package splunkexport

import "github.com/delgado-jacob/spl-toolkit/pkg/environment"

// Acquisition is the report's normalized collection coverage.
type Acquisition struct {
	Kind     string `json:"kind"`
	Coverage string `json:"coverage"`
	Reason   string `json:"reason,omitempty"`
}
type ExportLimits struct {
	RequestTimeoutSeconds float64 `json:"request_timeout_seconds"`
	JobTimeoutSeconds     float64 `json:"job_timeout_seconds"`
	OverallTimeoutSeconds float64 `json:"overall_timeout_seconds"`
	MaxRows               int     `json:"max_rows"`
	MaxResponseBytes      int64   `json:"max_response_bytes"`
	MaxArtifactBytes      int64   `json:"max_artifact_bytes"`
}
type ExportReport struct {
	SchemaVersion  int                           `json:"schema_version"`
	Status         string                        `json:"status"`
	ScopeID        string                        `json:"scope_id"`
	AllowInsecure  bool                          `json:"allow_insecure"`
	Origin         environment.Origin            `json:"origin"`
	Capture        environment.CaptureInterval   `json:"capture"`
	SnapshotDigest string                        `json:"snapshot_digest"`
	Observation    *environment.ObservationScope `json:"observation"`
	Limits         ExportLimits                  `json:"limits"`
	Collections    []Acquisition                 `json:"collections"`
	Diagnostics    []Diagnostic                  `json:"diagnostics"`
}
type Result struct {
	Snapshot environment.Snapshot
	Report   ExportReport
	ExitCode int
}

func exitFor(report *environment.Report, cleanupFailed bool) int {
	if report == nil || report.Status != "valid" && report.Status != "partial" {
		return 2
	}
	if report.Status == "partial" || cleanupFailed {
		return 3
	}
	return 0
}
