"""Frozen Go schema reports through real CLI, HTTP and installed native Python.

SPL_SCHEMA_EVIDENCE optionally retains full synthetic inputs/results. Catalogs are
identified by their committed relative fixture and byte hashes, never truncated.
"""
from copy import deepcopy
import gzip
import hashlib
import json
import os
from pathlib import Path
import subprocess
import time
from urllib.error import HTTPError
from urllib.request import Request, urlopen

import pytest

from spl_toolkit import SPLMapper, SPLMapperError
from test_surfaces import cli_path, server_url, post_json, required_absolute_path


FIXTURES = required_absolute_path("SPL_SCHEMA_FIXTURES")
CORPUS = json.loads((FIXTURES / "cases.json").read_text(encoding="utf-8"))
REQUESTS = json.loads((FIXTURES / "requests.json").read_text(encoding="utf-8"))
EXITS = {"valid": 0, "invalid": 1, "incomplete": 3}


def open_mapper():
    options = {"library_path": str(required_absolute_path("SPL_NATIVE_LIBRARY"))} if "SPL_NATIVE_LIBRARY" in os.environ else {}
    return SPLMapper(**options)


def digest(raw):
    return hashlib.sha256(raw).hexdigest()


def load_target(case):
    target = deepcopy(case["target"])
    if "catalog_fixture" in case:
        path = FIXTURES / case["catalog_fixture"]
        raw = gzip.decompress(path.read_bytes()) if path.suffix == ".gz" else path.read_bytes()
        target["catalog"] = json.loads(raw)
    return target


def cli_arguments(cli, target, directory, case=None):
    args = [str(cli), "validate-schema", "--format", "json"]
    for key, flag in (("schema", "--schema"), ("resources", "--schema-resources"), ("catalog", "--ocsf-catalog")):
        if key in target:
            path = directory / (key + ".json")
            if key == "catalog" and case and "catalog_fixture" in case:
                fixture = FIXTURES / case["catalog_fixture"]
                path.write_bytes(gzip.decompress(fixture.read_bytes()) if fixture.suffix == ".gz" else fixture.read_bytes())
            else:
                path.write_text(json.dumps(target[key]), encoding="utf-8")
            args += [flag, str(path)]
    if "base_uri" in target:
        args += ["--schema-base-uri", target["base_uri"]]
    if "selection" in target:
        selection = target["selection"]
        args += ["--ocsf-version", selection["version"]]
        for key in ("class", "category", "class_uid", "category_uid"):
            if key in selection:
                args += ["--ocsf-" + key.removesuffix("_uid"), str(selection[key])]
        for key in ("profiles", "extensions"):
            for name in selection.get(key, []):
                args += ["--ocsf-" + key[:-1], name]
    return args


@pytest.fixture(scope="session")
def evidence(cli_path):
    with open_mapper() as mapper:
        library = Path(mapper._lib._name).resolve()
    record = {"schema_version": 1, "fixture_root": str(FIXTURES),
              "fixture_hashes": {n: digest((FIXTURES / n).read_bytes()) for n in ("cases.json", "requests.json")},
              "artifacts": {str(p): digest(p.read_bytes()) for p in (cli_path, required_absolute_path("SPL_SERVER"), library)},
              "runtime_proxies": {k: os.environ.get(k) for k in ("HTTP_PROXY", "HTTPS_PROXY", "ALL_PROXY", "NO_PROXY")},
              "cases": [], "batches": [], "request_errors": [], "cli_errors": []}
    yield record
    if destination := os.environ.get("SPL_SCHEMA_EVIDENCE"):
        Path(destination).write_text(json.dumps(record, indent=2, sort_keys=True) + "\n", encoding="utf-8")


