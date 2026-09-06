#!/usr/bin/env python3
"""Build-independent acceptance checks for SPL Toolkit wheel and sdist artifacts."""

from __future__ import annotations

import argparse
from email.parser import Parser
import os
from pathlib import Path, PurePath
import platform
import shutil
import struct
import subprocess
import sys
import sysconfig
import tarfile
import tempfile
import textwrap
import venv
import zipfile

from packaging.utils import parse_wheel_filename


NATIVE_SUFFIXES = (".so", ".dylib", ".dll")
SDIST_FIXED_FILES = {
    "LICENSE", "MANIFEST.in", "PKG-INFO", "README.md", "VERSION", "build_support.py",
    "native-source-files.txt", "pyproject.toml", "requirements-build.txt",
    "requirements-dev.txt", "setup.cfg", "setup.py", "spl_toolkit/__init__.py",
    "spl_toolkit/exceptions.py", "spl_toolkit/libspl_toolkit.h", "spl_toolkit/mapper.py",
    "spl_toolkit.egg-info/PKG-INFO", "spl_toolkit.egg-info/SOURCES.txt",
    "spl_toolkit.egg-info/dependency_links.txt", "spl_toolkit.egg-info/top_level.txt",
    "tests/test_mapper.py", "tests/test_native_abi.py", "tests/test_native_mapper.py",
}
INSTALL_SCRIPT = """
import importlib.metadata, pathlib, sys
import spl_toolkit
from spl_toolkit import SPLMapper
assert pathlib.Path(spl_toolkit.__file__).resolve().is_relative_to(pathlib.Path(sys.prefix).resolve())
with SPLMapper() as mapper:
    assert isinstance(mapper.native_version, str)
    assert mapper.native_version == spl_toolkit.__version__
    assert spl_toolkit.__version__ == importlib.metadata.version('spl-toolkit')
    mapper.load_mappings([{'source':'src_ip','target':'source_ip'}])
    assert mapper.map_query('search src_ip=1') == 'search source_ip=1'
"""


def run(command: list[str], *, cwd: Path, env: dict[str, str] | None = None) -> subprocess.CompletedProcess[str]:
    return subprocess.run(command, cwd=cwd, env=env, check=True, text=True)


def clean_env() -> dict[str, str]:
    env = os.environ.copy()
    for name in ("PYTHONPATH", "PYTHONHOME", "SPL_NATIVE_LIBRARY", "SPL_EXPECTED_VERSION"):
        env.pop(name, None)
    return env


def native_library_name() -> str:
    if os.name == "nt":
        return "libspl_toolkit.dll"
    return "libspl_toolkit.dylib" if sys.platform == "darwin" else "libspl_toolkit.so"


def native_architecture(payload: bytes) -> str:
    if payload.startswith(b"\x7fELF"):
        endian = "<" if payload[5] == 1 else ">"
        machine = struct.unpack_from(f"{endian}H", payload, 18)[0]
        return {62: "x86_64", 183: "arm64"}.get(machine, f"elf-{machine}")
    if payload[:4] in (b"\xcf\xfa\xed\xfe", b"\xfe\xed\xfa\xcf"):
        endian = "<" if payload[:4] == b"\xcf\xfa\xed\xfe" else ">"
        cpu = struct.unpack_from(f"{endian}I", payload, 4)[0]
        return {0x01000007: "x86_64", 0x0100000C: "arm64"}.get(cpu, f"macho-{cpu}")
    if payload.startswith(b"MZ"):
        pe_offset = struct.unpack_from("<I", payload, 0x3C)[0]
        if payload[pe_offset:pe_offset + 4] != b"PE\0\0":
            raise AssertionError("invalid PE header")
        machine = struct.unpack_from("<H", payload, pe_offset + 4)[0]
        return {0x8664: "x86_64", 0xAA64: "arm64"}.get(machine, f"pe-{machine}")
    raise AssertionError("unrecognized native library format")


def venv_python(directory: Path) -> Path:
    return directory / ("Scripts/python.exe" if os.name == "nt" else "bin/python")


def create_test_environment(directory: Path) -> Path:
    venv.EnvBuilder(with_pip=True, symlinks=os.name != "nt").create(directory)
    return venv_python(directory)


