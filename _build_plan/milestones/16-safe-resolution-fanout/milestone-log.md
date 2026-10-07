## What's new in the app

- Resolve explicit Dataset, index, lookup, data model, source, and sourcetype names through Go, `spl-toolkit resolve`, HTTP `POST /api/v1/query/resolve`, C `spl_mapper_resolve`, and Python `SPLMapper.resolve`.
- Receive the ordered Cartesian combinations of supplied values, bounded to 100 variants by default or a positive caller override. Over-limit requests are rejected before any variants are generated.
- Read verified queries alongside each failed or incomplete sibling. Only `verified` variants publish `resolved_query`, after independent analysis, substitution proof, and compatibility assessment.
- Inspect exact changes and source locations, original input roles, environment/schema evidence, and per-variant provenance.

## What was built

`pkg/resolution/` owns version-1 request/report types, strict decoding, overflow-safe counting, deterministic sequential fanout, prepared orchestration, and text formatting. Public Go entry points are `Resolve`, `ResolveJSON`, `DecodeRequest`, `Prepare`, and `Prepared.Resolve`. Request/value order is preserved, with the last placeholder varying fastest. Exact product counts are decimal strings; limits and ordinals are uint64. A query without external markers still receives one independently assessed empty combination.

`pkg/analysis/resolution*.go` owns document-only typed discovery, simultaneous canonical identity rendering, opaque session-bound proof, exact UTF-8 byte audits, and original-to-candidate occurrence/requirement correspondence. Whole markers match `$[A-Za-z_][A-Za-z_0-9]*` in supported typed identity slots. Local symbols, aliases, comments, ordinary strings, embedded markers, and unsupported owners gain no substitution authority. Narrow private extensions reuse canonical rewrite machinery while preserving ordinary rewrite behavior. Serialized or mutated reporting evidence cannot manufacture proof authority.

`pkg/compatibility/resolution*.go` owns original-role binding admission, independent role assessment, and supplied-definition closure mapping. Small helper extractions in `fields.go` and `prepare.go` retain existing ordinary assessment rules. Equal candidate Dataset names do not merge original roles or permit schema borrowing. Required and conditional occurrences retain their original obligations; unproved ownership remains incomplete. Submitted dependency bindings and effective translated bindings are separately reported. `pkg/environment`, `pkg/validation`, and `pkg/closure` retain artifact, field, and definition authority.

Adapters are `cmd/resolution_cli.go`, `pkg/api/resolution.go`, the existing `pkg/bindings/bindings.go`, `python/spl_toolkit/libspl_toolkit.h`, and `python/spl_toolkit/mapper.py`. CLI writes the complete report before exit: 0 for all verified, 1 for any failed, 3 for incomplete without failure, and 2 for request/tooling errors. Existing output-file protection rejects aliases of the request. HTTP accepts at most 8 MiB; content reports return 200, configuration errors 400, and internal failures 500. C returns an owned result freed with `spl_result_free`; Python returns detached dictionaries and raises `SPLMapperError` for configuration failures.

`contracts/v1/resolution-request.schema.json`, `resolution.schema.json`, and shared definitions publish the additive contracts. Generated OpenAPI, `docs/resolution.md`, maintained surface guides, and `examples/resolution/request.json` document them. The example yields two variants: one verified, one failed. The existing SPL2 Dataset-parameter safe-rewriting ledger entry gained bounded proof evidence; other capability dimensions remain separate. Native/release inventories and installed/hosted acceptance registries include the new sources, example, and suites. `_build_plan/` is administrative guidance, never a runtime or packaging input.

### Validation and provenance

Implementation worktree: `/Users/jmdelgad/.worktrees/spl-toolkit/milestone-16-safe-resolution-fanout`, branch `codex/milestone-16-safe-resolution-fanout`, base `3842c614dd45a83186332b581174702d420cecea`. Product/artifact acceptance is bound to `e87bbc690de7ad34101ca95f90e48a59ed824a44`; final implementation head before this log is `f705d4578a7e7740bcd612b4002e4e03313e5b98`.

- `GOWORK=off python3 tools/check_go.py --race-timeout=20m` passed formatting, vet, and the complete Go race suite at `da1c8726e6660b707d18156de4ae27a13d85dd1a`. No Go source changed through `f705d45`. The receipt is `build/m16-final/check-go-record.json`.
- At `e87bbc6`, `make build-all` and source resolution acceptance passed: 54 tests, zero skips, covering 16 authored cases and five malformed cases across Go/CLI/HTTP/C/Python. Controls include single/multiple values, 2-by-3 ordering, default-bound rejection and override, role isolation, unsuccessful siblings, exact audits, and publication gates.
- At `e87bbc6`, actual direct-wheel and separately rebuilt-sdist installations each passed 661 native tests, 361 surface tests, and 30 machine-contract tests, with zero failures/skips. The receipt's surface count of 391 includes the 30 machine-contract tests. `python3 tools/check_clean_build.py --ref HEAD` passed a committed clean export with fresh caches, product/native builds, and all Go race tests.
- Contract/documented-CLI checks passed 52 tests plus 75 subtests during Task 12; generated documentation was byte-idempotent. Documentation checks passed in Task 13. At `f705d45`, the complete tooling suite passed 418 tests plus 179 subtests, zero skips. A behavioral test executed two actual Make dry runs and verified unique immutable receipt paths.
- Tasks 1–12 passed independent specification and quality reviews. Final specification and quality reviews passed at `f705d45`; reviewers independently checked focused behavior, source/fixture/manifest relationships, and actual artifact hashes.

