## Reopened acceptance gate (2026-09-28)

The 36/45 semantic-complete result below is historical evidence, not the final Milestone 11 acceptance target. The user approved typed public field identity in schema version 1 without a compatibility version because there are no existing consumers. Tasks 14–19 in the living plan preserve distinct atomic and structural identities through reports and require 45/45 standalone semantic completeness at the exact Linus ref. Definite unavailable structural reads remain invalid where proved; safe rewriting and live runtime behavior are separate claims. Task 14 public reference and binding identities passed bounded reviews at `d9510a9`; Task 15 typed removals and transitions at `359de99`; Task 16 exact field requirements and dynamic-field correction at `abdb2ea`; Task 17a schema and capability evidence at `6208141`; Task 17b document/graph projection and rewrite refusal at `28953e7`; Task 17c1 canonical report fixtures and unselected-join correction at `c4de46c`; Task 17c1b embedded validation reports at `576624e`; Task 17c2 transport and byte-exact examples at `b72bc14`. Task 18's renewed exact-ref audit and documentation passed bounded reviews at `70b0c69`, and a requirements-attribution correction passed quality review at `192afd8`. A packaged-acceptance name-only projection correction passed bounded reviews at `b05e1a7`. Final branch review found duplicate report order entries after a field was removed and recreated; RED public-report regressions and the central typed-key deduplication at `02cc07b` passed independent spec and quality reviews. The local validation gates below are green.

## Current exact-ref evidence (2026-09-28)

Remote Linus `refs/heads/main` and the clean 54-YAML sparse snapshot matched `d4db9bae4adba00bc59d9becbf4014908e7a3ef5`; toolkit `origin/main` was `ec9f626c03f67fe783c3dfc1d409ff75a12bbca9`. At feature code commit `02cc07bc7f948dfc2726caafaef19213d665440e`, the freshly built CLI and strict `tools/audit_linus_spl2.py` gate exited 0. Its aggregate-only result covered 49 searches (45 standalone, 4 predicate fragments) and 65 form obligations. All 45 standalone queries were syntax- and semantic-complete: 40 valid and 5 invalid. All 4 fragments were complete and valid. The invalid standalone reports retained 25 correctly categorized `SPL_UNAVAILABLE_FIELD` diagnostics; standalone requirements also retained 25, with 11 valid, 29 incomplete, and 5 invalid requirement reports. Neither surface emitted selected `SPL_AMBIGUOUS_FIELD` or `SPL_UNSUPPORTED_SEMANTICS` diagnostics for external content. The strict gate checks unavailable-field attribution per analysis and requirement report, not just aggregate totals. No external query text, names, source paths, or YAML metadata were copied into this repository or the audit output.

The public schema remains version 1 under the approved pre-consumer correction. Field references, bindings, removals, exact requirement items, and proved transitions expose typed identity so quoted atomic `actor.name` and structural `actor.name` are distinct despite equal display spelling. The current SPL2 capability manifest has 140 records, 146 evidence cases, semantic counts 76 supported / 38 unsupported / 26 unassessed, requirement counts 46 / 6 / 88, and revision `sha256:a765813624c6edfd754a4529556f5fd35288f464e759f8d969404a072bef1898`. Requirement completeness, safe rewrite, live Splunk execution, hosted CI, deployment, approval, and UAT are not implied by the 45/45 semantic result.

## Reopened local validation (2026-09-28)

- `GOWORK=off GOCACHE=/private/tmp/spl-toolkit-go-cache-final go test ./... -count=1`: passed on final code commit `02cc07b` with approved loopback access. The initial sandbox-only HTTP run could not bind a test server.
- `GOWORK=off GOCACHE=/private/tmp/spl-toolkit-go-cache-final python3 tools/check_go.py --race-timeout=20m`: formatting, vet, and all race-enabled Go package tests passed with approved loopback access. No Go code changed afterward.
- `GOWORK=off GOCACHE=/private/tmp/spl-toolkit-go-cache-final make build-all`: passed. The first sandboxed attempt could not replace the generated C header in the isolated worktree; the identical command passed with approved worktree access.
- `PIP_FIND_LINKS=file:///private/tmp/spl2-package-wheels-task13 GOWORK=off GOCACHE=/private/tmp/spl-toolkit-go-cache-final make python-test`: passed again at `02cc07b` with the pinned offline wheelhouse. The direct wheel and rebuilt sdist each passed 571 native, 259 installed-surface, and 19 machine-contract tests. An earlier run found the stale name-only acceptance projection, corrected at `b05e1a7`; a subsequent all-green run failed only its final Git-status guard when the separate audit correction was committed concurrently. Both stable reruns passed that guard.
- `GOWORK=off GOCACHE=/private/tmp/spl-toolkit-go-cache-final python3 -m pytest tools/tests -q`: 340 passed with 138 subtests. It emitted a dependency deprecation warning and a sandbox-only pytest-cache-write warning.
- `python3 tools/check_spl2_corpus.py`: 1,835 active obligations and 1,714 canonical queries passed; `python3 tools/check_docs.py`: all 17 checked pages passed; `python3 -m unittest tools.tests.test_linus_spl2_audit -q`: 27 tests passed. The final strict external audit passed at `02cc07b` with the aggregate counts above.
- The pinned ANTLR 4.13.2 complete JAR matched SHA-256 `eae2dfa119a64327444672aff63e9ec35a20180dc5b8090b7a6ab85125df4d76`; generation into a clean temporary directory reproduced all 10 tracked parser files byte for byte. No generated parser files were overwritten for this check.

