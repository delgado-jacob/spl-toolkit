#!/usr/bin/env python3
"""Run sarif-tools on actual candidate output and independently check locations."""
from __future__ import annotations
import argparse
import csv
import hashlib
import json
import os
from pathlib import Path
import re
import subprocess
import tempfile
from urllib.parse import quote, unquote, urljoin, urlparse

import jsonschema
from referencing import Registry, Resource
from referencing.exceptions import NoSuchResource

ROOT = Path(__file__).resolve().parents[1]


def check_semantics(value: dict, snapshots: dict[str, str], source_root: Path, cases: dict):
    schema = json.loads((ROOT / "contracts/sarif/sarif-schema-2.1.0.json").read_text())
    def deny(uri):
        raise NoSuchResource(ref=uri)
    registry = Registry(retrieve=deny).with_resource(schema["id"], Resource.from_contents(schema))
    validator = jsonschema.validators.validator_for(schema)
    validator.check_schema(schema)
    failures = list(validator(schema, registry=registry, format_checker=jsonschema.FormatChecker()).iter_errors(value))
    if failures:
        raise AssertionError([error.message for error in failures])
    run = value["runs"][0]
    assert run["columnKind"] == "unicodeCodePoints"
    assert run["originalUriBaseIds"]["SRCROOT"]["uri"] == source_root.as_uri() + "/"
    assert run["invocations"][0]["executionSuccessful"] is True
    artifacts = run["artifacts"]
    rules = run["tool"]["driver"]["rules"]
    located = []
    for finding in run["results"]:
        assert rules[finding["ruleIndex"]]["id"] == finding["ruleId"]
        for location in finding.get("locations", []):
            physical = location["physicalLocation"]
            artifact = physical["artifactLocation"]
            assert artifacts[artifact["index"]]["location"]["uri"] == artifact["uri"]
            relative = unquote(artifact["uri"])
            assert relative in snapshots
            assert artifact["uri"] == quote(relative, safe="/")
            resolved = urljoin(run["originalUriBaseIds"][artifact["uriBaseId"]]["uri"], artifact["uri"])
            assert Path(unquote(urlparse(resolved).path)) == source_root / relative
            region = physical["region"]
            text = snapshots[relative]
            lines = text.splitlines() or [""]
            for prefix in ("start", "end"):
                line, column = region[prefix + "Line"], region[prefix + "Column"]
                assert 1 <= line <= len(lines)
                assert 1 <= column <= len(lines[line - 1]) + 1
            located.append(finding)
    syntax = next(item for item in located if item["ruleId"] == cases["syntax_rule"])
    physical = syntax["locations"][0]["physicalLocation"]
    assert physical["artifactLocation"]["uri"] == cases["syntax_uri"]
    assert {key: physical["region"][key] for key in cases["syntax_region"]} == cases["syntax_region"]
    assert syntax["level"] == "error" and syntax["message"]["text"]
    assert any(item["ruleId"] != cases["syntax_rule"] for item in run["results"])
    return run


