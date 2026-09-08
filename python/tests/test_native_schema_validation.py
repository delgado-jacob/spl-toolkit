"""Schema validation through the real shared library and owned results."""

from concurrent.futures import ThreadPoolExecutor
from copy import deepcopy
import ctypes
import gzip
import json
import os
from pathlib import Path
import threading
import time

import pytest

from spl_toolkit import SPLMapper, SPLMapperError
from spl_toolkit.exceptions import MapperNotFoundError


FIXTURES = Path(os.environ["SPL_SCHEMA_FIXTURES"])
assert FIXTURES.is_absolute(), "SPL_SCHEMA_FIXTURES must be absolute"
CASES = json.loads((FIXTURES / "cases.json").read_text(encoding="utf-8"))
REQUESTS = json.loads((FIXTURES / "requests.json").read_text(encoding="utf-8"))
TARGET = {"kind": "json_schema", "schema": {"type": "object", "properties": {"host": True},
          "required": ["host"], "additionalProperties": False}}


def mapper_kwargs():
    return {"library_path": os.environ["SPL_NATIVE_LIBRARY"]} if "SPL_NATIVE_LIBRARY" in os.environ else {}


def case_target(case):
    target = deepcopy(case["target"])
    if "catalog_fixture" in case:
        path = FIXTURES / case["catalog_fixture"]
        opener = gzip.open if path.suffix == ".gz" else open
        with opener(path, "rt", encoding="utf-8") as source:
            target["catalog"] = json.load(source)
    return target


def test_native_schema_required_and_owned_result():
    target = {"kind": "json_schema", "schema": {"type": "object", "properties": {"host": True}, "required": ["host"], "additionalProperties": False}}
    kwargs = {"library_path": os.environ["SPL_NATIVE_LIBRARY"]} if "SPL_NATIVE_LIBRARY" in os.environ else {}
    with SPLMapper(**kwargs) as mapper:
        report = mapper.validate_schema("table host", target)
        assert report["status"] == "valid"
        assert report["outcomes"][0]["outcome"] == "required"
        report["outcomes"].clear()
        assert mapper.validate_schema("table host", target)["outcomes"]


@pytest.mark.parametrize("case", CASES["cases"], ids=lambda case: case["id"])
def test_native_schema_corpus_full_reports(case):
    assert CASES["schema_version"] == 1
    target = case_target(case)
    document = deepcopy(case["document"])
    original = deepcopy((target, document))
    with SPLMapper(**mapper_kwargs()) as mapper:
        single = mapper.validate_schema(document["text"], target,
                                        **{k: v for k, v in document.items() if k != "text"})
        batch = mapper.validate_schema_batch([document], target)
    assert single == case["expected"]
    assert batch == {"schema_version": 1, "status": single["status"], "reports": [single]}
    assert (target, document) == original
    assert single["analysis"]["document"]["text"] == document["text"]
    for key in ("resource_uris", "members", "extensions", "limitations"):
        assert isinstance(single["target"][key], list)
    assert isinstance(single["diagnostics"], list)
    assert isinstance(single["coverage"]["reasons"], list)
    for outcome in single["outcomes"]:
        for key in ("matches", "evidence", "supporting_classes", "missing_classes", "indeterminate_classes"):
            assert isinstance(outcome[key], list)
        for match in outcome["matches"]:
            assert match["binding"] in ("source", "derived")
            assert isinstance(match["evidence"], list)


@pytest.mark.parametrize("case", REQUESTS["cases"], ids=lambda case: case["id"])
def test_raw_c_schema_rejects_request_corpus(case):
    assert REQUESTS["schema_version"] == 1
    payload = bytes.fromhex(case["hex"]) if "hex" in case else case["input"].encode("utf-8")
    with SPLMapper(**mapper_kwargs()) as mapper:
        native = mapper._lib.spl_mapper_validate_schema_batch if case["batch"] else mapper._lib.spl_mapper_validate_schema
        pointer = native(mapper._mapper_id, payload)
        try:
            assert pointer and pointer.contents.error
            assert pointer.contents.result is None
        finally:
            mapper._lib.spl_result_free(pointer)
        assert mapper.validate_schema("table host", TARGET)["status"] == "valid"


@pytest.mark.parametrize("batch", [False, True])
def test_raw_c_schema_nil_request_and_read_only_input(batch):
    with SPLMapper(**mapper_kwargs()) as mapper:
        native = mapper._lib.spl_mapper_validate_schema_batch if batch else mapper._lib.spl_mapper_validate_schema
        pointer = native(mapper._mapper_id, None)
        try:
            assert pointer and pointer.contents.error and pointer.contents.result is None
        finally:
            mapper._lib.spl_result_free(pointer)
        document = {"text": "table host", "source_id": "😀\x00"}
        request = {"documents": [document], "target": TARGET} if batch else {"document": document, "target": TARGET}
        payload = ctypes.create_string_buffer(json.dumps(request).encode())
        before = payload.raw
        pointer = native(mapper._mapper_id, payload)
        try:
            assert pointer and not pointer.contents.error
            report = json.loads(pointer.contents.result)
            assert (report["reports"][0] if batch else report)["analysis"]["document"]["source_id"] == "😀\x00"
            assert payload.raw == before
        finally:
            mapper._lib.spl_result_free(pointer)


