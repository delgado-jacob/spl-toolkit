# Milestone 6 — Safe Rewrite and Mapping 2.0 Design

Status: architectural design and complete written spec approved by the delegated controller on 2026-09-07. Exact internal interfaces and the implementation plan require the verified Milestone 5 handoff.

## Purpose and scope

Provide deterministic, source-preserving mappings for exact supported field, index, source, sourcetype, lookup, dataset and data-model references in SPL and standalone SPL2. A caller previews a candidate and its exact audit, then requests apply through the same Go operation. Static source identity, derived bindings, aliases, scope and lineage govern eligibility. Every refusal explains whether a rule did not match, evidence was unavailable, or alternatives were ambiguous.

Go owns behavior; CLI, real native Python and stateless REST expose equivalent reports. Preserve legacy mapping/configuration behavior through its existing entry points. The new operation does not convert languages, beautify queries, rewrite dynamic identifiers, learn mappings, repair arbitrary syntax, infer raw-to-data-model or index-to-tstats transformations, execute searches, or promise runtime equivalence. No account, network, persistence or schema download is introduced. Temporary `_build_plan/` artifacts are never runtime dependencies.

## Architecture and alternatives

Add a canonical rewrite consumer over the finished analysis and validation packages. The selected frontend remains responsible for parsing, reference interpretation, quoting and replacement eligibility; the shared kernel remains responsible for source identity, flow, scopes, predicate facts, binding and lineage. Rewrite owns strict rule preparation, condition selection, edit groups, conflict detection, source edits, audit composition and the verification gate. Adapters own transport only.

Necessary predicate-fact and render-eligibility hooks belong in the canonical frontends/kernel. Choose their actual types only against verified Milestone 5. The public report must not expose ANTLR contexts or tokens. Do not parse Boolean meaning, identifier paths, quoting, command structure or scope from strings in rewrite. Do not replay a second transfer engine. Conditions consume source-scoped facts produced during canonical flow; post-verification consumes canonical analyses and mapped reference identities.

Three conditional approaches were considered. Guaranteed scoped facts are conservative and explainable. Occurrence matching is smaller but unsafe for OR, NOT, nested searches and later filters. A general satisfiability/value engine substantially exceeds the milestone. Guaranteed scoped facts were approved.

Three field-edit approaches were considered. Editing source tokens alone can break implicit aggregate output reads; general alias renaming changes explicitly created user names. Source-identity mapping with atomic linked edits preserves explicit aliases while maintaining proven implicit-name relationships. This grouped approach was approved.

A global complete-analysis prerequisite would exclude independently provable edits in nested queries. Publishing candidates despite failed verification would weaken safe apply. The approved gate permits unrelated preexisting semantic incompleteness, while refusing affected unknown bindings, new uncertainty, definite errors and inconclusive explicitly requested schema validation.

## Strict request and rule contract

The new request has integer `schema_version: 1`, `mode` (`preview` or `apply`, omitted means preview), a canonical `document`, ordered `rules`, and an optional `validation_target`. Reuse the existing field-list, JSON Schema and OCSF target contracts through a strict tagged selection; do not invent a new catalog or schema format. Targets describe the destination query and are evaluated after proposed edits. They are not preconditions that the original source names satisfy the destination schema. Exact Go target types and adapter envelopes are selected against finished M5 APIs.

Each rule has a unique nonempty `id`, `kind`, `source`, `target`, and optional `when`. Supported kinds are `field`, `index`, `source`, `sourcetype`, `lookup`, `dataset`, and `data_model`. Source/target identities are strict tagged values: an atom has `name: <logical string>`; a statically navigated field path has `path: [<logical segment>, ...]`. Exactly one form is allowed. Paths are field-only, nonempty and contain nonempty segments. A dotted atom such as `{"name":"actor.name"}` differs from `{"path":["actor","name"]}`. A frontend must prove that it can render the requested identity in the original role; an unsupported atom/path conversion is refused. Query normalization and case rules stay frontend-owned. Configuration names are logical names, never query fragments, quoting syntax, wildcards or executable text.

Reject unknown keys/kinds/modes/versions, null required values, empty/whitespace-only names, malformed conditions, duplicate rule IDs and invalid target requests. Preserve meaningful name whitespace and case. Do not trim or parse caller names as SPL. Empty rules are a valid no-op. Duplicate semantic rules with distinct IDs are permitted and their identical edits coalesce. A rule mapping an identity to itself is an explained no-op. Rules have no priority, implicit override or cascading evaluation. Caller-injected condition contexts, legacy regex operators, arbitrary expressions and executable rule code are not accepted.

