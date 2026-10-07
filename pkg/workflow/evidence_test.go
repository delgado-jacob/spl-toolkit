package workflow

import (
	"encoding/json"
	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/compatibility"
	"github.com/delgado-jacob/spl-toolkit/pkg/impact"
	"strings"
	"testing"
)

func TestEvidenceFailureMessageDisclosure(t *testing.T) {
	r, err := Assess(seedRequest(t, "from [{id:1}]"))
	if err != nil {
		t.Fatal(err)
	}
	r.Entries[0].Failure = &Failure{Phase: "configuration", Code: "request_invalid", Message: "PRIVATE_MESSAGE_17"}
	r.Entries[0].Status = analysis.Incomplete
	r.Entries[0].Compatibility = nil
	r.Status = analysis.Incomplete
	r.ExecutionComplete = false
	r.CIExitCode = 2
	r.Counts.Valid = 0
	r.Counts.Incomplete = 1
	r.Counts.ConfigurationFailed = 1
	r.Counts.Assessed = 0
	got, err := Evidence(EvidenceRequest{SchemaVersion: 1, Report: r, Include: []string{}})
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "PRIVATE_MESSAGE_17") {
		t.Fatal("failure message disclosed by default")
	}
}

func TestEvidenceComparisonAdmission(t *testing.T) {
	r, err := Assess(seedRequest(t, "from [{id:1}]"))
	if err != nil {
		t.Fatal(err)
	}
	c, err := Compare(CompareRequest{SchemaVersion: 1, Before: *r, After: *r})
	if err != nil {
		t.Fatal(err)
	}
	got, err := Evidence(EvidenceRequest{SchemaVersion: 1, Comparison: c, Include: []string{}})
	if err != nil {
		t.Fatal(err)
	}
	if got.SourceKind != "comparison" || got.Counts["unchanged"] != 1 || got.SourceCIExitCode != 0 {
		t.Fatalf("unexpected projection %+v", got)
	}
	for _, mutate := range []func(*ComparisonReport){
		func(c *ComparisonReport) { c.Counts.Unchanged++ },
		func(c *ComparisonReport) { c.Entries[0].Classification = "affected" },
		func(c *ComparisonReport) { c.Entries[0].Pairs[0].BeforePointer = "/PRIVATE_POINTER" },
		func(c *ComparisonReport) {
			c.Entries[0].Deltas = append(c.Entries[0].Deltas, impact.EvidenceDelta{Category: "coverage", Change: "introduced", Key: "/PRIVATE_POINTER", After: json.RawMessage(`{"PRIVATE_DELTA":true}`)})
		},
		func(c *ComparisonReport) { c.Entries[0].Before.SourceHash = "PRIVATE_HASH" },
	} {
		copy, err := exportCopy(*c)
		if err != nil {
			t.Fatal(err)
		}
		mutate(&copy)
		if _, err := Evidence(EvidenceRequest{SchemaVersion: 1, Comparison: &copy, Include: []string{}}); err == nil {
			t.Fatal("inconsistent comparison admitted")
		} else if strings.Contains(err.Error(), "PRIVATE") {
			t.Fatal("rejection leaked input")
		}
	}
	c.ExecutionComplete = false
	c.CIExitCode = 2
	got, err = Evidence(EvidenceRequest{SchemaVersion: 1, Comparison: c, Include: []string{}})
	if err != nil {
		t.Fatal(err)
	}
	if got.ExecutionComplete || got.SourceCIExitCode != 2 {
		t.Fatal("lost incomplete execution")
	}
}

func TestEvidenceStrictAdmission(t *testing.T) {
	r, err := Assess(seedRequest(t, "from [{id:1}]"))
	if err != nil {
		t.Fatal(err)
	}
	q := EvidenceRequest{SchemaVersion: 1, Report: r, Include: []string{}}
	raw, _ := json.Marshal(q)
	if _, err := EvidenceJSON(raw); err != nil {
		t.Fatal(err)
	}
	for _, invalid := range []string{
		strings.Replace(string(raw), `"include":[]`, `"include":null`, 1),
		strings.Replace(string(raw), `"include":[]`, `"include":["query_text","query_text"]`, 1),
		strings.Replace(string(raw), `"include":[]`, `"include":["PRIVATE_CATEGORY"]`, 1),
		strings.Replace(string(raw), `"include":[]`, `"include":[],"PRIVATE_PROPERTY":true`, 1),
		strings.Replace(string(raw), `"include":[]`, `"include":[],"PRIVATE_DUPLICATE":1,"PRIVATE_DUPLICATE":2`, 1),
		strings.Replace(string(raw), `"include":[]`, `"include":[],"comparison":null`, 1),
		string(raw) + ` {"PRIVATE_TRAILING":true}`,
		strings.Replace(string(raw), `"schema_version":1`, `"schema_version":1,"schema_version":1`, 1),
		strings.Replace(string(raw), `"source_id":"d1"`, `"source_id":"\ud800"`, 1),
	} {
		_, err := EvidenceJSON([]byte(invalid))
		if err == nil {
			t.Fatal("invalid JSON admitted")
		}
		if strings.Contains(err.Error(), "PRIVATE") {
			t.Fatalf("error leaked input: %v", err)
		}
	}
	q.Include = nil
	if _, err := Evidence(q); err == nil {
		t.Fatal("missing include admitted")
	}
	q.Include = []string{}
	q.Report.Entries[0].ID = string([]byte{0xff})
	if _, err := Evidence(q); err == nil {
		t.Fatal("invalid typed UTF-8 admitted")
	}
}

