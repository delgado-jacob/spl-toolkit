---
title: "CLI"
layout: page
---

# CLI

The following commands are executed by `tests/acceptance/test_documented_cli.py`. Mapping requires a configuration. Discovery and query validation do not.

<!-- cli-example: map-basic -->
```bash
spl-toolkit map --config testdata/baseline/mappings.json --query 'search src_ip=1'
```

Prints `search source_ip=1`.

<!-- cli-example: discover-json -->
```bash
spl-toolkit discover --query 'search sourcetype=web src_ip=1' --format json
```

<!-- cli-example: validate-query-text -->
```bash
spl-toolkit validate --query 'search src_ip=1'
```

Prints `Valid`.

<!-- cli-example: validate-config-json -->
```bash
spl-toolkit validate --config testdata/baseline/mappings.json --format json
```

Prints `{"target":"configuration","valid":true}`.

<!-- cli-example: discover-output-file -->
```bash
spl-toolkit discover --query 'search src_ip=1' --format json --output discovery.json
```

Successful `--output` writes the result to the file and emits no standard output.

<!-- cli-example: version -->
```bash
spl-toolkit version
```

<!-- cli-example: help -->
```bash
spl-toolkit help
```

<!-- cli-example: demo -->
```bash
spl-toolkit demo
```

`--query` and a positional query are equivalent, but supplying both is a usage error. `--format` accepts `text` or `json`. Exit status 0 means success, 1 means rejected query or configuration content, and 2 means usage or file-access failure. The demo returns nonzero if any demonstrated operation fails.
