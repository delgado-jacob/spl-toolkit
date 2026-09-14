# Feature Implementation Plan: Milestone 8 Product Interface and Requirements API

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development to implement this plan task-by-task. Mark completed Progress items and keep all living sections current.

**Goal:** Make query requirements a canonical, deterministic product interface across Go, CLI, REST, C, Python, contracts, packages, and permanent documentation without changing the meaning of existing analysis evidence.

**Architecture:** Extend the single analysis pass with a private query-only semantic trace that survives refinement without inheriting target-dependent facts. Project the public `RequirementSet` only after canonical reference finalization, embed it in every runtime `analysis.Result`, and expose detached copies through thin adapters. Contracts preserve archived v1 compatibility by making the new property optional when decoding stored analysis and document snapshots, while every current runtime producer emits it.

**Tech Stack:** Go 1.22, Cobra, `net/http`, cgo and a stable C ABI, Python 3.11 through 3.14 with `ctypes`, JSON Schema draft 2020-12, Swag/OpenAPI generation, pytest, Go race tests, AddressSanitizer, GitHub Actions, and Claude CLI review.

---

This file is a living execution plan. The implementing orchestrator must update `Progress`, `Surprises & Discoveries`, `Decision Log`, and `Outcomes & Retrospective` as work proceeds. The approved behavioral source is `_build_plan/milestones/8-product-interface-requirements/design-spec.md`; this plan restates the required behavior so implementers and reviewers must not load `_build_plan/` at runtime or from tests, package manifests, release archives, generated documentation, or installed distributions.

## Progress

- [x] (2026-09-14) Read the approved Milestone 8 design specification, current source and tests, release tooling, GitHub Actions, and relevant prior milestone plans.
- [x] (2026-09-14) Resolve the public model, single-pass refinement strategy, adapter boundaries, compatibility rules, fixture ownership, review protocol, and release procedure in this plan.
- [ ] Task 1: Add public requirement value types, digest helpers, constants, and deep-copy primitives with focused tests.
- [ ] Task 2: Capture a private query-only trace in the existing analysis traversal and prove refinement cannot contaminate it.
- [ ] Task 3: Project, embed, and expose canonical requirements and update the current analysis corpus.
- [ ] Task 4: Propagate requirements through document snapshots and prove refinement and downstream product parity.
- [ ] Task 5: Add CLI and REST product surfaces with exact status and transport behavior.
- [ ] Task 6: Add C and Python product surfaces with ownership and native-memory coverage.
- [ ] Task 7: Publish machine contracts, OpenAPI, registry, native-source, and release manifests.
- [ ] Task 8: Add permanent documentation, cross-surface acceptance, installed-package checks, and release closure.
- [ ] Complete independent final specification, code-quality, security, and architecture reviews, including Claude Opus, and resolve every actionable finding.
- [ ] Complete an independent acceptance validation at the exact feature-branch head.
- [ ] Push the feature branch, dispatch and pass GitHub Actions at the pushed SHA, fast-forward `main`, push it, and pass the `main` pipeline at the merge SHA.
- [ ] Record final evidence and lessons in `Outcomes & Retrospective`.

## Surprises & Discoveries

- The existing analysis reference pipeline assigns public `ref-N` identifiers only in `finalizeReferences`. The requirements trace must therefore hold pending reference identity and use that same final mapping. Matching diagnostics back to references by location would be ambiguous and is prohibited.
- Source refinement intentionally changes public field binding and diagnostics. A projector over the final public `Result` would make requirements target-dependent, so query-only field provenance and diagnostic completeness must be captured during the original semantic traversal.
- The feature branch is based on two documentation commits beyond `origin/main`. The eventual feature diff and CI `headSha` checks must include those commits, and landing must remain a fast-forward.
- GitHub Actions does not run this repository's CI workflow on an ordinary feature-branch push. The branch pipeline must be started explicitly with `workflow_dispatch` after pushing.
- Existing current fixtures embed complete analysis results in analysis, validation, schema, and CLI acceptance data. Runtime embedding of `requirements` requires deliberate regeneration of those current goldens, but historical release evidence and receipts must remain byte-for-byte untouched.

## Decision Log

- **Decision:** Keep `analysis.Result.Requirements` a non-pointer value and initialize every nested slice to a non-nil empty slice. **Rationale:** Every runtime analysis must emit the product view, and stable empty arrays avoid surface-specific `null` values.
- **Decision:** Implement `analysis.Requirements(QueryDocument)` by calling `Analyze` exactly once and deep-copying the embedded value. **Rationale:** This makes `Analyze` the only analysis authority and gives callers a detached result without replay.
- **Decision:** Compute `query_digest` from the exact UTF-8 query text bytes only and `capability_revision` from compact JSON encoding of the fully normalized typed `CapabilityManifest`. **Rationale:** The digests then have one deterministic implementation rather than adapter-specific serialization.
- **Decision:** Maintain a private query-only environment beside the public semantic environment. It follows the same traversal and control-flow clones but ignores refinement target facts. **Rationale:** Requirements from refined reports must be byte-identical to plain analysis without a second parse, traversal, or analysis call.
- **Decision:** Carry explicit private diagnostic-to-reference ownership links and event ordinals. **Rationale:** Gap construction cannot infer ownership by overlapping locations, and event order gives deterministic gap ordering.
- **Decision:** A wildcard or dynamic obligation uses an existing linked `SPL_UNRESOLVED_WILDCARD` diagnostic when one exists; otherwise the requirement gap uses `SPL_REQUIREMENT_DYNAMIC`. An indeterminate obligation uses `SPL_REQUIREMENT_INDETERMINATE`. A final uncovered incomplete state uses `SPL_REQUIREMENT_COVERAGE_INCOMPLETE`. **Rationale:** This reuses canonical analysis evidence and introduces only the three approved requirement codes.
- **Decision:** Add `requirements` to analysis and document-snapshot schema properties but not to their v1 `required` arrays. **Rationale:** Current producers always emit the field while archived v1 payloads remain decodable.
- **Decision:** Run implementation serially in the shared feature worktree. Each task receives an implementer, then a specification reviewer, then a quality reviewer, with fix and rereview loops before the next task. **Rationale:** The tasks share analysis and fixture files, so concurrent writers would create unsafe overlap.
- **Decision:** Do not create a pull request. Push the branch, run the dispatchable workflow directly, then fast-forward and push `main` only after all local, review, acceptance, and branch-CI gates pass. **Rationale:** This follows the requested hosted workflow without adding an unrequested PR artifact.

