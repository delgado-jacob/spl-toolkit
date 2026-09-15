// Package document exposes caller-owned views of canonical analysis evidence.
package document

import (
	"errors"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
)

// ErrInvalidRange reports a byte range that cannot describe an exact UTF-8
// source boundary in this snapshot.
var ErrInvalidRange = errors.New("invalid source range")

// RevisionContext identifies the canonical tool and refinement context that
// produced a snapshot. TargetDigest is empty only for analysis without a
// refinement target.
type RevisionContext struct {
	ToolVersion     string `json:"tool_version"`
	ContractVersion string `json:"contract_version"`
	TargetDigest    string `json:"target_digest,omitempty"`
}

// Snapshot is a deep-detached copy of canonical analysis evidence. Its IDs are
// meaningful only with this document and Revision. Callers may mutate a
// snapshot, but that neither changes engine state nor retains its attestation.
//
// Stages remain in canonical lexical source order; Stage.Position is the
// canonical logical position. Lineage retains explicit phase and execution
// order instead of inferring either from source order.
type Snapshot struct {
	SchemaVersion int                     `json:"schema_version"`
	Revision      RevisionContext         `json:"revision"`
	SourceHash    string                  `json:"source_hash"`
	Document      analysis.QueryDocument  `json:"document"`
	Status        analysis.Status         `json:"status"`
	Coverage      analysis.Coverage       `json:"coverage"`
	Stages        []analysis.Stage        `json:"stages"`
	Scopes        []analysis.Scope        `json:"scopes"`
	References    []analysis.Reference    `json:"references"`
	Lineage       []analysis.Lineage      `json:"lineage"`
	Dependencies  analysis.Dependencies   `json:"dependencies"`
	Diagnostics   []analysis.Diagnostic   `json:"diagnostics"`
	Requirements  analysis.RequirementSet `json:"requirements"`
}
