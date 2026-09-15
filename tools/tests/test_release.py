"""Tests for controlled, reproducible release artifacts."""

from __future__ import annotations

import gzip
import hashlib
import importlib.util
import io
import json
from pathlib import Path
import tarfile
from types import SimpleNamespace
import zipfile

import pytest

import tools.release as release
import tools.check_reproducible as reproducible
import tools.check_package as package_check
from tools.check_reproducible import compare_artifacts
from tools.release import artifact_hashes, normalize_archive, parser_attribution, require_python_archives, verify_wheel_native


EPOCH = 1788652800
ROOT = Path(__file__).resolve().parents[2]


def test_requirement_contract_and_sources_are_release_inputs():
    expected = {
        "contracts/v1/requirements.schema.json",
        "pkg/analysis/requirements.go",
        "pkg/analysis/requirements_trace.go",
        "pkg/analysis/resource_limit.go",
    }
    native = set((ROOT / "python/native-source-files.txt").read_text(encoding="utf-8").splitlines())
    content = set((ROOT / "tools/release-content-files.txt").read_text(encoding="utf-8").splitlines())
    sources = set((ROOT / "tools/release-source-files.txt").read_text(encoding="utf-8").splitlines())

    assert expected <= native
    assert expected <= sources
    assert "contracts/v1/requirements.schema.json" in content
    assert not expected.intersection({line for line in content if line.startswith("pkg/")})
    for entries in (native, content, sources):
        assert not any(line.startswith(("tests/", "_build_plan/")) for line in entries)


def test_requirement_contract_reaches_wheel_through_native_source_manifest(tmp_path: Path, monkeypatch):
    spec = importlib.util.spec_from_file_location(
        "spl_toolkit_release_build_support", ROOT / "python/build_support.py"
    )
    assert spec and spec.loader
    support = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(support)

    distribution = support.NativeDistribution({"script_name": str(ROOT / "python/setup.py")})
    command = support.BuildPy(distribution)
    command.initialize_options()
    command.build_lib = str(tmp_path / "build")

    monkeypatch.setattr(support.build_py, "run", lambda _command: None)

    def stage_native_stub(_source: Path, output: Path, _version: str) -> None:
        output.parent.mkdir(parents=True, exist_ok=True)
        output.write_bytes(b"native")

    monkeypatch.setattr(support, "build_native", stage_native_stub)
    monkeypatch.setattr(support, "parser_attribution", lambda _setup_dir: "parser attribution\n")
    command.run()

    relative = Path("contracts/v1/requirements.schema.json")
    staged = Path(command.build_lib) / "spl_toolkit" / relative
    assert staged.read_bytes() == (ROOT / relative).read_bytes()


def test_staged_sdist_native_source_closure_compiles_through_build_py(
    tmp_path: Path, monkeypatch
):
    spec = importlib.util.spec_from_file_location(
        "spl_toolkit_release_build_support", ROOT / "python/build_support.py"
    )
    assert spec and spec.loader
    support = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(support)

    checkout_distribution = support.NativeDistribution(
        {"script_name": str(ROOT / "python/setup.py")}
    )
    source_command = support.SourceDistribution(checkout_distribution)
    source_command.ensure_finalized()
    staged = tmp_path / "spl_toolkit-0.1.1"
    monkeypatch.chdir(ROOT / "python")
    source_command.make_release_tree(
        str(staged), ["build_support.py", "native-source-files.txt", "setup.py"]
    )

    staged_distribution = support.NativeDistribution(
        {"script_name": str(staged / "setup.py")}
    )
    build_command = support.BuildPy(staged_distribution)
    build_command.initialize_options()
    build_command.build_lib = str(tmp_path / "build")
    monkeypatch.setattr(support.build_py, "run", lambda _command: None)
    monkeypatch.setenv("GOWORK", "off")
    monkeypatch.setenv("GOPROXY", "off")
    monkeypatch.setenv("GOSUMDB", "off")
    monkeypatch.setenv("GOCACHE", str(tmp_path / "go-cache"))
    monkeypatch.setenv("GOTMPDIR", str(tmp_path / "go-tmp"))

    build_command.run()

    native = Path(build_command.build_lib) / "spl_toolkit" / support.native_library_name()
    assert native.is_file()
    assert native.with_suffix(".h").is_file()


