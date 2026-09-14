# Feature Implementation Plan: Milestone 8 Product Interface and Requirements API

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development to implement this plan task-by-task. Mark completed Progress items and keep all living sections current.

**Goal:** Make query requirements a canonical, deterministic product interface across Go, CLI, REST, C, Python, contracts, packages, and permanent documentation without changing the meaning of existing analysis evidence.

**Architecture:** Extend the single analysis pass with a private query-only semantic trace that survives refinement without inheriting target-dependent facts. Project the public `RequirementSet` only after canonical reference finalization, embed it in every runtime `analysis.Result`, and expose detached copies through thin adapters. Contracts preserve archived v1 compatibility by making the new property optional when decoding stored analysis and document snapshots, while every current runtime producer emits it.

**Tech Stack:** Go 1.22, the repository's custom CLI parser, `net/http`, cgo and a stable C ABI, Python 3.11 through 3.14 with `ctypes`, JSON Schema draft 2020-12, Swag/OpenAPI generation, pytest, Go race tests, AddressSanitizer, GitHub Actions, and Claude CLI review.

---

This file is a living execution plan. The implementing orchestrator must update `Progress`, `Surprises & Discoveries`, `Decision Log`, and `Outcomes & Retrospective` as work proceeds. The approved behavioral source is `_build_plan/milestones/8-product-interface-requirements/design-spec.md`; this plan restates the required behavior so implementers and reviewers must not load `_build_plan/` at runtime or from tests, package manifests, release archives, generated documentation, or installed distributions. This plan and `_build_plan/milestones/8-product-interface-requirements/milestone-log.md` are administrative records and are always allowed task edits even when a task's owned-file list omits them. Finish and commit all administrative updates before final whole-feature review and freeze.

## Progress

- [x] (2026-09-14) Read the approved Milestone 8 design specification, current source and tests, release tooling, GitHub Actions, and relevant prior milestone plans.
- [x] (2026-09-14) Resolve the public model, single-pass refinement strategy, adapter boundaries, compatibility rules, fixture ownership, review protocol, and release procedure in this plan.
- [x] (2026-09-14) Task 1: Added public requirement value types, digest helpers, constants, and deep-copy primitives from base `c5de9590fe41b79a8b1181f73b13b2c92e7105a5`. The first focused run failed on the missing types and `queryDigest`; later RED runs failed on the missing `capabilityRevision`, `cloneRequirementSet`, and constants. The exact focused command and the full `pkg/analysis` package passed after implementation. The completion commit is the commit containing this entry.
- [x] (2026-09-14) Task 2: Captured a private query-only trace during the existing analysis traversal. Focused RED runs exposed the missing trace API, shadow removal and projection state, null-test policy, semantic diagnostic capture, macro ownership, refinement-only diagnostic duplication, SPL2 lexical-stage remapping, child-scope trace loss, qualified catalog classification, and downstream dotted-name contamination. Independent specification review then found incomplete sidecar mirrors and ownership cases. Exact RED witnesses proved a renamed-away source, a field absent from a finite SPL2 dataset literal, a field absent after a sound SQL projection, a field read after a syntax-damaged SPL stage, and a hidden SQL `HAVING` field all had incorrect query-only classifications; the first of two macro diagnostics was also unowned. Those six witnesses, the full `TestRequirementTrace` suite, the complete `pkg/analysis` package, the no-reentry scan, and `git diff --check` pass after correction; the latest correction commit is the commit containing this entry.
- [ ] Task 3: Project, embed, and expose canonical requirements and update the current analysis corpus.
- [ ] Task 4: Propagate requirements through document snapshots and prove refinement and downstream product parity.
- [ ] Task 5: Add CLI and REST product surfaces with exact status and transport behavior.
- [ ] Task 6: Add C and Python product surfaces with ownership and native-memory coverage.
- [ ] Task 7: Publish machine contracts, OpenAPI, registry, native-source, and release manifests.
- [ ] Task 8: Add permanent documentation, cross-surface acceptance, installed-package checks, and release closure.
- [ ] Finish all tracked plan and milestone-log updates and commit the final administrative state before whole-feature review.

After that final tracked update, whole-feature review, local independent acceptance, feature-branch CI, merge, and post-merge CI are terminal gates reported out of band. Their outcomes intentionally do not mutate these Progress checkboxes or any other tracked file after freeze.

## Surprises & Discoveries

- The existing analysis reference pipeline assigns public `ref-N` identifiers only in `finalizeReferences`. The requirements trace must therefore hold pending reference identity and use that same final mapping. Matching diagnostics back to references by location would be ambiguous and is prohibited.
- Source refinement intentionally changes public field binding and diagnostics. A projector over the final public `Result` would make requirements target-dependent, so query-only field provenance and diagnostic completeness must be captured during the original semantic traversal.
- The feature branch is based on two documentation commits beyond `origin/main`. The eventual feature diff and CI `headSha` checks must include those commits, and landing must remain a fast-forward.
- GitHub Actions does not run this repository's CI workflow on an ordinary feature-branch push. The branch pipeline must be started explicitly with `workflow_dispatch` after pushing.
- Existing current fixtures embed complete analysis results in analysis, validation, schema, and CLI acceptance data. Runtime embedding of `requirements` requires deliberate regeneration of those current goldens, but historical release evidence and receipts must remain byte-for-byte untouched.
- `pkg/analysis/transfers_test.go:transferParityWitnesses` and `pkg/rewrite/rewrite_test.go:explicitAliasReport` are independent hand-pinned full-result witnesses outside the JSON fixture directories. Both must gain hand-derived requirements when `analysis.Result` changes.
- The primary checkout intentionally contains only the roadmap inputs listed in the landing procedure as untracked files. It is not an empty-status checkout, and previously removed legacy Markdown and text files must stay removed.
- `CapabilitiesFor` already owns selector validation and default normalization while returning a fresh typed manifest. Task 1 could calculate capability revisions without adding a second selector path, maps, or mutable cache state.
- Attaching the requirement environment to the existing semantic environment makes every existing `clone()` site deep-copy query-only field state while sharing only the analysis-wide trace collector. Independent SPL and SPL2 scopes still require a fresh sidecar that points to the same collector.
- SPL2 child scopes reorder lexical stage identifiers before reference finalization. Returning the already-built private stage map from `spl2FinalizeStages` lets the trace remap its stage links before the reference invariant check; no second stage walk or public API is needed.
- Public field closure does not automatically close the private sidecar. Successful rename removal, a non-empty intact SPL2 dataset literal, and a sound SQL `SELECT` projection each need to mirror their proven output shape explicitly or later local absence is misclassified as an external obligation.
- Linking a macro reference to the latest diagnostic with the same code and stage is ambiguous when one stage contains multiple macros. The reference and its expansion diagnostic must be recorded together for the exact macro context.
- Parser recovery can set public uncertainty outside the ordinary diagnostic helper. SPL stage-boundary recovery and SPL2 recovery phase boundaries must mirror that state into the query-only environment or later reads can be promoted to direct source requirements.
- SQL restricted-expression analysis reads from a temporary visibility environment. Hidden-field diagnostics must use the reference IDs returned by that same expression evaluation before the temporary environment is discarded; emitting the diagnostic first cannot establish exact ownership.

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
- **Decision:** Local independent acceptance signs off before any push and does not depend on hosted CI. After final review freezes the candidate, review, acceptance, and hosted-run results are reported out of band and do not cause tracked evidence commits. **Rationale:** The exact SHA cannot contain a record of checks that run only after that SHA exists.
- **Decision:** Preserve the primary checkout's intentional untracked roadmap baseline exactly across landing. Require a clean tracked index and worktree, check incoming-path collisions, and compare the saved untracked list after the fast-forward. **Rationale:** An untracked roadmap file is user state, not a dirty-tree defect or permission to restore previously removed files.
- **Decision:** Copy every slice layer in `cloneRequirementSet` with an owned empty destination. **Rationale:** This detaches coverage reasons, item occurrences, gap links, and diagnostics while keeping empty cloned collections array-valued instead of `null`.
- **Decision:** Store `requirementEnvironment` as a private sidecar on `environment`; `environment.clone()` deep-copies its maps and origin slices while retaining the shared trace pointer. **Rationale:** Branch and scope semantics remain in the established traversal, refinement mutates only public field state, and trace code does not need a second parser, lowerer, transfer pass, or analysis call.
- **Decision:** Let the private `spl2FinalizeStages` return its existing old-to-final stage map and use it only from `spl2_lower.go` to remap trace references and diagnostics. **Rationale:** SPL2 child scopes otherwise leave trace stage IDs stale after lexical reordering; reusing the existing map preserves single-pass ownership with a three-line private plumbing change.
- **Decision:** Mirror only structurally proven removal and finite output-shape effects into `requirementEnvironment`: successful rename sources become tombstones, intact non-empty dataset literals close their shape, and sound SQL projections retain only selected fields and close uncertainty. **Rationale:** Downstream absent fields are query-local unavailable references, while incomplete or ambiguous effects still retain their established indeterminate behavior.
- **Decision:** Create each SPL macro reference and its owned dynamic-expansion diagnostic in one `macro` semantic event, keyed by the exact parser context so the later dependency walk cannot duplicate it. **Rationale:** Multiple macros in one stage then have exact ownership without code, stage, or location searches.
- **Decision:** Mirror public uncertainty into `requirementEnvironment` at every non-refinement parser and stage recovery boundary, while leaving refinement-only diagnostics out of the private trace. **Rationale:** Query damage makes later field provenance indeterminate, but target-specific evidence must not change standalone requirements.
- **Decision:** For hidden SQL restricted expressions, mark the temporary query-only environment uncertain before evaluating the expression, then attach each incomplete visibility diagnostic to the corresponding returned reference ID. **Rationale:** This classifies the hidden read conditionally and records direct ownership without matching diagnostics back to references by location.

