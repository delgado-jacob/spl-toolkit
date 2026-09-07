---
title: "Quickstart"
layout: page
---

# Quickstart

Build and test the Go surfaces:

```bash
make deps
make test
make build-all
```

The CLI commands are maintained as executable cases in the [CLI guide](cli.md).

```bash
spl-toolkit map --config testdata/baseline/mappings.json --query 'search src_ip=1'
```

In Go, check configuration and operation errors:

```go
m := mapper.New()
if err := m.LoadMappings([]byte(`[{"source":"src_ip","target":"source_ip"}]`)); err != nil {
    return err
}
mapped, err := m.MapQuery("search src_ip=1")
```

In Python, close each mapper deterministically:

```python
from spl_toolkit import SPLMapper

with SPLMapper() as mapper:
    mapper.load_mappings([{"source": "src_ip", "target": "source_ip"}])
    mapped = mapper.map_query("search src_ip=1")
```

Conditional rules require `enabled: true`. Lower numeric priorities run first, equal priorities keep configuration order, and only the first matching rule contributes mappings. Its mapping overrides a base mapping for the same source field.

```json
{
  "version": "1.0",
  "mappings": [{"source": "src_ip", "target": "base_ip"}],
  "rules": [{
    "id": "web",
    "enabled": true,
    "priority": 1,
    "conditions": [{"type": "sourcetype", "operator": "equals", "value": "web"}],
    "mappings": [{"source": "src_ip", "target": "source_ip"}]
  }]
}
```

The CLI extracts source and sourcetype context from the query. Go, Python, and REST can also accept explicit context through their mapping APIs.

## Validate a field catalog

```bash
printf '%s\n' '["host"]' > fields.json
spl-toolkit validate-fields --fields fields.json --query 'eval label=host | table label' --format json
printf '%s\n' '[{"text":"table host","source_id":"good.spl"},{"text":"table missing","source_id":"missing.spl"}]' > queries.json
spl-toolkit validate-fields --fields fields.json --batch queries.json --format json --output reports.json
```

The first query is valid because `label` is derived from declared `host`; it needs no separate catalog declaration. The batch emits both reports and exits 1 for the missing field. Valid exits 0, invalid exits 1, incomplete exits 3, and request/I/O errors exit 2. The [field-list API](API.md#field-list-validation) includes Go, Python, and REST single/batch workflows, catalog metadata and optional fields, source identities, and exact report semantics. Validation checks declarations, not event presence or types.
