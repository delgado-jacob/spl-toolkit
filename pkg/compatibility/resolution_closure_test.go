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

func TestTask8SpecSharedDigestValidBinding(t *testing.T) {
	text := `from $root | union [from users] | fields id`
	proof, r, a := resolutionCheckFixture(t, text)
	r.Snapshot.Objects[0].Document = &analysis.QueryDocument{Text: text, Language: "spl2", SourceID: "definition"}
	original, _, _, _ := proof.AssessmentEvidence()
	for _, ref := range original.Analysis.References {
		if ref.Kind == "dataset" && ref.NormalizedName == "users" {
			a.DependencyBindings = append(a.DependencyBindings, closure.Binding{DocumentDigest: queryDigest(text), Kind: "dataset", Start: ref.Location.Start.Offset, End: ref.Location.End.Offset, ObjectID: "users"})
		}
	}
	if len(a.DependencyBindings) != 1 {
		t.Fatalf("bindings=%+v", a.DependencyBindings)
	}
	p, err := Prepare(r.Snapshot, r.SchemaBundle)
	if err != nil {
		t.Fatal(err)
	}
	bundle, err := p.env.DefinitionBundle(a.QueryScope)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := closure.Evaluate(closure.Request{SchemaVersion: 1, Document: original.Analysis.Document, Bundle: bundle, Bindings: a.DependencyBindings}); err != nil {
		t.Fatalf("original admission: %v", err)
	}
	report, err := p.CheckResolution(proof, a)
	if err != nil {
		t.Fatalf("valid submitted binding rejected after rendering: %v", err)
	}
	if report == nil {
		t.Fatal("nil report")
	}
	if !slices.Contains(report.EffectiveDependencyBindings, a.DependencyBindings[0]) {
		t.Fatal("immutable definition binding lost")
	}
	_, candidate, _, _ := proof.AssessmentEvidence()
	translated := false
	for _, binding := range report.EffectiveDependencyBindings {
		if binding.ObjectID == "users" && binding.DocumentDigest == queryDigest(candidate.Analysis.Document.Text) && candidate.Analysis.Document.Text[binding.Start:binding.End] == "users" {
			translated = true
		}
	}
	if !translated {
		t.Fatal("translated root binding lost")
	}
	for _, def := range report.Closure.DefinitionAnalyses {
		if def.ObjectID == "events" && def.DirectAnalysis.Document.Text != text {
			t.Fatal("immutable definition rewritten")
		}
	}
	if !reflect.DeepEqual(report.DependencyBindings, a.DependencyBindings) {
		t.Fatal("submitted bindings changed")
	}
	// The public selection gate also admits original dependency coordinates before rendering.
	if err := p.ValidateResolutionBindings(original, []analysis.ResolutionChoice{{Placeholder: "$root", Kind: "dataset", Value: "events"}}, a); err != nil {
		t.Fatal(err)
	}
	a.DependencyBindings[0].DocumentDigest = queryDigest("stale")
	if err := p.ValidateResolutionBindings(original, []analysis.ResolutionChoice{{Placeholder: "$root", Kind: "dataset", Value: "events"}}, a); err == nil {
		t.Fatal("preflight admitted stale dependency binding")
	}

}

