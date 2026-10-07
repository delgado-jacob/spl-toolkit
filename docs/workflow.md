---
title: "Offline detection workflows"
layout: page
---

# Offline detection workflows

Assess explicitly selected detections against captured environment and field evidence, retain results for CI, compare saved facts, and project evidence for an external assistant. Every operation is offline and deterministic for the supplied inputs and toolkit capabilities. The toolkit does not execute queries, retrieve live inventories, call an AI provider, or establish that a proposed change preserves detection intent.

## Assess and retain results

From the repository root after building, run:

```bash
./build/spl-toolkit workflow assess --request examples/workflow/request.json --format json --output workflow.json
./build/spl-toolkit workflow assess --manifest examples/workflow/manifest.json --settings examples/workflow/settings.json --format json --output workflow-manifest.json
```

The [inline request](https://github.com/delgado-jacob/spl-toolkit/blob/main/examples/workflow/request.json) selects two documents. `complete` uses literal SPL2 input and compatibility mode. `portable` resolves `$events` to two captured dataset names using inline schemas and explicit original-role bindings. Its `events_good` sibling verifies; `events_missing` fails because its complete schema lacks required `id`. Both commands retain the report and exit **1**, although workflow execution completes. The manifest reads contained query paths; their bytes match the inline documents, including trailing newlines. File origins and source IDs may differ between acquisition modes.

Each document ID must have exactly one settings entry containing `compatibility` or `resolution`. Settings contain Snapshot v1/v2, optional SchemaBundle v1, and explicit scope and bindings. Resolution uses original input IDs derived from canonical analysis of the exact query. The settings file is the same inline settings object used by the request. Directory selection also requires `--settings`; filesystem selection belongs to the CLI and is unavailable over HTTP.

| CI exit | Meaning |
| --- | --- |
| 0 | Complete execution and valid assessment; comparison has complete unchanged evidence |
| 1 | Complete execution with definite invalid content or affected comparison evidence |
| 2 | Input/configuration/tooling failure, or retained acquisition/internal/traversal failure |
| 3 | Complete execution with incomplete assessment or indeterminate comparison |

`--format text`, `json`, `sarif`, `graph`, and `bom` select assessment renderings with the same CI outcome. Graph and BOM retain subjects for unsuccessful candidates and pointers into saved canonical evidence. Full reports and exports can contain query text, names, source paths, diagnostics, captured definitions, and object metadata. Treat them as sensitive artifacts. SARIF source snippets are omitted by default; this does not make names, messages, or paths public-safe.

## Compare captured facts

```bash
./build/spl-toolkit workflow compare --before workflow.json --after workflow.json --format json --output comparison.json
```

Comparison never reevaluates queries. It retains before/after evidence, classifies affected, unchanged, indeterminate, or failed entries, and reports unmatched or ambiguous correspondence. Equal JSON alone does not establish complete evidence. The mixed example compares unchanged (exit 0): it retains complete relevant evidence for both the verified and failed siblings. Comparing a complete literal assessment to itself also exits 0; equal saved reports with incomplete relevant evidence remain indeterminate (exit 3). A failed resolution sibling is content evidence, distinct from an entry's tooling failure, which produces comparison exit 2. Captured capability revisions remain historical facts; comparison does not substitute the currently installed revision.

## Project evidence with explicit disclosure

```bash
./build/spl-toolkit workflow evidence --report workflow.json --output evidence.json
./build/spl-toolkit workflow evidence --comparison comparison.json --include query_text --output evidence-with-query.json
```

Successful projection exits 0 even when `source_ci_exit_code` records 1, 2, or 3. The default contains fixed vocabulary, counts, coverage, and local traversal tokens; requested/emitted disclosure arrays are empty and items contain no `details`. Tokens identify positions within this projection and are not identities across edits.

Each `--include` opts into one potentially sensitive category: `query_text`, `requirement_names`, `source_identity`, `environment_metadata`, `definitions`, `diagnostic_details`, or `artifact_identity`. Repeat the flag for multiple categories. JSON APIs require an explicit `include` array, including `[]` for the default. The disclosure object records requested, emitted, omitted, and potentially sensitive categories. Review opt-ins before sending evidence to an external assistant; projection does not restore trusted resolution proof.

## Recheck a proposal with fresh context

```bash
./build/spl-toolkit workflow recheck --request examples/workflow/recheck-query.json --format json --output recheck-query.json
./build/spl-toolkit workflow recheck --request examples/workflow/recheck-resolution.json --format json --output recheck-resolution.json
```

A query proposal supplies a complete replacement document and compatibility settings. A resolution proposal supplies resolution settings and reuses the original document. Both carry explicit original context, snapshot, optional schemas, and fresh settings. The query example changes literal `id` from 1 to 2 and exits 0; the resolution example retains one verified and one failed sibling and exits 1. Original/proposed hashes link exact source bytes. Neither a hash nor a valid fresh assessment proves preserved detection intent.

## Native APIs and HTTP

Go offers `workflow.AssessJSON`, `CompareJSON`, `EvidenceJSON`, and `RecheckJSON`, plus typed operations and graph/BOM/SARIF exports. Run the [Go example](https://github.com/delgado-jacob/spl-toolkit/blob/main/examples/go/workflow/main.go):

```bash
go run ./examples/go/workflow examples/workflow/request.json
```

The installed Python `SPLMapper` exposes `workflow_assess`, `workflow_compare`, `workflow_evidence`, and `workflow_recheck`. The [native Python example](https://github.com/delgado-jacob/spl-toolkit/blob/main/examples/workflow/native.py) reads its sibling request, assesses it, and prints the default evidence projection:

```bash
python3 examples/workflow/native.py
```

`POST /workflow/assess`, `/workflow/compare`, `/workflow/evidence`, and `/workflow/recheck` accept strict inline JSON with an 8 MiB body limit. Assessment `format` selects JSON, text, SARIF, graph, or BOM; other HTTP routes return their canonical JSON report. Content outcomes and retained entry failures return HTTP 200; malformed requests return 400. HTTP status and evidence-projection success are separate from assessment CI status.

The [machine contracts](contracts.md) publish strict v1 workflow envelopes and reuse existing canonical report definitions. Unknown members, duplicate JSON keys, null optional members, malformed Unicode, and trailing JSON are rejected by runtime decoders. JSON Schema cannot check duplicate keys, lexical integer encoding, byte-exact hashes, cross-document IDs, or proof authority; those remain runtime checks.
