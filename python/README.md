# SPL Toolkit Python bindings

This package provides Python 3.11+ bindings for SPL Toolkit 0.1.1's offline mapping, discovery, structured analysis, caller-supplied knowledge-object closure, environment artifact validation, safe rewriting, field-list validation, and JSON Schema/OCSF field declaration APIs. Wheels include the native Go library and do not require Go at installation or runtime.

```python
from spl_toolkit import SPLMapper

with SPLMapper() as mapper:
    mapper.load_mappings([{"source": "src_ip", "target": "source_ip"}])
    print(mapper.map_query("search src_ip=1"))
```

Use `SPLMapper` as a context manager or call `close()`. Operations after close raise `MapperNotFoundError`. Loading invalid configuration raises `ConfigurationError`, and rejected queries raise `ParseError`. The package verifies that the native and Python package versions match when the native library is opened.

Discovery returns `data_models`, `datasets`, `lookups`, `macros`, `sources`, `source_types`, and flat `input_fields`. Flat fields are not lineage. When unsupported macro syntax prevents normal parsing, discovery may return recovered macro names with the other categories empty.

Source distributions require a local Go toolchain and C compiler. Development installs use built wheels; rebuild and reinstall the wheel after changing the Python wrapper or native Go code.

## Corpus, graph, impact and document tools

`SPLMapper.scan_corpus(request)`, `export_graph(request)` and `export_sarif(request)`
accept `{"schema_version": 1, "documents": [{"id": "q", "document": {"text": "table host"}}]}`
with an optional canonical `validation_target`. They return complete dictionaries.
`impact_schema` accepts that document list plus `before_target`/`after_target`;
`impact_mapping` accepts `before_rules`/`after_rules` and optional validation targets.
`document_view` accepts a QueryDocument dictionary and returns a detached evidence
snapshot. All calls use the owned native boundary and existing close/concurrency
guards; no CLI subprocess or Python semantic implementation is involved.

Wheels contain versioned machine schemas and the official SARIF schema with
provenance/notices under `spl_toolkit/contracts`. Python file users can read UTF-8
sources into QueryDocuments explicitly. The native API accepts inline snapshots
and does not implicitly read caller filenames. See the repository's
`examples/tooling/native.py` and `docs/tooling.md` for runnable use.

## Structured analysis

```python
from spl_toolkit import SPLMapper

with SPLMapper() as mapper:
    report = mapper.analyze_query(
        "search src=1 | eval a=src, b=a+1 | table b",
        language="spl", profile="splunkd", version="current",
        source_id="example.spl",
    )
    print(report["status"])
    capabilities = mapper.capabilities()
```

Both methods return dictionaries with integer `schema_version: 1`. The keyword-only defaults are `language='spl'`, `profile='splunkd'`, `version='current'`, and `source_id=''`. Empty compatibility options normalize to these defaults; unsupported options raise `SPLMapperError`. Analysis returns `valid`, `invalid`, or `incomplete` reports rather than raising `ParseError` for query diagnostics. Structural validity does not validate an external field catalog or schema. Inspect `coverage.syntax_complete`, `coverage.semantic_complete`, and diagnostics before treating findings as conclusive.

Reports preserve source text and source ID and expose ordered stages/scopes, references, lineage, dependencies, and diagnostics; collections are arrays even when empty. Locations are half-open ranges with zero-based UTF-8 byte offsets and one-based Unicode code-point line/column coordinates. Tabs count as one column; CRLF is one newline. Slice `query.encode("utf-8")[start_offset:end_offset]` for a byte range. Invalid UTF-8 and lone JSON/Python surrogate values raise errors before replacement; paired non-BMP escapes are preserved.

Capabilities expose `records`, `summary`, and `evidence` in addition to the legacy command/function projections. Records keep syntax, semantics, requirements, linting, and safe rewriting separate, using supported, partial, unsupported, not applicable, and unassessed states. For each dimension, applicable equals supported plus partial plus unsupported plus unassessed; records equal applicable plus not applicable; covered equals supported. The manifest has no percentage or composite score.

