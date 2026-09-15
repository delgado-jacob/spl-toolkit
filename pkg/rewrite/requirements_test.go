package rewrite

import (
	"reflect"
	"strings"
	"testing"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/validation"
)

func TestRequirementSetRewriteParity(t *testing.T) {
	const original = "search src=x | table src"
	const candidate = "search user=x | table user"
	for _, mode := range []Mode{Preview, Apply} {
		t.Run(string(mode), func(t *testing.T) {
			request := rewriteRequest(original)
			request.Mode = mode
			report := requireRewrite(t, request)
			wantOriginal, err := analysis.Requirements(report.Document)
			if err != nil {
				t.Fatal(err)
			}
			candidateDocument := report.Document
			candidateDocument.Text = candidate
			wantCandidate, err := analysis.Requirements(candidateDocument)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(report.OriginalAnalysis.Requirements, *wantOriginal) {
				t.Fatalf("original requirements differ: got %+v want %+v", report.OriginalAnalysis.Requirements, *wantOriginal)
			}
			if !reflect.DeepEqual(report.CandidateAnalysis.Requirements, *wantCandidate) {
				t.Fatalf("candidate requirements differ: got %+v want %+v", report.CandidateAnalysis.Requirements, *wantCandidate)
			}
		})
	}
}

func TestRewriteResourceLimitNoOp(t *testing.T) {
	text := strings.Repeat("a ", 4097)
	rules := []Rule{
		{ID: "first", Kind: "field", Source: conditionIdentity("a"), Target: conditionIdentity("b")},
		{ID: "second", Kind: "field", Source: conditionIdentity("x"), Target: conditionIdentity("y")},
	}
	for _, mode := range []Mode{Preview, Apply} {
		t.Run(string(mode), func(t *testing.T) {
			request := Request{
				SchemaVersion: 1,
				Mode:          mode,
				Document:      analysis.QueryDocument{Text: text, SourceID: "limited"},
				Rules:         rules,
				ValidationTarget: &ValidationTarget{Kind: "json_schema", SchemaTarget: &validation.SchemaTarget{
					Kind: "json_schema", Schema: []byte(`{"type":42}`),
				}},
			}
			got := requireRewrite(t, request)
			prepared, err := prepareRequest(request)
			if err != nil {
				t.Fatal(err)
			}
			session, err := analysis.PrepareRewrite(prepared.Document, conditionProbes(prepared.Rules))
			if err != nil {
				t.Fatal(err)
			}
			original := session.Evidence().Analysis
			if got.SchemaVersion != 1 || got.Document != prepared.Document || got.Mode != mode || got.Status != analysis.Incomplete {
				t.Fatalf("resource no-op identity or status: %+v", got)
			}
			wantCoverage := Coverage{SyntaxComplete: false, SemanticComplete: false, RewriteComplete: false, Reasons: []string{analysis.CodeAnalysisResourceLimit, ReasonPostVerificationFailed}}
			if !reflect.DeepEqual(got.Coverage, wantCoverage) || got.Coverage.Validation != nil || got.CandidateValidation != nil {
				t.Fatalf("resource no-op coverage or validation: %+v / %+v", got.Coverage, got.CandidateValidation)
			}
			if got.OriginalText != text || got.CandidateText != text || got.Text != text || got.Committed || got.Changes == nil || len(got.Changes) != 0 {
				t.Fatalf("resource no-op text or changes: %+v", got)
			}
			if len(got.RuleEvaluations) != len(rules) {
				t.Fatalf("rule evaluations: %+v", got.RuleEvaluations)
			}
			for i, evaluation := range got.RuleEvaluations {
				want := RuleEvaluation{RuleID: rules[i].ID, Outcome: "skipped", Reason: ReasonPostVerificationFailed, ReferenceIDs: []string{}}
				if !reflect.DeepEqual(evaluation, want) {
					t.Errorf("evaluation %d: got %+v want %+v", i, evaluation, want)
				}
			}
			if got.OriginalAnalysis == nil || got.CandidateAnalysis == nil || got.OriginalAnalysis == got.CandidateAnalysis || !reflect.DeepEqual(*got.OriginalAnalysis, original) || !reflect.DeepEqual(*got.CandidateAnalysis, original) {
				t.Fatalf("bounded analyses differ:\noriginal=%+v\ncandidate=%+v\nwant=%+v", got.OriginalAnalysis, got.CandidateAnalysis, original)
			}
			assertResourceAnalysesDetached(t, request, got)
		})
	}
}

func assertResourceAnalysesDetached(t *testing.T, request Request, got *Result) {
	t.Helper()
	wantCandidate := cloneAnalysisForTest(t, got.CandidateAnalysis)
	mutateAnalysisCollections(got.OriginalAnalysis)
	if !reflect.DeepEqual(got.CandidateAnalysis, wantCandidate) {
		t.Fatal("original analysis mutation leaked into candidate analysis")
	}

	fresh := requireRewrite(t, request)
	wantOriginal := cloneAnalysisForTest(t, fresh.OriginalAnalysis)
	mutateAnalysisCollections(fresh.CandidateAnalysis)
	if !reflect.DeepEqual(fresh.OriginalAnalysis, wantOriginal) {
		t.Fatal("candidate analysis mutation leaked into original analysis")
	}
}

