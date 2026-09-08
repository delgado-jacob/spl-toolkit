# Standalone SPL2 conformance matrix — preflight

Status: design evidence and purpose-written fixture seeds for controller review, retrieved 2026-09-07. No SPL2 grammar, implementation, test data, executed fixture result, production edits, implementation plan, or commit is represented here. The final shared lowering and SQL phase API waits for verified M4.

This matrix operationalizes the approved [design](design-spec.md) for one explicit SPL2 standalone Search & Reporting search in the splunkd profile. It is a documentation-snapshot contract, not a runtime certification against any deployed Splunk version. Every primary source below is an official Splunk reference; product context and contradictory guidance are recorded.

The bounded [form and function expansion](form-expansion.md) supplies per-option/form obligations and the finite approved positional arities. It preserves all 288 seed identities and H01–H12 below; its E/F obligations are not extra unique-fixture counts or executed tests. Named calls remain explicitly incomplete, and new unproved variants are separately labeled EH01–EH05.

## Counting and outcome rules

The official compatibility denominator has **53 rows: 50 splunkd entries and 3 other-profile entries**. FROM and SELECT occupy separate official rows but share one of the **35 mandatory dedicated command families**. The remaining **14 native entries** are accounted for as deferred grammar/effects. There are **25 language-family cards** and **8 additional SQL form cards**.

The 68 cards contain **272 labeled P/N seeds**, plus **16 profile/module/contract seeds**, for **288 candidate seeds** before duplicate-content review. Each card provides two purpose-written positives and two minimal negatives. These are design candidates, not passing tests. Definite syntax/contract/profile/module negatives are counted separately from unresolved H forms, which are never counted toward negative floors. Command option variants listed within a card need the typed-operand/delimiter assertions and additional pair expansion below before those individual variants are advertised as exhaustively covered. Repeating letter case or arbitrary data values does not add a meaningful form.

**P** means the documented outer syntax should be recognized, with the card's separate field-semantic expectation. It does not mean valid: embedded SPL body, dynamic effects, and disputed semantic visibility can remain incomplete. **N** means a located definite malformed or contract-invalid case; lexical/structural malformed syntax makes syntax coverage false, while a successfully parsed but invalid function arity/context can retain syntax coverage true and report a semantic/compatibility error. **K** denotes the approved complete field-model subset, conditional on supported arguments and verified kernel proofs. **U** denotes incomplete effects or unsupported syntax as specified. **H** denotes an evidenced conflict/unresolved form with incomplete outcome; no permissive parse success and no accidental syntax-invalid upgrade.

Command-card snippets are command tails. Prefix `FROM main | ` unless the snippet already starts a generating search/source (FROM, SELECT, search, index=, union, makeresults, loadjob, tstats, mstats, spl1 or backtick). Language, Q, and B cases are complete queries. The multiline comment case contains an actual newline. No leading-pipe assumption is needed for this corpus.

Each eventual durable fixture must record ID; dialect/profile/version; source ID; exact input; command/form labels; provenance key and retrieval date; expected syntax, semantic coverage, status and codes; references/bindings; source slices; scopes; lexical stage order; logical clause-entry positions; lineage phase/execution order when needed; and source-universe assumptions. A capability claim is tied to reviewed forms, never just a command name.

## Complete command inventory

Native membership comes from [compat](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/spl2-compatibility-profiles-and-quick-references/compatibility-quick-reference-for-spl2-commands) (updated 2026-07-14). Recognition/deferred state is the toolkit design, not a claim about Splunk support.

| Official entry | splunkd | M5 classification | Standalone context |
| --- | --- | --- | --- |
| addinfo | Yes | Deferred grammar/effects; U | Native inventory; not advertised grammar |
| appendcols | Yes | C19 dedicated; U | Selected standalone scope |
| append | Yes | C17 dedicated; U | Selected standalone scope |
| appendpipe | Yes | C18 dedicated; U | Selected standalone scope |
| bin | Yes | C23 dedicated; U | Selected standalone scope |
| branch | Yes | Deferred grammar/effects; U | Search syntax evidenced; terminal, each branch ends into |
| convert | Yes | Deferred grammar/effects; U | Native inventory; not advertised grammar |
| decrypt | No | Profile exclusion | Outside splunkd |
| dedup | Yes | C13 dedicated; K | Selected standalone scope |
| eval | Yes | C03 dedicated; K | Selected standalone scope |
| eventstats | Yes | C09 dedicated; K | Selected standalone scope |
| expand | Yes | Deferred grammar/effects; U | Native inventory; not advertised grammar |
| fields | Yes | C05 dedicated; K | Selected standalone scope |
| fieldsummary | Yes | Deferred grammar/effects; U | Native inventory; not advertised grammar |
| fillnull | Yes | C35 dedicated; U | Selected standalone scope |
| flatten | Yes | Deferred grammar/effects; U | Native inventory; not advertised grammar |
| from | Yes | C02 dedicated; K | Selected standalone scope |
| head | Yes | C14 dedicated; K | Selected standalone scope |
| if | Yes | C21 dedicated; U | Selected standalone scope |
| into | Yes | Deferred grammar/effects; U | Search sink exists; native, deferred |
| iplocation | Yes | Deferred grammar/effects; U | Native inventory; not advertised grammar |
| join | Yes | C16 dedicated; U | Selected standalone scope |
| loadjob | Yes | C27 dedicated; U | Selected standalone scope |
| lookup | Yes | C11 dedicated; K | Selected standalone scope |
| makemv | Yes | C32 dedicated; U | Selected standalone scope |
| makeresults | Yes | C26 dedicated; U | Selected standalone scope |
| mstats | Yes | C29 dedicated; U | Selected standalone scope |
| mvcombine | Yes | C34 dedicated; U | Selected standalone scope |
| mvexpand | Yes | C33 dedicated; U | Selected standalone scope |
| nomv | Yes | Deferred grammar/effects; U | Native inventory; not advertised grammar |
| ocsf | No | Profile exclusion | Outside splunkd |
| rename | Yes | C07 dedicated; K | Selected standalone scope |
| replace | Yes | Deferred grammar/effects; U | Native inventory; not advertised grammar |
| reverse | Yes | C15 dedicated; K | Selected standalone scope |
| rex | Yes | C24 dedicated; U | Selected standalone scope |
| route | No | Profile exclusion | Outside splunkd |
| search | Yes | C01 dedicated; K; | Selected standalone scope |
| select | Yes | C02 dedicated; K | Selected standalone scope |
| sort | Yes | C12 dedicated; K | Selected standalone scope |
| spath | Yes | C25 dedicated; U | Selected standalone scope |
| spl1 | Yes | C22 dedicated; U | Selected standalone scope |
| stats | Yes | C08 dedicated; K | Selected standalone scope |
| streamstats | Yes | C10 dedicated; K | Selected standalone scope |
| table | Yes | C06 dedicated; K | Selected standalone scope |
| tags | Yes | Deferred grammar/effects; U | Native inventory; not advertised grammar |
| thru | Yes | Deferred grammar/effects; U | Search sink/pass-through exists; native, deferred |
| timechart | Yes | C30 dedicated; U | Selected standalone scope |
| timewrap | Yes | C31 dedicated; U | Selected standalone scope |
| tstats | Yes | C28 dedicated; U | Selected standalone scope |
| typer | Yes | Deferred grammar/effects; U | Native inventory; not advertised grammar |
| union | Yes | C20 dedicated; U | Selected standalone scope |
| untable | Yes | Deferred grammar/effects; U | Native inventory; not advertised grammar |
| where | Yes | C04 dedicated; K | Selected standalone scope |

Deferred entries preserve an unsupported-syntax limitation; balanced opaque tokens are not syntax conformance. Known profile exclusions use SPL_PROFILE_MISMATCH. Native branch/into/thru are not module errors solely because they write a sink. Branch's individual search section proves standalone search applicability; each branch terminates in into and branch is terminal. Expand's table entry was fetched successfully, but its individual linked page returned a cache miss; no additional grammar claim is made.

## Dedicated command/form cards

The following cards state the required operands, optional operands/options, list delimiters, aliases/placement and field effects. Clauses use documented upper/lower spelling; no inherited global case-insensitive lexer is authorized. Literal strings, identifier arguments, Boolean-only options and command keywords are separate token roles.

### C01 — search

Required search expression: literal term/phrase, field comparison, IN list, time/index selector or Boolean combination. IN uses parenthesized comma-separated literals; CASE/TERM directives have parentheses. Search equality RHS is a literal. NOT > OR > AND > XOR; adjacent search terms imply AND. Command-specific comments restriction applies.

Field semantics: K; ordinary fields/selector roles complete; search-literal subqueries and unproved directives remain U. Source: [doc163](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/search-command/search-command-overview-and-syntax).

```text
C01-P1  search index=main severity=warn
C01-P2  search index=audit status IN (401,403)
C01-N1  search index=main severity=
C01-N2  search index=audit status IN (401,,403)
```

### C02 — from/select

FROM dataset [AS alias] JOIN* WHERE? GROUP BY? SELECT? HAVING? ORDER BY? LIMIT? OFFSET?; SELECT-first moves SELECT before FROM. Dataset always required, SELECT required with GROUP BY. Comma expression lists, optional AS aliases; DISTINCT follows SELECT. Joins require aliases and equality ON. Details and SQL cells below.

Field semantics: K for plain source and supported no-join clauses; joins, uncertain names/visibility U. Source: [doc143](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/from-command/from-command-overview).

```text
C02-P1  FROM requests WHERE size>1 SELECT server
C02-P2  SELECT sum(size) AS total FROM requests GROUP BY server
C02-N1  FROM WHERE size>1 SELECT server
C02-N2  SELECT sum(size) AS total GROUP BY server
```

### C03 — eval

One or more target=expression assignments, comma-separated, evaluated left-to-right. Identifiers and field templates are distinct targets. Quoted reserved targets allowed except narrower documented alias restrictions. Boolean assignments admitted by ruling R01. Exact null removes target.

Field semantics: K core expressions; dynamic targets/lambdas/unmodeled calls U. Read-before-write and conditional availability preserved. Source: [doc136](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/eval-command/eval-command-overview-and-syntax).

```text
C03-P1  eval size_kb=size/1024, rounded=round(size_kb,1)
C03-P2  eval enabled=false, retired=null
C03-N1  eval size_kb=size/1024 rounded=round(size_kb,1)
C03-N2  eval enabled=, retired=null
```

### C04 — where

One required predicate expression; explicit logical separators between predicates. NOT > AND > OR > XOR, unlike search. Equality may use = or ==; field operands on both sides are reads. Predicate family includes BETWEEN, IN, IS, LIKE and EXISTS (L07).

Field semantics: K ordinary core predicates; EXISTS/correlation and unmodeled calls U. Source: [doc178](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/where-command/where-command-overview-syntax-and-usage).

