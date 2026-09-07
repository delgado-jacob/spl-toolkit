#!/usr/bin/env python3
"""Build the small, controlled SPL Toolkit release payload set."""

from __future__ import annotations

import argparse
import copy
import datetime as dt
import gzip
import hashlib
import importlib.util
import io
import json
import os
from pathlib import Path, PurePosixPath
import platform
import shutil
import subprocess
import sys
import tarfile
import zipfile
import zlib


ROOT = Path(__file__).resolve().parents[1]
ENVIRONMENT_FILE = ROOT / "tools" / "release-env.json"
CHECKSUM_FILE = "SHA256SUMS"
VERSION_SYMBOL = "github.com/delgado-jacob/spl-toolkit/internal/buildinfo.Version"


def _safe_archive_name(name: str) -> None:
    path = PurePosixPath(name.replace("\\", "/"))
    if path.is_absolute() or not path.parts or any(part in ("", ".", "..") for part in path.parts):
        raise ValueError(f"unsafe archive path: {name}")


def _zip_timestamp(epoch: int) -> tuple[int, int, int, int, int, int]:
    value = dt.datetime.fromtimestamp(epoch, tz=dt.timezone.utc)
    if value.year < 1980:
        value = value.replace(year=1980, month=1, day=1, hour=0, minute=0, second=0)
    if value.year > 2107:
        raise ValueError("archive epoch is outside the ZIP timestamp range")
    return value.year, value.month, value.day, value.hour, value.minute, value.second


def _normalize_zip(path: Path, epoch: int) -> None:
    entries: list[tuple[zipfile.ZipInfo, bytes]] = []
    with zipfile.ZipFile(path, "r") as source:
        for original in source.infolist():
            _safe_archive_name(original.filename)
            entries.append((original, source.read(original)))
    temporary = path.with_name(path.name + ".normalize.tmp")
    try:
        with zipfile.ZipFile(temporary, "w", compression=zipfile.ZIP_DEFLATED, compresslevel=9) as target:
            for original, data in sorted(entries, key=lambda item: item[0].filename):
                info = zipfile.ZipInfo(original.filename, _zip_timestamp(epoch))
                info.compress_type = zipfile.ZIP_DEFLATED
                info.create_system = 3
                info.flag_bits = 0
                info.comment = b""
                executable = bool((original.external_attr >> 16) & 0o111)
                mode = 0o755 if original.is_dir() or executable else 0o644
                file_type = 0o040000 if original.is_dir() else 0o100000
                info.external_attr = (file_type | mode) << 16
                target.writestr(info, data, compress_type=zipfile.ZIP_DEFLATED, compresslevel=9)
        temporary.replace(path)
    finally:
        temporary.unlink(missing_ok=True)


def _normalize_tar_gz(path: Path, epoch: int) -> None:
    entries: list[tuple[tarfile.TarInfo, bytes | None]] = []
    with tarfile.open(path, "r:gz") as source:
        for original in source.getmembers():
            _safe_archive_name(original.name)
            if original.issym() or original.islnk():
                parent = PurePosixPath(original.name).parent
                target = PurePosixPath(original.linkname.replace("\\", "/"))
                resolved = target if target.is_absolute() else parent / target
                if target.is_absolute() or ".." in resolved.parts:
                    raise ValueError(f"unsafe archive link: {original.name} -> {original.linkname}")
            extracted = source.extractfile(original) if original.isfile() else None
            entries.append((original, extracted.read() if extracted is not None else None))
    raw = io.BytesIO()
    with tarfile.open(fileobj=raw, mode="w", format=tarfile.PAX_FORMAT) as target:
        for original, data in sorted(entries, key=lambda item: item[0].name):
            info = copy.copy(original)
            info.uid = info.gid = 0
            info.uname = info.gname = ""
            info.mtime = epoch
            info.pax_headers = {}
            if info.isdir():
                info.mode = 0o755
            elif info.isfile():
                info.mode = 0o755 if info.mode & 0o111 else 0o644
            elif info.issym() or info.islnk():
                info.mode = 0o777
            target.addfile(info, io.BytesIO(data) if data is not None else None)
    temporary = path.with_name(path.name + ".normalize.tmp")
    try:
        with temporary.open("wb") as destination:
            with gzip.GzipFile(filename="", mode="wb", fileobj=destination, mtime=epoch, compresslevel=9) as compressed:
                compressed.write(raw.getvalue())
        temporary.replace(path)
    finally:
        temporary.unlink(missing_ok=True)


def normalize_archive(path: Path, epoch: int) -> None:
    """Atomically normalize ZIP/wheel or gzip/tar archive metadata."""
    path = Path(path)
    if path.suffix == ".whl" or zipfile.is_zipfile(path):
        _normalize_zip(path, epoch)
    elif path.name.endswith(".tar.gz"):
        _normalize_tar_gz(path, epoch)
    else:
        raise ValueError(f"unsupported release archive: {path}")


