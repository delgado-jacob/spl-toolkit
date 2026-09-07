## What's new in the app

- Map existing SPL queries from a configuration file using the documented CLI options.
- Install and use the native Python package with complete discovery results and explicit cleanup.
- Use the public library and bindings concurrently with consistent mapping configuration.
- Preserve caller-owned leading and trailing whitespace byte-for-byte when mapping or leaving a query unchanged.
- Build the documented release artifacts and inspect their compatibility and performance evidence.

## What was built

Milestone 1 establishes a trustworthy baseline for the existing Go library, CLI, Python/native package, and REST service. The supported surface performs syntax-aware configured field mapping, flat discovery across seven existing categories, and validation against the bundled grammar. The Python wrapper owns native results safely, verifies package/native version agreement, and supports explicit or context-manager cleanup. Mapper state is now locked for concurrent use; the existing REST request-configuration and global-fallback behavior was preserved.

The release pipeline builds reproducible CLI, server, native-library, source-distribution, and wheel artifacts for Linux x86-64, macOS x86-64, macOS arm64, and Windows x86-64. It also checks Go 1.22.12 compatibility, Go race tests, GCC 13 AddressSanitizer and leak detection, clean-source builds, documented Make targets, and Docker CLI, Python/native, and server examples.

The tested implementation is `6b55f8902ff1a990aea8951cd78934b32c90660a`. [GitHub Actions run 34077738907](https://github.com/delgado-jacob/spl-toolkit/actions/runs/34077738907) completed all 24 jobs successfully. Strict aggregation accepted 28 exact-source records. All 16 combinations of four targets with Python 3.11.9, 3.12.10, 3.13.7, and 3.14.0 passed 11 required native tests and 5 surface tests without failures or skips. The shared surface suite exercised 10 fixtures, including mapped and unchanged queries with boundary spaces, tabs, and newlines. Four reproducible target sets retained eight checksum-linked payloads each. Exact runner images, build environments, test counts, and all 32 artifact identities are in [`docs/evidence/milestone-1-acceptance.json`](../../../docs/evidence/milestone-1-acceptance.json).

Performance remains an observed baseline rather than a release threshold. The fixed-fixture measurements at the tested source `6b55f8902ff1a990aea8951cd78934b32c90660a` recorded median Parse, Discover, Map, and Parallel Map costs across five repetitions; see [`docs/performance.md`](../../../docs/performance.md) for the environment, command, ranges, allocation counts, and results.

## Decisions made during implementation

- Conditional rules run by ascending priority, retain configuration order for ties, and use only the first matching rule. Its mappings override base mappings for the same source field. This makes precedence deterministic and avoids combining multiple matching rules implicitly.
- CLI mapping requires `--config`; discovery and query validation do not. The CLI accepts `--query` or a positional query, supports text and JSON output, distinguishes rejected content from usage failures with exit codes, and writes successful `--output` results without also printing them.
- Existing REST behavior was preserved: requests may provide mapping configuration, otherwise the server uses its global fallback, and the process-global mapping endpoint remains disabled by default unless explicitly enabled for development.
- Release wheels carry their target's native library. Installed-wheel acceptance runs outside the checkout with source-path injection removed, and release artifacts are promoted only after package and reproducibility checks pass.
- Flat `InputFields` remains available for compatibility even though it cannot express lineage or read/write roles. The existing unsupported-macro fallback returns known macros and leaves the other categories empty. It has no completeness model; Milestone 2 owns that distinction.

## What the next milestone needs to know

Milestone 2 still owns structured analysis and completeness: stage-aware references, lineage, read/write roles, stronger handling of dynamic or unsupported syntax, and any expanded SPL/SPL2 grammar. `InputFields` is a compatibility view, not the structured model. Current mapping preserves surrounding query text for supported syntax; it is not a general semantic rewrite guarantee.

The public Go implementation remains canonical across the CLI, Python binding, and REST layer. Generated parser sources are committed build inputs. Reuse the exact-source evidence protocol when release inputs change, and keep accepted artifacts tied to their recorded source, environment, and checksums.

## Deviations from the PRD and why

Rule evaluation intentionally uses one matching rule with stable priority rather than merging every matching rule. The earlier behavior made precedence ambiguous and could silently combine incompatible mappings. CLI mapping now requires an explicit configuration file so a successful command cannot imply mappings that were never loaded. REST configuration behavior was retained from the baseline and is not an implementation deviation.

The first remote Task 10 run exposed that most fresh matrix interpreters did not include the checker's `packaging` dependency. The workflow was corrected to install the pinned build requirements before running the isolated package check; the inner test environment and wheel-only, no-Go installation contract remained unchanged. A later whole-branch review found that the token rewriter's final `TrimSpace` violated the documented byte-preservation boundary. The focused fix removed that blanket trim, added shared mapped and no-op boundary fixtures, and triggered replacement run 34077738907 against the corrected source. This documentation-only follow-up descends from that tested implementation and does not claim that its own commit was part of the CI run. The final scoped review of the complete fix and evidence wave is recorded separately.