## Outcomes & Retrospective

Tasks 1 and 2 now define the public requirement value model and capture its private query-only evidence in the canonical traversal. The trace owns explicit event order and diagnostic links, remaps through the canonical reference and SPL2 stage maps, and remains byte-equivalent across finite, partial, wildcard, and dotted-name refinement cases. Reviewed regressions now prove query-local unavailability after rename, finite literal, and sound SQL projection effects, conditional provenance after SPL recovery and hidden SQL reads, and exact diagnostic ownership for hidden SQL fields and multiple macros in one stage. `analysis.Result` still has no requirements field, and projection remains unimplemented as required by the Task 2 boundary. Tasks 3 through 8, whole-feature review, local acceptance, hosted CI, merge, and post-merge verification remain open.

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

Use these named helper boundaries so every task and reviewer can trace ownership:

```go
func queryDigest(text string) string
func capabilityRevision(document QueryDocument) (string, error)
func cloneRequirementSet(in RequirementSet) RequirementSet
func newRequirementTrace() *requirementTrace
func (t *requirementTrace) nextEvent() int
func (t *requirementTrace) recordReference(reference Reference, directExternal, conditional bool, eventOrdinal int)
func (t *requirementTrace) recordDiagnostic(diagnostic Diagnostic, incomplete bool, pendingReferenceIDs []string, eventOrdinal int)
func (t *requirementTrace) remapReferences(mapping map[string]string)
func (e requirementEnvironment) clone() requirementEnvironment
func projectRequirements(document QueryDocument, trace *requirementTrace) (RequirementSet, error)
```

`queryDigest` hashes `[]byte(text)` and formats `sha256:<hex>`. `capabilityRevision` normalizes selectors through `CapabilitiesFor`, marshals the returned typed manifest with `json.Marshal`, and hashes those exact bytes. `projectRequirements` consumes only the normalized document and finalized query trace. If implementation discovers that an extra argument is necessary, record the reason before changing the signature and prove the argument cannot expose target-refined evidence.

`requirementTraceReference.reference.ID` remains pending until `finalizeReferences` obtains the same pending-to-public `ref-N` map used by the public result. `finalizeReferences` remaps trace references and explicit diagnostic links, then asserts that each trace reference has the same canonical identity, kind, role, location, stage, and scope as its public counterpart. Trace field state clones wherever the existing `environment` clones for branches or scopes.

At each semantic event, record public and trace evidence together. Parser diagnostics seed both public diagnostics and query trace diagnostics. Normal semantic diagnostics record their `incomplete` classification and explicit pending reference owners in the trace. Refinement-only diagnostics use a separate helper and never enter the trace. The trace records baseline wildcard, indeterminate, source, field, knowledge-object, and macro-expansion uncertainty before target refinement can resolve or alter public facts.

The projector runs only after reference finalization. It derives query status and completeness from the trace, builds ordered items and gaps, copies query-only diagnostics, and stores the result on `Result.Requirements`. It does not inspect AST nodes, call `Analyze`, invoke the parser, rerun transfers, or infer diagnostic ownership from locations. `Requirements(document)` calls `Analyze(document)` once and returns `cloneRequirementSet(result.Requirements)`. `document.New(result, context RevisionContext)` uses an equivalent package-local deep clone when it builds `Snapshot`; mutation tests must prove every nested slice is detached.

## Orchestration and review protocol

The orchestrator starts each implementation task only after the preceding task and its reviews are accepted. Fresh subagents receive the approved design specification, this plan, the exact current task, the starting and ending SHAs, and a statement that `_build_plan/` is planning input only. They must use `superpowers:test-driven-development` and must not change files owned by later tasks unless a failing dependency makes the listed scope impossible.

For every Task 1 through Task 8:

- [ ] Record the concrete 40-character task base SHA and assign one fresh implementation agent. The agent writes the named failing test first, runs it to observe the expected failure, makes the smallest production change, reruns the focused test, then runs the task verification commands. It updates the living sections and creates the task's scoped commit.
- [ ] Record the concrete implementation SHA and assign a fresh specification reviewer. The reviewer reads the approved specification and reviews the exact `TASK_BASE_SHA..TASK_HEAD_SHA` range for missing, extra, or contradictory behavior. The reviewer does not edit.
- [ ] Only after specification approval, assign a different fresh quality reviewer. The reviewer examines the same concrete immutable range for correctness, robustness, deterministic behavior, error paths, security, memory ownership, architecture, test quality, and unnecessary scope. The reviewer does not edit.
- [ ] If either reviewer finds an actionable issue, assign a fresh fix agent with the exact findings. The fix agent uses TDD where behavior changes and makes a new scoped commit. Repeat both reviews over the original concrete base SHA through the new concrete head SHA until both approve.
- [ ] Update `Progress`, `Surprises & Discoveries`, and `Decision Log` with evidence and accepted deviations. Release the next task only after the current range is approved.

Do not run two code-writing agents concurrently in this worktree. Review agents may inspect immutable ranges concurrently only when neither writes. No agent may amend another agent's commit, rewrite history, stash user changes, or reset the worktree.

## Task 1: Public model, digests, constants, and cloning

**Owned files:** `pkg/analysis/requirements.go`, `pkg/analysis/requirements_test.go`, and `pkg/analysis/diagnostics.go`. Do not add `Requirements` to `Result` yet.

