package environment

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/closure"
)

func TestDefinitionBundleScopeFence(t *testing.T) {
	var snapshot Snapshot
	if err := json.Unmarshal(fixtureRaw(t, snapshotFixture()), &snapshot); err != nil {
		t.Fatal(err)
	}
	selected := func(values ...string) Selector { return Selector{Values: values} }
	snapshot.CaptureScope = CaptureScope{Namespace: selected("search"), App: selected("main"), Owner: selected("nobody")}
	snapshot.Collections = []Collection{
		{Kind: "macro", Coverage: "complete"},
		{Kind: "lookup", Coverage: "complete"},
		{Kind: "index", Coverage: "partial", Reason: "global inventory sampled"},
	}
	zero, one, no, yes, validation, property := 0, 1, false, true, "isnum($value$)", "/dependencies/0"
	macroDocument := analysis.QueryDocument{Text: "lookup users user OUTPUT role\n", Language: "spl", Profile: "splunkd", Version: "current", SourceID: "body.spl"}
	provenance := func(id string) Provenance {
		return Provenance{SourceKind: "rest", SourceID: id, ObservedAt: "2026-10-01T12:02:00Z"}
	}
	snapshot.Objects = []Object{
		{ID: "macro-enrich", Kind: "macro", Name: "enrich", Namespace: "search", App: "main", Owner: "nobody", Sharing: "app", Provenance: provenance("macros.conf:enrich"), Document: &macroDocument, Arity: &zero, Arguments: []string{}, EvalBased: &no, Relations: []closure.Relation{{Kind: "lookup", Name: "users", Property: &property}}},
		{ID: "macro-guarded", Kind: "macro", Name: "guarded", Namespace: "search", App: "main", Owner: "nobody", Sharing: "app", Provenance: provenance("macros.conf:guarded"), Document: &analysis.QueryDocument{Text: "eval x=$value$", Language: "spl", Profile: "splunkd", Version: "current", SourceID: "guarded.spl"}, Arity: &one, Arguments: []string{"value"}, Validation: &validation},
		{ID: "lookup-users", Kind: "lookup", Name: "users", Namespace: "search", App: "main", Owner: "nobody", Sharing: "app", Provenance: provenance("transforms.conf:users")},
		{ID: "index-main", Kind: "index", Name: "main", Provenance: provenance("indexes.conf:main")},
	}
	prepared, report, err := PrepareSnapshot(snapshot)
	if err != nil || prepared == nil || report.Status != "partial" {
		t.Fatalf("prepare snapshot: %v %#v", err, report)
	}
	env, pairReport, err := Pair(prepared, nil)
	if err != nil || env == nil || pairReport.Status != "partial" {
		t.Fatalf("pair snapshot: %v %#v", err, pairReport)
	}
	inScope := snapshot.CaptureScope
	bundle, err := env.DefinitionBundle(inScope)
	if err != nil {
		t.Fatal(err)
	}
	want := closure.DefinitionBundle{SchemaVersion: 1, ScopeID: "capture-a", Collections: []closure.Collection{{Kind: "lookup", Coverage: "complete"}, {Kind: "macro", Coverage: "complete"}}, Objects: []closure.Definition{
		{ID: "lookup-users", Kind: "lookup", Name: "users", App: "main", Owner: "nobody", Sharing: "app", SourceID: "transforms.conf:users"},
		{ID: "macro-enrich", Kind: "macro", Name: "enrich", App: "main", Owner: "nobody", Sharing: "app", SourceID: "macros.conf:enrich", Document: &macroDocument, Arity: &zero, EvalBased: &no, Relations: []closure.Relation{{Kind: "lookup", Name: "users", Property: &property}}},
		{ID: "macro-guarded", Kind: "macro", Name: "guarded", App: "main", Owner: "nobody", Sharing: "app", SourceID: "macros.conf:guarded", Document: &analysis.QueryDocument{Text: "eval x=$value$", Language: "spl", Profile: "splunkd", Version: "current", SourceID: "guarded.spl"}, Arity: &one, Arguments: []string{"value"}, Validation: &validation},
	}}
	if !reflect.DeepEqual(bundle, want) {
		t.Fatalf("definition bundle differs from explicit closure input:\ngot:  %#v\nwant: %#v", bundle, want)
	}
	first, err := closure.Evaluate(closure.Request{SchemaVersion: 1, Document: analysis.QueryDocument{Text: "`enrich` | eval access=role", SourceID: "detection-a", Language: "spl"}, Bundle: bundle, Bindings: []closure.Binding{}})
	if err != nil || first == nil || !first.Coverage.Complete {
		t.Fatalf("in-scope detection closure: %v %#v", err, first)
	}
	second, err := closure.Evaluate(closure.Request{SchemaVersion: 1, Document: analysis.QueryDocument{Text: "lookup users user OUTPUT role", SourceID: "detection-b", Language: "spl"}, Bundle: bundle, Bindings: []closure.Binding{}})
	if err != nil || second == nil || !second.Coverage.Complete {
		t.Fatalf("second detection closure: %v %#v", err, second)
	}
	bundle.Objects[1].Document.Text = "changed"
	again, err := env.DefinitionBundle(inScope)
	if err != nil || !reflect.DeepEqual(again, want) {
		t.Fatalf("returned bundle mutated environment: %v %#v", err, again)
	}
	for name, outScope := range map[string]CaptureScope{
		"namespace": {Namespace: selected("other"), App: selected("main"), Owner: selected("nobody")},
		"app":       {Namespace: selected("search"), App: selected("other"), Owner: selected("nobody")},
		"owner":     {Namespace: selected("search"), App: selected("main"), Owner: selected("alice")},
		"all":       {Namespace: selected("search"), App: Selector{All: &yes}, Owner: selected("nobody")},
	} {
		t.Run(name, func(t *testing.T) {
			outBundle, err := env.DefinitionBundle(outScope)
			if err != nil {
				t.Fatal(err)
			}
			if len(outBundle.Collections) != 2 || outBundle.Collections[0].Coverage != "partial" || outBundle.Collections[1].Coverage != "partial" {
				t.Fatalf("out-of-scope bundle promised complete coverage: %#v", outBundle)
			}
			if name != "all" && len(outBundle.Objects) != 0 {
				t.Fatalf("out-of-scope bundle invented definitions: %#v", outBundle)
			}
			outReport, err := closure.Evaluate(closure.Request{SchemaVersion: 1, Document: analysis.QueryDocument{Text: "`absent`", SourceID: "detection-out", Language: "spl"}, Bundle: outBundle, Bindings: []closure.Binding{}})
			if err != nil || outReport == nil || outReport.Coverage.Complete {
				t.Fatalf("out-of-scope detection claimed complete closure: %v %#v", err, outReport)
			}
		})
	}
}

