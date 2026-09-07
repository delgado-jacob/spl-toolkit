"""Focused tests for the Python package build helpers."""

from __future__ import annotations

import importlib.util
from pathlib import Path, PureWindowsPath
import platform
import subprocess
import sys
import json

import pytest


ROOT = Path(__file__).resolve().parents[2]
PYTHON_DIR = ROOT / "python"


def test_installed_wheel_job_bootstraps_package_checker_dependencies():
    workflow = (ROOT / ".github" / "workflows" / "ci.yml").read_text(encoding="utf-8")
    installed_job = workflow.split("  installed-wheel:\n", 1)[1].split("\n  go-floor:\n", 1)[0]
    bootstrap = (
        "python -m pip install --disable-pip-version-check "
        "-r python/requirements-build.txt"
    )
    checker = "python tools/check_package.py --wheel-only"

    assert bootstrap in installed_job
    assert installed_job.index(bootstrap) < installed_job.index(checker)
    assert "packaging==25.0" in (PYTHON_DIR / "requirements-build.txt").read_text(
        encoding="utf-8"
    ).splitlines()


def load_build_support():
    spec = importlib.util.spec_from_file_location(
        "spl_toolkit_build_support", PYTHON_DIR / "build_support.py"
    )
    assert spec and spec.loader
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


def load_package_checker():
    spec = importlib.util.spec_from_file_location("check_package", ROOT / "tools" / "check_package.py")
    assert spec and spec.loader
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


def test_checkout_source_and_version_are_resolved_from_repository():
    support = load_build_support()

    assert support.source_root(PYTHON_DIR) == ROOT
    assert support.read_version(PYTHON_DIR) == "0.1.1"


def test_staged_source_and_version_are_resolved_without_checkout(tmp_path: Path):
    support = load_build_support()
    staged = tmp_path / "spl_toolkit-0.1.1"
    native = staged / "_native_src"
    native.mkdir(parents=True)
    (native / "go.mod").write_text("module example.invalid/staged\n", encoding="utf-8")
    (staged / "VERSION").write_text("0.1.1\n", encoding="utf-8")
    (staged.parent / "VERSION").write_text("9.9.9\n", encoding="utf-8")

    assert support.source_root(staged) == native
    assert support.read_version(staged) == "0.1.1"


def test_source_root_failure_names_missing_native_source(tmp_path: Path):
    support = load_build_support()

    with pytest.raises(FileNotFoundError, match="_native_src/go.mod"):
        support.source_root(tmp_path)


def test_build_native_uses_release_flags_and_requires_output(tmp_path: Path, monkeypatch):
    support = load_build_support()
    source = tmp_path / "source"
    source.mkdir()
    output = tmp_path / "build" / support.native_library_name()
    gotmp = tmp_path / "go temp with spaces"
    gotmp.mkdir()
    monkeypatch.setenv("GOTMPDIR", str(gotmp))
    seen = {}

    def fake_run(command, *, cwd, env, check):
        seen.update(command=command, cwd=cwd, env=env, check=check)

    monkeypatch.setattr(subprocess, "run", fake_run)

    with pytest.raises(RuntimeError, match="produced no shared library"):
        support.build_native(source, output, "0.1.1")

    assert seen["command"] == [
        "go", "build", "-mod=readonly", "-trimpath", "-buildvcs=false",
        "-buildmode=c-shared", "-ldflags",
        "-buildid= -X=github.com/delgado-jacob/spl-toolkit/internal/buildinfo.Version=0.1.1 "
        f"-extldflags={support.native_linker_flag()}",
        "-o", str(output.resolve()), "./pkg/bindings",
    ]
    assert seen["cwd"] == source
    assert seen["env"]["CGO_ENABLED"] == "1"
    assert seen["env"]["GOTOOLCHAIN"] == "local"
    assert f"-ffile-prefix-map={source.resolve()}=." in seen["env"]["CGO_CFLAGS"]
    assert f"-fdebug-prefix-map={source.resolve()}=." in seen["env"]["CGO_CFLAGS"]
    assert f"-ffile-prefix-map={gotmp.resolve()}=." in seen["env"]["CGO_CFLAGS"]
    assert seen["env"]["CGO_CFLAGS"] == seen["env"]["CGO_CPPFLAGS"]
    if sys.platform == "darwin":
        assert seen["env"]["MACOSX_DEPLOYMENT_TARGET"] == "15.0"
        assert "-mmacosx-version-min=15.0" in seen["env"]["CGO_CFLAGS"]
        assert "-mmacosx-version-min=15.0" in seen["env"]["CGO_LDFLAGS"]
    assert seen["check"] is True


def test_build_native_requires_generated_header(tmp_path: Path, monkeypatch):
    support = load_build_support()
    source = tmp_path / "source"
    source.mkdir()
    gotmp = tmp_path / "gotmp"
    gotmp.mkdir()
    output = tmp_path / "build" / support.native_library_name()
    monkeypatch.setenv("GOTMPDIR", str(gotmp))

    def fake_run(*_args, **_kwargs):
        output.parent.mkdir(parents=True, exist_ok=True)
        output.write_bytes(b"native")

    monkeypatch.setattr(subprocess, "run", fake_run)

    with pytest.raises(RuntimeError, match="produced no C header"):
        support.build_native(source, output, "0.1.1")


