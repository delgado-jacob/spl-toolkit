# Standalone SPL2 Implementation Plan


> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development. Root released production on 2026-09-08 against fully accepted M4 0bb169c9fb1f0212d1e7758be6ab35a1804bbbaf. Execute serially with a fresh implementer per task, a combined independent specification/code review of each immutable task commit, fixes and scoped re-review, then a separate broad milestone review and independent root acceptance. Execute the accepted serial tasks within the stated ownership and root acceptance windows.

**Goal:** Let callers explicitly analyze and validate standalone splunkd SPL2, including native pipeline and FROM/SELECT search forms, with truthful syntax coverage, located field evidence and shared field/schema semantics.

**Architecture:** Add a dedicated Go-target ANTLR SPL2 frontend. Its typed contexts supply source-located operands to the existing field-state, reference, source-universe and validation machinery; they do not create a second flow engine. Register SQL clauses in lexical order, execute their real semantic phases in logical order, and retain the original source throughout.

**Tech Stack:** Go 1.22 minimum; existing github.com/antlr4-go/antlr/v4 v4.13.1; pinned ANTLR generator 4.13.2; existing Go CLI/HTTP/C ABI; Python 3.11+ ctypes and installed-package acceptance. No new runtime service, query execution or network dependency.

**Spec:** The approved design, conformance matrix, form expansion and their reviews in this directory. Their required behavior is restated below; runtime code, fixtures, installed packages and permanent documentation must never depend on _build_plan.

This ExecPlan is a living document maintained under /Users/jacobdelgado/.codex/PLANS.md. That file requires checkboxes only in Progress, so the detailed task steps below use numbered actions rather than the writing-plans skill's generic checkbox template. The controller explicitly selected this path, serial SDD and draft-only scope; the skill's default plan location and execution-choice question do not apply.


## Purpose / Big Picture


After acceptance, a caller can select language spl2 and analyze `SELECT host FROM main WHERE bytes>0`. The result retains SELECT/FROM/WHERE in original source order, records bytes as an input read before SELECT projection, and reports lineage in FROM/WHERE/SELECT order. Field-list, local JSON Schema and pinned OCSF validation consume those same references and state. A malformed supported form is invalid; a recognized but unmodeled form remains incomplete even if a target schema contains every mentioned field. Existing omitted-language SPL calls and legacy fixed-SPL Go/C methods retain their behavior.

This is static analysis, not execution or certification against a live Splunk server. The splunkd/current contract describes this build's documentation snapshot. Modules, imports, exports, reusable declarations and user-defined functions remain excluded. Inline lambdas and named built-in calls are recognized syntax with deliberately incomplete semantics.


## Global Constraints


All query parsing uses Go plus ANTLR grammar contexts. Do not split source text at pipes, implement a regex parser, transpile whole queries into SPL, manufacture SPL contexts for SPL2, or add Python/REST semantic inference.

Explicit QueryDocument language is spl or spl2; empty/omitted means spl. Profile is splunkd; version is current. Unknown nonempty selectors or invalid Unicode are input errors, never fallback selection. Text and source_id stay exact caller data.

The approved breadth is 35 dedicated command families, 25 language families, all 53 compatibility-table entries accounted for, and at least 200 deduplicated meaningful fixtures including 80 definite-negative/profile cases and 40 deliberate SQL/mixed cases. These are floors, not a substitute for per-form coverage. Every admitted option, alternative and finite positional arity has the reviewed positive/minimal-negative obligations; optional omissions and variadic unbounded forms must not acquire false negatives.

Original source bytes, UTF-8 byte offsets, Unicode columns, CRLF semantics, copied public values, stable IDs and deterministic array ordering are required. All mutable parsing, recovery, lowering, scope and source-refinement state is per call. Keep the legacy SPL analysis CharStream marker immutable and local; do not extend it into a global dialect mode.

Known internals and tombstones are preserved by the canonical field engine. Fields inclusion must not fabricate _raw/_time or resurrect earlier removals. External source membership is distinct from event presence and structural field availability. Derived fields bypass external schema membership; unknown stages cannot become complete because a catalog contains matching names.

Module/declaration/multiple-statement exclusions produce SPL_UNSUPPORTED_MODULE; known wrong-profile constructs produce SPL_PROFILE_MISMATCH. Both are located invalid findings for this selected contract with incomplete coverage. Unknown standalone syntax stays incomplete. Supported malformed syntax remains invalid despite surrounding recovery. Definite errors outrank incompleteness without hiding incomplete coverage.

Keep default SPL wire values unchanged where new metadata is unnecessary. No push, merge, publication, release tag, destructive cleanup or worktree deletion is authorized by this plan.


## Progress


- [x] (2026-09-08 UTC) Read writing-plans and PLANS.md; reviewed approved M5 source artifacts and accepted predecessor code without running predecessor checks.
- [x] (2026-09-08 UTC) Controller approved CapabilityOptions/CapabilitiesFor and concrete additive SQL phase/order fields for this draft.
- [x] (2026-09-08 UTC) Draft independent tasks, corpus obligations, ownership and verification recipes.
- [x] (2026-09-08 UTC) Read accepted M4 Task 2 schema types/preparer at bacc2172291005b0c70cdbba948512173090f9aa; public schema report/dispatcher remains gated.
- [x] (2026-09-08 UTC) Reconciled reviewed M4 Tasks 1–4 at 5e3d4ec3a54b6878830b83a8ff08405011a507d9: actual schema request/report/decoder/validator APIs and string-only path boundary; no analysis-hook changes.
- [x] (2026-09-08 UTC) Reconciled accepted M4 Task 5 CLI/REST at c91ca8464e916ae0b3792679a195a76e936c9385 and Task 6 native/Python/source closure at 7b6580fff46917045230668141d23f17489891df through immutable code and their accepted reports/reviews. Active Task 7 product files were not inspected.
- [x] (2026-09-08 UTC) Reconciled accepted M4 Task 7 and I1 fix at 92879052b0454db12aead045188ce4b15fa9d83a: complete-header comparison, schema surface/docs registration, qualified installed evidence and strict-offline package recipe. Root's local acceptance metadata is recorded as predecessor evidence.
- [x] (2026-09-08 UTC) Full M4 broad-review/controller closure accepted at 0bb169c9fb1f0212d1e7758be6ab35a1804bbbaf; exact five-path docs/evidence-only delta from 9287905, no interface/code change.
- [x] (2026-09-08 UTC) Root approved the reconciled plan plus genuine-toolchain header-fixture clarification and explicitly released serial production.
- [x] (2026-09-08 UTC) Task 1: shared located transfer boundary and selector contract; 09a167845956f909639824a1781d4895ea26972d, combined review clean; focused/full analysis+validation pass and fourteen baseline report pairs preserved.
- [x] (2026-09-08 UTC) Task 2: dedicated SPL2 lexer, expressions and durable conformance harness; 5f85b8b64933cd0c020331ee1d53c3418730b100, combined review plus two scoped fix reviews clean. Current focused syntax/corpus/tooling and deterministic generation pass; intermediate 179 queries / 36 meaningful groups / 13 negative groups / zero SQL credit, semantics pending.
- [x] (2026-09-08 UTC) Task 3: ordinary pipeline grammar; e3c1a256192ef8e7dcb1c775156a9e5b1c65c2bc, combined review plus two scoped fix reviews clean. Current 781 focused checks, 16 tooling tests, two-pass generation and 27 input hashes pass; 663 queries / 223 meaningful / 106 negative groups, with 681 obligations and all semantics still pending.
- [x] (2026-09-08 UTC) Task 4: both SQL clause hierarchies and located contracts; 42e261db46d796b057da431dec175d3bbe0b063c, combined review plus scoped fix review clean. Current 1136 focused checks, 16 tooling tests and 61 input hashes pass; retained grammar/generated identities unchanged after fix. 945 queries / 288 meaningful / 139 negative / 45 SQL groups, 565 obligations and all semantics still pending.
- [x] (2026-09-08 UTC) Task 5: remaining dedicated grammar and typed child scopes; 2d7a758257c753112b33ca3db2951364d209f788, combined review plus scoped fix review clean. Current 1926 focused checks, 16 tooling tests, two-pass generation and 66 input hashes pass; 1545 queries / 288 meaningful / 139 negative / 45 SQL groups, 189 obligations and all semantics still pending. Descriptor-specific preimplementation RED sequencing limitation remains disclosed; no historical reconstruction.
- [x] Task 6: complete ordinary SPL2 lowering.
- [x] Task 7: SQL phases and supported SQL lowering.
- [x] Task 8: recovery, deferred boundaries, capability truth and canonical validation.
- [x] Root's committed Go-only canonical acceptance window: 16 analysis, 9 schema and 4 field-list cases, legacy/mixed batches and 68 interleaved analysis calls passed at 471e28746c288fea544d5cc4652d37ecad636a2c before Task 9 dispatch.
- [x] Task 9: CLI and REST dialect parity at b9c98c76e1b0b1da4ae2f66c3094715fbd0b6235; combined review and scoped Unicode-fold fix review clean, full report/docs/OpenAPI checks and qualified frozen-input proof retained.
- [x] Task 10: native Python, C ABI additions and package closure at b754b9bfa80eb990cb83ee05d3b4588c2c2e54c3; combined review plus scoped refusal-category fix review clean. 65-source closure, genuine 18-export compiler headers and both installed package paths are verified; historical evidence qualifications remain explicit.
- [x] (2026-09-08 UTC) Task 11: full corpus/surface acceptance and maintained docs at fcbb74ac60da312bc06275e4e3de29b570a3bf2e; fresh combined spec/quality review Approved with no actionable findings. All 1,714 canonical cases executed after integration (732 independent and 982 independently reviewed snapshots); required local Go, tooling, source and both offline package gates pass. Final milestone acceptance and authored handoff remain the separate entries below.
- [x] (2026-09-08 UTC) Root's final stable built/native/surface/Python 3.14 acceptance passed at fcbb74ac60da312bc06275e4e3de29b570a3bf2e. Python 3.14.1 ran 339 tests with zero skips; actual native/CLI/HTTP analysis, field/schema validation, concurrency, batch/ownership/refusal and file/stdin/output checks passed. Product source/build inputs remained fixed. Root receipt SHAdaa4b3e94fbc514e08c2b4caaf58c1111daa0333693b9d0fac268f175a4aac95 preserves the initial private harness-option failure and its corrected documented route; no product change or duplicate Python run.
- [x] (2026-09-08 UTC) Separate broad milestone review completed across all 91 handwritten paths. M5-BR-01 (len/split result domains) and M5-BR-02 (module category) were reproduced and fixed in 7571f1cb578968caf92996b050ebfc3092f02605; scoped review Approved at evidence-only child 40e6e171a1d215e5c67758fbf346cc5b78892ee2, with no actionable findings.
- [x] (2026-09-08 UTC) Root reacceptance passed at 40e6e171 with fixed product/build inputs: 16 analysis cases plus legacy, 14 Task11 semantic controls, two fix validation controls, nine schema/four field cases, 68 concurrent calls and 339 Python 3.14 tests with zero skips. The original Task8 baseline is preserved with only the independently approved module-category comparison amendment.
- [x] (2026-09-08 UTC) Root released the exact fourteen-path authored-document/evidence closure. The ten original design/conformance/research/review documents remain byte-for-byte unchanged; this living plan and milestone log record the accepted results. The closure identity check binds the resulting documentation-only commit to runtime-tested 40e6e171 and unchanged builds.


## Surprises & Discoveries


At accepted M4 hook commit 47b560d710ed264490e6c21711ad157146b8fa78, pkg/analysis/flow.go and model.go are byte-identical to accepted M3 e401402e5d8cd80f469884b2c3ad80245266ad66. The environment itself has no parser dependency. Most coupling instead occurs in typed command/expression wrappers, parsedDocument, scope scheduling and selector extraction. This supports a small shared located-operand boundary rather than a new public intermediate representation.

The accepted function registry gives legacy SPL coalesce a minimum of two arguments and reports unsupported arity through its existing warning policy. SPL2's approved positional coalesce admits one or more, and definite bad arity has a located contract-invalid outcome. Keep these policies dialect-specific; do not change SPL to make SPL2 tests pass.

The existing originalTokenBoundaries recovery helper treats brackets as SPL subquery ownership. SPL2 also has arrays, named lists and inherited branch bodies, so that helper cannot be reused unchanged. Original typed delimiter ownership must decide the child kind.

The form-expansion review corrected a dangerous assembly instruction: all E.L01.start candidates are complete inputs. In particular, E.L01.start.N1 is exactly `failure index=app`; prefixing FROM would change the intended invalid start into a different unknown-command scenario. Preserve its bytes.

The accepted source-universe resolver accepts a name string rather than a path object. Reviewed M4 Task 4 at 5e3d4ec confirms that prepared.project receives ref.NormalizedName or a proven expansion member name, without typed query path/alias metadata. Literal dotted identifiers, nested navigation and SQL alias qualification must not be collapsed into a conclusive schema path by M5. Retain the approved incomplete boundary where these typed frontend distinctions cannot be proved by the existing canonical representation.


## Decision Log


2026-09-08, controller: Draft planning may proceed during M4, using e401402 and 47b560d as immutable evidence. Active uncommitted M4 schema fixes are not final interfaces. Final approval and all implementation still require reviewed integration and fully accepted M4.

2026-09-08, controller: Add CapabilityOptions with Language/Profile/Version JSON selectors and CapabilitiesFor(options) returning a manifest or input error. Keep Capabilities(), old C ABI and no-argument Python behavior unchanged. Use one normalized selector contract across analysis and capabilities.

2026-09-08, controller: Add Lineage.Phase string with omitempty and ExecutionOrder *int with omitempty. SQL lineage entries carry a finite phase vocabulary and a zero-based index in actual report lineage order. ScopeID determines scope meaning; global order does not imply scope inheritance. Existing SPL omits both fields.

