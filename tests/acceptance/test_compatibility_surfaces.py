"""Full canonical compatibility values and error details across built transports."""

import copy
import json
import os
import subprocess
import shutil
from urllib.error import HTTPError
from urllib.request import Request, urlopen

import pytest

from spl_toolkit import SPLMapper, SPLMapperError
from test_surfaces import cli_path, required_absolute_path, server_url
from test_requirements_surfaces import (call_python, call_raw_c, expected_requirements_exit,
                                        post_document, run_cli)

FIXTURES = required_absolute_path("SPL_COMPATIBILITY_FIXTURES")
assert FIXTURES.is_file(), "SPL_COMPATIBILITY_FIXTURES must name cases.json"
CASES = json.loads(FIXTURES.read_text(encoding="utf-8"))
EXITS = {"satisfied": 0, "unsatisfied": 1, "incomplete": 3, "not assessed": 3}


def open_mapper():
    return SPLMapper(library_path=str(required_absolute_path("SPL_NATIVE_LIBRARY")))


def make_request(mapper, case):
    document = case["document"]
    requirements = mapper.requirements_query(document["text"], **{key: document[key] for key in
                                              ("language", "profile", "version", "source_id")})
    request = {"schema_version": 1, "requirements": requirements, "snapshot": copy.deepcopy(case["snapshot"]),
               "query_scope": copy.deepcopy(case["query_scope"]), "input_bindings": []}
    if case.get("schema_bundle") is not None:
        request["schema_bundle"] = copy.deepcopy(case["schema_bundle"])
    bindings_by_key = {}
    for source in requirements["inputs"]:
        key = source["kind"] + ":" + source["name"]
        if key not in case["bindings"]:
            continue
        binding = copy.deepcopy(case["bindings"][key])
        binding.pop("input", None)
        binding["input_id"] = source["id"]
        if "expected" not in binding:
            obj = next(obj for obj in case["snapshot"]["objects"] if obj["id"] == binding["object_id"])
            binding["expected"] = {key: obj[key] for key in ("kind", "name", "namespace", "app", "owner")}
        request["input_bindings"].append(binding)
        bindings_by_key[key] = binding
    for mutation in case.get("request_mutations", []):
        operation = mutation["op"]
        if operation in ("set", "remove"):
            parts = mutation["path"].strip("/").split("/")
            target = request
            for part in parts[:-1]:
                target = target[part]
            if operation == "set":
                target[parts[-1]] = copy.deepcopy(mutation["value"])
            else:
                del target[parts[-1]]
        elif operation == "duplicate_binding":
            request["input_bindings"].append(copy.deepcopy(bindings_by_key[mutation["input"]]))
        else:
            field = {"set_binding_input_id": "input_id", "set_binding_expected": "expected",
                     "set_binding_schema_id": "schema_id"}[operation]
            bindings_by_key[mutation["input"]][field] = copy.deepcopy(mutation["value"])
    return request



def post_raw(base_url, raw, content_type="application/json"):
    request = Request(base_url + "/query/compatibility", data=raw,
                      headers={"Content-Type": content_type}, method="POST")
    try:
        with urlopen(request, timeout=10) as response:
            return response.status, json.load(response)
    except HTTPError as response:
        with response:
            return response.code, json.load(response)


def native_raw(mapper, raw):
    pointer = mapper._lib.spl_mapper_check_compatibility(mapper._mapper_id, raw)
    assert pointer
    try:
        if pointer.contents.error is not None:
            assert pointer.contents.result is None
            return "error", json.loads(pointer.contents.error)
        assert pointer.contents.result is not None
        return "report", json.loads(pointer.contents.result)
    finally:
        mapper._lib.spl_result_free(pointer)


