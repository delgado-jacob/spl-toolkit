package impact

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/corpus"
	"github.com/delgado-jacob/spl-toolkit/pkg/corpusio"
	"github.com/delgado-jacob/spl-toolkit/pkg/rewrite"
	"github.com/delgado-jacob/spl-toolkit/pkg/validation"
)

func TestImpactStrictInlineRequests(t *testing.T) {
	goodSchema := []byte(`{"schema_version":1,"documents":[{"id":"q","document":{"text":"table host"}}],"before_target":{"kind":"field_list","catalog":{"fields":["host"]}},"after_target":{"kind":"field_list","catalog":{"fields":[]}}}`)
	r, err := DecodeSchemaRequest(goodSchema)
	if err != nil || len(r.Input.Entries) != 1 {
		t.Fatalf("valid inline request: %+v %v", r, err)
	}
	if _, err := CompareSchemas(r); err != nil {
		t.Fatal(err)
	}
	for _, raw := range [][]byte{
		[]byte(`{"schema_version":1,"schema_version":1,"documents":[],"before_target":{},"after_target":{}}`),
		[]byte(`{"schema_version":1,"documents":[],"before_target":{},"after_target":{},"path":"q.spl"}`),
		[]byte(`{"schema_version":1,"documents":[{"id":"q","document":{"text":"x"},"failure":{"code":"x"}}],"before_target":{},"after_target":{}}`),
		[]byte(`{"schema_version":1,"documents":null,"before_target":{},"after_target":{}}`),
	} {
		if _, err := DecodeSchemaRequest(raw); !IsInputError(err) {
			t.Fatalf("malformed schema request accepted: %s, %v", raw, err)
		}
	}
	goodMapping := []byte(`{"schema_version":1,"documents":[{"id":"q","document":{"text":"table src"}}],"before_rules":{"schema_version":1,"rules":[]},"after_rules":{"schema_version":1,"rules":[]}}`)
	m, err := DecodeMappingRequest(goodMapping)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := CompareMappings(m); err != nil {
		t.Fatal(err)
	}
	if _, err := DecodeMappingRequest([]byte(`{"schema_version":1,"documents":[{"id":"q","document":{"text":"table src"}}],"before_rules":{"schema_version":1,"rules":[]},"after_rules":{"schema_version":1,"rules":[]},"before_target":null}`)); !IsInputError(err) {
		t.Fatalf("null target accepted: %v", err)
	}
}

func TestImpactTargetDigestAndAllFailures(t *testing.T) {
	before := fieldTarget("host")
	after := fieldTarget("other")
	before.Catalog.Identity, after.Catalog.Identity = "same", "same"
	p, err := PrepareSchemas(before, after)
	if err != nil {
		t.Fatal(err)
	}
	input := corpus.Input{Selection: corpus.Selection{Mode: "manifest", Complete: true}, Entries: []corpus.Entry{{ID: "failed", Origin: corpus.Origin{Kind: "file", RelativePath: "failed.spl"}, Failure: &corpus.AcquisitionError{Code: "not_found", Phase: "open"}}}}
	r, err := p.Compare(input)
	if err != nil {
		t.Fatal(err)
	}
	if r.BeforeDigest == r.AfterDigest || r.ExecutionComplete || r.Counts.Failed != 1 || r.Counts.Compared != 0 || r.Entries[0].Failure == nil {
		t.Fatalf("target content or failure lost: %+v", r)
	}
	mp, err := PrepareMappings(MappingSide{Rules: rewrite.RuleSet{SchemaVersion: 1}, ValidationTarget: &before}, MappingSide{Rules: rewrite.RuleSet{SchemaVersion: 1}, ValidationTarget: &after})
	if err != nil {
		t.Fatal(err)
	}
	mr, err := mp.Compare(input)
	if err != nil || mr.Counts.Failed != 1 || mr.Counts.Compared != 0 || mr.BeforeDigest == mr.AfterDigest {
		t.Fatalf("zero-success mapping: %+v, %v", mr, err)
	}
	bad := fieldTarget("host", "host")
	if p, err := PrepareMappings(MappingSide{Rules: rewrite.RuleSet{SchemaVersion: 1}}, MappingSide{Rules: rewrite.RuleSet{SchemaVersion: 1}, ValidationTarget: &bad}); p != nil || !IsInputError(err) {
		t.Fatalf("invalid target escaped preflight: %+v, %v", p, err)
	}
}

