"""Independent future M7 filesystem fixtures; no product invocation or API claims.

Import materialize(parent) only during an authorized M7 acceptance window. Each
call creates a fresh private directory. Runtime races require separate loader
controls; these static layouts do not attest race resistance.
"""
from __future__ import annotations

import hashlib
import json
import os
import tempfile
from pathlib import Path, PurePosixPath

CASES = Path('/private/tmp/spl-toolkit-root-m7-acquisition-cases.json')
CASES_SHA256 = '8c63c74b92a8d118c02fdd40bdfbd479f8cc1f136aaa04857e795f6ece61668c'


def _file(root: Path, name: str, contents: bytes) -> Path:
    relative = PurePosixPath(name)
    if relative.is_absolute() or '..' in relative.parts or '\\' in name:
        raise ValueError('Fixture path must be relative and contained')
    path = root.joinpath(*relative.parts)
    path.parent.mkdir(parents=True, exist_ok=True)
    with path.open('xb') as stream:
        stream.write(contents)
    return path


def preservation(paths: dict[str, Path]) -> dict:
    """Snapshot explicit controls without traversing directories or symlinks."""
    result = {}
    for label, path in paths.items():
        if path.is_symlink():
            result[label] = {'path': str(path), 'kind': 'symlink', 'target': os.readlink(path)}
        elif not path.exists():
            result[label] = {'path': str(path), 'kind': 'absent'}
        elif path.is_file():
            result[label] = {'path': str(path), 'kind': 'file',
                             'sha256': hashlib.sha256(path.read_bytes()).hexdigest()}
        else:
            result[label] = {'path': str(path), 'kind': 'nonregular'}
    return result


def materialize(parent: Path) -> dict:
    raw = CASES.read_bytes()
    if hashlib.sha256(raw).hexdigest() != CASES_SHA256:
        raise ValueError('Independent fixture inputs changed; reconcile before use')
    data = json.loads(raw)
    base = Path(tempfile.mkdtemp(prefix='root-m7-acquisition-', dir=parent))
    directory = base / 'directory'
    manifest = base / 'manifest'
    controls = {}
    for root in (directory, manifest):
        for item in data['files']:
            path = _file(root, item['path'], item['text'].encode('utf-8'))
            controls[f'{root.name}/{item["path"]}'] = path
    for item in data['manifest_cases']:
        if 'bytes_hex' in item:
            path = _file(manifest, item['file'], bytes.fromhex(item['bytes_hex']))
            controls[f'manifest/{item["file"]}'] = path
        elif item['expected'] == 'acquisition_failure':
            controls[f'manifest/{item["file"]}'] = manifest / item['file']

    # All outside sentinels are still inside this private fixture directory.
    outside = base / 'outside'
    outside_file = _file(outside, 'outside.spl', b'search outside_sentinel=do_not_acquire')
    controls['outside-sentinel'] = outside_file
    guarded = base / 'guarded'
    inside = _file(guarded, 'inside.spl', b'search host=inside')
    controls['guarded-query'] = inside
    (guarded / 'file-link.spl').symlink_to(outside_file)
    (guarded / 'directory-link').symlink_to(outside, target_is_directory=True)
    (guarded / 'selected-directory.spl').mkdir()
    fifo = guarded / 'selected-fifo.spl'
    fifo_supported = hasattr(os, 'mkfifo')
    if fifo_supported:
        os.mkfifo(fifo, 0o600)
    (base / 'root-final-link').symlink_to(guarded, target_is_directory=True)
    ancestor = base / 'ancestor'
    ancestor_root = ancestor / 'real-root'
    _file(ancestor_root, 'query.spl', b'search host=ancestor')
    (base / 'ancestor-link').symlink_to(ancestor, target_is_directory=True)

    aliases = base / 'output-aliases'
    aliases.mkdir()
    os.link(inside, aliases / 'hardlink.spl')
    (aliases / 'leaf-link.spl').symlink_to(inside)
    (aliases / 'parent-link').symlink_to(guarded, target_is_directory=True)
    missing = guarded / 'missing-selected-input.spl'
    controls['missing-selected-input'] = missing
    for name in ('file-link.spl', 'directory-link'):
        controls[f'guarded/{name}'] = guarded / name
    controls['output-hardlink'] = aliases / 'hardlink.spl'
    controls['output-leaf-link'] = aliases / 'leaf-link.spl'
    controls['output-parent-link'] = aliases / 'parent-link'
    result = {
        'status': 'fixtures-materialized-product-unexecuted',
        'base': str(base),
        'input_cases_sha256': CASES_SHA256,
        'directory_root': str(directory),
        'manifest_root': str(manifest),
        'guarded_root': str(guarded),
        'root_final_symlink': str(base / 'root-final-link'),
        'root_with_ancestor_alias': str(base / 'ancestor-link' / 'real-root'),
        'output_aliases': str(aliases),
        'missing_selected_input': str(missing),
        'fifo_supported': fifo_supported,
        'preservation_before': preservation(controls),
        'control_paths': {name: str(path) for name, path in controls.items()},
        'limitations': [
            'No product operation, result, CLI exit, or platform acceptance is implied',
            'Manifest serialization awaits the actual accepted M7 wire contract',
            'Rules, target and resources collision inputs are supplied by the actual surface harness',
            'Static layouts do not prove race resistance or held-descriptor semantics',
            'The caller preserves the fixture directory and raw execution evidence',
        ],
    }
    (base / 'fixture-receipt.json').write_text(json.dumps(result, ensure_ascii=False, indent=2) + '\n')
    return result
