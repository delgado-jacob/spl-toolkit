#!/usr/bin/env python3
"""Audit repository-owned SPL2 evidence. This is not a query/semantic parser.

Pending obligations are visible intermediate work, never success or floor credit.
Meaningful IDs conservatively coalesce related witnesses; exact query aliases
share one case and retain every original obligation identity.
"""
import hashlib
import json
from pathlib import Path
import sys


# Version 1 pins the complete immutable projection independently reviewed at
# 762b97e: original IDs/candidates, source links/metadata, inventory and holds.
# Activations, case aliases, assembly and owner/disposition state are excluded.
# New supplemental task witnesses live outside the original C/L/Q/B/E/F/I IDs.
CANONICAL_PROVENANCE_SHA256_V1 = "3345cf5712b1bdbf467d1651784fdb8bccc596805038da0d54e7a123384e3a4e"


def original_id(identity):
    return identity == "exclusions" or identity.startswith(("E.", "F.", "I.")) or (
        len(identity) >= 3 and identity[0] in "CLQB" and identity[1:3].isdigit()
    )


def canonical_provenance_digest(provenance):
    def records(name, fields):
        return sorted(
            ({key: record[key] for key in fields if key in record}
             for record in provenance[name] if original_id(record["id"])),
            key=lambda record: record["id"],
        )

    projection = {
        "projection_version": 1,
        "schema_version": provenance["schema_version"],
        "design_snapshots": provenance["design_snapshots"],
        "seed_ids": sorted(provenance["seed_ids"]),
        "sources": provenance["sources"],
        "inventory": sorted(provenance["inventory"], key=lambda record: record["entry"]),
        "holds": sorted(provenance["holds"], key=lambda record: record["id"]),
        "forms": records("forms", ("id", "description", "source_keys")),
        "obligations": records("obligations", ("id", "form_id", "candidate", "source_keys", "evidence")),
    }
    encoded = json.dumps(projection, sort_keys=True, ensure_ascii=False, separators=(",", ":")).encode("utf-8")
    return hashlib.sha256(encoded).hexdigest()


def load(root: Path):
    manifest = json.loads((root / "manifest.json").read_text(encoding="utf-8"))
    provenance = json.loads((root / manifest["provenance_file"]).read_text(encoding="utf-8"))
    cases = []
    for name in manifest["case_files"]:
        path = (root / name).resolve()
        if not path.is_relative_to(root.resolve()):
            raise ValueError("case path escapes corpus")
        cases.extend(json.loads(path.read_text(encoding="utf-8")))
    return manifest, provenance, cases


def require(condition, message):
    if not condition:
        raise ValueError(message)


def unique(records, label):
    result = {}
    for record in records:
        require(record["id"] not in result, f"duplicate {label}: {record['id']}")
        result[record["id"]] = record
    return result


# Fixed context assemblies preserve every original candidate byte. These are
# conformance evidence wrappers, never runtime query rewriting.
ASSEMBLIES = {"standalone": ("", ""), "pipeline-tail": ("FROM main | ", ""),
              "eval-rhs": ("FROM main | eval result=", "")}


def audit_recovery_classification(case, expected, inventory):
    """Validate explicit evidence/credit metadata, never parse a query here.

    The Go corpus verifies the original token, owning canonical stage and its
    located unsupported diagnostic. Raw parser errors remain a separate layer.
    """
    recovery = case.get("recovery_classification")
    if recovery is None:
        return False
    require(isinstance(recovery, dict) and set(recovery) == {"kind", "command", "start", "end", "stage_id"}, "invalid recovery classification metadata")
    require(case.get("floor_credit") is False and not case.get("hold_ids"), "recovery classification has no floor credit")
    require(case["status"] == "invalid" and not case["syntax_complete"] and case.get("expected_codes") == ["SPL_SYNTAX_ERROR"], "recovery classification must retain original parser errors")
    require(expected["status"] == "incomplete" and not expected["syntax_complete"] and not expected["semantic_complete"], "recovery classification cannot promote complete/valid support")
    require(expected["expected_codes"] == ["SPL_UNSUPPORTED_SEMANTICS"], "recovery classification cannot demote definite/profile/module findings")
    command, start, end = recovery["command"], recovery["start"], recovery["end"]
    text = case["document"]["text"].encode("utf-8")
    require(isinstance(command, str) and command.isascii() and command.isidentifier() and command == command.lower(), "invalid recovery classification command")
    require(type(start) is int and type(end) is int and 0 <= start < end <= len(text) and text[start:end] == command.encode("utf-8"), "recovery classification needs original nonempty token slice")
    stages = {f"stage-{i}": name for i, name in enumerate(expected["stage_commands"])}
    require(stages.get(recovery["stage_id"]) == command, "recovery classification owning stage mismatch")
    index = int(recovery["stage_id"].removeprefix("stage-"))
    require(not expected["stage_complete"][index], "recovery classification stage cannot be complete")
    by_name = {entry["entry"]: entry for entry in inventory}
    if recovery["kind"] == "unknown_command":
        require(command not in by_name and command not in {"import", "export", "function"}, "unknown classification names a known command")
    elif recovery["kind"] == "native_deferred":
        entry = by_name.get(command)
        require(entry is not None and entry["splunkd"] and entry["classification"] == "Deferred grammar/effects; U", "native_deferred classification lacks durable inventory")
    else:
        require(False, "unknown recovery classification kind")
    return True


