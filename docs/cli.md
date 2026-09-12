---
title: "CLI"
layout: page
---

# CLI

The following commands are executed by `tests/acceptance/test_documented_cli.py`. Mapping requires a configuration. Discovery and query validation do not.

<!-- cli-example: map-basic -->
```bash
spl-toolkit map --config testdata/baseline/mappings.json --query 'search src_ip=1'
```

Prints `search source_ip=1`.

<!-- cli-example: discover-json -->
```bash
spl-toolkit discover --query 'search sourcetype=web src_ip=1' --format json
```

<!-- cli-example: validate-query-text -->
```bash
spl-toolkit validate --query 'search src_ip=1'
```

Prints `Valid`.

<!-- cli-example: validate-config-json -->
```bash
spl-toolkit validate --config testdata/baseline/mappings.json --format json
```

Prints `{"target":"configuration","valid":true}`.

<!-- cli-example: discover-output-file -->
```bash
spl-toolkit discover --query 'search src_ip=1' --format json --output discovery.json
```

Successful `--output` writes the result to the file and emits no standard output.

<!-- cli-example: version -->
```bash
spl-toolkit version
```

<!-- cli-example: help -->
```bash
spl-toolkit help
```

<!-- cli-example: demo -->
```bash
spl-toolkit demo
```

`--query` and a positional query are equivalent, but supplying both is a usage error. `--format` accepts `text` or `json`. Exit status 0 means success, 1 means rejected query or configuration content, and 2 means usage or file-access failure. The demo returns nonzero if any demonstrated operation fails.

## Explicit query dialects

`analyze`, `capabilities`, `validate-fields`, and `validate-schema` select `--language spl` (default) or `--language spl2`, with `--profile splunkd` and `--compatibility-version current`. Omitted or explicitly empty compatibility selectors use these defaults. Unknown nonempty values, invalid Unicode, and duplicate flags are input errors (exit 2). The tool never guesses the dialect from query syntax.

`analyze` accepts positional or `--query` text and `--source-id ID`; `capabilities` takes selectors without a query or source identity. JSON analysis includes the exact source, located references and diagnostics, coverage, dependencies, and lineage. SQL lineage includes `phase` and `execution_order` without changing lexical source locations. SPL2 capabilities include the pinned `documentation_snapshot` identity.

<!-- cli-example: spl2-analysis -->
```bash
spl-toolkit analyze --language spl2 --profile splunkd --compatibility-version current --source-id sql-example --query 'SELECT host FROM main WHERE bytes>0' --format json
```

Exit 0. SQL phases run in source/filter/evaluate/project order while references retain their original byte locations.

<!-- cli-example: spl2-capabilities -->
```bash
spl-toolkit capabilities --language spl2 --profile splunkd --compatibility-version current --format json --output spl2-capabilities.json
```

Exit 0; stdout is empty and the file contains the complete selected manifest. Text output also includes the documentation snapshot when present.

Content reports use exit 0 for valid, 1 for invalid, and 3 for incomplete. Malformed supported SPL2 forms remain invalid. Modules/declarations and wrong-profile constructs report `SPL_UNSUPPORTED_MODULE` and `SPL_PROFILE_MISMATCH` with incomplete coverage; unknown or deferred forms remain incomplete. Input errors write to stderr and emit no report. `--output FILE` preserves the full report even for invalid or incomplete content.

Legacy `map`, `discover`, and `validate` accept only the default SPL compatibility contract. Explicit `--language spl2` returns exit 2 with `unsupported_dialect_for_operation` and guidance to `analyze` or structured validation. Use `rewrite` for local, rule-driven SPL or SPL2 rewrites, and inspect `capabilities` for the selected dialect's supported rewrite forms and limitations; support is not universal. Default legacy success and error formats are retained.

## Local field validation

`validate-fields` checks the canonical field obligations against a local catalog. It requires `--fields FILE` and exactly one positional/`--query` value, `--file FILE`, `--stdin`, or `--batch FILE`. The legacy `validate` command retains its syntax/configuration behavior.

Create a catalog as a JSON string array, such as `["host", "bytes"]`, or use the object form for optional fields and metadata:

```json
{"fields":["host","bytes"],"optional_fields":["user"],"identity":"web-logs","version":"1"}
```

`fields` is required in the object form; `optional_fields`, `identity`, and `version` are optional. Names are nonempty, unique, case-sensitive, and cannot occur in both arrays. Nested names match exactly: a parent or leaf does not imply another path. Optional fields produce optional-equivalent matches, not event-presence guarantees. Metadata is descriptive and never instructs a file read or URL fetch. Catalogs reject nulls, unknown or duplicate object keys, malformed Unicode, trailing JSON, and non-string entries. `--fields -` is rejected.

