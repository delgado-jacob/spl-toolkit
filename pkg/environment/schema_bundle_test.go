package environment

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
)

func bundleFixture(t *testing.T) SchemaBundle {
	t.Helper()
	ocsf, err := os.ReadFile("../../testdata/schemas/ocsf/edge-cases.json")
	if err != nil {
		t.Fatal(err)
	}
	p := Provenance{SourceKind: "fixture", SourceID: "schema-source", ObservedAt: "2026-10-01T12:00:00Z"}
	identity := ObjectIdentity{Kind: "sourcetype", Name: "events"}
	return SchemaBundle{SchemaVersion: 1, BundleID: "bundle-a", Provenance: p,
		Schemas: []SchemaEntry{
			{ID: "fields", Kind: "field_list", Catalog: json.RawMessage(`{"fields":["host"],"optional_fields":[],"identity":"local","version":"1"}`), Provenance: p},
			{ID: "closed", Kind: "json_schema", Target: json.RawMessage(`{"kind":"json_schema","schema":{"type":"object","properties":{"host":true},"additionalProperties":false}}`), Provenance: p},
			{ID: "open", Kind: "json_schema", Target: json.RawMessage(`{"kind":"json_schema","schema":{"type":"object","properties":{"host":true},"x-large":9223372036854775808}}`), Provenance: p},
			{ID: "ocsf", Kind: "ocsf", Target: fixtureRaw(t, map[string]any{"kind": "ocsf", "catalog": json.RawMessage(ocsf), "selection": map[string]any{"version": "1.6.0", "class": "a"}}), Provenance: p},
		},
		Bindings: []SchemaBinding{
			{SchemaID: "fields", ObjectID: "source-a", Expected: identity, SourceCoverage: "complete"},
			{SchemaID: "closed", ObjectID: "source-a", Expected: identity, SourceCoverage: "complete"},
			{SchemaID: "open", ObjectID: "source-a", Expected: identity, SourceCoverage: "complete"},
			{SchemaID: "ocsf", ObjectID: "source-a", Expected: identity, SourceCoverage: "complete"},
		}}
}

func TestSchemaBundleKindsAndReuse(t *testing.T) {
	b := bundleFixture(t)
	before := fixtureRaw(t, b)
	prepared, report, err := PrepareSchemaBundle(b)
	if err != nil || prepared == nil || report.Status != "valid" || !strings.HasPrefix(report.SchemaBundleDigest, "sha256:") {
		t.Fatalf("prepare: %v %#v", err, report)
	}
	if !reflect.DeepEqual(before, fixtureRaw(t, b)) {
		t.Fatal("caller bundle mutated")
	}
	if len(prepared.targets) != 4 || prepared.targets["fields"].field == nil || prepared.targets["closed"].schema == nil || prepared.targets["open"].schema == nil || prepared.targets["ocsf"].schema == nil {
		t.Fatalf("compiled targets: %#v", prepared.targets)
	}
	open, err := prepared.targets["open"].schema.Validate(analysis.QueryDocument{Text: "eval label=1 | table *"})
	if err != nil || len(open.Outcomes) != 1 || open.Outcomes[0].Outcome != "indeterminate" || open.Coverage.SchemaComplete {
		t.Fatalf("open target: %v %#v", err, open)
	}
	closed, err := prepared.targets["closed"].schema.Validate(analysis.QueryDocument{Text: "table unknown"})
	if err != nil || len(closed.Outcomes) != 1 || closed.Outcomes[0].Outcome != "missing" {
		t.Fatalf("closed target: %v %#v", err, closed)
	}
	ocsf, err := prepared.targets["ocsf"].schema.Validate(analysis.QueryDocument{Text: "table time"})
	if err != nil || len(ocsf.Outcomes) != 1 || ocsf.Outcomes[0].Outcome != "required" {
		t.Fatalf("ocsf target: %v %#v", err, ocsf)
	}
	fields, err := prepared.targets["fields"].field.Validate(analysis.QueryDocument{Text: "table host"})
	if err != nil || len(fields.Outcomes) != 1 {
		t.Fatalf("field target: %v %#v", err, fields)
	}
	first := report.SchemaBundleDigest
	copyOfBundle := prepared.Bundle()
	originalCatalog := append([]byte(nil), copyOfBundle.Schemas[1].Catalog...)
	copyOfBundle.Schemas[1].Catalog[0] = ' '
	if !reflect.DeepEqual(prepared.Bundle().Schemas[1].Catalog, json.RawMessage(originalCatalog)) {
		t.Fatal("bundle accessor leaked payload")
	}
	report.Coverage[0].Coverage = "partial"
	if prepared.Report().Coverage[0].Coverage != "complete" {
		t.Fatal("report accessor leaked coverage")
	}
	large := bundleFixture(t)
	large.Schemas[2].Target = json.RawMessage(strings.Replace(string(large.Schemas[2].Target), "9223372036854775808", "9223372036854775809", 1))
	_, changedLarge, err := PrepareSchemaBundle(large)
	if err != nil || changedLarge.SchemaBundleDigest == first {
		t.Fatalf("large numeric change lost: %v %#v", err, changedLarge)
	}
	b.Schemas[0], b.Schemas[3] = b.Schemas[3], b.Schemas[0]
	b.Bindings[0], b.Bindings[3] = b.Bindings[3], b.Bindings[0]
	_, reordered, err := PrepareSchemaBundle(b)
	if err != nil || reordered.SchemaBundleDigest != first {
		t.Fatalf("order changed digest: %v %#v", err, reordered)
	}
	b.Digest = first
	_, asserted, err := PrepareSchemaBundle(b)
	if err != nil || asserted.Status != "valid" {
		t.Fatalf("matching assertion: %v %#v", err, asserted)
	}
	b.Schemas[0].ID = "changed"
	_, changed, err := PrepareSchemaBundle(b)
	if err != nil || changed.Status != "invalid" {
		t.Fatalf("stale digest accepted: %v %#v", err, changed)
	}
}

