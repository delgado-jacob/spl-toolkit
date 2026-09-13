"""Canonical tooling through actual native allocations and guarded handles."""
from concurrent.futures import ThreadPoolExecutor
from copy import deepcopy
import ctypes
import json
import os
from pathlib import Path
import subprocess

import pytest

from spl_toolkit import SPLMapper, SPLMapperError
from spl_toolkit.exceptions import MapperNotFoundError

OPERATIONS = ("scan_corpus", "export_graph", "export_sarif", "impact_schema",
              "impact_mapping", "document_view")
DOCS = [{"id": "valid", "document": {"text": 'search host="😀"\r\n| table host', "source_id": "valid"}},
        {"id": "missing", "document": {"text": "search absent=x", "source_id": "missing"}},
        {"id": "unknown", "document": {"text": "| mystery | table host", "source_id": "unknown"}}]
BEFORE = {"kind": "field_list", "catalog": {"fields": ["host", "absent"]}}
AFTER = {"kind": "field_list", "catalog": {"fields": ["host"]}}


def request_for(operation):
    if operation == "document_view":
        return deepcopy(DOCS[0]["document"])
    request = {"schema_version": 1, "documents": deepcopy(DOCS)}
    if operation == "impact_schema":
        request.update(before_target=BEFORE, after_target=AFTER)
    elif operation == "impact_mapping":
        request.update(before_rules={"schema_version": 1, "rules": []},
                       after_rules={"schema_version": 1, "rules": []})
    else:
        request["validation_target"] = AFTER
    return deepcopy(request)


def mapper_kwargs():
    return {"library_path": os.environ["SPL_NATIVE_LIBRARY"]} if "SPL_NATIVE_LIBRARY" in os.environ else {}


def raw_result(mapper, operation, payload, handle=None, reject=False):
    pointer = getattr(mapper._lib, "spl_mapper_" + operation)(
        mapper._mapper_id if handle is None else handle, payload)
    try:
        assert pointer
        if reject:
            assert pointer.contents.error and pointer.contents.result is None
            return pointer.contents.error.decode()
        assert not pointer.contents.error, pointer.contents.error
        return json.loads(pointer.contents.result)
    finally:
        mapper._lib.spl_result_free(pointer)


@pytest.mark.parametrize("operation", OPERATIONS)
def test_native_tooling_reports_and_owned_results(operation):
    request = request_for(operation)
    original = deepcopy(request)
    with SPLMapper(**mapper_kwargs()) as mapper:
        report = getattr(mapper, operation)(request)
        assert report == raw_result(mapper, operation, json.dumps(request).encode())
        if operation == "export_sarif":
            assert report["version"] == "2.1.0"
            assert report["runs"][0]["results"]
        elif operation == "export_graph":
            assert report["nodes"] and report["edges"]
        elif operation == "document_view":
            assert report["document"]["text"] == request["text"]
            assert report["references"]
        elif operation.startswith("impact_"):
            assert report["counts"]["indeterminate"] == 1
            assert report["counts"]["affected"] == (1 if operation == "impact_schema" else 0)
            assert len(report["entries"]) == 3
        else:
            assert report["status"] == "invalid"
            assert len(report["entries"]) == 3
        expected = deepcopy(report)
        report.clear()
        assert getattr(mapper, operation)(request) == expected
    assert request == original


@pytest.mark.parametrize("operation", OPERATIONS)
@pytest.mark.parametrize("payload", [None, b"null", b"[]", b"{", b"{} {}",
                                     b'{"text":"\\ud800"}', b'{"text":"\xff"}',
                                     b'{"schema_version":1,"schema_version":1}',
                                     b'{"directory":"/"}'])
def test_raw_tooling_malformed_requests_are_owned_errors(operation, payload):
    with SPLMapper(**mapper_kwargs()) as mapper:
        raw_result(mapper, operation, payload, reject=True)
        assert getattr(mapper, operation)(request_for(operation))


@pytest.mark.parametrize("operation", OPERATIONS)
def test_tooling_closed_handles_and_input_bytes(operation):
    mapper = SPLMapper(**mapper_kwargs())
    request = request_for(operation)
    payload = ctypes.create_string_buffer(json.dumps(request).encode())
    before = payload.raw
    raw_result(mapper, operation, payload)
    assert payload.raw == before
    handle = mapper._mapper_id
    mapper.close()
    with pytest.raises(MapperNotFoundError):
        getattr(mapper, operation)(request)
    assert "Mapper not found" in raw_result(mapper, operation, payload, handle, reject=True)


