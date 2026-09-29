package closure

import (
	"fmt"
	"sort"
)

// DependencyGraph projects the traversal without inferring additional dependencies.
type DependencyGraph struct {
	Nodes []GraphNode `json:"nodes"`
	Edges []GraphEdge `json:"edges"`
}

type GraphNode struct {
	ID              string `json:"id"`
	Kind            string `json:"kind"`
	ObjectID        string `json:"object_id,omitempty"`
	InstanceID      string `json:"instance_id,omitempty"`
	EvidencePointer string `json:"evidence_pointer"`
}

// A graph edge with Unknown set has no target node. Its resolution explains why.
type GraphEdge struct {
	ID                string         `json:"id"`
	From              string         `json:"from"`
	To                string         `json:"to,omitempty"`
	Unknown           bool           `json:"unknown"`
	Kind              string         `json:"kind"`
	Name              string         `json:"name"`
	Resolution        string         `json:"resolution"`
	TraversalPointer  string         `json:"traversal_pointer"`
	Source            SourceInterval `json:"source"`
	Property          string         `json:"property,omitempty"`
	Path              []string       `json:"path"`
	CyclePath         []string       `json:"cycle_path"`
	InvocationNodeIDs []string       `json:"invocation_node_ids"`
}

type BOMEntry struct {
	ObjectID    string          `json:"object_id"`
	Kind        string          `json:"kind"`
	Name        string          `json:"name"`
	Direct      bool            `json:"direct"`
	Transitive  bool            `json:"transitive"`
	Incomplete  bool            `json:"incomplete"`
	Occurrences []BOMOccurrence `json:"occurrences"`
}

type BOMOccurrence struct {
	EdgeID      string   `json:"edge_id"`
	Path        []string `json:"path"`
	InstanceIDs []string `json:"instance_ids"`
}

func objectNodeID(id string) string   { return "object:" + id }
func instanceNodeID(id string) string { return "instance:" + id }

// projectReport reads only canonical traversal, provenance, and gaps from Evaluate.
func projectReport(report *Report) {
	graph := DependencyGraph{Nodes: []GraphNode{{ID: "root", Kind: "query", EvidencePointer: "/query"}}, Edges: []GraphEdge{}}
	nodes := map[string]GraphNode{"root": graph.Nodes[0]}
	entries := map[string]*BOMEntry{}
	addObject := func(id, pointer string) {
		nodeID := objectNodeID(id)
		if _, exists := nodes[nodeID]; !exists {
			nodes[nodeID] = GraphNode{ID: nodeID, Kind: "object", ObjectID: id, EvidencePointer: pointer}
		}
	}
	addFrame := func(frame InvocationFrame, pointer string) {
		addObject(frame.ObjectID, pointer)
		nodeID := instanceNodeID(frame.InstanceID)
		if _, exists := nodes[nodeID]; !exists {
			nodes[nodeID] = GraphNode{ID: nodeID, Kind: "expansion", ObjectID: frame.ObjectID, InstanceID: frame.InstanceID, EvidencePointer: pointer}
		}
	}
	for i, segment := range report.Provenance {
		for j, frame := range segment.InvocationChain {
			addFrame(frame, fmt.Sprintf("/provenance/%d/invocation_chain/%d", i, j))
		}
	}
	for i, edge := range report.Traversal {
		pointer := fmt.Sprintf("/traversal/%d", i)
		from := "root"
		if edge.FromObjectID != "" {
			addObject(edge.FromObjectID, pointer)
			from = objectNodeID(edge.FromObjectID)
		}
		to := ""
		if edge.ToObjectID != "" {
			addObject(edge.ToObjectID, pointer)
			to = objectNodeID(edge.ToObjectID)
		}
		graphEdge := GraphEdge{ID: edge.ID, From: from, To: to, Unknown: to == "", Kind: edge.Kind, Name: edge.Name, Resolution: edge.Resolution, TraversalPointer: pointer, Source: edge.Source, Property: edge.Property, Path: append([]string{}, edge.Path...), CyclePath: append([]string{}, edge.CyclePath...), InvocationNodeIDs: []string{}}
		for j, frame := range edge.InvocationChain {
			addFrame(frame, fmt.Sprintf("%s/invocation_chain/%d", pointer, j))
			graphEdge.InvocationNodeIDs = append(graphEdge.InvocationNodeIDs, instanceNodeID(frame.InstanceID))
		}
		graph.Edges = append(graph.Edges, graphEdge)
		if to == "" {
			continue
		}
		entry := entries[edge.ToObjectID]
		if entry == nil {
			entry = &BOMEntry{ObjectID: edge.ToObjectID, Kind: edge.Kind, Name: edge.Name, Occurrences: []BOMOccurrence{}}
			entries[edge.ToObjectID] = entry
		}
		direct := edge.FromObjectID == "" && len(edge.Path) == 0
		entry.Direct = entry.Direct || direct
		entry.Transitive = entry.Transitive || !direct
		if edge.Resolution == "cycle" {
			entry.Incomplete = true
		}
		path := append([]string{}, edge.Path...)
		path = append(path, edge.ToObjectID)
		occurrence := BOMOccurrence{EdgeID: edge.ID, Path: path, InstanceIDs: []string{}}
		for _, frame := range edge.InvocationChain {
			occurrence.InstanceIDs = append(occurrence.InstanceIDs, frame.InstanceID)
		}
		entry.Occurrences = append(entry.Occurrences, occurrence)
	}
	for _, gap := range report.Gaps {
		for id, entry := range entries {
			if gap.Kind == entry.Kind && gap.Name == entry.Name {
				entry.Incomplete = true
			}
			for _, pathID := range gap.Path {
				if pathID == id {
					entry.Incomplete = true
				}
			}
		}
	}
	otherNodes := make([]GraphNode, 0, len(nodes)-1)
	for id, node := range nodes {
		if id != "root" {
			otherNodes = append(otherNodes, node)
		}
	}
	sort.Slice(otherNodes, func(i, j int) bool { return otherNodes[i].ID < otherNodes[j].ID })
	graph.Nodes = append(graph.Nodes, otherNodes...)
	report.Graph = graph
	report.BOM = make([]BOMEntry, 0, len(entries))
	for _, entry := range entries {
		report.BOM = append(report.BOM, *entry)
	}
	sort.Slice(report.BOM, func(i, j int) bool {
		a, b := report.BOM[i], report.BOM[j]
		if a.Kind != b.Kind {
			return a.Kind < b.Kind
		}
		if a.Name != b.Name {
			return a.Name < b.Name
		}
		return a.ObjectID < b.ObjectID
	})
}