def install_and_check(
    wheel: Path,
    directory: Path,
    outside_checkout: Path,
    expected_version: str,
    requirements: Path,
) -> None:
    python = create_test_environment(directory)
    env = clean_env()
    path_entries = [str(python.parent)]
    if os.name == "nt":
        path_entries.append(str(Path(os.environ.get("SystemRoot", r"C:\Windows")) / "System32"))
    else:
        path_entries.extend(("/usr/bin", "/bin"))
    install_env = env | {"PATH": os.pathsep.join(path_entries)}
    run([str(python), "-m", "pip", "install", "--disable-pip-version-check", "-r", str(requirements.resolve())], cwd=outside_checkout, env=install_env)
    run([str(python), "-m", "pip", "install", "--disable-pip-version-check", "--no-deps", str(wheel.resolve())], cwd=outside_checkout, env=install_env)
    controller_site = str(Path(sysconfig.get_paths()["purelib"]).resolve())
    check_script = (
        f"import importlib.metadata, pathlib, sys\nassert {expected_version!r} == importlib.metadata.version('spl-toolkit')\n"
        f"assert {controller_site!r} not in [str(pathlib.Path(p).resolve()) for p in sys.path]\n"
        + INSTALL_SCRIPT
    )
    run([str(python), "-I", "-c", textwrap.dedent(check_script)], cwd=outside_checkout, env=install_env)

    installed_test_dir = outside_checkout / f"tests-{directory.name}"
    shutil.copytree(Path(__file__).resolve().parents[1] / "python" / "tests", installed_test_dir)
    run([str(python), "-I", "-m", "pytest", str(installed_test_dir), "-q"], cwd=outside_checkout, env=install_env)


def inspect_wheel(wheel: Path, expected_version: str) -> None:
    name, version, _build, tags = parse_wheel_filename(wheel.name)
    if name != "spl-toolkit" or str(version) != expected_version:
        raise AssertionError(f"unexpected wheel identity: {name} {version}")
    if any(tag.platform == "any" or tag.interpreter != "py3" or tag.abi != "none" for tag in tags):
        raise AssertionError(f"wheel must use py3-none-platform tag: {sorted(map(str, tags))}")

    with zipfile.ZipFile(wheel) as archive:
        names = archive.namelist()
        native = [name for name in names if name.startswith("spl_toolkit/libspl_toolkit") and name.endswith(NATIVE_SUFFIXES)]
        if len(native) != 1:
            raise AssertionError(f"wheel contains {len(native)} native libraries")
        metadata_name = next(name for name in names if name.endswith(".dist-info/METADATA"))
        metadata = Parser().parsestr(archive.read(metadata_name).decode())
        if metadata["Requires-Python"] != ">=3.11":
            raise AssertionError(f"unexpected Requires-Python: {metadata['Requires-Python']}")
        if metadata["Version"] != expected_version:
            raise AssertionError(f"unexpected metadata version: {metadata['Version']}")
        if metadata["Description-Content-Type"] != "text/markdown":
            raise AssertionError("README metadata is missing")
        if metadata["License-Expression"] != "MIT":
            raise AssertionError("license metadata is missing")

        payload = archive.read(native[0])
        machine = platform.machine().lower()
        expected = "arm64" if machine in ("arm64", "aarch64") else "x86_64"
        actual = native_architecture(payload)
        if actual != expected:
            raise AssertionError(f"native architecture {actual} does not match {machine}")

        with tempfile.TemporaryDirectory(prefix="spl-wheel-native-") as temp:
            native_path = Path(temp) / Path(native[0]).name
            native_path.write_bytes(payload)
            if sys.platform == "darwin":
                required_platform = f"macosx_15_0_{actual}"
                platforms = {tag.platform for tag in tags}
                if platforms != {required_platform}:
                    raise AssertionError(
                        f"wheel platform {sorted(platforms)} does not match native minimum/architecture {required_platform}"
                    )
                load_commands = subprocess.run(
                    ["otool", "-l", str(native_path)], check=True, capture_output=True, text=True
                ).stdout
                if "minos 15.0" not in load_commands:
                    raise AssertionError("native library does not declare macOS 15.0 minimum deployment")


def unpack_sdist(sdist: Path, destination: Path) -> Path:
    with tarfile.open(sdist, "r:gz") as archive:
        members = archive.getmembers()
        if any(Path(member.name).is_absolute() or ".." in Path(member.name).parts for member in members):
            raise AssertionError("sdist contains an unsafe path")
        archive.extractall(destination)
    roots = [entry for entry in destination.iterdir() if entry.is_dir()]
    if len(roots) != 1:
        raise AssertionError(f"sdist has {len(roots)} roots")
    return roots[0]


def inspect_sdist(source: Path) -> None:
    if any(path.suffix in NATIVE_SUFFIXES for path in source.rglob("*")):
        raise AssertionError("sdist contains a prebuilt native library")
    native_manifest = source / "native-source-files.txt"
    native_files = {
        line.strip() for line in native_manifest.read_text(encoding="utf-8").splitlines()
        if line.strip() and not line.lstrip().startswith("#")
    }
    expected = SDIST_FIXED_FILES | {f"_native_src/{relative}" for relative in native_files}
    actual = {sdist_member_name(path, source) for path in source.rglob("*") if path.is_file()}
    reject_unexpected_members(actual, expected)


