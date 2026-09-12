# SPL Toolkit

> **About these build-plan files:** Everything in `_build_plan/` (this PRD and the per-milestone folders) is a **temporary documentation and guidance artifact** for the initial build-out of this codebase. These files are not functional — no code, configuration, runtime logic, tests, or deployment process should import, read, reference, or depend on anything in `_build_plan/`. Once the initial milestones are built and shipped, the entire `_build_plan/` folder is expected to be deleted from the codebase. Do not treat it as long-living documentation.

## What we're building

SPL Toolkit will become a dependable, embeddable analysis and safe-rewrite library for standalone SPL and SPL2 queries. It will broadly understand the SPL2 `splunkd` search language, discover field and data dependencies, validate referenced fields against field lists, JSON Schema, and OCSF schemas, and return actionable diagnostics with an explicit `valid`, `invalid`, or `incomplete` result—never false confidence.

The product is a deterministic, offline-capable toolkit rather than a hosted service. Its Go API is canonical, while the CLI, Python package, and optional stateless REST adapter expose the same behavior. The work is divided into seven independently useful milestones so the existing SPL capabilities become trustworthy before schema validation, broad standalone SPL2, rewriting, and developer tooling build on them.

---

### Current state and product opportunity

The repository is a functional alpha with useful foundations: an ANTLR grammar, Go discovery and token-preserving mapping behavior, a CLI, a REST adapter, Python bindings, tests, and release scaffolding. At the roadmap audit point, the Go race-enabled suite passed, and the mocked Python suite reported 10 passing tests.

The current product cannot yet support a strong correctness claim:

- The grammar covers a limited, SPL1-oriented subset and rejects common expression, quoting, list, and standalone SPL2 forms.
- Discovery returns flat name lists rather than source locations, roles, scopes, lineage, or an explicit completeness result.
- Field discovery misses common reads, can misclassify functions and derived fields, and may discard otherwise useful discoveries after a parse failure.
- Some documented CLI, configuration, mapping, API, and performance capabilities are ahead of the implemented behavior.
- The Python boundary has correctness, memory-management, and concurrency risks, while its current tests do not exercise the real native library.
- Clean-checkout build and release paths need attention, generated parser code triggers static-analysis warnings, and parser/binding coverage is limited.

These findings make the opportunity clear: evolve the project from a promising mapper into a transparent static-analysis and safe-transformation foundation that other security tools can confidently embed.

---

### Product principles

- **Honesty over apparent coverage:** Parse coverage and semantic coverage are reported separately.
- **No silent uncertainty:** Unsupported or dynamic behavior produces `incomplete` with an explanation.
- **One shared meaning:** Go, CLI, Python, and REST return equivalent results.
- **Offline by default:** Queries and schemas remain local, and validation requires no account or network service.
- **Source-aware output:** Every finding and change links back to the relevant query text.
- **Safe transformation:** Ambiguous rewrites are skipped and reported rather than guessed.
- **Backward-aware evolution:** Existing supported SPL behavior remains available while public contracts mature deliberately.

---

### What the toolkit does

- Parses SPL and broadly parses standalone SPL2 for the `splunkd` compatibility profile.
- Discovers fields, indexes, sources, source types, datasets, lookups, data models, and macros.
- Tracks fields as commands read, create, rename, remove, aggregate, and pass them downstream.
- Validates source-required fields against field lists, JSON Schema, and versioned OCSF selections.
- Returns stable diagnostics, exact source locations, coverage details, and `valid`, `invalid`, or `incomplete` outcomes.
- Safely maps supported query references while preserving unaffected query text.
- Provides consistent Go, CLI, Python, and optional REST access for individual and batch workflows.
- Supports later CI, editor, dependency-graph, and migration-impact tooling through the same analysis model.

---

### Already provided by the existing codebase

- A Go 1.22 module and public mapper package.
- ANTLR4 lexer and parser grammars with generated Go parser code.
- Existing source, dependency, input-field, and derived-field discovery behavior.
- Token-stream-based field mapping that can preserve unaffected query text.
- A command-line application with validation, discovery, mapping, version, and demo entry points.
- A stateless REST server adapter and generated API-documentation scaffolding.
- A C-compatible binding layer and Python package structure.
- Go and Python tests, Docker files, Make targets, and release-oriented automation.

The roadmap retains Go and ANTLR as the core stack. The Go API, CLI, and Python package are first-class surfaces; the REST server remains a thin optional adapter. Milestone 1 standardizes and documents Python 3.11+ as the supported baseline even though the current package metadata permits older Python versions. No database, frontend framework, hosted infrastructure, or runtime credential is introduced. Field lists and schemas are provided locally, and versioned OCSF catalogs work without a live network connection.

