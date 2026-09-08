"""SPL2 adapter parity and strict owned-result boundaries through real native code."""

from concurrent.futures import ThreadPoolExecutor
from copy import deepcopy
import json
import os
from pathlib import Path
import threading
import time

import pytest

from spl_toolkit import SPLMapper, SPLMapperError
from spl_toolkit.exceptions import MapperNotFoundError

FIXTURES = Path(os.environ['SPL_SPL2_FIXTURES'])
assert FIXTURES.is_absolute(), 'SPL_SPL2_FIXTURES must be absolute'
CASES = json.loads((FIXTURES / 'canonical-core.json').read_text(encoding='utf-8'))
SNAPSHOT = 'spl2-provenance-v1:sha256:3345cf5712b1bdbf467d1651784fdb8bccc596805038da0d54e7a123384e3a4e'
TARGET = {'kind': 'json_schema', 'schema': {'type': 'object', 'properties': {'host': True, 'value': True},
          'required': ['host', 'value'], 'additionalProperties': False}}
CATALOG = {'fields': ['host', 'value']}


def mapper_kwargs():
    return {'library_path': os.environ['SPL_NATIVE_LIBRARY']} if 'SPL_NATIVE_LIBRARY' in os.environ else {}


def native_json(mapper, operation, payload, *, reject=False, handle=None):
    native = getattr(mapper._lib, 'spl_mapper_' + operation)
    pointer = native(mapper._mapper_id if handle is None else handle, payload)
    try:
        assert pointer
        if reject:
            assert pointer.contents.error and pointer.contents.result is None
            return pointer.contents.error.decode('utf-8')
        assert not pointer.contents.error, pointer.contents.error
        return json.loads(pointer.contents.result)
    finally:
        mapper._lib.spl_result_free(pointer)


def test_capability_selectors_preserve_defaults_and_snapshot():
    with SPLMapper(**mapper_kwargs()) as mapper:
        original = mapper.capabilities()
        assert original == mapper.capabilities(language='', profile='', version='')
        assert original == native_json(mapper, 'capabilities_for', b'{}')
        pointer = mapper._lib.spl_mapper_capabilities(mapper._mapper_id)
        try:
            assert json.loads(pointer.contents.result) == original
            assert pointer.contents.result == json.dumps(original, separators=(',', ':'), ensure_ascii=False).encode()
        finally:
            mapper._lib.spl_result_free(pointer)
        assert 'documentation_snapshot' not in original
        manifest = mapper.capabilities(language='spl2')
        assert manifest['documentation_snapshot'] == SNAPSHOT
        assert manifest['language'] == 'spl2'
        assert any(c['name'] == 'from' and c['semantic_supported'] for c in manifest['commands'])
        assert manifest == native_json(mapper, 'capabilities_for', b'{"language":"spl2"}')
        mapper.capabilities(language='spl2')['commands'].clear()
        assert mapper.capabilities(language='spl2') == manifest


@pytest.mark.parametrize('payload', [None, b'null', b'[]', b'{', b'{} {}', b'{"text":"x"}',
    b'{"language":null}', b'{"profile":false}', b'{"version":42}', b'{"language":"spl","language":"spl2"}',
    b'{"Language":"spl2"}', b'{"ver\\u017fion":"current"}', b'{"language":"SPL2"}',
    b'{"language":"spl2","profile":"cloud"}', b'{"version":"next"}', b'{"version":NaN}',
    b'{"language":"\\ud800"}', b'{"language":"\xff"}'])
def test_native_capabilities_strict_owned_errors(payload):
    with SPLMapper(**mapper_kwargs()) as mapper:
        native_json(mapper, 'capabilities_for', payload, reject=True)
        assert mapper.capabilities(language='spl2')['language'] == 'spl2'


@pytest.mark.parametrize('payload', [b'null', b'{}', b'{"text":null}', b'{"text":"x","extra":1}',
    b'{"text":"x","text":"y"}', b'{"Text":"x"}', b'{"text":"x","language":null}',
    b'{"text":"x"},{"text":"y"}', b'{"text":NaN}', b'{"text":"x","ver\\u017fion":"current"}'])