This local evidence does not establish hosted CI, live Splunk execution, release, deployment, approval, or UAT. Direct-requirement coverage is separately incomplete for 29 of the 45 standalone searches, and the supported safe-rewrite ledger forms do not imply that arbitrary external queries are safe to rewrite.

The earlier Task 13 record below is retained for provenance. Its 36/45 gate, collision behavior, capability revision, and next-step wording describe that prior snapshot, not the current implementation.

## What's new in the app

- SPL2 analysis under `splunkd/current` now models selected field and dataset forms, same-document views and scalar functions, and selected `if`, guarded `branch`, `union`, and qualified pipeline `join` merges.
- Unquoted dotted SPL2 access is structural; a quoted identifier containing dots is atomic. They retain separate private identities. A collision in their shared public name reports `SPL_AMBIGUOUS_FIELD` and leaves semantic and requirement coverage incomplete.
- Exact static dataset descriptors have canonical JSON identities in dependencies and direct requirements. Dynamic descriptors retain located conditional evidence without an exact dependency.
- External imports are recorded but not resolved. Unused imports can leave analysis valid while direct requirement coverage is incomplete; using an imported member makes analysis incomplete.
- The SPL2 capability manifest now has 140 records and 146 evidence cases. Its semantic revision is `sha256:69b166318f99909d0ffbad378f0369fd9377a1f56945e2c3c0b69eaa32c03e95`.

## What was built

The parser and analyzer now handle the selected `splunkd/current` forms through the canonical Go result and its CLI, HTTP, native, and Python surfaces. Bounded expression and function forms, source and aggregate effects, static dataset descriptors, local declarations, and selected flow merges produce located references, lineage, diagnostics, dependencies, and direct requirements. The current ledger has 89 supported, 19 unsupported, and 32 unassessed syntax records; 75 supported, 39 unsupported, and 26 unassessed semantic records; and 45 supported, 7 unsupported, and 88 unassessed requirement records. All 140 lint records are unassessed. Safe rewriting has 14 supported, 3 unsupported, and 123 unassessed records.

The external exact-ref audit read 54 YAML documents containing 49 searches: 45 standalone queries and 4 fragments. The form inventory held 65 obligations. All 45 standalone queries and all 4 fragments passed syntax analysis. All 4 fragments were also semantic and requirement valid. Of the 45 standalone queries, 36 were semantically complete: 34 valid and 2 invalid. The other 9 had incomplete status and semantics, and emitted 30 `SPL_AMBIGUOUS_FIELD` findings caused by quoted atomic and unquoted structural identities sharing a public name. The 2 semantically complete but invalid reports emitted 7 `SPL_UNAVAILABLE_FIELD` errors from definite absent structural reads after only an atomic aggregate output remained. External search content emitted no `SPL_UNSUPPORTED_SEMANTICS`; combined generic form and content counters do include unsupported diagnostics.

## Implementation decisions

The analyzer keeps atomic and structural field identities in private keys while retaining the established string-valued public report. When two private identities project to one public name, it combines their visible origins and marks the result uncertain under the previously approved schema-version-1 collision rule. On 2026-09-28, the user confirmed that unquoted dotted access is structural and a quoted identifier containing dots is atomic. The audit must retain incomplete coverage rather than equate those identities.

Static descriptor identity is canonical JSON over a proved literal kind and literal properties. A duplicate decoded key or dynamic value cannot become an exact dataset dependency. Local views and pure scalar functions bind within the submitted document, with forward references, duplicate detection, and cycle detection. Imported modules are not fetched or linked.