## Outcomes & Retrospective

Implementation has not started. At completion, replace this paragraph with the delivered behavior, exact implementation and review commit SHAs, local verification results, Claude review disposition, acceptance evidence, branch and `main` GitHub Actions run URLs and conclusions, merge SHA, any deviations from this plan, and remaining follow-up work. Do not claim runtime or installed-package evidence that was not actually run.

## User-visible behavior

Milestone 8 adds two equivalent product operations:

    func Analyze(document QueryDocument) (*Result, error)
    func Requirements(document QueryDocument) (*RequirementSet, error)

`Analyze` continues to return all existing canonical evidence and additionally embeds `requirements`. `Requirements` performs one canonical analysis and returns a deep detached copy of that embedded value. No adapter may parse SPL, walk an AST, derive evidence independently, or call analysis more than once.

The public model in `pkg/analysis/requirements.go` is:

```go
type RequirementQueryIdentity struct {
	SourceID    string `json:"source_id"`
	Language    string `json:"language"`
	Profile     string `json:"profile"`
	Version     string `json:"version"`
	QueryDigest string `json:"query_digest"`
}

type RequirementCoverage struct {
	Complete bool     `json:"complete"`
	Reasons  []string `json:"reasons"`
}

type RequirementOccurrence struct {
	ReferenceID  string   `json:"reference_id"`
	OriginalName string   `json:"original_name"`
	Binding      string   `json:"binding"`
	StageID      string   `json:"stage_id"`
	ScopeID      string   `json:"scope_id"`
	Location     Location `json:"location"`
}

type RequirementItem struct {
	ID          string                  `json:"id"`
	Kind        string                  `json:"kind"`
	Identity    string                  `json:"identity"`
	Role        string                  `json:"role"`
	Necessity   string                  `json:"necessity"`
	Origin      string                  `json:"origin"`
	Resolution  string                  `json:"resolution"`
	Occurrences []RequirementOccurrence `json:"occurrences"`
}

type RequirementGap struct {
	Code            string   `json:"code"`
	Message         string   `json:"message"`
	ReferenceIDs    []string `json:"reference_ids"`
	DiagnosticCodes []string `json:"diagnostic_codes"`
}

type RequirementSet struct {
	SchemaVersion     int                      `json:"schema_version"`
	Query             RequirementQueryIdentity `json:"query"`
	CapabilityRevision string                   `json:"capability_revision"`
	QueryStatus       Status                   `json:"query_status"`
	Coverage          RequirementCoverage      `json:"coverage"`
	Items             []RequirementItem        `json:"items"`
	Gaps              []RequirementGap         `json:"gaps"`
	Diagnostics       []Diagnostic             `json:"diagnostics"`
}
```

Before committing Task 1, run `gofmt` and align the fields visually; do not rename JSON keys or replace the named query and coverage types with maps. Add this exact field to `analysis.Result` in Task 3:

```go
Requirements RequirementSet `json:"requirements"`
```

Add the package-qualified value to `document.Snapshot` in Task 4:

```go
Requirements analysis.RequirementSet `json:"requirements"`
```

The v1 requirement rules are:

1. `schema_version` is `1`. `query_digest` and `capability_revision` are `sha256:` followed by 64 lowercase hexadecimal characters. The query digest hashes `QueryDocument.Text` exactly, after UTF-8 validation but before selector normalization. The capability digest hashes the bytes from `json.Marshal(manifest)`, where `manifest` is the typed value returned by `CapabilitiesFor(CapabilityOptions{Language: document.Language, Profile: document.Profile, Version: document.Version})` after normalizing selector defaults.
2. Items group canonical direct external obligations by `(kind, identity, role, resolution)`. Each item receives `req-1`, `req-2`, and so on only after deterministic sorting. Item order is first reference order, then occurrence location, kind, role, identity, and resolution. Occurrences are in canonical reference order.
3. `origin` is always `direct`. `necessity` is `required` when any occurrence is a definite direct external obligation and otherwise `conditional`.
4. Required exact fields are those consumed from an external source. Indeterminate fields remain conditional with a gap. Wildcard or dynamic field reads remain conditional with a gap. Derived fields, create/output/rename targets, removals, null tests, and unavailable local fields are omitted. A rename source can be an external requirement.
5. Knowledge kinds are `index`, `source`, `sourcetype`, `dataset`, `data_model`, `lookup`, and `macro`. Exact knowledge objects are required. Wildcard or dynamic objects become conditional when an identity is defensible, otherwise they produce only a gap. An exact macro is an item plus an expansion gap.
6. `query_status` is the canonical query-only `valid`, `invalid`, or `incomplete` status. Requirement coverage is independent. An invalid query can still have complete requirement coverage, such as a read that follows removal.
7. A gap has `code`, `message`, ordered `reference_ids`, and ordered `diagnostic_codes`. Deduplicate by code plus the exact ordered reference and diagnostic links. `coverage.reasons` contains first-seen unique gap codes in gap order. Create ownership links in the semantic event that owns the evidence. Never use source-location overlap to connect them.
8. Requirement diagnostics are canonical query-only diagnostics. Refinement-only diagnostics never enter the embedded or standalone requirement set. The only new diagnostic codes are `SPL_REQUIREMENT_INDETERMINATE`, `SPL_REQUIREMENT_DYNAMIC`, and `SPL_REQUIREMENT_COVERAGE_INCOMPLETE`.

