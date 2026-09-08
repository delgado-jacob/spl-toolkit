# Milestone 5 conformance preflight — independent review

Reviewed 2026-09-07. This is a read-only design/conformance audit, with this report as its only written artifact. No parser fixtures, production tests, generation, implementation, or Splunk runtime checks were run.

## Verdict

**One P2 correction is required before adopting the matrix unchanged as TDD input.** Subject to that narrow correction, the artifact is suitable to seed the later fixture expansion and implementation process with its stated gates. It is not a completed conformance corpus, an implementation approval, or proof of runtime compatibility. The verified M4 handoff, exact shared lowering/SQL phase API, per-form expansion, and executed adapter/semantic assertions remain open as documented.

## Finding

### P2 — Keep slash-regex alternation outside unconditional positive syntax seeds

- Artifact: `conformance-matrix.md:836`, candidate **L23-P1**; its paired unterminated form is **L23-N1**, line 838.
- Candidate: `FROM main | rex field=payload /(?<tag>red|blue)/`.
- The matrix defines P as documented outer syntax that should be recognized (line 13). The cited [rex overview/usage](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-reference/rex-command/rex-command-overview-syntax-and-usage), updated 2026-06-18T09:04:43.715Z, has two relevant sections: “Pipe characters” specifically requires double quotation marks around regex containing a pipe; “Character classes and string expressions” separately demonstrates a slash-delimited regex containing `\d+`. The latter establishes slash literals but does not settle the former delimiter-specific restriction. The [built-in types page](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-manual/data-types/built-in-data-types) also shows a slash regex, without an internal pipe. Targeted primary-source searches found no corroborating slash-plus-pipe example.
- Consequence: retaining this as an unconditional P seed would silently settle an additional source conflict in the lexer/command-boundary contract. Leaving extraction semantics U does not repair the unsupported complete-outer-syntax expectation.
- Narrow correction: use an uncontested slash positive such as `FROM main | rex field=payload /(?<tag>\w+)/`, and remove its closing slash for the paired malformed negative. Preserve quoted alternation as a separately expanded positive. Record slash-plus-pipe as a held/incomplete form unless additional primary evidence or a controller ruling resolves it. Do **not** reclassify the balanced slash-plus-pipe query as definitely invalid solely from this wording conflict.

## Checks and conclusions

The complete matrix text, all 288 candidate classifications, the approved 91-line design, and the 88-line research notes were read. Targeted primary references were retrieved for the command/profile inventory, evaluation-function profile boundary, branch, FROM syntax, predicates/EXISTS, lambda bodies, arrays/objects, built-in literals, general syntax, lookup, appendpipe, union, fillnull, spath, rex, and streamstats. This was targeted source sampling, not a repeat retrieval and exhaustive audit of every ledger URL or every modifier/function arity.

1. **Inventory and scope:** the document contains 53 official inventory rows: 50 Yes and three No for splunkd, with 35 dedicated command families covering 36 native rows and 14 native deferred entries. Separate FROM/SELECT rows are not counted as two dedicated families. Deferred branch/into/thru and resource qualification retain the native/standalone distinction; no module error is inferred merely from a sink or qualified resource. The refreshed command compatibility table and branch search section support that separation. The function compatibility table corroborates B06/B07's batch-function profile exclusions.
2. **Negative classifications:** no other concrete misclassified N/B seed was found in the full candidate scan and targeted source checks. Sampled strict cases include missing/incorrect spath path quoting and output dependency, uppercase-only lookup OUTPUT, comma lists, Boolean-only options, malformed lambda bodies, and EXISTS context/LIMIT restrictions. R01/R02 remain the controlling Boolean-eval/rename rulings. H01–H11 are excluded from the negative totals; H02's same-page streamstats layout conflict is handled as incomplete. This conclusion is a document audit with sampling, not proof that every source interpretation will survive later fixture review.
3. **Counts:** an independent document-only script found 68 cards (35 C, 25 L, eight Q), 272 card seeds, 16 B seeds, **288 total**, **152 N/B**, and **56 deliberately SQL-focused** seeds (C02, L15–L19, Q01–Q08). No duplicate IDs, raw candidate source texts, or fully assembled source texts were found. Assembly applied the documented command-prefix rule and retained L04-P2's real newline. These counts are candidate counts; they do not prove 288 distinct semantic forms or 152 independently established negative contracts.
4. **Coverage honesty:** lines 11, 1019–1064, and 1182–1185 explicitly retain modifier/alternative/function-arity expansion and exact expectation assertions. The counts do not claim every option is paired or tested. Future expansion should explicitly include expression-position search literals (L24 currently discusses the distinction but seeds only command embeddings), and the documented rejection of duplicate object keys from the [object-literal reference](https://help.splunk.com/en/splunk-enterprise/search/spl2-search-manual/expressions-and-predicates/array-and-object-literals-in-expressions). These are concrete expansion reminders under the existing gate, not new claims that today's seed-only artifact is exhaustive.
5. **Architecture and partial findings:** no contradiction was found with a dedicated typed SPL2 frontend feeding the shared kernel, syntax/semantic coverage separation, lexical SQL stages with logical clause-entry positions, SELECT-owned aggregate/projection phases, or retained sound findings under incomplete coverage. Q06 explicitly avoids premature projection and fabricated hidden-field failures. Real source text, Unicode, multiline comments, alias scopes, template reads, and implicit internal-field span prohibitions provide feasible source-location requirements. Exact span correctness remains unexecuted.

## Reviewed artifact identity

The matrix had 1,187 lines at review time (including the final blank line). SHA-256:

| Artifact | SHA-256 |
| --- | --- |
| conformance-matrix.md | `f1da97889de3318e7a8a0a49dada18914082a43cfdadb0ae7b04699e6ef8e891` |
| design-spec.md | `05660c261d860eb0ed09e9dadbf30309ae82a01498e4fdfa144828205139b911` |
| research-notes.md | `0aaecc9b0d6fe20620706f54c72fee0bb769eabcf9424cebba290dedb5d814d7` |

This report does not assess any later edits or production HEAD. It makes no finding merely from absent implementation or already-declared held forms.
