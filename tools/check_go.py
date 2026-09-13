#!/usr/bin/env python3
"""Run formatting, vet, and test checks for the canonical Go source."""

import argparse
import os
from pathlib import Path
import subprocess
import sys


GENERATED_FILES = {"docs/docs.go"}
GENERATED_DIRECTORIES = {"gen", "parser"}


def command(arguments: list[str], *, capture: bool = False) -> subprocess.CompletedProcess[str]:
    return subprocess.run(
        arguments,
        check=False,
        text=True,
        stdout=subprocess.PIPE if capture else None,
    )


def tracked_handwritten_go() -> list[str]:
    result = subprocess.run(
        ["git", "ls-files", "-z", "--", "*.go"],
        check=True,
        stdout=subprocess.PIPE,
    )
    names = [os.fsdecode(name) for name in result.stdout.split(b"\0") if name]
    return [
        name for name in names
        if name not in GENERATED_FILES and Path(name).parts[0] not in GENERATED_DIRECTORIES
    ]


def unformatted_handwritten_go(names: list[str]) -> tuple[list[str], int]:
    formatting = command(["gofmt", "-l", *names], capture=True)
    if formatting.returncode:
        return [], formatting.returncode

    unformatted = []
    for name in formatting.stdout.splitlines():
        rendered = subprocess.run(
            ["gofmt", name],
            check=False,
            stdout=subprocess.PIPE,
        )
        if rendered.returncode:
            return [], rendered.returncode
        source = Path(name).read_bytes().replace(b"\r\n", b"\n")
        if source != rendered.stdout:
            unformatted.append(name)
    return unformatted, 0


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--race-timeout", default="10m", help="Go race suite timeout (default: 10m)")
    args = parser.parse_args()

    try:
        handwritten = tracked_handwritten_go()
    except subprocess.CalledProcessError as error:
        return error.returncode or 1

    if handwritten:
        unformatted, returncode = unformatted_handwritten_go(handwritten)
        if returncode:
            return returncode
        if unformatted:
            print("handwritten Go files need formatting:", file=sys.stderr)
            for name in unformatted:
                print(name, file=sys.stderr)
            return 1

    packages_result = command(["go", "list", "-mod=readonly", "./..."], capture=True)
    if packages_result.returncode:
        return packages_result.returncode
    packages = packages_result.stdout.splitlines()
    module_result = command(["go", "list", "-mod=readonly", "-m"], capture=True)
    if module_result.returncode:
        return module_result.returncode
    module = module_result.stdout.strip()
    generated_package_names = {
        f"{module}/{name}" for name in GENERATED_DIRECTORIES | {"docs"}
    }
    generated_packages = [
        name for name in packages
        if any(name == prefix or name.startswith(prefix + "/") for prefix in generated_package_names)
    ]
    handwritten_packages = [name for name in packages if name not in generated_packages]

    if generated_packages:
        print(
            "excluding generated packages from handwritten go vet gate: "
            + ", ".join(generated_packages)
            + "; generated ANTLR and OpenAPI sources are committed build inputs",
            file=sys.stderr,
        )

    if handwritten_packages:
        vet = command(["go", "vet", "-mod=readonly", *handwritten_packages])
        if vet.returncode:
            return vet.returncode

    tests = command(["go", "test", "-mod=readonly", "-race", f"-timeout={args.race_timeout}", "./..."])
    return tests.returncode


if __name__ == "__main__":
    raise SystemExit(main())
