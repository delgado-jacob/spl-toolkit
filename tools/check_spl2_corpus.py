#!/usr/bin/env python3
"""Audit repository-owned SPL2 evidence. This is not a query/semantic parser.

Pending obligations are visible intermediate work, never success or floor credit.
Meaningful IDs conservatively coalesce related witnesses; exact query aliases
share one case and retain every original obligation identity.
"""
import json
from pathlib import Path
import sys


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
    aliases = set()
    for case in cases:
        require(case["id"] not in holds and not set(case["obligation_ids"]) & holds.keys(), "held case cannot receive floor credit")
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
            require(assembly in {"standalone", "pipeline-tail"}, "missing source assembly provenance")
            expected = ("FROM main | " if assembly == "pipeline-tail" else "") + obligation["candidate"]
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
    meaningful = {c["meaningful_id"] for c in cases}
    negative = {c["meaningful_id"] for c in cases if c["status"] == "invalid"}
    sql = {c["meaningful_id"] for c in cases if any(f.startswith(("Q", "E.L15", "E.L16", "E.L17")) for f in c["form_ids"])}
    pending = sum(o["disposition"] == "pending" for o in obligations.values())
    if manifest["enforce_final_floors"]:
        require(pending == 0, "pending mandatory obligations remain")
        for key, actual in {"meaningful":len(meaningful), "definite_negative":len(negative), "sql_mixed":len(sql)}.items():
            require(actual >= manifest["final_floors"][key], f"unmet {key} floor")
    return {"active_queries":len(cases), "active_obligations":len(aliases), "pending_obligations":pending, "meaningful":len(meaningful), "definite_negative":len(negative), "sql_mixed":len(sql), "held":len(holds)}


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