2026-09-08, controller: Include evaluate for SELECT-owned nonaggregate preparation, distinct from aggregate and final project. Register plain projection reads once during evaluation; final restriction reuses those reference IDs. The phase does not authorize additional HAVING contexts, hidden fields, unselected aggregates or alias collisions.

Earlier approved decisions retained: dedicated SPL2 frontend, shared field/schema transfer, no source transpilation; independent exact rename pairs only, rejecting duplicate/chain/overlap forms; count() implicit label count only where proved; named colon calls typed but semantically incomplete throughout M5; ordinary positional equality expressions remain legal; hidden HAVING/ORDER visibility incomplete; final projection SELECT-owned; modules and profile mismatches explicitly invalid for this contract.

2026-09-08, draft author: Keep generated SPL2 code under parser/spl2 and handwritten lowering under pkg/analysis. The existing parser namespace and legacy frontend stay intact. Private location-based transfer functions are preferred over exporting parser-independent AST types or implementing a second whole-query engine. Reconcile this small boundary against accepted M4 before final approval.

2026-09-08, controller: Architecture/task sequence at 0c7c2e787ee8ff42f2c90faeb2d339ca1a9439f8a6980b458fb38519964a5dc3 is approved, with no repeated broad design/architecture gate. Reconcile only actual reviewed M4 core APIs now; full adapter/package acceptance and final production release remain open. Reserve root acceptance windows after Task 8 review and after adapter/package work.


2026-09-08, root and qualifying Task11 coordinator: Retain original query/ID/raw-syntax/source/provenance identities during conformance closure. Exact independently authored canonical values are separate from 982 representation/recovery snapshots, which require independently frozen nonvacuous assertions and semantic review. Capture failures and every versioned authority correction remain retained. The only additional original-record metadata is the approved E.L01.start.N1 unknown-command recovery classification and floor exclusion; its source bytes and raw negative syntax obligation are unchanged. Full-Go transport is parity evidence and adds no semantic credit.

2026-09-08, root and qualifying Task11 coordinator: Recovery preserves independently sound original typed owners and operands, not every apparent source-written intention. No additional SqlSpanCall, WHERE, projection, or late-JOIN reconstruction is authorized. The existing empty-GROUP SELECT retry, Q06 bare-key/empty-HAVING Identifier proof, intact-original-GROUP input EOF attribution, and same-range recovered-native-count EOF attribution have distinct private typed/token ownership requirements; all unrelated diagnostics and invalidity remain. No generic EOF waiver, fake node, broad message match, or global diagnostic deletion is allowed.

2026-09-08, root and qualifying Task11 coordinator: Parent option errors do not erase an independently intact originally bracketed body owned by an actual recognized command. Missing original delimiters or damaged containing bodies do not create child scopes. Named UNION inputs retain dataset dependencies without merge effects. Intact deferred input/output intentions remain located without invented state; generating metrics/timechart outputs remain conditional/open/uncertain, and wildcard grouping does not install literal pattern members. Static loadjob SIDs are search_job references without a new Dependencies member; tstats datamodel_name is one atomic data_model reference/dependency. These identities alone do not grant rewrite authority.

2026-09-08, root and qualifying Task11 coordinator: An intact supported zero-argument count has a per-output presence proof only with original projection/destination soundness and no source, sibling, dynamic, or unproved destination collision. Qualified suffix evidence is used only to check potential collisions, never to install a root field. A finite final SELECT membership set may be known while individual fields remain conditional and coverage incomplete. If selected inputs/origins depend on incomplete recorded source expansion, uncertainty remains. This distinction explicitly supersedes the older uncertainty inference for E.L16.alias.P2 and I.literal.P2 while preserving their source, opaque-literal, conditionality and coverage obligations.

2026-09-08, root: The user-required efficient-delegation capability floor governs all remaining protected work. The M5 replacement coordinator and retained/new protected delegates are explicitly assigned gpt-6-astra/xhigh with non-overriding default/worker roles. Assignment/runtime configuration is recorded honestly without backend attestation. Historical accepted work and raw failures remain evidence rather than being re-audited solely because coordination changed.


2026-09-08, root and qualifying broad reviewer: The len result domain is numeric; split uses the existing unknown domain because a multivalue result does not prove scalar string compatibility. Preserve the existing argument, nullability, source and arity guards. Conditional field binding alone need not make analysis status or semantic coverage incomplete; field and schema validation add their own indeterminate coverage when a queried field is conditional. The scoped fix is source commit 7571f1cb578968caf92996b050ebfc3092f02605, with the exact nested-len failure retained before mutation.

2026-09-08, root: Module declarations and suffixes must emit SPL_UNSUPPORTED_MODULE with severity error and category unsupported_syntax, as the original design requires. The scoped fix changes only the two category arguments and focused assertions. All 1,714 compact canonical records and original fixtures remain unchanged; a pre-capture source-derived amendment permits exactly five full-report category leaves for B08–B12, with the other 1,709 reports equal. Root retains its original Task8 baseline and separately authorizes only the matching module-category path.


## Outcomes & Retrospective


This plan is root-approved for implementation from accepted M4 0bb169c9fb1f0212d1e7758be6ab35a1804bbbaf. Planning evidence does not claim parser conformance. The syntax research and draft architecture gates are closed by controller approval. M4 core API/path reconciliation is complete at reviewed 5e3d4ec, CLI/REST at accepted c91ca84, native/Python/source machinery at accepted 7b6580f, and Task 7 checker/header/surface/docs/evidence at accepted 9287905. Root local M4 acceptance passed at that pin. Full M4 broad-review/controller closure and plan/production release were accepted before M5 implementation. That preparation changed no production files, tests, fixtures, builds or commits. Tasks 1–11 and their independent combined/scoped reviews are now complete through fcbb74ac60da312bc06275e4e3de29b570a3bf2e. Final root built/native/Python 3.14 acceptance passed with fixed inputs. The broad review completed with two scoped findings: len/split return-domain classification and module diagnostic category. Both are fixed in 7571f1c and independently approved at evidence-only candidate 40e6e171. Affected floor, vet/race, source and both installed-package checks passed, with 1,714 unchanged canonical objects and exactly five approved full-report category amendments. Root repeated its independent runtime acceptance at 40e6e171, including the two new composition/validation controls and Python 3.14, and approved the scoped authored-document closure. M5 implementation, verification and independent reviews are complete. Later milestones retain their own plan and production release gates; no merge, push, release matrix or live Splunk certification is claimed.


## Context and Orientation


Work in /Users/jacobdelgado/repos/spl-toolkit/.worktrees/remaining-milestones on codex/remaining-milestones. All paths below are relative to that root. Other agents share the tree; each fresh implementer owns only its listed files, must preserve others' work, and must not restore or delete unrelated changes.

pkg/analysis/analyze.go normalizes a QueryDocument, creates a Result, parses, performs semantic analysis, then finalizes status. AnalyzeWithSourceFields and AnalyzeWithSourceUniverse in source_fields.go funnel through the same private analyze function. A SourceUniverse carries candidate names, a completeness promise and an optional pure admission callback. Fields is copied; callback state is not. The callback answers external name admission, never query flow or event presence.

At 47b560d the important public contract is:

    type SourceUniverse struct {
        Fields []string
        Complete bool
        Resolve func(name string) SourceFieldAdmission
    }
    func Analyze(document QueryDocument) (*Result, error)
    func AnalyzeWithSourceFields(document QueryDocument, fields []string) (*SourceAnalysis, error)
    func AnalyzeWithSourceUniverse(document QueryDocument, universe SourceUniverse) (*SourceAnalysis, error)

SourceFieldAdmission has indeterminate, admitted and prohibited values. Complete=true prohibits unlisted names without invoking Resolve. Partial universes can admit unlisted exact names. SourceAnalysis contains Result and FieldExpansion entries; finalizeReferences remaps their IDs. The older finite hook retains its empty-on-incomplete compatibility behavior. Preserve the accepted tracked-internal admission regression and preceding-read/projection behavior.

flow.go owns trackedField and environment, including source/derived provenance, conditional binding, removed names, open schema state and historical uncertainty. references.go mixes source-context wrappers with referenceAt, reads, creation and final ID remapping. commands.go mixes SPL context extraction with projection, rename, aggregate and lookup effects. functions.go has the legacy registry and typed SPL call checking. scopes.go allocates lexical stages, runs SPL transfers and records one lineage entry per stage. source.go converts ANTLR rune indexes to public original-byte locations. recovery.go uses SPL-specific token classes and reparsing; it remains an SPL frontend concern.

pkg/validation/validate.go consumes canonical source analysis rather than replaying transfers. Its request.go has a separate selector normalizer that currently admits only spl; align it with the canonical selector helper while preserving InputError wrappers and strict JSON/Unicode behavior. Reviewed M4 5e3d4ec adds the shared finalizeValidation helper used by field-list and schema reports; preserve its deterministic sorting, exact diagnostic deduplication and invalid-over-incomplete precedence. Public schema reports/decoders/validators are reviewed core evidence; accepted adapter/package reconciliation is recorded below through 9287905, with only full M4 closure outstanding.

Accepted M4 Task 2 bacc2172291005b0c70cdbba948512173090f9aa adds SchemaTarget, SchemaTargetInfo, SchemaEvidence and the private preparedSchemaTarget interface in pkg/validation/schema_model.go, plus DecodeSchemaTarget in schema_target.go. Reviewed Tasks 1–4 at 5e3d4ec retain these shapes and add the public reports/dispatch below, including OCSF preparation. SchemaTarget carries Kind/Identity, raw Schema, BaseURI, local Resources, raw Catalog and Selection. The private interface is project(string) fieldProjection, universe() analysis.SourceUniverse and info() SchemaTargetInfo. This confirms string-name admission, including schema-side literal/nested ambiguity detection; it does not supply a typed query path or alias discriminator. Retain distinct SPL2 identifier/navigation/alias contexts and incomplete validation where the string-only canonical evidence cannot prove the intended interpretation. No new validator or string-based guessed resolution is authorized.

cmd/cli.go owns command dispatch, selector flags, presence tracking and legacy operations. cmd/validation_cli.go implements field-list validation; accepted cmd/schema_validation_cli.go implements schema/OCSF validation. pkg/api/analysis.go handles /api/v1/query/analyze and /api/v1/capabilities; pkg/api/schema_validation.go contains the accepted schema handlers and bounded body reader. pkg/api/models.go/handlers.go hold legacy request types and handlers. pkg/bindings/bindings.go has ownedMapperJSONResult, native handle admission and result allocation/free. python/spl_toolkit/mapper.py coordinates calls with _operation and finally/free. python/native-source-files.txt explicitly enumerates Go source closure. tools/check_package.py copies required acceptance inputs outside the checkout and requires nonzero, zero-skip real installed tests.

The root-approved design evidence is conformance-matrix.md SHA256 84d07d0ba4638cff5e3e7cf021698a9d570d0b4a81f46414a670f0eb8491fe51; form-expansion.md 6eb5f7242f8dd8078ce9ec2daed06a374c0fec873dfb314156cc83d5cb36e2ba; research-notes.md 002b2176f662c95058ecd894bcb27f9b3df629580a3723a0597ac7fbd04166ad. Preserve conformance-review.md, conformance-rereview.md, form-expansion-review.md and the controller's scoped re-review. They are implementation obligations, never runtime package inputs.


## Reviewed M4 core reconciliation and remaining preflight


Core reconciliation uses only git show at 5e3d4ec3a54b6878830b83a8ff08405011a507d9, supplied by root as independently reviewed M4 Tasks 1–4. The pkg/analysis tree has no differences from accepted hook 47b560d. SourceUniverse, partial/finite sidecar behavior, tracked-internal admission, model.go Lineage and finalizeReferences therefore retain their verified shapes; the approved prepared-selection, per-output lookup and additive SQL phase proposals need no architectural change. No predecessor tests were rerun.

The following public functions now exist in pkg/validation/schema_request.go, schema_target.go, schema_validate.go and schema_batch.go:

    func DecodeSchemaTarget(data []byte) (SchemaTarget, error)
    func DecodeSchemaRequest(data []byte) (SchemaRequest, error)
    func DecodeSchemaBatchRequest(data []byte) (SchemaBatchRequest, error)
    func ValidateSchema(document analysis.QueryDocument, target SchemaTarget) (*SchemaReport, error)
    func ValidateSchemaBatch(documents []analysis.QueryDocument, target SchemaTarget) (*SchemaBatchReport, error)

SchemaRequest is {Document analysis.QueryDocument, Target SchemaTarget} with JSON document/target. SchemaBatchRequest is {Documents []analysis.QueryDocument, Target SchemaTarget} with documents/target. Decoders enforce the strict wrappers, normalize each document through the existing validation normalizer, and use DecodeSchemaTarget; target semantic preparation occurs in ValidateSchema/ValidateSchemaBatch. Invalid wrappers, selectors, Unicode, target configuration and empty batches return InputError, identified with validation.IsInputError. Query syntax failures remain invalid reports. M5 changes selector admission at the shared normalizer; it must not duplicate or bypass these strict decoders.

SchemaReport contains SchemaVersion int, Target SchemaTargetInfo, Analysis *analysis.Result, Status analysis.Status, Coverage validation.Coverage, Outcomes []SchemaReferenceOutcome and Diagnostics []analysis.Diagnostic, using the corresponding snake_case JSON members. SchemaBatchReport has SchemaVersion, Status and Reports []*SchemaReport. SchemaReferenceOutcome carries ReferenceID, Outcome, MatchesComplete, Matches []SchemaMatch, Evidence []SchemaEvidence, SupportingClasses, MissingClasses and IndeterminateClasses. SchemaMatch is Name/Binding/Outcome/Evidence. Keep matches_complete distinct from syntax/semantic/schema completeness and requiredness: proven partial matches can coexist with an incomplete selector, and admitted exact membership can have uncertain requiredness. Preserve the complete reviewed report, including empty collections and owned metadata, across adapters.

