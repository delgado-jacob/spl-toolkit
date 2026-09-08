#!/usr/bin/env python3
"""Reconcile pinned Swag output with the strict canonical validation decoders.

Swag cannot express strict M3/M4 inputs or inline JSON unions through Go
DTOs. Keep generated report schemas and the M2 document schema unchanged.
Requires the repository's pinned development PyYAML dependency.
"""
import copy
import json
from pathlib import Path
import re
import sys

import yaml


PREFIX = "#/components/schemas/"


def update(directory: Path) -> None:
    json_path, yaml_path, go_path = (directory / name for name in ("swagger.json", "swagger.yaml", "docs.go"))
    spec = json.loads(json_path.read_text(encoding="utf-8"))
    yaml_spec = yaml.safe_load(yaml_path.read_text(encoding="utf-8"))
    template = go_path.read_text(encoding="utf-8")
    matches = list(re.finditer(r'^    "components": (\{.*\}),$', template, re.MULTILINE))
    if len(matches) != 1 or json.loads(matches[0][1]) != spec["components"] or yaml_spec != spec:
        raise ValueError("unexpected pinned Swag output: JSON, YAML, and Go template must agree")
    if spec.get("openapi") != "3.1.0":
        raise ValueError("expected pinned OpenAPI 3.1.0 output")
    schemas = spec["components"]["schemas"]

    def shape(name, keys):
        schema = schemas[name]
        if schema.get("type") != "object" or set(schema["properties"]) != set(keys):
            raise ValueError(f"unexpected pinned schema shape: {name}")
        return schema

    document = copy.deepcopy(shape("analysis.QueryDocument", ("text", "language", "profile", "version", "source_id")))
    if any(value != {"type": "string"} for value in document["properties"].values()):
        raise ValueError("unexpected pinned query document property shape")
    document.update(required=["text"], additionalProperties=False)
    schemas["validation.QueryDocument"] = document
    catalog = shape("validation.FieldCatalog", ("fields", "optional_fields", "identity", "version"))
    for key in ("identity", "version"):
        if catalog["properties"][key] != {"type": "string"}:
            raise ValueError(f"unexpected pinned catalog metadata shape: {key}")
    for key in ("fields", "optional_fields"):
        field = catalog["properties"][key]
        if field.get("type") != "array" or field.get("items") not in ({"type": "string"}, {"type": "string", "minLength": 1}):
            raise ValueError(f"unexpected pinned catalog names shape: {key}")
    catalog.update(required=["fields"], additionalProperties=False)
    names = {"type": "array", "items": {"type": "string", "minLength": 1}, "uniqueItems": True}
    catalog["properties"]["fields"] = copy.deepcopy(names)
    catalog["properties"]["optional_fields"] = copy.deepcopy(names)
    union = {"oneOf": [names, {"$ref": PREFIX + "validation.FieldCatalog"}]}
    for name, key in (("validation.Request", "document"), ("validation.BatchRequest", "documents")):
        request = shape(name, (key, "catalog"))
        old_catalog = request["properties"]["catalog"]
        if old_catalog not in ({"$ref": PREFIX + "validation.FieldCatalog"}, union):
            raise ValueError(f"unexpected pinned catalog reference: {name}")
        doc_ref = {"$ref": PREFIX + "validation.QueryDocument"}
        old_doc_ref = {"$ref": PREFIX + "analysis.QueryDocument"}
        value = doc_ref if key == "document" else {"type": "array", "items": doc_ref, "minItems": 1}
        current = request["properties"][key]
        # Swag emits uniqueItems:false on slice fields; it is not a constraint.
        unpatched = old_doc_ref if key == "document" else {"type": "array", "items": old_doc_ref}
        current = {k: v for k, v in current.items() if k != "uniqueItems"}
        if current not in (unpatched, value):
            raise ValueError(f"unexpected pinned document reference: {name}")
        request.update(required=[key, "catalog"], additionalProperties=False)
        request["properties"].update({key: value, "catalog": copy.deepcopy(union)})

    # M4 request-only schemas: do not constrain the report's normalized selection.
    nonblank = {"type": "string", "minLength": 1, "pattern": r"\S"}
    selection = copy.deepcopy(shape("validation.OCSFSelection", ("version", "class", "class_uid", "category", "category_uid", "profiles", "extensions")))
    for key in ("version", "class", "category"):
        if selection["properties"][key] != {"type": "string"}:
            raise ValueError(f"unexpected pinned selection string: {key}")
        selection["properties"][key] = copy.deepcopy(nonblank)
    for key in ("class_uid", "category_uid"):
        if selection["properties"][key] != {"type": "integer"}:
            raise ValueError(f"unexpected pinned selection UID: {key}")
        selection["properties"][key] = {"type": "integer", "minimum": -(2**63), "maximum": 2**63 - 1}
    for key in ("profiles", "extensions"):
        current = {k: v for k, v in selection["properties"][key].items() if k != "uniqueItems"}
        if current != {"type": "array", "items": {"type": "string"}}:
            raise ValueError(f"unexpected pinned selection names: {key}")
        selection["properties"][key] = {"type": "array", "items": copy.deepcopy(nonblank), "uniqueItems": True}
    selection.update(required=["version"], additionalProperties=False,
                     oneOf=[{"required": [key]} for key in ("class", "class_uid", "category", "category_uid")],
                     description="Exact version and exactly one selector. Omitted profiles/extensions normalize to empty arrays; null is invalid. Names are nonblank UTF-8, unique and case-sensitive.")
    schemas["api.SchemaValidationSelection"] = selection
    inline = {"oneOf": [{"type": "object"}, {"type": "boolean"}]}
    target_union = {"oneOf": [
        {"type": "object", "required": ["kind", "schema"], "additionalProperties": False,
         "properties": {"kind": {"const": "json_schema"}, "identity": copy.deepcopy(nonblank),
                        "schema": copy.deepcopy(inline), "base_uri": copy.deepcopy(nonblank),
                        "resources": {"type": "object", "propertyNames": copy.deepcopy(nonblank), "additionalProperties": copy.deepcopy(inline),
                                      "description": "URI-keyed local inline JSON Schema resources. No URI is fetched."}}},
        {"type": "object", "required": ["kind", "catalog", "selection"], "additionalProperties": False,
         "properties": {"kind": {"const": "ocsf"}, "identity": copy.deepcopy(nonblank),
                        "catalog": {"type": "object", "description": "Official compiled OCSF catalog with compile_version 1, exact version and full extension set. Catalog vocabulary and annotation members are allowed."},
                        "selection": {"$ref": PREFIX + "api.SchemaValidationSelection"}}}],
        "description": "Strict tagged target: reject unknown or mixed members, nulls, duplicate JSON keys throughout and malformed Unicode. Schema defaults to Draft 2020-12; unsupported explicit dialects are input errors. Schema/catalog payloads retain their own vocabulary; unsupported field semantics produce incomplete results. Offline only."}
    target = schemas["api.SchemaValidationTarget"]
    unpatched_target = {"type": "object", "properties": {
        **{key: {"type": "string"} for key in ("kind", "identity", "base_uri")},
        **{key: {"type": "object"} for key in ("schema", "resources", "catalog")},
        "selection": {"$ref": PREFIX + "validation.OCSFSelection"}}}
    if target not in (unpatched_target, target_union):
        raise ValueError("unexpected pinned schema target shape")
    schemas["api.SchemaValidationTarget"] = target_union
    for name, key in (("api.SchemaValidationRequest", "document"), ("api.SchemaValidationBatchRequest", "documents")):
        request = shape(name, (key, "target"))
        if request["properties"]["target"] != {"$ref": PREFIX + "api.SchemaValidationTarget"}:
            raise ValueError(f"unexpected pinned schema target reference: {name}")
        doc_ref = {"$ref": PREFIX + "validation.QueryDocument"}
        old_doc_ref = {"$ref": PREFIX + "analysis.QueryDocument"}
        value = doc_ref if key == "document" else {"type": "array", "items": doc_ref, "minItems": 1}
        current = {k: v for k, v in request["properties"][key].items() if k != "uniqueItems"}
        unpatched = old_doc_ref if key == "document" else {"type": "array", "items": old_doc_ref}
        if current not in (unpatched, value):
            raise ValueError(f"unexpected pinned schema document reference: {name}")
        request.update(required=[key, "target"], additionalProperties=False)
        request["properties"][key] = value

    components = json.dumps(spec["components"], ensure_ascii=True, separators=(",", ":"))
    match = matches[0]
    template = template[:match.start(1)] + components + template[match.end(1):]
    # Validate all inputs and prepare outputs before modifying generated files.
    json_output = "{\n" + ",\n".join(
        "    " + json.dumps(key) + ": " + json.dumps(value, separators=(",", ":"))
        for key, value in spec.items()) + "\n}\n"
    outputs = ((json_path, json_output),
               (yaml_path, yaml.safe_dump(spec, sort_keys=True, allow_unicode=True)),
               (go_path, template))
    for path, content in outputs:
        path.write_text(content, encoding="utf-8")


if __name__ == "__main__":
    try:
        update(Path(sys.argv[1]) if len(sys.argv) == 2 else Path("docs"))
    except (KeyError, TypeError, ValueError, OSError, yaml.YAMLError) as error:
        sys.exit(f"OpenAPI validation reconciliation failed: {error}")