def test_platform_linker_flags_are_explicit_and_shell_free():
    support = load_build_support()

    assert support.native_linker_flag("Linux") == "-Wl,--build-id=none"
    assert support.native_linker_flag("Darwin") == "-Wl,-reproducible"
    assert support.native_linker_flag("Windows") == "-Wl,--no-insert-timestamp"


def test_build_native_propagates_missing_compiler_without_artifact(tmp_path: Path, monkeypatch):
    support = load_build_support()
    output = tmp_path / "build" / support.native_library_name()

    def missing_go(*_args, **_kwargs):
        raise FileNotFoundError("go")

    monkeypatch.setattr(subprocess, "run", missing_go)

    with pytest.raises(FileNotFoundError, match="go"):
        support.build_native(tmp_path, output, "0.1.1")
    assert not output.exists()


def test_release_tree_contains_only_allowed_native_source(tmp_path: Path):
    support = load_build_support()
    distribution = support.NativeDistribution({"script_name": str(PYTHON_DIR / "setup.py")})
    command = support.SourceDistribution(distribution)
    command.ensure_finalized()
    release = tmp_path / "release"

    command.make_release_tree(str(release), [])

    assert (release / "LICENSE").is_file()
    assert "Copyright (c) 2024 Clemens Sageder" in (release / "PARSER-LICENSE").read_text(encoding="utf-8")
    assert (release / "README.md").is_file()
    assert (release / "VERSION").read_text(encoding="utf-8").strip() == "0.1.1"
    native = release / "_native_src"
    for relative in (
        "go.mod", "go.sum", "pkg/mapper", "pkg/bindings", "parser", "internal/buildinfo"
    ):
        assert (native / relative).exists(), relative
    assert not (release / ".git").exists()
    assert not (release / "_build_plan").exists()
    assert not list(release.rglob("*.so"))
    assert not list(release.rglob("*.dylib"))
    assert not list(release.rglob("*.dll"))


def test_parser_attribution_is_available_from_checkout_and_staged_sdist(tmp_path: Path):
    support = load_build_support()
    notice = support.parser_attribution(PYTHON_DIR)
    assert "Redistribution and use in source and binary forms" in notice

    staged = tmp_path / "sdist"
    staged.mkdir()
    (staged / "PARSER-LICENSE").write_text(notice, encoding="utf-8")
    assert support.parser_attribution(staged) == notice


def test_native_source_manifest_excludes_unlisted_files(tmp_path: Path):
    support = load_build_support()
    source = tmp_path / "source"
    destination = tmp_path / "staged"
    manifest = source / "native-source-files.txt"
    allowed = ["go.mod", "pkg/mapper/mapper.go", "parser/spl_parser.go"]
    for relative in allowed:
        path = source / relative
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text(relative, encoding="utf-8")
    for relative in ("pkg/mapper/private-notes.txt", "parser/.editor.swp", "pkg/mapper/.coverage"):
        path = source / relative
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text("private", encoding="utf-8")
    manifest.write_text("\n".join(allowed) + "\n", encoding="utf-8")

    support.stage_native_source(source, destination, manifest)

    assert sorted(str(path.relative_to(destination)) for path in destination.rglob("*") if path.is_file()) == sorted(allowed)


def test_installed_checks_copy_tests_from_explicit_source_root(tmp_path: Path, monkeypatch):
    checker = load_package_checker()
    source_root = tmp_path / "exported-source"
    (source_root / "python" / "tests").mkdir(parents=True)
    (source_root / "tests" / "acceptance").mkdir(parents=True)
    for name in checker.NATIVE_TESTS:
        (source_root / "python" / "tests" / name).write_text(name, encoding="utf-8")
    for name in checker.ACCEPTANCE_FILES:
        (source_root / "tests" / "acceptance" / name).write_text(name, encoding="utf-8")

    native = tmp_path / "native"
    acceptance = tmp_path / "acceptance"
    checker._copy_required_files(source_root / "python" / "tests", native, checker.NATIVE_TESTS)
    checker._copy_required_files(source_root / "tests" / "acceptance", acceptance, checker.ACCEPTANCE_FILES)

    assert sorted(path.name for path in native.iterdir()) == sorted(checker.NATIVE_TESTS)
    assert sorted(path.name for path in acceptance.iterdir()) == sorted(checker.ACCEPTANCE_FILES)


def test_binary_wheel_forces_macos_15_tag_from_older_interpreter_target():
    if platform.system() != "Darwin":
        pytest.skip("macOS tag behavior")
    support = load_build_support()
    distribution = support.NativeDistribution({"script_name": str(PYTHON_DIR / "setup.py")})
    command = support.BinaryWheel(distribution)
    command.initialize_options()
    command.plat_name = "macosx_11_0_arm64"

    command.finalize_options()

    assert command.plat_name == "macosx_15_0_arm64"
    assert command.get_tag() == ("py3", "none", "macosx_15_0_arm64")


