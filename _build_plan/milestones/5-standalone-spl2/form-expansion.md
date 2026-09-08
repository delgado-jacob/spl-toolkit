# Standalone SPL2 form and function expansion

Status: bounded design/research supplement, retrieved 2026-09-07 EDT (latest retrievals 2026-09-08 UTC). No M5 parser or fixture execution is represented. The final shared lowering and SQL phase contract still awaits verified M4 and controller release. This is not an implementation plan.

The approved [conformance matrix](conformance-matrix.md) retains its 53 inventory rows, 35 mandatory command families, 25 language families, 68 cards, 288 seed identities, and H01–H12. Its pre-expansion SHA256 was `3e402c278bdca70a152eaf0b0ea8b97ed2c3dc9b4bbde04958957a541199fc63`. This supplement records **149 form/option obligation rows and 34 approved positional function spellings**. These are cross-referenced design obligations, not additional unique-fixture totals. Several deliberately share candidate queries and negative bounds. Do not add their counts to 288 or claim tests pass.

## Reading the obligations

All rows concern standalone Search & Reporting in **splunkd**. The original command compatibility inventory governs membership; function membership is corroborated by [evalcompat](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/spl2-compatibility-profiles-and-quick-references/compatibility-quick-reference-for-spl2-evaluation-functions) and [aggcompat](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/spl2-compatibility-profiles-and-quick-references/compatibility-quick-reference-for-spl2-statistical-functions). Shared language pages provide syntax, not permission to import modules, Edge/Ingest behavior or custom function declarations. Sources and publisher metadata are recorded below.

**A** means intended recognition of the evidenced grammar. **K** means the approved complete field-model subset, conditional on supported operands, sound source bindings, and verified shared-kernel proofs. **U** means deliberately incomplete semantics or additional unmodeled syntax. Neither A nor K claims current production implementation. An A/U positive must preserve honest incomplete status. EH entries are separate held research variants; they never count as definite negatives or complete support.

P/N labels identify future obligations. Each form supplies two positive and two minimal-negative candidates; where an operand is repeatable or optional, later fixture expansion must assert every listed alternative at its distinct boundary. N means malformed syntax **or a located documented contract violation**, as in the base matrix. Wrong arity, duplicate object keys, documented rename restrictions and explicit numeric option bounds can parse successfully and still be invalid; do not force syntax coverage false for semantic violations. Runtime field-value computation and general type checking remain excluded.

All E.L01.start candidates are complete queries and must never receive a prefix, including deliberately invalid starts: E.L01.start.N1 remains exactly `failure index=app`. For other form snippets, prepend `FROM main | ` unless already beginning FROM, SELECT, search, index=, union, makeresults, loadjob, tstats, mstats, spl1 or a command-position backtick. This rule also applies to other language-tail snippets here, unlike the base matrix's complete language queries. Scalar F expressions use `FROM main | eval result=<expression>`; aggregate F snippets use `FROM main | <stats tail>`. Preserve every source character when assembling durable fixtures. Boolean expressions are legal eval values under the prior controller ruling.

Syntax notation is compact prose, not a replacement EBNF. Each alternative, omission, repeated-list separator and source-specific option position carries its own typed-tree assertion. Existing IDs in “Base” refer to the immutable card or seed; E/F labels here are obligation keys, not replacements. Do not count case-only or data-only permutations as new meaningful coverage.

## Explicit controller decisions

Named built-in calls use **colon labels**. Admit all-named reordering, a positional prefix followed by named arguments, and array-wrapped named lists in typed grammar. Their field semantics remain **U throughout M5**, even when a particular call seems easy to model. Do not infer a generic parameter registry or automatically promote named forms. Labels are non-field syntax; their value expressions and arrays retain located reads. The detailed ordering section and valid/invalid table in [namingargs](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-manual/functions/naming-function-arguments) resolve its introductory typo; that typo is not a held ambiguity.

`round(precision:2,num:bytes)`, `round(bytes,precision:2)`, and `coalesce(values:[primary,backup])` are recognized/incomplete examples. `round(num:bytes,2)` and `coalesce(values:primary,backup)` violate the documented ordering/list grammar. Custom default arguments do not establish equals-sign built-in labels. **A call containing = can still contain a legitimate positional comparison expression**; for example `if(code=200,1,0)` is an ordinary positional call. Do not guess named intent or introduce type checking to reject such expressions. Unproved duplicate/unknown named labels, named case-pair encoding, and named aggregate forms stay U.

Predicate type tests are `expr IS type` or `NOT (expr IS type)`; the independently documented `expr IS NOT NULL` does not establish `expr IS NOT type`. The L07 compact summary is corrected without changing any approved seeds. The extra IS NOT type layout is EH04/U, not a manufactured invalid case.

## Coverage crosswalk

Every mandatory family remains represented; absence of an additional option row does not create an unreviewed option promise.

| Command family | Base | Expansion |
| --- | --- | --- |
| search | C01 | E.C01.* |
| from/select | C02 | E.L15.*, E.L16.*, E.L17.*, E.L18.*, E.L19.*, base Q01–Q08 |
| eval | C03 | E.C03.* |
| where | C04 | E.C04.* |
| fields | C05 | E.C05.* |
| table | C06 | E.C06.* |
| rename | C07 | E.C07.* |
| stats | C08 | E.C08.* |
| eventstats | C09 | E.C09.* |
| streamstats | C10 | E.C10.* |
| lookup | C11 | E.C11.* |
| sort | C12 | E.C12.* |
| dedup | C13 | E.C13.* |
| head | C14 | E.C14.* |
| reverse | C15 | No operands/options; existing C15 positive/negative pairs suffice |
| join | C16 | E.C16.* |
| append | C17 | E.C17.* |
| appendpipe | C18 | E.C18.* |
| appendcols | C19 | E.C19.* |
| union | C20 | E.C20.* |
| if | C21 | E.C21.* |
| spl1 | C22 | E.C22.* |
| bin | C23 | E.C23.* |
| rex | C24 | E.C24.* |
| spath | C25 | E.C25.* |
| makeresults | C26 | E.C26.* |
| loadjob | C27 | E.C27.* |
| tstats | C28 | E.C28.* |
| mstats | C29 | E.C29.* |
| timechart | C30 | E.C30.* |
| timewrap | C31 | E.C31.* |
| makemv | C32 | E.C32.* |
| mvexpand | C33 | E.C33.* |
| mvcombine | C34 | E.C34.* |
| fillnull | C35 | E.C35.* |

| Language family | Base | Expansion |
| --- | --- | --- |
| starts | L01 | E.L01.* |
| identifiers | L02 | E.L02.* |
| literals | L03 | E.L03.* |
| comments | L04 | E.L04.* |
| search predicates | L05 | E.C01.* |
| expression precedence | L06 | E.L06.* |
| conditional predicates | L07 | E.L07.* |
| arrays/objects | L08 | E.L08.* |
| access paths | L09 | E.L09.* |
| templates | L10 | E.L10.* |
| inline lambdas | L11 | E.L11.* |
| projection/rename lists | L12 | E.C05.*, C06.*, C07.* |
| aggregation/window forms | L13 | E.C08.*, C09.*, C10.*, F.* |
| lookup lists | L14 | E.C11.* |
| FROM-first SQL | L15 | E.L15.* |
| SELECT-first SQL | L16 | E.L16.* |
| SQL followed by pipes | L17 | E.L17.* |
| dataset literals | L18 | E.L18.* |
| SQL joins | L19 | E.L19.* |
| pipeline joins/subsearches | L20 | E.C16.*, C17.*, C18.*, C19.* |
| union | L21 | E.C20.* |
| metrics/data models | L22 | E.C28.*, C29.* |
| extraction/options | L23 | E.C23.*, C24.*, C25.* |
| embedded SPL | L24 | E.C22.*, L24.search_literal |
| conditional paths | L25 | E.C21.* |

## Shared domains, delimiters and capability boundaries

Commands use their evidenced lower-case spellings and source-specific case rules; H11 still holds unreviewed command/operator casing. SQL uses the documented keyword families. Search and expression Boolean precedence differ; only search receives adjacency AND. Quoted identifiers, strings, raw strings, regex, templates and embedded searches retain distinct token categories. Do not broaden one command's parameter quoting to every command. Boolean option values are the documented lowercase `true`/`false`.

Commas separate fields, aliases, aggregate lists and most SQL lists. Parentheses delimit calls/predicates; square brackets delimit independent searches, inherited subpipes, arrays or named aggregate/byfields lists according to context. They are not interchangeable. Stats/eventstats grouping spans follow grouping fields; streamstats normative BY/options order is separately evidenced and its contradictory examples remain H02. SQL ordering alternatives, hidden projection visibility, and disputed timechart aggregate separators remain held under their original IDs.

Span units require **command-specific** domains. Bin's own table lists seconds/minutes/hours/days/months and millisecond/centisecond/decisecond spellings; its examples corroborate short m/h/d. The general time-span vocabulary alone does not prove every bin alias or weekly span (EH01). Timechart separately documents seconds s, minutes m, hours h, days d, weeks w, months mon and us/ms/cs/ds subdivisions. Timewrap distinguishes m=month from min=minute and documents quarter/year families. For each admitted unit class, the eventual fixture must distinguish its token span from a trailing field, plus missing magnitude/unit/snap operand as applicable. Do not pad fixtures with synonymous spellings.

Repeated/default options are not silently discarded. Each documented option gets a parse-tree position/value assertion and an omission/default counterpart. Unknown unreviewed options stay U; specifically documented removed options retain base negative contracts. Domain restrictions in E rows are limited to explicitly evidenced static option literals; they do not promise evaluating dynamic field values. Deprecation alone does not make a search invalid.

K effects preserve only known fields and tombstones, including known internals. Open schemas do not fabricate _raw/_time. Exact null assignments remove fields; non-nullability requires an actual proof, not just a known pure-function name. Aggregate/group/order/SELECT references keep original lexical locations while lineage follows real evaluation order. SELECT aggregation and final projection remain distinct phases under the approved design; no fake lexical stages or projection attributed to HAVING.

## Command and language obligations

### E.C01.literal

Syntax: search followed by a nonempty term or double-quoted phrase; implicit AND between adjacent search terms. Target: **A/K**. Source: [doc163](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/search-command/search-command-overview-and-syntax). Base: C01;L01;L05.

Keyword RHS remains literal; assert original search stage and selector fields.

```text
E.C01.literal.P1  search index=app failure
E.C01.literal.P2  search index=app "disk nearly full"
E.C01.literal.N1  search index=app "disk nearly full
E.C01.literal.N2  search index=app severity=
```

### E.C01.comparison

Syntax: field (= | != | < | <= | > | >=) literal. Target: **A/K**. Source: [doc163](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/search-command/search-command-overview-and-syntax). Base: C01-P1.

Schedule separate tree assertions for each operator; do not apply expression RHS field reads.

```text
E.C01.comparison.P1  search index=app retries!=0
E.C01.comparison.P2  search index=app duration>=2.5
E.C01.comparison.N1  search index=app retries!=
E.C01.comparison.N2  search index=app duration>=
```

### E.C01.in

Syntax: field IN (literal, literal, ...); documented list is >=2; wildcard values allowed. Target: **A/K**. Source: [doc163](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/search-command/search-command-overview-and-syntax). Base: C01-P2,N2;L05-N1.

Single-element search IN has weak evidence: mark extra-form incomplete pending confirmation, not invented definite error.

```text
E.C01.in.P1  search index=app code IN (401,403)
E.C01.in.P2  search index=app code IN (50*,404)
E.C01.in.N1  search index=app code IN (401 403)
E.C01.in.N2  search index=app code IN (50*,)
```

### E.C01.term

Syntax: TERM(term), including minor-segmenter text. Target: **A/K**. Source: [doc163](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/search-command/search-command-overview-and-syntax). Base: L05-P2.

No field requirements inferred from directive content. Runtime indexing behavior not evaluated.

```text
E.C01.term.P1  search index=app TERM(192.0.2.19)
E.C01.term.P2  search index=app marker=TERM(error_code)
E.C01.term.N1  search index=app TERM()
E.C01.term.N2  search index=app TERM(192.0.2.19
```

### E.C01.case

Syntax: CASE(term) case-sensitive search directive. Target: **A/K**. Source: [doc163](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/search-command/search-command-overview-and-syntax). Base: C01.

Distinct from lowercase scalar case(predicate,value,...).

```text
E.C01.case.P1  search index=app CASE(Failure)
E.C01.case.P2  search index=app marker=CASE(WARN)
E.C01.case.N1  search index=app CASE()
E.C01.case.N2  search index=app marker=CASE(WARN
```

### E.C01.boolean

Syntax: Search precedence: parentheses, NOT, OR, AND, XOR; adjacency AND only search. Target: **A/K**. Source: [predicate](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-manual/expressions-and-predicates/predicate-expressions). Base: L05-P1.

Require operator-tree expectations, not merely field lists. H11 casing boundaries remain held.

```text
E.C01.boolean.P1  search index=app NOT severity=debug OR code=503 AND retry=1
E.C01.boolean.P2  search index=app (code=400 OR code=404) XOR severity=error
E.C01.boolean.N1  search index=app NOT
E.C01.boolean.N2  search index=app (code=400 OR)
```

### E.C01.selector

Syntax: index/source/sourcetype/host source modifiers with literal values in source-selector positions. Target: **A/K**. Source: [doc163](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/search-command/search-command-overview-and-syntax). Base: C01.

Do not select a validation schema from arbitrary value predicates. Piped search source roles must use grammar context.

```text
E.C01.selector.P1  search index=app sourcetype=access source="gateway.log"
E.C01.selector.P2  search index=audit host="api-2"
E.C01.selector.N1  search index=app sourcetype=
E.C01.selector.N2  search index=audit host="api-2
```

### E.C01.relative_time

Syntax: earliest/latest = signed optional-count unit, optional @snap and post-snap offset; snap-only legal. Target: **A/K**. Source: [relativetime](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-manual/dates-and-time/specifying-relative-time). Base: L05-P2;C02 SQL WHERE time range.

Time unit and snap grammar; no clock evaluation. Comparison operator controversy is EH03.

```text
E.C01.relative_time.P1  search index=app earliest=-2h@h latest=@h
E.C01.relative_time.P2  search index=app earliest=@d+3h latest=now()
E.C01.relative_time.N1  search index=app earliest=-2h@
E.C01.relative_time.N2  search index=app earliest=@d+
```

### E.C01.absolute_time

Syntax: earliest/latest = epoch integer or double-quoted timestamp; timeformat is quoted format for starttime/endtime. Target: **A/K**. Source: [timemodifiers](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-manual/dates-and-time/time-modifiers). Base: C01.

Malformed quotation versus invalid date/value separate. Do not infer custom-format semantics for earliest from starttime rule.

```text
E.C01.absolute_time.P1  search index=app earliest=1 latest=now()
E.C01.absolute_time.P2  search index=app timeformat="%Y-%m-%d" starttime="2026-01-02" endtime="2026-01-03"
E.C01.absolute_time.N1  search index=app earliest=
E.C01.absolute_time.N2  search index=app timeformat="%Y-%m-%d starttime="2026-01-02"
```

### E.C01.index_time

Syntax: _index_earliest/_index_latest use the relative-time modifier syntax. Target: **A/K**. Source: [timemodifiers](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-manual/dates-and-time/time-modifiers). Base: C01.

Index-time constraints require a suitable event window at runtime; offline analyzer does not simulate retrieval.

```text
E.C01.index_time.P1  search index=app earliest=-1d _index_earliest=-h@h
E.C01.index_time.P2  search index=app earliest=1 _index_latest=@h
E.C01.index_time.N1  search index=app _index_earliest=@
E.C01.index_time.N2  search index=app _index_latest=-
```

### E.C03.assign

Syntax: target=expression; comma-separated assignments; exact target and field-template target distinct. Target: **A/K or U dynamic**. Source: [doc136](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/eval-command/eval-command-overview-and-syntax). Base: C03;L02;L03;L10.

Left-to-right read/write order; null removal; Boolean R01, templates U. Missing required value is structural invalid.

```text
E.C03.assign.P1  eval initial=bytes, total=initial+5
E.C03.assign.P2  eval 'display name'=user, gone=null
E.C03.assign.N1  eval initial=bytes total=initial+5
E.C03.assign.N2  eval 'display name'=
```

### E.C04.predicate

Syntax: where predicate; explicit AND/OR/NOT/XOR separates predicate expressions. Target: **A/K**. Source: [doc178](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/where-command/where-command-overview-syntax-and-usage). Base: C04;L06;L07.

Expression equality RHS is a field read; both = and == admitted by predicate docs; no arbitrary implicit AND.

```text
E.C04.predicate.P1  where left_value=right_value AND amount>0
E.C04.predicate.P2  where NOT active=true OR isnull(owner)
E.C04.predicate.N1  where left_value=right_value amount>0
E.C04.predicate.N2  where NOT
```

### E.C05.include

Syntax: fields [+] selector (, selector)*. Target: **A/K universe-dependent**. Source: [doc139](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/fields-command/fields-command-overview-syntax-and-usage). Base: C05-P1;L12.

Finite/partial source universe affects wildcard completeness; preserved known internals/tombstones tested.

```text
E.C05.include.P1  fields user, bytes
E.C05.include.P2  fields + 'user*', host
E.C05.include.N1  fields user bytes
E.C05.include.N2  fields +
```

### E.C05.exclude

Syntax: fields - selector (, selector)*; single list-wide sign. Target: **A/K universe-dependent**. Source: [doc139](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/fields-command/fields-command-overview-syntax-and-usage). Base: C05-P2.

Do not add per-item +/- syntax without evidence; exact and wildcard exclusion share kernel.

```text
E.C05.exclude.P1  fields - debug, trace
E.C05.exclude.P2  fields - 'temp*', 'private.*'
E.C05.exclude.N1  fields - debug trace
E.C05.exclude.N2  fields -
```

### E.C06.exact

Syntax: table field (, field)*; no AS or options. Target: **A/K**. Source: [doc169](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/table-command/table-command-overview-syntax-and-usage). Base: C06.

H01 all wildcard quote variants remain separately incomplete; do not silently expand exact table rule.

```text
E.C06.exact.P1  table user, bytes
E.C06.exact.P2  table 'display name', 'literal.path'
E.C06.exact.N1  table user bytes
E.C06.exact.N2  table user AS name
```

