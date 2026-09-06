#!/usr/bin/env python3
"""Verify that documented builds succeed from a committed, clean export."""

import argparse
import os
from pathlib import Path, PurePosixPath
import shutil
import subprocess
import sys
import tarfile
import tempfile


def hashes(root, names):
    import hashlib
    return {name: hashlib.sha256((root / name).read_bytes()).hexdigest()
            for name in names if (root / name).is_file()}


def tracked_files(repository: Path, ref: str) -> list[str]:
    result = subprocess.run(
        ["git", "ls-tree", "-r", "--name-only", "-z", ref],
        cwd=repository,
        check=True,
        stdout=subprocess.PIPE,
    )
    return [os.fsdecode(name) for name in result.stdout.split(b"\0") if name]


def validate_archive(archive: tarfile.TarFile, destination: Path) -> None:
    destination = destination.resolve()
    for member in archive.getmembers():
        name = PurePosixPath(member.name)
        if name.is_absolute() or ".." in name.parts:
            raise ValueError(f"unsafe archive path: {member.name}")

        target = (destination / member.name).resolve(strict=False)
        if not target.is_relative_to(destination):
            raise ValueError(f"archive path escapes destination: {member.name}")

        if member.issym():
            link_target = (target.parent / member.linkname).resolve(strict=False)
            if not link_target.is_relative_to(destination):
                raise ValueError(f"archive link escapes destination: {member.name}")
        elif member.islnk():
            link_target = (destination / member.linkname).resolve(strict=False)
            if not link_target.is_relative_to(destination):
                raise ValueError(f"archive link escapes destination: {member.name}")
        elif not (member.isfile() or member.isdir()):
            raise ValueError(f"unsupported archive member: {member.name}")


def export_ref(repository: Path, ref: str, archive_path: Path, destination: Path) -> None:
    with archive_path.open("wb") as archive_file:
        subprocess.run(
            ["git", "archive", "--format=tar", ref],
            cwd=repository,
            stdout=archive_file,
            check=True,
        )
    with tarfile.open(archive_path, mode="r:") as archive:
        validate_archive(archive, destination)
        archive.extractall(destination)


def windows_build(source: Path, env: dict[str, str]) -> None:
    build = source / "build"
    build.mkdir()
    commands = [
        ["go", "build", "-mod=readonly", "-trimpath", "-o", "build/spl-toolkit", "./cmd"],
        ["go", "build", "-mod=readonly", "-trimpath", "-o", "build/spl-toolkit-server", "./cmd/server"],
        [
            "go", "build", "-mod=readonly", "-trimpath", "-buildmode=c-shared",
            "-o", "build/libspl_toolkit.dll", "./pkg/bindings",
        ],
        ["go", "test", "-mod=readonly", "-race", "./..."],
    ]
    for command in commands:
        subprocess.run(command, cwd=source, env=env, check=True)


def check(repository: Path, ref: str) -> None:
    tracked_names = tracked_files(repository, ref)
    with tempfile.TemporaryDirectory(prefix="spl-toolkit-clean-build-") as directory:
        temporary = Path(directory)
        source = temporary / "source"
        source.mkdir()
        export_ref(repository, ref, temporary / "source.tar", source)

        env = os.environ.copy()
        env["GOCACHE"] = str(temporary / "go-build-cache")
        env["GOMODCACHE"] = str(temporary / "go-module-cache")
        env["GOTOOLCHAIN"] = "local"

        before = hashes(source, tracked_names)
        if os.name == "nt" and shutil.which("make") is None:
            windows_build(source, env)
        else:
            subprocess.run(["make", "build", "build-server", "build-shared", "test"],
                           cwd=source, env=env, check=True)
        try:
            assert hashes(source, tracked_names) == before, "build changed tracked source"
        except AssertionError as error:
            after = hashes(source, tracked_names)
            changed = sorted(
                name for name in set(before) | set(after)
                if before.get(name) != after.get(name)
            )
            raise AssertionError(f"{error}: {', '.join(changed)}") from error


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--ref", required=True, help="committed Git ref to export and check")
    args = parser.parse_args()

    try:
        repository = Path(
            subprocess.run(
                ["git", "rev-parse", "--show-toplevel"],
                check=True,
                text=True,
                stdout=subprocess.PIPE,
            ).stdout.strip()
        )
        check(repository, args.ref)
    except subprocess.CalledProcessError as error:
        print(f"build command failed with exit code {error.returncode}", file=sys.stderr)
        return error.returncode or 1
    except (AssertionError, OSError, tarfile.TarError, ValueError) as error:
        print(error, file=sys.stderr)
        return 1

    print(f"clean build verified for {args.ref}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