The SPL snapshot contains 106 records and 107 evidence cases; SPL2 has 142 records and 164 evidence cases. SPL syntax is 77 supported, 1 unsupported, and 28 unassessed; semantics is 68 supported, 10 unsupported, and 28 unassessed; requirements are 19 supported, 6 unsupported, and 81 unassessed; linting has 106 unassessed records; and safe rewriting is 17 supported, 2 unsupported, and 87 unassessed. SPL2 syntax is 90 supported, 19 unsupported, and 33 unassessed; semantics is 76 supported, 39 unsupported, and 27 unassessed; requirements are 46 supported, 1 partial, 7 unsupported, and 88 unassessed; linting has 142 unassessed; and safe rewriting has 15 supported, 3 unsupported, and 124 unassessed. Evidence IDs resolve to typed local observations and provenance. A broadened form receives a new ID unless a reviewed scope correction changes the original boundary. `grammar_registered` records parser registration only. It does not grant syntax coverage, and ordinary analysis diagnostics do not grant lint coverage.

Source Go tests may report toolkit version `dev`. Tagged native libraries and Python packages report the exact `VERSION` and retain the same semantic capability revision because `toolkit_version` is the only manifest field excluded from that revision. The revision includes selectors, documentation snapshot, legacy projections, rewrite, records, summary, and evidence. The SPL revision is `sha256:5b7c15002c426b163a5488d18ae0fb83c68809a194a15f9a6584ea95cbb3a09b`; the SPL2 revision is `sha256:216bb4be25e623d48098144be8bf006b80b4c1b4b6eeec05a2b7b4c2c3559842`.

The bounded SPL field-flow semantics support exact `tstats` data-model sources, predicates, registered aggregates, aliases, groups, and supported literal options; exact modeled forms of `fillnull`, `rex`, `spath`, `bin`, `bucket`, `regex`, and `mvexpand`; and the selected evaluation and aggregate function arities. Exact macros produce direct macro requirements and source-located unresolved-expansion gaps. Branch children retain evidence and direct requirements, while merged output fields remain uncertain.

Unknown commands and functions, dynamic operands, wildcard `tstats` groups, `PREFIX(...)`, true result-shape modes, unsupported options, ambiguous or sed-mode `rex`, `spath` auto-extraction, macro expansion, branch merging, and unresolved wildcard membership remain incomplete. Open-input `fields` inclusion also remains incomplete when retained internal membership is unknown. A later stage cannot erase an earlier coverage gap. The local minimized detection corpus improved from 12/1/35 to 14/16/13 for syntax-complete cases, semantic-complete selected target stages, and non-macro gaps; it does not measure whole-query completeness. The toolkit does not execute SPL, evaluate regular expressions, compare result rows, model acceleration, load macro definitions, or certify runtime compatibility. The corpus also does not prove authorization or upstream support.

For `language="spl2"`, selected local views and pure scalar functions bind within one submitted document. Used external imports remain incomplete; unused imports can leave analysis valid while direct requirement coverage is incomplete. Selected `if`, guarded `branch`, `union`, and qualified pipeline `join` forms merge field flow and requirement evidence. Public `field_identity` on exact references, bindings, removals, and field requirement items distinguishes quoted atomic dotted fields from structural paths even when their display names match; proved transitions carry `output_identity`. Genuinely unproved join output remains incomplete. Exact static dataset descriptors have canonical JSON requirement identities. Inspect `report["requirements"]["coverage"]` separately from analysis status. Parser acceptance, semantic completeness, requirement completeness, lint support, safe rewriting, live runtime compatibility, and deployment are separate claims.

