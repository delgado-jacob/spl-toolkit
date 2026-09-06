# Milestone 1: Trustworthy Baseline

Date: 2026-09-06

Status: Written spec approved on 2026-09-06; the user requested the implementation plan. Implementation has not started. The ExecPlan follows `/Users/jacobdelgado/.codex/PLANS.md` and the writing-plans workflow and requires plan confirmation before execution.

## Purpose and scope

Turn the existing SPL alpha into a dependable baseline that a user can build from a clean checkout, exercise through Go, CLI, REST, and the actual Python native package, and understand from accurate documentation. The Go implementation remains canonical. The existing ANTLR grammar, flat discovery model, and token-preserving mapper remain the foundation.

This design implements only milestone 1 of the approved PRD. It does not introduce structured analysis, lineage, field-list or schema validation, standalone SPL2, a grammar rewrite, mapping 2.0, batch scanning, additional SDKs, or hosted services. Syntax validation does not promise semantic correctness or complete analysis. The later `valid` / `invalid` / `incomplete` analysis contract is not introduced here.

The PRD and milestone prompt under `_build_plan/` are temporary planning guidance. Runtime code, build configuration, tests, and release automation must not depend on that directory. Permanent documentation must stand on its own. At implementation completion, the milestone log goes in `_build_plan/milestones/1-trustworthy-baseline/milestone-log.md` and begins with `## What's new in the app`, followed by implementation details, decisions, handoff notes, and deviations.

## Repository context

The inspected baseline is commit `a276154689d946c9fd667765a7ef48759edfa021` on `main`. The working tree has pre-existing untracked `.omo/`, `CLAUDE.md`, `_build_plan/`, `bin/`, `bindings`, `go.sum`, and `spl-field-mapper` entries. Preserve this work. In particular, do not delete or silently adopt the existing untracked dependency checksum file as proof of a clean checkout. The implementation plan must establish dependency state from a clean, isolated source tree and compare any proposed checksum changes before integration.

Relevant existing components are:

- `pkg/mapper/mapper.go`, `parser.go`, and `schema.go`: public mapping, parsing, discovery, and mapping configuration.
- `pkg/bindings/bindings.go`: C-compatible handles, result structures, allocation, and cleanup.
- `python/spl_toolkit/`: ctypes wrapper, discovery values, exceptions, and package version.
- `cmd/main.go` and `cmd/server/main.go`: command-line and server entry points.
- `pkg/api/`: optional REST adapter and mapper caching.
- `Makefile`, `.github/workflows/ci.yml`, `python/setup.py`, `python/pyproject.toml`, and `VERSION`: builds, packaging, CI, and version metadata.
- `README.md`, `docs/`, examples, and Python tests: the user-facing claims and workflows that need to agree with executable behavior.

Inspection confirmed that normal builds run dependency tidy and formatting, CLI mapping constructs an unconfigured mapper, mapping priority is ignored, accepted regex conditions are not evaluated, the parser reuses mutable error state, native handles use an unsynchronized map, macro arrays are not populated, data-model pointer arithmetic is incorrect, and several native allocations have no effective cleanup path. Python packaging still declares Python 3.8+ and can continue despite a failed native build. These are observations from source inspection, not newly executed regression-test results.

## Chosen approach and alternatives

Repair the existing contracts and adapters with focused changes. Keep Go API signatures and the C result layouts, correct documented rule behavior, add the core documented CLI flags, and make package installation an acceptance test.

A documentation-only retreat would leave basic mapping configuration and native correctness unresolved. Replacing the analysis model or native protocol would expand scope and require consumer migration before the current functionality is trustworthy. A single global operation lock would make unrelated mappers block each other. Instead, use mapper-local state protection and short native-registry critical sections.

## Mapping configuration behavior

Base mappings apply first. Evaluate enabled conditional rules in ascending numeric priority. Equal priorities preserve configuration order. Only the first matching rule contributes mappings, and its mappings override base mappings for the same source field. Disabled rules never match. Unmapped fields and unaffected query text retain existing behavior.

Make regex conditions work wherever their supported condition type accepts string comparisons, including source and sourcetype values with the existing any-match behavior for string lists. Invalid patterns fail configuration validation. Validate operator/type combinations so accepted configurations do not silently select an unimplemented evaluator. Unsupported operators shown in existing documentation must be removed from current examples or explicitly identified as unsupported; implementing new condition families is outside this milestone.

Preserve the existing all-conditions-must-match behavior within a rule and existing explicit AND/OR child groups. Preserve the current query-context extraction limits and document them; do not replace them with lineage or a new query analysis engine. Configuration validation means validation of mapping settings, not validation against an event schema.

Keep `LoadMappings` as an update to the mapper's base mappings. An update is applied atomically, so an operation sees a complete state before or after the update. Base mappings copied from the initial configuration must not later overwrite an explicit update accidentally. Conditional overrides still apply according to the approved rule order.

Changing from all matching rules to first-match priority is intentional and must appear in compatibility notes, with regression fixtures showing overlapping rules, priority ties, disabled rules, overrides, no-match behavior, regex success, and invalid patterns.

