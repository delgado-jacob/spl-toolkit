"""Build entry point for SPL Toolkit's native Python package."""

import importlib.util
from pathlib import Path

from setuptools import setup

SETUP_DIR = Path(__file__).resolve().parent
SPEC = importlib.util.spec_from_file_location("spl_toolkit_build_support", SETUP_DIR / "build_support.py")
if SPEC is None or SPEC.loader is None:
    raise RuntimeError("could not load build_support.py")
BUILD_SUPPORT = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(BUILD_SUPPORT)

setup(
    version=BUILD_SUPPORT.read_version(SETUP_DIR),
    distclass=BUILD_SUPPORT.NativeDistribution,
    cmdclass={
        "build_py": BUILD_SUPPORT.BuildPy,
        "sdist": BUILD_SUPPORT.SourceDistribution,
        "bdist_wheel": BUILD_SUPPORT.BinaryWheel,
    },
)
