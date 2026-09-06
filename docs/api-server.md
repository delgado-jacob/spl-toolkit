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

Requests require `Content-Type: application/json`, reject unknown JSON fields, limit request bodies to 1 MiB, and limit queries to 64 KiB. Structural or configuration errors return 400. Syntax rejection from `/query/validate` returns 422. Successful operations return 200.

Admin mapping updates are disabled by default. Set `ENABLE_ADMIN_ENDPOINTS=true` only for development if process-global mutable mappings are explicitly wanted.

Query processing and the JSON endpoints run offline. The Swagger UI at `/docs` references unpkg CDN assets, so displaying that page needs browser network access. `/openapi.json` remains local.
