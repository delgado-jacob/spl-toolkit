package resolution

import (
	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/closure"
	"github.com/delgado-jacob/spl-toolkit/pkg/compatibility"
	"github.com/delgado-jacob/spl-toolkit/pkg/environment"
)

type Resolution struct {
	Placeholder string   `json:"placeholder"`
	Kind        string   `json:"kind"`
	Values      []string `json:"values"`
}
type CompatibilityInputs struct {
	Snapshot           environment.Snapshot              `json:"snapshot"`
	SchemaBundle       *environment.SchemaBundle         `json:"schema_bundle,omitempty"`
	QueryScope         environment.CaptureScope          `json:"query_scope"`
	InputBindings      []compatibility.ResolutionBinding `json:"input_bindings"`
	DependencyBindings []closure.Binding                 `json:"dependency_bindings,omitempty"`
}
type Request struct {
	SchemaVersion int                    `json:"schema_version"`
	Document      analysis.QueryDocument `json:"document"`
	Resolutions   []Resolution           `json:"resolutions"`
	MaxVariants   *uint64                `json:"max_variants,omitempty"`
	Compatibility CompatibilityInputs    `json:"compatibility"`
}
type PreparedRequest struct {
	SchemaVersion int                                `json:"schema_version"`
	Document      analysis.QueryDocument             `json:"document"`
	Resolutions   []Resolution                       `json:"resolutions"`
	MaxVariants   *uint64                            `json:"max_variants,omitempty"`
	Compatibility compatibility.ResolutionAssessment `json:"compatibility"`
}
type Diagnostic struct {
	Code        string             `json:"code"`
	Message     string             `json:"message"`
	Location    *analysis.Location `json:"location,omitempty"`
	Placeholder string             `json:"placeholder,omitempty"`
}
type Provenance struct {
	QueryDigest             string `json:"query_digest"`
	CapabilityRevision      string `json:"capability_revision"`
	AnalysisContractVersion int    `json:"analysis_contract_version"`
	RequirementSetVersion   int    `json:"requirement_set_version"`
	EnvironmentDigest       string `json:"environment_digest"`
	SchemaBundleDigest      string `json:"schema_bundle_digest,omitempty"`
	ResolutionInputDigest   string `json:"resolution_input_digest"`
	AssessmentInputDigest   string `json:"assessment_input_digest"`
}
type Variant struct {
	ID                string                           `json:"id"`
	Ordinal           uint64                           `json:"ordinal"`
	Selection         []analysis.ResolutionChoice      `json:"selection"`
	Outcome           string                           `json:"outcome"`
	CandidateText     *string                          `json:"candidate_text,omitempty"`
	ResolvedQuery     *string                          `json:"resolved_query,omitempty"`
	Changes           []analysis.ResolutionChange      `json:"changes"`
	CandidateAnalysis *analysis.Result                 `json:"candidate_analysis,omitempty"`
	Proof             analysis.ResolutionProofEvidence `json:"proof"`
	Compatibility     *compatibility.ResolutionReport  `json:"compatibility,omitempty"`
	Diagnostics       []Diagnostic                     `json:"diagnostics"`
	Provenance        Provenance                       `json:"provenance"`
}
type Counts struct {
	Verified   uint64 `json:"verified"`
	Failed     uint64 `json:"failed"`
	Incomplete uint64 `json:"incomplete"`
}
type Report struct {
	SchemaVersion     int                         `json:"schema_version"`
	Original          analysis.ResolutionEvidence `json:"original"`
	Resolutions       []Resolution                `json:"resolutions"`
	MaxVariants       uint64                      `json:"max_variants"`
	TotalCombinations string                      `json:"total_combinations"`
	GeneratedCount    uint64                      `json:"generated_count"`
	Counts            Counts                      `json:"counts"`
	Variants          []Variant                   `json:"variants"`
	Provenance        Provenance                  `json:"provenance"`
}
type RequestErrorDetail struct {
	Code              string  `json:"code"`
	Path              string  `json:"path"`
	ByteOffset        *int    `json:"byte_offset,omitempty"`
	Message           string  `json:"message"`
	TotalCombinations string  `json:"total_combinations,omitempty"`
	MaxVariants       *uint64 `json:"max_variants,omitempty"`
}

// Prepared retains detached environment authority; query choices are call-local.
type Prepared struct {
	compatibility *compatibility.Prepared
	identity      compatibility.ArtifactIdentity
}