def artifact_hashes(directory: Path) -> dict[str, str]:
    """Hash every release payload by its relative POSIX path."""
    directory = Path(directory)
    return {
        path.relative_to(directory).as_posix(): hashlib.sha256(path.read_bytes()).hexdigest()
        for path in sorted(directory.rglob("*"))
        if path.is_file() and path.relative_to(directory).as_posix() != CHECKSUM_FILE
    }


def require_python_archives(output: Path) -> tuple[Path, Path]:
    wheels = sorted(output.glob("spl_toolkit-*.whl"))
    if len(wheels) != 1:
        raise RuntimeError(f"release requires exactly one wheel; found {len(wheels)}")
    sdists = sorted(output.glob("spl_toolkit-*.tar.gz"))
    if len(sdists) != 1:
        raise RuntimeError(f"release requires exactly one sdist; found {len(sdists)}")
    return wheels[0], sdists[0]


def verify_wheel_native(wheel: Path, native: Path) -> None:
    with zipfile.ZipFile(wheel, "r") as archive:
        members = [
            name for name in archive.namelist()
            if name.startswith("spl_toolkit/libspl_toolkit") and Path(name).suffix in (".so", ".dylib", ".dll")
        ]
        if len(members) != 1:
            raise RuntimeError(f"wheel must contain exactly one native payload; found {len(members)}")
        packaged = archive.read(members[0])
    if packaged != native.read_bytes():
        raise RuntimeError("wheel native payload differs from standalone native payload")


def _load_build_support(source: Path):
    path = source / "python" / "build_support.py"
    spec = importlib.util.spec_from_file_location("spl_toolkit_release_build_support", path)
    if spec is None or spec.loader is None:
        raise RuntimeError(f"could not load {path}")
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


def _run(command: list[str], *, cwd: Path, env: dict[str, str], log: Path) -> str:
    completed = subprocess.run(command, cwd=cwd, env=env, text=True, capture_output=True, check=False)
    log.parent.mkdir(parents=True, exist_ok=True)
    log.write_text(
        "$ " + " ".join(command) + "\n" + completed.stdout + completed.stderr,
        encoding="utf-8",
    )
    if completed.returncode:
        raise RuntimeError(f"command failed ({completed.returncode}); see {log}")
    return completed.stdout.strip()


def _target_name() -> str:
    system = platform.system().lower()
    machine = platform.machine().lower()
    arch = "amd64" if machine in ("amd64", "x86_64") else "arm64" if machine in ("arm64", "aarch64") else machine
    if system == "windows":
        return "windows-amd64" if arch == "amd64" else f"windows-{arch}"
    return f"{system}-{arch}"


def _version_output(command: list[str]) -> str:
    return subprocess.run(command, check=True, text=True, capture_output=True).stdout.strip()


def validate_environment(config: dict, target: str) -> dict[str, object]:
    expected = config["targets"].get(target)
    if expected is None:
        raise RuntimeError(f"unsupported release target: {target}")
    go_version = _version_output(["go", "version"])
    if f"go{config['go']}" not in go_version.split():
        raise RuntimeError(f"Go version mismatch: expected {config['go']}, got {go_version}")
    python_version = platform.python_version()
    if python_version != config["build_python"]:
        raise RuntimeError(f"Python version mismatch: expected {config['build_python']}, got {python_version}")
    configured_arch = os.environ.get("GOARCH", expected["goarch"])
    if configured_arch != expected["goarch"]:
        raise RuntimeError(f"GOARCH mismatch: expected {expected['goarch']}, got {configured_arch}")
    cc = shutil.which(os.environ.get("CC", expected["cc"]))
    if cc is None:
        raise RuntimeError(f"required compiler not found: {expected['cc']}")
    compiler_version = subprocess.run([cc, "--version"], check=True, text=True, capture_output=True).stdout.strip()
    if target.startswith("windows-"):
        machine = _version_output([cc, "-dumpmachine"])
        version = _version_output([cc, "-dumpfullversion"])
        if machine != "x86_64-w64-mingw32" or version != expected["gcc_version"]:
            raise RuntimeError(f"MinGW compiler mismatch: expected x86_64-w64-mingw32 {expected['gcc_version']}, got {machine} {version}")
    sdk = None
    linker = None
    if target.startswith("darwin-"):
        developer_dir = os.environ.get("DEVELOPER_DIR")
        if developer_dir != expected["developer_dir"]:
            raise RuntimeError(f"DEVELOPER_DIR mismatch: expected {expected['developer_dir']}, got {developer_dir!r}")
        if os.environ.get("MACOSX_DEPLOYMENT_TARGET") != expected["deployment_target"]:
            raise RuntimeError(f"MACOSX_DEPLOYMENT_TARGET must be {expected['deployment_target']}")
        sdk = _version_output(["xcrun", "--show-sdk-path"])
        linker = _version_output(["xcrun", "ld", "-v"])
    return {
        "target": target,
        "runner": expected["runner"],
        "runner_image_os": os.environ.get("ImageOS", "unavailable"),
        "runner_image_version": os.environ.get("ImageVersion", "unavailable"),
        "os": platform.platform(),
        "architecture": platform.machine(),
        "go": go_version,
        "python": python_version,
        "packaging": {
            name: __import__(name).__version__ for name in ("setuptools", "wheel", "build", "packaging")
        },
        "cc": str(Path(cc).resolve()),
        "cc_version": compiler_version,
        "linker": linker,
        "sdk": sdk,
        "zlib": zlib.ZLIB_VERSION,
        "wheel_platform": expected["wheel_platform"],
    }


