"""The Go evaluator, CLI, REST, C ABI, and Python wrapper return one report."""

from copy import deepcopy
import json
import os
from pathlib import Path
import subprocess
import tempfile

import pytest

from spl_toolkit import SPLMapper


ROOT = Path(__file__).resolve().parents[2]
CASES = json.loads((ROOT / "testdata/closure/cases.json").read_text(encoding="utf-8"))["cases"]
GO_HELPER = r'''
package main
import (
    "bytes"
    "encoding/json"
    "io"
    "net/http"
    "net/http/httptest"
    "os"
    "github.com/delgado-jacob/spl-toolkit/pkg/api"
    "github.com/delgado-jacob/spl-toolkit/pkg/closure"
)
func main() {
    body, err := io.ReadAll(os.Stdin); if err != nil { panic(err) }
    request, err := closure.DecodeRequest(body); if err != nil { panic(err) }
    report, err := closure.Evaluate(request); if err != nil { panic(err) }
    r := httptest.NewRequest(http.MethodPost, "/api/v1/query/closure", bytes.NewReader(body))
    r.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()
    api.NewServer().Handler().ServeHTTP(w, r)
    var rest any
    if err := json.Unmarshal(w.Body.Bytes(), &rest); err != nil { panic(err) }
    out := map[string]any{"go": report, "rest": rest, "rest_status": w.Code}
    if err := json.NewEncoder(os.Stdout).Encode(out); err != nil { panic(err) }
}
'''


def fixture_requests():
    resolved = deepcopy(CASES[0]["request"])
    incomplete = deepcopy(CASES[1]["request"])
    invalid = {
        "schema_version": 1,
        "document": {"text": "`bad`", "language": "spl", "source_id": "root.spl"},
        "bundle": {"schema_version": 1, "scope_id": "synthetic", "collections": [
            {"kind": "macro", "coverage": "complete"}], "objects": [
            {"id": "bad", "kind": "macro", "name": "bad", "source_id": "bad.conf",
             "document": {"text": "eval =", "language": "spl"}, "arity": 0,
             "arguments": [], "relations": []}]},
        "bindings": [],
    }
    cycle = deepcopy(invalid)
    cycle["document"]["text"] = "`loop`"
    cycle["bundle"]["objects"][0].update({
        "id": "loop", "name": "loop", "source_id": "loop.conf",
        "document": {"text": "`loop`", "language": "spl"}})
    return [("resolved", resolved, "valid", 0),
            ("missing", incomplete, "incomplete", 3),
            ("invalid", invalid, "invalid", 1),
            ("cycle", cycle, "incomplete", 3)]


@pytest.fixture(scope="module")
def go_helper():
    with tempfile.TemporaryDirectory(prefix="spl-closure-parity-") as folder:
        source = Path(folder) / "main.go"
        binary = Path(folder) / "closure-parity"
        source.write_text(GO_HELPER, encoding="utf-8")
        env = dict(os.environ, GOWORK="off")
        subprocess.run(["go", "build", "-o", str(binary), str(source)], cwd=ROOT, env=env, check=True)
        cli_binary = Path(folder) / "spl-toolkit"
        subprocess.run(["go", "build", "-o", str(cli_binary), "./cmd"], cwd=ROOT, env=env, check=True)
        yield binary, cli_binary


def mapper_kwargs():
    if "SPL_NATIVE_LIBRARY" in os.environ:
        return {"library_path": os.environ["SPL_NATIVE_LIBRARY"]}
    built = ROOT / "build/libspl_toolkit.dylib"
    return {"library_path": str(built)} if built.is_file() else {}


@pytest.mark.parametrize("name,closure_request,status,exit_code", fixture_requests())
def test_closure_complete_cross_surface_values(go_helper, tmp_path, name, closure_request, status, exit_code):
    request_bytes = json.dumps(closure_request).encode("utf-8")
    env = dict(os.environ, GOWORK="off")
    direct = subprocess.run([str(go_helper[0])], input=request_bytes, cwd=ROOT,
                            capture_output=True, check=True, env=env)
    response = json.loads(direct.stdout.splitlines()[-1])
    expected = response["go"]
    assert expected["status"] == status
    assert response["rest_status"] == 200 and response["rest"] == expected
    bundle = tmp_path / "bundle.json"
    bundle.write_text(json.dumps(closure_request["bundle"]), encoding="utf-8")
    args = [str(go_helper[1]), "closure", "--bundle", str(bundle),
            "--query", closure_request["document"]["text"], "--format=json"]
    for source, option in (("language", "--language"), ("profile", "--profile"),
                           ("version", "--compatibility-version"), ("source_id", "--source-id")):
        if source in closure_request["document"]:
            args.extend([option, closure_request["document"][source]])
    if closure_request.get("bindings"):
        bindings = tmp_path / "bindings.json"
        bindings.write_text(json.dumps(closure_request["bindings"]), encoding="utf-8")
        args.extend(["--bindings", str(bindings)])
    cli = subprocess.run(args, cwd=ROOT, capture_output=True, env=env)
    assert cli.returncode == exit_code, cli.stderr.decode()
    assert json.loads(cli.stdout) == expected
    with SPLMapper(**mapper_kwargs()) as mapper:
        pointer = mapper._lib.spl_mapper_closure_query(mapper._mapper_id, request_bytes)
        try:
            assert pointer and not pointer.contents.error
            native = json.loads(pointer.contents.result)
        finally:
            mapper._lib.spl_result_free(pointer)
        assert native == expected
        assert mapper.closure_query(closure_request) == expected
    for projection, key in (("graph", "graph"), ("bom", "bom")):
        projected = subprocess.run(["--format=" + projection if arg == "--format=json" else arg for arg in args], cwd=ROOT,
                                   capture_output=True, env=env)
        assert projected.returncode == exit_code, projected.stderr.decode()
        assert json.loads(projected.stdout) == expected[key]
    if name == "cycle":
        assert any(edge["cycle_path"] for edge in expected["graph"]["edges"])
        assert any(entry["incomplete"] for entry in expected["bom"])
