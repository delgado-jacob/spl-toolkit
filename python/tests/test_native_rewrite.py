"""Safe rewrite requests and result ownership through the real native library."""

from concurrent.futures import ThreadPoolExecutor
from copy import deepcopy
import ctypes
import hashlib
import json
import os
from pathlib import Path
import threading
import time

import pytest

from spl_toolkit import SPLMapper, SPLMapperError
from spl_toolkit.exceptions import MapperNotFoundError


FIXTURES = Path(os.environ["SPL_REWRITE_FIXTURES"])
assert FIXTURES.is_absolute(), "SPL_REWRITE_FIXTURES must be absolute"
CASES = json.loads((FIXTURES / "cases.json").read_text(encoding="utf-8"))
FORMS = json.loads((FIXTURES / "forms.json").read_text(encoding="utf-8"))["forms"]
RULES = [{"id": "map", "kind": "field", "source": {"name": "src"}, "target": {"name": "user"}}]


def mapper_kwargs():
    return {"library_path": os.environ["SPL_NATIVE_LIBRARY"]} if "SPL_NATIVE_LIBRARY" in os.environ else {}


def raw_result(mapper, operation, payload, *, handle=None, reject=False):
    pointer = getattr(mapper._lib, "spl_mapper_" + operation)(
        mapper._mapper_id if handle is None else handle, payload)
    try:
        assert pointer
        if reject:
            assert pointer.contents.error and pointer.contents.result is None
            return pointer.contents.error.decode("utf-8")
        assert not pointer.contents.error, pointer.contents.error
        return json.loads(pointer.contents.result)
    finally:
        mapper._lib.spl_result_free(pointer)


@pytest.mark.parametrize("operator", ["AND", "OR"])
@pytest.mark.parametrize("left,right,equal", [
    ("1e1000001", "1e1000001", True),
    ("1e1000001", "10e1000000", True),
    ("1e1000001", "2e1000001", False),
    ("1e-1000001", "1e-1000001", True),
    ("1e-1000001", "10e-1000002", True),
    ("1e-1000001", "1e-1000002", False),
    ("1e9223372036854775809", "1e9223372036854775809", True),
    ("1e9223372036854775809", "10e9223372036854775808", True),
    ("1e9223372036854775809", "1e9223372036854775810", False),
    ("1e-9223372036854775809", "1e-9223372036854775809", True),
    ("1e-9223372036854775809", "10e-9223372036854775810", True),
    ("1e-9223372036854775809", "2e-9223372036854775809", False),
])
def test_raw_rewrite_exact_large_exponent_guarantees(left, right, equal, operator):
    # The Python value API cannot express an arbitrary-precision numeric JSON
    # token. Exercise the owned C entrypoint directly, with no float conversion.
    query = f"FROM main | where n={left} {operator} n={right} | table src"
    rules = deepcopy(RULES)
    rules[0]["when"] = {"fact": "literal", "kind": "field", "identity": {"name": "n"},
                        "operator": "equals", "value": "__EXACT_NUMBER__"}
    document = {"text": query, "language": "spl2"}
    request = {"schema_version": 1, "mode": "apply", "document": document, "rules": rules}
    payload = json.dumps(request).replace('"__EXACT_NUMBER__"', left).encode()
    with SPLMapper(**mapper_kwargs()) as mapper:
        report = raw_result(mapper, "rewrite", payload)
        batch = {k: v for k, v in request.items() if k != "document"} | {"documents": [document]}
        batch_payload = json.dumps(batch).replace('"__EXACT_NUMBER__"', left).encode()
        assert raw_result(mapper, "rewrite_batch", batch_payload) == {
            "schema_version": 1, "status": "valid", "reports": [report]}
    expected = query.replace("table src", "table user") if equal else query
    assert report["status"] == "valid" and report["committed"] is equal
    assert report["candidate_text"] == report["text"] == expected
    assert len(report["rule_evaluations"]) == 1
    assert report["rule_evaluations"][0]["condition"] == {
        "state": "true" if equal else "false", "reason": "condition_true" if equal else "condition_false",
        "reference_ids": ["ref-1", "ref-2"], "children": []}


