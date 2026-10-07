package analysis_test

import (
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
)

// These expectations are source-authored before publishing replay receipts.
// Schema evidence and alias spellings cannot supply an owner.
func TestCapabilityPublicationSourceOwnership(t *testing.T) {
	type witness struct {
		name, query  string
		fields       []string
		nodes, edges int
		outcome      string
	}
	witnesses := []witness{
		{"default-inner", `from $events | join left=e right=u where e.id=u.uid [from $users]`, []string{"$events:id:required", "$users:uid:required"}, 2, 1, "connected"},
		{"explicit-inner", `from $events | join type=inner left=e right=u where e.id=u.uid [from $users]`, []string{"$events:id:required", "$users:uid:required"}, 2, 1, "connected"},
		{"max-zero", `from $events | join max=0 left=e right=u where e.id=u.uid [from $users]`, []string{"$events:id:required", "$users:uid:required"}, 2, 1, "connected"},
		{"max-two", `from $events | join max=2 left=e right=u where e.id=u.uid [from $users]`, []string{"$events:id:required", "$users:uid:required"}, 2, 1, "connected"},
		{"left", `from $events | join type=left left=e right=u where e.id=u.uid [from $users]`, []string{"$events:id:required", "$users:uid:required"}, 2, 1, "connected"},
		{"outer", `from $events | join type=outer left=e right=u where e.id=u.uid [from $users]`, []string{"$events:id:required", "$users:uid:required"}, 2, 1, "connected"},
		{"sql-inner", `SELECT e.id, u.uid FROM $events AS e INNER JOIN $users AS u ON e.id=u.uid`, []string{"$events:id:required", "$users:uid:required"}, 2, 1, "connected"},
		{"sql-reversed", `SELECT e.id, u.uid FROM $events AS e INNER JOIN $users AS u ON u.uid=e.id`, []string{"$events:id:required", "$users:uid:required"}, 2, 1, "connected"},
		{"sql-left", `SELECT e.id, u.uid FROM $events AS e LEFT JOIN $users AS u ON e.id=u.uid`, []string{"$events:id:required", "$users:uid:conditional"}, 2, 1, "connected"},
		{"chain", `from $events | fields id | join left=e right=u where e.id=u.uid [from $users | fields uid] | join left=p right=a where p.uid=a.aid [from $accounts | fields aid]`, []string{"$events:id:required", "$users:uid:required", "$accounts:aid:required"}, 3, 2, "connected"},
		{"renamed-chain", `from $events | fields user_id, asset_id | join left=e right=u where e.user_id=u.id [from $users | fields id] | join left=eu right=a where eu.asset_id=a.asset_key [from $assets | rename id AS asset_key | fields asset_key] | eval copied_id=user_id | fields copied_id`, []string{"$events:user_id:required", "$events:asset_id:required", "$users:id:required", "$assets:id:required"}, 3, 2, "connected"},
		{"repeated", `from $events | join left=e right=u where e.id=u.id [from $events]`, []string{"$events:id:required"}, 2, 1, "connected"},
		{"independent", `from $events | join left=e right=u where e.id=u.id [from $events] | union [from $events]`, []string{"$events:id:required"}, 3, 1, "disconnected"},
		{"parameter", `FROM $target_1 | where actor.name="sample"`, []string{"$target_1:actor.name:required"}, 1, 0, "not applicable"},
	}
	for _, tc := range witnesses {
		t.Run(tc.name, func(t *testing.T) {
			r, err := analysis.Analyze(analysis.QueryDocument{Text: tc.query, Language: "spl2", SourceID: "capability:ownership:" + tc.name})
			if err != nil {
				t.Fatal(err)
			}
			wantStatus, wantComplete := analysis.Valid, true
			if tc.name == "sql-left" {
				wantComplete = false
			}
			if tc.name == "repeated" || tc.name == "independent" {
				wantStatus = analysis.Incomplete
				wantComplete = false
			}
			if r.Status != wantStatus || r.Requirements.Coverage.Complete != wantComplete || r.InputCoverage.State != "complete" || r.FieldAttributionCoverage.State != "complete" {
				t.Fatalf("incomplete witness: %+v", r.Requirements)
			}
			if len(r.Correlation.Nodes) != tc.nodes || len(r.Correlation.Edges) != tc.edges || r.Correlation.Outcome != tc.outcome {
				t.Fatalf("correlation: %+v", r.Correlation)
			}
			inputs := map[string]analysis.QueryInput{}
			occurrenceOwners := map[string]string{}
			for _, input := range r.Inputs {
				inputs[input.ID] = input
				for _, occ := range input.Occurrences {
					occurrenceOwners[occ.ID] = input.ID
				}
			}
			got := []string{}
			for _, item := range r.Requirements.Items {
				if item.Kind != "field" {
					continue
				}
				input, ok := inputs[item.InputID]
				if !ok || item.Ownership.State != "proved" || !reflect.DeepEqual(item.Ownership.CandidateInputIDs, []string{item.InputID}) || item.FieldIdentity == nil || item.FieldIdentity.Qualifier != "" {
					t.Fatalf("unproved source-relative field: %+v", item)
				}
				got = append(got, input.Name+":"+strings.Join(item.FieldIdentity.Segments, ".")+":"+item.Necessity)
				for _, occ := range item.Occurrences {
					if len(occ.InputOccurrenceIDs) == 0 || occ.Necessity != item.Necessity {
						t.Fatalf("lost obligation occurrence: %+v", occ)
					}
					for _, id := range occ.InputOccurrenceIDs {
						if occurrenceOwners[id] != item.InputID {
							t.Fatalf("wrong occurrence owner: %+v", occ)
						}
					}
				}
			}
			want := append([]string{}, tc.fields...)
			sort.Strings(got)
			sort.Strings(want)
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("fields=%v want=%v", got, want)
			}
			if !reflect.DeepEqual(r.Inputs, r.Requirements.Inputs) || !reflect.DeepEqual(r.Correlation, r.Requirements.Correlation) {
				t.Fatal("analysis/requirements projections differ")
			}
		})
	}
}

