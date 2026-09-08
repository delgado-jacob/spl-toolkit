# Milestone 3 Field-List Validation Implementation Plan


> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development to implement this plan task-by-task after controller release. Progress checkboxes track each task's test, implementation, and review cycle. Do not dispatch implementation while either predecessor checkpoint remains open.

**Goal:** Validate single queries and ordered collections against local allowed-field catalogs through Go, CLI, actual native Python, and REST, with precise matching, missing, unavailable, optional-equivalent, and indeterminate outcomes.

**Architecture:** Extend the existing canonical analysis transfer flow with an optional finite source-name universe and per-wildcard expansion evidence. Add pkg/validation for strict catalogs, membership decisions, reports, and batches. Adapters handle input/output and native ownership without inferring field behavior.

**Tech Stack:** Go 1.22+, the existing ANTLR4 parser/runtime, Go standard-library JSON, existing C-compatible native boundary, Python 3.11+, and the existing stateless REST adapter. No new dependency or runtime service.

**Spec:** _build_plan/milestones/3-field-list-validation/design-spec.md, approved by the controller. This plan repeats its operative contracts so execution can resume from this file alone.

This living ExecPlan follows /Users/jacobdelgado/.codex/PLANS.md. Keep Progress, Surprises & Discoveries, Decision Log, and Outcomes & Retrospective current. The user authorized the controller to make milestone design decisions and run serial subagent implementation with review gates. This document authorizes no production changes before the controller's explicit M3 release.

## Global Constraints


The shared worktree is /Users/jacobdelgado/repos/spl-toolkit/.worktrees/remaining-milestones on codex/remaining-milestones. Work only there. Other agents own later milestone preparation and user-owned untracked planning inputs; preserve their edits and never stage broad directories or revert their changes. The controller has authorized committing this approved plan and design as final M3 task artifacts; the prompt, PRD, and root ledger remain protected.

Go is canonical, parsing is grammar-aware, operation is offline, and adapters contain no field-transfer or matching logic. Preserve accepted plain analysis.Analyze wire output and behavior. M2's separately reviewed wildcard-removal role correctness repair belongs to the predecessor, not a mode-specific M3 behavior. New refinement evidence belongs to the new Go hook and validation report. All report schema_version values are integer 1. Status precedence is invalid, then incomplete, then valid. Locations retain half-open UTF-8 byte offsets and one-based Unicode code-point columns from the kernel. Arrays are [], never null. Field names remain case-sensitive.

No JSON Schema, OCSF, type checks, event checks, broad SPL2, fixes, suppressions, remote catalogs, provider calls, or release publication belongs in this milestone. No production code, runtime test, configuration, or release process may depend on _build_plan.

The plan basename is uniquely milestone-3-implementation-plan.md. Its subagent-development workspace must be derived from this basename, never Milestone 2's implementation-plan.md. Fresh implementers work serially by task. Each implementer completes self-review and tests, stages exact owned paths, and commits the task before one independent reviewer combines specification and code-quality review of immutable BASE..HEAD. Scoped fixes are tested and committed, then the affected immutable range is re-reviewed until clean. Record BASE and HEAD for every review; do not review an uncommitted moving diff or split this gate into separate spec/quality reviewers. Every implementer is told they are not alone and must preserve others' work. The controller controls commits and integration.

## Purpose / Big Picture


A user can supply ["host","actor.user.name"], run validate-fields --fields catalog.json --query 'eval label=host | table label', and receive a valid report because label is created within the query. A query reading an absent external field is invalid with the original source location. Unknown command behavior remains incomplete. Optional declarations do not assert that event records contain those fields. The same report is returned by all four public surfaces, including ordered batches and automation-friendly exits.

## Progress


- [x] (2026-09-07) Controller approved the written M3 design.
- [x] (2026-09-07) Reviewed M2 kernel through e3c468484815ec570df02babb9ec0936dbd7f916 with controller documentation commit 06bc198; this is a planning baseline, not final M2 acceptance.
- [x] (2026-09-07) Controller authorized planning before final adapter acceptance and approved finite-universe expansion evidence.
- [x] (2026-09-07) Checkpoint A: root released execution at accepted M2 closure HEAD 1f5c98605aba26cab402fa46b6605f9e650cf643, product 347d12fb5f35d5a13079c1e0232f7bf529d62198 ancestor; completed log and final adapter/package evidence read.
- [x] (2026-09-07) Checkpoint B: linked worktree/branch and accepted ancestry verified; tracked status clean, copied user planning inputs preserved. Final selector/removal/Unicode helpers reconciled, permanent Go docs corrected to docs/API.md. Root fresh matching executable/82-input/5-artifact evidence reused; no redundant baseline run.
- [x] (2026-09-07) Task 1: RED/GREEN kernel refinement and flow/isolation/plain-analysis regressions complete at 3cb50d1; analysis/mapper race pass. Combined immutable review passed as recorded below.
- [x] (2026-09-07) Task 1: immutable 1f5c986..3cb50d1 independently approved for spec and quality; no findings. Later external-validation/adapter acceptance remains assigned to Tasks2-5.
- [x] Task 2: observe failing strict catalog/report/batch tests, implement canonical validation, and prove fixtures/concurrency.
- [x] Task 2: complete implementer self-review/tests, commit exact owned paths, and pass one independent combined review of immutable BASE..HEAD plus any scoped fix/re-review loop.
- [x] Task 3: observe failing CLI/REST tests, implement transports and documentation, and prove input modes/exits/legacy behavior.
- [x] Task 3: complete implementer self-review/tests, commit exact owned paths, and pass one independent combined review of immutable BASE..HEAD plus any scoped fix/re-review loop.
- [x] Task 4: observe failing real-native/source-closure tests, implement owned Python/native operations, and pass real native tests.
- [x] Task 4: complete implementer self-review/tests, commit exact owned paths, and pass one independent combined review of immutable BASE..HEAD plus any scoped fix/re-review loop.
- [x] Task 5: observe failing installed acceptance wiring, implement cross-surface packaging/docs, and complete local installed acceptance (26 shared validation reports; each wheel 120 native + 26 surface tests, zero failures/skips).
- [x] Task 5: complete implementer self-review/tests, commit exact owned paths, and pass one independent combined review of immutable BASE..HEAD plus any scoped fix/re-review loop.
- [x] (2026-09-07) Final whole-milestone combined review approved immutable 1f5c986..c6ae684; no blocking findings. Exact-source/native/artifact continuity and completed milestone log verified.
- [x] (2026-09-07) Root independent runtime gate passed source 78964a7; later evidence-only documentation changes preserve executable/package/test inputs.
- [x] (2026-09-07) Root accepted Milestone3 COMPLETE at `753bd5a5247830c6afdcef9aee7031c948a77070`; evidence `/private/tmp/spl-toolkit-root-m3-closure-acceptance.json`. This final documentation record does not reopen acceptance. Unrun release-matrix acceptance remains explicitly open.

## Surprises & Discoveries


The reviewed kernel stores source-versus-derived provenance in private trackedField.source; public FieldBinding omits it. Its wildcard selector returns names and one aggregate reference whose binding remains indeterminate even when a closed environment makes membership complete. Evidence is in pkg/analysis/flow.go, commands.go:selector, and references.go:read. Therefore a validator cannot recover each expanded field's binding from public lineage without guessing. The new kernel hook emits explicit evidence while it has that information.

