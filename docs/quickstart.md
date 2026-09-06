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
