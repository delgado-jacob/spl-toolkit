package validation

import (
	"encoding/json"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
)

func typedField(kind string, segments ...string) analysis.FieldIdentity {
	return analysis.FieldIdentity{Kind: kind, Segments: segments}
}

func TestProjectionFieldCatalog(t *testing.T) {
	p, err := PrepareFieldCatalog(FieldCatalog{Fields: []string{"actor.name", "host"}, OptionalFields: []string{"optional"}})
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct {
		field     analysis.FieldIdentity
		outcome   string
		admission analysis.SourceFieldAdmission
	}{
		{typedField("atomic", "actor.name"), "required", analysis.SourceFieldAdmitted},
		{typedField("path", "host"), "required", analysis.SourceFieldAdmitted},
		{typedField("atomic", "optional"), "optional", analysis.SourceFieldAdmitted},
		{typedField("atomic", "missing"), "missing", analysis.SourceFieldProhibited},
		{typedField("path", "actor", "name"), "indeterminate", analysis.SourceFieldIndeterminate},
	} {
		got, err := p.ProjectField(c.field)
		if got.Evidence == nil || got.SupportingClasses == nil || got.MissingClasses == nil || got.IndeterminateClasses == nil {
			t.Fatal("projection contains null arrays")
		}
		if err != nil || got.Outcome != c.outcome || got.Admission != c.admission {
			t.Fatalf("%+v: %+v %v", c.field, got, err)
		}
		if c.outcome == "indeterminate" && !hasOCSFReason(got.Evidence, "typed_path_not_represented") {
			t.Fatal("missing typed path evidence")
		}
	}
}

func TestProjectionMalformedCalls(t *testing.T) {
	catalog, _ := PrepareFieldCatalog(FieldCatalog{})
	schema, _ := PrepareSchemaTarget(SchemaTarget{Kind: "json_schema", Schema: json.RawMessage(`true`)})
	calls := []func(analysis.FieldIdentity) (FieldProjection, error){catalog.ProjectField, schema.ProjectField}
	for _, field := range []analysis.FieldIdentity{
		{}, typedField("dynamic", "x"), typedField("segmented", "x"), typedField("atomic"), typedField("atomic", "a", "b"),
		typedField("path", ""), typedField("path", " "), typedField("atomic", string([]byte{0xff})),
		{Kind: "atomic", Segments: []string{"x"}, Qualifier: "left"},
		typedField("path", strings.Split(strings.Repeat("x.", schemaPathSegmentBudget)+"x", ".")...),
	} {
		for _, call := range calls {
			if _, err := call(field); !IsInputError(err) {
				t.Fatalf("%+v: expected InputError, got %v", field, err)
			}
		}
	}
	for _, p := range []*PreparedFieldCatalog{nil, {}} {
		if _, err := p.ProjectField(typedField("atomic", "x")); !IsInputError(err) {
			t.Fatal(err)
		}
	}
	for _, p := range []*PreparedSchemaTarget{nil, {}} {
		if _, err := p.ProjectField(typedField("atomic", "x")); !IsInputError(err) {
			t.Fatal(err)
		}
	}
}

func TestProjectionJSONTypedPaths(t *testing.T) {
	p, err := PrepareSchemaTarget(SchemaTarget{Kind: "json_schema", Schema: json.RawMessage(`{"type":"object","properties":{"actor.name":true,"actor":{"type":"object","properties":{"name":true},"required":["name"],"additionalProperties":false}},"required":["actor"],"additionalProperties":false}`)})
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct {
		field   analysis.FieldIdentity
		outcome string
	}{
		{typedField("atomic", "actor.name"), "optional"}, {typedField("path", "actor", "name"), "required"},
		{typedField("path", "actor.name"), "optional"}, {typedField("atomic", "actor.missing"), "missing"},
		{typedField("path", "actor", "missing"), "missing"},
	} {
		got, err := p.ProjectField(c.field)
		if err != nil || got.Outcome != c.outcome {
			t.Fatalf("%+v: %+v %v", c.field, got, err)
		}
	}
	for _, raw := range []string{`true`, `{"type":"object","additionalProperties":true}`, `{"type":"object","properties":{"x":true},"if":{},"then":{}}`, `{"$ref":"urn:missing"}`, `{"type":"object","additionalProperties":false,"$ref":"urn:missing"}`, `{"type":"object","additionalProperties":false,"if":{},"then":{}}`} {
		p, err := PrepareSchemaTarget(SchemaTarget{Kind: "json_schema", Schema: json.RawMessage(raw)})
		if err != nil {
			t.Fatal(err)
		}
		got, err := p.ProjectField(typedField("atomic", "x"))
		if err != nil || got.Admission != analysis.SourceFieldIndeterminate {
			t.Fatalf("open/unknown %s: %+v %v", raw, got, err)
		}
	}
}

