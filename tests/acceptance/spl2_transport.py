"""Full Go reports are transport evidence, never semantic conformance authority."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import subprocess

REPORT_KEYS = {"schema_version", "document", "status", "coverage", "stages", "scopes", "references", "lineage", "dependencies", "diagnostics", "requirements"}


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def source_hashes(root):
    paths = list((root / "pkg/analysis").glob("*.go")) + list((root / "parser").rglob("*.go")) + list((root / "grammar").glob("*.g4"))
    paths += [root / name for name in ("go.mod", "go.sum") if (root / name).is_file()]
    return {p.relative_to(root).as_posix(): digest(p) for p in sorted(paths)}


def load_transport(path, fixtures, expected_sha256, source_root=None):
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
        assert set(report) == REPORT_KEYS, "empty or truncated Go report"
        assert report["document"] == expected, "wrong normalized Query Document"
        assert report["schema_version"] == 1 and report["status"] in {"valid", "invalid", "incomplete"}
        assert set(report["coverage"]) == {"syntax_complete", "semantic_complete", "reasons"}
        assert isinstance(report["requirements"], dict), "truncated Go requirements"
        for key in ("stages", "scopes", "references", "lineage", "diagnostics"):
            assert isinstance(report[key], list), f"truncated Go {key}"
        reports[identity] = report
    assert documents and set(reports) == set(documents), "missing Go reports"
    return artifact, reports


def prepare_transport(root, output):
    """Run the actual Go corpus checks once; installed consumers only load JSON."""
    env = os.environ.copy()
    env["SPL_SPL2_GO_REPORTS"] = str(output.resolve())
    subprocess.run(["go", "test", "-mod=readonly", "./pkg/analysis", "-run", "^TestSPL2CorpusCanonical$", "-count=1"], cwd=root, env=env, check=True)
    load_transport(output, root / "testdata/spl2", digest(output), root)
    return output


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--root", required=True, type=Path)
    parser.add_argument("--output", required=True, type=Path)
    args = parser.parse_args()
    prepare_transport(args.root.resolve(), args.output.resolve())