GO_REPORT_HELPER = r"""
package main
import (
    "encoding/json"
    "fmt"
    "io"
    "os"
    "sync"
    "github.com/delgado-jacob/spl-toolkit/pkg/analysis"
    "github.com/delgado-jacob/spl-toolkit/pkg/compatibility"
)
func main() {
    raw, err := io.ReadAll(os.Stdin); if err != nil { panic(err) }
    if len(os.Args) > 1 && os.Args[1] == "discovery" {
        var doc analysis.QueryDocument
        if err := json.Unmarshal(raw, &doc); err != nil { panic(err) }
        analyzed, err := analysis.Analyze(doc); if err != nil { panic(err) }
        requirements, err := analysis.Requirements(doc); if err != nil { panic(err) }
        if err := json.NewEncoder(os.Stdout).Encode(map[string]interface{}{"analysis": analyzed, "requirements": requirements}); err != nil { panic(err) }
        return
    }
    if len(os.Args) > 1 && os.Args[1] == "prepared" {
        var requests []compatibility.Request
        if err := json.Unmarshal(raw, &requests); err != nil { panic(err) }
        prepared, err := compatibility.Prepare(requests[0].Snapshot, requests[0].SchemaBundle)
        if err != nil { panic(err) }
        // Mutate caller-owned evidence after preparation, using only public data.
        requests[0].Snapshot.Objects[0].Name = "caller mutation"
        requests[0].SchemaBundle.Schemas[0].Catalog[0] = 'X'
        assess := func(index int) *compatibility.Report {
            r := requests[index]
            report, err := prepared.Check(compatibility.AssessmentRequest{
                SchemaVersion: r.SchemaVersion, Requirements: r.Requirements,
                QueryScope: r.QueryScope, InputBindings: r.InputBindings,
                Document: r.Document, DependencyBindings: r.DependencyBindings})
            if err != nil { panic(err) }; return report
        }
        mutated := assess(0)
        mutated.Requirements.Items[0].Identity = "report mutation"
        mutated.Inputs[0].Objects[0].Object.Name = "report mutation"
        mutated.Reasons = nil
        if mutated.Closure != nil {
            mutated.Closure.EffectiveAnalysis.Document.Text = "report mutation"
            mutated.Inputs[0].Occurrences[0].SourceIntervals[0].ObjectID = "report mutation"
        }
        reports := []*compatibility.Report{assess(0), assess(1)}
        concurrent := make([]*compatibility.Report, 8)
        var wait sync.WaitGroup
        for i := range concurrent {
            wait.Add(1); go func(i int) { defer wait.Done(); concurrent[i] = assess(i % 2) }(i)
        }
        wait.Wait()
        if err := json.NewEncoder(os.Stdout).Encode(map[string]interface{}{"reports": reports, "concurrent": concurrent}); err != nil { panic(err) }
        return
    }
    report, err := compatibility.CheckJSON(raw)
    if err != nil {
        detail, ok := compatibility.RequestErrorDetails(err); if !ok { panic(err) }
        if err := json.NewEncoder(os.Stdout).Encode(detail); err != nil { panic(err) }
        return
    }
    if len(os.Args) > 1 && os.Args[1] == "text" {
        fmt.Print(compatibility.FormatReport(report))
        return
    }
    if err := json.NewEncoder(os.Stdout).Encode(report); err != nil { panic(err) }
}
"""


@pytest.fixture(scope="module")
def compatibility_go_reporter(tmp_path_factory):
    root = required_absolute_path("SPL_TOOLING_SOURCE_ROOT")
    directory = tmp_path_factory.mktemp("compatibility-go")
    try:
        helper = directory / "main.go"
        helper.write_text(GO_REPORT_HELPER, encoding="utf-8")
        binary = directory / ("compatibility-report.exe" if os.name == "nt" else "compatibility-report")
        env = dict(os.environ, GOWORK="off")
        subprocess.run([os.environ.get("SPL_TOOLING_GO", "go"), "build", "-mod=readonly", "-o", str(binary), str(helper)],
                       cwd=root, env=env, check=True, capture_output=True)
        yield binary, root, env
    finally:
        shutil.rmtree(directory)


@pytest.fixture
def compatibility_temp_dir(tmp_path):
    directory = tmp_path / "compatibility-run"
    directory.mkdir()
    try:
        yield directory
    finally:
        shutil.rmtree(directory)


def go_value(reporter, raw, *args):
    binary, root, env = reporter
    return subprocess.run([str(binary), *args], input=raw, cwd=root, env=env,
                          check=True, capture_output=True, timeout=10).stdout


def assert_authored_facts(case, report, analysis):
    expected = case["expected"]
    requirements = report["requirements"]
    assert requirements == analysis["requirements"]
    assert report["outcome"] == expected["compatibility"]["outcome"]
    assert report["correlation"]["outcome"] == expected["correlation"]["outcome"]
    assert len(report["correlation"]["edges"]) == expected["correlation"]["edge_count"]
    assert set(expected["compatibility"]["reason_codes"]) <= {reason["code"] for reason in report["reasons"]}
    assert len(requirements["inputs"]) == expected["input_count"]
    assert sum(len(source["occurrences"]) for source in requirements["inputs"]) == expected["occurrence_count"]
    sources = {source["kind"] + ":" + source["name"]: source for source in requirements["inputs"]}
    names = {source["id"]: key for key, source in sources.items()}
    text = case["document"]["text"].encode("utf-8")
    for want in expected["inputs"]:
        source = sources[want["key"]]
        assert (source["kind"], source["name"], source["identity"]["form"]) == (want["kind"], want["name"], want["form"])
        slices = [text[o["location"]["start"]["offset"]:o["location"]["end"]["offset"]].decode("utf-8")
                  for o in source["occurrences"]]
        assert sorted(slices) == sorted(want["source_slices"])
    for want in expected["field_input_pairs"]:
        assert any(item["kind"] == "field" and names.get(item.get("input_id")) == want["input"]
                   and item["field_identity"] == want["field_identity"] and item["necessity"] == want["necessity"]
                   for item in requirements["items"])
    for excluded in expected["excluded_source_fields"]:
        assert not any(item["kind"] == "field" and item["identity"] == excluded for item in requirements["items"])
    for excluded in expected.get("excluded_inputs", []):
        assert not any(source["name"] == excluded for source in sources.values())
    outcomes = {outcome["requirement_id"]: outcome for outcome in report["requirement_outcomes"]}
    for want in expected["requirement_outcomes"]:
        matches = [item for item in requirements["items"] if item["kind"] == want["kind"] and item["identity"] == want["identity"]
                   and (names.get(item.get("input_id"), "") == want.get("input", "") if want["kind"] == "field" or "input" in want else True)
                   and (item["necessity"] == want["necessity"] if "necessity" in want else True)
                   and (item.get("field_identity") == want["field_identity"] if "field_identity" in want else True)]
        assert any(outcomes[item["id"]]["outcome"] == want["outcome"] for item in matches), want
    references = {reference["id"]: reference for reference in analysis["references"]}
    def endpoint_supplies(endpoint, want):
        source = sources[want["input"]]
        if endpoint["input_id"] != source["id"] or endpoint["occurrence_id"] != source["occurrences"][want["occurrence"]]["id"]:
            return False
        # Correlation identifies the situated comparison operand. The authored
        # edge names the original supplying field, including proved renames.
        lineage = set()
        pending = list(endpoint["reference_ids"])
        while pending:
            reference_id = pending.pop()
            if reference_id in lineage:
                continue
            lineage.add(reference_id)
            pending.extend(references[reference_id]["origin_reference_ids"])
        return any(item["kind"] == "field" and item.get("input_id") == source["id"]
                   and item["field_identity"]["segments"] == [want["field"]]
                   and any(o["reference_id"] in lineage and endpoint["occurrence_id"] in o["input_occurrence_ids"]
                           for o in item["occurrences"]) for item in requirements["items"])
    for want in expected["correlation"]["edges"]:
        assert any(all(endpoint_supplies(edge[side], want[side]) for side in ("left", "right"))
                   for edge in report["correlation"]["edges"])


