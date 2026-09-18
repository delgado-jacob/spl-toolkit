---
title: "Machine contracts"
layout: page
---

# Machine contracts

The [contract registry](../contracts/README.md) lists every v1 schema and request
entry point. It covers QueryDocument, capabilities, analysis, direct query
requirements, field-list and JSON Schema/OCSF validation, rewrite, corpus,
manifest, graph, impact, advanced document view and LSP configuration. Python
wheels include these schemas under
`spl_toolkit/contracts`; source/release artifacts retain their provenance and
third-party notices. The product version remains 0.1.1.

Owned contracts use JSON Schema Draft 2020-12 and integer `schema_version: 1`.
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
68 SPL records with 67 evidence cases and 98 SPL2 records with 108 evidence cases.
The corpus establishes static local toolkit behavior only, not live Splunk
execution, runtime equivalence, environment compatibility, authorization or
upstream support.

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
