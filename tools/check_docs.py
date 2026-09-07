"""Validate front matter for Markdown pages in the Jekyll docs source."""

import yaml
import sys
import os
from pathlib import Path

errors = []

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
