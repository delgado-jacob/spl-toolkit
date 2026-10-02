"""Focused integrity tests for committed-source export checks."""

from pathlib import Path

import pytest

from tools.check_clean_build import hashes


def test_tracked_hashes_reject_missing_files(tmp_path: Path):
    (tmp_path / "present.txt").write_text("present\n", encoding="utf-8")

    with pytest.raises(FileNotFoundError, match="tracked file is missing: missing.txt"):
        hashes(tmp_path, ["present.txt", "missing.txt"])


def test_windows_clean_build_propagates_exporter_compile_failure(tmp_path, monkeypatch):
    from tools import check_clean_build as checker
    import subprocess
    seen = []
    def run(command, **kwargs):
        seen.append(command)
        if command[-1] == "./cmd/splunk-export":
            raise subprocess.CalledProcessError(7, command)
    monkeypatch.setattr(checker.subprocess, "run", run)
    with pytest.raises(subprocess.CalledProcessError) as error:
        checker.windows_build(tmp_path, {})
    assert error.value.returncode == 7
    assert seen[-1][-1] == "./cmd/splunk-export"
    assert not any(command[1] == "test" for command in seen)
