# Feature Implementation Plan: Milestone 8 Product Interface and Requirements API

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development to implement this plan task-by-task. Mark completed Progress items and keep all living sections current.

**Goal:** Make query requirements a canonical, deterministic product interface across Go, CLI, REST, C, Python, contracts, packages, and permanent documentation without changing the meaning of existing analysis evidence.

**Architecture:** Extend the single analysis pass with a private query-only semantic trace that survives refinement without inheriting target-dependent facts. Before parser prediction, enforce one 4,096-unit lexer work budget across SPL and SPL2, then return a bounded incomplete report when unit 4,097 would be consumed. Project the public `RequirementSet` only after canonical reference finalization, embed it in every runtime `analysis.Result`, and expose detached copies through thin adapters. Contracts preserve archived v1 compatibility by making the new property optional when decoding stored analysis and document snapshots, while every current runtime producer emits it.

**Tech Stack:** Go 1.22, the repository's custom CLI parser, `net/http`, cgo and a stable C ABI, Python 3.11 through 3.14 with `ctypes`, JSON Schema draft 2020-12, Swag/OpenAPI generation, pytest, Go race tests, AddressSanitizer, GitHub Actions, and Claude CLI review.

---

This file is a living execution plan. The implementing orchestrator must update `Progress`, `Surprises & Discoveries`, `Decision Log`, and `Outcomes & Retrospective` as work proceeds. The approved behavioral source is `_build_plan/milestones/8-product-interface-requirements/design-spec.md`; this plan restates the required behavior so implementers and reviewers must not load `_build_plan/` at runtime or from tests, package manifests, release archives, generated documentation, or installed distributions. This plan and `_build_plan/milestones/8-product-interface-requirements/milestone-log.md` are administrative records and are always allowed task edits even when a task's owned-file list omits them. Finish and commit all administrative updates before final whole-feature review and freeze.

## Progress

