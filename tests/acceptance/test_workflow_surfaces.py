"""Public, authored workflow semantics across five actual product boundaries."""
import copy
import ctypes
import hashlib
import json
import os
from pathlib import Path
import shutil
import subprocess
from urllib.error import HTTPError
from urllib.request import Request, urlopen

import pytest
from spl_toolkit import SPLMapper, SPLMapperError
from test_surfaces import cli_path, required_absolute_path, server_url

FIXTURES = required_absolute_path("SPL_WORKFLOW_FIXTURES")
LIBRARY = required_absolute_path("SPL_NATIVE_LIBRARY")
CASES = json.loads((FIXTURES / "cases.json").read_text())
COMPARISONS = json.loads((FIXTURES / "comparisons.json").read_text())
DISCLOSURE = json.loads((FIXTURES / "disclosure.json").read_text())
SURFACES = ["go", "cli", "http", "c", "python"]
GO_DRIVER = r'''
package main
import("encoding/json";"io";"os";"github.com/delgado-jacob/spl-toolkit/pkg/workflow")
func main(){ raw,e:=io.ReadAll(os.Stdin);if e!=nil{panic(e)};var value any
switch os.Args[1]{case "assess":value,e=workflow.AssessOutputJSON(raw)
case "compare":value,e=workflow.CompareJSON(raw)
case "evidence":value,e=workflow.EvidenceJSON(raw)
case "recheck":value,e=workflow.RecheckJSON(raw)}
if e!=nil {d,ok:=workflow.RequestErrorDetails(e);if !ok{panic(e)};value=d}
if e=json.NewEncoder(os.Stdout).Encode(value);e!=nil{panic(e)}}
'''

@pytest.fixture(scope="module")
def workflow_go_driver(tmp_path_factory):
    owner = required_absolute_path("SPL_TOOLING_SOURCE_ROOT")
    with SPLMapper(library_path=str(LIBRARY)) as mapper:
        version = mapper.native_version
    directory = tmp_path_factory.mktemp("workflow-go")
    try:
        helper = directory / "main.go"
        helper.write_text(GO_DRIVER)
        binary = directory / ("workflow.exe" if os.name == "nt" else "workflow")
        env = dict(os.environ, GOWORK="off")
        subprocess.run([os.environ.get("SPL_TOOLING_GO", "go"), "build", "-mod=readonly", "-buildvcs=false",
                        "-ldflags=-X=github.com/delgado-jacob/spl-toolkit/internal/buildinfo.Version=" + version,
                        "-o", str(binary), str(helper)], cwd=owner, env=env, check=True, capture_output=True)
        yield binary, owner, env
    finally:
        shutil.rmtree(directory)

@pytest.fixture(scope="module")
def workflow_evidence():
    record = {"schema_version": 1, "source_sha": os.environ["SPL_WORKFLOW_SOURCE_SHA"],
              "fixture_hashes": {p.name: hashlib.sha256(p.read_bytes()).hexdigest()
                                 for p in sorted(FIXTURES.glob("*.json"))},
              "operations": {op: 0 for op in ("assess", "compare", "evidence", "recheck")},
              "controls": {}, "surfaces": []}
    yield record
    destination = Path(os.environ["SPL_WORKFLOW_EVIDENCE"])
    assert destination.is_absolute() and not destination.exists()
    destination.write_text(json.dumps(record, indent=2, sort_keys=True) + "\n")


def make_request(case):
    q = copy.deepcopy(case["request"])
    if case.get("selectors"):
        doc = q["documents"][0]["document"]
        with SPLMapper(library_path=str(LIBRARY)) as mapper:
            requirements = mapper.requirements_query(doc["text"], **{k: doc[k] for k in ("language", "profile", "version", "source_id")})
        for selector in case["selectors"]:
            matches = [i for i in requirements["inputs"] if i["kind"] == selector["kind"] and i["name"] == selector["name"]]
            assert len(matches) == 1
            binding = copy.deepcopy(selector["binding"]); binding["input_id"] = matches[0]["id"]
            q["settings"]["entries"][0]["compatibility"]["input_bindings"].append(binding)
    return q


def assess(case):
    with SPLMapper(library_path=str(LIBRARY)) as mapper:
        value = mapper.workflow_assess(make_request(case))
    assert_assessment(value, case["expected"])
    return value


