package compatibility

import (
	"errors"
	"fmt"
	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/closure"
	"github.com/delgado-jacob/spl-toolkit/pkg/environment"
	"reflect"
	"sync"
	"testing"
)

func closureFixture(t *testing.T) Request {
	r := assessmentFixture(t, "from $events | fields id")
	doc := analysis.QueryDocument{Text: "`source`", Language: "spl", SourceID: "root.spl"}
	set, err := analysis.Requirements(doc)
	if err != nil {
		t.Fatal(err)
	}
	r.Requirements = *set
	r.Document = &doc
	r.Snapshot.Capabilities = append(r.Snapshot.Capabilities, environment.Capability{ID: "language:" + set.Query.Language + ":profile:" + set.Query.Profile, Version: set.Query.Version, State: "available", Provenance: r.Snapshot.Objects[0].Provenance})
	arity := 0
	no := false
	r.Snapshot.Collections = append(r.Snapshot.Collections, environment.Collection{Kind: "macro", Coverage: "complete"})
	r.Snapshot.Objects = append(r.Snapshot.Objects, environment.Object{ID: "macro", Kind: "macro", Name: "source", Namespace: "search", App: "app", Owner: "nobody", Provenance: r.Snapshot.Objects[0].Provenance, Arity: &arity, EvalBased: &no, Arguments: []string{}, Document: &analysis.QueryDocument{Text: "| makeresults", Language: "spl", SourceID: "macro.spl"}})
	property := "/dependency"
	r.Snapshot.Objects[len(r.Snapshot.Objects)-1].Relations = []closure.Relation{{Kind: "saved_search", Name: "Daily", Property: &property}}
	r.Snapshot.Collections = append(r.Snapshot.Collections, environment.Collection{Kind: "saved_search", Coverage: "complete"})
	r.Snapshot.Objects = append(r.Snapshot.Objects, environment.Object{ID: "daily", Kind: "saved_search", Name: "Daily", Namespace: "search", App: "app", Owner: "nobody", Provenance: r.Snapshot.Objects[0].Provenance, Document: &analysis.QueryDocument{Text: "from $events | fields id", Language: "spl2", SourceID: "daily.spl"}})
	return r
}

func TestClosureHiddenBindingAdmission(t *testing.T) {
	r := closureFixture(t)
	prepared, err := Prepare(r.Snapshot, r.SchemaBundle)
	if err != nil {
		t.Fatal(err)
	}
	for name, check := range map[string]func() (*Report, error){"Check": func() (*Report, error) { return Check(r) }, "JSON": func() (*Report, error) { return CheckJSON(requestRaw(t, r)) }, "Prepared": func() (*Report, error) { return prepared.Check(r.assessment()) }} {
		t.Run(name, func(t *testing.T) {
			report, err := check()
			if err != nil {
				t.Fatal(err)
			}
			if report.Outcome != "incomplete" || hasReason(report, "dependency_closure_incomplete") || len(report.Inputs) != 1 || report.Inputs[0].Outcome != "satisfied" {
				t.Fatalf("%s", assessmentSummary(report))
			}
			if report.Closure == nil || report.EffectiveRequirements == nil {
				t.Fatal("closure evidence absent")
			}
		})
	}
	r.InputBindings = []InputBinding{}
	_, err = Check(r)
	requireRequestError(t, err, "missing_input_binding", "/input_bindings")
}

func TestClosureDocumentFreshness(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(*Request)
	}{
		{"text", func(r *Request) { r.Document.Text += " | fields other" }},
		{"source", func(r *Request) { r.Document.SourceID = "other" }},
		{"dialect", func(r *Request) { r.Document.Language = "spl2" }},
		{"full evidence", func(r *Request) { r.Requirements.Correlation.Coverage.State = "not_applicable" }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := closureFixture(t)
			tc.mutate(&r)
			_, err := Check(r)
			if err == nil {
				t.Fatal("stale/inconsistent requirements admitted")
			}
		})
	}
}

