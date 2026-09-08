# Milestone 4 research notes

Research date: 2026-09-07. These are development references, never runtime dependencies. The design controller approved consuming the official OCSF compiled format directly, with exact compiled-extension-set matching. No production changes or fixture-generation claims are made here.

## Practical OCSF input provenance

The [official OCSF compiler](https://github.com/ocsf/ocsf-schema-compiler) emits one local JSON document. Its documented normal output is `compile_version: 1`; it resolves source-schema inheritance and includes before consumers see class attributes. The [format documentation](https://github.com/ocsf/ocsf-schema-compiler/blob/main/docs/format.md) describes `version`, category/class/object/profile/extension tables, scoped extension keys, and merged attributes. An extension can patch a base definition without giving that definition a simple extension-origin tag. Therefore dropping extension-tagged entries from an already compiled document cannot recover the corresponding base schema.

The [official server preparation instructions](https://github.com/ocsf/ocsf-server#compile-an-ocsf-schema) and [compiler CLI source](https://github.com/ocsf/ocsf-schema-compiler/blob/main/src/ocsf_schema_compiler/__main__.py) establish the following preparation commands. They are documentation-verified commands, not commands executed during this design task:

```sh
# Base only; the compiler otherwise includes platform extensions by default.
ocsf-schema-compiler /path/to/pinned/ocsf-schema --ignore-platform-extensions > base-catalog.json

# The pinned schema with all its default platform extensions.
ocsf-schema-compiler /path/to/pinned/ocsf-schema > platform-catalog.json

# Explicit local extension input with default platform loading disabled.
ocsf-schema-compiler /path/to/pinned/ocsf-schema --ignore-platform-extensions --extensions-path /path/to/pinned/extension > selected-catalog.json
```

`--extensions-path` is repeatable. Consumers must inspect the resulting `extensions` table and select that exact set, rather than infer included names from directory names. The official source tree at [OCSF v1.6.0 extensions](https://github.com/ocsf/ocsf-schema/tree/v1.6.0/extensions) includes `linux` and `windows` directories. The current compiler instructions require Python 3.14+, which is a preparation dependency only; SPL Toolkit keeps its Python 3.11+ runtime baseline. No compiler, Git clone, download, account, or network connection belongs in validation.

Acceptance work must generate or obtain at least one genuine complete compiled catalog from a pinned official release, plus a selected-extension variant, and record source repository/revision, compiler version/revision, commands, output hash, catalog version, extension identities and versions, and upstream license/notice. Synthetic edge fixtures supplement that evidence. No generated-catalog hash was established in this design task; that belongs in the actual acceptance-fixture provenance. Initial shell network access could not resolve GitHub and Python's TLS trust configuration failed; the system curl subsequently read the official source with normal certificate verification after approved network escalation. Web research also successfully read the cited primary pages.

## OCSF semantics relevant to query references

The [official overview](https://github.com/ocsf/ocsf-docs/blob/main/overview/understanding-ocsf.md#attribute-requirement-flags) distinguishes required, recommended, and optional attributes in their class/object context. Recommended does not guarantee presence. Category membership comes from class metadata. Category-only validation therefore compares the selected member classes; it must never infer a class from the query's values.

The [official server profile filter](https://github.com/ocsf/ocsf-server/blob/main/lib/schema/utils.ex) retains an attribute if it has no profile dependency or if its profiles intersect the selected set. The compiler's documented null/missing/empty profile metadata has no profile dependency; honor that explicit format contract. Resolve exact scoped keys and retain query-independent selection evidence.

The [pinned OCSF dictionary](https://github.com/ocsf/ocsf-schema/blob/v1.6.0/dictionary.json) defines `json_t` as arbitrary JSON, including objects and arrays. Descendants under a generic JSON field cannot be rejected as if it were a closed typed object. The compiler's object references (`object_type`) must resolve through its local object table. Reading array contents requires an explicit toolkit path convention; no such inference is approved in this milestone draft.

## Verified compiler profile limitations

The GitHub API returned compiler commit `d6b0b781d51a6b9682ea99396636ae01437b41b4`. The read compiler source's Git blob hash was verified as `4dd6486c36c71536b93d86cabf423274edbbe73f`, matching the API source listing. Inspect these pinned locations:

- [Profile inclusion labeling, lines 907-960](https://github.com/ocsf/ocsf-schema-compiler/blob/d6b0b781d51a6b9682ea99396636ae01437b41b4/src/ocsf_schema_compiler/compiler.py#L907).
- [Class inheritance expansion, line 1156](https://github.com/ocsf/ocsf-schema-compiler/blob/d6b0b781d51a6b9682ea99396636ae01437b41b4/src/ocsf_schema_compiler/compiler.py#L1156), and [object expansion, line 1311](https://github.com/ocsf/ocsf-schema-compiler/blob/d6b0b781d51a6b9682ea99396636ae01437b41b4/src/ocsf_schema_compiler/compiler.py#L1311). The reviewed compile pipeline has no profile counterpart.
- [Profile union and strongest-requirement merge, lines 1734-1764](https://github.com/ocsf/ocsf-schema-compiler/blob/d6b0b781d51a6b9682ea99396636ae01437b41b4/src/ocsf_schema_compiler/compiler.py#L1734).

Inference from this code: the normal artifact does not establish profile-inheritance expansion or a requiredness decision for every strict subset of merged profiles. The controller approved indeterminate results where those gaps affect a queried path, retaining conclusive unconditional declarations and avoiding false missing results. Do not implement unverified selector closure or assume the merged strongest requirement came from a selected profile.

## JSON Schema authority

- [Draft 2020-12 Core](https://json-schema.org/draft/2020-12/json-schema-core): resource identity, local resolution, anchors, `$ref` siblings, conjunction/disjunction/exclusivity, and keyword-scoped object applicators.
- [Draft 2020-12 Validation](https://json-schema.org/draft/2020-12/json-schema-validation): required-property meaning and field-affecting dependency/cardinality keywords.
- [Official object guidance](https://json-schema.org/understanding-json-schema/reference/object): optional properties, open objects, patterns, and the interaction of closed objects with composition.
- [Official composition guidance](https://json-schema.org/understanding-json-schema/reference/combining): `allOf` is conjunction, `anyOf` is alternatives, and `oneOf` requires exactly one branch; an `allOf` wrapper does not provide inheritance-style extension of a closed object.
- [Official regex guidance](https://json-schema.org/understanding-json-schema/reference/regular_expressions): ECMAScript regular expressions and the practical portable subset. An implementation may support a documented conservative subset, but may not silently substitute Go-specific matching semantics.
- [Official static-analysis discussion](https://json-schema.org/blog/posts/schema-static-analysis): schemas can describe unenumerated locations. This supports a partial source universe in the shared kernel rather than pretending every schema is a finite field list.

The proposed toolkit contract is static field projection, not instance validation or a general schema-satisfiability checker. Unsupported keywords must be classified by whether they affect the requested path; a warning must not be silently discarded merely because JSON parsing succeeded.

## Executed official-catalog preflight

The subsequent authorized preflight completed on 2026-09-07. This section supersedes the earlier design-stage statement that no catalog hashes had yet been established. Everything generated remains under `/private/tmp/spl-toolkit-ocsf-preflight`; no catalogs were copied into runtime, test, or package source, and no implementation plan was started.

### Pins, runtime, and reproducible preparation

- Official schema release: `v1.6.0`, resolved through the official GitHub API to commit `d0cd8a0fef198bf93086044e33fda6d80c74b9ef`. The extracted `version.json` confirms `1.6.0`.
- Official compiler: approved commit `d6b0b781d51a6b9682ea99396636ae01437b41b4`. The source checkout reports `0.0.0-dev`; upstream replaces that version string during publishing, so the immutable commit is the meaningful compiler pin here. This is not represented as a tagged compiler release.
- Runtime: existing `/opt/homebrew/bin/python3.14`, Python `3.14.1`. No installation, global change, or virtual environment was needed. Direct module execution from the compiler's `src` directory uses its standard-library-only runtime dependencies.
- Sources were downloaded using normal certificate verification from official GitHub codeload URLs at the exact commits, then extracted locally with Python's `tarfile` data filter. Compilation ran successfully in the ordinary restricted execution environment after the downloads.
- Full exact commands, source archive hashes, runtime details, catalog tables, and all license/notice hashes are saved in `/private/tmp/spl-toolkit-ocsf-preflight/provenance.json`. `/private/tmp/spl-toolkit-ocsf-preflight/inspect_catalogs.py` regenerates that inventory and checks the observed format/link invariants.

Source download commands:

```sh
curl --fail --silent --show-error --location --output /private/tmp/spl-toolkit-ocsf-preflight/compiler.tar.gz https://codeload.github.com/ocsf/ocsf-schema-compiler/tar.gz/d6b0b781d51a6b9682ea99396636ae01437b41b4
curl --fail --silent --show-error --location --output /private/tmp/spl-toolkit-ocsf-preflight/schema.tar.gz https://codeload.github.com/ocsf/ocsf-schema/tar.gz/d0cd8a0fef198bf93086044e33fda6d80c74b9ef
```

The following commands were run from `/private/tmp/spl-toolkit-ocsf-preflight/ocsf-schema-compiler-d6b0b781d51a6b9682ea99396636ae01437b41b4/src`. Each command's stderr was captured separately in the corresponding `*-compile.log`:

```sh
PYTHONDONTWRITEBYTECODE=1 /opt/homebrew/bin/python3.14 -m ocsf_schema_compiler /private/tmp/spl-toolkit-ocsf-preflight/ocsf-schema-d0cd8a0fef198bf93086044e33fda6d80c74b9ef --ignore-platform-extensions > /private/tmp/spl-toolkit-ocsf-preflight/base-catalog.json
PYTHONDONTWRITEBYTECODE=1 /opt/homebrew/bin/python3.14 -m ocsf_schema_compiler /private/tmp/spl-toolkit-ocsf-preflight/ocsf-schema-d0cd8a0fef198bf93086044e33fda6d80c74b9ef --ignore-platform-extensions --extensions-path /private/tmp/spl-toolkit-ocsf-preflight/ocsf-schema-d0cd8a0fef198bf93086044e33fda6d80c74b9ef/extensions/windows > /private/tmp/spl-toolkit-ocsf-preflight/windows-catalog.json
PYTHONDONTWRITEBYTECODE=1 /opt/homebrew/bin/python3.14 -m ocsf_schema_compiler /private/tmp/spl-toolkit-ocsf-preflight/ocsf-schema-d0cd8a0fef198bf93086044e33fda6d80c74b9ef > /private/tmp/spl-toolkit-ocsf-preflight/platform-catalog.json
```

All three succeeded with normal `compile_version: 1`, schema version `1.6.0`. The base and explicit-Windows runs were repeated and produced byte-identical output. There were no dangling compiled object links or dangling category links for concrete classes, and every class/object attribute had a recognized required/recommended/optional requirement. The retained `base_event` sentinel has UID/category UID 0 and category `other`, which is intentionally absent from the ordinary category table; skip that sentinel before validating concrete-class links.

| Temporary catalog | Bytes | SHA-256 | Extension set | Classes / objects / profiles |
| --- | ---: | --- | --- | --- |
| `base-catalog.json` | 3,122,505 | `9b609f8fb670772f04191c1c276b46d34d6e9110d2417c71fa89c4f54c585137` | `[]` | 75 / 163 / 11 |
| `windows-catalog.json` | 3,344,790 | `19af77ce259f3ff57debc33e52da51b8220a1af59399d06553f29629678595e9` | `[win]` | 82 / 167 / 11 |
| `platform-catalog.json` | 3,348,659 | `db9ff7b6c2e8903050852c1fbe94885e6d4f7877fd8dfb4ef8cfad2a6bd0b858` | `[linux, win]` | 82 / 167 / 12 |

Class counts include `base_event`; concrete counts are 74, 81, and 81. The selected Windows source comes from the same pinned official schema commit, specifically `extensions/windows/extension.json`: name `win`, UID 2, version `1.6.0`. Passing it with `--extensions-path` records `platform_extension?: false`, as expected for that explicit loading route. The default-platform output instead records `linux` (UID 1) and `win` (UID 2), both version `1.6.0` and both `platform_extension?: true`. Therefore consumers must use the catalog's actual extension table, not assume that a directory name or source location identifies its compiled metadata.

Both repositories are Apache-2.0 licensed. Their original `LICENSE` and `NOTICE` files remain in the extracted source roots. The schema notice attributes OCSF/LF Projects and the included ICD Schema developed by Symantec/Broadcom; preserve these notices with any eventual vendored fixtures. The compiler notice attributes OCSF/LF Projects. Exact file hashes are in `provenance.json`.

### Real field and selection examples

These are observations of the official compiled outputs, not executed SPL Toolkit validation outcomes:

- `authentication`, UID 3002, category `iam`: `time` and `user` are required; `actor` is recommended with explicit null profile dependency; `unmapped` is optional; `observables` is an array. `cloud` is required only under its `cloud` profile dependency; `time_dt` is optional under `datetime`.
- `file_activity`, UID 1001, category `system`: `file` is required, and its resolved `file.name` child is required. `file.xattributes` is an optional generic object.
- The six concrete IAM members are `account_change` (3001), `authentication` (3002), `authorize_session` (3003), `entity_management` (3004), `user_access` (3005), and `group_management` (3006). `group` is declared only in `authorize_session` and `group_management`, giving a real category-conditional acceptance example. `time` is required across these members.
- The Windows catalog contains `win/registry_key_activity`, UID 201001, category `system`, plus six other scoped Windows classes. Exact key and UID selection can be exercised without class-name guessing.
- Compiled Windows patches add `reg_key`, `reg_value`, and/or `win_service` to existing `evidences`, `query_evidence`, and `startup_item` objects while those objects have no `extension` origin. This directly confirms that filtering extension-origin tags cannot undo the compiled extension set.
- Base and explicit-Windows normal profiles contain no `extends` definitions and no multi-profile attribute lists. They exercise ordinary complete profile cases; inheritance and strict-subset requirement limitations still require separate carefully labeled edge fixtures.

### Source-grounded generic-object correction

Real catalogs exposed a second open boundary beyond `json_t`. The controller approved this narrow correction and owns its spec edit:

- [Pinned `objects/object.json`](https://github.com/ocsf/ocsf-schema/blob/d0cd8a0fef198bf93086044e33fda6d80c74b9ef/objects/object.json) explicitly defines the base `object` as usable for generic data outside named schema objects. Compiled `/objects/object` preserves that definition.
- [Pinned `dictionary.json`](https://github.com/ocsf/ocsf-schema/blob/d0cd8a0fef198bf93086044e33fda6d80c74b9ef/dictionary.json#L6319), pointer `/attributes/unmapped`, declares raw type `object`. Compiled `/classes/authentication/attributes/unmapped` resolves to `type: object_t`, `object_type: object`, requirement optional. `unmapped.vendor_field` must be admitted without a named declaration.
- The same dictionary's `/attributes/xattributes` (line 6555), [file usage](https://github.com/ocsf/ocsf-schema/blob/d0cd8a0fef198bf93086044e33fda6d80c74b9ef/objects/file.json#L193), and [process usage](https://github.com/ocsf/ocsf-schema/blob/d0cd8a0fef198bf93086044e33fda6d80c74b9ef/objects/process.json#L65) supply a second practical case. Compiled `/objects/file/attributes/xattributes` and `/objects/process/attributes/xattributes` resolve the same generic base key. Their arbitrary user/application attribute descendants need permitted-unspecified outcomes.
- Apply this openness only when `object_t` resolves exactly to the base generic key `object`, alongside the already approved `json_t` rule. An empty attribute map alone is not evidence of openness. A typed object merely extending `object` is not automatically open: real `http_cookie` and `request` objects have 11 and 4 declared attributes respectively and remain governed by their declared structure.

The standalone `json_t` path is also present: compiled `/objects/request/attributes/data`. Array traversal remains a separate incomplete case even though direct array declarations are checkable.

### Required transport follow-up and completion limits

The genuine catalogs exceed the current REST decoder ceiling at `pkg/api/models.go:117`, which is `1<<20` bytes. The controller designated a deliberate bounded catalog-capable request limit as required M4 implementation-plan work, including real base/Windows payload acceptance and oversized-request rejection. No production limit was changed here. Synthetic-only REST fixtures would miss this real interoperability requirement.

Catalog preparation has no blocker. This preflight proves the official preparation path and inspects real artifact structure; it does not establish that any not-yet-built M4 runtime accepts the catalogs. The verified-M3 implementation plan retains final fixture selection, copying, packaging, and cross-surface validation ownership. Temporary source trees, outputs, logs, repeat outputs, inspection script, and provenance inventory are ready at the path above.

## Implementation acceptance follow-through — 2026-09-08

The pinned preparation research was exercised by the production fixtures and accepted at `92879052b0454db12aead045188ce4b15fa9d83a`. Genuine base/Windows catalogs worked through Go, CLI, REST, source native Python and both installed package paths. All seven task reviews, broad review and root independent Go/all-surface/Python3.14 acceptance passed. This closes the earlier preparation-only handoff; no catalog/compiler download or regeneration was repeated at runtime. Exact execution/provenance and local platform qualifications are retained in `milestone-log.md` and `docs/evidence/milestone-4/closure.json`.
