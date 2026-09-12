"""Independent rewrite obligations plus complete Go/CLI/HTTP/native parity.

This module also prepares the Go transport artifact before package installation.
The artifact grants no semantic credit: every consumer checks authored expectations.
"""
from copy import deepcopy
import gzip
import hashlib
import json
import os
from pathlib import Path
import subprocess
from urllib.error import HTTPError
from urllib.request import Request, urlopen


REQUIRED_GROUPS = {"preview", "apply", "aliases", "implicit", "conditions", "boolean", "contains",
                   "fact-order", "scopes", "sql", "identity-kinds", "atom-path", "collisions",
                   "chains-swaps", "dependent-skips", "bytes", "syntax", "validation", "no-op",
                   "refusals", "lexical-controls", "batch"}
REQUIRED_GROUPS.update({
    "all-false-unknown", "all-true-unknown", "and-compatible-repeat", "and-conflict",
    "and-conflict-contains", "any-false-unknown", "any-true-unknown", "atom-path-control",
    "chain", "coalescing", "common-or", "common-or-spl",
    "conditional-spl", "conditional-spl2", "conflict-independent", "conflict-or",
    "conflict-unknown", "contains-pattern-unknown", "contains-query-text-control", "dataset-component",
    "dependency-context-control", "dependency-index", "dependency-source", "dependency-sourcetype",
    "dependency-value-control", "dependent-collision", "empty-rules", "eval-alias",
    "field-target-invalid", "field-target-valid", "implicit-consumer", "implicit-refusal",
    "independent-child", "inherited-child", "json-schema-conditional", "json-schema-invalid",
    "json-schema-unresolved", "json-schema-valid", "later-fact", "literal-comment-control",
    "literal-target", "lookup-catalog-spl2", "macro-refusal", "metric-index",
    "navigation-refusal", "no-change", "no-match", "noncommon-or",
    "noncommon-or-spl", "not", "not-spl", "null-spl2",
    "ocsf-local", "overwritten-fact", "qualified-corender", "reference-present",
    "removed-fact", "rename-spl2", "sql-logical-order", "swap",
    "typed-boolean", "typed-contains-exact-star", "typed-null", "typed-number-string",
    "unicode-crlf", "wildcard-refusal",
    "having-compatible",
    "having-conflict",
    "having-conflict-unknown",
    "having-derived",
    "having-hidden",
    "having-only",
    "having-phase-order",
    "having-ungrouped",
    "having-unsupported",
    "having-where-control",
    "number-decimal-spl",
    "number-exponent-and",
    "number-exponent-different",
    "number-exponent-int64",
    "number-exponent-million",
    "number-exponent-negative",
    "number-exponent-or",
    "number-exponent-positive",
})
REPORT_KEYS = {"schema_version", "document", "mode", "status", "coverage", "original_text", "candidate_text",
               "text", "committed", "changes", "rule_evaluations", "original_analysis", "candidate_analysis"}


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def source_hashes(root):
    paths = [p for directory in ("pkg", "internal", "parser") for p in (root / directory).rglob("*.go")]
    paths += list((root / "grammar").glob("*.g4")) + [root / "go.mod", root / "go.sum"]
    return {p.relative_to(root).as_posix(): digest(p) for p in sorted(paths)}


def load_cases(fixtures):
    cases = json.loads((fixtures / "corpus.json").read_text(encoding="utf-8"))
    ids = [c["id"] for c in cases]
    assert ids and len(ids) == len(set(ids)), "empty corpus or duplicate rewrite IDs"
    assert REQUIRED_GROUPS <= {g for c in cases for g in c["groups"]}, "empty required rewrite group"
    for case in cases:
        assert case["request"]["document"]["source_id"] == case["id"]
        assert set(case["request"]["document"]) == {"text", "language", "profile", "version", "source_id"}
        assert case["target"] == case["request"].get("validation_target")
        assert not case["target"] or "validation_summary" in case["expected"]
        if "catalog_fixture" in case:
            assert case["catalog_fixture"] == "ocsf/1.6.0/base.json.gz", "unknown rewrite catalog fixture"
            schemas = Path(os.environ.get("SPL_SCHEMA_FIXTURES", fixtures.parent / "schemas"))
            raw = gzip.decompress((schemas / case["catalog_fixture"]).read_bytes())
            assert hashlib.sha256(raw).hexdigest() == case["catalog_raw_sha256"] == "9b609f8fb670772f04191c1c276b46d34d6e9110d2417c71fa89c4f54c585137"
            case["target"]["catalog"] = json.loads(raw)
            case["request"]["validation_target"] = deepcopy(case["target"])
    return cases