func definitionFixture(t *testing.T) Request {
	r := assessmentFixture(t, "from $events | fields id")
	doc := analysis.QueryDocument{Text: "from view | fields id", Language: "spl2", SourceID: "root.spl"}
	set, err := analysis.Requirements(doc)
	if err != nil {
		t.Fatal(err)
	}
	r.Requirements = *set
	r.Document = &doc
	object := environment.Object{ID: "view", Kind: "dataset", Name: "view", Namespace: "search", App: "app", Owner: "nobody", Provenance: r.Snapshot.Objects[0].Provenance, Document: &analysis.QueryDocument{Text: "from $events | fields id", Language: "spl2", SourceID: "shared.spl"}}
	r.Snapshot.Objects = append(r.Snapshot.Objects, object)
	r.InputBindings = append(r.InputBindings, InputBinding{InputID: set.Inputs[0].ID, ObjectID: object.ID, Expected: objectIdentity(object), SchemaID: "fields"})
	r.SchemaBundle.Bindings = append(r.SchemaBundle.Bindings, environment.SchemaBinding{SchemaID: "fields", ObjectID: object.ID, Expected: objectIdentity(object), SourceCoverage: "complete"})
	return r
}

func TestClosureDefinitionEffectivePositive(t *testing.T) {
	r := definitionFixture(t)
	report := checked(t, r)
	if report.Outcome != "satisfied" || len(report.Inputs) != 2 {
		t.Fatalf("outcome=%s inputs=%+v reasons=%+v", report.Outcome, report.Inputs, report.Reasons)
	}
	if !reflect.DeepEqual(report.Requirements, r.Requirements) {
		t.Fatal("direct requirements changed")
	}
	if report.Closure.EffectiveAnalysis.Document.Text != r.Document.Text {
		t.Fatal("Dataset placeholder substitution occurred")
	}
	for _, def := range report.Closure.DefinitionAnalyses {
		if def.EffectiveAnalysis != nil && def.EffectiveAnalysis.Document.Text != "from $events | fields id" {
			t.Fatal("definition placeholder substituted")
		}
	}
	hiddenFields := 0
	for _, out := range report.RequirementOutcomes {
		if out.DefinitionObjectID == "view" && out.FieldProjection != nil {
			hiddenFields++
			if out.Outcome != "satisfied" || len(out.SourceIntervals) != 1 || out.SourceIntervals[0].ObjectID != "view" || out.SourceIntervals[0].Start != 22 || out.SourceIntervals[0].End != 24 || len(out.InvocationProvenance) != 1 {
				t.Fatalf("hidden field provenance=%+v", out)
			}
		}
	}
	if hiddenFields != 1 {
		t.Fatal("hidden field obligation absent")
	}
}

func TestClosureLeafAndSelectedBodyBoundary(t *testing.T) {
	for _, withDocument := range []bool{false, true} {
		for _, withBody := range []bool{false, true} {
			t.Run(fmt.Sprintf("doc=%v/body=%v", withDocument, withBody), func(t *testing.T) {
				r := assessmentFixture(t, "from $events | fields id")
				if withDocument {
					r.Document = &analysis.QueryDocument{Text: "from $events | fields id", Language: "spl2"}
				}
				if withBody {
					r.Snapshot.Objects[0].Document = &analysis.QueryDocument{Text: "from $secret | fields missing", Language: "spl2"}
				}
				report := checked(t, r)
				want := "satisfied"
				if withBody {
					want = "incomplete"
				}
				if report.Outcome != want || fieldOutcome(t, report, "id").Outcome != "satisfied" || hasReason(report, "dependency_closure_incomplete") != withBody {
					t.Fatalf("outcome=%s reasons=%+v", report.Outcome, report.Reasons)
				}
				if withDocument && len(report.Closure.Traversal) != 0 {
					t.Fatal("placeholder became knowledge-object name")
				}
				for _, coverage := range report.Coverage {
					if coverage.Dimension == "dependency_closure" && !withBody && coverage.State != "not_applicable" {
						t.Fatalf("leaf closure state=%s", coverage.State)
					}
				}
			})
		}
	}
	r := definitionFixture(t)
	r.Document = nil
	r.InputBindings = []InputBinding{}
	report := checked(t, r)
	if report.Outcome != "incomplete" || !hasReason(report, "dependency_closure_incomplete") {
		t.Fatal("definition existence proved hidden body without document")
	}
}

