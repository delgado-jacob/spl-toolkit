## What's new in the app

- Capability manifests now explain five evidence-backed dimensions and five explicit states.
- SPL reports 86 records with 85 evidence cases. SPL2 reports 114 records with 124 evidence cases.
- All 34 advertised rewrite forms now have canonical ledger records and executable success or boundary evidence.
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

### Final local acceptance

Task 9 ran against commit `5662b7870f791077b33384ca70fb9aa34f5bd763` on macOS arm64 on 2026-09-18. Final review later found that the rewrite compatibility projection was checked against its fixture but not against the canonical ledger. The correction added exact records and independently replayed evidence for all 31 supported forms and three unsupported boundaries.

Correction validation passed `env GOWORK=off go test ./... -count=1`, the 297-test native/CLI/HTTP acceptance selection, `python3 -m pytest -q tools/tests` with 283 tests and 59 subtests, all 17 documentation pages, and `git diff --check`. The offline `make python-test` package gate passed for both the direct wheel and rebuilt-sdist wheel, each with 562 native tests, 257 cross-surface tests, and 18 tooling tests. The semantic revisions changed to the values recorded below.

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

The 12 approved design completion criteria map to passing evidence in design order:

1. The curated ledger represents every required construct kind for SPL and SPL2. `pkg/analysis/capabilitydata/ledger.json` and `TestCapabilityLedgerCoversEveryKindPerLanguage` passed. The built CLI reported 85 SPL records and 112 SPL2 records.
2. Every record has five explicit support states, one for each of syntax, semantics, requirements, linting, and safe rewriting. `TestEmbeddedCapabilityAssetsAreStructurallyValid`, the strict `decodeCapabilityAssets` path, and `test_capability_ledger_shapes_are_strict_and_optional` passed.
3. Every `supported`, `partial`, and `unsupported` claim cites passing local evidence for its exact dimension and form. `TestValidateCapabilityClaims` and the external `TestCapabilityEvidenceCorpus` execution passed against `pkg/analysis/capabilitydata/ledger.json` and `pkg/analysis/capabilitydata/corpus.json`.
4. Every `not_applicable` claim has a reason and no evidence citation. The `not applicable` cases in `TestValidateCapabilityClaims` and the strict capability-claim schema checks in `test_capability_claim_state_boundaries_and_semantics_nonempty` passed.
5. Every `unassessed` claim remains visibly unassessed and contributes no covered credit. `TestValidateCapabilityClaims`, `TestCapabilitySummaryUsesStrictDenominator`, and `test_capability_ledger_contract_and_additive_v1_compatibility` passed. Current SPL and SPL2 manifests report every linting record as unassessed.
6. Positive, negative, and incomplete corpus cases have stable IDs. `pkg/analysis/capabilitydata/corpus.json`, `TestCapabilityDataCanonicalOrder`, `TestDecodeCapabilityAssetsRejectsMalformedInput`, `TestDecodeCapabilityAssetsRejectsMalformedObservations`, and `TestCapabilityEvidenceCorpus` passed.
7. JSON and text reports show exact, reproducible per-dimension counts for the tagged toolkit version. The built CLI reported `toolkit_version: 0.1.1`, 86 SPL records with 85 evidence cases, and 114 SPL2 records with 124 evidence cases. `TestCapabilitiesCLIFormatsAndOutput`, `TestCapabilitiesCLITextIsCanonicalAndComplete`, `TestCapabilitySummaryUsesStrictDenominator`, `TestCapabilitiesPublishCanonicalLedger`, and `test_source_capability_version_is_independent_of_package_and_native` passed.
8. Users can follow each evidence-bearing claim to the corresponding compact corpus case. `TestCapabilitiesPublishCanonicalLedger` verified that each selected claim ID resolves within the returned evidence, `TestCapabilityEvidenceIsReferenced` verified the corpus-to-claim relationship, and `TestCapabilityEvidenceCorpus` replayed those compact cases.
9. Go, CLI, REST, native C, and Python return equivalent canonical values. `TestDialectCapabilitiesCLIHTTPParity`, `TestCapabilitiesRESTMatchesKernel`, `TestCapabilitiesBindingsReturnCompleteCanonicalOwnedManifests`, and `test_analysis_capabilities_parity` passed.
10. Existing capability consumers retain their command/function compatibility fields. `TestCapabilitiesPreserveLegacyProjection` passed for exact legacy values. The `capabilities-archived-v1` artifact in `testdata/tooling/contracts.json` and `test_capability_ledger_contract_and_additive_v1_compatibility` passed for archived v1 compatibility.
11. The machine contract, OpenAPI, documentation, source package, and installed package checks pass. `test_capability_ledger_contract_and_additive_v1_compatibility`, `test_capability_ledger_shapes_are_strict_and_optional`, the semantic JSON/YAML/docs.go reconciliation check, `python3 tools/check_docs.py`, `test_native_source_manifest_contains_capability_ledger_closure`, `test_capability_ledger_files_are_release_source_inputs_only`, `test_copied_capability_tests_do_not_require_checkout_version`, and `make python-test` passed. `python3 tools/check_docs.py` validated the 17 Markdown pages under `docs/`. The `main..HEAD` diff review covered the other maintained documents listed above and this milestone log, and `git diff --check` passed. The package gate checked both the built wheel and the wheel rebuilt from the sdist.
12. No external repository, live service, query execution, or later-milestone feature is required. Production ledger and corpus data are compiled through `go:embed`; the installed-package checks run outside the checkout. The full `main..HEAD` review found no runtime external checkout, network, live Splunk, Git, or worktree read, no parser or grammar change, no new endpoint, CLI command, C export, or Python method, and no adapter classification or historical `docs/evidence` refresh.

Separate contract checks verified that `capability_revision` includes the typed semantic contract and excludes only `toolkit_version`. SPL and SPL2 revisions are `sha256:117785ed5ff93fd72fa0030c56325ecf9daf50034eaed22c2e2f80bd22eae59a` and `sha256:f1391296cfbc616e9bb1b1828e2471e37b60a35c0555654c0734640e072a0437`; `TestCapabilityRevisionUsesNormalizedTypedManifest` and `TestAnalysisRevisionMatchesCanonicalManifestPayload` passed. `TestCapabilityGrammarRegistrationDoesNotAddSyntaxCoverage` separately verified that parser registration adds no evidence-backed coverage.

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
