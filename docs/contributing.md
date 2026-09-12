---
title: "Contributing"
layout: page
---

# Contributing

Use the pinned Go 1.26.8 release toolchain and Python 3.11 or newer. The declared Go source floor is 1.22.12 and is exercised in Linux CI; it is not a claim that every older-Go/current-macOS linker combination works. Install pinned dependencies, then run the checks appropriate to your change:

```bash
make deps
make test
make lint
make python-test
make bench
```

CLI syntax and expected output are executable cases; see the [CLI guide](cli.md). Keep new current-user examples tied to that manifest.

The Go implementation is canonical. Preserve exported mapper signatures, all seven discovery categories, the C result layouts, and existing exit codes. Mapping and discovery changes should have behavior-level tests at the Go boundary and across any affected CLI, Python, or REST surface.

The current bounded contracts include legacy mapping/discovery, structured SPL/SPL2 analysis and lineage, offline field/schema validation, and explicit safe rewriting. Preserve their documented limitations and independent corpus expectations. Further grammar expansion, SPL2 modules and new SDKs require approved design rather than opportunistic changes. See the [rewrite contract](rewrite.md) before changing source-binding or commit-gate behavior.

Do not edit generated parser output to suppress tool warnings. Regenerate it through the pinned workflow when an approved grammar change requires regeneration. Build and runtime paths must not depend on temporary planning files.