def projection(report):
    result = {k: deepcopy(report[k]) for k in ("status", "candidate_text", "text", "committed", "coverage", "changes", "rule_evaluations")}
    if wrapper := report.get("candidate_validation"):
        validation = wrapper["field_list" if wrapper["kind"] == "field_list" else "schema"]
        result["validation_summary"] = [wrapper["kind"], validation["status"],
                                        [[item["reference_id"], item["outcome"]] for item in validation["outcomes"]]]
    for key in ("changes", "rule_evaluations"):
        for entry in result[key]:
            for name in ("location", "original_location", "candidate_location"):
                if name in entry:
                    entry[name] = [entry[name]["start"]["offset"], entry[name]["end"]["offset"]]
    for prefix in ("original", "candidate"):
        analysis = report[prefix + "_analysis"]
        result[prefix + "_references"] = [[r[k] for k in ("id", "normalized_name", "kind", "role", "binding", "origin_reference_ids")] for r in analysis["references"]]
        result[prefix + "_lineage"] = [[l["stage_id"], l["scope_id"], l.get("phase", ""),
                                       [[f["name"], f["conditional"]] for f in l["after"]["fields"]], l["transitions"]]
                                      for l in analysis["lineage"]]
    return result


def load_transport(path, fixtures, expected_hash, root=None):
    assert digest(path) == expected_hash, "rewrite Go transport hash mismatch"
    artifact = json.loads(path.read_text(encoding="utf-8"))
    assert (artifact["schema_version"], artifact["kind"], artifact["conformance_credit"]) == (1, "rewrite-go-transport", 0)
    assert artifact["source_hashes"], "missing producing rewrite sources"
    if root is not None:
        assert artifact["source_hashes"] == source_hashes(root), "stale producing rewrite sources"
    assert artifact["fixture_hashes"] == {p.name: digest(p) for p in fixtures.glob("*.json")}, "stale rewrite fixtures"
    cases = {c["id"]: c for c in load_cases(fixtures)}
    reports = {}
    for entry in artifact["reports"]:
        identity = entry["id"]
        assert identity in cases and identity not in reports, "extra or duplicate Go rewrite ID"
        case, report = cases[identity], entry["report"]
        assert entry["request"] == case["request"]
        assert set(report) == REPORT_KEYS | ({"candidate_validation"} if case["target"] else set()), "truncated Go rewrite report"
        assert report["document"] == case["request"]["document"]
        assert projection(report) == case["expected"], identity
        reports[identity] = report
    assert set(reports) == set(cases), "missing Go rewrite reports"
    return artifact, reports


def prepare_transport(root, output):
    env = os.environ.copy() | {"SPL_REWRITE_GO_REPORTS": str(output.resolve())}
    subprocess.run(["go", "test", "-mod=readonly", "./pkg/rewrite", "-run", "^TestRewriteConformance$", "-count=1"], cwd=root, env=env, check=True)
    load_transport(output, root / "testdata/rewrite", digest(output), root)
    return output


if __name__ == "__main__":
    import argparse
    parser = argparse.ArgumentParser()
    parser.add_argument("--root", required=True, type=Path)
    parser.add_argument("--output", required=True, type=Path)
    arguments = parser.parse_args()
    prepare_transport(arguments.root.resolve(), arguments.output.resolve())
    raise SystemExit(0)


import pytest
from spl_toolkit import SPLMapper, SPLMapperError
from test_surfaces import cli_path, server_url, post_json, required_absolute_path
from test_schema_surfaces import cli_arguments


FIXTURES = required_absolute_path("SPL_REWRITE_FIXTURES")
CASES = load_cases(FIXTURES)
EXITS = {"valid": 0, "invalid": 1, "incomplete": 3}


