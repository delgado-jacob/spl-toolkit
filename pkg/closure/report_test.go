package closure

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
)

func TestReportDeterministicProjections(t *testing.T) {
	users := evalDef("users", "lookup", "users", "")
	child := evalDef("child", "saved_search", "Child", "lookup users user OUTPUT role")
	a := evalDef("a", "saved_search", "A", "search index=main", Relation{Kind: "saved_search", Name: "Child", Property: stringPointer("/child")})
	b := evalDef("b", "saved_search", "B", "search index=main", Relation{Kind: "saved_search", Name: "Child", Property: stringPointer("/child")})
	collections := []Collection{{Kind: "saved_search", Coverage: "complete"}, {Kind: "lookup", Coverage: "complete"}}
	req := evalRequest("| from savedsearch:A | from savedsearch:B", collections, users, child, b, a)
	first, err := Evaluate(req)
	if err != nil {
		t.Fatal(err)
	}
	req.Bundle.Objects = []Definition{a, b, child, users}
	req.Bundle.Collections = []Collection{collections[1], collections[0]}
	second, err := Evaluate(req)
	if err != nil {
		t.Fatal(err)
	}
	x, _ := json.Marshal(first)
	y, _ := json.Marshal(second)
	if string(x) != string(y) {
		t.Fatal("equivalent bundle permutations changed report JSON")
	}
	entries := map[string]BOMEntry{}
	for _, entry := range first.BOM {
		entries[entry.ObjectID] = entry
	}
	if len(entries) != 4 || len(entries["users"].Occurrences) != 2 || !entries["users"].Transitive || entries["users"].Direct {
		t.Fatalf("BOM identity or occurrences: %+v", first.BOM)
	}
	paths := []string{}
	for _, occurrence := range entries["users"].Occurrences {
		paths = append(paths, strings.Join(occurrence.Path, "/"))
	}
	sort.Strings(paths)
	if !reflect.DeepEqual(paths, []string{"a/child/users", "b/child/users"}) {
		t.Fatalf("occurrence paths: %v", paths)
	}
	if !strings.Contains(FormatInventory(first), "transitive") || !strings.Contains(FormatInventory(first), "users") {
		t.Fatalf("inventory: %s", FormatInventory(first))
	}
}

func TestGraphEvidenceAndUnknownEdges(t *testing.T) {
	req := evalRequest("| lookup users user OUTPUT role | lookup missing user OUTPUT role", []Collection{{Kind: "lookup", Coverage: "complete"}}, evalDef("users", "lookup", "users", ""))
	report, err := Evaluate(req)
	if err != nil {
		t.Fatal(err)
	}
	nodes := map[string]bool{}
	for _, node := range report.Graph.Nodes {
		nodes[node.ID] = true
	}
	if !nodes["root"] || !nodes["object:users"] {
		t.Fatalf("graph nodes: %+v", report.Graph.Nodes)
	}
	for _, node := range report.Graph.Nodes {
		if !reportPointerExists(t, report, node.EvidencePointer) {
			t.Fatalf("invalid node evidence pointer: %+v", node)
		}
	}
	unknown := false
	for _, edge := range report.Graph.Edges {
		if !nodes[edge.From] || edge.To != "" && !nodes[edge.To] {
			t.Fatalf("dangling graph edge: %+v", edge)
		}
		if !reportPointerExists(t, report, edge.TraversalPointer) {
			t.Fatalf("invalid edge evidence pointer: %+v", edge)
		}
		var index int
		if _, err := fmt.Sscanf(edge.TraversalPointer, "/traversal/%d", &index); err != nil || index < 0 || index >= len(report.Traversal) || edge.ID != report.Traversal[index].ID {
			t.Fatalf("invalid evidence pointer: %+v", edge)
		}
		if edge.Unknown {
			unknown = true
			if edge.To != "" {
				t.Fatalf("unknown edge has target: %+v", edge)
			}
		}
		if edge.Source.Kind == "query" && (edge.Source.Start < 0 || edge.Source.End > len(req.Document.Text) || edge.Source.Start > edge.Source.End) {
			t.Fatalf("invalid source interval: %+v", edge.Source)
		}
	}
	if !unknown {
		t.Fatalf("missing explicit unknown edge: %+v", report.Graph.Edges)
	}
}

