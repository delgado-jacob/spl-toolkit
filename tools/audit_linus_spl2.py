#!/usr/bin/env python3
"""Audit an exact Linus SPL2 snapshot without emitting repository content."""

from __future__ import annotations

import argparse
from collections import Counter
import importlib.util
import json
from pathlib import Path
import re
import subprocess
import sys
from typing import NamedTuple


EXPECTED_INVENTORY = {
    "yaml_documents": 54,
    "search_blocks": 49,
    "standalone_programs": 45,
    "predicate_fragments": 4,
}
EXACT_LINUS_COMMIT = "d4db9bae4adba00bc59d9becbf4014908e7a3ef5"
EXPECTED_CONTENT_ACCEPTANCE = {
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
}
EXPECTED_CONTENT_DIAGNOSTIC_COUNTS = {
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
}
STATUS_EXITS = {"valid": 0, "invalid": 1, "incomplete": 3}
TOP_LEVEL_SEARCH = re.compile(r"^search\s*:\s*(.*)$")
AMBIGUOUS_SEARCH_KEY = re.compile(
    r"^(?:(?:[ ]+|[ ]*-[ ]+)search|[ ]*(?:-[ ]+)?(?:\"search\"|'search'))[ ]*:"
)
QUOTED_MAPPING_KEY = re.compile(
    r'''^[ ]*(?:-[ ]+)?(?:"(?:\\.|[^"\\])*"|'(?:''|[^'])*')[ ]*:'''
)
TAGGED_MAPPING_KEY = re.compile(r"^[ ]*(?:-[ ]+)?![^ \t]+[ ]+[^:]+[ ]*:")
EXPLICIT_MAPPING_KEY = re.compile(r"^[ ]*(?:-[ ]+)?\?(?:[ ]|$)")
NODE_PROPERTY_MAPPING_KEY = re.compile(
    r"^[ ]*(?:-[ ]+)?(?:&[^ \t]+[ \t]+[^:]+|\*[^ \t:]+)[ ]*:"
)
TOP_LEVEL_SCALAR = re.compile(r"^([A-Za-z_][A-Za-z0-9_-]*)\s*:\s*(.+?)\s*$")
STANDALONE = re.compile(r"^(?:FROM|SELECT)\b", re.IGNORECASE)


class AuditError(ValueError):
    pass


class YAMLDocument(NamedTuple):
    relative_path: str
    query: str | None
    protected_values: frozenset[str]


def _split_yaml_documents(text: str) -> list[list[str]]:
    lines = text.splitlines()
    documents = []
    current = []
    for line in lines:
        if "\t" in line[:len(line) - len(line.lstrip())]:
            raise AuditError("ambiguous YAML tab indentation")
        if line in {"---", "..."}:
            if current or line == "...":
                documents.append(current)
                current = []
            continue
        current.append(line)
    if current or not documents:
        documents.append(current)
    return documents


def _unquote_scalar(value: str) -> str | None:
    value = value.strip()
    if not value or value[0] in "|>&*!{[":
        return None
    if len(value) >= 2 and value[0] == value[-1] and value[0] in "\"'":
        value = value[1:-1]
    return value if len(value) >= 4 else None


def _extract_document(lines: list[str], relative_path: str) -> YAMLDocument:
    query = None
    protected = {relative_path, Path(relative_path).name, Path(relative_path).stem}
    index = 0
    while index < len(lines):
        line = lines[index]
        if any(
            pattern.match(line)
            for pattern in (
                QUOTED_MAPPING_KEY,
                TAGGED_MAPPING_KEY,
                EXPLICIT_MAPPING_KEY,
                NODE_PROPERTY_MAPPING_KEY,
            )
        ):
            raise AuditError("unsupported YAML mapping key syntax")
        if AMBIGUOUS_SEARCH_KEY.match(line):
            raise AuditError("ambiguous, nested, or quoted YAML search key")
        match = TOP_LEVEL_SEARCH.match(line)
        if match:
            if query is not None:
                raise AuditError("duplicate top-level YAML search key")
            indicator = match.group(1)
            if indicator not in {"|", "|-", "|+"}:
                raise AuditError("search must be a top-level literal block scalar")
            block = []
            index += 1
            while index < len(lines):
                candidate = lines[index]
                if candidate and not candidate.startswith(" "):
                    break
                block.append(candidate)
                index += 1
            nonempty = [len(value) - len(value.lstrip(" ")) for value in block if value.strip()]
            if not nonempty or min(nonempty) == 0:
                raise AuditError("search literal block scalar is empty or ambiguously indented")
            indentation = min(nonempty)
            if any(value.strip() and len(value) - len(value.lstrip(" ")) < indentation for value in block):
                raise AuditError("ambiguous YAML block indentation")
            query = "\n".join(value[indentation:] if value else "" for value in block).rstrip("\n")
            if not query.strip():
                raise AuditError("search literal block scalar is empty")
            continue
        scalar = TOP_LEVEL_SCALAR.match(line)
        if scalar and scalar.group(1) != "search":
            value = _unquote_scalar(scalar.group(2))
            if value:
                protected.add(value)
        index += 1
    if query:
        protected.add(query)
    return YAMLDocument(relative_path, query, frozenset(protected))


