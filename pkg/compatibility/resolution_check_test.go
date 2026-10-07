package compatibility

import (
	"encoding/json"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/environment"
)

func resolutionCheckFixture(t *testing.T, text string) (*analysis.ResolutionProof, Request, ResolutionAssessment) {
	t.Helper()
	s, err := analysis.PrepareResolution(analysis.QueryDocument{Text: text, Language: "spl2"})
	if err != nil {
		t.Fatal(err)
	}
	o := s.Evidence()
	choices := []analysis.ResolutionChoice{}
	for _, group := range o.Placeholders {
		choices = append(choices, analysis.ResolutionChoice{Placeholder: group.Placeholder, Kind: group.Kind, Value: "events"})
	}
	rendered, err := s.Render(choices)
	if err != nil {
		t.Fatal(err)
	}
	proof, err := s.Verify(rendered)
	if err != nil {
		t.Fatal(err)
	}
	if !proof.Evidence().Proven {
		t.Fatalf("unproved fixture: %+v", proof.Evidence())
	}
	r := assessmentFixture(t, text)
	expected := objectIdentity(r.Snapshot.Objects[0])
	r.SchemaBundle.Bindings = []environment.SchemaBinding{{SchemaID: "fields", ObjectID: "events", Expected: expected, SourceCoverage: "complete"}, {SchemaID: "missing", ObjectID: "events", Expected: expected, SourceCoverage: "complete"}}
	r.SchemaBundle.Schemas = append(r.SchemaBundle.Schemas, environment.SchemaEntry{ID: "missing", Kind: "field_list", Catalog: json.RawMessage(`{"fields":["other"],"optional_fields":[],"identity":"missing"}`), Provenance: r.SchemaBundle.Provenance})
	a := ResolutionAssessment{QueryScope: r.QueryScope, InputBindings: []ResolutionBinding{}}
	for _, input := range o.Analysis.Inputs {
		b := ResolutionBinding{OriginalInputID: input.ID, ObjectID: "events", Expected: expected, SchemaID: "fields"}
		for _, g := range o.Placeholders {
			for _, id := range g.OriginalInputIDs {
				if id == input.ID {
					v := "events"
					b.ResolvedValue = &v
				}
			}
		}
		if input.Kind == "explicit_dataset" {
			b.ObjectID = input.Name
			b.Expected.Name = input.Name
			b.SchemaID = ""
		}
		a.InputBindings = append(a.InputBindings, b)
	}
	return proof, r, a
}
func checkedResolution(t *testing.T, proof *analysis.ResolutionProof, r Request, a ResolutionAssessment) *ResolutionReport {
	t.Helper()
	p, err := Prepare(r.Snapshot, r.SchemaBundle)
	if err != nil {
		t.Fatal(err)
	}
	report, err := p.CheckResolution(proof, a)
	if err != nil {
		t.Fatal(err)
	}
	return report
}
func TestResolutionCheckSharedCandidateRoleIsolation(t *testing.T) {
	for _, text := range []string{`from $left | join type=inner left=L right=R where L.id=R.id [from $right]`, `from $left | join type=inner left=L right=R where L.id=R.id [from events]`} {
		proof, r, a := resolutionCheckFixture(t, text)
		for i := range a.InputBindings {
			if a.InputBindings[i].ResolvedValue == nil || i == 1 {
				a.InputBindings[i].SchemaID = "missing"
			}
		}
		report := checkedResolution(t, proof, r, a)
		if report.Outcome != "unsatisfied" {
			t.Fatalf("got %s: %+v", report.Outcome, report)
		}
		if len(report.Inputs) != 2 || report.Inputs[0].OriginalInputID == report.Inputs[1].OriginalInputID || report.Inputs[0].CandidateInputID != report.Inputs[1].CandidateInputID {
			t.Fatalf("roles lost: %+v", report.Inputs)
		}
		states := map[string]bool{}
		caps := 0
		for _, out := range report.RequirementOutcomes {
			if out.CandidateRequirementID == "capability:language:spl2:profile:splunkd" {
				caps++
				continue
			}
			if out.Evidence.RequirementID != out.CandidateRequirementID || out.Evidence.InputID != out.CandidateInputID {
				t.Fatal("noncanonical child IDs")
			}
			if out.Evidence.FieldProjection != nil {
				states[out.Evidence.Outcome] = true
				if len(out.AssessedOccurrences) != 1 {
					t.Fatal("occurrences not isolated")
				}
			}
		}
		if !states["missing"] || !states["satisfied"] || caps != 1 {
			t.Fatalf("lost outcomes: %+v", report.RequirementOutcomes)
		}
		_, candidate, _, _ := proof.AssessmentEvidence()
		if !reflect.DeepEqual(report.Requirements, candidate.Analysis.Requirements) {
			t.Fatal("canonical candidate requirements modified")
		}
	}
}
func TestResolutionCheckAuthority(t *testing.T) {
	proof, r, a := resolutionCheckFixture(t, `from $left | fields id`)
	p, err := Prepare(r.Snapshot, r.SchemaBundle)
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range []*analysis.ResolutionProof{nil, {}, new(analysis.ResolutionProof)} {
		if _, err := p.CheckResolution(v, a); err == nil {
			t.Fatal("authority accepted")
		}
	}
	if _, err := p.CheckResolution(proof, a); err != nil {
		t.Fatal(err)
	}
}
func TestResolutionCheckOriginalUnowned(t *testing.T) {
	proof, r, a := resolutionCheckFixture(t, `from $left | join type=inner left=L right=R where L.id=R.id [from $right] | where id=1`)
	report := checkedResolution(t, proof, r, a)
	if report.Outcome != "incomplete" {
		t.Fatalf("original ownership became complete: %+v", report)
	}
	found := false
	for _, out := range report.RequirementOutcomes {
		if out.OriginalInputID == "" && out.Evidence.Outcome != "satisfied" && out.Evidence.RequirementID != "capability:language:spl2:profile:splunkd" {
			found = true
		}
	}
	if !found {
		t.Fatal("unattributable source obligation lost")
	}
	for _, out := range report.RequirementOutcomes {
		for _, item := range report.Requirements.Items {
			if item.ID == out.CandidateRequirementID && (out.Evidence.InputID != item.InputID || out.CandidateInputID != item.InputID) {
				t.Fatal("conservative ownership projection changed canonical coordinates")
			}
		}
	}
}

