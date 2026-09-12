# Safe Rewrite and Mapping 2.0 Implementation Plan — ROOT APPROVED EXECUTION


> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development only after root accepts the reconciled final plan and releases production. Execute serially with a fresh implementer for each task, an independent specification/code review of the immutable task commit, fixes and scoped re-review, then a separate broad milestone review and independent root acceptance. Root approved private plan 86ce9d9c in full and released its repository application and serial implementation on 2026-09-08. M5 is finally accepted at documentation head 237e62ac, preserving runtime-tested 40e6e171/source commit 7571f1cb. Execute Tasks 1–5, reserve root's Go-only acceptance checkpoint, then continue the approved adapter/final-review sequence under the reserved root windows.

**Goal:** Let callers preview and apply exact, conditional source mappings with byte-preserving edits, an honest audit, canonical static verification and optional destination validation through Go, CLI, real native Python and REST.

**Architecture:** A new rewrite package prepares strict rules and consumes canonical analysis evidence; the existing dialect frontends own parsed identity, literal facts and safe rendering, while the existing field engine owns flow and lineage. Rewrite groups linked edits, rejects conflicts, builds a candidate and calls canonical analysis/validation again before returning any committed text. No second parser, transfer engine or public parser-independent AST is introduced.

**Tech Stack:** Go 1.22 minimum, existing ANTLR4 Go runtime and pinned generator, existing CLI/HTTP/C ABI, Python 3.11+ ctypes, local field-list/JSON Schema/OCSF targets. No new service or network dependency.

**Spec:** `_build_plan/milestones/6-safe-rewrite-mapping/design-spec.md`, SHA256 `4355ca99ec2807025dafda26cff66b6768287f004a6f9930a0482c066063adba`, reviewed and approved in full by root. This plan restates its execution requirements; the design travels with the final plan.

This is a living ExecPlan under `/Users/jacobdelgado/.codex/PLANS.md`. Its required Progress, Surprises & Discoveries, Decision Log and Outcomes & Retrospective sections remain current. PLANS permits checkboxes only in Progress, so task actions use numbered steps. The controller selected the existing repository filename, private preparation during broad M5 review and later serial SDD; no default-location, execution-choice or renewed brainstorming gate applies. This private integration binds the approved design and preflight proposals to M5's root-accepted fixed source/runtime at `40e6e171a1d215e5c67758fbf346cc5b78892ee2`, whose source commit is `7571f1cb578968caf92996b050ebfc3092f02605`. The scoped rereview is Approved and root runtime acceptance passed. The six source/test changes and twelve evidence additions since provisional `fcbb74ac60da312bc06275e4e3de29b570a3bf2e` are reconciled below. Root has also accepted and this plan has reconciled final authored-document closure at `237e62ac063755559207870b45744db6b2544a17`, the direct fourteen-path documentation/evidence-only child of 40e6e171. Root has now granted exact repository-plan acceptance/application and implementation release; the receipt is `/private/tmp/spl-toolkit-root-m6-production-release.json`, SHA256 `4d21108dd0eac7d1271a9e6e7a21ed54240fa3429e6bc40f97a0032eb874937b`.

Administrative execution note (2026-09-08): Earlier private-preparation restrictions and pending-release wording retained below describe the preserved preparation history; this explicit root release supersedes those administrative gates. Behavioral contracts, all eight task bodies, design/preflight decisions and the coordinate contract are unchanged. Authorized work is confined to M6 living-plan/SDD coordination and each dispatched task's owned files in the existing worktree. Preserve root aggregate, user original checkout and all M5 artifacts/history; no merge, push, publication, cleanup or M7 production is released.


## Purpose / Big Picture


After implementation, a caller can preview mapping src to user in `search src=alice | rename src AS owner | table owner`. The report shows a candidate containing `search user=alice | rename user AS owner | table owner`, exact original/candidate locations and the preserved owner alias. Preview keeps authoritative returned text equal to the original and committed false. Apply returns candidate text only after static binding/lineage verification succeeds. A failed destination schema check keeps original text, preserves candidate evidence and marks every proposed edit uncommitted.

The same versioned request through Go, a CLI file/stdin call, native Python or HTTP yields equal report values. A false rule is an explained skip; an unresolved rule or conflicting target remains incomplete rather than silently choosing an answer. Valid describes complete error-free static work, not runtime equivalence or actual event-field presence.


## Global Constraints


Only Milestone 6 is in scope. Preserve existing fixed-SPL mapper methods, configuration precedence and old native signatures. Do not change source files in place, perform whole-query conversion, beautify queries, rewrite wildcards/dynamic identifiers, infer raw-to-data-model/tstats changes, learn rules, execute searches or claim equivalent runtime results. Report output explicitly requested by a CLI caller is the only filesystem write performed by the operation.

Select language spl or spl2 explicitly through the existing canonical QueryDocument, with omitted language defaulting to spl, profile splunkd and version current. Normalize/reject options through the accepted canonical selector boundary. Maintain source text and source identity byte-for-byte, half-open UTF-8 byte offsets, Unicode columns, CRLF semantics, copied public values, deterministic IDs and array ordering. Unknown options and malformed Unicode remain request errors.

Conditions use only original, guaranteed source/flow facts. AND combines compatible guarantees, OR intersects guarantees, and query NOT establishes no positive equality. Later facts cannot authorize earlier edits; independent child queries never establish parent facts. Source-reference-present is a static query-reference fact, never event presence. The new rule language contains all/any, proven literal equals/string contains and source-reference-present, without regex, arbitrary code, condition-language NOT or caller-injected facts.

Canonical frontends own typed contexts, identity interpretation and render eligibility. Canonical transfer owns bindings, alias/implicit-output origins, inherited environments, unknown barriers, SQL logical phases and source universes. Rewrite must neither discover these by substring/regex/token guesses nor replay their transfer algorithms. The selected M6 rewrite bridge below is new implementation work; existing M5 helpers named below are real at the accepted fixed source. Final M5 closure at 237e62ac preserves those source identities; any later actual relevant change must be reconciled before it is used.

Atom/path rule identities remain distinct. On 2026-09-08 root confirmed that M6 adds only minimal fact/render/implicit-name hooks for already proven forms; do not expand canonical path binding just to force a mapping or schema match. Use the reconciled proven-form inventory below; every navigated or alias-qualified form that lacks canonical binding proof remains a located refusal.

Apply has one whole-request post-verification gate. Existing unrelated semantic incompleteness may remain without a requested target if every edited group is proven. Requested destination validation must produce a valid, complete report; invalid or incomplete validation prevents commit. An unknown/ambiguous group may remain unchanged alongside independently committed safe groups, yielding incomplete with committed true. This never permits an edit to depend on a skipped edit for safety.

This private preparation authorizes no repository/source/plan/ledger mutation, product tests/builds/generators, implementation dispatch or commit. Only artifacts under `/private/tmp/spl-toolkit-m6-final-m5-delta-preparation/` may be authored now. No push, merge, publication, tag, destructive cleanup or worktree deletion is authorized by the later implementation plan.


## Progress


- [x] (2026-09-08 UTC) Adopted the root-approved design, existing eight-task ExecPlan and all three approved core/CLI/native preflights; verified their hashes.
- [x] (2026-09-08 UTC) Verified provisional candidate `fcbb74ac60da312bc06275e4e3de29b570a3bf2e` descends from each preflight pin and inspected only relevant immutable deltas.
- [x] (2026-09-08 UTC) Integrated concrete facade/renderer/fact/phase/adapter/native/package proposals into this private plan, consuming M5's tstats extraction and final surface registrations.
- [x] (2026-09-08 UTC) Confirmed the original facade lacks candidate-coordinate conversion; root approved the single canonical LocateRewriteBytes bridge below and its exact validation/position contract.
- [x] (2026-09-08 UTC) Root accepted M5 fixed source/runtime at `40e6e171a1d215e5c67758fbf346cc5b78892ee2`; source commit `7571f1cb578968caf92996b050ebfc3092f02605` and scoped rereview Approved.
- [x] (2026-09-08 UTC) Reconciled exactly six source/test changes and twelve evidence additions since fcbb74ac, closing the len/split domain and module-category delta; preserved the root-read e518c17d private plan before revision.
- [x] (2026-09-08 UTC) Root finally accepted M5 authored-document closure at `237e62ac063755559207870b45744db6b2544a17`; reconciled its exact fourteen documentation/evidence paths and consumed the released public handoff, preserving accepted source/build/evidence qualifications.
- [x] (2026-09-08 UTC) Root approved private plan 86ce9d9c, authorized exact repository application with administrative release wording, and released serial production through the Task 5 root checkpoint.
- [x] (2026-09-08 UTC) Task 1: strict request/rule/target and report contracts; commits 237e62a..3a29058, independent combined review approved after scoped I1 fix/rereview.
- [x] Task 2: minimal canonical evidence and render eligibility for proven forms. Reviewed at 0b8dd620; I1–I5 and M1 closed. Root-authorized completeness/O1 producer follow-up independently approved at 57450416 with no open findings.
- [x] Task 3: three-valued conditional rule selection. Reviewed at 6c18c32a after scoped I1 known-derived ordinary-skip correction; zero open findings.
- [x] Task 4: linked edit groups, conflicts and byte-preserving candidate/audit. Independent combined review approved at 14a3a0dc with zero open findings; investigated limits retained.
- [ ] Task 5: post-verification, optional validation, apply rollback and ordered batch API. Fresh implementation released from reviewed Task4 head 14a3a0dc; root Go-only checkpoint follows its immutable review.
- [ ] Root receives a stable Go-only canonical acceptance window before adapters.
- [ ] Task 6: CLI/HTTP reports and maintained executable examples.
- [ ] Task 7: owned C/native Python API and installed source/fixture closure.
- [ ] Task 8: full durable corpus, documentation, surface parity and milestone handoff.
- [ ] Root receives final stable built/native/HTTP/Python 3.14 acceptance alongside independent broad review, then gives exact-head milestone acceptance.


## Surprises & Discoveries


The relevant adapter diff from `b9c98c76e1b0b1da4ae2f66c3094715fbd0b6235` to provisional M5 is empty. Strict document/target decoding, CLI option/body helpers, API examples and OpenAPI tooling keep the already approved preflight interfaces. No new shared JSON object decoder appeared: `internal/jsoninput` supplies Unicode validation; rewrite owns strict wrappers and reuses `validation.DecodeDocuments` for canonical documents.

M5 now extracts tstats DATAMODEL_NAME in `pkg/analysis/spl2_commands.go:561` (`metricsInputs`), as one `data_model` atom including its whole quoted token. The four prior positive corpus forms already contain that canonical reference; damaged N1/N2 forms do not. M6 needs private rendering ownership only, not another dependency helper or changes to ordinary Analyze JSON. `search_job` is a nonconsuming external-job reference, outside the seven mapping kinds. Deferred output references and named UNION/join dependencies do not establish output fields, source bindings or rewrite permission.

`prepareSQLSelection` now returns `([]preparedSelection, map[string]bool, bool)`. The final Boolean proves finite selected membership separately from conditional field presence. An origin crossing an incomplete wildcard expansion keeps that finite proof false. An intact noncolliding zero-argument count can have independent presence proof, without upgrading sibling semantics. `sqlProjectionCollisionName` may inspect a static alias-qualified suffix for collision purposes only; it cannot authorize a suffix/root source binding. M6 must preserve these distinctions when tracking implicit ownership and postverification.

Original typed recovery records and original delimiter checks became stricter in `spl2_parse.go`, `spl2_recovery.go` and `spl2_scopes.go`. Some intact original operands survive local parser damage; all original document diagnostics remain. M6's original-syntax-error gate still refuses edits for the document. It must not generalize the predecessor's narrow local retry into token reconstruction or a new parser.

Installed package acceptance now includes `test_spl2_surfaces.py`, the helper `spl2_transport.py`, hash-checked full Go reports and copied maintained docs. The helper is an input, not a collected test module. `tools/check_acceptance.py` requires named native/acceptance suites instead of relying only on old count floors. Native manifest membership remains 65 files and native exports/header inputs did not change after their preflight; `_legacy_selectors` now preserves `unsupported_dialect_for_operation`. These are inspected predecessor facts, not new executed acceptance.


## Decision Log


2026-09-07, root: Approved guaranteed scoped three-valued facts, simultaneous source-identity atomic edit groups, no priority winner, condition authorization at every linked member, implicit-name linkage, observed static collision checks, stable removal of safety dependencies on skipped edits, whole-request postverification, strict optional destination validation, exact audit and additive single/batch surfaces. These product choices remain settled.

