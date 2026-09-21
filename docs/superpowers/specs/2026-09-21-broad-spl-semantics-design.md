# Milestone 10 broad SPL semantic coverage design

**Date:** 2026-09-21
**Status:** Approved
**Scope:** Milestone 10 only
**Implementation base:** `origin/main` at `63f4eaf4708c368611d943f813f4ca2326232e7a`

## Context

Milestones 1 through 8 established the canonical Go analyzer, structured field
flow, direct requirement extraction, validation, safe rewriting, and shared Go,
CLI, REST, native C, and Python reports. Milestone 9 added an authored capability
ledger and compact executable corpus. The ledger makes syntax, semantics,
requirements, linting, and safe rewriting separate evidence-backed claims.

The Milestone 9 SPL ledger contains 86 records and 85 evidence cases. Fifteen of
22 command records have supported semantic claims, and 34 of 35 function records
have supported semantic claims. Seven command forms remain semantically
unsupported. Requirements claims for the existing 22 command records and 35
function records remain unassessed.

Milestone 10 expands classic SPL semantics for security detections. It does not
attempt complete stock-command coverage. It selects bounded forms using observed
detection impact, implements their syntax, field flow, requirements, and
diagnostics together, and updates the ledger in the same change.

Everything under `_build_plan/` remains temporary planning material. Production
code, tests, packages, documentation generation, and runtime behavior must not
read or depend on it.

## Research baseline

The selected scope is informed by read-only scans of two local repositories:

- `/Users/jmdelgad/repos/security_content` at `424a788f19`, for classic SPL.
- `/Users/jmdelgad/repos/linus_security_content` at `ec5910c`, for SPL2 patterns.

The scans used a binary built from the exact SPL Toolkit implementation base.
YAML `search` strings were supplied through inline corpus manifests. The classic
SPL scan selected 2,170 active detections after excluding the deprecated
detection directory and deprecated or removed statuses. It acquired and analyzed
every selected query.

The current SPL results are:

- 0 valid, 1,351 invalid, and 819 incomplete reports.
- 823 syntax-complete and 1,347 syntax-incomplete reports.
- 0 semantic-complete and 2,170 semantic-incomplete reports.
- 1,088 occurrences of the dominant syntax failure, `no viable alternative at
  input 'tstats'`.
- 6,834 observed macro stages across 2,063 detections. Unresolved repository
  source and filter macros keep those whole-query reports semantically
  incomplete until definitions can be expanded.

The most common unsupported or partial command families are:

| Family | Stage occurrences | Distinct detections in selected union |
|---|---:|---:|
| `tstats` | 1,126 | 1,097 |
| `fillnull`, `rex`, `spath`, `bin`, `bucket`, `regex`, `mvexpand` | 912 combined | 615 |
| `join`, `append`, `appendpipe` | 60 combined | 60 |

The selected function gaps occur in 199 distinct detections. Counting each
function once per affected detection yields 328 occurrences: `mvindex` 60,
`latest` 49, `true` 41, `earliest` 36, `stdev` 31, `relative_time` 24, `now` 22,
`mvfind` 20, `like` 17, `null` 16, and `strftime` 12.

The SPL2 scan selected 49 Linus detections. It produced 45 invalid and four
incomplete reports, with no syntax-complete or semantic-complete report. Those
queries are research input for Milestone 11 only. Milestone 10 adds no SPL2
syntax, semantics, or capability claims.

These repositories are optional research inputs. Their queries, layouts, and
availability are not build, test, package, or runtime dependencies. Maintained
acceptance uses locally authored, minimized cases.

## Goals

- Add evidence-backed syntax, field-flow semantics, and direct requirements for
  the highest-impact classic SPL forms found in representative security
  detections.
- Turn the dominant `tstats` syntax failure into sound analysis of known
  operands, outputs, and knowledge-object requirements.
- Add selected field-shaping commands and functions needed by representative
  multi-stage detection flows.
- Preserve exact locations, scope ownership, requirements, and diagnostics when
  a neighboring macro, branch effect, option, or output remains unsupported.
- Measure report-level syntax improvement and stage-level semantic improvement
  because unresolved repository macros keep many whole-query reports
  incomplete.
- Update the capability ledger and compact corpus for every support claim.
- Preserve all current product interfaces and adapter parity.

## Non-goals