func TestProjectionOCSFTypedPathsAndOwnership(t *testing.T) {
	var catalog map[string]any
	if err := json.Unmarshal(edgeOCSF(t), &catalog); err != nil {
		t.Fatal(err)
	}
	attrs := catalog["classes"].(map[string]any)["a"].(map[string]any)["attributes"].(map[string]any)
	attrs["single.x"] = map[string]any{"type": "string_t", "requirement": "optional"}
	raw, _ := json.Marshal(catalog)
	p, err := PrepareSchemaTarget(SchemaTarget{Kind: "ocsf", Catalog: raw, Selection: &OCSFSelection{Version: "1.6.0", Class: "a"}})
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct {
		field   analysis.FieldIdentity
		outcome string
	}{
		{typedField("atomic", "single.x"), "optional"}, {typedField("path", "single", "x"), "required"},
		{typedField("atomic", "cycle.value"), "missing"}, {typedField("path", "cycle", "value"), "optional"},
		{typedField("path", "generic", "x"), "permitted_unspecified"}, {typedField("path", "array", "x"), "indeterminate"},
	} {
		got, err := p.ProjectField(c.field)
		if err != nil || got.Outcome != c.outcome {
			t.Fatalf("%+v: %+v %v", c.field, got, err)
		}
	}
	want, err := p.ProjectField(typedField("path", "single", "x"))
	if err != nil {
		t.Fatal(err)
	}
	got, _ := p.ProjectField(typedField("path", "single", "x"))
	got.Evidence[0].Reason = "changed"
	*got.Evidence[0].ClassUID = -1
	got.SupportingClasses[0].Key = "changed"
	again, _ := p.ProjectField(typedField("path", "single", "x"))
	if !reflect.DeepEqual(want, again) {
		t.Fatal("projection aliases prepared target")
	}
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				got, err := p.ProjectField(typedField("path", "single", "x"))
				if err != nil || !reflect.DeepEqual(want, got) {
					t.Errorf("concurrent projection: %+v %v", got, err)
				}
			}
		}()
	}
	wg.Wait()
}

func TestProjectionOCSFConstraintCollisionAndLegacy(t *testing.T) {
	var catalog map[string]any
	if err := json.Unmarshal(edgeOCSF(t), &catalog); err != nil {
		t.Fatal(err)
	}
	a := catalog["classes"].(map[string]any)["a"].(map[string]any)
	a["constraints"] = map[string]any{"at_least_one": []string{"cycle.value"}}
	raw, _ := json.Marshal(catalog)
	p, err := PrepareSchemaTarget(SchemaTarget{Kind: "ocsf", Catalog: raw, Selection: &OCSFSelection{Version: "1.6.0", Class: "a"}})
	if err != nil {
		t.Fatal(err)
	}
	got, err := p.ProjectField(typedField("path", "cycle", "value"))
	if err != nil || got.Outcome != "required" {
		t.Fatalf("unambiguous constraint: %+v %v", got, err)
	}
	a["attributes"].(map[string]any)["cycle.value"] = map[string]any{"type": "string_t", "requirement": "optional"}
	raw, _ = json.Marshal(catalog)
	p, err = PrepareSchemaTarget(SchemaTarget{Kind: "ocsf", Catalog: raw, Selection: &OCSFSelection{Version: "1.6.0", Class: "a"}})
	if err != nil {
		t.Fatal(err)
	}
	for _, field := range []analysis.FieldIdentity{typedField("atomic", "cycle.value"), typedField("path", "cycle", "value")} {
		got, err := p.ProjectField(field)
		if err != nil || got.Outcome != "indeterminate" || !hasOCSFReason(got.Evidence, "ocsf_constraint") {
			t.Fatalf("ambiguous constraint %+v: %+v %v", field, got, err)
		}
	}
	legacy, err := p.Validate(analysis.QueryDocument{Text: "table cycle.value"})
	if err != nil || len(legacy.Outcomes) != 1 || legacy.Outcomes[0].Outcome != "indeterminate" || !hasOCSFReason(legacy.Outcomes[0].Evidence, "literal_path_collision") {
		t.Fatalf("legacy string collision changed: %+v %v", legacy, err)
	}
}

