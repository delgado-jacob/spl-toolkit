package resolution_test

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"reflect"
	"strconv"
	"testing"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/compatibility"
	"github.com/delgado-jacob/spl-toolkit/pkg/environment"
	"github.com/delgado-jacob/spl-toolkit/pkg/resolution"
)

// These fixtures freeze authored expectations before the resolver exists. The
// operation assertions below compare producer output with these authored values;
// this reader never derives expected outcomes or candidate text from output.
type acceptanceCase struct {
	ID            string                 `json:"id"`
	Document      analysis.QueryDocument `json:"document"`
	Resolutions   []fixtureResolution    `json:"resolutions"`
	MaxVariants   *int                   `json:"max_variants,omitempty"`
	Compatibility fixtureEvidence        `json:"compatibility"`
	Expected      fixtureExpectation     `json:"expected"`
}
type fixtureResolution struct {
	Marker string   `json:"marker"`
	Kind   string   `json:"kind"`
	Values []string `json:"values"`
}
type fixtureInputSelector struct {
	Kind     string                 `json:"kind"`
	Identity analysis.InputIdentity `json:"identity"`
}
type fixtureBinding struct {
	OriginalInput fixtureInputSelector       `json:"original_input"`
	ResolvedValue string                     `json:"resolved_value,omitempty"`
	ObjectID      string                     `json:"object_id"`
	Expected      environment.ObjectIdentity `json:"expected"`
	SchemaID      string                     `json:"schema_id,omitempty"`
}
type fixtureEvidence struct {
	Snapshot      environment.Snapshot      `json:"snapshot"`
	SchemaBundle  *environment.SchemaBundle `json:"schema_bundle"`
	QueryScope    environment.CaptureScope  `json:"query_scope"`
	InputBindings []fixtureBinding          `json:"input_bindings"`
}
type fixtureVariant struct {
	Selection     []string `json:"selection"`
	Outcome       string   `json:"outcome"`
	ResolvedQuery *string  `json:"resolved_query"`
}
type fixtureExpectation struct {
	Variants           []fixtureVariant `json:"variants"`
	Counts             map[string]int   `json:"counts"`
	SemanticAssertions []string         `json:"semantic_assertions"`
	CombinationCount   int              `json:"combination_count,omitempty"`
	RequestError       string           `json:"request_error,omitempty"`
}