2026-09-08, root: Approved the complete written design and all concrete core, CLI/REST and native/package preflight proposals, including the opaque per-call facade/value types, typed form inventory, SQL/finalizer correspondence and compact additive rewrite capability shape. Render is pure and session-bound; Verify checks document identity and rendering provenance before proof. Capability Role is a documented render context, not merely read/create and not a generated parser type.

2026-09-08, root: Approved exactly LocateRewriteBytes(text, ranges), RewriteByteRange and private sourceIndex.byteLocation after reviewing proposal 663a20a3. Require strict all-or-error UTF-8/exact byte boundaries, original canonical Position values, ordered results and an empty nonnil result slice. Keep helper/value/private method in rewrite_evidence.go, preserving source.go behavior and ownership. This closes only the bounded coordinate interface; final accepted-M5 reconciliation and plan/repository/implementation release remain required.

2026-09-08, root: Preserve the typed atom/path distinction and refuse individual forms that lack canonical binding/render proof. Add minimal canonical facts/render/implicit-name hooks for already proven forms; do not expand navigation, join or schema semantics for rewrite coverage. M5 owns the atomic tstats dependency; M6 adds private render ownership only. Static search_job and deferred output intentions do not add a rule kind or permission.

2026-09-08, root: Both new rewrite HTTP routes reuse the schema route's 8 MiB body helper, including exact-limit/+1 and real catalog checks; legacy 1 MiB routes remain unchanged. Native header changes require both whole Go1.22/Go1.25 generated fixtures from actual builds, never synthetic declaration insertion.

2026-09-08, root: Keep eight serial task implementations and immutable task reviews. Reserve independent Go-only acceptance after Task 5, then final built/native/HTTP/Python 3.14 acceptance after Task 8 alongside broad review. Root owns its independent expectations and aggregate progress; implementers neither read nor execute those private oracles. Root already answers product questions under the user's authorization.

2026-09-08, root: Authorized this private read-only delta preparation while M5 broad review runs. Provisional candidate `fcbb74ac60da312bc06275e4e3de29b570a3bf2e` is not accepted final M5 and may be superseded. Repository plan edits and production remain gated on explicit root acceptance and release.

2026-09-08, root: Accepted M5 fixed source/runtime at 40e6e171 after the two broad findings were fixed and scoped rereview Approved. Authorized only private reconciliation from fcbb74ac: six source/test files and twelve evidence additions. M5's lead owns authored-document closure; this authorization does not release M6 repository-plan edits or implementation. The approved design, preflights and LocateRewriteBytes contract remain settled.

2026-09-08, root: Finally accepted M5 documentation head 237e62ac, the exact fourteen-path documentation/evidence-only child of runtime-tested 40e6e171. Released the public successor handoff and authorized final private closure reconciliation. Source interfaces and settled architecture remain unchanged; root will separately approve/apply the exact M6 plan and release production.

2026-09-08, coordinator: The intervening changes require concrete hook/registration updates, not a new architecture. Candidate validation still uses existing public batch validators once over ordered candidate documents; private prepared schema types stay private. Shared-state writers remain serial and protected decisions stay at the Astra/XHigh capability floor.


Root ruling 2026-09-08, Task1 I1: Keep CandidateValidation.FieldList *validation.Report complete and unchanged in Go. Only rewrite CandidateValidation JSON field_list.target projects original {kind, identity, version}; all three keys remain present, including empty metadata strings, and fields/optional_fields keys are absent. Preserve every other canonical report field/value/analysis/outcome/coverage/diagnostic. No new metadata/default/hash/validation logic. Canonical validation.Report serialization and JSON Schema/OCSF branch remain unchanged. Use a small rewrite-only projection without mutating the original report; retain array/union guards. Test a populated required+optional catalog with nonvacuous report evidence, exact JSON equivalence after excluding ONLY those two target keys, unchanged Go original, standalone canonical arrays still present and unchanged schema serialization.

Task2 ownership ruling: Permit pkg/analysis/dependencies.go only to register private render-owner information inside the existing qualifiedCatalog(ctx, from), adjacent to its existing data_model/dataset reference emissions. Accepted/current source SHA256 122c652ba117a02904bdd37496114a940f300b6f1930365a9428e0612970e620; qualifiedCatalog at lines 84–114 retains the original typed context, datamodel prefix, quote form and bounded overlapping components that referenceAt alone does not receive. Preserve extraction, dependencies, diagnostics, normalization, public reference locations and ordinary Analyze results exactly. This supplies the already approved composite-owner bridge; no new forms, reconstruction, duplicate extraction or SPL2 tstats behavior is authorized.

## Outcomes & Retrospective


This phase produced a private integrated plan, bounded delta receipts and immutable source/hash records only. It did not run product tests, build artifacts, change repository files or dispatch implementation. No M6 behavior is implemented or verified. Root approved the exact LocateRewriteBytes contract below and accepted M5's fixed source/runtime at 40e6e171. Its six source/test changes and twelve evidence additions are reconciled without changing the approved interfaces or task scope. Final M5 authored-document closure at 237e62ac is now accepted and reconciled. Only root's exact plan/application/production release remains required.



2026-09-08, root/coordinator execution update: Task3 real-query integration exposed a canonical LiteralComplete gap at accepted Task2 head0b8dd620. Root confirmed that fully supported, fully observed complete OR branches with no common guarantee, NOT without a positive guarantee and an absent earlier restriction yield empty GuaranteedValues with LiteralComplete=true: absence of a proven query guarantee, never event/field absence. Unknown scalar/pattern/owner/binding/scope evidence remains unknown, separately guaranteed matching values survive other incomplete evidence, and missing map entries never become complete by default. Task3 stopped at a safe PARTIAL checkpoint with its three untracked files and18 source/report/evidence artifacts frozen (receipt397fc83b). Original eligible Task2 writer exclusively owns one bounded canonical follow-up for this gap and the queued O1 metric Boolean reproduction/fix-if-confirmed; then scoped qualified review and Task3 resume against the accepted producer identity. No consumer coercion, API/architecture expansion, parallel writers or new approval gate. Later Task5/Task8 root windows remain unchanged.

## Context and Orientation


Work eventually executes in `/Users/jacobdelgado/repos/spl-toolkit/.worktrees/remaining-milestones` on `codex/remaining-milestones`. The user-owned original checkout is protected. Other milestone agents share the worktree: never reset, clean, stash, delete or rebuild outputs another agent owns. Root supplies accepted immutable commits and reserves stable acceptance windows.

M5's final accepted documentation head is `237e62ac063755559207870b45744db6b2544a17`, whose direct parent is the accepted fixed-source/runtime pin `40e6e171a1d215e5c67758fbf346cc5b78892ee2`, subject `docs(evidence): record scoped SPL2 broad-review fixes`. Its direct parent/source commit is `7571f1cb578968caf92996b050ebfc3092f02605`, subject `fix(spl2): correct function result domains and module category`, whose direct parent is the prior provisional `fcbb74ac60da312bc06275e4e3de29b570a3bf2e`. Earlier approved source pins are core `471e28746c288fea544d5cc4652d37ecad636a2c`, CLI/REST `b9c98c76e1b0b1da4ae2f66c3094715fbd0b6235`, native/package `641e616f8db68b4dad7e8eed3eecb64e4bf12f0c`, and Task10 fix `b754b9bfa80eb990cb83ee05d3b4588c2c2e54c3`; verified ancestry carries through to 40e6e171. The final closure identity and released public handoff are reconciled below. Root must still approve/apply and release this exact plan before execution.

`pkg/analysis/analyze.go` is the canonical analysis entry for both selected dialects. SPL2 has separate parsed source/token ownership in `spl2_parse.go`, typed lowering in `spl2_lower.go` and command/expression files, and a shared field-transfer kernel in `transfers.go` and `flow.go`. The SPL2 semantic stage embeds the common semantic stage but its legacy parser pointer remains nil. Existing located operands carry name/location/resolution/soundness/unresolved-source evidence; they do not contain the new rewrite rendering/fact/binding-lifetime contracts. Do not call legacy parser paths for SPL2.

A binding says whether a reference selects a source field, an explicit derived name, an unavailable name or unknown evidence. A source epoch is the kernel's lifetime for one source environment, not QueryDocument.SourceID. A directly selected binding key is separate from transitive origin ancestry: aliases can retain source origins while being explicit user-owned names. Lineage describes assignments/projections/aggregates and their input/output relationships. `sourceIndex` in `source.go` owns byte offsets and Unicode/CRLF positions. `finalizeReferences` in `references.go` source-sorts IDs and remaps origins/transitions. `spl2FinalizeStages` in `spl2_scopes.go:214` remaps stage IDs and actual execution order after nested scheduling.

SQL execution in `spl2_sql.go:24` follows source, filter, group, evaluate/aggregate, having, order, project, limit and offset. SELECT preparation at `prepareSQLSelection:217` and final project share one stage and prepared references. Its third return value proves finite membership, not unconditional presence. `runChildren` in `spl2_scopes.go:168` executes children at their actual owner phase; only inherited children receive copied environments and outputs never merge into parents. Original delimiters must exist. A recovered reference can be useful diagnostic evidence without authorizing a rewrite of a syntax-damaged original.

Validation request APIs are real and unchanged from the accepted preflights:

    func DecodeDocuments(data []byte) ([]analysis.QueryDocument, error)
    func DecodeFieldCatalog(data []byte) (FieldCatalog, error)
    func DecodeSchemaTarget(data []byte) (SchemaTarget, error)
    func Validate(document analysis.QueryDocument, catalog FieldCatalog) (*Report, error)
    func ValidateBatch(documents []analysis.QueryDocument, catalog FieldCatalog) (*BatchReport, error)
    func ValidateSchema(document analysis.QueryDocument, target SchemaTarget) (*SchemaReport, error)
    func ValidateSchemaBatch(documents []analysis.QueryDocument, target SchemaTarget) (*SchemaBatchReport, error)
    func IsInputError(err error) bool

`validation.DecodeDocuments` performs exact duplicate/member/null/Unicode checks and canonical dialect normalization. Decode a singleton raw document through its established one-element-array pattern and a batch through its array directly. `internal/jsoninput` exports Unicode validation only; rewrite still implements strict wrappers/rules and may narrowly extract shared decoding only when tested to preserve predecessor behavior. SchemaTarget is the existing strict json_schema/ocsf tagged input, with inline resources or local versioned catalog/selection. Prepared indexes and the validator finalizer stay private. The canonical public batch API prepares one target for the whole ordered candidate slice.

The maintained frontend surfaces are `cmd/cli.go`, `cmd/validation_cli.go`, `cmd/schema_validation_cli.go`, `pkg/api/server.go`, `pkg/api/schema_validation.go`, `pkg/bindings/bindings.go` and `python/spl_toolkit/mapper.py`. Generated OpenAPI artifacts are `docs/swagger.json`, `docs/swagger.yaml`, `docs/docs.go`; the existing `tools/update_validation_openapi.py` reconciles all three atomically. `python/native-source-files.txt` is an exact native source allowlist. `tools/check_package.py` verifies wheel/sdist sources, copied fixtures/docs, real imports/libraries and mandatory suites outside the checkout.


## Accepted M5 reconciliation and remaining release gate


The five original preflight items are integrated below: typed access/facade, proven render forms, SQL/finalization correspondence, additive capability metadata, and strict adapter/native/package integration. No accepted architecture choice is reopened. The initial source inspection at fcbb74ac is retained in `source-identities.json`; the bounded fcbb74ac-to-40e6e171 reconciliation is in `final-source-delta-40e6e171.json`, SHA256 `3b8e09cb2ad85963b88f7988275f518e0e03037235551848e40096ab9c970ed5`. The latter records all eighteen paths with exact SHA256/Git blob identities and the three read report hashes. Exactly six source/test files changed and twelve evidence files were added. Adapter, public validation, common-kernel, source-index, native/package, fixtures and maintained product documentation retain the inspected identities; no public API shape or scope expanded. Unchanged input continuity is not a repeated runtime check.

Final authored-document closure is accepted at `237e62ac063755559207870b45744db6b2544a17`. Its direct parent and exact fourteen documentation/evidence-only paths were verified against the supplied root audit; every immutable path hash matches that audit. `final-closure-delta-237e62ac.json`, SHA256 `3970e28c11746fe4dd1db56e4265c86a5f366631aaead40bd4f4426c714fb315`, records those identities and read-only command exits. The root audit SHA256 `73f10808f7d73f99091652e8d0f1ddb58a9f6167237f09cdb0253f2ac49bc9a9` retains verified continuity for 398 nonowned files, ten original authored documents and nine held artifacts. No source, fixture, adapter or package interface changed. The public successor handoff SHA256 `1f3878483ca6ac601044e28a47f9fda856d2126058689a67005448de17f06fc9` was read in full and preserves all already integrated M5 boundaries and verification qualifications. No unchanged source research or tests were repeated. Root's exact repository-plan acceptance/application and implementation release are the only remaining gates; no implementation worker starts before that step.