def call_request(mapper, request):
    request = deepcopy(request)
    document = request.pop("document")
    request.pop("schema_version")
    return mapper.rewrite(document.pop("text"), **document, **request)


def expected_search(mode, language="spl"):
    """Hand-derived report: SPL filter/SPL2 read, one source binding and edit."""
    def location(start, end):
        return {"start": {"offset": start, "line": 1, "column": start + 1},
                "end": {"offset": end, "line": 1, "column": end + 1}}

    analyses = []
    for name, text, end in (("src", "search src=x", 12), ("user", "search user=x", 13)):
        document = {"text": text, "language": language, "profile": "splunkd", "version": "current", "source_id": "é😀\x00.spl"}
        empty = {"fields": [], "removed": [], "open": True, "uncertain": False}
        source = {"fields": [{"name": name, "origin_reference_ids": ["ref-0"], "conditional": False}],
                  "removed": [], "open": True, "uncertain": False}
        reference_location = location(7, end - 2)
        requirements = {
            "schema_version": 1,
            "query": {
                "source_id": document["source_id"],
                "language": language,
                "profile": "splunkd",
                "version": "current",
                "query_digest": "sha256:" + hashlib.sha256(text.encode()).hexdigest(),
            },
            "capability_revision": {
                "spl": "sha256:117785ed5ff93fd72fa0030c56325ecf9daf50034eaed22c2e2f80bd22eae59a",
                "spl2": "sha256:f1391296cfbc616e9bb1b1828e2471e37b60a35c0555654c0734640e072a0437",
            }[language],
            "query_status": "valid",
            "coverage": {"complete": True, "reasons": []},
            "items": [{
                "id": "req-1", "kind": "field", "identity": name,
                "role": "filter" if language == "spl" else "read",
                "necessity": "required", "origin": "direct", "resolution": "exact",
                "occurrences": [{
                    "reference_id": "ref-0", "original_name": name, "binding": "source",
                    "stage_id": "stage-0", "scope_id": "scope-0", "location": reference_location,
                }],
            }],
            "gaps": [],
            "diagnostics": [],
        }
        analyses.append({
            "schema_version": 1, "document": document, "status": "valid",
            "coverage": {"syntax_complete": True, "semantic_complete": True, "reasons": []},
            "stages": [{"id": "stage-0", "command": "search", "position": 0, "scope_id": "scope-0",
                        "location": location(0, end), "semantic_complete": True}],
            "scopes": [{"id": "scope-0", "parent_id": "", "kind": "root", "stage_id": "", "location": location(0, end)}],
            "references": [{"id": "ref-0", "original_name": name, "normalized_name": name, "kind": "field",
                            "role": "filter" if language == "spl" else "read", "stage_id": "stage-0", "scope_id": "scope-0", "location": reference_location,
                            "resolution": "exact", "binding": "source", "origin_reference_ids": []}],
            "lineage": [{"stage_id": "stage-0", "scope_id": "scope-0", "before": empty, "after": source, "transitions": []}],
            "dependencies": {k: [] for k in ("indexes", "sources", "source_types", "datasets", "lookups", "data_models", "macros")},
            "diagnostics": [],
            "requirements": requirements,
        })
    committed = mode == "apply"
    return {
        "schema_version": 1, "document": analyses[0]["document"], "mode": mode, "status": "valid",
        "coverage": {"syntax_complete": True, "semantic_complete": True, "rewrite_complete": True, "reasons": []},
        "original_text": "search src=x", "candidate_text": "search user=x",
        "text": "search user=x" if committed else "search src=x", "committed": committed,
        "changes": [{"outcome": "applied", "reason": "matched", "group_id": "group-0", "rule_ids": ["map"],
                     "original_reference_ids": ["ref-0"], "candidate_reference_ids": ["ref-0"],
                     "original_location": location(7, 10), "candidate_location": location(7, 11),
                     "old_text": "src", "new_text": "user", "candidate_applied": True, "committed": committed}],
        "rule_evaluations": [{"rule_id": "map", "outcome": "proposed", "reason": "matched",
                              "reference_ids": ["ref-0"], "location": location(7, 10)}],
        "original_analysis": analyses[0], "candidate_analysis": analyses[1],
    }


