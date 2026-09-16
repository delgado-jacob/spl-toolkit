# Milestone 8 - Product interface and canonical requirements

Status: approved architectural design. This document defines Milestone 8 only. It does not authorize implementation outside the milestone, and nothing under `_build_plan/` may become a runtime, build, package, test-data, or documentation-generation dependency.

## Purpose and scope

Milestone 8 adds two product-level query operations to the existing canonical analysis system:

- `analyze(query)` returns the current structured meaning, diagnostics, coverage, provenance, and an embedded requirement set.
- `requirements(query)` returns that same first-class requirement set without requiring callers to consume the rest of the analysis report.

`pkg/analysis` remains the canonical Go implementation. The milestone extends that package instead of introducing a facade or a second requirements package. CLI, REST, native/C, and Python adapters delegate to the canonical operations and preserve the same report values.

The requirement set describes direct external obligations visible in the submitted query. It does not load environment metadata, inspect event data, expand supplied knowledge-object definitions, validate compatibility, resolve placeholders, generate query variants, or add broad SPL or SPL2 language coverage. Transitive requirements, environment snapshots, compatibility, and resolution remain owned by later milestones.

## Architecture

Canonical SPL or SPL2 analysis first normalizes the query document and enforces a shared lexer work budget before parser prediction. Documents admitted by that boundary perform the existing parse, field-flow, dependency, reference, diagnostic, coverage, and status work. After reference IDs and query-only analysis evidence are final, a deterministic projector builds the requirement set. The projector consumes canonical evidence. It does not parse query text, replay field transfers, or guess identities from source spans.

`analysis.Result` always contains a non-optional `Requirements` value. `analysis.Requirements(document)` performs one canonical analysis pass and returns a deeply detached copy of the set embedded in that pass. Calling `Analyze` gives callers meaning and requirements together without repeated work.

### Query-only requirement trace

Some existing operations call `AnalyzeWithSourceFields` or `AnalyzeWithSourceUniverse`. Field-list validation, schema validation, and rewrite proof may refine public reference bindings, wildcard membership, diagnostics, and coverage using a caller-supplied target. Milestone 8 requirements must remain a function of the query and selected language capability contract only.

The analysis pass therefore maintains a private requirement trace alongside its public refinement-aware evidence. The trace records the complete query-only consuming references, diagnostics, status inputs, and completeness inputs before source-universe admission changes them. Reference finalization remaps trace links to the same final `ref-N` identifiers used by the public result. Pending-reference lookup and incomplete-stage membership use O(1) indexes that remain synchronized through recording, remapping, and completeness updates; environment clones retain access to the shared trace indexes. The projector uses this trace, not target-refined conclusions.

For plain `Analyze`, trace evidence and public analysis evidence coincide. For validation and rewrite analysis, the embedded requirement set must be byte-equivalent to the set from plain analysis of the same normalized query. It excludes refinement-target diagnostics and proofs. This trace is a projection feed within the existing traversal, not a second parser or semantic engine.

### Bounded canonical lexer work

Each canonical analysis pass admits at most 4,096 lexer work units. For each lexer call, real lexer errors consume units in listener occurrence order before the next token returned by that call; the returned token then consumes one unit when it is not EOF. EOF does not count. The analyzer detects the event that would consume unit 4,097 before parser prediction, records that first omitted token or lexer error location, and stops. First omitted event wins. It emits no partial stages, scopes, references, lineage, dependencies, or requirements evidence from the abandoned parse.

When SPL2 reaches EOF within 4,096 units, it performs its existing unmatched-literal-mode closure inspection on the collected tokens before constructing a parser. The single existing synthetic unterminated-literal error is ordered after the last non-EOF token and before EOF and consumes one unit. If that synthetic error would consume unit 4,097, analysis returns the resource-limit outcome at the earliest unmatched opener. If token or real lexer-error overflow already occurred before EOF, closure inspection does not run. A synthetic closure error within the budget remains an ordinary syntax diagnostic and parsing proceeds. Neither dialect constructs a parser object until lexer preflight and SPL2 closure accounting admit the document.

