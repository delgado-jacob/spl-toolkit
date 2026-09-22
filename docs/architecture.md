---
title: "Architecture"
layout: page
---

# Architecture

## Corpus and editor orchestration

`pkg/corpus` prepares canonical validation and aggregates immutable snapshots;
`pkg/corpusio` and `internal/corpusfs` acquire explicitly selected local files
with anchored no-follow handles. `pkg/graph` and `pkg/sarif` project existing
evidence. `pkg/impact` compares canonical before/after validation/rewrite results.
`pkg/document` exposes detached snapshots and checked evidence lookups.
`internal/lsp` owns framed protocol/document lifecycles and UTF-16 conversion,
delegating all analysis and binding interpretation to canonical packages.
CLI, REST and owned native/Python adapters compose these operations without
adding semantic engines. See [tooling](tooling.md) and [contracts](contracts.md).

## Canonical requirements and parser admission

`pkg/analysis` owns both structured analysis and direct query requirements. `Analyze` normalizes the query document, preflights bounded lexer work, runs the selected SPL or SPL2 parser only after admission, and finalizes canonical references and diagnostics. A deterministic projector then creates `Result.Requirements` from analysis-owned evidence. `Requirements` performs one `Analyze` call and returns a deeply detached copy of that embedded set. It does not parse again, replay field transfers, or classify text spans.

Refinement-aware analysis keeps a private query-only requirement trace beside its public evidence. Field-list, JSON Schema, OCSF, and rewrite validation may refine public binding, wildcard, diagnostic, or coverage conclusions against a supplied target. The projector consumes the query-only trace, so the embedded set remains equal to plain analysis of the same normalized document. Final reference IDs are shared, and indexed pending-reference and incomplete-stage records preserve deterministic ordering without repeated linear searches.

Requirement items represent direct external obligations. Source-bound consuming fields and direct knowledge-object references can become items. Derived fields and non-consuming definitions remain analysis evidence but are not requirements. Adapters do not add environment state, expand knowledge objects, resolve placeholders, assess compatibility, generate variants, or execute queries. Query and capability digests identify supplied values and are not security credentials, signatures, authorization decisions, or execution permission.

`pkg/analysis` also owns the embedded capability ledger and corpus. It validates stable record and evidence IDs, selector agreement, typed observations, provenance, canonical ordering, and state-specific evidence rules before constructing a manifest. The five independent dimensions are syntax, semantics, requirements, linting, and safe rewriting. Supported, partial, unsupported, not-applicable, and unassessed claims stay distinct. Summary construction uses exact identities: applicable equals supported plus partial plus unsupported plus unassessed; record count equals applicable plus not applicable; covered equals supported. It computes no composite score.

Grammar registration is stored separately from syntax claims. The legacy command/function projection may expose registration as `syntax_supported`, but the ledger grants coverage only from supported evidence. Linting also has its own evidence boundary; parser or semantic diagnostics do not establish lint coverage. Evidence IDs bind records to the embedded static cases. A broader form gets a new ID unless a reviewed scope correction changes the original boundary. The selected semantic revision covers selectors, snapshot, legacy projections, rewrite forms, records, summary, and evidence, excluding only the display-oriented toolkit version.

Source Go builds may expose toolkit version `dev`; tagged CLI, server, native, and packaged builds use the exact `VERSION` while retaining the same semantic revision. The current SPL revision is `sha256:08901c84ac8c420f59ddb86168c534f0e83484c05d3c8a81078a716c878a1733`; the SPL2 revision remains `sha256:f1391296cfbc616e9bb1b1828e2471e37b60a35c0555654c0734640e072a0437`. The embedded corpus proves deterministic local toolkit behavior. The toolkit does not execute SPL, evaluate regular expressions, compare result rows, model acceleration, load macro definitions, or certify runtime compatibility. It also does not prove authorization or upstream support.

Lexer admission allows 4,096 work units. Real errors from one lexer call are counted in listener order before the returned non-EOF token; EOF does not count. SPL2 closure inspection runs after admitted EOF and accounts for its synthetic unterminated-literal error before parser construction. The first event that would consume unit 4,097 records the omitted source range and returns the canonical incomplete resource-limit result before parser prediction. Long sparse documents remain admitted when they stay within the work budget.

Every newly produced analysis result carries its requirements. Existing validation, rewrite, corpus, and impact reports inherit that member where they already serialize an analysis result. `pkg/document` deep-copies it into detached snapshots. This addition creates no requirement-specific graph or SARIF projection, corpus aggregate, impact comparison logic, or LSP behavior.

## Canonical safe rewriting

`pkg/rewrite` strictly prepares explicit rules and optional validation targets, then consumes the analysis-owned `PrepareRewrite` evidence, typed rendering and whole-candidate `Verify` facade. It selects against original facts, resolves simultaneous linked groups and collisions, reconstructs only declared byte edits, reparses the candidate, and applies one final publication gate. Adapters never infer aliases, implicit labels, SQL phases or source bindings themselves.

The report preserves both analyses and exact original/candidate audit locations even when apply returns the original string. CLI, real HTTP and owned C results/native Python serialize that report without reclassification. Legacy mapper/configuration precedence remains separate. The [rewrite guide](rewrite.md) describes the supported contract.

Durable rewrite assertions own candidate/returned text, status, commit, audit, source bindings and lineage independently of Go-produced transport reports. Installed acceptance copies every used fixture and maintained example outside the checkout and compares full reports across Go, CLI, HTTP and packaged native Python. Source/fixture hashes bind transport to the producing inputs; transport snapshots grant no semantic credit. Required native and surface suites cannot disappear or skip tests.

