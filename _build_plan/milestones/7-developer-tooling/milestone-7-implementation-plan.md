# Developer Tooling and Ecosystem Implementation Plan — DRAFT


> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development only after root accepts the reconciled final plan and releases production. Use one implementer at a time, then independent specification/code review of its immutable owned-file commit, fixes and scoped re-review before the next task. This draft is not released for execution.

**Goal:** Let users scan local query repositories, consume honest aggregate/SARIF/graph results, compare schema and mapping changes, embed a detached document view, and receive actual local editor diagnostics/highlights from the canonical engine.

**Architecture:** The Go analysis/validation/rewrite packages own semantics. New corpus acquisition, in-memory orchestration, export and impact components preserve those reports and their uncertainty; a local language server translates editor lifecycle and source positions without parsing SPL itself. Existing CLI, native Python and optional HTTP adapters expose the same versioned values.

**Tech Stack:** Go 1.22 floor, existing ANTLR runtime/frontends, Python 3.11+ real native binding, offline JSON Schema/OCSF inputs, SARIF 2.1.0 plus approved Errata 01, LSP 3.17 baseline and the existing release tooling. Consumer preparation uses official vscode-languageclient 9.0.1/protocol 3.17.5 and sarif-tools 3.0.5 in an isolated temporary directory; these are acceptance tools, not runtime product dependencies.

**Spec:** `_build_plan/milestones/7-developer-tooling/design-spec.md`, approved in full by root. This plan restates the behavior an executor needs. Read the design alongside the final plan.

This is a living ExecPlan maintained under `/Users/jacobdelgado/.codex/PLANS.md`. Checkboxes belong only in Progress; task steps use numbered actions. The named file, serial SDD workflow and draft-only authorization override the skill's default plan location and execution-choice question. This is deliberately not a final novice-executable release: the bounded reconciliation register must be replaced with accepted M6 interfaces and concrete platform/dependency recipes before final approval.


## Purpose / Big Picture


A user currently has canonical single/batch query operations. After M7, `scan` can read a repository of dedicated `.spl`/`.spl2` files, retain useful findings when another file cannot be read, and state exactly what was and was not checked. CI can consume SARIF, embedding tools can use graph/document data, and migration projects can compare explicit before/after schemas or mapping previews without changing source files. A real local editor displays the same findings and proven field relationships through a standard language client.

Success is observable: a corpus containing one valid query, one definite missing-field error and one unknown command retains all three reports with invalid aggregate status and incomplete coverage; an unreadable fourth file also causes operational exit 2. A schema change that alters one field outcome appears in impact output even if both sides remain invalid. The editor receives exact diagnostics and highlights in its own storage/provider APIs, rather than merely passing a handcrafted wire harness.


## Global Constraints


The Go language floor remains 1.22 and Python remains 3.11+. Preserve actual accepted release platforms and predecessor APIs. Do not use newer `os.Root` APIs or silently require a newer Go dependency. Do not change toolchain pins, package version, supported platforms or remote release workflows just to pass local checks.

The toolkit remains offline by default; no query execution, automatic schema retrieval, runtime credentials, hosted state, autocomplete, graphical IDE, bulk migration or arbitrary host-language extraction is introduced. Files contain one whole UTF-8 query. Explicit manifests carry literal text or a dedicated query-file path; neither syntax implies host JSON/dashboard/code coordinates.

Canonical syntax, semantic, schema and rewrite coverage stay independent. Invalid wins over incomplete, and valid requires complete error-free relevant static analysis. Operational failure is distinct from content findings. Graph/impact/editor code must not invent parser, binding, lineage, schema, predicate or rewrite semantics. No generic AST or generated parser object crosses a public boundary.

Source hashes cover exact original UTF-8 bytes. Analysis revision identity also covers normalized language/profile/version and tool/contract context. Validation revision identity covers prepared target content/options/resources, not merely caller labels. Stage/reference IDs remain revision-local. Source spans provide evidence, not persistent identity across edits.

Preserve all other users/agents' changes in the shared worktree. During this draft, do not implement, test, build, generate fixtures, dispatch SDD workers, launch GUI/editor clients, download dependencies or commit. Future execution also grants no push, merge, tag, upload, publication, destructive cleanup or automatic migration writes. `_build_plan/` is temporary guidance and never a runtime/test/package dependency.


## Progress


- [x] (2026-09-08 03:07 UTC) Read writing-plans, PLANS.md, approved M7 design and immutable reviewed M4 core `5e3d4ec3a54b6878830b83a8ff08405011a507d9` without running predecessor checks.
- [x] (2026-09-08 UTC) Read approved M5/M6 draft proposals as unimplemented contracts and rechecked existence of pinned consumer-preparation artifacts.
- [x] (2026-09-08 UTC) Root approved narrow OS-specific anchored opens and minimal canonical target-preparation access reconciliation.
- [x] (2026-09-08 UTC) Completed draft self-review for spec coverage, interface consistency, ownership and bounded pending facts.
- [ ] Root reviews this draft; no production release follows automatically.
- [ ] Obtain full accepted M6 handoff and reconcile all bounded pending interfaces/dependencies below.
- [ ] Root approves the final plan and releases serial production.
- [ ] Task 1: strict corpus/manifest models and revision identity.
- [ ] Task 2: contained filesystem acquisition and deterministic selection.
- [ ] Task 3: canonical prepared target access and in-memory corpus reports.
- [ ] Task 4: detached advanced document access.
- [ ] Task 5: evidence-preserving graph export.
- [ ] Task 6: standards-correct SARIF export.
- [ ] Task 7: before/after schema and mapping impact.
- [ ] Task 8: published machine schemas and independent contract validation.
- [ ] Root completes a stable Go-only core acceptance window before any adapters.
- [ ] Task 9: local LSP lifecycle and source conversion.
- [ ] Task 10: CLI/HTTP operations and executable examples.
- [ ] Task 11: native Python and installed package closure.
- [ ] Task 12: actual editor/SARIF consumers and release evidence closure.
- [ ] Root completes a stable final built/native/real-consumer window; broad review and independent exact-head acceptance close M7.


## Surprises & Discoveries


At reviewed M4 core 5e3d4ec, schema target preparation is private: `preparedSchemaTarget`, `prepareSchemaTarget` and `validatePreparedSchema` live in pkg/validation. Public ValidateSchemaBatch prepares once but rejects an empty document slice before preparation. A corpus with all file acquisitions failing must still reject an invalid shared target before acquisition, so an artificial query or empty batch is not an acceptable preparation trick. Root approved reconciling the smallest canonical prepare/validate access against accepted M6; no second validator is permitted.

The reviewed core's `sourceIndex` in pkg/analysis/source.go handles UTF-8 bytes, Unicode columns, LF, CRLF and lone CR but is private and uses ANTLR rune indices. LSP requires zero-based UTF-16 positions. A source-coordinate adapter can inspect bytes without interpreting SPL, but must not claim it can call private source helpers or subtract one from a Unicode column.

M5 proposes CapabilitiesFor(CapabilityOptions), per-form limitations and optional Lineage.Phase/ExecutionOrder; M6 proposes public Request/RuleSet/Result plus original/candidate audit correspondence. Neither proposal is implemented evidence for M7. In particular, M5 private locatedOperand/readAt helpers are not public tooling APIs, and M4 NormalizedName alone cannot distinguish an atom containing a dot from a navigated path.

The existing consumer directory is `/private/tmp/spl-toolkit-m7-consumer-preflight`. VS Code 1.117.0 arm64 CLI works but emitted `SecCodeCheckValidity` / `NSOSStatusErrorDomain Code=-2147409622`; no GUI/client run occurred. sarif-tools 3.0.5 CLI and dependency checks passed during preparation. Its flat CSV retains rule/message/severity/encoded URI/startLine, but does not prove base-ID resolution, decoded navigation or Unicode columns. Those require independent semantic tests. No preparation result is product acceptance.


## Decision Log


