---
title: "Developer tooling"
layout: page
---

# Developer tooling

SPL Toolkit 0.1.1 scans dedicated query files, exports graph/SARIF evidence,
compares local schema or mapping changes, and serves editor diagnostics and
document highlights. Every operation uses the canonical offline Go engine.

## Repository and CI use

```sh
spl-toolkit scan --directory examples/tooling/queries --format json
spl-toolkit graph --directory examples/tooling/queries --output graph.json
python examples/tooling/graph_import.py graph.json
spl-toolkit scan --directory examples/tooling/queries --format sarif --output findings.sarif
spl-toolkit impact-schema --directory examples/tooling/queries --before-target examples/tooling/before-target.json --after-target examples/tooling/after-target.json --format json
spl-toolkit impact-mapping --directory examples/tooling/queries --before-rules examples/tooling/before-rules.json --after-rules examples/tooling/after-rules.json --format json
```

Exit 0 means valid, 1 invalid and 3 incomplete. Exit 2 takes precedence for a
request, acquisition, output or internal failure. Available findings are still
written before a content exit. In CI retain reports for exits 1, 2 and 3, then
apply your policy to both content status and execution completeness. Sources
remain unchanged, including during mapping impact previews.

Directory selection recursively includes exact lowercase `.spl`/`.spl2` files,
preserves UTF-8 bytes and lexical relative-path order, and skips `.git`, `.hg`,
`.svn`, `.worktrees`, `node_modules`, `vendor`, `.venv`, `venv`, `build`, `dist`
and symlink entries. Manifest selection preserves explicit order and supports
inline text or contained file paths. Paths and their ancestors must be physical
directories without symlinks/reparse points; on macOS use `/private/tmp` instead
of `/tmp`. Manifest query paths reject absolute paths, backslashes and `..`.
Unreadable selected files remain failures with their IDs; traversal gaps make
selection incomplete. See [manifest and request contracts](contracts.md).

Local file snapshots/reports are memory-resident, with no silent truncation.
The tooling HTTP routes use an 8 MiB body limit; LSP bodies are bounded at 64 MiB
and headers at 8 KiB. These transport bounds do not imply that arbitrary corpus
sizes are inexpensive. Split large workloads into explicit corpora.

Graph nodes and edges retain canonical evidence IDs, source ranges, scope,
lineage phase and uncertainty. Resolve every edge endpoint against the node
table; inspect the explicit relationship and evidence fields before grouping
dependencies. IDs are document/revision-qualified and can change after edits.
Impact classifications are static `affected`, `unchanged`, `indeterminate` or
`failed`; unchanged requires complete compared evidence. They do not establish
runtime equivalence or permission to apply migrations.

## Embedding and HTTP

The runnable [Go example](../examples/go/tooling/main.go) uses `corpus.DecodeRequest`,
`corpus.Scan` and `graph.Export`. Native Python adds `scan_corpus`, `export_graph`,
`export_sarif`, `impact_schema`, `impact_mapping` and `document_view` to the existing
`SPLMapper` lifecycle; see [the Python example](../examples/tooling/native.py).
All return dictionaries and use the packaged native library. The advanced
document snapshot is detached; mutating it does not reanalyze or attest new data.
Go lookup helpers use checked original byte ranges and return all defensible
intersecting references, without exposing parser internals.

Stateless POST routes are `/api/v1/corpus/scan`, `/api/v1/corpus/graph`,
`/api/v1/corpus/sarif`, `/api/v1/corpus/impact-schema`,
`/api/v1/corpus/impact-mapping`, and `/api/v1/query/document`.
They consume inline canonical requests and return direct reports. Server
requests cannot select local filesystem paths or fetch remote schemas.

## Local editor

Run `spl-toolkit lsp --stdio` from a compatible LSP 3.17 client. Register language
IDs `spl` and `spl2`, set the executable to the installed CLI and arguments to
`["lsp", "--stdio"]`. No product editor extension is distributed. A minimal
client configuration is in [examples/tooling/editor.json](../examples/tooling/editor.json).

