package validation

import "github.com/delgado-jacob/spl-toolkit/pkg/analysis"

// ValidateSchemaBatch prepares the target once and preserves document order.
// Input or internal errors discard the whole batch; query syntax errors are reports.
func ValidateSchemaBatch(documents []analysis.QueryDocument, target SchemaTarget) (*SchemaBatchReport, error) {
	return validateSchemaBatchWith(documents, target, prepareSchemaTarget)
}

func validateSchemaBatchWith(documents []analysis.QueryDocument, target SchemaTarget, prepare func(SchemaTarget) (preparedSchemaTarget, error)) (*SchemaBatchReport, error) {
	if len(documents) == 0 {
		return nil, inputError("documents must be nonempty")
	}
	normalized := make([]analysis.QueryDocument, len(documents))
	for i, document := range documents {
		var err error
		normalized[i], err = normalizeDocument(document)
		if err != nil {
			return nil, err
		}
	}
	prepared, err := prepare(target)
	if err != nil {
		return nil, err
	}
	return validateSchemaBatchPrepared(normalized, prepared)
}

func validateSchemaBatchPrepared(documents []analysis.QueryDocument, prepared preparedSchemaTarget) (*SchemaBatchReport, error) {
	batch := &SchemaBatchReport{SchemaVersion: 1, Status: analysis.Valid, Reports: make([]*SchemaReport, 0, len(documents))}
	for _, document := range documents {
		report, err := validatePreparedSchema(document, prepared)
		if err != nil {
			return nil, err
		}
		batch.Reports = append(batch.Reports, report)
		if report.Status == analysis.Invalid || (report.Status == analysis.Incomplete && batch.Status == analysis.Valid) {
			batch.Status = report.Status
		}
	}
	return batch, nil
}