```text
C04-P1  where inbound=outbound AND size>0
C04-P2  where isnotnull(account) OR severity="warn"
C04-N1  where inbound=outbound size>0
C04-N2  where isnotnull(account) OR
```

### C05 — fields

One optional list-wide + or -, followed by a nonempty comma-separated field selector list. Wildcard selectors single-quoted. No repeated sign per item. Known _raw/_time persist by default; prior removals remain removed, open internal membership is not fabricated.

Field semantics: K exact; wildcard completeness depends on M3 finite/partial universe and internal policy. Source: [doc139](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/fields-command/fields-command-overview-syntax-and-usage).

```text
C05-P1  fields + server, account
C05-P2  fields - 'scratch*', debug
C05-N1  fields + server account
C05-N2  fields -
```

### C06 — table

Required nonempty comma-separated fields. No options or AS aliases. Exact identifiers admitted. Wildcard quoting conflicts: table page gives double quotes, general selector rules single; tracked H01 until ruling/evidence. Do not make either quote an unsupported-syntax error by assumption.

Field semantics: K exact; wildcard form held and finite-universe requirements apply. Source: [doc169](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/table-command/table-command-overview-syntax-and-usage).

```text
C06-P1  table server, account
C06-P2  table 'remote-address'
C06-N1  table server account
C06-N2  table server AS machine
```

### C07 — rename

Comma-separated source AS target pairs; exact names or quoted wildcard patterns. AS/as. Newer usage forbids repeated source, repeated target, chained and overlapping/circular pairs. R02 supersedes older sequential example; reserved replacement names and wildcard cardinality need separate validation.

Field semantics: K independent exact pairs; wildcard refinement M3. Conflicts definite invalid, not unknown effects. Source: [doc158](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/rename-command/rename-command-overview-syntax-and-usage).

```text
C07-P1  rename server AS machine, account AS actor
C07-P2  rename 'remote-address' AS remote_addr
C07-N1  rename server AS machine account AS actor
C07-N2  rename server AS
```

### C08 — stats

Search options before aggregates: allnum=bool, delim=string, partitions=num. Nonempty comma-separated aggregate calls with optional AS output; parentheses mandatory, including count(). Optional BY comma-separated exact fields, each optionally span=timespan. BY wildcards disallowed. Pipeline-only mode/prestats/annotations are profile errors.

Field semantics: K core aggregate forms, closes output. allnum=true or unsupported option/form must retain conditional outputs or U; no unconditional availability. Source: [doc167](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/stats-command/stats-command-overview-syntax-and-usage).

```text
C08-P1  stats count(), sum(size) AS total BY server, account
C08-P2  stats delim=";" values(account) AS actors
C08-N1  stats count, sum(size) AS total BY server, account
C08-N2  stats count() sum(size) AS total
```

### C09 — eventstats

Optional allnum=bool precedes nonempty comma-separated aggregates with AS aliases. Optional BY exact comma-separated fields with per-field span. Synopsis stray [bool] is a publication error; named by-clause definition governs. No BY wildcard.

Field semantics: K core forms preserve input/add outputs; allnum=true numerical groups can suppress outputs, therefore conditional or U. Source: [doc137](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/eventstats-command/eventstats-command-overview-syntax-and-usage).

```text
C09-P1  eventstats avg(size) AS typical BY server
C09-P2  eventstats count() AS rows, dc(account) AS actors
C09-N1  eventstats avg(size) AS typical BY
C09-N2  eventstats count() AS rows dc(account) AS actors
```

### C10 — streamstats

Normative synopsis: optional BY list, current=bool, reset clause, window=nonnegative-int all BEFORE comma-separated aggregates. Reset has before expression, after expression, onchange, requiring a condition. Older/post-aggregate layouts are H02, not uncritically inherited SPL. Removed SPL option aliases not advertised.

Field semantics: K core aggregates preserve input; previous/current/reset conditional output proof may require U. Reset operands are reads, outputs use original locations. Source: [doc168](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/streamstats-command/streamstats-command-overview-syntax-and-usage).

```text
C10-P1  streamstats window=4 avg(size) AS moving
C10-P2  streamstats BY server current=true count() AS seen
C10-N1  streamstats window=4
C10-N2  streamstats window=-1 avg(size)
```

### C11 — lookup

Required dataset plus nonempty comma match list, each lookup-field [AS event-field]. Optional uppercase OUTPUT or OUTPUTNEW followed by comma output list with per-item AS. AS/as allowed. local/update removed. Catalog column names and local field names have distinct roles.

Field semantics: K explicit outputs with OUTPUT overwrite/OUTPUTNEW conditional semantics; omitted output list unknown catalog shape U. Source: [doc150](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/lookup-command/lookup-command-overview-syntax-and-usage).

```text
C11-P1  lookup identities uid AS account OUTPUT display AS person, team
C11-P2  lookup regions code OUTPUTNEW country
C11-N1  lookup identities uid AS account output display AS person
C11-N2  lookup regions code OUTPUTNEW
```

### C12 — sort

Optional positional integer count before nonempty comma sort terms. Terms may have +/-, and auto/ip/num/str(field) wrappers. Count 0 means all. No count= option. Parentheses belong to sort wrappers; signed terms apply independently.

Field semantics: K field-preserving reads, no execution/collation evaluation. Source: [doc164](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/sort-command/sort-command-overview-syntax-and-usage).

```text
C12-P1  sort 0 -num(size), +server
C12-P2  sort ip(remote_addr), str(account)
C12-N1  sort 0 -num(size) +server
C12-N2  sort ip(), str(account)
```

### C13 — dedup

Optional positive positional count, keepempty=bool, consecutive=bool BEFORE nonempty comma field list. Removed keepevents and sortby. No arbitrary options after fields.

Field semantics: K reads fields, preserves field set; value loss is not field deletion. Source: [doc135](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/dedup-command/dedup-command-overview-syntax-and-usage).

```text
C13-P1  dedup 2 keepempty=true server, account
C13-P2  dedup consecutive=true request_id
C13-N1  dedup 2 keepempty=true server account
C13-N2  dedup 0 request_id
```

### C14 — head

Optional keeplast=bool, then optional while(parenthesized predicate), then optional positional integer; no operands defaults 10. No limit= or null=. Synopsis order admitted; contradictory count-before-while is H03. Statistical functions not allowed in while.

Field semantics: K ordinary head/while core reads; uncertain predicate function U. Source: [doc144](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/head-command/head-command-overview-syntax-and-usage).

```text
C14-P1  head 7
C14-P2  head keeplast=false while (size<100) 9
C14-N1  head limit=7
C14-N2  head keeplast=false while size<100
```

### C15 — reverse

No arguments or options; reverses rows while preserving field values and multivalue order.

Field semantics: K field-preserving, no reads. Source: [doc160](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/reverse-command/reverse-command-overview-syntax-and-usage).

```text
C15-P1  FROM main | reverse
C15-P2  FROM [{a:1},{a:2}] | reverse | head 1
C15-N1  FROM main | reverse a
C15-N2  FROM main | reverse limit=2
```

### C16 — join

Optional type=inner|outer|left and max=int with required left=alias right=alias before required where equality list. Alias-qualified equality joined with AND. Literal bracketed right subsearch begins FROM. Synopsis/options examples allow options amongst pre-where aliases; max=0 unlimited, outer means left outer. Removed usetime/earlier/overwrite.

Field semantics: U merge; independent right scope, real predicate reads and dependency roles. Source: [doc148](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/join-command/join-command-overview-syntax-and-usage).

```text
C16-P1  join left=L right=R where L.id=R.id [FROM identities]
C16-P2  join type=left max=0 left=L right=R where R.id=L.id AND L.region=R.region [FROM identities]
C16-N1  join left=L right=R where L.id=R.id FROM identities
C16-N2  join type=left max=0 left=L where L.id=R.id [FROM identities]
```

### C17 — append

One literal bracketed independent subsearch, no options. Removed extendtimerange/maxtime/maxout/timeout. Child not inherited; historical-use limitation not static grammar restriction.

Field semantics: U parent/child merge, dependencies and child local lineage retained. Source: [doc129](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/append-command/append-command-overview-syntax-and-usage).

```text
C17-P1  append [FROM archive]
C17-P2  append [search index=audit | stats count()]
C17-N1  append FROM archive
C17-N2  append [search index=audit | stats count()
```

### C18 — appendpipe

Optional run_in_preview=bool before required literal bracketed nonempty subpipe. Child executes at this point with inherited parent environment. Transforming-command placement requires contextual checks, not merely bracket recognition.

Field semantics: U merge, inherited copied child state. Source: [doc130](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/appendpipe-command/appendpipe-command-overview-syntax-and-usage).

```text
C18-P1  appendpipe [stats count()]
C18-P2  appendpipe run_in_preview=false [where size>0 | eval tag="large"]
C18-N1  appendpipe []
C18-N2  appendpipe run_in_preview=1 [stats count()]
```

### C19 — appendcols

One literal bracketed independent subsearch. No override/maxtime/maxout/timeout. If transforming operations appear, command follows them. Main collision values win; internal child fields excluded.

Field semantics: U row alignment/merge. Independent child scope. Source: [doc128](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/appendcols-command/appendcols-command-overview-syntax-and-usage).

```text
C19-P1  appendcols [FROM identities | fields account]
C19-P2  appendcols [search index=audit | stats count()]
C19-N1  appendcols FROM identities
C19-N2  appendcols override=true [FROM identities]
```

### C20 — union

Generating union requires at least two comma-separated datasets; piped union requires at least one. Dataset can plain name, literal row list or bracketed independent subsearch. Dataset-kind qualification is separate from module namespace and must be explicitly admitted before complete syntax claim.

Field semantics: U merge; child scope/dependencies retained. Source: [doc176](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/union-command/union-command-overview-syntax-and-usage).

```text
C20-P1  union main, archive
C20-P2  FROM main | union [FROM audit | fields account], [{account:"local"}]
C20-N1  union main
C20-N2  FROM main | union archive audit
```

### C21 — if

Required if (predicate) [nonempty subpipe]; zero+ elseif (predicate) [subpipe], optional final else [subpipe]. Literal parentheses and brackets. No pipeline separator before elseif/else; each internal subpipe may have pipes. Inline if() is a separate expression.

Field semantics: U conditional merge; inherited child scopes and predicate reads sound. Source: [doc145](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/if-command/if-command-overview-syntax-and-usage).

```text
C21-P1  if (size>50) [eval tier="large"] else [eval tier="small"]
C21-P2  if (code=500) [eval severity=3] elseif (code=404) [eval severity=1]
C21-N1  if size>50 [eval tier="large"]
C21-N2  if (code=500) [eval severity=3] elseif
```

