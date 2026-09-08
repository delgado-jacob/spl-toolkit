---
title: "Compatibility"
layout: page
---

# Compatibility

## Structured analysis contract and local verification

The additive [structured analysis API](API.md) uses report format integer `1` and compatibility `spl` / `splunkd` / `current`. Package version `0.1.1` and report format are separate identifiers. Empty compatibility values normalize to those defaults; other values are rejected. Go 1.22+ and Python 3.11+ remain the source/API floors.

Analysis is verified locally on macOS arm64 with Go race/regression tests and Python 3.12.6 source/native tests. Both a checkout-built wheel and a wheel rebuilt from the source distribution run mandatory copied analysis and legacy surface tests outside the checkout, with installed native libraries and zero required skips. All 34 reviewed analysis reports are compared as complete JSON values; only object-key order is irrelevant. Capability manifests also match across CLI, native Python, and REST. Go independently asserts the same corpus reports and semantic oracles. Additional local Python 3.14.1 checks are separate from the pinned package acceptance environment.

Go 1.22.12 targeted packages pass locally using `-ldflags=-linkmode=external`. On this host, the default-linked API test previously failed before assertions with a macOS `missing LC_UUID` loader error; external-link success does not establish default-link execution. The current Go bindings race build can emit an `LC_DYSYMTAB` linker warning despite passing tests.

These checks establish local analysis acceptance only. The historical release matrix below predates structured analysis and does not establish analysis acceptance on its other platforms or Python versions. No new cross-platform release or external Splunk-runtime conformance is claimed. Analysis deliberately reports incomplete coverage for the unsupported forms documented in the API reference.

## Field-list validation local verification

Field-list validation uses the same `spl` / `splunkd` / `current` compatibility defaults and integer report format `1`. Go, CLI, native Python, and REST compare all 26 shared validation reports exactly, including nested names, optional metadata, derived fields, removal, wildcard matches, conditional uncertainty, Unicode locations, and incomplete/error precedence. Batch reports preserve input order. Declaration optionality is valid equivalence; it does not assert event presence or override structural availability.

Local macOS arm64 verification uses Go 1.25.5 race checks, Go 1.22.12 affected-package tests with external linking, and Python 3.12.6. Each of the built wheel and rebuilt-sdist wheel runs 120 required native tests and 26 surface tests outside the checkout, with zero failures or skips. Surface acceptance includes the 34-case analysis corpus, the 26-case validation corpus, documented CLI commands, complete validation file/stdin/output equivalence, ordered batches, and separate request-error checks. The source Python suite has 137 passing tests; the tooling suite has 79. These are local verification counts, not a new release matrix. Existing Go linker and Python tar-extraction deprecation warnings remain qualified; no product defect is inferred from those warnings.

Unsupported wildcard command forms, dynamic constructs, macros, and conditional or unresolved branch behavior retain incomplete coverage. JSON Schema and OCSF validation have not shipped. The historical release evidence below does not establish acceptance for these new capabilities on other operating systems or Python versions. All five independent task reviews are complete. Additional local Python 3.14.1 source tests passed 137 tests with zero skips, and eight independent full Go/CLI/REST/native cases passed alongside concurrent repeats, ordered batches, input/output preservation, and strict-request rejection at source `78964a7775bb2e74d2022bb760d17be4e6e27e3a`. Those checks are separate from pinned package acceptance. Broad whole-milestone review and final controller milestone acceptance remain pending.

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
