"""Prepared M6 surface runner; reconcile actual APIs before the final locked window."""
import concurrent.futures
import copy
import hashlib
import json
import os
from pathlib import Path
import socket
import subprocess
import sys
import time
import urllib.error
import urllib.request

root = Path(sys.argv[1]).resolve()
expected_head = sys.argv[2]
prefix = Path('/private/tmp')
work = prefix / 'spl-toolkit-root-m6-surfaces'
work.mkdir(exist_ok=True)


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def head():
    return subprocess.check_output(['git', 'rev-parse', 'HEAD'], cwd=root, text=True).strip()


def clean():
    subprocess.run(['git', 'diff', '--quiet'], cwd=root, check=True)
    subprocess.run(['git', 'diff', '--cached', '--quiet'], cwd=root, check=True)
    paths = subprocess.check_output(['git', 'ls-files', '--others', '--exclude-standard', '-z'], cwd=root).decode().split('\0')
    product = {'pkg', 'parser', 'cmd', 'grammar', 'tools', 'testdata', 'tests', 'python', 'internal', 'Makefile', 'go.mod', 'go.sum'}
    assert not [path for path in paths if path and path.split('/')[0] in product], 'Uncommitted product/test input'


def encode(value):
    return json.dumps(value, ensure_ascii=False, allow_nan=False).encode()


assert head() == expected_head
clean()
subprocess.run([sys.executable, str(prefix / 'spl-toolkit-root-m6-go-acceptance.py'), str(root), expected_head], cwd=root, check=True)
canonical_path = prefix / 'spl-toolkit-root-m6-go-results.json'
baseline_path = prefix / 'spl-toolkit-root-m6-go-results-task5.json'
canonical = json.loads(canonical_path.read_text())
assert canonical == json.loads(baseline_path.read_text()), 'Canonical reports changed after root Task5 acceptance; inspect before accepting'
case_path = prefix / 'spl-toolkit-root-m6-cases.json'
validation_path = prefix / 'spl-toolkit-root-m6-validation-cases.json'
cases = json.loads(case_path.read_text())['cases']
validation = json.loads(validation_path.read_text())
cases.extend(validation['cases'])
batch = validation['batch']['request']
input_paths = {Path(__file__), case_path, validation_path, baseline_path, canonical_path}
for case in cases:
    if 'catalog_path' in case:
        path = Path(case['catalog_path'])
        assert digest(path) == case['catalog_sha256']
        case['request']['validation_target']['catalog'] = json.loads(path.read_text())
        input_paths.add(path)
input_hashes = {str(path): digest(path) for path in sorted(input_paths)}
tracked = subprocess.check_output(['git', 'ls-files', '-z'], cwd=root).decode().split('\0')
source_hashes = {path: digest(root / path) for path in tracked if path}
artifact_hashes = {name: digest(root / 'build' / name) for name in ['spl-toolkit', 'spl-toolkit-server', 'libspl_toolkit.dylib']}
sys.path.insert(0, str(root / 'python'))
from spl_toolkit import SPLMapper, SPLMapperError


def write_json(name, value):
    path = work / name
    path.write_bytes(encode(value) + b'\n')
    return str(path)


def selector_flags(document):
    flags = []
    for key, option in [('language', '--language'), ('profile', '--profile'), ('version', '--compatibility-version'), ('source_id', '--source-id')]:
        flags += [option, document[key]]
    return flags


def target_flags(name, target):
    if target is None:
        return []
    if target['kind'] == 'field_list':
        return ['--fields', write_json(name + '-fields.json', target['catalog'])]
    assert not target.get('identity'), 'Independent targets use a surface-neutral default identity'
    if target['kind'] == 'json_schema':
        flags = ['--schema', write_json(name + '-schema.json', target['schema'])]
        if 'base_uri' in target:
            flags += ['--schema-base-uri', target['base_uri']]
        if 'resources' in target:
            flags += ['--schema-resources', write_json(name + '-resources.json', target['resources'])]
        return flags
    assert target['kind'] == 'ocsf'
    selection = target['selection']
    flags = ['--ocsf-catalog', write_json(name + '-catalog.json', target['catalog']), '--ocsf-version', selection['version']]
    for key, option in [('class', '--ocsf-class'), ('class_uid', '--ocsf-class'), ('category', '--ocsf-category'), ('category_uid', '--ocsf-category')]:
        if key in selection:
            flags += [option, str(selection[key])]
    for key, option in [('profiles', '--ocsf-profile'), ('extensions', '--ocsf-extension')]:
        for value in selection.get(key, []):
            flags += [option, value]
    return flags