The accepted correction in `pkg/analysis/spl2_functions.go`, SHA256 `d7576fde248da9e134dfbf3a79297abcf9783d77fff660378592630359ff57b7`, assigns len a numeric return domain and split an unknown scalar domain. Canonical analysis can now prove the numeric composition abs(len("abc")) when its existing argument/presence guards hold. Split is still modeled, but its multivalue output supplies no scalar-string proof to lower/len or another scalar consumer. Invalid/null/source-dependent arguments and unknown, named or wrong-arity functions retain existing uncertainty; valid modeled analysis may still have conditional field presence. M6 consumes these canonical results without inventing new literal facts or rewrite permissions. The corrected module suffix/declaration sites in `pkg/analysis/spl2_syntax.go`, SHA256 `8438f5174b3a930f80425d71c316bb1c6cdacab304b450af042766a5c4a7d46c`, use unsupported_syntax while preserving SPL_UNSUPPORTED_MODULE, messages, severity, locations and false syntax coverage. The four changed test files cover nested domains, conservative controls, validation presence and both module locations. All six source/test hashes match the supplied fix report.

The supplied broad fix report SHA256 `347300bde92cd949fa9ee74d8ec4692bef84925b2178e51c195e539a190623b5` and scoped rereview SHA256 `c906a7a32b5dcdd045d688f293e9fd15c2d036d09e56c5fa80efe6cdcb1296bc` retain the intentional REDs and qualified worker verification. All thirteen compact fixture inputs remain unchanged; the full 1714-report corpus changes only the five approved B08–B12 module-category leaves. The root receipt `/private/tmp/spl-toolkit-root-m5-40e6e171-final-acceptance-receipt.json`, SHA256 `567ba8d92ec4b63f54c02b1e2c6ae49d86d094013b3ae69934ddf51337249014`, records fixed runtime acceptance at 40e6e171, including 339 Python 3.14.1 native tests and full built-surface checks. Worker checks ran on precommit bytes identical to 7571f1cb; root checks ran at 40e6e171 while held build artifacts retain precommit VCS metadata and the accepted source identity. Do not relabel either record. These are retained predecessor results, not M6 execution or new tests by this reconciliation. Root's receipt supersedes the older reports' then-pending runtime gate; authored-document closure, open when that receipt was written, is now accepted at 237e62ac. Preserve remaining platform/runtime limits. The canonical tstats dependency addition remains owned by M5; no ordinary Analyze exception is now planned.


## Interfaces and file responsibilities


Create pkg/rewrite rather than replacing pkg/mapper. Proposed public operations and decoders are:

    func DecodeRequest(data []byte) (Request, error)
    func DecodeBatchRequest(data []byte) (BatchRequest, error)
    func DecodeRuleSet(data []byte) (RuleSet, error)
    func Rewrite(request Request) (*Result, error)
    func RewriteBatch(request BatchRequest) (*BatchResult, error)
    func IsInputError(err error) bool

Request is the strict schema_version/mode/document/rules/validation_target wrapper. BatchRequest replaces document with documents and shares mode/rules/target. RuleSet is the CLI file wrapper containing schema_version and rules only; mode/document/target are supplied by CLI flags. Require schema_version integer 1; missing mode normalizes to preview; an explicitly null/malformed mode is rejected. Empty rules are a valid no-op; empty documents are a batch input error. Validate Go-constructed values through the same preparation path as decoded requests.

An Identity contains either Name *string or Path []string, never both, serialized as name or path. Path is field-only and retains exact segments. Rule contains ID, Kind, Source, Target and optional When. Condition is a strict all/any/literal/source_reference_present union. Literal contains kind, identity, operator equals/contains and a scalar value retained without lossy float64 decoding; string, number, bool and null identities remain different. Numeric equality uses supported exact literal values, and values outside the frontend's proven scalar representation make only affected conditions unknown rather than coercing data. The strict decoder distinguishes a present null scalar from a missing value member.

ValidationTarget composes existing types: field_list uses `{kind: field_list, catalog: <existing catalog input>}` and delegates catalog decoding to DecodeFieldCatalog; json_schema/ocsf use the existing SchemaTarget wire shape unchanged and delegate to DecodeSchemaTarget. This wrapper is target selection, not a second schema format or a copy of schema logic. CandidateValidation retains exactly one complete original validation.Report or SchemaReport and its kind in Go. Only rewrite CandidateValidation JSON projects field_list.target to original {kind, identity, version}, keeping all three keys present even for empty metadata strings and omitting only fields and optional_fields. Preserve every other canonical report field/value, analysis, outcome, coverage and diagnostic, without mutating the original Go report or adding metadata/default/hash/validation logic. Canonical validation.Report serialization and the JSON Schema/OCSF branch remain unchanged. Do not flatten away report evidence or repeat the large catalog in each rewrite report.

Result contains SchemaVersion, normalized Document, Mode, Status, Coverage, OriginalText, CandidateText, Text, Committed, Changes, RuleEvaluations, OriginalAnalysis, CandidateAnalysis and optional CandidateValidation, with the approved snake_case wire keys. BatchResult has SchemaVersion, Status and ordered Reports. Coverage retains syntax/semantic/rewrite completeness, optional validation coverage and ordered reasons without confusing rewrite ambiguity with a parser failure. Changes carry outcome, reason, group/rule IDs, original/candidate reference IDs, optional real original/candidate locations, old/new text and separate candidate_applied/committed flags. Missing references use rule-level evaluations with absent location, never a fake zero-length span. Task 1 implements these snake_case fields and copied collection shapes without changing their approved meaning.

Keep model.go/request.go/target.go for these public contracts and strict preparation, conditions.go for pure three-valued condition composition, groups.go for linked selection/conflict closure, edits.go for byte edits and coordinate translation, verify.go for canonical correspondence and validation policy, rewrite.go/batch.go for orchestration, and diagnostics.go for deterministic audit codes/order. Do not create an interface per file or a generic plugin framework. Focused *_test.go files live beside each behavior. Canonical hooks live in the canonical analysis/frontend files and new evidence/render files named in the bridge section below, but do not add empty placeholder files or expose generated contexts.


## Canonical bridge, facts and proof ownership


Task 2 creates `pkg/analysis/rewrite_evidence.go`, `rewrite_facts.go`, `rewrite_render_spl.go`, `rewrite_render_spl2.go`, `rewrite_correspondence.go` and focused tests, with small hooks in `analyze.go`, `references.go`, `transfers.go`, `flow.go`, SPL `commands.go`, narrowly scoped private typed-owner registration in `dependencies.go:qualifiedCatalog`, and SPL2 `spl2_lower.go`, `spl2_commands.go`, `spl2_expressions.go`, `spl2_sql.go`, `spl2_scopes.go`. The coordinate helper/value and private sourceIndex.byteLocation method live together in rewrite_evidence.go; existing source.go behavior and ownership remain unchanged. Capability types/tables live with existing `model.go`, `capabilities.go` and `spl2_capabilities.go`. No new parser generation, independent AST or tstats extraction helper is needed. Analysis cannot import rewrite; rewrite imports analysis.

Implement the approved narrow source-level facade below. These Go helper values are not a replacement public JSON AST. Session/rendering internals are unexported and owned by the call. All exposed collections are copied; IDs are session-local handles, never persistence or cross-query identities.

    func PrepareRewrite(document QueryDocument, probes []RewriteFactProbe) (*RewriteSession, error)
    func (s *RewriteSession) Evidence() RewriteEvidence
    func (s *RewriteSession) Render(changes []RewriteReplacement) (*RewriteRendering, error)
    func (s *RewriteSession) Verify(candidate *RewriteSession, rendered *RewriteRendering) RewriteProof

    type RewriteIdentity struct { Name *string; Path []string }
    type RewriteFactProbe struct { Kind string; Identity RewriteIdentity }
    type RewriteReplacement struct { SiteID string; Target RewriteIdentity }
    type RewriteEvidence struct { Analysis Result; Sites []RewriteSite }
    type RewriteSite struct {
        ID, ReferenceID, BindingID, SourceEpochID string
        Kind, Role, Eligibility string
        Identity RewriteIdentity
        Location, OwnerLocation Location
        Point RewritePoint
        Facts []RewriteFactEvidence
        Limitations []RewriteLimitation
    }
    type RewritePoint struct {
        ScopeID, StageID, Phase string
        LineageIndex, Ordinal int
    }
    type RewriteFactEvidence struct {
        ProbeIndex int
        GuaranteedValues []RewriteScalar
        LiteralComplete bool
        ReferenceState string
        ReferenceIDs []string
        Locations []Location
        Limitations []RewriteLimitation
    }
    type RewriteScalar struct { Kind string; Value json.RawMessage }
    type RewriteLimitation struct { Code, Message string; Location Location }
    type RewriteTextEdit struct { Location Location; Before, After string; SiteIDs []string }
    type RewriteRequirement struct {
        CauseSiteIDs []string
        RequiredChanges []RewriteReplacement
        Limitations []RewriteLimitation
    }
    type RewriteIdentityEffect struct { ReferenceID string; Before, After RewriteIdentity }
    func (r *RewriteRendering) Edits() []RewriteTextEdit
    func (r *RewriteRendering) Requirements() []RewriteRequirement
    func (r *RewriteRendering) Effects() []RewriteIdentityEffect
    type RewriteReferencePair struct { OriginalID, CandidateID string }
    type RewriteProof struct {
        Proven bool
        References []RewriteReferencePair
        Limitations []RewriteLimitation
    }

Root approved the following bounded coordinate clarification after reviewing the complete proposal; it was not part of the original approved preflight's facade. The preflight assigns Unicode/CRLF conversion to sourceIndex, but exposes no converter to pkg/rewrite. At fcbb74ac and unchanged through accepted 40e6e171, `pkg/analysis/source.go` SHA256 is `c0421ed1db6a70c689396801a40d515fa2c887f273571cd704baa9af314c628b` (Git blob `8f1b17b002a5a5cf3af24a23eff0c24ae741afd1`). Its sourceIndex, constructor and position/location methods are private; location accepts rune offsets and clamps them. The package's existing exported API provides no arbitrary-text byte-coordinate service. Passing candidate byte offsets to the rune-indexed method would be incorrect.

    type RewriteByteRange struct { Start, End int }
    func LocateRewriteBytes(text string, ranges []RewriteByteRange) ([]Location, error)

Task 2 places the public helper/value and private sourceIndex.byteLocation method together in rewrite_evidence.go. It validates UTF-8 and every half-open byte range: 0 <= Start <= End <= len(text), with both endpoints exactly on UTF-8 rune boundaries or document EOF. Empty ranges at a valid boundary are allowed; invalid UTF-8, reversed/out-of-bounds ranges and endpoints inside a multibyte rune return an ordinary error and no partial results. Empty ranges input returns an empty nonnil Location slice. Results follow input order. There is no clamping, parsing, normalization, text change, alias/binding inference or semantic evidence creation.

The helper builds the existing newSourceIndex once for the supplied text, then resolves all range endpoints by exact Position.Offset lookup in its already populated positions. Define only a private `func (s *sourceIndex) byteLocation(start, end int) (Location, bool)` in rewrite_evidence.go for that exact lookup, using ordered search of Position.Offset and rejecting missing boundaries. Reuse the existing Position values so Unicode columns, tabs, CRLF and EOF behavior remain canonical. Preserve all existing rune-indexed position/location behavior. This is a batched coordinate utility, not a public SourceIndex/AST or another line/column implementation.

Task 4 keeps candidate construction, nonoverlapping edit validation and cumulative byte-offset translation. Once the final selected edits have produced the exact candidate string, it batches their translated replacement ranges through analysis.LocateRewriteBytes to obtain audit CandidateLocations. This covers arbitrary replacements and zero-width insertions even if no candidate Reference exists. Candidate references, scopes or diagnostics that exist only after analysis remain Task 5/Verify evidence; any needed byte-range conversion uses the same helper/private facility. Coordinates from this utility never establish render/session provenance or semantic proof. Render's approved values and Verify's authority remain unchanged, so there is no duplicate candidate construction or premature parse introduced merely to obtain locations.

