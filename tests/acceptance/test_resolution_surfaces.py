"""Authored resolution semantics and full reports across five real product surfaces."""
import copy
import hashlib
import json
import os
from pathlib import Path
import shutil
import subprocess
from urllib.error import HTTPError
from urllib.request import Request, urlopen

import pytest
from spl_toolkit import SPLMapper, SPLMapperError
# Import resource fixtures directly: unrelated suites require their own fixtures.
from test_surfaces import cli_path, required_absolute_path, server_url

FIXTURES = required_absolute_path("SPL_RESOLUTION_FIXTURES")
LIBRARY = required_absolute_path("SPL_NATIVE_LIBRARY")
assert FIXTURES.is_file() and LIBRARY.is_file()
CASES = json.loads(FIXTURES.read_text(encoding="utf-8"))
assert len(CASES) == 16

def open_mapper():
    return SPLMapper(library_path=str(LIBRARY))


def make_request(mapper, case):
    document = copy.deepcopy(case["document"])
    analysis = mapper.analyze_query(document["text"], **{key: document[key] for key in
                                  ("language", "profile", "version", "source_id")})
    request = {"schema_version": 1, "document": document,
               "resolutions": [{"placeholder": item["marker"], "kind": item["kind"],
                                "values": copy.deepcopy(item["values"])} for item in case["resolutions"]],
               "compatibility": copy.deepcopy(case["compatibility"])}
    if request["compatibility"].get("schema_bundle") is None:
        request["compatibility"].pop("schema_bundle", None)
    if "max_variants" in case:
        request["max_variants"] = case["max_variants"]
    for binding in request["compatibility"]["input_bindings"]:
        selector = binding.pop("original_input")
        matches = [item for item in analysis["inputs"] if item["kind"] == selector["kind"]
                   and item["identity"] == selector["identity"]]
        assert len(matches) == 1, selector
        binding["original_input_id"] = matches[0]["id"]
    return request


def assert_authored_report(result, case):
    expected = case["expected"]
    assert result["counts"] == expected["counts"]
    assert result["generated_count"] == len(expected["variants"])
    assert result["total_combinations"] == str(len(expected["variants"]))
    assert len(result["variants"]) == len(expected["variants"])
    for ordinal, (got, want) in enumerate(zip(result["variants"], expected["variants"]), 1):
        assert got["ordinal"] == ordinal
        assert [choice["value"] for choice in got["selection"]] == want["selection"]
        assert got["outcome"] == want["outcome"]
        assert got.get("resolved_query") == want["resolved_query"]
        if want["resolved_query"] is None:
            assert "resolved_query" not in got
        if got["outcome"] == "verified":
            assert got["proof"]["proven"]
            assert got["compatibility"]["outcome"] == "satisfied"
        for change in got["changes"]:
            location = change["original_location"]
            assert case["document"]["text"].encode()[location["start"]["offset"]:
                                                       location["end"]["offset"]].decode() == change["before"]
            if "candidate_location" in change:
                location = change["candidate_location"]
                assert got["candidate_text"].encode()[location["start"]["offset"]:
                                                      location["end"]["offset"]].decode() == change["after"]


def raw_result(mapper, raw):
    pointer = mapper._lib.spl_mapper_resolve(mapper._mapper_id, raw)
    assert pointer
    try:
        if pointer.contents.error is not None:
            assert pointer.contents.result is None
            return "error", json.loads(pointer.contents.error)
        assert pointer.contents.result is not None
        return "report", json.loads(pointer.contents.result)
    finally:
        mapper._lib.spl_result_free(pointer)



GO_DRIVER = r"""
package main
import (
 "encoding/json"
 "io"
 "os"
 "github.com/delgado-jacob/spl-toolkit/pkg/resolution"
)
func main() {
 raw, err := io.ReadAll(os.Stdin); if err != nil { panic(err) }
 report, err := resolution.ResolveJSON(raw)
 if err != nil {
  detail, ok := resolution.RequestErrorDetails(err); if !ok { panic(err) }
  if err := json.NewEncoder(os.Stdout).Encode(detail); err != nil { panic(err) }
  return
 }
 if err := json.NewEncoder(os.Stdout).Encode(report); err != nil { panic(err) }
}
"""

@pytest.fixture(scope="module")
def resolution_go_driver(tmp_path_factory):
    # Installed runs use the copied, hashed module owner. Source runs locate the
    # module containing this suite, never an ambient parent workspace.
    value = os.environ.get("SPL_TOOLING_SOURCE_ROOT")
    owner = required_absolute_path("SPL_TOOLING_SOURCE_ROOT") if value else Path(__file__).resolve().parents[2]
    assert (owner / "go.mod").is_file()
    directory = tmp_path_factory.mktemp("resolution-go")
    try:
        helper = directory / "main.go"
        helper.write_text(GO_DRIVER, encoding="utf-8")
        binary = directory / ("resolution-report.exe" if os.name == "nt" else "resolution-report")
        env = dict(os.environ, GOWORK="off")
        subprocess.run([os.environ.get("SPL_TOOLING_GO", "go"), "build", "-mod=readonly", "-o", str(binary), str(helper)],
                       cwd=owner, env=env, check=True, capture_output=True)
        yield binary, owner, env
    finally:
        shutil.rmtree(directory)

