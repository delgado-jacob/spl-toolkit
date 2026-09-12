// Package graph projects canonical corpus evidence into a versioned, static graph.
// It does not resolve live dependencies or infer additional lineage.
package graph

import (
	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/corpus"
)

// Report is graph JSON v1. Array order follows corpus order, lexical stage and
// reference order, then canonical lineage order; it is deterministic for an
// unchanged corpus report. ReportPointer is an RFC 6901 pointer into that report.
type Report struct {
	SchemaVersion     int                   `json:"schema_version"`
	Status            analysis.Status       `json:"status"`
	ExecutionComplete bool                  `json:"execution_complete"`
	Mode              string                `json:"mode"`
	Coverage          corpus.CoverageCounts `json:"coverage"`
	CoverageReasons   []string              `json:"coverage_reasons"`
	Nodes             []Node                `json:"nodes"`
	Edges             []Edge                `json:"edges"`
}

// Node preserves the evidence needed for inspection without giving a resource
// name the meaning of a verified catalog entity. Source locations are evidence,
// never persistent symbol identity.
type Node struct {
	ID               string                   `json:"id"`
	Kind             string                   `json:"kind"`
	ReportPointer    string                   `json:"report_pointer"`
	DocumentID       string                   `json:"document_id,omitempty"`
	AnalysisRevision string                   `json:"analysis_revision,omitempty"`
	TargetDigest     string                   `json:"target_digest,omitempty"`
	SourceHash       string                   `json:"source_hash,omitempty"`
	Failure          *corpus.AcquisitionError `json:"failure,omitempty"`
	CanonicalID      string                   `json:"canonical_id,omitempty"`
	ScopeKind        string                   `json:"scope_kind,omitempty"`
	ScopeID          string                   `json:"scope_id,omitempty"`
	StageID          string                   `json:"stage_id,omitempty"`
	Command          string                   `json:"command,omitempty"`
	Position         *int                     `json:"position,omitempty"`
	SemanticComplete *bool                    `json:"semantic_complete,omitempty"`
	Name             string                   `json:"name,omitempty"`
	OriginalName     string                   `json:"original_name,omitempty"`
	DependencyKind   string                   `json:"dependency_kind,omitempty"`
	Role             string                   `json:"role,omitempty"`
	Resolution       string                   `json:"resolution,omitempty"`
	Binding          string                   `json:"binding,omitempty"`
	Location         *analysis.Location       `json:"location,omitempty"`
	Phase            string                   `json:"phase,omitempty"`
	ExecutionOrder   *int                     `json:"execution_order,omitempty"`
	Occurrence       *int                     `json:"occurrence,omitempty"`
	Operation        string                   `json:"operation,omitempty"`
	Conditional      bool                     `json:"conditional,omitempty"`
	Before           *analysis.FieldState     `json:"before,omitempty"`
	After            *analysis.FieldState     `json:"after,omitempty"`
}

// Edges are only direct canonical relations. No edge claims completeness when
// the source analysis is incomplete or a transition has unknown membership.
type Edge struct {
	ID               string `json:"id"`
	From             string `json:"from"`
	To               string `json:"to"`
	Relation         string `json:"relation"`
	ReportPointer    string `json:"report_pointer"`
	DocumentID       string `json:"document_id,omitempty"`
	AnalysisRevision string `json:"analysis_revision,omitempty"`
	TargetDigest     string `json:"target_digest,omitempty"`
	Conditional      bool   `json:"conditional,omitempty"`
}
