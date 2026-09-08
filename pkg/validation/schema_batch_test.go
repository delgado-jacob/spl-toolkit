package validation

import (
	"encoding/json"
	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"reflect"
	"testing"
)

func TestSchemaBatchPreparesOnceAndOrders(t *testing.T) {
	target := SchemaTarget{Kind: "json_schema", Schema: json.RawMessage(`{"type":"object","properties":{"host":true},"additionalProperties":false}`)}
	docs := []analysis.QueryDocument{{Text: "table host"}, {Text: "| mystery"}, {Text: "table absent"}}
	calls := 0
	b, e := validateSchemaBatchWith(docs, target, func(t SchemaTarget) (preparedSchemaTarget, error) { calls++; return prepareSchemaTarget(t) })
	if e != nil || calls != 1 || b.Status != analysis.Invalid || len(b.Reports) != 3 {
		t.Fatalf("%+v %v calls=%d", b, e, calls)
	}
	for i, d := range docs {
		r, e := ValidateSchema(d, target)
		if e != nil || !reflect.DeepEqual(r, b.Reports[i]) {
			t.Fatalf("single/batch %d: %v", i, e)
		}
	}
	for _, ds := range [][]analysis.QueryDocument{nil, {{Text: "x", Language: "sql"}}} {
		if b, e := ValidateSchemaBatch(ds, target); b != nil || !IsInputError(e) {
			t.Fatalf("%+v %v", b, e)
		}
	}
	if b, e := ValidateSchemaBatch(docs, SchemaTarget{Kind: "json_schema", Schema: json.RawMessage(`null`)}); b != nil || !IsInputError(e) {
		t.Fatalf("%+v %v", b, e)
	}
}
