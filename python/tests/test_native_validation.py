"""Field-list validation and result ownership through the real shared library."""

from concurrent.futures import ThreadPoolExecutor
import json
import os
import threading
import time

import pytest

from spl_toolkit import SPLMapper, SPLMapperError
from spl_toolkit.exceptions import MapperNotFoundError


def mapper_kwargs():
    return {"library_path": os.environ["SPL_NATIVE_LIBRARY"]} if "SPL_NATIVE_LIBRARY" in os.environ else {}


def test_native_validation_matches_created_field():
    with SPLMapper(**mapper_kwargs()) as mapper:
        report = mapper.validate_fields("eval label=host | table label", ["host"])
    assert report["status"] == "valid"
    assert [item["outcome"] for item in report["outcomes"]] == ["matching", "matching"]
    assert report["analysis"]["document"]["text"] == "eval label=host | table label"


@pytest.mark.parametrize("query, catalog, status, outcomes", [
    ("table absent", [], "invalid", ["missing"]),
    ("table host", {"fields": [], "optional_fields": ["host"]}, "valid", ["optional_equivalent"]),
    ("| mystery", ["host"], "incomplete", []),
    ("", [], "invalid", []),
    (" \t\r\n", [], "invalid", []),
])
def test_native_validation_content_is_a_report(query, catalog, status, outcomes):
    with SPLMapper(**mapper_kwargs()) as mapper:
        report = mapper.validate_fields(query, catalog)
    assert type(report["schema_version"]) is int and report["schema_version"] == 1
    assert report["status"] == status
    assert [item["outcome"] for item in report["outcomes"]] == outcomes
    assert isinstance(report["diagnostics"], list)
    assert isinstance(report["coverage"]["reasons"], list)
    assert isinstance(report["target"]["optional_fields"], list)
    assert all(isinstance(item["matches"], list) for item in report["outcomes"])


@pytest.mark.parametrize("documents, statuses, status", [
    ([{"text": "table host"}], ["valid"], "valid"),
    ([{"text": "table host"}, {"text": "| mystery"}], ["valid", "incomplete"], "incomplete"),
    ([{"text": "| mystery"}, {"text": "table missing"}], ["incomplete", "invalid"], "invalid"),
])
def test_native_batch_preserves_order_and_status_precedence(documents, statuses, status):
    with SPLMapper(**mapper_kwargs()) as mapper:
        report = mapper.validate_fields_batch(documents, ["host"])
    assert type(report["schema_version"]) is int and report["schema_version"] == 1
    assert report["status"] == status
    assert [item["status"] for item in report["reports"]] == statuses
    assert [item["analysis"]["document"]["text"] for item in report["reports"]] == [item["text"] for item in documents]


def test_native_validation_preserves_metadata_unicode_and_defaults():
    query = 'search host="😀\x00"\r\n| eval café=host | table café'
    catalog = {"fields": ["host"], "optional_fields": [], "identity": "é😀\x00", "version": "v😀"}
    document = {"text": query, "source_id": "é😀\x00.spl", "language": "spl", "profile": "splunkd", "version": "current"}
    with SPLMapper(**mapper_kwargs()) as mapper:
        single = mapper.validate_fields(query, catalog, source_id=document["source_id"])
        batch = mapper.validate_fields_batch([document], catalog)
        assert single == mapper.validate_fields(query, catalog, language="", profile="", version="", source_id=document["source_id"])
    assert batch["reports"] == [single]
    assert single["analysis"]["document"] == document
    assert single["target"] == {"kind": "field_list", **catalog}
    for reference in single["analysis"]["references"]:
        location = reference["location"]
        assert query.encode()[location["start"]["offset"]:location["end"]["offset"]].decode() == reference["original_name"]


@pytest.mark.parametrize("catalog", [None, {}, [1], [""], ["host", "host"],
    {"fields": ["host"], "optional_fields": ["host"]}, {"fields": [], "unknown": []},
    {"fields": [], "identity": None}, ["\ud800"], {"fields": [], "version": "\udc00"}])
@pytest.mark.parametrize("batch", [False, True])
def test_native_validation_catalog_errors_are_mapper_errors(catalog, batch):
    with SPLMapper(**mapper_kwargs()) as mapper:
        with pytest.raises(SPLMapperError):
            if batch:
                mapper.validate_fields_batch([{"text": "table host"}], catalog)
            else:
                mapper.validate_fields("table host", catalog)
        assert mapper.validate_fields("table host", ["host"])["status"] == "valid"


@pytest.mark.parametrize("options", [{"language": "spl2"}, {"profile": "other"},
    {"version": "9"}, {"source_id": "\ud800"}, {"source_id": None}, {"language": 1}])
def test_native_validation_options_are_mapper_errors(options):
    with SPLMapper(**mapper_kwargs()) as mapper:
        with pytest.raises(SPLMapperError):
            mapper.validate_fields("table host", ["host"], **options)
        with pytest.raises(SPLMapperError):
            mapper.validate_fields_batch([{"text": "table host", **options}], ["host"])


@pytest.mark.parametrize("documents", [None, [], [{}], [{"text": None}], [{"text": 1}],
    [{"text": "\ud800"}], [{"text": "table host", "unknown": "x"}]])
def test_native_batch_document_errors_are_mapper_errors(documents):
    with SPLMapper(**mapper_kwargs()) as mapper:
        with pytest.raises(SPLMapperError):
            mapper.validate_fields_batch(documents, ["host"])


