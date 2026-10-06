package analysis

import (
	"slices"
	"sort"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	"github.com/delgado-jacob/spl-toolkit/parser/spl2"
)

// CorrelationGraph contains only query-proved edges. Co-occurrence and matching
// names never supply an edge.
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
	Location      Location      `json:"location"`
}
type CorrelationKeyEvidence struct {
	Predicate string              `json:"predicate"`
	Left      CorrelationEndpoint `json:"left"`
	Right     CorrelationEndpoint `json:"right"`
	Location  Location            `json:"location"`
}
type CorrelationEdge struct {
	ID       string                   `json:"id"`
	Left     CorrelationEndpoint      `json:"left"`
	Right    CorrelationEndpoint      `json:"right"`
	StageID  string                   `json:"stage_id"`
	ScopeID  string                   `json:"scope_id"`
	Location Location                 `json:"location"`
	Keys     []CorrelationKeyEvidence `json:"keys"`
}

// An event retains source facts before output environments merge. Logical field
// ownership alone is insufficient: each side must have one supplying row source.
type correlationEvent struct {
	left, right                                              sourceOwner
	leftReference, rightReference                            string
	leftIdentity, rightIdentity                              fieldIdentity
	leftLocation, rightLocation, location, predicateLocation Location
	stageID, scopeID                                         string
	proved                                                   bool
}

func cloneCorrelationEvents(in []correlationEvent) []correlationEvent {
	out := append([]correlationEvent{}, in...)
	for i := range out {
		out[i].left = cloneSourceOwners([]sourceOwner{in[i].left})[0]
		out[i].right = cloneSourceOwners([]sourceOwner{in[i].right})[0]
		out[i].leftIdentity = cloneRequirementTraceIdentity(in[i].leftIdentity)
		out[i].rightIdentity = cloneRequirementTraceIdentity(in[i].rightIdentity)
	}
	return out
}

// Reference pairs identify equality occurrences, independently of projected
// view use sites. Collision buckets retain any distinct situated owner facts.
type correlationEventIdentity struct {
	leftReference, rightReference, stageID, scopeID string
	predicateLocation                               Location
}

func correlationIdentity(event *correlationEvent) correlationEventIdentity {
	return correlationEventIdentity{event.leftReference, event.rightReference, event.stageID, event.scopeID, event.predicateLocation}
}
func mergeCorrelationEvents(base, additions []correlationEvent) []correlationEvent {
	out := cloneCorrelationEvents(base)
	indexes := make(map[correlationEventIdentity][]int, len(base)+len(additions))
	for i := range out {
		key := correlationIdentity(&out[i])
		indexes[key] = append(indexes[key], i)
	}
	for i := range additions {
		event := &additions[i]
		key := correlationIdentity(event)
		found := false
		for _, index := range indexes[key] {
			// Only candidates for this equality occurrence are compared. Explicit
			// comparisons avoid boxing and preserve clone-normalized empty use lists.
			if sameCorrelationEvent(&out[index], event) {
				found = true
				break
			}
		}
		if !found {
			indexes[key] = append(indexes[key], len(out))
			out = append(out, cloneCorrelationEvents(additions[i : i+1])[0])
		}
	}
	return out
}

