package validation

import (
	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"strings"
	"testing"
)

func TestJSONSchemaLocalResources(t *testing.T) {
	raw := `{"kind":"json_schema","schema":{"$id":"https://schemas.example.test/event","properties":{"actor":{"$ref":"user#/$defs/user"}},"additionalProperties":false},"resources":{"https://schemas.example.test/user":{"$defs":{"user":{"type":"object","properties":{"name":{"type":"string"}},"additionalProperties":false}}}}}`
	target, e := DecodeSchemaTarget([]byte(raw))
	if e != nil {
		t.Fatal(e)
	}
	p, e := prepareJSONSchema(target)
	if e != nil {
		t.Fatal(e)
	}
	if got := p.project("actor.name"); got.Admission != analysis.SourceFieldAdmitted {
		t.Fatal(got)
	}
	if got := p.project("actor.no"); got.Admission != analysis.SourceFieldProhibited {
		t.Fatal(got)
	}
}
func TestJSONSchemaResourceVectors(t *testing.T) {
	for _, s := range []string{`{"$ref":"#/$defs/a~1b","$defs":{"a/b":{"properties":{"x":{"type":"string"}}}}}`, `{"$ref":"#thing","$defs":{"a":{"$anchor":"thing","properties":{"x":{"type":"string"}}}}}`, `{"$id":"https://s.test/root","$ref":"child","$defs":{"a":{"$id":"child","properties":{"x":{"type":"string"}}}}}`, `{"$ref":"#/%24defs/a~0b","$defs":{"a~b":{"properties":{"x":{"type":"string"}}}}}`} {
		p := preparedJSON(t, s)
		if p.project("x").Admission != analysis.SourceFieldAdmitted {
			t.Fatal(s, p.project("x"))
		}
	}
	p := preparedJSON(t, `{"$ref":"#","properties":{"x":true}}`)
	if p.project("z").Admission != analysis.SourceFieldIndeterminate {
		t.Fatal("cycle")
	}
	p = preparedJSON(t, `{"type":"object","properties":{"name":{"type":"string"},"child":{"$ref":"#"}},"additionalProperties":false}`)
	if p.project("child.child.name").Admission != analysis.SourceFieldAdmitted {
		t.Fatal(p.project("child.child.name"))
	}
	p = preparedJSON(t, `{"$ref":"missing","properties":{"host":{"type":"string"}}}`)
	if p.project("host").Admission != analysis.SourceFieldIndeterminate {
		t.Fatal("containing ref must affect host")
	}
	p = preparedJSON(t, `{"$ref":"#/$defs/open","additionalProperties":false,"$defs":{"open":{"properties":{"x":true}}}}`)
	if p.project("x").Admission != analysis.SourceFieldProhibited {
		t.Fatal("ref siblings must intersect")
	}
}
func TestJSONSchemaResourceInputErrors(t *testing.T) {
	for _, s := range []string{
		`{"$defs":{"x":{"$schema":"http://json-schema.org/draft-07/schema#"}}}`, `{"$defs":{"x":{"$id":"https://s.test/x"},"y":{"$id":"https://s.test/x"}}}`, `{"$defs":{"x":{"$anchor":"a"},"y":{"$anchor":"a"}}}`, `{"$ref":"#/examples/0","examples":[true]}`, `{"$ref":"#/$defs/a~2b"}`, `{"$ref":"#/%GG"}`, `{"type":["object","bogus"]}`, `{"required":["a","a"]}`, `{"minProperties":-1}`, `{"properties":{"x":null}}`,
	} {
		target, e := DecodeSchemaTarget([]byte(`{"kind":"json_schema","schema":` + s + `}`))
		if e == nil {
			_, e = prepareJSONSchema(target)
		}
		if !IsInputError(e) {
			t.Errorf("want input error %s: %v", s, e)
		}
	}
	p := preparedJSON(t, `{"examples":[{"$id":"https://s.test/fake"}],"$ref":"https://s.test/fake"}`)
	if p.project("x").Admission != analysis.SourceFieldIndeterminate {
		t.Fatal("annotation indexed")
	}
}
func TestJSONSchemaExplicitRootIdentity(t *testing.T) {
	p := preparedJSON(t, `{"$id":"https://s.test/root"}`)
	for _, uri := range p.info().ResourceURIs {
		if uri == jsonSchemaRoot {
			t.Fatal("fabricated fallback identity alongside explicit root id")
		}
	}
}
func TestJSONSchemaRecognizedKeywordShapes(t *testing.T) {
	for _, s := range []string{`{"uniqueItems":"yes"}`, `{"minimum":"low"}`, `{"multipleOf":0}`, `{"dependentSchemas":[]}`, `{"$vocabulary":{"https://json-schema.org/draft/2020-12/vocab/made-up":true},"properties":{"x":true}}`} {
		target, e := DecodeSchemaTarget([]byte(`{"kind":"json_schema","schema":` + s + `}`))
		var p *jsonSchemaTarget
		if e == nil {
			p, e = prepareJSONSchema(target)
		}
		if strings.Contains(s, "made-up") {
			if e != nil {
				t.Fatal(e)
			}
			if p.project("x").Admission != analysis.SourceFieldIndeterminate {
				t.Fatal("unknown vocabulary treated supported")
			}
		} else if !IsInputError(e) {
			t.Errorf("invalid shape accepted %s: %v", s, e)
		}
	}
	p := preparedJSON(t, `{"minLength":1.0,"maxItems":1e2,"properties":{"x":true}}`)
	if p.project("x").Admission != analysis.SourceFieldAdmitted {
		t.Fatal("integral number rejected")
	}
}
func TestJSONSchemaRelativeIDDoesNotResolveAgainstAncestor(t *testing.T) {
	p := preparedJSON(t, `{"properties":{"a":{"$id":"relative","$ref":"#","properties":{"x":true}}},"additionalProperties":false}`)
	got := p.project("a.x")
	if got.Admission != analysis.SourceFieldIndeterminate {
		t.Fatalf("relative resource inherited unrelated root: %+v", got)
	}
}
