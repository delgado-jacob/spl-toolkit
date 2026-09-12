# M7 draft architecture review

Verdict: bounded architectural review is clean apart from the already identified unknown-command example correction. No additional actionable P1/P2 architectural or task-dependency gap was found in the reviewed draft. This is not final-plan approval, production authorization, or runtime/OS/consumer acceptance.

## Reviewed inputs

- `milestone-7-implementation-plan.md`: all 416 lines; SHA-256 `42fe1312efcfecb5f24e092b8fa882bfef5e93e580e0612a81c1001fda30a5a8`.
- Approved `design-spec.md`: all 135 lines; SHA-256 `c73d27bfd4fb879481ce4cff218cf85cac1299881895a8279f6f2275eed1c0ba`.
- Root's bounded review instructions and clarification that the thin `document` CLI is approved advanced-view serialization with no additional analysis semantics.

The line anchors below identify the reviewed draft, before root's pending example correction. Any subsequent edit has a different review identity.

## Known correction owned by root

**P2 — Use an explicit unknown command in the mixed-status examples (plan lines 205 and 384).** Both examples currently use `mystery | table host`. Root has already established that the existing SPL frontend treats this bare initial term as implicit search. The intended unsupported-command fixture must be `| mystery | table host`; otherwise its expected incomplete semantic coverage and unsupported-command finding do not follow from the selected input. Root will make the correction. This review did not run a duplicate probe or change the plan.

## Architectural and dependency checks

| Area | Draft anchors | Review result |
| --- | --- | --- |
| Acquisition versus canonical in-memory input | 142–164, 180–196, 284–288 | The manifest/loader owns physical acquisition and strict path structure. In-memory wire requests carry canonical inline documents, cannot request file loading, and cannot assert acquisition failures. Local normalized input can preserve actual acquisition failures without fabricating analysis. |
| Anchored filesystem boundary | 86, 130, 191–197, 323 | The plan requires handle-relative opening/enumeration, no-follow/no-reparse enforcement, opened-object checks and deterministic component-swap tests. It rejects pathname prechecks as the containment proof and assigns the boundary to one owner. Supported-OS runtime proof remains distinct from compilation. |
| Shared preparation before acquisition | 68, 88, 128, 148–156, 172, 203–209, 249, 285 | Separate prepared handles allow invalid shared inputs to fail before acquisition, including zero successful reads, without an artificial query or changed empty-batch semantics. Both impact sides prepare before the same corpus is loaded once. Exact accepted canonical access is explicitly pending reconciliation. |
| Report identity, aggregation and coverage | 32–34, 144–156, 182–184, 205–209 | Identity covers exact source bytes, normalized selectors, tool/contracts and target content/options/resources. The report keeps ordered canonical evaluations, failures, explicit coverage denominators, schema `not_requested`, and operational completeness independent of invalid/incomplete/valid content status. |
| Graph evidence and source integrity | 168, 225–229 | IDs are qualified by document/revision and transition occurrence/phase, endpoint/report pointers must resolve, and SQL lexical stages remain separate from evaluation order. Conditional and incomplete evidence is retained; malformed external evidence is rejected rather than repaired with invented nodes. |
| SARIF source and invocation integrity | 170, 235–240, 257–262, 312 | Semantic tests own URI encoding, original-snapshot bounds, Unicode units, indices, virtual-source contents and invocation/content separation. Official-schema validation and actual converter acceptance are assigned separately. The converter's navigation/column limitations are stated explicitly. |
| Impact correspondence | 172, 246–251 | The comparison retains complete before/after canonical evidence, verifies exact original-reference correspondence, uses M6 audit provenance for candidates and leaves duplicate/unproven relations ambiguous. It compares finding/outcome changes beyond coarse status; equal incomplete evidence remains indeterminate. |
| Detached mutable document access | 166, 215–219 | Nested data is detached and lookups inspect the caller's current snapshot instead of stale indices. Range checks, half-open matching, overlapping references and EOF behavior are explicit. Caller mutation does not preserve engine attestation or invent persistent cross-edit identity. |
| LSP lifecycle and configuration | 90, 174, 268–276 | URI/open generation/version/configuration revision govern work and publication. Invalid requested configuration visibly pauses publication while buffer lifecycle continues; correction resumes current snapshots. Initialization failure, clear-on-close, reopening, full-sync versions and UTF-16 conversion are explicit. |
| LSP requests versus diagnostic publication | 272–276 | Obsolete diagnostics are suppressed at publication time, while position requests capture their snapshot in message order and still complete after ordinary edits. Cancellation receives a response; configuration invalidation is distinguished from ordinary edits. Bounded scheduling does not claim forced engine interruption. |
| Machine-schema validation and compatibility | 132, 257–262 | Task 8 owns published schema families, a full standards-conforming instance validator, offline registry, actual emitted JSON, negative fixtures and old-wire compatibility. Decoder duplicate-key checks remain separate from JSON Schema validation. The schema dialect/dependency choice remains an explicit preflight item. |
| Adapter, installed-package and real-consumer ownership | 282–327, 369–378 | CLI/HTTP/native operations delegate canonical values; advanced-view serialization does not add semantics. Package checks require outside-checkout native origins, wheel-from-sdist, mandatory suites and artifact hashes. The actual editor receives diagnostics/highlights through the official client; converter runs and source semantics are independently asserted. Serial ownership and two immutable root windows prevent those results being conflated with unit checks. |

The ordering supports the declared dependencies: contracts and loader precede orchestration; document/graph/SARIF/impact and wire schemas precede root's Go-only window; LSP/adapters/native packaging and real consumers follow that window. Task 3 owns canonical target-access reconciliation, Task 2 owns platform opening, and Task 9 owns protocol behavior; later gaps return through scoped reviewed follow-ups. Task 12 cannot replace actual client evidence with a fake server or injected diagnostics.

## Limits and remaining gates

The six reconciliation items at plan lines 123–136 are intentional pending accepted-M5/M6/API/dependency/platform/package facts. Their absence is not reported as a draft defect. The reviewed draft expressly requires replacing them with actual facts and updating affected interfaces/tasks before final release; this review does not freeze proposed symbols or infer they exist.

Only filesystem reads, SHA-256/line inspection and this review-artifact write were performed. No active product source or private root oracles were inspected. No immutable source fallback was needed. No toolkit tests, builds, runtime probes, GUI/client runs, dependency provisioning, commits or subagents were used. No implementation, OS containment, installed-package, SARIF consumer or real-editor result is established by this review.

Root's example correction, accepted predecessor reconciliation, final-plan approval, implementation review, and the two independent root acceptance windows remain separate required work.
