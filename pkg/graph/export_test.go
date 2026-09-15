package graph

import (
	"bytes"
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/corpus"
	"github.com/delgado-jacob/spl-toolkit/pkg/validation"
)

type graphCase struct {
	Name      string                   `json:"name"`
	Documents []corpus.RequestDocument `json:"documents"`
}

func graphCases(t *testing.T) []graphCase {
	t.Helper()
	raw, err := os.ReadFile("../../testdata/tooling/graph-cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []graphCase
	if err := json.Unmarshal(raw, &cases); err != nil {
		t.Fatal(err)
	}
	return cases
}

func scanCase(t *testing.T, name string) *corpus.Report {
	t.Helper()
	for _, tc := range graphCases(t) {
		if tc.Name == name {
			r, err := corpus.Scan(corpus.Request{SchemaVersion: 1, Documents: tc.Documents})
			if err != nil {
				t.Fatal(err)
			}
			return r
		}
	}
	t.Fatalf("missing graph case %q", name)
	return nil
}

func TestGraphEndpointIntegrity(t *testing.T) {
	report := scanCase(t, "mixed")
	got, err := Export(report)
	if err != nil {
		t.Fatal(err)
	}
	ids := map[string]bool{}
	for _, node := range got.Nodes {
		if ids[node.ID] {
			t.Fatalf("duplicate node %q", node.ID)
		}
		ids[node.ID] = true
		if node.ReportPointer == "" || !strings.HasPrefix(node.ReportPointer, "/") {
			t.Fatalf("unresolvable node pointer: %+v", node)
		}
		if !pointerExists(t, report, node.ReportPointer) {
			t.Fatalf("node pointer absent: %+v", node)
		}
	}
	for _, edge := range got.Edges {
		if ids[edge.ID] || !ids[edge.From] || !ids[edge.To] {
			t.Fatalf("duplicate edge or missing endpoint: %+v", edge)
		}
		ids[edge.ID] = true
		if edge.ReportPointer == "" || !strings.HasPrefix(edge.ReportPointer, "/") {
			t.Fatalf("missing edge evidence: %+v", edge)
		}
		if !pointerExists(t, report, edge.ReportPointer) {
			t.Fatalf("edge pointer absent: %+v", edge)
		}
	}
	if len(got.Nodes) == 0 || len(got.Edges) == 0 || got.Status != report.Status || got.Coverage != report.Coverage {
		t.Fatalf("lost report evidence: %+v", got)
	}
	a, _ := json.Marshal(got)
	for i := 0; i < 5; i++ {
		b, err := Export(report)
		if err != nil {
			t.Fatal(err)
		}
		encoded, _ := json.Marshal(b)
		if !bytes.Equal(a, encoded) {
			t.Fatal("nondeterministic graph JSON")
		}
	}
	// A broken canonical pointer must fail, not become an invented graph endpoint.
	for i := range report.Entries {
		if a := report.Entries[i].Evaluation.Analysis; a != nil && len(a.References) > 0 {
			a.References[0].StageID = "missing-stage"
			if _, err := Export(report); err == nil {
				t.Fatal("accepted dangling stage reference")
			}
			break
		}
	}
}

func TestGraphExcludesRequirementSpecificProjection(t *testing.T) {
	report := scanCase(t, "mixed")
	got, err := Export(report)
	if err != nil {
		t.Fatal(err)
	}
	for _, node := range got.Nodes {
		if strings.Contains(strings.ToLower(node.Kind), "requirement") || strings.Contains(node.ReportPointer, "/requirements") {
			t.Fatalf("requirement-specific graph node introduced: %+v", node)
		}
	}
	for _, edge := range got.Edges {
		if strings.Contains(strings.ToLower(edge.Relation), "requirement") || strings.Contains(edge.ReportPointer, "/requirements") {
			t.Fatalf("requirement-specific graph edge introduced: %+v", edge)
		}
	}
}

func pointerExists(t *testing.T, value any, ptr string) bool {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	var current any
	if err := json.Unmarshal(raw, &current); err != nil {
		t.Fatal(err)
	}
	for _, component := range strings.Split(strings.TrimPrefix(ptr, "/"), "/") {
		switch typed := current.(type) {
		case map[string]any:
			var ok bool
			current, ok = typed[component]
			if !ok {
				return false
			}
		case []any:
			i, err := strconv.Atoi(component)
			if err != nil || i < 0 || i >= len(typed) {
				return false
			}
			current = typed[i]
		default:
			return false
		}
	}
	return current != nil
}

func TestGraphRevisionIsolation(t *testing.T) {
	report := scanCase(t, "revision")
	got, err := Export(report)
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]map[string]bool{}
	for _, n := range got.Nodes {
		if n.Kind != "reference" {
			continue
		}
		if seen[n.DocumentID] == nil {
			seen[n.DocumentID] = map[string]bool{}
		}
		seen[n.DocumentID][n.ID] = true
	}
	if len(seen) != 2 {
		t.Fatalf("expected both revisions, got %v", seen)
	}
	for id := range seen["spl"] {
		if seen["spl2"][id] {
			t.Fatalf("cross-revision reference collision %q", id)
		}
	}
}