### E.C07.exact

Syntax: rename source AS target (, source AS target)*; independent pairs only. Target: **A/K**. Source: [doc158](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/rename-command/rename-command-overview-syntax-and-usage). Base: C07;B13-B16.

R02 duplicate-source/target/chain/overlap invalid contract even if grammar parses. Use B13–B16 and L12 negatives.

```text
E.C07.exact.P1  rename user AS owner
E.C07.exact.P2  rename 'request.id' AS request_id, host AS machine
E.C07.exact.N1  rename user AS
E.C07.exact.N2  rename user AS owner host AS machine
```

### E.C07.wildcard

Syntax: Single-quoted source/target wildcard patterns, e.g. prefix* AS replacement*. Target: **A/K finite, otherwise U**. Source: [renameexamples](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/rename-command/rename-command-examples). Base: C07.

Do not invent wildcard-count errors from wording alone; M3 finite mapping ambiguity/collision/tombstone evidence controls semantics.

```text
E.C07.wildcard.P1  rename 'old_*' AS 'new_*'
E.C07.wildcard.P2  rename 'event.*' AS '*'
E.C07.wildcard.N1  rename 'old_*' AS
E.C07.wildcard.N2  rename 'event.*' AS '*
```

### E.C08.calls

Syntax: stats aggregation (, aggregation)* [BY field [span=timespan] (, field [span=timespan])*]. Target: **A/K**. Source: [doc167](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/stats-command/stats-command-overview-syntax-and-usage). Base: C08;L13.

Count output count, core aggregate explicit/implicit names; pure-function rows cover arities.

```text
E.C08.calls.P1  stats count(),sum(bytes) AS total
E.C08.calls.P2  stats avg(duration) AS mean BY host,service
E.C08.calls.N1  stats count() sum(bytes) AS total
E.C08.calls.N2  stats avg(duration) AS mean BY host service
```

### E.C08.by_span

Syntax: BY timestamp-field span=timespan, other-field; no wildcard grouping. Target: **A/K exact group or U span effect**. Source: [doc167](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/stats-command/stats-command-overview-syntax-and-usage). Base: C08.

Unit-only timespan evidenced in general time-span doc; timestamp type not assumed from name; unsupported span mapping U.

```text
E.C08.by_span.P1  stats count() BY _time span=10min
E.C08.by_span.P2  stats sum(bytes) BY event_time span=hour,host
E.C08.by_span.N1  stats count() BY _time span=
E.C08.by_span.N2  stats sum(bytes) BY 'host*'
```

### E.C09.calls

Syntax: eventstats aggregates [BY exact comma field list with optional span]. Target: **A/K**. Source: [eventtail](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/eventstats-command/eventstats-command-overview-syntax-and-usage). Base: C09.

Same pure aggregate registry, preserves input; apparent [bool] synopsis typo resolved by explicit by-clause.

```text
E.C09.calls.P1  eventstats count() AS rows BY host
E.C09.calls.P2  eventstats values(user) AS users,sum(bytes) AS total
E.C09.calls.N1  eventstats count() AS rows BY
E.C09.calls.N2  eventstats values(user) AS users sum(bytes) AS total
```

### E.C09.by_span

Syntax: BY timestamp-field span=timespan (, field [span=timespan])*. Target: **A/K exact group or U span effect**. Source: [eventtail](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/eventstats-command/eventstats-command-overview-syntax-and-usage). Base: C09.

Do not fabricate precise per-span output membership absent shared proof.

```text
E.C09.by_span.P1  eventstats count() BY _time span=5min
E.C09.by_span.P2  eventstats sum(bytes) AS total BY event_time span=2hr,host
E.C09.by_span.N1  eventstats count() BY _time span=
E.C09.by_span.N2  eventstats sum(bytes) BY host service
```

### E.C10.calls

Syntax: streamstats [BY fields] [current=bool] [reset...] [window=int] aggregates. Target: **A/K ordinary or U conditional**. Source: [doc168](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/streamstats-command/streamstats-command-overview-syntax-and-usage). Base: C10;L13-P2.

Normative options before aggregates only; H02 post-aggregate layouts stay incomplete.

```text
E.C10.calls.P1  streamstats count() AS seen
E.C10.calls.P2  streamstats BY host sum(bytes) AS total,avg(bytes) AS mean
E.C10.calls.N1  streamstats count() AS
E.C10.calls.N2  streamstats BY host sum(bytes) AS total avg(bytes) AS mean
```

### E.C10.reset_before

Syntax: reset before predicate before aggregate. Target: **A/K reads; conditional output U if unproved**. Source: [doc168](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/streamstats-command/streamstats-command-overview-syntax-and-usage). Base: C10.

Do not consume aggregate as missing reset predicate via generic fallback.

```text
E.C10.reset_before.P1  streamstats reset before action="START" count() AS n
E.C10.reset_before.P2  streamstats BY host reset before code>=500 sum(bytes) AS total
E.C10.reset_before.N1  streamstats reset before
E.C10.reset_before.N2  streamstats reset before action= count()
```

### E.C10.reset_after

Syntax: reset after predicate before aggregate. Target: **A/K reads; conditional output U if unproved**. Source: [doc168](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/streamstats-command/streamstats-command-overview-syntax-and-usage). Base: C10.

Evaluation applies after current event; don't collapse before/after or read postaggregate alias without evidence.

```text
E.C10.reset_after.P1  streamstats reset after action="STOP" count() AS n
E.C10.reset_after.P2  streamstats reset after amount>10 window=4 sum(amount) AS running
E.C10.reset_after.N1  streamstats reset after
E.C10.reset_after.N2  streamstats reset after amount> window=4 sum(amount)
```

### E.C10.reset_onchange

Syntax: reset onchange, optionally combined with before/after, before aggregate. Target: **A/K grouping reads; conditional output U if unproved**. Source: [doc168](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/streamstats-command/streamstats-command-overview-syntax-and-usage). Base: C10.

Onchange uses grouping fields. Combination OR resets documented; preserve source order/reads.

```text
E.C10.reset_onchange.P1  streamstats BY host reset onchange count() AS n
E.C10.reset_onchange.P2  streamstats BY host reset before code=500 after action="STOP" onchange sum(bytes)
E.C10.reset_onchange.N1  streamstats BY host reset
E.C10.reset_onchange.N2  streamstats BY host reset before after action="STOP" sum(bytes)
```

### E.C11.matches

Syntax: lookup dataset match [AS local] (,match [AS local])*. Target: **A/U absent output shape**. Source: [doc150](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/lookup-command/lookup-command-overview-syntax-and-usage). Base: C11;L14.

Catalog match columns are not source fields; aliases identify local inputs.

```text
E.C11.matches.P1  lookup users uid AS user
E.C11.matches.P2  lookup accounts id AS account, realm AS domain
E.C11.matches.N1  lookup users
E.C11.matches.N2  lookup accounts id AS account realm AS domain
```

### E.C11.output

Syntax: uppercase OUTPUT dest [AS local] (,dest [AS local])*. Target: **A/K explicit**. Source: [doc150](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/lookup-command/lookup-command-overview-syntax-and-usage). Base: C11.

Overwrites target; alias/location roles distinct. lowercase output is proven unsupported by usage.

```text
E.C11.output.P1  lookup users uid AS user OUTPUT display AS name
E.C11.output.P2  lookup accounts id OUTPUT email, team AS group_name
E.C11.output.N1  lookup users uid AS user OUTPUT
E.C11.output.N2  lookup accounts id OUTPUT email team
```

### E.C11.outputnew

Syntax: uppercase OUTPUTNEW dest [AS local] (,dest [AS local])*. Target: **A/K conditional**. Source: [doc150](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/lookup-command/lookup-command-overview-syntax-and-usage). Base: C11-P2;L14-P2.

Preserves existing local destination; source-universe membership affects overwrite vs conditional creation.

```text
E.C11.outputnew.P1  lookup users uid OUTPUTNEW display AS name
E.C11.outputnew.P2  lookup accounts id OUTPUTNEW email, team
E.C11.outputnew.N1  lookup users uid OUTPUTNEW
E.C11.outputnew.N2  lookup accounts id OUTPUTNEW email team
```

### E.C12.order_terms

Syntax: sort [integer-count] [+|-][auto|ip|num|str](field) or signed plain field, comma separated. Target: **A/K**. Source: [sortexamples](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/sort-command/sort-command-examples). Base: C12.

Each wrapper and plain +/- term receives exact reference-slice assertion; wrapper names not fields.

```text
E.C12.order_terms.P1  sort -host,+num(bytes)
E.C12.order_terms.P2  sort auto(status),-ip(client),str(user)
E.C12.order_terms.N1  sort -host +num(bytes)
E.C12.order_terms.N2  sort auto(),-ip(client)
```

### E.C12.count

Syntax: sort positional integer count before fields; 0 means all. Target: **A/K**. Source: [doc164](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/sort-command/sort-command-overview-syntax-and-usage). Base: C12.

Do not reject omitted count/default10000; data ordering not evaluated.

```text
E.C12.count.P1  sort 0 bytes
E.C12.count.P2  sort 5 -bytes,user
E.C12.count.N1  sort 0
E.C12.count.N2  sort 5.5 -bytes
```

### E.C13.count

Syntax: dedup [positive integer] [keepempty=bool] [consecutive=bool] comma fields. Target: **A/K**. Source: [doc135](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/dedup-command/dedup-command-overview-syntax-and-usage). Base: C13.

All options precede fields; default one; sortby/keepevents excluded.

```text
E.C13.count.P1  dedup 1 user
E.C13.count.P2  dedup 3 user,host
E.C13.count.N1  dedup 0 user
E.C13.count.N2  dedup 2.5 user
```

### E.C14.while

Syntax: head [keeplast=bool] [while (predicate)] [integer]. Target: **A/K ordinary predicates**. Source: [doc144](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/head-command/head-command-overview-syntax-and-usage). Base: C14.

Statistical functions forbidden in while; count-before-while is H03, not negative.

```text
E.C14.while.P1  head while (amount>0)
E.C14.while.P2  head keeplast=false while (isnotnull(user)) 5
E.C14.while.N1  head while amount>0
E.C14.while.N2  head while (amount>)
```

### E.C14.count

Syntax: head optional positional integer after other options; omitted default10. Target: **A/K**. Source: [doc144](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/head-command/head-command-overview-syntax-and-usage). Base: C14-P1,N1.

H03 alternative placement remains incomplete; head0/range not guessed from int syntax.

```text
E.C14.count.P1  head
E.C14.count.P2  head 4
E.C14.count.N1  head limit=4
E.C14.count.N2  head 4.5
```

### E.C08.allnum

Syntax: allnum=true|false before command arguments. Target: **A/K conditional or U**. Source: [doc167](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/stats-command/stats-command-overview-syntax-and-usage). Base: C08.

Numerical allnum=true groups can suppress outputs; false must not acquire that rule.

```text
E.C08.allnum.P1  stats allnum=true count()
E.C08.allnum.P2  stats allnum=false sum(bytes) AS total BY host
E.C08.allnum.N1  stats allnum=1 count()
E.C08.allnum.N2  stats allnum=TRUE count()
```

### E.C09.allnum

Syntax: allnum=true|false before command arguments. Target: **A/K conditional or U**. Source: [eventtail](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/eventstats-command/eventstats-command-overview-syntax-and-usage). Base: C09.

Conditional absence applies per group, not unconditional derived availability.

```text
E.C09.allnum.P1  eventstats allnum=true count() AS n
E.C09.allnum.P2  eventstats allnum=false avg(bytes) AS average BY host
E.C09.allnum.N1  eventstats allnum=1 count() AS n
E.C09.allnum.N2  eventstats allnum=TRUE count() AS n
```

### E.C10.current

Syntax: current=true|false before command arguments. Target: **A/K conditional or U**. Source: [doc168](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/streamstats-command/streamstats-command-overview-syntax-and-usage). Base: C10.

false excludes current event, empty initial history must not promise output presence.

```text
E.C10.current.P1  streamstats current=true count() AS n
E.C10.current.P2  streamstats current=false window=3 avg(bytes) AS moving
E.C10.current.N1  streamstats current=1 count() AS n
E.C10.current.N2  streamstats current=TRUE count() AS n
```

### E.C13.keepempty

Syntax: keepempty=true|false before command arguments. Target: **A/K**. Source: [doc135](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/dedup-command/dedup-command-overview-syntax-and-usage). Base: C13.

Row preservation differs; does not create a missing grouping field.

```text
E.C13.keepempty.P1  dedup keepempty=true user
E.C13.keepempty.P2  dedup keepempty=false user,host
E.C13.keepempty.N1  dedup keepempty=1 user
E.C13.keepempty.N2  dedup keepempty=TRUE user
```

### E.C13.consecutive

Syntax: consecutive=true|false before command arguments. Target: **A/K**. Source: [doc135](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/dedup-command/dedup-command-overview-syntax-and-usage). Base: C13.

Consecutive versus global row dedup, field membership unchanged.

```text
E.C13.consecutive.P1  dedup consecutive=true user
E.C13.consecutive.P2  dedup consecutive=false user,host
E.C13.consecutive.N1  dedup consecutive=1 user
E.C13.consecutive.N2  dedup consecutive=TRUE user
```

### E.C14.keeplast

Syntax: keeplast=true|false before command arguments. Target: **A/K**. Source: [doc144](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/head-command/head-command-overview-syntax-and-usage). Base: C14.

False/true decides boundary event; no fields created.

```text
E.C14.keeplast.P1  head keeplast=true while (bytes>0) 3
E.C14.keeplast.P2  head keeplast=false while (isnotnull(user)) 4
E.C14.keeplast.N1  head keeplast=1 while (bytes>0) 3
E.C14.keeplast.N2  head keeplast=TRUE while (bytes>0) 3
```

### E.C18.run_in_preview

Syntax: run_in_preview=true|false before command arguments. Target: **A/U**. Source: [doc130](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/appendpipe-command/appendpipe-command-overview-syntax-and-usage). Base: C18.

Inherited child copy and unknown parent/child merge retained.

```text
E.C18.run_in_preview.P1  appendpipe run_in_preview=true [stats count()]
E.C18.run_in_preview.P2  appendpipe run_in_preview=false [where bytes>0 | fields bytes]
E.C18.run_in_preview.N1  appendpipe run_in_preview=1 [stats count()]
E.C18.run_in_preview.N2  appendpipe run_in_preview=TRUE [stats count()]
```

### E.C30.partial

Syntax: partial=true|false before command arguments. Target: **A/U**. Source: [doc172](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/timechart-command/timechart-command-overview-and-syntax). Base: C30.

Partial end bins affect rows, unmodeled series shape.

```text
E.C30.partial.P1  timechart partial=true count()
E.C30.partial.P2  timechart partial=false avg(bytes) BY host
E.C30.partial.N1  timechart partial=1 count()
E.C30.partial.N2  timechart partial=TRUE count()
```

### E.C30.cont

Syntax: cont=true|false before command arguments. Target: **A/U**. Source: [doc172](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/timechart-command/timechart-command-overview-and-syntax). Base: C30.

Gap filling can produce time rows; do not silently preserve complete field set.

```text
E.C30.cont.P1  timechart cont=true count()
E.C30.cont.P2  timechart cont=false avg(bytes) BY host
E.C30.cont.N1  timechart cont=1 count()
E.C30.cont.N2  timechart cont=TRUE count()
```

### E.C30.fixedrange

Syntax: fixedrange=true|false before command arguments. Target: **A/U**. Source: [doc172](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/timechart-command/timechart-command-overview-and-syntax). Base: C30.

Changes covered range only at execution; no clock inference.

```text
E.C30.fixedrange.P1  timechart fixedrange=true count()
E.C30.fixedrange.P2  timechart fixedrange=false avg(bytes) BY host
E.C30.fixedrange.N1  timechart fixedrange=1 count()
E.C30.fixedrange.N2  timechart fixedrange=TRUE count()
```

### E.C30.usenull

Syntax: usenull=true|false after BY split field. Target: **A/U**. Source: [doc172](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/timechart-command/timechart-command-overview-and-syntax). Base: timechartusage.

Suffix-only option: wrong prefix placement invalid per explicit usage; do not confuse time-axis prefix bin options.

```text
E.C30.usenull.P1  timechart count() BY host usenull=true
E.C30.usenull.P2  timechart avg(bytes) BY service usenull=false
E.C30.usenull.N1  timechart count() BY host usenull=1
E.C30.usenull.N2  timechart usenull=false count() BY host
```

### E.C30.useother

Syntax: useother=true|false after BY split field. Target: **A/U**. Source: [doc172](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/timechart-command/timechart-command-overview-and-syntax). Base: timechartusage.

Suffix-only option: wrong prefix placement invalid per explicit usage; do not confuse time-axis prefix bin options.

```text
E.C30.useother.P1  timechart count() BY host useother=true
E.C30.useother.P2  timechart avg(bytes) BY service useother=false
E.C30.useother.N1  timechart count() BY host useother=1
E.C30.useother.N2  timechart useother=false count() BY host
```

### E.C08.delim

Syntax: delim=double-quoted string before aggregate. Target: **A/K**. Source: [doc167](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/stats-command/stats-command-overview-syntax-and-usage). Base: C08.

Delimiter is literal, not field read.

```text
E.C08.delim.P1  stats delim="::" values(user) AS users
E.C08.delim.P2  stats delim="($AGG$)-$VAL$" list(host) AS hosts
E.C08.delim.N1  stats delim= values(user) AS users
E.C08.delim.N2  stats delim="unterminated values(user) AS users
```

### E.C30.sep

Syntax: sep=double-quoted string before aggregate. Target: **A/U**. Source: [doc172](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/timechart-command/timechart-command-overview-and-syntax). Base: C30.

Separator contributes dynamic series names.

```text
E.C30.sep.P1  timechart sep="::" count() BY host
E.C30.sep.P2  timechart sep="($AGG$)-$VAL$" avg(bytes) BY service
E.C30.sep.N1  timechart sep= count() BY host
E.C30.sep.N2  timechart sep="unterminated count() BY host
```

### E.C30.format

Syntax: format=double-quoted string before aggregate. Target: **A/U**. Source: [doc172](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/timechart-command/timechart-command-overview-and-syntax). Base: C30.