- Complete stock SPL command or function coverage.
- Any SPL2 implementation or claim.
- Macro-definition loading or transitive knowledge-object expansion.
- Complete `join`, `append`, or `appendpipe` field-set merging.
- Search execution, result comparison, value evaluation, regex execution,
  performance modeling, or optimizer behavior.
- Event sampling, environment snapshots, compatibility assessment, placeholder
  resolution, or query fanout.
- New public commands, endpoints, C symbols, Python methods, report versions, or
  production dependencies.
- Reading either research repository during builds, tests, packaging, or
  runtime.

## Chosen approach

Use detection-first vertical slices. Each selected form receives typed grammar,
canonical Go semantics, requirement extraction, diagnostics, capability records,
and executable evidence together. This maximizes measured detection impact while
preserving the analyzer's conservative boundaries.

Two alternatives were rejected:

- A family-floor approach would add one shallow form for every family named in
  the PRD. It would spend significant effort on lower-frequency branch forms and
  produce a smaller measured corpus improvement.
- A framework-first approach would introduce generalized command, function, and
  branch registries before repeated implementations demonstrated the necessary
  abstraction. The current focused handler and transfer architecture is adequate
  for the selected forms.

## Architecture and ownership

`pkg/analysis` remains the canonical owner of syntax interpretation, field flow,
references, dependencies, requirements, diagnostics, status, and capabilities.
Adapters serialize and transport its values without command-specific logic.

The analysis flow remains:

```text
QueryDocument
  -> bounded lexer and analysis-only SPL grammar
  -> typed stage and expression handlers
  -> field environment, references, dependencies, diagnostics
  -> query-only requirement trace
  -> Analysis Result and Requirement Set
  -> capability ledger and executable evidence
```

### Grammar boundary

The legacy `query` entry point used by mapping and discovery remains unchanged.
New syntax belongs under the analysis-specific rules in `grammar/SPLParser.g4`
and, only where required, `grammar/SPLLexer.g4`. Generated parser files are
refreshed through the repository's pinned generator and committed using current
project conventions.

The analysis grammar must provide typed contexts for:

- `tstats` options, exact option macros, aggregates, data-model sources,
  predicates, grouping fields, and spans;
- selected field-command arguments and options;
- exact join keys and child subsearches;
- the selected function and `like(...)` forms.

The grammar must not use text slicing to recover distinctions it can express as
typed children. Syntax recovery may retain a sound prefix but cannot invent
operands or outputs after damaged input.

### Handler boundaries

The existing command and function registries remain the dispatch points. New
handlers are separated by responsibility:

- `spl_tstats.go` owns `tstats`, data-model inputs, grouping, aggregate outputs,
  and exact inline option macros.
- `spl_field_commands.go` owns `fillnull`, `rex`, `spath`, `bin`, `bucket`,
  `regex`, and `mvexpand`.
- The existing `functions.go` owns function arity, aggregate context, and
  deterministic argument-read behavior.
- Existing scope and dependency code owns subsearch traversal. Narrow branch
  handlers add join-key reads and uncertainty without introducing a general
  branch algebra.

These filenames describe intended responsibility. The implementation plan may
adjust a filename to match nearby conventions, but it must preserve the module
boundaries and avoid growing one mixed command file with unrelated parsing and
transfer logic.

### Reuse of the transfer kernel

New handlers use existing `semanticStage` operations for exact reads,
conditional reads, definitions, removals, projections, aggregation, source
reset, dependencies, diagnostics, and requirement tracing. They do not replay
analysis in an adapter or introduce a second field environment.

## Selected semantic slices

### `tstats` and data-model analysis

The selected form is:

```text
| tstats [supported literal options or exact macro]
    aggregate-expression [AS alias] ...
    [FROM datamodel=Model[.Dataset]]
    [WHERE search-predicate]
    [BY exact-field ... [_time span=literal]]
```

