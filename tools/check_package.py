#!/usr/bin/env python3
"""Build-independent acceptance checks for SPL Toolkit wheel and sdist artifacts."""

from __future__ import annotations

import argparse
from email.parser import Parser
import hashlib
import json
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
    "LICENSE", "MANIFEST.in", "PARSER-LICENSE", "PKG-INFO", "README.md", "VERSION", "build_support.py",
    "native-source-files.txt", "pyproject.toml", "requirements-build.txt",
    "requirements-dev.txt", "setup.cfg", "setup.py", "spl_toolkit/__init__.py",
    "spl_toolkit/exceptions.py", "spl_toolkit/libspl_toolkit.h", "spl_toolkit/mapper.py",
    "spl_toolkit.egg-info/PKG-INFO", "spl_toolkit.egg-info/SOURCES.txt",
    "spl_toolkit.egg-info/dependency_links.txt", "spl_toolkit.egg-info/top_level.txt",
    "tests/test_mapper.py", "tests/test_native_abi.py", "tests/test_native_mapper.py",
    "tests/test_native_analysis.py",
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
NATIVE_TESTS = ("test_native_abi.py", "test_native_mapper.py", "test_native_analysis.py")
ACCEPTANCE_FILES = ("test_documented_cli.py", "test_surfaces.py", "test_analysis_surfaces.py", "cli_examples.json")
REQUIRED_PYTEST_PLUGIN = r'''\
import json
import os

collected = 0
passed = set()
failed = set()
skipped = set()

def pytest_collection_finish(session):
    global collected
    collected = len(session.items)

def pytest_runtest_logreport(report):
    if report.skipped:
        skipped.add(report.nodeid)
    elif report.failed:
        failed.add(report.nodeid)
    elif report.when == "call" and report.passed:
        passed.add(report.nodeid)

def pytest_sessionfinish(session, exitstatus):
    counts = {
        "collected": collected,
        "passed": len(passed),
        "failed": len(failed),
        "skipped": len(skipped),
    }
    destination = os.environ.get("SPL_TEST_COUNTS")
    if destination:
        temporary = destination + ".tmp"
        with open(temporary, "w", encoding="utf-8") as output:
            json.dump(counts, output, sort_keys=True)
            output.write("\n")
        os.replace(temporary, destination)
    if collected == 0 or not passed or failed or skipped or len(passed) != collected:
        session.exitstatus = 1
'''