The CLI command is `spl-toolkit requirements [query]`. It accepts one positional or `--query` value, compatibility selectors, optional `--source-id`, `--format text|json`, and optional `--output`. It does not add batch, file, or stdin input. Text mode prints status, coverage, items, gaps, and diagnostics; JSON mode emits the exact `RequirementSet`. Exit codes are `0` for valid plus complete, `1` for invalid, `2` for request, option, output, or internal failure, and `3` when query status is incomplete or requirement coverage is incomplete.

The REST operation is `POST /api/v1/query/requirements`. It uses the shared strict `QueryDocument` decoder, accepted content type rules, and 1 MiB body limit. Content outcomes return HTTP 200 even when invalid or incomplete; malformed or oversized requests return 400. The endpoint performs no I/O.

The native ABI exports:

```c
SPLResult* spl_mapper_requirements_query(int mapperID, char* documentJSON);
```

It accepts one strict `QueryDocument` JSON object and returns owned JSON through the existing `SPLResult` allocation and `spl_result_free` lifecycle. Python exposes:

```python
def requirements_query(
    self,
    query: str,
    *,
    language: str = "spl",
    profile: str = "splunkd",
    version: str = "current",
    source_id: str = "",
) -> dict[str, Any]:
    document = {"text": query, "language": language, "profile": profile,
                "version": version, "source_id": source_id}
    return self._validate_fields_request(
        self._lib.spl_mapper_requirements_query, document, operation="requirements")
```

The method uses the same validation, JSON boundary, native error mapping, and `finally`-based free path as `analyze_query`.

## Internal architecture and invariants

Add `pkg/analysis/requirements_trace.go` for private evidence. Names may be adjusted to existing unexported style, but the responsibilities and data flow are fixed:

```go
type requirementTrace struct {
	references  []requirementTraceReference
	diagnostics []requirementTraceDiagnostic
	syntaxComplete   bool
	semanticComplete bool
	nextEventOrdinal int
}

type requirementTraceReference struct {
	pendingID string
	reference Reference
	directExternal bool
	conditional    bool
	eventOrdinal   int
}

type requirementTraceDiagnostic struct {
	diagnostic Diagnostic
	incomplete bool
	pendingReferenceIDs []string
	eventOrdinal int
}

type requirementField struct {
	source      bool
	unavailable bool
}

type requirementEnvironment struct {
	fields    map[string]requirementField
	removed   map[string]bool
	open      bool
	uncertain bool
}
```

`requirementTraceReference.reference.ID` remains pending until `finalizeReferences` obtains the same pending-to-public `ref-N` map used by the public result. `finalizeReferences` remaps trace references and explicit diagnostic links, then asserts that each trace reference has the same canonical identity, kind, role, location, stage, and scope as its public counterpart. Trace field state clones wherever the existing `environment` clones for branches or scopes.

At each semantic event, record public and trace evidence together. Parser diagnostics seed both public diagnostics and query trace diagnostics. Normal semantic diagnostics record their `incomplete` classification and explicit pending reference owners in the trace. Refinement-only diagnostics use a separate helper and never enter the trace. The trace records baseline wildcard, indeterminate, source, field, knowledge-object, and macro-expansion uncertainty before target refinement can resolve or alter public facts.

The projector runs only after reference finalization. It derives query status and completeness from the trace, builds ordered items and gaps, copies query-only diagnostics, and stores the result on `Result.Requirements`. It does not inspect AST nodes, call `Analyze`, invoke the parser, rerun transfers, or infer diagnostic ownership from locations. `Requirements(document)` calls `Analyze(document)` once and returns `cloneRequirementSet(result.Requirements)`. `document.New` uses the same deep-clone helper or an equivalent package-safe clone; mutation tests must prove every nested slice is detached.

## Orchestration and review protocol

The orchestrator starts each implementation task only after the preceding task and its reviews are accepted. Fresh subagents receive the approved design specification, this plan, the exact current task, the starting and ending SHAs, and a statement that `_build_plan/` is planning input only. They must use `superpowers:test-driven-development` and must not change files owned by later tasks unless a failing dependency makes the listed scope impossible.

For every Task 1 through Task 8:

1. Record the immutable starting SHA and assign one fresh implementation agent. The agent writes the named failing test first, runs it to observe the expected failure, makes the smallest production change, reruns the focused test, then runs the task verification commands. It updates the living sections in this plan and creates the task's scoped commit.
2. Record the implementation SHA and assign a fresh specification reviewer. The reviewer reads the approved specification and reviews `starting-sha..implementation-sha` for missing, extra, or contradictory behavior. The reviewer does not edit.
3. Only after specification approval, assign a different fresh quality reviewer. The reviewer examines the same immutable range for correctness, robustness, deterministic behavior, error paths, security, memory ownership, architecture, test quality, and unnecessary scope. The reviewer does not edit.
4. If either reviewer finds an actionable issue, assign a fresh fix agent with the exact findings. The fix agent uses TDD where behavior changes, makes a new scoped commit, and returns the new head. Repeat both reviews over `starting-sha..new-head` until both approve.
5. Update `Progress`, `Surprises & Discoveries`, and `Decision Log` with evidence and accepted deviations. Release the next task only after the current range is approved.

Do not run two code-writing agents concurrently in this worktree. Review agents may inspect immutable ranges concurrently only when neither writes. No agent may amend another agent's commit, rewrite history, stash user changes, or reset the worktree.

## Task 1: Public model, digests, constants, and cloning

**Owned files:** `pkg/analysis/requirements.go`, `pkg/analysis/requirements_test.go`, and `pkg/analysis/diagnostics.go`. Do not add `Requirements` to `Result` yet.