func cloneAnalysisForTest(t *testing.T, result *analysis.Result) *analysis.Result {
	t.Helper()
	clone := *result
	clone.Coverage.Reasons = append([]string{}, result.Coverage.Reasons...)
	clone.Stages = append([]analysis.Stage{}, result.Stages...)
	clone.Scopes = append([]analysis.Scope{}, result.Scopes...)
	clone.References = append([]analysis.Reference{}, result.References...)
	clone.Lineage = append([]analysis.Lineage{}, result.Lineage...)
	clone.Dependencies.Indexes = append([]string{}, result.Dependencies.Indexes...)
	clone.Dependencies.Sources = append([]string{}, result.Dependencies.Sources...)
	clone.Dependencies.SourceTypes = append([]string{}, result.Dependencies.SourceTypes...)
	clone.Dependencies.Datasets = append([]string{}, result.Dependencies.Datasets...)
	clone.Dependencies.Lookups = append([]string{}, result.Dependencies.Lookups...)
	clone.Dependencies.DataModels = append([]string{}, result.Dependencies.DataModels...)
	clone.Dependencies.Macros = append([]string{}, result.Dependencies.Macros...)
	clone.Diagnostics = append([]analysis.Diagnostic{}, result.Diagnostics...)
	clone.Requirements = cloneRequirementSetForTest(result.Requirements)
	return &clone
}

func cloneRequirementSetForTest(in analysis.RequirementSet) analysis.RequirementSet {
	out := in
	out.Coverage.Reasons = append([]string{}, in.Coverage.Reasons...)
	out.Items = append([]analysis.RequirementItem{}, in.Items...)
	for i := range out.Items {
		out.Items[i].Occurrences = append([]analysis.RequirementOccurrence{}, in.Items[i].Occurrences...)
	}
	out.Gaps = append([]analysis.RequirementGap{}, in.Gaps...)
	for i := range out.Gaps {
		out.Gaps[i].ReferenceIDs = append([]string{}, in.Gaps[i].ReferenceIDs...)
		out.Gaps[i].DiagnosticCodes = append([]string{}, in.Gaps[i].DiagnosticCodes...)
	}
	out.Diagnostics = append([]analysis.Diagnostic{}, in.Diagnostics...)
	return out
}

func mutateAnalysisCollections(result *analysis.Result) {
	result.Coverage.Reasons = append(result.Coverage.Reasons, "mutated")
	result.Stages = append(result.Stages, analysis.Stage{ID: "mutated"})
	result.Scopes = append(result.Scopes, analysis.Scope{ID: "mutated"})
	result.References = append(result.References, analysis.Reference{ID: "mutated", OriginReferenceIDs: []string{"mutated"}})
	result.Lineage = append(result.Lineage, analysis.Lineage{StageID: "mutated", Before: analysis.FieldState{Fields: []analysis.FieldBinding{{Name: "mutated", OriginReferenceIDs: []string{"mutated"}}}, Removed: []string{"mutated"}}, After: analysis.FieldState{Fields: []analysis.FieldBinding{}, Removed: []string{}}, Transitions: []analysis.Transition{{InputReferenceIDs: []string{"mutated"}}}})
	result.Dependencies.Indexes = append(result.Dependencies.Indexes, "mutated")
	result.Dependencies.Sources = append(result.Dependencies.Sources, "mutated")
	result.Dependencies.SourceTypes = append(result.Dependencies.SourceTypes, "mutated")
	result.Dependencies.Datasets = append(result.Dependencies.Datasets, "mutated")
	result.Dependencies.Lookups = append(result.Dependencies.Lookups, "mutated")
	result.Dependencies.DataModels = append(result.Dependencies.DataModels, "mutated")
	result.Dependencies.Macros = append(result.Dependencies.Macros, "mutated")
	result.Diagnostics = append(result.Diagnostics, analysis.Diagnostic{Code: "mutated"})
	result.Requirements.Coverage.Reasons = append(result.Requirements.Coverage.Reasons, "mutated")
	result.Requirements.Items = append(result.Requirements.Items, analysis.RequirementItem{ID: "mutated", Occurrences: []analysis.RequirementOccurrence{{ReferenceID: "mutated"}}})
	for i := range result.Requirements.Gaps {
		result.Requirements.Gaps[i].ReferenceIDs = append(result.Requirements.Gaps[i].ReferenceIDs, "mutated")
		result.Requirements.Gaps[i].DiagnosticCodes = append(result.Requirements.Gaps[i].DiagnosticCodes, "mutated")
	}
	result.Requirements.Gaps = append(result.Requirements.Gaps, analysis.RequirementGap{Code: "mutated", ReferenceIDs: []string{"mutated"}, DiagnosticCodes: []string{"mutated"}})
	result.Requirements.Diagnostics = append(result.Requirements.Diagnostics, analysis.Diagnostic{Code: "mutated"})
}