@pytest.mark.parametrize("case", CORPUS["cases"], ids=lambda case: case["id"])
def test_schema_full_report_parity(case, cli_path, server_url, tmp_path, evidence):
    assert CORPUS["schema_version"] == 1
    target, document, expected = load_target(case), case["document"], case["expected"]
    args = cli_arguments(cli_path, target, tmp_path, case)
    for key, flag in (("language", "--language"), ("profile", "--profile"), ("version", "--compatibility-version"), ("source_id", "--source-id")):
        if key in document:
            args += [flag, document[key]]
    start = time.monotonic()
    completed = subprocess.run(args + ["--query", document["text"]], capture_output=True, text=True, timeout=10)
    assert (completed.returncode, completed.stderr) == (EXITS[expected["status"]], ""), case["id"]
    cli_report = json.loads(completed.stdout)
    with open_mapper() as mapper:
        native = mapper.validate_schema(document["text"], target, **{k: v for k, v in document.items() if k != "text"})
    status, rest = post_json(server_url, "/query/validate-schema", {"document": document, "target": target})
    elapsed = time.monotonic() - start
    assert status == 200
    assert cli_report == native == rest == expected, case["id"]
    if case["id"] == "unresolved_ref":
        assert elapsed < 5, "unresolved local reference must not wait on retrieval"
    item = {"id": case["id"], "document": document, "target": case["target"],
            "target_json_sha256": digest(json.dumps(target, sort_keys=True).encode()),
            "oracle": {k: case[k] for k in ("status", "schema_complete", "outcomes") if k in case},
            "expected": expected, "cli": {"exit": completed.returncode, "report": cli_report},
            "native": native, "rest": {"http_status": status, "report": rest}, "elapsed_seconds": elapsed}
    if "catalog_fixture" in case:
        path = FIXTURES / case["catalog_fixture"]
        raw = gzip.decompress(path.read_bytes()) if path.suffix == ".gz" else path.read_bytes()
        item.update(catalog_fixture=case["catalog_fixture"], catalog_file_sha256=digest(path.read_bytes()), catalog_raw_sha256=digest(raw))
    evidence["cases"].append(item)


def test_schema_ordered_batches_full_reports(cli_path, server_url, tmp_path, evidence):
    groups = {}
    for case in CORPUS["cases"]:
        key = json.dumps([case["target"], case.get("catalog_fixture")], sort_keys=True)
        groups.setdefault(key, []).append(case)
    assert any(len(group) > 1 for group in groups.values())
    for group in groups.values():
        group = list(reversed(group))
        documents = [case["document"] for case in group]
        target = load_target(group[0])
        reports = [case["expected"] for case in group]
        status = "invalid" if any(r["status"] == "invalid" for r in reports) else "incomplete" if any(r["status"] == "incomplete" for r in reports) else "valid"
        expected = {"schema_version": 1, "status": status, "reports": reports}
        args = cli_arguments(cli_path, target, tmp_path, group[0])
        completed = subprocess.run(args + ["--batch", "-"], input=json.dumps(documents), capture_output=True, text=True, timeout=10)
        assert (completed.returncode, completed.stderr) == (EXITS[status], "")
        with open_mapper() as mapper:
            native = mapper.validate_schema_batch(documents, target)
        http_status, rest = post_json(server_url, "/query/validate-schema/batch", {"documents": documents, "target": target})
        cli_report = json.loads(completed.stdout)
        assert http_status == 200
        assert cli_report == native == rest == expected
        evidence["batches"].append({"case_ids": [c["id"] for c in group], "documents": documents, "expected": expected,
                                    "cli": {"exit": completed.returncode, "report": cli_report}, "native": native,
                                    "rest": {"http_status": http_status, "report": rest}})


