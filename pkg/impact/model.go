// Package impact compares canonical static evidence for one immutable corpus.
// It does not execute queries or write migration candidates to source files.
package impact

import (
	"encoding/json"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/corpus"
	"github.com/delgado-jacob/spl-toolkit/pkg/rewrite"
)

type Classification string

const (
	Affected      Classification = "affected"
	Unchanged     Classification = "unchanged"
	Indeterminate Classification = "indeterminate"
	Failed        Classification = "failed"
)

// EvidenceDelta contains complete canonical values for a changed finding.
// Null sides mean introduction or resolution, not missing comparison data.
type EvidenceDelta struct {
	Category string          `json:"category"`
	Change   string          `json:"change"` // introduced, resolved, changed, unmatched, ambiguous
	Key      string          `json:"key"`
	Before   json.RawMessage `json:"before,omitempty"`
	After    json.RawMessage `json:"after,omitempty"`
}

type ReferencePair struct {
	BeforeID string `json:"before_id"`
	AfterID  string `json:"after_id"`
	Domain   string `json:"domain"` // original or candidate
	Basis    string `json:"basis"`
}

// Alignment never guesses candidate relationships from shifted coordinates.
type Alignment struct {
	Pairs     []ReferencePair `json:"pairs"`
	Unmatched []string        `json:"unmatched"`
	Ambiguous []string        `json:"ambiguous"`
}

type ImpactEntry struct {
	ID               string                   `json:"id"`
	Origin           corpus.Origin            `json:"origin"`
	SourceHash       string                   `json:"source_hash,omitempty"`
	AnalysisRevision string                   `json:"analysis_revision,omitempty"`
	Before           *corpus.Evaluation       `json:"before,omitempty"`
	After            *corpus.Evaluation       `json:"after,omitempty"`
	BeforeRewrite    *rewrite.Result          `json:"before_rewrite,omitempty"`
	AfterRewrite     *rewrite.Result          `json:"after_rewrite,omitempty"`
	Failure          *corpus.AcquisitionError `json:"failure,omitempty"`
	BeforeStatus     analysis.Status          `json:"before_status,omitempty"`
	AfterStatus      analysis.Status          `json:"after_status,omitempty"`
	Deltas           []EvidenceDelta          `json:"deltas"`
	Alignment        Alignment                `json:"alignment"`
	Classification   Classification           `json:"classification"`
	Reasons          []string                 `json:"reasons"`
}

type Counts struct {
	Selected        int `json:"selected"`
	Compared        int `json:"compared"`
	Denominator     int `json:"denominator"`
	TraversalFailed int `json:"traversal_failed"`
	Affected        int `json:"affected"`
	Unchanged       int `json:"unchanged"`
	Indeterminate   int `json:"indeterminate"`
	Failed          int `json:"failed"`
	Introduced      int `json:"introduced"`
	Resolved        int `json:"resolved"`
	Changed         int `json:"changed"`
	Unmatched       int `json:"unmatched"`
	Ambiguous       int `json:"ambiguous"`
}

type Report struct {
	SchemaVersion     int              `json:"schema_version"`
	Kind              string           `json:"kind"`
	ExecutionComplete bool             `json:"execution_complete"`
	Selection         corpus.Selection `json:"selection"`
	BeforeDigest      string           `json:"before_digest"`
	AfterDigest       string           `json:"after_digest"`
	Counts            Counts           `json:"counts"`
	Entries           []ImpactEntry    `json:"entries"`
}

func emptyEntry(entry corpus.Entry) ImpactEntry {
	return ImpactEntry{ID: entry.ID, Origin: entry.Origin, SourceHash: entry.SourceHash,
		Deltas: []EvidenceDelta{}, Alignment: Alignment{Pairs: []ReferencePair{}, Unmatched: []string{}, Ambiguous: []string{}}, Reasons: []string{}}
}

func (r *Report) append(entry ImpactEntry) {
	r.Entries = append(r.Entries, entry)
	switch entry.Classification {
	case Affected:
		r.Counts.Affected++
	case Unchanged:
		r.Counts.Unchanged++
	case Indeterminate:
		r.Counts.Indeterminate++
	case Failed:
		r.Counts.Failed++
	}
	if entry.Failure == nil {
		r.Counts.Compared++
	}
	for _, d := range entry.Deltas {
		switch d.Change {
		case "introduced":
			r.Counts.Introduced++
		case "resolved":
			r.Counts.Resolved++
		case "changed":
			r.Counts.Changed++
		case "unmatched":
			r.Counts.Unmatched++
		case "ambiguous":
			r.Counts.Ambiguous++
		}
	}
}
