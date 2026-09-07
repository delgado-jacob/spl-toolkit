#!/usr/bin/env python3
"""Require complete, current evidence for the Milestone 1 release baseline."""

from __future__ import annotations

import argparse
import json
from pathlib import Path, PurePosixPath
import re
import subprocess
import sys


ROOT = Path(__file__).resolve().parents[1]
CONFIG = json.loads((ROOT / "tools" / "release-env.json").read_text(encoding="utf-8"))
TARGETS = CONFIG["targets"]
PYTHON_VERSIONS = tuple(CONFIG["test_python"])
EXPECTED_COUNTS = {"required_native": 11, "surface_acceptance": 5}
HASH_RE = re.compile(r"^[0-9a-f]{64}$")
SHA_RE = re.compile(r"^[0-9a-f]{40}$")
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
        "cli_examples", "surface_parity", "version_agreement",
    },
    "go-floor": {"go_version"},
    "native-memory": {"compiler", "sanitizer"},
    "clean-source": {"tracked_hash_sha256"},
    "docker-examples": {"cli", "python_native", "server", "make_workflows"},
}
SINGLETONS = {"go-floor", "native-memory", "clean-source", "docker-examples"}


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
        extra = sorted(set(record) - required)
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

        if kind == "native" and not isinstance(record.get("environment"), dict):
            errors.append(f"{label}: environment must be an object")
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
            if not isinstance(record.get("environment"), dict) or record["environment"].get("pinned_environment") is not True:
                errors.append(f"{label}: reproducibility environment is not pinned")
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
