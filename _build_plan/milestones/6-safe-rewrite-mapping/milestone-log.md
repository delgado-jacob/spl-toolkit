## What's new in the app

Broad-review follow-up: the later `fbde61b` checkpoint was not accepted as final
M6. The broad-review fix integrates supported SQL HAVING predicates at their
restricted logical phase and shares exact input-sized JSON decimal normalization
between canonical fact merges and condition comparison. The final corpus has 125
public cases and 106 required groups, with 24 additional raw-native exact-number
regressions. The broad-review verification section below records the final
candidate placeholder, fresh artifact closure and the explicit Python numeric
serialization boundary. Broad rereview and renewed root acceptance remain separate.

Task 8 review follow-up: the initial `c3a175b` evidence below is historical and
was not accepted as final. Scoped producer correction
`e76d351d58a73d17b04d4f7e1728a3ccedaca2a9` is independently approved: literal
dependency facts retain grammar selector identities and conflicting exact AND
restrictions cannot authorize rewrites. The final corpus now has 106 independent
public cases, 88 required semantic/form subgroups, all 31 advertised positive
render roles represented in candidate-applied full reports, and the preserved
34-row capability matrix plus two private proof-fault cases. It adds SQL logical
ordering, typed SPL2 conditions, selector/context/conflict controls, concrete
JSON Schema outcomes and local versioned OCSF, exact missing forms, byte/alias/
macro/wildcard/navigation boundaries, and real HTTP exact/over-limit checks.
Final refreshed artifact counts/hashes and follow-up commit identity are recorded
in the review-follow-up verification section below; broad review/root acceptance
and unexecuted release/runtime gates remain open.

SPL Toolkit now previews and applies explicit safe rewrite rules through canonical Go, CLI, REST and native Python. Rules map field, index, source, sourcetype, lookup, dataset and data-model identities in supported SPL/SPL2 forms. Reports retain original/candidate/returned text, exact audit entries, original facts, bindings, lineage, refusal evidence and optional destination validation. Explicit aliases remain fixed. A separately proved group may commit beside a refused group with incomplete status; a failed whole-candidate or destination gate returns the original string. Legacy mapping precedence and context semantics remain unchanged.

Task8 implementation is ready for independent review. This is not final milestone/root acceptance, release-platform certification or Splunk runtime equivalence.

## Accepted dependencies and candidate identity

Task8 began at exact `a7290ccbb530ff8f918362f0362e2f7c7ed1030d` (Task7 acceptance). A corpus capability check exposed the missing SPL null-inspection render role. The coordinator routed the narrow producer correction through independent review and released Task8 from `f6572c35e87a9a891f31ac06ece48bd086f204b6`; it preserves public SPL read semantics and bindings. Task8 makes no further production semantic changes.

Accepted implementation tips consumed:

| Unit | Commit |
| --- | --- |
| Task1 request/report contract | `3a29058949b5b0000e4b90d9406764dbe11ce2ee` |
| Task2 canonical evidence/render/proof bridge | `5745041676c8d6cddfb7b769c2aaa2630d3ab8a2` |
| Task3 original three-valued facts | `6c18c32a073b944375d1d37da7e47b768d16656a` |
| Task4 linked candidates/audit | `14a3a0dcfe4652240dc8b949461f905b206f215a` |
| Task5 core/batch/publication correction | `c8be29f6fead17e9e40a06073878edf2b2efa35e` |
| Task6 CLI/HTTP contract correction | `e724af4cd22ba6f29e53de99c572275db7627843` |
| Task7 native/package/floor header | `8e0e0b89a3c4b2825ca10f510d3a191f9b6dd1db` |
| Reviewed null-role producer dependency | `f6572c35e87a9a891f31ac06ece48bd086f204b6` |

Task8 final candidate is the immutable commit containing this log. Its full SHA is recorded after commit in the ignored `task-8-report.md` and post-commit evidence, avoiding a self-referential hash here. No merge or push is performed.

## Task8 files and durable coverage

