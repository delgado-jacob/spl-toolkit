package workflow

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/closure"
	"github.com/delgado-jacob/spl-toolkit/pkg/compatibility"
	"github.com/delgado-jacob/spl-toolkit/pkg/corpus"
	"github.com/delgado-jacob/spl-toolkit/pkg/environment"
)

// Retained closure evidence uses the canonical owner for macro expansion and
// traversal. Workflow exports have no reason to regenerate this owner evidence.
func closureExportFixture(t *testing.T, text string, definitions []closure.Definition) *Report {
	t.Helper()
	doc := analysis.QueryDocument{Text: text, Language: "spl", Profile: "splunkd", Version: "current", SourceID: "d1"}
	c, err := closure.Evaluate(closure.Request{SchemaVersion: 1, Document: doc, Bundle: closure.DefinitionBundle{SchemaVersion: 1, ScopeID: "capture", Collections: []closure.Collection{{Kind: "macro", Coverage: "complete"}, {Kind: "lookup", Coverage: "complete"}, {Kind: "saved_search", Coverage: "complete"}}, Objects: definitions}, Bindings: []closure.Binding{}})
	if err != nil {
		t.Fatal(err)
	}
	r, err := Assess(seedRequest(t, "from [{id:1}]"))
	if err != nil {
		t.Fatal(err)
	}
	r.Entries[0].Analysis = c.DirectAnalysis
	r.Entries[0].SourceHash = corpus.SourceHash(text)
	r.Entries[0].Compatibility = &compatibility.Report{Closure: c}
	return r
}
func macroExportDefinition(id, name, text string) closure.Definition {
	zero := 0
	no := false
	doc := analysis.QueryDocument{Text: text, Language: "spl", Profile: "splunkd", Version: "current", SourceID: id}
	return closure.Definition{ID: id, Kind: "macro", Name: name, SourceID: id, Document: &doc, Arity: &zero, Arguments: []string{}, EvalBased: &no, Relations: []closure.Relation{}}
}
func TestWorkflowBOMCapturedIdentityOccurrencesAndIsolation(t *testing.T) {
	r := closureExportFixture(t, "`m` | `m`", []closure.Definition{macroExportDefinition("macro-id", "m", "lookup users user OUTPUT role | lookup missing user OUTPUT role"), {ID: "lookup-id", Kind: "lookup", Name: "users", SourceID: "users", Relations: []closure.Relation{}}})
	// The same captured object occurs in another detection without merging roots.
	other := r.Entries[0]
	other.ID = "d2"
	r.Entries = append(r.Entries, other)
	before, _ := json.Marshal(r)
	g, err := ExportGraph(r)
	if err != nil {
		t.Fatal(err)
	}
	unknown := false
	for _, s := range g.Subjects {
		assertClosureGraph(t, r, s.Closure)
		for _, e := range s.Closure.Edges {
			unknown = unknown || e.Unknown
		}
	}
	if !unknown {
		t.Fatal("unknown closure edge lost")
	}
	b, err := ExportBOM(r)
	if err != nil {
		t.Fatal(err)
	}
	if len(b.Subjects) != 2 || len(b.SharedDependencies) != 2 {
		t.Fatalf("unexpected captured index: %+v", b)
	}
	transitive := false
	for _, s := range b.Subjects {
		workflowPointer(t, r, s.EvidencePointer)
		for _, e := range s.Entries {
			if e.ObjectID == "lookup-id" {
				transitive = e.Transitive && !e.Direct && len(e.Occurrences) == 2
			}
			if e.Name == "missing" {
				t.Fatal("unknown object invented")
			}
		}
	}
	if !transitive {
		t.Fatalf("transitive occurrences lost: %+v", b.Subjects)
	}
	for _, shared := range b.SharedDependencies {
		if shared.EnvironmentDigest != r.Provenance.EnvironmentDigest || len(shared.OccurrencePointers) != 4 {
			t.Fatalf("shared occurrence index: %+v", shared)
		}
		for _, p := range shared.OccurrencePointers {
			workflowPointer(t, r, p)
		}
	}
	again, err := ExportBOM(r)
	if err != nil {
		t.Fatal(err)
	}
	one, _ := json.Marshal(b)
	two, _ := json.Marshal(again)
	if !bytes.Equal(one, two) {
		t.Fatal("nondeterministic BOM")
	}
	b.Subjects[0].Entries[0].Occurrences[0].Path[0] = "mutated"
	b.SharedDependencies[0].OccurrencePointers[0] = "mutated"
	g.Subjects[0].Closure.Edges[0].Path = append(g.Subjects[0].Closure.Edges[0].Path, "mutated")
	after, _ := json.Marshal(r)
	if !bytes.Equal(before, after) {
		t.Fatal("export mutated or aliases canonical evidence")
	}
}

