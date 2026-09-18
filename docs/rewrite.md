---
title: "Safe rewrite"
layout: page
---

# Safe rewrite

`rewrite` previews or applies explicit source-identity mappings in SPL and
standalone SPL2. The Go package `pkg/rewrite`, CLI, native Python and REST return
the same versioned report. Select `language: spl2` explicitly; defaults are
`spl`, `splunkd`, and `current`. `current` names the contract bundled in this local
build, not a promise to follow a remote Splunk release. No query is executed.

## Preview and apply

The maintained rules file `testdata/rewrite/example-rules.json` contains:

```json
{"schema_version":1,"rules":[{"id":"source-user","kind":"field","source":{"name":"src"},"target":{"name":"user"}}]}
```

Run `build/spl-toolkit rewrite --rules testdata/rewrite/example-rules.json --query
'search src=alice | rename src AS owner | table owner' --format json` on one line.
The candidate is `search user=alice | rename user AS owner | table owner`.
Preview returns the original query as `text`, with `committed: false`.
Adding `--apply` returns the candidate as `text` and sets `committed: true` after
static verification. The explicit alias `owner` stays fixed. The [CLI guide](cli.md)
contains executable file, stdin, batch, output and validation examples.

Read these report fields separately:

| Field | Meaning |
| --- | --- |
| `document`, `original_text` | Original source and normalized selectors; source bytes and source ID are retained. |
| `candidate_text`, `candidate_analysis` | Proposed query and its canonical analysis, including a candidate rejected by apply. |
| `text`, `committed` | Authoritative returned query and whether candidate edits were committed to this returned string. |
| `changes` | Exact old/new text, rule/group IDs, original/candidate reference IDs and locations, candidate inclusion and commitment. |
| `rule_evaluations` | Every rule decision, including unmatched rules without fabricated locations. |
| `status`, `coverage` | Static errors and incompleteness; neither is a runtime equivalence guarantee. |

Locations use half-open UTF-8 byte offsets, one-based Unicode columns and CRLF
line handling. Untouched whitespace, quoting and comments retain their original
bytes. Query files are never modified in place; `--output` writes a report only.

CLI content exits are 0 (valid), 1 (invalid) and 3 (incomplete); request errors
exit 2. An independent safe group may commit beside an ambiguous or unknown
group, so **incomplete does not imply uncommitted**. REST uses HTTP 200 for all
query statuses, 400 for request errors and 500 for internal failures.

## Rules and original facts

A request has integer `schema_version: 1`, `mode` (`preview` by default or
`apply`), a Query Document, `rules`, and an optional `validation_target`. Rule
IDs are unique. Kinds are `field`, `index`, `source`, `sourcetype`, `lookup`,
`dataset` and `data_model`. Identities are logical data, never query fragments.
`{"name":"actor.name"}` is an atom distinct from `{"path":["actor","name"]}`.
Path identities are accepted only for field rules; current navigated source
bindings are unproved and remain refused. Names preserve meaningful spaces/case.

Rules may have `when` with nonempty `all`/`any` arrays, a
`source_reference_present` field fact, or a `literal` fact with `equals` or
string-only `contains`. Literal kinds are field/index/source/sourcetype; scalar
strings, numbers, Booleans and null remain distinct. Contains checks proven
literal values, not query text. There is no condition-language NOT, regex,
priority, caller-supplied context or executable expression.

Dependency literal conditions name the grammar selector, while dependency mapping
rules name its value. For `search sourcetype=auth src=alice | table src`, this
condition authorizes a `src` to `user` rule:

```json
{"fact":"literal","kind":"sourcetype","identity":{"name":"sourcetype"},"operator":"equals","value":"auth"}
```

The equivalent selector identities are `index` and `source`. An expression field
named `sourcetype` remains a `field`-kind fact; it is not a dependency by spelling.

Facts come from the original canonical query and its flow point. Compatible AND
guarantees combine; OR keeps only common guarantees; query NOT establishes no
positive equality. Later restrictions cannot authorize earlier references.
Incompatible exact AND restrictions on one identity establish no usable guarantee;
subsequent repeated restrictions or OR branches cannot revive that conflict.
Complete supported absence is false, while unsupported evidence remains unknown.
Unrelated independently proven facts can still authorize their own rules.
Independent children neither inherit nor export parent facts; inherited subpipes
receive the existing environment. Overwrites, removals and unknown flow invalidate
affected facts. A conclusive false condition is an ordinary skip; unknown evidence
keeps the report incomplete. Static reference presence does not prove event-field
presence. All condition children retain evidence even when one controls the result.

Supported SQL HAVING equalities enter the original fact flow at the HAVING phase,
after grouping/selection and before later consumers. They do not authorize earlier
phases or manufacture source facts for derived aggregate aliases. Conflicting
WHERE/HAVING restrictions cannot authorize a later rewrite.

Numeric equality is exact decimal equality without expanding powers of ten.
SPL2 supports exponent notation in query literals; the current SPL grammar does
not. CLI/HTTP/Go/native JSON requests can retain arbitrary numeric tokens, including
huge exponents. Python convenience methods accept Python JSON values, not raw
numeric tokens or Decimal objects: they cannot express a nonzero arbitrary-precision
exponent token without a representable Python value. No float coercion is performed
to widen that interface.

