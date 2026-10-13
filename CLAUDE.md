# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

SPL Toolkit is an offline Go library, CLI, REST server, and native-backed Python package for bounded operations on Splunk SPL and SPL2 queries: legacy field mapping and flat discovery, structured analysis and lineage, query requirements, knowledge-object closure, environment/schema validation, compatibility assessment, safe rewriting, resolution, detection workflows, and corpus/editor tooling (scan, graph, SARIF, impact, LSP).

The Go implementation is canonical. CLI (`cmd/`), REST (`pkg/api`), the C ABI (`pkg/bindings`), and Python (`python/spl_toolkit`, ctypes over the C ABI) serialize canonical Go reports. They must not reimplement or reclassify semantics.

## Commands

Toolchain: pinned Go 1.26.8 (CI sets `GOTOOLCHAIN=local`), with a declared source floor of Go 1.22.12 that CI also tests. Python 3.11+.

```bash
make deps                  # go mod download
make test                  # go test -mod=readonly -race ./...
make lint                  # python3 tools/check_go.py: gofmt check + go vet on handwritten packages + race tests
make build                 # build/spl-toolkit           (./cmd)
make build-server          # build/spl-toolkit-server    (./cmd/server)
make build-exporter        # build/spl-toolkit-export    (./cmd/splunk-export, CGO_ENABLED=0)
make build-shared          # build/libspl_toolkit.{so,dylib,dll}  (./pkg/bindings, c-shared)
make python-deps           # pip install -r python/requirements-dev.txt
make python-test           # build wheel+sdist, then tools/check_package.py runs native + acceptance suites outside the checkout
make generate-docs         # regenerate OpenAPI docs (docs/docs.go etc.) via swag, then tools/update_validation_openapi.py
make bench
```

Single Go test: `go test -mod=readonly -race -run 'TestName' ./pkg/analysis/`

Always pass `-mod=readonly`; the Makefile and CI never modify `go.mod`/`go.sum` implicitly (`make deps-update` is the deliberate tidy).

`go vet` must not run over generated code. `tools/check_go.py` excludes `parser/`, `gen/`, and `docs/docs.go` from gofmt/vet. Use it instead of a bare `go vet ./...`.

### Python and acceptance tests

`python/tests/` (native binding tests) and `tests/acceptance/` (cross-surface tests comparing Go, CLI, HTTP, and packaged Python reports) are designed to run against installed artifacts, not the source tree. Acceptance tests need absolute-path env vars such as `SPL_CLI` and `SPL_FIXTURES` and fail rather than skip when they are missing. `make python-test` / `tools/check_package.py` handles copying fixtures out of the checkout, setting env, and requiring zero skips. Prefer that over invoking pytest directly.

`tests/acceptance/cli_examples.json` is the executable manifest behind `docs/cli.md`. Documented CLI syntax and output are tested (`test_documented_cli.py`), so doc examples and this manifest must change together.

## Architecture

### Parsing

- `grammar/*.g4` holds the ANTLR grammars for legacy SPL and SPL2. `parser/` (SPL) and `parser/spl2/` (SPL2) are ANTLR 4.13.2 generated output, committed as build inputs. Never hand-edit them or edit them to silence warnings; regenerate only for an approved grammar change. `gen/spl_lexer.go` is an older generated lexer and nothing imports it.
- `pkg/analysis` admits input before parsing: a lexer work budget of 4,096 units. Exceeding it returns a canonical *incomplete* resource-limit result instead of parsing.

### Core engine: `pkg/analysis`

It owns Query Documents, stages, scopes, located references, lineage, diagnostics, explicit coverage, and **requirements**. `Analyze` runs once; `Result.Requirements` is projected from analysis-owned evidence, and `Requirements()` returns a detached copy rather than re-parsing. SPL2 is lowered (`spl2_lower.go`) into the same field-transfer kernel as SPL. The kernel never inspects grammar objects directly. Refinement hooks (`AnalyzeWithSourceFields`, `AnalyzeWithSourceUniverse`) let validators refine conclusions inside the kernel; adapters must not replay field-flow semantics.

