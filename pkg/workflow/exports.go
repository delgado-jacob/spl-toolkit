package workflow

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/closure"
	"github.com/delgado-jacob/spl-toolkit/pkg/corpus"
	"github.com/delgado-jacob/spl-toolkit/pkg/graph"
)

type exportSubject struct {
	id, domain, pointer, analysisPointer, closurePointer string
	selection                                            []analysis.ResolutionChoice
	analysis                                             *analysis.Result
	closure                                              *closure.Report
	failure                                              *Failure
}

// exportSubjects follows detection order, then original and canonical variant order.
// Unsuccessful variants remain subjects even when no candidate analysis is available.
func exportSubjects(r *Report) ([]exportSubject, error) {
	if r == nil || r.SchemaVersion != 1 {
		return nil, fmt.Errorf("workflow: report v1 is required")
	}
	subjects := []exportSubject{}
	for i, e := range r.Entries {
		base := fmt.Sprintf("/entries/%d", i)
		s := exportSubject{id: e.ID, domain: "original", pointer: base, analysisPointer: base + "/analysis", analysis: e.Analysis, failure: e.Failure}
		if e.Compatibility != nil {
			s.closure = e.Compatibility.Closure
			s.closurePointer = base + "/compatibility/closure"
		}
		subjects = append(subjects, s)
		if e.Resolution == nil {
			continue
		}
		for j, v := range e.Resolution.Variants {
			vp := fmt.Sprintf("%s/resolution/variants/%d", base, j)
			s := exportSubject{id: e.ID, domain: "candidate", pointer: vp, analysisPointer: vp + "/candidate_analysis", analysis: v.CandidateAnalysis, selection: v.Selection}
			if v.Compatibility != nil {
				s.closure = v.Compatibility.Closure
				s.closurePointer = vp + "/compatibility/closure"
			}
			subjects = append(subjects, s)
		}
	}
	return subjects, nil
}

// ExportGraph projects retained canonical evidence only; it never analyzes,
// resolves, or expands queries. All evidence pointers address source.
func ExportGraph(source *Report) (*GraphReport, error) {
	subjects, err := exportSubjects(source)
	if err != nil {
		return nil, err
	}
	out := &GraphReport{SchemaVersion: 1, Status: source.Status, ExecutionComplete: source.ExecutionComplete, CIExitCode: source.CIExitCode, Selection: source.Selection, Counts: source.Counts, Subjects: []GraphSubject{}}
	for _, s := range subjects {
		coverage, reasons := exportClosureCoverage(s)
		subject := GraphSubject{DetectionID: s.id, Domain: s.domain, VariantSelection: s.selection, EvidencePointer: s.pointer, ClosureCoverage: coverage, CoverageReasons: reasons, Failure: s.failure}
		prefix := s.pointer + ":"
		if s.analysis != nil {
			subject.Analysis, err = exportAnalysis(s, prefix)
			if err != nil {
				return nil, fmt.Errorf("workflow: %s: %w", s.analysisPointer, err)
			}
		}
		if s.closure != nil {
			// JSON detaches nested occurrence/origin arrays before remapping local links.
			c, err := exportCopy(s.closure.Graph)
			if err != nil {
				return nil, err
			}
			for i := range c.Nodes {
				n := &c.Nodes[i]
				n.ID = prefix + n.ID
				n.EvidencePointer = s.closurePointer + n.EvidencePointer
			}
			for i := range c.Edges {
				e := &c.Edges[i]
				e.ID = prefix + e.ID
				e.From = prefix + e.From
				if e.To != "" {
					e.To = prefix + e.To
				}
				e.TraversalPointer = s.closurePointer + e.TraversalPointer
				for j := range e.InvocationNodeIDs {
					e.InvocationNodeIDs[j] = prefix + e.InvocationNodeIDs[j]
				}
			}
			subject.Closure = &c
		}
		out.Subjects = append(out.Subjects, subject)
	}
	return detachedExport(out)
}

