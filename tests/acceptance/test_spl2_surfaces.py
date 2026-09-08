"""Canonical SPL2 obligations and complete CLI/HTTP/native report equality.

Canonical expectations are authored independently in the durable corpus and are
also checked by Go. Transport equality never replaces those semantic assertions.
Source runs use the explicit absolute library override; installed runs omit it.
"""
from concurrent.futures import ThreadPoolExecutor
from copy import deepcopy
import hashlib
import json
import os
from pathlib import Path
import subprocess

import pytest

from spl_toolkit import SPLMapper
from test_surfaces import cli_path, server_url, post_json, required_absolute_path
from test_schema_surfaces import cli_arguments, load_target
from spl2_transport import digest, load_transport, prepare_transport


FIXTURES = required_absolute_path("SPL_SPL2_FIXTURES")
MANIFEST = json.loads((FIXTURES / "manifest.json").read_text(encoding="utf-8"))
EXITS = {"valid": 0, "invalid": 1, "incomplete": 3}
SELECTORS = (("language", "--language"), ("profile", "--profile"),
             ("version", "--compatibility-version"), ("source_id", "--source-id"))


def open_mapper():
    options = {"library_path": str(required_absolute_path("SPL_NATIVE_LIBRARY"))} if "SPL_NATIVE_LIBRARY" in os.environ else {}
    return SPLMapper(**options)


def canonical_projection(report):
    after = report["lineage"][-1]["after"] if report["lineage"] else {"fields": [], "removed": [], "open": False, "uncertain": False}
    return {"phase": "analysis", "scope": "canonical-result", "status": report["status"],
            "syntax_complete": report["coverage"]["syntax_complete"],
            "semantic_complete": report["coverage"]["semantic_complete"],
            "expected_codes": sorted({d["code"] for d in report["diagnostics"]}),
            "references": [{k:r[k] for k in ("original_name", "normalized_name", "kind", "role", "binding")} |
                           {"start":r["location"]["start"]["offset"], "end":r["location"]["end"]["offset"]}
                           for r in report["references"]],
            "fields": [{"name": f["name"], "conditional": f["conditional"]} for f in after["fields"]],
            "removed": after["removed"], "open": after["open"], "uncertain": after["uncertain"],
            "stage_commands": [s["command"] for s in report["stages"]],
            "stage_complete": [s["semantic_complete"] for s in report["stages"]]}


def run_cli(cli, command, document, extra=()):
    document = {"source_id": ""} | document
    args = [str(cli), command, "--format", "json", *extra]
    args += [value for key, flag in SELECTORS if key in document for value in (flag, document[key])]
    query_args = ["--query", document["text"]] if command == "analyze" else ["--stdin"]
    result = subprocess.run(args + query_args, input=document["text"].encode("utf-8") if command != "analyze" else None,
                            capture_output=True, timeout=15)
    assert result.stdout, (document, result.returncode, result.stderr)
    report = json.loads(result.stdout)
    assert (result.returncode, result.stderr) == (EXITS[report["status"]], b"")
    return report


def assert_document(report, document):
    assert report["document"] == {"language": "spl", "profile": "splunkd", "version": "current", "source_id": ""} | document
    raw = document["text"].encode("utf-8")
    for reference in report["references"]:
        start, end = reference["location"]["start"]["offset"], reference["location"]["end"]["offset"]
        assert 0 <= start < end <= len(raw)
        assert raw[start:end].decode("utf-8") == reference["original_name"]


@pytest.fixture(scope="session")
def spl2_go_reports(tmp_path_factory):
    root = Path(__file__).resolve().parents[2]
    source_root = root if (root / "go.mod").is_file() else None
    if "SPL_SPL2_GO_REPORTS" in os.environ:
        path = required_absolute_path("SPL_SPL2_GO_REPORTS")
        expected = os.environ["SPL_SPL2_GO_SHA256"]
    else:
        assert source_root is not None, "installed acceptance requires copied Go transport evidence"
        path = prepare_transport(source_root, tmp_path_factory.mktemp("spl2-go") / "reports.json")
        expected = digest(path)
    artifact, reports = load_transport(path, FIXTURES, expected, source_root)
    return {"path": str(path), "sha256": expected, "bytes": path.stat().st_size, "source_hashes": artifact["source_hashes"], "reports": reports}


