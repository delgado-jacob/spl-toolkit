package validation_test

import (
	"bytes"
	"encoding/json"
	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/validation"
	"reflect"
	"strings"
	"testing"
)

func TestValidateDistinguishesExternalAndStructuralAbsence(t *testing.T) {
	catalog := validation.FieldCatalog{Fields: []string{"host", "user"}}
	external, err := validation.Validate(analysis.QueryDocument{Text: "search missing=x"}, catalog)
	if err != nil {
		t.Fatal(err)
	}
	if external.Status != analysis.Invalid || external.Outcomes[0].Outcome != "missing" {
		t.Fatalf("external = %+v", external)
	}
	structural, err := validation.Validate(analysis.QueryDocument{Text: "fields host | table user"}, catalog)
	if err != nil {
		t.Fatal(err)
	}
	last := structural.Outcomes[len(structural.Outcomes)-1]
	if structural.Status != analysis.Invalid || last.Outcome != "unavailable" {
		t.Fatalf("structural = %+v", structural)
	}
}

func TestValidateDistinguishesRemovalOfAbsentSource(t *testing.T) {
	report, err := validation.Validate(analysis.QueryDocument{Text: "search missing=x | fields -miss* | where missing=2"}, validation.FieldCatalog{Fields: []string{"host"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Outcomes) != 2 || report.Outcomes[0].Outcome != "missing" || report.Outcomes[1].Outcome != "unavailable" {
		t.Fatalf("outcomes = %+v", report.Outcomes)
	}
	for _, outcome := range report.Outcomes {
		if len(outcome.Matches) != 0 {
			t.Fatalf("fabricated evidence: %+v", outcome)
		}
	}
	if len(report.Diagnostics) != 2 || report.Diagnostics[0].Code != "SPL_UNKNOWN_FIELD" || report.Diagnostics[0].Location.Start.Offset != 7 || report.Diagnostics[1].Code != "SPL_UNAVAILABLE_FIELD" || report.Diagnostics[1].Location.Start.Offset != 41 {
		t.Fatalf("diagnostics = %+v", report.Diagnostics)
	}
	if report.Analysis.Status != analysis.Invalid || len(report.Analysis.Diagnostics) != 1 {
		t.Fatalf("analysis mutated: %+v", report.Analysis)
	}
}

func TestValidationEvidenceAndCoverage(t *testing.T) {
	for _, tc := range []struct {
		query     string
		catalog   validation.FieldCatalog
		want      []validation.Match
		aggregate string
	}{
		{"table opt*", validation.FieldCatalog{OptionalFields: []string{"optB", "optA"}}, []validation.Match{{Name: "optA", Binding: "source", Outcome: "optional_equivalent"}, {Name: "optB", Binding: "source", Outcome: "optional_equivalent"}}, "optional_equivalent"},
		{"table host*", validation.FieldCatalog{Fields: []string{"host"}, OptionalFields: []string{"hostname"}}, []validation.Match{{Name: "host", Binding: "source", Outcome: "matching"}, {Name: "hostname", Binding: "source", Outcome: "optional_equivalent"}}, "matching"},
		{"eval label=host | table *", validation.FieldCatalog{OptionalFields: []string{"host"}}, []validation.Match{{Name: "host", Binding: "source", Outcome: "optional_equivalent"}, {Name: "label", Binding: "derived", Outcome: "matching"}}, "matching"},
		{"eval host=1 | table host", validation.FieldCatalog{OptionalFields: []string{"host"}}, []validation.Match{{Name: "host", Binding: "derived", Outcome: "matching"}}, "matching"},
	} {
		r, err := validation.Validate(analysis.QueryDocument{Text: tc.query}, tc.catalog)
		if err != nil {
			t.Fatal(err)
		}
		last := r.Outcomes[len(r.Outcomes)-1]
		if last.Outcome != tc.aggregate || !reflect.DeepEqual(last.Matches, tc.want) {
			t.Errorf("%s evidence = %+v; want %+v", tc.query, last, tc.want)
		}
	}
	report, err := validation.Validate(analysis.QueryDocument{Text: "search missing=x | mystery | table host"}, validation.FieldCatalog{Fields: []string{"host"}})
	if err != nil {
		t.Fatal(err)
	}
	if !report.Coverage.SyntaxComplete || report.Coverage.SemanticComplete || report.Coverage.SchemaComplete || report.Status != analysis.Invalid {
		t.Fatalf("invalid incomplete coverage %+v", report.Coverage)
	}
	if !reflect.DeepEqual(report.Coverage.Reasons, []string{"SPL_UNKNOWN_FIELD", "SPL_UNSUPPORTED_COMMAND", "SPL_INDETERMINATE_FIELD"}) {
		t.Fatalf("reasons %+v", report.Coverage.Reasons)
	}
	source, err := analysis.AnalyzeWithSourceFields(report.Analysis.Document, []string{"host"})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(report.Analysis, source.Result) {
		t.Fatal("embedded analysis was modified")
	}
	if report.Analysis.Status != analysis.Incomplete || len(report.Analysis.Diagnostics) != 1 {
		t.Fatal("report diagnostics leaked into analysis")
	}
	if report.Diagnostics[0].Severity != "error" || report.Diagnostics[0].Category != "unknown_field" || report.Diagnostics[2].Severity != "warning" || report.Diagnostics[2].Category != "schema_ambiguity" || !strings.Contains(report.Diagnostics[2].Message, "could not prove") {
		t.Fatalf("validation diagnostics %+v", report.Diagnostics)
	}
}

func TestCatalogCaseWhitespaceAndLiteralStars(t *testing.T) {
	catalog := validation.FieldCatalog{Fields: []string{"Host", " first name ", "a*b"}}
	r, err := validation.Validate(analysis.QueryDocument{Text: "table Host host ' first name ' 'a*b'"}, catalog)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"matching", "missing", "matching", "matching"}
	for i, outcome := range r.Outcomes {
		if outcome.Outcome != want[i] {
			t.Errorf("outcome %d: %+v", i, outcome)
		}
	}
	if len(r.Outcomes) != len(want) {
		t.Fatalf("outcomes %+v", r.Outcomes)
	}
	if !reflect.DeepEqual(r.Target.Fields, []string{" first name ", "Host", "a*b"}) {
		t.Fatalf("names normalized destructively: %+v", r.Target)
	}
}

func TestValidationArraysAndDiagnosticIsolation(t *testing.T) {
	for _, query := range []string{"eval answer=1 | table answer", "fields -absent*", " ", "| mystery | table *"} {
		r, err := validation.Validate(analysis.QueryDocument{Text: query}, validation.FieldCatalog{})
		if err != nil {
			t.Fatal(err)
		}
		data, err := json.Marshal(r)
		if err != nil {
			t.Fatal(err)
		}
		if bytes.Contains(data, []byte(":null")) {
			t.Fatalf("null value: %s", data)
		}
	}
	r, err := validation.Validate(analysis.QueryDocument{Text: "fields host | table user"}, validation.FieldCatalog{Fields: []string{"host", "user"}})
	if err != nil {
		t.Fatal(err)
	}
	r.Diagnostics[0].Message = "caller edit"
	if r.Analysis.Diagnostics[0].Message == "caller edit" {
		t.Fatal("report findings alias embedded analysis findings")
	}
}