def audit_canonical(case, inventory=()):
    if "canonical" not in case:
        require(not (case.get("canonical_evidence", "independent") == "independent" and "canonical_assertions" in case), "independent assertions require full exact canonical object")
        require("recovery_classification" not in case, "recovery classification needs canonical evidence")
        return False
    expected = case["canonical"]
    require(isinstance(expected, dict), "canonical expectation must be an object")
    require(expected.get("phase") == "analysis" and expected.get("scope") == "canonical-result", "invalid canonical layer")
    keys = {"status", "syntax_complete", "semantic_complete", "expected_codes", "references", "fields", "removed", "open", "uncertain", "stage_commands", "stage_complete"}
    require(keys <= expected.keys(), "incomplete canonical expectation")
    require(expected["status"] in {"valid", "invalid", "incomplete"}, "invalid canonical status")
    require(not expected["syntax_complete"] or case["syntax_complete"], "canonical cannot promote unproved syntax")
    require(expected["syntax_complete"] or not expected["semantic_complete"], "canonical cannot promote semantics of unproved syntax")
    recovery = audit_recovery_classification(case, expected, inventory)
    require(case["status"] != "invalid" or expected["status"] == "invalid" or recovery, "canonical cannot erase definite syntax findings")
    if case.get("hold_ids"):
        require(expected["status"] != "valid" and not expected["semantic_complete"], "canonical cannot promote held evidence")
    if expected["status"] == "valid":
        require(expected["syntax_complete"] and expected["semantic_complete"], "canonical valid requires complete coverage")
    require(len(expected["stage_commands"]) == len(expected["stage_complete"]), "canonical stage assertions disagree")
    require(all(isinstance(expected[key], list) for key in ("expected_codes", "references", "fields", "removed", "stage_commands", "stage_complete")), "canonical arrays required")
    for ref in expected["references"]:
        require(isinstance(ref, dict) and {"original_name", "normalized_name", "kind", "role", "binding", "start", "end"} <= ref.keys(), "incomplete canonical reference")
        raw, start, end = case["document"]["text"].encode("utf-8"), ref["start"], ref["end"]
        require(type(start) is int and type(end) is int and 0 <= start < end <= len(raw), "canonical reference requires nonempty original bounds")
        require(isinstance(ref["original_name"], str) and raw[start:end] == ref["original_name"].encode("utf-8"), "canonical reference source changed")
    audit_canonical_assertions(case)
    return True


