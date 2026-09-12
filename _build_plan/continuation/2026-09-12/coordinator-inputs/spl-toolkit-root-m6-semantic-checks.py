"""Independent M6 spec-derived assertions; reconcile wire names, not semantics, after preflight."""


def location(text, value):
    data = text.encode('utf-8')
    start, end = value['start'], value['end']
    assert 0 <= start['offset'] <= end['offset'] <= len(data), value
    for position in (start, end):
        prefix = data[:position['offset']].decode('utf-8').replace('\r\n', '\n').replace('\r', '\n')
        lines = prefix.split('\n')
        assert (position['line'], position['column']) == (len(lines), len(lines[-1]) + 1), position
    return data[start['offset']:end['offset']].decode('utf-8')


def analysis_locations(report):
    text = report['document']['text']
    def walk(value):
        if isinstance(value, dict):
            if set(value) == {'start', 'end'} and all(isinstance(value[key], dict) and 'offset' in value[key] for key in ('start', 'end')):
                location(text, value)
            for child in value.values():
                walk(child)
        elif isinstance(value, list):
            for child in value:
                walk(child)
    walk(report)
    ids = [ref['id'] for ref in report['references']]
    assert len(ids) == len(set(ids))


def exact_nested_count(value, expected):
    if value == expected:
        return 1
    if isinstance(value, dict):
        return sum(exact_nested_count(child, expected) for child in value.values())
    if isinstance(value, list):
        return sum(exact_nested_count(child, expected) for child in value)
    return 0


