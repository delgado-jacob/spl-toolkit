package graph

import (
	"fmt"
	"strconv"
	"unicode/utf8"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/corpus"
)

var dependencyKinds = map[string]bool{"index": true, "source": true, "sourcetype": true, "dataset": true, "lookup": true, "data_model": true, "macro": true}

// Export projects a corpus report without reinterpreting query text. It rejects
// dangling or contradictory external evidence instead of silently repairing it.
func Export(source *corpus.Report) (*Report, error) {
	if source == nil || source.SchemaVersion != 1 {
		return nil, fmt.Errorf("graph: corpus report v1 is required")
	}
	g := &Report{SchemaVersion: 1, Status: source.Status, ExecutionComplete: source.ExecutionComplete, Mode: source.Mode, Coverage: source.Coverage, CoverageReasons: append([]string{}, source.CoverageReasons...), Nodes: []Node{}, Edges: []Edge{}}
	b := builder{graph: g, source: source, nodes: map[string]bool{}, edges: map[string]bool{}, dependencies: map[string]corpus.DependencySummary{}, observedDependencies: map[string]*dependencyObservation{}}
	for i, summary := range source.Dependencies {
		if !dependencyKinds[summary.Kind] || summary.Name == "" {
			return nil, fmt.Errorf("graph: malformed dependency %d", i)
		}
		id := dependencyID(summary.Kind, summary.Name)
		if _, duplicate := b.dependencies[id]; duplicate {
			return nil, fmt.Errorf("graph: duplicate dependency %s", id)
		}
		b.dependencies[id] = summary
	}
	documents := map[string]bool{}
	for i, entry := range source.Entries {
		if entry.ID == "" || documents[entry.ID] {
			return nil, fmt.Errorf("graph: duplicate or empty document id at entry %d", i)
		}
		documents[entry.ID] = true
		if err := b.entry(i, entry); err != nil {
			return nil, err
		}
	}
	for i, summary := range source.Dependencies {
		id := dependencyID(summary.Kind, summary.Name)
		observation := b.observedDependencies[id]
		if observation == nil || len(observation.documents) != len(summary.DocumentIDs) || len(observation.references) == 0 {
			return nil, fmt.Errorf("graph: dependency summary %d does not match references", i)
		}
		for _, doc := range summary.DocumentIDs {
			if !observation.documents[doc] {
				return nil, fmt.Errorf("graph: dangling dependency document %q", doc)
			}
		}
		actual := map[string]int{}
		for _, ref := range summary.ReferenceIDs {
			actual[ref]++
		}
		if len(actual) != len(observation.references) {
			return nil, fmt.Errorf("graph: dependency reference count mismatch for %s", id)
		}
		for ref, count := range observation.references {
			if actual[ref] != count {
				return nil, fmt.Errorf("graph: dangling or duplicated dependency reference %q", ref)
			}
		}
		if err := b.addNode(Node{ID: id, Kind: "dependency", ReportPointer: pointer("dependencies", i), DependencyKind: summary.Kind, Name: summary.Name}); err != nil {
			return nil, err
		}
	}
	for _, link := range b.dependencyLinks {
		if err := b.addEdge(link); err != nil {
			return nil, err
		}
	}
	for _, edge := range g.Edges {
		if !b.nodes[edge.From] || !b.nodes[edge.To] {
			return nil, fmt.Errorf("graph: dangling edge %s", edge.ID)
		}
	}
	return g, nil
}

type builder struct {
	graph                *Report
	source               *corpus.Report
	nodes                map[string]bool
	edges                map[string]bool
	dependencies         map[string]corpus.DependencySummary
	observedDependencies map[string]*dependencyObservation
	dependencyLinks      []Edge
}
type dependencyObservation struct {
	documents  map[string]bool
	references map[string]int
}

func pointer(parts ...any) string {
	out := ""
	for _, p := range parts {
		out += "/" + fmt.Sprint(p)
	}
	return out
}
func (b *builder) addNode(n Node) error {
	if b.nodes[n.ID] {
		return fmt.Errorf("graph: duplicate node %s", n.ID)
	}
	b.nodes[n.ID] = true
	b.graph.Nodes = append(b.graph.Nodes, n)
	return nil
}
func (b *builder) addEdge(e Edge) error {
	if b.edges[e.ID] {
		return fmt.Errorf("graph: duplicate edge %s", e.ID)
	}
	b.edges[e.ID] = true
	b.graph.Edges = append(b.graph.Edges, e)
	return nil
}
func (b *builder) edge(relation, from, to, ptr string, entry corpus.ReportEntry, conditional bool) error {
	return b.addEdge(Edge{ID: edgeID(relation, from, to, ptr), From: from, To: to, Relation: relation, ReportPointer: ptr, DocumentID: entry.ID, AnalysisRevision: entry.AnalysisRevision, TargetDigest: entry.TargetDigest, Conditional: conditional})
}