def assert_assessment(value, expected):
    assert value["ci_exit_code"] == expected["ci_exit_code"]
    assert value["execution_complete"] == expected["execution_complete"]
    assert [entry["status"] for entry in value["entries"]] == expected["statuses"]
    assert value["counts"]["selected"] == len(expected["statuses"])
    assert sum(value["counts"][s] for s in ("valid", "invalid", "incomplete")) == len(expected["statuses"])


def pointer(value, path):
    for token in path.strip("/").split("/"):
        token = token.replace("~1", "/").replace("~0", "~")
        value = value[int(token)] if isinstance(value, list) else value[token]
    return value


def exercise(operation, request, authored, code, cli_path, server_url, driver, tmp_path, evidence, error=False):
    before = copy.deepcopy(request)
    raw = json.dumps(request, ensure_ascii=False).encode()
    binary, owner, env = driver
    go = json.loads(subprocess.run([str(binary), operation], input=raw, cwd=owner, env=env,
                                  check=True, capture_output=True, timeout=30).stdout)
    path = tmp_path / "request.json"; path.write_bytes(raw)
    args = [str(cli_path), "workflow", operation]
    if operation in ("assess", "recheck"):
        args += ["--request", str(path), "--format", request.get("format", "json")]
    elif operation == "compare":
        for key in ("before", "after"):
            f = tmp_path / (key + ".json"); f.write_text(json.dumps(request[key]))
            args += ["--" + key, str(f)]
        args += ["--format", "json"]
    else:
        key = "report" if "report" in request else "comparison"
        f = tmp_path / "saved.json"; f.write_text(json.dumps(request[key]))
        args += ["--" + key, str(f)]
        for include in request["include"]: args += ["--include", include]
    cli = subprocess.run(args, capture_output=True, timeout=30)
    assert cli.returncode == code, cli.stderr
    assert (cli.stdout if error else cli.stderr) == b""
    output = cli.stderr.removeprefix(b"Error: ") if error else cli.stdout
    cli_value = output.decode() if request.get("format") == "text" and not error else json.loads(output)
    req = Request(server_url + "/workflow/" + operation, data=raw, headers={"Content-Type": "application/json"})
    try: response = urlopen(req, timeout=30)
    except HTTPError as failed: response = failed
    with response:
        assert response.status == (400 if error else 200)
        http = response.read().decode() if request.get("format") == "text" and not error else json.load(response)
    with SPLMapper(library_path=str(LIBRARY)) as mapper:
        buffer = ctypes.create_string_buffer(raw); original = buffer.raw
        owned = getattr(mapper._lib, "spl_mapper_workflow_" + operation)(mapper._mapper_id, ctypes.cast(buffer, ctypes.c_char_p))
        assert owned
        try:
            assert buffer.raw == original
            assert bool(owned.contents.error) == error
            c = json.loads(owned.contents.error if error else owned.contents.result)
        finally: mapper._lib.spl_result_free(owned)
        if error:
            with pytest.raises(SPLMapperError) as raised: getattr(mapper, "workflow_" + operation)(request)
            python = json.loads(str(raised.value))
        else: python = getattr(mapper, "workflow_" + operation)(request)
    for value in (go, cli_value, http, c, python): authored(value)
    assert go == cli_value == http == c == python
    assert request == before and path.read_bytes() == raw
    # Detached values remain usable after owned buffers and mapper are freed.
    assert copy.deepcopy(python) == python
    evidence["operations"][operation] += 1
    evidence["surfaces"] = SURFACES.copy()
    evidence["controls"].update(source_immutability="passed", native_detachment="passed")
    return python


