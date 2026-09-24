## What's new in the app

- SPL2 analysis under `splunkd/current` now models selected field and dataset forms, same-document views and scalar functions, and selected `if`, guarded `branch`, `union`, and qualified pipeline `join` merges.
- Quoted atomic dotted fields and structural paths retain separate private identities. A collision in their shared public name reports `SPL_AMBIGUOUS_FIELD` and leaves semantic and requirement coverage incomplete.
- Exact static dataset descriptors have canonical JSON identities in dependencies and direct requirements. Dynamic descriptors retain located conditional evidence without an exact dependency.
- External imports are recorded but not resolved. Unused imports can leave analysis valid while direct requirement coverage is incomplete; using an imported member makes analysis incomplete.
- The SPL2 capability manifest now has 140 records and 146 evidence cases. Its semantic revision is `sha256:69b166318f99909d0ffbad378f0369fd9377a1f56945e2c3c0b69eaa32c03e95`.

## What was built

The parser and analyzer now handle the selected `splunkd/current` forms through the canonical Go result and its CLI, HTTP, native, and Python surfaces. Bounded expression and function forms, source and aggregate effects, static dataset descriptors, local declarations, and selected flow merges produce located references, lineage, diagnostics, dependencies, and direct requirements. The current ledger has 89 supported, 19 unsupported, and 32 unassessed syntax records; 75 supported, 39 unsupported, and 26 unassessed semantic records; and 45 supported, 7 unsupported, and 88 unassessed requirement records. All 140 lint records are unassessed. Safe rewriting has 14 supported, 3 unsupported, and 123 unassessed records.

The external exact-ref audit read 54 YAML documents containing 49 searches: 45 standalone queries and 4 fragments. The form inventory held 65 obligations. All 45 standalone queries and all 4 fragments passed syntax analysis. All 4 fragments were also semantic and requirement valid. Of the 45 standalone queries, 36 were semantically complete: 34 valid and 2 invalid. The other 9 had incomplete status and semantics, and emitted 30 `SPL_AMBIGUOUS_FIELD` findings caused by quoted atomic and unquoted structural identities sharing a public name. The 2 semantically complete but invalid reports emitted 7 `SPL_UNAVAILABLE_FIELD` errors from definite absent structural reads after only an atomic aggregate output remained. External search content emitted no `SPL_UNSUPPORTED_SEMANTICS`; combined generic form and content counters do include unsupported diagnostics.

## Implementation decisions

The analyzer keeps atomic and structural field identities in private keys while retaining the established string-valued public report. When two private identities project to one public name, it combines their visible origins and marks the result uncertain. This collision rule is intentional pending a user choice about whether to retain or revise it.

Static descriptor identity is canonical JSON over a proved literal kind and literal properties. A duplicate decoded key or dynamic value cannot become an exact dataset dependency. Local views and pure scalar functions bind within the submitted document, with forward references, duplicate detection, and cycle detection. Imported modules are not fetched or linked.

Alternative flow paths use copied environments and forked requirement traces. Fields or obligations present on only some paths remain conditional. Selected pipeline joins compose matched left and right fields, then include unmatched sides for left or outer joins. Unproved layouts preserve independently sound child evidence without installing guessed parent outputs. Parser registration, semantic completeness, requirement completeness, lint support, and safe rewrite support retain separate evidence rules.

## What the next milestone needs to know

Milestone 11 acceptance is open. The audit command exited 0, but only 36 of 45 standalone searches are semantically complete under the current collision rule. A user decision is pending on whether that rule should remain conservative or be revised. The two invalid reports also require preserving the structural-versus-atomic distinction when assessing post-aggregate reads. No external search text, document names, or source paths are recorded here.

## Deviations from the plan

The external audit used a temporary sparse checkout of the fetched exact ref, with only YAML inputs in its working tree, as the `--content-root`. It did not use a tar archive. The snapshot resolved to `d4db9bae4adba00bc59d9becbf4014908e7a3ef5`, verified against remote main. The toolkit implementation base was `ec9f626c03f67fe783c3dfc1d409ff75a12bbca9`; the researched grammar baseline was `4d3b7850a47c7f493f4e4c3cae8c3f30ae4cc5fe`. Task 10 was committed as `df982b2a3c7fc487f1373a430df3046b2c5a8521`, and the Task 11 audit tooling as `a569f4fd3a0882e8717ea51eccd43ba1fed2fd0d`.

Four additional reference guides received only current SPL2 ledger-total and revision corrections. Their older milestone evidence and other prose were left intact.

The documentation landing page and contributor guide received separate narrow corrections to distinguish selected same-document declarations from unselected SPL2 module forms and external module resolution.

## Exact local validation

Task 10 validation included:

- `go test ./cmd ./pkg/api ./pkg/bindings -count=1`: passed.
- After `make build-shared`, `PYTHONPATH=python SPL_NATIVE_LIBRARY=$PWD/build/libspl_toolkit.dylib SPL_SPL2_FIXTURES=$PWD/testdata/spl2 python3 -m pytest python/tests/test_native_analysis.py python/tests/test_native_requirements.py python/tests/test_native_spl2.py tools/tests/test_package.py -q`: 266 passed.
- Installed direct-wheel and rebuilt-sdist-wheel acceptance: each passed 568 native, 258 surface, and 18 machine-contract tests.

Task 11 validation included:

- `GOWORK=off make build`: passed.
- `python3 -m unittest tools.tests.test_linus_spl2_audit`: 18 tests passed.
- `python3 tools/audit_linus_spl2.py --content-root "$LINUS_AUDIT_SNAPSHOT" --toolkit-bin build/spl-toolkit --forms testdata/spl2/linus-forms.json`: exited 0 against the exact-ref sparse checkout. Its semantic acceptance result remains 36/45 standalone queries.
- `python3 tools/check_docs.py` and `git diff --check`: passed for this documentation update.

Other earlier tasks had focused green checks; no full Milestone 11 acceptance run is claimed here.

## Unclaimed proof levels

Local parser and surface checks establish only the tested toolkit behavior. They do not establish complete SPL2 semantics, complete direct requirements for every form, lint support, safe rewriting for forms outside its evidence ledger, live Splunk runtime compatibility, release-platform CI, deployment, authorization, approval, or UAT. The external audit does not establish 45/45 standalone semantic acceptance while the collision decision and remaining findings are open.
