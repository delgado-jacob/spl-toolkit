## What's new in the app

- See which fields a query reads, creates, renames, or removes.
- Follow fields through query steps and nested searches.
- Locate findings in the original query and see where analysis is incomplete.
- Keep using existing mapping and discovery workflows.

## Built files and contracts

- Public kernel `pkg/analysis`: QueryDocument/Result/CapabilityManifest, source index, typed ANTLR traversal, transfer rules, scope/dependency/recovery handling, and 34 independently reviewed full-report fixtures in `testdata/analysis/cases.json`.
- CLI `analyze` and `capabilities`, REST `POST /api/v1/query/analyze` / `GET /api/v1/capabilities`, C owned-result exports and guarded Python `analyze_query`/`capabilities` all preserve canonical values.
- Native source distribution includes the analysis kernel and strict JSON Unicode input validator. Installed suites now require `tests/acceptance/test_analysis_surfaces.py`, copied corpus identity, and zero required skips on both wheel paths.
- User-facing contracts live in README, `docs/API.md`, `docs/compatibility.md`, and `python/README.md`. Runtime and permanent user documentation do not depend on this build plan.

Additive analysis preserves the old flat API while giving new consumers flow/coverage data. Report `schema_version` is integer 1; fixture wrapper version is string "1"; compatibility version is `current`. Half-open locations use zero-based UTF-8 byte offsets, one-based code-point columns/lines, tabs as one column, and CRLF as one newline. Arrays/order/IDs/messages/source text are preserved; only JSON object-key order is irrelevant in parity checks. Invalid options/encoding are request errors; syntax/structural failures are report diagnostics. `invalid` takes precedence over incomplete coverage.

## Accepted limitations and next milestones

M3 must refine **open-input fields-inclusion retained-internal membership using a finite field catalog** before treating ordinary projections as conclusive. Do not invent `_raw`/`_time`, lose known internals/tombstones, or turn an unknown into false structural absence. External schema/field-list existence is separate from the current structural availability model. Removal references are non-consuming and create no schema obligation; exact versus wildcard roles are grammar/context dependent.

M6 must retain **component and overlapping data-model/dataset source spans**, normalized qualified identities, distinct reference IDs, and synthetic `macro` stages. Separate `datamodel Web All_Traffic` has a dataset component span normalized to `Web.All_Traffic`; qualified operands can expose overlapping root/model and dataset ranges. Future rewriting requires typed eligibility/rendering proof and conflict handling; do not directly replace normalized names at every located span. Macro expansion is unresolved and branch merging is incomplete.

Other accepted conservative limits: unknown commands/functions/dynamic semantics, unresolved wildcard membership, rename overlap/chains/swaps/duplicate claims/collisions, unmodeled command options, streamstats windows, dependency-only datamodel/from/tstats field effects, and branch merge effects remain incomplete. Exact local table/stats membership can narrow later state but cannot erase earlier report-level coverage gaps. No field-list/schema validation, broad SPL2, new rewriting, batch/editor tooling, runtime network calls, or external Splunk conformance was added here.

## Chronological rulings and refinements

The following exhaustively retains all 17 `Ruling:` entries from the controller ledger in original chronological order, including rationale and cost-if-wrong. The permissive rename ruling is historical and superseded; quotation handling is refined by typed role.

Ruling: Task 1's extraction-free Unicode test must assert the source-index helper directly, and Task 2 must assert a nonempty located target — avoids a vacuous range loop — cost if wrong: source-coordinate regressions would evade the initial gate.

Ruling: schema_version is integer 1, while the shared fixture wrapper version is string "1" — controller explicitly approved the machine-report integer — cost if wrong: fixtures/adapters would disagree on the public schema type.

Ruling: Ordinary commandless bare search terms (for example error) are valid SPL syntax, with literal context preserved; "bare malformed input" means damaged syntax, not implicit search — avoids rejecting supported SPL search forms — cost if wrong: overly broad implicit-search acceptance could need narrower grammar fixtures.

