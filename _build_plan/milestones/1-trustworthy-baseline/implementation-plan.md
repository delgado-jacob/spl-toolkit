# Trustworthy Baseline Implementation Plan


> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking in Progress; the matching numbered instructions appear under Concrete Steps.

**Goal:** Make existing SPL discovery and mapping installable, concurrent, accurately documented, and verifiable through Go, CLI, REST, and native Python on four required platform targets.

**Architecture:** Retain Go, ANTLR, the flat discovery model, token rewriting, and existing public signatures. Repair configuration evaluation and ownership in the core, allocation and lifecycle at the C/Python boundary, and the existing adapters. Use one release version source and install real artifacts in isolated acceptance environments.

**Tech Stack:** Go module language floor 1.22, release compiler 1.26.8, ANTLR Go runtime 4.13.1, C-compatible shared libraries, Python 3.11+, ctypes, setuptools, pytest, GitHub Actions, and Make.

**Spec:** `docs/superpowers/specs/2026-09-06-trustworthy-baseline-design.md`, approved in conversation before this plan. Read it with this document; the executable requirements are repeated here so the plan is self-contained.

## Global Constraints


"This design implements only milestone 1 of the approved PRD."

"The Go implementation remains canonical."

"Standardize package metadata and documentation on Python 3.11+."

"Required native execution targets are Linux x86-64, macOS x86-64, macOS arm64, and Windows x86-64."

"Keep existing C result structure layouts and exported operation signatures."

"Base mappings apply first. Evaluate enabled conditional rules in ascending numeric priority. Equal priorities preserve configuration order. Only the first matching rule contributes mappings, and its mappings override base mappings for the same source field."

"Exit codes are 0 for success, 1 for rejected query or configuration content, and 2 for usage or file-access errors."

"Reproducibility means rebuilding the same source revision with the same pinned toolchain and environment produces matching final artifact checksums."

"Runtime code, build configuration, tests, and release automation must not depend on that directory." Here, "that directory" means `_build_plan/`.

No structured analysis, lineage, schemas, SPL2, batch scanning, grammar expansion, new SDKs, or mapping-2.0 architecture. No publishing, remote push, tag creation, or merge is part of plan execution without the user's corresponding authorization. Runtime processing is offline; acquiring pinned build dependencies is setup work.

---

## Purpose / Big Picture


A user will be able to clone the project, build its CLI/server/native package without modifying tracked source or dependencies, load an actual mapping configuration, and obtain equivalent existing SPL results through all four surfaces. Python users will receive complete native discovery arrays and deterministic cleanup rather than relying on mocks or garbage-collection timing. Release evidence will distinguish a successful local test from successful execution on all required platforms.

The principal demonstration is `search src_ip=1` mapped by `{"version":"1.0","mappings":[{"source":"src_ip","target":"source_ip"}]}` to exactly `search source_ip=1`. Equivalent discovery fixtures include two macro names and three distinct data-model references. Successful syntax validation means accepted by the current grammar only.

This is a living ExecPlan maintained according to `/Users/jacobdelgado/.codex/PLANS.md`. Preserve Progress, Surprises & Discoveries, Decision Log, and Outcomes & Retrospective through execution. To reconcile the writing-plans step checklist with ExecPlan prose requirements, maintain the per-step checkboxes only in Progress and use numbered prose plus indented executable examples below. Execution status is recorded below.

## Progress


- [x] (2026-09-06) Approved design read; source contracts and public build metadata inspected.
- [x] (2026-09-06) Plan authored and checked against the approved spec.
- [x] (2026-09-06) User approved continuous subagent implementation and review; no intermediate user-review stops.
- [x] Task 1.1: Establish isolated baseline and dependency failure evidence.
- [x] Task 1.2: Add clean-build acceptance check.
- [x] Task 1.3: Repair checksum state and non-mutating Make workflows.
- [x] Task 1.4: Verify clean builds and commit only Task 1 files.
- [x] Task 2.1: Add rule-priority, regex, and validation regression cases.
- [x] Task 2.2: Observe failing rule tests.
- [x] Task 2.3: Implement ordered first-match evaluation and strict supported settings.
- [x] Task 2.4: Verify rule tests and commit Task 2.
- [x] Task 3.1: Add configuration-ownership, atomic-update, and parser-race tests.
- [x] Task 3.2: Observe baseline failures/races.
- [x] Task 3.3: Implement owned configuration and operation-local state.
- [x] Task 3.4: Run race tests and commit Task 3.
- [x] Task 4.1: Add native handle and array regression tests.
- [x] Task 4.2: Observe native failures in child processes.
- [x] Task 4.3: Repair registry, allocations, and string cleanup.
- [x] Task 4.4: Verify actual native behavior and commit Task 4.
- [x] Task 5.1: Add Python lifecycle, ownership, and concurrency tests.
- [x] Task 5.2: Observe failing lifecycle tests.
- [x] Task 5.3: Implement close/context manager, locking, and pointer release.
- [x] Task 5.4: Run wrapper/native tests and commit Task 5.
- [x] Task 6.1: Add CLI invocation and output regression tests.
- [x] Task 6.2: Observe baseline CLI failures.
- [x] Task 6.3: Implement argument parsing, config loading, and output contracts.
- [x] Task 6.4: Verify CLI tests and commit Task 6.
- [x] Task 7.1: Add version and source-distribution/wheel installation tests.
- [x] Task 7.2: Observe packaging/version failures.
- [x] Task 7.3: Implement single-source Go/native version reporting.
- [x] Task 7.4: Implement Python metadata and fail-closed wheel builds.
- [x] Task 7.5: Stage self-contained source distributions and package Make targets.
- [x] Task 7.6: Verify installed wheel and extracted sdist and commit Task 7.
- [x] Task 8.1: Add common cross-surface fixtures and HTTP parity tests.
- [x] Task 8.2: Add Go/HTTP fixture tests and repair demonstrated REST inconsistencies.
- [x] Task 8.3: Implement installed-package cross-surface acceptance.
- [x] Task 8.4: Turn current CLI examples into executable acceptance cases.
- [x] Task 8.5: Align current documentation and examples with implemented behavior.
- [x] Task 8.6: Record fixed-fixture performance measurements.
- [x] Task 8.7: Verify parity, documentation, and benchmark report; commit Task 8.
- [x] Task 9.1: Add artifact normalization and mismatch regression tests.
- [x] Task 9.2: Implement deterministic release artifact creation.
- [x] Task 9.3: Implement two-clean-build equality and installation verification.
- [x] Task 9.4: Record and enforce release toolchain/target inputs.
- [x] Task 9.5: Connect non-publishing Make release targets.
- [x] Task 9.6: Repair and verify existing container workflows.
- [x] Task 9.7: Verify artifact equality and commit Task 9.
- [x] Task 10.1: Wire four native runner jobs and interpreter matrix.
- [x] Task 10.2: Exercise installed wheels across the interpreter matrix.
- [x] Task 10.3: Wire minimum-Go, native-memory, Make, and container gates.
- [x] Task 10.4: Implement required-evidence aggregation and its failure tests.
- [x] Task 10.5: Commit/review the exact candidate and obtain current CI results.
- [x] Task 10.6: Write the milestone log with proven results and open gates.
- [x] Task 10.7: Commit final documentation and verify evidence ancestry.
- [x] Final review correction: preserve boundary whitespace for mapped and no-op queries with shared exact-output fixtures.
- [x] Final review acceptance: rerun all hosted gates and the measured baseline on corrected source.

## Surprises & Discoveries


At plan entry, HEAD is `b7b39157bb4822f77390faf7c41921fe64645daa`, the design-only commit above implementation baseline `a276154689d946c9fd667765a7ef48759edfa021`. `go.sum` exists locally but is untracked. A successful test in this checkout cannot establish that a clone has complete checksums.

`Parser.Parse` clears and appends to a shared error listener. Mapping locks alone will not make public parser reuse safe. The C data-model array loop advances the base pointer cumulatively; a three-element fixture is necessary to expose the out-of-bounds path. Native discovery omits macros altogether. Python's `c_char_p` conversion for a standalone error return discards the address needed to free that allocation.

`NewWithConfig(*MappingConfig) *Mapper` has no error result. Preserve this signature: store constructor validation errors internally and return them from mapping operations; JSON loaders still reject invalid configurations immediately. Discovery and syntax-only validation need not fail because unused mapping settings are invalid.

The current Python tests prepend the source directory to `sys.path`. Those tests cannot prove that an installed wheel contains the library. New installed-package tests must run under `python -I`, outside the checkout, with copies of only test files and fixtures.

Read-only fixture probes confirmed three data models and two macros through the current Go CLI. These probes verify fixture suitability, not milestone correctness. No required platform acceptance suite or release reproducibility test ran during planning.

The local tooling observed during planning is Go 1.25.5 and Python 3.12.6 on macOS arm64. Those are not the pinned release tools below; do not label their results as pinned release evidence.

Execution evidence, 2026-09-06: the clean-export baseline failed on missing checksums, the native three-model regression crashed its child process with exit -5, and shared parser reuse produced race reports and cross-call errors. Tasks 1–4 now pass independent task reviews. Task 1 review also caught an unguarded vet command; explicit -mod=readonly and a permissive-GOFLAGS regression closed it. Local Go 1.26.8, Go 1.22.12, and all four pinned Python interpreters have been acquired in task-owned locations. Native platform and release gates are still open.

Task 7 review found that apparent isolated-package success still relied on a .pth exposing outer test dependencies, and recursive source-directory staging could include unrelated user files. The fix rounds closed both defects and explicit macOS tag/minimum coupling, then repeated isolated wheel and extracted-source installation checks successfully.