def sdist_member_name(path: PurePath, root: PurePath) -> str:
    return path.relative_to(root).as_posix()


def reject_unexpected_members(actual: set[str], expected: set[str]) -> None:
    unexpected = sorted(actual - expected)
    missing = sorted(expected - actual)
    if unexpected:
        raise AssertionError(f"unexpected sdist members: {unexpected}")
    if missing:
        raise AssertionError(f"missing sdist members: {missing}")


def build_sdist_wheel(source: Path, output: Path, env: dict[str, str]) -> Path:
    output.mkdir()
    completed = subprocess.run(
        [sys.executable, "-m", "build", "--no-isolation", "--wheel", "--outdir", str(output), "."],
        cwd=source, env=env, check=True, capture_output=True, text=True,
    )
    print(completed.stdout, end="")
    print(completed.stderr, end="", file=sys.stderr)
    warning = "wheel needs a higher macOS version"
    if warning in completed.stdout or warning in completed.stderr:
        raise AssertionError("wheel target was repaired after build instead of configured before build")
    wheels = list(output.glob("spl_toolkit-*.whl"))
    if len(wheels) != 1:
        raise AssertionError(f"sdist build produced {len(wheels)} wheels")
    return wheels[0]


def check_missing_compiler(source: Path, output: Path, env: dict[str, str]) -> None:
    copy = output.parent / "missing-compiler-source"
    shutil.copytree(source, copy)
    unrelated = copy / "spl_toolkit" / native_library_name()
    unrelated.write_bytes(b"unrelated prebuilt payload")
    output.mkdir()
    no_go_env = env.copy()
    no_go_env["PATH"] = str(Path(sys.executable).parent)
    completed = subprocess.run(
        [sys.executable, "-m", "build", "--no-isolation", "--wheel", "--outdir", str(output), "."],
        cwd=copy, env=no_go_env, text=True, capture_output=True, check=False,
    )
    if completed.returncode == 0:
        raise AssertionError("source wheel succeeded without a Go compiler")
    if list(output.glob("*.whl")):
        raise AssertionError("failed compiler build left an installable artifact")


def check_metadata_without_compiler(source: Path, output: Path, env: dict[str, str]) -> None:
    output.mkdir()
    no_go_env = env.copy()
    no_go_env["PATH"] = str(Path(sys.executable).parent)
    run(
        [sys.executable, "setup.py", "egg_info", "--egg-base", str(output)],
        cwd=source, env=no_go_env,
    )


def git_status(root: Path) -> str | None:
    completed = subprocess.run(
        ["git", "status", "--porcelain=v1", "--untracked-files=all"], cwd=root,
        text=True, capture_output=True, check=False,
    )
    return completed.stdout if completed.returncode == 0 else None


def check_package(sdist: Path, wheel_dir: Path, expected_version: str | None) -> None:
    root = Path(__file__).resolve().parents[1]
    before = git_status(root)
    wheels = list(wheel_dir.glob("spl_toolkit-*.whl"))
    if len(wheels) != 1:
        raise AssertionError(f"expected exactly one spl-toolkit wheel, found {len(wheels)}")
    wheel = wheels[0]
    _name, wheel_version, _build, _tags = parse_wheel_filename(wheel.name)
    version = expected_version or str(wheel_version)
    inspect_wheel(wheel, version)

    with tempfile.TemporaryDirectory(prefix="spl-package-check-") as temporary:
        temp = Path(temporary)
        outside = temp / "outside"
        outside.mkdir()
        requirements = root / "python" / "requirements-dev.txt"
        install_and_check(wheel, temp / "wheel-venv", outside, version, requirements)

        source = unpack_sdist(sdist, temp / "sdist")
        inspect_sdist(source)
        check_metadata_without_compiler(source, temp / "metadata", clean_env())
        source_wheel = build_sdist_wheel(source, temp / "sdist-wheel", clean_env())
        inspect_wheel(source_wheel, version)
        install_and_check(source_wheel, temp / "sdist-venv", outside, version, source / "requirements-dev.txt")
        check_missing_compiler(source, temp / "failed-wheel", clean_env())

    after = git_status(root)
    if before is not None and after != before:
        raise AssertionError("package check changed tracked or untracked repository files")


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--sdist", type=Path, required=True)
    parser.add_argument("--wheel-dir", type=Path, required=True)
    parser.add_argument("--expected-version")
    args = parser.parse_args()
    check_package(args.sdist.resolve(), args.wheel_dir.resolve(), args.expected_version)
    print("package acceptance passed")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
