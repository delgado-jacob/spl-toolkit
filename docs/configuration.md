---
title: "Configuration"
layout: page
---

# Configuration

A mapping configuration is JSON with a required non-empty `version` and a `mappings` array. `name`, `description`, `rules`, and `metadata` are optional.

```json
{
  "version": "1.0",
  "mappings": [
    {"source": "src_ip", "target": "source_ip"},
    {"source": "dst_ip", "target": "destination_ip"}
  ]
}
```

Field mappings contain only `source` and `target`. Empty source or target names are rejected.

## Conditional rules

```json
{
  "version": "1.0",
  "mappings": [{"source": "src_ip", "target": "base_ip"}],
  "rules": [{
    "id": "web",
    "name": "Web logs",
    "description": "Use the web field name",
    "enabled": true,
    "priority": 1,
    "conditions": [{"type": "sourcetype", "operator": "equals", "value": "web"}],
    "mappings": [{"source": "src_ip", "target": "source_ip"}]
  }]
}
```

Enabled rules are evaluated by ascending numeric priority. Equal priorities preserve configuration order. Only the first matching rule contributes mappings, and its mapping overrides the base mapping for the same source field.

Supported conditions are:

| Type | Operators | Value |
|---|---|---|
| `source`, `sourcetype` | `equals`, `contains`, `regex` | string |
| `field_value` | `equals`, `contains`, `regex` | scalar for equals; string otherwise |
| `field_exists` | `exists`, `not_exists` | no value |
| `combination` | `and`, `or` | at least two child conditions |

Source and sourcetype conditions accept a string or a JSON array of strings in explicit context; any matching array item satisfies the condition. Invalid regular expressions and unsupported operators are rejected. `starts_with`, `ends_with`, `not_equals`, numeric comparisons, and unary `not` are unsupported.

The current validation checks required values, supported condition shapes and operators, regex compilation, mapping field names, and the unsupported data-model rewrite field. It does not perform JSON Schema validation, version compatibility negotiation, unique-rule-ID checks, or circular mapping analysis. A non-empty `datamodels` configuration is rejected because data-model rewriting is outside Milestone 1.

Validate a file using the executable case in the [CLI guide](cli.md).
