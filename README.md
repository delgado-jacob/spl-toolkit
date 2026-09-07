# SPL Toolkit

SPL Toolkit 0.1.1 is an offline library and command-line tool for bounded operations on supported Splunk SPL queries:

- replace configured field names while preserving the query text around them;
- discover data models, datasets, lookups, macros, sources, sourcetypes, and input fields;
- validate queries against the bundled legacy grammar;
- analyze query flow, located references, lineage, dependencies, and coverage.

The Go implementation is canonical. The Python package includes the native Go library, and the REST server calls the same Go APIs. Structured analysis supports the bounded SPL/splunkd/current contract. It does not provide schema validation, raw-to-data-model translation, data-model rewriting, learned mappings, SPL2 modules, or complete Splunk syntax coverage.

## Structured analysis

```bash
spl-toolkit analyze --query 'search src=1 | eval a=src, b=a+1 | table b' --format json
spl-toolkit capabilities --format json
```

Analysis emits report format `1` with original source, located references, scope/stage IDs, lineage, dependencies, diagnostics, and explicit coverage. Exit codes are 0 valid, 1 invalid, 3 incomplete, and 2 usage/options/I/O errors. Structural validity does not prove external schema membership or successful Splunk execution. Go, native Python, CLI, and REST share the canonical report.

See the [structured analysis API](docs/API.md) for Go/Python examples, REST routes, exact UTF-8 byte and Unicode column coordinates, supported forms, and migration from flat discovery. See [compatibility](docs/compatibility.md) for the distinction between local analysis verification and released platform evidence.

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

Mapping requires a configuration file. Discovery and query validation do not. See the executable [CLI usage](docs/cli.md) for commands, output, and exit codes.

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

Later milestones may expand standalone SPL2 parsing and structured analysis. Those capabilities are outside this Milestone 1 release.