- [x] Add `TestRequirementTypesJSONShape`. Construct a fully populated `RequirementSet`, marshal it, and assert the keys and named nested values shown in `User-visible behavior`. The first run must fail to compile because the public types do not exist.
- [x] Add `TestQueryDigestExactBytes` with `""`, `"café 😀"`, `"a\nb"`, and `"a\r\nb"`. Assert the digest equals `"sha256:" + hex.EncodeToString(sha256.Sum256([]byte(text))[:])` using an addressable sum variable, and assert selector-only changes do not affect it.
- [x] Run `env GOWORK=off GOCACHE=/private/tmp/spl-toolkit-m8-gocache go test -mod=readonly ./pkg/analysis -run 'TestRequirementTypesJSONShape|TestQueryDigestExactBytes' -count=1`. Record the missing-type/helper compile failure.
- [x] Define the exact public structs and `func queryDigest(text string) string`; add only enough code to pass those two tests.
- [x] Add `TestCapabilityRevisionUsesNormalizedTypedManifest`. Compare empty selectors with `spl/splunkd/current`, compare repeated calls, compare SPL with SPL2, and independently marshal `CapabilitiesFor(CapabilityOptions{...})` to calculate the expected digest.
- [x] Implement `func capabilityRevision(document QueryDocument) (string, error)` without maps, indentation, trailing newline, environment data, or cached mutable state.
- [x] Add `TestCloneRequirementSetOwnsNestedSlices`. Mutate cloned coverage reasons, item occurrences, gap reference IDs, gap diagnostic codes, and diagnostics, then assert the source is unchanged.
- [x] Add `TestRequirementSetEmptyCollectionsAreArrays`. Marshal an empty initialized set and assert `coverage.reasons`, `items`, `gaps`, and `diagnostics` are `[]`, not `null`.
- [x] Implement `func cloneRequirementSet(in RequirementSet) RequirementSet` and add exactly `CodeRequirementIndeterminate`, `CodeRequirementDynamic`, and `CodeRequirementCoverageIncomplete` with their approved wire strings in `pkg/analysis/diagnostics.go`.
- [x] Run:

       env GOWORK=off GOCACHE=/private/tmp/spl-toolkit-m8-gocache go test -mod=readonly ./pkg/analysis -run 'TestRequirementTypes|TestQueryDigest|TestCapabilityRevision|TestCloneRequirementSet' -count=1
       env GOWORK=off GOCACHE=/private/tmp/spl-toolkit-m8-gocache go test -mod=readonly ./pkg/analysis -count=1
       git diff --check

   Expected evidence is all tests passing and no whitespace errors.
- [x] Update the administrative plan sections with Task 1 evidence, then commit the owned files plus administrative updates with `git commit -m "feat(analysis): define requirement product model"`.

## Task 2: Single-pass query-only trace

**Owned files:** `pkg/analysis/requirements_trace.go`, `pkg/analysis/requirements_trace_test.go`, and minimal trace plumbing in the existing `pkg/analysis/analyze.go`, `pkg/analysis/flow.go`, `pkg/analysis/references.go`, `pkg/analysis/transfers.go`, `pkg/analysis/source_fields.go`, `pkg/analysis/dependencies.go`, `pkg/analysis/scopes.go`, and `pkg/analysis/spl2_lower.go` owners.

- [x] Add `TestRequirementTraceClassifiesFieldOrigins` for `search src=* | eval derived=src | table src derived`. Assert direct source reads are trace obligations, the derived consumer is not, and create/output references are not consuming obligations. Run it and record the missing-trace failure.
- [x] Add table rows to that test for rename source versus target, removals, `isnull`, unavailable local fields, wildcard reads, and an indeterminate source. Each row asserts the pending reference ID, query-only binding, direct/conditional flags, and event order.
- [x] Implement `newRequirementTrace`, `requirementEnvironment.clone`, and trace initialization beside the existing `environment`. Clone both environments at the same branch and scope sites.
- [x] Route `referenceAt`, `readAt`, create, projection, rename, aggregation, removal, and source operations through `recordReference` at the same semantic event that creates the public reference. Do not walk syntax again.
- [x] Add `TestRequirementTraceKnowledgeObjects` for `index=main source=access.log sourcetype=web`, `| inputlookup users`, `| datamodel Web`, and a macro. Assert exact direct objects, wildcard or dynamic states, and unresolved macro-expansion evidence.
- [x] Record parser and ordinary semantic diagnostics through `recordDiagnostic`, including the exact owning pending reference IDs. Add a separate refinement-only diagnostic helper that never records into the query trace.
- [x] Add `TestRequirementTraceDiagnosticOwnership` with a wildcard, unknown command, unsupported semantics, syntax error, and macro expansion. Assert explicit ownership and event order without comparing locations.
- [x] Extend the existing `finalizeReferences` mapping handoff to call `requirementTrace.remapReferences(mapping)`. Add `TestRequirementTraceRemapsEachPendingIDOnce` and assert every trace reference matches its public identity, kind, role, stage, scope, and location after remapping.
- [x] Add `TestRequirementTraceRefinementParity` rows for a finite field list, partial source universe, wildcard selectors, dotted SPL2 names, and resolved versus unresolved public bindings. Compare the serialized private trace from plain and refined analysis while asserting that at least one public result differs.
- [x] Inspect `rg -n 'Analyze\(|parseDocument|parseSPL2|analyzeRewrite|transfer' pkg/analysis/requirements_trace.go pkg/analysis/requirements.go` and confirm trace code contains no recursive analysis, parser, lowerer, transfer, or projector call.
- [x] Run:

       env GOWORK=off GOCACHE=/private/tmp/spl-toolkit-m8-gocache go test -mod=readonly ./pkg/analysis -run '^TestRequirementTrace' -count=1
       env GOWORK=off GOCACHE=/private/tmp/spl-toolkit-m8-gocache go test -mod=readonly ./pkg/analysis -count=1
       git diff --check

   Expected evidence is byte-equivalent query traces across refinement pairs, one traversal per case, and a green analysis package.
- [x] Update the administrative plan sections with Task 2 evidence, then commit with `git commit -m "feat(analysis): capture query requirement evidence"`.

## Task 3: Projection, embedding, standalone Go API, and analysis corpus

**Owned files:** `pkg/analysis/model.go`, `pkg/analysis/analyze.go`, `pkg/analysis/requirements.go`, `pkg/analysis/requirements_test.go`, `pkg/analysis/corpus_test.go`, `pkg/analysis/transfers_test.go`, `testdata/requirements/cases.json`, and `testdata/analysis/cases.json`.