func TestGraphTargetContentIsolation(t *testing.T) {
	request := corpus.Request{SchemaVersion: 1, Documents: []corpus.RequestDocument{{ID: "same", Document: analysis.QueryDocument{Text: "search host=web | table host"}}}}
	var ids []string
	for _, field := range []string{"host", "other"} {
		request.ValidationTarget = &corpus.ValidationTarget{Kind: "field_list", Catalog: &validation.FieldCatalog{Fields: []string{field}, Identity: "same-label", Version: "same-version"}}
		r, err := corpus.Scan(request)
		if err != nil {
			t.Fatal(err)
		}
		g, err := Export(r)
		if err != nil {
			t.Fatal(err)
		}
		for _, n := range g.Nodes {
			if n.Kind == "reference" {
				ids = append(ids, n.ID)
				break
			}
		}
	}
	if len(ids) != 2 || ids[0] == ids[1] {
		t.Fatalf("target content collision: %v", ids)
	}
}

func TestGraphRejectsDanglingDependencySummary(t *testing.T) {
	r := scanCase(t, "sql")
	if len(r.Dependencies) == 0 {
		t.Fatal("fixture has no dependency")
	}
	r.Dependencies[0].ReferenceIDs = append(r.Dependencies[0].ReferenceIDs, "missing-ref")
	if _, err := Export(r); err == nil {
		t.Fatal("accepted dangling dependency reference")
	}
}

func TestGraphConditionalLineage(t *testing.T) {
	report := scanCase(t, "conditional")
	got, err := Export(report)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != analysis.Incomplete {
		t.Fatalf("expected incomplete canonical status, got %s", got.Status)
	}
	if len(got.CoverageReasons) == 0 {
		t.Fatal("lost incomplete coverage reasons")
	}
	uncertain := false
	for _, n := range got.Nodes {
		if n.Kind == "phase" && n.Before != nil && n.After != nil && (n.Before.Uncertain || n.After.Uncertain || n.Before.Open || n.After.Open) {
			uncertain = true
		}
	}
	if !uncertain {
		t.Fatal("lost uncertain canonical lineage")
	}
	for _, e := range got.Edges {
		if e.Relation == "complete_derivation" {
			t.Fatalf("fabricated complete edge %+v", e)
		}
	}
}

func TestGraphConditionalTransition(t *testing.T) {
	r := scanCase(t, "conditional_transition")
	g, err := Export(r)
	if err != nil {
		t.Fatal(err)
	}
	var transitionID string
	for _, n := range g.Nodes {
		if n.Kind == "transition" && n.Operation == "rename" && n.Name == "b" {
			if !n.Conditional {
				t.Fatal("conditional rename promoted to certain")
			}
			transitionID = n.ID
		}
	}
	if transitionID == "" {
		t.Fatal("missing conditional rename transition")
	}
	foundConditionalEdge := false
	for _, e := range g.Edges {
		if e.To == transitionID && e.Relation == "transition_input" && e.Conditional {
			foundConditionalEdge = true
		}
	}
	if !foundConditionalEdge {
		t.Fatal("conditional lineage edge lost")
	}
}