- [x] (2026-09-14) Read the approved Milestone 8 design specification, current source and tests, release tooling, GitHub Actions, and relevant prior milestone plans.
- [x] (2026-09-14) Resolve the public model, single-pass refinement strategy, adapter boundaries, compatibility rules, fixture ownership, review protocol, and release procedure in this plan.
- [x] (2026-09-14) Task 1: Added public requirement value types, digest helpers, constants, and deep-copy primitives from base `c5de9590fe41b79a8b1181f73b13b2c92e7105a5`. The first focused run failed on the missing types and `queryDigest`; later RED runs failed on the missing `capabilityRevision`, `cloneRequirementSet`, and constants. The exact focused command and the full `pkg/analysis` package passed after implementation. The completion commit is the commit containing this entry.
- [x] (2026-09-14) Task 2: Captured a private query-only trace during the existing analysis traversal. Focused RED runs exposed the missing trace API, shadow removal and projection state, null-test policy, semantic diagnostic capture, macro ownership, refinement-only diagnostic duplication, SPL2 lexical-stage remapping, child-scope trace loss, qualified catalog classification, downstream dotted-name contamination, and dropped conditional field state. Independent specification review then found incomplete sidecar mirrors and ownership cases. Exact RED witnesses proved a renamed-away source, a field absent from a finite SPL2 dataset literal, a field absent after a sound SQL projection, a field read after a syntax-damaged SPL stage, a hidden SQL `HAVING` field, an SPL `OUTPUTNEW` field, a partial SPL2 dataset-literal field, and a field read after deferred SPL2 `bin` effects all had incorrect query-only classifications; the first of two macro diagnostics was also unowned. A final isolation review proved that refinement-aware public expression nullability could leak into private field installation. Query-only expression nullability and a dotted-name parity witness prevent that leak. A subsequent `createAt` audit found that rename, aggregate, and lookup outputs also installed public refinement conditionality into the private sidecar. Three exact plain-versus-refined RED witnesses now cover those transfer paths, and each field creation supplies its query-only conditional separately. Repository-wide validation then exposed a hidden-HAVING rewrite regression caused by sharing one diagnostic incompleteness flag between public rewrite state and the private trace. A focused RED witness now protects the public predicate reference IDs while the existing hidden-field witness protects trace incompleteness and exact ownership. The full `TestRequirementTrace` suite, the complete `pkg/analysis` package, the no-reentry scan, `git diff --check`, and the complete `pkg/rewrite` package pass after correction; the latest correction commit is the commit containing this entry.
- [x] (2026-09-14) Task 2 quality follow-up: An exact RED witness for `eval a=1, b=2 | rename a AS b | table a` showed that the private sidecar retained `a` as a definite derived field after the public conflict path cleared it. Mirroring the same source and destination invalidation sets changed the final query-only read to conditional `indeterminate`. A compile-time RED witness also rejected the variadic `finalizeReferences` signature; its trace argument is now explicit at every production and test call site. The focused trace suite, focused race run, complete `pkg/analysis` and `pkg/rewrite` packages, `go vet` for both packages, no-reentry scan, and `git diff --check` pass.
- [x] (2026-09-14) Task 2 rename-isolation follow-up: The exact `search alpha=1 | rename a* AS b | table alpha` RED witness proved that refinement-aware wildcard expansion fed private conflict invalidation and changed the final query-only `alpha` read from conditional `indeterminate` to direct `source`. The rename audit found a second RED witness, `search alpha=1 | fields a* | rename x AS alpha | table alpha`, where refinement changed existing-destination conflict detection and inserted a target reference that renumbered the downstream consuming reference. Public and query-only conflicts, invalidation sets, diagnostics, and field updates are now computed separately from the same lexical pairs. Every sound exact rename target emits located evidence even when the operation is rejected, without installing a field, adding a transition, or changing rewrite state. The two witnesses assert both public reference-list shape and byte-equivalent private trace parity. The bounded cross-task fixture deviation updates only eight current SPL2 canonical cases and the one hand-pinned conflicting-rename parity row; their field state, transitions, status, coverage, and diagnostic codes did not change.
- [x] (2026-09-14) Task 2 wildcard-target follow-up: The exact SPL2 RED witness `FROM [{'old_a':1,'new_a':2}] | rename 'old_*' AS 'new_*' | table new_a` showed that public conflict handling retained `new_a` as a definite derived field while the query-only sidecar invalidated it. Public wildcard targets now expand against the pre-rename public field snapshot, independently of the query-only expansion against its own snapshot; exact targets still enter both invalidation sets directly. The final public and query-only `new_a` reads are both conditional `indeterminate`. The focused witness and trace suite, focused race run, complete `pkg/analysis` and `pkg/rewrite` packages, `go vet` for both packages, no-reentry scan, and `git diff --check` pass without fixture changes.
- [x] (2026-09-14) Task 3 initial projection implementation, with acceptance later superseded: Projected finalized query-only trace evidence into deterministic requirement items and gaps, embedded the initialized set in every successful analysis result, and exposed a one-analysis detached Go operation. Named RED runs failed first on the missing projector, conditional gaps, unexplained completeness fallback, query status, `Result.Requirements`, `Requirements`, the absent 17-case requirement corpus, and additive full-result witnesses. The exact Task 3 gate and complete `pkg/analysis` package pass. The complete `pkg/rewrite` package reaches only the expected Task 4-owned `explicitAliasReport` mismatch because that hand-built result has not yet gained its explicit requirements value. The implementation commit is `b94814299f9007f7e37e3203c56596d4f0899829`.
- [x] (2026-09-14) Task 3 specification-review fix, with approval later superseded: Focused RED cases proved that intact single-field SPL `isnull` and `isnotnull` operands remained canonical `read` references and became external requirements. Canonical SPL analysis now reuses the typed ownership classifier already used by rewrite evidence, emits `null_test` for eligible operands, and retains the prior conservative behavior for multi-argument, compound, parenthesized, value-prefixed, aggregate, dynamic-parent, and malformed forms. The two focused suites, exact Task 3 gate, complete `pkg/analysis` package, focused rewrite conformance case, and `git diff --check` pass. Only the affected current requirement and rewrite conformance cases changed; historical evidence and current full-analysis fixtures were untouched. The fix commit is `b60e1ff3d0c524dd866b86bc8af28930ed0881cc`.
- [x] (2026-09-14) Task 3 quality-review fix, with approval later superseded: An adversarial projector benchmark with one owned incomplete diagnostic per wildcard reference reproduced the quadratic scans. After warm-up, 8,192 references took 280 to 283 ms/op. A compile-time RED test then required an ownership and gap index with exact evidence-key behavior. The projector now builds diagnostic ownership once by `(reference ID, code)`, deduplicates gaps by code plus ordered reference and diagnostic-code lists, and records coverage reasons through a first-seen set. The same benchmark now takes 6.7 to 7.8 ms/op at 8,192 references. Focused projector, requirements, corpus, and null-inspection tests, the complete `pkg/analysis` package, focused race tests, `go vet`, and `git diff --check` pass. A full repository run outside the listener-restricted sandbox passes every package except the Task 4-owned rewrite witness and validation and schema fixture updates. Current analysis fixtures remain byte-identical, and historical evidence was untouched. The fix commit is `d1db37cf022bd1cbb640fa5ef0c9bcfc3c65f28b`.
- [x] (2026-09-14) Task 3 security-review reopening: A blocking resource-exhaustion review at `d1db37cf022bd1cbb640fa5ef0c9bcfc3c65f28b` superseded the prior Task 3 specification and quality approvals for the range `d94451ca1ef2e37a8f754a1a60434c11672c0e3a..d1db37cf022bd1cbb640fa5ef0c9bcfc3c65f28b`. Every prior implementation SHA and verification result remains historical evidence, but Task 3 is not accepted. Fresh specification, quality, and security reviews must approve the original Task 3 base through the new head.
- [x] (2026-09-14) Task 3 resource-limit specification follow-up: Independent plan review pinned exact lexer event order, SPL2 synthetic closure accounting, canonical messages and ranges, resource-limited rewrite output, restoration of the fixed trace pointer, and bounded ASCII dense-fixture response evidence. The implementation evidence is recorded below, but no post-amendment review has completed and all prior Task 3 approvals remain superseded.
- [x] (2026-09-14) Task 3 resource-limit implementation: Added one shared 4,096-unit lexer tracker around each dialect's existing lexer and token stream. Both frontends fill that stream and account for real lexer errors before returned tokens; SPL2 also accounts for its existing synthetic literal-closure error before either frontend constructs a parser. Unit 4,097 returns the exact initialized incomplete analysis and requirement values with the full-text digest and no partial evidence. Trace pending-reference and incomplete-stage lookups now use synchronized maps retained through environment clones. RED first failed to compile on the absent limit and index contracts. The exact focused gate, focused race gate, benchmark, complete analysis package, vet, and diff check pass. The complete repository run outside the listener-restricted sandbox passes except for the Task 4-owned rewrite witness and validation and schema fixture updates. The 65,533-byte dense fixtures serialize to 959 requirement bytes and 67,258 result bytes for SPL, and 960 requirement bytes and 67,260 result bytes for SPL2. The 256 KiB fixtures serialize to 959 requirement bytes and 263,869 result bytes for SPL, and 960 requirement bytes and 263,871 result bytes for SPL2. The completion commit is the commit containing this entry. Fresh specification, quality, and security reviews remain required.
- [x] (2026-09-14) Task 3 lexer error-storm specification fix: Fresh specification review found that one generated lexer `NextToken` call could report and recover from every character in a consecutive invalid-character run after the tracker rejected unit 4,097. The SPL and SPL2 RED cases each observed all 32,768 lexer errors instead of stopping at the first omitted error. A private sentinel now unwinds only that generated lexer call and is recovered at the work-limiter boundary; unrelated panic values are re-panicked. GREEN observes exactly two errors, the admitted unit 4,096 error and the omitted unit 4,097 error, with the input cursor still on the omitted character. Only the admitted error enters the abandoned frontend diagnostics, and the returned public result contains only the resource-limit diagnostic. The SPL2 witness places an unmatched quote after the invalid run and proves that neither the later token nor closure evidence enters the preflight stream. The focused and race gates, complete analysis package, vet, adversarial benchmark, and diff check pass. The repository run retains only Task 4-owned rewrite and validation/schema fixture failures, and `pkg/api` passes outside the listener-restricted sandbox. This finding supersedes every Task 3 review result through `33616e24b9ac862a886786ec9e67d742477ecffe`; fresh specification, quality, and security reviews must approve the original Task 3 base through the new head.
- [x] (2026-09-14) Task 3 sentinel ownership quality-review fix: A same-package `panickingLexer` carrying the real package sentinel and an untouched tracker supplied the RED witness. The limiter swallowed that unowned panic as synthetic EOF because recovery checked pointer identity alone. Recovery now also requires this limiter to hold a non-nil tracker with a recorded resource limit; every other panic is re-panicked unchanged. GREEN preserves the real SPL and SPL2 error-storm aborts. The focused resource suite, focused race run, complete analysis package, vet, adversarial benchmark, and diff check pass. The repository run retains only the Task 4-owned rewrite and validation/schema fixture failures, while `pkg/api` passes with local listener access. This finding supersedes every Task 3 review result through `7d5854ab07a9ae64226b6614594aec81fe0cf8cc`; fresh specification, quality, and security reviews must approve the original Task 3 base through the new head.
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
- Public conditional field effects were not represented in the private field sidecar. The three exact RED witnesses returned public `indeterminate` reads while the trace returned `derived`: `search user=* | lookup users user OUTPUTNEW role | table role`, `FROM [{a:1},{b:2}] | table a`, and `FROM main | eval a=host | bin a | table a`.
- Passing public SPL2 expression nullability directly into the private sidecar broke refinement parity for `FROM main | eval local='actor.name' | table local`. Private nullability must follow query-only trace bindings while the public result continues to use refinement-aware bindings.
- Rename, aggregate, and lookup output installation reused public conditionality. Under dotted-name source refinement, this changed private destination reads from `derived,false` to `indeterminate,true` for `FROM main | rename 'actor.name' AS actor | table actor`, `FROM main | stats count('actor.name') AS total | table total`, and `FROM main | lookup users 'actor.name' OUTPUT role | table role`.
- An additional repository-wide test run found that the sandbox blocks the existing API `httptest` listener and that `pkg/rewrite` had a `sql-having-conflict-hidden` corpus mismatch. The full rewrite package passed at Task 2 commits `69967e3` and `7a53b06`, then first failed at `73fe6b4`. That commit changed the hidden-field diagnostic from public `incomplete=false` with an explicit incomplete stage to public `incomplete=true`, which invoked `rewriteUncertain` and discarded the visible predicate reference IDs.
- The rename conflict path invalidated public source and destination fields but left the same derived fields in `requirementEnvironment`. Existing-destination, duplicate-source, duplicate-destination, source/destination-overlap, and wildcard conflicts all share this return path and its `sources` and `dests` sets.
- Refinement-aware wildcard expansion and existing-destination checks can make public rename conflict outcomes differ from the query-only outcome. Reusing either public decision for the private sidecar changes requirement classification. Omitting exact targets on rejected renames also makes those divergent outcomes renumber later `ref-N` identities, so rejected exact targets must retain reference evidence without applying state.
- Unsupported wildcard rename targets can overlap concrete fields already present in a closed public snapshot. Retaining only the target pattern in the public invalidation set leaves those concrete fields definite even though the operation is unresolved.
- A variadic trace parameter let a caller omit the query-only trace or supply more than one without a compile-time error. The two production callers already supplied one trace, while two trace-free test helpers relied on omission.
- Adding a non-optional value to `analysis.Result` exposed two additional complete-result witnesses in `pkg/analysis/spl2_lower_test.go` and `pkg/analysis/spl2_sql_test.go`. Both needed explicit hand-derived requirement values for the full analysis package to remain an independent oracle. The rewrite package has the analogous `explicitAliasReport` witness, which is explicitly owned by Task 4 and remains the only expected Task 3-era rewrite failure.
- The analyzer has no current query syntax that emits a `Reference` with `resolution: dynamic`. Task 3 therefore covers defensible and gap-only dynamic identities at the projector boundary, while the full query corpus covers the current macro dynamic-expansion diagnostic and wildcard references.
- SPL rewrite evidence already classified eligible `isnull` and `isnotnull` operands as `null_test`, but ordinary SPL analysis still recorded them as `read`. Because the requirement projector correctly excludes `null_test`, the disagreement promoted a non-consuming inspection into an external field requirement.
- The original requirement projector searched every incomplete diagnostic for each conditional wildcard reference, then searched every accepted gap and coverage reason for each candidate. Its allocations grew linearly, but its CPU time grew quadratically on traces with many independently owned wildcard gaps.
- The Task 3 security review found a broader resource-exhaustion path before projection. A 65,533-byte dense wildcard query at pre-Milestone-8 `f02009d` took 5.28 seconds, used 203 MB, and serialized 12.8 MB of JSON. At `d1db37c`, the same query took 5.65 seconds, used 317 MB, and serialized 27.0 MB of JSON, including 14.2 MB of requirements. A 256 KiB dense query exceeded 60 seconds.
- Existing schema projection and enumeration code already uses 4,096-unit budgets. Reusing that scale for canonical lexer work gives one measurable boundary across SPL and SPL2 without rejecting long sparse documents that remain below the token and lexer-error count.
- The lexer-work bound limits parser admission and therefore every product surface, but it cannot cap every serialized result. A below-budget document can still amplify inherited lineage evidence, so final review and acceptance must describe this residual boundary rather than claim a universal output-size limit.
- SPL2 closure inspection creates one synthetic unterminated-literal error after token collection. Without explicit ordering, an implementation could construct the parser too early, omit that error from the budget, or report the wrong opener when it becomes unit 4,097.
- A resource-limited original rewrite session reaches `pkg/rewrite` after request normalization. The safe behavior needs a fully specified no-op report because returning an API error would contradict the admitted-content contract, while continuing to rule selection would use analysis that intentionally contains no references.
- A generated ANTLR lexer can report, recover from, and skip many consecutive invalid characters inside one `NextToken` call. Recording a rejected work event without unwinding that call does not enforce the hard stop, even when the wrapper returns a synthetic EOF after the call eventually completes.
- A package-global panic sentinel is only a transport marker. Pointer identity does not prove that a particular limiter recorded the overflow, so recovery must also verify that limiter's tracker state.
- Filling one `CommonTokenStream` before parser construction preserves the admitted parser token indexes and avoids a second lexer pass. The complete `pkg/analysis` package remained green without fixture changes after moving SPL2 literal-closure inspection ahead of parser construction while retaining its prior diagnostic insertion point.
- On the Apple M4 Max benchmark host, the 65,533-byte and 256 KiB dense SPL fixtures completed in 11.60 ms/op and 11.42 ms/op with 16.75 MB and 23.06 MB allocated per operation. The corresponding SPL2 fixtures completed in 1.55 ms/op and 2.05 ms/op with 3.28 MB and 9.24 MB allocated per operation. These measurements include full analysis and result JSON serialization.
- The sandboxed complete repository run failed only because `httptest` could not bind IPv6 loopback, plus the already assigned Task 4 rewrite and validation fixture work. The unrestricted rerun passed `pkg/api` and retained only the named Task 4 failures.

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
- **Decision:** Use the existing typed SPL ownership classifier to identify canonical null inspections instead of adding a second spelling-based function check. **Rationale:** One classifier preserves the established eligibility boundaries for nested, aggregate, malformed, and otherwise unproved function contexts while allowing analysis and rewrite evidence to agree on `null_test`.
- **Decision:** Build one projection-local index for incomplete diagnostic ownership, exact gap evidence keys, and first-seen coverage reasons. Encode each ordered evidence list with byte-length prefixes before using it in a comparable map key. **Rationale:** Map lookups remove the repeated scans without changing message selection, semantic event order, evidence-list order, reason order, or serialized requirement output; length prefixes prevent distinct ordered string lists from sharing a key.
- **Decision:** Copy every slice layer in `cloneRequirementSet` with an owned empty destination. **Rationale:** This detaches coverage reasons, item occurrences, gap links, and diagnostics while keeping empty cloned collections array-valued instead of `null`.
- **Decision:** Store `requirementEnvironment` as a private sidecar on `environment`; `environment.clone()` deep-copies its maps and origin slices while retaining the shared trace pointer. **Rationale:** Branch and scope semantics remain in the established traversal, refinement mutates only public field state, and trace code does not need a second parser, lowerer, transfer pass, or analysis call.
- **Decision:** Let the private `spl2FinalizeStages` return its existing old-to-final stage map and use it only from `spl2_lower.go` to remap trace references and diagnostics. **Rationale:** SPL2 child scopes otherwise leave trace stage IDs stale after lexical reordering; reusing the existing map preserves single-pass ownership with a three-line private plumbing change.
- **Decision:** Mirror only structurally proven removal and finite output-shape effects into `requirementEnvironment`: successful rename sources become tombstones, intact non-empty dataset literals close their shape, and sound SQL projections retain only selected fields and close uncertainty. **Rationale:** Downstream absent fields are query-local unavailable references, while incomplete or ambiguous effects still retain their established indeterminate behavior.
- **Decision:** Create each SPL macro reference and its owned dynamic-expansion diagnostic in one `macro` semantic event, keyed by the exact parser context so the later dependency walk cannot duplicate it. **Rationale:** Multiple macros in one stage then have exact ownership without code, stage, or location searches.
- **Decision:** Mirror public uncertainty into `requirementEnvironment` at every non-refinement parser and stage recovery boundary, while leaving refinement-only diagnostics out of the private trace. **Rationale:** Query damage makes later field provenance indeterminate, but target-specific evidence must not change standalone requirements.
- **Decision:** For hidden SQL restricted expressions, mark the temporary query-only environment uncertain before evaluating the expression, then attach each incomplete visibility diagnostic to the corresponding returned reference ID. **Rationale:** This classifies the hidden read conditionally and records direct ownership without matching diagnostics back to references by location.
- **Decision:** Store conditional field presence in `requirementField`, propagate it through private installation and cloning, and classify later reads as conditional and indeterminate. Mirror deferred SPL2 effects, lookup-without-output effects, and restricted SQL visibility only for the fields whose public effects are conditional. **Rationale:** Query-only evidence must retain unproved presence without turning a derived field into a definite source obligation or broadening uncertainty to unrelated known fields.
- **Decision:** Carry query-only SPL2 expression nullability beside public nullability and pass both values through assignment installation. **Rationale:** Source refinement can make a public operand indeterminate, but it cannot change standalone requirements for the same query.
- **Decision:** Pass separate public and query-only conditionality into field creation whenever refinement can affect the public value. Rename derives the private destination from the query-only source binding, aggregates use query-only expression nullability and trace-owned stage diagnostics, and lookups use query-owned `OUTPUTNEW` intent and trace-owned stage diagnostics. **Rationale:** Public source-universe evidence may change result lineage, but it must not change the standalone requirement trace.
- **Decision:** Carry public semantic incompleteness and requirement-trace incompleteness as separate flags at the diagnostic append boundary. A hidden SQL visibility diagnostic leaves public rewrite facts intact while explicitly marking its stage incomplete, but the same single public diagnostic records incomplete private evidence with exact reference owners. **Rationale:** Requirement coverage needs uncertainty that the established rewrite contract does not treat as a fact barrier.
- **Decision:** Compute public and query-only rename conflicts from separate field snapshots and invalidation sets while sharing the same lexical pairs and source-reference events. Record public-only or requirement-only conflict and unavailable-field diagnostics when their conclusions diverge. Every sound exact target receives its ordinary located reference and origin evidence even when the rename is rejected, but rejected targets do not alter field state, rewrite facts, or transitions. **Rationale:** Public source refinement remains target-aware, query-only requirements remain target-independent, and later consuming references keep the same final `ref-N` identities in both results.
- **Decision:** Expand wildcard rename targets separately against the public and query-only pre-rename field snapshots, while adding exact targets directly to both destination sets. **Rationale:** Unsupported wildcard substitution invalidates every known matching destination without letting target refinement change query-only evidence.
- **Decision:** Give `finalizeReferences` one explicit `*requirementTrace` parameter and require trace-free helpers to pass `nil`. **Rationale:** The compiler now rejects omitted or multiple traces instead of silently accepting them.
- **Decision:** Sort projector input by finalized `ref-N` order before grouping, retain occurrences in that order, and assign `req-N` only after the first occurrence establishes a group. Order gap candidates by their recorded semantic event and deduplicate only exact code and ordered evidence-link tuples. **Rationale:** Item, occurrence, gap, and reason order remain deterministic without parsing the query or inferring diagnostic ownership from locations.
- **Decision:** Update the two complete SPL2 analysis witnesses during Task 3 even though their files were absent from the initial owned-file list, while leaving `pkg/rewrite/rewrite_test.go:explicitAliasReport` to Task 4. **Rationale:** The exact Task 3 instructions require a green complete analysis package, and these two tests compare the entire current `Result`; the rewrite witness has an explicit later owner and does not indicate a production regression.
- **Decision:** Limit every canonical SPL and SPL2 analysis pass to 4,096 lexer work units, counting each non-EOF token and each lexer error. Detect the event that would consume unit 4,097 and stop before parser prediction. **Rationale:** The value matches existing 4,096-unit schema projection and enumeration budgets, and the recorded dense-query measurements show that transport-size limits alone do not bound parser, lineage, or requirement projection work.
- **Decision:** Treat a normalized over-budget document as an incomplete analysis content result with a nil Go error. Emit no partial canonical evidence and return exactly one `SPL_ANALYSIS_RESOURCE_LIMIT` warning and the corresponding initialized requirement gap. Preserve CLI exit `3`, REST HTTP 200 below the 1 MiB request limit, and owned native and Python success. **Rationale:** Every adapter keeps its existing distinction between admitted content outcomes and request or transport errors while sharing one canonical fail-closed result.
- **Decision:** Add O(1) indexes for requirement-trace pending-reference lookup and incomplete-stage membership as Task 3 hardening. **Rationale:** The security review found that bounding lexer work should not leave trace synchronization with avoidable repeated scans inside the admitted 4,096-unit envelope.
- **Decision:** Count real lexer errors in listener occurrence order before the next returned non-EOF token from the same lexer call. For SPL2 EOF reached within budget, count the existing synthetic unmatched-literal closure error after the last non-EOF token and before EOF, then construct the parser only if all accounting is admitted. **Rationale:** This defines one deterministic order for mixed token and error events and retains the existing ordinary syntax outcome for a within-budget synthetic closure error.
- **Decision:** Pin the resource-limit diagnostic message to `analysis stopped before parser prediction after reaching the 4,096-unit lexer work limit` and the gap message to `requirement coverage is incomplete because analysis exceeded the 4,096-unit lexer work limit`. Use the specified source-indexed half-open rune ranges and let the first omitted event win. **Rationale:** Deterministic cross-surface JSON requires canonical strings and one location rule for returned tokens, real lexer errors, and the SPL2 synthetic closure error.
- **Decision:** A rewrite whose original `PrepareRewrite` session is resource limited returns the specified incomplete no-op report and nil error before rule selection, candidate construction, validation, or verification. Candidate analysis is a deeply detached equal clone made without another analysis call. **Rationale:** Rewrite preserves the content-outcome contract, refuses to act on absent canonical evidence, and retains a complete audit shape for preview, apply, and batch.
- **Decision:** For the ASCII dense 64 KiB and 256 KiB acceptance fixtures, require serialized `RequirementSet` output at most 4,096 bytes and serialized `Result` output at most the full query byte length plus 4,096 bytes. **Rationale:** These fixture-specific checks catch accidental partial evidence amplification while acknowledging that arbitrary JSON escaping and inherited below-budget lineage prevent a universal byte guarantee.
- **Decision:** Wrap the existing generated lexer with a private work-counting lexer and fill its sole `CommonTokenStream` before parser construction. Real lexer listeners charge errors immediately, and the wrapper charges the returned non-EOF token afterward. On overflow it supplies an internal EOF only to terminate preflight; no parser consumes that abandoned stream. **Rationale:** This implements the specified event order and first-omitted-event rule without editing generated ANTLR code, duplicating lexing, or changing admitted token indexes.
- **Decision:** When a real lexer error would consume unit 4,097, the analysis listener records the exact resource-limit location and raises one private sentinel panic. Recover that sentinel only at `lexerWorkLimiter.NextToken`, return the synthetic EOF there, and re-panic every unrelated value. **Rationale:** ANTLR performs error recovery inside the generated lexer's `NextToken` loop, so unwinding to the private wrapper is the smallest way to prevent inspection of later characters without editing generated sources.
- **Decision:** Convert the lexer abort sentinel to synthetic EOF only when the recovering limiter has a non-nil tracker whose `resourceLimit` is non-nil. Re-panic the original value otherwise, including the same package sentinel from another or nested lexer path. **Rationale:** The tracker side effect binds recovery to an overflow recorded by this limiter without introducing a new payload or changing the real error-storm path.
- **Decision:** Build resource-limited requirements by recording the one canonical diagnostic in an otherwise empty requirement trace and running the existing projector, with only the resource-limit gap message specialized. **Rationale:** Query and capability identities, query status, coverage ordering, and detached standalone behavior stay under the same canonical projection path.
- **Decision:** Index pending references by their immutable pending IDs and incomplete stages by stage ID on the shared trace. Rebuild the stage index after parser diagnostic synchronization or SPL2 stage remapping. **Rationale:** Hot lookups are O(1), remapping keeps exact final evidence, and cloned environments retain one fixed trace pointer.

