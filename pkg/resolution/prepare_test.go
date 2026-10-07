package resolution

import (
	"sync"
	"testing"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/compatibility"
)

func TestResolutionPrepareDetachedArtifacts(t *testing.T) {
	r := requestFixture(t)
	p, err := Prepare(r.Compatibility.Snapshot, r.Compatibility.SchemaBundle)
	if err != nil {
		t.Fatal(err)
	}
	owner, err := compatibility.Prepare(r.Compatibility.Snapshot, r.Compatibility.SchemaBundle)
	if err != nil {
		t.Fatal(err)
	}
	want := owner.ArtifactIdentity()
	binding := detach(r.Compatibility.InputBindings[0])
	original := analysis.ResolutionEvidence{Analysis: analysis.Result{Inputs: []analysis.QueryInput{{ID: binding.OriginalInputID, Kind: "named_placeholder"}}}, Coverage: analysis.InputCoverage{State: "complete"}, Placeholders: []analysis.ResolutionPlaceholder{{Placeholder: "$events", Kind: "dataset", OriginalInputIDs: []string{binding.OriginalInputID}}}}
	choices := []analysis.ResolutionChoice{{Placeholder: "$events", Kind: "dataset", Value: *binding.ResolvedValue}}
	assessment := compatibility.ResolutionAssessment{QueryScope: detach(r.Compatibility.QueryScope), InputBindings: []compatibility.ResolutionBinding{binding}}
	if p.identity != want {
		t.Fatal("artifact owners' digests must be retained")
	}
	r.Compatibility.Snapshot.Collections[0].Coverage = "caller"
	r.Compatibility.Snapshot.Objects[0].Name = "caller"
	if r.Compatibility.SchemaBundle != nil {
		r.Compatibility.SchemaBundle.Schemas[0].Catalog[0] = 'X'
		r.Compatibility.SchemaBundle.Bindings[0].Expected.Name = "caller"
	}
	if err := p.compatibility.ValidateResolutionBindings(original, choices, assessment); err != nil {
		t.Fatal("prepared artifacts alias caller", err)
	}
	if p.compatibility.ArtifactIdentity() != want {
		t.Fatal("prepared artifacts alias caller")
	}
}

func TestResolutionPrepareRequestNormalization(t *testing.T) {
	r := requestFixture(t)
	input := PreparedRequest{SchemaVersion: r.SchemaVersion, Document: r.Document, Resolutions: []Resolution{{Placeholder: "$events", Kind: "dataset", Values: []string{"events"}}}, MaxVariants: uint64Pointer(2), Compatibility: compatibility.ResolutionAssessment{QueryScope: r.Compatibility.QueryScope, InputBindings: r.Compatibility.InputBindings, DependencyBindings: r.Compatibility.DependencyBindings}}
	out, err := normalizePreparedRequest(input)
	if err != nil {
		t.Fatal(err)
	}
	out.Resolutions[0].Values[0] = "changed"
	out.Compatibility.InputBindings[0].Expected.Name = "changed"
	*out.Compatibility.InputBindings[0].ResolvedValue = "changed"
	*out.MaxVariants = 100
	if input.Resolutions[0].Values[0] != "events" || input.Compatibility.InputBindings[0].Expected.Name == "changed" || *input.Compatibility.InputBindings[0].ResolvedValue == "changed" || *input.MaxVariants != 2 {
		t.Fatal("normalized call state aliases caller")
	}
	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			local, err := normalizePreparedRequest(input)
			if err != nil {
				t.Error(err)
				return
			}
			local.Resolutions[0].Values[0] = "local"
			*local.Compatibility.InputBindings[0].ResolvedValue = "local"
		}()
	}
	wg.Wait()
	for _, mutate := range []func(*PreparedRequest){func(r *PreparedRequest) { r.SchemaVersion = 2 }, func(r *PreparedRequest) { r.Resolutions = nil }, func(r *PreparedRequest) { r.Compatibility.InputBindings = nil }, func(r *PreparedRequest) { r.Document.Text = string([]byte{255}) }, func(r *PreparedRequest) { r.Resolutions[0].Values = []string{"events", "events"} }} {
		local := detach(input)
		mutate(&local)
		if _, err := normalizePreparedRequest(local); err == nil {
			t.Fatal("invalid prepared request admitted")
		}
	}
}