@pytest.fixture(scope="session")
def spl2_evidence(cli_path, spl2_go_reports):
    with open_mapper() as mapper:
        library = Path(mapper._lib._name).resolve()
    digest = lambda path: hashlib.sha256(path.read_bytes()).hexdigest()
    record = {"schema_version": 1, "fixture_hashes": {p.name: digest(p) for p in sorted(FIXTURES.glob("*.json"))},
              "artifacts": {str(p): digest(p) for p in (cli_path, required_absolute_path("SPL_SERVER"), library)},
              "go_transport": {k:v for k,v in spl2_go_reports.items() if k != "reports"},
              "corpus": [], "semantic_controls": [], "validation": [], "batches": [], "concurrent_calls": 0}
    yield record
    if destination := os.environ.get("SPL_SPL2_EVIDENCE"):
        Path(destination).write_text(json.dumps(record, indent=2, sort_keys=True, ensure_ascii=False) + "\n", encoding="utf-8")


@pytest.mark.parametrize("filename", MANIFEST["case_files"])
def test_spl2_corpus_full_report_parity(filename, cli_path, server_url, spl2_evidence, spl2_go_reports):
    cases = json.loads((FIXTURES / filename).read_text(encoding="utf-8"))
    assert cases, "required corpus file is empty"
    with open_mapper() as mapper:
        for case in cases:
            assert isinstance(case.get("canonical"), dict), f"missing canonical obligation: {case['id']}"
            document = deepcopy(case["document"])
            options = {k:v for k,v in document.items() if k != "text"}
            native = mapper.analyze_query(document["text"], **options)
            assert canonical_projection(native) == case["canonical"], case["id"]
            assert native == spl2_go_reports["reports"][case["id"]], case["id"]
            assert_document(native, document)
            status, http = post_json(server_url, "/query/analyze", document)
            cli = run_cli(cli_path, "analyze", document)
            assert status == 200 and cli == http == native, case["id"]
            spl2_evidence["corpus"].append({"id": case["id"], "document": document,
                                           "canonical": case["canonical"], "cli": cli, "http": http, "native": native})


@pytest.mark.parametrize("query,status,output,binding,conditional", [
    ("SELECT count() AS n FROM main HAVING n>2", "valid", "n", "derived", False),
    ("FROM main SELECT count() AS n HAVING n>2", "valid", "n", "derived", False),
    ("FROM main SELECT host HAVING host>2", "incomplete", "host", "source", False),
    ("FROM main SELECT count() AS n HAVING hidden>2", "incomplete", "n", "indeterminate", False),
    ("FROM main SELECT sum(bytes) AS total HAVING total>2", "valid", "total", "indeterminate", True),
])
def test_spl2_whole_input_having_and_visibility(query, status, output, binding, conditional, cli_path, server_url, spl2_evidence):
    document = {"text": query, "language": "spl2", "source_id": "having-é😀\r\n"}
    with open_mapper() as mapper:
        native = mapper.analyze_query(query, language="spl2", source_id=document["source_id"])
    assert native["status"] == status
    assert native["coverage"]["syntax_complete"] is True
    assert native["coverage"]["semantic_complete"] == (status == "valid")
    assert [l["phase"] for l in native["lineage"]] == ["source", "evaluate" if output == "host" else "aggregate", "having", "project"]
    assert [l["execution_order"] for l in native["lineage"]] == [0, 1, 2, 3]
    assert canonical_projection(native)["fields"] == [{"name": output, "conditional": conditional}]
    having = next(s for s in native["stages"] if s["command"] == "having")
    reads = [r for r in native["references"] if r["stage_id"] == having["id"] and r["kind"] == "field"]
    assert len(reads) == 1 and reads[0]["binding"] == binding
    if binding == "derived":
        created = next(r for r in native["references"] if r["role"] == "output" and r["normalized_name"] == output)
        assert reads[0]["origin_reference_ids"] == [created["id"]]
    assert_document(native, document)
    http_status, http = post_json(server_url, "/query/analyze", document)
    cli = run_cli(cli_path, "analyze", document)
    assert http_status == 200 and cli == http == native
    spl2_evidence["semantic_controls"].append({"document": document, "cli": cli, "http": http, "native": native})


