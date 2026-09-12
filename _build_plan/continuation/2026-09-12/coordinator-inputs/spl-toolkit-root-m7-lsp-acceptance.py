"""Prepared local LSP transcript; actual configuration names require final M7 reconciliation.

This is independent stdio acceptance, not the required real-editor consumer evidence.
"""
import hashlib
import json
import os
from pathlib import Path
import queue
import subprocess
import sys
import threading
import time

root = Path(sys.argv[1]).resolve()
expected_head = sys.argv[2]
work = Path('/private/tmp/spl-toolkit-root-m7-lsp')
work.mkdir(exist_ok=True)
binary = root / 'build/spl-toolkit'


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def head():
    return subprocess.check_output(['git', 'rev-parse', 'HEAD'], cwd=root, text=True).strip()


def clean():
    for command in [['git', 'diff', '--quiet'], ['git', 'diff', '--cached', '--quiet']]:
        subprocess.run(command, cwd=root, check=True)
    paths = subprocess.check_output(['git', 'ls-files', '--others', '--exclude-standard', '-z'], cwd=root).decode().split('\0')
    product = {'pkg', 'parser', 'cmd', 'grammar', 'tools', 'testdata', 'tests', 'python', 'internal', 'Makefile', 'go.mod', 'go.sum', 'contracts'}
    assert not [path for path in paths if path and path.split('/')[0] in product]


assert head() == expected_head
clean()
tracked = subprocess.check_output(['git', 'ls-files', '-z'], cwd=root).decode().split('\0')
source_hashes = {path: digest(root / path) for path in tracked if path}
binary_hash = digest(binary)
inbox = queue.Queue()
transcript = []
unsaved_path = work / 'unsaved-é-🙂.spl'
unsaved_path.write_text('search disk_only=x | table disk_only')
disk_hash = digest(unsaved_path)
uri = unsaved_path.as_uri()
valid = 'eval marker="🙂", alias=host | table alias'
invalid = 'eval marker="🙂"\r\n| eval alias='