---

### Out of scope

- **SPL2 modules and declarations** — Imports, exports, namespaces, named module searches, reusable functions, and user-defined declarations are deferred.
- **Non-`splunkd` compatibility profiles** — Edge Processor, Ingest Processor, and federated-search differences are later work.
- **Search execution** — The toolkit does not run searches against Splunk or inspect their results.
- **Immediate semantics for every parsed command** — Broad syntax coverage is required, but incomplete command semantics remain explicitly `incomplete`.
- **Whole-query SPL1-to-SPL2 transpilation** — Targeted safe mappings are supported; automatic language conversion is not.
- **Event-record validation** — The toolkit checks query references against schemas, not whether event payloads conform to them.
- **Hosted product features** — Accounts, teams, billing, persistent query history, and a hosted dashboard are excluded.
- **Arbitrary extensions and extra SDKs** — Go, CLI, Python, and the existing REST adapter stabilize before more bindings or a plugin system.
- **Machine-learned mappings and repairs** — Analysis and transformation remain deterministic and explainable.
- **Automatic raw-to-data-model translation** — Inferred index-to-data-model or `tstats` conversion is deferred.

---

### External integration policy

No external service, credential, or live network connection is required. The official OCSF project is the authoritative upstream schema source, but users validate against local, versioned catalogs. Splunk documentation and tooling may inform conformance fixtures without becoming runtime dependencies.

---

### Data model

This toolkit has no persistence requirement. Its data model is the set of structured documents and reports exchanged with callers.

#### Query Document

- **Original query text** — The source that is analyzed or rewritten.
- **Language** — Whether the caller selected SPL or SPL2.
- **Compatibility profile and version** — The language contract used to interpret the query.
- **Source identity** — An optional filename or caller-provided identifier.

Each Query Document produces query stages, references, lineage, and diagnostics.

#### Query Stage

- **Command or clause** — The operation represented by this stage.
- **Pipeline position** — Where the stage occurs in query flow.
- **Source location** — The exact portion of the original query.
- **Scope** — Its relationship to the root query, subsearch, join, or branch.
- **Semantic coverage** — Whether the toolkit fully understands the stage's field behavior.

Each Query Stage belongs to a Query Document and owns the references and field transitions that occur there.

#### Reference

- **Original and normalized name** — The spelling in the query and its comparable field path.
- **Kind** — Event field, index, source, source type, dataset, lookup, data model, or macro.
- **Role** — Read, create, rename, remove, filter, group, sort, or output.
- **Stage, scope, and location** — Where and why the reference appears.
- **Resolution form** — Exact, wildcard-based, or dynamically computed.

References participate in discovery, lineage, validation, and rewriting.

#### Field Lineage

- **Available fields** — What is known before and after a query stage.
- **Input relationships** — Which source or derived fields contribute to another field.
- **Transitions** — Creation, rename, removal, aggregation, and lookup output.
- **Uncertainty** — Transitions that cannot be resolved statically.

Field Lineage connects references across stages and scopes.

#### Diagnostic

- **Stable code** — A machine-readable identifier suitable for automation.
- **Severity and explanation** — The importance and human-readable meaning.
- **Location and context** — The relevant query text and command.
- **Category** — Syntax, unknown field, unavailable field, unsupported semantics, or schema ambiguity.

Diagnostics can belong to parsing, analysis, validation, or rewriting results.

#### Schema Target

- **Input kind** — Field list, JSON Schema, or OCSF.
- **Identity and version** — The selected schema and release.
- **Field expectations** — Allowed, required, optional, patterned, and undeclared-property behavior.
- **OCSF selection** — Activity category or event class, with any profiles and extensions.

A Schema Target combines with query analysis to produce a Validation Report.

#### Validation Report

- **Overall result** — `valid`, `invalid`, or `incomplete`.
- **Reference outcomes** — Matching, missing, optional, conditional, permitted-but-unspecified, and indeterminate fields.
- **Diagnostics** — Actionable findings with locations.
- **Coverage** — What was completely analyzed and why anything remained incomplete.

A definite error makes the result `invalid` while completeness details remain available. With no definite error, uncertainty produces `incomplete`; `valid` is reserved for complete, error-free analysis.

#### Rewrite Request and Result

- **Original query and requested mappings** — What should be transformed.
- **Rewritten query** — The source-preserving output.
- **Change audit** — Applied, skipped, and ambiguous changes.
- **Diagnostics and completeness** — The confidence and limitations of the transformation.

A Rewrite Result can be parsed, analyzed, and validated again as verification.

---

## Milestone 1 — Trustworthy Baseline

