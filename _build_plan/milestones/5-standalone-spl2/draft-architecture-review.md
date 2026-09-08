# M5 draft architecture review

Reviewed draft SHA256: `7b2306443030e6517aa3b4e97b7acc03fd0cf8cf654f2871d64bf42dd3ec5777`.

Scope: the proposed shared SPL/SPL2 semantic extraction, recovery/source refinement, SQL lexical and evaluation ownership, and literal dotted/nested/qualified paths. Read the approved design and relevant form-expansion obligations. Production evidence was inspected with `git show` at accepted M4 Task 2 `bacc2172291005b0c70cdbba948512173090f9aa`, analysis hook `47b560d710ed264490e6c21711ad157146b8fa78`, and M3 `e401402e5d8cd80f469884b2c3ad80245266ad66`. The analysis tree has no differences between the first two revisions; `flow.go` and `model.go` have no differences between the latter two. No production runs, tests, builds, source changes or commits were performed.

## Findings

### 1. [P2] Give shared projection a boundary for already prepared references

Plan evidence: lines 172–189 extract the substantive projection body into `applyProjection(selectors []locatedOperand, mode string, retainKnownInternals bool)`. `locatedOperand` at lines 155–160 has no prepared reference identity or selected binding. Lines 210 and 367–373 require SELECT evaluation to register reads exactly once and final projection to reuse those identities; Task 7's ownership at line 365 excludes the shared transfer file.

Immutable implementation evidence: `47b560d:pkg/analysis/commands.go:132–135` registers a new exact read whenever a selector is evaluated. The projection body at lines 272–287 both evaluates selectors and installs the selected field state, with project transitions referencing the newly returned IDs. Those are currently one operation.

Extracting that operation with the proposed input cannot directly implement the required later SELECT restriction: invoking it normally repeats the lexical reads, while manually restricting state in `spl2_sql.go` duplicates the substantive projection behavior that the plan requires to be shared. The same source name/location alone is not an explicit contract for selecting an earlier reference and binding after alias/aggregate preparation.

Minimal correction: define a private prepared-selection boundary carrying selected names/bindings and their existing reference IDs, and have ordinary projection and SQL final restriction use its common installation/transition operation. Keep frontend selector evaluation separate. Assign this work explicitly to Task 1, or extend Task 7 ownership to the shared file and require reconciliation before dispatch. Add a focused shared-boundary witness proving that final restriction adds no read references and preserves prepared source/derived provenance. No public AST or schema API expansion is needed.

### 2. [P2] Preserve per-output lookup policy in the shared extraction

Plan evidence: line 187 proposes `applyLookup(matches, outputs []locatedOperand, outputNew bool)`, while lines 172 and 248–266 require extraction without changing existing SPL reports.

Immutable implementation evidence: `47b560d:grammar/SPLParser.g4:178–180` permits ordered `analysisOutput*` blocks, each independently `OUTPUT` or `OUTPUTNEW`. `47b560d:pkg/analysis/commands.go:385–395` evaluates local match references once, then lines 400–418 process each output block with its own `OUTPUTNEW` policy and reuse the same match IDs. A legacy input such as `search user=* | lookup users user OUTPUT name OUTPUTNEW role` therefore requires two different output policies in a single transfer.

A single Boolean for all outputs loses this accepted shape. Calling the proposed operation once per block would ordinarily repeat match reads and can read an environment already changed by a preceding output, violating both the original match provenance and unchanged SPL wire contract. Leaving these effects in the SPL wrapper would also leave substantive lookup transfer outside the shared boundary.

Minimal correction: give each ordered output operand its preserve-existing policy, or evaluate matches once and pass the prepared IDs into a shared output-group operation. Add a Task 1 full-result regression for mixed `OUTPUT`/`OUTPUTNEW`, including an existing destination and an output that overlaps a match field. Preserve output order, one lexical match reference, and the accepted per-output transitions; this does not broaden admitted SPL2 syntax.

## Scoped conclusion

Two concrete extraction corrections are needed before final plan approval. The remaining reviewed architecture is consistent with the approved scope: original typed delimiter ownership remains frontend-specific; field admission, tombstones, conditional bindings and expansion evidence remain canonical; SQL stages/IDs stay lexical while positions and phase lineage express execution; pre-group and selected-output environments stay separate; hidden HAVING/ORDER and alias collisions remain incomplete. The explicit string-only schema/path limitation correctly prevents equating nested navigation, literal dotted identifiers and qualified aliases.

Pending public M4 schema report/dispatcher and package interfaces are explicit preflight gates, not findings. Controller-approved capability selectors and additive phase metadata were treated as fixed decisions. This is a draft architecture review, not implementation acceptance or production release. The root's separate whole-plan review owns its additional Task 6 sequential-assignment sample correction and the overall decision.
