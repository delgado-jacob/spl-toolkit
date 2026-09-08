# Milestone 3 — Field-List Validation Design

Status: approved by the delegated controller on 2026-09-07 and implemented after final Milestone2 acceptance at1f5c986. All five task reviews and the whole-milestone review through c6ae684 are approved; root runtime acceptance passed78964a7 with verified documentation-only ancestry. Final exact-SHA milestone acceptance remains controller-owned. The behavioral contract below is unchanged.

## Purpose and scope

Deliver the smallest complete offline field-validation workflow through canonical Go, CLI, real native Python, and the stateless REST adapter. A caller supplies a finite allowed input-field catalog and receives source-aware findings for a query or ordered query collection. The result distinguishes `valid`, `invalid`, and `incomplete` without confusing external field declaration with structural availability in the pipeline.

This milestone includes exact and nested field paths, statically resolvable query wildcards, explicit optional catalog declarations, file/stdin/batch workflows, human-readable findings, and versioned machine reports. It excludes JSON Schema, OCSF, event-record or field-value validation, expression types, new SPL2 syntax, fixes, suppressions, and network schema retrieval. Existing syntax-only validation remains available unchanged.

The design follows the architectural brainstorming path. The controller approved the catalog interpretation, kernel approach and behavioral boundaries, report/transport/batch contracts, and error/verification section. This file is intentionally in the milestone folder at the controller's request. The controller subsequently released implementation and authorized final design/plan artifact commits. The implementation plan and milestone log record the completed release, review, and verification chain.

## Chosen architecture

Add `pkg/validation` as a consumer of the structured analysis kernel. It owns catalog validation, field membership, per-reference outcomes, report composition, and batch aggregation. The analysis kernel continues owning grammar interpretation, field binding, scope handling, transfer functions, source locations, and semantic uncertainty. Adapters only decode requests, invoke Go, and encode reports. Catalog refinement need not materialize every catalog name in every stage state: preserve the kernel's useful compactness and consult the finite source universe for the requested expansion.

The preferred approach adds a narrow optional finite-source-universe input to the canonical kernel, so supported wildcard operations can resolve catalog-backed names alongside currently available derived names. Plain `analysis.Analyze` keeps its current behavior. The approved narrow hook is `analysis.AnalyzeWithSourceFields(document, fields)`, returning a `SourceAnalysis` with the canonical `Result` and per-reference `FieldExpansion` evidence. Each expansion records its reference ID, completeness, and sorted source/derived matches. The exact signatures and implementation checkpoints are recorded in `milestone-3-implementation-plan.md`; they must be reconciled with the final accepted predecessor before production work.

Two alternatives were considered. Comparing only the original analysis report is simpler, but cannot finish otherwise resolvable flows such as `fields host* | table hostname`. Replaying lineage in the validator could refine those flows, but would duplicate field-transfer semantics. Neither alternative meets the intended shared-meaning boundary as well as the narrow kernel extension.

The kernel must retain initial source openness separately from catalog membership. A missing initial source name is not structurally unavailable. The catalog supplies a finite universe for wildcard refinement, not a replacement rule that makes all absent initial fields unavailable. There is no string-based query reparser, regex field discovery, or adapter-side semantic engine.

## Catalog contract

The Go catalog contains `Fields []string`, `OptionalFields []string`, `Identity string`, and `Version string`. Identity and version are caller metadata, never instructions to fetch anything. The zero-value Go catalog represents an empty allowed set.

The canonical local JSON decoder accepts either an ordinary string array:

```json
["host", "actor.user.name"]
```

or this strict object:

```json
{
  "fields": ["host", "actor.user.name"],
  "optional_fields": ["actor.user.email"],
  "identity": "local-security-events",
  "version": "1"
}
```

The object requires `fields`; the other keys are optional. Omitted `optional_fields` normalizes to `[]`, and omitted metadata to empty strings. Empty arrays are valid closed catalogs. Null values, unknown keys, non-string entries, empty/whitespace-only field names, duplicate names, and overlap between ordinary and optional declarations are input errors. Duplicate and contradictory entries are rejected deterministically rather than silently deduplicated or assigned precedence.

Catalog entries are concrete normalized field names, not SPL fragments or schema patterns. Names retain case and meaningful whitespace; the decoder does not strip quote delimiters, lowercase names, or interpret query syntax inside a catalog string. The query grammar owns normalization of quoted/escaped query references. Arrays are normalized into lexical name order for deterministic reports. The catalog's allowed set is exactly the union of both arrays.

