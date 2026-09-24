---
title: "SPL2 static analysis"
layout: page
---

# SPL2 static analysis

SPL Toolkit analyzes standalone SPL2 queries and selected same-document module
declarations offline. Select `spl2` explicitly;
omitting the language still selects the existing SPL contract. Both dialects use
profile `splunkd`, compatibility version `current`, and integer report format `1`.
`current` names this build's bounded capability snapshot. It does not select a
running Splunk installation or certify query execution against one.

## Start a document

An explicit `search`, a `FROM` pipeline, a `SELECT … FROM …` query, and approved
generating commands are standalone starts. The documented initial selector form
`index=app failure` may use implicit search. An arbitrary leading word such as
`failure index=app` remains an incomplete unknown command; the toolkit does not
infer search intent or insert a search prefix. Raw parser findings and the
canonical unknown-command recovery classification remain separate evidence.
SQL-first and FROM-first clauses may feed later pipe commands. Original text,
CRLF, UTF-8 byte offsets, Unicode names, and opaque `source_id` values are retained.

The [CLI guide](cli.md) contains executable commands, exact help and stdin/batch
examples. The [HTTP guide](api-server.md) covers exact request bodies and errors.

```bash
build/spl-toolkit analyze --language spl2 --query 'SELECT host FROM main WHERE bytes>0' --format json
```

In Go, call `analysis.Analyze(analysis.QueryDocument{Text: query, Language: "spl2"})`.
Use `analysis.CapabilitiesFor(analysis.CapabilityOptions{Language: "spl2"})` to
inspect command forms, function arities and limitations. The unchanged
`analysis.Capabilities()` returns the default SPL manifest.

```python
from spl_toolkit import SPLMapper

with SPLMapper() as mapper:
    report = mapper.analyze_query("SELECT host FROM main WHERE bytes>0", language="spl2")
    fields = mapper.validate_fields("FROM main SELECT host", {"fields": ["host"]}, language="spl2")
    schema = mapper.validate_schema(
        "FROM main SELECT host",
        {"kind": "json_schema", "schema": {"type": "object", "properties": {"host": True}}},
        language="spl2",
    )
    capabilities = mapper.capabilities(language="spl2")
```

HTTP `POST /api/v1/query/analyze` accepts
`{"text":"FROM main SELECT host","language":"spl2"}`. Field and schema routes
wrap that same document with `catalog` or `target`. Their `/batch` routes receive
ordered `documents`, each with its own selectors; SPL and SPL2 may share a batch.
The CLI uses `--language spl2`, `--profile splunkd`, and
`--compatibility-version current`; batches carry those values in each document
and reject global document flags. Python methods use keyword-only `language`,
`profile`, `version`, and `source_id`. Empty compatibility strings normalize to
the defaults; selector spelling and case remain exact.

The C functions `spl_mapper_analyze_query`, both field-validation functions and
both schema-validation functions take canonical JSON wrappers. The additive
`spl_mapper_capabilities_for` takes a JSON selector object. Each returns an owned
`SPLResult`; release it with `spl_result_free`. Native Python uses those same
functions and contains no second semantic registry.

## Syntax and modeled effects

The grammar has dedicated forms for the reviewed command and language families,
including expressions, quoted identifiers, raw and interpolated strings,
arrays/objects, lambdas, SQL clauses, child queries, local declarations, and
command-specific options. Parser acceptance is separate from complete field
effects. The capability manifest lists evidence and limitations by form.

The evidence ledger is more granular than that legacy inventory projection. The
current SPL2 manifest has 140 form records and 146 evidence cases. Syntax has 89
supported, 19 unsupported and 32 unassessed records. Semantics has 75 supported,
39 unsupported and 26 unassessed records. Requirements has 45 supported, 7
unsupported and 88 unassessed; linting has 140 unassessed. Safe rewriting has
14 supported, 3 unsupported and 123 unassessed. Supported is the
only state counted as covered; partial, unsupported and unassessed remain in the
applicable denominator. Not-applicable records are outside that denominator. No
composite score is produced.