func TestResolutionCheckConditionalRoleNecessity(t *testing.T) {
	proof, r, a := resolutionCheckFixture(t, `$required = from $left | where id=1; $conditional = from seed | branch (guard=true) [from $right | where id=1], (guard=false) [from $right | where other=2];`)
	seed := r.Snapshot.Objects[0]
	seed.ID = "seed"
	seed.Name = "seed"
	r.Snapshot.Objects = append(r.Snapshot.Objects, seed)
	r.SchemaBundle.Schemas[0].Catalog = json.RawMessage(`{"fields":["id","other","guard"],"optional_fields":[],"identity":"fields"}`)
	seedExpected := objectIdentity(seed)
	r.SchemaBundle.Bindings = append(r.SchemaBundle.Bindings, environment.SchemaBinding{ObjectID: "seed", Expected: seedExpected, SchemaID: "fields", SourceCoverage: "complete"})
	original, _, _, _ := proof.AssessmentEvidence()
	right := ""
	for _, input := range original.Analysis.Inputs {
		if input.Name == "$right" {
			right = input.ID
		}
	}
	for i := range a.InputBindings {
		if a.InputBindings[i].OriginalInputID == right {
			a.InputBindings[i].SchemaID = "missing"
		}
		if a.InputBindings[i].Expected.Name == "seed" {
			a.InputBindings[i].SchemaID = "fields"
		}
	}
	report := checkedResolution(t, proof, r, a)
	if report.Outcome != "incomplete" {
		t.Fatalf("conditional failure became required: %s %+v", report.Outcome, report.Reasons)
	}
	found := false
	for _, out := range report.RequirementOutcomes {
		if out.OriginalInputID == right && out.Evidence.FieldProjection != nil && out.Evidence.Outcome == "missing" {
			found = true
			if out.Evidence.Applicability != "indeterminate" {
				t.Fatal("canonical strongest necessity copied")
			}
		}
	}
	if !found {
		t.Fatal("conditional missing fact lost")
	}
	// Use a document containing only a conditional role, keeping the seed present.
	proof, r, a = resolutionCheckFixture(t, `from seed | branch (guard=true) [from $right], (guard=false) [from seed]`)
	r.Snapshot.Objects = []environment.Object{seed}
	r.SchemaBundle.Schemas[0].Catalog = json.RawMessage(`{"fields":["guard"],"optional_fields":[],"identity":"fields"}`)
	r.SchemaBundle.Bindings = append(r.SchemaBundle.Bindings, environment.SchemaBinding{ObjectID: "seed", Expected: seedExpected, SchemaID: "fields", SourceCoverage: "complete"})
	for i := range a.InputBindings {
		if a.InputBindings[i].Expected.Name == "seed" {
			a.InputBindings[i].SchemaID = "fields"
		}
	}
	report = checkedResolution(t, proof, r, a)
	if report.Outcome != "incomplete" {
		t.Fatalf("conditional object absence: %s %+v", report.Outcome, report.Reasons)
	}
	for _, out := range report.Inputs {
		if out.Evidence.Outcome == "missing" {
			t.Fatal("conditional source independently blocked")
		}
	}
}