REQUEST_ERRORS = {
    "request-unsupported-request-version": ("request_invalid", "/schema_version"),
    "request-missing-required-input-bindings": ("request_invalid", "/input_bindings"),
    "request-missing-snapshot": ("request_invalid", "/snapshot"),
    "request-invalid-query-scope": ("request_invalid", "/query_scope/app"),
    "request-duplicate-input-binding": ("binding_invalid", "/input_bindings/1/input_id"),
    "request-unknown-input-binding": ("binding_invalid", "/input_bindings/0/input_id"),
    "request-binding-identity-mismatch": ("binding_invalid", "/input_bindings/0/expected"),
    "request-unknown-bound-schema": ("binding_invalid", "/input_bindings/0/schema_id"),
    "request-schema-belongs-to-different-object": ("binding_invalid", "/input_bindings/0/schema_id"),
    "request-stale-artifact-digest": ("snapshot_invalid", "/snapshot/digest"),
}


def assert_analysis_embeddings(analysis):
    for key in ("inputs", "input_coverage", "field_attribution_coverage", "correlation"):
        assert analysis[key] == analysis["requirements"][key]


def assert_discovery_parity(mapper, document, cli_path, server_url, reporter):
    analysis = call_python(mapper, "analyze_query", document)
    requirements = call_python(mapper, "requirements_query", document)
    assert analysis["requirements"] == requirements
    assert_analysis_embeddings(analysis)
    go = json.loads(go_value(reporter, json.dumps(document).encode(), "discovery"))
    for operation, cli_operation, key, want in (("analyze_query", "analyze", "analysis", analysis),
                                                ("requirements_query", "requirements", "requirements", requirements)):
        exit_code, cli, _ = run_cli(cli_path, cli_operation, document)
        status, http, _ = post_document(server_url, cli_operation, document)
        assert status == 200
        assert exit_code == ({"valid": 0, "invalid": 1, "incomplete": 3}[analysis["status"]]
                             if cli_operation == "analyze" else expected_requirements_exit(requirements))
        assert want == call_raw_c(mapper, operation, document) == cli == http == go[key]
    return analysis


@pytest.mark.parametrize("case", CASES, ids=lambda case: case["name"])
def test_compatibility_full_value_parity(case, cli_path, server_url, compatibility_go_reporter, compatibility_temp_dir):
    with open_mapper() as mapper:
        request = make_request(mapper, case)
    def authored(report, mapper):
        analysis = call_python(mapper, "analyze_query", case["document"])
        assert_authored_facts(case, report, analysis)
        assert_discovery_parity(mapper, case["document"], cli_path, server_url, compatibility_go_reporter)
    error = REQUEST_ERRORS[case["name"]] if case["expected"].get("request_error") else None
    exercise_request(request, authored, error, cli_path, server_url, compatibility_go_reporter, compatibility_temp_dir)


def exercise_request(request, authored, error, cli_path, server_url, compatibility_go_reporter, compatibility_temp_dir):
    raw = json.dumps(request).encode("utf-8")
    with open_mapper() as mapper:
        kind, c_value = native_raw(mapper, raw)
        if error:
            assert kind == "error"
            assert (c_value["code"], c_value["path"]) == error
            with pytest.raises(SPLMapperError) as raised:
                mapper.check_compatibility(request)
            python_value = json.loads(str(raised.value))
        else:
            assert kind == "report", c_value
            authored(c_value, mapper)
            python_value = mapper.check_compatibility(request)
    path = compatibility_temp_dir / "request.json"
    path.write_bytes(raw)
    completed = subprocess.run([str(cli_path), "compatibility", "--request", str(path), "--format", "json"],
                               capture_output=True, timeout=10)
    assert completed.returncode == (2 if kind == "error" else EXITS[c_value["outcome"]])
    if kind == "error":
        assert completed.stdout == b""
    else:
        assert completed.stderr == b""
    cli_bytes = completed.stderr if kind == "error" else completed.stdout
    status, http_value = post_raw(server_url, raw)
    assert status == (400 if kind == "error" else 200)
    assert json.loads(go_value(compatibility_go_reporter, raw)) == json.loads(cli_bytes) == http_value == c_value == python_value
    connected = subprocess.run([str(cli_path), "compatibility", "--request", str(path), "--format", "json", "--require-connected"],
                               capture_output=True, timeout=10)
    assert (connected.stdout, connected.stderr) == (completed.stdout, completed.stderr)
    exit_code = completed.returncode
    if kind == "report" and c_value["correlation"]["outcome"] == "disconnected":
        exit_code = 1
    elif kind == "report" and c_value["correlation"]["outcome"] == "indeterminate" and exit_code == 0:
        exit_code = 3
    assert connected.returncode == exit_code
    if kind == "report":
        output = compatibility_temp_dir / "report.txt"
        text = subprocess.run([str(cli_path), "compatibility", "--request", str(path), "--output", str(output)],
                              capture_output=True, timeout=10)
        assert (text.returncode, text.stdout, text.stderr) == (EXITS[c_value["outcome"]], b"", b"")
        assert output.read_bytes() == go_value(compatibility_go_reporter, raw, "text")
        assert json.loads(output.read_text(encoding="utf-8").split("Evidence:\n", 1)[1]) == c_value
    return c_value