func sameCorrelationEvent(a, b *correlationEvent) bool {
	return correlationIdentity(a) == correlationIdentity(b) && a.proved == b.proved &&
		a.leftLocation == b.leftLocation && a.rightLocation == b.rightLocation && a.location == b.location &&
		sameCorrelationFieldIdentity(a.leftIdentity, b.leftIdentity) && sameCorrelationFieldIdentity(a.rightIdentity, b.rightIdentity) &&
		sameCorrelationOwner(a.left, b.left) && sameCorrelationOwner(a.right, b.right)
}
func sameCorrelationFieldIdentity(a, b fieldIdentity) bool {
	return a.Kind == b.Kind && a.Qualifier == b.Qualifier && a.PublicName == b.PublicName && a.exact == b.exact &&
		(a.Segments == nil) == (b.Segments == nil) && slices.Equal(a.Segments, b.Segments)
}
func sameCorrelationOwner(a, b sourceOwner) bool {
	if a.unresolved != b.unresolved || !sameCorrelationFieldIdentity(a.identity, b.identity) ||
		a.input.kind != b.input.kind || a.input.name != b.input.name || a.input.sourceID != b.input.sourceID || a.input.identity != b.input.identity {
		return false
	}
	left, right := a.input.occurrence, b.input.occurrence
	// cloneInputFacts makes empty use arrays non-nil, so nil and empty arrays
	// represent the same inherited fact at this merge boundary.
	return left.ID == right.ID && left.ReferenceID == right.ReferenceID && left.OriginalReferenceID == right.OriginalReferenceID &&
		left.StageID == right.StageID && left.ScopeID == right.ScopeID && left.Alias == right.Alias && left.Location == right.Location &&
		slices.Equal(left.UseSiteLocations, right.UseSiteLocations) && slices.Equal(left.UseSiteReferenceIDs, right.UseSiteReferenceIDs)
}
func supplyingOccurrence(owners []sourceOwner) (sourceOwner, bool) {
	if len(owners) == 0 {
		return sourceOwner{}, false
	}
	first := owners[0]
	for _, owner := range owners {
		if owner.unresolved || owner.input.kind == "unresolved_source" || inputOccurrenceID(owner.input) != inputOccurrenceID(first.input) {
			return sourceOwner{}, false
		}
	}
	return cloneSourceOwners([]sourceOwner{first})[0], true
}
func (s *spl2SemanticStage) recordJoinCorrelations(command *spl2.JoinCommandContext, ids []string, rightEnvironment *environment) {
	s.recordSourceJoinCorrelations(command, command.SqlJoinPredicate(), ids, s.selectedJoin(command).rightAlias, rightEnvironment)
}

func (s *spl2SemanticStage) recordSourceJoinCorrelations(owner antlr.ParserRuleContext, predicate spl2.ISqlJoinPredicateContext, ids []string, rightAlias string, rightEnvironment *environment) {
	trace := s.env.requirements.trace
	if trace == nil {
		return
	}
	for i, equality := range predicate.AllSqlJoinEquality() {
		left, right := trace.reference(ids[2*i]), trace.reference(ids[2*i+1])
		if left.reference.FieldIdentity.Qualifier == rightAlias {
			left, right = right, left
		}
		// A generated-only join has no external occurrence obligation. Constant
		// keys on external rows still require supplier uncertainty: no field
		// lineage proves which external occurrence supplies those rows.
		if len(s.env.inputs) == 0 && len(rightEnvironment.inputs) == 0 {
			continue
		}
		lo, lok := supplyingOccurrence(left.owners)
		ro, rok := supplyingOccurrence(right.owners)
		stage := s.result.Stages[s.stage]
		trace.correlations = append(trace.correlations, correlationEvent{
			left: lo, right: ro, leftReference: left.pendingID, rightReference: right.pendingID,
			leftIdentity: left.fieldIdentity.clone(), rightIdentity: right.fieldIdentity.clone(),
			leftLocation: left.reference.Location, rightLocation: right.reference.Location,
			location: s.parsed2.source.contextLocation(owner), predicateLocation: s.parsed2.source.contextLocation(equality),
			stageID: stage.ID, scopeID: stage.ScopeID, proved: lok && rok && left.reference.Binding != "unavailable" && right.reference.Binding != "unavailable",
		})
	}
}
func (t *requirementTrace) remapCorrelationReferences(mapping map[string]string) {
	for i := range t.correlations {
		e := &t.correlations[i]
		e.leftReference = mapping[e.leftReference]
		e.rightReference = mapping[e.rightReference]
		owners := []sourceOwner{e.left, e.right}
		remapSourceOwnerReferences(owners, mapping)
		e.left, e.right = owners[0], owners[1]
	}
}
func (t *requirementTrace) remapCorrelationStages(mapping map[string]string) {
	for i := range t.correlations {
		e := &t.correlations[i]
		e.stageID = mapping[e.stageID]
		owners := []sourceOwner{e.left, e.right}
		remapSourceOwnerStages(owners, mapping)
		e.left, e.right = owners[0], owners[1]
	}
}