func TestResolutionCheckEvidenceLimits(t *testing.T) {
	for _, tc := range []struct {
		name, want string
		change     func(*Request, *ResolutionAssessment)
	}{
		{"partial absent collection", "incomplete", func(r *Request, a *ResolutionAssessment) {
			r.Snapshot.Objects = []environment.Object{}
			r.Snapshot.Collections[0].Coverage = "partial"
			r.Snapshot.Collections[0].Reason = "bounded"
		}},
		{"complete absent collection", "unsatisfied", func(r *Request, a *ResolutionAssessment) { r.Snapshot.Objects = []environment.Object{} }},
		{"partial schema negative", "incomplete", func(r *Request, a *ResolutionAssessment) {
			a.InputBindings[0].SchemaID = "missing"
			r.SchemaBundle.Bindings[1].SourceCoverage = "partial"
			r.SchemaBundle.Bindings[1].Reason = "bounded"
		}},
		{"missing plus unknown", "unsatisfied", func(r *Request, a *ResolutionAssessment) {
			r.Snapshot.Objects = []environment.Object{}
			a.InputBindings[0].SchemaID = ""
		}},
		{"binding absent", "incomplete", func(r *Request, a *ResolutionAssessment) { a.InputBindings = []ResolutionBinding{} }},
		{"partial positive", "satisfied", func(r *Request, a *ResolutionAssessment) {
			r.Snapshot.Collections[0].Coverage = "partial"
			r.Snapshot.Collections[0].Reason = "bounded"
		}},
		{"hidden selected dataset", "satisfied", func(r *Request, a *ResolutionAssessment) {
			r.Snapshot.Objects[0].Document = &analysis.QueryDocument{Language: "spl2", Text: "from users"}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			proof, r, a := resolutionCheckFixture(t, `from $left | fields id`)
			tc.change(&r, &a)
			report := checkedResolution(t, proof, r, a)
			if tc.name == "hidden selected dataset" {
				rootField, hiddenDataset := false, false
				for _, out := range report.RequirementOutcomes {
					if out.OriginalInputID != "" && out.Evidence.FieldProjection != nil {
						rootField = out.Evidence.Outcome == "satisfied" && len(out.Evidence.Schemas) == 1 && out.Evidence.Schemas[0].SchemaID == "fields"
					}
					if out.Evidence.DefinitionObjectID == "events" && len(out.AssessedOccurrences) > 0 && out.AssessedOccurrences[0].OriginalName == "users" {
						hiddenDataset = out.OriginalInputID == "" && out.Evidence.Outcome == "satisfied" && len(out.Evidence.Schemas) == 0 && len(out.Evidence.InvocationProvenance) > 0
					}
				}
				_, candidate, _, _ := proof.AssessmentEvidence()
				if report.Closure.EffectiveAnalysis.Document.Text != candidate.Analysis.Document.Text || len(report.EffectiveRequirements.Inputs) != 1 || report.EffectiveRequirements.Inputs[0].Name != "events" {
					t.Fatal("dataset body changed effective root supplier")
				}
				ordinaryBindings := []InputBinding{}
				for _, binding := range a.InputBindings {
					ordinaryBindings = append(ordinaryBindings, InputBinding{InputID: candidate.Analysis.Inputs[0].ID, ObjectID: binding.ObjectID, Expected: binding.Expected, SchemaID: binding.SchemaID})
				}
				prepared, err := Prepare(r.Snapshot, r.SchemaBundle)
				if err != nil {
					t.Fatal(err)
				}
				doc := candidate.Analysis.Document
				ordinary, err := prepared.Check(AssessmentRequest{SchemaVersion: 1, Document: &doc, Requirements: candidate.Analysis.Requirements, QueryScope: a.QueryScope, InputBindings: ordinaryBindings})
				if err != nil || ordinary.Outcome != "satisfied" {
					t.Fatalf("ordinary closure parity: %+v %v", ordinary, err)
				}
				for _, out := range report.RequirementOutcomes {
					if out.OriginalInputID != "" && out.Evidence.FieldProjection != nil {
						if len(out.Evidence.SourceIntervals) != 1 || out.Evidence.SourceIntervals[0].Kind != "query" || out.Evidence.SourceIntervals[0].Start != 21 || out.Evidence.SourceIntervals[0].End != 23 {
							t.Fatalf("root field interval=%+v", out.Evidence.SourceIntervals)
						}
					}
				}
				if !rootField || !hiddenDataset {
					t.Fatalf("root field or independent hidden dataset evidence lost: %+v", report.RequirementOutcomes)
				}
			}
			if report.Outcome != tc.want {
				t.Fatalf("got %s want %s: %+v", report.Outcome, tc.want, report.Reasons)
			}
		})
	}
}

