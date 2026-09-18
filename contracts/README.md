# Machine contracts v1

These JSON Schemas describe serialized canonical Go values and strict request
objects. They validate JSON instances independently of the M4 field-projection
engine. They do not execute SPL, resolve live resources, or prove query equivalence.

Owned schemas use Draft 2020-12. Their absolute GitHub-based `$id` values are stable
registry identities, not retrieval endpoints. Load every file in `v1/` into a
local registry before validating; network retrieval is neither needed nor allowed.
Shared wire definitions reside only in `v1/shared.schema.json`. No competing
runtime schema location existed when these contracts were introduced.

| Named schema | Root instance | Additional entry points under `$defs` |
| --- | --- | --- |
| `query-document.schema.json` | Strict query input; only `text` required | `Normalized` (all five emitted fields) |
| `capabilities.schema.json` | Per-dialect capability manifest | |
| `analysis.schema.json` | Canonical analysis result | |
| `requirements.schema.json` | Canonical direct query requirements | |
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
schemas so archived reports produced before this property existed continue to
validate. Current producers always emit it.

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
The current manifests contain 85 SPL records with 84 evidence cases and 112 SPL2
records with 122 evidence cases. Safe rewriting has 17 supported, 1 unsupported
and 67 unassessed SPL records; SPL2 has 14 supported, 1 unsupported and 97
unassessed records. These cases establish static local toolkit
behavior only, not live Splunk execution, runtime equivalence, environment
compatibility, authorization or upstream support.

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