def test_requirement_release_manifests_package_exact_inputs(tmp_path: Path):
    release.package_tooling_content(ROOT, tmp_path, "0.1.1", EPOCH)

    contract = Path("contracts/v1/requirements.schema.json")
    assert (tmp_path / contract).read_bytes() == (ROOT / contract).read_bytes()
    with tarfile.open(tmp_path / "spl-toolkit-source-0.1.1.tar.gz", "r:gz") as archive:
        names = set(archive.getnames())
    assert {
        contract.as_posix(),
        "pkg/analysis/requirements.go",
        "pkg/analysis/requirements_trace.go",
        "pkg/analysis/resource_limit.go",
    } <= names


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


@pytest.mark.parametrize(
    "name",
    ["../escape", "/absolute", "pkg/../../escape", r"C:\escape", "C:/escape", r"\\server\share\escape"],
)
def test_archive_normalization_rejects_unsafe_member_names(tmp_path: Path, name: str):
    archive_path = tmp_path / "unsafe.whl"
    with zipfile.ZipFile(archive_path, "w") as archive:
        archive.writestr(name, b"payload")

    with pytest.raises(ValueError, match="unsafe archive path"):
        normalize_archive(archive_path, EPOCH)


@pytest.mark.parametrize("name", [r"C:\escape", "C:/escape", r"\\server\share\escape"])
def test_sdist_normalization_rejects_windows_absolute_member_names(tmp_path: Path, name: str):
    archive_path = tmp_path / "unsafe.tar.gz"
    with tarfile.open(archive_path, "w:gz") as archive:
        info = tarfile.TarInfo(name)
        info.size = 1
        archive.addfile(info, io.BytesIO(b"x"))

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


def test_parser_attribution_extracts_the_complete_license_notice():
    notice = parser_attribution(ROOT / "grammar" / "SPLParser.g4")

    assert "Copyright (c) 2024 Clemens Sageder" in notice
    assert "Redistribution and use in source and binary forms" in notice
    assert "THIS SOFTWARE IS PROVIDED BY THE AUTHOR" in notice
    assert "parser grammar" not in notice


def test_windows_gcc_resolution_uses_powershell_get_command(monkeypatch):
    seen = []
    monkeypatch.setattr(release.shutil, "which", lambda name: "C:/PowerShell/pwsh.exe" if name == "powershell" else None)

    def fake_version(command):
        seen.append(command)
        return "C:/mingw64/bin/gcc.exe"

    monkeypatch.setattr(release, "_version_output", fake_version)

    assert release._windows_gcc_from_powershell() == "C:/mingw64/bin/gcc.exe"
    assert seen == [[
        "C:/PowerShell/pwsh.exe", "-NoProfile", "-Command", "(Get-Command gcc -ErrorAction Stop).Source",
    ]]


def test_production_runner_identity_rejects_mismatching_observed_image(monkeypatch):
    monkeypatch.delenv("ImageOS", raising=False)
    monkeypatch.delenv("ImageVersion", raising=False)

    identity = release._runner_identity({"runner": "ubuntu-24.04"})
    assert identity["evidence_kind"] == "pinned-local"
    assert identity["observed_image_os"] == "unavailable"

    monkeypatch.setenv("ImageOS", "ubuntu22")
    monkeypatch.setenv("ImageVersion", "20260901.1")
    with pytest.raises(RuntimeError, match="runner image mismatch"):
        release._runner_identity({"runner": "ubuntu-24.04"})


def test_diagnostic_runner_identity_is_structurally_non_accepting(monkeypatch):
    monkeypatch.delenv("ImageOS", raising=False)
    monkeypatch.delenv("ImageVersion", raising=False)

    identity = release._runner_identity({"runner": "diagnostic-macos-26.2"})

    assert identity["evidence_kind"] == "diagnostic"
    assert identity["pinned_environment"] is False
    assert identity["configured_runner"] == "diagnostic-macos-26.2"
    assert identity["observed_image_os"] == "unavailable"
    assert reproducible._result_status(identity) == "diagnostic-passed"


