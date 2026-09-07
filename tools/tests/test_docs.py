from pathlib import Path
import subprocess
import sys
import tempfile
import unittest


CHECKER = Path(__file__).resolve().parents[1] / "check_docs.py"


class DocumentationTests(unittest.TestCase):
    def check(self, files):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            for name, content in files.items():
                path = root / name
                path.parent.mkdir(parents=True, exist_ok=True)
                path.write_text(content, encoding="utf-8")
            return subprocess.run(
                [sys.executable, str(CHECKER)], cwd=root, text=True,
                stdout=subprocess.PIPE, stderr=subprocess.STDOUT, check=False,
            )

    def test_repository_readme_and_internal_specs_are_not_site_pages(self):
        result = self.check({
            "README.md": "# Repository readme\n",
            "docs/_config.yml": "exclude:\n  - superpowers/\n",
            "docs/superpowers/specs/design.md": "# Internal design\n",
            "docs/api/go.md": "---\ntitle: Go API\n---\n# Go API\n",
        })
        self.assertEqual(result.returncode, 0, result.stdout)

    def test_nested_public_pages_still_require_valid_front_matter(self):
        for content in ("# Missing\n", "---\ntitle: Unclosed\n", "---\ntitle: [invalid\n---\n"):
            with self.subTest(content=content):
                result = self.check({
                    "docs/_config.yml": "exclude:\n  - superpowers/\n",
                    "docs/api/new.md": content,
                })
                self.assertNotEqual(result.returncode, 0, result.stdout)
                self.assertIn("docs/api/new.md", result.stdout)

    def test_internal_docs_are_validated_unless_excluded_from_site(self):
        result = self.check({
            "docs/_config.yml": "exclude: []\n",
            "docs/superpowers/specs/design.md": "# Design\n",
        })
        self.assertNotEqual(result.returncode, 0, result.stdout)
        self.assertIn("docs/superpowers/specs/design.md", result.stdout)


if __name__ == "__main__":
    unittest.main()