Ruling: Tasks 2 and 3 may make narrowly tested grammar/generated-parser extensions necessary for their assigned supported forms and typed dependency contexts — semantic handlers cannot recover grammar distinctions by text guessing — cost if wrong: wider parser changes require extra legacy regression review.

Ruling: Rerun Go 1.22.12's API package with -ldflags=-linkmode=external after macOS dyld rejected the default binary for missing LC_UUID — the failure precedes test assertions and is loader metadata, so retain both evidence paths without source changes — cost if wrong: the alternative link path alone would not prove default-link execution on this host.

Ruling: Enable analysis-specific dot disambiguation through an immutable per-document CharStream marker recognized by the ANTLR lexer predicate; unmarked legacy inputs retain original tokenization — shared text has different approved legacy/analysis interpretation and global lookahead cannot preserve both — cost if wrong: lexer mode coupling could leak unless direct and concurrent mixed-mode regressions prove isolation. No source rewriting, hand tokenizer, or global mutable state is allowed.

**Superseded by the next rename ruling.**

Ruling: Snapshot rename permits chains/swaps with distinct destinations because every source resolves from the before-state; duplicate sources or duplicate destination claims are incomplete — this makes the approved snapshot semantics coherent and preserves provenance — cost if wrong: unusual rename forms would need narrower declared semantic coverage.

Ruling: Superseding the earlier permissive rename interpretation, overlap chains/swaps and collisions remain incomplete; only ordinary distinct non-overlapping renames use before-state resolution — root confirms the approved uncertainty boundary and official reference does not establish overlap evaluation — cost if wrong: some deterministic rename forms receive conservative incomplete reports until conformance supports them. Preserve sound source reads, withhold unsupported resulting bindings. Reference: https://help.splunk.com/en/splunk-enterprise/search/spl-search-reference/9.1/search-commands/rename

Ruling: fields inclusion preserves known internal bindings/tombstones; closed-input projection remains exact, but open-input unknown retained-internal membership emits incomplete rather than inventing _raw/_time or false absence — root approved this conservative M2 limitation after official docs check — cost if wrong: broader indeterminate results reduce precision until M3 finite-catalog refinement. Table/stats remain exact. Preserve provable non-internal exclusions privately if safe.

Ruling: Unsupported rename overlap/collision processing retains sound reads and unaffected bindings but forgets affected source/destination claims without tombstones — avoids later false source/derived provenance while uncertainty applies — cost if wrong: affected later reads remain conservative until overlap semantics are proven.

Ruling: Qualified datamodel sources expose root data_model plus qualified dataset (Network_Traffic and Network_Traffic.All_Traffic); unqualified sources expose model only — existing mapper_test.go:860–918 verifies this dependency convention — cost if wrong: downstream dependency grouping would need migration, so retain exact source spans and grammar-established qualification. savedsearch:Daily remains a qualified dataset identity.

Ruling: A typed macro-only pipeline stage uses synthetic command `macro`, with actual macro identity in its located dependency and full invocation diagnostic — macro expansion can represent a pipeline and is not an ordinary named command — cost if wrong: consumers need documented synthetic stage handling; unresolved semantics always remain incomplete.

Ruling: Quoted literal field identifiers remain exact, while grammar-established search value patterns containing asterisk retain wildcard resolution even when double quoted — official SPL search example src="10.9.165.*" confirms quote delimiters do not literalize search-value patterns — cost if wrong: downstream dependency matching would confuse exact catalog names and patterns. Restrict decoding/pattern interpretation to typed AnalysisSearchValue roles. Reference: https://help.splunk.com/en/splunk-cloud-platform/search/search-reference/10.5.2605/search-commands/search

**Refined by root:** quoted expression identifiers may be exact, but quoted projection selectors containing asterisks remain wildcard selectors; quotation alone is not literal-field proof.