func TestGraphPreservesAcquisitionFailure(t *testing.T) {
	p, err := corpus.Prepare(corpus.ScanOptions{})
	if err != nil {
		t.Fatal(err)
	}
	r, err := p.Scan(corpus.Input{Selection: corpus.Selection{Mode: "manifest", Complete: true}, Entries: []corpus.Entry{{ID: "ok", Origin: corpus.Origin{Kind: "inline"}, Document: &analysis.QueryDocument{Text: "search host=web"}}, {ID: "missing", Origin: corpus.Origin{Kind: "file", RelativePath: "missing.spl"}, Failure: &corpus.AcquisitionError{Code: "not_found", Phase: "open", Path: "missing.spl"}}}})
	if err != nil {
		t.Fatal(err)
	}
	g, err := Export(r)
	if err != nil {
		t.Fatal(err)
	}
	if g.ExecutionComplete || g.Status != analysis.Incomplete {
		t.Fatalf("lost operational incompleteness: %+v", g)
	}
	found := false
	for _, n := range g.Nodes {
		if n.Kind == "document" && n.DocumentID == "missing" {
			found = n.Failure != nil && n.Failure.Code == "not_found" && pointerExists(t, r, n.ReportPointer)
		}
	}
	if !found {
		t.Fatal("missing acquisition failure evidence")
	}
}

func TestGraphSQLPhaseOrder(t *testing.T) {
	report := scanCase(t, "sql")
	got, err := Export(report)
	if err != nil {
		t.Fatal(err)
	}
	var lexical, logical []string
	var selectPhases []string
	for _, n := range got.Nodes {
		if n.Kind == "stage" {
			lexical = append(lexical, n.Command)
		}
		if n.Kind == "phase" {
			logical = append(logical, n.Phase)
			if n.StageID == "stage-0" {
				selectPhases = append(selectPhases, n.ID)
			}
		}
	}
	if strings.Join(lexical, ",") != "select,from,where" || strings.Join(logical, ",") != "source,filter,evaluate,project" {
		t.Fatalf("lexical %v logical %v", lexical, logical)
	}
	if len(selectPhases) != 2 || selectPhases[0] == selectPhases[1] {
		t.Fatalf("SELECT-owned phases conflated: %v", selectPhases)
	}
	for _, n := range got.Nodes {
		if n.Kind == "reference" && n.Location != nil {
			doc := report.Entries[0].Evaluation.Analysis.Document.Text
			start, end := n.Location.Start.Offset, n.Location.End.Offset
			if start < 0 || end > len(doc) || doc[start:end] != n.OriginalName {
				t.Fatalf("source slice mismatch: %+v", n)
			}
		}
	}
}

func TestGraphDetachedPhaseAndScopeIdentity(t *testing.T) {
	report := scanCase(t, "scopes")
	got, err := Export(report)
	if err != nil {
		t.Fatal(err)
	}
	var sameName []Node
	for _, n := range got.Nodes {
		if n.Kind == "reference" && n.Name == "root" {
			sameName = append(sameName, n)
		}
	}
	if len(sameName) < 2 {
		t.Fatalf("missing repeated field evidence: %+v", sameName)
	}
	separate := false
	for i := range sameName {
		for j := i + 1; j < len(sameName); j++ {
			if sameName[i].ScopeID != sameName[j].ScopeID && sameName[i].ID != sameName[j].ID {
				separate = true
			}
		}
	}
	if !separate {
		t.Fatalf("repeated names conflated: %+v", sameName)
	}
	before, _ := json.Marshal(got)
	for i := range report.Entries[0].Evaluation.Analysis.Lineage {
		l := &report.Entries[0].Evaluation.Analysis.Lineage[i]
		if l.ExecutionOrder != nil {
			*l.ExecutionOrder = 999
		}
		if len(l.After.Fields) > 0 {
			l.After.Fields[0].Name = "mutated"
		}
	}
	after, _ := json.Marshal(got)
	if !bytes.Equal(before, after) {
		t.Fatal("export aliases mutable canonical lineage")
	}
}

func TestGraphUnicodeSourceSlices(t *testing.T) {
	r, err := corpus.Scan(corpus.Request{SchemaVersion: 1, Documents: []corpus.RequestDocument{{ID: "unicode", Document: analysis.QueryDocument{Text: "search note=\"😀\" | table host"}}}})
	if err != nil {
		t.Fatal(err)
	}
	g, err := Export(r)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, n := range g.Nodes {
		if n.Kind != "reference" || n.Location == nil {
			continue
		}
		start, end := n.Location.Start.Offset, n.Location.End.Offset
		if start < 0 || end > len(r.Entries[0].Evaluation.Analysis.Document.Text) || r.Entries[0].Evaluation.Analysis.Document.Text[start:end] != n.OriginalName {
			t.Fatalf("bad UTF-8 source slice: %+v", n)
		}
		if n.Name == "host" {
			found = true
		}
	}
	if !found {
		t.Fatal("missing post-emoji reference")
	}
}