### C22 — spl1

Wrapper is spl1 double-quoted SPL source or a command-position backtick block. Embedded body must use explicit search unless generating; macros/subsearches forbidden by source. Full embedded conformance requires its own grammar-established body check, never regex/rewrite. Wrapper-only success reports body syntax unproved.

Field semantics: U embedded syntax/body/effects until separately validated; never claim valid on wrapper alone. Source: [doc166](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/spl1-command/spl1-command-overview-syntax-and-usage).

```text
C22-P1  spl1 "search index=legacy | head 4"
C22-P2  `search index=archive | stats count`
C22-N1  spl1 search index=legacy
C22-N2  `search index=archive
```

### C23 — bin

Options before required field, optional AS output. bins=int, minspan=span-length, span=length|logspan, start/end=num, aligntime=earliest|latest|time-specifier. Span length int[unit]; logspan coefficient/base constraints and timescale aliases listed below. No field-first options.

Field semantics: U currently; retains input read and alias-write candidate without definite field effect. Source: [doc131](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/bin-command/bin-command-overview-syntax-and-usage).

```text
C23-P1  bin span=15m _time AS slot
C23-P2  bin bins=8 start=0 end=100 size
C23-N1  bin _time span=15m
C23-N2  bin bins=8 start=0 end=100
```

### C24 — rex

Options field=field,max_match=int,offset_field=field precede regex argument, or mode=sed plus sed string. Regex may quoted, raw or slash literal per usage. Default input _raw; max_match0 unlimited. Wrapper parsing is not PCRE validation. Ignore Edge/Ingest PCRE2 migration note for splunkd promise.

Field semantics: U extraction/regex effects, reliable input read only; no guessed named-capture outputs. Source: [doc161](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/rex-command/rex-command-overview-syntax-and-usage).

```text
C24-P1  rex field=message @"(?<digits>\d+)"
C24-P2  rex mode=sed "s/secret/hidden/g"
C24-N1  rex @"(?<digits>\d+)" field=message
C24-N2  rex mode=sed
```

### C25 — spath

All operands optional: input=field,output=field,path=quoted-path. output requires path. Without path, autoextract. Dot/curly-index path is a string, not ordinary field-expression reads. Default input _raw; output defaults to path name when path explicit.

Field semantics: U extraction, source input trustworthy; implicit _raw never invented span. Source: [doc165](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/spath-command/spath-command-overview-syntax-and-usage).

```text
C25-P1  spath
C25-P2  spath input=payload output=user_name path="actor.name"
C25-N1  spath output=user_name
C25-N2  spath input=payload output=user_name path=actor.name
```

### C26 — makeresults

Optional positional integer, defaults 1; no count=, annotate/format/data options advertised in SPL2. Required-arguments None governs unbracketed synopsis count.

Field semantics: U generated effect in initial milestone; eventual known _time support requires explicit model. Source: [doc152](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/makeresults-command/makeresults-command-overview-syntax-and-usage).

```text
C26-P1  makeresults
C26-P2  makeresults 3
C26-N1  makeresults count=3
C26-N2  makeresults 3 4
```

### C27 — loadjob

Exactly one positional job identifier; generating, must begin search. Savedsearch loading and result_event/delegate/artifact_offset/ignore_running options removed.

Field semantics: U external job data; not an index named by sid. Source: [doc149](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/loadjob-command/loadjob-command-overview-syntax-and-usage).

```text
C27-P1  loadjob 1780000000.42
C27-P2  loadjob 1780000500.7 | fields account
C27-N1  loadjob
C27-N2  FROM main | loadjob 1780000000.42
```

### C28 — tstats

Required aggregates=[comma calls]. Optional datamodel_name='Model.Root', predicate=(expression), byfields=[comma exact fields]. Brackets/parentheses literal; wildcard byfields forbidden. Removed prestats/local/append/summariesonly/include_reduced_buckets/allow_old_summaries/chunk_size/fillnull_value.

Field semantics: U metric/data-model semantics; preserve selector/dependency and explicit field roles. Source: [doc174](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/tstats-command/tstats-command-overview-syntax-usage).

```text
C28-P1  tstats aggregates=[count()]
C28-P2  tstats aggregates=[sum(bytes)] datamodel_name='Traffic.All' predicate=(port=443) byfields=[host,source]
C28-N1  tstats count()
C28-N2  tstats aggregates=[sum(bytes)] byfields=[host source]
```

### C29 — mstats

Required aggregates=[comma calls]; optional predicate=(expression), byfields=[comma exact fields]. Dotted metric names single-quoted. No wildcard grouping. Removed append/backfill/chart/chart-options/chunk_size/fillnull_value/prestats/span-length/update_period.

Field semantics: U metric semantics; source selectors and dimensions distinct. Source: [doc153](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/mstats-command/mstats-command-overview-syntax-and-usage).

```text
C29-P1  mstats aggregates=[avg('cpu.load')]
C29-P2  mstats aggregates=[max('disk.used'),min('disk.used')] predicate=(index="metrics") byfields=[host]
C29-N1  mstats aggregates=avg('cpu.load')
C29-N2  mstats aggregates=[avg('cpu.load')] byfields=['host*']
```

### C30 — timechart

Required single aggregate OR parenthesized expression with required BY split field. Prefix options sep/format strings, partial/cont/fixedrange bool, limit=int, agg=(aggregate [AS alias]), bin options. BY accepts exactly one split field with suffix bin/usenull/useother/nullstr/otherstr options. Multiple aggregation delimiter is H04. Timespans below.

Field semantics: U dynamic series/output effects, original aggregate and split reads retained. Source: [doc172](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/timechart-command/timechart-command-overview-and-syntax).

```text
C30-P1  timechart span=10m count()
C30-P2  timechart cont=false avg(size) BY server usenull=false
C30-N1  timechart span=10m
C30-N2  timechart avg(size) BY server, account
```

### C31 — timewrap

Required [integer]timescale then optional align=now|end (command-specific exception to generic option-first). Must follow timechart. No series/time_format. Here m means MONTHS; min means minutes. Accepted unit aliases below.

Field semantics: U dynamically named periods, no fabricated outputs. Source: [doc173](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/timewrap-command/timewrap-command-overview-syntax-and-usage).

```text
C31-P1  FROM main | timechart count() | timewrap 2day
C31-P2  FROM main | timechart avg(size) | timewrap week align=now
C31-N1  FROM main | timechart count() | timewrap
C31-N2  FROM main | timechart count() | timewrap day align=start
```

### C32 — makemv

Optional delim=string and tokenizer=regex-string before exactly one field; no evidence that both options are syntactically exclusive. Defaults space delimiter. No internal fields; allowempty/setsv removed.

Field semantics: U cardinality/value effects until modeled; input role retained. Source: [doc151](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/makemv-command/makemv-command-overview-syntax-and-usage).

```text
C32-P1  makemv delim=":" labels
C32-P2  makemv tokenizer="([a-z]+)" tokens
C32-N1  makemv delim=":"
C32-N2  makemv allowempty=true labels
```

### C33 — mvexpand

Optional limit=int BEFORE exactly one field; 0 unlimited. Multiple fields illegal. Command-specific placement strict.

Field semantics: U initial semantic boundary; retains field read, no merge invention. Source: [doc155](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/mvexpand-command/mvexpand-command-overview-syntax-and-usage).

```text
C33-P1  mvexpand labels
C33-P2  mvexpand limit=2 tokens
C33-N1  mvexpand labels limit=2
C33-N2  mvexpand labels, tokens
```

### C34 — mvcombine

Optional delim=string before exactly one field. Does not apply to internal fields; string delimiter default space.

Field semantics: U row/value combination; one input field read. Source: [doc154](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/mvcombine-command/mvcombine-command-overview-syntax-and-usage).

```text
C34-P1  mvcombine account
C34-P2  mvcombine delim=";" labels
C34-N1  mvcombine
C34-N2  mvcombine delim=";" labels, account
```

### C35 — fillnull

No required arguments. Optional value=DOUBLE-QUOTED-string before optional comma field list; default value 0. Missing named field may be created (value-dependent effects), not safely a pure preservation command.

Field semantics: U output membership/conditional creation. Source: [doc141](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/fillnull-command/fillnull-command-overview-syntax-and-usage).

```text
C35-P1  fillnull
C35-P2  fillnull value="unknown" account, region
C35-N1  fillnull value=unknown account
C35-N2  fillnull value="unknown" account region
```

## Language-family form matrix

### L01 — starts

One standalone search; generating command or explicit search, with only index-first omitted search admitted. Optional leading pipe must be separately evidenced; seed queries omit it. Module statements not inferred from standalone text.

Field semantics: K by selected command; unsupported entry syntax U, excluded module/profile invalid. Source: [start](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-manual/getting-started/quick-start-write-and-run-a-basic-spl2-search/start-searching-data-using-spl2).

```text
L01-P1  index=telemetry warning
L01-P2  search 502 index=proxy
L01-N1  502 index=proxy
L01-N2  SELECT account
```

### L02 — identifiers

Bare simple identifiers; single-quoted names for punctuation/Unicode/reserved words. Double-quoted tokens are string values except command-specific literal parameters. Literal dotted identifier differs from path.

Field semantics: K exact reference spelling/slices. Source: [syntax](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/introduction/understanding-spl2-syntax).

```text
L02-P1  FROM main | eval 'métrique'=size, n='métrique'
L02-P2  FROM main | fields 'group', 'peer.addr'
L02-N1  FROM main | eval 'métrique=size
L02-N2  FROM main | fields group
```

### L03 — literals

Lowercase true/false; null literal; signed int/long L/float F/double D, decimals and exponents; double strings/raw @strings. Type-invalid bool alternatives tested in bool-only options, not as ordinary identifiers.

Field semantics: K literal-only assignments where supported; null deletes. Runtime number overflow is not automatic field failure. Source: [types](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-manual/data-types/built-in-data-types).

```text
L03-P1  FROM main | eval long_n=91L, fraction=2.5F, precise=4D
L03-P2  FROM main | eval path=@"C:\logs", text="a\"b"
L03-N1  FROM main | streamstats current=TRUE count()
L03-N2  FROM main | eval n=2.5e+
```

### L04 — comments

// through newline and /*...*/ outside search-command portions. Block comments inside search portions expressly prohibited. Strings/regex comments inert, preserve UTF-8 and CRLF source accounting.

Field semantics: K no invented reads/stages from comments. Source: [comments](https://help.splunk.com/en/splunk-cloud-platform/search/spl2-search-manual/comments/using-comments-in-spl2).

```text
L04-P1  FROM main /* pipeline | is inert */ | fields account
L04-P2  FROM main // ignored | text
| eval note="/*literal*/"
L04-N1  FROM main /* never closed
L04-N2  search index=main /* blocked */ account=guest
```

### L05 — search predicates

Literal keyword, field comparison, IN literals (comma list), TERM/CASE directives, time selectors, adjacency AND. RHS bare identifiers literal; no where-style read-right inference.

Field semantics: K ordinary filters; verify Boolean tree differs from where. Source: [doc163](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/search-command/search-command-overview-and-syntax).

```text
L05-P1  search index=web user=guest OR action=read status=200
L05-P2  search index=logs TERM(192.0.2.9) earliest=-2h@h latest=@h
L05-N1  search index=web status IN (200 204)
L05-N2  search index=logs TERM()
```

### L06 — expression precedence

Unary +/-; arithmetic * / % before +/-, explicit parentheses; + also string concatenation. Dot is path access, not concatenation. Function calls comma-delimited. Tree assertions required, not just discovered names.

Field semantics: K core expressions; type/value evaluation excluded. Source: [expressions](https://help.splunk.com/en/splunk-cloud-platform/search/spl2-search-manual/expressions-and-predicates/types-of-expressions).

```text
L06-P1  FROM main | eval calc=-size+rate*2, label=account+":"+server
L06-P2  FROM main | eval calc=(size+rate)/2, rest=size%3
L06-N1  FROM main | eval calc=-size+
L06-N2  FROM main | eval calc=(size+rate/2
```

### L07 — conditional predicates

Comparison =/==/!=/</<=/>/>=, [NOT] BETWEEN low AND high, [NOT] IN (...), [NOT] LIKE pattern, IS [NOT] NULL, IS type or NOT(expr IS type), EXISTS(parenthesized correlated search). IS NOT type is an unproved extra layout held as EH04 in the form expansion, not a definite negative. Operator case is per documented subfamily; BETWEEN lower allowed despite generic uppercase advice. EXISTS is restricted to SQL WHERE/HAVING, not the pipe where command; see Q07.

Field semantics: K ordinary expressions; correlated EXISTS U with bound aliases; don't synthesize source fields from type names. Source: [predicate](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-manual/expressions-and-predicates/predicate-expressions).

```text
L07-P1  FROM main | where size NOT BETWEEN 2 AND 8
L07-P2  FROM main | where account IS NOT NULL AND server LIKE "api%"
L07-N1  FROM main | where size BETWEEN 2
L07-N2  FROM main | where account IS NOT
```

### L08 — arrays/objects

Array comma expressions; object comma key:expression entries, keys bare or single/double quoted. Empty container/trailing object-comma details need source coverage; trailing array comma H05. Nested delimiters parsed structurally.

Field semantics: K statically known objects/arrays only as M4 supports; keys not source reads, conditional missing rows retained. Source: [objects](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-manual/expressions-and-predicates/array-and-object-literals-in-expressions).

```text
L08-P1  FROM main | eval pair=[account,size], obj={owner:account,n:size}
L08-P2  FROM main | eval obj={"two words":server,'dash-key':[1,2]}
L08-N1  FROM main | eval obj={owner account}
L08-N2  FROM main | eval pair=[account,,size]
```

### L09 — access paths

Dot object keys (reserved keys single-quoted), bracket integer/string/expression indexes; arbitrary chained accesses. Bracket double string key and bracket field index differ. Dot after a number may be numeric token boundary, not field name.

Field semantics: K static paths only verified M4 contract; dynamic indexes U retaining base/index reads. Source: [access](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-manual/expressions-and-predicates/access-expressions-for-arrays-and-objects).

```text
L09-P1  FROM main | eval owner=actor.user.name
L09-P2  FROM main | eval selected=items[position]["name"]
L09-N1  FROM main | eval owner=actor.
L09-N2  FROM main | eval selected=items[]
```

### L10 — templates

Double-quoted string templates and single-quoted field templates contain ${expression}; nested quote contexts supported through lexer modes. Field-template target computes a name, never an exact source token reference.

Field semantics: K ordinary supported string-template reads; dynamic field names U. Source: [templates](https://help.splunk.com/en/splunk-cloud-platform/search/spl2-search-manual/expressions-and-predicates/string-templates-in-expressions).

```text
L10-P1  FROM main | eval message="owner ${account}, bytes ${size}"
L10-P2  FROM main | eval '${category}'=amount
L10-N1  FROM main | eval message="owner ${}"
L10-N2  FROM main | eval '${category'=amount
```

### L11 — inline lambdas

Dollar parameters, optional built-in type/default constant. Single param may omit parens, zero/multi require them. -> expression or block with assignment(s) then single last return; semicolon or newline separators. $it shortcut supported contextually. Nested lambdas disallowed, reusable declarations excluded.

Field semantics: U call effects; lexical locals excluded from event reads, body ordinary free reads retained. Source: [lambda](https://help.splunk.com/en/splunk-cloud-platform/search/spl2-search-manual/expressions-and-predicates/lambda-expressions).

```text
L11-P1  FROM main | eval doubled=map(items,$v -> $v*2)
L11-P2  FROM main | eval total=reduce(items,0,($a,$v) -> {$next=$a+$v; return $next})
L11-N1  FROM main | eval doubled=map(items,$v $v*2)
L11-N2  FROM main | eval total=reduce(items,0,($a,$v) -> {$next=$a+$v})
```

### L12 — projection/rename lists

Comma-separated exact fields and aliases; wildcard quoting command-specific. Independent rename pairs only under R02. Fields internal policy preserves only known membership/tombstones; table/SELECT close differently.

Field semantics: K exact and finite resolved wildcard; ambiguous universe U. Source: [doc139](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/fields-command/fields-command-overview-syntax-and-usage).

```text
L12-P1  FROM main | fields server, size | rename server AS machine
L12-P2  FROM [{a:1,b:2}] | rename a AS x, b AS y | table x,y
L12-N1  FROM main | rename server AS machine, server AS node
L12-N2  FROM main | rename a AS b, b AS c
```

### L13 — aggregation/window forms

Parenthesized native calls, per-call AS, comma aggregate and group lists. Native names count, sum(size), dc(account). Current=false and allnum=true conditional writes retained. Use streamstats normative pre-aggregate options.

Field semantics: K core call forms; windows/resets conditional availability must remain honest. Source: [doc168](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/streamstats-command/streamstats-command-overview-syntax-and-usage).

```text
L13-P1  FROM main | stats dc(account), count() BY server
L13-P2  FROM main | streamstats BY server reset before code=500 window=3 sum(size) AS total
L13-N1  FROM main | stats sum(size), BY server
L13-N2  FROM main | streamstats window=3 sum()
```

### L14 — lookup lists

Dataset/cached catalog column roles separate from event roles. OUTPUT/OUTPUTNEW uppercase; missing clause legal but shape U. Commas between aliased mappings and between outputs mandatory.

Field semantics: K explicit well-modeled output effects; omitted clause U. Source: [doc150](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/lookup-command/lookup-command-overview-syntax-and-usage).

```text
L14-P1  FROM main | lookup people key AS account, realm AS domain OUTPUT title AS role
L14-P2  FROM main | lookup codes id AS status OUTPUTNEW label, category
L14-N1  FROM main | lookup people key AS account realm AS domain OUTPUT title
L14-N2  FROM main | lookup codes id AS status OUTPUT label category
```

### L15 — FROM-first SQL

FROM then optional JOIN/WHERE/GROUP BY/SELECT/HAVING/ORDER BY/LIMIT/OFFSET in fixed hierarchy. SELECT may omitted if no group. Source dependency established before WHERE read, source references original clause.

Field semantics: K approved no-join subset; visibility U retained. Source: [fromusage](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/from-command/from-command-usage).

```text
L15-P1  FROM sales WHERE price>0 GROUP BY region SELECT region,sum(price) AS total
L15-P2  FROM events ORDER BY size DESC LIMIT 12 OFFSET 3
L15-N1  FROM sales SELECT region GROUP BY region
L15-N2  FROM events ORDER BY size DESC WHERE size>0
```

### L16 — SELECT-first SQL

SELECT [DISTINCT] expression list then required FROM, optional JOIN/WHERE/GROUP BY/HAVING/ORDER BY/LIMIT/OFFSET. AS aliases optional for field/core aggregates, required for complete uncertain expression label.

Field semantics: K flow FROM/WHERE/group/aggregate/HAVING/order/project; SELECT-owned phase metadata. Source: [fromsyntax](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/from-command/from-command-syntax).

```text
L16-P1  SELECT DISTINCT region FROM sales
L16-P2  SELECT region,max(price) AS peak FROM sales GROUP BY region HAVING peak>9
L16-N1  SELECT DISTINCT FROM sales
L16-N2  SELECT region,max(price) AS peak FROM sales HAVING peak>9 GROUP BY region
```

### L17 — SQL followed by pipes

Complete SQL unit may feed native commands. Pipe is actual command boundary, never part of SQL clause order. SQL clauses cannot resume after another pipeline command.

Field semantics: K SQL final projection before next command; source spans original. Source: [process](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-manual/getting-started/quick-start-write-and-run-a-basic-spl2-search/process-your-search-results).

```text
L17-P1  SELECT account FROM main WHERE size>0 | eval actor=account | table actor
L17-P2  FROM main GROUP BY server SELECT server,count() AS n | where n>2
L17-N1  SELECT account FROM main | eval actor=account WHERE size>0
L17-N2  FROM main SELECT server | LIMIT 2
```

### L18 — dataset literals

Array of object rows as source, including quoted keys; no external dataset dependency. Scalar expression array does not become row dataset. Missing row keys conditional, not external obligations.

Field semantics: K finite source universe; nested shapes per verified M4. Source: [fromusage](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/from-command/from-command-usage).

```text
L18-P1  FROM [{region:"east",n:1},{region:"west",n:2}] SELECT region
L18-P2  SELECT item FROM [{item:"a"},{other:2}]
L18-N1  FROM [1,2] SELECT item
L18-N2  FROM [{item:"a"} {item:"b"}] SELECT item
```

### L19 — SQL joins

JOIN/INNER JOIN/LEFT JOIN/LEFT OUTER JOIN; both datasets require AS aliases, ON cross-alias equality and optional AND equalities. Multiple joins allowed. RIGHT/FULL outside advertised syntax. Qualifiers not retained as output name prefix per usage.

Field semantics: U merge; retain scope/alias evidence, no fake L.id root fields. Source: [fromsyntax](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/from-command/from-command-syntax).

```text
L19-P1  FROM purchases AS p JOIN inventory AS i ON p.item=i.id SELECT p.item
L19-P2  SELECT p.item FROM purchases AS p LEFT OUTER JOIN inventory AS i ON p.item=i.id AND p.site=i.site
L19-N1  FROM purchases JOIN inventory AS i ON p.item=i.id SELECT p.item
L19-N2  SELECT p.item FROM purchases AS p RIGHT JOIN inventory AS i ON p.item=i.id
```

### L20 — pipeline joins/subsearches

Join right child is independent FROM subquery; append child independent, appendpipe child inherited. Literal brackets are syntax, not optional notation. Output merge unknown doesn't erase known input lineage.

Field semantics: U scopes and merges. Source: [doc148](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/join-command/join-command-overview-syntax-and-usage).

```text
L20-P1  FROM main | join left=A right=B where A.uid=B.id [FROM directory | fields id]
L20-P2  FROM main | appendpipe [eval scratch=account | fields scratch]
L20-N1  FROM main | join left=A right=B where A.uid=B.id []
L20-N2  FROM main | appendpipe [eval scratch=account |]
```

### L21 — union

Plain/literal/subsearch datasets, comma-separated; standalone>=2, piped>=1. Child SQL scopes don't inherit parent fields. Multiple SQL child trees counted individually in SQL structural assertions.

Field semantics: U merge, trustworthy independent scopes. Source: [doc176](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/union-command/union-command-overview-syntax-and-usage).

```text
L21-P1  union [FROM live SELECT account], [SELECT account FROM archive]
L21-P2  FROM [{n:1}] | union [{n:2}], [{n:3}]
L21-N1  union [FROM live SELECT account] [SELECT account FROM archive]
L21-N2  FROM [{n:1}] | union
```

### L22 — metrics/data models

Named aggregates/byfields lists and parenthesized predicate; data-model root argument distinct from event field. Metric dotted names single-quoted. Legacy bare SPL options never guessed into new named syntax.

Field semantics: U selector/data-model/metric roles and unsupported effects explicit. Source: [doc174](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/tstats-command/tstats-command-overview-syntax-usage).

```text
L22-P1  tstats aggregates=[count()] datamodel_name='Network.Flow' byfields=[host]
L22-P2  mstats aggregates=[sum('request.count')] predicate=(index="telemetry")
L22-N1  tstats aggregates=[count()] predicate=index="main"
L22-N2  mstats aggregates=[sum('request.count')] byfields=host
```

### L23 — extraction/options

Command-specific option placement; quoted/raw/slash regex wrappers; spath quoted location string with output dependency; bin/timechart spans typed context. No parsing of pipes inside strings as command boundaries.

Field semantics: U extraction/regex language unless dedicated support added. Source: [doc161](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/rex-command/rex-command-overview-syntax-and-usage).

```text
L23-P1  FROM main | rex field=payload /(?<tag>\w+)/
L23-P2  FROM main | spath input=payload path="items{0}.name" output=first_name
L23-N1  FROM main | rex field=payload /(?<tag>\w+)
L23-N2  FROM main | spath input=payload output=first_name
```

### L24 — embedded SPL

Backtick command wrapper versus expression search literal; explicit double-string wrapper. Macro/subsearch restrictions need grammar validation of embedded body. Balanced outer wrapper alone is not proved SPL syntax.

Field semantics: U body and effects; known forbidden macro/subsearch invalid once proved. Source: [doc166](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/spl1-command/spl1-command-overview-syntax-and-usage).

```text
L24-P1  FROM main | spl1 "stats count by host"
L24-P2  FROM main | `stats sum(bytes) AS total`
L24-N1  FROM main | spl1 "stats count by host
L24-N2  FROM main | spl1
```

### L25 — conditional paths

if/elseif/else paths use parentheses/bracketed nonempty subpipes; else final and once. Distinguish command path from scalar if(condition,a,b). Newer Boolean assignment evidence applies.

Field semantics: U merge, condition and local-scope reads retained. Source: [doc145](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/if-command/if-command-overview-syntax-and-usage).

```text
L25-P1  FROM main | if (flag=true) [eval a=1] elseif (size>0) [eval a=2] else [eval a=3]
L25-P2  FROM main | if (account="service") [where enabled=true | eval checked=true]
L25-N1  FROM main | if (flag=true) [eval a=1] else [eval a=2] else [eval a=3]
L25-N2  FROM main | if (account="service") []
```

## Additional SQL phase/form matrix

These 32 queries supplement C02 and L15–L19; those named cards alone yield **56 explicitly SQL-focused cases**, without counting incidental FROM prefixes on pipeline tests. Q06 has accepted grammar and explicitly incomplete visibility semantics. No case relies on generic SQL assumptions.

### Q01 — group keys and aggregate phases

Group input/aggregate arguments survive until calculation; HAVING sees selected aggregate outputs; final projection SELECT-owned.

Field semantics: K with explicit phase/execution-order; N cases remove required SELECT structure. Mixed nongrouped aggregate projections are H08 due contradictory examples. Source: [fromsyntax](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/from-command/from-command-syntax).

```text
Q01-P1  SELECT region,sum(cost) AS total FROM charges GROUP BY region HAVING total>3
Q01-P2  FROM charges GROUP BY region SELECT region,count() HAVING count>2
Q01-N1  FROM charges GROUP BY region
Q01-N2  FROM charges GROUP BY region SELECT
```

### Q02 — DISTINCT and projected aliases

DISTINCT precedes expression list; source alias and output alias separate syntax roles.

Field semantics: K no-join exact names; lexical SELECT remains source-first stage. Source: [fromsyntax](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/from-command/from-command-syntax).

```text
Q02-P1  SELECT DISTINCT account,region FROM main
Q02-P2  FROM main SELECT lower(account) AS normalized ORDER BY normalized
Q02-N1  SELECT account DISTINCT FROM main
Q02-N2  FROM main SELECT lower(account) AS ORDER BY normalized
```

### Q03 — integer LIMIT/OFFSET and ordering

LIMIT then OFFSET integer operands; either may occur without the other, no duplicate clauses. Value-range evidence tracked H06.

Field semantics: K preserves selected fields, no hidden reads. Source: [fromsyntax](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/from-command/from-command-syntax).

```text
Q03-P1  SELECT account FROM main LIMIT 5 OFFSET 2
Q03-P2  FROM main OFFSET 4
Q03-N1  SELECT account FROM main LIMIT 5 OFFSET
Q03-N2  FROM main OFFSET 4 LIMIT 5
```

### Q04 — clause aliases and comma lists

GROUPBY/ORDERBY aliases; expression lists require commas, permitted all-upper/all-lower clauses. Per-term sort directions beyond synopsis H07.

Field semantics: K exact group/core calls. Source: [fromsyntax](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/from-command/from-command-syntax).

```text
Q04-P1  SELECT region,count() AS n FROM main GROUPBY region ORDERBY n DESC
Q04-P2  FROM main GROUP BY region,account SELECT region,account,sum(size) AS total
Q04-N1  SELECT region,count() AS n FROM main Group By region
Q04-N2  FROM main GROUP BY region account SELECT region,account,sum(size) AS total
```

### Q05 — group span forms

span(field),span(field,int?unit) and field span=(int?unit); timestamps required. Exact output labeling/shape can be U.

Field semantics: U when span output semantics not proven; cannot pretend ordinary field grouping. Source: [fromsyntax](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/from-command/from-command-syntax).

```text
Q05-P1  FROM main GROUP BY span(_time,5min) SELECT count()
Q05-P2  SELECT count() FROM main GROUP BY _time span=(hour)
Q05-N1  FROM main GROUP BY span(_time,) SELECT count()
Q05-N2  SELECT count() FROM main GROUP BY _time span=()
```

### Q06 — hidden visibility remains incomplete

Parse hidden group key/unselected aggregate references; retain sound grouping/aggregate proof, don't prematurely close environment. P expectations U; negatives below syntax only.

Field semantics: U hidden HAVING/ORDER visibility as approved. Source: [fromsyntax](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/from-command/from-command-syntax).

```text
Q06-P1  SELECT sum(size) AS total FROM main GROUP BY server HAVING server="api"
Q06-P2  FROM main GROUP BY server SELECT sum(size) AS total ORDER BY server
Q06-N1  SELECT sum(size) AS total FROM main GROUP BY server HAVING
Q06-N2  FROM main GROUP BY server SELECT sum(size) AS total ORDER BY
```

### Q07 — correlated EXISTS restricted SQL contexts

Outer AS required; parenthesized child WHERE correlates using equality. Child LIMIT/OFFSET forbidden; EXISTS not allowed in pipe where; with both WHERE/HAVING, EXISTS belongs HAVING.

Field semantics: U correlated binding, no root alias-prefix fabrication. N are definite contextual contract errors. Source: [predicate](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-manual/expressions-and-predicates/predicate-expressions).

```text
Q07-P1  FROM orders AS o WHERE EXISTS (SELECT id FROM inventory WHERE id=o.item) SELECT item
Q07-P2  SELECT item FROM orders AS o WHERE NOT EXISTS (SELECT id FROM inventory WHERE id=o.item AND active=true)
Q07-N1  FROM orders AS o WHERE EXISTS (SELECT id FROM inventory WHERE id=o.item LIMIT 1) SELECT item
Q07-N2  FROM orders AS o | where EXISTS (SELECT id FROM inventory WHERE id=o.item)
```

### Q08 — source scopes and mixed Unicode slices

Dataset name quote is not projected field quote; actual expressions and aliases preserve source byte slices. SQL finishes before next pipe.

Field semantics: K literal source fields and explicit aliases, no external obligations. Source: [fromusage](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/from-command/from-command-usage).

```text
Q08-P1  SELECT 'café' AS label FROM [{'café':"yes"}] | table label
Q08-P2  FROM [{n:1},{n:2}] SELECT n+1 AS next_n | where next_n>1
Q08-N1  SELECT 'café' AS label FROM [{'café':"yes"}] | table
Q08-N2  FROM [{n:1},{n:2}] SELECT n+1 AS next_n | where next_n>
```

## Excluded profile/module and definite contract seeds

All cases below are invalid for the selected toolkit contract. Profile/module outcomes keep both coverage dimensions false and a located explanation. A recognized declaration need not be fully parsed as a reusable module to identify its excluded construct. Other syntax faults may coexist; no fallback to SPL is allowed.

```text
B01  FROM main | decrypt
B02  FROM main | ocsf
B03  FROM main | route
B04  FROM main | stats mode=summary count()
B05  FROM main | stats prestats=raw count()
B06  FROM main | eval id=batch_id()
B07  FROM main | eval clock=batch_time()
B08  $saved = FROM main;
B09  import * from example;
B10  export $saved = FROM main;
B11  function label($x) { return $x }
B12  FROM main; FROM archive
B13  FROM main | rename a AS x, a AS y
B14  FROM main | rename a AS x, b AS x
B15  FROM main | rename alpha AS beta, beta AS gamma
B16  FROM main | rename a AS b, c AS a
```

| IDs | Expected code/outcome | Reason |
| --- | --- | --- |
| B01 | SPL_PROFILE_MISMATCH | Known command belongs to Edge/Ingest, not splunkd. |
| B02 | SPL_PROFILE_MISMATCH | OCSF command != offline OCSF schema validation. |
| B03 | SPL_PROFILE_MISMATCH | Known other-profile command; don't let missing operands hide context finding. |
| B04 | SPL_PROFILE_MISMATCH | stats option is pipeline-only. |
| B05 | SPL_PROFILE_MISMATCH | stats option is pipeline-only. |
| B06 | SPL_PROFILE_MISMATCH | Function inventory separate from command inventory. |
| B07 | SPL_PROFILE_MISMATCH | Known non-splunkd function. |
| B08 | SPL_UNSUPPORTED_MODULE | Named module search. |
| B09 | SPL_UNSUPPORTED_MODULE | Import construct, not ordinary dataset. |
| B10 | SPL_UNSUPPORTED_MODULE | Export/named declaration. |
| B11 | SPL_UNSUPPORTED_MODULE | Reusable function declaration, distinct from inline lambda. |
| B12 | SPL_UNSUPPORTED_MODULE | Multiple top-level statements. |
| B13 | invalid semantic contract | Repeated source R02. |
| B14 | invalid semantic contract | Repeated target R02. |
| B15 | invalid semantic contract | Chain R02. |
| B16 | invalid semantic contract | Overlapping/circular R02. |

## Exact modifier/alias expansion obligations

These are finite syntax axes, not permission for an opaque options map. Each advertised modifier and distinct alternative must ultimately obtain two meaningful positive/minimal-negative pairs, including operand presence/type and placement. Combined examples can share a query only if assertions target every represented axis; fixture count is by distinct source/expectation, not labels. The current seed set intentionally does not pretend every listed modifier has already received all pairs.

| Family | Required expansion axes and boundaries |
| --- | --- |
| search | Explicit and index-first starts; literal/phrase; field relation; IN; TERM/CASE; relative/absolute time; timeformat; index/source/sourcetype/host selectors; search versus where precedence; XOR lowest. Search grammar never accepts arbitrary expression RHS as field reads. |
| eval/expressions | Assignment lists; quoted/dotted identifiers; null deletion; Boolean values; nested calls/operators; primitive suffixes; templates; lambda local scope. Parser admits function calls independently of function semantic registry. |
| fields/table/rename | Include/exclude/default plus; exact and wildcard selectors; distinct table quote dispute H01; missing/comma lists; alias syntax; R02 duplicate/chain/overlap contracts; finite versus open universe. Do not restore tombstoned internals. |
| stats/eventstats | count() versus count(field), field/expression aggregation, explicit/implicit output name, aggregate lists, BY lists/per-field time span, options allnum/delim/partitions where applicable. Wildcard grouping prohibited. allnum=true absence is conditional or incomplete. |
| streamstats | BY pre-list; current bool; reset before/after/onchange combinations; window>=0; calls/aliases. Missing reset condition invalid; alternate placements H02. No reset_on_change/reset_before/reset_after alias copied from SPL. |
| lookup | Single/multiple matches; aliases; OUTPUT versus OUTPUTNEW; explicit and absent output clause; multiple outputs and collisions; uppercase output keyword. No local/update. |
| sort/dedup/head | Signed plain/auto/ip/num/str sort terms and integer count; dedup positive integer/keepempty/consecutive; head absent/count/while/keeplast. Bool is true/false only; limit= and null= removed. Head disputed layout H03. |
| join/branches | Required aliases; either equality orientation and AND lists; type inner/left/outer,max0/positive; independent FROM right child; append/appendcols independent children; appendpipe inherited subpipe and run_in_preview; if/elseif/else inherited paths. No invented combined output universe. |
| bin | bins integer; minspan length; span numeric/time/log; start/end numeric; aligntime earliest/latest/time; AS alias. Coefficient >=1 and <base; base>1. Numeric/domain failure is distinct from delimiter failure. |
| rex/spath | Rex quoted/raw/slash regex and quoted sed s/y forms; max_match0/positive,field,offset_field; options before body. Expand quoted alternation as admitted syntax, for example `rex field=payload "(?<tag>red\|blue)"`; slash-plus-internal-pipe is H12. PCRE wrapper/body separate coverage. Spath input/path/output combinations; output implies path; omitted input default does not fabricate a reference span. |
| generators/metrics | makeresults absent/positional count; loadjob SID and first-command restriction; tstats aggregates/datamodel_name/predicate/byfields; mstats aggregates/predicate/byfields. Named lists nonempty, commas/brackets mandatory, legacy options rejected. |
| timechart | Single aggregate; expression plus BY; prefix sep,format,partial,cont,fixedrange,limit,agg; prefix and split bin options; split usenull,useother,nullstr,otherstr. One split field. Multiple-aggregate syntax H04, not guessed. |
| timewrap | Required unit/optional integer; align now/end after span; preceding timechart required; no series/time_format. Distinct unit vocabulary below. |
| multivalue/fillnull | makemv delim/tokenizer/one noninternal field; mvexpand limit-before-field; mvcombine delim/one noninternal field; fillnull absent/value/string/list. Internal field applicability is a command contract, not identifier grammar. |
| SQL | Both complete hierarchies, optional clauses/dependencies; DISTINCT; AS; GROUPBY/ORDERBY; exact/dynamic expressions; group span alternatives; HAVING/ORDER visibility; LIMIT/OFFSET; aliases/joins; correlated EXISTS restrictions; final projection before pipe. |
| literals/lambdas | Empty/nested containers and key quotes; duplicate-object-key rejection; object comma versus H05 array trailing comma; expression-position search literals distinct from command embeddings; $it shortcut; 0/1/multiple lambda params; typed/default constants; newline/semicolon block separators; return last/once; no nested lambda; no top-level reusable function syntax. |

For bin/timechart spans the current documented aliases are: seconds s/sec/secs/second/seconds; minutes m/min/mins/minute/minutes; hours h/hr/hrs/hour/hours; days d/day/days; weeks w/week/weeks; months mon/month/months; subseconds us/ms/cs/ds. Timechart also documents weekly snap spans with optional sign/integer and @ snap unit; do not accept arbitrary nonweekly snap-span values without evidence. Timewrap separately allows sec aliases above, min/mins/minute/minutes, hour/day/week aliases, months **m/mon/month/months**, quarters qtr/quarter/quarters, years y/yr/year/years. Share concepts, not a mistaken universal m unit. Sources: [doc131](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/bin-command/bin-command-overview-syntax-and-usage), [doc172](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/timechart-command/timechart-command-overview-and-syntax), [doc173](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/timewrap-command/timewrap-command-overview-syntax-and-usage).

## Function semantics are a separate matrix

All required functions below appear as splunkd-supported in [evalcompat](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/spl2-compatibility-profiles-and-quick-references/compatibility-quick-reference-for-spl2-evaluation-functions) or [aggcompat](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/spl2-compatibility-profiles-and-quick-references/compatibility-quick-reference-for-spl2-statistical-functions). Synonym rows denote documented functions, not a generic alias-permission mechanism. Native functions can parse while remaining semantically unmodeled; wrong-profile calls receive the profile code. Field-value/type execution is outside the static model.

| Required function family | Reviewed ordinary positional arity/form | Complete-model boundary / source |
| --- | --- | --- |
| abs; ceil/ceiling; floor | 1 | Ordinary scalar reads; [mathfunc](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/evaluation-functions/mathematical-functions). |
| round | 1 or 2 (optional precision) | Both arities must be registered; [mathfunc](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/evaluation-functions/mathematical-functions). |
| len; lower; upper | 1 | String operand role, no constant result evaluation; [textfunc](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/evaluation-functions/text-functions). |
| trim/ltrim/rtrim | 1 or 2 (optional trim characters) | No implicit event read from omitted argument; [textfunc](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/evaluation-functions/text-functions). |
| substr | 2 or 3 (optional length) | Starting index positional argument; [textfunc](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/evaluation-functions/text-functions). |
| replace | 3 (input, regex, replacement) | Scalar name distinct from replace command; PCRE validity not inferred; [textfunc](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/evaluation-functions/text-functions). |
| coalesce | 1+ | Nullability/conditional availability must remain sound; [evalcompat](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/spl2-compatibility-profiles-and-quick-references/compatibility-quick-reference-for-spl2-evaluation-functions), [comparefunc](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/evaluation-functions/comparison-and-conditional-functions). |
| if; case | if exactly3; case alternating predicate/value pairs | if predicate context separate from if command; missing case fallback can yield null; [comparefunc](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/evaluation-functions/comparison-and-conditional-functions). |
| isnull/isnotnull | 1 | Informational-function usage and per-arity obligations are now recorded in F.isnull/F.isnotnull in the [form expansion](form-expansion.md); final registry execution remains pending. |
| tonumber/tostring | 1 or 2 | Base/format optional; exact vocabulary and positional obligations are now recorded in F.tonumber/F.tostring in the [form expansion](form-expansion.md). Named calls are recognized/incomplete by controller ruling. [evalcompat](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/spl2-compatibility-profiles-and-quick-references/compatibility-quick-reference-for-spl2-evaluation-functions), [convertfunc](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/evaluation-functions/conversion-functions). |
| mvcount; split; match | 1; 2; 2 | mvcount missing field can yield null; split creates multivalue; match regex wrapper != regex validity. [mvfunc](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/evaluation-functions/multivalue-eval-functions), [evalcompat](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/spl2-compatibility-profiles-and-quick-references/compatibility-quick-reference-for-spl2-evaluation-functions). |
| count | 0 or1 | Native count() gives count; count(field) counts field values. c() alias documented but implicit c-output name unproved; do not silently add complete alias support. [aggregatefunc](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/statistical-and-charting-functions/aggregate-functions). |
| sum/avg/min/max; dc/distinct_count | 1 ordinary field/expression | Command contexts and value forms reviewed separately; implicit complex names U. [aggregatefunc](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/statistical-and-charting-functions/aggregate-functions), [aggcompat](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/spl2-compatibility-profiles-and-quick-references/compatibility-quick-reference-for-spl2-statistical-functions). |
| values/list; first/last | 1 ordinary field/expression | Order-sensitive versus distinct values; no runtime ordering computation. [aggcompat](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/spl2-compatibility-profiles-and-quick-references/compatibility-quick-reference-for-spl2-statistical-functions). |

A final production function registry still needs executed positive/minimal-negative tests per admitted arity/form, including explicit versus implicit outputs and field-nullability expectations. The [form expansion](form-expansion.md) now records those design obligations and finite positional signatures; optional/variadic signatures share the correct out-of-range boundaries instead of inventing invalid optional omissions. Named calls have typed recognition but explicitly incomplete semantics. Regex flags, wildcard aggregations and undocumented aliases are not automatically covered by ordinary positional forms. Durable fixtures and their execution remain a gate, not an instruction to reduce the approved ordinary core.

## Conflicts, held forms and review rulings

**R01 — Boolean eval.** Controller approved 2026-09-07: newer [doc145](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/if-command/if-command-overview-syntax-and-usage) examples assign true/false, so the older [doc136](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/eval-command/eval-command-overview-and-syntax) statement excluding Boolean eval results does not create negative fixtures. Boolean-only options still reject numeric/uppercase equivalents per [types](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-manual/data-types/built-in-data-types).

**R02 — rename.** Controller approved 2026-09-07: newer [doc158](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/rename-command/rename-command-overview-syntax-and-usage) forbids duplicate-source, duplicate-target, chain and overlap/circular pairs. They are definite invalid contract findings. Older [renameexamples](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/rename-command/rename-command-examples) left-to-right description does not authorize those pairs. Independent exact pairs remain complete; do not add sequencing abstraction merely for contradicted forms. Select the smallest shared-kernel policy after M4.

**R03 — disputed syntax outcome.** Controller approved 2026-09-07: recognized same-page conflicting layouts remain incomplete. The later dedicated grammar must deliberately retain their recognition and limitation without claiming complete grammar or accidentally treating them as malformed supported syntax. This is narrow disputed-form recognition, not an opaque fallback for arbitrary malformed commands.

| Hold | Evidence conflict / gap | Example candidates (not positive/negative floor entries) | Required outcome |
| --- | --- | --- | --- |
| H01 | Table quotes wildcard with double quotes; generic fields/select syntax uses single. [doc169](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/table-command/table-command-overview-syntax-and-usage), [doc139](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/fields-command/fields-command-overview-syntax-and-usage) | `FROM main \| table "host*"`; `FROM main \| table 'host*'` | Specific unresolved syntax limitation until corroborated/ruling; no blanket rejection from lexer policy. |
| H02 | Streamstats synopsis says all options/BY precede aggregate; same page examples put BY/reset later. [doc168](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/streamstats-command/streamstats-command-overview-syntax-and-usage) | `FROM main \| streamstats sum(size) BY server`; `FROM main \| streamstats count() reset after code=500` | Incomplete disputed layout; don't count as malformed. |
| H03 | Head generic/older examples disagree with command synopsis order. [doc144](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/head-command/head-command-overview-syntax-and-usage), [syntax](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/introduction/understanding-spl2-syntax) | `FROM main \| head 4 while (size>0)`; `FROM main \| head 6 keeplast=false while (size>0)` | Incomplete if recognized disputed placement; no silent normalization. |
| H04 | Timechart says single aggregate in synopsis but discusses multiple; examples also use eval wrapper. [doc172](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/timechart-command/timechart-command-overview-and-syntax), [aggregatefunc](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/statistical-and-charting-functions/aggregate-functions) | `FROM main \| timechart avg(size),max(size)`; `FROM main \| timechart avg(size) max(size)` | Delimiter/alternative unresolved; no positive multiple-agg claim. |
| H05 | Array trailing comma not yet corroborated; object trailing comma evidence must be preserved independently. [objects](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-manual/expressions-and-predicates/array-and-object-literals-in-expressions), [types](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-manual/data-types/built-in-data-types) | `FROM main \| eval a=[1,2,]`; `FROM main \| eval a=[account,]` | Unsupported/incomplete additional form until reviewed; missing middle values remain invalid. |
| H06 | SQL integer syntax clear, exact negative/range limits not fully audited. [fromsyntax](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/from-command/from-command-syntax) | `FROM main LIMIT -1`; `FROM main OFFSET -2` | Do not manufacture runtime-range invalidity; incomplete if unsupported. |
| H07 | ORDER BY synopsis ends with one direction; per-expression direction scope needs explicit evidence. [fromsyntax](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/from-command/from-command-syntax) | `FROM main ORDER BY account ASC,size DESC`; `FROM main ORDER BY account DESC,size ASC` | Keep separate from admitted one trailing direction. |
| H08 | Specific GROUP BY dependencies versus generic examples mixing ungrouped fields/aggregates conflict. [fromsyntax](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/from-command/from-command-syntax), [process](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-manual/getting-started/quick-start-write-and-run-a-basic-spl2-search/process-your-search-results), [predicate](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-manual/expressions-and-predicates/predicate-expressions) | `SELECT account,sum(size) FROM main`; `FROM main GROUP BY region SELECT account,count()` | Unsupported/indeterminate semantic grouping rule; no false available/unavailable conclusion. |
| H09 | Qualified dataset kinds/resources, repeat(), empty dataset literals and full type annotation grammar need finer standalone evidence. [fromusage](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/from-command/from-command-usage), [lambda](https://help.splunk.com/en/splunk-cloud-platform/search/spl2-search-manual/expressions-and-predicates/lambda-expressions) | `FROM repeat({},3)`; `FROM lookup.people` | Incomplete unsupported form unless subsequently admitted. Do not classify resource qualification alone as module. |
| H10 | Embedded SPL syntax cannot be proved from a string wrapper; quoted macro/subsearch content needs grammar analysis. [doc166](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/spl1-command/spl1-command-overview-syntax-and-usage) | `spl1 "search index=main [search index=audit]"` | Proved excluded embedding invalid; otherwise incomplete body syntax/effects, never valid wrapper-only. |
| H11 | Upper/lower logical advice conflicts with newer operator-specific lowercase examples. [syntax](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/introduction/understanding-spl2-syntax), [predicate](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-manual/expressions-and-predicates/predicate-expressions) | lower BETWEEN/AND explicitly documented; arbitrary mixed logical tokens not | Per-operator source rule; don't inherit global case insensitivity or classify unreviewed spelling as definite invalid. |
| H12 | Rex's pipe-character section requires double quotes around regex containing a pipe; its slash-literal example establishes character classes without an internal pipe. No corroborating slash-plus-pipe example was found in the independent review. [doc161](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/rex-command/rex-command-overview-syntax-and-usage), [types](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-manual/data-types/built-in-data-types) | `FROM main \| rex field=payload /(?<tag>red\|blue)/` | Recognized disputed/incomplete syntax, outside positive/negative floors; no definite-invalid claim absent corroboration. |

The new comments restriction is command-context-specific: block comments are not allowed within search-command portions ([comments](https://help.splunk.com/en/splunk-cloud-platform/search/spl2-search-manual/comments/using-comments-in-spl2)). Lexed comments outside those portions remain normal language syntax. The source's prose says implied AND broadly, but command-specific WHERE documentation requires explicit operators; only search receives adjacency AND.

Rex's slash-regex example is explicitly documented in its native command usage. The Edge/Ingest PCRE2 migration note on that page does not establish a splunkd engine version. Regex-body syntax/flags require separate evidence if ever advertised.

Publication metadata varied across retrievals for expression, object/access, into and legacy redirected pages; the source ledger records exactly what this retrieval returned. Product-context sections and content govern; a later timestamp alone does not resolve contradictory rules.

## Adapter and field-model obligations attached to this matrix

Explicit language selection must work through Go Query Documents, CLI canonical analyze/validation, native Python and REST, including mixed batches. No byte sniffing or SPL fallback. Legacy fixed-SPL Go methods stay fixed; legacy selectors reject spl2 with unsupported_dialect_for_operation and canonical-API guidance. Excluded modules/profile use SPL_UNSUPPORTED_MODULE / SPL_PROFILE_MISMATCH as approved.

A P/K seed becomes valid only when the shared M4 kernel supports every effect and source binding. Parent source-universe omissions, lookup unknown columns, unknown functions or scopes cannot become valid because a schema happens to contain similarly named fields. Fields inclusion preserves **known** internals/tombstones; it never creates proof that _raw/_time were present. SQL references/stages keep lexical IDs and original slices; position is logical clause-entry order. Actual lineage can repeat SELECT ownership only with explicit phase and execution-order metadata and coherent states. Final projection is SELECT-owned, not HAVING-owned. Q06 must remain incomplete instead of falsely missing a hidden grouping key.

## Source ledger

All URLs were consulted via read-only web retrieval on **2026-09-07**. “Updated” below is publisher/tool-returned metadata, not toolkit execution proof. Sources shared by cards also supply their exact context; command inventory membership alone never establishes grammar. Missing individual expand retrieval is recorded explicitly.

| Key | Primary reference | Updated metadata |
| --- | --- | --- |
| compat | [Compatibility Quick Reference for SPL2 commands](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/spl2-compatibility-profiles-and-quick-references/compatibility-quick-reference-for-spl2-commands) | 2026-07-14T22:38:02.365Z |
| doc127 | [addinfo command: Overview, syntax, and usage](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/addinfo-command/addinfo-command-overview-syntax-and-usage) | 2026-06-16T03:49:23.226Z |
| doc128 | [appendcols command: Overview, syntax, and usage](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/appendcols-command/appendcols-command-overview-syntax-and-usage) | 2026-06-16T03:49:29.063Z |
| doc129 | [append command: Overview, syntax, and usage](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/append-command/append-command-overview-syntax-and-usage) | 2026-06-16T03:49:15.944Z |
| doc130 | [appendpipe command: Overview, syntax, and usage](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/appendpipe-command/appendpipe-command-overview-syntax-and-usage) | 2026-06-16T03:49:26.243Z |
| doc131 | [bin command: Overview, syntax, and usage](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/bin-command/bin-command-overview-syntax-and-usage) | 2026-06-16T03:49:28.404Z |
| doc132 | [branch command: Overview, syntax, and usage](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/branch-command/branch-command-overview-syntax-and-usage) | 2026-06-16T03:49:25.891Z |
| doc133 | [convert command: Overview, syntax, and usage](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/convert-command/convert-command-overview-syntax-and-usage) | 2026-06-16T03:49:26.930Z |
| doc135 | [dedup command: Overview, syntax, and usage](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/dedup-command/dedup-command-overview-syntax-and-usage) | 2026-06-16T03:49:17.646Z |
| doc136 | [eval command: Overview and syntax](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/eval-command/eval-command-overview-and-syntax) | 2026-06-16T03:48:58.817Z |
| doc137 | [eventstats command: Overview, syntax, and usage](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/eventstats-command/eventstats-command-overview-syntax-and-usage) | 2026-06-16T03:49:26.621Z |
| doc139 | [fields command: Overview, syntax, and usage](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/fields-command/fields-command-overview-syntax-and-usage) | 2026-07-14T22:31:06.977Z |
| doc140 | [fieldsummary command: Overview, syntax, and usage](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/fieldsummary-command/fieldsummary-command-overview-syntax-and-usage) | 2026-06-16T03:49:15.051Z |
| doc141 | [fillnull command: Overview, syntax, and usage](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/fillnull-command/fillnull-command-overview-syntax-and-usage) | 2026-06-16T03:49:34.600Z |
| doc142 | [flatten command: Overview, syntax, and usage](https://help.splunk.com/?resourceId=SCS_SearchReference_flattencommandoverview) | 2026-01-16T00:19:21.629Z |
| doc143 | [from command: Overview](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/from-command/from-command-overview) | 2026-06-16T03:49:25.745Z |
| doc144 | [head command: Overview, syntax, and usage](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/head-command/head-command-overview-syntax-and-usage) | 2026-06-16T03:49:09.722Z |
| doc145 | [if command: Overview, syntax, and usage](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/if-command/if-command-overview-syntax-and-usage) | 2026-07-14T22:25:10.340Z |
| doc146 | [into command: Overview, syntax, and usage](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/into-command/into-command-overview-syntax-and-usage) | 2026-01-16T01:09:50.509Z |
| doc147 | [iplocation command: Overview, syntax, and usage](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/iplocation-command/iplocation-command-overview-syntax-and-usage) | 2026-06-22T06:13:22.727Z |
| doc148 | [join command: Overview, syntax, and usage](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/join-command/join-command-overview-syntax-and-usage) | 2026-06-22T06:10:19.972Z |
| doc149 | [loadjob command: Overview, syntax, and usage](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/loadjob-command/loadjob-command-overview-syntax-and-usage) | 2026-06-20T15:03:23.100Z |
| doc150 | [lookup command: Overview, syntax, and usage](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/lookup-command/lookup-command-overview-syntax-and-usage) | 2026-06-16T03:49:28.119Z |
| doc151 | [makemv command: Overview, syntax, and usage](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/makemv-command/makemv-command-overview-syntax-and-usage) | 2026-06-19T04:43:45.874Z |
| doc152 | [makeresults command: Overview, syntax, and usage](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/makeresults-command/makeresults-command-overview-syntax-and-usage) | 2026-06-19T04:41:34.012Z |
| doc153 | [mstats command: Overview, syntax, and usage](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/mstats-command/mstats-command-overview-syntax-and-usage) | 2026-06-19T04:36:55.476Z |
| doc154 | [mvcombine command: Overview, syntax, and usage](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/mvcombine-command/mvcombine-command-overview-syntax-and-usage) | 2026-06-18T18:05:52.219Z |
| doc155 | [mvexpand command: Overview, syntax, and usage](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/mvexpand-command/mvexpand-command-overview-syntax-and-usage) | 2026-06-16T03:48:57.365Z |
| doc156 | [nomv command: Overview, syntax, and usage](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/nomv-command/nomv-command-overview-syntax-and-usage) | 2026-06-18T15:37:12.015Z |
| doc158 | [rename command: Overview, syntax, and usage](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/rename-command/rename-command-overview-syntax-and-usage) | 2026-07-14T22:31:29.013Z |
| doc159 | [replace command: Overview, syntax, and usage](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/replace-command/replace-command-overview-syntax-and-usage) | 2026-06-18T14:57:00.447Z |
| doc160 | [reverse command: Overview, syntax, and usage](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/reverse-command/reverse-command-overview-syntax-and-usage) | 2026-06-16T03:48:57.711Z |
| doc161 | [rex command: Overview, syntax, and usage](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/rex-command/rex-command-overview-syntax-and-usage) | 2026-06-18T09:04:43.715Z |
| doc163 | [search command: Overview and syntax](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/search-command/search-command-overview-and-syntax) | 2026-06-16T03:49:12.086Z |
| doc164 | [sort command: Overview, syntax, and usage](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/sort-command/sort-command-overview-syntax-and-usage) | 2026-06-16T03:49:11.080Z |
| doc165 | [spath command: Overview, syntax, and usage](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/spath-command/spath-command-overview-syntax-and-usage) | 2026-06-16T04:28:00.227Z |
| doc166 | [spl1 command: Overview, syntax, and usage](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/spl1-command/spl1-command-overview-syntax-and-usage) | 2026-06-16T03:49:32.110Z |
| doc167 | [stats command: Overview, syntax, and usage](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/stats-command/stats-command-overview-syntax-and-usage) | 2026-06-16T03:49:04.920Z |
| doc168 | [streamstats command: Overview, syntax, and usage](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/streamstats-command/streamstats-command-overview-syntax-and-usage) | 2026-06-16T03:49:25.435Z |
| doc169 | [table command: Overview, syntax, and usage](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/table-command/table-command-overview-syntax-and-usage) | 2026-06-16T03:56:32.628Z |
| doc170 | [tags command: Overview, syntax, and usage](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/tags-command/tags-command-overview-syntax-and-usage) | 2026-06-16T03:56:32.011Z |
| doc171 | [thru command: Overview, syntax, and usage](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/thru-command/thru-command-overview-syntax-and-usage) | 2026-06-16T03:49:35.227Z |
| doc172 | [timechart command: Overview and syntax](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/timechart-command/timechart-command-overview-and-syntax) | 2026-06-16T03:48:57.036Z |
| doc173 | [timewrap command: Overview, syntax, and usage](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/timewrap-command/timewrap-command-overview-syntax-and-usage) | 2026-06-16T03:49:24.644Z |
| doc174 | [tstats command: Overview, syntax, usage](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/tstats-command/tstats-command-overview-syntax-usage) | 2026-06-16T03:55:44.015Z |
| doc175 | [typer command: Overview, syntax, and usage](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/typer-command/typer-command-overview-syntax-and-usage) | 2026-06-16T03:49:31.960Z |
| doc176 | [union command: Overview, syntax, and usage](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/union-command/union-command-overview-syntax-and-usage) | 2026-06-16T03:49:14.314Z |
| doc177 | [untable command: Overview, syntax, and usage](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/untable-command/untable-command-overview-syntax-and-usage) | 2026-06-16T03:55:45.977Z |
| doc178 | [where command: Overview, syntax, and usage](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/where-command/where-command-overview-syntax-and-usage) | 2026-06-16T03:49:03.447Z |
| start | [Start searching data using SPL2](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-manual/getting-started/quick-start-write-and-run-a-basic-spl2-search/start-searching-data-using-spl2) | 2026-07-27T18:10:44.254Z |
| syntax | [Understanding SPL2 syntax](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/introduction/understanding-spl2-syntax) | 2026-06-16T03:49:14.478Z |
| types | [Built-in data types](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-manual/data-types/built-in-data-types) | 2026-07-27T18:10:46.819Z |
| comments | [Using comments in SPL2](https://help.splunk.com/en/splunk-cloud-platform/search/spl2-search-manual/comments/using-comments-in-spl2) | 2026-07-27T18:10:51.652Z |
| expressions | [Types of expressions](https://help.splunk.com/en/splunk-cloud-platform/search/spl2-search-manual/expressions-and-predicates/types-of-expressions) | 2026-02-06T04:57:27.306Z |
| predicate | [Predicate expressions](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-manual/expressions-and-predicates/predicate-expressions) | 2026-07-27T18:10:44.526Z |
| objects | [Array and object literals in expressions](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-manual/expressions-and-predicates/array-and-object-literals-in-expressions) | 2026-07-27T18:10:48.823Z |
| access | [Access expressions for arrays and objects](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-manual/expressions-and-predicates/access-expressions-for-arrays-and-objects) | 2026-07-27T18:10:48.858Z |
| templates | [String templates in expressions](https://help.splunk.com/en/splunk-cloud-platform/search/spl2-search-manual/expressions-and-predicates/string-templates-in-expressions) | 2026-07-27T18:10:48.768Z |
| lambda | [Lambda expressions](https://help.splunk.com/en/splunk-cloud-platform/search/spl2-search-manual/expressions-and-predicates/lambda-expressions) | 2026-07-27T18:10:48.930Z |
| fromusage | [from command: Usage](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/from-command/from-command-usage) | 2026-06-16T03:49:00.226Z |
| fromsyntax | [from command: Syntax](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/from-command/from-command-syntax) | 2026-01-16T00:21:47.256Z |
| process | [Process your search results](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-manual/getting-started/quick-start-write-and-run-a-basic-spl2-search/process-your-search-results) | 2026-07-27T18:10:44.598Z |
| fieldtemplates | [Field templates in expressions](https://help.splunk.com/en/splunk-cloud-platform/search/spl2-search-manual/expressions-and-predicates/field-templates-in-expressions) | 2026-07-27T18:10:48.895Z |
| renameexamples | [rename command: Examples](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/rename-command/rename-command-examples) | 2026-06-16T03:49:15.515Z |
| sortexamples | [sort command: Examples](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/sort-command/sort-command-examples) | 2026-06-16T03:49:34.167Z |
| evalcompat | [Compatibility Quick Reference for SPL2 evaluation functions](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/spl2-compatibility-profiles-and-quick-references/compatibility-quick-reference-for-spl2-evaluation-functions) | 2026-07-06T19:41:16.114Z |
| aggcompat | [Compatibility Quick Reference for SPL2 statistical functions](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/spl2-compatibility-profiles-and-quick-references/compatibility-quick-reference-for-spl2-statistical-functions) | 2026-06-16T03:50:01.647Z |
| mathfunc | [Mathematical functions](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/evaluation-functions/mathematical-functions) | 2026-06-16T03:49:07.004Z |
| textfunc | [Text functions](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/evaluation-functions/text-functions) | 2026-06-16T03:49:34.458Z |
| comparefunc | [Comparison and Conditional functions](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/evaluation-functions/comparison-and-conditional-functions) | 2026-06-16T03:48:59.359Z |
| convertfunc | [Conversion functions](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/evaluation-functions/conversion-functions) | 2026-06-16T03:49:01.131Z |
| mvfunc | [Multivalue eval functions](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/evaluation-functions/multivalue-eval-functions) | 2026-06-18T18:13:46.678Z |
| aggregatefunc | [Aggregate functions](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/statistical-and-charting-functions/aggregate-functions) | 2026-06-16T03:55:23.106Z |
| doc138 | Individual expand page linked from compatibility table returned cache miss | No individual-content claim |

## Review/handoff checklist

1. Controller review of this evidence matrix and narrow R01/R02 spec correction; keep H forms distinct from malformed forms.
2. After verified M4, finalize the real shared lowering/phase fields; reconcile every K/conditional claim with actual kernel behavior.
3. Convert the exact option/subform and ordinary function-arity obligations in [form-expansion.md](form-expansion.md) into durable fixtures; document remaining H/EH forms as capability limitations. Preserve at least 200 meaningful cases, 80 definite negative/profile cases and 40 deliberately SQL/mixed cases after deduplication/review.
4. Convert candidates into durable production fixture data during the authorized SDD plan; add exact source-slice, scope, flow, tri-state and adapter parity assertions. Run ANTLR generation and fixtures then. This artifact makes no test-pass or runtime-conformance claim.