def test_native_schema_metadata_locations_and_derived_evidence():
    query = 'search host="😀\x00"\r\n| eval café=host | table café'
    target = deepcopy(TARGET) | {"identity": "é😀\x00"}
    document = {"text": query, "source_id": "é😀\x00.spl", "language": "spl", "profile": "splunkd", "version": "current"}
    with SPLMapper(**mapper_kwargs()) as mapper:
        report = mapper.validate_schema(query, target, source_id=document["source_id"])
        assert mapper.validate_schema_batch([document], target)["reports"] == [report]
        assert mapper.validate_schema(query, target, language="", profile="", version="", source_id=document["source_id"]) == report
    assert report["analysis"]["document"] == document
    assert report["target"]["identity"] == target["identity"]
    assert report["status"] == "valid"
    bindings = {match["binding"] for item in report["outcomes"] for match in item["matches"]}
    assert bindings == {"source", "derived"}
    evidence = [e for item in report["outcomes"] for e in item["evidence"]]
    assert any(e.get("requirement") == "required" and e.get("pointer") == "/properties/host" for e in evidence)
    assert any(e.get("declaration_basis") == "derived" for e in evidence)
    for reference in report["analysis"]["references"]:
        location = reference["location"]
        assert query.encode()[location["start"]["offset"]:location["end"]["offset"]].decode() == reference["original_name"]


@pytest.mark.parametrize("documents, status, statuses", [
    ([{"text": "table host"}, {"text": "| mystery"}], "incomplete", ["valid", "incomplete"]),
    ([{"text": "| mystery"}, {"text": "table missing"}], "invalid", ["incomplete", "invalid"]),
])
def test_native_schema_batch_precedence(documents, status, statuses):
    with SPLMapper(**mapper_kwargs()) as mapper:
        result = mapper.validate_schema_batch(documents, TARGET)
    assert result["status"] == status
    assert [r["status"] for r in result["reports"]] == statuses


@pytest.mark.parametrize("target", [None, {}, {"kind": "json_schema", "schema": None},
    {"kind": "json_schema", "schema": True, "identity": "\ud800"},
    {"kind": "json_schema", "schema": {"properties": {"\udc00": True}}}])
@pytest.mark.parametrize("batch", [False, True])
def test_native_schema_bad_target_does_not_poison_calls(target, batch):
    with SPLMapper(**mapper_kwargs()) as mapper:
        with pytest.raises(SPLMapperError):
            if batch:
                mapper.validate_schema_batch([{"text": "table host"}], target)
            else:
                mapper.validate_schema("table host", target)
        assert mapper.validate_schema("table host", TARGET)["status"] == "valid"


@pytest.mark.parametrize("options", [{"language": "spl2"}, {"profile": "other"},
    {"version": "9"}, {"source_id": "\ud800"}, {"source_id": None}, {"language": 1}])
def test_native_schema_options_rejected(options):
    with SPLMapper(**mapper_kwargs()) as mapper:
        with pytest.raises(SPLMapperError):
            mapper.validate_schema("table host", TARGET, **options)
        with pytest.raises(SPLMapperError):
            mapper.validate_schema_batch([{"text": "table host", **options}], TARGET)


@pytest.mark.parametrize("documents", [None, [], [{}], [{"text": None}], [{"text": 1}],
    [{"text": "\ud800"}], [{"text": "table host", "unknown": "x"}]])
def test_native_schema_batch_invalid_documents(documents):
    with SPLMapper(**mapper_kwargs()) as mapper:
        with pytest.raises(SPLMapperError):
            mapper.validate_schema_batch(documents, TARGET)


def test_native_schema_input_and_nested_result_isolation():
    target = deepcopy(TARGET)
    documents = [{"text": "table host", "source_id": "original"}]
    with SPLMapper(**mapper_kwargs()) as mapper:
        result = mapper.validate_schema_batch(documents, target)
        expected = deepcopy(result)
        result["reports"][0]["outcomes"][0]["matches"][0]["evidence"].clear()
        result["reports"][0]["target"]["limitations"].clear()
        assert mapper.validate_schema_batch(documents, target) == expected
        target["schema"]["properties"].clear()
        documents[0]["text"] = "table missing"
        assert expected["reports"][0]["analysis"]["document"]["text"] == "table host"
        assert mapper.validate_schema_batch(documents, target)["status"] == "invalid"