2026-09-07, root: Approved dedicated files plus strict manifest, fixed ignore policy, explicit source identity, acquisition-error preservation, deterministic summaries, revision-qualified graphs, static before/after impact and detached canonical document view. Approved LSP full-sync/UTF-16/diagnostics/highlights with real-client acceptance; no extra AST or editor extension distribution.

2026-09-07, root: Approved written spec and required target/refinement identity to include content/options/resources even when caller identity/version labels are unchanged. Prepared consumer dependencies were accepted as future inputs, not product verification.

2026-09-08, root: Authorized only this bounded draft overlap. Full accepted M6, final interface reconciliation and root release still precede production. Reserve a stable Go-only window before adapters and a stable final built/native/real-consumer window.

2026-09-08, root: Preserve Go1.22/no-symlink containment using narrow OS-specific anchored opens on supported platforms. Test component swaps deterministically; Lstat followed by Open is insufficient. Prefer accepted dependencies; a minimal compatible addition is a final preflight selection. Unsupported operating systems fail explicitly rather than use weaker traversal.

2026-09-08, root: Approve minimal canonical target preparation access before filesystem acquisition, including all-files-fail cases, without a fabricated Analyze query or duplicated schema logic. Exact signatures/ownership wait for accepted M6.

2026-09-08, root: Invalid initialization target/options fail initialization. Invalid configuration changes visibly error, clear diagnostics and pause affected publication while buffers/lifecycle continue; corrected configuration resumes latest snapshots. Config-dependent requests fail while paused. Explicit valid target removal selects analysis mode. New corpus/impact HTTP routes use 8 MiB while old limits remain unchanged.


## Outcomes & Retrospective


This revision prepares a draft task sequence and explicit acceptance contracts only. No M7 implementation, fixture, build, editor launch or product result exists from drafting. Approved behavior is retained; the pending register concerns actual predecessor access, final dependency/platform mechanics and package/transport closure. A future executor must not interpret the draft's named new types as already-existing APIs.


## Context and Orientation


Work only in `/Users/jacobdelgado/repos/spl-toolkit/.worktrees/remaining-milestones`, branch `codex/remaining-milestones`. Current adapter/native files are active earlier-milestone work and cannot be used as final dependency authority. The immutable reviewed M4 core is `5e3d4ec3a54b6878830b83a8ff08405011a507d9`, subject `feat(validation): expose schema reports and batch APIs`. Root supplies the later accepted commit; inspect it through immutable reads before changing production.

At this core, `pkg/analysis/model.go` defines QueryDocument with Text, Language, Profile, Version and SourceID. Analyze in analyze.go returns `*analysis.Result`, containing SchemaVersion, Document, Status, Coverage, Stages, Scopes, References, Lineage, Dependencies and Diagnostics. Locations are half-open UTF-8 byte offsets with one-based Unicode line/column evidence. References expose ID, kind, role, scope/stage, resolution, binding and origin IDs. Binding means source-required, created upstream, proven unavailable or indeterminate; tooling consumes that fact instead of recomputing it.

`pkg/validation/model.go` defines FieldCatalog, Report, BatchReport and Coverage. A field Report embeds Analysis, target, outcomes and combined diagnostics. `schema_model.go` defines SchemaTarget, SchemaReport and SchemaBatchReport; schema outcomes retain evidence, partial matches and class information. Real reviewed functions are:

    analysis.Analyze(document analysis.QueryDocument) (*analysis.Result, error)
    validation.DecodeFieldCatalog(data []byte) (validation.FieldCatalog, error)
    validation.Validate(document analysis.QueryDocument, catalog validation.FieldCatalog) (*validation.Report, error)
    validation.ValidateBatch(documents []analysis.QueryDocument, catalog validation.FieldCatalog) (*validation.BatchReport, error)
    validation.DecodeSchemaTarget(data []byte) (validation.SchemaTarget, error)
    validation.ValidateSchema(document analysis.QueryDocument, target validation.SchemaTarget) (*validation.SchemaReport, error)
    validation.ValidateSchemaBatch(documents []analysis.QueryDocument, target validation.SchemaTarget) (*validation.SchemaBatchReport, error)

SchemaTarget is a tagged union: json_schema carries raw root schema, base URI and supplied resource map; ocsf carries raw compiled catalog and exact selection. Decode and preparation reject malformed input; incomplete field projection is a report. `internal/jsoninput/unicode.go` supplies strict Unicode checks; validation/request.go supplies private duplicate-key/presence/unknown-key handling. Share a narrowly factored strict decoder only if accepted M6 already did so or the same behavior needs reuse; never copy permissive encoding/json unmarshalling and lose duplicate/surrogate rejection.

`cmd/cli.go` dispatches CLI operations; `cmd/validation_cli.go` contains existing file/stdin/batch conventions. `pkg/api/server.go` registers HTTP routes and `pkg/api/validation.go` wraps field validation at this core. `pkg/bindings/bindings.go` owns guarded C handles and allocated result JSON; `python/spl_toolkit/mapper.py` frees results in finally. `python/native-source-files.txt`, `tools/check_package.py`, tools/release.py, tools/check_acceptance.py and existing tests/acceptance own explicit native/package/evidence closure. Preserve accepted M4–M6 additions rather than copying older files over them.

M5's approved draft proposes a dedicated SPL2 frontend under parser/spl2 and shared lowering in pkg/analysis. SQL source stages remain lexical, with logical lineage Phase and ExecutionOrder; SELECT final projection can be a later phase of the same stage. M6's approved draft proposes pkg/rewrite.Rewrite/RewriteBatch and strict RuleSet/Request, retaining original/candidate text, analyses, validation and group/rule/reference audit. Empty rules are proposed valid no-ops. These proposal details guide reconciliation but do not establish final symbol names or safety proof access.


## Bounded pending reconciliation register


First, root must identify full accepted M6 including accepted M5, with immutable SHA, source/package versions and actual final reports/capability vocabulary. Record actual names/tags for SQL phase/evaluation order, source IDs, typed identity distinctions, copied ownership and rewrite original-to-candidate provenance. Do not expose M5 private frontend helpers or M6 internal predicate state to tooling. An absent proven alignment produces indeterminate impact, not guessed matching.

Second, choose the smallest canonical prepared field/schema target handle and shared target selector against accepted M6, plus any existing independent rule/preview preparation access needed to prepare mapping comparisons before acquisition. Record preparation, normalization and per-document/batch evaluation signatures plus input-error classification. Preparation must run once per target before acquisition and be reusable without leaking internal projection/transfer state. The wrapper must permit a prepared target with zero acquired documents while keeping existing public empty-batch rejection unchanged. If accepted M6 already supplies this access, reuse it. Otherwise Task 3 owns a narrow addition inside pkg/validation, with regression proof and no semantic changes.

Third, finalize the Go1.22-compatible anchored-open dependency/OS API calls in Task 2 before production release. Record supported Darwin/Linux/Windows release builds, concrete no-follow/no-reparse semantics, root-handle lifetime, relative directory enumeration and deterministic race-test hooks. Select only existing compatible dependencies or one minimal audited addition; no download occurs while drafting. Unsupported OS builds return explicit unsupported-platform errors. Real OS runtime evidence, cross-compilation and local mocks are different proof categories.

Fourth, select the full JSON Schema instance validator and local schema registry for Task 8. Use Python jsonschema's validator_for/check_schema/iter_errors APIs or the accepted equivalent, pin all package versions and hashes compatible with Python3.11, and verify supported meta-schema dialects for both M7 Draft2020-12 contracts and the actual official SARIF schema. Do not use M4 field projection as a wire validator. Also select a maintained Go1.22-compatible LSP JSON-RPC transport if an accepted dependency exists; otherwise use a narrowly tested stdio framing implementation without adding a semantic SDK. Record exact APIs before Task 9.

Fifth, reconcile actual M6 CLI/HTTP flags/routes, body-limit helpers, C exports/header generation, Python guards/methods, OpenAPI update tool, native manifests, mandatory installed suites/fixtures, package-data paths and release recipes. New corpus/impact HTTP routes must carry both ordinary before/after OCSF catalogs plus bounded documents: use the root-approved 8 MiB (8,388,608-byte) per-route policy, retain explicit too-large failure, and preserve legacy 1 MiB routes. Reconcile only actual accepted helper names and boundary status behavior; no global limit increase or silent truncation.

