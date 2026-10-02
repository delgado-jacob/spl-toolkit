---
title: "Machine contracts"
layout: page
---

# Machine contracts

The [contract registry](https://github.com/delgado-jacob/spl-toolkit/blob/main/contracts/README.md) lists every v1 and v2 schema and request
entry point. It covers QueryDocument, capabilities, analysis, direct query
requirements, caller-supplied knowledge-object closure, field-list and JSON
Schema/OCSF validation, rewrite, corpus, manifest, graph, impact, advanced
document view and LSP configuration. Python
wheels include these schemas under
`spl_toolkit/contracts`; source/release artifacts retain their provenance and
third-party notices. The product version remains 0.1.1.

The closure family uses `definitions.schema.json` for the supplied scope,
collections, and objects; `closure-request.schema.json` for the inline request;
`closure.schema.json` for the full report; and `closure-graph.schema.json` and
`detection-bom.schema.json` for its two projections. A `complete` collection is
a caller assertion about one object kind in `scope_id`. Report completeness
requires effective-query, traversed-definition, resolution, collection, and
expansion coverage with no gaps or definite invalid finding. Bindings use the
exact original-document SHA256 digest and half-open UTF-8 byte range; source-map
ranges retain query or definition origin. These schemas validate local values,
not inventory truth, live execution, authorization, or deployment.

The environment family uses `environment-snapshot.schema.json` for a strict
offline capture, `field-schema-bundle.schema.json` for independently supplied
field evidence, and `environment-validation.schema.json` for the canonical
report and its `#/$defs/Request` inline input. A snapshot's per-kind collection
coverage is a producer promise within its capture scope. Each schema binding
has separate source coverage; it cannot make an open schema projection
conclusive. Computed snapshot and bundle digests identify normalized supplied
content independently. `valid` means the supplied claims are internally
consistent with no coverage gaps; `partial` preserves unknown evidence; and
`invalid` records malformed or contradictory input. Schema checks cannot prove
collection truth, live query compatibility, authorization, or execution.

Observed Snapshot v2 uses `v2/environment-snapshot.schema.json` and `v2/shared.schema.json`, retaining unchanged v1 object/provenance references. The [exporter](splunk-exporter.md) acquisition report uses `v1/environment-export-report.schema.json`. Register both directories locally; source/release artifacts and native Python wheels carry both versions. Snapshot v1 remains supported. Observation is scoped to method, index/datatype, visibility and bucket-overlap window; complete observed absence means `not_observed`, not configured absence. The validation envelope and report remain v1.

Owned contracts use JSON Schema Draft 2020-12 and integer schema versions. Snapshot v2 uses `schema_version: 2`; existing report families retain `schema_version: 1`.
The named family establishes what that integer versions. The unmodified official
OASIS SARIF Errata 01 schema uses Draft 4 and SARIF `version: "2.1.0"`.
Register all declared IDs locally and deny unknown retrieval. Schema IDs are
identities, not permission to fetch network resources.

Canonical/new requests reject unknown keys, null optional objects, invalid
selectors and conflicting union members. Outputs tolerate unknown additive
properties at documented extension points. Required-field removal, changed
meaning/types and new values in closed enums require a contract version change.
Diagnostic code/reason strings remain extensible; consumers must preserve
unrecognized values. Existing legacy endpoint decoding remains compatible.

This milestone makes a one-time, reviewed pre-consumer v1 correction to the
new SPL2 typed-field projection: `field_identity` and `output_identity` retain
distinct atomic and path identities even when their display names coincide.
No released consumer had adopted the earlier lossy projection. Once consumed,
the version-change rule above still applies to changed meaning or types.

Canonical source offsets are zero-based UTF-8 bytes; line/columns are one-based
Unicode code points with half-open ranges. SARIF declares `unicodeCodePoints`;
LSP converts to zero-based UTF-16. Source IDs are opaque metadata, not physical
paths. SHA256 covers the exact UTF-8 snapshot. Analysis revision/context IDs also
include normalized selectors, tool/capability identity and complete prepared
target contents; equal caller labels cannot disguise changed schemas.

Content status uses invalid before incomplete before valid, independently of
execution completeness. Coverage denominators describe analyzed documents;
acquisition failures have no synthetic canonical analysis. Analysis-only schema
coverage is explicitly not requested. Unknown commands/forms, dynamic bindings
and uncertain lineage remain visible in capabilities, diagnostics and reports.

Diagnostic objects carry code, severity, category, message, location and context.
Syntax errors (`SPL_SYNTAX_ERROR`) mean the supported grammar rejected the source;
semantic/unsupported diagnostics describe modeled limitations or binding facts;
validation diagnostics describe the selected field/schema contract. Errors
contribute invalidity, while uncertainty can produce incomplete without an error.
Consumers should use the emitted category/severity and capability form record,
not infer complete support from a command name alone. The current CLI
`capabilities --language spl` and `capabilities --language spl2` reports publish
the actual supported forms and restrictions.

Capability records have five separately reported and independently evidenced
dimensions: syntax, semantics, requirements, linting and safe rewriting. Each is
supported, partial, unsupported, not applicable or unassessed. Summary counts are
exact integers with these identities:
`applicable = supported + partial + unsupported + unassessed`; record count equals
applicable plus not applicable; covered equals
supported. Partial and unsupported claims do not receive covered credit. The
contract defines no percentage or composite score.

Evidence IDs resolve within the manifest to typed observations and provenance.
The same ID keeps the same reviewed scope; a broadened form uses a new ID unless
a reviewed scope correction changes the original boundary. `grammar_registered`
records parser registration separately from syntax coverage, and ordinary parser
or semantic diagnostics do not count as lint evidence. Current exact totals are
106 SPL records with 107 evidence cases and 141 SPL2 records with 149 evidence
cases. The current SPL revision is
`sha256:8a612f2064da24552a68faec261968731a6f1f1c7279ffabb77a59c7dfdb2008`;
the SPL2 revision is
`sha256:6495f77a0747f04e9690f5360cf4d024cf95a727bf05ab7cb7c9e1a96e1383dc`.

The SPL records cover bounded field-flow semantics for exact `tstats`, selected
field commands, and selected function arities. Exact macro invocations emit a
direct macro requirement and source-located unresolved-expansion gap. Branch
children retain direct requirements and evidence, while merge effects keep parent
outputs uncertain. Dynamic identities, unsupported options, unresolved macro
expansion, and branch merging remain incomplete. The toolkit does not execute SPL,
evaluate regular expressions, compare result rows, model acceleration, load macro
definitions, or certify runtime compatibility. The corpus also does not prove
authorization or upstream support.

| Stable code | Meaning |
| --- | --- |
| `SPL_SYNTAX_ERROR` | Supported grammar rejected the source; error/syntax finding. |
| `SPL_UNAVAILABLE_FIELD` | A required field is known unavailable after pipeline transfer; error/unavailable-field finding. |
| `SPL_UNSUPPORTED_COMMAND` | Command effects are unmodeled; warning and incomplete semantics. |
| `SPL_UNSUPPORTED_FUNCTION` | Function interpretation is unmodeled; warning and incomplete semantics. |
| `SPL_UNSUPPORTED_SEMANTICS` | This command/expression form lacks proved semantics; warning. |
| `SPL_DYNAMIC_REFERENCE` | Dynamic identity or expansion cannot be resolved statically; warning. |
| `SPL_UNRESOLVED_WILDCARD` | Exact wildcard membership is unproved; warning. |
| `SPL_UNKNOWN_FIELD` | A source obligation is absent from the selected closed field catalog; error. |
| `SPL_INDETERMINATE_FIELD` | Field-catalog validation is inconclusive; preserve the emitted reason. |
| `SPL_TOOLING_INCOMPLETE` | SARIF-only unlocated completeness summary; warning, with no invented source range. |

Schema/OCSF outcomes carry their own canonical messages and context; inspect the
complete validation report. The table explains common codes and is not a closed
list of all future codes or rewrite audit reasons.

JSON Schema instance validation is independent of the field-projection engine.
`tests/acceptance/test_machine_contracts.py` validates emitted reports and authored
positive/negative requests with a full validator, format checking and denied
external retrieval. Decoder tests additionally cover duplicate JSON keys,
integer spelling, unique IDs and semantic preparation constraints that a parsed
JSON Schema instance alone cannot prove.
