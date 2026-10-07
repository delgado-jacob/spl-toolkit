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
            files = {"docs/resolution.md": "---\ntitle: Resolution\n---\n",
                     "examples/resolution/request.json": "{}"} | files
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

    def test_retained_evidence_readme_exclusion_does_not_hide_sibling_pages(self):
        config = (CHECKER.parents[1] / "docs/_config.yml").read_text(encoding="utf-8")
        evidence = "docs/evidence/milestone-5/broad-fix-1/"
        files = {"docs/_config.yml": config, evidence + "README.md": "# Retained evidence\n"}
        result = self.check(files)
        self.assertEqual(result.returncode, 0, result.stdout)
        files[evidence + "public.md"] = "# Public page without metadata\n"
        result = self.check(files)
        self.assertNotEqual(result.returncode, 0, result.stdout)
        self.assertIn(evidence + "public.md", result.stdout)
        self.assertNotIn(evidence + "README.md", result.stdout)


class RenderedDocumentationTests(unittest.TestCase):
    def check(self, files):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            site = root / "site"
            site.mkdir()
            # An existing sibling must not satisfy a link escaping the build.
            (root / "secret.html").write_text("secret", encoding="utf-8")
            for name, content in files.items():
                path = site / name
                path.parent.mkdir(parents=True, exist_ok=True)
                path.write_text(content, encoding="utf-8")
            return subprocess.run(
                [sys.executable, str(CHECKER), "--site-dir", str(site),
                 "--site-url", "https://example.org/project/"],
                text=True, stdout=subprocess.PIPE, stderr=subprocess.STDOUT,
                check=False,
            )

    def test_published_links_resolve_against_build(self):
        result = self.check({
            "index.html": '<link rel="canonical" href="https://example.org/project/index.html">'
                          '<a href="/project/api/">API</a>'
                          '<img src="/project/assets/style.css">'
                          '<a href="https://other.example/missing.html">External</a>'
                          '<a href="//other.example/missing.html">External</a>'
                          '<a href="mailto:owner@example.org">Mail</a>',
            "api/index.html": '<a href="go.html?view=all#usage">Go</a>'
                              '<a href="../index.html">Home</a>',
            "api/go.html": '<meta property="og:url" content="https://example.org/project/api/go.html">',
            "assets/style.css": "body {}",
        })
        self.assertEqual(result.returncode, 0, result.stdout)

    def test_missing_same_site_destinations_fail(self):
        for link in ("missing.html", "/project/missing.html", "/project/missing/",
                     "https://example.org/project/missing.html",
                     "//example.org/project/missing.html"):
            with self.subTest(link=link):
                result = self.check({"api/index.html": f'<a href="{link}">Broken</a>'})
                self.assertNotEqual(result.returncode, 0, result.stdout)
                self.assertIn(link, result.stdout)

    def test_missing_project_prefix_cannot_match_existing_asset(self):
        for link in ("/assets/style.css", "//example.org/assets/style.css", "../assets/style.css"):
            with self.subTest(link=link):
                result = self.check({
                    "index.html": f'<img src="{link}">',
                    "assets/style.css": "body {}",
                })
                self.assertNotEqual(result.returncode, 0, result.stdout)
                self.assertIn("outside published site prefix", result.stdout)

    def test_absolute_same_host_url_outside_project_is_left_for_lychee(self):
        result = self.check({
            "index.html": '<img src="https://example.org/assets/style.css">',
        })
        self.assertEqual(result.returncode, 0, result.stdout)

    def test_project_root_with_or_without_slash_uses_index(self):
        result = self.check({
            "index.html": '<a href="/project">Home</a><a href="/project/">Home</a>',
        })
        self.assertEqual(result.returncode, 0, result.stdout)

    def test_escaping_link_cannot_use_file_outside_build(self):
        result = self.check({"index.html": '<a href="/project/%2e%2e/secret.html">Escapes</a>'})
        self.assertNotEqual(result.returncode, 0, result.stdout)
        self.assertIn("escapes rendered site", result.stdout)

    def test_empty_rendered_site_fails(self):
        result = self.check({})
        self.assertNotEqual(result.returncode, 0, result.stdout)
        self.assertIn("no rendered HTML pages", result.stdout)


if __name__ == "__main__":
    unittest.main()