def open_mapper():
    options = {"library_path": str(required_absolute_path("SPL_NATIVE_LIBRARY"))} if "SPL_NATIVE_LIBRARY" in os.environ else {}
    return SPLMapper(**options)


def native_request(mapper, request):
    options = {k: v for k, v in request.items() if k not in ("schema_version", "document")}
    if "documents" in request:
        return mapper.rewrite_batch(**options)
    document = request["document"]
    return mapper.rewrite(document["text"], **{k: v for k, v in document.items() if k != "text"}, **options)


def run_cli(cli, request, directory):
    rules = directory / "rules.json"
    rules.write_text(json.dumps({"schema_version": 1, "rules": request["rules"]}), encoding="utf-8")
    args = [str(cli), "rewrite", "--rules", str(rules), "--format", "json"]
    if request["mode"] == "apply":
        args += ["--apply"]
    if target := request.get("validation_target"):
        if target["kind"] == "field_list":
            fields = directory / "fields.json"
            fields.write_text(json.dumps(target["catalog"]), encoding="utf-8")
            args += ["--fields", str(fields)]
        else:
            args += cli_arguments(cli, target, directory)[4:]
    if "documents" in request:
        args += ["--batch", "-"]
        data = json.dumps(request["documents"]).encode()
    else:
        document = request["document"]
        for key, flag in (("language", "--language"), ("profile", "--profile"), ("version", "--compatibility-version"), ("source_id", "--source-id")):
            args += [flag, document[key]]
        args += ["--stdin"]
        data = document["text"].encode("utf-8")
    result = subprocess.run(args, input=data, capture_output=True, timeout=20)
    assert result.stdout, (result.returncode, result.stderr)
    report = json.loads(result.stdout)
    assert (result.returncode, result.stderr) == (EXITS[report["status"]], b"")
    return report


def assert_locations(report):
    # Independently count Unicode columns and CRLF lines from the actual bytes.
    def position(text, offset):
        prefix = text.encode("utf-8")[:offset].decode("utf-8")
        normalized = prefix.replace("\r\n", "\n").replace("\r", "\n")
        return {"offset": offset, "line": normalized.count("\n") + 1, "column": len(normalized.rsplit("\n", 1)[-1]) + 1}
    for prefix in ("original", "candidate"):
        analysis = report[prefix + "_analysis"]
        text = analysis["document"]["text"]
        assert text == report[prefix + "_text"]
        for reference in analysis["references"]:
            location = reference["location"]
            assert text.encode()[location["start"]["offset"]:location["end"]["offset"]].decode() == reference["original_name"]
            for endpoint in ("start", "end"):
                assert location[endpoint] == position(text, location[endpoint]["offset"])
    a_end = b_end = 0
    for change in report["changes"]:
        assert change["committed"] == (report["committed"] and change["candidate_applied"])
        if not change["candidate_applied"]:
            continue
        a, b = change["original_location"], change["candidate_location"]
        original, candidate = report["original_text"].encode(), report["candidate_text"].encode()
        assert original[a_end:a["start"]["offset"]] == candidate[b_end:b["start"]["offset"]]
        for location, key, expected in ((a, "original_text", change["old_text"]), (b, "candidate_text", change["new_text"])):
            text = report[key]
            assert text.encode()[location["start"]["offset"]:location["end"]["offset"]].decode() == expected
            for endpoint in ("start", "end"):
                assert location[endpoint] == position(text, location[endpoint]["offset"])
        a_end, b_end = a["end"]["offset"], b["end"]["offset"]
    assert report["original_text"].encode()[a_end:] == report["candidate_text"].encode()[b_end:]


@pytest.fixture(scope="session")
def rewrite_go_reports(tmp_path_factory):
    root = Path(__file__).resolve().parents[2]
    source = root if (root / "go.mod").is_file() else None
    if "SPL_REWRITE_GO_REPORTS" in os.environ:
        path = required_absolute_path("SPL_REWRITE_GO_REPORTS")
        expected = os.environ["SPL_REWRITE_GO_SHA256"]
    else:
        assert source is not None, "installed rewrite acceptance requires copied Go evidence"
        path = prepare_transport(root, tmp_path_factory.mktemp("rewrite-go") / "reports.json")
        expected = digest(path)
    artifact, reports = load_transport(path, FIXTURES, expected, source)
    return {"sha256": expected, "source_hashes": artifact["source_hashes"], "reports": reports, "batches": artifact["batches"]}