func TestCapabilityRequirementComparisonRejectsOwnershipMismatch(t *testing.T) {
	set, err := analysis.Requirements(analysis.QueryDocument{Text: `SELECT e.id, u.uid FROM $events AS e LEFT JOIN $users AS u ON e.id=u.uid`, Language: "spl2"})
	if err != nil {
		t.Fatal(err)
	}
	mutations := map[string]func(*analysis.CapabilityRequirementExpectation){
		"wrong owner":          func(v *analysis.CapabilityRequirementExpectation) { v.InputID = "other" },
		"wrong candidate":      func(v *analysis.CapabilityRequirementExpectation) { v.Ownership.CandidateInputIDs[0] = "other" },
		"fabricated proof":     func(v *analysis.CapabilityRequirementExpectation) { v.Ownership.State = "unproved" },
		"lost occurrence":      func(v *analysis.CapabilityRequirementExpectation) { v.Occurrences[0].InputOccurrenceIDs = []string{} },
		"wrong conditionality": func(v *analysis.CapabilityRequirementExpectation) { v.Occurrences[0].Necessity = "required" },
	}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			facts := requirementItemFacts(set.Items)
			at := -1
			for i, v := range facts {
				if v.Kind == "field" && v.Necessity == "conditional" {
					at = i
					break
				}
			}
			if at < 0 {
				t.Fatal("missing conditional RHS")
			}
			expected := analysis.CapabilityRequirementsObservation{QueryStatus: set.QueryStatus, Complete: false, Items: []analysis.CapabilityRequirementExpectation{facts[at]}, GapCodes: []string{}}
			if !containsAllRequirementFacts(facts, expected.Items, false) {
				t.Fatal("baseline mismatch")
			}
			mutate(&expected.Items[0])
			if containsAllRequirementFacts(requirementItemFacts(set.Items), expected.Items, false) {
				t.Fatal("incomplete subset accepted fabricated ownership")
			}
		})
	}
}

