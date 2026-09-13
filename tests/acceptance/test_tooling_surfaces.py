"""Real CLI and loopback HTTP tooling examples, without a native dependency.

Go package tests compare each complete adapter report with its canonical Go
operation. Here manifest/inline acquisition provenance is asserted explicitly;
canonical per-entry evidence is compared without filtering it. Native and actual
editor/SARIF consumers have separate acceptance gates.
"""
from __future__ import annotations

import json
import os
from pathlib import Path
import socket
import subprocess
import sys
import time
from urllib.error import HTTPError, URLError
from urllib.request import Request, urlopen

import pytest

ROOT = Path(__file__).resolve().parents[2]


@pytest.fixture(scope="module")
def tooling_binaries(tmp_path_factory):
    directory = tmp_path_factory.mktemp("tooling-binaries")
    binaries = []
    for variable, package, name in (("SPL_CLI", "./cmd", "spl-toolkit"),
                                    ("SPL_SERVER", "./cmd/server", "spl-server")):
        binary = Path(os.environ.get(variable, directory / name))
        if variable not in os.environ:
            command = [os.environ.get("SPL_TOOLING_GO", "go"), "build", "-mod=readonly"]
            if sys.platform == "darwin":
                command += ["-ldflags=-linkmode=external"]
            subprocess.run(command + ["-o", str(binary), package], cwd=ROOT,
                           check=True, capture_output=True, text=True)
        assert binary.is_absolute() and binary.is_file()
        binaries.append(binary)
    return binaries


@pytest.fixture(scope="module")
def tooling_server(tooling_binaries, tmp_path_factory):
    # Same bounded ephemeral-port/readiness/child-cleanup recipe as the accepted
    # test_surfaces harness, without importing that module's native dependency.
    log_path = tmp_path_factory.mktemp("tooling-server") / "server.log"
    for attempt in range(3):
        with socket.socket() as listener:
            listener.bind(("127.0.0.1", 0))
            port = listener.getsockname()[1]
        with log_path.open("w+", encoding="utf-8") as log:
            child = subprocess.Popen([str(tooling_binaries[1])], cwd=log_path.parent,
                                     env=dict(os.environ, PORT=str(port)), stdout=log, stderr=log)
            base = f"http://127.0.0.1:{port}/api/v1"
            ready = False
            try:
                deadline = time.monotonic() + 5
                while child.poll() is None and time.monotonic() < deadline:
                    try:
                        with urlopen(base + "/health", timeout=1) as response:
                            ready = response.status == 200
                        if ready:
                            break
                    except (OSError, URLError):
                        time.sleep(0.05)
                if ready:
                    yield base
                    return
            finally:
                if child.poll() is None:
                    child.terminate()
                    try:
                        child.wait(timeout=5)
                    except subprocess.TimeoutExpired:
                        child.kill()
                        child.wait(timeout=5)
                assert child.poll() is not None, "server cleanup failed"
    pytest.fail(f"server failed readiness: {log_path.read_text()}")


def post(base, route, value):
    request = Request(base + "/" + route, data=json.dumps(value).encode(),
                      headers={"Content-Type": "application/json"})
    with urlopen(request, timeout=30) as response:
        assert response.status == 200
        return json.load(response)


def cli(binary, *args):
    result = subprocess.run([str(binary), *args], capture_output=True, text=True)
    assert not result.stderr, result.stderr
    return result.returncode, json.loads(result.stdout)


def test_corpus_graph_sarif_and_impact(tooling_binaries, tooling_server, tmp_path):
    tmp_path = tmp_path.resolve()  # Anchored loader requires physical ancestors.
    docs = [{"id": "valid", "document": {"text": 'search host="😀"\r\n| table host', "source_id": "valid"}},
            {"id": "missing", "document": {"text": "search absent=x", "source_id": "missing"}},
            {"id": "unknown", "document": {"text": "| mystery | table host", "source_id": "unknown"}}]
    manifest = tmp_path / "manifest.json"
    manifest.write_text(json.dumps({"schema_version": 1, "documents": [
        {"id": item["id"], **item["document"]} for item in docs]}), encoding="utf-8")
    before = {"kind": "field_list", "catalog": {"fields": ["host", "absent"]}}
    after = {"kind": "field_list", "catalog": {"fields": ["host"]}}
    files = {}
    for name, value in (("before", before), ("after", after),
                        ("rules", {"schema_version": 1, "rules": []})):
        files[name] = tmp_path / f"{name}.json"
        files[name].write_text(json.dumps(value), encoding="utf-8")
    snapshot = {p: p.read_bytes() for p in (manifest, *files.values())}
    request = {"schema_version": 1, "documents": docs, "validation_target": after}
    for command, format_, route in (("scan", "json", "scan"), ("graph", "json", "graph"),
                                    ("scan", "sarif", "sarif")):
        code, local = cli(tooling_binaries[0], command, "--manifest", str(manifest),
                          "--target", str(files["after"]), "--format", format_)
        remote = post(tooling_server, "corpus/" + route, request)
        assert code == 1
        if route == "scan":
            assert local["selection"]["mode"] == "manifest"
            assert remote["selection"]["mode"] == "inline"
            assert local["entries"] == remote["entries"]
            for key in local.keys() - {"selection", "entries"}:
                assert local[key] == remote[key]
            assert local["status"] == "invalid" and local["counts"]["analyzed"] == 3
        else:
            assert local == remote
    for kind in ("schema", "mapping"):
        request = {"schema_version": 1, "documents": docs}
        args = ["impact-" + kind, "--manifest", str(manifest), "--format", "json"]
        if kind == "schema":
            request.update(before_target=before, after_target=after)
            args += ["--before-target", str(files["before"]), "--after-target", str(files["after"])]
        else:
            rules = {"schema_version": 1, "rules": []}
            request.update(before_rules=rules, after_rules=rules)
            args += ["--before-rules", str(files["rules"]), "--after-rules", str(files["rules"])]
        code, local = cli(tooling_binaries[0], *args)
        remote = post(tooling_server, "corpus/impact-" + kind, request)
        assert code == (1 if kind == "schema" else 3)
        assert local["selection"]["mode"] == "manifest" and remote["selection"]["mode"] == "inline"
        for key in local.keys() - {"selection"}:
            assert local[key] == remote[key]
    assert snapshot == {p: p.read_bytes() for p in snapshot}


def test_document_and_transport_errors(tooling_binaries, tooling_server):
    text = ' search host="😀"\r\n| table host\n'
    code, local = cli(tooling_binaries[0], "document", "--query", text, "--source-id", "exact")
    assert code == 0 and local["document"]["text"] == text
    assert local == post(tooling_server, "query/document", {"text": text, "source_id": "exact"})
    for route in ("corpus/scan", "corpus/graph", "corpus/sarif", "corpus/impact-schema",
                  "corpus/impact-mapping", "query/document"):
        with pytest.raises(HTTPError) as error:
            post(tooling_server, route, {"directory": "/", "schema_version": 1})
        assert error.value.code == 400
    help_result = subprocess.run([str(tooling_binaries[0]), "help"], capture_output=True, text=True, check=True)
    assert "scan --directory" in help_result.stdout and "lsp --stdio" in help_result.stdout