- [ ] Add `TestProjectRequirementsGroupsAndOrdersOccurrences`. Supply a finalized trace with two source reads sharing `(field, host, read, exact)` and one different role. Assert two ordered items, canonical occurrence order, and IDs `req-1` and `req-2`. Run it and record the missing-projector failure.
- [ ] Implement `projectRequirements(document QueryDocument, trace *requirementTrace) (RequirementSet, error)` through the empty-set and grouping stages only. Initialize all arrays and omit full query text.
- [ ] Add `TestProjectRequirementsFieldPolicy` with subtests for exact source, indeterminate source, wildcard, dynamic, derived, create, output, rename source and target, remove, null test, and unavailable local references. Assert the exact inclusion, necessity, binding, resolution, and gap outcome.
- [ ] Add `TestProjectRequirementsKnowledgePolicy` covering `index`, `source`, `sourcetype`, `dataset`, `data_model`, `lookup`, and `macro`, including exact, defensible wildcard/dynamic, gap-only dynamic, and exact macro plus expansion gap.
- [ ] Add `TestProjectRequirementsGapOrderingAndDeduplication`. Supply duplicate diagnostic ownership evidence and assert deduplication by code plus ordered links, deterministic message selection, gap order, and first-seen unique `coverage.reasons`.
- [ ] Add `TestProjectRequirementsStatusIndependentFromCoverage` for `search host=* | fields - host | table host`. Assert query status `invalid`, coverage complete, no external item for the unavailable local read, and no requirement gap for the local defect.
- [ ] Add `Requirements RequirementSet` to `Result` and call the projector once after canonical references and query-only diagnostics are finalized. Add a focused empty-query test proving every successful `Analyze` path emits a non-null requirement object with array-valued collections.
- [ ] Implement `Requirements(document)` with exactly one `Analyze(document)` call, unchanged error return, local deep clone, and pointer to the clone. Add `TestRequirementsMatchesEmbeddedAndIsDetached` for every requirement fixture.
- [ ] Create `testdata/requirements/cases.json` with full expected sets for valid, invalid, and incomplete SPL and SPL2; exact and derived fields; indeterminate and wildcard fields; all direct knowledge kinds; dynamic identities; macros; removals; null tests; rename source and target; branch and scope evidence; dotted SPL2; non-ASCII text; and empty output.
- [ ] Update `testdata/analysis/cases.json` with the additive runtime field. Review a mechanical before/after sample for valid, invalid, and incomplete cases; do not modify historical evidence or receipts.
- [ ] Update the hand-pinned full-result JSON strings in `pkg/analysis/transfers_test.go` variable `transferParityWitnesses`. Preserve every pre-Milestone-8 byte except insertion of the correctly derived `requirements` member, and retain the comments that forbid runtime regeneration inside the test.
- [ ] Run:

       env GOWORK=off GOCACHE=/private/tmp/spl-toolkit-m8-gocache go test -mod=readonly ./pkg/analysis -run 'TestProjectRequirements|TestRequirements|TestAnalysisCorpus|TestLookupExtractionFullReportParity|TestLocatedTransferEquivalence' -count=1
       env GOWORK=off GOCACHE=/private/tmp/spl-toolkit-m8-gocache go test -mod=readonly ./pkg/analysis -count=1
       git diff --check

   Expected evidence is exact fixture parity, a single analysis invocation, deterministic repeat runs, and a green analysis corpus.
- [ ] Update the administrative plan sections with Task 3 evidence, then commit with `git commit -m "feat(analysis): project canonical query requirements"`.

## Task 4: Refinement and downstream product propagation

**Owned files:** `pkg/analysis/requirements_refinement_test.go`, `pkg/analysis/source_fields_test.go`, focused existing `pkg/analysis/rewrite_*_test.go` files only when their assertions embed full analysis results, `pkg/document/model.go`, `pkg/document/snapshot.go`, `pkg/document/snapshot_test.go`, focused `pkg/validation/*_test.go` files, `pkg/rewrite/rewrite_test.go`, `pkg/rewrite/requirements_test.go`, `pkg/corpus/requirements_test.go`, `pkg/graph/export_test.go`, `pkg/sarif/export_test.go`, `pkg/impact/compare_test.go`, `internal/lsp/server_test.go`, `testdata/validation/cases.json`, and `testdata/schemas/cases.json`. Changes to production validation, rewrite, corpus, graph, SARIF, impact, or LSP logic require a demonstrated propagation bug and a `Decision Log` entry.

- [ ] Add `TestRequirementSetFieldRefinementParity` for plain analysis, `AnalyzeWithSourceFields`, and `AnalyzeWithSourceUniverse`. Use exact, partial, wildcard, unresolved wildcard, and dotted SPL2 cases. Marshal only the embedded sets and assert byte equality.
- [ ] Add `TestRequirementSetValidationRefinementParity` in `pkg/validation` with a field-list target, JSON Schema target, and OCSF target. Compare each report's embedded requirement set with `analysis.Requirements` for the same normalized `QueryDocument`.
- [ ] Add `TestRequirementSetRewriteParity` in `pkg/rewrite/requirements_test.go`. Compare original and candidate analysis requirements with plain analysis of their respective normalized query texts in preview and apply modes.
- [ ] Update `pkg/rewrite/rewrite_test.go` helper `explicitAliasReport` so both hand-built `analysis.Result` values contain exact hand-derived requirement sets. Preserve its independent full-report witness role and do not call `Analyze` from the helper.
- [ ] Add `TestSnapshotRequirementSetDetached`. Call `document.New(result, RevisionContext{ToolVersion: "test", ContractVersion: "v1"})`, mutate every nested requirement slice on the result and snapshot in turn, and assert no alias in either direction.
- [ ] Add `Requirements analysis.RequirementSet` to `document.Snapshot` and a package-local deep clone in `document.New`. Do not serialize through JSON to clone it.
- [ ] Add `TestCorpusEvaluationCarriesRequirements` for the existing `corpus.Evaluation.Analysis` pointer. Assert the corpus aggregate remains unchanged and no requirement aggregate is introduced.
- [ ] Add or extend exclusion tests proving no requirement-specific graph nodes or edges, SARIF rules, impact comparison fields, or LSP messages appear.
- [ ] Update only the current full-result witnesses in `testdata/validation/cases.json` and `testdata/schemas/cases.json`. Inspect `pkg/rewrite/rewrite_test.go:explicitAliasReport` and all `rg -n '\*analysis\.Result|analysis\.Result{' pkg --glob '*_test.go'` hits for equivalent hand-built complete results; update only witnesses whose equality contract covers the complete runtime result.
- [ ] Run:

       env GOWORK=off GOCACHE=/private/tmp/spl-toolkit-m8-gocache go test -mod=readonly ./pkg/analysis -run 'Test.*Requirement.*Refinement|Test.*Requirement.*Rewrite' -count=1
       env GOWORK=off GOCACHE=/private/tmp/spl-toolkit-m8-gocache go test -mod=readonly ./pkg/document ./pkg/validation ./pkg/rewrite ./pkg/corpus -count=1
       env GOWORK=off GOCACHE=/private/tmp/spl-toolkit-m8-gocache go test -mod=readonly ./pkg/... -count=1
       git diff --check

   Expected evidence is byte-for-byte plain/refined parity and detached propagation through every existing result-bearing product.
- [ ] Update the administrative plan sections with Task 4 evidence, then commit with `git commit -m "feat(products): propagate canonical requirements"`.

## Task 5: CLI and REST adapters

**Owned files:** `cmd/analysis.go`, `cmd/analysis_test.go`, `cmd/cli.go`, `cmd/cli_test.go`, `cmd/main.go`, `pkg/api/analysis.go`, `pkg/api/analysis_test.go`, `pkg/api/server.go`, and `pkg/api/server_test.go`.

