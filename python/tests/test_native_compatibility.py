"""Compatibility delegates retain the existing owned native and Python lifetimes."""

import copy
import ctypes
import json
import os
from pathlib import Path
import threading
import time
from concurrent.futures import ThreadPoolExecutor

import pytest

from spl_toolkit import SPLMapper, SPLMapperError
from spl_toolkit.exceptions import MapperNotFoundError

FIXTURES = Path(os.environ["SPL_COMPATIBILITY_FIXTURES"])
LIBRARY = Path(os.environ["SPL_NATIVE_LIBRARY"])
assert FIXTURES.is_absolute() and FIXTURES.is_file()
assert LIBRARY.is_absolute() and LIBRARY.is_file()
CASES = json.loads(FIXTURES.read_text(encoding="utf-8"))


def open_mapper():
    return SPLMapper(library_path=str(LIBRARY))


def make_request(mapper, case):
    document = case["document"]
    requirements = mapper.requirements_query(document["text"], **{key: document[key] for key in
                                              ("language", "profile", "version", "source_id")})
    request = {"schema_version": 1, "requirements": requirements, "snapshot": copy.deepcopy(case["snapshot"]),
               "query_scope": copy.deepcopy(case["query_scope"]), "input_bindings": []}
    if case.get("schema_bundle") is not None:
        request["schema_bundle"] = copy.deepcopy(case["schema_bundle"])
    bindings_by_key = {}
    for source in requirements["inputs"]:
        key = source["kind"] + ":" + source["name"]
        if key not in case["bindings"]:
            continue
        binding = copy.deepcopy(case["bindings"][key])
        binding.pop("input", None)
        binding["input_id"] = source["id"]
        if "expected" not in binding:
            obj = next(obj for obj in case["snapshot"]["objects"] if obj["id"] == binding["object_id"])
            binding["expected"] = {key: obj[key] for key in ("kind", "name", "namespace", "app", "owner")}
        request["input_bindings"].append(binding)
        bindings_by_key[key] = binding
    for mutation in case.get("request_mutations", []):
        operation = mutation["op"]
        if operation in ("set", "remove"):
            parts = mutation["path"].strip("/").split("/")
            target = request
            for part in parts[:-1]:
                target = target[part]
            if operation == "set":
                target[parts[-1]] = copy.deepcopy(mutation["value"])
            else:
                del target[parts[-1]]
        elif operation == "duplicate_binding":
            request["input_bindings"].append(copy.deepcopy(bindings_by_key[mutation["input"]]))
        else:
            field = {"set_binding_input_id": "input_id", "set_binding_expected": "expected",
                     "set_binding_schema_id": "schema_id"}[operation]
            bindings_by_key[mutation["input"]][field] = copy.deepcopy(mutation["value"])
    return request


def raw_result(mapper, raw):
    pointer = mapper._lib.spl_mapper_check_compatibility(mapper._mapper_id, raw)
    assert pointer
    try:
        if pointer.contents.error is not None:
            assert pointer.contents.result is None
            return "error", json.loads(pointer.contents.error)
        assert pointer.contents.result is not None
        return "report", json.loads(pointer.contents.result)
    finally:
        mapper._lib.spl_result_free(pointer)


@pytest.mark.parametrize("case", CASES, ids=lambda case: case["name"])
def test_compatibility_native_python_parity_and_owned_copy(case):
    with open_mapper() as mapper:
        request = make_request(mapper, case)
        before = copy.deepcopy(request)
        raw = json.dumps(request).encode("utf-8")
        buffer = ctypes.create_string_buffer(raw)
        kind, result = raw_result(mapper, buffer)
        assert buffer.raw == raw + b"\0"
        if case["expected"].get("request_error"):
            assert kind == "error"
            assert result["code"] and isinstance(result["path"], str) and result["message"]
            with pytest.raises(SPLMapperError) as raised:
                mapper.check_compatibility(request)
            assert json.loads(str(raised.value)) == result
        else:
            assert kind == "report"
            assert result["outcome"] == case["expected"]["compatibility"]["outcome"]
            assert result["correlation"]["outcome"] == case["expected"]["correlation"]["outcome"]
            assert result == mapper.check_compatibility(request)
            result["reasons"].append({"code": "caller mutation"})
            assert mapper.check_compatibility(request) == raw_result(mapper, raw)[1]
        assert request == before