func TestSchemaBundleRejectsConflicts(t *testing.T) {
	cases := map[string]func(*SchemaBundle){
		"duplicate schema IDs":          func(b *SchemaBundle) { b.Schemas = append(b.Schemas, b.Schemas[0]) },
		"duplicate schema-object pairs": func(b *SchemaBundle) { b.Bindings = append(b.Bindings, b.Bindings[0]) },
		"wrong union payload":           func(b *SchemaBundle) { b.Schemas[0].Target = json.RawMessage(`{"kind":"json_schema","schema":true}`) },
		"kind mismatch": func(b *SchemaBundle) {
			b.Schemas[1].Target = json.RawMessage(`{"kind":"ocsf","catalog":{},"selection":{"version":"1","class":"a"}}`)
		},
		"unknown member": func(b *SchemaBundle) { b.Schemas[0].Catalog = json.RawMessage(`{"fields":["x"],"unknown":true}`) },
		"invalid OCSF selection": func(b *SchemaBundle) {
			b.Schemas[3].Target = json.RawMessage(`{"kind":"ocsf","catalog":{},"selection":{"version":"1.6.0","class":"missing"}}`)
		},
		"nonlocal JSON Schema resource": func(b *SchemaBundle) {
			b.Schemas[1].Target = json.RawMessage(`{"kind":"json_schema","schema":true,"resources":{"relative":true}}`)
		},
		"malformed nested JSON": func(b *SchemaBundle) {
			b.Schemas[1].Target = json.RawMessage(`{"kind":"json_schema","schema":{"properties":{"x":1,"x":2}}}`)
		},
		"partial without reason":   func(b *SchemaBundle) { b.Bindings[0].SourceCoverage = "partial" },
		"asserted digest mismatch": func(b *SchemaBundle) { b.Digest = "sha256:" + strings.Repeat("0", 64) },
	}
	for name, change := range cases {
		t.Run(name, func(t *testing.T) {
			b := bundleFixture(t)
			change(&b)
			prepared, report, err := PrepareSchemaBundle(b)
			if err != nil || prepared != nil || report == nil || report.Status != "invalid" || len(report.Diagnostics) == 0 || report.Coverage == nil || report.Diagnostics == nil {
				t.Fatalf("accepted invalid bundle: %v %#v", err, report)
			}
		})
	}
}

func TestSchemaBundleArtifactValidation(t *testing.T) {
	bundle := fixtureRaw(t, bundleFixture(t))
	snapshot := fixtureRaw(t, snapshotFixture())
	standalone, err := ValidateArtifacts(nil, bundle)
	if err != nil || standalone.Status != "valid" || standalone.SnapshotDigest != "" || standalone.SchemaBundleDigest == "" {
		t.Fatalf("standalone: %v %#v", err, standalone)
	}
	combined, err := ValidateArtifacts(snapshot, bundle)
	if err != nil || combined.Status != "valid" || combined.SnapshotDigest == "" || combined.SchemaBundleDigest != standalone.SchemaBundleDigest {
		t.Fatalf("combined: %v %#v", err, combined)
	}
	inline, err := ValidateJSON(fixtureRaw(t, map[string]any{"schema_version": 1, "schema_bundle": json.RawMessage(bundle)}))
	if err != nil || inline.SchemaBundleDigest != standalone.SchemaBundleDigest {
		t.Fatalf("inline: %v %#v", err, inline)
	}
	badSnapshot := []byte(`{"schema_version":1}`)
	badBundle := []byte(`{"schema_version":1,"bundle_id":"bad","schemas":[]}`)
	invalid, err := ValidateArtifacts(badSnapshot, badBundle)
	if err != nil || invalid.Status != "invalid" || len(invalid.Diagnostics) != 2 || invalid.Coverage == nil {
		t.Fatalf("aggregate invalid: %v %#v", err, invalid)
	}
	for name, raw := range map[string][]byte{
		"unknown property": []byte(strings.Replace(string(bundle), `"bundle_id":"bundle-a"`, `"bundle_id":"bundle-a","extra":true`, 1)),
		"missing bindings": []byte(strings.Replace(string(bundle), `,"bindings":[`, `,"lost_bindings":[`, 1)),
	} {
		t.Run(name, func(t *testing.T) {
			report, err := ValidateArtifacts(nil, raw)
			if err != nil || report.Status != "invalid" || len(report.Diagnostics) != 1 {
				t.Fatalf("raw bundle accepted: %v %#v", err, report)
			}
		})
	}
}
