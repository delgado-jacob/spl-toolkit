"""Canonical requirement values across Go, CLI, HTTP, C, and Python."""

from __future__ import annotations

from concurrent.futures import ThreadPoolExecutor
import hashlib
import json
import os
from pathlib import Path
import subprocess
from urllib.error import HTTPError
from urllib.request import Request, urlopen

import pytest

from spl_toolkit import SPLMapper
from test_surfaces import cli_path, required_absolute_path, server_url


EXITS = {"valid": 0, "invalid": 1, "incomplete": 3}
SELECTORS = (
    ("language", "--language"),
    ("profile", "--profile"),
    ("version", "--compatibility-version"),
    ("source_id", "--source-id"),
)
RESOURCE_CODE = "SPL_ANALYSIS_RESOURCE_LIMIT"
GO_HELPER = r'''
package main

import (
	"encoding/json"
	"os"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
)

type reports struct {
	Analysis     *analysis.Result         `json:"analysis"`
	Requirements *analysis.RequirementSet `json:"requirements"`
}

func main() {
	var documents []analysis.QueryDocument
	if err := json.NewDecoder(os.Stdin).Decode(&documents); err != nil {
		panic(err)
	}
	results := make([]reports, 0, len(documents))
	for _, document := range documents {
		analyzed, err := analysis.Analyze(document)
		if err != nil {
			panic(err)
		}
		requirements, err := analysis.Requirements(document)
		if err != nil {
			panic(err)
		}
		results = append(results, reports{Analysis: analyzed, Requirements: requirements})
	}
	if err := json.NewEncoder(os.Stdout).Encode(results); err != nil {
		panic(err)
	}
}
'''


def open_mapper() -> SPLMapper:
    options = {}
    if "SPL_NATIVE_LIBRARY" in os.environ:
        options["library_path"] = str(required_absolute_path("SPL_NATIVE_LIBRARY"))
    return SPLMapper(**options)


