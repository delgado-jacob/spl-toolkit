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


NATIVE_SOURCE_PATHS = (
    "go.mod", "go.sum", "pkg/mapper", "pkg/bindings", "parser", "internal/buildinfo",
)
ROOT_FILES = ("README.md", "LICENSE", "VERSION")
NATIVE_SUFFIXES = (".so", ".dylib", ".dll")


def native_library_name() -> str:
    if os.name == "nt":
        return "libspl_toolkit.dll"
    if platform.system() == "Darwin":
        return "libspl_toolkit.dylib"
    return "libspl_toolkit.so"


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


def _ignore_native_build_products(_directory: str, names: list[str]) -> set[str]:
    ignored = {name for name in names if name in {"__pycache__", ".pytest_cache", ".git", "_build_plan"}}
    ignored.update(name for name in names if Path(name).suffix in NATIVE_SUFFIXES)
    return ignored


class SourceDistribution(sdist):
    """Stage the precise native source allowlist in the release tree."""

    def make_release_tree(self, base_dir: str, files: list[str]) -> None:
        super().make_release_tree(base_dir, files)
        destination = Path(base_dir)
        setup_dir = Path(self.distribution.script_name).resolve().parent
        source = source_root(setup_dir)
        for name in ROOT_FILES:
            shutil.copy2(source / name, destination / name)
        native = destination / "_native_src"
        for relative in NATIVE_SOURCE_PATHS:
            source_path = source / relative
            target = native / relative
            target.parent.mkdir(parents=True, exist_ok=True)
            if source_path.is_dir():
                shutil.copytree(source_path, target, ignore=_ignore_native_build_products)
            else:
                shutil.copy2(source_path, target)


class BinaryWheel(bdist_wheel):
    """Use one Python-independent tag for the bundled C ABI library."""

    def finalize_options(self) -> None:
        super().finalize_options()
        self.root_is_pure = False

    def get_tag(self) -> tuple[str, str, str]:
        _python, _abi, platform_tag = super().get_tag()
        return "py3", "none", platform_tag
