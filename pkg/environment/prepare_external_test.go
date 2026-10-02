package environment_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/environment"
)

func TestPairedTargetsReuseAcrossDocuments(t *testing.T) {
	read := func(path string, out any) {
		t.Helper()
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(raw, out); err != nil {
			t.Fatal(err)
		}
	}
	var snapshot environment.Snapshot
	read("../../examples/environment/partial-snapshot.json", &snapshot)
	var bundle environment.SchemaBundle
	read("../../examples/environment/fields.json", &bundle)
	var ocsf json.RawMessage
	read("../../testdata/schemas/ocsf/edge-cases.json", &ocsf)
	provenance := bundle.Provenance
	identity := environment.ObjectIdentity{Kind: "sourcetype", Name: "events"}
	bundle.Schemas = append(bundle.Schemas,
		environment.SchemaEntry{ID: "closed", Kind: "json_schema", Target: json.RawMessage(`{"kind":"json_schema","schema":{"type":"object","properties":{"host":true},"additionalProperties":false}}`), Provenance: provenance},
		environment.SchemaEntry{ID: "ocsf", Kind: "ocsf", Target: json.RawMessage(`{"kind":"ocsf","catalog":` + string(ocsf) + `,"selection":{"version":"1.6.0","class":"a"}}`), Provenance: provenance},
	)
	bundle.Bindings = append(bundle.Bindings,
		environment.SchemaBinding{SchemaID: "closed", ObjectID: "source-a", Expected: identity, SourceCoverage: "complete"},
		environment.SchemaBinding{SchemaID: "ocsf", ObjectID: "source-a", Expected: identity, SourceCoverage: "complete"},
		environment.SchemaBinding{SchemaID: "closed", ObjectID: "missing", Expected: environment.ObjectIdentity{Kind: "macro", Name: "later"}, SourceCoverage: "complete"},
	)
	preparedSnapshot, snapshotReport, err := environment.PrepareSnapshot(snapshot)
	if err != nil || preparedSnapshot == nil || snapshotReport.Status != "partial" {
		t.Fatalf("prepare snapshot: %v %#v", err, snapshotReport)
	}
	preparedBundle, bundleReport, err := environment.PrepareSchemaBundle(bundle)
	if err != nil || preparedBundle == nil || bundleReport.Status != "partial" {
		t.Fatalf("prepare bundle: %v %#v", err, bundleReport)
	}
	paired, report, err := environment.Pair(preparedSnapshot, preparedBundle)
	if err != nil || paired == nil || report.Status != "partial" {
		t.Fatalf("pair: %v %#v", err, report)
	}
	bindings := paired.Bindings("source-a")
	partialFieldSource := false
	for _, binding := range bindings {
		if binding.SchemaID == "events-fields" {
			partialFieldSource = binding.SourceCoverage == "partial" && binding.Reason != ""
		}
	}
	if len(bindings) != 3 || !partialFieldSource {
		t.Fatalf("resolved bindings lost independent coverage: %#v", bindings)
	}
	fields, ok := paired.FieldCatalog("source-a", "events-fields")
	if !ok || fields == nil {
		t.Fatal("resolved field catalog unavailable")
	}
	closed, ok := paired.SchemaTarget("source-a", "closed")
	if !ok || closed == nil {
		t.Fatal("resolved JSON Schema target unavailable")
	}
	ocsfTarget, ok := paired.SchemaTarget("source-a", "ocsf")
	if !ok || ocsfTarget == nil {
		t.Fatal("resolved OCSF target unavailable")
	}
	if again, ok := paired.FieldCatalog("source-a", "events-fields"); !ok || again != fields {
		t.Fatal("field catalog was not reused")
	}
	if again, ok := paired.SchemaTarget("source-a", "closed"); !ok || again != closed {
		t.Fatal("schema target was not reused")
	}
	for _, check := range []struct {
		document      analysis.QueryDocument
		fieldOutcome  string
		schemaOutcome string
	}{
		{document: analysis.QueryDocument{Text: "table host"}, fieldOutcome: "matching", schemaOutcome: "optional"},
		{document: analysis.QueryDocument{Text: "table unknown"}, fieldOutcome: "missing", schemaOutcome: "missing"},
	} {
		if result, err := fields.Validate(check.document); err != nil || len(result.Outcomes) != 1 || result.Outcomes[0].Outcome != check.fieldOutcome {
			t.Fatalf("field validation: %v %#v", err, result)
		}
		if result, err := closed.Validate(check.document); err != nil || len(result.Outcomes) != 1 || result.Outcomes[0].Outcome != check.schemaOutcome {
			t.Fatalf("JSON Schema validation: %v %#v", err, result)
		}
		if result, err := ocsfTarget.Validate(check.document); err != nil || len(result.Outcomes) != 1 {
			t.Fatalf("OCSF validation: %v %#v", err, result)
		}
	}
	if _, ok := paired.FieldCatalog("source-a", "closed"); ok {
		t.Fatal("schema target returned as field catalog")
	}
	if _, ok := paired.SchemaTarget("source-a", "events-fields"); ok {
		t.Fatal("field catalog returned as schema target")
	}
	if _, ok := paired.SchemaTarget("missing", "closed"); ok {
		t.Fatal("unresolved binding returned a target")
	}
	if _, ok := paired.SchemaTarget("other", "closed"); ok {
		t.Fatal("unbound object returned a target")
	}
	mismatched := snapshot
	mismatched.Objects = append([]environment.Object(nil), snapshot.Objects...)
	mismatched.Objects[0].Name = "different"
	preparedMismatch, _, err := environment.PrepareSnapshot(mismatched)
	if err != nil || preparedMismatch == nil {
		t.Fatalf("prepare mismatched snapshot: %v", err)
	}
	if environmentWithMismatch, mismatchReport, err := environment.Pair(preparedMismatch, preparedBundle); err != nil || environmentWithMismatch != nil || mismatchReport.Status != "invalid" {
		t.Fatalf("mismatched binding should not produce a prepared environment: %v %#v", err, mismatchReport)
	}

	observed := snapshot
	observed.SchemaVersion = 2
	observed.Collections = append([]environment.Collection{{Kind: "index", Coverage: "complete"}}, snapshot.Collections...)
	observed.Objects = append([]environment.Object{{ID: "index-main", Kind: "index", Name: "main", Provenance: snapshot.Objects[0].Provenance}}, snapshot.Objects...)
	yes := true
	observed.Observation = &environment.ObservationScope{
		IndexSelection: environment.Selector{All: &yes},
		Enumeration:    environment.IndexEnumeration{Method: "distributed_rest", PeerScope: "configured_search_peers", Coverage: "complete", Provenance: snapshot.Objects[0].Provenance},
		Indexes:        []environment.ObservationIndex{{IndexID: "index-main", CatalogDatatypes: []string{"event"}, RequiredDatatypes: []string{"event"}}}, UnmatchedIndexes: []string{},
		Method: "splunk_metadata", Visibility: "exporting_principal", Window: environment.ObservationWindow{Mode: "all_retained"}, TimePrecision: "bucket_overlap", AbsenceMeaning: "not_observed",
		Captures: []environment.ObservationCapture{{Kind: "sourcetype", IndexID: "index-main", Datatype: "event", ObjectIDs: []string{"source-a"}, Coverage: "complete", Provenance: snapshot.Objects[0].Provenance}},
	}
	for _, captured := range []bool{true, false} {
		if !captured {
			observed.Objects = observed.Objects[:1]
			observed.Observation.Captures[0].ObjectIDs = []string{}
		}
		preparedObserved, observedReport, err := environment.PrepareSnapshot(observed)
		if err != nil || preparedObserved == nil || observedReport.Status != "partial" {
			t.Fatalf("prepare v2 observation: %v %#v", err, observedReport)
		}
		observedEnvironment, observedReport, err := environment.Pair(preparedObserved, preparedBundle)
		if err != nil || observedEnvironment == nil || observedReport.Status != "partial" {
			t.Fatalf("pair v2 captured=%v: %v %#v", captured, err, observedReport)
		}
		observedFields, fieldsOK := observedEnvironment.FieldCatalog("source-a", "events-fields")
		observedClosed, schemaOK := observedEnvironment.SchemaTarget("source-a", "closed")
		if captured {
			if !fieldsOK || observedFields != fields || !schemaOK || observedClosed != closed {
				t.Fatal("same v1 bundle lost reusable targets for a v2 captured identity")
			}
		} else if fieldsOK || observedFields != nil || schemaOK || observedClosed != nil || len(observedEnvironment.Bindings("source-a")) != 0 {
			t.Fatal("schemas proved presence of an unobserved sourcetype")
		}
	}
}