For example, the conceptual request below changes a source field only where the query establishes the source-type fact:

```json
{
  "schema_version": 1,
  "mode": "preview",
  "document": {"text": "search sourcetype=auth src=alice | table src", "language": "spl"},
  "rules": [{
    "id": "auth-user",
    "kind": "field",
    "source": {"name": "src"},
    "target": {"name": "user"},
    "when": {"fact": "literal", "kind": "sourcetype", "identity": {"name": "sourcetype"}, "operator": "equals", "value": "auth"}
  }]
}
```

The configuration's literal-fact identity selects the grammar-recognized dependency selector or event-field identity. A SQL expression field named sourcetype remains a field-kind fact; it is not reclassified as a dependency selector by spelling.

## Conditions and flow scope

A condition is exactly one leaf or combinator. Combinators are `{"all":[...]}` and `{"any":[...]}` with nonempty child lists. A literal leaf has `fact: literal`, kind `field`, `index`, `source` or `sourcetype`, a typed `identity`, operator `equals` or `contains`, and a grammar-compatible scalar `value`. Contains requires a string and applies only to an established literal value, using exact case-sensitive substring comparison. It never searches query text. Scalar equality compares canonical literal values without coercing strings to numbers or booleans. A reference leaf has `fact: source_reference_present`, `kind: field`, and a typed `identity`; no operator/value is allowed. There is no condition-language NOT in this milestone.

Conditions have true, false and unknown results. True requires the requested fact to be established. False means complete relevant evidence does not establish that query fact; it does not assert that an event lacks the field or cannot contain a value. Unknown means unsupported syntax, effects, binding or scope prevent this decision. Unknown rules skip with explicit uncertainty; a false condition is an ordinary explained skip. All/any use three-valued composition: all is false if any child is false, true when all are true, otherwise unknown; any is true if any child is true, false when all are false, otherwise unknown. Retain evidence for each child so a controlling known result does not obscure unrelated analysis limitations.

Canonical predicate lowering establishes positive exact literal restrictions. AND combines compatible guaranteed facts; OR intersects facts proven in every branch. Query-language NOT establishes no positive equality, and no broad theorem simplification or double-negation solver is required. Conflicting or otherwise unproven constraints never justify vacuous application. `sourcetype=a OR sourcetype=b` does not establish `sourcetype=a`; an unsupported predicate that can change the relevant conclusion makes it unknown. Known facts independent of unsupported predicates can remain usable when canonical structure proves their guarantee.

Facts are evaluated from the original document at each candidate's source/flow point, including the complete relevant predicate at that point. Later filters cannot retroactively authorize earlier edits. Propagation and invalidation follow canonical transfer: redefine, remove, unknown effects, source resets and unresolved merges may invalidate relevant evidence. Source-reference-present means a proven static source reference with that identity in the applicable scope/flow context. It never means actual event presence, and it never implies an equality. Reference evidence hidden behind indeterminate bindings stays unknown.

Independent child searches do not inherit parent facts or establish parent facts. Inherited subpipes receive only the copied facts the canonical flow proves apply, and their changes do not leak into parents. SQL uses logical evaluation phases from M5 even when SELECT text appears first; source-order scanning cannot substitute for flow. Facts and reference eligibility are determined once from the original query, so mapping one dependency does not trigger another rule in the same request.

## Field identity, aliases and linked changes

Match exact canonical identity and kind, not every token with the same spelling. Ordinary source-bound reads and proven passthrough source selectors are eligible. Creation-only references and explicitly created eval, rename, SQL and lookup aliases are not source-name replacements. Exact removal/exclusion operands are eligible only when the kernel proves which source identity they select; their lack of a schema read obligation is not itself a rewrite decision. Derived, unavailable or indeterminate bindings do not become source mappings merely because names match a rule.

For `rename src AS alias | table alias`, src-to-dst edits the rename input to dst and keeps alias and its downstream reads. An alias-to-other source rule cannot rename that derived alias. For `eval alias=src | table alias`, only src changes. Scope-specific identity permits the same spelling to map in an independent source subquery while staying unchanged as a parent-derived field.