An over-budget normalized document is a content outcome rather than an API error. Go `Analyze` returns a non-nil `analysis.Result` and a nil error. The result retains the normalized document and full supplied text, uses `status: incomplete`, sets both `coverage.syntax_complete` and `coverage.semantic_complete` to false, and has `coverage.reasons: ["SPL_ANALYSIS_RESOURCE_LIMIT"]`. Stages, scopes, references, and lineage are initialized empty arrays. Every dependency collection is an initialized empty array. The result contains exactly one diagnostic with code `SPL_ANALYSIS_RESOURCE_LIMIT`, severity `warning`, category `resource_limit`, message `analysis stopped before parser prediction after reaching the 4,096-unit lexer work limit`, empty stage and scope IDs, and the exact location of the first omitted token or lexer error.

Locations use the analyzer's existing source-indexed, half-open rune ranges. An omitted returned token uses `[token.Start, token.Stop+1)`, clamped by `sourceIndex`. A real lexer error uses `[lexer input index, lexer input index+1)`, clamped so an EOF error is `[len,len)`. The synthetic SPL2 closure error uses the earliest unmatched opener token range `[Start, Stop+1)`. These rules apply to the first event that would exceed the budget; later events are not inspected.

The embedded requirement set is fully initialized. Its query and capability identities are the same values a normally admitted document would receive, including a `query_digest` over the full supplied text. It uses `query_status: incomplete`, `coverage.complete: false`, and `coverage.reasons: ["SPL_ANALYSIS_RESOURCE_LIMIT"]`. Its items are empty. It has exactly one gap with code `SPL_ANALYSIS_RESOURCE_LIMIT`, message `requirement coverage is incomplete because analysis exceeded the 4,096-unit lexer work limit`, no reference IDs, and `diagnostic_codes: ["SPL_ANALYSIS_RESOURCE_LIMIT"]`. Its diagnostics array contains the same single diagnostic as the analysis result. Every collection remains array-valued when empty. Go `Requirements` returns a detached copy of that set and a nil error.

The 4,096-unit value matches the existing schema projection and enumeration budgets and is supported by the security-review measurements recorded in the implementation plan. It bounds canonical parsing for every product surface. The ASCII dense-fixture response limits used in acceptance are not universal byte guarantees. Arbitrary query text may expand under JSON escaping, and a document below the lexer budget can still amplify inherited lineage evidence. Long sparse input remains admitted when it stays within the work budget.

## Requirement report contract

`RequirementSet` is a new named report family with integer `schema_version: 1`. It contains:

- `query`: compact query identity with `source_id`, normalized `language`, `profile`, `version`, and `query_digest`.
- `capability_revision`: the identity of the selected capability contract.
- `query_status`: the query-only analysis status, using `valid`, `invalid`, or `incomplete`.
- `coverage`: requirement-specific `complete` and ordered `reasons` values.
- `items`: ordered direct requirement items.
- `gaps`: ordered explanations of evidence that prevents complete requirement discovery.
- `diagnostics`: a detached copy of canonical query-only analysis diagnostics.

The requirement set omits full query text. Its digest binds it to the submitted text, and each occurrence retains the spelling and exact source range needed to interpret the evidence.

All collections serialize as arrays, including when empty. Returned sets own their nested slices. Mutation of a standalone set, one analysis result, or one document snapshot must not alter any other result or global capability value.

### Provenance identities

Both identities use the string form `sha256:<64 lowercase hex>`.

`query_digest` is SHA-256 over the exact valid UTF-8 query-text bytes. It does not include `source_id`, language, profile, or compatibility version. Selector defaults are normalized and retained as separate query identity fields. Query text itself is not whitespace-normalized, repaired, or line-ending-normalized.

