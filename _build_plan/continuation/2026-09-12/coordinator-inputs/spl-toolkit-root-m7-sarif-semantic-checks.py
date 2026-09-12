"""Prepared independent SARIF checks; no M7 product output has been tested.

Call only after the official schema validator passes. Expected findings come
from canonical diagnostics, and expected artifact URIs come from the accepted
producer contract and original corpus inputs, never from the SARIF result.
"""
from collections import Counter
from urllib.parse import urljoin


def _encoded_uri(uri):
    assert isinstance(uri, str) and uri
    assert all(33 <= ord(char) < 127 for char in uri), ('unencoded URI', uri)
    return uri


def _artifact_uri(run, location, seen=()):
    """Resolve declared artifact/base indices and check redundant URI identity."""
    indexed = None
    index = location.get('index', -1)
    if index != -1:
        artifacts = run.get('artifacts', [])
        assert type(index) is int and 0 <= index < len(artifacts)
        token = ('artifact', index)
        assert token not in seen, ('cyclic artifact index', index)
        indexed = _artifact_uri(run, artifacts[index]['location'], seen + (token,))
    if 'uri' not in location:
        assert indexed is not None
        return indexed
    uri = _encoded_uri(location['uri'])
    if 'uriBaseId' in location:
        key = location['uriBaseId']
        bases = run.get('originalUriBaseIds', {})
        token = ('base', key)
        assert key in bases and token not in seen, ('unknown or cyclic URI base', key)
        base = _artifact_uri(run, bases[key], seen + (token,))
        uri = urljoin(base, uri)
    assert indexed is None or indexed == uri, ('artifact index/URI mismatch', indexed, uri)
    return uri


def _region_tuple(location):
    return (location['start']['line'], location['start']['column'],
            location['end']['line'], location['end']['column'])


def check_sarif(report, expected_findings, documents_by_uri, *,
                virtual_uris=(), execution_complete=True):
    """Compare full diagnostic multiset, encoded sources, and invocation state.

    expected_findings entries: {uri, code, severity, message, location}; location
    is the canonical UTF-8 half-open location, or None for separately expected
    completeness/file-level evidence. Omit uri for a fully unlocated result.
    The selected producer encoding supplies explicit result.level; this oracle
    does not implement optional SARIF rule-configuration severity resolution.
    documents_by_uri maps independently
    derived final URIs to original text. Virtual artifacts must embed exact text.
    URI spelling and virtual-file naming must be reconciled before first use.
    """
    assert report['version'] == '2.1.0'
    assert 'schema_version' not in report
    assert len(report['runs']) == 1
    run = report['runs'][0]
    assert run['columnKind'] == 'unicodeCodePoints'
    rules = run['tool']['driver'].get('rules', [])
    rule_ids = [rule['id'] for rule in rules]
    assert len(rule_ids) == len(set(rule_ids))
    assert run['invocations']
    assert all(invocation['executionSuccessful'] is execution_complete
               for invocation in run['invocations'])
    artifact_text = {}
    for artifact in run.get('artifacts', []):
        uri = _artifact_uri(run, artifact['location'])
        if 'contents' in artifact:
            assert uri in documents_by_uri
            assert artifact['contents']['text'] == documents_by_uri[uri]
            artifact_text[uri] = artifact['contents']['text']
    assert set(virtual_uris) <= artifact_text.keys()
    severity = {'error': 'error', 'warning': 'warning', 'info': 'note'}
    expected = Counter()
    for finding in expected_findings:
        uri = _encoded_uri(finding['uri']) if finding.get('uri') is not None else None
        loc = finding.get('location')
        coordinates = None
        if loc is not None:
            assert uri in documents_by_uri
            data = documents_by_uri[uri].encode('utf-8')
            assert 0 <= loc['start']['offset'] <= loc['end']['offset'] <= len(data)
            for position in (loc['start'], loc['end']):
                prefix = data[:position['offset']].decode('utf-8')
                lines = prefix.replace('\r\n', '\n').replace('\r', '\n').split('\n')
                assert (position['line'], position['column']) == (len(lines), len(lines[-1]) + 1)
            coordinates = _region_tuple(loc)
        expected[(uri, finding['code'], severity[finding['severity']],
                  finding['message'], coordinates)] += 1
    actual = Counter()
    for result in run['results']:
        code = result['ruleId']
        assert code in rule_ids, 'Producer result is absent from its deterministic rule table'
        if result.get('ruleIndex', -1) != -1:
            index = result['ruleIndex']
            assert type(index) is int and 0 <= index < len(rules)
            assert rules[index]['id'] == code
        locations = result.get('locations', [])
        assert len(locations) <= 1
        uri = None
        coordinates = None
        if locations:
            physical = locations[0]['physicalLocation']
            uri = _artifact_uri(run, physical['artifactLocation'])
            if 'region' in physical:
                assert uri in documents_by_uri
                region = physical['region']
                lines = documents_by_uri[uri].replace('\r\n', '\n').replace('\r', '\n').split('\n')
                start_line = region['startLine']
                end_line = region.get('endLine', start_line)
                assert 1 <= start_line <= end_line <= len(lines)
                start_column = region.get('startColumn', 1)
                end_column = region.get('endColumn', len(lines[end_line - 1]) + 1)
                assert 1 <= start_column <= len(lines[start_line - 1]) + 1
                assert 1 <= end_column <= len(lines[end_line - 1]) + 1
                assert start_line < end_line or start_column <= end_column
                coordinates = (start_line, start_column, end_line, end_column)
        assert result['level'] in {'error', 'warning', 'note'}, 'Selected producer must encode severity explicitly'
        actual[(uri, code, result['level'], result['message']['text'], coordinates)] += 1
    assert actual == expected, {'missing': list((expected - actual).elements()),
                                'extra': list((actual - expected).elements())}