Sixth, verify temporary consumer dependencies still exist and match recorded locks/hashes. Record the actual VS Code executable/version, candidate artifact path, isolated launch recipe and accepted native/Python3.14/package recipes. Do not treat this planning host's Go1.25.5/Xcode26.2 as pinned Go1.26.8/Xcode16.4 release acceptance. Root controls build-output and independent-input windows; do not read or execute its private oracles. Replace these six access/dependency entries with actual facts, revise affected signatures/tasks consistently, and obtain root final-plan release before dispatching implementation.


## Interfaces and file responsibilities


The following are concrete proposed new M7 interfaces, not claims about predecessor code. Final reconciliation may rename a shared target alias or immutable predecessor report type, but must keep every consumer consistent. Create pkg/corpus for strict normalized input, identity and canonical report orchestration; pkg/corpusio for local manifest/discovery with internal/corpusfs for platform opens; pkg/document for detached access; pkg/graph and pkg/sarif for pure exports; pkg/impact for static comparison; internal/lsp for the process adapter. Do not create a plugin framework or one interface per report field.

In pkg/corpus/model.go, define Input with Entries []Entry and Selection metadata. Entry contains ID, Origin and exactly one of Document *analysis.QueryDocument or Failure *AcquisitionError. Origin records kind, relative path and optional actual base URI; SourceID stays inside the canonical document as opaque metadata. Selection records mode, selection-complete flag, fixed ignored names, skipped symlinks and traversal failures. AcquisitionError has stable code, phase and message plus selected path/ID as applicable, without a forged query range. ScanOptions contains optional ValidationTarget. ValidationTarget is the single accepted M6 field_list/json_schema/ocsf selector or a thin shared alias, not a new schema format.

Define Evaluation with Kind and exactly one of Analysis *analysis.Result, FieldValidation *validation.Report or SchemaValidation *validation.SchemaReport. A Report has SchemaVersion, Status, ExecutionComplete, Selection, Counts, Coverage, Dependencies and ordered Entries. Each ReportEntry retains ID, Origin, SourceHash, AnalysisRevision and exactly one Evaluation or Failure. Counts distinguish selected/analyzed/acquisition_failed/traversal_failed; coverage counters explicitly carry complete/incomplete/denominator and schema not_requested. Empty collections always encode as arrays.

Define an opaque PreparedScan handle in pkg/corpus that owns normalized options and canonical prepared targets. Proposed calls are:

    func DecodeRequest(data []byte) (Request, error)
    func Prepare(options ScanOptions) (*PreparedScan, error)
    func (p *PreparedScan) Scan(input Input) (*Report, error)
    func Scan(request Request) (*Report, error)
    func IsInputError(err error) bool

Request is strict schema_version 1, nonempty documents [{id, document}] and optional validation_target. In-memory requests cannot ask for file loading; the loader produces Input separately. Prepare before calling the loader. Request errors return nil/error; a valid selection with acquisition failures returns a report with ExecutionComplete false. Internal engine failures return an error and cannot be converted into successful entries. Successful neighboring acquisition results are retained when failures are local I/O; earlier atomic batch APIs stay unchanged.

In pkg/corpusio, define Manifest, ManifestEntry and decoding/loading calls:

    func DecodeManifest(data []byte) (Manifest, error)
    func LoadManifest(manifest Manifest, manifestPath string) (corpus.Input, error)
    func LoadDirectory(root string) (corpus.Input, error)

ManifestEntry uses presence-aware Text *string versus Path *string and explicit selectors/source_id; normalize into Entry only after strict structure/path checks. A manifest's absent base defaults to `.` relative to its parent. Paths are slash-separated relative names within the declared base; reject drive/UNC/backslash/absolute/parent traversal inputs before opening. File-level symlinks/nonregular/unreadable/non-UTF8 are acquisition failures after valid selection, while malformed manifest/options and a symlink root are upfront errors.

In pkg/document, create a detached Snapshot containing copied canonical Result and revision context. RevisionContext contains ToolVersion, ContractVersion and TargetDigest strings; the normalized document supplies language/profile/version and exact source hash. These identify the supplied snapshot, not a cryptographic proof that an arbitrary caller-provided report came from the engine. Proposed APIs New(result *analysis.Result, context RevisionContext) (*Snapshot, error), Reference(id string) (analysis.Reference, bool), Stage(id string) (analysis.Stage, bool), Scope(id string) (analysis.Scope, bool), ReferencesIn(startByte, endByte int) ([]analysis.Reference, error), ReferencesAt(byteOffset int) ([]analysis.Reference, error), and Source(startByte, endByte int) (string, error). These inspect the current caller-owned copy, never stale cached indices; mutation voids engine attestation. Ranges are half-open, bounds/rune boundaries are checked, overlapping matches return all in canonical order, zero-width matches require exact-position lookup and EOF does not select a preceding reference by accident.

In pkg/graph expose Export(report *corpus.Report) (*Report, error). Graph Report is versioned nodes/edges plus source/revision/context/coverage metadata. Node/edge IDs encode an unambiguous tuple of document/revision/entity identity. References/stages use canonical IDs; transitions add owner, occurrence and actual phase identity. Edges retain report pointers and conditional/incomplete evidence. Reject malformed externally supplied report references rather than invent an endpoint.

In pkg/sarif expose Export(report *corpus.Report) (*Log, error). Log has SARIF version/schema/runs and typed required SARIF structures, without adding schema_version to the standardized top level. Deterministic driver rules/artifacts/results map original findings. Use explicit unicodeCodePoints, correct encoded root-relative or virtual URIs and invocation status; do not encode byte offsets as charOffset.

In pkg/impact expose strict DecodeSchemaRequest(data []byte) (SchemaRequest, error), DecodeMappingRequest(data []byte) (MappingRequest, error), CompareSchemas(request SchemaRequest) (*Report, error) and CompareMappings(request MappingRequest) (*Report, error). Decoded requests contain normalized corpus.Input and explicit before/after targets or accepted rewrite rule sets with optional before/after validation targets. Wire requests accept inline documents, not acquisition-error claims. For local acquisition, separate preparation from comparison: PrepareSchemas(before, after corpus.ValidationTarget) (*PreparedSchemaComparison, error) and PrepareMappings(before, after MappingSide) (*PreparedMappingComparison, error), each with Compare(input corpus.Input) (*Report, error). MappingSide contains the accepted RuleSet and optional canonical validation target. These handles own canonical preparation only, not new matching/evaluation logic. Convenience CompareSchemas/CompareMappings prepare and compare already-supplied input; CLI prepares, loads once, then calls the handle so acquisition failures remain in the result. Final preflight selects actual M6 preparation access rather than reparsing rules or inventing a second validator. Mapping calls preview only. Each ImpactEntry retains complete before/after canonical evaluation or rewrite results, finding/outcome/dependency deltas, explicit alignment evidence and affected/unchanged/indeterminate/failed classification. Root must reconcile exact accepted rewrite target/audit types before these compile-time fields are frozen.

In internal/lsp expose Serve(ctx context.Context, in io.Reader, out io.Writer, log io.Writer, options Options) error. Options selects supported profile/version and optional canonical local target input; the supplied analysis/corpus preparation boundary is reused, not a second matcher. Lifecycle state keys URI/open-generation/document-version/configuration-revision. Stdout writes are serialized framed JSON-RPC; logs go only to the provided stderr writer. Only full synchronization, diagnostics and document highlights are advertised.


## Task 1: Strict corpus and manifest contracts with revision identity


Own pkg/corpus/model.go, request.go, identity.go and focused tests; pkg/corpusio/manifest.go and manifest_test.go; durable testdata/tooling/requests.json. These files establish the proposed request/report types and decoder behavior; they do not fake successful analysis. Consume actual QueryDocument and canonical target decoders selected in preflight. Produce the contracts used by all later tasks.