@pytest.mark.parametrize("operation", ["rewrite", "rewrite_batch"])
def test_rewrite_native_operation_is_available(operation):
    with SPLMapper(**mapper_kwargs()) as mapper:
        assert callable(getattr(mapper, operation, None)), f"missing {operation} operation"
        assert callable(getattr(mapper._lib, "spl_mapper_" + operation, None))


@pytest.mark.parametrize("mode", ["preview", "apply"])
@pytest.mark.parametrize("language", ["spl", "spl2"])
def test_rewrite_complete_expected_report_and_owned_copy(mode, language):
    document = {"text": "search src=x", "source_id": "é😀\x00.spl", "language": language}
    with SPLMapper(**mapper_kwargs()) as mapper:
        expected = expected_search(mode, language)
        assert mapper.rewrite(document["text"], RULES, mode=mode, source_id=document["source_id"], language=language) == expected
        assert mapper.rewrite_batch([document], RULES, mode=mode) == {"schema_version": 1, "status": "valid", "reports": [expected]}
        result = mapper.rewrite_batch([document], RULES, mode=mode)
        result["reports"][0]["original_analysis"]["references"].clear()
        assert mapper.rewrite_batch([document], RULES, mode=mode)["reports"] == [expected]


@pytest.mark.parametrize("case", CASES, ids=lambda case: case["name"])
def test_rewrite_corpus_public_semantics_and_complete_c_python_parity(case):
    request = deepcopy(case["request"])
    before = deepcopy(request)
    with SPLMapper(**mapper_kwargs()) as mapper:
        report = call_request(mapper, request)
        assert report == raw_result(mapper, "rewrite", json.dumps(request).encode())
        batch = {k: v for k, v in request.items() if k != "document"} | {"documents": [request["document"]]}
        batch_args = {k: v for k, v in batch.items() if k != "schema_version"}
        assert mapper.rewrite_batch(**batch_args) == {"schema_version": 1, "status": report["status"], "reports": [report]}
        assert mapper.rewrite_batch(**batch_args) == raw_result(mapper, "rewrite_batch", json.dumps(batch).encode())
    assert request == before
    for key, expected in case["want"].items():
        if key.endswith("_complete"):
            assert report["coverage"][key] == expected
        elif key == "change_outcomes":
            assert [c["outcome"] for c in report["changes"]] == expected
        else:
            assert report[key] == expected
    for change in report["changes"]:
        for prefix, text_key in (("original", "original_text"), ("candidate", "candidate_text")):
            location = change.get(prefix + "_location")
            if location and change["candidate_applied"]:
                sliced = report[text_key].encode()[location["start"]["offset"]:location["end"]["offset"]].decode()
                assert sliced == change["old_text" if prefix == "original" else "new_text"]


@pytest.mark.parametrize("form", [form for form in FORMS if form["eligible"]], ids=lambda form: form["id"])
def test_native_rewrite_all_registered_rule_kinds(form):
    rules = [{"id": "form", "kind": form["kind"], "source": {"name": form["name"]}, "target": {"name": form["target"]}}]
    with SPLMapper(**mapper_kwargs()) as mapper:
        report = mapper.rewrite(form["query"], rules, mode="apply", language=form["language"])
    assert report["candidate_text"] == form["candidate"]
    assert report["text"] == form["candidate"] and report["committed"]