@pytest.mark.parametrize("case", CASES, ids=lambda c: c["id"])
def test_workflow_assessment(case, cli_path, server_url, workflow_go_driver, tmp_path, workflow_evidence):
    report = exercise("assess", make_request(case), lambda v: assert_assessment(v, case["expected"]),
                      case["expected"]["ci_exit_code"], cli_path, server_url, workflow_go_driver, tmp_path, workflow_evidence)
    if case["id"] == "resolution-siblings":
        entry = report["entries"][1]
        variants = entry["resolution"]["variants"]
        assert [v["outcome"] for v in variants] == ["verified", "failed"]
        assert "resolved_query" in variants[0] and "resolved_query" not in variants[1]
        assert entry["analysis"]["document"] == entry["resolution"]["original"]["analysis"]["document"]
        for variant in variants:
            for change in variant["changes"]:
                loc=change["original_location"]
                assert entry["analysis"]["document"]["text"].encode()[loc["start"]["offset"]:loc["end"]["offset"]].decode() == change["before"]
                if "candidate_location" in change:
                    candidate = change["candidate_location"]
                    assert variant["candidate_text"].encode()[candidate["start"]["offset"]:candidate["end"]["offset"]].decode() == change["after"]
    if case["id"] == "role-isolation":
        e = report["entries"][0]
        assert e["compatibility"]["requirements"] == e["analysis"]["requirements"]
        assert e["compatibility"]["outcome"] == "satisfied"
        assert len(e["compatibility"]["inputs"]) == 3
    if case["id"] == "role-missing-field":
        c = report["entries"][0]["compatibility"]
        assert c["outcome"] == "unsatisfied"
        assert c["requirements"] == report["entries"][0]["analysis"]["requirements"]
        workflow_evidence["controls"]["role_isolation"] = "passed"
    if case["id"] == "closure-implicit-source":
        closure = report["entries"][0]["compatibility"]["closure"]
        assert closure["effective_analysis"]["document"]["text"] == "eval local=1"
        assert report["entries"][0]["compatibility"]["outcome"] == "incomplete"
    if case["id"].startswith("resolution-limit"):
        assert report["entries"][1]["failure"]["phase"] == "configuration"
        assert "resolution" not in report["entries"][1]


@pytest.mark.parametrize("case", [c for c in COMPARISONS if "assessment" in c], ids=lambda c: c["id"])
def test_workflow_comparison(case, cli_path, server_url, workflow_go_driver, tmp_path, workflow_evidence):
    saved = assess(next(c for c in CASES if c["id"] == case["assessment"]))
    def authored(v):
        assert v["ci_exit_code"] == case["expected"]["ci_exit_code"]
        assert v["entries"][-1]["classification"] == case["expected"]["classification"]
    exercise("compare", {"schema_version": 1, "before": saved, "after": saved}, authored,
             case["expected"]["ci_exit_code"], cli_path, server_url, workflow_go_driver, tmp_path, workflow_evidence)


@pytest.mark.parametrize("includes", [[], ["query_text"], ["requirement_names"], ["source_identity"],
                                           ["environment_metadata"], ["diagnostic_details"], ["artifact_identity"]])
def test_workflow_disclosure(includes, cli_path, server_url, workflow_go_driver, tmp_path, workflow_evidence):
    saved = assess(next(c for c in CASES if c["id"] == "disclosure-retained-values"))
    e = saved["entries"][0]; c = e["compatibility"]
    e["origin"] = {"kind": "file", "relative_path": DISCLOSURE["source_identity"] + ".spl", "base_uri": "https://" + DISCLOSURE["source_identity"]}
    saved["provenance"]["environment_digest"] = DISCLOSURE["artifact_identity"]
    saved["provenance"]["schema_bundle_digest"] = DISCLOSURE["artifact_identity"] + "_SCHEMA"
    c["provenance"]["environment_digest"] = saved["provenance"]["environment_digest"]
    c["provenance"]["schema_bundle_digest"] = saved["provenance"]["schema_bundle_digest"]
    c["diagnostics"].append({"code": DISCLOSURE["unknown_code"], "severity": "warning", "artifact": "snapshot",
                             "path": DISCLOSURE["source_identity"], "message": DISCLOSURE["diagnostic_details"]})
    c["reasons"].append({"code": DISCLOSURE["unknown_code"], "message": DISCLOSURE["diagnostic_details"],
                         "locations": [], "reference_ids": [], "candidate_input_ids": []})
    for group in (c["inputs"], c["requirement_outcomes"]):
        for item in group:
            for match in item.get("objects", []):
                if match.get("object"):
                    match["object"]["owner"] = DISCLOSURE["environment_metadata"]
    def authored(v):
        assert v["disclosure"]["requested"] == includes
        raw = json.dumps(v, ensure_ascii=False)
        if not includes:
            assert all(marker not in raw for marker in DISCLOSURE.values())
            assert all("details" not in item for item in v["items"])
            assert len(v["disclosure"]["omitted"]) == 7
        else:
            assert DISCLOSURE[includes[0]] in raw
            for category in ("query_text", "requirement_names", "source_identity", "environment_metadata", "artifact_identity"):
                if category != includes[0] and not (includes == ["query_text"] and category == "requirement_names"):
                    assert DISCLOSURE[category] not in raw
        assert all("PRIVATE" not in item["pointer"] and "PRIVATE" not in item["token"] for item in v["items"])
    exercise("evidence", {"schema_version": 1, "report": saved, "include": includes}, authored, 0,
             cli_path, server_url, workflow_go_driver, tmp_path, workflow_evidence)


