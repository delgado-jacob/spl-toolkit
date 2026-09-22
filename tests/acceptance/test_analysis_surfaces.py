"""Full canonical reports through real CLI, HTTP, and installed native Python.

The same reviewed fixture values are asserted directly by Go's corpus tests.
Only JSON object key order is ignored; no report values are normalized here.
"""

from __future__ import annotations

import json
import os
from pathlib import Path
import subprocess
from urllib.error import HTTPError
from urllib.request import Request, urlopen

import pytest

from spl_toolkit import SPLMapper, __version__
from test_surfaces import (
    cli_path,
    get_json,
    post_json,
    required_absolute_path,
    server_url,
)


EXPECTED_VERSION = os.environ.get("SPL_EXPECTED_VERSION")
CAPABILITY_DIMENSIONS = ("syntax", "semantics", "requirements", "linting", "safe_rewriting")
CAPABILITY_STATES = ("supported", "partial", "unsupported", "not_applicable", "unassessed")


def mapper_kwargs() -> dict[str, str]:
    return {"library_path": str(required_absolute_path("SPL_NATIVE_LIBRARY"))} if "SPL_NATIVE_LIBRARY" in os.environ else {}


def raw_native_capabilities(mapper: SPLMapper, language: str) -> dict:
    if language == "spl":
        pointer = mapper._lib.spl_mapper_capabilities(mapper._mapper_id)
    else:
        options = json.dumps({"language": language, "profile": "splunkd", "version": "current"}).encode()
        pointer = mapper._lib.spl_mapper_capabilities_for(mapper._mapper_id, options)
    try:
        assert pointer and not pointer.contents.error
        return json.loads(pointer.contents.result)
    finally:
        mapper._lib.spl_result_free(pointer)


def source_tree_version() -> str | None:
    for root in Path(__file__).resolve().parents:
        version_file = root / "VERSION"
        if (root / "go.mod").is_file() and version_file.is_file():
            version = version_file.read_text(encoding="utf-8").strip()
            assert version
            return version
    return None


def expected_toolkit_version(_mapper: SPLMapper) -> str:
    expected = EXPECTED_VERSION or source_tree_version()
    if expected:
        return expected
    assert __version__ != "dev", "expected tagged version is unavailable outside a source tree"
    return __version__


def assert_capability_ledger_is_self_consistent(manifest: dict, language: str, expected_version: str) -> None:
    assert manifest["toolkit_version"] == expected_version
    assert (manifest["language"], manifest["profile"], manifest["version"]) == (language, "splunkd", "current")
    assert manifest["records"] and manifest["evidence"]
    assert set(manifest["summary"]) == set(CAPABILITY_DIMENSIONS)

    recomputed = {
        dimension: {
            "applicable": 0, "covered": 0, "supported": 0, "partial": 0,
            "unsupported": 0, "not_applicable": 0, "unassessed": 0,
        }
        for dimension in CAPABILITY_DIMENSIONS
    }
    cited = set()
    for record in manifest["records"]:
        assert record["language"] == language and record["profile"] == "splunkd"
        assert set(record["dimensions"]) == set(CAPABILITY_DIMENSIONS)
        for dimension in CAPABILITY_DIMENSIONS:
            claim = record["dimensions"][dimension]
            state = claim["state"]
            assert state in CAPABILITY_STATES
            recomputed[dimension][state] += 1
            if state != "not_applicable":
                recomputed[dimension]["applicable"] += 1
            if state == "supported":
                recomputed[dimension]["covered"] += 1
            cited.update(claim["evidence_ids"])
    assert recomputed == manifest["summary"]
    evidence_ids = {item["id"] for item in manifest["evidence"]}
    assert len(evidence_ids) == len(manifest["evidence"])
    assert cited <= evidence_ids


@pytest.fixture(scope="session")
def analysis_cases() -> list[dict]:
    path = required_absolute_path("SPL_ANALYSIS_FIXTURES")
    fixture = json.loads(path.read_text(encoding="utf-8"))
    assert fixture["version"] == "1" and len(fixture["cases"]) == 38
    return fixture["cases"]


def test_analysis_full_report_parity(analysis_cases: list[dict], cli_path: Path, server_url: str) -> None:
    with SPLMapper(**mapper_kwargs()) as mapper:
        for case in analysis_cases:
            document = case["document"]
            options = {key: value for key, value in document.items() if key != "text"}
            actual = mapper.analyze_query(document["text"], **options)
            assert actual == case["expected"], case["id"]

            status, response = post_json(server_url, "/query/analyze", document)
            assert status == 200, case["id"]
            assert response == case["expected"], case["id"]

            args = [str(cli_path), "analyze", "--query", document["text"], "--format", "json"]
            for key, flag in (("language", "--language"), ("profile", "--profile"),
                              ("version", "--compatibility-version"), ("source_id", "--source-id")):
                if key in document:
                    args.extend([flag, document[key]])
            completed = subprocess.run(args, capture_output=True, text=True, check=False)
            assert completed.stderr == "", case["id"]
            assert completed.returncode == {"valid": 0, "invalid": 1, "incomplete": 3}[actual["status"]], case["id"]
            assert json.loads(completed.stdout) == case["expected"], case["id"]


def test_analysis_capabilities_parity(cli_path: Path, server_url: str) -> None:
    with SPLMapper(**mapper_kwargs()) as mapper:
        expected_version = expected_toolkit_version(mapper)
        for language in ("spl", "spl2"):
            expected = mapper.capabilities(language=language, profile="splunkd", version="current")
            assert raw_native_capabilities(mapper, language) == expected
            assert_capability_ledger_is_self_consistent(expected, language, expected_version)
            assert type(expected["schema_version"]) is int and expected["schema_version"] == 1
            assert expected["commands"] and expected["functions"] and expected["rewrite"]["forms"]

            url = server_url + "/capabilities?language=" + language + "&profile=splunkd&version=current"
            assert get_json(url) == expected
            completed = subprocess.run(
                [str(cli_path), "capabilities", "--language", language, "--profile", "splunkd",
                 "--compatibility-version", "current", "--format", "json"],
                capture_output=True, text=True, check=False,
            )
            assert (completed.returncode, completed.stderr) == (0, "")
            assert json.loads(completed.stdout) == expected


@pytest.mark.parametrize("payload", [
    b'{"text":"\xff"}', b'{"text":"x","source_id":"\xff"}',
    b'{"text":"\\ud800"}', b'{"source_id":"\\udc00"}',
    b'{"text":"\\ud800x"}', b'{"text":"\\ud800\\u0041"}',
])
def test_analysis_http_rejects_lossy_unicode(server_url: str, payload: bytes) -> None:
    request = Request(server_url + "/query/analyze", data=payload,
                      headers={"Content-Type": "application/json"}, method="POST")
    with pytest.raises(HTTPError) as rejected:
        urlopen(request, timeout=5)
    with rejected.value as response:
        assert response.code == 400
        error = json.load(response)
    assert "error" in error and "schema_version" not in error


@pytest.mark.parametrize("query", ['search host="😀"', 'search host="\\ud800"'])
def test_analysis_http_valid_unicode_controls(server_url: str, query: str) -> None:
    document = {"text": query, "source_id": "é😀.spl"}
    with SPLMapper(**mapper_kwargs()) as mapper:
        expected = mapper.analyze_query(query, source_id=document["source_id"])
    # post_json emits escaped non-BMP pairs and escaped literal backslashes.
    status, actual = post_json(server_url, "/query/analyze", document)
    assert status == 200 and actual == expected
    assert actual["document"]["text"] == query