Execution platform diagnostic: Go1.22.12 with cgo/race on macOS26.2 can emit a binary without LC_UUID that the dynamic loader rejects before tests. A minimal net-import probe confirmed old default-link failure and current Go1.26.8 success; the required Go floor gate remains Linux. Separately, a minimal dylib built with the initially proposed -no_uuid failed to load; -reproducible retained UUIDs, loaded, and produced identical bytes across two directories. The actual Go native payload must still pass Task9 equality/load checks. Evidence is in task scratch linker diagnostics and the Go1.24 linker release note (https://go.dev/doc/go1.24#linker).

Task 9 independent review closed four release-gate defects: validation now executes from the exported revision, environment identity is observed and checked, Docker inputs use an exact allowlist, and Windows drive/UNC archive paths are rejected. Diagnostic builds carry a separate status and cannot stand in for pinned acceptance. Matching production/diagnostic-base Git tree IDs and the sole environment substitution were verified directly. The original source floor passed on Linux Go 1.22.12; an early Linux Go 1.26.8/GCC 12 ASan diagnostic also passed, with final GCC 13 acceptance still required.

Task 10 acceptance, 2026-09-06: independent review closed incomplete nested environment validation and evidence overwrite behavior. The first full hosted run exposed a missing pinned dependency bootstrap in 15 installed-wheel jobs; its other gates passed and aggregation correctly rejected the incomplete matrix. The reviewed bootstrap correction at `5191206124590f69a18c81fc51e756b6e75c90e8` passed all 24 jobs in run `34075500813`, including four reproducible releases, all 16 installed-wheel combinations, Go 1.22.12, GCC 13 ASan, Linux Docker/Make/clean-source workflows, and final strict aggregation. Controller validation of the 28 downloaded records passed and all 32 release payload checksums matched. Each installed combination passed 11 native and five surface tests with zero skips. Complete logs and payloads are retained under `build/evidence/ci-34075500813/`; the later documentation commits preserve this tested source identity and do not change runtime, test, or build inputs. Final whole-branch review is recorded separately.

Final whole-branch review found one contract gap: the token rewriter removed caller-owned boundary whitespace after mapping. Commit `6b55f8902ff1a990aea8951cd78934b32c90660a` removes only the blanket trim, retains exact EOF-marker cleanup, and adds mapped/no-op spaces, tabs, and newlines to the shared fixtures. The CLI assertion now compares exact bytes plus its one output newline. Local Go and installed Python/CLI/HTTP tests passed; [run 34077738907](https://github.com/delgado-jacob/spl-toolkit/actions/runs/34077738907) passed all 24 jobs on the corrected source. The controller revalidated all 28 downloaded records and 32 payload hashes. A fresh five-repetition benchmark at the same source is retained in `build/evidence/final-benchmark.txt`; the permanent performance report and acceptance records use this corrected source. The scoped fix-review verdict is recorded separately.

## Decision Log


Decision: Keep one milestone plan with ten separately reviewable tasks, rather than unrelated subsystem plans. Rationale: this is one approved baseline whose final acceptance depends on cross-surface equivalence and packaging; each task below still produces an independently testable change. Date/Author: 2026-09-06, Codex.

Decision: Use `_build_plan/milestones/1-trustworthy-baseline/implementation-plan.md` for this temporary ExecPlan, matching the approved design's handoff location. Rationale: avoid making permanent runtime guidance depend on temporary milestone artifacts. Date/Author: 2026-09-06, Codex.

Decision: Release builds use Go 1.26.8; retain `go 1.22` in go.mod and test Go 1.22.12 separately. Rationale: pin an available compiler without silently increasing the public language floor. Do not use language/library features newer than 1.22 in project source. Date/Author: 2026-09-06, Codex.

Decision: Test CPython 3.11.9, 3.12.10, 3.13.7, and 3.14.0 on all four platform jobs. Rationale: these are explicit reproducible compatibility test versions; they are not claims about the newest patch release. The ordinary GIL-enabled interpreter is the target; free-threaded builds are not claimed. Date/Author: 2026-09-06, Codex.

Decision: Initially distribute native platform wheels as downloadable release artifacts: `py3-none-linux_x86_64`, `py3-none-macosx_15_0_x86_64`, `py3-none-macosx_15_0_arm64`, and `py3-none-win_amd64`. Rationale: ctypes does not use the CPython extension ABI; native execution on the four approved targets is required, while broad manylinux compatibility and PyPI publication are not claimed by this milestone. Linux support is explicitly Ubuntu 24.04/glibc 2.39 for these wheel artifacts; macOS deployment target is 15.0; Windows native tests run on Server 2022 x64. Do not relabel Linux wheels as manylinux without the corresponding build/repair/conformance work. Date/Author: 2026-09-06, Codex.

Decision: Build twice in separate source/output/cache directories on the same CI job and record the resolved OS image/compiler. Rationale: a runner label is not an immutable machine image; the approved reproducibility claim is equality within an identical recorded environment. Compare final payload bytes and archives; never hide native mismatches by comparing only unpacked Python sources. Date/Author: 2026-09-06, Codex.

Decision: Keep recursive configuration ownership in focused pkg/mapper/config_clone.go, rejecting unsupported or cyclic input and preserving the existing scalar domain. Rationale: the ownership logic has one responsibility and should not obscure mapping synchronization; independent review accepted the separation. Date/Author: 2026-09-06, Codex.

Decision: Replace the planned macOS native -no_uuid flag with -reproducible while retaining content-derived UUIDs. Rationale: local minimal tests prove macOS26 rejects UUID-less dylibs and prove equal/loadable outputs with reproducible linking; Task9 still verifies actual native artifacts. Date/Author: 2026-09-06, Codex.

## Outcomes & Retrospective


Tasks 1–10 have passed their independent code reviews, including focused fix reviews. Runtime and release acceptance is established at implementation SHA `6b55f8902ff1a990aea8951cd78934b32c90660a` by all 24 successful jobs in [run 34077738907](https://github.com/delgado-jacob/spl-toolkit/actions/runs/34077738907). The controller independently accepted all 28 downloaded records and verified all 32 release payload hashes. This covers actual execution on all four targets and all 16 pinned interpreter combinations, alongside the language floor, real native sanitizer, source preservation, documented workflows, and reproducibility gates. The milestone log and permanent compatibility/evidence records were committed separately after those results, with tested-source ancestry and unchanged runtime/configuration/test inputs verified. The final whole-branch review identified and prompted the whitespace correction described above; its scoped fix-review verdict is recorded separately; no merge, tag, or release publication has occurred.

## Context and Orientation


Run commands from the isolated implementation checkout, called `WORKTREE` in the prose. On macOS/Linux examples, activate a task-local `.venv` so `python` means its interpreter. On Windows, use the equivalent `.venv/Scripts/python.exe` and Git Bash for shell examples; CI can invoke Python tools directly in PowerShell. Use absolute paths when a subprocess changes working directory.

Go entry points live in `cmd/main.go` and `cmd/server/main.go`. `pkg/mapper` owns behavior, `pkg/api` wraps it in HTTP, and `pkg/bindings` is a `package main` built with `-buildmode=c-shared`, which produces a shared library and generated C header. ctypes calls that library through numeric handles. A wheel is the installable Python archive; an sdist is a source archive from which a wheel must be buildable without the checkout.

Do not copy existing untracked build products into the implementation tree. Preserve `.omo/`, `CLAUDE.md`, `_build_plan/`, `bin/`, `bindings`, `go.sum`, and `spl-field-mapper` in the user's original checkout. Only the approved spec and this plan are relevant planning inputs. If using subagents, use the available GPT-5.6 family model at medium effort (`gpt-5.6-sol`, `medium`), assign exact file ownership, and tell each worker that others may be editing the same repository.

## File Structure and Responsibilities


Task 1 modifies `Makefile` and adds tracked `go.sum`, `tools/check_clean_build.py`, and `tools/check_go.py`. The first tool checks a build from an exported committed tree; the second provides read-only formatting/static-analysis checks with generated-code exclusions. `go.mod` may change only if clean dependency resolution demonstrates a required correction, with the Go 1.22 floor preserved.

Task 2 modifies `pkg/mapper/schema.go`, `schema_test.go`, and configuration semantics in `mapper.go`; it adds focused cases to existing tests rather than moving all schema code. Task 3 modifies `mapper.go`, `parser.go`, and their tests, adding `pkg/mapper/concurrency_test.go` for observable race/ownership behavior.

Task 4 adds `pkg/bindings/registry.go`, `registry_test.go`, `python/tests/test_native_abi.py`, and `tests/native/memory.c`; it repairs `pkg/bindings/bindings.go`. Registry code contains only Go state, permitting Go race tests without `import "C"` in test files. The small C harness exercises allocation/free pairs directly. Task 5 modifies the Python wrapper and tests and adds `python/tests/test_native_mapper.py` for real-library lifecycle tests.

Task 6 keeps command dispatch in `cmd/main.go`, adds `cmd/cli.go` for a testable CLI runner, and adds `cmd/cli_test.go`. Demo content remains in its current location except for changes necessary to keep examples truthful.

Task 7 adds `internal/buildinfo/version.go`, `python/build_support.py`, `python/MANIFEST.in`, and `tools/check_package.py`; it modifies version consumers, Python metadata/setup, and Make packaging targets. The existing generated C header, if tracked, is refreshed by the documented header-generation step, not hand-edited. No generated native library is committed.

Task 8 adds `testdata/baseline/cases.json`, `testdata/baseline/mappings.json`, `tests/acceptance/test_surfaces.py`, `tests/acceptance/test_documented_cli.py`, `tests/acceptance/cli_examples.json`, `pkg/api/parity_test.go`, `pkg/mapper/baseline_test.go`, and `pkg/mapper/benchmark_test.go`. It modifies relevant README/docs/Python examples and adds `docs/compatibility.md`, `docs/performance.md`, and `docs/cli.md`. Only demonstrated adapter defects justify edits to `pkg/api/handlers.go`, `models.go`, or `server.go`.

Task 9 adds `tools/release.py`, `tools/check_reproducible.py`, `tools/tests/test_release.py`, and `tools/release-env.json`; modifies release targets, both existing Dockerfiles, and the package acceptance tests created in Task 7; and adds a tracked `.dockerignore` with its exact `.gitignore` exception. Task 10 replaces relevant CI jobs in `.github/workflows/ci.yml`, adds `tools/check_acceptance.py` and `tools/tests/test_acceptance.py`, and writes the temporary milestone log. Avoid changes to the unrelated documentation-deployment workflow.

## Interfaces and Dependencies


Preserved Go interfaces are `New() *Mapper`, `NewWithConfig(*MappingConfig) *Mapper`, `(*Mapper).LoadMappings([]byte) error`, `MapQuery(string) (string, error)`, `MapQueryWithContext(string, map[string]interface{}) (string, error)`, `DiscoverQuery(string) (*QueryInfo, error)`, `ValidateQuery(string) error`, and `(*Parser).Parse(string) (*ASTNode, error)`. Preserve all seven `QueryInfo` categories and JSON keys: `datamodels`, `datasets`, `lookups`, `macros`, `sources`, `sourcetypes`, and `input_fields`.

Task 2 produces `(*MappingConfig).matchingRuleMappings(map[string]interface{}) []FieldMapping` as an unexported helper returning only the first rule's mappings. Public `GetMappingsForConditions` returns base mappings followed by that helper's result. Task 3 uses the helper directly when overlaying current mutable base mappings, preventing old config mappings from undoing `LoadMappings` updates.

Task 4 produces `type mapperRegistry`, `newMapperRegistry() *mapperRegistry`, `(*mapperRegistry).add(*mapper.Mapper) (int, bool)`, `get(int) (*mapper.Mapper, bool)`, and `remove(int)`. The counter never reuses IDs; return `false` after maximum signed 32-bit handle capacity rather than wrap. Exported constructors return -1 on handle exhaustion or configuration-validation failure. It also exports `void spl_string_free(char*)` for caller-owned standalone strings, without changing existing structs. No new recoverable out-of-memory contract is added to cgo allocation calls.

Task 5 produces `SPLMapper.close() -> None`, `__enter__() -> SPLMapper`, and `__exit__(exc_type, exc, traceback) -> None`. Closed operations raise existing `MapperNotFoundError`, a subclass of `SPLMapperError`. Configuration loading errors use `ConfigurationError`; parse failures continue to use `ParseError`. Do not create a new exception hierarchy.

Task 6 produces `runCLI(args []string, stdout, stderr io.Writer) int` in package main. `main` calls it with `os.Args[1:]` and exits with its result. File I/O stays real and testable using temporary directories. No new CLI dependency is needed.

Task 7 produces `internal/buildinfo.Version`, initialized to `"dev"` for direct builds and injected from root VERSION for release builds, plus `char* spl_toolkit_version(void)` whose returned allocated string is freed with `spl_string_free`. Python exposes `__version__` from installed package metadata and a read-only `SPLMapper.native_version: str`, and verifies native/package version agreement during native initialization; direct uninstalled source use can report `dev`.

The Python packaging helpers are `source_root(setup_dir: Path) -> Path`, `read_version(setup_dir: Path) -> str`, and `build_native(source: Path, output: Path, version: str) -> None`. Custom setuptools commands `BuildPy`, `SourceDistribution`, and `BinaryWheel` live in `python/build_support.py`. No import of `spl_toolkit` is required to build metadata.

The task tools expose command-line interfaces as follows. Every command returns nonzero on failed checks; tools do not import this plan.

    python tools/check_go.py
    python tools/check_clean_build.py --ref HEAD
    python tools/check_package.py --sdist dist/spl_toolkit-0.1.1.tar.gz --wheel-dir dist
    python tools/release.py --source . --output dist --epoch 1788652800
    python tools/check_reproducible.py --ref HEAD --output build/reproducibility
    python tools/check_acceptance.py --evidence build/acceptance --ref HEAD

The epoch above is an example command argument, not an authored build timestamp. For release commands use the source commit timestamp obtained by `git show -s --format=%ct HEAD`; the check tools calculate it when only a ref is supplied.

Release dependency pins are `pip==25.2`, `setuptools==80.9.0`, `wheel==0.45.1`, `build==1.3.0`, `packaging==25.0`, `pyproject-hooks==1.2.0`, and `colorama==0.4.6` on Windows. Test pins are `pytest==8.4.2`, `pytest-mock==3.15.1`, `iniconfig==2.1.0`, `pluggy==1.6.0`, and `pygments==2.19.2`, plus the shared packaging/colorama pins. Put the minimal complete build set in `python/requirements-build.txt` and the test set, including `-r requirements-build.txt`, in `python/requirements-dev.txt`. Optional format/docs tools are not installed by normal build/test targets. `[build-system].requires` pins setuptools and wheel exactly; use `--no-isolation` only after installing the complete pinned build requirements in the dedicated environment.

Keep `github.com/antlr4-go/antlr/v4 v4.13.1`, `github.com/swaggo/swag/v2 v2.0.0-rc4`, and the existing module graph. Explicit API-document generation uses the same pinned swag version. Standard builds use committed generated parser/docs. Generated-parser vet warnings are documented and excluded by package-specific checks, not by editing generated goto statements.

The native runner matrix is `ubuntu-24.04`/amd64, `macos-15-intel`/amd64, `macos-15`/arm64, and `windows-2022`/amd64. macOS selects `/Applications/Xcode_16.4.app/Contents/Developer` with `MACOSX_DEPLOYMENT_TARGET=15.0`; Linux selects `gcc-13`; Windows selects the runner's MinGW-w64 x64 GCC explicitly and records its executable path/version. Check the expected architecture with `go env GOARCH` and `platform.machine()` before accepting any result. Record runner image version, full compiler version, linker, SDK, Python, Go, environment, and hashes in each release job.

## Plan of Work


Deliver the first independently useful slice with Task 1: reproducible dependency state and non-mutating Go builds. Tasks 2–3 make the core's existing mapping behavior deterministic and concurrent. Tasks 4–5 make native discovery and Python resource lifetimes correct. Task 6 exposes the approved CLI functionality. Task 7 creates installable native packages with a shared version identity. Task 8 proves parity through the real surfaces and replaces documentation claims with exercised examples. Tasks 9–10 produce controlled release artifacts and close the four-platform gates.

Run each task's failing tests before implementing its behavior. A missing compiler, import failure, or malformed test fixture is not evidence that the product regression was reproduced; resolve the prerequisite or test setup first. Commit only task-owned files after the stated checks pass. Task-local code snippets below are implementation anchors, not instructions to replace working code wholesale.

## Concrete Steps


### Task 1: Clean dependency state and non-mutating builds


Files: modify `Makefile`; add `go.sum`, `tools/check_clean_build.py`, and `tools/check_go.py`. Consumes the existing go.mod and generated parser. Produces build/test targets usable without dependency or source changes.

1. After plan approval, load using-git-worktrees and establish `codex/milestone-1-trustworthy-baseline` in an isolated checkout. If the app already supplied one, use it. Record `git rev-parse HEAD`, `git status --short`, and the original checkout's untracked inventory. Copy only this approved temporary plan into the worktree if it is not in history. Export HEAD to a new temporary directory and show that the original `go build -mod=readonly ./cmd` lacks tracked checksum state. Do not copy the user's untracked go.sum into the export.

    git rev-parse HEAD
    git status --short
    git ls-files go.sum
    go build -mod=readonly ./cmd

Run the last command in the pristine export, not the user's populated checkout. Capture expected missing-go.sum/dependency-checksum errors. If it unexpectedly passes from existing module cache, use fresh task-owned GOMODCACHE and GOCACHE directories and repeat only this baseline check.

2. Implement the clean-export checker using Python's standard library. It exports the requested committed ref, records hashes of all tracked regular files, runs the documented build/test commands, and verifies unchanged hashes. Use `git archive --format=tar REF` captured to a temporary file and extract the trusted local archive with checked relative paths. Build outputs are allowed only outside tracked source. This core assertion must remain literal:

    def hashes(root, names):
        import hashlib
        return {name: hashlib.sha256((root / name).read_bytes()).hexdigest()
                for name in names if (root / name).is_file()}

    before = hashes(source, tracked_names)
    subprocess.run(["make", "build", "build-server", "build-shared", "test"],
                   cwd=source, env=env, check=True)
    assert hashes(source, tracked_names) == before, "build changed tracked source"

Set `env["GOCACHE"]` and `env["GOMODCACHE"]` to task-owned directories, and `env["GOTOOLCHAIN"]="local"`. For a Windows environment without Make, the checker invokes the exact equivalent Go commands in step 3, while a separate Linux job exercises the documented Make targets. The temporary export checker returns the subprocess failure and changed filenames rather than catching errors as success.

3. In the isolated implementation tree only, run `go mod download all` with Go 1.26.8, inspect any changes, and commit the necessary go.sum. If readonly build still requests module changes, diagnose the exact module requirement; do not run unconditional tidy. Change Make defaults so `build`, `build-server`, `build-shared`, and `test` neither depend on `deps`/`fmt` nor install tools. Use these command forms, retaining current binary names and Windows suffix detection:

    go build -mod=readonly -trimpath -o build/spl-toolkit ./cmd
    go build -mod=readonly -trimpath -o build/spl-toolkit-server ./cmd/server
    go build -mod=readonly -trimpath -buildmode=c-shared -o build/libspl_toolkit.dylib ./pkg/bindings
    go test -mod=readonly -race ./...

Use `.so` on Linux and `.dll` on Windows for the shared library. Later Task 7 adds common version ldflags. Keep `deps` as an explicit `go mod download`, introduce `deps-update` for an intentional tidy, and keep `fmt` explicitly mutating. `lint` must be a read-only check. `generate-docs` must use `go run github.com/swaggo/swag/v2/cmd/swag@v2.0.0-rc4 init --v3.1 -g cmd/server/main.go -o docs` and is never a normal build dependency.

4. Add `tools/check_go.py`: enumerate tracked `.go` files outside `parser/`, `gen/`, and generated `docs/docs.go`, run `gofmt -l`, and fail if any handwritten file needs formatting. Planning inspection found no existing handwritten formatting differences, so no baseline waiver or mass reformat is needed. Obtain package names with `go list -mod=readonly ./...`, exclude the generated parser/gen/docs packages themselves, and run `go vet` on the remaining explicit package names. Vet on explicit packages does not vet their imported generated dependency packages. Run `go test -mod=readonly -race ./...`; if the test command itself enables a generated-code analyzer that fails on the pinned compiler, explicitly separate test execution with `-vet=off` and retain the positive handwritten `go vet` gate. Record the exact exclusion and reason, not a general static-analysis waiver.

Verify `make build build-server build-shared test`, `git diff --check`, and tracked-file hashes in a clean committed export. Stage Makefile, go.sum, the two tools, and their focused behavior regression tests, then commit `build: make baseline Go workflows non-mutating`. Run the export checker against that commit; expected exit 0 and no tracked-file changes.

### Task 2: Documented rule priority and working regex conditions


Files: `pkg/mapper/schema.go`, `schema_test.go`, and rule integration in `mapper.go`. Consumes `MappingConfig`, `ConditionalRule`, `Condition`, and `FieldMapping`. Produces `matchingRuleMappings` and strict supported condition validation without new mapping capabilities.

1. Add this regression to `schema_test.go` (package mapper, imports testing and reflect). It distinguishes first-match behavior from the current last-write outcome and proves that sorting does not mutate user order:

    func TestRulePriorityFirstMatch(t *testing.T) {
        cfg := &MappingConfig{Version: "1.0", Mappings: []FieldMapping{{Source:"src", Target:"base"}}}
        condition := []Condition{{Type:"source", Operator:"equals", Value:"web"}}
        cfg.Rules = []ConditionalRule{
            {ID:"low", Enabled:true, Priority:20, Conditions:condition, Mappings:[]FieldMapping{{Source:"src", Target:"low"}}},
            {ID:"winner", Enabled:true, Priority:1, Conditions:condition, Mappings:[]FieldMapping{{Source:"src", Target:"winner"}}},
            {ID:"tie", Enabled:true, Priority:1, Conditions:condition, Mappings:[]FieldMapping{{Source:"src", Target:"tie"}}},
        }
        got := cfg.GetMappingsForConditions(map[string]interface{}{"source":"web"})
        want := []FieldMapping{{Source:"src", Target:"base"}, {Source:"src", Target:"winner"}}
        if !reflect.DeepEqual(got,want) { t.Fatalf("got %#v, want %#v",got,want) }
        if cfg.Rules[0].ID != "low" { t.Fatal("evaluation reordered caller config") }
    }

Add table cases with the same configuration: winner disabled selects tie; context source `other` returns only base; missing source returns only base. Add table-driven regex tests using `^web_[0-9]+$` against `web_12`/`mail_12`, and sourcetype list `[]string{"mail_12","web_12"}`. Invalid regex `[` must fail `LoadMappingConfig`. A `field_value/and`, `combination/equals`, or unsupported `starts_with` condition must fail validation, not silently evaluate false. Empty/whitespace rule mapping source or target must also fail.

2. Run `go test -mod=readonly ./pkg/mapper -run 'TestRulePriority|TestRegex|TestSupportedCondition' -count=1`. Expected first-match and regex regressions fail for their intended assertions. Record exact failures.

3. Implement first-match evaluation by sorting a copied rule slice stably, checking enabled rules, and returning immediately. Use this anchor with imports sort and regexp:

    func (mc *MappingConfig) matchingRuleMappings(ctx map[string]interface{}) []FieldMapping {
        rules := append([]ConditionalRule(nil), mc.Rules...)
        sort.SliceStable(rules, func(i,j int) bool { return rules[i].Priority < rules[j].Priority })
        for _, rule := range rules {
            if rule.Enabled && mc.evaluateConditions(rule.Conditions, ctx) {
                return rule.Mappings
            }
        }
        return nil
    }

    func matchString(actual, expected, operator string) bool {
        switch operator {
        case "equals": return actual == expected
        case "contains": return strings.Contains(actual, expected)
        case "regex":
            pattern, err := regexp.Compile(expected)
            return err == nil && pattern.MatchString(actual)
        }
        return false
    }

Define supported type/operator pairs exactly: `field_exists` requires `exists|not_exists` and a field; `field_value` requires `equals|contains|regex` and a field; `source|sourcetype` requires `equals|contains|regex`; `combination` requires `and|or` and at least two recursively valid children. Equals accepts JSON scalar string/number/bool/null; contains/regex require a string condition value. For scalar equality, use a type switch or `reflect.DeepEqual` after rejecting compound configured values, avoiding interface equality panics on a caller-provided list/map. No implicit numeric coercion is added. Regex validation calls `regexp.Compile` before accepting configuration. Keep existing any-match source list behavior for both `[]string` and JSON-decoded `[]interface{}` string elements, so Python/REST context arrays match Go context arrays. Add both representations to the regex test table; ignore non-string array elements for string comparisons. Invalid direct method inputs must not panic.

Validate whitespace-only mapping names and nested rule mappings. Reject a nonempty `datamodels` rewrite configuration as unsupported instead of silently accepting ignored rewrite requests; leave its exported types for source compatibility and label actual data-model discovery separately. Make unsupported settings explicit in compatibility notes in Task 8.

4. Run all mapper tests, then full Go tests. Update only tests whose old all-match expectations conflict with the approved intentional behavior; do not rewrite unrelated expectations. Commit `fix: honor mapping rule priority and regex conditions` with the three task files.

### Task 3: Owned configuration and concurrent public Go use


Files: `pkg/mapper/mapper.go`, `parser.go`, their existing tests, and new `concurrency_test.go`. Consumes the rule-only helper from Task 2. Produces mapper-owned configuration, atomic base updates, and a reusable concurrent parser.

1. Add an ownership test that constructs a mapper with nested conditions, mutates the original base mapping, rule mapping, and nested child afterwards, and proves the mapper still produces its original output. Add this update regression:

    func TestLoadMappingsOverridesInitialBase(t *testing.T) {
        m := NewWithConfig(&MappingConfig{Version:"1.0", Mappings:[]FieldMapping{{Source:"src",Target:"old"}}})
        if err := m.LoadMappings([]byte(`[{"source":"src","target":"new"}]`)); err != nil { t.Fatal(err) }
        got, err := m.MapQueryWithContext("search src=1", map[string]interface{}{})
        if err != nil || got != "search new=1" { t.Fatalf("%q %v",got,err) }
    }

For atomicity, initialize `a->x,b->y`, run a writer alternating that pair with `a->p,b->q`, and readers mapping `search a=1 b=2`; only `search x=1 y=2` or `search p=1 q=2` is accepted. Start goroutines together with a closed channel, collect errors through a buffered channel, and wait with sync.WaitGroup. For parser reuse, run valid `search host=web` and invalid `|` parses concurrently on the same `NewParser()` for 100 iterations each; assert valid calls succeed and invalid calls fail independently. Add direct-constructor invalid-regex coverage: `MapQuery` must return configuration error, without a panic or signature change.

2. Run `go test -mod=readonly -race ./pkg/mapper -run 'TestLoadMappingsOverrides|TestOwned|TestAtomic|TestConcurrentParser|TestInvalidConstructor' -count=1`. Capture ownership/assertion failures and race reports separately.

3. Add a `sync.RWMutex` for mutable base mappings and `configErr error` to Mapper. Validate and deep-copy supported configuration values at construction. Copy slices recursively, including children and mapping slices; copy metadata maps/slices recursively for JSON-like values without marshaling errors being ignored. Configuration inputs outside the documented supported value domain produce `configErr`. `NewWithConfig(nil)` returns an empty mapper, preserving a useful non-panicking constructor behavior. Do not sort the original caller slice. In `LoadMappings`, decode and validate every entry into a temporary collection before acquiring the write lock; then update all mappings as one operation. A failed load leaves the old configuration intact.

    m.mu.Lock()
    defer m.mu.Unlock()
    for _, mapping := range mappings {
        m.fieldMappings[mapping.Source] = mapping.Target
    }

In `getEffectiveMappings`, copy the mutable base map under a read lock, release it, and overlay only `m.config.matchingRuleMappings(context)`. At the start of mapping operations return stored `configErr`, if any. Immutable owned configuration needs no per-query write lock. Read-only discovery remains independent of mapping configuration.

Change `Parser` to hold no mutable per-parse listener. In `Parse`, create the listener and attach it only to the new lexer/parser for that operation:

    listener := &CustomErrorListener{DefaultErrorListener: antlr.NewDefaultErrorListener()}
    lexer.AddErrorListener(listener)
    splParser.AddErrorListener(listener)

Check `listener.errors` instead of `p.errorListener.errors`. Existing parse trees/listeners stay local. Do not broaden the grammar or alter AST structures.

4. Run the targeted race cases and `go test -mod=readonly -race ./...`. Compare tracked source status and commit `fix: isolate mapper configuration and parser state`. No stress repetition is needed after these tests pass unless another change affects synchronization.

### Task 4: Safe C handles and complete native discovery arrays


Files: new `pkg/bindings/registry.go`, `registry_test.go`, `python/tests/test_native_abi.py`, `tests/native/memory.c`; modify `pkg/bindings/bindings.go`. Consumes concurrent mapper behavior. Produces the registry and standalone string release export defined above, preserving every existing C struct layout.

1. Test registry add/get/remove through pure Go test files. A removed ID never resolves again; an already obtained mapper reference remains callable; 32 goroutines each create/remove 100 mappers without duplicate IDs or races. Set the registry counter near signed-32-bit maximum in a package-local test and assert exhaustion returns false without wrapping.

Use native ctypes tests in a child process so a corrupt free crashes that child with a useful failing test result. Build the real library with `make build-shared`, then run the test file with the absolute path in `SPL_NATIVE_LIBRARY`. Define ctypes signatures matching existing `SPLQueryInfoC`; use `spl_mapper_discover_query` and always `spl_query_info_free` in finally. This query must return exactly the three named data models, comparing order only where the core guarantees it:

    THREE_MODELS = (
        "| tstats count from datamodel=Web by Web.status "
        "| tstats count from datamodel=Authentication by Authentication.user "
        "| tstats count from datamodel=Network_Traffic by Network_Traffic.src_ip"
    )
    MACROS = "`get_data` | eval result=`calculate_score(field1, field2)` | where result>10"
    assert set(read_models(THREE_MODELS)) == {"Web", "Authentication", "Network_Traffic"}
    assert set(read_macros(MACROS)) == {"get_data", "calculate_score"}

Define `read_models` and `read_macros` in that test file as calls to a shared local `discover(query)` helper that copies all seven arrays to Python lists before freeing the native result. It creates/removes its own mapper handle and fails on any nonempty native error; it does not use a mocked CDLL. Use SPL_NATIVE_LIBRARY during Task 4; in the installed suite obtain the native path from the installed spl_toolkit package directory, with no checkout fallback. Test repeated calls, empty arrays, invalid handle, null cleanup, and an invalid mapping JSON load whose returned pointer is copied and released.

2. Execute the existing native library tests before editing. The macros case fails because it returns an empty list; the three-model case may crash or corrupt data. Keep the child return code and stderr. Run pure-Go registry tests after adding the tests; missing registry methods are the expected preimplementation failure.

3. Implement the Go registry with RWMutex, map[int]*mapper.Mapper, and a monotonic int64 counter checked against 2147483647. All exported operations call `get` once and keep the returned pointer alive until completion; no global registry lock is held while parsing. Removal is idempotent. Replace each native string-array loop with the same base-preserving helper; this anchor is valid Go with cgo:

    func cStrings(values []string) **C.char {
        if len(values) == 0 { return nil }
        base := (**C.char)(C.malloc(C.size_t(len(values)) * C.size_t(unsafe.Sizeof(uintptr(0)))))
        items := unsafe.Slice(base, len(values))
        for i, value := range values { items[i] = C.CString(value) }
        return base
    }

Use `C.CString` exactly once per value, remove the double-copy allocator, populate macros, and set each count to the corresponding slice length. Initialize result structs before filling, retain existing result free functions, and add:

    //export spl_string_free
    func spl_string_free(value *C.char) { C.free(unsafe.Pointer(value)) }

Do not change C struct field order. Do not add `import "C"` to `_test.go`: Go does not support that test-file form. Direct allocation behavior is verified by the shared library and C harness.

4. Write the C harness with the generated header and this complete exercise, compiled without NDEBUG so assertions execute:

    #include "libspl_toolkit.h"
    #include <assert.h>
    #include <string.h>
    int main(void) {
        char models[] = "| tstats count from datamodel=Web by Web.status | tstats count from datamodel=Authentication by Authentication.user | tstats count from datamodel=Network_Traffic by Network_Traffic.src_ip";
        char macros[] = "`get_data` | eval result=`calculate_score(field1, field2)` | where result>10";
        for (int i = 0; i < 200; i++) {
            int handle = spl_mapper_new();
            assert(handle > 0);
            SPLQueryInfo *a = spl_mapper_discover_query(handle, models);
            assert(a && !a->error && a->data_models_count == 3);
            assert(strcmp(a->data_models[0], "Web") == 0);
            assert(strcmp(a->data_models[1], "Authentication") == 0);
            assert(strcmp(a->data_models[2], "Network_Traffic") == 0);
            spl_query_info_free(a);
            SPLQueryInfo *b = spl_mapper_discover_query(handle, macros);
            assert(b && !b->error && b->macros_count == 2);
            assert(strcmp(b->macros[0], "get_data") == 0);
            assert(strcmp(b->macros[1], "calculate_score") == 0);
            spl_query_info_free(b);
            char *error = spl_mapper_load_mappings(handle, "[");
            assert(error && strlen(error) > 0);
            spl_string_free(error);
            spl_mapper_free(handle);
            SPLResult *closed = spl_mapper_map_query(handle, "search src_ip=1");
            assert(closed && closed->error);
            spl_result_free(closed);
            spl_mapper_free(handle);
        }
        spl_result_free(NULL);
        spl_query_info_free(NULL);
        spl_string_free(NULL);
        return 0;
    }

Compile and run it against the actual library. On Linux add the Task 10 AddressSanitizer job with a Go `-asan` shared library and an `-fsanitize=address` C harness; use `ASAN_OPTIONS=detect_leaks=1:halt_on_error=1`. No broad leak suppressions are allowed; attributable process-runtime allocations must be distinguished from binding allocations with recorded stacks.

Run `go test -mod=readonly -race ./pkg/bindings`, the native ctypes tests, and the harness. Expected complete arrays, valid cleanup, and no registry races. Commit `fix: preserve native result ownership and synchronize handles` with only Task 4 files.

### Task 5: Explicit Python lifecycle and real native calls


Files: `python/spl_toolkit/mapper.py`, `python/tests/test_mapper.py`, new `python/tests/test_native_mapper.py`; use existing `exceptions.py` classes. Consumes `spl_string_free`, complete native arrays, and the registry's held-reference semantics. Produces close/context-manager behavior and per-instance lifecycle synchronization without serializing all mapper instances.

1. Add real-library tests with this exact representative behavior. Before Task 7, pass the absolute `SPL_NATIVE_LIBRARY` environment value as `library_path`; after installation, the same tests omit the override to verify package-local lookup.

    def test_close_and_context_manager():
        kwargs = {"library_path": os.environ["SPL_NATIVE_LIBRARY"]} if "SPL_NATIVE_LIBRARY" in os.environ else {}
        mapper = SPLMapper(**kwargs)
        with mapper as active:
            assert active is mapper
            active.load_mappings([{"source": "src_ip", "target": "source_ip"}])
            assert active.map_query("search src_ip=1") == "search source_ip=1"
        mapper.close()
        with pytest.raises(MapperNotFoundError):
            mapper.map_query("search src_ip=1")

    def test_json_context_array_matches():
        config = {"version":"1.0", "mappings":[], "rules":[{
            "id":"web", "enabled":True, "priority":1,
            "conditions":[{"type":"sourcetype","operator":"regex","value":"^web_"}],
            "mappings":[{"source":"src_ip","target":"client_ip"}]}]}
        kwargs = {"library_path": os.environ["SPL_NATIVE_LIBRARY"]} if "SPL_NATIVE_LIBRARY" in os.environ else {}
        with SPLMapper(config=config, **kwargs) as mapper:
            assert mapper.map_query_with_context("search src_ip=1", {"sourcetype":["mail","web_access"]}) == "search client_ip=1"

Add repeated invalid `load_mappings([{"source":1,"target":"bad"}])` calls, asserting `ConfigurationError`; a valid call must still succeed afterwards. Add 16-worker ThreadPoolExecutor tests sharing one mapper for mapping/discovery; add separate-instance construction/destruction tests. For concurrent close, coordinate workers using threading.Event: calls already inside the operation helper finish, new calls receive MapperNotFoundError, multiple close callers return, and the underlying free is called exactly once. A small mock test may hold an operation with an Event to make that boundary deterministic, but the normal concurrent workload also uses the native library.

2. Run `python -m pytest python/tests/test_native_mapper.py -q` with a real library path. Expect missing close/context-manager behavior before implementation. Existing mock tests remain separately executable; fix mock argument setup when new symbols are configured rather than removing assertions.

3. Initialize a `threading.Condition`, `_active_calls=0`, `_closing=False`, `_closed=False`, and `_mapper_id=None` before loading the library. Add a private context manager `_operation` that admits calls under the condition, rejects closed/closing state, increments the count, yields the handle, then decrements and notifies in finally. The native call and result-copy/free occur outside the condition lock so read operations may run concurrently:

    @contextmanager
    def _operation(self):
        with self._condition:
            if self._closing or self._closed or self._mapper_id is None:
                raise MapperNotFoundError("Mapper is closed")
            self._active_calls += 1
            handle = self._mapper_id
        try:
            yield handle
        finally:
            with self._condition:
                self._active_calls -= 1
                self._condition.notify_all()

Wrap load/map/map-with-context/discover using this helper. `get_input_fields` delegates to discovery instead of starting a redundant nested operation. `close` sets closing, blocks new operations, waits for active count zero, takes and clears the handle, frees it exactly once, marks closed, and notifies other closers. A second closer waits until closed if the first is still freeing. For partial construction with no handle it simply marks closed. `__enter__` rejects closed state and returns self; `__exit__` closes and returns None. `__del__` defensively closes without raising at interpreter shutdown.

Set standalone string function signatures and release errors after copying:

    self._lib.spl_mapper_load_mappings.restype = ctypes.c_void_p
    self._lib.spl_string_free.argtypes = [ctypes.c_void_p]
    self._lib.spl_string_free.restype = None
    if error_pointer:
        try:
            message = ctypes.string_at(error_pointer).decode("utf-8")
        finally:
            self._lib.spl_string_free(error_pointer)
        raise ConfigurationError(message)

Do not use c_void_p for result struct fields that are freed with the enclosing result; existing struct pointer/free logic is appropriate there. Default library lookup checks only the package directory for the current platform's suffix. Keep explicit library_path support. Update mock tests to pass a path or mock the existence check intentionally; no arbitrary current-directory shared library fallback remains.

4. Run the complete Python unit/native suite with the built library and rerun the C harness under the memory checker when available. Expected no hangs, no double-free, all copied arrays intact, clear closed errors, and native error allocations released. Commit `fix: make Python native mapper lifecycle explicit`.

### Task 6: CLI configuration, input, formatting, and exit contracts


Files: `cmd/main.go`, new `cmd/cli.go`, `cmd/cli_test.go`. Consumes mapper configuration loading and existing operations. Produces `runCLI(args []string, stdout, stderr io.Writer) int` and the approved flags, without adding a CLI framework dependency.

1. Add table-driven tests calling runCLI with buffers and temporary config files. This complete test is the initial failing case (imports bytes, os, filepath, strings, testing):

    func TestCLIMapLoadsConfig(t *testing.T) {
        path := filepath.Join(t.TempDir(), "mapping.json")
        if err := os.WriteFile(path, []byte(`{"version":"1.0","mappings":[{"source":"src_ip","target":"source_ip"}]}`), 0600); err != nil { t.Fatal(err) }
        var out, errs bytes.Buffer
        code := runCLI([]string{"map","--config",path,"--query","search src_ip=1"}, &out, &errs)
        if code != 0 || out.String() != "search source_ip=1\n" || errs.Len() != 0 {
            t.Fatalf("code=%d out=%q err=%q",code,out.String(),errs.String())
        }
    }

Use explicit rows: map without config => 2; missing config file => 2; malformed config => 1; invalid regex => 1; invalid query `|` => 1; positional query plus --query => 2; unknown flag => 2; extra positional => 2; unsupported format => 2. Valid config with no matching fields => unchanged query and 0. `validate --config FILE` succeeds without a query; validate with both inputs => 2. `discover --config` => 2. Test options before and after a positional query, `--name=value`, and `--` termination for a literal query beginning with a dash.

Test JSON map exactly `{"query":"search source_ip=1"}`; validate JSON exactly `{"target":"query","valid":true}` or target `configuration`. Discovery retains existing seven JSON fields and current empty/null representation. JSON errors are `{"error":"MESSAGE"}` on stderr and no stdout. Format detection for errors uses a successfully parsed --format; if argument parsing fails before a valid format can be determined, emit a text usage error. Test zero-query help/version, output-file content and empty stdout, output failure => 2, and an existing output file preserved when input is rejected.

2. Run `go test -mod=readonly ./cmd -run TestCLI -count=1`, initially failing because runCLI is absent. Keep a binary subprocess test after implementation to prove real process exit codes, rather than only return values.

3. Reduce main to `os.Exit(runCLI(os.Args[1:], os.Stdout, os.Stderr))`. Keep help/demo printing testable by giving their print functions an io.Writer, with no os.Exit/log.Fatal inside the dispatcher. Implement a small explicit option parser for the four named value flags plus help. It consumes complete value tokens, rejects duplicates/missing values/unknown names, supports `--flag=value`, and distinguishes the single positional query. Do not use strings.Fields on query text.

Use a local result value and encode only after computation succeeds. For example:

    mapped, err := m.MapQuery(query)
    if err != nil { return writeCLIError(stderr, format, err.Error(), 1) }
    var payload []byte
    if format == "json" {
        payload, err = json.Marshal(struct { Query string `json:"query"` }{mapped})
        payload = append(payload, '\n')
    } else { payload = []byte(mapped + "\n") }

`writeCLIError(w io.Writer, format, message string, code int) int` serializes JSON or prints text to the supplied error writer and returns the selected nonzero code. `writeCLIResult(payload []byte, output string, stdout io.Writer) error` writes stdout or calls os.WriteFile only after a successful result; it propagates writer failures. Text discovery lists `Data models`, `Datasets`, `Lookups`, `Macros`, `Sources`, `Source types`, and `Input fields` in that order with comma-separated values, using `(none)` for empty lists. Syntax/configuration validation call the existing dedicated APIs. Do not introduce the future analysis tri-state result.

4. Run all CLI tests and the binary subprocess cases, then Go tests. Commit `feat: make documented CLI mapping and output options work`. Task 8 updates current documentation comprehensively; this commit includes CLI help matching the implemented contract.

### Task 7: One version identity and independently buildable Python packages


Files: new `internal/buildinfo/version.go`, `python/build_support.py`, `python/MANIFEST.in`, `python/requirements-build.txt`, `tools/check_package.py`, `tools/tests/test_package.py`; modify `Makefile`, `cmd/main.go`, `cmd/server/main.go`, `pkg/bindings/bindings.go`, Python `setup.py`, `pyproject.toml`, `requirements-dev.txt`, `spl_toolkit/__init__.py`, and wrapper version setup. Refresh the tracked native header if applicable. Consumes working Go/C/Python operations. Produces wheel and sdist artifacts whose versions and native payloads agree.

1. Add package acceptance that builds a wheel, installs it in a new venv outside the checkout, runs native tests under `-I`, and checks `importlib.metadata.version("spl-toolkit") == spl_toolkit.__version__ == mapper.native_version`. The read-only `native_version` attribute is populated during wrapper initialization by copying and freeing `spl_toolkit_version()`; it is a string, not a new analysis API. Install once with Go removed from PATH to prove wheel installation needs no compiler. This subprocess assertion is mandatory:

    script = """
    import importlib.metadata, pathlib, sys
    import spl_toolkit
    from spl_toolkit import SPLMapper
    assert pathlib.Path(spl_toolkit.__file__).resolve().is_relative_to(pathlib.Path(sys.prefix).resolve())
    with SPLMapper() as mapper:
        assert mapper.native_version == spl_toolkit.__version__
        assert spl_toolkit.__version__ == importlib.metadata.version('spl-toolkit')
        mapper.load_mappings([{'source':'src_ip','target':'source_ip'}])
        assert mapper.map_query('search src_ip=1') == 'search source_ip=1'
    """
    subprocess.run([str(venv_python), "-I", "-c", textwrap.dedent(script)], cwd=outside_checkout, check=True)

Build an sdist, unpack it into another external directory, verify no prebuilt native library is present, build a wheel there, and run the same installed-package checks. Check metadata for Requires-Python >=3.11, non-any platform tag, license/readme, native suffix/architecture, and version. Intentionally make the compiler executable unavailable when building a source wheel; packaging must fail and produce no installable artifact. Rebuilding must not reuse an unrelated source-directory library. These are behavioral package tests, not assertions on setup.py text.

2. Run existing packaging under the new acceptance checker and record its actual failure outside the source checkout. Do not equate a repository import with a wheel install. The checker may be initially written with individual subprocess calls until the package build is fixed; keep the final tool reusable.

3. Add a standard internal version package and use one injected symbol in all Go entry points:

    package buildinfo
    var Version = "dev"

    //export spl_toolkit_version
    func spl_toolkit_version() *C.char { return C.CString(buildinfo.Version) }

Build with `-ldflags=-X=github.com/delgado-jacob/spl-toolkit/internal/buildinfo.Version=0.1.1`, where the actual value is read from root VERSION. Do not bump VERSION during this milestone. Remove duplicated main.Version defaults in favor of buildinfo.Version; preserve NewServerWithVersion(version) and its current dynamic OpenAPI version behavior. Test `/api/v1/health` and `/api/v1/openapi.json` against the injected value. Generated OpenAPI source still gets refreshed only explicitly; runtime reads remain per-server and do not mutate shared SwaggerInfo during requests.

4. Consolidate Python project metadata in pyproject.toml; declare `dynamic=["version"]`, `requires-python=">=3.11"`, exact build requirements, and classifiers only for the chosen tested interpreter minors. setup.py supplies the version from read_version and registers build commands, without building native code at import/metadata time or duplicating project metadata. Runtime `__version__` uses importlib.metadata with a `dev` fallback only when uninstalled.

Replace the dummy extension and warning-only builder with a custom distribution marking native payloads and a BinaryWheel command returning `("py3", "none", platform_tag)`. BuildPy first copies Python modules, then builds the native library directly into its `build_lib/spl_toolkit` directory. It never copies an existing arbitrary `.so/.dylib/.dll` from source. `build_native` uses subprocess arguments as a list, checked return status, `CGO_ENABLED=1`, `GOTOOLCHAIN=local`, and read-only modules:

    subprocess.run([
        "go", "build", "-mod=readonly", "-trimpath", "-buildvcs=false",
        "-buildmode=c-shared", "-ldflags",
        f"-X=github.com/delgado-jacob/spl-toolkit/internal/buildinfo.Version={version}",
        "-o", str(output.resolve()), "./pkg/bindings"
    ], cwd=source, env=env, check=True)
    if not output.is_file():
        raise RuntimeError("native build produced no shared library")

Task 9 adds deterministic external-linker flags through the same build_native path; do not create a second divergent native builder. Parse the native version return with c_void_p and free it; reject a bundled library whose version disagrees with installed metadata. Explicit custom library_path follows the same version check for released packages; source `dev` mode reports its library version without fabricating release equality.

5. Make SourceDistribution stage a self-contained tree. Its root contains the Python package, setup.py, pyproject.toml, build_support.py, requirements-build.txt, README.md, LICENSE, VERSION, and `_native_src/`. The native source subtree contains go.mod, go.sum, `pkg/mapper/`, `pkg/bindings/`, `parser/`, and `internal/buildinfo/`, with license notices retained. Exclude `.git`, user files, caches, test output, built native libraries, and `_build_plan/`. `source_root` checks the original checkout's parent go.mod first when present, otherwise `_native_src/go.mod`; failure names the missing source rather than packaging anyway. `read_version` uses original root VERSION or the staged sdist VERSION.

Override SourceDistribution.make_release_tree to copy an explicit native source-file allowlist into its temporary release tree. The allowlist must remain available from a Git export and extracted sdist without .git; recursive directory copying must not include unrelated user or editor files. Acceptance rejects unexpected archive members. Include build_support and the required package inputs through MANIFEST.in. Do not permanently duplicate Go source into the repository. When building a wheel from an extracted sdist, every helper comes from the sdist; no absolute original path or Git history is needed.

Use `python -m build --no-isolation --sdist --wheel --outdir dist python` after installing pinned requirements. Change python-build/python-wheel/python-sdist/python-install targets accordingly. `python-install` installs a built wheel, avoiding an editable-install contract that could omit the binary; document that developers rebuild/reinstall after wrapper/native changes. `python-test` builds/installs into a dedicated test venv and runs copied tests outside the tree, not by modifying sys.path. Keep pure mock tests separate from native-required tests; missing libraries fail native-required suites rather than skip.

6. Run `python tools/check_package.py --sdist dist/spl_toolkit-0.1.1.tar.gz --wheel-dir dist` in the actual isolated checkout (normalize the chosen distribution filenames consistently). The tool locates exactly one wheel for the target, creates two dedicated environments with their own pinned test dependencies for original-wheel and sdist-built-wheel checks (no outer-site-packages .pth exposure), removes PYTHONPATH, invokes `python -I -m pytest` on copied tests, and verifies no tracked files changed. It accepts an optional `--expected-version` value for version-override tests in a temporary source export, without rewriting user VERSION.

Run `make build build-server build-shared`, native tests, and package acceptance. Commit `build: ship self-contained native Python packages with one version`. Preserve explicit build failures if a platform cannot produce the library; do not generate a pure wheel fallback.

### Task 8: Cross-surface parity, executable documentation, and performance baseline


Files: common fixtures, Go/API parity tests, Python acceptance tests, CLI example manifest, benchmark file, and documentation paths listed under File Structure. Consumes CLI runner, native package build, and shared version identity. Produces evidence that user-visible examples and transport results reflect the actual core.

1. Create `testdata/baseline/cases.json` with a small versioned fixture schema, consumed only by tests. Each case has id, query, optional config/context, expected mapped query when applicable, and optional expected seven-category discovery object. Start with these independently meaningful cases:

    [
      {"id":"basic-map","query":"search src_ip=1","config":{"version":"1.0","mappings":[{"source":"src_ip","target":"source_ip"}]},"mapped":"search source_ip=1"},
      {"id":"macros","query":"`get_data` | eval result=`calculate_score(field1, field2)` | where result>10","discovery":{"datamodels":[],"datasets":[],"lookups":[],"macros":["get_data","calculate_score"],"sources":[],"sourcetypes":[],"input_fields":[]}},
      {"id":"three-models","query":"| tstats count from datamodel=Web by Web.status | tstats count from datamodel=Authentication by Authentication.user | tstats count from datamodel=Network_Traffic by Network_Traffic.src_ip","discovery":{"datamodels":["Web","Authentication","Network_Traffic"],"datasets":[],"lookups":[],"macros":[],"sources":[],"sourcetypes":[],"input_fields":["Web.status","Authentication.user","Network_Traffic.src_ip"]}}
    ]

Validate each asserted empty category against a Go baseline run before freezing it; this does not permit changing expected mappings to match a regression. Keep these two discovery fixtures separate because current macro recovery deliberately returns only macros after some parse failures. Add supported source/sourcetype/lookup examples from existing mapper tests, preserving their actual limitations. Add first-match rule and JSON-array-context mapping fixtures. CLI does not add a context flag: context-bearing fixtures compare Go/Python/REST, and the equivalent source/sourcetype query fixture covers CLI automatic context extraction.

Define test-local normalization that maps nil/null and empty lists to empty lists and sorts category values for set comparison. Do not change production ordering/JSON fields simply to simplify assertions. Mapped query text is compared byte-for-byte, including preserved whitespace.

2. Add `TestBaselineSurfaceFixtures` in `pkg/mapper/baseline_test.go` to load the common JSON and assert Go results. Add `TestHTTPBaselineParity` in pkg/api using httptest.NewServer(NewServerWithVersion("0.1.1").Handler()), sending real JSON requests to `/api/v1/query/map` and `/api/v1/query/discover`. This request/response anchor verifies actual mapped text, unlike existing success-only tests:

    request := `{"query":"search src_ip=1","mappings":[{"source":"src_ip","target":"source_ip"}]}`
    response, err := http.Post(server.URL+"/api/v1/query/map", "application/json", strings.NewReader(request))
    if err != nil { t.Fatal(err) }
    defer response.Body.Close()
    var body MapQueryResponse
    if err := json.NewDecoder(response.Body).Decode(&body); err != nil { t.Fatal(err) }
    if response.StatusCode != 200 || body.MappedQuery != "search source_ip=1" { t.Fatalf("%d %#v",response.StatusCode,body) }

Exercise parallel HTTP requests with distinct configs to ensure cache entries do not leak mappings across callers. Test invalid regex returns HTTP 400 through existing configuration validation and syntax rejection returns existing 422 on `/query/validate`. Keep admin endpoints disabled by default and the existing fallback API behavior; the CLI's required config is not silently imposed on the Go/REST APIs. If existing handlers already pass, add tests without refactoring the server.

3. In tests/acceptance/test_surfaces.py, require `SPL_CLI`, `SPL_SERVER`, and `SPL_FIXTURES` absolute environment paths. Missing required variables fail the acceptance invocation. Start the server on a temporary available loopback port, poll `/api/v1/health` with a bounded deadline, compare real CLI/Python/HTTP results, and terminate only the child server in finally. If port allocation races, retry server startup with a new port at most three times; do not touch an existing user server. Verify health and OpenAPI version. Use the standard library urllib client, not a new requests dependency. Installed acceptance runs in the external venv prepared by check_package, with the acceptance tests and fixtures copied there.

4. Create `docs/cli.md` as the canonical current CLI usage page and a concrete JSON manifest in `tests/acceptance/cli_examples.json`. Each entry contains `id`, `argv`, `exit`, `stdout`, optional `stderr_contains`, optional fixture file setup, and optional expected output-file content. `argv` is a JSON argument list, never a shell command. The checker substitutes only explicit `{cli}` and `{tmp}` tokens and writes fixture files to its own temporary directory. Execute with subprocess.run without shell=True; assert return code and exact output or parsed JSON as appropriate. The canonical documentation references the corresponding case ID in an adjacent HTML comment, for example `<!-- cli-example: map-basic -->`. A coverage test verifies every `cli-example` ID exists and every current spl-toolkit command block in README/docs uses an ID or links to canonical usage. Roadmap prose and historical files under docs/superpowers are excluded from current-user command scanning.

Use these actual supported examples and expected outcomes, including output-file behavior:

    spl-toolkit map --config testdata/baseline/mappings.json --query 'search src_ip=1'
    spl-toolkit discover --query 'search sourcetype=web src_ip=1' --format json
    spl-toolkit validate --query 'search src_ip=1'
    spl-toolkit validate --config testdata/baseline/mappings.json --format json
    spl-toolkit discover --query 'search src_ip=1' --format json --output discovery.json
    spl-toolkit version
    spl-toolkit help
    spl-toolkit demo

Basic map outputs `search source_ip=1`; text validation outputs `Valid`; config JSON validation returns target configuration and valid true; output-file examples emit no stdout. Demo must exit nonzero if any demonstrated operation fails; it may not log an error and continue to apparent success. Preserve testdata as a permanent example/fixture location, never point current docs into `_build_plan/`.

5. Audit README.md and all current docs pages, especially quickstart, installation, configuration, API Go examples, server docs, architecture, contributing, and Python examples. Fix API calls, error checks, mapping flags, Python baseline, and build commands to match the implemented surfaces. Remove current or roadmap promises for automatic raw/data-model translation, learned mappings, module support, and throughput that the approved PRD excludes. Retain a concise roadmap for the actual later milestones, without a runtime link dependency on temporary files. Explain flat InputFields and macro-only recovery limitations; do not claim semantic completeness, general rewrite safety, full Splunk syntax coverage, or schema validation. Mark unsupported condition operators and data-model rewrite configuration as unsupported; document first-match precedence, stable ties, required CLI config, Python close, and version mismatch behavior as compatibility corrections. REST Swagger UI's existing CDN assets need a browser network connection; document that distinction from offline REST processing rather than introducing a UI asset project.

6. Add fixed-fixture benchmarks in pkg/mapper/benchmark_test.go using the existing public APIs. At minimum include Parse, Discover, Map, and ParallelMap. A benchmark must fail on operation error:

    func BenchmarkBaselineMap(b *testing.B) {
        m := New()
        if err := m.LoadMappings([]byte(`[{"source":"src_ip","target":"source_ip"}]`)); err != nil { b.Fatal(err) }
        b.ReportAllocs()
        b.ResetTimer()
        for i := 0; i < b.N; i++ {
            if _, err := m.MapQuery("search src_ip=1 | stats count by src_ip"); err != nil { b.Fatal(err) }
        }
    }

ParallelMap uses b.RunParallel with the same preconfigured mapper; Parse and Discover use the existing simple supported fixture. Run `go test -mod=readonly ./pkg/mapper -run '^$' -bench '^BenchmarkBaseline' -benchmem -count=5` once on the final relevant code. Store raw output and command/environment in build evidence; summarize repetitions, ns/op, B/op, and allocs/op in docs/performance.md with the exact code SHA, CPU, OS, Go version, and queries. No invented performance threshold or comparison to unmeasured old claims. If native/core code changes afterwards, refresh only affected benchmark evidence.

7. Run Go/API parity tests, installed native acceptance, and `python -m pytest tests/acceptance/test_documented_cli.py -q` with the built CLI path. Run a temporary Go example compilation and Python example execution where docs claim runnable snippets; snippet error-handling must compile rather than be pseudocode. Commit `test: verify surface parity and document the implemented baseline`. Record benchmark code SHA accurately; a subsequent documentation-only commit does not invalidate the measured code if ancestry and unchanged relevant paths are recorded.

### Task 9: Controlled release builds and artifact equality


Files: `tools/release.py`, `tools/check_reproducible.py`, `tools/tests/test_release.py`, `tools/release-env.json`, `Makefile`, both Dockerfiles, and deterministic build options in `python/build_support.py`. Consumes package builders and version injection. Produces checksummed release artifacts from two independent source trees in a recorded environment.

1. Add tests for archive normalization and mismatch detection before implementing the release tools. Expose `normalize_archive(path: Path, epoch: int) -> None` and `artifact_hashes(directory: Path) -> dict[str, str]` in tools/release.py. The hash function includes every release payload, excluding only explicitly separate evidence JSON and the checksum file itself. This test must fail until normalization is implemented:

    def test_wheel_metadata_is_reproducible(tmp_path):
        import zipfile
        from tools.release import normalize_archive
        first, second = tmp_path / "first.whl", tmp_path / "second.whl"
        for path, stamp in [(first,(2020,1,1,0,0,0)),(second,(2025,2,2,2,2,2))]:
            with zipfile.ZipFile(path,"w") as archive:
                info=zipfile.ZipInfo("spl_toolkit/native.dll",stamp)
                archive.writestr(info,b"same-native-payload")
        normalize_archive(first,1788652800)
        normalize_archive(second,1788652800)
        assert first.read_bytes() == second.read_bytes()

Add an sdist test with differing gzip/tar timestamps, owners, and input ordering that normalizes to equal bytes. Add a deliberately changed native payload to the second directory and assert the comparison fails even after archive normalization. Add a missing-wheel test: release.py must fail rather than emit only Go artifacts. Tests use small synthetic archives to prove tool behavior; final reproducibility acceptance uses real binaries.

2. Implement tools/release.py as a small CLI orchestrator, not a general build framework. It reads root VERSION and tools/release-env.json, validates expected compiler/interpreter, checks source inputs, creates a new output directory, builds CLI/server/shared library, calls `python -m build` for sdist/wheel, normalizes archive metadata, writes native payloads and C header, and produces a sorted SHA256SUMS. Reuse build_native from python/build_support.py for the native payload so package and standalone libraries use identical compiler/linker options. Keep build logs and environment evidence in a sibling evidence directory, outside the hashed payload set.

Build CLI/server with `CGO_ENABLED=0`, `-trimpath`, `-buildvcs=false`, injected version, and an empty Go build ID. Native builds use `CGO_ENABLED=1`, the same Go settings, and deterministic external-linker flags: Linux `-Wl,--build-id=none`, Windows `-Wl,--no-insert-timestamp`, macOS `-Wl,-reproducible`, retaining its content-derived UUID (do not use `-no_uuid`, which makes native dylibs unloadable on macOS 26). Pass these through Go's `-extldflags` within its ldflags argument using a Python list, not shell interpolation. Map source and temporary native compiler paths using `-ffile-prefix-map=ACTUAL_SOURCE=.` and `-fdebug-prefix-map=ACTUAL_SOURCE=.` (and the dedicated GOTMPDIR equivalent), substituting the actual resolved paths programmatically. Record all flags. If a platform linker rejects a proposed flag, diagnose against its actual compiler and adjust the platform-specific recipe with a regression/evidence note; never strip or alter native payloads after comparison solely to hide nondeterminism.

The Python builder sets `SOURCE_DATE_EPOCH` from the source commit. Wheel normalization sorts entries, sets deterministic timestamps, permissions, create_system, and compression settings without modifying member contents. Since ZIP metadata is outside wheel RECORD hashes, content-preserving normalization does not invalidate RECORD. Sdist normalization sets sorted tar member order, uid/gid 0, empty uname/gname, stable modes and mtime, clears volatile PAX metadata, and writes gzip with `filename=""` and the fixed epoch. Fail on unexpected absolute/traversal archive paths. Keep native library bytes unchanged during normalization.

Use this equality check in check_reproducible:

    left = artifact_hashes(first_output)
    right = artifact_hashes(second_output)
    if left.keys() != right.keys():
        raise RuntimeError("release artifact sets differ")
    different = [name for name in sorted(left) if left[name] != right[name]]
    if different:
        raise RuntimeError("non-reproducible artifacts: " + ", ".join(different))

Release artifact names include version and platform for the CLI/server/shared-library bundle, plus standard wheel/sdist filenames. Contents include LICENSE and applicable generated/parser attribution. Package metadata remains identical across both source roots. The environment report includes OS image identifier/version, architecture, Go/Python/packaging versions, CC path and full version output, linker/SDK selection, zlib version, SOURCE_DATE_EPOCH, GOFLAGS/CGO flags, and artifact SHA256 values. No user environment dump or secrets are captured.

3. Implement check_reproducible.py with `--ref` and `--output`: export the same committed ref twice to different task-owned directories, use independent output/GOCACHE/GOTMPDIR locations, and run release.py in each. GOMODCACHE may be warmed independently from the same verified pinned graph; neither pass copies compiled artifacts from the other. Both builds run on the same recorded runner and Python environment. Verify tracked-file hashes before/after. Compare the complete payload lists, then run check_package on the accepted wheel/sdist and extract/run the CLI/server/native payloads. Write `result.json` with `source_sha`, `target`, `status`, `artifact_hashes`, `environment`, and explicit failed checks. It is only `passed` after equality and installation checks succeed. Preserve failed artifacts for diagnosis under the task output directory.

4. Put this executable configuration in tools/release-env.json, consumed and validated by release.py, with dependency pins from Interfaces and Dependencies:

    {
      "go": "1.26.8",
      "go_language_floor_test": "1.22.12",
      "build_python": "3.11.9",
      "test_python": ["3.11.9", "3.12.10", "3.13.7", "3.14.0"],
      "targets": {
        "linux-amd64": {"runner":"ubuntu-24.04","goarch":"amd64","cc":"gcc-13","wheel_platform":"linux_x86_64"},
        "darwin-amd64": {"runner":"macos-15-intel","goarch":"amd64","cc":"clang","deployment_target":"15.0","developer_dir":"/Applications/Xcode_16.4.app/Contents/Developer","wheel_platform":"macosx_15_0_x86_64"},
        "darwin-arm64": {"runner":"macos-15","goarch":"arm64","cc":"clang","deployment_target":"15.0","developer_dir":"/Applications/Xcode_16.4.app/Contents/Developer","wheel_platform":"macosx_15_0_arm64"},
        "windows-amd64": {"runner":"windows-2022","goarch":"amd64","cc":"gcc","gcc_version":"14.2.0","wheel_platform":"win_amd64"}
      }
    }

Windows setup must choose the runner's MinGW-w64 GCC that reports 14.2.0 and x86_64-w64-mingw32; fail on a different target instead of using MSVC cl or a WSL compiler. Resolve its executable with `Get-Command gcc`, set CC to its absolute path, and include its bin directory in PATH when needed by Go. The reviewed runner image advertises this version; record the full resolved path in evidence. Linux records the exact GCC 13 patch/build from its runner. macOS checks selected Xcode and SDK. A disappearing pinned tool/runner is an infrastructure failure, not authorization to silently switch architecture or remove a gate.

5. Change Make release targets to call the controlled builder and stop swallowing artifact-copy failures. Keep `make release` local-only; `make tag` is never an automatic dependency. `make release-build` may build the current source for local inspection, while `check_reproducible --ref HEAD` is the authoritative committed-source acceptance path. `make python-test` and `make dev-test` must use the actual native suite from Task 7. Delete stale commented release workflow fragments only when replacing the same functionality in Task 10, not as unrelated cleanup.

6. Repair the existing Docker recipes so retained current Docker examples are true. Preserve CLI container/Python functionality and the separate server container. Use these exact manifest-list references, resolved from the public Docker Hub metadata during planning:

    golang:1.26.8-bookworm@sha256:9fdc884aacc3bec89b20ffc69f4bb369c78210e3e4f600387b5128b12c199f81
    python:3.11.9-slim-bookworm@sha256:8fb099199b9f2d70342674bd9dbccd3ed03a258f26bbd1d556822c6dfc60c317
    debian:bookworm-slim@sha256:88200866dfff7ea7f5cbcb6ec7c8a701889efe6fe859fe64d6990e4b07ea4171

The corresponding Linux amd64 child manifests are `sha256:bc6beb46032d45f421cf400036bf031cdc64f683ba9cdc124e31d063e71670bd`, `sha256:2856e6af199e8128161abd320575eb9b341f3b76f017b5d0c9cd364f60d8a050`, and `sha256:5ae3c39ebd15e229dcedd5cee596b2497182493d41ff162e824ba13fc1b2b867`. Docker acceptance uses linux/amd64 explicitly and verifies the resolved identities. Missing digests are setup failures, not permission to silently replace the pinned image.

Copy go.mod and go.sum before dependency download, use readonly build commands with VERSION injection, remove implicit tidy, and build/install the wheel rather than copying a source package with a missing native source tree. Keep a non-root runtime user. Server builds remain CGO-disabled and serve the existing health route. Add a task-owned `.dockerignore` by deliberately unignoring that exact path in `.gitignore`; exclude `.git`, `_build_plan`, caches, user untracked artifacts, and built libraries. Do not send unrelated local files to Docker's build context; run Docker acceptance from the clean exported tree.

Verify `docker build -t spl-toolkit-baseline-cli .`, CLI version and map with a mounted fixture, Python native mapping using `--entrypoint python`, and `docker build -f Dockerfile.server -t spl-toolkit-baseline-server .` followed by loopback health/map requests. Use a unique container name and published temporary port; stop/remove only that container. Docker functionality is checked on Linux; do not claim Docker image byte reproducibility as a milestone release artifact unless image creation itself is normalized and compared. The required checksum gate applies to the CLI/server/native files and Python archives explicitly enumerated above.

7. Run `python -m pytest tools/tests/test_release.py tools/tests/test_package.py -q` and the real two-build checker on the available matching pinned platform. Expected all payload hashes identical, successful install/rebuild, working container examples, and no tracked-file changes. An unmatched local compiler/interpreter may provide diagnostic evidence but does not pass the pinned platform gate. Commit `build: verify reproducible baseline release artifacts` with exact task-owned files and the recorded Docker digest choices.

### Task 10: Four-platform CI execution and milestone acceptance


Files: `.github/workflows/ci.yml`, `tools/check_acceptance.py`, `tools/tests/test_acceptance.py`, `tools/check_package.py` for wheel-only and required-test evidence, `tools/check_reproducible.py` for post-verification accepted-artifact promotion, their focused tests, and the final milestone log; relevant compatibility/performance evidence updates. Consumes all previous tasks. Produces mandatory platform/interpreter, reproducibility, and exact-source acceptance results with no silent skips.

1. Replace the current build/native test jobs with an explicit four-target release matrix and 16 installed-wheel compatibility combinations. Use immutable action commits verified during planning:

    actions/checkout@11bd71901bbe5b1630ceea73d27597364c9af683
    actions/setup-go@d35c59abb061a4a6fb18e82ac0862c26744d6ab5
    actions/setup-python@a26af69be951a213d495a4c3e4e4022e16d87065
    actions/upload-artifact@ea165f8d65b6e75b540449e92b4886f43607fa02
    actions/download-artifact@d3f86a106a0bac45b974a628896c90dbdf5c8093

Keep existing repository events and add workflow_dispatch for an explicitly requested run. Do not add automatic tag/publish jobs. The release job core is:

    release-baseline:
      strategy:
        fail-fast: false
        matrix:
          include:
            - {target: linux-amd64, runner: ubuntu-24.04}
            - {target: darwin-amd64, runner: macos-15-intel}
            - {target: darwin-arm64, runner: macos-15}
            - {target: windows-amd64, runner: windows-2022}
      runs-on: ${{ matrix.runner }}
      steps:
        - uses: actions/checkout@11bd71901bbe5b1630ceea73d27597364c9af683
          with:
            fetch-depth: 0
        - uses: actions/setup-go@d35c59abb061a4a6fb18e82ac0862c26744d6ab5
          with:
            go-version: '1.26.8'
        - uses: actions/setup-python@a26af69be951a213d495a4c3e4e4022e16d87065
          with:
            python-version: '3.11.9'
        - run: python -m pip install -r python/requirements-dev.txt
        - run: python tools/check_reproducible.py --ref HEAD --output build/reproducibility
        - uses: actions/upload-artifact@ea165f8d65b6e75b540449e92b4886f43607fa02
          with:
            name: baseline-${{ matrix.target }}
            path: build/reproducibility/accepted/
            if-no-files-found: error

Add compiler/SDK selection before the checker using the release-env values, and `go test -mod=readonly -race ./...` with explicit handwritten vet checks. Set task-local cache directories and GOTOOLCHAIN=local. The checker creates accepted/ only after its tests pass; copy only verified payloads and status evidence there. Upload failure logs separately with always(), clearly labeled diagnostics, never under the accepted artifact name. Ensure source SHA is the checked out commit; for PR workflows it may be a merge-test commit, which must be recorded rather than mislabeled as branch HEAD.

2. The installed-wheel job depends on release-baseline and combines each target/runner with all four Python versions. It downloads baseline-TARGET, installs the wheel under the selected interpreter in an external venv, removes PYTHONPATH and source-path injection, and runs both native wrapper tests and surface acceptance. Add `--wheel-only` to check_package so these interpreter combinations do not unnecessarily rebuild the sdist; the release job already rebuilt it on every platform. Missing wheel/native library, mismatched architecture/version, skip-only collection, or a skipped native test is a job failure. Set a pytest hook for this required suite that treats skipped required native tests as failure; optional unit tests are not counted toward this gate.

Use job result files recording target, full Python version, source SHA, wheel checksum, tests collected/passed/failed/skipped, and status. The tool copies tests/fixtures explicitly, checks the installed module path lies inside the venv, and executes `python -I -m pytest` from outside the checkout. It must not use pip editable installs. Runtime smoke tests run with no query/schema network calls; HTTP parity is loopback only.

3. Add a Linux Go-floor job using Go 1.22.12, `GOTOOLCHAIN=local`, and `go test -mod=readonly -race ./...` plus CLI/server compilation. No syntax or standard-library API newer than Go 1.22 may enter production sources. A separate Linux required native-memory job uses gcc-13, Go 1.26.8, and the Task 4 harness:

    CGO_ENABLED=1 CC=gcc-13 go build -mod=readonly -asan -buildmode=c-shared -o build/asan/libspl_toolkit.so ./pkg/bindings
    gcc-13 -fsanitize=address -g -Ibuild/asan tests/native/memory.c -Lbuild/asan -lspl_toolkit -Wl,-rpath,'$ORIGIN' -o build/asan/memory-check
    ASAN_OPTIONS=detect_leaks=1:halt_on_error=1 build/asan/memory-check

Create build/asan first. The harness includes the generated libspl_toolkit.h from that directory, not a stale checked-in header. If sanitizer runtime incompatibility prevents meaningful execution, record it and resolve the instrumentation; do not replace memory acceptance with an RSS plateau or mocked free counts. Run Docker acceptance and documented Make workflows in a separate Linux step so Windows shell limitations do not weaken those checks.

4. Implement check_acceptance.py to read explicit evidence JSON files from the completed jobs and require the exact set of four native targets, all 16 interpreter combinations, Go-floor, native-memory, CLI examples, surface parity, clean-source hash check, version agreement, Docker examples, and four reproducibility results. It validates source_sha against the requested ref and rejects missing, duplicate, failed, skipped, or mismatched-source records. An older passing run cannot close the current commit's gate. Add a test that starts with all required passing records, deletes darwin-arm64, and asserts a nonzero result; repeat with a mismatched SHA and a changed wheel checksum.

    def test_missing_arm64_is_not_complete():
        from tools.check_acceptance import validate_records
        records = [{"kind":"native", "target":target, "source_sha":"abc", "status":"passed"}
                   for target in ("linux-amd64","darwin-amd64","windows-amd64")]
        errors = validate_records(records, "abc")
        assert any("darwin-arm64" in error for error in errors)

Define `validate_records(records: list[dict], source_sha: str) -> list[str]` in that tool. It returns all missing/mismatched gate messages, and CLI exits 1 if the list is nonempty. Do not invent passing records in real evidence; synthetic records are test fixtures only. The final aggregation job requires every job via needs and runs even after failures solely to report an incomplete result; it cannot convert failed upstream jobs to success.

5. Run the tool tests and local workflow checks, then commit the CI/acceptance changes as `ci: require native baseline acceptance on four platforms` before obtaining remote results. Review that exact candidate with `git rev-parse HEAD`, `git status --short`, `git diff BASE..HEAD --stat`, and relevant reports, replacing BASE with the recorded implementation base SHA. Use independent code review under the normal execution workflow after task tests. Resolve findings with targeted regression tests and rerun affected checks against the updated committed candidate. Because this task may require remote CI access, implementation can finish locally while remote checks remain open. If pushing/running CI is not yet authorized, present the concrete local candidate and request the required action; do not state that milestone 1 is complete.

6. After every required gate actually passes, write `_build_plan/milestones/1-trustworthy-baseline/milestone-log.md`, starting with the required heading and concise user-facing bullets. Then include built files/surfaces, decisions beyond the PRD, handoff context, deviations, tested source SHA, artifact/checksum identities, exact platform/interpreter results, performance evidence, and any limitations. The intended opening is:

    ## What's new in the app

    - Map existing SPL queries from a configuration file using the documented CLI options.
    - Install and use the native Python package with complete discovery results and explicit cleanup.
    - Use the public library and bindings concurrently with consistent mapping configuration.
    - Build the documented release artifacts and inspect their compatibility and performance evidence.

Those bullets are written as completed capabilities only once proven. If work stops with remote gates open, use a clearly labeled interim log with pending items and avoid a completed-milestone claim. Include the four required detail sections: What was built; Decisions made during implementation; What the next milestone needs to know; Deviations from the PRD and why. State that milestone 2 still owns structured analysis and completeness; do not imply this milestone implemented them.

7. Commit the final evidence/log separately only after the tested code SHA from step 5 is established. A log cannot contain its own future commit hash: record the tested implementation SHA and later report the documentation-only final HEAD with ancestry and unchanged runtime-file proof. If runtime/configuration files changed in the evidence commit, the earlier run is insufficient and the affected gates must run again. Do not merge, push, or remove the worktree as an unrequested cleanup step.

## Validation and Acceptance


The following is the normal local acceptance sequence after Tasks 1–9, from the implementation checkout with the pinned tools and dedicated Python environment. Windows substitutes its venv interpreter path and binary suffix. These commands are implementation-time requirements, not results already obtained:

    python -m venv .venv
    .venv/bin/python -m pip install -r python/requirements-dev.txt
    make build build-server build-shared
    go test -mod=readonly -race ./...
    .venv/bin/python tools/check_go.py
    .venv/bin/python -m build --no-isolation --sdist --wheel --outdir dist python
    .venv/bin/python tools/check_package.py --sdist dist/spl_toolkit-0.1.1.tar.gz --wheel-dir dist
    .venv/bin/python tools/check_clean_build.py --ref HEAD
    .venv/bin/python tools/check_reproducible.py --ref HEAD --output build/reproducibility
    go test -mod=readonly ./pkg/mapper -run '^$' -bench '^BenchmarkBaseline' -benchmem -count=5
    git diff --check
    git status --short

Use the actual normalized artifact filename returned by the builder if setuptools normalizes spl-toolkit to spl_toolkit; choose one canonical spelling in the package tool and Make targets and test it rather than letting wildcard matches choose unrelated distributions. check_package requires exactly one matching wheel and one selected source archive.

Success means observed mapping text, complete existing discovery categories, clear errors/lifecycle, and working installed packages. It is not just compilation, an import, a mocked result, or unchanged go.mod while go.sum is missing. Final completion also requires Task 10's four-platform reports and the explicitly tested source identity.

## Idempotence and Recovery


All task tools allocate unique temporary directories and clean up only directories they created, unless the caller explicitly selected an evidence output location. Refuse to overwrite a nonempty evidence/output directory without an explicit tool option; the default is preservation. Use .venv, task-owned caches, and build directories in the isolated worktree. Never use git clean/reset on the original checkout or remove its untracked native products.

If a dependency fetch fails, preserve go.mod/go.sum and retry the pinned fetch; a network failure is not a product regression. If a toolchain is unavailable, record which pin is missing and acquire that exact version or explicitly revise the recorded pin with compatibility evidence. Do not automatically pick latest. If a package build fails, retain logs and the staged source tree; do not upload/install a partial wheel. If a native child crashes, retain its exact query and return code, then fix the ownership defect before resuming repeated calls.

If an implementation step discovers required work outside milestone 1, record it as deferred unless it blocks a documented baseline capability. Do not turn it into a grammar/analysis redesign. If the approved public contract genuinely must change, present the concrete conflict to the user before implementing that change. Routine fixes and implementation choices within this plan do not require renewed approval.

## Artifacts and Notes


Permanent artifacts are the implemented code, executable fixtures/tests, compatibility statement, truthful current docs, and performance description. Release artifacts and detailed logs live under build/dist and CI artifact storage; `_build_plan/` remains temporary. The spec and plan are guidance only.

Planning research verified runner labels and architecture distinctions against [GitHub's runner reference](https://docs.github.com/en/actions/reference/runners/github-hosted-runners), selected Xcode availability against the [Intel](https://github.com/actions/runner-images/blob/main/images/macos/macos-15-Readme.md) and [arm64](https://github.com/actions/runner-images/blob/main/images/macos/macos-15-arm64-Readme.md) image inventories, and GCC availability against the [Windows 2022 inventory](https://raw.githubusercontent.com/actions/runner-images/main/images/windows/Windows2022-Readme.md). Image inventories are mutable; record the actual executed image version and fail incompatible changes.

The Go release pin was checked against [official Go release metadata](https://go.dev/dl/?mode=json). Interpreter identities were checked against the official release pages for [3.11.9](https://www.python.org/downloads/release/python-3119/), [3.12.10](https://www.python.org/downloads/release/python-31210/), [3.13.7](https://www.python.org/downloads/release/python-3137/), and [3.14.0](https://www.python.org/downloads/release/python-3140/). These checks establish availability, not that this project passes on those versions.

Build/test package versions and requirements were read from public PyPI JSON metadata for the exact pins, including [setuptools 80.9.0](https://pypi.org/pypi/setuptools/80.9.0/json), [wheel 0.45.1](https://pypi.org/pypi/wheel/0.45.1/json), [build 1.3.0](https://pypi.org/pypi/build/1.3.0/json), and [pytest 8.4.2](https://pypi.org/pypi/pytest/8.4.2/json). The native wheel tag decision follows the [PyPA platform-tag specification](https://packaging.python.org/en/latest/specifications/platform-compatibility-tags/); it is not a manylinux portability claim. Reproducible epoch handling follows the [SOURCE_DATE_EPOCH specification](https://reproducible-builds.org/specs/source-date-epoch/), with final equality still enforced by tests.

Self-review coverage: scope and temporary guidance are captured in Global Constraints/Task 10; rule behavior in Task 2; ownership/concurrency in Task 3; C memory and Python lifecycle in Tasks 4–5; CLI in Task 6; versions/packages in Task 7; REST/parity/documentation/performance in Task 8; controlled artifacts/containers in Task 9; four-platform and exact-source acceptance in Task 10. No source implementation was modified while authoring this plan.

Revision note, 2026-09-06: Initial implementation plan based on the approved design. Resolved concrete tool/interpreter/runner choices, source-distribution contents, native error ownership, JSON-array context parity, rule-update precedence, CLI output shapes, and completion evidence. The user subsequently authorized continuous subagent implementation and independent review through all acceptance gates; Progress records the current execution state.

Execution clarification, 2026-09-06: Task 1 includes tools/tests/test_build_checks.py so the authorized focused tool regressions ship with the tools; the staging list was illustrative, not a prohibition on behavior tests.
