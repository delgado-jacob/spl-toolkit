"""Fail closed on stale or incomplete Go transport evidence."""
from copy import deepcopy
import hashlib
import importlib.util
import json
from pathlib import Path

import pytest

ROOT = Path(__file__).resolve().parents[2]
spec = importlib.util.spec_from_file_location("spl2_transport", ROOT / "tests/acceptance/spl2_transport.py")
transport = importlib.util.module_from_spec(spec)
spec.loader.exec_module(transport)
sha = lambda path: hashlib.sha256(path.read_bytes()).hexdigest()


def evidence(tmp_path):
    fixtures = tmp_path / "fixtures"
    fixtures.mkdir()
    document = {"text": "FROM main", "language": "spl2"}
    (fixtures / "manifest.json").write_text(json.dumps({"case_files": ["cases.json"]}))
    (fixtures / "cases.json").write_text(json.dumps([{"id": "sample", "document": document}]))
    source = tmp_path / "source"
    (source / "pkg/analysis").mkdir(parents=True)
    (source / "pkg/analysis/core.go").write_text("package analysis")
    report = {"schema_version": 1, "document": {"profile": "splunkd", "version": "current", "source_id": ""} | document,
              "status": "incomplete", "coverage": {"syntax_complete": True, "semantic_complete": False, "reasons": []},
              "stages": [], "scopes": [], "references": [], "lineage": [], "diagnostics": [],
              "requirements": {"schema_version": 1, "query": {}, "capability_revision": "sha256:" + "0" * 64,
                               "query_status": "incomplete", "coverage": {"complete": False, "reasons": []},
                               "items": [], "gaps": [], "diagnostics": []},
              "dependencies": {k: [] for k in ("indexes", "sources", "source_types", "datasets", "lookups", "data_models", "macros")}}
    input_evidence = {
        "inputs": [],
        "input_coverage": {"state": "partial", "reasons": []},
        "field_attribution_coverage": {"state": "partial", "reasons": []},
        "correlation": {"outcome": "indeterminate", "coverage": {"state": "partial", "reasons": []},
                        "nodes": [], "edges": [], "components": []},
    }
    report.update(deepcopy(input_evidence))
    report["requirements"].update(deepcopy(input_evidence))
    milestone_document = {
        "text": "FROM {kind: \"index\"}",
        "language": "spl2",
        "profile": "splunkd",
        "version": "current",
        "source_id": "milestone11:descriptor",
    }
    milestone_requirements = deepcopy(report["requirements"])
    milestone_requirements["query_status"] = "valid"
    milestone_requirements["coverage"] = {"complete": True, "reasons": []}
    milestone_analysis = deepcopy(report)
    milestone_analysis["document"] = milestone_document
    milestone_analysis["status"] = "valid"
    milestone_analysis["coverage"] = {
        "syntax_complete": True,
        "semantic_complete": True,
        "reasons": [],
    }
    milestone_analysis["requirements"] = deepcopy(milestone_requirements)
    artifact = {"schema_version": 1, "kind": "spl2-go-transport", "conformance_credit": 0,
                "source_hashes": {"pkg/analysis/core.go": sha(source / "pkg/analysis/core.go")},
                "fixture_hashes": {p.name: sha(p) for p in fixtures.glob("*.json")},
                "reports": [{"id": "sample", "document": {"profile": "", "version": "", "source_id": ""} | document, "report": report}],
                "milestone11": [{"id": "descriptor", "document": milestone_document,
                                 "analysis": milestone_analysis,
                                 "requirements": milestone_requirements}]}
    return fixtures, source, artifact, [
        {"id": "descriptor", "document": deepcopy(milestone_document)}
    ]


@pytest.mark.parametrize("mutation", ["duplicate", "missing", "extra", "wrong-document", "empty-report", "truncated-report", "fixture-hash", "source-hash", "credit"])
def test_transport_rejects_invalid_evidence(tmp_path, mutation):
    fixtures, source, artifact, milestone_documents = evidence(tmp_path)
    if mutation == "duplicate": artifact["reports"].append(deepcopy(artifact["reports"][0]))
    if mutation == "missing": artifact["reports"] = []
    if mutation == "extra":
        artifact["reports"].append(deepcopy(artifact["reports"][0])); artifact["reports"][-1]["id"] = "extra"
    if mutation == "wrong-document": artifact["reports"][0]["report"]["document"]["text"] = "FROM other"
    if mutation == "empty-report": artifact["reports"][0]["report"] = {}
    if mutation == "truncated-report": del artifact["reports"][0]["report"]["references"]
    if mutation == "fixture-hash": artifact["fixture_hashes"]["cases.json"] = "0" * 64
    if mutation == "source-hash": artifact["source_hashes"]["pkg/analysis/core.go"] = "0" * 64
    if mutation == "credit": artifact["conformance_credit"] = 1
    path = tmp_path / "transport.json"
    path.write_text(json.dumps(artifact))
    with pytest.raises(AssertionError):
        transport.load_transport(path, fixtures, sha(path), source, milestone_documents)


