package analysis

// CorrelationGraph contains only query-proved edges. Mere co-occurrence never
// supplies an edge; later semantic join handlers may add exact equality proof.
type CorrelationGraph struct {
	Outcome    string            `json:"outcome"`
	Coverage   InputCoverage     `json:"coverage"`
	Nodes      []CorrelationNode `json:"nodes"`
	Edges      []CorrelationEdge `json:"edges"`
	Components [][]string        `json:"components"`
}
type CorrelationNode struct {
	InputID      string `json:"input_id"`
	OccurrenceID string `json:"occurrence_id"`
}
type CorrelationEndpoint struct {
	InputID       string        `json:"input_id"`
	OccurrenceID  string        `json:"occurrence_id"`
	FieldIdentity FieldIdentity `json:"field_identity"`
	ReferenceIDs  []string      `json:"reference_ids"`
}
type CorrelationEdge struct {
	ID       string              `json:"id"`
	Left     CorrelationEndpoint `json:"left"`
	Right    CorrelationEndpoint `json:"right"`
	StageID  string              `json:"stage_id"`
	ScopeID  string              `json:"scope_id"`
	Location Location            `json:"location"`
}

func initialCorrelation(inputs []QueryInput, coverage InputCoverage) CorrelationGraph {
	graph := CorrelationGraph{Outcome: "not_applicable", Coverage: InputCoverage{State: "not_applicable", Reasons: []InputReason{}}, Nodes: []CorrelationNode{}, Edges: []CorrelationEdge{}, Components: [][]string{}}
	for _, input := range inputs {
		for _, occurrence := range input.Occurrences {
			graph.Nodes = append(graph.Nodes, CorrelationNode{InputID: input.ID, OccurrenceID: occurrence.ID})
			graph.Components = append(graph.Components, []string{occurrence.ID})
		}
	}
	if len(graph.Nodes) > 1 || coverage.State == "partial" {
		graph.Outcome = "indeterminate"
		graph.Coverage = InputCoverage{State: "partial", Reasons: []InputReason{}}
		if len(graph.Nodes) > 1 {
			o := inputs[0].Occurrences[0]
			graph.Coverage.Reasons = append(graph.Coverage.Reasons, InputReason{Code: "correlation_incomplete", Message: "no complete source correlation proof is available", Location: o.Location, StageID: o.StageID, ScopeID: o.ScopeID, ReferenceIDs: []string{}})
		}
		for _, reason := range coverage.Reasons {
			graph.Coverage.Reasons = append(graph.Coverage.Reasons, reason)
		}
	}
	return graph
}
func cloneCorrelation(in CorrelationGraph) CorrelationGraph {
	out := in
	out.Coverage = cloneInputCoverage(in.Coverage)
	out.Nodes = append([]CorrelationNode{}, in.Nodes...)
	out.Edges = append([]CorrelationEdge{}, in.Edges...)
	out.Components = make([][]string, len(in.Components))
	for i := range out.Components {
		out.Components[i] = copyIDs(in.Components[i])
	}
	for i := range out.Edges {
		out.Edges[i].Left.FieldIdentity.Segments = copyIDs(in.Edges[i].Left.FieldIdentity.Segments)
		out.Edges[i].Right.FieldIdentity.Segments = copyIDs(in.Edges[i].Right.FieldIdentity.Segments)
		out.Edges[i].Left.ReferenceIDs = copyIDs(in.Edges[i].Left.ReferenceIDs)
		out.Edges[i].Right.ReferenceIDs = copyIDs(in.Edges[i].Right.ReferenceIDs)
	}
	return out
}
