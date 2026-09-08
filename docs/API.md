---
title: "Structured analysis API"
layout: page
---

# Structured analysis API

Structured analysis describes query stages, nested scopes, located references, field availability, lineage, dependencies, and incomplete understanding. Processing is deterministic and offline. Go is canonical; the CLI, native Python library, and REST API expose the same report values.

## Go

```go
package main

import (
    "encoding/json"
    "log"
    "os"

    "github.com/delgado-jacob/spl-toolkit/pkg/analysis"
)

func main() {
    report, err := analysis.Analyze(analysis.QueryDocument{
        Text: "search src=1 | eval a=src, b=a+1 | table b",
        SourceID: "example.spl",
    })
    if err != nil {
        log.Fatal(err)
    }
    if err := json.NewEncoder(os.Stdout).Encode(report); err != nil {
        log.Fatal(err)
    }
}
```

`analysis.Analyze(document) (*analysis.Result, error)` returns query diagnostics inside the report. Invalid options or invalid source encoding return an error. `analysis.Capabilities()` returns a fresh `analysis.CapabilityManifest`. Public types contain no parser internals.

## CLI

See [CLI usage](cli.md) for command conventions and existing operations.

```bash
spl-toolkit analyze --query 'search src=1 | eval a=src, b=a+1 | table b' --source-id example.spl --format json
spl-toolkit analyze --query 'search src=1 | fields - src | where src>1' --format json
spl-toolkit analyze --query 'search src=1 | mystery x' --format json
spl-toolkit capabilities --format json
```

These analysis examples return `valid`/0, `invalid`/1, and `incomplete`/3 respectively. The invalid example reports `SPL_UNAVAILABLE_FIELD`; the incomplete example preserves the `src` finding and reports `SPL_UNSUPPORTED_COMMAND`.

| Exit | Meaning for `analyze` |
|---|---|
| 0 | Valid within the selected analysis contract |
| 1 | Invalid syntax or a proven unavailable field |
| 3 | Incomplete syntax/semantic coverage without a proven error |
| 2 | Usage, options, or I/O error |

The report is written before returning its content status. `--output report.json` writes to a file. `--format text` presents status, coverage, located references, and diagnostics; `--format json` emits the direct canonical report. A positional query is also supported. `--language spl`, `--profile splunkd`, and `--compatibility-version current` select the only delivered contract. `capabilities` supports text/JSON and `--output`, with exit 0 on success and 2 on usage/I/O errors. Existing commands retain their exit behavior.

## Python

```python
from spl_toolkit import SPLMapper

with SPLMapper() as mapper:
    report = mapper.analyze_query(
        "search src=1 | eval a=src, b=a+1 | table b",
        language="spl", profile="splunkd", version="current",
        source_id="example.spl",
    )
    print(report["status"])
    capabilities = mapper.capabilities()
```

Both methods return dictionaries. `analyze_query` accepts keyword-only `language='spl'`, `profile='splunkd'`, `version='current'`, and `source_id=''`. Invalid/incomplete queries return reports; invalid document options or encoding raise `SPLMapperError`. Operations on a closed mapper raise `MapperNotFoundError`. Use a context manager or call `close()`; the wrapper frees every owned native result even when decoding fails.

## REST

Start the server with `PORT=8080 ./build/spl-toolkit-server`. Send the Query Document directly:

```bash
curl -sS http://localhost:8080/api/v1/query/analyze \
  -H 'Content-Type: application/json' \
  -d '{"text":"search src=1 | eval a=src, b=a+1 | table b","source_id":"example.spl"}'
curl -sS http://localhost:8080/api/v1/capabilities
```

`POST /api/v1/query/analyze` returns HTTP 200 with the direct report for all three query statuses. The following is an excerpt of the first response; the complete response also contains stages, scopes, references, lineage, dependencies, and diagnostics:

```json
{
  "schema_version": 1,
  "document": {
    "text": "search src=1 | eval a=src, b=a+1 | table b",
    "language": "spl",
    "profile": "splunkd",
    "version": "current",
    "source_id": "example.spl"
  },
  "status": "valid",
  "coverage": {"syntax_complete": true, "semantic_complete": true, "reasons": []}
}
```