`grammar_registered` is separate from evidence-backed syntax coverage. In
particular, `spl2.command.spl1.quoted-pipeline` has grammar registration because
the parser owns the form, but its syntax claim is unsupported and receives zero
covered syntax credit. Its cited incomplete case preserves the H10 embedded-body
boundary. Analysis findings also do not count as lint evidence; linting remains
unassessed until exact lint observations are reviewed.

Approved ordinary forms of `search`, `from`, `eval`, `where`, `fields`, `table`,
`rename`, `stats`, `eventstats`, `streamstats`, `lookup`, `sort`, `dedup`, `head`,
`reverse` and `select` have modeled field effects. These include sequential eval
assignments, supported positional scalar/aggregate arities, source and derived
reads, exact projection/removal, independent rename pairs, and ordered lookup
outputs. Complete support is conditional on the actual operands and field evidence.
Without explicit lookup outputs, unknown remote columns may overwrite existing
local aliases, so later consumers can remain indeterminate even when the match
input was a definite source read. Explicit OUTPUT and OUTPUTNEW retain their
separate overwrite and preserve-existing policies.

Selected `bin` and `mvexpand` forms have modeled transfers. Selected guarded
`if`/`branch`, `union`, and qualified pipeline `join` forms merge child output
under the rules below. Other forms of these commands remain incomplete.
`append`, `appendcols`, `appendpipe`, `spl1`, `rex`, `spath`, `makeresults`,
`loadjob`, `tstats`, `mstats`, `timechart`, `timewrap`, `makemv`, `mvcombine`,
and `fillnull` have dedicated syntax but incomplete overall effects. Intact
child reads remain visible without proving the parent's output merge. An intact
output identifier in a held form records output intent without installing a
field.

The generating effects of `tstats`, `mstats` and `timechart` remain unmodeled:
known input names and origins become conditional, with an open and uncertain
output universe. Wildcard grouping never installs a literal pattern name or
claims catalog membership. In held `union` forms, named inputs retain dataset
dependencies, while array and child-query inputs do not establish merged output
fields. Selected `union` forms follow the merge rules below.

A tstats `datamodel_name` value is one atomic `data_model` reference and dependency,
including the whole original quoted token. Static loadjob SIDs use reference kind
`search_job`, with no new dependency-list member or event field. These located
identities do not by themselves authorize rewriting them.

`timewrap` requires a preceding `timechart` in a proved pipeline. A known missing
prerequisite is a contract error even when an unreviewed unit remains incomplete;
the unit itself does not become invalid merely because it is unreviewed.

The native deferred inventory remains incomplete without invented child grammar.
Unknown commands do not acquire field semantics from their spelling. Malformed
supported syntax still produces a definite error.

Named arguments, unknown functions, dynamic field templates, unresolved navigation,
ambiguous aliases and unproved output labels retain incomplete coverage even when
the surrounding schema matches. Function arity/context support is in the selected
manifest. For example, `round(precision:2,num:bytes)` retains the `bytes` read but
does not claim complete named-argument semantics. A held form such as
`where user IS NOT string` remains incomplete. The 17 provenance holds are retained
as disputed boundaries and receive no completed corpus credit.

Rename pairs must be independent. Duplicate sources, duplicate destinations,
chains such as `a AS b, b AS c`, and overlaps are invalid. An input renamed to a
derived name still requires its original source declaration. `fields` retains
known internal fields unless explicitly removed. With an open source universe,
unproved internal retention or wildcard expansion can remain incomplete.

## SQL source order and execution

For `SELECT host FROM main WHERE bytes>0`, the source-ordered stages are
`select`, `from`, `where`; their logical positions are `2`, `0`, `1`. IDs, spans
and references preserve lexical order. Lineage records actual evaluation order:

| Phase | Owning stage | Result |
|---|---|---|
| `source` | `from` | Open input from `main` |
| `filter` | `where` | Read `bytes` from the input |
| `evaluate` | `select` | Read `host` while the source is still available |
| `project` | `select` | Restrict the final output to `host` |

