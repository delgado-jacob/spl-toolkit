package analysis_test

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
	"testing"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
)

func TestPartialSourceExactProjectionRecoversLocalExpansion(t *testing.T) {
	doc := analysis.QueryDocument{Text: "table custom | table cus*"}
	got, err := analysis.AnalyzeWithSourceUniverse(doc, analysis.SourceUniverse{
		Complete: false,
		Resolve:  func(string) analysis.SourceFieldAdmission { return analysis.SourceFieldAdmitted },
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Result.Status != analysis.Valid || len(got.Expansions) != 1 {
		t.Fatalf("result=%+v expansions=%+v", got.Result, got.Expansions)
	}
	want := []analysis.ExpandedField{{Name: "custom", Binding: "source"}}
	if !got.Expansions[0].Complete || !reflect.DeepEqual(got.Expansions[0].Matches, want) {
		t.Fatalf("expansion=%+v", got.Expansions[0])
	}
}

func TestPartialSourceAdmissionAndTransfers(t *testing.T) {
	admitted := func(string) analysis.SourceFieldAdmission { return analysis.SourceFieldAdmitted }
	prohibited := func(string) analysis.SourceFieldAdmission { return analysis.SourceFieldProhibited }
	unknown := func(string) analysis.SourceFieldAdmission { return analysis.SourceFieldIndeterminate }
	invalid := func(string) analysis.SourceFieldAdmission { return analysis.SourceFieldAdmission(99) }
	sourceHost := []analysis.ExpandedField{{Name: "host", Binding: "source"}}
	for _, tc := range []struct {
		name, query string
		universe    analysis.SourceUniverse
		status      analysis.Status
		binding     string
		complete    bool
		matches     []analysis.ExpandedField
	}{
		{"complete prevents callback expansion", "table custom | table cus*", analysis.SourceUniverse{Complete: true, Resolve: admitted}, analysis.Valid, "source", true, []analysis.ExpandedField{}},
		{"nil complete listed", "table host*", analysis.SourceUniverse{Fields: []string{"host"}, Complete: true}, analysis.Valid, "source", true, sourceHost},
		{"nil partial listed", "table host*", analysis.SourceUniverse{Fields: []string{"host"}}, analysis.Incomplete, "indeterminate", false, sourceHost},
		{"nil partial unlisted", "table custom | table cus*", analysis.SourceUniverse{}, analysis.Incomplete, "indeterminate", false, []analysis.ExpandedField{}},
		{"listed unknown", "table host*", analysis.SourceUniverse{Fields: []string{"host"}, Complete: true, Resolve: unknown}, analysis.Incomplete, "indeterminate", false, []analysis.ExpandedField{}},
		{"listed prohibited", "table host*", analysis.SourceUniverse{Fields: []string{"host"}, Complete: true, Resolve: prohibited}, analysis.Valid, "source", true, []analysis.ExpandedField{}},
		{"invalid enum", "table host*", analysis.SourceUniverse{Fields: []string{"host"}, Complete: true, Resolve: invalid}, analysis.Incomplete, "indeterminate", false, []analysis.ExpandedField{}},
		{"partial inclusion", "table host* | table hostname", analysis.SourceUniverse{Fields: []string{"host"}}, analysis.Incomplete, "indeterminate", false, sourceHost},
		{"partial removal", "fields -host* | table hostname", analysis.SourceUniverse{Fields: []string{"host"}}, analysis.Incomplete, "indeterminate", false, sourceHost},
		{"partial internals", "fields host | table _*", analysis.SourceUniverse{Fields: []string{"host"}}, analysis.Incomplete, "indeterminate", false, []analysis.ExpandedField{}},
		{"unknown exact projection", "| mystery | table host | table host*", analysis.SourceUniverse{Fields: []string{"host"}}, analysis.Incomplete, "indeterminate", false, []analysis.ExpandedField{}},
		{"aggregate precision", "| mystery | stats count AS total | table total*", analysis.SourceUniverse{}, analysis.Incomplete, "derived", true, []analysis.ExpandedField{{Name: "total", Binding: "derived"}}},
		{"derived partial evidence", "eval label=1 | table *", analysis.SourceUniverse{}, analysis.Incomplete, "indeterminate", false, []analysis.ExpandedField{{Name: "label", Binding: "derived"}}},
		{"prohibited obligation", "table prohibited | table pro*", analysis.SourceUniverse{Resolve: prohibited}, analysis.Valid, "source", true, []analysis.ExpandedField{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := analysis.AnalyzeWithSourceUniverse(analysis.QueryDocument{Text: tc.query}, tc.universe)
			if err != nil {
				t.Fatal(err)
			}
			refs := got.Result.References
			if tc.status == analysis.Incomplete && (got.Result.Coverage.SemanticComplete || len(got.Result.Coverage.Reasons) == 0) {
				t.Errorf("lost historical coverage: %+v", got.Result.Coverage)
			}
			for i, ref := range refs {
				if ref.ID != fmt.Sprintf("ref-%d", i) || tc.query[ref.Location.Start.Offset:ref.Location.End.Offset] != ref.OriginalName {
					t.Errorf("reference identity/span=%+v", ref)
				}
			}
			if got.Result.Status != tc.status || refs[len(refs)-1].Binding != tc.binding {
				t.Errorf("result=%+v refs=%+v", got.Result, refs)
			}
			if len(got.Expansions) != 1 {
				t.Fatalf("expansions=%+v", got.Expansions)
			}
			e := got.Expansions[0]
			if e.Complete != tc.complete || !reflect.DeepEqual(e.Matches, tc.matches) {
				t.Errorf("expansion=%+v want complete=%v matches=%+v", e, tc.complete, tc.matches)
			}
			if tc.name == "prohibited obligation" && refs[0].Binding != "source" {
				t.Errorf("initial source obligation=%+v", refs[0])
			}
			if tc.name == "partial removal" && (refs[0].Role != "remove" || refs[0].Binding != "not_applicable") {
				t.Errorf("removal=%+v", refs[0])
			}
		})
	}
}

func TestPartialSourceNamesAndDeterminism(t *testing.T) {
	for _, fields := range [][]string{{""}, {" \t\n"}, {string([]byte{0xff})}, {"host", "host"}} {
		if got, err := analysis.AnalyzeWithSourceUniverse(analysis.QueryDocument{Text: "table *"}, analysis.SourceUniverse{Fields: fields}); err == nil || got != nil {
			t.Fatalf("accepted invalid names %q: %+v %v", fields, got, err)
		}
	}
	fields := []string{"z", "café", "Host", "host", " a "}
	original := append([]string{}, fields...)
	membership := map[string]analysis.SourceFieldAdmission{"z": analysis.SourceFieldAdmitted, "café": analysis.SourceFieldAdmitted, "Host": analysis.SourceFieldProhibited, "host": analysis.SourceFieldAdmitted, " a ": analysis.SourceFieldAdmitted}
	universe := analysis.SourceUniverse{Fields: fields, Resolve: func(name string) analysis.SourceFieldAdmission { return membership[name] }}
	doc := analysis.QueryDocument{Text: "eval label=1\r\n| table 'café'* *", SourceID: "query-π"}
	got, err := analysis.AnalyzeWithSourceUniverse(doc, universe)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(fields, original) {
		t.Fatalf("input mutated: %q", fields)
	}
	want := []analysis.FieldExpansion{
		{ReferenceID: "ref-1", Complete: false, Matches: []analysis.ExpandedField{{Name: "café", Binding: "source"}}},
		{ReferenceID: "ref-2", Complete: false, Matches: []analysis.ExpandedField{{Name: "café", Binding: "source"}, {Name: "label", Binding: "derived"}}},
	}
	// The second selector observes the first selector's unresolved transfer:
	// only its already tracked source and derived bindings remain proven.
	if !reflect.DeepEqual(got.Expansions, want) {
		t.Fatalf("expansions=%+v", got.Expansions)
	}
	ref := got.Result.References[1]
	if ref.OriginalName != "'café'*" || ref.Location.Start.Offset != 22 || ref.Location.End.Offset != 30 || ref.Location.Start.Line != 2 || ref.Location.Start.Column != 9 {
		t.Fatalf("span=%+v", ref)
	}
	baseline, _ := json.Marshal(got)
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			other, err := analysis.AnalyzeWithSourceUniverse(doc, universe)
			if err != nil {
				t.Error(err)
				return
			}
			encoded, _ := json.Marshal(other)
			if string(encoded) != string(baseline) {
				t.Error("nondeterministic result")
			}
			other.Expansions[0].Matches[0].Name = "changed"
		}()
	}
	wg.Wait()
	encoded, _ := json.Marshal(got)
	if string(encoded) != string(baseline) {
		t.Error("aliased result")
	}
	for _, query := range []string{"table *", "table host | table host*", "search [ table user* ] host=x"} {
		legacy, err := analysis.AnalyzeWithSourceFields(analysis.QueryDocument{Text: query}, []string{"host", "user"})
		if err != nil {
			t.Fatal(err)
		}
		modern, err := analysis.AnalyzeWithSourceUniverse(analysis.QueryDocument{Text: query}, analysis.SourceUniverse{Fields: []string{"host", "user"}, Complete: true})
		if err != nil {
			t.Fatal(err)
		}
		// Conclusive membership and child ID fixtures agree byte for byte.
		a, _ := json.Marshal(legacy)
		b, _ := json.Marshal(modern)
		if string(a) != string(b) {
			t.Errorf("finite compatibility %q:\n%s\n%s", query, a, b)
		}
	}
}

