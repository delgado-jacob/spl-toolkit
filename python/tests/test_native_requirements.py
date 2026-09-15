"""Canonical requirements and owned-result lifecycle through the real library."""

from concurrent.futures import ThreadPoolExecutor
import hashlib
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
    path = Path(os.environ.get("SPL_REQUIREMENTS_FIXTURES", Path(__file__).resolve().parents[2] / "testdata/requirements/cases.json"))
    data = json.loads(path.read_text(encoding="utf-8"))
    assert data["version"] == "1" and data["cases"]
    return data["cases"]


def test_requirements_matches_canonical_fixture():
    case = corpus()[0]
    document = case["document"].copy()
    query = document.pop("text")
    with SPLMapper(**mapper_kwargs()) as mapper:
        assert mapper.requirements_query(query, **document) == case["expected"]


def test_requirements_matches_every_canonical_fixture_and_embedded_analysis():
    with SPLMapper(**mapper_kwargs()) as mapper:
        for case in corpus():
            document = case["document"].copy()
            query = document.pop("text")
            standalone = mapper.requirements_query(query, **document)
            assert standalone == case["expected"], case["id"]
            assert standalone == mapper.analyze_query(query, **document)["requirements"], case["id"]


def test_requirements_defaults_and_explicit_selectors():
    query = "search host=web"
    with SPLMapper(**mapper_kwargs()) as mapper:
        default = mapper.requirements_query(query, source_id="q.spl")
        assert default == mapper.requirements_query(
            query, language="spl", profile="splunkd", version="current", source_id="q.spl"
        )
        assert default == mapper.requirements_query(
            query, language="", profile="", version="", source_id="q.spl"
        )
        spl2 = mapper.requirements_query(
            "FROM main | where host='web'",
            language="spl2",
            profile="splunkd",
            version="current",
            source_id="q.spl2",
        )
    assert default["query"] == {
        "source_id": "q.spl",
        "language": "spl",
        "profile": "splunkd",
        "version": "current",
        "query_digest": "sha256:" + hashlib.sha256(query.encode()).hexdigest(),
    }
    assert (spl2["query"]["language"], spl2["query"]["profile"], spl2["query"]["version"]) == (
        "spl2", "splunkd", "current"
    )


@pytest.mark.parametrize("query", ["search host=\"😀\" | eval 'café'=host", "search host=\"a\x00b\""])
def test_requirements_preserve_non_ascii_and_nul_text(query):
    source_id = "é😀\x00.spl"
    with SPLMapper(**mapper_kwargs()) as mapper:
        requirements = mapper.requirements_query(query, source_id=source_id)
        analysis = mapper.analyze_query(query, source_id=source_id)
    assert requirements == analysis["requirements"]
    assert requirements["query"]["source_id"] == source_id
    assert requirements["query"]["query_digest"] == "sha256:" + hashlib.sha256(query.encode()).hexdigest()
    for item in requirements["items"]:
        for occurrence in item["occurrences"]:
            location = occurrence["location"]
            assert query.encode()[location["start"]["offset"]:location["end"]["offset"]].decode() == occurrence["original_name"]


@pytest.mark.parametrize(
    "options, message",
    [
        ({"language": "sql"}, "unsupported language"),
        ({"profile": "other"}, "unsupported profile"),
        ({"version": "9"}, "unsupported compatibility version"),
        ({"source_id": "\ud800"}, "unpaired UTF-16"),
    ],
)
def test_requirements_options_are_native_errors(options, message):
    with SPLMapper(**mapper_kwargs()) as mapper:
        with pytest.raises(SPLMapperError, match=message):
            mapper.requirements_query("search host=web", **options)
        assert mapper.requirements_query("search host=web")["query_status"] == "valid"


@pytest.mark.parametrize("query", ["\ud800", "\udc00"])
def test_requirements_reject_lone_python_surrogates(query):
    with SPLMapper(**mapper_kwargs()) as mapper:
        with pytest.raises(SPLMapperError, match="unpaired UTF-16"):
            mapper.requirements_query(query)