func TestProjectionJSONConcurrencyAndBudget(t *testing.T) {
	raw := json.RawMessage(`{"type":"object","properties":{"x":true},"required":["x"],"additionalProperties":false}`)
	p, err := PrepareSchemaTarget(SchemaTarget{Kind: "json_schema", Schema: raw})
	if err != nil {
		t.Fatal(err)
	}
	field := typedField("atomic", "x")
	want, _ := p.ProjectField(field)
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				got, err := p.ProjectField(field)
				if err != nil || !reflect.DeepEqual(want, got) {
					t.Errorf("projection: %+v %v", got, err)
				}
				got.Evidence[0].Reason = "changed"
			}
		}()
	}
	wg.Wait()
	branches := make([]json.RawMessage, schemaProjectionBudget+1)
	for i := range branches {
		branches[i] = raw
	}
	expanded, _ := json.Marshal(map[string]any{"allOf": branches})
	p, err = PrepareSchemaTarget(SchemaTarget{Kind: "json_schema", Schema: expanded})
	if err != nil {
		t.Fatal(err)
	}
	got, err := p.ProjectField(field)
	if err != nil || got.Admission != analysis.SourceFieldIndeterminate || !hasOCSFReason(got.Evidence, "traversal_budget") {
		t.Fatalf("bounded projection: %+v %v", got, err)
	}
}

func TestProjectionOCSFUnrelatedAtomicConstraint(t *testing.T) {
	var catalog map[string]any
	if err := json.Unmarshal(edgeOCSF(t), &catalog); err != nil {
		t.Fatal(err)
	}
	a := catalog["classes"].(map[string]any)["a"].(map[string]any)
	a["constraints"] = map[string]any{"just_one": []string{"cycle"}}
	a["attributes"].(map[string]any)["cycle.value"] = map[string]any{"type": "string_t", "requirement": "optional"}
	raw, _ := json.Marshal(catalog)
	p, err := PrepareSchemaTarget(SchemaTarget{Kind: "ocsf", Catalog: raw, Selection: &OCSFSelection{Version: "1.6.0", Class: "a"}})
	if err != nil {
		t.Fatal(err)
	}
	got, err := p.ProjectField(typedField("atomic", "cycle.value"))
	if err != nil || got.Outcome != "optional" {
		t.Fatalf("structural ancestor constraint conflated literal declaration: %+v %v", got, err)
	}
}

func TestProjectionConditionalAndBoundedOCSF(t *testing.T) {
	p, err := PrepareSchemaTarget(SchemaTarget{Kind: "ocsf", Catalog: edgeOCSF(t), Selection: &OCSFSelection{Version: "1.6.0", Category: "test", Profiles: []string{"p"}}})
	if err != nil {
		t.Fatal(err)
	}
	got, err := p.ProjectField(typedField("atomic", "p_only"))
	if err != nil || got.Outcome != "conditional" || got.Admission != analysis.SourceFieldIndeterminate || len(got.SupportingClasses) == 0 || len(got.MissingClasses) == 0 {
		t.Fatalf("conditional category: %+v %v", got, err)
	}
	var catalog map[string]any
	if err := json.Unmarshal(edgeOCSF(t), &catalog); err != nil {
		t.Fatal(err)
	}
	b := catalog["classes"].(map[string]any)["b"].(map[string]any)
	names := make([]string, schemaProjectionBudget+1)
	for i := range names {
		names[i] = "unrelated"
	}
	b["constraints"] = map[string]any{"future": names}
	raw, _ := json.Marshal(catalog)
	p, err = PrepareSchemaTarget(SchemaTarget{Kind: "ocsf", Catalog: raw, Selection: &OCSFSelection{Version: "1.6.0", Class: "b"}})
	if err != nil {
		t.Fatal(err)
	}
	got, err = p.ProjectField(typedField("atomic", "time"))
	if err != nil || got.Admission != analysis.SourceFieldIndeterminate || !hasOCSFReason(got.Evidence, "traversal_budget") {
		t.Fatalf("constraint budget: %+v %v", got, err)
	}
}