1. Add compile-time tests for the exact public types and JSON names above. Add table tests that hash empty text, non-ASCII UTF-8, line endings, and documents with identical text but different selectors. Assert query digests change only with exact text bytes and match `sha256:<64 lowercase hex>`.
2. Add capability revision tests across selector aliases and repeated calls. Equivalent normalized selectors must match, materially different supported selectors must differ, and the digest must equal SHA-256 over the compact JSON bytes from the typed normalized manifest.
3. Add `cloneRequirementSet` tests that populate every nested slice and mutate the clone's coverage reasons, item occurrences, gap reference and diagnostic code arrays, and diagnostics. The source must remain unchanged and all empty arrays must serialize as `[]` rather than `null`.
4. Run the tests before implementation. Expected evidence is compile failure for missing requirement types and helpers. Implement the value types, digest helpers, clone helper, and exactly three new requirement diagnostic constants.
5. Run:

       env GOWORK=off GOCACHE=/private/tmp/spl-toolkit-m8-gocache go test -mod=readonly ./pkg/analysis -run 'TestRequirementTypes|TestQueryDigest|TestCapabilityRevision|TestCloneRequirementSet' -count=1
       env GOWORK=off GOCACHE=/private/tmp/spl-toolkit-m8-gocache go test -mod=readonly ./pkg/analysis -count=1
       git diff --check

   Expected evidence is all tests passing and no whitespace errors.
6. Commit only the owned files with `git commit -m "feat(analysis): define requirement product model"`.

## Task 2: Single-pass query-only trace

**Owned files:** `pkg/analysis/requirements_trace.go`, `pkg/analysis/requirements_trace_test.go`, and the minimal trace plumbing in `pkg/analysis/analyze.go`, `pkg/analysis/flow.go`, `pkg/analysis/references.go`, `pkg/analysis/transfers.go`, `pkg/analysis/source_fields.go`, `pkg/analysis/dependencies.go`, `pkg/analysis/scopes.go`, and `pkg/analysis/spl2_lower.go`. If an exact filename differs, locate the existing owner with `rg` and record the correction in `Surprises & Discoveries` before editing.

1. Add internal tests that observe trace evidence through package-private helpers. Cover direct field reads, derived fields, rename sources and targets, removals, null tests, unavailable local fields, exact and wildcard knowledge objects, macros, syntax diagnostics, unsupported semantics, branching, subsearch scopes, and dotted SPL2 identifiers.
2. Add paired plain and source-refined tests for field lists, partial universes, JSON Schema, OCSF, original-query rewrite, and candidate rewrite. The private trace and its finalized reference links must be identical in each pair even when public bindings or diagnostics differ.
3. Instrument the existing traversal. Initialize one trace in the analysis context, carry a query-only environment beside public field state, clone it at the same branch and scope boundaries, and record both views at the same semantic events. Add an explicit refinement-only diagnostic path. Do not call the parser, lowerer, analyzer, or projector a second time.
4. Extend `finalizeReferences` to remap trace pending IDs with the canonical pending-to-public map and validate correspondence. Keep diagnostic ownership as explicit pending IDs established by the owning semantic operation.
5. Assert each private trace reference is recorded at one semantic event and receives exactly one canonical public ID. Review the implementation call graph to prove it contains no recursive `Analyze`, parser, lowerer, or transfer invocation. The first run should fail because the trace does not exist; the passing run must prove no replay was added.
6. Run:

       env GOWORK=off GOCACHE=/private/tmp/spl-toolkit-m8-gocache go test -mod=readonly ./pkg/analysis -run 'TestRequirementTrace|TestRequirementTraceRefinementParity|TestRequirementTraceSinglePass' -count=1
       env GOWORK=off GOCACHE=/private/tmp/spl-toolkit-m8-gocache go test -mod=readonly ./pkg/analysis -count=1
       git diff --check

   Expected evidence is byte-equivalent query traces across refinement pairs, one traversal per case, and a green analysis package.
7. Commit with `git commit -m "feat(analysis): capture query requirement evidence"`.

## Task 3: Projection, embedding, standalone Go API, and analysis corpus

**Owned files:** `pkg/analysis/model.go`, `pkg/analysis/analyze.go`, `pkg/analysis/requirements.go`, `pkg/analysis/requirements_test.go`, `pkg/analysis/corpus_test.go`, `testdata/requirements/cases.json`, and `testdata/analysis/cases.json`.

1. Create `testdata/requirements/cases.json` as the durable shared product fixture. Include valid, invalid, and incomplete SPL and SPL2; exact and derived fields; indeterminate and wildcard fields; direct knowledge kinds; dynamic identities; macros; removals; null tests; rename sources and targets; branching; and empty results. Expected values must be full `RequirementSet` objects, not fragments.
2. Add projector tests for exact grouping, first-reference and occurrence ordering, `req-N` numbering, necessity folding, field inclusion and omission, knowledge-kind handling, gap codes, deduplication by code and ordered evidence links, first-seen unique coverage reasons, diagnostic ordering, query status, independent coverage, and non-nil arrays. Add a case where an invalid read-after-removal query has complete requirement coverage.
3. Add `Requirements RequirementSet` to `Result`. Invoke the projector once at the end of canonical analysis after references and diagnostics are finalized. Ensure every return path either returns an error or a fully populated embedded value.
4. Add `Requirements(document)` with exactly this control flow: call `Analyze(document)` once, return its error unchanged, deep-clone `result.Requirements` into a local value, and return that value's address. Add deep-detachment tests and compare its JSON bytes with the embedded `Analyze(document).Requirements` bytes for every shared fixture. The implementation contains no second call or fallback path.
5. Update only the current `testdata/analysis/cases.json` expected results to include `requirements`. Use the existing corpus update mechanism if present; otherwise make a narrow mechanical update from current Go output and review the diff. Do not edit any historical evidence, receipts, or release archives.
6. Run:

       env GOWORK=off GOCACHE=/private/tmp/spl-toolkit-m8-gocache go test -mod=readonly ./pkg/analysis -run 'TestRequirementProjection|TestRequirements|TestAnalysisCorpus' -count=1
       env GOWORK=off GOCACHE=/private/tmp/spl-toolkit-m8-gocache go test -mod=readonly ./pkg/analysis -count=1
       git diff --check

   Expected evidence is exact fixture parity, a single analysis invocation, deterministic repeat runs, and a green analysis corpus.
