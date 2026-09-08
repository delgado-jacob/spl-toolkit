package validation

import (
	"encoding/json"
	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"reflect"
	"testing"
)

func TestDecodeSchemaTargetStrict(t *testing.T) {
	bad := []string{
		`{"kind":"json_schema","schema":true,"catalog":null}`, `{"kind":"json_schema","schema":{"examples":[{"a":1,"a":2}]}}`,
		`{"kind":"json_schema","schema":null}`, `{"kind":"json_schema","schema":{"type":"wat"}}`,
		`{"kind":"ocsf","catalog":{},"selection":{"version":"1","class":"x","profiles":null}}`,
		`{"kind":"ocsf","catalog":{},"selection":{"version":"1","class_uid":1.0}}`,
		`{"kind":"ocsf","catalog":{},"selection":{"version":"1","class":"x","category":"y"}}`,
		`{"kind":"ocsf","catalog":{},"selection":{"version":"1","class":"x","profiles":["a","a"]}}`,
	}
	for _, raw := range bad {
		target, e := DecodeSchemaTarget([]byte(raw))
		if e == nil && target.Kind == "json_schema" {
			_, e = prepareJSONSchema(target)
		}
		if !IsInputError(e) {
			t.Errorf("expected InputError: %s, %v", raw, e)
		}
	}
	target, e := DecodeSchemaTarget([]byte(`{"kind":"ocsf","catalog":{},"selection":{"version":"1","class_uid":9223372036854775807}}`))
	if e != nil {
		t.Fatal(e)
	}
	if target.Selection.Profiles == nil || target.Selection.Extensions == nil {
		t.Fatal("nil normalized arrays")
	}
}
func TestJSONSchemaDefensiveCopies(t *testing.T) {
	raw := json.RawMessage(`{"properties":{"x":{"type":"string"}},"additionalProperties":false}`)
	target := SchemaTarget{Kind: "json_schema", Schema: raw}
	before := append([]byte{}, raw...)
	p, e := prepareJSONSchema(target)
	if e != nil {
		t.Fatal(e)
	}
	if !reflect.DeepEqual([]byte(raw), before) {
		t.Fatal("mutated input")
	}
	for i := range raw {
		raw[i] = ' '
	}
	if p.project("x").Admission != analysis.SourceFieldAdmitted {
		t.Fatal("retained raw input")
	}
	u := p.universe()
	u.Fields[0] = "broken"
	if p.universe().Fields[0] != "x" {
		t.Fatal("shared fields")
	}
}
func TestJSONSchemaResourceMapOwnership(t *testing.T) {
	target := SchemaTarget{Kind: "json_schema", Schema: json.RawMessage(`{"$ref":"https://s.test/external"}`), Resources: map[string]json.RawMessage{"https://s.test/external": json.RawMessage(`{"properties":{"host":{"type":"string"}},"additionalProperties":false}`)}}
	before, _ := json.Marshal(target)
	p, e := prepareJSONSchema(target)
	if e != nil {
		t.Fatal(e)
	}
	after, _ := json.Marshal(target)
	if !reflect.DeepEqual(before, after) {
		t.Fatal("preparation mutated caller map")
	}
	target.Resources["https://s.test/external"][0] = ' '
	delete(target.Resources, "https://s.test/external")
	expected := p.project("host")
	if expected.Admission != analysis.SourceFieldAdmitted {
		t.Fatal(expected)
	}
	expected.Evidence[0].Reason = "mutated"
	if p.project("host").Evidence[0].Reason == "mutated" {
		t.Fatal("shared projection evidence")
	}
	info := p.info()
	info.ResourceURIs[0] = "changed"
	if p.info().ResourceURIs[0] == "changed" {
		t.Fatal("shared metadata")
	}
}
func TestTypedSchemaTargetSemanticRejection(t *testing.T) {
	for _, target := range []SchemaTarget{
		{Kind: "json_schema", Schema: json.RawMessage(`true`), Selection: &OCSFSelection{}},
		{Kind: "json_schema", Schema: json.RawMessage(`true`), Catalog: json.RawMessage(`null`)},
		{Kind: "json_schema", Schema: json.RawMessage(`{"a":1,"a":2}`)},
		{Kind: "json_schema", Schema: json.RawMessage(`true`), Identity: string([]byte{0xff})},
	} {
		if _, e := prepareJSONSchema(target); !IsInputError(e) {
			t.Errorf("invalid typed target accepted: %v", e)
		}
	}
}