func TestBOMCycleAndReportDetachment(t *testing.T) {
	a := evalDef("a", "saved_search", "A", "search index=main", Relation{Kind: "saved_search", Name: "B", Property: stringPointer("/next")})
	b := evalDef("b", "saved_search", "B", "search index=main", Relation{Kind: "saved_search", Name: "A", Property: stringPointer("/next")})
	req := evalRequest("| from savedsearch:A", []Collection{{Kind: "saved_search", Coverage: "complete"}}, a, b)
	report, err := Evaluate(req)
	if err != nil {
		t.Fatal(err)
	}
	cycle := false
	for _, edge := range report.Graph.Edges {
		if len(edge.CyclePath) > 0 {
			cycle = true
			if edge.To != "object:a" {
				t.Fatalf("cycle target: %+v", edge)
			}
		}
	}
	if !cycle {
		t.Fatal("cycle edge omitted")
	}
	if !strings.Contains(FormatInventory(report), "cycle") {
		t.Fatalf("partial reasons missing: %s", FormatInventory(report))
	}
	req.Bundle.Objects[0].Name = "changed"
	req.Bundle.Collections[0].Coverage = "partial"
	if report.BOM[0].Name == "changed" || report.Status == analysis.Valid {
		t.Fatalf("report changed with request: %+v", report.BOM)
	}
}

func TestReportSyntheticCases(t *testing.T) {
	raw, err := os.ReadFile("../../testdata/closure/cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases struct {
		Cases []struct {
			Name       string          `json:"name"`
			Request    Request         `json:"request"`
			WantStatus analysis.Status `json:"want_status"`
			WantBOMIDs []string        `json:"want_bom_ids"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(raw, &cases); err != nil {
		t.Fatal(err)
	}
	if len(cases.Cases) == 0 {
		t.Fatal("empty fixture")
	}
	for _, tc := range cases.Cases {
		t.Run(tc.Name, func(t *testing.T) {
			report, err := Evaluate(tc.Request)
			if err != nil {
				t.Fatal(err)
			}
			ids := []string{}
			for _, entry := range report.BOM {
				ids = append(ids, entry.ObjectID)
			}
			sort.Strings(ids)
			if report.Status != tc.WantStatus || !reflect.DeepEqual(ids, tc.WantBOMIDs) {
				t.Fatalf("status=%s IDs=%v", report.Status, ids)
			}
		})
	}
}

func TestGraphExpansionInstancesAndBOMScopes(t *testing.T) {
	macro := macroDef("m", "enrich", "lookup users user OUTPUT role")
	users := evalDef("users", "lookup", "users", "")
	req := evalRequest("`enrich` | lookup users user OUTPUT role", []Collection{{Kind: "macro", Coverage: "complete"}, {Kind: "lookup", Coverage: "complete"}}, macro, users)
	report, err := Evaluate(req)
	if err != nil {
		t.Fatal(err)
	}
	nodes := map[string]bool{}
	expansions := 0
	for _, node := range report.Graph.Nodes {
		nodes[node.ID] = true
		if node.EvidencePointer == "" {
			t.Fatalf("node missing evidence pointer: %+v", node)
		}
		if node.Kind == "expansion" {
			expansions++
			if node.InstanceID == "" || node.ObjectID != "m" {
				t.Fatalf("expansion node: %+v", node)
			}
		}
	}
	if expansions == 0 {
		t.Fatalf("no expansion nodes: %+v", report.Graph.Nodes)
	}
	for _, edge := range report.Graph.Edges {
		for _, id := range edge.InvocationNodeIDs {
			if !nodes[id] {
				t.Fatalf("dangling expansion node %s", id)
			}
		}
	}
	var lookup BOMEntry
	for _, entry := range report.BOM {
		if entry.ObjectID == "users" {
			lookup = entry
		}
	}
	if !lookup.Direct || !lookup.Transitive || len(lookup.Occurrences) != 2 {
		t.Fatalf("direct and transitive lookup: %+v", lookup)
	}
	inventory := FormatInventory(report)
	if !strings.Contains(inventory, "direct: users") || !strings.Contains(inventory, "transitive: users") {
		t.Fatalf("inventory scopes: %s", inventory)
	}
}

func reportPointerExists(t *testing.T, report *Report, pointer string) bool {
	t.Helper()
	if !strings.HasPrefix(pointer, "/") {
		return false
	}
	data, err := json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	var current any
	if err := json.Unmarshal(data, &current); err != nil {
		t.Fatal(err)
	}
	for _, part := range strings.Split(pointer[1:], "/") {
		switch value := current.(type) {
		case map[string]any:
			next, ok := value[part]
			if !ok {
				return false
			}
			current = next
		case []any:
			index, err := strconv.Atoi(part)
			if err != nil || index < 0 || index >= len(value) {
				return false
			}
			current = value[index]
		default:
			return false
		}
	}
	return true
}
