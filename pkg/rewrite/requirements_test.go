package rewrite

import (
	"encoding/json"
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
			assertResourceAnalysesDetached(t, got)
		})
	}
}

func TestResourceLimitedRewriteUsesCachedEvidenceAfterCandidateFormation(t *testing.T) {
	text := strings.Repeat("a ", 4097)
	rules := []Rule{{ID: "limited", Kind: "field", Source: conditionIdentity("a"), Target: conditionIdentity("b")}}
	pending, err := formCandidate(
		analysis.QueryDocument{Text: text, SourceID: "cached-limited"},
		rules,
		conditionProbes(rules),
	)
	if err != nil {
		t.Fatal(err)
	}
	if !resourceLimited(pending) {
		t.Fatal("expected resource-limited candidate formation")
	}

	want := cloneAnalysisForTest(t, &pending.originalEvidence.Analysis)
	pending.original = nil
	if !resourceLimited(pending) {
		t.Fatal("resource-limit detection consulted the rewrite session instead of cached state")
	}
	got := finishResourceLimitedRewrite(pending, Preview, rules)
	if !reflect.DeepEqual(got.OriginalAnalysis, want) || !reflect.DeepEqual(got.CandidateAnalysis, want) {
		t.Fatalf("resource-limited result did not reuse cached evidence:\noriginal=%+v\ncandidate=%+v\nwant=%+v", got.OriginalAnalysis, got.CandidateAnalysis, want)
	}
	assertResourceAnalysesDetached(t, got)
}

func assertResourceAnalysesDetached(t *testing.T, got *Result) {
	t.Helper()
	wantOriginal := cloneAnalysisForTest(t, got.OriginalAnalysis)
	wantCandidate := cloneAnalysisForTest(t, got.CandidateAnalysis)
	mutateAnalysisCollections(got.OriginalAnalysis, "original-mutated")
	if !reflect.DeepEqual(got.CandidateAnalysis, wantCandidate) {
		t.Fatal("original analysis mutation leaked into candidate analysis")
	}

	mutatedOriginal := cloneAnalysisForTest(t, got.OriginalAnalysis)
	mutateAnalysisCollections(got.CandidateAnalysis, "candidate-mutated")
	if !reflect.DeepEqual(got.OriginalAnalysis, mutatedOriginal) {
		t.Fatal("candidate analysis mutation leaked into original analysis")
	}
	if reflect.DeepEqual(got.OriginalAnalysis, wantOriginal) || reflect.DeepEqual(got.CandidateAnalysis, wantCandidate) {
		t.Fatal("detachment proof did not mutate both analyses")
	}
}

func cloneAnalysisForTest(t *testing.T, result *analysis.Result) *analysis.Result {
	t.Helper()
	data, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	var clone analysis.Result
	if err := json.Unmarshal(data, &clone); err != nil {
		t.Fatal(err)
	}
	return &clone
}

func mutateStringCollection(values *[]string, marker string) {
	if len(*values) == 0 {
		*values = append(*values, marker)
		return
	}
	(*values)[0] = marker
}

func mutateFieldState(state *analysis.FieldState, marker string) {
	if len(state.Fields) == 0 {
		state.Fields = append(state.Fields, analysis.FieldBinding{Name: marker + "-field", OriginReferenceIDs: []string{marker + "-field-origin"}})
	} else {
		state.Fields[0].Name = marker + "-field"
		mutateStringCollection(&state.Fields[0].OriginReferenceIDs, marker+"-field-origin")
	}
	mutateStringCollection(&state.Removed, marker+"-removed")
}

