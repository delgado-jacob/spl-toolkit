#!/usr/bin/env python3
"""Reconcile pinned Swag output with the strict canonical validation decoders.

Swag cannot express strict M3/M4 inputs or inline JSON unions through Go
DTOs. Constrain request-only document copies; keep generated report schemas unchanged.
Requires the repository's pinned development PyYAML dependency.
"""
import copy
import json
from pathlib import Path
import re
import sys

import yaml


PREFIX = "#/components/schemas/"


def add_tooling(spec):
    """Embed reviewed owned contracts without changing legacy generated types."""
    source = Path(__file__).resolve().parents[1] / "contracts/v1/shared.schema.json"
    shared = json.loads(source.read_text(encoding="utf-8"))
    origin = shared["$id"] + "#/$defs/"
    schemas = spec["components"]["schemas"]
    needed = set()

    def convert(value, *, preserve_analysis=False):
        if isinstance(value, list):
            return [convert(item, preserve_analysis=preserve_analysis) for item in value]
        if not isinstance(value, dict):
            return value
        result = {
            key: convert(item, preserve_analysis=preserve_analysis)
            for key, item in value.items()
        }
        if "$ref" in result:
            ref = result["$ref"]
            if not ref.startswith(origin):
                raise ValueError("unexpected external tooling contract reference")
            name = ref.removeprefix(origin)
            if name.startswith("analysis.Requirement"):
                result["$ref"] = PREFIX + name
                return result
            if preserve_analysis and name.startswith("analysis."):
                if name not in schemas:
                    raise ValueError(f"missing generated analysis dependency: {name}")
                result["$ref"] = PREFIX + name
                return result
            needed.add(name)
            result["$ref"] = PREFIX + "tooling." + name
        return result

    def reference(name):
        needed.add(name)
        return {"$ref": PREFIX + "tooling." + name}

    requirement_names = (
        "analysis.RequirementQueryIdentity",
        "analysis.RequirementCoverage",
        "analysis.RequirementOccurrence",
        "analysis.RequirementItem",
        "analysis.RequirementGap",
        "analysis.RequirementSet",
    )
    for name in ("analysis.Position", "analysis.Location", "analysis.Diagnostic"):
        generated = schemas.get(name)
        canonical = shared["$defs"][name]
        if (not isinstance(generated, dict) or generated.get("type") != "object"
                or set(generated.get("properties", {})) != set(canonical["properties"])):
            raise ValueError(f"unexpected pinned analysis dependency shape: {name}")
        schemas[name] = convert(canonical, preserve_analysis=True)
    for name in requirement_names:
        generated = schemas.get(name)
        canonical = shared["$defs"][name]
        if (not isinstance(generated, dict) or generated.get("type") != "object"
                or set(generated.get("properties", {})) != set(canonical["properties"])):
            raise ValueError(f"unexpected pinned requirement schema shape: {name}")
        schemas[name] = convert(canonical, preserve_analysis=True)
    requirement_ref = {"$ref": PREFIX + "analysis.RequirementSet"}
    for name in ("analysis.Result",):
        generated = schemas.get(name)
        if (not isinstance(generated, dict)
                or generated.get("properties", {}).get("requirements") != requirement_ref
                or "requirements" in generated.get("required", [])):
            raise ValueError(f"unexpected optional requirement embedding: {name}")
    response = spec.get("paths", {}).get("/query/requirements", {}).get("post", {}).get("responses", {}).get("200", {})
    if response.get("content", {}).get("application/json", {}).get("schema") != requirement_ref:
        raise ValueError("unexpected requirements response schema")

    for route, request, report in (
        ("corpus/scan", "corpus.Request", "corpus.Report"),
        ("corpus/graph", "corpus.Request", "graph.Report"),
        ("corpus/sarif", "corpus.Request", None),
        ("corpus/impact-schema", "impact.SchemaRequest", "impact.Report"),
        ("corpus/impact-mapping", "impact.MappingRequest", "impact.Report"),
        ("query/document", "QueryDocumentRequest", "document.Snapshot"),
    ):
        response = reference(report) if report else {
            "$ref": "https://docs.oasis-open.org/sarif/sarif/v2.1.0/errata01/os/schemas/sarif-schema-2.1.0.json"}
        spec["paths"]["/" + route] = {"post": {
            "summary": "Canonical " + route.replace("/", " "),
            "description": "Strict inline JSON only; no filesystem paths or remote retrieval. Reject unknown or duplicate members, nulls, malformed Unicode and trailing JSON. Exact body limit 8,388,608 bytes. Content statuses valid, invalid and incomplete return 200. Full canonical evidence is retained. SARIF's official schema is vendored under contracts/sarif for offline validation.",
            "tags": ["tooling"],
            "requestBody": {"required": True, "content": {"application/json": {"schema": reference(request)}}},
            "responses": {
                "200": {"description": "Canonical report", "content": {"application/json": {"schema": response}}},
                **{code: {"description": message, "content": {"application/json": {"schema": {"$ref": PREFIX + "api.ErrorResponse"}}}}
                   for code, message in (("400", "Invalid input, content type or body limit"), ("500", "Internal failure"))},
            },
        }}
    visited = set()
    while needed - visited:
        name = sorted(needed - visited)[0]
        schemas["tooling." + name] = convert(shared["$defs"][name])
        visited.add(name)
    snapshot = schemas.get("tooling.document.Snapshot")
    if (not isinstance(snapshot, dict)
            or snapshot.get("properties", {}).get("requirements") != requirement_ref
            or "requirements" in snapshot.get("required", [])):
        raise ValueError("unexpected optional requirement embedding: tooling.document.Snapshot")


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
    for key, values, default in (("language", ["", "spl", "spl2"], "spl"), ("profile", ["", "splunkd"], "splunkd"), ("version", ["", "current"], "current")):
        document["properties"][key].update(enum=values, default=default, description="Omitted or empty selects the default; other values are input errors.")
    analysis_request = shape("api.AnalysisRequest", ("text", "language", "profile", "version", "source_id"))
    if analysis_request not in (schemas["analysis.QueryDocument"], document):
        raise ValueError("unexpected pinned analysis request shape")
    schemas["api.AnalysisRequest"] = copy.deepcopy(document)
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

    # M6 request-only rewrite schemas. Runtime remains the canonical strict
    # decoder; these definitions make that contract executable for clients.
    identity = shape("api.RewriteIdentity", ("name", "path"))
    expected_identity = {"type": "object", "properties": {
        "name": {"type": "string"},
        "path": {"type": "array", "items": {"type": "string"}, "uniqueItems": False}}}
    patched_identity = {"type": "object", "properties": {
        "name": copy.deepcopy(nonblank),
        "path": {"type": "array", "items": copy.deepcopy(nonblank), "minItems": 1}},
        "additionalProperties": False, "oneOf": [{"required": ["name"]}, {"required": ["path"]}]}
    if identity not in (expected_identity, patched_identity):
        raise ValueError("unexpected pinned rewrite identity shape")
    identity.update(additionalProperties=False, oneOf=[{"required": ["name"]}, {"required": ["path"]}])
    identity["properties"] = {
        "name": copy.deepcopy(nonblank),
        "path": {"type": "array", "items": copy.deepcopy(nonblank), "minItems": 1}}

    name_identity = {"type": "object", "required": ["name"], "additionalProperties": False,
                     "properties": {"name": copy.deepcopy(nonblank)}}
    field_identity = {"$ref": PREFIX + "api.RewriteIdentity"}
    scalar = {"oneOf": [{"type": "string"}, {"type": "number"}, {"type": "boolean"}, {"type": "null"}]}
    condition = shape("api.RewriteCondition", ("all", "any", "fact", "kind", "identity", "operator", "value"))
    condition_ref = {"$ref": PREFIX + "api.RewriteCondition"}
    for key in ("all", "any"):
        condition["properties"][key] = {"type": "array", "items": condition_ref, "minItems": 1}
    condition["properties"]["fact"] = {"type": "string", "enum": ["literal", "source_reference_present"]}
    condition["properties"]["kind"] = {"type": "string", "enum": ["field", "index", "source", "sourcetype", "lookup", "dataset", "data_model"]}
    condition["properties"]["identity"] = {"$ref": PREFIX + "api.RewriteIdentity"}
    condition["properties"]["operator"] = {"type": "string", "enum": ["equals", "contains"]}
    condition["properties"]["value"] = copy.deepcopy(scalar)
    combinator_other = ["any", "fact", "kind", "identity", "operator", "value"]
    all_branch = {"required": ["all"], "not": {"anyOf": [{"required": [key]} for key in combinator_other]}}
    any_branch = {"required": ["any"], "not": {"anyOf": [{"required": [key]} for key in ["all"] + combinator_other[1:]]}}
    leaf_other = {"not": {"anyOf": [{"required": ["all"]}, {"required": ["any"]}]}}

    def literal_branch(kinds, identity_schema, operator, value_schema):
        branch = copy.deepcopy(leaf_other)
        branch.update(required=["fact", "kind", "identity", "operator", "value"], properties={
            "fact": {"const": "literal"}, "kind": {"type": "string", "enum": kinds},
            "identity": copy.deepcopy(identity_schema), "operator": {"const": operator},
            "value": copy.deepcopy(value_schema)})
        return branch

    source_reference = copy.deepcopy(leaf_other)
    source_reference.update(required=["fact", "kind", "identity"], properties={
        "fact": {"const": "source_reference_present"}, "kind": {"const": "field"},
        "identity": copy.deepcopy(field_identity)})
    source_reference["not"] = {"anyOf": [{"required": [key]} for key in ("all", "any", "operator", "value")]}
    dependency_kinds = ["index", "source", "sourcetype"]
    condition.update(additionalProperties=False, oneOf=[
        all_branch, any_branch, source_reference,
        literal_branch(["field"], field_identity, "equals", scalar),
        literal_branch(["field"], field_identity, "contains", {"type": "string"}),
        literal_branch(dependency_kinds, name_identity, "equals", scalar),
        literal_branch(dependency_kinds, name_identity, "contains", {"type": "string"})])

    rule = shape("api.RewriteRule", ("id", "kind", "source", "target", "when"))
    rule.update(required=["id", "kind", "source", "target"], additionalProperties=False)
    rule["properties"]["id"] = copy.deepcopy(nonblank)
    rule_kinds = ["field", "index", "source", "sourcetype", "lookup", "dataset", "data_model"]
    rule["properties"]["kind"] = {"type": "string", "enum": rule_kinds}
    for key in ("source", "target"):
        rule["properties"][key] = {"$ref": PREFIX + "api.RewriteIdentity"}
    rule["properties"]["when"] = {"$ref": PREFIX + "api.RewriteCondition"}
    rule["oneOf"] = [
        {"properties": {"kind": {"const": "field"}, "source": copy.deepcopy(field_identity), "target": copy.deepcopy(field_identity)}},
        {"properties": {"kind": {"type": "string", "enum": rule_kinds[1:]},
                        "source": copy.deepcopy(name_identity), "target": copy.deepcopy(name_identity)}}]

    field_target = {"type": "object", "required": ["kind", "catalog"], "additionalProperties": False,
                    "properties": {"kind": {"const": "field_list"}, "catalog": copy.deepcopy(union)}}
    rewrite_target = {"oneOf": [field_target] + copy.deepcopy(target_union["oneOf"]),
                      "description": "Optional inline offline validation target. Paths and retrieval URLs are not accepted."}
    for name, key in (("api.RewriteRequest", "document"), ("api.RewriteBatchRequest", "documents")):
        request = shape(name, ("schema_version", "mode", key, "rules", "validation_target"))
        doc_ref = {"$ref": PREFIX + "validation.QueryDocument"}
        old_doc_ref = {"$ref": PREFIX + "analysis.QueryDocument"}
        value = doc_ref if key == "document" else {"type": "array", "items": doc_ref, "minItems": 1}
        current = {k: v for k, v in request["properties"][key].items() if k != "uniqueItems"}
        unpatched = old_doc_ref if key == "document" else {"type": "array", "items": old_doc_ref}
        if current not in (unpatched, value):
            raise ValueError(f"unexpected pinned rewrite document reference: {name}")
        request.update(required=["schema_version", key, "rules"], additionalProperties=False)
        request["properties"]["schema_version"] = {"type": "integer", "const": 1}
        request["properties"]["mode"] = {"type": "string", "enum": ["preview", "apply"], "default": "preview"}
        request["properties"][key] = value
        request["properties"]["rules"] = {"type": "array", "items": {"$ref": PREFIX + "api.RewriteRule"}}
        request["properties"]["validation_target"] = copy.deepcopy(rewrite_target)

    capability_form = shape("analysis.RewriteCapabilityForm", ("kind", "role", "identity_forms", "supported", "limitations"))
    capability_form.update(required=["kind", "role", "identity_forms", "supported", "limitations"])
    capability = shape("analysis.RewriteCapabilityManifest", ("schema_version", "forms"))
    capability.update(required=["schema_version", "forms"])
    capability["properties"]["schema_version"] = {"type": "integer", "const": 1}

    shared_path = Path(__file__).resolve().parents[1] / "contracts/v1/shared.schema.json"
    shared = json.loads(shared_path.read_text(encoding="utf-8"))
    shared_origin = shared["$id"] + "#/$defs/"

    def capability_contract(value):
        if isinstance(value, list):
            return [capability_contract(item) for item in value]
        if not isinstance(value, dict):
            return value
        result = {key: capability_contract(item) for key, item in value.items()}
        if "$ref" in result:
            ref = result["$ref"]
            if not ref.startswith(shared_origin):
                raise ValueError("unexpected external capability contract reference")
            result["$ref"] = PREFIX + ref.removeprefix(shared_origin)
        return result

    capability_names = (
        "analysis.CapabilityClaim",
        "analysis.CapabilityDimensions",
        "analysis.CapabilityProvenance",
        "analysis.CapabilityRecord",
        "analysis.CapabilityStateCounts",
        "analysis.CapabilitySummary",
        "analysis.CapabilityDiagnosticExpectation",
        "analysis.CapabilitySyntaxObservation",
        "analysis.CapabilityStageExpectation",
        "analysis.CapabilityReferenceExpectation",
        "analysis.CapabilityDependencyExpectation",
        "analysis.CapabilityTransitionExpectation",
        "analysis.CapabilitySemanticsObservation",
        "analysis.CapabilityRequirementExpectation",
        "analysis.CapabilityRequirementsObservation",
        "analysis.CapabilityLintingObservation",
        "analysis.CapabilityRewriteObservation",
        "analysis.CapabilityEvidenceObservations",
        "analysis.CapabilityEvidence",
    )
    for name in capability_names:
        canonical = shared["$defs"][name]
        shape(name, tuple(canonical["properties"]))
        schemas[name] = capability_contract(canonical)

    canonical_manifest = shared["$defs"]["analysis.CapabilityManifest"]
    shape("analysis.CapabilityManifest", tuple(canonical_manifest["properties"]))
    schemas["analysis.CapabilityManifest"] = capability_contract(canonical_manifest)
    manifest = schemas["analysis.CapabilityManifest"]
    manifest["properties"]["rewrite"] = {
        "$ref": PREFIX + "analysis.RewriteCapabilityManifest",
        "description": "Optional rewrite support for this selected dialect; inspect each form rather than assuming universal support."}

    path_matches = list(re.finditer(r'^    "paths": (\{.*\}),?$', template, re.MULTILINE))
    if len(path_matches) != 1 or json.loads(path_matches[0][1]) != spec["paths"]:
        raise ValueError("unexpected pinned paths template")
    add_tooling(spec)
    path_match = path_matches[0]
    template = template[:path_match.start(1)] + json.dumps(spec["paths"], ensure_ascii=True, separators=(",", ":")) + template[path_match.end(1):]
    # Recalculate the component span after replacing paths.
    matches = list(re.finditer(r'^    "components": (\{.*\}),$', template, re.MULTILINE))
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