- `testdata/rewrite/corpus.json` and `pkg/rewrite/corpus_test.go`: 62 independently specified requests/reports across 46 distinct dialect/query texts. Counts are 44 SPL / 18 SPL2, 42 valid / 16 incomplete / 4 invalid, with 34 committed candidates. Repeated text tests distinct modes, rules or targets, not spelling variants. Twenty-two required public groups fail when empty.
- `testdata/rewrite/forms.json`: all 34 advertised dialect/kind/role rows, with actual positive producer checks and negative context controls; the 19 predecessor facade forms remain. Two private candidate-syntax/binding fault fixtures retain full compact independent audit/reference/lineage expectations and force the real final gate to reject stale successful proof summaries. These are not additional public report conformance credit.
- The exact required `example-rules.json` drives preview/apply CLI acceptance for `search src=alice | rename src AS owner | table owner`.
- `tests/acceptance/test_rewrite_surfaces.py`: full canonical Go/actual CLI/real HTTP/native equality for all 62 reports; actual Go batch results for 39 rule/mode/target groups in reversed document order, including both dialects and a valid/incomplete/invalid group; three atomic malformed batch inputs. Batch groups reuse the single-query obligations rather than claiming 39 new semantic examples. UTF-8 source slices, Unicode coordinates, CRLF handling, inserted aliases and untouched bytes are independently checked.
- Package/acceptance tools register both rewrite suites and all seven used rewrite JSON inputs, preserving predecessor native/schema/SPL2/docs evidence, nonzero collection and zero-skip rules. Tests reject lost groups, duplicate IDs, missing/truncated/stale transport, stale fixtures and changed semantics. Go reports are transport only, never the semantic oracle.
- `python/examples/basic_usage.py` exercises real native preview, apply, destination rollback and ordered mixed-status batches. Maintained README/API/architecture/compatibility/install/configuration/contributor pages distinguish the new contract from legacy behavior; `docs/rewrite.md` documents supported identities, conditions, aliases, implicit names, collisions and remaining limits.

## Decisions, failures and deviations

TDD receipts retain missing-corpus, missing-suite/fixture/docs closure, missing capability matrix, missing private fault groups and missing executable example failures. The null-role defect was escalated, not hidden by changing a fixture. After the reviewed producer fix, independently reconciled draft errors corrected byte positions, reference IDs, public read/create vocabulary, and conditional SPL2 aggregate lineage. Unsupported source-less SPL2 pipelines, SPL2 inputlookup and arbitrary implicit labels remain explicit negatives. Pattern predicates weaken fact completeness; isolated overwritten-field cases exercise false/unknown boolean siblings without inventing absent-fact proof. The batch harness uses the canonical strict request decoder for schema targets.

The predecessor docs check failed on the retained `docs/evidence/milestone-5/broad-fix-1/README.md`. The coordinator explicitly authorized excluding exactly that file in Jekyll config, adding rewrite navigation and replacing the stale site description. The evidence README is byte-preserved (SHA256 `6f2593ea22619815de13791c90147b9e6fcc851adbc2999faefe6bca672838a7`); sibling evidence remains available and a regression checks that sibling public pages still require metadata.

The plan's historical module-cache path was incomplete. All qualified runs use the already provisioned `/Users/jacobdelgado/go/pkg/mod` with `GOPROXY=off`, `GOSUMDB=off` and `GOTOOLCHAIN=local`, plus official Go1.22.12 at `/private/tmp/spl-toolkit-go1.22.12-official/go/bin/go`. Packaging uses only `/private/tmp/spl-toolkit-offline-wheelhouse`, `PIP_NO_INDEX=1` and pinned build dependencies. No further downloads, dependency changes, generated parser/OpenAPI changes or private root acceptance inputs were used.

Shared outputs are preserved. Task-local builds live in ignored `task-8-evidence/build` and `dist`; existing ignored Python build staging was moved intact to `task-8-evidence/prior-python-build`. A wrapper runs the unchanged formal checker and copies its actual rebuilt wheel bytes before temporary environments are removed; it changes no validation or build result. The final installed run was repeated after the last maintained-document clarifications so copied documentation hashes match the candidate bytes.