This milestone turns the current alpha into a dependable starting point. Users can install the project from a clean checkout, follow its documentation, and get consistent behavior from the existing SPL capabilities.

### What gets built

- Clean, repeatable Go, CLI, REST, and Python build and test workflows.
- CLI inputs, configuration, commands, output options, and examples that work as documented.
- Working mapping configuration rather than an accidental identity-only default.
- Accurate Python discovery and mapping results through the real native library, including macros and data-model references.
- Safe concurrent use of the public library and bindings.
- Documentation that describes only implemented APIs and clearly labels roadmap capabilities.
- Reproducible release artifacts and a maintained compatibility statement.
- A recorded performance baseline without unsupported performance claims.

### What milestone 1 explicitly does NOT include

- A replacement analysis model or field-lineage engine.
- Field-list, JSON Schema, or OCSF validation.
- Standalone SPL2 support.
- A rewrite of the grammar or existing mapping behavior beyond baseline correctness.
- New SDKs, editor integrations, or hosted services.

### Done when

A user can clone the repository, run the documented build and test commands without changing tracked dependency state, execute every documented current CLI example, and exercise the actual Python native package successfully. Documentation, packages, and reported version all agree on the functionality that exists.

---

## Milestone 2 — Structured Analysis Kernel

This milestone gives all later features one trustworthy description of a query. Existing SPL users receive precise references, scopes, lineage, diagnostics, and completeness instead of only flat name lists.

### What gets built

- A stable Query Document result available through every public surface.
- Structured stages, references, dependency summaries, source locations, and nested scopes.
- Flow-aware field lineage for creation, reads, renames, removals, aggregations, lookups, and subqueries.
- Stable diagnostic codes and explicit analysis completeness.
- Partial discoveries that remain available even when another construct is invalid or unsupported.
- Correct classification of common comparisons, functions, multiple assignments, command field lists, case variations, and derived-field reads.
- A capability manifest separating syntax support from semantic support.
- Compatibility guidance for callers moving from the existing flat discovery result.

### What milestone 2 explicitly does NOT include

- Comparing fields with an external field list or schema.
- JSON Schema or OCSF support.
- Broad standalone SPL2 parsing.
- The mapping and rewrite modernization.
- SARIF, language-server behavior, or dependency graph presentation.

### Done when

Across a representative SPL corpus, users can see every supported reference with its role, stage, scope, and location; trace derived fields through the pipeline; and understand any limitation from an explicit diagnostic. Unsupported behavior never appears as complete analysis, and existing supported SPL use cases remain available.

---

## Milestone 3 — Field-List Validation

This milestone delivers the smallest complete schema-validation workflow. Users can supply an allowed input-field catalog and determine whether a single query or query collection refers to fields outside it.

### What gets built

- Local field-list input through Go, CLI, Python, and the REST adapter.
- Validation of source-required event fields while respecting fields created earlier in the query.
- Exact and nested field-path checks plus statically resolvable field patterns.
- Separate outcomes for matching, missing, unavailable, optional-equivalent, and indeterminate references.
- `valid`, `invalid`, and `incomplete` output for interactive and automated use.
- Human-readable findings and versioned machine-readable reports.
- Individual-file, standard-input, and batch validation workflows.
- Automation-friendly outcomes that distinguish invalid queries from inconclusive analysis.

### What milestone 3 explicitly does NOT include

- JSON Schema keywords or OCSF catalogs.
- Field value or event-record validation.
- Expression type checking.
- SPL2 support beyond syntax already understood by the existing product.
- Automatic fixes or suppressions.

### Done when

A user can validate known-good, missing-field, derived-field, nested-field, wildcard, and unsupported-command examples from the CLI and both library surfaces. Definite missing fields are invalid with exact locations, uncertainty is incomplete with a reason, and a complete match is valid.

---

## Milestone 4 — JSON Schema and OCSF Validation

This milestone expands validation from simple catalogs to real-world structured schemas. Users can assess queries against local JSON Schema documents and versioned OCSF activity categories or event classes without a network connection.

### What gets built

- Local JSON Schema input with nested properties, required and optional fields, open and closed objects, field patterns, references, and composed alternatives.
- Results that distinguish declared fields, optional fields, conditionally available fields, and fields permitted without a declaration.
- Offline, versioned OCSF catalog selection.
- OCSF event-class validation with optional profiles and extensions.
- OCSF activity-category validation that separates category-wide fields from fields present only in some member classes.
- Clear diagnostics for ambiguous schema branches and category-only uncertainty.
- Consistent individual and batch reports through every public surface.
- Examples covering common security-event validation workflows.

### What milestone 4 explicitly does NOT include