func TestImpactTraversalFailureDenominator(t *testing.T) {
	p, err := PrepareSchemas(fieldTarget(), fieldTarget())
	if err != nil {
		t.Fatal(err)
	}
	input := corpus.Input{Selection: corpus.Selection{Mode: "directory", Complete: false, TraversalFailures: []corpus.AcquisitionError{{Code: "traversal_failed", Phase: "traverse", Path: "sub"}}}}
	r, err := p.Compare(input)
	if err != nil {
		t.Fatal(err)
	}
	if r.ExecutionComplete || r.Counts.Selected != 0 || r.Counts.Compared != 0 || r.Counts.Denominator != 0 || r.Counts.TraversalFailed != 1 || len(r.Entries) != 0 {
		t.Fatalf("traversal denominator understated: %+v", r)
	}
}

func TestImpactProgrammaticInputErrorsAreClassified(t *testing.T) {
	input := corpus.Input{Selection: corpus.Selection{Mode: "inline", Complete: true}, Entries: []corpus.Entry{{ID: "duplicate", Document: &analysis.QueryDocument{Text: "table src"}}, {ID: "duplicate", Document: &analysis.QueryDocument{Text: "table src"}}}}
	sp, err := PrepareSchemas(fieldTarget(), fieldTarget())
	if err != nil {
		t.Fatal(err)
	}
	if report, err := sp.Compare(input); report != nil || !IsInputError(err) {
		t.Fatalf("schema input error classification: %+v %v", report, err)
	}
	mp, err := PrepareMappings(MappingSide{Rules: rewrite.RuleSet{SchemaVersion: 1}}, MappingSide{Rules: rewrite.RuleSet{SchemaVersion: 1}})
	if err != nil {
		t.Fatal(err)
	}
	if report, err := mp.Compare(input); report != nil || !IsInputError(err) {
		t.Fatalf("mapping input error classification: %+v %v", report, err)
	}
}

func TestImpactUnalignedReferencesStayUnmatched(t *testing.T) {
	b := &analysis.Result{References: []analysis.Reference{{ID: "same", Kind: "field", Role: "read", ScopeID: "scope-a", Location: analysis.Location{Start: analysis.Position{Offset: 3}, End: analysis.Position{Offset: 6}}}}}
	a := &analysis.Result{References: []analysis.Reference{{ID: "same", Kind: "field", Role: "read", ScopeID: "scope-b", Location: analysis.Location{Start: analysis.Position{Offset: 4}, End: analysis.Position{Offset: 7}}}}}
	alignment := alignOriginal(b, a, "revision", "revision")
	if len(alignment.Pairs) != 0 || len(alignment.Unmatched) != 2 {
		t.Fatalf("same ID with changed evidence paired: %+v", alignment)
	}
	deltas := outcomeDeltas([]outcomeEvidence{{"same", validation.ReferenceOutcome{ReferenceID: "same", Outcome: "matching"}}}, []outcomeEvidence{{"same", validation.ReferenceOutcome{ReferenceID: "same", Outcome: "missing"}}}, alignment, "original")
	if len(deltas) != 2 || deltas[0].Change != "unmatched" || deltas[1].Change != "unmatched" {
		t.Fatalf("unaligned same-ID outcomes falsely paired: %+v", deltas)
	}
	beforeDiagnostic := analysis.Diagnostic{Code: "SPL_UNKNOWN_FIELD", Category: "unknown_field", Location: b.References[0].Location, ScopeID: "scope-a"}
	afterDiagnostic := analysis.Diagnostic{Code: "SPL_UNKNOWN_FIELD", Category: "unknown_field", Location: a.References[0].Location, ScopeID: "scope-b"}
	deltas = diagnosticDeltas([]analysis.Diagnostic{beforeDiagnostic}, []analysis.Diagnostic{afterDiagnostic}, b, a, alignment, "original")
	if len(deltas) != 2 || deltas[0].Change != "unmatched" || deltas[1].Change != "unmatched" {
		t.Fatalf("unaligned same-ID diagnostics falsely paired: %+v", deltas)
	}
}