func mutateAnalysisCollections(result *analysis.Result, marker string) {
	mutateStringCollection(&result.Coverage.Reasons, marker+"-coverage")
	if len(result.Stages) == 0 {
		result.Stages = append(result.Stages, analysis.Stage{ID: marker + "-stage"})
	} else {
		result.Stages[0].ID = marker + "-stage"
	}
	if len(result.Scopes) == 0 {
		result.Scopes = append(result.Scopes, analysis.Scope{ID: marker + "-scope"})
	} else {
		result.Scopes[0].ID = marker + "-scope"
	}
	if len(result.References) == 0 {
		result.References = append(result.References, analysis.Reference{ID: marker + "-reference", OriginReferenceIDs: []string{marker + "-origin"}})
	} else {
		result.References[0].ID = marker + "-reference"
		mutateStringCollection(&result.References[0].OriginReferenceIDs, marker+"-origin")
	}
	if len(result.Lineage) == 0 {
		result.Lineage = append(result.Lineage, analysis.Lineage{
			StageID:     marker + "-lineage",
			Transitions: []analysis.Transition{{InputReferenceIDs: []string{marker + "-transition"}}},
		})
	}
	lineage := &result.Lineage[0]
	lineage.StageID = marker + "-lineage"
	mutateFieldState(&lineage.Before, marker+"-before")
	mutateFieldState(&lineage.After, marker+"-after")
	if len(lineage.Transitions) == 0 {
		lineage.Transitions = append(lineage.Transitions, analysis.Transition{InputReferenceIDs: []string{marker + "-transition"}})
	} else {
		mutateStringCollection(&lineage.Transitions[0].InputReferenceIDs, marker+"-transition")
	}
	mutateStringCollection(&result.Dependencies.Indexes, marker+"-index")
	mutateStringCollection(&result.Dependencies.Sources, marker+"-source")
	mutateStringCollection(&result.Dependencies.SourceTypes, marker+"-sourcetype")
	mutateStringCollection(&result.Dependencies.Datasets, marker+"-dataset")
	mutateStringCollection(&result.Dependencies.Lookups, marker+"-lookup")
	mutateStringCollection(&result.Dependencies.DataModels, marker+"-data-model")
	mutateStringCollection(&result.Dependencies.Macros, marker+"-macro")
	if len(result.Diagnostics) == 0 {
		result.Diagnostics = append(result.Diagnostics, analysis.Diagnostic{Code: marker + "-diagnostic"})
	} else {
		result.Diagnostics[0].Code = marker + "-diagnostic"
	}
	mutateStringCollection(&result.Requirements.Coverage.Reasons, marker+"-requirement-coverage")
	if len(result.Requirements.Items) == 0 {
		result.Requirements.Items = append(result.Requirements.Items, analysis.RequirementItem{ID: marker + "-item", Occurrences: []analysis.RequirementOccurrence{{ReferenceID: marker + "-occurrence"}}})
	} else {
		result.Requirements.Items[0].ID = marker + "-item"
		if len(result.Requirements.Items[0].Occurrences) == 0 {
			result.Requirements.Items[0].Occurrences = append(result.Requirements.Items[0].Occurrences, analysis.RequirementOccurrence{ReferenceID: marker + "-occurrence"})
		} else {
			result.Requirements.Items[0].Occurrences[0].ReferenceID = marker + "-occurrence"
		}
	}
	if len(result.Requirements.Gaps) == 0 {
		result.Requirements.Gaps = append(result.Requirements.Gaps, analysis.RequirementGap{
			Code:            marker + "-gap",
			ReferenceIDs:    []string{marker + "-gap-reference"},
			DiagnosticCodes: []string{marker + "-gap-diagnostic"},
		})
	} else {
		result.Requirements.Gaps[0].Code = marker + "-gap"
		mutateStringCollection(&result.Requirements.Gaps[0].ReferenceIDs, marker+"-gap-reference")
		mutateStringCollection(&result.Requirements.Gaps[0].DiagnosticCodes, marker+"-gap-diagnostic")
	}
	if len(result.Requirements.Diagnostics) == 0 {
		result.Requirements.Diagnostics = append(result.Requirements.Diagnostics, analysis.Diagnostic{Code: marker + "-requirement-diagnostic"})
	} else {
		result.Requirements.Diagnostics[0].Code = marker + "-requirement-diagnostic"
	}
}
