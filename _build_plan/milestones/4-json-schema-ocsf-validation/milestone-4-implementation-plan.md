# JSON Schema and Offline OCSF Validation Implementation Plan

Status: all seven tasks, scoped fixes, broad review and root final independent acceptance passed at `92879052b0454db12aead045188ce4b15fa9d83a` on 2026-09-08. The remaining text preserves the approved execution instructions and historical gates; documentation-only closure introduces no new runtime or test behavior.


> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development to implement this plan task by task. Track steps with checkboxes in Progress; the narrative task steps below define their test and review gates. Root has selected serial production implementers. This file is a planning artifact, not production authorization.

**Goal:** Let a user check SPL query field references against a local JSON Schema or an explicitly selected official OCSF catalog, with the same explainable single and batch results in Go, CLI, REST, and installed native Python.

**Architecture:** Extend the existing canonical analysis engine with an additive partial source-universe hook. Prepare local schema targets once, resolve individual source paths without replaying query transfers, and build schema-specific reports alongside the unchanged M3 field-list reports. CLI, REST, and native Python decode inputs and return those Go reports.

**Tech Stack:** Go 1.22+, existing ANTLR parser and standard-library JSON/URI/regexp/gzip packages; Python 3.11+ ctypes bindings; existing HTTP server, CLI, packaging, and pytest acceptance harness. Official catalog preparation uses Python 3.14.1 only outside the runtime.

**Spec:** `_build_plan/milestones/4-json-schema-ocsf-validation/design-spec.md`; provenance and primary-source observations are in the adjacent `research-notes.md`. This ExecPlan is maintained under `/Users/jacobdelgado/.codex/PLANS.md` and repeats the operative behavior so it can be executed without conversation history.

## Purpose and historical production gate


After this work, `spl-toolkit validate-schema --schema event.schema.json --query 'table actor.name' --format json` explains whether `actor.name` is required, optional, admitted without a declaration, conditional, missing, or indeterminate. An OCSF user can supply a normal official compiled catalog, choose version `1.6.0` and class `authentication`, and see field evidence from that exact local catalog. Selecting the IAM category explains which classes support a conditional field. No event data or remote schema service is needed.

Historical gate record (satisfied by the production release recorded below): planning started against independently accepted M3 core commit `aeb71a920f1ff1ee5c022077898e5c078241cfb5`, following accepted M2 `1f5c98605aba26cab402fa46b6605f9e650cf643`. The finite hook was introduced at `3cb50d1`; its validation consumer is at `aeb71a9`. The subsequent read-only interface preflight inspected completed M3 adapters at `c6ae68441abee859ee7e4e5b4addf43d202096ab`; broad milestone acceptance was still pending at that inspection. Before Task 1 production work, root must record the final accepted M3 SHA and explicitly release M4. Reconcile every adapter/helper/manifest path below against that SHA and incorporate any accepted core corrections; do not reset the worktree or overwrite newer M3 behavior. Record the reconciliation and actual baseline SHA here. Until that gate is satisfied, only this plan, M4 spec, and M4 research notes may change. No fixture copying, code, tests, builds, commits, or implementer dispatch is authorized by this planning release.

## Global Constraints


Runtime is offline, with no schema/OCSF compiler, downloading, account, proprietary registry, URL dereference, or implicit filesystem lookup. `_build_plan/` is never a runtime, test-fixture, or package dependency.

Preserve Go 1.22+ and Python 3.11+ floors, plain `analysis.Analyze`, existing native ownership, legacy mapping/syntax APIs, and every existing M3 field-list wire shape. No grammar expansion, event validation, expression typechecking, query execution, selection from query values, or rewriting work belongs here.

Keep the canonical engine responsible for field transfer, scopes, source locations, source/derived binding, structural unavailability, and final reference IDs. External schema absence cannot make an initial exact source read structurally unavailable. Removal/exclusion roles do not consume schema obligations. Derived fields bypass external schema membership.

Do not claim an open, patterned, recursive, or provenance-limited source universe is finite and exhaustive. Preserve historical incompleteness and unknown-stage binding effects even when a later stage establishes a locally complete output. Known partial matches never imply exhaustive expansion.

JSON Schema defaults to Draft 2020-12; explicit unsupported dialects are input errors. Unsupported field-affecting semantics are located incomplete results, not silently ignored keywords. OCSF requires normal `compile_version: 1`, exact version, explicit concrete class/category, profile selection, and full compiled extension-set equality.

Use the worktree `/Users/jacobdelgado/repos/spl-toolkit/.worktrees/remaining-milestones`, branch `codex/remaining-milestones`. Preserve other agents' files and changes. Each serial task follows implement/test/self-review, exact-path commit, independent combined spec/code review, then tested committed fixes and scoped rereview. Root controls task release and integration. No broad `git add .`, reset, cleanup, or parallel production worker.

## Progress


- [x] (2026-09-07) Read the approved behavior, accepted M3 core, and upstream preparation evidence; rechecked retained base/Windows catalog hashes.
- [x] (2026-09-07) Root approved the additive partial-universe contract direction, schema-specific report/request types, and an 8 MiB limit only for new schema REST routes.
- [x] (2026-09-07) Root approved the corrected written plan at SHA-256 `3a98ebf4e977f0ee1d50e48110b57077419d37e90f09841ab783e9540d312257`; this read-only interface reconciliation preserves its contracts.
- [x] (2026-09-07) Root accepted M3 at753bd5a and released M4 on final documentation handoff `e401402e5d8cd80f469884b2c3ad80245266ad66`; exact successor changes are only three M3 guidance files. Reconciled final symbols with unchanged production inputs and initialized the unique SDD workspace.
- [x] (2026-09-07) Task 1: Canonical partial-universe hook accepted through47b560d710ed264490e6c21711ad157146b8fa78; tracked internal admission defect reproduced/fixed; combined review plus scoped rereview clean; focused/race gates passed.
- [x] (2026-09-07) Task 2: Strict local target/JSON Schema projection accepted through `bacc2172291005b0c70cdbba948512173090f9aa`; combined review plus two committed scoped fix rounds closed all five findings.
- [x] (2026-09-07) Task 3: Official OCSF preparation/projection and pinned gzip/provenance fixtures accepted through `4675faf779d8427732d54cb2214aa8305021bcc2`; combined review plus committed singleton-depth fix/rereview clean.
- [x] (2026-09-07) Task 4: Canonical schema report/request/batch APIs accepted at `5e3d4ec3a54b6878830b83a8ff08405011a507d9`; combined review clean,148 top-level race tests passed with28 schema snapshots/22 malformed cases/26 M3 byte comparisons. Root17/4/13 Go-only oracle also passed with clean-source guards.
- [x] (2026-09-07) Task 5: CLI/REST/OpenAPI accepted at `c91ca8464e916ae0b3792679a195a76e936c9385`; required affected suites, handler race, real raw/compact catalog limits, offline generation and reconciler passed. The nonblocking negative-test isolation observation was closed in Task7 and the broad review.
- [x] (2026-09-07) Task 6: Native/Python/package closure accepted at `7b6580fff46917045230668141d23f17489891df`;213 native tests/98 tooling tests passed,47 source/11 fixture closure and generated-header identity verified; review clean.
- [x] (2026-09-08) Task 7: Final surfaces/installed artifacts/docs and independent task review accepted through `92879052b0454db12aead045188ce4b15fa9d83a`; one header-checker finding fixed and scoped rereview clean.
- [x] (2026-09-08) Broad review approved all147 changed files with zero actionable findings; root final17 semantic/4 batch/13 malformed Go/all-surface checks and230 Python3.14 tests passed at9287905.
- [x] (2026-09-08) Documentation-only closure records exact accepted identities, initial guidance hashes and limits; final closure commit identity is reported to root separately to avoid self-referential hashes.

## Surprises & Discoveries


The accepted finite hook treats its supplied names as complete and drops every match from incomplete expansions. Merely adding a `Complete` boolean would still lose an exact name admitted by an open schema after `table custom | table cus*`. Conversely, unioning syntactic property names would admit a name prohibited by another `allOf` conjunct. The new hook therefore needs both partial enumeration and query-independent exact admission.

Official compiler `d6b0b781d51a6b9682ea99396636ae01437b41b4` successfully compiled official schema release `v1.6.0`, commit `d0cd8a0fef198bf93086044e33fda6d80c74b9ef`, using Python 3.14.1. Base output is 3,122,505 bytes; selected-Windows output is 3,344,790 bytes. Both exceed the existing 1 MiB REST ceiling. The new routes need a catalog-capable bound and tests using these actual bytes.

Normal catalogs retain `base_event` with UID/category UID 0 and category `other`, absent from the ordinary category table. Exclude this recognized sentinel before concrete-category link validation. Other malformed links remain input errors.

The generic `object` definition explicitly permits unmodeled content: compiled `unmapped` and `file.xattributes` resolve to `object_t` with `object_type: "object"`. This exact base key is open alongside `json_t`. Empty attributes and subclasses merely extending `object` are not evidence of openness.

Compiler code merges profile names and strongest requirements, but does not prove profile-inheritance expansion or requiredness for every selected strict subset of merged profile names. Real base/Windows catalogs contain ordinary profiles without these edge cases. Use genuine catalogs for normal acceptance and clearly labeled synthetic fixtures for the metadata-loss cases.

## Decision Log


2026-09-07, root/design lead: Extend the canonical engine with `AnalyzeWithSourceUniverse`, preserving `AnalyzeWithSourceFields` semantics and its empty-on-incomplete sidecar. A pure resolver answers external name admission only. This handles open exact reads without a second flow engine.

2026-09-07, root/design lead: Keep M3 `Target`, `Report`, `Request`, `BatchRequest`, and field-list outcomes unchanged. Add schema-specific types and reuse diagnostic finalization at the smallest private helper boundary. Compact target metadata avoids embedding catalogs in each report.

2026-09-07, root/design lead: New schema REST routes have an 8 MiB request-body bound; old routes retain 1 MiB. Match existing HTTP input-error behavior for oversized requests: return 400 with an explicit body-limit message. Tests cover the genuine catalogs, the exact bound, and requests exceeding it, including unknown content length.

2026-09-07, design lead: Use a deliberately small portable ASCII property-pattern subset and conservative indeterminacy outside it; do not claim Go RE2 implements ECMAScript. Keep unsupported traversal/regex/keyword limits visible in compact target capability metadata and permanent documentation.

2026-09-07, design lead: Vendor deterministic gzip copies of the two verified raw official catalogs as test inputs, preserve raw-byte hashes and both upstream notices, and decompress with standard libraries in tests. The production CLI still takes ordinary local JSON. No runtime compiler or gzip-input feature is introduced.

