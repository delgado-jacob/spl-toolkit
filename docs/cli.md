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

File and stdin query text is retained byte-for-byte, including newlines. Default source identity is the supplied file path, `<stdin>`, or an empty string for inline text. `--source-id` overrides it. Single queries also accept `--language spl`, `--profile splunkd`, and `--compatibility-version current`; unsupported options are input errors. Empty query text produces an invalid content report.

Batches are nonempty JSON arrays of documents, not lines of query text. Each object requires string `text` and permits only string `language`, `profile`, `version`, and `source_id`. Defaults are `spl`, `splunkd`, `current`, and empty identity. Global document options are rejected with `--batch`, including options explicitly set to their defaults. Reports preserve document order, identity, and text. Batch status precedence is invalid, then incomplete, then valid.

JSON output is the full canonical report (schema version 1), including analysis, target catalog, coverage, located diagnostics, and per-reference matching outcomes. Batch JSON contains `schema_version`, `status`, and ordered `reports`. Text output shows status, identity, completeness, located outcomes, matches, and diagnostics. All content reports are written before returning their exit status; `--output` writes only to the selected file.

| Exit | Meaning |
|---|---|
| 0 | Valid field report or all-valid batch |
| 1 | Missing/unavailable field or syntax-invalid content |
| 3 | Incomplete validation |
| 2 | Usage, malformed catalog/document, unsupported options, Unicode, input I/O, or output-write error |
