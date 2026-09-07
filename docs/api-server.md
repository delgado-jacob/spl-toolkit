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
| POST | `/query/validate` | syntax validation |
| POST | `/query/validate-fields` | canonical local field validation |
| POST | `/query/validate-fields/batch` | ordered local field validation batch |
| GET | `/openapi.json` | OpenAPI 3.1 document |
| GET | `/docs` | Swagger UI |
| POST | `/mappings` | development-only global configuration; disabled by default |

Map request:

```json
{
  "query": "search src_ip=1",
  "mappings": [{"source": "src_ip", "target": "source_ip"}]
}
```

Instead of `mappings`, a request may provide a complete `config` and optional `context`. When neither is present, the server retains the existing process-global fallback behavior. The CLI requirement for `--config` does not apply to the Go or REST APIs.

Legacy mapping/discovery/validation requests require `Content-Type: application/json`, reject unknown JSON fields, limit request bodies to 1 MiB, and limit queries to 64 KiB. Structural or configuration errors return 400. Syntax rejection from `/query/validate` returns 422. Successful operations return 200.

Admin mapping updates are disabled by default. Set `ENABLE_ADMIN_ENDPOINTS=true` only for development if process-global mutable mappings are explicitly wanted.

Query processing and the JSON endpoints run offline. The Swagger UI at `/docs` references unpkg CDN assets, so displaying that page needs browser network access. `/openapi.json` remains local.

## Local field validation

`POST /api/v1/query/validate-fields` accepts a strict object with exactly `document` and `catalog`. For example:

```json
{"document":{"text":"search host=web\n","source_id":"query.spl"},"catalog":["host"]}
```

Catalogs accept a string array or an object with required `fields` and optional `optional_fields`, `identity`, and `version`:

```json
{"document":{"text":"search user=alice"},"catalog":{"fields":["host"],"optional_fields":["user"],"identity":"web-logs","version":"1"}}
```

`POST /api/v1/query/validate-fields/batch` requires exactly `documents` and `catalog`. Documents form a nonempty array; each supplies its own options and identity:

```json
{"documents":[{"text":"search host=web\n","source_id":"first"},{"text":"search missing=x","language":"spl","profile":"splunkd","version":"current","source_id":"second"}],"catalog":["host"]}
```

Each document requires string `text` and permits only string `language`, `profile`, `version`, and `source_id`. Omitted options default to `spl`, `splunkd`, `current`, and empty identity. Text and identity are retained without trimming; empty query text produces invalid content. Unsupported options, missing properties, nulls, unknown or duplicate keys, malformed Unicode, and trailing JSON are input errors. Catalog names must be nonempty unique strings without overlap between ordinary and optional fields. Names are case-sensitive and nested names match exactly. Optional-equivalent matches do not assert event presence. Catalog metadata is descriptive: no filename or URL is resolved, and catalogs are isolated per request.

Both endpoints require `Content-Type: application/json` (charset parameters are accepted), enforce a 1 MiB body limit, and preserve existing middleware and method enforcement. They return HTTP 200 with the full canonical report for valid, invalid, and incomplete content. Input errors, including content-type/body-limit errors, return 400; unexpected internal errors return 500. They do not use the legacy `/query/validate` status-422 behavior.

A single report has integer `schema_version: 1`, `target`, `analysis`, `status`, `coverage`, `outcomes`, and `diagnostics`. All arrays are non-null. Outcomes include matching, missing, unavailable, optional_equivalent, and indeterminate, with concrete matches where known. Reports include original document options and identity, half-open UTF-8 byte offsets, and one-based Unicode code-point columns. Batch output has integer `schema_version: 1`, aggregate `status`, and ordered `reports`; status precedence is invalid, then incomplete, then valid. The machine-readable schemas are available at `/api/v1/openapi.json`.

OpenAPI generation uses `make generate-docs` with pinned Swag v2.0.0-rc4 and the pinned PyYAML development dependency from `python/requirements-dev.txt`. The generation postprocessor reconciles strict validation input schemas across JSON, YAML, and the served Go template; it preserves the shared analysis document schema.