@pytest.mark.parametrize("raw", [b"", b"{", b"\xff", b'{"schema_version":1,"unknown":true}'])
def test_compatibility_malformed_raw_parity(raw, cli_path, server_url, compatibility_go_reporter, compatibility_temp_dir):
    path = compatibility_temp_dir / "malformed.json"
    path.write_bytes(raw)
    completed = subprocess.run([str(cli_path), "compatibility", "--request", str(path), "--format", "json"],
                               capture_output=True, timeout=10)
    assert (completed.returncode, completed.stdout) == (2, b"")
    status, http_value = post_raw(server_url, raw)
    assert status == 400
    with open_mapper() as mapper:
        kind, c_value = native_raw(mapper, raw)
    assert kind == "error"
    assert json.loads(go_value(compatibility_go_reporter, raw)) == json.loads(completed.stderr) == http_value == c_value
    assert c_value["code"] == "request_invalid" and c_value["message"]


def test_compatibility_real_http_body_boundaries(server_url):
    with open_mapper() as mapper:
        raw = json.dumps(make_request(mapper, CASES[0])).encode("utf-8")
    assert post_raw(server_url, raw, "text/plain")[0] == 400
    exact = raw + b" " * ((8 << 20) - len(raw))
    assert post_raw(server_url, exact)[0] == 200
    status, error = post_raw(server_url, exact + b" ")
    assert status == 400 and error["error"] is True and "8 MiB" in error["message"]


# These controls extend evidence from the immutable corpus; queries and expected
# behavior are authored here, never constructed from observed reports.
def authored_case(name):
    return copy.deepcopy(next(case for case in CASES if case["name"] == name))


def request_for(mapper, text="from $events | fields id", language="spl2"):
    case = authored_case("schema-source-partial")
    case["document"].update(text=text, language=language, source_id="task12:" + text)
    case["schema_bundle"]["bindings"][0].pop("reason")
    case["schema_bundle"]["bindings"][0]["source_coverage"] = "complete"
    case["snapshot"]["capabilities"][0]["id"] = "language:" + language + ":profile:splunkd"
    return make_request(mapper, case), case["document"]


def assert_field_facts(report, expected):
    requirements = report["requirements"]
    owners = {source["id"]: source["name"] for source in requirements["inputs"]}
    fields = {(owners.get(item.get("input_id")), item["identity"], item["necessity"])
              for item in requirements["items"] if item["kind"] == "field"}
    assert fields == set(expected)
    for item in requirements["items"]:
        if item["kind"] == "field":
            assert item["ownership"]["state"] == "proved"


def requirement_outcome(report, kind, identity, owner=None):
    requirements = report["requirements"]
    owners = {source["id"]: source["name"] for source in requirements["inputs"]}
    ids = {item["id"] for item in requirements["items"] if item["kind"] == kind and item["identity"] == identity
           and (owner is None or owners.get(item.get("input_id")) == owner)}
    return next(out for out in report["requirement_outcomes"] if out["requirement_id"] in ids)


EVIDENCE_LIMITS = [
    ("partial-object-absence", "dataset-collection-partial", "environment_collection_partial"),
    ("outside-capture-absence", "query-scope-outside-capture", "environment_scope_not_covered"),
    ("covered-bound-absence-broad-query", "query-scope-outside-capture", "environment_object_missing"),
    ("partial-field-absence", "schema-source-partial", "schema_partial"),
    ("capability-unknown", "schema-source-partial", "environment_capability_unknown"),
    ("capability-omitted", "schema-source-partial", "environment_capability_not_supplied"),
    ("capability-other-version", "schema-source-partial", "environment_capability_version_unproven"),
    ("mixed-missing-and-unknown", "schema-not-supplied", "schema_not_supplied"),
]


