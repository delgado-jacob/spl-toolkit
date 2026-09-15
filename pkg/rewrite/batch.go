package rewrite

import (
	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/validation"
)

// RewriteBatch forms ordered candidates before validating one destination target.
// Input and internal failures return no batch; syntax damage remains per-query.
func RewriteBatch(request BatchRequest) (*BatchResult, error) {
	return rewriteBatch(request, formCandidate, validation.ValidateBatch, validation.ValidateSchemaBatch)
}

// Dependencies stay call-local. The public operation always uses canonical
// analysis and validation; package tests can exercise internal failure paths.
func rewriteBatch(request BatchRequest,
	form func(analysis.QueryDocument, []Rule, []analysis.RewriteFactProbe) (*pendingRewrite, error),
	validateFields func([]analysis.QueryDocument, validation.FieldCatalog) (*validation.BatchReport, error),
	validateSchema func([]analysis.QueryDocument, validation.SchemaTarget) (*validation.SchemaBatchReport, error),
) (*BatchResult, error) {
	prepared, err := prepareBatchRequest(request)
	if err != nil {
		return nil, err
	}
	probes := conditionProbes(prepared.Rules)
	pending := make([]*pendingRewrite, len(prepared.Documents))
	documents := make([]analysis.QueryDocument, 0, len(pending))
	documentIndexes := make([]int, 0, len(pending))
	for i, document := range prepared.Documents {
		pending[i], err = form(document, prepared.Rules, probes)
		if err != nil {
			return nil, err
		}
		if resourceLimited(pending[i]) {
			continue
		}
		documents = append(documents, pending[i].candidate.Session.Evidence().Analysis.Document)
		documentIndexes = append(documentIndexes, i)
	}
	reports := make([]*CandidateValidation, len(pending))
	if target := prepared.ValidationTarget; target != nil && len(documents) > 0 {
		if target.Kind == "field_list" {
			batch, err := validateFields(documents, *target.Catalog)
			if err != nil {
				return nil, rewriteValidationError(err)
			}
			for i, report := range batch.Reports {
				reports[documentIndexes[i]] = &CandidateValidation{Kind: target.Kind, FieldList: report}
			}
		} else {
			batch, err := validateSchema(documents, *target.SchemaTarget)
			if err != nil {
				return nil, rewriteValidationError(err)
			}
			for i, report := range batch.Reports {
				reports[documentIndexes[i]] = &CandidateValidation{Kind: target.Kind, Schema: report}
			}
		}
	}
	result := &BatchResult{SchemaVersion: 1, Status: analysis.Valid, Reports: make([]*Result, len(pending))}
	for i, candidate := range pending {
		if resourceLimited(candidate) {
			result.Reports[i] = finishResourceLimitedRewrite(candidate, prepared.Mode, prepared.Rules)
		} else {
			result.Reports[i] = finishRewrite(candidate, prepared.Mode, reports[i])
		}
		result.Status = combineStatus(result.Status, result.Reports[i].Status)
	}
	return result, nil
}
