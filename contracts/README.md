# Machine contracts

These JSON Schemas describe serialized canonical Go values and strict request
objects. They validate JSON instances independently of the M4 field-projection
engine. They do not execute SPL, resolve live resources, or prove query equivalence.

Owned schemas use Draft 2020-12. Their absolute GitHub-based `$id` values are stable
registry identities, not retrieval endpoints. Load every schema in `v1/` and
`v2/` into a local registry before validating; network retrieval is neither needed
nor allowed. Shared wire definitions reside in `v1/shared.schema.json` and
`v2/shared.schema.json`; v2 references unchanged v1 object and provenance IDs.

| Named schema | Root instance | Additional entry points under `$defs` |
| --- | --- | --- |
| `query-document.schema.json` | Strict query input; only `text` required | `Normalized` (all five emitted fields) |
| `capabilities.schema.json` | Per-dialect capability manifest | |
| `analysis.schema.json` | Canonical analysis result | |
| `requirements.schema.json` | Canonical direct query requirements | |
| `definitions.schema.json` | Caller-supplied definition bundle and collection promises | |
| `closure-request.schema.json` | Strict inline knowledge-object closure request | |
| `closure.schema.json` | Canonical closure report | |
| `closure-graph.schema.json` | Dependency graph projection | |
| `detection-bom.schema.json` | Detection bill of materials projection | |
| `v1/environment-snapshot.schema.json` | Strict offline Snapshot v1 input | |
| `v2/environment-snapshot.schema.json` | Strict offline observed Snapshot v2 input | |
| `v1/environment-export-report.schema.json` | Exporter acquisition report v1 for Snapshot v2 | |
| `field-schema-bundle.schema.json` | Strict independent field evidence input | |
| `compatibility-request.schema.json` | Strict offline assessment envelope | |
| `compatibility.schema.json` | Canonical assessment report, including all four outcomes | |
| `environment-validation.schema.json` | Canonical artifact validation report | `Request` |
| `field-validation.schema.json` | Field-list validation report | `Request`, `BatchRequest`, `BatchReport`, `Catalog` |
| `schema-validation.schema.json` | JSON Schema or OCSF field-validation report | `Request`, `BatchRequest`, `BatchReport`, `Target` |
| `rewrite.schema.json` | Rewrite result | `Request`, `BatchRequest`, `BatchReport`, `RuleSet`, `Rule`, `ValidationTarget` |
| `corpus.schema.json` | Corpus report | `Request`, `Evaluation` |
| `manifest.schema.json` | Local manifest request | |
| `graph.schema.json` | Graph report | |
| `impact.schema.json` | Static comparison report | `SchemaRequest`, `MappingRequest` |
| `document-view.schema.json` | Detached document snapshot | |
| `lsp-configuration.schema.json` | Initialization options/settings object | |

For example, validate a corpus request against
`corpus.schema.json#/$defs/Request`, not against the report root. New reports use
integer `schema_version: 1`; their named schema establishes the report family.
LSP configuration accepts optional `profile`, `version`, and inline `validation_target`.
Omitting the target selects ordinary analysis. The stdio server consumes this object
as initialization options; configuration notifications wrap it under
`settings.splToolkit`. That protocol envelope is outside this schema.

## Compatibility rules

Outputs tolerate additive properties at object extension points, including nested
canonical evidence. They retain required known fields and types. Unknown properties
cannot substitute for a required union member or permit conflicting known members.
Adding a value to an explicitly closed enum, removing required output fields, or
changing their meaning requires a new contract version. Strings without an enum
(for example diagnostic codes and explanatory reasons) remain extensible.
Every member defined by the requirement-set schema is required, and its query and
capability digests use `sha256:` followed by 64 lowercase hexadecimal characters.
The `requirements` property remains optional in the analysis and document-view v1
schemas; current producers always emit it. Current analysis and requirement sets
require `inputs`, `input_coverage`, `field_attribution_coverage`, and `correlation`.
This reviewed pre-release v1 correction also requires item ownership and occurrence
input IDs/necessity. Reanalyze older evidence before assessment; an archived report
without the current input shape does not validate as current evidence. Report
format integers remain 1. Historical receipts retain their original bytes.

The capability schema retains the legacy command/function projections and adds
the evidence ledger. Each record requires syntax, semantics, requirements,
linting and safe-rewriting claims with one of five states: supported, partial,
unsupported, not applicable or unassessed. Summary integers obey three identities:
applicable equals supported plus partial plus unsupported plus unassessed; record
count equals applicable plus not applicable; covered equals supported. Percentages,
weighted totals and composite scores are outside the contract.

