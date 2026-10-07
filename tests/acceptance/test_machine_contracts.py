"""Offline instance contracts, independent of SPL field projection.

Run this file directly with pytest; it needs Go but no native/Python binding.
SPL_CONTRACT_GO may select an absolute Go executable; Go caches follow Go env.
"""
from __future__ import annotations

import copy
import hashlib
import json
import os
from pathlib import Path
import re
import subprocess
from urllib.parse import urljoin

import jsonschema
import pytest
from referencing import Registry, Resource
from referencing.exceptions import NoSuchResource, Unresolvable

ROOT = Path(__file__).resolve().parents[2]
CONTRACTS = ROOT / "contracts"
FAMILIES = ("query-document", "capabilities", "analysis", "requirements", "field-validation",
            "schema-validation", "rewrite", "corpus", "manifest", "graph",
            "impact", "document-view", "lsp-configuration", "closure-request", "closure",
            "environment-snapshot", "field-schema-bundle", "environment-validation",
            "compatibility-request", "compatibility", "resolution-request", "resolution")


def deny_unknown(uri):
    raise NoSuchResource(ref=uri)


def references(value, base):
    if isinstance(value, dict):
        identity = value.get("$id", value.get("id", ""))
        if isinstance(identity, str):
            base = urljoin(base, identity)
        for key, child in value.items():
            if key in ("$ref", "$dynamicRef"):
                yield urljoin(base, child)
            else:
                yield from references(child, base)
    elif isinstance(value, list):
        for child in value:
            yield from references(child, base)


@pytest.fixture(scope="module")
def schemas():
    expected = [CONTRACTS / "v1" / f"{name}.schema.json" for name in FAMILIES]
    assert all(p.is_file() for p in expected), "published contract family missing"
    paths = sorted(CONTRACTS.rglob("*.schema.json"))
    paths.append(CONTRACTS / "sarif" / "sarif-schema-2.1.0.json")
    documents = [json.loads(p.read_text()) for p in paths]
    identifiers = [d.get("$id", d.get("id")) for d in documents]
    assert all(identifiers)
    assert len(set(identifiers)) == len(documents)
    registry = Registry(retrieve=deny_unknown).with_resources(
        (identity, Resource.from_contents(d)) for identity, d in zip(identifiers, documents))
    result = {}
    for path, schema, identity in zip(paths, documents, identifiers):
        expected_dialect = ("http://json-schema.org/draft-04/schema#" if path.parent.name == "sarif"
                            else "https://json-schema.org/draft/2020-12/schema")
        assert schema["$schema"] == expected_dialect
        validator = jsonschema.validators.validator_for(schema)
        validator.check_schema(schema)
        for ref in references(schema, identity):
            registry.resolver().lookup(ref)  # Eager closure, including unused branches.
        result[path.relative_to(CONTRACTS).as_posix()] = validator(schema, registry=registry,
                                     format_checker=jsonschema.FormatChecker())
    return result


def errors(schemas, family, instance, definition=None, *, version="v1"):
    name = f"{version}/{family}.schema.json" if family != "sarif" else "sarif/sarif-schema-2.1.0.json"
    validator = schemas[name]
    if definition:
        validator = validator.evolve(
            schema={"$ref": validator.schema["$id"] + "#/$defs/" + definition})
    return list(validator.iter_errors(instance))  # Never short-circuit the iterator.


CAPABILITY_DIMENSIONS = ("syntax", "semantics", "requirements", "linting", "safe_rewriting")
FORBIDDEN_CAPABILITY_SCORE_KEYS = {"percent", "percentage", "score", "composite_score"}


def capability_ledger_consistency_errors(manifest):
    failures = []
    evidence_ids = {item["id"] for item in manifest["evidence"]}
    for record in manifest["records"]:
        for dimension in CAPABILITY_DIMENSIONS:
            for evidence_id in record["dimensions"][dimension]["evidence_ids"]:
                if evidence_id not in evidence_ids:
                    failures.append(f"{record['id']} {dimension} references missing evidence {evidence_id}")

    expected = {}
    for dimension in CAPABILITY_DIMENSIONS:
        counts = {
            "applicable": 0, "covered": 0, "supported": 0, "partial": 0,
            "unsupported": 0, "not_applicable": 0, "unassessed": 0,
        }
        for record in manifest["records"]:
            state = record["dimensions"][dimension]["state"]
            counts[state] += 1
            if state != "not_applicable":
                counts["applicable"] += 1
            if state == "supported":
                counts["covered"] += 1
        expected[dimension] = counts
    if manifest["summary"] != expected:
        failures.append("summary does not match records")
    return failures


def forbidden_capability_score_paths(value, path=()):
    found = []
    if isinstance(value, dict):
        for key, child in value.items():
            child_path = path + (key,)
            if key in FORBIDDEN_CAPABILITY_SCORE_KEYS:
                found.append(child_path)
            found.extend(forbidden_capability_score_paths(child, child_path))
    elif isinstance(value, list):
        for index, child in enumerate(value):
            found.extend(forbidden_capability_score_paths(child, path + (index,)))
    return found


def shared_errors(schemas, definition, instance, *, version="v1"):
    validator = schemas[f"{version}/shared.schema.json"]
    validator = validator.evolve(
        schema={"$ref": validator.schema["$id"] + "#/$defs/" + definition})
    return list(validator.iter_errors(instance))


def test_published_contracts_and_offline_closure(schemas):
    assert len(schemas) >= len(FAMILIES) + 1
    with pytest.raises(NoSuchResource):
        Registry(retrieve=deny_unknown).get_or_retrieve("https://unregistered.invalid/schema")


def test_authored_contract_cases(schemas):
    fixture = json.loads((ROOT / "testdata/tooling/contracts.json").read_text())
    for case in fixture["cases"]:
        actual = errors(schemas, case["family"], case["instance"], case.get("definition"))
        assert (not actual) == case["valid"], (case["id"], [e.message for e in actual])


def test_required_and_optional_output_members(schemas, emitted):
    for family in ("analysis", "capabilities", "field-validation", "schema-validation",
                   "rewrite", "corpus", "graph", "impact", "document-view"):
        instance = emitted[family]
        assert not errors(schemas, family, instance), family
        additive = copy.deepcopy(instance)
        additive["future_annotation"] = {"arbitrary": [None, True, 3]}
        assert not errors(schemas, family, additive), family
        missing = copy.deepcopy(instance)
        del missing["schema_version"]
        assert errors(schemas, family, missing), family
        wrong = copy.deepcopy(instance)
        wrong["schema_version"] = 2
        assert errors(schemas, family, wrong), family
    assert "candidate_validation" not in emitted["rewrite"]
    wrong = copy.deepcopy(emitted["rewrite"])
    wrong["candidate_validation"] = None
    assert errors(schemas, "rewrite", wrong)
    for family, member in (("analysis", "diagnostics"), ("corpus", "entries"),
                           ("graph", "edges"), ("impact", "entries")):
        wrong = copy.deepcopy(emitted[family])
        wrong[member] = None
        assert errors(schemas, family, wrong)


def test_official_sarif_schema_and_formats(schemas, emitted):
    path = CONTRACTS / "sarif/sarif-schema-2.1.0.json"
    assert hashlib.sha256(path.read_bytes()).hexdigest() == "c3b4bb2d6093897483348925aaa73af03b3e3f4bd4ca38cef26dcb4212a2682e"
    assert not errors(schemas, "sarif", emitted["sarif"])
    valid = copy.deepcopy(emitted["sarif"])
    valid["runs"][0]["invocations"] = [{"executionSuccessful": True,
                                         "startTimeUtc": "2026-09-12T00:00:00Z"}]
    assert not errors(schemas, "sarif", valid)
    for property_name, value in (("startTimeUtc", "not-a-date"), ("endTimeUtc", "2026-99-01T00:00:00Z")):
        log = copy.deepcopy(emitted["sarif"])
        log["runs"][0]["invocations"] = [{"executionSuccessful": True, property_name: value}]
        assert errors(schemas, "sarif", log)
    log = copy.deepcopy(emitted["sarif"])
    log["$schema"] = "not a URI"
    assert errors(schemas, "sarif", log)
    log = copy.deepcopy(emitted["sarif"])
    log["version"] = "2.0.0"
    assert errors(schemas, "sarif", log)