@pytest.mark.parametrize("operation", ["validate_fields", "validate_fields_batch"])
def test_validation_serialization_failure_releases_operation(operation):
    circular = []
    circular.append(circular)
    with SPLMapper(**mapper_kwargs()) as mapper:
        for catalog in ([object()], circular, [float("nan")]):
            args = ("table host", catalog) if operation == "validate_fields" else ([{"text": "table host"}], catalog)
            with pytest.raises(SPLMapperError):
                getattr(mapper, operation)(*args)
            assert mapper._active_calls == 0


@pytest.mark.parametrize("operation, key, document", [
    ("validate_fields", "document", b'{"text":"table host"}'),
    ("validate_fields_batch", "documents", b'[{"text":"table host"}]'),
])
@pytest.mark.parametrize("bad", [None, b"{", b"[]", b"null", b"missing", b"trailing",
    b"duplicate", b"duplicate_catalog", b"bad_utf8", b"lone_surrogate"])
def test_raw_c_validation_rejects_malformed_requests(operation, key, document, bad):
    payload = b'{"' + key.encode() + b'":' + document + b',"catalog":["host"]}'
    if bad == b"missing":
        payload = b'{"catalog":[]}'
    elif bad == b"trailing":
        payload += b" {}"
    elif bad == b"duplicate":
        payload = payload[:-1] + b',"catalog":[]}'
    elif bad == b"duplicate_catalog":
        payload = payload.replace(b'["host"]', b'{"fields":[],"fields":[]}')
    elif bad == b"bad_utf8":
        payload = payload.replace(b'host', b'\xff')
    elif bad == b"lone_surrogate":
        payload = payload.replace(b'host', b'\\ud800')
    else:
        payload = bad
    with SPLMapper(**mapper_kwargs()) as mapper:
        pointer = getattr(mapper._lib, "spl_mapper_" + operation)(mapper._mapper_id, payload)
        try:
            assert pointer and pointer.contents.error
            assert pointer.contents.result is None
        finally:
            mapper._lib.spl_result_free(pointer)


@pytest.mark.parametrize("batch", [False, True])
@pytest.mark.parametrize("escaped, source_id", [(br'\ud83d\ude00', "😀"), (br'\\ud800', r"\ud800")])
def test_raw_c_validation_preserves_valid_unicode_escapes(batch, escaped, source_id):
    document = b'{"text":"table host","source_id":"' + escaped + b'"}'
    payload = (b'{"documents":[' + document + b']' if batch else b'{"document":' + document) + b',"catalog":["host"]}'
    with SPLMapper(**mapper_kwargs()) as mapper:
        native = mapper._lib.spl_mapper_validate_fields_batch if batch else mapper._lib.spl_mapper_validate_fields
        pointer = native(mapper._mapper_id, payload)
        try:
            assert pointer and not pointer.contents.error
            report = json.loads(pointer.contents.result)
            if batch:
                report = report["reports"][0]
            assert report["analysis"]["document"]["source_id"] == source_id
        finally:
            mapper._lib.spl_result_free(pointer)


@pytest.mark.parametrize("operation", ["validate_fields", "validate_fields_batch"])
def test_validation_invalid_and_closed_handles_return_owned_errors(operation):
    mapper = SPLMapper(**mapper_kwargs())
    lib, handle = mapper._lib, mapper._mapper_id
    mapper.close()
    args = ("table host", ["host"]) if operation == "validate_fields" else ([{"text": "table host"}], ["host"])
    with pytest.raises(MapperNotFoundError):
        getattr(mapper, operation)(*args)
    for invalid in (handle, -1, 2_147_483_647):
        pointer = getattr(lib, "spl_mapper_" + operation)(invalid, b"{}")
        try:
            assert pointer and pointer.contents.error == b"Mapper not found"
            assert pointer.contents.result is None
        finally:
            lib.spl_result_free(pointer)


@pytest.mark.parametrize("operation", ["validate_fields", "validate_fields_batch"])
def test_validation_owned_result_freed_when_json_decode_fails(operation, monkeypatch):
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
        args = ("table host", ["host"]) if operation == "validate_fields" else ([{"text": "table host"}], ["host"])
        with pytest.raises(ValueError, match="decode failed"):
            getattr(mapper, operation)(*args)
        assert freed == [True]
        assert mapper._active_calls == 0


def test_native_validation_repeat_and_concurrent_calls_are_deterministic():
    with SPLMapper(**mapper_kwargs()) as mapper:
        expected = mapper.validate_fields("eval label=host | table label", ["host"])

        def work(_):
            assert mapper.validate_fields("eval label=host | table label", ["host"]) == expected
            assert mapper.validate_fields_batch([{"text": "eval label=host | table label"}], ["host"])["reports"] == [expected]

        for _ in range(100):
            work(0)
        with ThreadPoolExecutor(max_workers=16) as pool:
            list(pool.map(work, range(128)))


@pytest.mark.parametrize("operation", ["validate_fields", "validate_fields_batch"])
def test_validation_close_waits_for_admitted_native_operation(operation, monkeypatch):
    mapper = SPLMapper(**mapper_kwargs())
    admitted, proceed = threading.Event(), threading.Event()
    native_name = "spl_mapper_" + operation
    native = getattr(mapper._lib, native_name)

    def paused_native(*args):
        admitted.set()
        assert proceed.wait(timeout=5)
        return native(*args)

    monkeypatch.setattr(mapper._lib, native_name, paused_native)
    args = ("table host", ["host"]) if operation == "validate_fields" else ([{"text": "table host"}], ["host"])
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