func TestResolutionCheckTypedFields(t *testing.T) {
	proof, r, a := resolutionCheckFixture(t, `from $left | fields 'actor.name', actor.name`)
	r.SchemaBundle.Schemas[0].Catalog = nil
	r.SchemaBundle.Schemas[0].Kind = "json_schema"
	r.SchemaBundle.Schemas[0].Target = json.RawMessage(`{"kind":"json_schema","identity":"fields","schema":{"type":"object","properties":{"actor.name":true,"actor":{"type":"object","properties":{"name":true},"required":["name"],"additionalProperties":false}},"required":["actor"],"additionalProperties":false}}`)
	report := checkedResolution(t, proof, r, a)
	if report.Outcome != "satisfied" {
		t.Fatalf("typed fields: %+v", report.Reasons)
	}
	kinds := map[string]string{}
	for _, out := range report.RequirementOutcomes {
		if out.Evidence.FieldProjection != nil {
			for _, item := range report.Requirements.Items {
				if item.ID == out.CandidateRequirementID {
					kinds[item.FieldIdentity.Kind] = out.Evidence.FieldProjection.Outcome
				}
			}
		}
	}
	if kinds["atomic"] != "optional" || kinds["path"] != "required" {
		t.Fatalf("typed identities conflated: %+v", kinds)
	}
}

