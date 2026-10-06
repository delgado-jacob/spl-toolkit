"""Full Go reports are transport evidence, never semantic conformance authority."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import subprocess
import tempfile

INPUT_EVIDENCE_KEYS = {"inputs", "input_coverage", "field_attribution_coverage", "correlation"}
REPORT_KEYS = {"schema_version", "document", "status", "coverage", "stages", "scopes", "references", "lineage", "dependencies", "diagnostics", "requirements"} | INPUT_EVIDENCE_KEYS
GO_HELPER = r'''
package main

import (
	"encoding/json"
	"os"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
)

type documentCase struct {
	ID       string                 `json:"id"`
	Document analysis.QueryDocument `json:"document"`
}

type reportCase struct {
	ID           string                   `json:"id"`
	Document     analysis.QueryDocument   `json:"document"`
	Analysis     *analysis.Result         `json:"analysis"`
	Requirements *analysis.RequirementSet `json:"requirements"`
}

func main() {
	var request struct {
		Documents []documentCase `json:"documents"`
	}
	if err := json.NewDecoder(os.Stdin).Decode(&request); err != nil {
		panic(err)
	}
	response := struct {
		Milestone11 []reportCase `json:"milestone11"`
	}{Milestone11: make([]reportCase, 0, len(request.Documents))}
	for _, item := range request.Documents {
		analyzed, err := analysis.Analyze(item.Document)
		if err != nil {
			panic(err)
		}
		requirements, err := analysis.Requirements(item.Document)
		if err != nil {
			panic(err)
		}
		response.Milestone11 = append(response.Milestone11, reportCase{
			ID: item.ID, Document: item.Document,
			Analysis: analyzed, Requirements: requirements,
		})
	}
	if err := json.NewEncoder(os.Stdout).Encode(response); err != nil {
		panic(err)
	}
}
'''


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def source_hashes(root):
    paths = list((root / "pkg/analysis").glob("*.go")) + list((root / "parser").rglob("*.go")) + list((root / "grammar").glob("*.g4"))
    paths += [root / name for name in ("go.mod", "go.sum") if (root / name).is_file()]
    return {p.relative_to(root).as_posix(): digest(p) for p in sorted(paths)}


def milestone11_documents(path=None):
    path = path or Path(__file__).with_name("cli_examples.json")
    manifest = json.loads(path.read_text(encoding="utf-8"))
    documents = manifest.get("milestone11_documents")
    assert isinstance(documents, list) and documents, "missing Milestone 11 documents"
    assert len({item["id"] for item in documents}) == len(documents), "duplicate Milestone 11 document ID"
    for item in documents:
        assert set(item) == {"id", "document"}
        assert set(item["document"]) == {"text", "language", "profile", "version", "source_id"}
        assert item["document"]["language"] == "spl2"
        assert item["document"]["profile"] == "splunkd"
        assert item["document"]["version"] == "current"
        assert item["document"]["source_id"] == "milestone11:" + item["id"]
    return documents


def assert_full_report(report, document):
    assert set(report) == REPORT_KEYS, "empty or truncated Go report"
    assert report["document"] == document, "wrong normalized Query Document"
    assert report["schema_version"] == 1 and report["status"] in {"valid", "invalid", "incomplete"}
    assert set(report["coverage"]) == {"syntax_complete", "semantic_complete", "reasons"}
    assert isinstance(report["requirements"], dict), "truncated Go requirements"
    for key in ("stages", "scopes", "references", "lineage", "diagnostics"):
        assert isinstance(report[key], list), f"truncated Go {key}"

    # These are canonical transport facts, not independently authored expectations.
    for key in INPUT_EVIDENCE_KEYS:
        assert key in report["requirements"] and report[key] == report["requirements"][key], f"inconsistent Go {key}"
    assert isinstance(report["inputs"], list), "truncated Go inputs"
    correlation = report["correlation"]
    assert isinstance(correlation, dict) and set(correlation) == {"outcome", "coverage", "nodes", "edges", "components"}
    assert correlation["outcome"] in {"connected", "disconnected", "indeterminate", "not applicable"}
    for key in ("nodes", "edges", "components"):
        assert isinstance(correlation[key], list), f"truncated Go correlation {key}"
    for coverage in (report["input_coverage"], report["field_attribution_coverage"], correlation["coverage"]):
        assert isinstance(coverage, dict) and set(coverage) == {"state", "reasons"}
        assert coverage["state"] in {"complete", "partial", "not_applicable"}
        assert isinstance(coverage["reasons"], list) and all(isinstance(reason, dict) for reason in coverage["reasons"])
    occurrences = set()
    input_ids = set()
    for source in report["inputs"]:
        assert isinstance(source, dict) and set(source) == {"id", "kind", "name", "identity", "evidence", "occurrences"}
        assert isinstance(source["id"], str) and source["id"] and source["id"] not in input_ids
        input_ids.add(source["id"])
        assert isinstance(source["occurrences"], list) and source["occurrences"]
        for occurrence in source["occurrences"]:
            assert isinstance(occurrence, dict) and isinstance(occurrence.get("id"), str) and occurrence["id"]
            pair = (source["id"], occurrence["id"])
            assert pair not in occurrences
            occurrences.add(pair)
    nodes = []
    for node in correlation["nodes"]:
        assert isinstance(node, dict) and set(node) == {"input_id", "occurrence_id"}
        nodes.append((node["input_id"], node["occurrence_id"]))
    assert len(nodes) == len(set(nodes)) and set(nodes) == occurrences, "inconsistent Go correlation source occurrences"


def load_transport(path, fixtures, expected_sha256, source_root=None, expected_milestone11=None):
    assert digest(path) == expected_sha256, "Go transport artifact hash mismatch"
    artifact = json.loads(path.read_text(encoding="utf-8"))
    assert artifact.get("schema_version") == 1 and artifact.get("kind") == "spl2-go-transport"
    assert artifact.get("conformance_credit") == 0, "transport evidence cannot grant conformance credit"
    assert artifact.get("source_hashes"), "missing producing Go sources"
    if source_root is not None:
        assert artifact["source_hashes"] == source_hashes(source_root), "stale producing Go sources"
    assert artifact.get("fixture_hashes") == {p.name: digest(p) for p in sorted(fixtures.glob("*.json"))}, "stale corpus fixtures"
    manifest = json.loads((fixtures / "manifest.json").read_text(encoding="utf-8"))
    documents = {}
    for filename in manifest["case_files"]:
        cases = json.loads((fixtures / filename).read_text(encoding="utf-8"))
        assert cases, "empty required corpus file"
        for case in cases:
            assert case["id"] not in documents, "duplicate corpus ID"
            documents[case["id"]] = case["document"]
    reports = {}
    for entry in artifact["reports"]:
        identity = entry["id"]
        assert identity in documents and identity not in reports, "extra or duplicate Go report ID"
        document = {"language": "", "profile": "", "version": "", "source_id": ""} | documents[identity]
        assert entry["document"] == document, "wrong original Query Document"
        expected = document | {key: document[key] or default for key, default in (("language", "spl"), ("profile", "splunkd"), ("version", "current"))}
        report = entry["report"]
        assert_full_report(report, expected)
        reports[identity] = report
    assert documents and set(reports) == set(documents), "missing Go reports"

    expected_milestone11 = expected_milestone11 or milestone11_documents()
    milestone_documents = {item["id"]: item["document"] for item in expected_milestone11}
    milestone_reports = {}
    for entry in artifact.get("milestone11", []):
        identity = entry["id"]
        assert identity in milestone_documents and identity not in milestone_reports, "extra or duplicate Milestone 11 report ID"
        document = milestone_documents[identity]
        assert entry["document"] == document, "wrong Milestone 11 Query Document"
        assert_full_report(entry["analysis"], document)
        assert entry["analysis"]["requirements"] == entry["requirements"], "embedded and standalone Go requirements differ"
        assert entry["analysis"]["status"] == "valid"
        assert entry["analysis"]["coverage"] == {
            "syntax_complete": True, "semantic_complete": True, "reasons": [],
        }
        assert entry["requirements"]["query_status"] == "valid"
        assert entry["requirements"]["coverage"] == {"complete": True, "reasons": []}
        milestone_reports[identity] = entry
    assert set(milestone_reports) == set(milestone_documents), "missing Milestone 11 Go reports"
    return artifact, reports


def prepare_transport(root, output):
    """Run the actual Go corpus checks once; installed consumers only load JSON."""
    env = os.environ.copy()
    env["SPL_SPL2_GO_REPORTS"] = str(output.resolve())
    subprocess.run(["go", "test", "-mod=readonly", "./pkg/analysis", "-run", "^TestSPL2CorpusCanonical$", "-count=1"], cwd=root, env=env, check=True)
    documents = milestone11_documents(root / "tests/acceptance/cli_examples.json")
    with tempfile.TemporaryDirectory(prefix="spl2-milestone11-go-") as temporary:
        helper = Path(temporary) / "main.go"
        helper.write_text(GO_HELPER, encoding="utf-8")
        helper_env = env.copy()
        helper_env["GOWORK"] = "off"
        completed = subprocess.run(
            ["go", "run", "-mod=readonly", str(helper)], cwd=root, env=helper_env,
            input=json.dumps({"documents": documents}, ensure_ascii=False),
            capture_output=True, text=True, check=True, timeout=90,
        )
    artifact = json.loads(output.read_text(encoding="utf-8"))
    artifact["milestone11"] = json.loads(completed.stdout)["milestone11"]
    output.write_text(
        json.dumps(artifact, ensure_ascii=False, separators=(",", ":")) + "\n",
        encoding="utf-8",
    )
    load_transport(output, root / "testdata/spl2", digest(output), root, documents)
    return output


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--root", required=True, type=Path)
    parser.add_argument("--output", required=True, type=Path)
    args = parser.parse_args()
    prepare_transport(args.root.resolve(), args.output.resolve())