## Outcomes & Retrospective


M4 is implemented and independently accepted at `92879052b0454db12aead045188ce4b15fa9d83a`, descending from the final accepted M3 handoff `e401402e5d8cd80f469884b2c3ad80245266ad66`. The seven serial task reviews and all committed scoped fixes passed. The broad review accounted for all147 changed files and found no actionable issue. The partial-universe hook, local JSON Schema projection, official OCSF selection and all schema adapters preserve canonical field flow and the M3 compatibility boundary.

Task review regressions closed retained-internal admission, effective literal-path/URI/type/conditional handling, strongest singleton required depth, and complete generated-header verification. The Task5 negative-test isolation observation was closed using a real valid catalog baseline. The final checker compares complete headers after only exact known compiler normalizations; historical installed runs and supplemental corrected-checker proof are distinguished.

Verification includes the full Go checker and Go1.22 external-link floor,230 source Python3.12 tests,57 source schema/documented CLI tests,114 pre-fix tooling tests and71 affected checker/docs tests. Both installed wheel and rebuilt-sdist runs passed213 native plus81 surface tests with no failures/skips; full28 report/19 batch/22 malformed-byte schema corpus evidence is retained. Root separately passed17 semantic cases,4 batch documents and13 malformed requests through Go/CLI/REST/native, including nine native calls per case, and230 tests on Python3.14. Root verified290 tracked identities,47 native manifest entries and5 artifacts; the broad reviewer independently checked provenance, evidence and cross-task closure.

The only final successor is documentation. Initial guidance identities and exact acceptance/review hashes are recorded in `docs/evidence/milestone-4/closure.json`; the milestone log retains the commit chain, local platform limitations and rebuilt-header-only supplemental qualification. No default-link Go1.22, Linux/Windows release matrix or live Splunk conformance is claimed. No merge, push, release or cleanup is performed.

## Context and Orientation


`pkg/analysis/analyze.go` parses once and finalizes the canonical result. `flow.go`, `references.go`, and `commands.go` maintain each pipeline's field environment. A tracked field records whether its binding originated at a source or was derived and whether it is conditional. `source_fields.go` already refines selectors using a complete list of source names and adds a `SourceAnalysis` sidecar with `FieldExpansion` records; its `Result` has the ordinary analysis wire shape. `selectorStructurallyAbsent` distinguishes source-catalog absence from a pipeline projection/removal. `finalizeExpansions` preserves finalized reference IDs. Plain analysis must not acquire external schema knowledge.

`pkg/validation/model.go` contains M3 `FieldCatalog`, `Target`, `Coverage`, `Match`, `ReferenceOutcome`, `Report`, `BatchReport`, `Request`, and `BatchRequest`. `catalog.go` normalizes copied finite names. `request.go` has `InputError`, `IsInputError`, strict object/document decoding, and `DecodeDocuments`; `internal/jsoninput` rejects malformed UTF-8 and unpaired surrogate escapes before JSON decoding. `validate.go` exposes `Validate`, translates source obligations, and has private diagnostic sorting/status finalization. `batch.go` exposes `ValidateBatch`, normalizes once, preserves order, and rejects an empty batch. A schema projection is a query-independent answer for one external field path: admission, requirement, and source evidence. It is not an event validator.

CLI dispatch/options live in `cmd/cli.go`; current M3 adapter work uses `cmd/validation_cli.go`. HTTP registration lives in `pkg/api/server.go`, with current M3 handlers in `pkg/api/validation.go`. Native exports are in `pkg/bindings/bindings.go`; Python ctypes calls and `_operation` lifetime coordination are in `python/spl_toolkit/mapper.py`. Public C declarations live in `python/spl_toolkit/libspl_toolkit.h`. The actual Go documentation is `docs/API.md`. The completed M3 helper and packaging details are recorded in the final-interface preflight section below. `python/native-source-files.txt` explicitly enumerates source-package Go inputs. `tools/check_package.py` checks exact sdist closure, copies installed acceptance inputs out of the checkout, and rejects missing/skipped required tests. New sources and fixtures must enter these checks deliberately.

## Final-interface preflight (read-only, production gate remains open)


On 2026-09-07, inspected immutable current HEAD `c6ae68441abee859ee7e4e5b4addf43d202096ab`. M3 production/installed/root acceptance had passed through `78964a7775bb2e74d2022bb760d17be4e6e27e3a`; the newer commit changes only evidence/docs closure and was entering broad review. This preflight found no approved design-contract discrepancy. Reconcile any subsequent M3 review fixes at explicit production release; this note is not final M3 acceptance.

The core symbols and wire shapes used by this plan are unchanged. Preserve `TestSourceUniverseUnknownMaterializationStaysConditional` in `pkg/analysis/source_fields_test.go`, which exercises `| mystery | table host* | table host | table host*` and prevents catalog membership from promoting unknown-stage source binding. The existing child-before-parent expansion-ID regression also remains required.

CLI `cmd/validation_cli.go` supplies `runValidationCLI`, `validateFieldCLIOptions`, `computeValidationCLIResult`, and `formatValidationText`. Shared parsing in `cmd/cli.go` has command-specific gates in `parseCLIOptions` and `analysisOptionMayBeEmpty`, with `setCLIOption` preserving explicit flag presence. Task 5 extends those gates for the new operation and uses existing `runCLIWithInput`, `marshalCLILine`, output/error helpers, and status exits. REST `readValidationBody` is deliberately still a 1 MiB reader; Task 5 adds a separate 8 MiB schema reader and can reuse `writeValidationError`. `NewServer().Handler().ServeHTTP` is the exact test setup. OpenAPI generation remains pinned Swag plus `tools/update_validation_openapi.py` and its same-named tooling test.

Native `ownedMapperJSONResult` in `pkg/bindings/bindings.go` already owns result allocation, handle admission, `runtime.KeepAlive`, error strings and JSON serialization. Task 6 supplies closures invoking the new canonical decoders/validators. Python `_validate_fields_request(self, native, request)` already owns JSON serialization with allow_nan false, `_operation`, native error decoding and finally/free; use it for schema wrappers unchanged. The cgo-generated header uses `char* requestJSON`, not a handwritten const variant. Source native tests may pass `library_path` from SPL_NATIVE_LIBRARY; installed acceptance must instantiate `SPLMapper()` and load the packaged library without such an override.

