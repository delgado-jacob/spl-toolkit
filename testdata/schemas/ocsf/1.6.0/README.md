# Official compiled OCSF 1.6.0 test catalogs

These are genuine normal `compile_version: 1` outputs of the official OCSF
compiler. Tests decompress them, verify the pinned raw SHA-256, then pass the
bytes to the offline target preparation API. Neither production nor tests run
the compiler or retrieve schemas. `../edge-cases.json` is explicitly synthetic
and covers cases not represented in this pinned release.

`base.json.gz` contains no extensions. `windows.json.gz` contains exactly `win`,
UID 2, version 1.6.0, with `platform_extension?: false`. The Windows extension
also patches base objects without marking each added attribute as extension
owned. The complete extension set therefore must match the caller's selection.
The compiler's default invocation includes platform extensions; it is not the
base catalog invocation used here.

## Reproduction

Use Python 3.14.1 for the official compiler preparation (reported compiler
version `0.0.0-dev`). This preparation-only requirement does not change the
toolkit's Go 1.22+ or Python 3.11+ runtime floors. Pin both sources exactly:

```sh
git clone https://github.com/ocsf/ocsf-schema.git
git -C ocsf-schema checkout --detach d0cd8a0fef198bf93086044e33fda6d80c74b9ef
git clone https://github.com/ocsf/ocsf-schema-compiler.git
git -C ocsf-schema-compiler checkout --detach d6b0b781d51a6b9682ea99396636ae01437b41b4
SCHEMA="$(pwd)/ocsf-schema"
cd ocsf-schema-compiler/src
python3.14 -m ocsf_schema_compiler "$SCHEMA" --ignore-platform-extensions > base-catalog.json
python3.14 -m ocsf_schema_compiler "$SCHEMA" --ignore-platform-extensions --extensions-path "$SCHEMA/extensions/windows" > windows-catalog.json
```

The schema revision is release `v1.6.0`. `provenance.json` records full source
revisions, official repository URLs, source archive hashes, exact portable
compiler commands, raw sizes and hashes, compressed hashes, extension identities,
and notice hashes. Preparation used the official commit archives whose hashes
are recorded there; equivalent commit-pinned checkouts are shown above.

From the directory containing the raw outputs, verify before compression:

```python
from pathlib import Path
import gzip
import hashlib

expected = {
    "base": "9b609f8fb670772f04191c1c276b46d34d6e9110d2417c71fa89c4f54c585137",
    "windows": "19af77ce259f3ff57debc33e52da51b8220a1af59399d06553f29629678595e9",
}
for name, digest in expected.items():
    raw = Path(name + "-catalog.json").read_bytes()
    assert hashlib.sha256(raw).hexdigest() == digest
    with Path(name + ".json.gz").open("wb") as output:
        with gzip.GzipFile(filename="", mode="wb", fileobj=output, mtime=0) as compressed:
            compressed.write(raw)
```

This compression code is Python 3.11 compatible. The four `SCHEMA-*` and
`COMPILER-*` license/notice files preserve upstream bytes, including the
Broadcom/Symantec ICD attribution. The tests establish source-structure
projection behavior; public toolkit query/schema report validation is a
separate implementation task.
