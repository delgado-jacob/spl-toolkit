"""Canonical requirement values across Go, CLI, HTTP, C, and Python."""

from __future__ import annotations

import copy
from concurrent.futures import ThreadPoolExecutor
import hashlib
import json
import os
from pathlib import Path
import signal
import subprocess
import sys
import tempfile
import threading
import time
from urllib.error import HTTPError
from urllib.request import Request, urlopen

import pytest

from spl_toolkit import SPLMapper
from test_surfaces import cli_path, required_absolute_path, server_url


EXITS = {"valid": 0, "invalid": 1, "incomplete": 3}
SELECTORS = (
    ("language", "--language"),
    ("profile", "--profile"),
    ("version", "--compatibility-version"),
    ("source_id", "--source-id"),
)
RESOURCE_CODE = "SPL_ANALYSIS_RESOURCE_LIMIT"
RESOURCE_MESSAGE = (
    "analysis stopped before parser prediction after reaching the 4,096-unit lexer work limit"
)
RESOURCE_GAP_MESSAGE = (
    "requirement coverage is incomplete because analysis exceeded the 4,096-unit lexer work limit"
)
DEPENDENCY_KEYS = (
    "indexes",
    "sources",
    "source_types",
    "datasets",
    "data_models",
    "lookups",
    "macros",
)
CAPABILITY_REVISIONS = {
    "spl": "sha256:1e6c75800f843931ec517dba27f3baa5af928a8a908d97dd1c62513e2ea24d31",
    "spl2": "sha256:0203cbeec2e0fd484080b1f512582a1c8bdc4bc541fb73ac42c013060695f84c",
}
DENSE_QUERY_DIGESTS = {
    ("spl", 65_536): "sha256:4ef76589e31ca84b778eb3e15f8cd320319dd03746fff5bf7ba21ba66865dace",
    ("spl", 262_144): "sha256:64ec67a0b6bcecae397863ac2b0336f8c59fa39cadb134fd2757b3d68b66e502",
    ("spl2", 65_536): "sha256:c5b1c0ba7bc519bde8181f7b8469f84fb4a04e35e16cc37d027a67a3477fcf3b",
    ("spl2", 262_144): "sha256:ad99be139221b8721e1d2319d4be832e2df915731a1f33e45444320a9115c445",
}
DENSE_OMITTED_OFFSETS = {"spl": 7026, "spl2": 7564}
PORTABLE_RESOURCE_QUERY = "a " * 2048 + "a"
GO_HELPER = r'''
package main

import (
	"encoding/json"
	"os"
	"sync"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
)

type reports struct {
	Analysis     *analysis.Result         `json:"analysis"`
	Requirements *analysis.RequirementSet `json:"requirements"`
}

type request struct {
	Documents          []analysis.QueryDocument `json:"documents"`
	ConcurrentDocuments []analysis.QueryDocument `json:"concurrent_documents"`
}

type response struct {
	Reports    []reports `json:"reports"`
	Concurrent []reports `json:"concurrent"`
}

func analyze(document analysis.QueryDocument) reports {
	analyzed, err := analysis.Analyze(document)
	if err != nil {
		panic(err)
	}
	requirements, err := analysis.Requirements(document)
	if err != nil {
		panic(err)
	}
	return reports{Analysis: analyzed, Requirements: requirements}
}

func main() {
	var input request
	if err := json.NewDecoder(os.Stdin).Decode(&input); err != nil {
		panic(err)
	}
	output := response{
		Reports: make([]reports, len(input.Documents)),
		Concurrent: make([]reports, len(input.ConcurrentDocuments)),
	}
	for index, document := range input.Documents {
		output.Reports[index] = analyze(document)
	}
	var wait sync.WaitGroup
	for index, document := range input.ConcurrentDocuments {
		wait.Add(1)
		go func(index int, document analysis.QueryDocument) {
			defer wait.Done()
			output.Concurrent[index] = analyze(document)
		}(index, document)
	}
	wait.Wait()
	if err := json.NewEncoder(os.Stdout).Encode(output); err != nil {
		panic(err)
	}
}
'''


def open_mapper() -> SPLMapper:
    options = {}
    if "SPL_NATIVE_LIBRARY" in os.environ:
        options["library_path"] = str(required_absolute_path("SPL_NATIVE_LIBRARY"))
    return SPLMapper(**options)


