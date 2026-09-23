import importlib.util
import copy
import json
from pathlib import Path
import subprocess
import tempfile
import unittest


TOOLS = Path(__file__).resolve().parents[1]
SPEC = importlib.util.spec_from_file_location("audit_linus_spl2", TOOLS / "audit_linus_spl2.py")
AUDIT = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(AUDIT)


class LinusSPL2AuditTests(unittest.TestCase):
    def make_content(self, root: Path, *, standalone_count=45, metadata_count=4, ignored=None):
        detections = root / "detections"
        detections.mkdir()
        for index in range(standalone_count):
            (detections / f"generic-{index:02}.yaml").write_text(
                "title: Private standalone detection %02d\nsearch: |-\n  FROM synthetic_events | where synthetic_value > %d\n"
                % (index, index),
                encoding="utf-8",
            )
        for index in range(4):
            (detections / f"fragment-{index:02}.yaml").write_text(
                "title: Private predicate detection %02d\nsearch: |\n  synthetic_value > %d\n"
                % (index, index),
                encoding="utf-8",
            )
        for index in range(metadata_count):
            (detections / f"metadata-{index:02}.yaml").write_text(
                f"title: Private metadata only {index:02}\n",
                encoding="utf-8",
            )
        if ignored:
            (root / ".gitignore").write_text(ignored + "\n", encoding="utf-8")
        subprocess.run(["git", "init", "-q", str(root)], check=True)
        subprocess.run(["git", "-C", str(root), "config", "user.email", "audit@example.invalid"], check=True)
        subprocess.run(["git", "-C", str(root), "config", "user.name", "Audit Fixture"], check=True)
        add = ["git", "-C", str(root), "add", "detections"]
        if ignored:
            add.append(".gitignore")
        subprocess.run(add, check=True)
        subprocess.run(["git", "-C", str(root), "commit", "-qm", "fixture"], check=True)

    def write_forms(self, path: Path):
        source = TOOLS.parent / "testdata/spl2/linus-forms.json"
        path.write_bytes(source.read_bytes())

    @staticmethod
    def successful_runner(calls):
        def run(args, **kwargs):
            calls.append((list(args), dict(kwargs)))
            operation = args[1]
            if operation == "analyze":
                report = {
                    "status": "valid",
                    "diagnostics": [{"code": "SPL_GENERIC_NOTE", "category": "coverage"}],
                }
                return subprocess.CompletedProcess(args, 0, json.dumps(report), "")
            report = {
                "query_status": "incomplete",
                "coverage": {"complete": False, "reasons": ["SPL_GENERIC_GAP"]},
                "diagnostics": [{"code": "SPL_GENERIC_GAP", "category": "semantics"}],
            }
            return subprocess.CompletedProcess(args, 3, json.dumps(report), "")
        return run

    def test_strict_top_level_yaml_block_extraction(self):
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            path = root / "case.yaml"
            path.write_text(
                "title: Protected detection label\nsearch: |-\n  FROM synthetic_events\n  | where synthetic_value > 0\n",
                encoding="utf-8",
            )
            documents = AUDIT.extract_yaml_documents(root)
            self.assertEqual(len(documents), 1)
            self.assertEqual(documents[0].query, "FROM synthetic_events\n| where synthetic_value > 0")
            self.assertIn("Protected detection label", documents[0].protected_values)

            ambiguous = {
                "nested.yaml": "rule:\n  search: |\n    synthetic_value > 0\n",
                "nested-sequence.yaml": "rules:\n  - search: |\n      synthetic_value > 0\n",
                "nested-double-quoted.yaml": "rule:\n  \"search\": |\n    synthetic_value > 0\n",
                "nested-single-quoted.yaml": "rule:\n  'search': |\n    synthetic_value > 0\n",
                "nested-sequence-quoted.yaml": "rules:\n  - \"search\": |\n      synthetic_value > 0\n",
                "tagged-key.yaml": "!private search: |\n  FROM synthetic_events\n",
                "explicit-key.yaml": "? search\n: |\n  FROM synthetic_events\n",
                "folded.yaml": "search: >\n  FROM synthetic_events\n",
                "inline.yaml": "search: FROM synthetic_events\n",
                "duplicate.yaml": "search: |\n  FROM synthetic_events\nsearch: |\n  FROM synthetic_events\n",
            }
            for name, content in ambiguous.items():
                with self.subTest(name=name):
                    for existing in root.glob("*.yaml"):
                        existing.unlink()
                    (root / name).write_text(content, encoding="utf-8")
                    with self.assertRaisesRegex(AUDIT.AuditError, "ambiguous|mapping key|block scalar|duplicate"):
                        AUDIT.extract_yaml_documents(root)

    def test_accepts_top_level_literal_block_chomping_indicators(self):
        for indicator in ("|", "|-", "|+"):
            with self.subTest(indicator=indicator), tempfile.TemporaryDirectory() as temporary:
                root = Path(temporary)
                (root / "case.yaml").write_text(
                    f"search: {indicator}\n  FROM synthetic_events\n",
                    encoding="utf-8",
                )
                documents = AUDIT.extract_yaml_documents(root)
                self.assertEqual(documents[0].query, "FROM synthetic_events")

    def test_classifies_45_standalone_and_four_predicate_fragments(self):
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            self.make_content(root)
            documents = AUDIT.extract_yaml_documents(root / "detections")
            classifications = [AUDIT.classify_query(document.query) for document in documents if document.query]
            self.assertEqual(classifications.count("standalone"), 45)
            self.assertEqual(classifications.count("predicate_fragment"), 4)

    def test_full_audit_rejects_escaped_search_key_hidden_from_inventory(self):
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            self.make_content(root)
            hidden = root / "detections" / "metadata-00.yaml"
            hidden.write_text(
                '"se\\u0061rch": |\n  FROM synthetic_events | where synthetic_hidden=true\n',
                encoding="utf-8",
            )
            subprocess.run(["git", "-C", str(root), "add", str(hidden)], check=True)
            subprocess.run(["git", "-C", str(root), "commit", "-qm", "escaped key"], check=True)
            forms = root / "forms.json"
            self.write_forms(forms)
            with self.assertRaisesRegex(AUDIT.AuditError, "mapping key"):
                AUDIT.audit(root, root / "toolkit", forms, runner=self.successful_runner([]))

    def test_full_audit_rejects_anchor_and_alias_search_keys_hidden_from_inventory(self):
        hidden_keys = {
            "anchor": "&private search: |-\n  FROM synthetic_events | where synthetic_hidden=true\n",
            "anchor-sequence": "rules:\n  - &private search: |-\n      FROM synthetic_events | where synthetic_hidden=true\n",
            "alias": "private_key: &private search\n*private: |-\n  FROM synthetic_events | where synthetic_hidden=true\n",
            "alias-sequence": "private_key: &private search\nrules:\n  - *private: |-\n      FROM synthetic_events | where synthetic_hidden=true\n",
        }
        for name, content in hidden_keys.items():
            with self.subTest(name=name), tempfile.TemporaryDirectory() as temporary:
                root = Path(temporary)
                self.make_content(root)
                hidden = root / "detections" / "metadata-00.yaml"
                hidden.write_text(content, encoding="utf-8")
                subprocess.run(["git", "-C", str(root), "add", str(hidden)], check=True)
                subprocess.run(["git", "-C", str(root), "commit", "-qm", "node property key"], check=True)
                forms = root / "forms.json"
                self.write_forms(forms)
                with self.assertRaisesRegex(AUDIT.AuditError, "mapping key"):
                    AUDIT.audit(root, root / "toolkit", forms, runner=self.successful_runner([]))

    def test_full_audit_rejects_ignored_untracked_yaml_even_when_counts_match(self):
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            ignored = "detections/ignored.yaml"
            self.make_content(root, standalone_count=44, ignored=ignored)
            (root / ignored).write_text(
                "search: |\n  FROM synthetic_events | where synthetic_ignored=true\n",
                encoding="utf-8",
            )
            forms = root / "forms.json"
            self.write_forms(forms)
            with self.assertRaisesRegex(AUDIT.AuditError, "tracked at HEAD"):
                AUDIT.audit(root, root / "toolkit", forms, runner=self.successful_runner([]))

    def test_audit_embeds_fragments_and_emits_aggregate_only_output(self):
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            self.make_content(root)
            forms = root / "forms.json"
            self.write_forms(forms)
            calls = []
            result = AUDIT.audit(root, root / "toolkit", forms, runner=self.successful_runner(calls))

            commit = subprocess.run(
                ["git", "-C", str(root), "rev-parse", "HEAD"],
                check=True, capture_output=True, text=True,
            ).stdout.strip()
            self.assertEqual(result["audited_commit"], commit)
            self.assertEqual(result["counts"], {
                "yaml_documents": 53,
                "search_blocks": 49,
                "standalone_programs": 45,
                "predicate_fragments": 4,
                "form_obligations": 65,
                "analyze_invocations": 114,
                "requirements_invocations": 114,
            })
            self.assertEqual(len(result["form_ids"]), 65)
            self.assertEqual(result["status_classes"], {
                "analyze": {"valid": 114},
                "requirements": {"incomplete": 114},
            })
            self.assertEqual(result["diagnostic_counts"]["analyze"]["codes"], {"SPL_GENERIC_NOTE": 114})
            self.assertEqual(result["diagnostic_counts"]["requirements"]["categories"], {"semantics": 114})

            query_arguments = [args[args.index("--query") + 1] for args, _ in calls]
            self.assertIn("FROM synthetic_events | where (synthetic_value > 0)", query_arguments)
            self.assertIn("FROM synthetic_events | where synthetic_value > 0", query_arguments)
            for args, kwargs in calls:
                self.assertEqual(args[2:8], [
                    "--language", "spl2", "--profile", "splunkd",
                    "--compatibility-version", "current",
                ])
                self.assertEqual(args[-2], "--format")
                self.assertEqual(args[-1], "json")
                self.assertTrue(kwargs["capture_output"])

            rendered = json.dumps(result, sort_keys=True)
            self.assertNotIn("Private standalone detection", rendered)
            self.assertNotIn("generic-00.yaml", rendered)
            self.assertNotIn("synthetic_value > 0", rendered)
            self.assertEqual(set(result), {
                "audited_commit", "counts", "form_ids", "status_classes", "diagnostic_counts",
            })

    def test_subprocess_status_exits_are_checked_without_leaking_output(self):
        valid = {"status": "invalid", "diagnostics": []}
        completed = subprocess.CompletedProcess([], 1, json.dumps(valid), "protected stderr")
        self.assertEqual(AUDIT.parse_toolkit_result("analyze", completed)[0], "invalid")

        failed = subprocess.CompletedProcess([], 2, "protected search text", "protected path")
        with self.assertRaises(AUDIT.AuditError) as raised:
            AUDIT.parse_toolkit_result("analyze", failed)
        self.assertNotIn("protected", str(raised.exception))

    def test_output_leak_rejection(self):
        with self.assertRaisesRegex(AUDIT.AuditError, "protected input"):
            AUDIT.reject_protected_output(
                {"form_ids": ["M11.layout.standalone"], "leak": "Private detection label"},
                {"Private detection label"},
            )

    def test_form_inputs_cannot_turn_output_ids_or_queries_into_a_leak(self):
        with tempfile.TemporaryDirectory() as temporary:
            path = Path(temporary) / "forms.json"
            self.write_forms(path)
            payload = json.loads(path.read_text(encoding="utf-8"))
            original = payload["obligations"][0]
            mutations = {
                "identity": {**original, "id": "M11.private-detection-name"},
                "query": {**original, "query": "FROM customer_production_events"},
            }
            for name, mutation in mutations.items():
                with self.subTest(name=name):
                    changed = dict(payload)
                    changed["obligations"] = [mutation, *payload["obligations"][1:]]
                    path.write_text(json.dumps(changed), encoding="utf-8")
                    with self.assertRaisesRegex(AUDIT.AuditError, "form"):
                        AUDIT.load_forms(path)

    def test_external_audit_rejects_weakened_approved_form_contract(self):
        with tempfile.TemporaryDirectory() as temporary:
            path = Path(temporary) / "forms.json"
            self.write_forms(path)
            approved = json.loads(path.read_text(encoding="utf-8"))
            mutations = {
                "authorship": lambda matrix: matrix.update(authorship="external-source"),
                "disposition": lambda matrix: matrix["obligations"][0].update(disposition="pending"),
                "function": lambda matrix: next(
                    obligation for obligation in matrix["obligations"]
                    if obligation["family"] == "M11.function"
                ).update(function="upper"),
            }
            for name, mutate in mutations.items():
                with self.subTest(name=name):
                    matrix = copy.deepcopy(approved)
                    mutate(matrix)
                    path.write_text(json.dumps(matrix), encoding="utf-8")
                    with self.assertRaisesRegex(AUDIT.AuditError, "approved form matrix"):
                        AUDIT.load_forms(path)

    def test_exact_commit_rejects_uncommitted_content(self):
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            self.make_content(root)
            (root / "detections" / "uncommitted.yaml").write_text(
                "search: |\n  FROM synthetic_events\n", encoding="utf-8",
            )
            with self.assertRaisesRegex(AUDIT.AuditError, "exact committed snapshot"):
                AUDIT.exact_commit(root / "detections")


if __name__ == "__main__":
    unittest.main()
