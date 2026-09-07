package validation

import "github.com/delgado-jacob/spl-toolkit/pkg/analysis"

// ValidateBatch validates in source order with one normalized catalog. Any
// request or internal failure discards the batch; syntax errors remain reports.
func ValidateBatch(documents []analysis.QueryDocument, catalog FieldCatalog) (*BatchReport, error) {
	if len(documents) == 0 {
		return nil, inputError("documents must be nonempty")
	}
	catalog, err := normalizeCatalog(catalog)
	if err != nil {
		return nil, err
	}
	normalized := make([]analysis.QueryDocument, len(documents))
	for i, document := range documents {
		normalized[i], err = normalizeDocument(document)
		if err != nil {
			return nil, err
		}
	}
	batch := &BatchReport{SchemaVersion: 1, Status: analysis.Valid, Reports: make([]*Report, 0, len(documents))}
	for _, document := range normalized {
		report, err := validate(document, catalog)
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