ValidateSchema normalizes the document, prepares the target once and calls validatePreparedSchema. That helper calls analysis.AnalyzeWithSourceUniverse, consumes finalized FieldExpansion IDs and ordinary Reference bindings, and never replays transfers. Source matches call prepared.project(name); derived matches use derived evidence without external membership; unavailable stays unavailable; indeterminate bindings stay indeterminate. Removal/non-field/not-applicable references do not create schema obligations. A missing constituent still adds an error even if another uncertain constituent makes the descriptive outcome indeterminate. Both field-list and schema finalization use finalizeValidation. Batch validates all documents, prepares its target once, preserves order and source IDs, discards the whole result on input/internal error, and reduces content status invalid before incomplete before valid.

Reviewed core fixtures are testdata/schemas/cases.json and requests.json, with tests in schema_validate_test.go, schema_request_test.go and schema_batch_test.go. Preserve TestSchemaFrozenCorpus, TestM3FinalizationByteParity, TestSchemaConcurrentPreparedAndOwnership and TestSchemaPublicUniverseLimitsDoNotPoisonExactReads as relevant M5 regressions. Root independently reports its M4 17/4/13 Go oracle passed at this pin; that is predecessor evidence, not a test run performed by this planning task.

The reviewed public core does not add a typed path discriminator. schema_validate.go passes ref.NormalizedName for source reads and expansion member names for wildcard matches to project(string). A nested access, quoted literal-dot identifier and dataset-qualified alias must retain their distinct typed frontend interpretation; agreement of normalized strings cannot justify a source match. Publish sound base/free reads and specific incomplete analysis/binding evidence where the intended member cannot be represented, so validation retains indeterminate rather than consulting a guessed dotted name. Alias qualification never becomes an external root property merely by concatenation. This is the approved incompleteness boundary, not a gap requiring a new schema engine or an unapproved public path API.

Adapter reconciliation now uses only immutable git show at accepted M4 Task 5 c91ca8464e916ae0b3792679a195a76e936c9385 and Task 6 7b6580fff46917045230668141d23f17489891df, plus .superpowers/sdd/milestone-4-implementation-plan/task-5-report.md, task-5-review.md, task-6-report.md and task-6-review.md. Their reported checks are predecessor evidence, not checks rerun here. The earlier core-only absence of schema adapters at 5e3d4ec is superseded by these accepted additions; existing field-list adapters and ABI remain intact.

cmd/schema_validation_cli.go supplies runSchemaValidationCLI(args []string, stdin io.Reader, stdout, stderr io.Writer) int, readSchemaCLITarget(options cliOptions) (validation.SchemaTarget, error) and computeSchemaCLIResult(options cliOptions, stdin io.Reader) ([]byte, analysis.Status, error). Its validate-schema command requires exactly one local --schema or --ocsf-catalog and one positional/--query, --file, --stdin or --batch input. JSON Schema options are --schema-base-uri and --schema-resources; OCSF options are --ocsf-version, --ocsf-class, --ocsf-category and repeatable --ocsf-profile/--ocsf-extension. Do not mix target families; distinguish document --profile from the separate --ocsf-profile selection, which can coexist. Numeric class/category strings become int64 UIDs; nonnumeric strings remain keys. Target/resource paths cannot be '-' and --config is rejected. Query file/stdin bytes remain verbatim, with path/<stdin> source IDs unless explicitly overridden; batch '-' reads stdin and rejects any global document selector/source-ID presence, including explicit empty flags. Local target JSON goes through DecodeSchemaTarget; document arrays use the strict batch input path and canonical validation. Preserve scalar-duplicate/Unicode/input/output gates and content exits 0/1/3 versus input error 2.

pkg/api/schema_validation.go supplies handleValidateSchema, handleValidateSchemaBatch and readSchemaValidationBody(w http.ResponseWriter, r *http.Request) ([]byte, error); server.go registers POST /api/v1/query/validate-schema and /api/v1/query/validate-schema/batch exactly once. The reader validates mime.ParseMediaType for application/json and reads through http.MaxBytesReader at 8<<20 bytes before canonical DecodeSchemaRequest/DecodeSchemaBatchRequest and ValidateSchema/ValidateSchemaBatch. Exactly 8,388,608 bytes may be accepted; one more returns 'request body exceeds 8 MiB limit' even with unknown ContentLength. Existing analysis/field-list routes retain 1 MiB. Content statuses all return 200; input/body/media errors are 400 and unexpected errors use the existing 500 helper. Never replace these decoders with ordinary struct unmarshalling: unknown/duplicate/null wrapper members and malformed Unicode retain strict input errors while schema/catalog payloads retain their own vocabulary. SchemaValidationRequest/SchemaValidationBatchRequest/SchemaValidationTarget are documentation-only DTOs for pinned Swag; tools/update_validation_openapi.py reconciles strict request target/selection unions, object-or-boolean schemas and URI-keyed inline resources without changing canonical report selection. No REST file lookup or fetching is added.

Accepted native exports in pkg/bindings/bindings.go are spl_mapper_validate_schema and spl_mapper_validate_schema_batch, both (mapperID C.int, requestJSON *C.char) *C.SPLResult. They delegate through ownedMapperJSONResult to the canonical strict decoder and validator. The tracked header declares extern SPLResult* spl_mapper_validate_schema(int mapperID, char* requestJSON), and the corresponding batch name. Python ctypes binds both as [ctypes.c_int, ctypes.c_char_p] to POINTER(SPLResult). mapper.py provides validate_schema(self, query, target, *, language='spl', profile='splunkd', version='current', source_id='') -> dict and validate_schema_batch(self, documents, target) -> dict. Both reuse _validate_fields_request(self, native, request): admitted _operation, json.dumps(..., allow_nan=False), native error conversion and unconditional spl_result_free in finally. Batch selectors remain inside documents. Add SPL2 by extending canonical selector admission and parity; no new schema wrapper or Python semantic inference is needed.

The accepted Task 6 native manifest has 47 entries, SHA256 2d3c8d7bd3c5b0dcbce9f9d878f0662e22b5beef185e2decd36945f455c66a22. It retains all 37 M3 entries and adds pkg/validation/json_schema.go, json_schema_patterns.go, json_schema_resources.go, ocsf.go, ocsf_projection.go, schema_batch.go, schema_model.go, schema_request.go, schema_target.go and schema_validate.go. M5 extends this real closure with its handwritten/generated inputs; it never recreates the former 37-entry list. test_native_schema_validation.py is already in NATIVE_TESTS and SDIST_FIXED_FILES. copy_schema_fixtures explicitly copies and hashes eleven files beneath testdata/schemas: cases.json, requests.json, ocsf/edge-cases.json, and ocsf/1.6.0/{base.json.gz,windows.json.gz,provenance.json,README.md,SCHEMA-NOTICE,SCHEMA-LICENSE,COMPILER-NOTICE,COMPILER-LICENSE}. install_and_check copies them before native tests, sets absolute SPL_SCHEMA_FIXTURES for both native and surface suites, and records fixture_hashes.schema. clean_env removes the source SPL_SCHEMA_FIXTURES override alongside PYTHONPATH/PYTHONHOME/SPL_NATIVE_LIBRARY/SPL_EXPECTED_VERSION. Source-native tests use an explicit library override; installed tests use SPLMapper() and Python -I outside checkout, preserving module/library provenance and wheel-payload hash binding. Catalog test inputs remain separate from runtime wheel resources.

Accepted M4 Task 7 and scoped I1 fix at 92879052b0454db12aead045188ce4b15fa9d83a close the package/header/surface/docs reconciliation. Evidence is immutable code at that pin, task-7-report.md/task-7-review.md/task-7-rereview-1.md, and docs/evidence/milestone-4/{verification.json,source-identity.json,package.json,review-i1/verification.json}. tools/check_package.py's toolkit_header_contract compares the complete reconstructed header after checking unique ordered cgo boundaries. It normalizes only the exact Go 1.25 guarded string-helper block and MSVC complex branch to their actual Go 1.22 forms within their respective compiler regions, once each; valid whole-line #line metadata; and the two known empty-parameter prototypes spl_mapper_new and spl_toolkit_version. All other prefix/preamble/prologue/export bytes remain compared, including unexpected declarations/directives and altered typedefs. Preserve tools/tests/fixtures/cgo-go1.22.h and cgo-go1.25.h plus the original twelve and I1 eight mutation controls. Do not restore raw wheel-header equality or the initial comparator that omitted compiler regions. Wrapper and manifest-listed sdist source/header comparisons remain byte-exact, with actual source/header/wheel/library hashes retained separately.

ACCEPTANCE_FILES now includes test_schema_surfaces.py alongside test_documented_cli.py, test_surfaces.py, test_analysis_surfaces.py, test_validation_surfaces.py and cli_examples.json. The schema suite's open_mapper uses an explicit absolute SPL_NATIVE_LIBRARY when present for source testing, otherwise SPLMapper() for installed testing; retain this distinction for the new SPL2 surface suite. It consumes absolute SPL_SCHEMA_FIXTURES and optionally emits full SPL_SCHEMA_EVIDENCE. The package checker passes copied schema fixtures before both suites, unusable external HTTP/HTTPS/ALL_PROXY variants with loopback NO_PROXY exceptions during runtime checks, and records schema_surface_evidence. Permanent CLI docs/manifests include schema-local-resources, schema-unresolved-resource and schema-local-batch witnesses and updated help. Preserve those frozen examples and the documented local target/resource, no-retrieval, requiredness/uncertainty and atomic batch contracts while adding SPL2. docs/API.md, quickstart.md, cli.md, api-server.md, compatibility.md, architecture.md, both READMEs and python/examples/basic_usage.py already document M4; M5 extends their actual accepted content rather than restoring predecessor drafts.

Historical installed original-wheel and rebuilt-sdist paths each passed 213 native plus 81 surface tests with zero skips and eleven copied schema fixtures. Those executed suites used the earlier installed checker, SHA256 9dba4b8ffb782c78466a9d6f97b76f1adeb48f87a337421a8f2234b512125712; package.json stays unchanged at 71a35cc7daa3890394457c306940f4e933962f2f19a62b4f8a05129dfb9a7227. The corrected checker SHA256 6baf6db778421cbbf5aa57f2146de69425bab9148cb58182f5c6d013e73414cf passed 71 affected package/documentation tests and freshly verified the retained original wheel. The rebuilt archive had been removed by its historical TemporaryDirectory cleanup, so no fresh full rebuilt-archive verification or installed-suite rerun is claimed; retained genuine header bytes matched its recorded header hash and passed corrected complete-header comparison. This accepted evidence distinction is required in later handoff reporting, not a request to rerun predecessor builds. Root's /private/tmp/spl-toolkit-root-m4-final-verification.json separately records local acceptance at 9287905: 230 source Python 3.14 tests, seventeen full canonical/CLI/REST/native cases, four batch documents and thirteen raw HTTP rejections, with 290 tracked sources, 47 native sources and five artifacts stable. These are root/implementer results, not checks performed by this planning task or M5 acceptance.

Root accepted full M4 at 0bb169c9fb1f0212d1e7758be6ab35a1804bbbaf, its direct five-path docs/evidence-only successor to 9287905. Immutable comparison confirms no API/code delta. Root approved this plan and released production; no repeated broad M5 architecture review is required. All M5 task/broad reviews and reserved independent root acceptance windows remain mandatory. Preserve the strict-offline wheelhouse, cache/Go 1.22 external-linker and offline-Swag recipes below without predecessor reruns.


## Interfaces and Dependencies


Use the existing ANTLR Go runtime without a dependency upgrade. The private SPL2 parsed result owns its generated QueryContext, token stream, sourceIndex, diagnostics and original delimiter ownership. It is not inserted into parsedDocument.tree, which is an SPL context. Analyze dispatches by normalized language to the correct frontend and shares finalization.

The proposed minimal transfer extraction for Task 1 is private to pkg/analysis. A located operand carries a normalized semantic name, original Location, explicit resolution and a soundness flag established by its frontend. Context wrappers determine whether syntax is intact before invoking these helpers. No generated type crosses public APIs.

    type locatedOperand struct {
        Name string
        Location Location
        Resolution string
        Sound bool
    }

    func (s *semanticStage) readAt(operand locatedOperand, role string) string
    func (s *semanticStage) createAt(operand locatedOperand, role, operation string,
        inputs []string, conditional bool) string
    func (s *semanticStage) diagnosticAt(code, severity, category, message string,
        location Location, incomplete bool)
    func (s *semanticStage) refinedSelectorAt(operand locatedOperand,
        role, referenceID string) []string

These are draft extraction signatures, not claims that they exist at 47b560d. The accepted refinedSelector returns transfer-candidate names and separately records FieldExpansion completeness/proven matches through recordExpansion. Preserve that distinction exactly in refinedSelectorAt; returned names are not proof that expansion is complete. Reconcile the same signatures and sidecar behavior against fully accepted M4 before final approval. The stable requirement is one canonical admission/selector algorithm, supplied with located operands from either frontend.

Move substantive projection, rename, assignment, aggregation and lookup transfer bodies into the following draft shared private operations. SPL wrappers retain decoding/recovery and pass explicit dialect policies, such as known-internal retention and frontend-resolved aggregate output names. SPL2 wrappers enforce their syntax and duplicate-rename restrictions before calling the same operations. Do not create a generic plugin framework, public AST or registry inferred from parameter names. Reconcile these concrete extraction proposals against accepted M4 before final approval; Task 1's immutable review then fixes the downstream implementation boundary.

    type renameOperands struct { Source, Target locatedOperand }
    type aggregateOutput struct {
        Target locatedOperand
        InputReferenceIDs []string
        Conditional bool
    }
    type preparedSelection struct {
        Field trackedField
        InputReferenceIDs []string
        EmitProjectTransition bool
    }
    type lookupOutput struct {
        Target locatedOperand
        PreserveExisting bool
    }
    func (s *semanticStage) applyAssignment(target locatedOperand, inputs []string,
        conditional, removeNull bool)
    func (s *semanticStage) applyProjection(selectors []locatedOperand,
        mode string, retainKnownInternals bool)
    func (s *semanticStage) applyPreparedProjection(selected []preparedSelection,
        mode string)
    func (s *semanticStage) applyRename(pairs []renameOperands)
    func (s *semanticStage) applyAggregation(outputs []aggregateOutput,
        groups []locatedOperand, preserveInput bool)
    func (s *semanticStage) applyLookupOutputs(matchReferenceIDs []string,
        outputs []lookupOutput)