def test_native_analysis_rejects_malformed_document_wrappers(payload):
    with SPLMapper(**mapper_kwargs()) as mapper:
        native_json(mapper, 'analyze_query', payload, reject=True)


@pytest.mark.parametrize('operation,args', [('map_query', ('search host=web',)),
    ('map_query_with_context', ('search host=web', {})), ('discover_query', ('search host=web',)),
    ('get_input_fields', ('search host=web',))])
def test_legacy_selectors_reject_spl2_before_legacy_abi(operation, args, monkeypatch):
    with SPLMapper(**mapper_kwargs()) as mapper:
        original = getattr(mapper, operation)(*args)
        assert getattr(mapper, operation)(*args, language='', profile='', version='') == original
        native_name = 'spl_mapper_' + ('discover_query' if operation == 'get_input_fields' else operation)
        def forbidden(*_args):
            pytest.fail('legacy ABI was invoked with unsupported selectors')
        monkeypatch.setattr(mapper._lib, native_name, forbidden)
        for options in ({'language': 'spl2'}, {'language': 'SPL2'}, {'profile': 'cloud'}, {'version': None}):
            with pytest.raises(SPLMapperError):
                getattr(mapper, operation)(*args, **options)
        with pytest.raises(SPLMapperError, match='analyze_query.*validate_fields.*validate_schema'):
            getattr(mapper, operation)(*args, language='spl2')


@pytest.mark.parametrize('operation,args', [('analyze_query', ('FROM main',)), ('capabilities', ())])
@pytest.mark.parametrize('value', [float('nan'), float('inf'), object()])
def test_python_invalid_json_rejected_before_native_allocation(operation, args, value, monkeypatch):
    with SPLMapper(**mapper_kwargs()) as mapper:
        native_name = 'spl_mapper_' + ('capabilities_for' if operation == 'capabilities' else operation)
        def forbidden(*_args):
            pytest.fail('non-JSON input reached native code')
        monkeypatch.setattr(mapper._lib, native_name, forbidden)
        with pytest.raises(SPLMapperError):
            getattr(mapper, operation)(*args, language=value)
        assert mapper._active_calls == 0


@pytest.mark.parametrize('case', CASES, ids=lambda c: c['id'])
def test_native_spl2_canonical_corpus_and_full_c_python_parity(case):
    document = deepcopy(case['document'])
    document['source_id'] = 'opaque:é😀\x00.spl2'
    with SPLMapper(**mapper_kwargs()) as mapper:
        report = mapper.analyze_query(document['text'], **{k:v for k,v in document.items() if k != 'text'})
        assert report == native_json(mapper, 'analyze_query', json.dumps(document).encode())
    expected = case['canonical']  # Separate accepted semantic expectations, not syntax dispositions.
    assert report['status'] == expected['status']
    assert report['coverage']['syntax_complete'] == expected['syntax_complete']
    assert report['coverage']['semantic_complete'] == expected['semantic_complete']
    assert [d['code'] for d in report['diagnostics']] == expected['expected_codes']
    assert [s['command'] for s in report['stages']] == expected['stage_commands']
    assert report['document'] == document | {'profile':'splunkd', 'version':'current'}
    assert [{k:r[k] for k in ('original_name', 'normalized_name', 'kind', 'role', 'binding')} |
            {'start':r['location']['start']['offset'], 'end':r['location']['end']['offset']}
            for r in report['references']] == expected['references']
    for ref in report['references']:
        location = ref['location']
        assert document['text'].encode()[location['start']['offset']:location['end']['offset']].decode() == ref['original_name']


