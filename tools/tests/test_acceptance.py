"""Focused tests for the strict milestone acceptance evidence gate."""

from __future__ import annotations

import copy
import json
from pathlib import Path

from tools.check_acceptance import load_records, validate_records


SHA = "a" * 40
HASH = "b" * 64
ROOT = Path(__file__).resolve().parents[2]
TARGETS = {
    "linux-amd64": "x86_64",
    "darwin-amd64": "x86_64",
    "darwin-arm64": "arm64",
    "windows-amd64": "x86_64",
}
PYTHONS = ("3.11.9", "3.12.10", "3.13.7", "3.14.0")


def release_environment(target: str, architecture: str) -> dict:
    runner = {
        "linux-amd64": "ubuntu-24.04",
        "darwin-amd64": "macos-15-intel",
        "darwin-arm64": "macos-15",
        "windows-amd64": "windows-2022",
    }[target]
    darwin = target.startswith("darwin-")
    return {
        "configured_runner": runner,
        "observed_runner_name": "GitHub Actions 1",
        "observed_image_os": "macos15" if darwin else "ubuntu24" if target.startswith("linux-") else "win22",
        "observed_image_version": "20260901.1",
        "evidence_kind": "pinned-runner",
        "pinned_environment": True,
        "target": target,
        "os": "observed platform",
        "architecture": architecture,
        "go": "go version go1.26.8 target/arch",
        "python": "3.11.9",
        "packaging": {"setuptools": "80.9.0", "wheel": "0.45.1", "build": "1.3.0", "packaging": "25.0"},
        "cc": "/toolchain/clang" if darwin else "/toolchain/gcc",
        "cc_version": "Apple clang version 16.0.0" if darwin else "gcc 13.3.0" if target.startswith("linux-") else "gcc 14.2.0",
        "xcode": "Xcode 16.4" if darwin else None,
        "linker_path": "/toolchain/ld",
        "linker": "observed linker",
        "sdk": "/Applications/Xcode_16.4.app/SDK" if darwin else None,
        "zlib": "1.2.12",
        "wheel_platform": {
            "linux-amd64": "linux_x86_64", "darwin-amd64": "macosx_15_0_x86_64",
            "darwin-arm64": "macosx_15_0_arm64", "windows-amd64": "win_amd64",
        }[target],
        "source_date_epoch": 1788652800,
        "goflags": ["-mod=readonly", "-trimpath", "-buildvcs=false", "-buildid=", "version"],
        "cgo_enabled_cli_server": "0",
        "cgo_enabled_native": "1",
        "native_argv": ["go", "build"],
        "cgo_cflags": "flags",
        "cgo_cppflags": "flags",
        "cgo_ldflags": "flags",
        "artifacts": {"wheel.whl": HASH},
    }