An input replacement may alter a generated implicit output name, for example sum(src). Use canonical origin/lineage and frontend naming evidence to assemble a linked edit group that changes the source input and every affected exact downstream implicit-name read. Explicit user aliases stay unchanged. If naming, consumers, projection, qualification or scope linkage cannot be proven, skip the entire group. Do not silently add AS, manufacture aliases, or guess implicit names by formatting strings.

A group's proof must include its affected statically modeled uses and transitions. A relevant wildcard, dynamic reference, unknown command effect or unmodeled merge that prevents this closure makes the group ineligible. Unrelated incomplete scopes do not automatically taint an independently proven group. No unseen dynamic name or macro body is rewritten.

Where preserving one source identity requires linked edits at several flow points, every required edit must be independently authorized by its rule condition at that point. A later true condition cannot backfill an earlier false or unknown condition. If such a group cannot be completed without an unauthorized edit, skip the whole group as linked_edit_unproven. Do not silently split one proven identity into differently named consumers merely to publish a partial mapping.

## Dependency references and supported forms

Index/source/sourcetype mappings edit only exact grammar-established dependency values. Identically spelled event fields, literals, comments and options are not dependency matches. Lookup mappings edit exact catalog dependency identities and preserve known local-field/catalog-column distinctions. Dataset and data-model mappings edit supported exact dependency identities, including qualified forms only when the frontend proves their structure and replacement interval.

Mapping a model or dataset does not infer a corresponding field schema, alias or raw-to-tstats conversion. Explicit related mappings can participate in one proven edit group where canonical evidence establishes a coupling. An independently supplied qualified field mapping must retain its atom/path/alias distinction. Unsupported tstats/metrics/join semantics cannot be upgraded merely because a dependency token is recognized. The capability report states exact rewrite-supported roles and limitations separately from parse support and full command semantic coverage.

Macros, dynamically computed datasets/fields, wildcard names, unresolved aliases and damaged contexts are refused with located reasons. A target is encoded as a single proven identity in the original grammatical role; punctuation in caller text never becomes a pipeline command or an unquoted query fragment.

## Conflict and collision policy

Build candidates from original identities simultaneously. Source a-to-b and b-to-c do not become a-to-c. Coalesce identical replacements at the same interval with all contributing rule IDs. Two true rules proposing different replacements, incompatible overlapping intervals, or required linked edits that disagree make the connected edit group ambiguous and unchanged. Disjoint safe groups can proceed. Audit conflicting alternatives and their rule evidence rather than choosing request order or legacy priority.

Check static capture and collapse: distinct statically live source/derived identities may not silently become one binding; an existing alias must not capture a newly renamed source read. Swaps and chains are accepted only if canonical before/after relationships under the simultaneous mapping prove the intended distinct identities remain correctly bound. Name equality alone is not a proof of a safe swap.

This is an observed static collision boundary. It does not require proving that unreferenced event fields are absent, checking event payloads, or proving equivalent search results. Ordinary src-to-dst source migration is not blocked just because an open source could contain an unseen dst field. Report these limits without presenting a runtime-equivalence guarantee.

A dependent group's exclusion can change whether another proposed group is safe. Resolve conflict groups deterministically to a stable candidate, and verify the final selected set together. A surviving group cannot rely on a skipped edit to prevent capture. This is conflict closure over canonical identities, not a search for an optimally large subset of rewrites.

## Source preservation and audit coordinates

Construct the candidate from validated, nonoverlapping edit intervals over the original bytes or their canonical token equivalents. Preserve every byte outside those intervals, including comments, spaces, tabs, line breaks, CRLF, case and unaffected quote spelling. Within an edited identifier retain its existing quote style when the selected dialect and role can encode the target correctly; otherwise use the minimal required proven quoting. Escape the target through that frontend's rules. Never normalize the whole query or strip an EOF-looking user suffix.

Audit every edit's exact original and replacement text, original location, candidate location when included, original reference IDs, candidate reference IDs when resolved, rule IDs, group ID and reason. Locations use existing half-open UTF-8 byte offsets, one-based Unicode columns and CRLF behavior. Compute translation through the ordered edit map; downstream locations after multibyte or multiline replacements must point into the actual candidate. Do not compare original/candidate offsets directly or carry stale original IDs into candidate analysis.

Ordering is deterministic: rules preserve request order; candidate references and changes follow canonical source ordering with stable tie-breakers; contributing rule IDs follow request order; diagnostics follow the canonical location/code/message order. Empty collections serialize as arrays, never null. State is call-local, inputs are copied as needed, and concurrent identical calls return equivalent JSON values.

