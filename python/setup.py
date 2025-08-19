"""
Setup script for SPL Toolkit Python bindings
"""

from setuptools import setup, find_packages
import os
import subprocess
import shutil
from pathlib import Path

# Read the README file
readme_path = Path(__file__).parent.parent / "README.md"
if readme_path.exists():
    with open(readme_path, "r", encoding="utf-8") as f:
        long_description = f.read()
else:
    long_description = "SPL Toolkit - Python bindings for programmatic SPL query analysis and manipulation"

# Build the Go shared library
def build_go_library():
    """Build the Go shared library"""
    print("Building Go shared library...")
    
    # Change to the project root directory
    project_root = Path(__file__).parent.parent
    os.chdir(project_root)
    
    # Build the shared library
    cmd = [
        "go", "build", 
        "-buildmode=c-shared",
        "-o", "python/spl_toolkit/libspl_toolkit.so",
        "./pkg/bindings"
    ]
    
    try:
        result = subprocess.run(cmd, check=True, capture_output=True, text=True)
        print("Go library built successfully")
        return True
    except subprocess.CalledProcessError as e:
        print(f"Failed to build Go library: {e}")
        print(f"stdout: {e.stdout}")
        print(f"stderr: {e.stderr}")
        return False
    except FileNotFoundError:
        print("Go compiler not found. Please install Go.")
        return False

# Custom build command
class BuildCommand:
    def run(self):
        if not build_go_library():
            raise RuntimeError("Failed to build Go shared library")

# Try to build the library during setup
try:
    build_go_library()
except Exception as e:
    print(f"Warning: Could not build Go library during setup: {e}")
    print("You may need to build it manually with: go build -buildmode=c-shared -o python/spl_toolkit/libspl_toolkit.so ./pkg/bindings")

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
        ],
    },
    package_data={
        "spl_toolkit": [
            "libspl_toolkit.so",
            "libspl_toolkit.dylib",
            "*.h",
        ],
    },
    include_package_data=True,
    zip_safe=False,  # Due to shared library
    keywords="splunk spl query parser field mapping analysis",
    project_urls={
        "Bug Reports": "https://github.com/delgado-jacob/spl-toolkit/issues",
        "Source": "https://github.com/delgado-jacob/spl-toolkit",
        "Documentation": "https://github.com/delgado-jacob/spl-toolkit/docs",
    },
)