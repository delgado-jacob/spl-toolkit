# SPL Toolkit Python bindings

This package provides Python 3.11+ bindings for SPL Toolkit 0.1.1's offline mapping, discovery, and structured analysis APIs. Wheels include the native Go library and do not require Go at installation or runtime.

```python
from spl_toolkit import SPLMapper

with SPLMapper() as mapper:
    mapper.load_mappings([{"source": "src_ip", "target": "source_ip"}])
    print(mapper.map_query("search src_ip=1"))
```

Use `SPLMapper` as a context manager or call `close()`. Operations after close raise `MapperNotFoundError`. Loading invalid configuration raises `ConfigurationError`, and rejected queries raise `ParseError`. The package verifies that the native and Python package versions match when the native library is opened.

Discovery returns `data_models`, `datasets`, `lookups`, `macros`, `sources`, `source_types`, and flat `input_fields`. Flat fields are not lineage. When unsupported macro syntax prevents normal parsing, discovery may return recovered macro names with the other categories empty.

Source distributions require a local Go toolchain and C compiler. Development installs use built wheels; rebuild and reinstall the wheel after changing the Python wrapper or native Go code.

## Structured analysis

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

Both methods return dictionaries with integer `schema_version: 1`. The keyword-only defaults are `language='spl'`, `profile='splunkd'`, `version='current'`, and `source_id=''`. Empty compatibility options normalize to these defaults; unsupported options raise `SPLMapperError`. Analysis returns `valid`, `invalid`, or `incomplete` reports rather than raising `ParseError` for query diagnostics. Structural validity does not validate an external field catalog or schema. Inspect `coverage.syntax_complete`, `coverage.semantic_complete`, and diagnostics before treating findings as conclusive.

Reports preserve source text and source ID and expose ordered stages/scopes, references, lineage, dependencies, and diagnostics; collections are arrays even when empty. Locations are half-open ranges with zero-based UTF-8 byte offsets and one-based Unicode code-point line/column coordinates. Tabs count as one column; CRLF is one newline. Slice `query.encode("utf-8")[start_offset:end_offset]` for a byte range. Invalid UTF-8 and lone JSON/Python surrogate values raise errors before replacement; paired non-BMP escapes are preserved.

Capabilities separate syntax from semantic support and list limitations and function arities. Unknown commands/functions, unexpanded macros, branch merges, unresolved wildcard membership, and unsupported options remain incomplete. Open-input `fields` inclusion also remains incomplete when retained internal membership is unknown. A later stage cannot erase an earlier coverage gap.

Legacy discovery and mapping retain their own contracts. Flat `input_fields` does not encode flow, scope, or completeness and is not guaranteed to match structured classifications. New consumers should use reference roles/bindings and status/coverage. The [structured analysis API reference](https://github.com/delgado-jacob/spl-toolkit/blob/main/docs/API.md) describes every surface and the supported forms; this package does not perform field-list/schema validation, broad SPL2 analysis, or new rewriting.