func TestClosureMissingDirectAndHiddenBindings(t *testing.T) {
	r := definitionFixture(t)
	doc := analysis.QueryDocument{Text: "from $direct | union [from view]", Language: "spl2"}
	set, err := analysis.Requirements(doc)
	if err != nil {
		t.Fatal(err)
	}
	r.Requirements = *set
	r.Document = &doc
	hidden := r.InputBindings[0]
	direct := hidden
	direct.InputID = set.Inputs[0].ID
	r.InputBindings = []InputBinding{hidden}
	_, err = Check(r)
	requireRequestError(t, err, "missing_input_binding", "/input_bindings")
	r.InputBindings = []InputBinding{direct}
	_, err = Check(r)
	requireRequestError(t, err, "missing_input_binding", "/input_bindings")
	r.InputBindings = []InputBinding{direct, hidden}
	if _, err = Check(r); err != nil {
		t.Fatal(err)
	}
}

func TestClosureIncompleteKnownFacts(t *testing.T) {
	for _, kind := range []string{"opaque", "cycle", "scoped absence"} {
		t.Run(kind, func(t *testing.T) {
			r := definitionFixture(t)
			switch kind {
			case "opaque":
				// A macro's opaque body has no invented hidden input obligation.
				r = closureFixture(t)
				r.Snapshot.Objects[2].Document = nil
				r.Snapshot.Objects[2].Relations = nil
				r.InputBindings = []InputBinding{}
			case "cycle":
				r.Snapshot.Objects[2].Document = &analysis.QueryDocument{Text: "from view | fields id", Language: "spl2"}
				r.InputBindings = []InputBinding{}
			case "scoped absence":
				r.Snapshot.Objects[2].App = "elsewhere"
				r.QueryScope.App = environment.Selector{Values: []string{"app"}}
				r.SchemaBundle.Bindings = r.SchemaBundle.Bindings[:1]
				r.InputBindings = []InputBinding{}
			}
			report := checked(t, r)
			if !hasReason(report, "dependency_closure_incomplete") {
				t.Fatalf("closure gap lost: %+v", report.Reasons)
			}
			if kind == "scoped absence" {
				if report.Outcome != "unsatisfied" {
					t.Fatal("missing direct dependency not retained")
				}
			} else if report.Outcome != "incomplete" {
				t.Fatalf("got %s", report.Outcome)
			}
			if report.Closure == nil || len(report.Closure.Gaps) == 0 {
				t.Fatal("raw closure gaps absent")
			}
		})
	}
}

func TestClosureSharedQueryIdentityAndInputCoalescing(t *testing.T) {
	r := definitionFixture(t)
	other := detach(r.Snapshot.Objects[2])
	other.ID = "other"
	other.Name = "other"
	r.Snapshot.Objects = append(r.Snapshot.Objects, other)
	r.SchemaBundle.Bindings = append(r.SchemaBundle.Bindings, environment.SchemaBinding{SchemaID: "fields", ObjectID: other.ID, Expected: objectIdentity(other), SourceCoverage: "complete"})
	doc := analysis.QueryDocument{Text: "from view | union [from other]", Language: "spl2", SourceID: "root.spl"}
	set, err := analysis.Requirements(doc)
	if err != nil {
		t.Fatal(err)
	}
	r.Requirements = *set
	r.Document = &doc
	report := checked(t, r)
	ids := map[string]bool{}
	for _, out := range report.RequirementOutcomes {
		if out.RequirementID == "req-1" && out.DefinitionObjectID != "" {
			ids[out.DefinitionObjectID] = true
			if out.Outcome != "satisfied" {
				t.Fatalf("hidden outcome=%+v", out)
			}
		}
	}
	if !ids["view"] || !ids["other"] {
		t.Fatalf("local req-1 collided: %+v", ids)
	}
	count := 0
	for _, input := range report.Inputs {
		if input.InputID == r.InputBindings[0].InputID {
			count++
			if len(input.Occurrences) != 2 {
				t.Fatalf("contextual input occurrences=%+v", input.Occurrences)
			}
		}
	}
	if count != 1 {
		t.Fatal("logical placeholder did not coalesce")
	}
}