func TestDefinitionBundleSnapshotVersionParity(t *testing.T) {
	selected := func(value string) Selector { return Selector{Values: []string{value}} }
	scope := CaptureScope{Namespace: selected("search"), App: selected("main"), Owner: selected("nobody")}
	macro := preparedPairSnapshot(t, "first", "eval x=1").Snapshot().Objects[0]
	saved := Object{ID: "saved-daily", Kind: "saved_search", Name: "Daily", Namespace: "search", App: "main", Owner: "nobody", Provenance: macro.Provenance, Document: &analysis.QueryDocument{Text: "`first`", Language: "spl"}}
	var first closure.DefinitionBundle
	firstReports := map[string]*closure.Report{}
	for _, version := range []int{1, 2} {
		t.Run(fmt.Sprintf("v%d", version), func(t *testing.T) {
			value := observedFixture(t)
			value.CaptureScope = scope
			value.Objects = append(value.Objects, macro, saved)
			if version == 1 {
				value.SchemaVersion, value.Observation = 1, nil
				for _, kind := range []string{"index", "source", "sourcetype"} {
					setObservedCollection(&value, kind, "partial", "global inventory sampled")
				}
			}
			prepared, report, err := PrepareSnapshot(value)
			if err != nil || prepared == nil {
				t.Fatalf("prepare snapshot: %v %#v", err, report)
			}
			env, report, err := Pair(prepared, nil)
			if err != nil || env == nil {
				t.Fatalf("pair: %v %#v", err, report)
			}
			bundle, err := env.DefinitionBundle(scope)
			if err != nil || bundle.SchemaVersion != 1 || len(bundle.Objects) != 2 {
				t.Fatalf("definition bundle: %v %#v", err, bundle)
			}
			if version == 1 {
				first = bundle
			} else if !reflect.DeepEqual(bundle, first) {
				t.Fatalf("snapshot version changed definitions:\nv1=%#v\nv2=%#v", first, bundle)
			}
			for _, query := range []string{"`first`", "| from savedsearch:Daily"} {
				result, err := closure.Evaluate(closure.Request{SchemaVersion: 1, Document: analysis.QueryDocument{Text: query, Language: "spl"}, Bundle: bundle, Bindings: []closure.Binding{}})
				if err != nil || result == nil || !result.Coverage.Resolution || !result.Coverage.Collections || !result.Coverage.TraversedDefinitions {
					t.Fatalf("definition resolution for %q: %v %#v", query, err, result)
				}
				if query == "`first`" && !result.Coverage.Complete {
					t.Fatalf("macro closure incomplete: %#v", result.Coverage)
				}
				if version == 1 {
					firstReports[query] = result
				} else if !reflect.DeepEqual(result, firstReports[query]) {
					t.Fatalf("snapshot version changed closure for %q", query)
				}
			}
			outside := scope
			outside.App = selected("other")
			outBundle, err := env.DefinitionBundle(outside)
			if err != nil || len(outBundle.Objects) != 0 {
				t.Fatalf("out-of-scope definitions: %v %#v", err, outBundle)
			}
			for _, collection := range outBundle.Collections {
				if collection.Coverage != "partial" {
					t.Fatalf("out-of-scope collection promised completeness: %#v", collection)
				}
			}
			result, err := closure.Evaluate(closure.Request{SchemaVersion: 1, Document: analysis.QueryDocument{Text: "| from savedsearch:Daily", Language: "spl"}, Bundle: outBundle, Bindings: []closure.Binding{}})
			if err != nil || result == nil || result.Coverage.Collections || result.Coverage.Resolution {
				t.Fatalf("out-of-scope closure claimed collection or resolution coverage: %v %#v", err, result)
			}
		})
	}
}