## Outcomes & Retrospective

Tasks 1 and 2 are accepted. Task 3 now includes the public requirement model, query-only trace, deterministic projection, 17-case corpus, 4,096-unit canonical lexer boundary, prompt abort of generated lexer error storms, limiter-local sentinel ownership, exact bounded incomplete result, and O(1) trace indexes. Its focused tests, complete analysis package, race gate, benchmark, and vet pass, and current analysis fixtures remain unchanged. Every Task 3 review through `7d5854a` is superseded until fresh specification, quality, and security reviewers approve the original base through the new head. The complete repository run retains only Task 4-owned rewrite and validation fixture failures. Tasks 4 through 8, whole-feature review, local acceptance, hosted CI, merge, and post-merge verification remain open.

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
8. Requirement diagnostics are canonical query-only diagnostics. Refinement-only diagnostics never enter the embedded or standalone requirement set. The only new requirement-projector codes are `SPL_REQUIREMENT_INDETERMINATE`, `SPL_REQUIREMENT_DYNAMIC`, and `SPL_REQUIREMENT_COVERAGE_INCOMPLETE`; the canonical analysis boundary additionally defines `SPL_ANALYSIS_RESOURCE_LIMIT`.

Canonical analysis admits at most 4,096 lexer work units for either language. Within each lexer call, count real lexer errors in listener occurrence order before counting the next returned non-EOF token. EOF does not count. When SPL2 reaches EOF within budget, run the existing unmatched-literal-mode closure inspection before parser construction. Its single synthetic unterminated-literal error follows the last non-EOF token, precedes EOF, and consumes one unit. If it would consume unit 4,097, stop at the earliest unmatched opener. Do not run closure inspection after an earlier token or real-error overflow. A within-budget synthetic closure error remains an ordinary syntax diagnostic and parsing proceeds. No parser object is constructed until this preflight and closure accounting admit the document.

