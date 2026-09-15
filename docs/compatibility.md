---
title: "Compatibility"
layout: page
---

# Compatibility

## Developer tooling contracts

SPL Toolkit 0.1.1 adds corpus, graph, impact and document-view report families at
schema version 1. The local language server deliberately targets LSP 3.17 full
synchronization, diagnostics and document highlights. SARIF targets 2.1.0 Errata
01 with Unicode code-point columns; editor positions use UTF-16. Public source
offsets remain UTF-8 bytes. See [contracts](contracts.md) for additive-output and
strict-request compatibility rules, and [tooling](tooling.md) for limits and
real-consumer recipes. The four release targets and Python 3.11+ support policy
remain unchanged; local consumer checks do not certify the entire matrix.

## Structured analysis contract and local verification

The additive [structured analysis API](API.md) uses report format integer `1` and compatibility `spl` or explicitly selected `spl2`, profile `splunkd`, and version `current`. Package version `0.1.1` and report format are separate identifiers. Empty compatibility values normalize to those defaults; unsupported selector values are rejected. Go 1.22+ and Python 3.11+ remain the source/API floors.

Analysis is verified locally on macOS arm64 with Go race/regression tests and Python 3.12.6 source/native tests. Both a checkout-built wheel and a wheel rebuilt from the source distribution run mandatory copied analysis and legacy surface tests outside the checkout, with installed native libraries and zero required skips. All 34 reviewed analysis reports are compared as complete JSON values; only object-key order is irrelevant. Capability manifests also match across CLI, native Python, and REST. Go independently asserts the same corpus reports and semantic oracles. Additional local Python 3.14.1 checks are separate from the pinned package acceptance environment.

Go 1.22.12 targeted packages pass locally using `-ldflags=-linkmode=external`. On this host, the default-linked API test previously failed before assertions with a macOS `missing LC_UUID` loader error; external-link success does not establish default-link execution. The current Go bindings race build can emit an `LC_DYSYMTAB` linker warning despite passing tests.

These checks establish local analysis acceptance only. The historical release matrix below predates structured analysis and does not establish analysis acceptance on its other platforms or Python versions. No new cross-platform release or external Splunk-runtime conformance is claimed. Analysis deliberately reports incomplete coverage for the unsupported forms documented in the API reference.

## Query requirements contract