## Qualified verification

Evidence directory: `.superpowers/sdd/milestone-6-implementation-plan/task-8-evidence/`. Every command receipt records exact argv, cwd, environment, exit and raw-log hash.

| Check | Actual result |
| --- | --- |
| `go test -mod=readonly ./pkg/rewrite -count=1`, corrected harness | passed |
| Official Go1.22.12 external-link rewrite/analysis/validation/mapper/bindings/cmd/api floor | all seven packages passed |
| `tools/check_go.py` | formatting/vet/full race passed; eight tested packages |
| `tools/tests` | 211 passed, zero skips |
| `tools/check_docs.py` | all 15 current site pages passed |
| Source-native `python/tests` | 435 passed, zero skips |
| Documented CLI/schema/SPL2/rewrite source surfaces | 147 passed, zero skips |
| Direct wheel outside checkout | 414 native + 171 surface tests passed; zero failures/skips |
| Rebuilt-sdist wheel outside checkout | 414 native + 171 surface tests passed; zero failures/skips |
| Full closure audit | current source/fixture/docs hashes and complete source/direct/rebuilt reports and batches match |

The installed suites preserve all 1,714 SPL2 documents and existing schema, validation, analysis, legacy mapping and CLI examples. Both installed libraries have SHA256 `e138aaee06568510133848530b0c1aba65e0b029fcc3c1e9ba2095c8214f12a6`. Module/library paths were observed inside separate Python3.12 site-packages trees outside checkout, not source imports; full origins are in final package evidence. Source Python is 3.12.6 on macOS26.2 arm64.

## Immutable artifact hashes

Paths below are relative to the ignored evidence directory; SHA256 values describe retained bytes.

| Artifact | SHA256 |
| --- | --- |
| CLI `build/spl-toolkit` | `298f551044522c6a4e39818831ab8f6ec482d5799147c1d64a4097c94a993524` |
| Server `build/spl-toolkit-server` | `192ecdc573c34a1b936eca3fd9b01ad44218ac82bcd315443528c6421e7305d4` |
| Source native `build/libspl_toolkit.dylib` | `1dceaa2e7f44717dce92aa3dbdd0858084417c0de9247274b6cd66f2700dafdf` |
| Actual Go1.22 header `build/libspl_toolkit.h` | `ce5f777f85c773ed7397c7e66fcf6ab59e9a6592e816f871cb1f4c6ea2e519ef` |
| Direct wheel | `199b888f18091c1120b37532dc6d0d3f4ad163129265f61f02eb640674679c3d` |
| Source distribution | `636b9cb6941eb5d2ea74ffe3c3623d611e5a8de4c32d36cda581b5234634c90a` |
| Final retained rebuilt-sdist wheel | `b2ac0af969f6c9c7b9a3ebe5686a6e3ab80b85e4566c298bc954fba0237347d7` |
| Final package evidence | `b65a6b247554514ea24e27378ab9e671d1b0d75b3102697f8032c9d0444d0320` |
| Source rewrite parity evidence | `48fda147d1f47b51830c028a806627c2f36c7f9a92e9f149ffa4d8c8e2e46ff9` |
| Final closure record | `5b51db113805357367b34593f98db0d44851eff1f014b6d30c95bcdaf5e3e730` |

Corpus SHA256 is `370e01b0cb84334b985dac8c1a36b743cdadb8df83778932f3dc5e22b1a7774d`; forms SHA256 is `9ab44b9538d1248b27c4c5d39963ab637d415b75fdbfcb4324ae3bfcfd7fac55`; example-rules SHA256 is `80f1953b3259e3df6244e96ea642112095a88fb4111d7d7689840376fd58584e`. Full Go transport SHA256 is `499c88b9b5ad31eac79d3501b594ae37514ee62429820fb4287b7ff3dbda6b47` and grants zero semantic credit.

## Remaining gates and qualifications

