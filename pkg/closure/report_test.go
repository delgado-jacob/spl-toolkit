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

func TestStandaloneGraphRetainsOrderedTraversalOrigins(t *testing.T) {
	macro := macroDef("m", "m", "lookup us$x$ user OUTPUT role", "x")
	lookup := evalDef("users", "lookup", "users", "")
	report, err := Evaluate(evalRequest("`m(ers)`", []Collection{{Kind: "macro", Coverage: "complete"}, {Kind: "lookup", Coverage: "complete"}}, macro, lookup))
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(report.Graph)
	if err != nil {
		t.Fatal(err)
	}
	var graph DependencyGraph
	if err := json.Unmarshal(raw, &graph); err != nil {
		t.Fatal(err)
	}
	for _, edge := range graph.Edges {
		if edge.Kind != "lookup" || edge.Name != "users" {
			continue
		}
		var index int
		if _, err := fmt.Sscanf(edge.TraversalPointer, "/traversal/%d", &index); err != nil || index >= len(report.Traversal) {
			t.Fatalf("invalid traversal pointer: %+v", edge)
		}
		traversal := report.Traversal[index]
		if len(traversal.Origins) < 3 || !reflect.DeepEqual(edge.Origins, traversal.Origins) || edge.Source != edge.Origins[0] {
			t.Fatalf("graph origins=%+v traversal origins=%+v source=%+v", edge.Origins, traversal.Origins, edge.Source)
		}
		if edge.Origins[0].Kind != "definition" || edge.Origins[1].Kind != "query" || edge.Origins[2].Kind != "definition" {
			t.Fatalf("unexpected ordered origins: %+v", edge.Origins)
		}
		report.Traversal[index].Origins[0].Start = -1
		if report.Graph.Edges[index].Origins[0].Start < 0 {
			t.Fatal("graph origins alias traversal origins")
		}
		return
	}
	t.Fatalf("lookup edge missing from standalone graph: %+v", graph.Edges)
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

func TestBOMRootMacroSelfCycleIsIncomplete(t *testing.T) {
	macro := macroDef("m0", "m", "`m`")
	report, err := Evaluate(evalRequest("`m`", []Collection{{Kind: "macro", Coverage: "complete"}}, macro))
	if err != nil {
		t.Fatal(err)
	}
	if len(report.BOM) != 1 || report.BOM[0].ObjectID != "m0" {
		t.Fatalf("BOM: %+v", report.BOM)
	}
	cycle := false
	for _, edge := range report.Graph.Edges {
		if edge.Resolution == "cycle" {
			cycle = true
			if edge.To != "object:m0" || edge.Unknown {
				t.Fatalf("proven cycle target is unknown: %+v", edge)
			}
		}
	}
	cycleEdgeID := ""
	for _, edge := range report.Traversal {
		if edge.Resolution == "cycle" {
			if edge.ToObjectID != "m0" {
				t.Fatalf("cycle traversal target: %+v", edge)
			}
			cycleEdgeID = edge.ID
		}
	}
	if len(report.BOM[0].Occurrences) != 2 || cycleEdgeID == "" {
		t.Fatalf("cycle occurrence omitted: %+v", report.BOM[0])
	}
	occurrenceFound := false
	for _, occurrence := range report.BOM[0].Occurrences {
		if occurrence.EdgeID == cycleEdgeID {
			occurrenceFound = true
		}
	}
	if !occurrenceFound {
		t.Fatalf("BOM omits cycle traversal %s: %+v", cycleEdgeID, report.BOM[0])
	}
	if !cycle || !report.BOM[0].Incomplete {
		t.Fatalf("cycle edge/BOM mismatch: graph=%+v BOM=%+v gaps=%+v", report.Graph.Edges, report.BOM, report.Gaps)
	}
	if !strings.Contains(FormatInventory(report), "m (m0) [incomplete]") {
		t.Fatalf("inventory cycle: %s", FormatInventory(report))
	}
}

func TestBOMMissingMacroOverloadDoesNotTaintResolvedObject(t *testing.T) {
	zero := macroDef("m0", "m", "eval x=1")
	report, err := Evaluate(evalRequest("`m` | `m(1)`", []Collection{{Kind: "macro", Coverage: "complete"}}, zero))
	if err != nil {
		t.Fatal(err)
	}
	if len(report.BOM) != 1 || report.BOM[0].ObjectID != "m0" {
		t.Fatalf("BOM: %+v", report.BOM)
	}
	if report.BOM[0].Incomplete {
		t.Fatalf("missing arity-one overload tainted m/0: BOM=%+v traversal=%+v gaps=%+v", report.BOM, report.Traversal, report.Gaps)
	}
	inventory := FormatInventory(report)
	if !strings.Contains(inventory, "missing_object") || strings.Contains(inventory, "m (m0) [incomplete]") {
		t.Fatalf("inventory overload: %s", inventory)
	}
}

func TestBOMPartialCollectionMarksResolvedOccurrence(t *testing.T) {
	users := evalDef("users", "lookup", "users", "")
	report, err := Evaluate(evalRequest("| lookup users user OUTPUT role", []Collection{{Kind: "lookup", Coverage: "partial"}}, users))
	if err != nil {
		t.Fatal(err)
	}
	if len(report.BOM) != 1 || report.BOM[0].ObjectID != "users" || !report.BOM[0].Incomplete {
		t.Fatalf("partial collection object: BOM=%+v traversal=%+v gaps=%+v", report.BOM, report.Traversal, report.Gaps)
	}
	if !strings.Contains(FormatInventory(report), "users (users) [incomplete]") || !strings.Contains(FormatInventory(report), "collection_incomplete") {
		t.Fatalf("inventory partial collection: %s", FormatInventory(report))
	}
}

func TestBOMHeldRootMacroExpansionIsIncomplete(t *testing.T) {
	cases := []struct {
		name       string
		definition Definition
		reason     string
	}{
		{name: "missing_body", definition: func() Definition { d := macroDef("m0", "m", ""); d.Document = nil; return d }(), reason: "expansion_missing"},
		{name: "eval_based", definition: func() Definition { d := macroDef("m0", "m", "eval x=1"); value := true; d.EvalBased = &value; return d }(), reason: "expansion_eval_based"},
		{name: "validation_held", definition: func() Definition {
			d := macroDef("m0", "m", "eval x=1")
			value := "isnum($x$)"
			d.Validation = &value
			return d
		}(), reason: "expansion_validation_held"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			report, err := Evaluate(evalRequest("`m`", []Collection{{Kind: "macro", Coverage: "complete"}}, tc.definition))
			if err != nil {
				t.Fatal(err)
			}
			if len(report.BOM) != 1 || report.BOM[0].ObjectID != "m0" {
				t.Fatalf("BOM: %+v", report.BOM)
			}
			foundGap := false
			for _, gap := range report.Gaps {
				if gap.Code == tc.reason {
					foundGap = true
				}
			}
			if !foundGap || !report.BOM[0].Incomplete {
				t.Fatalf("held expansion status: gap=%t BOM=%+v traversal=%+v gaps=%+v", foundGap, report.BOM, report.Traversal, report.Gaps)
			}
			if !strings.Contains(FormatInventory(report), "m (m0) [incomplete]") || !strings.Contains(FormatInventory(report), tc.reason) {
				t.Fatalf("inventory held expansion: %s", FormatInventory(report))
			}
		})
	}
}

func TestGraphBoundCycleUsesActiveTarget(t *testing.T) {
	a := macroDef("a", "m", "`m` | eval a=1")
	b := macroDef("b", "m", "eval b=1 | `m`")
	root := "`m`"
	nested := strings.Index(b.Document.Text, "`m`")
	req := evalRequest(root, []Collection{{Kind: "macro", Coverage: "complete"}}, a, b)
	req.Bindings = []Binding{
		bindingFor(root, 0, len(root), "a"),
		bindingFor(a.Document.Text, 0, len(root), "b"),
		bindingFor(b.Document.Text, nested, nested+len(root), "b"),
	}
	report, err := Evaluate(req)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, edge := range report.Graph.Edges {
		if edge.Resolution == "cycle" {
			found = true
			if edge.To != "object:b" || edge.Unknown || !reflect.DeepEqual(edge.CyclePath, []string{"b", "b"}) {
				t.Fatalf("bound cycle target: %+v", edge)
			}
		}
	}
	if !found {
		t.Fatalf("cycle edge absent: %+v", report.Graph.Edges)
	}
}
