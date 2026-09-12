package rewrite

import (
	"reflect"
	"testing"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/validation"
)

// internalProposedEdit bypasses group selection only inside package tests. It
// retains real canonical rendering and analysis so finishRewrite must reject a
// bad proposal even when a stale candidate summary says it was proven.
func internalProposedEdit(t *testing.T, document analysis.QueryDocument, source, target string, corrupt func([]analysis.RewriteTextEdit)) *pendingRewrite {
	t.Helper()
	original, err := analysis.PrepareRewrite(document, nil)
	if err != nil {
		t.Fatal(err)
	}
	var replacements []analysis.RewriteReplacement
	for _, site := range original.Evidence().Sites {
		if site.Kind == "field" && site.Identity.Name != nil && *site.Identity.Name == source {
			replacements = append(replacements, analysis.RewriteReplacement{SiteID: site.ID, Target: analysis.RewriteIdentity{Name: &target}})
		}
	}
	rendering, err := original.Render(replacements)
	if err != nil {
		t.Fatal(err)
	}
	edits := rendering.Edits()
	if len(edits) == 0 {
		t.Fatal("test requires actual proposed edits")
	}
	if corrupt != nil {
		corrupt(edits)
	}
	text, locations, err := reconstructEdits(document.Text, edits)
	if err != nil {
		t.Fatal(err)
	}
	document = original.Evidence().Analysis.Document
	document.Text = text
	next, err := analysis.PrepareRewrite(document, nil)
	if err != nil {
		t.Fatal(err)
	}
	candidate := &rewriteCandidate{Text: text, Session: next, Rendering: rendering,
		Proof: analysis.RewriteProof{Proven: true}, Changes: []Change{}, Evaluations: []RuleEvaluation{}}
	for i, edit := range edits {
		old, next := edit.Location, locations[i]
		candidate.Changes = append(candidate.Changes, Change{Outcome: "proposed", Reason: "matched", GroupID: "internal-test-group",
			RuleIDs: []string{"internal-test-rule"}, OriginalReferenceIDs: []string{}, CandidateReferenceIDs: []string{},
			OriginalLocation: &old, CandidateLocation: &next, OldText: edit.Before, NewText: edit.After, CandidateApplied: true})
	}
	return &pendingRewrite{original: original, candidate: candidate}
}

// Trusting parser success, a cached proof, or publishing a reduced candidate
// would commit one of these malformed internal proposals.
func TestRewritePostVerificationFailure(t *testing.T) {
	t.Run("original syntax damage", func(t *testing.T) {
		request := rewriteRequest("search src=x | eval =")
		got := requireRewrite(t, request)
		if got.Status != analysis.Invalid || got.Committed || got.Text != request.Document.Text || got.CandidateText != request.Document.Text || got.Coverage.SyntaxComplete {
			t.Fatalf("damaged original rewritten: %+v", got)
		}
		for _, change := range got.Changes {
			if change.CandidateApplied || change.Committed || change.CandidateLocation != nil {
				t.Errorf("damaged original proposed an edit: %+v", change)
			}
		}
	})
	for _, tc := range []struct {
		name, query, language, target, proofCode string
		corrupt                                  func([]analysis.RewriteTextEdit)
		syntaxComplete, semanticComplete         bool
	}{
		{name: "candidate syntax damage", query: "table src", target: "user", proofCode: "candidate_mismatch",
			corrupt: func(edits []analysis.RewriteTextEdit) { edits[0].After = "'" }},
		{name: "derived alias capture", language: "spl2", query: "FROM main | eval account=1 | table src", target: "account", proofCode: "reference_correspondence", syntaxComplete: true, semanticComplete: true},
		{name: "lost implicit linkage", query: "stats sum(src) | table 'sum(src)'", target: "user", proofCode: "required_change_missing", syntaxComplete: true, semanticComplete: true},
		{name: "new affected uncertainty", query: "table src", target: "user", proofCode: "candidate_mismatch", syntaxComplete: true,
			corrupt: func(edits []analysis.RewriteTextEdit) { edits[0].After = "user | mystery" }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			pending := internalProposedEdit(t, analysis.QueryDocument{Text: tc.query, Language: tc.language, SourceID: "fault.spl"}, "src", tc.target, tc.corrupt)
			proof := pending.original.Verify(pending.candidate.Session, pending.candidate.Rendering)
			if proof.Proven || len(proof.Limitations) == 0 || proof.Limitations[0].Code != tc.proofCode {
				t.Fatalf("wrong negative proof fixture: %+v", proof)
			}
			got := finishRewrite(pending, Apply, nil)
			assertRewriteRollback(t, got)
			if got.Coverage.RewriteComplete || got.CandidateAnalysis.Coverage.SyntaxComplete != tc.syntaxComplete || got.CandidateAnalysis.Coverage.SemanticComplete != tc.semanticComplete || got.Status == analysis.Valid {
				t.Fatalf("failure classified as proven or wrong candidate evidence: %+v", got.Coverage)
			}
			for _, change := range got.Changes {
				location := change.CandidateLocation
				if got.CandidateText[location.Start.Offset:location.End.Offset] != change.NewText {
					t.Errorf("malformed edit evidence discarded: %+v", change)
				}
			}
		})
	}
}