Nested paths use exact comparable dotted names. Declaring `actor` does not implicitly declare `actor.user` or `actor.user.name`; declaring a leaf does not automatically declare its ancestors. No object traversal, array-item inference, or undeclared-property policy is introduced here.

Ordinary entries yield `matching`: the catalog declares the name. Explicit optional entries yield `optional_equivalent`: the catalog declares it as optional. Neither outcome asserts that an actual event contains a field. Ordinary entries do not impose a required-event contract. This explicit metadata preserves the PRD's separate optional-equivalent outcome while leaving string arrays as the easy default.

One catalog applies to the query's source-required event fields across its independently analyzed source scopes. Source-specific catalog selection and lookup dataset schemas are outside this milestone. Lookup-produced derived outputs remain governed by kernel lineage rather than being invented as catalog input requirements.

## Flow, wildcard, and membership rules

Only semantically consuming event-field references produce validation obligations. Use the kernel's binding and role together: writes/creations and non-field references are excluded, as are exclusion/removal selectors. Do not assume every `output` or `rename` reference is a read; their semantic binding determines whether they require an input. Every included reference receives one outcome tied to its original reference ID.

- A source-bound exact read present in ordinary fields is `matching`; presence only in optional fields is `optional_equivalent`; absence is `missing`.
- A provably available derived read is `matching` without requiring a catalog declaration. An upstream missing source still invalidates the report through that upstream reference.
- A kernel-proven unavailable read is `unavailable`, even if its name exists in the external catalog.
- An indeterminate exact binding remains `indeterminate`. Catalog membership cannot turn an unknown command's possible output into a proven source or derived read. A wildcard's single aggregate binding may be indeterminate solely because it combines proven source and derived matches; only complete kernel expansion evidence for every member permits validation to resolve that wildcard conclusively.
- A dynamically computed reference remains indeterminate with a concrete reason.

The finite catalog can refine only grammar-recognized wildcard forms whose field effects the kernel models. Expansion considers the catalog-backed source names still available at that point and currently available derived names. Removed names do not participate. Matching follows the supported SPL selector wildcard semantics; it is not filesystem globbing and does not introduce catalog-side patterns. Unsupported wildcard command forms remain incomplete.

A source-required or inclusion/output wildcard with no matches is `missing` only when the finite catalog and current environment prove the empty result. A kernel-proven structural absence remains `unavailable`. Exclusion/removal selectors with no matches are harmless and create no schema obligation. Wildcard removal still acts on matching names already tracked structurally even when those source obligations are absent from the finite catalog; perform that transfer before validation-candidate filtering. Thus `search missing=x | fields -miss* | where missing=2` against `["host"]` retains the first external missing finding, removes the tracked name, and makes the final read unavailable, without a validation outcome or matching catalog claim for the removal. If the current environment remains uncertain after an unknown-effect stage, dynamic construct, or unresolved scope merge, the wildcard is indeterminate instead of definitely missing. A later exact table/stats projection can restore local membership precision while earlier report coverage limitations remain; conditional per-match bindings must still remain indeterminate. Exclude removals and honor independently proven structural unavailability before treating a complete empty source expansion as missing.

Refinement occurs in the kernel's existing transfer flow, including its supported scope inheritance. It may discharge only the particular wildcard uncertainty that the supplied universe actually resolves. Never clear diagnostic codes globally, force a stage complete, or override aggregate status in a way that erases another unsupported construct. Source locations always cover the original exact reference or wildcard, never a synthesized expanded name.

Approved examples:

| Query and catalog | Outcome |
| --- | --- |
| `search missing=x`, catalog `["host"]` | Source `missing` is missing/invalid, not unavailable. |
| `table no_match*`, complete catalog `["host"]` | Empty source selection is missing/invalid at the wildcard span. |
| `fields -no_match*`, complete catalog `["host"]` | Harmless removal; no missing-field obligation. |
| `fields host \| table user`, catalog `["host","user"]` | `user` is structurally unavailable after projection. |
| `eval derived=host \| table derived`, catalog `["host"]` | Source host and derived read match; derived needs no catalog entry. |
| `fields host* \| table hostname`, catalog `["hostname"]` | Supported expansion and downstream binding are conclusive. |
| `eval label=host \| table lab*`, catalog `["host"]` | The wildcard includes the available derived name label. |
| `mystery \| table missing`, catalog `["host"]` | Unknown effects retain indeterminate/incomplete behavior. |
| `search missing=x \| mystery`, catalog `["host"]` | Definite missing makes invalid; incomplete coverage remains visible. |