7. Commit with `git commit -m "feat(analysis): project canonical query requirements"`.

## Task 4: Refinement and downstream product propagation

**Owned files:** `pkg/analysis/requirements_refinement_test.go`, `pkg/analysis/source_fields_test.go`, focused existing `pkg/analysis/rewrite_*_test.go` files only when their existing assertions embed full analysis results, `pkg/document/model.go`, `pkg/document/snapshot.go`, `pkg/document/snapshot_test.go`, focused `pkg/validation/*_test.go` files, `pkg/rewrite/requirements_test.go`, `pkg/corpus/requirements_test.go`, `testdata/validation/cases.json`, and `testdata/schemas/cases.json`. Changes to production validation, rewrite, or corpus logic require a demonstrated propagation bug and a `Decision Log` entry.

1. Add a refinement parity table that runs plain analysis and every supported field-source refinement mode for field list, partial universe, JSON Schema, OCSF, original-query rewrite, and candidate rewrite. Assert byte equality of the full embedded `requirements`, not selected fields. Include wildcard, unresolved wildcard, dotted SPL2, syntax error, and unsupported semantics.
2. Add snapshot tests that assert `document.New(result)` copies the complete requirement set. Mutate all nested slices in the source and snapshot independently and prove no aliasing in either direction.
3. Add field-validation and schema-validation tests that compare requirements inherited through their embedded analysis result with direct `analysis.Requirements` for the same query. Add focused propagation tests in new `pkg/rewrite/requirements_test.go` and `pkg/corpus/requirements_test.go` for the existing `*analysis.Result` fields. Confirm no requirement-specific aggregate, graph node or edge, SARIF rule, impact comparison, or LSP behavior is added.
4. Update current validation and schema goldens that embed complete analysis results. Keep rewrite expectations unchanged unless they already store full analysis JSON; provenance strings such as `"phase":"analysis"` are not a reason to change a fixture.
5. Run:

       env GOWORK=off GOCACHE=/private/tmp/spl-toolkit-m8-gocache go test -mod=readonly ./pkg/analysis -run 'Test.*Requirement.*Refinement|Test.*Requirement.*Rewrite' -count=1
       env GOWORK=off GOCACHE=/private/tmp/spl-toolkit-m8-gocache go test -mod=readonly ./pkg/document ./pkg/validation ./pkg/rewrite ./pkg/corpus -count=1
       env GOWORK=off GOCACHE=/private/tmp/spl-toolkit-m8-gocache go test -mod=readonly ./pkg/... -count=1
       git diff --check

   Expected evidence is byte-for-byte plain/refined parity and detached propagation through every existing result-bearing product.
6. Commit with `git commit -m "feat(products): propagate canonical requirements"`.

## Task 5: CLI and REST adapters

**Owned files:** `cmd/analysis.go`, `cmd/analysis_test.go`, `cmd/cli.go`, `cmd/cli_test.go`, `cmd/main.go`, `pkg/api/analysis.go`, `pkg/api/analysis_test.go`, `pkg/api/server.go`, and `pkg/api/server_test.go`.

1. Add failing CLI tests for positional query text, `--query`, shared selector flags including `--source-id`, JSON output, deterministic text sections, `--output`, and the exact exit table. Assert file, stdin, and batch inputs are rejected. Also assert help lists `requirements` and invalid option, output, and internal failures return `2` without partial JSON.
2. Refactor only enough shared adapter code to prevent drift between `analyze` and `requirements`. Call `analysis.Requirements` once. Text output must always show query status and requirement coverage, then ordered items, gaps, and diagnostics without recomputing order.
3. Add failing real-handler REST tests for valid, invalid, and incomplete SPL and SPL2; strict duplicate and unknown-field rejection; empty and malformed JSON; invalid Unicode; wrong content type; trailing JSON; unsupported selectors; and bodies at and beyond 1 MiB. Assert content outcomes are HTTP 200 and request-boundary failures are HTTP 400.
4. Register `POST /api/v1/query/requirements` through the same server composition as other query endpoints. Reuse the strict `QueryDocument` request parser and write the `RequirementSet` directly. Assert no filesystem or network collaborator is invoked.
5. Run:

       env GOWORK=off GOCACHE=/private/tmp/spl-toolkit-m8-gocache go test -mod=readonly ./cmd -run 'Test.*Requirements' -count=1
       env GOWORK=off GOCACHE=/private/tmp/spl-toolkit-m8-gocache go test -mod=readonly ./pkg/api -run 'Test.*Requirements' -count=1
       env GOWORK=off GOCACHE=/private/tmp/spl-toolkit-m8-gocache go test -mod=readonly ./cmd ./pkg/api -count=1
       git diff --check

   Expected evidence is exact JSON parity with the Go API, stable text, exact process exits, strict transport behavior, and no adapter-owned analysis logic.
6. Commit with `git commit -m "feat(adapters): expose requirement CLI and REST APIs"`.

## Task 6: Native C ABI and Python API

**Owned files:** `pkg/bindings/bindings.go`, `pkg/bindings/requirements_test.go`, `python/spl_toolkit/mapper.py`, `python/spl_toolkit/libspl_toolkit.h`, `python/tests/test_native_analysis.py`, `python/tests/test_native_requirements.py`, `python/tests/test_native_abi.py`, and `tests/native/memory.c`.