def _yaml_paths(root: Path) -> list[Path]:
    root = Path(root)
    if not root.is_dir():
        raise AuditError("content root is not a directory")
    paths = sorted({*root.rglob("*.yaml"), *root.rglob("*.yml")})
    if not paths:
        raise AuditError("content root contains no YAML documents")
    if any(path.is_symlink() or not path.is_file() for path in paths):
        raise AuditError("content tree contains an unsupported YAML entry")
    return paths


def _extract_yaml_inputs(root: Path, inputs: list[tuple[Path, str]]) -> list[YAMLDocument]:
    documents = []
    for path, text in inputs:
        relative = path.relative_to(root).as_posix()
        for lines in _split_yaml_documents(text):
            documents.append(_extract_document(lines, relative))
    return documents


def extract_yaml_documents(root: Path) -> list[YAMLDocument]:
    root = Path(root)
    return _extract_yaml_inputs(
        root,
        [(path, path.read_text(encoding="utf-8")) for path in _yaml_paths(root)],
    )


def _git_root(root: Path) -> Path:
    top = subprocess.run(
        ["git", "-C", str(root), "rev-parse", "--show-toplevel"],
        capture_output=True, text=True, check=False,
    )
    if top.returncode != 0:
        raise AuditError("audited YAML inputs are not tracked at HEAD")
    return Path(top.stdout.strip()).resolve()


def extract_tracked_yaml_documents(root: Path, commit: str) -> list[YAMLDocument]:
    root = Path(root).resolve()
    git_root = _git_root(root)
    try:
        root.relative_to(git_root)
    except ValueError as error:
        raise AuditError("audited YAML inputs are not tracked at HEAD") from error
    inputs = []
    for path in _yaml_paths(root):
        relative = path.resolve().relative_to(git_root).as_posix()
        tracked = subprocess.run(
            ["git", "-C", str(git_root), "ls-files", "--error-unmatch", "--", relative],
            capture_output=True, check=False,
        )
        if tracked.returncode != 0:
            raise AuditError("every audited YAML input must be tracked at HEAD")
        blob = subprocess.run(
            ["git", "-C", str(git_root), "show", f"{commit}:{relative}"],
            capture_output=True, check=False,
        )
        if blob.returncode != 0 or path.read_bytes() != blob.stdout:
            raise AuditError("audited YAML bytes differ from the exact committed snapshot")
        try:
            text = blob.stdout.decode("utf-8")
        except UnicodeDecodeError as error:
            raise AuditError("audited YAML input is not UTF-8") from error
        inputs.append((path, text))
    return _extract_yaml_inputs(root, inputs)


def classify_query(query: str) -> str:
    stripped = query.strip()
    if STANDALONE.match(stripped):
        return "standalone"
    if not stripped or "|" in stripped or ";" in stripped:
        raise AuditError("predicate fragment is not safe for neutral local embedding")
    return "predicate_fragment"


def executable_query(query: str, classification: str) -> str:
    if classification == "standalone":
        return query
    if classification == "predicate_fragment":
        return f"FROM synthetic_events | where ({query})"
    raise AuditError("unknown query classification")