The fields command preserves internal names beginning with underscore, whereas table projects exactly. M2 marks open-source fields inclusion incomplete because retained internal membership is unknown; it does not invent _raw or _time. Finite refinement must retain catalog-backed or known derived internal names, respect tombstones (names explicitly removed earlier), and allow conclusive projection without changing plain Analyze's conservative result.

The planning snapshot's wildcard exclusion used role read while exact exclusion used remove. The controller assigned this defect to M2 for a separately reviewed correction. Consume role remove consistently after handoff; do not introduce refinement-only role classifications.

M2 Task4 introduced shared internal/jsoninput validation because valid raw UTF-8 can still contain unpaired UTF-16 surrogate escapes that encoding/json silently replaces. M3 must reuse that helper, including its valid non-BMP-pair and escaped-backslash behavior. The helper does not by itself enforce duplicate-key or missing-text rules.

M2 adapters and packaging were active during drafting. The final decoder/helper names, native manifests, and acceptance wiring must be reconciled at Checkpoint B. This is why inspected interim adapters are not treated as a final interface contract.

## Decision Log


Decision: String arrays are the easy catalog default; a strict object adds optional_fields, identity, and version. Rationale: matching establishes declaration only, optional-equivalent records explicit optional metadata, and neither implies event presence. Date/Author: 2026-09-07, controller and M3 designer.

Decision: Keep finite-source refinement in the existing kernel and expose a SourceAnalysis wrapper with expansion evidence. Rationale: the kernel alone knows stage/scope flow and each match's source/derived binding. Date/Author: 2026-09-07, controller approval.

Decision: Complete expansion means exhaustive membership and proven binding for every candidate. Rationale: unknown effects and conditional outputs cannot become conclusive through catalog names. Mixed proven bindings can leave an aggregate Reference.Binding indeterminate; only complete per-match evidence resolves it in validation. Date/Author: 2026-09-07, controller approval.

Decision: Explicit nil/empty source names on the refinement hook mean a known empty universe; plain Analyze supplies no universe. Rationale: an empty catalog must differ from unknown membership. Date/Author: 2026-09-07, controller approval.

Decision: Add validate-fields, retain validate, and accept JSON document batches. Rationale: preserve syntax validation and avoid guessing multiline SPL document boundaries. Date/Author: 2026-09-07, controller approval.

Decision: A precise later projection may restore local membership while earlier report coverage stays incomplete; remove selectors are excluded and structural unavailable takes precedence over complete empty source expansion. Rationale: finite-catalog evidence must not override canonical transfer facts or erase earlier limitations. Date/Author: 2026-09-07, controller draft review.

Decision: Reuse fresh matching exact-M2 acceptance evidence, provisioned offline caches, and proactive Go1.22 external linking; commit each self-reviewed/tested task before one combined immutable-range review. Rationale: avoid redundant checks and known loader failures while preserving artifact-backed review. Date/Author: 2026-09-07, controller draft review.

Decision: Planning may overlap M2 adapters, but production and SDD wait for final acceptance. Reuse the final M2 removal-role correction and Unicode helper. Rationale: preparation must not create semantic divergence or duplicate unstable boundary code. Date/Author: 2026-09-07, controller instruction.

## Outcomes & Retrospective


All five serial implementation tasks and independent combined task reviews are complete. Canonical refinement committed at 3cb50d1, validation at aeb71a9, CLI/REST at 8a62cdb, native/Python at 390557b, and installed parity/documentation at 560ae71 with analysis-count correction 78964a7. Evidence-only documentation closure c6ae684 records the independent outer gate. Whole-milestone review of accepted M2 closure 1f5c986 through c6ae684 is specification compliant and quality approved with no blocking findings.

Verification includes the 34-case unchanged analysis corpus and 26 full validation reports; all-package Go race, Go1.22 external-link floor, 79 tooling tests, 137 source Python tests, and 13 documentation pages pass. Both built and rebuilt-sdist wheels pass 120 required native plus 26 surface tests each with zero failures/skips. Root additionally passed Python3.14.1 source137 and eight independent full Go/CLI/REST/native cases, concurrent calls, four-document batch, malformed input, and file/stdin/output checks at78964a7. Documentation-only ancestry is backed by retained source/native/artifact hashes, not repeated unchanged suites.

The bounded implementation reconciliations were explicit initial-command fixture syntax, deterministic strict OpenAPI postprocessing without changing M2 schemas, and real appended CLI example acceptance. No validator flow replay or semantic logic in adapters was introduced. Existing host linker and tar-extraction warning qualifications remain; cross-platform release, live Splunk, event/type validation, JSON Schema/OCSF, and later milestone production are not inferred. Root accepted the complete milestone at `753bd5a5247830c6afdcef9aee7031c948a77070` after independent closure verification. Durable task reports/reviews and ruling history remain in .superpowers/sdd/milestone-3-implementation-plan; permanent product documentation is outside _build_plan.

## Context and Orientation


The repository module is github.com/delgado-jacob/spl-toolkit. pkg/analysis/analyze.go normalizes Query Documents and invokes parsing/semantics. pkg/analysis/scopes.go:analyzeParsed walks typed ANTLR contexts, creates per-scope environments, runs command handlers, snapshots lineage, and finalizes IDs. A scope has independent source availability or inherits a branch environment. A transfer is one command's effect on available names.

pkg/analysis/flow.go defines private environments: known fields, removed-name tombstones, open (other source names may exist), and uncertain (unmodeled effects may alter availability). pkg/analysis/commands.go owns transfers and the '*' selector matcher. pkg/analysis/references.go creates located references, binds reads, and remaps temporary IDs after sorting. Its diagnostic helper marks unsupported effects uncertain; do not emit a wildcard warning and later try to undo those effects.

Public analysis.Result contains schema version, document, status, coverage, stages, scopes, references, lineage, dependencies, and diagnostics. Writes use not_applicable binding; reads use source, derived, unavailable, or indeterminate. Unavailable means an earlier projection/removal prevents a read regardless of catalog declaration. Missing means a source obligation is absent from the catalog.

cmd/cli.go and cmd/main.go route commands and help. pkg/api/server.go registers REST routes/middleware. pkg/bindings/bindings.go owns C-exported results. python/spl_toolkit/mapper.py guards handle lifetime and frees native results. python/native-source-files.txt enumerates Go sources needed to rebuild a source-distribution wheel. tools/check_package.py explicitly selects required native tests and installed acceptance files. Permanent docs live in docs/ and Python's README.

## Interfaces and Dependencies


### Kernel refinement


Create pkg/analysis/source_fields.go with these public Go interfaces. The wrapper is consumed by validation, not exposed as a replacement CLI/REST analysis format:

    type ExpandedField struct {
        Name    string `json:"name"`
        Binding string `json:"binding"`
    }
    type FieldExpansion struct {
        ReferenceID string          `json:"reference_id"`
        Complete    bool            `json:"complete"`
        Matches     []ExpandedField `json:"matches"`
    }
    type SourceAnalysis struct {
        Result     *Result          `json:"analysis"`
        Expansions []FieldExpansion `json:"expansions"`
    }
    func AnalyzeWithSourceFields(document QueryDocument, fields []string) (*SourceAnalysis, error)