func TestResolutionClosureRootMacroRoleAndArgumentEnvelope(t *testing.T) {
	doc := analysis.QueryDocument{Text: "search index=\"$root\" | `pass(id,id)` | fields id", Language: "spl", SourceID: "root"}
	session, err := analysis.PrepareResolution(doc)
	if err != nil {
		t.Fatal(err)
	}
	choices := []analysis.ResolutionChoice{{Placeholder: "$root", Kind: "index", Value: "main"}}
	rendered, err := session.Render(choices)
	if err != nil {
		t.Fatal(err)
	}
	proof, err := session.Verify(rendered)
	if err != nil || !proof.Evidence().Proven {
		t.Fatalf("proof=%+v error=%v", proof, err)
	}
	original, candidate, roles, _ := proof.AssessmentEvidence()
	r := closureFixture(t)
	macroIndex := slices.IndexFunc(r.Snapshot.Objects, func(object environment.Object) bool { return object.Kind == "macro" })
	r.Snapshot.Objects = []environment.Object{r.Snapshot.Objects[0], r.Snapshot.Objects[macroIndex]}
	r.Snapshot.Objects[1].Name = "pass"
	r.Snapshot.Objects[1].Relations = nil
	arity := 2
	r.Snapshot.Objects[1].Arity = &arity
	r.Snapshot.Objects[1].Arguments = []string{"a", "b"}
	body := analysis.QueryDocument{Text: "fields $a$, $b$", Language: "spl", SourceID: "macro"}
	r.Snapshot.Objects[1].Document = &body

	r.Snapshot.Objects = append(r.Snapshot.Objects, environment.Object{ID: "main", Kind: "index", Name: "main", Provenance: r.Snapshot.Objects[0].Provenance})
	a := ResolutionAssessment{QueryScope: r.QueryScope, InputBindings: []ResolutionBinding{{OriginalInputID: original.Analysis.Inputs[0].ID, ObjectID: "events", Expected: objectIdentity(r.Snapshot.Objects[0]), SchemaID: "fields"}}}
	start := strings.Index(doc.Text, "`pass")
	a.DependencyBindings = []closure.Binding{{DocumentDigest: queryDigest(doc.Text), Kind: "macro", Start: start, End: start + len("`pass(id,id)`"), ObjectID: "macro"}}
	report := checkedResolution(t, proof, r, a)
	if report.Closure.EffectiveAnalysis.Document.Text == candidate.Analysis.Document.Text || !strings.Contains(report.Closure.EffectiveAnalysis.Document.Text, "fields id, id") {
		t.Fatal("root macro did not expand")
	}
	if report.Outcome != "incomplete" {
		t.Fatalf("implicit source coverage lost: %s", report.Outcome)
	}
	effectiveStart := strings.Index(candidate.Analysis.Document.Text, "`pass")
	translated := closure.Binding{DocumentDigest: queryDigest(candidate.Analysis.Document.Text), Kind: "macro", Start: effectiveStart, End: effectiveStart + len("`pass(id,id)`"), ObjectID: "macro"}
	if !slices.Contains(report.EffectiveDependencyBindings, translated) || effectiveStart == start {
		t.Fatalf("macro invocation not exactly translated: %+v", report.EffectiveDependencyBindings)
	}
	if report.Closure.DefinitionAnalyses[0].DirectAnalysis.Document.Text != body.Text || report.Closure.DefinitionAnalyses[0].DirectAnalysis.Requirements.Query.QueryDigest != queryDigest(body.Text) {
		t.Fatal("macro definition bytes changed")
	}
	outside := false
	definition := false
	rootInput := false
	for _, out := range report.Inputs {
		if out.OriginalInputID == roles[0].OriginalInput.ID {
			rootInput = true
			if len(out.Occurrences) != 1 || len(out.Evidence.Occurrences) != 1 || len(out.Evidence.Occurrences[0].SourceIntervals) != 1 || out.Evidence.Occurrences[0].SourceIntervals[0].Kind != "query" {
				t.Fatalf("root source role provenance lost: %+v", out)
			}
		}
	}
	for _, out := range report.RequirementOutcomes {
		if out.OriginalInputID == roles[0].OriginalInput.ID && len(out.AssessedOccurrences) > 0 && out.AssessedOccurrences[0].OriginalName == "id" {
			outside = true
			if len(out.Evidence.SourceIntervals) != 1 || out.Evidence.SourceIntervals[0].Kind != "query" || out.Evidence.SourceIntervals[0].Start != strings.LastIndex(candidate.Analysis.Document.Text, "id") || out.AssessedOccurrences[0].Necessity != roles[0].Requirements[0].OriginalOccurrence.Necessity {
				t.Fatalf("outside field mapping lost: %+v", out)
			}
			// Evidence keeps the original conditional necessity despite the stronger expansion.
			if out.Evidence.Applicability == "applicable" {
				t.Fatalf("original necessity upgraded: %+v", out)
			}
		}
		for _, source := range out.Evidence.SourceIntervals {
			if source.Kind == "definition" && source.ObjectID == "macro" {
				definition = true
				if out.OriginalInputID != "" || out.Evidence.Outcome == "satisfied" || len(out.Evidence.Schemas) > 0 || len(out.Evidence.InvocationProvenance) == 0 {
					t.Fatalf("definition field borrowed root schema: %+v", out)
				}
			}
		}
	}
	if !outside || !definition || !rootInput {
		t.Fatalf("outside=%v definition=%v root=%v outcomes=%+v", outside, definition, rootInput, report.RequirementOutcomes)
	}
	// Every effective occurrence remains represented, independent of grouping.
	for _, item := range report.EffectiveRequirements.Items {
		for _, occ := range item.Occurrences {
			if !slices.ContainsFunc(report.RequirementOutcomes, func(out ResolutionRequirementOutcome) bool {
				return out.Evidence.Query == report.EffectiveRequirements.Query && out.CandidateRequirementID == item.ID && slices.ContainsFunc(out.AssessedOccurrences, func(assessed analysis.RequirementOccurrence) bool { return reflect.DeepEqual(assessed, occ) })
			}) {
				t.Fatalf("effective obligation omitted: %s %+v", item.ID, occ)
			}
		}
	}
}