Alternative flow paths use copied environments and forked requirement traces. Fields or obligations present on only some paths remain conditional. Selected pipeline joins compose matched left and right fields, then include unmatched sides for left or outer joins. Unproved layouts preserve independently sound child evidence without installing guessed parent outputs. Parser registration, semantic completeness, requirement completeness, lint support, and safe rewrite support retain separate evidence rules.

## What the next milestone needs to know

Milestone 11 local acceptance passed with the exact-ref classification enforced by the audit at `c33b592`. It requires 45/45 standalone syntax-complete searches, 36/45 semantic-complete searches (34 valid, 2 invalid), 9 incomplete searches with `SPL_AMBIGUOUS_FIELD`, and 4/4 valid, complete fragments. The 2 invalid searches have definite unavailable structural reads after atomic aggregate output. A later inventory or ref change needs a new evidence-backed baseline. No external search text, document names, or source paths are recorded here.

## Deviations from the plan

The external audit used a temporary sparse checkout of the fetched exact ref, with only YAML inputs in its working tree, as the `--content-root`. It did not use a tar archive. The snapshot resolved to `d4db9bae4adba00bc59d9becbf4014908e7a3ef5`, verified against remote main. The toolkit implementation base was `ec9f626c03f67fe783c3dfc1d409ff75a12bbca9`; the researched grammar baseline was `4d3b7850a47c7f493f4e4c3cae8c3f30ae4cc5fe`. Task 10 was committed as `df982b2a3c7fc487f1373a430df3046b2c5a8521`, and the Task 11 audit tooling as `a569f4fd3a0882e8717ea51eccd43ba1fed2fd0d`.

Four additional reference guides received only current SPL2 ledger-total and revision corrections. Their older milestone evidence and other prose were left intact.

The documentation landing page and contributor guide received separate narrow corrections to distinguish selected same-document declarations from unselected SPL2 module forms and external module resolution.

## Exact local validation

Task 10 validation included:

- `go test ./cmd ./pkg/api ./pkg/bindings -count=1`: passed.
- After `make build-shared`, `PYTHONPATH=python SPL_NATIVE_LIBRARY=$PWD/build/libspl_toolkit.dylib SPL_SPL2_FIXTURES=$PWD/testdata/spl2 python3 -m pytest python/tests/test_native_analysis.py python/tests/test_native_requirements.py python/tests/test_native_spl2.py tools/tests/test_package.py -q`: 266 passed.
- Installed direct-wheel and rebuilt-sdist-wheel acceptance: each passed 568 native and 258 surface tests, including 18 machine-contract tests.

Task 11 validation included:

- `GOWORK=off make build`: passed.
- `python3 -m unittest tools.tests.test_linus_spl2_audit`: 18 tests passed.
- `python3 tools/audit_linus_spl2.py --content-root "$LINUS_AUDIT_SNAPSHOT" --toolkit-bin build/spl-toolkit --forms testdata/spl2/linus-forms.json`: exited 0 against the exact-ref sparse checkout. It reported 36/45 semantic-complete standalone queries.
- `python3 tools/check_docs.py` and `git diff --check`: passed for this documentation update.

Task 13 fresh local acceptance at `2c5dac416122813fbaebab64bba538da774a655b` included:

- `python3 -m unittest tools.tests.test_linus_spl2_audit tools.tests.test_spl2_corpus tools.tests.test_package`: 81 passed.
- `python3 tools/check_spl2_corpus.py`: passed with 1,835 active obligations, 1,714 active and canonical queries, 288 meaningful cases, and 65 supplemental form obligations.
- `python3 tools/check_docs.py`: passed for all 17 checked documentation pages.
- `go test ./...`: passed. `python3 tools/check_go.py --race-timeout=20m`: passed with authorized loopback access. Its first sandboxed run was blocked by `httptest` binding to `[::1]:0`.
- `make build-all`: passed. `make python-test`: passed after provisioning the hash-locked offline wheelhouse and setting `PIP_FIND_LINKS`. The direct wheel and rebuilt source-distribution wheel each passed 568 native and 258 surface tests, including 18 machine-contract tests. The first package run lacked the required offline wheelhouse and stopped before installed-package tests. The successful run used:

  ```sh
  test ! -e /private/tmp/spl2-package-wheels-task13 && mkdir /private/tmp/spl2-package-wheels-task13 && printf 'wheelhouse ready\n'
  python3 -m pip --isolated download --disable-pip-version-check --no-cache-dir --index-url https://pypi.org/simple --only-binary=:all: --require-hashes -r tools/requirements-package-check-hashed.lock --dest /private/tmp/spl2-package-wheels-task13
  PIP_FIND_LINKS=file:///private/tmp/spl2-package-wheels-task13 make python-test
  ```