def operation_flags(name, value):
    flags = ['--rules', write_json(name + '-rules.json', {'schema_version': value['schema_version'], 'rules': value['rules']})]
    if value.get('mode', 'preview') == 'apply':
        flags += ['--apply']
    flags += target_flags(name, value.get('validation_target'))
    return flags


def cli(arguments, expected=None, data=None, output=None, error=False):
    command = [str(root / 'build/spl-toolkit'), 'rewrite', *arguments, '--format', 'json']
    result = subprocess.run(command, input=data, capture_output=True, timeout=90)
    if error:
        assert result.returncode == 2 and result.stdout == b'' and result.stderr, result
        return
    assert result.returncode == {'valid': 0, 'invalid': 1, 'incomplete': 3}[expected['status']], (result.returncode, result.stderr)
    assert result.stderr == b'', result.stderr
    if output is not None:
        assert result.stdout == b'', 'Output report duplicated on stdout'
        payload = output.read_bytes()
    else:
        payload = result.stdout
    assert json.loads(payload) == expected, arguments


with socket.socket() as listener:
    listener.bind(('127.0.0.1', 0))
    port = listener.getsockname()[1]
base = f'http://127.0.0.1:{port}/api/v1'


def request(path, payload=None, status=200):
    req = urllib.request.Request(base + path, data=payload, headers={'Content-Type': 'application/json'})
    try:
        response = urllib.request.urlopen(req, timeout=60)
    except urllib.error.HTTPError as error:
        assert error.code == status, (error.code, error.read())
        value = json.load(error)
        assert not {'reports', 'changes', 'candidate_analysis', 'text'} & value.keys(), value
        return value
    with response:
        assert response.status == status, response.status
        return json.load(response)


def native_call(mapper, value, batch_mode=False):
    keywords = {'mode': value.get('mode', 'preview')}
    if 'validation_target' in value:
        keywords['validation_target'] = value['validation_target']
    if batch_mode:
        return mapper.rewrite_batch(value['documents'], value['rules'], **keywords)
    document = value['document']
    keywords.update({key: document[key] for key in ('language', 'profile', 'version', 'source_id')})
    return mapper.rewrite(document['text'], value['rules'], **keywords)


def native_rejected(call):
    try:
        call()
    except SPLMapperError:
        return
    raise AssertionError('Native operation accepted an invalid request')


strict_invalid = []
simple = cases[0]['request']
for patch in [{'schema_version': 2}, {'mode': None}, {'mode': 'execute'}, {'unknown': True}]:
    value = copy.deepcopy(simple)
    value.update(patch)
    strict_invalid.append(encode(value))
value = copy.deepcopy(simple)
value['rules'].append(copy.deepcopy(value['rules'][0]))
strict_invalid.append(encode(value))
value = copy.deepcopy(simple)
value['rules'][0]['source']['path'] = ['src']
strict_invalid.append(encode(value))
strict_invalid.append(encode(simple).replace(b'"schema_version": 1', b'"schema_version": 1, "schema_version": 1', 1))
strict_invalid.append(encode(simple).replace(b'"language": "spl"', b'"language": "\\ud800"', 1))
assert len(set(strict_invalid)) == 8

