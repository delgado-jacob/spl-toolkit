# Structured Analysis Kernel Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking in Progress; the task instructions use numbered steps to honor the ExecPlan rule restricting checklists to Progress.

**Goal:** Give SPL callers a source-aware, flow-aware query report with honest completeness through Go, CLI, real native Python, and REST while retaining legacy behavior.

**Architecture:** Add a stateless public `pkg/analysis` kernel that consumes an additional ANTLR grammar entry point. Typed parse contexts establish references and scopes; command transfer functions update field availability and lineage. Public adapters serialize the same Go result without their own semantic rules.

**Tech Stack:** Go 1.22 minimum, existing github.com/antlr4-go/antlr/v4 v4.13.1, ANTLR generator 4.13.2, existing Go HTTP/C bindings, Python 3.11+ and ctypes. No new runtime service or database.

**Spec:** `_build_plan/milestones/2-structured-analysis-kernel/design-spec.md`, approved by the delegated product decision maker on 2026-09-07, including canonical implicit aggregate names.

## Global Constraints


Go 1.22 minimum. Python 3.11+ minimum. ANTLR establishes syntactic structure; no regex or string-splitting query parser. Runtime analysis is deterministic and offline. Generated parser internals are not public result types. Existing mapper discovery, mapping, and CLI/API behavior remains available. No field-list/schema validation, broad SPL2, new rewriting, batch tooling, SARIF, or editor features. `_build_plan/` is temporary guidance and cannot become a runtime dependency. All new functionality must be exercised through Go, CLI, installed native Python, and REST.

This is a living ExecPlan maintained under `/Users/jacobdelgado/.codex/PLANS.md`. Keep Progress, Surprises & Discoveries, Decision Log, and Outcomes & Retrospective current at each stopping point. The plan embeds the required implementation context; the spec accompanies it for approved requirements. Work only in `/Users/jacobdelgado/repos/spl-toolkit/.worktrees/remaining-milestones` on `codex/remaining-milestones`. Preserve other agents' changes and untracked guidance. Coordinate commits with the controller and stage exact owned paths, never `git add .`.

## Purpose / Big Picture


Today callers obtain flat field names and parser errors without knowing where a name was read, whether an earlier command created it, or what is uncertain. After this work, analyzing `search src=1 | eval a=src, b=a+1 | table b` explains src as a source requirement, a and b as derived bindings, and the final projection. An unsupported command produces an incomplete report retaining known findings. A read after definite removal produces an invalid report with the exact source location. A user can inspect identical report JSON through each supported adapter.

## Progress


- [x] (2026-09-07) Read PRD, milestone prompt/log, CLAUDE.md, PLANS.md, grammar, mapper, CLI, REST, C/Python, and packaging paths.
- [x] (2026-09-07) Delegated decision maker approved architecture, contract, flow semantics, adapters, written spec, and aggregate-name clarification.
- [x] (2026-09-07) Controller verified baseline: full Go race suite, build-all, 69 tool tests, Go 1.22.12 targeted tests, documentation checks, and source Python 27 passed/1 intentional installed-version skip.
- [x] (2026-09-07) Task 1 complete at c57144b after two scoped fix rounds: public model, grammar/source/recovery foundation, combined review and clean re-review; current-Go race and Go1.22 covering checks passed.
- [x] (2026-09-07) Task 2 complete through b430722 after one scoped fix round and clean independent re-review: references, flow, functions/capabilities; full Go race and covering Go1.22 checks passed.
- [x] Task 3: scoped dependencies/recovery, role-specific wildcard resolution, invalid-UTF8 rejection, and 24 full-report corpus cases complete through e3c4684 after malformed-child ownership fix passed independent re-review; race and Go1.22 analysis passed.
- [x] Task 4: CLI/REST analysis and capabilities, strict shared Unicode input validation, generated OpenAPI, and 24-case adapter parity complete at df1bb0b; independent combined review clean.
- [x] Follow-up core gate after Task4: wildcard fields exclusions use role=remove/binding=not_applicable at 829c4c1, with closed/open regressions, 26-case Go/CLI/REST parity, and clean independent scoped review.
- [ ] Task 5: C/Python analysis ownership and native package source closure; red real-library tests, implementation, green package checks, review, commit.
- [ ] Task 6: shared full-report parity, installed-wheel acceptance, docs, final review and milestone log.

## Surprises & Discoveries


The legacy grammar contains a generic `operation` tree, and `pkg/mapper/parser.go` discards a recovered tree whenever an error listener sees an error. Its AST visitor does not preserve nested subqueries. Evidence: `Parser.Parse` returns before `ASTVisitor`, and `VisitInitCommand`/`VisitNextCommand` visit operations only. Consequently the new kernel must use the new parse entry point directly rather than extend the public legacy AST.

Generated sources report ANTLR 4.13.2 while the Go runtime is pinned at v4.13.1; this is the working baseline, not justification for a dependency upgrade. `python/native-source-files.txt` and `tools/check_package.py` explicitly enumerate package inputs/test files, so new source-tree tests alone do not prove an installed wheel includes or exercises the kernel.

Baseline `httptest` initially failed because the sandbox disallowed loopback binding; rerunning with the controller's loopback approval passed. Do not weaken server tests to accommodate that environment. The controller's existing cache emitted nonfatal readonly module-stat-cache warnings.

## Decision Log


Decision: Add `pkg/analysis` and an additional grammar entry rule while preserving legacy entry rules and discovery ABI. Rationale: the structured result must mature without rewriting baseline mapping. Date/author: 2026-09-07, design lead and delegated product decision maker.

Decision: Use UTF-8 byte offsets plus one-based Unicode code-point lines/columns, explicit source/derived/unavailable/indeterminate bindings, and independent syntax/semantic coverage. Rationale: later validation and safe rewriting require exact locations and honest provenance. Date/author: 2026-09-07, same approvers.