@pytest.mark.parametrize('kind,target', [('fields', CATALOG), ('schema', TARGET)])
def test_mixed_single_batch_full_native_parity_and_conditional_bindings(kind, target):
    documents = [
        {'text':'table host', 'source_id':'spl'},
        {'text':'FROM main | eval x=tonumber(value) | table x', 'language':'spl2', 'source_id':'é😀\x00'},
        {'text':'SELECT host AS h FROM main WHERE isnotnull(host) ORDER BY h', 'language':'spl2', 'source_id':'sql'},
        {'text':'FROM main | table missing', 'language':'spl2', 'source_id':'invalid'},
    ]
    original = deepcopy((documents, target))
    operation = 'validate_' + kind
    key = 'catalog' if kind == 'fields' else 'target'
    with SPLMapper(**mapper_kwargs()) as mapper:
        reports = []
        for doc in documents:
            report = getattr(mapper, operation)(doc['text'], target, **{k:v for k,v in doc.items() if k != 'text'})
            assert report == native_json(mapper, operation, json.dumps({'document':doc, key:target}).encode())
            assert report['analysis'] == mapper.analyze_query(doc['text'], **{k:v for k,v in doc.items() if k != 'text'})
            reports.append(report)
        batch = getattr(mapper, operation + '_batch')(documents, target)
        assert batch == native_json(mapper, operation + '_batch', json.dumps({'documents':documents, key:target}).encode())
    assert batch['reports'] == reports
    assert (documents, target) == original
    assert reports[1]['analysis']['coverage']['semantic_complete'] is True
    x = [r for r in reports[1]['analysis']['references'] if r['normalized_name']=='x' and r['role']=='read']
    assert x and x[0]['binding'] == 'indeterminate'
    sql = reports[2]['analysis']
    assert any(r['role'] == 'null_test' for r in sql['references'])
    assert all(o['reference_id'] not in {r['id'] for r in sql['references'] if r['role']=='null_test'} for o in reports[2]['outcomes'])
    assert any('execution_order' in stage and 'phase' in stage for stage in sql['lineage'])
    assert all('phase' not in s and 'execution_order' not in s for s in reports[0]['analysis']['lineage'])


@pytest.mark.parametrize('operation', ['capabilities_for', 'analyze_query', 'validate_fields', 'validate_fields_batch', 'validate_schema', 'validate_schema_batch'])
def test_spl2_owned_errors_and_closed_handles(operation):
    mapper = SPLMapper(**mapper_kwargs())
    handle = mapper._mapper_id
    mapper.close()
    for bad in (handle, -1, 2147483647):
        assert native_json(mapper, operation, b'{}', reject=True, handle=bad) == 'Mapper not found'
    with pytest.raises(MapperNotFoundError):
        mapper.capabilities(language='spl2')


def test_spl2_capability_results_free_on_success_error_and_decode_error(monkeypatch):
    with SPLMapper(**mapper_kwargs()) as mapper:
        freed = []
        real_free = mapper._lib.spl_result_free
        def free(pointer):
            freed.append(bool(pointer))
            real_free(pointer)
        monkeypatch.setattr(mapper._lib, 'spl_result_free', free)
        mapper.capabilities(language='spl2')
        with pytest.raises(SPLMapperError):
            mapper.capabilities(language='unsupported')
        def fail(_value):
            raise ValueError('decode failed')
        monkeypatch.setattr('spl_toolkit.mapper.json.loads', fail)
        with pytest.raises(ValueError, match='decode failed'):
            mapper.capabilities(language='spl2')
        assert freed == [True, True, True]
        assert mapper._active_calls == 0


def test_spl2_capability_close_waits_for_admission(monkeypatch):
    mapper = SPLMapper(**mapper_kwargs())
    admitted, proceed = threading.Event(), threading.Event()
    real = mapper._lib.spl_mapper_capabilities_for
    def paused(*args):
        admitted.set()
        assert proceed.wait(5)
        return real(*args)
    monkeypatch.setattr(mapper._lib, 'spl_mapper_capabilities_for', paused)
    with ThreadPoolExecutor(max_workers=3) as pool:
        running = pool.submit(mapper.capabilities, language='spl2')
        assert admitted.wait(5)
        closing = pool.submit(mapper.close)
        try:
            deadline = time.monotonic()+5
            while not mapper._closing:
                assert time.monotonic() < deadline
                time.sleep(.001)
            assert not closing.done()
            with pytest.raises(MapperNotFoundError):
                mapper.capabilities(language='spl2')
        finally:
            proceed.set()
        assert running.result(5)['language'] == 'spl2'
        closing.result(5)