`capability_revision` is SHA-256 over the bytes produced by Go `encoding/json` for the fully normalized typed `CapabilityManifest`, with no indentation or trailing newline. The digest includes its language, profile, version, documentation snapshot, commands, functions, limitations, and rewrite manifest when present. It includes no environment state or later corpus evidence. Milestone 9 may enrich the manifest while preserving this revision mechanism.

### Requirement items

Each item has:

- `id`: `req-N`, assigned after deterministic ordering.
- `kind`: the canonical reference kind.
- `identity`: the analyzer's normalized identity.
- `role`: the canonical consuming role.
- `necessity`: `required` or `conditional`.
- `origin`: exactly `direct` in Milestone 8.
- `resolution`: the canonical exact, wildcard, or dynamic resolution state.
- `occurrences`: ordered source evidence.

An occurrence contains `reference_id`, `original_name`, query-structural `binding`, `stage_id`, `scope_id`, and the exact `location`. In a target-refined parent analysis, the occurrence binding may differ from the public refined reference binding. The occurrence always states the query-only requirement classification.

Items group by `(kind, identity, role, resolution)`. Groups follow the first canonical reference occurrence, then explicit start offset, end offset, kind, role, identity, and resolution tie-breakers. Occurrences remain in canonical reference order. An item is required when at least one occurrence is a definite direct external obligation. Otherwise it is conditional.

For fields, inclusion depends on whether the reference consumes an existing source-bound value:

- Exact source-bound consuming references are required.
- Indeterminate consuming references are conditional and produce a gap.
- Wildcard or dynamic consuming references are conditional and produce a gap, even if a refinement target can enumerate matches.
- Query-derived fields are omitted. Their references remain in the parent analysis.
- Create and output definitions, rename targets, removals, and null tests are not external obligations.
- A query-local unavailable field is omitted. Its analysis diagnostic describes the invalid query; it is not converted into an environment requirement.

The current analyzer represents a rename source with a source-bound read and its target with the non-consuming rename reference. The source read can become a requirement; the target cannot.

Exact direct knowledge-object references are required items. Initial kinds are `index`, `source`, `sourcetype`, `dataset`, `data_model`, `lookup`, and `macro`, matching canonical analyzer names. Wildcard or dynamic identities are conditional when the analyzer has a defensible identity. If it does not, the limitation appears only as a gap. A macro name can be an exact required item while unresolved macro expansion independently produces an incomplete-coverage gap.

### Coverage, gaps, and query status

Requirement coverage is independent of query status. `coverage.complete` is false when query-only syntax or semantics are incomplete, or when any requirement evidence is indeterminate, wildcard, or dynamic. A fully understood query-local defect can therefore yield `query_status: invalid` with `coverage.complete: true`. For example, reading a field after the query removed it is an invalid query, but it does not create an unknown external obligation.

Each gap has `code`, `message`, ordered `reference_ids`, and ordered `diagnostic_codes`. Links are based on canonical ownership recorded during analysis. The projector must not infer links from overlapping or adjacent locations.

Diagnostic-driven gaps reuse the existing canonical diagnostic code when it directly states the requirement limitation. Milestone 8 adds only these requirement-specific codes:

- `SPL_REQUIREMENT_INDETERMINATE` for a query-only obligation whose source or derived origin cannot be proved.
- `SPL_REQUIREMENT_DYNAMIC` for dynamic requirement evidence not already expressed by a canonical diagnostic.
- `SPL_REQUIREMENT_COVERAGE_INCOMPLETE` when a query-only coverage flag is incomplete but no more specific diagnostic or requirement fact explains it.

Wildcard gaps use `SPL_UNRESOLVED_WILDCARD` when that canonical diagnostic exists. Syntax, unknown-command, unsupported-semantics, and similar gaps keep their existing diagnostic codes. The projector deduplicates gaps by code plus exact ordered evidence links. `coverage.reasons` contains first-seen unique gap codes in gap order.

## Product interfaces

The standalone operations are:

| Interface | Operation |
|---|---|
| Go | `analysis.Requirements(analysis.QueryDocument) (*analysis.RequirementSet, error)` |
| CLI | `spl-toolkit requirements [query]` |
| REST | `POST /api/v1/query/requirements` |
| Native/C | `spl_mapper_requirements_query` |
| Python | `SPLMapper.requirements_query(...)` |

The Python method uses the same keyword-only `language`, `profile`, `version`, and `source_id` parameters as `analyze_query`. The C operation accepts the strict query-document JSON shape and returns an owned `SPLResult` released by `spl_result_free`.

The CLI mirrors the existing `analyze` input boundary: one positional or `--query` value, compatibility selectors, optional `--source-id`, `--format text|json`, and optional `--output`. Milestone 8 does not add batch, file, or stdin input to either command. Text output prints query status and requirement coverage separately, followed by items, gaps, and diagnostics.

Portable CLI resource-limit acceptance uses the compact deterministic query `strings.Repeat("a ", 2048) + "a"`, or its byte-identical fixture equivalent. It contains 2,049 identifier tokens and 2,048 hidden whitespace tokens, for exactly 4,097 lexer work units. It is exactly 4,097 ASCII bytes and UTF-16 code units, well below the Windows 32,767 UTF-16 command-line boundary. The 64 KiB and 256 KiB dense fixtures and the 300,000-byte long-sparse fixture do not travel through CLI argv. They remain mandatory through Go, HTTP, raw C, Python, an installed wheel, and a rebuilt sdist, whose body, stdin, or in-process transports admit those sizes. This operating-system argv constraint is a transport-only test constraint. It does not lower or otherwise change the canonical 4,096-unit analyzer budget, and it does not authorize file, stdin, or batch input for either CLI command.

The `requirements` CLI exit codes are:

| Code | Meaning |
|---|---|
| 0 | Query valid and requirement coverage complete |
| 1 | Query status invalid |
| 2 | Request, option, output, or internal failure |
| 3 | Query status incomplete or requirement coverage incomplete |

The existing `analyze` command retains status-only process exits, so a valid analysis with incomplete embedded requirement coverage exits `0`; an over-budget analysis still exits `3` because its `analysis.Status` is `incomplete`.

An over-budget query follows the existing successful content paths for both operations. The `analyze` and `requirements` CLI commands emit their canonical values and exit `3`. Their REST routes return HTTP 200. Native/C returns owned successful `SPLResult` values, and Python returns the decoded successful values through its owned native-result paths. The existing REST request boundary remains separate: a body larger than 1 MiB returns HTTP 400 before analysis.

REST uses the existing strict query-document decoder, content-type policy, middleware, and 1 MiB body limit. Valid, invalid, and incomplete content returns HTTP 200 with the canonical set. Malformed JSON, invalid Unicode, duplicate or unknown properties, and unsupported selectors return HTTP 400. No server-side file or network access is introduced.

Native and Python preserve existing mapper-handle admission, operation guards, owned-result allocation, UTF-8 behavior, error propagation, and finally-based freeing. Adapters perform no requirement classification or gap construction. Full report parity ignores only JSON object-key order.

Dense raw-C and Python acceptance runs in helper subprocesses rather than the pytest parent. The parent sends the complete probe request through stdin, applies an explicit timeout, captures the structured result, and always reaps the helper. A timeout or abnormal exit fails acceptance after the parent terminates the helper and, if necessary, kills and reaps it. The helper retains the same complete-value equality checks, exactly-once `spl_result_free` checks, mapper ownership checks, repeated calls, and concurrent-call checks as the in-process API contract.

## Propagation through existing reports

Every newly produced `analysis.Result` emits its embedded requirement set. Validation, rewrite, corpus, and impact reports that already serialize `*analysis.Result` inherit the field without new requirement-specific orchestration.

`document.Snapshot` adds and deeply copies the requirement set because it promises a detached view of canonical analysis evidence. Milestone 8 does not add requirement-specific graph nodes, graph edges, SARIF rules, corpus aggregates, impact comparison logic, or LSP behavior. Those projections retain their current responsibilities.