def test_external_venv_does_not_expose_controller_site_packages(tmp_path: Path):
    checker = load_package_checker()
    python = checker.create_test_environment(tmp_path / "venv")
    controller_site = str(Path(checker.sysconfig.get_paths()["purelib"]).resolve())

    completed = subprocess.run(
        [str(python), "-I", "-c", "import pathlib,sys; print('\\n'.join(str(pathlib.Path(p).resolve()) for p in sys.path))"],
        check=True, capture_output=True, text=True,
    )

    assert controller_site not in completed.stdout.splitlines()


def test_unexpected_sdist_member_is_rejected():
    checker = load_package_checker()
    with pytest.raises(AssertionError, match="unexpected sdist members"):
        checker.reject_unexpected_members({"setup.py", "private-notes.txt"}, {"setup.py"})


def test_sdist_member_name_normalizes_nested_windows_path():
    checker = load_package_checker()
    root = PureWindowsPath(r"C:\release\spl_toolkit-0.1.1")
    member = root / "_native_src" / "pkg" / "mapper" / "mapper.go"

    assert checker.sdist_member_name(member, root) == "_native_src/pkg/mapper/mapper.go"


def test_native_architecture_reads_supported_binary_headers():
    checker = load_package_checker()

    elf = bytearray(64)
    elf[:6] = b"\x7fELF\x02\x01"
    elf[18:20] = (62).to_bytes(2, "little")
    macho = b"\xcf\xfa\xed\xfe" + (0x0100000C).to_bytes(4, "little")

    assert checker.native_architecture(bytes(elf)) == "x86_64"
    assert checker.native_architecture(macho) == "arm64"


def test_required_pytest_plugin_rejects_skips_and_writes_counts(tmp_path: Path):
    checker = load_package_checker()
    suite = tmp_path / "suite"
    suite.mkdir()
    (suite / "test_required.py").write_text(
        "import pytest\n\ndef test_pass(): pass\n\n@pytest.mark.skip(reason='no')\ndef test_skip(): pass\n",
        encoding="utf-8",
    )
    result = tmp_path / "counts.json"
    checker.write_required_pytest_plugin(suite)
    completed = subprocess.run(
        [sys.executable, "-m", "pytest", str(suite), "-q"],
        env=checker.clean_env() | {"SPL_TEST_COUNTS": str(result)},
        check=False,
    )
    counts = json.loads(result.read_text(encoding="utf-8"))
    assert completed.returncode != 0
    assert counts == {"collected": 2, "passed": 1, "failed": 0, "skipped": 1}


def test_restore_verified_executables_changes_only_named_payloads(tmp_path: Path):
    checker = load_package_checker()
    cli = tmp_path / "cli"
    server = tmp_path / "server"
    plain = tmp_path / "plain.txt"
    for path in (cli, server, plain):
        path.write_bytes(path.name.encode())
        path.chmod(0o644)
    hashes = {path.name: checker.sha256(path) for path in (cli, server, plain)}

    checker.restore_verified_executables(tmp_path, hashes, ("cli", "server"))

    assert cli.stat().st_mode & 0o111
    assert server.stat().st_mode & 0o111
    assert not (plain.stat().st_mode & 0o111)


@pytest.mark.parametrize("existing", ["final", "staging"])
def test_evidence_write_preserves_existing_files(tmp_path: Path, existing: str):
    checker = load_package_checker()
    destination = tmp_path / "evidence.json"
    occupied = destination if existing == "final" else destination.with_suffix(".json.tmp")
    occupied.write_text("original\n", encoding="utf-8")

    with pytest.raises(FileExistsError, match="evidence"):
        checker.write_evidence(destination, {"status": "passed"})

    assert occupied.read_text(encoding="utf-8") == "original\n"
    if existing == "staging":
        assert not destination.exists()
    else:
        assert destination.read_text(encoding="utf-8") == "original\n"


def test_release_tree_contains_complete_analysis_source_closure(tmp_path: Path):
    support = load_build_support()
    distribution = support.NativeDistribution({"script_name": str(PYTHON_DIR / "setup.py")})
    command = support.SourceDistribution(distribution)
    command.ensure_finalized()
    release = tmp_path / "release"
    command.make_release_tree(str(release), [])
    required = [p for p in (ROOT / "pkg/analysis").glob("*.go") if not p.name.endswith("_test.go")]
    required.append(ROOT / "internal/jsoninput/unicode.go")
    for path in required:
        assert (release / "_native_src" / path.relative_to(ROOT)).read_bytes() == path.read_bytes()


def test_installed_native_suite_includes_analysis(tmp_path: Path):
    checker = load_package_checker()
    destination = tmp_path / "tests"
    checker._copy_required_files(ROOT / "python/tests", destination, checker.NATIVE_TESTS)
    assert (destination / "test_native_analysis.py").read_bytes() == (ROOT / "python/tests/test_native_analysis.py").read_bytes()
    assert "tests/test_native_analysis.py" in checker.SDIST_FIXED_FILES
