---
title: "API surfaces"
layout: page
---

# API surfaces

All current surfaces use the canonical Go implementation. Structured analysis, field-list and JSON Schema/OCSF validation, and safe single/batch rewriting are described in the [structured API](../API.md) and [rewrite contract](../rewrite.md). They are distinct from the legacy operations below.

- [Go](go.md) exposes mapping, explicit-context mapping, discovery, validation, and parsing.
- Python exposes `SPLMapper` with `load_mappings`, `map_query`, `map_query_with_context`, `discover_query`, and `get_input_fields`.
- The [REST server](../api-server.md) exposes health, query mapping, discovery, validation, OpenAPI, and optional development-only mapping administration.
- The [CLI](../cli.md) exposes map, discover, validate, version, help, and demo commands.

Python example:

```python
from spl_toolkit import SPLMapper

config = {
    "version": "1.0",
    "mappings": [{"source": "src_ip", "target": "source_ip"}],
}
with SPLMapper(config=config) as mapper:
    mapped = mapper.map_query("search src_ip=1")
    info = mapper.discover_query("search sourcetype=web src_ip=1")
```

Python mapping errors use the existing `SPLMapperError` hierarchy. Invalid configuration uses `ConfigurationError`; parse failures use `ParseError`; operations after `close()` use `MapperNotFoundError`.

Discovery fields use Python names `data_models`, `datasets`, `lookups`, `macros`, `sources`, `source_types`, and `input_fields`. REST and CLI JSON use `datamodels` and `sourcetypes`. The acceptance suite normalizes empty arrays and category order when checking parity; mapped query text remains byte-exact.
