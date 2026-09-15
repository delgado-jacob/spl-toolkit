import copy
import json
from pathlib import Path
import re
import subprocess
import sys
import tempfile
import unittest

import jsonschema
from referencing import Registry, Resource
import yaml


SCRIPT = Path(__file__).resolve().parents[1] / "update_validation_openapi.py"


class ValidationOpenAPITests(unittest.TestCase):
    def schema_accepts(self, schemas, name, instance):
        def valid(schema, value):
            if "$ref" in schema:
                return valid(schemas[schema["$ref"].removeprefix("#/components/schemas/")], value)
            if "const" in schema and value != schema["const"]:
                return False
            if "enum" in schema and value not in schema["enum"]:
                return False
            if "type" in schema:
                expected = schema["type"]
                matches = {
                    "object": isinstance(value, dict),
                    "array": isinstance(value, list),
                    "string": isinstance(value, str),
                    "integer": isinstance(value, int) and not isinstance(value, bool),
                    "number": isinstance(value, (int, float)) and not isinstance(value, bool),
                    "boolean": isinstance(value, bool),
                    "null": value is None,
                }[expected]
                if not matches:
                    return False
            if "oneOf" in schema and sum(valid(branch, value) for branch in schema["oneOf"]) != 1:
                return False
            if "anyOf" in schema and not any(valid(branch, value) for branch in schema["anyOf"]):
                return False
            if "not" in schema and valid(schema["not"], value):
                return False
            if isinstance(value, dict):
                if any(key not in value for key in schema.get("required", [])):
                    return False
                properties = schema.get("properties", {})
                if schema.get("additionalProperties") is False and any(key not in properties for key in value):
                    return False
                if any(not valid(properties[key], item) for key, item in value.items() if key in properties):
                    return False
            if isinstance(value, list):
                if len(value) < schema.get("minItems", 0):
                    return False
                if "items" in schema and any(not valid(schema["items"], item) for item in value):
                    return False
            if isinstance(value, str):
                if len(value) < schema.get("minLength", 0):
                    return False
                if "pattern" in schema and re.search(schema["pattern"], value) is None:
                    return False
            return True

        return valid({"$ref": "#/components/schemas/" + name}, instance)

    def fixture(self, root):
        document = {"type": "object", "properties": {key: {"type": "string"} for key in ("text", "language", "profile", "version", "source_id")}}
        catalog = {"type": "object", "properties": {"fields": {"type": "array", "items": {"type": "string"}}, "optional_fields": {"type": "array", "items": {"type": "string"}}, "identity": {"type": "string"}, "version": {"type": "string"}}}
        schemas = {"api.AnalysisRequest": copy.deepcopy(document), "analysis.QueryDocument": document, "validation.FieldCatalog": catalog, "unrelated": {"type": "string", "description": "retain me"}}
        schemas["analysis.Position"] = {"type": "object", "properties": {key: {"type": "integer"} for key in ("offset", "line", "column")}}
        schemas["analysis.Location"] = {"type": "object", "properties": {key: {"$ref": "#/components/schemas/analysis.Position"} for key in ("start", "end")}}
        schemas["analysis.Diagnostic"] = {"type": "object", "properties": {
            **{key: {"type": "string"} for key in ("code", "severity", "category", "message", "stage_id", "scope_id")},
            "location": {"$ref": "#/components/schemas/analysis.Location"}}}
        schemas["analysis.RequirementQueryIdentity"] = {"type": "object", "properties": {key: {"type": "string"} for key in ("source_id", "language", "profile", "version", "query_digest")}}
        schemas["analysis.RequirementCoverage"] = {"type": "object", "properties": {
            "complete": {"type": "boolean"}, "reasons": {"type": "array", "items": {"type": "string"}, "uniqueItems": False}}}
        schemas["analysis.RequirementOccurrence"] = {"type": "object", "properties": {
            **{key: {"type": "string"} for key in ("reference_id", "original_name", "binding", "stage_id", "scope_id")},
            "location": {"$ref": "#/components/schemas/analysis.Location"}}}
        schemas["analysis.RequirementItem"] = {"type": "object", "properties": {
            **{key: {"type": "string"} for key in ("id", "kind", "identity", "role", "necessity", "origin", "resolution")},
            "occurrences": {"type": "array", "items": {"$ref": "#/components/schemas/analysis.RequirementOccurrence"}, "uniqueItems": False}}}
        schemas["analysis.RequirementGap"] = {"type": "object", "properties": {
            "code": {"type": "string"}, "message": {"type": "string"},
            "reference_ids": {"type": "array", "items": {"type": "string"}, "uniqueItems": False},
            "diagnostic_codes": {"type": "array", "items": {"type": "string"}, "uniqueItems": False}}}
        schemas["analysis.RequirementSet"] = {"type": "object", "properties": {
            "schema_version": {"type": "integer"},
            "query": {"$ref": "#/components/schemas/analysis.RequirementQueryIdentity"},
            "capability_revision": {"type": "string"}, "query_status": {"type": "string"},
            "coverage": {"$ref": "#/components/schemas/analysis.RequirementCoverage"},
            "items": {"type": "array", "items": {"$ref": "#/components/schemas/analysis.RequirementItem"}, "uniqueItems": False},
            "gaps": {"type": "array", "items": {"$ref": "#/components/schemas/analysis.RequirementGap"}, "uniqueItems": False},
            "diagnostics": {"type": "array", "items": {"$ref": "#/components/schemas/analysis.Diagnostic"}, "uniqueItems": False}}}
        for name in ("analysis.Result", "document.Snapshot"):
            schemas[name] = {"type": "object", "properties": {
                "schema_version": {"type": "integer"},
                "requirements": {"$ref": "#/components/schemas/analysis.RequirementSet"}}}
        for name, key in (("Request", "document"), ("BatchRequest", "documents")):
            value = {"$ref": "#/components/schemas/analysis.QueryDocument"}
            if key == "documents":
                value = {"type": "array", "items": value}
            schemas["validation." + name] = {"type": "object", "properties": {key: value, "catalog": {"$ref": "#/components/schemas/validation.FieldCatalog"}}}
        schemas["api.SchemaValidationTarget"] = {"type": "object", "properties": {
            **{key: {"type": "string"} for key in ("kind", "identity", "base_uri")},
            **{key: {"type": "object"} for key in ("schema", "resources", "catalog")},
            "selection": {"$ref": "#/components/schemas/validation.OCSFSelection"}}}
        schemas["validation.OCSFSelection"] = {"type": "object", "properties": {
            **{key: {"type": "string"} for key in ("version", "class", "category")},
            **{key: {"type": "integer"} for key in ("class_uid", "category_uid")},
            **{key: {"type": "array", "items": {"type": "string"}, "uniqueItems": False} for key in ("profiles", "extensions")}}}
        for name, key in (("SchemaValidationRequest", "document"), ("SchemaValidationBatchRequest", "documents")):
            value = {"$ref": "#/components/schemas/analysis.QueryDocument"}
            if key == "documents":
                value = {"type": "array", "items": value, "uniqueItems": False}
            schemas["api." + name] = {"type": "object", "properties": {key: value, "target": {"$ref": "#/components/schemas/api.SchemaValidationTarget"}}}
        schemas["api.RewriteIdentity"] = {"type": "object", "properties": {"name": {"type": "string"}, "path": {"type": "array", "items": {"type": "string"}, "uniqueItems": False}}}
        schemas["api.RewriteCondition"] = {"type": "object", "properties": {"all": {"type": "array", "items": {"$ref": "#/components/schemas/api.RewriteCondition"}, "uniqueItems": False}, "any": {"type": "array", "items": {"$ref": "#/components/schemas/api.RewriteCondition"}, "uniqueItems": False}, "fact": {"type": "string"}, "kind": {"type": "string"}, "identity": {"$ref": "#/components/schemas/api.RewriteIdentity"}, "operator": {"type": "string"}, "value": {}}}
        schemas["api.RewriteRule"] = {"type": "object", "properties": {"id": {"type": "string"}, "kind": {"type": "string"}, "source": {"$ref": "#/components/schemas/api.RewriteIdentity"}, "target": {"$ref": "#/components/schemas/api.RewriteIdentity"}, "when": {"$ref": "#/components/schemas/api.RewriteCondition"}}}
        for name, key in (("RewriteRequest", "document"), ("RewriteBatchRequest", "documents")):
            value = {"$ref": "#/components/schemas/analysis.QueryDocument"}
            if key == "documents": value = {"type": "array", "items": value, "uniqueItems": False}
            schemas["api." + name] = {"type": "object", "properties": {"schema_version": {"type": "integer"}, "mode": {"type": "string"}, key: value, "rules": {"type": "array", "items": {"$ref": "#/components/schemas/api.RewriteRule"}, "uniqueItems": False}, "validation_target": {"type": "object"}}}
        schemas["analysis.RewriteCapabilityForm"] = {"type": "object", "properties": {"kind": {"type": "string"}, "role": {"type": "string"}, "identity_forms": {"type": "array", "items": {"type": "string"}, "uniqueItems": False}, "supported": {"type": "boolean"}, "limitations": {"type": "array", "items": {"type": "string"}, "uniqueItems": False}}}
        schemas["analysis.RewriteCapabilityManifest"] = {"type": "object", "properties": {"schema_version": {"type": "integer"}, "forms": {"type": "array", "items": {"$ref": "#/components/schemas/analysis.RewriteCapabilityForm"}, "uniqueItems": False}}}
        schemas["analysis.CapabilityManifest"] = {"type": "object", "properties": {"rewrite": {"$ref": "#/components/schemas/analysis.RewriteCapabilityManifest"}, "schema_version": {"type": "integer"}, "language": {"type": "string"}, "profile": {"type": "string"}, "version": {"type": "string"}, "documentation_snapshot": {"type": "string"}, "commands": {"type": "array", "items": {"type": "object"}}, "functions": {"type": "array", "items": {"type": "object"}}}}
        spec = {"openapi": "3.1.0", "components": {"schemas": schemas}, "paths": {
            "/unrelated": {},
            "/query/requirements": {"post": {"responses": {"200": {"content": {"application/json": {
                "schema": {"$ref": "#/components/schemas/analysis.RequirementSet"}}}}}}},
        }}
        (root / "swagger.json").write_text(json.dumps(spec), encoding="utf-8")
        (root / "swagger.yaml").write_text(yaml.safe_dump(spec), encoding="utf-8")
        (root / "docs.go").write_text('package docs\nconst docTemplate = `{\n    "components": ' + json.dumps(spec["components"]) + ',\n    "paths": ' + json.dumps(spec["paths"]) + ',\n    "info": {"title": "{{.Title}}"}\n}`\n', encoding="utf-8")
        return copy.deepcopy(spec)

    def run_script(self, root):
        return subprocess.run([sys.executable, str(SCRIPT), str(root)], text=True, capture_output=True)

    def test_shapes_agree_and_repeated_generation_is_idempotent(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            original = self.fixture(root)
            result = self.run_script(root)
            self.assertEqual(result.returncode, 0, result.stderr)
            spec = json.loads((root / "swagger.json").read_text())
            schemas = spec["components"]["schemas"]
            self.assertEqual(schemas["analysis.QueryDocument"], original["components"]["schemas"]["analysis.QueryDocument"])
            self.assertEqual(schemas["validation.OCSFSelection"], original["components"]["schemas"]["validation.OCSFSelection"])
            self.assertEqual(schemas["unrelated"], original["components"]["schemas"]["unrelated"])
            self.assertEqual(spec["paths"]["/unrelated"], original["paths"]["/unrelated"])
            self.assertEqual(set(spec["paths"]) - set(original["paths"]), {
                "/corpus/scan", "/corpus/graph", "/corpus/sarif",
                "/corpus/impact-schema", "/corpus/impact-mapping", "/query/document"})
            self.assertEqual(spec["paths"]["/query/document"]["post"]["requestBody"]["content"]["application/json"]["schema"],
                             {"$ref": "#/components/schemas/tooling.QueryDocumentRequest"})
            self.assertIs(schemas["tooling.corpus.Request"]["additionalProperties"], False)
            for name, keys in (("validation.Request", ["document", "catalog"]), ("validation.BatchRequest", ["documents", "catalog"]), ("validation.QueryDocument", ["text"]), ("validation.FieldCatalog", ["fields"])):
                self.assertEqual(schemas[name]["required"], keys)
                self.assertIs(schemas[name]["additionalProperties"], False)
            for name in ("api.AnalysisRequest", "validation.QueryDocument"):
                request = schemas[name]
                self.assertEqual(request["required"], ["text"])
                self.assertIs(request["additionalProperties"], False)
                for key, values, default in (("language", ["", "spl", "spl2"], "spl"), ("profile", ["", "splunkd"], "splunkd"), ("version", ["", "current"], "current")):
                    self.assertEqual(request["properties"][key]["enum"], values)
                    self.assertEqual(request["properties"][key]["default"], default)
                self.assertEqual(request["properties"]["text"], {"type": "string"})
                self.assertEqual(request["properties"]["source_id"], {"type": "string"})
            union = schemas["validation.Request"]["properties"]["catalog"]["oneOf"]
            self.assertEqual(union[0], {"type": "array", "items": {"type": "string", "minLength": 1}, "uniqueItems": True})
            self.assertEqual(union[1], {"$ref": "#/components/schemas/validation.FieldCatalog"})
            self.assertEqual(schemas["validation.BatchRequest"]["properties"]["documents"]["minItems"], 1)
            self.assertEqual(yaml.safe_load((root / "swagger.yaml").read_text()), spec)
            line = next(line for line in (root / "docs.go").read_text().splitlines() if line.startswith('    "components": '))
            self.assertEqual(json.loads(line.removeprefix('    "components": ').removesuffix(',')), spec["components"])
            before = {p.name: p.read_bytes() for p in root.iterdir()}
            self.assertEqual(self.run_script(root).returncode, 0)
            self.assertEqual(before, {p.name: p.read_bytes() for p in root.iterdir()})

    def test_schema_targets_and_selection_are_strict_inline_unions(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            self.fixture(root)
            result = self.run_script(root)
            self.assertEqual(result.returncode, 0, result.stderr)
            schemas = json.loads((root / "swagger.json").read_text())["components"]["schemas"]
            union = schemas["api.SchemaValidationTarget"]["oneOf"]
            self.assertEqual(len(union), 2)
            by_kind = {branch["properties"]["kind"]["const"]: branch for branch in union}
            js, ocsf = by_kind["json_schema"], by_kind["ocsf"]
            self.assertEqual(js["required"], ["kind", "schema"])
            self.assertEqual(ocsf["required"], ["kind", "catalog", "selection"])
            for branch in union:
                self.assertIs(branch["additionalProperties"], False)
            self.assertEqual(set(js["properties"]), {"kind", "identity", "schema", "base_uri", "resources"})
            self.assertEqual(set(ocsf["properties"]), {"kind", "identity", "catalog", "selection"})
            self.assertEqual(js["properties"]["schema"]["oneOf"], [{"type": "object"}, {"type": "boolean"}])
            self.assertEqual(js["properties"]["resources"]["additionalProperties"], js["properties"]["schema"])
            self.assertEqual(ocsf["properties"]["catalog"]["type"], "object")
            self.assertNotIn("additionalProperties", ocsf["properties"]["catalog"])
            selection = schemas["api.SchemaValidationSelection"]
            self.assertEqual(selection["required"], ["version"])
            self.assertEqual(selection["oneOf"], [{"required": [key]} for key in ("class", "class_uid", "category", "category_uid")])
            self.assertIs(selection["additionalProperties"], False)
            for key in ("class_uid", "category_uid"):
                self.assertEqual(selection["properties"][key], {"type": "integer", "minimum": -9223372036854775808, "maximum": 9223372036854775807})
            for key in ("profiles", "extensions"):
                self.assertEqual(selection["properties"][key]["type"], "array")
                self.assertIs(selection["properties"][key]["uniqueItems"], True)
            for name, key in (("SchemaValidationRequest", "document"), ("SchemaValidationBatchRequest", "documents")):
                wrapper = schemas["api." + name]
                self.assertEqual(wrapper["required"], [key, "target"])
                self.assertIs(wrapper["additionalProperties"], False)
            before = {p.name: p.read_bytes() for p in root.iterdir()}
            self.assertEqual(self.run_script(root).returncode, 0)
            self.assertEqual(before, {p.name: p.read_bytes() for p in root.iterdir()})

    def test_rewrite_requests_and_optional_capability_are_strict(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            self.fixture(root)
            result = self.run_script(root)
            self.assertEqual(result.returncode, 0, result.stderr)
            schemas = json.loads((root / "swagger.json").read_text())["components"]["schemas"]
            for name, key in (("api.RewriteRequest", "document"), ("api.RewriteBatchRequest", "documents")):
                request = schemas[name]
                self.assertEqual(request["required"], ["schema_version", key, "rules"])
                self.assertIs(request["additionalProperties"], False)
                self.assertEqual(request["properties"]["schema_version"], {"type": "integer", "const": 1})
                self.assertEqual(request["properties"]["mode"], {"type": "string", "enum": ["preview", "apply"], "default": "preview"})
                if key == "documents": self.assertEqual(request["properties"][key]["minItems"], 1)
            identity = schemas["api.RewriteIdentity"]
            self.assertEqual(identity["oneOf"], [{"required": ["name"]}, {"required": ["path"]}])
            self.assertIs(identity["additionalProperties"], False)
            document = {"text": "search src=x"}
            accepted_rules = [
                {"id": "field-path", "kind": "field", "source": {"path": ["actor", "name"]}, "target": {"path": ["user", "name"]}},
                {"id": "source-name", "kind": "source", "source": {"name": "old"}, "target": {"name": "new"},
                 "when": {"fact": "literal", "kind": "source", "identity": {"name": "old"}, "operator": "contains", "value": "prod"}},
                {"id": "presence", "kind": "field", "source": {"name": "src"}, "target": {"name": "user"},
                 "when": {"fact": "source_reference_present", "kind": "field", "identity": {"path": ["actor", "name"]}}},
            ]
            for rule in accepted_rules:
                request = {"schema_version": 1, "document": document, "rules": [rule]}
                self.assertTrue(self.schema_accepts(schemas, "api.RewriteRequest", request), rule)
            rejected_rules = [
                {"id": "presence-index", "kind": "index", "source": {"name": "old"}, "target": {"name": "new"},
                 "when": {"fact": "source_reference_present", "kind": "index", "identity": {"name": "main"}}},
                {"id": "literal-lookup", "kind": "lookup", "source": {"name": "old"}, "target": {"name": "new"},
                 "when": {"fact": "literal", "kind": "lookup", "identity": {"name": "users"}, "operator": "equals", "value": "users"}},
                {"id": "contains-number", "kind": "field", "source": {"name": "src"}, "target": {"name": "user"},
                 "when": {"fact": "literal", "kind": "field", "identity": {"name": "src"}, "operator": "contains", "value": 7}},
                {"id": "rule-path-index", "kind": "index", "source": {"path": ["old"]}, "target": {"name": "new"}},
                {"id": "rule-target-path-index", "kind": "index", "source": {"name": "old"}, "target": {"path": ["new"]}},
                {"id": "condition-path-source", "kind": "source", "source": {"name": "old"}, "target": {"name": "new"},
                 "when": {"fact": "literal", "kind": "source", "identity": {"path": ["old"]}, "operator": "equals", "value": "old"}},
            ]
            for rule in rejected_rules:
                request = {"schema_version": 1, "document": document, "rules": [rule]}
                self.assertFalse(self.schema_accepts(schemas, "api.RewriteRequest", request), rule)
            batch = {"schema_version": 1, "documents": [document], "rules": accepted_rules}
            self.assertTrue(self.schema_accepts(schemas, "api.RewriteBatchRequest", batch))
            self.assertEqual(schemas["analysis.CapabilityManifest"]["properties"]["rewrite"], {"$ref": "#/components/schemas/analysis.RewriteCapabilityManifest", "description": "Optional rewrite support for this selected dialect; inspect each form rather than assuming universal support."})
            self.assertEqual(schemas["analysis.RewriteCapabilityManifest"]["required"], ["schema_version", "forms"])
            before = {p.name: p.read_bytes() for p in root.iterdir()}
            self.assertEqual(self.run_script(root).returncode, 0)
            self.assertEqual(before, {p.name: p.read_bytes() for p in root.iterdir()})

    def test_requirement_reports_are_exact_and_existing_v1_embedding_is_optional(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            self.fixture(root)
            result = self.run_script(root)
            self.assertEqual(result.returncode, 0, result.stderr)
            spec = json.loads((root / "swagger.json").read_text())
            schemas = spec["components"]["schemas"]
            required = {
                "analysis.RequirementQueryIdentity": ["source_id", "language", "profile", "version", "query_digest"],
                "analysis.RequirementCoverage": ["complete", "reasons"],
                "analysis.RequirementOccurrence": ["reference_id", "original_name", "binding", "stage_id", "scope_id", "location"],
                "analysis.RequirementItem": ["id", "kind", "identity", "role", "necessity", "origin", "resolution", "occurrences"],
                "analysis.RequirementGap": ["code", "message", "reference_ids", "diagnostic_codes"],
                "analysis.RequirementSet": ["schema_version", "query", "capability_revision", "query_status", "coverage", "items", "gaps", "diagnostics"],
            }
            for name, members in required.items():
                self.assertEqual(schemas[name]["required"], members)
                self.assertIs(schemas[name]["additionalProperties"], True)
            self.assertEqual(schemas["analysis.RequirementSet"]["properties"]["schema_version"], {"type": "integer", "const": 1})
            self.assertEqual(schemas["analysis.RequirementSet"]["properties"]["query_status"]["enum"], ["valid", "invalid", "incomplete"])
            self.assertEqual(schemas["analysis.RequirementItem"]["properties"]["kind"]["enum"], ["field", "index", "source", "sourcetype", "dataset", "data_model", "lookup", "macro"])
            self.assertEqual(schemas["analysis.RequirementItem"]["properties"]["necessity"]["enum"], ["required", "conditional"])
            self.assertEqual(schemas["analysis.RequirementItem"]["properties"]["origin"], {"const": "direct"})
            self.assertEqual(schemas["analysis.RequirementItem"]["properties"]["resolution"]["enum"], ["exact", "wildcard", "dynamic"])
            self.assertEqual(schemas["analysis.RequirementItem"]["properties"]["id"]["pattern"], "^req-[1-9][0-9]*$")
            self.assertEqual(schemas["analysis.RequirementOccurrence"]["properties"]["reference_id"]["pattern"], "^ref-(0|[1-9][0-9]*)$")
            self.assertEqual(schemas["analysis.RequirementOccurrence"]["properties"]["location"], {"$ref": "#/components/schemas/analysis.Location"})
            self.assertEqual(schemas["analysis.RequirementSet"]["properties"]["capability_revision"]["pattern"], "^sha256:[0-9a-f]{64}$")
            self.assertEqual(schemas["analysis.RequirementQueryIdentity"]["properties"]["query_digest"]["pattern"], "^sha256:[0-9a-f]{64}$")
            self.assertEqual(schemas["analysis.RequirementSet"]["properties"]["diagnostics"]["items"], {"$ref": "#/components/schemas/analysis.Diagnostic"})
            for name in ("analysis.Result", "tooling.document.Snapshot"):
                self.assertEqual(schemas[name]["properties"]["requirements"], {"$ref": "#/components/schemas/analysis.RequirementSet"})
                self.assertNotIn("requirements", schemas[name].get("required", []))
            response = spec["paths"]["/query/requirements"]["post"]["responses"]["200"]
            self.assertEqual(response["content"]["application/json"]["schema"], {"$ref": "#/components/schemas/analysis.RequirementSet"})
            before = {p.name: p.read_bytes() for p in root.iterdir()}
            self.assertEqual(self.run_script(root).returncode, 0)
            self.assertEqual(before, {p.name: p.read_bytes() for p in root.iterdir()})

    def test_requirement_diagnostics_match_canonical_draft_2020_12_contract(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            self.fixture(root)
            result = self.run_script(root)
            self.assertEqual(result.returncode, 0, result.stderr)

            spec = json.loads((root / "swagger.json").read_text())
            schemas = spec["components"]["schemas"]
            shared_path = SCRIPT.parents[1] / "contracts/v1/shared.schema.json"
            requirements_path = SCRIPT.parents[1] / "contracts/v1/requirements.schema.json"
            shared = json.loads(shared_path.read_text())
            requirements = json.loads(requirements_path.read_text())
            registry = Registry().with_resources((
                (shared["$id"], Resource.from_contents(shared)),
                (requirements["$id"], Resource.from_contents(requirements)),
            ))
            canonical = jsonschema.Draft202012Validator(requirements, registry=registry)
            openapi_schema = copy.deepcopy(spec)
            openapi_schema["$schema"] = "https://json-schema.org/draft/2020-12/schema"
            openapi_schema["$ref"] = "#/components/schemas/analysis.RequirementSet"
            openapi = jsonschema.Draft202012Validator(openapi_schema)

            authored = json.loads(
                (SCRIPT.parents[1] / "testdata/tooling/contracts.json").read_text()
            )
            minimal = copy.deepcopy(next(
                case["instance"] for case in authored["cases"]
                if case["id"] == "requirements-minimal"
            ))
            minimal["diagnostics"] = [{
                "code": "SPL_TEST",
                "severity": "warning",
                "category": "semantic",
                "message": "test diagnostic",
                "location": {
                    "start": {"offset": 0, "line": 1, "column": 1},
                    "end": {"offset": 4, "line": 1, "column": 5},
                },
                "stage_id": "stage-0",
                "scope_id": "scope-0",
            }]

            cases = [("valid", minimal, True)]
            for member in ("code", "severity", "category", "message", "location", "stage_id", "scope_id"):
                invalid = copy.deepcopy(minimal)
                del invalid["diagnostics"][0][member]
                cases.append((f"missing diagnostic {member}", invalid, False))
            for member in ("start", "end"):
                invalid = copy.deepcopy(minimal)
                del invalid["diagnostics"][0]["location"][member]
                cases.append((f"missing location {member}", invalid, False))
            for member in ("offset", "line", "column"):
                invalid = copy.deepcopy(minimal)
                del invalid["diagnostics"][0]["location"]["start"][member]
                cases.append((f"missing position {member}", invalid, False))
            invalid = copy.deepcopy(minimal)
            invalid["diagnostics"][0]["severity"] = "notice"
            cases.append(("invalid severity", invalid, False))
            for member, value in (("offset", -1), ("line", 0), ("column", 0)):
                invalid = copy.deepcopy(minimal)
                invalid["diagnostics"][0]["location"]["start"][member] = value
                cases.append((f"invalid position {member}", invalid, False))

            for label, instance, accepted in cases:
                with self.subTest(label=label):
                    self.assertEqual(canonical.is_valid(instance), accepted)
                    self.assertEqual(openapi.is_valid(instance), accepted)

            self.assertEqual(
                schemas["analysis.Diagnostic"]["required"],
                ["code", "severity", "category", "message", "location", "stage_id", "scope_id"],
            )
            self.assertEqual(
                schemas["analysis.Diagnostic"]["properties"]["severity"]["enum"],
                ["error", "warning", "info"],
            )
            self.assertEqual(schemas["analysis.Location"]["required"], ["start", "end"])
            self.assertEqual(
                schemas["analysis.Position"]["required"], ["offset", "line", "column"]
            )
            self.assertEqual(
                {key: schemas["analysis.Position"]["properties"][key]["minimum"]
                 for key in ("offset", "line", "column")},
                {"offset": 0, "line": 1, "column": 1},
            )

    def test_unexpected_shape_fails_without_writes(self):
        for failure in ("missing schema", "template mismatch", "yaml mismatch", "catalog type drift", "rewrite identity drift"):
            with self.subTest(failure=failure), tempfile.TemporaryDirectory() as directory:
                root = Path(directory)
                spec = self.fixture(root)
                if failure == "missing schema":
                    del spec["components"]["schemas"]["validation.Request"]
                    (root / "swagger.json").write_text(json.dumps(spec))
                elif failure == "template mismatch":
                    (root / "docs.go").write_text("unexpected generator output")
                elif failure == "yaml mismatch":
                    (root / "swagger.yaml").write_text("{}")
                else:
                    for path in root.iterdir():
                        content = path.read_text()
                        if path.suffix == ".yaml":
                            data = yaml.safe_load(content)
                            name = "validation.FieldCatalog" if failure == "catalog type drift" else "api.RewriteIdentity"
                            key = "identity" if failure == "catalog type drift" else "name"
                            data["components"]["schemas"][name]["properties"][key]["type"] = "integer"
                            path.write_text(yaml.safe_dump(data))
                        else:
                            old = '"identity": {"type": "string"}' if failure == "catalog type drift" else '"name": {"type": "string"}'
                            new = '"identity": {"type": "integer"}' if failure == "catalog type drift" else '"name": {"type": "integer"}'
                            path.write_text(content.replace(old, new))
                before = {p.name: p.read_bytes() for p in root.iterdir()}
                result = self.run_script(root)
                self.assertNotEqual(result.returncode, 0)
                self.assertEqual(before, {p.name: p.read_bytes() for p in root.iterdir()})
