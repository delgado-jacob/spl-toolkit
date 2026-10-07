package compatibility

import (
	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/closure"
	"github.com/delgado-jacob/spl-toolkit/pkg/environment"
	"reflect"
	"slices"
	"strings"
	"testing"
)

func TestResolutionClosureDefinitionMarkerIsolation(t *testing.T) {
	proof, r, a := resolutionCheckFixture(t, `from $root | fields id`)
	r.Snapshot.Objects[0].Document = &analysis.QueryDocument{Text: `from $definition_only | fields id`, Language: "spl2", SourceID: "definition"}
	report := checkedResolution(t, proof, r, a)
	if report.Closure == nil || report.EffectiveRequirements == nil {
		t.Fatal("closure evidence missing")
	}
	if report.Outcome != "incomplete" {
		t.Fatalf("outcome=%s", report.Outcome)
	}
	found := false
	for _, out := range report.RequirementOutcomes {
		if out.Evidence.DefinitionObjectID == "events" && len(out.AssessedOccurrences) > 0 && out.AssessedOccurrences[0].OriginalName == "id" {
			found = true
			if out.OriginalInputID != "" || out.Evidence.Outcome == "satisfied" || len(out.Evidence.InvocationProvenance) == 0 {
				t.Fatalf("definition borrowed root authority: %+v", out)
			}
		}
	}
	if !found {
		t.Fatal("definition field not assessed")
	}
	_, candidate, _, _ := proof.AssessmentEvidence()
	if !reflect.DeepEqual(report.Requirements, candidate.Analysis.Requirements) {
		t.Fatal("candidate changed")
	}
}

func TestResolutionClosureRejectsStaleBindings(t *testing.T) {
	proof, r, a := resolutionCheckFixture(t, `from $root | fields id`)
	for _, binding := range []closure.Binding{
		{DocumentDigest: queryDigest("stale"), Kind: "dataset", Start: 5, End: 11, ObjectID: "events"},
		{DocumentDigest: queryDigest(`from $root | fields id`), Kind: "dataset", Start: 0, End: 4, ObjectID: "events"},
	} {
		a.DependencyBindings = []closure.Binding{binding}
		p, err := Prepare(r.Snapshot, r.SchemaBundle)
		if err != nil {
			t.Fatal(err)
		}
		if report, err := p.CheckResolution(proof, a); err == nil || report != nil {
			t.Fatalf("invalid binding admitted: %+v %v", report, err)
		}
	}
}

func TestResolutionClosureUnavailableDefinition(t *testing.T) {
	proof, r, a := resolutionCheckFixture(t, `from $root | fields id`)
	property := "/dependency"
	r.Snapshot.Objects[0].Relations = []closure.Relation{{Kind: "saved_search", Name: "unavailable", Property: &property}}
	r.Snapshot.Collections = append(r.Snapshot.Collections, environment.Collection{Kind: "saved_search", Coverage: "complete"})
	report := checkedResolution(t, proof, r, a)
	if report.Closure == nil || len(report.Closure.Gaps) == 0 || report.Outcome != "incomplete" {
		t.Fatalf("missing closure gaps: %+v", report)
	}
	for _, out := range report.RequirementOutcomes {
		if out.OriginalInputID != "" && out.Evidence.FieldProjection != nil && out.Evidence.Outcome != "satisfied" {
			t.Fatalf("root lost: %+v", out)
		}
	}
}

