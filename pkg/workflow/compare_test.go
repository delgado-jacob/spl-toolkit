package workflow

import (
	"bytes"
	"encoding/json"
	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/compatibility"
	"github.com/delgado-jacob/spl-toolkit/pkg/corpus"
	"github.com/delgado-jacob/spl-toolkit/pkg/environment"
	"github.com/delgado-jacob/spl-toolkit/pkg/impact"
	"strings"
	"testing"
)

func TestWorkflowCompareCompleteEquality(t *testing.T) {
	r, err := Assess(seedRequest(t, "from [{id:1}]"))
	if err != nil {
		t.Fatal(err)
	}
	got, err := Compare(CompareRequest{SchemaVersion: 1, Before: *r, After: *r})
	if err != nil {
		t.Fatal(err)
	}
	if got.CIExitCode != 0 || got.Entries[0].Classification != impact.Unchanged {
		t.Fatalf("complete equality: %+v", got)
	}
}

func boundComparisonRequest(t *testing.T, text string) Request {
	t.Helper()
	req := seedRequest(t, text)
	a, err := analysis.Analyze(req.Documents[0].Document)
	if err != nil {
		t.Fatal(err)
	}
	for _, o := range req.Settings.Snapshot.Objects {
		if o.Name == "events_good" {
			req.Settings.Entries[0].Compatibility.InputBindings = []compatibility.InputBinding{{InputID: a.Inputs[0].ID, ObjectID: o.ID, Expected: environment.ObjectIdentity{Kind: o.Kind, Name: o.Name, Namespace: o.Namespace, App: o.App, Owner: o.Owner}, SchemaID: "schema-0"}}
		}
	}
	return req
}
func compareAssessed(t *testing.T, b, a Request) *ComparisonReport {
	t.Helper()
	before, err := Assess(b)
	if err != nil {
		t.Fatal(err)
	}
	after, err := Assess(a)
	if err != nil {
		t.Fatal(err)
	}
	got, err := Compare(CompareRequest{SchemaVersion: 1, Before: *before, After: *after})
	if err != nil {
		t.Fatal(err)
	}
	return got
}
func TestWorkflowCompareRelevantEvidence(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(*Request)
		want   impact.Classification
		code   int
	}{
		{"required schema lost", func(r *Request) {
			r.Settings.SchemaBundle.Schemas[0].Catalog = json.RawMessage(`{"fields":["other"],"identity":"schema-0","optional_fields":[],"version":"1"}`)
		}, impact.Affected, 1},
		{"selected fact same outcome", func(r *Request) { r.Settings.Snapshot.Objects[0].Sharing = "global" }, impact.Affected, 1},
		{"unrelated object", func(r *Request) { r.Settings.Snapshot.Objects[1].Sharing = "global" }, impact.Unchanged, 0},
		{"metadata only", func(r *Request) { r.Settings.Snapshot.Objects[0].Provenance.ObservedAt = "2026-10-06T12:03:00Z" }, impact.Unchanged, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			b := boundComparisonRequest(t, "from events_good | fields id")
			a, _ := exportCopy(b)
			tc.mutate(&a)
			got := compareAssessed(t, b, a)
			if got.Entries[0].Classification != tc.want || got.CIExitCode != tc.code {
				t.Fatalf("classification %s CI%d delta categories %v", got.Entries[0].Classification, got.CIExitCode, deltaKeys(got.Entries[0]))
			}
			if tc.name == "selected fact same outcome" && got.Entries[0].Before.Status != got.Entries[0].After.Status {
				t.Fatal("supporting-fact fixture changed final status")
			}
		})
	}
}
func TestWorkflowComparePartialAndInvalid(t *testing.T) {
	req := boundComparisonRequest(t, "from events_good | fields id")
	for i := range req.Settings.Snapshot.Collections {
		req.Settings.Snapshot.Collections[i].Coverage = "partial"
		req.Settings.Snapshot.Collections[i].Reason = "partial capture"
	}
	for i := range req.Settings.SchemaBundle.Bindings {
		req.Settings.SchemaBundle.Bindings[i].SourceCoverage = "partial"
		req.Settings.SchemaBundle.Bindings[i].Reason = "partial capture"
	}
	got := compareAssessed(t, req, req)
	if got.Entries[0].Classification != impact.Indeterminate || got.CIExitCode != 3 {
		t.Fatalf("partial equality: %+v", got)
	}
	req = boundComparisonRequest(t, "from events_good | fields absent")
	got = compareAssessed(t, req, req)
	if got.Entries[0].Before.Status != analysis.Invalid || got.Entries[0].Classification != impact.Unchanged || got.CIExitCode != 0 {
		t.Fatalf("invalid equality: %+v", got)
	}
}
func TestWorkflowCompareMismatchAndFormatting(t *testing.T) {
	b := seedRequest(t, "from [{id:1}]")
	a := seedRequest(t, "from [{id:2}]")
	got := compareAssessed(t, b, a)
	if got.Entries[0].Classification != impact.Indeterminate || len(got.Entries[0].Deltas) != 0 || got.Entries[0].Reasons[0] != "query_input_mismatch" {
		t.Fatalf("mismatch: %+v", got)
	}
	text := FormatComparison(got)
	for _, want := range []string{"indeterminate", "before: valid", "after: valid", "query_input_mismatch", "Offline evidence comparison"} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %q", want)
		}
	}
}
func TestWorkflowCompareResolutionEnumeration(t *testing.T) {
	b := seedResolutionRequest(t)
	a, _ := exportCopy(b)
	values := a.Settings.Entries[0].Resolution.Resolutions[0].Values
	values[0], values[1] = values[1], values[0]
	got := compareAssessed(t, b, a)
	if got.Entries[0].Classification != impact.Affected {
		t.Fatalf("order: %+v", got)
	}
	for _, d := range got.Entries[0].Deltas {
		if d.Change == "introduced" || d.Change == "resolved" {
			t.Fatalf("fake variant membership: %+v", d)
		}
	}
}

