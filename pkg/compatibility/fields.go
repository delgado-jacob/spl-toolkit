package compatibility

import (
	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/validation"
)

func (p *Prepared) assessField(item analysis.RequirementItem, query analysis.RequirementQueryIdentity, resolved map[string]resolvedInput) RequirementOutcome {
	out := baseOutcome(item, query)
	if item.Ownership.State != "proved" {
		if len(item.Ownership.CandidateInputIDs) > 1 {
			out.Outcome = "ambiguous"
		}
		return outcomeReason(out, item, "field_ownership_ambiguous", "The query has not proved one supplying input. Recheck with supported ownership evidence; schema contents cannot choose the owner.", "field_attribution")
	}
	r := resolved[item.InputID]
	out.Objects = objectEvidence(r)
	if !r.explicitSourceEvidence {
		return outcomeReason(out, item, "target_discovery_incomplete", "The logical owner is proved but its full supplying source is not. Supply supported source-selection evidence.", "target_discovery")
	}
	if r.binding == nil || r.binding.SchemaID == "" {
		return outcomeReason(out, item, "schema_not_supplied", "Select schema evidence for this input to assess its field obligations.", "field_schema")
	}
	for _, binding := range p.env.Bindings(r.binding.ObjectID) {
		if binding.SchemaID == r.binding.SchemaID {
			out.Schemas = append(out.Schemas, SchemaEvidence{SchemaID: binding.SchemaID, Binding: binding})
		}
	}
	if item.Resolution != "exact" || item.FieldIdentity == nil {
		return outcomeReason(out, item, "schema_projection_indeterminate", "A dynamic or wildcard field requires proof of the whole obligation; exact declarations cannot establish it.", "field_schema")
	}
	var projection validation.FieldProjection
	var err error
	if catalog, ok := p.env.FieldCatalog(r.binding.ObjectID, r.binding.SchemaID); ok {
		projection, err = catalog.ProjectField(*item.FieldIdentity)
	} else if schema, ok := p.env.SchemaTarget(r.binding.ObjectID, r.binding.SchemaID); ok {
		projection, err = schema.ProjectField(*item.FieldIdentity)
	} else {
		return outcomeReason(out, item, "schema_not_supplied", "Supply a compiled schema target linked to the selected input.", "field_schema")
	}
	if err != nil {
		return outcomeReason(out, item, "schema_projection_indeterminate", "The typed field exceeds the supported projection boundary; supply supported field/schema evidence.", "field_schema")
	}
	out.FieldProjection = &projection
	switch projection.Outcome {
	case "required", "optional":
		out.Outcome = "satisfied"
		return out
	case "missing":
		if len(out.Schemas) == 1 && out.Schemas[0].Binding.SourceCoverage == "complete" {
			out.Outcome = "missing"
			return outcomeReason(out, item, "schema_field_missing", "The selected schema conclusively excludes this field and its applicable source coverage is complete.", "field_schema")
		}
		return outcomeReason(out, item, "schema_partial", "Complete the selected schema's source coverage before its negative field projection can prove absence.", "field_schema")
	default:
		return outcomeReason(out, item, "schema_projection_indeterminate", "The selected schema's open, conditional, unknown or bounded projection cannot prove declaration membership. Supply conclusive typed field evidence.", "field_schema")
	}
}
