"""Guard package acceptance against a host-only or unprovisioned wheel closure."""
import json
import os
from pathlib import Path
import re
import subprocess
import sys

from packaging.requirements import Requirement
from packaging.tags import compatible_tags, cpython_tags, mac_platforms
from packaging.utils import canonicalize_name, parse_wheel_filename
import pytest
import yaml


ROOT = Path(__file__).resolve().parents[2]


def locked_requirements():
    text = (ROOT / "tools/requirements-package-check-hashed.lock").read_text()
    return {
        canonicalize_name(Requirement(line.split(" --hash=", 1)[0]).name):
        (Requirement(line.split(" --hash=", 1)[0]), set(re.findall(r"--hash=sha256:([0-9a-f]{64})", line)))
        for line in text.replace("\\\n", "").splitlines()
        if line and not line.startswith("#")
    }


@pytest.mark.parametrize("minor", [11, 12, 13, 14])
@pytest.mark.parametrize("target", ["linux-amd64", "darwin-amd64", "darwin-arm64", "windows-amd64"])
def test_locked_wheels_cover_release_matrix(minor, target):
    platforms = {
        "linux-amd64": ["manylinux_2_17_x86_64"],
        "windows-amd64": ["win_amd64"],
        "darwin-amd64": list(mac_platforms((15, 0), "x86_64")),
        "darwin-arm64": list(mac_platforms((15, 0), "arm64")),
    }[target]
    tags = set(cpython_tags((3, minor), platforms=platforms))
    tags.update(compatible_tags((3, minor), platforms=platforms))
    locked = locked_requirements()
    manifest = json.loads((ROOT / "tools/package-check-wheels.json").read_text())
    assert set(locked) == {package["name"] for package in manifest["packages"]}
    environment = {"python_version": f"3.{minor}", "platform_system": "Windows" if target.startswith("windows") else "Darwin" if target.startswith("darwin") else "Linux"}
    assert "colorama" in locked
    for package in manifest["packages"]:
        requirement, hashes = locked[package["name"]]
        assert hashes == {wheel["sha256"] for wheel in package["wheels"]}
        assert package["metadata_url"] == f'https://pypi.org/pypi/{package["name"]}/{package["version"]}/json'
        if requirement.marker and not requirement.marker.evaluate(environment):
            continue
        candidates = []
        for wheel in package["wheels"]:
            name, version, _, wheel_tags = parse_wheel_filename(wheel["filename"])
            assert name == canonicalize_name(package["name"])
            assert version in requirement.specifier
            if tags.intersection(wheel_tags):
                candidates.append(wheel)
        assert candidates, (target, minor, package["name"])


def test_package_lock_preserves_all_existing_pins():
    locked = locked_requirements()
    for filename in ("requirements-dev.txt", "requirements-build.txt"):
        for line in (ROOT / "python" / filename).read_text().splitlines():
            if not line or line.startswith(("#", "-r")):
                continue
            requirement = Requirement(line)
            actual, hashes = locked[canonicalize_name(requirement.name)]
            assert actual == requirement
            assert hashes


@pytest.mark.parametrize("job", ["release-baseline", "installed-wheel"])
def test_ci_provisions_wheelhouse_before_isolated_acceptance(tmp_path, job):
    workflow = yaml.safe_load((ROOT / ".github/workflows/ci.yml").read_text())
    steps = workflow["jobs"][job]["steps"]
    configuration = next(step for step in steps if step.get("name") == "Configure isolated package-check wheelhouse")
    provision = next(step for step in steps if step.get("name") == "Provision hash-locked package-check dependencies")
    acceptance = next(step for step in steps if "tools/check_package.py" in step.get("run", "") or "tools/check_reproducible.py" in step.get("run", ""))
    assert steps.index(configuration) < steps.index(provision) < steps.index(acceptance)
    temporary = tmp_path / "runner temp with spaces"
    temporary.mkdir()
    github_env = tmp_path / "github-env"
    # Execute the platform-neutral configuration itself, including URI escaping.
    command = configuration["run"].removeprefix('python -c "').removesuffix('"')
    command = command.replace('\\"', '"')
    subprocess.run([sys.executable, "-c", command], check=True,
                   env=os.environ | {"RUNNER_TEMP": str(temporary), "GITHUB_ENV": str(github_env)})
    wheels = temporary / "package-check-wheels"
    assert wheels.is_dir()
    assert github_env.read_text() == f"PIP_FIND_LINKS={wheels.as_uri()}\n"
    assert all(option in provision["run"] for option in (
        "--isolated download", "--no-cache-dir", "--only-binary=:all:", "--require-hashes",
        "--index-url https://pypi.org/simple", "-r tools/requirements-package-check-hashed.lock",
        '--dest "${{ runner.temp }}/package-check-wheels"',
    ))