func TestProjectionJSONWholePathDeclaration(t *testing.T) {
	for _, c := range []struct {
		name, schema, outcome string
		admission             analysis.SourceFieldAdmission
	}{
		{"true ancestor", `{"type":"object","properties":{"actor":true},"additionalProperties":false}`, "indeterminate", analysis.SourceFieldIndeterminate},
		{"open required ancestor", `{"type":"object","properties":{"actor":{"type":"object","additionalProperties":true}},"required":["actor"],"additionalProperties":false}`, "indeterminate", analysis.SourceFieldIndeterminate},
		{"schema additional property", `{"type":"object","properties":{"actor":{"type":"object","additionalProperties":{"type":"string"}}},"additionalProperties":false}`, "indeterminate", analysis.SourceFieldIndeterminate},
		{"explicit true leaf", `{"type":"object","properties":{"actor":{"type":"object","properties":{"name":true},"additionalProperties":false}},"additionalProperties":false}`, "optional", analysis.SourceFieldAdmitted},
		{"generic conjunction with declaration", `{"allOf":[{"type":"object","properties":{"actor":true}},{"type":"object","properties":{"actor":{"type":"object","properties":{"name":true}}}}]}`, "optional", analysis.SourceFieldAdmitted},
		{"conjunctive prefix and descendant", `{"allOf":[{"type":"object","properties":{"actor":true}},{"type":"object","additionalProperties":{"type":"object","properties":{"name":true}}}]}`, "optional", analysis.SourceFieldAdmitted},
		{"pattern declaration", `{"type":"object","patternProperties":{"^actor$":{"type":"object","patternProperties":{"^name$":true},"additionalProperties":false}},"additionalProperties":false}`, "optional", analysis.SourceFieldAdmitted},
		{"resolved declaration", `{"$defs":{"actor":{"type":"object","properties":{"name":true},"additionalProperties":false}},"type":"object","properties":{"actor":{"$ref":"#/$defs/actor"}},"additionalProperties":false}`, "optional", analysis.SourceFieldAdmitted},
		{"unresolved ancestor", `{"type":"object","properties":{"actor":{"$ref":"urn:missing"}},"additionalProperties":false}`, "indeterminate", analysis.SourceFieldIndeterminate},
		{"alternative generic ancestor", `{"anyOf":[{"type":"object","properties":{"actor":true}},{"type":"object","properties":{"actor":{"type":"object","properties":{"name":true}}}}]}`, "indeterminate", analysis.SourceFieldIndeterminate},
		{"common alternative declarations", `{"anyOf":[{"type":"object","properties":{"actor":{"type":"object","properties":{"name":true}}}},{"type":"object","properties":{"actor":{"type":"object","properties":{"name":{"type":"string"}}}}}]}`, "optional", analysis.SourceFieldAdmitted},
		{"recursive ancestor", `{"$defs":{"actor":{"$ref":"#/$defs/actor"}},"type":"object","properties":{"actor":{"$ref":"#/$defs/actor"}}}`, "indeterminate", analysis.SourceFieldIndeterminate},
		{"unsupported leaf pattern", `{"type":"object","properties":{"actor":{"type":"object","patternProperties":{"(?=name)name":true},"additionalProperties":false}}}`, "indeterminate", analysis.SourceFieldIndeterminate},
		{"required-only leaf", `{"type":"object","properties":{"actor":{"type":"object","required":["name"]}},"required":["actor"]}`, "required", analysis.SourceFieldAdmitted},
		{"undeclared ancestor", `{"type":"object","additionalProperties":{"type":"object","properties":{"name":true}}}`, "indeterminate", analysis.SourceFieldIndeterminate},
	} {
		t.Run(c.name, func(t *testing.T) {
			p, err := PrepareSchemaTarget(SchemaTarget{Kind: "json_schema", Schema: json.RawMessage(c.schema)})
			if err != nil {
				t.Fatal(err)
			}
			got, err := p.ProjectField(typedField("path", "actor", "name"))
			if err != nil || got.Outcome != c.outcome || got.Admission != c.admission {
				t.Fatalf("projection: %+v %v", got, err)
			}
		})
	}
}

func TestProjectionJSONDeepDeclaration(t *testing.T) {
	for _, declared := range []bool{false, true} {
		child := any(true)
		if declared {
			child = map[string]any{"type": "object", "properties": map[string]any{"name": true}, "additionalProperties": false}
		}
		raw, _ := json.Marshal(map[string]any{"type": "object", "properties": map[string]any{"actor": map[string]any{"type": "object", "properties": map[string]any{"user": child}, "additionalProperties": false}}, "additionalProperties": false})
		p, err := PrepareSchemaTarget(SchemaTarget{Kind: "json_schema", Schema: raw})
		if err != nil {
			t.Fatal(err)
		}
		got, err := p.ProjectField(typedField("path", "actor", "user", "name"))
		want := analysis.SourceFieldIndeterminate
		if declared {
			want = analysis.SourceFieldAdmitted
		}
		if err != nil || got.Admission != want {
			t.Fatalf("declared %v: %+v %v", declared, got, err)
		}
	}
}
