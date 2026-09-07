"""Exact shared Go reports through installed CLI, native Python, and HTTP."""

from __future__ import annotations

import json
from pathlib import Path
import subprocess

import pytest

from spl_toolkit import SPLMapper, SPLMapperError
from test_surfaces import cli_path, post_json, required_absolute_path, server_url


EXITS = {"valid": 0, "invalid": 1, "incomplete": 3}


@pytest.fixture(scope="session")
def validation_cases() -> list[dict]:
    fixture = json.loads(required_absolute_path("SPL_VALIDATION_FIXTURES").read_text(encoding="utf-8"))
    assert type(fixture["schema_version"]) is int and fixture["schema_version"] == 1
    assert fixture["cases"]
    return fixture["cases"]


def cli_arguments(cli_path: Path, catalog_path: Path, document: dict) -> list[str]:
    args = [str(cli_path), "validate-fields", "--fields", str(catalog_path), "--format", "json"]
    for key, flag in (("language", "--language"), ("profile", "--profile"),
                      ("version", "--compatibility-version"), ("source_id", "--source-id")):
        if key in document:
            args.extend([flag, document[key]])
    return args


def test_validation_full_report_parity(validation_cases, cli_path, server_url, tmp_path):
    catalog_path = tmp_path / "fields.json"
    with SPLMapper() as mapper:
        for case in validation_cases:
            document, catalog, expected = case["document"], case["catalog"], case["expected"]
            catalog_path.write_text(json.dumps(catalog), encoding="utf-8")
            args = cli_arguments(cli_path, catalog_path, document) + ["--query", document["text"]]
            completed = subprocess.run(args, capture_output=True, text=True, check=False)
            assert completed.returncode == EXITS[expected["status"]], case["id"]
            assert completed.stderr == "", case["id"]
            assert json.loads(completed.stdout) == expected, case["id"]
            assert mapper.validate_fields(
                document["text"], catalog, language=document.get("language", "spl"),
                profile=document.get("profile", "splunkd"), version=document.get("version", "current"),
                source_id=document.get("source_id", ""),
            ) == expected, case["id"]
            status, report = post_json(server_url, "/query/validate-fields", {"document": document, "catalog": catalog})
            assert status == 200, case["id"]
            assert report == expected, case["id"]


def test_validation_batch_full_report_parity(validation_cases, cli_path, server_url, tmp_path):
    cases = {case["id"]: case for case in validation_cases}
    selected = [cases[name] for name in ("empty_removal", "missing", "unknown_command")]
    catalog = selected[0]["catalog"]
    assert all(case["catalog"] == catalog for case in selected)
    documents = [case["document"] for case in selected]
    expected = {"schema_version": 1, "status": "invalid", "reports": [case["expected"] for case in selected]}
    catalog_path, batch_path = tmp_path / "fields.json", tmp_path / "batch.json"
    catalog_path.write_text(json.dumps(catalog), encoding="utf-8")
    batch_path.write_text(json.dumps(documents), encoding="utf-8")
    args = cli_arguments(cli_path, catalog_path, {})
    for source, stdin in ((str(batch_path), None), ("-", batch_path.read_text(encoding="utf-8"))):
        completed = subprocess.run(args + ["--batch", source], input=stdin, capture_output=True, text=True, check=False)
        assert (completed.returncode, completed.stderr) == (1, "")
        assert json.loads(completed.stdout) == expected
    with SPLMapper() as mapper:
        assert mapper.validate_fields_batch(documents, catalog) == expected
    status, report = post_json(server_url, "/query/validate-fields/batch", {"documents": documents, "catalog": catalog})
    assert status == 200
    assert report == expected


def test_validation_file_stdin_and_output_preserve_full_reports(validation_cases, cli_path, tmp_path):
    catalog_path, query_path, output_path = (tmp_path / name for name in ("fields.json", "query.spl", "report.json"))
    for case in validation_cases:
        document, expected = case["document"], case["expected"]
        catalog_path.write_text(json.dumps(case["catalog"]), encoding="utf-8")
        query_path.write_bytes(document["text"].encode("utf-8"))
        args = cli_arguments(cli_path, catalog_path, document)
        for source in (["--file", str(query_path)], ["--stdin"]):
            completed = subprocess.run(args + source + ["--output", str(output_path)],
                                       input=query_path.read_bytes(), capture_output=True, check=False)
            assert (completed.returncode, completed.stdout, completed.stderr) == (EXITS[expected["status"]], b"", b""), case["id"]
            assert json.loads(output_path.read_text(encoding="utf-8")) == expected, case["id"]


@pytest.mark.parametrize("catalog", [None, {}, ["host", "host"], {"fields": ["host"], "optional_fields": ["host"]}, {"fields": [], "extra": 1}])
def test_validation_catalog_request_errors_are_not_content_reports(catalog, cli_path, server_url, tmp_path):
    catalog_path = tmp_path / "bad-fields.json"
    catalog_path.write_text(json.dumps(catalog), encoding="utf-8")
    completed = subprocess.run(cli_arguments(cli_path, catalog_path, {}) + ["--query", "table host"],
                               capture_output=True, text=True, check=False)
    assert completed.returncode == 2
    assert completed.stdout == ""
    with SPLMapper() as mapper, pytest.raises(SPLMapperError):
        mapper.validate_fields("table host", catalog)
    status, report = post_json(server_url, "/query/validate-fields", {"document": {"text": "table host"}, "catalog": catalog})
    assert status == 400
    assert "analysis" not in report and "outcomes" not in report


@pytest.mark.parametrize("documents", [[], [{"text": "table host", "profile": "unsupported"}], [{"source_id": "missing-text"}]])
def test_validation_batch_request_errors_are_atomic(documents, cli_path, server_url, tmp_path):
    catalog_path = tmp_path / "fields.json"
    catalog_path.write_text('["host"]', encoding="utf-8")
    completed = subprocess.run(cli_arguments(cli_path, catalog_path, {}) + ["--batch", "-"], input=json.dumps(documents),
                               capture_output=True, text=True, check=False)
    assert completed.returncode == 2
    assert completed.stdout == ""
    with SPLMapper() as mapper, pytest.raises(SPLMapperError):
        mapper.validate_fields_batch(documents, ["host"])
    status, report = post_json(server_url, "/query/validate-fields/batch", {"documents": documents, "catalog": ["host"]})
    assert status == 400
    assert "reports" not in report