@pytest.fixture(scope="session")
def rewrite_evidence(cli_path, rewrite_go_reports):
    with open_mapper() as mapper:
        library = Path(mapper._lib._name).resolve()
    record = {"schema_version": 1, "fixture_hashes": {p.name: digest(p) for p in FIXTURES.glob("*.json")},
              "artifacts": {str(p): digest(p) for p in (cli_path, required_absolute_path("SPL_SERVER"), library)},
              "go_transport": {k: v for k, v in rewrite_go_reports.items() if k not in ("reports", "batches")},
              "cases": [], "batches": [], "request_errors": [], "body_limits": [],
              "catalogs": [{"fixture": c["catalog_fixture"], "raw_sha256": c["catalog_raw_sha256"]}
                           for c in CASES if "catalog_fixture" in c]}
    yield record
    if destination := os.environ.get("SPL_REWRITE_EVIDENCE"):
        Path(destination).write_text(json.dumps(record, sort_keys=True, ensure_ascii=False) + "\n", encoding="utf-8")


@pytest.mark.parametrize("case", CASES, ids=lambda c: c["id"])
def test_rewrite_full_canonical_parity(case, cli_path, server_url, tmp_path, rewrite_go_reports, rewrite_evidence):
    request = deepcopy(case["request"])
    with open_mapper() as mapper:
        native = native_request(mapper, request)
    assert projection(native) == case["expected"], case["id"]
    assert_locations(native)
    status, http = post_json(server_url, "/query/rewrite", request)
    cli = run_cli(cli_path, request, tmp_path)
    assert status == 200 and cli == http == native == rewrite_go_reports["reports"][case["id"]]
    rewrite_evidence["cases"].append({"id": case["id"], "request": request, "report": native})


def test_rewrite_ordered_batches(cli_path, server_url, tmp_path, rewrite_go_reports, rewrite_evidence):
    groups = {}
    for case in CASES:
        shared = {k: v for k, v in case["request"].items() if k != "document"}
        groups.setdefault(json.dumps(shared, sort_keys=True), []).append(case)
    assert any(len(group) > 1 for group in groups.values())
    go_batches = {tuple(b["ids"]): b["report"] for b in rewrite_go_reports["batches"]}
    assert len(go_batches) == len(groups) == len(rewrite_go_reports["batches"])
    assert any({rewrite_go_reports["reports"][c["id"]]["status"] for c in group} == {"valid", "incomplete", "invalid"} for group in groups.values())
    for shared, group in groups.items():
        ordered = list(reversed(group))
        reports = [rewrite_go_reports["reports"][c["id"]] for c in ordered]
        status = "invalid" if any(r["status"] == "invalid" for r in reports) else "incomplete" if any(r["status"] == "incomplete" for r in reports) else "valid"
        request = json.loads(shared) | {"documents": [c["request"]["document"] for c in ordered]}
        expected = {"schema_version": 1, "status": status, "reports": reports}
        with open_mapper() as mapper:
            native = native_request(mapper, request)
        code, http = post_json(server_url, "/query/rewrite/batch", request)
        assert code == 200 and native == http == run_cli(cli_path, request, tmp_path) == expected == go_batches[tuple(c["id"] for c in ordered)]
        rewrite_evidence["batches"].append({"ids": [c["id"] for c in ordered], "report": native})


@pytest.mark.parametrize("mode", ["preview", "apply"])
def test_rewrite_required_documented_example(mode, cli_path, rewrite_go_reports):
    case = next(c for c in CASES if c["id"] == "spl-alias-" + mode)
    rules_file = FIXTURES / "example-rules.json"
    assert json.loads(rules_file.read_text()) == {"schema_version": 1, "rules": case["request"]["rules"]}
    args = [str(cli_path), "rewrite", "--rules", str(rules_file), "--query", case["request"]["document"]["text"], "--format", "json"]
    if mode == "apply":
        args.append("--apply")
    result = subprocess.run(args, capture_output=True, text=True, timeout=20)
    assert result.returncode == 0 and not result.stderr
    expected = deepcopy(rewrite_go_reports["reports"][case["id"]])
    for document in (expected["document"], expected["original_analysis"]["document"], expected["candidate_analysis"]["document"]):
        document["source_id"] = ""
    assert json.loads(result.stdout) == expected