- Downloading or updating schemas during validation.
- Guessing an OCSF category or class from a query.
- Checking actual events against OCSF.
- Type-checking SPL expressions against schema value types.
- Supporting proprietary schema registries.

### Done when

A user can select a local JSON Schema or OCSF version and receive correct, explainable results for required, optional, nested, open, closed, composed, class-specific, and category-conditional fields. The same fixtures yield equivalent reports in Go, CLI, Python, and REST.

---

## Milestone 5 — Broad Standalone SPL2 Support

This milestone makes standalone SPL2 a first-class input while preserving the truthfulness established earlier. The toolkit broadly understands the `splunkd` search profile and feeds both SPL and SPL2 into the same public analysis, lineage, and validation experience.

### What gets built

- Explicit SPL or SPL2 dialect selection on every public surface.
- Broad syntax coverage for standalone SPL2 searches in the `splunkd` compatibility profile, including its pipeline and SQL-style search forms.
- SPL2 identifiers, literals, expressions, command and function forms, clauses, subsearches, joins, and field lists represented in the Query Document.
- Dependency discovery, field lineage, and schema validation for the supported semantics of parsed SPL2 constructs.
- `incomplete` results for parsed constructs whose field effects are not yet modeled.
- A conformance corpus drawn from authoritative standalone SPL2 examples, negative cases, and compatibility references.
- A public capability manifest that makes command and function semantic coverage visible.
- Regression coverage preserving supported SPL behavior.

### What milestone 5 explicitly does NOT include

- SPL2 modules, imports, exports, namespaces, named module searches, or user-defined declarations.
- Edge Processor, Ingest Processor, or federated-search profiles.
- A promise of complete semantics for every successfully parsed command.
- Automatic SPL/SPL2 dialect guessing.
- Search execution or whole-query SPL1-to-SPL2 conversion.

### Done when

A user can explicitly analyze representative standalone SPL2 searches across both supported syntax styles, obtain the same structured references and schema-validation experience as SPL, and see every semantic limitation as `incomplete`. The conformance corpus clearly demonstrates broad parser coverage for the declared profile without regressing SPL.

---

## Milestone 6 — Safe Rewrite and Mapping 2.0

This milestone makes transformation a trusted consumer of the analysis model. Users can preview and apply deterministic mappings with a complete audit of what changed, what did not, and why.

### What gets built

- Mappings for fields and supported indexes, sources, source types, lookups, datasets, and data models.
- Context-aware handling of source fields, derived fields, renames, scopes, and subsearches.
- Unconditional mappings and deterministic conditions based on discovered query facts.
- Preview results containing applied, skipped, and ambiguous changes.
- Preservation of unaffected source text, quoting, comments, and formatting.
- Refusal and diagnostics for unsupported or dynamically constructed references.
- Post-rewrite parsing, analysis, and optional schema validation.
- Equivalent rewrite behavior through Go, CLI, Python, and REST.

### What milestone 6 explicitly does NOT include

- Whole-query SPL1-to-SPL2 transpilation.
- Query formatting or beautification.
- Dynamic identifier rewriting.
- Runtime-equivalence guarantees.
- Learned mappings, AI repairs, or automatic raw-to-data-model conversion.

### Done when

A user can preview and apply representative conditional and unconditional mappings across nested, derived, renamed, and quoted references; inspect an exact change audit; and receive a successfully reanalyzed result. Ambiguous changes are never silently applied.

---

## Milestone 7 — Developer Tooling and Ecosystem

This milestone turns the stable core into a practical foundation for repositories, CI systems, migration projects, and editors. Each tool remains a view over the same local analysis results rather than a separate source of truth.

### What gets built

- Query-corpus scanning with aggregate dependency, validation, completeness, and coverage summaries.
- SARIF reports suitable for code-scanning and CI consumers.
- Exportable dependency and field-lineage graph data.
- Schema- and mapping-change impact reports across a query corpus.
- Language-server diagnostics and source highlighting for compatible editors.
- A stable advanced document representation that does not expose generated parser internals.
- Published, versioned machine-output contracts and capability documentation.
- Polished cross-platform release artifacts and examples for embedding the toolkit.

### What milestone 7 explicitly does NOT include

- A graphical IDE or hosted dashboard.
- Query execution, result previews, or live Splunk catalog browsing.
- Full autocomplete or automatic bulk migration.
- Multiple editor-specific extensions before the language-server contract is proven.
- A general static-analysis platform for languages other than SPL.

### Done when

A user can scan a repository of queries, consume its findings in terminal and machine-readable forms, import standards-based diagnostics into a compatible code-review workflow, export dependency and lineage data, assess migration impact, and receive local editor diagnostics from the same analysis engine.
