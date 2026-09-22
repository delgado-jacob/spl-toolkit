#!/usr/bin/env python3
"""Require complete, current evidence for the Milestone 1 release baseline."""

from __future__ import annotations

import argparse
import hashlib
import json
from pathlib import Path, PurePosixPath
import re
import subprocess
import sys


ROOT = Path(__file__).resolve().parents[1]
CONFIG = json.loads((ROOT / "tools" / "release-env.json").read_text(encoding="utf-8"))
TARGETS = CONFIG["targets"]
PYTHON_VERSIONS = tuple(CONFIG["test_python"])
EXPECTED_COUNTS = {"required_native": 11, "surface_acceptance": 6}
EXPECTED_MACHINE_CONTRACT_COUNT = 13
REQUIRED_TEST_FILES = {
    "native": {"test_native_abi.py", "test_native_mapper.py", "test_native_analysis.py",
               "test_native_validation.py", "test_native_schema_validation.py", "test_native_spl2.py",
               "test_native_rewrite.py", "test_native_requirements.py"},
    "acceptance": {"test_documented_cli.py", "test_surfaces.py", "test_analysis_surfaces.py",
                   "test_requirements_surfaces.py", "test_validation_surfaces.py", "test_schema_surfaces.py",
                   "test_spl2_surfaces.py", "test_rewrite_surfaces.py"},
}
REQUIRED_TEST_HASH_PATHS = {
    "native": {
        "test_native_requirements.py": ROOT / "python/tests/test_native_requirements.py",
        "test_native_analysis.py": ROOT / "python/tests/test_native_analysis.py",
        "test_native_spl2.py": ROOT / "python/tests/test_native_spl2.py",
    },
    "acceptance": {
        "test_requirements_surfaces.py": ROOT / "tests/acceptance/test_requirements_surfaces.py",
        "test_analysis_surfaces.py": ROOT / "tests/acceptance/test_analysis_surfaces.py",
    },
}
REQUIRED_TEST_HASHES = {
    suite: {
        name: hashlib.sha256(path.read_bytes()).hexdigest()
        for name, path in paths.items()
    }
    for suite, paths in REQUIRED_TEST_HASH_PATHS.items()
}
REQUIREMENTS_EVIDENCE_FIELDS = {
    "schema_version", "fixture_sha256", "corpus_cases", "dense_cases",
    "concurrent_calls", "long_sparse_cases", "request_error_cases",
}
HASH_RE = re.compile(r"^[0-9a-f]{64}$")
SHA_RE = re.compile(r"^[0-9a-f]{40}$")
WHEEL_CONTRACT_KEYS = {
    "spl_toolkit/" + path.relative_to(ROOT).as_posix()
    for path in (ROOT / "contracts").rglob("*") if path.is_file()
}
TOOLING_SOURCE_HASHES = {
    line: hashlib.sha256((ROOT / line).read_bytes()).hexdigest()
    for line in (ROOT / "python/native-source-files.txt").read_text(encoding="utf-8").splitlines()
    if line and not line.startswith("#")
}
TOOLING_FIXTURE_KEYS = {
    "contracts.json", "graph-cases.json", "impact-cases.json", "requests.json",
    "sarif-cases.json", "example-corpus.json", "example-target.json",
    "example-corpus-missing-file.json", "../rewrite/forms.json",
}
COMMON_FIELDS = {"schema_version", "kind", "source_sha", "status"}
KIND_FIELDS = {
    "native": {"target", "architecture", "environment"},
    "reproducibility": {
        "target", "architecture", "wheel_sha256", "artifact_hashes",
        "accepted_payloads", "environment", "checks",
    },
    "installed-wheel": {
        "target", "architecture", "python_version", "python_runtime", "wheel_sha256", "venv_prefix",
        "installed_module", "package_version", "native_version", "tests",
        "cli_examples", "surface_parity", "version_agreement", "required_test_files",
        "required_test_hashes",
        "wheel_contract_hashes", "tooling_source_hashes", "tooling_fixture_hashes",
        "machine_contract_tests", "fixture_hashes", "requirements_surface_evidence",
    },
    "go-floor": {"go_version"},
    "native-memory": {"compiler", "sanitizer"},
    "clean-source": {"tracked_hash_sha256"},
    "docker-examples": {"cli", "python_native", "server", "make_workflows"},
}
OPTIONAL_KIND_FIELDS = {
    "installed-wheel": {"loaded_library", "native_sha256", "wheel_payload_hashes", "source_header_sha256",
                        "fixture_hashes", "schema_surface_evidence", "spl2_surface_evidence", "rewrite_surface_evidence", "documentation_hashes"},
}
SINGLETONS = {"go-floor", "native-memory", "clean-source", "docker-examples"}
ENVIRONMENT_FIELDS = {
    "configured_runner", "observed_runner_name", "observed_image_os", "observed_image_version",
    "evidence_kind", "pinned_environment", "target", "os", "architecture", "go", "python",
    "packaging", "cc", "cc_version", "xcode", "linker_path", "linker", "sdk", "zlib",
    "wheel_platform", "source_date_epoch", "goflags", "cgo_enabled_cli_server",
    "cgo_enabled_native", "native_argv", "cgo_cflags", "cgo_cppflags", "cgo_ldflags", "artifacts",
}


