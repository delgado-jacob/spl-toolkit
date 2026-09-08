# Milestone 5 broad fix 1

The scoped correction is committed as `7571f1cb578968caf92996b050ebfc3092f02605` above `fcbb74ac60da312bc06275e4e3de29b570a3bf2e`. `len` now carries numeric result evidence; `split` retains supported presence but does not claim a scalar-string result. The two module exclusion sites now emit `SPL_UNSUPPORTED_MODULE` with category `unsupported_syntax`.

The exact nested-len failure and the module-category failures were reproduced before their fixes. The 17 focused analysis queries and six field-list/JSON-Schema validation compositions preserve source reads, nullability, unknown/named/arity boundaries, and the distinction between conditional presence and analysis completeness.

All 1,714 compact canonical cases passed without fixture changes: 732 independent expectations and 982 separately reviewed snapshots remain unchanged. Full reports differ from the prior accepted transport only at five pre-authorized category leaves, B08–B12. The other 1,709 reports, including nine len/split text candidates, are identical. Full reports provide no additional conformance credit.

Fresh verification passed Go 1.22 analysis/validation tests, current-Go vet/race for those packages, native/CLI/server builds, 339 source Python tests, and 78 source acceptance tests. Both the offline-built wheel and rebuilt sdist passed 322 native and 102 acceptance tests with zero skips, copied fixtures outside the checkout, and the packaged library. The existing tarfile Python 3.14 deprecation warning remains qualified.

[verification.json](verification.json) contains exact commands, exits, source/fixture/artifact/tool hashes, checked-input maps, continuity for reused generation/OpenAPI/tooling/docs checks, and open review/root gates. [corpus-delta.json](corpus-delta.json) and [category-authority-freeze.json](category-authority-freeze.json) bind the pre-capture category amendment. Compressed reports preserve complete transport/source/installed evidence; `execution-records.tar.gz` preserves RED/GREEN raw logs, per-check runner source, and source-commit binding. Historical Task11 evidence is unchanged.

Build/test receipts retain their actual `fcbb74a` dirty-tree HEAD. Their final source blobs equal the implementation commit exactly. Scoped review and root reacceptance are still required; this record does not claim them.
