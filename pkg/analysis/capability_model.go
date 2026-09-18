package analysis

import "encoding/json"

type CapabilityState string

const (
	CapabilitySupported     CapabilityState = "supported"
	CapabilityPartial       CapabilityState = "partial"
	CapabilityUnsupported   CapabilityState = "unsupported"
	CapabilityNotApplicable CapabilityState = "not_applicable"
	CapabilityUnassessed    CapabilityState = "unassessed"
)

type CapabilityClaim struct {
	State       CapabilityState `json:"state"`
	EvidenceIDs []string        `json:"evidence_ids"`
	Limitations []string        `json:"limitations"`
}

type CapabilityDimensions struct {
	Syntax        CapabilityClaim `json:"syntax"`
	Semantics     CapabilityClaim `json:"semantics"`
	Requirements  CapabilityClaim `json:"requirements"`
	Linting       CapabilityClaim `json:"linting"`
	SafeRewriting CapabilityClaim `json:"safe_rewriting"`
}

type CapabilityRecord struct {
	ID       string `json:"id"`
	Language string `json:"language"`
	Profile  string `json:"profile"`
	Kind     string `json:"kind"`
	Name     string `json:"name"`
	Form     string `json:"form"`
	// GrammarRegistered records a local parser registration fact, not syntax coverage.
	GrammarRegistered bool                 `json:"grammar_registered"`
	Provenance        CapabilityProvenance `json:"provenance"`
	Dimensions        CapabilityDimensions `json:"dimensions"`
}

type CapabilityStateCounts struct {
	Applicable    int `json:"applicable"`
	Covered       int `json:"covered"`
	Supported     int `json:"supported"`
	Partial       int `json:"partial"`
	Unsupported   int `json:"unsupported"`
	NotApplicable int `json:"not_applicable"`
	Unassessed    int `json:"unassessed"`
}

type CapabilitySummary struct {
	Syntax        CapabilityStateCounts `json:"syntax"`
	Semantics     CapabilityStateCounts `json:"semantics"`
	Requirements  CapabilityStateCounts `json:"requirements"`
	Linting       CapabilityStateCounts `json:"linting"`
	SafeRewriting CapabilityStateCounts `json:"safe_rewriting"`
}

type CapabilityProvenance struct {
	SourceFamily string `json:"source_family"`
	Reference    string `json:"reference,omitempty"`
	Note         string `json:"note"`
}

type CapabilityEvidenceClassification string

const (
	CapabilityEvidencePositive   CapabilityEvidenceClassification = "positive"
	CapabilityEvidenceNegative   CapabilityEvidenceClassification = "negative"
	CapabilityEvidenceIncomplete CapabilityEvidenceClassification = "incomplete"
)

type CapabilityDiagnosticExpectation struct {
	Code     string   `json:"code"`
	Category string   `json:"category"`
	Severity string   `json:"severity"`
	Location Location `json:"location"`
}

type CapabilitySyntaxObservation struct {
	Complete    bool                              `json:"complete"`
	Diagnostics []CapabilityDiagnosticExpectation `json:"diagnostics"`
}

type CapabilityStageExpectation struct {
	Command          string `json:"command"`
	SemanticComplete bool   `json:"semantic_complete"`
}

type CapabilityReferenceExpectation struct {
	NormalizedName string   `json:"normalized_name"`
	Kind           string   `json:"kind"`
	Role           string   `json:"role"`
	Resolution     string   `json:"resolution"`
	Binding        string   `json:"binding"`
	Location       Location `json:"location"`
}

type CapabilityDependencyExpectation struct {
	Kind string `json:"kind"`
	Name string `json:"name"`
}

type CapabilityTransitionExpectation struct {
	Operation string `json:"operation"`
	Output    string `json:"output"`
}

type CapabilitySemanticsObservation struct {
	Status       Status                            `json:"status"`
	Complete     bool                              `json:"complete"`
	Stages       []CapabilityStageExpectation      `json:"stages"`
	References   []CapabilityReferenceExpectation  `json:"references"`
	Dependencies []CapabilityDependencyExpectation `json:"dependencies"`
	Transitions  []CapabilityTransitionExpectation `json:"transitions"`
	Diagnostics  []CapabilityDiagnosticExpectation `json:"diagnostics"`
}

type CapabilityRequirementExpectation struct {
	Kind       string `json:"kind"`
	Identity   string `json:"identity"`
	Role       string `json:"role"`
	Necessity  string `json:"necessity"`
	Resolution string `json:"resolution"`
}

type CapabilityRequirementsObservation struct {
	QueryStatus Status                             `json:"query_status"`
	Complete    bool                               `json:"complete"`
	Items       []CapabilityRequirementExpectation `json:"items"`
	GapCodes    []string                           `json:"gap_codes"`
}

type CapabilityLintingObservation struct {
	Diagnostics []CapabilityDiagnosticExpectation `json:"diagnostics"`
}