func TestEvidencePositivePartialCompleteness(t *testing.T) {
	req := boundComparisonRequest(t, "from events_good | fields id")
	for i := range req.Settings.Snapshot.Collections {
		req.Settings.Snapshot.Collections[i].Coverage = "partial"
		req.Settings.Snapshot.Collections[i].Reason = "partial capture"
	}
	for i := range req.Settings.SchemaBundle.Bindings {
		req.Settings.SchemaBundle.Bindings[i].SourceCoverage = "partial"
		req.Settings.SchemaBundle.Bindings[i].Reason = "partial capture"
	}
	r, err := Assess(req)
	if err != nil {
		t.Fatal(err)
	}
	if r.Entries[0].Compatibility.Outcome != "satisfied" {
		t.Fatal("fixture is not positive partial")
	}
	got, _ := evidenceBytes(t, EvidenceRequest{SchemaVersion: 1, Report: r, Include: []string{}})
	seen := false
	for _, item := range got.Items {
		if item.Kind == "compatibility" || item.Kind == "input_assessment" || item.Kind == "requirement_assessment" {
			seen = true
			if item.Complete {
				t.Fatalf("partial evidence reported complete: %+v", item)
			}
		}
	}
	if !seen {
		t.Fatal("missing compatibility evidence")
	}
}

func TestEvidenceHistoricalRevisionAndUnknownSelectors(t *testing.T) {
	r, err := Assess(seedRequest(t, "from [{id:1}]"))
	if err != nil {
		t.Fatal(err)
	}
	historical := "sha256:" + strings.Repeat("a", 64)
	rawSaved, _ := json.Marshal(r)
	rawSaved = []byte(strings.ReplaceAll(string(rawSaved), r.Entries[0].Analysis.Requirements.CapabilityRevision, historical))
	rawSaved = []byte(strings.ReplaceAll(string(rawSaved), "spl2", "PRIVATE_LANGUAGE"))
	if err := json.Unmarshal(rawSaved, r); err != nil {
		t.Fatal(err)
	}
	got, raw := evidenceBytes(t, EvidenceRequest{SchemaVersion: 1, Report: r, Include: []string{}})
	if strings.Contains(raw, "PRIVATE_LANGUAGE") {
		t.Fatal("unknown selector disclosed")
	}
	found := false
	for _, item := range got.Items {
		if item.Capability != nil {
			found = true
			if item.Capability.Language != "unrecognized" || item.Capability.Revision != historical {
				t.Fatalf("historical capability not preserved %+v", item.Capability)
			}
		}
	}
	if !found {
		t.Fatal("no capability")
	}
}

func TestEvidenceCoverageAndUnknownClosureCompleteness(t *testing.T) {
	for _, tc := range []struct {
		name    string
		request func(*testing.T) Request
	}{
		{"literal", func(t *testing.T) Request { return seedRequest(t, "from [{id:1}]") }},
		{"bound", func(t *testing.T) Request { return boundComparisonRequest(t, "from events_good | fields id") }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r, err := Assess(tc.request(t))
			if err != nil {
				t.Fatal(err)
			}
			got, _ := evidenceBytes(t, EvidenceRequest{SchemaVersion: 1, Report: r, Include: []string{}})
			for _, item := range got.Items {
				if item.Kind == "detection" || item.Kind == "compatibility" || item.Kind == "closure" {
					if !item.Complete {
						t.Errorf("canonical captured evidence incomplete: %+v", item)
					}
				}
				for _, coverage := range item.Coverage {
					if coverage.Dimension == "unrecognized" {
						t.Errorf("canonical dimension unrecognized: %+v", item)
					}
				}
			}
			empty, err := exportCopy(*r)
			if err != nil {
				t.Fatal(err)
			}
			empty.Entries[0].Compatibility.Coverage = []compatibility.Coverage{}
			got, _ = evidenceBytes(t, EvidenceRequest{SchemaVersion: 1, Report: &empty, Include: []string{}})
			for _, item := range got.Items {
				if (item.Kind == "detection" || item.Kind == "compatibility") && item.Complete {
					t.Errorf("missing obligations marked complete: %+v", item)
				}
			}
			unknown, err := exportCopy(*r)
			if err != nil {
				t.Fatal(err)
			}
			unknown.Entries[0].Compatibility.Closure.Status = "PRIVATE_CLOSURE_STATUS_17"
			got, raw := evidenceBytes(t, EvidenceRequest{SchemaVersion: 1, Report: &unknown, Include: []string{}})
			if strings.Contains(raw, "PRIVATE_CLOSURE_STATUS_17") {
				t.Fatal("unknown closure label disclosed")
			}
			for _, item := range got.Items {
				if (item.Kind == "detection" || item.Kind == "compatibility" || item.Kind == "closure") && item.Complete {
					t.Errorf("unknown closure marked complete: %+v", item)
				}
			}
		})
	}
}

