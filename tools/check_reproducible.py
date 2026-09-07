#!/usr/bin/env python3
"""Rebuild one committed ref twice and require byte-identical release payloads."""

from __future__ import annotations

import argparse
import ctypes
import hashlib
import json
import os
from pathlib import Path
import shutil
import socket
import subprocess
import sys
import time
import urllib.request

ROOT = Path(__file__).resolve().parents[1]
if __package__ in (None, ""):
    sys.path.insert(0, str(ROOT))

from tools.check_clean_build import export_ref, hashes as hash_tracked_files, tracked_files
from tools.release import _target_name, artifact_hashes


def _sha256(path: Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as source:
        for chunk in iter(lambda: source.read(1024 * 1024), b""):
            digest.update(chunk)
    return digest.hexdigest()


def _promote_accepted(output: Path, result: dict[str, object]) -> None:
    """Copy only a fully verified build into the CI transport directory."""
    if result.get("status") != "passed" or result.get("failed_checks") != []:
        raise RuntimeError("cannot promote a failed or diagnostic reproducibility result")
    environment = result.get("environment")
    if not isinstance(environment, dict) or environment.get("pinned_environment") is not True:
        raise RuntimeError("cannot promote an unpinned reproducibility result")
    hashes = result.get("artifact_hashes")
    payloads = result.get("accepted_payloads")
    if not isinstance(hashes, dict) or not isinstance(payloads, dict):
        raise RuntimeError("cannot promote incomplete reproducibility evidence")
    if environment.get("artifacts") != hashes:
        raise RuntimeError("cannot promote inconsistent environment artifact hashes")
    source = output / "build-1"
    if artifact_hashes(source) != hashes:
        raise RuntimeError("cannot promote payloads whose hashes changed after verification")
    accepted = output / "accepted"
    if accepted.exists():
        raise RuntimeError(f"accepted output already exists: {accepted}")
    staging = output / "accepted.staging"
    if staging.exists():
        raise RuntimeError(f"accepted staging output already exists: {staging}")
    shutil.copytree(source, staging)
    if os.name != "nt":
        for role in ("cli", "server"):
            name = payloads.get(role)
            if not isinstance(name, str) or name not in hashes:
                raise RuntimeError(f"accepted {role} payload is missing from verified hashes")
            path = staging / name
            if _sha256(path) != hashes[name]:
                raise RuntimeError(f"accepted {role} payload hash changed during promotion")
            path.chmod(path.stat().st_mode | 0o755)
    wheel_names = [name for name in hashes if name.endswith(".whl")]
    if len(wheel_names) != 1:
        raise RuntimeError("accepted payload set must contain exactly one wheel")
    architecture = "arm64" if str(result["target"]).endswith("arm64") else "x86_64"
    common = {
        "schema_version": 1,
        "source_sha": result["source_sha"],
        "status": "passed",
        "target": result["target"],
        "architecture": architecture,
    }
    native_record = common | {
        "kind": "native",
        "environment": environment,
    }
    reproducibility_record = common | {
        "kind": "reproducibility",
        "wheel_sha256": hashes[wheel_names[0]],
        "artifact_hashes": hashes,
        "accepted_payloads": payloads,
        "environment": environment,
        "checks": {"clean_source": "passed", "payloads": "passed", "package": "passed"},
    }
    (staging / "result.json").write_text(json.dumps(result, indent=2, sort_keys=True) + "\n", encoding="utf-8")
    (staging / "environment.json").write_text(json.dumps(environment, indent=2, sort_keys=True) + "\n", encoding="utf-8")
    (staging / "native.json").write_text(json.dumps(native_record, indent=2, sort_keys=True) + "\n", encoding="utf-8")
    (staging / "reproducibility.json").write_text(json.dumps(reproducibility_record, indent=2, sort_keys=True) + "\n", encoding="utf-8")
    staging.rename(accepted)


def compare_artifacts(first_output: Path, second_output: Path) -> dict[str, str]:
    left = artifact_hashes(first_output)
    right = artifact_hashes(second_output)
    if left.keys() != right.keys():
        raise RuntimeError("release artifact sets differ")
    different = [name for name in sorted(left) if left[name] != right[name]]
    if different:
        raise RuntimeError("non-reproducible artifacts: " + ", ".join(different))
    return left


def _result_status(environment: dict[str, object]) -> str:
    return "passed" if environment.get("pinned_environment") is True else "diagnostic-passed"


def _git(root: Path, *args: str) -> str:
    return subprocess.run(["git", *args], cwd=root, check=True, text=True, capture_output=True).stdout.strip()


def _run_payloads(output: Path, version: str, fixture: Path) -> dict[str, str]:
    cli = next(output.glob(f"spl-toolkit-{version}-*"))
    server = next(output.glob(f"spl-toolkit-server-{version}-*"))
    native = next(path for path in output.glob(f"libspl_toolkit-{version}-*") if path.suffix in (".so", ".dylib", ".dll"))
    cli_version = subprocess.run([str(cli), "version"], check=True, text=True, capture_output=True).stdout.strip()
    if version not in cli_version:
        raise RuntimeError("accepted CLI reports the wrong version")
    mapped = subprocess.run(
        [str(cli), "map", "--config", str(fixture), "search src_ip=1"],
        check=True, text=True, capture_output=True,
    ).stdout.strip()
    if mapped != "search source_ip=1":
        raise RuntimeError(f"accepted CLI mapping failed: {mapped}")
    library = ctypes.CDLL(str(native))
    library.spl_toolkit_version.restype = ctypes.c_void_p
    library.spl_string_free.argtypes = [ctypes.c_void_p]
    pointer = library.spl_toolkit_version()
    if not pointer:
        raise RuntimeError("accepted native library returned no version")
    try:
        native_version = ctypes.string_at(pointer).decode()
    finally:
        library.spl_string_free(pointer)
    if native_version != version:
        raise RuntimeError(f"accepted native library reports {native_version}, expected {version}")
    listener = socket.socket()
    listener.bind(("127.0.0.1", 0))
    port = listener.getsockname()[1]
    listener.close()
    process = subprocess.Popen([str(server)], env=os.environ | {"PORT": str(port)}, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)
    try:
        deadline = time.monotonic() + 15
        body = ""
        while time.monotonic() < deadline:
            if process.poll() is not None:
                stdout, stderr = process.communicate()
                raise RuntimeError(f"accepted server exited early: {stdout}{stderr}")
            try:
                body = urllib.request.urlopen(f"http://127.0.0.1:{port}/api/v1/health", timeout=1).read().decode()
                break
            except OSError:
                time.sleep(0.1)
        else:
            raise RuntimeError("accepted server health check timed out")
        if version not in body:
            raise RuntimeError("accepted server health response omits the version")
        request = urllib.request.Request(
            f"http://127.0.0.1:{port}/api/v1/query/map",
            data=json.dumps({
                "query": "search src_ip=1",
                "mappings": [{"source": "src_ip", "target": "source_ip"}],
            }).encode(),
            headers={"Content-Type": "application/json"},
        )
        response = json.loads(urllib.request.urlopen(request, timeout=2).read())
        if response.get("mapped_query") != "search source_ip=1":
            raise RuntimeError("accepted server mapping failed")
    finally:
        process.terminate()
        try:
            process.wait(timeout=5)
        except subprocess.TimeoutExpired:
            process.kill()
            process.wait()
    return {"cli": cli.name, "server": server.name, "native": native.name}


def _run_package_check(source: Path, sdist: Path, wheel_dir: Path, version: str, log: Path) -> None:
    completed = subprocess.run(
        [
            sys.executable,
            str(source / "tools" / "check_package.py"),
            "--sdist", str(sdist),
            "--wheel-dir", str(wheel_dir),
            "--expected-version", version,
            "--source-root", str(source),
        ],
        cwd=source,
        text=True,
        capture_output=True,
        check=False,
    )
    log.write_text(completed.stdout + completed.stderr, encoding="utf-8")
    if completed.returncode:
        raise RuntimeError(f"exported package acceptance failed; see {log}")


def _check_reproducible_worker(repository: Path, source_sha: str, epoch: int, output: Path) -> dict[str, object]:
    files = tracked_files(repository, source_sha)
    result: dict[str, object] = {"source_sha": source_sha, "target": _target_name(), "status": "failed", "artifact_hashes": {}, "environment": {}, "failed_checks": []}
    try:
        environments = []
        for number in (1, 2):
            source = output / f"source-{number}"
            source.mkdir()
            export_ref(repository, source_sha, output / f"source-{number}.tar", source)
            if number == 1:
                before = hash_tracked_files(source, files)
            build_output = output / f"build-{number}"
            cache = output / f"gocache-{number}"
            temporary = output / f"gotmp-{number}"
            module_cache = output / f"gomodcache-{number}"
            cache.mkdir(); temporary.mkdir(); module_cache.mkdir()
            env = os.environ.copy()
            env.update({"GOCACHE": str(cache), "GOTMPDIR": str(temporary), "GOMODCACHE": str(module_cache)})
            completed = subprocess.run(
                [sys.executable, "tools/release.py", "--source", str(source), "--output", str(build_output), "--epoch", str(epoch)],
                cwd=source, env=env, text=True, capture_output=True, check=False,
            )
            (output / f"release-{number}.log").write_text(completed.stdout + completed.stderr, encoding="utf-8")
            if completed.returncode:
                detail = next((line for line in reversed(completed.stderr.splitlines()) if line.strip()), "unknown error")
                raise RuntimeError(f"release build {number} failed: {detail}; see {output / f'release-{number}.log'}")
            environments.append(json.loads((output / f"build-{number}-evidence" / "environment.json").read_text(encoding="utf-8")))
        hashes = compare_artifacts(output / "build-1", output / "build-2")
        version = (output / "source-1" / "VERSION").read_text(encoding="utf-8").strip()
        accepted = _run_payloads(output / "build-1", version, output / "source-1" / "testdata" / "baseline" / "mappings.json")
        wheel = next((output / "build-1").glob("spl_toolkit-*.whl"))
        sdist = next((output / "build-1").glob("spl_toolkit-*.tar.gz"))
        _run_package_check(output / "source-1", sdist, wheel.parent, version, output / "package-check.log")
        after = hash_tracked_files(output / "source-1", files)
        if before != after:
            raise RuntimeError("exported tracked source files changed during reproducibility check")
        result.update(
            status=_result_status(environments[0]),
            target=environments[0]["target"],
            artifact_hashes=hashes,
            environment=environments[0],
            accepted_payloads=accepted,
        )
        _promote_accepted(output, result)
    except Exception as error:
        result["status"] = "failed"
        result["failed_checks"] = [str(error)]
        raise
    finally:
        (output / "result.json").write_text(json.dumps(result, indent=2, sort_keys=True) + "\n", encoding="utf-8")
    return result


def check_reproducible(ref: str, output: Path) -> dict[str, object]:
    root = ROOT
    output = output.resolve()
    if output.exists() and any(output.iterdir()):
        raise RuntimeError(f"reproducibility output must be new or empty: {output}")
    output.mkdir(parents=True, exist_ok=True)
    source_sha = _git(root, "rev-parse", "--verify", f"{ref}^{{commit}}")
    epoch = int(_git(root, "show", "-s", "--format=%ct", source_sha))
    checker_source = output / "checker-source"
    checker_source.mkdir()
    export_ref(root, source_sha, output / "checker-source.tar", checker_source)
    command = [
        sys.executable,
        str(checker_source / "tools" / "check_reproducible.py"),
        "--output", str(output),
        "--worker-repository", str(root),
        "--worker-source-sha", source_sha,
        "--worker-epoch", str(epoch),
    ]
    completed = subprocess.run(command, cwd=checker_source, text=True, capture_output=True, check=False)
    (output / "checker.log").write_text(completed.stdout + completed.stderr, encoding="utf-8")
    if completed.returncode:
        detail = next(
            (line for line in reversed(completed.stderr.splitlines()) if line.strip()),
            "exported checker failed without an error message",
        )
        raise RuntimeError(f"exported checker failed: {detail}; see {output / 'checker.log'}")
    return json.loads((output / "result.json").read_text(encoding="utf-8"))


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--ref", default="HEAD")
    parser.add_argument("--output", type=Path, required=True)
    parser.add_argument("--worker-repository", type=Path, help=argparse.SUPPRESS)
    parser.add_argument("--worker-source-sha", help=argparse.SUPPRESS)
    parser.add_argument("--worker-epoch", type=int, help=argparse.SUPPRESS)
    args = parser.parse_args()
    worker_values = (args.worker_repository, args.worker_source_sha, args.worker_epoch)
    if any(value is not None for value in worker_values):
        if any(value is None for value in worker_values):
            parser.error("all exported-checker worker arguments are required together")
        _check_reproducible_worker(
            args.worker_repository.resolve(), args.worker_source_sha, args.worker_epoch, args.output.resolve()
        )
    else:
        check_reproducible(args.ref, args.output)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