// ExportBOM preserves captured object identity and canonical occurrence order.
// Shared dependencies are indexed by environment digest and captured ObjectID,
// never by a resource display name. Occurrence pointers address the source BOM.
func ExportBOM(source *Report) (*BOMReport, error) {
	subjects, err := exportSubjects(source)
	if err != nil {
		return nil, err
	}
	out := &BOMReport{SchemaVersion: 1, Status: source.Status, ExecutionComplete: source.ExecutionComplete, CIExitCode: source.CIExitCode, Selection: source.Selection, Counts: source.Counts, Subjects: []BOMSubject{}, SharedDependencies: []SharedDependency{}}
	index := map[string]int{}
	for _, s := range subjects {
		coverage, reasons := exportClosureCoverage(s)
		subject := BOMSubject{DetectionID: s.id, Domain: s.domain, VariantSelection: s.selection, EvidencePointer: s.pointer, Entries: []closure.BOMEntry{}, ClosureCoverage: coverage, CoverageReasons: reasons, Failure: s.failure}
		if s.closure != nil {
			subject.Entries, err = exportCopy(s.closure.BOM)
			if err != nil {
				return nil, err
			}
			for i := range subject.Entries {
				for j := range subject.Entries[i].Occurrences {
					subject.Entries[i].Occurrences[j].EdgeID = s.pointer + ":" + subject.Entries[i].Occurrences[j].EdgeID
				}
			}
			for i, entry := range s.closure.BOM {
				key := source.Provenance.EnvironmentDigest + "\x00" + entry.ObjectID
				at, ok := index[key]
				if !ok {
					at = len(out.SharedDependencies)
					index[key] = at
					out.SharedDependencies = append(out.SharedDependencies, SharedDependency{EnvironmentDigest: source.Provenance.EnvironmentDigest, ObjectID: entry.ObjectID, OccurrencePointers: []string{}})
				}
				for j := range entry.Occurrences {
					out.SharedDependencies[at].OccurrencePointers = append(out.SharedDependencies[at].OccurrencePointers, fmt.Sprintf("%s/bom/%d/occurrences/%d", s.closurePointer, i, j))
				}
			}
		}
		out.Subjects = append(out.Subjects, subject)
	}
	sort.Slice(out.SharedDependencies, func(i, j int) bool {
		a, b := out.SharedDependencies[i], out.SharedDependencies[j]
		if a.EnvironmentDigest != b.EnvironmentDigest {
			return a.EnvironmentDigest < b.EnvironmentDigest
		}
		return a.ObjectID < b.ObjectID
	})
	return detachedExport(out)
}

func exportClosureCoverage(s exportSubject) (string, []string) {
	if s.closure != nil {
		status := "incomplete"
		if s.closure.Coverage.Complete {
			status = "complete"
		}
		return status, append([]string{}, s.closure.Coverage.Reasons...)
	}
	if s.analysis == nil {
		return "unavailable", []string{"analysis_unavailable"}
	}
	// Only knowledge-object kinds belong to closure. Direct source references
	// retain their analysis graph but do not invent captured knowledge objects.
	for _, ref := range s.analysis.References {
		switch ref.Kind {
		case "macro", "lookup", "dataset", "data_model", "saved_search":
			return "unavailable", []string{"closure_unavailable"}
		}
	}
	return "not_applicable", []string{}
}