func TestResolutionCheckBindingAdmissionAndDeterminism(t *testing.T) {
	proof, r, a := resolutionCheckFixture(t, `from $left | join type=inner left=L right=R where L.id=R.id [from $right]`)
	p, err := Prepare(r.Snapshot, r.SchemaBundle)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name   string
		change func(*ResolutionAssessment)
	}{
		{"unknown role", func(a *ResolutionAssessment) { a.InputBindings[0].OriginalInputID = "invented" }},
		{"duplicate", func(a *ResolutionAssessment) { a.InputBindings = append(a.InputBindings, a.InputBindings[0]) }},
		{"wrong identity", func(a *ResolutionAssessment) { a.InputBindings[0].Expected.Name = "users" }},
		{"wrong schema", func(a *ResolutionAssessment) { a.InputBindings[0].SchemaID = "invented" }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			v := detach(a)
			tc.change(&v)
			if _, err := p.CheckResolution(proof, v); err == nil {
				t.Fatal("forged selection admitted")
			}
		})
	}
	alternative := detach(a.InputBindings[0])
	value := "users"
	alternative.ResolvedValue = &value
	alternative.ObjectID = "users"
	alternative.Expected.Name = value
	alternative.SchemaID = ""
	a.InputBindings = append(a.InputBindings, alternative)
	report, err := p.CheckResolution(proof, a)
	if err != nil {
		t.Fatal("unused alternative rejected", err)
	}
	raw, _ := json.Marshal(report)
	reverse := detach(a)
	for i, j := 0, len(reverse.InputBindings)-1; i < j; i, j = i+1, j-1 {
		reverse.InputBindings[i], reverse.InputBindings[j] = reverse.InputBindings[j], reverse.InputBindings[i]
	}
	again, err := p.CheckResolution(proof, reverse)
	if err != nil {
		t.Fatal(err)
	}
	other, _ := json.Marshal(again)
	if string(raw) != string(other) {
		t.Fatal("binding order changed report")
	}
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			next, err := p.CheckResolution(proof, a)
			if err != nil {
				t.Error(err)
				return
			}
			got, _ := json.Marshal(next)
			if string(raw) != string(got) {
				t.Error("prepared assessment state leaked")
			}
		}()
	}
	wg.Wait()
}

func TestResolutionCheckClassicImplicitStream(t *testing.T) {
	document := analysis.QueryDocument{Text: `search index=main | table id`, Language: "spl"}
	s, err := analysis.PrepareResolution(document)
	if err != nil {
		t.Fatal(err)
	}
	rendered, err := s.Render([]analysis.ResolutionChoice{})
	if err != nil {
		t.Fatal(err)
	}
	proof, err := s.Verify(rendered)
	if err != nil {
		t.Fatal(err)
	}
	_, candidate, _, ok := proof.AssessmentEvidence()
	if !ok {
		t.Fatal("implicit stream proof refused")
	}
	r := assessmentFixture(t, `from $left`)
	r.SchemaBundle = nil
	r.Snapshot.Objects = append(r.Snapshot.Objects, environment.Object{ID: "main", Kind: "index", Name: "main", Provenance: r.Snapshot.Objects[0].Provenance})
	r.Snapshot.Capabilities[0].ID = "language:" + candidate.Analysis.Requirements.Query.Language + ":profile:" + candidate.Analysis.Requirements.Query.Profile
	r.Snapshot.Capabilities[0].Version = candidate.Analysis.Requirements.Query.Version
	a := ResolutionAssessment{QueryScope: r.QueryScope, InputBindings: []ResolutionBinding{}}
	for _, input := range s.Evidence().Analysis.Inputs {
		a.InputBindings = append(a.InputBindings, ResolutionBinding{OriginalInputID: input.ID, ObjectID: "events", Expected: objectIdentity(r.Snapshot.Objects[0])})
	}
	report := checkedResolution(t, proof, r, a)
	if report.Outcome != "incomplete" {
		t.Fatalf("implicit stream became source proof: %s", report.Outcome)
	}
	found := false
	for _, reason := range report.Reasons {
		if reason.Evidence.Code == "target_discovery_incomplete" {
			found = true
		}
	}
	if !found {
		t.Fatal("implicit selection limit lost")
	}
}