func TestResolutionClosureShiftedRootBindingsAndRepeatedDefinitions(t *testing.T) {
	text := `from $root | union [from view] | union [from view] | fields id`
	proof, r, a := resolutionCheckFixture(t, text)
	object := environment.Object{ID: "view", Kind: "dataset", Name: "view", Namespace: "search", App: "app", Owner: "nobody", Provenance: r.Snapshot.Objects[0].Provenance, Document: &analysis.QueryDocument{Text: `from events | fields id`, Language: "spl2", SourceID: "definition"}}
	r.Snapshot.Objects = append(r.Snapshot.Objects, object)
	original, candidate, _, _ := proof.AssessmentEvidence()
	for _, ref := range original.Analysis.References {
		if ref.Kind == "dataset" && ref.NormalizedName == "view" {
			a.DependencyBindings = append(a.DependencyBindings, closure.Binding{DocumentDigest: queryDigest(text), Kind: "dataset", Start: ref.Location.Start.Offset, End: ref.Location.End.Offset, ObjectID: "view"})
		}
	}
	if len(a.DependencyBindings) != 2 {
		t.Fatalf("references=%+v", original.Analysis.References)
	}
	report := checkedResolution(t, proof, r, a)
	if report.Outcome != "incomplete" {
		t.Fatalf("definition fields require separate selection: %s", report.Outcome)
	}
	if !reflect.DeepEqual(report.DependencyBindings, a.DependencyBindings) {
		t.Fatal("submitted binding coordinates changed")
	}
	translated := 0
	for _, binding := range report.EffectiveDependencyBindings {
		if binding.ObjectID == "view" {
			translated++
			if binding.DocumentDigest != queryDigest(candidate.Analysis.Document.Text) || candidate.Analysis.Document.Text[binding.Start:binding.End] != "view" {
				t.Fatalf("wrong translated coordinates: %+v", binding)
			}
		}
	}
	if translated != 2 {
		t.Fatalf("effective bindings=%+v", report.EffectiveDependencyBindings)
	}
	contexts := map[string]bool{}
	for _, out := range report.RequirementOutcomes {
		if out.Evidence.DefinitionObjectID == "view" && len(out.AssessedOccurrences) > 0 && out.AssessedOccurrences[0].OriginalName == "id" {
			if out.OriginalInputID != "" || out.Evidence.Outcome == "satisfied" || len(out.Evidence.InvocationProvenance) == 0 {
				t.Fatalf("definition borrowed root schema: %+v", out)
			}
			contexts[out.Evidence.TraversalEdgeID] = true
		}
	}
	if len(contexts) != 2 {
		t.Fatalf("repeated definition contexts merged: %+v", contexts)
	}
}

func TestResolutionClosureDefinitionBindingAdmission(t *testing.T) {
	proof, r, a := resolutionCheckFixture(t, `from $root | fields id`)
	r.Snapshot.Objects[0].Document = &analysis.QueryDocument{Text: `from users`, Language: "spl2", SourceID: "body"}
	original, _, _, _ := proof.AssessmentEvidence()
	for _, b := range []closure.Binding{
		{DocumentDigest: queryDigest("stale body"), Kind: "dataset", Start: 5, End: 10, ObjectID: "users"},
		{DocumentDigest: queryDigest("from users"), Kind: "dataset", Start: 0, End: 4, ObjectID: "users"},
		{DocumentDigest: queryDigest("from users"), Kind: "dataset", Start: 5, End: 10, ObjectID: "unknown"},
		{DocumentDigest: original.Analysis.Requirements.Query.QueryDigest, Kind: "dataset", Start: 5, End: 10, ObjectID: "users"},
	} {
		a.DependencyBindings = []closure.Binding{b}
		p, err := Prepare(r.Snapshot, r.SchemaBundle)
		if err != nil {
			t.Fatal(err)
		}
		report, err := p.CheckResolution(proof, a)
		if err == nil || report != nil {
			t.Fatalf("invalid definition binding admitted: %+v %v", report, err)
		}
	}
	a.DependencyBindings = []closure.Binding{{DocumentDigest: queryDigest("from users"), Kind: "dataset", Start: 5, End: 10, ObjectID: "users"}}
	report := checkedResolution(t, proof, r, a)
	if !slices.Contains(report.EffectiveDependencyBindings, a.DependencyBindings[0]) {
		t.Fatal("definition binding digest/range rewritten")
	}
}