@pytest.mark.parametrize("name,base,reason", EVIDENCE_LIMITS, ids=[item[0] for item in EVIDENCE_LIMITS])
def test_evidence_limits(name, base, reason, cli_path, server_url, compatibility_go_reporter, compatibility_temp_dir):
    case = authored_case(base)
    if name in ("partial-object-absence", "outside-capture-absence", "covered-bound-absence-broad-query", "mixed-missing-and-unknown"):
        case["snapshot"]["objects"] = []
    elif name == "partial-field-absence":
        case["schema_bundle"]["schemas"][0]["catalog"]["fields"] = ["other"]
    elif name == "capability-omitted":
        case["snapshot"]["capabilities"] = []
    elif name == "capability-unknown":
        case["snapshot"]["capabilities"][0]["state"] = "unknown"
    elif name == "capability-other-version":
        case["snapshot"]["capabilities"][0]["version"] = "other"
    if name == "outside-capture-absence":
        # The selected exact identity must itself lie outside the capture. A
        # broader query alone cannot invalidate covered bound-identity absence.
        case["bindings"]["named_placeholder:$events"]["expected"]["app"] = "other-app"
        case["schema_bundle"]["bindings"][0]["expected"]["app"] = "other-app"
    with open_mapper() as mapper:
        request = make_request(mapper, case)
    def authored(report, mapper):
        assert report["outcome"] == ("unsatisfied" if name in ("mixed-missing-and-unknown", "covered-bound-absence-broad-query") else "incomplete")
        assert reason in {r["code"] for r in report["reasons"]}
        assert_field_facts(report, [("$events", "id", "required")])
        field = requirement_outcome(report, "field", "id")
        assert field["outcome"] == ("indeterminate" if name in ("partial-field-absence", "mixed-missing-and-unknown") else "satisfied")
        if name == "partial-field-absence":
            assert field["field_projection"]["outcome"] == "missing"
        if name in ("partial-object-absence", "outside-capture-absence", "covered-bound-absence-broad-query", "mixed-missing-and-unknown"):
            want = "indeterminate" if name in ("partial-object-absence", "outside-capture-absence") else "missing"
            assert report["inputs"][0]["outcome"] == want
            assert requirement_outcome(report, "dataset", "$events")["outcome"] == want
        assert_discovery_parity(mapper, case["document"], cli_path, server_url, compatibility_go_reporter)
    exercise_request(request, authored, None, cli_path, server_url, compatibility_go_reporter, compatibility_temp_dir)


@pytest.mark.parametrize("present", [False, True], ids=["bounded-observed-absence", "bounded-observed-positive"])
def test_observation_meaning(present, cli_path, server_url, compatibility_go_reporter, compatibility_temp_dir):
    text = 'from {kind:"sourcetype",properties:{name:"audit"}}' if present else 'from {kind:"source",properties:{name:"absent"}}'
    with open_mapper() as mapper:
        request, document = request_for(mapper, text)
    request.pop("schema_bundle")
    request["input_bindings"] = []
    snapshot = request["snapshot"]
    provenance = snapshot["objects"][0]["provenance"]
    snapshot["schema_version"] = 2
    snapshot["objects"] = [{"id": "index-main", "kind": "index", "name": "main", "provenance": provenance},
                           {"id": "type-audit", "kind": "sourcetype", "name": "audit", "provenance": provenance}]
    snapshot["observation"] = {
        "index_selection": {"all": True},
        "enumeration": {"method": "distributed_rest", "peer_scope": "configured_search_peers", "coverage": "complete", "provenance": provenance},
        "indexes": [{"index_id": "index-main", "catalog_datatypes": ["event"], "required_datatypes": ["event"]}],
        "unmatched_indexes": [], "method": "splunk_metadata", "visibility": "exporting_principal",
        "window": {"mode": "all_retained"}, "time_precision": "bucket_overlap", "absence_meaning": "not_observed",
        "captures": [{"kind": "source", "index_id": "index-main", "datatype": "event", "object_ids": [], "coverage": "complete", "provenance": provenance},
                     {"kind": "sourcetype", "index_id": "index-main", "datatype": "event", "object_ids": ["type-audit"], "coverage": "complete", "provenance": provenance}]}
    def authored(report, mapper):
        assert report["outcome"] == ("satisfied" if present else "incomplete")
        assert report["observation"] == snapshot["observation"]
        assert report["inputs"][0]["outcome"] == ("satisfied" if present else "indeterminate")
        if not present:
            assert "observation_scope_insufficient" in {r["code"] for r in report["reasons"]}
        assert_field_facts(report, [])
        assert_discovery_parity(mapper, document, cli_path, server_url, compatibility_go_reporter)
    exercise_request(request, authored, None, cli_path, server_url, compatibility_go_reporter, compatibility_temp_dir)


@pytest.mark.parametrize("stale", [False, True], ids=["omitted-placeholder-binding", "stale-requirements"])
def test_additional_admission_errors(stale, cli_path, server_url, compatibility_go_reporter, compatibility_temp_dir):
    with open_mapper() as mapper:
        request, _ = request_for(mapper)
    if stale:
        request["requirements"]["capability_revision"] = "old"
        error = ("requirements_stale", "/requirements/capability_revision")
    else:
        request["input_bindings"] = []
        error = ("missing_input_binding", "/input_bindings")
    exercise_request(request, None, error, cli_path, server_url, compatibility_go_reporter, compatibility_temp_dir)