PrepareRewrite uses the same canonical Analyze path and retains its own typed owners/source bytes. Attach call-local evidence at existing typed operand/lowering sites and actual `readAt`, `createAt`, `selectorAt`, source reset, clone, assignment, rename, projection, aggregate, lookup and removal transfers. Retain explicit alias boundaries independent of transitive origins. `LineageIndex` is actual execution sequence; `Ordinal` distinguishes multiple reads/transfers within a phase. Missing canonical references use an empty ReferenceID with a real located refusal, never a fabricated reference.

Collect literal guarantees from typed per-dialect search/comparison/Boolean owners before annotating candidate sites in the same predicate. Carry and invalidate them in the actual environment rather than replaying public lineage. Existing expression presence is not equality evidence. Scalar Kind is string/number/boolean/null and Value is one strictly decoded canonical scalar with exact numeric precision. Unknown literal syntax stays unknown. LiteralComplete=false does not erase a separately guaranteed matching value; an unestablished value remains unknown. Conclusive empty guarantees mean false query-fact evidence, never event absence. Every probe gets explicit evidence at every site, including ReferenceState true/false/unknown. Rewrite owns equals/contains and all/any composition, not predicate parsing.

Render consumes simultaneous proposals at original sites. It deterministically encodes original role contexts, composite intervals and already-proven implicit-name changes; it returns edits plus required linked consumer changes. It does not authorize those changes, resolve rule winners or cascade identities. The rewrite layer independently checks each required site's original-point condition and withholds a group if any member fails. Composite/implicit references can have normalization Effects even when their whole semantic location is not a standalone editable token. Rendering is pure over original session state; mutation of copied evidence or a rendering from another session cannot create proof.

Verify first checks candidate document language/profile/version/source_id and original rendering/session provenance. It uses canonical candidate analysis, translated typed ownership and selected edits to compare intended identities, direct binding relationships, mapped origins, transition operations/inputs/outputs/conditionality and actual SQL phases. It neither reparses independently nor promises runtime equivalence. Remap all sidecar StageIDs at `spl2FinalizeStages` and reference/fact/implicit IDs at `finalizeReferences`; equal spans or equal ref-N values are insufficient. Translate every affected reference/owner/scope/diagnostic endpoint, including enclosing references and zero-width lookup AS insertion. SELECT preparation creates one physical reference/read owner; final project supplies a later transition using that reference. Unknown evidence is compared by translated ownership and effect, never diagnostic counts. Task 5 may complete a specifically scoped reviewed correspondence follow-up in analysis; adapters never own proof.

## Proven forms and explicit refusal boundary


SPL exact source reads/selectors, rename inputs and exact removals can map only with a selected source binding and typed identifier encoding. Explicit eval/rename/SQL/lookup outputs and downstream explicit aliases remain fixed. Wildcard selectors, dynamic names and creation-only or unavailable/indeterminate bindings do not become source mappings merely because names match. Null inspection keeps `null_test`: a source-bound exact operand can render, but unknown/unavailable null inspection never materializes a source field or proves event presence.

SPL2 exact bare or quoted field atoms use direct FieldName.Identifier without navigation. `'actor.name'` is an atom. Unresolved AccessPart navigation and alias-qualified joins remain located refusals; do not rewrite a source-looking root while the enclosing whole identity is unproved. Dotted resolver ambiguity remains governed by the existing typed source policy. Literal punctuation is assessed by role: an actual wildcard selector remains refused even quoted, while an exact quoted expression atom can render if its frontend proves encoding. Meaningful whitespace, keyword and punctuation targets are logical data, never raw query fragments.

Index/source/sourcetype mappings target typed dependency values, not same-spelled event fields. Search values, metric predicate values, SQL fields and quoted Boolean/string literals retain their own grammar categories. M5's `metricsPredicate` recognizes only an intact command-owned bare-index equality; BY/aggregate fields remain fields, wildcard/dynamic selectors stay refused. Dependency discovery through OR/NOT is not positive equality evidence. Interpolation, held signed/punctuation spellings or uncertain literal decoding never receive false definite render/fact credit.

Lookup catalog names are independent of columns. Explicit local match aliases can follow a proven source identity. An unaliased remote/local dual-purpose input must preserve the remote column and insert a documented AS local target only when canonical correspondence proves it; otherwise refuse it. Lookup outputs stay explicit derived names. Without explicit outputs, M5's possible remote overwrite makes affected aliases conditional; M6 must not infer linked consumer safety across that uncertainty.

Named external datasets can render independently of source aliases, including typed UNION/join dependency intentions only when local edit and final correspondence are independently proved; none of these tokens establishes a parent merge. Literal-row keys are derived field definitions. SPL datamodel/dataset composites retain their exact component owner and prefix/quotes: `datamodel Model Dataset` normalizes Model.Dataset across a Dataset-only interval, and qualified FROM/tstats references may overlap. Co-render compatible component edits and their normalization Effects; refuse conflict or unproved backslash forms instead of writing a full normalized name into a component interval.

For SPL2 tstats, consume `metricsInputs`' existing atomic data_model reference/dependency. It includes the entire quoted token, e.g. `'Traffic.All'`; never split model/dataset/path or duplicate extraction. Preserve canonical expectations for `C28-P2`, `L22-P1`, `E.C28.datamodel_name.P1/P2`, and invalid N1/N2 plus held mstats boundaries. Add only the private render owner and rewrite positive/held/dynamic/damaged cases. Recognition does not waive affected-effects/group/postverification proof. `search_job` is not an accepted Rule.Kind. Deferred field output intentions are not editable source permissions.

SPL's implicit aggregate label is frontend-owned and can require an atomic input/consumer group: `search index=main | stats sum(bytes) | table 'sum(bytes)'` with bytes→octets must become `search index=main | stats sum(octets) | table 'sum(octets)'`, or retain the whole group if proof fails. Never add a synthetic AS. SPL2's accepted implicit labels remain narrow (`count()`, `sum(bytes)`, `dc(action)`); mapping bytes to octets in `FROM main | stats sum(bytes) | table 'sum(bytes)'` cannot establish the changed label or indeterminate consumer and must withhold the whole group. Explicit AS outputs stay fixed. Direct unaliased SQL field projection is passthrough, not a synthetic name. Do not broaden implicit naming simply because SELECT now proves a finite output set or isolated count presence.

The original syntax-error gate applies before edits even where M5 recovery retains an intact typed operand. Comments, strings, object-key definitions, option labels, function names, source aliases and unmodeled named arguments are not inferred source sites. Add private naming owners at real SPL/SPL2 aggregate and SELECT creation only; preserve M5 assignment, deferred-effect, incomplete-wildcard and SQL collision/conditionality guards.

## Additive rewrite capabilities


Add `Rewrite *RewriteCapabilityManifest` with JSON key `rewrite,omitempty` to analysis.CapabilityManifest. Task 2 owns the canonical implemented role/form table; Task 8 completes documentation/corpus correspondence. Use this compact approved shape:

    type RewriteCapabilityManifest struct {
        SchemaVersion int `json:"schema_version"`
        Forms []RewriteCapabilityForm `json:"forms"`
    }
    type RewriteCapabilityForm struct {
        Kind string `json:"kind"`
        Role string `json:"role"`
        IdentityForms []string `json:"identity_forms"`
        Supported bool `json:"supported"`
        Limitations []string `json:"limitations"`
    }

Role is a stable documented rendering context fine-grained enough to distinguish exact expression atoms, selector atoms, null inspections, lookup local operands, dependency-value slots, composite catalog components and implicit consumers from refused navigation/wildcard/dynamic forms. It is not merely canonical read/create or a generated ANTLR type. Populate both selected dialect manifests from implemented renderer eligibility; preserve all existing default values, order, documentation_snapshot and copied nested slices. Intentionally additive capability JSON differs as a whole; ordinary Analyze reports stay unchanged. Text capabilities and OpenAPI expose the same optional addition without changing selector APIs or claiming universal seven-kind support.

## Strict wire examples and result decisions


All request objects use exact member names with unknown/duplicate/null-required members rejected. schema_version is integer 1. mode is preview or apply; omission alone defaults to preview. A rule requires unique nonempty id, kind, source and target, with optional when. Kind is exactly field/index/source/sourcetype/lookup/dataset/data_model. Identity contains exactly one name string or nonempty path array of nonempty logical segments; path is field-only. Reject empty/whitespace-only names without trimming meaningful whitespace or case. Empty rules are valid; duplicate semantic rules with different IDs coalesce identical edits; source=target is an explained no_change. No priority, cascading, regex, caller context, arbitrary expression or executable rule exists.

Use this complete request as an independent Task 1 decoder example and later Task 3/5 behavior case:

    {
      "schema_version": 1,
      "mode": "preview",
      "document": {"text": "search sourcetype=auth src=alice | table src", "language": "spl"},
      "rules": [{
        "id": "auth-user", "kind": "field",
        "source": {"name": "src"}, "target": {"name": "user"},
        "when": {"fact": "literal", "kind": "sourcetype", "identity": {"name": "sourcetype"}, "operator": "equals", "value": "auth"}
      }]
    }

Condition is exactly one leaf/combinator. Literal leaves have fact=literal, kind field/index/source/sourcetype, identity, operator equals/contains and a scalar value. Contains requires a string, uses exact case-sensitive substring search on proven literal values and never scans source text. Reference leaves have only fact=source_reference_present, kind=field and identity. Combinators are {all:[conditions]} or {any:[conditions]} with nonempty children; there is no condition-language NOT. A literal null is present data and differs from missing value. Numeric equality retains exact precision, including 9007199254740992 versus 9007199254740993. Preserve evidence for every child even when a known false all-child or true any-child controls the result.

For example, a field reference leaf is {"fact":"source_reference_present","kind":"field","identity":{"name":"src"}}; a dotted atom {"name":"actor.name"} is different from {"path":["actor","name"]}. A field-list destination target is {"kind":"field_list","catalog":{"fields":["user"]}}. JSON Schema and OCSF targets use their existing strict SchemaTarget objects unchanged. The CLI RuleSet has only schema_version and rules; apply mode never comes from a rules file. Batch replaces document with a nonempty documents array and shares rules/mode/target; all request preparation and internal operations must complete before publishing any batch report.

Result keys include schema_version, document, mode, status, coverage, original_text, candidate_text, text, committed, changes, rule_evaluations, original_analysis, candidate_analysis and optional candidate_validation. Empty collections encode as arrays. Top-level invalid outranks incomplete outranks valid. Preview may mark a change applied/candidate_applied=true but keeps committed=false and text original. Apply publishes candidate text only after all selected-group/static/error/optional-target gates pass. A failed apply keeps candidate evidence but changes candidate-included applied audit entries to skipped/post_verification_failed, candidate_applied=true, committed=false and text original. No-op never fabricates applied entries or commits; a requested target still validates the unchanged candidate. Ambiguous changes never alter candidate text. An independent safe group can commit beside an unchanged unknown/ambiguous group with overall incomplete status; exit 3 therefore is not synonymous with no commit.

Stable uncertainty/conflict reasons include condition_unknown, unsupported_reference, dynamic_reference, conflicting_targets, overlapping_edits, binding_collision, linked_edit_unproven, target_not_renderable and post_verification_failed. Ordinary no_match, condition_false, no_change and explicit derived names outside source mapping do not alone make a result incomplete. Keep original analysis/schema diagnostics and real locations; absent references have rule-level evaluations without invented spans. Do not confuse mapped static identity with event presence, unseen-field absence or equivalent query execution.

## Task 1: Strict independent request, rule, target and result contracts


Own new pkg/rewrite/model.go, request.go, target.go, diagnostics.go and their focused tests, plus initial durable testdata/rewrite/requests.json. Consume real analysis.QueryDocument and validation FieldCatalog/SchemaTarget/report types from the accepted predecessor. Produce the approved public request/rule/decoder/report types and input-error classification used by every later task; no rewrite engine or successful stub is needed yet.

1. Write table-driven TestDecodeRewriteRequest and TestDecodeRewriteBatchRequest covering required keys, duplicate keys/IDs, unknown members, UTF-8/lone-surrogate rejection, identity atom/path distinctions, kind/identity incompatibility, null versus missing, exact scalar values, source-equals-target, empty rules and malformed documents/options/targets. Include numbers around 9007199254740992 so float rounding cannot make different rules equal. Seed requests.json with independently specified accepted/rejected shapes; no generated expected snapshots.

2. Run `go test -mod=readonly ./pkg/rewrite -run 'TestDecodeRewrite|TestRewriteModel'` and record the expected missing-definition failures. Add minimal types, strict key/presence decoding using the accepted internal/jsoninput facilities, deep-copy normalization and delegation to canonical document/target normalization. Do not copy a query parser or add Python-side JSON policy. Source/target names preserve meaningful whitespace; they are never parsed as caller-supplied SPL fragments.