## Preview, apply and post-verification

Preview builds the candidate from independently eligible, unambiguous groups, then reparses and reanalyzes it in the same document language/profile/version. If requested, validate that destination candidate through the existing canonical field-list or schema validator. Retain original analysis, candidate analysis and validation evidence. Do not apply from syntax-damaged original input where exact token/binding evidence cannot be established; the conservative default for original syntax errors is no edits and an invalid report with original bytes retained.

Verification is stronger than parser success. Map original reference identities and locations through the edit map, and prove preservation of the intended static binding/lineage relationships under the selected mapping. Check changed implicit outputs, capture, scope boundaries and SQL phases. New uncertainty is assessed by affected scopes, identities, transitions and translated source evidence, never by diagnostic counts alone. An uncertainty barrier touching an edited group fails its proof.

Apply uses one whole-request verification gate after safe-group selection. It publishes the candidate only if candidate parsing succeeds, static relationship checks succeed, no new relevant uncertainty is introduced, and candidate analysis has no definite errors. Without requested schema validation, unrelated preexisting semantic incompleteness may remain when every edited group is independently proven. A supplied validation target is a strict additional gate: its resulting validation report must be conclusive and error-free. Invalid or incomplete requested validation prevents commit, including incomplete analysis retained in that validation report.

If any gate fails, returned authoritative text is the original query, `committed` is false, and candidate text survives solely as preview evidence. Every proposed edit is marked not committed; apply audit entries that were included in the candidate become skipped with `post_verification_failed` and retain their proposed replacement/candidate coordinates. No automatic rollback subset search or second attempt publishes only some failed-request edits.

Independent safe groups may commit while ambiguous/unknown groups remain unchanged if the resulting candidate passes the gate. Such a result is explicitly incomplete with `committed: true`. A syntactically and semantically valid candidate does not erase unknown mapping eligibility. Unmatched rules, false conditions, explicit derived names outside source mapping, and source-equals-target rules are ordinary explained skips and do not automatically make the result incomplete.

## Result and failure contracts

The versioned result contains the normalized request document, mode, `status`, coverage and reasons, `original_text`, `candidate_text`, authoritative returned `text`, `committed`, source-ordered change audit, rule evaluations, original analysis, candidate analysis and optional candidate validation. The exact serialization/type factoring must fit verified M5, while preserving these observable fields and distinctions. A result's top-level status combines query/verification errors and rewrite uncertainty using invalid, incomplete, valid precedence. Request errors remain separate API/transport errors.

Audit outcomes are `applied`, `skipped`, and `ambiguous`. In preview, applied explicitly means included in the candidate, while `committed: false` and authoritative text remains original. In successful apply, applied entries identify committed edits. In failed apply they become skipped/post_verification_failed as above. An audit entry records `candidate_applied` and `committed` separately so preview and failed-apply proposal evidence is unambiguous. Ambiguous entries never modify candidate or returned text.

No-op preview/apply retains original bytes, `committed: false`, and no fabricated applied entry. Candidate analysis may reuse equivalent original analysis when text is unchanged, but an optional destination target is still evaluated. Rule-level skip entries explain no_match/condition_false/no_change even without a source location; a location is never fabricated for an absent reference. Rejected dynamic/unsupported references retain their real locations and diagnostics.

Stable rewrite reason codes distinguish condition_unknown, unsupported_reference, dynamic_reference, conflicting_targets, overlapping_edits, binding_collision, linked_edit_unproven, target_not_renderable, and post_verification_failed. Definite verification errors yield invalid, while ambiguous or unresolved eligibility yields incomplete. Preserve the originating analysis/schema diagnostic evidence instead of replacing every failure with a generic rewrite message. Malformed requests, inaccessible files and internal failures never return an empty successful report.

## Public interfaces and batches

The public Go operations are Rewrite and RewriteBatch in the canonical rewrite package. Batch requests share prepared rules/mode/validation target and contain ordered canonical documents; each document selects its language and retains source identity. Empty batches are request errors. Invalid query content produces its ordered per-query result. Malformed rules/targets/options/document request shapes or internal failures reject the batch without publishing a partial success report. There is no cross-document transaction or filesystem mutation: individual apply results commit only their returned text. Aggregate status follows invalid, incomplete, valid precedence.