- [ ] Add `TestRequirementsCLIInputBoundary` with `requirements 'search host=web'`, `requirements --query 'search host=web'`, duplicate input, `--file`, `--stdin`, and `--batch`. Assert the first two succeed and unsupported acquisition modes exit `2` without a report.
- [ ] Add `TestRequirementsCLIJSONMatchesGo` using one valid SPL and one incomplete SPL2 fixture with selectors and `--source-id`. Decode stdout and compare the complete value with `analysis.Requirements`.
- [ ] Add `TestRequirementsCLIExitCodes` with exact rows: valid plus complete `0`, invalid `1`, incomplete query `3`, valid query plus incomplete requirement coverage `3`, invalid option `2`, output write failure `2`, and an unknown `RequirementSet.QueryStatus` passed directly to the exit helper `2`.
- [ ] Add `func requirementsExitCode(set *analysis.RequirementSet) int` and `func formatRequirementsText(set *analysis.RequirementSet) []byte` in `cmd/analysis.go`. Text must print query status and requirement coverage separately, followed by items, gaps, and diagnostics in canonical order.
- [ ] Register `requirements` in the custom command switch and help text. Reuse existing `parseCLIOptions`, validation, JSON writing, and `--output` behavior; do not add a third-party parser dependency.
- [ ] Add `TestRequirementsRESTContentStatuses` against `NewServer` for valid, invalid, and incomplete SPL and SPL2. Assert HTTP 200 and full equality with `analysis.Requirements`.
- [ ] Add `TestRequirementsRESTRejectsRequestBoundaryErrors` for duplicate and unknown members, empty and malformed JSON, invalid Unicode, wrong or missing content type, trailing JSON, unsupported selectors, and bodies over 1 MiB. Assert HTTP 400 and no partial report.
- [ ] Implement `func (s *Server) handleRequirementsQuery(w http.ResponseWriter, r *http.Request)` by calling existing `parseAnalysisDocument`, then `analysis.Requirements`, then `writeJSONResponse`. Register only `POST /api/v1/query/requirements` through the existing middleware chain.
- [ ] Add a server test proving GET is rejected and the handler has no file or network input path. Inspect the handler diff to confirm all classification remains in `pkg/analysis`.
- [ ] Run:

       env GOWORK=off GOCACHE=/private/tmp/spl-toolkit-m8-gocache go test -mod=readonly ./cmd -run 'Test.*Requirements' -count=1
       env GOWORK=off GOCACHE=/private/tmp/spl-toolkit-m8-gocache go test -mod=readonly ./pkg/api -run 'Test.*Requirements' -count=1
       env GOWORK=off GOCACHE=/private/tmp/spl-toolkit-m8-gocache go test -mod=readonly ./cmd ./pkg/api -count=1
       git diff --check

   Expected evidence is exact JSON parity with the Go API, stable text, exact process exits, strict transport behavior, and no adapter-owned analysis logic.
- [ ] Update the administrative plan sections with Task 5 evidence, then commit with `git commit -m "feat(adapters): expose requirement CLI and REST APIs"`.

## Task 6: Native C ABI and Python API

**Owned files:** `pkg/bindings/bindings.go`, `pkg/bindings/requirements_test.go`, `python/spl_toolkit/mapper.py`, `python/spl_toolkit/libspl_toolkit.h`, `python/tests/test_native_analysis.py`, `python/tests/test_native_requirements.py`, `python/tests/test_native_abi.py`, and `tests/native/memory.c`.

- [ ] Add `TestRequirementsExportReturnsOwnedJSON` in `pkg/bindings/requirements_test.go`. Create a handle, call the missing export with `{"text":"search host=web"}`, decode the result, compare with `analysis.Requirements`, and free it. Record the compile failure before implementation.
- [ ] Add `TestRequirementsExportRejectsBadDocumentsAndClosedHandles` for malformed, array, duplicate-member, unknown-member, invalid UTF-8, and closed-handle inputs. Assert an owned error and nil result, then call `spl_result_free` exactly once.
- [ ] Implement `func spl_mapper_requirements_query(mapperID C.int, documentJSON *C.char) *C.SPLResult` using `ownedMapperJSONResult`, the same strict one-document decoder as `spl_mapper_analyze_query`, and `analysis.Requirements`.
- [ ] Add `test_requirements_matches_canonical_fixture` in `python/tests/test_native_requirements.py` and register the native function as `[ctypes.c_int, ctypes.c_char_p] -> ctypes.POINTER(SPLResult)`. Run the test and record the missing Python method failure.
- [ ] Implement the exact `SPLMapper.requirements_query` signature and delegate through `_validate_fields_request(..., operation="requirements")`. Do not add a Python classifier or fallback.
- [ ] Add Python tests for defaults, all selectors, non-ASCII and NUL text, invalid selectors, lone surrogates, native error mapping, and complete parity with embedded analysis requirements.
- [ ] Add lifecycle tests that monkeypatch JSON decoding and the native call. Assert `spl_result_free` and `_operation` cleanup on success, native content error, JSON decode error, and Python exception.
- [ ] Add repeated, 16-worker concurrent, close-waits-for-admitted-call, invalid-handle, and closed-handle tests modeled on `python/tests/test_native_analysis.py`.
- [ ] Extend `tests/native/memory.c` with successful and failing requirement calls inside its existing loop, and free every non-null `SPLResult` exactly once. Extend exported-symbol expectations in `python/tests/test_native_abi.py`.
- [ ] Run `make build-shared`, copy the generated `build/libspl_toolkit.h` to `python/spl_toolkit/libspl_toolkit.h`, rerun `make build-shared`, and require `cmp build/libspl_toolkit.h python/spl_toolkit/libspl_toolkit.h` to succeed.
- [ ] Run:

       env GOWORK=off GOCACHE=/private/tmp/spl-toolkit-m8-gocache go test -mod=readonly ./pkg/bindings -count=1
       make build-shared
       python3 -m pytest -q python/tests/test_native_analysis.py python/tests/test_native_requirements.py python/tests/test_native_abi.py
       cmp build/libspl_toolkit.h python/spl_toolkit/libspl_toolkit.h
       git diff --check

   Expected evidence is ABI parity, exactly-once frees, concurrent-call safety, and an idempotently generated header. On Linux with GCC 13, also run the exact `native-memory` commands from `.github/workflows/ci.yml`: build the shared library with Go `-asan`, compile `tests/native/memory.c` with `-fsanitize=address`, and run with `ASAN_OPTIONS=detect_leaks=1:halt_on_error=1`. On other hosts, record the local limitation. The hosted exact-SHA `native-memory` job is a later branch-CI gate, not a prerequisite for local independent acceptance.
- [ ] Update the administrative plan sections with Task 6 evidence, then commit with `git commit -m "feat(native): expose canonical query requirements"`.

## Task 7: JSON Schema, OpenAPI, registries, and source manifests

**Owned files:** `contracts/v1/requirements.schema.json`, `contracts/v1/shared.schema.json`, `contracts/README.md`, `tests/acceptance/test_machine_contracts.py`, `testdata/tooling/contracts.json`, `tools/update_validation_openapi.py`, `tools/tests/test_update_validation_openapi.py`, generated `docs/docs.go`, `docs/swagger.json`, `docs/swagger.yaml`, `python/native-source-files.txt`, `tools/release-content-files.txt`, `tools/release-source-files.txt`, and `tools/tests/test_release.py`.

- [ ] Add a failing `requirements` family entry to `tests/acceptance/test_machine_contracts.py` and a minimal valid report in `testdata/tooling/contracts.json`. Run the family test and record the missing schema failure.
- [ ] Add `analysis.RequirementQueryIdentity`, `analysis.RequirementCoverage`, `analysis.RequirementOccurrence`, `analysis.RequirementItem`, `analysis.RequirementGap`, and `analysis.RequirementSet` definitions to `contracts/v1/shared.schema.json`. Require every defined field, preserve `additionalProperties: true`, and constrain status, necessity, origin, resolution, kind, IDs, and digest formats.
- [ ] Add `contracts/v1/requirements.schema.json` as a root reference to the shared `analysis.RequirementSet` definition.
- [ ] Add `requirements` to the property maps for shared `analysis.Result` and `document.Snapshot`, but do not add it to either existing v1 `required` array.
- [ ] Add negative fixtures only for missing required requirement-set fields, wrong primitive or collection types, invalid enum values, malformed `req-N` and `ref-N` IDs, and malformed digests. Do not expect an unknown output property to fail.
- [ ] Add an explicit positive additive-property witness with an unknown top-level `RequirementSet` member and assert it validates. Add archived analysis and snapshot witnesses that omit `requirements` and assert both still validate.
- [ ] Register local `$ref` closure and update `contracts/README.md`. Run with network disabled by the existing harness and assert no remote fetch.
- [ ] Add the REST operation's Swag response annotation and only the minimum reconciler changes needed by `tools/update_validation_openapi.py`. Extend `tools/tests/test_update_validation_openapi.py` first.
- [ ] Run `make generate-docs` once and inspect OpenAPI path `/query/requirements` under the existing `/api/v1` server base, `analysis.RequirementSet`, the optional analysis and snapshot properties, enums, arrays, and digests in all three generated files.
- [ ] Snapshot that accepted first generation, rerun the generator, and compare the second generation with the snapshot:

       mkdir -p /private/tmp/spl-toolkit-m8-openapi-first
       cp docs/docs.go docs/swagger.json docs/swagger.yaml /private/tmp/spl-toolkit-m8-openapi-first/
       make generate-docs
       cmp /private/tmp/spl-toolkit-m8-openapi-first/docs.go docs/docs.go
       cmp /private/tmp/spl-toolkit-m8-openapi-first/swagger.json docs/swagger.json
       cmp /private/tmp/spl-toolkit-m8-openapi-first/swagger.yaml docs/swagger.yaml

