package validation

import (
	"encoding/json"
	"fmt"
	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"
)

func TestJSONSchemaAllOfClosureExcludesSyntacticCandidate(t *testing.T) {
	target, err := DecodeSchemaTarget([]byte(`{"kind":"json_schema","schema":{"allOf":[{"type":"object","properties":{"a":true},"additionalProperties":false},{"properties":{"b":true}}]}}`))
	if err != nil {
		t.Fatal(err)
	}
	prepared, err := prepareJSONSchema(target)
	if err != nil {
		t.Fatal(err)
	}
	if got := prepared.project("b"); got.Outcome != "missing" || got.Admission != analysis.SourceFieldProhibited {
		t.Fatalf("b=%+v", got)
	}
	for _, name := range prepared.universe().Fields {
		if name == "b" {
			t.Fatal("allOf-prohibited syntactic name was enumerated")
		}
	}
}
func preparedJSON(t *testing.T, s string) *jsonSchemaTarget {
	t.Helper()
	target, e := DecodeSchemaTarget([]byte(`{"kind":"json_schema","schema":` + s + `}`))
	if e != nil {
		t.Fatal(e)
	}
	p, e := prepareJSONSchema(target)
	if e != nil {
		t.Fatal(e)
	}
	return p
}
func TestJSONSchemaProjectionVectors(t *testing.T) {
	tests := []struct {
		schema, path string
		want         analysis.SourceFieldAdmission
	}{
		{`{"type":"object","properties":{"a":{"type":"string"}},"additionalProperties":false}`, "a.x", analysis.SourceFieldProhibited},
		{`{"properties":{"a":{"type":"array","items":{"properties":{"x":true}}}}}`, "a.x", analysis.SourceFieldIndeterminate},
		{`{"properties":{"a":{"type":"array"}}}`, "a", analysis.SourceFieldAdmitted},
		{`{"properties":{"a.b":{"type":"string"},"a":{"properties":{"b":false}}}}`, "a.b", analysis.SourceFieldAdmitted},
		{`{"required":["x"],"additionalProperties":false}`, "x", analysis.SourceFieldProhibited},
		{`{"additionalProperties":{"type":"string"}}`, "unknown.x", analysis.SourceFieldProhibited},
		{`{"additionalProperties":{"type":"string"}}`, "unknown", analysis.SourceFieldAdmitted},
		{`{"anyOf":[{"properties":{"x":true},"additionalProperties":false},{"additionalProperties":false}]}`, "x", analysis.SourceFieldIndeterminate},
		{`{"oneOf":[{"properties":{"x":true}},{"properties":{"x":true}}]}`, "x", analysis.SourceFieldIndeterminate},
		{`{"properties":{"x":true},"oneOf":[{"properties":{"y":true}},{"properties":{"z":true}}]}`, "x", analysis.SourceFieldAdmitted},
		{`{"properties":{"host":{"type":"string"},"other":{"$ref":"https://absent.test/x"}}}`, "host", analysis.SourceFieldAdmitted},
		{`{"properties":{"host":true},"unevaluatedProperties":false}`, "host", analysis.SourceFieldAdmitted},
		{`{"unevaluatedProperties":false}`, "other", analysis.SourceFieldIndeterminate},
		{`{"then":false,"else":false}`, "x", analysis.SourceFieldAdmitted},
		{`{"properties":{"x":true},"dependentRequired":{"x":["y"]}}`, "x", analysis.SourceFieldAdmitted},
		{`{"not":{"required":["x"]}}`, "x", analysis.SourceFieldIndeterminate},
	}
	for _, tt := range tests {
		p := preparedJSON(t, tt.schema)
		if got := p.project(tt.path); got.Admission != tt.want {
			t.Errorf("%s %s: %+v want %v", tt.schema, tt.path, got, tt.want)
		}
	}
}
func TestJSONSchemaRequiredness(t *testing.T) {
	p := preparedJSON(t, `{"type":"object","properties":{"a":{"type":"object","required":["b"],"properties":{"b":{"type":"string"}}}},"additionalProperties":false}`)
	got := p.project("a.b")
	foundRequired, foundOptional := false, false
	for _, e := range got.Evidence {
		foundRequired = foundRequired || e.Requirement == "required"
		foundOptional = foundOptional || e.Requirement == "optional"
	}
	if !foundRequired || !foundOptional {
		t.Fatalf("ancestor requiredness lost: %+v", got)
	}
}
func TestJSONSchemaRequirementAndLiteralDetails(t *testing.T) {
	for _, tt := range []struct{ s, path, out string }{
		{`{"type":"object","required":["x"],"additionalProperties":true}`, "x", "required"},
		{`{"type":"object","required":["x"]}`, "x", "required"},
		{`{"type":"object","properties":{"a.b":{"type":"string"}},"additionalProperties":false}`, "a.b", "optional"},
		{`{"properties":{"a.b":{"type":"string"}}}`, "a.b", "indeterminate"},
	} {
		p := preparedJSON(t, tt.s)
		if got := p.project(tt.path); got.Outcome != tt.out {
			t.Errorf("%s: %+v want %s", tt.s, got, tt.out)
		}
	}
}
func TestJSONSchemaCompleteness(t *testing.T) {
	for _, tt := range []struct {
		s        string
		complete bool
	}{
		{`{"properties":{"x":{"type":"string"}},"additionalProperties":false}`, true},
		{`{"properties":{"a":{"properties":{"b":{"type":"integer"}},"additionalProperties":false}},"additionalProperties":false}`, true},
		{`{"properties":{"x":true},"additionalProperties":false}`, false},
		{`{"properties":{"x":{"type":"array"}},"additionalProperties":false}`, false},
		{`{"allOf":[{"properties":{"x":{"type":"string"}},"additionalProperties":false},{"properties":{"y":{"type":"string"}}}]}`, true},
	} {
		p := preparedJSON(t, tt.s)
		if p.universe().Complete != tt.complete {
			t.Fatalf("%s: %+v", tt.s, p.universe())
		}
	}
}
func TestJSONSchemaEnumerationDAGAndConcurrency(t *testing.T) {
	defs := map[string]any{}
	for i := 0; i < 13; i++ {
		ref := map[string]any{"$ref": "#/$defs/n" + strconv.Itoa(i+1)}
		defs["n"+strconv.Itoa(i)] = map[string]any{"type": "object", "properties": map[string]any{"a": ref, "b": ref}, "additionalProperties": false}
	}
	defs["n13"] = map[string]any{"type": "string"}
	raw, _ := json.Marshal(map[string]any{"$ref": "#/$defs/n0", "$defs": defs})
	target := SchemaTarget{Kind: "json_schema", Schema: raw}
	p, e := prepareJSONSchema(target)
	if e != nil {
		t.Fatal(e)
	}
	late := strings.TrimSuffix(strings.Repeat("b.", 13), ".")
	if p.enumerationWork != 4096 || p.universe().Complete || !slices.Contains(p.info().Limitations, "enumeration_budget") {
		t.Fatalf("budget: %d %+v", p.enumerationWork, p.info())
	}
	if slices.Contains(p.universe().Fields, late) {
		t.Fatal("late path should not be seeded")
	}
	if p.project(late).Admission != analysis.SourceFieldAdmitted {
		t.Fatal(p.project(late))
	}
	got, e := analysis.AnalyzeWithSourceUniverse(analysis.QueryDocument{Text: "table *"}, p.universe())
	if e != nil {
		t.Fatal(e)
	}
	if len(got.Expansions) != 1 || got.Expansions[0].Complete || got.Result.Status != analysis.Incomplete {
		t.Fatalf("expansion %+v", got)
	}
	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			q, e := prepareJSONSchema(target)
			if e != nil {
				t.Error(e)
				return
			}
			if !reflect.DeepEqual(p.universe().Fields, q.universe().Fields) || !reflect.DeepEqual(p.info(), q.info()) || !reflect.DeepEqual(p.project(late), q.project(late)) {
				t.Error("nondeterministic prepared target")
			}
			for j := 0; j < 10; j++ {
				if !reflect.DeepEqual(p.project(late), q.project(late)) {
					t.Error("concurrent projection changed")
				}
			}
		}()
	}
	wg.Wait()
}
func TestJSONSchemaManyPropertiesBound(t *testing.T) {
	props := map[string]any{}
	for i := 0; i < 5000; i++ {
		props[fmt.Sprintf("n%04d", i)] = map[string]any{"type": "string"}
	}
	raw, _ := json.Marshal(map[string]any{"properties": props, "additionalProperties": false})
	p, e := prepareJSONSchema(SchemaTarget{Kind: "json_schema", Schema: raw})
	if e != nil {
		t.Fatal(e)
	}
	if p.enumerationWork != 4096 || len(p.universe().Fields) > 4096 || p.universe().Complete {
		t.Fatal("enumeration bound")
	}
	if p.project("n4999").Admission != analysis.SourceFieldAdmitted {
		t.Fatal("truncated discovery poisoned exact resolution")
	}
}
func TestJSONSchemaProjectionSegmentBudget(t *testing.T) {
	p := preparedJSON(t, `{"properties":{"child":{"$ref":"#"}}}`)
	got := p.project(strings.Repeat("child.", 128) + "x")
	if got.Admission != analysis.SourceFieldIndeterminate || !slices.ContainsFunc(got.Evidence, func(e SchemaEvidence) bool { return e.Reason == "traversal_budget" }) {
		t.Fatal(got)
	}
}
func TestJSONSchemaIntermediateLiteralPaths(t *testing.T) {
	for _, tt := range []struct {
		s    string
		want analysis.SourceFieldAdmission
	}{
		{`{"properties":{"a.b":{"properties":{"c":{"type":"string"}},"additionalProperties":false}},"additionalProperties":false}`, analysis.SourceFieldAdmitted},
		{`{"properties":{"a.b":{"properties":{"c":{"type":"string"}}}}}`, analysis.SourceFieldIndeterminate},
		{`{"properties":{"a.b":{"properties":{"c":true}},"a":{"properties":{"b":{"properties":{"c":true}}}}}}`, analysis.SourceFieldIndeterminate},
		{`true`, analysis.SourceFieldAdmitted},
	} {
		p := preparedJSON(t, tt.s)
		if got := p.project("a.b.c"); got.Admission != tt.want {
			t.Errorf("%s: %+v", tt.s, got)
		}
	}
}
func TestJSONSchemaRequirednessUncertainMembershipCertain(t *testing.T) {
	p := preparedJSON(t, `{"type":"object","properties":{"x":{"type":"string"}},"dependentRequired":{"y":["x"]}}`)
	got := p.project("x")
	if got.Admission != analysis.SourceFieldAdmitted || got.Outcome != "indeterminate" {
		t.Fatalf("requiredness uncertainty must not fabricate optional certainty: %+v", got)
	}
}
func TestJSONSchemaConjunctiveRequirednessAndShape(t *testing.T) {
	for _, tt := range []struct {
		s, out string
		a      analysis.SourceFieldAdmission
	}{
		{`{"type":["object","string"],"required":["x"],"properties":{"x":true}}`, "indeterminate", analysis.SourceFieldAdmitted},
		{`{"type":["object","array"],"properties":{"x":true}}`, "indeterminate", analysis.SourceFieldIndeterminate},
		{`{"allOf":[{"type":"object"},{"required":["x"]},{"properties":{"x":{"type":"string"}}}]}`, "required", analysis.SourceFieldAdmitted},
		{`{"type":"object","required":["x"],"properties":{"x":true},"dependentRequired":{"y":["x"]}}`, "required", analysis.SourceFieldAdmitted},
	} {
		p := preparedJSON(t, tt.s)
		if got := p.project("x"); got.Outcome != tt.out || got.Admission != tt.a {
			t.Errorf("%s: %+v", tt.s, got)
		}
	}
}
func TestJSONSchemaConjunctiveNestedRequiredness(t *testing.T) {
	p := preparedJSON(t, `{"allOf":[{"type":"object","required":["a"]},{"additionalProperties":{"type":"object","properties":{"b":{"type":"string"}},"required":["b"]}}]}`)
	if got := p.project("a.b"); got.Outcome != "required" {
		t.Fatal(got)
	}
}
func TestJSONSchemaProjectionStateBudget(t *testing.T) {
	nodes := make([]any, 4100)
	for i := range nodes {
		nodes[i] = map[string]any{"properties": map[string]any{"x": map[string]any{"type": "string"}}}
	}
	raw, _ := json.Marshal(map[string]any{"allOf": nodes})
	p, e := prepareJSONSchema(SchemaTarget{Kind: "json_schema", Schema: raw})
	if e != nil {
		t.Fatal(e)
	}
	got := p.project("x")
	if got.Admission != analysis.SourceFieldIndeterminate || !slices.ContainsFunc(got.Evidence, func(e SchemaEvidence) bool { return e.Reason == "traversal_budget" }) {
		t.Fatal("missing state limit")
	}
}
func TestJSONSchemaBlankPropertyDoesNotEscapeUniverse(t *testing.T) {
	p := preparedJSON(t, `{"properties":{"":{"type":"string"}," ":{"type":"string"},"host":{"type":"string"}},"additionalProperties":false}`)
	if p.project("host").Admission != analysis.SourceFieldAdmitted {
		t.Fatal("unrelated exact path poisoned")
	}
	u := p.universe()
	if u.Complete || !slices.Contains(p.info().Limitations, "unrepresentable_source_name") {
		t.Fatalf("blank name omitted with false completeness: %+v", p.info())
	}
	for _, name := range u.Fields {
		if !validSchemaName(name) {
			t.Fatal("invalid blank seed")
		}
	}
	got, e := analysis.AnalyzeWithSourceUniverse(analysis.QueryDocument{Text: "table *"}, u)
	if e != nil {
		t.Fatal(e)
	}
	if got.Expansions[0].Complete {
		t.Fatal("blank name wildcard falsely complete")
	}
}
func TestJSONSchemaProjectionSharedConjunctionDAG(t *testing.T) {
	defs := map[string]any{}
	for i := 0; i < 13; i++ {
		ref := map[string]any{"$ref": "#/$defs/n" + strconv.Itoa(i+1)}
		defs["n"+strconv.Itoa(i)] = map[string]any{"allOf": []any{ref, ref}}
	}
	defs["n13"] = map[string]any{"properties": map[string]any{"x": map[string]any{"type": "string"}}}
	raw, _ := json.Marshal(map[string]any{"$ref": "#/$defs/n0", "$defs": defs})
	p, e := prepareJSONSchema(SchemaTarget{Kind: "json_schema", Schema: raw})
	if e != nil {
		t.Fatal(e)
	}
	got := p.project("x")
	if got.Admission != analysis.SourceFieldAdmitted || len(got.Evidence) > 1024 {
		t.Fatalf("repeated projection evidence expanded: admission %d, evidence %d", got.Admission, len(got.Evidence))
	}
}
func TestJSONSchemaAnyOfUnspecifiedConclusion(t *testing.T) {
	p := preparedJSON(t, `{"anyOf":[{"properties":{"x":{"type":"string"}}},true]}`)
	if got := p.project("x"); got.Outcome != "permitted_unspecified" {
		t.Fatal(got)
	}
}