1. Add failing binding tests for the `spl_mapper_requirements_query` symbol, strict `QueryDocument` JSON, valid, invalid, and incomplete results, embedded-versus-standalone parity, invalid and closed handles, malformed input errors, and repeated allocate/free cycles.
2. Implement the C export as a thin wrapper over `analysis.Requirements` and `ownedMapperJSONResult`. Regenerate the header using `make build-shared`; copy the generated declaration into the source header only if that is the repository's established flow, then rerun generation and require no diff.
3. Add failing Python tests for the exact keyword-only signature, selector defaults and validation, non-ASCII input, JSON parity, native error mapping, repeated and concurrent calls, close waiting for an admitted call, and unconditional `spl_result_free` on success, content failure, JSON decode failure, and raised Python exceptions.
4. Register `argtypes` and `restype` in `_setup_function_signatures` and implement `SPLMapper.requirements_query` using the same request builder and lifecycle as `analyze_query`. Do not add a Python-side fallback or projector.
5. Extend `tests/native/memory.c` to exercise successful and failing requirement calls in the ASAN loop, including freeing every owned result exactly once. Extend ABI export expectations in `python/tests/test_native_abi.py`.
6. Run:

       env GOWORK=off GOCACHE=/private/tmp/spl-toolkit-m8-gocache go test -mod=readonly ./pkg/bindings -count=1
       make build-shared
       python3 -m pytest -q python/tests/test_native_analysis.py python/tests/test_native_requirements.py python/tests/test_native_abi.py
       make build-shared
       cmp build/libspl_toolkit.h python/spl_toolkit/libspl_toolkit.h
       git diff --check

   Expected evidence is ABI parity, exactly-once frees, concurrent-call safety, and an idempotently generated header. On Linux with GCC 13, also run the exact `native-memory` commands from `.github/workflows/ci.yml`: build the shared library with Go `-asan`, compile `tests/native/memory.c` with `-fsanitize=address`, and run with `ASAN_OPTIONS=detect_leaks=1:halt_on_error=1`. On other hosts, record that this gate is CI-only and require the exact-SHA CI `native-memory` job before acceptance.
7. Commit with `git commit -m "feat(native): expose canonical query requirements"`.

## Task 7: JSON Schema, OpenAPI, registries, and source manifests

**Owned files:** `contracts/v1/requirements.schema.json`, `contracts/v1/shared.schema.json`, `contracts/README.md`, `tests/acceptance/test_machine_contracts.py`, `testdata/tooling/contracts.json`, `tools/update_validation_openapi.py`, its focused tests, generated `docs/docs.go`, `docs/swagger.json`, `docs/swagger.yaml`, `python/native-source-files.txt`, `tools/release-content-files.txt`, `tools/release-source-files.txt`, and manifest expectation tests that fail because of these additions.

1. Add `RequirementQueryIdentity`, `RequirementCoverage`, `RequirementOccurrence`, `RequirementItem`, `RequirementGap`, and `RequirementSet` definitions to the shared v1 schema. Require every field inside a requirement set and preserve the repository's additive object-property policy. Constrain enums, `^req-[1-9][0-9]*$`, canonical reference IDs, and `^sha256:[0-9a-f]{64}$` where applicable.
2. Add `contracts/v1/requirements.schema.json` as a root reference to the shared `RequirementSet`. Add `requirements` properties to analysis `Result` and document `Snapshot`, but leave both existing v1 required arrays unchanged. Add positive and negative contract fixtures for missing new-set fields, bad enum values, bad IDs, bad digests, unknown properties, and archived analysis and snapshot documents that omit requirements.
3. Register the family in `tests/acceptance/test_machine_contracts.py` and document it in `contracts/README.md`. Assert the registry runs offline and resolves only repository-local `$ref` targets.
4. Add Swag annotations or generator normalization only where needed to publish `/api/v1/query/requirements` and the exact response model. Run `make generate-docs`, inspect the operation and schemas, rerun generation, and require a clean second generation.
5. Add every new production Go source and contract file to `python/native-source-files.txt`, `tools/release-source-files.txt`, and `tools/release-content-files.txt` according to current ownership. Never add `_build_plan/`, test-only files to runtime manifests, or requirement schema files to Python package data unless they already fall under the package's existing contract-data policy.
6. Run:

       python3 -m pytest -q tests/acceptance/test_machine_contracts.py
       python3 -m pytest -q tools/tests/test_update_validation_openapi.py tools/tests/test_release.py
       make generate-docs
       git diff --exit-code -- docs/docs.go docs/swagger.json docs/swagger.yaml
       git diff --check

   Expected evidence is offline schema validation, archived-v1 compatibility, deterministic OpenAPI, complete manifest test coverage, and no `_build_plan` path in any runtime or distribution manifest.
7. Commit with `git commit -m "feat(contracts): publish requirement interface contracts"`.

## Task 8: Permanent docs, packaged acceptance, and release closure

**Owned files:** `README.md`, `docs/API.md`, `docs/cli.md`, `docs/architecture.md`, `docs/compatibility.md`, `python/README.md`, `tests/acceptance/test_requirements_surfaces.py`, `tests/acceptance/cli_examples.json`, `python/tests/test_native_requirements.py`, `tools/check_package.py`, `tools/check_acceptance.py`, their focused tests, and the minimum fixture-copy and hash plumbing for `testdata/requirements/cases.json`.