Mappings are simultaneous: a→b and b→c use original identities, and do not cascade
a into c. Identical replacements coalesce, while different targets are ambiguous.
Linked consumers must be authorized together. Static checks reject observed live
binding collisions, alias capture and groups dependent on skipped edits. They do
not assume that unreferenced fields are absent in actual events.

## Supported forms and limits

The selected capability manifest's additive `rewrite` member contains
`schema_version` and `forms`. Each form has `kind`, stable rendering `role`,
`identity_forms`, `supported` and `limitations`. A supported role still requires
the actual operand's source binding, rendering and whole-candidate proof.

The ledger's `safe_rewriting` dimension is an evidence claim per language form,
not a projection of `rewrite.forms[].supported`. It uses the same supported,
partial, unsupported, not-applicable and unassessed states as syntax, semantics,
requirements and linting. Supported alone contributes to `covered`; applicable
equals supported plus partial plus unsupported plus unassessed, and total records
equal applicable plus not applicable. The current SPL summary is 1 unsupported
and 67 unassessed; SPL2 is 1 unsupported and 97 unassessed. These exact counts do
not form a score or percentage.

Evidence IDs bind a claim to typed local requests and observations. A broader
rewrite form gets a new ID unless a reviewed scope correction changes the
original boundary. Grammar registration and analysis diagnostics do not provide
safe-rewrite or lint coverage. The corpus establishes static local toolkit
behavior only, not live Splunk execution, runtime equivalence, environment
compatibility, authorization or upstream support.

Exact search/expression/selector atoms, rename inputs, null-inspection operands
and proven lookup-local inputs can map. Explicit eval/rename/SQL/lookup outputs
and downstream derived aliases remain fixed. Lookup catalogs are independent of
column labels. An unaliased lookup input preserves the remote column and can
insert `AS` for the local target when canonical proof succeeds.

Static index/source/sourcetype values are dependencies, separate from similarly
spelled event fields. Named datasets and lookup catalogs can map in their typed
slots. SPL model/dataset composites preserve component intervals and normalization;
compatible overlapping changes co-render, contradictory ones are refused. SPL2
tstats `datamodel_name='Traffic.All'` is one atomic data-model identity. Its
generating field effects remain incomplete even when that dependency edit commits.
SPL2 metric index mapping requires intact command-owned bare-index equality.

SPL implicit aggregate inputs and exact generated-name consumers move together,
without a synthetic AS. SPL2's narrow known implicit names do not authorize a
changed label: `FROM main | stats sum(bytes) | table 'sum(bytes)'` with
bytes→octets is withheld. Direct SQL passthrough projections preserve source
bindings; SQL aliases and logical phases are checked independently.

Wildcard selectors, dynamic/interpolated identities, navigated paths, unproved
alias-qualified joins, uncertain literal decoding, unknown owners and affected
unknown flow remain located refusals. Damaged original syntax prevents edits.
Comments, string literals, option labels/values, function names and remote lookup
column labels are not inferred field sites. A source-less SPL2 `where` and SPL's
`inputlookup` spelling used as a SPL2 command do not gain support through rewriting.

## Destination validation and batches

Optional targets reuse the existing field-list, JSON Schema and local OCSF
contracts. Validation checks the candidate's destination source obligations,
not whether the old names belonged to the target. A catalog containing only
`src` rejects a src→user candidate: apply returns the original, retains the
candidate, and marks included changes skipped with `post_verification_failed`,
`candidate_applied: true`, `committed: false`. Invalid or incomplete target
validation prevents every candidate edit from committing. Unresolved schema URLs
remain unresolved; no network retrieval occurs.

`candidate_validation` retains the full canonical validator report. In its
field-list JSON branch only, `target` retains `kind`, `identity`, `version` and
omits repeated `fields`/`optional_fields`; the Go validator report is unchanged.
No-op requests still validate an explicitly requested target.

`RewriteBatch`, Python `rewrite_batch`, CLI `--batch` and
`POST /api/v1/query/rewrite/batch` preserve document order and share rules/mode/target.
Invalid query syntax is a per-query result. Malformed documents, rules, selectors
or targets reject the whole request without partial reports. Single/batch REST
bodies use the existing 8 MiB schema-route limit. Rules/resources remain local CLI
inputs or inline HTTP data; HTTP accepts no server file paths.

## Existing mapper compatibility

Legacy `map`, `map_query`, `MapQuery` and context mapping keep their SPL-only
configuration, precedence and caller-context semantics. They do not acquire the
new original-fact, conflict, audit or destination-validation policy. New rewrite
calls ignore legacy mapper configuration. Neither API promises Splunk runtime
equivalence, raw-to-data-model conversion, whole-query translation, event instance
validation, learned mappings, or SPL2 module execution.

The durable corpus separates authored semantic expectations from full Go transport
reports. Installed acceptance copies and hashes rewrite/schema/SPL2 inputs and
maintained examples outside the checkout, requires each native/surface suite with
nonzero collection and zero skips, and compares every JSON array/value across
Go, CLI, real HTTP and native Python for both the wheel and rebuilt-sdist wheel.
Local automated evidence does not substitute for unexecuted release-platform jobs
or actual Splunk execution.
