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
