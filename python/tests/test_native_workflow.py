"""Workflow delegates use the real owned-result ABI and detached Python values."""

import copy
import ctypes
import json
import os
import threading
from pathlib import Path
from concurrent.futures import ThreadPoolExecutor

import pytest

from spl_toolkit import SPLMapper, SPLMapperError
from spl_toolkit.exceptions import MapperNotFoundError

LIBRARY = Path(os.environ["SPL_NATIVE_LIBRARY"])
FIXTURE = Path(os.environ["SPL_WORKFLOW_SEED"])
assert LIBRARY.is_absolute() and LIBRARY.is_file()
assert FIXTURE.is_absolute() and FIXTURE.is_file()
SEED = json.loads(FIXTURE.read_text(encoding="utf-8"))
OPERATIONS = ("assess", "compare", "evidence", "recheck")


def open_mapper():
    return SPLMapper(library_path=str(LIBRARY))


def request():
    evidence = copy.deepcopy(SEED["compatibility"])
    return {"schema_version": 1, "documents": [{"id": "d1", "document": {
        "text": "from [{id:1}]", "language": "spl2", "profile": "splunkd",
        "version": "current", "source_id": "d1"}}], "settings": {
        "schema_version": 1, "snapshot": evidence["snapshot"],
        "schema_bundle": evidence["schema_bundle"], "entries": [{"id": "d1", "compatibility": {
            "query_scope": evidence["query_scope"], "input_bindings": []}}]}}


def requests(mapper):
    assess = request()
    report = mapper.workflow_assess(assess)
    return {"assess": assess, "compare": {"schema_version": 1, "before": report, "after": report},
            "evidence": {"schema_version": 1, "report": report, "include": []},
            "recheck": {"schema_version": 1, "context": {"original": assess["documents"][0],
                "snapshot": assess["settings"]["snapshot"], "schema_bundle": assess["settings"]["schema_bundle"]},
                "proposal": {"document": copy.deepcopy(assess["documents"][0]["document"]),
                             "settings": assess["settings"]["entries"][0]}}}


def raw_result(mapper, operation, raw, *, handle=None):
    buffer = ctypes.create_string_buffer(raw)
    before = buffer.raw
    pointer = getattr(mapper._lib, "spl_mapper_workflow_" + operation)(
        mapper._mapper_id if handle is None else handle, ctypes.cast(buffer, ctypes.c_char_p))
    assert pointer
    try:
        assert buffer.raw == before
        if pointer.contents.error:
            assert not pointer.contents.result
            return "error", pointer.contents.error.decode()
        assert pointer.contents.result
        return "report", json.loads(pointer.contents.result.decode())
    finally:
        mapper._lib.spl_result_free(pointer)


def test_all_workflow_delegates_match_raw_owned_results():
    with open_mapper() as mapper:
        inputs = requests(mapper)
        for operation, value in inputs.items():
            before = copy.deepcopy(value)
            result = getattr(mapper, "workflow_" + operation)(value)
            assert isinstance(result, dict)
            assert raw_result(mapper, operation, json.dumps(value).encode()) == ("report", result)
            assert value == before
        assert inputs["compare"]["before"]["ci_exit_code"] == 0
        evidence = mapper.workflow_evidence(inputs["evidence"])
        assert evidence["disclosure"]["requested"] == []
        assert evidence["disclosure"]["emitted"] == []
    # Values survive freeing their native result and mapper.
    assert result["assessment"]["ci_exit_code"] == 0


def test_text_and_mixed_content_are_values():
    with open_mapper() as mapper:
        value = request()
        value["format"] = "text"
        result = mapper.workflow_assess(value)
        assert isinstance(result, str) and result
        assert raw_result(mapper, "assess", json.dumps(value).encode()) == ("report", result)
        value.pop("format")
        bad = copy.deepcopy(value["documents"][0])
        bad["id"] = "bad"
        bad["document"]["text"] = "from [{id:1}] | eval broken ="
        value["documents"].append(bad)
        settings = copy.deepcopy(value["settings"]["entries"][0])
        settings["id"] = "bad"
        value["settings"]["entries"].append(settings)
        report = mapper.workflow_assess(value)
        assert report["ci_exit_code"] == 1
        assert report["execution_complete"]
        assert [e["status"] for e in report["entries"]] == ["valid", "invalid"]
        settings["compatibility"]["input_bindings"] = [{"input_id": "unknown", "object_id": "missing",
            "expected": {"kind": "dataset", "name": "missing", "namespace": "", "app": "", "owner": ""}}]
        bad["document"]["text"] = "from [{id:1}]"
        report = mapper.workflow_assess(value)
        assert report["ci_exit_code"] == 2 and not report["execution_complete"]
        assert report["counts"]["configuration_failed"] == 1
        assert report["entries"][0]["status"] == "valid"
        assert report["entries"][1]["failure"]["phase"] == "configuration"


