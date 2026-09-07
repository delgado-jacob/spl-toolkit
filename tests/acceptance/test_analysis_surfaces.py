"""Full canonical reports through real CLI, HTTP, and installed native Python.

The same reviewed fixture values are asserted directly by Go's corpus tests.
Only JSON object key order is ignored; no report values are normalized here.
"""

from __future__ import annotations

import json
from pathlib import Path
import subprocess
from urllib.error import HTTPError
from urllib.request import Request, urlopen

import pytest

from spl_toolkit import SPLMapper
from test_surfaces import (
    cli_path,
    get_json,
    post_json,
    required_absolute_path,
    server_url,
)


@pytest.fixture(scope="session")
def analysis_cases() -> list[dict]:
    path = required_absolute_path("SPL_ANALYSIS_FIXTURES")
    fixture = json.loads(path.read_text(encoding="utf-8"))
    assert fixture["version"] == "1" and fixture["cases"]
    return fixture["cases"]


def test_analysis_full_report_parity(analysis_cases: list[dict], cli_path: Path, server_url: str) -> None:
    with SPLMapper() as mapper:
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
    with SPLMapper() as mapper:
        expected = mapper.capabilities()
    assert type(expected["schema_version"]) is int and expected["schema_version"] == 1
    assert (expected["language"], expected["profile"], expected["version"]) == ("spl", "splunkd", "current")
    assert expected["commands"] and expected["functions"]
    assert get_json(server_url + "/capabilities") == expected
    completed = subprocess.run(
        [str(cli_path), "capabilities", "--format", "json"],
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
    with SPLMapper() as mapper:
        expected = mapper.analyze_query(query, source_id=document["source_id"])
    # post_json emits escaped non-BMP pairs and escaped literal backslashes.
    status, actual = post_json(server_url, "/query/analyze", document)
    assert status == 200 and actual == expected
    assert actual["document"]["text"] == query