## Core ownership and concurrency

Each mapper owns a copy of its input configuration, including nested rule data used by evaluation. Mutating the original configuration after construction must not change mapper behavior. Callers remain responsible for not concurrently mutating an input object while handing it to an API; this milestone does not promise safety for arbitrary races in caller-owned maps.

Protect mutable base mappings within each mapper and obtain a consistent effective-mapping snapshot for each operation. Parsing and rewriting use operation-local parser/listener state. The public parser must also be safe to reuse concurrently; its current shared error listener cannot remain shared across parses.

Independent mappers can operate concurrently. Synchronization should cover shared state, not all parsing work across the process. Preserve existing public method signatures and returned discovery fields. No new stage, reference, lineage, or completeness model is introduced.

## Native boundary and Python lifecycle

Keep existing C result structure layouts and exported operation signatures. Synchronize handle creation, lookup, and deletion with short registry locks. A successful lookup obtains a Go reference that remains alive for the operation even if the handle is subsequently removed. Later lookups of a closed handle return the existing missing-mapper failure. Do not reuse a removed handle in a way that lets a stale caller access a different mapper.

Populate all discovery arrays from the corresponding Go fields, including macros and multiple data-model references. Preserve each array's base pointer and pair each allocation with one release. Remove duplicate string allocation. Native error strings returned independently of a result structure need an explicit matching release function; the Python wrapper must retain the native pointer until it has copied and freed the message. Existing result cleanup remains in `finally` paths.

Python adds idempotent `close()` and context-manager methods. Operations after close raise a clear mapper-lifecycle exception. Coordinate close with active calls on the same Python instance so cleanup cannot invalidate their use. Finalization is a fallback and tolerates partially initialized instances. Existing usage without a context manager remains supported.

The installed package loads its bundled native library by default. An explicit caller-provided library path remains supported. Release verification must not accidentally use a library in the repository or current working directory.

Test native behavior through the actual shared library and installed wheel. Mock tests may remain useful for wrapper branches, but cannot count as native acceptance. Exercise both successful and failing calls, repeated cleanup, multiple discovery values, independent instances, concurrent operations, and concurrent lifecycle activity. Use Go race tests for Go shared-state races and native allocation/lifecycle checks for memory ownership; one does not substitute for the other.

## CLI contract

Keep `map`, `discover`, `validate`, `version`, `help`, and `demo`. Query commands accept one positional query or `--query`. Supplying both, extra query arguments, unknown flags, missing values, or unsupported formats is a usage error. This milestone adds the documented core flags, not batch testing, corpus scanning, stdin, or file-query workflows from later milestones.

`map` requires `--config` pointing to a mapping configuration document. The CLI must load and validate it before mapping. A valid loaded configuration may legitimately leave a query unchanged when no field matches. Requiring configuration is an intentional CLI compatibility change that prevents the current accidental identity-only command.

`discover` analyzes the supplied query without mapping configuration. `validate` checks query syntax when a query is supplied, or mapping configuration when `--config` is supplied alone. Supplying both a query and configuration to `validate` is a usage error, keeping the two meanings explicit. Configuration is not a supported option for `discover`.

All three commands support `--format text|json` and `--output`. Preserve defaults: mapped query text for `map`, the existing discovery JSON field names for `discover`, and `Valid` text for successful `validate`. JSON mapping output contains the mapped query under `query`; JSON validation output identifies the target as `query` or `configuration` and reports `valid: true` on success. Text discovery lists the existing discovery categories in a fixed order. These are CLI representations, not a replacement shared analysis report.

Results go to stdout unless `--output` selects a file. Errors go to stderr, with no success payload on failure; in JSON mode, emit a JSON error object there. Perform validation and computation before opening the output file so a rejected input does not truncate it. A successful output request replaces the selected file. File read/write failures must return an error rather than success.

Exit codes are 0 for success, 1 for rejected query or configuration content, and 2 for usage or file-access errors. Missing required flags are usage errors; malformed JSON, invalid regex, and parse rejection are content errors. Help and version succeed without requiring a query or configuration.

Every documented current CLI example must be runnable and checked for its expected result, not only its exit status. Remove the unsupported batch `test` command examples. Keep roadmap descriptions separate from current usage documentation.

## REST and surface parity

REST remains a thin optional adapter over the same core. Keep its existing endpoints and response structures except for focused correctness repairs needed by this milestone. Shared-mapper caching must remain safe under concurrent requests and inherit corrected rule behavior. No account, persistent query history, or external runtime dependency is introduced.

Use common supported-query and mapping fixtures to compare discovery values and mapped text across Go, CLI, Python, and REST. Differences in transport envelopes and established error representations are allowed; differences in the underlying mapping and discovery behavior are not. Test actual HTTP request/response handling as well as Go methods.

## Builds, dependencies, and version identity

Document clean-checkout commands for CLI, REST, shared library, Python package, tests, and benchmarks. Standard builds/tests consume committed dependency checksums in read-only mode and do not run dependency tidy, formatting, or generated-document updates. Dependency changes, formatting, parser regeneration, and API-document generation become explicit maintenance actions. Checks report problems without rewriting tracked source.

