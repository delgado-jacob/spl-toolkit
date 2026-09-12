package validation

import "github.com/delgado-jacob/spl-toolkit/pkg/analysis"

// PreparedFieldCatalog owns a normalized catalog reusable across documents.
type PreparedFieldCatalog struct{ catalog *FieldCatalog }

func PrepareFieldCatalog(catalog FieldCatalog) (*PreparedFieldCatalog, error) {
	normalized, err := normalizeCatalog(catalog)
	if err != nil {
		return nil, err
	}
	return &PreparedFieldCatalog{catalog: &normalized}, nil
}

func (p *PreparedFieldCatalog) Validate(document analysis.QueryDocument) (*Report, error) {
	if p == nil || p.catalog == nil {
		return nil, inputError("field catalog is not prepared")
	}
	normalized, err := normalizeDocument(document)
	if err != nil {
		return nil, err
	}
	catalog := *p.catalog
	catalog.Fields = append([]string{}, catalog.Fields...)
	catalog.OptionalFields = append([]string{}, catalog.OptionalFields...)
	return validate(normalized, catalog)
}

// PreparedSchemaTarget holds the canonical compiled offline schema target.
type PreparedSchemaTarget struct{ target preparedSchemaTarget }

func PrepareSchemaTarget(target SchemaTarget) (*PreparedSchemaTarget, error) {
	prepared, err := prepareSchemaTarget(target)
	if err != nil {
		return nil, err
	}
	return &PreparedSchemaTarget{target: prepared}, nil
}

func (p *PreparedSchemaTarget) Validate(document analysis.QueryDocument) (*SchemaReport, error) {
	if p == nil || p.target == nil {
		return nil, inputError("schema target is not prepared")
	}
	normalized, err := normalizeDocument(document)
	if err != nil {
		return nil, err
	}
	return validatePreparedSchema(normalized, p.target)
}
