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
    "github.com/delgado-jacob/spl-toolkit/pkg/compatibility"
)
func main() {
    raw, err := io.ReadAll(os.Stdin); if err != nil { panic(err) }
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


@pytest.mark.parametrize("case", CASES, ids=lambda case: case["name"])
def test_compatibility_full_value_parity(case, cli_path, server_url, compatibility_go_reporter, compatibility_temp_dir):
    with open_mapper() as mapper:
        request = make_request(mapper, case)
        raw = json.dumps(request).encode("utf-8")
        kind, c_value = native_raw(mapper, raw)
        if case["expected"].get("request_error"):
            assert kind == "error"
            assert (c_value["code"], c_value["path"]) == REQUEST_ERRORS[case["name"]]
            with pytest.raises(SPLMapperError) as raised:
                mapper.check_compatibility(request)
            python_value = json.loads(str(raised.value))
        else:
            assert kind == "report"
            document = case["document"]
            analysis = mapper.analyze_query(document["text"], **{key: document[key] for key in
                                            ("language", "profile", "version", "source_id")})
            assert_authored_facts(case, c_value, analysis)
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
