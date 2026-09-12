"""Prepared M6 canonical runner; execute only after reviewed Task5 and wire reconciliation."""
import copy
import hashlib
import importlib.util
import json
from pathlib import Path
import subprocess
import sys

root = Path(sys.argv[1]).resolve()
expected_head = sys.argv[2]
base = Path('/private/tmp')


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def clean():
    subprocess.run(['git', 'diff', '--quiet'], cwd=root, check=True)
    subprocess.run(['git', 'diff', '--cached', '--quiet'], cwd=root, check=True)
    paths = subprocess.check_output(['git', 'ls-files', '--others', '--exclude-standard', '-z'], cwd=root).decode().split('\0')
    product = {'pkg', 'parser', 'cmd', 'grammar', 'tools', 'testdata', 'tests', 'python', 'internal', 'Makefile', 'go.mod', 'go.sum'}
    assert not [path for path in paths if path and path.split('/')[0] in product], 'Uncommitted product/test input'


def head():
    return subprocess.check_output(['git', 'rev-parse', 'HEAD'], cwd=root, text=True).strip()


assert head() == expected_head
clean()
paths = {key: base / ('spl-toolkit-root-m6-' + suffix) for key, suffix in {
    'cases': 'cases.json', 'validation': 'validation-cases.json', 'driver': 'go.go', 'oracle': 'semantic-checks.py'}.items()}
cases = json.loads(paths['cases'].read_text())['cases']
validation = json.loads(paths['validation'].read_text())
cases.extend(validation['cases'])
input_paths = set(paths.values()) | {Path(__file__)}
for case in cases:
    if 'catalog_path' in case:
        path = Path(case['catalog_path'])
        assert digest(path) == case['catalog_sha256']
        case['request']['validation_target']['catalog'] = json.loads(path.read_text())
        input_paths.add(path)
input_hashes = {str(path): digest(path) for path in sorted(input_paths)}
tracked = subprocess.check_output(['git', 'ls-files', '-z'], cwd=root).decode().split('\0')
source_hashes = {path: digest(root / path) for path in tracked if path}
simple = cases[0]['request']
invalid = []
for patch in [{'schema_version': 2}, {'mode': None}, {'mode': 'execute'}, {'unknown': True}]:
    value = copy.deepcopy(simple)
    value.update(patch)
    invalid.append(json.dumps(value))
value = copy.deepcopy(simple)
value['rules'].append(copy.deepcopy(value['rules'][0]))
invalid.append(json.dumps(value))
value = copy.deepcopy(simple)
value['rules'][0]['source']['path'] = ['src']
invalid.append(json.dumps(value))
invalid.append(json.dumps(simple).replace('"schema_version": 1', '"schema_version": 1, "schema_version": 1', 1))
invalid.append(json.dumps(simple).replace('"language": "spl"', '"language": "\\ud800"', 1))
request = {'cases': cases, 'batch': validation['batch']['request'], 'invalid_requests': invalid}
result = subprocess.run(['go', 'run', '-mod=readonly', str(paths['driver'])], cwd=root,
                        input=json.dumps(request, ensure_ascii=False).encode(), capture_output=True, timeout=300)
raw_output_path = base / 'spl-toolkit-root-m6-go-attempt.stdout.json'
raw_error_path = base / 'spl-toolkit-root-m6-go-attempt.stderr.txt'
raw_output_path.write_bytes(result.stdout)
raw_error_path.write_bytes(result.stderr)
(base / 'spl-toolkit-root-m6-go-attempt.json').write_text(json.dumps(dict(
    status='executed-not-yet-accepted', source_sha=expected_head, cwd=str(root),
    argv=['go', 'run', '-mod=readonly', str(paths['driver'])], exit_code=result.returncode,
    input_sha256=input_hashes, tracked_source_sha256=source_hashes,
    stdout_path=str(raw_output_path), stdout_sha256=digest(raw_output_path),
    stderr_path=str(raw_error_path), stderr_sha256=digest(raw_error_path)), indent=2) + '\n')
if result.returncode:
    raise RuntimeError(result.stderr.decode(errors='replace'))
reports = json.loads(result.stdout)
spec = importlib.util.spec_from_file_location('m6_semantic_oracles', paths['oracle'])
module = importlib.util.module_from_spec(spec)
spec.loader.exec_module(module)
for case in cases:
    name = case['name']
    module.check(case, reports['cases'][name], reports['original_analyses'][name], reports['candidate_analyses'][name], reports['candidate_validations'].get(name))
    print(name + ': exact text, edit translation, canonical provenance, and apply gate passed', flush=True)
batch = validation['batch']
assert reports['batch']['status'] == batch['expected_status']
assert reports['batch']['reports'] == reports['singles']
assert [report['status'] for report in reports['singles']] == batch['expected_report_statuses']
assert [report['document'] for report in reports['singles']] == batch['request']['documents']
assert reports['concurrent_repeats'] == 2 and reports['invalid_requests'] == len(invalid)
assert head() == expected_head
clean()
for path, wanted in source_hashes.items():
    assert digest(root / path) == wanted, path
for path, wanted in input_hashes.items():
    assert digest(Path(path)) == wanted, path
output_path = base / 'spl-toolkit-root-m6-go-results.json'
output_path.write_text(json.dumps(reports, ensure_ascii=False, indent=2) + '\n')
evidence = {
    'status': 'passed', 'source_sha': expected_head,
    'scope': 'Independent canonical M6 rewrite and destination validation; adapters, installed and release acceptance separate',
    'cases': [case['name'] for case in cases], 'destination_validation_cases': [case['name'] for case in validation['cases']],
    'batch_documents': len(batch['request']['documents']), 'concurrent_repeats': 2,
    'strict_invalid_requests': len(invalid), 'late_invalid_batch_atomicity': 'passed',
    'caller_request_target_ownership': 'passed', 'canonical_original_candidate_validation_parity': 'passed', 'field_list_target_wire_projection': 'Only fields and optional_fields omitted; original kind, identity and version retained, including empty strings',
    'exact_audited_byte_reconstruction_and_coordinates': 'passed',
    'input_sha256': input_hashes, 'tracked_source_sha256': source_hashes,
    'results_sha256': digest(output_path), 'go_stderr': result.stderr.decode(errors='replace'),
}
evidence_path = base / 'spl-toolkit-root-m6-go-acceptance.json'
evidence_path.write_text(json.dumps(evidence, indent=2) + '\n')
print('Evidence:', evidence_path, flush=True)
