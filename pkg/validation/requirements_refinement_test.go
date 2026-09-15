package validation

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
)

func TestRequirementSetValidationRefinementParity(t *testing.T) {
	ocsf := readOCSFFixture(t, "base")
	cases := []struct {
		name     string
		document analysis.QueryDocument
		validate func(analysis.QueryDocument) (*analysis.Result, error)
	}{
		{
			name:     "field list wildcard",
			document: analysis.QueryDocument{Text: "fields host* | table hostname"},
			validate: func(document analysis.QueryDocument) (*analysis.Result, error) {
				report, err := Validate(document, FieldCatalog{Fields: []string{"hostname"}, Identity: "fields", Version: "1"})
				if err != nil {
					return nil, err
				}
				return report.Analysis, nil
			},
		},
		{
			name:     "JSON Schema dotted SPL2",
			document: analysis.QueryDocument{Text: "FROM main SELECT actor.user.name", Language: "spl2"},
			validate: func(document analysis.QueryDocument) (*analysis.Result, error) {
				report, err := ValidateSchema(document, SchemaTarget{Kind: "json_schema", Schema: json.RawMessage(`{"type":"object","properties":{"actor.user.name":true},"additionalProperties":false}`)})
				if err != nil {
					return nil, err
				}
				return report.Analysis, nil
			},
		},
		{
			name:     "OCSF exact",
			document: analysis.QueryDocument{Text: "table time"},
			validate: func(document analysis.QueryDocument) (*analysis.Result, error) {
				report, err := ValidateSchema(document, SchemaTarget{Kind: "ocsf", Catalog: ocsf, Selection: &OCSFSelection{Version: "1.6.0", Class: "authentication", Profiles: []string{}, Extensions: []string{}}})
				if err != nil {
					return nil, err
				}
				return report.Analysis, nil
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			want, err := analysis.Requirements(tc.document)
			if err != nil {
				t.Fatal(err)
			}
			result, err := tc.validate(tc.document)
			if err != nil {
				t.Fatal(err)
			}
			wantJSON, err := json.Marshal(want)
			if err != nil {
				t.Fatal(err)
			}
			gotJSON, err := json.Marshal(result.Requirements)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(gotJSON, wantJSON) {
				t.Fatalf("validation refinement changed requirements:\nwant=%s\ngot=%s", wantJSON, gotJSON)
			}
		})
	}
}

func TestRequirementSetValidationResourceLimitParity(t *testing.T) {
	document := analysis.QueryDocument{Text: strings.Repeat("a ", 4097), SourceID: "resource-limited"}
	want, err := analysis.Analyze(document)
	if err != nil {
		t.Fatal(err)
	}
	if len(want.Diagnostics) != 1 || want.Diagnostics[0].Code != analysis.CodeAnalysisResourceLimit || len(want.Stages) != 0 || len(want.References) != 0 {
		t.Fatalf("fixture did not produce the bounded resource outcome: %+v", want)
	}
	fieldReport, err := Validate(document, FieldCatalog{Fields: []string{"a"}})
	if err != nil {
		t.Fatal(err)
	}
	jsonReport, err := ValidateSchema(document, SchemaTarget{Kind: "json_schema", Schema: json.RawMessage(`{"type":"object","properties":{"a":true},"additionalProperties":false}`)})
	if err != nil {
		t.Fatal(err)
	}
	ocsfReport, err := ValidateSchema(document, SchemaTarget{Kind: "ocsf", Catalog: readOCSFFixture(t, "base"), Selection: &OCSFSelection{Version: "1.6.0", Class: "authentication", Profiles: []string{}, Extensions: []string{}}})
	if err != nil {
		t.Fatal(err)
	}
	for name, got := range map[string]*analysis.Result{
		"field list":  fieldReport.Analysis,
		"JSON Schema": jsonReport.Analysis,
		"OCSF":        ocsfReport.Analysis,
	} {
		if !reflect.DeepEqual(got, want) {
			t.Errorf("%s validation changed bounded incomplete analysis", name)
		}
	}
}