The additive [query requirements API](API.md#query-requirements) uses report format integer `1` and the existing SPL or explicit standalone SPL2 compatibility selectors. Current runtime analysis and document-snapshot reports always emit `requirements`. The version-1 analysis and snapshot schemas allow this property without adding it to their `required` arrays, so archived version-1 payloads that predate requirements still validate. The standalone requirement schema requires its complete shape. Consumers must continue to tolerate additive output properties under the version-1 policy.

For one normalized document, standalone requirements and the requirements embedded in plain analysis, field-list or schema refinement, rewrite analysis, and document snapshots contain the same query-only value. CLI, REST, C, and Python serialize that canonical Go value without reclassification; only JSON object-key order is immaterial. Query status and requirement coverage are independent, so consumers must inspect both. An invalid query can have complete requirement coverage, while a valid query with dynamic or indeterminate obligations can have incomplete requirement coverage.

The C export `spl_mapper_requirements_query` follows the existing owned-result ABI. A non-null `SPLResult*`, including an error result, must be released exactly once with `spl_result_free`. Python uses the same mapper admission and close synchronization and frees the native result after decoding on success or failure.

Canonical parsing admits 4,096 lexer work units for SPL and SPL2 before parser construction or prediction. Real lexer errors precede the token returned by the same call, and that non-EOF token follows them in the count. SPL2 unmatched-literal closure accounting follows the last admitted token and precedes EOF. Unit 4,097 is omitted, and its exact half-open location is reported. The [API reference](API.md#canonical-lexer-work-boundary) records the canonical diagnostic and gap messages, token, lexer-error, and SPL2 opener range rules, the bounded incomplete report, and the resource-limited rewrite no-op contract.

Resource exhaustion is a successful content outcome. Go and native/Python return owned values without an API error, CLI emits the report and exits 3, and REST returns HTTP 200. The REST transport boundary remains separate: a request body larger than 1 MiB returns HTTP 400 before analysis. Long sparse input remains admitted when it stays within the lexer work budget.

Acceptance checks bind two ASCII dense fixtures, 64 KiB and 256 KiB, to a serialized `RequirementSet` no larger than 4,096 bytes and a serialized analysis `Result` no larger than the query bytes plus 4,096 bytes. These are fixture-specific checks, not universal byte guarantees. JSON escaping can expand arbitrary text, and an admitted document can amplify inherited lineage evidence below the lexer budget.

Requirement extraction is deterministic and offline. It reads no path, URL, environment variable, event data, credential, or global registry. It does not expand knowledge objects, resolve placeholders, assess an environment, generate query variants, or execute a query. Query and capability digests identify supplied data; they are not authentication, authorization, signatures, compatibility approval, or execution permission.

## Field-list validation local verification

Field-list validation uses the same `spl` / `splunkd` / `current` compatibility defaults and integer report format `1`. Go, CLI, native Python, and REST compare all 26 shared validation reports exactly, including nested names, optional metadata, derived fields, removal, wildcard matches, conditional uncertainty, Unicode locations, and incomplete/error precedence. Batch reports preserve input order. Declaration optionality is valid equivalence; it does not assert event presence or override structural availability.

Local macOS arm64 verification uses Go 1.25.5 race checks, Go 1.22.12 affected-package tests with external linking, and Python 3.12.6. Each of the built wheel and rebuilt-sdist wheel runs 120 required native tests and 26 surface tests outside the checkout, with zero failures or skips. Surface acceptance includes the 34-case analysis corpus, the 26-case validation corpus, documented CLI commands, complete validation file/stdin/output equivalence, ordered batches, and separate request-error checks. The source Python suite has 137 passing tests; the tooling suite has 79. These are local verification counts, not a new release matrix. Existing Go linker and Python tar-extraction deprecation warnings remain qualified; no product defect is inferred from those warnings.

Unsupported wildcard command forms, dynamic constructs, macros, and conditional or unresolved branch behavior retain incomplete coverage. This paragraph records the earlier field-list checks; the schema field-validation evidence is recorded separately below. The historical release evidence below does not establish acceptance for these new capabilities on other operating systems or Python versions. All five independent task reviews are complete. Additional local Python 3.14.1 source tests passed 137 tests with zero skips, and eight independent full Go/CLI/REST/native cases passed alongside concurrent repeats, ordered batches, input/output preservation, and strict-request rejection at source `78964a7775bb2e74d2022bb760d17be4e6e27e3a`. Those checks are separate from pinned package acceptance. Broad whole-milestone review and final controller milestone acceptance remain pending.

## JSON Schema and OCSF local verification

The additive schema field-validation contract uses report format integer `1`, compatibility `spl` / `splunkd` / `current`, Go 1.22+ and Python 3.11+ floors. It checks declarations rather than JSON event instances or expression types. Local source verification ran Python 3.12.6 (230 source tests), the 28 frozen schema reports, 22 raw malformed requests, ordered batches, real OCSF file activity/generic descendants and executable local-resource examples. Source schema/CLI acceptance passed 57 tests with zero skips. Go 1.22.12 passed the six required packages using external linking.

Both the Go1.22-built wheel and the wheel rebuilt from sdist outside the checkout passed 213 required native tests and 81 surface tests each, with zero failures or skips and no checkout imports. Final tooling passed 114 tests. Installed wheel and rebuilt-sdist checks, exact input manifests and artifact identities are retained in [Milestone 4 evidence](evidence/milestone-4/verification.json). The new schema suite is mandatory in copied installed acceptance, alongside existing mapping, analysis and field-list suites. Full reports are compared, not only statuses. External proxies are unusable during runtime checks; loopback remains reachable for real HTTP. Code inspection separately confirms preparation has no HTTP/file opener. All catalog bytes are pinned and unchanged.

This evidence is local macOS arm64 acceptance work. It does not establish an unrun Linux/Windows release matrix, all supported Python versions, a new release, external Splunk conformance, or complete JSON Schema instance validation. The known Go 1.22 default-link LC_UUID failure was not repeated; external linking is the verified local path. The current Go race linker may emit the existing LC_DYSYMTAB warning. The package checker compares complete toolkit C layouts/exports across Go-generated header boilerplate, while retaining exact wrapper, source-header, actual wheel-header and loaded-library hashes. Tooling, compiler preparation and runtime version requirements remain distinct. Task 7 independent review, broad milestone review, and controller acceptance are separate open gates at this documentation checkpoint.

## Standalone SPL2 verification boundary

Explicit `spl2` selection is available through canonical Go, CLI, HTTP, C/native Python analysis and field/schema validation. Omitted language continues to select SPL. The [standalone SPL2 guide](spl2.md) describes the build capability snapshot, dedicated syntax families, complete and conditional effects, and held or unmodeled boundaries. Neither `current` nor the documentation snapshot is live Splunk certification.

The durable corpus keeps original private syntax expectations separate from canonical semantics. Independently determined full canonical expectations are distinct from deterministic recovery/representation snapshots. A snapshot earns conformance credit only with independently authored semantic assertions frozen before capture and independently reviewed location, ownership and soundness evidence. Exact final-field sets are asserted where proved; unknown recovery representation does not authorize invented fields, source bindings or valid/complete promotion. Original source links, IDs and provenance holds remain intact.

The closed corpus contains 1,714 original documents and 1,835 active obligations: 732 independently exact canonical objects and 982 independently constrained, reviewed regression snapshots. All canonical objects are executed by Go; the final auditor requires zero pending obligations. Deduplicated floors are 288 meaningful inputs, 139 definite negatives and 45 SQL/mixed groups, with 17 holds preserved. The V1 documentation/source projection remains `3345cf5712b1bdbf467d1651784fdb8bccc596805038da0d54e7a123384e3a4e`.

Local source verification passed the Go1.22.12 external-link floor, the full Go formatting/vet/race runner, 191 tooling tests, 339 Python tests and 78 documented CLI/schema/SPL2 surface tests. Complete reports for every corpus document match Go, CLI, real HTTP and source native Python. [Milestone 5 verification](evidence/milestone-5/verification.json) records source and installed wheel/rebuilt-sdist outcomes, copied fixtures, required-suite counts and exact artifact identities. The installed checker requires the new SPL2 suite and rejects missing, empty or skipped required coverage; its full-Go transport artifact grants no semantic conformance credit.

The initial floor API attempt was blocked by sandbox loopback permissions; the same API scope passed with authorized local listener access. The known race-linker warning remains nonblocking. Python3.14 acceptance, independent task review, broad milestone review and final root acceptance are separate gates. Local macOS arm64 evidence does not establish unrun Linux/Windows/Python release combinations or external Splunk execution conformance.

## Safe rewrite verification boundary

The [safe rewrite contract](rewrite.md) adds explicit preview/apply and ordered batches for SPL and standalone SPL2 without changing legacy mapping precedence. The durable suite contains 62 independently specified public reports, a complete 34-entry dialect/kind/render-role capability matrix, and two private candidate-syntax/binding fault cases. The matrix and fault cases are separate checks, not extra full-report conformance credit. Required groups cannot become empty; unsupported syntax and changed SPL2 implicit labels remain refusal cases.

Go checks authored candidate/returned text, commit/status, exact audit, source bindings and lineage before producing transport reports. CLI, real local HTTP, source-native Python, a directly built wheel and a wheel rebuilt from sdist compare every full canonical JSON value; only object key order is immaterial. Both installed paths require the rewrite native/surface suites alongside predecessor schema/SPL2/documentation checks, exact copied fixture hashes, isolated module/library origins, nonzero collection and zero required skips.

Local automated verification uses official Go 1.22.12 with external linking and Python 3.12.6 on macOS arm64, with offline modules and pinned packaging wheels. It does not establish a new release, other OS/CPU/Python combinations, ASAN/leak-sanitizer acceptance, or Splunk runtime equivalence. Independent Task8 review, broad milestone review and root-owned stable-built/installed/Python3.14 acceptance are separate gates. The retained milestone log identifies actual command results and artifact hashes; the historical matrix below does not certify these new APIs.

## Historical 0.1.1 release acceptance

SPL Toolkit 0.1.1 was accepted at source commit `6b55f8902ff1a990aea8951cd78934b32c90660a` by [GitHub Actions run 34077738907](https://github.com/delgado-jacob/spl-toolkit/actions/runs/34077738907). All 24 jobs passed, including exact-source aggregation of 28 evidence records.

| Release target | GitHub runner | Observed runner image | Compiler | Wheel tag |
|---|---|---|---|---|
| Linux x86-64 | `ubuntu-24.04` | `ubuntu24` `20260831.293.1` | GCC 13.3.0 | `linux_x86_64` |
| macOS x86-64 | `macos-15-intel` | `macos15` `20260824.0482.1` | Apple clang 17.0.0, Xcode 16.4 | `macosx_15_0_x86_64` |
| macOS arm64 | `macos-15` | `macos15` `20260829.0321.1` | Apple clang 17.0.0, Xcode 16.4 | `macosx_15_0_arm64` |
| Windows x86-64 | `windows-2022` | `win22` `20260830.290.1` | MinGW GCC 14.2.0 | `win_amd64` |

Each platform wheel passed under Python `3.11.9`, `3.12.10`, `3.13.7`, and `3.14.0`. Every one of the 16 combinations ran 11 required native tests and 5 surface acceptance tests with no failures or skips. The surface suite used 10 shared cases and compared mapped text exactly across Go, the CLI, installed Python, and HTTP, including mapped and no-op queries with boundary spaces, tabs, and newlines. The release builds used Python 3.11.9, Go 1.26.8, and pinned packaging tools (`build` 1.3.0, `packaging` 25.0, `setuptools` 80.9.0, and `wheel` 0.45.1). A separate Linux job passed the Go 1.22.12 minimum, and GCC 13 AddressSanitizer plus leak detection passed the native memory harness.

Wheels contain the native library and do not require Go at installation or runtime. They are specific to the listed operating system and CPU architecture; the Linux wheel uses the `linux_x86_64` platform tag and does not claim manylinux portability. The macOS wheels target macOS 15.0 or newer. Source builds require Go 1.22.12 or newer, Python 3.11 or newer, and a C compiler.

Checksums, exact environment identities, and installed-test counts are preserved in [the Milestone 1 acceptance record](evidence/milestone-1-acceptance.json).
