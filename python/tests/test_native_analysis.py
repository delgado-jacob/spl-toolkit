"""Canonical analysis and owned-result lifecycle through the real shared library."""

from concurrent.futures import ThreadPoolExecutor
import json
import os
from pathlib import Path
import threading
import time

import pytest

from spl_toolkit import SPLMapper, SPLMapperError
from spl_toolkit.exceptions import MapperNotFoundError


def mapper_kwargs():
    return {"library_path": os.environ["SPL_NATIVE_LIBRARY"]} if "SPL_NATIVE_LIBRARY" in os.environ else {}


def corpus():
    path = Path(os.environ.get("SPL_ANALYSIS_FIXTURES", Path(__file__).resolve().parents[2] / "testdata/analysis/cases.json"))
    data = json.loads(path.read_text(encoding="utf-8"))
    assert data["version"] == "1" and data["cases"]
    return data["cases"]


def test_analysis_matches_every_canonical_report():
    with SPLMapper(**mapper_kwargs()) as mapper:
        for case in corpus():
            document = case["document"].copy()
            query = document.pop("text")
            assert mapper.analyze_query(query, **document) == case["expected"], case["id"]


def test_analysis_defaults_options_and_partial_result():
    with SPLMapper(**mapper_kwargs()) as mapper:
        query = "search host=web | mystery x"
        report = mapper.analyze_query(query, source_id="q.spl")
        assert report == mapper.analyze_query(query, language="spl", profile="splunkd", version="current", source_id="q.spl")
        assert report == mapper.analyze_query(query, language="", profile="", version="", source_id="q.spl")
    assert type(report["schema_version"]) is int and report["schema_version"] == 1
    assert report["status"] == "incomplete"
    assert report["document"] == {"text": query, "source_id": "q.spl", "language": "spl", "profile": "splunkd", "version": "current"}
    assert any(r["normalized_name"] == "host" for r in report["references"])
    assert any(d["code"] == "SPL_UNSUPPORTED_COMMAND" for d in report["diagnostics"])


@pytest.mark.parametrize("query", ["", " \t\r\n", "search host=\"😀\"\r\n| eval 'café'=host", "search host=\"a\x00b\""])
def test_analysis_preserves_source_and_arrays(query):
    with SPLMapper(**mapper_kwargs()) as mapper:
        report = mapper.analyze_query(query, source_id="é😀\x00.spl")
    assert report["document"]["text"] == query
    assert report["document"]["source_id"] == "é😀\x00.spl"
    for key in ("references", "stages", "scopes", "lineage", "diagnostics"):
        assert isinstance(report[key], list)
    assert all(isinstance(value, list) for value in report["dependencies"].values())
    for reference in report["references"]:
        location = reference["location"]
        assert query.encode()[location["start"]["offset"]:location["end"]["offset"]].decode() == reference["original_name"]


@pytest.mark.parametrize("options, message", [({"language": "spl2"}, "unsupported language"), ({"profile": "other"}, "unsupported profile"), ({"version": "9"}, "unsupported compatibility version"), ({"source_id": "\ud800"}, "unpaired UTF-16")])
def test_analysis_options_are_api_errors(options, message):
    with SPLMapper(**mapper_kwargs()) as mapper:
        with pytest.raises(SPLMapperError, match=message):
            mapper.analyze_query("search host=web", **options)
        assert mapper.analyze_query("search host=web")["status"] == "valid"


@pytest.mark.parametrize("query", ["\ud800", "\udc00"])
def test_analysis_rejects_lone_python_surrogates(query):
    with SPLMapper(**mapper_kwargs()) as mapper:
        with pytest.raises(SPLMapperError, match="unpaired UTF-16"):
            mapper.analyze_query(query)


@pytest.mark.parametrize("payload", [b"{", b"[]", b'{"text":123}', b'{"text":"x"} trailing', b'{"text":"\xff"}', b'{"text":"x","source_id":"\xff"}', b'{"text":"\\ud800"}', b'{"source_id":"\\udc00"}', b'{"text":"\\ud800x"}', b'{"text":"\\ud800\\u0041"}', None])
def test_c_analysis_rejects_bad_json_without_report(payload):
    with SPLMapper(**mapper_kwargs()) as mapper:
        pointer = mapper._lib.spl_mapper_analyze_query(mapper._mapper_id, payload)
        assert pointer
        try:
            assert pointer.contents.error
            assert pointer.contents.result is None
        finally:
            mapper._lib.spl_result_free(pointer)


