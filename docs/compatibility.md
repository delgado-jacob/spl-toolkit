---
title: "Compatibility"
layout: page
---

# Compatibility

SPL Toolkit 0.1.1 was accepted at source commit `5191206124590f69a18c81fc51e756b6e75c90e8` by [GitHub Actions run 34075500813](https://github.com/delgado-jacob/spl-toolkit/actions/runs/34075500813). All 24 jobs passed, including exact-source aggregation of 28 evidence records.

| Release target | GitHub runner | Observed runner image | Compiler | Wheel tag |
|---|---|---|---|---|
| Linux x86-64 | `ubuntu-24.04` | `ubuntu24` `20260831.293.1` | GCC 13.3.0 | `linux_x86_64` |
| macOS x86-64 | `macos-15-intel` | `macos15` `20260824.0482.1` | Apple clang 17.0.0, Xcode 16.4 | `macosx_15_0_x86_64` |
| macOS arm64 | `macos-15` | `macos15` `20260829.0321.1` | Apple clang 17.0.0, Xcode 16.4 | `macosx_15_0_arm64` |
| Windows x86-64 | `windows-2022` | `win22` `20260830.290.1` | MinGW GCC 14.2.0 | `win_amd64` |

Each platform wheel passed under Python `3.11.9`, `3.12.10`, `3.13.7`, and `3.14.0`. Every one of the 16 combinations ran 11 required native tests and 5 surface acceptance tests with no failures or skips. The release builds used Python 3.11.9, Go 1.26.8, and pinned packaging tools (`build` 1.3.0, `packaging` 25.0, `setuptools` 80.9.0, and `wheel` 0.45.1). A separate Linux job passed the Go 1.22.12 minimum, and GCC 13 AddressSanitizer plus leak detection passed the native memory harness.

Wheels contain the native library and do not require Go at installation or runtime. They are specific to the listed operating system and CPU architecture; the Linux wheel uses the `linux_x86_64` platform tag and does not claim manylinux portability. The macOS wheels target macOS 15.0 or newer. Source builds require Go 1.22.12 or newer, Python 3.11 or newer, and a C compiler.

Checksums, exact environment identities, and installed-test counts are preserved in [the Milestone 1 acceptance record](evidence/milestone-1-acceptance.json).