def _assert_resource_limited_requirements(requirements, query, language, source_id):
    code = "SPL_ANALYSIS_RESOURCE_LIMIT"
    assert requirements["schema_version"] == 1
    assert requirements["query"] == {
        "source_id": source_id,
        "language": language,
        "profile": "splunkd",
        "version": "current",
        "query_digest": "sha256:" + hashlib.sha256(query.encode()).hexdigest(),
    }
    assert requirements["query_status"] == "incomplete"
    assert requirements["coverage"] == {"complete": False, "reasons": [code]}
    assert requirements["items"] == []
    assert requirements["gaps"] == [{
        "code": code,
        "message": "requirement coverage is incomplete because analysis exceeded the 4,096-unit lexer work limit",
        "reference_ids": [],
        "diagnostic_codes": [code],
    }]
    assert len(requirements["diagnostics"]) == 1
    diagnostic = requirements["diagnostics"][0]
    assert diagnostic["code"] == code
    assert diagnostic["severity"] == "warning"
    assert diagnostic["category"] == "resource_limit"
    assert diagnostic["message"] == "analysis stopped before parser prediction after reaching the 4,096-unit lexer work limit"
    assert diagnostic["stage_id"] == diagnostic["scope_id"] == ""
    assert diagnostic["location"]["start"]["offset"] == 4096
    assert diagnostic["location"]["end"]["offset"] == 4097


@pytest.mark.parametrize("language, profile", [("spl", "splunkd"), ("spl2", "splunkd")])
def test_resource_limit_is_owned_success_for_native_and_python(language, profile, monkeypatch):
    query = "a " * 2048 + "a"
    source_id = "dense." + language
    document = {"text": query, "language": language, "profile": profile,
                "version": "current", "source_id": source_id}
    payload = json.dumps(document).encode()
    with SPLMapper(**mapper_kwargs()) as mapper:
        freed = []
        real_free = mapper._lib.spl_result_free

        def free(pointer):
            freed.append(bool(pointer))
            real_free(pointer)

        monkeypatch.setattr(mapper._lib, "spl_result_free", free)
        native_values = {}
        for operation in ("analyze_query", "requirements_query"):
            pointer = getattr(mapper._lib, "spl_mapper_" + operation)(mapper._mapper_id, payload)
            assert pointer and not pointer.contents.error and pointer.contents.result
            try:
                native_values[operation] = json.loads(pointer.contents.result.decode())
            finally:
                mapper._lib.spl_result_free(pointer)
        analysis = mapper.analyze_query(query, language=language, profile=profile, source_id=source_id)
        requirements = mapper.requirements_query(query, language=language, profile=profile, source_id=source_id)
        assert freed == [True, True, True, True]
        assert mapper._active_calls == 0

    assert native_values["analyze_query"] == analysis
    assert native_values["requirements_query"] == requirements
    assert analysis["status"] == "incomplete"
    assert analysis["document"] == document
    assert analysis["coverage"] == {
        "syntax_complete": False,
        "semantic_complete": False,
        "reasons": ["SPL_ANALYSIS_RESOURCE_LIMIT"],
    }
    assert analysis["requirements"] == requirements
    assert analysis["diagnostics"] == requirements["diagnostics"]
    assert analysis["stages"] == analysis["scopes"] == analysis["references"] == analysis["lineage"] == []
    assert all(values == [] for values in analysis["dependencies"].values())
    _assert_resource_limited_requirements(requirements, query, language, source_id)


