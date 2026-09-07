package validation_test

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/validation"
)

type semanticCase struct {
	id, query        string
	fields, optional []string
	status           analysis.Status
	complete         bool
	outcomes         string // Ordered consuming reference name:outcome; writes/removals must be absent.
}

// These independently stated semantic obligations precede the full-report
// golden fixtures. A changed binding, missing obligation, or invented match
// fails here even if a fixture is regenerated.
var semanticCases = []semanticCase{
	{"exact_nested_optional", "search Host=x host=y 'user.name'=z | table optional", []string{"Host", "host", "user.name"}, []string{"optional"}, analysis.Valid, true, "Host:matching,host:matching,user.name:matching,optional:optional_equivalent"},
	{"nested_no_parent", "table user user.name", []string{"user.name"}, nil, analysis.Invalid, true, "user:missing,user.name:matching"},
	{"nested_no_child", "table user.name", []string{"user"}, nil, analysis.Invalid, true, "user.name:missing"},
	{"empty_derived", "eval answer=1 | table answer", nil, nil, analysis.Valid, true, "answer:matching"},
	{"missing", "search missing=x", []string{"host"}, nil, analysis.Invalid, true, "missing:missing"},
	{"empty_inclusion", "table host*", nil, nil, analysis.Invalid, true, "host*:missing"},
	{"empty_removal", "fields -absent*", []string{"host"}, nil, analysis.Valid, true, ""},
	{"removed_wildcard", "fields -host | table host*", []string{"host"}, nil, analysis.Invalid, true, "host*:unavailable"},
	{"removed_absent", "search missing=x | fields -miss* | where missing=2", []string{"host"}, nil, analysis.Invalid, true, "missing:missing,missing:unavailable"},
	{"absent_not_structural", "table missing | table *", []string{"host"}, nil, analysis.Invalid, true, "missing:missing,*:missing"},
	{"all_optional", "table opt*", nil, []string{"optA", "optB"}, analysis.Valid, true, "opt*:optional_equivalent"},
	{"ordinary_optional", "table host*", []string{"host"}, []string{"hostname"}, analysis.Valid, true, "host*:matching"},
	{"mixed_bindings", "eval label=host | table *", nil, []string{"host"}, analysis.Valid, true, "host:optional_equivalent,*:matching"},
	{"closed", "fields host | table user", []string{"host", "user"}, nil, analysis.Invalid, true, "host:matching,user:unavailable"},
	{"internal", "fields host | table _time", []string{"host", "_time"}, nil, analysis.Valid, true, "host:matching,_time:matching"},
	{"internal_removed", "fields -_time | fields host | table _time", []string{"host", "_time"}, nil, analysis.Invalid, true, "host:matching,_time:unavailable"},
	{"unknown_command", "| mystery | table host*", []string{"host"}, nil, analysis.Incomplete, false, "host*:indeterminate"},
	{"unknown_function", "eval out=mystery(host) | table out", []string{"host"}, nil, analysis.Incomplete, false, "host:indeterminate,out:indeterminate"},
	{"macro", "search host=x | `expand` | table host", []string{"host"}, nil, analysis.Incomplete, false, "host:matching,host:matching"},
	{"lookup", "lookup assets host OUTPUTNEW label | table lab*", []string{"host", "label"}, nil, analysis.Incomplete, false, "host:matching,lab*:indeterminate"},
	{"independent_scope", "eval label=host | append [ table * ] | table lab*", []string{"host", "user"}, nil, analysis.Incomplete, false, "host:matching,*:matching,lab*:indeterminate"},
	{"inherited_scope", "table host | appendpipe [ table * ]", []string{"host", "user"}, nil, analysis.Incomplete, false, "host:matching,*:matching"},
	{"syntax_empty", " ", nil, nil, analysis.Invalid, false, ""},
	{"unicode_crlf", "eval label='café'\r\n| table 'café'* lab*", []string{"café"}, nil, analysis.Valid, true, "café:matching,café*:matching,lab*:matching"},
	{"invalid_incomplete", "search missing=x | mystery | table host", []string{"host"}, nil, analysis.Invalid, false, "missing:missing,host:indeterminate"},
	{"rename", "rename host AS renamed | table renamed", []string{"host"}, nil, analysis.Valid, true, "host:matching,renamed:matching"},
}