def load_records(directory: Path) -> tuple[list[dict], list[str]]:
    records: list[dict] = []
    errors: list[str] = []
    if not directory.is_dir():
        return [], [f"evidence directory does not exist: {directory}"]
    for path in sorted(directory.rglob("*.json")):
        try:
            value = json.loads(path.read_text(encoding="utf-8"))
        except (OSError, json.JSONDecodeError) as error:
            errors.append(f"{path}: malformed JSON: {error}")
            continue
        if not isinstance(value, dict):
            errors.append(f"{path}: evidence must be one JSON object")
            continue
        records.append(value)
    if not records and not errors:
        errors.append(f"no evidence JSON files found in {directory}")
    return records, errors


def _key(record: dict) -> tuple[object, ...] | None:
    kind = record.get("kind")
    if kind in ("native", "reproducibility"):
        return kind, record.get("target")
    if kind == "installed-wheel":
        return kind, record.get("target"), record.get("python_version")
    if kind in SINGLETONS:
        return (kind,)
    return None


def _within(module: object, prefix: object) -> bool:
    if not isinstance(module, str) or not isinstance(prefix, str) or not module or not prefix:
        return False
    module_path = PurePosixPath(module.replace("\\", "/"))
    prefix_path = PurePosixPath(prefix.replace("\\", "/"))
    try:
        module_path.relative_to(prefix_path)
        return True
    except ValueError:
        return False


def _validate_counts(record: dict, errors: list[str], label: str) -> None:
    tests = record.get("tests")
    if not isinstance(tests, dict) or set(tests) != set(EXPECTED_COUNTS):
        errors.append(f"{label}: tests must contain required_native and surface_acceptance")
        return
    for suite, minimum in EXPECTED_COUNTS.items():
        counts = tests[suite]
        if not isinstance(counts, dict) or set(counts) != {"collected", "passed", "failed", "skipped"}:
            errors.append(f"{label}: {suite} has invalid count fields")
            continue
        if any(type(counts[name]) is not int or counts[name] < 0 for name in counts):
            errors.append(f"{label}: {suite} counts must be nonnegative integers")
            continue
        if counts["collected"] < minimum:
            errors.append(f"{label}: {suite} collected {counts['collected']}, expected at least {minimum}")
        if counts["passed"] != counts["collected"]:
            errors.append(f"{label}: {suite} passed does not equal collected")
        if counts["failed"]:
            errors.append(f"{label}: {suite} has failed tests")
        if counts["skipped"]:
            errors.append(f"{label}: {suite} has skipped tests")


def _validate_hash_map(value: object, expected: set[str] | dict[str, str], field: str,
                       errors: list[str], label: str) -> None:
    expected_paths = set(expected)
    if not isinstance(value, dict) or set(value) != expected_paths:
        errors.append(f"{label}: {field} has incorrect paths")
        return
    if any(not isinstance(digest, str) or not HASH_RE.fullmatch(digest) for digest in value.values()):
        errors.append(f"{label}: {field} contains an invalid SHA-256")
        return
    if isinstance(expected, dict):
        for name, digest in value.items():
            if digest != expected[name]:
                errors.append(f"{label}: {field} {name} differs from current source")