func TestEvidenceUnknownInputKindPropagatesCompleteness(t *testing.T) {
	for _, tc := range []struct {
		name    string
		request Request
	}{
		{"bound", boundComparisonRequest(t, "from events_good | fields id")},
		{"supported closure", definitionComparisonRequest(t, "from events_good | fields id")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r, err := Assess(tc.request)
			if err != nil {
				t.Fatal(err)
			}
			got, _ := evidenceBytes(t, EvidenceRequest{SchemaVersion: 1, Report: r, Include: []string{}})
			for _, item := range got.Items {
				if item.Kind == "detection" || item.Kind == "compatibility" || item.Kind == "closure" {
					if !item.Complete {
						t.Fatalf("canonical closure discharge lost: %+v", item)
					}
				}
			}
			saved, _ := json.Marshal(r)
			for _, mutation := range []struct{ old, replacement string }{
				{"explicit_dataset", "PRIVATE_UNKNOWN_KIND_17"},
				{`"kind":"field"`, `"kind":"PRIVATE_UNKNOWN_FIELD_KIND_17"`},
				{`"resolution":"exact"`, `"resolution":"PRIVATE_UNKNOWN_RESOLUTION_17"`},
				{`"state":"complete"`, `"state":"PRIVATE_UNKNOWN_STATE_17"`},
			} {
				changed := strings.ReplaceAll(string(saved), mutation.old, mutation.replacement)
				if changed == string(saved) {
					t.Fatal("fixture has no target label")
				}
				var imported Report
				if err := json.Unmarshal([]byte(changed), &imported); err != nil {
					t.Fatal(err)
				}
				got, raw := evidenceBytes(t, EvidenceRequest{SchemaVersion: 1, Report: &imported, Include: []string{}})
				if strings.Contains(raw, "PRIVATE_UNKNOWN_") {
					t.Fatal("unknown label leaked")
				}
				for _, item := range got.Items {
					if (item.Kind == "detection" || item.Kind == "compatibility" || item.Kind == "closure") && item.Complete {
						t.Errorf("unknown label %s did not propagate: %+v", mutation.old, item)
					}
				}
			}
		})
	}
}