GO_EMITTER = r'''
package main
import (
 "encoding/json"
 "fmt"
 "os"
 "strings"
 "github.com/delgado-jacob/spl-toolkit/pkg/analysis"
 "github.com/delgado-jacob/spl-toolkit/pkg/validation"
 "github.com/delgado-jacob/spl-toolkit/pkg/compatibility"
 "github.com/delgado-jacob/spl-toolkit/pkg/environment"
 "github.com/delgado-jacob/spl-toolkit/pkg/rewrite"
 "github.com/delgado-jacob/spl-toolkit/pkg/resolution"
 "github.com/delgado-jacob/spl-toolkit/pkg/corpus"
 "github.com/delgado-jacob/spl-toolkit/pkg/corpusio"
 "github.com/delgado-jacob/spl-toolkit/pkg/document"
 "github.com/delgado-jacob/spl-toolkit/pkg/graph"
 "github.com/delgado-jacob/spl-toolkit/pkg/impact"
 "github.com/delgado-jacob/spl-toolkit/pkg/sarif"
)
func boolptr(v bool)*bool{return &v}
func must[T any](v T, e error) T { if e != nil { panic(e) }; return v }
func compatibilityExamples() map[string]any {
 path:=os.Getenv("SPL_COMPATIBILITY_FIXTURES"); if path=="" {path="testdata/compatibility/cases.json"}
 raw:=must(os.ReadFile(path));var cases []struct{Name string `json:"name"`;Document analysis.QueryDocument `json:"document"`;Snapshot environment.Snapshot `json:"snapshot"`;SchemaBundle *environment.SchemaBundle `json:"schema_bundle"`;QueryScope environment.CaptureScope `json:"query_scope"`;Bindings map[string]compatibility.InputBinding `json:"bindings"`}
 must(0,json.Unmarshal(raw,&cases));out:=map[string]any{}
 selected:=map[string]string{"explicit-dataset":"satisfied","closed-schema-missing-field":"unsatisfied","schema-not-supplied":"incomplete","malformed-query":"not assessed","pipeline-three-input-renamed-key":"chain","select-first-left-joins":"sql-left","placeholder-binding-absent-object":"absent"}
 for _,c:=range cases {label,ok:=selected[c.Name];if !ok {continue};req:=compatibility.Request{SchemaVersion:1,Requirements:*must(analysis.Requirements(c.Document)),Snapshot:c.Snapshot,SchemaBundle:c.SchemaBundle,QueryScope:c.QueryScope,InputBindings:[]compatibility.InputBinding{}}
  for _,input:=range req.Requirements.Inputs {binding,ok:=c.Bindings[input.Kind+":"+input.Name];if !ok {continue};binding.InputID=input.ID;if binding.Expected.Kind=="" {for _,obj:=range c.Snapshot.Objects {if obj.ID==binding.ObjectID {binding.Expected=environment.ObjectIdentity{Kind:obj.Kind,Name:obj.Name,Namespace:obj.Namespace,App:obj.App,Owner:obj.Owner}}}};req.InputBindings=append(req.InputBindings,binding)}
  out[label+"-request"]=req;out[label]=must(compatibility.Check(req))
 }
 return out
}
func main() {
 if len(os.Args)>2 && os.Args[1]=="resolution-request" { _,err:=resolution.ResolveJSON([]byte(os.Args[2]));if err!=nil{fmt.Print("rejected")}else{fmt.Print("accepted")};return }
 if len(os.Args)>1 {
  raw := []byte(os.Args[2]); var e error
  switch os.Args[1] {
  case "compatibility-request": _,e=compatibility.CheckJSON(raw)
  case "corpus": _,e=corpus.DecodeRequest(raw)
  case "manifest": _,e=corpusio.DecodeManifest(raw)
  case "rewrite": _,e=rewrite.DecodeRequest(raw)
  case "impact": _,e=impact.DecodeSchemaRequest(raw)
  case "query-document": _,e=validation.DecodeDocuments(append(append([]byte{'['},raw...),']'))
  default: panic("unknown decoder family: "+os.Args[1])
  }
  if e==nil { fmt.Print("accepted") } else { fmt.Print("rejected") }; return
 }
 d:=analysis.QueryDocument{Text:"search host=web | table host"}
 a:=must(analysis.Analyze(d))
 f:=must(validation.Validate(d,validation.FieldCatalog{Fields:[]string{"host"}}))
 s:=must(validation.ValidateSchema(d,validation.SchemaTarget{Kind:"json_schema",Schema:json.RawMessage(`{"type":"object","properties":{"host":{"type":"string"}},"additionalProperties":false}`)}))
 r:=must(rewrite.Rewrite(rewrite.Request{SchemaVersion:1,Document:d,Rules:[]rewrite.Rule{}}))
 cv:=must(rewrite.Rewrite(rewrite.Request{SchemaVersion:1,Document:d,Rules:[]rewrite.Rule{},ValidationTarget:&corpus.ValidationTarget{Kind:"field_list",Catalog:&validation.FieldCatalog{Fields:[]string{"host"}}}}))
 c:=must(corpus.Scan(corpus.Request{SchemaVersion:1,Documents:[]corpus.RequestDocument{{ID:"q",Document:d}}}))
 imp:=must(impact.DecodeSchemaRequest([]byte(`{"schema_version":1,"documents":[{"id":"q","document":{"text":"search host=web | table host"}}],"before_target":{"kind":"field_list","catalog":{"fields":["host"]}},"after_target":{"kind":"field_list","catalog":{"fields":[]}}}`)))
 m:=map[string]any{"query-document":a.Document,"analysis":a,"field-validation":f,"schema-validation":s,"rewrite":r,"rewrite-validated":cv,"capabilities":analysis.Capabilities(),"corpus":c,"graph":must(graph.Export(c)),"impact":must(impact.CompareSchemas(imp)),"document-view":must(document.New(a,document.RevisionContext{ToolVersion:"0.1.1",ContractVersion:"1"})),"sarif":must(sarif.Export(c))}
 m["requirements"]=must(analysis.Requirements(d))
 limitedDocument:=analysis.QueryDocument{Text:strings.Repeat("x ",2048)+"x"}
 m["resource-analysis"]=must(analysis.Analyze(limitedDocument))
 m["resource-requirements"]=must(analysis.Requirements(limitedDocument))
 m["spl2-analysis"]=must(analysis.Analyze(analysis.QueryDocument{Text:"from main | where host=\"web\" | select host",Language:"spl2"}))
 dotted:=analysis.QueryDocument{Text:"FROM main | fields actor.name, 'actor.name' | fields - 'actor.name'",Language:"spl2"}
 m["dotted-analysis"]=must(analysis.Analyze(dotted))
 m["dotted-requirements"]=must(analysis.Requirements(dotted))
 m["spl2-capabilities"]=must(analysis.CapabilitiesFor(analysis.CapabilityOptions{Language:"spl2"}))
 m["unsupported"]=must(analysis.Analyze(analysis.QueryDocument{Text:"search host=web | mystery host"}))
 m["unknown-corpus"]=must(corpus.Scan(corpus.Request{SchemaVersion:1,Documents:[]corpus.RequestDocument{{ID:"unknown",Document:analysis.QueryDocument{Text:"search host=web | mystery host"}}}}))
 m["empty-analysis"]=must(analysis.Analyze(analysis.QueryDocument{Text:""}))
 m["field-batch"]=must(validation.ValidateBatch([]analysis.QueryDocument{d},validation.FieldCatalog{Fields:[]string{"host"}}))
 m["schema-batch"]=must(validation.ValidateSchemaBatch([]analysis.QueryDocument{d},validation.SchemaTarget{Kind:"json_schema",Schema:json.RawMessage(`true`)}))
 m["rewrite-batch"]=must(rewrite.RewriteBatch(rewrite.BatchRequest{SchemaVersion:1,Documents:[]analysis.QueryDocument{d},Rules:[]rewrite.Rule{}}))
 mapping:=must(impact.DecodeMappingRequest([]byte(`{"schema_version":1,"documents":[{"id":"q","document":{"text":"search host=web | table host"}}],"before_rules":{"schema_version":1,"rules":[]},"after_rules":{"schema_version":1,"rules":[{"id":"host-server","kind":"field","source":{"name":"host"},"target":{"name":"server"}}]}}`)))
 m["mapping-impact"]=must(impact.CompareMappings(mapping))
 fieldTarget:=&corpus.ValidationTarget{Kind:"field_list",Catalog:&validation.FieldCatalog{Fields:[]string{}}}
 fieldScan:=must(corpus.Scan(corpus.Request{SchemaVersion:1,Documents:[]corpus.RequestDocument{{ID:"q",Document:d}},ValidationTarget:fieldTarget}))
 m["field-corpus"]=fieldScan
 m["invalid-sarif"]=must(sarif.Export(fieldScan))
 schemaTarget:=&corpus.ValidationTarget{Kind:"json_schema",SchemaTarget:&validation.SchemaTarget{Kind:"json_schema",Schema:json.RawMessage(`true`)}}
 m["schema-corpus"]=must(corpus.Scan(corpus.Request{SchemaVersion:1,Documents:[]corpus.RequestDocument{{ID:"q",Document:d}},ValidationTarget:schemaTarget}))
 selection:=corpus.Selection{Mode:"directory",Complete:true,IgnoredNames:[]string{},SkippedSymlinks:[]string{},TraversalFailures:[]corpus.AcquisitionError{}}
 prepared:=must(corpus.Prepare(corpus.ScanOptions{}))
 incompleteSelection:=selection
 incompleteSelection.Complete=false
 incompleteSelection.TraversalFailures=[]corpus.AcquisitionError{{Code:"traversal_failed",Phase:"traverse",Message:"unreadable directory"}}
 emptyInput:=corpus.Input{Selection:incompleteSelection,Entries:[]corpus.Entry{}}
 empty:=must(prepared.Scan(emptyInput))
 failedInput:=corpus.Input{Selection:selection,Entries:[]corpus.Entry{{ID:"missing",Origin:corpus.Origin{Kind:"file",RelativePath:"missing.spl",BaseURI:"file:///queries/"},Failure:&corpus.AcquisitionError{Code:"not_found",Phase:"open",Message:"missing"}}}}
 failed:=must(prepared.Scan(failedInput))
 m["empty-corpus"],m["empty-graph"],m["failed-corpus"]=empty,must(graph.Export(empty)),failed
 m["failed-graph"],m["failed-sarif"]=must(graph.Export(failed)),must(sarif.Export(failed))
 comparison:=must(impact.PrepareSchemas(*fieldTarget,*fieldTarget))
 m["failed-impact"]=must(comparison.Compare(failedInput))
 m["empty-impact"]=must(comparison.Compare(emptyInput))
 scope:=environment.CaptureScope{Namespace:environment.Selector{All:boolptr(true)},App:environment.Selector{All:boolptr(true)},Owner:environment.Selector{All:boolptr(true)}}
 snapshot:=environment.Snapshot{SchemaVersion:1,ScopeID:"offline",CaptureScope:scope,Origin:environment.Origin{InstanceID:"local",ProductVersion:"9.4",Producer:"fixture",ProducerVersion:"1"},Capture:environment.CaptureInterval{Start:"2026-10-01T12:00:00Z",End:"2026-10-01T12:05:00Z"},Capabilities:[]environment.Capability{},Collections:[]environment.Collection{},Objects:[]environment.Object{}}
 request:=compatibility.Request{SchemaVersion:1,Requirements:*must(analysis.Requirements(analysis.QueryDocument{Text:"from [{id:1}]",Language:"spl2"})),QueryScope:scope,InputBindings:[]compatibility.InputBinding{},Snapshot:snapshot}
 rawResolution:=must(os.ReadFile("examples/resolution/request.json"))
 resolutionRequest:=must(resolution.DecodeRequest(rawResolution))
 m["resolution-request"]=resolutionRequest
 m["resolution"]=must(resolution.Resolve(resolutionRequest))
 resolutionRequest.Document=analysis.QueryDocument{Text:"from [{id:1}]",Language:"spl2"}
 resolutionRequest.Resolutions=[]resolution.Resolution{}
 resolutionRequest.Compatibility.InputBindings=[]compatibility.ResolutionBinding{}
 m["resolution-noop"]=must(resolution.Resolve(resolutionRequest))
 m["compatibility-examples"]=compatibilityExamples()
 m["compatibility-request"]=request
 m["compatibility"]=must(compatibility.Check(request))
 must(0,json.NewEncoder(os.Stdout).Encode(m))
}
'''