func TestImpactEmptyCollectionsAreArrays(t *testing.T) {
	p, err := PrepareSchemas(fieldTarget(), fieldTarget())
	if err != nil {
		t.Fatal(err)
	}
	r, err := p.Compare(impactInput("table host"))
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	var top map[string]json.RawMessage
	if err := json.Unmarshal(raw, &top); err != nil {
		t.Fatal(err)
	}
	if string(top["entries"]) == "null" {
		t.Fatalf("null entries: %s", raw)
	}
	var entries []map[string]json.RawMessage
	if err := json.Unmarshal(top["entries"], &entries); err != nil {
		t.Fatal(err)
	}
	entry := entries[0]
	for _, key := range []string{"deltas", "reasons"} {
		if string(entry[key]) == "null" {
			t.Fatalf("null %s: %s", key, raw)
		}
	}
}

func TestAmbiguousAlignmentWithoutOutcomeDeltaIsIndeterminate(t *testing.T) {
	classification, reasons := classify(nil, true, false, Alignment{Ambiguous: []string{"before:ref-0"}})
	if classification != Indeterminate || len(reasons) == 0 {
		t.Fatalf("ambiguous evidence called unchanged: %s %v", classification, reasons)
	}
}

func TestCandidateAlignmentRequiresCandidateRevision(t *testing.T) {
	ref := analysis.Reference{ID: "ref-0", Kind: "field", Role: "read", ScopeID: "scope-0", StageID: "stage-0", Location: analysis.Location{Start: analysis.Position{Offset: 6}, End: analysis.Position{Offset: 9}}}
	before := &rewrite.Result{OriginalText: "table src", CandidateText: "table src", OriginalAnalysis: &analysis.Result{}, CandidateAnalysis: &analysis.Result{Document: analysis.QueryDocument{Text: "table src", Language: "spl"}, References: []analysis.Reference{ref}}}
	after := &rewrite.Result{OriginalText: "table src", CandidateText: "table src", OriginalAnalysis: &analysis.Result{}, CandidateAnalysis: &analysis.Result{Document: analysis.QueryDocument{Text: "table src", Language: "spl2"}, References: []analysis.Reference{ref}}}
	alignment := alignMapping(before, after, "original-revision")
	for _, pair := range alignment.Pairs {
		if pair.Domain == "candidate" {
			t.Fatalf("candidate revision mismatch paired: %+v", alignment)
		}
	}
	if len(alignment.Unmatched) == 0 {
		t.Fatalf("candidate revision mismatch hidden: %+v", alignment)
	}
}

func TestCandidateProvenanceConflictNeverRecoversByOrder(t *testing.T) {
	report := &rewrite.Result{OriginalText: "table src", CandidateText: "table first", OriginalAnalysis: &analysis.Result{References: []analysis.Reference{{ID: "original"}}}, CandidateAnalysis: &analysis.Result{References: []analysis.Reference{{ID: "first"}, {ID: "second"}}}, Changes: []rewrite.Change{
		{OriginalReferenceIDs: []string{"original"}, CandidateReferenceIDs: []string{"first"}, CandidateApplied: true},
		{OriginalReferenceIDs: []string{"original"}, CandidateReferenceIDs: []string{"second"}, CandidateApplied: true},
		{OriginalReferenceIDs: []string{"original"}, CandidateReferenceIDs: []string{"first"}, CandidateApplied: true},
	}}
	if got := provenCandidateOrigins(report); len(got) != 0 {
		t.Fatalf("conflicting provenance resolved by audit order: %+v", got)
	}
}

func TestCandidateManyToOneProvenanceIsAmbiguous(t *testing.T) {
	original := &analysis.Result{Document: analysis.QueryDocument{Text: "table a b"}, References: []analysis.Reference{{ID: "o1", Kind: "field", Role: "read", Location: analysis.Location{Start: analysis.Position{Offset: 6}, End: analysis.Position{Offset: 7}}}, {ID: "o2", Kind: "field", Role: "read", Location: analysis.Location{Start: analysis.Position{Offset: 8}, End: analysis.Position{Offset: 9}}}}}
	before := &rewrite.Result{OriginalText: "table a b", CandidateText: "table a b", OriginalAnalysis: original, CandidateAnalysis: original}
	after := &rewrite.Result{OriginalText: "table a b", CandidateText: "table c", OriginalAnalysis: original, CandidateAnalysis: &analysis.Result{Document: analysis.QueryDocument{Text: "table c"}, References: []analysis.Reference{{ID: "dest"}}}, Changes: []rewrite.Change{{GroupID: "g1", OriginalReferenceIDs: []string{"o1"}, CandidateReferenceIDs: []string{"dest"}, CandidateApplied: true}, {GroupID: "g2", OriginalReferenceIDs: []string{"o2"}, CandidateReferenceIDs: []string{"dest"}, CandidateApplied: true}}}
	alignment := alignMapping(before, after, "same-revision")
	for _, pair := range alignment.Pairs {
		if pair.Domain == "candidate" {
			t.Fatalf("many-to-one candidate correspondence paired: %+v", alignment)
		}
	}
	if len(alignment.Ambiguous) == 0 {
		t.Fatalf("many-to-one candidate correspondence not marked ambiguous: %+v", alignment)
	}
}

