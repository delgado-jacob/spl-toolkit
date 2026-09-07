"""Tests for controlled, reproducible release artifacts."""

from __future__ import annotations

import gzip
import hashlib
import io
from pathlib import Path
import tarfile
import zipfile

import pytest

from tools.check_reproducible import compare_artifacts
from tools.release import artifact_hashes, normalize_archive, require_python_archives, verify_wheel_native


EPOCH = 1788652800


def _write_wheel(path: Path, stamp: tuple[int, int, int, int, int, int], payload: bytes) -> None:
    with zipfile.ZipFile(path, "w") as archive:
        info = zipfile.ZipInfo("spl_toolkit/native.dll", stamp)
        info.external_attr = 0o600 << 16
        archive.writestr(info, payload)


def _write_sdist(path: Path, *, stamp: int, reverse: bool) -> None:
    members = [("pkg/module.py", b"value = 1\n"), ("pkg/README.md", b"same\n")]
    if reverse:
        members.reverse()
    raw = io.BytesIO()
    with tarfile.open(fileobj=raw, mode="w", format=tarfile.PAX_FORMAT) as archive:
        for index, (name, data) in enumerate(members):
            info = tarfile.TarInfo(name)
            info.size = len(data)
            info.mtime = stamp + index
            info.uid = 501 + index
            info.gid = 20 + index
            info.uname = "developer"
            info.gname = "staff"
            info.pax_headers = {"atime": str(stamp), "ctime": str(stamp)}
            archive.addfile(info, io.BytesIO(data))
    with path.open("wb") as destination:
        with gzip.GzipFile(filename="source-name", mode="wb", fileobj=destination, mtime=stamp) as compressed:
            compressed.write(raw.getvalue())


def test_wheel_metadata_is_reproducible(tmp_path: Path):
    first, second = tmp_path / "first.whl", tmp_path / "second.whl"
    _write_wheel(first, (2020, 1, 1, 0, 0, 0), b"same-native-payload")
    _write_wheel(second, (2025, 2, 2, 2, 2, 2), b"same-native-payload")

    normalize_archive(first, EPOCH)
    normalize_archive(second, EPOCH)

    assert first.read_bytes() == second.read_bytes()


def test_sdist_metadata_and_input_order_are_reproducible(tmp_path: Path):
    first, second = tmp_path / "first.tar.gz", tmp_path / "second.tar.gz"
    _write_sdist(first, stamp=1_600_000_000, reverse=False)
    _write_sdist(second, stamp=1_700_000_000, reverse=True)

    normalize_archive(first, EPOCH)
    normalize_archive(second, EPOCH)

    assert first.read_bytes() == second.read_bytes()
    with tarfile.open(first, "r:gz") as archive:
        assert archive.getnames() == sorted(archive.getnames())
        assert all(member.mtime == EPOCH for member in archive.getmembers())
        assert all((member.uid, member.gid, member.uname, member.gname) == (0, 0, "", "") for member in archive.getmembers())


@pytest.mark.parametrize("name", ["../escape", "/absolute", "pkg/../../escape"])
def test_archive_normalization_rejects_unsafe_member_names(tmp_path: Path, name: str):
    archive_path = tmp_path / "unsafe.whl"
    with zipfile.ZipFile(archive_path, "w") as archive:
        archive.writestr(name, b"payload")

    with pytest.raises(ValueError, match="unsafe archive path"):
        normalize_archive(archive_path, EPOCH)


def test_changed_native_payload_fails_comparison_after_normalization(tmp_path: Path):
    first, second = tmp_path / "first", tmp_path / "second"
    first.mkdir()
    second.mkdir()
    _write_wheel(first / "spl_toolkit-0.1.1-py3-none-any.whl", (2020, 1, 1, 0, 0, 0), b"native-one")
    _write_wheel(second / "spl_toolkit-0.1.1-py3-none-any.whl", (2025, 2, 2, 2, 2, 2), b"native-two")
    normalize_archive(next(first.glob("*.whl")), EPOCH)
    normalize_archive(next(second.glob("*.whl")), EPOCH)

    with pytest.raises(RuntimeError, match="non-reproducible artifacts: spl_toolkit"):
        compare_artifacts(first, second)


def test_differing_artifact_sets_fail_comparison(tmp_path: Path):
    first, second = tmp_path / "first", tmp_path / "second"
    first.mkdir()
    second.mkdir()
    (first / "cli").write_bytes(b"same")

    with pytest.raises(RuntimeError, match="release artifact sets differ"):
        compare_artifacts(first, second)


def test_hashes_include_nested_payloads_and_exclude_only_checksum(tmp_path: Path):
    (tmp_path / "nested").mkdir()
    (tmp_path / "nested" / "payload").write_bytes(b"payload")
    (tmp_path / "SHA256SUMS").write_text("not recursively hashed\n", encoding="utf-8")

    assert artifact_hashes(tmp_path) == {
        "nested/payload": hashlib.sha256(b"payload").hexdigest(),
    }


def test_release_requires_exactly_one_wheel_and_sdist(tmp_path: Path):
    (tmp_path / "spl-toolkit-0.1.1-darwin-arm64").write_bytes(b"cli")

    with pytest.raises(RuntimeError, match="exactly one wheel"):
        require_python_archives(tmp_path)

    (tmp_path / "spl_toolkit-0.1.1-py3-none-any.whl").write_bytes(b"wheel")
    with pytest.raises(RuntimeError, match="exactly one sdist"):
        require_python_archives(tmp_path)


def test_wheel_contains_the_unchanged_standalone_native_payload(tmp_path: Path):
    wheel = tmp_path / "spl_toolkit-0.1.1-py3-none-any.whl"
    native = tmp_path / "libspl_toolkit-0.1.1-linux-amd64.so"
    native.write_bytes(b"same-native")
    with zipfile.ZipFile(wheel, "w") as archive:
        archive.writestr("spl_toolkit/libspl_toolkit.so", b"same-native")

    verify_wheel_native(wheel, native)

    native.write_bytes(b"different-native")
    with pytest.raises(RuntimeError, match="wheel native payload differs"):
        verify_wheel_native(wheel, native)