format overrides sep; $AGG$/$VAL$ placeholders in option string are not field-template ${...}.

```text
E.C30.format.P1  timechart format="::" count() BY host
E.C30.format.P2  timechart format="($AGG$)-$VAL$" avg(bytes) BY service
E.C30.format.N1  timechart format= count() BY host
E.C30.format.N2  timechart format="unterminated count() BY host
```

### E.C30.nullstr

Syntax: nullstr=string after BY split field. Target: **A/U**. Source: [doc172](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/timechart-command/timechart-command-overview-and-syntax). Base: timechartusage.

Controls label only when corresponding usenull/useother enabled; option remains grammatical when effect disabled.

```text
E.C30.nullstr.P1  timechart count() BY host nullstr="missing"
E.C30.nullstr.P2  timechart avg(bytes) BY service nullstr="other series"
E.C30.nullstr.N1  timechart count() BY host nullstr=
E.C30.nullstr.N2  timechart nullstr="missing" count() BY host
```

### E.C30.otherstr

Syntax: otherstr=string after BY split field. Target: **A/U**. Source: [doc172](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/timechart-command/timechart-command-overview-and-syntax). Base: timechartusage.

Controls label only when corresponding usenull/useother enabled; option remains grammatical when effect disabled.

```text
E.C30.otherstr.P1  timechart count() BY host otherstr="missing"
E.C30.otherstr.P2  timechart avg(bytes) BY service otherstr="other series"
E.C30.otherstr.N1  timechart count() BY host otherstr=
E.C30.otherstr.N2  timechart otherstr="missing" count() BY host
```

### E.C08.partitions

Syntax: partitions=num before aggregate, default1. Target: **A/K**. Source: [doc167](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/stats-command/stats-command-overview-syntax-and-usage). Base: C08.

Do not invent fractional/range restrictions from execution rationale alone.

```text
E.C08.partitions.P1  stats partitions=1 count()
E.C08.partitions.P2  stats partitions=3 sum(bytes) BY host
E.C08.partitions.N1  stats partitions= count()
E.C08.partitions.N2  stats partitions="many" count()
```

### E.C10.window

Syntax: window=nonnegative integer before aggregate, 0 unbounded. Target: **A/K conditional or U**. Source: [doc168](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/streamstats-command/streamstats-command-overview-syntax-and-usage). Base: C10-P1,N2.

Empty history/absent values require conditional output proof.

```text
E.C10.window.P1  streamstats window=0 count()
E.C10.window.P2  streamstats current=false window=3 avg(bytes)
E.C10.window.N1  streamstats window=-1 count()
E.C10.window.N2  streamstats window=2.5 count()
```

### E.C16.type

Syntax: join type=inner|left|outer in pre-where options with required aliases. Target: **A/U**. Source: [doc148](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/join-command/join-command-overview-syntax-and-usage). Base: C16.

Also cover left explicitly; outer means left outer, not full. Both evidenced pre-where option positions.

```text
E.C16.type.P1  join type=inner left=L right=R where L.id=R.id [FROM other]
E.C16.type.P2  join left=L right=R type=outer where L.id=R.id [FROM other]
E.C16.type.N1  join type=full left=L right=R where L.id=R.id [FROM other]
E.C16.type.N2  join type= left=L right=R where L.id=R.id [FROM other]
```

### E.C16.max

Syntax: join max=int before where; 0 unlimited, default1. Target: **A/U**. Source: [doc148](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/join-command/join-command-overview-syntax-and-usage). Base: C16.

Negative maximum/range not inferred from int alone; merge still U.

```text
E.C16.max.P1  join max=0 left=L right=R where L.id=R.id [FROM other]
E.C16.max.P2  join left=L right=R max=2 where L.id=R.id [FROM other]
E.C16.max.N1  join max= left=L right=R where L.id=R.id [FROM other]
E.C16.max.N2  join max=2.5 left=L right=R where L.id=R.id [FROM other]
```

### E.C16.alias_predicate

Syntax: left=alias right=alias where alias.field=alias.field [AND ...] [FROM independent search]. Target: **A/U**. Source: [doc148](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/join-command/join-command-overview-syntax-and-usage). Base: C16;L20.

Equality orientation can reverse; alias qualifiers not root literal dotted fields.

```text
E.C16.alias_predicate.P1  join left=A right=B where A.uid=B.id [FROM directory]
E.C16.alias_predicate.P2  join left=A right=B where B.id=A.uid AND A.realm=B.realm [FROM directory | fields id,realm]
E.C16.alias_predicate.N1  join left=A where A.uid=B.id [FROM directory]
E.C16.alias_predicate.N2  join left=A right=B where A.uid=B.id [directory]
```

### E.C17.subsearch

Syntax: append [nonempty independent search]; no options. Target: **A/U**. Source: [doc129](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/append-command/append-command-overview-syntax-and-usage). Base: C17.

Child independent; historical-only execution caveat not grammar error.

```text
E.C17.subsearch.P1  append [FROM archive SELECT user]
E.C17.subsearch.P2  append [search index=old | stats count()]
E.C17.subsearch.N1  append []
E.C17.subsearch.N2  append maxtime=30 [FROM archive]
```

### E.C18.subpipe

Syntax: appendpipe [nonempty inherited subpipe]. Target: **A/U**. Source: [doc130](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/appendpipe-command/appendpipe-command-overview-syntax-and-usage). Base: C18;L20-P2.

Transforming-stage contextual placement separately checked; no output leak from unmodeled merge.

```text
E.C18.subpipe.P1  appendpipe [fields user]
E.C18.subpipe.P2  appendpipe [where bytes>0 | eval extra=bytes]
E.C18.subpipe.N1  appendpipe []
E.C18.subpipe.N2  appendpipe [where bytes>0 |]
```

### E.C19.subsearch

Syntax: appendcols [nonempty independent search]; no override/time/maxout options. Target: **A/U**. Source: [doc128](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/appendcols-command/appendcols-command-overview-syntax-and-usage). Base: C19.

Main value wins on collisions; child internals excluded; row alignment unknown.

```text
E.C19.subsearch.P1  appendcols [FROM other | fields user]
E.C19.subsearch.P2  appendcols [FROM totals SELECT count()]
E.C19.subsearch.N1  appendcols []
E.C19.subsearch.N2  appendcols override=true [FROM other]
```

### E.C20.dataset_types

Syntax: union comma-separated plain, row-literal or bracketed-subsearch datasets. Target: **A/U**. Source: [doc176](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/union-command/union-command-overview-syntax-and-usage). Base: C20.

Generating minimum2, piped minimum1. Existing C20/L21 cover minima and independent scope.

```text
E.C20.dataset_types.P1  union main, [{user:"local"}]
E.C20.dataset_types.P2  FROM main | union [FROM archive SELECT user], other
E.C20.dataset_types.N1  union main [{user:"local"}]
E.C20.dataset_types.N2  FROM main | union [FROM archive SELECT user],
```

### E.C21.paths

Syntax: if (predicate)[subpipe] elseif (predicate)[subpipe]* else [subpipe]?. Target: **A/U**. Source: [doc145](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/if-command/if-command-overview-syntax-and-usage). Base: C21;L25.

No pipe before elseif/else; final else once; inherited copied branch state.

```text
E.C21.paths.P1  if (code=500) [eval rank=2] else [eval rank=0]
E.C21.paths.P2  if (code=400) [eval rank=1] elseif (code=500) [eval rank=2] elseif (code=503) [eval rank=3]
E.C21.paths.N1  if (code=500) []
E.C21.paths.N2  if (code=400) [eval rank=1] else [eval rank=0] elseif (code=500) [eval rank=2]
```

### E.C22.explicit

Syntax: spl1 double-quoted embedded SPL command source. Target: **A/U body and effects**. Source: [doc166](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/spl1-command/spl1-command-overview-syntax-and-usage). Base: C22;L24.

Wrapper's successful grammar doesn't validate embedded body. Macro/subsearch restrictions require grammar evidence.

```text
E.C22.explicit.P1  spl1 "search index=old | head 2"
E.C22.explicit.P2  FROM main | spl1 "stats count by host"
E.C22.explicit.N1  spl1 search index=old
E.C22.explicit.N2  FROM main | spl1 "stats count by host
```

### E.C22.backtick

Syntax: command-position backtick embedded SPL source. Target: **A/U body and effects**. Source: [doc166](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/spl1-command/spl1-command-overview-syntax-and-usage). Base: C22.

Expression-position backticks are L24.search_literal, not automatic SPL command body.

```text
E.C22.backtick.P1  `search index=old | head 2`
E.C22.backtick.P2  FROM main | `stats sum(bytes) AS total`
E.C22.backtick.N1  `search index=old | head 2
E.C22.backtick.N2  FROM main | `stats sum(bytes) AS total
```

### E.C23.bins

Syntax: bins=int before field. Target: **A/U**. Source: [doc131](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/bin-command/bin-command-overview-syntax-and-usage). Base: C23.

No output-model promotion; bins numeric-domain constraints beyond syntax only if documented.

```text
E.C23.bins.P1  bin bins=10 bytes
E.C23.bins.P2  bin bins=4 duration AS bucket
E.C23.bins.N1  bin bins= bytes
E.C23.bins.N2  bin bins=2.5 bytes
```

### E.C23.minspan

Syntax: minspan=span-length before field. Target: **A/U**. Source: [doc131](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/bin-command/bin-command-overview-syntax-and-usage). Base: C23.

Use explicit integer+unit witnesses; unit-only bin spans have narrower evidence than SQL/timewrap.

```text
E.C23.minspan.P1  bin minspan=2min _time
E.C23.minspan.P2  bin minspan=10ms stamp AS bucket
E.C23.minspan.N1  bin minspan= _time
E.C23.minspan.N2  bin minspan=2min
```

### E.C23.span_length

Syntax: span=integer[time-unit], numeric bin length without unit allowed. Target: **A/U**. Source: [doc131](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/bin-command/bin-command-overview-syntax-and-usage). Base: C23.

Command-specific aliases and general timespan table cross-audit in EH01. Do not make bare numeric span a syntax failure.

```text
E.C23.span_length.P1  bin span=50 bytes
E.C23.span_length.P2  bin span=2hr _time AS bucket
E.C23.span_length.N1  bin span= bytes
E.C23.span_length.N2  bin span=2hr
```

### E.C23.span_log

Syntax: span=[coefficient]log[base]; coefficient>=1,<base; base>1. Target: **A/U**. Source: [doc131](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/bin-command/bin-command-overview-syntax-and-usage). Base: C23.

N are documented numeric contract restrictions, not tokenizer failure. Numeric literals need no query execution.

```text
E.C23.span_log.P1  bin span=2log10 bytes
E.C23.span_log.P2  bin span=log2 duration
E.C23.span_log.N1  bin span=0.5log10 bytes
E.C23.span_log.N2  bin span=2log1 bytes
```

### E.C23.start

Syntax: start=num before field. Target: **A/U**. Source: [doc131](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/bin-command/bin-command-overview-syntax-and-usage). Base: C23.

Both legal together; only expand range when span absent; ignored combinations are not invalid.

```text
E.C23.start.P1  bin start=0 bytes
E.C23.start.P2  bin start=1000 bins=10 duration
E.C23.start.N1  bin start= bytes
E.C23.start.N2  bin start="large" bytes
```

### E.C23.end

Syntax: end=num before field. Target: **A/U**. Source: [doc131](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/bin-command/bin-command-overview-syntax-and-usage). Base: C23.

Both legal together; only expand range when span absent; ignored combinations are not invalid.

```text
E.C23.end.P1  bin end=0 bytes
E.C23.end.P2  bin end=1000 bins=10 duration
E.C23.end.N1  bin end= bytes
E.C23.end.N2  bin end="large" bytes
```

### E.C23.aligntime

Syntax: aligntime=earliest|latest|time-specifier before field; snap+offset evidenced. Target: **A/U**. Source: [binexamples](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/bin-command/bin-command-examples). Base: C23.

Additional positive obligation latest; ignored for day/month/year spans is not a parse error.

```text
E.C23.aligntime.P1  bin span=2hr aligntime=earliest _time
E.C23.aligntime.P2  bin span=12h aligntime=@d+3h _time
E.C23.aligntime.N1  bin span=2hr aligntime= _time
E.C23.aligntime.N2  bin span=12h aligntime=@d+ _time
```

### E.C23.alias

Syntax: field AS output after bin options. Target: **A/U**. Source: [doc131](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/bin-command/bin-command-overview-syntax-and-usage). Base: C23.

Output identifier is write target, not source read; do not reverse option placement.

```text
E.C23.alias.P1  bin span=5min _time AS bucket
E.C23.alias.P2  bin bins=5 bytes AS 'byte bucket'
E.C23.alias.N1  bin span=5min _time AS
E.C23.alias.N2  bin bins=5 bytes AS 'byte bucket
```

### E.C24.field

Syntax: rex field=identifier before regex/sed body; default _raw. Target: **A/U**. Source: [doc161](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/rex-command/rex-command-overview-syntax-and-usage). Base: C24.

Default _raw has no fabricated source span; literal dotted source and navigation differ.

```text
E.C24.field.P1  rex field=payload "(?<code>[0-9]+)"
E.C24.field.P2  rex field='raw message' @"(?<word>\w+)"
E.C24.field.N1  rex field= "(?<code>[0-9]+)"
E.C24.field.N2  rex "(?<code>[0-9]+)" field=payload
```

### E.C24.max_match

Syntax: max_match=int before regex; 0 unlimited/default1. Target: **A/U**. Source: [doc161](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/rex-command/rex-command-overview-syntax-and-usage). Base: C24.

Multiple matches produce multivalue, no static extraction promise.

```text
E.C24.max_match.P1  rex max_match=0 "(?<word>[a-z]+)"
E.C24.max_match.P2  rex field=payload max_match=3 @"(?<digit>\d)"
E.C24.max_match.N1  rex max_match= "(?<word>[a-z]+)"
E.C24.max_match.N2  rex max_match=1.5 "(?<word>[a-z]+)"
```

### E.C24.offset_field

Syntax: offset_field=output-identifier before regex. Target: **A/U**. Source: [doc161](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/rex-command/rex-command-overview-syntax-and-usage). Base: C24.

Offset target role retained as candidate; no effect completeness from matching regex text.

```text
E.C24.offset_field.P1  rex offset_field=offsets "(?<word>[a-z]+)"
E.C24.offset_field.P2  rex field=payload max_match=0 offset_field=positions "(?<n>[0-9]+)"
E.C24.offset_field.N1  rex offset_field= "(?<word>[a-z]+)"
E.C24.offset_field.N2  rex "(?<word>[a-z]+)" offset_field=offsets
```

### E.C24.quoted_regex

Syntax: double-quoted regex, including internal alternation pipe. Target: **A/U**. Source: [doc161](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/rex-command/rex-command-overview-syntax-and-usage). Base: C24.

H12 specifically remains slash+internal-pipe, not quoted alternation. Regex-body validity separate.

```text
E.C24.quoted_regex.P1  rex field=payload "(?<tag>red|blue)"
E.C24.quoted_regex.P2  rex "(?<word>alpha|beta)"
E.C24.quoted_regex.N1  rex field=payload "(?<tag>red|blue)
E.C24.quoted_regex.N2  rex "(?<word>alpha|beta)
```

### E.C24.raw_regex

Syntax: @ double-quoted raw regex, doubled quote escape inside raw string. Target: **A/U**. Source: [doc161](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/rex-command/rex-command-overview-syntax-and-usage). Base: C24.

Raw-string escaping uses global expression source; PCRE result unproved.

```text
E.C24.raw_regex.P1  rex @"(?<digit>\d+)"
E.C24.raw_regex.P2  rex field=payload @"(?<quoted>""word"")"
E.C24.raw_regex.N1  rex @"(?<digit>\d+)
E.C24.raw_regex.N2  rex field=payload @"(?<quoted>word)
```

### E.C24.slash_regex

Syntax: slash regex with evidenced character-class form; internal pipe H12. Target: **A/U**. Source: [doc161](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/rex-command/rex-command-overview-syntax-and-usage). Base: L23-P1,N1.

Preserve approved L23 correction and do not add slash flags without standalone evidence.

```text
E.C24.slash_regex.P1  rex field=payload /(?<digits>\d+)/
E.C24.slash_regex.P2  rex /(?<word>\w+)/
E.C24.slash_regex.N1  rex field=payload /(?<digits>\d+)
E.C24.slash_regex.N2  rex /(?<word>\w+)
```

### E.C24.sed

Syntax: mode=sed plus quoted s/regex/replacement/[g|N] or y/characters/characters/. Target: **A/U**. Source: [doc161](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/rex-command/rex-command-overview-syntax-and-usage). Base: C24-P2.

Separate s Nth occurrence positive plus allmax/offset interaction; body operations other than s/y not advertised. String wrapper recognition not full sed validation.

```text
E.C24.sed.P1  rex mode=sed "s/private/public/g"
E.C24.sed.P2  rex field=payload mode=sed "y/abc/ABC/"
E.C24.sed.N1  rex mode=sed
E.C24.sed.N2  rex mode=sed "y/abc/ABC/
```

### E.C25.input

Syntax: spath [input=field] without path autoextracts. Target: **A/U**. Source: [doc165](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/spath-command/spath-command-overview-syntax-and-usage). Base: C25-P1.

Default _raw has no fabricated source reference. Explicit input identifier is a read; autoextracted output names remain unknown.

```text
E.C25.input.P1  spath
E.C25.input.P2  spath input=payload
E.C25.input.N1  spath input=
E.C25.input.N2  spath input='payload
```

### E.C25.path

Syntax: path=double-quoted location string; default output=path name. Target: **A/U**. Source: [doc165](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/spath-command/spath-command-overview-syntax-and-usage). Base: C25.

Curly indexes/XML attribute path are literal path, not input event expression. Do not turn components into reads.

```text
E.C25.path.P1  spath path="actor.name"
E.C25.path.P2  spath input=payload path="rows{1}.name"
E.C25.path.N1  spath path=actor.name
E.C25.path.N2  spath path="actor.name
```

### E.C25.output

Syntax: output=field only with path=quoted-path. Target: **A/U**. Source: [doc165](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/spath-command/spath-command-overview-syntax-and-usage). Base: C25-P2,N1;L23-N2.

