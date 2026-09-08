---
title: "Architecture"
layout: page
---

# Architecture

The Go core is the source of behavior. The generated ANTLR lexer and parser build the supported Milestone 1 parse tree. Listener code discovers seven flat categories, while a token rewriter replaces configured field tokens without reformatting the surrounding query.

```text
CLI ───────────────┐
REST handlers ─────┼──> Go mapper/parser
Python ctypes ─> C ABI ┘
```

The native C boundary keeps the existing exported operation signatures and result layouts. It owns mapper handles in a synchronized registry, and callers free returned results through the matching exported free functions. Python packages that boundary with the native library, checks native/package version agreement, and provides deterministic `close()` and context-manager lifecycle.

REST mapping configurations are request-scoped and may be cached by configuration content. Concurrent callers with different configurations remain isolated. A global fallback mapper remains for compatibility; its mutation endpoint is disabled by default.

Discovery is intentionally flat. It does not model stages, structured references, lineage, schemas, or read/write roles. Macro recovery is deliberately limited: when unsupported macro syntax causes parse errors, the result contains recovered macros and empty arrays for other categories.

Generated parser sources are committed build inputs. Runtime code and build configuration do not depend on temporary planning artifacts. Expanding the grammar and adding SPL2 modules remain later work.

## Canonical analysis and validation

`pkg/analysis` models Query Documents, stages, scopes, references, lineage, diagnostics, and explicit completeness. `Analyze` keeps its catalog-free contract. The additive `AnalyzeWithSourceFields` hook refines the same field-transfer kernel using a finite source universe and returns complete/incomplete selector expansions as a sidecar. Wildcard refinement changes downstream lineage inside the kernel; adapters never replay field flow.

`pkg/validation` strictly decodes and normalizes field catalogs, calls that canonical hook, and classifies source obligations and per-match evidence. Ordinary and optional declarations share the finite universe; optionality affects declaration equivalence only. Single/batch reports preserve embedded analysis and status/coverage separately. JSON Schema and OCSF field projection extend this layer; no event or expression-type validation is performed.

CLI and REST validate transport inputs and serialize reports. The C exports `spl_mapper_validate_fields` and `spl_mapper_validate_fields_batch` return owned `SPLResult` values, which callers release through `spl_result_free`. Python serializes requests and decodes results through the existing admission/close lifecycle. No adapter contains field-transfer, wildcard-matching, or catalog semantics.

Shared full reports in `testdata/validation/cases.json` are asserted directly in Go and in copied installed CLI/Python/HTTP acceptance tests. The package checker copies fixtures outside the checkout, supplies absolute environment paths, records hashes, and requires nonempty collection with zero failures or skips for both the built wheel and the wheel rebuilt from the source distribution.

## Local schema projection

`AnalyzeWithSourceUniverse` accepts copied candidate names, an exhaustive-universe promise, and a per-name admission resolver. An exact initial source read remains structurally available even when the external target prohibits it. Schema absence is classified by validation, preserving the original binding. The engine owns transfers, nested scopes, structural availability, partial selectors, and final reference IDs. Legacy finite field-list behavior and plain `Analyze` remain unchanged.

JSON Schema preparation builds local URI/anchor indexes. OCSF preparation reads normal compiler output and validates exact version, full extensions and explicit class/category/profile selection. Both use immutable prepared indexes and call-local recursion state. Preparation has no HTTP or file opener; CLI input file reads happen only at its explicit transport boundary. Neither runtime, package nor fixture loading depends on `_build_plan`.

The schema C exports return owned `SPLResult` values using the existing result/free ABI. Python single/batch calls use the existing admitted-operation and close lifecycle. The complete 28-case schema corpus and 22 malformed raw requests run through native libraries installed outside the checkout; CLI and real HTTP surface checks compare the frozen full Go reports. The package checker records copied fixture, wheel payload, loaded native library, and rebuilt-sdist source hashes. See [schema semantics and bounded projection](API.md#json-schema-and-ocsf-field-validation).