def sha256(path: Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as source:
        for chunk in iter(lambda: source.read(1024 * 1024), b""):
            digest.update(chunk)
    return digest.hexdigest()


def write_evidence(destination: Path, record: dict[str, object]) -> None:
    destination.parent.mkdir(parents=True, exist_ok=True)
    temporary = destination.with_suffix(destination.suffix + ".tmp")
    if destination.exists():
        raise FileExistsError(f"evidence destination already exists: {destination}")
    if temporary.exists():
        raise FileExistsError(f"evidence staging file already exists: {temporary}")
    with temporary.open("x", encoding="utf-8") as output:
        output.write(json.dumps(record, indent=2, sort_keys=True) + "\n")
    try:
        os.link(temporary, destination)
    finally:
        temporary.unlink(missing_ok=True)


def restore_verified_executables(directory: Path, hashes: dict[str, str], names: tuple[str, ...]) -> None:
    if os.name == "nt":
        return
    for name in names:
        if name not in hashes:
            raise AssertionError(f"executable {name} is absent from accepted hashes")
        path = directory / name
        if not path.is_file() or sha256(path) != hashes[name]:
            raise AssertionError(f"executable {name} does not match accepted hash")
        path.chmod(path.stat().st_mode | 0o755)


def write_required_pytest_plugin(directory: Path) -> None:
    (directory / "conftest.py").write_text(REQUIRED_PYTEST_PLUGIN, encoding="utf-8")


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


def _copy_required_files(source: Path, destination: Path, names: tuple[str, ...]) -> None:
    destination.mkdir()
    for name in names:
        path = source / name
        if not path.is_file():
            raise FileNotFoundError(f"required test input is missing: {path}")
        shutil.copy2(path, destination / name)


def _run_required_suite(python: Path, suite: Path, result: Path, outside: Path, env: dict[str, str]) -> dict[str, int]:
    write_required_pytest_plugin(suite)
    child_env = env | {"SPL_TEST_COUNTS": str(result)}
    run([str(python), "-I", "-m", "pytest", str(suite), "-q"], cwd=outside, env=child_env)
    counts = json.loads(result.read_text(encoding="utf-8"))
    if set(counts) != {"collected", "passed", "failed", "skipped"}:
        raise AssertionError(f"required suite returned invalid counts: {counts}")
    return counts


def install_and_check(
    wheel: Path,
    directory: Path,
    outside_checkout: Path,
    expected_version: str,
    requirements: Path,
    cli: Path,
    server: Path,
    fixture_source: Path,
    docs_root: Path,
) -> dict[str, object]:
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
    metadata_script = check_script + "\nimport json, platform\nwith SPLMapper() as _metadata_mapper:\n    _native_version = _metadata_mapper.native_version\nprint(json.dumps({'installed_module': str(pathlib.Path(spl_toolkit.__file__).resolve()), 'venv_prefix': str(pathlib.Path(sys.prefix).resolve()), 'python_version': platform.python_version(), 'python_runtime': sys.version, 'package_version': spl_toolkit.__version__, 'native_version': _native_version}, sort_keys=True))\n"
    completed = subprocess.run(
        [str(python), "-I", "-c", textwrap.dedent(metadata_script)], cwd=outside_checkout,
        env=install_env, check=True, text=True, capture_output=True,
    )
    metadata = json.loads(completed.stdout.splitlines()[-1])

    installed_test_dir = outside_checkout / f"tests-{directory.name}"
    _copy_required_files(docs_root / "python" / "tests", installed_test_dir, NATIVE_TESTS)
    analysis_fixture = outside_checkout / f"analysis-cases-{directory.name}.json"
    shutil.copy2(docs_root / "testdata" / "analysis" / "cases.json", analysis_fixture)
    analysis_env = {"SPL_ANALYSIS_FIXTURES": str(analysis_fixture.resolve())}
    native_counts = _run_required_suite(
        python, installed_test_dir, outside_checkout / f"native-counts-{directory.name}.json", outside_checkout, install_env | analysis_env
    )

    acceptance_dir = outside_checkout / f"acceptance-{directory.name}"
    _copy_required_files(docs_root / "tests" / "acceptance", acceptance_dir, ACCEPTANCE_FILES)
    fixture = outside_checkout / f"cases-{directory.name}.json"
    shutil.copy2(fixture_source, fixture)
    acceptance_env = install_env | analysis_env | {
        "SPL_CLI": str(cli.resolve()),
        "SPL_SERVER": str(server.resolve()),
        "SPL_FIXTURES": str(fixture.resolve()),
        "SPL_DOCS_ROOT": str(docs_root.resolve()),
    }
    surface_counts = _run_required_suite(
        python, acceptance_dir, outside_checkout / f"surface-counts-{directory.name}.json",
        outside_checkout, acceptance_env,
    )
    return metadata | {
        "wheel_sha256": sha256(wheel),
        "fixture_hashes": {"baseline": sha256(fixture), "analysis": sha256(analysis_fixture)},
        "tests": {"required_native": native_counts, "surface_acceptance": surface_counts},
        "cli_examples": "passed",
        "surface_parity": "passed",
        "version_agreement": "passed",
    }


def build_surface_binaries(root: Path, output: Path, version: str) -> tuple[Path, Path]:
    output.mkdir()
    suffix = ".exe" if os.name == "nt" else ""
    cli = output / f"spl-toolkit{suffix}"
    server = output / f"spl-toolkit-server{suffix}"
    ldflags = f"-X=github.com/delgado-jacob/spl-toolkit/internal/buildinfo.Version={version}"
    env = clean_env() | {"GOTOOLCHAIN": "local"}
    run(
        ["go", "build", "-mod=readonly", "-trimpath", "-ldflags", ldflags, "-o", str(cli), "./cmd"],
        cwd=root,
        env=env,
    )
    run(
        ["go", "build", "-mod=readonly", "-trimpath", "-ldflags", ldflags, "-o", str(server), "./cmd/server"],
        cwd=root,
        env=env,
    )
    return cli, server


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
        parser_licenses = [name for name in names if name == "spl_toolkit/PARSER-LICENSE"]
        if len(parser_licenses) != 1 or b"Copyright (c) 2024 Clemens Sageder" not in archive.read(parser_licenses[0]):
            raise AssertionError("wheel parser attribution is missing")
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


def _check_package(
    sdist: Path | None, wheel_dir: Path, expected_version: str | None, root: Path,
    *, wheel_only: bool = False, wheel_path: Path | None = None, cli_path: Path | None = None,
    server_path: Path | None = None,
) -> dict[str, object]:
    root = root.resolve()
    before = git_status(root)
    wheels = [wheel_path] if wheel_path else list(wheel_dir.glob("spl_toolkit-*.whl"))
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
        if cli_path is not None and server_path is not None:
            cli, server = cli_path.resolve(), server_path.resolve()
        elif wheel_only:
            raise AssertionError("wheel-only checks require accepted --cli and --server payloads")
        else:
            cli, server = build_surface_binaries(root, temp / "surface-binaries", version)
        fixture = root / "testdata" / "baseline" / "cases.json"
        evidence = install_and_check(
            wheel, temp / "wheel-venv", outside, version, requirements,
            cli, server, fixture, root,
        )

        if wheel_only:
            source = None
        elif sdist is None:
            raise AssertionError("source distribution is required unless --wheel-only is used")
        else:
            source = unpack_sdist(sdist, temp / "sdist")
        if source is not None:
            inspect_sdist(source)
            check_metadata_without_compiler(source, temp / "metadata", clean_env())
            source_wheel = build_sdist_wheel(source, temp / "sdist-wheel", clean_env())
            inspect_wheel(source_wheel, version)
            evidence["rebuilt_sdist"] = install_and_check(
                source_wheel, temp / "sdist-venv", outside, version,
                source / "requirements-dev.txt", cli, server, fixture, root,
            )
            check_missing_compiler(source, temp / "failed-wheel", clean_env())

    after = git_status(root)
    if before is not None and after != before:
        raise AssertionError("package check changed tracked or untracked repository files")
    return evidence


def check_package(sdist: Path, wheel_dir: Path, expected_version: str | None) -> None:
    _check_package(sdist, wheel_dir, expected_version, Path(__file__).resolve().parents[1])


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--sdist", type=Path)
    parser.add_argument("--wheel-dir", type=Path)
    parser.add_argument("--wheel", type=Path)
    parser.add_argument("--wheel-only", action="store_true")
    parser.add_argument("--cli", type=Path)
    parser.add_argument("--server", type=Path)
    parser.add_argument("--accepted-result", type=Path)
    parser.add_argument("--accepted-dir", type=Path)
    parser.add_argument("--source-sha")
    parser.add_argument("--target")
    parser.add_argument("--evidence", type=Path)
    parser.add_argument("--expected-version")
    parser.add_argument("--source-root", type=Path)
    args = parser.parse_args()
    root = args.source_root.resolve() if args.source_root else Path(__file__).resolve().parents[1]
    if args.accepted_dir:
        accepted_dir = args.accepted_dir.resolve()
        args.accepted_result = accepted_dir / "result.json"
        accepted = json.loads(args.accepted_result.read_text(encoding="utf-8"))
        hashes = accepted.get("artifact_hashes", {})
        wheel_names = [name for name in hashes if name.endswith(".whl")]
        if len(wheel_names) != 1:
            parser.error("accepted result must identify exactly one wheel")
        payloads = accepted.get("accepted_payloads", {})
        args.wheel = accepted_dir / wheel_names[0]
        args.cli = accepted_dir / str(payloads.get("cli", ""))
        args.server = accepted_dir / str(payloads.get("server", ""))
    if args.source_sha:
        args.source_sha = subprocess.run(
            ["git", "rev-parse", "--verify", f"{args.source_sha}^{{commit}}"], cwd=root,
            check=True, text=True, capture_output=True,
        ).stdout.strip()
    if not args.wheel_dir and not args.wheel:
        parser.error("one of --wheel-dir or --wheel is required")
    if not args.wheel_only and not args.sdist:
        parser.error("--sdist is required unless --wheel-only is used")
    if args.wheel_only and not all((args.wheel, args.cli, args.server, args.accepted_result, args.source_sha, args.target, args.evidence)):
        parser.error("--wheel-only requires --wheel, --cli, --server, --accepted-result, --source-sha, --target, and --evidence")
    wheel_dir = args.wheel_dir.resolve() if args.wheel_dir else args.wheel.resolve().parent
    if args.wheel_only:
        accepted = json.loads(args.accepted_result.read_text(encoding="utf-8"))
        if accepted.get("status") != "passed" or accepted.get("source_sha") != args.source_sha or accepted.get("target") != args.target:
            raise AssertionError("accepted result identity/status does not match this installed-wheel job")
        hashes = accepted.get("artifact_hashes", {})
        names = accepted.get("accepted_payloads", {})
        if hashes.get(args.wheel.name) != sha256(args.wheel):
            raise AssertionError("wheel does not match accepted release hash")
        if names.get("cli") != args.cli.name or names.get("server") != args.server.name:
            raise AssertionError("CLI/server do not match accepted payload identities")
        restore_verified_executables(args.cli.parent, hashes, (args.cli.name, args.server.name))
    evidence = _check_package(
        args.sdist.resolve() if args.sdist else None, wheel_dir, args.expected_version, root,
        wheel_only=args.wheel_only, wheel_path=args.wheel.resolve() if args.wheel else None,
        cli_path=args.cli, server_path=args.server,
    )
    if args.wheel_only:
        record = {
            "schema_version": 1, "kind": "installed-wheel", "source_sha": args.source_sha,
            "status": "passed", "target": args.target,
            "architecture": "arm64" if args.target.endswith("arm64") else "x86_64",
        } | evidence
        write_evidence(args.evidence, record)
    elif args.evidence:
        write_evidence(args.evidence, evidence)
    print("package acceptance passed")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