## Report contract

`validation.Validate(document analysis.QueryDocument, catalog FieldCatalog) (*Report, error)` returns a report or an input/internal error. `validation.ValidateBatch(documents []analysis.QueryDocument, catalog FieldCatalog) (*BatchReport, error)` returns an ordered batch report or an input/internal error. The chosen kernel refinement hook is reconciled against the final accepted Milestone 2 model before implementation.

A single report has integer `schema_version: 1`, a normalized `target` with `kind: field_list`, `identity`, `version`, `fields`, and `optional_fields`, the refined canonical `analysis` result, `status`, `coverage`, source-ordered `outcomes`, and combined `diagnostics`. Including the canonical analysis retains document text, source identity, references, lineage, scopes, locations, and kernel coverage for downstream consumers. The analysis result remains schema-version 1; external membership findings belong to the validation report rather than mutating analysis-only findings.

Validation coverage contains `syntax_complete`, `semantic_complete`, `schema_complete`, and ordered distinct diagnostic-code `reasons`. Syntax and semantic values reflect refined kernel coverage; schema coverage is false when the complete set of source obligations or an included reference's membership cannot be established. A definite missing or unavailable field can be conclusively classified while the report is invalid. Unsupported semantics continue to prevent an overall valid result.

Each outcome contains `reference_id`, `outcome`, and `matches`. Allowed outcome values are `matching`, `missing`, `unavailable`, `optional_equivalent`, and `indeterminate`. Each match records `name`, `binding` (`source` or `derived`), and its matching/optional-equivalent outcome. Exact matches have one entry; a conclusive wildcard lists all matched names lexically. Missing, unavailable, and unresolved membership do not fabricate matches. A wildcard containing only explicitly optional source matches is optional-equivalent; one containing any ordinary or derived match is matching while individual entries preserve optional distinctions. Reference IDs lead directly to the original source locations in the embedded analysis.

All collections serialize as `[]`, never null. Outcome order follows canonical reference order; matches sort by exact name. Combined diagnostics use the kernel's source-start/code/message ordering. Identical calls, including concurrent calls, return equivalent JSON values. UTF-8 byte offsets, Unicode code-point columns, half-open ranges, and CRLF handling stay kernel-owned.

A batch report has integer `schema_version: 1`, aggregate `status`, and input-ordered `reports`. Each report retains its own document identity. Aggregate precedence is invalid, then incomplete, then valid. An empty batch is an input error. Invalid query text produces an individual syntax-invalid report; malformed batch shape, invalid catalog, unsupported document options, or an internal error rejects the operation without publishing a partial batch report. No parallel execution policy or streaming response protocol is required.

## Diagnostics and failures

Add `SPL_UNKNOWN_FIELD` with error severity and `unknown_field` category for proven missing source requirements. Its explanation identifies the exact name or wildcard and the local catalog decision. Reuse the kernel's `SPL_UNAVAILABLE_FIELD` instead of relabeling structural absence as unknown or duplicating its diagnostic.

Use `SPL_INDETERMINATE_FIELD` with warning severity and `schema_ambiguity` category for unresolved validation obligations, giving the specific reason without discarding the originating unsupported/dynamic diagnostic. Optional declarations are valid outcomes, not warnings. Input errors such as malformed JSON, contradictory declarations, bad document options, or inaccessible files stay outside the query report. Internal errors are not converted to empty successful results.

Any definite query/schema error makes overall status invalid while coverage explains remaining uncertainty. With no definite errors, incomplete query or schema coverage makes status incomplete. Valid requires complete, error-free analysis and validation. The report is not an assertion of successful Splunk execution or event presence.

## CLI

Add `validate-fields` rather than changing the existing `validate` command's syntax/config validation behavior. Require `--fields <local-catalog.json>`. Catalog stdin support is not included, avoiding competing stdin consumers.

Exactly one query source is accepted:

- `--query <text>` or one positional query.
- `--file <path>` for the complete contents of one query file.
- `--stdin` for the complete standard-input query text.
- `--batch <path>` for a JSON array of Query Documents, with `-` selecting stdin.

