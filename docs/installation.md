---
title: "Installation"
layout: page
---

# Installation

Source builds require Go 1.22.12 or newer and Python 3.11 or newer. Native Python builds also require a C compiler. The pinned release toolchain is Go 1.26.8; use it on current macOS hosts. Release wheels are platform-specific and contain the native library, so their installation and runtime do not require Go.

Build the current checkout with the repository targets:

```bash
make deps
make build-all
make python-build
```

Outputs are written under `build/` and `dist/`. Install a freshly built wheel rather than importing the package from the source directory:

```bash
python -m pip install dist/spl_toolkit-0.1.1-*.whl
```

Source distributions contain the native Go sources needed to build a wheel. Building one requires the pinned packages from `python/requirements-build.txt`, plus Go and a platform C compiler. Normal development and test dependencies are pinned in `python/requirements-dev.txt`.

Supported native release targets are Linux x86-64, macOS x86-64, macOS arm64, and Windows x86-64. Use a wheel matching the operating system and architecture. The Python wrapper raises `ConfigurationError` if package and native versions differ.

Verify from source with:

```bash
make test
make python-test
```

See the [CLI guide](cli.md) for installed commands. Use `SPLMapper` as a context manager or call `close()` when embedding Python.