1. Add one cross-surface acceptance suite driven by `testdata/requirements/cases.json`. For every valid, invalid, and incomplete SPL and SPL2 case, compare canonical JSON from Go, CLI, real HTTP, native C through Python, and direct Python to the expected fixture and to the embedded analysis requirements.
2. Extend CLI examples with requirements help, positional and `--query` input, text, JSON, output-file, input rejection, and exit-code cases. Because analysis JSON now embeds requirements, update only exact current examples affected by the runtime output. Preserve all unrelated examples.
3. Teach `tools/check_package.py` to copy the requirement fixture outside the checkout, provide a dedicated `SPL_REQUIREMENTS_FIXTURES` environment variable, hash the fixture in evidence, include the new native and acceptance tests in the source distribution, and run them against both an installed wheel and sdist. Extend its unit tests first.
4. Teach `tools/check_acceptance.py` to require the new cross-surface and native requirement tests without weakening existing minimum counts or evidence checks. Extend its unit tests first. Installed checks must not import from the repository checkout or read `_build_plan/`.
5. Document the operation, fields, evidence rules, exit and HTTP behavior, ABI ownership, Python signature, deterministic/offline guarantee, refinement parity, archived-v1 compatibility, and explicit exclusions in every named permanent document. State that the digests identify supplied data and are not authentication, authorization, signatures, or proof that a principal can execute a query. Do not link permanent docs to `_build_plan/`.
6. Run focused acceptance and documentation checks:

       python3 -m pytest -q tests/acceptance/test_requirements_surfaces.py
       python3 -m pytest -q tools/tests/test_package.py tools/tests/test_acceptance.py
       python3 tools/check_docs.py
       git diff --unified=0 origin/main -- README.md docs cmd internal pkg python tools tests contracts go.mod go.sum | rg '^\+.*_build_plan'

   Expected evidence is cross-surface byte parity, strict out-of-checkout fixtures, green documentation checks, and no newly added `_build_plan` reference in runtime, test, package, contract, or permanent-documentation files. Existing historical evidence and negative dependency assertions may retain their current references.
7. Run the repository-level local gates:

       make fmt
       git diff --check
       python3 tools/check_go.py
       env GOWORK=off GOCACHE=/private/tmp/spl-toolkit-m8-race-gocache go test -mod=readonly -race ./...
       make python-test
       python3 -m pytest -q tools/tests/test_package.py tools/tests/test_acceptance.py tools/tests/test_release.py
       python3 tools/check_docs.py

   Expected evidence is every local gate passing. The Linux GCC 13 AddressSanitizer and full multi-target evidence-aggregation gates remain mandatory exact-SHA CI checks. Record environment-limited failures precisely and do not treat CI as a silent substitute. Any deterministic source, fixture, or generated-file diff after a second run is a defect.
8. Commit with `git commit -m "docs: complete requirement product acceptance"`.

## Final independent review and revision gate

After Task 8 is accepted, freeze `HEAD` as the candidate SHA and run four independent subagent reviews over `origin/main...HEAD`: full design-spec compliance, code quality and robustness, security and native-memory safety, and architecture and repository organization. Each reviewer must inspect source, tests, contracts, generated artifacts, package manifests, and permanent docs, and must return findings with exact file and line evidence. Reviewers do not edit.

Run the available external reviewer from the feature worktree:

    claude --model opus --print "Review git diff origin/main...HEAD in this repository for Milestone 8 requirement-product specification compliance, correctness, robustness, security, native memory ownership, deterministic behavior, architecture, test adequacy, packaging, and documentation. Do not modify files. Return actionable findings with file and line evidence, followed by residual risks."

Record its output and exit status in this plan. Do not waive a finding because it came from an external model. Validate it against the approved specification and source. For every confirmed finding, assign a fresh fix agent, require a focused regression test, create a scoped commit, then rerun all affected task reviews and all four final reviews over the new immutable head. Repeat the Claude command after material fixes. Completion requires no unresolved actionable findings; disputed findings require a concrete disposition in `Decision Log`.

Reviewers must explicitly confirm:

1. Query requirements are produced from one canonical traversal and are byte-equivalent across refinement modes and all adapters.
2. Classification, grouping, ordering, gaps, diagnostic ownership, status, coverage, digests, and deep-copy behavior meet every rule in this plan.
3. REST, CLI, C, and Python error and ownership boundaries are exact and do not introduce filesystem, network, injection, unsafe allocation, or use-after-free behavior.
4. Archived v1 analysis and snapshot payloads remain valid, while current runtime payloads always emit requirements.
5. `_build_plan/` is absent from runtime imports, test inputs, package inputs, installed artifacts, release manifests, generated permanent docs, and product documentation links. Tests may retain explicit negative assertions that distributions do not contain it.
6. Historical receipts and evidence were not edited, generated artifacts are reproducible, and the source and release manifests contain exactly the necessary new files.

## Independent acceptance validation

Assign a fresh acceptance agent that did not implement or review the feature. Give it the frozen candidate SHA, this plan's user-visible behavior, and the approved design spec. It must begin from a clean worktree, run the shared requirement fixture across Go, CLI text and JSON, a real HTTP server, native C, Python, installed wheel, and installed sdist, and independently inspect the machine contracts and permanent docs.

The acceptance agent runs at minimum:

    git status --short
    git diff --check origin/main...HEAD
    env GOWORK=off GOCACHE=/private/tmp/spl-toolkit-m8-accept-gocache go test -mod=readonly -race ./...
    make python-test
    python3 -m pytest -q tests/acceptance/test_requirements_surfaces.py tests/acceptance/test_machine_contracts.py
    python3 -m pytest -q tools/tests/test_package.py tools/tests/test_acceptance.py tools/tests/test_release.py
    python3 tools/check_docs.py

It must also manually sample at least one valid, one invalid, and one incomplete SPL and SPL2 case and report response JSON, CLI exit, HTTP status, native/Python parity, and requirement coverage. The acceptance agent must confirm that the candidate's exact-SHA CI `native-memory` job and aggregate acceptance job pass before signing off. Acceptance fails on any unexplained skip, dirty generated artifact, checkout-relative installed-package read, or unverified required surface. Fix failures through the same fresh fix, specification review, quality review, final review, and reacceptance loop.