The implementation follows the published Splunk Enterprise
[`tstats` syntax](https://help.splunk.com/en/splunk-enterprise/spl-search-reference/10.2/search-commands/tstats)
only to define static operands and output shape. It does not claim execution,
acceleration, summary freshness, or performance equivalence.

The handler performs these steps in source order while applying effects at the
canonical command boundary:

1. Record supported literal options. Options whose values do not affect field
   identity remain value metadata within the typed stage.
2. Record an exact inline macro as a `macro` dependency and requirement at its
   source location. Do not expand or assume its body.
3. Reset the source environment because `tstats` is dataset-producing.
4. Record an exact `data_model` requirement for `Model` and, when qualified, a
   `dataset` requirement for `Model.Dataset` using the existing normalization
   and overlapping-location conventions.
5. Analyze `WHERE` through the typed search predicate reader. Field operands are
   source reads. Exact index, source, or sourcetype selectors retain their
   existing typed dependency behavior.
6. Record aggregate input reads and exact aliases. A supported aggregate without
   an alias uses the existing canonical implicit-output identity.
7. Record exact `BY` field reads and projected outputs. A literal span changes a
   grouping value but not field identity.
8. Produce a closed field environment containing grouping and aggregate outputs
   only when all output-affecting inputs are understood.

The following boundaries remain incomplete:

- an exact inline macro, because its body can alter options or semantics;
- `PREFIX(...)` or another output identity that the analyzer cannot state
  exactly;
- `prestats=true`, because downstream output ownership differs from ordinary
  aggregate output;
- dynamic data-model or dataset identities;
- wildcard grouping fields;
- unknown or malformed options.

Known aggregates, grouping fields, dependencies, outputs, references, and
requirements survive these boundaries. The stage remains semantic-incomplete,
and downstream unknown fields become indeterminate rather than falsely absent.

### Common field-shaping commands

#### `fillnull`

Supported syntax is `fillnull [value=literal] [exact-field ...]`.

- Without a field list, the handler preserves the known field shape and keeps an
  open input universe open.
- With exact targets, each known target records a consuming read, then becomes a
  derived conditional output that retains the input origin.
- A target not known statically does not become an unconditional external field
  requirement. Its existence and fill behavior are data-dependent.
- Dynamic selectors and unsupported options remain incomplete.

This boundary reflects the published
[`fillnull` behavior](https://help.splunk.com/en/splunk-enterprise/spl-search-reference/9.0/search-commands/fillnull),
including its dependence on whether a field exists in the result schema.

#### `rex`

Supported syntax is an exact input field, default `_raw`, optional exact
`max_match`, optional exact `offset_field`, and an exact quoted extraction
expression.

- The input field is a consuming read.
- Exact named captures become conditional derived outputs because a runtime row
  might not match.
- An exact `offset_field` becomes a conditional derived output.
- Sed mode preserves field identity but remains semantic-partial because value
  rewriting is outside the field-flow model.
- Dynamic expressions or ambiguous capture declarations preserve the input read
  and emit a source-located incomplete diagnostic.

Named captures are discovered by a bounded scanner over the exact string. It
recognizes only unambiguous named-capture openers outside escapes and character
classes and validates each output identifier. It does not compile or execute
PCRE and adds no regex dependency.

#### `spath`

Supported syntax is an exact input field, default `_raw`, an exact path, and an
explicit exact `output` field.

- The input is a consuming read.
- The explicit output is a conditional derived field.
- Auto-extraction, omitted output, or a dynamic path remains incomplete because
  the static output set is not known.

This follows the published
[`spath` input, path, and output roles](https://help.splunk.com/en/splunk-enterprise/spl-search-reference/10.4/search-commands/spath)
without parsing event JSON or XML.

#### `bin` and `bucket`

`bucket` is an alias of `bin`. The supported form has one exact input field,
supported literal binning options, and an optional exact alias.

- The source field is a consuming read.
- Without an alias, the source identity is replaced by a derived value with the
  same name.
- With an alias, the exact alias is a derived output and the source remains
  available.
- Binning options affect values, not field identity. Dynamic fields, aliases, or
  unsupported options remain incomplete.

#### `regex`

The supported form has either an exact quoted regex expression against `_raw`,
or an exact field followed by `=` or `!=` and an exact quoted regex expression.

- The selected field is a filter read.
- The command preserves field identity and output shape.
- Regex execution and match truth are not modeled.
- `NOT`, a dynamic field, or malformed syntax remains invalid or incomplete as
  appropriate.

The filter-only field effect follows the published
[`regex` command](https://help.splunk.com/en/splunk-enterprise/spl-search-reference/9.4/search-commands/regex).

#### `mvexpand`

The supported form reads one exact field and preserves that field identity.
Row-count and memory effects are not represented by the field-flow model.
Dynamic fields and unsupported options remain incomplete.

### High-frequency functions and expressions

The function registry adds these exact forms:

| Function | Context | Arity | Static field effect |
|---|---|---:|---|
| `mvindex` | eval or predicate | 2 or 3 | Reads its multivalue argument and any expression arguments |
| `mvfind` | eval or predicate | 2 | Reads its multivalue argument and expression arguments |
| `true` | eval or predicate | 0 | Constant, no field read |
| `null` | eval or predicate | 0 | Constant, no field read |
| `now` | eval or predicate | 0 | Constant for one search, no field read |
| `relative_time` | eval or predicate | 2 | Reads expression arguments |
| `strftime` | eval or predicate | 2 | Reads expression arguments |
| `like` | eval or predicate | 2 | Reads expression arguments |
| `earliest` | aggregate | 1 | Reads one aggregate input and creates the existing aggregate output form |
| `latest` | aggregate | 1 | Reads one aggregate input and creates the existing aggregate output form |
| `stdev` | aggregate | 1 | Reads one aggregate input and creates the existing aggregate output form |

These forms follow the official evaluation and date-time function references:

- [Evaluation functions](https://help.splunk.com/en/splunk-enterprise/spl-search-reference/9.4/evaluation-functions/evaluation-functions)
- [Date and time functions](https://help.splunk.com/en/splunk-enterprise/spl-search-reference/10.2/evaluation-functions/date-and-time-functions)

The analyzer models argument reads and output identity, not returned values.
Wrong arity, use in the wrong aggregate context, dynamic multivalue lambdas, and
functions outside the selected set remain source-located and incomplete.

### Conservative branch handling

`join`, `append`, and `appendpipe` continue to analyze child pipelines in child
scopes. Milestone 10 adds these bounded facts:

- exact `join` key fields are consuming reads in the parent scope;
- child references, dependencies, diagnostics, and direct requirements remain
  available in their child scope;
- known parent bindings remain available after the branch;
- the parent environment becomes uncertain after the branch, so a later unknown
  field is indeterminate rather than falsely source-bound or unavailable.

The analyzer does not claim a complete merged output field set, row
cardinality, collision behavior, join type behavior, or branch ordering.
Each branch stage retains a source-located `SPL_UNSUPPORTED_SEMANTICS` boundary
and a partial semantic ledger claim.

### Macro handling

An exact macro invocation in a supported position records:

- the exact normalized macro identity;
- its original spelling and source location;
- a direct exact macro requirement;
- an incomplete coverage gap for unresolved expansion.

Known sibling syntax and semantics continue after that gap when the grammar can
identify their boundaries. A macro never grants syntax or semantic support to
its unknown body. Loading definitions and computing transitive closure remain
Milestone 12 responsibilities.

## Requirement extraction

The existing private query-only requirement trace remains authoritative. New
handlers record facts during the canonical traversal; the projector does not
reparse command text.

For the selected forms:

- exact external source-field reads become direct field requirements;
- conditional or indeterminate reads remain conditional and produce gaps;
- exact data-model, dataset, and macro references become knowledge-object
  requirements;
- derived aggregate, capture, path, alias, and replacement outputs remain out of
  the requirement set;
- filters consume their selected fields without creating outputs;
- malformed or unsupported forms preserve sound requirements before the owned
  boundary and do not invent later obligations.

Every newly claimed form receives requirements evidence in the capability
corpus. Existing unrelated unassessed requirements claims are not upgraded
merely because their commands share helpers.

## Diagnostics and error handling

Unsupported options and dynamic forms emit `SPL_UNSUPPORTED_SEMANTICS` at the
smallest grammar-owned source range. Existing more-specific diagnostics remain
preferred when they already describe the limitation.

Known sibling facts survive an unsupported boundary. The stage and report retain
incomplete coverage. Syntax damage remains invalid under existing status
precedence, and no handler creates references or fields from unsound contexts.

The named-capture scanner and all new grammar traversals are bounded by input
length and the existing 4,096 lexer work-unit admission limit. They cannot read
files, access a network, execute a regex, or panic on malformed user text.

Existing behavior remains unchanged for:

- invalid UTF-8 and strict JSON admission;
- resource-limit reports;
- stable reference, requirement, stage, scope, and diagnostic IDs;
- deep-copy guarantees;
- invalid-over-incomplete status precedence;
- CLI exit codes and HTTP content outcomes.

## Capability ledger and corpus

The ledger remains the support authority. Parser registration, registry
presence, upstream documentation, and external corpus frequency do not prove a
claim.

For each selected form:

- add a new narrow record when the form is broader than an existing record;
- change an existing record only when its exact reviewed scope gains evidence;
- provide positive syntax, semantics, and requirements observations for
  supported behavior;
- provide positive plus incomplete observations for partial behavior;
- retain unsupported or unassessed states for dimensions without direct proof;
- leave linting and safe rewriting unchanged unless this milestone supplies
  exact dimension-specific evidence.

Record and evidence IDs remain stable and canonically ordered. The manifest
summary is recomputed. `capability_revision` changes with the semantic contract;
`toolkit_version` remains excluded from that revision. Existing command and
function compatibility projections remain additive compatibility surfaces.

## Detection-impact corpus

A new minimized SPL impact corpus provides multi-stage evidence without copying
complete external detections. Each case contains:

- a stable local ID;
- locally authored query text;
- source-family provenance and selection rationale;
- targeted stage selectors using command and exact source range;
- exact expected syntax, semantic, reference, dependency, transition,
  requirement, diagnostic, and completeness facts;
- a classification as positive, negative, or incomplete.

The fixture binds its pre-Milestone 10 aggregate baseline to the exact local
query bytes. Tests recompute:

- syntax-complete impact cases;
- semantically complete target stages;
- source-located non-macro coverage gaps;
- requirement items and gaps;
- preserved facts outside changed forms.

Every positive target stage must reach its declared expected state. Every held
boundary must remain incomplete. Aggregate syntax-complete case count and
semantic-complete target-stage count must strictly increase from the bound
baseline. Non-macro gaps must strictly decrease. No previously sound reference,
dependency, transition, requirement, or diagnostic outside the changed form may
disappear.

Whole-query status is not the success metric. An exact unresolved source or
filter macro can keep a representative query incomplete after all selected
stages improve.

The impact corpus complements the capability corpus:

- capability cases prove one narrow claim and dimension;
- impact cases prove that several supported forms compose without hiding a
  neighboring incomplete boundary.

Neither corpus reads an external checkout or stores runtime output as an
unreviewed oracle.

## Public interfaces and compatibility

No public operation is added. The existing operations carry the improved
canonical values:

- Go `Analyze`, `Requirements`, validation, rewrite, corpus, impact, and
  document operations;
- CLI `analyze`, `requirements`, validation, rewrite, and developer tooling;
- REST query and corpus routes;
- native C owned-result operations;
- Python `SPLMapper` methods.

No adapter classifies a new command or function. Analysis reports retain schema
version 1. Requirement sets retain schema version 1. Existing JSON schemas remain
additive and archived version-1 artifacts remain valid.

Generated OpenAPI artifacts are refreshed only through the repository's current
generator. If no public shape changes, regeneration must be byte-stable apart
from intentionally embedded examples or capability values.

New handwritten Go files, generated parser files, embedded ledger and corpus
assets, and copied acceptance tests must be included in
`python/native-source-files.txt` and `tools/release-source-files.txt` as required
by their existing roles. Runtime and release-content documentation manifests do
not receive source assets.

## Verification strategy

### Grammar and parser tests

- Prove accepted syntax for every selected exact form.
- Prove malformed and held forms at exact source locations.
- Prove the legacy parser entry point and mapper behavior are unchanged.
- Regenerate with the pinned ANTLR toolchain and verify deterministic output.
- Exercise Unicode, CRLF, quoted identifiers, commas, whitespace, and nested
  subsearch boundaries where relevant.

### Semantic and requirements tests

- Assert full stages, references, transitions, dependencies, diagnostics,
  coverage, and requirement sets for positive forms.
- Assert downstream field flow after each new command.
- Assert conditional outputs for `rex` captures and explicit path extraction.
- Assert data-model, dataset, and macro requirements with exact locations.
- Assert wrong arity, unsupported options, dynamic identities, and ambiguous
  outputs remain incomplete.
- Assert branch children retain evidence and downstream unknown fields become
  indeterminate.
- Assert returned reports and capability values remain deeply detached.

### Ledger and evidence tests

- Validate every changed claim against a same-dimension observation.
- Recompute state counts and revisions.
- Reject broadened records without new IDs or an explicit reviewed scope
  correction.
- Replay the compact evidence corpus through canonical analysis and
  requirements operations.
- Preserve current linting and safe-rewrite counts unless direct evidence is
  added.

### Impact tests

- Replay all minimized multi-stage cases.
- Compare target-stage and non-macro-gap metrics with the bound baseline.
- Prove held macro and branch boundaries remain visible.
- Prove no sound sibling fact disappears when one stage improves.

### Surface and packaging acceptance

- Compare representative canonical values across Go, CLI, REST, native C, and
  Python.
- Run existing analysis, requirements, validation, rewrite, corpus, impact,
  graph, SARIF, document, LSP, and mapping regressions.
- Run Go formatting, vet, and race checks.
- Run source-native Python and documentation checks.
- Build direct wheels and source distributions, rebuild a wheel from the source
  distribution outside the checkout, and run the required installed suites with
  zero required skips.
- Verify embedded assets and all native build sources are available without the
  checkout or either research repository.

Hosted CI, cross-platform release jobs, sanitizers, publication, deployment,
live Splunk execution, and runtime equivalence remain separate gates. Local
checks do not imply them.

## Documentation

Update maintained capability, analysis, requirements, CLI, API, architecture,
compatibility, and Python documentation where the selected forms affect current
examples or limitations. Documentation must state:

- the exact supported forms;
- conditional and incomplete boundaries;
- the distinction between field-flow semantics and runtime value behavior;
- the distinction between exact macro requirements and macro expansion;
- the detection-impact measurement method;
- the lack of live Splunk execution or runtime certification.

After implementation and acceptance, write
`_build_plan/milestones/10-broad-spl-semantics/milestone-log.md` using the
required headings. Its first section is `## What's new in the app` and contains
concise user-visible bullets.

## Acceptance criteria

Milestone 10 is complete when all of the following are true:

1. Every selected positive form has typed grammar, canonical field-flow
   semantics, source-located diagnostics, direct requirement extraction, and
   executable ledger evidence.
2. Every selected held form remains explicitly partial, unsupported, or
   unassessed with exact boundary evidence.
3. Supported `tstats` forms produce exact data-model and dataset requirements,
   predicate and grouping reads, aggregate outputs, and closed output shapes.
   Inline macros retain exact macro requirements and incomplete expansion.
4. Selected `fillnull`, `rex`, `spath`, `bin`/`bucket`, `regex`, and `mvexpand`
   forms produce the field effects defined in this design without executing
   values or regexes.
5. Selected evaluation and aggregate functions enforce exact arity and context
   and contribute the expected field reads and outputs.
6. `join`, `append`, and `appendpipe` preserve child evidence and requirements,
   record exact join-key reads, and make downstream unknown fields indeterminate
   without claiming complete branch merging.
7. The minimized impact corpus has strictly more syntax-complete impact cases
   and semantic-complete target stages, strictly fewer non-macro gaps, no
   unexpected target gaps, and no lost sound sibling facts compared with its
   exact bound baseline.
8. Capability summaries and semantic revisions are deterministic. Existing
   record meanings, compatibility projections, archived schemas, and adapter
   values remain compatible.
9. Go, CLI, REST, native C, and Python return equivalent canonical values for
   representative new forms.
10. Existing analyzer, requirements, validation, rewrite, tooling, packaging,
    and documentation gates pass without required skips or unreviewed fixture
    regeneration.
11. Builds, tests, packages, and runtime require no external repository,
    network, live service, `_build_plan` file, environment snapshot, macro
    definition, or later-milestone feature.

## Later milestone handoff

Milestone 11 may reuse the research patterns when it expands SPL2, but it must
define SPL2 forms and evidence independently. Classic SPL support does not imply
SPL2 support.

Milestone 12 owns supplied knowledge-object definitions, macro expansion,
transitive dependency closure, cycle detection, and Detection BOM output. The
exact macro requirements and source locations produced here are its inputs, not
an implementation shortcut.

Milestones 13 through 17 retain ownership of environment contracts, live
metadata export, compatibility assessment, resolution, and repository-scale CI
composition. No Milestone 10 result claims that a target environment can run a
query or that runtime results are correct.
