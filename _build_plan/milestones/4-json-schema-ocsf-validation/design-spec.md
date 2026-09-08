# Milestone 4 — JSON Schema and OCSF Validation Design

Status: implemented and independently accepted at `92879052b0454db12aead045188ce4b15fa9d83a` on 2026-09-08. All seven task reviews, the broad milestone review, and root final Go/CLI/REST/native/Python 3.14 acceptance passed. Documentation-only closure preserves those tested inputs. The design was approved on 2026-09-07 and production was released after final M3 handoff `e401402e5d8cd80f469884b2c3ad80245266ad66`; earlier planning gates are completed history.

## Purpose and boundaries

Extend the canonical offline validation workflow to local Draft 2020-12 JSON Schema and explicitly selected versioned OCSF categories or event classes. Every result distinguishes schema declarations, requiredness, branch/category conditionality, openness, structural pipeline availability, and uncertainty. Go owns behavior; CLI, real native Python, and REST expose equivalent single and batch results.

The milestone checks query field references, not event records, field values, or expression types. It does not execute searches, download schemas, infer OCSF selections from query predicates, implement a proprietary registry, expand standalone SPL2 coverage, or modernize rewriting. Existing analysis, field-list validation, and legacy syntax validation stay compatible. `_build_plan/` remains temporary guidance, never a runtime dependency.

This follows the architectural brainstorming process. The controller is the delegated design approver and retains commit/integration authority. This document records the approved design; the adjacent implementation plan and milestone log record completed execution and acceptance.

## Shared architecture

Extend the Milestone 3 validation package with local target preparation and schema field projection. Keep command transfer, bindings, source locations, scope behavior, and supported wildcard operations in the canonical analysis kernel. There is no schema-side SPL parser or second field-flow engine.

JSON Schema projection evaluates a requested field path against local schema resources. OCSF projection evaluates the same path against explicitly selected compiled classes and objects. Both supply schema evidence and a source universe to the existing validator/kernel collaboration. Neither treats external schema absence as a structurally unavailable initial source name.

A schema universe can be partial. It contains known named candidates and an explicit statement of whether the relevant expansion is exhaustive; exact path resolution can remain conclusive even when name enumeration is unbounded. Do not flatten an open object, a regular-expression property set, or a recursive graph into a supposedly complete finite catalog. Known derived names remain kernel-owned. A wildcard whose complete expansion cannot be established stays indeterminate, retaining proven partial evidence without claiming its names are exhaustive. Unknown command effects and unresolved scope merges remain unknown even when schema membership is known. Query-independent candidate discovery also has a deterministic bound of 4096 work units and 4096 distinct candidates per target, covering acyclic shared-reference graphs as well as cycles. Reaching a bound retains safe partial candidates, makes enumeration incomplete, and records enumeration_budget evidence; it never proves an unenumerated name missing. Exact path resolution remains independent and may establish an unenumerated name conclusively within its own 4096-state/128-segment limits. Consuming wildcards that rely on truncated discovery remain incomplete.

After review of the accepted Milestone 3 core, the controller approved additive `analysis.AnalyzeWithSourceUniverse(document, SourceUniverse)` alongside the unchanged finite hook. The input contains copied concrete `Fields`, `Complete`, and an optional pure, deterministic, concurrency-safe `Resolve(name)` callback returning admitted, prohibited, or indeterminate external name membership. Complete means every potentially admitted source name is listed: unlisted names remain prohibited even if a callback would admit them. With a nil resolver, listed names are admitted and unlisted names are indeterminate for a partial universe or prohibited for a complete one. Listed resolver indeterminacy prevents conclusive expansion; invalid admission values become indeterminate. Existing duplicate/Unicode/name validation applies. The callback knows no query transfers or event presence. Initial exact source reads keep structural source binding regardless of external admission, and derived bindings bypass it.

The new hook retains only proven source/derived partial matches; the existing finite hook keeps its empty-on-incomplete sidecar. Local table/stats precision may recover only with proven membership and binding, never by turning unknown-stage provenance into certain source origins. Partial selection/removal and unknown internal retention keep later-flow uncertainty visible. This extends the existing refinement in place, discharges only uncertainty it actually resolves, and never clears diagnostic classes globally or overrides unrelated semantic coverage.