Decision: Use versioned direct JSON on all adapters and Go errors only for invalid document options. Query errors remain reports, with invalid winning over incomplete. Rationale: partial discoveries must survive and adapters must not invent semantics. Date/author: 2026-09-07, same approvers.

Decision: Rename evaluates all sources from the before-state; eval assignments run left-to-right; branch merge/OUTPUTNEW uncertainty stays explicit. Rationale: these avoid fabricated dependencies and accidental scope leaks. Date/author: 2026-09-07, same approvers.

Decision: Implicit aggregate names are semantic names such as `sum(bytes)`, not incidental source whitespace or function case. Unknown naming forms are incomplete. Rationale: downstream reads address SPL output fields, while source ranges separately preserve original text. Date/author: 2026-09-07, delegated product decision maker.

## Outcomes & Retrospective


Task 1 foundation is implemented and independently reviewed through c57144b; two grammar-boundary issues were corrected with exact source and legacy mixed-mode regressions. Task 2 semantics is independently reviewed through b430722. Task 3 is independently reviewed through e3c4684 with 24 full-report corpus cases, source-encoding rejection, and malformed-child ownership recovery; Tasks 5–6 remain; the removal-role follow-up is independently reviewed through 829c4c1, with 26 shared cases. Task 4 CLI/REST is independently reviewed at df1bb0b with 24-case parity and shared Unicode input validation. Open-input fields internal-membership uncertainty and unsupported rename overlaps are explicit conservative limitations. Completion requires every task below, meaningful installed-wheel parity, and a milestone log that distinguishes local checks from unrun platform release acceptance. Do not represent the inherited milestone 1 remote evidence as proof for new source.

## Context and Orientation


The repository root contains `grammar/SPLLexer.g4` and `grammar/SPLParser.g4`; committed generated Go lives in `parser/`. Existing analysis-like behavior is `pkg/mapper/discovery_listener.go`; legacy mapper parsing and source-preserving mapping must remain usable. `cmd/cli.go` parses options and returns process exit codes through `runCLI`. Its current `runQueryCommand` returns zero after writing successful payloads, which must be carefully extended to preserve an analysis report's nonzero content status. `pkg/api/server.go` registers routes, `pkg/api/models.go` validates request transport, and `pkg/api/handlers.go` holds existing endpoint logic. Put new handlers in `pkg/api/analysis.go`.

The C-compatible library is `pkg/bindings/bindings.go` (package main) with a concurrency-safe handle registry in `registry.go`. `SPLResult` owns two C strings and is freed by `spl_result_free`. Python `SPLMapper` in `python/spl_toolkit/mapper.py` protects calls using `_operation()` and synchronizes close. New methods must follow those guards and free results even on JSON decoding exceptions. `python/native-source-files.txt` defines the source distribution's Go closure. `tools/check_package.py` owns an explicit source manifest plus native and adapter acceptance file lists. The Makefile builds CLI, server, native library, wheel, and source distribution.

A scope is a pipeline's field environment. A stage is one command within it. A binding records whether a reference comes from the source, a prior derived field, a provably absent field, or uncertainty. A transfer function changes that environment according to a command's parsed arguments. An open environment permits unknown source names; a closed environment describes a projection's complete set. Neither asserts external schema existence. Lineage records which read references contribute to a created output, and links derived reads back to their creator IDs.

## Interfaces and Dependencies


Create `pkg/analysis/model.go`. All fields below have snake_case JSON tags exactly matching the spec; no omitempty on arrays. Use string-backed status constants. Positions and locations are values, not pointers. Define:

    type QueryDocument struct {
        Text string `json:"text"`
        Language string `json:"language"`
        Profile string `json:"profile"`
        Version string `json:"version"`
        SourceID string `json:"source_id"`
    }
    type Status string
    const (Valid Status = "valid"; Invalid Status = "invalid"; Incomplete Status = "incomplete")
    type Position struct { Offset int `json:"offset"`; Line int `json:"line"`; Column int `json:"column"` }
    type Location struct { Start Position `json:"start"`; End Position `json:"end"` }
    type Coverage struct { SyntaxComplete bool `json:"syntax_complete"`; SemanticComplete bool `json:"semantic_complete"`; Reasons []string `json:"reasons"` }
    type Stage struct { ID string `json:"id"`; Command string `json:"command"`; Position int `json:"position"`; ScopeID string `json:"scope_id"`; Location Location `json:"location"`; SemanticComplete bool `json:"semantic_complete"` }
    type Scope struct { ID string `json:"id"`; ParentID string `json:"parent_id"`; Kind string `json:"kind"`; StageID string `json:"stage_id"`; Location Location `json:"location"` }
    type Reference struct {
        ID string `json:"id"`; OriginalName string `json:"original_name"`; NormalizedName string `json:"normalized_name"`
        Kind string `json:"kind"`; Role string `json:"role"`; StageID string `json:"stage_id"`; ScopeID string `json:"scope_id"`
        Location Location `json:"location"`; Resolution string `json:"resolution"`; Binding string `json:"binding"`
        OriginReferenceIDs []string `json:"origin_reference_ids"`
    }
    type FieldBinding struct { Name string `json:"name"`; OriginReferenceIDs []string `json:"origin_reference_ids"`; Conditional bool `json:"conditional"` }
    type FieldState struct { Fields []FieldBinding `json:"fields"`; Removed []string `json:"removed"`; Open bool `json:"open"`; Uncertain bool `json:"uncertain"` }
    type Transition struct { Operation string `json:"operation"`; Output string `json:"output"`; InputReferenceIDs []string `json:"input_reference_ids"`; OutputReferenceID string `json:"output_reference_id"`; Conditional bool `json:"conditional"` }
    type Lineage struct { StageID string `json:"stage_id"`; ScopeID string `json:"scope_id"`; Before FieldState `json:"before"`; After FieldState `json:"after"`; Transitions []Transition `json:"transitions"` }
    type Dependencies struct { Indexes []string `json:"indexes"`; Sources []string `json:"sources"`; SourceTypes []string `json:"source_types"`; Datasets []string `json:"datasets"`; Lookups []string `json:"lookups"`; DataModels []string `json:"data_models"`; Macros []string `json:"macros"` }
    type Diagnostic struct { Code string `json:"code"`; Severity string `json:"severity"`; Category string `json:"category"`; Message string `json:"message"`; Location Location `json:"location"`; StageID string `json:"stage_id"`; ScopeID string `json:"scope_id"` }
    type Result struct { SchemaVersion int `json:"schema_version"`; Document QueryDocument `json:"document"`; Status Status `json:"status"`; Coverage Coverage `json:"coverage"`; Stages []Stage `json:"stages"`; Scopes []Scope `json:"scopes"`; References []Reference `json:"references"`; Lineage []Lineage `json:"lineage"`; Dependencies Dependencies `json:"dependencies"`; Diagnostics []Diagnostic `json:"diagnostics"` }
    type Capability struct { Name string `json:"name"`; SyntaxSupported bool `json:"syntax_supported"`; SemanticSupported bool `json:"semantic_supported"`; Limitations []string `json:"limitations"` }
    type CapabilityManifest struct { SchemaVersion int `json:"schema_version"`; Language string `json:"language"`; Profile string `json:"profile"`; Version string `json:"version"`; Commands []Capability `json:"commands"`; Functions []Capability `json:"functions"` }
    func Analyze(document QueryDocument) (*Result, error)
    func Capabilities() CapabilityManifest

