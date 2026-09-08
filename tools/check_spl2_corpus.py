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


def audit_canonical(case):
    if "canonical" not in case:
        return False
    expected = case["canonical"]
    require(isinstance(expected, dict), "canonical expectation must be an object")
    require(expected.get("phase") == "analysis" and expected.get("scope") == "canonical-result", "invalid canonical layer")
    keys = {"status", "syntax_complete", "semantic_complete", "expected_codes", "references", "fields", "removed", "open", "uncertain", "stage_commands", "stage_complete"}
    require(keys <= expected.keys(), "incomplete canonical expectation")
    require(expected["status"] in {"valid", "invalid", "incomplete"}, "invalid canonical status")
    require(not expected["syntax_complete"] or case["syntax_complete"], "canonical cannot promote unproved syntax")
    require(expected["syntax_complete"] or not expected["semantic_complete"], "canonical cannot promote semantics of unproved syntax")
    require(case["status"] != "invalid" or expected["status"] == "invalid", "canonical cannot erase definite syntax findings")
    if case.get("hold_ids"):
        require(expected["status"] != "valid" and not expected["semantic_complete"], "canonical cannot promote held evidence")
    if expected["status"] == "valid":
        require(expected["syntax_complete"] and expected["semantic_complete"], "canonical valid requires complete coverage")
    require(len(expected["stage_commands"]) == len(expected["stage_complete"]), "canonical stage assertions disagree")
    require(all(isinstance(expected[key], list) for key in ("expected_codes", "references", "fields", "removed", "stage_commands", "stage_complete")), "canonical arrays required")
    for ref in expected["references"]:
        require({"original_name", "normalized_name", "kind", "role", "binding", "start", "end"} <= ref.keys(), "incomplete canonical reference")
        require(case["document"]["text"].encode("utf-8")[ref["start"]:ref["end"]].decode("utf-8") == ref["original_name"], "canonical reference source changed")
    return True


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
        if audit_canonical(case):
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
    return {"active_queries":len(cases), "active_obligations":len(aliases), "pending_obligations":pending, "meaningful":len(meaningful), "definite_negative":len(negative), "sql_mixed":len(sql), "held":len(holds), "canonical_queries":len(canonical_ids)}


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
