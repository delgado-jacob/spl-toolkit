## What's new in the app

- Check nested fields against your local JSON Schema or an exact OCSF version.
- See which fields are required, optional, or depend on a category or schema branch.
- Keep uncertainty visible when schemas leave names open or cannot fully describe a field.
- Use the same results in the command line, Python, Go, and the local API, including ordered batches.

## What was built

The canonical query engine now accepts partial source universes without losing source bindings or claiming exhaustive wildcard expansion. JSON Schema projection uses supplied local resources, nested requiredness and bounded composition. Official compiled OCSF catalogs use explicit concrete class/category, profiles, exact version and full extension identity. All adapters expose the same schema reports while preserving legacy analysis and field-list contracts.

Task 7 adds full installed surface acceptance, an isolated REST selection-negative regression, maintained tutorials/reference docs, and permanent source/artifact evidence. Its 28 frozen reports and 22 malformed requests remain unchanged. Genuine catalogs are not reduced for HTTP. Required arrays, optional ancestors, generic descendants, branch/category evidence and profile-provenance uncertainty are covered without claiming event instance validation.

## Decisions and next-milestone constraints

- Go owns flow, scopes, locations, availability, binding, selector expansion and final reference IDs. Adapters serialize and delegate.
- Derived fields bypass external declarations; removals have no schema obligations. Missing schema membership cannot retroactively make an initial source read structurally unavailable.
- Known partial wildcard matches never prove an exhaustive source universe. Historical coverage gaps survive later locally complete output.
- Draft 2020-12 is the default. References resolve from explicit inline local resources only. No runtime compiler, account, URL/file retrieval, or global catalog registry exists.
- OCSF normal `compile_version: 1`, exact version and complete extension set are mandatory. Category members and profile provenance qualify results. Pinned catalogs/notices remain in testdata; development planning is never a runtime or package dependency.
- Projection bounds are 4096 states and 128 path segments; enumeration bound 4096. The precise supported ASCII regex subset and 25 stable reason identifiers are maintained in docs/API.md.
- Declared arrays are supported; descendants remain incomplete. Requiredness is a schema statement, not event presence. No event validation, expression typechecking, query execution, SPL2 grammar or rewriting was added.

## Deviations and reasons

The Go 1.22 floor commands proactively used external linking, avoiding the previously understood macOS LC_UUID default-link failure. API/socket acceptance requires authorized loopback execution; one sandbox-denied API floor scope was rerun alone, and subsequent Go/installed suites used that context.

The initial prescribed Go 1.22 Python build passed, but exact generated-header hashing rejected harmless cgo boilerplate differences from the committed Go 1.25 header. Root approved a narrow checker correction: validate unique ordered cgo boundaries; compare complete toolkit preamble/layout and complete exports; remove only compiler-owned prologues and #line metadata; normalize only the known empty-parameter spelling. Exact wrapper/source-header and actual wheel-header/native payload hashes remain recorded. Actual Go 1.22/1.25 headers are tooling-only fixtures; missing/new/wrong exports, struct changes and malformed/ambiguous boundaries are rejected. The original floor wheel was retained and passed without a rebuild.

New source schema acceptance uses an explicit native-library path; pre-existing acceptance suites continue to run against installed packages. Tutorial snapshots were authored and independently checked once before tests; tests never regenerate expected results. Existing help expectations were updated only for accepted schema command additions. Swagger artifacts were already correct and were verified without regeneration.

## Verified commits and evidence

Accepted M3 base: `e401402e5d8cd80f469884b2c3ad80245266ad66`.

- Task 1: `8e94df25110af708b94f0d83a724ed07599b1174`, correction `47b560d710ed264490e6c21711ad157146b8fa78`.
- Task 2: `94714a64c5674d7ce10fa428479c57f28e9050f3`, corrections `1a59404244edb727293db95252b2a7e665701acc`, `397efc32c9ea2f9ff86f1b977e179d2944cfc481`, `bacc2172291005b0c70cdbba948512173090f9aa`.
- Task 3: `de09f38f2f086751dc7f5c1e879d5b3d68fa77e5`, correction `4675faf779d8427732d54cb2214aa8305021bcc2`.
- Task 4: `5e3d4ec3a54b6878830b83a8ff08405011a507d9`.
- Task 5: `c91ca8464e916ae0b3792679a195a76e936c9385`.
- Task 6 / Task 7 input base: `7b6580fff46917045230668141d23f17489891df`.

