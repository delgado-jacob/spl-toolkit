"""Wrapper guard and release closure tests; these do not substitute for real consumers."""
from __future__ import annotations
import json
from pathlib import Path
import tarfile

import pytest

from tools import check_editor_client, check_sarif_consumer, release

ROOT = Path(__file__).resolve().parents[2]


@pytest.mark.parametrize("sha", ["HEAD", "1234567", "g" * 40, "a" * 41])
def test_consumers_reject_unbound_source_identity(tmp_path, sha):
    with pytest.raises(ValueError, match="full Git SHA"):
        check_editor_client.run(tmp_path, tmp_path, tmp_path, sha, tmp_path / "editor.json")
    with pytest.raises(ValueError, match="full Git SHA"):
        check_sarif_consumer.run(tmp_path, tmp_path, sha, tmp_path / "sarif.json")


def test_editor_requires_real_offline_client_install(tmp_path):
    cli = tmp_path / "cli"
    cli.touch()
    with pytest.raises(ValueError, match="installation missing"):
        check_editor_client.run(cli, cli, tmp_path, "a" * 40, tmp_path / "evidence.json")
    assert not (tmp_path / "evidence.json").exists()


def test_reviewed_consumer_lock_closures():
    fixture = ROOT / "tests/editor-client"
    packages = json.loads((fixture / "package-lock.json").read_text())["packages"]
    assert packages["node_modules/vscode-languageclient"]["version"] == "9.0.1"
    assert packages["node_modules/vscode-languageserver-protocol"]["version"] == "3.17.5"
    for name, package in packages.items():
        if name:
            assert package["resolved"].startswith("https://registry.npmjs.org/")
            assert package["integrity"].startswith("sha512-")
    provenance = json.loads((fixture / "consumer-provenance.json").read_text())
    lock = (fixture / "requirements-sarif-local-hashed.lock").read_text()
    for wheel in provenance["wheels"]:
        assert f'{wheel["name"]}=={wheel["version"]} --hash=sha256:{wheel["sha256"]}' in lock
        assert wheel["metadata_url"].startswith("https://pypi.org/pypi/")
        assert wheel["url"].startswith("https://files.pythonhosted.org/")


def test_release_source_and_content_are_explicit_and_reproducible(tmp_path):
    first, second = tmp_path / "first", tmp_path / "second"
    first.mkdir()
    second.mkdir()
    release.package_tooling_content(ROOT, first, "0.1.1", 1788652800)
    release.package_tooling_content(ROOT, second, "0.1.1", 1788652800)
    assert release.artifact_hashes(first) == release.artifact_hashes(second)
    source = first / "spl-toolkit-source-0.1.1.tar.gz"
    with tarfile.open(source) as archive:
        names = set(archive.getnames())
        for path in ("cmd/tooling_cli.go", "internal/lsp/server.go", "internal/corpusfs/open_windows.go",
                     "go.mod", "go.sum", "contracts/sarif/LICENSE.txt", "examples/tooling/native.py"):
            assert path in names
        assert not any(name.startswith(("_build_plan/", ".git/", ".superpowers/")) for name in names)
        assert archive.extractfile("cmd/tooling_cli.go").read() == (ROOT / "cmd/tooling_cli.go").read_bytes()
    assert (first / "docs/tooling.md").read_bytes() == (ROOT / "docs/tooling.md").read_bytes()
    assert (first / "contracts/sarif/LICENSE.txt").is_file()


def test_release_manifest_rejects_missing_input(tmp_path):
    (tmp_path / "tools").mkdir()
    (tmp_path / "tools/release-content-files.txt").write_text("missing.md\n")
    (tmp_path / "tools/release-source-files.txt").write_text("go.mod\n")
    output = tmp_path / "out"
    output.mkdir()
    with pytest.raises((ValueError, FileNotFoundError)):
        release.package_tooling_content(tmp_path, output, "0.1.1", 1788652800)