func TestPartialSourceUnresolvedRetainedInternal(t *testing.T) {
	got, err := analysis.AnalyzeWithSourceUniverse(analysis.QueryDocument{Text: "fields host"}, analysis.SourceUniverse{Fields: []string{"host", "_custom"}, Complete: true, Resolve: func(name string) analysis.SourceFieldAdmission {
		if name == "host" {
			return analysis.SourceFieldAdmitted
		}
		return analysis.SourceFieldIndeterminate
	}})
	if err != nil {
		t.Fatal(err)
	}
	if got.Result.Status != analysis.Incomplete || got.Result.Coverage.SemanticComplete {
		t.Fatalf("unresolved retained internal reported complete: %+v", got.Result)
	}
}

func TestPartialSourceAdmissionBoundaries(t *testing.T) {
	for _, query := range []string{"table custom | table cus*", "eval host=1 | table host*"} {
		got, err := analysis.AnalyzeWithSourceUniverse(analysis.QueryDocument{Text: query}, analysis.SourceUniverse{Fields: []string{"host"}, Complete: true, Resolve: func(string) analysis.SourceFieldAdmission {
			panic("resolver called for an unlisted obligation or proven derived binding")
		}})
		if err != nil {
			t.Fatal(err)
		}
		if got.Result.Status != analysis.Valid || !got.Expansions[0].Complete {
			t.Fatalf("result=%+v expansions=%+v", got.Result, got.Expansions)
		}
	}
	for _, complete := range []bool{false, true} {
		got, err := analysis.AnalyzeWithSourceUniverse(analysis.QueryDocument{Text: "| mystery | table host* | table host | table host*"}, analysis.SourceUniverse{Fields: []string{"host"}, Complete: complete})
		if err != nil {
			t.Fatal(err)
		}
		want := []analysis.FieldExpansion{{ReferenceID: "ref-0", Complete: false, Matches: []analysis.ExpandedField{}}, {ReferenceID: "ref-2", Complete: false, Matches: []analysis.ExpandedField{}}}
		if got.Result.Status != analysis.Incomplete || got.Result.References[1].Binding != "indeterminate" || !reflect.DeepEqual(got.Expansions, want) {
			t.Fatalf("conditional provenance promoted: %+v %+v", got.Result, got.Expansions)
		}
	}
}