func TestImpactDurableWireCases(t *testing.T) {
	raw, err := os.ReadFile("../../testdata/tooling/impact-cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases map[string]struct {
		Request             json.RawMessage `json:"request"`
		Classification      Classification  `json:"classification"`
		BeforeStatus        analysis.Status `json:"before_status"`
		AfterStatus         analysis.Status `json:"after_status"`
		BeforeCandidateText string          `json:"before_candidate_text"`
		AfterCandidateText  string          `json:"after_candidate_text"`
	}
	if err := json.Unmarshal(raw, &cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) != 3 {
		t.Fatalf("fixture case count: %d", len(cases))
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			var report *Report
			var err error
			if name == "schema_same_status_changed_outcome" {
				request, e := DecodeSchemaRequest(tc.Request)
				if e != nil {
					t.Fatal(e)
				}
				report, err = CompareSchemas(request)
			} else {
				request, e := DecodeMappingRequest(tc.Request)
				if e != nil {
					t.Fatal(e)
				}
				report, err = CompareMappings(request)
			}
			if err != nil {
				t.Fatal(err)
			}
			entry := report.Entries[0]
			if entry.Classification != tc.Classification || (tc.BeforeStatus != "" && entry.BeforeStatus != tc.BeforeStatus) || (tc.AfterStatus != "" && entry.AfterStatus != tc.AfterStatus) {
				t.Fatalf("wire case classification/status: %+v", entry)
			}
			if tc.BeforeCandidateText != "" && (entry.BeforeRewrite.CandidateText != tc.BeforeCandidateText || entry.AfterRewrite.CandidateText != tc.AfterCandidateText) {
				t.Fatalf("wire case candidate text: %+v", entry)
			}
		})
	}
}

func impactInput(text string) corpus.Input {
	return corpus.Input{Selection: corpus.Selection{Mode: "inline", Complete: true}, Entries: []corpus.Entry{{ID: "q", Origin: corpus.Origin{Kind: "inline"}, Document: &analysis.QueryDocument{Text: text}}}}
}

func fieldTarget(fields ...string) corpus.ValidationTarget {
	return corpus.ValidationTarget{Kind: "field_list", Catalog: &validation.FieldCatalog{Fields: fields}}
}

func fieldRule(id, source, target string) rewrite.Rule {
	return rewrite.Rule{ID: id, Kind: "field", Source: rewrite.Identity{Name: &source}, Target: rewrite.Identity{Name: &target}}
}

func TestSchemaSameStatusChangedOutcome(t *testing.T) {
	p, err := PrepareSchemas(fieldTarget("host"), fieldTarget())
	if err != nil {
		t.Fatal(err)
	}
	r, err := p.Compare(impactInput("search host=x missing=y"))
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Entries) != 1 || r.Entries[0].Before == nil || r.Entries[0].After == nil {
		t.Fatalf("canonical sides absent: %+v", r)
	}
	e := r.Entries[0]
	if e.Before.FieldValidation.Status != analysis.Invalid || e.After.FieldValidation.Status != analysis.Invalid || e.Classification != Affected {
		t.Fatalf("same-status changed outcome hidden: %+v", e)
	}
	var hostChanged bool
	for _, d := range e.Deltas {
		if d.Category == "outcome" && d.Change == "changed" {
			hostChanged = true
		}
	}
	if !hostChanged || r.Counts.Affected != 1 || r.Counts.Denominator != 1 {
		t.Fatalf("outcome delta/counts absent: %+v", r)
	}
}

