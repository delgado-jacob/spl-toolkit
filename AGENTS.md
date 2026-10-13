# SPL Toolkit contributor guidance

## Scope and product boundary

- This file applies throughout the repository; more-specific instructions and direct user requests take precedence.
- Work as a contributor: keep changes focused and preserve unrelated files. Follow [docs/contributing.md](docs/contributing.md) for project policy.
- SPL Toolkit provides bounded, offline SPL/SPL2 analysis, validation, mapping, rewriting, compatibility assessment, and detection tooling. The Go implementation is canonical.
- CLI, HTTP, native C, and Python adapters serialize canonical Go reports; do not add independent parsers, field-flow logic, or status classification in adapters.
- Live Splunk acquisition belongs to the separate exporter. Offline operations must not fetch schemas, resolve external modules, inspect live inventory, or execute queries.
- Grammar expansion, additional SPL2 module forms, and new SDKs require an approved design. Do not broaden supported behavior opportunistically.

## Where to work

| Path | Responsibility |
| --- | --- |
| `pkg/analysis/` | Query documents, SPL/SPL2 lowering, field transfers, lineage, requirements, capability ledger, rewrite proof |
| `pkg/mapper/` | Legacy token mapping and seven-category flat discovery |
| `pkg/validation/` | Field catalogs, local JSON Schema and exact-version OCSF validation through analysis refinement |
| `pkg/environment/`, `pkg/closure/`, `pkg/compatibility/` | Captured artifacts, supplied definition closure, and offline assessment |
| `pkg/rewrite/`, `pkg/resolution/`, `pkg/workflow/` | Explicit safe edits, bounded named-input candidates, and detection orchestration |
| `pkg/corpus/`, `pkg/corpusio/`, `internal/corpusfs/` | Corpus evaluation and explicit local file acquisition |
| `pkg/graph/`, `pkg/sarif/`, `pkg/impact/`, `pkg/document/`, `internal/lsp/` | Evidence projections, comparisons, detached snapshots, editor protocol |
| `cmd/`, `cmd/server/`, `pkg/api/` | CLI and HTTP surfaces |
| `pkg/bindings/`, `python/spl_toolkit/` | Owned C ABI and ctypes Python bindings |
| `internal/splunkexport/`, `cmd/splunk-export/` | Bounded live acquisition producing Snapshot v2 |
| `grammar/`, `parser/`, `gen/` | Grammar sources and committed generated ANTLR build inputs |
| `contracts/v1/`, `contracts/v2/` | Versioned public machine schemas; see [contracts/README.md](contracts/README.md) |
| `testdata/`, `python/tests/`, `tests/acceptance/`, `tests/native/` | Behavioral fixtures, native Python, cross-surface and C ownership checks |
| `tools/`, `.github/workflows/` | Packaging, reproducibility, documentation, and exact-source CI checks |

Read the nearest implementation and tests before editing. Consult [docs/architecture.md](docs/architecture.md) for ownership seams and [docs/API.md](docs/API.md) for public report contracts.

## Setup and commands

- Run commands from the repository root. Prefer a local virtual environment for Python dependencies.
- Release/CI Go is pinned to **1.26.8**; the source floor **1.22.12** is separately tested on Linux. Python requires **3.11+**. Check CI files for current platform/compiler pins.
- Native builds and race tests require a C compiler and CGO. macOS native builds target macOS 15.0; the Makefile supplies deployment flags.
- Use `-mod=readonly` for Go builds/tests. `make deps` downloads modules; `make deps-update` deliberately changes module metadata and is not routine setup.

| Command | Use |
| --- | --- |
| `make deps` | Download Go dependencies |
| `make python-deps` | Install pinned Python development requirements |
| `go test -mod=readonly -race -run 'TestName' ./pkg/analysis/` | Focused Go test; substitute the affected package and test |
| `make test` | Full Go race suite |
| `make lint` | Handwritten Go formatting check, filtered vet, and full race suite |
| `make build` | Build versioned CLI at `build/spl-toolkit` |
| `make build-server` | Build `build/spl-toolkit-server` |
| `make build-exporter` | Build `build/spl-toolkit-export` without CGO |
| `make build-shared` | Build platform native library for Python/C |
| `make python-test` | Build wheel/sdist and test installed artifacts outside the checkout |
| `make generate-docs` | Regenerate OpenAPI via pinned swag and the validation postprocessor |
| `python3 tools/check_docs.py` | Validate documentation source/front matter; needs PyYAML |
| `make bench` | Go benchmark suite, when performance is relevant |
| `python3 tools/check_clean_build.py --ref HEAD` | Verify a committed ref builds without untracked inputs |
| `python3 tools/check_reproducible.py --ref HEAD --output build/reproducibility` | Controlled committed-source release acceptance; see CI for prerequisites |