type CapabilityRewriteObservation struct {
	Status                Status   `json:"status"`
	Committed             bool     `json:"committed"`
	RewriteComplete       bool     `json:"rewrite_complete"`
	Text                  string   `json:"text"`
	CandidateText         string   `json:"candidate_text"`
	CoverageReasons       []string `json:"coverage_reasons"`
	ChangeReasons         []string `json:"change_reasons"`
	RuleEvaluationReasons []string `json:"rule_evaluation_reasons"`
}

type CapabilityEvidenceObservations struct {
	Syntax        *CapabilitySyntaxObservation       `json:"syntax,omitempty"`
	Semantics     *CapabilitySemanticsObservation    `json:"semantics,omitempty"`
	Requirements  *CapabilityRequirementsObservation `json:"requirements,omitempty"`
	Linting       *CapabilityLintingObservation      `json:"linting,omitempty"`
	SafeRewriting *CapabilityRewriteObservation      `json:"safe_rewriting,omitempty"`
}

type CapabilityEvidence struct {
	ID             string                           `json:"id"`
	Classification CapabilityEvidenceClassification `json:"classification"`
	Document       QueryDocument                    `json:"document"`
	Observations   CapabilityEvidenceObservations   `json:"observations"`
	RewriteRequest json.RawMessage                  `json:"rewrite_request,omitempty"`
	Provenance     CapabilityProvenance             `json:"provenance"`
}

func cloneCapabilityRecords(records []CapabilityRecord) []CapabilityRecord {
	if records == nil {
		return nil
	}
	cloned := make([]CapabilityRecord, len(records))
	for i := range records {
		cloned[i] = cloneCapabilityRecord(records[i])
	}
	return cloned
}

func cloneCapabilityRecord(record CapabilityRecord) CapabilityRecord {
	record.Dimensions = cloneCapabilityDimensions(record.Dimensions)
	return record
}

func cloneCapabilityDimensions(dimensions CapabilityDimensions) CapabilityDimensions {
	dimensions.Syntax = cloneCapabilityClaim(dimensions.Syntax)
	dimensions.Semantics = cloneCapabilityClaim(dimensions.Semantics)
	dimensions.Requirements = cloneCapabilityClaim(dimensions.Requirements)
	dimensions.Linting = cloneCapabilityClaim(dimensions.Linting)
	dimensions.SafeRewriting = cloneCapabilityClaim(dimensions.SafeRewriting)
	return dimensions
}

func cloneCapabilityClaim(claim CapabilityClaim) CapabilityClaim {
	claim.EvidenceIDs = cloneCapabilitySlice(claim.EvidenceIDs)
	claim.Limitations = cloneCapabilitySlice(claim.Limitations)
	return claim
}

func cloneCapabilityEvidenceCases(cases []CapabilityEvidence) []CapabilityEvidence {
	if cases == nil {
		return nil
	}
	cloned := make([]CapabilityEvidence, len(cases))
	for i := range cases {
		cloned[i] = cloneCapabilityEvidence(cases[i])
	}
	return cloned
}

func cloneCapabilityEvidence(evidence CapabilityEvidence) CapabilityEvidence {
	evidence.Observations = cloneCapabilityEvidenceObservations(evidence.Observations)
	evidence.RewriteRequest = cloneCapabilitySlice(evidence.RewriteRequest)
	return evidence
}

func cloneCapabilityEvidenceObservations(observations CapabilityEvidenceObservations) CapabilityEvidenceObservations {
	if observations.Syntax != nil {
		value := *observations.Syntax
		value.Diagnostics = cloneCapabilitySlice(value.Diagnostics)
		observations.Syntax = &value
	}
	if observations.Semantics != nil {
		value := *observations.Semantics
		value.Stages = cloneCapabilitySlice(value.Stages)
		value.References = cloneCapabilitySlice(value.References)
		value.Dependencies = cloneCapabilitySlice(value.Dependencies)
		value.Transitions = cloneCapabilitySlice(value.Transitions)
		value.Diagnostics = cloneCapabilitySlice(value.Diagnostics)
		observations.Semantics = &value
	}
	if observations.Requirements != nil {
		value := *observations.Requirements
		value.Items = cloneCapabilitySlice(value.Items)
		value.GapCodes = cloneCapabilitySlice(value.GapCodes)
		observations.Requirements = &value
	}
	if observations.Linting != nil {
		value := *observations.Linting
		value.Diagnostics = cloneCapabilitySlice(value.Diagnostics)
		observations.Linting = &value
	}
	if observations.SafeRewriting != nil {
		value := *observations.SafeRewriting
		value.CoverageReasons = cloneCapabilitySlice(value.CoverageReasons)
		value.ChangeReasons = cloneCapabilitySlice(value.ChangeReasons)
		value.RuleEvaluationReasons = cloneCapabilitySlice(value.RuleEvaluationReasons)
		observations.SafeRewriting = &value
	}
	return observations
}

func cloneCapabilitySlice[T any](values []T) []T {
	if values == nil {
		return nil
	}
	cloned := make([]T, len(values))
	copy(cloned, values)
	return cloned
}