Use report-format integer `1` consistently in Result and CapabilityManifest. JSON document defaults normalize before analysis. Unsupported options return `fmt.Errorf("unsupported language %q", value)`, corresponding `profile` or `compatibility version` messages. Define stable public constants for all diagnostic codes in `diagnostics.go`; statuses and code strings must not change between adapters.

## Plan of Work


The first three tasks establish and test the kernel. Task 1 publishes types and parsing/source behavior; Task 2 supplies ordinary pipeline semantics; Task 3 completes scope/dependency/partial-result behavior. Tasks 4 and 5 run serially after Task 3: Task 4 owns CLI/REST, then Task 5 owns bindings/Python/package infrastructure, following the approved controller workflow. Task 6 integrates their results and owns end-to-end fixtures and final docs. Each serial implementer receives an SDD task brief containing the owned paths and required interface/spec excerpts, not the whole plan. One independent reviewer combines specification and code-quality verdicts per task, followed by a broad milestone review. The tool-created `.superpowers/sdd` ledger and this living plan preserve progress. Other workers' edits must be preserved.

### Task 1: Public model and grammar-backed source analysis


Files: create `pkg/analysis/model.go`, `analyze.go`, `parse.go`, `source.go`, `diagnostics.go`, `model_test.go`, `parse_test.go`, and `source_test.go`; modify `grammar/SPLLexer.g4`, `grammar/SPLParser.g4`, and regenerated `parser/` files only. The public interfaces are the types and Analyze signature above. Internal parser data stays inside pkg/analysis. Later tasks consume Analyze and extend command processing; adapters see no parser context.

1. Write contract, source, and parser regression tests first. The following test is the minimum source-preservation contract; add unsupported-option cases and a recursive JSON walk rejecting null collections. A raw query with Unicode before and inside a single-quoted field is necessary to distinguish byte and rune offsets.

       func TestDocumentAndUnicodeSource(t *testing.T) {
           text := "  search host=\"é\"\n| eval 'café'=1  "
           r, err := Analyze(QueryDocument{Text: text, SourceID: "queries/é.spl"})
           if err != nil { t.Fatal(err) }
           if r.Document.Text != text || r.Document.SourceID != "queries/é.spl" { t.Fatal(r.Document) }
           if r.Document.Language != "spl" || r.Document.Profile != "splunkd" || r.Document.Version != "current" { t.Fatal(r.Document) }
           for _, ref := range r.References {
               if text[ref.Location.Start.Offset:ref.Location.End.Offset] != ref.OriginalName { t.Fatal(ref) }
           }
       }

   Add assertions that the eval target starts at the actual byte index of `'café'`, line 2, column 8, and has normalized name café once extraction lands in Task 2. In this task test the source-index helper independently with the same span rather than permitting an empty reference loop as proof.

2. Run `GOCACHE=/private/tmp/spl-toolkit-go-cache GOTOOLCHAIN=local go test -mod=readonly ./pkg/analysis -run 'TestDocument|TestSource|TestParse' -v` and record missing package/types as the expected initial failure.

3. Implement the model and rune-to-byte source index. Normalize defaults, reject unsupported document options, initialize arrays, and attach lexer/parser error listeners. Add `analysisQuery` and typed pipeline, stage, argument, expression, alias/output/group, and subquery rules alongside existing `query`. Retain legacy labeled alternatives/methods. New rules must structurally recognize a comparison, function call, literal, identifier, alias, argument list, macro invocation, and nested query. Use grammar precedence for arithmetic, concatenation, comparisons, and logical operators. Preserve literal search RHS versus identifier expression RHS through the owning command context. The internal control flow is:

       func Analyze(document QueryDocument) (*Result, error) {
           normalized, err := normalizeDocument(document)
           if err != nil { return nil, err }
           result := newResult(normalized)
           parsed := parseDocument(normalized.Text) // owns token stream, tree, source index, diagnostics
           result.Diagnostics = append(result.Diagnostics, parsed.diagnostics...)
           analyzeParsed(result, parsed) // typed parse contexts; extended in Tasks 2 and 3
           finalizeResult(result) // sorts, assigns IDs, remaps lineage IDs, calculates status
           return result, nil
       }

   Define those internal helpers where used; this pseudocode states control flow, not permission to leave no-op stubs. At Task 1 completion a parsed but unmodeled stage must have unsupported-semantics diagnostics. Support empty/bare malformed input with syntax diagnostics. Never claim full semantics until Task 2 provides a handler. Keep lexical additions narrow (single-quoted/Unicode identifiers, macro delimiters, expression punctuation); test that baseline IP/path/wildcard tokens retain legacy meaning. If new tokens affect existing token-kind checks, adapt only the necessary legacy rule alternatives and prove regressions.