1. Write TestCorpusRequestStrict, TestManifestExactlyOneSource and TestRevisionIncludesTargetContent with concrete inputs: duplicate id, duplicate JSON key, text/path both present, text/path both absent, null versus empty text, escaped unpaired surrogate, unsupported language, absolute/drive/UNC/parent paths, unchanged schema labels with changed schema content, and differently normalized selector context. A valid inline empty text remains a document request, not a decoding error.
2. Run `go test -mod=readonly ./pkg/corpus ./pkg/corpusio -run 'TestCorpusRequest|TestManifest|TestRevision'`; observe missing definitions/failing requirements. Implement only strict tagged/presence decoding, canonical selector/target delegation and collision-safe identity. Use encoding/json with duplicate-key/presence checks and jsoninput Unicode validation, not round-tripping lossy numbers or cleaning source text.
3. Define source hash as SHA256([]byte(Text)). Hash revision tuples using canonical JSON arrays with sorted resource-map keys and exact raw target content or accepted normalized target serialization; caller labels never replace content. If two logically equivalent differently formatted schemas produce different content revisions, that is permitted conservative identity, not a semantic difference claim.
4. Run the focused suite and existing request-normalization regressions. Commit only owned files and obtain immutable specification/code review before Task 2. No implementation fixture may read `_build_plan/`.


## Task 2: Contained file acquisition and deterministic discovery


Own internal/corpusfs/open_unix.go, open_windows.go, open_unsupported.go and focused platform tests, plus pkg/corpusio/load.go, discover.go, load_test.go and discovery_test.go. Exact platform dependencies/syscalls are fixed by reconciliation before this task. Consume strict Manifest and corpus.Input; produce immutable loaded bytes or explicit acquisition failures without language analysis.

1. Write TestDirectorySelectionOrder, TestManifestContainment, TestExplicitSymlinkFailure and TestComponentSwapDoesNotEscape. Build temporary roots containing `.spl`/`.spl2`, unrelated files, ignored directories, an empty query, invalid UTF-8, nonregular selected files, same path under distinct IDs and an external sentinel. Inject a barrier between resolving a parent component and opening the child; swap it for a symlink/reparse point, then require either an original-root handle read or explicit failure, never sentinel bytes. Use test hooks local to the filesystem module instead of timing-sensitive sleep loops.
2. Observe failure with `go test -mod=readonly ./internal/corpusfs ./pkg/corpusio`. Implement root-handle acquisition and component-by-component relative opens with no-follow/no-reparse checks. Enumerate/read through anchored handles; inspect opened object type, not only pre-open pathname metadata. Reject unsupported platforms explicitly. Do not weaken the contract to EvalSymlinks/Lstat followed by ordinary Open.
3. Implement exact lowercase extension selection, fixed ignored directory names `.git`, `.hg`, `.svn`, `.worktrees`, `node_modules`, `vendor`, `.venv`, `venv`, `build`, `dist`, no .gitignore parser, slash-normalized lexical ordering and explicit manifest order. No symlink traversal during discovery; record skipped symlinks and traversal failures. Explicit selected symlinks fail acquisition; ignored subtrees are not counted as checked. Inline manifests use virtual coordinates, not escaped-JSON locations.
4. Preserve bytes including BOM/newlines, validate UTF-8, close every handle after snapshot, and continue independent read failures. Reject a complete empty selection; if traversal failed, report that operational failure rather than an empty-success explanation. Test permission failures with controllable injected I/O errors when root/host permissions cannot reproduce them honestly.
5. Run focused real local tests and Go1.22 build checks. Record Darwin runtime evidence separately from Windows/Linux compile-only or unexecuted runtime gates. Commit/review; root does not accept a platform safety claim solely from a cross-compile.


## Task 3: Canonical target preparation and in-memory corpus assessment


Own pkg/corpus/prepare.go, scan.go, aggregate.go and tests plus the exact narrowly approved pkg/validation preparation-access files from reconciliation. Consume canonical analysis, prepared field/schema validation and normalized Input. Produce complete corpus.Report values; no export/UI logic belongs here.

1. Write TestSharedTargetRejectedBeforeAcquisition, TestAllAcquisitionsFail, TestCorpusMixedStatusCoverage and TestCorpusSourceParity. A closed catalog `["host"]` with `search host=web`, `search missing=x` and `| mystery | table host` must yield analyzed counts 3, invalid overall, incomplete semantic coverage and retained missing/unsupported findings. Add a failed fourth entry: selected 4, analyzed 3, acquisition_failed 1, ExecutionComplete false. An invalid target must fail Prepare before the injected loader is called, including all-failed-source cases.
2. Run `go test -mod=readonly ./pkg/corpus ./pkg/validation -run 'TestSharedTarget|TestAllAcquisitions|TestCorpus|TestSchemaBatch'`. Add the minimal canonical prepare handle if absent, preserving existing empty-batch errors and one-target-per-batch preparation. Its public surface does not expose project/universe/finalize internals. Prepared targets must own copied input and be safe for the actual supported concurrent use.
3. Implement corpus preparation and sequential ordered evaluation first. Use canonical reports directly. One acquisition failure cannot erase another document; a canonical internal error is still an operation error. Do not catch panics as empty reports. Default analysis-only schema coverage to not_requested. Reduce invalid before incomplete before valid, retaining operational-failure state separately. Generate dependencies only from grammar-established canonical references.
4. Test same target identity/version with changed content, caller mutation, repeated/concurrent equivalence, independent scopes and partial schema wildcards. Compare embedded reports to direct public operations by full JSON value, normalizing only object-key order. A content error can still have complete coverage; no averaging completeness into a misleading percentage.
5. Run focused core tests and immutable review. Preserve predecessor validation/analysis wire shapes and batch atomicity. The root Go-only window occurs after all later core export/impact/schema tasks, before adapter Task 9.


## Task 4: Detached advanced document access


Own pkg/document/model.go, snapshot.go, query.go and snapshot_test.go. Consume canonical Result/revision context only. Produce typed source/ID/range lookup helpers without parser internals or a new AST.

1. Write TestSnapshotDetachedMutation, TestOverlappingReferenceLookup, TestSourceRangeBoundaries and TestSQLPhasePreservation. Mutate a returned reference slice and prove the original analysis and a second snapshot remain unchanged; mutate the owned snapshot and prove lookup sees that copy rather than stale cache. Unknown ID returns found=false; invalid/rune-splitting byte ranges error; adjacent half-open spans and EOF do not overselect. Include non-BMP text and overlapping roles on one source range.
2. Observe red tests with `go test -mod=readonly ./pkg/document`. Implement deep copying of every nested slice/pointer, direct or rebuilt lookup against the supplied copy, checked source slicing and explicit revision provenance. A caller-modified copy is no longer an engine-verified report. Do not copy ANTLR contexts/tokens into an exported type.
3. Retain exact canonical origin links and SQL phase/evaluation-order fields from accepted M6. Helpers expose facts, not transitive inferred equivalence. Run focused tests, inspect exported API dependency types for parser leakage, commit/review.


## Task 5: Exportable dependency and lineage graph


Own pkg/graph/model.go, export.go, identity.go, export_test.go and durable testdata/tooling/graph-cases.json. Consume corpus.Report and canonical document evidence. Produce deterministic graph nodes/edges whose report pointers and endpoints are checkable.

1. Write TestGraphEndpointIntegrity, TestGraphRevisionIsolation, TestGraphConditionalLineage and TestGraphSQLPhaseOrder. Use same-spelled fields in separate scopes, repeated transition owners, invalid/incomplete neighbors and the same bytes with different dialect/target content. Require all IDs unique and endpoints/report pointers resolvable; names or offsets alone cannot serve as ID.
2. Run `go test -mod=readonly ./pkg/graph`. Implement stable tuple-encoded IDs, exact dependency grouping by kind/name and per-reference edges carrying original document/revision/evidence pointers. Preserve uncertainty, conditionality and incomplete wildcard membership. A graph node named host is not proof of a live catalog resource.
3. Iterate source lexical stage arrays and explicit lineage evaluation evidence separately. A SELECT-owned project phase does not create a synthetic source stage. Reject malformed external evidence references rather than creating missing nodes. Run deterministic JSON and source-slice tests, commit/review.


