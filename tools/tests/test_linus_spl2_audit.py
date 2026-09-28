import importlib.util
import copy
from contextlib import redirect_stderr, redirect_stdout
import io
import json
from pathlib import Path
import subprocess
import tempfile
import unittest
from unittest.mock import patch


TOOLS = Path(__file__).resolve().parents[1]
SPEC = importlib.util.spec_from_file_location("audit_linus_spl2", TOOLS / "audit_linus_spl2.py")
AUDIT = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(AUDIT)


class LinusSPL2AuditTests(unittest.TestCase):
    @staticmethod
    def accepted_exact_ref_aggregate():
        return {
            "audited_commit": "d4db9bae4adba00bc59d9becbf4014908e7a3ef5",
            "counts": {
                "yaml_documents": 54,
                "search_blocks": 49,
                "standalone_programs": 45,
                "predicate_fragments": 4,
                "form_obligations": 65,
                "analyze_invocations": 114,
                "requirements_invocations": 114,
            },
            "content_syntax_counts": {
                "standalone": {"complete": 45, "incomplete": 0},
                "predicate_fragment": {"complete": 4, "incomplete": 0},
            },
            "content_semantic_counts": {
                "standalone": {"complete": 45, "incomplete": 0},
                "predicate_fragment": {"complete": 4, "incomplete": 0},
            },
            "content_status_classes": {
                "standalone": {
                    "analyze": {"valid": 40, "invalid": 5},
                    "requirements": {"valid": 11, "incomplete": 29, "invalid": 5},
                },
                "predicate_fragment": {
                    "analyze": {"valid": 4},
                    "requirements": {"valid": 4},
                },
            },
            "content_diagnostic_counts": {
                "standalone": {
                    "analyze": {
                        "codes": {"SPL_UNAVAILABLE_FIELD": 25},
                        "categories": {"unavailable_field": 25},
                    },
                    "requirements": {
                        "codes": {"SPL_UNAVAILABLE_FIELD": 25},
                        "categories": {"unavailable_field": 25},
                    },
                },
                "predicate_fragment": {
                    "analyze": {"codes": {}, "categories": {}},
                    "requirements": {"codes": {}, "categories": {}},
                },
            },
            "form_ids": [f"M11.synthetic.{index:02}" for index in range(65)],
        }

    def test_exact_ref_acceptance_requires_complete_standalone_semantics(self):
        AUDIT.validate_exact_ref_acceptance(self.accepted_exact_ref_aggregate())

    def test_exact_ref_acceptance_rejects_aggregate_drift_without_leaking_content(self):
        mutations = {
            "commit": lambda result: result.update(audited_commit="private-source-path"),
            "semantic": lambda result: result["content_semantic_counts"]["standalone"].update(complete=44, incomplete=1),
            "status": lambda result: result["content_status_classes"]["standalone"]["analyze"].update(valid=39, incomplete=1),
            "requirements": lambda result: result["content_status_classes"]["standalone"]["requirements"].update(valid=10, incomplete=30),
            "fragment_semantic": lambda result: result["content_semantic_counts"]["predicate_fragment"].update(complete=3, incomplete=1),
            "fragment": lambda result: result["content_status_classes"]["predicate_fragment"]["requirements"].update(valid=3, incomplete=1),
            "unsupported": lambda result: result["content_diagnostic_counts"]["standalone"]["analyze"]["codes"].update(SPL_UNSUPPORTED_SEMANTICS=1),
            "ambiguity": lambda result: result["content_diagnostic_counts"]["standalone"]["analyze"]["codes"].update(SPL_AMBIGUOUS_FIELD=1),
            "requirements_ambiguity": lambda result: result["content_diagnostic_counts"]["standalone"]["requirements"]["codes"].update(SPL_AMBIGUOUS_FIELD=1),
            "category": lambda result: result["content_diagnostic_counts"]["standalone"]["analyze"]["categories"].update(unsupported_semantics=1, unavailable_field=24),
            "private_diagnostic": lambda result: result["content_diagnostic_counts"]["standalone"]["analyze"]["codes"].update({"private-query-text": 1}),
        }
        for name, mutate in mutations.items():
            with self.subTest(name=name):
                aggregate = self.accepted_exact_ref_aggregate()
                mutate(aggregate)
                with self.assertRaises(AUDIT.AuditError) as raised:
                    AUDIT.validate_exact_ref_acceptance(aggregate)
                self.assertNotIn("private", str(raised.exception))

    def test_cli_enforces_exact_ref_acceptance_after_reusable_audit(self):
        options = ["--content-root", "unused", "--toolkit-bin", "unused", "--forms", "unused"]
        aggregate = self.accepted_exact_ref_aggregate()
        output, errors = io.StringIO(), io.StringIO()
        with patch.object(AUDIT, "audit", return_value=aggregate), redirect_stdout(output), redirect_stderr(errors):
            self.assertEqual(AUDIT.main(options), 0)
        self.assertEqual(json.loads(output.getvalue()), aggregate)
        self.assertEqual(errors.getvalue(), "")

        aggregate["audited_commit"] = "private-source-path"
        output, errors = io.StringIO(), io.StringIO()
        with patch.object(AUDIT, "audit", return_value=aggregate), redirect_stdout(output), redirect_stderr(errors):
            self.assertEqual(AUDIT.main(options), 1)
        self.assertEqual(output.getvalue(), "")
        self.assertNotIn("private", errors.getvalue())

    def test_standalone_attribution_requires_matching_status_coverage_and_diagnostic(self):
        ambiguous = [{"code": "SPL_AMBIGUOUS_FIELD", "category": "unsupported_semantics"}]
        unavailable = [{"code": "SPL_UNAVAILABLE_FIELD", "category": "unavailable_field"}]
        note = [{"code": "SPL_GENERIC_NOTE", "category": "coverage"}]
        for status, complete, diagnostics in (
            ("valid", True, note),
            ("invalid", True, unavailable),
        ):
            with self.subTest(status=status):
                AUDIT.validate_standalone_attribution(status, complete, diagnostics)
        for status, complete, diagnostics in (
            ("valid", True, ambiguous),
            ("valid", False, note),
            ("valid", True, [{"code": "SPL_UNSUPPORTED_SEMANTICS", "category": "unsupported_semantics"}]),
            ("incomplete", False, note),
            ("incomplete", True, ambiguous),
            ("incomplete", False, ambiguous + unavailable),
            ("incomplete", False, ambiguous),
            ("incomplete", False, [{"code": "SPL_AMBIGUOUS_FIELD", "category": "unavailable_field"}]),
            ("invalid", True, note),
            ("invalid", False, unavailable),
            ("invalid", True, unavailable + ambiguous),
            ("invalid", True, unavailable + [{"code": "SPL_UNSUPPORTED_SEMANTICS", "category": "unsupported_semantics"}]),
            ("invalid", True, [{"code": "SPL_UNAVAILABLE_FIELD", "category": "unsupported_semantics"}]),
        ):
            with self.subTest(status=status, complete=complete, diagnostics=diagnostics):
                with self.assertRaises(AUDIT.AuditError) as raised:
                    AUDIT.validate_standalone_attribution(status, complete, diagnostics)
                self.assertNotIn("private", str(raised.exception))

    def test_exact_ref_attribution_gate_rejects_shifted_unavailable_with_unchanged_aggregate(self):
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            content = root / "content"
            content.mkdir()
            self.make_content(content)
            forms = root / "forms.json"
            self.write_forms(forms)
            classes = [
                AUDIT.classify_query(document.query)
                for document in AUDIT.extract_yaml_documents(content)
                if document.query
            ]
            invalid_index, valid_index = [
                index for index, classification in enumerate(classes)
                if classification == "standalone"
            ][:2]

            def runner(shift_unavailable=False):
                calls = []
                successful = self.successful_runner(calls)

                def run(args, **kwargs):
                    query_index = len(calls) // 2
                    completed = successful(args, **kwargs)
                    if args[1] != "analyze" or query_index not in {invalid_index, valid_index}:
                        return completed
                    status = "valid"
                    diagnostic = {"code": "SPL_GENERIC_NOTE", "category": "coverage"}
                    if query_index == invalid_index:
                        status = "invalid"
                        if not shift_unavailable:
                            diagnostic = {"code": "SPL_UNAVAILABLE_FIELD", "category": "unavailable_field"}
                    elif shift_unavailable:
                        diagnostic = {"code": "SPL_UNAVAILABLE_FIELD", "category": "unavailable_field"}
                    report = {
                        "status": status,
                        "coverage": {"syntax_complete": True, "semantic_complete": True},
                        "diagnostics": [diagnostic],
                    }
                    return subprocess.CompletedProcess(args, AUDIT.STATUS_EXITS[status], json.dumps(report), "")

                return run

            expected = AUDIT.audit(content, root / "toolkit", forms, runner=runner())
            AUDIT.audit(content, root / "toolkit", forms, runner=runner(), enforce_field_attribution=True)
            changed = AUDIT.audit(content, root / "toolkit", forms, runner=runner(shift_unavailable=True))
            self.assertEqual(changed, expected)
            with self.assertRaises(AUDIT.AuditError) as raised:
                AUDIT.audit(
                    content, root / "toolkit", forms,
                    runner=runner(shift_unavailable=True), enforce_field_attribution=True,
                )
            self.assertNotIn(str(content), str(raised.exception))

    def test_malformed_diagnostic_code_fails_cli_with_sanitized_error(self):
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            content = root / "content"
            content.mkdir()
            self.make_content(content)
            forms = root / "forms.json"
            self.write_forms(forms)
            classes = [
                AUDIT.classify_query(document.query)
                for document in AUDIT.extract_yaml_documents(content)
                if document.query
            ]
            first_standalone = classes.index("standalone")
            calls = []
            successful = self.successful_runner(calls)

            def malformed_runner(args, **kwargs):
                query_index = len(calls) // 2
                completed = successful(args, **kwargs)
                if query_index != first_standalone or args[1] != "analyze":
                    return completed
                report = {
                    "status": "incomplete",
                    "coverage": {"syntax_complete": True, "semantic_complete": False},
                    "diagnostics": [{"code": ["private-query-text"], "category": "unsupported_semantics"}],
                }
                return subprocess.CompletedProcess(args, 3, json.dumps(report), "private stderr")

            real_audit = AUDIT.audit

            def run_with_malformed_code(*args, **kwargs):
                return real_audit(*args, runner=malformed_runner, **kwargs)

            options = ["--content-root", str(content), "--toolkit-bin", str(root / "toolkit"), "--forms", str(forms)]
            output, errors = io.StringIO(), io.StringIO()
            with patch.object(AUDIT, "audit", side_effect=run_with_malformed_code), redirect_stdout(output), redirect_stderr(errors):
                self.assertEqual(AUDIT.main(options), 1)
            self.assertEqual(output.getvalue(), "")
            self.assertNotIn("private", errors.getvalue())

    def test_cli_opts_into_per_report_attribution(self):
        aggregate = self.accepted_exact_ref_aggregate()

        def acceptance_aware_audit(*args, **kwargs):
            if not kwargs.get("enforce_field_attribution"):
                raise AUDIT.AuditError("per-report attribution was not checked")
            return aggregate

        output, errors = io.StringIO(), io.StringIO()
        options = ["--content-root", "unused", "--toolkit-bin", "unused", "--forms", "unused"]
        with patch.object(AUDIT, "audit", side_effect=acceptance_aware_audit), redirect_stdout(output), redirect_stderr(errors):
            self.assertEqual(AUDIT.main(options), 0)
        self.assertEqual(json.loads(output.getvalue()), aggregate)
        self.assertEqual(errors.getvalue(), "")

    def make_content(
        self,
        root: Path,
        *,
        standalone_count=45,
        metadata_count=5,
        metadata_outside_detections=False,
        ignored=None,
    ):
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
        metadata = root / "metadata" if metadata_outside_detections else detections
        metadata.mkdir(exist_ok=True)
        for index in range(metadata_count):
            (metadata / f"metadata-{index:02}.yaml").write_text(
                f"title: Private metadata only {index:02}\n",
                encoding="utf-8",
            )
        if ignored:
            (root / ".gitignore").write_text(ignored + "\n", encoding="utf-8")
        subprocess.run(["git", "init", "-q", str(root)], check=True)
        subprocess.run(["git", "-C", str(root), "config", "user.email", "audit@example.invalid"], check=True)
        subprocess.run(["git", "-C", str(root), "config", "user.name", "Audit Fixture"], check=True)
        add = ["git", "-C", str(root), "add", "detections"]
        if metadata_outside_detections:
            add.append("metadata")
        if ignored:
            add.append(".gitignore")
        subprocess.run(add, check=True)
        subprocess.run(["git", "-C", str(root), "commit", "-qm", "fixture"], check=True)

    def write_forms(self, path: Path):
        source = TOOLS.parent / "testdata/spl2/linus-forms.json"
        path.write_bytes(source.read_bytes())

    @staticmethod
    def successful_runner(calls, syntax_incomplete=frozenset()):
        def run(args, **kwargs):
            calls.append((list(args), dict(kwargs)))
            operation = args[1]
            if operation == "analyze":
                query = args[args.index("--query") + 1]
                syntax_complete = query not in syntax_incomplete
                report = {
                    "status": "valid" if syntax_complete else "invalid",
                    "coverage": {"syntax_complete": syntax_complete, "semantic_complete": syntax_complete},
                    "diagnostics": [{
                        "code": "SPL_GENERIC_NOTE" if syntax_complete else "SPL_SYNTAX_ERROR",
                        "category": "coverage" if syntax_complete else "syntax",
                    }],
                }
                return subprocess.CompletedProcess(args, 0 if syntax_complete else 1, json.dumps(report), "")
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
        with tempfile.TemporaryDirectory() as temporary, tempfile.TemporaryDirectory() as auxiliary:
            root = Path(temporary)
            self.make_content(root)
            hidden = root / "detections" / "metadata-00.yaml"
            hidden.write_text(
                '"se\\u0061rch": |\n  FROM synthetic_events | where synthetic_hidden=true\n',
                encoding="utf-8",
            )
            subprocess.run(["git", "-C", str(root), "add", str(hidden)], check=True)
            subprocess.run(["git", "-C", str(root), "commit", "-qm", "escaped key"], check=True)
            forms = Path(auxiliary) / "forms.json"
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
            with (
                self.subTest(name=name),
                tempfile.TemporaryDirectory() as temporary,
                tempfile.TemporaryDirectory() as auxiliary,
            ):
                root = Path(temporary)
                self.make_content(root)
                hidden = root / "detections" / "metadata-00.yaml"
                hidden.write_text(content, encoding="utf-8")
                subprocess.run(["git", "-C", str(root), "add", str(hidden)], check=True)
                subprocess.run(["git", "-C", str(root), "commit", "-qm", "node property key"], check=True)
                forms = Path(auxiliary) / "forms.json"
                self.write_forms(forms)
                with self.assertRaisesRegex(AUDIT.AuditError, "mapping key"):
                    AUDIT.audit(root, root / "toolkit", forms, runner=self.successful_runner([]))

    def test_full_audit_rejects_ignored_untracked_yaml_even_when_counts_match(self):
        with tempfile.TemporaryDirectory() as temporary, tempfile.TemporaryDirectory() as auxiliary:
            root = Path(temporary)
            ignored = "detections/ignored.yaml"
            self.make_content(root, standalone_count=44, ignored=ignored)
            (root / ignored).write_text(
                "search: |\n  FROM synthetic_events | where synthetic_ignored=true\n",
                encoding="utf-8",
            )
            forms = Path(auxiliary) / "forms.json"
            self.write_forms(forms)
            with self.assertRaisesRegex(AUDIT.AuditError, "tracked at HEAD"):
                AUDIT.audit(root, root / "toolkit", forms, runner=self.successful_runner([]))

    def test_audit_embeds_fragments_and_emits_aggregate_only_output(self):
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            content = root / "content"
            content.mkdir()
            self.make_content(content, metadata_outside_detections=True)
            forms = root / "forms.json"
            self.write_forms(forms)
            calls = []
            result = AUDIT.audit(content, root / "toolkit", forms, runner=self.successful_runner(calls))

            commit = subprocess.run(
                ["git", "-C", str(content), "rev-parse", "HEAD"],
                check=True, capture_output=True, text=True,
            ).stdout.strip()
            self.assertEqual(result["audited_commit"], commit)
            self.assertEqual(result["counts"], {
                "yaml_documents": 54,
                "search_blocks": 49,
                "standalone_programs": 45,
                "predicate_fragments": 4,
                "form_obligations": 65,
                "analyze_invocations": 114,
                "requirements_invocations": 114,
            })
            self.assertEqual(result["content_syntax_counts"], {
                "predicate_fragment": {"complete": 4, "incomplete": 0},
                "standalone": {"complete": 45, "incomplete": 0},
            })
            self.assertEqual(result["content_semantic_counts"], {
                "predicate_fragment": {"complete": 4, "incomplete": 0},
                "standalone": {"complete": 45, "incomplete": 0},
            })
            self.assertEqual(result["content_status_classes"], {
                "predicate_fragment": {"analyze": {"valid": 4}, "requirements": {"incomplete": 4}},
                "standalone": {"analyze": {"valid": 45}, "requirements": {"incomplete": 45}},
            })
            self.assertEqual(result["content_diagnostic_counts"], {
                "predicate_fragment": {
                    "analyze": {"codes": {"SPL_GENERIC_NOTE": 4}, "categories": {"coverage": 4}},
                    "requirements": {"codes": {"SPL_GENERIC_GAP": 4}, "categories": {"semantics": 4}},
                },
                "standalone": {
                    "analyze": {"codes": {"SPL_GENERIC_NOTE": 45}, "categories": {"coverage": 45}},
                    "requirements": {"codes": {"SPL_GENERIC_GAP": 45}, "categories": {"semantics": 45}},
                },
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
            for document in AUDIT.extract_yaml_documents(content):
                for protected in document.protected_values:
                    if len(protected) >= 4:
                        self.assertNotIn(protected, rendered)
            self.assertNotIn(str(content), rendered)
            self.assertNotIn(str(content.resolve()), rendered)
            self.assertEqual(set(result), {
                "audited_commit", "counts", "content_syntax_counts", "content_semantic_counts",
                "content_status_classes", "content_diagnostic_counts", "form_ids", "status_classes", "diagnostic_counts",
            })

    def test_content_aggregates_distinguish_valid_incomplete_and_invalid(self):
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            content = root / "content"
            content.mkdir()
            self.make_content(content)
            forms = root / "forms.json"
            self.write_forms(forms)
            classes = [
                AUDIT.classify_query(document.query)
                for document in AUDIT.extract_yaml_documents(content)
                if document.query
            ]
            standalone = [index for index, classification in enumerate(classes) if classification == "standalone"]
            calls = []
            successful = self.successful_runner(calls)

            def mixed_runner(args, **kwargs):
                query_index = len(calls) // 2
                completed = successful(args, **kwargs)
                if query_index >= len(classes):
                    return completed
                classification = classes[query_index]
                status = "valid"
                if query_index == standalone[0]:
                    status = "incomplete"
                elif query_index == standalone[1]:
                    status = "invalid"
                code = {"incomplete": "SPL_GENERIC_INCOMPLETE", "invalid": "SPL_GENERIC_INVALID"}.get(status)
                diagnostics = [{"code": code, "category": "generic_gap"}] if code else []
                if args[1] == "analyze":
                    report = {
                        "status": status,
                        "coverage": {"syntax_complete": True, "semantic_complete": status != "incomplete"},
                        "diagnostics": diagnostics,
                    }
                else:
                    report = {
                        "query_status": status,
                        "coverage": {"complete": status == "valid"},
                        "diagnostics": diagnostics,
                    }
                return subprocess.CompletedProcess(args, AUDIT.STATUS_EXITS[status], json.dumps(report), "")

            result = AUDIT.audit(content, root / "toolkit", forms, runner=mixed_runner)
            self.assertEqual(result["content_semantic_counts"], {
                "predicate_fragment": {"complete": 4, "incomplete": 0},
                "standalone": {"complete": 44, "incomplete": 1},
            })
            self.assertEqual(result["content_status_classes"], {
                "predicate_fragment": {"analyze": {"valid": 4}, "requirements": {"valid": 4}},
                "standalone": {
                    "analyze": {"valid": 43, "incomplete": 1, "invalid": 1},
                    "requirements": {"valid": 43, "incomplete": 1, "invalid": 1},
                },
            })
            self.assertEqual(result["content_diagnostic_counts"]["standalone"], {
                "analyze": {
                    "codes": {"SPL_GENERIC_INCOMPLETE": 1, "SPL_GENERIC_INVALID": 1},
                    "categories": {"generic_gap": 2},
                },
                "requirements": {
                    "codes": {"SPL_GENERIC_INCOMPLETE": 1, "SPL_GENERIC_INVALID": 1},
                    "categories": {"generic_gap": 2},
                },
            })
            self.assertEqual(result["content_diagnostic_counts"]["predicate_fragment"], {
                "analyze": {"codes": {}, "categories": {}},
                "requirements": {"codes": {}, "categories": {}},
            })
            rendered = json.dumps(result, sort_keys=True)
            for document in AUDIT.extract_yaml_documents(content):
                for protected in document.protected_values:
                    if len(protected) >= 4:
                        self.assertNotIn(protected, rendered)
            self.assertNotIn(str(content), rendered)
            self.assertNotIn(str(content.resolve()), rendered)

    def test_external_syntax_gate_rejects_content_without_leaking_it(self):
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            content = root / "content"
            content.mkdir()
            self.make_content(content)
            protected_query = "FROM synthetic_events | where ("
            target = content / "detections" / "generic-00.yaml"
            target.write_text(
                "title: Private invalid syntax detection\nsearch: |-\n  " + protected_query + "\n",
                encoding="utf-8",
            )
            subprocess.run(["git", "-C", str(content), "add", str(target)], check=True)
            subprocess.run(["git", "-C", str(content), "commit", "-qm", "invalid syntax"], check=True)
            forms = root / "forms.json"
            self.write_forms(forms)

            with self.assertRaisesRegex(AUDIT.AuditError, "external content syntax coverage is incomplete") as raised:
                AUDIT.audit(
                    content,
                    root / "toolkit",
                    forms,
                    runner=self.successful_runner([], {protected_query}),
                )
            self.assertNotIn(protected_query, str(raised.exception))
            self.assertNotIn("generic-00.yaml", str(raised.exception))
            self.assertNotIn("Private invalid syntax detection", str(raised.exception))

    def test_generic_form_syntax_boundary_does_not_fail_external_gate(self):
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            content = root / "content"
            content.mkdir()
            self.make_content(content)
            forms = root / "forms.json"
            self.write_forms(forms)
            protected_form = next(
                query for _, query in AUDIT.load_forms(forms)
                if "| eval " in query
            )

            result = AUDIT.audit(
                content,
                root / "toolkit",
                forms,
                runner=self.successful_runner([], {protected_form}),
            )
            self.assertEqual(result["content_syntax_counts"], {
                "predicate_fragment": {"complete": 4, "incomplete": 0},
                "standalone": {"complete": 45, "incomplete": 0},
            })
            self.assertEqual(result["status_classes"]["analyze"]["invalid"], 1)
            self.assertNotIn(protected_form, json.dumps(result, sort_keys=True))

    def test_subprocess_status_exits_are_checked_without_leaking_output(self):
        valid = {"status": "invalid", "coverage": {"syntax_complete": False, "semantic_complete": False}, "diagnostics": []}
        completed = subprocess.CompletedProcess([], 1, json.dumps(valid), "protected stderr")
        self.assertEqual(AUDIT.parse_toolkit_result("analyze", completed)[0], "invalid")

        failed = subprocess.CompletedProcess([], 2, "protected search text", "protected path")
        with self.assertRaises(AUDIT.AuditError) as raised:
            AUDIT.parse_toolkit_result("analyze", failed)
        self.assertNotIn("protected", str(raised.exception))

    def test_analyze_requires_boolean_semantic_coverage(self):
        for missing in (None, "yes"):
            with self.subTest(missing=missing):
                report = {
                    "status": "valid",
                    "coverage": {"syntax_complete": True, "semantic_complete": missing},
                    "diagnostics": [],
                }
                completed = subprocess.CompletedProcess([], 0, json.dumps(report), "")
                with self.assertRaisesRegex(AUDIT.AuditError, "semantic coverage"):
                    AUDIT.parse_toolkit_result("analyze", completed)

    def test_output_leak_rejection(self):
        with self.assertRaisesRegex(AUDIT.AuditError, "protected input"):
            AUDIT.reject_protected_output(
                {"form_ids": ["M11.layout.standalone"], "leak": "Private detection label"},
                {"Private detection label"},
            )

    def test_audit_rejects_absolute_repository_path_in_aggregate_output(self):
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            content = root / "content"
            content.mkdir()
            self.make_content(content)
            alias = root / "content-alias"
            alias.symlink_to(content, target_is_directory=True)
            forms = root / "forms.json"
            self.write_forms(forms)
            for path_kind, leaked in (("supplied", str(alias)), ("resolved", str(alias.resolve()))):
                with self.subTest(path_kind=path_kind):
                    calls = []
                    successful = self.successful_runner(calls)

                    def leaking_runner(args, **kwargs):
                        completed = successful(args, **kwargs)
                        if len(calls) == 1:
                            report = json.loads(completed.stdout)
                            report["diagnostics"][0]["code"] = leaked
                            return subprocess.CompletedProcess(args, completed.returncode, json.dumps(report), "")
                        return completed

                    with self.assertRaisesRegex(AUDIT.AuditError, "protected input"):
                        AUDIT.audit(alias, root / "toolkit", forms, runner=leaking_runner)

    def test_audit_rejects_git_root_when_content_root_is_subdirectory(self):
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            repository = root / "content"
            repository.mkdir()
            self.make_content(repository)
            forms = root / "forms.json"
            self.write_forms(forms)
            for path_kind, leaked in (("supplied", str(repository)), ("resolved", str(repository.resolve()))):
                with self.subTest(path_kind=path_kind):
                    calls = []
                    successful = self.successful_runner(calls)

                    def leaking_runner(args, **kwargs):
                        completed = successful(args, **kwargs)
                        if len(calls) == 1:
                            report = json.loads(completed.stdout)
                            report["diagnostics"][0]["code"] = leaked
                            return subprocess.CompletedProcess(args, completed.returncode, json.dumps(report), "")
                        return completed

                    with self.assertRaisesRegex(AUDIT.AuditError, "protected input"):
                        AUDIT.audit(repository / "detections", root / "toolkit", forms, runner=leaking_runner)

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