4. Regenerate with the exact tool procedure in Concrete Steps. Tests should show parse acceptance of multiple assignments, nested function predicates, IN lists, aliases, lookup OUTPUT/OUTPUTNEW, whitespace/comma field lists, wildcard selectors, and nested branches. Test an unknown command as syntactically represented but semantically incomplete. Preserve syntax errors and intact stages on either side of malformed arguments; deeper uncertainty assertions belong to Task 3.

5. Run analysis parser/source tests and `go test -mod=readonly ./pkg/mapper ./cmd ./pkg/api ./pkg/bindings`. Review the generated diff for legacy entry point removal. Update Progress with red/green evidence, have the combined independent review gate, and coordinate an exact-path commit named `feat: add source-aware analysis document and grammar entry`.

### Task 2: Expression references, field transfers, and capabilities


Files: create `pkg/analysis/references.go`, `flow.go`, `commands.go`, `functions.go`, `capabilities.go`, `flow_test.go`, `references_test.go`, and `capabilities_test.go`; modify `analyze.go`/`parse.go` only for semantic dispatch. Task 2 owns semantic internals, not adapters. It may extend grammar rules and regenerate parser files narrowly when tests show an assigned supported form lacks necessary typed structure; preserve legacy parser regressions. Consumes typed contexts from Task 1 and produces finalized Reference/Lineage/Coverage values through `Analyze`; `Capabilities() CapabilityManifest` becomes complete for the supported subset.

1. Add table-driven tests for these exact representative cases. Assert input reference bindings and final environments, not merely status:

       search host=web | where src=dest
           // host, src, dest are reads; web is not; search and where comparisons differ.
       search src=1 | eval a=src, b=a+1 | table b
           // src source, a derived when b reads it, final only b, complete.
       search a=1 | rename a AS b | where a=2
           // b created from a; final a read unavailable; invalid.
       search a=1 b=2 | fields - a | where a=3
           // a definitely removed; invalid.
       search a=1 b=2 | stats sum(a) AS total BY b | where a>1
           // aggregate total + b only; a unavailable.
       search a=1 | eval a=a+1 | where a>1
           // eval RHS reads old source binding, where reads created binding.
       search a=1 | eval b=if(a>0,lower(c),d)
           // a, c, d reads; if/lower function names and numeric literals are not fields.
       search a=1 | lookup users key AS a OUTPUT name AS user
           // lookup users; local a input, user derived; catalog key/name not event reads.
       search a=1 | lookup users a OUTPUTNEW user | where user=x
           // user conditional/indeterminate; never a fabricated source requirement.
       search a=1 | stats SUM (a) | table 'sum(a)'
           // semantic aggregate output sum(a), located source spelling SUM (a).

   Include case preservation, unary sort direction, supported pure/aggregate functions, argument arity, and command options such as a numeric head limit. Wildcards and unknown functions must not become complete accidentally.

2. Run `go test -mod=readonly ./pkg/analysis -run 'TestFlow|TestReferences|TestCapabilities' -v` and save a meaningful failing assertion before implementation.

3. Walk typed expression contexts recursively. Literals and function-name nodes produce no field reference; function argument expressions do. Handle search selector values as typed dependencies. Define a small internal environment with lookup and transfer methods; do not put semantic branching in Reference marshaling. The minimum binding decision is:

       if known && !field.Conditional { binding = "derived" /* or retained source provenance */
       } else if state.Uncertain || (known && field.Conditional) { binding = "indeterminate"
       } else if removed || !state.Open { binding = "unavailable"; addUnavailableDiagnostic()
       } else { binding = "source" }

   Preserve whether a known input binding originated at the source; being previously observed does not make it derived. Implement eval left-to-right; ordinary non-overlapping rename sources from a snapshot; overlapping/colliding forms incomplete with affected binding claims withheld; exact fields removal and projection with known internal retention/open-membership uncertainty; exact table projection; stats closed output; eventstats/streamstats additive output; lookup aliased input/output and OUTPUTNEW conditional output; inputlookup open source; sort/dedup reads and numeric head/tail pass-through. Unsupported variants report precise unsupported-semantics diagnostics and mark the environment uncertain. Resolve provable wildcards against known fields but mark unresolved membership incomplete. Ensure provenance graph references remain correct after deterministic reference-ID assignment.

4. Implement capability entries from the actual dispatch/function registry so advertised support cannot drift from handlers. `Capabilities` returns fresh slices. Unknown functions report `SPL_UNSUPPORTED_FUNCTION`; dynamic functions/macros report the corresponding unresolved reason. Unknown command effects taint the state instead of behaving as identity. Finalization uses the following outcome precedence and must preserve all diagnostics:

       result.Status = Valid
       if !result.Coverage.SyntaxComplete || !result.Coverage.SemanticComplete { result.Status = Incomplete }
       for _, diagnostic := range result.Diagnostics {
           if diagnostic.Severity == "error" { result.Status = Invalid }
       }