## Task 6: Standards-correct SARIF export


Own pkg/sarif/model.go, export.go, locations.go, export_test.go and durable testdata/tooling/sarif-cases.json. Task 8 owns the official vendored schema and full validator; this task owns semantic export tests. Consume corpus.Report; produce SARIF2.1.0+Errata01 Log.

1. Write TestSARIFRuleAndArtifactIndices, TestSARIFURIsAndUnicode, TestSARIFInvocationVsContent and TestSARIFIncompleteDiagnostics. Inputs include spaces/#/%/non-ASCII file paths, Windows-style root metadata supplied as URI, original Unicode/non-BMP/CRLF source, EOF/zero-width diagnostics, virtual inline sources, no findings and acquisition failure beside an invalid query.
2. Observe red with `go test -mod=readonly ./pkg/sarif`. Implement deterministic rules/results/artifacts, code→ruleId, severity→level, messages and unicodeCodePoints regions. Construct URIs with standard net/url path escaping per segment; do not concatenate raw Windows paths or double-encode percent escapes. Use a trailing-slash absolute file-URI root mapping and relative artifact URIs for files; virtual artifacts contain exact query text and do not claim host-manifest coordinates.
3. Set invocation success false for acquisition/tool failures; content exits 1/3 alone can still describe successful invocation. Export every incomplete diagnostic and independent coverage properties. Unlocated incompleteness stays unlocated. Omit charOffset/charLength rather than insert UTF-8 byte units. Verify artifact/rule indices and region bounds against original snapshots with independent assertions.
4. Run focused tests and compare repeated output. Commit/review; official schema validation in Task 8 and actual consumer conversion in Task 12 remain separate required gates.


## Task 7: Static schema and mapping change impact


Own pkg/impact/model.go, request.go, schema.go, mapping.go, alignment.go, compare_test.go and durable testdata/tooling/impact-cases.json. Consume corpus preparation, accepted canonical validation and M6 preview/audit APIs. Produce before/after canonical evidence with explicit deltas and honest alignment.

1. Write TestSchemaSameStatusChangedOutcome, TestMappingIdentityBaseline, TestMappingAlignmentAmbiguous and TestImpactNeverWrites. Compare a source requiring host and missing against before catalog `["host"]` and after empty catalog: both sides invalid, but host gains a missing outcome/diagnostic and must appear affected. Use one incomplete query with equal partial reports and require indeterminate, not unchanged.
2. Run `go test -mod=readonly ./pkg/impact`. Prepare both targets/rule sets before acquiring snapshots; hash complete target/options/resource contents. Reuse one loaded corpus for both sides. Validate unchanged source against each schema; invoke M6 in preview only for each mapping side, with explicit empty-rule no-op baseline and optional selected validation. Retain full original/candidate audits and reports without publishing migration edits.
3. Align exact original-reference identities under the same source revision and verify kind/role/scope/range. If refinement changes final IDs, allow only a unique exact original-evidence match; duplicates remain ambiguous. Use accepted M6 audit provenance for candidate correspondence. Do not match by nearest offset, same spelling or array position. Compare diagnostic/outcome severity/message/evidence changes, not status or error count alone.
4. Sort deltas deterministically, preserve introduced/resolved/changed/unmatched/ambiguous findings and denominators for affected/unchanged/indeterminate/failed documents. Unchanged requires equal compared evidence and complete comparison coverage. Test target-label reuse, partial rewrites, shifted candidate coordinates, independent scope names and input files unchanged by byte hash. Commit/review after focused tests.


## Task 8: Published wire contracts and independent schema validation


Own contracts/v1/query-document.schema.json, capabilities.schema.json, analysis.schema.json, field-validation.schema.json, schema-validation.schema.json, rewrite.schema.json, corpus.schema.json, manifest.schema.json, graph.schema.json, impact.schema.json, document-view.schema.json and lsp-configuration.schema.json (all under contracts/v1/), contracts/README.md, contracts/sarif/sarif-schema-2.1.0.json with provenance/license files, tests/acceptance/test_machine_contracts.py, testdata/tooling/contracts.json and the narrow accepted development-dependency lock updates. Runtime contract sources live outside `_build_plan/`. Reconcile existing schema locations first and reuse them rather than publish competing definitions.

1. Write positive/negative schema fixtures for canonical QueryDocument, capabilities, analysis, field/schema validation, rewrite, corpus request/report, graph, impact, advanced view and LSP initialization configuration. Include malformed required fields, invalid discriminated unions, unknown new-request keys, duplicate-key decoder cases and additive report-property compatibility. The JSON Schema validator cannot detect duplicate object keys after parsing; test that at the actual decoder too.
2. After final approved dependency provisioning, invoke the full instance validator locally. Select its implementation using validator_for(schema), run check_schema(schema), then iter_errors(instance) with an explicit local registry and no network retrieval. Use Draft2020-12 for owned contracts; honor the actual official SARIF schema's declared dialect. This is a development check, not M4's field-projection engine.
3. Generate actual representative Go reports in test fixtures and validate their serialized JSON, including empty arrays and optional missing versus null behavior. Preserve accepted old wire fixtures. New families use integer schema_version1; family identity comes from the named schema. Document additive output tolerance, strict canonical/new requests, closed-enum changes requiring new contract version and no breaking retrofit to legacy endpoints.
4. Compare capability/diagnostic documentation and manifest values to actual accepted per-dialect forms; no unknown command receives full support because it parses. Official SARIF schema checks supplement Task6 source/URI/index tests. Run the focused schema suite and Go core suites, commit/review, then stop and reserve root's stable Go-only acceptance window before Task9.


## Task 9: Local LSP lifecycle and position conversion


Dispatch only after root completes the Go-only window. Own internal/lsp/protocol.go, server.go, documents.go, diagnostics.go, highlights.go, positions.go and focused tests, plus tests/acceptance/test_lsp_protocol.py. Consume corpus preparation and canonical analysis/document evidence. Produce stdio Serve with LSP3.17 full-sync/diagnostics/documentHighlight support, not new SPL semantics.

1. Write TestInitializeAndShutdown, TestFullSyncDocumentGenerations, TestUTF16Positions, TestStaleDiagnosticsSuppressed and TestHighlightSnapshotCancellation. A transcript must initialize, open with explicit spl/spl2 languageId, modify full text with increasing versions, receive replacement diagnostics, close/clear and shutdown/exit. Include non-BMP characters before/inside identifiers, LF/CRLF/lone CR, duplicate/stale versions, close/reopen resetting version, and unknown method/request versus notification behavior.
2. Run `go test -mod=readonly ./internal/lsp` and observe red. Implement serialized Content-Length framing measured in UTF-8 bytes, JSON-RPC ID/error handling, bounded input sizes and protocol-only stdout. Reuse the final accepted framing dependency or narrow stdio implementation. Initialize precedes ordinary traffic; shutdown gets a response, exit follows protocol status; canceled requests always receive a response.
3. Advertise positionEncoding utf-16, full document synchronization and documentHighlightProvider only. Capture URI/open generation/version/config revision with each analysis snapshot; use bounded work coalescing and serialized publication. Suppress stale diagnostic sets at publication time even for clients without diagnostic-version support. Cancel superseded work where supported without claiming forced ANTLR interruption. Closing disposes state and publishes an empty set; reopening cannot accept a prior generation's results.
4. Convert offsets from exact snapshot bytes to zero-based UTF-16 units, and incoming UTF-16 positions back to checked byte boundaries. Follow protocol line-end clamping rules but reject structurally invalid positions/ranges. A valid surrogate pair contributes two character units. Do not use Unicode columns minus one, tab display widths or JSON-escaped string offsets.
5. Use initializationOptions `{profile, version, validation_target}` with optional inline canonical local target data; configuration notifications use a documented splToolkit settings object with the same shape. Targets are already local data and never retrieval URLs. Prepare with the canonical handle and snapshot configuration revisions. Malformed initialization options/target fail initialization. A malformed configuration change reports a visible service error, invalidates pending publication, clears affected diagnostics and pauses affected analysis/validation publication. Continue tracking buffers/lifecycle; configuration-dependent requests fail clearly. A corrected configuration resumes from the latest open-buffer versions and a new configuration revision, so old work cannot publish. Explicit valid removal of the optional target selects ordinary analysis mode. There is no last-good fallback or extra stale-configuration UI. Unsupported didOpen language IDs do not become line-one syntax findings.
6. Highlight only canonical binding/origin relationships. For `eval a=host | table a`, select the connected write/read evidence; two same-spelled fields in independent scopes do not merge. Ambiguous overlapping cursor references yield no invented group. Capture each position request snapshot in incoming message order and complete it even if later ordinary edits arrive; honor client cancellation with RequestCancelled, and reserve ContentModified for qualifying internal/config invalidation. Ordinary queued edits alone are not justification to drop a request.
7. Use deterministic barriers in version/config/race tests instead of timing guesses. Assert every request ID receives one response, stdout remains parseable after errors and no old diagnostics republish after close. Run focused protocol/Go tests, commit/review. Actual VS Code client delivery remains Task12, not replaced by this harness.


