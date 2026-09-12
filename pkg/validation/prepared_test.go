package validation

import (
	"encoding/json"
	"reflect"
	"sync"
	"testing"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
)

func TestPreparedFieldCatalogParityAndOwnership(t *testing.T) {
	input := FieldCatalog{Fields: []string{"host"}, OptionalFields: []string{"user"}, Identity: "local"}
	prepared, err := PrepareFieldCatalog(input)
	if err != nil {
		t.Fatal(err)
	}
	input.Fields[0] = "changed"
	input.OptionalFields[0] = "changed-too"
	doc := analysis.QueryDocument{Text: "search host=web user=alice"}
	want, err := Validate(doc, FieldCatalog{Fields: []string{"host"}, OptionalFields: []string{"user"}, Identity: "local"})
	if err != nil {
		t.Fatal(err)
	}
	got, err := prepared.Validate(doc)
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("prepared mismatch: %v\n%+v\n%+v", err, got, want)
	}
	got.Target.Fields[0] = "corrupted"
	got.Target.OptionalFields[0] = "corrupted"
	again, err := prepared.Validate(doc)
	if err != nil || !reflect.DeepEqual(again, want) {
		t.Fatalf("report mutation escaped: %v\n%+v", err, again)
	}
}

func TestPreparedSchemaTargetParityAndConcurrentUse(t *testing.T) {
	target := SchemaTarget{Kind: "json_schema", Schema: json.RawMessage(`{"type":"object","properties":{"host":{"type":"string"}}}`)}
	prepared, err := PrepareSchemaTarget(target)
	if err != nil {
		t.Fatal(err)
	}
	target.Schema[0] = ' '
	doc := analysis.QueryDocument{Text: "search host=web"}
	want, err := ValidateSchema(doc, SchemaTarget{Kind: "json_schema", Schema: json.RawMessage(`{"type":"object","properties":{"host":{"type":"string"}}}`)})
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	errors := make(chan string, 12)
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			got, err := prepared.Validate(doc)
			if err != nil || !reflect.DeepEqual(got, want) {
				errors <- "concurrent prepared schema differed"
			}
		}()
	}
	wg.Wait()
	close(errors)
	for failure := range errors {
		t.Fatal(failure)
	}
}

func TestPreparedNilAndZeroHandles(t *testing.T) {
	doc := analysis.QueryDocument{Text: "search host=web"}
	for _, p := range []*PreparedFieldCatalog{nil, {}} {
		if _, err := p.Validate(doc); !IsInputError(err) {
			t.Fatalf("field handle: %v", err)
		}
	}
	for _, p := range []*PreparedSchemaTarget{nil, {}} {
		if _, err := p.Validate(doc); !IsInputError(err) {
			t.Fatalf("schema handle: %v", err)
		}
	}
}