func TestCapabilityPublicationConditionalAndHeldOwnership(t *testing.T) {
	r, err := analysis.Analyze(analysis.QueryDocument{Text: `from $events | fields id | join type=left left=e right=u where e.id=u.uid [from $users | fields uid,team] | where team="security"`, Language: "spl2"})
	if err != nil {
		t.Fatal(err)
	}
	owners := map[string]string{}
	for _, input := range r.Inputs {
		owners[input.ID] = input.Name
	}
	foundKey, foundConditional := false, false
	for _, item := range r.Requirements.Items {
		if item.Kind != "field" {
			continue
		}
		if owners[item.InputID] != "$users" && !(item.InputID == "" && len(item.Ownership.CandidateInputIDs) == 1 && owners[item.Ownership.CandidateInputIDs[0]] == "$users") {
			continue
		}
		if item.Identity == "uid" {
			foundKey = item.Necessity == "required" && item.Ownership.State == "proved"
		}
		if item.Identity == "team" {
			for _, occ := range item.Occurrences {
				if occ.OriginalName == "team" && occ.Location.Start.Offset > 110 {
					foundConditional = occ.Necessity == "conditional" && item.InputID == "" && item.Ownership.State == "unproved" && len(item.Ownership.CandidateInputIDs) == 1
				}
			}
		}
	}
	if !foundKey || !foundConditional {
		t.Fatalf("lost predicate or conditional read: %+v", r.Requirements.Items)
	}
	held, err := analysis.Analyze(analysis.QueryDocument{Text: `from $events | join left=e right=u where e.actor.id=u.id [from $users] | fields id`, Language: "spl2"})
	if err != nil {
		t.Fatal(err)
	}
	if held.Correlation.Outcome != "indeterminate" || len(held.Correlation.Edges) != 0 || held.FieldAttributionCoverage.State != "partial" {
		t.Fatalf("held predicate proof: %+v", held.Correlation)
	}
	found := false
	for _, item := range held.Requirements.Items {
		if item.Kind == "field" && item.Identity == "id" && len(item.Occurrences) > 0 && item.Occurrences[len(item.Occurrences)-1].OriginalName == "id" {
			found = item.InputID == "" && item.Ownership.State == "unproved" && len(item.Ownership.CandidateInputIDs) == 2
		}
	}
	if !found {
		t.Fatalf("held downstream read lost candidates: %+v", held.Requirements.Items)
	}
}

// Publication uses the existing Dataset denominator and independent proof controls.
// Environment evidence cannot promote a grammar-only descriptor to a named slot.
func TestCapabilityPublicationResolutionControls(t *testing.T) {
	manifest, err := analysis.CapabilitiesFor(analysis.CapabilityOptions{Language: "spl2"})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, record := range manifest.Records {
		if record.ID == "spl2.dataset.dataset.parameter" {
			found = record.Dimensions.SafeRewriting.State == "supported" &&
				reflect.DeepEqual(record.Dimensions.SafeRewriting.EvidenceIDs, []string{"spl2.dataset.parameter.resolution-render"})
		}
	}
	if !found {
		t.Fatal("named Dataset publication lacks narrow replay evidence")
	}
	session, err := analysis.PrepareResolution(analysis.QueryDocument{Text: "from $events | fields id", Language: "spl2"})
	if err != nil {
		t.Fatal(err)
	}
	rendering, err := session.Render([]analysis.ResolutionChoice{{Placeholder: "$events", Kind: "dataset", Value: "events"}})
	if err != nil {
		t.Fatal(err)
	}
	proof, err := session.Verify(rendering)
	if err != nil || !proof.Evidence().Proven || len(proof.Evidence().Roles) != 1 {
		t.Fatalf("named slot proof: %v", err)
	}
	if rendering.CandidateDocument().Text != "from events | fields id" {
		t.Fatal("candidate text")
	}
	dynamic, err := analysis.PrepareResolution(analysis.QueryDocument{Text: `FROM {kind: lower("index"), properties: {name: dataset_name}}`, Language: "spl2"})
	if err != nil {
		t.Fatal(err)
	}
	if len(dynamic.Evidence().Placeholders) != 0 {
		t.Fatal("dynamic descriptor advertised as a named slot")
	}
	if _, err := dynamic.Render([]analysis.ResolutionChoice{{Placeholder: "$events", Kind: "dataset", Value: "events"}}); err == nil {
		t.Fatal("dynamic descriptor accepted an unrelated resolution selection")
	}
}
