"""Rewrite acceptance cannot disappear while predecessor evidence stays green."""
import json
import importlib.util
from copy import deepcopy
from pathlib import Path

import pytest

from tools import check_package
from tools.check_acceptance import validate_records
from tools.tests.test_acceptance import passing_records, SHA


ROOT = Path(__file__).resolve().parents[2]


@pytest.mark.parametrize("suite,name", [("native", "test_native_rewrite.py"),
                                        ("acceptance", "test_rewrite_surfaces.py")])
def test_rewrite_suite_cannot_be_omitted(suite, name):
    records = passing_records()
    installed = next(r for r in records if r["kind"] == "installed-wheel")
    installed["required_test_files"][suite] = [n for n in installed["required_test_files"][suite] if n != name]
    assert any("missing required suite " + name in e for e in validate_records(records, SHA))


def test_installed_rewrite_inputs_are_complete(tmp_path):
    copied = tmp_path / "rewrite"
    hashes = check_package.copy_rewrite_fixtures(ROOT / "testdata/rewrite", copied)
    assert set(hashes) == {p.name for p in (ROOT / "testdata/rewrite").glob("*.json")}
    assert "example-rules.json" in hashes and "corpus.json" in hashes
    assert "test_rewrite_surfaces.py" in check_package.ACCEPTANCE_FILES
    for name, digest in hashes.items():
        assert check_package.sha256(copied / name) == digest


def test_rewrite_documentation_required_for_installed_examples(tmp_path):
    source = tmp_path / "source"
    (source / "docs").mkdir(parents=True)
    for name in ("README.md", "docs/cli.md", "docs/spl2.md"):
        (source / name).write_text("Maintained document\n")
    with pytest.raises(FileNotFoundError, match="rewrite"):
        check_package.copy_documentation(source, tmp_path / "copy")


@pytest.fixture
def rewrite_transport_helpers(monkeypatch):
    # Reuse the independently hand-derived native oracle, never a live report.
    for directory in ("python", "python/tests", "tests/acceptance"):
        monkeypatch.syspath_prepend(str(ROOT / directory))
    for name, directory in (("SPL_REWRITE_FIXTURES", "rewrite"), ("SPL_SCHEMA_FIXTURES", "schemas")):
        monkeypatch.setenv(name, str(ROOT / "testdata" / directory))
    def load(name, path):
        spec = importlib.util.spec_from_file_location(name, ROOT / path)
        module = importlib.util.module_from_spec(spec)
        spec.loader.exec_module(module)
        return module
    return (load("rewrite_transport_helpers", "tests/acceptance/test_rewrite_surfaces.py"),
            load("rewrite_independent_oracle", "python/tests/test_native_rewrite.py"))


@pytest.mark.parametrize("damage", ["hash", "missing", "duplicate", "truncated", "sources", "fixtures", "semantics"])
def test_rewrite_transport_rejects_lost_or_stale_evidence(tmp_path, rewrite_transport_helpers, damage):
    transport, oracle = rewrite_transport_helpers
    report = oracle.expected_search("apply")
    for document in (report["document"], report["original_analysis"]["document"], report["candidate_analysis"]["document"]):
        document["source_id"] = "unit"
    request = {"schema_version": 1, "mode": "apply", "document": report["document"], "rules": oracle.RULES}
    case = {"id": "unit", "groups": sorted(transport.REQUIRED_GROUPS), "target": None,
            "request": request, "expected": transport.projection(report)}
    fixtures = tmp_path / "fixtures"
    fixtures.mkdir()
    (fixtures / "corpus.json").write_text(json.dumps([case]))
    artifact = {"schema_version": 1, "kind": "rewrite-go-transport", "conformance_credit": 0,
                "source_hashes": transport.source_hashes(ROOT),
                "fixture_hashes": {"corpus.json": transport.digest(fixtures / "corpus.json")},
                "reports": [{"id": "unit", "request": request, "report": report}]}
    path = tmp_path / "transport.json"
    path.write_text(json.dumps(artifact))
    expected = transport.digest(path)
    assert transport.load_transport(path, fixtures, expected, ROOT)[1] == {"unit": report}
    if damage == "missing":
        artifact["reports"].clear()
    elif damage == "duplicate":
        artifact["reports"].append(deepcopy(artifact["reports"][0]))
    elif damage == "truncated":
        del artifact["reports"][0]["report"]["original_analysis"]
    elif damage == "sources":
        artifact["source_hashes"]["go.mod"] = "0" * 64
    elif damage == "fixtures":
        artifact["fixture_hashes"]["corpus.json"] = "0" * 64
    elif damage == "semantics":
        artifact["reports"][0]["report"]["text"] = "search unexpected=x"
    path.write_text(json.dumps(artifact) + "\n")
    with pytest.raises(AssertionError):
        transport.load_transport(path, fixtures, expected if damage == "hash" else transport.digest(path), ROOT)