- [ ] Add `contracts/v1/requirements.schema.json`, `pkg/analysis/requirements.go`, and `pkg/analysis/requirements_trace.go` to `python/native-source-files.txt` and the release manifests according to their existing inclusion policy. Add no test or `_build_plan/` entry.
- [ ] Extend `tools/tests/test_release.py` manifest expectations and verify the built wheel receives the new contract through the existing `native-source-files.txt` contract-copy path.
- [ ] Run:

       python3 -m pytest -q tests/acceptance/test_machine_contracts.py
       python3 -m pytest -q tools/tests/test_update_validation_openapi.py tools/tests/test_release.py
       git diff --check

   Expected evidence is offline schema validation, archived-v1 compatibility, deterministic OpenAPI, complete manifest test coverage, and no `_build_plan` path in any runtime or distribution manifest.
- [ ] Update the administrative plan sections with Task 7 evidence, then commit with `git commit -m "feat(contracts): publish requirement interface contracts"`.

## Task 8: Permanent docs, packaged acceptance, and release closure

**Owned files:** `README.md`, `docs/API.md`, `docs/cli.md`, `docs/architecture.md`, `docs/compatibility.md`, `python/README.md`, `tests/acceptance/test_requirements_surfaces.py`, `tests/acceptance/cli_examples.json`, `python/tests/test_native_requirements.py`, `tools/check_package.py`, `tools/check_acceptance.py`, `tools/tests/test_package.py`, `tools/tests/test_acceptance.py`, the fixture-copy and hash plumbing for `testdata/requirements/cases.json`, and the always-allowed administrative files described above.

- [ ] Add `test_requirements_fixture_matches_go_cli_http_and_python` in `tests/acceptance/test_requirements_surfaces.py`. Drive valid, invalid, and incomplete SPL and SPL2 from `testdata/requirements/cases.json`; compare complete decoded values, allowing only JSON object-key order to differ.
- [ ] Add real loopback-server rows for valid content, invalid content, incomplete content, malformed requests, and the 1 MiB boundary. Compare the complete response with the Go fixture.
- [ ] Add CLI rows for help, positional and `--query`, text, JSON, `--output`, rejected file/stdin/batch modes, and exits `0`, `1`, `2`, and `3` in `tests/acceptance/cli_examples.json`.
- [ ] Update existing current CLI example outputs that serialize a complete `analysis.Result`. Preserve all unrelated bytes and historical evidence.
- [ ] Add failing `tools/tests/test_package.py` cases asserting the requirement fixture, new native test, and new acceptance test are copied to the out-of-checkout tooling root, named in evidence, and hashed.
- [ ] Extend `tools/check_package.py` constants and copy logic. Set `SPL_REQUIREMENTS_FIXTURES` to an absolute copied path and run requirement tests against both installed wheel and rebuilt sdist without repository imports.
- [ ] Add failing `tools/tests/test_acceptance.py` cases for the new required test files and evidence keys. Extend `tools/check_acceptance.py` without lowering counts, allowing skips, or weakening existing hashes.
- [ ] Update `README.md`, `docs/API.md`, `docs/cli.md`, `docs/architecture.md`, `docs/compatibility.md`, and `python/README.md` with the full report shape, Go and Python examples, CLI and REST examples, digest rules, grouping, source versus derived policy, independent statuses, exit behavior, ABI ownership, refinement parity, additive v1 compatibility, offline boundaries, and exclusions.
- [ ] State in permanent docs that digests identify supplied data and are not authentication, authorization, signatures, or execution permission. Add no link or runtime dependency on `_build_plan/`.
- [ ] Complete `_build_plan/milestones/8-product-interface-requirements/milestone-log.md` and the living sections in this plan with implementation decisions, task SHAs, local test evidence gathered so far, deviations, and known local environment limits. Do not attempt to record future final-review, hosted-CI, merge, or post-merge results.
- [ ] Run focused acceptance and documentation checks:

       python3 -m pytest -q tests/acceptance/test_requirements_surfaces.py
       python3 -m pytest -q tools/tests/test_package.py tools/tests/test_acceptance.py
       python3 tools/check_docs.py

- [ ] Check for newly added `_build_plan` dependencies without treating normal `rg` no-match status as a failure and without hiding a `git diff` or `rg` execution error:

       set -e
       git diff --unified=0 origin/main -- README.md docs cmd internal pkg python tools tests contracts go.mod go.sum > /private/tmp/spl-toolkit-m8-runtime.diff
       if rg '^\+.*_build_plan' /private/tmp/spl-toolkit-m8-runtime.diff; then exit 1; else m8_rg_status=$?; test "$m8_rg_status" -eq 1; fi

   Expected evidence is no newly added `_build_plan` reference in runtime, test input, package, contract, or permanent-documentation files. Existing historical evidence and negative distribution assertions may retain their current references.
- [ ] Run the repository-level local gates:

       make fmt
       git diff --check
       python3 tools/check_go.py
       env GOWORK=off GOCACHE=/private/tmp/spl-toolkit-m8-race-gocache go test -mod=readonly -race ./...
       make python-test
       python3 -m pytest -q tools/tests/test_package.py tools/tests/test_acceptance.py tools/tests/test_release.py
       python3 tools/check_docs.py

   Expected evidence is every local gate passing. The Linux GCC 13 AddressSanitizer and full multi-target evidence-aggregation gates remain mandatory exact-SHA CI checks. Record environment-limited failures precisely and do not treat CI as a silent substitute. Any deterministic source, fixture, or generated-file diff after a second run is a defect.
- [ ] Rerun generated-header and generated-OpenAPI comparisons, confirm the administrative files are final for pre-freeze execution, and commit with `git commit -m "test(acceptance): prove requirement product surfaces"`.

## Final independent review and revision gate

After Task 8 is accepted and all tracked administrative updates are committed, record `HEAD` as the provisional candidate SHA. Run four independent subagent reviews over `origin/main...HEAD`: full design-spec compliance, code quality and robustness, security and native-memory safety, and architecture and repository organization. Each reviewer must inspect source, tests, contracts, generated artifacts, package manifests, permanent docs, and the final administrative delta, and must return findings with exact file and line evidence. Reviewers do not edit.

Run the available external reviewer from the feature worktree:

    claude --model opus --print "Review git diff origin/main...HEAD in this repository for Milestone 8 requirement-product specification compliance, correctness, robustness, security, native memory ownership, deterministic behavior, architecture, test adequacy, packaging, and documentation. Do not modify files. Return actionable findings with file and line evidence, followed by residual risks."