5. Run all `pkg/analysis` tests plus legacy mapper tests. Add mutation tests for capability return-value isolation and repeated JSON determinism. After independent review, coordinate commit `feat: track structured references and pipeline field lineage`.

### Task 3: Nested scopes, dependencies, and honest partial analysis


Files: create `pkg/analysis/scopes.go`, `dependencies.go`, `scopes_test.go`, `recovery_test.go`, `corpus_test.go`, and `testdata/analysis/cases.json`; modify semantic/parser helpers only where needed, including narrowly tested grammar/generated-parser extensions for typed dependency contexts. Consumes public interfaces/transfer behavior from Tasks 1–2 and produces the finished canonical kernel used by all adapters. Shared fixtures are version `"1"` with `cases`, each containing `id`, `document`, and fully reviewed `expected` Result JSON. During development focused assertions may precede full goldens; checked-in expected documents must be reviewed rather than blindly accepted from current output.

1. Write failing tests proving independent versus inherited environments:

       search root=1 | join id [ search child=1 | eval local=child ] | where local=2
           // child reads child; its local does not become a known parent derived field;
           // parent merge incomplete, parent local indeterminate, explicit child scope.
       search root=1 | append [ search child=1 | eval root=child ]
           // independent child source; parent root origin unchanged, merge incomplete.
       search root=1 | eval local=root | appendpipe [ where local>1 ]
           // inherited copy sees local derived; no mutation leaks to original environment.
       search host=web | eval broken= | stats count by user
           // invalid syntax; sound host and user findings remain; no invented broken output.
       search a=1 | fields - a | mystery x | where a=2
           // incomplete unknown effects; later a not falsely definite unavailable.
       search a=1 | fields - a | where a=2 | mystery x
           // invalid from already-proven read and incomplete from mystery both retained.

   Add each dependency kind: `index=main source="/var/log/a" sourcetype=syslog`, `inputlookup users`, baseline datamodel/from/tstats forms, and an unresolved backtick macro between intact stages. Verify exact located dependency names even when field semantics remain incomplete. Reject field names inferred merely from opaque argument text.

2. Run `go test -mod=readonly ./pkg/analysis -run 'TestScopes|TestRecovery|TestCorpus' -v` and record the failures. Then implement explicit scope push/pop with copied or fresh environments, child ownership by parent stage, and conservative merge uncertainty. Parent stage IDs precede child stages; position counters are per scope. Aggregate all scope dependencies without losing their located references.

3. Strengthen grammar-aware error synchronization if needed. Damage intervals from lexer/parser errors must withhold references relying on synthetic inserted/deleted nodes, while intact later pipeline contexts are still analyzed. A recovery strategy can synchronize at a tokenized pipe only when bracket/parenthesis nesting matches the stage; never split source text. Unclosed delimiters may make a remainder unsound; retain earlier sound references and explain that limit rather than guessing recovered structure.

4. Create the reviewed full-report fixture corpus, including Unicode CRLF/tabs, repeated field names, macros, unknown pure-looking functions, wildcard names, comparison variants, functions/multiple assignments, every transfer family, independent scopes, partial recovery, and invalid precedence. Every reference span must round-trip to source bytes. Run corpus calls concurrently and compare encoded JSON bytes for the same document. Negative tests assert no panic for truncated inputs and no complete result for unknown effects.

5. Run `GOCACHE=/private/tmp/spl-toolkit-go-cache GOTOOLCHAIN=local go test -mod=readonly -race ./...` with permitted loopback access. Record syntax and semantic limitations in the manifest rather than reducing assertions to get green. After review, coordinate `feat: preserve scoped dependencies and partial analysis findings`.

### Task 4: CLI and REST report adapters


Files: modify `cmd/cli.go`, `pkg/api/server.go`, and API model/annotation integration; create `cmd/analysis.go`, `cmd/analysis_test.go`, `pkg/api/analysis.go`, `pkg/api/analysis_test.go`, and a narrow shared `internal/jsoninput` Unicode-validation helper with tests; regenerate `docs/docs.go`, `docs/swagger.json`, and `docs/swagger.yaml` when API docs generation produces them (inspect actual generator filenames). Consumes `analysis.Analyze` and `analysis.Capabilities`; produces CLI analyze/capabilities and REST POST `/api/v1/query/analyze`, GET `/api/v1/capabilities`. Own no Python/binding files.

1. Write tests loading `testdata/analysis/cases.json`, invoke `runCLI`, and compare parsed JSON to each expected Result. Assert exits 0/1/3 and report bytes on stdout, empty stderr for successful report delivery, code 2 for unsupported document options, and byte-preserving `--output` without duplicate stdout. REST httptest must compare full response values, 200 for all content statuses, and 400 for malformed JSON/options. Existing discover/map/validate tests stay unchanged.

       var out, errOut bytes.Buffer
       code := runCLI([]string{"analyze", "--query", "search a=1 | mystery x", "--format", "json"}, &out, &errOut)
       if code != 3 || errOut.Len() != 0 { t.Fatalf("%d %s", code, errOut.String()) }
       var report analysis.Result
       if err := json.Unmarshal(out.Bytes(), &report); err != nil { t.Fatal(err) }
       if report.Status != analysis.Incomplete { t.Fatal(report.Status) }

2. Run `go test -mod=readonly ./cmd ./pkg/api -run 'TestAnalysis|TestCapabilities' -v` to prove missing routes/commands fail. Add flags only where permitted; legacy commands reject analysis-only flags. Default option values must come from Go document normalization. `capabilities` requires no query and supports the existing format/output convention.