The server supports full-text open/change/close synchronization, replacement
diagnostics, document highlights, configuration, cancellation and shutdown.
Positions are zero-based UTF-16. Open buffers are authoritative. A newer version
replaces the whole text and supersedes older diagnostic work; close clears
diagnostics. Highlight groups follow canonical origins/bindings, so equal
spelling in unrelated assignments/scopes does not merge symbols. No completion,
incremental edits, semantic tokens or query execution is advertised.

Initialization options accept `profile`, `version` and an optional inline
`validation_target`. `workspace/didChangeConfiguration` sends that object under
`settings.splToolkit`. Omitted targets select analysis only; invalid requested
targets produce visible configuration errors. CLI `--target FILE` reads the
same local target wrapper before server startup.

## Reproducible consumer acceptance

`tests/editor-client` is test-only code for a real isolated VS Code Extension
Development Host using vscode-languageclient 9.0.1/protocol 3.17.5. Its npm lock
records every transitive tarball integrity. The colocated Python consumer lock
is independent of the extension and pins sarif-tools 3.0.5 plus its complete
CPython 3.12 macOS arm64 wheel closure. `consumer-provenance.json` records official
PyPI metadata/download URLs and verified SHA256 values. Broader platforms need
their own reviewed wheel closures.

After reviewing/populating an isolated cache and wheelhouse, install npm with
`npm ci --offline --ignore-scripts` and Python with `pip install --no-index
--find-links WHEELHOUSE --require-hashes -r requirements-sarif-local-hashed.lock`.
Place installed npm modules in `CONSUMER_ROOT/node-consumer/node_modules`.
Use the existing editor executable; no installation into the user's profile is
needed. The wrapper supplies temporary user-data/extensions/workspace paths,
disables sync, and waits for actual client assertions before cleanup.

```sh
python tools/check_editor_client.py --cli /absolute/spl-toolkit --vscode /usr/local/bin/code --consumer-root /absolute/consumer-root --source-sha FULL_GIT_SHA --evidence /absolute/editor-evidence.json
python tools/check_sarif_consumer.py --cli /absolute/spl-toolkit --sarif /absolute/consumer-venv/bin/sarif --source-sha FULL_GIT_SHA --evidence /absolute/sarif-evidence.json
```

The SARIF wrapper needs the pinned independent JSON Schema validator environment.
It retains source snapshots and real summary/CSV/HTML conversions. CSV preserves
encoded paths and HTML uses `--no-autotrim`; sarif-tools does not resolve
`uriBaseId` or prove editor navigation/Unicode columns. Independent assertions
check the official OASIS schema, root URI resolution, single percent encoding,
source ranges, rule/artifact indices and invocation/coverage state. HTML charts
use Matplotlib's noninteractive Agg renderer.

Consumers supplement the Go/native/transport suites. A startup/version probe or
schema pass alone does not establish actual client behavior. Evidence records
the source SHA, executable hash, versions, command exits and delivered findings.
Local macOS evidence does not certify the other configured release platforms.

## Release content

The existing release builder packages CLI/server/native/header/Python wheel and
sdist plus the explicit `tools/release-content-files.txt` documents, contracts,
notices and examples. `spl-toolkit-source-0.1.1.tar.gz` is the separate complete
Go build-source archive selected by `tools/release-source-files.txt`; it includes
CLI/LSP and both generated frontends, but is not the full repository test corpus.
From its extracted root, provision the pinned go.mod/go.sum module closure and
run `go build ./cmd ./cmd/server` or `go build -buildmode=c-shared -o libspl_toolkit.so ./pkg/bindings`
(choose the native suffix for the host). Go 1.22 is the language floor. Python's
sdist remains independently rebuildable through its native-source manifest.
All release payloads, including normalized source archives, enter SHA256SUMS and
the existing two-build reproducibility comparison. Release versions and the four
configured targets remain controlled by VERSION and tools/release-env.json.