def test_recheck_is_fresh_after_mutating_previous_report():
    with open_mapper() as mapper:
        inputs = requests(mapper)
        inputs["compare"]["before"]["ci_exit_code"] = 99
        inputs["compare"]["before"]["entries"][0]["analysis"]["document"]["text"] = "MUTATED"
        proposed = copy.deepcopy(inputs["assess"]["documents"][0]["document"])
        proposed["text"] = "from [{id:1}] | eval broken ="
        inputs["recheck"]["proposal"]["document"] = proposed
        report = mapper.workflow_recheck(inputs["recheck"])
        assert report["assessment"]["ci_exit_code"] == 1
        assert report["assessment"]["entries"][0]["analysis"]["document"]["text"] == proposed["text"]
        assert mapper.workflow_assess(inputs["assess"])["ci_exit_code"] == 0


@pytest.mark.parametrize("operation", OPERATIONS)
def test_global_errors_unicode_and_closed_mapper(operation):
    with open_mapper() as mapper:
        handle = mapper._mapper_id
        for raw in (b"{", b"null", b'{"id":"\\ud800"}', b'\xff'):
            kind, error = raw_result(mapper, operation, raw)
            assert kind == "error" and json.loads(error)["code"] == "request_invalid"
        with pytest.raises(SPLMapperError):
            getattr(mapper, "workflow_" + operation)({"schema_version": 2})
        with pytest.raises(SPLMapperError):
            getattr(mapper, "workflow_" + operation)({"id": "\ud800"})
        with pytest.raises(SPLMapperError):
            getattr(mapper, "workflow_" + operation)({"value": float("nan")})
    with pytest.raises(MapperNotFoundError):
        getattr(mapper, "workflow_" + operation)({})
    assert raw_result(mapper, operation, b"{}", handle=handle) == ("error", "Mapper not found")


def test_shared_mapper_concurrent_workflow_calls():
    with open_mapper() as mapper:
        inputs = requests(mapper)
        expected = {op: getattr(mapper, "workflow_" + op)(value) for op, value in inputs.items()}
        def work(index):
            operation = OPERATIONS[index % len(OPERATIONS)]
            return operation, getattr(mapper, "workflow_" + operation)(inputs[operation])
        with ThreadPoolExecutor(max_workers=8) as pool:
            values = list(pool.map(work, range(32)))
        assert all(value == expected[operation] for operation, value in values)


def test_workflow_calls_race_with_close():
    # Use the same lifecycle synchronization pattern as the native mapper tests.
    mapper = open_mapper()
    inputs = requests(mapper)
    start = threading.Event()
    first_call = threading.Event()

    def operate(index):
        assert start.wait(timeout=5)
        operation = OPERATIONS[index % len(OPERATIONS)]
        while True:
            try:
                getattr(mapper, "workflow_" + operation)(inputs[operation])
                first_call.set()
            except MapperNotFoundError:
                return

    try:
        with ThreadPoolExecutor(max_workers=8) as pool:
            workers = [pool.submit(operate, index) for index in range(8)]
            start.set()
            assert first_call.wait(timeout=5)
            mapper.close()
            for worker in workers:
                assert worker.result(timeout=10) is None
    finally:
        mapper.close()


@pytest.mark.parametrize("format", ["json", "sarif", "graph", "bom"])
def test_assessment_export_objects_and_unicode(format):
    with open_mapper() as mapper:
        value = request()
        value["format"] = format
        value["documents"][0]["document"]["text"] = 'from [{id:1}] | eval label="雪é"'
        result = mapper.workflow_assess(value)
        assert isinstance(result, dict)
        assert raw_result(mapper, "assess", json.dumps(value, ensure_ascii=False).encode()) == ("report", result)
