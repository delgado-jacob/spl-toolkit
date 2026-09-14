# Milestone 8 - Product interface and canonical requirements

Status: approved architectural design. This document defines Milestone 8 only. It does not authorize implementation outside the milestone, and nothing under `_build_plan/` may become a runtime, build, package, test-data, or documentation-generation dependency.

## Purpose and scope

Milestone 8 adds two product-level query operations to the existing canonical analysis system:

- `analyze(query)` returns the current structured meaning, diagnostics, coverage, provenance, and an embedded requirement set.
- `requirements(query)` returns that same first-class requirement set without requiring callers to consume the rest of the analysis report.

`pkg/analysis` remains the canonical Go implementation. The milestone extends that package instead of introducing a facade or a second requirements package. CLI, REST, native/C, and Python adapters delegate to the canonical operations and preserve the same report values.

The requirement set describes direct external obligations visible in the submitted query. It does not load environment metadata, inspect event data, expand supplied knowledge-object definitions, validate compatibility, resolve placeholders, generate query variants, or add broad SPL or SPL2 language coverage. Transitive requirements, environment snapshots, compatibility, and resolution remain owned by later milestones.

## Architecture

Canonical SPL or SPL2 analysis performs its existing parse, field-flow, dependency, reference, diagnostic, coverage, and status work. After reference IDs and query-only analysis evidence are final, a deterministic projector builds the requirement set. The projector consumes canonical evidence. It does not parse query text, replay field transfers, or guess identities from source spans.

`analysis.Result` always contains a non-optional `Requirements` value. `analysis.Requirements(document)` performs one canonical analysis pass and returns a deeply detached copy of the set embedded in that pass. Calling `Analyze` gives callers meaning and requirements together without repeated work.

### Query-only requirement trace

Some existing operations call `AnalyzeWithSourceFields` or `AnalyzeWithSourceUniverse`. Field-list validation, schema validation, and rewrite proof may refine public reference bindings, wildcard membership, diagnostics, and coverage using a caller-supplied target. Milestone 8 requirements must remain a function of the query and selected language capability contract only.

The analysis pass therefore maintains a private requirement trace alongside its public refinement-aware evidence. The trace records the complete query-only consuming references, diagnostics, status inputs, and completeness inputs before source-universe admission changes them. Reference finalization remaps trace links to the same final `ref-N` identifiers used by the public result. The projector uses this trace, not target-refined conclusions.

For plain `Analyze`, trace evidence and public analysis evidence coincide. For validation and rewrite analysis, the embedded requirement set must be byte-equivalent to the set from plain analysis of the same normalized query. It excludes refinement-target diagnostics and proofs. This trace is a projection feed within the existing traversal, not a second parser or semantic engine.

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

CLI exit codes are:

| Code | Meaning |
|---|---|
| 0 | Query valid and requirement coverage complete |
| 1 | Query status invalid |
| 2 | Request, option, output, or internal failure |
| 3 | Query status incomplete or requirement coverage incomplete |

REST uses the existing strict query-document decoder, content-type policy, middleware, and 1 MiB body limit. Valid, invalid, and incomplete content returns HTTP 200 with the canonical set. Malformed JSON, invalid Unicode, duplicate or unknown properties, and unsupported selectors return HTTP 400. No server-side file or network access is introduced.

Native and Python preserve existing mapper-handle admission, operation guards, owned-result allocation, UTF-8 behavior, error propagation, and finally-based freeing. Adapters perform no requirement classification or gap construction. Full report parity ignores only JSON object-key order.

## Propagation through existing reports

Every newly produced `analysis.Result` emits its embedded requirement set. Validation, rewrite, corpus, and impact reports that already serialize `*analysis.Result` inherit the field without new requirement-specific orchestration.

`document.Snapshot` adds and deeply copies the requirement set because it promises a detached view of canonical analysis evidence. Milestone 8 does not add requirement-specific graph nodes, graph edges, SARIF rules, corpus aggregates, impact comparison logic, or LSP behavior. Those projections retain their current responsibilities.

## Machine contracts and compatibility

Publish `contracts/v1/requirements.schema.json` with shared definitions for `RequirementSet`, query identity, coverage, item, occurrence, and gap. The root schema requires the complete Milestone 8 shape.

The existing analysis v1 and document-snapshot schemas add `requirements` as an allowed property but do not add it to their v1 `required` arrays. New runtime reports always emit the property. Keeping it optional in those existing schemas allows archived pre-Milestone-8 version-1 reports to continue validating under the additive-output compatibility policy. Analysis retains `schema_version: 1`.

