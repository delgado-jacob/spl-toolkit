// Package environment validates offline descriptions of captured Splunk environments.
package environment

import (
	"encoding/json"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/closure"
	"github.com/delgado-jacob/spl-toolkit/pkg/validation"
)

// Selector names an entire capture dimension or a finite set of exact values.
type Selector struct {
	All    *bool    `json:"all,omitempty"`
	Values []string `json:"values,omitempty"`
}

type CaptureScope struct {
	Namespace Selector `json:"namespace"`
	App       Selector `json:"app"`
	Owner     Selector `json:"owner"`
}

type Origin struct {
	InstanceID      string `json:"instance_id"`
	ProductVersion  string `json:"product_version"`
	Producer        string `json:"producer"`
	ProducerVersion string `json:"producer_version"`
}

type CaptureInterval struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

type Provenance struct {
	SourceKind string `json:"source_kind"`
	SourceID   string `json:"source_id"`
	ObservedAt string `json:"observed_at"`
}

type Capability struct {
	ID         string     `json:"id"`
	Version    string     `json:"version"`
	State      string     `json:"state"`
	Provenance Provenance `json:"provenance"`
}

type Collection struct {
	Kind     string `json:"kind"`
	Coverage string `json:"coverage"`
	Reason   string `json:"reason,omitempty"`
}

type Object struct {
	ID         string                  `json:"id"`
	Kind       string                  `json:"kind"`
	Name       string                  `json:"name"`
	Namespace  string                  `json:"namespace,omitempty"`
	App        string                  `json:"app,omitempty"`
	Owner      string                  `json:"owner,omitempty"`
	Sharing    string                  `json:"sharing,omitempty"`
	Provenance Provenance              `json:"provenance"`
	Document   *analysis.QueryDocument `json:"document,omitempty"`
	Arity      *int                    `json:"arity,omitempty"`
	Arguments  []string                `json:"arguments,omitempty"`
	EvalBased  *bool                   `json:"eval_based,omitempty"`
	Validation *string                 `json:"validation,omitempty"`
	Relations  []closure.Relation      `json:"relations,omitempty"`
}

type Snapshot struct {
	SchemaVersion int             `json:"schema_version"`
	ScopeID       string          `json:"scope_id"`
	Digest        string          `json:"digest,omitempty"`
	CaptureScope  CaptureScope    `json:"capture_scope"`
	Origin        Origin          `json:"origin"`
	Capture       CaptureInterval `json:"capture"`
	Capabilities  []Capability    `json:"capabilities"`
	Collections   []Collection    `json:"collections"`
	Objects       []Object        `json:"objects"`
}

// SchemaBundle records offline schema evidence independently of a snapshot.
type SchemaBundle struct {
	SchemaVersion int             `json:"schema_version"`
	BundleID      string          `json:"bundle_id"`
	Digest        string          `json:"digest,omitempty"`
	Provenance    Provenance      `json:"provenance"`
	Schemas       []SchemaEntry   `json:"schemas"`
	Bindings      []SchemaBinding `json:"bindings"`
}

type SchemaEntry struct {
	ID         string          `json:"id"`
	Kind       string          `json:"kind"`
	Catalog    json.RawMessage `json:"catalog,omitempty"`
	Target     json.RawMessage `json:"target,omitempty"`
	Provenance Provenance      `json:"provenance"`
}

type ObjectIdentity struct {
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
	App       string `json:"app,omitempty"`
	Owner     string `json:"owner,omitempty"`
}

type SchemaBinding struct {
	SchemaID       string         `json:"schema_id"`
	ObjectID       string         `json:"object_id"`
	Expected       ObjectIdentity `json:"expected"`
	SourceCoverage string         `json:"source_coverage"`
	Reason         string         `json:"reason,omitempty"`
}

type preparedBundleTarget struct {
	field  *validation.PreparedFieldCatalog
	schema *validation.PreparedSchemaTarget
}

// PreparedSchemaBundle owns normalized evidence and reusable compiled targets.
type PreparedSchemaBundle struct {
	bundle  SchemaBundle
	report  Report
	targets map[string]preparedBundleTarget
}

type CoverageEntry struct {
	Artifact string `json:"artifact"`
	Kind     string `json:"kind"`
	Coverage string `json:"coverage"`
	Reason   string `json:"reason,omitempty"`
}

type Diagnostic struct {
	Code       string `json:"code"`
	Severity   string `json:"severity"`
	Artifact   string `json:"artifact"`
	Path       string `json:"path"`
	ByteOffset *int   `json:"byte_offset,omitempty"`
	Message    string `json:"message"`
}

type Report struct {
	SchemaVersion      int             `json:"schema_version"`
	Status             string          `json:"status"`
	SnapshotDigest     string          `json:"snapshot_digest,omitempty"`
	SchemaBundleDigest string          `json:"schema_bundle_digest,omitempty"`
	Coverage           []CoverageEntry `json:"coverage"`
	Diagnostics        []Diagnostic    `json:"diagnostics"`
}

// PreparedSnapshot owns a detached, normalized snapshot.
type PreparedSnapshot struct {
	snapshot Snapshot
	report   Report
}

func (p *PreparedSnapshot) Snapshot() Snapshot {
	if p == nil {
		return Snapshot{}
	}
	return cloneSnapshot(p.snapshot)
}
func (p *PreparedSnapshot) Report() Report {
	if p == nil {
		return Report{}
	}
	out := p.report
	out.Coverage = append([]CoverageEntry{}, out.Coverage...)
	out.Diagnostics = append([]Diagnostic{}, out.Diagnostics...)
	return out
}
