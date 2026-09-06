"""Setuptools commands for self-contained SPL Toolkit native packages."""

from __future__ import annotations

import os
from pathlib import Path
import platform
import shutil
import subprocess

from setuptools import Distribution
from setuptools.command.bdist_wheel import bdist_wheel
from setuptools.command.build_py import build_py
from setuptools.command.sdist import sdist


ROOT_FILES = ("README.md", "LICENSE", "VERSION")
NATIVE_SOURCE_MANIFEST = "native-source-files.txt"


def native_library_name() -> str:
    if os.name == "nt":
        return "libspl_toolkit.dll"
    if platform.system() == "Darwin":
        return "libspl_toolkit.dylib"
    return "libspl_toolkit.so"


def macos_architecture() -> str:
    architecture = os.environ.get("GOARCH", platform.machine()).lower()
    if architecture in ("arm64", "aarch64"):
        return "arm64"
    if architecture in ("amd64", "x86_64"):
        return "x86_64"
    raise RuntimeError(f"unsupported macOS wheel architecture: {architecture}")


def required_platform_tag() -> str | None:
    if platform.system() == "Darwin":
        return f"macosx_15_0_{macos_architecture()}"
    return None


def source_root(setup_dir: Path) -> Path:
    setup_dir = setup_dir.resolve()
    checkout = setup_dir.parent
    if (checkout / "go.mod").is_file():
        return checkout
    staged = setup_dir / "_native_src"
    if (staged / "go.mod").is_file():
        return staged
    raise FileNotFoundError(f"native source not found: {staged / 'go.mod'}")


def read_version(setup_dir: Path) -> str:
    setup_dir = setup_dir.resolve()
    candidates = []
    if (setup_dir.parent / "go.mod").is_file():
        candidates.append(setup_dir.parent / "VERSION")
    candidates.append(setup_dir / "VERSION")
    for version_file in candidates:
        if version_file.is_file():
            version = version_file.read_text(encoding="utf-8").strip()
            if version:
                return version
            raise ValueError(f"empty version file: {version_file}")
    raise FileNotFoundError(f"version file not found for {setup_dir}")


def build_native(source: Path, output: Path, version: str) -> None:
    output.parent.mkdir(parents=True, exist_ok=True)
    env = os.environ.copy()
    env["CGO_ENABLED"] = "1"
    env["GOTOOLCHAIN"] = "local"
    if platform.system() == "Darwin":
        env.setdefault("MACOSX_DEPLOYMENT_TARGET", "15.0")
        env["CGO_CFLAGS"] = f"{env.get('CGO_CFLAGS', '')} -mmacosx-version-min=15.0".strip()
        env["CGO_LDFLAGS"] = f"{env.get('CGO_LDFLAGS', '')} -mmacosx-version-min=15.0".strip()
    subprocess.run([
        "go", "build", "-mod=readonly", "-trimpath", "-buildvcs=false",
        "-buildmode=c-shared", "-ldflags",
        f"-X=github.com/delgado-jacob/spl-toolkit/internal/buildinfo.Version={version}",
        "-o", str(output.resolve()), "./pkg/bindings",
    ], cwd=source, env=env, check=True)
    if not output.is_file():
        raise RuntimeError("native build produced no shared library")


class NativeDistribution(Distribution):
    """Mark wheels as platform-specific without a dummy Python extension."""

    def has_ext_modules(self) -> bool:
        return True


class BuildPy(build_py):
    """Copy Python modules, then compile the native payload into build_lib."""

    def run(self) -> None:
        super().run()
        setup_dir = Path(self.distribution.script_name).resolve().parent
        output = Path(self.build_lib) / "spl_toolkit" / native_library_name()
        build_native(source_root(setup_dir), output, read_version(setup_dir))


def native_source_files(manifest: Path) -> list[Path]:
    entries = []
    for line in manifest.read_text(encoding="utf-8").splitlines():
        value = line.strip()
        if not value or value.startswith("#"):
            continue
        relative = Path(value)
        if relative.is_absolute() or ".." in relative.parts:
            raise ValueError(f"invalid native source path: {value}")
        entries.append(relative)
    if len(entries) != len(set(entries)):
        raise ValueError(f"duplicate native source path in {manifest}")
    return entries


def stage_native_source(source: Path, destination: Path, manifest: Path) -> None:
    for relative in native_source_files(manifest):
        source_path = source / relative
        if not source_path.is_file():
            raise FileNotFoundError(f"native source manifest entry is missing: {source_path}")
        target = destination / relative
        target.parent.mkdir(parents=True, exist_ok=True)
        shutil.copy2(source_path, target)


class SourceDistribution(sdist):
    """Stage the precise native source allowlist in the release tree."""

    def make_release_tree(self, base_dir: str, files: list[str]) -> None:
        super().make_release_tree(base_dir, files)
        destination = Path(base_dir)
        setup_dir = Path(self.distribution.script_name).resolve().parent
        source = source_root(setup_dir)
        documents = setup_dir if source.name == "_native_src" else source
        for name in ROOT_FILES:
            shutil.copy2(documents / name, destination / name)
        stage_native_source(source, destination / "_native_src", setup_dir / NATIVE_SOURCE_MANIFEST)


class BinaryWheel(bdist_wheel):
    """Use one Python-independent tag for the bundled C ABI library."""

    def finalize_options(self) -> None:
        super().finalize_options()
        self.root_is_pure = False
        required = required_platform_tag()
        if required is not None:
            self.plat_name = required
            self.plat_name_supplied = True

    def get_tag(self) -> tuple[str, str, str]:
        _python, _abi, platform_tag = super().get_tag()
        return "py3", "none", required_platform_tag() or platform_tag
