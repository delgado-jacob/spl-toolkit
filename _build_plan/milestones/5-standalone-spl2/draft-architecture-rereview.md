# M5 draft architecture scoped re-review

Actual reviewed plan SHA256: `0c7c2e787ee8ff42f2c90faeb2d339ca1a9439f8a6980b458fb38519964a5dc3`.

Scope: only the correction batch addressing the two findings in `draft-architecture-review.md`, its immediate Task 1/Task 7 integration consequences, and confirmation that the root's corrected Task 6 sample does not demand a fields-retention regression. The maintained CLI/API/documented-example ownership additions were observed but are not independently accepted by this architecture review. The original review remains unchanged. No broad re-review, source research, production tests, builds, source changes or commits were performed.

## Finding dispositions

1. **Prepared projection boundary — addressed.** Plan lines 180–184 and 193–194 define prepared binding evidence and existing reference IDs plus a shared installer. Lines 201–203 explicitly preserve copied source/derived/conditional provenance, ordered transition evidence and implicit internal-field retention without fabricated reads/transitions. The installer performs no read registration, selector evaluation or admission; ordinary projection and SELECT final restriction share it. Task 1 owns extraction and the no-new-reference/finalized-ID proof at lines 266–280. Task 7 consumes that reviewed boundary, permits only a documented narrow shared adjustment, and checks prepared aliases/aggregates, reference reuse and downstream output at lines 385–393. This closes the earlier input/ownership gap without introducing a SQL-only transfer engine.

2. **Per-output lookup policy — addressed.** Plan lines 185–188 and 198–199 replace the single command-wide Boolean with ordered per-output policy and once-prepared match IDs. Line 207 states that matches are evaluated before outputs, each output retains its own location and policy, existing OUTPUTNEW destinations preserve the accepted reference-only behavior, and overlapping outputs cannot re-read or change the match provenance. Task 1's full-result witnesses at line 268 cover mixed blocks, existing destinations and match/output overlap, including one lexical match read and ordered references/transitions. Line 280 assigns the actual extraction. The correction preserves accepted SPL behavior without broadening SPL2 syntax.

## Adjacent correction and conclusion

The Task 6 valid assignment witness now ends with `table y` at lines 358–370. Its separate `fields y` control at line 373 explicitly retains incomplete open-source internal membership and requires finite-universe/removal controls. The sample therefore no longer asks an implementer to weaken fields retention or fabricate internal fields to satisfy `Valid`.

Both findings are closed for this draft revision. No new architectural defect was identified in the correction batch or its directly affected shared-transfer/SQL boundaries. This scoped conclusion does not replace the root's whole-plan review, establish implementation correctness, or authorize execution. Final reviewed M4 interfaces and handoff, final plan approval, and explicit production release remain open gates.
