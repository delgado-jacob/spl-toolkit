package compatibility

import (
	"sync"
	"testing"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/environment"
)

func resolutionBindingsFixture(t *testing.T) (*Prepared, analysis.ResolutionEvidence, []analysis.ResolutionChoice, ResolutionAssessment) {
	t.Helper()
	r := assessmentFixture(t, "from $left | join left=L right=R $right on L.id=R.id")
	// Separate original typed roles may deliberately select the same object value.
	original := analysis.ResolutionEvidence{Analysis: analysis.Result{Inputs: []analysis.QueryInput{{ID: "left", Kind: "named_placeholder"}, {ID: "right", Kind: "named_placeholder"}, {ID: "fixed", Kind: "explicit_dataset", Identity: analysis.InputIdentity{Form: "identifier", Value: "events"}}}}, Coverage: analysis.InputCoverage{State: "complete"}, Placeholders: []analysis.ResolutionPlaceholder{{Placeholder: "$left", Kind: "dataset", OriginalInputIDs: []string{"left"}}, {Placeholder: "$right", Kind: "dataset", OriginalInputIDs: []string{"right"}}}}
	expected := objectIdentity(r.Snapshot.Objects[0])
	r.SchemaBundle.Bindings = []environment.SchemaBinding{{SchemaID: "fields", ObjectID: "events", Expected: expected, SourceCoverage: "complete"}, {SchemaID: "other", ObjectID: "events", Expected: expected, SourceCoverage: "complete"}}
	second := detach(r.SchemaBundle.Schemas[0])
	second.ID = "other"
	r.SchemaBundle.Schemas = append(r.SchemaBundle.Schemas, second)
	p, err := Prepare(r.Snapshot, r.SchemaBundle)
	if err != nil {
		t.Fatal(err)
	}
	value := "events"
	choices := []analysis.ResolutionChoice{{Placeholder: "$left", Kind: "dataset", Value: value}, {Placeholder: "$right", Kind: "dataset", Value: value}}
	a := ResolutionAssessment{QueryScope: r.QueryScope, InputBindings: []ResolutionBinding{{OriginalInputID: "left", ResolvedValue: &value, ObjectID: "events", Expected: expected, SchemaID: "fields"}, {OriginalInputID: "right", ResolvedValue: &value, ObjectID: "events", Expected: expected, SchemaID: "other"}, {OriginalInputID: "fixed", ObjectID: "events", Expected: expected}}}
	return p, original, choices, a
}

func TestResolutionBindingsAdmission(t *testing.T) {
	p, original, choices, assessment := resolutionBindingsFixture(t)
	if err := p.ValidateResolutionBindings(original, choices, assessment); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name   string
		mutate func(*analysis.ResolutionEvidence, *[]analysis.ResolutionChoice, *ResolutionAssessment)
	}{
		{"unknown exhaustive role", func(o *analysis.ResolutionEvidence, c *[]analysis.ResolutionChoice, a *ResolutionAssessment) {
			a.InputBindings[0].OriginalInputID = "unknown"
		}},
		{"duplicate key", func(o *analysis.ResolutionEvidence, c *[]analysis.ResolutionChoice, a *ResolutionAssessment) {
			a.InputBindings = append(a.InputBindings, a.InputBindings[0])
		}},
		{"substituted role without value", func(o *analysis.ResolutionEvidence, c *[]analysis.ResolutionChoice, a *ResolutionAssessment) {
			a.InputBindings[0].ResolvedValue = nil
		}},
		{"unchanged role with value", func(o *analysis.ResolutionEvidence, c *[]analysis.ResolutionChoice, a *ResolutionAssessment) {
			v := "events"
			a.InputBindings[2].ResolvedValue = &v
		}},
		{"case differs", func(o *analysis.ResolutionEvidence, c *[]analysis.ResolutionChoice, a *ResolutionAssessment) {
			v := "Events"
			a.InputBindings[0].ResolvedValue = &v
		}},
		{"another role choice", func(o *analysis.ResolutionEvidence, c *[]analysis.ResolutionChoice, a *ResolutionAssessment) {
			*c = append(*c, analysis.ResolutionChoice{Placeholder: "$right", Kind: "dataset", Value: "users"})
			v := "users"
			a.InputBindings[0].ResolvedValue = &v
		}},
		{"expected mismatch", func(o *analysis.ResolutionEvidence, c *[]analysis.ResolutionChoice, a *ResolutionAssessment) {
			a.InputBindings[0].Expected.Name = "users"
		}},
		{"contradictory object across values", func(o *analysis.ResolutionEvidence, c *[]analysis.ResolutionChoice, a *ResolutionAssessment) {
			v := "users"
			*c = append(*c, analysis.ResolutionChoice{Placeholder: "$left", Kind: "dataset", Value: v})
			b := a.InputBindings[0]
			b.ResolvedValue = &v
			b.Expected.Name = v
			a.InputBindings = append(a.InputBindings, b)
		}},
		{"schema mismatch", func(o *analysis.ResolutionEvidence, c *[]analysis.ResolutionChoice, a *ResolutionAssessment) {
			a.InputBindings[0].SchemaID = "missing"
		}},
		{"object mismatch", func(o *analysis.ResolutionEvidence, c *[]analysis.ResolutionChoice, a *ResolutionAssessment) {
			a.InputBindings[0].ObjectID = "users"
		}},
		{"role membership uses ids", func(o *analysis.ResolutionEvidence, c *[]analysis.ResolutionChoice, a *ResolutionAssessment) {
			o.Placeholders[0].OriginalInputIDs = []string{"other"}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, o, c, a := resolutionBindingsFixture(t)
			tc.mutate(&o, &c, &a)
			if err := p.ValidateResolutionBindings(o, c, a); err == nil {
				t.Fatal("expected rejection")
			}
		})
	}
}

