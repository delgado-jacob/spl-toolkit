package workflow

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/compatibility"
	"github.com/delgado-jacob/spl-toolkit/pkg/corpus"
	"github.com/delgado-jacob/spl-toolkit/pkg/environment"
)

func TestRecheckChangedQueryUsesFreshAnalysis(t *testing.T) {
	req := seedRequest(t, "from [{id:1}]")
	candidate := req.Documents[0].Document
	candidate.Text = "from [{id:1}] | eval broken ="
	report, err := Recheck(RecheckRequest{SchemaVersion: 1, Context: RecheckContext{Original: req.Documents[0], Snapshot: req.Settings.Snapshot, SchemaBundle: req.Settings.SchemaBundle}, Proposal: Proposal{Document: &candidate, Settings: req.Settings.Entries[0]}})
	if err != nil {
		t.Fatal(err)
	}
	if report.Assessment.CIExitCode != 1 || report.Assessment.Entries[0].Analysis.Document.Text != candidate.Text {
		t.Fatalf("fresh candidate assessment: %+v", report)
	}
}

func queryRecheck(t *testing.T, text string) RecheckRequest {
	t.Helper()
	req := seedRequest(t, "from [{id:1}]")
	candidate := req.Documents[0].Document
	candidate.Text, candidate.SourceID = text, "opaque-caller-token"
	return RecheckRequest{SchemaVersion: 1, Context: RecheckContext{Original: req.Documents[0], Snapshot: req.Settings.Snapshot, SchemaBundle: req.Settings.SchemaBundle}, Proposal: Proposal{Document: &candidate, Settings: req.Settings.Entries[0]}}
}
func resolutionRecheck(t *testing.T) RecheckRequest {
	t.Helper()
	req := seedResolutionRequest(t)
	return RecheckRequest{SchemaVersion: 1, Context: RecheckContext{Original: req.Documents[0], Snapshot: req.Settings.Snapshot, SchemaBundle: req.Settings.SchemaBundle}, Proposal: Proposal{Settings: req.Settings.Entries[0]}}
}

func TestRecheckOutcomesAndExactLinks(t *testing.T) {
	for _, test := range []struct {
		name, text string
		code       int
		status     analysis.Status
	}{
		{"valid", "from [{id:1}] | fields id", 0, analysis.Valid},
		{"unresolved query", "from [{id:1}] | unknowable_command", 3, analysis.Incomplete},
		{"malformed", "from [{id:1}] | eval broken =", 1, analysis.Invalid},
	} {
		t.Run(test.name, func(t *testing.T) {
			q := queryRecheck(t, test.text)
			before, _ := json.Marshal(q)
			r, err := Recheck(q)
			if err != nil {
				t.Fatal(err)
			}
			if r.Assessment.CIExitCode != test.code || r.Assessment.Status != test.status || !r.Assessment.ExecutionComplete || r.Assessment.Entries[0].Failure != nil {
				t.Fatalf("outcome: %+v", r.Assessment)
			}
			if r.DetectionID != q.Context.Original.ID || r.OriginalSourceHash != corpus.SourceHash(q.Context.Original.Document.Text) || r.ProposedSourceHash != corpus.SourceHash(test.text) || r.Assessment.Entries[0].SourceHash != r.ProposedSourceHash {
				t.Fatal("lost exact source links")
			}
			after, _ := json.Marshal(q)
			if !bytes.Equal(before, after) {
				t.Fatal("mutated caller context/configuration")
			}
			raw, _ := json.Marshal(q)
			wire, err := RecheckJSON(raw)
			if err != nil {
				t.Fatal(err)
			}
			want, _ := json.Marshal(r)
			got, _ := json.Marshal(wire)
			if !bytes.Equal(want, got) {
				t.Fatal("typed and wire assessment differ")
			}
		})
	}
}