3. Add TestRewriteModelJSON distinguishing empty arrays from null, preview candidate-applied from committed, optional absent target/location, and rule-level unmatched evidence without a fabricated reference. Ensure direct Go requests reject the same invalid values as decoded JSON. A null literal value is accepted only in the supported literal condition union, never as an omitted required member.

4. Rerun focused tests and the accepted strict validation request regressions if a shared JSON helper changed. Commit only owned files after root execution release; pass the immutable commit to a fresh reviewer for combined spec/code review. Resolve findings before Task 2. Do not commit other workers' files.


## Task 2: Canonical evidence and rendering for proven forms


Own only the new analysis evidence/facts/render/correspondence files and surgical frontend/kernel/finalizer/capability hooks enumerated above, their focused tests, and durable testdata/rewrite/forms.json. Consume actual typed M5 contexts, field environment/source provenance, source-index conversion, phase and finalization functions. Produce the preflight-approved canonical evidence/render bridge used by pkg/rewrite; it is a narrow per-call edit service, not an independent field engine or public AST. The facade signatures and concrete hooks are recorded above; production still awaits root's exact plan/application/implementation release.

1. Add TestRewriteEvidenceSPL and TestRewriteEvidenceSPL2 that demand exact identity, original render interval, role, binding, scope/flow point, origin IDs, guaranteed literal/reference facts and limitations for ordinary source fields, rename inputs, explicit aliases, lookup input/output distinctions and each of the six nonfield mapping kinds. Test lexical-context negative controls: equal text in comments, string literals, option values and catalog-column labels is not an eligible source field. The tests must initially fail because canonical edit evidence is absent, not because the query corpus is malformed.

2. Attach evidence collection to existing typed frontend lowering and canonical transfer/phase callbacks. Record positive predicate facts from typed comparison/Boolean contexts and their supporting source references; maintain them with the existing environment copies/barriers rather than reconstructing transfer in rewrite. Field redefinition/removal/source reset invalidates affected facts; independent child scopes do not inherit or export parent facts. Source-reference-present is tied to the applicable source identity/flow evidence. Add the narrow typed static literal decoders/encoders in the canonical frontend bridge; normalizedName and spl2DecodeKey are not universal encoders. Unfamiliar scalar forms stay unknown.

3. Add TestRewriteRenderIdentity with quoted identifiers, reserved words, dependencies, static path forms actually proven at preflight, Unicode and punctuation such as a target containing `| eval pwn=1`. Rendering must yield one exact logical identity in the original grammar role or a located target_not_renderable refusal; it never inserts a command. Preserve original quote style when valid and choose minimal correct quoting otherwise. Atom `actor.name` never silently becomes path actor/name. Test both dialects rather than inheriting SPL quoting in SPL2. Add TestLocateRewriteBytes for the new helper: `LocateRewriteBytes("é\r\n😀x", []RewriteByteRange{{4,8},{9,9}})` returns line 2 columns 1–2 for the emoji range and line 2 column 3 at EOF, retaining byte offsets 4/8 and 9/9. Assert tabs, CRLF boundaries, empty text/range arrays, arbitrary ranges without references, zero-width insertion points and all-or-error rejection of invalid UTF-8, mid-rune endpoints, reversed and out-of-bounds ranges.

4. Preserve ordinary public analysis JSON byte values, including M5's already present tstats dependency. Add no duplicate extraction or new rule kind. The optional rewrite capability section is intentionally additive to capabilities only. Reuse finalization so evidence reference IDs align with ordinary reports after nested scopes/SQL phases. Test duplicate SELECT phase observations, implicit output name ownership, wildcard/dynamic refusal and unknown/unsupported contexts. A supported read in an unrelated complete scope may retain evidence even when another scope is incomplete; a damaged original query is not declared eligible for apply.

5. Run `go test -mod=readonly ./pkg/analysis -run 'TestRewriteEvidence|TestRewriteRender|TestLocateRewriteBytes|TestFinalize'` and focused existing SPL/SPL2/validation regressions for touched helpers. No broad new path semantics, parser alternative or schema engine belongs to this task. Commit/review the exact hook change, and record the implemented signatures and immutable commit in this living plan for subsequent implementers.



Task2 implementation record (2026-09-08): completed and independently reviewed at 0b8dd62068be0638e470a532544acd1cc0a48507 (initial 02c2e321, six-file correction). The facade/value signatures in the canonical bridge section are implemented unchanged. Eligibility values are eligible/ineligible; ReferenceState is true/false/unknown; all facts are indexed by supplied ProbeIndex. PrepareRewrite uses canonical Analyze; rendering accessors return copied edits/requirements/effects and Verify consumes session-bound provenance. LocateRewriteBytes lives with private byteLocation in rewrite_evidence.go; source.go remains unchanged. Final focused and analysis/validation/rewrite checks passed on the recorded committed bytes. Full report/review/rereview and source receipts are in this plan's SDD workspace. I1–I5 and M1 are closed; the rereview's separate metric Boolean typing observation O1 is now queued for the root-authorized producer interlude described below. Text/OpenAPI exposure and downstream engine gates remain at their assigned consuming tasks.

## Task 3: Three-valued rule selection from canonical facts


Own pkg/rewrite/conditions.go, conditions_test.go and durable testdata/rewrite/conditions.json. Consume prepared Condition/Rule from Task 1 and immutable canonical candidate-point evidence from Task 2. Produce deterministic rule evaluations and per-candidate proposals containing all rule/evidence IDs and true/false/unknown reasons. Do not update field flow or parse expressions here.

1. Write TestConditionTruthTables for all/any. All is false if any child is false, true if every child is true, otherwise unknown. Any is true if any child is true, false if every child is false, otherwise unknown. Add known-control cases with unknown siblings, retaining all child evidence even when the controlling result is conclusive.

2. Write TestRewriteConditionScopeAndOrder using `search sourcetype=a src=x | table src`, `search (sourcetype=a OR sourcetype=b) src=x | table src`, a common-a OR predicate, query NOT and `search src=x | where EventCode=1 | table src`. The later equality cannot authorize earlier src. Independent child equality cannot authorize parent src; a redefinition/removal/unknown barrier makes only affected evidence unusable. Derive expected false/unknown results from complete supported evidence, not assumptions about event values.

3. Implement equals, string contains and source_reference_present over canonical facts only. Contains compares established string values, never raw query text. Distinguish typed strings/numbers/bools/null without lossy coercion. Unknown literal representation, dynamic binding or unsupported relevant predicate produces unknown, while a fully understood lack of the requested query fact produces false. A reference fact does not become equality or event existence.

4. Check that rule ordering never acts as precedence and all decisions use original facts, so a source-type mapping does not trigger another rule in the same request. Return ordinary condition_false skips and explicit condition_unknown incomplete evidence. Run `go test -mod=readonly ./pkg/rewrite -run 'TestCondition|TestRewriteCondition'`, then commit/review before candidate construction.


## Task 4: Linked edits, conflict closure and byte-preserving preview candidate


Own pkg/rewrite/groups.go, edits.go, their tests and durable testdata/rewrite/edits.json. Consume proposals/evidence from Tasks 2–3 and produce a candidate plus precise proposed-edit audit. This task does not publish apply text or bypass the later verification gate. An edit group is the smallest connected set that must change together to preserve a proven source/implicit-name relationship.

1. Add TestRewriteLinkedAliases asserting src-to-user changes `search src=x | rename src AS owner | table owner` to `search user=x | rename user AS owner | table owner`. owner remains explicit and unchanged; an owner source rule cannot rename it. Add implicit sum(src) and exact downstream implicit-name references using valid dialect quoting from forms.json; either all canonically linked edits occur or the whole group is skipped. Explicit AS names do not change.

2. Build linked groups from canonical identities and origin/transition evidence. Require condition authorization at each required edit's own flow point. A later true rule cannot authorize an earlier false/unknown occurrence; refuse incomplete linkage rather than split one source identity into differently named consumers. Wildcards, dynamic consumers or a relevant unknown merge make the affected group unproven. Unrelated scopes remain independent.

3. Add TestRewriteConflicts for two different true targets, identical replacements with multiple IDs, overlapping spans, target capture by an explicit alias, distinct observed live identities collapsing onto one target, original-identity swaps/chains and a skipped edit that another group needed for safety. Resolve conflicts deterministically to a stable surviving set. Coalesce identical edits, reject connected contradictory groups, and never choose first/last rule or search for a maximum-count rewrite set. Do not require proof that unreferenced event fields are absent.

4. Add TestRewriteBytePreservationAndLocations. Apply validated disjoint original intervals in a single forward reconstruction; check old text against the original slice, copy untouched bytes exactly and compute cumulative byte-offset translation. Batch the resulting half-open replacement ranges with the exact candidate text through `analysis.LocateRewriteBytes`; use its returned Locations for candidate audit positions. The rewrite package does not rescan runes/CRLF to compute lines or columns, and does not pass byte offsets to the private rune-indexed sourceIndex.location method. Test CRLF, tabs, comments containing pipes, quotes, Unicode before/inside edits, multibyte and newline-containing logical-name rendering, arbitrary edit spans without a canonical reference and zero-width insertions. Unchanged text, including literal `<EOF>` text, is not cleaned up.

5. Audit original/candidate text, location, reference IDs when resolvable, group/rule IDs and proposed inclusion. Ambiguous entries never enter the candidate. No_match/condition_false/no_change are ordinary rule-level skips; unknown eligibility adds incomplete evidence. Run `go test -mod=readonly ./pkg/rewrite -run 'TestRewriteLinked|TestRewriteConflicts|TestRewriteByte'`, commit/review, and retain counterexamples as durable fixtures rather than loose temporary files.


## Task 5: Canonical post-verification, apply policy and ordered batches


Own pkg/rewrite/verify.go, rewrite.go, batch.go, focused tests and durable testdata/rewrite/cases.json. Consume the accepted canonical evidence bridge, candidate edits and real public validation APIs. Produce Rewrite/RewriteBatch and complete canonical Result/BatchResult values. Any necessary narrow canonical proof helper is reviewed with analysis ownership rather than implemented as a second flow engine in verify.go.

1. Write TestRewritePreviewApply with the explicit-alias example and an independently reviewed full report. Preview returns candidate_applied true for included edits, committed false and authoritative original text. Apply returns candidate text and committed true only after proof. Empty/no-match/no-change rules yield unchanged bytes, committed false and no fabricated applied entries. Reanalyze an unchanged candidate or safely reuse equal original analysis; still evaluate an explicitly requested target.

2. Reanalyze each candidate with identical dialect/profile/version/source_id and build original-to-candidate reference correspondence from the real edit translation and canonical identities. Compare affected roles, scopes, binding provenance, origin chains, transitions and SQL phases under the requested mapping. Parser success is insufficient. Check explicit aliases unchanged, implicit generated-name consumers moved consistently, and no newly introduced capture/unavailable bindings or affected uncertainty. Unchanged unsupported evidence is compared by translated locations and semantic ownership, never diagnostic counts alone.

3. Add TestRewritePostVerificationFailure for original syntax damage, invalid candidate syntax induced at the internal edit-test seam, a syntactically valid but wrong binding/capture, lost implicit linkage and newly uncertain affected flow. The test seam supplies an internal malformed proposed edit only in package tests; no caller API bypass exists. Failed apply retains original text, keeps candidate analysis/text as evidence, and changes previously proposed applied entries to skipped/post_verification_failed with candidate_applied true and committed false. Do not silently publish a second reduced candidate.

4. Add TestRewriteDestinationValidation covering field-list missing target field, JSON Schema missing/conditional/unresolved cases, and a real local OCSF catalog target from accepted durable fixtures. Validation describes destination inputs, so old source names absent from the target do not pre-reject a valid migration. Require the final validation report status valid for commit, including complete analysis/schema coverage. Preserve original canonical validation report data. Without a target, unrelated incomplete scope evidence may remain if edited groups are independently proven; with an incomplete target report, commit fails. A definite validation error yields invalid; otherwise unknown validation/rewrite evidence yields incomplete.

5. Implement batches in two phases: prepare/validate request shapes and rules, then form ordered candidate documents and call existing ValidateBatch or ValidateSchemaBatch once if requested. Attach each returned canonical report at its original index, run per-document gates and only publish the batch after all input/internal operations succeed. Invalid query syntax is a per-query invalid result; malformed options, target, empty batch or internal failure rejects the whole request without a partial report. There are no external mutation transactions: each successful apply result commits only its returned string. Aggregate invalid outranks incomplete outranks valid.