func TestWorkflowBOMCycleAndDistinctCapturedIDs(t *testing.T) {
	property := "/next"
	a := closure.Definition{ID: "a", Kind: "saved_search", Name: "A", SourceID: "a", Relations: []closure.Relation{{Kind: "saved_search", Name: "B", Property: &property}}}
	b := closure.Definition{ID: "b", Kind: "saved_search", Name: "B", SourceID: "b", Relations: []closure.Relation{{Kind: "saved_search", Name: "A", Property: &property}}}
	r := closureExportFixture(t, "| from savedsearch:A", []closure.Definition{a, b})
	g, err := ExportGraph(r)
	if err != nil {
		t.Fatal(err)
	}
	cycle := false
	for _, e := range g.Subjects[0].Closure.Edges {
		cycle = cycle || len(e.CyclePath) > 0
	}
	if !cycle {
		t.Fatal("cycle edge lost")
	}
	bom, err := ExportBOM(r)
	if err != nil {
		t.Fatal(err)
	}
	if len(bom.Subjects[0].Entries) != 2 {
		t.Fatalf("cycle BOM: %+v", bom)
	}
	for _, e := range bom.Subjects[0].Entries {
		if !e.Incomplete {
			t.Fatal("cycle completeness lost")
		}
	}
	// Separate scopes can capture different objects with the same display name.
	one := closureExportFixture(t, "lookup users user OUTPUT role", []closure.Definition{{ID: "first", Kind: "lookup", Name: "users", SourceID: "first", Relations: []closure.Relation{}}})
	two := closureExportFixture(t, "lookup users user OUTPUT role", []closure.Definition{{ID: "second", Kind: "lookup", Name: "users", SourceID: "second", Relations: []closure.Relation{}}})
	other := two.Entries[0]
	other.ID = "d2"
	one.Entries = append(one.Entries, other)
	bom, err = ExportBOM(one)
	if err != nil {
		t.Fatal(err)
	}
	if len(bom.SharedDependencies) != 2 || bom.SharedDependencies[0].ObjectID == bom.SharedDependencies[1].ObjectID {
		t.Fatalf("display names coalesced: %+v", bom.SharedDependencies)
	}
}

func TestWorkflowExportsMissingClosureAndFailures(t *testing.T) {
	r, err := Assess(seedRequest(t, "from [{id:1}]"))
	if err != nil {
		t.Fatal(err)
	}
	r.Entries[0].Compatibility.Closure = nil
	g, err := ExportGraph(r)
	if err != nil {
		t.Fatal(err)
	}
	if g.Subjects[0].ClosureCoverage != "not_applicable" {
		t.Fatalf("literal coverage: %+v", g.Subjects[0])
	}
	r, err = Assess(seedRequest(t, "from events_good"))
	if err != nil {
		t.Fatal(err)
	}
	r.Entries[0].Compatibility.Closure = nil
	g, err = ExportGraph(r)
	if err != nil {
		t.Fatal(err)
	}
	if g.Subjects[0].ClosureCoverage != "unavailable" {
		t.Fatalf("missing closure claimed complete: %+v", g.Subjects[0])
	}
	r.Entries = append(r.Entries, ReportEntry{ID: "failed", Failure: &Failure{Phase: "acquisition", Code: "unreadable", Message: "failure"}})
	bom, err := ExportBOM(r)
	if err != nil {
		t.Fatal(err)
	}
	if len(bom.Subjects[0].Entries) != 0 || bom.Subjects[1].Failure == nil || bom.Subjects[1].ClosureCoverage != "unavailable" {
		t.Fatalf("failure subject lost: %+v", bom.Subjects)
	}
	bom.Subjects[1].Failure.Message = "changed"
	if r.Entries[1].Failure.Message != "failure" {
		t.Fatal("failure aliases report")
	}
	if _, err := ExportGraph(nil); err == nil {
		t.Fatal("nil report accepted")
	}
	if _, err := ExportBOM(&Report{SchemaVersion: 99}); err == nil {
		t.Fatal("unknown report version accepted")
	}
}

func TestWorkflowDatasetDefinitionExportsAssessedEvidence(t *testing.T) {
	req := seedRequest(t, "from events_good | fields id")
	direct, err := analysis.Analyze(req.Documents[0].Document)
	if err != nil {
		t.Fatal(err)
	}
	for i := range req.Settings.Snapshot.Objects {
		o := &req.Settings.Snapshot.Objects[i]
		if o.Kind == "dataset" && o.Name == "events_good" {
			o.Document = &analysis.QueryDocument{Text: "from [{id:1}] | fields id", Language: "spl2", SourceID: "dataset-body"}
			req.Settings.Entries[0].Compatibility.InputBindings = []compatibility.InputBinding{{InputID: direct.Inputs[0].ID, ObjectID: o.ID, Expected: environment.ObjectIdentity{Kind: o.Kind, Name: o.Name, Namespace: o.Namespace, App: o.App, Owner: o.Owner}, SchemaID: "schema-0"}}
		}
	}
	r, err := Assess(req)
	if err != nil {
		t.Fatal(err)
	}
	if r.Entries[0].Failure != nil || r.CIExitCode != 0 {
		t.Fatalf("dataset assessment: %+v", r.Entries[0])
	}
	g, err := ExportGraph(r)
	if err != nil {
		t.Fatal(err)
	}
	assertClosureGraph(t, r, g.Subjects[0].Closure)
	b, err := ExportBOM(r)
	if err != nil {
		t.Fatal(err)
	}
	if len(b.Subjects[0].Entries) != 1 || !b.Subjects[0].Entries[0].Direct {
		t.Fatalf("captured dataset omitted: %+v", b)
	}
	for _, s := range b.SharedDependencies {
		for _, p := range s.OccurrencePointers {
			workflowPointer(t, r, p)
		}
	}
	edges := map[string]bool{}
	for _, e := range g.Subjects[0].Closure.Edges {
		edges[e.ID] = true
	}
	for _, entry := range b.Subjects[0].Entries {
		for _, o := range entry.Occurrences {
			if !edges[o.EdgeID] {
				t.Fatalf("BOM occurrence edge missing from graph %s", o.EdgeID)
			}
		}
	}
}