func TestMappingIdentityBaseline(t *testing.T) {
	p, err := PrepareMappings(MappingSide{Rules: rewrite.RuleSet{SchemaVersion: 1, Rules: []rewrite.Rule{}}}, MappingSide{Rules: rewrite.RuleSet{SchemaVersion: 1, Rules: []rewrite.Rule{fieldRule("rename", "src", "user")}}})
	if err != nil {
		t.Fatal(err)
	}
	r, err := p.Compare(impactInput("table src"))
	if err != nil {
		t.Fatal(err)
	}
	e := r.Entries[0]
	if e.BeforeRewrite == nil || e.AfterRewrite == nil || e.BeforeRewrite.CandidateText != "table src" || e.AfterRewrite.CandidateText != "table user" || e.AfterRewrite.Text != "table src" || e.AfterRewrite.Committed || e.Classification != Affected {
		t.Fatalf("identity baseline or preview policy lost: %+v", e)
	}
}

func TestMappingIncompleteCandidateObservedValidationRegression(t *testing.T) {
	for _, tc := range []struct {
		query        string
		beforeStatus analysis.Status
	}{
		{"search host=x missing=y | mystery", analysis.Invalid},
		{"table host | mystery", analysis.Incomplete},
	} {
		t.Run(tc.query, func(t *testing.T) {
			request, err := DecodeMappingRequest([]byte(fmt.Sprintf(`{"schema_version":1,"documents":[{"id":"q","document":{"text":%q}}],"before_rules":{"schema_version":1,"rules":[]},"after_rules":{"schema_version":1,"rules":[]},"before_target":{"kind":"field_list","catalog":{"fields":["host"]}},"after_target":{"kind":"field_list","catalog":{"fields":[]}}}`, tc.query)))
			if err != nil {
				t.Fatal(err)
			}
			report, err := CompareMappings(request)
			if err != nil {
				t.Fatal(err)
			}
			e := report.Entries[0]
			if e.BeforeStatus != tc.beforeStatus || e.AfterStatus != analysis.Invalid || e.BeforeRewrite.CandidateText != tc.query || e.AfterRewrite.CandidateText != tc.query {
				t.Fatalf("canonical regression setup: before=%s after=%s candidates=%q/%q", e.BeforeStatus, e.AfterStatus, e.BeforeRewrite.CandidateText, e.AfterRewrite.CandidateText)
			}
			var hostOutcome, hostDiagnostic bool
			for _, d := range e.Deltas {
				hostOutcome = hostOutcome || (d.Category == "outcome" && d.Change == "changed" && d.Key == "ref-0")
				hostDiagnostic = hostDiagnostic || (d.Category == "diagnostic" && d.Change == "introduced" && strings.Contains(d.Key, "SPL_UNKNOWN_FIELD") && strings.HasSuffix(d.Key, "|ref-0"))
			}
			if !hostOutcome || !hostDiagnostic || e.Classification != Affected || report.Counts.Affected != 1 || report.Counts.Indeterminate != 0 {
				t.Fatalf("observed host regression hidden: classification=%s counts=%+v hostOutcome=%v hostDiagnostic=%v", e.Classification, report.Counts, hostOutcome, hostDiagnostic)
			}
			if e.BeforeRewrite.Coverage.SemanticComplete || e.AfterRewrite.Coverage.SemanticComplete || len(e.Alignment.Unmatched) == 0 && report.Counts.Unmatched == 0 {
				t.Fatal("incomplete background evidence was lost")
			}
		})
	}
}

func TestMappingIdenticalRuleMultipleOccurrencesUnchanged(t *testing.T) {
	for _, query := range []string{"search src=x | table src", "table src src"} {
		t.Run(query, func(t *testing.T) {
			rules := `{"schema_version":1,"rules":[{"id":"r","kind":"field","source":{"name":"src"},"target":{"name":"user"}}]}`
			request, err := DecodeMappingRequest([]byte(fmt.Sprintf(`{"schema_version":1,"documents":[{"id":"q","document":{"text":%q}}],"before_rules":%s,"after_rules":%s}`, query, rules, rules)))
			if err != nil {
				t.Fatal(err)
			}
			report, err := CompareMappings(request)
			if err != nil {
				t.Fatal(err)
			}
			e := report.Entries[0]
			if e.BeforeStatus != analysis.Valid || e.AfterStatus != analysis.Valid || len(e.BeforeRewrite.RuleEvaluations) != 2 || len(e.AfterRewrite.RuleEvaluations) != 2 {
				t.Fatal("expected two complete canonical rule occurrences")
			}
			if e.Classification != Unchanged || len(e.Deltas) != 0 || report.Counts.Ambiguous != 0 || report.Counts.Unchanged != 1 {
				t.Fatalf("equal occurrences invented ambiguity: classification=%s counts=%+v", e.Classification, report.Counts)
			}
		})
	}
}