Ruling: Reject invalid UTF-8 text/source_id and invalid raw REST/C JSON encoding as document/request errors — successful report serialization otherwise substitutes U+FFFD and breaks byte-preserving source contract; root expressly required this — cost if wrong: callers with previously tolerated malformed encoding now receive errors rather than lossy reports.

Ruling: New analysis REST handler owns a narrow strict single-document/raw-UTF8 decoder retaining existing content-type/1MiB protections, while legacy decoder stays unchanged; CLI parsing becomes command-aware for explicit empty analysis values — this satisfies source/input contracts without unrelated legacy behavior changes — cost if wrong: the two bounded decoder paths require maintaining shared protection values.

Ruling: Add a narrow shared internal transport Unicode validator for raw UTF8 and surrogate escape pairing before REST/C decoding — root demonstrated encoding/json substitutes malformed UTF16 escapes even with valid raw UTF8, breaking source preservation — cost if wrong: validator could reject legitimate escaped literals, so paired non-BMP and escaped-backslash controls are mandatory. Standard decoder retains ordinary JSON syntax; no new missing-text/duplicate-key policy. Task4 owns helper; Task5 consumes and packages it.

Ruling: Allow post-Task4 correction worker to change CLI/REST corpus-loader guards from exact24 count to version1 plus nonempty, retaining full per-case equality — reviewed corpus growth should be consumed automatically while kernel corpus tests enforce exact fixture membership — cost if wrong: adapter guards alone no longer detect a nonempty truncated corpus, so final acceptance relies on the independent kernel membership/oracle gate. New correction adds two reviewed cases (26 total), prior24 expected values unchanged.

## Scope and source history

No product semantic source changed in Task 6. `docs/API.md` did not exist and was created with Jekyll front matter, while unrelated API-server docs remain intact. Package reporting now preserves the existing full-mode `--evidence` output for both installation results and fixture hashes. This is additive acceptance evidence. No corpus expectation changed.

Milestone base 6898cc052eb7109fbcc7495ee826c993aaa93352; design a0b5a03810d44f2464590286f1298dda8cb3080d. Product commits: 3589300 (source-aware entry), 1ea90bc/c57144b (lexical boundaries/isolation), d8c2da6/b430722 (flow/projection), 0c48604/e3c4684 (scopes/dependencies/recovery), df1bb0b (CLI/REST/Unicode), 829c4c1 (non-consuming removal references), 7cf9571 (C/Python/package closure). Task 6 starts from 8431e537a486bf5f1228b4df3796f88e6967583d. Detailed historical task reviews and correction rounds remain in controller artifacts; no open earlier task product finding was silently dismissed.

## Historical Task 6 local verification

The following records the pre-fix Task 6 run. Current final-fix evidence appears below.

All commands run from `/Users/jacobdelgado/repos/spl-toolkit/.worktrees/remaining-milestones`. Go commands/builds/tool pytest/package checker use the following environment (supplied inline in actual invocations):

```sh
GOCACHE=/private/tmp/spl-toolkit-go-cache
GOMODCACHE=/private/tmp/spl-toolkit-remaining-gomodcache
GOPROXY=off
GOTOOLCHAIN=local
```

Python is `/private/tmp/spl-toolkit-remaining-venv/bin/python` (3.12.6). The already-authorized server-inclusive checks ran with `require_escalated` for local loopback access. Commands and results:

