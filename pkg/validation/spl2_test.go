package validation

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
)

func TestSPL2FieldListCanonical(t *testing.T) {
	catalog := FieldCatalog{Fields: []string{"host", "bytes"}, OptionalFields: []string{"optional"}, Identity: "task8", Version: "1"}
	for _, tt := range []struct {
		text     string
		status   analysis.Status
		complete bool
		outcomes []string
	}{
		{`FROM main SELECT host`, analysis.Valid, true, []string{"matching"}},
		{`FROM main SELECT absent`, analysis.Invalid, true, []string{"missing"}},
		{`FROM main | eval output=bytes | table output`, analysis.Valid, true, []string{"matching", "matching"}},
		{`FROM main | eval output=null | table output`, analysis.Invalid, true, []string{"unavailable"}},
		{`FROM main | eval output=null, answer=isnull(output) | table answer`, analysis.Valid, true, []string{"matching"}},
		{`FROM main | eval answer=isnull(absent) | where absent>0`, analysis.Invalid, true, []string{"missing"}},
		{`FROM main | table absent | mystery x=1`, analysis.Invalid, false, []string{"missing"}},
		{`FROM main | fields 'b*'`, analysis.Valid, true, []string{"matching"}},
		{`FROM main | mystery x=1 | fields 'b*'`, analysis.Incomplete, false, []string{"indeterminate"}},
		{`FROM main | table optional`, analysis.Valid, true, []string{"optional_equivalent"}},
		{`FROM [{host:"a"}] | table host`, analysis.Valid, true, []string{"matching"}},
		{`FROM main | eval stable=1 | append [FROM child | table absent] | appendpipe [table stable]`, analysis.Invalid, false, []string{"missing", "matching"}},
	} {
		t.Run(tt.text, func(t *testing.T) {
			doc := analysis.QueryDocument{Text: tt.text, Language: "spl2", SourceID: " exact\r\n "}
			r, err := Validate(doc, catalog)
			if err != nil {
				t.Fatal(err)
			}
			outcomes := []string{}
			for _, o := range r.Outcomes {
				outcomes = append(outcomes, o.Outcome)
				if o.Matches == nil {
					t.Fatal("nil matches")
				}
			}
			if r.Status != tt.status || r.Coverage.SchemaComplete != tt.complete || !reflect.DeepEqual(outcomes, tt.outcomes) || r.Analysis.Document.SourceID != doc.SourceID {
				t.Fatalf("canonical field report: %+v outcomes %v", r, outcomes)
			}
			raw, _ := json.Marshal(Request{Document: doc, Catalog: catalog})
			decoded, err := DecodeRequest(raw)
			if err != nil {
				t.Fatal(err)
			}
			wrapped, err := Validate(decoded.Document, decoded.Catalog)
			if err != nil || !reflect.DeepEqual(r, wrapped) {
				t.Fatalf("strict wrapper parity: %v", err)
			}
		})
	}
}

func TestSPL2FunctionResultDomainValidation(t *testing.T) {
	for _, tc := range []struct {
		expression string
		status     analysis.Status
		outcome    string
	}{
		{`abs(len("abc"))`, analysis.Valid, "matching"},
		{`substr("abc",len("x"))`, analysis.Valid, "matching"},
		{`split("a:b",":")`, analysis.Valid, "matching"},
		{`lower(split("a:b",":"))`, analysis.Incomplete, "indeterminate"},
		{`lower(len("abc"))`, analysis.Incomplete, "indeterminate"},
		{`abs(len(null))`, analysis.Incomplete, "indeterminate"},
	} {
		t.Run(tc.expression, func(t *testing.T) {
			doc := analysis.QueryDocument{Text: "FROM main | eval n=" + tc.expression + " | table n", Language: "spl2", SourceID: "result-domain"}
			field, err := Validate(doc, FieldCatalog{Fields: []string{}})
			if err != nil {
				t.Fatal(err)
			}
			schema, err := ValidateSchema(doc, SchemaTarget{Kind: "json_schema", Schema: json.RawMessage(`false`)})
			if err != nil {
				t.Fatal(err)
			}
			complete := tc.status == analysis.Valid
			if field.Status != tc.status || field.Coverage.SchemaComplete != complete || len(field.Outcomes) != 1 || field.Outcomes[0].Outcome != tc.outcome || schema.Status != tc.status || schema.Coverage.SchemaComplete != complete || len(schema.Outcomes) != 1 || schema.Outcomes[0].Outcome != tc.outcome {
				t.Fatalf("validation presence: field=%+v schema=%+v", field, schema)
			}
			if field.Analysis.Status != analysis.Valid || !field.Analysis.Coverage.SemanticComplete || schema.Analysis.Status != analysis.Valid || !schema.Analysis.Coverage.SemanticComplete {
				t.Fatal("validation uncertainty changed modeled analysis status/coverage")
			}
		})
	}
}

func TestSPL2FieldListMixedBatchAndInputs(t *testing.T) {
	docs := []analysis.QueryDocument{{Text: "table host", SourceID: "legacy"}, {Text: "FROM main SELECT host", Language: "spl2", SourceID: "spl2"}, {Text: "FROM main | mystery", Language: "spl2", SourceID: "unknown"}, {Text: "table absent", SourceID: "missing"}}
	catalog := FieldCatalog{Fields: []string{"host"}, OptionalFields: []string{}}
	r, err := ValidateBatch(docs, catalog)
	if err != nil || r.Status != analysis.Invalid || len(r.Reports) != 4 {
		t.Fatalf("batch: %+v %v", r, err)
	}
	for i, doc := range docs {
		single, e := Validate(doc, catalog)
		if e != nil || !reflect.DeepEqual(r.Reports[i], single) || r.Reports[i].Analysis.Document.SourceID != doc.SourceID {
			t.Fatal("mixed order/parity")
		}
	}
	raw, _ := json.Marshal(BatchRequest{Documents: docs, Catalog: catalog})
	decoded, err := DecodeBatchRequest(raw)
	if err != nil {
		t.Fatal(err)
	}
	wrapped, err := ValidateBatch(decoded.Documents, decoded.Catalog)
	if err != nil || !reflect.DeepEqual(r, wrapped) {
		t.Fatal("batch wrapper")
	}
	for _, bad := range []analysis.QueryDocument{{Text: "FROM main", Language: "SPL2"}, {Text: "FROM main", Language: "spl2", Profile: "edge"}, {Text: "FROM main", Language: "spl2", Version: "next"}, {Text: string([]byte{255}), Language: "spl2"}, {Text: "FROM main", SourceID: string([]byte{255})}} {
		if report, e := Validate(bad, catalog); report != nil || !IsInputError(e) {
			t.Fatalf("input classification: %+v %v", report, e)
		}
		if batch, e := ValidateBatch(append(append([]analysis.QueryDocument{}, docs...), bad), catalog); batch != nil || !IsInputError(e) {
			t.Fatalf("partial invalid batch: %+v %v", batch, e)
		}
	}
}