def test_requirements_owned_result_and_operation_cleanup(monkeypatch):
    with SPLMapper(**mapper_kwargs()) as mapper:
        real_free = mapper._lib.spl_result_free

        def run_with_free_count(call, expected_frees):
            freed = []

            def free(pointer):
                freed.append(bool(pointer))
                real_free(pointer)

            with monkeypatch.context() as patcher:
                patcher.setattr(mapper._lib, "spl_result_free", free)
                call()
            assert freed == expected_frees
            assert mapper._active_calls == 0

        run_with_free_count(lambda: mapper.requirements_query("search host=web"), [True])

        def native_error():
            with pytest.raises(SPLMapperError, match="unsupported language"):
                mapper.requirements_query("search host=web", language="sql")

        run_with_free_count(native_error, [True])

        def decode_error():
            with monkeypatch.context() as patcher:
                patcher.setattr("spl_toolkit.mapper.json.loads", lambda _value: (_ for _ in ()).throw(ValueError("decode failed")))
                with pytest.raises(ValueError, match="decode failed"):
                    mapper.requirements_query("search host=web")

        run_with_free_count(decode_error, [True])

        native = mapper._lib.spl_mapper_requirements_query
        with monkeypatch.context() as patcher:
            patcher.setattr(mapper._lib, "spl_mapper_requirements_query", lambda *_args: (_ for _ in ()).throw(RuntimeError("native failed")))
            with pytest.raises(RuntimeError, match="native failed"):
                mapper.requirements_query("search host=web")
        assert mapper._active_calls == 0
        monkeypatch.setattr(mapper._lib, "spl_mapper_requirements_query", native)


def test_requirements_invalid_python_value_uses_operation_name():
    with SPLMapper(**mapper_kwargs()) as mapper:
        with pytest.raises(SPLMapperError, match="Invalid requirements request JSON"):
            mapper.requirements_query({"not a string"})
        assert mapper._active_calls == 0


def test_requirements_repeated_calls_return_fresh_values():
    case = corpus()[1]
    document = case["document"].copy()
    query = document.pop("text")
    with SPLMapper(**mapper_kwargs()) as mapper:
        first = mapper.requirements_query(query, **document)
        for _ in range(100):
            assert mapper.requirements_query(query, **document) == case["expected"]
        first["items"].clear()
        assert mapper.requirements_query(query, **document) == case["expected"]


def test_shared_native_requirements_are_deterministic_under_concurrency():
    case = corpus()[1]
    document = case["document"].copy()
    query = document.pop("text")
    with SPLMapper(**mapper_kwargs()) as mapper:
        def work(_):
            standalone = mapper.requirements_query(query, **document)
            assert standalone == mapper.analyze_query(query, **document)["requirements"]
            return standalone

        with ThreadPoolExecutor(max_workers=16) as pool:
            assert all(report == case["expected"] for report in pool.map(work, range(128)))


def test_close_waits_for_admitted_requirements_call(monkeypatch):
    mapper = SPLMapper(**mapper_kwargs())
    admitted = threading.Event()
    proceed = threading.Event()
    native = mapper._lib.spl_mapper_requirements_query

    def paused_native(*args):
        admitted.set()
        assert proceed.wait(timeout=5)
        return native(*args)

    monkeypatch.setattr(mapper._lib, "spl_mapper_requirements_query", paused_native)
    with ThreadPoolExecutor(max_workers=3) as pool:
        result = pool.submit(mapper.requirements_query, "search host=web")
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
                mapper.requirements_query("search host=web")
        finally:
            proceed.set()
        assert result.result(timeout=5)["schema_version"] == 1
        closing.result(timeout=5)
    mapper.close()


def test_requirements_invalid_and_closed_handles_return_owned_errors():
    mapper = SPLMapper(**mapper_kwargs())
    lib = mapper._lib
    old_handle = mapper._mapper_id
    mapper.close()
    with pytest.raises(MapperNotFoundError):
        mapper.requirements_query("search host=web")
    payload = b'{"text":"search host=web"}'
    for handle in (old_handle, -1, 2_147_483_647):
        pointer = lib.spl_mapper_requirements_query(handle, payload)
        try:
            assert pointer and pointer.contents.error == b"Mapper not found"
            assert pointer.contents.result is None
        finally:
            lib.spl_result_free(pointer)
    lib.spl_result_free(None)