# These literal obligations reuse the accepted public Go validation tests. They
# also exercise unknown semantics with an otherwise matching target, and an
# independently definite missing read alongside incomplete semantics.
VALIDATION_CASES = [
    ("legacy", "table host", "spl", "valid", ["matching"], ["required"]),
    ("selected", "FROM main SELECT host", "spl2", "valid", ["matching"], ["required"]),
    ("missing", "FROM main SELECT absent", "spl2", "invalid", ["missing"], ["missing"]),
    ("unknown", "FROM main | mystery x=1 | table host", "spl2", "incomplete", ["indeterminate"], ["indeterminate"]),
    ("mixed", "FROM main | table absent | mystery x=1", "spl2", "invalid", ["missing"], ["missing"]),
    ("null-inspection", "FROM main SELECT isnull(absent) AS answer", "spl2", "valid", [], []),
    ("ordinary-read", "FROM main | eval answer=isnull(absent) | where absent>0", "spl2", "invalid", ["missing"], ["missing"]),
    ("null-deletion", "FROM main | eval output=null | table output", "spl2", "invalid", ["unavailable"], ["unavailable"]),
    ("conditional", 'FROM [{host:"a"},{other:1}] SELECT host', "spl2", "incomplete", ["indeterminate"], ["indeterminate"]),
]
FIELD_TARGET = {"fields": ["host"], "identity": "surface-fields", "version": "1"}
SCHEMA_TARGET = {"kind": "json_schema", "identity": "surface-schema", "schema": {
    "type": "object", "properties": {"host": {"type": "string"}}, "required": ["host"], "additionalProperties": False}}


def validation_cli_args(cli, kind, target, directory):
    if kind == "schema":
        # CLI has no free-form target identity flag; parity uses the target's
        # transport-supported fields, keeping all returned metadata exact.
        return cli_arguments(cli, target, directory)[4:]
    path = directory / "fields.json"
    path.write_text(json.dumps(target), encoding="utf-8")
    return ["--fields", str(path)]


@pytest.mark.parametrize("kind", ["fields", "schema"])
def test_spl2_validation_single_and_ordered_batches(kind, cli_path, server_url, tmp_path, spl2_evidence):
    target = deepcopy(FIELD_TARGET if kind == "fields" else SCHEMA_TARGET)
    if kind == "schema":
        target.pop("identity")
    reports, documents = [], []
    args = validation_cli_args(cli_path, kind, target, tmp_path)
    with open_mapper() as mapper:
        for identity, query, language, expected_status, fields, schema in VALIDATION_CASES:
            document = {"text": query, "language": language, "source_id": "é😀\r\n:" + identity}
            native = getattr(mapper, "validate_" + kind)(query, target, language=language, source_id=document["source_id"])
            assert native["status"] == expected_status, identity
            assert [o["outcome"] for o in native["outcomes"]] == (fields if kind == "fields" else schema), identity
            assert_document(native["analysis"], document)
            null_reads = {r["id"] for r in native["analysis"]["references"] if r["role"] == "null_test"}
            assert not null_reads.intersection(o["reference_id"] for o in native["outcomes"])
            status, http = post_json(server_url, "/query/validate-" + kind, {"document": document, "catalog" if kind == "fields" else "target": target})
            cli = run_cli(cli_path, "validate-" + kind, document, args)
            assert status == 200 and cli == http == native, identity
            documents.append(document)
            reports.append(native)
            spl2_evidence["validation"].append({"id": identity, "kind": kind, "document": document, "target": target,
                                               "cli": cli, "http": http, "native": native})
        documents.reverse()
        expected = {"schema_version": 1, "status": "invalid", "reports": list(reversed(reports))}
        native = getattr(mapper, "validate_" + kind + "_batch")(documents, target)
        status, http = post_json(server_url, "/query/validate-" + kind + "/batch", {"documents": documents, "catalog" if kind == "fields" else "target": target})
        result = subprocess.run([str(cli_path), "validate-" + kind, "--format", "json", *args, "--batch", "-"],
                                input=json.dumps(documents), capture_output=True, text=True, timeout=15)
        assert (result.returncode, result.stderr) == (1, "")
        assert status == 200 and json.loads(result.stdout) == http == native == expected
        spl2_evidence["batches"].append({"kind": kind, "documents": documents, "target": target, "report": expected})