func TestResolutionClosureCyclePreservesRoot(t *testing.T) {
	proof, r, a := resolutionCheckFixture(t, `from $root | fields id`)
	r.Snapshot.Objects[0].Document = &analysis.QueryDocument{Text: `from events`, Language: "spl2", SourceID: "cycle"}
	report := checkedResolution(t, proof, r, a)
	if report.Outcome != "incomplete" {
		t.Fatalf("cycle=%s", report.Outcome)
	}
	cycle := false
	for _, gap := range report.Closure.Gaps {
		if strings.Contains(gap.Code, "cycle") {
			cycle = true
		}
	}
	if !cycle {
		t.Fatal("cycle gap absent")
	}
	root := false
	for _, out := range report.RequirementOutcomes {
		if out.OriginalInputID != "" && out.Evidence.FieldProjection != nil {
			root = true
			if out.Evidence.Outcome != "satisfied" {
				t.Fatalf("root lost: %+v", out)
			}
		}
	}
	if !root {
		t.Fatal("root field dropped")
	}
}

func TestResolutionClosureMacroDefinitionMarker(t *testing.T) {
	proof, r, a := resolutionCheckFixture(t, `from $root | fields id`)
	r.Snapshot.Objects[0].Document = &analysis.QueryDocument{Text: "`hidden`", Language: "spl", SourceID: "view"}
	arity := 0
	no := false
	r.Snapshot.Collections = append(r.Snapshot.Collections, environment.Collection{Kind: "macro", Coverage: "complete"})
	r.Snapshot.Objects = append(r.Snapshot.Objects, environment.Object{ID: "hidden", Kind: "macro", Name: "hidden", Namespace: "search", App: "app", Owner: "nobody", Provenance: r.Snapshot.Objects[0].Provenance, Arity: &arity, EvalBased: &no, Arguments: []string{}, Document: &analysis.QueryDocument{Text: `from $definition_only | fields id`, Language: "spl2", SourceID: "macro"}})
	report := checkedResolution(t, proof, r, a)
	if report.Closure == nil || report.Outcome != "incomplete" {
		t.Fatalf("macro=%+v", report)
	}
	found := false
	for _, def := range report.Closure.DefinitionAnalyses {
		if def.ObjectID == "hidden" && def.DirectAnalysis != nil {
			found = strings.Contains(def.DirectAnalysis.Document.Text, "$definition_only")
		}
	}
	if !found {
		t.Fatal("definition marker lost")
	}
	for _, out := range report.RequirementOutcomes {
		if out.Evidence.DefinitionObjectID == "hidden" && out.OriginalInputID != "" {
			t.Fatal("definition acquired root role")
		}
	}
}

func TestResolutionClosureMacroInvocationBinding(t *testing.T) {
	r := closureFixture(t)
	// The resolution session also supports a document with no substitution choices.
	doc := analysis.QueryDocument{Text: "search index=main | `source`", Language: "spl", SourceID: "root"}
	session, err := analysis.PrepareResolution(doc)
	if err != nil {
		t.Fatal(err)
	}
	rendered, err := session.Render([]analysis.ResolutionChoice{})
	if err != nil {
		t.Fatal(err)
	}
	proof, err := session.Verify(rendered)
	if err != nil {
		t.Fatal(err)
	}
	if !proof.Evidence().Proven {
		t.Fatalf("proof=%+v", proof.Evidence())
	}
	a := ResolutionAssessment{QueryScope: r.QueryScope, InputBindings: []ResolutionBinding{}, DependencyBindings: []closure.Binding{{DocumentDigest: queryDigest(doc.Text), Kind: "macro", Start: 20, End: 28, ObjectID: "macro"}}}
	report := checkedResolution(t, proof, r, a)
	if !slices.Contains(report.EffectiveDependencyBindings, a.DependencyBindings[0]) {
		t.Fatal("valid macro invocation binding lost")
	}
}