## Branch CI, fast-forward merge, and post-merge verification

Only after all local, review, Claude, and acceptance gates pass:

1. Confirm the worktree is clean and record the exact candidate SHA:

       git status --short
       git diff --check origin/main...HEAD
       git rev-parse HEAD

   Expected evidence is empty status and diff-check output. If the living plan received final evidence after the last code commit, commit only that plan update as `docs: record milestone 8 acceptance evidence`, rerun the documentation and manifest checks, and freeze the new SHA.
2. Push the feature branch without force:

       git push -u origin codex/milestone-8-product-interface-requirements

3. Dispatch CI because ordinary feature pushes do not trigger the workflow:

       gh workflow run .github/workflows/ci.yml --ref codex/milestone-8-product-interface-requirements
       m8_run_id="$(gh run list --workflow ci.yml --branch codex/milestone-8-product-interface-requirements --event workflow_dispatch --limit 1 --json databaseId --jq '.[0].databaseId')"
       gh run watch "$m8_run_id" --exit-status
       gh run view "$m8_run_id" --json url,headSha,conclusion,event,workflowName

   Verify `headSha` equals the pushed `git rev-parse HEAD`, `event` is `workflow_dispatch`, and `conclusion` is `success`. A successful run at another SHA is not evidence. Diagnose failures from job logs, fix them on the branch, repeat all affected reviews and acceptance, push, and dispatch a new run.
4. Inspect the primary checkout without modifying user work:

       git -C /Users/jmdelgad/repos/spl-toolkit status --short --branch
       git -C /Users/jmdelgad/repos/spl-toolkit fetch origin
       git -C /Users/jmdelgad/repos/spl-toolkit rev-parse --abbrev-ref HEAD
       git -C /Users/jmdelgad/repos/spl-toolkit merge-base --is-ancestor origin/main codex/milestone-8-product-interface-requirements

   Stop and coordinate if it is dirty, not on `main`, or `origin/main` is not an ancestor. Do not stash, reset, rebase, or overwrite user changes.
5. Fast-forward and push `main`:

       git -C /Users/jmdelgad/repos/spl-toolkit merge --ff-only codex/milestone-8-product-interface-requirements
       git -C /Users/jmdelgad/repos/spl-toolkit push origin main

   Record the merge SHA, which must equal the accepted feature SHA.
6. Watch the automatically triggered `main` push workflow at that exact SHA:

       m8_main_run_id="$(gh run list --workflow ci.yml --branch main --event push --limit 1 --json databaseId --jq '.[0].databaseId')"
       gh run watch "$m8_main_run_id" --exit-status
       gh run view "$m8_main_run_id" --json url,headSha,conclusion,event,workflowName
       git -C /Users/jmdelgad/repos/spl-toolkit fetch origin
       git -C /Users/jmdelgad/repos/spl-toolkit rev-parse HEAD origin/main
       git -C /Users/jmdelgad/repos/spl-toolkit status --short --branch

   Verify the run `headSha` and both local and remote `main` SHAs equal the accepted SHA, the conclusion is `success`, and the primary checkout is clean. Record the run URL and outcome in `Outcomes & Retrospective`.

## Acceptance criteria

Milestone 8 is complete only when all of the following are evidenced at the landed SHA:

1. Plain and every supported refinement path produce byte-identical `RequirementSet` values from a single parse, lower, and semantic traversal.
2. Field and knowledge-object classification, ordering, grouping, gaps, diagnostics, digests, status, and coverage match the shared fixture for valid, invalid, and incomplete SPL and SPL2.
3. Go, CLI, REST, C, Python, document snapshot, validation, and schema products expose identical canonical requirement content with exact boundary semantics and detached ownership.
4. The v1 requirement schema is strict, current analysis and snapshot schemas expose the property without invalidating archived payloads, OpenAPI is reproducible, and offline contract validation passes.
5. Wheel and sdist checks run outside the checkout, include and hash the requirement fixture, exercise the real native library, and do not depend on `_build_plan/`.
6. Permanent docs, source manifests, content manifests, and release manifests are complete; historical evidence is unchanged; local gates, independent reviews, Claude Opus review, independent acceptance, feature-branch CI, and post-merge `main` CI all pass at the exact accepted SHA.

## Idempotence and recovery

All generators, fixture updates, formatters, and manifest checks must be safe to rerun and produce no second-run diff. Use fixed cache directories under `/private/tmp` only for build cache, never for source worktrees. Test and package tools must create their own isolated output directories and clean them through their existing mechanisms.

If a focused test fails, retain the failing command and first relevant error in `Surprises & Discoveries`, fix the smallest owning layer, and rerun the focused command before broadening. If generated output drifts, regenerate from its authoritative source, never hand-edit a generated artifact. If a native or installed-package check fails after a partial build, rerun the repository's build target rather than copying ad hoc libraries into the package.

If an implementation agent leaves uncommitted changes, the orchestrator inspects and assigns ownership before continuing. Do not reset or discard them. If a review or CI fix changes behavior, reopen both specification and quality review for the complete task range and rerun acceptance. If branch CI fails after push, add new commits and push normally; never force-push. If `main` advances before landing, stop and re-evaluate ancestry and the plan rather than rebasing or merging silently. If the primary checkout is dirty, stop and ask the user to resolve it.

The analysis fixture, requirement fixture, generated OpenAPI, generated header, contract registry, native source manifest, and release manifests are authoritative checkpoints. Compare them before and after recovery to distinguish a stale build from a semantic change.

## Plan revision note

Created on 2026-09-14 from the approved Milestone 8 design specification and the repository state at `8029a44`. The plan fixes the public types, query-only trace boundary, serial subagent review gates, exact adapters and contracts, packaging and acceptance evidence, and no-PR fast-forward release procedure. Update this note whenever execution changes a public shape, task boundary, verification command, or landing procedure, and explain why in `Decision Log`.