Update the contract registry, OpenAPI output, native header and source manifests, Python package data, release-content manifests, and current parity fixtures through their established tooling. Current golden outputs that serialize a live `analysis.Result` require deliberate updates for the additive field. Do not rewrite historical evidence receipts or make historical outputs claim Milestone 8 data.

## Security and operational boundaries

Requirement extraction is deterministic and offline. It accepts only the query document already admitted by analysis. It opens no path, follows no URL, reads no environment variable, consults no global registry, executes no query, and stores no credential or event data.

The REST route retains bounded request handling and strict decoding. Native and Python retain existing memory ownership and concurrency guards. Capability and query digests identify supplied data; they are not authentication, authorization, signatures, or proof that a principal can execute a query.

## Testing and verification

Durable requirement fixtures live outside `_build_plan/`. Core tests cover:

- Exact source fields and every current knowledge-object kind.
- Repeated identities under different roles and repeated occurrence grouping.
- Derived fields, create/output definitions, rename targets, removals, null tests, and unavailable query-local fields.
- Unknown commands, malformed syntax, macros, indeterminate bindings, wildcard and dynamic references, and fallback coverage gaps.
- An invalid query with complete requirement coverage.
- SPL2 SQL, mixed pipelines, and dotted source identities.
- Unicode and CRLF locations, exact query digests, stable and distinct capability revisions, deterministic ordering, empty arrays, deep detachment, and concurrent calls.

Refinement-isolation tests compare byte-equivalent embedded `RequirementSet` values for plain analysis, field-list refinement, partial source-universe refinement, JSON Schema and OCSF validation, and original or candidate rewrite analysis of the same normalized query. The cases must include wildcard selectors and dotted SPL2 source identities. No test may obtain this parity by running a second query analysis inside the implementation.

Cross-surface acceptance submits representative valid, invalid, and incomplete SPL and SPL2 through Go, CLI, a real loopback HTTP server, native C/Python, and installed Python packages. It compares complete requirement-set values, including provenance, occurrences, gaps, diagnostics, and empty collections. It also checks:

- `Analyze(...).Requirements` equals `Requirements(...)` for the same document.
- CLI text fields and all four exit outcomes.
- HTTP content outcomes versus request failures, strict Unicode handling, and body limits.
- Invalid and closed native handles, owned-result freeing, repeated calls, and concurrent calls.
- Direct-wheel and rebuilt-sdist installations outside the checkout with no source-path injection.
- Positive and negative JSON Schema instances for the new contract and compatibility checks for archived analysis v1 reports.
- OpenAPI regeneration idempotence, native header/source closure, documentation checks, release-content closure, and absence of `_build_plan/` from runtime and package inputs.

Run the established Go 1.22 floor checks, current-Go formatting, vet and race checks, native source tests, package isolation suites, contract validator, documentation checks, and the repository's main CI pipeline. Green local tests do not substitute for the requested main pipeline result.

## Permanent documentation

Update the repository README, `docs/API.md`, `docs/cli.md`, `docs/architecture.md`, compatibility guidance, and Python README. The documentation must show both operations, the complete report shape, Go and Python examples, CLI and REST examples, digest rules, item grouping, source versus derived behavior, independent query and coverage statuses, exit behavior, and the milestone's exclusions.

Permanent documentation and tests must not link to or load this design file. User-facing limits must state that requirements are direct and query-only, dynamic or unsupported behavior remains incomplete, no knowledge-object expansion occurs, and a requirement set does not prove environment compatibility or runtime execution.

## Acceptance criteria

Milestone 8 is complete when:

1. Representative SPL and SPL2 documents produce equivalent standalone requirement sets through Go, CLI, REST, native/C, and Python.
2. Every new analysis result embeds a set equal to the standalone query-only set, including inside refinement-aware reports.
3. Source obligations, query-derived fields, query-local invalidity, conditional evidence, dynamic identities, and incomplete coverage remain distinguishable without inference from omission.
4. Provenance identities, ordering, links, diagnostics, and empty collections are deterministic and deeply detached.
5. Current contracts, installed packages, native closure, documentation, and main CI pass their required checks while archived version-1 reports remain valid.
6. No environment snapshot, compatibility assessment, transitive expansion, placeholder resolution, query fanout, broad language family, or `_build_plan/` runtime dependency is introduced.
