package rewrite

import (
	"encoding/json"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/validation"
)

// Rewrite constructs and verifies one candidate. Apply commits only the returned
// string; original source bytes and both canonical analyses remain in the report.
func Rewrite(request Request) (*Result, error) {
	prepared, err := prepareRequest(request)
	if err != nil {
		return nil, err
	}
	pending, err := formCandidate(prepared.Document, prepared.Rules, conditionProbes(prepared.Rules))
	if err != nil {
		return nil, err
	}
	if resourceLimited(pending) {
		return finishResourceLimitedRewrite(pending, prepared.Mode, prepared.Rules), nil
	}
	var report *CandidateValidation
	if target := prepared.ValidationTarget; target != nil {
		document := pending.candidate.Session.Evidence().Analysis.Document
		report = &CandidateValidation{Kind: target.Kind}
		if target.Kind == "field_list" {
			report.FieldList, err = validation.Validate(document, *target.Catalog)
		} else {
			report.Schema, err = validation.ValidateSchema(document, *target.SchemaTarget)
		}
		if err != nil {
			return nil, rewriteValidationError(err)
		}
	}
	return finishRewrite(pending, prepared.Mode, report), nil
}

type pendingRewrite struct {
	original         *analysis.RewriteSession
	originalEvidence analysis.RewriteEvidence
	candidate        *rewriteCandidate
	resourceLimited  bool
}

func formCandidate(document analysis.QueryDocument, rules []Rule, probes []analysis.RewriteFactProbe) (*pendingRewrite, error) {
	original, err := analysis.PrepareRewrite(document, probes)
	if err != nil {
		return nil, err
	}
	evidence := original.Evidence()
	pending := &pendingRewrite{
		original:        original,
		resourceLimited: resourceLimitedAnalysis(evidence.Analysis),
	}
	if pending.resourceLimited {
		pending.originalEvidence = evidence
		return pending, nil
	}
	selection := selectRules(rules, probes, evidence)
	candidate, err := buildCandidate(original, rules, probes, selection)
	if err != nil {
		return nil, err
	}
	pending.candidate = candidate
	return pending, nil
}

func resourceLimited(pending *pendingRewrite) bool {
	return pending != nil && pending.resourceLimited
}

func resourceLimitedAnalysis(result analysis.Result) bool {
	for _, diagnostic := range result.Diagnostics {
		if diagnostic.Code == analysis.CodeAnalysisResourceLimit {
			return true
		}
	}
	return false
}

func finishResourceLimitedRewrite(pending *pendingRewrite, mode Mode, rules []Rule) *Result {
	original := pending.originalEvidence.Analysis
	candidate := detachedAnalysis(original)
	evaluations := make([]RuleEvaluation, len(rules))
	for i, rule := range rules {
		evaluations[i] = RuleEvaluation{
			RuleID:       rule.ID,
			Outcome:      "skipped",
			Reason:       ReasonPostVerificationFailed,
			ReferenceIDs: []string{},
		}
	}
	return &Result{
		SchemaVersion: 1,
		Document:      original.Document,
		Mode:          mode,
		Status:        analysis.Incomplete,
		Coverage: Coverage{
			SyntaxComplete:   false,
			SemanticComplete: false,
			RewriteComplete:  false,
			Reasons:          []string{analysis.CodeAnalysisResourceLimit, ReasonPostVerificationFailed},
		},
		OriginalText:      original.Document.Text,
		CandidateText:     original.Document.Text,
		Text:              original.Document.Text,
		Committed:         false,
		Changes:           []Change{},
		RuleEvaluations:   evaluations,
		OriginalAnalysis:  &original,
		CandidateAnalysis: &candidate,
	}
}

func detachedAnalysis(result analysis.Result) analysis.Result {
	data, _ := json.Marshal(result)
	var clone analysis.Result
	_ = json.Unmarshal(data, &clone)
	return clone
}

func rewriteValidationError(err error) error {
	if validation.IsInputError(err) {
		return inputError("validation_target: %w", err)
	}
	return err
}