The fresh CLI's embedded VCS metadata names outer baseline `6898cc052eb7109fbcc7495ee826c993aaa93352` with `modified=true`, not the Task8 candidate. Recorded worktree build commands, actual Go package directories, source hashes, native manifests and full report parity bind its inputs; no final-commit stamp is claimed. Native wheel builds use the existing `-buildvcs=false` release recipe. The local Go external-link path is qualified for the known LC_UUID issue; race runs retain LC_DYSYMTAB warnings and floor logs retain deployment-target warnings without claiming other-host execution.

Independent Task8 review, broad M6 review and root's reserved stable Go/built/native/installed/Python3.14 acceptance remain separate. Linux/Windows/other architectures, supported Python release combinations, ASAN/leak-sanitizer execution and external Splunk runtime equivalence were not run here. No merge, push, release or Milestone7 implementation is implied. User-owned untracked files, GOAL_STATE, predecessor evidence and root-private inputs remain untouched.

## Review-follow-up verification

This section supersedes the historical Task8 candidate counts and artifacts above.
Review round one identified selector-identity and conflicting-AND producer defects
plus missing full-parity obligations. The separately committed producer fix
`e76d351d58a73d17b04d4f7e1728a3ccedaca2a9` received independent approval with
zero findings (`task-8-fact-fix-review.md`, SHA256
`5a1d302cd9edad40b7af7c3f54373ded5134cbe9a62e252e4827f7e17a951f5c`).
The follow-up commit containing this log changes only corpus/tests and documentation;
its exact SHA is resolved in `task-8-report.md` and `r1-post-commit.json` after commit.

The final 106 public cases (+44) cover 81 distinct dialect/query texts: 65 SPL,
41 SPL2; 76 valid, 25 incomplete, five invalid; 60 committed. All 88 required
groups (22 original umbrellas plus 66 semantic/form subgroups) have deletion
guards. All 31 advertised supported roles occur in actually changed full reports;
the 34-row capability matrix and two private proof-fault cases remain separate.
The 56 real ordered batch groups include 18 multi-document groups. Three malformed
batch inputs retain atomic rejection. Eight real-HTTP checks exercise exact 8 MiB
and one-byte-over bodies for single/batch and known-length/chunked requests.
Destination cases independently assert field-list, JSON Schema valid/invalid/
conditional/unresolved, and real local OCSF outcomes. The existing OCSF 1.6.0
catalog is reused with its raw SHA256, not duplicated or downloaded.

Seven obligation-deletion tests first failed because the old broad guards let
their cases disappear (`r1-obligation-red.log`, SHA256
`e4992e54217e8f851d113f9e40b1ae45eeaebd5aaa75b865e952220d4b820a11`).
Expanded guards and independent cases then passed. Draft corrections were
reconciled against canonical public roles, SQL phase order, deferred tstats output
effects, null/rename provenance, removal transitions and refusal audit semantics;
the report preserves that authoring ledger. No live report became an oracle.

Fresh official Go1.22.12/offline runs passed: seven-package floor; full formatting,
vet and race checks (eight tested packages); 222 tooling tests; 435 source-native
Python tests; 199 documented/schema/SPL2/rewrite surface tests; and a final 121-test
rewrite run binding the final fixture bytes. Both direct and rebuilt-sdist installed
paths passed 414 native plus 223 surface tests, with zero failures/skips. The docs
checker passed all 15 current pages. The initial floor attempt was denied a local
HTTP listener by the sandbox; the authorized local-listener rerun passed and is
the counted floor evidence. No test failure was waived.

Task-local refreshed artifacts are under `task-8-evidence/r1-build`, `r1-dist` and
`r1-rebuilt-sdist`; previous artifacts and ignored Python build staging were kept
intact. Full source/direct/rebuilt JSON values, ordered batches, atomic errors and
HTTP limits match. Closure retains all seven rewrite fixtures, 18 copied docs/
examples, 86 sdist source files, 165 producing Go source hashes and all predecessor
schema/SPL2 suites, including 1,714 SPL2 documents. Installed module/native origins
were observed in separate outside-checkout Python3.12 site-packages environments.