Output-without-path explicitly invalid context; original output token span retained.

```text
E.C25.output.P1  spath output=name path="actor.name"
E.C25.output.P2  spath input=payload output=first path="rows{0}.name"
E.C25.output.N1  spath output=name
E.C25.output.N2  spath input=payload output=first
```

### E.C26.count

Syntax: makeresults [positional integer]; omitted1; no count= option. Target: **A/U**. Source: [doc152](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/makeresults-command/makeresults-command-overview-syntax-and-usage). Base: C26.

Initial generator source state must reflect supported known fields only; optional-count omission positive.

```text
E.C26.count.P1  makeresults
E.C26.count.P2  makeresults 4 | eval label="local"
E.C26.count.N1  makeresults count=4
E.C26.count.N2  makeresults 2.5
```

### E.C27.sid

Syntax: loadjob sid, exactly one job identifier, generating first command. Target: **A/U**. Source: [doc149](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/loadjob-command/loadjob-command-overview-syntax-and-usage). Base: C27.

Job identifier dependency distinct from field/index; savedsearch loading/removed options explicitly excluded.

```text
E.C27.sid.P1  loadjob 1780123000.5
E.C27.sid.P2  loadjob 1780123000.5 | fields host
E.C27.sid.N1  loadjob
E.C27.sid.N2  FROM main | loadjob 1780123000.5
```

### E.C28.aggregates

Syntax: aggregates=[call (,call)*] required. Target: **A/U**. Source: [doc174](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/tstats-command/tstats-command-overview-syntax-usage). Base: C28.

Arity/context registry separate; metric field names with dot quoted for mstats.

```text
E.C28.aggregates.P1  tstats aggregates=[count()]
E.C28.aggregates.P2  tstats aggregates=[sum(bytes),max(bytes)]
E.C28.aggregates.N1  tstats aggregates=count()
E.C28.aggregates.N2  tstats aggregates=[count(),]
```

### E.C28.predicate

Syntax: predicate=(expression) optional. Target: **A/U**. Source: [doc174](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/tstats-command/tstats-command-overview-syntax-usage). Base: C28.

Source selector roles only in grammar-established positions; unproved metric resolution U.

```text
E.C28.predicate.P1  tstats aggregates=[count()] predicate=(index="metrics")
E.C28.predicate.P2  tstats aggregates=[sum(bytes)] predicate=(host="api" AND bytes>0)
E.C28.predicate.N1  tstats aggregates=[count()] predicate=index="metrics"
E.C28.predicate.N2  tstats aggregates=[count()] predicate=()
```

### E.C28.byfields

Syntax: byfields=[exact-field (,exact-field)*] optional, no wildcard. Target: **A/U**. Source: [doc174](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/tstats-command/tstats-command-overview-syntax-usage). Base: C28.

Optional omission accepted; no wildcard syntax fallback.

```text
E.C28.byfields.P1  tstats aggregates=[count()] byfields=[host]
E.C28.byfields.P2  tstats aggregates=[sum(bytes)] byfields=[host,service]
E.C28.byfields.N1  tstats aggregates=[count()] byfields=[host service]
E.C28.byfields.N2  tstats aggregates=[count()] byfields=['host*']
```

### E.C29.aggregates

Syntax: aggregates=[call (,call)*] required. Target: **A/U**. Source: [doc153](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/mstats-command/mstats-command-overview-syntax-and-usage). Base: C29.

Arity/context registry separate; metric field names with dot quoted for mstats.

```text
E.C29.aggregates.P1  mstats aggregates=[count()]
E.C29.aggregates.P2  mstats aggregates=[sum(bytes),max(bytes)]
E.C29.aggregates.N1  mstats aggregates=count()
E.C29.aggregates.N2  mstats aggregates=[count(),]
```

### E.C29.predicate

Syntax: predicate=(expression) optional. Target: **A/U**. Source: [doc153](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/mstats-command/mstats-command-overview-syntax-and-usage). Base: C29.

Source selector roles only in grammar-established positions; unproved metric resolution U.

```text
E.C29.predicate.P1  mstats aggregates=[count()] predicate=(index="metrics")
E.C29.predicate.P2  mstats aggregates=[sum(bytes)] predicate=(host="api" AND bytes>0)
E.C29.predicate.N1  mstats aggregates=[count()] predicate=index="metrics"
E.C29.predicate.N2  mstats aggregates=[count()] predicate=()
```

### E.C29.byfields

Syntax: byfields=[exact-field (,exact-field)*] optional, no wildcard. Target: **A/U**. Source: [doc153](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/mstats-command/mstats-command-overview-syntax-and-usage). Base: C29.

Optional omission accepted; no wildcard syntax fallback.

```text
E.C29.byfields.P1  mstats aggregates=[count()] byfields=[host]
E.C29.byfields.P2  mstats aggregates=[sum(bytes)] byfields=[host,service]
E.C29.byfields.N1  mstats aggregates=[count()] byfields=[host service]
E.C29.byfields.N2  mstats aggregates=[count()] byfields=['host*']
```

### E.C28.datamodel_name

Syntax: datamodel_name='Model.Root' optional. Target: **A/U**. Source: [doc174](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/tstats-command/tstats-command-overview-syntax-usage). Base: C28.

Model/root is dependency, not ordinary dotted event input. Subnode names in predicate; no catalog resolution.

```text
E.C28.datamodel_name.P1  tstats aggregates=[count()] datamodel_name='Traffic.All'
E.C28.datamodel_name.P2  tstats aggregates=[sum(bytes)] datamodel_name='Network.Flow' byfields=[host]
E.C28.datamodel_name.N1  tstats aggregates=[count()] datamodel_name=
E.C28.datamodel_name.N2  tstats aggregates=[count()] datamodel_name='Traffic.All
```

### E.C30.limit

Syntax: limit=int prefix; 0 disables series filtering. Target: **A/U**. Source: [doc172](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/timechart-command/timechart-command-overview-and-syntax). Base: C30.

Interacts with agg,useother; do not promise fixed series names.

```text
E.C30.limit.P1  timechart limit=0 count() BY host
E.C30.limit.P2  timechart limit=4 avg(bytes) BY service
E.C30.limit.N1  timechart limit= count() BY host
E.C30.limit.N2  timechart limit=2.5 count() BY host
```

### E.C30.agg

Syntax: agg=(aggregate [AS identifier]) prefix. Target: **A/U**. Source: [doc172](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/timechart-command/timechart-command-overview-and-syntax). Base: C30.

Bare agg=sum is separately evidenced conflicting syntax (EH02), not a malformed negative.

```text
E.C30.agg.P1  timechart agg=(sum(bytes)) limit=5 avg(bytes) BY host
E.C30.agg.P2  timechart agg=(max(duration) AS peak) count() BY service
E.C30.agg.N1  timechart agg=(sum(bytes) limit=5 avg(bytes) BY host
E.C30.agg.N2  timechart agg=(max(duration) AS) count() BY service
```

### E.C30.eval_expression

Syntax: eval(expression involving aggregates) BY split field. Target: **A/U**. Source: [timechartexamples](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/timechart-command/timechart-command-examples). Base: C30.

A fieldless ordinary aggregate count() need not BY; expression form does. Raw expression/extra outer parentheses do not inherit proof automatically.

```text
E.C30.eval_expression.P1  timechart eval(avg(bytes)*avg(duration)) BY host
E.C30.eval_expression.P2  timechart eval(round(avg(duration),2)) BY service
E.C30.eval_expression.N1  timechart eval(avg(bytes)*avg(duration))
E.C30.eval_expression.N2  timechart eval(round(avg(duration),2)) BY
```

### E.C30.single_aggregate

Syntax: count() or one aggregate(field), optional BY one split field. Target: **A/U**. Source: [doc172](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/timechart-command/timechart-command-overview-and-syntax). Base: C30.

Specific command forbids wildcard single aggregate fields; same field used as aggregate input and split is documented runtime restriction, not arbitrary parse failure. H04 multiple aggregates held.

```text
E.C30.single_aggregate.P1  timechart sum(bytes)
E.C30.single_aggregate.P2  timechart count() BY host
E.C30.single_aggregate.N1  timechart sum()
E.C30.single_aggregate.N2  timechart count() BY host,service
```

### E.C30.bins

Syntax: bins=int prefix time-axis OR after BY split field. Target: **A/U**. Source: [doc172](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/timechart-command/timechart-command-overview-and-syntax). Base: C30.

Distinct placement means distinct axis, not forbidden post-BY options. bins+span allowed; span wins. Model remains U.

```text
E.C30.bins.P1  timechart bins=25 count()
E.C30.bins.P2  timechart avg(bytes) BY bucket bins=100
E.C30.bins.N1  timechart bins= count()
E.C30.bins.N2  timechart avg(bytes) BY bucket bins=
```

### E.C30.minspan

Syntax: minspan=span-length prefix time-axis OR after BY split field. Target: **A/U**. Source: [doc172](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/timechart-command/timechart-command-overview-and-syntax). Base: C30.

Distinct placement means distinct axis, not forbidden post-BY options. bins+span allowed; span wins. Model remains U.

```text
E.C30.minspan.P1  timechart minspan=5min count()
E.C30.minspan.P2  timechart avg(bytes) BY bucket minspan=10ms
E.C30.minspan.N1  timechart minspan= count()
E.C30.minspan.N2  timechart avg(bytes) BY bucket minspan=
```

### E.C30.span

Syntax: span=span-length|logspan|weekly-snap-span prefix time-axis OR after BY split field. Target: **A/U**. Source: [doc172](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/timechart-command/timechart-command-overview-and-syntax). Base: C30.

Distinct placement means distinct axis, not forbidden post-BY options. bins+span allowed; span wins. Model remains U.

```text
E.C30.span.P1  timechart span=10min count()
E.C30.span.P2  timechart avg(bytes) BY bucket span=log2
E.C30.span.N1  timechart span= count()
E.C30.span.N2  timechart avg(bytes) BY bucket span=
```

### E.C30.start

Syntax: start=num prefix time-axis OR after BY split field. Target: **A/U**. Source: [doc172](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/timechart-command/timechart-command-overview-and-syntax). Base: C30.

Distinct placement means distinct axis, not forbidden post-BY options. bins+span allowed; span wins. Model remains U.

```text
E.C30.start.P1  timechart start=0 count()
E.C30.start.P2  timechart avg(bytes) BY bucket start=-5
E.C30.start.N1  timechart start= count()
E.C30.start.N2  timechart avg(bytes) BY bucket start=
```

### E.C30.end

Syntax: end=num prefix time-axis OR after BY split field. Target: **A/U**. Source: [doc172](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/timechart-command/timechart-command-overview-and-syntax). Base: C30.

Distinct placement means distinct axis, not forbidden post-BY options. bins+span allowed; span wins. Model remains U.

```text
E.C30.end.P1  timechart end=50 count()
E.C30.end.P2  timechart avg(bytes) BY bucket end=1000
E.C30.end.N1  timechart end= count()
E.C30.end.N2  timechart avg(bytes) BY bucket end=
```

### E.C30.aligntime

Syntax: aligntime=earliest|latest|time-specifier prefix time-axis OR after BY split field. Target: **A/U**. Source: [doc172](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/timechart-command/timechart-command-overview-and-syntax). Base: C30.

Distinct placement means distinct axis, not forbidden post-BY options. bins+span allowed; span wins. Model remains U.

```text
E.C30.aligntime.P1  timechart aligntime=earliest count()
E.C30.aligntime.P2  timechart avg(bytes) BY bucket aligntime=@d+2h
E.C30.aligntime.N1  timechart aligntime= count()
E.C30.aligntime.N2  timechart avg(bytes) BY bucket aligntime=
```

### E.C30.weekly_snap

Syntax: span=[+|-][integer]week-unit@snap-unit, weekly relative unit only. Target: **A/U**. Source: [doc172](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/timechart-command/timechart-command-overview-and-syntax). Base: C30.

Nonweekly snap spans are not admitted; H source gap versus definite native restriction must be kept explicit.

```text
E.C30.weekly_snap.P1  timechart span=w@w1 count()
E.C30.weekly_snap.P2  timechart span=2w@w0 count() BY host
E.C30.weekly_snap.N1  timechart span=w@ count()
E.C30.weekly_snap.N2  timechart span=2w@ count() BY host
```

### E.C31.span

Syntax: timewrap [integer]time-unit, required unit; preceding timechart. Target: **A/U**. Source: [doc173](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/timewrap-command/timewrap-command-overview-syntax-and-usage). Base: C31.

m is month; min is minute. Separately cover quarter/year unit classes, not just spelling variants.

```text
E.C31.span.P1  FROM main | timechart count() | timewrap day
E.C31.span.P2  FROM main | timechart count() | timewrap 2m
E.C31.span.N1  FROM main | timechart count() | timewrap
E.C31.span.N2  FROM main | timechart count() | timewrap 2
```

### E.C31.align

Syntax: align=now|end after required timewrap span. Target: **A/U**. Source: [doc173](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/timewrap-command/timewrap-command-overview-syntax-and-usage). Base: C31.

Commandspecific option-after-span; earlier timechart precondition is structural context.

```text
E.C31.align.P1  FROM main | timechart count() | timewrap week align=now
E.C31.align.P2  FROM main | timechart count() | timewrap 2day align=end
E.C31.align.N1  FROM main | timechart count() | timewrap week align=start
E.C31.align.N2  FROM main | timechart count() | timewrap week align=
```

### E.C32.delim

Syntax: makemv [delim=string] field, before one noninternal field. Target: **A/U**. Source: [doc151](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/makemv-command/makemv-command-overview-syntax-and-usage). Base: C32.

allowempty/setsv removed; both delim+tokenizer syntactic exclusivity unproven, no false negative.

```text
E.C32.delim.P1  makemv delim=":" labels
E.C32.delim.P2  makemv delim="::" 'tag list'
E.C32.delim.N1  makemv delim=":"
E.C32.delim.N2  makemv delim="::" labels,other
```

### E.C32.tokenizer

Syntax: makemv [tokenizer=regex-string] field. Target: **A/U**. Source: [doc151](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/makemv-command/makemv-command-overview-syntax-and-usage). Base: C32.

Raw-string acceptance follows global string-expression syntax; regex matching/group semantics remain U.

```text
E.C32.tokenizer.P1  makemv tokenizer="([a-z]+)" labels
E.C32.tokenizer.P2  makemv tokenizer=@"(\d+)" ids
E.C32.tokenizer.N1  makemv tokenizer="([a-z]+)"
E.C32.tokenizer.N2  makemv tokenizer="([a-z]+)
```

### E.C33.limit

Syntax: mvexpand [limit=int] field; option before exactly one field. Target: **A/U**. Source: [doc155](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/mvexpand-command/mvexpand-command-overview-syntax-and-usage). Base: C33.

0 unlimited versus bounded expansion; no implicit availability promotion.

```text
E.C33.limit.P1  mvexpand limit=0 labels
E.C33.limit.P2  mvexpand limit=3 'tag list'
E.C33.limit.N1  mvexpand labels limit=3
E.C33.limit.N2  mvexpand limit=2.5 labels
```

### E.C34.delim

Syntax: mvcombine [delim=string] field; one noninternal field. Target: **A/U**. Source: [doc154](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/mvcombine-command/mvcombine-command-overview-syntax-and-usage). Base: C34.

No internal fields; row combination effect still U.

```text
E.C34.delim.P1  mvcombine delim=":" labels
E.C34.delim.P2  mvcombine delim=" / " 'tag list'
E.C34.delim.N1  mvcombine delim=":"
E.C34.delim.N2  mvcombine delim=" / " labels,other
```

### E.C35.value

Syntax: fillnull [value=double-quoted-string] [comma fields]. Target: **A/U**. Source: [doc141](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/fillnull-command/fillnull-command-overview-syntax-and-usage). Base: C35.

Default0 omitted positive; explicit numeric unquoted is not string. Creation conditional/value-dependent.

```text
E.C35.value.P1  fillnull value="missing"
E.C35.value.P2  fillnull value="0" user,host
E.C35.value.N1  fillnull value=0 user
E.C35.value.N2  fillnull value="missing
```

### E.C35.fields

Syntax: optional nonempty comma-separated field list after optional value. Target: **A/U**. Source: [doc141](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/fillnull-command/fillnull-command-overview-syntax-and-usage). Base: C35.

No list means all applicable fields, unknown names/effects; don't classify that omission as malformed.

```text
E.C35.fields.P1  fillnull user
E.C35.fields.P2  fillnull value="unknown" user,host
E.C35.fields.N1  fillnull user host
E.C35.fields.N2  fillnull user,
```

### E.L01.start

Syntax: Generating standalone source, explicit search, or documented index-first implicit search. Target: **A/K selected command**. Source: [start](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-manual/getting-started/quick-start-write-and-run-a-basic-spl2-search/start-searching-data-using-spl2). Base: L01;C01;C02.

Existing starts preserved, no broad missing initial command inference.

```text
E.L01.start.P1  index=app failure
E.L01.start.P2  SELECT host FROM app
E.L01.start.N1  failure index=app
E.L01.start.N2  SELECT host
```

### E.L02.identifier

Syntax: Single-quoted reserved/Unicode/punctuation identifiers; double strings distinct. Target: **A/K exact**. Source: [syntax](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/introduction/understanding-spl2-syntax). Base: L02.

Assert semantic decoded name and exact original multibyte source slice; aliases may have narrower restrictions.

```text
E.L02.identifier.P1  eval 'résumé'=user,'literal.path'='résumé'
E.L02.identifier.P2  fields 'group','user-id'
E.L02.identifier.N1  eval 'résumé=user
E.L02.identifier.N2  fields group
```

### E.L03.number

Syntax: int, long L, double decimals/exponents/D, float F; unary sign distinct from binary operator. Target: **A/K reads/no evaluation**. Source: [types](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-manual/data-types/built-in-data-types). Base: L03.

Do not reject number overflow or invalid operation types by grammar guessing. One sample per suffix and exponent lexical form plus source offsets.

```text
E.L03.number.P1  eval a=42L,b=1.25F,c=8D
E.L03.number.P2  eval x=2.5e-3,y=-4E2F
E.L03.number.N1  eval x=2.5e-
E.L03.number.N2  eval x=4+
```