Detect unit 4,097 before parser prediction and stop at its exact location. First omitted event wins. A returned token uses the existing source-indexed half-open rune range `[token.Start, token.Stop+1)`, clamped by `sourceIndex`. A real lexer error uses `[lexer input index, lexer input index+1)`, clamped so an EOF error is `[len,len)`. The synthetic SPL2 closure error uses the earliest unmatched opener token range `[Start, Stop+1)`.

An over-budget normalized document makes `Analyze` return a non-nil report with a nil Go error. Its document is normalized and retains the full text; status is `incomplete`; syntax and semantic coverage are false with reason `SPL_ANALYSIS_RESOURCE_LIMIT`; stages, scopes, references, lineage, and every dependency array are initialized and empty; and its only diagnostic has code `SPL_ANALYSIS_RESOURCE_LIMIT`, severity `warning`, category `resource_limit`, message `analysis stopped before parser prediction after reaching the 4,096-unit lexer work limit`, empty stage and scope IDs, and the first omitted event location. No evidence from the abandoned parse may survive.

The embedded and standalone requirement sets retain full normalized query identity and capability identity. The query digest covers the full supplied text. Query status is `incomplete`; coverage is false with only reason `SPL_ANALYSIS_RESOURCE_LIMIT`; items are empty; the only gap uses that code, message `requirement coverage is incomplete because analysis exceeded the 4,096-unit lexer work limit`, no reference IDs, and `diagnostic_codes: ["SPL_ANALYSIS_RESOURCE_LIMIT"]`; and the same one diagnostic appears in the set. All collections are non-null arrays. `Requirements` returns the detached standalone set with a nil Go error. This behavior bounds canonical parsing across product surfaces, but it is not a universal serialized-result bound for inherited lineage amplification below the lexer budget. Long sparse input below the work budget remains admitted.

The CLI command is `spl-toolkit requirements [query]`. It accepts one positional or `--query` value, compatibility selectors, optional `--source-id`, `--format text|json`, and optional `--output`. It does not add batch, file, or stdin input. Text mode prints status, coverage, items, gaps, and diagnostics; JSON mode emits the exact `RequirementSet`. Exit codes are `0` for valid plus complete, `1` for invalid, `2` for request, option, output, or internal failure, and `3` when query status is incomplete or requirement coverage is incomplete.