func TestRecheckFreshProofAndDetachedReports(t *testing.T) {
	q := queryRecheck(t, "from [{id:1}]")
	first, err := Recheck(q)
	if err != nil {
		t.Fatal(err)
	}
	want, _ := json.Marshal(first)
	first.Assessment.Entries[0].Analysis.Document.Text = "mutated"
	first.Assessment.Entries[0].Compatibility.QueryScope.Namespace.Values = []string{"mutated"}
	first.Assessment.CIExitCode = 0
	second, err := Recheck(q)
	if err != nil {
		t.Fatal(err)
	}
	got, _ := json.Marshal(second)
	if !bytes.Equal(want, got) {
		t.Fatal("earlier report mutated fresh outcome")
	}
	q.Proposal.Document.Text = "from [{id:1}] | eval broken ="
	third, err := Recheck(q)
	if err != nil {
		t.Fatal(err)
	}
	if third.Assessment.CIExitCode != 1 || third.Assessment.Entries[0].Analysis.Document.SourceID != "opaque-caller-token" || second.Assessment.CIExitCode != 0 {
		t.Fatal("opaque caller token cached/authorized proof")
	}
}

func TestRecheckResolutionPublicationAndIsolation(t *testing.T) {
	for _, name := range []string{"mixed", "valid", "incomplete", "over limit", "bad binding", "namespace"} {
		t.Run(name, func(t *testing.T) {
			q := resolutionRecheck(t)
			settings := q.Proposal.Settings.Resolution
			expected := 1
			switch name {
			case "valid":
				settings.Resolutions[0].Values = settings.Resolutions[0].Values[:1]
				settings.Compatibility.InputBindings = settings.Compatibility.InputBindings[:1]
				expected = 0
			case "incomplete":
				settings.Compatibility.InputBindings = []compatibility.ResolutionBinding{}
				expected = 3
			case "over limit":
				n := uint64(1)
				settings.MaxVariants = &n
				expected = 2
			case "bad binding":
				settings.Compatibility.InputBindings[0].OriginalInputID = "opaque-evidence-token"
				expected = 2
			case "namespace":
				expected = 2
				settings.Compatibility.QueryScope.Namespace = environment.Selector{Values: []string{"unobserved"}}
			}
			before, _ := json.Marshal(q)
			r, err := Recheck(q)
			if err != nil {
				t.Fatal(err)
			}
			if r.Assessment.CIExitCode != expected {
				t.Fatalf("outcome: %+v", r.Assessment)
			}
			e := r.Assessment.Entries[0]
			if e.Analysis.Document.Text != q.Context.Original.Document.Text || r.ProposedSourceHash != r.OriginalSourceHash {
				t.Fatal("resolution changed original query")
			}
			if name == "bad binding" || name == "over limit" || name == "namespace" {
				if e.Failure == nil || e.Resolution != nil {
					t.Fatal("request failure not retained")
				}
			} else {
				if e.Failure != nil || e.Resolution == nil {
					t.Fatal("canonical resolution missing")
				}
				for _, v := range e.Resolution.Variants {
					if v.Outcome != "verified" && v.ResolvedQuery != nil {
						t.Fatal("unsuccessful variant published resolved query")
					}
				}
				if name == "mixed" && (e.Resolution.Variants[0].ResolvedQuery == nil || e.Resolution.Variants[1].ResolvedQuery != nil) {
					t.Fatal("mixed siblings lost")
				}
			}
			after, _ := json.Marshal(q)
			if !bytes.Equal(before, after) {
				t.Fatal("context artifacts or settings changed")
			}
			if e.Resolution != nil {
				e.Resolution.Variants[0].CandidateAnalysis.Document.Text = "tampered"
			}
			fresh, err := Recheck(q)
			if err != nil {
				t.Fatal(err)
			}
			if fresh.Assessment.CIExitCode != expected || fresh.Assessment.Entries[0].Analysis.Document.Text != q.Context.Original.Document.Text {
				t.Fatal("returned proof leaked into next session")
			}
		})
	}
}