## Task 10: CLI and HTTP adapters with maintained examples


Own new cmd/tooling_cli.go and tooling_cli_test.go, minimal cmd/cli.go dispatch/help, pkg/api/tooling.go and tests, minimal route registrations, accepted OpenAPI update inputs and tests/acceptance/test_tooling_surfaces.py. Task12 owns final broad documentation/package closure; do not rewrite earlier adapter semantics. Consume the final reviewed Go APIs and decoded target wrappers.

1. Write TestScanCLIExitPrecedence, TestCorpusSourceConflict, TestImpactCLIReadonly and TestToolingHTTPBounds. CLI commands are `scan`, `graph`, `impact-schema`, `impact-mapping` and `lsp --stdio`. Corpus commands accept exactly one `--directory <path>` or `--manifest <file>`; reject query/stdin/batch selectors from unrelated operations. scan supports text/json/sarif; graph supports JSON; impact supports text/json. Reuse --output conventions and explicit local target decoding.
2. Define --target <file> for scan/graph optional canonical field_list/json_schema/ocsf target. Schema impact requires --before-target/--after-target. Mapping impact requires --before-rules/--after-rules and optional --before-target/--after-target, using accepted M6 RuleSet/target files. Prepare targets/rules before loader access. Do not let local file argument names imply server filesystem access. Default no target means analysis only; graph contains the selected canonical evaluation context.
3. Run `go test -mod=readonly ./cmd ./pkg/api -run 'TestScanCLI|TestCorpusSource|TestImpactCLI|TestToolingHTTP'`. Implement only decoding, shared loader/service invocation and output. Write content before returning exits 0 valid,1 invalid,3 incomplete; acquisition/traversal/request/write/internal failures yield2. Never overwrite an input query, manifest, target or rule file through --output: detect input/output collision using opened-file/path identity, including aliases, and fail before replacing source content. No source migration writes are supported.
4. Proposed new HTTP routes are POST `/api/v1/corpus/scan`, `/api/v1/corpus/graph`, `/api/v1/corpus/sarif`, `/api/v1/corpus/impact-schema`, `/api/v1/corpus/impact-mapping` and POST `/api/v1/query/document` for advanced view. They accept strict versioned inline documents/options/targets and return direct canonical reports; path-based query acquisition and LSP-over-HTTP are rejected. Reconcile route spelling/helpers with accepted M6 during final release. Content states return200; invalid requests400; body too large follows accepted transport status; internal errors remain500. Body-limit exact-size/+1 tests include two genuine before/after local OCSF catalogs within the approved boundary; limits never silently drop documents.
5. Expose advanced document output through `document` single-query CLI using accepted QueryDocument options and through HTTP/native equivalents; it returns the detached snapshot contract. Existing analyze/discover/map/validate behavior and legacy decoder strictness remain compatible. Add these operations to maintained CLI/API help and tested examples outside `_build_plan/`.
6. Exercise full canonical JSON parity from Go, CLI and local HTTP, normalizing object-key order only. Bind temporary server ports using the accepted harness and collect readiness/cleanup evidence. Regenerate only the actual owned OpenAPI documents through accepted offline commands. Commit/review; do not modify native/package files while Task11 is absent.


## Task 11: Native Python and installed package closure


Own pkg/bindings/bindings.go additions, existing generated C header inputs, python/spl_toolkit/mapper.py methods, python/tests/test_native_tooling.py, python/native-source-files.txt, python/MANIFEST.in/setup/build-support files only where required, tools/check_package.py and focused package-tool tests. Consume final Go operations/strict request decoders. Preserve all accepted M4–M6 exports, guards and installed fixture lists.

1. Write real-library tests for scan_corpus, export_graph, export_sarif, impact_schema, impact_mapping and document_view. Proposed C exports are spl_mapper_scan_corpus, spl_mapper_export_graph, spl_mapper_export_sarif, spl_mapper_impact_schema, spl_mapper_impact_mapping and spl_mapper_document_view, each using the existing handle/document JSON/owned result convention. Final reconciliation fixes exact names consistently across C, Python and docs. Python in-memory documents are canonical dictionaries with explicit IDs; file discovery uses normal caller I/O or documented manifests, not subprocess parsing.
2. Observe failures against the freshly built actual native library; mocks or import success are not the red/green proof. Implement thin native JSON delegation and guarded Python methods that decode and free ownership in finally. Invalid/incomplete content returns dictionaries; malformed requests/internal native failures raise the existing exception. Test concurrent mixed operations, closed handles, invalid Unicode, malformed results/error handling and release ownership exactly once.
3. Extend source-built native closure with every new Go source package actually linked, preserving both accepted frontends and predecessor sources. Schema data/docs included in Python wheels live under an explicit package-data path such as spl_toolkit/contracts; use one canonical source and build-copy rules rather than maintain duplicate schema definitions. New stdio executable code need not become native library source unless imported; include its source in the sdist for release builds as appropriate.
4. Extend mandatory installed-suite and fixture copying with test_native_tooling.py, test_tooling_surfaces.py, test_machine_contracts.py and testdata/tooling/ plus contracts/ and all accepted predecessor inputs. Preserve required tests with nonzero collection and zero skips. Build wheel and sdist through the accepted offline toolchain recipe; rebuild from sdist and execute installed package tests outside checkout. Verify module/native-library origins and exact artifact hashes. A source PYTHONPATH run does not substitute for this.
5. Publish representative installed Go/CLI/native Python equivalent reports for mixed statuses/targets and malformed request errors. Native callbacks contain no field matching, graph inference or rewrite policy. Commit/review the coherent binding/package closure and schedule final build outputs only when root can grant a stable acceptance window.


## Task 12: Real consumers, published artifacts and final evidence


Own tests/editor-client/package.json, package-lock.json, extension.js and test/index.js; tools/check_editor_client.py; tools/check_sarif_consumer.py; tests/acceptance/test_consumer_evidence.py; docs/tooling.md, docs/contracts.md and exact maintained README/architecture/compatibility/Python documentation updates; tools/release.py and package/release manifests only as required; final testdata/tooling cases/examples and `_build_plan/milestones/7-developer-tooling/milestone-log.md`. Read root's candidate identity but never its private expected-input oracles. Do not write fixtures from current output without independent approved assertions.