@pytest.fixture(scope="module")
def resolution_evidence():
    record = {"schema_version": 1, "source_sha": os.environ.get("SPL_RESOLUTION_SOURCE_SHA"),
              "fixture_sha256": hashlib.sha256(FIXTURES.read_bytes()).hexdigest(),
              "corpus_cases": 0, "malformed_cases": 0,
              "surfaces": ["go", "cli", "http", "c", "python"]}
    yield record
    if destination := os.environ.get("SPL_RESOLUTION_EVIDENCE"):
        Path(destination).write_text(json.dumps(record, indent=2, sort_keys=True) + "\n", encoding="utf-8")


def assert_authored(value, case):
    if case["expected"].get("request_error"):
        assert "variants" not in value
        assert value["code"] == case["expected"]["request_error"]
        assert value["total_combinations"] == str(case["expected"]["combination_count"])
        assert isinstance(value["path"], str) and value["message"]
    else:
        assert_authored_report(value, case)


def exercise(raw, request, authored, error, cli_path, server_url, driver, tmp_path):
    binary, owner, env = driver
    go = json.loads(subprocess.run([str(binary)], input=raw, cwd=owner, env=env,
                                  check=True, capture_output=True, timeout=30).stdout)
    authored(go)
    path = tmp_path / "request.json"
    path.write_bytes(raw)
    cli = subprocess.run([str(cli_path), "resolve", "--request", str(path), "--format", "json"],
                         capture_output=True, timeout=30)
    assert cli.returncode == (2 if error else 1 if go["counts"]["failed"] else 3 if go["counts"]["incomplete"] else 0)
    assert (cli.stdout if error else cli.stderr) == b""
    cli_value = json.loads(cli.stderr if error else cli.stdout)
    authored(cli_value)
    http_request = Request(server_url + "/query/resolve", data=raw,
                           headers={"Content-Type": "application/json"}, method="POST")
    try:
        response = urlopen(http_request, timeout=30)
    except HTTPError as response_error:
        response = response_error
    with response:
        assert response.status == (400 if error else 200)
        http = json.load(response)
    authored(http)
    with open_mapper() as mapper:
        kind, c = raw_result(mapper, raw)
        assert kind == ("error" if error else "report")
        authored(c)
        if error:
            with pytest.raises(SPLMapperError) as raised:
                mapper.resolve(request)
            python = json.loads(str(raised.value))
        else:
            python = mapper.resolve(request)
        authored(python)
    # Go is a report producer; independently authored facts above are the oracle.
    assert go == cli_value == http == c == python


@pytest.mark.parametrize("case", CASES, ids=lambda case: case["id"])
def test_resolution_full_report_parity(case, cli_path, server_url, resolution_go_driver, tmp_path, resolution_evidence):
    with open_mapper() as mapper:
        request = make_request(mapper, case)
    exercise(json.dumps(request).encode(), request, lambda report: assert_authored(report, case),
             bool(case["expected"].get("request_error")), cli_path, server_url, resolution_go_driver, tmp_path)
    resolution_evidence["corpus_cases"] += 1


@pytest.mark.parametrize("raw", [b"", b"{", b"{}", b"\xff", b'{"schema_version":1,"unknown":true}'])
def test_resolution_malformed_parity(raw, cli_path, server_url, resolution_go_driver, tmp_path, resolution_evidence):
    def authored(detail):
        assert detail["code"] == "request_invalid" and detail["message"]
        if raw in (b"", b"{", b"\xff"):
            assert isinstance(detail["byte_offset"], int)
    # Python accepts dicts, so use its raw decoder boundary for malformed bytes.
    # Its valid JSON invalid-request delegate is covered across all five below.
    if raw in (b"", b"{", b"\xff"):
        binary, owner, env = resolution_go_driver
        go = json.loads(subprocess.run([str(binary)], input=raw, cwd=owner, env=env,
                                      check=True, capture_output=True).stdout)
        authored(go)
        path = tmp_path / "malformed.json"
        path.write_bytes(raw)
        cli = subprocess.run([str(cli_path), "resolve", "--request", str(path), "--format", "json"], capture_output=True)
        assert (cli.returncode, cli.stdout) == (2, b"")
        cli_value = json.loads(cli.stderr)
        authored(cli_value)
        req = Request(server_url + "/query/resolve", data=raw, headers={"Content-Type": "application/json"})
        with pytest.raises(HTTPError) as raised:
            urlopen(req, timeout=10)
        with raised.value as response:
            assert response.status == 400
            http = json.load(response)
        authored(http)
        with open_mapper() as mapper:
            kind, c = raw_result(mapper, raw)
            assert kind == "error"
            authored(c)
        assert go == cli_value == http == c
    else:
        exercise(raw, json.loads(raw), authored, True, cli_path, server_url, resolution_go_driver, tmp_path)
    resolution_evidence["malformed_cases"] += 1