func checkSemantics(t *testing.T, tc semanticCase, report *validation.Report) {
	t.Helper()
	if report.Status != tc.status || report.Coverage.SchemaComplete != tc.complete {
		t.Errorf("%s status/coverage = %s %+v; want %s %t", tc.id, report.Status, report.Coverage, tc.status, tc.complete)
	}
	refs := map[string]analysis.Reference{}
	for _, ref := range report.Analysis.References {
		refs[ref.ID] = ref
	}
	parts := []string{}
	for _, outcome := range report.Outcomes {
		ref, ok := refs[outcome.ReferenceID]
		if !ok {
			t.Fatalf("unknown reference %q", outcome.ReferenceID)
		}
		parts = append(parts, ref.NormalizedName+":"+outcome.Outcome)
		if ref.Role == "remove" || ref.Binding == "not_applicable" || ref.Kind != "field" {
			t.Errorf("non-consuming outcome: %+v", ref)
		}
		loc := ref.Location
		// Check byte spans against original source, independently of fixture offsets.
		token := tc.query[loc.Start.Offset:loc.End.Offset]
		if token != ref.OriginalName {
			t.Errorf("%s span token %q != original %q", tc.id, token, ref.OriginalName)
		}
		if outcome.Matches == nil {
			t.Error("null matches")
		}
		if outcome.Outcome == "matching" || outcome.Outcome == "optional_equivalent" {
			if len(outcome.Matches) == 0 {
				t.Errorf("match without evidence %+v", outcome)
			}
		} else if len(outcome.Matches) != 0 {
			t.Errorf("invented evidence %+v", outcome)
		}
	}
	if strings.Join(parts, ",") != tc.outcomes {
		t.Errorf("%s outcomes %q; want %q", tc.id, strings.Join(parts, ","), tc.outcomes)
	}
	if tc.id == "unicode_crlf" {
		a := report.Analysis.References
		for _, ref := range a {
			if ref.NormalizedName == "café*" && (ref.Location.Start.Offset != 28 || ref.Location.Start.Line != 2 || ref.Location.Start.Column != 9 || ref.Location.End.Offset != 36 || ref.Location.End.Column != 16) {
				t.Errorf("Unicode span %+v", ref.Location)
			}
		}
	}
	for _, d := range report.Diagnostics {
		if d.Code == "SPL_UNKNOWN_FIELD" || d.Code == "SPL_INDETERMINATE_FIELD" {
			found := false
			for _, ref := range report.Analysis.References {
				if d.Location == ref.Location && d.StageID == ref.StageID && d.ScopeID == ref.ScopeID {
					found = true
				}
			}
			if !found {
				t.Errorf("unlocated validation finding %+v", d)
			}
		}
	}
}

func TestValidationSemanticCases(t *testing.T) {
	for _, tc := range semanticCases {
		t.Run(tc.id, func(t *testing.T) {
			r, err := validation.Validate(analysis.QueryDocument{Text: tc.query, SourceID: tc.id}, validation.FieldCatalog{Fields: tc.fields, OptionalFields: tc.optional, Identity: "local-fixture", Version: "1"})
			if err != nil {
				t.Fatal(err)
			}
			checkSemantics(t, tc, r)
		})
	}
}

type fixtureCase struct {
	ID       string                  `json:"id"`
	Document analysis.QueryDocument  `json:"document"`
	Catalog  validation.FieldCatalog `json:"catalog"`
	Expected *validation.Report      `json:"expected"`
}
type fixtureCorpus struct {
	SchemaVersion int           `json:"schema_version"`
	Cases         []fixtureCase `json:"cases"`
}

func TestValidationCorpus(t *testing.T) {
	data, err := os.ReadFile("../../testdata/validation/cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var corpus fixtureCorpus
	if err := json.Unmarshal(data, &corpus); err != nil {
		t.Fatal(err)
	}
	if corpus.SchemaVersion != 1 || len(corpus.Cases) != len(semanticCases) {
		t.Fatal("incomplete corpus")
	}
	for i, tc := range corpus.Cases {
		t.Run(tc.ID, func(t *testing.T) {
			if tc.ID != semanticCases[i].id || tc.Document.Text != semanticCases[i].query {
				t.Fatal("fixture source drift")
			}
			checkSemantics(t, semanticCases[i], tc.Expected)
			actual, err := validation.Validate(tc.Document, tc.Catalog)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(actual, tc.Expected) {
				t.Fatalf("full report differs from reviewed %s fixture", tc.ID)
			}
		})
	}
}

func TestValidationDeterminismAndIsolation(t *testing.T) {
	document := analysis.QueryDocument{Text: "eval label=host | table *"}
	catalog := validation.FieldCatalog{Fields: []string{"user", "host"}, OptionalFields: []string{"z", "a"}}
	before := validation.FieldCatalog{Fields: append([]string{}, catalog.Fields...), OptionalFields: append([]string{}, catalog.OptionalFields...)}
	baseline, err := validation.Validate(document, catalog)
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 3; j++ {
				got, err := validation.Validate(document, catalog)
				if err != nil || !reflect.DeepEqual(got, baseline) {
					t.Errorf("unstable report: %v", err)
				}
			}
		}()
	}
	wg.Wait()
	if !reflect.DeepEqual(before, catalog) {
		t.Fatal("caller catalog mutated")
	}
	catalog.Fields[0] = "mutated"
	catalog.OptionalFields[0] = "mutated"
	if !reflect.DeepEqual(baseline.Target.Fields, []string{"host", "user"}) || !reflect.DeepEqual(baseline.Target.OptionalFields, []string{"a", "z"}) {
		t.Fatal("target aliases input")
	}
}