6. Add TestRewriteBatchOrderAndAtomicErrors and TestRewriteConcurrentOwnership for mixed dialect/status documents, safe groups beside ambiguous groups, caller mutation after request preparation and repeated/concurrent equality. Run `go test -mod=readonly ./pkg/rewrite ./pkg/analysis ./pkg/validation` with targeted race tests in this task if root's stable window permits. Commit/review Task 5, then reserve a Go-only acceptance window with immutable input identity before dispatching adapters.


## Exact CLI/HTTP and OpenAPI integration


Task 6 adds `cmd/rewrite_cli.go`, `cmd/rewrite_cli_test.go`, `pkg/api/rewrite.go`, `pkg/api/rewrite_test.go`, and only the necessary changes to `cmd/cli.go`, `cmd/main.go`, `cmd/analysis.go`, `pkg/api/server.go`, legacy guidance in `pkg/api/models.go`, maintained example assertions in `pkg/api/dialect_test.go`, CLI/API guides and `tests/acceptance/cli_examples.json`. Existing helper signatures remain:

    func runCLIWithInput(args []string, stdin io.Reader, stdout, stderr io.Writer) int
    func parseCLIOptions(command string, args []string) (cliOptions, string, error)
    func marshalCLILine(value any) ([]byte, int, error)
    func writeCLIError(w io.Writer, format, message string, code int) int
    func writeCLIResult(payload []byte, output string, stdout io.Writer) error
    func analysisStatusExitCode(status analysis.Status) int
    func formatAnalysisLocation(location analysis.Location) string
    func setSchemaCLIOption(options *cliOptions, name, value string) error
    func readSchemaCLITarget(options cliOptions) (validation.SchemaTarget, error)

Implement `runRewriteCLI(args []string, stdin io.Reader, stdout, stderr io.Writer) int`. Extend the shared parser's validation-only file/stdin/batch admission to rewrite, add `--rules FILE` and valueless nonduplicate `--apply`, allow empty query/source-id via analysisOptionMayBeEmpty, and reject legacy --config. One query source is allowed: positional/--query, --file, --stdin or --batch FILE; --batch - reads stdin. Every explicitly supplied global document option is rejected for batches, even defaults/empty values. File source_id defaults to supplied path; stdin uses `<stdin>`; inline uses empty unless explicitly overridden. Shared duplicate, --name=value, -- termination and text/json rules remain unchanged.

Require rules file schema_version/rules only. Allow zero or one target family --fields/--schema/--ocsf-catalog; reject cross-family or orphan modifiers. Input target/resource/rules files are local paths, not `-`. Reuse DecodeFieldCatalog and readSchemaCLITarget, but not validateSchemaCLIOptions wholesale because rewrite targets are optional. Keep canonical OCSF list/UID validation. Successful content JSON is the complete canonical report plus newline; JSON request errors go to stderr as `{"error":"..."}` plus newline, text errors as `Error: ...`. Report --output uses existing 0644 creation semantics and leaves stdout empty. Preserve exits 0/1/3/2, including incomplete committed=true.

Register handlers `handleRewrite` and `handleRewriteBatch` as `POST /api/v1/query/rewrite` and `/api/v1/query/rewrite/batch`. Both call `readSchemaValidationBody(w,r)` directly: application/json, 8<<20 bytes, error `request body exceeds 8 MiB limit`. Keep known/unknown-length boundary tests and genuine base/windows OCSF catalogs in compact/raw form through both routes, using existing schemaREST/readSchemaFixture support. `writeRewriteError` recognizes rewrite.IsInputError and wrapped validation input errors, then reuses writeErrorResponse. It cannot blindly reuse writeValidationError, which knows only validation errors. Return complete content reports with HTTP 200; malformed transport/request gets 400 and unexpected internal failure 500.

Update stale legacy “SPL2 rewriting is not available” guidance in cmd/cli.go and pkg/api/models.go without changing legacy rejection category/shape or routing legacy operations through rewrite. The legacy 64 KiB cap and mapper context do not become canonical rewrite policy. Keep the 14 current marked API examples and add explicit rewrite dispatch/count assertions in dialect_test.go; CLI examples must remain paired with executable markers.

Extend `tools/update_validation_openapi.py` and its tests for strict schema_version/mode/rule/condition/identity/optional-target/single/batch request shapes while preserving report schemas and original target selection metadata. The actual generator is pinned swag v2.0.0-rc4 with `init --v3.1 -g cmd/server/main.go -o docs`, followed by the existing reconciler. Generate/reconcile JSON/YAML/Go together; do not hand-patch one output or substitute latest Swag. Keep idempotence, shape-drift/no-write and optional capability/phase checks. Raw JSON literals must not become base64 bytes. Documentation-only adapter DTOs can follow the existing schema-validation DTO pattern while canonical decoding remains authoritative.

## Task 6: CLI and HTTP parity with executable documentation


Own new cmd/rewrite_cli.go and tests, pkg/api/rewrite.go and tests, minimal registrations in cmd/cli.go and pkg/api/server.go, actual accepted OpenAPI reconciliation tooling/generated docs, docs/cli.md, docs/api-server.md and tests/acceptance/cli_examples.json. Consume Task 5 Request/BatchRequest decoders and canonical operations; produce thin report transports. Use the exact integrated helper and request-limit contracts above, reconciling only any later accepted M5 delta and preserving maintained examples.

1. Add `rewrite --rules <local-json>` with preview default and explicit `--apply`. Reuse query/positional/file/stdin/batch and document options, field-list/schema/OCSF selectors, format text/json and report output behavior. Rules file uses RuleSet; never let a file silently supply apply mode. Reject competing inputs/options and global dialect flags in document-batch mode. No in-place source-file editing flag is added.

2. Test full canonical JSON equality, exact source bytes/source_id, text labels for candidate versus returned text, no-op and failed-apply audit, unknown/ambiguous partial commit, all exits 0/1/3/2 and output write failures. A report must be written before a content-status exit; file output must not duplicate stdout. Exit 3 alone cannot be described as a refusal because committed may be true.

3. Register POST `/api/v1/query/rewrite` and `/api/v1/query/rewrite/batch` through existing middleware and canonical decoders, selecting M4's 8 MiB schema-route body policy for both. Preserve 1 MiB legacy routes and add no global limit/config feature. HTTP 200 carries all query-level statuses; request errors are 400 and internal failures are server errors. Inline targets/rules are data; server filesystem paths and retrieval URLs are not accepted. Test malformed duplicate/null/selector/target inputs and single/batch JSON equality. Reuse the accepted M4 exact-bound and +1-byte mechanisms to prove the 8,388,608-byte bound, including its established transport error on overflow. Exercise genuine normal/selected-extension OCSF catalogs through single and batch rewrite requests; do not substitute tiny synthetic tables as payload/compatibility proof.

4. Add maintained docs/cli.md example markers with matching argv/stdin/exit/stdout/stderr/output_files entries in cli_examples.json. Execute the actual documented CLI harness, including stdin and output-file behavior, rather than copying expected strings from docs without execution. Update HTTP example assertions and OpenAPI through the accepted offline recipe; run idempotence/shape reconciliation tests. Run `go test -mod=readonly ./cmd ./pkg/api` and focused documented CLI/tooling tests. Commit/review the complete surface slice before native packaging.


## Exact native, package and surface integration


Task 7 adds the following functions in `pkg/bindings/bindings.go` using `ownedMapperJSONResult(mapperID C.int, operation func() (any, error)) *C.SPLResult`, canonical decode/rewrite operations and no mapper-config injection:

    //export spl_mapper_rewrite
    func spl_mapper_rewrite(mapperID C.int, requestJSON *C.char) *C.SPLResult
    //export spl_mapper_rewrite_batch
    func spl_mapper_rewrite_batch(mapperID C.int, requestJSON *C.char) *C.SPLResult

The existing helper allocates/initializes the owned result, admits a live registry handle, retains the pointer across the operation and marshals canonical JSON. Query statuses stay in result JSON; malformed/null requests become owned errors. `spl_result_free` frees both optional strings and the outer object exactly once; nil remains permitted. Registry removal prevents later admission but does not cancel an admitted operation. Python close separately waits for admitted calls. Existing 18 exports become 20 if final accepted M5 adds none; verify actual counts instead of hardcoding them as acceptance.

Add ctypes signatures `[ctypes.c_int, ctypes.c_char_p] -> ctypes.POINTER(SPLResult)` and these wrapper methods:

    def rewrite(self, query, rules, *, mode="preview", validation_target=None,
                language="spl", profile="splunkd", version="current", source_id="") -> dict:
        request = {"schema_version": 1, "mode": mode, "rules": rules,
                   "document": {"text": query, "language": language, "profile": profile,
                                "version": version, "source_id": source_id}}
        if validation_target is not None:
            request["validation_target"] = validation_target
        return self._validate_fields_request(self._lib.spl_mapper_rewrite, request, operation="rewrite")
    def rewrite_batch(self, documents, rules, *, mode="preview", validation_target=None) -> dict:
        request = {"schema_version": 1, "mode": mode, "rules": rules, "documents": documents}
        if validation_target is not None:
            request["validation_target"] = validation_target
        return self._validate_fields_request(self._lib.spl_mapper_rewrite_batch, request, operation="rewrite")

Bodies build schema_version=1 requests, preserve caller query/documents/rules, omit validation_target when None and forward other values to strict native decoding. Reuse `_validate_fields_request(self,native,request)` by adding keyword `operation="validation"`; keep existing message text by default and use operation="rewrite" for these methods. Keep `_operation`, allow_nan=False, copied decoded JSON, finally-free, close wait/idempotence and all existing loader/version behavior. Do not duplicate the ctypes/free loop or add a Python parser/validator. Preserve `_legacy_selectors`' `unsupported_dialect_for_operation`; update only guidance to include rewrite. Raw C tests cover duplicates and malformed Unicode that Python dicts cannot express; Python tests cover NaN/Inf/nonserializable rejection, invalid/freed handles, decode failure, concurrency and close admission. Add focused new ownership calls in `tests/native/memory.c`; source Python green is not ASAN acceptance.

Use `python/build_support.py`'s actual native build recipe and require both the library and generated .h. Update `python/spl_toolkit/libspl_toolkit.h` from actual output. Refresh both complete `tools/tests/fixtures/cgo-go1.22.h` and `cgo-go1.25.h` from real corresponding Go builds of the same source. Never insert two declarations synthetically or relax `toolkit_header_contract`'s exact compiler allowances. Preserve existing whole-header drift tests and recorded source/actual wheel header hashes.

Append every new non-test pkg/rewrite Go file, new analysis bridge/helper and any actual shared internal helper to `python/native-source-files.txt`. M5 already owns tstats extraction in an existing listed file; no new dependency file is required. Preserve all 65 predecessor entries and the 13 SPL2 fixture registrations. Extend tools/tests/test_package.py's native closure enumeration to pkg/rewrite and any real shared helper. Keep missing/duplicate/traversal/sdist-source/hash checks intact. Add tests/test_native_rewrite.py to SDIST_FIXED_FILES, test_native_rewrite.py to NATIVE_TESTS and verify_sdist_sources' explicit handwritten hash list. MANIFEST.in inclusion alone does not satisfy exact checker closure.

The checker, unchanged through accepted 40e6e171, has 28 fixed sdist members, six mandatory native modules, 11 schema fixture inputs, 13 SPL2 fixture inputs and eight acceptance inputs (six test modules, spl2_transport.py and cli_examples.json). Preserve the named suites and helper; do not infer test counts from file counts. Add test_rewrite_surfaces.py to ACCEPTANCE_FILES. In tools/check_acceptance.py add test_native_rewrite.py to REQUIRED_TEST_FILES.native and test_rewrite_surfaces.py to REQUIRED_TEST_FILES.acceptance. Register explicit REWRITE_FIXTURE_FILES for every durable rules/conditions/forms/edits/cases input used by native/surface suites. Copy and hash them before both suites under an absolute outside-checkout SPL_REWRITE_FIXTURES, clear inherited overrides in clean_env, and record fixture_hashes.rewrite. Copy maintained rewrite docs with existing copy_documentation; preserve README/docs/cli/docs/spl2 requirements.

