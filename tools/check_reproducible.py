#!/usr/bin/env python3
"""Rebuild one committed ref twice and require byte-identical release payloads."""

from __future__ import annotations

import argparse
import ctypes
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
from tools.check_package import check_package
from tools.release import _target_name, artifact_hashes


def compare_artifacts(first_output: Path, second_output: Path) -> dict[str, str]:
    left = artifact_hashes(first_output)
    right = artifact_hashes(second_output)
    if left.keys() != right.keys():
        raise RuntimeError("release artifact sets differ")
    different = [name for name in sorted(left) if left[name] != right[name]]
    if different:
        raise RuntimeError("non-reproducible artifacts: " + ", ".join(different))
    return left


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


def check_reproducible(ref: str, output: Path) -> dict[str, object]:
    root = ROOT
    output = output.resolve()
    if output.exists() and any(output.iterdir()):
        raise RuntimeError(f"reproducibility output must be new or empty: {output}")
    output.mkdir(parents=True, exist_ok=True)
    source_sha = _git(root, "rev-parse", "--verify", f"{ref}^{{commit}}")
    epoch = int(_git(root, "show", "-s", "--format=%ct", source_sha))
    files = tracked_files(root, source_sha)
    before = hash_tracked_files(root, files)
    result: dict[str, object] = {"source_sha": source_sha, "target": _target_name(), "status": "failed", "artifact_hashes": {}, "environment": {}, "failed_checks": []}
    try:
        environments = []
        for number in (1, 2):
            source = output / f"source-{number}"
            source.mkdir()
            export_ref(root, source_sha, output / f"source-{number}.tar", source)
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
        check_package(sdist, wheel.parent, version)
        after = hash_tracked_files(root, files)
        if before != after:
            raise RuntimeError("tracked source files changed during reproducibility check")
        result.update(status="passed", target=environments[0]["target"], artifact_hashes=hashes, environment=environments[0], accepted_payloads=accepted)
    except Exception as error:
        result["failed_checks"] = [str(error)]
        raise
    finally:
        (output / "result.json").write_text(json.dumps(result, indent=2, sort_keys=True) + "\n", encoding="utf-8")
    return result


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--ref", default="HEAD")
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    check_reproducible(args.ref, args.output)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
