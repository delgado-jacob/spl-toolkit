package rewrite

import (
	"sort"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
)

// finishRewrite has one publication gate. Canonical Verify owns correspondence,
// including translated unknown evidence, bindings, origins and logical phases.
// Candidate selection has already refused groups crossing affected uncertainty.
func finishRewrite(pending *pendingRewrite, mode Mode, validation *CandidateValidation) *Result {
	candidate := pending.candidate
	original := pending.original.Evidence().Analysis
	next := candidate.Session.Evidence().Analysis
	proof := pending.original.Verify(candidate.Session, candidate.Rendering)
	result := &Result{SchemaVersion: 1, Document: original.Document, Mode: mode,
		Status:       combineStatus(original.Status, next.Status),
		OriginalText: original.Document.Text, CandidateText: candidate.Text, Text: original.Document.Text,
		Changes: append([]Change{}, candidate.Changes...), RuleEvaluations: copyGroupEvaluations(candidate.Evaluations),
		OriginalAnalysis: &original, CandidateAnalysis: &next, CandidateValidation: validation,
		Coverage: Coverage{SyntaxComplete: original.Coverage.SyntaxComplete && next.Coverage.SyntaxComplete,
			SemanticComplete: original.Coverage.SemanticComplete && next.Coverage.SemanticComplete,
			RewriteComplete:  !candidate.Incomplete && proof.Proven, Reasons: []string{}},
	}
	gate := proof.Proven && result.Coverage.SyntaxComplete && original.Status != analysis.Invalid && next.Status != analysis.Invalid
	reasons := append(append([]string{}, original.Coverage.Reasons...), next.Coverage.Reasons...)
	if !result.Coverage.RewriteComplete {
		result.Status = combineStatus(result.Status, analysis.Incomplete)
	}
	for _, group := range candidate.Groups {
		if group.Reason != "" {
			reasons = append(reasons, group.Reason)
		}
	}
	if candidate.Incomplete {
		for _, evaluation := range candidate.Evaluations {
			switch evaluation.Reason {
			case ReasonConditionUnknown, ReasonDynamicReference, ReasonUnsupportedReference:
				reasons = append(reasons, evaluation.Reason)
			}
		}
	}
	if validation != nil {
		var status analysis.Status
		if validation.Kind == "field_list" {
			coverage := validation.FieldList.Coverage
			result.Coverage.Validation, status = &coverage, validation.FieldList.Status
		} else {
			coverage := validation.Schema.Coverage
			result.Coverage.Validation, status = &coverage, validation.Schema.Status
		}
		result.Status = combineStatus(result.Status, status)
		reasons = append(reasons, result.Coverage.Validation.Reasons...)
		gate = gate && status == analysis.Valid && result.Coverage.Validation.SyntaxComplete && result.Coverage.Validation.SemanticComplete && result.Coverage.Validation.SchemaComplete
	}
	if !gate {
		reasons = append(reasons, ReasonPostVerificationFailed)
	}
	sort.Strings(reasons)
	for _, reason := range reasons {
		if len(result.Coverage.Reasons) == 0 || result.Coverage.Reasons[len(result.Coverage.Reasons)-1] != reason {
			result.Coverage.Reasons = append(result.Coverage.Reasons, reason)
		}
	}
	if mode == Apply && gate && candidate.Text != original.Document.Text {
		result.Text, result.Committed = candidate.Text, true
	}
	for i := range result.Changes {
		change := &result.Changes[i]
		change.Committed = result.Committed && change.CandidateApplied
		if change.CandidateApplied {
			// Public applied means inclusion in the candidate, including preview.
			change.Outcome = "applied"
			if mode == Apply && !gate {
				change.Outcome, change.Reason = "skipped", ReasonPostVerificationFailed
			}
		}
	}
	return result
}

func combineStatus(a, b analysis.Status) analysis.Status {
	if a == analysis.Invalid || b == analysis.Invalid {
		return analysis.Invalid
	}
	if a == analysis.Incomplete || b == analysis.Incomplete {
		return analysis.Incomplete
	}
	return analysis.Valid
}