def passing_records() -> list[dict]:
    records: list[dict] = []
    for target, architecture in TARGETS.items():
        environment = release_environment(target, architecture)
        records.extend([
            {
                "schema_version": 1, "kind": "native", "source_sha": SHA,
                "status": "passed", "target": target, "architecture": architecture,
                "environment": environment,
            },
            {
                "schema_version": 1, "kind": "reproducibility", "source_sha": SHA,
                "status": "passed", "target": target, "architecture": architecture,
                "wheel_sha256": HASH, "artifact_hashes": {"wheel.whl": HASH},
                "accepted_payloads": {"cli": "cli", "server": "server", "native": "native"},
                "environment": environment,
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
                "wheel_contract_hashes": {
                    "spl_toolkit/" + path.relative_to(ROOT).as_posix(): HASH
                    for path in (ROOT / "contracts").rglob("*") if path.is_file()
                },
                "tooling_source_hashes": {
                    line: HASH for line in (ROOT / "python/native-source-files.txt").read_text().splitlines()
                    if line and not line.startswith("#")
                },
                "tooling_fixture_hashes": {name: HASH for name in (
                    "contracts.json", "graph-cases.json", "impact-cases.json", "requests.json",
                    "sarif-cases.json", "example-corpus.json", "example-target.json",
                    "example-corpus-missing-file.json", "../rewrite/forms.json",
                )},
                "machine_contract_tests": {"collected": 10, "passed": 10, "failed": 0, "skipped": 0},
                "fixture_hashes": {"requirements": HASH},
                "requirements_surface_evidence": {
                    "schema_version": 1,
                    "fixture_sha256": HASH,
                    "corpus_cases": 17,
                    "dense_cases": 4,
                    "concurrent_calls": 16,
                    "long_sparse_cases": 2,
                    "request_error_cases": 3,
                },
                "tests": {
                    "required_native": {"collected": 11, "passed": 11, "failed": 0, "skipped": 0},
                    "surface_acceptance": {"collected": 6, "passed": 6, "failed": 0, "skipped": 0},
                },
                "required_test_files": {
                    "native": ["test_native_abi.py", "test_native_mapper.py", "test_native_analysis.py",
                               "test_native_validation.py", "test_native_schema_validation.py", "test_native_spl2.py",
                               "test_native_rewrite.py", "test_native_requirements.py"],
                    "acceptance": ["test_documented_cli.py", "test_surfaces.py", "test_analysis_surfaces.py",
                                   "test_validation_surfaces.py", "test_schema_surfaces.py", "test_spl2_surfaces.py",
                                   "test_rewrite_surfaces.py", "test_requirements_surfaces.py"],
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


def test_installed_contract_evidence_cannot_be_missing_or_malformed():
    original = passing_records()
    index = next(i for i, record in enumerate(original) if record["kind"] == "installed-wheel")
    for field in ("wheel_contract_hashes", "tooling_source_hashes",
                  "tooling_fixture_hashes", "machine_contract_tests"):
        records = copy.deepcopy(original)
        del records[index][field]
        assert any(f"missing fields: {field}" in error for error in validate_records(records, SHA))

    records = copy.deepcopy(original)
    records[index]["machine_contract_tests"]["passed"] = 9
    assert any("machine_contract_tests must all pass" in error for error in validate_records(records, SHA))

    for field in ("wheel_contract_hashes", "tooling_source_hashes", "tooling_fixture_hashes"):
        records = copy.deepcopy(original)
        first = next(iter(records[index][field]))
        records[index][field][first] = "not a hash"
        assert any(f"{field} contains an invalid SHA-256" in error for error in validate_records(records, SHA))
        del records[index][field][first]
        assert any(f"{field} has incorrect paths" in error for error in validate_records(records, SHA))


def test_installed_evidence_cannot_omit_spl2_surface_suite():
    records = passing_records()
    installed = next(record for record in records if record["kind"] == "installed-wheel")
    installed["required_test_files"] = {
        "native": ["test_native_abi.py", "test_native_mapper.py", "test_native_analysis.py",
                   "test_native_validation.py", "test_native_schema_validation.py", "test_native_spl2.py"],
        "acceptance": ["test_documented_cli.py", "test_surfaces.py", "test_analysis_surfaces.py",
                       "test_validation_surfaces.py", "test_schema_surfaces.py"],
    }
    assert any("missing required suite test_spl2_surfaces.py" in error for error in validate_records(records, SHA))


def test_installed_evidence_requires_requirement_surfaces_and_bound_fixture_hash():
    original = passing_records()
    index = next(i for i, record in enumerate(original) if record["kind"] == "installed-wheel")

    for field in ("fixture_hashes", "requirements_surface_evidence"):
        records = copy.deepcopy(original)
        del records[index][field]
        assert any(f"missing fields: {field}" in error for error in validate_records(records, SHA))

    records = copy.deepcopy(original)
    records[index]["requirements_surface_evidence"]["fixture_sha256"] = "c" * 64
    assert any("requirements fixture hash" in error for error in validate_records(records, SHA))

    records = copy.deepcopy(original)
    records[index]["required_test_files"]["native"].remove("test_native_requirements.py")
    records[index]["required_test_files"]["acceptance"].remove("test_requirements_surfaces.py")
    errors = validate_records(records, SHA)
    assert any("test_native_requirements.py" in error for error in errors)
    assert any("test_requirements_surfaces.py" in error for error in errors)


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


def test_release_environment_is_required_and_cross_checked():
    records = passing_records()
    native = next(r for r in records if r["kind"] == "native" and r["target"] == "linux-amd64")
    reproducibility = next(r for r in records if r["kind"] == "reproducibility" and r["target"] == "darwin-arm64")
    native["environment"] = {"pinned_environment": True}
    reproducibility["environment"]["target"] = "darwin-amd64"
    reproducibility["environment"]["artifacts"] = {"wheel.whl": "d" * 64}

    errors = validate_records(records, SHA)

    assert any("environment missing fields" in error for error in errors)
    assert any("environment target" in error for error in errors)
    assert any("environment artifacts" in error for error in errors)


def test_mislabeled_runner_and_diagnostic_environment_are_rejected():
    records = passing_records()
    linux = next(r for r in records if r["kind"] == "native" and r["target"] == "linux-amd64")
    linux["environment"]["configured_runner"] = "macos-15"
    darwin = next(r for r in records if r["kind"] == "native" and r["target"] == "darwin-arm64")
    darwin["environment"].update(evidence_kind="diagnostic", pinned_environment=False)

    errors = validate_records(records, SHA)

    assert any("configured_runner" in error for error in errors)
    assert any("diagnostic" in error or "not pinned" in error for error in errors)


def test_task9_validated_pinned_local_environment_remains_acceptable():
    records = passing_records()
    environment = next(r for r in records if r["kind"] == "native" and r["target"] == "linux-amd64")["environment"]
    environment.update(
        evidence_kind="pinned-local",
        observed_runner_name="local",
        observed_image_os="unavailable",
        observed_image_version="unavailable",
    )

    assert validate_records(records, SHA) == []