- `python3 -m pytest tools/tests -q`: 329 passed with 96 subtests. It reported a dependency deprecation warning and a sandbox-only pytest cache-write warning.
- Pinned ANTLR 4.13.2 generation into a clean temporary directory reproduced all 10 tracked `parser/spl2/` files byte for byte; the generator JAR matched SHA-256 `eae2dfa119a64327444672aff63e9ec35a20180dc5b8090b7a6ab85125df4d76`. The commands run from the repository root were:

  ```sh
  shasum -a 256 /private/tmp/spl-toolkit-remaining-tools/antlr-4.13.2-complete.jar
  set -eu
  gen_tmp=$(mktemp -d /private/tmp/spl2-antlr-acceptance.XXXXXX)
  java -jar /private/tmp/spl-toolkit-remaining-tools/antlr-4.13.2-complete.jar -Dlanguage=Go -package spl2 -visitor -listener -Xexact-output-dir -o "$gen_tmp" grammar/SPL2Lexer.g4
  java -jar /private/tmp/spl-toolkit-remaining-tools/antlr-4.13.2-complete.jar -Dlanguage=Go -package spl2 -visitor -listener -Xexact-output-dir -lib "$gen_tmp" -o "$gen_tmp" grammar/SPL2Parser.g4
  tracked_count=0
  for tracked_file in $(git ls-files parser/spl2); do
    cmp "$tracked_file" "$gen_tmp/${tracked_file##*/}"
    tracked_count=$((tracked_count+1))
  done
  generated_count=$(find "$gen_tmp" -type f | wc -l | tr -d ' ')
  test "$tracked_count" -eq "$generated_count"
  ```

  Both counts were 10, every `cmp` exited 0, and tracked generated files were not overwritten.
- The external repository's remote main matched `d4db9bae4adba00bc59d9becbf4014908e7a3ef5`. A fresh exact-ref sparse worktree contained only its 54 tracked YAML inputs. With `LINUS_SOURCE` set to the private source checkout and `LINUS_AUDIT_SNAPSHOT` set to a fresh path under the persistent `~/.worktrees` area, the reproducible setup and audit commands were:

  ```sh
  set -eu
  test "$(git -C "$LINUS_SOURCE" ls-remote origin refs/heads/main | awk '{print $1}')" = d4db9bae4adba00bc59d9becbf4014908e7a3ef5
  git -C "$LINUS_SOURCE" fetch --no-tags origin refs/heads/main
  git -C "$LINUS_SOURCE" worktree add --detach --no-checkout "$LINUS_AUDIT_SNAPSHOT" d4db9bae4adba00bc59d9becbf4014908e7a3ef5
  git -C "$LINUS_AUDIT_SNAPSHOT" sparse-checkout set --no-cone '*.yaml' '*.yml'
  git -C "$LINUS_AUDIT_SNAPSHOT" sparse-checkout reapply
  git -C "$LINUS_AUDIT_SNAPSHOT" read-tree -mu HEAD
  test "$(git -C "$LINUS_AUDIT_SNAPSHOT" rev-parse HEAD)" = d4db9bae4adba00bc59d9becbf4014908e7a3ef5
  test "$(find "$LINUS_AUDIT_SNAPSHOT" -type f \( -name '*.yaml' -o -name '*.yml' \) | wc -l | tr -d ' ')" = 54
  test -z "$(git -C "$LINUS_AUDIT_SNAPSHOT" status --porcelain)"
  python3 tools/audit_linus_spl2.py --content-root "$LINUS_AUDIT_SNAPSHOT" --toolkit-bin build/spl-toolkit --forms testdata/spl2/linus-forms.json
  ```

  The variables' private path values are omitted to avoid exposing source paths. The audit exited 0 after the final build: 49 searches (45 standalone, 4 fragments), 65 form obligations, 45/45 standalone syntax complete, 36/45 standalone semantics complete (34 valid, 2 invalid, 9 incomplete), and 4/4 fragments complete and valid. Standalone analysis emitted 30 `SPL_AMBIGUOUS_FIELD` and 7 `SPL_UNAVAILABLE_FIELD` findings; selected external content emitted no `SPL_UNSUPPORTED_SEMANTICS`. The command did not yet enforce the revised semantic classification.
- `git diff --check origin/main...HEAD`, `git status --short --branch`, and `git diff --stat origin/main...HEAD` were checked. The branch diff contained 109 files (11 added, 98 modified), all within the Milestone 11 parser, analysis, evidence, tests, packaging, audit tooling, and documentation scope. No dependency lock, public schema, workflow, or Milestone 12 file changed. The original checkout's user-owned plan and design files remained untracked and uncommitted.