@pytest.mark.parametrize("operation", ["rewrite", "rewrite_batch"])
@pytest.mark.parametrize("payload", [None, b"null", b"[]", b"{", b"{} {}",
    b'{"schema_version":1,"schema_version":1,"document":{"text":""},"rules":[]}',
    b'{"schema_version":1,"document":{"text":"\\ud800"},"rules":[]}',
    b'{"schema_version":1,"document":{"text":"\xff"},"rules":[]}',
    b'{"schema_version":1,"mode":null,"document":{"text":""},"rules":[]}',
    b'{"schema_version":1,"documents":[],"rules":[]}'])
def test_raw_rewrite_malformed_requests_return_owned_error(operation, payload):
    with SPLMapper(**mapper_kwargs()) as mapper:
        raw_result(mapper, operation, payload, reject=True)
        assert mapper.rewrite("search src=x", RULES)["status"] == "valid"


@pytest.mark.parametrize("operation", ["rewrite", "rewrite_batch"])
def test_raw_rewrite_input_is_read_only_and_closed_handles_fail(operation):
    mapper = SPLMapper(**mapper_kwargs())
    request = {"schema_version": 1, "rules": RULES}
    document = {"text": "search src=x", "source_id": "é😀\x00.spl"}
    request.update({"document": document} if operation == "rewrite" else {"documents": [document]})
    payload = ctypes.create_string_buffer(json.dumps(request).encode())
    before = payload.raw
    assert raw_result(mapper, operation, payload)["status"] == "valid"
    assert payload.raw == before
    handle = mapper._mapper_id
    mapper.close()
    for invalid in (handle, -1, 2_147_483_647):
        assert raw_result(mapper, operation, payload, handle=invalid, reject=True) == "Mapper not found"
    args = (document["text"], RULES) if operation == "rewrite" else ([document], RULES)
    with pytest.raises(MapperNotFoundError):
        getattr(mapper, operation)(*args)
    mapper.close()


@pytest.mark.parametrize("operation", ["rewrite", "rewrite_batch"])
@pytest.mark.parametrize("value", [float("nan"), float("inf"), -float("inf"), object()])
def test_rewrite_invalid_json_rejected_before_native_allocation(operation, value, monkeypatch):
    with SPLMapper(**mapper_kwargs()) as mapper:
        def forbidden(*_args):
            pytest.fail("non-JSON input reached native code")
        monkeypatch.setattr(mapper._lib, "spl_mapper_" + operation, forbidden)
        args = ("search src=x", [value]) if operation == "rewrite" else ([{"text": "search src=x"}], [value])
        with pytest.raises(SPLMapperError, match="Invalid rewrite request JSON"):
            getattr(mapper, operation)(*args)
        assert mapper._active_calls == 0


@pytest.mark.parametrize("kwargs", [{"rules": None}, {"rules": [{}]}, {"rules": RULES + RULES},
    {"rules": [RULES[0] | {"context": {"EventCode": 1}}]}, {"mode": None}, {"mode": "unsafe"},
    {"validation_target": {}}, {"language": "sql"}, {"profile": "cloud"}, {"version": None},
    {"source_id": "\ud800"}, {"source_id": None}])
def test_rewrite_strict_request_failures_are_errors_not_query_reports(kwargs):
    with SPLMapper(**mapper_kwargs()) as mapper:
        with pytest.raises(SPLMapperError):
            mapper.rewrite("search src=x", **({"rules": RULES} | kwargs))
        assert mapper.rewrite("search src=x", RULES)["status"] == "valid"


@pytest.mark.parametrize("documents", [None, [], [{}], [{"text": None}], [{"text": "\ud800"}],
    [{"text": "search src=x", "unknown": True}]])
def test_rewrite_batch_strict_document_failures(documents):
    with SPLMapper(**mapper_kwargs()) as mapper:
        with pytest.raises(SPLMapperError):
            mapper.rewrite_batch(documents, RULES)


def test_rewrite_rules_and_original_facts_ignore_legacy_mapper_configuration():
    config = {"version": "1.0", "mappings": [{"source": "src", "target": "legacy"}]}
    with SPLMapper(config=config, **mapper_kwargs()) as mapper:
        assert mapper.map_query("search src=x") == "search legacy=x"
        assert mapper.rewrite("search src=x", RULES, source_id="é😀\x00.spl") == expected_search("preview")


