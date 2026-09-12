## What's new in the app

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