Legacy discovery and mapping retain their own contracts. Flat `input_fields` does not encode flow, scope, or completeness and is not guaranteed to match structured classifications. New consumers should use reference roles/bindings and status/coverage. The [API reference](https://github.com/delgado-jacob/spl-toolkit/blob/main/docs/API.md) describes every surface and the supported forms; this package does not perform event instance validation or external SPL2 module execution.

## Query requirements

```python
from spl_toolkit import SPLMapper

with SPLMapper() as mapper:
    requirements = mapper.requirements_query(
        "search index=main host=web | eval label=host | table label",
        language="spl", profile="splunkd", version="current",
        source_id="example.spl",
    )
    assert requirements == mapper.analyze_query(
        "search index=main host=web | eval label=host | table label",
        language="spl", profile="splunkd", version="current",
        source_id="example.spl",
    )["requirements"]
    print(requirements["query_status"], requirements["coverage"]["complete"])
```

`requirements_query(query, *, language='spl', profile='splunkd', version='current', source_id='')` returns the complete canonical `RequirementSet` dictionary. It contains integer `schema_version: 1`, query identity and digest, capability revision, query status, requirement coverage, ordered items, gaps, and diagnostics. Item occurrences retain reference, spelling, binding, stage, scope, and exact source location. Every collection remains a list when empty.

Items group occurrences by kind, normalized identity, consuming role, and resolution. Source-bound consumers and direct knowledge-object references can be required. Indeterminate, wildcard, and dynamic consumers are conditional and produce gaps. Fields created by the query, rename targets, removals, null tests, and query-local unavailable fields are not external obligations. Requirement coverage is independent of query status, so inspect both values.

Validation and rewrite can refine their public analysis against a supplied target, but their embedded requirements remain equal to plain query-only analysis of the same normalized document. The operation does not expand knowledge objects, read environment metadata or event data, assess compatibility, resolve placeholders, generate variants, or execute the query.

`query_digest` is SHA-256 over the exact valid UTF-8 query bytes. It excludes source ID and compatibility selectors and performs no whitespace or line-ending normalization. `capability_revision` identifies the compact normalized capability manifest. Both use `sha256:<64 lowercase hex>`. They identify supplied data and are not authentication, authorization, signatures, environment compatibility, or execution permission.

Canonical analysis admits at most 4,096 lexer work units. A query that would consume unit 4,097 returns a successful dictionary with incomplete status and coverage, one `SPL_ANALYSIS_RESOURCE_LIMIT` diagnostic and gap, the full-text digest, and no partial requirement evidence. Long sparse input remains admitted when it stays within the work budget. Preview, apply, and batch rewrite return an incomplete no-op for a resource-limited original; apply never commits. The [API reference](https://github.com/delgado-jacob/spl-toolkit/blob/main/docs/API.md#canonical-lexer-work-boundary) defines exact ordering, SPL2 closure accounting, messages, half-open ranges, and the fixture-specific response-size checks.

The method uses the owned native call `spl_mapper_requirements_query`. Mapper admission and close guards match the other operations, and the wrapper releases every returned `SPLResult` with `spl_result_free`, including decoding and native error paths. Repeated calls return detached Python values.

## Offline environment validation

`SPLMapper.validate_environment(request)` accepts an inline version 1 request
containing a `snapshot`, a `schema_bundle`, or both. It returns the canonical
Go report as a detached dictionary, including `status`, computed digests when
available, `coverage`, and `diagnostics`. From the repository checkout:

```python
import json
from spl_toolkit import SPLMapper

with open("examples/environment/partial-snapshot.json", encoding="utf-8") as source:
    snapshot = json.load(source)
with open("examples/environment/fields.json", encoding="utf-8") as source:
    schemas = json.load(source)
with SPLMapper() as mapper:
    report = mapper.validate_environment({
        "schema_version": 1, "snapshot": snapshot, "schema_bundle": schemas,
    })
    print(report["status"])  # partial
```

The Python method accepts data, not paths. Invalid artifact content returns an
`invalid` report; serialization errors and native handle errors raise
`SPLMapperError`. The wrapper releases each owned native result, including on
decoding errors. Direct C callers use
`spl_mapper_validate_environment(int mapperID, char* requestJSON)` and free
every non-null `SPLResult*` with `spl_result_free`. Collection completeness and
field source coverage describe producer claims. The operation does not fetch a
live inventory, read event rows, or assess whether a query can execute.

## Knowledge-object closure

`SPLMapper.closure_query(request)` accepts the strict version 1 inline request
and returns the canonical Go closure report as a dictionary. Supply a query
`document`, a `bundle` with `scope_id`, per-kind collection coverage and
definitions, and optional occurrence `bindings`. For example:

```python
with SPLMapper() as mapper:
    report = mapper.closure_query({
        "schema_version": 1,
        "document": {"text": "| lookup users user OUTPUT role", "source_id": "query.spl"},
        "bundle": {"schema_version": 1, "scope_id": "local-example",
                   "collections": [{"kind": "lookup", "coverage": "complete"}],
                   "objects": [{"id": "lookup-users", "kind": "lookup",
                                "name": "users", "source_id": "lookups.conf",
                                "relations": []}]},
        "bindings": [],
    })
    assert report["status"] == "valid"
```

Invalid and incomplete query content returns a report; malformed requests raise
`SPLMapperError`. The wrapper releases the owned native result. Direct C callers
use `spl_mapper_closure_query(int mapperID, char* requestJSON)` and free every
non-null `SPLResult*` with `spl_result_free`. `coverage.complete` is scoped to
the supplied collection promises, traversed definitions, exact resolution, and
bounded macro expansion. It is not a live Splunk inventory or execution check.
See the [closure API
contract](https://github.com/delgado-jacob/spl-toolkit/blob/main/docs/API.md#knowledge-object-closure).

## Safe rewrite

```python
rules = [{"id": "source-user", "kind": "field",
          "source": {"name": "src"}, "target": {"name": "user"}}]
with SPLMapper() as mapper:
    preview = mapper.rewrite("search src=alice", rules)
    applied = mapper.rewrite("search src=alice", rules, mode="apply")
    assert preview["text"] == "search src=alice" and not preview["committed"]
    assert applied["text"] == "search user=alice" and applied["committed"]
```

`rewrite(query, rules, *, mode="preview", validation_target=None, language="spl", profile="splunkd", version="current", source_id="")` returns the Go report dictionary. `rewrite_batch(documents, rules, *, mode="preview", validation_target=None)` preserves document order. Explicit aliases stay fixed; source-bound uses and provable implicit consumers change together. A safe group can commit beside a refused group with `status="incomplete"`; invalid or incomplete destination validation prevents all commits. The candidate and audit remain visible on rollback. Input errors raise `SPLMapperError` and publish no partial batch.

Rules are independent of legacy configuration/precedence and use only original canonical facts. The [rewrite guide](https://github.com/delgado-jacob/spl-toolkit/blob/main/docs/rewrite.md) defines identities, conditions, targets and limitations. `examples/basic_usage.py` runs preview, apply, validation rollback and ordered valid/incomplete/invalid batches with the real native library. Rebuild and reinstall after native API changes; older same-version libraries need not contain these additive exports.

## Field-list validation

SPL2 uses explicit keyword-only `language="spl2"` on `analyze_query`,
`validate_fields`, `validate_schema` and `capabilities`. Batch documents carry
their own dialects and preserve order. For example:

```python
with SPLMapper() as mapper:
    report = mapper.analyze_query("SELECT host FROM main WHERE bytes>0", language="spl2")
    manifest = mapper.capabilities(language="spl2")
```

SPL, `splunkd`, and `current` remain the defaults. `current` is the build's bounded
capability snapshot. SQL stages retain lexical order, while positions and lineage
phases describe evaluation. Direct `null_test` inspections retain their source
evidence without an existence outcome; ordinary reads remain obligations.
Conditional presence can be fully modeled while later consumers are indeterminate.
Named arguments, dynamic paths and held forms remain incomplete. Legacy mapping,
context mapping, discovery and input-field methods reject explicit SPL2 selectors
with canonical-operation guidance. See the
[SPL2 contract](https://github.com/delgado-jacob/spl-toolkit/blob/main/docs/spl2.md).

```python
from spl_toolkit import SPLMapper

with SPLMapper() as mapper:
    report = mapper.validate_fields(
        "eval label=host | table label", ["host"],
        language="spl", profile="splunkd", version="current",
        source_id="example.spl",
    )
    assert report["status"] == "valid"
    batch = mapper.validate_fields_batch(
        [{"text": "table host", "source_id": "first.spl"},
         {"text": "table missing", "source_id": "second.spl"}],
        {"fields": ["host"], "optional_fields": ["user"],
         "identity": "local-catalog", "version": "1"},
    )
    assert batch["status"] == "invalid"
```

`validate_fields(query, catalog, *, language='spl', profile='splunkd', version='current', source_id='')` returns a dictionary with `schema_version`, `target`, `analysis`, `status`, `coverage`, `outcomes`, and `diagnostics`. `validate_fields_batch(documents, catalog)` takes a nonempty list of document dictionaries with required string `text` and optional `language`, `profile`, `version`, and `source_id`. It returns `schema_version`, `status`, and ordered `reports`; batch status uses invalid, then incomplete, then valid precedence. Schema versions are integer `1` and collections are arrays, including when empty.

A catalog is a string list or an object with required `fields` and optional `optional_fields`, `identity`, and `version`. An empty field list is valid. Field names are case-sensitive and nested names match exactly: a parent or leaf declaration does not imply the other. Empty names, duplicate or overlapping ordinary/optional names, unknown properties, malformed Unicode, and unsupported document options raise `SPLMapperError`. Catalog metadata is preserved; it never triggers a file read or network request. Python only serializes the request and returns the Go report.

Validation follows fields through the query, including created and removed fields and supported wildcards. Reference outcomes are `matching`, `missing`, `unavailable`, `optional_equivalent`, or `indeterminate`; individual matches are `matching` or `optional_equivalent`. Invalid or incomplete query content is a report, not an API error. Inspect `coverage.syntax_complete`, `coverage.semantic_complete`, `coverage.schema_complete`, reasons, and diagnostics. Unsupported semantics remain incomplete and cannot be erased by a later stage. The embedded analysis preserves source metadata and Unicode locations; plain `analyze_query` continues to analyze without a catalog.

Both validation operations use the same context manager and close rules as other mapper methods. Native results are released after decoding, including failures. For direct C callers, `spl_mapper_validate_fields(int mapperID, char* requestJSON)` takes `{"document": {...}, "catalog": ...}` and `spl_mapper_validate_fields_batch` takes `{"documents": [...], "catalog": ...}`. Each returns an owned `SPLResult*` to release with `spl_result_free`, including request/handle errors. Null requests are errors.

Optionality is declaration-only: `optional_equivalent` is valid but says nothing about whether an event contains the field. It cannot repair a structurally removed or conditional field. Each outcome retains every concrete match with its `name`, `binding` (`source` or `derived`), and individual `outcome`; mixed ordinary/optional matches must not be collapsed. A conclusively empty wildcard inclusion is missing, while an empty exclusion is harmless. Unsupported wildcard command forms, dynamic references, and conditional/branch effects remain incomplete. Plain `analyze_query` remains catalog-free.

## JSON Schema and OCSF fields

```python
from spl_toolkit import SPLMapper

target = {"kind": "json_schema", "schema": {"type": "object",
          "properties": {"actor": {"type": "object", "properties": {"name": {"type": "string"}},
                                   "required": ["name"], "additionalProperties": False}},
          "required": ["actor"], "additionalProperties": False}}
with SPLMapper() as mapper:
    report = mapper.validate_schema("table actor.name", target, source_id="first.spl")
    assert report["outcomes"][0]["outcome"] == "required"
    batch = mapper.validate_schema_batch(
        [{"text": "table actor.name", "source_id": "first.spl"},
         {"text": "table absent", "source_id": "second.spl"}], target)
    assert batch["status"] == "invalid"
```

`validate_schema` uses the same keyword-only document defaults as analysis. `validate_schema_batch` accepts a nonempty list of document dictionaries with their own options. Both return full dictionaries with report format integer `1`, original documents, target limitations, located evidence and separate syntax/semantic/schema coverage. Status precedence is invalid, incomplete, valid. Invalid target/document inputs raise `SPLMapperError`; content diagnostics stay in reports. Requests are copied, results are independently owned, and native results are freed on success or error.

Supply JSON Schema object/boolean `schema`, optional `base_uri` and URI-keyed inline `resources`. Draft 2020-12 is the default. For OCSF use `{"kind":"ocsf","catalog":catalog,"selection":{"version":"1.6.0","class":"authentication","profiles":[],"extensions":[]}}`, loading your prepared catalog explicitly with `json.loads(Path("base-catalog.json").read_text())`. The complete compiled extension set must match selection; Windows requires `["win"]`. Choose `category: "iam"` instead of `class` for category evidence, or select `profiles: ["cloud", "datetime"]` explicitly. Preparation uses the official compiler outside the toolkit runtime. No URI is retrieved and no process-global catalog is consulted.

Nested optional ancestors, category/branch-dependent membership, open wildcard sets, unsupported patterns/keywords, array descendants and missing profile provenance remain qualified. A declared array is supported, but descent is incomplete. Requiredness is a schema statement, not event presence. This does not validate JSON event instances or expression types. See the [full schema contract and pinned compiler preparation](https://github.com/delgado-jacob/spl-toolkit/blob/main/docs/API.md#json-schema-and-ocsf-field-validation).

## Offline compatibility assessment

```python
import json
from pathlib import Path
from spl_toolkit import SPLMapper

request = json.loads(Path("compatibility-request.json").read_text())
with SPLMapper() as mapper:
    report = mapper.check_compatibility(request)
    print(report["outcome"], report["correlation"]["outcome"])
```

Generate the complete envelope with the
[canonical recipe](https://github.com/delgado-jacob/spl-toolkit/blob/main/docs/API.md#offline-compatibility-assessment).
`check_compatibility(request)` uses native `spl_mapper_check_compatibility`; native
results are freed after decoding on success or failure under the existing mapper
close/admission lifecycle. Content outcomes return dictionaries; malformed/stale
requirements, duplicate/unknown bindings and mismatched schema pairing raise
request errors. Required current input arrays, ownership and occurrence evidence
must come from reanalysis. Format integers remain 1.

Bind each discovered placeholder explicitly and isolate schema evidence by proved
input owner. Exact positives survive partial capture; covered negatives require
complete relevant evidence. Independent connectedness does not alter the core
compatibility outcome. This offline operation does not execute SPL or prove live
runtime compatibility, authorization or deployment.

### Named-input resolution

`SPLMapper.resolve(request)` returns all canonical verified/failed/incomplete variants as a detached dictionary; request errors raise `SPLMapperError`. Pass only verified variants’ `resolved_query` downstream. See the [complete Python example and names-only input contract](../docs/resolution.md), including separate offline environment/schema bindings and native result ownership.

## Offline detection workflows

`SPLMapper.workflow_assess(request)`, `workflow_compare(request)`, `workflow_evidence(request)`, and `workflow_recheck(request)` delegate to the native workflow owner. Assessment text format returns a string; canonical JSON formats and other operations return dictionaries. Successful evidence projection preserves `source_ci_exit_code` without treating it as a Python error. Run `python3 examples/workflow/native.py` from the repository root with the installed package. See the [workflow guide](../docs/workflow.md) for runnable examples, CI status and source disclosure controls.