func TestClosureLiteralSelectionConsistency(t *testing.T) {
	r := assessmentFixture(t, "from $events | fields id")
	doc := analysis.QueryDocument{Text: "from events | fields id", Language: "spl2"}
	set, err := analysis.Requirements(doc)
	if err != nil {
		t.Fatal(err)
	}
	r.Requirements = *set
	r.Document = &doc
	r.InputBindings[0].InputID = set.Inputs[0].ID
	other := detach(r.Snapshot.Objects[0])
	other.ID = "other"
	other.App = "other"
	r.Snapshot.Objects = append(r.Snapshot.Objects, other)
	report := checked(t, r)
	if report.Outcome != "satisfied" {
		t.Fatalf("selected leaf demoted: %+v", report.Reasons)
	}
	r.DependencyBindings = []closure.Binding{{DocumentDigest: set.Query.QueryDigest, Kind: "dataset", Start: 5, End: 11, ObjectID: other.ID}}
	_, err = Check(r)
	requireRequestError(t, err, "binding_invalid", "/dependency_bindings")
}

func TestClosureDependencyIdentityAndAdmission(t *testing.T) {
	r := closureFixture(t)
	r.InputBindings = []InputBinding{}
	// Give both definitions source-free literal bodies; uncertainty is retained,
	// while the binding itself determines one exact dependency selection.
	r.Snapshot.Objects[2].Relations = nil
	other := detach(r.Snapshot.Objects[2])
	other.ID = "other"
	other.App = "other"
	r.Snapshot.Objects = append(r.Snapshot.Objects, other)
	binding := closure.Binding{DocumentDigest: r.Requirements.Query.QueryDigest, Kind: "macro", Start: 0, End: 8, ObjectID: "macro"}
	r.DependencyBindings = []closure.Binding{binding}
	first := checked(t, r)
	r.DependencyBindings[0].ObjectID = "other"
	second := checked(t, r)
	if first.Provenance.AssessmentIdentityDigest == second.Provenance.AssessmentIdentityDigest {
		t.Fatal("dependency selection omitted from identity")
	}
	for _, report := range []*Report{first, second} {
		for _, out := range report.RequirementOutcomes {
			if out.Query == r.Requirements.Query && out.RequirementID == "req-1" && out.Outcome != "satisfied" {
				t.Fatalf("selected macro existence still ambiguous: %+v", out)
			}
		}
	}
	r.DependencyBindings[0].Start = 1
	for name, check := range map[string]func() (*Report, error){"Check": func() (*Report, error) { return Check(r) }, "JSON": func() (*Report, error) { return CheckJSON(requestRaw(t, r)) }} {
		t.Run(name, func(t *testing.T) {
			report, err := check()
			requireRequestError(t, err, "dependency_closure_invalid", "")
			var cause *closure.InputError
			if report != nil || !errors.As(err, &cause) {
				t.Fatalf("closure admission cause lost: %v", err)
			}
			detail, _ := RequestErrorDetails(err)
			if detail.Message != cause.Error() {
				t.Fatal("owner message changed")
			}
		})
	}

}

func TestClosureSelectedLeafPartialPositive(t *testing.T) {
	r := assessmentFixture(t, "from $events | fields id")
	doc := analysis.QueryDocument{Text: "from events | fields id", Language: "spl2"}
	set, err := analysis.Requirements(doc)
	if err != nil {
		t.Fatal(err)
	}
	r.Requirements = *set
	r.Document = &doc
	r.InputBindings[0].InputID = set.Inputs[0].ID
	r.Snapshot.Collections[0].Coverage = "partial"
	r.Snapshot.Collections[0].Reason = "bounded collection"
	report := checked(t, r)
	if report.Outcome != "satisfied" {
		t.Fatalf("positive leaf demoted: %+v", report.Reasons)
	}
	for _, coverage := range report.Coverage {
		if coverage.Dimension == "dependency_closure" && coverage.State != "not_applicable" {
			t.Fatalf("leaf is not an expansion obligation: %+v", coverage)
		}
	}
	if report.Closure.Coverage.Collections {
		t.Fatal("raw closure partial collection evidence was erased")
	}
}