def load_forms(path: Path) -> list[tuple[str, str]]:
    try:
        payload = json.loads(Path(path).read_text(encoding="utf-8"))
    except (OSError, json.JSONDecodeError) as error:
        raise AuditError("form matrix could not be loaded") from error
    spec = importlib.util.spec_from_file_location(
        "linus_spl2_form_contract", Path(__file__).with_name("check_spl2_corpus.py"),
    )
    if spec is None or spec.loader is None:
        raise AuditError("approved form matrix validator is unavailable")
    validator = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(validator)
    try:
        validator.audit_linus_forms(payload, {"linus_m11"})
    except (KeyError, TypeError, ValueError) as error:
        raise AuditError("approved form matrix validation failed") from error
    return [(obligation["id"], obligation["query"]) for obligation in payload["obligations"]]


def exact_commit(content_root: Path) -> str:
    completed = subprocess.run(
        ["git", "-C", str(content_root), "rev-parse", "HEAD"],
        capture_output=True, text=True, check=False,
    )
    commit = completed.stdout.strip()
    if completed.returncode != 0 or not re.fullmatch(r"[0-9a-f]{40}", commit):
        raise AuditError("content root does not resolve an exact Git commit")
    status = subprocess.run(
        ["git", "-C", str(content_root), "status", "--porcelain=v1", "--untracked-files=all", "--", "."],
        capture_output=True, text=True, check=False,
    )
    if status.returncode != 0 or status.stdout:
        raise AuditError("content root is not an exact committed snapshot")
    return commit


def _requirements_status(report: dict) -> str | None:
    status = report.get("query_status")
    coverage = report.get("coverage")
    if status in {"invalid", "incomplete"}:
        return status
    if status == "valid" and isinstance(coverage, dict):
        if coverage.get("complete") is True:
            return "valid"
        if coverage.get("complete") is False:
            return "incomplete"
    return None


def parse_toolkit_result(operation: str, completed: subprocess.CompletedProcess[str]) -> tuple[str, list[dict], bool | None, bool | None]:
    if operation not in {"analyze", "requirements"}:
        raise AuditError("unknown toolkit operation")
    try:
        report = json.loads(completed.stdout)
    except (TypeError, json.JSONDecodeError) as error:
        raise AuditError(f"toolkit {operation} returned invalid JSON") from error
    if not isinstance(report, dict):
        raise AuditError(f"toolkit {operation} returned an invalid report")
    status = report.get("status") if operation == "analyze" else _requirements_status(report)
    if status not in STATUS_EXITS or completed.returncode != STATUS_EXITS[status]:
        raise AuditError(f"toolkit {operation} returned an unexpected status or exit")
    diagnostics = report.get("diagnostics", [])
    if not isinstance(diagnostics, list) or not all(isinstance(item, dict) for item in diagnostics):
        raise AuditError(f"toolkit {operation} returned invalid diagnostics")
    syntax_complete = None
    semantic_complete = None
    if operation == "analyze":
        coverage = report.get("coverage")
        if not isinstance(coverage, dict) or not isinstance(coverage.get("syntax_complete"), bool):
            raise AuditError("toolkit analyze returned invalid syntax coverage")
        if not isinstance(coverage.get("semantic_complete"), bool):
            raise AuditError("toolkit analyze returned invalid semantic coverage")
        syntax_complete = coverage["syntax_complete"]
        semantic_complete = coverage["semantic_complete"]
    return status, diagnostics, syntax_complete, semantic_complete


def _count_diagnostics(diagnostics: list[dict], codes: Counter, categories: Counter) -> None:
    for diagnostic in diagnostics:
        code, category = diagnostic.get("code"), diagnostic.get("category")
        if not isinstance(code, str) or not code or not isinstance(category, str) or not category:
            raise AuditError("toolkit diagnostic lacks aggregate code or category")
        codes[code] += 1
        categories[category] += 1


def _strings(value):
    if isinstance(value, dict):
        for key, item in value.items():
            yield str(key)
            yield from _strings(item)
    elif isinstance(value, list):
        for item in value:
            yield from _strings(item)
    elif isinstance(value, str):
        yield value


def reject_protected_output(output: dict, protected_values: set[str]) -> None:
    emitted = list(_strings(output))
    for protected in protected_values:
        if len(protected) < 4:
            continue
        if any(protected in value for value in emitted):
            raise AuditError("aggregate output would expose protected input")