Evidence IDs must resolve to typed cases in the same manifest. Stable IDs retain
the exact reviewed scope; a broadened form receives a new ID unless a reviewed
scope correction changes the original boundary. `grammar_registered` is a parser
fact independent of syntax coverage. Linting likewise requires its own evidence.
The current manifests contain 106 SPL records with 107 evidence cases and 142 SPL2
records with 163 evidence cases. SPL requirements have 19 supported, 6 unsupported,
and 81 unassessed records. Safe rewriting has 17 supported, 2 unsupported and 87
unassessed SPL records; SPL2 has 14 supported, 3 unsupported and 125 unassessed
records. The current SPL revision is
`sha256:5b7c15002c426b163a5488d18ae0fb83c68809a194a15f9a6584ea95cbb3a09b`;
the SPL2 revision is
`sha256:7134e06d345f6b2c6e58c3d29c727868320b47ff1fc0a94842aec35615223d9f`.

The SPL ledger includes bounded field-flow semantics for exact `tstats`, selected
field commands, and selected function arities. Exact macros produce direct macro
requirements and source-located unresolved-expansion gaps. Branch children retain
their evidence and direct requirements, while unmodeled merge effects leave parent
outputs uncertain. These cases establish static local toolkit behavior only. The
toolkit does not execute SPL, evaluate regular expressions, compare result rows,
model acceleration, load macro definitions, or certify runtime compatibility. It
also does not prove authorization or upstream support.

Strict canonical and new request objects reject unknown members, wrong types,
null optional objects, and conflicting selectors. Query selectors accept their
existing empty-string defaults (`spl`, `splunkd`, `current`); normalized output
includes every QueryDocument field, including empty opaque `source_id`. Catalog
inputs preserve both the legacy string-array shorthand and the object shape.
Rewrite candidate field-list validation deliberately retains only target
`kind`, `identity`, and `version`; standalone field-list reports retain catalog
arrays. Existing endpoint decoding and wire shapes are not retrofitted or changed.

JSON Schema checks structure, not every decoder/engine invariant. Duplicate JSON
object keys and the lexical difference between `1` and `1.0` are lost in the JSON
data model; actual decoder tests cover those boundaries. Duplicate document/rule
IDs, cross-array catalog overlap, selector preparation, schema compilation and
filesystem containment remain canonical decoder/engine checks. JSON Schema target
payloads and OCSF catalogs are local input data, not a request to retrieve URLs.
Manifest query paths are relative slash-separated names without parent segments;
the base may select a parent before freezing the root. Root paths and their ancestors
must be physical, without symlink/reparse components (use `/private/tmp` on macOS).

Positions retain zero-based UTF-8 byte offsets and one-based Unicode code-point
line/columns. LSP UTF-16 conversion is a separate adapter contract. Graph IDs,
locations and static dependency names are evidence scoped to the report revision,
not persistent catalog identities.

## Closure coverage