1. Recheck the prepared consumer lockfiles, wheelhouse and installed tools under `/private/tmp/spl-toolkit-m7-consumer-preflight`. Record missing/mismatched inputs rather than silently install a newer dependency. After final root permission to provision missing accepted inputs, reproduce exact pinned consumers or revise pins with review. Never install into the user VS Code profile. The version-controlled test-only editor client uses official vscode-languageclient9.0.1/protocol3.17.5 inside the real Extension Development Host; it is not distributed as a product editor extension.
2. Launch the existing VS Code executable with temporary --user-data-dir, --extensions-dir, --extensionDevelopmentPath and --extensionTestsPath, sync off and a temporary workspace. The extension contributes spl/spl2 IDs, starts the exact candidate binary `lsp --stdio`, awaits startup and opens/edits documents through VS Code. Read delivered diagnostics with languages.getDiagnostics(uri) after onDidChangeDiagnostics, and invoke vscode.executeDocumentHighlights to exercise the registered real client provider. Capture code/message/severity/UTF16 range and highlight kinds, clearing, version changes and independent same-spelling scopes. A successful process launch or standalone framing transcript is insufficient. Stop the language client and isolated host when finished.
3. Preserve the observed SecCodeCheckValidity warning if still present; diagnose failed GUI/extension-host startup without changing user settings or relaxing sandbox flags. Actual editor automation produces structured evidence keyed to candidate SHA/client version plus its assertions and exit status. GUI start remains a genuine unrun gate until this operation succeeds; CLI --version is merely preflight.
4. Generate SARIF using the exact packaged candidate against durable files with spaces/#/%/non-ASCII names and invalid/incomplete findings, then run sarif-tools3.0.5 summary, CSV and HTML commands locally. Retain real output rows/counts and assert rule/message/severity/encoded URI/start line. CSV preserves paths by default; HTML uses --no-autotrim. sarif-tools does not resolve uriBaseId or verify columns/navigation. Independently validate official schema plus root base-URI resolution, correct single percent-encoding, Unicode columns/end ranges, rule/artifact indices and invocation/coverage state against original snapshots. Never claim the converter proves source navigation.
5. Package machine schemas, capability/diagnostic documentation, licenses and runnable Go/Python/corpus/graph/impact/CI/editor configuration examples in explicit manifests. Reuse tools/release.py, check_package.py, check_clean_build.py, check_reproducible.py and check_acceptance.py. Preserve source-derived version agreement and accepted target set; a version bump/new release target requires the actual final decision rather than implication from this draft. Keep Go1.22 builds and actual installed Python3.14 checks separate from local development/native tests.
6. Run the concrete qualified checks below once immutable candidate inputs are ready. Perform the relevant full/broad review, resolve findings with scoped commits and rechecks, then reserve root's final stable built/native/real-consumer window before claiming completion. Do not mutate HEAD, headers, binaries, wheel/sdist or fixture lists while that window is active.
7. Write the milestone log with `## What's new in the app` first, then changed files, decisions/deviations, exact accepted predecessor/final SHAs, contract/capability versions, actual corpus counts and test/consumer/platform results. State every unrun release matrix or manual gate plainly. Root independently accepts the exact final source/artifacts; no merge/push/publish follows automatically.


## Serial SDD ownership and root windows


After final release only, dispatch one fresh implementer for one task with exact file ownership and all consumed interfaces copied into its prompt. Tell each it is not alone and must preserve every other edit. Each task observes red tests, implements minimally, obtains focused green evidence, commits only owned paths, then receives independent specification and code review against that immutable commit. Fixes produce a new scoped commit/re-review before the next task. Existing untracked CLAUDE.md/copied milestone prompts and other agents' files must never be swept into a broad git add.

Task3 is the sole ordinary owner of canonical target access; later semantic/preparation gaps go back through a scoped reviewed follow-up. Task2 alone owns the platform open module; adapters cannot substitute weaker file loading. Task9 alone owns LSP protocol behavior; Task12's consumer cannot add a fake server or direct diagnostics to make client checks pass. Native/header/package/release owners are serial because they share generated outputs and source manifests. No concurrent implementers own cmd, pkg/api, pkg/bindings, docs manifests or build artifacts.

After Task8 core/schema review, stop mutation and reserve root's Go-only window. Supply immutable commit, source status, approved plan/spec, core reports and focused evidence. Root independently evaluates its own inputs; do not read, run or copy those private cases into fixtures. No Task9–12 adapter work begins until root reports acceptance or specific findings.

After Task12 review and candidate builds, reserve root's final window. Supply exact candidate source SHA and hashes of CLI/server/native/header/wheel/sdist/contracts, real installed Python/native environment and real consumer evidence. Root may rerun its independent Go/CLI/HTTP/Python3.14/editor/SARIF cases. Keep outputs immutable until root responds. Broad review and independent root acceptance remain distinct from implementer unit/surface checks and from configured remote CI.


## Concrete verification commands


All following commands are future execution recipes, not actions taken while drafting. Working directory is `/Users/jacobdelgado/repos/spl-toolkit/.worktrees/remaining-milestones`. Before final release, reconcile paths against accepted M6 without rerunning predecessor suites merely to orient yourself. Provision only approved dependencies; no implicit network fallback.

    export GOCACHE=/private/tmp/spl-toolkit-go-cache
    export GOMODCACHE=/private/tmp/spl-toolkit-remaining-gomodcache
    export GOPROXY=off
    export GOTOOLCHAIN=local
    export PIP_CACHE_DIR=/private/tmp/spl-toolkit-pip-cache
    go test -mod=readonly ./internal/corpusfs ./pkg/corpusio ./pkg/corpus ./pkg/document ./pkg/graph ./pkg/sarif ./pkg/impact
    /private/tmp/spl-toolkit-floor-tools/gomodcache/golang.org/toolchain@v0.0.1-go1.22.12.darwin-arm64/bin/go test -mod=readonly -ldflags=-linkmode=external ./internal/corpusfs ./pkg/corpusio ./pkg/corpus ./pkg/document ./pkg/graph ./pkg/sarif ./pkg/impact

Expected: focused tests pass, acquired sources are unchanged, and error/status assertions retain uncertainty. A missing dependency/cache or unavailable runtime OS is an environmental gate, not a product semantic failure and not permission to fetch newer modules. The Go1.22 external-link flag is the accepted preceding-plan workaround for local dyld LC_UUID behavior; verify the final M6 recipe before finalizing.

After the core window and reviewed adapter work:

    go test -mod=readonly ./internal/lsp ./cmd ./pkg/api ./pkg/bindings
    make build-all
    /private/tmp/spl-toolkit-remaining-venv/bin/python tools/check_go.py
    /private/tmp/spl-toolkit-remaining-venv/bin/python -m pytest tools/tests -q
    /private/tmp/spl-toolkit-remaining-venv/bin/python tools/check_docs.py
    /private/tmp/spl-toolkit-remaining-venv/bin/python -m pytest tests/acceptance/test_machine_contracts.py -q

The accepted check_go tool owns the full race suite; repeat full races only after relevant changes/failures justify it. If M7 does not change grammar, do not regenerate parser outputs. Core target access, source conversion and protocol work do not by themselves require grammar edits.

CLI/HTTP/native parity harnesses must use accepted executable/environment conventions and isolated temporary ports. Proposed new tooling fixtures environment is SPL_TOOLING_FIXTURES; preserve all actual accepted predecessor fixture variables. Source-native smoke commands after an actual build include:

    SPL_CLI=/Users/jacobdelgado/repos/spl-toolkit/.worktrees/remaining-milestones/build/spl-toolkit SPL_TOOLING_FIXTURES=/Users/jacobdelgado/repos/spl-toolkit/.worktrees/remaining-milestones/testdata/tooling /private/tmp/spl-toolkit-remaining-venv/bin/python -m pytest tests/acceptance/test_tooling_surfaces.py tests/acceptance/test_lsp_protocol.py -q
    PYTHONPATH=python SPL_NATIVE_LIBRARY=/Users/jacobdelgado/repos/spl-toolkit/.worktrees/remaining-milestones/build/libspl_toolkit.dylib SPL_TOOLING_FIXTURES=/Users/jacobdelgado/repos/spl-toolkit/.worktrees/remaining-milestones/testdata/tooling /private/tmp/spl-toolkit-remaining-venv/bin/python -m pytest python/tests/test_native_tooling.py -q

Reconcile full mandatory native fixture/environment names before running the entire installed suite; do not skip a predecessor test merely because its fixtures were not copied. No code/test depends on these temporary plan documents.