func TestMappingRuleEvaluationDeltaIsPerOccurrence(t *testing.T) {
	first := analysis.Location{Start: analysis.Position{Offset: 6}, End: analysis.Position{Offset: 9}}
	second := analysis.Location{Start: analysis.Position{Offset: 10}, End: analysis.Position{Offset: 13}}
	before := []rewrite.RuleEvaluation{
		{RuleID: "r", Outcome: "proposed", Reason: "matched", ReferenceIDs: []string{"ref-0"}, Location: &first},
		{RuleID: "r", Outcome: "proposed", Reason: "matched", ReferenceIDs: []string{"ref-1"}, Location: &second},
	}
	after := append([]rewrite.RuleEvaluation{}, before...)
	after[1].Outcome, after[1].Reason = "skipped", rewrite.ReasonConditionFalse
	deltas := compareEvidence(evidence{mapping: true, rules: before}, evidence{mapping: true, rules: after}, Alignment{Pairs: []ReferencePair{{BeforeID: "ref-0", AfterID: "ref-0", Domain: "original"}, {BeforeID: "ref-1", AfterID: "ref-1", Domain: "original"}}})
	if len(deltas) != 1 || deltas[0].Category != "rule_evaluation" || deltas[0].Change != "changed" {
		t.Fatalf("per-occurrence change lost: %s", canonicalRaw(deltas))
	}
	var b, a rewrite.RuleEvaluation
	if err := json.Unmarshal(deltas[0].Before, &b); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(deltas[0].After, &a); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(b.ReferenceIDs, []string{"ref-1"}) || !reflect.DeepEqual(a.ReferenceIDs, []string{"ref-1"}) || b.Outcome != "proposed" || a.Outcome != "skipped" || a.Location == nil || *a.Location != second {
		t.Fatalf("wrong occurrence paired: before=%+v after=%+v", b, a)
	}
}

func TestMappingUnresolvedAlternativesRetainObservedValidationChange(t *testing.T) {
	beforeTarget, afterTarget := fieldTarget("host"), fieldTarget()
	p, err := PrepareMappings(
		MappingSide{Rules: rewrite.RuleSet{SchemaVersion: 1}, ValidationTarget: &beforeTarget},
		MappingSide{Rules: rewrite.RuleSet{SchemaVersion: 1, Rules: []rewrite.Rule{fieldRule("a", "host", "user"), fieldRule("b", "host", "owner")}}, ValidationTarget: &afterTarget},
	)
	if err != nil {
		t.Fatal(err)
	}
	report, err := p.Compare(impactInput("table host"))
	if err != nil {
		t.Fatal(err)
	}
	e := report.Entries[0]
	if e.BeforeRewrite.CandidateText != "table host" || e.AfterRewrite.CandidateText != "table host" || e.AfterRewrite.Coverage.RewriteComplete || len(e.Alignment.Ambiguous) == 0 {
		t.Fatal("expected unresolved alternatives with equal candidate text")
	}
	if e.Classification != Affected || report.Counts.Affected != 1 || e.BeforeStatus != analysis.Valid || e.AfterStatus != analysis.Invalid {
		t.Fatalf("validation regression masked by alternatives: classification=%s statuses=%s/%s counts=%+v", e.Classification, e.BeforeStatus, e.AfterStatus, report.Counts)
	}
}

func TestMappingRuleEvaluationDuplicateAndUnalignedEvidence(t *testing.T) {
	loc := analysis.Location{Start: analysis.Position{Offset: 6}, End: analysis.Position{Offset: 9}}
	r := rewrite.RuleEvaluation{RuleID: "r", Outcome: "proposed", Reason: "matched", ReferenceIDs: []string{"ref-0"}, Location: &loc}
	deltas := compareEvidence(evidence{mapping: true, rules: []rewrite.RuleEvaluation{r, r}}, evidence{mapping: true, rules: []rewrite.RuleEvaluation{r}}, Alignment{Pairs: []ReferencePair{{BeforeID: "ref-0", AfterID: "ref-0", Domain: "original"}}})
	if len(deltas) != 3 {
		t.Fatalf("duplicate occurrence evidence lost: %s", canonicalRaw(deltas))
	}
	for _, d := range deltas {
		if d.Category != "rule_evaluation" || d.Change != "ambiguous" {
			t.Fatalf("duplicate occurrence falsely paired: %s", canonicalRaw(deltas))
		}
	}
	deltas = compareEvidence(evidence{mapping: true, rules: []rewrite.RuleEvaluation{r}}, evidence{mapping: true, rules: []rewrite.RuleEvaluation{r}}, Alignment{})
	if len(deltas) != 2 {
		t.Fatalf("unaligned evaluation disappeared: %s", canonicalRaw(deltas))
	}
	for _, d := range deltas {
		if d.Change != "unmatched" {
			t.Fatalf("unaligned evaluation paired: %s", canonicalRaw(deltas))
		}
	}
}

