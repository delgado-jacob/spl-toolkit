package analysis_test

import (
	"encoding/json"
	"reflect"
	"sync"
	"testing"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
)

func TestSourceUniverseRefinesFieldsWithoutChangingAnalyze(t *testing.T) {
	document := analysis.QueryDocument{Text: "fields host* | table hostname"}
	plain, err := analysis.Analyze(document)
	if err != nil {
		t.Fatal(err)
	}
	if plain.Status != analysis.Incomplete {
		t.Fatalf("plain = %s", plain.Status)
	}
	refined, err := analysis.AnalyzeWithSourceFields(document, []string{"hostname"})
	if err != nil {
		t.Fatal(err)
	}
	if refined.Result.Status != analysis.Valid {
		t.Fatalf("refined = %+v", refined.Result)
	}
	if len(refined.Expansions) != 1 || !refined.Expansions[0].Complete {
		t.Fatalf("evidence = %+v", refined.Expansions)
	}
	matches := refined.Expansions[0].Matches
	if len(matches) != 1 || matches[0].Name != "hostname" || matches[0].Binding != "source" {
		t.Fatalf("matches = %+v", matches)
	}
}

// Catches invented source membership, reopened projections, and lost uncertainty.
func TestSourceUniverseFlow(t *testing.T) {
	for _, tc := range []struct {
		name, query string
		fields      []string
		status      analysis.Status
		lastBinding string
		complete    bool
		matches     []analysis.ExpandedField
		wildcards   int
	}{
		{"exact absent remains source", "search missing=x", []string{"host"}, analysis.Valid, "source", false, nil, 0},
		{"empty source selection", "table host*", nil, analysis.Valid, "source", true, []analysis.ExpandedField{}, 1},
		{"removed source selection", "fields -host | table host*", []string{"host"}, analysis.Invalid, "unavailable", true, []analysis.ExpandedField{}, 1},
		{"empty removal", "fields -absent*", []string{"host"}, analysis.Valid, "not_applicable", true, []analysis.ExpandedField{}, 1},
		{"closed exact projection", "fields host | table user", []string{"host", "user"}, analysis.Invalid, "unavailable", false, nil, 0},
		{"derived", "eval label=host | table lab*", []string{"host"}, analysis.Valid, "derived", true, []analysis.ExpandedField{{Name: "label", Binding: "derived"}}, 1},
		{"mixed", "eval label=host | table *", []string{"host"}, analysis.Valid, "indeterminate", true, []analysis.ExpandedField{{Name: "host", Binding: "source"}, {Name: "label", Binding: "derived"}}, 1},
		{"unknown", "| mystery | table *", []string{"host"}, analysis.Incomplete, "indeterminate", false, []analysis.ExpandedField{}, 1},
		{"restored aggregate", "| mystery | stats count AS total | table total*", []string{"host"}, analysis.Incomplete, "derived", true, []analysis.ExpandedField{{Name: "total", Binding: "derived"}}, 1},
		{"local absence", "| mystery | table host | table absent*", []string{"host"}, analysis.Invalid, "unavailable", true, []analysis.ExpandedField{}, 1},
		{"conditional projection", "| mystery | table host | table host*", []string{"host"}, analysis.Incomplete, "indeterminate", false, []analysis.ExpandedField{}, 1},
		{"conditional shadows catalog", "lookup assets host OUTPUTNEW label | table lab*", []string{"host", "label"}, analysis.Incomplete, "indeterminate", false, []analysis.ExpandedField{}, 1},
		{"retained internal", "fields host | table _time", []string{"host", "_time"}, analysis.Valid, "source", false, nil, 0},
		{"removed internal", "fields -_time | fields host | table _time", []string{"host", "_time"}, analysis.Invalid, "unavailable", false, nil, 0},
		{"no internal defaults", "fields host | table _time", nil, analysis.Invalid, "unavailable", false, nil, 0},
		{"tracked absent removed", "search missing=x | fields -miss* | where missing=2", []string{"host"}, analysis.Invalid, "unavailable", true, []analysis.ExpandedField{}, 1},
		{"absent obligation not a member", "table missing | table *", []string{"host"}, analysis.Valid, "source", true, []analysis.ExpandedField{}, 1},
		{"reset source", "fields -host | inputlookup assets | table host*", []string{"host"}, analysis.Valid, "source", true, []analysis.ExpandedField{{Name: "host", Binding: "source"}}, 1},
		{"closed wildcard projection", "table host | table user*", []string{"host", "user"}, analysis.Invalid, "unavailable", true, []analysis.ExpandedField{}, 1},
		{"derived shadow", "eval host=1 | table host*", []string{"host"}, analysis.Valid, "derived", true, []analysis.ExpandedField{{Name: "host", Binding: "derived"}}, 1},
		{"unsupported sort", "sort host*", []string{"host"}, analysis.Incomplete, "source", true, []analysis.ExpandedField{{Name: "host", Binding: "source"}}, 1},
		{"unsupported stats", "stats count BY host*", []string{"host"}, analysis.Incomplete, "source", true, []analysis.ExpandedField{{Name: "host", Binding: "source"}}, 1},
		{"unsupported rename", "rename 'host*' AS 'name*'", []string{"host"}, analysis.Incomplete, "source", true, []analysis.ExpandedField{{Name: "host", Binding: "source"}}, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a, err := analysis.AnalyzeWithSourceFields(analysis.QueryDocument{Text: tc.query}, tc.fields)
			if err != nil {
				t.Fatal(err)
			}
			if a.Result.Status != tc.status {
				t.Errorf("status = %s; want %s; diagnostics = %+v", a.Result.Status, tc.status, a.Result.Diagnostics)
			}
			if n := len(a.Result.References); n == 0 || a.Result.References[n-1].Binding != tc.lastBinding {
				t.Errorf("references = %+v; want last %s", a.Result.References, tc.lastBinding)
			}
			if len(a.Expansions) != tc.wildcards {
				t.Fatalf("expansions = %+v", a.Expansions)
			}
			if tc.wildcards > 0 {
				e := a.Expansions[0]
				if e.Complete != tc.complete || !reflect.DeepEqual(e.Matches, tc.matches) {
					t.Errorf("expansion = %+v; want complete %t, matches %+v", e, tc.complete, tc.matches)
				}
			}
			if tc.name == "tracked absent removed" {
				if a.Result.References[0].Binding != "source" || a.Result.References[1].Role != "remove" || a.Result.References[1].Binding != "not_applicable" {
					t.Error("lost external obligation/removal role", a.Result.References)
				}
				if len(a.Result.Diagnostics) != 1 || a.Result.Diagnostics[0].Code != analysis.CodeUnavailableField || a.Result.Diagnostics[0].Location != a.Result.References[2].Location {
					t.Error("removal created an obligation", a.Result.Diagnostics)
				}
			}
		})
	}
}