@pytest.mark.parametrize("bound", [False, True], ids=["ambiguous-literal-object", "explicit-literal-selection"])
def test_literal_object_selection(bound, cli_path, server_url, compatibility_go_reporter, compatibility_temp_dir):
    with open_mapper() as mapper:
        request, document = request_for(mapper, "from events")
    request.pop("schema_bundle")
    second = copy.deepcopy(request["snapshot"]["objects"][0])
    second.update(id="events-second", app="other")
    request["snapshot"]["objects"].append(second)
    request["query_scope"]["app"] = {"all": True}
    request["input_bindings"] = []
    if bound:
        original = request["snapshot"]["objects"][0]
        request["input_bindings"] = [{"input_id": request["requirements"]["inputs"][0]["id"], "object_id": original["id"],
                                      "expected": {key: original[key] for key in ("kind", "name", "namespace", "app", "owner")}}]
    def authored(report, mapper):
        assert report["outcome"] == ("satisfied" if bound else "incomplete")
        assert report["inputs"][0]["outcome"] == ("satisfied" if bound else "ambiguous")
        assert len(report["inputs"][0]["objects"]) == (1 if bound else 2)
        if not bound:
            assert "environment_object_ambiguous" in {r["code"] for r in report["reasons"]}
        assert_field_facts(report, [])
        assert_discovery_parity(mapper, document, cli_path, server_url, compatibility_go_reporter)
    exercise_request(request, authored, None, cli_path, server_url, compatibility_go_reporter, compatibility_temp_dir)


def test_unmapped_static_descriptor(cli_path, server_url, compatibility_go_reporter, compatibility_temp_dir):
    with open_mapper() as mapper:
        request, document = request_for(mapper, 'from {kind:"index",properties:{name:"main",enabled:true}}')
    request.pop("schema_bundle")
    request["input_bindings"] = []
    def authored(report, mapper):
        assert report["outcome"] == "incomplete"
        assert {"dataset_identity_unmapped", "target_discovery_incomplete"} <= {r["code"] for r in report["reasons"]}
        assert_field_facts(report, [])
        assert_discovery_parity(mapper, document, cli_path, server_url, compatibility_go_reporter)
    exercise_request(request, authored, None, cli_path, server_url, compatibility_go_reporter, compatibility_temp_dir)


def test_conditional_missing_preserves_fact(cli_path, server_url, compatibility_go_reporter, compatibility_temp_dir):
    case = authored_case("select-first-left-joins")
    case["schema_bundle"]["schemas"][1]["catalog"]["fields"] = ["other"]
    with open_mapper() as mapper:
        request = make_request(mapper, case)
    def authored(report, mapper):
        assert report["outcome"] == "incomplete"
        assert "conditional_applicability_unproven" in {r["code"] for r in report["reasons"]}
        users = requirement_outcome(report, "field", "id", "$users")
        assert (users["outcome"], users["applicability"]) == ("missing", "indeterminate")
        assert_field_facts(report, [("$events", "user_id", "required"), ("$events", "asset_id", "required"),
                                   ("$users", "id", "conditional"), ("$assets", "id", "conditional")])
        assert_discovery_parity(mapper, case["document"], cli_path, server_url, compatibility_go_reporter)
    exercise_request(request, authored, None, cli_path, server_url, compatibility_go_reporter, compatibility_temp_dir)


def definition_request(mapper):
    hidden, body = request_for(mapper)
    request, document = request_for(mapper, "from view | fields id")
    request["document"] = document
    view = copy.deepcopy(request["snapshot"]["objects"][0])
    view.update(id="view", name="view", document=dict(body, source_id="shared.spl"))
    request["snapshot"]["objects"].append(view)
    expected = {key: view[key] for key in ("kind", "name", "namespace", "app", "owner")}
    request["input_bindings"] = hidden["input_bindings"] + [{"input_id": request["requirements"]["inputs"][0]["id"],
        "object_id": "view", "expected": expected, "schema_id": "schema-events"}]
    request["schema_bundle"]["bindings"].append({"schema_id": "schema-events", "object_id": "view", "expected": expected,
                                               "source_coverage": "complete"})
    return request, document


def macro_request(mapper, repeated=False):
    hidden, body = request_for(mapper)
    text = "`source(1)` | `source(2)`" if repeated else "`source`"
    request, document = request_for(mapper, text, "spl")
    request["document"] = document
    request["snapshot"]["capabilities"].append(copy.deepcopy(hidden["snapshot"]["capabilities"][0]))
    base = copy.deepcopy(request["snapshot"]["objects"][0])
    macro = dict(base, id="macro", kind="macro", name="source", arity=1 if repeated else 0,
                 eval_based=False, arguments=["events"] if repeated else [],
                 document={"text": "eval local=$events$" if repeated else "| makeresults", "language": "spl", "source_id": "macro.spl"},
                 relations=[{"kind": "saved_search", "name": "Daily", "property": "/dependency"}])
    daily = dict(base, id="daily", kind="saved_search", name="Daily", document=dict(body, source_id="daily.spl"))
    request["snapshot"]["objects"].extend([macro, daily])
    request["input_bindings"] = hidden["input_bindings"]
    return request, document


@pytest.mark.parametrize("name", ["effective-definition", "missing-hidden-binding", "missing-direct-binding",
                                    "opaque-macro", "definition-cycle", "repeated-formal", "missing-repeated-hidden-binding"])
