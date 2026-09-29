package closure

import (
	"fmt"
	"sort"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
)

// Report retains both the caller's exact source analysis and the macro-expanded analysis.
type Report struct {
	SchemaVersion      int                               `json:"schema_version"`
	Query              analysis.RequirementQueryIdentity `json:"query"`
	BundleDigest       string                            `json:"bundle_digest"`
	ScopeID            string                            `json:"scope_id"`
	Status             analysis.Status                   `json:"status"`
	DirectAnalysis     *analysis.Result                  `json:"direct_analysis"`
	DirectRequirements analysis.RequirementSet           `json:"direct_requirements"`
	EffectiveAnalysis  *analysis.Result                  `json:"effective_analysis"`
	DefinitionAnalyses []DefinitionAnalysis              `json:"definition_analyses"`
	Provenance         []ProvenanceSegment               `json:"provenance"`
	Coverage           ClosureCoverage                   `json:"coverage"`
	Gaps               []ClosureGap                      `json:"gaps"`
	Diagnostics        []ClosureDiagnostic               `json:"diagnostics"`
	Traversal          []TraversalEdge                   `json:"traversal"`
}
type DefinitionAnalysis struct {
	ObjectID          string           `json:"object_id"`
	DirectAnalysis    *analysis.Result `json:"direct_analysis"`
	EffectiveAnalysis *analysis.Result `json:"effective_analysis"`
}
type SourceInterval struct {
	Kind     string `json:"kind"`
	SourceID string `json:"source_id"`
	ObjectID string `json:"object_id,omitempty"`
	Start    int    `json:"start"`
	End      int    `json:"end"`
}
type InvocationFrame struct {
	ObjectID   string           `json:"object_id"`
	InstanceID string           `json:"instance_id"`
	Invocation []SourceInterval `json:"invocation"`
}
type ProvenanceSegment struct {
	EffectiveStart  int               `json:"effective_start"`
	EffectiveEnd    int               `json:"effective_end"`
	Source          SourceInterval    `json:"source"`
	InvocationChain []InvocationFrame `json:"invocation_chain"`
	Placeholder     *SourceInterval   `json:"placeholder,omitempty"`
	Placeholders    []SourceInterval  `json:"placeholders"`
}
type ClosureCoverage struct {
	Complete             bool     `json:"complete"`
	EffectiveQuery       bool     `json:"effective_query"`
	TraversedDefinitions bool     `json:"traversed_definitions"`
	Resolution           bool     `json:"resolution"`
	Collections          bool     `json:"collections"`
	Expansion            bool     `json:"expansion"`
	Reasons              []string `json:"reasons"`
}
type ClosureGap struct {
	Code     string         `json:"code"`
	Kind     string         `json:"kind,omitempty"`
	Name     string         `json:"name,omitempty"`
	Source   SourceInterval `json:"source"`
	Property string         `json:"property,omitempty"`
	Path     []string       `json:"path"`
}
type ClosureDiagnostic struct {
	Source     SourceInterval      `json:"source"`
	Diagnostic analysis.Diagnostic `json:"diagnostic"`
}
type TraversalEdge struct {
	ID              string            `json:"id"`
	FromObjectID    string            `json:"from_object_id,omitempty"`
	Kind            string            `json:"kind"`
	Name            string            `json:"name"`
	Source          SourceInterval    `json:"source"`
	Property        string            `json:"property,omitempty"`
	ReferenceID     string            `json:"reference_id,omitempty"`
	Resolution      string            `json:"resolution"`
	ToObjectID      string            `json:"to_object_id,omitempty"`
	Path            []string          `json:"path"`
	CyclePath       []string          `json:"cycle_path"`
	InvocationChain []InvocationFrame `json:"invocation_chain"`
	Origins         []SourceInterval  `json:"origins"`
}

func publicInterval(s sourceInterval) SourceInterval {
	return SourceInterval{Kind: s.Kind, SourceID: s.SourceID, ObjectID: s.ObjectID, Start: s.Start, End: s.End}
}
func publicFrames(in []invocationFrame) []InvocationFrame {
	out := make([]InvocationFrame, 0, len(in))
	for _, f := range in {
		frame := InvocationFrame{ObjectID: f.ObjectID, InstanceID: f.InstanceID, Invocation: []SourceInterval{}}
		for _, v := range f.Invocation {
			frame.Invocation = append(frame.Invocation, publicInterval(v))
		}
		out = append(out, frame)
	}
	return out
}
func publicProvenance(in []provenanceSegment) []ProvenanceSegment {
	out := make([]ProvenanceSegment, 0, len(in))
	for _, s := range in {
		p := ProvenanceSegment{EffectiveStart: s.EffectiveStart, EffectiveEnd: s.EffectiveEnd, Source: publicInterval(s.Source), InvocationChain: []InvocationFrame{}, Placeholders: []SourceInterval{}}
		if s.Placeholder != nil {
			v := publicInterval(*s.Placeholder)
			p.Placeholder = &v
		}
		for _, v := range s.Placeholders {
			p.Placeholders = append(p.Placeholders, publicInterval(v))
		}
		p.InvocationChain = publicFrames(s.InvocationChain)
		out = append(out, p)
	}
	return out
}