Expanded bindings are source or derived. Incomplete expansions assert no matches. Expansion records refer to actual wildcard selectors, sort by final reference order, and use [] when absent. A complete empty expansion is different from absent/incomplete evidence.

Direct hook names must be valid UTF-8, nonempty/non-whitespace-only concrete names without duplicates. Preserve case and meaningful whitespace; do not strip quotes, lowercase, trim valid names, or interpret catalog-side patterns. Clone/sort inputs. Literal stars in source names are concrete characters; query selector matching alone interprets wildcard syntax. Explicit nil/empty means a known empty universe. Query options/text/source identity use existing analysis normalization and errors. Optional metadata and external membership errors never enter this hook.

### Validation model and entry points


Create pkg/validation/model.go. The following shapes specify all public fields and exact JSON names; use the normal snake_case JSON tag for each field named below.

FieldCatalog contains Fields []string (fields), OptionalFields []string (optional_fields), Identity string (identity), and Version string (version). Target contains Kind string (kind) and embedded FieldCatalog, giving a flat kind/fields/optional_fields/identity/version object. Kind is field_list. Coverage contains SyntaxComplete, SemanticComplete, SchemaComplete booleans and Reasons []string. Match contains Name, Binding, Outcome strings. ReferenceOutcome contains ReferenceID, Outcome strings and Matches []Match. Use these exact report interfaces:

    type Report struct {
        SchemaVersion int
        Target        Target
        Analysis      *analysis.Result
        Status        analysis.Status
        Coverage      Coverage
        Outcomes      []ReferenceOutcome
        Diagnostics   []analysis.Diagnostic
    }
    type BatchReport struct {
        SchemaVersion int
        Status        analysis.Status
        Reports       []*Report
    }
    func DecodeFieldCatalog(data []byte) (FieldCatalog, error)
    func Validate(document analysis.QueryDocument, catalog FieldCatalog) (*Report, error)
    func ValidateBatch(documents []analysis.QueryDocument, catalog FieldCatalog) (*BatchReport, error)

Report JSON keys are schema_version, target, analysis, status, coverage, outcomes, diagnostics. Batch keys are schema_version, status, reports. Match keys are name, binding, outcome; outcome keys are reference_id, outcome, matches; coverage keys are syntax_complete, semantic_complete, schema_complete, reasons. All arrays are non-null. Outcomes are matching, missing, unavailable, optional_equivalent, indeterminate. Match outcomes are matching or optional_equivalent.

The Go zero-value catalog is a valid empty set. DecodeFieldCatalog accepts a JSON string array or strict object with required fields and optional optional_fields/identity/version. Reject null, unknown/duplicate object keys, missing fields, trailing JSON, non-string entries, malformed Unicode, empty names, duplicate names, and ordinary/optional overlap. Metadata omitted in JSON becomes empty strings/arrays. Nested names match exactly; parent/leaf declarations do not imply each other. Catalog metadata is not a path/fetch instruction.

Canonical request DTOs and decoders are shared by native and REST:

    type Request struct {
        Document analysis.QueryDocument
        Catalog  FieldCatalog
    }
    type BatchRequest struct {
        Documents []analysis.QueryDocument
        Catalog   FieldCatalog
    }
    func DecodeRequest(data []byte) (Request, error)
    func DecodeBatchRequest(data []byte) (BatchRequest, error)
    func DecodeDocuments(data []byte) ([]analysis.QueryDocument, error)

Their wire keys are document/catalog and documents/catalog. Request decoders read raw catalog JSON through DecodeFieldCatalog so both allowed shapes work. DecodeDocuments serves CLI batches and batch requests. Require each document object to have string text; permit only text/language/profile/version/source_id. Optional fields use analysis defaults. Reject null/non-string properties, unknown/duplicate keys, malformed Unicode, trailing values, and empty document arrays. Empty/whitespace query text becomes syntax-invalid content, not a request error. Reuse final M2 internal/jsoninput Unicode checking; implement remaining validation-specific strict shape/presence rules once in pkg/validation.

Define InputError implementing error and IsInputError(error) bool. Request/catalog/document-option errors are InputError; REST maps them to 400. Unexpected internal errors remain errors/500. Do not catch panics as valid/empty reports. Query errors remain report diagnostics.

## Plan of Work


### Checkpoints A and B — reconcile the accepted predecessor


These are gates. Obtain the final accepted M2 SHA, completed milestone log, adapter/package handoff, and controller M3 release; record them in Progress and Decision Log. Kernel-only acceptance is insufficient. Do not write production tests or start SDD while waiting.

After release run from the shared worktree:

    git branch --show-current
    git rev-parse HEAD
    git status --short
    git log -5 --oneline
    git merge-base --is-ancestor e3c468484815ec570df02babb9ec0936dbd7f916 HEAD

Expect codex/remaining-milestones and ancestry success. Record allowed dirty planning inputs rather than claiming a clean tree. Read final M2 analysis, CLI/native/REST decoders and routes, internal/jsoninput helper, package manifests/checker, and milestone log. Verify wildcard exclusions consistently have role remove; verify retained-internal fields behavior; incorporate final helper signatures and file ownership into this plan. Do not assume interim adapters are final. If accepted interfaces already supply proposed behavior, reuse them. If a contract conflicts, present the exact discrepancy to the controller before dependent work.

Inspect the controller's fresh final-M2 evidence and confirm its exact HEAD and executable inputs match the handoff checkout. Reuse that just-passed evidence when unchanged; do not repeat full acceptance merely to establish a ritual baseline. Run targeted predecessor checks only for missing evidence or identified executable/toolchain drift, then run the required affected checks after M3 changes and final acceptance. Product failures block dependent work; cache/loopback/sandbox denials are reported separately and rerun in the permitted environment. No live service or release matrix is required.

### Task 1 — canonical finite-source refinement


Deliver a Go hook that makes finite-catalog selectors/projections conclusive and returns truthful binding evidence. Own new pkg/analysis/source_fields.go and source_fields_test.go, and narrow changes to analyze.go, flow.go, commands.go, references.go, scopes.go. Consume current QueryDocument/Result/environments/transfers; produce the kernel interfaces above. Do not edit grammar/adapters or weaken M2 expectations.

Step 1: write this red regression in package analysis_test, importing testing and pkg/analysis:

    func TestSourceUniverseRefinesFieldsWithoutChangingAnalyze(t *testing.T) {
        document := analysis.QueryDocument{Text: "fields host* | table hostname"}
        plain, err := analysis.Analyze(document)
        if err != nil { t.Fatal(err) }
        if plain.Status != analysis.Incomplete { t.Fatalf("plain = %s", plain.Status) }
        refined, err := analysis.AnalyzeWithSourceFields(document, []string{"hostname"})
        if err != nil { t.Fatal(err) }
        if refined.Result.Status != analysis.Valid { t.Fatalf("refined = %+v", refined.Result) }
        if len(refined.Expansions) != 1 || !refined.Expansions[0].Complete {
            t.Fatalf("evidence = %+v", refined.Expansions)
        }
        matches := refined.Expansions[0].Matches
        if len(matches) != 1 || matches[0].Name != "hostname" || matches[0].Binding != "source" {
            t.Fatalf("matches = %+v", matches)
        }
    }

