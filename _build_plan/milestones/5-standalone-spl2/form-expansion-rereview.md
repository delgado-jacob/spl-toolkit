# Milestone 5 form expansion — scoped re-review

Reviewed 2026-09-07 EDT / 2026-09-08 UTC. Scope: the P2 assembly correction, the count implicit-label clarification, and the appended research note. **The P2 is addressed. No finding remains open from the bounded expansion review, and no new inconsistency was found in these corrections.** Overall scope approval and implementation release remain with the controller.

Only this new report was written. The original review remains immutable. No product tests, parser or fixture execution, broad source re-review, web retrieval, production edit, or Git mutation was performed.

- `form-expansion.md:15` now exempts every E.L01.start candidate from prefix assembly, explicitly including invalid starts. N1 at line 1524 remains exactly `failure index=app`. It therefore tests the intended standalone-start boundary rather than becoming an unknown pipeline command. Other snippets retain their previous assembly contexts.
- `form-expansion.md:2383` now limits the proved implicit output `count` to `count()` without AS. Operand-bearing implicit labels remain U unless separately evidenced and approved; an explicit alias resolves only the output name. The existing operand-read and nested expression-coverage conditions remain in force. The explicit-alias F.count positives and arity negatives at lines 2386–2391 are unchanged.
- `research-notes.md:114–116` accurately records those two document corrections and their scope. It introduces no parser, semantic registry, fixture-execution, plan, or release claim.

Document-only verification reconstructed the prior expansion by substituting only the two original lines in memory; the result matched the previously reviewed SHA-256 `4281ef7dec370f349994af7305f5230bc2bd938465fc078df0d7c8469d5e6744`. The first 112 research-note lines likewise matched the prior hash `0bdbe54e0ae5423eba312b088152dbaea164f4b5aadab8199d8d13f84fd4538d`. This confirms that candidate blocks, IDs, holds and other obligations were not changed by the correction. The recount remains 149 E sections, 34 F sections, five EH rows, and 778 unique E/F/I candidate IDs. These are document counts, not executed or deduplicated fixture totals.

## Exact reviewed input identity

| Artifact | SHA-256 |
| --- | --- |
| form-expansion.md | `6eb5f7242f8dd8078ce9ec2daed06a374c0fec873dfb314156cc83d5cb36e2ba` |
| research-notes.md | `002b2176f662c95058ecd894bcb27f9b3df629580a3723a0597ac7fbd04166ad` |
| conformance-matrix.md | `84d07d0ba4638cff5e3e7cf021698a9d570d0b4a81f46414a670f0eb8491fe51` |

The first two hashes match the controller's requested snapshots. The matrix hash is unchanged from the expansion review. The design remains `05660c261d860eb0ed09e9dadbf30309ae82a01498e4fdfa144828205139b911`; the original expansion review is `b6765781547ad0de524fe095f19689071995a9cba071f215f33f3b5ef3e6a6b1`.

The original bounded source-review limitations still apply. Verified M4/shared-interface selection, controller release, durable fixture conversion, exact tree/source/scope/state/phase expectations, parser recovery, adapter parity and production execution remain open. This re-review does not establish that fixtures exist or pass and does not review later edits or production HEAD.