@pytest.mark.parametrize("case", REQUESTS["cases"], ids=lambda case: case["id"])
def test_schema_raw_request_rejection(case, server_url, evidence):
    assert REQUESTS["schema_version"] == 1
    payload = bytes.fromhex(case["hex"]) if "hex" in case else case["input"].encode("utf-8")
    route = "/query/validate-schema" + ("/batch" if case["batch"] else "")
    request = Request(server_url + route, data=payload, headers={"Content-Type": "application/json"}, method="POST")
    with pytest.raises(HTTPError) as caught:
        urlopen(request, timeout=5)
    with caught.value as response:
        assert response.code == 400
        rest = json.load(response)
    assert rest["error"] is True and "reports" not in rest and "analysis" not in rest
    with open_mapper() as mapper:
        native = mapper._lib.spl_mapper_validate_schema_batch if case["batch"] else mapper._lib.spl_mapper_validate_schema
        pointer = native(mapper._mapper_id, payload)
        try:
            assert pointer and pointer.contents.error and pointer.contents.result is None
            error = pointer.contents.error.decode()
        finally:
            mapper._lib.spl_result_free(pointer)
    evidence["request_errors"].append({"id": case["id"], "batch": case["batch"], "input_hex": payload.hex(),
                                       "http_status": 400, "rest": rest, "native_error": error})


@pytest.mark.parametrize("schema", [None, {"$schema": "https://json-schema.org/draft-07/schema"}])
def test_schema_cli_target_errors(schema, cli_path, tmp_path, evidence):
    target = {"kind": "json_schema", "schema": schema}
    completed = subprocess.run(cli_arguments(cli_path, target, tmp_path) + ["--query", "table host"], capture_output=True, text=True, timeout=5)
    assert completed.returncode == 2 and completed.stdout == "" and completed.stderr
    with open_mapper() as mapper, pytest.raises(SPLMapperError):
        mapper.validate_schema("table host", target)
    evidence["cli_errors"].append({"target": target, "query": "table host", "exit": 2, "stdout": completed.stdout, "stderr": completed.stderr})


def test_schema_cli_failed_batch_is_atomic(cli_path, tmp_path, evidence):
    target = {"kind": "json_schema", "schema": True}
    documents = [{"text": "table host"}, {"text": "table host", "profile": "unsupported"}]
    completed = subprocess.run(cli_arguments(cli_path, target, tmp_path) + ["--batch", "-"], input=json.dumps(documents), capture_output=True, text=True, timeout=5)
    assert completed.returncode == 2 and completed.stdout == "" and completed.stderr
    evidence["cli_errors"].append({"target": target, "documents": documents, "exit": 2, "stdout": completed.stdout, "stderr": completed.stderr})


def test_schema_real_file_activity_and_generic_descendants(cli_path, server_url, tmp_path, evidence):
    case = {"target": {"kind": "ocsf", "selection": {"version": "1.6.0", "class": "file_activity", "profiles": [], "extensions": []}},
            "catalog_fixture": "ocsf/1.6.0/base.json.gz"}
    target = load_target(case)
    document = {"text": "table file.name file.xattributes.vendor_field", "source_id": "file-activity.spl"}
    with open_mapper() as mapper:
        native = mapper.validate_schema(document["text"], target, source_id=document["source_id"])
    # Independent field facts from the genuine catalog, before transport equality.
    assert native["status"] == "valid"
    assert [o["outcome"] for o in native["outcomes"]] == ["required", "permitted_unspecified"]
    assert len(native["outcomes"][0]["evidence"]) >= 2
    args = cli_arguments(cli_path, target, tmp_path, case)
    completed = subprocess.run(args + ["--query", document["text"], "--source-id", document["source_id"]], capture_output=True, text=True, timeout=10)
    assert (completed.returncode, completed.stderr) == (0, "")
    status, rest = post_json(server_url, "/query/validate-schema", {"document": document, "target": target})
    assert status == 200 and json.loads(completed.stdout) == rest == native
    evidence["practical_file_activity"] = {"document": document, **case, "oracle_outcomes": ["required", "permitted_unspecified"],
                                           "cli": {"exit": 0, "report": json.loads(completed.stdout)}, "native": native,
                                           "rest": {"http_status": status, "report": rest}}