@pytest.mark.parametrize(
    "mutation",
    ["missing", "duplicate", "extra", "wrong-document", "truncated-analysis", "requirements-mismatch"],
)
def test_transport_rejects_invalid_milestone11_evidence(tmp_path, mutation):
    fixtures, source, artifact, milestone_documents = evidence(tmp_path)
    if mutation == "missing":
        artifact["milestone11"] = []
    if mutation == "duplicate":
        artifact["milestone11"].append(deepcopy(artifact["milestone11"][0]))
    if mutation == "extra":
        artifact["milestone11"].append(deepcopy(artifact["milestone11"][0]))
        artifact["milestone11"][-1]["id"] = "extra"
    if mutation == "wrong-document":
        artifact["milestone11"][0]["document"]["text"] = "FROM other"
    if mutation == "truncated-analysis":
        del artifact["milestone11"][0]["analysis"]["references"]
    if mutation == "requirements-mismatch":
        artifact["milestone11"][0]["requirements"]["query_status"] = "invalid"
    path = tmp_path / "transport.json"
    path.write_text(json.dumps(artifact))
    with pytest.raises(AssertionError):
        transport.load_transport(path, fixtures, sha(path), source, milestone_documents)


def test_transport_hash_and_original_identity(tmp_path):
    fixtures, source, artifact, milestone_documents = evidence(tmp_path)
    path = tmp_path / "transport.json"
    path.write_text(json.dumps(artifact))
    parsed, reports = transport.load_transport(
        path, fixtures, sha(path), source, milestone_documents
    )
    assert parsed == artifact and reports == {"sample": artifact["reports"][0]["report"]}
    with pytest.raises(AssertionError):
        transport.load_transport(path, fixtures, "0" * 64, source, milestone_documents)
    artifact["reports"][0]["document"]["source_id"] = "changed"
    path.write_text(json.dumps(artifact))
    with pytest.raises(AssertionError):
        transport.load_transport(path, fixtures, sha(path), source, milestone_documents)


@pytest.mark.parametrize("member", sorted(transport.INPUT_EVIDENCE_KEYS))
@pytest.mark.parametrize("mutation", ["missing", "null", "embedded-mismatch"])
def test_transport_rejects_invalid_input_evidence(tmp_path, member, mutation):
    fixtures, source, artifact, milestone_documents = evidence(tmp_path)
    report = artifact["reports"][0]["report"]
    if mutation == "missing":
        del report[member]
    elif mutation == "null":
        report[member] = report["requirements"][member] = None
    else:
        report["requirements"][member] = None
    path = tmp_path / "transport.json"
    path.write_text(json.dumps(artifact))
    with pytest.raises(AssertionError):
        transport.load_transport(path, fixtures, sha(path), source, milestone_documents)


@pytest.mark.parametrize("mutation", ["coverage-state", "coverage-reasons", "correlation-outcome", "correlation-edges", "dangling-node"])
def test_transport_rejects_malformed_input_structure(tmp_path, mutation):
    fixtures, source, artifact, milestone_documents = evidence(tmp_path)
    report = artifact["reports"][0]["report"]
    if mutation == "coverage-state":
        report["input_coverage"]["state"] = "unknown"
    elif mutation == "coverage-reasons":
        report["field_attribution_coverage"]["reasons"] = None
    elif mutation == "correlation-outcome":
        report["correlation"]["outcome"] = "unknown"
    elif mutation == "correlation-edges":
        report["correlation"]["edges"] = None
    else:
        report["correlation"]["nodes"] = [{"input_id": "absent", "occurrence_id": "absent"}]
    for key in transport.INPUT_EVIDENCE_KEYS:
        report["requirements"][key] = deepcopy(report[key])
    path = tmp_path / "transport.json"
    path.write_text(json.dumps(artifact))
    with pytest.raises(AssertionError):
        transport.load_transport(path, fixtures, sha(path), source, milestone_documents)