def test_spl2_mixed_concurrent_analysis_preserves_dialect(cli_path, server_url, spl2_evidence):
    documents = [{"text": "table host", "source_id": "spl"},
                 {"text": "SELECT host FROM main WHERE bytes>0", "language": "spl2", "source_id": "sql"},
                 {"text": 'FROM main | mystery "open', "language": "spl2", "source_id": "invalid-literal"},
                 {"text": "FROM main | where user IS NOT string", "language": "spl2", "source_id": "held"}]
    statuses = ["valid", "valid", "invalid", "incomplete"]
    with open_mapper() as mapper:
        expected = [mapper.analyze_query(d["text"], **{k:v for k,v in d.items() if k != "text"}) for d in documents]
        assert [r["status"] for r in expected] == statuses
        sql = expected[1]
        assert [s["command"] for s in sql["stages"]] == ["select", "from", "where"]
        assert [s["position"] for s in sql["stages"]] == [2, 0, 1]
        assert [l["phase"] for l in sql["lineage"]] == ["source", "filter", "evaluate", "project"]
        assert [l["execution_order"] for l in sql["lineage"]] == [0, 1, 2, 3]
        assert [r["normalized_name"] for r in sql["references"] if r["role"] == "read" and r["kind"] == "field"] == ["host", "bytes"]
        assert "SPL_SYNTAX_ERROR" in {d["code"] for d in expected[2]["diagnostics"]}
        def call(index):
            document, report = documents[index % len(documents)], expected[index % len(documents)]
            native = mapper.analyze_query(document["text"], **{k:v for k,v in document.items() if k != "text"})
            status, http = post_json(server_url, "/query/analyze", document)
            cli = run_cli(cli_path, "analyze", document)
            assert status == 200 and cli == native == http == report
            assert_document(native, document)
        with ThreadPoolExecutor(max_workers=4) as executor:
            list(executor.map(call, range(32)))
    spl2_evidence["concurrent_calls"] = 32


@pytest.mark.parametrize("language,query", [("spl", "table file.name file.xattributes.vendor_field"),
                                           ("spl2", "FROM main SELECT file.name")])
def test_spl2_owned_ocsf_path_boundary(language, query, cli_path, server_url, tmp_path, spl2_evidence):
    case = {"target": {"kind": "ocsf", "selection": {"version": "1.6.0", "class": "file_activity", "profiles": [], "extensions": []}},
            "catalog_fixture": "ocsf/1.6.0/base.json.gz"}
    target = load_target(case)
    document = {"text": query, "language": language, "source_id": "ocsf-path"}
    with open_mapper() as mapper:
        native = mapper.validate_schema(query, target, language=language, source_id=document["source_id"])
    if language == "spl":
        assert native["status"] == "valid"
        assert [o["outcome"] for o in native["outcomes"]] == ["required", "permitted_unspecified"]
    else:
        assert native["status"] == "incomplete"
        assert [o["outcome"] for o in native["outcomes"]] == ["required", "indeterminate"]
    status, http = post_json(server_url, "/query/validate-schema", {"document": document, "target": target})
    args = cli_arguments(cli_path, target, tmp_path, case)[4:]
    cli = run_cli(cli_path, "validate-schema", document, args)
    assert status == 200 and cli == native == http
    spl2_evidence["validation"].append({"id": "ocsf-" + language, "document": document,
                                       "catalog_fixture": case["catalog_fixture"], "cli": cli, "http": http, "native": native})