`GET /api/v1/capabilities` returns the direct capability manifest. Malformed JSON, invalid Unicode, and unsupported options return HTTP 400 transport errors instead of analysis reports. Existing JSON content-type, 1 MiB request-body limit, and middleware protections apply. See [REST server usage](api-server.md) for deployment and legacy endpoints.

## Document, positions, and report format

Query Documents have `text`, `language`, `profile`, `version`, and `source_id`. Empty compatibility options normalize to `spl`, `splunkd`, and `current`; source ID defaults to an empty string. `version` selects the compatibility contract, while integer `schema_version: 1` identifies the machine report format. This is separate from the package version.

Successful reports preserve original query text and source ID, including spaces, tabs, CRLF, Unicode, and embedded NUL supported by JSON/native APIs. Invalid UTF-8 or unpaired JSON UTF-16 surrogate escapes are rejected before lossy replacement. Valid paired non-BMP escapes and escaped literal backslashes are accepted. CLI arguments cannot carry embedded NUL.

Locations use half-open `[start, end)` ranges into the original text. `offset` is a zero-based UTF-8 byte offset; `line` and `column` are one-based, with columns counting Unicode code points, not UTF-16 units or display cells. Tabs count as one column. CRLF counts as one newline; lone CR and LF each start a new line. Slice the UTF-8 bytes to recover `original_name`; do not use byte offsets as Python string indices.

| Report field | Meaning |
|---|---|
| `stages`, `scopes` | Ordered command/scope records with deterministic IDs and locations |
| `references` | Original and normalized names, kind, role, stage/scope IDs, location, resolution, binding, and origin reference IDs |
| `lineage` | Per-stage before/after field state and ordered transitions; fields include conditionality and origins |
| `dependencies` | Indexes, sources, sourcetypes, datasets, lookups, data models, and macros |
| `diagnostics` | Stable code, severity, category, message, location, stage ID, and scope ID |
| `coverage` | Separate syntax/semantic completeness and reason codes |

Collections are arrays, including empty arrays, never null. Names preserve field case; command/function matching is case-insensitive. IDs and collection order are deterministic for the same document. Reference binding distinguishes source requirements, derived values, indeterminate origins, and non-consuming operands. Removal references describe operations, not required source inputs.

Implicit search stages use command `search`; a macro-only stage uses synthetic command `macro`. Macro identity is a located dependency, with unexpanded effects incomplete. Data-model and dataset references may overlap: `datamodel:Web.All_Traffic` identifies both the root model and qualified dataset. In `datamodel Web All_Traffic`, the dataset's source range covers `All_Traffic` while its normalized name is `Web.All_Traffic`. Consumers must retain the distinct reference identities and component spans.

## Coverage and limitations

`valid` means no proven structural error and complete analysis for the supported forms. It does not assert that fields exist in an external schema, that a dependency exists on a Splunk instance, or that Splunk will execute a query successfully. `invalid` takes precedence over incomplete coverage when syntax errors or provably unavailable fields occur. Partial trustworthy findings survive neighboring unsupported or damaged stages; later supported stages cannot erase earlier coverage gaps. Read both `status` and `coverage`.

The capability manifest has integer `schema_version: 1`, `language`, `profile`, `version`, `commands`, and `functions`. Each entry has `name`, `syntax_supported`, `semantic_supported`, and `limitations`. Semantic support applies to the listed forms, not every option of that command. Use the runtime manifest for exact function arities and supported contexts.