Capture its output and exit status in the orchestrator's out-of-band handoff, not in a tracked file. Do not waive a finding because it came from an external model. Validate it against the approved specification and source. For every confirmed finding, assign a fresh fix agent, require a focused regression test, create a scoped commit, bring the plan and milestone log up to date, then rerun all affected task reviews and all four final reviews. Repeat the Claude command after material fixes. Completion requires no unresolved actionable findings; report disputed findings and their concrete disposition out of band.

When all reviewers approve, declare the reviewed `HEAD` the frozen candidate SHA. From that point through local acceptance, branch CI, merge, and post-merge verification, do not edit or commit the plan, milestone log, or any other tracked file. Report final review conclusions, acceptance results, hosted URLs, merge results, and post-merge results out of band. If a later gate requires a fix, explicitly unfreeze the candidate, make and review the fix, finish all administrative updates, rerun this entire final review gate, and freeze the new reviewed `HEAD`.

Reviewers must explicitly confirm:

1. Query requirements are produced from one canonical traversal and are byte-equivalent across refinement modes and all adapters.
2. Classification, grouping, ordering, gaps, diagnostic ownership, status, coverage, digests, and deep-copy behavior meet every rule in this plan.
3. REST, CLI, C, and Python error and ownership boundaries are exact and do not introduce filesystem, network, injection, unsafe allocation, or use-after-free behavior.
4. Archived v1 analysis and snapshot payloads remain valid, while current runtime payloads always emit requirements.
5. `_build_plan/` is absent from runtime imports, test inputs, package inputs, installed artifacts, release manifests, generated permanent docs, and product documentation links. Tests may retain explicit negative assertions that distributions do not contain it.
6. Historical receipts and evidence were not edited, generated artifacts are reproducible, and the source and release manifests contain exactly the necessary new files.

## Independent acceptance validation

Assign a fresh acceptance agent that did not implement or review the feature. Give it the frozen candidate SHA, this plan's user-visible behavior, and the approved design spec. It must begin from the clean feature worktree, run the shared requirement fixture across Go, CLI text and JSON, a real HTTP server, native C, Python, installed wheel, and installed sdist, and independently inspect the machine contracts and permanent docs. This is a local pre-push gate. It must finish and sign off without waiting for GitHub Actions.

The acceptance agent runs at minimum:

    git status --short
    git diff --check origin/main...HEAD
    env GOWORK=off GOCACHE=/private/tmp/spl-toolkit-m8-accept-gocache go test -mod=readonly -race ./...
    make python-test
    python3 -m pytest -q tests/acceptance/test_requirements_surfaces.py tests/acceptance/test_machine_contracts.py
    python3 -m pytest -q tools/tests/test_package.py tools/tests/test_acceptance.py tools/tests/test_release.py
    python3 tools/check_docs.py

It must also manually sample at least one valid, one invalid, and one incomplete SPL and SPL2 case and report response JSON, CLI exit, HTTP status, native/Python parity, and requirement coverage. On a non-Linux host, it records the local GCC 13 AddressSanitizer limitation without failing local signoff; the hosted `native-memory` job remains a later branch-CI gate. Acceptance fails on any unexplained skip other than that known platform limitation, dirty generated artifact, checkout-relative installed-package read, or unverified required surface. Report the signoff out of band and do not edit tracked files. A failure unfreezes the candidate and restarts the fix, task review, administrative update, final whole-feature review, freeze, and local acceptance sequence.

## Branch CI, fast-forward merge, and post-merge verification

Only after all local, review, Claude, and acceptance gates pass:

- [ ] Confirm the feature worktree is clean and bind the frozen SHA:

       git status --short
       git diff --check origin/main...HEAD
       m8_candidate_sha="$(git rev-parse HEAD)"
       git rev-parse HEAD > /private/tmp/spl-toolkit-m8-candidate-sha.txt

   Expected evidence is empty status and diff-check output. Do not make another tracked edit or evidence commit after this point.
- [ ] Push the feature branch without force:

       git push -u origin codex/milestone-8-product-interface-requirements
       test "$(git ls-remote origin refs/heads/codex/milestone-8-product-interface-requirements | cut -f1)" = "$(cat /private/tmp/spl-toolkit-m8-candidate-sha.txt)"

- [ ] Dispatch CI because ordinary feature pushes do not trigger the workflow. Query by the frozen commit so an older run cannot be mistaken for evidence:

       m8_candidate_sha="$(cat /private/tmp/spl-toolkit-m8-candidate-sha.txt)"
       gh workflow run .github/workflows/ci.yml --ref codex/milestone-8-product-interface-requirements
       m8_run_id="$(gh run list --workflow ci.yml --branch codex/milestone-8-product-interface-requirements --commit "$m8_candidate_sha" --event workflow_dispatch --limit 1 --json databaseId --jq '.[0].databaseId')"
       test -n "$m8_run_id"
       gh run watch "$m8_run_id" --exit-status
       gh run view "$m8_run_id" --json url,headSha,conclusion,event,workflowName

   If the first `gh run list` is empty because dispatch registration is delayed, repeat only that list command until it returns the run; do not dispatch a duplicate. Verify `headSha` equals `$m8_candidate_sha`, `event` is `workflow_dispatch`, and `conclusion` is `success`. A successful run at another SHA is not evidence. This hosted gate is later than local acceptance and does not retroactively change its signoff.
- [ ] If branch CI fails, report the run out of band, unfreeze the candidate, fix it on the branch, update administrative records before review, rerun affected task reviews, final whole-feature reviews, local independent acceptance, push, and dispatch a new exact-SHA run. Never edit tracked files merely to record a hosted result.
- [ ] Inspect the primary checkout's tracked state and branch without treating intentional untracked roadmap inputs as an error:

       git -C /Users/jmdelgad/repos/spl-toolkit diff --quiet
       git -C /Users/jmdelgad/repos/spl-toolkit diff --cached --quiet
       test "$(git -C /Users/jmdelgad/repos/spl-toolkit rev-parse --abbrev-ref HEAD)" = "main"
       git -C /Users/jmdelgad/repos/spl-toolkit ls-files --others --exclude-standard > /private/tmp/spl-toolkit-m8-primary-untracked-before.txt
       LC_ALL=C sort -o /private/tmp/spl-toolkit-m8-primary-untracked-before.txt /private/tmp/spl-toolkit-m8-primary-untracked-before.txt

       python3 -c 'from pathlib import Path; actual=set(Path("/private/tmp/spl-toolkit-m8-primary-untracked-before.txt").read_text().splitlines()); expected={"CLAUDE.md","_build_plan/prd.md","_build_plan/prd.html","_build_plan/milestones/8-product-interface-requirements/prompt.md","_build_plan/milestones/9-language-capability-ledger/prompt.md","_build_plan/milestones/10-broad-spl-semantics/prompt.md","_build_plan/milestones/11-broad-spl2-semantics/prompt.md","_build_plan/milestones/12-knowledge-object-closure/prompt.md","_build_plan/milestones/13-environment-snapshot-contracts/prompt.md","_build_plan/milestones/14-live-splunk-exporter/prompt.md","_build_plan/milestones/15-compatibility-assessment/prompt.md","_build_plan/milestones/16-safe-resolution-fanout/prompt.md","_build_plan/milestones/17-detection-ci-ai-evidence/prompt.md"}; assert actual == expected, f"untracked baseline mismatch: missing={sorted(expected-actual)!r} extra={sorted(actual-expected)!r}"'

   The branch must be `main`, both tracked diff commands must exit `0`, and the saved untracked list must contain exactly `CLAUDE.md`, `_build_plan/prd.md`, `_build_plan/prd.html`, and one `prompt.md` in each existing Milestone 8 through Milestone 17 prompt directory:

       _build_plan/milestones/8-product-interface-requirements/prompt.md
       _build_plan/milestones/9-language-capability-ledger/prompt.md
       _build_plan/milestones/10-broad-spl-semantics/prompt.md
       _build_plan/milestones/11-broad-spl2-semantics/prompt.md
       _build_plan/milestones/12-knowledge-object-closure/prompt.md
       _build_plan/milestones/13-environment-snapshot-contracts/prompt.md
       _build_plan/milestones/14-live-splunk-exporter/prompt.md
       _build_plan/milestones/15-compatibility-assessment/prompt.md
       _build_plan/milestones/16-safe-resolution-fanout/prompt.md
       _build_plan/milestones/17-detection-ci-ai-evidence/prompt.md

   Stop and coordinate if any other untracked file exists or any listed file is absent. In particular, do not restore the legacy untracked Markdown or text files that the user intentionally removed.
