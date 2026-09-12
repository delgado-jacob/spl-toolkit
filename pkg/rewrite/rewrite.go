package rewrite

import (
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
	original  *analysis.RewriteSession
	candidate *rewriteCandidate
}

func formCandidate(document analysis.QueryDocument, rules []Rule, probes []analysis.RewriteFactProbe) (*pendingRewrite, error) {
	original, err := analysis.PrepareRewrite(document, probes)
	if err != nil {
		return nil, err
	}
	selection := selectRules(rules, probes, original.Evidence())
	candidate, err := buildCandidate(original, rules, probes, selection)
	if err != nil {
		return nil, err
	}
	return &pendingRewrite{original: original, candidate: candidate}, nil
}

func rewriteValidationError(err error) error {
	if validation.IsInputError(err) {
		return inputError("validation_target: %w", err)
	}
	return err
}