| Refreshed artifact | SHA256 |
| --- | --- |
| CLI | `d32e067be0f37cad0fb39b47a63d9d76a5e78054f0816f716ad16ed44732c524` |
| Server | `49ffd68641c647885489482f7727de1c4a8cfab7774e905b3e70a3e8f103b3b8` |
| Source native | `e0bfbea75b44b628a1fdca14816e3747439e6801b45a781846d8c068168213c7` |
| Both installed native libraries | `be4c7feb64d5d8e7565990b19604210b4816ae18fe3c39f2245226e6137186fd` |
| Direct wheel | `5ed1af89e7bdc7259f9e4f5e0444883765ef504d3b9e37248aa691cc34902b03` |
| Sdist | `e16362b51fd7bfb77f44f359ea87b2a9272f798566bbab42856f6bc2e071e540` |
| Rebuilt-sdist wheel | `af2750ade8d4959e6c602c91fec982ba1860525134f91f2ecc72cbd50f235531` |
| r1-package-evidence.json | `c0a2262ac2b46bb4ee9edef5245794652248a6de4e2d66d75aac467e368ad14c` |
| r1-final-source-rewrite-evidence.json | `58bb05c747e7e986cc916993a865a6b8c2bd102922d2bc7279b394b44a895d00` |
| r1-closure.json | `40f9774d795723e8de6a659caf2fb0ccfa08321177c9a1eb516abfb2d547339c` |

Final corpus SHA256 is `30078c0abe82f0e2c82c9efc90fb3acc8c99d252b1c920961dcd41abcf5e8c83`;
full Go transport SHA256 is `854fc4acd93d385964b0829a50899bd4845c0b76a9779c128e13d4d8c377b2d3`
(transport only, zero semantic credit). Exact command/log/input hashes are in the
closure record. The previously disclosed outer-revision `6898cc052`, modified=true,
CLI metadata and external-link warnings still apply; actual worktree source hashes
bind these binaries, not a claimed clean candidate stamp. Results remain local
macOS26.2 arm64/Python3.12.6 proof. Task8 rereview, broad M6 review, root stable and
Python3.14 acceptance, other release platforms, sanitizers and external runtime
conformance are unexecuted separate gates. No private acceptance inputs, new
dependencies, generated schema changes, user-file edits, merge or push occurred.

## Broad-review verification

This section supersedes the preceding historical Task8/r1 artifact counts. The
broad review at `fbde61b4383e473522355319fb45f7bf5b332d35` required two producer
corrections. The exact candidate containing this log is resolved after commit in
`milestone-6-broad-fix-report.md` and `task-8-evidence/broad-post-commit.json`;
it is a direct child of that reviewed base, not a claimed independently approved
producer. Existing accepted dependencies, including `f6572c35` and `e76d351`,
remain ancestors.

Supported HAVING predicates now use the shared canonical predicate reader inside
the same restricted visibility view as SQL references, at the HAVING phase.
Only its fact/reference evidence survives the temporary view. WHERE, GROUP BY,
selection, aliases, later ordering and diagnostics keep their established phase
contracts. Hidden operands do not suppress a conflicting visible restriction.
Ungrouped HAVING and derived/unsupported guarantees remain non-authorizing.
Fact merges and condition comparison share the existing exact decimal
coefficient/exponent normalization, factored into `internal/jsoninput/number.go`
and included in the native sdist manifest. It uses input-sized storage, never
expands powers of ten and never coerces numeric values through floats.

The independently authored corpus adds 19 cases: nine HAVING/visibility/phase
cases and ten numeric AND/OR cases. There are now 125 public reports over 100
distinct dialect/query texts (67 SPL, 58 SPL2): 90 valid, 30 incomplete, five
invalid, 71 committed. All 106 required groups have deletion guards, including
18 new HAVING/numeric groups. The 31 positively exercised supported roles, 34-row
capability matrix and two private proof-fault cases remain unchanged. Full
six-surface parity covers 61 ordered batches, including 20 multi-document groups,
three atomic request errors, eight real HTTP limits and the existing local OCSF
catalog. All predecessor schema/SPL2/native/surface obligations remain registered.