def _new_directory(path: Path) -> None:
    if path.exists() and (not path.is_dir() or any(path.iterdir())):
        raise RuntimeError(f"release output must be new or empty: {path}")
    path.mkdir(parents=True, exist_ok=True)


def build_release(source: Path, output: Path, epoch: int) -> dict[str, object]:
    source, output = source.resolve(), output.resolve()
    if not (source / "go.mod").is_file() or not (source / "VERSION").is_file():
        raise RuntimeError(f"release source is incomplete: {source}")
    _new_directory(output)
    evidence = output.parent / f"{output.name}-evidence"
    _new_directory(evidence)
    config = json.loads((source / "tools" / "release-env.json").read_text(encoding="utf-8"))
    target = _target_name()
    environment = validate_environment(config, target)
    support = _load_build_support(source)
    support.native_source_files(source / "python" / support.NATIVE_SOURCE_MANIFEST)
    version = (source / "VERSION").read_text(encoding="utf-8").strip()
    if version != support.read_version(source / "python"):
        raise RuntimeError("root and Python package versions differ")
    env = os.environ.copy()
    env.update({
        "SOURCE_DATE_EPOCH": str(epoch),
        "GOTOOLCHAIN": "local",
        "GOARCH": config["targets"][target]["goarch"],
        "CC": str(environment["cc"]),
    })
    if target.startswith("darwin-"):
        env["_PYTHON_HOST_PLATFORM"] = config["targets"][target]["wheel_platform"].replace("_", "-")
    executable_suffix = ".exe" if target.startswith("windows-") else ""
    version_flag = f"-X={VERSION_SYMBOL}={version}"
    go_ldflags = f"-buildid= {version_flag}"
    go_base = ["go", "build", "-mod=readonly", "-trimpath", "-buildvcs=false", "-ldflags", go_ldflags]
    cli = output / f"spl-toolkit-{version}-{target}{executable_suffix}"
    server = output / f"spl-toolkit-server-{version}-{target}{executable_suffix}"
    cgo0 = env | {"CGO_ENABLED": "0"}
    _run(go_base + ["-o", str(cli), "./cmd"], cwd=source, env=cgo0, log=evidence / "cli-build.log")
    _run(go_base + ["-o", str(server), "./cmd/server"], cwd=source, env=cgo0, log=evidence / "server-build.log")
    native_suffix = Path(support.native_library_name()).suffix
    native = output / f"libspl_toolkit-{version}-{target}{native_suffix}"
    native_build = evidence / "native" / support.native_library_name()
    prior = os.environ.copy()
    try:
        os.environ.clear()
        os.environ.update(env)
        command, native_env = support._native_build_plan(source, native_build, version)
        (evidence / "native-build.log").write_text("argv: " + json.dumps(command) + "\n", encoding="utf-8")
        support.build_native(source, native_build, version)
    finally:
        os.environ.clear()
        os.environ.update(prior)
    native.parent.mkdir(parents=True, exist_ok=True)
    shutil.copyfile(native_build, native)
    header = native.with_suffix(".h")
    shutil.copyfile(native_build.with_suffix(".h"), header)
    shutil.copyfile(source / "LICENSE", output / "LICENSE")
    _run(
        [sys.executable, "-m", "build", "--no-isolation", "--sdist", "--wheel", "--outdir", str(output), "python"],
        cwd=source, env=env, log=evidence / "python-build.log",
    )
    wheel, sdist = require_python_archives(output)
    expected_wheel_platform = config["targets"][target]["wheel_platform"]
    if expected_wheel_platform not in wheel.name:
        raise RuntimeError(f"wheel platform mismatch: expected {expected_wheel_platform}, got {wheel.name}")
    verify_wheel_native(wheel, native)
    normalize_archive(wheel, epoch)
    normalize_archive(sdist, epoch)
    hashes = artifact_hashes(output)
    (output / CHECKSUM_FILE).write_text(
        "".join(f"{digest}  {name}\n" for name, digest in hashes.items()), encoding="utf-8"
    )
    environment.update({
        "source_date_epoch": epoch,
        "goflags": ["-mod=readonly", "-trimpath", "-buildvcs=false", "-buildid=", version_flag],
        "cgo_enabled_cli_server": "0",
        "cgo_enabled_native": "1",
        "native_argv": command,
        "cgo_cflags": native_env.get("CGO_CFLAGS", ""),
        "cgo_cppflags": native_env.get("CGO_CPPFLAGS", ""),
        "cgo_ldflags": native_env.get("CGO_LDFLAGS", ""),
        "artifacts": hashes,
    })
    (evidence / "environment.json").write_text(json.dumps(environment, indent=2, sort_keys=True) + "\n", encoding="utf-8")
    return environment


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--source", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    parser.add_argument("--epoch", type=int, required=True)
    args = parser.parse_args()
    build_release(args.source, args.output, args.epoch)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
