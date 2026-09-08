"""Keep the public schema reason reference complete as the registry evolves."""
from pathlib import Path
import re

ROOT = Path(__file__).resolve().parents[2]


def test_schema_capability_reason_documentation_parity():
    source = (ROOT / "pkg/validation/json_schema.go").read_text()
    registry = source.split("var schemaCapabilityReasons = map[string]string{", 1)[1].split("\n}", 1)[0]
    reasons = set(re.findall(r'"([a-z_]+)"\s*:', registry))
    document = (ROOT / "docs/API.md").read_text()
    document = document.split("### Schema diagnostic reasons", 1)[1]
    documented = set(re.findall(r'^\| `([a-z_]+)` \|', document, re.MULTILINE))
    assert reasons and reasons == documented, {"missing": sorted(reasons - documented), "unknown": sorted(documented - reasons)}