Single-query input reuses `--language`, `--profile`, `--compatibility-version`, and `--source-id` from analysis. The file path is the default source ID for file input, `<stdin>` for stdin, and an empty source ID for inline text; explicit `--source-id` overrides each. Preserve query bytes without trimming. Batch documents carry their own options and source IDs; batch mode rejects global document/dialect/source-ID flags and other conflicting query inputs.

Support `--format text|json` and `--output`. JSON is the direct canonical single or batch report. Text shows status, completeness reasons, source identity, per-reference outcome and location, wildcard matches when useful, and diagnostics. Content reports are written before returning status: exit 0 valid, 1 invalid, 3 incomplete, and 2 usage/request/I/O/internal errors. Output-file mode does not also print the report to stdout. Failed input or write operations cannot imply a successful content result.

## Go, Python, and REST

Go is the authoritative API; callers can read files or streams into Query Documents using ordinary local I/O. The shared strict catalog decoder provides string-array/object convenience without duplicating parsing in adapters.

Python adds `SPLMapper.validate_fields(query, catalog, *, language='spl', profile='splunkd', version='current', source_id='')` and `SPLMapper.validate_fields_batch(documents, catalog)`, returning canonical dictionaries. Catalog accepts a string list or the documented object. Batch documents use the canonical Query Document shape. Python callers can supply file/stdin contents using normal Python I/O. Add native exports for both operations using the existing owned-result structure and operation guard. Decode results and free native ownership in `finally`; invalid/incomplete queries return reports, request errors raise the established mapper exception, and closed handles behave consistently. Keep package source/native build manifests complete.

REST adds `POST /api/v1/query/validate-fields` with `{document, catalog}` and `POST /api/v1/query/validate-fields/batch` with `{documents, catalog}`. Catalog is inline; server filesystem paths and remote URLs are not accepted as catalog sources. Return HTTP 200 with the canonical report for all query statuses. Invalid request JSON/catalog/options produces HTTP 400 through existing transport errors; internal failures remain server errors. Existing body limits and middleware apply. Document both routes and strict shapes. No adapter performs field matching or transfer logic.

## Verification and acceptance

Add a shared fixture corpus for exact, nested, optional, derived, wildcard, unavailable, and indeterminate outcomes. Cover unknown source names, absent parents versus declared nested leaves, case distinctions, duplicate/contradictory catalogs, empty closed catalogs, and invalid-plus-incomplete precedence. Include quoted/Unicode reference locations and assert exact source slices.

Kernel refinement tests establish the approved examples, inclusion/exclusion distinctions, derived wildcard membership, original Analyze behavior without catalog input, and retention of unrelated uncertainty. Conditional lookup outputs and unsupported branch merges must not become definitely available merely because a similarly named catalog field exists.

Compare complete canonical JSON values across Go, CLI, actual native Python, and REST, normalizing only object-key order. Include single and mixed-status batch inputs. Verify input order and identity, strict malformed-request rejection without partial batch results, all automation exits, file/stdin/batch-stdin contents, output files, missing files, conflicting flags, and unchanged legacy `validate` behavior.

Run relevant Go tests with the race detector, actual native Python tests, and package/native build checks. Test repeated/concurrent validation and caller-data isolation. The runtime uses no network or external account. Local acceptance does not imply a new cross-platform release matrix unless that matrix is actually run.

## Handoff boundary

Before implementation execution, obtain the controller's verified final Milestone 2 commit and handoff and reconcile the chosen kernel hook, final adapter helpers, shared `internal/jsoninput` Unicode validation, and the predecessor's reviewed wildcard-exclusion role correction. Explicit finite-universe refinement accepts nil/empty as a known empty set; plain Analyze means no universe was supplied. Validate/copy names rather than mutating caller slices. Complete expansion means exhaustive membership and proven binding for every match; conditional and unknown-effect membership remains incomplete. Any necessary refinement must remain in the canonical engine; no validator transfer replay or blanket diagnostic clearing is acceptable.

The later plan must be named `milestone-3-implementation-plan.md` in this milestone directory. Do not modify Milestone 2's `implementation-plan.md` or reuse its subagent-development workspace. The controller retains milestone integration and commit authority. The separately authorized implementation plan contains the task sequence and mandatory predecessor checkpoints; neither document starts Milestone 3 production changes.