def test_closure_semantics(name, cli_path, server_url, compatibility_go_reporter, compatibility_temp_dir):
    with open_mapper() as mapper:
        request, document = (macro_request(mapper, "repeated" in name)
                             if "macro" in name or "repeated" in name else definition_request(mapper))
    error = None
    if name in ("missing-hidden-binding", "missing-repeated-hidden-binding"):
        request["input_bindings"] = [b for b in request["input_bindings"] if b["object_id"] == "view"]
        error = ("missing_input_binding", "/input_bindings")
    elif name == "missing-direct-binding":
        request["input_bindings"] = [b for b in request["input_bindings"] if b["object_id"] != "view"]
        # A named direct source must be admitted before the hidden source.
        document["text"] = "from $direct | union [from view]"
        request["document"] = document
        with open_mapper() as mapper:
            request["requirements"] = call_python(mapper, "requirements_query", document)
        error = ("missing_input_binding", "/input_bindings")
    elif name == "opaque-macro":
        request["snapshot"]["objects"][1].pop("document")
        request["snapshot"]["objects"][1].pop("relations")
        request["input_bindings"] = []
    elif name == "definition-cycle":
        request["snapshot"]["objects"][1]["document"]["text"] = "from view | fields id"
        request["input_bindings"] = []
    def authored(report, mapper):
        codes = {r["code"] for r in report["reasons"]}
        assert report["requirements"] == request["requirements"]
        closure = report["closure"]
        assert closure["effective_analysis"]["document"]["text"] == (
            "eval local=1 | eval local=2" if name == "repeated-formal" else document["text"])
        if name in ("opaque-macro", "definition-cycle"):
            assert report["outcome"] == "incomplete"
            assert "dependency_closure_incomplete" in codes
            assert closure["gaps"]
            assert len(report["inputs"]) == 1
            if name == "opaque-macro":
                # Opaque invocation retains the root implicit stream and cannot
                # invent a hidden Dataset from a body that was never supplied.
                assert all(not o.get("definition_object_id") for source in report["inputs"] for o in source["occurrences"])
            assert not any(source.get("name") == "$events" for source in report["effective_requirements"]["inputs"])
            if name == "definition-cycle":
                assert any(edge["cycle_path"] == ["view", "view"] for edge in closure["traversal"])
        elif name == "effective-definition":
            assert report["outcome"] == "satisfied"
            assert len(report["inputs"]) == 2
            assert_field_facts(report, [("view", "id", "required")])
            definition = next(d for d in closure["definition_analyses"] if d["object_id"] == "view")
            assert definition["effective_analysis"]["document"]["text"] == "from $events | fields id"
            hidden = [out for out in report["requirement_outcomes"] if out.get("definition_object_id") == "view" and "field_projection" in out]
            assert len(hidden) == 1
            assert hidden[0]["outcome"] == "satisfied"
            assert [(i["object_id"], i["start"], i["end"]) for i in hidden[0]["source_intervals"]] == [("view", 22, 24)]
            assert len(hidden[0]["invocation_provenance"]) == 1
        else:
            # Expanding eval preserves an unresolved implicit root stream;
            # the separate real Dataset source remains proved and bound.
            assert report["outcome"] == "incomplete"
            assert "target_discovery_incomplete" in codes
            assert "dependency_closure_incomplete" not in codes
            assert len(report["inputs"]) == 2
            hidden = [source for source in report["inputs"] if source["input_id"] == request["input_bindings"][0]["input_id"]]
            assert len(hidden) == 1 and hidden[0]["outcome"] == "satisfied"
            assert len([source for source in report["inputs"] if source["outcome"] == "indeterminate"]) == 1
            occurrences = hidden[0]["occurrences"]
            assert len({o["traversal_edge_id"] for o in occurrences}) == 2
            for occurrence in occurrences:
                assert occurrence["definition_object_id"] == "daily"
                frames = occurrence["invocation_provenance"]
                assert [frame["object_id"] for frame in frames] == ["macro", "daily"]
                assert [len(frame["invocation"]) for frame in frames] == [1, 0]
            assert_field_facts(report, [])
        assert closure["direct_requirements"] == request["requirements"]
        assert closure["effective_analysis"]["requirements"] == report["effective_requirements"]
        assert_analysis_embeddings(closure["effective_analysis"])
        for definition in closure["definition_analyses"]:
            assert_analysis_embeddings(definition["direct_analysis"])
            if definition.get("effective_analysis") is not None:
                assert_analysis_embeddings(definition["effective_analysis"])
        assert_discovery_parity(mapper, document, cli_path, server_url, compatibility_go_reporter)
    exercise_request(request, authored, error, cli_path, server_url, compatibility_go_reporter, compatibility_temp_dir)


# Literal queries protect transfer and join semantics without generating variants.
TRANSFER_QUERIES = [
    ("aggregate-derived", "from $events | eval copied=id | rename copied AS key | stats count(key) AS total | fields total", "satisfied", [("$events", "id", "required")]),
    ("removed-read", "from $events | join type=inner left=e right=u where e.id=u.user_id [from $users] | fields - id | where id=1", "not assessed", [("$events", "id", "required"), ("$users", "user_id", "required")]),
    ("aggregate-closed-read", "from $events | join type=inner left=e right=u where e.id=u.user_id [from $users] | stats count() AS total | where missing=1", "not assessed", [("$events", "id", "required"), ("$users", "user_id", "required")]),
    ("bounded-join", "from $events | join max=2 left=e right=u where e.id=u.id [from $users]", "satisfied", [("$events", "id", "required"), ("$users", "id", "required")]),
    ("outer-join", "from $events | join type=outer left=e right=u where e.id=u.id [from $users]", "satisfied", [("$events", "id", "required"), ("$users", "id", "required")]),
]