func TestClosureEffectiveFormalAndRepeatedInstanceProvenance(t *testing.T) {
	r := closureFixture(t)
	doc := analysis.QueryDocument{Text: "`source(1)` | `source(2)`", Language: "spl", SourceID: "root.spl"}
	set, err := analysis.Requirements(doc)
	if err != nil {
		t.Fatal(err)
	}
	r.Requirements = *set
	r.Document = &doc
	arity := 1
	r.Snapshot.Objects[2].Arity = &arity
	r.Snapshot.Objects[2].Arguments = []string{"events"}
	r.Snapshot.Objects[2].Document.Text = "eval local=$events$"
	report := checked(t, r)
	if report.Closure.EffectiveAnalysis.Document.Text != "eval local=1 | eval local=2" {
		t.Fatalf("effective=%q", report.Closure.EffectiveAnalysis.Document.Text)
	}
	hidden := 0
	instances := map[string]bool{}
	for _, input := range report.Inputs {
		if input.InputID == r.InputBindings[0].InputID {
			hidden++
			for _, occurrence := range input.Occurrences {
				if occurrence.DefinitionObjectID != "daily" {
					t.Fatalf("formal token became hidden source: %+v", occurrence)
				}
				if len(occurrence.InvocationProvenance) != 2 || len(occurrence.InvocationProvenance[0].Invocation) != 1 || len(occurrence.InvocationProvenance[1].Invocation) != 0 {
					t.Fatalf("invocation chain or fake property range: %+v", occurrence.InvocationProvenance)
				}
				instances[occurrence.TraversalEdgeID] = true
			}
		}
	}
	if hidden != 1 || len(instances) != 2 {
		t.Fatalf("hidden=%d invocation contexts=%+v", hidden, instances)
	}
	if hasReason(report, "dependency_closure_incomplete") {
		t.Fatal("known contextual body was demoted by formal token analysis")
	}
	// The real typed Dataset placeholder shares the formal name and still needs
	// its one source binding. No spelling-based formal-name filter is permitted.
	r.InputBindings = []InputBinding{}
	_, err = Check(r)
	requireRequestError(t, err, "missing_input_binding", "/input_bindings")
}

func TestClosurePreparedReuseAndIsolation(t *testing.T) {
	r := definitionFixture(t)
	prepared, err := Prepare(r.Snapshot, r.SchemaBundle)
	if err != nil {
		t.Fatal(err)
	}
	want := checked(t, r)
	first, err := prepared.Check(r.assessment())
	if err != nil {
		t.Fatal(err)
	}
	first.Closure.EffectiveAnalysis.Document.Text = "mutated"
	first.Closure.DefinitionContexts[0].Provenance[0].Source.ObjectID = "mutated"
	first.Closure.DefinitionAnalyses[0].EffectiveAnalysis.Requirements.Inputs[0].Occurrences[0].Location.Start.Offset = 999
	first.Inputs[0].Occurrences[0].SourceIntervals[0].ObjectID = "mutated"
	got, err := prepared.Check(r.assessment())
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(want, got) {
		t.Fatal("closure report mutation leaked into prepared reuse")
	}
	failures := make(chan error, 8)
	var wait sync.WaitGroup
	for i := 0; i < 8; i++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			report, err := prepared.Check(r.assessment())
			if err == nil && !reflect.DeepEqual(want, report) {
				err = fmt.Errorf("concurrent report differs")
			}
			failures <- err
		}()
	}
	wait.Wait()
	close(failures)
	for err := range failures {
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestClosureNormalizedDependencyChoices(t *testing.T) {
	r := definitionFixture(t)
	first := checked(t, r)
	r.DependencyBindings = []closure.Binding{}
	second := checked(t, r)
	if first.Provenance.AssessmentIdentityDigest != second.Provenance.AssessmentIdentityDigest {
		t.Fatal("empty dependency choices are not canonical")
	}
	binding := closure.Binding{DocumentDigest: r.Requirements.Query.QueryDigest, Kind: "dataset", Start: 5, End: 9, ObjectID: "view"}
	r.DependencyBindings = []closure.Binding{binding}
	first = checked(t, r)
	r.DependencyBindings = append(r.DependencyBindings, binding)
	second = checked(t, r)
	if first.Provenance.AssessmentIdentityDigest != second.Provenance.AssessmentIdentityDigest {
		t.Fatal("equivalent duplicate dependency choices are not canonical")
	}
}