@pytest.fixture(scope="module")
def emitter(tmp_path_factory):
    directory = tmp_path_factory.mktemp("machine-contract-go")
    source = directory / "main.go"
    source.write_text(GO_EMITTER)
    binary = directory / "emit"
    subprocess.run([os.environ.get("SPL_CONTRACT_GO", "go"), "build", "-mod=readonly",
                    "-o", str(binary), str(source)], cwd=ROOT, check=True,
                   capture_output=True, text=True)
    return binary


@pytest.fixture(scope="module")
def emitted(emitter):
    return json.loads(subprocess.check_output([str(emitter)], cwd=ROOT, text=True))


def test_emitted_semantics_and_legacy_wire(schemas, emitted):
    a = emitted["analysis"]
    assert a["document"] == {"text": "search host=web | table host", "language": "spl",
                             "profile": "splunkd", "version": "current", "source_id": ""}
    assert a["status"] == "valid"
    assert [stage["command"] for stage in a["stages"]] == ["search", "table"]
    assert a["diagnostics"] == []
    assert emitted["rewrite"]["text"] == a["document"]["text"]
    assert emitted["rewrite"]["committed"] is False
    assert emitted["rewrite"]["changes"] == []
    assert emitted["corpus"]["counts"]["selected"] == 1
    assert emitted["impact"]["counts"]["affected"] == 1
    assert emitted["unsupported"]["coverage"]["semantic_complete"] is False
    for key in ("spl2-analysis", "unsupported", "empty-analysis"):
        assert not errors(schemas, "analysis", emitted[key]), key
    assert not errors(schemas, "capabilities", emitted["spl2-capabilities"])
    assert not errors(schemas, "rewrite", emitted["rewrite-validated"])
    assert emitted["rewrite-validated"]["candidate_validation"]["field_list"]["target"] == {
        "kind": "field_list", "identity": "", "version": ""}


def test_requirement_contract_and_additive_v1_compatibility(schemas, emitted):
    requirement_set = emitted["requirements"]
    assert not errors(schemas, "requirements", requirement_set)
    assert not errors(schemas, "analysis", emitted["analysis"])
    assert not errors(schemas, "document-view", emitted["document-view"])

    additive = copy.deepcopy(requirement_set)
    additive["future_annotation"] = {"arbitrary": [None, True, 3]}
    assert not errors(schemas, "requirements", additive)

    archived_analysis = copy.deepcopy(emitted["analysis"])
    del archived_analysis["requirements"]
    assert not errors(schemas, "analysis", archived_analysis)
    archived_snapshot = copy.deepcopy(emitted["document-view"])
    del archived_snapshot["requirements"]
    assert not errors(schemas, "document-view", archived_snapshot)


def test_structural_and_atomic_dotted_identity_machine_contracts(schemas, emitted):
    analysis_report = emitted["dotted-analysis"]
    requirements = emitted["dotted-requirements"]
    assert not errors(schemas, "analysis", analysis_report)
    assert not errors(schemas, "requirements", requirements)
    assert analysis_report["requirements"] == requirements
    path = {"kind": "path", "segments": ["actor", "name"]}
    atomic = {"kind": "atomic", "segments": ["actor.name"]}
    assert [reference["field_identity"] for reference in analysis_report["references"][1:]] == [path, atomic, atomic]
    assert [item["field_identity"] for item in requirements["items"][1:]] == [path, atomic]
    assert [field["field_identity"] for field in analysis_report["lineage"][1]["after"]["fields"]] == [atomic, path]
    assert [field["field_identity"] for field in analysis_report["lineage"][2]["after"]["fields"]] == [path]
    assert analysis_report["lineage"][2]["after"]["removed"] == [{"name": "actor.name", "field_identity": atomic}]
    assert analysis_report["lineage"][2]["transitions"][0]["output_identity"] == atomic
    invalid = copy.deepcopy(analysis_report)
    del invalid["lineage"][2]["after"]["removed"][0]["field_identity"]
    assert errors(schemas, "analysis", invalid)