Add CLI `rewrite`, defaulting to preview; `--apply` selects apply. A local rules file supplies the strict versioned mapping definition. Reuse existing query/positional/file/stdin/batch inputs, language/profile/compatibility-version/source-ID flags, target selectors, conflict rules, `--format text|json` and requested report `--output`. Batch mode takes dialect options from documents. Report output is the only filesystem mutation; there is no in-place query editing. Text output clearly labels candidate versus returned text, committed state, skipped/ambiguous groups and validation failures. JSON is canonical report data. Write content reports before exits 0 valid, 1 invalid, 3 incomplete; usage/request/I/O/internal errors exit 2. Do not imply that exit 3 means nothing committed: callers inspect committed and audit.

Python adds rewrite and rewrite_batch methods through the actual native operation boundary, with document dialect options and canonical dictionaries. Preserve operation guards, closed-handle behavior, owned results, exception conventions and finally freeing. Update native source packaging manifests. No Python parser, condition evaluator or matcher is added.

REST adds POST `/api/v1/query/rewrite` and `/api/v1/query/rewrite/batch` with inline canonical requests, rules and targets. Existing body limits and error middleware apply. Return HTTP 200 for valid/invalid/incomplete query-level results; malformed input is HTTP 400 and internal failure is a server error. Server filesystem paths and retrieval URLs are not rule/schema sources. Existing fixed-SPL legacy mapping routes and methods remain available without silently adopting new rule precedence or syntax.

## Verification and acceptance

Create a shared maintained corpus outside `_build_plan/` covering both selected dialects and every advertised exact dependency/field form. Compare complete canonical JSON across Go, CLI, real native Python and REST for preview/apply and ordered mixed-dialect batches; normalize only JSON object-key order. Include request errors and runtime-offline operation. Capabilities and maintained documentation must list exact supported render roles and semantic refusal cases.

Required scenarios include:

- Source reads versus eval/rename/lookup/SQL-created aliases; repeated spellings in independent and inherited scopes; supported SQL lexical/flow-order differences.
- Linked implicit aggregate outputs and downstream names, explicit aliases, lost/indeterminate lineage and refusal of the complete affected group.
- Positive AND facts, common/noncommon OR facts, query NOT, conflicting restrictions, unsupported predicates, later-stage facts, overwritten/removed bindings, source-reference-present, contains on known strings, typed literal distinctions, and false versus unknown condition outcomes.
- Conflicting true rules, identical coalesced edits, interval overlaps, static target capture/collapse, simultaneous swaps/chains, and ordinary migration with an open source universe without imaginary event-absence requirements.
- Exact index/source/sourcetype/lookup/dataset/data-model references versus similar literals, fields and comments; alias qualification and atom/path distinctions; explicit refusal of macros, wildcard/dynamic identifiers and unsupported roles.
- Spaces, tabs, CRLF, Unicode, quotes/escapes, keyword targets, punctuation resembling commands, comments containing pipes, multiline sources, multibyte replacements and exact original/candidate source slices.
- Original malformed syntax, post-parse failure, static binding drift without a syntax error, newly affected uncertainty, unrelated incomplete scopes, invalid/incomplete optional destination validation, candidate retained after failed apply, partial safe-group commit with incomplete status, and no-op behavior.
- Field-list, JSON Schema and local versioned OCSF destination validation through their existing report contracts, with no condition-driven schema or class selection.

Assert unaffected byte identity and exact edit application; do not merely compare formatted queries. Assert static before/after identities and lineage rather than only successful parsing. Test request input isolation, deterministic rule/evidence ordering, repeated/concurrent mixed-dialect calls, Go race checks, native Python ownership/concurrency and packaging. Preserve all legacy mapper/configuration tests. Run established build/generation checks if canonical grammar/render support changes. Local checks do not imply unexecuted Splunk runtime or cross-platform release acceptance.

## Handoff and next gate

The controller must review this written spec. Before final planning, obtain the verified M5 commit and handoff, inspect canonical predicate/flow/path/render/implicit-name support, SQL phases, validation target types, adapter conventions and capability APIs. Select the smallest canonical additions necessary for this behavior; if required evidence is missing, explicitly extend the canonical frontend/kernel instead of building a rewrite-side parser or flow engine.

The later plan must be `_build_plan/milestones/6-safe-rewrite-mapping/milestone-6-implementation-plan.md`. This document creates no final implementation task sequence, workers, commits or production changes. The controller retains milestone integration authority. Record implemented supported forms, refusal cases, source preservation, static verification and full surface parity in the eventual milestone log without claiming runtime equivalence.
