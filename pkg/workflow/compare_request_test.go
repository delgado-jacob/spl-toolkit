package workflow

import (
	"encoding/json"
	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/corpus"
	"github.com/delgado-jacob/spl-toolkit/pkg/environment"
	"testing"
)

func comparisonSeed(t *testing.T) CompareRequest {
	t.Helper()
	r, err := Assess(seedRequest(t, "from main | where host=\"a\""))
	if err != nil {
		t.Fatal(err)
	}
	return CompareRequest{SchemaVersion: 1, Before: *r, After: *r}
}
func TestCompareAdmission(t *testing.T) {
	q := comparisonSeed(t)
	if _, err := Compare(q); err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name   string
		mutate func(*CompareRequest)
	}{
		{"hash", func(q *CompareRequest) { q.After.Entries[0].SourceHash = "wrong" }},
		{"counts", func(q *CompareRequest) { q.After.Counts.Assessed++ }},
		{"mode", func(q *CompareRequest) { q.After.Entries[0].Mode = "resolution" }},
		{"array", func(q *CompareRequest) { q.After.Entries = nil }},
		{"location", func(q *CompareRequest) { q.After.Entries[0].Analysis.References[0].Location.Start.Offset = -1 }},
		{"reference", func(q *CompareRequest) { q.After.Entries[0].Analysis.Inputs[0].Occurrences[0].ReferenceID = "missing" }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			q := comparisonSeed(t)
			q, _ = exportCopy(q)
			tc.mutate(&q)
			if _, err := Compare(q); err == nil {
				t.Fatal("accepted malformed saved evidence")
			}
		})
	}
	raw, _ := json.Marshal(comparisonSeed(t))
	raw = append([]byte(`{"schema_version":1,`), raw[1:]...)
	if _, err := CompareJSON(raw); err == nil {
		t.Fatal("accepted duplicate member")
	}
}

func TestCompareAdmissionCanonicalRoundtrip(t *testing.T) {
	for _, text := range []string{"from [{id:1}]", "from main | where host=\"a\"", "from [{id:1}] | eval broken =", "from $events | fields host"} {
		t.Run(text, func(t *testing.T) {
			r, err := Assess(seedRequest(t, text))
			if err != nil {
				t.Fatal(err)
			}
			raw, _ := json.Marshal(CompareRequest{SchemaVersion: 1, Before: *r, After: *r})
			if _, err = CompareJSON(raw); err != nil {
				t.Fatal(err)
			}
		})
	}
	for _, failure := range []bool{false, true} {
		req := seedResolutionRequest(t)
		if failure {
			n := uint64(1)
			req.Settings.Entries[0].Resolution.MaxVariants = &n
		}
		r, err := Assess(req)
		if err != nil {
			t.Fatal(err)
		}
		raw, _ := json.Marshal(CompareRequest{SchemaVersion: 1, Before: *r, After: *r})
		if _, err = CompareJSON(raw); err != nil {
			t.Fatal(err)
		}
	}
}

func TestCompareAdmissionClosureRoundtrip(t *testing.T) {
	req := seedRequest(t, "`source`")
	req.Documents[0].Document.Language = "spl"
	n := 0
	no := false
	doc := analysis.QueryDocument{Text: "| makeresults", Language: "spl", SourceID: "macro.spl"}
	for i := range req.Settings.Snapshot.Collections {
		if req.Settings.Snapshot.Collections[i].Kind == "macro" {
			req.Settings.Snapshot.Collections[i].Coverage = "complete"
		}
	}
	req.Settings.Snapshot.Objects = append(req.Settings.Snapshot.Objects, environment.Object{ID: "macro", Kind: "macro", Name: "source", Namespace: "search", App: "resolution", Owner: "nobody", Provenance: req.Settings.Snapshot.Objects[0].Provenance, Arity: &n, EvalBased: &no, Arguments: []string{}, Document: &doc})
	r, err := Assess(req)
	if err != nil {
		t.Fatal(err)
	}
	if r.Entries[0].Compatibility == nil || r.Entries[0].Compatibility.Closure == nil {
		t.Fatalf("no closure: %+v", r.Entries[0])
	}
	raw, _ := json.Marshal(CompareRequest{SchemaVersion: 1, Before: *r, After: *r})
	if _, err = CompareJSON(raw); err != nil {
		t.Fatal(err)
	}
}

func TestCompareAdmissionSavedFailureAndSelection(t *testing.T) {
	req := seedRequest(t, "from [{id:1}]")
	p, err := Prepare(req.Settings)
	if err != nil {
		t.Fatal(err)
	}
	in := inlineInput(req)
	in.Entries[0].Document = nil
	in.Entries[0].SourceHash = ""
	in.Entries[0].Failure = &corpus.AcquisitionError{ID: "d1", Code: "not_found", Phase: "open", Message: "missing"}
	r, err := p.Assess(in)
	if err != nil {
		t.Fatal(err)
	}
	q := CompareRequest{SchemaVersion: 1, Before: *r, After: *r}
	got, err := Compare(q)
	if err != nil {
		t.Fatal(err)
	}
	if got.Counts.Failed != 1 || got.CIExitCode != 2 || got.ExecutionComplete {
		t.Fatalf("failure lost: %+v", got)
	}
	q.After, _ = exportCopy(q.After)
	q.After.Entries[0].ID = "other"
	if _, err = Compare(q); err == nil {
		t.Fatal("accepted changed selected ID set")
	}
}
func TestCompareAdmissionStrictSavedWire(t *testing.T) {
	raw, _ := json.Marshal(comparisonSeed(t))
	for _, name := range []string{"missing_array", "null_array", "optional_null", "fractional_integer", "unknown", "duplicate_id"} {
		t.Run(name, func(t *testing.T) {
			var v map[string]any
			json.Unmarshal(raw, &v)
			after := v["after"].(map[string]any)
			entries := after["entries"].([]any)
			entry := entries[0].(map[string]any)
			switch name {
			case "missing_array":
				delete(after, "entries")
			case "null_array":
				after["entries"] = nil
			case "optional_null":
				entry["failure"] = nil
			case "fractional_integer":
				after["schema_version"] = 1.5
			case "unknown":
				after["unknown"] = true
			case "duplicate_id":
				after["entries"] = append(entries, entry)
			}
			b, _ := json.Marshal(v)
			if _, err := CompareJSON(b); err == nil {
				t.Fatal("accepted malformed saved JSON")
			}
		})
	}
}