def test_resource_limit_contracts(schemas, emitted):
    diagnostic_message = (
        "analysis stopped before parser prediction after reaching the "
        "4,096-unit lexer work limit"
    )
    gap_message = (
        "requirement coverage is incomplete because analysis exceeded the "
        "4,096-unit lexer work limit"
    )
    analysis_report = emitted["resource-analysis"]
    requirement_set = emitted["resource-requirements"]

    assert not errors(schemas, "analysis", analysis_report)
    assert not errors(schemas, "requirements", requirement_set)
    assert analysis_report["status"] == "incomplete"
    assert analysis_report["coverage"] == {
        "syntax_complete": False,
        "semantic_complete": False,
        "reasons": ["SPL_ANALYSIS_RESOURCE_LIMIT"],
    }
    for member in ("stages", "scopes", "references", "lineage"):
        assert analysis_report[member] == []
    assert all(value == [] for value in analysis_report["dependencies"].values())
    assert len(analysis_report["diagnostics"]) == 1
    diagnostic = analysis_report["diagnostics"][0]
    assert (diagnostic["code"], diagnostic["category"], diagnostic["severity"],
            diagnostic["message"]) == (
        "SPL_ANALYSIS_RESOURCE_LIMIT", "resource_limit", "warning", diagnostic_message)

    assert requirement_set["query_status"] == "incomplete"
    assert requirement_set["coverage"] == {
        "complete": False,
        "reasons": ["SPL_ANALYSIS_RESOURCE_LIMIT"],
    }
    assert requirement_set["items"] == []
    assert requirement_set["gaps"] == [{
        "code": "SPL_ANALYSIS_RESOURCE_LIMIT",
        "message": gap_message,
        "reference_ids": [],
        "diagnostic_codes": ["SPL_ANALYSIS_RESOURCE_LIMIT"],
    }]
    assert requirement_set["diagnostics"] == [diagnostic]
    assert re.fullmatch(r"sha256:[0-9a-f]{64}", requirement_set["query"]["query_digest"])
    assert re.fullmatch(r"sha256:[0-9a-f]{64}", requirement_set["capability_revision"])


def test_requirement_required_types_enums_ids_and_digests(schemas, emitted):
    complete = emitted["requirements"]
    limited = emitted["resource-requirements"]

    for member in ("schema_version", "query", "capability_revision", "query_status",
                   "coverage", "items", "gaps", "diagnostics"):
        invalid = copy.deepcopy(complete)
        del invalid[member]
        assert errors(schemas, "requirements", invalid), member
    for member in ("source_id", "language", "profile", "version", "query_digest"):
        invalid = copy.deepcopy(complete)
        del invalid["query"][member]
        assert errors(schemas, "requirements", invalid), member
    for member in ("complete", "reasons"):
        invalid = copy.deepcopy(complete)
        del invalid["coverage"][member]
        assert errors(schemas, "requirements", invalid), member

    item = complete["items"][0]
    for member in ("id", "kind", "identity", "role", "necessity", "origin",
                   "resolution", "occurrences"):
        invalid = copy.deepcopy(complete)
        del invalid["items"][0][member]
        assert errors(schemas, "requirements", invalid), member
    for member in ("reference_id", "original_name", "binding", "stage_id", "scope_id",
                   "location"):
        invalid = copy.deepcopy(complete)
        del invalid["items"][0]["occurrences"][0][member]
        assert errors(schemas, "requirements", invalid), member
    for member in ("code", "message", "reference_ids", "diagnostic_codes"):
        invalid = copy.deepcopy(limited)
        del invalid["gaps"][0][member]
        assert errors(schemas, "requirements", invalid), member

    for member, value in (
        ("schema_version", "1"), ("query", []), ("capability_revision", 1),
        ("query_status", 1), ("coverage", []), ("items", {}), ("gaps", {}),
        ("diagnostics", {}),
    ):
        invalid = copy.deepcopy(complete)
        invalid[member] = value
        assert errors(schemas, "requirements", invalid), member
    nested_types = (
        (("query", "source_id"), 1), (("coverage", "complete"), "true"),
        (("coverage", "reasons"), {}), (("items", 0, "identity"), 1),
        (("items", 0, "occurrences"), {}),
        (("items", 0, "occurrences", 0, "location"), []),
        (("gaps", 0, "reference_ids"), {}),
        (("gaps", 0, "diagnostic_codes"), {}),
    )
    for path, value in nested_types:
        invalid = copy.deepcopy(complete if path[0] != "gaps" else limited)
        target = invalid
        for component in path[:-1]:
            target = target[component]
        target[path[-1]] = value
        assert errors(schemas, "requirements", invalid), path

    for member, value in (
        ("query_status", "unknown"),
        ("items.0.kind", "command"),
        ("items.0.necessity", "optional"),
        ("items.0.origin", "inferred"),
        ("items.0.resolution", "unresolved"),
    ):
        invalid = copy.deepcopy(complete)
        target = invalid
        parts = member.split(".")
        for component in parts[:-1]:
            target = target[int(component)] if component.isdigit() else target[component]
        target[parts[-1]] = value
        assert errors(schemas, "requirements", invalid), member
    for member, value in (("language", "sql"), ("profile", "cloud"), ("version", "latest")):
        invalid = copy.deepcopy(complete)
        invalid["query"][member] = value
        assert errors(schemas, "requirements", invalid), member

    for path, value in (
        (("items", 0, "id"), "req-0"),
        (("items", 0, "occurrences", 0, "reference_id"), "ref-x"),
        (("items", 0, "occurrences", 0, "reference_id"), "ref-00"),
    ):
        invalid = copy.deepcopy(complete)
        target = invalid
        for component in path[:-1]:
            target = target[component]
        target[path[-1]] = value
        assert errors(schemas, "requirements", invalid), path
    invalid = copy.deepcopy(limited)
    invalid["gaps"][0]["reference_ids"] = ["reference-1"]
    assert errors(schemas, "requirements", invalid)

    for member in ("query.query_digest", "capability_revision"):
        for digest in ("sha256:abc", "sha256:" + "A" * 64, "0" * 64):
            invalid = copy.deepcopy(complete)
            if member.startswith("query."):
                invalid["query"][member.split(".")[1]] = digest
            else:
                invalid[member] = digest
            assert errors(schemas, "requirements", invalid), (member, digest)

    assert item["id"].startswith("req-")


def test_duplicate_keys_rejected_by_actual_decoders(emitter):
    cases = json.loads((ROOT / "testdata/tooling/contracts.json").read_text())["decoder_cases"]
    for case in cases:
        result = subprocess.check_output([str(emitter), case["family"], case["raw"]], text=True)
        assert result == case["expected"], case["id"]


def test_invalid_discriminated_output_members(schemas, emitted):
    impact = copy.deepcopy(emitted["impact"])
    impact["entries"][0]["failure"] = {"code": "not_found", "phase": "open", "message": "missing"}
    assert errors(schemas, "impact", impact)
    corpus = copy.deepcopy(emitted["corpus"])
    corpus["entries"][0]["evaluation"]["field_validation"] = emitted["field-validation"]
    assert errors(schemas, "corpus", corpus)
    graph = copy.deepcopy(emitted["graph"])
    stage = next(n for n in graph["nodes"] if n["kind"] == "stage")
    del stage["position"]
    assert errors(schemas, "graph", graph)


def test_live_report_variants_and_empty_arrays(schemas, emitted):
    for key, family, definition in (
        ("field-batch", "field-validation", "BatchReport"),
        ("schema-batch", "schema-validation", "BatchReport"),
        ("rewrite-batch", "rewrite", "BatchReport"),
        ("mapping-impact", "impact", None), ("failed-impact", "impact", None),
        ("empty-impact", "impact", None), ("field-corpus", "corpus", None),
        ("schema-corpus", "corpus", None), ("empty-corpus", "corpus", None),
        ("failed-corpus", "corpus", None), ("empty-graph", "graph", None),
        ("failed-graph", "graph", None), ("invalid-sarif", "sarif", None),
        ("failed-sarif", "sarif", None),
        ("unknown-corpus", "corpus", None),
    ):
        failures = errors(schemas, family, emitted[key], definition)
        assert not failures, (key, [e.message for e in failures])
    for key in ("empty-corpus", "empty-impact"):
        assert emitted[key]["entries"] == []
    assert emitted["empty-graph"]["nodes"] == []
    assert emitted["empty-graph"]["edges"] == []
    assert emitted["failed-corpus"]["execution_complete"] is False
    assert emitted["failed-impact"]["entries"][0]["classification"] == "failed"
    assert emitted["mapping-impact"]["entries"][0]["after_rewrite"]["candidate_text"] == "search server=web | table server"
    assert emitted["mapping-impact"]["counts"]["affected"] == 1
    assert emitted["invalid-sarif"]["runs"][0]["results"]


