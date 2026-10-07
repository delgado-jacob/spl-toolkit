package workflow

import (
	"encoding/json"
	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/compatibility"
	"github.com/delgado-jacob/spl-toolkit/pkg/corpus"
	"github.com/delgado-jacob/spl-toolkit/pkg/environment"
	"reflect"
	"strings"
	"testing"
)

func TestAssessCompleteLiteralQuery(t *testing.T) {
	req := seedRequest(t, "from [{id:1}]")
	before, _ := json.Marshal(req)
	report, err := Assess(req)
	if err != nil {
		t.Fatal(err)
	}
	after, _ := json.Marshal(req)
	if !reflect.DeepEqual(before, after) {
		t.Fatal("request mutated")
	}
	if report.CIExitCode != 0 || !report.ExecutionComplete || len(report.Entries) != 1 {
		t.Fatalf("report: %+v", report)
	}
	entry := report.Entries[0]
	if entry.Status != analysis.Valid || entry.Compatibility.Outcome != "satisfied" {
		t.Fatalf("entry: %+v", entry)
	}
	if !reflect.DeepEqual(entry.Analysis.Requirements, entry.Compatibility.Requirements) {
		t.Fatal("original requirements changed")
	}
}

func inlineInput(req Request) corpus.Input {
	input := corpus.Input{Selection: corpus.Selection{Mode: "inline", Complete: true}, Entries: []corpus.Entry{}}
	for _, e := range req.Documents {
		doc := e.Document
		input.Entries = append(input.Entries, corpus.Entry{ID: e.ID, Document: &doc, Origin: corpus.Origin{Kind: "inline"}, SourceHash: corpus.SourceHash(doc.Text)})
	}
	return input
}

func TestAssessContentStatuses(t *testing.T) {
	for _, tc := range []struct {
		name, text string
		status     analysis.Status
		code       int
	}{
		{"literal", "from [{id:1}]", analysis.Valid, 0},
		{"malformed", "from [{id:1}] | eval broken =", analysis.Invalid, 1},
		{"unknown command", "from [{id:1}] | unknowable_command", analysis.Incomplete, 3},
		{"missing field", "from events_good | fields absent", analysis.Invalid, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := seedRequest(t, tc.text)
			if tc.name == "missing field" {
				// Find the unique canonical source identity, independently of schema membership.
				result, err := analysis.Analyze(req.Documents[0].Document)
				if err != nil {
					t.Fatal(err)
				}
				if len(result.Inputs) != 1 || result.Inputs[0].Identity.Value != "events_good" {
					t.Fatalf("source identity: %+v", result.Inputs)
				}
				for _, object := range req.Settings.Snapshot.Objects {
					if object.Kind == "dataset" && object.Name == "events_good" {
						req.Settings.Entries[0].Compatibility.InputBindings = []compatibility.InputBinding{{InputID: result.Inputs[0].ID, ObjectID: object.ID, Expected: environment.ObjectIdentity{Kind: object.Kind, Name: object.Name, Namespace: object.Namespace, App: object.App, Owner: object.Owner}, SchemaID: "schema-0"}}
					}
				}
			}
			raw, _ := json.Marshal(req)
			p, err := Prepare(req.Settings)
			if err != nil {
				t.Fatal(err)
			}
			for name, assess := range map[string]func() (*Report, error){"typed": func() (*Report, error) { return Assess(req) }, "json": func() (*Report, error) { return AssessJSON(raw) }, "prepared": func() (*Report, error) { return p.Assess(inlineInput(req)) }} {
				t.Run(name, func(t *testing.T) {
					r, err := assess()
					if err != nil {
						t.Fatal(err)
					}
					if !r.ExecutionComplete || r.CIExitCode != tc.code || r.Status != tc.status || r.Counts.Assessed != 1 {
						t.Fatalf("report: %+v, entry: %+v", r, r.Entries[0])
					}
				})
			}
		})
	}
}

