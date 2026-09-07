"""Focused tests for the strict milestone acceptance evidence gate."""

from __future__ import annotations

import copy
import json
from pathlib import Path

from tools.check_acceptance import load_records, validate_records


SHA = "a" * 40
HASH = "b" * 64
TARGETS = {
    "linux-amd64": "x86_64",
    "darwin-amd64": "x86_64",
    "darwin-arm64": "arm64",
    "windows-amd64": "x86_64",
}
PYTHONS = ("3.11.9", "3.12.10", "3.13.7", "3.14.0")


def passing_records() -> list[dict]:
    records: list[dict] = []
    for target, architecture in TARGETS.items():
        records.extend([
            {
                "schema_version": 1, "kind": "native", "source_sha": SHA,
                "status": "passed", "target": target, "architecture": architecture,
                "environment": {"compiler": "observed", "runner": "observed"},
            },
            {
                "schema_version": 1, "kind": "reproducibility", "source_sha": SHA,
                "status": "passed", "target": target, "architecture": architecture,
                "wheel_sha256": HASH, "artifact_hashes": {"wheel.whl": HASH},
                "accepted_payloads": {"cli": "cli", "server": "server", "native": "native"},
                "environment": {"pinned_environment": True},
                "checks": {"clean_source": "passed", "payloads": "passed", "package": "passed"},
            },
        ])
        for version in PYTHONS:
            records.append({
                "schema_version": 1, "kind": "installed-wheel", "source_sha": SHA,
                "status": "passed", "target": target, "architecture": architecture,
                "python_version": version, "python_runtime": version + " (main)", "wheel_sha256": HASH,
                "venv_prefix": f"/tmp/{target}/{version}/venv",
                "installed_module": f"/tmp/{target}/{version}/venv/site/spl_toolkit/__init__.py",
                "package_version": "0.1.1", "native_version": "0.1.1",
                "tests": {
                    "required_native": {"collected": 11, "passed": 11, "failed": 0, "skipped": 0},
                    "surface_acceptance": {"collected": 5, "passed": 5, "failed": 0, "skipped": 0},
                },
                "cli_examples": "passed", "surface_parity": "passed", "version_agreement": "passed",
            })
    records.extend([
        {"schema_version": 1, "kind": "go-floor", "source_sha": SHA, "status": "passed", "go_version": "1.22.12"},
        {"schema_version": 1, "kind": "native-memory", "source_sha": SHA, "status": "passed", "compiler": "gcc-13", "sanitizer": "address+leak"},
        {"schema_version": 1, "kind": "clean-source", "source_sha": SHA, "status": "passed", "tracked_hash_sha256": "c" * 64},
        {"schema_version": 1, "kind": "docker-examples", "source_sha": SHA, "status": "passed", "cli": "passed", "python_native": "passed", "server": "passed", "make_workflows": "passed"},
    ])
    return records


def test_complete_current_evidence_passes():
    assert validate_records(passing_records(), SHA) == []


def test_missing_arm64_is_not_complete():
    records = [r for r in passing_records() if not (r["kind"] == "native" and r.get("target") == "darwin-arm64")]
    assert any("darwin-arm64" in error and "native" in error for error in validate_records(records, SHA))


def test_mismatched_source_is_rejected():
    records = passing_records()
    records[0]["source_sha"] = "d" * 40
    assert any("source_sha" in error for error in validate_records(records, SHA))


def test_changed_wheel_checksum_is_rejected():
    records = passing_records()
    changed = next(r for r in records if r["kind"] == "installed-wheel")
    changed["wheel_sha256"] = "e" * 64
    assert any("wheel_sha256" in error for error in validate_records(records, SHA))


def test_duplicate_unknown_failed_and_stale_records_are_all_rejected():
    records = passing_records()
    records.append(copy.deepcopy(records[0]))
    records[1]["surprise"] = True
    del records[1]["checks"]
    records[2]["status"] = "failed"
    records.append({"schema_version": 1, "kind": "go-floor", "source_sha": "f" * 40, "status": "passed", "go_version": "1.22.12"})
    errors = validate_records(records, SHA)
    assert any("duplicate" in error for error in errors)
    assert any("unknown fields" in error for error in errors)
    assert any("missing fields" in error for error in errors)
    assert any("status" in error for error in errors)
    assert any("source_sha" in error for error in errors)


def test_skips_zero_collection_and_outside_module_are_rejected():
    records = passing_records()
    first, second = [r for r in records if r["kind"] == "installed-wheel"][:2]
    first["tests"]["required_native"]["skipped"] = 1
    first["tests"]["required_native"]["passed"] = 10
    second["tests"]["surface_acceptance"] = {"collected": 0, "passed": 0, "failed": 0, "skipped": 0}
    second["installed_module"] = "/tmp/controller/spl_toolkit/__init__.py"
    errors = validate_records(records, SHA)
    assert any("skipped" in error for error in errors)
    assert any("collected" in error for error in errors)
    assert any("installed_module" in error for error in errors)


def test_loader_rejects_arrays_and_malformed_json(tmp_path: Path):
    (tmp_path / "array.json").write_text("[]\n", encoding="utf-8")
    (tmp_path / "broken.json").write_text("{", encoding="utf-8")
    records, errors = load_records(tmp_path)
    assert records == []
    assert len(errors) == 2
