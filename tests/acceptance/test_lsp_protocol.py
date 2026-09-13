"""Real framed stdio LSP 3.17 transcript, without an editor dependency.

SPL_LSP_BINARY selects a built CLI (invoked with ``lsp --stdio``). Without it,
the fixture compiles the Go stdio test entry point, which calls production Serve.
SPL_LSP_GO may select an offline Go toolchain. Actual editor-client delivery is a
separate acceptance gate; this harness does not represent VS Code acceptance.
"""
from __future__ import annotations

import json
import os
from pathlib import Path
import queue
import subprocess
import threading

import pytest


ROOT = Path(__file__).resolve().parents[2]


@pytest.fixture(scope="module")
def lsp_command(tmp_path_factory):
    binary = os.environ.get("SPL_LSP_BINARY")
    if binary:
        return [binary, "lsp", "--stdio"], os.environ.copy()
    binary = tmp_path_factory.mktemp("lsp-protocol") / "lsp.test"
    subprocess.run(
        [os.environ.get("SPL_LSP_GO", "go"), "test", "-mod=readonly", "-c",
         "-o", str(binary), "./internal/lsp"],
        cwd=ROOT, check=True, capture_output=True, text=True,
    )
    env = dict(os.environ, SPL_LSP_STDIO_HELPER="1")
    return [str(binary), "-test.run=^TestStdioHelper$"], env


class Client:
    def __init__(self, command, env):
        self.process = subprocess.Popen(command, env=env, stdin=subprocess.PIPE,
                                        stdout=subprocess.PIPE, stderr=subprocess.PIPE)
        self.messages = queue.Queue()
        self.thread = threading.Thread(target=self._read, daemon=True)
        self.thread.start()

    def _read(self):
        try:
            while True:
                first = self.process.stdout.readline()
                if not first:
                    return
                assert first.startswith(b"Content-Length: "), first
                length = int(first[len(b"Content-Length: "):])
                assert self.process.stdout.readline() == b"\r\n"
                body = self.process.stdout.read(length)
                assert len(body) == length
                self.messages.put(json.loads(body))
        except Exception as exc:
            self.messages.put(exc)

    def send(self, method, params=None, *, id=None):
        value = {"jsonrpc": "2.0", "method": method}
        if params is not None:
            value["params"] = params
        if id is not None:
            value["id"] = id
        self.raw(json.dumps(value, ensure_ascii=False).encode())

    def raw(self, body):
        self.process.stdin.write(f"Content-Length: {len(body)}\r\n\r\n".encode() + body)
        self.process.stdin.flush()

    def receive(self):
        value = self.messages.get(timeout=15)
        if isinstance(value, Exception):
            raise value
        return value

    def close(self):
        if self.process.poll() is None:
            self.process.kill()
        self.process.wait(timeout=10)
        self.thread.join(timeout=10)
        for stream in (self.process.stdin, self.process.stdout, self.process.stderr):
            stream.close()


@pytest.fixture
def client(lsp_command):
    client = Client(*lsp_command)
    try:
        yield client
    finally:
        client.close()


def test_real_stdio_lifecycle_and_unicode(client):
    client.send("initialize", {"capabilities": {"textDocument": {
        "publishDiagnostics": {"versionSupport": True}}}}, id=9007199254740993)
    initialized = client.receive()
    assert initialized["id"] == 9007199254740993
    caps = initialized["result"]["capabilities"]
    assert caps["positionEncoding"] == "utf-16"
    assert caps["textDocumentSync"] == {"openClose": True, "change": 1}
    assert caps["documentHighlightProvider"] is True
    client.send("initialized", {})
    uri = "file:///buffer.spl"
    text = 'eval emoji="😀", a=host | table a'
    client.send("textDocument/didOpen", {"textDocument": {
        "uri": uri, "languageId": "spl", "version": 1, "text": text}})
    published = client.receive()
    assert published["method"] == "textDocument/publishDiagnostics"
    assert published["params"]["version"] == 1
    assert published["params"]["diagnostics"] == []
    # The non-BMP character occupies two UTF-16 units before assignment a.
    client.send("textDocument/documentHighlight", {"textDocument": {"uri": uri},
                "position": {"line": 0, "character": 17}}, id="highlight")
    response = client.receive()
    assert response["id"] == "highlight"
    locations = {(h["range"]["start"]["character"], h["kind"])
                 for h in response["result"]}
    assert (17, 3) in locations
    assert (32, 2) in locations
    client.send("textDocument/didChange", {"textDocument": {"uri": uri, "version": 2},
                "contentChanges": [{"text": "eval a="}]})
    published = client.receive()
    assert published["params"]["version"] == 2
    assert published["params"]["diagnostics"]
    client.send("textDocument/didClose", {"textDocument": {"uri": uri}})
    assert client.receive()["params"]["diagnostics"] == []
    client.send("unknownNotification")
    client.send("unknownRequest", id="unknown")
    assert client.receive()["error"]["code"] == -32601
    client.raw(b"[]")
    assert client.receive()["error"]["code"] == -32600
    client.raw(b"{")
    assert client.receive()["error"]["code"] == -32700
    client.send("shutdown", id="shutdown")
    assert client.receive() == {"jsonrpc": "2.0", "id": "shutdown", "result": None}
    client.send("exit")
    assert client.process.wait(timeout=10) == 0
    assert client.process.stderr.read() == b""


def test_exit_before_shutdown_is_failure(client):
    client.send("exit")
    assert client.process.wait(timeout=10) != 0
    assert b"exit before shutdown" in client.process.stderr.read()


def test_oversized_frame_terminates_without_stdout(client):
    client.process.stdin.write(b"Content-Length: 67108865\r\n\r\n")
    client.process.stdin.flush()
    assert client.process.wait(timeout=10) != 0
    assert b"Content-Length exceeds" in client.process.stderr.read()
    assert client.messages.empty()