def test_capabilities_match_accepted_dialect_forms(emitted):
    cases = json.loads((ROOT / "testdata/rewrite/forms.json").read_text())["capability_checks"]
    expected = {(c["language"], c["kind"], c["role"]): c["supported"] for c in cases}
    actual = {}
    for key in ("capabilities", "spl2-capabilities"):
        manifest = emitted[key]
        assert manifest["language"] in ("spl", "spl2")
        for form in manifest["rewrite"]["forms"]:
            actual[(manifest["language"], form["kind"], form["role"])] = form["supported"]
            assert form["identity_forms"] == (["path"] if form["role"] == "navigation" else ["atom"])
    assert actual == expected
    assert "documentation_snapshot" not in emitted["capabilities"]
    assert emitted["spl2-capabilities"]["documentation_snapshot"] == (
        "spl2-provenance-v1:sha256:3345cf5712b1bdbf467d1651784fdb8bccc596805038da0d54e7a123384e3a4e")
    unknown = next(c for c in emitted["unknown-corpus"]["command_coverage"] if c["command"] == "mystery")
    assert (unknown["declared"], unknown["syntax_supported"], unknown["semantic_supported"]) == (False, False, False)
    assert emitted["unknown-corpus"]["status"] == "incomplete"


def test_capability_ledger_contract_and_additive_v1_compatibility(schemas, emitted):
    for key in ("capabilities", "spl2-capabilities"):
        manifest = emitted[key]
        assert not errors(schemas, "capabilities", manifest)
        assert not capability_ledger_consistency_errors(manifest)
        assert set(manifest["summary"]) == set(CAPABILITY_DIMENSIONS)
        assert manifest["toolkit_version"]
        assert not forbidden_capability_score_paths(manifest)

        archived = {
            member: copy.deepcopy(manifest[member])
            for member in ("schema_version", "language", "profile", "version", "commands", "functions")
        }
        if "rewrite" in manifest:
            archived["rewrite"] = copy.deepcopy(manifest["rewrite"])
        if "documentation_snapshot" in manifest:
            archived["documentation_snapshot"] = manifest["documentation_snapshot"]
        assert not errors(schemas, "capabilities", archived)

        for member, value in (
            ("toolkit_version", 1), ("records", {}), ("summary", []), ("evidence", {}),
        ):
            invalid = copy.deepcopy(manifest)
            invalid[member] = value
            assert errors(schemas, "capabilities", invalid), (key, member)

    manifest = copy.deepcopy(emitted["capabilities"])
    for path, value in (
        (("records", 0, "language"), "sql"),
        (("records", 0, "profile"), "cloud"),
        (("records", 0, "kind"), "operator"),
        (("records", 0, "dimensions", "syntax", "state"), "complete"),
        (("records", 0, "provenance", "source_family"), "unknown"),
        (("evidence", 0, "classification"), "unknown"),
    ):
        invalid = copy.deepcopy(manifest)
        target = invalid
        for component in path[:-1]:
            target = target[component]
        target[path[-1]] = value
        assert errors(schemas, "capabilities", invalid), path

    for path in (
        ("records", 0, "id"),
        ("records", 0, "grammar_registered"),
        ("records", 0, "dimensions", "syntax", "state"),
        ("records", 0, "provenance", "source_family"),
        ("summary", "syntax", "applicable"),
        ("evidence", 0, "observations"),
    ):
        invalid = copy.deepcopy(manifest)
        target = invalid
        for component in path[:-1]:
            target = target[component]
        del target[path[-1]]
        assert errors(schemas, "capabilities", invalid), path

    for path in (
        ("records", 0),
        ("records", 0, "dimensions", "syntax"),
        ("records", 0, "provenance"),
        ("summary", "syntax"),
        ("evidence", 0),
        ("evidence", 0, "observations"),
    ):
        invalid = copy.deepcopy(manifest)
        target = invalid
        for component in path:
            target = target[component]
        target["future_field"] = True
        assert errors(schemas, "capabilities", invalid), path

    for forbidden in ("percent", "score"):
        invalid = copy.deepcopy(manifest)
        invalid["summary"]["syntax"][forbidden] = 100
        assert errors(schemas, "capabilities", invalid), forbidden

    dangling = copy.deepcopy(manifest)
    dangling["records"][0]["dimensions"]["syntax"]["evidence_ids"] = ["missing-evidence"]
    assert capability_ledger_consistency_errors(dangling)

    mismatched = copy.deepcopy(manifest)
    mismatched["summary"]["syntax"]["supported"] += 1
    assert capability_ledger_consistency_errors(mismatched)

    registration_only = copy.deepcopy(manifest)
    registration_only["records"][0]["grammar_registered"] = not registration_only["records"][0]["grammar_registered"]
    assert not errors(schemas, "capabilities", registration_only)
    assert not capability_ledger_consistency_errors(registration_only)

    for path in (("score",), ("composite_score",), ("summary", "syntax", "percentage")):
        scored = copy.deepcopy(manifest)
        target = scored
        for component in path[:-1]:
            target = target[component]
        target[path[-1]] = 100
        assert path in forbidden_capability_score_paths(scored)


def test_capability_evidence_observation_definitions_are_strict(schemas):
    location = {
        "start": {"offset": 0, "line": 1, "column": 1},
        "end": {"offset": 1, "line": 1, "column": 2},
    }
    diagnostic = {
        "code": "SPL_TEST", "category": "semantic", "severity": "warning",
        "location": location,
    }
    definitions = {
        "analysis.CapabilityDiagnosticExpectation": diagnostic,
        "analysis.CapabilitySyntaxObservation": {"complete": True, "diagnostics": [diagnostic]},
        "analysis.CapabilityStageExpectation": {"command": "search", "semantic_complete": True},
        "analysis.CapabilityReferenceExpectation": {
            "normalized_name": "host", "kind": "field", "role": "search_field",
            "resolution": "exact", "binding": "source", "location": location,
        },
        "analysis.CapabilityDependencyExpectation": {"kind": "index", "name": "main"},
        "analysis.CapabilityTransitionExpectation": {"operation": "create", "output": "host"},
        "analysis.CapabilitySemanticsObservation": {
            "status": "valid", "complete": True,
            "stages": [{"command": "search", "semantic_complete": True}],
            "references": [], "dependencies": [], "transitions": [], "diagnostics": [],
        },
        "analysis.CapabilityRequirementExpectation": {
            "kind": "field", "identity": "host", "role": "search_field",
            "necessity": "required", "resolution": "exact",
        },
        "analysis.CapabilityRequirementsObservation": {
            "query_status": "valid", "complete": True,
            "items": [{"kind": "field", "identity": "host", "role": "search_field",
                       "necessity": "required", "resolution": "exact"}],
            "gap_codes": [],
        },
        "analysis.CapabilityLintingObservation": {"diagnostics": [diagnostic]},
        "analysis.CapabilityRewriteObservation": {
            "status": "valid", "committed": True, "rewrite_complete": True,
            "text": "search host=web", "candidate_text": "search host=web",
            "coverage_reasons": [], "change_reasons": [], "rule_evaluation_reasons": [],
        },
    }
    for definition, instance in definitions.items():
        assert not shared_errors(schemas, definition, instance), definition
        extra = copy.deepcopy(instance)
        extra["future_field"] = True
        assert shared_errors(schemas, definition, extra), definition
        for member in instance:
            missing = copy.deepcopy(instance)
            del missing[member]
            assert shared_errors(schemas, definition, missing), (definition, member)

    observations = {"syntax": definitions["analysis.CapabilitySyntaxObservation"]}
    assert not shared_errors(schemas, "analysis.CapabilityEvidenceObservations", observations)
    assert shared_errors(schemas, "analysis.CapabilityEvidenceObservations", {})
    observations["future_field"] = True
    assert shared_errors(schemas, "analysis.CapabilityEvidenceObservations", observations)