- `make lint` already runs race tests; avoid repeating the full suite without a reason.
- Use `tools/check_go.py` rather than bare `go vet ./...`: generated ANTLR/OpenAPI packages are excluded from handwritten vet/format gates. Format only changed handwritten Go files with `gofmt`.
- `--ref HEAD` checks committed source, not uncommitted work. Do not describe it as verification of pending edits.
- `VERSION` owns release versions, injected through `internal/buildinfo`; plain Go builds may report `dev`.

## Behavioral contracts

- Preserve incomplete, partial, unsupported, and unassessed evidence. A parser success, supplied capture promise, or local corpus pass does not establish live Splunk compatibility or authorization.
- `pkg/analysis` owns direct requirements. Validators may refine public analysis through its hooks, but requirements retain query-only semantics. Do not replay field transfers downstream.
- Preserve deterministic identities/order, source locations, and detached report ownership. Query offsets use UTF-8 bytes; LSP positions require the existing UTF-16 conversion.
- Keep SPL and explicitly selected SPL2 contracts distinct. Atomic dotted names and structural field paths can share display text without sharing identity.
- Preserve exported mapper signatures, all seven discovery categories, C export signatures/result layouts, and versioned JSON shapes unless the requested change explicitly authorizes a contract change.
- CLI status codes are contractual: `0` valid/satisfied, `1` invalid/unsatisfied, `3` incomplete/not assessed, `2` request/usage/I/O errors. Check the command's documented policy before changing precedence.
- C results must be released using their matching exported free function. Python must retain deterministic `close()`/context-manager behavior and native/package version agreement.
- REST mapping configuration stays request-scoped; process-global mapping mutation is disabled by default.
- Read [docs/rewrite.md](docs/rewrite.md) before changing rewrite bindings or publication gates. Select against original evidence, apply declared edits, reparse, and preserve proof/audit outcomes.
- Resolution and workflow coordinate existing owners. Preserve unsuccessful siblings and entry-local failures; saved comparisons consume captured evidence without rerunning analysis or substituting current capability evidence. See [docs/resolution.md](docs/resolution.md) and [docs/workflow.md](docs/workflow.md).

## Generated files and packaging

- Do not hand-edit `parser/`, `gen/`, or generated OpenAPI artifacts such as `docs/docs.go`. ANTLR parser generation uses 4.13.2; the Go runtime module version is separate. Regenerate only for an authorized grammar change using the documented pinned procedure.
- For HTTP documentation changes, update source annotations/postprocessor as appropriate and run `make generate-docs`; review the resulting diff for unrelated churn.
- Packaging uses explicit source/content lists. When adding a needed source, schema, example, or fixture, check `python/native-source-files.txt`, `tools/release-source-files.txt`, and `tools/release-content-files.txt`; update only the applicable manifests.
- Capability claims/evidence change semantic revision digests. Run ledger tests and update documentation that quotes affected revisions; do not hard-code the current digest here.
- Keep independent corpus expectations meaningful. Do not regenerate expected reports solely to make changed behavior pass.
- Build/runtime paths must not depend on temporary planning artifacts. `docs/evidence/` contains historical receipts; do not rewrite them to imply current verification.

## Validation and delivery

- Start with focused Go behavioral tests, then run the relevant broader gate. Add regression coverage at the public boundary when practical; do not add tests that merely repeat documentation or static values.
- For shared report, native, packaging, or adapter changes, use `make python-test`. The package checker copies fixtures outside the checkout and runs native/cross-surface suites against installed artifacts, requiring collection and zero required skips.
- Direct acceptance pytest runs need explicit artifact/fixture paths (including `SPL_CLI` and `SPL_FIXTURES`); source-tree Python tests alone do not prove wheel/sdist correctness.
- Keep [docs/cli.md](docs/cli.md) examples synchronized with `tests/acceptance/cli_examples.json`, which is exercised by documented CLI acceptance tests.
- Changes to machine reports need affected schema and cross-surface checks. Native ownership changes also need the `tests/native/memory.c` sanitizer gate from CI.
- CI covers Linux amd64, macOS arm64, Windows amd64, Python installations, the Go floor, native memory, clean/Docker builds, and exact-source acceptance. Local success proves only the checks actually run.
- Distinguish product failures from unavailable compiler, cache, network, loopback, or credential prerequisites. Report blocked/skipped checks without weakening assertions.
- Clean up resources created by verification, including task-owned containers, networks, volumes, and images. Avoid broad prune commands.
- Review the final diff for unrelated changes. Report changed behavior and actual validation; keep local results, hosted CI, publication, and live runtime evidence distinct.
- Do not use release/tag/install targets as routine validation: `make tag` commits and pushes, and install targets write system paths. Publishing, merging, tagging, and live acquisition require authorization within the task.
