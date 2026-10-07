"""Validate front matter for Markdown pages in the Jekyll docs source."""

import sys
import os
from pathlib import Path
from html.parser import HTMLParser
from urllib.parse import unquote, urljoin, urlsplit


def check_site_links(site_dir, site_url):
    """Check published site destinations against the rendered build, not live Pages."""
    site = Path(site_dir).resolve()
    base = urlsplit(site_url)
    prefix = base.path.rstrip("/")
    missing = []

    class Links(HTMLParser):
        def __init__(self):
            super().__init__()
            self.urls = []

        def handle_starttag(self, tag, attrs):
            attrs = dict(attrs)
            self.urls.extend(value for key, value in attrs.items()
                             if key in ("href", "src", "action", "poster") and value)
            if tag == "meta" and attrs.get("property") == "og:url":
                self.urls.append(attrs.get("content", ""))

    pages = sorted(site.rglob("*.html"))
    if not pages:
        return [f"{site}: no rendered HTML pages"]
    for page in pages:
        parser = Links()
        parser.feed(page.read_text(encoding="utf-8"))
        page_url = site_url.rstrip("/") + "/" + page.relative_to(site).as_posix()
        for link in parser.urls:
            target = urlsplit(urljoin(page_url, link))
            if target.scheme not in ("http", "https") or target.netloc != base.netloc:
                continue
            path = unquote(target.path)
            if prefix:
                if path != prefix and not path.startswith(prefix + "/"):
                    # Absolute URLs outside this project remain network checks.
                    # Lychee 0.15 drops root-relative links, so reject escapes.
                    if not urlsplit(link).scheme:
                        missing.append(f"{page}: link outside published site prefix: {link}")
                    continue
                path = path[len(prefix):]
            destination = (site / path.lstrip("/")).resolve()
            if not destination.is_relative_to(site):
                missing.append(f"{page}: link escapes rendered site: {link}")
            elif not (destination.is_file() or (destination / "index.html").is_file()):
                missing.append(f"{page}: missing site destination: {link}")
    return missing


# Rendered links need their own mode: lychee 0.15 drops root-relative links
# without --base, while --base resolves nested relative links at the site root.
if len(sys.argv) > 1:
    import argparse
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--site-dir", required=True)
    parser.add_argument("--site-url", required=True)
    args = parser.parse_args()
    failures = check_site_links(args.site_dir, args.site_url)
    for failure in failures:
        print(failure)
    sys.exit(bool(failures))

import yaml

errors = []

# The resolution landing page and committed request are part of published docs.
resolution_page = Path('docs/resolution.md')
resolution_request = Path('examples/resolution/request.json')
if not resolution_page.is_file() or not resolution_request.is_file():
    errors.append('resolution documentation or complete request missing')
else:
    import json
    try:
        json.loads(resolution_request.read_text(encoding='utf-8'))
    except (ValueError, OSError) as error:
        errors.append(f'resolution example: {error}')

def check_yaml_frontmatter(file_path):
    try:
        with open(file_path, 'r', encoding='utf-8') as f:
            content = f.read()
    except Exception as e:
        print(f'✗ {file_path}: Cannot read file - {e}')
        errors.append(file_path)
        return

    if not content.startswith('---'):
        print(f'✗ {file_path}: Missing YAML front matter (must start at first line)')
        errors.append(file_path)
        return

    lines = content.splitlines()
    if not lines or lines[0].strip() != '---':
        print(f'✗ {file_path}: Missing starting --- delimiter')
        errors.append(file_path)
        return

    # Find ending delimiter line
    end_idx = None
    for i in range(1, len(lines)):
        if lines[i].strip() == '---':
            end_idx = i
            break

    if end_idx is None:
        print(f'✗ {file_path}: Incomplete YAML front matter (no closing ---)')
        errors.append(file_path)
        return

    yaml_content = '\n'.join(lines[1:end_idx])
    try:
        yaml.safe_load(yaml_content) if yaml_content.strip() else {}
        print(f'✓ {file_path}: Valid YAML front matter')
    except yaml.YAMLError as e:
        print(f'✗ {file_path}: Invalid YAML front matter - {e}')
        errors.append(file_path)

# Share the site's explicit path exclusions; repository README is outside
# the Jekyll source and does not require page metadata.
source = Path('docs')
config = yaml.safe_load((source / '_config.yml').read_text(encoding='utf-8'))
excluded = {source / path for path in config.get('exclude', [])}
for root, dirs, files in os.walk(source):
    dirs[:] = sorted(name for name in dirs if Path(root) / name not in excluded)
    for file in sorted(files):
        path = Path(root) / file
        if path.suffix == '.md' and path not in excluded:
            check_yaml_frontmatter(path)

if errors:
    print('\nFront matter validation failed for files:')
    for f in errors:
        print(f' - {f}')
    sys.exit(1)
