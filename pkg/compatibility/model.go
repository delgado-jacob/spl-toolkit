// Package compatibility admits offline compatibility assessments against captured evidence.
package compatibility

import (
	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/closure"
	"github.com/delgado-jacob/spl-toolkit/pkg/environment"
	"github.com/delgado-jacob/spl-toolkit/pkg/validation"
)

type InputBinding struct {
	InputID  string                     `json:"input_id"`
	ObjectID string                     `json:"object_id"`
	Expected environment.ObjectIdentity `json:"expected"`
	SchemaID string                     `json:"schema_id,omitempty"`
}
type AssessmentRequest struct {
	SchemaVersion      int                      `json:"schema_version"`
	Requirements       analysis.RequirementSet  `json:"requirements"`
	QueryScope         environment.CaptureScope `json:"query_scope"`
	InputBindings      []InputBinding           `json:"input_bindings"`
	Document           *analysis.QueryDocument  `json:"document,omitempty"`
	DependencyBindings []closure.Binding        `json:"dependency_bindings,omitempty"`
}
type Request struct {
	SchemaVersion      int                       `json:"schema_version"`
	Requirements       analysis.RequirementSet   `json:"requirements"`
	QueryScope         environment.CaptureScope  `json:"query_scope"`
	InputBindings      []InputBinding            `json:"input_bindings"`
	Document           *analysis.QueryDocument   `json:"document,omitempty"`
	DependencyBindings []closure.Binding         `json:"dependency_bindings,omitempty"`
	Snapshot           environment.Snapshot      `json:"snapshot"`
	SchemaBundle       *environment.SchemaBundle `json:"schema_bundle,omitempty"`
}

// Reason locates the evidence limit or finding on the affected obligation.
type Reason struct {
	Dimension         string              `json:"dimension,omitempty"`
	Locations         []analysis.Location `json:"locations"`
	ReferenceIDs      []string            `json:"reference_ids"`
	CandidateInputIDs []string            `json:"candidate_input_ids"`
	Code              string              `json:"code"`
	Message           string              `json:"message"`
	InputID           string              `json:"input_id,omitempty"`
	RequirementID     string              `json:"requirement_id,omitempty"`
	ObjectID          string              `json:"object_id,omitempty"`
	SchemaID          string              `json:"schema_id,omitempty"`
	Location          *analysis.Location  `json:"location,omitempty"`
	Path              string              `json:"path,omitempty"`
}
type ObjectEvidence struct {
	ObjectID string                     `json:"object_id"`
	Expected environment.ObjectIdentity `json:"expected"`
	Object   *environment.Object        `json:"object,omitempty"`
}
type SchemaEvidence struct {
	SchemaID string                    `json:"schema_id"`
	Binding  environment.SchemaBinding `json:"binding"`
}
type RequirementOutcome struct {
	Query                analysis.RequirementQueryIdentity `json:"query"`
	RequirementID        string                            `json:"requirement_id"`
	DefinitionObjectID   string                            `json:"definition_object_id,omitempty"`
	InputID              string                            `json:"input_id,omitempty"`
	Applicability        string                            `json:"applicability"`
	Outcome              string                            `json:"outcome"`
	Reasons              []Reason                          `json:"reasons"`
	Capabilities         []environment.Capability          `json:"capabilities"`
	Objects              []ObjectEvidence                  `json:"objects"`
	Schemas              []SchemaEvidence                  `json:"schemas"`
	FieldProjection      *validation.FieldProjection       `json:"field_projection,omitempty"`
	SourceIntervals      []closure.SourceInterval          `json:"source_intervals"`
	InvocationProvenance []closure.InvocationFrame         `json:"invocation_provenance"`
}
type InputOutcome struct {
	InputID        string           `json:"input_id"`
	Outcome        string           `json:"outcome"`
	Objects        []ObjectEvidence `json:"objects"`
	RequirementIDs []string         `json:"requirement_ids"`
	Reasons        []Reason         `json:"reasons"`
}
type Coverage struct {
	CollectionKind string   `json:"collection_kind,omitempty"`
	Dimension      string   `json:"dimension"`
	State          string   `json:"state"`
	InputID        string   `json:"input_id,omitempty"`
	ObjectID       string   `json:"object_id,omitempty"`
	SchemaID       string   `json:"schema_id,omitempty"`
	Reasons        []Reason `json:"reasons"`
}
type Provenance struct {
	QueryDigest              string `json:"query_digest"`
	SourceID                 string `json:"source_id"`
	CapabilityRevision       string `json:"capability_revision"`
	AnalysisContractVersion  int    `json:"analysis_contract_version"`
	RequirementSetVersion    int    `json:"requirement_set_version"`
	EnvironmentDigest        string `json:"environment_digest"`
	SchemaBundleDigest       string `json:"schema_bundle_digest,omitempty"`
	AssessmentIdentityDigest string `json:"assessment_identity_digest"`
}

type Report struct {
	QueryScope            environment.CaptureScope      `json:"query_scope"`
	InputBindings         []InputBinding                `json:"input_bindings"`
	Observation           *environment.ObservationScope `json:"observation,omitempty"`
	SchemaVersion         int                           `json:"schema_version"`
	Outcome               string                        `json:"outcome"`
	Requirements          analysis.RequirementSet       `json:"requirements"`
	EffectiveRequirements *analysis.RequirementSet      `json:"effective_requirements,omitempty"`
	Closure               *closure.Report               `json:"closure,omitempty"`
	Inputs                []InputOutcome                `json:"inputs"`
	RequirementOutcomes   []RequirementOutcome          `json:"requirement_outcomes"`
	Correlation           analysis.CorrelationGraph     `json:"correlation"`
	Coverage              []Coverage                    `json:"coverage"`
	Diagnostics           []environment.Diagnostic      `json:"diagnostics"`
	Reasons               []Reason                      `json:"reasons"`
	Provenance            Provenance                    `json:"provenance"`
}
