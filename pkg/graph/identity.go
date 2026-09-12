package graph

import "strconv"

// graphID is a collision-free encoding of a string tuple: each UTF-8 byte
// length precedes its value, so separators inside names cannot alias tuples.
// Revision and target digest are explicit tuple members for local evidence.
func graphID(parts ...string) string {
	out := "graph-v1:"
	for _, part := range parts {
		out += strconv.Itoa(len(part)) + ":" + part
	}
	return out
}

func documentID(id string) string { return graphID("document", id) }
func evidenceID(kind, document, revision, target, canonical string) string {
	return graphID(kind, document, revision, target, canonical)
}
func dependencyID(kind, name string) string { return graphID("dependency", kind, name) }
func edgeID(relation, from, to, pointer string) string {
	return graphID("edge", relation, from, to, pointer)
}
