"""Prepared core M6 attempt capture; do not execute before reviewed Task5."""
import datetime
import hashlib
import json
import os
from pathlib import Path
import signal
import shutil
import subprocess
import sys


root = Path(sys.argv[1]).resolve()
expected_head = sys.argv[2]
scope = "core"
base = Path("/private/tmp")
scripts = {"core": (base / "spl-toolkit-remaining-venv/bin/python", base / "spl-toolkit-root-m6-go-acceptance.py")}
go_binary = Path(shutil.which("go")).resolve()


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def git(*arguments):
    return subprocess.check_output(["git", *arguments], cwd=root).decode()


def snapshot():
    tracked = [path for path in git("ls-files", "-z").split("\0") if path]
    untracked = [path for path in git("ls-files", "--others", "--exclude-standard", "-z").split("\0") if path]
    artifact_paths = [root / "build" / name for name in ("spl-toolkit", "spl-toolkit-server", "libspl_toolkit.dylib")]
    artifact_paths += sorted(path for path in (root / "dist").glob("*") if path.name.endswith((".whl", ".tar.gz")))
    input_paths = {Path(__file__), go_binary, scripts[scope][0].resolve()}
    input_paths.update(base / ("spl-toolkit-root-m6-" + name) for name in (
        "cases.json", "validation-cases.json", "semantic-checks.py", "go.go", "go-acceptance.py"))
    for case in json.loads((base / "spl-toolkit-root-m6-validation-cases.json").read_text())["cases"]:
        if "catalog_path" in case:
            input_paths.add(Path(case["catalog_path"]))
    return {
        "head": git("rev-parse", "HEAD").strip(),
        "tracked_diff": git("diff", "--name-only"),
        "index_diff": git("diff", "--cached", "--name-only"),
        "tracked": {path: digest(root / path) for path in tracked},
        "untracked": {path: digest(root / path) for path in untracked},
        "artifacts": {str(path): digest(path) for path in artifact_paths},
        "independent_inputs": {str(path): digest(path) for path in sorted(input_paths)},
    }


stamp = datetime.datetime.now(datetime.timezone.utc).strftime("%Y%m%dT%H%M%S%fZ")
attempt = base / "spl-toolkit-root-m6-core-attempts" / (expected_head[:12] + "-" + scope + "-" + stamp)
attempt.mkdir(parents=True, exist_ok=False)
before = snapshot()
assert before["head"] == expected_head and not before["tracked_diff"] and not before["index_diff"]
interpreter, script = scripts[scope]
command = [str(interpreter), str(script), str(root), expected_head]

child_paths = [base / ("spl-toolkit-root-m6-" + name) for name in (
    "go-attempt.json", "go-attempt.stdout.json", "go-attempt.stderr.txt", "go-results.json", "go-acceptance.json")]
child_before = {str(path): digest(path) if path.is_file() else None for path in child_paths}

record = {"status": "running", "scope": scope, "command": command, "cwd": str(root), "before": before,
          "tooling": {"go_executable": str(go_binary),
                      "go_version": subprocess.check_output([str(go_binary), "version"], text=True).strip(),
                      "interpreter": str(interpreter.resolve()),
                      "environment": {key: os.environ.get(key) for key in
                                      ("GOTOOLCHAIN", "GOPROXY", "GOSUMDB", "GOCACHE", "GOMODCACHE", "CGO_ENABLED", "GOFLAGS")}}}
record_path = attempt / "attempt.json"
record_path.write_text(json.dumps(record, indent=2) + "\n")
print("Attempt:", attempt, flush=True)
with (attempt / "stdout.log").open("wb") as stdout, (attempt / "stderr.log").open("wb") as stderr:
    process = subprocess.Popen(command, cwd=root, env=os.environ.copy(), stdout=stdout, stderr=stderr, start_new_session=True)
    try:
        record["exit_code"] = process.wait(timeout=900)
    except subprocess.TimeoutExpired:
        try:
            os.killpg(process.pid, signal.SIGTERM)
        except ProcessLookupError:
            pass
        try:
            process.wait(timeout=5)
        except subprocess.TimeoutExpired:
            os.killpg(process.pid, signal.SIGKILL)
            process.wait()
        record["exit_code"] = 124
        record["timeout_seconds"] = 900
record["child_returncode"] = process.returncode
record["after"] = snapshot()
record["inputs_unchanged"] = record["before"] == record["after"]

record["child_outputs"] = {}
for path in child_paths:
    if path.is_file():
        copied = attempt / "child_outputs" / path.relative_to(base)
        copied.parent.mkdir(parents=True, exist_ok=True)
        shutil.copyfile(path, copied)
        record["child_outputs"][str(path)] = {
            "archived_path": str(copied), "sha256": digest(copied),
            "before_sha256": child_before[str(path)],
            "qualification": "Observed after this attempt; presence or equal bytes alone does not prove the child step executed."}

record["status"] = "passed" if record["exit_code"] == 0 and record["inputs_unchanged"] else "failed"
record["logs"] = {name: {"path": str(attempt / name), "sha256": digest(attempt / name)} for name in ("stdout.log", "stderr.log")}
record_path.write_text(json.dumps(record, indent=2) + "\n")
for name in ("stdout.log", "stderr.log"):
    print((attempt / name).read_bytes().decode("utf-8", errors="backslashreplace"), end="")
print("Attempt record:", record_path, flush=True)
assert record["status"] == "passed", record_path