Projection mode is a private finite value include, exclude or table, never arbitrary user text. Ordinary applyProjection evaluates located selectors once, prepares their selected trackedField evidence and existing reference IDs, then delegates installation/project transitions to applyPreparedProjection for include/table. Exact/wildcard exclusion retains the existing removal path. A preparedSelection copies the complete trackedField, including its field name, source/derived origin, conditionality and origin reference IDs. Ordered records preserve accepted transition order even when repeated selectors choose the same final name. EmitProjectTransition is false for implicitly retained known internals, which must not acquire fabricated lexical reads or transitions; explicit selections and SELECT outputs carry the already-created reference IDs and true. No prepared record proves more than its binding evidence.

applyPreparedProjection never calls readAt, referenceAt, selector evaluation or source admission. It installs copied selected bindings and emits project transitions using supplied IDs, sharing the existing open-remainder/uncertainty and tombstone policy rather than deriving a second SQL policy. Ordinary include/table and SQL final restriction call this same installation operation. SQL evaluate/aggregate preparation stores selected bindings and existing lexical IDs; project passes that prepared selection with table restriction semantics and adds no read/output reference. Alias output evidence and source evidence remain distinct, and finalization remaps supplied IDs normally. Do not resolve a prepared selection again against an environment changed by HAVING/ORDER. This is a private transfer boundary, not a public AST.

applyRename preserves the accepted snapshot-source transfer, which is equivalent for admitted independent SPL2 pairs; it does not add sequential rename support. removeNull is the frontend's proved exact-null assignment policy, including the root-approved Task6 extension to intact finite supported structural always-null compositions; it is not a general evaluator and preserves all RHS reads/diagnostics. Aggregate inputs are evaluated in their proper original environment before output installation; Target already contains the frontend-resolved implicit name or explicit alias. Group operands remain real located reads, and preserveInput distinguishes stats from eventstats/streamstats.

Lookup frontend wrappers evaluate every local match read exactly once, before any output changes the environment, then pass copied matchReferenceIDs plus ordered lookupOutput records. PreserveExisting is per output, retaining each legacy OUTPUT/OUTPUTNEW block's policy and source order. applyLookupOutputs reuses the same prepared match IDs for all outputs, registers each output location once and applies the accepted per-output policy against the environment at that output's turn. For an existing destination with PreserveExisting, retain the accepted output reference without replacement/lookup transition; otherwise create with the accepted conditionality. An earlier output overlapping a match field cannot cause match reads to be repeated or switch their provenance. External column labels/dependencies remain separate frontend evidence. This preserves legacy mixed blocks without broadening admitted SPL2 syntax. All functions reuse environment and sourceRefinement rather than recreating their membership/conditionality algorithms.

The approved public selector shape is:

    type CapabilityOptions struct {
        Language string `json:"language"`
        Profile string `json:"profile"`
        Version string `json:"version"`
    }

    func CapabilitiesFor(options CapabilityOptions) (CapabilityManifest, error)

Capabilities() remains the no-error default SPL API. One private normalizeSelectors helper supplies defaults and allowed values for both capabilities and QueryDocument normalization; validation wraps failures as InputError. Strings must be valid UTF-8, including selectors. Add strict capability-options decoding for the owned C JSON request rather than invoking Analyze on a fake query.

Preserve existing Capability fields. Add an optional DocumentationSnapshot string to CapabilityManifest, populated only for the new SPL2 contract so the legacy SPL wire stays unchanged. Encode per-form/arity completeness and held limitations in deterministic Capability.Limitations strings keyed by durable manifest form IDs; do not pretend the command-level SemanticSupported flag grants every option complete effects. The durable manifest must separately record all inventory profile memberships and each form's grammar/semantics status. Caller-returned slices must be deep copies.

The approved additive lineage fields are:

    Phase string `json:"phase,omitempty"`
    ExecutionOrder *int `json:"execution_order,omitempty"`

For SQL-owned entries the finite Phase vocabulary is source, filter, group, evaluate, aggregate, having, order, project, limit, offset. ExecutionOrder is the zero-based index of that entry in the actual Result.Lineage array, including entries from other scopes. The fields are omitted on unchanged SPL entries; a pointer distinguishes the SQL index zero from omission. ScopeID supplies scope meaning, not adjacency. A nonaggregate SELECT has evaluate and final project phases; an aggregate SELECT has aggregate preparation and project phases, without creating a second stage. Plain field reads and alias output references are registered once in evaluate/preparation and reused by final restriction.


## Corpus contract and review rules


Create durable, repository-owned evidence under testdata/spl2. The fixture schema must distinguish provenance/obligation IDs from deduplicated query IDs. Preserve original base IDs when materializing their queries. Every E.L01.start input is already complete, including negatives. E/F/I obligation aliases may point to the same unique input without inflating meaningful counts. H01–H12 and EH01–EH05 get held records, never definite-negative floor credit. Snapshot expectations must be reviewed independently rather than dumped from the new implementation and blessed.

The 35 command families are search, from/select, eval, where, fields, table, rename, stats, eventstats, streamstats, lookup, sort, dedup, head, reverse, join, append, appendpipe, appendcols, union, if, spl1, bin, rex, spath, makeresults, loadjob, tstats, mstats, timechart, timewrap, makemv, mvexpand, mvcombine, fillnull. Keep all 53 inventory rows, including 50 native entries and the three other-profile entries decrypt/ocsf/route. FROM and SELECT share one family. Native deferred branch/into/thru do not become module errors because examples involve sinks.

The 25 language families cover starts, identifiers, literals, comments, search predicates, precedence, conditional predicates, arrays/objects, access paths, templates, lambdas, projection/rename lists, aggregation/window forms, lookup lists, FROM-first, SELECT-first, SQL/pipes, dataset literals, SQL joins, pipeline children, union, metrics/data models, extraction/options, embedded SPL and conditional paths. The expansion contains 149 E sections, 34 F spellings and 778 obligation labels; these are not a claim of 778 unique fixtures.

A durable case carries id, obligation_ids, source_keys, document, form_ids, syntax_complete, semantic_complete, status, expected/forbidden codes and explicit source/scoping/phase assertions. For precedence/lexical grammar cases also store a small normalized typed-tree shape, not only discovered fields. Core flow cases and surface parity cases carry complete expected canonical reports. Source excerpts must match exact byte slices independently; retain actual newlines and multibyte text.

One minimal expectation record used by Task 2 is:

    {
      "id": "E.L01.start.N1",
      "obligation_ids": ["E.L01.start.N1"],
      "source_keys": ["start"],
      "document": {"text":"failure index=app","language":"spl2"},
      "form_ids": ["L01.starts"],
      "syntax_complete": false,
      "semantic_complete": false,
      "status": "invalid",
      "expected_codes": ["SPL_SYNTAX_ERROR"]
    }

The expected code is the existing CodeSyntaxError value SPL_SYNTAX_ERROR, verified in accepted diagnostics.go. During initial frontend tasks test private parse diagnostics/tree coverage; public semantic expectations become active as their owning lowering task lands. Unimplemented core semantics must be explicitly incomplete in an intermediate build, never skipped or temporarily marked valid. Do not leave a permanent expected-failure/skip hiding mandatory scope.

The 34 positional spellings are abs, ceil, ceiling, floor, round, len, lower, upper, trim, ltrim, rtrim, substr, replace, coalesce, if, case, match, isnull, isnotnull, tonumber, tostring, mvcount, split, count, sum, avg, min, max, dc, distinct_count, values, list, first, last. Arity one: abs/ceil/ceiling/floor/len/lower/upper/isnull/isnotnull/mvcount; one or two: round/trim/ltrim/rtrim/tonumber/tostring; two or three: substr; exactly three: replace/if; exactly two: match/split; coalesce one or more; case positive even argument count. Aggregate count accepts zero or one; other approved aggregates one. Scalar min/max are a separate unmodeled variadic family, not malformed aggregate calls. Named colon calls remain incomplete; argument labels are not fields and value expressions retain reads.

Each task follows the same serial review gate. The implementer first writes a specific failing case and runs only its relevant focused suite, records why it fails, makes the smallest production change, runs that suite and relevant regressions, then self-reviews and commits exact owned paths. Root requests a fresh combined specification/code review against that immutable commit. Fix findings in committed follow-ups and request scoped re-review. Only then dispatch the next task. A separate broad milestone review follows Task 11; it does not replace Task 11's own task review. Root independently accepts the final milestone. No implementation agents are dispatched during draft planning.


## Task 1: Extract located shared transfers and one selector contract


Own pkg/analysis/analyze.go, references.go, source_fields.go, commands.go, functions.go, capabilities.go, model.go's CapabilityOptions and new pkg/analysis/transfers.go/options.go; focused new transfers_test.go/options_test.go and relevant existing tests. Own the narrow pkg/validation/request.go normalization change and its request tests. Preserve flow.go behavior; change it only if extraction requires a small location-independent helper. Produce shared located transfers, normalized selector handling, the default-compatible CapabilitiesFor entrypoint and unchanged SPL behavior; do not expose SPL2 analysis before its frontend exists.

1. Add a failing located-transfer equivalence test in package analysis using the same initial environment, source refinement and exact byte locations as an existing SPL command test. Cover source read, derived creation, preceding removal, closed projection, partial wildcard admission and known-internal retention. Task 1 owns the prepared-selection extraction and proof: prepare one source binding and one derived/conditional binding with existing reference IDs, call applyPreparedProjection, and assert reference count/content unchanged, copied source/derived/conditional origins preserved, project transitions use those IDs, and finalization remaps them. Include implicit known internals with no invented transition. Add full SPL Result/SourceAnalysis parity witnesses for `search user=* | lookup users user OUTPUT name OUTPUTNEW role`, `search user=* | eval role="local" | lookup users user OUTPUT name OUTPUTNEW role`, and `search user=* | lookup users user OUTPUT user OUTPUTNEW role` where the first output overlaps the match field. Require one lexical local-match read, immutable pre-output match origins, ordered per-output references/transitions and unchanged existing-destination behavior. Add selector tests for defaults, invalid UTF-8 and invalid profile/version. Before changing accepted behavior, pin full SPL Result and SourceAnalysis values for these and representative existing corpus cases.

       func TestSelectorsDefaultsAndInvalidProfile(t *testing.T) {
           got, err := normalizeSelectors(CapabilityOptions{})
           if err != nil || got.Language != "spl" || got.Profile != "splunkd" || got.Version != "current" {
               t.Fatalf("unexpected defaults: %#v %v", got, err)
           }
           if _, err := normalizeSelectors(CapabilityOptions{Profile: "edge"}); err == nil {
               t.Fatal("unknown selected profile must be an input error")
           }
       }

2. Run `go test -mod=readonly ./pkg/analysis -run 'TestSelectors|TestLocatedTransfer|TestPreparedProjection|TestLookupExtraction'`; the new entrypoints should fail to compile or assertions should fail before extraction. Move reference resolution/creation/admission bodies into shared located methods, leaving syntax decoding and soundness in the SPL wrappers. Separate ordinary projection's selection/read preparation from its shared prepared installation, and separate lookup's once-only local match preparation from ordered per-output transfer. Use explicit diagnostic severity/category/incompleteness; preserve the old wrapper's legacy code policy. Extract shared command transfer only once, including output naming as a frontend operand rather than source formatting in the kernel.

3. Keep normalizeDocument's source text validation and canonical defaults. Implement CapabilitiesFor here for available frontends using the same normalizeSelectors helper; zero options returns existing SPL values. An unavailable frontend returns an input error, never a default manifest. Validation calls CapabilitiesFor to validate its three selectors and copies the manifest's normalized values, without performing an analysis pass or duplicating the allowed-language table. When Task 6 wires SPL2, its registry starts with explicit incomplete form capabilities and Task 8 completes the manifest audit; no intermediate unmodeled effect is advertised as complete.

4. Run `go test -mod=readonly ./pkg/analysis ./pkg/validation`, including SourceUniverse tests and the unchanged SPL corpus. Verify no duplicated admission/flow implementation and no lost finalized expansion IDs. Confirm the reconciled private transfer signatures in the task evidence before downstream dispatch. Commit the owned extraction and request the combined immutable Task 1 review.


## Task 2: Add the dedicated lexer, expressions and corpus harness


Own new grammar/SPL2Lexer.g4, grammar/SPL2Parser.g4, generated parser/spl2 files, pkg/analysis/spl2_parse.go, spl2_syntax.go, spl2_parse_test.go, spl2_corpus_test.go, testdata/spl2/manifest.json, provenance.json and initial lexical/expression JSON cases. Own narrow tools/check_go.py and tools/tests/test_build_checks.py changes needed to classify parser/spl2 as generated, plus new tools/check_spl2_corpus.py and its tooling test. Do not modify legacy grammar. Produce parseSPL2Document(text string) with a dedicated generated query context, call-local source map/tokens/diagnostics and typed tree facts.

1. Implement the durable loader/auditor and add red tests for L01 starts, L02 quoted Unicode names, strings/raw strings, comments, null/Boolean tokens, arithmetic/Boolean precedence, object/array syntax, access paths, templates, inline lambdas and positional/named call delimiters. The audit checks unique IDs and meaningful-query aliases, provenance closure, every mandatory family/form disposition and no H/EH floor counting. N1 must remain exact:

       func TestSPL2StartNegativeSource(t *testing.T) {
           text := "failure index=app"
           parsed := parseSPL2Document(text)
           if parsed.source.text != text || len(parsed.diagnostics) == 0 {
               t.Fatal("invalid standalone start was prefixed, lost, or accepted")
           }
       }