def audit_canonical_assertions(case):
    """Separate independent soundness obligations from representation snapshots."""
    evidence = case.get("canonical_evidence", "independent")
    require(evidence in {"independent", "regression_snapshot"}, "unknown canonical evidence kind")
    if evidence == "independent":
        if "canonical_assertions" not in case:
            return
        require(isinstance(case.get("canonical"), dict), "independent assertions require full exact canonical object")
    a = case.get("canonical_assertions")
    require(isinstance(a, dict), "regression snapshot requires independent assertions")
    arrays = {"required_codes", "forbidden_codes", "required_references", "forbidden_references",
              "required_fields", "forbidden_field_names", "required_stages"}
    keys = arrays | {"category", "basis", "status", "syntax_complete", "semantic_complete", "max_scopes", "final_state"}
    require(keys <= set(a) <= keys | {"required_scopes"}, "invalid independent assertion keys")
    require(a["category"] in {"recovery", "representation"} and isinstance(a["basis"], str) and a["basis"].strip(), "snapshot needs independent eligibility basis")
    require(a["status"] in {"valid", "invalid", "incomplete"}, "invalid independent status")
    require(type(a["syntax_complete"]) is bool and type(a["semantic_complete"]) is bool, "independent coverage must be boolean")
    require(not a["semantic_complete"] or a["syntax_complete"], "independent semantics cannot promote unproved syntax")
    require(a["status"] != "valid" or a["syntax_complete"] and a["semantic_complete"], "independent valid requires complete coverage")
    require(type(a["max_scopes"]) is int and a["max_scopes"] > 0, "independent scope bound required")
    require(all(isinstance(a[k], list) for k in arrays), "independent assertion arrays required")
    require(any(a[k] for k in ("required_references", "forbidden_references", "required_fields", "forbidden_field_names", "required_stages")), "snapshot needs nonvacuous soundness assertions")
    for key in ("required_codes", "forbidden_codes", "forbidden_field_names"):
        require(all(isinstance(v, str) and v for v in a[key]), "independent names/codes must be nonempty strings")
    require(not set(a["required_codes"]) & set(a["forbidden_codes"]), "contradictory independent codes")
    expected, raw = case["canonical"], case["document"]["text"].encode("utf-8")
    for key in ("status", "syntax_complete", "semantic_complete"):
        require(expected[key] == a[key], "snapshot differs from independent status/coverage")
    require(set(a["required_codes"]) <= set(expected["expected_codes"]), "snapshot lacks required code")
    require(not set(a["forbidden_codes"]) & set(expected["expected_codes"]), "snapshot includes forbidden code")
    ref_keys = {"original_name", "normalized_name", "kind", "role", "binding", "start", "end"}
    scope_keys = {"scope_start", "scope_end"}
    def scope_bounds(record):
        if not scope_keys & set(record):
            return
        require(scope_keys <= set(record), "independent scope bounds require a pair")
        start, end = record["scope_start"], record["scope_end"]
        require(type(start) is int and type(end) is int and 0 <= start < end <= len(raw), "invalid independent scope bounds")
    def located(record, keys):
        require(isinstance(record, dict) and set(record) == keys, "invalid independent located assertion")
        start, end = record["start"], record["end"]
        require(type(start) is int and type(end) is int and 0 <= start < end <= len(raw), "independent source bounds changed")
        require(isinstance(record["original_name"], str) and raw[start:end].decode("utf-8") == record["original_name"], "independent source slice changed")
    for ref in a["required_references"]:
        require(isinstance(ref, dict) and ref_keys - {"binding"} <= set(ref) <= ref_keys | scope_keys, "invalid independent required reference")
        located(ref, set(ref))
        scope_bounds(ref)
        require("binding" not in ref or ref["binding"] in {"source", "derived", "indeterminate", "unavailable", "not_applicable"}, "invalid required reference binding")
        require(any(all(actual[k] == v for k, v in ref.items() if k not in scope_keys) for actual in expected["references"]), "snapshot lacks required reference")
    for ref in a["forbidden_references"]:
        require(isinstance(ref, dict) and {"original_name", "start", "end"} <= set(ref) <= {"original_name", "start", "end", "role", "binding"}, "invalid forbidden reference selector")
        located(ref, set(ref))
        require("role" not in ref or ref["role"] in {"read", "create", "output", "remove", "rename", "group", "null_test"}, "invalid forbidden reference role")
        require("binding" not in ref or ref["binding"] in {"source", "derived", "indeterminate", "unavailable", "not_applicable"}, "invalid forbidden reference binding")
        require(not any(all(actual[k] == v for k,v in ref.items()) for actual in expected["references"]), "snapshot includes forbidden reference")
    for field in a["required_fields"]:
        require(isinstance(field, dict) and set(field) == {"name", "conditional"} and isinstance(field["name"], str) and field["name"] and type(field["conditional"]) is bool, "invalid independent field assertion")
        require(field in expected["fields"], "snapshot lacks required field")
    require(not set(a["forbidden_field_names"]) & {f["name"] for f in expected["fields"]}, "snapshot includes forbidden field")
    final = a["final_state"]
    if final is not None:
        require(isinstance(final, dict) and set(final) == {"fields", "removed", "open", "uncertain"}, "invalid independent exact final state")
        require(type(final["open"]) is bool and type(final["uncertain"]) is bool and isinstance(final["fields"], list) and isinstance(final["removed"], list), "invalid independent exact final state types")
        require(all(isinstance(f, dict) and set(f) == {"name", "conditional"} and isinstance(f["name"], str) and f["name"] and type(f["conditional"]) is bool for f in final["fields"]) and all(isinstance(n, str) and n for n in final["removed"]), "invalid independent exact final state fields")
        require(all(expected[key] == final[key] for key in final), "snapshot differs from independent exact final state")
    for stage in a["required_stages"]:
        require(isinstance(stage, dict) and {"command", "start", "semantic_complete"} <= set(stage) <= {"command", "start", "semantic_complete"} | scope_keys, "invalid independent stage assertion")
        scope_bounds(stage)
        command, start = stage["command"], stage["start"]
        require(isinstance(command, str) and command and command.isascii() and command == command.lower() and type(start) is int and start >= 0 and type(stage["semantic_complete"]) is bool, "invalid independent stage identity")
        implicit_search = command == "search" and start == 0 and raw.startswith(b"index=")
        require(raw[start:start+len(command)].lower() == command.encode() or implicit_search, "independent stage source changed")
        require((command, stage["semantic_complete"]) in zip(expected["stage_commands"], expected["stage_complete"]), "snapshot lacks required stage")
    scopes = a.get("required_scopes", [])
    require(isinstance(scopes, list), "independent required scopes must be an array")
    for scope in scopes:
        require(isinstance(scope, dict) and set(scope) == {"kind", "start", "end", "owner_start", "parent_start", "parent_end"}, "invalid independent scope descriptor")
        require(scope["kind"] in {"search", "subpipe", "exists"}, "invalid independent scope kind")
        require(all(type(scope[k]) is int for k in ("start", "end", "owner_start", "parent_start", "parent_end")), "invalid independent scope coordinate")
        require(0 <= scope["parent_start"] <= scope["owner_start"] <= scope["start"] < scope["end"] <= scope["parent_end"] <= len(raw), "invalid independent scope containment")