- Search predicates treat bare right-hand values as literals; `where`/`eval` expression identifiers are reads. Index/source/sourcetype selectors are dependencies.
- Eval assignments resolve left-to-right, with self-assignment reading the prior binding. Ordinary non-overlapping rename sources resolve before the stage. Chains, swaps, duplicate claims, and collisions remain incomplete.
- Exact removals preserve tombstones. Table/stats close the known field set. Fields inclusion retains known internal fields; open-input unknown internal membership remains incomplete. Wildcards require provable membership; quoting a projection selector does not make `*` literal. Quoted expression field identifiers can be exact.
- Stats uses registered aggregates and exact grouping fields. Implicit output names include `count` and canonical names such as `sum(bytes)`. Eventstats/streamstats add outputs; options, window behavior, and wildcard grouping remain unmodeled.
- Lookup models explicit inputs/outputs; catalog aliases are not event-field reads. OUTPUTNEW preserves known bindings and records conditional outputs. Inputlookup introduces an open source without inventing catalog columns.
- Sort/dedup support exact field lists and numeric limits; head/tail support optional numeric limits. Unsupported options and dynamic/unknown functions remain incomplete.
- Ordinary child searches, join, and append use independent environments. Appendpipe inherits a copy. Child outputs do not leak into parent state; branch merging remains incomplete.
- Datamodel/from/tstats expose grammatical dependencies while field effects remain incomplete. Unknown commands are opaque; macros and dynamic query semantics are unresolved. No default complete handler is assumed.

Stable diagnostic codes are `SPL_SYNTAX_ERROR`, `SPL_UNAVAILABLE_FIELD`, `SPL_UNSUPPORTED_COMMAND`, `SPL_UNSUPPORTED_SEMANTICS`, `SPL_UNSUPPORTED_FUNCTION`, `SPL_DYNAMIC_REFERENCE`, and `SPL_UNRESOLVED_WILDCARD`.

## Migrating from discovery

Legacy Go `DiscoverQuery`, Python `QueryInfo`/`discover_query`, CLI `discover`, REST discovery, and mapping remain available. Flat `InputFields`/`input_fields` cannot describe read timing, derived fields, scopes, source positions, or coverage. Structured consumers should call analysis, inspect each reference's role/binding and scope, and check status/coverage before treating the result as conclusive. Legacy and structured field classification are distinct contracts; do not infer identical flat field lists or replace mapping behavior based on an analysis report. Use field-list validation below for external declarations. JSON Schema validation, broad SPL2, and new rewrite semantics remain outside this API.

## Field-list validation

Use `pkg/validation` to validate source obligations against an offline declaration of concrete fields. It follows the canonical query flow: derived names need no catalog entry, removed fields remain unavailable, and supported selectors expand against the fields available at that stage. Ordinary consumers should call validation rather than reconstructing obligations from flat discovery or replaying field transfers.

```go
package main

import (
    "fmt"
    "log"

    "github.com/delgado-jacob/spl-toolkit/pkg/analysis"
    "github.com/delgado-jacob/spl-toolkit/pkg/validation"
)

func main() {
    catalog := validation.FieldCatalog{
        Fields: []string{"host"}, OptionalFields: []string{"user"},
        Identity: "local-fields", Version: "1",
    }
    report, err := validation.Validate(analysis.QueryDocument{
        Text: "eval label=host | table label", SourceID: "example.spl",
    }, catalog)
    if err != nil { log.Fatal(err) }
    fmt.Println(report.Status) // valid
    batch, err := validation.ValidateBatch([]analysis.QueryDocument{
        {Text: "table host", SourceID: "good.spl"},
        {Text: "table missing", SourceID: "missing.spl"},
        {Text: "| mystery | table host", SourceID: "unknown.spl"},
    }, catalog)
    if err != nil { log.Fatal(err) }
    fmt.Println(batch.Status) // invalid; reports retain input order
}
```

`validation.Validate(document, catalog) (*validation.Report, error)` and `validation.ValidateBatch(documents, catalog) (*validation.BatchReport, error)` return content findings in reports and request failures as errors. JSON callers use `DecodeFieldCatalog`, `DecodeRequest`, `DecodeBatchRequest`, or `DecodeDocuments` for strict decoding. A catalog is either a JSON string array such as `["host"]` or an object:

```json
{"fields":["host","user.name"],"optional_fields":["hostname"],"identity":"local-fields","version":"1"}
```

Object `fields` is required. `optional_fields` defaults to `[]`; `identity` and `version` default to empty strings. Names are concrete, nonblank, case-sensitive UTF-8 strings; duplicates, ordinary/optional overlap, nulls, unknown properties, duplicate JSON keys, trailing JSON, invalid Unicode, and wrong value types are rejected. Catalog order normalizes deterministically. An empty catalog is useful and valid. Metadata is retained in `target`; it does not resolve a file, URL, or remote schema. A declaration of `user` does not declare `user.name`, and a declaration of `user.name` does not declare `user`.