Add table-driven cases before each implementation slice: search missing=x with host-only universe remains source and has no unavailable error; initial table host* with an empty universe produces a complete empty source expansion and validates missing, while fields -host | table host* with catalog ["host"] is independently structurally unavailable; fields -absent* is harmless role remove; fields host | table user with host/user universe is unavailable; eval label=host | table lab* proves derived label; eval label=host | table * has mixed host/source and label/derived evidence; | mystery | table * stays incomplete; | mystery | stats count AS total | table total* has conclusive local derived total expansion while the earlier unknown-command diagnostic keeps overall coverage incomplete; | mystery | table host | table absent* proves local structural absence despite earlier incomplete coverage, while | mystery | table host | table host* retains conditional-member uncertainty; lookup assets host OUTPUTNEW label | table lab* stays conditional/incomplete; fields host | table _time with host/_time universe retains _time; removing _time first makes it unavailable; empty catalog invents no internal defaults. Closed projections must not reopen unrelated catalog names. Add search missing=x | fields -miss* | where missing=2 against catalog ["host"]: the first missing read remains an external source obligation, the wildcard exclusion structurally removes that already tracked name even though it is absent from the catalog, and the final missing read is unavailable. The removal reference creates no validation outcome or matching catalog evidence. Add `table missing | table *` with host-only universe: the first exact source obligation remains externally missing and the later selector must not fabricate a matching source member or relabel catalog absence alone as structural absence. Inputlookup resets to a fresh source universe. Appendpipe inherits closure; independent children do not leak derived names. Unsupported merges and wildcard rename/sort/stats semantics remain incomplete.

Step 2: run the focused test and observe missing-hook failure:

    GOCACHE=/private/tmp/spl-toolkit-go-cache GOMODCACHE=/private/tmp/spl-toolkit-remaining-gomodcache GOPROXY=off GOTOOLCHAIN=local go test -mod=readonly ./pkg/analysis -run TestSourceUniverse -count=1

Step 3: add a private per-call refinement context containing an immutable normalized source-name set and pending expansion evidence. Route Analyze through a common helper without a context and the new hook through it with a context. Pass context through analyzeParsed and semanticStage; never attach mutable state to globals or public Result. Initial environments remain structurally open. The same immutable catalog serves fresh independent scopes/inputlookup; table/stats closures stop source enumeration, while appendpipe clones existing flow state.

Refine the existing selector using its matcher. Candidates are currently available known derived fields plus catalog-backed source names allowed by open/closed state and tombstones. A derived binding shadows the same catalog name. Conditional bindings also shadow catalog names and prevent completeness; never replace them with a certain source. For wildcard exclusion, execute the removal transfer over structurally tracked matching names before any finite validation-candidate filtering; a tracked source obligation absent from the catalog is still a name the command removes. Keep those transfer candidates separate from any source-membership evidence used for validation, using the existing canonical selector/transfer rather than validator replay. The remove-role reference is excluded from validation and must not invent a matching catalog declaration. Materialize only requested source candidates and retained internal projection fields, preserving compact state and source provenance. Missing exact source obligations stay as references for external validation; the hook must not invent them as existing wildcard members.

Use this control flow within the existing selector, with focused private helpers in source_fields.go:

    names, bindings, exhaustive := s.refinedSelectorCandidates(pattern)
    complete := exhaustive && !s.env.uncertain && allBindingsProven(bindings)
    if !complete {
        s.recordExpansion(id, false, nil)
        s.diagnostic(CodeUnresolvedWildcard, reason, selectorContext)
    } else {
        s.recordExpansion(id, true, sortedExpandedFields(names, bindings))
    }

These helper names describe the required local algorithm; define their concrete types where used. allBindingsProven means every candidate is nonconditional and source/derived. The helper consumes parsed selector/environment state and never parses query strings. Complete zero-match source selection is external missing evidence, not initial structural unavailable. Sound prior closure/removal may instead prove unavailable, but only when the structural environment independently excludes the selector: inspect structural candidates before filtering source names by the finite universe. A known exact source obligation absent only from the catalog is not a tombstone or proof of structural absence. Exclusion selections never create missing/unavailable obligations from an empty set. Do not issue a warning and later delete it: avoid the warning only when the exact uncertainty is proven resolved.

Complete mixed bindings may leave the aggregate wildcard reference indeterminate; the sidecar resolves each member. Unsupported wildcard operations retain their own diagnostics even if their membership can be enumerated. Unknown effects prevent conclusive membership while the current environment remains uncertain. A later exact table/stats projection may restore local membership precision; still require proven per-match binding, and retain all earlier report-level coverage limitations and diagnostics. In fields inclusion, retain catalog-backed/known derived underscore names and tombstones, without inventing defaults; avoid only the old open-internal-membership warning when the supplied universe proves it. Plain Analyze behavior remains untouched except already accepted M2 corrections.

Step 4: remap expansion IDs with the same final reference mapping in finalizeReferences, sort records by final reference order and matches lexically, and return independent slices. Test Unicode/CRLF source spans, child scopes and deterministic remapping, invalid/direct-Go name inputs, duplicates, nil/empty universe, caller slice preservation, repeated JSON equality, and concurrent calls. Keep existing plain corpus JSON unchanged.

Step 5: run race-enabled analysis/mapper tests and implementer self-review, especially unknown/conditional/closed state. Record BASE, stage exact owned paths, and commit feat: refine analysis with finite source field evidence under controller coordination. Obtain one independent combined spec/quality review of immutable BASE..HEAD. Test and commit scoped fixes, then re-review the affected immutable range until clean.

### Task 2 — strict catalogs, canonical reports, and batches


Deliver complete Go validation. Own new pkg/validation/model.go, catalog.go, request.go, validate.go, batch.go and focused tests; create testdata/validation/cases.json and pkg/validation/corpus_test.go. Consume AnalyzeWithSourceFields and canonical types; produce all validation interfaces above. Do not replay transfers or add adapter/grammar code.

Step 1: write this red test in package validation_test:

    func TestValidateDistinguishesExternalAndStructuralAbsence(t *testing.T) {
        catalog := validation.FieldCatalog{Fields: []string{"host", "user"}}
        external, err := validation.Validate(analysis.QueryDocument{Text: "search missing=x"}, catalog)
        if err != nil { t.Fatal(err) }
        if external.Status != analysis.Invalid || external.Outcomes[0].Outcome != "missing" {
            t.Fatalf("external = %+v", external)
        }
        structural, err := validation.Validate(analysis.QueryDocument{Text: "fields host | table user"}, catalog)
        if err != nil { t.Fatal(err) }
        last := structural.Outcomes[len(structural.Outcomes)-1]
        if structural.Status != analysis.Invalid || last.Outcome != "unavailable" {
            t.Fatalf("structural = %+v", structural)
        }
    }

Add the Task 2 regression search missing=x | fields -miss* | where missing=2 with FieldCatalog{Fields: []string{"host"}}. Assert exactly two validation outcomes in reference order: initial missing and final unavailable; assert no outcome for the remove-role wildcard, SPL_UNKNOWN_FIELD at the initial read, SPL_UNAVAILABLE_FIELD at the final read, and no fabricated matching evidence. Run go test -mod=readonly ./pkg/validation -run TestValidateDistinguishes -count=1 with the task cache and observe failure. Add decoder cases for [], {"fields":[]}, optional entries, metadata, Host versus host, nested exact paths, meaningful spaces, duplicate names/keys, overlaps, nulls/wrong types, missing fields/text, extra keys, trailing values, invalid UTF-8/surrogates, and valid non-BMP/escaped-backslash controls.