func TestSourceUniverseInputNames(t *testing.T) {
	for _, fields := range [][]string{{""}, {" \t\n"}, {string([]byte{0xff})}, {"host", "host"}} {
		if a, err := analysis.AnalyzeWithSourceFields(analysis.QueryDocument{Text: "table *"}, fields); err == nil || a != nil {
			t.Errorf("accepted invalid names %#v: %+v %v", fields, a, err)
		}
	}
	fields := []string{"z", " a ", "'quoted'", "a*b", "Host", "host", "café"}
	original := append([]string{}, fields...)
	a, err := analysis.AnalyzeWithSourceFields(analysis.QueryDocument{Text: "table *"}, fields)
	if err != nil {
		t.Fatal(err)
	}
	want := []analysis.ExpandedField{{Name: " a ", Binding: "source"}, {Name: "'quoted'", Binding: "source"}, {Name: "Host", Binding: "source"}, {Name: "a*b", Binding: "source"}, {Name: "café", Binding: "source"}, {Name: "host", Binding: "source"}, {Name: "z", Binding: "source"}}
	if !reflect.DeepEqual(fields, original) || !reflect.DeepEqual(a.Expansions[0].Matches, want) {
		t.Errorf("normalization changed concrete names: input %q, output %+v", fields, a.Expansions)
	}
	for _, fields := range [][]string{nil, {}} {
		a, err := analysis.AnalyzeWithSourceFields(analysis.QueryDocument{Text: "table *"}, fields)
		if err != nil || a.Expansions == nil || len(a.Expansions) != 1 || !a.Expansions[0].Complete || a.Expansions[0].Matches == nil || len(a.Expansions[0].Matches) != 0 {
			t.Fatalf("empty universe = %+v, %v", a, err)
		}
	}
	for _, document := range []analysis.QueryDocument{{Language: "bad"}, {Profile: "bad"}, {Version: "bad"}, {Text: string([]byte{0xff})}, {SourceID: string([]byte{0xff})}} {
		_, plainErr := analysis.Analyze(document)
		_, refinedErr := analysis.AnalyzeWithSourceFields(document, nil)
		if plainErr == nil || refinedErr == nil || plainErr.Error() != refinedErr.Error() {
			t.Errorf("normalization mismatch: %v / %v", plainErr, refinedErr)
		}
	}
}

