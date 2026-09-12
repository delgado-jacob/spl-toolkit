import copy
import json
from pathlib import Path
import subprocess
import sys
import tempfile
import unittest

import yaml


SCRIPT = Path(__file__).resolve().parents[1] / "update_validation_openapi.py"


class ValidationOpenAPITests(unittest.TestCase):
    def fixture(self, root):
        document = {"type": "object", "properties": {key: {"type": "string"} for key in ("text", "language", "profile", "version", "source_id")}}
        catalog = {"type": "object", "properties": {"fields": {"type": "array", "items": {"type": "string"}}, "optional_fields": {"type": "array", "items": {"type": "string"}}, "identity": {"type": "string"}, "version": {"type": "string"}}}
        schemas = {"api.AnalysisRequest": copy.deepcopy(document), "analysis.QueryDocument": document, "validation.FieldCatalog": catalog, "unrelated": {"type": "string", "description": "retain me"}}
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
        spec = {"openapi": "3.1.0", "components": {"schemas": schemas}, "paths": {"/unrelated": {}}}
        (root / "swagger.json").write_text(json.dumps(spec), encoding="utf-8")
        (root / "swagger.yaml").write_text(yaml.safe_dump(spec), encoding="utf-8")
        (root / "docs.go").write_text('package docs\nconst docTemplate = `{\n    "components": ' + json.dumps(spec["components"]) + ',\n    "info": {"title": "{{.Title}}"}\n}`\n', encoding="utf-8")
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
            self.assertEqual(spec["paths"], original["paths"])
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
            rule = schemas["api.RewriteRule"]
            self.assertEqual(rule["required"], ["id", "kind", "source", "target"])
            self.assertEqual(rule["properties"]["kind"]["enum"], ["field", "index", "source", "sourcetype", "lookup", "dataset", "data_model"])
            condition = schemas["api.RewriteCondition"]
            self.assertEqual(len(condition["oneOf"]), 3)
            self.assertEqual(condition["properties"]["value"]["oneOf"], [{"type": "string"}, {"type": "number"}, {"type": "boolean"}, {"type": "null"}])
            self.assertEqual(schemas["analysis.CapabilityManifest"]["properties"]["rewrite"], {"$ref": "#/components/schemas/analysis.RewriteCapabilityManifest", "description": "Optional rewrite support for this selected dialect; inspect each form rather than assuming universal support."})
            self.assertEqual(schemas["analysis.RewriteCapabilityManifest"]["required"], ["schema_version", "forms"])
            before = {p.name: p.read_bytes() for p in root.iterdir()}
            self.assertEqual(self.run_script(root).returncode, 0)
            self.assertEqual(before, {p.name: p.read_bytes() for p in root.iterdir()})

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
