package validation

import (
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
)

// FieldProjection describes schema declaration membership, not event presence.
// Only required and optional outcomes conclusively admit a typed field.
type FieldProjection struct {
	Admission            analysis.SourceFieldAdmission `json:"admission"`
	Outcome              string                        `json:"outcome"`
	Evidence             []SchemaEvidence              `json:"evidence"`
	SupportingClasses    []SchemaClass                 `json:"supporting_classes"`
	MissingClasses       []SchemaClass                 `json:"missing_classes"`
	IndeterminateClasses []SchemaClass                 `json:"indeterminate_classes"`
}

func validateProjectionField(field analysis.FieldIdentity) error {
	if field.Qualifier != "" || (field.Kind != "atomic" && field.Kind != "path") || len(field.Segments) == 0 || len(field.Segments) > schemaPathSegmentBudget || (field.Kind == "atomic" && len(field.Segments) != 1) {
		return inputError("projection requires an unqualified exact atomic or path field with at most %d segments", schemaPathSegmentBudget)
	}
	for _, segment := range field.Segments {
		if !utf8.ValidString(segment) || strings.TrimSpace(segment) == "" {
			return inputError("projection field segments must be nonblank UTF-8")
		}
	}
	return nil
}

func publicFieldProjection(value fieldProjection) FieldProjection {
	// Legacy intersection can retain a prohibition alongside uncertainty.
	// Typed compatibility must preserve that relevant uncertainty as evidence.
	for _, evidence := range value.Evidence {
		if evidence.Reason != "" {
			value.Admission = analysis.SourceFieldIndeterminate
			if value.Outcome == "required" || value.Outcome == "optional" || value.Outcome == "missing" {
				value.Outcome = "indeterminate"
			}
			break
		}
	}
	// Schema openness and unknown requiredness are not conclusive declarations.
	if value.Outcome != "required" && value.Outcome != "optional" && value.Outcome != "missing" {
		value.Admission = analysis.SourceFieldIndeterminate
	}
	return FieldProjection{
		Admission: value.Admission, Outcome: value.Outcome,
		Evidence: canonicalSchemaEvidence(value.Evidence), SupportingClasses: canonicalSchemaClasses(value.SupportingClasses),
		MissingClasses: canonicalSchemaClasses(value.MissingClasses), IndeterminateClasses: canonicalSchemaClasses(value.IndeterminateClasses),
	}
}

// ProjectField projects a typed exact field through a flat declaration catalog.
func (p *PreparedFieldCatalog) ProjectField(field analysis.FieldIdentity) (FieldProjection, error) {
	if err := validateProjectionField(field); err != nil {
		return FieldProjection{}, err
	}
	if p == nil || p.catalog == nil {
		return FieldProjection{}, inputError("field catalog is not prepared")
	}
	if len(field.Segments) > 1 {
		return publicFieldProjection(newProjection(analysis.SourceFieldIndeterminate, "indeterminate", SchemaEvidence{DeclarationBasis: "unknown", Requirement: "unknown", Reason: "typed_path_not_represented"})), nil
	}
	outcome, admission := "missing", analysis.SourceFieldProhibited
	if slices.Contains(p.catalog.Fields, field.Segments[0]) {
		outcome, admission = "required", analysis.SourceFieldAdmitted
	} else if slices.Contains(p.catalog.OptionalFields, field.Segments[0]) {
		outcome, admission = "optional", analysis.SourceFieldAdmitted
	}
	return publicFieldProjection(newProjection(admission, outcome, SchemaEvidence{DeclarationBasis: "field_catalog", Requirement: outcome})), nil
}

// ProjectField projects exactly the supplied segments without dotted-name reinterpretation.
func (p *PreparedSchemaTarget) ProjectField(field analysis.FieldIdentity) (FieldProjection, error) {
	if err := validateProjectionField(field); err != nil {
		return FieldProjection{}, err
	}
	if p == nil || p.target == nil {
		return FieldProjection{}, inputError("schema target is not prepared")
	}
	var value fieldProjection
	switch target := p.target.(type) {
	case *jsonSchemaTarget:
		value = target.projectInterpretation(field.Segments, &projectionContext{active: map[projectionState]bool{}, seen: map[projectionState]bool{}})
	case *ocsfTarget:
		value = target.projectTyped(field.Segments)
	default:
		value = newProjection(analysis.SourceFieldIndeterminate, "indeterminate", SchemaEvidence{DeclarationBasis: "unknown", Requirement: "unknown", Reason: "unknown_projection"})
	}
	return publicFieldProjection(value), nil
}
