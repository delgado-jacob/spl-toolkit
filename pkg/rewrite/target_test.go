package rewrite

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/delgado-jacob/spl-toolkit/pkg/validation"
)

func TestDecodeRewriteValidationTarget(t *testing.T) {
	for _, tc := range []struct {
		name, target string
		valid        bool
	}{
		{"field catalog object", `{"kind":"field_list","catalog":{"fields":["user"],"optional_fields":["extra"]}}`, true},
		{"field catalog array", `{"kind":"field_list","catalog":["user"]}`, true},
		{"json schema", `{"kind":"json_schema","identity":"dest","schema":{"type":"object","properties":{"user":{"type":"string"}}}}`, true},
		{"ocsf", `{"kind":"ocsf","catalog":{},"selection":{"version":"1.0","class":"event"}}`, true},
		{"null", `null`, false}, {"unknown", `{"kind":"other"}`, false},
		{"missing catalog", `{"kind":"field_list"}`, false}, {"null catalog", `{"kind":"field_list","catalog":null}`, false},
		{"duplicate fields", `{"kind":"field_list","catalog":["a","a"]}`, false},
		{"catalog null names", `{"kind":"field_list","catalog":{"fields":null}}`, false},
		{"unknown catalog key", `{"kind":"field_list","catalog":{"fields":[],"schema":{}}}`, false},
		{"mixed field target", `{"kind":"field_list","catalog":[],"schema":{}}`, false},
		{"duplicate target kind", `{"kind":"field_list","kind":"field_list","catalog":[]}`, false},
		{"schema missing", `{"kind":"json_schema"}`, false},
		{"schema mixed", `{"kind":"json_schema","schema":{},"catalog":{}}`, false},
		{"schema duplicate nested key", `{"kind":"json_schema","schema":{"type":"object","type":"object"}}`, false},
		{"schema malformed unicode", `{"kind":"json_schema","schema":{"title":"\ud800"}}`, false},
		{"ocsf selection missing", `{"kind":"ocsf","catalog":{}}`, false},
		{"ocsf selection null", `{"kind":"ocsf","catalog":{},"selection":null}`, false},
		{"ocsf selector conflict", `{"kind":"ocsf","catalog":{},"selection":{"version":"1","class":"x","category":"y"}}`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			raw := strings.TrimSuffix(testRequest(""), "}") + `,"validation_target":` + tc.target + `}`
			r, err := DecodeRequest([]byte(raw))
			if !tc.valid {
				if err == nil || !IsInputError(err) {
					t.Fatalf("want input error: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if r.ValidationTarget == nil {
				t.Fatal("target lost")
			}
			wire, err := json.Marshal(r.ValidationTarget)
			if err != nil {
				t.Fatal(err)
			}
			var got map[string]json.RawMessage
			if err := json.Unmarshal(wire, &got); err != nil {
				t.Fatal(err)
			}
			if _, found := got["schema_target"]; found {
				t.Fatalf("schema wire shape changed: %s", wire)
			}
			if r.ValidationTarget.Kind == "field_list" && (r.ValidationTarget.Catalog == nil || r.ValidationTarget.SchemaTarget != nil) {
				t.Fatal("wrong field target union")
			}
			if r.ValidationTarget.Kind != "field_list" && (r.ValidationTarget.Catalog != nil || r.ValidationTarget.SchemaTarget == nil) {
				t.Fatal("wrong schema target union")
			}
		})
	}
}

func TestRewriteModelDirectTarget(t *testing.T) {
	for _, tc := range []struct {
		name   string
		target ValidationTarget
	}{
		{"missing payload", ValidationTarget{Kind: "field_list"}},
		{"mixed union", ValidationTarget{Kind: "field_list", Catalog: &validation.FieldCatalog{}, SchemaTarget: &validation.SchemaTarget{}}},
		{"kind mismatch", ValidationTarget{Kind: "json_schema", SchemaTarget: &validation.SchemaTarget{Kind: "ocsf", Catalog: json.RawMessage(`{}`)}}},
		{"invalid name UTF8", ValidationTarget{Kind: "field_list", Catalog: &validation.FieldCatalog{Fields: []string{string([]byte{0xff})}}}},
		{"invalid schema identity UTF8", ValidationTarget{Kind: "json_schema", SchemaTarget: &validation.SchemaTarget{Kind: "json_schema", Identity: string([]byte{0xff}), Schema: json.RawMessage(`{}`)}}},
		{"empty raw mixed union", ValidationTarget{Kind: "json_schema", SchemaTarget: &validation.SchemaTarget{Kind: "json_schema", Schema: json.RawMessage(`{}`), Catalog: json.RawMessage{}}}},
		{"empty map mixed union", ValidationTarget{Kind: "ocsf", SchemaTarget: &validation.SchemaTarget{Kind: "ocsf", Catalog: json.RawMessage(`{}`), Resources: map[string]json.RawMessage{}, Selection: &validation.OCSFSelection{Version: "1", Class: "event"}}}},
		{"invalid resource key UTF8", ValidationTarget{Kind: "json_schema", SchemaTarget: &validation.SchemaTarget{Kind: "json_schema", Schema: json.RawMessage(`{}`), Resources: map[string]json.RawMessage{string([]byte{0xff}): json.RawMessage(`{}`)}}}},
		{"invalid selection UTF8", ValidationTarget{Kind: "ocsf", SchemaTarget: &validation.SchemaTarget{Kind: "ocsf", Catalog: json.RawMessage(`{}`), Selection: &validation.OCSFSelection{Version: "1", Class: "event", Profiles: []string{string([]byte{0xff})}}}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := prepareRequest(Request{SchemaVersion: 1, ValidationTarget: &tc.target}); err == nil || !IsInputError(err) {
				t.Fatalf("want input error: %v", err)
			}
		})
	}
}

func TestRewriteModelTargetOwnership(t *testing.T) {
	catalog := &validation.FieldCatalog{Fields: []string{"z", "a"}, OptionalFields: []string{"optional"}}
	r, err := prepareRequest(Request{SchemaVersion: 1, ValidationTarget: &ValidationTarget{Kind: "field_list", Catalog: catalog}})
	if err != nil {
		t.Fatal(err)
	}
	catalog.Fields[0] = "changed"
	if r.ValidationTarget.Catalog.Fields[0] != "a" || r.ValidationTarget.Catalog.Fields[1] != "z" {
		t.Fatal("catalog not normalized/owned")
	}
	uid := int64(1001)
	target := &validation.SchemaTarget{Kind: "ocsf", Catalog: json.RawMessage(`{"version":"1"}`), Selection: &validation.OCSFSelection{Version: "1", ClassUID: &uid, Profiles: []string{"p"}}}
	r, err = prepareRequest(Request{SchemaVersion: 1, ValidationTarget: &ValidationTarget{Kind: "ocsf", SchemaTarget: target}})
	if err != nil {
		t.Fatal(err)
	}
	uid = 2002
	target.Catalog[2] = 'X'
	target.Selection.Profiles[0] = "changed"
	got := r.ValidationTarget.SchemaTarget
	if *got.Selection.ClassUID != 1001 || string(got.Catalog) != `{"version":"1"}` || got.Selection.Profiles[0] != "p" {
		t.Fatal("OCSF target aliases caller")
	}
	target = &validation.SchemaTarget{Kind: "json_schema", Schema: json.RawMessage(`{"type":"object"}`), Resources: map[string]json.RawMessage{"local": json.RawMessage(`{"type":"string"}`)}}
	r, err = prepareRequest(Request{SchemaVersion: 1, ValidationTarget: &ValidationTarget{Kind: "json_schema", SchemaTarget: target}})
	if err != nil {
		t.Fatal(err)
	}
	target.Schema[2] = 'X'
	target.Resources["local"][2] = 'X'
	if string(r.ValidationTarget.SchemaTarget.Schema) != `{"type":"object"}` || string(r.ValidationTarget.SchemaTarget.Resources["local"]) != `{"type":"string"}` {
		t.Fatal("schema raw bytes alias caller")
	}
}
