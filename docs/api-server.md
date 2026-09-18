---
title: "REST server"
layout: page
---

# REST server

Build and start the server:

```bash
make build-server
PORT=8080 ./build/spl-toolkit-server
```

Routes are under `/api/v1`:

| Method | Route | Behavior |
|---|---|---|
| GET | `/health` | service and version |
| POST | `/query/map` | request-scoped mapping |
| POST | `/query/discover` | seven-category discovery |
| POST | `/query/validate` | legacy SPL syntax validation |
| POST | `/query/analyze` | canonical SPL/SPL2 analysis |
| POST | `/query/requirements` | canonical direct query requirements |
| GET | `/capabilities` | selected canonical capability manifest |
| POST | `/query/validate-fields` | canonical local field validation |
| POST | `/query/validate-fields/batch` | ordered local field validation batch |
| POST | `/query/validate-schema` | canonical local JSON Schema/OCSF field validation |
| POST | `/query/validate-schema/batch` | ordered schema validation batch |
| POST | `/query/rewrite` | canonical safe rewrite preview/apply report |
| POST | `/query/rewrite/batch` | ordered safe rewrite batch |
| GET | `/openapi.json` | OpenAPI 3.1 document |
| GET | `/docs` | Swagger UI |
| POST | `/mappings` | development-only global configuration; disabled by default |

Map request:

<!-- api-example: map-basic /query/map legacy -->
```json
{
  "query": "search src_ip=1",
  "mappings": [{"source": "src_ip", "target": "source_ip"}]
}
```

Instead of `mappings`, a request may provide a complete `config` and optional `context`. When neither is present, the server retains the existing process-global fallback behavior. The CLI requirement for `--config` does not apply to the Go or REST APIs.

Legacy mapping/discovery/validation accept optional string `language`, `profile`, and `version` selectors, defaulting to `spl`, `splunkd`, and `current`; empty strings select defaults. Unknown values, null selectors, and duplicate selector members are input errors. Explicit `spl2` returns 400 with `unsupported_dialect_for_operation` and guidance to canonical analysis, structured validation, or the rewrite route where the selected capability form is supported. Default legacy response shapes are retained.

Legacy mapping/discovery/validation requests require `Content-Type: application/json`, reject unknown JSON fields, limit request bodies to 1 MiB, and limit queries to 64 KiB. Structural or configuration errors return 400. Syntax rejection from `/query/validate` returns 422. Successful operations return 200.

Admin mapping updates are disabled by default. Set `ENABLE_ADMIN_ENDPOINTS=true` only for development if process-global mutable mappings are explicitly wanted.

Query processing and the JSON endpoints run offline. The Swagger UI at `/docs` references unpkg CDN assets, so displaying that page needs browser network access. `/openapi.json` remains local.

## Canonical analysis and capabilities

`POST /api/v1/query/analyze` accepts exactly one strict query document: required string `text`, optional strings `language`, `profile`, `version`, and `source_id`. Select `spl` (default) or `spl2`, profile `splunkd`, version `current`; empty compatibility selectors use defaults. No dialect inference occurs. Text and identity remain exact, including whitespace, CRLF, and Unicode. Null, duplicate/unknown members, missing text, unsupported selectors, malformed Unicode, and trailing JSON return 400. `Content-Type: application/json` is required; the body limit is exactly 1 MiB, including streamed/unknown-length bodies. One byte over the limit returns 400.

<!-- api-example: analyze-sql /query/analyze valid -->
```json
{"text":"SELECT host FROM main WHERE bytes>0","language":"spl2","profile":"splunkd","version":"current","source_id":"sql-example"}
```

The full report includes SQL lineage `phase` and `execution_order`, with lexical stage/reference locations still referring to the original text. Content status `valid`, `invalid`, or `incomplete` always returns 200; input errors return 400. These examples intentionally exercise invalid and incomplete content:

<!-- api-example: analyze-malformed /query/analyze invalid -->
```json
{"text":"FROM main | eval owner=","language":"spl2"}
```

<!-- api-example: analyze-deferred /query/analyze incomplete -->
```json
{"text":"FROM main | mystery host","language":"spl2"}
```

<!-- api-example: analyze-module /query/analyze invalid -->
```json
{"text":"$q = FROM main;","language":"spl2"}
```

<!-- api-example: analyze-profile /query/analyze invalid -->
```json
{"text":"FROM main | route output","language":"spl2"}
```

Modules/declarations and known wrong-profile constructs carry located `SPL_UNSUPPORTED_MODULE` and `SPL_PROFILE_MISMATCH` findings with incomplete coverage. Unknown standalone syntax remains incomplete.

`POST /api/v1/query/requirements` accepts the same strict query document and returns its canonical direct requirements without environment metadata, knowledge-object expansion, or compatibility proof:

<!-- api-example: requirements-direct /query/requirements valid -->
```json
{"text":"search host=web | table host","source_id":"requirements-example"}
```

Valid, invalid, and incomplete content returns HTTP 200. Input errors return 400.

`GET /api/v1/capabilities?language=spl2&profile=splunkd&version=current` returns the selected manifest, including `documentation_snapshot` for SPL2. The only query parameters are `language`, `profile`, and `version`; omitted or empty values use defaults. Unknown parameters/values, malformed encodings/Unicode, and duplicate or conflicting keys return 400, including repeated equal values. With no selectors it returns the SPL manifest. Query body/source identity are not capability selectors.

The manifest contains five evidence-backed dimensions per record: syntax, semantics, requirements, linting, and safe rewriting. Their states are supported, partial, unsupported, not applicable, and unassessed. Summary counts obey `applicable = supported + partial + unsupported + unassessed`, `records = applicable + not_applicable`, and `covered = supported`; no composite score is emitted. The SPL response contains 68 records and 67 evidence cases. SPL2 contains 98 records and 108 evidence cases. `grammar_registered` is a parser fact, not syntax coverage, and the legacy command/function projections do not replace the ledger. A tagged server reports the exact `VERSION` in `toolkit_version`; source Go execution may report `dev` without changing the semantic capability revision.

Evidence IDs join claims to typed local observations and provenance. Broader forms receive new IDs unless a reviewed scope correction changes the original boundary. The corpus proves static local toolkit behavior only, not live Splunk execution, runtime equivalence, environment compatibility, authorization, or upstream support.

## Local field validation

`POST /api/v1/query/validate-fields` accepts a strict object with exactly `document` and `catalog`. For example:

<!-- api-example: fields-single /query/validate-fields valid -->
```json
{"document":{"text":"search host=web\n","source_id":"query.spl"},"catalog":["host"]}
```

Catalogs accept a string array or an object with required `fields` and optional `optional_fields`, `identity`, and `version`:

<!-- api-example: fields-optional /query/validate-fields valid -->
```json
{"document":{"text":"search user=alice"},"catalog":{"fields":["host"],"optional_fields":["user"],"identity":"web-logs","version":"1"}}
```

`POST /api/v1/query/validate-fields/batch` requires exactly `documents` and `catalog`. Documents form a nonempty array; each supplies its own options and identity:

<!-- api-example: fields-batch /query/validate-fields/batch invalid -->
```json
{"documents":[{"text":"search host=web\n","source_id":"first"},{"text":"search missing=x","language":"spl","profile":"splunkd","version":"current","source_id":"second"}],"catalog":["host"]}
```

Each document requires string `text` and permits only string `language`, `profile`, `version`, and `source_id`. Language is `spl` or `spl2`; profile is `splunkd`; version is `current`. Omitted or empty compatibility selectors default to `spl`, `splunkd`, `current`, and omitted identity defaults to empty. Text and identity are retained without trimming; empty query text produces invalid content. Unsupported options, missing properties, nulls, unknown or duplicate keys, malformed Unicode, and trailing JSON are input errors. Catalog names must be nonempty unique strings without overlap between ordinary and optional fields. Names are case-sensitive and nested names match exactly. Optional-equivalent matches do not assert event presence. Catalog metadata is descriptive: no filename or URL is resolved, and catalogs are isolated per request.

Both endpoints require `Content-Type: application/json` (charset parameters are accepted), enforce a 1 MiB body limit, and preserve existing middleware and method enforcement. They return HTTP 200 with the full canonical report for valid, invalid, and incomplete content. Input errors, including content-type/body-limit errors, return 400; unexpected internal errors return 500. They do not use the legacy `/query/validate` status-422 behavior.

A single report has integer `schema_version: 1`, `target`, `analysis`, `status`, `coverage`, `outcomes`, and `diagnostics`. All arrays are non-null. Outcomes include matching, missing, unavailable, optional_equivalent, and indeterminate, with concrete matches where known. Reports include original document options and identity, half-open UTF-8 byte offsets, and one-based Unicode code-point columns. Batch output has integer `schema_version: 1`, aggregate `status`, and ordered `reports`; status precedence is invalid, then incomplete, then valid. The machine-readable schemas are available at `/api/v1/openapi.json`.

OpenAPI generation uses `make generate-docs` with pinned Swag v2.0.0-rc4 and the pinned PyYAML development dependency from `python/requirements-dev.txt`. The generation postprocessor reconciles strict validation and rewrite input schemas across JSON, YAML, and the served Go template; it supplies strict request-only document, rule, condition, identity, target, and optional capability schemas while preserving report document and normalized OCSF selection schemas.

## JSON Schema and OCSF requests