3. Implement handlers as transport-only delegation. Preserve request body size/content-type protections and existing middleware; validate raw UTF-8 and JSON surrogate pairing with the shared helper before decoding; avoid requiring nonempty Text in transport validation because empty query text is a reportable syntax error. Return direct canonical report JSON. Modify successful CLI output flow to return the analysis status code after writing the payload; keep legacy successful codes at zero. Text output includes status, separate syntax/semantic coverage, located references, and diagnostics.

       report, err := analysis.Analyze(document)
       if err != nil { s.writeErrorResponse(w, http.StatusBadRequest, err.Error()); return }
       s.writeJSONResponse(w, http.StatusOK, report)

4. Annotate and regenerate API documentation, inspect the served spec for both routes and report version. Run `go test -mod=readonly ./cmd ./pkg/api` and legacy documented CLI checks. Coordinate review/commit `feat: expose structured analysis in CLI and REST`.

### Task 5: Real native Python and package source closure


Files: modify `pkg/bindings/bindings.go`, `python/spl_toolkit/mapper.py`, `python/spl_toolkit/libspl_toolkit.h`, `python/native-source-files.txt`, and `tools/check_package.py`; create `python/tests/test_native_analysis.py`; modify `tools/tests/test_package.py` for manifest expectations. Consumes canonical Go interfaces and existing SPLResult/free/registry helpers. Produces C exports `spl_mapper_analyze_query(C.int, *C.char) *C.SPLResult`, `spl_mapper_capabilities(C.int) *C.SPLResult`, Python `analyze_query`/`capabilities` signatures from spec. Own no CLI/REST files.

1. Write real-library tests for valid/invalid/incomplete reports, defaults and explicit options, source bytes/ID, capability manifest, bad options, malformed request JSON at C boundary, invalid/closed handles, repeated calls, concurrent analysis and close, and result cleanup on JSON decode failure. Do not mock away the native path for acceptance. A focused regression is:

       def test_analysis_returns_partial_result():
           with SPLMapper(**mapper_kwargs()) as mapper:
               report = mapper.analyze_query("search host=web | mystery x", source_id="q.spl")
           assert report["status"] == "incomplete"
           assert report["document"]["source_id"] == "q.spl"
           assert any(ref["normalized_name"] == "host" for ref in report["references"])
           assert any(d["code"] == "SPL_UNSUPPORTED_COMMAND" for d in report["diagnostics"])

2. Run the new real-library test against the freshly built baseline library with `SPL_NATIVE_LIBRARY` set and expect missing method/export failure. Implement C functions that resolve and retain a registry handle, decode a document, call Go, and JSON-encode into an owned SPLResult; always initialize both pointers before an error path. Invalid options go in result.error, while syntax errors remain result.result JSON. Use the existing result free function.

3. Add ctypes signatures and wrapper methods. The required ownership shape is:

       with self._operation() as handle:
           pointer = self._lib.spl_mapper_analyze_query(handle, json.dumps(document).encode("utf-8"))
           if not pointer:
               raise SPLMapperError("Native analysis returned no result")
           try:
               if pointer.contents.error:
                   raise SPLMapperError(pointer.contents.error.decode("utf-8"))
               return json.loads(pointer.contents.result.decode("utf-8"))
           finally:
               self._lib.spl_result_free(pointer)

   Reject embedded NUL in text/options before C conversion if using a NUL-terminated JSON transport could otherwise truncate: JSON encoding normally escapes NUL, so test round-trip behavior rather than adding an unnecessary text restriction. Capabilities also uses the operation guard and frees its owned JSON result.

4. Add every new handwritten analysis Go source to `python/native-source-files.txt`; generated parser files retain existing paths. Add `tests/test_native_analysis.py` to `SDIST_FIXED_FILES` and `NATIVE_TESTS` in `tools/check_package.py`. Package manifest tests must fail when an analysis source or native analysis test is absent. Task 6 will add cross-adapter test files and copied fixture paths to `ACCEPTANCE_FILES`/runner environment. Do not satisfy exact-manifest tests by removing new required sources.

5. Rebuild `make build-shared`; run new and existing real-library Python tests. Build a wheel and source distribution using the command below, then run checker tests for source closure. After review, coordinate commit `feat: expose owned native analysis reports to Python`.

### Task 6: Full-report parity, installed-wheel acceptance, and completion


Files: create `tests/acceptance/test_analysis_surfaces.py`; modify `tools/check_package.py`, `tools/tests/test_package.py`, `README.md`, `docs/API.md`, `docs/compatibility.md` if present (otherwise create it), and `python/README.md`; update `testdata/analysis/cases.json` only for independently reviewed expectations; create `_build_plan/milestones/2-structured-analysis-kernel/milestone-log.md`. Update this plan's living sections. Own cross-surface testing and user-facing docs, not semantic changes without coordinating the responsible worker.

1. Use the shared fixture expected report in direct Go corpus tests, CLI tests, native Python, and REST. The new pytest module may import lifecycle/HTTP helpers from copied `test_surfaces.py`; explicitly import fixtures required by pytest rather than relying on incidental discovery. It requires absolute `SPL_ANALYSIS_FIXTURES`, `SPL_CLI`, and `SPL_SERVER`; installed mode loads SPLMapper without library_path. Core comparison is:

       for case in fixture["cases"]:
           document = case["document"]
           with SPLMapper() as mapper:
               actual = mapper.analyze_query(document["text"], language=document.get("language", "spl"),
                   profile=document.get("profile", "splunkd"), version=document.get("version", "current"),
                   source_id=document.get("source_id", ""))
           assert actual == case["expected"], case["id"]
           status, rest = post_json(server_url, "/query/analyze", document)
           assert status == 200 and rest == case["expected"], case["id"]
           args = [str(cli_path), "analyze", "--query", document["text"], "--format", "json"]
           for key, flag in (("language", "--language"), ("profile", "--profile"),
                             ("version", "--compatibility-version"), ("source_id", "--source-id")):
               if document.get(key):
                   args.extend([flag, document[key]])
           process = subprocess.run(args, capture_output=True, text=True, check=False)
           assert process.stderr == ""
           assert process.returncode == {"valid": 0, "invalid": 1, "incomplete": 3}[actual["status"]]
           assert json.loads(process.stdout) == case["expected"], case["id"]

   Implement the CLI invocation concretely using a list argument per value; no shell interpolation. Compare capability JSON across the same surfaces. Only object-key order may normalize; arrays, source bytes, diagnostic messages, IDs, and locations must match.

