---
title: "Safe resolution"
layout: page
---

# Safe resolution

Resolution substitutes explicitly selected names, proves correspondence to the original query roles, and assesses each candidate against separately supplied offline environment and schema evidence. It does not discover objects, consult a registry, retrieve remote data, or execute queries.

The [complete request](../examples/resolution/request.json) selects `$events` from `events_good` and `events_missing`. The first object's complete schema contains required `id`; the second complete schema contains `other` and lacks `id`. Original role IDs were obtained from canonical `analysis.Analyze` on the exact submitted document, rather than invented opaque identifiers. The resulting report contains two variants: one verified, one failed, zero incomplete.

## Input and publication contract

`schema_version` is 1. `document`, `resolutions`, and `compatibility` are required and non-null. Supply `resolutions: []` for a no-op assessment, which still generates one variant. Resolution entries contain `placeholder`, `kind`, and a nonempty ordered `values` array of unique nonblank names. The supported kinds are `index`, `source`, `sourcetype`, `lookup`, `dataset`, and `data_model`; support depends on a proved typed slot, not grammar acceptance alone. Values are names only, never inline schemas, object handles, conditions, or executable fragments. Identical marker/kind occurrences are substituted simultaneously across supported root query sites. Multiple entries form a Cartesian product in request order, with the last entry varying fastest.

The default `max_variants` is 100. An explicit override must be a positive uint64. The exact product is reported as decimal-string `total_combinations`; exceeding the limit rejects the entire request before generating any variants. `ordinal` is one-based uint64; counts are nonnegative uint64. Invalid keys, unsupported sites, and malformed requests produce structured request errors. Go admission checks marker/site membership and original-role binding membership; JSON Schema alone cannot establish those relationships. Unknown members, duplicate JSON keys, null optional inputs, malformed Unicode, and fractional numeric encodings are rejected by the strict decoder.

`compatibility` contains a captured Snapshot v1/v2, optional SchemaBundle v1, explicit `query_scope`, required `input_bindings`, and optional `dependency_bindings`. Each input binding selects an `original_input_id`, optionally one `resolved_value`, and an explicit `object_id`, `expected` identity, and optional `schema_id`. Candidate values and compatibility evidence are separate inputs. `input_bindings: []` is valid when the existing evaluator can assess a unique explicit identity directly from captured objects. Field requirements still need a schema selected for that original role; an unbound role cannot borrow another role’s schema. Equal resolved names never erase original roles or let one role borrow another role's schema. Required and conditional requirements remain distinct. `dependency_bindings` retain submitted selection; report `effective_dependency_bindings`, when present, records selection used for effective closure assessment.

Publication is per variant. Outcomes are `verified`, `failed`, or `incomplete`. Only a verified variant contains `resolved_query`; downstream consumers must read that member from verified variants. `candidate_text` remains diagnostic evidence even when the variant fails. Every variant retains selection, changes, proof, diagnostics, and provenance, with optional candidate analysis and compatibility evidence. Required arrays are emitted as arrays even when empty; omitted optional arrays must never be null.

Proof roles preserve original and candidate input identities, occurrences, and requirements. Serialized proof evidence cannot create trusted authority: the Go owner verifies its session-bound rendering before compatibility assessment. Digests identify exact query text, normalized resolution/assessment inputs, capability revision, and detached artifacts; they are provenance, not signatures.

Resolution edits root query sites only. Captured definitions are immutable; markers inside definitions remain subject to existing closure limitations. Classic SPL implicit source ownership and ambiguous/dynamic Dataset sites cannot gain proof merely from replacement or schema evidence. Grammar registration, proved rendering, and source compatibility assessment are separate evidence boundaries. Inspect the existing [capability manifest CLI](cli.md) and each report's coverage rather than assuming every command supports rewriting.

## CLI

See [canonical CLI usage](cli.md).

```bash
spl-toolkit resolve --request examples/resolution/request.json --format json
```

This mixed example intentionally exits 1 after writing the full report. Exit 0 means every variant verified; exit 1 means at least one failed; exit 3 means incomplete variants and no failed variants; request/configuration errors exit 2. The command takes one local request path, supports text or JSON output, and `--output` rejects aliases of the request path, including symlink/hardlink aliases. It does not overwrite the input request.

## HTTP

```bash
curl -sS -X POST http://localhost:8080/query/resolve \
  -H 'Content-Type: application/json' \
  --data-binary @examples/resolution/request.json
```

The application/json transport accepts at most 8,388,608 bytes (8 MiB). All content outcomes return HTTP 200 with the complete canonical report. Request/configuration errors return structured 400 details; transport content type/body-limit errors use the existing ErrorResponse. Internal failures return 500. HTTP status alone does not mean a candidate verified.

## Python and native C

```python
import json
from spl_toolkit import SPLMapper

with open("examples/resolution/request.json", encoding="utf-8") as source:
    request = json.load(source)
with SPLMapper() as mapper:
    report = mapper.resolve(request)
    verified_queries = [v["resolved_query"] for v in report["variants"]
                        if v["outcome"] == "verified"]
```

Python returns a detached dictionary for all content outcomes and raises `SPLMapperError` for request/configuration errors. The native `spl_mapper_resolve` delegate returns an owned result; the caller must free it exactly once with `spl_result_free`. Mapper handles keep their existing lifetimes; neither JSON output nor a decoded proof owns a trusted resolution session.

## Prepared Go

```go
request, err := resolution.DecodeRequest(raw)
if err != nil { return err }
prepared, err := resolution.Prepare(request.Compatibility.Snapshot, request.Compatibility.SchemaBundle)
if err != nil { return err }
report, err := prepared.Resolve(resolution.PreparedRequest{
    SchemaVersion: request.SchemaVersion,
    Document: request.Document,
    Resolutions: request.Resolutions,
    MaxVariants: request.MaxVariants,
    Compatibility: compatibility.ResolutionAssessment{
        QueryScope: request.Compatibility.QueryScope,
        InputBindings: request.Compatibility.InputBindings,
        DependencyBindings: request.Compatibility.DependencyBindings,
    },
})
```

Prepared instances retain detached immutable artifact authority; query sessions and binding choices are call-local. `resolution.ResolveJSON` is the strict one-shot wire boundary. Published [request and report contracts](contracts.md) reuse canonical analysis, environment, schema, and closure evidence.