func TestRecheckRejectsModeAndTypedAdmission(t *testing.T) {
	for name, mutate := range map[string]func(*RecheckRequest){
		"version":         func(q *RecheckRequest) { q.SchemaVersion = 2 },
		"id mismatch":     func(q *RecheckRequest) { q.Proposal.Settings.ID = "other" },
		"missing query":   func(q *RecheckRequest) { q.Proposal.Document = nil },
		"neither mode":    func(q *RecheckRequest) { q.Proposal.Settings.Compatibility = nil },
		"typed utf8":      func(q *RecheckRequest) { q.Context.Original.Document.Text = string([]byte{0xff}) },
		"proposal utf8":   func(q *RecheckRequest) { q.Proposal.Document.Text = string([]byte{0xff}) },
		"null bindings":   func(q *RecheckRequest) { q.Proposal.Settings.Compatibility.InputBindings = nil },
		"global artifact": func(q *RecheckRequest) { q.Context.Snapshot.Capabilities = nil },
	} {
		t.Run(name, func(t *testing.T) {
			q := queryRecheck(t, "from [{id:1}]")
			mutate(&q)
			r, err := Recheck(q)
			if err == nil || r != nil {
				t.Fatal("accepted invalid typed request")
			}
		})
	}
	q := resolutionRecheck(t)
	q.Proposal.Document = &q.Context.Original.Document
	if _, err := Recheck(q); err == nil {
		t.Fatal("accepted resolution query replacement")
	}
	q = queryRecheck(t, "from [{id:1}]")
	q.Proposal.Settings.Resolution = resolutionRecheck(t).Proposal.Settings.Resolution
	if _, err := Recheck(q); err == nil {
		t.Fatal("accepted both modes")
	}
}

func TestRecheckStrictWireAdmission(t *testing.T) {
	raw, _ := json.Marshal(queryRecheck(t, "from [{id:1}]"))
	tests := map[string][]byte{
		"duplicate":     append([]byte(`{"schema_version":1,`), raw[1:]...),
		"trailing":      append(append([]byte{}, raw...), []byte(` {}`)...),
		"surrogate":     bytes.Replace(raw, []byte(`opaque-caller-token`), []byte(`\ud800`), 1),
		"utf8":          bytes.Replace(raw, []byte(`opaque-caller-token`), []byte{0xff}, 1),
		"fraction":      bytes.Replace(raw, []byte(`"schema_version":1`), []byte(`"schema_version":1.0`), 1),
		"missing array": bytes.Replace(raw, []byte(`,"input_bindings":[]`), nil, 1),
	}
	var nullDocument map[string]any
	if err := json.Unmarshal(raw, &nullDocument); err != nil {
		t.Fatal(err)
	}
	nullDocument["proposal"].(map[string]any)["document"] = nil
	tests["null document"], _ = json.Marshal(nullDocument)
	for _, field := range []string{"success", "proof", "report", "candidate_patch", "evidence_token", "snapshot", "schema_bundle"} {
		tests["proposal "+field] = bytes.Replace(raw, []byte(`"proposal":{`), []byte(`"proposal":{"`+field+`":{},`), 1)
	}
	for name, bad := range tests {
		t.Run(name, func(t *testing.T) {
			if r, err := RecheckJSON(bad); err == nil || r != nil {
				t.Fatal("accepted invalid wire proposal")
			}
		})
	}
	// Artifact duplicate diagnostics retain the context owner path and exact offset.
	bad := bytes.Replace(raw, []byte(`"snapshot":{`), []byte(`"snapshot":{"scope_id":"duplicate",`), 1)
	_, err := RecheckJSON(bad)
	detail, ok := RequestErrorDetails(err)
	if !ok || detail.Path != "/context/snapshot/scope_id" || detail.ByteOffset == nil {
		t.Fatalf("artifact detail: %+v %v", detail, err)
	}
	if *detail.ByteOffset <= bytes.Index(bad, []byte(`"snapshot":`)) {
		t.Fatal("artifact offset not rebased")
	}
}