@pytest.mark.parametrize("payload, text", [(b'{"text":"search host=\\"\\ud83d\\ude00\\""}', 'search host="😀"'), (b'{"text":"search host=\\"\\\\ud800\\""}', 'search host="\\ud800"')])
def test_c_analysis_preserves_valid_unicode_escapes(payload, text):
    with SPLMapper(**mapper_kwargs()) as mapper:
        pointer = mapper._lib.spl_mapper_analyze_query(mapper._mapper_id, payload)
        try:
            assert pointer and not pointer.contents.error
            assert json.loads(pointer.contents.result)["document"]["text"] == text
        finally:
            mapper._lib.spl_result_free(pointer)


def test_capabilities_are_owned_fresh_json():
    with SPLMapper(**mapper_kwargs()) as mapper:
        manifest = mapper.capabilities()
        assert type(manifest["schema_version"]) is int and manifest["schema_version"] == 1
        assert (manifest["language"], manifest["profile"], manifest["version"]) == ("spl", "splunkd", "current")
        assert any(c["name"] == "eval" and c["semantic_supported"] for c in manifest["commands"])
        assert any(f["name"] == "lower" and f["semantic_supported"] for f in manifest["functions"])
        for _ in range(100):
            pointer = mapper._lib.spl_mapper_capabilities(mapper._mapper_id)
            try:
                assert pointer and not pointer.contents.error
                assert json.loads(pointer.contents.result) == manifest
            finally:
                mapper._lib.spl_result_free(pointer)
        mapper.capabilities()["commands"].clear()
        assert mapper.capabilities() == manifest


@pytest.mark.parametrize("operation", ["analyze_query", "capabilities"])
def test_invalid_and_closed_handles_return_owned_errors(operation):
    mapper = SPLMapper(**mapper_kwargs())
    lib = mapper._lib
    old_handle = mapper._mapper_id
    mapper.close()
    method = getattr(mapper, operation)
    args = ("search host=web",) if operation == "analyze_query" else ()
    with pytest.raises(MapperNotFoundError):
        method(*args)
    native = getattr(lib, "spl_mapper_" + operation)
    native_args = (b'{"text":"search host=web"}',) if args else ()
    for handle in (old_handle, -1, 2_147_483_647):
        pointer = native(handle, *native_args)
        try:
            assert pointer and pointer.contents.error == b"Mapper not found"
            assert pointer.contents.result is None
        finally:
            lib.spl_result_free(pointer)
    lib.spl_result_free(None)


@pytest.mark.parametrize("operation", ["analyze_query", "capabilities"])
def test_owned_result_freed_when_json_decode_fails(operation, monkeypatch):
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
        with pytest.raises(ValueError, match="decode failed"):
            getattr(mapper, operation)(*("search host=web",) if operation == "analyze_query" else ())
        assert freed == [True]
        assert mapper._active_calls == 0


def test_shared_native_analysis_is_deterministic_under_concurrency():
    case = corpus()[0]
    with SPLMapper(**mapper_kwargs()) as mapper:
        manifest = mapper.capabilities()
        def work(_):
            report = mapper.analyze_query(case["document"]["text"], source_id=case["document"]["source_id"])
            assert mapper.capabilities() == manifest
            return report
        with ThreadPoolExecutor(max_workers=16) as pool:
            assert all(report == case["expected"] for report in pool.map(work, range(128)))


@pytest.mark.parametrize("operation", ["analyze_query", "capabilities"])
def test_close_waits_for_admitted_native_operation(operation, monkeypatch):
    mapper = SPLMapper(**mapper_kwargs())
    admitted = threading.Event()
    proceed = threading.Event()
    native_name = "spl_mapper_" + operation
    native = getattr(mapper._lib, native_name)

    def paused_native(*args):
        admitted.set()
        assert proceed.wait(timeout=5)
        return native(*args)

    monkeypatch.setattr(mapper._lib, native_name, paused_native)
    args = ("search host=web",) if operation == "analyze_query" else ()
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
        assert result.result(timeout=5)["schema_version"] == 1
        closing.result(timeout=5)
    mapper.close()