Each SQL lineage record carries `phase` and `execution_order`. Grouped queries
may add `group`, `aggregate`, `having`, `order`, `limit` and `offset` phases.
The final projection reuses the original reference IDs and creates no extra read.
Thus a field catalog containing only `host` still reports the earlier missing
`bytes`; a pipe command after the SELECT cannot read projected-away `bytes`.
Scope-local positions must not be confused with global lexical IDs.

An aggregate-only SELECT can prepare supported aggregates for HAVING without a
GROUP clause, for example `SELECT count() AS n FROM main HAVING n>2`. Hidden
inputs, mixed non-grouped projections and unproved aggregate expressions retain
their limitations; the presence of aggregate syntax alone is insufficient.

The zero-argument aggregate `count()` has the proven implicit output name `count`.
An operand-bearing count label without an established spelling needs an alias or
remains incomplete. Hidden ORDER/HAVING inputs, sibling alias reuse, source/alias
collisions, SQL joins and unproved compound aggregate labels remain bounded. A
WHERE read is not silently converted into a SELECT alias read.

An exact finite SELECT output set can be known even when one of its fields is
conditionally present and the query's effects remain incomplete. An intact
`count()` output has its own presence proof when its destination is exact and
does not collide with another possible output. This does not make unrelated
grouping expressions, aliases or other outputs complete. An incomplete wildcard
expansion retains uncertainty even when its expression has an explicit alias.

## Null, conditional presence and paths

Complete modeled effects may include `conditional: true` fields. This is a
structural presence result, with no promise about any event's value. A consuming
read of a conditionally present field remains indeterminate; matching schema
declarations cannot turn it into a certain match. This is separate from an
optional declaration, which may be an accepted declaration match.

`eval x=null` removes `x` and preserves a tombstone. A later ordinary read of `x`
is unavailable. Finite intact structural all-null compositions follow the same
removal rule only where the expression contract proves an all-null result; real
RHS reads are still retained. Dynamic or unproved nullability does not authorize
speculative deletion. An intact supported string template produces a string even
when an interpolation is nullable; unknown or damaged child expressions still
prevent a complete semantic claim.

Direct `isnull(field)` / `isnotnull(field)`
use reference role `null_test`. The reference retains its original location,
binding and origins but has no required-existence validation outcome. Ordinary
sibling or later consuming reads still require presence. For example,
`eval answer=isnull(absent) | where absent>0` still requires `absent` for the WHERE.

SPL2 quoted literal dots and nested navigation are distinct identities.
`'actor.name'` is one atomic field name; `actor.name` is a structural path.
The analyzer keeps those identities separate internally, including for exact
reads, removals and wildcard matching. The public field state has string names,
so both identities project to `actor.name`. If they coexist, it reports
`SPL_AMBIGUOUS_FIELD`, merges their visible origins into one uncertain public
binding, and leaves semantic and requirement coverage incomplete. A later
projection cannot turn that collision into a proved identity. Resolver expansion
checks every candidate, including broad `*`, against typed dotted identity.
Exact unaffected members and their schema evidence remain visible.

An exact static dataset descriptor produces one canonical JSON dataset identity.
For example, `FROM {kind: "index", properties: {name: "events"}}` has the identity
`{"kind":"index","properties":{"name":"events"}}` in dependencies and direct
requirements. Object keys are ordered canonically while array order is retained.
Duplicate decoded keys, a missing or dynamic kind, or a dynamic property value
cannot claim an exact dataset. The analyzer keeps a located dynamic reference,
any independently parsed child evidence, and an incomplete requirement instead.

## Same-document declarations

Selected module documents can declare local `$view` pipelines and pure scalar
`function` bodies, attach annotations, and export local names. Forward references
resolve within the submitted document. A local view contributes its field-flow
summary to a later `FROM $view`; it is not an external dataset requirement. A
local function substitutes positional parameters into its reviewed expression
body, so the call keeps the caller's field origins. Duplicate symbols, unresolved
exports, invalid calls and declaration cycles are contract errors. A damaged
declaration cannot make its dependents complete.

```spl
$base = FROM main | fields host;
function normalize($value) { return lower($value); }
$output = FROM $base | eval key=normalize(host);
export output;
```

