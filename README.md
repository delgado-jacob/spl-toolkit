# SPL Toolkit

SPL Toolkit 0.1.1 is an offline library and command-line tool for bounded operations on supported Splunk SPL queries:

- replace configured field names while preserving the query text around them;
- discover data models, datasets, lookups, macros, sources, sourcetypes, and input fields;
- validate queries against the bundled legacy grammar;
- analyze query flow, located references, lineage, dependencies, and coverage;
- scan dedicated query corpora, export graph/SARIF evidence, compare schema/mapping changes, and serve local editor diagnostics and highlights;
- validate source field obligations against an offline field catalog, singly or in ordered batches;
- preview or apply explicit source-identity rewrites with linked-edit proof, audit trails, and optional destination validation;
- check nested declarations against local JSON Schema resources and exact compiled OCSF versions, keeping optional and category-dependent findings visible.

The Go implementation is canonical. The Python package includes the native Go library, and the REST server calls the same Go APIs. Structured analysis and safe rewriting support bounded SPL and standalone SPL2 contracts under splunkd/current; SPL2 requires explicit selection. This does not provide event instance validation, expression typechecking, raw-to-data-model translation, learned mappings, SPL2 modules, or complete Splunk syntax coverage.

## Developer tooling

`scan`, `graph`, `impact-schema`, `impact-mapping`, `document`, and `lsp --stdio`
reuse the same canonical engine. New corpus/export/impact/document operations
also have native Python and stateless HTTP adapters. See the
[tooling guide and runnable examples](docs/tooling.md) and
[versioned machine contracts](docs/contracts.md). All operations are offline;
impact assessment does not write query sources.

## Safe rewrite

Use `rewrite` with explicit versioned rules for fields, indexes, sources, sourcetypes, lookups, data models, or datasets. Preview returns original `text` plus `candidate_text`; `--apply` returns the candidate only when its syntax, affected bindings and optional destination validation are proved. Explicit aliases stay fixed. An independently safe group may commit beside refused groups, so inspect both `status` and `committed`. Legacy `map` retains its separate configuration and context precedence. See the [rewrite contract and runnable example](docs/rewrite.md).

## Structured analysis

```bash
spl-toolkit analyze --query 'search src=1 | eval a=src, b=a+1 | table b' --format json
spl-toolkit capabilities --format json
```

Analysis emits report format `1` with original source, located references, scope/stage IDs, lineage, dependencies, diagnostics, and explicit coverage. Exit codes are 0 valid, 1 invalid, 3 incomplete, and 2 usage/options/I/O errors. Structural validity does not prove external schema membership or successful Splunk execution. Go, native Python, CLI, and REST share the canonical report.

See the [structured analysis API](docs/API.md) for Go/Python examples, REST routes, exact UTF-8 byte and Unicode column coordinates, supported forms, and migration from flat discovery. See [compatibility](docs/compatibility.md) for the distinction between local analysis verification and released platform evidence.

## Field-list validation

```bash
printf '%s\n' '["host"]' > fields.json
spl-toolkit validate-fields --fields fields.json --query 'eval label=host | table label' --format json
```

Validation follows derived fields, removals, and supported wildcards through canonical query flow. Nested names match exactly; optional declarations are valid without asserting event presence. Reports distinguish valid, invalid, and incomplete coverage, with exits 0, 1, and 3 (2 for request/I/O errors). Go, CLI, native Python, and REST expose identical single/batch reports. See the [validation API](docs/API.md#field-list-validation) for catalog shapes, options, examples, and limitations.

## JSON Schema and OCSF

Check nested fields using your local schema or compiled OCSF catalog. Reports distinguish required, optional, category/branch-dependent and unspecified fields. Open wildcard sets, unsupported patterns/keywords, array descendants, and profile-provenance gaps stay visibly incomplete. Declared arrays are supported; requiredness describes a schema, not event presence. No schema URL is fetched. See the [local-resource CLI tutorial](docs/cli.md#json-schema-with-local-resources), [schema API contract](docs/API.md#json-schema-and-ocsf-field-validation), and [local verification limits](docs/compatibility.md).

## Go

```go
package main

import (
    "fmt"
    "log"

    "github.com/delgado-jacob/spl-toolkit/pkg/mapper"
)

func main() {
    m := mapper.New()
    if err := m.LoadMappings([]byte(`[{"source":"src_ip","target":"source_ip"}]`)); err != nil {
        log.Fatal(err)
    }
    mapped, err := m.MapQuery("search src_ip=1")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(mapped)
}
```

The result is `search source_ip=1`. Mapping is a syntax-aware token rewrite for the supported grammar; it is not a guarantee that an arbitrary rewritten query has the intended semantics.
The same program is available at `examples/go/basic/main.go`.

## Python

Python 3.11 or newer is required. A wheel contains the platform-specific native library.

```python
from spl_toolkit import SPLMapper

with SPLMapper() as mapper:
    mapper.load_mappings([{"source": "src_ip", "target": "source_ip"}])
    print(mapper.map_query("search src_ip=1"))
```

Use a context manager or call `close()`. The package checks that its metadata version matches the bundled native version when it opens the native library.

## CLI

Legacy mapping requires a configuration file; safe rewriting requires a rules file; field-list validation requires a catalog file. Discovery and legacy grammar validation do not. See the executable [CLI usage](docs/cli.md) for commands, output, and exit codes.

```bash
spl-toolkit map --config testdata/baseline/mappings.json --query 'search src_ip=1'
```

## REST

```bash
make build-server
PORT=8080 ./build/spl-toolkit-server
```

The REST service is offline during query processing. Its Swagger page loads assets from a CDN and therefore needs browser network access. Mapping configuration is request-scoped by default; the process-global `/api/v1/mappings` endpoint is disabled unless explicitly enabled for development.

## Discovery limits

`InputFields` is a flat list. It does not describe lineage or distinguish every read/write role. `source` and `sourcetype` keys currently also appear in that list. When unsupported macro syntax causes parsing to fail, discovery can recover macro names but returns the other six categories empty.

## Documentation

- [Quickstart](docs/quickstart.md)
- [Installation](docs/installation.md)
- [Configuration](docs/configuration.md)
- [API surfaces](docs/api/index.md)
- [REST server](docs/api-server.md)
- [Performance baseline](docs/performance.md)

Structured analysis, field-list validation and JSON Schema/OCSF validation support explicitly selected standalone SPL2. See the [SPL2 contract](docs/spl2.md) for syntax, effects and held boundaries, and [compatibility](docs/compatibility.md) for the exact local evidence and remaining platform gates.
