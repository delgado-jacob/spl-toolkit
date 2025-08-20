"""
Setup script for SPL Toolkit Python bindings
"""

from setuptools import setup, find_packages
import os
import subprocess
from pathlib import Path
import platform
import shutil
import sys

# Detect shared library extension based on OS
SYSTEM = platform.system()
if SYSTEM == "Windows":
    SHARED_EXT = ".dll"
elif SYSTEM == "Darwin":
    SHARED_EXT = ".dylib"
else:
    SHARED_EXT = ".so"

# Library base name
LIB_BASE_NAME = "libspl_toolkit"

# Read the README file
readme_path = Path(__file__).parent.parent / "README.md"
if readme_path.exists():
    with open(readme_path, "r", encoding="utf-8") as f:
        long_description = f.read()
else:
    long_description = "SPL Toolkit - Python bindings for programmatic SPL query analysis and manipulation"

def find_go_library():
    """Find an existing Go shared library"""
    # Look for the library in several possible locations
    possible_locations = [
        Path(__file__).parent / "spl_toolkit" / f"{LIB_BASE_NAME}{SHARED_EXT}",
        Path(__file__).parent / f"{LIB_BASE_NAME}{SHARED_EXT}",
        # Also check without the 'lib' prefix on Windows
        Path(__file__).parent / "spl_toolkit" / f"spl_toolkit{SHARED_EXT}",
        Path(__file__).parent / f"spl_toolkit{SHARED_EXT}",
        ]

    for path in possible_locations:
        if path.exists():
            print(f"Found existing Go library at: {path}")
            return path

    return None

def build_go_library():
    """Build the Go shared library"""
    print("Building Go shared library...")

    # Check if we're in cibuildwheel environment
    in_cibuildwheel = os.environ.get("CIBUILDWHEEL") == "1"

    # If in cibuildwheel, the library should already be built
    if in_cibuildwheel:
        existing_lib = find_go_library()
        if existing_lib:
            # Ensure it's in the right place
            target_path = Path(__file__).parent / "spl_toolkit" / f"{LIB_BASE_NAME}{SHARED_EXT}"
            if existing_lib != target_path:
                target_path.parent.mkdir(parents=True, exist_ok=True)
                shutil.copy2(existing_lib, target_path)
                print(f"Copied library to: {target_path}")
            return True
        else:
            print("WARNING: In cibuildwheel but library not found. Will attempt to build.")

    # Determine the project root
    setup_dir = Path(__file__).parent
    if setup_dir.name == "python":
        project_root = setup_dir.parent
    else:
        project_root = setup_dir

    # Build the shared library
    output_path = setup_dir / "spl_toolkit" / f"{LIB_BASE_NAME}{SHARED_EXT}"
    output_path.parent.mkdir(parents=True, exist_ok=True)

    # Build command
    cmd = [
        "go", "build",
        "-buildmode=c-shared",
        "-o", str(output_path),
        "./pkg/bindings",
    ]

    env = os.environ.copy()
    # Ensure CGO is enabled for c-shared builds
    env["CGO_ENABLED"] = "1"

    # Handle cross-compilation for macOS universal builds
    if SYSTEM == "Darwin" and "ARCHFLAGS" in os.environ:
        if "arm64" in os.environ["ARCHFLAGS"]:
            env["GOARCH"] = "arm64"
        elif "x86_64" in os.environ["ARCHFLAGS"]:
            env["GOARCH"] = "amd64"

    try:
        print(f"Running: {' '.join(cmd)}")
        print(f"Working directory: {project_root}")
        result = subprocess.run(
            cmd,
            check=True,
            capture_output=True,
            text=True,
            env=env,
            cwd=str(project_root)
        )

        # Verify output file exists
        if not output_path.exists():
            print(f"Go build reported success but output not found: {output_path}")
            return False

        print(f"Go library built successfully: {output_path}")

        # On Linux, also create a symlink without 'lib' prefix if needed
        if SYSTEM == "Linux":
            alt_name = output_path.parent / f"spl_toolkit{SHARED_EXT}"
            if not alt_name.exists():
                alt_name.symlink_to(output_path.name)
                print(f"Created symlink: {alt_name}")

        return True

    except subprocess.CalledProcessError as e:
        print(f"Failed to build Go library: {e}")
        print(f"stdout: {e.stdout}")
        print(f"stderr: {e.stderr}")
        return False
    except FileNotFoundError:
        print("Go compiler not found. Please install Go.")
        return False

# Custom build commands
from setuptools.command.build_py import build_py as _build_py
from setuptools.command.build_ext import build_ext as _build_ext
from setuptools import Extension

