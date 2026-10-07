"""Run with an installed spl_toolkit package and its native library."""
import json
import os
from pathlib import Path

from spl_toolkit import SPLMapper

request = json.loads(Path(__file__).with_name("request.json").read_text(encoding="utf-8"))
with SPLMapper(library_path=os.environ.get("SPL_NATIVE_LIBRARY")) as mapper:
    report = mapper.workflow_assess(request)
    evidence = mapper.workflow_evidence({"schema_version": 1, "report": report, "include": []})
    print(json.dumps(evidence, indent=2))