func deltaKeys(e ComparisonEntry) []string {
	out := []string{}
	for _, d := range e.Deltas {
		out = append(out, d.Category+":"+d.Key)
	}
	return out
}

func TestWorkflowCompareHistoricalCanonicalRenumbering(t *testing.T) {
	r, err := Assess(boundComparisonRequest(t, "from events_good | fields id"))
	if err != nil {
		t.Fatal(err)
	}
	q := CompareRequest{SchemaVersion: 1, Before: *r, After: *r}
	q, _ = exportCopy(q)
	ids := map[string]string{}
	a := q.After.Entries[0].Analysis
	for _, v := range a.Stages {
		ids[v.ID] = "historical-" + v.ID
	}
	for _, v := range a.Scopes {
		ids[v.ID] = "historical-" + v.ID
	}
	for _, v := range a.References {
		ids[v.ID] = "historical-" + v.ID
	}
	for _, v := range a.Inputs {
		ids[v.ID] = "historical-" + v.ID
		for _, o := range v.Occurrences {
			ids[o.ID] = "historical-" + o.ID
		}
	}
	for _, v := range a.Requirements.Items {
		ids[v.ID] = "historical-" + v.ID
	}
	raw, _ := json.Marshal(q.After)
	var wire any
	_ = json.Unmarshal(raw, &wire)
	var rename func(any) any
	rename = func(v any) any {
		switch x := v.(type) {
		case map[string]any:
			for k, v := range x {
				if k == "capability_revision" {
					x[k] = "sha256:" + strings.Repeat("a", 64)
				} else {
					x[k] = rename(v)
				}
			}
		case []any:
			for i, v := range x {
				x[i] = rename(v)
			}
		case string:
			if y, ok := ids[x]; ok {
				return y
			}
		}
		return v
	}
	raw, _ = json.Marshal(rename(wire))
	_ = json.Unmarshal(raw, &q.After)
	got, err := Compare(q)
	if err != nil {
		t.Fatal(err)
	}
	if got.CIExitCode != 0 || got.Entries[0].Classification != impact.Unchanged || len(got.Entries[0].Deltas) != 0 {
		t.Fatalf("renumber classification %s CI%d deltas %v unmatched %v", got.Entries[0].Classification, got.CIExitCode, deltaKeys(got.Entries[0]), got.Entries[0].Unmatched)
	}
}
func TestWorkflowCompareProofDecisionAndExactDeltas(t *testing.T) {
	r, err := Assess(seedResolutionRequest(t))
	if err != nil {
		t.Fatal(err)
	}
	q := CompareRequest{SchemaVersion: 1, Before: *r, After: *r}
	q, _ = exportCopy(q)
	// The retained failed variant remains failed; this imported limitation is
	// evidence, never restored engine authority or permission to publish.
	v := &q.After.Entries[0].Resolution.Variants[1]
	v.Proof.Limitations = append(v.Proof.Limitations, analysis.ResolutionLimitation{Code: "historical_limit", Message: "captured proof limit"})
	got, err := Compare(q)
	if err != nil {
		t.Fatal(err)
	}
	if got.Entries[0].Classification != impact.Affected || got.CIExitCode != 1 {
		t.Fatalf("proof classification %s", got.Entries[0].Classification)
	}
	raw, _ := json.Marshal(got)
	var tree any
	_ = json.Unmarshal(raw, &tree)
	for _, d := range got.Entries[0].Deltas {
		actual, _ := json.Marshal(pointerValue(tree, d.Key))
		var saved any
		_ = json.Unmarshal(d.Before, &saved)
		want, _ := json.Marshal(saved)
		if !bytes.Equal(actual, want) {
			t.Fatalf("delta not retained canonical value at %s", d.Key)
		}
	}
	again, err := Compare(q)
	if err != nil {
		t.Fatal(err)
	}
	next, _ := json.Marshal(again)
	if !bytes.Equal(raw, next) {
		t.Fatal("comparison is nondeterministic")
	}
}
func TestWorkflowCompareDefiniteChangeAlongsideGap(t *testing.T) {
	b := boundComparisonRequest(t, "from events_good | fields id")
	a, _ := exportCopy(b)
	for i := range a.Settings.Snapshot.Collections {
		a.Settings.Snapshot.Collections[i].Coverage = "partial"
		a.Settings.Snapshot.Collections[i].Reason = "partial capture"
	}
	a.Settings.SchemaBundle.Schemas[0].Catalog = json.RawMessage(`{"fields":["other"],"identity":"schema-0","optional_fields":[],"version":"1"}`)
	got := compareAssessed(t, b, a)
	if got.Entries[0].Classification != impact.Affected || got.CIExitCode != 1 {
		t.Fatalf("definite change lost to gap: %s CI%d", got.Entries[0].Classification, got.CIExitCode)
	}
}