func TestAssessSelectionOrderFailuresAndIsolation(t *testing.T) {
	req := seedRequest(t, "from [{id:1}]")
	req.Settings.Entries = append(req.Settings.Entries, EntrySettings{ID: "failed", Compatibility: detach(req.Settings.Entries[0].Compatibility)})
	req.Documents = append(req.Documents, corpus.RequestDocument{ID: "failed", Document: req.Documents[0].Document})
	p, err := Prepare(req.Settings)
	if err != nil {
		t.Fatal(err)
	}
	input := inlineInput(req)
	input.Entries[1].Document = nil
	input.Entries[1].Failure = &corpus.AcquisitionError{Code: "read_failed", Phase: "acquisition", Message: "cannot read", Path: "failed.spl"}
	input.Entries[0], input.Entries[1] = input.Entries[1], input.Entries[0]
	input.Selection.Complete = false
	input.Selection.TraversalFailures = []corpus.AcquisitionError{{Code: "walk_failed", Phase: "traversal", Message: "cannot walk", Path: "queries"}}
	before, _ := json.Marshal(input)
	first, err := p.Assess(input)
	if err != nil {
		t.Fatal(err)
	}
	if first.ExecutionComplete || first.CIExitCode != 2 || first.Counts.Selected != 2 || first.Counts.Assessed != 1 || first.Counts.TraversalFailed != 1 || first.Counts.AcquisitionFailed != 1 || first.Counts.Valid != 1 || first.Counts.Incomplete != 1 {
		t.Fatalf("report: %+v", first)
	}
	if first.Entries[0].ID != "failed" || first.Entries[1].ID != "d1" || first.Entries[0].Failure.Phase != "acquisition" {
		t.Fatalf("order/failure: %+v", first.Entries)
	}
	after, _ := json.Marshal(input)
	if !reflect.DeepEqual(before, after) {
		t.Fatal("input mutated")
	}
	want, _ := json.Marshal(first)
	first.Selection.TraversalFailures[0].Message = "mutated"
	first.Entries[1].Analysis.Requirements.Query.SourceID = "mutated"
	first.Entries[1].Compatibility.Requirements.Query.SourceID = "mutated"
	req.Settings.Snapshot.ScopeID = "mutated"
	req.Settings.Entries[0].Compatibility.InputBindings = nil
	second, err := p.Assess(input)
	if err != nil {
		t.Fatal(err)
	}
	got, _ := json.Marshal(second)
	if !reflect.DeepEqual(want, got) {
		t.Fatal("reports or settings share mutable state")
	}
	input.Entries[1].Document.Text = "from [{changed:1}]"
	if second.Entries[1].Analysis.Document.Text != "from [{id:1}]" {
		t.Fatal("report retains caller document")
	}
}

func TestAssessAdmissionAndConfigurationFailure(t *testing.T) {
	req := seedRequest(t, "from [{id:1}]")
	p, err := Prepare(req.Settings)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name   string
		change func(*corpus.Input)
	}{
		{"mode", func(i *corpus.Input) { i.Selection.Mode = "" }},
		{"completeness", func(i *corpus.Input) { i.Selection.Complete = false }},
		{"empty", func(i *corpus.Input) { i.Entries = nil }},
		{"duplicate", func(i *corpus.Input) { i.Entries = append(i.Entries, i.Entries[0]) }},
		{"missing settings", func(i *corpus.Input) { i.Entries[0].ID = "other" }},
		{"dual payload", func(i *corpus.Input) { i.Entries[0].Failure = &corpus.AcquisitionError{Code: "failed"} }},
		{"hash", func(i *corpus.Input) { i.Entries[0].SourceHash = "wrong" }},
		{"utf8", func(i *corpus.Input) { i.Entries[0].ID = string([]byte{255}) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			i := inlineInput(req)
			tc.change(&i)
			if _, err := p.Assess(i); err == nil {
				t.Fatal("invalid input admitted")
			}
		})
	}
	req.Settings.Entries[0].Compatibility.InputBindings = []compatibility.InputBinding{{InputID: "unknown", ObjectID: "missing", Expected: environment.ObjectIdentity{Kind: "dataset", Name: "missing"}}}
	r, err := Assess(req)
	if err != nil {
		t.Fatal(err)
	}
	if r.ExecutionComplete || r.CIExitCode != 2 || r.Counts.ConfigurationFailed != 1 || r.Counts.Assessed != 0 || r.Counts.Incomplete != 1 || r.Entries[0].Analysis == nil {
		t.Fatalf("report: %+v", r)
	}
	var detail compatibility.RequestErrorDetail
	if err := json.Unmarshal(r.Entries[0].Failure.Detail, &detail); err != nil {
		t.Fatal(err)
	}
	if detail.Code == "" || detail.Path == "" || detail.Code != r.Entries[0].Failure.Code {
		t.Fatalf("detail: %+v", detail)
	}
}