Two alternatives were considered. Treating every schema as a flattened field list is small but unsound for open objects and composition. Replaying query flow in a schema validator duplicates kernel semantics. The selected shared-kernel design retains precise source membership while allowing honest partial field universes.

## JSON Schema input and local resources

Support the standard Draft 2020-12 dialect. If `$schema` is absent, use Draft 2020-12 explicitly as the documented default. A declared different or unknown dialect is a request/input error, including in an embedded or externally supplied resource. Do not guess older `$ref` or composition semantics.

The caller supplies a root object or boolean schema, an optional absolute base URI, and an optional explicit map from absolute resource URIs to object/boolean schemas. A URI is an identifier only. Validation does not open it as a URL, file path, or other resource. Root-local references work without a caller base URI through a deterministic internal root identity. Relative cross-resource references need a resolvable supplied base or `$id`; no current working directory inference occurs.

Build a per-target local resource index respecting `$id` resource boundaries, `$anchor`, JSON Pointer URI fragments, and `$defs`. Resolve relative identifiers against the correct enclosing resource. Draft 2020-12 `$ref` siblings remain conjunctive constraints. Index only schema-bearing positions: an object inside examples, defaults, or arbitrary annotation data is not automatically a schema resource.

Malformed JSON, schema keyword shapes, URI/pointer syntax, conflicting resource identities, duplicate anchors in one resource, and explicit unsupported dialects are input errors. A syntactically valid reference whose target is not locally supplied yields an indeterminate path and an incomplete report with the unresolved URI. It never triggers retrieval. An existing pointer that resolves to a non-schema value is an invalid schema input.

Recursive object schemas are legitimate: traversal that consumes a requested path segment can follow a local cycle. Memoize by schema resource/location and remaining path. A non-progressing reference cycle or exhausted documented traversal budget yields indeterminate coverage and evidence; it is not fabricated absence or a process hang. A recursive universe is not exhaustively enumerated.

