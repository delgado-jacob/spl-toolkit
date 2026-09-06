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

Generated parser sources are committed build inputs. Runtime code and build configuration do not depend on `_build_plan/`. Expanding the grammar, adding SPL2 modules, introducing structured analysis, or changing the public ABI belongs to later work.