func TestAssessClosureRetainsImplicitSourceLimit(t *testing.T) {
	req := seedRequest(t, "`source`")
	req.Documents[0].Document.Language = "spl"
	provenance := req.Settings.Snapshot.Objects[0].Provenance
	req.Settings.Snapshot.Capabilities = append(req.Settings.Snapshot.Capabilities, environment.Capability{ID: "language:spl:profile:splunkd", Version: "current", State: "available", Provenance: provenance})
	for i := range req.Settings.Snapshot.Collections {
		if req.Settings.Snapshot.Collections[i].Kind == "macro" {
			req.Settings.Snapshot.Collections[i].Coverage = "complete"
			req.Settings.Snapshot.Collections[i].Reason = ""
		}
	}
	arity := 0
	eval := false
	req.Settings.Snapshot.Objects = append(req.Settings.Snapshot.Objects, environment.Object{ID: "macro", Kind: "macro", Name: "source", Namespace: "search", App: "resolution", Owner: "nobody", Provenance: provenance, Arity: &arity, EvalBased: &eval, Arguments: []string{}, Document: &analysis.QueryDocument{Text: "eval local=1", Language: "spl", SourceID: "macro.spl"}})
	r, err := Assess(req)
	if err != nil {
		t.Fatal(err)
	}
	if r.Entries[0].Analysis.Status != analysis.Incomplete || r.Entries[0].Compatibility.Closure == nil || r.Entries[0].Compatibility.Outcome != "incomplete" || r.Status != analysis.Incomplete || r.CIExitCode != 3 {
		t.Fatalf("closure assessment: status=%s outcome=%s CI=%d", r.Status, r.Entries[0].Compatibility.Outcome, r.CIExitCode)
	}
	if r.Entries[0].Compatibility.Closure.EffectiveAnalysis.Document.Text != "eval local=1" {
		t.Fatal("supported macro body was not expanded")
	}

	req.Settings.Snapshot.Objects[len(req.Settings.Snapshot.Objects)-1].Document = nil
	r, err = Assess(req)
	if err != nil {
		t.Fatal(err)
	}
	if r.Status != analysis.Incomplete || r.CIExitCode != 3 {
		t.Fatalf("unresolved macro: %+v", r)
	}
}

func TestAssessPartialPositiveEvidence(t *testing.T) {
	req := seedRequest(t, "from events_good | fields id")
	result, err := analysis.Analyze(req.Documents[0].Document)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Inputs) != 1 || result.Inputs[0].Identity.Value != "events_good" {
		t.Fatalf("source: %+v", result.Inputs)
	}
	for _, o := range req.Settings.Snapshot.Objects {
		if o.Kind == "dataset" && o.Name == "events_good" {
			req.Settings.Entries[0].Compatibility.InputBindings = []compatibility.InputBinding{{InputID: result.Inputs[0].ID, ObjectID: o.ID, Expected: environment.ObjectIdentity{Kind: o.Kind, Name: o.Name, Namespace: o.Namespace, App: o.App, Owner: o.Owner}, SchemaID: "schema-0"}}
		}
	}
	for i := range req.Settings.Snapshot.Collections {
		req.Settings.Snapshot.Collections[i].Coverage = "partial"
		req.Settings.Snapshot.Collections[i].Reason = "limited capture"
	}
	for i := range req.Settings.SchemaBundle.Bindings {
		req.Settings.SchemaBundle.Bindings[i].SourceCoverage = "partial"
		req.Settings.SchemaBundle.Bindings[i].Reason = "limited source capture"
	}
	r, err := Assess(req)
	if err != nil {
		t.Fatal(err)
	}
	if r.CIExitCode != 0 || r.Status != analysis.Valid || r.Entries[0].Compatibility.Inputs[0].Outcome != "satisfied" {
		t.Fatalf("positive evidence: %+v", r.Entries[0].Compatibility)
	}
}