### E.L03.raw

Syntax: @"raw text" with doubled embedded double quotation mark. Target: **A/K literal**. Source: [expressions](https://help.splunk.com/en/splunk-cloud-platform/search/spl2-search-manual/expressions-and-predicates/types-of-expressions). Base: L03.

No template/event field reads from literal raw body unless explicitly documented; escaped backslashes not processed as normal-string escapes.

```text
E.L03.raw.P1  eval path=@"C:\data\logs"
E.L03.raw.P2  eval quote=@"She said ""yes"""
E.L03.raw.N1  eval path=@"C:\data\logs
E.L03.raw.N2  eval quote=@"She said ""yes""
```

### E.L03.null_boolean

Syntax: lowercase true/false/null; null assignment removes field. Target: **A/K conditional**. Source: [types](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-manual/data-types/built-in-data-types). Base: L03.

Do not make `eval x=TRUE` a Boolean-literal syntax negative: it may be identifier syntax. R01 Boolean eval allowed.

```text
E.L03.null_boolean.P1  eval enabled=true,retired=null
E.L03.null_boolean.P2  eval enabled=false,copy=enabled
E.L03.null_boolean.N1  head keeplast=TRUE 2
E.L03.null_boolean.N2  dedup keepempty=1 host
```

### E.L04.comments

Syntax: // line and /*block*/ outside search portions; preserve newlines and strings. Target: **A/K**. Source: [comments](https://help.splunk.com/en/splunk-cloud-platform/search/spl2-search-manual/comments/using-comments-in-spl2). Base: L04.

Actual newline variant plus CRLF duplicate-for-location assertion, not a new semantic-form count.

```text
E.L04.comments.P1  FROM main /* | inert */ | eval note="//text"
E.L04.comments.P2  FROM main // | inert
| fields host
E.L04.comments.N1  FROM main /* open
E.L04.comments.N2  search index=main /* disallowed */ host=api
```

### E.L06.precedence

Syntax: Unary +/-, * / %, + -; + concatenation; dot navigation. Target: **A/K supported operands**. Source: [expressions](https://help.splunk.com/en/splunk-cloud-platform/search/spl2-search-manual/expressions-and-predicates/types-of-expressions). Base: L06.

Tree grouping required; parentheses/text pipes inert; no global SPL concatenation rewrite.

```text
E.L06.precedence.P1  eval x=-(a+b)*c%2
E.L06.precedence.P2  eval x=a/b-c, label=user+":"+host
E.L06.precedence.N1  eval x=-(a+b)*
E.L06.precedence.N2  eval label=user+
```

### E.L07.between

Syntax: expr [NOT] BETWEEN low AND high. Target: **A/K structural reads**. Source: [predicate](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-manual/expressions-and-predicates/predicate-expressions). Base: L07.

Operator-specific casing per source; H11 remains held for unreviewed mixed forms. Type-test names are syntax, not source fields.

```text
E.L07.between.P1  where amount BETWEEN 0 AND 5
E.L07.between.P2  where amount NOT BETWEEN -2 AND 0
E.L07.between.N1  where amount BETWEEN 0
E.L07.between.N2  where amount NOT BETWEEN AND 0
```

### E.L07.in

Syntax: expr [NOT] IN (comma expression list). Target: **A/K structural reads**. Source: [predicate](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-manual/expressions-and-predicates/predicate-expressions). Base: L07.

Operator-specific casing per source; H11 remains held for unreviewed mixed forms. Type-test names are syntax, not source fields.

```text
E.L07.in.P1  where code IN (200,201)
E.L07.in.P2  where region NOT IN (primary,backup)
E.L07.in.N1  where code IN (200 201)
E.L07.in.N2  where region NOT IN (primary,)
```

### E.L07.like

Syntax: expr [NOT] LIKE pattern. Target: **A/K structural reads**. Source: [predicate](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-manual/expressions-and-predicates/predicate-expressions). Base: L07.

Operator-specific casing per source; H11 remains held for unreviewed mixed forms. Type-test names are syntax, not source fields.

```text
E.L07.like.P1  where user LIKE "admin%"
E.L07.like.P2  where user NOT LIKE "svc_"
E.L07.like.N1  where user LIKE
E.L07.like.N2  where user NOT LIKE
```

### E.L07.is_null

Syntax: expr IS [NOT] NULL. Target: **A/K structural reads**. Source: [predicate](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-manual/expressions-and-predicates/predicate-expressions). Base: L07.

Operator-specific casing per source; H11 remains held for unreviewed mixed forms. Type-test names are syntax, not source fields.

```text
E.L07.is_null.P1  where user IS NULL
E.L07.is_null.P2  where user IS NOT NULL
E.L07.is_null.N1  where user IS NOT
E.L07.is_null.N2  where IS NULL
```

### E.L07.is_type

Syntax: expr IS built-in-type; NOT (expr IS built-in-type) for negation. Target: **A/K structural reads**. Source: [predicate](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-manual/expressions-and-predicates/predicate-expressions). Base: L07.

Operator-specific casing per source; H11 remains held for unreviewed mixed forms. Type-test names are syntax, not source fields. IS NOT type is not proven by this source; keep incomplete, distinct from documented IS NOT NULL.

```text
E.L07.is_type.P1  where amount IS int
E.L07.is_type.P2  where NOT (user IS string)
E.L07.is_type.N1  where amount IS
E.L07.is_type.N2  where NOT (user IS)
```

### E.L08.array

Syntax: [comma expressions], nested arrays/literals; middle missing elements invalid. Target: **A/K static or U dynamic**. Source: [objects](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-manual/expressions-and-predicates/array-and-object-literals-in-expressions). Base: L08.

H05 trailing array comma unchanged; [] empty array needs built-in-types evidence before treating as malformed or source dataset.

```text
E.L08.array.P1  eval pair=[host,1]
E.L08.array.P2  eval nested=[[1,2],bytes+1]
E.L08.array.N1  eval pair=[host,,1]
E.L08.array.N2  eval nested=[[1,2],bytes+1
```

### E.L08.object_keys

Syntax: {key:expression (, key:expression)* [,]}; bare/single/double quoted keys. Target: **A/K static**. Source: [objects](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-manual/expressions-and-predicates/array-and-object-literals-in-expressions). Base: L08.

Second negative duplicate decoded key has semantic contract error even if grammar parses; key names not source reads.

```text
E.L08.object_keys.P1  eval obj={owner:user,bytes:bytes,}
E.L08.object_keys.P2  eval obj={'display name':user,"role-name":role}
E.L08.object_keys.N1  eval obj={owner user}
E.L08.object_keys.N2  eval obj={'display name':user,"display name":role}
```

### E.L08.duplicate_keys

Syntax: Object key names unique after supported quote decoding. Target: **A/K static; invalid duplicate**. Source: [objects](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-manual/expressions-and-predicates/array-and-object-literals-in-expressions). Base: L08.

Do not merge duplicates silently; escaped-equivalent keys need follow-up decoding assertions. Duplicate nesting scoped per object.

```text
E.L08.duplicate_keys.P1  eval obj={a:1,b:2}
E.L08.duplicate_keys.P2  eval obj={name:user,'other-name':host}
E.L08.duplicate_keys.N1  eval obj={a:1,a:2}
E.L08.duplicate_keys.N2  eval obj={name:user,"name":host}
```

### E.L09.dot

Syntax: base.member, including single-quoted reserved member, chained. Target: **A/K path per M4 or U**. Source: [access](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-manual/expressions-and-predicates/access-expressions-for-arrays-and-objects). Base: L09.

Literal 'actor.user.name' remains different from path; alias-qualified SQL read is a third role.

```text
E.L09.dot.P1  eval name=actor.user.name
E.L09.dot.P2  eval name=actor.'group'.owner
E.L09.dot.N1  eval name=actor.
E.L09.dot.N2  eval name=actor.'group
```

### E.L09.bracket

Syntax: base[index-expression] with integer/string/computed indexes. Target: **A/K static or U dynamic**. Source: [access](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-manual/expressions-and-predicates/access-expressions-for-arrays-and-objects). Base: L09.

Computed index retains base and index expression reads; quoted static key has no field read.

```text
E.L09.bracket.P1  eval x=rows[0]["name"]
E.L09.bracket.P2  eval x=rows[position+1].value
E.L09.bracket.N1  eval x=rows[]
E.L09.bracket.N2  eval x=rows[position+1
```

### E.L10.string_template

Syntax: Double-quoted ${expression} interpolation, nested expression quote contexts. Target: **A/K supported interpolation**. Source: [templates](https://help.splunk.com/en/splunk-cloud-platform/search/spl2-search-manual/expressions-and-predicates/string-templates-in-expressions). Base: L10-P1,N1.

Known expression reads retained; expression strings containing braces/quotes need exact location assertions, no lexer whole-string opacity.

```text
E.L10.string_template.P1  eval label="host=${host}; size=${bytes}"
E.L10.string_template.P2  eval label="delta=${abs(high-low)}"
E.L10.string_template.N1  eval label="host=${}"
E.L10.string_template.N2  eval label="host=${host"
```

### E.L10.field_template

Syntax: Single-quoted ${expression} computes field name. Target: **A/U dynamic target**. Source: [fieldtemplates](https://help.splunk.com/en/splunk-cloud-platform/search/spl2-search-manual/expressions-and-predicates/field-templates-in-expressions). Base: L10-P2,N2.

Nested single quotes are documented lexer-mode syntax, not early closing delimiter; target is computed, not exact created name.

```text
E.L10.field_template.P1  eval '${kind}'=value
E.L10.field_template.P2  eval '${'résumé'}-value'=bytes
E.L10.field_template.N1  eval '${}'=value
E.L10.field_template.N2  eval '${kind'=value
```

### E.L11.one_param

Syntax: $param -> expr or ($param) -> expr; parameter begins $letter/_. Target: **A/U**. Source: [lambda](https://help.splunk.com/en/splunk-cloud-platform/search/spl2-search-manual/expressions-and-predicates/lambda-expressions). Base: L11.

Lambda local x excluded; free factor retained; callback effect U.

```text
E.L11.one_param.P1  eval result=map(items,$x -> $x+1)
E.L11.one_param.P2  eval result=map(items,($x) -> $x+factor)
E.L11.one_param.N1  eval result=map(items,$x $x+1)
E.L11.one_param.N2  eval result=map(items,($x) ->)
```

### E.L11.zero_multi

Syntax: () or ($a,$b,...) mandatory parentheses for zero/multi parameters. Target: **A/U**. Source: [lambda](https://help.splunk.com/en/splunk-cloud-platform/search/spl2-search-manual/expressions-and-predicates/lambda-expressions). Base: L11.

Callback invocation arity/type not simulated. Zero-param syntax documented even if function callback semantics unknown.

```text
E.L11.zero_multi.P1  eval result=map(items,() -> 1)
E.L11.zero_multi.P2  eval result=reduce(items,0,($a,$b) -> $a+$b)
E.L11.zero_multi.N1  eval result=map(items,-> 1)
E.L11.zero_multi.N2  eval result=reduce(items,0,($a $b) -> $a+$b)
```

### E.L11.typed_default

Syntax: ($p[:built-in-type][=number|string constant],...) -> body. Target: **A/U**. Source: [lambda](https://help.splunk.com/en/splunk-cloud-platform/search/spl2-search-manual/expressions-and-predicates/lambda-expressions). Base: L11.

Default field-name invalid declaration contract, not source read; no top-level function support implied.

```text
E.L11.typed_default.P1  eval result=map(items,($x:int=2) -> $x+1)
E.L11.typed_default.P2  eval result=map(items,($x:string="none") -> $x)
E.L11.typed_default.N1  eval result=map(items,($x int=2) -> $x)
E.L11.typed_default.N2  eval result=map(items,($x:int=fallback) -> $x)
```

### E.L11.block

Syntax: { $local=expr; ... return expr } or newline separators; one final return. Target: **A/U**. Source: [lambda](https://help.splunk.com/en/splunk-cloud-platform/search/spl2-search-manual/expressions-and-predicates/lambda-expressions). Base: L11-P2,N2.

No nested lambda; per-parameter/local references scoped, ordinary free body reads preserved.

```text
E.L11.block.P1  eval result=map(items,$x -> {$y=$x*2; return $y})
E.L11.block.P2  eval result=map(items,$x -> {
$y=$x+1
return $y
})
E.L11.block.N1  eval result=map(items,$x -> {$y=$x*2})
E.L11.block.N2  eval result=map(items,$x -> {return $x; $y=2})
```

### E.L11.shortcut

Syntax: Contextual implicit $it expression as lambda argument. Target: **A/U**. Source: [lambda](https://help.splunk.com/en/splunk-cloud-platform/search/spl2-search-manual/expressions-and-predicates/lambda-expressions). Base: L11.

Do not treat any $identifier as module declaration; implicit local $it only lambda-accepting call context evidenced. Unknown call U.

```text
E.L11.shortcut.P1  eval result=map(items,$it+2)
E.L11.shortcut.P2  eval result=filter(items,$it.amount>0)
E.L11.shortcut.N1  eval result=map(items,$it+)
E.L11.shortcut.N2  eval result=filter(items,$it.amount>)
```

### E.L15.from_clauses

Syntax: FROM [JOIN] [WHERE] [GROUP BY] [SELECT] [HAVING] [ORDER BY] [LIMIT] [OFFSET] hierarchy. Target: **A/K nojoin subset**. Source: [fromusage](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/from-command/from-command-usage). Base: C02;L15;Q01.

Original lexical IDs and logical phase order, SELECT finalization after HAVING not source rewrite.

```text
E.L15.from_clauses.P1  FROM main WHERE bytes>0 SELECT user ORDER BY user LIMIT 3 OFFSET 1
E.L15.from_clauses.P2  FROM main GROUP BY host SELECT host,count() HAVING count>1
E.L15.from_clauses.N1  FROM main SELECT user WHERE bytes>0
E.L15.from_clauses.N2  FROM main GROUP BY host
```

### E.L16.select_clauses

Syntax: SELECT [DISTINCT] expressions FROM [JOIN] [WHERE] [GROUP BY] [HAVING] [ORDER BY] [LIMIT] [OFFSET]. Target: **A/K nojoin subset**. Source: [fromusage](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/from-command/from-command-usage). Base: L16.

Aliases cannot suppress original WHERE requirements; Q01/Q06 environment constraints preserved.

```text
E.L16.select_clauses.P1  SELECT user FROM main WHERE bytes>0 ORDER BY user
E.L16.select_clauses.P2  SELECT host,count() FROM main GROUP BY host HAVING count>1 LIMIT 5
E.L16.select_clauses.N1  SELECT user WHERE bytes>0 FROM main
E.L16.select_clauses.N2  SELECT host,count() FROM main HAVING count>1 GROUP BY host
```

### E.L16.distinct

Syntax: SELECT DISTINCT expression (,expression)*. Target: **A/K**. Source: [fromsyntax](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/from-command/from-command-syntax). Base: Q02.

Distinct affects rows, not field creation; expression aliases followed in source span.

```text
E.L16.distinct.P1  SELECT DISTINCT host FROM main
E.L16.distinct.P2  FROM main SELECT DISTINCT host,user
E.L16.distinct.N1  SELECT DISTINCT FROM main
E.L16.distinct.N2  FROM main SELECT host DISTINCT
```

### E.L16.alias

Syntax: expression [AS identifier] in SELECT; explicit AS source alias. Target: **A/K explicit known name, qualified-path U if unproved**. Source: [fromsyntax](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/from-command/from-command-syntax). Base: L16.

AS= not named-argument syntax; module names/resources separate. Unknown implicit expression output stays U.

```text
E.L16.alias.P1  SELECT lower(user) AS normalized FROM main
E.L16.alias.P2  FROM main AS events SELECT events.user AS owner
E.L16.alias.N1  SELECT lower(user) AS FROM main
E.L16.alias.N2  FROM main AS SELECT user
```

### E.L16.group_keys

Syntax: GROUP BY/GROUPBY comma expressions; requires SELECT, no wildcard key. Target: **A/K exact; general expression U**. Source: [fromsyntax](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/from-command/from-command-syntax). Base: L16.

H08 invalidity uncertainty for nongrouped expressions unchanged; expression grouping remains grammar A/effect U.

```text
E.L16.group_keys.P1  SELECT host,user,count() FROM main GROUP BY host,user
E.L16.group_keys.P2  FROM main GROUPBY lower(user) SELECT lower(user) AS normalized,count()
E.L16.group_keys.N1  SELECT count() FROM main GROUP BY 'host*'
E.L16.group_keys.N2  FROM main GROUP BY host user SELECT count()
```

### E.L16.group_span

Syntax: span(field), span(field,optional-int unit), field span=(optional-int unit). Target: **A/U span effect**. Source: [fromsyntax](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/from-command/from-command-syntax). Base: L16.

Combine with Q05 two-arg and unit-only positives; default span output name unproved. Do not require fabricated _time availability.

```text
E.L16.group_span.P1  SELECT count() FROM main GROUP BY span(_time)
E.L16.group_span.P2  FROM main GROUP BY _time span=(5min) SELECT count()
E.L16.group_span.N1  SELECT count() FROM main GROUP BY span()
E.L16.group_span.N2  FROM main GROUP BY _time span=() SELECT count()
```

### E.L16.having

Syntax: HAVING predicate after SELECT/group phase. Target: **A/K selected names, U hidden visibility**. Source: [fromsyntax](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/from-command/from-command-syntax). Base: L16.

Q06 hidden grouping/order visibility stays incomplete; grouping evidence survives until final SELECT projection.

```text
E.L16.having.P1  SELECT count() AS n FROM main HAVING n>2
E.L16.having.P2  FROM main GROUP BY host SELECT host,sum(bytes) AS total HAVING total>0
E.L16.having.N1  SELECT count() AS n FROM main HAVING
E.L16.having.N2  FROM main GROUP BY host SELECT host,sum(bytes) AS total HAVING total>
```

### E.L16.order

Syntax: ORDER BY/ORDERBY comma expressions [ASC|DESC]. Target: **A/K known selected fields; U hidden visibility**. Source: [fromsyntax](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/from-command/from-command-syntax). Base: L16.

H07 independently signed per-term directions held. ORDER without SELECT keeps open source state.

```text
E.L16.order.P1  SELECT host,user FROM main ORDER BY host,user DESC
E.L16.order.P2  FROM main ORDERBY bytes ASC
E.L16.order.N1  SELECT host,user FROM main ORDER BY host,
E.L16.order.N2  FROM main ORDERBY
```

### E.L16.limit_offset

Syntax: LIMIT integer then OFFSET integer, each optional. Target: **A/K**. Source: [fromsyntax](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/from-command/from-command-syntax). Base: L16.

H06 negative values held; do not assume LIMIT required when OFFSET appears. Q03 validates relative clause order.

```text
E.L16.limit_offset.P1  SELECT host FROM main LIMIT 4
E.L16.limit_offset.P2  FROM main OFFSET 2
E.L16.limit_offset.N1  SELECT host FROM main LIMIT
E.L16.limit_offset.N2  FROM main LIMIT 4 OFFSET
```

### E.L17.pipe

Syntax: Completed SQL unit may feed supported command, cannot resume SQL clauses after pipe. Target: **A/K or U per downstream**. Source: [process](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-manual/getting-started/quick-start-write-and-run-a-basic-spl2-search/process-your-search-results). Base: L17.

SELECT-owned final projection before downstream stage, correct source positions.

```text
E.L17.pipe.P1  SELECT user FROM main | eval owner=user
E.L17.pipe.P2  FROM main GROUP BY host SELECT host,count() | where count>2
E.L17.pipe.N1  SELECT user FROM main | eval owner=user GROUP BY host
E.L17.pipe.N2  FROM main SELECT user | ORDER BY user
```

### E.L18.rows

Syntax: FROM array of object rows or SELECT...FROM literal rows. Target: **A/K local finite/conditional**. Source: [fromusage](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/from-command/from-command-usage). Base: L18.

Missing keys conditional; [{}] empty object row evidenced objects page; [] empty dataset remains H09.

```text
E.L18.rows.P1  FROM [{a:1},{a:2}] SELECT a
E.L18.rows.P2  SELECT a FROM [{a:1},{b:2}]
E.L18.rows.N1  FROM [{a:1}{a:2}] SELECT a
E.L18.rows.N2  SELECT a FROM [{a:1],{a:2}]
```

### E.L19.inner

Syntax: FROM source AS a [INNER] JOIN source AS b ON a.key=b.key [AND ...]. Target: **A/U merge**. Source: [fromsyntax](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/from-command/from-command-syntax). Base: L19.

Join aliases retain namespace role but do not become root dotted fields.

```text
E.L19.inner.P1  FROM main AS m JOIN users AS u ON m.uid=u.id SELECT m.uid
E.L19.inner.P2  SELECT m.uid FROM main AS m INNER JOIN users AS u ON m.uid=u.id AND m.realm=u.realm
E.L19.inner.N1  FROM main JOIN users AS u ON m.uid=u.id SELECT m.uid
E.L19.inner.N2  SELECT m.uid FROM main AS m JOIN users AS u m.uid=u.id
```

### E.L19.left

Syntax: LEFT JOIN or LEFT OUTER JOIN, aliases/ON mandatory. Target: **A/U merge**. Source: [fromsyntax](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/from-command/from-command-syntax). Base: L19.

No RIGHT/FULL admission; many-to-one assumptions forbidden.

```text
E.L19.left.P1  FROM main AS m LEFT JOIN users AS u ON m.uid=u.id SELECT m.uid
E.L19.left.P2  SELECT m.uid FROM main AS m LEFT OUTER JOIN users AS u ON m.uid=u.id
E.L19.left.N1  FROM main AS m LEFT JOIN users ON m.uid=u.id SELECT m.uid
E.L19.left.N2  SELECT m.uid FROM main AS m FULL JOIN users AS u ON m.uid=u.id
```

### E.L19.multiple_join

Syntax: JOIN clauses repeat before WHERE/GROUP/SELECT. Target: **A/U merge**. Source: [fromsyntax](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/from-command/from-command-syntax). Base: L19.

Each child independent source scope; copied parent scopes not shared mutation.

```text
E.L19.multiple_join.P1  FROM main AS m JOIN users AS u ON m.uid=u.id JOIN teams AS t ON u.team=t.id SELECT m.uid
E.L19.multiple_join.P2  SELECT m.uid FROM main AS m LEFT JOIN users AS u ON m.uid=u.id LEFT JOIN teams AS t ON u.team=t.id
E.L19.multiple_join.N1  FROM main AS m JOIN users AS u ON m.uid=u.id JOIN SELECT m.uid
E.L19.multiple_join.N2  SELECT m.uid FROM main AS m JOIN users AS u ON m.uid=u.id WHERE m.uid>0 JOIN teams AS t ON u.team=t.id
```

### E.L07.exists

Syntax: [NOT] EXISTS(parenthesized SQL subsearch), aliased outer, equality correlation in child WHERE. Target: **A/U correlation**. Source: [predicate](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-manual/expressions-and-predicates/predicate-expressions). Base: Q07.

Child OFFSET also forbidden; no pipe where; if outer WHERE+HAVING, EXISTS belongs HAVING. Q07 covers context.

```text
E.L07.exists.P1  FROM main AS m WHERE EXISTS (SELECT id FROM users WHERE id=m.uid) SELECT uid
E.L07.exists.P2  SELECT uid FROM main AS m WHERE NOT EXISTS (SELECT id FROM users WHERE id=m.uid)
E.L07.exists.N1  FROM main WHERE EXISTS (SELECT id FROM users WHERE id=m.uid) SELECT uid
E.L07.exists.N2  SELECT uid FROM main AS m WHERE EXISTS (SELECT id FROM users WHERE id=m.uid LIMIT 1)
```

### E.L24.search_literal

Syntax: Expression-position backtick search predicate; implied AND inside body. Target: **A/U body until modeled**. Source: [searchliterals](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-manual/expressions-and-predicates/search-literals-in-expressions). Base: L24.

Also positive stats count(`failed`) BY host and pipe where. Tokens inside are search grammar, not macro identifiers or arbitrary SPL subpipeline. Blank/body restrictions require grammar evidence.

```text
E.L24.search_literal.P1  FROM main WHERE `502 failure` SELECT host
E.L24.search_literal.P2  FROM main | eval class=if(`status=4*`,"client","other")
E.L24.search_literal.N1  FROM main WHERE `502 failure SELECT host
E.L24.search_literal.N2  FROM main | eval class=if(`status=4*,"client","other")
```

### E.L06.named_arguments

Syntax: Built-in name:value; allnamed reorder; positional prefix before named suffix; list value array. Target: **A/U by explicit root ruling**. Source: [namingargs](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-manual/functions/naming-function-arguments). Base: L06.

Additional positives round(bytes,precision:2), if(false_value:0,predicate:code=200,true_value:1). Labels not fields; values retain reads. '=' can be legitimate positional comparison, not guessed named syntax.

```text
E.L06.named_arguments.P1  FROM main | eval x=round(precision:2,num:bytes)
E.L06.named_arguments.P2  FROM main | eval x=coalesce(values:[primary,backup])
E.L06.named_arguments.N1  FROM main | eval x=round(num:bytes,2)
E.L06.named_arguments.N2  FROM main | eval x=coalesce(values:primary,backup)
```

## Finite approved positional function signatures

All following spellings are approved positional scope, not a complete Splunk catalog. Profile applicability comes from [evalcompat](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/spl2-compatibility-profiles-and-quick-references/compatibility-quick-reference-for-spl2-evaluation-functions) / [aggcompat](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/spl2-compatibility-profiles-and-quick-references/compatibility-quick-reference-for-spl2-statistical-functions); individual pages below establish signatures. [statsintro](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/statistical-and-charting-functions/overview-of-spl2-stats-and-chart-functions) establishes aggregate applicability to stats, eventstats, streamstats and timechart without treating an omission in an individual page as a prohibition. Command grammar/effect boundaries still apply.

Optional arities each receive two positives. The shared too-few/too-many negatives bracket the **whole** signature; omitting an optional argument is valid. Variadic forms have a minimum and repeated-group constraints, not an invented maximum. Two representative repeat categories seed grammar tests; they cannot enumerate an infinite set. F negatives concern call arity/context, not runtime value types. Nested expressions retain their own coverage.

The core scalar list is abs, ceil/ceiling, floor, round, len, lower, upper, trim/ltrim/rtrim, substr, replace, coalesce, if, case, match, isnull/isnotnull, tonumber/tostring, mvcount and split. The core aggregate list is count, sum, avg, min, max, dc/distinct_count, values/list and first/last. Scalar min/max have a separately documented variadic signature ([scalarstats](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/evaluation-functions/statistical-eval-functions)) outside this finite scalar core: recognize as U rather than wrongly applying aggregate arity. Additional native aliases length and c are not automatically promoted; their extra spelling/implicit-label coverage stays U.

### F.abs

Signature: **1** positional arguments. Target: A/K for approved ordinary operands, conditional on source/nullability proof. Source: [mathfunc](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/evaluation-functions/mathematical-functions). Base: ordinary function matrix; L06/L13 and relevant C03/C08–C10.

No general non-null/type guarantee from one-operand syntax.

```text
F.abs.arity-1.P1  abs(-7)
F.abs.arity-1.P2  abs(delta)
F.abs.N1  abs()
F.abs.N2  abs(delta,2)
```

### F.ceil

Signature: **1** positional arguments. Target: A/K for approved ordinary operands, conditional on source/nullability proof. Source: [mathfunc](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/evaluation-functions/mathematical-functions). Base: ordinary function matrix; L06/L13 and relevant C03/C08–C10.

ceil/ceiling documented aliases; both explicitly approved. Field value computation excluded.

```text
F.ceil.arity-1.P1  ceil(2.4)
F.ceil.arity-1.P2  ceil(cost)
F.ceil.N1  ceil()
F.ceil.N2  ceil(cost,2)
```

### F.ceiling

Signature: **1** positional arguments. Target: A/K for approved ordinary operands, conditional on source/nullability proof. Source: [mathfunc](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/evaluation-functions/mathematical-functions). Base: ordinary function matrix; L06/L13 and relevant C03/C08–C10.

ceil/ceiling documented aliases; both explicitly approved. Field value computation excluded.

```text
F.ceiling.arity-1.P1  ceiling(2.4)
F.ceiling.arity-1.P2  ceiling(cost)
F.ceiling.N1  ceiling()
F.ceiling.N2  ceiling(cost,2)
```

### F.floor

Signature: **1** positional arguments. Target: A/K for approved ordinary operands, conditional on source/nullability proof. Source: [mathfunc](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/evaluation-functions/mathematical-functions). Base: ordinary function matrix; L06/L13 and relevant C03/C08–C10.

One required expression; field value computation excluded. No general non-null/type guarantee follows from this signature.

```text
F.floor.arity-1.P1  floor(2.4)
F.floor.arity-1.P2  floor(cost)
F.floor.N1  floor()
F.floor.N2  floor(cost,2)
```

### F.round

Signature: **1 or 2** positional arguments. Target: A/K for approved ordinary operands, conditional on source/nullability proof. Source: [mathfunc](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/evaluation-functions/mathematical-functions). Base: ordinary function matrix; L06/L13 and relevant C03/C08–C10.

Omitted precision defaults0. Native negative precision disallowed; separate value-contract observation, not arity failure. No automatic expression evaluation.

```text
F.round.arity-1.P1  round(4.6)
F.round.arity-1.P2  round(cost)
F.round.arity-2.P1  round(4.6,0)
F.round.arity-2.P2  round(cost,2)
F.round.N1  round()
F.round.N2  round(cost,2,3)
```

### F.len

Signature: **1** positional arguments. Target: A/K for approved ordinary operands, conditional on source/nullability proof. Source: [textfunc](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/evaluation-functions/text-functions). Base: ordinary function matrix; L06/L13 and relevant C03/C08–C10.

length is documented but not promoted into the mandatory semantic registry. Additional alias boundary: length: A/U additional spelling.

```text
F.len.arity-1.P1  len("Abc")
F.len.arity-1.P2  len(name)
F.len.N1  len()
F.len.N2  len(name,2)
```

### F.lower

Signature: **1** positional arguments. Target: A/K for approved ordinary operands, conditional on source/nullability proof. Source: [textfunc](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/evaluation-functions/text-functions). Base: ordinary function matrix; L06/L13 and relevant C03/C08–C10.

Ordinary scalar field read; missing/type outcome depends on shared proof.

```text
F.lower.arity-1.P1  lower("Abc")
F.lower.arity-1.P2  lower(name)
F.lower.N1  lower()
F.lower.N2  lower(name,2)
```

### F.upper

Signature: **1** positional arguments. Target: A/K for approved ordinary operands, conditional on source/nullability proof. Source: [textfunc](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/evaluation-functions/text-functions). Base: ordinary function matrix; L06/L13 and relevant C03/C08–C10.

Ordinary scalar field read; missing/type outcome depends on shared proof.

```text
F.upper.arity-1.P1  upper("Abc")
F.upper.arity-1.P2  upper(name)
F.upper.N1  upper()
F.upper.N2  upper(name,2)
```

### F.trim

Signature: **1 or 2** positional arguments. Target: A/K for approved ordinary operands, conditional on source/nullability proof. Source: [textfunc](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/evaluation-functions/text-functions). Base: ordinary function matrix; L06/L13 and relevant C03/C08–C10.

Omitted trim_chars removes spaces/tabs; no wildcard interpretation of the trim set.

```text
F.trim.arity-1.P1  trim(" abc ")
F.trim.arity-1.P2  trim(name)
F.trim.arity-2.P1  trim("__abc__","_")
F.trim.arity-2.P2  trim(name,"./")
F.trim.N1  trim()
F.trim.N2  trim(name,"_",2)
```

### F.ltrim

Signature: **1 or 2** positional arguments. Target: A/K for approved ordinary operands, conditional on source/nullability proof. Source: [textfunc](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/evaluation-functions/text-functions). Base: ordinary function matrix; L06/L13 and relevant C03/C08–C10.

Omitted trim_chars removes spaces/tabs; no wildcard interpretation of the trim set.

```text
F.ltrim.arity-1.P1  ltrim(" abc ")
F.ltrim.arity-1.P2  ltrim(name)
F.ltrim.arity-2.P1  ltrim("__abc__","_")
F.ltrim.arity-2.P2  ltrim(name,"./")
F.ltrim.N1  ltrim()
F.ltrim.N2  ltrim(name,"_",2)
```

### F.rtrim

Signature: **1 or 2** positional arguments. Target: A/K for approved ordinary operands, conditional on source/nullability proof. Source: [textfunc](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/evaluation-functions/text-functions). Base: ordinary function matrix; L06/L13 and relevant C03/C08–C10.

Omitted trim_chars removes spaces/tabs; no wildcard interpretation of the trim set.

```text
F.rtrim.arity-1.P1  rtrim(" abc ")
F.rtrim.arity-1.P2  rtrim(name)
F.rtrim.arity-2.P1  rtrim("__abc__","_")
F.rtrim.arity-2.P2  rtrim(name,"./")
F.rtrim.N1  rtrim()
F.rtrim.N2  rtrim(name,"_",2)
```

### F.substr

Signature: **2 or 3** positional arguments. Target: A/K for approved ordinary operands, conditional on source/nullability proof. Source: [textfunc](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/evaluation-functions/text-functions). Base: ordinary function matrix; L06/L13 and relevant C03/C08–C10.

Omitted length means remainder; one-based and negative starts are legitimate value syntax, not missing-argument negatives.

```text
F.substr.arity-2.P1  substr(name,2)
F.substr.arity-2.P2  substr("abcd",-2)
F.substr.arity-3.P1  substr(name,2,3)
F.substr.arity-3.P2  substr("abcd",1,2)
F.substr.N1  substr(name)
F.substr.N2  substr(name,1,2,3)
```

### F.replace

Signature: **3** positional arguments. Target: A/K for approved ordinary operands, conditional on source/nullability proof. Source: [textfunc](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/evaluation-functions/text-functions). Base: ordinary function matrix; L06/L13 and relevant C03/C08–C10.

Scalar replacement differs from replace command; PCRE-body/value execution not established by arity.

```text
F.replace.arity-3.P1  replace(name,"^a","b")
F.replace.arity-3.P2  replace(path,"/","_")
F.replace.N1  replace(name,"a")
F.replace.N2  replace(name,"a","b","c")
```

### F.coalesce

Signature: **1 or more** positional arguments. Target: A/K for approved ordinary operands, conditional on source/nullability proof. Source: [comparefunc](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/evaluation-functions/comparison-and-conditional-functions). Base: ordinary function matrix; L06/L13 and relevant C03/C08–C10.

No maximum-arity negative exists. N2 is missing-list-element syntax. All-null inputs do not prove output presence; one operand is explicitly supported.

```text
F.coalesce.arity-1.P1  coalesce(primary)
F.coalesce.arity-1.P2  coalesce("fallback")
F.coalesce.arity-2+.P1  coalesce(primary,backup)
F.coalesce.arity-2+.P2  coalesce(primary,backup,"unknown")
F.coalesce.N1  coalesce()
F.coalesce.N2  coalesce(primary,)
```

### F.if

Signature: **3** positional arguments. Target: A/K for approved ordinary operands, conditional on source/nullability proof. Source: [comparefunc](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/evaluation-functions/comparison-and-conditional-functions). Base: ordinary function matrix; L06/L13 and relevant C03/C08–C10.

Result presence depends on branch values and shared proof. Command if paths are unrelated grammar. Root R01 Boolean conflict remains resolved.

```text
F.if.arity-3.P1  if(code=201,"created","other")
F.if.arity-3.P2  if(isnull(name),"unknown",name)
F.if.N1  if(code=201,"created")
F.if.N2  if(code=201,"created","other",0)
```

### F.case

Signature: **even count >=2** positional arguments. Target: A/K for approved ordinary operands, conditional on source/nullability proof. Source: [comparefunc](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/evaluation-functions/comparison-and-conditional-functions). Base: ordinary function matrix; L06/L13 and relevant C03/C08–C10.

No maximum for complete predicate/value pairs; unmatched predicates fall through to null. For variadic repetition, test2/4/6 arguments and malformed odd counts, not every impossible finite arity.

```text
F.case.arity-2.P1  case(code=201,"created")
F.case.arity-2.P2  case(true,"fallback")
F.case.arity-4+.P1  case(code=201,"created",true,"other")
F.case.arity-4+.P2  case(n<0,-1,n>=0,1)
F.case.N1  case()
F.case.N2  case(code=201)
F.case.N3  case(code=201,"created",true)
```

### F.match

Signature: **2** positional arguments. Target: A/K for approved ordinary operands, conditional on source/nullability proof. Source: [comparefunc](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/evaluation-functions/comparison-and-conditional-functions). Base: ordinary function matrix; L06/L13 and relevant C03/C08–C10.

Regex literal/string syntax and profile separate; no PCRE execution inferred.

```text
F.match.arity-2.P1  match(name,"^a")
F.match.arity-2.P2  match(path,"/tmp/")
F.match.N1  match(name)
F.match.N2  match(name,"a","b")
```

### F.isnull

Signature: **1** positional arguments. Target: A/K for approved ordinary operands, conditional on source/nullability proof. Source: [infofunc](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/evaluation-functions/informational-functions). Base: ordinary function matrix; L06/L13 and relevant C03/C08–C10.

Inspect missing/null values; avoid turning guarded/nullable fields into unconditional source-required failures. Existing Boolean eval source conflict resolved by R01.

```text
F.isnull.arity-1.P1  isnull(name)
F.isnull.arity-1.P2  isnull(null)
F.isnull.N1  isnull()
F.isnull.N2  isnull(name,backup)
```

### F.isnotnull

Signature: **1** positional arguments. Target: A/K for approved ordinary operands, conditional on source/nullability proof. Source: [infofunc](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/evaluation-functions/informational-functions). Base: ordinary function matrix; L06/L13 and relevant C03/C08–C10.

Inspect missing/null values; avoid turning guarded/nullable fields into unconditional source-required failures. Existing Boolean eval source conflict resolved by R01.

```text
F.isnotnull.arity-1.P1  isnotnull(name)
F.isnotnull.arity-1.P2  isnotnull(null)
F.isnotnull.N1  isnotnull()
F.isnotnull.N2  isnotnull(name,backup)
```

### F.tonumber

Signature: **1 or 2** positional arguments. Target: A/K for approved ordinary operands, conditional on source/nullability proof. Source: [convertfunc](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/evaluation-functions/conversion-functions). Base: ordinary function matrix; L06/L13 and relevant C03/C08–C10.

Base defaults10, native domain2..36. Unparseable field value may yield null; invalid literal string native error. These are value semantics, not parser arity or an added constant-evaluation promise.

```text
F.tonumber.arity-1.P1  tonumber("17")
F.tonumber.arity-1.P2  tonumber(amount)
F.tonumber.arity-2.P1  tonumber("1f",16)
F.tonumber.arity-2.P2  tonumber(binary,2)
F.tonumber.N1  tonumber()
F.tonumber.N2  tonumber(amount,10,2)
```

### F.tostring

Signature: **1 or 2** positional arguments. Target: A/K for approved ordinary operands, conditional on source/nullability proof. Source: [convertfunc](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/evaluation-functions/conversion-functions). Base: ordinary function matrix; L06/L13 and relevant C03/C08–C10.

Documented formats hex/commas/duration apply to numeric values. Each format gets an output-reference obligation; missing format is valid.

```text
F.tostring.arity-1.P1  tostring(17)
F.tostring.arity-1.P2  tostring(amount)
F.tostring.arity-2.P1  tostring(31,"hex")
F.tostring.arity-2.P2  tostring(seconds,"duration")
F.tostring.arity-2.P3  tostring(amount,"commas")
F.tostring.N1  tostring()
F.tostring.N2  tostring(amount,"hex",2)
```

### F.mvcount

Signature: **1** positional arguments. Target: A/K for approved ordinary operands, conditional on source/nullability proof. Source: [mvfunc](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/evaluation-functions/multivalue-eval-functions). Base: ordinary function matrix; L06/L13 and relevant C03/C08–C10.

No values => null; single value =>1. Nullability cannot be erased by treating this as arbitrary non-null pure function.

```text
F.mvcount.arity-1.P1  mvcount(tags)
F.mvcount.arity-1.P2  mvcount(split(name,":"))
F.mvcount.N1  mvcount()
F.mvcount.N2  mvcount(tags,2)
```

### F.split

Signature: **2** positional arguments. Target: A/K for approved ordinary operands, conditional on source/nullability proof. Source: [mvfunc](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/evaluation-functions/multivalue-eval-functions). Base: ordinary function matrix; L06/L13 and relevant C03/C08–C10.

Delimiter required; produces multivalue. Keep field reads separate from delimiter literal.

```text
F.split.arity-2.P1  split(name,":")
F.split.arity-2.P2  split("a/b","/")
F.split.N1  split(name)
F.split.N2  split(name,":",2)
```

### F.count

Signature: **aggregate 0 or 1** positional arguments. Target: A/K for approved ordinary operands, conditional on source/nullability proof. Source: [aggregatefunc](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/statistical-and-charting-functions/aggregate-functions). Base: ordinary function matrix; L06/L13 and relevant C03/C08–C10.

count() has no operand reads; argument-bearing calls retain reads from their actual expressions. The proved implicit output of count() without AS is count. Operand-bearing count calls have unproved implicit labels and remain U unless separately evidenced and approved; an explicit alias resolves the output name. Predicate/search-literal argument forms need their own expression-coverage state. No below-zero arity negative; bare keyword tests mandatory parentheses. Additional alias boundary: c: A/U additional spelling; implicit label unproved.

```text
F.count.arity-0.P1  stats count() AS n
F.count.arity-0.P2  stats count() AS n BY host
F.count.arity-1.P1  stats count(user) AS n
F.count.arity-1.P2  stats count(code=201) AS n
F.count.N1  stats count(user,code) AS n
F.count.N2  stats count AS n
```

### F.sum

Signature: **aggregate 1** positional arguments. Target: A/K for approved ordinary operands, conditional on source/nullability proof. Source: [aggregatefunc](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/statistical-and-charting-functions/aggregate-functions). Base: ordinary function matrix; L06/L13 and relevant C03/C08–C10.

Ordinary field/expression input; empty-group/null result not universally documented.

```text
F.sum.arity-1.P1  stats sum(bytes) AS result
F.sum.arity-1.P2  stats sum(duration) AS result
F.sum.N1  stats sum() AS result
F.sum.N2  stats sum(bytes,duration) AS result
```

### F.avg

Signature: **aggregate 1** positional arguments. Target: A/K for approved ordinary operands, conditional on source/nullability proof. Source: [aggregatefunc](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/statistical-and-charting-functions/aggregate-functions). Base: ordinary function matrix; L06/L13 and relevant C03/C08–C10.

Ordinary field/expression input; empty-group/null result not universally documented.

```text
F.avg.arity-1.P1  stats avg(bytes) AS result
F.avg.arity-1.P2  stats avg(duration) AS result
F.avg.N1  stats avg() AS result
F.avg.N2  stats avg(bytes,duration) AS result
```

### F.min

Signature: **aggregate 1** positional arguments. Target: A/K for approved ordinary operands, conditional on source/nullability proof. Source: [aggregatefunc](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/statistical-and-charting-functions/aggregate-functions). Base: ordinary function matrix; L06/L13 and relevant C03/C08–C10.

Aggregate min/max one value; scalar min/max accepts1+, outside approved scalar registry and remains U rather than applying aggregate arity errors.

```text
F.min.arity-1.P1  stats min(bytes) AS result
F.min.arity-1.P2  stats min(duration) AS result
F.min.N1  stats min() AS result
F.min.N2  stats min(bytes,duration) AS result
```

### F.max

Signature: **aggregate 1** positional arguments. Target: A/K for approved ordinary operands, conditional on source/nullability proof. Source: [aggregatefunc](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/statistical-and-charting-functions/aggregate-functions). Base: ordinary function matrix; L06/L13 and relevant C03/C08–C10.

Aggregate min/max one value; scalar min/max accepts1+, outside approved scalar registry and remains U rather than applying aggregate arity errors.

```text
F.max.arity-1.P1  stats max(bytes) AS result
F.max.arity-1.P2  stats max(duration) AS result
F.max.N1  stats max() AS result
F.max.N2  stats max(bytes,duration) AS result
```

### F.dc

Signature: **aggregate 1** positional arguments. Target: A/K for approved ordinary operands, conditional on source/nullability proof. Source: [aggregatefunc](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/statistical-and-charting-functions/aggregate-functions). Base: ordinary function matrix; L06/L13 and relevant C03/C08–C10.

One required value; descriptive dc() shorthand does not prove zero-arg validity.

```text
F.dc.arity-1.P1  stats dc(bytes) AS result
F.dc.arity-1.P2  stats dc(duration) AS result
F.dc.N1  stats dc() AS result
F.dc.N2  stats dc(bytes,duration) AS result
```

### F.distinct_count

Signature: **aggregate 1** positional arguments. Target: A/K for approved ordinary operands, conditional on source/nullability proof. Source: [aggregatefunc](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/statistical-and-charting-functions/aggregate-functions). Base: ordinary function matrix; L06/L13 and relevant C03/C08–C10.

One required value; descriptive dc() shorthand does not prove zero-arg validity.

```text
F.distinct_count.arity-1.P1  stats distinct_count(bytes) AS result
F.distinct_count.arity-1.P2  stats distinct_count(duration) AS result
F.distinct_count.N1  stats distinct_count() AS result
F.distinct_count.N2  stats distinct_count(bytes,duration) AS result
```

### F.values

Signature: **aggregate 1** positional arguments. Target: A/K for approved ordinary operands, conditional on source/nullability proof. Source: [multiagg](https://help.splunk.com/en/splunk-cloud-platform/search/spl2-search-reference/statistical-and-charting-functions/multivalue-and-array-functions). Base: ordinary function matrix; L06/L13 and relevant C03/C08–C10.

Distinct lexicographically ordered multivalue; field effects not value evaluation.

```text
F.values.arity-1.P1  stats values(user) AS result
F.values.arity-1.P2  stats values(code) AS result
F.values.N1  stats values() AS result
F.values.N2  stats values(user,code) AS result
```

### F.list

Signature: **aggregate 1** positional arguments. Target: A/K for approved ordinary operands, conditional on source/nullability proof. Source: [multiagg](https://help.splunk.com/en/splunk-cloud-platform/search/spl2-search-reference/statistical-and-charting-functions/multivalue-and-array-functions). Base: ordinary function matrix; L06/L13 and relevant C03/C08–C10.

Encounter-order multivalue; native first100 values. Field effects do not require evaluating value ordering.

```text
F.list.arity-1.P1  stats list(user) AS result
F.list.arity-1.P2  stats list(code) AS result
F.list.N1  stats list() AS result
F.list.N2  stats list(user,code) AS result
```

### F.first

Signature: **aggregate 1** positional arguments. Target: A/K for approved ordinary operands, conditional on source/nullability proof. Source: [eventorder](https://help.splunk.com/en/splunk-cloud-platform/search/spl2-search-reference/statistical-and-charting-functions/event-order-functions). Base: ordinary function matrix; L06/L13 and relevant C03/C08–C10.

Encounter-order, not chronological timestamp order. Per-function page omits eventstats without forbidding it; general stats overview establishes command applicability.

```text
F.first.arity-1.P1  stats first(user) AS result
F.first.arity-1.P2  stats first(code) AS result
F.first.N1  stats first() AS result
F.first.N2  stats first(user,code) AS result
```

### F.last

Signature: **aggregate 1** positional arguments. Target: A/K for approved ordinary operands, conditional on source/nullability proof. Source: [eventorder](https://help.splunk.com/en/splunk-cloud-platform/search/spl2-search-reference/statistical-and-charting-functions/event-order-functions). Base: ordinary function matrix; L06/L13 and relevant C03/C08–C10.

Encounter-order, not chronological timestamp order. Per-function page omits eventstats without forbidding it; general stats overview establishes command applicability.

```text
F.last.arity-1.P1  stats last(user) AS result
F.last.arity-1.P2  stats last(code) AS result
F.last.N1  stats last() AS result
F.last.N2  stats last(user,code) AS result
```

## Finite alternative and interaction obligations

The following substitutions are **document-only fixture obligations**, not parsing or query-rewrite code. For each listed alternative, instantiate both positive contexts and both minimal-negative contexts; retain the matching E row's K/U status. They finish the alternative inventory without presenting repeated spellings as unique breadth. Blank optional alternatives must be tested as omission, never as missing-operand negatives.

| Obligation / source | Alternatives | Two positive contexts | Two negative contexts |
| --- | --- | --- | --- |
| Search comparisons; E.C01.comparison; [doc163](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/search-command/search-command-overview-and-syntax) | =, !=, <, <=, >, >= | `search index=app bytes OP 2`; `search index=app rate OP 2.5` | Same sources with RHS removed; same sources with unterminated quoted RHS |
| Expression comparisons; E.L07; [predicate](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-manual/expressions-and-predicates/predicate-expressions) | =, ==, !=, <, <=, >, >= | `where bytes OP limit`; `eval answer=(bytes OP 2)` | Missing RHS; missing final parenthesis in second context |
| Search selectors; E.C01.selector; [doc163](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/search-command/search-command-overview-and-syntax) | index, source, sourcetype, host | `search SELECTOR="app"`; `search SELECTOR="app" failure` | Missing value; unterminated value string |
| Boolean operators; C01/C04/L06; [predicate](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-manual/expressions-and-predicates/predicate-expressions), [doc163](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/search-command/search-command-overview-and-syntax) | Search AND, OR, XOR; expression AND, OR, XOR | `search index=app a=1 OP b=2`; `where a=1 OP b=2` | Missing second operand in each context; duplicate adjacent binary operator in each context. Separate unary NOT witnesses already in E rows. |
| Sort wrappers; E.C12.order_terms; [doc164](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/sort-command/sort-command-overview-syntax-and-usage) | auto, ip, num, str; separate plain field | `sort WRAPPER(user)`; `sort -WRAPPER(user),host` | Empty wrapper operand; missing closing parenthesis. Plain field uses E.C12 with missing field/list delimiter negatives. |
| Sort signs; E.C12.order_terms; [doc164](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/sort-command/sort-command-overview-syntax-and-usage) | omitted, +, - | `sort SIGNbytes`; `sort host,SIGNbytes` | Missing field after a present sign; trailing comma. Omitted sign is valid, not a negative. |
| Join type; E.C16.type; [doc148](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/join-command/join-command-overview-syntax-and-usage) | inner, left, outer | `join type=TYPE left=L right=R where L.id=R.id [FROM other]`; same with equality operands reversed | Missing type value; missing required right alias. No full join promise. |
| Bin align; E.C23.aligntime; [doc131](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/bin-command/bin-command-overview-syntax-and-usage) | earliest, latest, evidenced snap+offset | `bin span=2hr aligntime=ALIGN _time`; same with `AS bucket` | Missing align value; missing field. Malformed snap+offset has its E-row negative. |
| Timechart axis versus split options; E.C30; [doc172](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/timechart-command/timechart-command-overview-and-syntax) | bins=10, minspan=2min, span=5min, start=0, end=50, aligntime=latest | `timechart OPTION count()`; `timechart avg(bytes) BY bucket OPTION` | Missing OPTION value in both positions; incomplete BY field in split context. Each option retains U. |
| Timewrap unit classes; E.C31.span; [doc173](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/timewrap-command/timewrap-command-overview-syntax-and-usage) | second, minute/min, hour, day, week, month/m, quarter, year; exact synonymous spellings remain source-table assertions | `FROM main \| timechart count() \| timewrap UNIT`; same with `2UNIT align=end` | Missing complete span; `2` without unit. Do not classify an unreviewed spelling as invalid. |
| SQL grouping span forms; E.L16.group_span/Q05; [fromsyntax](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/from-command/from-command-syntax) | `span(_time)`, `span(_time,5min)`, `span(_time,hour)`, `_time span=(5min)` | `SELECT count() FROM main GROUP BY FORM`; `FROM main GROUP BY FORM SELECT count()` | Missing span field or empty explicit span value as appropriate; missing closing delimiter. Omitted optional span value inside span(field) is valid. |
| SQL group/order spellings; E.L16; [fromsyntax](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/from-command/from-command-syntax) | GROUP BY/GROUPBY; ORDER BY/ORDERBY | SELECT-first with `GROUPKEY host` and FROM-first with same; SELECT-first and FROM-first `ORDERKEY host` | Missing grouping/order expression; trailing comma. Per-term direction remains H07. |
| Positional optional args; F.round/trim/ltrim/rtrim/substr/tonumber/tostring | Every arity and delimiter already enumerated in F | Each arity's two F positives | Shared lower/upper signature violations plus trailing comma. Valid optional omissions are not errors. |
| Repeat groups; F.coalesce/F.case | One and multiple coalesce values; one and multiple condition/value pairs | Each F repeat category's two positives | Zero coalesce values; trailing comma; odd case operands. No artificial maximum argument count. |

Additional interactions require source, state and phase assertions beyond their outer form. These examples are attached obligations, not added to the unique 288-seed count:

```text
I.assign.P1  FROM main | eval first=bytes, next=first+1
I.assign.P2  FROM main | eval first=null, next=first+1
I.assign.N1  FROM main | eval first=bytes next=first+1
I.assign.N2  FROM main | eval first=null, next=
I.bin.P1  FROM main | bin bins=10 span=5 bytes
I.bin.P2  FROM main | bin start=0 end=100 bytes AS bucket
I.bin.N1  FROM main | bin bins= span=5 bytes
I.bin.N2  FROM main | bin start=0 end=100 bytes AS
I.timechart.P1  FROM main | timechart bins=10 span=5min count() BY host
I.timechart.P2  FROM main | timechart sep="::" format="$AGG$:$VAL$" count() BY host useother=false otherstr="rest"
I.timechart.N1  FROM main | timechart bins=10 span= count() BY host
I.timechart.N2  FROM main | timechart sep="::" format= count() BY host
I.named.P1  FROM main | eval result=round(bytes,precision:2)
I.named.P2  FROM main | eval result=if(false_value:0,predicate:code=200,true_value:1)
I.named.N1  FROM main | eval result=round(precision:2,bytes)
I.named.N2  FROM main | eval result=coalesce(values:a,b)
I.object.P1  FROM main | eval record={a:1,nested:{a:2}}
I.object.P2  FROM main | eval record={a:user,b:host}
I.object.N1  FROM main | eval record={a:1,a:2}
I.object.N2  FROM main | eval record={name:user,"name":host}
I.literal.P1  FROM main | eval class=if(`error=4*`,"user","server")
I.literal.P2  SELECT count(`500`) AS failures FROM main
I.literal.N1  FROM main | eval class=if(`error=4*,"user","server")
I.literal.N2  SELECT count(`500) AS failures FROM main
```

I.assign.P2 is grammatical but must diagnose a sound missing read when first was just removed; it is not a valid-state positive. I.bin and I.timechart stay U: span overrides bins, start/end expand a range where applicable, format overrides sep, and disabled useother does not make otherstr syntactically invalid. Keep separate axis placement and option-source slices ([doc131](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/bin-command/bin-command-overview-syntax-and-usage), [doc172](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/timechart-command/timechart-command-overview-and-syntax), [timechartusage](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/timechart-command/timechart-command-usage)). I.named stays U under the root ruling; colon labels are not reads. I.object duplicate keys are located static contract-invalid findings scoped per object; quoted-equivalent and escaped-equivalent keys require decoding assertions ([objects](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-manual/expressions-and-predicates/array-and-object-literals-in-expressions)). I.literal distinguishes expression-position search literals from command embeddings; inner syntax/semantics coverage remains separate ([searchliterals](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-manual/expressions-and-predicates/search-literals-in-expressions)).

For each aggregate F signature, pair explicit output aliases with native implicit outputs in stats and SQL, plus eventstats/streamstats preservation contexts. Scalar F signatures are exercised in eval, WHERE or SELECT expressions and supported aggregate operands, with context-appropriate field/Boolean rules; they are not direct aggregate calls merely because their arity is known. Expected native aggregate labels are count, sum(bytes), dc(action) for the approved examples; other labels need an explicit alias or U. Ordinary field reads, nested pure reads, null input, literal null, and conditional null branches need separate state assertions. coalesce(all-null), case without a matching fallback, nullable if branches and empty mvcount cannot fabricate definite output availability. Known string/numeric argument domains (round precision, tonumber base 2–36, tostring hex/commas/duration) are source observations, not a new general runtime evaluator/type checker. The F registry must report unsupported semantics independently from recognized call syntax.

## Additional held evidence — separate from H01–H12

These expansion keys preserve the approved base identities. They are U and outside complete-support or definite-negative floors. A strict supported-command grammar must not accidentally convert a recognized disputed layout to syntax-invalid. This does not authorize opaque acceptance of arbitrary malformed supported syntax.

| Key | Evidence and candidate | Disposition |
| --- | --- | --- |
| EH01 | Bin's own unit table/examples ([doc131](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/bin-command/bin-command-overview-syntax-and-usage), [binexamples](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/bin-command/bin-command-examples)) are narrower than the general July27 span vocabulary ([timespans](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-manual/dates-and-time/specifying-time-spans)). Extra `bin span=1w _time` or unit-only `bin span=hour _time` is unproved by the selected command evidence. | Additional aliases/weekly/unit-only bin spans stay U. Explicit admitted integer-plus-unit witnesses and bare numeric bins remain as in E.C23. Do not silently union all commands' unit vocabularies. |
| EH02 | Timechart's normative prefix `agg=(sum(bytes))` ([doc172](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/timechart-command/timechart-command-overview-and-syntax)) conflicts with usage's bare `agg=sum` ([timechartusage](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/timechart-command/timechart-command-usage)). Candidate `FROM main \| timechart agg=sum limit=5 avg(bytes) BY host`. | Bare prefix recognized/disputed U, not malformed negative. This is separate from original H04's multi-aggregate delimiters. Independently evidenced eval(expression) BY form stays A/U. |
| EH03 | [timemodifiers](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-manual/dates-and-time/time-modifiers) restricts time modifier comparisons to =/!= in prose but illustrates latest<@d and earliest>=@d+1h. | `search index=app latest<@d` and `search index=app earliest>=@d+1h` remain disputed U. General field comparison evidence does not settle time-selector semantics. |
| EH04 | [predicate](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-manual/expressions-and-predicates/predicate-expressions) documents IS NOT NULL and NOT(expr IS type), without proving IS NOT type. Candidate `FROM main \| where user IS NOT string`. | Unproved extra layout U, not definite-invalid. Corrected L07 summary has no changed seed expectations. |
| EH05 | [fromsyntax](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/from-command/from-command-syntax) specifies field span=(value) in its synopsis but illustrates _time span=h and _time span=3h in its detailed span table. Candidate `FROM main GROUP BY _time span=3h SELECT count()`. | The extra unparenthesized assignment layout stays disputed/incomplete. Existing Q05 and E.L16 parenthesized/function-call seeds are unchanged; no missing-parentheses negative is inferred from the synopsis. |

H12's slash-plus-internal-pipe rex pattern remains held; this supplement admits quoted alternation and the previously corrected slash character-class form only. H01–H11 retain their original meanings, including source-specific option order, SQL direction/range/visibility, and casing boundaries. Unproved single-element search IN, extra implicit aggregate names, named aggregate signatures and undocumented option spellings remain U; absence from an audited syntax synopsis alone is not a universal native-invalid proof.

The time-modifier page lists older still-valid spellings which may be removed in a future version. The advertised starttime/endtime/timeformat forms are covered above. Further daysago/enddaysago/endhoursago/endminutesago/endmonthsago/endtimeu/hoursago/minutesago/monthsago/searchtimespandays/searchtimespanhours/searchtimespanminutes/searchtimespanmonths/startdaysago/starthoursago/startminutesago/startmonthsago/starttimeu forms remain outside this finite expansion/U, not newly invalid merely because deprecated.

## Provenance and remaining execution gates

The ledger below contains primary Splunk pages actually reviewed in this task or its bounded read-only function audit. Retrieval date is 2026-09-07 EDT, extending into 2026-09-08 UTC. Updated timestamps are publisher/tool-returned values, not deployed-engine versions. Enterprise and Cloud URLs share language reference material; profile applicability is independently checked against the splunkd tables. Source-specific examples are distinguished from module examples and Edge/Ingest migration notes. This documentation contract does not claim execution against a Splunk instance.

| Key | Primary source | Publisher updated metadata |
| --- | --- | --- |
| doc163 | [search command: Overview and syntax](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/search-command/search-command-overview-and-syntax) | 2026-06-16T03:49:12.086Z |
| predicate | [Predicate expressions](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-manual/expressions-and-predicates/predicate-expressions) | 2026-07-27T18:10:44.526Z |
| relativetime | [Specifying relative time](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-manual/dates-and-time/specifying-relative-time) | 2026-07-27T18:10:52.125Z |
| timemodifiers | [Time modifiers](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-manual/dates-and-time/time-modifiers) | 2026-07-27T18:10:52.875Z |
| doc136 | [eval command: Overview and syntax](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/eval-command/eval-command-overview-and-syntax) | 2026-06-16T03:48:58.817Z |
| doc178 | [where command: Overview, syntax, and usage](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/where-command/where-command-overview-syntax-and-usage) | 2026-06-16T03:49:03.447Z |
| doc139 | [fields command: Overview, syntax, and usage](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/fields-command/fields-command-overview-syntax-and-usage) | 2026-07-14T22:31:06.977Z |
| doc169 | [table command: Overview, syntax, and usage](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/table-command/table-command-overview-syntax-and-usage) | 2026-06-16T03:56:32.628Z |
| doc158 | [rename command: Overview, syntax, and usage](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/rename-command/rename-command-overview-syntax-and-usage) | 2026-07-14T22:31:29.013Z |
| renameexamples | [rename command: Examples](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/rename-command/rename-command-examples) | 2026-06-16T03:49:15.515Z |
| doc167 | [stats command: Overview, syntax, and usage](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/stats-command/stats-command-overview-syntax-and-usage) | 2026-06-16T03:49:04.920Z |
| eventtail | [eventstats command: Overview, syntax, and usage](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/eventstats-command/eventstats-command-overview-syntax-and-usage) | 2026-06-16T03:49:26.621Z |
| doc168 | [streamstats command: Overview, syntax, and usage](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/streamstats-command/streamstats-command-overview-syntax-and-usage) | 2026-06-16T03:49:25.435Z |
| doc150 | [lookup command: Overview, syntax, and usage](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/lookup-command/lookup-command-overview-syntax-and-usage) | 2026-06-16T03:49:28.119Z |
| sortexamples | [sort command: Examples](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/sort-command/sort-command-examples) | 2026-06-16T03:49:34.167Z |
| doc164 | [sort command: Overview, syntax, and usage](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/sort-command/sort-command-overview-syntax-and-usage) | 2026-06-16T03:49:11.080Z |
| doc135 | [dedup command: Overview, syntax, and usage](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/dedup-command/dedup-command-overview-syntax-and-usage) | 2026-06-16T03:49:17.646Z |
| doc144 | [head command: Overview, syntax, and usage](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/head-command/head-command-overview-syntax-and-usage) | 2026-06-16T03:49:09.722Z |
| doc130 | [appendpipe command: Overview, syntax, and usage](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/appendpipe-command/appendpipe-command-overview-syntax-and-usage) | 2026-06-16T03:49:26.243Z |
| doc172 | [timechart command: Overview and syntax](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/timechart-command/timechart-command-overview-and-syntax) | 2026-06-16T03:48:57.036Z |
| doc148 | [join command: Overview, syntax, and usage](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/join-command/join-command-overview-syntax-and-usage) | 2026-06-22T06:10:19.972Z |
| doc129 | [append command: Overview, syntax, and usage](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/append-command/append-command-overview-syntax-and-usage) | 2026-06-16T03:49:15.944Z |
| doc128 | [appendcols command: Overview, syntax, and usage](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/appendcols-command/appendcols-command-overview-syntax-and-usage) | 2026-06-16T03:49:29.063Z |
| doc176 | [union command: Overview, syntax, and usage](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/union-command/union-command-overview-syntax-and-usage) | 2026-06-16T03:49:14.314Z |
| doc145 | [if command: Overview, syntax, and usage](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/if-command/if-command-overview-syntax-and-usage) | 2026-07-14T22:25:10.340Z |
| doc166 | [spl1 command: Overview, syntax, and usage](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/spl1-command/spl1-command-overview-syntax-and-usage) | 2026-06-16T03:49:32.110Z |
| doc131 | [bin command: Overview, syntax, and usage](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/bin-command/bin-command-overview-syntax-and-usage) | 2026-06-16T03:49:28.404Z |
| binexamples | [bin command: Examples](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/bin-command/bin-command-examples) | 2026-06-16T03:49:18.560Z |
| doc161 | [rex command: Overview, syntax, and usage](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/rex-command/rex-command-overview-syntax-and-usage) | 2026-06-18T09:04:43.715Z |
| doc165 | [spath command: Overview, syntax, and usage](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/spath-command/spath-command-overview-syntax-and-usage) | 2026-06-16T04:28:00.227Z |
| doc152 | [makeresults command: Overview, syntax, and usage](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/makeresults-command/makeresults-command-overview-syntax-and-usage) | 2026-06-19T04:41:34.012Z |
| doc149 | [loadjob command: Overview, syntax, and usage](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/loadjob-command/loadjob-command-overview-syntax-and-usage) | 2026-06-20T15:03:23.100Z |
| doc174 | [tstats command: Overview, syntax, usage](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/tstats-command/tstats-command-overview-syntax-usage) | 2026-06-16T03:55:44.015Z |
| doc153 | [mstats command: Overview, syntax, and usage](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/mstats-command/mstats-command-overview-syntax-and-usage) | 2026-06-19T04:36:55.476Z |
| timechartexamples | [timechart command: Examples](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/timechart-command/timechart-command-examples) | 2026-06-16T03:49:12.541Z |
| doc173 | [timewrap command: Overview, syntax, and usage](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/timewrap-command/timewrap-command-overview-syntax-and-usage) | 2026-06-16T03:49:24.644Z |
| doc151 | [makemv command: Overview, syntax, and usage](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/makemv-command/makemv-command-overview-syntax-and-usage) | 2026-06-19T04:43:45.874Z |
| doc155 | [mvexpand command: Overview, syntax, and usage](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/mvexpand-command/mvexpand-command-overview-syntax-and-usage) | 2026-06-16T03:48:57.365Z |
| doc154 | [mvcombine command: Overview, syntax, and usage](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/mvcombine-command/mvcombine-command-overview-syntax-and-usage) | 2026-06-18T18:05:52.219Z |
| doc141 | [fillnull command: Overview, syntax, and usage](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/fillnull-command/fillnull-command-overview-syntax-and-usage) | 2026-06-16T03:49:34.600Z |
| start | [Start searching data using SPL2](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-manual/getting-started/quick-start-write-and-run-a-basic-spl2-search/start-searching-data-using-spl2) | 2026-07-27T18:10:44.254Z |
| syntax | [Understanding SPL2 syntax](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/introduction/understanding-spl2-syntax) | 2026-06-16T03:49:14.478Z |
| types | [Built-in data types](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-manual/data-types/built-in-data-types) | 2026-07-27T18:10:46.819Z |
| expressions | [Types of expressions](https://help.splunk.com/en/splunk-cloud-platform/search/spl2-search-manual/expressions-and-predicates/types-of-expressions) | 2026-02-06T04:57:27.306Z |
| comments | [Using comments in SPL2](https://help.splunk.com/en/splunk-cloud-platform/search/spl2-search-manual/comments/using-comments-in-spl2) | 2026-07-27T18:10:51.652Z |
| objects | [Array and object literals in expressions](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-manual/expressions-and-predicates/array-and-object-literals-in-expressions) | 2026-07-27T18:10:48.823Z |
| access | [Access expressions for arrays and objects](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-manual/expressions-and-predicates/access-expressions-for-arrays-and-objects) | 2026-07-27T18:10:48.858Z |
| templates | [String templates in expressions](https://help.splunk.com/en/splunk-cloud-platform/search/spl2-search-manual/expressions-and-predicates/string-templates-in-expressions) | 2026-07-27T18:10:48.768Z |
| fieldtemplates | [Field templates in expressions](https://help.splunk.com/en/splunk-cloud-platform/search/spl2-search-manual/expressions-and-predicates/field-templates-in-expressions) | 2026-07-27T18:10:48.895Z |
| lambda | [Lambda expressions](https://help.splunk.com/en/splunk-cloud-platform/search/spl2-search-manual/expressions-and-predicates/lambda-expressions) | 2026-07-27T18:10:48.930Z |
| fromusage | [from command: Usage](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/from-command/from-command-usage) | 2026-06-16T03:49:00.226Z |
| fromsyntax | [from command: Syntax](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/from-command/from-command-syntax) | 2026-01-16T00:21:47.256Z |
| process | [Process your search results](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-manual/getting-started/quick-start-write-and-run-a-basic-spl2-search/process-your-search-results) | 2026-07-27T18:10:44.598Z |
| searchliterals | [Search literals in expressions](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-manual/expressions-and-predicates/search-literals-in-expressions) | 2026-07-27T18:10:48.796Z |
| namingargs | [Naming function arguments](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-manual/functions/naming-function-arguments) | 2026-07-27T18:10:50.954Z |
| mathfunc | [Mathematical functions](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/evaluation-functions/mathematical-functions) | 2026-06-16T03:49:07.004Z |
| textfunc | [Text functions](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/evaluation-functions/text-functions) | 2026-06-16T03:49:34.458Z |
| comparefunc | [Comparison and Conditional functions](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/evaluation-functions/comparison-and-conditional-functions) | 2026-06-16T03:48:59.359Z |
| infofunc | [Informational functions](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/evaluation-functions/informational-functions) | 2026-06-16T03:49:05.865Z |
| convertfunc | [Conversion functions](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/evaluation-functions/conversion-functions) | 2026-06-16T03:49:01.131Z |
| mvfunc | [Multivalue eval functions](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/evaluation-functions/multivalue-eval-functions) | 2026-06-18T18:13:46.678Z |
| aggregatefunc | [Aggregate functions](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/statistical-and-charting-functions/aggregate-functions) | 2026-06-16T03:55:23.106Z |
| multiagg | [Multivalue and array functions](https://help.splunk.com/en/splunk-cloud-platform/search/spl2-search-reference/statistical-and-charting-functions/multivalue-and-array-functions) | 2026-06-16T03:49:15.805Z |
| eventorder | [Event order functions](https://help.splunk.com/en/splunk-cloud-platform/search/spl2-search-reference/statistical-and-charting-functions/event-order-functions) | 2026-06-16T03:49:11.605Z |
| timechartusage | [timechart command: Usage](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/timechart-command/timechart-command-usage) | 2026-06-16T03:49:01.452Z |
| timespans | [Specifying time spans](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-manual/dates-and-time/specifying-time-spans) | 2026-07-27T18:10:52.158Z |
| statsintro | [Overview of SPL2 stats and chart functions](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/statistical-and-charting-functions/overview-of-spl2-stats-and-chart-functions) | 2026-06-16T03:48:58.663Z |
| scalarstats | [Statistical eval functions](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/evaluation-functions/statistical-eval-functions) | 2026-06-16T03:49:27.478Z |
| evalcompat | [Compatibility Quick Reference for SPL2 evaluation functions](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/spl2-compatibility-profiles-and-quick-references/compatibility-quick-reference-for-spl2-evaluation-functions) | 2026-07-06T19:41:16.114Z |
| aggcompat | [Compatibility Quick Reference for SPL2 statistical functions](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/spl2-compatibility-profiles-and-quick-references/compatibility-quick-reference-for-spl2-statistical-functions) | 2026-06-16T03:50:01.647Z |

The design expansion now identifies each advertised option/form and approved positional arity, with concrete positive/minimal-negative obligations and held boundaries. Before any implementation claim: controller review must accept this supplement; verified M4 must resolve the actual shared API and phase types; the authorized M5 plan must convert these obligations into deduplicated durable fixtures with exact source slices, scope/reads/writes, before/after states, dialect/profile outcomes and adapter parity; the new ANTLR frontend must run those fixtures. Preserve the approved meaningful-fixture floors after deduplication. No such plan, parser, runtime test, or commit was performed here. The original independent [review](conformance-review.md) and [re-review](conformance-rereview.md) remain immutable evidence.