@pytest.mark.parametrize("operation", OPERATIONS)
def test_tooling_invalid_python_json(operation):
    with SPLMapper(**mapper_kwargs()) as mapper:
        for invalid in ({"text": "\ud800"}, {"text": float("nan")}, {"text": object()}):
            with pytest.raises(SPLMapperError):
                getattr(mapper, operation)(invalid)


def test_concurrent_mixed_tooling_and_legacy_operations():
    with SPLMapper(**mapper_kwargs()) as mapper:
        expected = {op: getattr(mapper, op)(request_for(op)) for op in OPERATIONS}
        def call(index):
            op = OPERATIONS[index % len(OPERATIONS)]
            assert getattr(mapper, op)(request_for(op)) == expected[op]
            assert mapper.analyze_query("search host=x")["status"] == "valid"
        with ThreadPoolExecutor(max_workers=8) as pool:
            list(pool.map(call, range(48)))


@pytest.mark.parametrize("operation", OPERATIONS)
@pytest.mark.parametrize("outcome", ["success", "error", "decode"])
def test_tooling_releases_real_allocation_exactly_once(operation, outcome, monkeypatch):
    with SPLMapper(**mapper_kwargs()) as mapper:
        frees = []
        native_free = mapper._lib.spl_result_free
        def free(pointer):
            frees.append(bool(pointer.contents.error))
            native_free(pointer)
        monkeypatch.setattr(mapper._lib, "spl_result_free", free)
        if outcome == "decode":
            def fail_decode(_):
                raise ValueError("decode failed")
            monkeypatch.setattr("spl_toolkit.mapper.json.loads", fail_decode)
            with pytest.raises(ValueError, match="decode failed"):
                getattr(mapper, operation)(request_for(operation))
        elif outcome == "error":
            with pytest.raises(SPLMapperError):
                getattr(mapper, operation)({"unknown": True})
        else:
            assert getattr(mapper, operation)(request_for(operation))
        assert frees == [outcome == "error"]
        assert mapper._active_calls == 0


@pytest.mark.parametrize("operation", OPERATIONS)
def test_tooling_null_result_releases_admission(operation, monkeypatch):
    with SPLMapper(**mapper_kwargs()) as mapper:
        monkeypatch.setattr(mapper._lib, "spl_mapper_" + operation, lambda *_: None)
        with pytest.raises(SPLMapperError, match="returned no result"):
            getattr(mapper, operation)(request_for(operation))
        assert mapper._active_calls == 0


def test_tooling_durable_request_fixtures():
    fixtures = Path(os.environ["SPL_TOOLING_FIXTURES"])
    assert fixtures.is_absolute()
    cases = json.loads((fixtures / "requests.json").read_text())
    with SPLMapper(**mapper_kwargs()) as mapper:
        for case in cases:
            if case["kind"] != "corpus" or case["name"] == "duplicate document property":
                continue  # Duplicate JSON keys are exercised as raw bytes above.
            if case["valid"]:
                assert mapper.scan_corpus(case["request"])["entries"]
            else:
                with pytest.raises(SPLMapperError):
                    mapper.scan_corpus(case["request"])


# The independent Go executable calls public operations directly, bypassing the
# C exports and Python wrapper. Package verification stages only its Go closure.
GO_TOOLING = r'''
package main
import (
 "encoding/json"; "fmt"; "io"; "os"
 "github.com/delgado-jacob/spl-toolkit/pkg/analysis"
 "github.com/delgado-jacob/spl-toolkit/pkg/corpus"
 "github.com/delgado-jacob/spl-toolkit/pkg/document"
 "github.com/delgado-jacob/spl-toolkit/pkg/graph"
 "github.com/delgado-jacob/spl-toolkit/pkg/impact"
 "github.com/delgado-jacob/spl-toolkit/pkg/sarif"
 "github.com/delgado-jacob/spl-toolkit/pkg/validation"
)
func run(op string, raw []byte) (any,error) {
 switch op {
 case "document_view":
  ds,e:=validation.DecodeDocuments(append(append([]byte{'['},raw...),']')); if e!=nil{return nil,e}
  if len(ds)!=1{return nil,fmt.Errorf("expected one document")}
  a,e:=analysis.Analyze(ds[0]); if e!=nil{return nil,e}
  return document.New(a,document.RevisionContext{ToolVersion:"0.1.1",ContractVersion:"1"})
 case "impact_schema":
  r,e:=impact.DecodeSchemaRequest(raw); if e!=nil{return nil,e}; return impact.CompareSchemas(r)
 case "impact_mapping":
  r,e:=impact.DecodeMappingRequest(raw); if e!=nil{return nil,e}; return impact.CompareMappings(r)
 default:
  r,e:=corpus.DecodeRequest(raw); if e!=nil{return nil,e}
  c,e:=corpus.Scan(r); if e!=nil{return nil,e}
  switch op {case "scan_corpus":return c,nil;case "export_graph":return graph.Export(c);case "export_sarif":return sarif.Export(c)}
 }
 return nil,fmt.Errorf("unknown operation")
}
func main(){raw,e:=io.ReadAll(os.Stdin); if e!=nil{panic(e)}; v,e:=run(os.Args[1],raw)
 if e!=nil{v=map[string]string{"error":e.Error()}}
 if e=json.NewEncoder(os.Stdout).Encode(v);e!=nil{panic(e)}
}
'''