2. Add `test_analysis_surfaces.py` to `ACCEPTANCE_FILES`; have `check_package.py` copy `testdata/analysis/cases.json` into the outside-checkout acceptance directory and set `SPL_ANALYSIS_FIXTURES` to that absolute copy. Preserve baseline `SPL_FIXTURES`. Include the corpus identity in package evidence wherever fixture hashes are recorded. Tests must prove required analysis tests run without skips from the installed wheel and rebuilt-sdist wheel, not silently bypassed by missing environment variables. Verify source injection is absent in the installed runner's `clean_env` path.

3. Run the local built-wheel package checker and full regression commands below. Update package-tool unit tests for new accepted files. If parity fails, fix the canonical kernel or transport defect; never normalize away mismatches. A failing native ownership, missing export, missing source file, or uncollected required test blocks completion.

4. Document only delivered behavior: Go example, analysis/capability CLI examples and exit codes, REST request/report example, Python context-manager usage, report format `1`, compatibility version current, exact coordinate convention, coverage semantics, structural versus schema validity, supported capability limitations, and migration from flat discovery. `_build_plan/` cannot be a runtime/documentation dependency of these examples. Explain unchanged legacy APIs without promising identical flat and structured field classification.

5. Run final code/spec review and acceptance. Write milestone-log with `## What's new in the app` first, then built files/contracts, decisions, next-milestone guidance, scope deviations, exact local commands/counts, source commit IDs and any open platform acceptance. Coordinate final commit `docs: verify and document structured analysis milestone`. Re-read `git status --short`, `git rev-parse HEAD`, and `git diff --check` immediately before reporting. Do not claim completion while a required local acceptance check is unrun or failing.

## Concrete Steps


