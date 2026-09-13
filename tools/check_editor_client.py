#!/usr/bin/env python3
"""Exercise the real VS Code client in an isolated Extension Development Host."""
from __future__ import annotations
import argparse
import hashlib
import json
import os
from pathlib import Path
import re
import shutil
import signal
import subprocess
import tempfile

ROOT = Path(__file__).resolve().parents[1]


def run(cli: Path, vscode: Path, consumer_root: Path, source_sha: str, evidence: Path) -> dict:
    if not re.fullmatch(r"[0-9a-f]{40}", source_sha):
        raise ValueError("source-sha must be a full Git SHA")
    cli, vscode, consumer_root = cli.resolve(strict=True), vscode.resolve(strict=True), consumer_root.resolve(strict=True)
    fixture = ROOT / "tests/editor-client"
    modules = consumer_root / "node-consumer/node_modules"
    if not (modules / "vscode-languageclient/package.json").is_file():
        raise ValueError("reviewed offline client installation missing")
    if json.loads((modules / "vscode-languageclient/package.json").read_text())["version"] != "9.0.1":
        raise ValueError("client version differs from reviewed pin")
    result = {"status": "failed", "source_sha": source_sha, "cli": str(cli),
              "cli_sha256": hashlib.sha256(cli.read_bytes()).hexdigest(), "vscode": str(vscode),
              "lock_sha256": hashlib.sha256((fixture / "package-lock.json").read_bytes()).hexdigest()}
    result["fixture_sha256"] = {name: hashlib.sha256((fixture / name).read_bytes()).hexdigest()
                                 for name in ("package.json", "extension.js", "test/index.js")}
    evidence.parent.mkdir(parents=True, exist_ok=True)
    with tempfile.TemporaryDirectory(prefix="editor-run-", dir=consumer_root) as temporary:
        root = Path(temporary)
        extension = root / "extension"
        shutil.copytree(fixture, extension, ignore=shutil.ignore_patterns("node_modules"))
        shutil.copytree(modules, extension / "node_modules")
        workspace = root / "workspace"
        workspace.mkdir()
        query = workspace / "unicode.spl"
        query.write_text("eval '😀a'=", encoding="utf-8")
        test_result = root / "client-result.json"
        command = [str(vscode), "--wait", "--user-data-dir", str(root / "user-data"), "--extensions-dir", str(root / "extensions"),
                   "--extensionDevelopmentPath=" + str(extension), "--extensionTestsPath=" + str(extension / "test"),
                   "--sync", "off", "--disable-updates", "--skip-welcome", "--skip-release-notes",
                   "--disable-workspace-trust", "--disable-telemetry", str(workspace)]
        env = os.environ.copy()
        env.pop("ELECTRON_RUN_AS_NODE", None)
        env.update(SPL_CONSUMER_CLI=str(cli), SPL_CONSUMER_QUERY=str(query), SPL_CONSUMER_RESULT=str(test_result))
        version = subprocess.run([str(vscode), "--version"], capture_output=True, text=True, env=env, timeout=20)
        result.update(editor_version_output=version.stdout, editor_version_stderr=version.stderr, command=command)
        process = subprocess.Popen(command, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True, env=env, start_new_session=True)
        try:
            stdout, stderr = process.communicate(timeout=120)
            result.update(exit_code=process.returncode, stdout=stdout, stderr=stderr)
            if test_result.is_file():
                result["client"] = json.loads(test_result.read_text())
            if process.returncode != 0 or result.get("client", {}).get("status") != "passed":
                raise RuntimeError("real editor assertions did not pass; see evidence")
            if hashlib.sha256(cli.read_bytes()).hexdigest() != result["cli_sha256"]:
                raise RuntimeError("candidate executable changed during editor checks")
            result["status"] = "passed"
        except Exception as error:
            result["error"] = str(error)
        finally:
            if process.poll() is None:
                os.killpg(process.pid, signal.SIGTERM)
                try:
                    stdout, stderr = process.communicate(timeout=10)
                except subprocess.TimeoutExpired:
                    os.killpg(process.pid, signal.SIGKILL)
                    stdout, stderr = process.communicate()
                result.update(exit_code=process.returncode, stdout=stdout, stderr=stderr)
            evidence.write_text(json.dumps(result, indent=2) + "\n")
    return result


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    for name in ("cli", "vscode", "consumer-root", "evidence"):
        parser.add_argument("--" + name, type=Path, required=True)
    parser.add_argument("--source-sha", required=True)
    args = parser.parse_args()
    result = run(args.cli, args.vscode, args.consumer_root, args.source_sha, args.evidence)
    print(json.dumps({"status": result["status"], "evidence": str(args.evidence)}))
    return 0 if result["status"] == "passed" else 1


if __name__ == "__main__":
    raise SystemExit(main())