// Unrelated old unknown evidence may coexist with safe edits, but an incomplete
// destination report must close the commit gate even with known target fields.
func TestRewriteUnrelatedIncompleteDestinationGate(t *testing.T) {
	const query = "search other=x | mystery | append [ search src=x | table src ]"
	const candidate = "search other=x | mystery | append [ search user=x | table user ]"
	request := rewriteRequest(query)
	got := requireRewrite(t, request)
	if got.Status != analysis.Incomplete || !got.Committed || got.Text != candidate || !got.Coverage.RewriteComplete || got.Coverage.SemanticComplete {
		t.Fatalf("independent edited scope was not proven: %+v", got)
	}
	for _, target := range []*ValidationTarget{
		{Kind: "field_list", Catalog: &validation.FieldCatalog{Fields: []string{"other", "user"}}},
		schemaRewriteTarget(`{"type":"object","properties":{"other":true,"user":true},"additionalProperties":false}`),
	} {
		t.Run(target.Kind, func(t *testing.T) {
			request.ValidationTarget = target
			got := requireRewrite(t, request)
			assertRewriteRollback(t, got)
			if got.Status != analysis.Incomplete || got.Coverage.Validation == nil || got.Coverage.Validation.SemanticComplete {
				t.Fatalf("incomplete destination coverage lost: %+v", got.Coverage)
			}
			if got.CandidateText != candidate {
				t.Fatal("failed gate silently reduced candidate")
			}
		})
	}
}

// Independent safe edits still apply when another source has conflicting rules.
func TestRewriteSafeBesideAmbiguous(t *testing.T) {
	request := rewriteRequest("search src=x other=y | table src other")
	request.Rules = append(request.Rules,
		Rule{ID: "other-a", Kind: "field", Source: conditionIdentity("other"), Target: conditionIdentity("a")},
		Rule{ID: "other-b", Kind: "field", Source: conditionIdentity("other"), Target: conditionIdentity("b")})
	got := requireRewrite(t, request)
	if got.Status != analysis.Incomplete || !got.Committed || got.Text != "search user=x other=y | table user other" || got.Coverage.RewriteComplete {
		t.Fatalf("safe/ambiguous policy: %+v", got)
	}
	applied, ambiguous := 0, 0
	for _, change := range got.Changes {
		if change.OldText == "src" {
			applied++
			if change.Outcome != "applied" || !change.CandidateApplied || !change.Committed || !reflect.DeepEqual(change.RuleIDs, []string{"src-user"}) {
				t.Errorf("safe audit: %+v", change)
			}
		} else {
			ambiguous++
			if change.Outcome != "ambiguous" || change.Reason != ReasonConflictingTargets || change.CandidateApplied || change.Committed || change.CandidateLocation != nil {
				t.Errorf("ambiguous audit: %+v", change)
			}
		}
	}
	if applied != 2 || ambiguous != 4 {
		t.Fatalf("lost alternatives: applied=%d ambiguous=%d", applied, ambiguous)
	}
}

// A destination failure in one independent group must retain the whole preview
// and roll back all included groups, even though a smaller subset would validate.
func TestRewriteWholeRequestGateDoesNotRetrySubset(t *testing.T) {
	const original = "search src=x other=y | table src other"
	const candidate = "search user=x missing=y | table user missing"
	for _, mode := range []Mode{Preview, Apply} {
		t.Run(string(mode), func(t *testing.T) {
			request := rewriteRequest(original)
			request.Mode = mode
			request.Rules = append(request.Rules, Rule{ID: "other-missing", Kind: "field", Source: conditionIdentity("other"), Target: conditionIdentity("missing")})
			request.ValidationTarget = &ValidationTarget{Kind: "field_list", Catalog: &validation.FieldCatalog{Fields: []string{"user", "other"}}}
			got := requireRewrite(t, request)
			if got.Status != analysis.Invalid || got.Text != original || got.CandidateText != candidate || got.Committed || len(got.Changes) != 4 || got.CandidateValidation.FieldList.Status != analysis.Invalid {
				t.Fatalf("whole-request gate selected a reduced candidate: %+v", got)
			}
			for _, change := range got.Changes {
				outcome, reason := "proposed", "matched"
				if mode == Apply {
					outcome, reason = "skipped", ReasonPostVerificationFailed
				}
				if !change.CandidateApplied || change.Committed || change.Outcome != outcome || change.Reason != reason || change.CandidateLocation == nil {
					t.Errorf("lost proposed group evidence: %+v", change)
				}
			}
		})
	}
}