func TestWorkflowCompareSchemaPropertyNamedMetadata(t *testing.T) {
	b := boundComparisonRequest(t, "from events_good | fields source_id")
	b.Settings.SchemaBundle.Schemas[0].Kind = "json_schema"
	b.Settings.SchemaBundle.Schemas[0].Catalog = nil
	b.Settings.SchemaBundle.Schemas[0].Target = json.RawMessage(`{"kind":"json_schema","identity":"selected","schema":{"type":"object","properties":{"source_id":true},"required":["source_id"],"additionalProperties":false}}`)
	a, _ := exportCopy(b)
	a.Settings.SchemaBundle.Schemas[0].Target = json.RawMessage(`{"kind":"json_schema","identity":"selected","schema":{"type":"object","properties":{"source_id":true},"required":[],"additionalProperties":false}}`)
	got := compareAssessed(t, b, a)
	if got.Entries[0].Classification != impact.Affected || got.Entries[0].Before.Status != got.Entries[0].After.Status {
		t.Fatalf("selected property facts: %s before %s after %s", got.Entries[0].Classification, got.Entries[0].Before.Status, got.Entries[0].After.Status)
	}
}
func TestWorkflowCompareFailedSiblingPrecedesAffected(t *testing.T) {
	b := boundComparisonRequest(t, "from events_good | fields id")
	a, _ := exportCopy(b)
	a.Settings.SchemaBundle.Schemas[0].Catalog = json.RawMessage(`{"fields":["other"],"identity":"schema-0","optional_fields":[],"version":"1"}`)
	before, err := Assess(b)
	if err != nil {
		t.Fatal(err)
	}
	after, err := Assess(a)
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range []*Report{before, after} {
		r.Entries = append(r.Entries, ReportEntry{ID: "failed", Origin: corpus.Origin{Kind: "inline"}, Mode: "compatibility", Status: analysis.Incomplete, Failure: &Failure{Phase: "acquisition", Code: "not_found", Message: "unavailable"}})
		r.Counts = Counts{Selected: 2}
		r.Status = analysis.Valid
		r.ExecutionComplete = true
		finalize(r)
	}
	got, err := Compare(CompareRequest{SchemaVersion: 1, Before: *before, After: *after})
	if err != nil {
		t.Fatal(err)
	}
	if got.ExecutionComplete || got.CIExitCode != 2 || got.Counts.Affected != 1 || got.Counts.Failed != 1 || got.Counts.Compared != 1 {
		t.Fatalf("failure precedence/counts: %+v", got.Counts)
	}
}

