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

Legacy Go `DiscoverQuery`, Python `QueryInfo`/`discover_query`, CLI `discover`, REST discovery, and mapping remain available. Flat `InputFields`/`input_fields` cannot describe read timing, derived fields, scopes, source positions, or coverage. Structured consumers should call analysis, inspect each reference's role/binding and scope, and check status/coverage before treating the result as conclusive. Legacy and structured field classification are distinct contracts; do not infer identical flat field lists or replace mapping behavior based on an analysis report. External field-list/schema validation, broad SPL2, and new rewrite semantics are outside this API.