It also embeds the **capability ledger** (`capabilitydata/`, `capability_*.go`). It covers five independent dimensions (syntax, semantics, requirements, linting, safe_rewriting) with states supported/partial/unsupported/not_applicable/unassessed, each backed by stable evidence IDs. The ledger publishes a semantic revision digest (quoted in README and `docs/architecture.md`). Changing claims or evidence changes that digest, and the docs quoting it must be updated.

### Layered packages (each consumes `pkg/analysis` evidence; none adds its own parser semantics)

- `pkg/mapper`: legacy token-rewrite mapping and flat 7-category discovery (data models, datasets, lookups, macros, sources, sourcetypes, input fields). Preserve exported signatures and all seven categories.
- `pkg/validation`: field-catalog, JSON Schema, and OCSF validation. Compiles to finite source universes and calls the analysis refinement hooks.
- `pkg/environment`: offline snapshot (v1/v2) and schema-bundle decoding, digests, coverage, prepared indexes.
- `pkg/closure`: expands caller-supplied knowledge-object definitions into a dependency graph and detection BOM.
- `pkg/compatibility`: compares requirements against prepared environment artifacts, reusing validation and closure.
- `pkg/rewrite`: explicit rule-based safe rewrites, built on analysis `PrepareRewrite`/`Verify`. Select on original facts, apply declared byte edits, reparse, and publish through one final gate. See `docs/rewrite.md` before touching source-binding or commit-gate behavior.
- `pkg/resolution`, `pkg/workflow`: named-input resolution and detection-workflow orchestration over the above.
- `pkg/corpus`, `pkg/corpusio`, `internal/corpusfs`, `pkg/graph`, `pkg/sarif`, `pkg/impact`, `pkg/document`, `internal/lsp`: corpus scanning, evidence projection, before/after impact, detached document snapshots, LSP. These are projection and orchestration layers only.
- `internal/splunkexport` (`cmd/splunk-export`): the only component that talks to live Splunk. It produces Snapshot v2. Everything else is offline.

### Surfaces and contracts

- `contracts/v1`, `contracts/v2`: versioned JSON Schemas for machine contracts (tested by `tests/acceptance/test_machine_contracts.py`). Report shapes are public contracts.
- C ABI (`pkg/bindings`): handle registry; every export returns an owned `SPLResult` freed via `spl_result_free`. Keep existing exported signatures and result layouts stable. CI runs an ASan/leak harness (`tests/native/memory.c`).
- Exit codes are a stable contract across CLI commands: 0 valid/satisfied, 1 invalid/unsatisfied, 3 incomplete/not assessed, 2 usage/request/I/O error.
- REST mapping config is request-scoped. The process-global `/api/v1/mappings` mutation endpoint is disabled by default.

### Source manifests

Packaging and release are hash-checked against explicit file lists. Adding a Go source file, contract, or fixture that the native library or release needs means updating `python/native-source-files.txt` and/or `tools/release-source-files.txt`, or the package and reproducibility checks will fail.

## Repository conventions

- Analysis outputs are honest about limits: incomplete, partial, and gap evidence stays visible rather than being guessed away. Preserve documented limitations and independent corpus expectations (for example `testdata/`, embedded corpora) when changing behavior.
- Grammar expansion, SPL2 module forms beyond same-document declarations, and new SDKs need an approved design (per `docs/contributing.md`), not opportunistic changes.
- Build and runtime paths must not depend on temporary planning artifacts (e.g. `_build_plan`).
- Version comes from `VERSION`, injected via ldflags into `internal/buildinfo`. Plain `go build` reports `dev`.
- `docs/evidence/` holds committed milestone verification records. Treat it as historical, not as code.
