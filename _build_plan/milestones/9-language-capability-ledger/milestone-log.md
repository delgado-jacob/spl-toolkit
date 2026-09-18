## What's new in the app

- Capability manifests now explain five evidence-backed dimensions and five explicit states.
- SPL reports 68 records with 67 evidence cases. SPL2 reports 98 records with 108 evidence cases.
- Summary counts use strict denominators and publish no percentage or composite score.
- Parser registration, syntax coverage, lint evidence, and safe-rewrite evidence remain separate claims.

## What was built

Documentation now describes the ledger on every public capability surface:

- `README.md`
- `docs/API.md`
- `docs/api-server.md`
- `docs/architecture.md`
- `docs/cli.md`
- `docs/compatibility.md`
- `docs/contracts.md`
- `docs/spl2.md`
- `docs/rewrite.md`
- `python/README.md`
- `contracts/README.md`

The five dimensions are syntax, semantics, requirements, linting, and safe rewriting. Their states are supported, partial, unsupported, not applicable, and unassessed. For every dimension, applicable equals supported plus partial plus unsupported plus unassessed; total records equal applicable plus not applicable; covered equals supported. Partial and unsupported claims receive zero covered credit.

Evidence IDs connect claims to typed local documents, observations, classifications, and provenance. Stable IDs keep their exact reviewed scope. A broader form receives a new ID unless a reviewed scope correction establishes that the original scope was wrong.

The semantic `capability_revision` includes schema version, selectors, documentation snapshot, legacy command/function projections, rewrite forms, records, summary, and evidence. It excludes only `toolkit_version`. Source Go tests can report `dev`, while tagged CLI, server, native, and packaged artifacts report exact `VERSION` with the same semantic revision.

The corpus establishes static local toolkit behavior. It does not establish live Splunk execution, runtime equivalence, environment compatibility, authorization, deployment, publication, or upstream support.

## Decisions made during implementation

The user authorized one deviation from a model that would infer syntax coverage from parser registration: grammar registration is modeled separately. `grammar_registered` records local parser ownership and never adds evidence-backed coverage.

The SPL2 `spl1` quoted-pipeline record demonstrates the boundary. `spl2.command.spl1.quoted-pipeline` has `grammar_registered: true`, while its syntax claim is `unsupported` and receives zero covered syntax credit. Its incomplete evidence preserves the H10 embedded-body boundary.

Linting uses its own evidence boundary. Parser, semantic, and validation diagnostics do not become lint evidence. Both current language snapshots therefore report all linting records as unassessed.

Legacy `commands` and `functions` remain compatibility projections. Consumers that need evidence-backed decisions use `records`, `summary`, and `evidence`.

## What the next milestone needs to know

Task 9 must run and record the final source, package, and cross-surface gates. It must verify exact counts, semantic revisions, tagged version propagation, generated contracts, and evidence parity from the final commit. Hosted CI, publication, deployment, live runtime execution, and upstream acceptance remain unverified.

Task 9 may update this log with final verified commands and results. It must preserve the distinction between local source checks, packaged-artifact checks, hosted CI, publication, deployment, live Splunk execution, and upstream support.

## Deviations from the PRD and why

Grammar registration is reported separately from evidence-backed syntax coverage. This prevents a registered parser rule from receiving coverage credit without reviewed positive evidence. The SPL2 `spl1` form remains registered for local parsing while its disputed embedded-body syntax stays unsupported in the ledger.

Verified during Task 8:

- `python tools/check_docs.py` passed and validated all 17 Jekyll documentation pages.
- `SPL_CLI="$PWD/build/spl-toolkit" SPL_DOCS_ROOT="$PWD" python -m pytest tests/acceptance/test_documented_cli.py -q` passed 3 tests. Pytest reported the existing `jsonschema` deprecation warning and could not write its cache under the worktree; neither warning changed the test result.
- Final diff review found changes only in the 11 documentation files listed above and this milestone log. `git diff --check` passed.

Task 9 final gates are pending. This log does not claim hosted CI, artifact publication, deployment, live Splunk runtime equivalence, environment compatibility, authorization, or upstream support.
