import os
from pathlib import Path
import subprocess
import sys
import tempfile
import unittest


TOOLS = Path(__file__).resolve().parents[1]


def run(*args: str, cwd: Path, env: dict[str, str] | None = None) -> subprocess.CompletedProcess[str]:
    return subprocess.run(
        args,
        cwd=cwd,
        env=env,
        text=True,
        stdout=subprocess.PIPE,
        stderr=subprocess.STDOUT,
        check=False,
    )


def init_repository(root: Path, files: dict[str, str]) -> None:
    run("git", "init", "-q", cwd=root)
    run("git", "config", "user.name", "Build Check Test", cwd=root)
    run("git", "config", "user.email", "build-check@example.invalid", cwd=root)
    for name, contents in files.items():
        path = root / name
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text(contents, encoding="utf-8")
    run("git", "add", ".", cwd=root)
    result = run("git", "commit", "-q", "-m", "fixture", cwd=root)
    if result.returncode:
        raise AssertionError(result.stdout)


class CleanBuildTests(unittest.TestCase):
    def run_check(self, files: dict[str, str]) -> subprocess.CompletedProcess[str]:
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            init_repository(root, files)
            return run(
                sys.executable,
                str(TOOLS / "check_clean_build.py"),
                "--ref",
                "HEAD",
                cwd=root,
            )

    def test_reports_the_tracked_file_changed_by_a_build(self) -> None:
        result = self.run_check(
            {
                "Makefile": (
                    "build:\n\t@printf changed > tracked.txt\n"
                    "build-server build-shared test:\n\t@true\n"
                ),
                "tracked.txt": "original\n",
            }
        )

        self.assertEqual(result.returncode, 1, result.stdout)
        self.assertIn("tracked.txt", result.stdout)

    def test_returns_the_build_subprocess_failure(self) -> None:
        result = self.run_check(
            {
                "Makefile": (
                    "build:\n\t@exit 7\n"
                    "build-server build-shared test:\n\t@true\n"
                )
            }
        )

        self.assertEqual(result.returncode, 2, result.stdout)
        self.assertIn("build command failed", result.stdout)


class GoCheckTests(unittest.TestCase):
    def test_rejects_unformatted_handwritten_go(self) -> None:
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            init_repository(
                root,
                {
                    "go.mod": "module example.invalid/check\n\ngo 1.22\n",
                    "bad.go": "package check\nfunc Bad( ){ }\n",
                },
            )

            result = run(
                sys.executable,
                str(TOOLS / "check_go.py"),
                cwd=root,
                env={**os.environ, "GOTOOLCHAIN": "local"},
            )

        self.assertEqual(result.returncode, 1, result.stdout)
        self.assertIn("bad.go", result.stdout)

    def test_excludes_generated_files_from_the_handwritten_format_gate(self) -> None:
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            init_repository(
                root,
                {
                    "go.mod": "module example.invalid/check\n\ngo 1.22\n",
                    "main.go": "package check\n",
                    "parser/generated.go": "package parser\nfunc Generated( ){ }\n",
                    "gen/generated.go": "package gen\nfunc Generated( ){ }\n",
                    "docs/docs.go": "package docs\nfunc Generated( ){ }\n",
                },
            )

            result = run(
                sys.executable,
                str(TOOLS / "check_go.py"),
                cwd=root,
                env={**os.environ, "GOTOOLCHAIN": "local"},
            )

        self.assertEqual(result.returncode, 0, result.stdout)


if __name__ == "__main__":
    unittest.main()