func evaluatedAnalysis(e *corpus.Evaluation) *analysis.Result {
	if e == nil {
		return nil
	}
	switch e.Kind {
	case "analysis":
		return e.Analysis
	case "field_list":
		if e.FieldValidation != nil {
			return e.FieldValidation.Analysis
		}
	case "json_schema", "ocsf":
		if e.SchemaValidation != nil {
			return e.SchemaValidation.Analysis
		}
	}
	return nil
}
func analysisPointer(i int, kind string) string {
	base := pointer("entries", i, "evaluation")
	switch kind {
	case "analysis":
		return base + "/analysis"
	case "field_list":
		return base + "/field_validation/analysis"
	default:
		return base + "/schema_validation/analysis"
	}
}

func (b *builder) entry(i int, entry corpus.ReportEntry) error {
	base := pointer("entries", i)
	docID := documentID(entry.ID)
	var failure *corpus.AcquisitionError
	if entry.Failure != nil {
		copy := *entry.Failure
		failure = &copy
	}
	if err := b.addNode(Node{ID: docID, Kind: "document", ReportPointer: base, DocumentID: entry.ID, AnalysisRevision: entry.AnalysisRevision, TargetDigest: entry.TargetDigest, SourceHash: entry.SourceHash, Failure: failure}); err != nil {
		return err
	}
	if entry.Failure != nil {
		if entry.Evaluation != nil || entry.AnalysisRevision != "" {
			return fmt.Errorf("graph: failure entry %q also has evaluation", entry.ID)
		}
		return nil
	}
	r := evaluatedAnalysis(entry.Evaluation)
	if r == nil {
		return fmt.Errorf("graph: entry %q has no canonical analysis", entry.ID)
	}
	if entry.SourceHash != corpus.SourceHash(r.Document.Text) {
		return fmt.Errorf("graph: entry %q source hash mismatch", entry.ID)
	}
	revision, err := corpus.AnalysisRevision(r.Document)
	if err != nil || revision != entry.AnalysisRevision {
		return fmt.Errorf("graph: entry %q analysis revision mismatch: %v", entry.ID, err)
	}
	if entry.Evaluation.Kind != b.source.Mode {
		return fmt.Errorf("graph: entry %q evaluation mode mismatch", entry.ID)
	}
	if entry.Evaluation.Kind == "analysis" && entry.TargetDigest != "" {
		return fmt.Errorf("graph: analysis entry %q has target digest", entry.ID)
	}
	if entry.Evaluation.Kind != "analysis" && entry.TargetDigest == "" {
		return fmt.Errorf("graph: refined entry %q lacks target digest", entry.ID)
	}
	context := func(kind, canonical string) string {
		return evidenceID(kind, entry.ID, revision, entry.TargetDigest, canonical)
	}
	ap := analysisPointer(i, entry.Evaluation.Kind)
	stages := map[string]analysis.Stage{}
	scopes := map[string]analysis.Scope{}
	refs := map[string]analysis.Reference{}
	for _, scope := range r.Scopes {
		if scope.ID == "" || scopes[scope.ID].ID != "" {
			return fmt.Errorf("graph: duplicate or empty scope id %q", scope.ID)
		}
		scopes[scope.ID] = scope
	}
	for _, stage := range r.Stages {
		if stage.ID == "" || stages[stage.ID].ID != "" {
			return fmt.Errorf("graph: duplicate or empty stage id %q", stage.ID)
		}
		stages[stage.ID] = stage
	}
	for _, ref := range r.References {
		if ref.ID == "" || refs[ref.ID].ID != "" {
			return fmt.Errorf("graph: duplicate or empty reference id %q", ref.ID)
		}
		refs[ref.ID] = ref
	}
	for j, scope := range r.Scopes {
		if scope.ParentID != "" && scopes[scope.ParentID].ID == "" {
			return fmt.Errorf("graph: dangling parent scope %q", scope.ParentID)
		}
		if scope.StageID != "" && stages[scope.StageID].ID == "" {
			return fmt.Errorf("graph: dangling scope stage %q", scope.StageID)
		}
		ptr := ap + pointer("scopes", j)
		id := context("scope", scope.ID)
		loc := scope.Location
		if err := b.addNode(Node{ID: id, Kind: "scope", ReportPointer: ptr, DocumentID: entry.ID, AnalysisRevision: revision, TargetDigest: entry.TargetDigest, CanonicalID: scope.ID, ScopeKind: scope.Kind, StageID: scope.StageID, Location: &loc}); err != nil {
			return err
		}
		if err := b.edge("contains_scope", docID, id, ptr, entry, false); err != nil {
			return err
		}
		if scope.ParentID != "" {
			if err := b.edge("parent_scope", context("scope", scope.ParentID), id, ptr, entry, false); err != nil {
				return err
			}
		}
	}
	for j, stage := range r.Stages {
		if scopes[stage.ScopeID].ID == "" {
			return fmt.Errorf("graph: dangling stage scope %q", stage.ScopeID)
		}
		ptr := ap + pointer("stages", j)
		id := context("stage", stage.ID)
		loc, pos, complete := stage.Location, stage.Position, stage.SemanticComplete
		if err := b.addNode(Node{ID: id, Kind: "stage", ReportPointer: ptr, DocumentID: entry.ID, AnalysisRevision: revision, TargetDigest: entry.TargetDigest, CanonicalID: stage.ID, ScopeID: stage.ScopeID, Command: stage.Command, Position: &pos, SemanticComplete: &complete, Location: &loc}); err != nil {
			return err
		}
		if err := b.edge("owns_stage", context("scope", stage.ScopeID), id, ptr, entry, false); err != nil {
			return err
		}
	}
	for j, ref := range r.References {
		if stages[ref.StageID].ID == "" || scopes[ref.ScopeID].ID == "" {
			return fmt.Errorf("graph: dangling stage/scope on reference %q", ref.ID)
		}
		if err := checkedSourceSlice(r.Document.Text, ref.Location, ref.OriginalName); err != nil {
			return fmt.Errorf("graph: reference %q: %w", ref.ID, err)
		}
		for _, origin := range ref.OriginReferenceIDs {
			if refs[origin].ID == "" {
				return fmt.Errorf("graph: dangling reference origin %q", origin)
			}
		}
		ptr := ap + pointer("references", j)
		id := context("reference", ref.ID)
		loc := ref.Location
		if err := b.addNode(Node{ID: id, Kind: "reference", ReportPointer: ptr, DocumentID: entry.ID, AnalysisRevision: revision, TargetDigest: entry.TargetDigest, CanonicalID: ref.ID, StageID: ref.StageID, ScopeID: ref.ScopeID, DependencyKind: ref.Kind, Name: ref.NormalizedName, OriginalName: ref.OriginalName, Role: ref.Role, Resolution: ref.Resolution, Binding: ref.Binding, Location: &loc}); err != nil {
			return err
		}
		if err := b.edge("owns_reference", context("stage", ref.StageID), id, ptr, entry, false); err != nil {
			return err
		}
		for _, origin := range ref.OriginReferenceIDs {
			if err := b.edge("reference_origin", context("reference", origin), id, ptr, entry, false); err != nil {
				return err
			}
		}
		if dependencyKinds[ref.Kind] && ref.Resolution == "exact" {
			depID := dependencyID(ref.Kind, ref.NormalizedName)
			if summary, ok := b.dependencies[depID]; ok && hasString(summary.DocumentIDs, entry.ID) && hasString(summary.ReferenceIDs, ref.ID) {
				if b.observedDependencies[depID] == nil {
					b.observedDependencies[depID] = &dependencyObservation{documents: map[string]bool{}, references: map[string]int{}}
				}
				b.observedDependencies[depID].documents[entry.ID] = true
				b.observedDependencies[depID].references[ref.ID]++
				b.dependencyLinks = append(b.dependencyLinks, Edge{ID: edgeID("mentions_dependency", id, depID, ptr), From: id, To: depID, Relation: "mentions_dependency", ReportPointer: ptr, DocumentID: entry.ID, AnalysisRevision: revision, TargetDigest: entry.TargetDigest})
			} else {
				return fmt.Errorf("graph: missing dependency summary for reference %q", ref.ID)
			}
		}
	}
	for j, lineage := range r.Lineage {
		if stages[lineage.StageID].ID == "" || scopes[lineage.ScopeID].ID == "" {
			return fmt.Errorf("graph: dangling lineage owner at %d", j)
		}
		if err := checkedState(lineage.Before, refs); err != nil {
			return fmt.Errorf("graph: lineage %d before: %w", j, err)
		}
		if err := checkedState(lineage.After, refs); err != nil {
			return fmt.Errorf("graph: lineage %d after: %w", j, err)
		}
		ptr := ap + pointer("lineage", j)
		phaseKey := graphID(lineage.StageID, lineage.ScopeID, strconv.Itoa(j), lineage.Phase, orderString(lineage.ExecutionOrder))
		phaseID := context("phase", phaseKey)
		occ := j
		before, after := copyState(lineage.Before), copyState(lineage.After)
		if err := b.addNode(Node{ID: phaseID, Kind: "phase", ReportPointer: ptr, DocumentID: entry.ID, AnalysisRevision: revision, TargetDigest: entry.TargetDigest, StageID: lineage.StageID, ScopeID: lineage.ScopeID, Phase: lineage.Phase, ExecutionOrder: copyInt(lineage.ExecutionOrder), Occurrence: &occ, Before: &before, After: &after}); err != nil {
			return err
		}
		if err := b.edge("executes_phase", context("stage", lineage.StageID), phaseID, ptr, entry, false); err != nil {
			return err
		}
		for k, tr := range lineage.Transitions {
			for _, input := range tr.InputReferenceIDs {
				if refs[input].ID == "" {
					return fmt.Errorf("graph: dangling transition input %q", input)
				}
			}
			if tr.OutputReferenceID != "" && refs[tr.OutputReferenceID].ID == "" {
				return fmt.Errorf("graph: dangling transition output %q", tr.OutputReferenceID)
			}
			trPtr := ptr + pointer("transitions", k)
			trID := context("transition", graphID(phaseKey, strconv.Itoa(k)))
			trOcc := k
			if err := b.addNode(Node{ID: trID, Kind: "transition", ReportPointer: trPtr, DocumentID: entry.ID, AnalysisRevision: revision, TargetDigest: entry.TargetDigest, StageID: lineage.StageID, ScopeID: lineage.ScopeID, Phase: lineage.Phase, ExecutionOrder: copyInt(lineage.ExecutionOrder), Occurrence: &trOcc, Operation: tr.Operation, Name: tr.Output, Conditional: tr.Conditional}); err != nil {
				return err
			}
			if err := b.edge("records_transition", phaseID, trID, trPtr, entry, tr.Conditional); err != nil {
				return err
			}
			for _, input := range tr.InputReferenceIDs {
				if err := b.edge("transition_input", context("reference", input), trID, trPtr, entry, tr.Conditional); err != nil {
					return err
				}
			}
			if tr.OutputReferenceID != "" {
				if err := b.edge("transition_output", trID, context("reference", tr.OutputReferenceID), trPtr, entry, tr.Conditional); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func orderString(n *int) string {
	if n == nil {
		return "none"
	}
	return strconv.Itoa(*n)
}
func copyInt(n *int) *int {
	if n == nil {
		return nil
	}
	value := *n
	return &value
}
func copyState(s analysis.FieldState) analysis.FieldState {
	s.Fields = append([]analysis.FieldBinding{}, s.Fields...)
	for i := range s.Fields {
		s.Fields[i].OriginReferenceIDs = append([]string{}, s.Fields[i].OriginReferenceIDs...)
	}
	s.Removed = append([]string{}, s.Removed...)
	return s
}
func hasString(items []string, want string) bool {
	for _, item := range items {
		if item == want {
			return true
		}
	}
	return false
}
func checkedState(s analysis.FieldState, refs map[string]analysis.Reference) error {
	for _, field := range s.Fields {
		for _, id := range field.OriginReferenceIDs {
			if refs[id].ID == "" {
				return fmt.Errorf("dangling field origin %q", id)
			}
		}
	}
	return nil
}
func checkedSourceSlice(text string, loc analysis.Location, name string) error {
	start, end := loc.Start.Offset, loc.End.Offset
	if start < 0 || end < start || end > len(text) || !utf8.ValidString(text[:start]) || !utf8.ValidString(text[:end]) {
		return fmt.Errorf("invalid source span %d:%d", start, end)
	}
	if text[start:end] != name {
		return fmt.Errorf("source slice %q differs from %q", text[start:end], name)
	}
	return nil
}