@pytest.fixture(scope="module")
def tooling_go(tmp_path_factory):
    directory = tmp_path_factory.mktemp("native-tooling-go")
    source = directory / "main.go"
    source.write_text(GO_TOOLING)
    binary = directory / ("tooling.exe" if os.name == "nt" else "tooling")
    root = Path(os.environ.get("SPL_TOOLING_SOURCE_ROOT", Path(__file__).resolve().parents[2]))
    ldflags = "-X=github.com/delgado-jacob/spl-toolkit/internal/buildinfo.Version=0.1.1"
    if os.name != "nt":
        ldflags = "-linkmode=external " + ldflags
    try:
        subprocess.run([os.environ.get("SPL_TOOLING_GO", "go"), "build", "-mod=readonly",
                        "-ldflags=" + ldflags,
                        "-o", str(binary), str(source)],
                       cwd=root, check=True, capture_output=True, text=True)
    except subprocess.CalledProcessError as error:
        pytest.fail(f"Go tooling build failed:\n{error.stderr[-3000:]}")
    return binary


@pytest.mark.parametrize("operation", OPERATIONS)
def test_full_go_cli_native_tooling_parity(operation, tooling_go, tmp_path):
    tmp_path = tmp_path.resolve()
    request = request_for(operation)
    binary = Path(os.environ["SPL_CLI"])
    assert binary.is_absolute() and binary.is_file()
    go = json.loads(subprocess.check_output([str(tooling_go), operation], input=json.dumps(request).encode()))
    with SPLMapper(**mapper_kwargs()) as mapper:
        assert getattr(mapper, operation)(request) == go
        with pytest.raises(SPLMapperError):
            getattr(mapper, operation)({"directory": "/"})
    assert "error" in json.loads(subprocess.check_output([str(tooling_go), operation], input=b'{"directory":"/"}'))
    def write(name, value):
        path = tmp_path / name
        path.write_text(json.dumps(value))
        return str(path)
    if operation == "document_view":
        args = ["document", "--query", request["text"], "--source-id", request["source_id"]]
    else:
        manifest = write("manifest.json", {"schema_version": 1, "documents": [
            {"id": item["id"], **item["document"]} for item in DOCS]})
        if operation.startswith("impact_"):
            kind = operation.removeprefix("impact_")
            args = ["impact-" + kind, "--manifest", manifest, "--format", "json"]
            for side in ("before", "after"):
                field = side + ("_target" if kind == "schema" else "_rules")
                args += ["--" + field.replace("_", "-"), write(field + ".json", request[field])]
        else:
            args = ["graph" if operation == "export_graph" else "scan", "--manifest", manifest,
                    "--target", write("target.json", AFTER), "--format",
                    "sarif" if operation == "export_sarif" else "json"]
    snapshots = {p: p.read_bytes() for p in tmp_path.iterdir()}
    result = subprocess.run([str(binary), *args], capture_output=True, text=True, encoding="utf-8")
    assert not result.stderr
    assert result.returncode == (0 if operation == "document_view" else 3 if operation == "impact_mapping" else 1)
    local = json.loads(result.stdout)
    if "selection" in go:
        assert go["selection"]["mode"] == "inline" and local["selection"]["mode"] == "manifest"
        assert {k: v for k, v in go.items() if k != "selection"} == {k: v for k, v in local.items() if k != "selection"}
    else:
        assert local == go
    assert snapshots == {p: p.read_bytes() for p in tmp_path.iterdir()}