func macroComparisonRequest(t *testing.T, body string) Request {
	t.Helper()
	req := seedRequest(t, "`source`")
	req.Documents[0].Document.Language = "spl"
	n := 0
	no := false
	doc := analysis.QueryDocument{Text: body, Language: "spl", SourceID: "macro.spl"}
	req.Settings.Snapshot.Objects = append(req.Settings.Snapshot.Objects, environment.Object{ID: "macro", Kind: "macro", Name: "source", Namespace: "search", App: "resolution", Owner: "nobody", Provenance: req.Settings.Snapshot.Objects[0].Provenance, Arity: &n, EvalBased: &no, Arguments: []string{}, Document: &doc})
	return req
}
func TestWorkflowCompareClosureIncidentalStagePosition(t *testing.T) {
	r, err := Assess(macroComparisonRequest(t, "| makeresults"))
	if err != nil {
		t.Fatal(err)
	}
	q := CompareRequest{SchemaVersion: 1, Before: *r, After: *r}
	q, _ = exportCopy(q)
	q.After.Entries[0].Compatibility.Closure.DirectAnalysis.Stages[0].Position = 99
	got, err := Compare(q)
	if err != nil {
		t.Fatal(err)
	}
	if got.Entries[0].Classification == impact.Affected || len(got.Entries[0].Deltas) != 0 {
		t.Fatalf("incidental positioning: %s %v", got.Entries[0].Classification, deltaKeys(got.Entries[0]))
	}
}
func TestWorkflowCompareUnmatchedDefinitionBodyUncertainty(t *testing.T) {
	got := compareAssessed(t, macroComparisonRequest(t, "| makeresults"), macroComparisonRequest(t, "|  makeresults"))
	if got.Entries[0].Classification != impact.Indeterminate || got.CIExitCode != 3 {
		t.Fatalf("unmatched body: %s CI%d %v", got.Entries[0].Classification, got.CIExitCode, deltaKeys(got.Entries[0]))
	}
}

