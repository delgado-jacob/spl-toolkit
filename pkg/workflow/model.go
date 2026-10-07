// Package workflow coordinates explicitly selected queries against offline evidence.
package workflow

import (
	"encoding/json"
	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/closure"
	"github.com/delgado-jacob/spl-toolkit/pkg/compatibility"
	"github.com/delgado-jacob/spl-toolkit/pkg/corpus"
	"github.com/delgado-jacob/spl-toolkit/pkg/environment"
	"github.com/delgado-jacob/spl-toolkit/pkg/graph"
	"github.com/delgado-jacob/spl-toolkit/pkg/resolution"
)

type Request struct {
	SchemaVersion int                      `json:"schema_version"`
	Documents     []corpus.RequestDocument `json:"documents"`
	Settings      Settings                 `json:"settings"`
	Format        string                   `json:"format,omitempty"`
}
type Settings struct {
	SchemaVersion int                       `json:"schema_version"`
	Snapshot      environment.Snapshot      `json:"snapshot"`
	SchemaBundle  *environment.SchemaBundle `json:"schema_bundle,omitempty"`
	Entries       []EntrySettings           `json:"entries"`
}
type EntrySettings struct {
	ID            string           `json:"id"`
	Compatibility *CheckSettings   `json:"compatibility,omitempty"`
	Resolution    *ResolveSettings `json:"resolution,omitempty"`
}
type CheckSettings struct {
	QueryScope         environment.CaptureScope     `json:"query_scope"`
	InputBindings      []compatibility.InputBinding `json:"input_bindings"`
	DependencyBindings []closure.Binding            `json:"dependency_bindings,omitempty"`
}
type ResolveSettings struct {
	Resolutions   []resolution.Resolution            `json:"resolutions"`
	MaxVariants   *uint64                            `json:"max_variants,omitempty"`
	Compatibility compatibility.ResolutionAssessment `json:"compatibility"`
}
type Failure struct {
	Phase   string          `json:"phase"`
	Code    string          `json:"code"`
	Message string          `json:"message"`
	Detail  json.RawMessage `json:"detail,omitempty"`
}
type RequestErrorDetail struct {
	Code       string `json:"code"`
	Path       string `json:"path"`
	ByteOffset *int   `json:"byte_offset,omitempty"`
	Message    string `json:"message"`
}
type ReportEntry struct {
	ID            string                `json:"id"`
	Origin        corpus.Origin         `json:"origin"`
	SourceHash    string                `json:"source_hash,omitempty"`
	Mode          string                `json:"mode"`
	Status        analysis.Status       `json:"status"`
	Analysis      *analysis.Result      `json:"analysis,omitempty"`
	Compatibility *compatibility.Report `json:"compatibility,omitempty"`
	Resolution    *resolution.Report    `json:"resolution,omitempty"`
	Failure       *Failure              `json:"failure,omitempty"`
}

// Assessed counts completed operations. Every selected entry has one content status;
// failed entries are incomplete, independently of their failure phase.
type Counts struct {
	Selected            int `json:"selected"`
	Assessed            int `json:"assessed"`
	AcquisitionFailed   int `json:"acquisition_failed"`
	ConfigurationFailed int `json:"configuration_failed"`
	InternalFailed      int `json:"internal_failed"`
	TraversalFailed     int `json:"traversal_failed"`
	Valid               int `json:"valid"`
	Invalid             int `json:"invalid"`
	Incomplete          int `json:"incomplete"`
}
type Provenance struct {
	ToolkitVersion     string `json:"toolkit_version"`
	VCSRevision        string `json:"vcs_revision,omitempty"`
	VCSModified        *bool  `json:"vcs_modified,omitempty"`
	EnvironmentDigest  string `json:"environment_digest"`
	SchemaBundleDigest string `json:"schema_bundle_digest,omitempty"`
	SettingsDigest     string `json:"settings_digest"`
}
type Report struct {
	SchemaVersion     int              `json:"schema_version"`
	Status            analysis.Status  `json:"status"`
	ExecutionComplete bool             `json:"execution_complete"`
	CIExitCode        int              `json:"ci_exit_code"`
	Selection         corpus.Selection `json:"selection"`
	Counts            Counts           `json:"counts"`
	Entries           []ReportEntry    `json:"entries"`
	Provenance        Provenance       `json:"provenance"`
}

// GraphSubject projects one original query or resolution candidate. All pointers
// address the input workflow Report, and IDs are local to its subject context.
type GraphSubject struct {
	DetectionID      string                      `json:"detection_id"`
	Domain           string                      `json:"domain"`
	VariantSelection []analysis.ResolutionChoice `json:"variant_selection,omitempty"`
	EvidencePointer  string                      `json:"evidence_pointer"`
	Analysis         *graph.Report               `json:"analysis,omitempty"`
	Closure          *closure.DependencyGraph    `json:"closure,omitempty"`
	ClosureCoverage  string                      `json:"closure_coverage"`
	CoverageReasons  []string                    `json:"coverage_reasons"`
	Failure          *Failure                    `json:"failure,omitempty"`
}
type GraphReport struct {
	SchemaVersion     int              `json:"schema_version"`
	Status            analysis.Status  `json:"status"`
	ExecutionComplete bool             `json:"execution_complete"`
	CIExitCode        int              `json:"ci_exit_code"`
	Selection         corpus.Selection `json:"selection"`
	Counts            Counts           `json:"counts"`
	Subjects          []GraphSubject   `json:"subjects"`
}
type BOMSubject struct {
	DetectionID      string                      `json:"detection_id"`
	Domain           string                      `json:"domain"`
	VariantSelection []analysis.ResolutionChoice `json:"variant_selection,omitempty"`
	EvidencePointer  string                      `json:"evidence_pointer"`
	Entries          []closure.BOMEntry          `json:"entries"`
	ClosureCoverage  string                      `json:"closure_coverage"`
	CoverageReasons  []string                    `json:"coverage_reasons"`
	Failure          *Failure                    `json:"failure,omitempty"`
}
type SharedDependency struct {
	EnvironmentDigest  string   `json:"environment_digest"`
	ObjectID           string   `json:"object_id"`
	OccurrencePointers []string `json:"occurrence_pointers"`
}
type BOMReport struct {
	SchemaVersion      int                `json:"schema_version"`
	Status             analysis.Status    `json:"status"`
	ExecutionComplete  bool               `json:"execution_complete"`
	CIExitCode         int                `json:"ci_exit_code"`
	Selection          corpus.Selection   `json:"selection"`
	Counts             Counts             `json:"counts"`
	Subjects           []BOMSubject       `json:"subjects"`
	SharedDependencies []SharedDependency `json:"shared_dependencies"`
}
