## What's new in the app

- Structured SPL analysis now covers the selected `tstats`, field-command, and function forms with source-located field-flow semantics and direct requirements.
- Exact macro invocations retain their direct macro requirement and an unresolved-expansion gap instead of erasing sibling facts.
- `join`, `append`, and `appendpipe` retain parent and child evidence while marking unmodeled merged output fields uncertain.
- The minimized detection corpus improved from 12 to 14 syntax-complete cases, from 1 to 16 semantic-complete target stages, and from 35 to 13 non-macro gaps.
- The SPL capability revision is `sha256:08901c84ac8c420f59ddb86168c534f0e83484c05d3c8a81078a716c878a1733`. The SPL2 revision remains `sha256:f1391296cfbc616e9bb1b1828e2471e37b60a35c0555654c0734640e072a0437`.

## What was built

The analyzer now supports bounded classic SPL forms through the existing analysis, field environment, dependency, requirement, and capability paths:

- `tstats` supports registered aggregates with exact optional aliases; a literal `datamodel=Model` or `datamodel=Model.Dataset` source; exact WHERE dependencies and field reads; exact BY groups; positive unsigned numeric `_time` spans, with an integer magnitude when unit-bearing; case-insensitive boolean values for `summariesonly`, `local`, `include_reduced_buckets`, and `allow_old_summaries`; a non-negative integer `chunk_size`; an exact literal `fillnull_value`; and exact `prestats=false` and `append=false`.
- `fillnull` supports an optional literal value and exact target fields. All-fields mode preserves the current field shape.
- `rex` supports default `_raw` or exact input, a quoted pattern, literal `max_match`, exact `offset_field`, and unambiguous `(?<name>...)` or `(?P<name>...)` captures.
- `spath` supports default `_raw` or exact input, an exact literal or dotted bare path, and an explicit exact output.
- `bin` and `bucket` support one exact input, optional exact alias, non-zero numeric `span` or `minspan` with an optional unit and an integer magnitude when a unit is present, positive integer `bins`, numeric `start` or `end`, and `aligntime` as a time literal, `earliest`, or `latest`.
- `regex` supports a quoted expression against `_raw` or an exact field with `=` or `!=`.
- `mvexpand` supports one exact field and preserves field identity.
- Expression functions support two or three arguments for `mvindex`; two arguments for `mvfind`, `relative_time`, `strftime`, and `like`; and zero arguments for `true`, `null`, and `now`. Aggregate context supports one argument for `earliest`, `latest`, and `stdev`.

Typed command operands remain outside safe-rewrite support. Held forms stay source-located and incomplete: inline macro expansion, dynamic catalog or field identities, wildcard `tstats` groups, spans on non-`_time` groups, signed relative-time spans, malformed or quoted spans, zero spans, fractional spans with a unit, logarithmic or snapped spans, dynamic spans, `PREFIX(...)`, `prestats=true`, `append=true`, unknown options or output identities, ambiguous captures, `rex` sed mode, `spath` auto-extraction, wrong function arity or context, and all branch merge effects.

Exact macros produce one direct macro requirement and one unresolved-expansion gap at the invocation. The analyzer does not load the macro body or invent transitive requirements. Branch children keep references, dependencies, diagnostics, transitions, and direct requirements. Join keys are parent reads, appendpipe inherits a copy of the parent environment, and no child field is copied into the parent. The parent retains known bindings in an uncertain environment.

The capability ledger now has 105 SPL records backed by 104 evidence cases. SPL syntax is 76 supported, 1 unsupported, and 28 unassessed; semantics is 68 supported, 9 unsupported, and 28 unassessed; requirements are 19 supported, 5 unsupported, and 81 unassessed; linting has 105 unassessed records; and safe rewriting is 17 supported, 2 unsupported, and 86 unassessed. The SPL2 ledger remains at 114 records and 124 evidence cases.

The impact comparison uses the same local minimized query bytes before and after the change. Its three measures are syntax-complete cases, semantic-complete selected target stages, and non-macro gaps. Whole-query completeness is excluded because a correct macro or branch boundary can keep a query incomplete after the selected stage is understood.

## Decisions made during implementation

The analyzer uses typed command contexts and the existing transfer kernel rather than a second semantic engine. `tstats` owns its data-model, dataset, predicate, group, aggregate, and macro facts so generic dependency traversal cannot duplicate them.

Value-only options use key-specific domains. `tstats` closes aggregation only when every output-affecting operand is understood. A held form retains independently sound groups and outputs conditionally in an open, uncertain environment. `bin` accepts only the modeled literal domains for spans, counts, numeric bounds, and alignment.

Named `rex` captures use a linear rune scanner. It masks escapes, character classes, POSIX classes, comment groups, and quoted literal regions. It does not call a regular expression engine, and ambiguous or unsupported group syntax remains incomplete.

Macro identity remains analysis evidence, not a rewrite kind. This keeps the approved rewrite surface unchanged while exposing the exact invocation as a dependency and direct requirement.

The capability ledger uses separate records for complete forms and held macro or branch boundaries. A record is not labeled partial without complete and incomplete evidence for the same exact scope.

## What the next milestone needs to know

The local source and surface checks completed before this log included:

- Focused parser, field-flow, requirement, branch, macro, rewrite-boundary, capability-corpus, and detection-impact tests.
- `env GOWORK=off go test ./pkg/analysis ./cmd ./pkg/api ./pkg/bindings -count=1`.
- `python3 -m pytest tools/tests/test_acceptance.py -q`.
- `make build-all`.
- `env PIP_FIND_LINKS=file:///private/tmp/spl-toolkit-offline-wheelhouse make python-test`, including direct-wheel and rebuilt-sdist acceptance. Each package path passed 562 native, 257 cross-surface, and 18 tooling tests.
- Byte comparisons showing that regenerated `docs/docs.go`, `docs/swagger.json`, and `docs/swagger.yaml` were unchanged.
- `python3 tools/check_docs.py`, which validated YAML front matter for all 17 maintained documentation pages.

The frozen surface corpora contain 38 analysis and 24 requirements reports. Go, CLI, REST, native C, and Python use those same canonical reports. Both new analysis source files are listed in the native and release source manifests, while the test-only detection-impact fixture is excluded.

The local evidence does not certify hosted CI, cross-platform release jobs, live Splunk behavior, publication, deployment, or UAT. The toolkit does not execute SPL, evaluate regular expressions, compare result rows, model acceleration, load macro definitions, or certify runtime compatibility. Macro expansion, branch merging, and the separate SPL2 roadmap remain open boundaries.

## Deviations from the PRD and why

None
