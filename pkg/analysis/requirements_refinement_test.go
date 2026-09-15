package analysis_test

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
)

func TestRequirementSetFieldRefinementParity(t *testing.T) {
	cases := []struct {
		name     string
		document analysis.QueryDocument
		universe analysis.SourceUniverse
	}{
		{name: "exact finite", document: analysis.QueryDocument{Text: "search host=web | table host"}, universe: analysis.SourceUniverse{Fields: []string{"host"}, Complete: true}},
		{name: "partial", document: analysis.QueryDocument{Text: "search host=web user=admin | table host user"}, universe: analysis.SourceUniverse{Fields: []string{"host"}, Complete: false}},
		{name: "wildcard finite", document: analysis.QueryDocument{Text: "fields host* | table hostname"}, universe: analysis.SourceUniverse{Fields: []string{"hostname"}, Complete: true}},
		{name: "unresolved wildcard", document: analysis.QueryDocument{Text: "| mystery | table host*"}, universe: analysis.SourceUniverse{Fields: []string{"hostname"}, Complete: false}},
		{name: "dotted SPL2", document: analysis.QueryDocument{Text: "FROM main SELECT actor.user.name", Language: "spl2"}, universe: analysis.SourceUniverse{Fields: []string{"actor.user.name"}, Complete: false}},
		{name: "SPL2 recovered diagnostic prefix", document: analysis.QueryDocument{Text: "FROM main | foobar | stats c=count() BY host | eval y=other", Language: "spl2"}, universe: analysis.SourceUniverse{Fields: []string{"host", "other"}, Complete: true}},
		{name: "aggregation closes prior uncertainty", document: analysis.QueryDocument{Text: "search index=main | foobar | stats count by host | eval y=other"}, universe: analysis.SourceUniverse{Fields: []string{"host", "other"}, Complete: true}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			plain, err := analysis.Analyze(tc.document)
			if err != nil {
				t.Fatal(err)
			}
			withFields, err := analysis.AnalyzeWithSourceFields(tc.document, tc.universe.Fields)
			if err != nil {
				t.Fatal(err)
			}
			withUniverse, err := analysis.AnalyzeWithSourceUniverse(tc.document, tc.universe)
			if err != nil {
				t.Fatal(err)
			}
			want, err := json.Marshal(plain.Requirements)
			if err != nil {
				t.Fatal(err)
			}
			for name, set := range map[string]analysis.RequirementSet{
				"fields":   withFields.Result.Requirements,
				"universe": withUniverse.Result.Requirements,
			} {
				got, err := json.Marshal(set)
				if err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(got, want) {
					t.Fatalf("embedded requirements changed under %s refinement:\nplain=%s\nrefined=%s", name, want, got)
				}
			}
		})
	}
}

func TestRequirementSetFieldRefinementResourceLimitParity(t *testing.T) {
	document := analysis.QueryDocument{Text: strings.Repeat("a ", 4097), SourceID: "resource-limited"}
	plain, err := analysis.Analyze(document)
	if err != nil {
		t.Fatal(err)
	}
	if len(plain.Diagnostics) != 1 || plain.Diagnostics[0].Code != analysis.CodeAnalysisResourceLimit || len(plain.Stages) != 0 || len(plain.References) != 0 {
		t.Fatalf("fixture did not produce the bounded resource outcome: %+v", plain)
	}
	finite, err := analysis.AnalyzeWithSourceFields(document, []string{"a"})
	if err != nil {
		t.Fatal(err)
	}
	partial, err := analysis.AnalyzeWithSourceUniverse(document, analysis.SourceUniverse{Fields: []string{"a"}, Complete: false})
	if err != nil {
		t.Fatal(err)
	}
	for name, got := range map[string]*analysis.Result{"finite": finite.Result, "partial": partial.Result} {
		if !reflect.DeepEqual(got, plain) {
			t.Errorf("%s refinement changed bounded incomplete analysis", name)
		}
	}
}