Optionality is declaration-only: an optional declared source is `optional_equivalent`, which is valid. It does not establish event presence, conditional query availability, or a required property. Structurally removed fields remain unavailable even if declared optional. Created fields are derived and need no catalog declaration. Types, actual events, JSON Schema, and OCSF are outside this feature.

### CLI single and batch inputs

Create local catalog and batch files, then invoke the CLI:

```bash
printf '%s\n' '["host"]' > fields.json
spl-toolkit validate-fields --fields fields.json --query 'eval label=host | table label' --source-id example.spl --format json
printf '%s\n' '[{"text":"table host","source_id":"good.spl"},{"text":"table missing","source_id":"missing.spl"},{"text":"| mystery | table host","source_id":"unknown.spl"}]' > queries.json
spl-toolkit validate-fields --fields fields.json --batch queries.json --format json --output reports.json
```

The single example exits 0; the batch exits 1 after writing ordered valid/invalid/incomplete reports. Exactly one positional query, `--query`, `--file path`, `--stdin`, or `--batch path` is required. `--batch -` reads the JSON document array from stdin. `--fields` always names a local file and cannot be `-`. `--format text|json` and `--output path` work for either mode. Reports are emitted even when invalid or incomplete. Exit codes are 0 valid, 1 invalid, 3 incomplete, and 2 request/options/I/O errors.

Single-query options are `--language spl`, `--profile splunkd`, `--compatibility-version current`, and `--source-id ID`. File input preserves all bytes and defaults source identity to the supplied path; stdin defaults to `<stdin>`; inline queries default to an empty source ID. An explicit `--source-id` overrides either default, including an empty value. Align source IDs when comparing file and stdin reports. Batch documents supply their own options and source IDs; global document options are rejected in batch mode. The legacy `validate` command retains its grammar-only behavior.

### Python single and batch calls

```python
from spl_toolkit import SPLMapper

with SPLMapper() as mapper:
    report = mapper.validate_fields(
        "eval label=host | table label", ["host"],
        language="spl", profile="splunkd", version="current", source_id="example.spl",
    )
    batch = mapper.validate_fields_batch(
        [{"text": "table host", "source_id": "good.spl"},
         {"text": "table missing", "source_id": "missing.spl"}],
        {"fields": ["host"], "optional_fields": ["user"]},
    )
```

Both return the direct report dictionary. Single-call compatibility options are keyword-only and default as shown; `source_id` defaults to `''`. Batch documents are dictionaries with required string `text` and optional string `language`, `profile`, `version`, and `source_id`. Invalid requests raise `SPLMapperError`; invalid/incomplete content returns a report. Context-manager/close and native-result ownership rules also apply to validation.

### REST single and batch requests

```bash
curl -sS http://localhost:8080/api/v1/query/validate-fields \
  -H 'Content-Type: application/json' \
  -d '{"document":{"text":"eval label=host | table label","source_id":"example.spl"},"catalog":["host"]}'
curl -sS http://localhost:8080/api/v1/query/validate-fields/batch \
  -H 'Content-Type: application/json' \
  -d '{"documents":[{"text":"table host","source_id":"good.spl"},{"text":"table missing","source_id":"missing.spl"}],"catalog":{"fields":["host"],"optional_fields":["user"]}}'
```

The endpoints return HTTP 200 for all content statuses. Requests require `document` or a nonempty `documents` array plus `catalog`; unknown or malformed properties and unsupported document options return HTTP 400. Invalid batch requests produce no partial reports. Existing content-type and body-size protections apply. Every document preserves its original text and identity; empty compatibility strings normalize to `spl` / `splunkd` / `current`.

### Validation reports and finite-source evidence

A single report has integer `schema_version: 1`, `target` (`kind: "field_list"` plus normalized catalog), embedded `analysis`, overall `status`, `coverage`, `outcomes`, and `diagnostics`. Batch reports have integer `schema_version: 1`, overall `status`, and ordered `reports`. Status precedence is invalid, then incomplete, then valid, both within a query and across a batch. Definite missing, unavailable, or syntax errors take precedence without hiding incomplete coverage.