func TestOriginalFallbackRequiresMutualUniqueness(t *testing.T) {
	b1 := analysis.Reference{ID: "before-1", Kind: "field", Role: "read", ScopeID: "scope-0", StageID: "stage-0", OriginalName: "host", NormalizedName: "host", Location: analysis.Location{Start: analysis.Position{Offset: 6}, End: analysis.Position{Offset: 10}}}
	b2, a1 := b1, b1
	b2.ID, a1.ID = "before-2", "after-1"
	for _, refs := range [][]analysis.Reference{{b1, b2}, {b2, b1}} {
		for _, reverseSides := range []bool{false, true} {
			before, after := refs, []analysis.Reference{a1}
			if reverseSides {
				before, after = after, before
			}
			got := alignOriginal(&analysis.Result{References: before}, &analysis.Result{References: after}, "rev", "rev")
			if len(got.Pairs) != 0 || len(got.Ambiguous) != 3 {
				t.Errorf("duplicate fallback paired or hid ambiguity (first=%s reverse=%v): %+v", refs[0].ID, reverseSides, got)
			}
		}
	}
}

func TestOriginalAlignmentReservesExactIDsBeforeFallback(t *testing.T) {
	b := analysis.Reference{ID: "fallback", Kind: "field", Role: "read"}
	exact, after := b, b
	exact.ID, after.ID = "exact", "exact"
	for _, refs := range [][]analysis.Reference{{b, exact}, {exact, b}} {
		got := alignOriginal(&analysis.Result{References: refs}, &analysis.Result{References: []analysis.Reference{after}}, "rev", "rev")
		want := []ReferencePair{{BeforeID: "exact", AfterID: "exact", Domain: "original", Basis: "canonical_original_id"}}
		if !reflect.DeepEqual(got.Pairs, want) || !reflect.DeepEqual(got.Unmatched, []string{"before:fallback"}) {
			t.Errorf("fallback consumed canonical match: %+v", got)
		}
	}
}

func TestMappingIdenticalCandidateDoesNotDuplicateAlignment(t *testing.T) {
	side := MappingSide{Rules: rewrite.RuleSet{SchemaVersion: 1}}
	p, err := PrepareMappings(side, side)
	if err != nil {
		t.Fatal(err)
	}
	r, err := p.Compare(impactInput("table src"))
	if err != nil {
		t.Fatal(err)
	}
	e := r.Entries[0]
	if e.Classification != Unchanged {
		t.Fatalf("equal complete mapping evidence: %+v", e)
	}
	seen := map[string]bool{}
	for _, pair := range e.Alignment.Pairs {
		if pair.Domain != "original" && pair.Domain != "candidate" {
			t.Fatalf("alignment has no domain: %+v", pair)
		}
		key := pair.Domain + "\x00" + pair.BeforeID + "\x00" + pair.AfterID
		if seen[key] {
			t.Fatalf("duplicate reference alignment: %+v", e.Alignment)
		}
		seen[key] = true
	}
}