@pytest.mark.parametrize("operation", ["validate_schema", "validate_schema_batch"])
def test_schema_serialization_failure_releases_operation(operation):
    circular = []
    circular.append(circular)
    with SPLMapper(**mapper_kwargs()) as mapper:
        for catalog in ([object()], circular, [float("nan")]):
            args = ("table host", catalog) if operation == "validate_schema" else ([{"text": "table host"}], catalog)
            with pytest.raises(SPLMapperError):
                getattr(mapper, operation)(*args)
            assert mapper._active_calls == 0




@pytest.mark.parametrize("operation", ["validate_schema", "validate_schema_batch"])
def test_schema_invalid_and_closed_handles_return_owned_errors(operation):
    mapper = SPLMapper(**mapper_kwargs())
    lib, handle = mapper._lib, mapper._mapper_id
    mapper.close()
    args = ("table host", TARGET) if operation == "validate_schema" else ([{"text": "table host"}], TARGET)
    with pytest.raises(MapperNotFoundError):
        getattr(mapper, operation)(*args)
    for invalid in (handle, -1, 2_147_483_647):
        pointer = getattr(lib, "spl_mapper_" + operation)(invalid, b"{}")
        try:
            assert pointer and pointer.contents.error == b"Mapper not found"
            assert pointer.contents.result is None
        finally:
            lib.spl_result_free(pointer)


@pytest.mark.parametrize("operation", ["validate_schema", "validate_schema_batch"])
def test_schema_owned_result_freed_when_json_decode_fails(operation, monkeypatch):
    with SPLMapper(**mapper_kwargs()) as mapper:
        freed = []
        real_free = mapper._lib.spl_result_free

        def free(pointer):
            freed.append(bool(pointer))
            real_free(pointer)

        def fail_decode(_value):
            raise ValueError("decode failed")

        monkeypatch.setattr(mapper._lib, "spl_result_free", free)
        monkeypatch.setattr("spl_toolkit.mapper.json.loads", fail_decode)
        args = ("table host", TARGET) if operation == "validate_schema" else ([{"text": "table host"}], TARGET)
        with pytest.raises(ValueError, match="decode failed"):
            getattr(mapper, operation)(*args)
        assert freed == [True]
        assert mapper._active_calls == 0


def test_native_schema_repeat_and_concurrent_calls_are_deterministic():
    with SPLMapper(**mapper_kwargs()) as mapper:
        expected = mapper.validate_schema("eval label=host | table label", TARGET)

        def work(_):
            assert mapper.validate_schema("eval label=host | table label", TARGET) == expected
            assert mapper.validate_schema_batch([{"text": "eval label=host | table label"}], TARGET)["reports"] == [expected]

        for _ in range(100):
            work(0)
        with ThreadPoolExecutor(max_workers=16) as pool:
            list(pool.map(work, range(128)))


@pytest.mark.parametrize("operation", ["validate_schema", "validate_schema_batch"])
def test_schema_close_waits_for_admitted_native_operation(operation, monkeypatch):
    mapper = SPLMapper(**mapper_kwargs())
    admitted, proceed = threading.Event(), threading.Event()
    native_name = "spl_mapper_" + operation
    native = getattr(mapper._lib, native_name)

    def paused_native(*args):
        admitted.set()
        assert proceed.wait(timeout=5)
        return native(*args)

    monkeypatch.setattr(mapper._lib, native_name, paused_native)
    args = ("table host", TARGET) if operation == "validate_schema" else ([{"text": "table host"}], TARGET)
    with ThreadPoolExecutor(max_workers=3) as pool:
        result = pool.submit(getattr(mapper, operation), *args)
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
                getattr(mapper, operation)(*args)
        finally:
            proceed.set()
        assert result.result(timeout=5)["status"] == "valid"
        closing.result(timeout=5)
    mapper.close()


@pytest.mark.parametrize("operation", ["validate_schema", "validate_schema_batch"])
@pytest.mark.parametrize("target", [TARGET, {}], ids=["success", "error"])
def test_schema_python_frees_real_success_and_error_results(operation, target, monkeypatch):
    with SPLMapper(**mapper_kwargs()) as mapper:
        freed = []
        real_free = mapper._lib.spl_result_free

        def free(pointer):
            freed.append((bool(pointer), bool(pointer.contents.error)))
            real_free(pointer)

        monkeypatch.setattr(mapper._lib, "spl_result_free", free)
        args = ("table host", target) if operation == "validate_schema" else ([{"text": "table host"}], target)
        if target:
            assert getattr(mapper, operation)(*args)["status"] == "valid"
        else:
            with pytest.raises(SPLMapperError):
                getattr(mapper, operation)(*args)
        assert freed == [(True, not bool(target))]
        assert mapper._active_calls == 0
        assert mapper.validate_schema("table host", TARGET)["status"] == "valid"
