package workflow

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/compatibility"
	"github.com/delgado-jacob/spl-toolkit/pkg/corpus"
	"github.com/delgado-jacob/spl-toolkit/pkg/resolution"
)

func seedResolutionRequest(t *testing.T) Request {
	t.Helper()
	raw, err := os.ReadFile("../../examples/resolution/request.json")
	if err != nil {
		t.Fatal(err)
	}
	seed, err := resolution.DecodeRequest(raw)
	if err != nil {
		t.Fatal(err)
	}
	return Request{SchemaVersion: 1, Documents: []corpus.RequestDocument{{ID: "r1", Document: seed.Document}}, Settings: Settings{SchemaVersion: 1, Snapshot: seed.Compatibility.Snapshot, SchemaBundle: seed.Compatibility.SchemaBundle, Entries: []EntrySettings{{ID: "r1", Resolution: &ResolveSettings{Resolutions: seed.Resolutions, MaxVariants: seed.MaxVariants, Compatibility: compatibility.ResolutionAssessment{QueryScope: seed.Compatibility.QueryScope, InputBindings: seed.Compatibility.InputBindings, DependencyBindings: seed.Compatibility.DependencyBindings}}}}}}
}

func TestWorkflowResolutionRetainsUnsuccessfulSibling(t *testing.T) {
	req := seedResolutionRequest(t)
	before, _ := json.Marshal(req)
	report, err := Assess(req)
	if err != nil {
		t.Fatal(err)
	}
	e := report.Entries[0]
	if e.Compatibility != nil || e.Resolution == nil || report.CIExitCode != 1 || len(e.Resolution.Variants) != 2 {
		t.Fatalf("report: %+v entry: %+v", report, e)
	}
	if e.Resolution.Variants[0].ResolvedQuery == nil || e.Resolution.Variants[1].ResolvedQuery != nil || e.Analysis.Document.Text != req.Documents[0].Document.Text {
		t.Fatal("publication or original analysis lost")
	}
	after, _ := json.Marshal(req)
	if !reflect.DeepEqual(before, after) {
		t.Fatal("request mutated")
	}
}

func TestWorkflowResolutionNoPlaceholderDefault(t *testing.T) {
	req := seedResolutionRequest(t)
	req.Documents[0].Document.Text = "from [{id:1}]"
	req.Settings.Entries[0].Resolution.Resolutions = []resolution.Resolution{}
	req.Settings.Entries[0].Resolution.Compatibility.InputBindings = []compatibility.ResolutionBinding{}
	r, err := Assess(req)
	if err != nil {
		t.Fatal(err)
	}
	e := r.Entries[0]
	if r.CIExitCode != 0 || e.Resolution == nil || e.Resolution.GeneratedCount != 1 || e.Resolution.Variants[0].ResolvedQuery == nil {
		t.Fatalf("report: %+v entry: %+v", r, e)
	}
}

func TestWorkflowResolutionPreparedIsolation(t *testing.T) {
	req := seedResolutionRequest(t)
	p, err := Prepare(req.Settings)
	if err != nil {
		t.Fatal(err)
	}
	input := inlineInput(req)
	first, err := p.Assess(input)
	if err != nil {
		t.Fatal(err)
	}
	if first.Entries[0].Resolution == nil {
		t.Fatal("missing resolution")
	}
	want, _ := json.Marshal(first)
	first.Entries[0].Resolution.Variants[0].CandidateAnalysis.Document.Text = "mutated"
	first.Entries[0].Resolution.Resolutions[0].Values[0] = "mutated"
	req.Settings.Entries[0].Resolution.Resolutions[0].Values[0] = "mutated"
	req.Settings.Snapshot.ScopeID = "mutated"
	second, err := p.Assess(input)
	if err != nil {
		t.Fatal(err)
	}
	got, _ := json.Marshal(second)
	if !reflect.DeepEqual(want, got) {
		t.Fatal("prepared state shares mutable reports or settings")
	}
	input.Entries[0].Document.Text = "from $events | fields absent"
	input.Entries[0].SourceHash = corpus.SourceHash(input.Entries[0].Document.Text)
	third, err := p.Assess(input)
	if err != nil {
		t.Fatal(err)
	}
	if third.Entries[0].Analysis.Document.Text != input.Entries[0].Document.Text || third.Status != analysis.Invalid || second.Entries[0].Analysis.Document.Text == input.Entries[0].Document.Text {
		t.Fatal("prepared query state leaked")
	}
}