Rewrite stops immediately when normalized request preparation and the original `PrepareRewrite` produce a resource-limited session. It does not select rules, construct or validate a candidate, or run verification. The operation returns a contract-compatible incomplete no-op result and a nil error. The result has `schema_version: 1`, the normalized document, the requested mode, `status: incomplete`, and coverage with syntax, semantic, and rewrite completeness false. Validation and candidate validation are omitted. Coverage reasons are sorted and deduplicated exactly as `["SPL_ANALYSIS_RESOURCE_LIMIT", "post_verification_failed"]`.

The result's `original_text`, `candidate_text`, and `text` all equal the full normalized document text; `committed` is false; and `changes` is empty. It contains one rule evaluation per prepared rule in request order. Every evaluation uses outcome `skipped`, reason `post_verification_failed`, an empty `reference_ids` array, and omitted location and condition. `original_analysis` is the exact limited analysis. `candidate_analysis` is a deeply detached equal clone of it made without another `Analyze` call. Preview and apply differ only in `mode`, and apply never commits. Batch preserves input report order and combines these incomplete reports through the existing status precedence.

## Machine contracts and compatibility

Publish `contracts/v1/requirements.schema.json` with shared definitions for `RequirementSet`, query identity, coverage, item, occurrence, and gap. The root schema requires the complete Milestone 8 shape.

The existing analysis v1 and document-snapshot schemas add `requirements` as an allowed property but do not add it to their v1 `required` arrays. New runtime reports always emit the property. Keeping it optional in those existing schemas allows archived pre-Milestone-8 version-1 reports to continue validating under the additive-output compatibility policy. Analysis retains `schema_version: 1`.

Update the contract registry, OpenAPI output, native header and source manifests, Python package data, release-content manifests, and current parity fixtures through their established tooling. Current golden outputs that serialize a live `analysis.Result` require deliberate updates for the additive field. Do not rewrite historical evidence receipts or make historical outputs claim Milestone 8 data.

## Security and operational boundaries

Requirement extraction is deterministic and offline. It accepts only the query document already admitted by analysis. It opens no path, follows no URL, reads no environment variable, consults no global registry, executes no query, and stores no credential or event data.

The REST route retains bounded request handling and strict decoding. Canonical analysis additionally limits lexer work to 4,096 non-EOF tokens and lexer errors across SPL and SPL2, before parser prediction. Native and Python retain existing memory ownership and concurrency guards. Capability and query digests identify supplied data; they are not authentication, authorization, signatures, or proof that a principal can execute a query.

The bound responds to an adversarial dense-wildcard measurement. At the pre-Milestone-8 commit `f02009d`, a 65,533-byte query took 5.28 seconds, used 203 MB, and serialized 12.8 MB of JSON. At `d1db37c`, the same query took 5.65 seconds, used 317 MB, and serialized 27.0 MB of JSON, including 14.2 MB of requirements. A 256 KiB query exceeded 60 seconds. These measurements justify stopping canonical parsing at a fixed lexer-work boundary while keeping the 1 MiB REST transport limit unchanged.

## Testing and verification

Durable requirement fixtures live outside `_build_plan/`. Core tests cover:

- Exact source fields and every current knowledge-object kind.
- Repeated identities under different roles and repeated occurrence grouping.
- Derived fields, create/output definitions, rename targets, removals, null tests, and unavailable query-local fields.
- Unknown commands, malformed syntax, macros, indeterminate bindings, wildcard and dynamic references, and fallback coverage gaps.
- An invalid query with complete requirement coverage.
- SPL2 SQL, mixed pipelines, and dotted source identities.
- Unicode and CRLF locations, exact query digests, stable and distinct capability revisions, deterministic ordering, empty arrays, deep detachment, and concurrent calls.
- Exact 4,096-unit admission and 4,097-unit rejection for SPL and SPL2, with lexer errors counted as work units.
- The full-text digest, first-omitted-event location, deterministic JSON, single diagnostic and gap, empty canonical evidence, long sparse admission, and an end-to-end dense adversarial benchmark.
- Real lexer-error listener order, returned-token order, SPL2 synthetic closure accounting, exact half-open omitted-event ranges, and proof that no parser object is constructed before admission.
- O(1) pending-reference and incomplete-stage indexes that preserve existing trace order, environment-clone behavior, remapping, and serialized output.
- Exact preview, apply, and batch rewrite no-op values for a resource-limited original, including detached candidate analysis without a second analysis call.