def test_workflow_definition_disclosure(cli_path, server_url, workflow_go_driver, tmp_path, workflow_evidence):
    q = copy.deepcopy(next(c for c in CASES if c["id"] == "closure-implicit-source"))
    q["request"]["settings"]["snapshot"]["objects"][-1]["document"]["text"] = 'eval note="' + DISCLOSURE["definitions"] + '"'
    saved = assess(q)
    assert DISCLOSURE["definitions"] in saved["entries"][0]["compatibility"]["closure"]["effective_analysis"]["document"]["text"]
    for include in ([], ["query_text"], ["environment_metadata"], ["diagnostic_details"], ["definitions"]):
        def authored(v):
            assert (DISCLOSURE["definitions"] in json.dumps(v)) == (include == ["definitions"])
        exercise("evidence", {"schema_version": 1, "report": saved, "include": include}, authored, 0,
                 cli_path, server_url, workflow_go_driver, tmp_path, workflow_evidence)
    workflow_evidence["controls"]["disclosure"] = "passed"


def test_workflow_recheck(cli_path, server_url, workflow_go_driver, tmp_path, workflow_evidence):
    q = CASES[0]["request"]
    proposed = copy.deepcopy(q["documents"][0]["document"]); proposed["text"] = 'from [{id:1}] | eval broken ='
    request = {"schema_version": 1, "context": {"original": q["documents"][0], "snapshot": q["settings"]["snapshot"],
               "schema_bundle": q["settings"]["schema_bundle"]}, "proposal": {"document": proposed, "settings": q["settings"]["entries"][0]}}
    def authored(v):
        assert v["assessment"]["ci_exit_code"] == 1
        assert v["assessment"]["entries"][0]["analysis"]["document"] == proposed
        assert v["original_source_hash"] != v["proposed_source_hash"]
    exercise("recheck", request, authored, 1, cli_path, server_url, workflow_go_driver, tmp_path, workflow_evidence)


@pytest.mark.parametrize("format", ["json", "text", "sarif", "graph", "bom"])
def test_workflow_exports(format, cli_path, server_url, workflow_go_driver, tmp_path, workflow_evidence):
    q = copy.deepcopy(CASES[1]["request"]); q["format"] = format
    source = assess(CASES[1])
    def authored(v):
        if format == "json": assert_assessment(v, CASES[1]["expected"])
        elif format == "text": assert 'portable' in v and 'complete' in v
        elif format == "sarif":
            from jsonschema import Draft7Validator
            owner = required_absolute_path("SPL_TOOLING_SOURCE_ROOT")
            schema = json.loads((owner / "contracts/sarif/sarif-schema-2.1.0.json").read_text())
            Draft7Validator(schema).validate(v)
            candidates = [r for r in v["runs"][0]["results"] if r.get("properties", {}).get("domain") == "candidate"]
            assert candidates and all(not r.get("locations") for r in candidates)
        else:
            assert len(v["subjects"]) == 4
            for subject in v["subjects"]: assert pointer(source, subject["evidence_pointer"])
            if format == "bom":
                for shared in v["shared_dependencies"]:
                    assert shared["occurrence_pointers"]
                    for path in shared["occurrence_pointers"]: assert pointer(source, path)
    exercise("assess", q, authored, 1, cli_path, server_url, workflow_go_driver, tmp_path, workflow_evidence)
    if format == "sarif": workflow_evidence["controls"]["locations"] = "passed"
    if format == "graph": workflow_evidence["controls"]["graph_pointers"] = "passed"
    if format == "bom": workflow_evidence["controls"]["bom_occurrences"] = "passed"