@pytest.mark.parametrize("name,text,outcome,fields", TRANSFER_QUERIES, ids=[item[0] for item in TRANSFER_QUERIES])
def test_transfers_and_join_options(name, text, outcome, fields, cli_path, server_url, compatibility_go_reporter, compatibility_temp_dir):
    case = authored_case("pipeline-three-input-renamed-key")
    case["document"].update(text=text, source_id="task12:" + name)
    case["schema_bundle"]["schemas"][0]["catalog"]["fields"] = ["id"]
    case["schema_bundle"]["schemas"][1]["catalog"]["fields"] = ["id", "user_id"]
    with open_mapper() as mapper:
        request = make_request(mapper, case)
    def authored(report, mapper):
        assert report["outcome"] == outcome
        assert_field_facts(report, fields)
        assert not any(item["identity"] in ("copied", "key", "total", "missing")
                       for item in report["requirements"]["items"] if item["kind"] == "field")
        if name != "aggregate-derived":
            assert report["correlation"]["outcome"] == "connected"
            assert len(report["correlation"]["edges"]) == 1
            edge = report["correlation"]["edges"][0]
            right_field = "user_id" if outcome == "not assessed" else "id"
            owners = {source["name"]: source["id"] for source in report["requirements"]["inputs"]}
            for side, owner, field in (("left", "$events", "id"), ("right", "$users", right_field)):
                assert edge[side]["input_id"] == owners[owner]
                assert any(item["kind"] == "field" and item["identity"] == field and item.get("input_id") == owners[owner]
                           and any(o["reference_id"] in edge[side]["reference_ids"] for o in item["occurrences"])
                           for item in report["requirements"]["items"])
        if outcome == "not assessed":
            analysis = call_python(mapper, "analyze_query", case["document"])
            last = analysis["references"][-1]
            assert last["binding"] == "unavailable"
            assert not any(o["reference_id"] == last["id"] for item in report["requirements"]["items"] for o in item["occurrences"])
        assert_discovery_parity(mapper, case["document"], cli_path, server_url, compatibility_go_reporter)
    exercise_request(request, authored, None, cli_path, server_url, compatibility_go_reporter, compatibility_temp_dir)


@pytest.mark.parametrize("closure", [False, True], ids=["independent-bindings", "effective-definition"])
def test_prepared_detach_reuse_and_concurrency(closure, cli_path, server_url, compatibility_go_reporter, compatibility_temp_dir):
    with open_mapper() as mapper:
        first, document = definition_request(mapper) if closure else request_for(mapper)
    # Both requests use the same prepared artifacts and independently select a
    # compatible object; the second cannot alter the first report's binding.
    alternate = copy.deepcopy(first["snapshot"]["objects"][0])
    alternate.update(id="alternate", name="alternate")
    first["snapshot"]["objects"].append(alternate)
    expected = {key: alternate[key] for key in ("kind", "name", "namespace", "app", "owner")}
    first["schema_bundle"]["bindings"].append({"schema_id": "schema-events", "object_id": "alternate",
                                             "expected": expected, "source_coverage": "complete"})
    second = copy.deepcopy(first)
    second["input_bindings"][0].update(object_id="alternate", expected=expected)
    def authored(report, mapper):
        assert report["outcome"] == "satisfied"
        assert {b["object_id"] for b in report["input_bindings"]} <= {"dataset-events", "alternate", "view"}
        assert_field_facts(report, [("view" if closure else "$events", "id", "required")])
        assert_discovery_parity(mapper, document, cli_path, server_url, compatibility_go_reporter)
    canonical = [exercise_request(request, authored, None, cli_path, server_url, compatibility_go_reporter, compatibility_temp_dir)
                 for request in (first, second)]
    assert canonical[0]["input_bindings"][0]["object_id"] == "dataset-events"
    assert canonical[1]["input_bindings"][0]["object_id"] == "alternate"
    prepared = json.loads(go_value(compatibility_go_reporter, json.dumps([first, second]).encode(), "prepared"))
    assert prepared["reports"] == canonical
    assert prepared["concurrent"] == [canonical[i % 2] for i in range(8)]


def test_invalid_root_retains_hidden_missing_fact(cli_path, server_url, compatibility_go_reporter, compatibility_temp_dir):
    with open_mapper() as mapper:
        request, document = definition_request(mapper)
        document["text"] = "from view | where"
        request["requirements"] = call_python(mapper, "requirements_query", document)
    request["snapshot"]["objects"][1]["document"]["text"] = "from $events | fields missing"
    def authored(report, mapper):
        assert report["outcome"] == "not assessed"
        assert report["requirements"]["query_status"] == "invalid"
        assert {"unsupported_semantics", "schema_field_missing"} <= {r["code"] for r in report["reasons"]}
        hidden = [out for out in report["requirement_outcomes"] if out.get("definition_object_id") == "view" and "field_projection" in out]
        assert len(hidden) == 1 and hidden[0]["outcome"] == "missing"
        assert [(i["object_id"], i["start"], i["end"]) for i in hidden[0]["source_intervals"]] == [("view", 22, 29)]
        assert len(hidden[0]["invocation_provenance"]) == 1
        assert_discovery_parity(mapper, document, cli_path, server_url, compatibility_go_reporter)
    exercise_request(request, authored, None, cli_path, server_url, compatibility_go_reporter, compatibility_temp_dir)