@pytest.mark.parametrize("documents", [[], [{"text": "search src=x"}, {"text": None}],
                                       [{"text": "search src=x"}, {"text": "search src=y", "language": "sql"}]])
def test_rewrite_batch_input_errors_are_atomic(documents, cli_path, server_url, tmp_path, rewrite_evidence):
    rules = [{"id": "map", "kind": "field", "source": {"name": "src"}, "target": {"name": "user"}}]
    request = {"schema_version": 1, "mode": "apply", "documents": documents, "rules": rules}
    code, report = post_json(server_url, "/query/rewrite/batch", request)
    assert code == 400 and "reports" not in report
    with open_mapper() as mapper:
        with pytest.raises(SPLMapperError):
            native_request(mapper, request)
    path = tmp_path / "rules.json"
    path.write_text(json.dumps({"schema_version": 1, "rules": rules}))
    cli = subprocess.run([str(cli_path), "rewrite", "--rules", str(path), "--apply", "--batch", "-", "--format", "json"], input=json.dumps(documents), capture_output=True, text=True, timeout=10)
    assert cli.returncode == 2 and cli.stdout == "" and set(json.loads(cli.stderr)) == {"error"}
    rewrite_evidence["request_errors"].append({"request": request, "http_status": code, "cli_exit": cli.returncode})


@pytest.mark.parametrize("batch", [False, True])
@pytest.mark.parametrize("chunked", [False, True])
@pytest.mark.parametrize("overflow", [False, True])
def test_rewrite_real_http_body_limit(batch, chunked, overflow, server_url, rewrite_go_reports, rewrite_evidence):
    # Whitespace padding changes only transport size. The accepted exact-limit
    # request must retain the complete canonical single/batch result.
    case = next(c for c in CASES if c["id"] == "spl-search-apply")
    request = deepcopy(case["request"])
    expected = rewrite_go_reports["reports"][case["id"]]
    if batch:
        request["documents"] = [request.pop("document")]
        expected = {"schema_version": 1, "status": "valid", "reports": [expected]}
    raw = json.dumps(request).encode()
    raw += b" " * ((8 << 20) + int(overflow) - len(raw))
    url = server_url + "/query/rewrite" + ("/batch" if batch else "")
    data = iter([raw]) if chunked else raw
    http_request = Request(url, data=data, headers={"Content-Type": "application/json"}, method="POST")
    if overflow:
        with pytest.raises(HTTPError) as caught:
            urlopen(http_request, timeout=15)
        with caught.value as response:
            status, actual = response.code, json.load(response)
        assert status == 400 and "reports" not in actual and "candidate_text" not in actual
        assert "8 MiB" in json.dumps(actual)
    else:
        with urlopen(http_request, timeout=15) as response:
            status, actual = response.status, json.load(response)
        assert status == 200 and actual == expected
    rewrite_evidence["body_limits"].append({"batch": batch, "chunked": chunked, "bytes": len(raw),
                                            "sha256": hashlib.sha256(raw).hexdigest(), "http_status": status})


def test_python_rewrite_example_uses_real_native(monkeypatch, capsys):
    import importlib.util
    path = required_absolute_path("SPL_DOCS_ROOT") / "python/examples/basic_usage.py"
    spec = importlib.util.spec_from_file_location("rewrite_basic_usage", path)
    example = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(example)
    if "SPL_NATIVE_LIBRARY" in os.environ:
        monkeypatch.setattr(example, "SPLMapper", lambda **kwargs: SPLMapper(library_path=str(required_absolute_path("SPL_NATIVE_LIBRARY")), **kwargs))
    assert example.main() == 0
    output = capsys.readouterr().out
    assert "Rewrite preview committed: False; apply committed: True" in output
    assert "Rewrite ordered batch: valid, incomplete, invalid" in output
