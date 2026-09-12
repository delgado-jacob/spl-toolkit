# M6 draft architecture review

Verdict: no actionable P1/P2 architectural defect found within this bounded draft review. This is not final plan approval, implementation verification or production authorization.

Reviewed inputs:

- `milestone-6-implementation-plan.md`, all 372 lines, SHA256 `e087f92156a07b46de037beae754c063640efe4b9770d26e28a647d718c6e3a5`.
- Approved `design-spec.md`, SHA256 `4355ca99ec2807025dafda26cff66b6768287f004a6f9930a0482c066063adba`.
- Only the narrowly relevant validation files from immutable core `5e3d4ec3a54b6878830b83a8ff08405011a507d9`: `pkg/validation/schema_target.go`, `schema_batch.go`, `batch.go` and `catalog.go`.

Checks and evidence:

| Concern | Bounded conclusion and plan evidence |
| --- | --- |
| Strict requests and target composition | Lines 158–178 define independent single/batch/rules-file wrappers, strict presence and union checks, exact scalar handling, direct-Go preparation and reuse of existing target decoders. Lines 162 and 240 retain canonical destination validation reports and avoid testing old source names against the destination target. No competing target format or validation engine is prescribed. |
| Canonical facts and frontend access | Lines 129–143 explicitly gate actual helper access, supported forms, phases, capability and adapter/package details on accepted M5. Lines 186–196 place evidence, literal decoding, scope propagation and rendering in canonical frontend/transfer ownership. Lines 202–210 limit rewrite-side conditions to pure three-valued evaluation over that evidence. No second parser or transfer engine is prescribed. |
| Atomic linked groups and per-point conditions | Lines 216–220 require the smallest necessary linked set, preserve explicit aliases, close implicit-name consumers and authorize every required edit at its own flow point. Relevant dynamic/unknown linkage refuses the complete group; unrelated scopes remain separate. This matches design lines 64–72. |
| Surviving conflicts | Lines 222 and 362 require deterministic conflict closure and reevaluation of dependent safety after exclusions. They prohibit priority winners and subset optimization. A surviving edit cannot depend on an excluded edit to prevent capture, matching design lines 84–90. |
| Original/candidate correspondence | Lines 224 and 236 require original-byte edit translation, canonical candidate coordinates and comparison of affected identity, role, scope, provenance, origin, transition and SQL-phase relationships. Diagnostic counts are explicitly insufficient. The exact M5 correspondence interface remains an acknowledged preflight item at line 135, rather than an invented API. |
| Preview/apply and rollback | Lines 234–240 retain original authoritative preview text, distinguish candidate inclusion from commitment, recheck unchanged candidates when a target is supplied and roll back all included apply edits after failed proof. Lines 38, 242 and 350 preserve incomplete partial-group results only where independently safe groups pass the gate. They do not authorize a second reduced candidate after verification failure. |
| Batch validation and publication | Lines 92 and 242 call the existing public validator once with the complete ordered candidate slice and delay publication until all input/internal operations succeed. The immutable `ValidateSchemaBatch` prepares one target and discards the report on any input/internal error; `ValidateBatch` likewise normalizes one catalog and preserves document order. `DecodeSchemaTarget` is strict decoding rather than schema preparation. The proposed composition therefore supports the stated once-per-batch boundary without exposing private schema internals. |
| Task and proof ownership | Tasks 1–5 proceed from contracts to canonical evidence, pure conditions, groups/edits and verification; Tasks 6–8 add transports, native/package closure and full corpus/parity. Lines 272–290 explicitly require source-native, installed wheel and wheel-from-sdist evidence, mandatory nonzero suites and zero skips. Lines 296–300 reserve serial implementation, immutable review and distinct root acceptance windows. No dependency requires adapters to invent semantics or treats source-native results as installed proof. |

Limitations: the five M5 reconciliation items remain pending by design and were not treated as defects. Their eventual evidence, feasibility and precise signatures require review against the accepted M5 commit before final plan approval. This review did not inspect active M4 adapter source or root's external oracle cases; it ran no tests, builds, generators, fixtures or runtime experiments. It modified only this report and preserved both reviewed inputs unchanged. Any claimed runtime behavior or final surface/package parity remains unverified.
