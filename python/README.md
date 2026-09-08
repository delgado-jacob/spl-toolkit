# SPL Toolkit Python bindings

This package provides Python 3.11+ bindings for SPL Toolkit 0.1.1's offline mapping, discovery, structured analysis, field-list validation, and JSON Schema/OCSF field declaration APIs. Wheels include the native Go library and do not require Go at installation or runtime.

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

Legacy discovery and mapping retain their own contracts. Flat `input_fields` does not encode flow, scope, or completeness and is not guaranteed to match structured classifications. New consumers should use reference roles/bindings and status/coverage. The [API reference](https://github.com/delgado-jacob/spl-toolkit/blob/main/docs/API.md) describes every surface and the supported forms; this package does not perform event instance validation, broad SPL2 analysis, or new rewriting.

## Field-list validation

```python
from spl_toolkit import SPLMapper

with SPLMapper() as mapper:
    report = mapper.validate_fields(
        "eval label=host | table label", ["host"],
        language="spl", profile="splunkd", version="current",
        source_id="example.spl",
    )
    assert report["status"] == "valid"
    batch = mapper.validate_fields_batch(
        [{"text": "table host", "source_id": "first.spl"},
         {"text": "table missing", "source_id": "second.spl"}],
        {"fields": ["host"], "optional_fields": ["user"],
         "identity": "local-catalog", "version": "1"},
    )
    assert batch["status"] == "invalid"
```

`validate_fields(query, catalog, *, language='spl', profile='splunkd', version='current', source_id='')` returns a dictionary with `schema_version`, `target`, `analysis`, `status`, `coverage`, `outcomes`, and `diagnostics`. `validate_fields_batch(documents, catalog)` takes a nonempty list of document dictionaries with required string `text` and optional `language`, `profile`, `version`, and `source_id`. It returns `schema_version`, `status`, and ordered `reports`; batch status uses invalid, then incomplete, then valid precedence. Schema versions are integer `1` and collections are arrays, including when empty.

A catalog is a string list or an object with required `fields` and optional `optional_fields`, `identity`, and `version`. An empty field list is valid. Field names are case-sensitive and nested names match exactly: a parent or leaf declaration does not imply the other. Empty names, duplicate or overlapping ordinary/optional names, unknown properties, malformed Unicode, and unsupported document options raise `SPLMapperError`. Catalog metadata is preserved; it never triggers a file read or network request. Python only serializes the request and returns the Go report.

Validation follows fields through the query, including created and removed fields and supported wildcards. Reference outcomes are `matching`, `missing`, `unavailable`, `optional_equivalent`, or `indeterminate`; individual matches are `matching` or `optional_equivalent`. Invalid or incomplete query content is a report, not an API error. Inspect `coverage.syntax_complete`, `coverage.semantic_complete`, `coverage.schema_complete`, reasons, and diagnostics. Unsupported semantics remain incomplete and cannot be erased by a later stage. The embedded analysis preserves source metadata and Unicode locations; plain `analyze_query` continues to analyze without a catalog.

Both validation operations use the same context manager and close rules as other mapper methods. Native results are released after decoding, including failures. For direct C callers, `spl_mapper_validate_fields(int mapperID, char* requestJSON)` takes `{"document": {...}, "catalog": ...}` and `spl_mapper_validate_fields_batch` takes `{"documents": [...], "catalog": ...}`. Each returns an owned `SPLResult*` to release with `spl_result_free`, including request/handle errors. Null requests are errors.

Optionality is declaration-only: `optional_equivalent` is valid but says nothing about whether an event contains the field. It cannot repair a structurally removed or conditional field. Each outcome retains every concrete match with its `name`, `binding` (`source` or `derived`), and individual `outcome`; mixed ordinary/optional matches must not be collapsed. A conclusively empty wildcard inclusion is missing, while an empty exclusion is harmless. Unsupported wildcard command forms, dynamic references, and conditional/branch effects remain incomplete. Plain `analyze_query` remains catalog-free.

## JSON Schema and OCSF fields

```python
from spl_toolkit import SPLMapper

target = {"kind": "json_schema", "schema": {"type": "object",
          "properties": {"actor": {"type": "object", "properties": {"name": {"type": "string"}},
                                   "required": ["name"], "additionalProperties": False}},
          "required": ["actor"], "additionalProperties": False}}
with SPLMapper() as mapper:
    report = mapper.validate_schema("table actor.name", target, source_id="first.spl")
    assert report["outcomes"][0]["outcome"] == "required"
    batch = mapper.validate_schema_batch(
        [{"text": "table actor.name", "source_id": "first.spl"},
         {"text": "table absent", "source_id": "second.spl"}], target)
    assert batch["status"] == "invalid"
```

`validate_schema` uses the same keyword-only document defaults as analysis. `validate_schema_batch` accepts a nonempty list of document dictionaries with their own options. Both return full dictionaries with report format integer `1`, original documents, target limitations, located evidence and separate syntax/semantic/schema coverage. Status precedence is invalid, incomplete, valid. Invalid target/document inputs raise `SPLMapperError`; content diagnostics stay in reports. Requests are copied, results are independently owned, and native results are freed on success or error.

Supply JSON Schema object/boolean `schema`, optional `base_uri` and URI-keyed inline `resources`. Draft 2020-12 is the default. For OCSF use `{"kind":"ocsf","catalog":catalog,"selection":{"version":"1.6.0","class":"authentication","profiles":[],"extensions":[]}}`, loading your prepared catalog explicitly with `json.loads(Path("base-catalog.json").read_text())`. The complete compiled extension set must match selection; Windows requires `["win"]`. Choose `category: "iam"` instead of `class` for category evidence, or select `profiles: ["cloud", "datetime"]` explicitly. Preparation uses the official compiler outside the toolkit runtime. No URI is retrieved and no process-global catalog is consulted.

Nested optional ancestors, category/branch-dependent membership, open wildcard sets, unsupported patterns/keywords, array descendants and missing profile provenance remain qualified. A declared array is supported, but descent is incomplete. Requiredness is a schema statement, not event presence. This does not validate JSON event instances or expression types. See the [full schema contract and pinned compiler preparation](https://github.com/delgado-jacob/spl-toolkit/blob/main/docs/API.md#json-schema-and-ocsf-field-validation).