def validate_standalone_attribution(status: str, semantic_complete: bool, diagnostics: list[dict]) -> None:
    """Require complete semantics and attach definite unavailable fields to invalid reports."""
    pairs = set()
    for diagnostic in diagnostics:
        if not isinstance(diagnostic, dict):
            raise AuditError("toolkit diagnostic lacks aggregate code or category")
        code, category = diagnostic.get("code"), diagnostic.get("category")
        if not isinstance(code, str) or not code or not isinstance(category, str) or not category:
            raise AuditError("toolkit diagnostic lacks aggregate code or category")
        pairs.add((code, category))
    unavailable_pair = ("SPL_UNAVAILABLE_FIELD", "unavailable_field")
    selected = {pair for pair in pairs if pair[0] in {
        "SPL_AMBIGUOUS_FIELD", "SPL_UNAVAILABLE_FIELD", "SPL_UNSUPPORTED_SEMANTICS",
    }}
    if selected - {unavailable_pair}:
        raise AuditError("exact-ref standalone field attribution differs from the confirmed baseline")
    unavailable = unavailable_pair in selected
    accepted = {
        "valid": semantic_complete and not unavailable,
        "invalid": semantic_complete and unavailable,
    }
    if not accepted.get(status, False):
        raise AuditError("exact-ref standalone field attribution differs from the confirmed baseline")


def audit(
    content_root: Path,
    toolkit_bin: Path,
    forms_path: Path,
    *,
    runner=subprocess.run,
    enforce_field_attribution: bool = False,
) -> dict:
    scan_root = Path(content_root)
    commit = exact_commit(scan_root)
    documents = extract_tracked_yaml_documents(scan_root, commit)
    content_queries = []
    git_root = _git_root(scan_root)
    supplied_git_root = scan_root
    for _ in scan_root.resolve().relative_to(git_root).parts:
        supplied_git_root = supplied_git_root.parent
    protected = {str(scan_root), str(scan_root.resolve()), str(supplied_git_root), str(git_root)}
    classifications = Counter()
    for document in documents:
        protected.update(document.protected_values)
        if document.query is None:
            continue
        classification = classify_query(document.query)
        classifications[classification] += 1
        content_queries.append((classification, executable_query(document.query, classification)))

    forms = load_forms(forms_path)
    protected.update(query for _, query in forms)
    observed = {
        "yaml_documents": len(documents),
        "search_blocks": len(content_queries),
        "standalone_programs": classifications["standalone"],
        "predicate_fragments": classifications["predicate_fragment"],
    }
    if observed != EXPECTED_INVENTORY:
        raise AuditError("content inventory differs from the approved aggregate baseline")

    status_counts = {operation: Counter() for operation in ("analyze", "requirements")}
    diagnostic_codes = {operation: Counter() for operation in ("analyze", "requirements")}
    diagnostic_categories = {operation: Counter() for operation in ("analyze", "requirements")}
    content_classes = ("predicate_fragment", "standalone")
    content_syntax_counts = {
        classification: Counter({"complete": 0, "incomplete": 0})
        for classification in content_classes
    }
    content_semantic_counts = {
        classification: Counter({"complete": 0, "incomplete": 0})
        for classification in content_classes
    }
    content_status_classes = {
        classification: {operation: Counter() for operation in ("analyze", "requirements")}
        for classification in content_classes
    }
    content_diagnostic_codes = {
        classification: {operation: Counter() for operation in ("analyze", "requirements")}
        for classification in content_classes
    }
    content_diagnostic_categories = {
        classification: {operation: Counter() for operation in ("analyze", "requirements")}
        for classification in content_classes
    }
    queries = [
        ("content", classification, query)
        for classification, query in content_queries
    ] + [
        ("form", "", query)
        for _, query in forms
    ]
    toolkit = str(Path(toolkit_bin))
    for origin, classification, query in queries:
        for operation in ("analyze", "requirements"):
            args = [
                toolkit, operation,
                "--language", "spl2",
                "--profile", "splunkd",
                "--compatibility-version", "current",
                "--query", query,
                "--format", "json",
            ]
            try:
                completed = runner(args, capture_output=True, text=True, check=False)
            except (OSError, subprocess.SubprocessError) as error:
                raise AuditError(f"toolkit {operation} could not be executed") from error
            status, diagnostics, syntax_complete, semantic_complete = parse_toolkit_result(operation, completed)
            if enforce_field_attribution and origin == "content" and classification == "standalone" and operation == "analyze":
                validate_standalone_attribution(status, semantic_complete, diagnostics)
            status_counts[operation][status] += 1
            if origin == "content":
                content_status_classes[classification][operation][status] += 1
                _count_diagnostics(
                    diagnostics,
                    content_diagnostic_codes[classification][operation],
                    content_diagnostic_categories[classification][operation],
                )
                if operation == "analyze":
                    content_syntax_counts[classification]["complete" if syntax_complete else "incomplete"] += 1
                    content_semantic_counts[classification]["complete" if semantic_complete else "incomplete"] += 1
            _count_diagnostics(
                diagnostics,
                diagnostic_codes[operation],
                diagnostic_categories[operation],
            )

    if any(counts["incomplete"] for counts in content_syntax_counts.values()):
        raise AuditError("external content syntax coverage is incomplete")

    if exact_commit(scan_root) != commit:
        raise AuditError("content root changed during the audit")
    result = {
        "audited_commit": commit,
        "counts": observed | {
            "form_obligations": len(forms),
            "analyze_invocations": len(queries),
            "requirements_invocations": len(queries),
        },
        "content_syntax_counts": {
            classification: dict(sorted(counts.items()))
            for classification, counts in sorted(content_syntax_counts.items())
        },
        "content_semantic_counts": {
            classification: dict(sorted(counts.items()))
            for classification, counts in sorted(content_semantic_counts.items())
        },
        "content_status_classes": {
            classification: {
                operation: dict(sorted(counts.items()))
                for operation, counts in sorted(operations.items())
            }
            for classification, operations in sorted(content_status_classes.items())
        },
        "content_diagnostic_counts": {
            classification: {
                operation: {
                    "codes": dict(sorted(content_diagnostic_codes[classification][operation].items())),
                    "categories": dict(sorted(content_diagnostic_categories[classification][operation].items())),
                }
                for operation in ("analyze", "requirements")
            }
            for classification in content_classes
        },
        "form_ids": sorted(identity for identity, _ in forms),
        "status_classes": {
            operation: dict(sorted(status_counts[operation].items()))
            for operation in ("analyze", "requirements")
        },
        "diagnostic_counts": {
            operation: {
                "codes": dict(sorted(diagnostic_codes[operation].items())),
                "categories": dict(sorted(diagnostic_categories[operation].items())),
            }
            for operation in ("analyze", "requirements")
        },
    }
    reject_protected_output(result, protected)
    return result