Step 2: implement catalog/request decoders with standard JSON and the final internal/jsoninput Unicode validator. Add validation-specific duplicate-key detection and required/allowed-key/type checking once in request.go; do not reimplement Unicode normalization or duplicate it in adapters. Normalize catalogs to sorted independent copies, checking direct Go inputs equally. Union ordinary/optional names for the hook. Missing optional catalog keys normalize to empty values; ordinary fields do not imply required-event semantics.

Step 3: compose reports by traversing refined references and looking up expansion evidence. Include consuming field references with source/derived/unavailable/indeterminate binding, excluding role remove, not_applicable writes, and non-fields. Do not treat every output/rename reference as a read. Exclude removals first, then honor independently proven unavailable binding before consulting wildcard expansion evidence. Only after those checks may complete expansion refine an otherwise aggregate-indeterminate wildcard. Incomplete or absent evidence remains indeterminate and cannot be repaired by the validator. For exact source reads check exact catalog membership; derived reads match without catalog entries. Complete expansion must never override unavailable.

Ordinary source matches yield matching; optional source matches yield optional_equivalent. All-optional wildcard matches yield optional_equivalent; any ordinary/derived member makes aggregate matching while preserving individual outcomes. An empty complete source expansion yields missing only after removal exclusion and the structural-unavailable check. Initial table host* with an empty catalog is missing; fields -host | table host* with host declared is unavailable. In table missing | table *, catalog absence of the earlier exact source obligation must not itself be promoted into structural absence. Use this executable decision order inside reference traversal:

    if ref.Role == "remove" || ref.Binding == "not_applicable" || ref.Kind != "field" {
        continue
    }
    item := ReferenceOutcome{ReferenceID: ref.ID, Matches: []Match{}}
    if ref.Binding == "unavailable" {
        item.Outcome = "unavailable"
        report.Outcomes = append(report.Outcomes, item)
        continue // Preserve the kernel diagnostic; do not add SPL_UNKNOWN_FIELD.
    }
    if ref.Resolution == "wildcard" {
        expansion, found := expansions[ref.ID]
        switch {
        case !found || !expansion.Complete:
            item.Outcome = "indeterminate"
        case len(expansion.Matches) == 0:
            item.Outcome = "missing"
        default:
            // Classify each proven source/derived match, then aggregate.
        }
        // Add the matching outcome entries or located diagnostic for item.Outcome.
        report.Outcomes = append(report.Outcomes, item)
        continue
    }
    // Classify exact source/derived/indeterminate binding.

These comments specify report assembly at the stated branch; implement them with the report types above, not no-op branches. Matching nested names does not declare parents/children. No event presence/type claims are added.

Add report-only SPL_UNKNOWN_FIELD (error/unknown_field) for proven missing source obligations and SPL_INDETERMINATE_FIELD (warning/schema_ambiguity) with a concrete reason for unresolved obligations. Reuse existing SPL_UNAVAILABLE_FIELD without duplication. Copy analysis diagnostics and append located validation findings; sort by source start/code/message and deduplicate only identical findings. Do not mutate embedded analysis findings/status. Syntax/semantic coverage come from refined analysis; schema completeness is false when the full requirement set cannot be established or any obligation is indeterminate. Reasons are ordered distinct combined codes. Compute status as:

    status := analysis.Valid
    if !coverage.SyntaxComplete || !coverage.SemanticComplete || !coverage.SchemaComplete {
        status = analysis.Incomplete
    }
    for _, diagnostic := range diagnostics {
        if diagnostic.Severity == "error" { status = analysis.Invalid }
    }

A definite missing field can be conclusive, and an invalid report can also have incomplete coverage. Include normalized target metadata and arrays for reproducibility.

Step 4: implement nonempty ordered serial batches, normalizing the catalog once. Return no partial batch on invalid options/request/internal errors; malformed SPL text remains a per-document invalid report. Aggregate invalid beats incomplete beats valid. Do not add concurrency scheduling or streaming.

Step 5: create shared fixtures with schema_version 1, source documents, catalogs, and independently reviewed full expected reports. Write semantic expected outcomes/spans before accepting golden reports. Cover exact/nested/optional/derived, empty catalog with eval answer=1 | table answer valid, missing names, zero-match inclusion/removal, mixed bindings, closed/removed/internal fields, unknown commands/functions/macros, conditional lookup, independent/inherited scopes, syntax errors, Unicode/CRLF, and invalid-plus-incomplete.

Step 6: run go test -mod=readonly -race ./pkg/validation ./pkg/analysis with task cache, including repeated/concurrent equality and caller-input isolation. Complete implementer self-review and commit exact owned paths as feat: validate query fields against local catalogs, recording BASE/HEAD. Obtain one independent combined spec/quality review of that immutable range; test/commit scoped fixes and re-review until clean.

### Task 3 — CLI input modes and thin REST routes


Deliver interactive/automation access. Own new cmd/validation_cli.go, cmd/validation_cli_test.go, pkg/api/validation.go, pkg/api/validation_test.go; modify cmd/cli.go, cmd/main.go, pkg/api/server.go, docs/cli.md, docs/api-server.md and final-M2 OpenAPI registration/generated files actually in use; update only the help stdout expectation in tests/acceptance/cli_examples.json. Consume canonical validators and decoders; produce validate-fields and POST /api/v1/query/validate-fields plus /api/v1/query/validate-fields/batch. Reconcile final-M2 helpers first.

Step 1: add runCLIWithInput(args []string, stdin io.Reader, stdout, stderr io.Writer) int for testable stdin, retaining runCLI(args,stdout,stderr) as the existing wrapper supplying os.Stdin. Write this red test before implementation:

    func TestValidateFieldsStdinIdentity(t *testing.T) {
        path := filepath.Join(t.TempDir(), "fields.json")
        if err := os.WriteFile(path, []byte("[\"host\"]"), 0600); err != nil { t.Fatal(err) }
        var out, errOut bytes.Buffer
        code := runCLIWithInput([]string{"validate-fields", "--fields", path, "--stdin", "--format", "json"},
            strings.NewReader("search host=web\n"), &out, &errOut)
        var report validation.Report
        if err := json.Unmarshal(out.Bytes(), &report); err != nil { t.Fatal(err) }
        if code != 0 || errOut.Len() != 0 || report.Analysis.Document.SourceID != "<stdin>" ||
            report.Analysis.Document.Text != "search host=web\n" {
            t.Fatalf("code=%d report=%+v stderr=%s", code, report, errOut.String())
        }
    }

Add exits valid 0, missing/syntax-invalid 1, incomplete 3, usage/input/I/O/write errors 2. Test positional/--query, --file, --stdin, --batch path/'-', text/JSON, --output, source override, invalid Unicode/options, empty/mixed batches, identity/order and byte retention. Reject duplicate/absent fields, competing query sources, --fields '-', and global document flags in batch even if explicitly set to defaults. Keep legacy validate unchanged. Run focused CLI/API tests and observe red.