Use accepted offline OpenAPI/package recipes, after reconciling their actual M6 inputs:

    GOCACHE=/private/tmp/spl-toolkit-go-cache GOMODCACHE=/private/tmp/spl-toolkit-remaining-gomodcache GOPROXY=file:///private/tmp/spl-toolkit-remaining-gomodcache/cache/download GOSUMDB=off GOTOOLCHAIN=local make generate-docs PYTHON=/private/tmp/spl-toolkit-remaining-venv/bin/python
    PATH=/private/tmp/spl-toolkit-floor-tools/gomodcache/golang.org/toolchain@v0.0.1-go1.22.12.darwin-arm64/bin:$PATH PIP_NO_INDEX=1 make python-build PYTHON=/private/tmp/spl-toolkit-remaining-venv/bin/python PIP=/private/tmp/spl-toolkit-remaining-venv/bin/pip
    /private/tmp/spl-toolkit-remaining-venv/bin/python tools/check_package.py --sdist dist/spl_toolkit-$(cat VERSION).tar.gz --wheel-dir dist --evidence /private/tmp/spl-toolkit-m7-package-evidence.json

Expected package evidence records real outside-checkout installed module/native origins, source and artifact hashes, nonzero mandatory tooling and predecessor suite counts with zero skips. It must include wheel-from-sdist testing. A source-tree PYTHONPATH pass is not this result. Root owns independently selected actual Python3.14 environment/acceptance.

The planned real-editor wrapper receives an explicit candidate executable, existing editor path and isolated consumer directory; Task12 defines this CLI and prevents profile/network defaults:

    /private/tmp/spl-toolkit-remaining-venv/bin/python tools/check_editor_client.py --cli build/spl-toolkit --vscode /usr/local/bin/code --consumer-root /private/tmp/spl-toolkit-m7-consumer-preflight --source-sha $(git rev-parse HEAD) --evidence /private/tmp/spl-toolkit-m7-editor-evidence.json
    /private/tmp/spl-toolkit-remaining-venv/bin/python tools/check_sarif_consumer.py --cli build/spl-toolkit --sarif /private/tmp/spl-toolkit-m7-consumer-preflight/venv/bin/sarif --source-sha $(git rev-parse HEAD) --evidence /private/tmp/spl-toolkit-m7-sarif-evidence.json

The wrappers are proposed Task12 artifacts and do not exist because this draft names them. They execute the exact reviewed editor/consumer recipes, assert actual delivered findings and record artifact hashes; they cannot satisfy gates with file-open or startup success. Their development tests use explicit captured fixtures only for wrapper error handling, never as substitutes for the real acceptance invocation.

Local host pins currently differ from tools/release-env.json. check_reproducible labels unpinned successful runs diagnostic-passed and must not be edited to promote them. Do not install Xcode/change pins for local acceptance. Preserve the configured four-target CI matrix as future executable infrastructure, not current-source evidence; remote dispatch belongs to root and is not authorized by this plan.


## Validation and Acceptance


Task12 materializes `testdata/tooling/example-corpus.json` with schema_version1 and inline documents `{id: good, text: "search host=web"}`, `{id: missing, text: "search missing=x"}` and `{id: unknown, text: "| mystery | table host"}`. `testdata/tooling/example-target.json` contains `{kind: field_list, catalog: ["host"]}` using accepted target syntax. The CLI command `build/spl-toolkit scan --manifest testdata/tooling/example-corpus.json --target testdata/tooling/example-target.json --format json` exits1, retains three report entries, reports invalid aggregate status and incomplete semantic coverage. It does not call the unknown command complete because host matches a catalog.

A second manifest adds an explicit missing-file entry to the same inline entries. scan writes all known findings, shows selected4/analyzed3/acquisition_failed1 and ExecutionComplete false, and exits2. Reordering manifest entries changes entry order but not each stable ID; directory inputs sort by relative path. A symlink escape or deterministic component swap cannot read an external sentinel. Ignored .git/node_modules trees are selection exclusions, not cleanly analyzed sources.

The graph command produces unique revision-qualified nodes/edges whose original reference/source pointers resolve. A query `eval owner=host | table owner` contains a derived chain linked to host; independent child-scope owner names are not one symbol. Unknown effects remain partial graph coverage. Unicode range slices must match their original query bytes.

Schema impact with before catalog host and after an empty catalog changes host's field outcome even if another missing field keeps both statuses invalid. Mapping impact previews explicit before/after rules, retains candidate/audit/validation evidence and proves the original source files are byte-identical after the operation. Equal incomplete results classify as indeterminate about unobserved impact. No report promises runtime equivalence.

SARIF export validates with the official schema and semantic source tests and produces meaningful sarif-tools rows/counts for errors and incomplete warnings. Encoded URI/startLine ingestion is recorded; full root resolution/percent-encoding/Unicode-column correctness comes from independent semantic assertions, not a claim of converter navigation. Empty findings and operational failures have correct result/invocation states.

The real VS Code extension host uses its official languageclient to receive diagnostics in editor API storage and highlights through its provider command. Correcting text clears old findings; rapid versions/config changes do not publish stale results; scope boundaries prevent same-spelling over-highlighting. Capture actual ranges, kinds, client version and candidate identity. If the editor cannot start, retain the warning/failure and the gate remains open rather than substituting protocol harness success.


## Idempotence and Recovery


Every loader/export/impact operation is read-only with respect to query files; snapshot before comparing and fail source/output identity collisions. Re-running tests uses fresh task-owned temporary roots and deterministic fixtures. Clean only artifacts created by the current task; never delete another agent's worktree, untracked file, cache, native output or active user profile. Keep filesystem handles/stdio clients/subprocesses scoped and explicitly closed on success/failure.

On failed task checks, preserve the failing example and classify product behavior versus missing tool/dependency/runtime acceptance. Make the smallest owned fix, rerun focused affected checks and review a new immutable commit. Rebuild native/package artifacts after linked source changes; stale binaries do not provide current acceptance. During root windows, report a needed change and wait for root to release mutation before touching inputs. Do not rerun broad unchanged suites to obscure an unresolved specific gate.

If a final accepted predecessor differs from a proposal, update the bounded register and every affected interface/test/adapter entry before release. Do not reopen settled product choices unless the real dependency prevents their sound implementation; raise that concrete conflict to root. All updates keep Progress, discoveries, decisions and outcomes accurate.


## Artifacts and Notes


The prepared consumer dependency paths were rechecked during drafting, not executed: node-consumer/package-lock.json SHA256 `fbdc562f435c699c13a09088f10bc2c61866524c76ba936d5e3eecf1006f1757`; requirements.lock SHA256 `c7b47a61d1b509546af4d9a67050c164fe0117fd71f6bcd93f8dc2304e926e0b`. Those hashes were read from the accepted preflight evidence, not freshly recomputed in this draft. Rehash before later acceptance. The prepared state is expressly `consumer-dependencies-prepared-product-unverified`.

All proposed source/test/schema/client files in this plan are future task ownership. Drafting did not create them. The independently useful result of each task is named with its tests and immutable review gate; the milestone's two root windows and final actual-client evidence cannot be inferred from those task checks.

Revision note (2026-09-08): Initial bounded draft created from the approved M7 design, immutable reviewed M4 core, approved but unimplemented M5/M6 proposals and accepted consumer preparation. Added root-approved anchored-open and canonical-preparation reconciliation; no production authorization or final interface freeze is implied.

Revision note (2026-09-08, self-review): Recorded root-approved configuration pause/resume and 8 MiB route policy; made impact preparation-before-acquisition handles explicit, named each contract file and clarified detached revision identity. Actual canonical preparation internals remain accepted-M6 reconciliation, not a duplicate rules/schema engine.

Revision note (2026-09-08, root-reviewed correction): Both intended unknown-command examples now start with `| mystery`, distinguishing an explicit unknown command from implicit SPL search text. Independent bounded draft architecture review found no additional gap; root approved the thin document CLI as canonical advanced-view serialization. The six accepted-M6 reconciliation items and final production gate remain unchanged.