func TestResolutionCheckRefusedSQLProof(t *testing.T) {
	s, err := analysis.PrepareResolution(analysis.QueryDocument{Text: `FROM $left AS L JOIN $right AS R ON L.id=R.id | fields id`, Language: "spl2"})
	if err != nil {
		t.Fatal(err)
	}
	rendering, err := s.Render([]analysis.ResolutionChoice{{Placeholder: "$left", Kind: "dataset", Value: "events"}, {Placeholder: "$right", Kind: "dataset", Value: "events"}})
	if err != nil {
		t.Fatal(err)
	}
	proof, err := s.Verify(rendering)
	if err != nil {
		t.Fatal(err)
	}
	r := assessmentFixture(t, `from $left`)
	p, err := Prepare(r.Snapshot, r.SchemaBundle)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := p.CheckResolution(proof, ResolutionAssessment{QueryScope: r.QueryScope, InputBindings: []ResolutionBinding{}}); err == nil || !strings.Contains(err.Error(), "resolution_proof_invalid") {
		t.Fatalf("unsafe SQL accepted: %v", err)
	}
}

func TestResolutionCheckBoundedObservationAbsence(t *testing.T) {
	s, err := analysis.PrepareResolution(analysis.QueryDocument{Text: `from {kind:"source",properties:{name:"absent"}}`, Language: "spl2"})
	if err != nil {
		t.Fatal(err)
	}
	rendered, err := s.Render([]analysis.ResolutionChoice{})
	if err != nil {
		t.Fatal(err)
	}
	proof, err := s.Verify(rendered)
	if err != nil {
		t.Fatal(err)
	}
	r := assessmentFixture(t, `from $left`)
	r.SchemaBundle = nil
	provenance := r.Snapshot.Objects[0].Provenance
	r.Snapshot.SchemaVersion = 2
	r.Snapshot.Collections = append(r.Snapshot.Collections, environment.Collection{Kind: "source", Coverage: "complete"}, environment.Collection{Kind: "sourcetype", Coverage: "complete"})
	r.Snapshot.Objects = []environment.Object{{ID: "index-main", Kind: "index", Name: "main", Provenance: provenance}}
	yes := true
	r.Snapshot.Observation = &environment.ObservationScope{IndexSelection: environment.Selector{All: &yes}, Enumeration: environment.IndexEnumeration{Method: "distributed_rest", PeerScope: "configured_search_peers", Coverage: "complete", Provenance: provenance}, Indexes: []environment.ObservationIndex{{IndexID: "index-main", CatalogDatatypes: []string{"event"}, RequiredDatatypes: []string{"event"}}}, UnmatchedIndexes: []string{}, Method: "splunk_metadata", Visibility: "exporting_principal", Window: environment.ObservationWindow{Mode: "all_retained"}, TimePrecision: "bucket_overlap", AbsenceMeaning: "not_observed", Captures: []environment.ObservationCapture{{Kind: "source", IndexID: "index-main", Datatype: "event", ObjectIDs: []string{}, Coverage: "complete", Provenance: provenance}, {Kind: "sourcetype", IndexID: "index-main", Datatype: "event", ObjectIDs: []string{}, Coverage: "complete", Provenance: provenance}}}
	original, _, _, ok := proof.AssessmentEvidence()
	if !ok {
		t.Fatal("descriptor proof unproved")
	}
	a := ResolutionAssessment{QueryScope: r.QueryScope, InputBindings: []ResolutionBinding{{OriginalInputID: original.Analysis.Inputs[0].ID, ObjectID: "absent", Expected: environment.ObjectIdentity{Kind: "source", Name: "absent"}}}}
	report := checkedResolution(t, proof, r, a)
	if report.Outcome != "incomplete" {
		t.Fatalf("bounded absence became global missing: %+v", report)
	}
	found := false
	for _, reason := range report.Reasons {
		if reason.Evidence.Code == "observation_scope_insufficient" && reason.OriginalInputID == a.InputBindings[0].OriginalInputID {
			found = true
		}
	}
	if !found {
		t.Fatal("role-qualified observation limitation lost")
	}
}