func TestAssessDefinitionClosureSatisfied(t *testing.T) {
	req := seedRequest(t, "from view | fields id")
	body := analysis.QueryDocument{Text: "from events_good | fields id", Language: "spl2", SourceID: "view.spl"}
	provenance := req.Settings.Snapshot.Objects[0].Provenance
	view := environment.Object{ID: "view", Kind: "dataset", Name: "view", Namespace: "search", App: "resolution", Owner: "nobody", Provenance: provenance, Document: &body}
	req.Settings.Snapshot.Objects = append(req.Settings.Snapshot.Objects, view)
	identity := environment.ObjectIdentity{Kind: view.Kind, Name: view.Name, Namespace: view.Namespace, App: view.App, Owner: view.Owner}
	req.Settings.SchemaBundle.Bindings = append(req.Settings.SchemaBundle.Bindings, environment.SchemaBinding{SchemaID: "schema-0", ObjectID: view.ID, Expected: identity, SourceCoverage: "complete"})
	root, err := analysis.Analyze(req.Documents[0].Document)
	if err != nil {
		t.Fatal(err)
	}
	hidden, err := analysis.Analyze(body)
	if err != nil {
		t.Fatal(err)
	}
	if len(root.Inputs) != 1 || root.Inputs[0].Identity.Value != "view" || len(hidden.Inputs) != 1 || hidden.Inputs[0].Identity.Value != "events_good" {
		t.Fatal("canonical source identities not unique")
	}
	req.Settings.Entries[0].Compatibility.InputBindings = []compatibility.InputBinding{{InputID: root.Inputs[0].ID, ObjectID: view.ID, Expected: identity, SchemaID: "schema-0"}}
	for _, o := range req.Settings.Snapshot.Objects {
		if o.Kind == "dataset" && o.Name == "events_good" {
			req.Settings.Entries[0].Compatibility.InputBindings = append(req.Settings.Entries[0].Compatibility.InputBindings, compatibility.InputBinding{InputID: hidden.Inputs[0].ID, ObjectID: o.ID, Expected: environment.ObjectIdentity{Kind: o.Kind, Name: o.Name, Namespace: o.Namespace, App: o.App, Owner: o.Owner}, SchemaID: "schema-0"})
		}
	}
	r, err := Assess(req)
	if err != nil {
		t.Fatal(err)
	}
	if r.Status != analysis.Valid || r.CIExitCode != 0 || r.Counts.Selected != 1 || r.Counts.Assessed != 1 || r.Entries[0].Compatibility.Closure == nil || len(r.Entries[0].Compatibility.Inputs) != 2 {
		t.Fatalf("definition closure: %+v", r)
	}
	if !reflect.DeepEqual(root.Requirements, r.Entries[0].Compatibility.Requirements) {
		t.Fatal("closure replaced direct requirement set")
	}
}

func TestAssessGlobalPreparationErrorPreservesDetail(t *testing.T) {
	req := seedRequest(t, "from [{id:1}]")
	req.Settings.Snapshot.Objects[0].ID = ""
	req.Documents[0].ID = "" // Global evidence is validated before document linkage.
	_, err := Assess(req)
	detail, ok := RequestErrorDetails(err)
	if !ok || !strings.HasPrefix(detail.Path, "/snapshot/") || !strings.HasPrefix(detail.Code, "snapshot_") {
		t.Fatalf("detail: %+v err=%v", detail, err)
	}
	if _, ok := compatibility.RequestErrorDetails(err); !ok {
		t.Fatal("canonical preparation cause lost")
	}
}

func TestAssessMixedCorpusTransports(t *testing.T) {
	req := seedRequest(t, "from [{id:1}]")
	for _, e := range []struct{ id, text string }{{"bad", "from [{id:1}] | eval broken ="}, {"unknown", "from [{id:1}] | unknowable_command"}} {
		doc := req.Documents[0].Document
		doc.Text = e.text
		doc.SourceID = e.id
		req.Documents = append(req.Documents, corpus.RequestDocument{ID: e.id, Document: doc})
		req.Settings.Entries = append(req.Settings.Entries, EntrySettings{ID: e.id, Compatibility: detach(req.Settings.Entries[0].Compatibility)})
	}
	raw, _ := json.Marshal(req)
	p, err := Prepare(req.Settings)
	if err != nil {
		t.Fatal(err)
	}
	for name, assess := range map[string]func() (*Report, error){"typed": func() (*Report, error) { return Assess(req) }, "json": func() (*Report, error) { return AssessJSON(raw) }, "prepared": func() (*Report, error) { return p.Assess(inlineInput(req)) }} {
		t.Run(name, func(t *testing.T) {
			r, err := assess()
			if err != nil {
				t.Fatal(err)
			}
			if r.CIExitCode != 1 || !r.ExecutionComplete || r.Counts.Selected != 3 || r.Counts.Assessed != 3 || r.Counts.Valid != 1 || r.Counts.Invalid != 1 || r.Counts.Incomplete != 1 {
				t.Fatalf("report: %+v", r)
			}
			for i, e := range req.Documents {
				if r.Entries[i].ID != e.ID {
					t.Fatal("selection order changed")
				}
			}
		})
	}
}