2. Write separate lexical token classes for identifier/string/raw string/regex/template/embedded search contexts. Use grammar rules or ANTLR lexer modes/actions local to the instance for typed template contents; token ownership must distinguish arrays, named lists, subsearches and lambda blocks. Never scan original source with regex to infer commands/fields. Source spans come from sourceIndex. Exact slash character-class regex and quoted alternation are admitted; slash/internal-pipe remains H12. Do not apply one global case-insensitive option to SPL2.

3. Define explicit search versus expression precedence, context-specific RHS literals versus reads, colon named arguments and positional-before-named ordering. Preserve legitimate positional comparisons with =. Duplicate object keys are parsed then statically contract-invalid after supported decoding. Lambda local parameters/variables and object keys are non-field syntax; free expressions remain typed. Standalone module declaration alternatives are recognized for rejection rather than imported into the grammar's executable scope.

4. Generate with pinned ANTLR commands in the verification section, test two-pass byte equality of all new generated files, and make the code checker exclude nested generated packages by path prefix rather than accidentally excluding handwritten pkg/analysis. Run `go test -mod=readonly ./pkg/analysis -run 'TestSPL2(Start|Lex|Expression|Source|CorpusSyntax)'` and the focused tooling tests. Intermediate semantics remain incomplete. Commit/review Task 2.


## Task 3: Enforce ordinary pipeline forms


For Tasks 3–5, root approved the necessary lexer-token ownership correction without new syntax breadth. Preserve existing identifier/quoting behavior and typed mixed-context regressions; run both pinned generators and two-pass equality for affected outputs. Preserve stable source identities when activating durable cases; pending and held forms receive no completed credit.

Own grammar/SPL2Lexer.g4 for genuinely needed keyword/mode tokens, SPL2Parser.g4 and regenerated parser/spl2, pkg/analysis/spl2_syntax.go, focused spl2_pipeline_syntax_test.go and the C01/C03–C15 plus related L05–L14 durable cases, including narrow manifest/provenance activation. Consume Task 2 expression/source/token rules; produce typed contexts for ordinary commands without changing shared field transfer.

1. Add table-driven syntax tests for search, eval, where, fields, table, rename, stats, eventstats, streamstats, lookup, sort, dedup, head and reverse. Each option/list/call alternative receives the expansion's actual admitted/negative obligation; lexical-only arity errors remain separately classified. For example `FROM main | eval x=bytes y=x+1` is invalid missing comma, while the corrected comma form yields two ordered assignment contexts.

2. Add dedicated rules with required operands, permitted option order and comma delimiters. Search admits only index-first implicit entry; other omitted initial commands remain invalid at query entry. Supported unknown options do not disappear. Encode exact output aliases, grouping operands, selector include/exclude intent and lookup match versus output columns as distinct contexts.

3. Preserve held table wildcard quote variants, streamstats postaggregate layouts and head disputed option placements as recognized/incomplete alternatives, not malformed fallback. Distinguish definite duplicate/chain rename contract findings from independent pair syntax. Boolean eval is admitted. Statistical functions in a prohibited command context are contract errors rather than syntax-only failures.

4. Run `go test -mod=readonly ./pkg/analysis -run 'TestSPL2(Pipeline|CorpusSyntax)'`, the meaningful-count auditor and unchanged SPL parse/selector regressions. Check typed tree facts for options and list members, not just parser error count. Commit/review Task 3.


## Task 4: Parse both SQL clause hierarchies


Own grammar/SPL2Lexer.g4 for genuinely needed keyword/mode tokens, SPL2Parser.g4/generated parser/spl2, pkg/analysis/spl2_syntax.go for its required typed SQL contract/hold checks, focused pkg/analysis/spl2_sql_syntax_test.go and durable C02/L15–L19/Q01–Q08 cases, including narrow manifest/provenance activation. Produce typed FROM-first and SELECT-first clause contexts with original locations; semantic scheduling belongs Task 7.

1. Add red cases for required FROM, both clause orders, SELECT DISTINCT, aliases, WHERE, GROUP BY/GROUPBY, HAVING, ORDER BY/ORDERBY, LIMIT/OFFSET, dataset literals, INNER/LEFT JOIN and SQL followed by a pipeline. Assertions compare exact clause spans and typed nesting. `SELECT host WHERE bytes>0 FROM main` is invalid clause order; `SELECT host FROM main WHERE bytes>0` is grammatical.

2. Grammar permits only the documented optional clause positions and dependencies. SELECT is required with GROUP BY. SQL expressions use their own predicate rules, not search adjacency AND. Qualified alias fields remain distinct from dataset/resource names and literal dotted identifiers. EXISTS gets its documented SQL WHERE/HAVING restrictions, child alias correlation and prohibited child LIMIT/OFFSET contract checks. Unsupported RIGHT/FULL join forms cannot become opaque supported arguments.

3. Preserve hidden projection visibility as a semantic limitation, H07 per-term direction, H06 integer range and EH05 unparenthesized span assignment as held records. Keep fully evidenced span(field), span(field,unit) and field span=(unit). Do not infer arbitrary negative values invalid or missing parentheses invalid from a contradicted synopsis.

4. Run `go test -mod=readonly ./pkg/analysis -run 'TestSPL2(SQLSyntax|CorpusSyntax)'` and source-slice tests for SELECT-first Unicode/CRLF. At least 40 distinct SQL/mixed cases remain required after deduplication; do not count case-only repeats. Commit/review Task 4.


## Task 5: Complete the remaining dedicated grammar and child ownership


Own grammar/SPL2Lexer.g4 for genuinely needed keyword/mode tokens, SPL2Parser.g4/generated parser/spl2, pkg/analysis/spl2_scopes.go, spl2_scopes_test.go, spl2_extended_syntax_test.go and syntax cases C16–C35 excluding already-covered rows, including narrow manifest/provenance activation. Extend existing spl2_syntax.go only for required remaining-command operand/option/hold contracts; spl2_parse.go may receive a private descriptor attachment only if actual scope integration requires it, with no public or semantic API change. Cover join, append/appendpipe/appendcols, union, if/elseif/else, spl1, bin, rex, spath, makeresults, loadjob, tstats/mstats, timechart/timewrap, makemv/mvexpand/mvcombine/fillnull. Produce grammar-recognized but semantically incomplete typed forms and real child scope descriptors.

1. Add red option/operand tests from each remaining E row, including numeric versus logarithmic spans, timechart axis/split option positions, rex wrappers, spath quoted paths/output dependency, generating-command starts, metrics named arrays and removed SPL options. Confirm `FROM main | rex field=payload /(?<tag>\\w+)/` parses while the missing closing slash does not.

2. Build explicit forms even where effects are deferred. Use lexed ownership to distinguish independent append/join/union source searches from inherited appendpipe/if subpipes. Child scopes keep their owning original bracket/token spans. Metrics option labels, datamodel strings, regex capture names and object keys are not guessed ordinary reads. Embedded SPL bodies require syntax/semantics coverage beyond wrapper recognition; do not claim a valid wrapper proves the body.

3. Model supported-form versus native-deferred inventory membership separately. Native branch/into/thru remain known/deferred, not modules. Preserve EH01 unit gaps, EH02 timechart agg contradiction, EH03 time-selector comparison conflict and H12 regex conflict as incomplete. Named built-ins remain incomplete even when outer grammar is recognized.

4. Run `go test -mod=readonly ./pkg/analysis -run 'TestSPL2(ExtendedSyntax|Scopes|CorpusSyntax)'`; malformed supported forms must have located errors while disputed forms must not accidentally acquire invalid status. Commit/review Task 5.


## Task 6: Lower the complete ordinary pipeline subset


Own new pkg/analysis/spl2_lower.go, spl2_commands.go, spl2_expressions.go, spl2_functions.go, spl2_capabilities.go, their focused tests, canonical dispatch in analyze.go/capabilities.go, the minimal exact-spl2 admission in options.go/options_test.go through the existing normalized selector, and shared transfers.go only for missing common operations, plus source_fields.go for the approved resolver-present SPL2 per-candidate dotted-source identity guard. Own model.go role documentation and the narrow null_test filters in pkg/validation/validate.go/schema_validate.go with focused tests; update only the stale spl2-selector expectation/admission control in pkg/validation/request_test.go. Own the narrow additive corpus layer integration in pkg/analysis/spl2_corpus_test.go and focused canonical corpus tests, tools/check_spl2_corpus.py/tools/tests/test_spl2_corpus.py, testdata/spl2 manifest/provenance and F cases or canonical-expectation additions to existing deduplicated records. Complete plain FROM source establishment and source-aware search/where/eval/exact rename/fields/table/stats/eventstats/streamstats/lookup/sort/dedup/supported head/reverse effects. Consume shared transfers and typed grammar; produce canonical Result/SourceAnalysis through existing APIs. The initial SPL2 capability registry describes currently implemented versus deliberately incomplete forms; Task 8 finishes its audited inventory.

1. Add failing full-state tests with field lists, partial universes and bare analysis. Required witnesses include search RHS literals; WHERE RHS reads; eval left-to-right assignment; exact null removal; known internals preserved without creation; table closure; independent rename pairs; rejected duplicate source/target/chain/overlap; count() output count; optional function args; unknown function incompleteness; and coalesce(value) admitted only for SPL2.

       func TestSPL2SequentialAssignments(t *testing.T) {
           got, err := Analyze(QueryDocument{Text: "FROM main | eval x=bytes, y=x+1 | table y", Language: "spl2"})
           if err != nil || got.Status != Valid { t.Fatalf("%#v %v", got, err) }
           found := false
           for _, ref := range got.References {
               if ref.NormalizedName == "x" && ref.Role == "read" {
                   found = true
                   if ref.Binding != "derived" {
                       t.Fatalf("later assignment lost derived binding: %#v", ref)
                   }
               }
           }
           if !found { t.Fatal("missing located read of x in the second assignment") }
       }

   Keep a separate open-source control using `FROM main | eval x=bytes, y=x+1 | fields y`: the assignment read remains derived, but plain Analyze retains incomplete semantics for unresolved internal-field membership. Pair it with an explicit finite SourceUniverse and prior internal removal controls according to the accepted shared hook. Do not change fields retention or fabricate internals to make the Valid assignment example pass; table y deliberately isolates assignment/projection proof.

2. Wire normalized language dispatch to the dedicated frontend. Typed walks decode identifiers/strings according to SPL2, validate function arity/context/profile, collect located ordinary/free reads and frontend-resolved outputs, then invoke shared transfers. Move substantive duplicate field transfer into transfers.go instead of copying it into spl2_commands.go. Existing SPL wrappers retain their outputs and diagnostics. The Task6 private locatedOperand.UnresolvedSource flag may consistently withhold unproved literal-dot exact-source binding only when a custom refinement resolver is present; preserve plain/finite/partial flat-list exact reads, proved derived/unavailable evidence and old SPL zero-default behavior. Typed navigation remains separately incomplete after sound base reads, with no public path API or second validator. Apply the same resolver-present SPL2 identity policy per wildcard expansion candidate, including broad *: withhold ambiguous dotted source members/exhaustiveness without creating certain environment fields or validator refill, retain proved derived/flat members under existing partial expansion, and locate incompleteness on the original selector. Preserve no-resolver finite/partial flat and literal-star controls through existing matcher/admission/transfer.

3. Nullability is shared semantic evidence, not a lookup of function names presumed non-null. Preserve coalesce all-null, case fallthrough, nullable if branches, conversion and mvcount uncertainty. Supported explicitly modeled conditional outputs may leave the stage semantically complete with Conditional=true binding/transition; later reads and canonical validation remain indeterminate when presence is unproved. Preserve ordinary source/sequential-assignment and old SPL behavior. Intact finite supported compositions structurally proving always-null may remove the actual LHS while preserving every RHS read/diagnostic; no arbitrary evaluation, recovery-token proof or blanket name inference. Add paired conditional-only/consumer-validation, non-null fallback and all-null removal regressions. Direct intact supported positional isnull/isnotnull field operands (parentheses allowed) use located null_test references: shared bindings/origins remain, absent/tombstoned inspection emits no unavailable error, unknown presence is not installed, and ordinary consuming reads remain independent. Both canonical validators omit outcomes for this no-existence-obligation role. Preserve independent path/errors; arbitrary argument subtrees are not exempt. Add Analyze/flat/schema existing/absent/derived/conditional/tombstone and subsequent/nested-consuming controls. Only proved count() implicit name is count; operand-bearing count labels and complex unknown labels require AS or incomplete. sum(bytes)/dc(action) naming uses the approved frontend rule. Named-call labels remain non-field syntax but every value read survives; all named-call effects remain incomplete.

4. Preserve existing parser syntax expectations and attach separate explicit canonical-analysis expectations to the same deduplicated document records. Activate implemented F obligations immediately with exact original candidate/ID/source values and unchanged provenance V1 pin. Add only finite named assembly modes with exact fixed wrappers for eval RHS/aggregate operands when a candidate needs that context; existing command candidates retain pipeline-tail assembly. No arbitrary templates, candidate-byte rewriting or duplicate query credit. Auditor negative tests reject layer confusion, held/unknown promotion, changed assembly and final closure missing canonical expectations. Report grammar versus canonical evidence separately; positive function syntax need not imply complete whole-query semantics under nullability or another approved limitation. Task11 closes remaining corpus coverage rather than deferring Task6 arity verification.

5. Run `go test -mod=readonly ./pkg/analysis ./pkg/validation` with complete canonical reports and source-universe sidecars, plus concurrent mixed-dialect calls. Neither SPL2 support nor extraction can change default SPL JSON. Commit/review Task 6.


## Task 7: Execute SQL phases while preserving lexical evidence