func TestRecheckSameNameRolesRemainExplicit(t *testing.T) {
	q := resolutionRecheck(t)
	q.Context.Original.Document.Text = "from $events | join type=inner left=L right=R where L.id=R.id [from events_good]"
	s := q.Proposal.Settings.Resolution
	s.Resolutions[0].Values = []string{"events_good"}
	s.Compatibility.InputBindings = s.Compatibility.InputBindings[:1]
	a, err := analysis.Analyze(q.Context.Original.Document)
	if err != nil {
		t.Fatal(err)
	}
	var placeholderID, fixedID string
	for _, i := range a.Inputs {
		if i.Identity.Value == "$events" {
			placeholderID = i.ID
		}
		if i.Identity.Value == "events_good" {
			fixedID = i.ID
		}
	}
	s.Compatibility.InputBindings[0].OriginalInputID = placeholderID
	fixed := s.Compatibility.InputBindings[0]
	fixed.OriginalInputID, fixed.ResolvedValue = fixedID, nil
	s.Compatibility.InputBindings = append(s.Compatibility.InputBindings, fixed)
	r, err := Recheck(q)
	if err != nil {
		t.Fatal(err)
	}
	if r.Assessment.CIExitCode != 0 {
		t.Fatalf("role assessment: %+v", r.Assessment)
	}
	inputs := r.Assessment.Entries[0].Resolution.Variants[0].Compatibility.Inputs
	if len(inputs) != 2 || inputs[0].OriginalInputID == inputs[1].OriginalInputID || inputs[0].CandidateInputID != inputs[1].CandidateInputID {
		t.Fatal("original roles merged")
	}
	s.Compatibility.InputBindings[0].OriginalInputID = inputs[0].CandidateInputID
	rejected, err := Recheck(q)
	if err != nil {
		t.Fatal(err)
	}
	if rejected.Assessment.Entries[0].Failure == nil || rejected.Assessment.CIExitCode != 2 {
		t.Fatal("candidate identity authorized original binding")
	}
}

func TestFormatRecheckPreservesCanonicalAssessment(t *testing.T) {
	r, err := Recheck(queryRecheck(t, "from [{id:1}]"))
	if err != nil {
		t.Fatal(err)
	}
	if FormatRecheck(nil) != "" || !bytes.Contains([]byte(FormatRecheck(r)), []byte(r.ProposedSourceHash)) || !bytes.Contains([]byte(FormatRecheck(r)), []byte(FormatReport(&r.Assessment))) {
		t.Fatal("format lost assessment/source")
	}
	if !reflect.DeepEqual(r.Assessment.Counts, Counts{Selected: 1, Assessed: 1, Valid: 1}) {
		t.Fatal("counts changed")
	}
}

func TestRecheckNewRequiredBindingRemainsIncomplete(t *testing.T) {
	q := queryRecheck(t, "from events_good | fields id")
	r, err := Recheck(q)
	if err != nil {
		t.Fatal(err)
	}
	e := r.Assessment.Entries[0]
	if r.Assessment.CIExitCode != 3 || !r.Assessment.ExecutionComplete || e.Failure != nil || e.Compatibility == nil || e.Analysis.Document.Text != q.Proposal.Document.Text {
		t.Fatalf("canonical missing binding lost: %+v", r.Assessment)
	}
}

func TestRecheckCanonicalV2ContextAndUnicode(t *testing.T) {
	q := queryRecheck(t, "from [{message:\"😀\"}] | fields message")
	raw, err := os.ReadFile("../../examples/environment/observed-partial-snapshot.json")
	if err != nil {
		t.Fatal(err)
	}
	snapshot, diagnostics := environment.DecodeSnapshot(raw)
	if (diagnostics != nil && diagnostics.Status == "invalid") || snapshot.SchemaVersion != 2 {
		t.Fatalf("v2 fixture: %+v", diagnostics)
	}
	q.Context.Snapshot, q.Context.SchemaBundle = snapshot, nil
	before, _ := json.Marshal(q)
	typed, err := Recheck(q)
	if err != nil {
		t.Fatal(err)
	}
	wire, err := RecheckJSON(before)
	if err != nil {
		t.Fatal(err)
	}
	a, _ := json.Marshal(typed)
	b, _ := json.Marshal(wire)
	if !bytes.Equal(a, b) || wire.Assessment.Entries[0].Analysis.Document.Text != q.Proposal.Document.Text {
		t.Fatal("canonical v2 or Unicode lost")
	}
	after, _ := json.Marshal(q)
	if !bytes.Equal(before, after) {
		t.Fatal("v2 artifacts mutated")
	}
}