def test_capability_claim_state_boundaries_and_semantics_nonempty(schemas):
    valid_claims = (
        {"state": "supported", "evidence_ids": ["positive"], "limitations": []},
        {"state": "partial", "evidence_ids": ["positive", "incomplete"], "limitations": ["gap"]},
        {"state": "unsupported", "evidence_ids": ["negative"], "limitations": ["gap"]},
        {"state": "not_applicable", "evidence_ids": [], "limitations": ["not applicable"]},
        {"state": "unassessed", "evidence_ids": [], "limitations": []},
    )
    for claim in valid_claims:
        assert not shared_errors(schemas, "analysis.CapabilityClaim", claim), claim

    invalid_claims = (
        {"state": "supported", "evidence_ids": [], "limitations": []},
        {"state": "partial", "evidence_ids": ["positive"], "limitations": ["gap"]},
        {"state": "partial", "evidence_ids": ["positive", "incomplete"], "limitations": []},
        {"state": "unsupported", "evidence_ids": [], "limitations": ["gap"]},
        {"state": "unsupported", "evidence_ids": ["negative"], "limitations": []},
        {"state": "not_applicable", "evidence_ids": ["positive"], "limitations": ["not applicable"]},
        {"state": "not_applicable", "evidence_ids": [], "limitations": []},
        {"state": "unassessed", "evidence_ids": ["positive"], "limitations": []},
        {"state": "unassessed", "evidence_ids": [], "limitations": ["unknown"]},
    )
    for claim in invalid_claims:
        assert shared_errors(schemas, "analysis.CapabilityClaim", claim), claim

    empty_semantics = {
        "status": "valid", "complete": True, "stages": [], "references": [],
        "dependencies": [], "transitions": [], "diagnostics": [],
    }
    assert shared_errors(schemas, "analysis.CapabilitySemanticsObservation", empty_semantics)


def test_capability_evidence_rewrite_request_is_the_runtime_wrapper(schemas):
    rule = {
        "id": "rename-host", "kind": "field",
        "source": {"name": "host"}, "target": {"name": "server"},
    }
    request = {"schema_version": 1, "mode": "preview", "rules": [rule]}
    assert not shared_errors(schemas, "analysis.CapabilityEvidenceRewriteRequest", request)

    invalid_cases = {
        "wrong version": {**request, "schema_version": 2},
        "unknown key": {**request, "document": {"text": "search host=web"}},
        "missing rules": {"schema_version": 1, "mode": "preview"},
        "bad mode": {**request, "mode": "commit"},
        "bad rule": {**request, "rules": [{**rule, "kind": "command"}]},
    }
    for name, invalid in invalid_cases.items():
        assert shared_errors(schemas, "analysis.CapabilityEvidenceRewriteRequest", invalid), name


def test_capability_evidence_reuses_public_diagnostic_and_requirement_enums(schemas):
    location = {
        "start": {"offset": 0, "line": 1, "column": 1},
        "end": {"offset": 1, "line": 1, "column": 2},
    }
    diagnostic = {
        "code": "SPL_TEST", "category": "semantic", "severity": "warning",
        "location": location,
    }
    requirement = {
        "kind": "field", "identity": "host", "role": "search_field",
        "necessity": "required", "resolution": "exact",
    }
    assert not shared_errors(schemas, "analysis.CapabilityDiagnosticExpectation", diagnostic)
    assert not shared_errors(schemas, "analysis.CapabilityRequirementExpectation", requirement)

    for value in ("fatal", "notice"):
        invalid = copy.deepcopy(diagnostic)
        invalid["severity"] = value
        assert shared_errors(schemas, "analysis.CapabilityDiagnosticExpectation", invalid), value
    for member, value in (
        ("kind", "command"), ("necessity", "optional"), ("resolution", "unresolved"),
    ):
        invalid = copy.deepcopy(requirement)
        invalid[member] = value
        assert shared_errors(schemas, "analysis.CapabilityRequirementExpectation", invalid), member


def test_eager_reference_closure_rejects_unused_bad_resources(schemas):
    registry = Registry(retrieve=deny_unknown).with_resources(
        (v.schema.get("$id", v.schema.get("id")), Resource.from_contents(v.schema))
        for v in schemas.values())
    for ref in ("https://unregistered.invalid/unused",
                "https://github.com/delgado-jacob/spl-toolkit/contracts/v1/shared.schema.json#/$defs/Missing"):
        unused = {"$defs": {"unused": {"$ref": ref}}}
        with pytest.raises(Unresolvable):
            for address in references(unused, "https://example.invalid/root"):
                registry.resolver().lookup(address)


def test_environment_snapshot_bundle_request_and_report_contracts(schemas):
    cases = json.loads((ROOT / "testdata/environment/cases.json").read_text())
    for case in cases:
        snapshot = case["snapshot"]
        bundle = case["schema_bundle"]
        assert not errors(schemas, "environment-snapshot", snapshot, version=f"v{snapshot['schema_version']}"), case["name"]
        assert not errors(schemas, "field-schema-bundle", bundle), case["name"]
        request = {"schema_version": 1, "snapshot": snapshot, "schema_bundle": bundle}
        assert not errors(schemas, "environment-validation", request, "Request"), case["name"]

    snapshot = cases[0]["snapshot"]
    bundle = cases[0]["schema_bundle"]
    for family, source in (("environment-snapshot", snapshot), ("field-schema-bundle", bundle)):
        wrong = copy.deepcopy(source)
        wrong["schema_version"] = 2
        assert errors(schemas, family, wrong)
        wrong = copy.deepcopy(source)
        wrong["unexpected"] = True
        assert errors(schemas, family, wrong)
        wrong = copy.deepcopy(source)
        wrong["provenance" if family == "field-schema-bundle" else "origin"] = None
        assert errors(schemas, family, wrong)
    for changed in (
        {"schema_version": 2, "snapshot": snapshot},
        {"schema_version": 1},
        {"schema_version": 1, "snapshot": None},
        {"schema_version": 1, "snapshot": snapshot, "extra": True},
    ):
        assert errors(schemas, "environment-validation", changed, "Request")
    wrong = copy.deepcopy(snapshot)
    wrong["capture_scope"]["namespace"] = {"all": True, "values": ["default"]}
    assert errors(schemas, "environment-snapshot", wrong)
    wrong = copy.deepcopy(snapshot)
    wrong["collections"][0]["reason"] = None
    assert errors(schemas, "environment-snapshot", wrong)
    wrong = copy.deepcopy(bundle)
    wrong["schemas"][0]["target"] = {"kind": "json_schema", "schema": True}
    assert errors(schemas, "field-schema-bundle", wrong)
    wrong = copy.deepcopy(bundle)
    wrong["schemas"][0] = {"id": "crossed", "kind": "json_schema", "target": {"kind": "ocsf", "catalog": {}, "selection": {"version": "1", "class": "a"}}, "provenance": bundle["provenance"]}
    assert errors(schemas, "field-schema-bundle", wrong)
    utc_offset = copy.deepcopy(snapshot)
    utc_offset["capture"]["start"] = "2026-10-01T12:00:00+00:00"
    utc_offset["objects"][0]["provenance"]["observed_at"] = "2026-10-01T12:02:00+00:00"
    assert not errors(schemas, "environment-snapshot", utc_offset)
    empty_reason = copy.deepcopy(snapshot)
    empty_reason["collections"][0]["reason"] = ""
    assert not errors(schemas, "environment-snapshot", empty_reason)
    wrong = copy.deepcopy(bundle)
    wrong["schemas"][0]["catalog"]["extra"] = True
    assert errors(schemas, "field-schema-bundle", wrong)
    extensible = copy.deepcopy(bundle)
    extensible["schemas"][0] = {"id": "schema", "kind": "json_schema", "target": {"kind": "json_schema", "schema": {"type": "object", "x-custom": {"value": 1}}}, "provenance": bundle["provenance"]}
    assert not errors(schemas, "field-schema-bundle", extensible)
    extensible["schemas"][0] = {"id": "ocsf", "kind": "ocsf", "target": {"kind": "ocsf", "catalog": {"compile_version": 1, "classes": {"a": {"x-vendor": {"hint": None}}}}, "selection": {"version": "1.6.0", "class": "a"}}, "provenance": bundle["provenance"]}
    assert not errors(schemas, "field-schema-bundle", extensible)
    report = {"schema_version": 1, "status": "partial", "snapshot_digest": "sha256:" + "a" * 64, "coverage": [{"artifact": "snapshot", "kind": "index", "coverage": "unavailable", "reason": "omitted"}], "diagnostics": [{"code": "collection_unavailable", "severity": "warning", "artifact": "snapshot", "path": "/collections", "message": "index omitted"}]}
    assert not errors(schemas, "environment-validation", report)
    for field, value in (("schema_version", 2), ("status", "incomplete"), ("coverage", None), ("diagnostics", None)):
        wrong = copy.deepcopy(report)
        wrong[field] = value
        assert errors(schemas, "environment-validation", wrong)
    for field in ("schema_version", "status", "coverage", "diagnostics"):
        wrong = copy.deepcopy(report)
        del wrong[field]
        assert errors(schemas, "environment-validation", wrong)
    wrong = copy.deepcopy(report)
    wrong["snapshot_digest"] = "sha256:bad"
    assert errors(schemas, "environment-validation", wrong)