Save the array catalog as `fields.json` and the object catalog above as `catalog.json`.

<!-- cli-example: fields-inline -->
```bash
spl-toolkit validate-fields --fields fields.json 'search host=web'
```

<!-- cli-example: fields-object-json -->
```bash
spl-toolkit validate-fields --fields catalog.json --query 'search user=alice' --format json
```

<!-- cli-example: fields-file-output -->
```bash
printf 'search host=web\n' > query.spl
spl-toolkit validate-fields --fields fields.json --file query.spl --source-id query.spl --output report.json --format json
```

<!-- cli-example: fields-stdin -->
```bash
printf 'search host=web\n' | spl-toolkit validate-fields --fields fields.json --stdin --source-id pipeline
```

<!-- cli-example: fields-batch-file -->
```bash
printf '%s\n' '[{"text":"search host=web","source_id":"one"},{"text":"search missing=x","source_id":"two"}]' > queries.json
spl-toolkit validate-fields --fields fields.json --batch queries.json --format json
```

<!-- cli-example: fields-batch-stdin -->
```bash
cat queries.json | spl-toolkit validate-fields --fields fields.json --batch - --format json
```

File and stdin query text is retained byte-for-byte, including newlines. Default source identity is the supplied file path, `<stdin>`, or an empty string for inline text. `--source-id` overrides it. Single queries also accept `--language spl` or `--language spl2`, `--profile splunkd`, and `--compatibility-version current`; unsupported options are input errors. Empty query text produces an invalid content report.

Batches are nonempty JSON arrays of documents, not lines of query text. Each object requires string `text` and permits only string `language`, `profile`, `version`, and `source_id`. Defaults are `spl`, `splunkd`, `current`, and empty identity. Global document options are rejected with `--batch`, including options explicitly set to their defaults. Reports preserve document order, identity, and text. Batch status precedence is invalid, then incomplete, then valid.

JSON output is the full canonical report (schema version 1), including analysis, target catalog, coverage, located diagnostics, and per-reference matching outcomes. Batch JSON contains `schema_version`, `status`, and ordered `reports`. Text output shows status, identity, completeness, located outcomes, matches, and diagnostics. All content reports are written before returning their exit status; `--output` writes only to the selected file.

## Safe rewrite

`rewrite --rules FILE` reads a local versioned `RuleSet` containing only `schema_version` and `rules`. It previews by default; `--apply` requests a verified candidate. A rules file cannot select apply mode. The command accepts the shared positional, `--query`, `--file`, `--stdin`, and `--batch` sources. It never edits a source file in place. Add at most one optional `--fields`, `--schema`, or `--ocsf-catalog` validation target family.

<!-- cli-example: rewrite-preview -->
```bash
spl-toolkit rewrite --rules rules.json --query 'search src=alice'
```

<!-- cli-example: rewrite-stdin-apply -->
```bash
printf 'search src=alice\n' | spl-toolkit rewrite --rules rules.json --stdin --source-id pipeline --apply
```

<!-- cli-example: rewrite-output-file -->
```bash
spl-toolkit rewrite --rules rules.json --batch queries.json --apply --output rewrite-report.txt
```

Text output labels original, candidate, and returned text separately. A preview can show an applied candidate while returning the original text. Reports are written before exits 1 or 3; exit 3 denotes incomplete proof and does not by itself mean every document was refused or uncommitted. With `--output`, stdout remains empty.

| Exit | Meaning |
|---|---|
| 0 | Valid report or all-valid batch |
| 1 | Invalid content, including failed destination validation |
| 3 | Incomplete analysis or rewrite proof; some batch reports may still be committed |
| 2 | Usage, malformed rules/target/document, unsupported options, Unicode, input I/O, or output-write error |

## JSON Schema with local resources

After `make build`, the local binary is `build/spl-toolkit`; `export PATH="$PWD/build:$PATH"` makes the following commands available.

`validate-schema` checks nested field declarations using exactly one `--schema FILE` or `--ocsf-catalog FILE`. Inputs are explicit local files; `-` is not accepted for target/resource paths. Query sources are positional/`--query`, `--file`, `--stdin`, or `--batch FILE` (including `--batch -`). Single document flags are `--language spl` or `--language spl2`, `--profile splunkd`, `--compatibility-version current`, and `--source-id ID`. Batch objects carry their own options; global document flags are rejected in batch mode.

Create `event.schema.json` with this complete content:

```json
{"$id":"https://schemas.example.test/event","type":"object","properties":{"actor":{"$ref":"user#/$defs/user"}},"required":["actor"],"additionalProperties":false}
```

Create `resources.json` with this complete content:

```json
{"https://schemas.example.test/user":{"$defs":{"user":{"type":"object","properties":{"name":{"type":"string"}},"required":["name"],"additionalProperties":false}}}}
```

<!-- cli-example: schema-local-resources -->
```bash
spl-toolkit validate-schema --schema event.schema.json --schema-resources resources.json --query 'table actor.name' --format json
```

Exit 0; `actor.name` is required, with evidence from both local resources. `required` describes schema declarations, not event presence. `--schema-base-uri URI` supplies a base for relative references when needed. Neither HTTPS identities nor missing references trigger retrieval.

<!-- cli-example: schema-unresolved-resource -->
```bash
spl-toolkit validate-schema --schema event.schema.json --query 'table actor.name' --format json
```

Exit 3; the unresolved reference remains incomplete. The report locates the affected reference and retains its schema evidence.

Create `queries.json`:

```json
[{"text":"table actor.name","source_id":"first.spl"},{"text":"table absent","source_id":"second.spl"}]
```

<!-- cli-example: schema-local-batch -->
```bash
spl-toolkit validate-schema --schema event.schema.json --schema-resources resources.json --batch queries.json --format json
```

Exit 1; the batch preserves both documents in order and reports the closed missing field. Valid is exit 0, invalid is 1, incomplete is 3, and usage/target/I/O failure is 2. A content report is emitted before its exit status; failed requests do not emit partial reports. Text output includes target limitations, locations and evidence; JSON is the full canonical report. `--output FILE` writes reports to that path.

For OCSF, pass exact `--ocsf-version` and exactly one `--ocsf-class` or `--ocsf-category` (key or decimal UID). Repeat `--ocsf-profile` and `--ocsf-extension` for selections; duplicate names are rejected. These schema profiles are separate from query compatibility `--profile splunkd`. The [OCSF preparation and selection guide](API.md#ocsf-preparation-and-selection) includes pinned compiler commands and executable authentication, IAM category, cloud/datetime, and Windows examples. Selection must equal the entire compiled extension set.

Optional ancestors and category/branch dependence stay visible. Declared arrays are supported; descendants, open wildcard sets, unsupported patterns/keywords and profile-provenance gaps can remain incomplete. This is not event validation or expression typechecking. See the [schema API contract](API.md#json-schema-and-ocsf-field-validation) for exact pattern support and traversal bounds.

## SPL2 validation and mixed batches

Save `fields.json` as `["host","bytes"]`. Feed the following command the exact stdin text `FROM main | table host` followed by a newline:

<!-- cli-example: spl2-fields-stdin -->
```bash
spl-toolkit validate-fields --fields fields.json --stdin --language spl2 --profile splunkd --compatibility-version current --source-id pipeline
```

Exit 0. File and stdin bytes, including CRLF and Unicode, remain unchanged; the source identity is exact caller data. An intact `isnull`/`isnotnull` inspection does not create a required-existence outcome, while a separate consuming read still does. Schema wildcard expansion retains proven flat/derived members while ambiguous dotted paths remain incomplete.

Feed this exact JSON array to the next command's stdin:

```json
[{"text":"table host","source_id":"legacy"},{"text":"FROM main | table host","language":"spl2","profile":"splunkd","version":"current","source_id":"spl2"},{"text":"FROM main | mystery host","language":"spl2","source_id":"deferred"}]
```

<!-- cli-example: mixed-fields-batch-stdin -->
```bash
spl-toolkit validate-fields --fields fields.json --batch - --format json
```

Exit 3. The three reports stay in input order with their own dialects and source identities. Global selectors, even explicitly empty ones, are input errors in batch mode.

Save `host.schema.json` as `{"type":"object","properties":{"host":{"type":"string"}},"additionalProperties":false}`:

<!-- cli-example: spl2-schema -->
```bash
spl-toolkit validate-schema --schema host.schema.json --language spl2 --profile splunkd --compatibility-version current --source-id schema-example --query 'FROM main SELECT host' --format json
```

Exit 0; the JSON report retains the selected language and schema declaration evidence.

## Verification

Build the CLI with `make build`, then run the maintained examples with the existing stdin-aware harness:

```bash
SPL_CLI="$PWD/build/spl-toolkit" SPL_DOCS_ROOT="$PWD" python -m pytest tests/acceptance/test_documented_cli.py -q
```