Step 2: implement a focused parser/input path using final-M2 helpers. Require --fields local file and exactly one source. File/stdin contents are not trimmed; identity defaults to supplied file path, <stdin>, or empty for inline, overridable by --source-id. Reuse --language, --profile, --compatibility-version on single queries. Batches are JSON document arrays with their own options/identities; no global document flags or line splitting. Decode catalogs/documents centrally. Write canonical reports before content exits and do not echo --output reports to stdout. Output-write failure is 2. Text shows status, identity, completeness, located outcomes, useful matches, and diagnostics.

Step 3: add httptest requests and compare complete responses to direct validation. The test request is:

    request := httptest.NewRequest(http.MethodPost, "/api/v1/query/validate-fields",
        strings.NewReader("{\"document\":{\"text\":\"search missing=x\"},\"catalog\":[\"host\"]}"))
    request.Header.Set("Content-Type", "application/json")
    response := httptest.NewRecorder()

Use the final-M2 API test helper/handler entry point established at Checkpoint B to serve it; assert HTTP 200 and whole-report equality with Validate, not just status. Cover both routes, invalid/incomplete content, null/missing/unknown/duplicate keys, invalid options/Unicode, content-type/body limits, and caller isolation. Implement handlers as decode/call/encode only. Content statuses are HTTP 200, InputError is 400, unexpected internal errors are 500. Keep existing method enforcement/middleware. No catalog filename or URL resolution is allowed.

Step 4: update help, CLI/API docs and OpenAPI schemas/routes using the accepted existing documentation mechanism. Include array/object catalogs, file/stdin/batch examples, identities, optional metadata, all exits and transport behavior. Run focused/race tests, existing documented examples, and implementer self-review; commit exact owned paths as feat: expose local field validation through CLI and REST. Record BASE/HEAD and obtain one independent combined spec/quality review of the immutable range, followed by tested/committed scoped fixes and re-review until clean.

### Task 4 — owned native/Python calls and package closure


Deliver actual Python native validation and rebuildable source packages. Own pkg/bindings/bindings.go, python/spl_toolkit/mapper.py, python/spl_toolkit/libspl_toolkit.h, python/native-source-files.txt, tools/check_package.py, tools/tests/test_package.py; create python/tests/test_native_validation.py and update Python docs. Reuse final-M2 owned JSON result helpers and shared Unicode decoder rather than adding competing allocators.

Expose spl_mapper_validate_fields(mapperID C.int, requestJSON *C.char) *C.SPLResult and spl_mapper_validate_fields_batch with the same signature. Verify handle, keep admitted mapper alive, decode centrally, call validation, and return the existing owned result/error. Null requests and invalid handles are errors; invalid/incomplete content is JSON, not an API error.

Python methods are validate_fields(self, query, catalog, *, language='spl', profile='splunkd', version='current', source_id='') -> dict and validate_fields_batch(self, documents, catalog) -> dict. Use the established _operation guard and mapper exceptions. Encode/decode JSON safely, always free SPLResult in finally, including decoding failures, and preserve closed-handle behavior. Python may check serialization safety but does not normalize catalogs or decide outcomes.

Step 1: write this actual-native red case:

    def test_native_validation_matches_created_field():
        with SPLMapper(**mapper_kwargs()) as mapper:
            report = mapper.validate_fields("eval label=host | table label", ["host"])
        assert report["status"] == "valid"
        assert [item["outcome"] for item in report["outcomes"]] == ["matching", "matching"]
        assert report["analysis"]["document"]["text"] == "eval label=host | table label"

Define mapper_kwargs using SPL_NATIVE_LIBRARY only in source mode; installed tests use packaged resolution. Add single/batch statuses, metadata/Unicode, catalog/options errors, malformed raw C JSON, invalid handles, repeat/concurrency/close, and result cleanup after forced JSON decode failure. Run against preceding native output to observe missing method/export; mocks are not acceptance.

Step 2: implement both C exports, public header declarations, and ctypes signatures [c_int,c_char_p] -> POINTER(SPLResult), then wrappers using final M2 helpers. Add all new handwritten analysis/validation sources to native-source-files.txt, preserving predecessor entries. Add the new native test to SDIST_FIXED_FILES and NATIVE_TESTS in tools/check_package.py; update exact closure expectations in tools/tests/test_package.py. First prove manifest tests fail when new sources/tests are omitted.

Step 3: rebuild shared output using Make's version flags, run new and existing real native tests with the Concrete Steps environment, then checker unit tests. Build wheel/sdist and verify every required import source is included. Task 5 completes installed acceptance. Complete implementer self-review and commit exact owned paths as feat: expose native field validation to Python. Obtain one independent combined spec/quality review of immutable BASE..HEAD, then test/commit scoped fixes and re-review until clean.

### Task 5 — installed parity, permanent docs, and handoff


Prove the four surfaces agree and artifacts run outside the checkout. Own tests/acceptance/test_validation_surfaces.py, tools/check_package.py, tools/tests/test_package.py, relevant README.md/docs/API.md/docs/compatibility.md/docs/quickstart.md/docs/architecture.md/python/README.md updates, and _build_plan/milestones/3-field-list-validation/milestone-log.md. Update this plan's living sections. Coordinate semantic fixes with their owners; do not normalize around defects.

Step 1: create installed acceptance using full shared expected reports. Require absolute SPL_VALIDATION_FIXTURES, SPL_CLI, SPL_SERVER. Reuse copied existing acceptance HTTP/server helpers and explicitly import pytest fixtures they require. Tests must load fixture paths through declared environment, never _build_plan. Go corpus tests compare the same expected reports. For each fixture execute CLI with a temporary catalog file and matching document options, call Python, and POST the identical request:

    assert completed.returncode == {"valid": 0, "invalid": 1, "incomplete": 3}[expected["status"]]
    assert json.loads(completed.stdout) == expected
    assert mapper.validate_fields(document["text"], catalog,
        language=document.get("language", "spl"), profile=document.get("profile", "splunkd"),
        version=document.get("version", "current"), source_id=document.get("source_id", "")) == expected
    status, report = post_json(server_url, "/query/validate-fields", {"document": document, "catalog": catalog})
    assert status == 200
    assert report == expected

Only object-key order is irrelevant. Keep array order, exact strings, locations, IDs, lineage, diagnostics and metadata. Add batch parity, file/stdin equivalence with aligned identities, harmless exclusion, internal projection, empty catalog, --output and emitted invalid/incomplete reports. Request-error parity is tested separately from canonical content reports.

Step 2: register this test and fixture copying/environment with final-M2 tools/check_package.py acceptance wiring. Update copied-file/source-closure expectations. Prove the unwired acceptance fails before fixing it. Required installed checks must have nonzero collection and zero failures/skips. Never weaken exact source/fixture manifests or treat uncollected tests as passing.

Step 3: write permanent Go, CLI, Python, REST examples for single/batch workflows, exact input shapes, options, all statuses/exits, optional metadata, exact nested semantics, source identity, and declaration versus structural/event presence. Explain the advanced source-universe hook and complete per-match evidence; ordinary users should call validation. Document unsupported wildcard commands, dynamic constructs and conditional behavior as incomplete. No build-plan filenames belong in runtime examples.