func TestMappingPreservesCandidateValidationDeltas(t *testing.T) {
	target := fieldTarget("src")
	p, err := PrepareMappings(MappingSide{Rules: rewrite.RuleSet{SchemaVersion: 1}, ValidationTarget: &target}, MappingSide{Rules: rewrite.RuleSet{SchemaVersion: 1, Rules: []rewrite.Rule{fieldRule("rename", "src", "user")}}, ValidationTarget: &target})
	if err != nil {
		t.Fatal(err)
	}
	r, err := p.Compare(impactInput("table src"))
	if err != nil {
		t.Fatal(err)
	}
	e := r.Entries[0]
	if e.BeforeRewrite.CandidateValidation == nil || e.AfterRewrite.CandidateValidation == nil || e.BeforeRewrite.CandidateValidation.FieldList == nil || e.AfterRewrite.CandidateValidation.FieldList == nil || e.BeforeRewrite.CandidateValidation.FieldList.Status != analysis.Valid || e.AfterRewrite.CandidateValidation.FieldList.Status != analysis.Invalid {
		t.Fatalf("canonical validation evidence missing: %+v", e)
	}
	var changedOutcome, diagnostic bool
	for _, delta := range e.Deltas {
		if delta.Category == "outcome" {
			changedOutcome = true
		}
		if delta.Category == "diagnostic" {
			diagnostic = true
		}
	}
	if !changedOutcome || !diagnostic || e.Classification != Affected {
		t.Fatalf("candidate validation delta hidden: %+v", e)
	}
	var provenance bool
	for _, pair := range e.Alignment.Pairs {
		if pair.Domain == "candidate" && pair.Basis == "audit_original_reference_provenance" {
			provenance = true
		}
		if pair.Domain == "candidate" && pair.Basis != "audit_original_reference_provenance" {
			t.Fatalf("shifted candidate aligned without audit: %+v", pair)
		}
	}
	if !provenance {
		t.Fatalf("M6 candidate provenance not carried: %+v", e.Alignment)
	}
}

func TestShiftedUnprovenCandidateReferenceRemainsUnmatched(t *testing.T) {
	p, err := PrepareMappings(MappingSide{Rules: rewrite.RuleSet{SchemaVersion: 1}}, MappingSide{Rules: rewrite.RuleSet{SchemaVersion: 1, Rules: []rewrite.Rule{fieldRule("rename", "src", "longer_name")}}})
	if err != nil {
		t.Fatal(err)
	}
	r, err := p.Compare(impactInput("table src other"))
	if err != nil {
		t.Fatal(err)
	}
	e := r.Entries[0]
	if e.AfterRewrite.CandidateText != "table longer_name other" {
		t.Fatalf("candidate was not shifted: %q", e.AfterRewrite.CandidateText)
	}
	var unmatchedCandidate bool
	for _, value := range e.Alignment.Unmatched {
		if strings.HasPrefix(value, "candidate:") {
			unmatchedCandidate = true
		}
	}
	if !unmatchedCandidate {
		t.Fatalf("unproven shifted candidate reference hidden: %+v", e.Alignment)
	}
}

func TestMappingAlignmentAmbiguous(t *testing.T) {
	p, err := PrepareMappings(MappingSide{Rules: rewrite.RuleSet{SchemaVersion: 1}}, MappingSide{Rules: rewrite.RuleSet{SchemaVersion: 1, Rules: []rewrite.Rule{fieldRule("a", "src", "user"), fieldRule("b", "src", "owner")}}})
	if err != nil {
		t.Fatal(err)
	}
	r, err := p.Compare(impactInput("table src"))
	if err != nil {
		t.Fatal(err)
	}
	e := r.Entries[0]
	if e.AfterRewrite == nil || e.AfterRewrite.Coverage.RewriteComplete || e.Classification != Indeterminate || len(e.Alignment.Ambiguous) == 0 {
		t.Fatalf("ambiguous preview or alignment overstated: %+v", e)
	}

	p, err = PrepareMappings(MappingSide{Rules: rewrite.RuleSet{SchemaVersion: 1}}, MappingSide{Rules: rewrite.RuleSet{SchemaVersion: 1}})
	if err != nil {
		t.Fatal(err)
	}
	r, err = p.Compare(impactInput("| mystery | table src"))
	if err != nil {
		t.Fatal(err)
	}
	if r.Entries[0].Classification != Indeterminate {
		t.Fatalf("equal incomplete evidence called unchanged: %+v", r.Entries[0])
	}
}

func TestImpactNeverWrites(t *testing.T) {
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "q.spl")
	original := []byte("table src\n")
	if err := os.WriteFile(path, original, 0o600); err != nil {
		t.Fatal(err)
	}
	input, err := corpusio.LoadDirectory(dir)
	if err != nil {
		t.Fatal(err)
	}
	p, err := PrepareMappings(MappingSide{Rules: rewrite.RuleSet{SchemaVersion: 1}}, MappingSide{Rules: rewrite.RuleSet{SchemaVersion: 1, Rules: []rewrite.Rule{fieldRule("rename", "src", "user")}}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := p.Compare(input); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(original) {
		t.Fatalf("source changed: %q", got)
	}
}