def check(case, report, original_analysis, candidate_analysis, candidate_validation=None):
    request = case['request']
    original = request['document']['text']
    assert report['schema_version'] == 1
    assert report['mode'] == request.get('mode', 'preview')
    assert report['status'] == case['expected_status'], (case['name'], report['status'])
    assert report['committed'] is case['expected_committed'], case['name']
    assert report['original_text'] == original
    assert report['candidate_text'] == case['expected_candidate_text'], case['name']
    assert report['text'] == case['expected_text'], case['name']
    assert report['document'] == original_analysis['document']
    assert report['original_analysis'] == original_analysis
    assert report['candidate_analysis'] == candidate_analysis
    assert candidate_analysis['document'] == original_analysis['document'] | {'text': report['candidate_text']}
    assert isinstance(report['changes'], list) and isinstance(report['rule_evaluations'], list)
    analysis_locations(original_analysis)
    analysis_locations(candidate_analysis)
    if candidate_validation is not None:
        if request.get('validation_target', {}).get('kind') == 'field_list':
            # Root Task1 I1 ruling: only rewrite JSON omits the two catalog arrays.
            # Keep the independently obtained canonical report complete and unchanged.
            target = candidate_validation['target']
            assert set(target) == {'kind', 'fields', 'optional_fields', 'identity', 'version'}
            expected = candidate_validation | {'target': {
                key: target[key] for key in ('kind', 'identity', 'version')
            }}
            assert report['candidate_validation'] == {'kind': 'field_list', 'field_list': expected}, 'Field-list validation differs beyond the approved target projection'
        else:
            assert exact_nested_count(report['candidate_validation'], candidate_validation) == 1, 'Canonical schema validation report was flattened, modified or duplicated'
    else:
        assert not report.get('candidate_validation')
    physical = []
    located = []
    for change in report['changes']:
        assert change['outcome'] in {'applied', 'skipped', 'ambiguous'}
        assert isinstance(change['candidate_applied'], bool) and isinstance(change['committed'], bool)
        if change.get('original_location') is not None:
            assert location(original, change['original_location']) == change['old_text']
            located.append(change['original_location']['start']['offset'])
        if change['outcome'] == 'ambiguous':
            assert not change['candidate_applied'] and not change['committed']
        if change['candidate_applied']:
            assert change.get('original_location') and change.get('candidate_location')
            assert change['old_text'] != change['new_text'], 'No-op was fabricated as an applied edit'
            assert location(report['candidate_text'], change['candidate_location']) == change['new_text']
            physical.append(change)
        else:
            assert not change['committed']
        if not report['committed']:
            assert not change['committed']
    assert located == sorted(located), 'Audit is not in original source order'
    physical.sort(key=lambda change: change['original_location']['start']['offset'])
    original_bytes = original.encode('utf-8')
    output = bytearray()
    previous = 0
    seen = set()
    for change in physical:
        start = change['original_location']['start']['offset']
        end = change['original_location']['end']['offset']
        key = (start, end, change['new_text'])
        assert key not in seen, 'Identical physical edit was not coalesced'
        seen.add(key)
        assert previous <= start < end, 'Overlapping or invented empty edit interval'
        output.extend(original_bytes[previous:start])
        candidate_start = len(output)
        output.extend(change['new_text'].encode('utf-8'))
        assert change['candidate_location']['start']['offset'] == candidate_start
        assert change['candidate_location']['end']['offset'] == len(output)
        previous = end
    output.extend(original_bytes[previous:])
    assert bytes(output).decode('utf-8') == report['candidate_text'], 'Candidate changed bytes outside its audit intervals'
    if report['committed']:
        assert request['mode'] == 'apply' and physical and all(change['committed'] for change in physical)
        assert report['text'] == report['candidate_text']
    else:
        assert report['text'] == original
    if request['mode'] == 'preview':
        assert all(change['outcome'] == 'applied' for change in physical)
    if request['mode'] == 'apply' and physical and not report['committed']:
        assert all(change['outcome'] == 'skipped' and change['reason'] == 'post_verification_failed' for change in physical)
    name = case['name']
    if name in {'source_alias_unicode_crlf', 'preview_keeps_original', 'rename_keeps_alias'}:
        assert len(physical) == 2
        assert all(change['old_text'] == 'src' for change in physical)
        assert candidate_analysis['document']['text'].endswith('table alias')
    if name == 'dependency_kind_not_spelling':
        assert len(physical) == 1 and physical[0]['old_text'] == 'auth' and physical[0]['new_text'] == 'audit'
    if name == 'conflicting_targets':
        assert any(change['outcome'] == 'ambiguous' and change['reason'] == 'conflicting_targets' for change in report['changes'])
    if name == 'target_captures_derived_alias':
        assert any(change['reason'] == 'binding_collision' for change in report['changes'])
    if name == 'identical_targets_coalesce':
        assert len(physical) == 2
        assert all(set(change['rule_ids']) == {'same-first', 'same-second'} for change in physical)
    if name == 'simultaneous_swap':
        assert len(physical) == 4
    if name == 'malformed_original':
        assert not physical and original_analysis['status'] == 'invalid'
    if name == 'noop_still_validates_destination':
        assert not physical and candidate_validation['status'] == 'invalid'

    if name in {'spl2_sql_alias_unicode_crlf', 'spl2_sql_phase_condition'}:
        assert len(physical) == 1 and physical[0]['old_text'] == 'src' and physical[0]['new_text'] == 'actor'
        assert candidate_analysis['document']['text'].endswith('table alias')
        assert [e['phase'] for e in candidate_analysis['lineage'] if 'phase' in e] == ['source', 'filter', 'evaluate', 'project']
    if name == 'spl2_dataset_identity':
        assert len(physical) == 1 and physical[0]['old_text'] == 'main' and physical[0]['new_text'] == 'archive'
    if name == 'spl2_quoted_dotted_atom':
        assert len(physical) == 1
        assert (physical[0]['old_text'], physical[0]['new_text']) in {
            ("'actor.name'", "'actor_name'"), ('actor.name', 'actor_name')}

    if name == 'spl_implicit_aggregate_link':
        assert len(physical) == 2
        assert len({change['group_id'] for change in physical}) == 1, 'Implicit source and consumer must be one atomic group'
        consumers = [ref for ref in candidate_analysis['references']
                     if ref['kind'] == 'field' and ref['original_name'] == "'sum(octets)'" ]
        assert len(consumers) == 1 and consumers[0]['binding'] == 'derived'
        assert consumers[0]['origin_reference_ids']
    if name == 'spl2_unproved_implicit_link_refused':
        assert not physical
        assert any(change.get('original_location') and change.get('reason') in
                   {'linked_edit_unproven', 'unsupported_reference', 'target_not_renderable'}
                   for change in report['changes']), 'Unproved linked group needs a located refusal'

    if name in {'spl_quoted_pattern_no_literal_authority', 'spl2_quoted_pattern_no_literal_authority'}:
        assert not physical
        decisions = report['changes'] + report['rule_evaluations']
        assert any(item['reason'] == 'condition_unknown' for item in decisions), 'A wildcard pattern fabricated exact literal authority'
    if name in {'spl_exact_string_after_pattern', 'spl2_exact_string_after_pattern'}:
        assert len(physical) == 1
        assert (physical[0]['old_text'], physical[0]['new_text']) == ('people', 'accounts')