def dense_query(size: int, language: str) -> str:
    pattern = "FROM main | fields a* | " if language == "spl2" else "fields a* | "
    return (pattern * (size // len(pattern) + 1))[:size]


def normalized_document(document: dict) -> dict:
    return {
        "text": document["text"],
        "language": document.get("language") or "spl",
        "profile": document.get("profile") or "splunkd",
        "version": document.get("version") or "current",
        "source_id": document.get("source_id") or "",
    }


def compact_bytes(value: dict) -> bytes:
    return json.dumps(value, ensure_ascii=False, separators=(",", ":")).encode("utf-8")


def cli_arguments(cli: Path, operation: str, document: dict) -> list[str]:
    args = [str(cli), operation, "--format", "json", "--query", document["text"]]
    for key, flag in SELECTORS:
        if key in document:
            args.extend([flag, document[key]])
    return args


def windows_command_utf16_units(args: list[str]) -> int:
    command = subprocess.list2cmdline(args)
    return len(command.encode("utf-16-le")) // 2 + 1


def run_cli(cli: Path, operation: str, document: dict) -> tuple[int, dict, bytes]:
    args = cli_arguments(cli, operation, document)
    command_units = windows_command_utf16_units(args)
    assert command_units < 32_767, (
        operation,
        document.get("language") or "spl",
        len(document["text"].encode("utf-8")),
        command_units,
    )
    completed = subprocess.run(args, capture_output=True, check=False, timeout=20)
    assert completed.stderr == b"", (operation, document.get("source_id"), completed.stderr)
    assert completed.stdout, (operation, document.get("source_id"), completed.returncode)
    return completed.returncode, json.loads(completed.stdout), completed.stdout.rstrip(b"\n")


def expected_requirements_exit(requirements: dict) -> int:
    status = requirements.get("query_status")
    coverage = requirements.get("coverage")
    if status == "invalid":
        return 1
    if status == "incomplete":
        return 3
    if status == "valid" and isinstance(coverage, dict):
        if coverage.get("complete") is False:
            return 3
        if coverage.get("complete") is True:
            return 0
    return 2


def post_raw(server: str, operation: str, payload: bytes) -> tuple[int, dict, bytes]:
    request = Request(
        server + "/query/" + operation,
        data=payload,
        headers={"Content-Type": "application/json"},
        method="POST",
    )
    try:
        with urlopen(request, timeout=20) as response:
            body = response.read()
            return response.status, json.loads(body), body.rstrip(b"\n")
    except HTTPError as error:
        with error:
            body = error.read()
        return error.code, json.loads(body), body.rstrip(b"\n")


def post_document(server: str, operation: str, document: dict) -> tuple[int, dict, bytes]:
    return post_raw(server, operation, compact_bytes(document))


def call_raw_c(mapper: SPLMapper, operation: str, document: dict) -> dict:
    pointer = getattr(mapper._lib, "spl_mapper_" + operation)(
        mapper._mapper_id, compact_bytes(document)
    )
    try:
        assert pointer and not pointer.contents.error and pointer.contents.result
        return json.loads(pointer.contents.result)
    finally:
        mapper._lib.spl_result_free(pointer)


def call_python(mapper: SPLMapper, operation: str, document: dict) -> dict:
    options = {key: value for key, value in document.items() if key != "text"}
    return getattr(mapper, operation)(document["text"], **options)


def _helper_context(request: dict) -> str:
    documents = request.get("documents") or []
    if documents:
        languages = sorted({document.get("language") or "spl" for document in documents})
        sizes = sorted({len(document["text"].encode("utf-8")) for document in documents})
    else:
        languages = [request.get("language", "unknown")]
        sizes = [request.get("size", "unknown")]
    return (
        f"mode={request.get('mode')} operations=['analyze_query', 'requirements_query'] "
        f"languages={languages} sizes={sizes}"
    )


def _reap_helper(
    child: subprocess.Popen,
    terminate_timeout: float,
    kill_timeout: float,
    stdout: str | bytes | None = "",
    stderr: str | bytes | None = "",
) -> tuple[bool, bool, bool, str, str]:
    terminated = False
    killed = False
    if child.poll() is None:
        child.terminate()
        terminated = True
    try:
        final_stdout, final_stderr = child.communicate(timeout=terminate_timeout)
        stdout = _merge_process_output(stdout, final_stdout)
        stderr = _merge_process_output(stderr, final_stderr)
    except subprocess.TimeoutExpired as error:
        stdout = _merge_process_output(stdout, error.stdout)
        stderr = _merge_process_output(stderr, error.stderr)
        if child.poll() is None:
            child.kill()
            killed = True
        try:
            final_stdout, final_stderr = child.communicate(timeout=kill_timeout)
            stdout = _merge_process_output(stdout, final_stdout)
            stderr = _merge_process_output(stderr, final_stderr)
        except subprocess.TimeoutExpired as kill_error:
            stdout = _merge_process_output(stdout, kill_error.stdout)
            stderr = _merge_process_output(stderr, kill_error.stderr)
            if child.poll() is None:
                child.kill()
            try:
                child.wait(timeout=kill_timeout)
            except subprocess.TimeoutExpired:
                pass
            for pipe in (child.stdin, child.stdout, child.stderr):
                if pipe is not None:
                    pipe.close()
    return terminated, killed, child.poll() is not None, stdout, stderr


def _merge_process_output(
    captured: str | bytes | None, completed: str | bytes | None
) -> str:
    def text(value: str | bytes | None) -> str:
        if value is None:
            return ""
        if isinstance(value, bytes):
            return value.decode("utf-8", errors="replace")
        return value

    captured_text = text(captured)
    completed_text = text(completed)
    if not captured_text or completed_text.startswith(captured_text):
        return completed_text
    if not completed_text:
        return captured_text
    overlap = min(len(captured_text), len(completed_text))
    while overlap and not captured_text.endswith(completed_text[:overlap]):
        overlap -= 1
    return captured_text + completed_text[overlap:]


def _wait_for_helper_ready(
    child: subprocess.Popen, ready_path: Path, timeout: float
) -> bool:
    deadline = time.monotonic() + timeout
    while time.monotonic() < deadline:
        if ready_path.is_file():
            return True
        if child.poll() is not None:
            return False
        time.sleep(0.01)
    return ready_path.is_file()


def run_native_helper(
    request: dict,
    *,
    timeout: float,
    terminate_timeout: float = 5,
    kill_timeout: float = 5,
    ready_timeout: float = 20,
) -> dict:
    args = [sys.executable, str(Path(__file__).resolve()), "--requirements-native-helper"]
    child = subprocess.Popen(
        args,
        stdin=subprocess.PIPE,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        text=True,
        env=os.environ.copy(),
    )
    stdout = ""
    stderr = ""
    ready = False
    ready_directory = (
        tempfile.TemporaryDirectory(prefix="spl-native-helper-ready-")
        if request.get("mode") == "hang"
        else None
    )
    try:
        try:
            helper_request = dict(request)
            if ready_directory is not None:
                ready_path = Path(ready_directory.name) / "ready"
                helper_request["_ready_path"] = str(ready_path)
                assert child.stdin is not None
                child.stdin.write(json.dumps(helper_request, ensure_ascii=False))
                child.stdin.close()
                child.stdin = None
                ready = _wait_for_helper_ready(child, ready_path, ready_timeout)
                if not ready:
                    terminated, killed, reaped, stdout, stderr = _reap_helper(
                        child, terminate_timeout, kill_timeout
                    )
                    raise AssertionError(
                        f"native helper was not ready after {ready_timeout}s "
                        f"{_helper_context(request)} exit={child.returncode} "
                        f"terminated={terminated} killed={killed} reaped={reaped} "
                        f"stderr={stderr!r}"
                    )
                stdout, stderr = child.communicate(timeout=timeout)
            else:
                stdout, stderr = child.communicate(
                    json.dumps(helper_request, ensure_ascii=False), timeout=timeout
                )
        except subprocess.TimeoutExpired as error:
            terminated, killed, reaped, stdout, stderr = _reap_helper(
                child,
                terminate_timeout,
                kill_timeout,
                error.stdout,
                error.stderr,
            )
            raise AssertionError(
                f"native helper timed out after {timeout}s {_helper_context(request)} "
                f"ready={ready} exit={child.returncode} "
                f"terminated={terminated} killed={killed} "
                f"reaped={reaped} stderr={stderr!r}"
            ) from None

        if child.returncode != 0 or stderr:
            raise AssertionError(
                f"native helper exited {child.returncode} {_helper_context(request)} "
                f"stderr={stderr!r}"
            )
        try:
            value = json.loads(stdout)
        except json.JSONDecodeError as error:
            raise AssertionError(
                f"native helper returned invalid JSON {_helper_context(request)}: {error}"
            ) from error
        assert isinstance(value, dict), _helper_context(request)
        return value
    finally:
        if child.poll() is None:
            _reap_helper(child, terminate_timeout, kill_timeout)
        if ready_directory is not None:
            ready_directory.cleanup()


class _NativeFreeTracker:
    def __init__(self, mapper: SPLMapper):
        self.mapper = mapper
        self.real_free = mapper._lib.spl_result_free
        self.local = threading.local()
        self.lock = threading.Lock()
        self.counts = {"python": 0, "raw_c": 0, "total": 0}
        mapper._lib.spl_result_free = self._free

    def _free(self, pointer) -> None:
        current = getattr(self.local, "current", None)
        assert current is not None, "spl_result_free called outside a tracked operation"
        assert pointer, "spl_result_free received a null result"
        current["count"] += 1
        with self.lock:
            self.counts[current["owner"]] += 1
            self.counts["total"] += 1
        self.real_free(pointer)

    def call(self, owner: str, operation) -> dict:
        current = {"owner": owner, "count": 0}
        assert getattr(self.local, "current", None) is None
        self.local.current = current
        try:
            value = operation()
        finally:
            self.local.current = None
        assert current["count"] == 1, (owner, current["count"])
        return value

    def close(self) -> None:
        self.mapper._lib.spl_result_free = self.real_free


def _exercise_native_document(
    mapper: SPLMapper, tracker: _NativeFreeTracker, document: dict
) -> dict:
    raw_analysis = tracker.call(
        "raw_c", lambda: call_raw_c(mapper, "analyze_query", document)
    )
    raw_requirements = tracker.call(
        "raw_c", lambda: call_raw_c(mapper, "requirements_query", document)
    )
    python_analysis = tracker.call(
        "python", lambda: call_python(mapper, "analyze_query", document)
    )
    python_requirements = tracker.call(
        "python", lambda: call_python(mapper, "requirements_query", document)
    )
    assert python_analysis is not raw_analysis
    assert python_requirements is not raw_requirements
    assert compact_bytes(python_analysis) == compact_bytes(raw_analysis)
    assert compact_bytes(python_requirements) == compact_bytes(raw_requirements)
    assert_resource_limit(python_analysis, python_requirements, document)
    return {"analysis": python_analysis, "requirements": python_requirements}


def _native_helper_execute(request: dict) -> dict:
    mode = request.get("mode")
    if mode == "fail":
        raise RuntimeError("simulated native helper failure")
    if mode == "hang":
        time.sleep(request.get("test_ready_delay", 0))
        if os.name != "nt":
            signal.signal(signal.SIGTERM, signal.SIG_IGN)
        ready_path = request.get("_ready_path")
        if not ready_path:
            raise ValueError("native helper hang mode requires a ready path")
        Path(ready_path).write_text("ready\n", encoding="utf-8")
        if message := request.get("test_stderr"):
            sys.stderr.write(message + "\n")
            sys.stderr.flush()
        time.sleep(3600)
        raise AssertionError("unreachable")
    if mode not in {"serial", "concurrent"}:
        raise ValueError(f"unsupported native helper mode: {mode}")

    documents = request.get("documents")
    if not isinstance(documents, list) or not documents:
        raise ValueError("native helper documents must be a non-empty list")
    mapper = open_mapper()
    reports = {}
    calls = 0
    with mapper:
        tracker = _NativeFreeTracker(mapper)
        try:
            if mode == "serial":
                for _ in range(2):
                    for document in documents:
                        observed = _exercise_native_document(mapper, tracker, document)
                        source_id = document["source_id"]
                        if source_id in reports:
                            assert reports[source_id] is not observed
                            assert compact_bytes(reports[source_id]) == compact_bytes(observed)
                        else:
                            reports[source_id] = observed
                        calls += 1
            else:
                requested_calls = request.get("calls", 16)

                def exercise(index: int) -> tuple[str, dict]:
                    document = documents[index % len(documents)]
                    return document["source_id"], _exercise_native_document(
                        mapper, tracker, document
                    )

                with ThreadPoolExecutor(max_workers=4) as pool:
                    observed = list(pool.map(exercise, range(requested_calls)))
                assert len({id(report) for _, report in observed}) == requested_calls
                assert len({id(report["analysis"]) for _, report in observed}) == requested_calls
                assert len({id(report["requirements"]) for _, report in observed}) == requested_calls
                for source_id, report in observed:
                    if source_id in reports:
                        assert compact_bytes(reports[source_id]) == compact_bytes(report)
                    else:
                        reports[source_id] = report
                calls = requested_calls
            assert mapper._active_calls == 0
        finally:
            tracker.close()
    assert mapper._closed
    return {
        "calls": calls,
        "active_calls": 0,
        "free_counts": tracker.counts,
        "mapper_closed": mapper._closed,
        "reports": reports,
    }


def _native_helper_main() -> None:
    if sys.argv[1:] != ["--requirements-native-helper"]:
        raise SystemExit("expected --requirements-native-helper")
    request = json.load(sys.stdin)
    json.dump(_native_helper_execute(request), sys.stdout, ensure_ascii=False)
    sys.stdout.write("\n")


def requirement_fixture_path() -> Path:
    return required_absolute_path("SPL_REQUIREMENTS_FIXTURES")


@pytest.fixture(scope="session")
def requirement_cases() -> list[dict]:
    data = json.loads(requirement_fixture_path().read_text(encoding="utf-8"))
    assert data["version"] == "1" and len(data["cases"]) == 20
    assert {
        (case["document"].get("language") or "spl", case["expected"]["query_status"])
        for case in data["cases"]
    } >= {
        (language, status)
        for language in ("spl", "spl2")
        for status in ("valid", "invalid", "incomplete")
    }
    return data["cases"]


@pytest.fixture(scope="session")
def special_documents() -> list[dict]:
    dense = [
        {
            "text": dense_query(size, language),
            "language": language,
            "source_id": f"dense-{language}-{size}.spl",
        }
        for language in ("spl", "spl2")
        for size in (65_536, 256 * 1024)
    ]
    sparse = []
    for language in ("spl", "spl2"):
        base = "FROM main" if language == "spl2" else "search host=x"
        text = base + " " * (300_000 - len(base.encode("utf-8")))
        assert len(text.encode("utf-8")) == 300_000
        sparse.append(
            {
                "text": text,
                "language": language,
                "source_id": f"sparse-{language}.spl",
            }
        )
    return dense + sparse


def test_sparse_requirement_documents_are_exactly_300000_bytes(special_documents):
    sparse = [
        document
        for document in special_documents
        if document["source_id"].startswith("sparse-")
    ]
    assert {document["source_id"] for document in sparse} == {
        "sparse-spl.spl",
        "sparse-spl2.spl",
    }
    assert all(len(document["text"].encode("utf-8")) == 300_000 for document in sparse)


@pytest.fixture(scope="session")
def go_reports(requirement_cases, special_documents, tmp_path_factory) -> dict[str, dict]:
    root_value = os.environ.get("SPL_REQUIREMENTS_GO_ROOT")
    root = Path(root_value) if root_value else Path(__file__).resolve().parents[2]
    assert root.is_absolute() and (root / "go.mod").is_file(), "missing copied Go requirement source root"
    documents = [case["document"] for case in requirement_cases] + special_documents
    dense = [
        document
        for document in special_documents
        if document["source_id"].startswith("dense-")
    ]
    concurrent_documents = [dense[index % len(dense)] for index in range(16)]
    helper = tmp_path_factory.mktemp("requirements-go") / "main.go"
    helper.write_text(GO_HELPER, encoding="utf-8")
    env = os.environ.copy()
    env["GOWORK"] = "off"
    env["GOCACHE"] = str(tmp_path_factory.mktemp("requirements-go-cache"))
    completed = subprocess.run(
        [os.environ.get("SPL_TOOLING_GO", "go"), "run", "-mod=readonly", str(helper)],
        cwd=root,
        env=env,
        input=json.dumps(
            {
                "documents": documents,
                "concurrent_documents": concurrent_documents,
            },
            ensure_ascii=False,
        ),
        capture_output=True,
        text=True,
        check=True,
        timeout=90,
    )
    values = json.loads(completed.stdout)
    assert len(values["reports"]) == len(documents)
    assert len(values["concurrent"]) == len(concurrent_documents)
    reports = {
        document["source_id"]: value
        for document, value in zip(documents, values["reports"], strict=True)
    }
    reports["__concurrent__"] = list(
        zip(concurrent_documents, values["concurrent"], strict=True)
    )
    return reports


@pytest.fixture(scope="session")
def requirement_evidence():
    fixture_hash = hashlib.sha256(requirement_fixture_path().read_bytes()).hexdigest()
    record = {
        "schema_version": 1,
        "fixture_sha256": fixture_hash,
        "corpus_cases": 0,
        "dense_cases": 0,
        "concurrent_calls": 0,
        "long_sparse_cases": 0,
        "request_error_cases": 0,
    }
    yield record
    if destination := os.environ.get("SPL_REQUIREMENTS_EVIDENCE"):
        Path(destination).write_text(
            json.dumps(record, indent=2, sort_keys=True) + "\n", encoding="utf-8"
        )


def expected_resource_limit(document: dict) -> tuple[dict, dict]:
    normalized = normalized_document(document)
    language = normalized["language"]
    size = len(normalized["text"].encode("utf-8"))
    dense_key = (language, size)
    if dense_key in DENSE_QUERY_DIGESTS:
        assert normalized["source_id"] == f"dense-{language}-{size}.spl"
        query_digest = DENSE_QUERY_DIGESTS[dense_key]
        omitted_offset = DENSE_OMITTED_OFFSETS[language]
    else:
        assert normalized["text"] == PORTABLE_RESOURCE_QUERY
        query_digest = "sha256:" + hashlib.sha256(
            normalized["text"].encode("utf-8")
        ).hexdigest()
        omitted_offset = 4096

    diagnostic = {
        "code": RESOURCE_CODE,
        "severity": "warning",
        "category": "resource_limit",
        "message": RESOURCE_MESSAGE,
        "location": {
            "start": {
                "offset": omitted_offset,
                "line": 1,
                "column": omitted_offset + 1,
            },
            "end": {
                "offset": omitted_offset + 1,
                "line": 1,
                "column": omitted_offset + 2,
            },
        },
        "stage_id": "",
        "scope_id": "",
    }
    requirements = {
        "schema_version": 1,
        "query": {
            "source_id": normalized["source_id"],
            "language": language,
            "profile": normalized["profile"],
            "version": normalized["version"],
            "query_digest": query_digest,
        },
        "capability_revision": CAPABILITY_REVISIONS[language],
        "query_status": "incomplete",
        "coverage": {"complete": False, "reasons": [RESOURCE_CODE]},
        "items": [],
        "gaps": [
            {
                "code": RESOURCE_CODE,
                "message": RESOURCE_GAP_MESSAGE,
                "reference_ids": [],
                "diagnostic_codes": [RESOURCE_CODE],
            }
        ],
        "diagnostics": [diagnostic],
    }
    analysis = {
        "schema_version": 1,
        "document": normalized,
        "status": "incomplete",
        "coverage": {
            "syntax_complete": False,
            "semantic_complete": False,
            "reasons": [RESOURCE_CODE],
        },
        "stages": [],
        "scopes": [],
        "references": [],
        "lineage": [],
        "dependencies": {key: [] for key in DEPENDENCY_KEYS},
        "diagnostics": [copy.deepcopy(diagnostic)],
        "requirements": copy.deepcopy(requirements),
    }
    return analysis, requirements


def assert_resource_limit(
    analysis: dict | None, requirements: dict, document: dict
) -> None:
    expected_analysis, expected_requirements = expected_resource_limit(document)
    assert requirements == expected_requirements
    assert len(compact_bytes(requirements)) <= 4096
    if analysis is not None:
        assert analysis == expected_analysis
        assert analysis["requirements"] == requirements
        assert len(compact_bytes(analysis)) <= len(document["text"].encode("utf-8")) + 4096


def _leaf_paths(value, prefix=()):
    if isinstance(value, dict) and value:
        for key, child in value.items():
            yield from _leaf_paths(child, prefix + (key,))
    elif isinstance(value, list) and value:
        for index, child in enumerate(value):
            yield from _leaf_paths(child, prefix + (index,))
    else:
        yield prefix


def resource_limit_oracle_mutations(analysis: dict, requirements: dict) -> list[tuple]:
    paths = [
        *(("analysis", *path) for path in _leaf_paths(analysis)),
        *(("requirements", *path) for path in _leaf_paths(requirements)),
        ("analysis", "dependencies", "__extra_key__"),
        ("analysis", "__extra_key__"),
        ("requirements", "__extra_key__"),
    ]
    assert len(paths) >= 75
    assert len(paths) == len(set(paths))
    return paths


def mutate_resource_limit_value(
    analysis: dict, requirements: dict, path: tuple
) -> None:
    current = analysis if path[0] == "analysis" else requirements
    for key in path[1:-1]:
        current = current[key]
    key = path[-1]
    if key == "__extra_key__":
        current[key] = True
        return
    value = current[key]
    if isinstance(value, bool):
        current[key] = not value
    elif isinstance(value, int):
        current[key] = value + 1
    elif isinstance(value, str):
        current[key] = value + "-mutated"
    elif isinstance(value, list):
        current[key] = [None]
    else:
        raise AssertionError((path, type(value)))


def test_requirements_fixture_matches_go_cli_http_and_python(
    requirement_cases, go_reports, cli_path, server_url, requirement_evidence
):
    with open_mapper() as mapper:
        for case in requirement_cases:
            document = case["document"]
            go = go_reports[document["source_id"]]
            assert go["requirements"] == case["expected"], case["id"]
            assert go["analysis"]["requirements"] == case["expected"], case["id"]

            py_analysis = call_python(mapper, "analyze_query", document)
            py_requirements = call_python(mapper, "requirements_query", document)
            c_analysis = call_raw_c(mapper, "analyze_query", document)
            c_requirements = call_raw_c(mapper, "requirements_query", document)
            analysis_exit, cli_analysis, _ = run_cli(cli_path, "analyze", document)
            requirement_exit, cli_requirements, _ = run_cli(cli_path, "requirements", document)
            analysis_status, http_analysis, _ = post_document(server_url, "analyze", document)
            requirement_status, http_requirements, _ = post_document(server_url, "requirements", document)

            assert analysis_exit == EXITS[go["analysis"]["status"]]
            assert requirement_exit == expected_requirements_exit(go["requirements"])
            assert analysis_status == requirement_status == 200
            assert py_analysis == c_analysis == cli_analysis == http_analysis == go["analysis"]
            assert py_requirements == c_requirements == cli_requirements == http_requirements == go["requirements"]

        first = requirement_cases[0]["document"]
        detached = call_python(mapper, "requirements_query", first)
        detached["items"].clear()
        assert call_python(mapper, "requirements_query", first) == requirement_cases[0]["expected"]
    requirement_evidence["corpus_cases"] = len(requirement_cases)


def test_dense_and_sparse_requirements_match_every_surface(
    special_documents, go_reports, server_url, requirement_evidence
):
    dense_count = 0
    sparse_count = 0
    dense = [
        document
        for document in special_documents
        if document["source_id"].startswith("dense-")
    ]
    native_dense = run_native_helper(
        {"mode": "serial", "documents": dense}, timeout=60
    )
    expected_native_calls = len(dense) * 2
    expected_owner_frees = expected_native_calls * 2
    assert native_dense["calls"] == expected_native_calls
    assert native_dense["active_calls"] == 0
    assert native_dense["free_counts"] == {
        "python": expected_owner_frees,
        "raw_c": expected_owner_frees,
        "total": expected_owner_frees * 2,
    }
    assert native_dense["mapper_closed"] is True

    with open_mapper() as mapper:
        for document in special_documents:
            go = go_reports[document["source_id"]]
            analysis_status, http_analysis, http_analysis_bytes = post_document(
                server_url, "analyze", document
            )
            requirement_status, http_requirements, http_requirement_bytes = post_document(
                server_url, "requirements", document
            )
            assert analysis_status == requirement_status == 200

            if document["source_id"].startswith("dense-"):
                dense_count += 1
                native = native_dense["reports"][document["source_id"]]
                assert native["analysis"] == http_analysis == go["analysis"]
                assert native["requirements"] == http_requirements == go["requirements"]
                assert_resource_limit(go["analysis"], go["requirements"], document)
                assert len(compact_bytes(go["requirements"])) <= 4096
                assert len(compact_bytes(go["analysis"])) <= len(document["text"].encode()) + 4096
                assert len(http_requirement_bytes) <= 4096
                assert len(http_analysis_bytes) <= len(document["text"].encode()) + 4096
            else:
                sparse_count += 1
                py_analysis = call_python(mapper, "analyze_query", document)
                py_requirements = call_python(mapper, "requirements_query", document)
                c_analysis = call_raw_c(mapper, "analyze_query", document)
                c_requirements = call_raw_c(mapper, "requirements_query", document)
                assert py_analysis == c_analysis == http_analysis == go["analysis"]
                assert py_requirements == c_requirements == http_requirements == go["requirements"]
                assert not any(
                    diagnostic["code"] == RESOURCE_CODE for diagnostic in go["analysis"]["diagnostics"]
                )
                assert go["analysis"]["stages"]
    requirement_evidence["dense_cases"] = dense_count
    requirement_evidence["long_sparse_cases"] = sparse_count


def test_concurrent_dense_results_are_deterministic_bounded_and_owned(
    special_documents, go_reports, server_url, requirement_evidence
):
    dense = [document for document in special_documents if document["source_id"].startswith("dense-")]
    calls = 16
    expected = {document["source_id"]: go_reports[document["source_id"]] for document in dense}
    native = run_native_helper(
        {"mode": "concurrent", "documents": dense, "calls": calls}, timeout=120
    )
    assert native["calls"] == calls
    assert native["active_calls"] == 0
    assert native["free_counts"] == {"python": 32, "raw_c": 32, "total": 64}
    assert native["mapper_closed"] is True

    for document, report in go_reports["__concurrent__"]:
        assert report == expected[document["source_id"]]
        assert_resource_limit(report["analysis"], report["requirements"], document)

    def exercise_http(index: int) -> tuple[dict, dict, bytes, bytes]:
        document = dense[index % len(dense)]
        analysis_status, analysis, analysis_bytes = post_document(
            server_url, "analyze", document
        )
        requirement_status, requirements, requirement_bytes = post_document(
            server_url, "requirements", document
        )
        assert analysis_status == requirement_status == 200
        return analysis, requirements, analysis_bytes, requirement_bytes

    with ThreadPoolExecutor(max_workers=4) as pool:
        http_observed = list(pool.map(exercise_http, range(calls)))

    for index, (http_analysis, http_requirements, analysis_bytes, requirement_bytes) in enumerate(
        http_observed
    ):
        document = dense[index % len(dense)]
        want = expected[document["source_id"]]
        assert http_analysis == want["analysis"]
        assert http_requirements == want["requirements"]
        assert len(requirement_bytes) <= 4096
        assert len(analysis_bytes) <= len(document["text"].encode()) + 4096
        native_report = native["reports"][document["source_id"]]
        assert compact_bytes(native_report["analysis"]) == compact_bytes(want["analysis"])
        assert compact_bytes(native_report["requirements"]) == compact_bytes(want["requirements"])
    requirement_evidence["concurrent_calls"] = calls


def test_requirements_real_http_request_errors_and_one_mebibyte_boundary(
    server_url, requirement_evidence
):
    valid = b'{"text":"search host=web"}'
    exact = valid + b" " * (1024 * 1024 - len(valid))
    for operation, status_key in (("analyze", "status"), ("requirements", "query_status")):
        status, report, _ = post_raw(server_url, operation, exact)
        assert status == 200 and report[status_key] == "valid"

    errors = 0
    for operation in ("analyze", "requirements"):
        for payload in (b"{", b'{"text":"one","text":"two"}', exact + b" "):
            status, body, _ = post_raw(server_url, operation, payload)
            assert status == 400
            assert body.get("error")
            assert not {
                "schema_version", "query", "query_status", "coverage", "items", "gaps", "diagnostics"
            }.intersection(body)
            errors += 1
    requirement_evidence["request_error_cases"] = errors


def test_portable_cli_resource_limit_query(cli_path):
    query = "a " * 2048 + "a"
    assert len(query.encode("ascii")) == 4097
    assert len(query.encode("utf-16-le")) // 2 == 4097
    assert query.count("a") == 2049
    assert query.count(" ") == 2048

    for language in ("spl", "spl2"):
        document = {
            "text": query,
            "language": language,
            "source_id": f"portable-{language}-4097.spl",
        }
        for operation in ("analyze", "requirements"):
            args = cli_arguments(cli_path, operation, document)
            assert windows_command_utf16_units(args) < 32767
            exit_code, value, encoded = run_cli(cli_path, operation, document)
            assert exit_code == 3
            if operation == "analyze":
                assert_resource_limit(value, value["requirements"], document)
                assert len(encoded) <= len(query.encode()) + 4096
            else:
                assert_resource_limit(None, value, document)
                assert len(encoded) <= 4096


def test_requirements_cli_exit_accounts_for_incomplete_coverage(cli_path):
    document = {
        "text": "search key=1 | lookup users key OUTPUTNEW a | table a | where a=1",
        "source_id": "requirements-valid-incomplete-coverage.spl",
    }
    exit_code, requirements, _ = run_cli(cli_path, "requirements", document)
    assert requirements["query_status"] == "valid"
    assert requirements["coverage"]["complete"] is False
    assert expected_requirements_exit(requirements) == 3
    assert exit_code == 3


def test_dense_resource_limit_oracle_is_exact(go_reports, special_documents):
    document = next(
        value
        for value in special_documents
        if value["source_id"] == "dense-spl-65536.spl"
    )
    reports = go_reports[document["source_id"]]
    assert_resource_limit(reports["analysis"], reports["requirements"], document)

    mutations = resource_limit_oracle_mutations(
        reports["analysis"], reports["requirements"]
    )
    assert ("analysis", "dependencies", "indexes") in mutations
    assert ("analysis", "diagnostics", 0, "location", "start", "offset") in mutations
    assert ("requirements", "query", "query_digest") in mutations
    assert ("requirements", "gaps", 0, "diagnostic_codes", 0) in mutations
    for path in mutations:
        analysis = copy.deepcopy(reports["analysis"])
        requirements = copy.deepcopy(reports["requirements"])
        mutate_resource_limit_value(analysis, requirements, path)
        with pytest.raises(AssertionError):
            assert_resource_limit(analysis, requirements, document)


def test_dense_native_helper_is_bounded_and_owned(special_documents):
    dense = [
        document
        for document in special_documents
        if document["source_id"] == "dense-spl-65536.spl"
    ]
    result = run_native_helper({"mode": "serial", "documents": dense}, timeout=60)
    assert result["calls"] == 2
    assert result["active_calls"] == 0
    assert result["free_counts"] == {"python": 4, "raw_c": 4, "total": 8}
    assert result["mapper_closed"] is True

    with pytest.raises(AssertionError, match="exited 1.*simulated native helper failure"):
        run_native_helper(
            {"mode": "fail", "language": "spl", "size": 65_536}, timeout=5
        )

    with pytest.raises(AssertionError, match="timed out.*terminated=True.*reaped=True") as error:
        run_native_helper(
            {"mode": "hang", "language": "spl", "size": 65536},
            timeout=0.5,
            terminate_timeout=0.05,
            kill_timeout=1,
        )
    if os.name != "nt":
        assert "killed=True" in str(error.value)


def test_native_helper_timeout_preserves_early_stderr():
    message = "early native helper stderr"
    with pytest.raises(AssertionError) as error:
        run_native_helper(
            {"mode": "hang", "test_stderr": message},
            timeout=0.05,
            terminate_timeout=0.05,
            kill_timeout=1,
        )
    assert str(error.value).count(message) == 1


def test_native_helper_timeout_starts_after_ready():
    started = time.monotonic()
    with pytest.raises(AssertionError, match="ready=True"):
        run_native_helper(
            {"mode": "hang", "test_ready_delay": 0.2},
            timeout=0.05,
            terminate_timeout=0.05,
            kill_timeout=1,
        )
    assert time.monotonic() - started >= 0.2


if __name__ == "__main__":
    _native_helper_main()
