import copy
import importlib.util
from pathlib import Path
import unittest

TOOLS = Path(__file__).resolve().parents[1]
SPEC = importlib.util.spec_from_file_location("check_spl2_corpus", TOOLS / "check_spl2_corpus.py")
CHECK = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(CHECK)


class SPL2CorpusTests(unittest.TestCase):
    def setUp(self):
        self.manifest, self.provenance, self.cases = CHECK.load(TOOLS.parent / "testdata/spl2")

    def audit(self):
        return CHECK.audit(self.manifest, self.provenance, self.cases)

    def test_repository_corpus_closes(self):
        counts = self.audit()
        self.assertGreater(counts["active_queries"], 100)
        self.assertLess(counts["meaningful"], counts["active_queries"])

    def test_duplicate_case_identity_rejected(self):
        self.cases.append(copy.deepcopy(self.cases[0]))
        with self.assertRaisesRegex(ValueError, "duplicate case"):
            self.audit()

    def test_duplicate_query_must_be_an_obligation_alias(self):
        duplicate = copy.deepcopy(self.cases[0])
        duplicate["id"] = "fake"
        self.cases.append(duplicate)
        with self.assertRaisesRegex(ValueError, "duplicate query"):
            self.audit()

    def test_provenance_must_resolve(self):
        self.cases[0]["source_keys"].append("missing-source")
        with self.assertRaisesRegex(ValueError, "source"):
            self.audit()

    def test_mandatory_family_cannot_disappear(self):
        self.provenance["forms"] = [f for f in self.provenance["forms"] if f["id"] != "L11"]
        with self.assertRaisesRegex(ValueError, "family"):
            self.audit()

    def test_obligation_cannot_disappear(self):
        self.provenance["obligations"].pop()
        with self.assertRaisesRegex(ValueError, "obligation"):
            self.audit()

    def test_held_record_cannot_get_floor_credit(self):
        self.provenance["holds"][0]["floor_credit"] = True
        with self.assertRaisesRegex(ValueError, "held"):
            self.audit()

    def test_invalid_start_is_exact(self):
        case = next(c for c in self.cases if "E.L01.start.N1" in c["obligation_ids"])
        case["document"]["text"] = "FROM main | failure index=app"
        with self.assertRaisesRegex(ValueError, "exact invalid start"):
            self.audit()

    def test_pending_is_not_complete_or_floor_credit(self):
        self.manifest["enforce_final_floors"] = True
        with self.assertRaisesRegex(ValueError, "pending|floor"):
            self.audit()

    def test_assertions_are_required(self):
        self.cases[0]["assertions"] = {}
        with self.assertRaisesRegex(ValueError, "assertions"):
            self.audit()

    def test_same_count_form_and_obligation_rename_is_rejected(self):
        old, new = "E.C01.literal", "E.C01.invented_replacement"
        for form in self.provenance["forms"]:
            if form["id"] == old:
                form["id"] = new
        for obligation in self.provenance["obligations"]:
            if obligation["form_id"] == old:
                obligation["form_id"] = new
                obligation["id"] = obligation["id"].replace(old, new)
        with self.assertRaisesRegex(ValueError, "canonical provenance"):
            self.audit()

    def test_same_count_canonical_value_changes_are_rejected(self):
        changes = {
            "candidate": lambda p: next(o for o in p["obligations"] if o["id"] == "E.C01.literal.P1").update(candidate="search replacement"),
            "base_candidate": lambda p: next(o for o in p["obligations"] if o["id"] == "C01-P1").update(candidate="search replacement"),
            "source_url": lambda p: p["sources"]["start"].update(url="https://example.invalid/replacement"),
            "source_metadata": lambda p: p["sources"]["start"].update(retrieved="2000-01-01"),
            "form_source": lambda p: next(f for f in p["forms"] if f["id"] == "E.C01.literal").update(source_keys=["start"]),
            "obligation_source": lambda p: next(o for o in p["obligations"] if o["id"] == "E.C01.literal.P1").update(source_keys=["start"]),
            "inventory": lambda p: p["inventory"][0].update(context="replacement"),
            "held_evidence": lambda p: p["holds"][0].update(evidence="replacement"),
            "design_snapshot": lambda p: p["design_snapshots"].update(matrix_sha256="0" * 64),
        }
        original = copy.deepcopy(self.provenance)
        for name, change in changes.items():
            with self.subTest(name=name):
                self.provenance = copy.deepcopy(original)
                change(self.provenance)
                with self.assertRaisesRegex(ValueError, "canonical provenance"):
                    self.audit()

    def test_same_count_seed_rename_is_rejected(self):
        old, new = "C01-P1", "C01-replacement"
        self.provenance["seed_ids"] = [new if item == old else item for item in self.provenance["seed_ids"]]
        next(o for o in self.provenance["obligations"] if o["id"] == old)["id"] = new
        with self.assertRaisesRegex(ValueError, "canonical provenance"):
            self.audit()

    def test_supplemental_boundary_case_can_be_added(self):
        case = copy.deepcopy(self.cases[0])
        case.update(id="T2.audit.new-boundary", meaningful_id="T2.audit.new-boundary", obligation_ids=["T2.audit.new-boundary"], form_ids=["T2.audit.new-form"])
        case["document"]["text"] = "index=additional host=proof"
        self.cases.append(case)
        self.provenance["forms"].append({"id":"T2.audit.new-form", "source_keys":case["source_keys"], "disposition":"active", "owner":"test"})
        self.provenance["obligations"].append({"id":case["id"], "form_id":"T2.audit.new-form", "source_keys":case["source_keys"], "candidate":case["document"]["text"], "disposition":"active", "case_id":case["id"], "assembly":"standalone", "evidence":"supplemental-boundary"})
        self.audit()

    def test_legitimate_original_obligation_activation_is_allowed(self):
        obligation = next(o for o in self.provenance["obligations"] if o["id"] == "E.C01.literal.P1")
        # Establish the pending state in this isolated fixture even after later
        # tasks activate the original obligation, preserving any exact aliases.
        for case in list(self.cases):
            if obligation["id"] not in case["obligation_ids"]:
                continue
            case["obligation_ids"].remove(obligation["id"])
            if not case["obligation_ids"]:
                self.cases.remove(case)
            elif case["id"] == obligation["id"]:
                case["id"] = case["obligation_ids"][0]
                for alias in self.provenance["obligations"]:
                    if alias["id"] in case["obligation_ids"]:
                        alias["case_id"] = case["id"]
        obligation.update(disposition="pending")
        obligation.pop("case_id", None)
        obligation.pop("assembly", None)
        form = next(f for f in self.provenance["forms"] if f["id"] == obligation["form_id"])
        siblings = [o for o in self.provenance["obligations"] if o["form_id"] == form["id"]]
        form["disposition"] = "partial" if any(o["disposition"] == "active" for o in siblings) else "pending"
        before = self.audit()["active_obligations"]
        case = next((c for c in self.cases if c["document"]["text"] == obligation["candidate"]), None)
        if case is None:
            case = copy.deepcopy(self.cases[0])
            case.update(id=obligation["id"], meaningful_id=obligation["id"], obligation_ids=[], form_ids=[form["id"]], source_keys=obligation["source_keys"])
            case["document"]["text"] = obligation["candidate"]
            self.cases.append(case)
        case["obligation_ids"].append(obligation["id"])
        obligation.update(disposition="active", case_id=case["id"], assembly="standalone")
        form["disposition"] = "active" if all(o["disposition"] == "active" for o in siblings) else "partial"
        self.assertEqual(self.audit()["active_obligations"], before + 1)

    def test_held_boundary_is_excluded_from_meaningful_credit(self):
        before = self.audit()["meaningful"]
        case = copy.deepcopy(self.cases[0])
        case.update(id="T2.audit.held", meaningful_id="T2.audit.held", obligation_ids=["T2.audit.held"], form_ids=["T2.audit.held-form"], hold_ids=["H11"], floor_credit=False, syntax_complete=False, expected_codes=["SPL_UNSUPPORTED_SEMANTICS"])
        case["document"]["text"] = "FROM main | where amount BETWEEN 13 and 17"
        self.cases.append(case)
        self.provenance["forms"].append({"id":"T2.audit.held-form", "source_keys":case["source_keys"], "disposition":"active", "owner":"test"})
        self.provenance["obligations"].append({"id":case["id"], "form_id":"T2.audit.held-form", "source_keys":case["source_keys"], "candidate":case["document"]["text"], "disposition":"active", "case_id":case["id"], "assembly":"standalone", "evidence":"supplemental-boundary"})
        self.assertEqual(self.audit()["meaningful"], before)
        case["floor_credit"] = True
        with self.assertRaisesRegex(ValueError, "held"):
            self.audit()