@pytest.mark.parametrize("raw", [b"", b"{", b"{}", b"\xff", b'{"schema_version":1,"unknown":true}'])
def test_compatibility_raw_invalid_requests_are_owned_structured_errors(raw):
    with open_mapper() as mapper:
        kind, result = raw_result(mapper, raw)
        assert kind == "error"
        assert result["code"] == "request_invalid" and result["message"]
        if raw in (b"", b"{", b"\xff"):
            assert isinstance(result["byte_offset"], int)


@pytest.mark.parametrize("outcome", ["success", "error", "decode"])
def test_compatibility_frees_every_owned_result(outcome, monkeypatch):
    with open_mapper() as mapper:
        request = make_request(mapper, CASES[0])
        frees = []
        native_free = mapper._lib.spl_result_free
        def free(pointer):
            frees.append(bool(pointer.contents.error))
            native_free(pointer)
        monkeypatch.setattr(mapper._lib, "spl_result_free", free)
        if outcome == "decode":
            def fail_decode(_):
                raise ValueError("decode failed")
            monkeypatch.setattr("spl_toolkit.mapper.json.loads", fail_decode)
            with pytest.raises(ValueError, match="decode failed"):
                mapper.check_compatibility(request)
        elif outcome == "error":
            request["input_bindings"] = []
            with pytest.raises(SPLMapperError):
                mapper.check_compatibility(request)
        else:
            assert mapper.check_compatibility(request)["outcome"] == "satisfied"
        assert frees == [outcome == "error"]
        assert mapper._active_calls == 0


def test_compatibility_encoding_failure_and_closed_mapper_do_not_allocate(monkeypatch):
    mapper = open_mapper()
    calls = []
    native = mapper._lib.spl_mapper_check_compatibility
    def tracked(*args):
        calls.append(args)
        return native(*args)
    monkeypatch.setattr(mapper._lib, "spl_mapper_check_compatibility", tracked)
    try:
        for value in (object(), float("nan"), float("inf")):
            with pytest.raises(SPLMapperError, match="Invalid compatibility request JSON"):
                mapper.check_compatibility({"value": value})
        assert mapper._active_calls == 0
    finally:
        mapper.close()
    with pytest.raises(MapperNotFoundError):
        mapper.check_compatibility({})
    assert calls == []


def test_compatibility_invalid_and_closed_handles_return_owned_errors():
    mapper = open_mapper()
    lib = mapper._lib
    old_handle = mapper._mapper_id
    mapper.close()
    for handle in (old_handle, -1, 2_147_483_647):
        pointer = lib.spl_mapper_check_compatibility(handle, b"{}")
        try:
            assert pointer and pointer.contents.error == b"Mapper not found"
            assert pointer.contents.result is None
        finally:
            lib.spl_result_free(pointer)
    lib.spl_result_free(None)


def test_compatibility_close_waits_for_admitted_call(monkeypatch):
    mapper = open_mapper()
    try:
        request = make_request(mapper, CASES[0])
        admitted, proceed = threading.Event(), threading.Event()
        native = mapper._lib.spl_mapper_check_compatibility
        def paused(*args):
            admitted.set()
            assert proceed.wait(timeout=5)
            return native(*args)
        monkeypatch.setattr(mapper._lib, "spl_mapper_check_compatibility", paused)
        with ThreadPoolExecutor(max_workers=2) as pool:
            result = pool.submit(mapper.check_compatibility, request)
            assert admitted.wait(timeout=5)
            closing = pool.submit(mapper.close)
            try:
                deadline = time.monotonic() + 5
                while True:
                    with mapper._condition:
                        if mapper._closing:
                            break
                    assert time.monotonic() < deadline, "close did not begin"
                    time.sleep(0.001)
                assert not closing.done()
                with pytest.raises(MapperNotFoundError):
                    mapper.check_compatibility(request)
            finally:
                proceed.set()
            assert result.result(timeout=5)["outcome"] == "satisfied"
            closing.result(timeout=5)
    finally:
        mapper.close()