def _validate_required_test_hashes(record: dict, errors: list[str], label: str) -> None:
    value = record.get("required_test_hashes")
    if not isinstance(value, dict) or set(value) != set(REQUIRED_TEST_HASHES):
        errors.append(f"{label}: required_test_hashes must contain native and acceptance")
        return
    for suite, expected in REQUIRED_TEST_HASHES.items():
        hashes = value[suite]
        if not isinstance(hashes, dict) or set(hashes) != set(expected):
            errors.append(f"{label}: required_test_hashes {suite} has incorrect paths")
            continue
        for name, digest in hashes.items():
            if not isinstance(digest, str) or not HASH_RE.fullmatch(digest):
                errors.append(
                    f"{label}: required_test_hashes {suite}/{name} has an invalid SHA-256"
                )
            elif digest != expected[name]:
                errors.append(
                    f"{label}: required_test_hashes {suite}/{name} differs from current source"
                )


def _validate_requirements_evidence(record: dict, errors: list[str], label: str) -> None:
    fixture_hashes = record.get("fixture_hashes")
    if not isinstance(fixture_hashes, dict):
        errors.append(f"{label}: fixture_hashes must be an object")
        return
    requirement_hash = fixture_hashes.get("requirements")
    if not isinstance(requirement_hash, str) or not HASH_RE.fullmatch(requirement_hash):
        errors.append(f"{label}: requirements fixture hash must be a lowercase SHA-256")

    evidence = record.get("requirements_surface_evidence")
    if not isinstance(evidence, dict) or set(evidence) != REQUIREMENTS_EVIDENCE_FIELDS:
        errors.append(f"{label}: requirements_surface_evidence has incorrect fields")
        return
    if evidence.get("schema_version") != 1:
        errors.append(f"{label}: requirements_surface_evidence schema_version must be 1")
    if evidence.get("fixture_sha256") != requirement_hash:
        errors.append(f"{label}: requirements fixture hash differs from surface evidence")
    minimums = {
        "corpus_cases": 24,
        "dense_cases": 4,
        "concurrent_calls": 16,
        "long_sparse_cases": 2,
        "request_error_cases": 3,
    }
    for field, minimum in minimums.items():
        if type(evidence.get(field)) is not int or evidence[field] < minimum:
            errors.append(f"{label}: requirements_surface_evidence {field} is below {minimum}")


def _normalized_architecture(value: object) -> str:
    normalized = str(value).lower()
    if normalized in ("amd64", "x86_64"):
        return "x86_64"
    if normalized in ("arm64", "aarch64"):
        return "arm64"
    return normalized