Meaningful RED is retained for missing HAVING and exact nonzero exponent facts,
the hidden-operand conflict refinement, and two new obligation-deletion tests.
Against the old native, 16 positive exact-number tests failed while eight
different-value controls passed; all 24 pass against the final rebuilt native in
source and both installed lanes. Independent corpus authoring corrections were
resolved from the grammar and canonical origin/visibility contracts, not copied
from producer reports. Initial pre-refinement successful artifacts and all failed
evidence remain historical; only the `broad-final-*` set below is final.

| Final qualified check | Result |
| --- | --- |
| Official Go1.22.12 floor | all seven packages passed |
| Full Go formatting/vet/race | eight tested packages passed |
| Source-native Python | 459 passed, zero skips |
| Source documented/schema/SPL2/rewrite surfaces | 218 passed, zero skips |
| Tooling tests | 224 passed, zero skips |
| Docs checker | 15 current pages passed |
| Direct wheel outside checkout | 438 native + 242 surface passed; zero failures/skips |
| Rebuilt-sdist wheel outside checkout | 438 native + 242 surface passed; zero failures/skips |
| Exact closure audit | full reports/batches/errors/limits and current source/fixture/docs hashes agree |

Closure binds 169 producing Go source hashes, 87 sdist source files, all seven
rewrite fixtures and 18 copied documentation/example files. Both installed native
libraries have SHA256 `fde1fa58452845a516e7f601029f154f1bb308afe20773eed429c01c2abab416`.
Observed imports and loaded libraries are inside separate outside-checkout
Python3.12 site-packages environments. The 12 Python3.12 wheelhouse inputs and
official compiler remain byte-identical to the prior input receipt; a separately
available cp314 wheel is inventoried, not claimed as execution of Python3.14.

| Final retained artifact | SHA256 |
| --- | --- |
| CLI | `b9248dd8bb1488fc270c17e0ed77340263e0b1523c264ef3519d30a2e40500d2` |
| Server | `946ba20d58758cc7dcaf5cad5a68f0e912c5b7d0bd882cf80ebdc908aca90714` |
| Source native | `acb2609c034585d4b571c4dc0ad6d19940d6cb5e51bb47bafd977b0608b5927a` |
| Direct wheel | `e385183999e29743cc096b7eb7449bfc8c35c023421e2cf65e4ea1d575e9ec76` |
| Sdist | `1524ff8e3729a996136a0070262fa4d9b005b82eedfc03a696b64680fb04333e` |
| Rebuilt-sdist wheel | `e4699fe9ab98dfef533c04a73ea8d17a996469c81a72973eb161fe4c888df8ec` |
| broad-final-package-evidence.json | `8cf4d383b60494672ae347d615b064ceda2aa1d154164988844f75d18443f6cc` |
| broad-final-source-rewrite-evidence.json | `33ccd2ee29216733eab9238d276c632b61e0cdbd4517ca46cad623bea8abf61a` |
| broad-final-closure.json | `6e0583b51d7e22835543204c4f8da7771d60a689d0d03aa3f1094f04273d5fa5` |

Corpus SHA256 is `26ad72ded80f3e7611e1a4763bc57f3abea000c0621db0f4dcc639d5ccab4121`;
Go transport SHA256 is `2f4f92ab6417d64faef7e34c82f0a19322be5ed032dd5cb0644ebec51782d5c9`
(transport only, zero semantic credit). Exact commands, exit statuses, origins,
official input hashes and log hashes are retained in the final closure.

The coordinator approved keeping Python's convenience value-serialization API
unchanged. Six-surface zero/exponent AND/OR cases exercise the shared producer
merge and exact-decimal path. They are not six-surface proof for a nonzero
arbitrary-precision numeric JSON token: those tokens are covered by exact Go/raw-C
tests. No Decimal API, float coercion or broader SPL exponent grammar is claimed.