func TestSourceUniverseScopesAndReferenceOrder(t *testing.T) {
	for _, tc := range []struct {
		query string
		want  []analysis.FieldExpansion
	}{
		{"table host | appendpipe [ table * ]", []analysis.FieldExpansion{{ReferenceID: "ref-1", Complete: true, Matches: []analysis.ExpandedField{{Name: "host", Binding: "source"}}}}},
		{"eval label=host | append [ table * ] | table lab*", []analysis.FieldExpansion{{ReferenceID: "ref-2", Complete: true, Matches: []analysis.ExpandedField{{Name: "host", Binding: "source"}, {Name: "user", Binding: "source"}}}, {ReferenceID: "ref-3", Complete: false, Matches: []analysis.ExpandedField{}}}},
		{"search host=x [ table user* ] | table host*", []analysis.FieldExpansion{{ReferenceID: "ref-1", Complete: true, Matches: []analysis.ExpandedField{{Name: "user", Binding: "source"}}}, {ReferenceID: "ref-2", Complete: false, Matches: []analysis.ExpandedField{}}}},
	} {
		a, err := analysis.AnalyzeWithSourceFields(analysis.QueryDocument{Text: tc.query}, []string{"user", "host"})
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(a.Expansions, tc.want) {
			t.Errorf("%s: expansions = %+v; want %+v; result %+v", tc.query, a.Expansions, tc.want, a.Result)
		}
	}
}

func TestSourceUniverseDeterminismAndLocations(t *testing.T) {
	document := analysis.QueryDocument{Text: "eval label='café'\r\n| table 'café'* lab*", SourceID: "query-π"}
	fields := []string{"café", "caféteria"}
	a, err := analysis.AnalyzeWithSourceFields(document, fields)
	if err != nil {
		t.Fatal(err)
	}
	if a.Result.Status != analysis.Valid {
		t.Fatalf("result = %+v", a.Result)
	}
	if len(a.Expansions) != 2 || a.Expansions[0].ReferenceID != "ref-2" || a.Expansions[1].ReferenceID != "ref-3" {
		t.Fatalf("reference order %+v", a.Expansions)
	}
	ref := a.Result.References[2]
	if ref.OriginalName != "'café'*" || ref.Location.Start.Offset != 28 || ref.Location.End.Offset != 36 || ref.Location.Start.Line != 2 || ref.Location.Start.Column != 9 || ref.Location.End.Column != 16 {
		t.Errorf("UTF-8/CRLF span = %+v", ref)
	}
	baseline, _ := json.Marshal(a)
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			b, err := analysis.AnalyzeWithSourceFields(document, fields)
			if err != nil {
				t.Error(err)
				return
			}
			encoded, _ := json.Marshal(b)
			if string(encoded) != string(baseline) {
				t.Error("nondeterministic result")
			}
			b.Expansions[0].Matches[0].Name = "changed"
		}()
	}
	wg.Wait()
	encoded, _ := json.Marshal(a)
	if string(encoded) != string(baseline) {
		t.Error("aliased result slices")
	}
	plain, _ := analysis.Analyze(analysis.QueryDocument{Text: "search host=x"})
	refined, _ := analysis.AnalyzeWithSourceFields(analysis.QueryDocument{Text: "search host=x"}, fields)
	if !reflect.DeepEqual(plain, refined.Result) || refined.Expansions == nil {
		t.Error("unrequested source materialization or null expansions", refined)
	}
}

// A finite declaration cannot establish availability after unknown effects.
func TestSourceUniverseUnknownMaterializationStaysConditional(t *testing.T) {
	a, err := analysis.AnalyzeWithSourceFields(analysis.QueryDocument{Text: "| mystery | table host* | table host | table host*"}, []string{"host"})
	if err != nil {
		t.Fatal(err)
	}
	if a.Result.Status != analysis.Incomplete || len(a.Expansions) != 2 || a.Expansions[0].Complete || a.Expansions[1].Complete {
		t.Errorf("unknown membership became certain: %+v, %+v", a.Expansions, a.Result)
	}
	if len(a.Result.References) != 3 || a.Result.References[1].Binding != "indeterminate" {
		t.Errorf("exact projection promoted unknown source: %+v", a.Result.References)
	}
}

// The parent is analyzed before children, but final IDs follow source order.
func TestSourceUniverseChildExpansionRemapsBeforeParentReference(t *testing.T) {
	a, err := analysis.AnalyzeWithSourceFields(analysis.QueryDocument{Text: "search [ table user* ] host=x"}, []string{"user", "host"})
	if err != nil {
		t.Fatal(err)
	}
	if a.Result.Status != analysis.Incomplete || len(a.Expansions) != 1 || a.Expansions[0].ReferenceID != "ref-0" || a.Result.References[0].ScopeID != "scope-1" || a.Result.References[1].NormalizedName != "host" {
		t.Fatalf("reference remapping = %+v, result %+v", a.Expansions, a.Result)
	}
	if !reflect.DeepEqual(a.Expansions[0].Matches, []analysis.ExpandedField{{Name: "user", Binding: "source"}}) {
		t.Error("child source evidence", a.Expansions)
	}
}