@pytest.mark.parametrize("damage", ["empty", "duplicate", "required-group"])
def test_rewrite_collection_requires_unique_nonempty_groups(tmp_path, rewrite_transport_helpers, damage):
    transport, _ = rewrite_transport_helpers
    cases = json.loads((ROOT / "testdata/rewrite/corpus.json").read_text())
    if damage == "empty":
        cases = []
    elif damage == "duplicate":
        cases.append(deepcopy(cases[0]))
    else:
        cases = [c for c in cases if "implicit" not in c["groups"]]
    (tmp_path / "corpus.json").write_text(json.dumps(cases))
    with pytest.raises(AssertionError):
        transport.load_cases(tmp_path)


@pytest.mark.parametrize("missing", ["condition-common-or", "condition-noncommon-or",
                                    "condition-not-positive", "overwritten-fact-unknown",
                                    "independent-child-no-parent-fact", "inherited-child-fact",
                                    "validation-schema-unresolved"])
def test_rewrite_collection_cannot_lose_a_semantic_obligation(tmp_path, rewrite_transport_helpers, missing):
    transport, _ = rewrite_transport_helpers
    cases = json.loads((ROOT / "testdata/rewrite/corpus.json").read_text())
    remaining = [case for case in cases if case["id"] != missing]
    assert len(remaining) == len(cases) - 1
    (tmp_path / "corpus.json").write_text(json.dumps(remaining))
    with pytest.raises(AssertionError, match="required rewrite group"):
        transport.load_cases(tmp_path)


def test_rewrite_collection_rejects_every_empty_required_subgroup(tmp_path, rewrite_transport_helpers):
    transport, _ = rewrite_transport_helpers
    cases = json.loads((ROOT / "testdata/rewrite/corpus.json").read_text())
    for group in transport.REQUIRED_GROUPS:
        remaining = [case for case in cases if group not in case["groups"]]
        assert len(remaining) < len(cases), group
        (tmp_path / "corpus.json").write_text(json.dumps(remaining))
        with pytest.raises(AssertionError):
            transport.load_cases(tmp_path)


@pytest.mark.parametrize("damage", ["hash", "path", "outcomes"])
def test_rewrite_local_catalog_and_outcomes_cannot_be_substituted(tmp_path, rewrite_transport_helpers, damage):
    transport, _ = rewrite_transport_helpers
    cases = json.loads((ROOT / "testdata/rewrite/corpus.json").read_text())
    case = next(c for c in cases if c["id"] == "validation-local-ocsf")
    if damage == "hash":
        case["catalog_raw_sha256"] = "0" * 64
    elif damage == "path":
        case["catalog_fixture"] = "https://outside.invalid/catalog"
    else:
        del case["expected"]["validation_summary"]
    (tmp_path / "corpus.json").write_text(json.dumps(cases))
    with pytest.raises(AssertionError):
        transport.load_cases(tmp_path)
