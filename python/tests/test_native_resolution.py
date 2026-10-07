"""Resolution delegates retain the existing owned native and Python lifetimes."""

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

FIXTURES = Path(os.environ["SPL_RESOLUTION_FIXTURES"])
LIBRARY = Path(os.environ["SPL_NATIVE_LIBRARY"])
assert FIXTURES.is_absolute() and FIXTURES.is_file()
assert LIBRARY.is_absolute() and LIBRARY.is_file()
CASES = json.loads(FIXTURES.read_text(encoding="utf-8"))


def open_mapper():
    return SPLMapper(library_path=str(LIBRARY))


def make_request(mapper, case):
    document = copy.deepcopy(case["document"])
    analysis = mapper.analyze_query(document["text"], **{key: document[key] for key in
                                  ("language", "profile", "version", "source_id")})
    request = {"schema_version": 1, "document": document,
               "resolutions": [{"placeholder": item["marker"], "kind": item["kind"],
                                "values": copy.deepcopy(item["values"])} for item in case["resolutions"]],
               "compatibility": copy.deepcopy(case["compatibility"])}
    if request["compatibility"].get("schema_bundle") is None:
        request["compatibility"].pop("schema_bundle", None)
    if "max_variants" in case:
        request["max_variants"] = case["max_variants"]
    for binding in request["compatibility"]["input_bindings"]:
        selector = binding.pop("original_input")
        matches = [item for item in analysis["inputs"] if item["kind"] == selector["kind"]
                   and item["identity"] == selector["identity"]]
        assert len(matches) == 1, selector
        binding["original_input_id"] = matches[0]["id"]
    return request


def assert_authored_report(result, case):
    expected = case["expected"]
    assert result["counts"] == expected["counts"]
    assert result["generated_count"] == len(expected["variants"])
    assert result["total_combinations"] == str(len(expected["variants"]))
    assert len(result["variants"]) == len(expected["variants"])
    for ordinal, (got, want) in enumerate(zip(result["variants"], expected["variants"]), 1):
        assert got["ordinal"] == ordinal
        assert [choice["value"] for choice in got["selection"]] == want["selection"]
        assert got["outcome"] == want["outcome"]
        assert got.get("resolved_query") == want["resolved_query"]
        if got["outcome"] == "verified":
            assert got["proof"]["proven"]
            assert got["compatibility"]["outcome"] == "satisfied"
        for change in got["changes"]:
            location = change["original_location"]
            assert case["document"]["text"].encode()[location["start"]["offset"]:
                                                       location["end"]["offset"]].decode() == change["before"]
            if "candidate_location" in change:
                location = change["candidate_location"]
                assert got["candidate_text"].encode()[location["start"]["offset"]:
                                                      location["end"]["offset"]].decode() == change["after"]


def raw_result(mapper, raw):
    pointer = mapper._lib.spl_mapper_resolve(mapper._mapper_id, raw)
    assert pointer
    try:
        if pointer.contents.error is not None:
            assert pointer.contents.result is None
            return "error", json.loads(pointer.contents.error)
        assert pointer.contents.result is not None
        return "report", json.loads(pointer.contents.result)
    finally:
        mapper._lib.spl_result_free(pointer)


@pytest.mark.parametrize("case", CASES, ids=lambda case: case["id"])
def test_resolution_native_python_parity_and_owned_copy(case):
    with open_mapper() as mapper:
        request = make_request(mapper, case)
        before = copy.deepcopy(request)
        raw = json.dumps(request).encode("utf-8")
        buffer = ctypes.create_string_buffer(raw)
        kind, result = raw_result(mapper, buffer)
        assert buffer.raw == raw + b"\0"
        if case["expected"].get("request_error"):
            assert kind == "error"
            assert result["code"] == case["expected"]["request_error"]
            assert result["total_combinations"] == str(case["expected"]["combination_count"])
            assert isinstance(result["path"], str) and result["message"]
            with pytest.raises(SPLMapperError) as raised:
                mapper.resolve(request)
            assert json.loads(str(raised.value)) == result
        else:
            assert kind == "report"
            assert_authored_report(result, case)
            assert result == mapper.resolve(request)
            result["variants"][0]["diagnostics"].append({"code": "caller mutation"})
            assert mapper.resolve(request) == raw_result(mapper, raw)[1]
        assert request == before


@pytest.mark.parametrize("raw", [b"", b"{", b"{}", b"\xff", b'{"schema_version":1,"unknown":true}'])
def test_resolution_raw_invalid_requests_are_owned_structured_errors(raw):
    with open_mapper() as mapper:
        kind, result = raw_result(mapper, raw)
        assert kind == "error"
        assert result["code"] == "request_invalid" and result["message"]
        if raw in (b"", b"{", b"\xff"):
            assert isinstance(result["byte_offset"], int)