Refinement-isolation tests compare byte-equivalent embedded `RequirementSet` values for plain analysis, field-list refinement, partial source-universe refinement, JSON Schema and OCSF validation, and original or candidate rewrite analysis of the same normalized query. The cases must include wildcard selectors and dotted SPL2 source identities. No test may obtain this parity by running a second query analysis inside the implementation.

Cross-surface acceptance submits representative valid, invalid, and incomplete SPL and SPL2 through Go, CLI, a real loopback HTTP server, native C/Python, and installed Python packages. It compares complete requirement-set values, including provenance, occurrences, gaps, diagnostics, and empty collections. It also checks:

- `Analyze(...).Requirements` equals `Requirements(...)` for the same document.
- CLI text fields and all four exit outcomes.
- HTTP content outcomes versus request failures, strict Unicode handling, and body limits.
- Over-budget parity through the compact 4,097-work-unit CLI fixture, including CLI exit `3`, and through the 64 KiB and 256 KiB dense fixtures for Go, REST below the 1 MiB body boundary, raw C, Python, installed wheel, and rebuilt sdist. The same non-CLI surfaces admit the 300,000-byte long-sparse fixture. Dense native and Python probes, including concurrency, use the stdin-fed helper-subprocess boundary described above.
- For the ASCII dense 64 KiB and 256 KiB fixtures, serialized `RequirementSet` output no larger than 4,096 bytes and serialized `Result` output no larger than the full query byte length plus 4,096 bytes, stable across repeated and concurrent runs.
- Invalid and closed native handles, owned-result freeing, repeated calls, and concurrent calls.
- Direct-wheel and rebuilt-sdist installations outside the checkout with no source-path injection.
- Positive and negative JSON Schema instances for the new contract and compatibility checks for archived analysis v1 reports.
- OpenAPI regeneration idempotence, native header/source closure, documentation checks, release-content closure, and absence of `_build_plan/` from runtime and package inputs.

The dense-fixture assertion is an exact semantic oracle. Cross-surface equality alone is insufficient. For the exact 64 KiB, or 65,536-byte, and 256 KiB fixtures it requires `schema_version: 1` on the analysis and requirement reports; the full normalized document with the supplied text and source ID plus `profile: splunkd` and `version: current`; the complete normalized requirement query identity; and the exact capability revision. Source IDs are `dense-spl-65536.spl`, `dense-spl-262144.spl`, `dense-spl2-65536.spl`, and `dense-spl2-262144.spl`. SPL uses capability revision `sha256:dfb8cedde04204e0a876412fbe217e49405689b54ae7d7d8fc37bfcb7fb2335f`. SPL2 uses `sha256:c3217502697cee2595f20d2fe98b76422f837696d2861cee81e852ad142edcfc`. The exact query digests are:

- SPL, 64 KiB: `sha256:4ef76589e31ca84b778eb3e15f8cd320319dd03746fff5bf7ba21ba66865dace`.
- SPL, 256 KiB: `sha256:64ec67a0b6bcecae397863ac2b0336f8c59fa39cadb134fd2757b3d68b66e502`.
- SPL2, 64 KiB: `sha256:c5b1c0ba7bc519bde8181f7b8469f84fb4a04e35e16cc37d027a67a3477fcf3b`.
- SPL2, 256 KiB: `sha256:ad99be139221b8721e1d2319d4be832e2df915731a1f33e45444320a9115c445`.

