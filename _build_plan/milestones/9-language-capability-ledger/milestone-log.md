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

## Final local acceptance

Task 9 ran against commit `5662b7870f791077b33384ca70fb9aa34f5bd763` on macOS arm64 on 2026-09-18. No implementation correction was required.

Source and tooling gates:

- `env GOWORK=off GOCACHE=/private/tmp/spl-toolkit-m9-gocache go test -mod=readonly ./pkg/analysis ./pkg/rewrite -run 'Test.*(Capabilit|Ledger|Evidence)' -count=1` passed both packages.
- `python3 tools/check_go.py --race-timeout=20m` passed after an approved retry outside the sandbox. The first attempt stopped before testing because the sandbox denied access to the default Go cache. The passing run reported 16 tested packages and 9 packages with no test files.
- `python3 -m pytest -q tools/tests` passed 283 tests and 59 subtests. Pytest reported the existing `jsonschema` deprecation warning and a sandbox-blocked cache write; neither warning changed the result.
- `python3 tools/check_docs.py` passed all 17 documentation pages.

Built and cross-surface gates:

- `make build-all` built the CLI, server, and native shared library.
- With `PYTHONPATH`, `SPL_NATIVE_LIBRARY`, `SPL_CLI`, `SPL_SERVER`, `SPL_ANALYSIS_FIXTURES`, and `SPL_SPL2_FIXTURES` bound to absolute paths under this worktree, `python3 -m pytest -q python/tests/test_native_analysis.py python/tests/test_native_spl2.py tests/acceptance/test_analysis_surfaces.py tests/acceptance/test_machine_contracts.py` passed 170 tests. The passing run used an approved sandbox exception for the local HTTP server.
- The repository's hash-pinned package-check wheelhouse was provisioned under `/private/tmp`. With `PIP_FIND_LINKS`, `PIP_NO_INDEX=1`, `PIP_CONFIG_FILE=/dev/null`, and a writable Go cache, `make python-test` ended with `package acceptance passed`. Both the directly built wheel and the wheel rebuilt from the sdist passed 562 native tests, 257 cross-surface tests, and 18 tooling tests, for 1,674 passing pytest results across the two artifacts.
- `git diff --check` passed. `git status --short` was empty before this log update.

Acceptance criteria map:

- `pkg/analysis/capabilitydata/ledger.json` and `pkg/analysis/capabilitydata/corpus.json` contain the reviewed local records and evidence. `TestCapabilityLedgerCoversEveryKindPerLanguage`, `TestCapabilityEvidenceIsReferenced`, and `TestCapabilityEvidenceCorpus` passed.
- The built CLI reported 68 SPL records with 67 evidence cases and 98 SPL2 records with 108 evidence cases. `TestCapabilitySummaryUsesStrictDenominator` and `test_capability_ledger_contract_and_additive_v1_compatibility` passed; no percentage, score, or composite score is published.
- `TestCapabilityGrammarRegistrationDoesNotAddSyntaxCoverage` passed for the registered but unsupported SPL2 `spl1` quoted-pipeline form. Current SPL and SPL2 manifests report every linting record as unassessed.
- SPL and SPL2 semantic revisions are `sha256:1e6c75800f843931ec517dba27f3baa5af928a8a908d97dd1c62513e2ea24d31` and `sha256:0203cbeec2e0fd484080b1f512582a1c8bdc4bc541fb73ac42c013060695f84c`. `TestCapabilityRevisionUsesNormalizedTypedManifest` and `TestAnalysisRevisionMatchesCanonicalManifestPayload` passed.
- The CLI manifests reported `toolkit_version: 0.1.1`. `TestCapabilitiesPublishCanonicalLedger`, `test_source_capability_version_is_independent_of_package_and_native`, `test_analysis_capabilities_parity`, and the installed package gate passed, covering source `dev` isolation and tagged CLI, server, native, and wheel propagation.
- `TestCapabilitiesPreserveLegacyProjection` passed for the exact legacy `commands` and `functions` projections. `TestDialectCapabilitiesCLIHTTPParity`, `TestCapabilitiesRESTMatchesKernel`, `TestCapabilitiesBindingsReturnCompleteCanonicalOwnedManifests`, and `test_analysis_capabilities_parity` passed for existing CLI, HTTP, C, and Python surfaces.
- The full `main..HEAD` review found no parser or grammar change, no new endpoint, CLI command, C export, or Python method, no runtime external checkout, network, live Splunk, Git, or worktree read, no adapter classification, and no historical `docs/evidence` refresh. Production ledger and corpus data are compiled through `go:embed`.
- Generated OpenAPI changes are paired with `tools/update_validation_openapi.py` and its tests. `test_capability_ledger_shapes_are_strict_and_optional`, `test_capability_rewrite_evidence_declares_openapi_object_shape`, and the semantic JSON/YAML/docs.go reconciliation check passed.

The following gates remain separate and unverified:

- Hosted CI.
- Cross-platform wheel jobs.
- ASan.
- Merge readiness.
- Publication.
- Deployment.
- Live Splunk execution and runtime equivalence.

## Deviations from the PRD and why

Grammar registration is reported separately from evidence-backed syntax coverage. This prevents a registered parser rule from receiving coverage credit without reviewed positive evidence. The SPL2 `spl1` form remains registered for local parsing while its disputed embedded-body syntax stays unsupported in the ledger.

Verified during Task 8:

- `python tools/check_docs.py` passed and validated all 17 Jekyll documentation pages.
- `SPL_CLI="$PWD/build/spl-toolkit" SPL_DOCS_ROOT="$PWD" python -m pytest tests/acceptance/test_documented_cli.py -q` passed 3 tests. Pytest reported the existing `jsonschema` deprecation warning and could not write its cache under the worktree; neither warning changed the test result.
- Final diff review found changes only in the 11 documentation files listed above and this milestone log. `git diff --check` passed.

Task 9 final local acceptance passed. The unverified gates above are not implied by local source, package, or cross-surface results.
