package workflow

import (
	"bytes"
	"encoding/json"
	"github.com/delgado-jacob/spl-toolkit/pkg/closure"
	"strconv"
	"strings"
	"testing"
)

func TestWorkflowGraphEndpointsExist(t *testing.T) {
	r, err := Assess(seedRequest(t, "from [{id:1}] | eval label=id | fields label"))
	if err != nil {
		t.Fatal(err)
	}
	g, err := ExportGraph(r)
	if err != nil {
		t.Fatal(err)
	}
	for _, subject := range g.Subjects {
		if subject.Analysis == nil {
			continue
		}
		nodes := map[string]bool{}
		for _, n := range subject.Analysis.Nodes {
			nodes[n.ID] = true
		}
		for _, e := range subject.Analysis.Edges {
			if !nodes[e.From] || !nodes[e.To] {
				t.Fatalf("dangling edge: %+v", e)
			}
		}
	}
}

func TestWorkflowGraphEvidenceVariantsAndIsolation(t *testing.T) {
	r, err := Assess(seedResolutionRequest(t))
	if err != nil {
		t.Fatal(err)
	}
	before, _ := json.Marshal(r)
	g, err := ExportGraph(r)
	if err != nil {
		t.Fatal(err)
	}
	if len(g.Subjects) != 3 || g.Subjects[0].Domain != "original" || g.Subjects[1].Domain != "candidate" || g.Subjects[2].Analysis == nil {
		t.Fatalf("unsuccessful candidate omitted: %+v", g.Subjects)
	}
	seen := map[string]bool{}
	for _, s := range g.Subjects {
		workflowPointer(t, r, s.EvidencePointer)
		if s.Analysis != nil {
			for _, n := range s.Analysis.Nodes {
				workflowPointer(t, r, n.ReportPointer)
				if seen[n.ID] {
					t.Fatalf("duplicate qualified ID %q", n.ID)
				}
				seen[n.ID] = true
			}
			for _, e := range s.Analysis.Edges {
				workflowPointer(t, r, e.ReportPointer)
			}
		}
		if s.Closure != nil {
			assertClosureGraph(t, r, s.Closure)
		}
	}
	again, err := ExportGraph(r)
	if err != nil {
		t.Fatal(err)
	}
	one, _ := json.Marshal(g)
	two, _ := json.Marshal(again)
	if !bytes.Equal(one, two) {
		t.Fatal("nondeterministic graph")
	}
	g.Subjects[1].VariantSelection[0].Value = "mutated"
	for _, s := range g.Subjects {
		if s.Analysis != nil {
			for i := range s.Analysis.Nodes {
				n := &s.Analysis.Nodes[i]
				if n.Before != nil && len(n.Before.Fields) > 0 {
					n.Before.Fields[0].Name = "mutated"
				}
				if n.Location != nil {
					n.Location.Start.Offset = -99
				}
			}
		}
	}
	after, _ := json.Marshal(r)
	if !bytes.Equal(before, after) {
		t.Fatal("graph aliases report")
	}
	r.Entries[0].Analysis.Requirements.CapabilityRevision = "historical"
	if _, err := ExportGraph(r); err == nil {
		t.Fatal("historical capability relabeled as current")
	}
}

func workflowPointer(t *testing.T, r *Report, ptr string) any {
	t.Helper()
	raw, _ := json.Marshal(r)
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		t.Fatal(err)
	}
	for _, part := range strings.Split(strings.TrimPrefix(ptr, "/"), "/") {
		part = strings.ReplaceAll(strings.ReplaceAll(part, "~1", "/"), "~0", "~")
		switch x := v.(type) {
		case map[string]any:
			var ok bool
			v, ok = x[part]
			if !ok {
				t.Fatalf("missing pointer %s", ptr)
			}
		case []any:
			i, err := strconv.Atoi(part)
			if err != nil || i < 0 || i >= len(x) {
				t.Fatalf("invalid pointer %s", ptr)
			}
			v = x[i]
		default:
			t.Fatalf("non-container pointer %s", ptr)
		}
	}
	return v
}
func assertClosureGraph(t *testing.T, r *Report, g *closure.DependencyGraph) {
	t.Helper()
	nodes := map[string]bool{}
	for _, n := range g.Nodes {
		nodes[n.ID] = true
		workflowPointer(t, r, n.EvidencePointer)
	}
	for _, e := range g.Edges {
		workflowPointer(t, r, e.TraversalPointer)
		if !nodes[e.From] || e.To != "" && !nodes[e.To] {
			t.Fatalf("dangling closure edge: %+v", e)
		}
		if e.Unknown && e.To != "" {
			t.Fatal("unknown target invented")
		}
		for _, id := range e.InvocationNodeIDs {
			if !nodes[id] {
				t.Fatalf("dangling invocation %s", id)
			}
		}
	}
}