The Go core is the source of behavior. The generated ANTLR lexer and parser build the supported Milestone 1 parse tree. Listener code discovers seven flat categories, while a token rewriter replaces configured field tokens without reformatting the surrounding query.

```text
CLI ───────────────┐
REST handlers ─────┼──> Go mapper/parser
Python ctypes ─> C ABI ┘
```

The native C boundary keeps the existing exported operation signatures and result layouts. It owns mapper handles in a synchronized registry, and callers free returned results through the matching exported free functions. `spl_mapper_requirements_query` returns an owned `SPLResult`, including for request or handle errors; callers release it with `spl_result_free`. Python packages that boundary with the native library, checks native/package version agreement, and provides deterministic `close()` and context-manager lifecycle.

REST mapping configurations are request-scoped and may be cached by configuration content. Concurrent callers with different configurations remain isolated. A global fallback mapper remains for compatibility; its mutation endpoint is disabled by default.

Discovery is intentionally flat. It does not model stages, structured references, lineage, schemas, or read/write roles. Macro recovery is deliberately limited: when unsupported macro syntax causes parse errors, the result contains recovered macros and empty arrays for other categories.

Generated parser sources are committed build inputs. Runtime code and build configuration do not depend on temporary planning artifacts. Standalone SPL2 has its own generated ANTLR lexer/parser and typed syntax ownership. SPL2 modules remain excluded.

## Canonical analysis and validation

`pkg/analysis` models Query Documents, stages, scopes, references, lineage, diagnostics, and explicit completeness. `Analyze` keeps its catalog-free contract. The additive `AnalyzeWithSourceFields` hook refines the same field-transfer kernel using a finite source universe and returns complete/incomplete selector expansions as a sidecar. Wildcard refinement changes downstream lineage inside the kernel; adapters never replay field-flow semantics.

Typed SPL command handlers own the bounded Milestone 10 forms. `tstats` resets its generating source before collecting exact model, dataset, predicate, aggregate, and group evidence. The selected field commands and functions reuse the existing field environment and requirement trace. Exact macro invocations retain one direct requirement and a source-located unresolved-expansion gap without loading a definition. Join and append children start independently, appendpipe receives a copy of the parent environment, and every branch preserves child evidence without installing child fields into the parent. The parent keeps known fields in an uncertain environment because merge effects are unmodeled.

The detection-impact fixture is test-only evidence. It holds query bytes and a pre-change baseline of 12 syntax-complete cases, 1 semantic-complete target stage, and 35 non-macro gaps. Current analysis yields 14, 16, and 13. Measuring selected target stages allows exact macro or branch boundaries to remain visible without erasing a proved command-level improvement.

`pkg/validation` strictly decodes and normalizes field catalogs, calls that canonical hook, and classifies source obligations and per-match evidence. Ordinary and optional declarations share the finite universe; optionality affects declaration equivalence only. Single/batch reports preserve embedded analysis and status/coverage separately. JSON Schema and OCSF field projection extend this layer; no event or expression-type validation is performed.

CLI and REST validate transport inputs and serialize reports. The C exports `spl_mapper_validate_fields` and `spl_mapper_validate_fields_batch` return owned `SPLResult` values, which callers release through `spl_result_free`. Python serializes requests and decodes results through the existing admission/close lifecycle. No adapter contains field-transfer, wildcard-matching, or catalog semantics.

Shared full reports in `testdata/validation/cases.json` are asserted directly in Go and in copied installed CLI/Python/HTTP acceptance tests. The package checker copies fixtures outside the checkout, supplies absolute environment paths, records hashes, and requires nonempty collection with zero failures or skips for both the built wheel and the wheel rebuilt from the source distribution.

## Local schema projection

SPL2 lowering consumes its typed syntax and original token locations, then shares
the field-transfer kernel and canonical validators with SPL. The shared kernel
receives located operands and prepared projections; it does not inspect SPL2
grammar objects. SQL scheduling separates source/filter/group/evaluation from
final projection, preserving lexical references while recording actual lineage
phases. Independent children start with source environments; inherited subpipes
receive copies. Unproved parent merges never install guessed child outputs.

The corpus retains separate raw syntax and canonical-result expectations. Its
auditor checks immutable provenance, assembly rules, original obligations, held
boundaries and deduplicated floors without parsing queries. CLI, HTTP and Python
serialize the canonical results. Installed acceptance copies the SPL2 corpus and
maintained documentation outside the checkout and requires every registered
native/surface suite to collect tests with no required skips. The
[SPL2 guide](spl2.md) describes the public boundaries.

`AnalyzeWithSourceUniverse` accepts copied candidate names, an exhaustive-universe promise, and a per-name admission resolver. An exact initial source read remains structurally available even when the external target prohibits it. Schema absence is classified by validation, preserving the original binding. The engine owns transfers, nested scopes, structural availability, partial selectors, and final reference IDs. Legacy finite field-list behavior and plain `Analyze` remain unchanged.

JSON Schema preparation builds local URI/anchor indexes. OCSF preparation reads normal compiler output and validates exact version, full extensions and explicit class/category/profile selection. Both use immutable prepared indexes and call-local recursion state. Preparation has no HTTP or file opener; CLI input file reads happen only at its explicit transport boundary. Neither runtime, package nor fixture loading depends on `_build_plan`.

The schema C exports return owned `SPLResult` values using the existing result/free ABI. Python single/batch calls use the existing admitted-operation and close lifecycle. The complete 28-case schema corpus and 22 malformed raw requests run through native libraries installed outside the checkout; CLI and real HTTP surface checks compare the frozen full Go reports. The package checker records copied fixture, wheel payload, loaded native library, and rebuilt-sdist source hashes. See [schema semantics and bounded projection](API.md#json-schema-and-ocsf-field-validation).