- [ ] Fetch and prove fast-forward ancestry:

       git -C /Users/jmdelgad/repos/spl-toolkit fetch origin
       git -C /Users/jmdelgad/repos/spl-toolkit merge-base --is-ancestor origin/main codex/milestone-8-product-interface-requirements

   Stop if `origin/main` is not an ancestor. Do not stash, reset, rebase, restore, or overwrite user files.
- [ ] Compare the incoming tracked tree against the exact untracked baseline and require no path collision:

       git -C /Users/jmdelgad/repos/spl-toolkit ls-tree -r --name-only codex/milestone-8-product-interface-requirements > /private/tmp/spl-toolkit-m8-incoming-tracked.txt
       LC_ALL=C sort -o /private/tmp/spl-toolkit-m8-incoming-tracked.txt /private/tmp/spl-toolkit-m8-incoming-tracked.txt
       comm -12 /private/tmp/spl-toolkit-m8-primary-untracked-before.txt /private/tmp/spl-toolkit-m8-incoming-tracked.txt > /private/tmp/spl-toolkit-m8-untracked-collisions.txt
       test ! -s /private/tmp/spl-toolkit-m8-untracked-collisions.txt

   Both Git path lists are lexically ordered. If a collision appears, stop before merge and coordinate with the user; do not move or delete the untracked file.
- [ ] Fast-forward and push `main`:

       m8_candidate_sha="$(cat /private/tmp/spl-toolkit-m8-candidate-sha.txt)"
       git -C /Users/jmdelgad/repos/spl-toolkit merge --ff-only codex/milestone-8-product-interface-requirements
       test "$(git -C /Users/jmdelgad/repos/spl-toolkit rev-parse HEAD)" = "$m8_candidate_sha"
       git -C /Users/jmdelgad/repos/spl-toolkit push origin main

   The merge SHA must equal `$m8_candidate_sha`. Do not include, remove, restore, or modify any untracked roadmap input.
- [ ] Immediately verify tracked cleanliness and exact preservation of the untracked baseline:

       git -C /Users/jmdelgad/repos/spl-toolkit diff --quiet
       git -C /Users/jmdelgad/repos/spl-toolkit diff --cached --quiet
       git -C /Users/jmdelgad/repos/spl-toolkit ls-files --others --exclude-standard > /private/tmp/spl-toolkit-m8-primary-untracked-after.txt
       LC_ALL=C sort -o /private/tmp/spl-toolkit-m8-primary-untracked-after.txt /private/tmp/spl-toolkit-m8-primary-untracked-after.txt
       cmp /private/tmp/spl-toolkit-m8-primary-untracked-before.txt /private/tmp/spl-toolkit-m8-primary-untracked-after.txt

- [ ] Watch the automatically triggered `main` push workflow at that exact SHA:

       m8_candidate_sha="$(cat /private/tmp/spl-toolkit-m8-candidate-sha.txt)"
       m8_main_run_id="$(gh run list --workflow ci.yml --branch main --commit "$m8_candidate_sha" --event push --limit 1 --json databaseId --jq '.[0].databaseId')"
       test -n "$m8_main_run_id"
       gh run watch "$m8_main_run_id" --exit-status
       gh run view "$m8_main_run_id" --json url,headSha,conclusion,event,workflowName
       git -C /Users/jmdelgad/repos/spl-toolkit fetch origin
       git -C /Users/jmdelgad/repos/spl-toolkit rev-parse HEAD origin/main

   If the first list is empty, repeat only the list command. Verify the run `headSha` and both local and remote `main` SHAs equal `$m8_candidate_sha`, the event is `push`, and the conclusion is `success`. Report branch and main run URLs, conclusions, merge SHA, and preserved-untracked evidence out of band. Do not edit a tracked file to record them.

## Acceptance criteria

Milestone 8 is complete only when all of the following are evidenced at the landed SHA:

1. Plain and every supported refinement path produce byte-identical `RequirementSet` values from a single parse, lower, and semantic traversal.
2. Field and knowledge-object classification, ordering, grouping, gaps, diagnostics, digests, status, and coverage match the shared fixture for valid, invalid, and incomplete SPL and SPL2.
3. Go, CLI, REST, C, Python, document snapshot, validation, and schema products expose identical canonical requirement content with exact boundary semantics and detached ownership.
4. The v1 requirement schema requires all defined members and exact types while preserving the existing additive `additionalProperties: true` policy; current analysis and snapshot schemas expose the property without invalidating archived payloads, OpenAPI is reproducible, and offline contract validation passes.
5. Wheel and sdist checks run outside the checkout, include and hash the requirement fixture, exercise the real native library, and do not depend on `_build_plan/`.
6. Permanent docs, source manifests, content manifests, and release manifests are complete; historical evidence is unchanged; local gates, independent reviews, Claude Opus review, independent acceptance, feature-branch CI, and post-merge `main` CI all pass at the exact accepted SHA.

## Idempotence and recovery

All generators, fixture updates, formatters, and manifest checks must be safe to rerun and produce no second-run diff. Use fixed cache directories under `/private/tmp` only for build cache, never for source worktrees. Test and package tools must create their own isolated output directories and clean them through their existing mechanisms.

If a focused test fails, retain the failing command and first relevant error in `Surprises & Discoveries`, fix the smallest owning layer, and rerun the focused command before broadening. If generated output drifts, regenerate from its authoritative source, never hand-edit a generated artifact. If a native or installed-package check fails after a partial build, rerun the repository's build target rather than copying ad hoc libraries into the package.

If an implementation agent leaves uncommitted changes, the orchestrator inspects and assigns ownership before continuing. Do not reset or discard them. If a review or CI fix changes behavior, unfreeze the candidate, reopen both specification and quality review for the complete task range, finish administrative updates, rerun final review, and rerun local acceptance. If branch CI fails after push, add new commits and push normally; never force-push. If `main` advances before landing, stop and re-evaluate ancestry rather than rebasing or merging silently. If the primary checkout has tracked/index changes, an unexpected untracked file, a missing roadmap input, or an incoming collision, stop and ask the user to resolve it. Do not treat the expected untracked roadmap baseline as dirty, and do not restore the legacy Markdown or text files the user removed.

The analysis fixture, requirement fixture, generated OpenAPI, generated header, contract registry, native source manifest, and release manifests are authoritative checkpoints. Compare them before and after recovery to distinguish a stale build from a semantic change.

## Plan revision note

Created on 2026-09-14 from the approved Milestone 8 design specification and the repository state at `8029a44`. Revised after independent review to add checkbox-sized TDD actions and named helpers, include hand-pinned full-result witnesses, preserve additive output contracts, make OpenAPI and no-dependency checks executable, separate local acceptance from hosted CI, eliminate evidence SHA self-reference, and preserve the primary checkout's exact untracked roadmap baseline during landing. Update this note before final whole-feature review whenever execution changes a public shape, task boundary, verification command, or landing procedure, and explain why in `Decision Log`. After freeze, report terminal outcomes out of band rather than editing this file.