type evaluator struct {
	req              Request
	report           *Report
	objects          map[string]Definition
	collections      map[string]string
	directCache      map[string]*analysis.Result
	active           []string
	bodyDone         map[string]bool
	expandedByResult map[*analysis.Result]expansion
}

// Evaluate validates every binding against its original document before returning content findings.
func Evaluate(input Request) (*Report, error) {
	if input.SchemaVersion != 1 {
		return nil, inputError("schema_version must be integer 1")
	}
	document, err := normalizeDocument(input.Document)
	if err != nil {
		return nil, err
	}
	bundle, err := normalizeBundle(input.Bundle)
	if err != nil {
		return nil, err
	}
	digest, err := BundleDigest(bundle)
	if err != nil {
		return nil, err
	}
	req := Request{SchemaVersion: 1, Document: document, Bundle: bundle, Bindings: append([]Binding{}, input.Bindings...)}
	direct, err := analysis.Analyze(document)
	if err != nil {
		return nil, inputError("analyze document: %v", err)
	}
	cache, err := validateBindings(req, direct)
	if err != nil {
		return nil, err
	}
	expanded := expandMacros(req)
	effectiveDocument := document
	effectiveDocument.Text = expanded.Text
	effective, err := analysis.Analyze(effectiveDocument)
	if err != nil {
		return nil, inputError("analyze effective document: %v", err)
	}
	report := &Report{SchemaVersion: 1, Query: direct.Requirements.Query, BundleDigest: digest, ScopeID: bundle.ScopeID, DirectAnalysis: direct, DirectRequirements: direct.Requirements, EffectiveAnalysis: effective, DefinitionAnalyses: []DefinitionAnalysis{}, Provenance: publicProvenance(expanded.Segments), Coverage: ClosureCoverage{EffectiveQuery: true, TraversedDefinitions: true, Resolution: true, Collections: true, Expansion: true, Reasons: []string{}}, Gaps: []ClosureGap{}, Diagnostics: []ClosureDiagnostic{}, Traversal: []TraversalEdge{}}
	e := &evaluator{req: req, report: report, objects: map[string]Definition{}, collections: map[string]string{}, directCache: cache, bodyDone: map[string]bool{}, expandedByResult: map[*analysis.Result]expansion{effective: expanded}}
	for _, o := range bundle.Objects {
		e.objects[o.ID] = o
	}
	for _, c := range bundle.Collections {
		e.collections[c.Kind] = c.Coverage
	}
	e.inspectExpansion(expanded, sourceInterval{Kind: "query", SourceID: document.SourceID, Start: 0, End: len(document.Text)}, effective, true)
	e.inspectOriginalMacros(document, sourceInterval{Kind: "query", SourceID: document.SourceID, Start: 0, End: len(document.Text)}, nil)
	e.finalize()
	return report, nil
}

func (e *evaluator) addGap(g ClosureGap, dimension string) {
	if g.Path == nil {
		g.Path = []string{}
	}
	e.report.Gaps = append(e.report.Gaps, g)
	switch dimension {
	case "effective":
		e.report.Coverage.EffectiveQuery = false
	case "definitions":
		e.report.Coverage.TraversedDefinitions = false
	case "resolution":
		e.report.Coverage.Resolution = false
	case "collections":
		e.report.Coverage.Collections = false
	case "expansion":
		e.report.Coverage.Expansion = false
	}
}
func (e *evaluator) addEdge(edge TraversalEdge) {
	edge.ID = fmt.Sprintf("edge-%d", len(e.report.Traversal)+1)
	if edge.Path == nil {
		edge.Path = []string{}
	}
	if edge.CyclePath == nil {
		edge.CyclePath = []string{}
	}
	if edge.InvocationChain == nil {
		edge.InvocationChain = []InvocationFrame{}
	}
	if edge.Origins == nil {
		edge.Origins = []SourceInterval{}
	}
	e.report.Traversal = append(e.report.Traversal, edge)
}
func (e *evaluator) finalize() {
	coverage := &e.report.Coverage
	reasons := map[string]bool{}
	for _, g := range e.report.Gaps {
		if !reasons[g.Code] {
			coverage.Reasons = append(coverage.Reasons, g.Code)
			reasons[g.Code] = true
		}
	}
	sort.Strings(coverage.Reasons)
	coverage.Complete = coverage.EffectiveQuery && coverage.TraversedDefinitions && coverage.Resolution && coverage.Collections && coverage.Expansion && len(e.report.Gaps) == 0
	e.report.Status = analysis.Valid
	if !coverage.Complete {
		e.report.Status = analysis.Incomplete
	}
	if e.definiteInvalid() {
		e.report.Status = analysis.Invalid
		coverage.Complete = false
	}
}
func (e *evaluator) definiteInvalid() bool {
	for result, expanded := range e.expandedByResult {
		for _, d := range result.Diagnostics {
			if d.Severity != "error" {
				continue
			}
			opaque := false
			for _, gap := range expanded.Gaps {
				if d.Location.Start.Offset < gap.EffectiveEnd && d.Location.End.Offset > gap.EffectiveStart {
					opaque = true
					break
				}
			}
			if !opaque {
				return true
			}
		}
	}
	return false
}