func definitionComparisonRequest(t *testing.T, body string) Request {
	t.Helper()
	req := boundComparisonRequest(t, "from events_good | fields id")
	req.Documents[0].Document.Text = "from view | fields id"
	doc := analysis.QueryDocument{Text: body, Language: "spl2", SourceID: "view.spl"}
	view := environment.Object{ID: "view", Kind: "dataset", Name: "view", Namespace: "search", App: "resolution", Owner: "nobody", Provenance: req.Settings.Snapshot.Objects[0].Provenance, Document: &doc}
	identity := environment.ObjectIdentity{Kind: view.Kind, Name: view.Name, Namespace: view.Namespace, App: view.App, Owner: view.Owner}
	req.Settings.Snapshot.Objects = append(req.Settings.Snapshot.Objects, view)
	req.Settings.SchemaBundle.Bindings = append(req.Settings.SchemaBundle.Bindings, environment.SchemaBinding{ObjectID: view.ID, SchemaID: "schema-0", Expected: identity, SourceCoverage: "complete"})
	root, err := analysis.Analyze(req.Documents[0].Document)
	if err != nil {
		t.Fatal(err)
	}
	hidden, err := analysis.Analyze(doc)
	if err != nil {
		t.Fatal(err)
	}
	req.Settings.Entries[0].Compatibility.InputBindings[0].InputID = hidden.Inputs[0].ID
	req.Settings.Entries[0].Compatibility.InputBindings = append(req.Settings.Entries[0].Compatibility.InputBindings, compatibility.InputBinding{InputID: root.Inputs[0].ID, ObjectID: view.ID, Expected: identity, SchemaID: "schema-0"})
	return req
}
func TestWorkflowCompareDefinitionUncertaintyPreservesSchemaRegression(t *testing.T) {
	before := definitionComparisonRequest(t, "from events_good | fields id")
	after := definitionComparisonRequest(t, "from  events_good | fields id")
	after.Settings.SchemaBundle.Schemas[0].Catalog = json.RawMessage(`{"fields":["other"],"identity":"schema-0","optional_fields":[],"version":"1"}`)
	got := compareAssessed(t, before, after)
	if got.Entries[0].Classification != impact.Affected || len(got.Entries[0].Unmatched) == 0 {
		t.Fatalf("independent schema regression: %s unmatched %v deltas %v", got.Entries[0].Classification, got.Entries[0].Unmatched, deltaKeys(got.Entries[0]))
	}
}

func TestWorkflowCompareDefinitionUncertaintyPreservesSelectedObjectFact(t *testing.T) {
	before := macroComparisonRequest(t, "| makeresults")
	after := macroComparisonRequest(t, "|  makeresults")
	after.Settings.Snapshot.Objects[len(after.Settings.Snapshot.Objects)-1].Sharing = "global"
	got := compareAssessed(t, before, after)
	if got.Entries[0].Classification != impact.Affected || len(got.Entries[0].Unmatched) == 0 {
		t.Fatalf("selected captured fact: %s unmatched %v deltas %v", got.Entries[0].Classification, got.Entries[0].Unmatched, deltaKeys(got.Entries[0]))
	}
}

func TestWorkflowCompareUnmatchedBodyWithOriginalRequirementRenumbering(t *testing.T) {
	before, err := Assess(macroComparisonRequest(t, "| makeresults"))
	if err != nil {
		t.Fatal(err)
	}
	after, err := Assess(macroComparisonRequest(t, "|  makeresults"))
	if err != nil {
		t.Fatal(err)
	}
	q := CompareRequest{SchemaVersion: 1, Before: *before, After: *after}
	q, _ = exportCopy(q)
	e := &q.After.Entries[0]
	c := e.Compatibility
	ids := map[string]string{}
	for i := range e.Analysis.Requirements.Items {
		r := &e.Analysis.Requirements.Items[i]
		ids[r.ID] = "recorded-" + r.ID
		r.ID = ids[r.ID]
	}
	c.Requirements = e.Analysis.Requirements
	c.Closure.DirectAnalysis.Requirements = e.Analysis.Requirements
	c.Closure.DirectRequirements = e.Analysis.Requirements
	for i := range c.RequirementOutcomes {
		o := &c.RequirementOutcomes[i]
		if o.Query == e.Analysis.Requirements.Query {
			if id, ok := ids[o.RequirementID]; ok {
				o.RequirementID = id
			}
		}
	}
	got, err := Compare(q)
	if err != nil {
		t.Fatal(err)
	}
	if got.Entries[0].Classification != impact.Indeterminate {
		t.Fatalf("recorded requirement IDs became definite: %s %v", got.Entries[0].Classification, deltaKeys(got.Entries[0]))
	}
}
