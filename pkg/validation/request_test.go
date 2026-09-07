package validation_test

import (
	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/validation"
	"reflect"
	"testing"
)

func TestCatalogDecoding(t *testing.T) {
	for _, input := range []string{`[]`, `{"fields":[]}`, `{"fields":["host","Host","user.name"," user ","*"],"optional_fields":["optional"],"identity":"local","version":"v1"}`, `["\ud83d\ude00","\\ud800"]`} {
		if _, err := validation.DecodeFieldCatalog([]byte(input)); err != nil {
			t.Errorf("%s: %v", input, err)
		}
	}
	for _, input := range []string{`null`, `{}`, `{"fields":null}`, `{"fields":[] ,"extra":true}`, `{"fields":[],"fields":[]}`, `{"fields":[],"f\u0069elds":[]}`, `["host","host"]`, `[""]`, `["  "]`, `[1]`, `[null]`, `{"fields":[],"optional_fields":null}`, `{"fields":[],"optional_fields":["x","x"]}`, `{"fields":["x"],"optional_fields":["x"]}`, `{"fields":[],"identity":null}`, `{"fields":[],"version":1}`, `[] []`, `["\ud800"]`, `["\udc00"]`, "[\"\xff\"]"} {
		if _, err := validation.DecodeFieldCatalog([]byte(input)); !validation.IsInputError(err) {
			t.Errorf("%q: expected InputError, got %v", input, err)
		}
	}
	got, err := validation.DecodeFieldCatalog([]byte(`{"fields":["z","a"],"optional_fields":["y","b"]}`))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, validation.FieldCatalog{Fields: []string{"a", "z"}, OptionalFields: []string{"b", "y"}}) {
		t.Fatalf("catalog = %+v", got)
	}
	empty, _ := validation.DecodeFieldCatalog([]byte(`[]`))
	if empty.Fields == nil || empty.OptionalFields == nil {
		t.Fatal("null arrays")
	}
}

func TestRequestDecoding(t *testing.T) {
	good := []string{`{"document":{"text":""},"catalog":[]}`, `{"document":{"text":"search host=x","source_id":"\ud83d\ude00"},"catalog":{"fields":["host"]}}`}
	for _, input := range good {
		r, err := validation.DecodeRequest([]byte(input))
		if err != nil {
			t.Errorf("%s: %v", input, err)
		} else if r.Document.Language != "spl" || r.Document.Profile != "splunkd" || r.Document.Version != "current" {
			t.Fatalf("defaults: %+v", r)
		}
	}
	for _, input := range []string{`null`, `{}`, `{"document":{"text":"x"}}`, `{"document":{"text":"x"},"catalog":[],"extra":0}`, `{"document":{"text":"x"},"document":{"text":"y"},"catalog":[]}`, `{"document":null,"catalog":[]}`, `{"document":{},"catalog":[]}`, `{"document":{"text":null},"catalog":[]}`, `{"document":{"text":1},"catalog":[]}`, `{"document":{"text":"x","text":"y"},"catalog":[]}`, `{"document":{"text":"x","extra":"y"},"catalog":[]}`, `{"document":{"text":"x","language":null},"catalog":[]}`, `{"document":{"text":"x","profile":false},"catalog":[]}`, `{"document":{"text":"x","version":"old"},"catalog":[]}`, `{"document":{"text":"x","source_id":"\ud800"},"catalog":[]}`, `{"document":{"text":"x"},"catalog":[]} true`} {
		if _, err := validation.DecodeRequest([]byte(input)); !validation.IsInputError(err) {
			t.Errorf("%s: expected InputError, got %v", input, err)
		}
	}
}

func TestDocumentsAndBatchDecoding(t *testing.T) {
	for _, input := range []string{`null`, `[]`, `{}`, `[null]`, `[{}]`, `[{"text":"x","language":"sql"}]`, `[{"text":"x"}] null`, `[{"text":"\ud800"}]`} {
		if _, err := validation.DecodeDocuments([]byte(input)); !validation.IsInputError(err) {
			t.Errorf("%s: %v", input, err)
		}
	}
	docs, err := validation.DecodeDocuments([]byte(`[{"text":" "},{"text":"search host=x"}]`))
	if err != nil || len(docs) != 2 {
		t.Fatalf("%+v %v", docs, err)
	}
	r, err := validation.DecodeBatchRequest([]byte(`{"documents":[{"text":"x"}],"catalog":[]}`))
	if err != nil || len(r.Documents) != 1 {
		t.Fatalf("%+v %v", r, err)
	}
	for _, input := range []string{`{}`, `{"documents":[],"catalog":[]}`, `{"documents":[{"text":"x"}],"catalog":null}`, `{"documents":[{"text":"x"}],"catalog":[],"extra":0}`} {
		if _, err := validation.DecodeBatchRequest([]byte(input)); !validation.IsInputError(err) {
			t.Errorf("%s: %v", input, err)
		}
	}
}

func TestDirectInputErrors(t *testing.T) {
	for _, catalog := range []validation.FieldCatalog{{Fields: []string{"x", "x"}}, {Fields: []string{" "}}, {Fields: []string{"\xff"}}, {Fields: []string{"x"}, OptionalFields: []string{"x"}}, {Identity: "\xff"}, {Version: "\xff"}} {
		if r, err := validation.Validate(analysis.QueryDocument{Text: "search x=1"}, catalog); r != nil || !validation.IsInputError(err) {
			t.Fatalf("%+v %v", r, err)
		}
	}
	for _, doc := range []analysis.QueryDocument{{Text: "x", Language: "sql"}, {Text: "x", Profile: "other"}, {Text: "x", Version: "old"}, {Text: "\xff"}, {Text: "x", SourceID: "\xff"}} {
		if r, err := validation.Validate(doc, validation.FieldCatalog{}); r != nil || !validation.IsInputError(err) {
			t.Fatalf("%+v %v", r, err)
		}
	}
}
