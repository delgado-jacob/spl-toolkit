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
            "impact", "document-view", "lsp-configuration")


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
        result[path.name] = validator(schema, registry=registry,
                                     format_checker=jsonschema.FormatChecker())
    return result


def errors(schemas, family, instance, definition=None):
    name = f"{family}.schema.json" if family != "sarif" else "sarif-schema-2.1.0.json"
    validator = schemas[name]
    if definition:
        validator = validator.evolve(
            schema={"$ref": validator.schema["$id"] + "#/$defs/" + definition})
    return list(validator.iter_errors(instance))  # Never short-circuit the iterator.


CAPABILITY_DIMENSIONS = ("syntax", "semantics", "requirements", "linting", "safe_rewriting")


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


def shared_errors(schemas, definition, instance):
    validator = schemas["shared.schema.json"]
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
 "github.com/delgado-jacob/spl-toolkit/pkg/rewrite"
 "github.com/delgado-jacob/spl-toolkit/pkg/corpus"
 "github.com/delgado-jacob/spl-toolkit/pkg/corpusio"
 "github.com/delgado-jacob/spl-toolkit/pkg/document"
 "github.com/delgado-jacob/spl-toolkit/pkg/graph"
 "github.com/delgado-jacob/spl-toolkit/pkg/impact"
 "github.com/delgado-jacob/spl-toolkit/pkg/sarif"
)
func must[T any](v T, e error) T { if e != nil { panic(e) }; return v }
func main() {
 if len(os.Args)>1 {
  raw := []byte(os.Args[2]); var e error
  switch os.Args[1] {
  case "corpus": _,e=corpus.DecodeRequest(raw)
  case "manifest": _,e=corpusio.DecodeManifest(raw)
  case "rewrite": _,e=rewrite.DecodeRequest(raw)
  case "impact": _,e=impact.DecodeSchemaRequest(raw)
  case "query-document": _,e=validation.DecodeDocuments(append(append([]byte{'['},raw...),']'))
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