// Projecting a body event pairs the context appended after each original use
// prefix. Prefixes can differ for nested views, but a repeated terminal use must
// never connect to another use of the same body.
func projectedCorrelationOccurrences(owner sourceOwner, inputs []QueryInput) []InputOccurrence {
	ids := projectedOwnerOccurrenceIDs([]sourceOwner{owner}, inputs)
	found := map[string]bool{}
	for _, id := range ids {
		found[id] = true
	}
	out := []InputOccurrence{}
	for _, input := range inputs {
		for _, occurrence := range input.Occurrences {
			if found[occurrence.ID] {
				out = append(out, occurrence)
			}
		}
	}
	return out
}
func projectCorrelation(trace *requirementTrace, inputs []QueryInput, coverage InputCoverage) CorrelationGraph {
	graph := CorrelationGraph{Outcome: "not applicable", Coverage: InputCoverage{State: "not_applicable", Reasons: []InputReason{}}, Nodes: []CorrelationNode{}, Edges: []CorrelationEdge{}, Components: [][]string{}}
	nodeIndex := map[string]int{}
	for _, input := range inputs {
		for _, o := range input.Occurrences {
			nodeIndex[o.ID] = len(graph.Nodes)
			graph.Nodes = append(graph.Nodes, CorrelationNode{InputID: input.ID, OccurrenceID: o.ID})
		}
	}
	graph.Coverage.Reasons = append(graph.Coverage.Reasons, cloneInputCoverage(coverage).Reasons...)
	for _, entry := range trace.diagnostics {
		d := entry.diagnostic
		deferredSources := 0
		if d.Message == "SQL clause field scheduling is not yet modeled" {
			for _, fact := range trace.inputs {
				location := fact.occurrence.Location
				if uses := fact.occurrence.UseSiteLocations; len(uses) > 0 {
					location = uses[len(uses)-1]
				}
				if location.Start.Offset >= d.Location.Start.Offset && location.End.Offset <= d.Location.End.Offset {
					deferredSources++
				}
			}
		}
		if d.Code == CodeUnsupportedSemantics && (strings.Contains(strings.ToLower(d.Message), "join") || deferredSources > 1) {
			graph.Coverage.Reasons = append(graph.Coverage.Reasons, InputReason{Code: "correlation_incomplete", Message: d.Message, Location: d.Location, StageID: d.StageID, ScopeID: d.ScopeID, ReferenceIDs: copyIDs(entry.pendingReferenceIDs)})
		}
	}
	events := cloneCorrelationEvents(trace.correlations)
	sort.SliceStable(events, func(i, j int) bool {
		return events[i].predicateLocation.Start.Offset < events[j].predicateLocation.Start.Offset
	})
	edgeIndexes := map[string]int{}
	for _, event := range events {
		if !event.proved {
			graph.Coverage.Reasons = append(graph.Coverage.Reasons, InputReason{Code: "correlation_incomplete", Message: "join key supplying source occurrence has not been proved", Location: event.predicateLocation, StageID: event.stageID, ScopeID: event.scopeID, ReferenceIDs: []string{event.leftReference, event.rightReference}})
			continue
		}
		lefts := projectedCorrelationOccurrences(event.left, inputs)
		rights := projectedCorrelationOccurrences(event.right, inputs)
		rightContexts := map[string][]InputOccurrence{}
		for _, right := range rights {
			suffix := right.UseSiteLocations[len(event.right.input.occurrence.UseSiteLocations):]
			key := inputUsePrefixKey("", suffix)
			rightContexts[key] = append(rightContexts[key], right)
		}
		for _, left := range lefts {
			suffix := left.UseSiteLocations[len(event.left.input.occurrence.UseSiteLocations):]
			for _, right := range rightContexts[inputUsePrefixKey("", suffix)] {
				if left.ID == right.ID {
					continue
				}
				li, _ := event.leftIdentity.public()
				ri, _ := event.rightIdentity.public()
				key := CorrelationKeyEvidence{Predicate: "equals", Left: CorrelationEndpoint{InputID: opaqueInputID(event.left.input.kind, event.left.input.identity), OccurrenceID: left.ID, FieldIdentity: li, ReferenceIDs: []string{event.leftReference}, Location: event.leftLocation}, Right: CorrelationEndpoint{InputID: opaqueInputID(event.right.input.kind, event.right.input.identity), OccurrenceID: right.ID, FieldIdentity: ri, ReferenceIDs: []string{event.rightReference}, Location: event.rightLocation}, Location: event.predicateLocation}
				edgeKey := orderedStringSliceKey([]string{left.ID, right.ID, event.stageID})
				index, exists := edgeIndexes[edgeKey]
				if !exists {
					index = len(graph.Edges)
					edgeIndexes[edgeKey] = index
					graph.Edges = append(graph.Edges, CorrelationEdge{ID: opaqueEvidenceID("edge-", edgeKey), Left: key.Left, Right: key.Right, StageID: event.stageID, ScopeID: event.scopeID, Location: event.location, Keys: []CorrelationKeyEvidence{}})
				}
				graph.Edges[index].Keys = append(graph.Edges[index].Keys, key)
			}
		}
	}
	// Components include isolated participants, including repeated logical inputs.
	parents := make([]int, len(graph.Nodes))
	for i := range parents {
		parents[i] = i
	}
	var root func(int) int
	root = func(i int) int {
		for parents[i] != i {
			i = parents[i]
		}
		return i
	}
	for _, edge := range graph.Edges {
		l, r := root(nodeIndex[edge.Left.OccurrenceID]), root(nodeIndex[edge.Right.OccurrenceID])
		parents[r] = l
	}
	components := map[int]int{}
	for i, node := range graph.Nodes {
		r := root(i)
		index, ok := components[r]
		if !ok {
			index = len(graph.Components)
			components[r] = index
			graph.Components = append(graph.Components, []string{})
		}
		graph.Components[index] = append(graph.Components[index], node.OccurrenceID)
	}
	switch {
	case coverage.State == "partial" || len(graph.Coverage.Reasons) > 0:
		graph.Outcome = "indeterminate"
		graph.Coverage.State = "partial"
	case len(graph.Nodes) <= 1:
	case len(graph.Components) == 1:
		graph.Outcome = "connected"
		graph.Coverage.State = "complete"
	default:
		graph.Outcome = "disconnected"
		graph.Coverage.State = "complete"
	}
	return graph
}
func cloneCorrelationEndpoint(in CorrelationEndpoint) CorrelationEndpoint {
	out := in
	out.FieldIdentity.Segments = copyIDs(in.FieldIdentity.Segments)
	out.ReferenceIDs = copyIDs(in.ReferenceIDs)
	return out
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
		out.Edges[i].Left = cloneCorrelationEndpoint(in.Edges[i].Left)
		out.Edges[i].Right = cloneCorrelationEndpoint(in.Edges[i].Right)
		out.Edges[i].Keys = append([]CorrelationKeyEvidence{}, in.Edges[i].Keys...)
		for j := range out.Edges[i].Keys {
			out.Edges[i].Keys[j].Left = cloneCorrelationEndpoint(in.Edges[i].Keys[j].Left)
			out.Edges[i].Keys[j].Right = cloneCorrelationEndpoint(in.Edges[i].Keys[j].Right)
		}
	}
	return out
}