// exportAnalysis adapts exactly one retained analysis to graph's corpus owner.
// The current revision is valid only when the recorded capability identity agrees.
func exportAnalysis(s exportSubject, prefix string) (*graph.Report, error) {
	a := s.analysis
	revision, err := analysis.CapabilityRevisionFor(analysis.CapabilityOptions{Language: a.Document.Language, Profile: a.Document.Profile, Version: a.Document.Version})
	if err != nil || revision != a.Requirements.CapabilityRevision {
		return nil, fmt.Errorf("recorded capability revision differs from current canonical capabilities")
	}
	ar, err := corpus.AnalysisRevision(a.Document)
	if err != nil {
		return nil, err
	}
	coverage := corpus.CoverageCounts{Schema: corpus.CoverageCount{NotRequested: true}}
	for _, item := range []struct {
		count    *corpus.CoverageCount
		complete bool
	}{{&coverage.Syntax, a.Coverage.SyntaxComplete}, {&coverage.Semantic, a.Coverage.SemanticComplete}} {
		item.count.Denominator = 1
		if item.complete {
			item.count.Complete = 1
		} else {
			item.count.Incomplete = 1
		}
	}
	c := &corpus.Report{SchemaVersion: 1, Status: a.Status, ExecutionComplete: true, Mode: "analysis", Coverage: coverage, CoverageReasons: append([]string{}, a.Coverage.Reasons...), Counts: corpus.Counts{Selected: 1, Analyzed: 1}, Entries: []corpus.ReportEntry{{ID: s.id, SourceHash: corpus.SourceHash(a.Document.Text), AnalysisRevision: ar, Evaluation: &corpus.Evaluation{Kind: "analysis", Analysis: a}}}, Dependencies: []corpus.DependencySummary{}}
	known := map[string][]string{"index": a.Dependencies.Indexes, "source": a.Dependencies.Sources, "sourcetype": a.Dependencies.SourceTypes, "dataset": a.Dependencies.Datasets, "lookup": a.Dependencies.Lookups, "data_model": a.Dependencies.DataModels, "macro": a.Dependencies.Macros}
	firstReference := map[string]int{}
	for i, ref := range a.References {
		exact := false
		for _, name := range known[ref.Kind] {
			if name == ref.NormalizedName {
				exact = true
				break
			}
		}
		if !exact || ref.Resolution != "exact" {
			continue
		}
		at := -1
		for j, d := range c.Dependencies {
			if d.Kind == ref.Kind && d.Name == ref.NormalizedName {
				at = j
				break
			}
		}
		if at == -1 {
			at = len(c.Dependencies)
			c.Dependencies = append(c.Dependencies, corpus.DependencySummary{Kind: ref.Kind, Name: ref.NormalizedName, DocumentIDs: []string{s.id}, ReferenceIDs: []string{}})
			firstReference[ref.Kind+"\x00"+ref.NormalizedName] = i
		}
		c.Dependencies[at].ReferenceIDs = append(c.Dependencies[at].ReferenceIDs, ref.ID)
	}
	sort.Slice(c.Dependencies, func(i, j int) bool {
		a, b := c.Dependencies[i], c.Dependencies[j]
		if a.Kind != b.Kind {
			return a.Kind < b.Kind
		}
		return a.Name < b.Name
	})
	g, err := graph.Export(c)
	if err != nil {
		return nil, err
	}
	rebase := func(ptr string) (string, error) {
		const ap = "/entries/0/evaluation/analysis"
		if ptr == "/entries/0" {
			return s.pointer, nil
		}
		if strings.HasPrefix(ptr, ap) {
			return s.analysisPointer + strings.TrimPrefix(ptr, ap), nil
		}
		for i, d := range c.Dependencies {
			if ptr == fmt.Sprintf("/dependencies/%d", i) {
				return fmt.Sprintf("%s/references/%d", s.analysisPointer, firstReference[d.Kind+"\x00"+d.Name]), nil
			}
		}
		return "", fmt.Errorf("unmapped corpus evidence pointer %q", ptr)
	}
	for i := range g.Nodes {
		n := &g.Nodes[i]
		n.ID = prefix + n.ID
		n.ReportPointer, err = rebase(n.ReportPointer)
		if err != nil {
			return nil, err
		}
	}
	for i := range g.Edges {
		e := &g.Edges[i]
		e.ID = prefix + e.ID
		e.From = prefix + e.From
		e.To = prefix + e.To
		e.ReportPointer, err = rebase(e.ReportPointer)
		if err != nil {
			return nil, err
		}
	}
	// Includes field-state pointer members that the owner shallowly copies.
	return detachedExport(g)
}

func exportCopy[T any](value T) (T, error) {
	var out T
	raw, err := json.Marshal(value)
	if err != nil {
		return out, err
	}
	err = json.Unmarshal(raw, &out)
	return out, err
}
func detachedExport[T any](value *T) (*T, error) {
	copy, err := exportCopy(*value)
	if err != nil {
		return nil, err
	}
	return &copy, nil
}