```sh
/private/tmp/spl-toolkit-remaining-venv/bin/python tools/check_go.py
/private/tmp/spl-toolkit-floor-tools/gomodcache/golang.org/toolchain@v0.0.1-go1.22.12.darwin-arm64/bin/go test -mod=readonly -ldflags=-linkmode=external ./pkg/analysis ./pkg/mapper ./pkg/bindings ./cmd ./pkg/api
make build-all
/private/tmp/spl-toolkit-remaining-venv/bin/python -m pytest tools/tests -q
PYTHONPATH=python SPL_EXPECTED_VERSION=$(cat VERSION) SPL_NATIVE_LIBRARY=/Users/jacobdelgado/repos/spl-toolkit/.worktrees/remaining-milestones/build/libspl_toolkit.dylib /private/tmp/spl-toolkit-remaining-venv/bin/python -m pytest python/tests -q
/private/tmp/spl-toolkit-remaining-venv/bin/python tools/check_docs.py
make python-build PYTHON=/private/tmp/spl-toolkit-remaining-venv/bin/python PIP=/private/tmp/spl-toolkit-remaining-venv/bin/pip
/private/tmp/spl-toolkit-remaining-venv/bin/python tools/check_package.py --sdist dist/spl_toolkit-$(cat VERSION).tar.gz --wheel-dir dist --evidence /private/tmp/task6-package-evidence.json
```

- `check_go.py`: formatting and handwritten vet pass; it runs the required `go test -mod=readonly -race ./...`, with six tested packages passing (cached unchanged executable/test inputs), six packages with no test files, no test failures. Existing `LC_DYSYMTAB` bindings linker warning remains. Log `/private/tmp/task6-check-go.log`.
- Go 1.22.12: all five named packages pass with external linking; actual package durations/cached marker in `/private/tmp/task6-go-floor.log`. The earlier default-link API `missing LC_UUID` failure remains a loader limitation; this does not claim default-link execution on this host.
- `make build-all`: pass, `/private/tmp/task6-build-all.log`.
- Tool suite: **74 passed, zero skips**, `/private/tmp/task6-tools.log`.
- Source native Python suite: **61 passed, zero skips**, actual built dylib and VERSION override, `/private/tmp/task6-python-source.log`.
- Docs checker: **13 pages pass front-matter validation**, `/private/tmp/task6-docs.log`.
- Python wheel/sdist build: pass, `/private/tmp/task6-build-python.log`.
- Final package checker: **44/44 required native + 15/15 surface acceptance for each of the checkout-built and rebuilt-sdist wheels, zero failures/skips**; total 118 passing invocations across the two paths. Ten of the 15 surface tests are new analysis checks: full 26-case report parity, capability parity, six malformed-Unicode HTTP rejections, and two valid Unicode controls. Log `/private/tmp/task6-installed-final.log`; full JSON evidence `/private/tmp/task6-package-evidence.json`. Existing tar extraction deprecation warning remains.

Both installed paths use Python `-I`, isolated environments outside the checkout, copied mandatory tests/corpora, and `SPLMapper()` without a library path. `clean_env()` removes PYTHONPATH, PYTHONHOME, SPL_NATIVE_LIBRARY, and SPL_EXPECTED_VERSION. Required-suite plugin fails on zero collection, any skip/failure, or incomplete pass counts. Metadata verifies installed module under venv, controller site-packages absent, package/native version 0.1.1. The checker also verifies source closure, missing-compiler behavior, metadata without compiler, and native artifact identity. Temporary wheels/venvs are deleted after success; their exact observed paths and hashes remain in the JSON evidence.

## Historical Task 6 artifact hashes and reused evidence

These artifact identities and the 78-input root reconciliation predate the final selector corrections. They are historical and do not establish acceptance of the corrected executable inputs.

- `dist/spl_toolkit-0.1.1.tar.gz`: SHA256 `bb295a72ea2e8453a36491e237901ee0d2e6794fed06468aebd084ffa1334b57`.
- `dist/spl_toolkit-0.1.1-py3-none-macosx_15_0_arm64.whl`: SHA256 `08d3ed790fbec371b6c1207e415afb1eec0fae322806b20f89a90c2e1e8a4d46`.
- `testdata/analysis/cases.json`: SHA256 `a47e5f602ec726f5b15214df47ddff063b70b2d0d40e87eecaf486fc34c21370`.
- `testdata/baseline/cases.json`: SHA256 `322383ae58a7fead0e013c089f6ae947061207a568bad2ec47b6e8ce3afc8dfb`.
- Rebuilt-sdist wheel (temporary, removed after verification): SHA256 `e34f5951e3a9f91cee094161178692f22e2e422163eb14b035d4c48dc0a5d3ea`.