def audit(manifest, provenance, cases):
    require(manifest["schema_version"] == provenance["schema_version"] == 1, "unknown schema")
    by_id = unique(cases, "case")
    queries = set()
    for case in cases:
        key = json.dumps(case["document"], sort_keys=True, ensure_ascii=False)
        require(key not in queries, "duplicate query must use obligation aliases")
        queries.add(key)
    sources = provenance["sources"]
    forms = unique(provenance["forms"], "form")
    mandatory = {f"C{i:02}" for i in range(1, 36)} | {f"L{i:02}" for i in range(1, 26)}
    require(mandatory <= forms.keys(), "missing mandatory family disposition")
    require(set(manifest["mandatory_command_families"] + manifest["mandatory_language_families"]) == mandatory, "manifest mandatory family list changed")
    require(len([f for f in forms if f.startswith("E.")]) == 149, "expansion form inventory changed")
    require(len([f for f in forms if f.startswith("F.")]) == 34, "function form inventory changed")
    obligations = unique(provenance["obligations"], "obligation")
    require(sum(o.startswith(("E.", "F.", "I.")) for o in obligations) == 778, "expansion obligation inventory changed")
    seeds = provenance["seed_ids"]
    require(len(seeds) == len(set(seeds)) == 288 and set(seeds) <= obligations.keys(), "base seed obligation identities changed")
    inventory = provenance["inventory"]
    require(len(inventory) == len({r["entry"] for r in inventory}) == 53, "command inventory changed")
    require(sum(r["splunkd"] for r in inventory) == 50, "native inventory changed")
    require({r["entry"] for r in inventory if not r["splunkd"]} == {"decrypt", "ocsf", "route"}, "profile inventory changed")
    holds = unique(provenance["holds"], "held record")
    require(set(holds) == {f"H{i:02}" for i in range(1, 13)} | {f"EH{i:02}" for i in range(1, 6)}, "held identities changed")
    for hold in holds.values():
        require(hold["disposition"] == "held" and hold["floor_credit"] is False, "held record receives floor credit")
    for record in [*forms.values(), *obligations.values(), *holds.values(), *cases]:
        require(record["source_keys"] and set(record["source_keys"]) <= sources.keys(), f"unresolved source provenance: {record.get('id')}")
    for form in forms.values():
        require(form["disposition"] in {"active", "partial", "pending"}, "missing form disposition")
    require(canonical_provenance_digest(provenance) == CANONICAL_PROVENANCE_SHA256_V1,
            "canonical provenance identities or immutable values changed")
    aliases = set()
    canonical_ids = set()
    for case in cases:
        require(case["id"] not in holds and not set(case["obligation_ids"]) & holds.keys(), "held case cannot receive floor credit")
        if case.get("hold_ids"):
            require(set(case["hold_ids"]) <= holds.keys() and case.get("floor_credit") is False,
                    "held boundary cannot receive floor credit")
            require(case["status"] == "incomplete" and not case["syntax_complete"] and not case["semantic_complete"],
                    "held boundary cannot claim complete or invalid support")
        require(case["meaningful_id"], "missing meaningful-query alias")
        require(case["document"].get("language") == "spl2", "case dialect must be explicit")
        require(case["obligation_ids"] and case["id"] in case["obligation_ids"], "original identity not preserved")
        require(set(case["form_ids"]) <= forms.keys(), "unresolved case form")
        assertions = case["assertions"]
        require(assertions.get("phase") == "syntax" and assertions.get("scope") == "grammar-contexts-only" and assertions.get("source_exact") is True, "missing source/scope/phase assertions")
        require(case["status"] in {"invalid", "incomplete", "valid"}, "invalid status")
        require(case["semantic_complete"] is False, "frontend cannot advertise complete semantics")
        if case["status"] == "invalid":
            require(case["expected_codes"], "negative has no expected diagnostic")
        else:
            require(assertions.get("kinds") or assertions.get("shape_contains") or assertions.get("excerpts"), "missing typed syntax assertions")
        if audit_canonical(case, inventory):
            canonical_ids.add(case["id"])
        for oid in case["obligation_ids"]:
            require(oid not in aliases, "obligation aliases multiple queries")
            aliases.add(oid)
            require(oid in obligations and obligations[oid].get("case_id") == case["id"], "obligation alias does not close")
    for oid, obligation in obligations.items():
        require(obligation["form_id"] in forms, "obligation form missing")
        require(obligation["disposition"] in {"active", "pending"}, "obligation disposition missing")
        if obligation["disposition"] == "active":
            require(oid in aliases, "active obligation has no case")
            assembly = obligation.get("assembly")
            require(assembly in ASSEMBLIES, "missing source assembly provenance")
            if assembly == "eval-rhs":
                require(oid.startswith("F.") and obligation["candidate"].startswith(obligation["form_id"][2:] + "("), "invalid function source assembly")
            prefix, suffix = ASSEMBLIES[assembly]
            expected = prefix + obligation["candidate"] + suffix
            if oid.startswith("F."):
                require(obligation["case_id"] in canonical_ids, "active function requires canonical arity evidence")
            if oid != "E.L01.start.N1":
                require(by_id[obligation["case_id"]]["document"]["text"] == expected, "source candidate changed")
        else:
            require(oid not in aliases and "case_id" not in obligation, "pending obligation credited")
    for form in forms.values():
        owned = [o for o in obligations.values() if o["form_id"] == form["id"]]
        active = sum(o["disposition"] == "active" for o in owned)
        expected = "active" if owned and active == len(owned) else "partial" if active else "pending"
        require(form["disposition"] == expected, "form disposition disagrees with obligations")
    start = by_id[obligations["E.L01.start.N1"]["case_id"]]
    require(start["document"]["text"] == "failure index=app", "exact invalid start changed")
    credited = [c for c in cases if not c.get("hold_ids") and c.get("floor_credit", True)]
    meaningful = {c["meaningful_id"] for c in credited}
    negative = {c["meaningful_id"] for c in credited if c["status"] == "invalid"}
    sql = {c["meaningful_id"] for c in credited if any(f.startswith(("Q", "E.L15", "E.L16", "E.L17")) for f in c["form_ids"])}
    pending = sum(o["disposition"] == "pending" for o in obligations.values())
    if manifest["enforce_final_floors"]:
        require(len(canonical_ids) == len(cases), "final closure requires canonical expectations")
        require(pending == 0, "pending mandatory obligations remain")
        for key, actual in {"meaningful":len(meaningful), "definite_negative":len(negative), "sql_mixed":len(sql)}.items():
            require(actual >= manifest["final_floors"][key], f"unmet {key} floor")
    snapshots = sum(c.get("canonical_evidence") == "regression_snapshot" for c in cases if c["id"] in canonical_ids)
    return {"active_queries":len(cases), "active_obligations":len(aliases), "pending_obligations":pending, "meaningful":len(meaningful), "definite_negative":len(negative), "sql_mixed":len(sql), "held":len(holds), "canonical_queries":len(canonical_ids), "independent_canonical_queries":len(canonical_ids)-snapshots, "regression_snapshots":snapshots}


def main():
    try:
        result = audit(*load(Path(__file__).resolve().parents[1] / "testdata/spl2"))
    except (ValueError, KeyError, OSError) as error:
        print(f"SPL2 corpus audit failed: {error}", file=sys.stderr)
        return 1
    print(json.dumps(result, sort_keys=True))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
