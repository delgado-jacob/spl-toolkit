package compatibility

import (
	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/closure"
	"github.com/delgado-jacob/spl-toolkit/pkg/environment"
)

type ResolutionBinding struct {
	OriginalInputID string                     `json:"original_input_id"`
	ResolvedValue   *string                    `json:"resolved_value,omitempty"`
	ObjectID        string                     `json:"object_id"`
	Expected        environment.ObjectIdentity `json:"expected"`
	SchemaID        string                     `json:"schema_id,omitempty"`
}
type ResolutionAssessment struct {
	QueryScope         environment.CaptureScope `json:"query_scope"`
	InputBindings      []ResolutionBinding      `json:"input_bindings"`
	DependencyBindings []closure.Binding        `json:"dependency_bindings,omitempty"`
}
type ArtifactIdentity struct {
	EnvironmentDigest  string `json:"environment_digest"`
	SchemaBundleDigest string `json:"schema_bundle_digest,omitempty"`
}
type ResolutionRequirementOutcome struct {
	OriginalInputID        string                           `json:"original_input_id,omitempty"`
	OriginalRequirementID  string                           `json:"original_requirement_id,omitempty"`
	CandidateInputID       string                           `json:"candidate_input_id,omitempty"`
	CandidateRequirementID string                           `json:"candidate_requirement_id"`
	AssessedOccurrences    []analysis.RequirementOccurrence `json:"assessed_occurrences"`
	Evidence               RequirementOutcome               `json:"evidence"`
}
type ResolutionInputOutcome struct {
	OriginalInputID  string                              `json:"original_input_id,omitempty"`
	CandidateInputID string                              `json:"candidate_input_id"`
	Occurrences      []analysis.ResolutionOccurrencePair `json:"occurrences"`
	Evidence         InputOutcome                        `json:"evidence"`
}
type ResolutionReason struct {
	OriginalInputID string `json:"original_input_id,omitempty"`
	Evidence        Reason `json:"evidence"`
}
type ResolutionCoverage struct {
	OriginalInputID string   `json:"original_input_id,omitempty"`
	Evidence        Coverage `json:"evidence"`
}
type ResolutionReport struct {
	SchemaVersion         int                            `json:"schema_version"`
	Outcome               string                         `json:"outcome"`
	Requirements          analysis.RequirementSet        `json:"requirements"`
	EffectiveRequirements *analysis.RequirementSet       `json:"effective_requirements,omitempty"`
	Closure               *closure.Report                `json:"closure,omitempty"`
	InputBindings         []ResolutionBinding            `json:"input_bindings"`
	DependencyBindings    []closure.Binding              `json:"dependency_bindings"`
	Inputs                []ResolutionInputOutcome       `json:"inputs"`
	RequirementOutcomes   []ResolutionRequirementOutcome `json:"requirement_outcomes"`
	Correlation           analysis.CorrelationGraph      `json:"correlation"`
	Coverage              []ResolutionCoverage           `json:"coverage"`
	Reasons               []ResolutionReason             `json:"reasons"`
	Diagnostics           []environment.Diagnostic       `json:"diagnostics"`
	Provenance            Provenance                     `json:"provenance"`
}