func TestResolutionBindingsAbsenceAndPartialDiscovery(t *testing.T) {
	p, o, c, a := resolutionBindingsFixture(t)
	a.InputBindings[0].ObjectID = "absent"
	a.InputBindings[0].SchemaID = ""
	if err := p.ValidateResolutionBindings(o, c, a); err != nil {
		t.Fatal(err)
	}
	a.InputBindings[0].OriginalInputID = "undiscovered"
	o.Coverage.State = "partial"
	if err := p.ValidateResolutionBindings(o, c, a); err != nil {
		t.Fatal(err)
	}
	a.InputBindings = []ResolutionBinding{}
	if err := p.ValidateResolutionBindings(o, c, a); err != nil {
		t.Fatal("missing bindings must remain assessable uncertainty", err)
	}
}

func TestResolutionBindingsConcurrentDetachedArtifacts(t *testing.T) {
	p, o, c, a := resolutionBindingsFixture(t)
	report := p.env.Report()
	id := p.ArtifactIdentity()
	if id.EnvironmentDigest != report.SnapshotDigest || id.SchemaBundleDigest != report.SchemaBundleDigest {
		t.Fatal("artifact digests diverged")
	}
	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				local := detach(a)
				if err := p.ValidateResolutionBindings(o, c, local); err != nil {
					t.Error(err)
				}
				local.InputBindings[0].Expected.Name = "mutated"
				*local.InputBindings[0].ResolvedValue = "mutated"
				if p.ArtifactIdentity() != id {
					t.Error("prepared identity changed")
				}
			}
		}()
	}
	wg.Wait()
}

func TestResolutionBindingsCanonicalOriginalRoles(t *testing.T) {
	r := assessmentFixture(t, "from $events | fields id")
	p, err := Prepare(r.Snapshot, r.SchemaBundle)
	if err != nil {
		t.Fatal(err)
	}
	session, err := analysis.PrepareResolution(analysis.QueryDocument{Text: "from $events | fields id", Language: "spl2"})
	if err != nil {
		t.Fatal(err)
	}
	original := session.Evidence()
	if len(original.Placeholders) != 1 || len(original.Placeholders[0].OriginalInputIDs) != 1 {
		t.Fatalf("missing original role: %+v", original.Placeholders)
	}
	value := "events"
	binding := ResolutionBinding{OriginalInputID: original.Placeholders[0].OriginalInputIDs[0], ResolvedValue: &value, ObjectID: "events", Expected: objectIdentity(r.Snapshot.Objects[0]), SchemaID: "fields"}
	assessment := ResolutionAssessment{QueryScope: r.QueryScope, InputBindings: []ResolutionBinding{binding}}
	choices := []analysis.ResolutionChoice{{Placeholder: "$events", Kind: "dataset", Value: value}}
	if err := p.ValidateResolutionBindings(original, choices, assessment); err != nil {
		t.Fatal(err)
	}
	// Admission never mutates caller-owned role evidence or binding selections.
	if original.Placeholders[0].OriginalInputIDs[0] != binding.OriginalInputID || *assessment.InputBindings[0].ResolvedValue != "events" {
		t.Fatal("admission mutated call state")
	}
}

func TestResolutionBindingsBoundSchemaAbsentObject(t *testing.T) {
	r := assessmentFixture(t, "from $events | fields id")
	r.Snapshot.Objects = []environment.Object{}
	p, err := Prepare(r.Snapshot, r.SchemaBundle)
	if err != nil {
		t.Fatal(err)
	}
	value := "events"
	original := analysis.ResolutionEvidence{Analysis: analysis.Result{Inputs: []analysis.QueryInput{{ID: "role", Kind: "named_placeholder"}}}, Coverage: analysis.InputCoverage{State: "complete"}, Placeholders: []analysis.ResolutionPlaceholder{{Placeholder: "$events", Kind: "dataset", OriginalInputIDs: []string{"role"}}}}
	a := ResolutionAssessment{QueryScope: r.QueryScope, InputBindings: []ResolutionBinding{{OriginalInputID: "role", ResolvedValue: &value, ObjectID: "events", Expected: r.InputBindings[0].Expected, SchemaID: "fields"}}}
	r.SchemaBundle.Schemas[0].Catalog[0] = 'X'
	r.SchemaBundle.Bindings[0].Expected.Name = "caller"
	if err := p.ValidateResolutionBindings(original, []analysis.ResolutionChoice{{Placeholder: "$events", Kind: "dataset", Value: value}}, a); err != nil {
		t.Fatal("absent captured facts remain admissible with detached linked schema", err)
	}
}