# Create a dummy extension to force platlib
# This tells setuptools that this is a platform-specific package
ext_modules = [
    Extension("spl_toolkit._dummy", sources=[])
]

class build_py_go(_build_py):
    def run(self):
        # Build the Go shared library before packaging Python modules
        if not find_go_library():
            success = build_go_library()
            if not success:
                # Only fail if we're in a CI environment or explicitly required
                if os.environ.get("CIBUILDWHEEL") == "1" or os.environ.get("FORCE_BUILD_GO") == "1":
                    raise RuntimeError("Failed to build Go shared library. See logs above.")
                else:
                    print("WARNING: Could not build Go library. The package may not work correctly.")
        super().run()

class build_ext_go(_build_ext):
    def run(self):
        # Also handle in build_ext for compatibility
        if not find_go_library():
            build_go_library()
        # Skip actual extension building since we don't have real C extensions
        pass

# Handle bdist_wheel only if wheel is installed
try:
    from wheel.bdist_wheel import bdist_wheel

    class bdist_wheel_plat(bdist_wheel):
        def finalize_options(self):
            super().finalize_options()
            # Mark the wheel as non-pure so it gets a platform tag
            self.root_is_pure = False

        def get_tag(self):
            python, abi, plat = super().get_tag()
            # Ensure we get a platform-specific tag
            if plat == "any":
                plat = self.plat_name
            return python, abi, plat

    cmdclass = {
        "build_py": build_py_go,
        "build_ext": build_ext_go,
        "bdist_wheel": bdist_wheel_plat,
    }
except ImportError:
    # wheel not installed, just use basic commands
    cmdclass = {
        "build_py": build_py_go,
        "build_ext": build_ext_go,
    }

# Determine package data based on what files exist
package_data_files = [
    f"{LIB_BASE_NAME}{SHARED_EXT}",
    f"spl_toolkit{SHARED_EXT}",  # Alternative name
    f"{LIB_BASE_NAME}.h",
    "*.so",
    "*.dylib",
    "*.dll",
]

# If we're building a source distribution, don't require the library
if "sdist" not in sys.argv:
    # Try to ensure the library exists or can be built
    if not find_go_library() and not build_go_library():
        print("\n" + "="*60)
        print("WARNING: Go shared library not found and could not be built.")
        print("The package is being built anyway, but may not function correctly.")
        print("To build the library manually, run:")
        print(f"  go build -buildmode=c-shared -o python/spl_toolkit/{LIB_BASE_NAME}{SHARED_EXT} ./pkg/bindings")
        print("="*60 + "\n")

setup(
    name="spl-toolkit",
    version="0.1.0",
    author="SPL Toolkit Team",
    author_email="team@example.com",
    description="Python bindings for SPL Toolkit - programmatic SPL query analysis and manipulation",
    long_description=long_description,
    long_description_content_type="text/markdown",
    url="https://github.com/delgado-jacob/spl-toolkit",
    packages=find_packages(),
    license="MIT",
    classifiers=[
        "Development Status :: 4 - Beta",
        "Intended Audience :: Developers",
        "Intended Audience :: System Administrators",
        "Operating System :: OS Independent",
        "Programming Language :: Python :: 3",
        "Programming Language :: Python :: 3.8",
        "Programming Language :: Python :: 3.9",
        "Programming Language :: Python :: 3.10",
        "Programming Language :: Python :: 3.11",
        "Programming Language :: Python :: 3.12",
        "Programming Language :: Go",
        "Topic :: Software Development :: Libraries :: Python Modules",
        "Topic :: System :: Systems Administration",
        "Topic :: Text Processing :: Linguistic",
    ],
    python_requires=">=3.8",
    install_requires=[
        # No external Python dependencies - all logic is in Go
    ],
    extras_require={
        "dev": [
            "pytest>=6.0",
            "pytest-cov>=2.0",
            "black>=22.0",
            "flake8>=4.0",
            "mypy>=0.910",
            "wheel>=0.37.0",  # Add wheel as dev dependency
            "setuptools>=45.0",
        ],
    },
    package_data={
        "spl_toolkit": package_data_files,
    },
    include_package_data=True,
    zip_safe=False,  # Due to shared library
    ext_modules=ext_modules,  # Add dummy extension to mark as platlib
    cmdclass=cmdclass,
    keywords="splunk spl query parser field mapping analysis",
    project_urls={
        "Bug Reports": "https://github.com/delgado-jacob/spl-toolkit/issues",
        "Source": "https://github.com/delgado-jacob/spl-toolkit",
        "Documentation": "https://github.com/delgado-jacob/spl-toolkit/docs",
    },
)