"""Tooling source/data closure survives real staging and rejects damaged payloads."""
from pathlib import Path
import zipfile

import pytest

from tools import check_package
from tools.tests.test_package import load_build_support

ROOT = Path(__file__).resolve().parents[2]


def test_tooling_source_and_contract_closure_stages_offline(tmp_path):
    support = load_build_support()
    support.stage_native_source(ROOT, tmp_path / "source", ROOT / "python/native-source-files.txt")
    for package in ("corpus", "corpusio", "document", "graph", "sarif", "impact", "compatibility", "resolution", "workflow"):
        for source in (ROOT / "pkg" / package).glob("*.go"):
            if not source.name.endswith("_test.go"):
                copied = tmp_path / "source" / source.relative_to(ROOT)
                assert copied.read_bytes() == source.read_bytes()
    for source in (ROOT / "contracts").rglob("*"):
        if source.is_file():
            assert (tmp_path / "source" / source.relative_to(ROOT)).read_bytes() == source.read_bytes()


@pytest.mark.parametrize("contract", ["contracts/sarif/sarif-schema-2.1.0.json", "contracts/v1/compatibility-request.schema.json", "contracts/v1/compatibility.schema.json", "contracts/v1/resolution-request.schema.json", "contracts/v1/resolution.schema.json"])
def test_tooling_contract_wheel_hashes_reject_missing_or_changed_schema(tmp_path, contract):
    wheel = tmp_path / "test.whl"
    expected = {"spl_toolkit/" + p.relative_to(ROOT).as_posix(): p.read_bytes()
                for p in (ROOT / "contracts").rglob("*") if p.is_file()}
    for damage in (None, "missing", "changed"):
        entries = dict(expected)
        key = "spl_toolkit/" + contract
        if damage == "missing":
            del entries[key]
        elif damage == "changed":
            entries[key] = b"{}"
        with zipfile.ZipFile(wheel, "w") as archive:
            for name, payload in entries.items():
                archive.writestr(name, payload)
        if damage is None:
            assert set(check_package.verify_wheel_contracts(wheel, ROOT)) == set(expected)
        else:
            with pytest.raises((KeyError, AssertionError)):
                check_package.verify_wheel_contracts(wheel, ROOT)


def test_resolution_sources_tests_and_fixtures_are_registered(monkeypatch):
    native = set((ROOT / "python/native-source-files.txt").read_text().splitlines())
    release = set((ROOT / "tools/release-source-files.txt").read_text().splitlines())
    for source in (ROOT / "pkg").rglob("resolution*.go"):
        if not source.name.endswith("_test.go"):
            assert source.relative_to(ROOT).as_posix() in native & release
    assert "test_native_resolution.py" in check_package.NATIVE_TESTS
    assert "tests/test_native_resolution.py" in check_package.SDIST_FIXED_FILES
    assert "test_resolution_surfaces.py" in check_package.ACCEPTANCE_FILES
    monkeypatch.setenv("SPL_RESOLUTION_FIXTURES", "/host/untrusted/cases.json")
    env = check_package.clean_env()
    assert "SPL_RESOLUTION_FIXTURES" not in env


def test_workflow_sources_tests_and_examples_are_registered():
    assert "test_native_workflow.py" in check_package.NATIVE_TESTS
    assert "tests/test_native_workflow.py" in check_package.SDIST_FIXED_FILES
    assert "test_workflow_surfaces.py" in check_package.ACCEPTANCE_FILES
    manifest = set((ROOT / "python/native-source-files.txt").read_text().splitlines())
    for source in (ROOT / "pkg/workflow").glob("*.go"):
        if not source.name.endswith("_test.go"):
            assert source.relative_to(ROOT).as_posix() in manifest
    assert not any("testdata/workflow" in path for path in manifest)
