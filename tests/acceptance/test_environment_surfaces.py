"""Full environment report parity across built CLI, HTTP, and native Python."""

import json
import os
import subprocess
from urllib.error import HTTPError
from urllib.request import Request, urlopen

import pytest

from spl_toolkit import SPLMapper
from test_surfaces import cli_path, required_absolute_path, server_url


FIXTURES = required_absolute_path("SPL_ENVIRONMENT_FIXTURES")
CASES = json.loads((FIXTURES / "cases.json").read_text(encoding="utf-8"))
EXITS = {"valid": 0, "partial": 3, "invalid": 1}


def open_mapper():
    options = {"library_path": str(required_absolute_path("SPL_NATIVE_LIBRARY"))} if "SPL_NATIVE_LIBRARY" in os.environ else {}
    return SPLMapper(**options)


def post_raw(server_url, raw):
    request = Request(server_url + "/environment/validate", data=raw,
                      headers={"Content-Type": "application/json"}, method="POST")
    try:
        with urlopen(request, timeout=5) as response:
            return response.status, json.load(response)
    except HTTPError as response:
        with response:
            return response.code, json.load(response)


def native_raw(mapper, raw):
    pointer = mapper._lib.spl_mapper_validate_environment(mapper._mapper_id, raw)
    assert pointer
    try:
        assert pointer.contents.error is None
        assert pointer.contents.result is not None
        return json.loads(pointer.contents.result)
    finally:
        mapper._lib.spl_result_free(pointer)


GO_REPORT_HELPER = r"""
package main
import (
    "encoding/json"
    "io"
    "os"
    "github.com/delgado-jacob/spl-toolkit/pkg/environment"
)
func main() {
    raw, err := io.ReadAll(os.Stdin); if err != nil { panic(err) }
    report, err := environment.ValidateJSON(raw); if err != nil { panic(err) }
    if err := json.NewEncoder(os.Stdout).Encode(report); err != nil { panic(err) }
}
"""


@pytest.fixture(scope="module")
def environment_go_reporter(tmp_path_factory):
    # Reuse the package checker's isolated native-source closure, as the installed
    # closure/requirements suites do; no checkout module or Python import is needed.
    root = required_absolute_path("SPL_TOOLING_SOURCE_ROOT")
    directory = tmp_path_factory.mktemp("environment-go")
    helper = directory / "main.go"
    helper.write_text(GO_REPORT_HELPER, encoding="utf-8")
    binary = directory / ("environment-report.exe" if os.name == "nt" else "environment-report")
    env = dict(os.environ, GOWORK="off")
    subprocess.run([os.environ.get("SPL_TOOLING_GO", "go"), "build", "-mod=readonly", "-o", str(binary), str(helper)],
                   cwd=root, env=env, check=True, capture_output=True)
    return binary, root, env


@pytest.mark.parametrize("case", CASES, ids=lambda case: case["name"])
def test_environment_full_report_parity(case, cli_path, server_url, environment_go_reporter, tmp_path):
    snapshot = tmp_path / "snapshot.json"
    schemas = tmp_path / "schemas.json"
    snapshot.write_text(json.dumps(case["snapshot"]), encoding="utf-8")
    schemas.write_text(json.dumps(case["schema_bundle"]), encoding="utf-8")
    request = {"schema_version": 1, "snapshot": case["snapshot"], "schema_bundle": case["schema_bundle"]}
    raw = json.dumps(request).encode("utf-8")
    completed = subprocess.run(
        [str(cli_path), "environment", "validate", "--snapshot", str(snapshot),
         "--schemas", str(schemas), "--format", "json"],
        capture_output=True, text=True, timeout=10)
    assert (completed.returncode, completed.stderr) == (EXITS[case["status"]], "")
    cli_report = json.loads(completed.stdout)
    http_status, http_report = post_raw(server_url, raw)
    with open_mapper() as mapper:
        python_report = mapper.validate_environment(request)
        c_report = native_raw(mapper, raw)
    assert http_status == (400 if case["status"] == "invalid" else 200)
    binary, root, env = environment_go_reporter
    go_output = subprocess.run([str(binary)], input=raw, cwd=root, env=env,
                               check=True, capture_output=True, timeout=10).stdout
    assert json.loads(go_output) == cli_report == http_report == c_report == python_report
    assert cli_report["status"] == case["status"]


@pytest.mark.parametrize("artifact", ["snapshot", "schema_bundle"])
def test_environment_malformed_raw_artifact_is_attributed_to_file(artifact, cli_path, tmp_path):
    path = tmp_path / "broken.json"
    path.write_bytes(b'{"schema_version":1,}')
    flag = "--snapshot" if artifact == "snapshot" else "--schemas"
    completed = subprocess.run([str(cli_path), "environment", "validate", flag, str(path), "--format", "json"],
                               capture_output=True, text=True, timeout=10)
    assert (completed.returncode, completed.stderr) == (1, "")
    report = json.loads(completed.stdout)
    assert report["status"] == "invalid"
    assert report["diagnostics"][0]["artifact"] == artifact


def test_environment_malformed_inline_request_is_attributed_to_request(server_url):
    raw = b'{"schema_version":1,"snapshot":{"schema_version":1,},"schema_bundle":{}}'
    http_status, http_report = post_raw(server_url, raw)
    with open_mapper() as mapper:
        c_report = native_raw(mapper, raw)
    assert http_status == 400
    assert http_report == c_report
    assert c_report["status"] == "invalid"
    assert c_report["diagnostics"][0]["artifact"] == "request"