func readAcceptanceCases(t *testing.T) []acceptanceCase {
	t.Helper()
	data, err := os.ReadFile("../../testdata/resolution/cases.json")
	if err != nil {
		t.Fatal(err)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var cases []acceptanceCase
	if err := decoder.Decode(&cases); err != nil {
		t.Fatal(err)
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		t.Fatalf("trailing fixture data: %v", err)
	}
	return cases
}

// Canonical typed identity selects original inputs. No opaque input or occurrence
// hashes are fixture literals, and equal replacement names cannot alter selection.
func originalInputID(t *testing.T, result *analysis.Result, selector fixtureInputSelector) string {
	t.Helper()
	var id string
	for _, input := range result.Inputs {
		if input.Kind == selector.Kind && input.Identity == selector.Identity {
			if id != "" {
				t.Fatalf("ambiguous fixture selector: %+v", selector)
			}
			id = input.ID
		}
	}
	if id == "" {
		t.Fatalf("original input not found: %+v in %+v", selector, result.Inputs)
	}
	return id
}

func TestAcceptanceFixtureIntegrity(t *testing.T) {
	seen := map[string]bool{}
	for _, c := range readAcceptanceCases(t) {
		t.Run(c.ID, func(t *testing.T) {
			if c.ID == "" || seen[c.ID] {
				t.Fatal("missing or repeated fixture ID")
			}
			seen[c.ID] = true
			if len(c.Expected.SemanticAssertions) == 0 {
				t.Fatal("missing semantic expectations")
			}
			result, err := analysis.Analyze(c.Document)
			if err != nil {
				t.Fatal(err)
			}
			if result.Status == analysis.Invalid {
				t.Fatalf("invalid original fixture: %+v", result.Diagnostics)
			}
			for _, binding := range c.Compatibility.InputBindings {
				originalInputID(t, result, binding.OriginalInput)
			}
			if _, err := compatibility.Prepare(c.Compatibility.Snapshot, c.Compatibility.SchemaBundle); err != nil {
				t.Fatalf("fixture evidence invalid: %v", err)
			}
			counts := map[string]int{"verified": 0, "failed": 0, "incomplete": 0}
			total := 1
			for _, r := range c.Resolutions {
				if len(r.Values) == 0 {
					t.Fatal("empty fixture choices")
				}
				total *= len(r.Values)
			}
			if c.Expected.RequestError != "" {
				if len(c.Expected.Variants) != 0 || c.Expected.CombinationCount != total {
					t.Fatal("limit rejection must retain exact count and no variants")
				}
			} else {
				if len(c.Expected.Variants) != total {
					t.Fatalf("expected %d combinations, got %d", total, len(c.Expected.Variants))
				}
				for ordinal, v := range c.Expected.Variants {
					if _, ok := counts[v.Outcome]; !ok {
						t.Fatalf("invalid expected outcome %q", v.Outcome)
					}
					counts[v.Outcome]++
					if (v.Outcome == "verified") != (v.ResolvedQuery != nil) {
						t.Fatal("publication expectation contradicts outcome")
					}
					if len(v.Selection) != len(c.Resolutions) {
						t.Fatal("selection cardinality mismatch")
					}
					remainder := ordinal
					for i := len(c.Resolutions) - 1; i >= 0; i-- {
						values := c.Resolutions[i].Values
						if v.Selection[i] != values[remainder%len(values)] {
							t.Fatal("expected selection order must vary last placeholder fastest")
						}
						remainder /= len(values)
					}
				}
			}
			if !reflect.DeepEqual(counts, c.Expected.Counts) {
				t.Fatalf("counts disagree: %v / %v", counts, c.Expected.Counts)
			}
		})
	}
}

func fixtureRequest(t *testing.T, c acceptanceCase) resolution.Request {
	t.Helper()
	result, err := analysis.Analyze(c.Document)
	if err != nil {
		t.Fatal(err)
	}
	r := resolution.Request{SchemaVersion: 1, Document: c.Document, Resolutions: []resolution.Resolution{}, Compatibility: resolution.CompatibilityInputs{Snapshot: c.Compatibility.Snapshot, SchemaBundle: c.Compatibility.SchemaBundle, QueryScope: c.Compatibility.QueryScope, InputBindings: []compatibility.ResolutionBinding{}}}
	for _, v := range c.Resolutions {
		r.Resolutions = append(r.Resolutions, resolution.Resolution{Placeholder: v.Marker, Kind: v.Kind, Values: v.Values})
	}
	if c.MaxVariants != nil {
		maximum := uint64(*c.MaxVariants)
		r.MaxVariants = &maximum
	}
	for _, b := range c.Compatibility.InputBindings {
		binding := compatibility.ResolutionBinding{OriginalInputID: originalInputID(t, result, b.OriginalInput), ObjectID: b.ObjectID, Expected: b.Expected, SchemaID: b.SchemaID}
		if b.ResolvedValue != "" {
			value := b.ResolvedValue
			binding.ResolvedValue = &value
		}
		r.Compatibility.InputBindings = append(r.Compatibility.InputBindings, binding)
	}
	return r
}

func TestAuthoredResolutionAcceptance(t *testing.T) {
	for _, c := range readAcceptanceCases(t) {
		t.Run(c.ID, func(t *testing.T) {
			request := fixtureRequest(t, c)
			report, err := resolution.Resolve(request)
			if c.Expected.RequestError != "" {
				detail, ok := resolution.RequestErrorDetails(err)
				if !ok || detail.Code != c.Expected.RequestError || detail.TotalCombinations != strconv.Itoa(c.Expected.CombinationCount) || report != nil {
					t.Fatalf("atomic rejection: report=%+v error=%v", report, err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if report.GeneratedCount != uint64(len(c.Expected.Variants)) || report.TotalCombinations != strconv.Itoa(len(c.Expected.Variants)) {
				t.Fatalf("counts: %+v", report)
			}
			if report.Counts.Verified != uint64(c.Expected.Counts["verified"]) || report.Counts.Failed != uint64(c.Expected.Counts["failed"]) || report.Counts.Incomplete != uint64(c.Expected.Counts["incomplete"]) {
				t.Fatalf("outcome counts: %+v", report.Counts)
			}
			for i, expected := range c.Expected.Variants {
				got := report.Variants[i]
				if got.Ordinal != uint64(i+1) || got.Outcome != expected.Outcome || !reflect.DeepEqual(got.ResolvedQuery, expected.ResolvedQuery) {
					t.Fatalf("variant %d: outcome=%s want=%s resolved=%v want=%v proof=%+v compatibility=%+v", i, got.Outcome, expected.Outcome, got.ResolvedQuery, expected.ResolvedQuery, got.Proof, got.Compatibility)
				}
				if got.Outcome == "verified" && (!got.Proof.Proven || got.Compatibility == nil || got.Compatibility.Outcome != "satisfied") {
					t.Fatal("publication lacks independent proof and compatibility")
				}
				for _, change := range got.Changes {
					original := change.OriginalLocation
					if request.Document.Text[original.Start.Offset:original.End.Offset] != change.Before {
						t.Fatal("change original byte range differs")
					}
					if change.CandidateLocation != nil {
						candidate := *change.CandidateLocation
						if (*got.CandidateText)[candidate.Start.Offset:candidate.End.Offset] != change.After {
							t.Fatal("change candidate byte range differs")
						}
					}
				}
				for j, choice := range got.Selection {
					if choice.Value != expected.Selection[j] {
						t.Fatalf("selection %d: %+v", i, got.Selection)
					}
				}
			}
			prepared, err := resolution.Prepare(request.Compatibility.Snapshot, request.Compatibility.SchemaBundle)
			if err != nil {
				t.Fatal(err)
			}
			next, err := prepared.Resolve(resolution.PreparedRequest{SchemaVersion: 1, Document: request.Document, Resolutions: request.Resolutions, MaxVariants: request.MaxVariants, Compatibility: compatibility.ResolutionAssessment{QueryScope: request.Compatibility.QueryScope, InputBindings: request.Compatibility.InputBindings}})
			if err != nil {
				t.Fatal(err)
			}
			before, _ := json.Marshal(report)
			after, _ := json.Marshal(next)
			if !bytes.Equal(before, after) {
				t.Fatal("prepared and one-shot differ")
			}
			repeated, err := resolution.Resolve(request)
			if err != nil {
				t.Fatal(err)
			}
			again, _ := json.Marshal(repeated)
			if !bytes.Equal(before, again) {
				t.Fatal("repeated output differs")
			}
		})
	}
}
