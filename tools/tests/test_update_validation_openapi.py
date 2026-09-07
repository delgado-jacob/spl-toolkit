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
        schemas = {"analysis.QueryDocument": document, "validation.FieldCatalog": catalog, "unrelated": {"type": "string", "description": "retain me"}}
        for name, key in (("Request", "document"), ("BatchRequest", "documents")):
            value = {"$ref": "#/components/schemas/analysis.QueryDocument"}
            if key == "documents":
                value = {"type": "array", "items": value}
            schemas["validation." + name] = {"type": "object", "properties": {key: value, "catalog": {"$ref": "#/components/schemas/validation.FieldCatalog"}}}
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
            self.assertEqual(schemas["unrelated"], original["components"]["schemas"]["unrelated"])
            self.assertEqual(spec["paths"], original["paths"])
            for name, keys in (("validation.Request", ["document", "catalog"]), ("validation.BatchRequest", ["documents", "catalog"]), ("validation.QueryDocument", ["text"]), ("validation.FieldCatalog", ["fields"])):
                self.assertEqual(schemas[name]["required"], keys)
                self.assertIs(schemas[name]["additionalProperties"], False)
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

    def test_unexpected_shape_fails_without_writes(self):
        for failure in ("missing schema", "template mismatch", "yaml mismatch", "catalog type drift"):
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
                            data["components"]["schemas"]["validation.FieldCatalog"]["properties"]["identity"]["type"] = "integer"
                            path.write_text(yaml.safe_dump(data))
                        else:
                            path.write_text(content.replace('"identity": {"type": "string"}', '"identity": {"type": "integer"}'))
                before = {p.name: p.read_bytes() for p in root.iterdir()}
                result = self.run_script(root)
                self.assertNotEqual(result.returncode, 0)
                self.assertEqual(before, {p.name: p.read_bytes() for p in root.iterdir()})