Own pkg/analysis/spl2_sql.go, spl2_sql_test.go, model.go's two additive Lineage fields and focused finalization tests, plus the narrow SQL scheduler integration in spl2_lower.go using the accepted Task6 typed contexts. Consume Task 1's reviewed preparedSelection/applyPreparedProjection boundary, real SQL contexts, located transfers and expression/function policies. Produce supported no-join SQL semantics and phase metadata without changing original source or inventing stages. If actual integration proves a shared-boundary adjustment necessary, limit additional ownership to pkg/analysis/transfers.go and its focused tests, record the concrete need before dispatch and include affected SPL full-report parity in this task's scoped review. Do not implement a SQL-only projection installation engine.

1. Write failing tests for `SELECT host FROM main WHERE bytes>0`: lexical commands SELECT/FROM/WHERE, positions 2/0/1 within root, and actual lineage FROM/source, WHERE/filter, SELECT/evaluate, SELECT/project. The host projection read occurs once in evaluate, and project reuses its identity. Verify every Location slices the original clause. Every SQL lineage entry has Phase and ExecutionOrder equal to its actual index. Default SPL reports omit both fields.

2. Separate lexical registration from evaluation. Register each real clause once, in original text order. Plan per-scope logical clause-entry positions and execute FROM, WHERE, GROUP BY, SELECT aggregate or nonaggregate evaluate preparation, HAVING where documented, ORDER BY, SELECT final project, LIMIT, OFFSET, as present. For `SELECT lower(user) AS owner FROM main ORDER BY owner`, actual phases are source, SELECT/evaluate, order, SELECT/project. Register plain projection reads once during evaluation; final restriction must not duplicate lexical read obligations or fabricate reference spans. Preserve separately prepared aliases/selected fields without prematurely restricting source input. The evaluate phase does not approve nongrouped HAVING, hidden source fields, unselected aggregates or collision-dependent alias visibility.

3. Retain pre-group source environment for aggregate arguments; group/aggregate/alias environment for supported HAVING/ORDER; selected-output environment for final projection. For `SELECT sum(bytes) AS total FROM main GROUP BY host HAVING host="a"`, host grouping read is sound, hidden HAVING visibility is indeterminate/incomplete, and final projection remains SELECT-owned. Keeping pregroup state internally must not grant hidden field access. SELECT aliases never satisfy earlier WHERE source obligations. Unsupported grouping expressions, mixed non-grouped projection, alias collisions and unproved output labels remain explicit limitations or proven contract errors.

4. Test FROM-first/SELECT-first/pipeline equivalent field results while retaining different original locations. Call applyPreparedProjection for SELECT final restriction and assert it adds no references: its prepared source, alias and aggregate binding origins remain intact, with existing IDs reused in project transitions. Test SELECT-owned aggregate/project entries with coherent Before/After, source-ref ID remapping after all phases/children, and a next piped command seeing final projected output. Run `go test -mod=readonly ./pkg/analysis -run 'TestSPL2SQL|TestPreparedProjection|TestFinalize|TestSourceUniverse'` plus unchanged SPL corpus; any shared-boundary fix also runs Task 1 lookup/projection full-report parity controls. Commit/review Task 7.


## Task 8: Preserve recovery, deferred scopes, capabilities and validation truth


Own pkg/analysis/spl2_recovery.go, spl2_recovery_test.go, spl2_scopes.go, spl2_capabilities.go, capabilities.go/model.go optional snapshot field and tests; narrow integration in spl2_lower.go and spl2_sql.go for child scope/position scheduling and one finalization, plus spl2_parse.go/spl2_syntax.go only where typed recovery/classification requires it; durable manifest activation/coverage disposition closure; focused new pkg/validation/spl2_test.go and spl2_schema_test.go consuming the reviewed schema_request.go/schema_validate.go/schema_batch.go APIs. Preserve existing M4 frozen tests and add only focused integration regressions in their files when a demonstrated shared fix requires it. Do not implement schema membership or a second field matcher. Produce truthful end-to-end Go analysis and all canonical validation operations.

1. Add malformed-middle/adjacent-stage and nested ownership regressions with comments, strings, arrays and regex containing delimiters. Valid intact adjacent stages retain trusted reads; inserted tokens and damaged child ownership never become source references. Recovery is driven by original typed token boundaries and local parser contexts, never source pipe splitting. Unknown standalone command has incomplete syntax/effects; malformed supported command is invalid; known profile/module exclusions use their distinct codes.

2. Lower independent children from fresh source environments and inherited appendpipe/if children from copied pre-transfer parent environments. Keep unknown merge output incomplete, prevent child-created fields leaking as certain parent names and preserve unresolved qualified aliases. Dataset literals with known object keys establish local derived/conditional fields without external source requirements. Dynamic paths/templates/lambdas retain sound base/free reads with specific limitations.

3. Implement CapabilitiesFor using the approved options API, deep copies and one canonical selector normalization. Register the entire native/profile inventory and each admitted/held form from durable manifest data, never from _build_plan. Default Capabilities must compare byte-for-byte to its prior fixtures. SPL2 capability booleans and limitations must agree with actual corpus outcomes; no unconditional semantic support for unknown arguments. Strict input failures cannot return a plausible default manifest.

4. Run Go field-list validation with missing source, derived output, null removal, wildcard finite/partial refinement and invalid-plus-incomplete precedence. Use ValidateSchema(document, target) and ValidateSchemaBatch(documents, target) with the reviewed closed/open/conditional/nested/OCSF target fixtures; exercise DecodeSchemaRequest/DecodeSchemaBatchRequest before the same calls for strict wrapper parity. Preserve SchemaReferenceOutcome.MatchesComplete, complete SchemaMatch/Evidence/class lists and owned target metadata. If source path distinction is unrepresentable, assert indeterminate rather than guessed matching. Include crossed partners: a quoted literal dotted identifier against a nested-only schema, and typed navigation against a literal-dotted-only schema. Preserve independently proved base reads or absence. The accepted validator projects Binding=source through a normalized string, so a stage warning alone cannot prevent false conclusive matching; keep the unproved intended path binding indeterminate without adding a second validator or altering proven bindings. Mixed-language batches preserve order/source_id and prepare the target once; input failures discard the whole batch. Add this concrete closed-target control, then its absent-host, unknown-semantic and typed-path ambiguity partners:

       func TestSPL2SchemaClosedSource(t *testing.T) {
           target := SchemaTarget{Kind: "json_schema", Schema: json.RawMessage(
               `{"type":"object","properties":{"host":{"type":"string"}},"required":["host"],"additionalProperties":false}`)}
           report, err := ValidateSchema(analysis.QueryDocument{
               Text: "FROM main SELECT host", Language: "spl2", SourceID: "core-window",
           }, target)
           if err != nil || report.Status != analysis.Valid || !report.Coverage.SchemaComplete {
               t.Fatalf("closed exact source validation: %#v %v", report, err)
           }
       }

   Run `go test -mod=readonly ./pkg/analysis ./pkg/validation` and the durable coverage auditor. Commit/review Task 8, then reserve the root Go-only acceptance window below before Task 9 dispatch.


## Task 9: Expose explicit dialects through CLI and REST


Own cmd/cli.go, cmd/analysis.go and focused tests, cmd/main.go showHelp strings only, cmd/validation_cli.go/validation_cli_test.go, cmd/schema_validation_cli.go/schema_validation_cli_test.go, pkg/api/analysis.go, models.go, handlers.go, validation.go/validation_test.go, schema_validation.go/schema_validation_test.go and relevant focused tests. Own maintained docs/cli.md, docs/api-server.md and tests/acceptance/cli_examples.json. Own docs OpenAPI source/annotations plus tools/update_validation_openapi.py and its focused tests only where selector/manifest schema reconciliation requires it. Reuse tests/acceptance/test_documented_cli.py unchanged unless a demonstrated harness gap requires a narrow reviewed edit; it already accepts stdin. Consume canonical Go APIs; no adapter syntax guessing.

1. Add CLI red tests for analyze/capabilities/validate-fields and validate-schema with --language spl2, --profile splunkd and --compatibility-version current; exercise --source-id on document operations. Mixed batches use document-specific language; conflicting global batch selectors remain input errors. Preserve accepted schema/OCSF target options and repeatable selection arrays, local-file target reads, query/file/stdin bytes, explicit empty defaults, invalid selectors/Unicode, input order, stdout/stderr, output files and content exit codes 0/1/3 versus input error 2. Isolate selector negative witnesses using otherwise valid target/input data so an unrelated target error cannot make rejection tests vacuous.

2. Legacy map/discover/validate operations reject explicit spl2 with unsupported_dialect_for_operation and guidance to canonical analyze/structured validation or later rewrite. Do not change fixed-SPL Go mapper methods. Legacy REST structs must recognize language/profile/version rather than ignore selectors, validate them canonically and reject spl2 before legacy execution. Default legacy success/error shapes remain unchanged.

3. REST /api/v1/query/analyze and canonical validation routes pass normalized documents through unchanged. Capabilities accepts language/profile/version query selectors, rejects duplicate/unknown/conflicting selectors with 400, and calls CapabilitiesFor. Reuse schema_validation.go's readSchemaValidationBody, strict canonical decoders and registered single/batch handlers. Preserve analysis/field-list 1 MiB and schema 8 MiB limits, including exact-bound acceptance and one-byte-over rejection with unknown ContentLength. Content reports valid/invalid/incomplete all use 200; request errors 400; unexpected failures retain the accepted 500 policy. Keep the documentation-only schema DTOs out of runtime decoding and preserve request-union versus report-selection separation when adding selector documentation.

4. Add full JSON HTTP/CLI parity for representative complete, malformed, deferred, module/profile and SQL phase cases. Update docs/cli.md's canonical selector/help text and marked examples, each paired with its exact ID/argv/exit/stdout/stderr/output_files in cli_examples.json. Include explicit SPL2 analysis/capabilities, field validation via stdin and mixed-language batch stdin using the existing manifest stdin field; do not add a shell-pipe runner. Update existing help-case expectations when helpPayload changes. Keep the test_documentation_example_coverage one-to-one marker/manifest/command-block check, with other maintained docs linking canonical CLI usage. Update docs/api-server.md's capability query selectors, canonical document examples, legacy rejection, content/request status behavior and actual schema-route contract; exercise those exact request examples in HTTP parity tests. Regenerate OpenAPI with the offline recipe and run reconciliation idempotence/shape tests. Run `go test -mod=readonly ./cmd ./pkg/api`, the documented CLI command in the verification section and focused tooling docs tests. Commit/review Task 9.


## Task 10: Preserve native ownership and installed source closure


Own pkg/bindings/bindings.go and focused tests, generated python/spl_toolkit/libspl_toolkit.h, python/spl_toolkit/mapper.py, new python/tests/test_native_spl2.py and relevant Python unit/ABI tests, python/native-source-files.txt, tools/check_package.py, tools/tests/test_package.py, and python/build_support.py only if actual source/attribution closure needs it. Own tools/tests/fixtures/cgo-go1.22.h and tools/tests/fixtures/cgo-go1.25.h when the new C export requires current-ABI fixtures. Consume accepted M4 native/schema APIs and the new Go capability selector; produce native/Python parity without changing old ABI signatures.

1. Add owned JSON-options export `spl_mapper_capabilities_for(int mapper_id, char *options_json) -> SPLResult *` through ownedMapperJSONResult, strict options decoding and CapabilitiesFor. Keep spl_mapper_capabilities unchanged. Add Python keyword-only `capabilities(*, language="spl", profile="splunkd", version="current")`; no-argument output stays unchanged. Preserve _operation, handle admission, concurrent close safety and finally/free on success and error.

2. Existing analyze_query and validate_fields carry document selectors. Reuse accepted validate_schema(query, target, *, language='spl', profile='splunkd', version='current', source_id='') and validate_schema_batch(documents, target), their spl_mapper_validate_schema/spl_mapper_validate_schema_batch exports and unchanged _validate_fields_request helper. Extend full single/mixed-batch SPL2 parity through canonical normalization, including strict Unicode/NaN/malformed-wrapper failures and owned result/error/free behavior; do not implement local analysis or add duplicate schema methods. Add keyword-only selectors to legacy map_query/map_query_with_context/discover_query/get_input_fields and reject explicit spl2 before invoking legacy ABI with unsupported_dialect_for_operation plus canonical-operation guidance, preserving SPLMapperError. Use canonical options validation through the native selector API where necessary; do not infer a Python semantic registry. Verify default legacy calls unchanged.

3. Enumerate all new handwritten pkg/analysis files and all generated parser/spl2 Go inputs in the final accepted native-source manifest, retaining the accepted 47-entry closure. Verify all staged _native_src inputs and wrapper/header provenance under 9287905's complete-header comparison after exact known compiler deltas; preserve genuine compiler provenance and all normalization/rejection regressions, byte-exact sdist inputs and actual payload hashes. The new capability export changes the current ABI: refresh either real-header fixture as needed only from actual already-planned respective Go 1.22/1.25 build headers, never synthetic header editing. Record new fixture/build hashes and verify both complete headers against the current source header; preserving M4 provenance does not require stale M4 export bytes. Ensure installed wheels and sdist-built wheels compile with generated SPL2 code and need neither grammar generation nor network nor _build_plan at runtime. Retain proper attribution; the new grammar is purpose-written from documented syntax, not copied from an unlicensed grammar.

4. Extend check_package's final accepted required source/test/acceptance closure. Preserve test_native_schema_validation.py and test_schema_surfaces.py registration, full schema_surface_evidence and copy_schema_fixtures' eleven-file hash-checked closure. Copy durable testdata/spl2 outside checkout before native suites run; supply absolute copied SPL_SPL2_FIXTURES and SPL_SCHEMA_FIXTURES to native and acceptance environments. Require new suite collection, nonzero cases, zero skips, packaged library load and module origin, fixture hashes, loaded-library/wheel-payload identity and sdist source evidence. Preserve clean_env removal of PYTHONPATH/PYTHONHOME/SPL_NATIVE_LIBRARY/SPL_EXPECTED_VERSION/SPL_SCHEMA_FIXTURES, add source SPL_SPL2_FIXTURES removal before supplying its copied path, and retain Python -I installed isolation plus the accepted blocked-external-proxy/loopback runtime environment. Run source native tests, tooling closure tests and package checks using the final recipes; record installed evidence separately from source tests. Task 10 collects its new native SPL2 suite and existing acceptance suites; Task 11 creates/registers its owned test_spl2_surfaces.py and repeats final installed checks once that suite exists. Do not claim future-suite coverage at Task 10. Commit/review Task 10.


