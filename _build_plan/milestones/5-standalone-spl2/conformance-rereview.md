# Milestone 5 conformance correction — scoped re-review

Reviewed 2026-09-07. Scope: the single P2 finding in `conformance-review.md`, its correction and immediately related consistency/count checks. The original broad review was not repeated. This report is the only file written; no production, parser, test-suite or Git operation was performed.

**P2 addressed. No finding remains open from the independent conformance review.** The matrix is suitable as a design seed for later TDD under its documented gates. This verdict does not establish executed conformance or authorize implementation before the verified M4 handoff.

- `conformance-matrix.md:836` now uses `/(?<tag>\w+)/` for L23-P1. L23-N1 at line 838 removes only the closing slash, providing an uncontested positive/minimal-malformed pair consistent with the previously inspected primary evidence.
- H12 at line 1087 retains balanced slash-plus-pipe syntax as disputed/incomplete, expressly outside positive/negative floors. It does not manufacture a definite-invalid outcome. The rex expansion row at line 1033 preserves quoted alternation as admitted syntax. Its escaped pipe is Markdown table formatting, not an added regex escape in the rendered example.
- The two nonblocking expansion reminders remain explicit at line 1039: expression-position search literals and duplicate-object-key rejection. `research-notes.md:92–94` accurately describes the correction and its limits.
- Document-only recount: 53 inventory rows (50 splunkd, three exclusions), 68 cards, 288 candidates, 152 N/B candidates, 56 deliberately SQL-focused candidates. Counts are unchanged. No duplicate IDs or fully assembled source texts were found. H12 contributes no counted candidate.

The verified M4/shared-lowering and SQL-phase handoff, per-option/arity fixture expansion, exact source/reference/scope/flow assertions, and executed parser/adapter/regression checks remain open. No new local inconsistency was found in the corrected material. Limitations of the original targeted-source design audit continue to apply.

## Verified artifact identity

| Artifact | SHA-256 |
| --- | --- |
| conformance-matrix.md | `3e402c278bdca70a152eaf0b0ea8b97ed2c3dc9b4bbde04958957a541199fc63` |
| research-notes.md | `01070caa5ccb92b2568d0d61ef96f10eace28e5f3cd9f199b69df6b04f141616` |

Both hashes matched the controller's requested re-review snapshots. The original review remains preserved as historical evidence.