def position(text, offset):
    prefix = text.encode()[:offset].decode()
    lines = prefix.replace('\r\n', '\n').replace('\r', '\n').split('\n')
    return {'line': len(lines) - 1, 'character': len(lines[-1].encode('utf-16-le')) // 2}


def expected_highlights(text):
    first = text.index('alias')
    last = text.rindex('alias')
    values = []
    for start, kind in [(first, 3), (last, 2)]:
        begin = len(text[:start].encode())
        values.append({'range': {'start': position(text, begin), 'end': position(text, begin + len('alias'))}, 'kind': kind})
    return values


def diagnostic_core(value):
    return {key: value[key] for key in ('code', 'message', 'severity', 'range')}


def canonical_diagnostics(text):
    result = subprocess.run([str(binary), 'analyze', '--query', text, '--language', 'spl', '--format', 'json'], capture_output=True, timeout=90)
    report = json.loads(result.stdout)
    assert result.returncode == {'valid': 0, 'invalid': 1, 'incomplete': 3}[report['status']]
    values = []
    for diagnostic in report['diagnostics']:
        location = diagnostic['location']
        values.append({'code': diagnostic['code'], 'message': diagnostic['message'],
                       'severity': {'error': 1, 'warning': 2, 'information': 3, 'hint': 4}[diagnostic['severity']],
                       'range': {'start': position(text, location['start']['offset']),
                                 'end': position(text, location['end']['offset'])}})
    return values


with (work / 'stderr.log').open('wb') as stderr:
    process = subprocess.Popen([str(binary), 'lsp', '--stdio'], stdin=subprocess.PIPE, stdout=subprocess.PIPE, stderr=stderr, env=dict(os.environ))

    def receive_frames():
        try:
            while True:
                headers = {}
                while True:
                    line = process.stdout.readline()
                    if not line:
                        inbox.put(EOFError('LSP stdout closed'))
                        return
                    if line == b'\r\n':
                        break
                    assert line.endswith(b'\r\n'), ('Unframed stdout', line)
                    name, value = line[:-2].split(b':', 1)
                    name = name.lower()
                    assert name not in headers, 'Duplicate frame header'
                    headers[name] = value.strip()
                assert set(headers) <= {b'content-length', b'content-type'}
                size = int(headers[b'content-length'])
                assert 0 < size <= 8 * 1024 * 1024
                payload = process.stdout.read(size)
                assert len(payload) == size, 'Truncated response frame'
                value = json.loads(payload)
                assert value.get('jsonrpc') == '2.0'
                inbox.put(value)
        except BaseException as error:
            inbox.put(error)

    reader = threading.Thread(target=receive_frames, daemon=True)
    reader.start()

    def send_value(value):
        payload = json.dumps(value, ensure_ascii=False).encode()
        process.stdin.write(f'Content-Length: {len(payload)}\r\n\r\n'.encode() + payload)
        process.stdin.flush()
        transcript.append({'direction': 'client', 'message': value})

    def send(method, params=None, request_id=None):
        value = {'jsonrpc': '2.0', 'method': method}
        if params is not None:
            value['params'] = params
        if request_id is not None:
            value['id'] = request_id
        send_value(value)

    def wait_for(predicate, timeout=30):
        end = time.monotonic() + timeout
        while True:
            remaining = end - time.monotonic()
            assert remaining > 0, 'Expected LSP response/notification did not arrive'
            value = inbox.get(timeout=remaining)
            if isinstance(value, BaseException):
                raise value
            transcript.append({'direction': 'server', 'message': value})
            if predicate(value):
                return value

    def response(request_id):
        return wait_for(lambda value: value.get('id') == request_id)

    def diagnostics(version=None):
        return wait_for(lambda value: value.get('method') == 'textDocument/publishDiagnostics'
                        and value['params']['uri'] == uri
                        and (version is None or value['params'].get('version') == version))['params']['diagnostics']

    def change(version, text):
        send('textDocument/didChange', {'textDocument': {'uri': uri, 'version': version}, 'contentChanges': [{'text': text}]})

    try:
        send('initialize', {'processId': os.getpid(), 'rootUri': work.as_uri(),
                           'capabilities': {'general': {'positionEncodings': ['utf-8', 'utf-16']},
                                            'textDocument': {'publishDiagnostics': {'versionSupport': True}}},
                           'initializationOptions': {'profile': 'splunkd', 'version': 'current'}}, 1)
        initialized = response(1)
        assert 'error' not in initialized, initialized
        capabilities = initialized['result']['capabilities']
        assert capabilities['positionEncoding'] == 'utf-16'
        sync = capabilities['textDocumentSync']
        assert (sync if isinstance(sync, int) else sync['change']) == 1
        assert capabilities['documentHighlightProvider']
        send('initialized', {})
        # The supported LSP transport accepts one message object per frame.
        for batch_payload in [[], [{'jsonrpc': '2.0', 'id': 91, 'method': 'root/batchElementMustNotRun', 'params': {}}]]:
            send_value(batch_payload)
            rejected = wait_for(lambda value: 'id' in value and value['id'] is None)
            assert rejected['error']['code'] == -32600 and 'result' not in rejected
        send('textDocument/didOpen', {'textDocument': {'uri': uri, 'languageId': 'spl', 'version': 1, 'text': valid}})
        assert diagnostics(1) == []
        highlight_params = {'textDocument': {'uri': uri}, 'position': expected_highlights(valid)[0]['range']['start']}
        send('textDocument/documentHighlight', highlight_params, 2)
        highlights = response(2)
        assert highlights['result'] == expected_highlights(valid), highlights
        change(2, invalid)
        delivered = diagnostics(2)
        assert delivered and [diagnostic_core(value) for value in delivered] == canonical_diagnostics(invalid)
        change(3, valid)
        assert diagnostics(3) == []
        # A bad configuration must not quietly reuse a target or fall back to ordinary analysis.
        send('workspace/didChangeConfiguration', {'settings': {'splToolkit': {'version': 'unsupported-root-version'}}})
        configuration_events = set()
        def invalid_configuration_observed(value):
            if value.get('method') == 'window/showMessage' and value['params']['type'] == 1:
                configuration_events.add('error')
            if value.get('method') == 'textDocument/publishDiagnostics' and value['params']['uri'] == uri and value['params']['diagnostics'] == []:
                configuration_events.add('cleared')
            return configuration_events == {'error', 'cleared'}
        wait_for(invalid_configuration_observed)
        send('textDocument/documentHighlight', highlight_params, 3)
        assert 'error' in response(3), 'Configuration-dependent request succeeded while configuration was invalid'
        change(4, invalid)
        send('workspace/didChangeConfiguration', {'settings': {'splToolkit': {'profile': 'splunkd', 'version': 'current'}}})
        assert [diagnostic_core(value) for value in diagnostics(4)] == canonical_diagnostics(invalid)
        send('textDocument/didClose', {'textDocument': {'uri': uri}})
        assert diagnostics() == []
        send('root/unknownMethod', {}, 4)
        assert response(4)['error']['code'] == -32601
        send('shutdown', request_id=5)
        assert response(5).get('result', 'missing') is None
        send('exit')
        process.stdin.close()
        assert process.wait(timeout=10) == 0
        reader.join(timeout=10)
        assert not reader.is_alive()
        while not inbox.empty():
            remaining = inbox.get_nowait()
            if isinstance(remaining, EOFError):
                continue
            if isinstance(remaining, BaseException):
                raise remaining
            transcript.append({'direction': 'server', 'message': remaining})
        responses = [item['message']['id'] for item in transcript if item['direction'] == 'server' and 'id' in item['message']]
        assert responses.count(None) == 2, 'Array rejection did not produce exactly two null-ID errors'
        assert sorted(value for value in responses if value is not None) == [1, 2, 3, 4, 5], 'Missing/duplicate response or an array element was dispatched'
        assert digest(unsaved_path) == disk_hash, 'LSP changed the disk source'
    finally:
        if process.poll() is None:
            process.terminate()
            try:
                process.wait(timeout=10)
            except subprocess.TimeoutExpired:
                process.kill()
                process.wait(timeout=10)
        (work / 'transcript.json').write_text(json.dumps(transcript, ensure_ascii=False, indent=2) + '\n')

assert head() == expected_head
clean()
assert binary_hash == digest(binary)
assert source_hashes == {path: digest(root / path) for path in source_hashes}
evidence = {'status': 'passed', 'source_sha': expected_head,
            'scope': 'Independent local stdio LSP transcript; no real-editor or deterministic concurrency-race claim',
            'checks': ['Full-sync UTF16 capability negotiation', 'Unsaved-buffer authority', 'Read/write alias highlights after emoji',
                       'Exact canonical diagnostic code/message/severity and UTF16 CRLF ranges', 'Diagnostics clearing and version',
                       'Invalid configuration visible error, cleared diagnostics and request refusal', 'Corrected configuration latest-buffer recovery',
                       'Close clearing', 'Unknown method and clean shutdown/exit', 'Exactly one response for each request ID',
                       'Top-level empty/nonempty arrays rejected without dispatching embedded elements'],
            'binary_sha256': binary_hash, 'tracked_source_sha256': source_hashes,
            'harness_sha256': digest(Path(__file__)), 'transcript_sha256': digest(work / 'transcript.json'),
            'stderr_sha256': digest(work / 'stderr.log')}
output = Path('/private/tmp/spl-toolkit-root-m7-lsp-acceptance.json')
output.write_text(json.dumps(evidence, indent=2) + '\n')
print('Evidence:', output)