def test_release_environment_rejects_wrong_effective_go_target(monkeypatch):
    config = {
        "go": "1.26.8",
        "build_python": release.platform.python_version(),
        "targets": {
            "linux-amd64": {
                "runner": "diagnostic-linux-amd64",
                "goarch": "amd64",
                "cc": "gcc-13",
                "wheel_platform": "linux_x86_64",
            }
        },
    }
    monkeypatch.setattr(release.platform, "system", lambda: "Linux")
    monkeypatch.setattr(release.platform, "machine", lambda: "x86_64")
    monkeypatch.setattr(
        release,
        "_version_output",
        lambda command: {
            ("go", "version"): "go version go1.26.8 linux/amd64",
            ("go", "env", "GOOS"): "linux",
            ("go", "env", "GOARCH"): "arm64",
        }[tuple(command)],
    )

    with pytest.raises(RuntimeError, match="effective GOARCH mismatch"):
        release.validate_environment(config, "linux-amd64")


def test_release_environment_rejects_non_gcc13_linux_compiler(monkeypatch):
    config = {
        "go": "1.26.8",
        "build_python": release.platform.python_version(),
        "targets": {
            "linux-amd64": {
                "runner": "diagnostic-linux-amd64",
                "goarch": "amd64",
                "cc": "gcc-13",
                "wheel_platform": "linux_x86_64",
            }
        },
    }
    monkeypatch.setattr(release.platform, "system", lambda: "Linux")
    monkeypatch.setattr(release.platform, "machine", lambda: "x86_64")
    monkeypatch.setattr(release.shutil, "which", lambda name: "/usr/bin/gcc-13")
    monkeypatch.setattr(
        release,
        "_version_output",
        lambda command: {
            ("go", "version"): "go version go1.26.8 linux/amd64",
            ("go", "env", "GOOS"): "linux",
            ("go", "env", "GOARCH"): "amd64",
            ("/usr/bin/gcc-13", "-dumpmachine"): "x86_64-linux-gnu",
            ("/usr/bin/gcc-13", "-dumpfullversion"): "12.2.0",
        }[tuple(command)],
    )
    monkeypatch.setattr(
        release.subprocess,
        "run",
        lambda *args, **kwargs: SimpleNamespace(stdout="gcc (Debian) 12.2.0\n"),
    )

    with pytest.raises(RuntimeError, match="Linux compiler mismatch"):
        release.validate_environment(config, "linux-amd64")


def test_darwin_build_environment_uses_the_validated_sdk(monkeypatch):
    monkeypatch.setenv("PATH", "/usr/bin:/bin")
    config = {
        "targets": {
            "darwin-arm64": {
                "goarch": "arm64",
                "deployment_target": "15.0",
                "wheel_platform": "macosx_15_0_arm64",
            }
        }
    }
    environment = {
        "cc": "/Applications/Xcode.app/Contents/Developer/Toolchains/XcodeDefault.xctoolchain/usr/bin/clang",
        "sdk": "/Applications/Xcode.app/Contents/Developer/Platforms/MacOSX.platform/Developer/SDKs/MacOSX.sdk",
    }

    child = release._release_build_environment(config, "darwin-arm64", environment, 1788652800)

    assert child["SDKROOT"] == environment["sdk"]
    assert child["MACOSX_DEPLOYMENT_TARGET"] == "15.0"
    assert child["CC"] == environment["cc"]


def test_docker_context_is_an_exact_allowlist():
    entries = (ROOT / ".dockerignore").read_text(encoding="utf-8").splitlines()

    assert entries[0] == "**"
    allowed = [entry for entry in entries if entry.startswith("!")]
    assert allowed
    assert all("*" not in entry for entry in allowed)
    assert "!cmd/cli.go" in allowed
    assert "!python/pyproject.toml" in allowed
    assert "!docs/docs.go" in allowed
    assert "!cmd/local.env" not in allowed