M5's install_and_check now also accepts `go_transport: Path`; _check_package prepares the full Go SPL2 report artifact once, copies it for each install and sets SPL_SPL2_GO_REPORTS plus SPL_SPL2_GO_SHA256. `load_transport` validates producing source/fixture hashes, all document IDs and complete report shapes with conformance_credit=0. Preserve this mechanism and its source-freshness requirement after analysis changes; a matching transport snapshot never replaces independent semantic obligations. Both direct-wheel and rebuilt-sdist paths must keep required nonzero collection, zero skips/failures and collected=passed. Extend required-suite/missing-input/corrupted-copy/stale-source tests rather than increasing a numeric threshold alone. Carry schema_surface_evidence, spl2_surface_evidence, documentation_hashes and their canonical fixture/artifact identities into resulting evidence; add rewrite evidence consistently if required, without loosening record validation.

Task 8's new rewrite surface suite reuses `tests/acceptance/test_surfaces.py` helpers `cli_path`, `server_url`, `post_json`, `required_absolute_path`, and schema loader helpers when needed. server_url launches the supplied absolute SPL_SERVER with an ephemeral loopback PORT, a bounded readiness check at /api/v1/health and explicit cleanup/log retention; do not invent a second server lifecycle. Source runs use absolute SPL_NATIVE_LIBRARY; installed runs load the packaged module/library and omit that override. Source fixture assertions remain independently authored; full canonical transport parity compares every array/value/location/binding/diagnostic, ignoring only object-key order. Add safe preview/apply/ordered-batch examples to python/examples/basic_usage.py while retaining its existing schema and legacy examples. No release-platform/format change is indicated.

## Task 7: Owned native Python operations and installed package closure


Own pkg/bindings/bindings.go and focused tests, generated python/spl_toolkit/libspl_toolkit.h, python/spl_toolkit/mapper.py, new python/tests/test_native_rewrite.py, relevant ABI/unit tests, python/native-source-files.txt, tools/check_package.py and its focused tests. Include both real generated ABI fixture headers and minimal tests/native/memory.c coverage. Change python/build_support.py only when actual source/attribution closure requires it. Consume Task 5 Go APIs and the final accepted M5 owned-result/document-option conventions.

1. Add owned C JSON operations for rewrite and rewrite_batch without changing existing export signatures. Use the same handle admission and result-allocation/free protocol as accepted structured analysis/validation; malformed requests return the existing error envelope. Treat closed handles and concurrent destroy/operation behavior consistently. Do not use a free-form native context map to bypass rule facts.

2. Add Python rewrite/rewrite_batch methods accepting strict rules, mode/target and normal document selectors, returning canonical dictionaries. Forward document requests to native code; preserve _operation guard and finally freeing on decode success, error and exceptions. Request failures raise the established mapper exception; invalid/incomplete query results remain reports. Unit mocks supplement rather than replace actual library tests.

3. Execute test_native_rewrite against the source-built library for both dialects, representative all-rule kinds, preview/apply, failed validation, batch order, Unicode and concurrency. Compare complete expected canonical values, not just fields or counters. Native ownership tests must observe no lost result frees and unchanged closed-handle behavior.

4. Extend the actual accepted native source manifest with every new handwritten/generated source needed for rewrite. Extend check_package's mandatory test/fixture copy closure with testdata/rewrite and any referenced existing schema/SPL2 fixtures. Use an absolute SPL_REWRITE_FIXTURES outside checkout for installed tests, retain accepted schema/SPL2 environment inputs, and preserve isolated import/module provenance checks. Required suites must collect nonzero cases with zero skips; missing fixtures or library loading is failure, not a skip.

5. Build and check a wheel plus a wheel rebuilt from sdist through the accepted offline Go-floor/Python recipe. Capture package/native library hashes and installed module origin; source-native proof is not installed-package proof. Run focused tooling closure tests and native suites before root's package window; do not overwrite shared build outputs during another worker's acceptance. Commit/review all native and closure changes together.


## Task 8: Durable corpus, full parity, capability truth and handoff


Own final testdata/rewrite files, pkg/rewrite/corpus_test.go, tests/acceptance/test_rewrite_surfaces.py, required-suite registration in actual accepted tooling, docs/rewrite.md, minimal maintained README/API/architecture/compatibility/Python documentation, and the eventual `_build_plan/milestones/6-safe-rewrite-mapping/milestone-log.md`. Consume all prior tasks and the actual supported-form/capability bridge. Do not hide a missing supported form by shrinking the corpus or changing independent expectations to match current output.

1. Give every durable case a unique ID, explicit dialect/document/rules/mode/target, expected candidate/returned text, status/committed decision, exact audit expectations and source/binding/lineage assertions. Keep logical expectations independent from generated snapshots. The forms manifest lists all seven mapping kinds and both dialects, each admitted role with positive and negative context cases, and all advertised refusals. A query unsupported in one dialect remains an explicit limitation case rather than copied as passing coverage.

2. Complete obligations for compatible AND/common OR/noncommon OR/NOT, false versus unknown, contains only on proven literals, overwritten/later/child facts, explicit aliases, implicit names, independent/inherited scopes, SQL phases, all identity kinds, atom/path distinction, collisions/chains/swaps, dependent skipped groups, Unicode/CRLF/quoting/comments, malformed syntax, candidate syntax/binding failure, schema invalid/incomplete rollback, no-op and batch request atomicity. Required groups are named in the fixture harness so deleting all cases in one group fails collection. Report actual meaningful case counts without inflating spelling variants.

3. Compare full canonical JSON across Go, built CLI, real HTTP, actual source-native Python, installed wheel and wheel-from-sdist. Only object key order may differ; all arrays, text, IDs, locations, diagnostics and schema metadata remain significant. Validate both dialects and ordered mixed-status batches. Source/candidate slices are checked independently from reported locations. The full corpus runs offline and no schema/catalog URL is opened.

4. Publish rewrite capability limitations using the minimal accepted additive shape, and update maintained examples/contracts for preview versus apply, original versus candidate/returned text, partial commit with incomplete status, strict validation, explicit aliases and static collision scope. Documentation states current as the local build contract and denies no existing capability. Existing legacy mapping documentation identifies its different precedence/context semantics without changing them.

5. Run final qualified commands below on immutable final executable inputs. Preserve all predecessor mapper/analysis/validation/native/CLI/docs tests. Update the milestone log starting with `## What's new in the app`, then files, decisions/deviations, exact accepted dependency/final commits, actual corpus counts, current hashes, canonical/surface/installed evidence and unexecuted platform/runtime gates. Commit/review Task 8, reserve root's stable built/native/installed acceptance window, obtain broad milestone review and independent root acceptance. Do not merge or push.


## Serial SDD and review ownership


After final accepted M5, final plan approval and explicit implementation release only, the controller dispatches one implementer at a time with exact owned paths, actual predecessor/helper interfaces and its bounded task. Every implementer is told it is not alone and must preserve others' changes. Before any protected implementation/debugging/architecture/security/review dispatch, read current `/Users/jacobdelgado/.codex/AGENTS.md` and `skills/efficient-delegation/SKILL.md`, check catalog and role overrides, and explicitly assign at least `gpt-6-astra/xhigh` with a non-overriding default/worker role and isolated context. Retired High coordinators and fixed lower-capability specialist roles cannot do protected work. If no qualifying delegate exists, retain protected work locally. Assigned settings are not backend execution attestation. Routine mechanical extraction can use an eligible capable economical worker. Capture command identities, raw logs, exits, failures and qualifications; never redo unchanged scans or side effects merely for summaries. Each task follows red test, observed failure, minimal implementation, focused green checks and immutable owned-file commit. The independent reviewer assesses both approved behavior and code/test quality against that immutable task commit; concrete fixes get a new commit and scoped re-review before the next task.

Do not give multiple implementers concurrent ownership of analysis helpers, adapters, package lists, documentation manifests or generated output. Task 2 is the only ordinary owner of canonical hooks; later proof gaps require a specifically scoped reviewed follow-up rather than an adapter inventing semantics. Root retains integration/commit acceptance authority and decides exact acceptance windows. Reviewers do not replace root's separate Go-only and built/native/installed acceptance.

After Task 5's reviewed immutable core commit, reserve root's Go-only window before Task 6 dispatch. Root's current independent acceptance inputs remain root-owned, unread and unexecuted by implementers; root binds them to final public APIs. Do not copy its cases into expected production snapshots. The later window requires final stable CLI/server/native library/wheel/sdist identities plus root's independently selected actual Python 3.14 environment. No shared output or HEAD mutation occurs during either window until root reports findings or completion.


## Concrete verification commands


These are future commands after root release, not results of this private preparation. Run from `/Users/jacobdelgado/repos/spl-toolkit/.worktrees/remaining-milestones`. The accepted fixed-source M5 evidence records these provisioned recipes; final documentation-only closure at 237e62ac is reconciled without changing them. Preserve the recorded precommit build metadata qualifications; no silent toolchain upgrade or network fallback is allowed.

    export GOCACHE=/private/tmp/spl-toolkit-go-cache
    export GOMODCACHE=/private/tmp/spl-toolkit-remaining-gomodcache
    export GOPROXY=off
    export GOTOOLCHAIN=local
    export PIP_CACHE_DIR=/private/tmp/spl-toolkit-pip-cache
    go test -mod=readonly ./pkg/rewrite
    /private/tmp/spl-toolkit-floor-tools/gomodcache/golang.org/toolchain@v0.0.1-go1.22.12.darwin-arm64/bin/go test -mod=readonly -ldflags=-linkmode=external ./pkg/rewrite ./pkg/analysis ./pkg/validation ./pkg/mapper ./pkg/bindings ./cmd ./pkg/api
    make build-all
    /private/tmp/spl-toolkit-remaining-venv/bin/python tools/check_go.py
    /private/tmp/spl-toolkit-remaining-venv/bin/python -m pytest tools/tests -q
    /private/tmp/spl-toolkit-remaining-venv/bin/python tools/check_docs.py

M5 records Go1.22.12 external linking for the established local dyld LC_UUID issue. Its initial API floor attempt was blocked by sandbox loopback permission; the same API scope passed with authorized listener access. Keep that qualification instead of treating every historical exit as current success or rerunning the known default-link failure. tools/check_go.py owns the full race suite, so an unchanged extra full race run adds no evidence. Focused post-fix tests may run when warranted. If no grammar changed, do not regenerate predecessor parser files; if a minimal approved frontend syntax edit is needed, use the accepted pinned ANTLR 4.13.2/JDK/runtime recipe and hash-idempotence check for those files.

After Task 6, execute maintained CLI examples with the actual built binary:

    SPL_CLI=/Users/jacobdelgado/repos/spl-toolkit/.worktrees/remaining-milestones/build/spl-toolkit SPL_DOCS_ROOT=/Users/jacobdelgado/repos/spl-toolkit/.worktrees/remaining-milestones /private/tmp/spl-toolkit-remaining-venv/bin/python -m pytest tests/acceptance/test_documented_cli.py -q

After native build, source-native evidence uses the built library and durable source fixtures:

    PYTHONPATH=python SPL_REWRITE_FIXTURES=/Users/jacobdelgado/repos/spl-toolkit/.worktrees/remaining-milestones/testdata/rewrite SPL_SCHEMA_FIXTURES=/Users/jacobdelgado/repos/spl-toolkit/.worktrees/remaining-milestones/testdata/schemas SPL_SPL2_FIXTURES=/Users/jacobdelgado/repos/spl-toolkit/.worktrees/remaining-milestones/testdata/spl2 SPL_EXPECTED_VERSION=$(cat VERSION) SPL_NATIVE_LIBRARY=/Users/jacobdelgado/repos/spl-toolkit/.worktrees/remaining-milestones/build/libspl_toolkit.dylib /private/tmp/spl-toolkit-remaining-venv/bin/python -m pytest python/tests -q

Preserve SPL_SCHEMA_FIXTURES and SPL_SPL2_FIXTURES. SPL_REWRITE_FIXTURES is the new M6 harness input. Source surfaces use the same actual ephemeral server fixture described above. After the new suite exists, run it with all required absolute paths (the source SPL2 fixture can prepare its fresh transport once when no SPL_SPL2_GO_REPORTS is provided):

    PYTHONPATH=python SPL_REWRITE_FIXTURES=/Users/jacobdelgado/repos/spl-toolkit/.worktrees/remaining-milestones/testdata/rewrite SPL_SCHEMA_FIXTURES=/Users/jacobdelgado/repos/spl-toolkit/.worktrees/remaining-milestones/testdata/schemas SPL_SPL2_FIXTURES=/Users/jacobdelgado/repos/spl-toolkit/.worktrees/remaining-milestones/testdata/spl2 SPL_EXPECTED_VERSION=$(cat VERSION) SPL_NATIVE_LIBRARY=/Users/jacobdelgado/repos/spl-toolkit/.worktrees/remaining-milestones/build/libspl_toolkit.dylib SPL_CLI=/Users/jacobdelgado/repos/spl-toolkit/.worktrees/remaining-milestones/build/spl-toolkit SPL_SERVER=/Users/jacobdelgado/repos/spl-toolkit/.worktrees/remaining-milestones/build/spl-toolkit-server SPL_DOCS_ROOT=/Users/jacobdelgado/repos/spl-toolkit/.worktrees/remaining-milestones /private/tmp/spl-toolkit-remaining-venv/bin/python -m pytest tests/acceptance/test_documented_cli.py tests/acceptance/test_schema_surfaces.py tests/acceptance/test_spl2_surfaces.py tests/acceptance/test_rewrite_surfaces.py -q