Run commands from `/Users/jacobdelgado/repos/spl-toolkit/.worktrees/remaining-milestones`. Existing provisioned tools are task-local: Python `/private/tmp/spl-toolkit-remaining-venv/bin/python` (3.12.6, pinned dev dependencies), Go cache `/private/tmp/spl-toolkit-go-cache`, Java 21, and ANTLR jar `/private/tmp/spl-toolkit-remaining-tools/antlr-4.13.2-complete.jar`. Verify the JAR's SHA256 is `eae2dfa119a64327444672aff63e9ec35a20180dc5b8090b7a6ab85125df4d76` before generation. If missing, obtain exactly `https://www.antlr.org/download/antlr-4.13.2-complete.jar` into that task-local path; it is a build tool, never a runtime network dependency or vendored binary.

    shasum -a 256 /private/tmp/spl-toolkit-remaining-tools/antlr-4.13.2-complete.jar
    java -jar /private/tmp/spl-toolkit-remaining-tools/antlr-4.13.2-complete.jar -Dlanguage=Go -package parser -visitor -listener -Xexact-output-dir -o parser grammar/SPLLexer.g4
    java -jar /private/tmp/spl-toolkit-remaining-tools/antlr-4.13.2-complete.jar -Dlanguage=Go -package parser -visitor -listener -Xexact-output-dir -lib parser -o parser grammar/SPLParser.g4
    gofmt -w parser/*.go

The lexer is generated first and the parser locates its vocabulary with `-lib parser`. Compare generated semantic changes, not generator source-path header differences. Regenerate a second time and expect no new diff. Do not run go mod tidy or upgrade the runtime to address generator-path errors.

    GOCACHE=/private/tmp/spl-toolkit-go-cache GOTOOLCHAIN=local go test -mod=readonly -race ./...
    GOCACHE=/private/tmp/spl-toolkit-go-cache GOTOOLCHAIN=local /private/tmp/spl-toolkit-floor-tools/gomodcache/golang.org/toolchain@v0.0.1-go1.22.12.darwin-arm64/bin/go test -mod=readonly ./pkg/analysis ./pkg/mapper ./pkg/bindings ./cmd ./pkg/api
    GOCACHE=/private/tmp/spl-toolkit-go-cache GOTOOLCHAIN=local make build-all
    /private/tmp/spl-toolkit-remaining-venv/bin/python -m pytest tools/tests -q
    /private/tmp/spl-toolkit-remaining-venv/bin/python tools/check_go.py
    /private/tmp/spl-toolkit-remaining-venv/bin/python tools/check_docs.py

For source-mode native testing on this macOS host:

    PYTHONPATH=python SPL_EXPECTED_VERSION=$(cat VERSION) SPL_NATIVE_LIBRARY=/Users/jacobdelgado/repos/spl-toolkit/.worktrees/remaining-milestones/build/libspl_toolkit.dylib /private/tmp/spl-toolkit-remaining-venv/bin/python -m pytest python/tests -q

For package building and outside-checkout acceptance, use the repository's existing controlled Make targets. PYTHON/PIP overrides prevent accidental use of a different environment. Builds may need sandbox permission for loopback or task-local caches; classify such failures as environment issues and rerun through the controller without changing assertions.

    GOCACHE=/private/tmp/spl-toolkit-go-cache GOTOOLCHAIN=local make python-build PYTHON=/private/tmp/spl-toolkit-remaining-venv/bin/python PIP=/private/tmp/spl-toolkit-remaining-venv/bin/pip
    GOCACHE=/private/tmp/spl-toolkit-go-cache GOTOOLCHAIN=local /private/tmp/spl-toolkit-remaining-venv/bin/python tools/check_package.py --sdist dist/spl_toolkit-$(cat VERSION).tar.gz --wheel-dir dist

The checker must exercise copied required native and cross-surface tests under an installed wheel outside this checkout. Record actual collection/pass/skip counts and artifact paths. The development suite's intentional installed-version skip is not acceptable inside required installed-wheel tests. These commands establish local platform acceptance only.

## Validation and Acceptance


After building, the following commands demonstrate the feature. `analyze --format json` emits one canonical report; inspect its status, coverage, stage IDs, and references.

    build/spl-toolkit analyze --query 'search src=1 | eval a=src, b=a+1 | table b' --format json
    build/spl-toolkit analyze --query 'search src=1 | fields - src | where src>1' --format json
    build/spl-toolkit analyze --query 'search src=1 | mystery x' --format json
    build/spl-toolkit capabilities --format json

Expected content statuses/exits are valid/0, invalid/1 with SPL_UNAVAILABLE_FIELD, and incomplete/3 with SPL_UNSUPPORTED_COMMAND. The first query has one source field src and derived a/b relationships; the last preserves src. Capability format is `1`, SPL/splunkd/current, with syntax and semantics separated. Do not require literal Go test counts before implementation; record actual counts in Progress and the milestone log after execution.

Success requires supported corpus references to have correct role, scope, stage, exact source span, and provenance; all tested transitions and scopes to match approved semantics; partial findings to survive unsupported/invalid neighbors; unsupported effects never to be complete; all four surfaces to equal the reviewed expected reports; old Go/Python/CLI/REST behavior to pass regression tests; and local installed native package acceptance with zero required skips. No remote release matrix is implicitly claimed.

## Idempotence and Recovery


All parsing is local and read-only. Repeating tests or generation is safe. Rebuilding outputs under build/dist uses existing repository conventions; never run blanket cleanup of other agents' files. Preserve the original grammar diff until legacy and new tests pass; revert only an owned change after controller agreement, never reset the worktree. For partial parser recovery retain the error and trustworthy earlier nodes; do not label discarded information as absent or complete. Package failures caused by missing new sources must be fixed in explicit manifests, not bypassed with checkout imports. If a final golden is wrong, fix its independently justified expectation and explain the reason in Decision Log.

## Artifacts and Notes


Baseline evidence came from the controller on 2026-09-07: Go race passed; build-all passed; 69 tool tests passed; Python source tests 27 passed with one intentional installed-version skip; Go 1.22.12 targeted baseline and check_docs passed. It does not establish milestone 2 acceptance. Keep final focused command outputs and reviewed corpus in their owned test/evidence locations, and summarize paths/counts here as execution progresses. Do not put generated caches or a tool JAR under version control.

Revision note (2026-09-07): Initial approved plan incorporates canonical implicit aggregate naming and installed-wheel analysis test/corpus closure, because source-only tests and raw function-call spelling would not meet the public contract.

Revision note (2026-09-07): Final controller approval fixes schema_version as integer 1 on both reports, requires tool-created SDD briefs/combined task reviews with serial implementers, and sets SPL_EXPECTED_VERSION for source-mode Make-built native testing.

Revision note (2026-09-07): Tasks 2 and 3 may narrowly extend grammar for their supported forms when necessary, because preserving typed structure is more reliable than recovering it from token text in semantic handlers.

Revision note (2026-09-07): Root approved an immutable per-input analysis CharStream marker for narrow ANTLR dot-disambiguation predicates. Default legacy tokenization is preserved, with adjacent/spaced concatenation, legacy mapping, both sequential orders and concurrent mixed-mode tests; no global mode or post-lex rewriting.

Revision note (2026-09-07): Task 1 completed after clean scoped re-review; Task 2 begins. Explicit empty CLI query remains a syntax-invalid content result, unlike omitted query usage errors. Reuse GOMODCACHE=/private/tmp/spl-toolkit-remaining-gomodcache with GOPROXY=off for required Go checks.

Revision note (2026-09-07): Root clarified rename overlaps: chains/swaps and collisions remain incomplete; before-state resolution applies only to ordinary distinct non-overlapping pairs. This supersedes an initially permissive controller interpretation and follows the approved honesty-over-coverage boundary.

Revision note (2026-09-07): Official SPL fields behavior retains internal fields. Root approved conservative open-input membership uncertainty with known internals/tombstones preserved and exact closed-input/table/stats behavior; M3 handoff must flag finite-catalog refinement. Unsupported rename overlaps forget affected binding claims while retaining sound reads.

Revision note (2026-09-07): Task 2 completed after three transfer fixes passed scoped re-review. Task 3 owns the carried typed quoted-asterisk resolution finding and must resolve it before milestone acceptance; table/stats restore local exact membership while global previous incomplete coverage remains.

Revision note (2026-09-07): Task 3 uses existing tested root-model/qualified-dataset identity conventions with explicit component/overlapping spans; typed macro-only stages use synthetic command macro, while statically named macro dependencies retain exact resolution despite unknown expansion semantics.

Revision note (2026-09-07): Root required narrow invalid-UTF8 document rejection for source preservation, including pre-decode raw REST/C JSON validation. Task3 owns core regressions; Tasks4/5 own transport regressions.

Revision note (2026-09-07): Task3 closed after independent scoped review confirmed original-bracket ownership prevents malformed child recovery leaks. The shared corpus contains 24 cases; quoted resolution follows semantic roles. Tasks4–6 remain serial and must deliver canonical parity and installed acceptance.

Revision note (2026-09-07): Root identified encoding/json replacement of unpaired UTF16 escapes. Task4 now supplies a narrow shared internal transport Unicode validator, consumed/packaged by Task5; test valid pairs and escaped-backslash controls, preserving standard decoder syntax handling.