@pytest.mark.parametrize("operation", ["assess", "compare", "evidence", "recheck"])
def test_workflow_errors(operation, cli_path, server_url, workflow_go_driver, tmp_path, workflow_evidence):
    # CLI compare/evidence construct their request envelope from saved values.
    if operation in ("compare", "evidence"):
        q = {"schema_version": 1, "before": {}, "after": {}} if operation == "compare" else {"schema_version": 1, "report": {}, "include": []}
    else: q = {"schema_version": 2}
    def authored(v): assert v["code"] == "request_invalid" and v["message"]
    exercise(operation, q, authored, 2, cli_path, server_url, workflow_go_driver, tmp_path, workflow_evidence, error=True)


def test_acquisition_failure_retained_across_saved_operations(cli_path, server_url, workflow_go_driver, tmp_path, workflow_evidence):
    case = next(c for c in COMPARISONS if c["id"] == "acquisition-failure")
    manifest = case["manifest"]
    path = tmp_path / "manifest.json"; path.write_text(json.dumps(manifest))
    settings = tmp_path / "settings.json"; settings.write_text(json.dumps(CASES[0]["request"]["settings"]))
    before = (path.read_bytes(), settings.read_bytes())
    completed = subprocess.run([str(cli_path), "workflow", "assess", "--manifest", str(path),
                                "--settings", str(settings), "--format", "json"], capture_output=True, timeout=30)
    assert completed.returncode == 2 and not completed.stderr
    saved = json.loads(completed.stdout)
    assert saved["counts"]["acquisition_failed"] == case["expected"]["acquisition_failed"]
    assert saved["counts"]["assessed"] == case["expected"]["assessed"]
    assert saved["entries"][0]["failure"]["phase"] == "acquisition"
    assert not saved["execution_complete"] and saved["ci_exit_code"] == 2
    assert (path.read_bytes(), settings.read_bytes()) == before
    def comparison(v):
        assert v["ci_exit_code"] == 2 and not v["execution_complete"]
        assert v["entries"][0]["classification"] == "failed"
        assert v["entries"][0]["reasons"] == ["entry_failure"]
    exercise("compare", {"schema_version": 1, "before": saved, "after": saved}, comparison, 2,
             cli_path, server_url, workflow_go_driver, tmp_path, workflow_evidence)
    def evidence(v):
        assert v["source_ci_exit_code"] == 2 and not v["execution_complete"]
        assert "missing.spl2" not in json.dumps(v)
    exercise("evidence", {"schema_version": 1, "report": saved, "include": []}, evidence, 0,
             cli_path, server_url, workflow_go_driver, tmp_path, workflow_evidence)


def test_sarif_original_ranges(cli_path, server_url, workflow_go_driver, tmp_path, workflow_evidence):
    q = copy.deepcopy(next(c["request"] for c in CASES if c["id"] == "mixed-statuses"))
    q["format"] = "sarif"
    def authored(value):
        from jsonschema import Draft7Validator
        owner = required_absolute_path("SPL_TOOLING_SOURCE_ROOT")
        Draft7Validator(json.loads((owner / "contracts/sarif/sarif-schema-2.1.0.json").read_text())).validate(value)
        run = value["runs"][0]
        locations = [r for r in run["results"] if r.get("locations")]
        assert locations
        for result in locations:
            assert result["properties"]["domain"] == "original"
            doc = next(d for d in q["documents"] if d["id"] == result["properties"]["document_id"])
            for location in result["locations"]:
                physical = location["physicalLocation"]
                artifact = run["artifacts"][physical["artifactLocation"]["index"]]
                assert artifact["location"]["uri"] == physical["artifactLocation"]["uri"]
                region = physical["region"]
                assert region["startLine"] == region["endLine"] == 1
                assert 1 <= region["startColumn"] <= region["endColumn"] <= len(doc["document"]["text"]) + 1
        invalid = [r for r in locations if r["properties"]["document_id"] == "invalid"]
        assert any(r["locations"][0]["physicalLocation"]["region"] ==
                   {"startLine": 1, "startColumn": 17, "endLine": 1, "endColumn": 30} for r in invalid)
    exercise("assess", q, authored, 1, cli_path, server_url, workflow_go_driver, tmp_path, workflow_evidence)