The resource rules follow [Draft 2020-12 Core](https://json-schema.org/draft/2020-12/json-schema-core). The supported scope is field projection, not a claim of full JSON Schema instance-validation conformance.

## JSON Schema field projection

Paths use the kernel's exact comparable field spelling. Traverse nested object properties without case folding or quote guessing. Reading an object or array itself can match its declaration. Traversing array contents is unsupported in this milestone and produces located indeterminate output. A collision between a literal dotted property name and a nested path is indeterminate; never silently select one representation.

For each property segment, apply its explicit `properties` subschema and every matching `patternProperties` subschema together. Evaluate `additionalProperties` only when neither an explicit property nor a pattern matches in that same schema object. Missing additional-property policy is open. Boolean false prohibits an unmatched field; boolean true permits it without a declaration. A schema-valued additional-property policy permits an unnamed property subject to its nested structure, so a child can be constrained without claiming the parent was explicitly declared.

`required` means required in the containing object; a nested required leaf is not globally required if an ancestor is optional or its object shape is not guaranteed. Preserve per-level evidence. A name in `required` need not also appear in `properties`; preserve requiredness separately from whether admission came from an explicit property, a pattern, or an open policy. A scalar-only parent prohibits object descendants when the structural type is conclusive. Scalar leaf bounds, formats, and value assertions do not invoke expression typechecking.

Property-name matching follows a documented sound subset of ECMAScript regular expressions, with its supported syntax and semantic equivalence verified by tests. Go's regular-expression acceptance alone is not proof of equivalent JSON Schema matching. Valid unsupported regex syntax yields indeterminate membership where it could matter, not a silent nonmatch. JSON Schema property patterns are not implicitly anchored; query selector wildcards retain the separate SPL semantics owned by the kernel. See the [official regular-expression guidance](https://json-schema.org/understanding-json-schema/reference/regular_expressions).

Composition keeps its logical structure:

- `allOf` intersects every conjunct, including sibling keywords. A declaration in one conjunct does not escape another conjunct's closure. One conclusive prohibition can establish a prohibited path even when a different conjunct declares it; record the relevant schema locations.
- `anyOf` retains alternatives. A path is definitely missing only when every possible branch conclusively prohibits it. Mixed admitted/prohibited branches are conditional and incomplete. Universally admitted paths retain the strongest common required/optional/unspecified conclusion supported by the evidence.
- `oneOf` retains exactly-one semantics and branch identities. Do not lower it blindly into `anyOf`, select a branch from a query value, or assume competing branches are satisfiable/exclusive. Where exclusivity or branch viability can change the field conclusion and projection cannot prove it, report indeterminate evidence. Common field facts are conclusive only when the field projection proves them independently of the unresolved branch choice.

There is no general schema satisfiability solver. Reports do not assert that an event exists that satisfies the target. Boolean schemas and supported structural contradictions may provide local field proofs; value predicates never become a guessed branch-selection mechanism. The [official composition guidance](https://json-schema.org/understanding-json-schema/reference/combining) explains why composition cannot be treated as ordinary object inheritance.

Known field-affecting keywords beyond the supported projection include `unevaluatedProperties`, `not`, active `if`/`then`/`else`, `dependentSchemas`, `dependentRequired`, `propertyNames`, object-level `const`/`enum` and property-cardinality constraints, dynamic references, and custom required vocabularies. If such a construct can affect the requested path's admission or requiredness, return indeterminate with its exact schema URI/pointer and keyword. Unsupported constraints in unrelated branches or value-only scalar leaf constraints do not automatically invalidate an otherwise conclusive field result. Do not silently ignore an unsupported construct when its effect cannot be bounded. Keyword syntax checks do not claim event validation. Requiredness follows the [Draft 2020-12 Validation specification](https://json-schema.org/draft/2020-12/json-schema-validation).

Approved behavioral examples:

| Target condition | Reference result |
| --- | --- |
| Declared root property in `required` | Required declaration. |
| Declared optional parent with a required nested child | Optional full path, with local child requirement retained. |
| Unknown exact field in an otherwise supported open object | Permitted without a declaration. |
| Unknown exact field in a supported closed object | Missing/invalid. |
| `allOf` of closed `{a}` and a separate declaration `{b}` | `b` is prohibited by the closed conjunct. |
| Closed `anyOf` branches declaring `{a}` and `{b}` | `a` is conditional/incomplete, with both branch results. |
| Undeclared exact name matching a supported property pattern | Pattern declaration, retaining its local optional/required status. |
| Array property read directly / unsupported descent through that array | Declaration / indeterminate, respectively. |
| Open or patterned schema with unenumerable `table host*` expansion | Indeterminate wildcard expansion, never a false empty match. |
| Missing source reference followed by an unsupported command | Invalid overall, incomplete coverage still visible. |

## OCSF catalog choice

Use the [official normal compiled OCSF format](https://github.com/ocsf/ocsf-schema-compiler/blob/main/docs/format.md), `compile_version: 1`, directly. The compiler's local output is the public input; no proprietary intermediate schema format is introduced. The controller approved exact extension-set matching because compiled extension patches cannot safely be undone by removing origin-tagged fields.

Users prepare catalogs outside validation with the [official compiler](https://github.com/ocsf/ocsf-schema-compiler). Its default includes platform extensions. `--ignore-platform-extensions` produces a base-only input; repeatable `--extensions-path` adds explicitly provided local extension sources. The documentation-verified commands and source links are in [research-notes.md](research-notes.md). Compiler installation and Python 3.14+ preparation requirements do not change SPL Toolkit's Python 3.11+ runtime baseline.

## OCSF selection and paths

Require an exact requested OCSF version matching the catalog's `version`. Require exactly one concrete event class or activity category, identified by its exact scoped table key or numeric UID. There is no latest-version alias, automatic class/category selection, or interpretation of `class_uid`/`category_uid` query comparisons as a request override. Reject direct `base_event` and abstract/hidden class selection: this workflow promises concrete event classes. Category membership excludes those entries as well.

Profiles and extensions default to empty explicit selections. Normalize their exact keys into sorted sets; reject duplicates, unknown selectors, conflicting class/category selectors, ambiguous UID collisions, malformed selection values, version mismatches, and an extension set different from the full compiled catalog set. Extension identities and their versions are retained in report metadata. No filtering or attempted undo of already compiled extension patches occurs. A caller who wants a different extension combination supplies a separately compiled local catalog.

Reject malformed compiled-format structure, unsupported `compile_version`, missing required class/category/object links, malformed requirement/profile metadata, and contradictory identities as target input errors. Validate concrete-class category links after excluding the recognized `base_event` sentinel: official normal catalogs retain its UID/category UID 0 and category `other` without an ordinary `other` category-table entry. This specific sentinel is not a dangling concrete-class link and does not excuse malformed links elsewhere. Ordinary descriptive or browser-only metadata does not become a runtime directive. Allow normal compiled documents with additional browser metadata, but read only the normal tables; do not use browser metadata to invent schema semantics or include hidden entries.

Use the class/object `attributes` already compiled by upstream. Resolve object-valued attributes through exact local `object_type` keys. Do not admit every dictionary attribute, re-run inheritance, re-merge profile definitions, or fetch unresolved references. Nested presence incorporates every ancestor. Required stays required only with a conclusive full-path requirement; recommended and optional yield optional admission while preserving the original distinction as evidence. An array field can be declared and checked directly; descending through it remains indeterminate. A `json_t` field permits arbitrary JSON descendants without declaring their names, so those paths yield permitted-unspecified evidence instead of a closed-object missing result. The same open boundary applies when `object_t` resolves exactly to the base generic `object` definition through `object_type: "object"`, as in official `unmapped` and `xattributes` fields. This rule follows that definition's explicit generic-object semantics; an empty attribute map or a typed subclass merely extending `object` does not establish openness. Ordinary typed class/object declarations are the closed OCSF field contract.

An attribute with no profile dependency is enabled unconditionally. Missing/null/empty dependency metadata is unconditional according to the compiled-format contract. Otherwise enable it when the selected profile set intersects its dependency names. Require each selected profile to be applicable to the selected class or at least one category member. Keep every category member in the comparison even if a profile only applies to some of them; profile selection must not silently narrow category membership.

Primary-source review identified that profile `extends` and merged multi-profile requirement metadata need explicit conservative treatment. The final evidence and behavior are documented in the profile qualification section below; they must not be silently flattened into an unconditional requirement.

OCSF `at_least_one` and `just_one` presence constraints are retained as evidence. A multi-member constraint does not make every member individually required. A supported singleton constraint may establish its sole member's presence; do not infer that conclusion for an unsupported or malformed relationship. Unknown field-affecting constraint forms give indeterminate results for affected queried paths. The toolkit does not test an event or infer co-occurring field values from a query.

## Profile qualification from primary code

The controller approved conservative treatment of information the compiled artifact cannot establish. Review of official compiler commit `d6b0b781d51a6b9682ea99396636ae01437b41b4` found profile inclusion labeling in [`_merge_attributes_include`](https://github.com/ocsf/ocsf-schema-compiler/blob/d6b0b781d51a6b9682ea99396636ae01437b41b4/src/ocsf_schema_compiler/compiler.py#L907), class/object inheritance calls at lines 1156 and 1311, and profile-name/requirement merging in [`_merge_attribute_properties`](https://github.com/ocsf/ocsf-schema-compiler/blob/d6b0b781d51a6b9682ea99396636ae01437b41b4/src/ocsf_schema_compiler/compiler.py#L1734). There is no corresponding profile-inheritance expansion in the reviewed pipeline, and merged requirements retain the strongest requirement across profile contributions rather than one requirement per profile.

Therefore a selected profile containing `extends` introduces unresolved inherited-profile effects. Do not invent selector closure, re-merge ancestors, or fabricate inherited attributes. Known unconditional fields can retain conclusive declaration facts, but an absent field that an unresolved inherited profile could introduce is indeterminate, never definitely missing. Mark the relevant universe as partial and give an explicit unsupported-profile-inheritance reason. Where inheritance could affect a queried declaration's requiredness or nested shape, that path is indeterminate as well.

If an attribute is enabled by a strict subset of its multiple profile dependency names and the compiled requirement is required, the artifact cannot prove that the selected subset establishes the requirement. Preserve the compiled metadata as provenance but classify the affected path as indeterminate. When all enabling profiles are selected, the compiled requirement applies. Optional/recommended results need no fabricated requirement increase. Ordinary complete profile selections remain conclusive. These are field-relevant coverage limitations, not grounds to pretend all profile selections are unsupported.

## OCSF categories and explanation

Derive the complete member set from the catalog's concrete classes and category metadata, then project each requested source path against every member. Empty categories are invalid selections. Report members by stable scoped key and UID; captions are display-only content.

- Required: the full path is conclusively required in every member.
- Optional: the path is declared in every member, but is not universally required.
- Conditional: the path is admitted by some members and conclusively prohibited by others. The report is incomplete and identifies both sets of classes.
- Missing: every member conclusively prohibits the path. The query is invalid at the original reference location.
- Permitted-unspecified: every member admits the path, but one or more only do so through an open generic JSON or base-object boundary, so a category-wide declaration cannot be claimed.
- Indeterminate: an unsupported path, unresolved membership, or unknown requirement can change the conclusion. Retain the members with conclusive positive/negative evidence as well as the unknown members.

Every category finding provides sorted supporting, missing, and indeterminate class identities and local attribute pointers. It distinguishes class-dependent declaration from a field that is simply optional in every class. It does not claim that a declared optional field occurs in every event. A definite missing or pipeline-unavailable field elsewhere still makes the whole report invalid while category uncertainty remains visible.

## Reports and diagnostics

Extend the Milestone 3 report machinery while preserving `validate-fields`' existing wire shape. Schema-specific reports use a versioned schema contract and the same canonical refined analysis, status, coverage, source-order outcome ordering, deterministic diagnostics, and ordered batch envelope. Choose exact shared helpers/types against the verified Milestone 3 code rather than forcing a premature serialization refactor.

A report includes compact normalized target metadata. JSON Schema metadata identifies the target, default/declared dialect, base identity, and sorted supplied resource identities; it does not repeat every schema document. OCSF metadata identifies the catalog version and compile version, exact class/category selection, profiles, full extension set with versions, and member class identities. Do not put the multi-megabyte compiled catalog in every batch report.

Schema reference outcomes are `required`, `optional`, `conditional`, `permitted_unspecified`, `missing`, `unavailable`, and `indeterminate`. Kernel-proven derived reads remain `matching`, with source obligations reported where originally consumed. Creation-only references and non-field dependencies are not schema obligations. A derived name requires no catalog declaration. A kernel-proven unavailable read remains unavailable even if the external schema declares it.

Keep declaration provenance separate from presence requirement. A `required` name admitted by an open policy has requiredness evidence without a fabricated explicit declaration. Evidence records the local schema URI and JSON Pointer, branch identity/operator or OCSF class identity, declaration basis, requirement, and a concrete explanation. Conditions and missing results preserve negative branch/member evidence. Match names sort lexically; evidence is deterministically ordered by resource/pointer and branch/class identity. Source spans always refer to original query text, including original wildcard spans, not generated candidate names.

Schema wildcard outcomes explicitly expose `matches_complete`. A partial universe may retain proven candidate names with `matches_complete: false`; these are not an exhaustive expansion. Candidate enumeration excludes names conclusively prohibited by the effective schema, including another allOf conjunct's closure; a syntactic property declaration alone is not an admitted candidate. A conclusive expansion has `matches_complete: true`. Exact missing outcomes do not fabricate successful matches. Unknown membership has no proven exact match and matches_complete false. When membership is conclusively admitted but requiredness alone is unknown, preserve the exact source match with outcome indeterminate and matches_complete true; schema_complete remains false and status is incomplete absent a definite error. Preserve specific profile-requiredness provenance to distinguish these cases. Conditional/indeterminate membership evidence can remain informative without converting the conclusion into a match. M3 field-list outcomes remain unchanged.

For a multi-match wildcard, retain every candidate's own result. If an included obligation nevertheless resolves missing, it produces an error and invalid status before requirement-summary precedence. An incomplete expansion is indeterminate; otherwise any indeterminate candidate wins, followed by conditional, permitted-unspecified, optional, and required. Derived matches are neutral for this schema-requirement summary; a wildcard containing only derived matches is matching. A supported consuming wildcard with no candidates is missing only when the complete universe and kernel environment prove emptiness. Structural unavailability remains kernel-owned. Exclusion/removal selectors never become missing-field obligations merely because they match no names.

Preserve validation coverage components `syntax_complete`, `semantic_complete`, and `schema_complete`. Conditional and indeterminate schema obligations make schema completeness false. Required, optional, and permitted-unspecified decisions can be complete when proven. Incompleteness from an unsupported field-affecting target construct remains visible even when no successful match can be emitted. Definite missing/unavailable errors make status invalid; otherwise any incomplete coverage makes status incomplete; valid requires complete, error-free analysis and schema projection.

Reuse the M3 unknown-field and indeterminate-field diagnostics and the kernel unavailable-field diagnostic. Add stable schema-ambiguity reason codes for conditional membership, unsupported schema constructs, unresolved local references, and OCSF category conditionality as needed by the final shared types. Every new code is documented and tested. Do not repurpose input errors as invalid query reports. Input errors include malformed target requests, bad target options, invalid documents, and inaccessible CLI files; a syntactically invalid query remains an individual canonical invalid report. Internal failures do not become successful empty results.

## Canonical Go and public adapters

The implemented public Go operations are `ValidateSchema(document analysis.QueryDocument, target SchemaTarget) (*SchemaReport, error)` and `ValidateSchemaBatch(documents []analysis.QueryDocument, target SchemaTarget) (*SchemaBatchReport, error)` in `pkg/validation`. Strict wire decoding is separate: `DecodeSchemaRequest(data []byte) (SchemaRequest, error)` and `DecodeSchemaBatchRequest(data []byte) (SchemaBatchRequest, error)` return the document(s) and target supplied to those operations. See [the permanent API reference](../../../docs/API.md#json-schema-and-ocsf-field-validation) and `pkg/validation/schema_model.go` for the request/report fields. The target is a strict tagged request union: `json_schema` carries root schema, optional base URI, and local resource map; `ocsf` carries raw official compiled catalog and explicit selection. Identity metadata is optional caller data, not retrieval instructions. Validation prepares the target once per batch, performs canonical kernel analysis and field projection per document, and returns input-ordered reports. An empty batch or invalid target rejects the operation. Aggregate precedence remains invalid, incomplete, valid.

Add CLI `validate-schema` with exactly one target source:

- `--schema <local-json-schema>` accepts a direct object/boolean JSON Schema. `--schema-base-uri` supplies the optional base identifier, and `--schema-resources <local-map.json>` supplies a URI-to-inline-schema map.
- `--ocsf-catalog <local-compiled-catalog>` requires `--ocsf-version` and exactly one `--ocsf-class` or `--ocsf-category`. Repeated `--ocsf-profile` and `--ocsf-extension` supply exact selection names. Reject JSON-target flags in OCSF mode and vice versa.

Reuse Milestone 3 inline query, positional query, file, stdin, and batch inputs and their conflict rules. Reuse language/profile/compatibility-version/source-ID options for individual documents; OCSF profiles have separate flag names. Preserve file/stdin content bytes and source IDs. Reuse `--format text|json`, `--output`, and exits 0 valid, 1 invalid, 3 incomplete, 2 usage/request/I/O/internal error. Write the report before returning a query content status. Text output explains target selection, field outcomes, conditional class/branch evidence, source locations, and completeness.

The new schema REST routes use an 8 MiB request-body limit, sufficient for the verified 3.1–3.4 MB normal catalogs; existing routes retain their 1 MiB limit. Genuine base and selected-Windows single/batch payloads and oversized request rejection are required acceptance inputs.

For example, `event.schema.json` can contain:

```json
{
  "$id": "https://schemas.example.test/event",
  "type": "object",
  "properties": {"actor": {"$ref": "user#/$defs/user"}},
  "required": ["actor"],
  "additionalProperties": false
}
```

Its local `resources.json` can contain:

```json
{
  "https://schemas.example.test/user": {
    "$defs": {
      "user": {
        "type": "object",
        "properties": {"name": {"type": "string"}},
        "required": ["name"],
        "additionalProperties": false
      }
    }
  }
}
```

The intended CLI usage is `spl-toolkit validate-schema --schema event.schema.json --schema-resources resources.json --query 'table actor.name' --format json`. The `.test` identifiers are resolved exclusively from these local values; the CLI never contacts that host. The binary name is confirmed by the current Makefile.

Python adds `SPLMapper.validate_schema(query, target, *, language='spl', profile='splunkd', version='current', source_id='')` and `SPLMapper.validate_schema_batch(documents, target)`. They return canonical dictionaries through the real native boundary. Follow existing operation guards, closed-handle behavior, exception conventions, result ownership, and `finally` freeing. No Python matcher or compiler is introduced. Keep packaging manifests complete for all new Go/native sources.

REST adds `POST /api/v1/query/validate-schema` with `{document, target}` and `/api/v1/query/validate-schema/batch` with `{documents, target}`. Targets and their schema resources are inline data, never server filesystem paths or retrieval URLs. Return canonical reports with HTTP 200 for valid/invalid/incomplete query statuses; malformed requests are HTTP 400 and internal errors are server errors. Apply and document body limits; the documented acceptance configuration must admit the genuine normal catalog used for cross-surface tests. Reject an oversized catalog explicitly rather than publish a partial result. A deployed operator may choose a smaller limit, with the normal request-size error behavior.

## Verification and acceptance

Use a shared schema corpus and compare complete canonical report JSON across Go, CLI, actual native Python, and REST for single and ordered mixed-status batch operations. Normalize only JSON object-key order. Retain exact source text, identities, locations, evidence ordering, diagnostics, and compact target metadata. Test strict target-union shapes, query-source conflicts, automation exits, file/stdin/batch-stdin/output behavior, invalid batch atomicity, missing files, and existing field-list and syntax-validation compatibility.

JSON Schema examples cover required and optional nested ancestors, required-without-properties, open/closed/schema-valued additional properties, overlapping explicit/pattern declarations, supported/unrepresentable ECMAScript patterns, allOf closure conflicts, anyOf conditionality, oneOf exclusivity uncertainty, multiple composition keywords and siblings, local `$id`/anchors/pointers, external resource maps, duplicate identities, recursive and non-progressing refs, unsupported dialects and field-affecting keywords, scalar versus object descent, direct array reads versus unsupported array traversal, dotted-name ambiguity, and unbounded wildcard candidates without false finite closure. Include invalid-plus-incomplete precedence and derived/unavailable/unknown-command flows.

OCSF acceptance must include a genuine complete catalog compiled from a pinned official release and a genuine selected-extension variant. Record exact upstream schema/extension revisions, compiler version/revision, generation commands, included extension/version table, content hash, and upstream license/notice. Preserve generated catalog bytes or an explicitly documented reproducible compressed artifact outside `_build_plan/`; test fixtures cannot rely on this temporary design folder. Tiny synthetic catalogs supplement real upstream evidence for collisions, malformed selectors, profile and recursive-object edges, and category-specific counterexamples. Do not present synthetic invented field tables as proof of official-catalog compatibility.

The official fixtures must demonstrate class selection, category member integrity, profile enable/disable behavior, profile qualifications, required/recommended/optional attributes, inherited object paths, json_t openness, array descent incompleteness, exact version/extension-set rejection, category-wide declarations, and supporting/missing/indeterminate class evidence. Add a query whose field is present only in some real category member classes. A local catalog prepared using the documented compiler commands must be directly consumable through every surface without a toolkit-only schema conversion.

Verify kernel partial-universe handling against existing finite field-list refinements: derived candidates remain usable, open schemas cannot create false empty wildcard results, and optional schema facts cannot erase unknown effects. Check caller-data isolation, repeated/concurrent equality, Go race tests, actual native Python single/batch calls, and packaging/build paths. Run acceptance with runtime retrieval unavailable; no schema, OCSF compiler, or account should be contacted. Supported-keyword and array-traversal limitations appear in maintained capability/documentation output. Local checks do not claim an unexecuted platform release matrix.

## Handoff and written-spec gate

This design is a behavior contract, not the implementation task sequence. Before writing the final plan, obtain the controller's verified Milestone 3 commit/handoff, inspect the completed kernel/provider/report/adapter types, and choose the smallest compatible extension. Any schema-driven field refinement remains inside that canonical collaboration; no transfer replay or blanket diagnostic clearing is allowed.

The later plan must be `_build_plan/milestones/4-json-schema-ocsf-validation/milestone-4-implementation-plan.md`. Do not reuse another milestone's implementation-plan basename or subagent-development workspace. The controller must review the completed written spec before planning begins. No production edits, workers, or commits are part of this design task.