Local ignored evidence remains in `build/m16-task01/` through `build/m16-task13/` and `build/m16-final/`. `build/m16-task13/package-summary.json` is the compact extraction of the actual `build/package-check.json`, whose SHA-256 is `9cb12876abd2ab5fa86bfb6f0540e6353515a00790a1d7904589b116cc381ddf`. Direct wheel SHA-256: `d724c10715cbbdb237ab46ad095e8235686959ea857c2206feb7ef8a015b83a6`; rebuilt-sdist wheel: `ee3612a7836facbd07d1d1d83a874bfa710e35e2c063a78a9a4ad9ff410578b9`; sdist: `a9d6caf3c72205be78bf083390097322e307315a8cd38300e297931dd1b673ab`. Both installed native payloads hash to `60f9d569cc1d48b7afd2bd8c8c44ed03f37bc4c4e2b15570ccd913e885f4e58a`. Resolution fixture SHA-256: `17f6f39e4396413cd747a6e453bc4255fbd1227788c4b386506ac924386ad6fa`.

The available local functional gates satisfy the PRD's single/multi-value and independently checked bounded-set criteria. Hosted native/reproducibility/platform records were unavailable, so aggregate `tools/check_acceptance.py` remains **UNRUN**. This is local evidence on macOS arm64, not hosted acceptance, release readiness, live search acceptance, deployment, or UAT.

## Decisions made during implementation

Candidate names and compatibility selections are separate inputs. The final assessment consumes original-role/value bindings rather than trusting a prior success flag. Unique captured identities can supply object existence without a binding when the existing owner permits it; field evidence still requires a schema selected for that original role. No schema is inferred by convenient field membership.

Fanout uses exact `math/big` product counting, rejects atomically rather than truncating, and retains distinct combinations even when candidate bytes coincide. Prepared artifacts are detached and reusable, with call-local query sessions and selections. Typed `Resolve` compiles artifacts once through preparation; strict `ResolveJSON` preserves existing decode-time artifact validation and may compile again during preparation. This is an admission-preserving tradeoff, not a cross-call assessment cache.

The new compatibility report includes optional `effective_dependency_bindings` beside submitted `dependency_bindings` because the existing closure report has no exact binding carrier. Canonical analysis and requirements remain query-only reports; original-role wrappers carry assessment identity without changing candidate IDs.

## Anything the next milestone needs to know

Consume `resolved_query` only from verified variants. `candidate_text`, successful parsing, positive object evidence, and neighboring verified variants do not establish publication authority. Correlation/disconnectedness is a separate finding. Classic SPL implicit source attribution, dynamic/opaque owners, incomplete discovery, ambiguous ownership, and bounded/cyclic/unavailable closure retain conservative limits.

Resolution edits only the submitted document. Supplied definitions are immutable; definition-only markers do not become resolution selections and can prevent verified publication. This milestone adds no runtime discovery, Data Source selection, registry, search execution, definition rewriting, deployment, or Milestone 17 workflow. No production dependency or ANTLR/parser contract changed.

`f705d45` changes only the Make receipt filename recipe and its actual dry-run regression relative to accepted product head `e87bbc6`; products, fixtures, schemas, examples, manifests, and acceptance checker are unchanged. This milestone log adds documentation only. Old receipts retain their actual source SHA: neither the recipe change nor this log turns them into same-head evidence. Any future aggregate acceptance must obtain the required records for the source identity its verifier requires, including hosted records. Preserve run-owned immutable receipts and the persistent worktree.

## Deviations from the PRD and why

There is no expansion beyond Milestone 16. The approved whole-identity design makes the PRD's “other supported identities” concrete as six kinds; arbitrary fragments and embedded interpolation remain unsupported.

The planned new C `const char*` declaration disagreed with CGO's generated `char*` ABI. Real package header verification caught this, and the maintained declaration plus representative header fixtures were minimally corrected. Request buffers remain read-only by implementation and unchanged-buffer tests; symbol and result ownership contracts did not change.

Authored fixture preparation needed a required conditional seed guard. Limit-case bindings added during earlier preparation were restored to their original empty arrays after the focused `c550aaf` compatibility owner fix preserved unique unbound object evidence. Counts, order, query, outcome, and negative schema-borrowing oracles were retained. Real sdist acceptance also required copying/hash-binding the maintained resolution example into the existing installed contract module; no checker was weakened.

Installed and clean-export checks were not repeated for the receipt-recipe-only change or this log, because their tested product inputs did not change. The affected Make/tooling gate was rerun. Hosted aggregate acceptance remains the explicit external gap described above; no historical payload was relabeled and no hosted record was fabricated.