def validate_exact_ref_acceptance(result: dict) -> None:
    """Gate the confirmed private-content aggregate without exposing source details."""
    failure = "exact-ref external content acceptance aggregate differs from the confirmed baseline"
    expected_counts = EXPECTED_INVENTORY | {
        "form_obligations": 65,
        "analyze_invocations": 114,
        "requirements_invocations": 114,
    }
    if result.get("audited_commit") != EXACT_LINUS_COMMIT or result.get("counts") != expected_counts:
        raise AuditError(failure)
    if not isinstance(result.get("form_ids"), list) or len(result["form_ids"]) != 65:
        raise AuditError(failure)
    for key, expected in EXPECTED_CONTENT_ACCEPTANCE.items():
        if result.get(key) != expected:
            raise AuditError(failure)
    if result.get("content_diagnostic_counts") != EXPECTED_CONTENT_DIAGNOSTIC_COUNTS:
        raise AuditError(failure)


def main(argv=None) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--content-root", required=True, type=Path)
    parser.add_argument("--toolkit-bin", required=True, type=Path)
    parser.add_argument("--forms", required=True, type=Path)
    options = parser.parse_args(argv)
    try:
        result = audit(options.content_root, options.toolkit_bin, options.forms, enforce_field_attribution=True)
        validate_exact_ref_acceptance(result)
    except (AuditError, OSError, UnicodeError):
        print("Linus SPL2 audit failed: aggregate audit contract was not satisfied", file=sys.stderr)
        return 1
    print(json.dumps(result, sort_keys=True, separators=(",", ":")))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