Imports are recorded as module and member references, but external modules are
not fetched or bound. An unused import can leave analysis `valid` and semantic
coverage complete while requirement coverage is incomplete. Using an imported
member makes the affected analysis incomplete with `SPL_UNRESOLVED_MODULE`.
Import aliases do not become event fields or external datasets. This module
support is same-document static analysis, not module execution or linking.

## Selected branch merges

Supported `if` arms and guarded `branch` arms inherit a copy of the parent field
environment. A missing `else` adds the unchanged parent path. `union` merges
exact named datasets, static descriptors, local views, and independent child
queries; a pipeline `union` also includes its input path. Named external dataset
shapes remain open. Selected pipeline `join` requires distinct `left` and `right`
aliases, `type=inner|left|outer`, a qualified equality predicate, and an intact
right child query. SQL joins and other pipeline join layouts remain incomplete.

At an alternative merge, fields present on every reachable path remain present;
fields present on some paths become conditional. Different origins or removal
states can leave the result uncertain. A selected join combines matched left and
right fields, then adds unmatched sides for left or outer joins. Duplicate public
output names produce `SPL_AMBIGUOUS_FIELD` and incomplete coverage. Guard and
join-key reads keep their original scopes and direct requirement evidence.
Branch requirement traces are forked and merged with the same reachability rules;
a conditional requirement remains conditional. Unsupported child syntax or an
unproved merge does not install guessed fields into the parent.

## Boundaries and evidence

A `valid` result requires complete syntax and modeled effects with no error.
`invalid` records a definite error and can still retain incomplete coverage.
`incomplete` reports an unresolved contract without manufacturing validity or an
error. In schema/field validation, inspect `schema_complete` as well as both
analysis coverage flags. Only JSON object-key order is irrelevant when comparing
canonical reports; source, ordered arrays, locations, phase metadata, bindings,
dependencies, diagnostics and target evidence all remain part of the contract.

Selected local declarations are described above. Other module syntax remains
outside the bound contract and can report `SPL_UNSUPPORTED_MODULE`. `decrypt`,
`ocsf` and `route` are outside the `splunkd` query profile and use
`SPL_PROFILE_MISMATCH`. These located content
findings differ from malformed syntax (`SPL_SYNTAX_ERROR`), unmodeled semantics
(`SPL_UNSUPPORTED_SEMANTICS`), and a malformed request or unsupported selector.
Unknown-owner recovery cannot erase lexer errors, unterminated literal modes,
malformed adjacent supported commands, or damaged typed children. Recovery retains
independently sound original typed operands and owners. A token swallowed into an
error node does not acquire a field, source or clause role from its apparent intent.
Intact bracketed children can survive errors in separate parent options; missing
original delimiters or damaged containing bodies do not create child scopes.

Legacy map/discover/context-map/input-field operations retain their SPL behavior.
Explicit SPL2 selection rejects with `unsupported_dialect_for_operation` and
guides callers to canonical analysis or validation. It never falls back to SPL.

The SPL2 manifest's optional `documentation_snapshot` is
`spl2-provenance-v1:sha256:3345cf5712b1bdbf467d1651784fdb8bccc596805038da0d54e7a123384e3a4e`.
It identifies the durable V1 source/design, original IDs/candidates/source links,
inventory and hold projection. It is not a downloaded-page-body hash or the
mutable current corpus. The default SPL manifest omits it. Runtime execution
does not read `_build_plan` or retrieve documentation.

Record and evidence IDs retain their exact reviewed scope. Broadening a form
requires a new ID unless a reviewed scope correction establishes that the old
boundary was wrong. The embedded corpus proves static local toolkit behavior. It
does not prove live Splunk execution, runtime equivalence, environment
compatibility, authorization or upstream Splunk support. Parser acceptance,
semantic completeness and direct requirement completeness are separate report
claims. Lint support and safe rewrite support require their own ledger evidence;
neither follows from successful analysis. Local package and transport checks do
not prove deployment or live runtime compatibility.

The [compatibility record](compatibility.md) distinguishes local verification
from release-platform, interpreter and live/operator acceptance. No local corpus,
package or parity result alone establishes live Splunk execution compatibility.