## Task 11: Close conformance, surface parity and permanent documentation


Own final testdata/spl2 files, pkg/analysis/spl2_corpus_test.go, focused validation corpus additions, tests/acceptance/test_spl2_surfaces.py, tools/check_acceptance.py/check_package.py required-suite registration, affected tooling tests, permanent docs/API.md/quickstart.md/architecture.md/compatibility.md, README.md, python/README.md and new docs/spl2.md. Own final consistency corrections to docs/cli.md, docs/api-server.md and tests/acceptance/cli_examples.json introduced in Task 9; reuse the existing documented-example/stdin harness. Own _build_plan/milestones/5-standalone-spl2/milestone-log.md and necessary docs/evidence/milestone-5/ permanent verification records only after execution release for handoff evidence. After final acceptance the controller finalizes and commits exact authored M5 design/plan/research/conformance handoff paths; copied prompt/PRD/CLAUDE, unrelated untracked files and ignored SDD records stay preserved.

1. Complete the durable per-form crosswalk and all required positive/minimal-negative/held expectations. Consume the reviewed separate syntax/canonical expectation layers and finite assembly modes from Task6; final closure requires canonical expectations for every required record and rejects layer confusion or held/unknown promotion while preserving original syntax assertions and provenance V1. Run the auditor and publish actual deduplicated counts: all 35 families/25 language families, every 53 inventory entry classified, at least 200 meaningful inputs/80 definite-negative-profile/40 SQL-mixed, and every approved option/arity obligation linked. Record uncovered extra forms honestly. No count padding by uppercase/value variants, no skips for mandatory supported forms, no accepted malformed core syntax through an opaque fallback.

2. Compare complete canonical JSON reports across Go, CLI, real HTTP, source native Python, installed wheel and wheel built from sdist. The source SPL2 surface suite must explicitly honor an absolute SPL_NATIVE_LIBRARY override like accepted test_schema_surfaces.py; installed runs use SPLMapper() with no source override. Only object key order may differ. Preserve document text/source_id, stages, scopes, references, locations, dependencies, diagnostics, phase metadata, ordered batches and target metadata. Cover field-list and reviewed M4 JSON Schema/OCSF validation with both dialects, including otherwise matching schema plus unknown semantics and invalid-plus-incomplete results. Mixed concurrent calls prove no dialect leakage.

3. Document standalone starts, explicit selection on every surface, default SPL compatibility, dedicated syntax families, complete positional core and all named/dynamic/held limitations. Explain lexical stage order versus position and actual lineage phases using the SELECT-first example, count() naming, independent rename restriction, null deletion and known-internal retention. Document excluded modules/profile codes and legacy-operation guidance. Explain direct null_test field inspection and its omitted existence outcome without suppressing ordinary consuming-read requirements. Explain that complete modeled effects can include conditional field presence, with no per-event value promise; document the finite structural all-null removal boundary. Document current as a build capability snapshot, never live Splunk certification. Reconcile maintained CLI/server references, exact help output and every marked CLI example with the accepted Task 9 behavior. Execute test_documented_cli.py's examples and documentation-coverage checks against the final binary, then require the installed package checker to copy and execute the updated cli_examples.json/docs closure outside checkout; do not substitute a docs-only string scan for actual example execution. Server examples are covered by exact-request HTTP surface assertions.