def _validate_release_environment(
    environment: object, target: object, errors: list[str], label: str,
    artifact_hashes: object | None = None,
) -> None:
    if not isinstance(environment, dict):
        errors.append(f"{label}: environment must be an object")
        return
    missing = sorted(ENVIRONMENT_FIELDS - set(environment))
    extra = sorted(set(environment) - ENVIRONMENT_FIELDS)
    if missing:
        errors.append(f"{label}: environment missing fields: {', '.join(missing)}")
    if extra:
        errors.append(f"{label}: environment unknown fields: {', '.join(extra)}")
    if target not in TARGETS:
        return
    expected = TARGETS[target]
    expected_architecture = "arm64" if expected["goarch"] == "arm64" else "x86_64"
    if environment.get("target") != target:
        errors.append(f"{label}: environment target does not match record target")
    if _normalized_architecture(environment.get("architecture")) != expected_architecture:
        errors.append(f"{label}: environment architecture does not match {target}")
    if environment.get("configured_runner") != expected["runner"]:
        errors.append(f"{label}: environment configured_runner does not match {target}")
    evidence_kind = environment.get("evidence_kind")
    if environment.get("pinned_environment") is not True:
        errors.append(f"{label}: environment is not pinned")
    if evidence_kind not in ("pinned-runner", "pinned-local"):
        errors.append(f"{label}: diagnostic or unknown environment evidence_kind is not accepted")
    image_os = environment.get("observed_image_os")
    image_version = environment.get("observed_image_version")
    if evidence_kind == "pinned-runner":
        expected_image = "ubuntu24" if target.startswith("linux-") else "macos15" if target.startswith("darwin-") else "win22"
        if image_os != expected_image or not isinstance(image_version, str) or image_version == "unavailable":
            errors.append(f"{label}: observed runner image does not match {target}")
    elif evidence_kind == "pinned-local" and (image_os, image_version) != ("unavailable", "unavailable"):
        errors.append(f"{label}: pinned-local runner image identity must be unavailable")
    if not isinstance(environment.get("observed_runner_name"), str) or not environment["observed_runner_name"]:
        errors.append(f"{label}: observed_runner_name is missing")
    if not isinstance(environment.get("os"), str) or not environment["os"]:
        errors.append(f"{label}: observed OS identity is missing")
    go = environment.get("go")
    if not isinstance(go, str) or f"go{CONFIG['go']}" not in go.split():
        errors.append(f"{label}: environment Go version does not match {CONFIG['go']}")
    if environment.get("python") != CONFIG["build_python"]:
        errors.append(f"{label}: environment Python version does not match {CONFIG['build_python']}")
    packaging = environment.get("packaging")
    if not isinstance(packaging, dict) or set(packaging) != {"setuptools", "wheel", "build", "packaging"} or not all(isinstance(value, str) and value for value in packaging.values()):
        errors.append(f"{label}: environment packaging identity is incomplete")
    for field in ("cc", "cc_version", "linker_path", "linker", "zlib"):
        if not isinstance(environment.get(field), str) or not environment[field]:
            errors.append(f"{label}: environment {field} is missing")
    compiler = str(environment.get("cc_version", "")).lower()
    if target.startswith("linux-") and ("gcc" not in compiler or "13" not in compiler):
        errors.append(f"{label}: environment compiler is not Linux GCC 13")
    if target.startswith("windows-") and ("gcc" not in compiler or expected["gcc_version"] not in compiler):
        errors.append(f"{label}: environment compiler is not the pinned MinGW GCC")
    if target.startswith("darwin-"):
        if "clang" not in compiler or not all(isinstance(environment.get(field), str) and environment[field] for field in ("xcode", "sdk")):
            errors.append(f"{label}: environment Xcode compiler/SDK identity is incomplete")
    elif environment.get("xcode") is not None or environment.get("sdk") is not None:
        errors.append(f"{label}: non-Darwin environment has Xcode/SDK identity")
    if environment.get("wheel_platform") != expected["wheel_platform"]:
        errors.append(f"{label}: environment wheel platform does not match {target}")
    if type(environment.get("source_date_epoch")) is not int or environment["source_date_epoch"] <= 0:
        errors.append(f"{label}: environment source_date_epoch is invalid")
    goflags = environment.get("goflags")
    if not isinstance(goflags, list) or not {"-mod=readonly", "-trimpath", "-buildvcs=false", "-buildid="}.issubset(goflags):
        errors.append(f"{label}: environment Go flags are incomplete")
    if environment.get("cgo_enabled_cli_server") != "0" or environment.get("cgo_enabled_native") != "1":
        errors.append(f"{label}: environment CGO modes are invalid")
    if not isinstance(environment.get("native_argv"), list) or not environment["native_argv"]:
        errors.append(f"{label}: environment native build argv is missing")
    for field in ("cgo_cflags", "cgo_cppflags", "cgo_ldflags"):
        if not isinstance(environment.get(field), str):
            errors.append(f"{label}: environment {field} must be a string")
    hashes = environment.get("artifacts")
    if not isinstance(hashes, dict) or not hashes or any(not isinstance(value, str) or not HASH_RE.fullmatch(value) for value in hashes.values()):
        errors.append(f"{label}: environment artifacts contains an invalid SHA-256")
    if artifact_hashes is not None and hashes != artifact_hashes:
        errors.append(f"{label}: environment artifacts do not match artifact_hashes")


