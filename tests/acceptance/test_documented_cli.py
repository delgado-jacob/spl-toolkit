from __future__ import annotations

import json
import os
from pathlib import Path
import re
import subprocess

import pytest


def absolute_env_path(name: str) -> Path:
    value = os.environ.get(name)
    if not value:
        pytest.fail(f"{name} is required")
    path = Path(value)
    if not path.is_absolute():
        pytest.fail(f"{name} must be an absolute path")
    return path


def render(value, substitutions: dict[str, str]):
    if isinstance(value, str):
        rendered = value
        for token, replacement in substitutions.items():
            rendered = rendered.replace(token, replacement)
        return rendered
    if isinstance(value, list):
        return [render(item, substitutions) for item in value]
    return value


def test_documented_cli_examples(cli_path: Path, tmp_path: Path) -> None:
    manifest_path = Path(__file__).with_name("cli_examples.json")
    cases = json.loads(manifest_path.read_text(encoding="utf-8"))["cases"]
    for case in cases:
        case_dir = tmp_path / case["id"]
        case_dir.mkdir()
        substitutions = {"{cli}": str(cli_path), "{tmp}": str(case_dir)}
        for fixture in case.get("files", []):
            destination = Path(render(fixture["path"], substitutions))
            destination.parent.mkdir(parents=True, exist_ok=True)
            destination.write_text(fixture["content"], encoding="utf-8")
        completed = subprocess.run(
            render(case["argv"], substitutions),
            cwd=case_dir,
            check=False,
            capture_output=True,
            text=True,
        )
        assert completed.returncode == case["exit"], case["id"]
        assert completed.stdout == case["stdout"], case["id"]
        if "stderr_contains" in case:
            assert case["stderr_contains"] in completed.stderr, case["id"]
        else:
            assert completed.stderr == "", case["id"]
        for expected in case.get("output_files", []):
            destination = Path(render(expected["path"], substitutions))
            assert destination.read_text(encoding="utf-8") == expected["content"], case["id"]


def test_documentation_example_coverage() -> None:
    root = absolute_env_path("SPL_DOCS_ROOT")
    manifest = json.loads((Path(__file__).with_name("cli_examples.json")).read_text(encoding="utf-8"))
    ids = {case["id"] for case in manifest["cases"]}
    cli_doc = (root / "docs/cli.md").read_text(encoding="utf-8")
    referenced = set(re.findall(r"<!--\s*cli-example:\s*([a-z0-9-]+)\s*-->", cli_doc))
    assert referenced == ids

    current_docs = [root / "README.md"] + [
        path for path in (root / "docs").rglob("*.md") if "superpowers" not in path.parts
    ]
    fence = re.compile(r"```(?:bash|console|sh)\n(.*?)```", re.DOTALL)
    for path in current_docs:
        text = path.read_text(encoding="utf-8")
        for block in fence.findall(text):
            if re.search(r"(?:^|\s)(?:\./)?spl-toolkit\b", block):
                if path == root / "docs/cli.md":
                    continue
                canonical = "docs/cli.md" if path == root / "README.md" else "cli.md"
                assert f"]({canonical})" in text, f"{path} must link CLI commands to canonical usage"


@pytest.fixture
def cli_path() -> Path:
    return absolute_env_path("SPL_CLI")