func TestWorkflowResolutionAbsentBindingsRemainIncomplete(t *testing.T) {
	req := seedResolutionRequest(t)
	req.Settings.Entries[0].Resolution.Compatibility.InputBindings = []compatibility.ResolutionBinding{}
	r, err := Assess(req)
	if err != nil {
		t.Fatal(err)
	}
	e := r.Entries[0]
	if r.CIExitCode != 3 || !r.ExecutionComplete || e.Failure != nil || e.Resolution == nil || e.Resolution.Counts.Incomplete != 2 {
		t.Fatalf("report: %+v entry: %+v", r, e)
	}
	for _, v := range e.Resolution.Variants {
		if v.ResolvedQuery != nil {
			t.Fatal("absent binding published a query")
		}
	}
}

func TestWorkflowResolutionSameNameRolesRemainDistinct(t *testing.T) {
	req := seedResolutionRequest(t)
	req.Documents[0].Document.Text = "from $events | join type=inner left=L right=R where L.id=R.id [from events_good]"
	s := req.Settings.Entries[0].Resolution
	s.Resolutions[0].Values = []string{"events_good"}
	s.Compatibility.InputBindings = s.Compatibility.InputBindings[:1]
	original, err := analysis.Analyze(req.Documents[0].Document)
	if err != nil {
		t.Fatal(err)
	}
	var placeholderID, fixedID string
	for _, input := range original.Inputs {
		switch input.Identity.Value {
		case "$events":
			if placeholderID != "" {
				t.Fatal("ambiguous placeholder role")
			}
			placeholderID = input.ID
		case "events_good":
			if fixedID != "" {
				t.Fatal("ambiguous fixed role")
			}
			fixedID = input.ID
		}
	}
	if placeholderID == "" || fixedID == "" || placeholderID == fixedID {
		t.Fatalf("original roles: %+v", original.Inputs)
	}
	s.Compatibility.InputBindings[0].OriginalInputID = placeholderID
	fixed := s.Compatibility.InputBindings[0]
	fixed.OriginalInputID, fixed.ResolvedValue = fixedID, nil
	s.Compatibility.InputBindings = append(s.Compatibility.InputBindings, fixed)
	r, err := Assess(req)
	if err != nil {
		t.Fatal(err)
	}
	if r.CIExitCode != 0 || r.Entries[0].Resolution == nil || r.Entries[0].Resolution.Variants[0].ResolvedQuery == nil {
		t.Fatalf("report: %+v entry: %+v", r, r.Entries[0])
	}
	assessment := r.Entries[0].Resolution.Variants[0].Compatibility
	if len(assessment.Inputs) != 2 || assessment.Inputs[0].OriginalInputID == assessment.Inputs[1].OriginalInputID || assessment.Inputs[0].CandidateInputID != assessment.Inputs[1].CandidateInputID {
		t.Fatalf("original roles lost: %+v", assessment.Inputs)
	}
	candidate := r.Entries[0].Resolution.Variants[0].CandidateAnalysis
	var candidateID string
	for _, input := range candidate.Inputs {
		if input.Identity.Value == "events_good" {
			if candidateID != "" {
				t.Fatal("ambiguous candidate role")
			}
			candidateID = input.ID
		}
	}
	if candidateID == "" || candidateID == placeholderID {
		t.Fatalf("candidate roles: %+v", candidate.Inputs)
	}
	s.Compatibility.InputBindings[0].OriginalInputID = candidateID
	r, err = Assess(req)
	if err != nil {
		t.Fatal(err)
	}
	if r.CIExitCode != 2 || r.Entries[0].Failure == nil || r.Entries[0].Failure.Code != "binding_invalid" || r.Entries[0].Resolution != nil {
		t.Fatalf("candidate role admitted: %+v", r.Entries[0])
	}
}