Proof remains local macOS26.2 arm64/Python3.12.6/official Go1.22.12 with the
disclosed external-link workaround and retained LC_DYSYMTAB warnings. CLI VCS
metadata still reports outer baseline `6898cc052`, modified=true; actual worktree
source/command/hash evidence binds its inputs, not a claimed clean candidate
stamp. Broad re-review and renewed root stable/Python3.14 acceptance, other
release platforms, sanitizer/leak checks and external Splunk runtime conformance
remain separate. No coordinator-private inputs, dependency upgrades, generated
schema edits, user-file changes, merge or push occurred.

## Final root milestone acceptance — 2026-09-12

The preceding Task8/broad-fix producer checks remain intact. The independent
same-reviewer [broad rereview](../../../.superpowers/sdd/milestone-6-implementation-plan/milestone-6-broad-rereview-1.md)
approved exact product source `98e38b941ea868468376b9dd6533e684009d35c0`
with zero findings, closing both original Important issues. Root then renewed
all reserved acceptance windows against that exact clean tracked/index source:

- The official Go1.22.12-built CLI/server/native and independent 29-case Go,
  actual CLI, real HTTP and raw-native checks passed, with four ordered batch
  documents, strict errors, concurrency, exact original/candidate bytes and
  canonical validation parity. Receipt:
  `/private/tmp/spl-toolkit-root-m6-final-attempts/98e38b941ea8-surfaces-20260912T201835255501Z/attempt.json`,
  SHA256 `64ed6d3d053d18d3122633fcfcd1acc5c26d0611c66ccfa0713ed58a8299e509`.
  CLI/server/native SHA256 respectively `42dc3732aefd0d854565540040206d653b424f84b754b34c922af9117be635a0`,
  `738c54683a728dd6015bf255b7ce7d05da41efef175b3c92593bf3c9a0b42bbe`,
  `74a0feb7fe5368e5e967a42482b3344b8427278b3c527015a6f2a04e45e75f1f`.
- Fresh Python3.14.1 source/native tests passed: 459, zero skips.
  Receipt `/private/tmp/spl-toolkit-root-m6-final-attempts/98e38b941ea8-py314-20260912T201853032997Z/attempt.json`,
  SHA256 `0a543de55c4a470c3a52d81c5027d1224b174ee8f8cb46d242b83a210b12c660`;
  result `/private/tmp/spl-toolkit-root-m6-acceptance-py314.json`,
  SHA256 `d19b99718504c1d64851e986a0bd6b8eaf6762d6490c6cfcca7f412d6f4fe8ae`.
- Fresh outside-checkout direct wheel and independently rebuilt-sdist wheel
  each passed the unchanged formal checker: 438 native and 242 surface tests,
  zero failures/skips, distinct installed Python3.12 site-packages origins.
  Receipt `/private/tmp/spl-toolkit-root-m6-final-package-acceptance.json`,
  SHA256 `7794c826c713687e1d6b2221ebdad418ad4e3b4e91a2cea2aeccdfcbe0b9e069`.
  Direct wheel `ed37d91f749489f26cb2a4aa27876d75cae13564bdd604c78a9af9e331c4e01a`,
  sdist `ae71f137426b2489d02a1d7ca497ab23450bf55dec1cb9142356611bd7c5ca66`,
  rebuilt wheel `404c28a2c2fef5285d44940483c9d25f8156fb5b7f4fe7c60df53305f5502c79`.
  The older ignored worktree `dist/` archives are unrelated to this gate.

The final private surface oracle was reconciled to approved selector identity
`{name:"sourcetype"}` in three cases (its original `auth` identity was
inconsistent with the design), and the historical Task5 baseline comparison
allows only the reviewed additional `ref-2` evidence in three
`or_no_guaranteed_fact` cases. Both amendments and preceding failed attempts
are retained in the private attempt records; no production semantic failure
was waived. All validated input/source snapshots were unchanged by the runs.
The Go floor artifacts use external linking to supply macOS LC_UUID; the local
linker emits the retained LC_DYSYMTAB warning. Other OS/architectures, the
entire Python-version matrix, sanitizers/leak tooling and external Splunk
runtime equivalence were not executed. This is accepted M6 local evidence,
not a release-platform certification or a search-runtime guarantee.