@pytest.mark.parametrize("outcome", ["success", "error", "decode"])
def test_resolution_frees_every_owned_result(outcome, monkeypatch):
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
                mapper.resolve(request)
        elif outcome == "error":
            request["resolutions"] = None
            with pytest.raises(SPLMapperError):
                mapper.resolve(request)
        else:
            assert mapper.resolve(request)["counts"]["verified"] == 1
        assert frees == [outcome == "error"]
        assert mapper._active_calls == 0


def test_resolution_encoding_failure_and_closed_mapper_do_not_allocate(monkeypatch):
    mapper = open_mapper()
    calls = []
    native = mapper._lib.spl_mapper_resolve
    def tracked(*args):
        calls.append(args)
        return native(*args)
    monkeypatch.setattr(mapper._lib, "spl_mapper_resolve", tracked)
    try:
        for value in (object(), float("nan"), float("inf")):
            with pytest.raises(SPLMapperError, match="Invalid resolution request JSON"):
                mapper.resolve({"value": value})
        assert mapper._active_calls == 0
    finally:
        mapper.close()
    with pytest.raises(MapperNotFoundError):
        mapper.resolve({})
    assert calls == []


def test_resolution_invalid_and_closed_handles_return_owned_errors():
    mapper = open_mapper()
    lib = mapper._lib
    old_handle = mapper._mapper_id
    mapper.close()
    for handle in (old_handle, -1, 2_147_483_647):
        pointer = lib.spl_mapper_resolve(handle, None)
        try:
            assert pointer and pointer.contents.error == b"Mapper not found"
            assert pointer.contents.result is None
        finally:
            lib.spl_result_free(pointer)
    lib.spl_result_free(None)


def test_resolution_close_waits_for_admitted_call(monkeypatch):
    mapper = open_mapper()
    try:
        request = make_request(mapper, CASES[0])
        admitted, proceed = threading.Event(), threading.Event()
        native = mapper._lib.spl_mapper_resolve
        def paused(*args):
            admitted.set()
            assert proceed.wait(timeout=5)
            return native(*args)
        monkeypatch.setattr(mapper._lib, "spl_mapper_resolve", paused)
        with ThreadPoolExecutor(max_workers=2) as pool:
            result = pool.submit(mapper.resolve, request)
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
                    mapper.resolve(request)
            finally:
                proceed.set()
            assert result.result(timeout=5)["counts"]["verified"] == 1
            closing.result(timeout=5)
    finally:
        mapper.close()


def test_resolution_null_request_is_an_owned_error():
    with open_mapper() as mapper:
        kind, detail = raw_result(mapper, None)
        assert kind == "error" and detail["code"] == "request_invalid"


def test_resolution_unicode_request_and_audit_survive_native_free():
    case = copy.deepcopy(CASES[0])
    case["document"]["text"] = 'from $events | eval label="雪é" | fields id'
    case["document"]["source_id"] = "résolution:雪"
    with open_mapper() as mapper:
        request = make_request(mapper, case)
        before = copy.deepcopy(request)
        result = mapper.resolve(request)
        kind, native = raw_result(mapper, json.dumps(request, ensure_ascii=False).encode())
        assert kind == "report" and result == native
        assert request == before
        assert "雪é" in result["variants"][0]["resolved_query"]
        for change in result["variants"][0]["changes"]:
            location = change["original_location"]
            assert request["document"]["text"].encode()[location["start"]["offset"]:
                                                         location["end"]["offset"]].decode() == change["before"]
    # Both native results and the mapper have been freed; Python owns the decoded copy.
    assert "雪é" in result["variants"][0]["resolved_query"]


@pytest.mark.parametrize("mutation,code", [("version", "request_invalid"), ("kind", "request_invalid"),
                                         ("duplicate_binding", "binding_invalid"), ("null_schema", "request_invalid")])
def test_resolution_configuration_errors_match_native_details(mutation, code):
    with open_mapper() as mapper:
        request = make_request(mapper, CASES[0])
        if mutation == "version":
            request["schema_version"] = 2
        elif mutation == "kind":
            request["resolutions"][0]["kind"] = "unknown"
        elif mutation == "duplicate_binding":
            request["compatibility"]["input_bindings"].append(
                copy.deepcopy(request["compatibility"]["input_bindings"][0]))
        else:
            request["compatibility"]["schema_bundle"] = None
        kind, detail = raw_result(mapper, json.dumps(request).encode())
        assert kind == "error" and detail["code"] == code
        with pytest.raises(SPLMapperError) as raised:
            mapper.resolve(request)
        assert json.loads(str(raised.value)) == detail