def dense_query(size: int, language: str) -> str:
    pattern = "FROM main | fields a* | " if language == "spl2" else "fields a* | "
    return (pattern * (size // len(pattern) + 1))[:size]


def normalized_document(document: dict) -> dict:
    return {
        "text": document["text"],
        "language": document.get("language") or "spl",
        "profile": document.get("profile") or "splunkd",
        "version": document.get("version") or "current",
        "source_id": document.get("source_id") or "",
    }


def compact_bytes(value: dict) -> bytes:
    return json.dumps(value, ensure_ascii=False, separators=(",", ":")).encode("utf-8")


def run_cli(cli: Path, operation: str, document: dict) -> tuple[int, dict, bytes]:
    args = [str(cli), operation, "--format", "json", "--query", document["text"]]
    for key, flag in SELECTORS:
        if key in document:
            args.extend([flag, document[key]])
    completed = subprocess.run(args, capture_output=True, check=False, timeout=20)
    assert completed.stderr == b"", (operation, document.get("source_id"), completed.stderr)
    assert completed.stdout, (operation, document.get("source_id"), completed.returncode)
    return completed.returncode, json.loads(completed.stdout), completed.stdout.rstrip(b"\n")


def post_raw(server: str, operation: str, payload: bytes) -> tuple[int, dict, bytes]:
    request = Request(
        server + "/query/" + operation,
        data=payload,
        headers={"Content-Type": "application/json"},
        method="POST",
    )
    try:
        with urlopen(request, timeout=20) as response:
            body = response.read()
            return response.status, json.loads(body), body.rstrip(b"\n")
    except HTTPError as error:
        with error:
            body = error.read()
        return error.code, json.loads(body), body.rstrip(b"\n")


def post_document(server: str, operation: str, document: dict) -> tuple[int, dict, bytes]:
    return post_raw(server, operation, compact_bytes(document))


def call_raw_c(mapper: SPLMapper, operation: str, document: dict) -> dict:
    pointer = getattr(mapper._lib, "spl_mapper_" + operation)(
        mapper._mapper_id, compact_bytes(document)
    )
    try:
        assert pointer and not pointer.contents.error and pointer.contents.result
        return json.loads(pointer.contents.result)
    finally:
        mapper._lib.spl_result_free(pointer)


def call_python(mapper: SPLMapper, operation: str, document: dict) -> dict:
    options = {key: value for key, value in document.items() if key != "text"}
    return getattr(mapper, operation)(document["text"], **options)


def requirement_fixture_path() -> Path:
    return required_absolute_path("SPL_REQUIREMENTS_FIXTURES")


@pytest.fixture(scope="session")
def requirement_cases() -> list[dict]:
    data = json.loads(requirement_fixture_path().read_text(encoding="utf-8"))
    assert data["version"] == "1" and len(data["cases"]) == 17
    return data["cases"]


@pytest.fixture(scope="session")
def special_documents() -> list[dict]:
    dense = [
        {
            "text": dense_query(size, language),
            "language": language,
            "source_id": f"dense-{language}-{size}.spl",
        }
        for language in ("spl", "spl2")
        for size in (65_533, 256 * 1024)
    ]
    sparse = [
        {
            "text": ("FROM main" if language == "spl2" else "search host=x")
            + " " * 300_000,
            "language": language,
            "source_id": f"sparse-{language}.spl",
        }
        for language in ("spl", "spl2")
    ]
    return dense + sparse


@pytest.fixture(scope="session")
def go_reports(requirement_cases, special_documents, tmp_path_factory) -> dict[str, dict]:
    root_value = os.environ.get("SPL_REQUIREMENTS_GO_ROOT")
    root = Path(root_value) if root_value else Path(__file__).resolve().parents[2]
    assert root.is_absolute() and (root / "go.mod").is_file(), "missing copied Go requirement source root"
    documents = [case["document"] for case in requirement_cases] + special_documents
    helper = tmp_path_factory.mktemp("requirements-go") / "main.go"
    helper.write_text(GO_HELPER, encoding="utf-8")
    env = os.environ.copy()
    env["GOWORK"] = "off"
    env["GOCACHE"] = str(tmp_path_factory.mktemp("requirements-go-cache"))
    completed = subprocess.run(
        [os.environ.get("SPL_TOOLING_GO", "go"), "run", "-mod=readonly", str(helper)],
        cwd=root,
        env=env,
        input=json.dumps(documents, ensure_ascii=False),
        capture_output=True,
        text=True,
        check=True,
        timeout=90,
    )
    values = json.loads(completed.stdout)
    assert len(values) == len(documents)
    return {document["source_id"]: value for document, value in zip(documents, values, strict=True)}


@pytest.fixture(scope="session")
def requirement_evidence():
    fixture_hash = hashlib.sha256(requirement_fixture_path().read_bytes()).hexdigest()
    record = {
        "schema_version": 1,
        "fixture_sha256": fixture_hash,
        "corpus_cases": 0,
        "dense_cases": 0,
        "concurrent_calls": 0,
        "long_sparse_cases": 0,
        "request_error_cases": 0,
    }
    yield record
    if destination := os.environ.get("SPL_REQUIREMENTS_EVIDENCE"):
        Path(destination).write_text(
            json.dumps(record, indent=2, sort_keys=True) + "\n", encoding="utf-8"
        )


def assert_resource_limit(analysis: dict, requirements: dict, document: dict) -> None:
    normalized = normalized_document(document)
    assert analysis["document"] == normalized
    assert analysis["status"] == "incomplete"
    assert analysis["coverage"] == {
        "syntax_complete": False,
        "semantic_complete": False,
        "reasons": [RESOURCE_CODE],
    }
    assert analysis["stages"] == analysis["scopes"] == []
    assert analysis["references"] == analysis["lineage"] == []
    assert all(value == [] for value in analysis["dependencies"].values())
    assert analysis["requirements"] == requirements
    assert requirements["query_status"] == "incomplete"
    assert requirements["coverage"] == {"complete": False, "reasons": [RESOURCE_CODE]}
    assert requirements["items"] == []
    assert requirements["query"]["query_digest"] == "sha256:" + hashlib.sha256(
        document["text"].encode("utf-8")
    ).hexdigest()
    assert len(requirements["diagnostics"]) == len(analysis["diagnostics"]) == 1
    assert requirements["diagnostics"] == analysis["diagnostics"]
    assert requirements["diagnostics"][0]["code"] == RESOURCE_CODE
    assert requirements["diagnostics"][0]["message"] == (
        "analysis stopped before parser prediction after reaching the 4,096-unit lexer work limit"
    )
    assert requirements["gaps"] == [{
        "code": RESOURCE_CODE,
        "message": "requirement coverage is incomplete because analysis exceeded the 4,096-unit lexer work limit",
        "reference_ids": [],
        "diagnostic_codes": [RESOURCE_CODE],
    }]


def test_requirements_fixture_matches_go_cli_http_and_python(
    requirement_cases, go_reports, cli_path, server_url, requirement_evidence
):
    with open_mapper() as mapper:
        for case in requirement_cases:
            document = case["document"]
            go = go_reports[document["source_id"]]
            assert go["requirements"] == case["expected"], case["id"]
            assert go["analysis"]["requirements"] == case["expected"], case["id"]

            py_analysis = call_python(mapper, "analyze_query", document)
            py_requirements = call_python(mapper, "requirements_query", document)
            c_analysis = call_raw_c(mapper, "analyze_query", document)
            c_requirements = call_raw_c(mapper, "requirements_query", document)
            analysis_exit, cli_analysis, _ = run_cli(cli_path, "analyze", document)
            requirement_exit, cli_requirements, _ = run_cli(cli_path, "requirements", document)
            analysis_status, http_analysis, _ = post_document(server_url, "analyze", document)
            requirement_status, http_requirements, _ = post_document(server_url, "requirements", document)

            assert analysis_exit == EXITS[go["analysis"]["status"]]
            assert requirement_exit == EXITS[go["requirements"]["query_status"]]
            assert analysis_status == requirement_status == 200
            assert py_analysis == c_analysis == cli_analysis == http_analysis == go["analysis"]
            assert py_requirements == c_requirements == cli_requirements == http_requirements == go["requirements"]

        first = requirement_cases[0]["document"]
        detached = call_python(mapper, "requirements_query", first)
        detached["items"].clear()
        assert call_python(mapper, "requirements_query", first) == requirement_cases[0]["expected"]
    requirement_evidence["corpus_cases"] = len(requirement_cases)


def test_dense_and_sparse_requirements_match_every_surface(
    special_documents, go_reports, cli_path, server_url, requirement_evidence
):
    dense_count = 0
    sparse_count = 0
    with open_mapper() as mapper:
        for document in special_documents:
            go = go_reports[document["source_id"]]
            py_analysis = call_python(mapper, "analyze_query", document)
            py_requirements = call_python(mapper, "requirements_query", document)
            c_analysis = call_raw_c(mapper, "analyze_query", document)
            c_requirements = call_raw_c(mapper, "requirements_query", document)
            analysis_exit, cli_analysis, cli_analysis_bytes = run_cli(cli_path, "analyze", document)
            requirement_exit, cli_requirements, cli_requirement_bytes = run_cli(
                cli_path, "requirements", document
            )
            analysis_status, http_analysis, http_analysis_bytes = post_document(
                server_url, "analyze", document
            )
            requirement_status, http_requirements, http_requirement_bytes = post_document(
                server_url, "requirements", document
            )
            assert py_analysis == c_analysis == cli_analysis == http_analysis == go["analysis"]
            assert py_requirements == c_requirements == cli_requirements == http_requirements == go["requirements"]
            assert analysis_status == requirement_status == 200

            if document["source_id"].startswith("dense-"):
                dense_count += 1
                assert analysis_exit == requirement_exit == 3
                assert_resource_limit(go["analysis"], go["requirements"], document)
                assert len(compact_bytes(go["requirements"])) <= 4096
                assert len(compact_bytes(go["analysis"])) <= len(document["text"].encode()) + 4096
                assert len(cli_requirement_bytes) <= 4096
                assert len(http_requirement_bytes) <= 4096
                assert len(cli_analysis_bytes) <= len(document["text"].encode()) + 4096
                assert len(http_analysis_bytes) <= len(document["text"].encode()) + 4096
            else:
                sparse_count += 1
                assert analysis_exit == EXITS[go["analysis"]["status"]]
                assert requirement_exit == EXITS[go["requirements"]["query_status"]]
                assert not any(
                    diagnostic["code"] == RESOURCE_CODE for diagnostic in go["analysis"]["diagnostics"]
                )
                assert go["analysis"]["stages"]
    requirement_evidence["dense_cases"] = dense_count
    requirement_evidence["long_sparse_cases"] = sparse_count


def test_concurrent_dense_results_are_deterministic_bounded_and_owned(
    special_documents, go_reports, cli_path, server_url, requirement_evidence
):
    dense = [document for document in special_documents if document["source_id"].startswith("dense-")]
    calls = 16
    with open_mapper() as mapper:
        expected = {document["source_id"]: go_reports[document["source_id"]] for document in dense}

        def exercise(index: int) -> tuple[bytes, bytes]:
            document = dense[index % len(dense)]
            want = expected[document["source_id"]]
            py_analysis = call_python(mapper, "analyze_query", document)
            py_requirements = call_python(mapper, "requirements_query", document)
            c_analysis = call_raw_c(mapper, "analyze_query", document)
            c_requirements = call_raw_c(mapper, "requirements_query", document)
            analysis_exit, cli_analysis, cli_analysis_bytes = run_cli(cli_path, "analyze", document)
            requirement_exit, cli_requirements, cli_requirement_bytes = run_cli(
                cli_path, "requirements", document
            )
            analysis_status, http_analysis, http_analysis_bytes = post_document(
                server_url, "analyze", document
            )
            requirement_status, http_requirements, http_requirement_bytes = post_document(
                server_url, "requirements", document
            )
            assert py_analysis == c_analysis == cli_analysis == http_analysis == want["analysis"]
            assert py_requirements == c_requirements == cli_requirements == http_requirements == want["requirements"]
            assert analysis_exit == requirement_exit == 3
            assert analysis_status == requirement_status == 200
            assert len(cli_requirement_bytes) <= 4096 and len(http_requirement_bytes) <= 4096
            limit = len(document["text"].encode()) + 4096
            assert len(cli_analysis_bytes) <= limit and len(http_analysis_bytes) <= limit
            return compact_bytes(py_analysis), compact_bytes(py_requirements)

        with ThreadPoolExecutor(max_workers=4) as pool:
            observed = list(pool.map(exercise, range(calls)))
    for index, (analysis_bytes, requirement_bytes) in enumerate(observed):
        document = dense[index % len(dense)]
        want = expected[document["source_id"]]
        assert analysis_bytes == compact_bytes(want["analysis"])
        assert requirement_bytes == compact_bytes(want["requirements"])
    requirement_evidence["concurrent_calls"] = calls


def test_requirements_real_http_request_errors_and_one_mebibyte_boundary(
    server_url, requirement_evidence
):
    valid = b'{"text":"search host=web"}'
    exact = valid + b" " * (1024 * 1024 - len(valid))
    for operation, status_key in (("analyze", "status"), ("requirements", "query_status")):
        status, report, _ = post_raw(server_url, operation, exact)
        assert status == 200 and report[status_key] == "valid"

    errors = 0
    for operation in ("analyze", "requirements"):
        for payload in (b"{", b'{"text":"one","text":"two"}', exact + b" "):
            status, body, _ = post_raw(server_url, operation, payload)
            assert status == 400
            assert body.get("error")
            assert not {
                "schema_version", "query", "query_status", "coverage", "items", "gaps", "diagnostics"
            }.intersection(body)
            errors += 1
    requirement_evidence["request_error_cases"] = errors