def test_reproducibility_controller_delegates_to_exported_checker(tmp_path: Path, monkeypatch):
    output = tmp_path / "result"
    source_sha = "a" * 40
    calls = []

    monkeypatch.setattr(
        reproducible,
        "_git",
        lambda root, *args: source_sha if args[0] == "rev-parse" else "1788652800",
    )

    def fake_export(repository, ref, archive, destination):
        script = destination / "tools" / "check_reproducible.py"
        script.parent.mkdir(parents=True)
        script.write_text("# exported checker\n", encoding="utf-8")

    def fake_run(command, **kwargs):
        calls.append((command, kwargs))
        (output / "result.json").write_text('{"status":"passed"}\n', encoding="utf-8")
        return reproducible.subprocess.CompletedProcess(command, 0, "worker stdout\n", "")

    monkeypatch.setattr(reproducible, "export_ref", fake_export)
    monkeypatch.setattr(reproducible.subprocess, "run", fake_run)

    assert reproducible.check_reproducible("candidate", output) == {"status": "passed"}
    command, options = calls[0]
    assert command[1] == str(output / "checker-source" / "tools" / "check_reproducible.py")
    assert command[command.index("--worker-source-sha") + 1] == source_sha
    assert options["cwd"] == output / "checker-source"


def test_package_acceptance_executes_exported_checker(tmp_path: Path, monkeypatch):
    source = tmp_path / "source"
    (source / "tools").mkdir(parents=True)
    checker = source / "tools" / "check_package.py"
    checker.write_text("# exported package checker\n", encoding="utf-8")
    sdist = tmp_path / "artifact.tar.gz"
    wheel_dir = tmp_path / "wheel"
    log = tmp_path / "package.log"
    seen = {}

    def fake_run(command, **kwargs):
        seen.update(command=command, kwargs=kwargs)
        return reproducible.subprocess.CompletedProcess(command, 0, "accepted\n", "")

    monkeypatch.setattr(reproducible.subprocess, "run", fake_run)
    reproducible._run_package_check(source, sdist, wheel_dir, "0.1.1", log)

    assert seen["command"][1] == str(checker)
    assert seen["command"][-2:] == ["--source-root", str(source)]
    assert seen["kwargs"]["cwd"] == source


def test_verified_payloads_are_promoted_with_result_and_environment(tmp_path: Path):
    output = tmp_path / "result"
    build = output / "build-1"
    evidence = output / "build-1-evidence"
    build.mkdir(parents=True)
    evidence.mkdir()
    (build / "cli").write_bytes(b"verified")
    (build / "unix-server").write_bytes(b"server")
    (build / "mode-case.txt").write_bytes(b"plain")
    (build / "artifact.whl").write_bytes(b"wheel")
    (build / "unix-server").chmod(0o644)
    (build / "mode-case.txt").chmod(0o644)
    hashes = {name: reproducible._sha256(build / name) for name in ("cli", "unix-server", "mode-case.txt", "artifact.whl")}
    environment = {"pinned_environment": True, "artifacts": hashes}
    (evidence / "environment.json").write_text(json.dumps(environment), encoding="utf-8")
    result = {
        "source_sha": "a" * 40, "target": "linux-amd64", "status": "passed",
        "failed_checks": [], "artifact_hashes": hashes,
        "environment": environment,
        "accepted_payloads": {"cli": "cli", "server": "unix-server", "native": "mode-case.txt"},
    }

    reproducible._promote_accepted(output, result)

    accepted = output / "accepted"
    assert (accepted / "result.json").is_file()
    assert (accepted / "environment.json").is_file()
    assert (accepted / "unix-server").stat().st_mode & 0o111
    assert not ((accepted / "mode-case.txt").stat().st_mode & 0o111)


def test_failed_reproducibility_result_is_never_promoted(tmp_path: Path):
    output = tmp_path / "result"
    (output / "build-1").mkdir(parents=True)
    with pytest.raises(RuntimeError, match="cannot promote"):
        reproducible._promote_accepted(output, {"status": "failed", "failed_checks": ["bad"]})
    assert not (output / "accepted").exists()