Each outcome contains `reference_id`, `outcome`, and ordered `matches`. Outcome values are `matching`, `missing`, `unavailable`, `optional_equivalent`, or `indeterminate`. Each match retains concrete `name`, `binding` (`source` or `derived`), and its own `outcome` (`matching` or `optional_equivalent`); preserve every match even when the aggregate outcome is matching. Non-consuming removals do not create catalog obligations. A conclusively empty inclusion is missing; an empty exclusion is harmless. All arrays remain arrays when empty. IDs, messages, source locations, and array order are canonical, using the UTF-8 positions described above.

Coverage separately reports `syntax_complete`, `semantic_complete`, and `schema_complete`, with reasons. Supported projection/removal wildcards refine downstream canonical lineage as well as matches. Unsupported wildcard command forms (including wildcard rename and aggregate/grouping forms), unknown commands/functions, dynamic references, macros, uncertain conditional outputs, and unresolved branch behavior retain incomplete coverage. Optional declarations cannot resolve conditional availability.

Advanced Go integrations can call `analysis.AnalyzeWithSourceFields(document, fields) (*analysis.SourceAnalysis, error)`. This uses a finite universe of concrete source names; nil and empty both mean a known empty universe. It returns `Result` (`analysis` in JSON) and ordered `Expansions` (`expansions`), where each entry contains `reference_id`, `complete`, and `matches` with concrete `name` and `binding`. `complete: true` with no matches is conclusive empty membership; `complete: false` retains uncertainty. Consumers must preserve completeness and all per-match evidence. This hook refines canonical transfer and lineage but does not classify catalog optionality. Plain `analysis.Analyze` and its wire format remain unchanged. Call validation for ordinary field-catalog decisions.

## JSON Schema and OCSF field validation

Check nested field declarations, optional ancestors, and fields that depend on a schema branch or OCSF category member. Supply local JSON Schema resources or an exact compiled OCSF version. Reports retain uncertainty for open wildcard sets, unsupported patterns/keywords, array descendants, and missing profile provenance. This is field declaration projection, not complete JSON Schema instance validation: no events are checked, expressions typechecked, or queries executed. `required` is a schema statement, not evidence that an event contains the field. A declared array itself is supported; descending into it is incomplete.

Go exposes `validation.ValidateSchema(document, target)` and `validation.ValidateSchemaBatch(documents, target)`, returning `*SchemaReport` and `*SchemaBatchReport` plus an error. Targets are explicitly supplied values; Go has no network access or global catalog registry. `analysis.Analyze` and the existing analysis capability output remain unchanged. Each prepared target owns its copied inputs and immutable indexes; calls share no mutable projection memo state.

```go
target := validation.SchemaTarget{
    Kind: "json_schema",
    Schema: json.RawMessage(`{"properties":{"host":{}},"required":["host"],"additionalProperties":false}`),
}
report, err := validation.ValidateSchema(analysis.QueryDocument{Text: "table host"}, target)
```

A JSON Schema target has `kind: "json_schema"`, required inline object/boolean `schema`, optional `identity`, `base_uri`, and URI-keyed inline `resources`. Draft 2020-12 is the default; explicit unsupported dialects are input errors. `$id`, local `$ref`, JSON Pointer fragments and anchors resolve only in these supplied resources. An HTTPS URI is an identity, never an instruction to retrieve. Missing local references return located incomplete results. There is no implicit file lookup, URL dereference, schema compiler, account, or registry.

Outcomes are `required`, `optional`, `permitted_unspecified`, `conditional`, `indeterminate`, `missing`, `unavailable`, and `matching` for derived fields. Nested requiredness needs every ancestor to be required; an optional ancestor makes its descendant optional. `additionalProperties` can permit an unspecified name without enumerating the source universe. `allOf` intersects permissions; prohibited syntactic property names are excluded. `anyOf` membership disagreement is conditional with both positive and negative evidence; `oneOf` viability remains indeterminate. Derived fields bypass external declarations. Removal selectors incur no schema obligations, and a schema declaration cannot restore a structurally removed field.