func TestResolutionCheckCanonicalCoordinatesAndDetachment(t *testing.T) {
	proof, r, a := resolutionCheckFixture(t, `from $left | join type=inner left=L right=R where L.id=R.id [from $right]`)
	for i := range a.InputBindings {
		a.InputBindings[i].SchemaID = "missing"
	}
	report := checkedResolution(t, proof, r, a)
	_, candidate, _, _ := proof.AssessmentEvidence()
	refs := map[string]bool{}
	inputs := map[string]bool{}
	requirements := map[string]bool{}
	for _, item := range candidate.Analysis.Requirements.Items {
		requirements[item.ID] = true
		for _, o := range item.Occurrences {
			refs[o.ReferenceID] = true
		}
	}
	for _, input := range candidate.Analysis.Inputs {
		inputs[input.ID] = true
		for _, o := range input.Occurrences {
			refs[o.ReferenceID] = true
		}
	}
	for _, wrapped := range report.Reasons {
		reason := wrapped.Evidence
		if reason.InputID != "" && !inputs[reason.InputID] {
			t.Fatal("original input leaked into child reason")
		}
		if reason.RequirementID != "" && !requirements[reason.RequirementID] && !strings.HasPrefix(reason.RequirementID, "capability:") {
			t.Fatal("original requirement leaked into child reason")
		}
		for _, ref := range reason.ReferenceIDs {
			if !refs[ref] {
				t.Fatalf("noncandidate reference %s", ref)
			}
		}
	}
	baseline := stableKey(report)
	a.InputBindings[0].Expected.Name = "mutated"
	original, detached, roles, _ := proof.AssessmentEvidence()
	original.Analysis.Requirements.Items[0].Identity = "mutated"
	detached.Analysis.Requirements.Items[0].Identity = "mutated"
	roles[0].CandidateInput.Name = "mutated"
	if stableKey(report) != baseline {
		t.Fatal("report aliases caller/evidence")
	}
}

func TestResolutionCheckPartialDiscoveryUnassessableBinding(t *testing.T) {
	proof, r, a := resolutionCheckFixture(t, `from $right | append [FROM {kind: $kind, properties: {name: "main"}}]`)
	original, _, _, _ := proof.AssessmentEvidence()
	if original.Coverage.State != "partial" {
		t.Fatal("fixture requires incomplete discovery")
	}
	unresolved := map[string]bool{}
	for _, input := range original.Analysis.Inputs {
		if input.Kind == "unresolved_source" {
			unresolved[input.ID] = true
		}
	}
	kept := []ResolutionBinding{}
	for _, b := range a.InputBindings {
		if !unresolved[b.OriginalInputID] {
			kept = append(kept, b)
		}
	}
	a.InputBindings = kept
	a.InputBindings = append(a.InputBindings, ResolutionBinding{OriginalInputID: "undiscovered", ObjectID: "users", Expected: objectIdentity(r.Snapshot.Objects[1])})
	report := checkedResolution(t, proof, r, a)
	if report.Outcome != "incomplete" {
		t.Fatalf("partial discovery falsely concluded: %s", report.Outcome)
	}
	found := false
	for _, reason := range report.Reasons {
		if reason.OriginalInputID == "undiscovered" && reason.Evidence.Code == "resolution_role_incomplete" {
			found = true
			if reason.Evidence.InputID != "" {
				t.Fatal("unproved candidate input fabricated")
			}
		}
	}
	if !found {
		t.Fatal("partial undiscovered binding silently ignored")
	}
}