Task 13 acceptance was refreshed at reviewed code commit `9500aa2b25f4788dce841f231f9c304e6349bb04`. Its import re-export correction now treats exporting an imported symbol as a use of the unresolved external module; the new regression test checks incomplete semantic and requirement coverage with a located `SPL_UNRESOLVED_MODULE` finding. The refreshed checks were:

- `python3 -m unittest tools.tests.test_linus_spl2_audit tools.tests.test_spl2_corpus tools.tests.test_package`: 81 passed. `python3 tools/check_spl2_corpus.py`: passed with 1,835 active obligations, 1,714 canonical queries, and 65 supplemental form obligations. `python3 tools/check_docs.py`: passed for 17 pages.
- `go test ./...`, `python3 tools/check_go.py --race-timeout=20m` with authorized loopback access, and `make build-all`: passed.
- `PIP_FIND_LINKS=file:///private/tmp/spl2-package-wheels-task13 make python-test`: passed using the previously provisioned hash-locked offline wheelhouse. The direct wheel and rebuilt source-distribution wheel each passed 568 native and 258 surface tests, including 18 machine-contract tests.
- `python3 -m pytest tools/tests -q`: 329 passed with 96 subtests. It reported the same dependency deprecation and sandbox cache-write warnings.
- `git diff --name-only 02421b2...HEAD -- grammar/SPL2Lexer.g4 grammar/SPL2Parser.g4 parser/spl2`: empty at the reviewed code commit, so the earlier 10/10 pinned generated-byte comparison still applies.
- Remote main was reverified as `d4db9bae4adba00bc59d9becbf4014908e7a3ef5`. The clean exact-ref sparse snapshot still had 54 YAML inputs. `python3 tools/audit_linus_spl2.py --content-root "$LINUS_AUDIT_SNAPSHOT" --toolkit-bin build/spl-toolkit --forms testdata/spl2/linus-forms.json` exited 0 after `make build-all`: 49 searches (45 standalone, 4 fragments), 65 forms, 45/45 standalone syntax complete, 36/45 standalone semantics complete (34 valid, 2 invalid, 9 incomplete), and 4/4 fragments complete and valid. Standalone analysis again emitted 30 `SPL_AMBIGUOUS_FIELD` and 7 `SPL_UNAVAILABLE_FIELD` findings; selected external content emitted no `SPL_UNSUPPORTED_SEMANTICS`.

The full local checks passed at reviewed production code commit `9500aa2b25f4788dce841f231f9c304e6349bb04`. The later audit gate changes only documentation and audit tooling; it does not alter the production code covered by those checks.

Task 11's confirmed-contract update was committed as `55d68c9`, followed by audit enforcement at `c33b592`. After the final gate change, the focused audit suite passed 25 unit tests. The exact-ref CLI audit passed against `d4db9bae4adba00bc59d9becbf4014908e7a3ef5` and enforced 54 YAML documents, 49 searches, 45 standalone queries, 4 fragments, and 65 forms. All 45 standalone queries were syntax-complete; 36 were semantically complete (34 valid, 2 invalid), and 9 were incomplete with ambiguity. Standalone analysis counted 30 `SPL_AMBIGUOUS_FIELD` findings across the 9 incomplete reports and 7 `SPL_UNAVAILABLE_FIELD` findings across the 2 invalid reports. All 4 fragments were valid and complete in analysis and requirements. Selected external content emitted no `SPL_UNSUPPORTED_SEMANTICS`. Standalone requirements counted 10 valid, 30 incomplete, and 5 invalid reports, with 30 ambiguity and 25 unavailable-field findings. The gate checks aggregate counts, statuses, completeness, diagnostic codes and categories, and private per-query attribution; the aggregate JSON schema remains unchanged and privacy-safe. The full tooling suite passed with `GOCACHE` directed to a writable temporary cache and pytest's cache plugin disabled: 336 tests and 119 subtests, with one dependency deprecation warning. Its first run failed 3 Go-check tests because the default Go cache was unwritable; the corrected rerun passed. The revised exact-ref local acceptance gate is complete under the confirmed atomic-versus-structural contract.

## Unclaimed proof levels

Local parser and surface checks establish only the tested toolkit behavior. They do not establish complete SPL2 semantics, complete direct requirements for every form, lint support, safe rewriting for forms outside its evidence ledger, live Splunk runtime compatibility, hosted CI, deployment, authorization, approval, or UAT. The enforced external audit establishes the reported exact-ref classification, not 45/45 standalone semantic completeness. Nine standalone searches remain incomplete and two are semantically complete but invalid under the confirmed field-identity rule.