func TestEvidenceUnknownAssessmentLabelsPropagate(t *testing.T) {
	r, err := Assess(boundComparisonRequest(t, "from events_good | fields id"))
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name   string
		mutate func(*Report)
	}{
		{"input outcome", func(r *Report) { r.Entries[0].Compatibility.Inputs[0].Outcome = "PRIVATE_UNKNOWN_OUTCOME" }},
		{"requirement outcome", func(r *Report) { r.Entries[0].Compatibility.RequirementOutcomes[0].Outcome = "PRIVATE_UNKNOWN_OUTCOME" }},
		{"requirement applicability", func(r *Report) {
			r.Entries[0].Compatibility.RequirementOutcomes[0].Applicability = "PRIVATE_UNKNOWN_APPLICABILITY"
		}},
		{"schema projection outcome", func(r *Report) {
			for i := range r.Entries[0].Compatibility.RequirementOutcomes {
				outcome := &r.Entries[0].Compatibility.RequirementOutcomes[i]
				if outcome.FieldProjection != nil {
					outcome.FieldProjection.Outcome = "PRIVATE_UNKNOWN_OUTCOME"
					return
				}
			}
			t.Fatal("fixture has no field projection")
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			imported, err := exportCopy(*r)
			if err != nil {
				t.Fatal(err)
			}
			tc.mutate(&imported)
			got, raw := evidenceBytes(t, EvidenceRequest{SchemaVersion: 1, Report: &imported, Include: []string{}})
			if strings.Contains(raw, "PRIVATE_UNKNOWN") {
				t.Fatal("unknown assessment label disclosed")
			}
			for _, item := range got.Items {
				if (item.Kind == "detection" || item.Kind == "compatibility") && item.Complete {
					t.Errorf("unknown child assessment label did not propagate: %+v", item)
				}
			}
			c, err := Compare(CompareRequest{SchemaVersion: 1, Before: imported, After: imported})
			if err != nil {
				t.Fatal(err)
			}
			got, _ = evidenceBytes(t, EvidenceRequest{SchemaVersion: 1, Comparison: c, Include: []string{}})
			for _, item := range got.Items {
				if item.Kind == "comparison" && item.Complete {
					t.Errorf("comparison complete with unknown assessment: %+v", item)
				}
			}
		})
	}
	rr, err := Assess(seedResolutionRequest(t))
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name   string
		mutate func(*compatibility.ResolutionReport)
	}{
		{"wrapped input outcome", func(c *compatibility.ResolutionReport) { c.Inputs[0].Evidence.Outcome = "PRIVATE_UNKNOWN_OUTCOME" }},
		{"wrapped requirement outcome", func(c *compatibility.ResolutionReport) {
			c.RequirementOutcomes[0].Evidence.Outcome = "PRIVATE_UNKNOWN_OUTCOME"
		}},
		{"wrapped requirement applicability", func(c *compatibility.ResolutionReport) {
			c.RequirementOutcomes[0].Evidence.Applicability = "PRIVATE_UNKNOWN_APPLICABILITY"
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			imported, err := exportCopy(*rr)
			if err != nil {
				t.Fatal(err)
			}
			tc.mutate(imported.Entries[0].Resolution.Variants[0].Compatibility)
			got, _ := evidenceBytes(t, EvidenceRequest{SchemaVersion: 1, Report: &imported, Include: []string{}})
			for _, item := range got.Items {
				if (item.Kind == "detection" || (item.Kind == "variant" && strings.HasSuffix(item.Pointer, "/variants/0")) || (item.Kind == "compatibility" && strings.Contains(item.Pointer, "/variants/0/"))) && item.Complete {
					t.Errorf("wrapped unknown child assessment label did not propagate: %+v", item)
				}
			}
			c, err := Compare(CompareRequest{SchemaVersion: 1, Before: imported, After: imported})
			if err != nil {
				t.Fatal(err)
			}
			got, _ = evidenceBytes(t, EvidenceRequest{SchemaVersion: 1, Comparison: c, Include: []string{}})
			for _, item := range got.Items {
				if item.Kind == "comparison" && item.Complete {
					t.Errorf("comparison complete with wrapped unknown assessment: %+v", item)
				}
			}

		})
	}
}

func TestEvidenceKnownAssessmentOutcomesRemainComplete(t *testing.T) {
	for _, req := range []Request{seedRequest(t, "from [{id:1}]"), boundComparisonRequest(t, "from events_good | fields id"), boundComparisonRequest(t, "from events_good | fields absent"), seedResolutionRequest(t)} {
		r, err := Assess(req)
		if err != nil {
			t.Fatal(err)
		}
		got, _ := evidenceBytes(t, EvidenceRequest{SchemaVersion: 1, Report: r, Include: []string{}})
		for _, item := range got.Items {
			if item.Kind == "detection" || item.Kind == "compatibility" || item.Kind == "variant" {
				if !item.Complete {
					t.Fatalf("canonical terminal evidence incomplete: %+v", item)
				}
			}
			if item.Kind == "requirement_assessment" && item.Outcome == "missing" && !item.Complete {
				t.Fatalf("known negative finding incomplete: %+v", item)
			}
		}
	}
	r, err := Assess(seedRequest(t, "from [{id:1}]"))
	if err != nil {
		t.Fatal(err)
	}
	for _, outcome := range []string{"neutral", "indeterminate"} {
		imported, err := exportCopy(*r)
		if err != nil {
			t.Fatal(err)
		}
		imported.Entries[0].Compatibility.RequirementOutcomes[0].Outcome = outcome
		imported.Entries[0].Compatibility.RequirementOutcomes[0].Applicability = "inapplicable"
		got, _ := evidenceBytes(t, EvidenceRequest{SchemaVersion: 1, Report: &imported, Include: []string{}})
		for _, item := range got.Items {
			if (item.Kind == "detection" || item.Kind == "compatibility" || item.Kind == "requirement_assessment") && !item.Complete {
				t.Fatalf("known inapplicable obligation incomplete: %+v", item)
			}
		}
	}
}