Permanent evidence: `docs/evidence/milestone-4/verification.json`, `source-identity.json`, `source-surfaces.json`, `package.json`, and command logs/input manifests in its `logs/` directory. These record actual base SHA plus input hashes; later human documentation is distinguished from executable inputs. All 64 Go build inputs match the tested build. The final Task 7 commit SHA and clean-owned proof are recorded in its controller report after the exact-path commit.

| Check | Actual result |
|---|---|
| Go 1.22.12 floor, six required packages, external link | Passed; only sandbox-blocked API scope repeated |
| make build-all, Go 1.25.5 | Passed |
| Full Go formatting/vet/race gate | Passed, 7 test-bearing packages |
| Python 3.12.6 source suite | 230 passed, 0 skipped |
| Source schema + documented CLI acceptance | 57 passed, 0 skipped |
| Final tooling | 114 passed |
| Documentation front matter | 13 pages passed |
| Offline Go 1.22 wheel/sdist build | Passed |
| Installed floor wheel | 213 native + 81 surface,0 failed/skipped |
| Installed wheel rebuilt from sdist | 213 native + 81 surface,0 failed/skipped |
| Each schema surface evidence | 28 full reports, 19 ordered groups, 22 exact raw malformed requests, real file_activity/xattributes |

Runtime tests use unusable external proxies and allow loopback only for the real HTTP server. Unresolved HTTPS references return promptly without retrieval. Code inspection independently confirms no schema preparation HTTP/file opener. Installed fixture/library paths are outside the checkout, package imports belong to their venv, and all 11 schema fixture hashes and 51 sdist source hashes are retained. No dependency preflight or wheelhouse regeneration was repeated.

## Remaining gates and limits

Implementation and local checks are ready for independent Task 7 review. That review, the separate broad milestone review, root stable-build oracle and final controller acceptance remain open. No milestone completion, merge, push, release or cleanup is claimed here.

Execution was macOS 26.2 arm64, Go 1.25.5 plus Go 1.22.12 floor, Python 3.12.6 and Apple clang 17. The existing LC_DYSYMTAB race-link warning remains qualified. Go 1.22 default linking, Linux/Windows release jobs, other Python versions and external Splunk conformance were not established by these local runs. The official compiler's preparation-only Python 3.14.1 requirement does not change toolkit runtime floors.

## Task 7 review correction I1

Review of Task 7 commit `db7c05ea9f6059e3fba4e0ddc37ff1ebf9f014f7` found that the initial header checker could ignore unexpected declarations or preprocessor directives in compiler sections. This supersedes the earlier compiler-prologue exclusion described above: the corrected checker compares the complete header, normalizing only the exact known Go 1.22/1.25 guarded string-helper and MSVC complex blocks, valid whole-line `#line` metadata, and the two known empty-parameter prototypes. All other bytes remain in the comparison, with unique ordered cgo boundaries still required.

Eight new mutation regressions failed as expected before the fix. After the fix, all 70 package tests and the schema reason/documentation test passed (71 total, zero failures/skips), retaining both actual header acceptance cases and the previous 12 mutations. Permanent supplemental commands, exact transcripts, input manifests, checker/test hashes and direct artifact results are in `docs/evidence/milestone-4/review-i1/`, indexed by `verification.json`.

Fresh corrected `verify_wheel_sources` passed the retained original Go 1.22 wheel with its unchanged archive and payload hashes. Both unchanged actual Go 1.22/1.25 header fixtures pass complete-header comparison. The rebuilt-sdist wheel archive was removed by the original checker's temporary-directory cleanup; root approved revalidating its retained Go 1.25 header by the exact previously recorded payload hash `b800c47a3f7b1de0dfd0b78c45eb767c4ac54c4016441d6ca1a431d6e639fda1`. No fresh rebuilt full-archive verification is claimed. Original `package.json`, both historical installed-suite results and archive/native identities remain unchanged. No build or unchanged Go/native/installed suite was repeated. Scoped Task 7 rereview and the remaining milestone/root/controller gates are open.