def test_environment_kind_and_ocsf_selection_boundaries(schemas):
    case = json.loads((ROOT / "testdata/environment/cases.json").read_text())[0]
    snapshot = case["snapshot"]
    bundle = case["schema_bundle"]

    index_with_query = copy.deepcopy(snapshot)
    index_with_query["objects"][0]["kind"] = "index"
    index_with_query["objects"][0]["document"] = {"text": "index=main"}
    assert errors(schemas, "environment-snapshot", index_with_query)

    macro_without_arity = copy.deepcopy(snapshot)
    macro_without_arity["objects"][0]["kind"] = "macro"
    macro_without_arity["objects"][0]["namespace"] = "search"
    assert errors(schemas, "environment-snapshot", macro_without_arity)
    macro_without_arity["objects"][0]["arity"] = 0
    assert not errors(schemas, "environment-snapshot", macro_without_arity)
    saved_search = copy.deepcopy(snapshot)
    saved_search["objects"][0]["kind"] = "saved_search"
    saved_search["objects"][0]["document"] = {"text": "index=main | stats count"}
    assert not errors(schemas, "environment-snapshot", saved_search)
    wrong = copy.deepcopy(saved_search)
    wrong["objects"][0]["arity"] = 0
    assert errors(schemas, "environment-snapshot", wrong)

    for selection in ({"version": "1.6.0", "class": ""},
                      {"version": "1.6.0", "class_uid": 2 ** 64}):
        wrong = copy.deepcopy(bundle)
        wrong["schemas"][0] = {
            "id": "ocsf", "kind": "ocsf",
            "target": {"kind": "ocsf", "catalog": {"compile_version": 1, "extensions": {"custom": {"description": None}}}, "selection": selection},
            "provenance": bundle["provenance"],
        }
        assert errors(schemas, "field-schema-bundle", wrong), selection


def test_environment_scope_identity_and_macro_argument_contracts(schemas):
    case = json.loads((ROOT / "testdata/environment/cases.json").read_text())[0]
    snapshot = case["snapshot"]
    bundle = case["schema_bundle"]

    for selector in ("namespace", "app", "owner"):
        restricted = copy.deepcopy(snapshot)
        restricted["capture_scope"][selector] = {"values": ["search"]}
        assert errors(schemas, "environment-snapshot", restricted), selector
        for collection in restricted["collections"]:
            if collection["kind"] in ("index", "source", "sourcetype"):
                collection.update(coverage="partial", reason="restricted capture")
        assert not errors(schemas, "environment-snapshot", restricted), selector

    for context in ("namespace", "app", "owner"):
        wrong = copy.deepcopy(bundle)
        wrong["bindings"][0]["expected"][context] = "search"
        assert errors(schemas, "field-schema-bundle", wrong), context
    scoped_binding = copy.deepcopy(bundle)
    scoped_binding["bindings"][0]["expected"] = {"kind": "macro", "name": "daily", "app": "search"}
    assert not errors(schemas, "field-schema-bundle", scoped_binding)

    macro = copy.deepcopy(snapshot)
    macro["objects"][0].update(kind="macro", arity=2, arguments=["host", "host"])
    assert errors(schemas, "environment-snapshot", macro)
    macro["objects"][0]["arguments"] = ["host", "source"]
    assert not errors(schemas, "environment-snapshot", macro)


def test_environment_snapshot_versions_and_published_export_examples(schemas):
    cases = json.loads((ROOT / "testdata/environment/cases.json").read_text())
    for case in cases:
        snapshot = case["snapshot"]
        version = snapshot["schema_version"]
        assert errors(schemas, "environment-snapshot", snapshot, version=f"v{3 - version}")
        assert not errors(schemas, "environment-validation", {"schema_version": 1, "snapshot": snapshot}, "Request")
    snapshot = json.loads((ROOT / "examples/environment/observed-partial-snapshot.json").read_text())
    report = json.loads((ROOT / "examples/environment/export-report.json").read_text())
    assert not errors(schemas, "environment-snapshot", snapshot, version="v2")
    assert not errors(schemas, "environment-export-report", report)
    assert report["snapshot_digest"] == snapshot["digest"]
    assert report["observation"] == snapshot["observation"]
    assert report["collections"] == snapshot["collections"]
    assert {item["kind"] for item in report["collections"] if item.get("reason") == "adapter_unsupported"} == {"dataset", "module", "function", "external_command"}
    for member in report:
        wrong = copy.deepcopy(report)
        del wrong[member]
        assert errors(schemas, "environment-export-report", wrong), member
    for member, value in (("schema_version", 2), ("status", "valid"), ("observation", None), ("limits", {}), ("snapshot_digest", "sha256:bad")):
        wrong = copy.deepcopy(report)
        wrong[member] = value
        assert errors(schemas, "environment-export-report", wrong), member


def test_observed_v2_shape_boundaries(schemas):
    cases = json.loads((ROOT / "testdata/environment/cases.json").read_text())
    original = next(c["snapshot"] for c in cases if c["name"] == "observed-v2-complete")
    validate = lambda value: errors(schemas, "environment-snapshot", value, version="v2")
    for field in original["observation"]:
        wrong = copy.deepcopy(original)
        del wrong["observation"][field]
        assert validate(wrong), field
    for value in (None, {}, {"mode": "bounded"}, {"mode": "all_retained", "earliest": "2026-10-01T00:00:00Z"}):
        wrong = copy.deepcopy(original)
        wrong["observation"]["window"] = value
        assert validate(wrong), value
    for bound in ("2026-10-01T0:00:00Z", "2026-10-01T00:00:00,1Z", "2026-10-01T00:00:00.1234567890Z", "2026-10-01T00:00:00+01:00", ""):
        wrong = copy.deepcopy(original)
        wrong["observation"]["window"] = {"mode": "bounded", "earliest": bound}
        assert validate(wrong), bound
    for bound in ("2026-10-01T00:00:00Z", "2026-10-01T00:00:00.123456789+00:00", "2026-10-01T00:00:00-00:00"):
        value = copy.deepcopy(original)
        value["observation"]["window"] = {"mode": "bounded", "latest": bound}
        assert not validate(value), bound
    for field, value in (("sharing", "app"), ("arguments", []), ("relations", []), ("document", {"text": "index=main"})):
        wrong = copy.deepcopy(original)
        wrong["objects"][0][field] = value
        assert validate(wrong), field
    for field, value in (("method", "other"), ("visibility", "other"), ("time_precision", "exact"), ("absence_meaning", "absent"), ("indexes", None), ("captures", None)):
        wrong = copy.deepcopy(original)
        wrong["observation"][field] = value
        assert validate(wrong), field
    for coverage, reason, valid in (("complete", "", True), ("complete", "failure", False), ("partial", "", False), ("unavailable", "timeout", True)):
        value = copy.deepcopy(original)
        capture = value["observation"]["captures"][0]
        capture.update(coverage=coverage, reason=reason)
        assert (not validate(value)) == valid, (coverage, reason)
    duplicates = copy.deepcopy(original)
    duplicates["observation"]["indexes"][0]["catalog_datatypes"] *= 2
    duplicates["observation"]["indexes"][0]["required_datatypes"] *= 2
    duplicates["observation"]["captures"][1]["object_ids"] *= 2
    assert not validate(duplicates)
    # Shape success does not establish object linkage, capture rollups, or remote exhaustiveness.
    invalid_ref = next(c["snapshot"] for c in cases if c["name"] == "observed-v2-invalid-reference")
    assert not validate(invalid_ref)