@pytest.mark.parametrize("operation", ["rewrite", "rewrite_batch"])
def test_rewrite_null_result_and_circular_json_release_admission(operation, monkeypatch):
    with SPLMapper(**mapper_kwargs()) as mapper:
        args = ("search src=x", RULES) if operation == "rewrite" else ([{"text": "search src=x"}], RULES)
        monkeypatch.setattr(mapper._lib, "spl_mapper_" + operation, lambda *_: None)
        with pytest.raises(SPLMapperError, match="Native rewrite returned no result"):
            getattr(mapper, operation)(*args)
        circular = []
        circular.append(circular)
        with pytest.raises(SPLMapperError, match="Invalid rewrite request JSON"):
            getattr(mapper, operation)(args[0], circular)
        assert mapper._active_calls == 0


@pytest.mark.parametrize("operation", ["rewrite", "rewrite_batch"])
@pytest.mark.parametrize("outcome", ["success", "error", "decode"])
def test_rewrite_frees_every_real_owned_result(operation, outcome, monkeypatch):
    with SPLMapper(**mapper_kwargs()) as mapper:
        frees = []
        native_free = mapper._lib.spl_result_free
        def free(pointer):
            frees.append(bool(pointer.contents.error))
            native_free(pointer)
        monkeypatch.setattr(mapper._lib, "spl_result_free", free)
        args = ("search src=x", RULES) if operation == "rewrite" else ([{"text": "search src=x"}], RULES)
        if outcome == "decode":
            def fail_decode(_):
                raise ValueError("decode failed")
            monkeypatch.setattr("spl_toolkit.mapper.json.loads", fail_decode)
            with pytest.raises(ValueError, match="decode failed"):
                getattr(mapper, operation)(*args)
        elif outcome == "error":
            with pytest.raises(SPLMapperError):
                getattr(mapper, operation)(*args, mode="unsafe")
        else:
            assert getattr(mapper, operation)(*args)["status"] == "valid"
        assert frees == [outcome == "error"]
        assert mapper._active_calls == 0


def test_rewrite_ordered_batch_and_concurrent_calls():
    documents = [{"text": "search src=x"}, {"text": "FROM main SELECT src", "language": "spl2"}, {"text": "| mystery"}]
    with SPLMapper(**mapper_kwargs()) as mapper:
        singles = [mapper.rewrite(doc["text"], RULES, language=doc.get("language", "spl"), mode="apply") for doc in documents]
        expected = {"schema_version": 1, "status": "incomplete", "reports": singles}
        assert [r["status"] for r in singles] == ["valid", "valid", "incomplete"]
        def work(_):
            assert mapper.rewrite_batch(documents, RULES, mode="apply") == expected
        with ThreadPoolExecutor(max_workers=8) as pool:
            list(pool.map(work, range(64)))


@pytest.mark.parametrize("operation", ["rewrite", "rewrite_batch"])
def test_rewrite_close_waits_for_admitted_native_call(operation, monkeypatch):
    mapper = SPLMapper(**mapper_kwargs())
    admitted, proceed = threading.Event(), threading.Event()
    native = getattr(mapper._lib, "spl_mapper_" + operation)
    def paused(*args):
        admitted.set()
        assert proceed.wait(timeout=5)
        return native(*args)
    monkeypatch.setattr(mapper._lib, "spl_mapper_" + operation, paused)
    args = ("search src=x", RULES) if operation == "rewrite" else ([{"text": "search src=x"}], RULES)
    with ThreadPoolExecutor(max_workers=2) as pool:
        result = pool.submit(getattr(mapper, operation), *args)
        assert admitted.wait(timeout=5)
        closing = pool.submit(mapper.close)
        try:
            deadline = time.monotonic() + 5
            while not mapper._closing:
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