`POST /api/v1/query/validate-schema` takes exactly `document` and `target`. The batch route `/api/v1/query/validate-schema/batch` takes exactly `documents` (nonempty) and `target`:

<!-- api-example: schema-single /query/validate-schema incomplete -->
```json
{"document":{"text":"table host","source_id":"single.spl"},"target":{"kind":"json_schema","schema":{"properties":{"host":{}},"required":["host"],"additionalProperties":false}}}
```

<!-- api-example: schema-batch /query/validate-schema/batch invalid -->
```json
{"documents":[{"text":"table host","source_id":"first.spl"},{"text":"table missing","source_id":"second.spl"}],"target":{"kind":"json_schema","schema":{"properties":{"host":{}},"additionalProperties":false}}}
```

Both schema routes require JSON content type and allow bodies up to exactly 8 MiB (8,388,608 bytes); 8,388,609 bytes returns 400, including streamed/unknown-length bodies. Existing field-list/legacy limits stay 1 MiB. Genuine base and Windows OCSF catalogs fit without removing tables. Each request supplies its own inline raw catalog; no global catalog registry or URL/file retrieval is available.

HTTP 200 contains the full canonical valid, invalid or incomplete report. HTTP 400 means a malformed target/document, unsupported dialect/selection/compile version, duplicate JSON keys, mixed/unknown wrapper members, nulls, malformed Unicode, invalid content type or oversized body. A batch input error is atomic and has no partial reports. Unexpected internal errors return 500. An unresolved local reference is a valid request yielding incomplete content, not an input error. See the [schema API](API.md#json-schema-and-ocsf-field-validation) for target shapes, optional/category-dependent fields, locations and evidence.

## Explicit SPL2 validation examples

<!-- api-example: fields-spl2 /query/validate-fields valid -->
```json
{"document":{"text":"FROM main | table host\n","language":"spl2","profile":"splunkd","version":"current","source_id":"pipeline"},"catalog":["host","bytes"]}
```

<!-- api-example: fields-mixed /query/validate-fields/batch incomplete -->
```json
{"documents":[{"text":"table host","source_id":"legacy"},{"text":"FROM main | table host","language":"spl2","source_id":"spl2"},{"text":"FROM main | mystery host","language":"spl2","source_id":"deferred"}],"catalog":["host","bytes"]}
```

<!-- api-example: schema-spl2 /query/validate-schema valid -->
```json
{"document":{"text":"FROM main SELECT host","language":"spl2","profile":"splunkd","version":"current","source_id":"schema-example"},"target":{"kind":"json_schema","schema":{"type":"object","properties":{"host":{"type":"string"}},"additionalProperties":false}}}
```

Both schema routes use the same per-document language contract and their existing strict 8 MiB decoders. JSON Schema and OCSF targets are separate request union branches. OCSF selection requires the exact catalog version and exactly one class/class_uid/category/category_uid, with unique profiles/extensions arrays; omitted arrays normalize to empty, but null is rejected. Request selectors are separate from the report's normalized selection, which includes resolved class/category data. Canonical reports preserve null-test obligations, schema evidence, and unresolved dotted wildcard paths without adapter inference. See [canonical CLI usage](cli.md) for file/stdin and mixed-language batch equivalents.

## Safe rewrite requests

`POST /api/v1/query/rewrite` and `POST /api/v1/query/rewrite/batch` accept strict, versioned JSON with inline rules and an optional inline field-list, JSON Schema, or OCSF validation target. Preview is the default; `mode: "apply"` returns candidate text only when the canonical rewrite proof permits it. Filesystem paths, remote retrieval URLs, legacy mapper configuration, and global batch document selectors are not part of these routes. Both routes use the exact 8 MiB schema-body policy and return complete query reports with HTTP 200 for valid, invalid, and incomplete statuses.

<!-- api-example: rewrite-preview /query/rewrite valid -->
```json
{"schema_version":1,"document":{"text":"search src=alice | table src","source_id":"preview.spl"},"rules":[{"id":"rename-src","kind":"field","source":{"name":"src"},"target":{"name":"user"}}]}
```

<!-- api-example: rewrite-batch-apply /query/rewrite/batch incomplete -->
```json
{"schema_version":1,"mode":"apply","documents":[{"text":"search src=alice","source_id":"safe.spl"},{"text":"| mystery src","source_id":"partial.spl"}],"rules":[{"id":"rename-src","kind":"field","source":{"name":"src"},"target":{"name":"user"}}]}
```

An exit or report status of `incomplete` is not itself a refusal: individual reports can have `committed: true` when their candidate is proven, while the ordered batch aggregate remains incomplete. The selected capability manifest exposes rewrite support per kind, role, and identity form; clients must inspect those optional entries rather than assume universal rewriting support.