with (work / 'server.log').open('w') as log:
    server = subprocess.Popen([str(root / 'build/spl-toolkit-server')], env=os.environ | {'PORT': str(port)}, stdout=log, stderr=log)
    try:
        for _ in range(150):
            if server.poll() is not None:
                raise RuntimeError('Temporary server exited; inspect server.log')
            try:
                urllib.request.urlopen(base + '/health', timeout=1).close()
                break
            except OSError:
                time.sleep(0.1)
        else:
            raise RuntimeError('Temporary server startup timed out')
        with SPLMapper(library_path=str(root / 'build/libspl_toolkit.dylib')) as mapper:
            for case in cases:
                name, value = case['name'], case['request']
                document, expected = value['document'], canonical['cases'][name]
                before = encode(value)
                flags = operation_flags(name, value)
                cli(['--query', document['text'], *selector_flags(document), *flags], expected)
                assert request('/query/rewrite', encode(value)) == expected, name
                assert native_call(mapper, value) == expected, name
                assert encode(value) == before, ('Caller request mutated', name)
                print(name + ': full canonical CLI/HTTP/native rewrite and audit parity passed', flush=True)
            jobs = cases * 3
            with concurrent.futures.ThreadPoolExecutor(max_workers=6) as pool:
                results = list(pool.map(lambda case: native_call(mapper, case['request']), jobs))
            for case, report in zip(jobs, results):
                assert report == canonical['cases'][case['name']], ('Concurrent ownership', case['name'])
            case = cases[0]
            value, expected = case['request'], canonical['cases'][case['name']]
            path = work / 'unicode-crlf.spl'
            path.write_bytes(value['document']['text'].encode())
            original_bytes = path.read_bytes()
            flags = operation_flags('transports', value)
            cli(['--file', str(path), *selector_flags(value['document']), *flags], expected)
            cli(['--stdin', *selector_flags(value['document']), *flags], expected, original_bytes)
            output = work / 'rewrite-report.json'
            cli(['--file', str(path), *selector_flags(value['document']), *flags, '--output', str(output)], expected, output=output)
            assert path.read_bytes() == original_bytes, 'Apply changed the source file'
            cli(['--file', str(path), *selector_flags(value['document']), *flags, '--output', str(work)], error=True)
            batch_path = write_json('batch-documents.json', batch['documents'])
            batch_before = encode(batch)
            batch_flags = operation_flags('batch', batch)
            cli(['--batch', batch_path, *batch_flags], canonical['batch'])
            cli(['--batch', '-', *batch_flags], canonical['batch'], Path(batch_path).read_bytes())
            assert request('/query/rewrite/batch', encode(batch)) == canonical['batch']
            assert native_call(mapper, batch, batch_mode=True) == canonical['batch']
            assert encode(batch) == batch_before
            bad_batch = copy.deepcopy(batch)
            bad_batch['documents'][-1]['version'] = 'unsupported-root-version'
            request('/query/rewrite/batch', encode(bad_batch), 400)
            native_rejected(lambda: native_call(mapper, bad_batch, batch_mode=True))
            bad_batch_path = write_json('bad-batch-documents.json', bad_batch['documents'])
            cli(['--batch', bad_batch_path, *batch_flags], error=True)
            cli(['--batch', batch_path, '--language', 'spl', *batch_flags], error=True)
            for payload in strict_invalid:
                request('/query/rewrite', payload, 400)
            duplicate_rules = copy.deepcopy(simple)
            duplicate_rules['rules'].append(copy.deepcopy(duplicate_rules['rules'][0]))
            native_rejected(lambda: native_call(mapper, duplicate_rules))
            wrong_mode = copy.deepcopy(simple)
            wrong_mode['mode'] = 'execute'
            native_rejected(lambda: native_call(mapper, wrong_mode))
    finally:
        server.terminate()
        try:
            server.wait(timeout=10)
        except subprocess.TimeoutExpired:
            server.kill()
            server.wait(timeout=10)

assert head() == expected_head
clean()
assert source_hashes == {path: digest(root / path) for path in source_hashes}
assert artifact_hashes == {name: digest(root / 'build' / name) for name in artifact_hashes}
assert input_hashes == {path: digest(Path(path)) for path in input_hashes}
evidence = {
    'status': 'passed', 'source_sha': expected_head,
    'scope': 'Independent local Go/CLI/HTTP/source-native rewrite acceptance; installed and release checks separate',
    'cases': [case['name'] for case in cases], 'batch_documents': len(batch['documents']),
    'concurrent_native_calls': len(jobs), 'strict_invalid_http_requests': len(strict_invalid),
    'exact_original_candidate_validation_audit_parity': 'passed', 'field_list_target_wire_projection': 'Only fields and optional_fields omitted; original kind, identity and version retained, including empty strings', 'root_task5_full_result_parity': 'passed',
    'caller_ownership': 'passed', 'file_stdin_output_source_preservation': 'passed',
    'late_invalid_batch_atomicity': 'passed', 'output_write_failure': 'passed',
    'input_sha256': input_hashes, 'tracked_source_sha256': source_hashes,
    'artifact_sha256': artifact_hashes, 'server_log': str(work / 'server.log'),
}
evidence_path = prefix / 'spl-toolkit-root-m6-surface-acceptance.json'
evidence_path.write_text(json.dumps(evidence, indent=2) + '\n')
print('Evidence:', evidence_path, flush=True)