def test_compatibility_live_contracts_and_independent_source_facts(schemas, emitted):
    examples = emitted["compatibility-examples"]
    for outcome in ("satisfied", "unsatisfied", "incomplete", "not assessed"):
        report = examples[outcome]
        assert report["outcome"] == outcome
        assert not errors(schemas, "compatibility", report)
        assert not errors(schemas, "compatibility-request", examples[outcome + "-request"])
    chain = examples["chain"]
    assert chain["outcome"] == "satisfied"
    assert chain["correlation"]["outcome"] == "connected"
    assert len(chain["correlation"]["nodes"]) == 3 and len(chain["correlation"]["edges"]) == 2
    inputs = {source["id"]: source["name"] for source in chain["requirements"]["inputs"]}
    fields = {(inputs.get(item.get("input_id")), tuple(item["field_identity"]["segments"]))
              for item in chain["requirements"]["items"] if item["kind"] == "field"}
    assert fields == {("$events", ("user_id",)), ("$events", ("asset_id",)), ("$users", ("id",)), ("$assets", ("id",))}
    for name in ("chain", "sql-left", "absent"):
        assert not errors(schemas, "compatibility", examples[name]), name
        assert not errors(schemas, "compatibility-request", examples[name + "-request"]), name
    assert examples["absent"]["outcome"] == "unsatisfied"
    assert any(reason["code"] == "schema_not_supplied" for reason in examples["absent"]["reasons"])


def test_compatibility_shape_and_runtime_reference_admission(schemas, emitted, emitter):
    request = emitted["compatibility-examples"]["satisfied-request"]
    mutations = []
    for member in ("inputs", "input_coverage", "field_attribution_coverage", "correlation"):
        value = copy.deepcopy(request)
        del value["requirements"][member]
        mutations.append(("missing " + member, value, False))
    for member in ("document", "schema_bundle", "dependency_bindings"):
        value = copy.deepcopy(request)
        value[member] = None
        mutations.append(("null " + member, value, False))
    value = copy.deepcopy(request)
    value["requirements"]["future_annotation"] = True
    mutations.append(("unknown requirements member", value, False))
    value = copy.deepcopy(request)
    value["input_bindings"].append(copy.deepcopy(value["input_bindings"][0]))
    mutations.append(("duplicate binding", value, False))
    value = copy.deepcopy(request)
    value["input_bindings"][0]["input_id"] = "unknown"
    mutations.append(("unknown binding link", value, True))
    value = copy.deepcopy(request)
    value["requirements"]["correlation"]["nodes"][0]["occurrence_id"] = "unknown"
    mutations.append(("unknown graph link", value, True))
    for name, value, shape_valid in mutations:
        assert (not errors(schemas, "compatibility-request", value)) == shape_valid, name
        assert subprocess.check_output([str(emitter), "compatibility-request", json.dumps(value)], text=True) == "rejected", name
    assert subprocess.check_output([str(emitter), "compatibility-request", json.dumps(request)], text=True) == "accepted"
    raw = json.dumps(request)
    raw = raw.replace('"schema_version": 1', '"schema_version": 1, "schema_version": 1', 1)
    assert subprocess.check_output([str(emitter), "compatibility-request", raw], text=True) == "rejected"
    for member in ("inputs", "requirement_outcomes", "coverage", "reasons"):
        value = copy.deepcopy(emitted["compatibility-examples"]["satisfied"])
        value[member] = None
        assert errors(schemas, "compatibility", value), member


def test_compatibility_authored_runtime_cases(emitter):
    fixture = json.loads((ROOT / "testdata/tooling/contracts.json").read_text())
    for case in fixture["cases"]:
        if case["family"] == "compatibility-request" and "runtime_valid" in case:
            actual = subprocess.check_output([str(emitter), "compatibility-request", json.dumps(case["instance"])], text=True)
            assert (actual == "accepted") == case["runtime_valid"], case["id"]


def test_resolution_runtime_reports_and_conditional_publication(schemas, emitted):
    report = emitted["resolution"]
    assert not errors(schemas, "resolution-request", emitted["resolution-request"])
    assert report["counts"] == {"verified": 1, "failed": 1, "incomplete": 0}
    assert report["total_combinations"] == "2" and report["generated_count"] == 2
    assert [v["ordinal"] for v in report["variants"]] == [1, 2]
    assert [v["selection"][0]["value"] for v in report["variants"]] == ["events_good", "events_missing"]
    for name in ("resolution", "resolution-noop"):
        instance = emitted[name]
        assert not errors(schemas, "resolution", instance), [e.message for e in errors(schemas, "resolution", instance)]
        for variant in instance["variants"]:
            assert ("resolved_query" in variant) == (variant["outcome"] == "verified")
            for member in ("proof", "selection", "provenance"):
                wrong = copy.deepcopy(instance)
                del wrong["variants"][variant["ordinal"] - 1][member]
                assert errors(schemas, "resolution", wrong), member
            for member in ("selection", "changes", "diagnostics"):
                wrong = copy.deepcopy(instance)
                wrong["variants"][variant["ordinal"] - 1][member] = None
                assert errors(schemas, "resolution", wrong), member
        for member in ("resolutions", "variants"):
            wrong = copy.deepcopy(instance)
            wrong[member] = None
            assert errors(schemas, "resolution", wrong)
    assert emitted["resolution-noop"]["total_combinations"] == "1"
    assert emitted["resolution-noop"]["variants"][0]["changes"] == []
    failed = copy.deepcopy(report)
    failed["variants"][1]["resolved_query"] = failed["variants"][1]["candidate_text"]
    assert errors(schemas, "resolution", failed)
    failed["variants"][1]["outcome"] = "incomplete"
    assert errors(schemas, "resolution", failed)
    wrong = copy.deepcopy(report)
    del wrong["variants"][0]["resolved_query"]
    assert errors(schemas, "resolution", wrong)
    for member in ("references", "roles", "limitations"):
        wrong = copy.deepcopy(report)
        wrong["variants"][0]["proof"][member] = None
        assert errors(schemas, "resolution", wrong)
    for member in ("input_bindings", "dependency_bindings", "inputs", "requirement_outcomes", "coverage", "reasons", "diagnostics"):
        wrong = copy.deepcopy(report)
        wrong["variants"][0]["compatibility"][member] = None
        assert errors(schemas, "resolution", wrong), member
    wrong = copy.deepcopy(report)
    wrong["variants"][0]["compatibility"]["effective_dependency_bindings"] = None
    assert errors(schemas, "resolution", wrong)


def test_resolution_strict_requests_and_runtime_membership(schemas, emitted, emitter):
    request = emitted["resolution-request"]
    for value in (None, 0, -1, 1.5, 2**64):
        wrong = copy.deepcopy(request)
        wrong["max_variants"] = value
        assert errors(schemas, "resolution-request", wrong)
        assert subprocess.check_output([str(emitter), "resolution-request", json.dumps(wrong)], text=True) == "rejected"
    for member in ("document", "resolutions", "compatibility"):
        wrong = copy.deepcopy(request)
        wrong[member] = None
        assert errors(schemas, "resolution-request", wrong)
    wrong = copy.deepcopy(request)
    wrong["compatibility"]["input_bindings"][0]["original_input_id"] = "unknown-role"
    assert not errors(schemas, "resolution-request", wrong)  # Membership belongs to Go.
    assert subprocess.check_output([str(emitter), "resolution-request", json.dumps(wrong)], text=True) == "rejected"
    for member, value in (("values", []), ("values", [""]), ("kind", "field")):
        wrong = copy.deepcopy(request)
        wrong["resolutions"][0][member] = value
        assert errors(schemas, "resolution-request", wrong)
    assert subprocess.check_output([str(emitter), "resolution-request", json.dumps(request)], text=True) == "accepted"


def test_resolution_native_cli_go_contract_parity(schemas, emitted):
    library = os.environ.get("SPL_NATIVE_LIBRARY")
    cli = os.environ.get("SPL_CLI")
    if not library or not cli:
        pytest.skip("SPL_NATIVE_LIBRARY and SPL_CLI select built transport artifacts")
    from spl_toolkit import SPLMapper
    request = json.loads((ROOT / "examples/resolution/request.json").read_text())
    with SPLMapper(library_path=library) as mapper:
        native = mapper.resolve(request)
        assert not errors(schemas, "resolution", native)
        assert native == emitted["resolution"]
        request["document"] = {"text": "from [{id:1}]", "language": "spl2"}
        request["resolutions"] = []
        request["compatibility"]["input_bindings"] = []
        noop = mapper.resolve(request)
        assert not errors(schemas, "resolution", noop)
        assert noop == emitted["resolution-noop"]
    result = subprocess.run([cli, "resolve", "--request", str(ROOT / "examples/resolution/request.json"),
                             "--format", "json"], capture_output=True, text=True)
    assert result.returncode == 1 and result.stderr == ""
    assert json.loads(result.stdout) == native