Step 4: run Concrete Steps, inspect required installed counts, and record exact source/artifact identities. Fix canonical/adapter defects at their origin. Run documentation/Go checkers and implementer self-review, then commit exact owned paths as test: verify installed field validation parity. Obtain one independent combined spec/quality task review of immutable BASE..HEAD; test/commit scoped fixes and re-review until clean. After the milestone log is committed, request the whole-milestone combined review of the accepted predecessor SHA through immutable final HEAD. Reuse unchanged fresh executable-input evidence and rerun affected exact-HEAD checks after executable changes.

Step 5: write milestone-log.md beginning with "## What's new in the app" and capability bullets. Follow with built files/APIs/routes, implemented decisions, limitations, exact verification source/counts, next-milestone guidance, and justified deviations. State optionality is declaration-only, M4 must extend canonical validation, and JSON Schema/OCSF have not shipped. No unrun release matrix is claimed. Self-review and commit the exact-owned milestone log/documentation paths before the final immutable review. Controller accepts the exact final SHA before the next milestone.

## Concrete Steps


Run from /Users/jacobdelgado/repos/spl-toolkit/.worktrees/remaining-milestones. Provisioned Python is /private/tmp/spl-toolkit-remaining-venv/bin/python with pinned development dependencies. Use GOCACHE=/private/tmp/spl-toolkit-go-cache, GOMODCACHE=/private/tmp/spl-toolkit-remaining-gomodcache, GOPROXY=off, and PIP_CACHE_DIR=/private/tmp/spl-toolkit-pip-cache for every relevant tool invocation, including focused commands named in prose. Module dependencies are already provisioned; do not reach the network for them. Go floor binary is the explicit 1.22.12 path below. Verify tools at handoff; if missing, follow accepted M2 provisioning without changing tracked dependency files.

Reuse fresh exact-M2 evidence at Checkpoint B when HEAD and executable inputs are unchanged. Run these acceptance commands after M3 changes, and at predecessor handoff only for missing evidence or identified drift:

    GOCACHE=/private/tmp/spl-toolkit-go-cache GOMODCACHE=/private/tmp/spl-toolkit-remaining-gomodcache GOPROXY=off GOTOOLCHAIN=local go test -mod=readonly -race ./...
    GOCACHE=/private/tmp/spl-toolkit-go-cache GOMODCACHE=/private/tmp/spl-toolkit-remaining-gomodcache GOPROXY=off GOTOOLCHAIN=local /private/tmp/spl-toolkit-floor-tools/gomodcache/golang.org/toolchain@v0.0.1-go1.22.12.darwin-arm64/bin/go test -mod=readonly -ldflags=-linkmode=external ./pkg/analysis ./pkg/mapper ./pkg/bindings ./cmd ./pkg/api
    GOCACHE=/private/tmp/spl-toolkit-go-cache GOMODCACHE=/private/tmp/spl-toolkit-remaining-gomodcache GOPROXY=off GOTOOLCHAIN=local make build-all
    GOCACHE=/private/tmp/spl-toolkit-go-cache GOMODCACHE=/private/tmp/spl-toolkit-remaining-gomodcache GOPROXY=off GOTOOLCHAIN=local PIP_CACHE_DIR=/private/tmp/spl-toolkit-pip-cache /private/tmp/spl-toolkit-remaining-venv/bin/python -m pytest tools/tests -q
    GOCACHE=/private/tmp/spl-toolkit-go-cache GOMODCACHE=/private/tmp/spl-toolkit-remaining-gomodcache GOPROXY=off GOTOOLCHAIN=local PIP_CACHE_DIR=/private/tmp/spl-toolkit-pip-cache /private/tmp/spl-toolkit-remaining-venv/bin/python tools/check_go.py
    GOCACHE=/private/tmp/spl-toolkit-go-cache GOMODCACHE=/private/tmp/spl-toolkit-remaining-gomodcache GOPROXY=off GOTOOLCHAIN=local PIP_CACHE_DIR=/private/tmp/spl-toolkit-pip-cache /private/tmp/spl-toolkit-remaining-venv/bin/python tools/check_docs.py

The Go 1.22 floor command deliberately uses -ldflags=-linkmode=external on this macOS host; default-linked tests have the already documented dyld LC_UUID failure before assertions, so do not reproduce that known loader failure. After Task 2 add ./pkg/validation to the floor command. Expect no package failures; record actual counts, not predictions. Loopback tests may require the already permitted execution context; sandbox denial is not a product regression.

Source-mode native tests after build-all:

    GOCACHE=/private/tmp/spl-toolkit-go-cache GOMODCACHE=/private/tmp/spl-toolkit-remaining-gomodcache GOPROXY=off GOTOOLCHAIN=local PIP_CACHE_DIR=/private/tmp/spl-toolkit-pip-cache PYTHONPATH=python SPL_EXPECTED_VERSION=$(cat VERSION) SPL_NATIVE_LIBRARY=/Users/jacobdelgado/repos/spl-toolkit/.worktrees/remaining-milestones/build/libspl_toolkit.dylib /private/tmp/spl-toolkit-remaining-venv/bin/python -m pytest python/tests -q

Built artifacts and installed acceptance:

    GOCACHE=/private/tmp/spl-toolkit-go-cache GOMODCACHE=/private/tmp/spl-toolkit-remaining-gomodcache GOPROXY=off GOTOOLCHAIN=local PIP_CACHE_DIR=/private/tmp/spl-toolkit-pip-cache make python-build PYTHON=/private/tmp/spl-toolkit-remaining-venv/bin/python PIP=/private/tmp/spl-toolkit-remaining-venv/bin/pip
    GOCACHE=/private/tmp/spl-toolkit-go-cache GOMODCACHE=/private/tmp/spl-toolkit-remaining-gomodcache GOPROXY=off GOTOOLCHAIN=local PIP_CACHE_DIR=/private/tmp/spl-toolkit-pip-cache /private/tmp/spl-toolkit-remaining-venv/bin/python tools/check_package.py --sdist dist/spl_toolkit-$(cat VERSION).tar.gz --wheel-dir dist

Confirm no other active milestone process owns these build outputs before using Make's documented artifact replacement. Installed checking must rebuild a wheel from the sdist and load packaged native code outside the checkout, with nonzero required collection and zero skips. Optional source-development skips do not satisfy required installed tests.

For a user-facing demonstration after Task 3, create local temporary files safely:

    /private/tmp/spl-toolkit-remaining-venv/bin/python - <<'PY'
    from pathlib import Path
    import json, tempfile
    directory = Path(tempfile.mkdtemp(prefix="spl-m3-demo-"))
    (directory / "fields.json").write_text(json.dumps(["host", "hostname"]), encoding="utf-8")
    (directory / "query.spl").write_text("eval label=host | table label\n", encoding="utf-8")
    (directory / "batch.json").write_text(json.dumps([
        {"text": "table hostname", "source_id": "good.spl"},
        {"text": "search missing=x", "source_id": "missing.spl"},
        {"text": "| mystery | table host", "source_id": "unknown.spl"}
    ]), encoding="utf-8")
    print(directory)
    PY

Use the actual printed absolute directory with build/spl-toolkit validate-fields --fields <directory>/fields.json --file <directory>/query.spl --format json. Expect exit 0, retained newline/file identity, source host and derived label matching. The --batch variant returns exit 1, ordered valid/invalid/incomplete reports, and the exact missing location. Automated tests use temporary paths and assert those observations.