An over-budget query follows the successful content path for both product operations: CLI exit `3`, REST HTTP 200, owned successful native `SPLResult` values, and decoded Python success values. REST still rejects a body larger than 1 MiB with HTTP 400 before analysis.

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
	pendingReferenceIndex map[string]int
	incompleteStages map[string]struct{}
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
	trace     *requirementTrace
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
func (t *requirementTrace) reference(pendingID string) *requirementTraceReference
func (e *requirementEnvironment) stageIncomplete(stageID string) bool
func (e requirementEnvironment) clone() requirementEnvironment
func projectRequirements(document QueryDocument, trace *requirementTrace) (RequirementSet, error)
```

`queryDigest` hashes `[]byte(text)` and formats `sha256:<hex>`. `capabilityRevision` normalizes selectors through `CapabilitiesFor`, marshals the returned typed manifest with `json.Marshal`, and hashes those exact bytes. `projectRequirements` consumes only the normalized document and finalized query trace. If implementation discovers that an extra argument is necessary, record the reason before changing the signature and prove the argument cannot expose target-refined evidence.

`requirementTraceReference.reference.ID` remains pending until `finalizeReferences` obtains the same pending-to-public `ref-N` map used by the public result. `finalizeReferences` remaps trace references and explicit diagnostic links, then asserts that each trace reference has the same canonical identity, kind, role, location, stage, and scope as its public counterpart. Pending-reference lookup and incomplete-stage membership use O(1) indexes maintained with their slices and completeness flags. Environment clones retain the shared trace pointer and therefore see the same indexes. Trace field state clones wherever the existing `environment` clones for branches or scopes.

At each semantic event, record public and trace evidence together. Parser diagnostics seed both public diagnostics and query trace diagnostics. Normal semantic diagnostics record their `incomplete` classification and explicit pending reference owners in the trace. Refinement-only diagnostics use a separate helper and never enter the trace. The trace records baseline wildcard, indeterminate, source, field, knowledge-object, and macro-expansion uncertainty before target refinement can resolve or alter public facts.

The canonical entry point wraps SPL and SPL2 lexing with the same 4,096-unit tracker. For each lexer call, the tracker counts listener errors in occurrence order before the returned non-EOF token. After admitted SPL2 EOF, it performs and accounts for the existing synthetic unmatched-literal closure inspection. It captures the first event that would consume unit 4,097 and returns the bounded resource-limit result before constructing a parser object. That early result initializes every collection and computes capability and full-text query digests without constructing or retaining partial canonical evidence.

The projector runs only after reference finalization for admitted documents. It derives query status and completeness from the trace, builds ordered items and gaps, copies query-only diagnostics, and stores the result on `Result.Requirements`. It does not inspect AST nodes, call `Analyze`, invoke the parser, rerun transfers, or infer diagnostic ownership from locations. `Requirements(document)` calls `Analyze(document)` once and returns `cloneRequirementSet(result.Requirements)`. `document.New(result, context RevisionContext)` uses an equivalent package-local deep clone when it builds `Snapshot`; mutation tests must prove every nested slice is detached.

## Orchestration and review protocol

The orchestrator starts each implementation task only after the preceding task and its reviews are accepted. Fresh subagents receive the approved design specification, this plan, the exact current task, the starting and ending SHAs, and a statement that `_build_plan/` is planning input only. They must use `superpowers:test-driven-development` and must not change files owned by later tasks unless a failing dependency makes the listed scope impossible.

For every Task 1 through Task 8:

- [ ] Record the concrete 40-character task base SHA and assign one fresh implementation agent. The agent writes the named failing test first, runs it to observe the expected failure, makes the smallest production change, reruns the focused test, then runs the task verification commands. It updates the living sections and creates the task's scoped commit.
- [ ] Record the concrete implementation SHA and assign a fresh specification reviewer. The reviewer reads the approved specification and reviews the exact `TASK_BASE_SHA..TASK_HEAD_SHA` range for missing, extra, or contradictory behavior. The reviewer does not edit.
- [ ] Only after specification approval, assign a different fresh quality reviewer. The reviewer examines the same concrete immutable range for correctness, robustness, deterministic behavior, error paths, security, memory ownership, architecture, test quality, and unnecessary scope. The reviewer does not edit.
- [ ] If either reviewer finds an actionable issue, assign a fresh fix agent with the exact findings. The fix agent uses TDD where behavior changes and makes a new scoped commit. Repeat both reviews over the original concrete base SHA through the new concrete head SHA until both approve.
- [ ] Update `Progress`, `Surprises & Discoveries`, and `Decision Log` with evidence and accepted deviations. Release the next task only after the current range is approved.

Do not run two code-writing agents concurrently in this worktree. Review agents may inspect immutable ranges concurrently only when neither writes. No agent may amend another agent's commit, rewrite history, stash user changes, or reset the worktree.

The reopened Task 3 adds a fresh security review after specification and quality approval. All three reviewers inspect the original Task 3 base through the same new head. Any finding reopens all three reviews, and Task 4 remains blocked until they approve one exact range.

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

**Owned files:** `pkg/analysis/model.go`, `pkg/analysis/analyze.go`, `pkg/analysis/parse.go`, `pkg/analysis/spl2_parse.go`, new `pkg/analysis/resource_limit.go`, `pkg/analysis/diagnostics.go`, `pkg/analysis/requirements.go`, `pkg/analysis/requirements_trace.go`, `pkg/analysis/requirements_test.go`, `pkg/analysis/requirements_trace_test.go`, new `pkg/analysis/resource_limit_test.go`, `pkg/analysis/corpus_test.go`, `pkg/analysis/transfers_test.go`, `testdata/requirements/cases.json`, and `testdata/analysis/cases.json`. The parser and trace files were added to Task 3 ownership by the blocking security review; do not edit generated lexer or parser sources.

- [x] Add `TestProjectRequirementsGroupsAndOrdersOccurrences`. Supply a finalized trace with two source reads sharing `(field, host, read, exact)` and one different role. Assert two ordered items, canonical occurrence order, and IDs `req-1` and `req-2`. Run it and record the missing-projector failure.
- [x] Implement `projectRequirements(document QueryDocument, trace *requirementTrace) (RequirementSet, error)` through the empty-set and grouping stages only. Initialize all arrays and omit full query text.
- [x] Add `TestProjectRequirementsFieldPolicy` with subtests for exact source, indeterminate source, wildcard, dynamic, derived, create, output, rename source and target, remove, null test, and unavailable local references. Assert the exact inclusion, necessity, binding, resolution, and gap outcome.
- [x] Add `TestProjectRequirementsKnowledgePolicy` covering `index`, `source`, `sourcetype`, `dataset`, `data_model`, `lookup`, and `macro`, including exact, defensible wildcard/dynamic, gap-only dynamic, and exact macro plus expansion gap.
- [x] Add `TestProjectRequirementsGapOrderingAndDeduplication`. Supply duplicate diagnostic ownership evidence and assert deduplication by code plus ordered links, deterministic message selection, gap order, and first-seen unique `coverage.reasons`.
- [x] Add `TestProjectRequirementsStatusIndependentFromCoverage` for `search host=* | fields - host | table host`. Assert query status `invalid`, coverage complete, no external item for the unavailable local read, and no requirement gap for the local defect.
- [x] Add `Requirements RequirementSet` to `Result` and call the projector once after canonical references and query-only diagnostics are finalized. Add a focused empty-query test proving every successful `Analyze` path emits a non-null requirement object with array-valued collections.
- [x] Implement `Requirements(document)` with exactly one `Analyze(document)` call, unchanged error return, local deep clone, and pointer to the clone. Add `TestRequirementsMatchesEmbeddedAndIsDetached` for every requirement fixture.
- [x] Create `testdata/requirements/cases.json` with full expected sets for valid, invalid, and incomplete SPL and SPL2; exact and derived fields; indeterminate and wildcard fields; all direct knowledge kinds; dynamic identities; macros; removals; null tests; rename source and target; branch and scope evidence; dotted SPL2; non-ASCII text; and empty output.
- [x] Update `testdata/analysis/cases.json` with the additive runtime field. Review a mechanical before/after sample for valid, invalid, and incomplete cases; do not modify historical evidence or receipts.
- [x] Update the hand-pinned full-result JSON strings in `pkg/analysis/transfers_test.go` variable `transferParityWitnesses`. Preserve every pre-Milestone-8 byte except insertion of the correctly derived `requirements` member, and retain the comments that forbid runtime regeneration inside the test.
- [x] Run:

       env GOWORK=off GOCACHE=/private/tmp/spl-toolkit-m8-gocache go test -mod=readonly ./pkg/analysis -run 'TestProjectRequirements|TestRequirements|TestAnalysisCorpus|TestLookupExtractionFullReportParity|TestLocatedTransferEquivalence' -count=1
       env GOWORK=off GOCACHE=/private/tmp/spl-toolkit-m8-gocache go test -mod=readonly ./pkg/analysis -count=1
       git diff --check

   Expected evidence is exact fixture parity, a single analysis invocation, deterministic repeat runs, and a green analysis corpus.
- [x] Update the administrative plan sections with Task 3 evidence, then commit with `git commit -m "feat(analysis): project canonical query requirements"`.
- [x] Reopen Task 3 from original base `d94451ca1ef2e37a8f754a1a60434c11672c0e3a`. Preserve the implementation at `b94814299f9007f7e37e3203c56596d4f0899829`, specification fix at `b60e1ff3d0c524dd866b86bc8af28930ed0881cc`, and projection performance fix at `d1db37cf022bd1cbb640fa5ef0c9bcfc3c65f28b` as historical evidence. Mark every Task 3 approval through `d1db37c` superseded until the new complete range passes fresh specification, quality, and security review.
- [x] Add failing exact-boundary tests for SPL and SPL2. Construct documents with exactly 4,096 lexer work units and require ordinary admission, then add one token or lexer error and require the bounded result. Prove that real lexer errors consume units in listener occurrence order before the next returned token from that lexer call and that first omitted event wins.
- [x] Add SPL2 EOF closure tests. When EOF arrives within budget, require the existing unmatched-literal-mode inspection before parser construction. Count its single synthetic unterminated-literal error after the last non-EOF token and before EOF. If it is unit 4,097, require the resource-limit location at the earliest unmatched opener; if it is within budget, require the ordinary syntax diagnostic and parser path. Prove closure inspection does not run after an earlier overflow.
- [x] Implement one shared 4,096-unit lexer work tracker for both dialects. Do not count EOF and do not construct a parser object until lexer preflight and SPL2 closure accounting admit the document. Do not edit generated ANTLR sources.
- [x] Add `SPL_ANALYSIS_RESOURCE_LIMIT` and a focused full-result test. Require normalized document identity, full supplied text, `status: incomplete`, false syntax and semantic coverage, one coverage reason, initialized empty stages, scopes, references, lineage, and dependency arrays, and exactly one diagnostic with severity `warning`, category `resource_limit`, empty stage and scope IDs, and message `analysis stopped before parser prediction after reaching the 4,096-unit lexer work limit`. Assert that no partial parser, semantic, lineage, dependency, trace, or requirement evidence is emitted.
- [x] Add exact half-open rune-range rows. An omitted returned token uses `[token.Start, token.Stop+1)` clamped through `sourceIndex`; a real lexer error uses `[lexer input index, lexer input index+1)` clamped so EOF becomes `[len,len)`; the synthetic SPL2 closure error uses the earliest unmatched opener `[Start, Stop+1)` range. Require the first omitted event's range in the only diagnostic.
- [x] Extend projector and standalone tests for the bounded result. Require full query and capability identity, a digest over the complete supplied text, incomplete requirement status, false coverage with only the resource-limit reason, empty items, one gap with message `requirement coverage is incomplete because analysis exceeded the 4,096-unit lexer work limit`, no reference IDs, and one matching diagnostic code, the same one diagnostic, non-null arrays, deep detachment, and deterministic repeated JSON.
- [x] Add a long sparse input case whose byte length exceeds the dense witness while its lexer work stays at or below 4,096. Require ordinary parser admission so the work boundary cannot collapse into a hidden byte-size limit.
- [x] Add O(1) indexes for pending-reference lookup and incomplete-stage membership in `requirementTrace`. Add compile-time or focused behavioral tests that require those indexes to stay synchronized through recording, remapping, and stage-completeness updates, and visible through environment clones, without changing existing ordering or serialized output.
- [x] Add an end-to-end adversarial benchmark that analyzes dense wildcard documents, consumes the returned JSON, and reports time, allocations, and output bytes. Include the 65,533-byte witness and a 256 KiB witness; verify both stop at the same 4,097th work event with bounded empty evidence. For ASCII dense fixtures, require serialized `RequirementSet` at most 4,096 bytes and serialized `Result` at most the full query byte length plus 4,096 bytes. Record benchmark measurements without turning those fixture-specific limits into a universal guarantee for arbitrary JSON escaping or admitted below-budget lineage.
- [x] Run:

       env GOWORK=off GOCACHE=/private/tmp/spl-toolkit-m8-gocache go test -mod=readonly ./pkg/analysis -run 'Test.*LexerWorkBudget|Test.*ResourceLimit|TestRequirementTrace|TestProjectRequirements|TestRequirements' -count=1
       env GOWORK=off GOCACHE=/private/tmp/spl-toolkit-m8-gocache go test -mod=readonly -race ./pkg/analysis -run 'Test.*ResourceLimit|TestRequirementTrace' -count=1
       env GOWORK=off GOCACHE=/private/tmp/spl-toolkit-m8-gocache go test -mod=readonly ./pkg/analysis -run '^$' -bench 'Benchmark.*ResourceLimit.*Adversarial' -benchmem -count=1
       env GOWORK=off GOCACHE=/private/tmp/spl-toolkit-m8-gocache go test -mod=readonly ./pkg/analysis -count=1
       env GOWORK=off GOCACHE=/private/tmp/spl-toolkit-m8-gocache go vet ./pkg/analysis
       git diff --check

   Expected evidence is exact-limit admission, plus-one rejection before parser construction or prediction, deterministic event ordering and half-open ranges, SPL2 closure accounting, deterministic bounded output with both canonical messages, full-text digests, lexer-error accounting, sparse-input admission, synchronized O(1) trace indexes, and a green analysis package. The existing Task 4-owned rewrite and current validation/schema fixture failures remain deferred; any other repository regression blocks review.
- [x] Add adversarial SPL and SPL2 cases with 4,095 admitted tokens followed by 32,768 consecutive invalid characters. RED observes all 32,768 listener events. GREEN observes the admitted error and first omitted error only, leaves the lexer cursor at that omitted error, excludes the rejected error from ordinary diagnostics, prevents the later SPL2 unmatched quote from entering the token stream, and preserves the exact resource-limited public result. Add a focused test that requires unrelated panic values to escape the limiter unchanged.
- [x] Add a same-package spoof witness that panics with the real abort sentinel while the limiter's fresh tracker has no recorded limit. RED requires the swallowed panic to escape. GREEN recovers only a sentinel paired with a non-nil `resourceLimit` on that limiter's tracker and preserves the real error-storm abort path.
- [ ] Commit Task 3 security hardening with a scoped implementation commit, update the living administrative sections with only completed evidence, then assign fresh specification, quality, and security reviewers over `d94451ca1ef2e37a8f754a1a60434c11672c0e3a..TASK3_NEW_HEAD`. Do not release Task 4 until all three approve the same exact head.

## Task 4: Refinement and downstream product propagation

**Owned files:** `pkg/analysis/requirements_refinement_test.go`, `pkg/analysis/source_fields_test.go`, focused existing `pkg/analysis/rewrite_*_test.go` files only when their assertions embed full analysis results, `pkg/document/model.go`, `pkg/document/snapshot.go`, `pkg/document/snapshot_test.go`, focused `pkg/validation/*_test.go` files, `pkg/rewrite/rewrite.go`, `pkg/rewrite/batch.go`, `pkg/rewrite/rewrite_test.go`, `pkg/rewrite/requirements_test.go`, `pkg/rewrite/batch_test.go`, `pkg/corpus/requirements_test.go`, `pkg/graph/export_test.go`, `pkg/sarif/export_test.go`, `pkg/impact/compare_test.go`, `internal/lsp/server_test.go`, `testdata/validation/cases.json`, and `testdata/schemas/cases.json`. The blocking review demonstrated that resource-limited original analysis needs a production rewrite propagation fix, which is recorded in `Decision Log`. Other changes to production validation, rewrite, corpus, graph, SARIF, impact, or LSP logic still require a demonstrated propagation bug and a `Decision Log` entry.

- [ ] Add `TestRequirementSetFieldRefinementParity` for plain analysis, `AnalyzeWithSourceFields`, and `AnalyzeWithSourceUniverse`. Use exact, partial, wildcard, unresolved wildcard, and dotted SPL2 cases. Marshal only the embedded sets and assert byte equality.
- [ ] Add `TestRequirementSetValidationRefinementParity` in `pkg/validation` with a field-list target, JSON Schema target, and OCSF target. Compare each report's embedded requirement set with `analysis.Requirements` for the same normalized `QueryDocument`.
- [ ] Add `TestRequirementSetRewriteParity` in `pkg/rewrite/requirements_test.go`. Compare original and candidate analysis requirements with plain analysis of their respective normalized query texts in preview and apply modes.
- [ ] Add over-budget refinement parity for `AnalyzeWithSourceFields`, `AnalyzeWithSourceUniverse`, JSON Schema validation, and OCSF validation. Each path must preserve the exact bounded incomplete analysis and requirement set, with no target-refined or partial canonical evidence.
- [ ] Add `TestRewriteResourceLimitNoOp` for preview and apply. After normalized request preparation, call the original `PrepareRewrite`. When that session is resource limited, stop before rule selection, candidate construction, candidate validation, or verification and return a nil error.
- [ ] Assert the exact no-op result: `schema_version: 1`; normalized document; requested mode; `status: incomplete`; coverage syntax, semantic, and rewrite false; validation omitted; sorted and deduplicated reasons exactly `["SPL_ANALYSIS_RESOURCE_LIMIT", "post_verification_failed"]`; `original_text`, `candidate_text`, and `text` equal the full normalized text; `committed: false`; and an empty changes array.
- [ ] Require one rule evaluation per prepared rule in request order. Every evaluation has the prepared rule ID, outcome `skipped`, reason `post_verification_failed`, empty `reference_ids`, nil location, and nil condition. Preview and apply values are identical except for mode, and apply never commits.
- [ ] Require `original_analysis` to be the exact resource-limited analysis from the prepared session. Build `candidate_analysis` as a deeply detached equal clone without a second `Analyze` call. Omit `candidate_validation`. Mutate every nested collection in each analysis direction to prove no aliasing.
- [ ] Add batch rows with admitted, invalid, and resource-limited documents. Preserve report order and combine statuses through the existing precedence. A resource-limited report uses the same no-op shape, while other reports retain their ordinary behavior.
- [ ] Update `pkg/rewrite/rewrite_test.go` helper `explicitAliasReport` so both hand-built `analysis.Result` values contain exact hand-derived requirement sets. Preserve its independent full-report witness role and do not call `Analyze` from the helper.
- [ ] Add `TestSnapshotRequirementSetDetached`. Call `document.New(result, RevisionContext{ToolVersion: "test", ContractVersion: "v1"})`, mutate every nested requirement slice on the result and snapshot in turn, and assert no alias in either direction.
- [ ] Add `Requirements analysis.RequirementSet` to `document.Snapshot` and a package-local deep clone in `document.New`. Do not serialize through JSON to clone it.
- [ ] Add `TestCorpusEvaluationCarriesRequirements` for the existing `corpus.Evaluation.Analysis` pointer. Assert the corpus aggregate remains unchanged and no requirement aggregate is introduced.
- [ ] Add or extend exclusion tests proving no requirement-specific graph nodes or edges, SARIF rules, impact comparison fields, or LSP messages appear.
- [ ] Update only the current full-result witnesses in `testdata/validation/cases.json` and `testdata/schemas/cases.json`. Inspect `pkg/rewrite/rewrite_test.go:explicitAliasReport` and all `rg -n '\*analysis\.Result|analysis\.Result{' pkg --glob '*_test.go'` hits for equivalent hand-built complete results; update only witnesses whose equality contract covers the complete runtime result.
- [ ] Run:

       env GOWORK=off GOCACHE=/private/tmp/spl-toolkit-m8-gocache go test -mod=readonly ./pkg/analysis -run 'Test.*Requirement.*Refinement|Test.*Requirement.*Rewrite' -count=1
       env GOWORK=off GOCACHE=/private/tmp/spl-toolkit-m8-gocache go test -mod=readonly ./pkg/rewrite -run 'TestRewriteResourceLimit|TestRewriteBatchResourceLimit' -count=1
       env GOWORK=off GOCACHE=/private/tmp/spl-toolkit-m8-gocache go test -mod=readonly ./pkg/document ./pkg/validation ./pkg/rewrite ./pkg/corpus -count=1
       env GOWORK=off GOCACHE=/private/tmp/spl-toolkit-m8-gocache go test -mod=readonly ./pkg/... -count=1
       git diff --check

   Expected evidence is byte-for-byte plain/refined parity, fail-closed over-budget validation and rewrite behavior, and detached propagation through every existing result-bearing product.
- [ ] Update the administrative plan sections with Task 4 evidence, then commit with `git commit -m "feat(products): propagate canonical requirements"`.

## Task 5: CLI and REST adapters

**Owned files:** `cmd/analysis.go`, `cmd/analysis_test.go`, `cmd/cli.go`, `cmd/cli_test.go`, `cmd/main.go`, `pkg/api/analysis.go`, `pkg/api/analysis_test.go`, `pkg/api/server.go`, and `pkg/api/server_test.go`.

- [ ] Add `TestRequirementsCLIInputBoundary` with `requirements 'search host=web'`, `requirements --query 'search host=web'`, duplicate input, `--file`, `--stdin`, and `--batch`. Assert the first two succeed and unsupported acquisition modes exit `2` without a report.
- [ ] Add `TestRequirementsCLIJSONMatchesGo` using one valid SPL and one incomplete SPL2 fixture with selectors and `--source-id`. Decode stdout and compare the complete value with `analysis.Requirements`.
- [ ] Add `TestRequirementsCLIExitCodes` with exact rows: valid plus complete `0`, invalid `1`, incomplete query `3`, valid query plus incomplete requirement coverage `3`, invalid option `2`, output write failure `2`, and an unknown `RequirementSet.QueryStatus` passed directly to the exit helper `2`.
- [ ] Add exact lexer-budget adapter rows for both dialects and both product operations. At 4,096 work units, require ordinary `analyze` and `requirements` CLI and REST content behavior. At unit 4,097, require the bounded analysis result or standalone requirement set, CLI exit `3`, and REST HTTP 200. Require a body larger than 1 MiB to remain REST HTTP 400, and admit a long sparse request below the work budget.
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

   Expected evidence is exact JSON parity with the Go API, stable text, exact process exits, strict transport behavior, shared resource-limit semantics, and no adapter-owned analysis logic.
- [ ] Update the administrative plan sections with Task 5 evidence, then commit with `git commit -m "feat(adapters): expose requirement CLI and REST APIs"`.

## Task 6: Native C ABI and Python API

**Owned files:** `pkg/bindings/bindings.go`, `pkg/bindings/requirements_test.go`, `python/spl_toolkit/mapper.py`, `python/spl_toolkit/libspl_toolkit.h`, `python/tests/test_native_analysis.py`, `python/tests/test_native_requirements.py`, `python/tests/test_native_abi.py`, and `tests/native/memory.c`.

- [ ] Add `TestRequirementsExportReturnsOwnedJSON` in `pkg/bindings/requirements_test.go`. Create a handle, call the missing export with `{"text":"search host=web"}`, decode the result, compare with `analysis.Requirements`, and free it. Record the compile failure before implementation.
- [ ] Add `TestRequirementsExportRejectsBadDocumentsAndClosedHandles` for malformed, array, duplicate-member, unknown-member, invalid UTF-8, and closed-handle inputs. Assert an owned error and nil result, then call `spl_result_free` exactly once.
- [ ] Implement `func spl_mapper_requirements_query(mapperID C.int, documentJSON *C.char) *C.SPLResult` using `ownedMapperJSONResult`, the same strict one-document decoder as `spl_mapper_analyze_query`, and `analysis.Requirements`.
- [ ] Add `test_requirements_matches_canonical_fixture` in `python/tests/test_native_requirements.py` and register the native function as `[ctypes.c_int, ctypes.c_char_p] -> ctypes.POINTER(SPLResult)`. Run the test and record the missing Python method failure.
- [ ] Implement the exact `SPLMapper.requirements_query` signature and delegate through `_validate_fields_request(..., operation="requirements")`. Do not add a Python classifier or fallback.
- [ ] Add Python tests for defaults, all selectors, non-ASCII and NUL text, invalid selectors, lone surrogates, native error mapping, and complete parity with embedded analysis requirements.
- [ ] Add over-budget SPL and SPL2 native and Python cases for `analyze_query` and `requirements_query`. Require an owned successful native result and a decoded Python success value containing the exact canonical incomplete analysis or requirement set. Free the native result exactly once and prove that the resource-limit outcome does not enter native error mapping.
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

   Expected evidence is ABI parity, exactly-once frees, concurrent-call safety, owned resource-limit success, and an idempotently generated header. On Linux with GCC 13, also run the exact `native-memory` commands from `.github/workflows/ci.yml`: build the shared library with Go `-asan`, compile `tests/native/memory.c` with `-fsanitize=address`, and run with `ASAN_OPTIONS=detect_leaks=1:halt_on_error=1`. On other hosts, record the local limitation. The hosted exact-SHA `native-memory` job is a later branch-CI gate, not a prerequisite for local independent acceptance.
- [ ] Update the administrative plan sections with Task 6 evidence, then commit with `git commit -m "feat(native): expose canonical query requirements"`.

## Task 7: JSON Schema, OpenAPI, registries, and source manifests

**Owned files:** `contracts/v1/requirements.schema.json`, `contracts/v1/shared.schema.json`, `contracts/README.md`, `tests/acceptance/test_machine_contracts.py`, `testdata/tooling/contracts.json`, `tools/update_validation_openapi.py`, `tools/tests/test_update_validation_openapi.py`, generated `docs/docs.go`, `docs/swagger.json`, `docs/swagger.yaml`, `python/native-source-files.txt`, `tools/release-content-files.txt`, `tools/release-source-files.txt`, and `tools/tests/test_release.py`.

- [ ] Add a failing `requirements` family entry to `tests/acceptance/test_machine_contracts.py` and a minimal valid report in `testdata/tooling/contracts.json`. Run the family test and record the missing schema failure.
- [ ] Add `analysis.RequirementQueryIdentity`, `analysis.RequirementCoverage`, `analysis.RequirementOccurrence`, `analysis.RequirementItem`, `analysis.RequirementGap`, and `analysis.RequirementSet` definitions to `contracts/v1/shared.schema.json`. Require every defined field, preserve `additionalProperties: true`, and constrain status, necessity, origin, resolution, kind, IDs, and digest formats.
- [ ] Add `contracts/v1/requirements.schema.json` as a root reference to the shared `analysis.RequirementSet` definition.
- [ ] Add `requirements` to the property maps for shared `analysis.Result` and `document.Snapshot`, but do not add it to either existing v1 `required` array.
- [ ] Add negative fixtures only for missing required requirement-set fields, wrong primitive or collection types, invalid enum values, malformed `req-N` and `ref-N` IDs, and malformed digests. Do not expect an unknown output property to fail.
- [ ] Add positive resource-limit analysis and requirement fixtures. Require `SPL_ANALYSIS_RESOURCE_LIMIT`, diagnostic category `resource_limit`, severity `warning`, exact diagnostic message `analysis stopped before parser prediction after reaching the 4,096-unit lexer work limit`, incomplete statuses, false coverage, empty evidence arrays, one gap with exact message `requirement coverage is incomplete because analysis exceeded the 4,096-unit lexer work limit`, no reference IDs, and full digest formats. Keep the existing v1 additive compatibility policy.
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

   Expected evidence is offline schema validation including the bounded resource-limit result, archived-v1 compatibility, deterministic OpenAPI, complete manifest test coverage, and no `_build_plan` path in any runtime or distribution manifest.
- [ ] Update the administrative plan sections with Task 7 evidence, then commit with `git commit -m "feat(contracts): publish requirement interface contracts"`.

## Task 8: Permanent docs, packaged acceptance, and release closure

**Owned files:** `README.md`, `docs/API.md`, `docs/cli.md`, `docs/architecture.md`, `docs/compatibility.md`, `python/README.md`, `tests/acceptance/test_requirements_surfaces.py`, `tests/acceptance/cli_examples.json`, `python/tests/test_native_requirements.py`, `tools/check_package.py`, `tools/check_acceptance.py`, `tools/tests/test_package.py`, `tools/tests/test_acceptance.py`, the fixture-copy and hash plumbing for `testdata/requirements/cases.json`, and the always-allowed administrative files described above.

- [ ] Add `test_requirements_fixture_matches_go_cli_http_and_python` in `tests/acceptance/test_requirements_surfaces.py`. Drive valid, invalid, and incomplete SPL and SPL2 from `testdata/requirements/cases.json`; compare complete decoded values, allowing only JSON object-key order to differ.
- [ ] Add real loopback-server rows for valid content, invalid content, incomplete content, malformed requests, and the 1 MiB boundary. Compare the complete response with the Go fixture.
- [ ] Add 64 KiB and 256 KiB dense adversarial cases for SPL and SPL2 across both Go operations, both CLI operations, both real HTTP operations below the 1 MiB limit, native C, Python, installed wheel, and rebuilt sdist. Require the same bounded incomplete analysis or requirement value, CLI exit `3`, HTTP 200, owned native/Python success, full query digest, one diagnostic and gap, and empty partial evidence.
- [ ] Run concurrent resource-limit acceptance with the ASCII dense 64 KiB and 256 KiB cases. Require deterministic JSON and bounded canonical evidence without hangs, races, result aliasing, or native ownership failures. For each fixture, serialized `RequirementSet` must be at most 4,096 bytes and serialized `Result` must be at most the full query byte length plus 4,096 bytes across repeated and concurrent runs. Include a long sparse input below 4,096 work units and require ordinary admission.
- [ ] Add CLI rows for help, positional and `--query`, text, JSON, `--output`, rejected file/stdin/batch modes, and exits `0`, `1`, `2`, and `3` in `tests/acceptance/cli_examples.json`.
- [ ] Update existing current CLI example outputs that serialize a complete `analysis.Result`. Preserve all unrelated bytes and historical evidence.
- [ ] Add failing `tools/tests/test_package.py` cases asserting the requirement fixture, new native test, and new acceptance test are copied to the out-of-checkout tooling root, named in evidence, and hashed.
- [ ] Extend `tools/check_package.py` constants and copy logic. Set `SPL_REQUIREMENTS_FIXTURES` to an absolute copied path and run requirement tests against both installed wheel and rebuilt sdist without repository imports.
- [ ] Add failing `tools/tests/test_acceptance.py` cases for the new required test files and evidence keys. Extend `tools/check_acceptance.py` without lowering counts, allowing skips, or weakening existing hashes.
- [ ] Update `README.md`, `docs/API.md`, `docs/cli.md`, `docs/architecture.md`, `docs/compatibility.md`, and `python/README.md` with the full report shape, Go and Python examples, CLI and REST examples, digest rules, grouping, source versus derived policy, independent statuses, exit behavior, ABI ownership, refinement parity, additive v1 compatibility, offline boundaries, and exclusions.
- [ ] Document the 4,096-unit lexer work boundary, real-error and token ordering, SPL2 closure accounting, canonical messages and half-open ranges, bounded incomplete result, resource-limited rewrite no-op, cross-surface status behavior, unchanged 1 MiB REST request limit, and long sparse admission. State that the ASCII dense-fixture response limits are acceptance checks, not universal byte guarantees for arbitrary JSON escaping or inherited lineage below the lexer budget.
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

    claude --model opus --print "Review git diff origin/main...HEAD in this repository for Milestone 8 requirement-product specification compliance, correctness, robustness, resource-exhaustion resistance, the 4,096-unit lexer boundary, security, native memory ownership, deterministic behavior, architecture, test adequacy, packaging, and documentation. Verify lexer and SPL2 closure ordering, canonical messages and ranges, no parser construction before admission, the exact rewrite no-op contract, ASCII dense-fixture response bounds, and every cross-surface outcome. Do not modify files. Return actionable findings with file and line evidence, followed by residual risks."

Capture its output and exit status in the orchestrator's out-of-band handoff, not in a tracked file. Do not waive a finding because it came from an external model. Validate it against the approved specification and source. For every confirmed finding, assign a fresh fix agent, require a focused regression test, create a scoped commit, bring the plan and milestone log up to date, then rerun all affected task reviews and all four final reviews. Repeat the Claude command after material fixes. Completion requires no unresolved actionable findings; report disputed findings and their concrete disposition out of band.

When all reviewers approve, declare the reviewed `HEAD` the frozen candidate SHA. From that point through local acceptance, branch CI, merge, and post-merge verification, do not edit or commit the plan, milestone log, or any other tracked file. Report final review conclusions, acceptance results, hosted URLs, merge results, and post-merge results out of band. If a later gate requires a fix, explicitly unfreeze the candidate, make and review the fix, finish all administrative updates, rerun this entire final review gate, and freeze the new reviewed `HEAD`.

Reviewers must explicitly confirm:

1. Query requirements are produced from one canonical traversal and are byte-equivalent across refinement modes and all adapters.
2. Classification, grouping, ordering, gaps, diagnostic ownership, status, coverage, digests, and deep-copy behavior meet every rule in this plan.
3. REST, CLI, C, and Python error and ownership boundaries are exact and do not introduce filesystem, network, injection, unsafe allocation, or use-after-free behavior.
4. Archived v1 analysis and snapshot payloads remain valid, while current runtime payloads always emit requirements.
5. `_build_plan/` is absent from runtime imports, test inputs, package inputs, installed artifacts, release manifests, generated permanent docs, and product documentation links. Tests may retain explicit negative assertions that distributions do not contain it.
6. Historical receipts and evidence were not edited, generated artifacts are reproducible, and the source and release manifests contain exactly the necessary new files.
7. Both dialects use the specified real-error and returned-token order, SPL2 accounts for its synthetic closure error before EOF, and neither constructs a parser before admission. The first omitted event uses the canonical message and half-open range, no partial canonical evidence is emitted, and all Go, CLI, REST, native, and Python outcomes remain exact.
8. Resource-limited preview, apply, and batch rewrites return the exact incomplete no-op shape, never select or commit a rule, preserve report order and status precedence, and deep-clone candidate analysis without a second analysis call.
9. The ASCII dense 64 KiB and 256 KiB fixtures meet their exact serialized response bounds across repeated and concurrent runs. Documentation does not extend those checks into a universal guarantee for arbitrary JSON escaping or inherited below-budget lineage.

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

It must also manually sample at least one valid, one invalid, and one incomplete SPL and SPL2 case and report response JSON, CLI exit, HTTP status, native/Python parity, and requirement coverage. It must exercise exact-limit and plus-one lexer work, real lexer-error order, SPL2 closure accounting, a long sparse admitted document, and concurrent ASCII dense 64 KiB and 256 KiB documents in both dialects. The agent verifies full-text digests, canonical messages and half-open ranges, deterministic JSON, one resource-limit diagnostic and gap, empty partial evidence, CLI exit `3`, REST HTTP 200 below the 1 MiB request boundary, owned native and Python success, and the exact rewrite no-op shape. Each ASCII dense fixture must serialize `RequirementSet` in at most 4,096 bytes and `Result` in at most the full query byte length plus 4,096 bytes. These fixture checks do not claim a universal bound for arbitrary JSON escaping or inherited below-budget lineage. On a non-Linux host, the agent records the local GCC 13 AddressSanitizer limitation without failing local signoff; the hosted `native-memory` job remains a later branch-CI gate. Acceptance fails on any unexplained skip other than that known platform limitation, dirty generated artifact, checkout-relative installed-package read, or unverified required surface. Report the signoff out of band and do not edit tracked files. A failure unfreezes the candidate and restarts the fix, task review, administrative update, final whole-feature review, freeze, and local acceptance sequence.

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
6. Exact-limit, plus-one, real-error ordering, SPL2 closure, full-digest, range, long-sparse, deterministic-output, 64 KiB, 256 KiB, and concurrency evidence proves that canonical parsing stops before parser construction or prediction on unit 4,097 and every surface returns the specified bounded incomplete result without partial evidence.
7. Preview, apply, and batch return the exact resource-limited no-op rewrite contract, and apply never commits.
8. For the ASCII dense acceptance fixtures, serialized `RequirementSet` is at most 4,096 bytes and serialized `Result` is at most full query bytes plus 4,096 bytes across repeated and concurrent runs. Permanent docs limit that claim to those fixtures because arbitrary JSON escaping and inherited below-budget lineage remain outside a universal byte guarantee.
9. Permanent docs, source manifests, content manifests, and release manifests are complete; historical evidence is unchanged; local gates, independent reviews, Claude Opus review, independent acceptance, feature-branch CI, and post-merge `main` CI all pass at the exact accepted SHA.

## Idempotence and recovery

All generators, fixture updates, formatters, and manifest checks must be safe to rerun and produce no second-run diff. Use fixed cache directories under `/private/tmp` only for build cache, never for source worktrees. Test and package tools must create their own isolated output directories and clean them through their existing mechanisms.

If a focused test fails, retain the failing command and first relevant error in `Surprises & Discoveries`, fix the smallest owning layer, and rerun the focused command before broadening. If generated output drifts, regenerate from its authoritative source, never hand-edit a generated artifact. If a native or installed-package check fails after a partial build, rerun the repository's build target rather than copying ad hoc libraries into the package.

If an implementation agent leaves uncommitted changes, the orchestrator inspects and assigns ownership before continuing. Do not reset or discard them. If a review or CI fix changes behavior, unfreeze the candidate, reopen both specification and quality review for the complete task range, finish administrative updates, rerun final review, and rerun local acceptance. If branch CI fails after push, add new commits and push normally; never force-push. If `main` advances before landing, stop and re-evaluate ancestry rather than rebasing or merging silently. If the primary checkout has tracked/index changes, an unexpected untracked file, a missing roadmap input, or an incoming collision, stop and ask the user to resolve it. Do not treat the expected untracked roadmap baseline as dirty, and do not restore the legacy Markdown or text files the user removed.

The analysis fixture, requirement fixture, generated OpenAPI, generated header, contract registry, native source manifest, and release manifests are authoritative checkpoints. Compare them before and after recovery to distinguish a stale build from a semantic change.

## Plan revision note

Created on 2026-09-14 from the approved Milestone 8 design specification and the repository state at `8029a44`. Revised after independent review to add checkbox-sized TDD actions and named helpers, include hand-pinned full-result witnesses, preserve additive output contracts, make OpenAPI and no-dependency checks executable, separate local acceptance from hosted CI, eliminate evidence SHA self-reference, and preserve the primary checkout's exact untracked roadmap baseline during landing. Revised again after the blocking Task 3 security review at `d1db37c` to supersede prior Task 3 approvals, add the 4,096-unit canonical lexer budget and exact resource-limit contract, harden trace indexes, and assign fail-closed and cross-surface checks to Tasks 4 through 8. A subsequent independent plan review pinned lexer and SPL2 closure ordering, canonical messages and half-open ranges, the complete rewrite no-op result, the fixed trace pointer, and fixture-specific response bounds. No post-amendment implementation or review result is recorded yet. Update this note before final whole-feature review whenever execution changes a public shape, task boundary, verification command, or landing procedure, and explain why in `Decision Log`. After freeze, report terminal outcomes out of band rather than editing this file.