Root extra evidence `/private/tmp/spl-toolkit-root-m2-acceptance-py314.json` was produced at `7cf957100e22b75a5dee6a1836ef1c90b7067d93`. `git merge-base --is-ancestor 7cf9571 HEAD` passes. The only committed intervening change at Task 6 base 8431e537a486bf5f1228b4df3796f88e6967583d is the living plan. Task 6 changes no product executable inputs. `/private/tmp/task6-input-reconciliation.json` records **78 identical executable/build input hashes**, including all tracked Go/grammar sources, module files, VERSION, Makefile, native manifest/build configuration, and Python package sources. At Task 6, all three `make build-all` artifact hashes matched the root evidence:

- CLI `64f0f84813fb7844dd0d7f7677da47bd582f3df0f7331bc45a4178c632e61757`.
- Server `82715ebd783f105eb386dabd2efb20a50a0fcc7dc1ec9ae052d51f20492e8893`.
- Dylib `bcadea4668bf3b2812a8a4b7ab2ed1411b3b15b0b77fb42a5a70a82e8f1d4cf1`.

At Task 6, the root's **Python 3.14.1 source 61/zero-skip result and four extra Go/CLI/REST/native cases with eight concurrent native repeats per case plus three malformed-Unicode REST rejections** were reused under that input reconciliation. The final selector corrections invalidate that reuse for current acceptance. Root will rerun its expanded seven-case and Python 3.14 acceptance at the corrected committed inputs. Both local built/rebuilt-sdist installed paths were actually rerun by Task 6; the refreshed final-fix gates are recorded below.

Task 3 ANTLR reproducibility is reconciled against all 10 unchanged current parser output hashes in `.superpowers/sdd/implementation-plan/task-3-antlr-reproducibility.json`. The current ANTLR 4.13.2 JAR matches `eae2dfa119a64327444672aff63e9ec35a20180dc5b8090b7a6ab85125df4d76`. No grammar/generated source changed and no needless regeneration was performed. Task 3's original second-generation evidence remains the reproducibility proof; Task 6 validates its input/output continuity.

## Open acceptance

These are local macOS arm64 gates, not a new cross-platform release matrix. The historical release evidence predates analysis. Go 1.22 default-link API loading and existing linker/tar warnings remain documented environment limits. Task 6 independent review passed at `35da878fa35b203a83382829e84f824db5927d19`. Broad milestone review at `21e4e784bf243a99e1fb7541cde7276a3629c696` found two Important selector defects and one Minor CLI help omission. Their one final fix wave is recorded below. One independent scoped re-review, refreshed root seven-case/Python 3.14 verification, and root final acceptance remain pending. No push, merge, publish, or unrelated user guidance staging occurred.


## Final milestone fix wave

Fix base `21e4e784bf243a99e1fb7541cde7276a3629c696`. All three broad-review findings are addressed in this one scoped wave: wildcard pattern stars cannot enter the literal-equality branch; selector normalization now joins only the typed identifier and wildcard terminals; existing CLI help describes analyze/capabilities, document options, and content exits. Quoted expression identifiers remain exact. Selector source ranges and original names are unchanged. Removal references remain non-consuming `remove` / `not_applicable`; only a later unavailable read receives the structural diagnostic.

Focused TDD retained `/private/tmp/m2-final-fix-red-corrected.log`: 16 failing include/exclude subcases plus one failing neighboring-selector test, with eight passing controls. GREEN `/private/tmp/m2-final-fix-green.log`: all 24 include/exclude subcases and the neighboring-selector test pass. Controls cover literal-star prefix/suffix, backtracking, suffix mismatch, empty wildcard match, quoted prefix/suffix/both, escaped quotes, Unicode, fully quoted patterns, exact later reads, source spans, and provenance. Initial backtracking/mismatch test inputs were corrected to fully quoted patterns because unquoted interior stars are separate typed selectors in the existing grammar; no grammar coverage was expanded.

