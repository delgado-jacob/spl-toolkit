package validation_test

import (
	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/validation"
	"testing"
)

func TestBatchValidation(t *testing.T) {
	docs := []analysis.QueryDocument{{Text: "eval x=1 | table x", SourceID: "valid"}, {Text: "| mystery", SourceID: "incomplete"}, {Text: " ", SourceID: "invalid"}}
	batch, err := validation.ValidateBatch(docs, validation.FieldCatalog{})
	if err != nil {
		t.Fatal(err)
	}
	if batch.SchemaVersion != 1 || batch.Status != analysis.Invalid || len(batch.Reports) != 3 {
		t.Fatalf("batch = %+v", batch)
	}
	for i, r := range batch.Reports {
		if r.Analysis.Document.SourceID != docs[i].SourceID {
			t.Fatal("order changed")
		}
	}
	for n, want := range map[int]analysis.Status{1: analysis.Valid, 2: analysis.Incomplete} {
		b, err := validation.ValidateBatch(docs[:n], validation.FieldCatalog{})
		if err != nil || b.Status != want {
			t.Fatalf("%+v %v", b, err)
		}
	}
	docs[2].Language = "bad"
	if b, err := validation.ValidateBatch(docs, validation.FieldCatalog{}); b != nil || !validation.IsInputError(err) {
		t.Fatalf("partial batch or wrong error: %+v %v", b, err)
	}
	if b, err := validation.ValidateBatch(nil, validation.FieldCatalog{}); b != nil || !validation.IsInputError(err) {
		t.Fatalf("empty batch: %+v %v", b, err)
	}
	if b, err := validation.ValidateBatch(docs[:1], validation.FieldCatalog{Fields: []string{"x", "x"}}); b != nil || !validation.IsInputError(err) {
		t.Fatalf("bad catalog: %+v %v", b, err)
	}
}
