# SPL Toolkit Python bindings

This package provides Python 3.11+ bindings for SPL Toolkit 0.1.1's offline mapping and discovery APIs. Wheels include the native Go library and do not require Go at installation or runtime.

```python
from spl_toolkit import SPLMapper

with SPLMapper() as mapper:
    mapper.load_mappings([{"source": "src_ip", "target": "source_ip"}])
    print(mapper.map_query("search src_ip=1"))
```

Use `SPLMapper` as a context manager or call `close()`. Operations after close raise `MapperNotFoundError`. Loading invalid configuration raises `ConfigurationError`, and rejected queries raise `ParseError`. The package verifies that the native and Python package versions match when the native library is opened.

Discovery returns `data_models`, `datasets`, `lookups`, `macros`, `sources`, `source_types`, and flat `input_fields`. Flat fields are not lineage. When unsupported macro syntax prevents normal parsing, discovery may return recovered macro names with the other categories empty.

Source distributions require a local Go toolchain and C compiler. Development installs use built wheels; rebuild and reinstall the wheel after changing the Python wrapper or native Go code.