Single reports preserve `analysis`, `target`, `status`, `coverage`, `outcomes`, and located `diagnostics`, with integer `schema_version: 1`. Batches have `schema_version`, aggregate `status`, and ordered `reports`. Invalid wins over incomplete without erasing coverage gaps. `target.limitations` describes the target's capabilities; it does not itself make every query incomplete. Consult per-reference evidence and all three coverage flags. Known partial wildcard matches do not establish exhaustive expansion. A later complete output stage does not erase earlier semantic incompleteness.

For `table actor.name`, the reference spans UTF-8 offsets `[6,16)` and one-based columns `[7,17)`. An outcome links its `reference_id` to that analysis reference. Evidence identifies the URI, JSON Pointer, keyword, declaration basis, and requiredness, for example:

```json
{"resource_uri":"https://schemas.example.test/user","pointer":"/$defs/user/properties/name","keyword":"properties","declaration_basis":"property","requirement":"required"}
```

This is a compact evidence illustration; the full report also carries ancestor evidence from the event resource. An unresolved target may include `"limitations":["array_traversal","partial_name_universe","unresolved_ref"]`. The CLI guide has the [complete executable local-resource example](cli.md#json-schema-with-local-resources).

### OCSF preparation and selection

Use normal official compiler output with `compile_version: 1`. Pin schema commit `d0cd8a0fef198bf93086044e33fda6d80c74b9ef` (OCSF 1.6.0) and compiler commit `d6b0b781d51a6b9682ea99396636ae01437b41b4`. Compiler preparation used Python 3.14.1 and compiler version `0.0.0-dev`; toolkit runtime floors remain Go 1.22+ and Python 3.11+. Preparation is an explicit separate operation:

```sh
git clone https://github.com/ocsf/ocsf-schema.git
git -C ocsf-schema checkout --detach d0cd8a0fef198bf93086044e33fda6d80c74b9ef
git clone https://github.com/ocsf/ocsf-schema-compiler.git
git -C ocsf-schema-compiler checkout --detach d6b0b781d51a6b9682ea99396636ae01437b41b4
SCHEMA="$(pwd)/ocsf-schema"
cd ocsf-schema-compiler/src
python3.14 -m ocsf_schema_compiler "$SCHEMA" --ignore-platform-extensions > base-catalog.json
python3.14 -m ocsf_schema_compiler "$SCHEMA" --ignore-platform-extensions --extensions-path "$SCHEMA/extensions/windows" > windows-catalog.json
```

The compiler default includes platform extensions; use `--ignore-platform-extensions` for the base catalog. Do not prune normal classes/objects/tables to fit an API payload. The pinned catalog hashes, source archive identities, compiler commands and upstream license/notices live in `testdata/schemas/ocsf/1.6.0/provenance.json` and its neighboring notice files. Raw base SHA-256 is `9b609f8fb670772f04191c1c276b46d34d6e9110d2417c71fa89c4f54c585137`; raw Windows is `19af77ce259f3ff57debc33e52da51b8220a1af59399d06553f29629678595e9`.

An OCSF target has exactly `kind: "ocsf"`, inline `catalog`, `selection`, and optional `identity`. Selection requires exact `version` and exactly one concrete `class`, `class_uid`, `category`, or `category_uid`. Abstract/base classes are not concrete selections. `profiles` and `extensions` are unique case-sensitive arrays; omission means empty and null is invalid. Select the full compiled extension set, not just extensions thought relevant to a query. The Windows fixture requires `["win"]`, UID 2, version 1.6.0. Version and extension identity are not inferred from query values.

Using the [CLI conventions](cli.md):

```bash
spl-toolkit validate-schema --ocsf-catalog base-catalog.json --ocsf-version 1.6.0 --ocsf-class authentication --query 'table time actor.user.name unmapped.vendor_field' --format json
spl-toolkit validate-schema --ocsf-catalog base-catalog.json --ocsf-version 1.6.0 --ocsf-category iam --query 'table time group' --format json
spl-toolkit validate-schema --ocsf-catalog base-catalog.json --ocsf-version 1.6.0 --ocsf-class authentication --ocsf-profile cloud --ocsf-profile datetime --query 'table cloud time_dt' --format json
spl-toolkit validate-schema --ocsf-catalog windows-catalog.json --ocsf-version 1.6.0 --ocsf-class win/registry_key_activity --ocsf-extension win --query 'table time' --format json
```

A category result includes supporting, missing, and indeterminate classes. Generic `unmapped`/`xattributes` objects permit unspecified descendants where the compiled structure proves that behavior. Profile inclusion and merged strongest requiredness lose some source provenance: partial enabling-profile selection and unresolved inherited profile membership remain qualified. No finite/exhaustive claim is made for open, recursive, patterned, or provenance-limited universes.

### Pattern subset and bounds

The supported regex subset is ASCII literals, `^` only at the beginning, `$` only at the end, positive ASCII character classes and ranges, escaped punctuation from `\^$.*+?()[]{}|/-`, and single-atom `?`, `*`, `+`, `{n}`, `{n,}`, `{n,m}` quantifiers. Counts are at most 1000 with no leading zeros except `0`; a finite maximum must be at least its minimum. Matching is unanchored unless anchors are present. Bare `.`, groups, alternation, negated classes, shorthand/Unicode escapes, lookaround, backreferences, lazy quantifiers, non-ASCII patterns, and non-ASCII/CR/LF candidate names are outside this subset and yield uncertainty. Malformed supported patterns are input errors. This narrow gate avoids claiming Go regex semantics are full ECMAScript semantics.

Projection is bounded to 4096 visited states and 128 path segments; candidate enumeration is bounded to 4096 work/name units. Exhaustion yields incomplete evidence rather than an absent-field claim. Recursive references without path progress are incomplete; supplied references that consume path segments can progress within the bounds. Array descendants remain incomplete regardless of a declared `items` schema.

### Schema diagnostic reasons

`SPL_UNKNOWN_FIELD` is a proven missing declaration; `SPL_INDETERMINATE_FIELD` is a located schema-ambiguity warning. Existing `SPL_UNAVAILABLE_FIELD` and syntax/semantic diagnostics are preserved. `coverage.reasons` carries diagnostic codes. The following stable reason identifiers appear in target limitations or reference evidence, not as replacement diagnostic codes:

| Reason | Meaning |
|---|---|
| `alternative_branches` | Alternative branches disagree on membership. |
| `array_traversal` | Array traversal is outside field projection. |
| `conditional_schema` | Conditional field constraints are unsupported. |
| `dependent_required` | Dependent requiredness is unsupported. |
| `dependent_schema` | Dependent schemas are unsupported. |
| `dynamic_ref` | Dynamic reference evaluation is unsupported. |
| `enumeration_budget` | Candidate enumeration reached its work or name bound. |
| `exclusive_branches` | Exclusive branch viability is unresolved. |
| `literal_path_collision` | Literal dotted and nested names collide. |
| `negation` | Negated field constraints are unsupported. |
| `object_shape_unknown` | Requiredness depends on an unproved object shape. |
| `object_value_constraint` | Object value and cardinality constraints are unsupported. |
| `ocsf_constraint` | A relevant compiled OCSF constraint form is unsupported. |
| `ocsf_profile_inheritance` | Selected profile inheritance cannot be resolved from compiled provenance. |
| `ocsf_profile_requirement` | A strict subset of enabling profiles cannot establish merged requiredness. |
| `partial_name_universe` | The source name universe is not exhaustive. |
| `pattern_properties` | Patterns can admit unenumerated names. |
| `property_names` | Property name constraints are unsupported. |
| `recursive_ref` | A reference cycle made no path progress. |
| `required_vocabulary` | A required custom vocabulary is unsupported. |
| `traversal_budget` | The requested path exceeded the projection state or segment bound. |
| `unevaluated_properties` | Unevaluated property tracking is unsupported. |
| `unrepresentable_source_name` | An admitted blank source name cannot be represented by the canonical source universe. |
| `unresolved_ref` | The reference has no supplied local schema target. |
| `unsupported_pattern` | The pattern is outside the supported ASCII subset. |
