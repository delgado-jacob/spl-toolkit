## What's new in the app

SPL Toolkit can now scan dedicated SPL/SPL2 query files or strict manifests,
retain failures alongside available findings, export dependency/lineage graphs
and SARIF, and compare schema or mapping changes against unchanged snapshots.
A detached document view exposes checked canonical evidence. The local stdio
language server delivers diagnostics and binding-aware document highlights.
CLI, stateless HTTP and owned native Python expose the same canonical tooling
operations, with published machine contracts and runnable embedding examples.

## Scope and source identity

M6 accepted product source: `98e38b941ea868468376b9dd6533e684009d35c0`.
M7 Task11 accepted native/package source: `6dc178a20b080a40e22455732e264e5e2ba264e0`.
Task12 began at administrative HEAD `985383eb12f6fc206f4a6b9fc363a645183cb430`.
Task12's source commit is identified in its independent review and immutable
consumer evidence; final M7/root acceptance remains pending at this log revision.
No merge, push, upload, version bump or new release target was performed.

The product remains 0.1.1. New corpus, graph, impact and document-view families
use schema_version 1. Owned schemas use Draft2020-12. SARIF uses 2.1.0 Errata01
with the official unmodified OASIS Draft4 schema. The local server targets LSP
3.17 full-text synchronization with UTF16 positions; it is not a full IDE.

## Changes and decisions

Tasks 1–11 introduced `pkg/corpus`, `pkg/corpusio`, `internal/corpusfs`,
`pkg/document`, `pkg/graph`, `pkg/sarif`, `pkg/impact`, `internal/lsp`, tooling
CLI/HTTP/native adapters and `contracts/`. Each task received an independent
review before its successor was released. Scoped fixes retained no-follow
acquisition, exact original bytes, detached prepared target data, observed
regressions amid partial coverage and explicit ambiguous alignment.

Task12 adds the test-only real editor client and fully locked official consumer
dependencies, `tools/check_editor_client.py`, `tools/check_sarif_consumer.py`,
consumer guard/release tests, maintained tooling/contracts guides, Go/Python/
CLI/impact/editor examples and explicit release manifests. `tools/release.py`
now packages 38 documentation/contract/example files and a normalized 185-file
Go build-source tar, including CLI/LSP, both frontends and anchored platform
opens. The Go source tar is a build closure, not the full repository test corpus;
Python's independently rebuildable sdist retains its native-source manifest.

Historical consumer preflight locks/wheels were absent. A fresh task directory
was provisioned from official npm/PyPI metadata. Root reviewed all integrity
and wheel hashes before offline installation. npm uses client9.0.1,
protocol3.17.5 and eight locked packages; sarif-tools3.0.5 has a 19-wheel closure
for CPython3.12.6/macOS26.2 arm64 only. Broader platform closure was deliberately
not inferred or manufactured from metadata alone.

## Local verification

- The real VS Code1.117.0 arm64 Extension Development Host and
  vscode-languageclient9.0.1 delivered `SPL_SYNTAX_ERROR`, error severity, the
  original message and UTF16 EOF range `[0,11,0,11]` for `eval '😀a'=`.
  Editing advanced the document version and cleared findings. The registered
  highlight provider returned read/write ranges, excluded an unrelated later
  assignment and an independent child scope, and retained non-BMP UTF16 ranges.
  Leaving the SPL language emitted real client document closure and cleared
  diagnostics; the client shut down gracefully. Temporary profiles/workspaces
  were used. A process inventory found no remaining task-owned editor hosts.
- The two-document SARIF consumer corpus has one invalid and one incomplete
  document. The real consumer emitted three rows: syntax error, unsupported
  command warning and unlocated tooling completeness warning. Summary, CSV and
  HTML conversion commands exited 0. Official schema validation and independent
  URI/root/index/Unicode code-point region assertions passed; source bytes did
  not change. Paths contain spaces, `#`, `%` and `é` and are encoded once.
- sarif-tools renders unlocated results as its own `-`/line1 sentinel and does
  not resolve uriBaseId or prove navigation. Those limitations remain explicit.
  Matplotlib's Agg backend creates HTML charts without a GUI.
- 36 consumer-wrapper/release tests passed after observing two missing-function
  regression failures. Two independently normalized source/content packages
  matched byte-for-byte. CLI/server/Go example built from the extracted Go
  archive using official Go1.22.12 and the offline module cache.
- Maintained documentation front matter and git diff whitespace checks passed.
  CLI graph/schema-impact/mapping-impact examples exited 0/1/0 as expected;
  the schema change introduces a definite missing-field error. The native Python
  example ran with an actual Task11 wheel installed into a fresh environment
  outside the checkout. These example checks supplement predecessor acceptance.
- Task11's reviewed package evidence records both direct-wheel and rebuilt-sdist
  installations passing 542 native plus 254 acceptance tests, zero skipped,
  on Python3.12.6. These are Task11 results, not newly rerun Task12 package checks.

Initial diagnostic wrapper failures remain retained: the editor launch needed
explicit `--wait`; closing only a VS Code tab can retain its cached text model;
SARIF HTML needed a noninteractive renderer; unlocated converter rows needed
their actual sentinel asserted. No product semantic behavior was changed for
those consumer harness corrections. The preflight editor version probe emitted
the historical SecCodeCheckValidity/NSOSStatusErrorDomain -2147409622 warning;
actual isolated host runs nevertheless completed. An initial failed evidence
file also contained a mistyped source SHA and is not acceptance evidence.

## Evidence and remaining gates

Task-owned evidence root:
`/private/tmp/spl-toolkit-m7-consumers-task12.QOw6lh`.
The ignored Task12 report records final candidate source, command outcomes,
manifest/lock hashes and consumer evidence paths. Precommit diagnostic consumer
runs are provisional; final source/executable hashes and immutable reruns are
required before acceptance.

Independent Task12 review, the broad M7/cross-milestone review and root's stable
final built/native/Python3.14/real-consumer acceptance remain open. Root owns that
window; no shared artifact is rebuilt during it. Linux/Windows filesystem
runtime and the complete pinned four-target release matrix have not run locally.
Compile-only checks do not prove runtime platform guarantees. Existing pinned
release/reproducibility/acceptance tools remain authoritative and are not relaxed
to convert local diagnostic environments into release acceptance.

## Task12 fixture correction

Independent re-review found that the exact named acceptance manifest and target
from the implementation plan were absent. Added `testdata/tooling/example-corpus.json`,
`example-target.json` and `example-corpus-missing-file.json`, plus a real CLI
regression and their explicit installed-fixture copy registration. The regression
first failed with exit2 for the missing target, then passed with the required
three-entry invalid/incomplete report (exit1) and four-selected/three-analyzed/
one-acquisition-failure report (exit2). It checks source bytes, unknown semantic
coverage and complete entry/identity preservation after reordering. No runtime
code or consumer dependency changed. Three tooling surface tests, 84 package
tests and focused Go1.22 corpus/corpusio tests passed. Existing consumer/package
receipts remain bound to their prior commits; root owns fresh final acceptance.
