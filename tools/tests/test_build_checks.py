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


def init_repository(root: Path, files: dict[str, str | bytes]) -> None:
    run("git", "init", "-q", cwd=root)
    run("git", "config", "user.name", "Build Check Test", cwd=root)
    run("git", "config", "user.email", "build-check@example.invalid", cwd=root)
    for name, contents in files.items():
        path = root / name
        path.parent.mkdir(parents=True, exist_ok=True)
        if isinstance(contents, bytes):
            path.write_bytes(contents)
        else:
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
    def test_requested_race_timeout_terminates_a_slow_test(self) -> None:
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            init_repository(
                root,
                {
                    "go.mod": "module example.invalid/check\n\ngo 1.22\n",
                    "main.go": "package check\n",
                    "main_test.go": (
                        "package check\n\n"
                        "import (\n\t\"testing\"\n\t\"time\"\n)\n\n"
                        "func TestSlow(t *testing.T) {\n"
                        "\ttime.Sleep(200 * time.Millisecond)\n"
                        "}\n"
                    ),
                },
            )

            result = run(
                sys.executable,
                str(TOOLS / "check_go.py"),
                "--race-timeout=1ms",
                cwd=root,
                env={**os.environ, "GOTOOLCHAIN": "local"},
            )

        self.assertNotEqual(result.returncode, 0, result.stdout)
        self.assertIn("test timed out after 1ms", result.stdout)

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

    def test_accepts_windows_line_endings_when_go_code_is_formatted(self) -> None:
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            init_repository(
                root,
                {
                    "go.mod": "module example.invalid/check\n\ngo 1.22\n",
                    "main.go": b"package check\r\n\r\nfunc Main() {}\r\n",
                },
            )

            result = run(
                sys.executable,
                str(TOOLS / "check_go.py"),
                cwd=root,
                env={**os.environ, "GOTOOLCHAIN": "local"},
            )

        self.assertEqual(result.returncode, 0, result.stdout)

    def test_excludes_generated_files_from_the_handwritten_format_gate(self) -> None:
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            init_repository(
                root,
                {
                    "go.mod": "module example.invalid/check\n\ngo 1.22\n",
                    "main.go": "package check\n",
                    "parser/generated.go": "package parser\nfunc Generated( ){ }\n",
                    "parser/spl2/generated.go": "package spl2\nfunc Generated( ){ }\n",
                    "pkg/analysis/analysis.go": "package analysis\n",
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

        exclusion = next(line for line in result.stdout.splitlines() if line.startswith("excluding generated"))
        self.assertIn("example.invalid/check/parser/spl2", exclusion)
        self.assertNotIn("example.invalid/check/pkg/analysis", exclusion)

    def test_lint_does_not_write_missing_test_dependency_checksums(self) -> None:
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            init_repository(
                root,
                {
                    "go.mod": (
                        "module example.invalid/check\n\n"
                        "go 1.22\n\n"
                        "require github.com/pkg/errors v0.9.1\n"
                    ),
                    "main.go": "package check\n",
                    "main_test.go": (
                        "package check\n\n"
                        "import (\n"
                        '\t"testing"\n\n'
                        '\t"github.com/pkg/errors"\n'
                        ")\n\n"
                        "func TestDependency(t *testing.T) {\n"
                        '\tif errors.New("sentinel") == nil {\n'
                        '\t\tt.Fatal("expected an error")\n'
                        "\t}\n"
                        "}\n"
                    ),
                },
            )
            go_mod_before = (root / "go.mod").read_bytes()

            result = run(
                sys.executable,
                str(TOOLS / "check_go.py"),
                cwd=root,
                env={**os.environ, "GOFLAGS": "-mod=mod", "GOTOOLCHAIN": "local"},
            )

            self.assertEqual((root / "go.mod").read_bytes(), go_mod_before)
            self.assertFalse((root / "go.sum").exists(), result.stdout)
        self.assertNotEqual(result.returncode, 0, result.stdout)

    def test_vets_nested_handwritten_package_named_parser(self) -> None:
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            init_repository(
                root,
                {
                    "go.mod": "module example.invalid/check\n\ngo 1.22\n",
                    "main.go": "package check\n",
                    "internal/parser/value.go": (
                        "package parser\n\n"
                        "type Value struct {\n"
                        '\tField string `json:"field`\n'
                        "}\n"
                    ),
                },
            )

            result = run(
                sys.executable,
                str(TOOLS / "check_go.py"),
                cwd=root,
                env={**os.environ, "GOTOOLCHAIN": "local"},
            )

        self.assertNotEqual(result.returncode, 0, result.stdout)


if __name__ == "__main__":
    unittest.main()