The analysis dependency object has exactly the keys `indexes`, `sources`, `source_types`, `data_models`, `lookups`, `macros`, and `datasets`, each mapped to an empty array. Analysis status and requirement query status are `incomplete`; analysis syntax and semantic coverage are false; requirement coverage is false; and both coverage reason arrays contain only `SPL_ANALYSIS_RESOURCE_LIMIT`. Stages, scopes, references, lineage, requirement items, and all dependency values are empty arrays. The analysis embeds the exact standalone requirement set.

Each dense result contains exactly one diagnostic. It has code `SPL_ANALYSIS_RESOURCE_LIMIT`, severity `warning`, category `resource_limit`, message `analysis stopped before parser prediction after reaching the 4,096-unit lexer work limit`, and empty stage and scope IDs. Both dense sizes use the same first-omitted location for their language. SPL uses start `{offset: 7026, line: 1, column: 7027}` and end `{offset: 7027, line: 1, column: 7028}`. SPL2 uses start `{offset: 7564, line: 1, column: 7565}` and end `{offset: 7565, line: 1, column: 7566}`. The requirement report contains that same diagnostic and exactly one gap with code `SPL_ANALYSIS_RESOURCE_LIMIT`, message `requirement coverage is incomplete because analysis exceeded the 4,096-unit lexer work limit`, empty `reference_ids`, and `diagnostic_codes: ["SPL_ANALYSIS_RESOURCE_LIMIT"]`. No partial evidence is permitted.

Run the established Go 1.22 floor checks, current-Go formatting, vet and race checks, native source tests, package isolation suites, contract validator, documentation checks, and the repository's main CI pipeline. Green local tests do not substitute for the requested main pipeline result.

## Deferred fidelity defect

The release behavior is the production analyzer at `de296da4093e37eae67d953792ba4be6517dca83`. Later planning commits do not change that code. Its SPL2 structural-reference remediation keeps a per-name snapshot of uncertainty so an adjacent quoted atomic field such as `'actor.name'` remains distinct from grammar-level navigation such as `actor.name`. The snapshot does not record whether a different incomplete event occurred before a later same-name quoted read.

The exact known reproductions are:

- `FROM main | eval x=actor.name+foo(1)+'actor.name'`, where an unsupported function occurs between the structural and quoted reads.
- `FROM main | eval x=actor.name+other.value+'actor.name'`, where a different structural navigation occurs between them.
- `FROM main AS actor WHERE actor.name=1 SELECT 'actor.name'`, where the structural read occurs in the SQL filter phase and the quoted read occurs in the later select phase.

In these already-incomplete reports, the stale snapshot can classify the final quoted occurrence as `source`, make the grouped requirement item `required`, and install ordinary field state for the quoted name. That classification is too optimistic within the incomplete evidence. It does not make the report or requirement coverage complete: the existing structural, unsupported-function, different-name, or SQL-phase diagnostics and gaps remain, `status` and `query_status` remain `incomplete`, and requirement coverage remains false.

This defect has no false-complete, security, resource, data-loss, ABI, or package impact. It changes no parser admission bound, execution behavior, stored data, public signature, serialized schema, native ownership rule, or distribution content. The authored requirement corpus remains exactly 20 cases, with `testdata/requirements/cases.json` SHA-256 `663387c481472dee92a5478296f98b28c89101785293648e2c21176f2eee2ba0`.

Milestone 8 delivery is blocked by a material error in a normal supported query, false completeness, a security, resource, or data-loss defect, or a break in a published surface. A rare fidelity defect that occurs only inside an already-incomplete report is recorded for later correction and does not block this milestone. The cause-aware generation design in the living implementation plan remains future guidance and is not part of Milestone 8 acceptance.

## Permanent documentation

Update the repository README, `docs/API.md`, `docs/cli.md`, `docs/architecture.md`, compatibility guidance, and Python README. The documentation must show both operations, the complete report shape, Go and Python examples, CLI and REST examples, digest rules, item grouping, source versus derived behavior, independent query and coverage statuses, exit behavior, and the milestone's exclusions.