The closure request combines a query document with a caller-supplied definition
bundle and optional occurrence bindings. A bundle's `scope_id` identifies the
caller's visibility scope. Per-kind `complete`, `partial`, and `unavailable`
collection values are caller promises; the toolkit cannot verify that the bundle
matches a live Splunk instance. Bindings identify exact original-document
references by SHA256 text digest, kind, half-open UTF-8 byte range, and object
ID. The report retains direct and effective analysis, source-mapped traversal,
graph and BOM projections, and five separate coverage dimensions.
`coverage.complete` requires all dimensions and no gaps or definite invalid
analysis. Macro limits and held forms are in the [API
contract](../docs/API.md#knowledge-object-closure). Graph and BOM schemas
validate projections of the report, not independent discovery or runtime
readiness.

## Environment artifacts

The snapshot names a capture scope, source provenance, interval, capabilities,
per-kind collection promises, and captured objects. The field-schema bundle
holds field-list, JSON Schema, or OCSF entries and explicit object bindings.
Its digest is independent of the snapshot digest, so the same prepared bundle
can be checked against a later capture. `environment-validation.schema.json`
validates a standalone report at its root and an inline request at `#/$defs/Request`.
The request supplies at least one artifact. Computed digests identify normalized
supplied content; they are not signatures. `valid` means the declarations are
internally consistent and have no reported coverage gaps. `partial` retains
omitted, unavailable, partial, or unresolved evidence. Neither status proves
that the capture matches a live Splunk instance or that a query can execute.

The validation request envelope, schema bundle, and validation report keep
`schema_version: 1`; the nested snapshot accepts either version 1 or 2. Snapshot
v1 is unchanged. Snapshot v2 separates observed index/source/sourcetype evidence
from configured knowledge-object inventory. Its observation records selected
indexes, event/metric datatype attempts, exporting-principal visibility, and
bucket-overlap time bounds. Complete observed capture means a missing source or
sourcetype was not observed; an absent binding remains an unresolved warning.
Pre-release pairing correction: conclusive absence from a complete configured
collection emits `binding_object_absent` as a warning and makes the pairing
`partial`, rather than rejecting a structurally valid schema binding. Both known
absence and unresolved absence retain their selected compiled schema targets;
these declarations do not establish object presence. Captured identity mismatch
or contradictory expected identities for the same object ID remain invalid.
Typed field projection distinguishes an atomic dotted name from a structural
path. Flat field lists cannot establish a multi-segment path. Optional field
declarations establish schema membership without proving runtime event presence;
open allowance, conditional membership, and unsupported or exhausted projection
remain indeterminate for compatibility even with complete source coverage.

The exporter acquisition report has its own version 1 schema and embeds the v2
observation with fixed v1 origin/capture references. The
[synthetic partial snapshot](../examples/environment/observed-partial-snapshot.json)
and matching [export report](../examples/environment/export-report.json) explicitly
record unsupported dataset, module, function, and external-command acquisition.
The complete v2 parity fixture supplies all kinds independently; it is not an
initial exporter support claim. JSON Schema checks shape and conditional members;
runtime validation checks cross-references, datatype sets, capture rollups,
provenance intervals, chronological bounds, and canonical digests. Neither shape
success nor runtime consistency proves remote exhaustiveness.

## Offline validation

The acceptance suite uses `jsonschema==4.26.0`, `referencing==0.37.0`, explicit
format checkers and an offline registry whose retrieval function always raises
`NoSuchResource`. It checks every schema with `validator_for` and `check_schema`,
eagerly resolves all resource references/fragments (including unused definitions),
and consumes every `iter_errors` result. The separate
[official SARIF Errata01 schema](sarif/PROVENANCE.md) remains unmodified Draft 4,
with its legacy `id` registered exactly as declared. Schema validation supplements
the existing SARIF URI, source-location, index and consumer tests.

`python/requirements-dev.txt` pins the validator as development dependencies only.
The reviewed local lock
`python/requirements-contracts-local-hashed.lock` has SHA256
`8ab3ebc2e76e8f57c3606952a2fb7c7814e78b9cf1761831af72c3882ea9ceb3`.
Its 23 preflight wheels cover macOS arm64 Python 3.12 and 3.14 only. They do not
establish Python 3.11, Intel, Linux or Windows wheel availability. Install with
`--only-binary=:all:` so an absent wheel fails instead of invoking a Rust build.

```sh
python3 -m venv /private/tmp/spl-contract-validator
/private/tmp/spl-contract-validator/bin/python -m pip install --no-index --only-binary=:all: --find-links=/private/tmp/spl-toolkit-m7-schema-validator-preflight/wheelhouse --require-hashes -r python/requirements-contracts-local-hashed.lock
GOPROXY=off GOSUMDB=off /private/tmp/spl-contract-validator/bin/python -m pytest tests/acceptance/test_machine_contracts.py -q
```

Use `SPL_CONTRACT_GO` to select an absolute Go executable and ordinary Go cache
environment variables for an already populated offline module cache. The suite
compiles a temporary Go emitter from test-owned source, invokes real public APIs,
asserts independently specified query text, statuses, catalog metadata, changes,
counts, capability truth and empty arrays, then validates the emitted instances.
It also exercises durable authored positive/negative fixtures and actual strict
decoders in `testdata/tooling/contracts.json`. It never blesses regenerated output
as an expectation and never needs the native/Python binding or a running server.

## Offline compatibility assessment

`compatibility-request.schema.json` admits current requirements, exactly Snapshot
v1 or v2, query scope and input bindings. SchemaBundlev1 and a query document are
optional; omitted is valid and explicit null is invalid. Nonempty dependency
bindings require a document. Selectors are exactly `all: true` or a nonempty list
of unique values. Requests recursively reject unknown requirement members;
reports retain additive extension points.

Runtime admission also verifies the current capability revision, canonical input
and ownership identities, occurrence/reference/graph links, duplicate or unknown
bindings, expected object identities and schema pairing. JSON Schema cannot detect
raw duplicate keys or enforce these cross-links. Every discovered placeholder
requires one binding; explicit sources may use captured discovery. Selected schema
IDs must exist in the supplied bundle and pair with the selected object's full
expected identity. A paired object may be absent from the snapshot: absence is
assessment evidence, not automatically a malformed request.

Load all v1/v2 IDs offline, including unused reference branches. An assessment
reports `satisfied`, `unsatisfied`, `incomplete`, or `not assessed`; independent
correlation reports `connected`, `disconnected`, `indeterminate`, or
`not applicable`. See the complete generated-envelope recipe in
[the API guide](../docs/API.md#offline-compatibility-assessment).
