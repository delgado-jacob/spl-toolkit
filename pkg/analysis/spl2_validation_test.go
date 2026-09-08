package analysis_test

import (
	"encoding/json"
	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/validation"
	"testing"
)

func TestSPL2CanonicalSchemaBoundaries(t *testing.T) {
	for _, tc := range []struct {
		text, schema string
		status       analysis.Status
		outcomes     []string
	}{
		{`FROM main | where 'actor.name'="a"`, `{"type":"object","properties":{"actor":{"type":"object","properties":{"name":true},"required":["name"],"additionalProperties":false}},"required":["actor"],"additionalProperties":false}`, analysis.Incomplete, []string{"indeterminate"}},
		{`FROM main | where actor.name="a"`, `{"type":"object","properties":{"actor.name":true},"required":["actor.name"],"additionalProperties":false}`, analysis.Invalid, []string{"missing", "indeterminate"}},
		{`FROM main | where actor.name="a"`, `{"type":"object","properties":{"actor":{"type":"object","properties":{"name":true},"required":["name"],"additionalProperties":false}},"required":["actor"],"additionalProperties":false}`, analysis.Incomplete, []string{"required", "indeterminate"}},
		{`FROM main | eval x=tonumber("17") | table x`, `false`, analysis.Incomplete, []string{"indeterminate"}},
		{`FROM main | eval x=tonumber("17")`, `false`, analysis.Valid, []string{}},
		{`FROM main | eval x=coalesce(tonumber("17"),0) | table x`, `false`, analysis.Valid, []string{"matching"}},
		{`FROM main | eval 'actor.name'=1 | table 'actor.name'`, `false`, analysis.Valid, []string{"matching"}},
	} {
		t.Run(tc.text+tc.schema, func(t *testing.T) {
			r, e := validation.ValidateSchema(analysis.QueryDocument{Text: tc.text, Language: "spl2"}, validation.SchemaTarget{Kind: "json_schema", Schema: json.RawMessage(tc.schema)})
			if e != nil {
				t.Fatal(e)
			}
			if r.Status != tc.status || len(r.Outcomes) != len(tc.outcomes) {
				t.Fatalf("%+v %+v", r, r.Outcomes)
			}
			for i, o := range r.Outcomes {
				if o.Outcome != tc.outcomes[i] || o.Outcome == "indeterminate" && (o.MatchesComplete || len(o.Matches) > 0) {
					t.Fatalf("%+v", r.Outcomes)
				}
			}
		})
	}
}

func TestSPL2NullTestValidation(t *testing.T) {
	for _, tc := range []struct {
		text     string
		status   analysis.Status
		outcomes int
	}{
		{`FROM main | eval answer=isnull(absent)`, analysis.Valid, 0},
		{`FROM main | eval answer=isnotnull((absent))`, analysis.Valid, 0},
		{`FROM main | eval x=1, answer=isnull(x)`, analysis.Valid, 0},
		{`FROM main | eval x=tonumber("17"), answer=isnull(x)`, analysis.Valid, 0},
		{`FROM main | eval x=null, answer=isnull(x)`, analysis.Valid, 0},
		{`FROM main | eval x=null, answer=isnull(x) | table x`, analysis.Invalid, 1},
		{`FROM main | eval answer=isnull(absent) | where absent>0`, analysis.Invalid, 1},
		{`FROM main | eval answer=isnull(absent+1)`, analysis.Invalid, 1},
		{`FROM main | eval answer=isnull(abs(absent))`, analysis.Invalid, 1},
		{`FROM main | where absent>0 | eval answer=isnull(absent)`, analysis.Invalid, 1},
	} {
		t.Run(tc.text, func(t *testing.T) {
			doc := analysis.QueryDocument{Text: tc.text, Language: "spl2"}
			flat, e := validation.Validate(doc, validation.FieldCatalog{})
			if e != nil {
				t.Fatal(e)
			}
			schema, e := validation.ValidateSchema(doc, validation.SchemaTarget{Kind: "json_schema", Schema: json.RawMessage(`false`)})
			if e != nil {
				t.Fatal(e)
			}
			if flat.Status != tc.status || schema.Status != tc.status || len(flat.Outcomes) != tc.outcomes || len(schema.Outcomes) != tc.outcomes {
				t.Fatalf("flat %+v schema %+v", flat, schema)
			}
		})
	}
	for _, schema := range []string{`{"type":"object","properties":{"host":true},"required":["host"],"additionalProperties":false}`, `{"type":"object","properties":{"host":true}}`} {
		doc := analysis.QueryDocument{Text: `FROM main | eval answer=isnull(host) | where host>0`, Language: "spl2"}
		r, e := validation.ValidateSchema(doc, validation.SchemaTarget{Kind: "json_schema", Schema: json.RawMessage(schema)})
		if e != nil || r.Status != analysis.Valid || len(r.Outcomes) != 1 {
			t.Fatalf("%+v %v", r, e)
		}
	}
}

func TestSPL2SchemaWildcardCandidateIdentity(t *testing.T) {
	schema := `{"type":"object","properties":{"actor":{"type":"object","properties":{"name":{"type":"string"}},"required":["name"],"additionalProperties":false},"host":{"type":"string"}},"required":["actor","host"],"additionalProperties":false}`
	for _, pattern := range []string{"actor.*", "*"} {
		text := `FROM main | eval 'actor.local'=1 | fields '` + pattern + `' | where 'actor.local'>0`
		r, e := validation.ValidateSchema(analysis.QueryDocument{Text: text, Language: "spl2"}, validation.SchemaTarget{Kind: "json_schema", Schema: json.RawMessage(schema)})
		if e != nil {
			t.Fatal(e)
		}
		if r.Status != analysis.Incomplete {
			t.Fatalf("%+v", r)
		}
		derived := false
		for _, outcome := range r.Outcomes {
			for _, match := range outcome.Matches {
				if match.Name == "actor.name" {
					t.Fatalf("nested path became literal-name match %+v", r)
				}
				derived = derived || match.Name == "actor.local" && match.Binding == "derived"
			}
		}
		if !derived {
			t.Fatalf("lost derived proof %+v", r)
		}
	}
}