def run(cli: Path, sarif: Path, source_sha: str, evidence: Path):
    if not re.fullmatch(r"[0-9a-f]{40}", source_sha):
        raise ValueError("source-sha must be a full Git SHA")
    cli, sarif = cli.resolve(strict=True), sarif.resolve(strict=True)
    evidence.parent.mkdir(parents=True, exist_ok=True)
    result = {"status": "failed", "source_sha": source_sha, "cli": str(cli),
              "cli_sha256": hashlib.sha256(cli.read_bytes()).hexdigest(), "consumer": str(sarif), "commands": []}
    result["input_sha256"] = {name: hashlib.sha256((ROOT / name).read_bytes()).hexdigest() for name in (
        "tests/editor-client/requirements-sarif-local-hashed.lock", "tests/editor-client/consumer-provenance.json",
        "testdata/tooling/consumer-cases.json", "contracts/sarif/sarif-schema-2.1.0.json")}
    cases = json.loads((ROOT / "testdata/tooling/consumer-cases.json").read_text())
    # Retain durable snapshots/conversions next to evidence for source inspection.
    root = Path(tempfile.mkdtemp(prefix="sarif-run-", dir=evidence.parent))
    sources = root / "sources"
    sources.mkdir()
    snapshots = {case["path"]: case["text"] for case in cases["sources"]}
    for name, text in snapshots.items():
        (sources / name).write_text(text, encoding="utf-8")
    env = os.environ | {"MPLCONFIGDIR": str(root / "matplotlib"), "MPLBACKEND": "Agg"}
    def execute(command, allowed=(0,)):
        value = subprocess.run([str(arg) for arg in command], capture_output=True, text=True, env=env, timeout=90)
        result["commands"].append({"argv": [str(arg) for arg in command], "exit_code": value.returncode,
                                   "stdout": value.stdout, "stderr": value.stderr})
        if value.returncode not in allowed:
            raise RuntimeError(f"command exited {value.returncode}: {command}")
        return value
    try:
        version = execute([sarif, "--version"]).stdout.strip()
        assert "3.0.5" in version
        result["consumer_version"] = version
        output = root / "findings.sarif"
        execute([cli, "scan", "--directory", sources, "--format", "sarif", "--output", output], (1,))
        value = json.loads(output.read_text())
        report = check_semantics(value, snapshots, sources, cases)
        corpus = json.loads(execute([cli, "scan", "--directory", sources, "--format", "json"], (1,)).stdout)
        assert corpus["status"] == "invalid" and corpus["execution_complete"] is True
        for entry, expected in zip(corpus["entries"], cases["sources"], strict=True):
            assert entry["evaluation"]["analysis"]["status"] == expected["status"]
        execute([sarif, "summary", output])
        csv_path, html_path = root / "findings.csv", root / "findings.html"
        execute([sarif, "csv", "--output", csv_path, output])
        execute([sarif, "html", "--no-autotrim", "--output", html_path, output])
        with csv_path.open(newline="") as handle:
            rows = list(csv.DictReader(handle))
        assert len(rows) == len(report["results"])
        for finding in report["results"]:
            physical = finding.get("locations", [{}])[0].get("physicalLocation")
            # sarif-tools renders unlocated results as its own sentinel '-':1.
            expected_uri = physical["artifactLocation"]["uri"] if physical else "-"
            expected_line = physical["region"]["startLine"] if physical else 1
            assert any(row["Code"] == finding["ruleId"] and row["Description"] == finding["message"]["text"]
                       and row["Severity"] == finding["level"] and row["Location"] == expected_uri
                       and int(row["Line"] or 0) == expected_line for row in rows)
        html = html_path.read_text()
        assert cases["syntax_rule"] in html and cases["syntax_uri"] in html
        assert all((sources / name).read_bytes() == text.encode() for name, text in snapshots.items())
        assert hashlib.sha256(cli.read_bytes()).hexdigest() == result["cli_sha256"], "candidate executable changed"
        result.update(status="passed", csv_rows=rows, findings=len(report["results"]),
                      assertions=["official OASIS Draft04 schema", "independent URI/base/index/Unicode region checks",
                                  "invalid and incomplete retained with successful execution", "summary CSV HTML real consumer",
                                  "CSV rule/message/severity/path/line preservation", "source snapshots unchanged"],
                      artifacts={str(p): hashlib.sha256(p.read_bytes()).hexdigest() for p in [output, csv_path, html_path, *sources.iterdir()]})
    except Exception as error:
        result["error"] = f"{type(error).__name__}: {error}"
    evidence.write_text(json.dumps(result, indent=2) + "\n")
    return result


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    for name in ("cli", "sarif", "evidence"):
        parser.add_argument("--" + name, type=Path, required=True)
    parser.add_argument("--source-sha", required=True)
    args = parser.parse_args()
    result = run(args.cli, args.sarif, args.source_sha, args.evidence)
    print(json.dumps({"status": result["status"], "evidence": str(args.evidence)}))
    return 0 if result["status"] == "passed" else 1


if __name__ == "__main__":
    raise SystemExit(main())