func TestResolutionCheckDisconnectedCorrelationSeparate(t *testing.T) {
	proof, r, a := resolutionCheckFixture(t, `$a = from $left; $b = from $right;`)
	report := checkedResolution(t, proof, r, a)
	if report.Correlation.Outcome != "disconnected" {
		t.Fatalf("fixture correlation: %s", report.Correlation.Outcome)
	}
	if report.Outcome != "satisfied" {
		t.Fatalf("disconnected sources blocked compatibility: %s %+v", report.Outcome, report.Reasons)
	}
}

func TestResolutionCheckUnboundExplicitIdentity(t *testing.T) {
	for _, tc := range []struct {
		name, query, outcome string
		missing, partial     bool
	}{
		{"captured-no-fields", "from $left", "satisfied", false, false},
		{"complete-absence", "from $left", "unsatisfied", true, false},
		{"partial-absence", "from $left", "incomplete", true, true},
		{"unbound-field-schema", "from $left | fields id", "incomplete", false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			proof, r, a := resolutionCheckFixture(t, tc.query)
			a.InputBindings = []ResolutionBinding{}
			r.SchemaBundle = nil
			if tc.missing {
				r.Snapshot.Objects = []environment.Object{}
			}
			if tc.partial {
				for i := range r.Snapshot.Collections {
					if r.Snapshot.Collections[i].Kind == "dataset" {
						r.Snapshot.Collections[i].Coverage = "partial"
						r.Snapshot.Collections[i].Reason = "bounded capture"
					}
				}
			}
			report := checkedResolution(t, proof, r, a)
			if report.Outcome != tc.outcome {
				t.Fatalf("unbound explicit identity outcome = %s, want %s: %+v", report.Outcome, tc.outcome, report.Reasons)
			}
			_, candidate, _, _ := proof.AssessmentEvidence()
			ordinary := r
			ordinary.Requirements = candidate.Analysis.Requirements
			ordinary.InputBindings = []InputBinding{}
			baseline := checked(t, ordinary)
			if report.Outcome != baseline.Outcome {
				t.Fatalf("resolution %s disagrees with ordinary identity assessment %s", report.Outcome, baseline.Outcome)
			}
		})
	}
}

func TestResolutionCheckUnboundEqualNameDoesNotBorrowExplicitSchema(t *testing.T) {
	proof, r, a := resolutionCheckFixture(t, `from $left | join type=inner left=L right=R where L.id=R.id [from $right]`)
	unboundID := a.InputBindings[1].OriginalInputID
	boundID := a.InputBindings[0].OriginalInputID
	a.InputBindings = a.InputBindings[:1]
	report := checkedResolution(t, proof, r, a)
	if report.Outcome != "incomplete" {
		t.Fatalf("unbound role borrowed explicit schema: %+v", report)
	}
	boundSatisfied, unboundIndeterminate := false, false
	for _, out := range report.RequirementOutcomes {
		if out.OriginalInputID == boundID && out.Evidence.FieldProjection != nil && out.Evidence.Outcome == "satisfied" {
			boundSatisfied = true
		}
		if out.OriginalInputID == unboundID {
			for _, reason := range out.Evidence.Reasons {
				if reason.Code == "schema_not_supplied" && out.Evidence.Outcome == "indeterminate" && len(out.Evidence.Schemas) == 0 {
					unboundIndeterminate = true
				}
			}
		}
	}
	if !boundSatisfied || !unboundIndeterminate {
		t.Fatalf("role-specific schema assessment lost: %+v", report.RequirementOutcomes)
	}
}