## Validation and Acceptance


Acceptance requires approved examples and full shared corpus values in Go, CLI, actual native Python, and REST. Wildcard refinement must affect downstream canonical lineage as well as validation matches. Derived names need no catalog entry; exact nested names need their own declaration; optional metadata is valid; empty catalogs are useful; empty batches are errors. Unknown effects, conditional fields, unsupported wildcard commands and incomplete branches retain uncertainty. Definite missing/unavailable/syntax errors win overall status without hiding incomplete coverage.

Automation proves all four exits, emitted invalid/incomplete reports, output files, complete file/stdin bytes, identity overrides, ordered JSON batches, strict malformed input, and unchanged legacy validate. Native acceptance proves ownership, handles, close/concurrency, source closure, and installed loading outside the checkout. Arrays, source spans, messages, IDs and metadata match exactly. No live Splunk success or remote release acceptance is inferred.

## Idempotence and Recovery


Record status before each task and stage only owned files approved by the controller. Do not reset/clean broadly, amend others' commits, or revert predecessor changes to fit this planning snapshot. Preserve failing evidence, fix within ownership, and rerun affected tests. Coordinate cross-owner changes through the controller.

Re-running tests is safe. Native source changes require rebuilding before dependent Python tests so stale libraries cannot create false passes. Keep failed package-checker outputs until understood; remove only disposable task-created outputs when authorized. Preserve user-owned untracked plans. The only request-driven filesystem write is the CLI's explicit output file.

## Artifacts and Notes


Append actual commands, source SHAs, exit codes, and counts at each gate; do not pre-fill success evidence. Final evidence identifies the accepted predecessor, reviewed task boundaries, implementation SHA, platform, native/sdist/wheel paths, installed counts and any unrun release boundary. The milestone log is a temporary initial-build handoff, while permanent behavior docs remain outside _build_plan.

Revision note (2026-09-07): Initial plan drafted under controller authorization to overlap planning with M2 adapters. It selects canonical finite-universe expansion evidence against the reviewed kernel, consumes the final M2 removal-role and Unicode corrections, and retains final acceptance plus adapter reconciliation as hard pre-implementation gates. No production edits or SDD dispatch occurred.

Self-review note (2026-09-07): Clarified that an exact source obligation missing from the finite catalog must not become structural unavailability in a later closed selector; added the explicit table-missing-then-wildcard regression. Public signatures, report keys, shared Unicode ownership, strict boundary rules, and all five task dependencies were checked against the approved spec.


Revision note (2026-09-07, controller draft review): Clarified restoration of local membership after exact projection while retaining global uncertainty; made removal/unavailable/empty-source precedence executable with paired regressions; permitted reuse of fresh unchanged exact-M2 evidence; pinned provisioned offline caches and proactive Go1.22 external linking; corrected all task gates to self-review/tests, exact-path commit, one combined immutable-range review, and scoped fix/re-review. Final-M2 release remains mandatory.

Revision note (2026-09-07, targeted acceptance refinement): Added search missing=x | fields -miss* | where missing=2 against ["host"] to Tasks 1 and 2. Structural wildcard removal must include already tracked names absent from the catalog before validation-candidate filtering; the initial reference remains externally missing, the final read unavailable, and the removal has no validation outcome. This clarifies transfer ordering within the approved kernel/sidecar contract and does not introduce validator flow replay or release production work.

Execution handoff note (2026-09-07): M2 accepted at 1f5c98605aba26cab402fa46b6605f9e650cf643. Actual CLI helpers are cmd/analysis.go:analysisStatusExitCode, coverageLabel, formatAnalysisLocation; REST test entry is Server.Handler().ServeHTTP; shared Unicode API is internal/jsoninput.ValidateUnicode([]byte) error. Native JSON functions use explicit SPLResult allocation/free rather than a generic helper, and Python uses _operation plus try/finally. Reuse those conventions with narrow shared factoring only when new operations require it. Final selectorName handles typed quoted fragments and wildcardMatches gives pattern-star precedence. Permanent Go docs are docs/API.md. No contract discrepancy requires renewed design approval.

Task3 ownership note (2026-09-07): The existing documented-CLI acceptance snapshots full help output. Task3 may update only the help stdout fixture in tests/acceptance/cli_examples.json alongside its new command/help text; preserve every other legacy example and expected value.

## Approved OpenAPI generation reconciliation

Root approves narrow ownership extension to tools/update_validation_openapi.py, its focused tests, and Makefile generate-docs invocation, plus already-owned generated outputs. Pinned swag generates canonical catalog object only, but accepted M3 request permits string array or strict catalog object. Implement deterministic generation postprocessing; no runtime schema mutation or Swag upgrade. Require repeatable/idempotent generation, explicit failure on unexpected pinned output shape, JSON/YAML and served Go-template schema agreement, required/strict keys, and preservation of unrelated schemas. Do not globally harden shared analysis.QueryDocument: strict M3 required-text/unknown-key input needs its own schema to preserve accepted M2 analysis docs. Carry generator dependencies/source closure into later tasks only if actually used.

## Documented CLI example acceptance reconciliation

Extend Task3 ownership narrowly to append new M3 cases in tests/acceptance/cli_examples.json and add optional stdin payload handling in tests/acceptance/test_documented_cli.py if needed. Preserve all legacy cases exactly except approved help stdout. Existing strict coverage markers and bash/sh command fences remain; do not evade coverage by changing fence language. Use existing setup-files support and stable source-id overrides for deterministic outputs. Execute added examples through this existing installed acceptance path. This supersedes prior help-only ownership restriction only for these new cases and minimal input support.

Task5 implementation note (2026-09-07): Required copy registration failed before wiring and passed afterward (29 helper tests). Full tooling79, source Python137, Go race/checker and six-package Go1.22 external-link floor, documentation13 pages, and both actual installed wheel paths passed. Package evidence is `/private/tmp/spl-toolkit-m3-task5-package-evidence.json`; source/artifact identities and final scoped SHA are in task-5-report.md. Final compatibility/log-only edits follow installed checking; executable/package inputs are unchanged. Independent Task5 review, whole-milestone review, root final oracle, and controller acceptance subsequently passed; accepted snapshot `753bd5a5247830c6afdcef9aee7031c948a77070`.

Final review record (2026-09-07): All five task reviews and the broad 1f5c986..c6ae684 review are approved. Root authorized committing only this plan, the approved design, and owned log closure; no prompt/PRD/root-ledger staging. Final documentation closure updates lifecycle/evidence only. Root accepted exact snapshot `753bd5a5247830c6afdcef9aee7031c948a77070`; all milestone gates are complete.

Acceptance record (2026-09-07): Root independently verified accepted M2/tested78964a7/broadc6ae684 ancestry, the two documentation deltas, 167 unchanged prior tracked inputs, 37 native entries, five artifacts, preserved evidence hashes, and clean tracked tree/index before accepting `753bd5a5247830c6afdcef9aee7031c948a77070`. Evidence: `/private/tmp/spl-toolkit-root-m3-closure-acceptance.json`. The final successor commit only records that decision in the owned design/plan/log; no test, build, or review repetition and no reopened acceptance gate.