def validate_records(records: list[dict], source_sha: str) -> list[str]:
    errors: list[str] = []
    if not SHA_RE.fullmatch(source_sha):
        errors.append(f"requested source_sha is invalid: {source_sha!r}")
    seen: set[tuple[object, ...]] = set()
    by_key: dict[tuple[object, ...], dict] = {}

    for index, record in enumerate(records):
        label = f"record {index}"
        if not isinstance(record, dict):
            errors.append(f"{label}: evidence must be an object")
            continue
        kind = record.get("kind")
        if kind not in KIND_FIELDS:
            errors.append(f"{label}: unknown kind {kind!r}")
            continue
        required = COMMON_FIELDS | KIND_FIELDS[kind]
        missing = sorted(required - set(record))
        extra = sorted(set(record) - required - OPTIONAL_KIND_FIELDS.get(kind, set()))
        if missing:
            errors.append(f"{label}: missing fields: {', '.join(missing)}")
        if extra:
            errors.append(f"{label}: unknown fields: {', '.join(extra)}")
        if record.get("schema_version") != 1:
            errors.append(f"{label}: unsupported schema_version {record.get('schema_version')!r}")
        if record.get("source_sha") != source_sha:
            errors.append(f"{label}: source_sha {record.get('source_sha')!r} does not match {source_sha}")
        if record.get("status") != "passed":
            errors.append(f"{label}: status must be passed")
        key = _key(record)
        if key in seen:
            errors.append(f"{label}: duplicate evidence key {key}")
        elif key is not None:
            seen.add(key)
            by_key[key] = record

        if kind in ("native", "reproducibility", "installed-wheel"):
            target = record.get("target")
            if target not in TARGETS:
                errors.append(f"{label}: unknown target {target!r}")
            else:
                expected_arch = "arm64" if TARGETS[target]["goarch"] == "arm64" else "x86_64"
                if record.get("architecture") != expected_arch:
                    errors.append(f"{label}: architecture does not match {target}")

        if kind == "native":
            _validate_release_environment(record.get("environment"), record.get("target"), errors, label)
        elif kind == "reproducibility":
            wheel_hash = record.get("wheel_sha256")
            hashes = record.get("artifact_hashes")
            wheel_entries = [] if not isinstance(hashes, dict) else [value for name, value in hashes.items() if name.endswith(".whl")]
            if not isinstance(wheel_hash, str) or not HASH_RE.fullmatch(wheel_hash):
                errors.append(f"{label}: wheel_sha256 must be a lowercase SHA-256")
            if len(wheel_entries) != 1 or wheel_entries[0] != wheel_hash:
                errors.append(f"{label}: wheel_sha256 does not match the single accepted wheel")
            if not isinstance(hashes, dict) or any(not isinstance(value, str) or not HASH_RE.fullmatch(value) for value in hashes.values()):
                errors.append(f"{label}: artifact_hashes contains an invalid SHA-256")
            _validate_release_environment(record.get("environment"), record.get("target"), errors, label, hashes)
            checks = record.get("checks")
            if not isinstance(checks, dict) or set(checks) != {"clean_source", "payloads", "package"} or any(value != "passed" for value in checks.values()):
                errors.append(f"{label}: reproducibility checks are incomplete")
            payloads = record.get("accepted_payloads")
            if not isinstance(payloads, dict) or set(payloads) != {"cli", "server", "native"} or not all(isinstance(v, str) and v for v in payloads.values()):
                errors.append(f"{label}: accepted_payloads is invalid")
        elif kind == "installed-wheel":
            if record.get("python_version") not in PYTHON_VERSIONS:
                errors.append(f"{label}: unexpected python_version {record.get('python_version')!r}")
            if not isinstance(record.get("wheel_sha256"), str) or not HASH_RE.fullmatch(record["wheel_sha256"]):
                errors.append(f"{label}: wheel_sha256 must be a lowercase SHA-256")
            if not _within(record.get("installed_module"), record.get("venv_prefix")):
                errors.append(f"{label}: installed_module is outside venv_prefix")
            if record.get("package_version") != record.get("native_version"):
                errors.append(f"{label}: package/native version disagreement")
            for gate in ("cli_examples", "surface_parity", "version_agreement"):
                if record.get(gate) != "passed":
                    errors.append(f"{label}: {gate} must be passed")
            _validate_counts(record, errors, label)
            _validate_requirements_evidence(record, errors, label)
            _validate_required_test_hashes(record, errors, label)
            for field, expected in (
                ("wheel_contract_hashes", WHEEL_CONTRACT_KEYS),
                ("tooling_source_hashes", TOOLING_SOURCE_HASHES),
                ("tooling_fixture_hashes", TOOLING_FIXTURE_KEYS),
            ):
                _validate_hash_map(record.get(field), expected, field, errors, label)
            contract_counts = record.get("machine_contract_tests")
            count_fields = {"collected", "passed", "failed", "skipped"}
            if not isinstance(contract_counts, dict) or set(contract_counts) != count_fields:
                errors.append(f"{label}: machine_contract_tests has invalid count fields")
            elif any(type(contract_counts[name]) is not int or contract_counts[name] < 0
                     for name in count_fields):
                errors.append(f"{label}: machine_contract_tests counts must be nonnegative integers")
            elif (contract_counts["collected"] < EXPECTED_MACHINE_CONTRACT_COUNT
                  or contract_counts["passed"] != contract_counts["collected"]
                  or contract_counts["failed"] or contract_counts["skipped"]):
                errors.append(f"{label}: machine_contract_tests must all pass without skips")
            registered = record.get("required_test_files")
            if not isinstance(registered, dict) or set(registered) != set(REQUIRED_TEST_FILES):
                errors.append(f"{label}: required_test_files must contain native and acceptance")
            else:
                for suite, required_files in REQUIRED_TEST_FILES.items():
                    names = registered[suite]
                    if not isinstance(names, list) or not all(isinstance(name, str) for name in names) or len(names) != len(set(names)):
                        errors.append(f"{label}: {suite} required_test_files must be distinct filenames")
                        continue
                    for missing in sorted(required_files - set(names)):
                        errors.append(f"{label}: {suite} missing required suite {missing}")
        elif kind == "go-floor" and record.get("go_version") != CONFIG["go_language_floor_test"]:
            errors.append(f"{label}: go-floor must use {CONFIG['go_language_floor_test']}")
        elif kind == "native-memory":
            if record.get("compiler") != "gcc-13" or record.get("sanitizer") != "address+leak":
                errors.append(f"{label}: native-memory must use gcc-13 AddressSanitizer with leak detection")
        elif kind == "clean-source" and not HASH_RE.fullmatch(str(record.get("tracked_hash_sha256", ""))):
            errors.append(f"{label}: tracked_hash_sha256 must be a lowercase SHA-256")
        elif kind == "docker-examples":
            for gate in ("cli", "python_native", "server", "make_workflows"):
                if record.get(gate) != "passed":
                    errors.append(f"{label}: docker {gate} must be passed")

    required = {(kind, target) for kind in ("native", "reproducibility") for target in TARGETS}
    required |= {("installed-wheel", target, version) for target in TARGETS for version in PYTHON_VERSIONS}
    required |= {(kind,) for kind in SINGLETONS}
    for key in sorted(required - seen, key=str):
        errors.append(f"missing required evidence {key}")

    for target in TARGETS:
        release = by_key.get(("reproducibility", target))
        native = by_key.get(("native", target))
        if release and native and release.get("environment") != native.get("environment"):
            errors.append(f"native and reproducibility environment evidence differ for {target}")
        if not release:
            continue
        expected_hash = release.get("wheel_sha256")
        for version in PYTHON_VERSIONS:
            installed = by_key.get(("installed-wheel", target, version))
            if installed and installed.get("wheel_sha256") != expected_hash:
                errors.append(f"installed-wheel {target} {version}: wheel_sha256 differs from accepted reproducibility evidence")
    return errors


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--evidence", type=Path, required=True)
    parser.add_argument("--ref", default="HEAD")
    args = parser.parse_args()
    try:
        source_sha = subprocess.run(
            ["git", "rev-parse", "--verify", f"{args.ref}^{{commit}}"], cwd=ROOT,
            check=True, text=True, capture_output=True,
        ).stdout.strip()
    except subprocess.CalledProcessError as error:
        print(error.stderr, file=sys.stderr)
        return 1
    records, errors = load_records(args.evidence)
    errors.extend(validate_records(records, source_sha))
    if errors:
        for error in errors:
            print(error, file=sys.stderr)
        return 1
    print(f"milestone acceptance passed for {source_sha} with {len(records)} evidence records")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