Permanent documentation and tests must not link to or load this design file. User-facing limits must state that requirements are direct and query-only, dynamic or unsupported behavior remains incomplete, no knowledge-object expansion occurs, and a requirement set does not prove environment compatibility or runtime execution.

## Decision and deviation history

On 2026-09-15, Task 8 quality review replaced the original uniform dense-payload surface matrix. Passing 64 KiB and 256 KiB queries directly in CLI argv is not portable to Windows, whose command-line boundary is 32,767 UTF-16 code units. Adding file, stdin, or batch input would change the approved CLI contract. The corrected design therefore keeps the public CLI contract and proves its resource-limit content path with the compact 4,097-work-unit, 4,097-code-unit query, while retaining the full dense and long-sparse matrix on transports that admit those payloads. This is a test-transport deviation only; the analyzer still enforces one 4,096-unit lexer budget on every canonical analysis call.

The same review found that dense acceptance compared surfaces and checked only part of the resource-limit shape. The corrected design pins schema versions, normalized identity, capability revisions, query digests, dependency keys, the complete diagnostic and gap, exact omitted ranges, and absence of partial evidence so a shared adapter defect cannot become its own oracle.

The review also found that raw native calls ran inside the pytest process, including thread-pool probes. A native crash or hang could therefore terminate or strand the acceptance runner before it reported which boundary failed. The corrected design moves dense raw-C and Python probes into stdin-fed helper subprocesses with parent-enforced timeouts and clean termination while preserving equality, allocation, free, handle, and concurrency assertions.

On 2026-09-16, three independent reviews confirmed the stale structural-snapshot reproductions recorded above. Scope review classified them as a deferred fidelity defect because every reproduction is already incomplete and retains the diagnostic evidence that prevents false completeness. The release-blocker threshold is limited to material errors in normal supported queries, false completeness, security, resource, or data-loss defects, and published-surface breaks. The production analyzer at `de296da4093e37eae67d953792ba4be6517dca83` remains the release behavior; no cause-aware tracker is required for Milestone 8.

## Acceptance criteria

Milestone 8 is complete when:

1. Representative SPL and SPL2 documents produce equivalent standalone requirement sets through Go, CLI, REST, native/C, and Python. Portable resource-limit acceptance uses the compact 4,097-work-unit query for CLI and the full 64 KiB, 256 KiB, and 300,000-byte long-sparse fixtures for Go, REST, raw C, Python, installed wheel, and rebuilt sdist.
2. Every new analysis result embeds a set equal to the standalone query-only set, including inside refinement-aware reports.
3. Source, conditional, dynamic, and indeterminate external obligations are explicit. Derived and query-local unavailable fields are not misclassified as obligations: `Analyze` retains their canonical references, while standalone `Requirements` returns only the external-obligation projection. The documented stale structural-snapshot edge is the accepted deferred exception because it remains visibly incomplete.
4. Provenance identities, ordering, links, diagnostics, and empty collections are deterministic and deeply detached.
5. Current contracts, installed packages, native closure, documentation, and main CI pass their required checks while archived version-1 reports remain valid.
6. Canonical SPL and SPL2 parsing stops before parser construction or prediction on work unit 4,097 and returns the exact bounded incomplete report through each applicable transport. Exact-limit, compact portable CLI, lexer-error ordering, SPL2 closure, range, sparse-input, 64 KiB, 256 KiB, deterministic-output, and concurrent adversarial checks pass without partial canonical evidence. Dense Python and raw-C probes run in stdin-fed helper subprocesses with parent timeouts and clean termination. For the two ASCII dense acceptance fixtures, the serialized requirement set is at most 4,096 bytes and the serialized analysis result is at most the full query byte length plus 4,096 bytes.
7. No environment snapshot, compatibility assessment, transitive expansion, placeholder resolution, query fanout, broad language family, universal serialized-result bound, or `_build_plan/` runtime dependency is introduced.
