package validation

import (
	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"testing"
)

func TestJSONSchemaPatternVectors(t *testing.T) {
	for _, tt := range []struct {
		pattern, name string
		want          analysis.SourceFieldAdmission
	}{
		{`^host_[a-z]+$`, "host_abc", analysis.SourceFieldAdmitted}, {`^host_[a-z]+$`, "host_9", analysis.SourceFieldProhibited},
		{`ip_[0-9]{1,3}`, "prefix_ip_12suffix", analysis.SourceFieldAdmitted}, {`^a\\+b$`, "a+b", analysis.SourceFieldAdmitted},
		{`(?=host)host`, "x", analysis.SourceFieldIndeterminate}, {`^host_[a-z]+$`, "host_é", analysis.SourceFieldIndeterminate}, {`^host_[a-z]+$`, "host_a\n", analysis.SourceFieldIndeterminate},
		{`a{1001}`, "b", analysis.SourceFieldIndeterminate},
	} {
		p := preparedJSON(t, `{"patternProperties":{"`+tt.pattern+`":{"type":"string"}},"additionalProperties":false}`)
		if got := p.project(tt.name); got.Admission != tt.want {
			t.Errorf("%s %q: %+v", tt.pattern, tt.name, got)
		}
	}
	p := preparedJSON(t, `{"patternProperties":{"^x":true,"x$":false},"additionalProperties":false}`)
	if p.project("x").Admission != analysis.SourceFieldProhibited {
		t.Fatal("patterns do not intersect")
	}
	for _, pattern := range []string{"[a-", "a{3,2}", "*a"} {
		target, e := DecodeSchemaTarget([]byte(`{"kind":"json_schema","schema":{"patternProperties":{"` + pattern + `":true}}}`))
		if e == nil {
			_, e = prepareJSONSchema(target)
		}
		if !IsInputError(e) {
			t.Errorf("malformed %q: %v", pattern, e)
		}
	}
}
func TestJSONSchemaECMAScriptBoundary(t *testing.T) {
	for _, pattern := range []string{"a{01}", "a{01,02}", "a{1001}", "[]", "a{", "a{,2}", "{}", "a*?", "a??"} {
		p, e := compileSchemaPattern(pattern)
		if e != nil || p.re != nil {
			t.Errorf("valid unsupported %q: %+v %v", pattern, p, e)
		}
	}
	for _, pattern := range []string{"a{2,1}", "a**"} {
		if _, e := compileSchemaPattern(pattern); !IsInputError(e) {
			t.Errorf("malformed %q accepted", pattern)
		}
	}
}
