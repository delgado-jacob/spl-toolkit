"""Owned native closure result and Python wrapper behavior."""

from copy import deepcopy
import json
import os
import subprocess
import tempfile
from pathlib import Path

import pytest

from spl_toolkit import SPLMapper, SPLMapperError


ROOT = Path(__file__).resolve().parents[2]
FIXTURES = Path(os.environ.get("SPL_CLOSURE_FIXTURES", ROOT / "testdata/closure/cases.json"))
CASES = json.loads(FIXTURES.read_text(encoding="utf-8"))["cases"]


def mapper_kwargs():
    if "SPL_NATIVE_LIBRARY" in os.environ:
        return {"library_path": os.environ["SPL_NATIVE_LIBRARY"]}
    built = ROOT / "build/libspl_toolkit.dylib"
    return {"library_path": str(built)} if built.is_file() else {}


def test_closure_native_owned_result_and_errors():
    request = deepcopy(CASES[0]["request"])
    with SPLMapper(**mapper_kwargs()) as mapper:
        payload = json.dumps(request).encode("utf-8")
        native = mapper._lib.spl_mapper_closure_query
        pointer = native(mapper._mapper_id, payload)
        try:
            assert pointer and not pointer.contents.error
            expected = json.loads(pointer.contents.result)
            copied = deepcopy(expected)
        finally:
            mapper._lib.spl_result_free(pointer)
        assert copied == mapper.closure_query(request)
        assert copied["status"] == "valid"
        pointer = native(mapper._mapper_id, b"{}")
        try:
            assert pointer and pointer.contents.error and not pointer.contents.result
        finally:
            mapper._lib.spl_result_free(pointer)
        with pytest.raises(SPLMapperError):
            mapper.closure_query({"schema_version": 1})
        handle = mapper._mapper_id
        lib = mapper._lib
    pointer = lib.spl_mapper_closure_query(handle, payload)
    try:
        assert pointer and pointer.contents.error and not pointer.contents.result
    finally:
        lib.spl_result_free(pointer)


def test_closure_python_reports_content_statuses():
    with SPLMapper(**mapper_kwargs()) as mapper:
        assert mapper.closure_query(CASES[0]["request"])["status"] == "valid"
        assert mapper.closure_query(CASES[1]["request"])["status"] == "incomplete"


GO_REPORT_HELPER = r"""
package main
import (
    "encoding/json"
    "os"
    "github.com/delgado-jacob/spl-toolkit/pkg/closure"
)
func main() {
    var raw json.RawMessage
    if err := json.NewDecoder(os.Stdin).Decode(&raw); err != nil { panic(err) }
    request, err := closure.DecodeRequest(raw); if err != nil { panic(err) }
    report, err := closure.Evaluate(request); if err != nil { panic(err) }
    if err := json.NewEncoder(os.Stdout).Encode(report); err != nil { panic(err) }
}
"""


def test_installed_closure_matches_canonical_go_json():
    go_root = Path(os.environ.get("SPL_CLOSURE_GO_ROOT", ROOT))
    env = dict(os.environ, GOWORK="off")
    with tempfile.TemporaryDirectory(prefix="spl-closure-installed-") as folder:
        source = Path(folder) / "main.go"
        binary = Path(folder) / ("closure-report.exe" if os.name == "nt" else "closure-report")
        source.write_text(GO_REPORT_HELPER, encoding="utf-8")
        subprocess.run([os.environ.get("SPL_TOOLING_GO", "go"), "build", "-mod=readonly", "-o", str(binary), str(source)],
                       cwd=go_root, env=env, check=True)
        with SPLMapper(**mapper_kwargs()) as mapper:
            for case in CASES:
                request = case["request"]
                completed = subprocess.run([str(binary)], input=json.dumps(request).encode("utf-8"),
                                           cwd=go_root, env=env, check=True, capture_output=True)
                assert mapper.closure_query(request) == json.loads(completed.stdout), case["name"]