4. Run final qualified verification below once on the final executable inputs. Commit Task 11 owned paths, obtain its fresh combined task review and fix/re-review findings. Reserve root's final stable built/native/surface/Python 3.14 acceptance window before the separate broad review of the fully accepted M4 base through final M5. After fixes rerun affected scopes; do not duplicate unchanged full suites for docs-only closure. Start milestone-log.md with `## What's new in the app`, followed by concise user capabilities, decisions/deviations, exact accepted commits, actual corpus counts, package/native/parity proof and remaining platform/operator limits. Root independently accepts completion; do not merge or push.


## Root independent acceptance windows


Root owns /private/tmp/spl-toolkit-root-m5-cases.json, containing 12 independent design-derived analysis inputs. Its confirmed SHA256 is 751f75b936882c9c8931247fbcbd8191ceb778c732c0e52a297fec33998d61ab, superseding the initial fab19b8970243591edb8f43690eef78cc704f379420ebdf8db75b60132820c11 after root added BMP/non-BMP Unicode, opaque Unicode source_id and CRLF to the first case. Root's preparation ledger now also records /private/tmp/spl-toolkit-root-m5-validation-cases.json at 3e8aca6e85d027064aa4a6387632619d9043c9c62d22d057ebdafc19dd6a957d with six schema and two field cases, mixed four-document batches per target and planned repeated concurrent checks. Its Go driver, semantic checks and orchestration are prepared-not-executed artifacts. This planning task reads preparation metadata only, does not run those drivers or infer passing results, and never uses independent root inputs to manufacture production expected outputs. Root supplies current driver/input identity when each reserved window starts.

After Task 8 is committed, combined-reviewed and scoped fixes accepted, stop implementation dispatch before Task 9. Give root the exact committed SHA, relevant Go API names and the corpus/limitations report, then hold the checkout/HEAD stable while root runs its independent Go-only canonical acceptance. Do not modify source, fixtures, plan-driven implementation or HEAD in that window. No M5 CLI/native surface is required for this first window. Resume Task 9 only when root closes the window; root findings enter focused failing regressions, committed fixes and scoped re-review before continuing.

After adapter/package work and Task 11's immutable task review, reserve a second stable window with final CLI/server/native library/wheel/sdist identities for root's built/native/real-HTTP/installed-Python and Python 3.14 checks. Root owns the actual 3.14 environment/driver choice; record its verified interpreter/library/package provenance and results rather than guessing a path or treating the provisioned 3.12 development venv as 3.14 proof. Keep HEAD and shared build/package outputs unchanged while root uses them. Address concrete findings through affected checks and reviewed committed fixes, then proceed to broad milestone review and independent final root acceptance. These coordination windows do not waive any final M4 or production-release gate.


## Concrete verification commands


All commands below are execution steps under root release, from /Users/jacobdelgado/repos/spl-toolkit/.worktrees/remaining-milestones. Do not rerun accepted predecessor suites merely to start; execute task-appropriate RED/GREEN and required new-scope checks. Use the provisioned task-local tools; if absent, report/reestablish the same pinned tools rather than silently skipping a supported-floor check.

    export GOCACHE=/private/tmp/spl-toolkit-go-cache
    export GOMODCACHE=/private/tmp/spl-toolkit-remaining-gomodcache
    export GOPROXY=off
    export GOTOOLCHAIN=local
    export PIP_CACHE_DIR=/private/tmp/spl-toolkit-pip-cache
    shasum -a 256 /private/tmp/spl-toolkit-remaining-tools/antlr-4.13.2-complete.jar

The expected JAR SHA256 is eae2dfa119a64327444672aff63e9ec35a20180dc5b8090b7a6ab85125df4d76. Use Java 21 and this generator/runtime pairing without a dependency upgrade. Generate only SPL2 outputs; unchanged legacy generation evidence is reused.

    java -jar /private/tmp/spl-toolkit-remaining-tools/antlr-4.13.2-complete.jar -Dlanguage=Go -package spl2 -visitor -listener -Xexact-output-dir -o parser/spl2 grammar/SPL2Lexer.g4
    java -jar /private/tmp/spl-toolkit-remaining-tools/antlr-4.13.2-complete.jar -Dlanguage=Go -package spl2 -visitor -listener -Xexact-output-dir -lib parser/spl2 -o parser/spl2 grammar/SPL2Parser.g4

Hash all generated files, repeat these two commands and require equality. Generator metadata, token/interp files and generated source are committed build inputs. The script auditor counts durable candidates and provenance only; it is not a query parser.

    /private/tmp/spl-toolkit-remaining-venv/bin/python tools/check_spl2_corpus.py
    /private/tmp/spl-toolkit-floor-tools/gomodcache/golang.org/toolchain@v0.0.1-go1.22.12.darwin-arm64/bin/go test -mod=readonly -ldflags=-linkmode=external ./pkg/analysis ./pkg/validation ./pkg/mapper ./pkg/bindings ./cmd ./pkg/api
    make build-all
    /private/tmp/spl-toolkit-remaining-venv/bin/python -m pytest tools/tests -q
    /private/tmp/spl-toolkit-remaining-venv/bin/python tools/check_go.py
    /private/tmp/spl-toolkit-remaining-venv/bin/python tools/check_docs.py

Execute maintained CLI examples, help expectations, stdin cases and marker/manifest parity after Task 9 and against final executable inputs. This existing harness needs only the actual CLI and docs root; no harness expansion is required for stdin.

    SPL_CLI=/Users/jacobdelgado/repos/spl-toolkit/.worktrees/remaining-milestones/build/spl-toolkit SPL_DOCS_ROOT=/Users/jacobdelgado/repos/spl-toolkit/.worktrees/remaining-milestones /private/tmp/spl-toolkit-remaining-venv/bin/python -m pytest tests/acceptance/test_documented_cli.py -q

tools/check_go.py owns the full race suite; do not repeat an unchanged full go test -race separately. Go1.22 external linking is intentional: the local default linker had an established dyld LC_UUID failure. This distinct floor check is not replaced by the newer toolchain's tests. Run shared build outputs only when no other production task owns them.

For OpenAPI, use the provisioned local module file proxy and disabled online checksum lookup for the pinned go run tool. GOPROXY=off is insufficient for that pinned command even when the module cache is present.

    GOCACHE=/private/tmp/spl-toolkit-go-cache GOMODCACHE=/private/tmp/spl-toolkit-remaining-gomodcache GOPROXY=file:///private/tmp/spl-toolkit-remaining-gomodcache/cache/download GOSUMDB=off GOTOOLCHAIN=local make generate-docs PYTHON=/private/tmp/spl-toolkit-remaining-venv/bin/python
    /private/tmp/spl-toolkit-remaining-venv/bin/python -m pytest tools/tests/test_update_validation_openapi.py -q

The following environment is source-native proof only. Accepted Task 6 requires absolute SPL_SCHEMA_FIXTURES pointing to the eleven-file schema corpus/catalog closure above. Preserve it and add SPL_SPL2_FIXTURES; installed checks instead set copied outside-checkout paths and load SPLMapper() without a source-library override. The accepted schema surface runner explicitly honors the source override; existing installed-only suites remain exercised through the package checker. Only an actual later broad-review delta requires reconciliation.

    PYTHONPATH=python SPL_SPL2_FIXTURES=/Users/jacobdelgado/repos/spl-toolkit/.worktrees/remaining-milestones/testdata/spl2 SPL_SCHEMA_FIXTURES=/Users/jacobdelgado/repos/spl-toolkit/.worktrees/remaining-milestones/testdata/schemas SPL_EXPECTED_VERSION=$(cat VERSION) SPL_NATIVE_LIBRARY=/Users/jacobdelgado/repos/spl-toolkit/.worktrees/remaining-milestones/build/libspl_toolkit.dylib /private/tmp/spl-toolkit-remaining-venv/bin/python -m pytest python/tests -q
    export PIP_NO_INDEX=1
    export PIP_FIND_LINKS=/private/tmp/spl-toolkit-offline-wheelhouse
    export PIP_CONFIG_FILE=/dev/null
    export PIP_CACHE_DIR=/private/tmp/spl-toolkit-pip-cache
    PATH=/private/tmp/spl-toolkit-floor-tools/gomodcache/golang.org/toolchain@v0.0.1-go1.22.12.darwin-arm64/bin:$PATH make python-build PYTHON=/private/tmp/spl-toolkit-remaining-venv/bin/python PIP=/private/tmp/spl-toolkit-remaining-venv/bin/pip
    /private/tmp/spl-toolkit-remaining-venv/bin/python tools/check_package.py --sdist dist/spl_toolkit-$(cat VERSION).tar.gz --wheel-dir dist --evidence /private/tmp/spl-toolkit-m5-package-evidence.json

Both package build and checker inherit the strict-offline PIP variables above and the existing GOCACHE/GOMODCACHE, GOPROXY=off and GOTOOLCHAIN=local. The floor Go PATH applies to make python-build; keep the accepted local checker recipe and do not replace the provisioned wheelhouse with network provisioning. Installed proof must load SPLMapper() with the packaged library outside checkout and collect test_spl2_surfaces plus the native SPL2 suite. Capture exact executable source identity, wheel/sdist/library/CLI hashes, tool versions, test counts and zero-skip results. Clean source-built wheel checks must prove every new Go/parser input is included. Do not run the existing clean-build checker with fresh empty offline module caches and misclassify cache absence as a product failure; use the accepted package/reproducibility recipes and report unrun release jobs honestly.


## Validation and Acceptance


A human can run `build/spl-toolkit analyze --language spl2 --query 'SELECT host FROM main WHERE bytes>0' --format json` and see the normalized spl2 document, lexical SELECT/FROM/WHERE stages, logical positions 2/0/1, located source reads and actual phase lineage. The same document through native Python and HTTP yields equal full report values. A field-list target containing host but not bytes must still flag bytes; SELECT projection must not hide that earlier requirement.

`FROM main | eval x=null | where x=1` has a located unavailable read. An independently renamed input retains its original requirement, and duplicate/chain/overlap rename is invalid. `FROM main | eval n=round(precision:2,num:bytes)` retains bytes as a read but remains incomplete. The held `FROM main | where user IS NOT string` is incomplete rather than a manufactured syntax error. `failure index=app` remains invalid at the initial-search boundary without prefix rewriting.

All complete core forms are exercised with plain analysis, finite fields and the reviewed partial schema universe. Valid static analysis requires complete syntax/effects and no error, not merely successful parsing. Unknown functions/commands, unresolved paths/aliases and conditional merges cannot be promoted by an unrelated schema match. Module/profile contract rejection remains distinguishable from malformed syntax and from request selection errors.

The same reviewed corpus passes native/CLI/HTTP/single/batch/installed-package paths; every new adapter stays thin and uses canonical copied reports. Local macOS evidence does not establish unrun Linux/Windows release-matrix or live Splunk execution. Final acceptance requires root's exact-head review and remaining-gate statement.


## Idempotence and Recovery


Generation is deterministic, tests use task-local caches and immutable fixture inputs, and no migration or user-data mutation is required. On failure, keep the failing case, fix only owned files, rerun affected checks and commit the tested fix before re-review. Do not reset the shared tree or clean another task's build outputs. Installed package tests must fail loudly for missing source/fixture/suite closure rather than skip.

After every task, record its tested immutable commit, commands/outcomes and review evidence in Progress and Outcomes. Inspect git diff --check, staged paths and staged diff before committing exact ownership. If a docs-only correction follows a build, record unchanged executable-input hashes and run the relevant document check rather than claiming the old suite ran against a new binary.


## Artifacts and Notes


No code in this plan is a runtime implementation or an executed test result. Code snippets define contracts and concrete red-test intent. Durable fixtures and provenance must be authored in testdata/spl2 during their owning released task; do not import or copy expected outputs from _build_plan at runtime.

Draft self-review maps architecture/flow to Tasks 1/6/7, all grammar families to Tasks 2–5, incomplete/profile/module/recovery and canonical validation to Task 8, CLI/HTTP to Task 9, C/Python/package closure to Task 10, and full corpus/docs/acceptance to Task 11. Reviewed M4 core report/path reconciliation is complete at 5e3d4ec, CLI/REST at c91ca84 and native/Python/source machinery at 7b6580f; Task 7 complete-header/surface/docs/package evidence is reconciled at 9287905. Full M4 is accepted at 0bb169c9fb1f0212d1e7758be6ab35a1804bbbaf with no code/interface delta; root released serial M5 production. Root's Go-only 12+6+2 and final built/native/Python 3.14 windows are reserved without implying executed acceptance.

Revision 2026-09-08: Initial bounded draft against immutable e401402 and accepted partial-universe hook 47b560d, with newly accepted M4 Task 2 schema types/preparer at bacc217. Incorporated controller-approved capability selector, SQL phase field shapes and evaluate preparation semantics. Corrected the proposed refined-selector signature to preserve its existing transfer-name return plus separate proven-expansion sidecar, made the sequence test nonvacuous, and kept path/alias distinctions incomplete where string-name schema evidence cannot establish intent. No predecessor checks, production work, implementation dispatch or commits occurred. This revision is not final plan approval or execution release.

Revision 2026-09-08, bounded architecture/root review batch: Resolved the two P2 findings in draft-architecture-review.md by defining private preparedSelection/applyPreparedProjection with existing IDs and copied binding provenance, assigning extraction/no-new-read proof to Task 1 and consumption to Task 7; and replacing lookup's command-wide Boolean with ordered per-output PreserveExisting plus once-prepared match IDs, with full SPL mixed-output/existing-destination/overlap parity. Changed Task 6's Valid assignment witness to table y and retained a separate fields/internal uncertainty control. Assigned maintained docs/cli.md, docs/api-server.md and cli_examples.json to Tasks 9/11 and added real documented-example/help/stdin execution using the existing harness. The original review and approved source artifacts are unchanged. These are draft corrections only; no production/tests/fixtures/builds/commits were performed, and final M4/plan/release gates remain open.

Revision 2026-09-08, reviewed M4 core preflight: Pinned actual SchemaRequest/SchemaBatchRequest/SchemaReport/SchemaBatchReport and DecodeSchemaTarget/DecodeSchemaRequest/DecodeSchemaBatchRequest/ValidateSchema/ValidateSchemaBatch at 5e3d4ec3a54b6878830b83a8ff08405011a507d9. Recorded unchanged analysis hook, string-name path limitation, shared validation finalizer, strict input/batch semantics and exact core test paths. Confirmed schema adapters/package closure are absent at this pin and left their actual names for full accepted handoff. Reserved root's 12-input Go-only window after Task 8 review and final built/native/surface/Python 3.14 window, using root-confirmed superseding oracle hash. Only this plan changed; approved sources/reviews remain immutable. No production/test/build/fixture/commit or HEAD mutation was performed.

Revision 2026-09-08, accepted adapter/native preflight: Read only immutable M4 Task 5 c91ca8464e916ae0b3792679a195a76e936c9385 and Task 6 7b6580fff46917045230668141d23f17489891df plus their accepted reports/reviews. Recorded exact CLI/REST/native/Python interfaces, strict 8 MiB schema reader, eleven-file copied fixture environment, 47-source manifest identity and package evidence machinery; assigned actual Task 9/10 paths and isolated negative witnesses. Updated root coordination from twelve narrative inputs to the prepared-not-executed 12+6+2 ledger. Final Task 7 header/checker compatibility, actual installed/surface closure, full M4 handoff and production release remain pending. No active Task 7 product file, production/test/build/fixture/commit or implementation dispatch was touched; approved architecture and source/review artifacts are unchanged.

Revision 2026-09-08, accepted Task 7/I1 preflight: Reconciled only immutable 92879052b0454db12aead045188ce4b15fa9d83a and accepted task/scoped-review/permanent evidence. Recorded complete-header comparison after exact known Go deltas, actual schema surface/source-library behavior and docs examples, strict-offline wheelhouse variables for both build and installed checking, and historical installed versus corrected-header verification boundaries. Root local acceptance metadata is attributed separately. M5 architecture, Tasks 1–8 and root 12+6+2 inputs/windows are preserved. Full M4 broad-review/controller closure SHA, any actual later delta, final plan acceptance and production release remain open. Only this plan changed; no production/tests/builds/downloads/fixtures/commits or implementation dispatch occurred.

Revision 2026-09-08, final approval and production release: Root accepted full M4 at 0bb169c9fb1f0212d1e7758be6ab35a1804bbbaf. Immutable diff against accepted runtime/checker 9287905 contains exactly five M4 docs/evidence paths and no API/code delta. Root approved this final plan without another broad design gate and released Task 1 serial SDD. Task 10 owns refreshing genuine Go 1.22/1.25 header fixtures from its actual planned toolchain builds as the capability export changes the ABI, retaining every normalization/rejection control and recording new hashes. No M4 baseline suite was rerun; no M5 production result is claimed at release.

Revision 2026-09-08, Task2-review boundary: Root approved explicit SPL2Lexer.g4 keyword/mode ownership for existing Tasks3–5 forms and narrow manifest/provenance activation alongside durable cases. Both pinned generation commands/deterministic equality and existing identifier/quoting/mixed-context boundaries remain required. This routine omission repair adds no syntax breadth or public API. Applied only after Task2 review verdict to preserve that review input; Task3 still awaits Task2 fixes and scoped acceptance.

Revision 2026-09-08, Task6 release: Task5 accepted at2d7a758 after combined/scoped review. Incorporated root-approved additive parser/canonical corpus layers, finite fixed assembly and immediate implemented F activation with narrow harness/auditor/provenance ownership. Recorded both crossed typed-path schema partners and the existing source-binding projection hazard. No public API or architecture change, repeated design gate or predecessor verification rerun. Task8 root Go window and final stable native/surface window remain unchanged.

Revision 2026-09-08, Task6 accepted at e77a3a8c0e753289f405801adb17d64b8f3331b2 after one scoped tooling fix. All ordinary semantic/shared path/null-test and layered corpus work is independently reviewed; current171canonical expectations remain distinct from1683syntax documents. Task7 owns the necessary spl2_lower.go dispatch integration because accepted analyzeSPL2 now controls typed top-level lexical registration/evaluation. SQL implementation consumes accepted private expression/prepared-selection boundaries; no architecture change or extra public API. Root Task8 and final immutable acceptance windows remain reserved.

Revision 2026-09-08, Task7 accepted at d2b3af94b3d946a696409b5b5cf5391aec7c8170 after one scoped correction preserving scalar-wrapper contracts around aggregate descendants. Shared private callWithExpression keeps the one finite call policy with operand preparation once; valid compound SQL outputs remain incomplete. Task8 consumes reviewed SQL phase/source/ref interfaces and owns necessary existing lowerer/SQL/private parser integration, retaining all no-child reports. Its durable syntax/activation coverage closure is distinct from Task11 full canonical-layer closure. Root Go-only12+6+2 window follows Task8 review; no adapters before rootrelease.

Revision 2026-09-08, Task8 root API wording clarification: analysis Analyze/CapabilitiesFor preserve their accepted plain error contract. Invalid selectors/Unicode require a nonnil input-rejection error and no plausible result/fallback; analysis has no exported InputError/IsInputError and none is added. Typed InputError/IsInputError assertions apply only to existing validation wrapper APIs. Preserve canonical error text and validation wrapping without brittle message-based classification. Earlier generic InputError wording describes input rejection, not a new analysis type.

Revision 2026-09-08, Task8 raw-parser/canonical-recovery evidence boundary: Root approved Task8 corpus-only recovery_classification {kind:unknown_command|native_deferred,command,start,end} with narrow auditor/unit/Go-corpus ownership. Preserve raw private syntax invalid/error expectations and original candidate/ID/source projection pins. Exceptional canonical layer is incomplete, both coverage flags false, floor_credit=false with zero meaningful/negative credit. Require exact nonempty original UTF-8 command-token slice, real canonical owning stage, authoritative inventory/disposition agreement and located unsupported diagnostic. Unknown excludes known dedicated/deferred/wrong-profile/module constructs; native_deferred only durable deferred-grammar/effects rows. Reject valid/complete promotion, profile/module demotion, synthetic/out-of-bounds/wrong-stage ranges and arbitrary whole-query overrides. Adjacent supported-malformed and malformed independent/inherited children retain definite errors. Auditor validates metadata/credit structure; Go corpus and typed recovery prove actual status/codes/stages without a second query parser or public result metadata. Record actual count deltas and preserve provenance.

Revision 2026-09-08, Task8 implementation independently approved at23e3c0eb02adf0bf79b9bcdc35123b87f4fc1354 with no blocking findings. One minor ignored-report wording correction distinguishes focusedanalysis from finalcombinedGo evidence; no productiondelta. RootGo-only currentreservation is16analysis+9schema+4fieldlist, legacy/mixedbatches and68interleavedanalysiscalls. Source/HEAD must remainstableuntilrootrelease; Task9 notstarted. Corpus1714syntax/1835active/177canonical/0pending retains288/139/45meaningfulfloors and17holds; fullcanonicalclosure remainsTask11.

Revision 2026-09-08, root acceptance scoped correction: restore concrete approved DocumentationSnapshot requirement after the controller brief mistakenly prohibited it. SPL2 value spl2-provenance-v1:sha256:3345cf5712b1bdbf467d1651784fdb8bccc596805038da0d54e7a123384e3a4e names the existing stable source/design/original-ID/inventory/hold projection, not downloaded pages or mutable corpus; static string/omitempty/default SPL omitted, meaning documented in code andTask11. Root also requires original typed literal-mode EOF closure errors survive unknown recovery. Exact five-file fix/review/freshrootacceptance remains pending; noTask9release.

Revision 2026-09-08, root core acceptance and Task9 release: root independently accepted 471e28746c288fea544d5cc4652d37ecad636a2c after both root fixes and clean scoped review. All 16 analysis, 9 schema, 4 field-list, 68 interleaved, legacy/full-report, selectors/copies and mixed-batch checks passed with unchanged source/input hashes. Root reports immutable results SHA02456051aa87dedd1267d6b47f4f9ac818b60add2b9765cd9152acdb65833407 and acceptance SHA672d3d58ea6f0edb034244dbb4f497169f2aa19303dd06e20e1c1dd9fed852e2; controller did not inspect private oracles/results. Initial failed attempt and three justified harness corrections remain root evidence. Task9 is released, then Tasks10/11 through normal serial gates; final root build/native/Python3.14 window still waits reviewed Task11. No reopened core baseline checks for orientation.

Revision after Task 9 acceptance: root accepted the narrow showHelp ownership and Task10-native/Task11-surface installed-suite sequencing. Task9 combined review found a legacy JSON field-folding mismatch; b9c98c7 uses EqualFold recognition for every raw selector occurrence before the existing canonical strict decoder, with fresh42-case and affected legacy API proof. Scoped rereview is clean. Historical pre-build binary and protected-content snapshot limitations remain qualified; actual scoped30 protected contents are stable, and28 prior protected contents plus2 attributed orchestration deltas establish earlier continuity. Task10 is released serially from b9c98c7 without a new root permission gate. Root final built/native/Python3.14 and broad acceptance remain pending.

Revision after Task 10 acceptance: b754b9b closes the shared Python legacy-refusal category with fresh4-case RED/28-control GREEN and refreshed native-wheel/sdist-wheel proof (each322 native+81 maintained acceptance, zero skips). Initial339 source Python/74 tooling/floor bindings and genuine65-source/18-export header closure remain qualified evidence on unchanged inputs. Review verified353 committed blobs/341 relevant inputs/968 protected contents for the scoped fix. Task11 is released serially; its canonical corpus closure flag, new SPL2 surface registration, final build/package checks, broad review and root final acceptance remain open. The inherited tar extraction warning and historical missing early RED-test hash are retained, not expanded into unrelated work.


2026-09-08 revision note: Recorded accepted Tasks 6–11, the independently frozen canonical/snapshot partition, later bounded recovery and field-presence decisions, both broad-review fixes, and final root acceptance at 40e6e171. This documentation-only closure preserves all original source/design/provenance documents and distinguishes the accepted runtime candidate from its later authored-document commit. The changes make the final implementation and remaining successor gates recoverable from this living plan.
