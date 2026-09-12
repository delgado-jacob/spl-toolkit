// Package corpus defines strict multi-document inputs and ordered report contracts.
package corpus

import (
	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/rewrite"
	"github.com/delgado-jacob/spl-toolkit/pkg/validation"
)

// ValidationTarget is the canonical M6 destination selector, not a new schema format.
type ValidationTarget = rewrite.ValidationTarget

type Origin struct {
	Kind         string `json:"kind"` // file or inline
	RelativePath string `json:"relative_path,omitempty"`
	BaseURI      string `json:"base_uri,omitempty"`
}

type AcquisitionError struct {
	Code    string `json:"code"`
	Phase   string `json:"phase"`
	Message string `json:"message"`
	ID      string `json:"id,omitempty"`
	Path    string `json:"path,omitempty"`
}

type Entry struct {
	ID         string                  `json:"id"`
	Origin     Origin                  `json:"origin"`
	SourceHash string                  `json:"source_hash,omitempty"`
	Document   *analysis.QueryDocument `json:"document,omitempty"`
	Failure    *AcquisitionError       `json:"failure,omitempty"`
}

type Selection struct {
	Mode              string             `json:"mode"`
	Complete          bool               `json:"complete"`
	IgnoredNames      []string           `json:"ignored_names"`
	SkippedSymlinks   []string           `json:"skipped_symlinks"`
	TraversalFailures []AcquisitionError `json:"traversal_failures"`
}

type Input struct {
	Entries   []Entry   `json:"entries"`
	Selection Selection `json:"selection"`
}

type ScanOptions struct {
	ValidationTarget *ValidationTarget `json:"validation_target,omitempty"`
}

type RequestDocument struct {
	ID       string                 `json:"id"`
	Document analysis.QueryDocument `json:"document"`
}

type Request struct {
	SchemaVersion    int               `json:"schema_version"`
	Documents        []RequestDocument `json:"documents"`
	ValidationTarget *ValidationTarget `json:"validation_target,omitempty"`
}

type Evaluation struct {
	Kind             string                   `json:"kind"`
	Analysis         *analysis.Result         `json:"analysis,omitempty"`
	FieldValidation  *validation.Report       `json:"field_validation,omitempty"`
	SchemaValidation *validation.SchemaReport `json:"schema_validation,omitempty"`
}

type ReportEntry struct {
	ID               string            `json:"id"`
	Origin           Origin            `json:"origin"`
	SourceHash       string            `json:"source_hash,omitempty"`
	AnalysisRevision string            `json:"analysis_revision,omitempty"`
	TargetDigest     string            `json:"target_digest,omitempty"`
	Evaluation       *Evaluation       `json:"evaluation,omitempty"`
	Failure          *AcquisitionError `json:"failure,omitempty"`
}

type Counts struct {
	Selected          int `json:"selected"`
	Analyzed          int `json:"analyzed"`
	AcquisitionFailed int `json:"acquisition_failed"`
	TraversalFailed   int `json:"traversal_failed"`
}

type CoverageCount struct {
	Complete     int  `json:"complete"`
	Incomplete   int  `json:"incomplete"`
	Denominator  int  `json:"denominator"`
	NotRequested bool `json:"not_requested,omitempty"`
}

type CoverageCounts struct {
	Syntax   CoverageCount `json:"syntax"`
	Semantic CoverageCount `json:"semantic"`
	Schema   CoverageCount `json:"schema"`
}

// StatusCounts count only documents with canonical analysis or validation reports.
type StatusCounts struct {
	Valid      int `json:"valid"`
	Invalid    int `json:"invalid"`
	Incomplete int `json:"incomplete"`
}

// CommandCoverage describes observed stage commands against the exact selected
// language capability manifest. An undeclared command has Declared=false.
type CommandCoverage struct {
	Language          string   `json:"language"`
	Profile           string   `json:"profile"`
	Version           string   `json:"version"`
	Command           string   `json:"command"`
	Count             int      `json:"count"`
	Declared          bool     `json:"declared"`
	SyntaxSupported   bool     `json:"syntax_supported"`
	SemanticSupported bool     `json:"semantic_supported"`
	Limitations       []string `json:"limitations"`
}

// ObservedReferenceForm counts canonical reference kind/role/resolution tuples.
// It does not claim the distinct M6 rewrite render-form capability.
type ObservedReferenceForm struct {
	Kind       string `json:"kind"`
	Role       string `json:"role"`
	Resolution string `json:"resolution"`
	Count      int    `json:"count"`
}

type DependencySummary struct {
	Kind         string   `json:"kind"`
	Name         string   `json:"name"`
	DocumentIDs  []string `json:"document_ids"`
	ReferenceIDs []string `json:"reference_ids"`
}

type Report struct {
	SchemaVersion          int                     `json:"schema_version"`
	Status                 analysis.Status         `json:"status"`
	ExecutionComplete      bool                    `json:"execution_complete"`
	Mode                   string                  `json:"mode"`
	Selection              Selection               `json:"selection"`
	Counts                 Counts                  `json:"counts"`
	StatusCounts           StatusCounts            `json:"status_counts"`
	Coverage               CoverageCounts          `json:"coverage"`
	CoverageReasons        []string                `json:"coverage_reasons"`
	CommandCoverage        []CommandCoverage       `json:"command_coverage"`
	ObservedReferenceForms []ObservedReferenceForm `json:"observed_reference_forms"`
	Dependencies           []DependencySummary     `json:"dependencies"`
	Entries                []ReportEntry           `json:"entries"`
}