@pytest.mark.parametrize('kind,target', [('fields', CATALOG), ('schema', TARGET)])
@pytest.mark.parametrize('batch', [False, True])
@pytest.mark.parametrize('document', [b'null', b'{"language":"spl2"}', b'{"text":null,"language":"spl2"}',
    b'{"text":"FROM main","language":"spl2","extra":1}',
    b'{"text":"FROM main","language":"spl2","language":"spl"}',
    b'{"text":"FROM main","language":"spl2","source_id":"\\ud800"}',
    b'{"text":"FROM main","language":"spl2","source_id":"\xff"}',
    b'{"text":"FROM main","language":"spl2","version":NaN}'])
def test_spl2_validation_native_strict_owned_errors(kind, target, batch, document):
    key = 'catalog' if kind == 'fields' else 'target'
    payload = (b'{"documents":[' + document + b']' if batch else b'{"document":' + document)
    payload += b',"' + key.encode() + b'":' + json.dumps(target).encode() + b'}'
    with SPLMapper(**mapper_kwargs()) as mapper:
        native_json(mapper, 'validate_' + kind + ('_batch' if batch else ''), payload, reject=True)
        assert getattr(mapper, 'validate_' + kind)('FROM main | table host', target, language='spl2')['status'] == 'valid'


@pytest.mark.parametrize('kind,target', [('fields', CATALOG), ('schema', TARGET)])
@pytest.mark.parametrize('batch', [False, True])
@pytest.mark.parametrize('value', [float('nan'), float('inf'), object()])
def test_spl2_validation_python_bad_json_never_allocates(kind, target, batch, value, monkeypatch):
    operation = 'validate_' + kind + ('_batch' if batch else '')
    with SPLMapper(**mapper_kwargs()) as mapper:
        def forbidden(*_args):
            pytest.fail('non-JSON input reached native code')
        monkeypatch.setattr(mapper._lib, 'spl_mapper_' + operation, forbidden)
        with pytest.raises(SPLMapperError):
            if batch:
                getattr(mapper, operation)([{'text':'FROM main', 'language':'spl2', 'source_id':value}], target)
            else:
                getattr(mapper, operation)('FROM main', target, language='spl2', source_id=value)
        assert mapper._active_calls == 0


@pytest.mark.parametrize('query,schema,status,outcomes', [
    ("FROM main | where 'actor.name'=\"a\"", {'type':'object', 'properties':{'actor':{'type':'object', 'properties':{'name':True}, 'required':['name'], 'additionalProperties':False}}, 'required':['actor'], 'additionalProperties':False}, 'incomplete', ['indeterminate']),
    ('FROM main | where actor.name="a"', {'type':'object', 'properties':{'actor.name':True}, 'required':['actor.name'], 'additionalProperties':False}, 'invalid', ['missing', 'indeterminate']),
    ('SELECT isnull(missing) AS ok FROM main', False, 'valid', []),
])
def test_spl2_schema_identity_and_null_inspection_boundaries(query, schema, status, outcomes):
    target = {'kind':'json_schema', 'schema':schema}
    with SPLMapper(**mapper_kwargs()) as mapper:
        report = mapper.validate_schema(query, target, language='spl2')
        assert report == native_json(mapper, 'validate_schema', json.dumps({'document':{'text':query, 'language':'spl2'}, 'target':target}).encode())
    assert report['status'] == status
    assert [item['outcome'] for item in report['outcomes']] == outcomes
    for item in report['outcomes']:
        if item['outcome'] == 'indeterminate':
            assert item['matches_complete'] is False and item['matches'] == []


def test_spl2_shared_native_calls_are_deterministic():
    with SPLMapper(**mapper_kwargs()) as mapper:
        def work(_):
            return (mapper.capabilities(language='spl2'), mapper.analyze_query('SELECT host FROM main', language='spl2'),
                    mapper.validate_schema('FROM main | table host', TARGET, language='spl2'))
        expected = work(None)
        with ThreadPoolExecutor(max_workers=8) as pool:
            assert all(result == expected for result in pool.map(work, range(64)))