If supplying a reused SPL2 transport file instead, both its producing source and fixture hashes must match current immutable inputs; a stale M5 artifact is not reusable after M6 analysis edits. No runtime downloads are allowed; preserve dead external proxies and working loopback used by installed tests.

OpenAPI generation uses the accepted local module proxy rather than network retrieval:

    GOCACHE=/private/tmp/spl-toolkit-go-cache GOMODCACHE=/private/tmp/spl-toolkit-remaining-gomodcache GOPROXY=file:///private/tmp/spl-toolkit-remaining-gomodcache/cache/download GOSUMDB=off GOTOOLCHAIN=local make generate-docs PYTHON=/private/tmp/spl-toolkit-remaining-venv/bin/python
    /private/tmp/spl-toolkit-remaining-venv/bin/python -m pytest tools/tests/test_update_validation_openapi.py -q

Use the actual strictly offline M5 wheelhouse recipe, preserving installed suite/source closure:

    PATH=/private/tmp/spl-toolkit-floor-tools/gomodcache/golang.org/toolchain@v0.0.1-go1.22.12.darwin-arm64/bin:$PATH PIP_NO_INDEX=1 PIP_FIND_LINKS=/private/tmp/spl-toolkit-offline-wheelhouse PIP_CONFIG_FILE=/dev/null make python-build PYTHON=/private/tmp/spl-toolkit-remaining-venv/bin/python PIP=/private/tmp/spl-toolkit-remaining-venv/bin/pip
    PIP_NO_INDEX=1 PIP_FIND_LINKS=/private/tmp/spl-toolkit-offline-wheelhouse PIP_CONFIG_FILE=/dev/null /private/tmp/spl-toolkit-remaining-venv/bin/python tools/check_package.py --sdist dist/spl_toolkit-$(cat VERSION).tar.gz --wheel-dir dist --cli /Users/jacobdelgado/repos/spl-toolkit/.worktrees/remaining-milestones/build/spl-toolkit --server /Users/jacobdelgado/repos/spl-toolkit/.worktrees/remaining-milestones/build/spl-toolkit-server --evidence /private/tmp/spl-toolkit-m6-package-evidence.json

Installed acceptance must use packaged library/module paths outside checkout, collect new required rewrite suites with nonzero count and zero skips, and preserve accepted CLI/schema/SPL2 cases. A source venv run is not evidence for installed Python 3.14; root owns its exact interpreter/environment and independent acceptance recipe. Keep shared outputs immutable during its window. Do not invoke destructive clean targets or misclassify offline empty-cache failures as product regressions.


## Validation and Acceptance


Once a durable rules file contains schema_version 1 and rule `{id: source-user, kind: field, source: {name: src}, target: {name: user}}`, the observable CLI command is `build/spl-toolkit rewrite --rules testdata/rewrite/example-rules.json --query 'search src=alice | rename src AS owner | table owner' --format json`. In preview it exits 0, returns original text as text, shows the candidate with both source src references changed to user, preserves owner and reports committed false. Add `--apply` and it returns that candidate as text with committed true. Task 8 owns materializing this exact example-rules.json and reviewing the full expected report; it does not exist merely because this plan names it.

A target field list containing only src makes the user candidate fail destination validation. Apply then exits 1, reports invalid, returns original text, retains candidate evidence and marks all included edits uncommitted. An incomplete schema target or affected binding produces incomplete with no commit. Two different targets for the same original reference yield ambiguity; a separate independently safe group may still commit, producing incomplete and committed true with exact audit. Unmatched rules produce no-match skips and preserve an otherwise valid result.

SQL, nested scopes and implicit-output examples must meet the same source/binding checks in their own dialect rather than just share field-name sets. Every advertised mapping kind needs working positive examples and context-negative controls; unsupported forms remain specific refusals. Corpus review independently checks semantic expectations before accepting generated canonical snapshots.

Final proof records accepted base/final SHA, source status, exact executable/native/wheel/sdist hashes, actual test counts and tools, report parity and remaining limits. Local macOS automated proof is distinct from unexecuted cross-platform release jobs and Splunk runtime equivalence, which this product does not promise. Root gives the final exact-head conclusion after review and stable acceptance.


## Idempotence and Recovery


The operations produce strings/reports only and need no database or data migration. Request preparation, candidate construction and verification are deterministic and safe to repeat. A failed apply never modifies its input string/file and retains the preview candidate; a failed batch publishes no partial report for input/internal errors. Retrying with the same request yields equal report values.

During development preserve the failing fixture, fix the narrow owned paths and rerun affected checks. Do not reset the shared tree, replace manifests from stale commits, clean another milestone's build directory or automatically bless new snapshots. A removed conflict group must trigger dependent-safety reevaluation before verification. Once the final candidate fails verification, return original text rather than searching for a different reduced apply result.


## Artifacts and Notes


The approved design hash is `4355ca99ec2807025dafda26cff66b6768287f004a6f9930a0482c066063adba`. This private plan derives from repository draft SHA256 `e087f92156a07b46de037beae754c063640efe4b9770d26e28a647d718c6e3a5`, core preflight `7cca1be6826e956c43140e2d16703b25fa8586dab4c6480845817b4ba819830d`, CLI/REST preflight `1b8c9ba06e2d679e8b3bbbb63730b65ca218a88287dae604f8346597647ffb11`, and native/package preflight `8714af82801d9f60fdea1bb5e24e77bcb3af2889d039f2b62f6a7f46de10035a`. Root's preparation handoff hash is `e5a8596922095a0bdd3fb289560d0531ea87208bd04078dec8a65295ce32b696`.

The private preparation directory retains source-identities.json with exact SHA256/Git blob identities, inspection-commands.json with command exits, complete immutable scoped diffs, relevant candidate source snapshots and a concise delta receipt. The adapter diff is empty. These are preparation evidence only. Runtime/package/test behavior must never read temporary plan artifacts or private root expectations. Durable rewrite fixtures belong in testdata/rewrite and maintained documentation under docs.

Self-review maps strict request/target/result contracts to Task 1; canonical facts/render/phase/binding/form/capability evidence and the approved coordinate helper to Task 2; three-valued selection to Task 3; simultaneous linked edits/collisions/bytes and canonical candidate coordinates to Task 4; proof/schema/rollback/batches to Task 5; CLI/HTTP/OpenAPI/examples to Task 6; owned native/Python/headers/package closure to Task 7; durable full parity/required-suite truth/documentation/handoff to Task 8. Root approved the explicit coordinate bridge clarification. The two broad-review domain/category corrections are reconciled against accepted 40e6e171, with no change to the eight tasks or their approved interfaces. Final authored-document closure identity reconciliation is complete at 237e62ac; root's exact plan/application/implementation release remains mandatory.

Revision 2026-09-08: Integrated the existing approved preflights and actual intervening provisional M5 changes into a private copy. Replaced old M4/M5-proposal orientation with concrete interfaces, consumed M5's atomic tstats extraction, retained strict SQL/recovery boundaries, updated native refusal and mandatory installed transport/docs closure, and preserved eight serial tasks plus root's independent acceptance windows. The repository plan is unchanged. No product tests/builds/generators or implementation dispatch occurred; this file remains provisional until root accepts final M5 and the resulting plan.

Revision 2026-09-08, root-requested interface clarification: The preflight facade and existing exported API do not expose arbitrary candidate-coordinate conversion to pkg/rewrite. Proposed LocateRewriteBytes plus RewriteByteRange, backed by one private exact-byte lookup into existing sourceIndex positions, and connected Task 2's ownership/tests to Task 4's coordinate use. This proposal preserves byte-edit construction in rewrite and semantic proof in Render/Verify; it creates no second line/column implementation or new parser. Only this private plan was amended; the prior c6e82990 plan/delta receipt remains historical preparation evidence. At that revision the helper remained unimplemented and awaited root review; the subsequent approval is recorded below.


Revision 2026-09-08, coordinate approval: Root approved the exact LocateRewriteBytes proposal and connected Tasks 2/4 after reviewing private plan 663a20a3. Preserved that complete proposal unchanged as milestone-6-implementation-plan.coordinate-proposal-663a20a3.md, SHA256 663a20a3a296477c4ea045e659a9fa435966018a43d1f36a89da29f143b68411. Marked the helper contract approved and located its public helper/value plus private sourceIndex.byteLocation method in rewrite_evidence.go, preserving existing source.go. No implementation or product verification occurred; all final accepted-M5 and plan/repository/production gates remain open.

Revision 2026-09-08, accepted fixed-source reconciliation: Preserved the complete root-read private plan as milestone-6-implementation-plan.pre-final-source-e518c17d.md, SHA256 e518c17da6c8f3549c02a3356be2731488fb34deedae841454b383484c6c2af8. Reconciled only fcbb74ac-to-40e6e171: six source/test files and twelve owned evidence additions. Bound the numeric len/unknown-scalar split domains, approved module diagnostic category and retained verification to source commit 7571f1cb and root receipt 567ba8d9. The approved design, preflights, coordinate bridge, task ownership and protected Astra/XHigh floor remain unchanged. This revision closes the fixed-source delta only; authored-document closure and root repository-plan/production release remain open. No repository edits, product tests/builds or subordinate dispatch occurred.

Revision 2026-09-08, final accepted-M5 closure: Preserved the complete fixed-source private plan as milestone-6-implementation-plan.pre-closure-d167b733.md, SHA256 d167b733bfac6b68cc7a05034b274112811a638797edbe091bf3cece04bffdf7. Consumed the released public handoff and root closure audit, verified the direct 40e6e171-to-237e62ac parent and all fourteen documentation/evidence path hashes, and bound final M5 acceptance without changing source interfaces, tasks or approved coordinate behavior. The prior e518c17d and d167b733 versions remain recoverable; root receives the complete e518-to-final diff for narrow approval. No repository edits, product tests/builds, private-oracle access or subordinate dispatch occurred. Only root repository-plan acceptance/application and explicit production release remain open.

Revision 2026-09-08, root approval and production release: Applied the exact root-approved private 86ce9d9c plan with administrative title, introduction, Progress and this revision entry only. Preserved its full bytes as milestone-6-implementation-plan.root-approved-86ce9d9c.md in the private preparation directory. Root release receipt 4d21108d authorizes serial SDD with fresh Astra/XHigh implementers and independent combined task reviews, then reserved Task 5 and Task 8 root windows and broad review. No task contract or implementation requirement changed.

Revision 2026-09-08, root Task1 review ruling I1: Clarified the existing no-repeated-catalog wire requirement while retaining complete Go validation reports. Only rewrite field_list.target serialization omits fields/optional_fields; every other value remains canonical. Root approved the exact interpretation after reading the review, serialization source and design/plan contracts. The original design and private approved plan bytes remain historical and unchanged; scoped RED/fix/GREEN/re-review now enforce the clarification.

Revision 2026-09-08, Task2 typed composite ownership: Added only dependencies.go:qualifiedCatalog private owner registration to the existing surgical hook ownership list after checking the exact accepted/current source identity. Existing typed context/component spans justify the hook; all extraction/public behavior and approved rendering/form boundaries remain unchanged.


## Fresh-session checkpoint — 2026-09-12

The user requested a coherent handoff after the current atomic unit. M6 Tasks1–4 remain accepted through14a3a0dc. The interrupted single-query Rewrite/finishRewrite unit is now committed at 7927ac3c372d4c359b07e0bedab64e09db96a5a3 after independent bounded source review (zero findings) and fresh passing rewrite, analysis and validation package tests. A missing temporary dependency cache was restored from the existing local pinned cache; the setup failure remains recorded. Task5 is **not complete**: batches, negative-proof seam coverage, final corpus/concurrency checks, full Task5 review and private Go acceptance remain. M6 adapters/final verification and all M7 production remain gated. See GOAL_STATE.md and .superpowers/sdd/milestone-6-implementation-plan/task-5-checkpoint/closure.json. No old worker handle is live; do not infer a running task from earlier notes.
