"""Focused integrity tests for committed-source export checks."""

from pathlib import Path

import pytest

from tools.check_clean_build import hashes


def test_tracked_hashes_reject_missing_files(tmp_path: Path):
    (tmp_path / "present.txt").write_text("present\n", encoding="utf-8")

    with pytest.raises(FileNotFoundError, match="tracked file is missing: missing.txt"):
        hashes(tmp_path, ["present.txt", "missing.txt"])