Eight complete expected reports were constructed independently for closed literal-star and quoted-fragment prefix/suffix inclusion/exclusion. All prior 26 reports retain identical values and byte prefix. Current corpus: **34 cases**, SHA256 `b0b1c1e1e8276e3caf4512ab6ad34f6ef1dab9a0d8779b5673530a4932f5bb00`. The existing CLI help snapshot alone was refreshed; other legacy CLI examples retain identical values. The first installed attempt passed 44 native and 14 surface tests but failed the stale help snapshot. After updating that one expected stdout, both installation paths passed the complete checker.

Refreshed gates use the same recorded environment/commands above, with unique `m2-final-fix-*` evidence paths. Focused analysis and legacy mapper packages pass. Go checker passes formatting, handwritten vet, and full race tests: six tested packages pass, six have no test files (`/private/tmp/m2-final-fix-check-go-final.log`). Go 1.22.12 external-link floor passes all five named packages (`/private/tmp/m2-final-fix-go-floor.log`). Tools: **74 passed**; source native Python 3.12.6: **61 passed**, zero skips. `make build-all` and wheel/sdist build pass. Actual `build/spl-toolkit analyze --help` exits 0 and its reviewed output is `/private/tmp/m2-final-fix-help.txt`.

Both checkout-built wheel and rebuilt-sdist wheel pass **44 required native + 15 surface tests each, zero failures/skips** (118 passing invocations total). These include full 34-case report parity, capability parity, malformed-Unicode rejection, and CLI documentation acceptance. Current evidence: `/private/tmp/m2-final-fix-package-evidence.json`; log `/private/tmp/m2-final-fix-installed-final.log`. They use isolated installed modules outside the checkout, copied corpora, removed source/native overrides, and required-suite skip/collection enforcement. Existing LC_DYSYMTAB and tar-deprecation warnings remain qualified; Go 1.22 default-link LC_UUID loader behavior was not retested or claimed resolved.

Current artifact SHA256 values:

- `build/spl-toolkit`: `248f6487c4ce1d22512f23db610f3637375e54c72bc58faaade090b35d93363d`.
- `build/spl-toolkit-server`: `7d95437303c73b13846fc77b6e0e1330276059b9857a5935fdd1c65880c07db2`.
- `build/libspl_toolkit.dylib`: `d0ce0a8a2046719c50653aaf93ff8e775f120808f1db654b7eb39abea48400aa`.
- `dist/spl_toolkit-0.1.1.tar.gz`: `007b75aa5a06e749356a72d3526a4bd455d2e4103e1575bd665f1d87a2036dda`.
- `dist/spl_toolkit-0.1.1-py3-none-macosx_15_0_arm64.whl`: `a860c0196d5a9ed9fb10370727bae5d4b96aa53749b54dbefdacb88e4d806926`.
- Rebuilt-sdist wheel (temporary artifact removed by checker): `62f4c6c42bb4d220e04c8ffff3dc02123717f0de9447b976c9e5c55c51be6b8a`.

`/private/tmp/m2-final-fix-input-artifacts.json` retains the 82-path current inventory (historical 78 paths plus the new selector test, both corpora, and CLI example fixture), exact artifact hashes, and the final source commit. All 10 ANTLR output hashes, both grammar inputs, and all three generated OpenAPI outputs remain unchanged. No regeneration was needed. The old 78-input/binary root reconciliation is historical; it cannot establish corrected-head acceptance.

Detailed dispositions, exact commands, RED/GREEN output, self-review and scoped commit: `.superpowers/sdd/implementation-plan/milestone-fix-report.md`. All 17 chronological rulings and accepted M3/M6 handoffs remain unchanged. Independent scoped re-review, root seven-case/Python 3.14 verification, and root final acceptance remain open. After the scoped commit, source/build mutations stop for the controller's immutable verification window. No push, merge, or publishing is included.