Track the complete required Go checksum state. Pin release toolchain and packaging dependencies, with exact versions recorded in the implementation plan and executable configuration. Avoid `latest` tool installation or unconditional dependency upgrades in repeatable build paths. Separate generated-parser static-analysis limitations from checks on handwritten code instead of using a formatter as a lint substitute.

Use the root `VERSION` file as the canonical release version. CLI, server reporting, native version reporting, Python metadata, and Python runtime reporting derive from it. Source distributions carry the same version as package build input and do not require the original checkout to read it. Release builds must not silently fall back to `dev` or contradictory hardcoded values. A documented direct developer build may identify itself as `dev`; it is not a release artifact.

Builds may fetch pinned dependencies during setup. Analysis, mapping, and validation remain local and require no network service. Offline runtime operation does not mean a machine with no compiler or dependency cache can bootstrap a source build without acquiring prerequisites.

## Python packaging and compatibility

Standardize package metadata and documentation on Python 3.11+. List exact interpreter versions actually exercised in CI separately from the installation floor. The implementation plan must pin that interpreter matrix before coding; the minimum Python 3.11 is mandatory on every required platform. Do not imply that future interpreter versions are tested merely because the metadata accepts them.

Wheels are platform-specific and include the matching native library. Packaging fails if the library cannot be built or is missing; no warning-only unusable wheel is accepted. Source distributions contain the Go sources, module/checksum inputs, version, and package build inputs needed to build the native library after extraction outside the checkout. Document Go and C compiler prerequisites for source installation. A wheel install must not need either compiler.

Required native execution targets are Linux x86-64, macOS x86-64, macOS arm64, and Windows x86-64. Each builds its artifacts, installs the wheel into a clean environment, and runs meaningful native tests. Building or inspecting another architecture's artifact does not replace executing it on that architecture. Keep distinct macOS architecture artifacts; universal binaries are not required.

The compatibility statement records exact tested OS/interpreter/toolchain combinations and artifact platform requirements. Other architectures and environments are not claimed as verified. Exact runner images, OS deployment floors, packaging tool pins, and interpreter versions are execution-plan choices to establish from toolchain and runner availability, without weakening the four-target requirement.

## Release reproducibility and performance

Reproducibility means rebuilding the same source revision with the same pinned toolchain and environment produces matching final artifact checksums. It does not mean binaries for different platforms are identical. Control source paths, timestamps, archive metadata, file ordering, and build metadata where needed. Pin or record the resolved build environment rather than assuming an evolving runner label is immutable.

Verify two independent clean builds of the relevant release artifacts in each target environment. Include CLI/server binaries, native-library payloads, and Python distribution archives. Publish or retain checksums and environment details as build evidence. Release packaging must fail on missing artifacts; wildcard copies that ignore errors cannot establish success. Preparing local artifacts and CI workflows does not itself publish a release or push a tag.

Record a performance baseline for existing supported parsing, discovery, and mapping workloads, including a concurrent workload. Store the source revision, query fixtures, hardware, operating system, toolchain, benchmark commands, repetitions, timings, and allocations. Report observed measurements without an invented pass/fail speed target or unsupported general throughput claims.

## Verification and completion

The implementation plan will turn this design into small steps with behavioral regression checks. Required evidence includes:

- A clean source checkout builds and runs documented tests without changing tracked dependency or source files.
- Go race tests cover shared mappers, mapping updates, parser reuse, and native handle management.
- Configuration fixtures prove first-match priority, stable ties, overrides, regex handling, and invalid-setting rejection.
- CLI acceptance executes every current documented example and checks output, stderr, exit codes, and file behavior.
- Real installed-wheel tests establish mapping/discovery parity, native cleanup, errors, and concurrent use.
- REST requests return equivalent core results and behave correctly under concurrent use.
- All four required platforms execute their installed native packages successfully, with no architecture skips counted as passes.
- Version values agree across packages and release surfaces, and repeated controlled builds have matching artifact checksums.
- Documentation accurately labels implemented functionality, syntax limits, compatibility changes, and deferred roadmap features.
- A reproducible performance report and milestone log record results and any remaining limitations.

Local success is not cross-platform acceptance. If a runner, platform execution, reproducibility check, or other required check has not run successfully, record it as open and do not declare milestone 1 complete. No implementation tests or platform gates have been executed as part of this design document.

## Approval record and next step

The user approved the core CLI scope, first-match priority and regex correction, all four release targets, core ownership and native/Python lifecycle design, detailed CLI contract, and build/release acceptance design in conversation. This document consolidates those decisions and makes edge behavior explicit for review.

After written-spec approval, create a self-contained ExecPlan in the milestone planning folder using the writing-plans skill and `/Users/jacobdelgado/.codex/PLANS.md`. Resolve and record exact toolchain/runner pins and executable acceptance commands in that plan. Obtain the plan confirmation required by the milestone prompt before implementation. This spec approval does not assert that any milestone functionality is already built.