The 37-line `python/native-source-files.txt` has SHA-256 `b44010e742955c5622730c0ac89c85d4dccbd83db352f55ca08956f43f4e4bdd`; it already contains source_fields.go and all five M3 validation sources. Append new handwritten Go sources without losing those entries. `python/MANIFEST.in` already recursively includes tests/*.py, so no manifest edit is needed merely to include the new Python test. `tools/check_package.py` additionally requires exact `SDIST_FIXED_FILES`, `NATIVE_TESTS`, `ACCEPTANCE_FILES`, `_copy_required_files`, `_run_required_suite` nonzero/zero-skip gates, and fixture copies/hashes in `install_and_check`. Schema fixtures must be copied before the native suite and its environment must receive the same absolute tree as the surface suite. Current M3 test fixtures are copied from the checkout into isolated acceptance inputs, not bundled as runtime catalogs. `python/build_support.py:ROOT_FILES` copies the repository README into the sdist; keep the package checker consistent with that source. Task 6 tooling changes belong to `tools/tests/test_package.py`.

Task 7 uses installed `tests/acceptance/test_validation_surfaces.py` as the concrete fixture/import pattern. Its `SPLMapper()` is intentionally installed-only; running that suite with source PYTHONPATH plus SPL_NATIVE_LIBRARY is not valid installed proof. Preserve `clean_env` removal of PYTHONPATH/PYTHONHOME/SPL_NATIVE_LIBRARY/SPL_EXPECTED_VERSION and the checker's Python-I isolation. Use the existing Go1.22 PATH/PIP_NO_INDEX package-build recipe and `--evidence` option in the final command block. `tools/check_go.py` contains the full race suite; do not repeat it separately. Documentation inputs are docs/API.md, docs/quickstart.md, docs/architecture.md, docs/compatibility.md, repository README, Python README and the pinned OpenAPI files; include the new schema quickstart in Task 7 as well.

Root has independently prepared `/private/tmp/spl-toolkit-root-m4-cases.json` (reported 14 grammar-checked query texts; SHA-256 `61558b53b6e6f0ad6f9684fc22ea0f4b9d6d19cd73ffb7cee16f0f4e06677f35`) and `/private/tmp/spl-toolkit-root-m4-invalid-inputs/manifest.json` (reported 12 cases; SHA-256 `f88d2af181608770f1db58db98247dbcfbd2e316b7f74f3b11d8574ab3ace539`). Only their bytes were hashed here; schema behavior was not executed and these remain the controller's later independent oracle, not fixtures used to manufacture implementation expectations.

## Interfaces and shared rules


Task 1 adds these definitions in `pkg/analysis/source_fields.go`; no JSON representation is needed for the callback-bearing input:

    type SourceFieldAdmission uint8
    const (
        SourceFieldIndeterminate SourceFieldAdmission = iota
        SourceFieldAdmitted
        SourceFieldProhibited
    )
    type SourceUniverse struct {
        Fields []string
        Complete bool
        Resolve func(name string) SourceFieldAdmission
    }
    func AnalyzeWithSourceUniverse(document QueryDocument, universe SourceUniverse) (*SourceAnalysis, error)

`Fields` is a copied, sorted set of candidate concrete names; reject duplicate, invalid UTF-8, empty, or all-whitespace names using the existing finite-hook rules. Preserve case and spelling. `Complete` promises that every potentially admitted external source name is listed. Under `Complete: true`, an unlisted name is prohibited without invoking the resolver; a callback cannot silently expand a purportedly complete universe. Listed candidates may resolve prohibited (exclude them) or indeterminate (prevent conclusive expansion when relevant). With a nil resolver, listed names are admitted, unlisted names are prohibited when complete and indeterminate when partial. A nonnil resolver on a partial universe can admit an unlisted exact tracked source name. Unknown enum values normalize to indeterminate. The callback must be pure, deterministic, concurrency-safe, and independent of query text, pipeline state, and event presence; it may be called repeatedly and callers must not mutate captured target data during calls. The engine copies the names, not arbitrary callback state, and does not recover callback panics as fake schema results.

`AnalyzeWithSourceFields` retains known-finite nil/empty semantics and its existing wire behavior. Internally it can delegate through the new machinery with complete names and a private flag to discard incomplete matches. `Analyze` continues passing no refinement. For the new hook, an incomplete expansion retains only proven source/derived members, never conditional members. A complete expansion requires exhaustive local membership, conclusive admission of relevant candidates, and proven bindings. Complete output from a known `stats` or exact `table` can restore local precision while historical coverage remains incomplete. Exact projection of a name after an unknown stage cannot manufacture certain source provenance; conditional fields remain conditional even when a subsequent environment is closed.

Task 2 defines the public tagged request union in new `pkg/validation/schema_model.go`. Use these exact Go shapes with snake_case JSON tags; fields shown with `omitempty` are optional JSON members:

    type OCSFSelection struct {
        Version string `json:"version"`
        Class string `json:"class,omitempty"`
        ClassUID *int64 `json:"class_uid,omitempty"`
        Category string `json:"category,omitempty"`
        CategoryUID *int64 `json:"category_uid,omitempty"`
        Profiles []string `json:"profiles"`
        Extensions []string `json:"extensions"`
    }
    type SchemaTarget struct {
        Kind string `json:"kind"`
        Identity string `json:"identity,omitempty"`
        Schema json.RawMessage `json:"schema,omitempty"`
        BaseURI string `json:"base_uri,omitempty"`
        Resources map[string]json.RawMessage `json:"resources,omitempty"`
        Catalog json.RawMessage `json:"catalog,omitempty"`
        Selection *OCSFSelection `json:"selection,omitempty"`
    }
    func DecodeSchemaTarget(data []byte) (SchemaTarget, error)

`json_schema` requires `schema` and permits identity/base_uri/resources only. `ocsf` requires `catalog` and `selection` and permits identity only in addition. Reject mixed target members even when empty/null. Selection requires nonempty exact `version` and exactly one of class/class_uid/category/category_uid. Integer IDs must fit int64, with no float coercion. Omitted profile/extension arrays normalize to copied sorted empty arrays; explicitly null arrays are invalid in the request wrapper. Reject duplicate names, blank or invalid UTF-8 strings, unknown members, and duplicate JSON keys. Schema/catalog internals allow descriptive and annotation members but reject duplicate keys throughout to avoid parser-dependent meaning. Typed Go inputs receive equivalent semantic validation and defensive copies during preparation; invalid typed combinations are `InputError`s too.

Task 4 completes the public operation signatures:

    type SchemaRequest struct {
        Document analysis.QueryDocument `json:"document"`
        Target SchemaTarget `json:"target"`
    }
    type SchemaBatchRequest struct {
        Documents []analysis.QueryDocument `json:"documents"`
        Target SchemaTarget `json:"target"`
    }
    func DecodeSchemaRequest(data []byte) (SchemaRequest, error)
    func DecodeSchemaBatchRequest(data []byte) (SchemaBatchRequest, error)
    func ValidateSchema(document analysis.QueryDocument, target SchemaTarget) (*SchemaReport, error)
    func ValidateSchemaBatch(documents []analysis.QueryDocument, target SchemaTarget) (*SchemaBatchReport, error)

The report types live in `schema_model.go`. Retain M3 `Coverage`, `analysis.Result`, and `analysis.Diagnostic` rather than cloning their fields. `SchemaReport` has `SchemaVersion int`, `Target SchemaTargetInfo`, `Analysis *analysis.Result`, `Status analysis.Status`, `Coverage Coverage`, `Outcomes []SchemaReferenceOutcome`, and `Diagnostics []analysis.Diagnostic`, with the existing snake_case names and schema_version 1. `SchemaBatchReport` has schema_version, status, and reports (`[]*SchemaReport`). All report/evidence arrays are nonnull, stably ordered, and freshly owned.

`SchemaTargetInfo` contains string fields Kind/Identity/Version/Dialect/BaseURI, `ResourceURIs []string`, `CompileVersion int`, `Selection *OCSFSelection`, `Members []SchemaClass`, `Extensions []SchemaExtension`, and `Limitations []string`, with JSON names kind, identity, version, dialect, base_uri, resource_uris, compile_version, selection, members, extensions, and limitations. Use optional scalar/selection members where inapplicable, never raw schema/catalog. `selection` is the normalized `*OCSFSelection`; `members` is `[]SchemaClass{Key string, UID int64}`; `extensions` is `[]SchemaExtension{Key string, UID int64, Version string}`. Their tags are key/uid/version. `resource_uris` lists sorted supplied/indexed external resource identities; use `urn:spl-toolkit:json-schema:root` as the deterministic root identity when no base or root `$id` supplies one. `limitations` is a sorted `[]string` of stable capability identifiers, including `array_traversal`, `partial_name_universe`, and target-present unsupported constructs. It is descriptive; presence alone does not make an unrelated exact obligation incomplete. Reuse a private capability-reason registry for documentation tests; do not put schema runtime logic into `analysis.Capabilities`.

`SchemaEvidence` has string fields ResourceURI, Pointer, Keyword, Operator, Branch, ClassKey, DeclarationBasis, Requirement, Reason with JSON names resource_uri, pointer, keyword, operator, branch, class_key, declaration_basis, requirement, reason; `ClassUID *int64` is optional. Empty inapplicable strings may be omitted. `declaration_basis` is one of property/pattern/additional_properties/ocsf_attribute/generic_object/derived/unknown. `requirement` is required/optional/recommended/unknown, preserving local ancestor records even when the full path is optional. A branch is a stable schema pointer; an operator is allOf/anyOf/oneOf or the relevant constraint form. Reasons use stable identifiers and explanatory diagnostic messages, not raw exception strings.

`SchemaMatch` has `Name string`, `Binding string`, `Outcome string`, and `Evidence []SchemaEvidence`, with JSON names name, binding, outcome, and evidence. `SchemaReferenceOutcome` has `ReferenceID string`, `Outcome string`, `MatchesComplete bool`, `Matches []SchemaMatch`, `Evidence []SchemaEvidence`, and `SupportingClasses`, `MissingClasses`, `IndeterminateClasses` each `[]SchemaClass`, with the corresponding snake_case JSON names reference_id, outcome, matches_complete, matches, evidence, supporting_classes, missing_classes, and indeterminate_classes. Exact outcomes set matches_complete true only for conclusive membership, emit one match for conclusive admitted source/derived paths, and no fake successful match for missing or indeterminate membership. An admitted source path with indeterminate requiredness retains a match whose outcome is indeterminate and matches_complete is true, while schema_complete is false; match completeness describes membership, not established requiredness. Wildcard partial matches set matches_complete false. Categories preserve negative and positive class evidence even when no successful match is emitted.

Internally define `fieldProjection` with `Admission analysis.SourceFieldAdmission`, `Outcome string`, `Evidence []SchemaEvidence`, and the three class lists. Define `preparedSchemaTarget` with methods `project(name string) fieldProjection`, `universe() analysis.SourceUniverse`, and `info() SchemaTargetInfo`. `prepareSchemaTarget(target SchemaTarget) (preparedSchemaTarget, error)` selects JSON or OCSF preparation. Admission is admitted only when field membership is conclusive; conditional, unsupported, or unresolved membership is indeterminate. Requiredness alone may be unknown even with admitted membership. `universe()` must include every potentially admitted name when complete and must exclude conclusively prohibited names. Resolver calls `project` without consulting query state. Use immutable prepared indexes and call-local recursion state; do not share mutable memo maps between calls.

## Task 1: Preserve partial source evidence in the canonical engine


Own `pkg/analysis/source_fields.go`, focused transfer changes in `pkg/analysis/commands.go` and, only if necessary, `flow.go`/`references.go`; add `pkg/analysis/source_universe_test.go`. Consume the accepted finite hook/environment and produce the additive API above. Do not edit validation transfers.

1. Add a failing public-API test in package `analysis_test`:

       func TestPartialSourceExactProjectionRecoversLocalExpansion(t *testing.T) {
           doc := analysis.QueryDocument{Text: "table custom | table cus*"}
           got, err := analysis.AnalyzeWithSourceUniverse(doc, analysis.SourceUniverse{
               Complete: false,
               Resolve: func(string) analysis.SourceFieldAdmission { return analysis.SourceFieldAdmitted },
           })
           if err != nil { t.Fatal(err) }
           if got.Result.Status != analysis.Valid || len(got.Expansions) != 1 {
               t.Fatalf("result=%+v expansions=%+v", got.Result, got.Expansions)
           }
           want := []analysis.ExpandedField{{Name: "custom", Binding: "source"}}
           if !got.Expansions[0].Complete || !reflect.DeepEqual(got.Expansions[0].Matches, want) {
               t.Fatalf("expansion=%+v", got.Expansions[0])
           }
       }

   Run `go test -mod=readonly ./pkg/analysis -run TestPartialSourceExactProjectionRecoversLocalExpansion -count=1`; initially it fails because the API does not exist. Add independent cases for complete unlisted callback admission being ignored, nil resolver rules, listed unknown/prohibited admission, invalid enum, duplicate/Unicode/copy validation, and exact initial prohibited source binding remaining source. Use table-driven inputs with expected final binding, expansion completeness, partial matches, and status; do not assert only that a report exists.

2. Implement admission lookup once in the refinement, preserving a separate finite compatibility flag. The ordering is essential:

       if !listed && universe.Complete { return SourceFieldProhibited }
       if universe.Resolve != nil {
           result := universe.Resolve(name)
           if result == SourceFieldAdmitted || result == SourceFieldProhibited { return result }
           return SourceFieldIndeterminate
       }
       if listed { return SourceFieldAdmitted }
       return SourceFieldIndeterminate

   Adapt `refinedSelectorCandidates` to consult this lookup for source fields, bypass it for proven derived bindings, exclude prohibited candidates, and retain uncertain candidates internally without presenting them as proven matches. For a closed environment, exhaustiveness comes from the actual local environment, including unresolved tracked obligations; for an open environment it also requires the provider's completeness. Capture proven partial matches before adding an unresolved-selector diagnostic. Do not clear unknown/conditional provenance when materializing candidates.

3. Update `selectorStructurallyAbsent` so enumerated removals prove an open-universe absence only when enumeration and admission are conclusive. A closed environment's raw tracked obligations still prevent source-catalog filtering from fabricating structural absence. Partial wildcard inclusion/removal must preserve unknown transfer effects for later reads; it may retain proven names, but cannot turn the selected finite subset into a certain complete output. `fields` retaining unknown internal names must emit incompleteness for a partial source universe; never inject `_time` defaults. Exact `table` can close membership only for names actually proven at that stage. Keep stats' explicit new outputs and scope handling canonical.

4. Add regressions for `table host* | table hostname`, `fields -host* | table hostname`, `fields host | table _*`, `| mystery | table host | table host*`, `| mystery | stats count AS total | table total*`, partial `table *` with one proven derived name and an unknown source remainder, and `table prohibited | table pro*` with prohibited resolver. Check historical coverage, removal roles, full reference IDs, original spans, and per-match source/derived evidence. Repeat the new hook concurrently with immutable resolver data. Run new tests plus all accepted `source_fields_test.go` cases; existing finite output stays byte-equal, including empty incomplete matches. Run `go test -mod=readonly -race ./pkg/analysis` after green focused tests.

5. Self-review the exact diff, stage only the files actually changed from this task, and commit `feat(analysis): support partial source universes`. Root dispatches a fresh combined spec/code reviewer on that commit. Test and commit any fixes before scoped rereview; do not begin Task 2 while an actionable finding remains.

## Task 2: Prepare strict local JSON Schema targets and project paths


Own new `pkg/validation/schema_model.go`, `schema_target.go`, `json_schema.go`, `json_schema_resources.go`, `json_schema_patterns.go`, and focused tests `schema_target_test.go`, `json_schema_test.go`, `json_schema_resources_test.go`, `json_schema_patterns_test.go`. `schema_model.go` initially defines the target, metadata/evidence, and private projection types needed here; Task 4 adds public reports. Reuse `request.go` input-error/Unicode helpers through small private additions rather than a second inconsistent JSON decoder. Consume Task 1's universe API and produce `DecodeSchemaTarget`, `prepareJSONSchema`, and a `preparedSchemaTarget` implementation. The `ocsf` dispatch is wired only when Task 3 provides its implementation; Task 2 tests invoke the direct JSON preparer; Task 3 adds the complete dispatcher before public validation operations ship.

1. Write a failing internal test proving conjunction closure:

       func TestJSONSchemaAllOfClosureExcludesSyntacticCandidate(t *testing.T) {
           target, err := DecodeSchemaTarget([]byte(`{"kind":"json_schema","schema":{"allOf":[{"type":"object","properties":{"a":true},"additionalProperties":false},{"properties":{"b":true}}]}}`))
           if err != nil { t.Fatal(err) }
           prepared, err := prepareJSONSchema(target)
           if err != nil { t.Fatal(err) }
           if got := prepared.project("b"); got.Outcome != "missing" || got.Admission != analysis.SourceFieldProhibited {
               t.Fatalf("b=%+v", got)
           }
           for _, name := range prepared.universe().Fields {
               if name == "b" { t.Fatal("allOf-prohibited syntactic name was enumerated") }
           }
       }

   Run `go test -mod=readonly ./pkg/validation -run TestJSONSchemaAllOfClosureExcludesSyntacticCandidate -count=1`, observe missing definitions, then implement only the target decoder and projection slice needed to make it pass. Extend with nested required/optional parents, required names absent from properties, open unknown fields, boolean schemas, schema-valued additionalProperties, multiple matching patterns, scalar-parent prohibition, literal dotted/nested collisions, and direct arrays versus descent.

2. Build a duplicate-aware object/boolean schema tree and resource index using `encoding/json`, `net/url`, and internal Unicode validation. Default dialect exactly `https://json-schema.org/draft/2020-12/schema`; accept its canonical optional trailing `#` form. Reject other `$schema` values wherever a schema resource declares them. Validate recognized keyword shapes (schema-valued positions, string arrays, nonnegative cardinalities, object maps, type names, refs/IDs/anchors); annotation data is not a schema. Index `$id` and `$anchor` only through schema-bearing positions, including unsupported applicators' schema children. Respect `$id` resource boundaries and relative bases. Decode URI-fragment percent escaping and JSON Pointer `~0`/`~1` once; malformed escapes are input errors. Reject conflicting identities, duplicate anchors, and pointers to existing non-schema values. Missing well-formed local resources/anchors/pointers yield unresolved evidence, never network/file access. Relative external refs without a usable absolute base remain unresolved, not cwd-relative.

   Add exact local-resource tests using root `$id: https://schemas.example.test/event`, an `actor` ref to `user#/$defs/user`, and resource map key `https://schemas.example.test/user`. A root `$ref` plus a closed sibling must intersect. Test nested `$id`, anchors, escaped pointers, embedded resource dialect errors, no-progress cycles, recursive `child.child.name`, duplicate IDs, and `$id` inside `examples` not becoming a resource. No-progress recursion and a documented maximum of 4096 distinct projection states per requested path or 128 field path segments return indeterminate with a traversal-budget reason. These are query-projection limits, not runtime input size limits or silently rejected catalogs.

3. Evaluate one requested path against all conjunctive object constraints and alternatives. Preserve declaration basis separately from requiredness; full-path requiredness requires all ancestors and conclusive object shape. `allOf` uses intersection and can use a conclusive local prohibition despite uncertainty in another conjunct. `anyOf` records all branches; all prohibited means missing, mixed admitted/prohibited means conditional/incomplete, all admitted uses the common supported conclusion, unresolved branches prevent a stronger claim. `oneOf` retains exactly-one semantics and branch evidence; do not infer satisfiability/exclusivity from distinct property names or query values. Conservatively mark conclusions that depend on unresolved exclusive branch viability indeterminate, including duplicate equivalent branches. A single branch or a field fact independently proved by a supported sibling may still be conclusive. Do not add a general solver.

   Keep effects query-relevant: unresolved refs or unsupported constraints at an unrelated declared sibling need not poison `host`; constraints on the containing object can affect its membership/requirement. Recognize `unevaluatedProperties`, `not`, active conditionals, dependent schemas/required, propertyNames, object const/enum/cardinality, dynamic refs, and required custom vocabularies as field-affecting incomplete where relevant. Scalar leaf bounds/formats do not typecheck the query. Absent `if` makes `then`/`else` inactive. Test requiredness uncertainty separately from membership certainty and retain explanatory pointers.

4. Implement a documented sound ASCII regex subset before invoking Go `regexp`: ASCII literals; `^` and `$` at expression edges; positive ASCII character classes with literal ranges; escaped ASCII regex punctuation; and `?`, `*`, `+`, `{n}`, `{n,}`, `{n,m}` on a preceding literal/class, with canonical decimal counts (`0` or a nonzero digit followed by digits). No dot, negated classes, groups, alternation, shorthand/property escapes, lookaround, backreferences, flags, or non-ASCII syntax in this first subset. Non-ASCII or line-terminator-containing candidate names are indeterminate where pattern matching matters, avoiding Go/ECMAScript end-anchor and Unicode differences. Patterns are unanchored unless written with anchors. A rejected subset is unsupported, not a false nonmatch; valid syntax outside it remains indeterminate. Recognizably malformed patterns within the supported grammar are input errors; valid repetition counts exceeding Go regexp implementation limits are unsupported/incomplete, not mislabeled invalid ECMAScript. Do not claim to diagnose every malformed ECMAScript construct outside the subset. Leading-zero counts (`a{01}`, `a{01,02}`), empty classes (`[]`), and valid JavaScript literal-brace forms (`a{`, `a{,2}`, `{}`) are unsupported/incomplete rather than mislabeled malformed. Add these conformance vectors alongside genuinely malformed `a{2,1}` and `a**`; Go acceptance or compile failure alone is not ECMAScript validity evidence. Add golden membership vectors for `^host_[a-z]+$`, `ip_[0-9]{1,3}`, escaped punctuation, unanchored matching, and unsupported lookahead/Unicode/newlines. Ensure an unknown pattern cannot silently activate additionalProperties false.

5. Enumerate named candidate paths without traversing arrays indefinitely or cycles. Enumerate syntactic names only as an upper bound, then project and exclude effective-schema prohibited names. Closed conjunctions can bound candidates from an open conjunct; unsupported or patterned cases conservatively set Complete false if exhaustiveness cannot be proved. Ordinary supported finite closed nested objects with bounded leaf shape can be complete. A boolean-true or unconstrained leaf can itself admit arbitrary object descendants, so a closed root alone does not make the full path universe finite. Include object/array declarations themselves; arrays or unknown descendant shapes make exhaustive descendant enumeration unavailable where the selector could reach them. The exact resolver remains usable for open unknown names and nested generic descendants. Never materialize infinitely many pattern names. Bound query-independent enumeration separately from exact projection: at most 4096 work units and 4096 distinct candidate names per prepared target. Charge one work unit before following each schema/reference/object edge at a path prefix or emitting each candidate; stop before exceeding either bound, including before enqueueing further work. Traverse sorted member keys/property names and stable schema-array order so the retained prefix and evidence are deterministic. Shared schema locations reached under different path prefixes must not be collapsed in a way that loses distinct names. If either bound is reached while work remains, retain already discovered safe candidates, set Complete false, and record `enumeration_budget` in target limitations and incomplete consuming-wildcard evidence. The truncation is not an input error or a missing-name proof. The exact resolver runs independently with its own per-path budget and can conclusively admit a name absent from the truncated seeds; unrelated exact obligations need not become incomplete merely because enumeration was bounded.

   Add a small shared-reference acyclic DAG fixture: thirteen successive definitions each have closed object properties `a` and `b` referring to the next definition; the final definition is a scalar. Its compact schema has exponentially many distinct field paths despite no reference cycle. Assert the 4096 enumeration-work bound is reached, retained candidates and budget evidence are equal across repeated/concurrent preparations, Complete is false, the lexically late all-`b` path absent from seeds is admitted by exact resolution, and `table *` remains incomplete without a fabricated missing outcome. Also test both bounds with many named properties and the same accounting in OCSF object/category enumeration. Test the universe via Task 1's public hook as well as direct projection.

6. Run `go test -mod=readonly ./pkg/validation ./pkg/analysis`, then the focused race tests for shared prepared targets. Copy/mutation tests must prove preparation does not mutate caller maps/raw bytes and concurrent projections have deterministic evidence. Commit only this task's actual files with `feat(validation): project local JSON Schema field contracts`, then independent combined review and committed fixes before Task 3.

## Task 3: Consume genuine compiled OCSF catalogs


Own `pkg/validation/ocsf.go`, `ocsf_projection.go`, `ocsf_test.go`, `ocsf_projection_test.go`, the OCSF dispatch in `schema_target.go`, and `testdata/schemas/ocsf/1.6.0/` plus `testdata/schemas/ocsf/edge-cases.json`. Consume Task 2's projection/evidence contract. Produce `prepareOCSF(target SchemaTarget) (preparedSchemaTarget, error)` and complete the shared target dispatcher. No raw-schema inheritance compiler is added.

1. After production release, verify retained raw bytes before copying. The prepared root is `/private/tmp/spl-toolkit-ocsf-preflight`. Base SHA-256 is `9b609f8fb670772f04191c1c276b46d34d6e9110d2417c71fa89c4f54c585137`; Windows is `19af77ce259f3ff57debc33e52da51b8220a1af59399d06553f29629678595e9`. Use Python 3.11-compatible deterministic gzip preparation:

       from pathlib import Path
       import gzip, hashlib
       prepared = Path('/private/tmp/spl-toolkit-ocsf-preflight')
       destination = Path('testdata/schemas/ocsf/1.6.0')
       expected = {
           'base': '9b609f8fb670772f04191c1c276b46d34d6e9110d2417c71fa89c4f54c585137',
           'windows': '19af77ce259f3ff57debc33e52da51b8220a1af59399d06553f29629678595e9',
       }
       destination.mkdir(parents=True, exist_ok=True)
       for name, digest in expected.items():
           data = (prepared / (name + '-catalog.json')).read_bytes()
           assert hashlib.sha256(data).hexdigest() == digest
           with (destination / (name + '.json.gz')).open('wb') as output:
               with gzip.GzipFile(filename='', mode='wb', fileobj=output, mtime=0) as compressed:
                   compressed.write(data)

   Copy upstream LICENSE and NOTICE from both retained source roots as `SCHEMA-LICENSE`, `SCHEMA-NOTICE`, `COMPILER-LICENSE`, and `COMPILER-NOTICE`. Preserve their bytes, including Broadcom/Symantec ICD attribution. Write a portable `provenance.json` containing repository URLs, full revisions, release/version, compiler's reported `0.0.0-dev`, Python 3.14.1 preparation, exact commands, raw sizes/hashes, compressed hashes, and extension identities. Do not copy temporary absolute locations as the only reproduction instructions. Add README reproduction commands using ordinary official compiler invocation from its `src` directory: `python3.14 -m ocsf_schema_compiler SCHEMA --ignore-platform-extensions`, and add `--extensions-path SCHEMA/extensions/windows` for the Windows file. Official default compilation includes platform extensions; selected explicit output contains only `win`, UID 2, version 1.6.0, `platform_extension?: false`. Base contains no extensions. No runtime/test execution invokes the compiler.

   If temporary inputs vanished, retrieve the same official codeload commit archives with normal certificate verification into a new task-scoped temp directory, extract with a safe tar data filter, and rerun the recorded commands. Never substitute current upstream HEAD or regenerate expected hashes from different sources. Source pin mismatch is a preparation failure to report to root.

2. Add a failing public-shape preparation test reading `base.json.gz` with `compress/gzip` and verifying its raw hash:

       func TestOCSFRealAuthentication(t *testing.T) {
           raw := readOCSFFixture(t, "base")
           prepared, err := prepareOCSF(SchemaTarget{Kind: "ocsf", Catalog: raw,
               Selection: &OCSFSelection{Version: "1.6.0", Class: "authentication", Profiles: []string{}, Extensions: []string{}}})
           if err != nil { t.Fatal(err) }
           for name, outcome := range map[string]string{"time": "required", "unmapped.vendor_field": "permitted_unspecified"} {
               if got := prepared.project(name); got.Outcome != outcome { t.Fatalf("%s=%+v", name, got) }
           }
       }

   Define the test-only `readOCSFFixture(t *testing.T, variant string) json.RawMessage` to open `../../testdata/schemas/ocsf/1.6.0/` plus variant `.json.gz`, decompress with `gzip.NewReader` and `io.ReadAll`, close both readers, and check `sha256.Sum256` against the pinned variant hash above before returning bytes. Run `go test -mod=readonly ./pkg/validation -run TestOCSFRealAuthentication -count=1`; observe missing implementation, then decode only normal compiled tables and implement the selected class path. Keep a shared test-only helper that reads gzip bytes and checks provenance, not a production gzip API.

3. Validate compile_version/version/table shapes, exact key/UID identity, recognized attribute requirement/profile metadata, local object links, and concrete category links. Normal descriptive/browser metadata is tolerated. Reject direct base_event, hidden/abstract classes, empty categories, conflicting/unknown selections, ambiguous UIDs, invalid profile applicability, and extension-set mismatch. Exclude the exact base_event 0/other sentinel before concrete link validation. Accept key `authentication` and UID3002 as equivalent normalized member selections; accept Windows key `win/registry_key_activity` and UID201001 with extensions `["win"]`. Sorting profile/extension copies is allowed; duplicates are errors. Never infer selected class from `search class_uid=3002`.

4. Traverse compiled attributes and exact local object_type links. Declared structures are closed except json_t and object_t resolving exactly to the base key object. Direct array declarations are checkable; descendants produce located array_traversal incompleteness. Recommended means optional full-path admission with recommended evidence. Missing/null/empty compiled profile dependencies are unconditional; otherwise require intersection with selected profiles. Keep all category classes when profiles apply to a subset. Unknown profile inheritance may make absent names indeterminate while known unconditional declarations remain conclusive. A strict subset of multiple enabling profiles plus required merged metadata makes affected requirement indeterminate; all profiles selected may use required. Preserve at_least_one/just_one constraint evidence; only a supported singleton establishes its sole member's presence. Unknown relevant constraint forms are incomplete.

5. Use real catalog cases for required `time`, required `file.name` in file_activity, optional `actor`, selected `cloud` and `datetime`, generic `unmapped.vendor_field` and `file.xattributes.vendor_field`, declared `observables` versus unsupported descendant, and IAM `group`: supporting classes authorize_session3003 and group_management3006; missing classes account_change3001, authentication3002, entity_management3004, user_access3005. IAM `time` is required in all six. A windows-patched base object must be present only with the exact compiled set; test mismatched empty extension selection fails atomically. Typed empty object and a subclass extending object stay closed in labeled synthetic edge cases. Also cover synthetic selected profile extends, unconditional-known versus absent-unknown membership, strict-subset requiredness, singleton and multimember constraints, cyclic object paths, malformed sentinel lookalikes, dangling links, and colliding UIDs.

6. Run `go test -mod=readonly ./pkg/validation`, projection concurrency/race tests, and verify both compressed inputs round-trip to their pinned raw hashes. Commit the actual task files and four notices with `feat(validation): project official offline OCSF catalogs`; root obtains independent combined review before Task 4. Evidence must distinguish source-structure checks from successful toolkit query validation, which Task 4 adds.

## Task 4: Assemble canonical schema reports and ordered batches


Own `pkg/validation/schema_validate.go`, `schema_batch.go`, `schema_request.go`, `schema_validate_test.go`, `schema_request_test.go`, `schema_batch_test.go`, finish schema report types in `schema_model.go`, and make a narrow private finalization extraction in `validate.go` if needed. Add `testdata/schemas/cases.json` and `testdata/schemas/requests.json`. Consume both target preparers and the canonical source sidecar. Produce the four decode/validate operations and two report types specified above.

1. Add this failing end-to-end semantic test, with normal imports of analysis, encoding/json, and testing:

       func TestValidateSchemaInvalidWinsHistoricalIncomplete(t *testing.T) {
           target := SchemaTarget{Kind: "json_schema", Schema: json.RawMessage(`{"type":"object","properties":{"host":true},"additionalProperties":false}`)}
           report, err := ValidateSchema(analysis.QueryDocument{Text: "search absent=x | mystery host"}, target)
           if err != nil { t.Fatal(err) }
           if report.Status != analysis.Invalid || report.Coverage.SemanticComplete {
               t.Fatalf("report=%+v", report)
           }
           found := false
           for _, outcome := range report.Outcomes { found = found || outcome.Outcome == "missing" }
           if !found { t.Fatalf("no definite missing obligation: %+v", report.Outcomes) }
       }

   Run `go test -mod=readonly ./pkg/validation -run TestValidateSchemaInvalidWinsHistoricalIncomplete -count=1`, confirm missing API, then add preparation, canonical analysis, and source outcome translation. Normalize documents using M3 rules before analysis. A schema error returns nil report and `InputError`; syntactically invalid query text returns a normal invalid report. Initial exact prohibited source reads become missing; derived reads are matching; kernel unavailable wins before schema membership; remove/not_applicable and creation-only/nonfield references are not obligations.

2. For selectors use sidecar final reference IDs, preserve each proven partial match and its binding, and project only source matches. An incomplete expansion is indeterminate, with `matches_complete:false`; zero complete consuming expansion is missing unless structurally unavailable. Every included missing obligation emits an error even if requirement-summary precedence otherwise selects indeterminate. Complete mixed matches summarize indeterminate, conditional, permitted_unspecified, optional, then required; derived entries are neutral, all-derived is matching. Keep every candidate's evidence. Conditional/indeterminate schema obligations make schema_complete false. Status is invalid if any diagnostic is error, otherwise incomplete if any coverage component is false, otherwise valid. Preserve earlier syntax/semantic reasons and diagnostics. Use existing `SPL_UNKNOWN_FIELD` and `SPL_INDETERMINATE_FIELD`; retain kernel unavailable diagnostics once. Stable new evidence reason identifiers include conditional_membership, unsupported_schema_keyword, unsupported_pattern, unresolved_local_reference, array_traversal, path_ambiguity, traversal_budget, enumeration_budget, unrepresentable_source_name, ocsf_category_conditional, unsupported_profile_inheritance, and merged_profile_requirement.

3. Share sorting/deduplication/error-precedence code with the existing finalizer through a small private helper accepting diagnostics and coverage components, without widening old JSON structs. Test byte-equality of all M3 reports before/after extraction. Preserve locations from original references, not candidate names; stable sort evidence by resource/pointer/operator/branch/class/name/reason, and matches lexically. Retain both negative and positive allOf/alternative/category provenance. Compact report target fields never contain the raw catalog, even in a one-document batch.

4. Implement strict `{document,target}` and `{documents,target}` wrappers with existing object/decodeDocument/DecodeDocuments helpers. Reject null/extra/duplicate members, malformed Unicode, invalid document options, empty batches, and invalid targets before producing results. Prepare once per batch and use the immutable prepared object for each normalized document. Test internal preparation count without exporting a counter or mutable global hook: factor `validatePreparedSchema(document, prepared)` and `validateSchemaBatchPrepared` helpers and test the preparation wrapper's single call using a local injected function in a private testable helper if needed. Do not reparse catalogs inside the per-document loop.

5. Create a versioned static cases file with target inline JSON or a relative real fixture reference, document(s), selected expected outcomes/coverage, and full canonical expected report JSON. Generate a proposed full report once from Go, inspect it against hand-authored semantic expectations and original span slices, then freeze it. No expected report is generated from the implementation during test execution. Include closed/nested/open/pattern/ref/allOf/anyOf/oneOf/unsupported/array/derived/removal/partial recovery and both real catalog selections/category/profile cases. Add malformed request byte fixtures separately; they are not successful report cases. Run table cases through `ValidateSchema` and `ValidateSchemaBatch`, caller-data isolation, concurrent equality, and invalid-over-incomplete batch precedence.

6. Run `go test -mod=readonly -race ./pkg/analysis ./pkg/validation`. Commit task files with `feat(validation): expose schema reports and batch APIs`; obtain independent combined review. This is the first task whose acceptance establishes actual toolkit validation of the pinned catalogs.

## Task 5: Expose schema validation in CLI and REST


Own `cmd/schema_validation_cli.go`, `cmd/schema_validation_cli_test.go`, focused changes in `cmd/cli.go` and the new command/flag help declarations in `cmd/main.go` (plus directly affected existing help expectations); `pkg/api/schema_validation.go`, `schema_validation_test.go`, `server.go`, `tools/update_validation_openapi.py`, `tools/tests/test_update_validation_openapi.py`, and route model/documentation declarations used by the final accepted M3 adapter. Task 5 also owns the generated `docs/docs.go`, `docs/swagger.json`, and `docs/swagger.yaml` output of its required generation/reconciler step, so the route commit carries coherent OpenAPI artifacts. Task 7 may update these again only for final documentation reconciliation. Consume public schema request/report APIs; perform no schema semantics in adapters. The read-only preflight confirmed `runCLIWithInput`, `parseCLIOptions`, `setCLIOption`, `analysisOptionMayBeEmpty`, `writeCLIError`, `writeCLIResult`, `marshalCLILine`, and `analysisStatusExitCode`. Extend the existing command-specific option gates to include validate-schema; keep both flag-presence and empty-value handling consistent. Recheck these names after root closes the final M3 review, not by replacing the parser.

1. Add a CLI test using `runCLIWithInput(args, stdin, stdout, stderr)`, a temp schema file containing closed `host`, and `validate-schema --schema PATH --query 'table absent' --format json`. Require exit1, nonempty canonical report matching direct Go byte-decoded JSON, and empty stderr. A supported open unknown field exits0 with permitted_unspecified; unresolved reference/wildcard exits3; bad target file/options exit2 with no report. Write report output before returning a content exit code. Run `go test -mod=readonly ./cmd -run Schema -count=1` and observe command/flag failures before adding dispatch.

2. Extend only needed option fields for `--schema`, `--schema-base-uri`, `--schema-resources`, `--ocsf-catalog`, `--ocsf-version`, `--ocsf-class`, `--ocsf-category`, repeatable `--ocsf-profile`, and repeatable `--ocsf-extension`. Reuse the exact M3 query/document/output flags `--query`, positional query, `--file`, `--stdin`, `--batch`, `--language`, `--profile`, `--compatibility-version`, `--source-id`, `--format text|json`, and `--output`. Require exactly one target source and one query input under M3 rules. Schema/catalog/resources paths cannot be `-`; stdin remains query/batch input. Reject target-specific flags in the other mode, query-source conflicts, duplicate scalar flags, unknown flags, and global document options with batch. CLI numeric selector strings are nonnegative decimal UIDs; other strings are exact keys. Preserve separate document `--profile` and `--compatibility-version` versus OCSF options. Read schema resources as a local URI-to-inline-schema map; use target decode/preparation for semantics. Text output includes target selection, coverage, original locations, outcomes, and supporting/missing classes or branch evidence. JSON is the direct report. Test source byte preservation for file/stdin, explicit source_id override, CRLF/Unicode, output files, ordered batch, empty/malformed batch, and incompatible flags. Reuse M3 output/error helpers.

3. Add single and batch handlers at `/api/v1/query/validate-schema` and `/api/v1/query/validate-schema/batch`. Require application/json, use `http.MaxBytesReader(w, r.Body, 8<<20)` before `io.ReadAll`, and invoke the strict canonical decoders. Return 200 for valid/invalid/incomplete content, 400 for request/target/content-type/body-limit input errors, and 500 for unexpected internal failures. Normalize a MaxBytesError into a clear `request body exceeds 8 MiB limit` error. Leave M3 routes at 1 MiB. Register each new route once and add accurate OpenAPI annotations/full report schemas using the existing generation path. Extend `tools/update_validation_openapi.py` after pinned Swag generation to describe the strict tagged target union, object-or-boolean inline JSON Schema, URI-keyed local resources, official catalog object, and one-of-four OCSF selectors; do not let RawMessage become a base64 string schema. Wrapper unknown/null/member rules must match the decoder, while schema/catalog payloads allow their own vocabulary. Preserve the M3 reconciler checks. Extend its idempotence and JSON/YAML/Go-template equality tests, then run `GOCACHE=/private/tmp/spl-toolkit-go-cache GOMODCACHE=/private/tmp/spl-toolkit-remaining-gomodcache GOPROXY=file:///private/tmp/spl-toolkit-remaining-gomodcache/cache/download GOSUMDB=off GOTOOLCHAIN=local make generate-docs PYTHON=/private/tmp/spl-toolkit-remaining-venv/bin/python` and `/private/tmp/spl-toolkit-remaining-venv/bin/python -m pytest tools/tests/test_update_validation_openapi.py -q`. This is the proven offline recipe recorded in `.superpowers/sdd/milestone-3-implementation-plan/task-3-report.md` lines49–52: pinned `go run ...@version` needs the provisioned file proxy and disabled online checksum service, not `GOPROXY=off`.

       func TestSchemaRESTAcceptsOfficialWindowsCatalog(t *testing.T) {
           raw := readSchemaFixture(t, "windows") // test helper decompresses and checks pinned hash
           target := validation.SchemaTarget{Kind: "ocsf", Catalog: raw,
               Selection: &validation.OCSFSelection{Version: "1.6.0", Class: "win/registry_key_activity", Profiles: []string{}, Extensions: []string{"win"}}}
           request := validation.SchemaRequest{Document: analysis.QueryDocument{Text: "table time"}, Target: target}
           payload, err := json.Marshal(request)
           if err != nil { t.Fatal(err) }
           if len(payload) <= 1<<20 || len(payload) >= 8<<20 { t.Fatalf("wrong real payload size: %d", len(payload)) }
           httpRequest := httptest.NewRequest(http.MethodPost, "/api/v1/query/validate-schema", bytes.NewReader(payload))
           httpRequest.Header.Set("Content-Type", "application/json")
           response := httptest.NewRecorder()
           NewServer().Handler().ServeHTTP(response, httpRequest)
           if response.Code != http.StatusOK { t.Fatalf("status=%d body=%s", response.Code, response.Body) }
           var got validation.SchemaReport
           if err := json.Unmarshal(response.Body.Bytes(), &got); err != nil { t.Fatal(err) }
           want, err := validation.ValidateSchema(request.Document, target)
           if err != nil { t.Fatal(err) }
           if !reflect.DeepEqual(&got, want) { t.Fatalf("report mismatch: %s", response.Body) }
       }

   The test uses the observed M3 `NewServer().Handler().ServeHTTP` setup and must remain equivalent after final adapter reconciliation. A helper `readSchemaFixture(t, variant) json.RawMessage` in this test file uses `os.Open`, `gzip.NewReader`, `io.ReadAll`, and expected SHA constants; no production fixture dependency. Test both genuine raw catalogs through single and batch routes, including the compact target size. Use whitespace padding to exercise an exactly 8 MiB valid JSON request, then 8 MiB + 1 rejection; set content length to -1 in another oversized test. Assert legacy endpoints still reject >1MiB under their existing contract. Test malformed Unicode and all strict wrapper cases independently of adapter parity.

4. Run `go test -mod=readonly ./cmd ./pkg/api ./pkg/validation`, then focused handler concurrency tests. Commit changed task files with `feat(api): add schema validation CLI and REST surfaces`; obtain combined review before native work. Defer broad rebuilding and documentation acceptance to Task 7 after native changes.

## Task 6: Expose real native Python operations and package every input


Own focused changes to `pkg/bindings/bindings.go`, generated `python/spl_toolkit/libspl_toolkit.h`, `python/spl_toolkit/mapper.py`, new `python/tests/test_native_schema_validation.py`, `python/native-source-files.txt`, `tools/check_package.py`, and focused `tools/tests/test_package.py` changes. Consume schema decoders/validators and accepted M3 native result/lifetime helpers. Produce C `spl_mapper_validate_schema` and `spl_mapper_validate_schema_batch`, each `(int mapper_id, char *request_json) -> SPLResult *` in the generated cgo header (input bytes are read-only in behavior; match the existing generated ABI spelling), and Python `SPLMapper.validate_schema(query, target, *, language="spl", profile="splunkd", version="current", source_id="")` plus `validate_schema_batch(documents, target)`.

1. Write native Python tests before adding the export. Use the real library path from `SPL_NATIVE_LIBRARY` and the existing context manager:

       def test_native_schema_required_and_owned_result():
           target = {"kind": "json_schema", "schema": {"type": "object", "properties": {"host": True}, "required": ["host"], "additionalProperties": False}}
           kwargs = {"library_path": os.environ["SPL_NATIVE_LIBRARY"]} if "SPL_NATIVE_LIBRARY" in os.environ else {}
           with SPLMapper(**kwargs) as mapper:
               report = mapper.validate_schema("table host", target)
               assert report["status"] == "valid"
               assert report["outcomes"][0]["outcome"] == "required"
               report["outcomes"].clear()
               assert mapper.validate_schema("table host", target)["outcomes"]

   Build the current shared library and run the one test to observe the missing native symbol/method. Do not replace the library with a mock to make semantic acceptance pass.

2. Reuse the completed M3 `ownedMapperJSONResult(mapperID C.int, operation func() (any, error)) *C.SPLResult`; do not duplicate allocation or result serialization. Its accepted native guard allocates owned SPLResult; validate the handle through the registry; retain operation lifetime; reject nil JSON pointer, invalid UTF-8/surrogates, malformed wrappers, invalid target/document with error and no result; invoke canonical operations; marshal the entire report; populate owned C strings. Error paths remain owned and freeable. Register the two ctypes signatures as `[ctypes.c_int, ctypes.c_char_p]` returning POINTER(SPLResult). Python delegates both new wrappers through existing `_validate_fields_request(self, native, request)`, whose implementation is already operation-independent despite its private name. It serializes with `json.dumps(..., allow_nan=False)`, preserves NUL through JSON escapes, uses `_operation`, decodes success/error, and always frees the pointer in `finally`. Keep the private helper name and accepted field-list callers; no unrelated rename is necessary. No subprocess, REST client, Python field engine, or source-tree fallback. Closed mapper errors and concurrent close behavior follow M3 exactly.

3. Exercise full-report cases including real base and Windows catalogs, categories/profiles, single/batch invalid/incomplete, caller input/result isolation, repeated/concurrent operations, empty batch, null/malformed raw C requests, wrong handles, and lone Python surrogates. Assert arrays, original text/source_id/CRLF/Unicode locations, schema evidence, and per-match bindings. Race against close with the established `_operation` lifecycle test pattern; failed operations must not leak owned results or poison later valid calls.

4. Enumerate all new handwritten Go files in `python/native-source-files.txt`; keep M3 entries. Keep the schema corpus, compressed real fixtures, provenance and notices as repository test inputs and explicitly copy their exact closure into the checker's outside-checkout acceptance directory; follow M3's external-fixture pattern rather than adding catalogs as runtime package data. The sdist closure still adds the new Go sources under `_native_src` and the new native test, while the checker supplies the copied schema fixture tree to tests for each installed wheel. Add `test_native_schema_validation.py` to required native tests and fixed sdist entries; Task 7 adds test_schema_surfaces.py to copied acceptance inputs when it creates that file; do not register a nonexistent acceptance test during Task 6, because current tooling tests copy the full required acceptance list from the checkout. Task 6 already copies the schema fixture closure and supplies both future surface and current native fixture environments. In `install_and_check`, copy the schema fixture tree before `_run_required_suite` executes the native tests, then supply its absolute `SPL_SCHEMA_FIXTURES` path to both native and surface environments. Include the exact corpus, raw/compressed catalog, provenance, and notice hashes in returned evidence for each installation. A path set only in the later acceptance environment would leave the new native corpus tests without fixtures. Check exact file hashes in wheel/sdist build evidence, require real loaded library and installed module origin, and preserve zero-skip/nonzero-count guards. Extend tooling tests to fail if a required schema fixture/source/test is omitted. Keep fixture catalogs out of wheel runtime resources unless the existing test harness requires a separately copied acceptance fixture directory; validation never locates a bundled default catalog.

5. Run focused native tests against the newly built real library, all native ABI/validation tests, and `python -m pytest tools/tests -q` using the configured venv. Commit exact files with `feat(python): expose native schema validation and package checks`; obtain independent combined review. Installed wheel/sdist rebuild and all-surface parity are required Task 7 proof, not implied by source-native success.

## Task 7: Verify end-to-end parity, document the contract, and close review


Own `tests/acceptance/test_schema_surfaces.py`, the narrow OCSF selection-negative isolation correction in `pkg/api/schema_validation_test.go`, necessary schema corpus expectation refinements, `tests/acceptance/cli_examples.json`, `tools/check_package.py` acceptance wiring and focused `tools/tests/test_package.py` required-file assertions and `tools/tests/test_schema_docs.py` capability/documentation parity, permanent docs `README.md`, `docs/API.md`, `docs/quickstart.md`, `docs/cli.md`, `docs/api-server.md`, `docs/compatibility.md`, `docs/architecture.md`, `python/README.md`, `python/examples/basic_usage.py`, generated Swagger artifacts, and `_build_plan/milestones/4-json-schema-ocsf-validation/milestone-log.md`. The milestone log is development evidence, never a runtime dependency. Keep tutorial examples short and executable. No new semantics are invented during documentation; a semantic defect returns to a tested scoped fix and review. Root requested closure of Task5’s nonblocking negative-test blind spot during Task7: OCSF selection mutations must use an otherwise valid catalog/selection baseline or assert the precise canonical wrapper error, so empty catalog cannot independently cause HTTP400. Change only these cases, run their focused API test, and include the fix in Task7’s combined review; no full core/catalog suite is repeated solely for this test improvement.

1. Register the newly created test_schema_surfaces.py in tools/check_package.py ACCEPTANCE_FILES and add its required-copy regression in tools/tests/test_package.py. Explicitly import `cli_path`, `server_url`, `post_json`, and `required_absolute_path` from `test_surfaces`, as the M3 validation suite does; imported fixtures must be present for pytest collection. Implement a corpus-driven acceptance test that sends identical documents/targets to Go expected snapshots, the actual CLI binary, an actual running REST server on an isolated loopback port, and the installed native Python library. Compare complete decoded report JSON, not only statuses or counts. Include ordered batches and failed requests separately. Decompress the real catalogs to a task temp directory for CLI; send inline raw JSON for REST/native. Capture original inputs, target hashes/selections, expected outcome oracle, executable/library/wheel hashes, SHA, exit/HTTP status, and full result JSON in an evidence directory beneath `docs/evidence/` using the existing project convention. Redact nothing from synthetic public query inputs; do not use user secrets. Run in a subprocess environment with unusable proxy addresses and no external schema resources; an unresolved HTTPS ref must return quickly without retrieval. Runtime code review must confirm there is no HTTP/file opener in schema preparation; offline acceptance alone is not proof that no accidental branch could retrieve.

2. Add the practical local-resource CLI example to permanent docs with these complete files:

       event.schema.json:
       {"$id":"https://schemas.example.test/event","type":"object","properties":{"actor":{"$ref":"user#/$defs/user"}},"required":["actor"],"additionalProperties":false}
       resources.json:
       {"https://schemas.example.test/user":{"$defs":{"user":{"type":"object","properties":{"name":{"type":"string"}},"required":["name"],"additionalProperties":false}}}}

   Run `build/spl-toolkit validate-schema --schema event.schema.json --schema-resources resources.json --query 'table actor.name' --format json`; expect exit0 and required with evidence from both local resources. Removing resources yields exit3 unresolved reference, with no retrieval. Document raw OCSF compiler preparation from the pinned sources, platform-extension defaults, exact version and full extension selection, and `--ocsf-class authentication` versus `--ocsf-category iam` examples. Include `--batch` with two document objects, CLI exit codes, REST 8 MiB/400 limit behavior, and Python single/batch snippets. State that Go has no network or global catalog registry.

3. Begin the milestone log with `## What's new in the app` as its very first section, followed by concise nontechnical capability bullets. Then record what was built, implementation decisions, information for the next milestone, deviations and reasons, verified commits and exact test evidence. Publish the same user-facing capabilities in maintained docs: check nested field declarations; explain optional and category/branch-dependent fields; use local schemas and exact OCSF versions; keep uncertainty visible for open wildcard sets, unsupported patterns/keywords, array descendants, and profile-provenance gaps. Explain that declared arrays are supported while descent is incomplete, and required is a schema statement rather than event presence. Include the exact supported ASCII regex subset and traversal limits. Show compact target limitations and location/evidence examples. Update schema diagnostic reason documentation and ensure all new reason identifiers appear. Do not describe this as complete JSON Schema instance validation. Keep existing analysis capability output unchanged and test the schema capability registry/documentation parity.

4. Run the final verification commands below only after all production changes are ready. Capture command, actual input SHA, exit code, test counts, tool versions, and artifact hashes. The full required suite is run once on the final executable input; rerun an affected scope after fixes. If only documentation changes after a build, record exact source identity proof rather than pretending old tests ran against a new binary. After Task 7 self-review, commit exact paths and obtain its own independent combined spec/code task review, as for Tasks1–6. Address Task 7 findings with affected checks, committed fixes, and scoped rereview. Once that task gate passes, obtain the separate broad independent combined milestone review of the entire M4 range from the recorded final accepted M3 SHA through final M4 SHA. Address broad-review findings with affected checks, committed fixes, and scoped rereview. For documentation-only closure, preserve exact executable-input identity and rerun only relevant documentation checks; neither review gate requires duplicating unchanged production suites. Verify final SHA, ancestor relationship, status, and absence of unrelated staged changes. Root decides completion/integration; this task does not merge, push, or delete worktrees.

## Concrete verification commands


Run from `/Users/jacobdelgado/repos/spl-toolkit/.worktrees/remaining-milestones`. Existing prepared tools are `/private/tmp/spl-toolkit-remaining-venv/bin/python` and Go1.22.12 at `/private/tmp/spl-toolkit-floor-tools/gomodcache/golang.org/toolchain@v0.0.1-go1.22.12.darwin-arm64/bin/go`. Use task caches rather than global installs. Verify executables exist; if unavailable, report/reestablish the same documented toolchain rather than silently lowering the floor or skipping its checks.

    export GOCACHE=/private/tmp/spl-toolkit-go-cache
    export GOMODCACHE=/private/tmp/spl-toolkit-remaining-gomodcache
    export GOPROXY=off
    export GOTOOLCHAIN=local
    export PIP_CACHE_DIR=/private/tmp/spl-toolkit-pip-cache
    export PIP_NO_INDEX=1
    export PIP_FIND_LINKS=/private/tmp/spl-toolkit-offline-wheelhouse
    export PIP_CONFIG_FILE=/dev/null
    /private/tmp/spl-toolkit-floor-tools/gomodcache/golang.org/toolchain@v0.0.1-go1.22.12.darwin-arm64/bin/go test -mod=readonly -ldflags=-linkmode=external ./pkg/analysis ./pkg/validation ./pkg/mapper ./pkg/bindings ./cmd ./pkg/api
    make build-all
    /private/tmp/spl-toolkit-remaining-venv/bin/python -m pytest tools/tests -q
    /private/tmp/spl-toolkit-remaining-venv/bin/python tools/check_go.py
    /private/tmp/spl-toolkit-remaining-venv/bin/python tools/check_docs.py
    PYTHONPATH=python SPL_SCHEMA_FIXTURES=/Users/jacobdelgado/repos/spl-toolkit/.worktrees/remaining-milestones/testdata/schemas SPL_EXPECTED_VERSION=$(cat VERSION) SPL_NATIVE_LIBRARY=/Users/jacobdelgado/repos/spl-toolkit/.worktrees/remaining-milestones/build/libspl_toolkit.dylib /private/tmp/spl-toolkit-remaining-venv/bin/python -m pytest python/tests -q
    PATH=/private/tmp/spl-toolkit-floor-tools/gomodcache/golang.org/toolchain@v0.0.1-go1.22.12.darwin-arm64/bin:$PATH PIP_NO_INDEX=1 make python-build PYTHON=/private/tmp/spl-toolkit-remaining-venv/bin/python PIP=/private/tmp/spl-toolkit-remaining-venv/bin/pip
    /private/tmp/spl-toolkit-remaining-venv/bin/python tools/check_package.py --sdist dist/spl_toolkit-$(cat VERSION).tar.gz --wheel-dir dist --evidence /private/tmp/spl-toolkit-m4-package-evidence.json

`tools/check_go.py` already runs the full race suite as part of its formatting/vet/test gate; do not run a duplicate unchanged full race command separately. Keep the Go1.22 floor execution distinct because it verifies a different supported toolchain. The external-link option is proactive: the existing local macOS Go1.22 default linker produced a dyld LC_UUID failure, while external linking passed. Do not repeat that unchanged environmental failure as if it were a new product test. Preserve all newer tests added during M3 reconciliation in the floor scope. Existing acceptance harness environment variables and artifact locations must be read from its final M3 version before invoking standalone acceptance; the installed package command must demonstrably collect and pass the new schema suite, with no skips and no checkout imports. `make build-all` must run only when no other production task owns the shared build directory.

For each task, `git diff --check` must pass. Use explicit changed paths in `git add`; inspect `git diff --cached --stat` and `git diff --cached` before committing. Record the tested commit/range in root's review request and the review result in this plan. A passing predecessor suite is context, not evidence for changed M4 code.

## Validation and Acceptance


The feature is accepted when the same exact document and target produce equal full canonical reports through Go, CLI, real HTTP, source native Python, installed wheel, and a wheel built from sdist outside the checkout. The frozen corpus checks facts independently: closed absent invalid; nested required ancestors required; optional ancestor optional; open descendants permitted_unspecified; partial wildcards indeterminate with proven partial matches; allOf-prohibited syntactic names excluded; anyOf/category mixed membership conditional with negative evidence; exclusive branch uncertainty indeterminate; local refs only; direct arrays declared and descent incomplete; derived matches independent of catalog; structural absence preserved; invalid wins incomplete without hiding coverage.

Genuine base and Windows catalogs must work on all surfaces and both REST batch/single routes. Authentication time, file_activity file.name, IAM time/group, cloud/datetime profile selection, generic unmapped/xattributes, exact extension identity and Windows scoped class cases establish practical compatibility. The malformed-target corpus rejects unsupported dialect/compile version, invalid selections, corrupted local links, duplicate keys, null or mixed wrapper members, bad Unicode, missing/empty batches, and wrong types. Valid unresolved refs and unsupported field-affecting constructs return incomplete reports rather than input failures or false absence. No fixture is made smaller for REST by removing normal tables.

Task 1 retains exact initial source binding even when the external resolver prohibits it; Task 4 owns the resulting missing diagnosis. Proven partial source/derived matches do not erase conditional bindings or unknown later flow. A later stats output may have conclusive selectors while report historical semantic coverage stays false. The existing M3 finite nil/empty and incomplete sidecar cases remain byte-equivalent. Caller data, repeated/concurrent calls, and owned native result/free behavior are covered by meaningful tests and race execution.

The final review covers schema logic, engine integration, all public surfaces, installed packaging, docs, performance bounds, and pinned provenance/license closure. Report the exact tested SHA and all remaining platform/operator gates plainly. Green local macOS tests do not establish an unrun Linux/Windows release matrix.

## Idempotence and Recovery


Preparation and tests use immutable pinned inputs and task-scoped temporary output. Repeating gzip preparation yields equivalent raw hashes; record compressed hashes but raw bytes are the compatibility identity. If an existing fixture differs, stop and inspect provenance rather than overwrite it from new upstream data. Raw request/target data are copied at preparation, and prepared indexes are immutable; retries do not modify user data or the filesystem.

If a task fails, keep its red test, fix only owned paths, rerun affected checks, and commit the tested fix before rereview. Do not delete another agent's changes, reset shared build outputs during concurrent work, or bypass required installed tests. Network/toolchain preparation failures are reported separately from runtime regressions. No migration, deployment, remote write, or destructive cleanup is needed.

## Artifacts and Notes


The accepted preflight source roots are `/private/tmp/spl-toolkit-ocsf-preflight/ocsf-schema-d0cd8a0fef198bf93086044e33fda6d80c74b9ef` and `/private/tmp/spl-toolkit-ocsf-preflight/ocsf-schema-compiler-d6b0b781d51a6b9682ea99396636ae01437b41b4`. `provenance.json`, compilation logs, repeat outputs, and `inspect_catalogs.py` remain there. Base has 75 classes including the sentinel, 163 objects, 11 profiles; Windows has 82 classes including the sentinel, 167 objects, 11 profiles. Repeat runs were byte-identical. These observations were rechecked as planning input; future runtime evidence belongs to Task 4 onward.

Pinned source facts underlying qualified behavior: compiler `src/ocsf_schema_compiler/compiler.py` profile inclusion around907, class inheritance1156, object inheritance1311, profile union/strongest-requirement merge1734; schema `objects/object.json` generic definition; dictionary attributes unmapped around6319 and xattributes6555; file xattributes usage193 and process65. Preserve these pointers with fixture provenance so a reviewer can verify the narrow generic-object and profile decisions without trusting a synthetic format.

Plan revision 2026-09-07: Initial execution plan against accepted M3 core, including root-approved partial-universe admission rules and 8 MiB schema-route bound. Full M3 adapter acceptance and root production release remain explicit gates. No production work was performed while writing this plan.

Self-review 2026-09-07: Mapped every approved design section to Tasks1–7; checked public/private type names and adapter file ownership; replaced incomplete test snippets; verified the request-limit and real-catalog acceptance paths, milestone-log structure, OpenAPI reconciliation, and package closure. No production tests were run during planning. Final accepted M3 adapter reconciliation remains required.

Plan revision 2026-09-07, controller review corrections: Fixed the two initial unknown-command regressions to start with `| mystery`; bounded query-independent candidate discovery to 4096 work units/candidates with deterministic safe truncation, independent exact resolution, explicit enumeration_budget evidence and acyclic shared-reference DAG tests; repeated every approved CLI flag; removed the duplicate final full-race invocation; adopted the verified local-file-proxy/offline-Swag recipe; and separated Task 7's committed combined task review from the broad milestone review without redundant tests for unchanged executable inputs. The controller approved the underlying design, regex subset, per-path limits, types, transport bound and fixtures; final written-plan approval and production release remain pending.

Plan revision 2026-09-07, bounded final-interface preflight: Recorded completed M3 symbols at c6ae684, reused actual native/Python ownership helpers and generated ABI, identified parser gates and pre-native fixture copying, retained the 37-source manifest and strict installed-checker closure, and aligned package build/evidence commands with the proven M3 recipe. No design contract changed and no production work, tests, builds or commits ran. Full M3 acceptance and explicit controller release are still required.

Production release 2026-09-07: Root accepted M3 through753bd5a with final documentation handoff e401402e5d8cd80f469884b2c3ad80245266ad66 and explicitly released all seven serial SDD tasks. Read-only comparison from prior preflight found only three M3 guidance-file changes, so predecessor runtime/test/source/package proof is reused without baseline reruns. Source Python schema tests now receive absolute SPL_SCHEMA_FIXTURES alongside the library path; installed tests retain checker-copied fixtures. No push, merge, publish, or destructive cleanup is authorized. Earlier pending-gate statements in historical planning sections are superseded by this release record.

Execution clarification 2026-09-07: Root independent regex comparison approved canonical decimal repetition counts and explicit unsupported/incomplete classification for leading-zero counts, empty classes and literal-brace JavaScript forms. Task2 receives focused conformance controls before its review; this refines the approved conservative subset without a JS runtime dependency.

Execution clarification 2026-09-07: Root clarified dotted-name evidence: an actual declared literal dotted property is conclusive only when the competing nested interpretation is conclusively prohibited. Both admitted or an open/unresolved nested interpretation means indeterminate, including literal dotted segments at intermediate object levels. Do not invent literal-dot alternatives from generic openness alone; ordinary nested/generic descendants remain checkable. Add literal-only closed, explicit collision, open competing-path and intermediate-segment controls within existing budgets.

Execution ruling 2026-09-07: Valid blank JSON property names remain legal schemas. Omit only candidate names that cannot satisfy the canonical nonblank SourceUniverse input contract, retain representable descendants, and mark affected enumeration partial with unrepresentable_source_name; normal exact paths remain conclusive. This avoids turning seed construction into a false malformed-schema error and adds no SPL syntax.

Supported-floor package verification correction, approved by root during Task7: Root-approved supported-floor packaging correction: Go1.22 and Go1.25 cgo headers differ in compiler-owned prologue and known empty-parameter spelling, despite identical toolkit layouts/17exports. Keep exact mapper.py and sdist source/header hashes, actual wheel-header hash evidence, and loaded-library hash binding. In verify_wheel_sources compare COMPLETE toolkit C preamble/layout and COMPLETE export region through validated unique cgo section boundaries; exclude only compiler-owned prologue and #line metadata, normalize only known ()/(void) empty-parameter spelling. Do not extract only matching spl_* lines, ignore unexpected declarations or malformed/ambiguous markers, or add a broad C parser. Test real1.22/1.25 acceptance and missing/new/wrong signature, changed struct field and malformed/ambiguous marker rejection. Task7 owns narrow checker/regression/report changes and resumes the already-built artifacts after affected checks; no product/toolchain change or unrelated Go/native rerun.

Task7 header regression ownership clarification: retain the actual generated Go1.22 and Go1.25 headers as tooling-only fixtures at tools/tests/fixtures/cgo-go1.22.h and tools/tests/fixtures/cgo-go1.25.h, with exact hashes/toolchain provenance. These are not native/package/runtime inputs; use them with the approved complete-section mutation regressions.

Task7 I1 evidence qualification approved by root: successful checker TemporaryDirectory cleanup removed the rebuilt-sdist wheel; no retained archive exists. Do not rebuild or reconstruct. Supplement with corrected verify_wheel_sources on the retained original wheel and corrected complete-header comparison on both exact actual header fixtures. Bind rebuilt-sdist header bytes by prior package.json wheel_payload_hashes b800c47a3f7b1de0dfd0b78c45eb767c4ac54c4016441d6ca1a431d6e639fda1, equal to retained Go1.25/source header. Prior rebuilt-wheel/native hashes and installed/source evidence remain historical executed proof; fresh whole-archive verification is original-wheel only, rebuilt-header revalidation is by exact recorded payload hash. Do not relabel old checker run or repeat unrelated builds/tests.

Final closure 2026-09-08: all earlier pending-production/task-review/root-oracle statements are historical instructions superseded by the accepted status and Outcomes above. Task7 I1 now compares the complete header with exact known compiler-delta normalization; no unchecked compiler region is discarded. Approved guidance is tracked with this documentation-only closure; copied prompt/PRD/other milestone guidance remains protected.
