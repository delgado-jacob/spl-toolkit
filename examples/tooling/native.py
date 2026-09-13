"""Run with an installed native SPL Toolkit wheel: python native.py."""
import json
from spl_toolkit import SPLMapper

documents = [{"id": "hosts", "document": {"text": "search host=web | table host"}}]
request = {"schema_version": 1, "documents": documents}
with SPLMapper() as mapper:
    print(json.dumps(mapper.scan_corpus(request), indent=2))
    print(json.dumps(mapper.export_graph(request), indent=2))
    print(json.dumps(mapper.impact_schema({
        **request,
        "before_target": {"kind": "field_list", "catalog": {"fields": ["host"]}},
        "after_target": {"kind": "field_list", "catalog": {"fields": ["server"]}},
    }), indent=2))
