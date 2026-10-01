package environment

import (
	"encoding/json"
	"reflect"
	"strings"
	"sync"
	"testing"
)

func preparedPairSnapshot(t *testing.T, name, definition string) *PreparedSnapshot {
	t.Helper()
	fixture := snapshotFixture()
	fixture["objects"] = []any{macroFixture("macro-a", name, definition)}
	var value Snapshot
	if err := json.Unmarshal(fixtureRaw(t, fixture), &value); err != nil {
		t.Fatal(err)
	}
	prepared, report, err := PrepareSnapshot(value)
	if err != nil || prepared == nil || report.Status != "valid" {
		t.Fatalf("prepare snapshot: %v %#v", err, report)
	}
	return prepared
}

func preparedPairBundle(t *testing.T) *PreparedSchemaBundle {
	t.Helper()
	b := bundleFixture(t)
	b.Schemas = b.Schemas[:2]
	b.Bindings = []SchemaBinding{
		{SchemaID: "fields", ObjectID: "macro-a", Expected: ObjectIdentity{Kind: "macro", Name: "first", Namespace: "search", App: "main", Owner: "nobody"}, SourceCoverage: "complete"},
		{SchemaID: "closed", ObjectID: "macro-a", Expected: ObjectIdentity{Kind: "macro", Name: "first", Namespace: "search", App: "main", Owner: "nobody"}, SourceCoverage: "complete"},
	}
	prepared, report, err := PrepareSchemaBundle(b)
	if err != nil || prepared == nil || report.Status != "valid" {
		t.Fatalf("prepare bundle: %v %#v", err, report)
	}
	return prepared
}

func TestPairReusesSchemaBundleAcrossSnapshots(t *testing.T) {
	bundle := preparedPairBundle(t)
	first, second := preparedPairSnapshot(t, "first", "x=1"), preparedPairSnapshot(t, "first", "x=2")
	for _, snapshot := range []*PreparedSnapshot{first, second} {
		env, report, err := Pair(snapshot, bundle)
		if err != nil || env == nil || report.Status != "valid" || len(env.Bindings("macro-a")) != 2 {
			t.Fatalf("valid pair: %v %#v %#v", err, env, report)
		}
		if report.SnapshotDigest != snapshot.Report().SnapshotDigest || report.SchemaBundleDigest != bundle.Report().SchemaBundleDigest {
			t.Fatalf("lost digests: %#v", report)
		}
	}
	third := preparedPairSnapshot(t, "another", "x=3")
	env, report, err := Pair(third, bundle)
	if err != nil || env != nil || report.Status != "invalid" || !hasDiagnostic(report, "binding_identity_mismatch") {
		t.Fatalf("reused ID with different identity: %v %#v", err, report)
	}
}

func hasDiagnostic(report *Report, code string) bool {
	for _, diagnostic := range report.Diagnostics {
		if diagnostic.Code == code {
			return true
		}
	}
	return false
}

func TestPairCoverageAndIsolation(t *testing.T) {
	bundle := preparedPairBundle(t)
	complete := preparedPairSnapshot(t, "first", "x=1")
	env, report, err := Pair(complete, bundle)
	if err != nil || env == nil || report.Status != "valid" || len(env.Bindings("macro-a")) != 2 {
		t.Fatalf("two independent bindings: %v %#v", err, report)
	}
	object, ok := env.Object("macro-a")
	if !ok || object.Name != "first" {
		t.Fatalf("object index: %#v %v", object, ok)
	}
	object.Name = "mutated"
	object.Document.Text = "mutated"
	bindings := env.Bindings("macro-a")
	bindings[0].Expected.Name = "mutated"
	if again, ok := env.Object("macro-a"); !ok || again.Name != "first" || again.Document.Text != "x=1" || env.Bindings("macro-a")[0].Expected.Name != "first" {
		t.Fatal("accessor leaked mutable internals")
	}
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				_, _ = env.Object("macro-a")
				_ = env.Bindings("macro-a")
				_ = env.Report()
			}
		}()
	}
	wg.Wait()

	var value Snapshot
	if err := json.Unmarshal(fixtureRaw(t, snapshotFixture()), &value); err != nil {
		t.Fatal(err)
	}
	missing, _, _ := PrepareSnapshot(value)
	env, report, err = Pair(missing, bundle)
	if err != nil || env != nil || report.Status != "invalid" || !hasDiagnostic(report, "binding_object_absent") {
		t.Fatalf("complete absence: %v %#v", err, report)
	}
	for _, coverage := range []string{"partial", "unavailable"} {
		value.Collections = []Collection{{Kind: "macro", Coverage: coverage, Reason: "limited"}}
		missing, _, _ = PrepareSnapshot(value)
		env, report, err = Pair(missing, bundle)
		if err != nil || env == nil || report.Status != "partial" || !hasDiagnostic(report, "binding_unresolved") || len(env.Bindings("macro-a")) != 0 {
			t.Fatalf("%s absence: %v %#v", coverage, err, report)
		}
	}
	value.Collections = []Collection{}
	missing, _, _ = PrepareSnapshot(value)
	env, report, err = Pair(missing, bundle)
	if err != nil || env == nil || report.Status != "partial" || !hasDiagnostic(report, "binding_unresolved") {
		t.Fatalf("omitted collection: %v %#v", err, report)
	}
	value.Collections = []Collection{{Kind: "macro", Coverage: "complete"}}
	value.CaptureScope.Namespace = Selector{Values: []string{"other"}}
	missing, _, _ = PrepareSnapshot(value)
	env, report, err = Pair(missing, bundle)
	if err != nil || env == nil || report.Status != "partial" || !hasDiagnostic(report, "binding_unresolved") {
		t.Fatalf("outside scope: %v %#v", err, report)
	}

	env, report, err = Pair(complete, nil)
	if err != nil || env == nil || report.Status != "valid" || len(env.Bindings("macro-a")) != 0 {
		t.Fatalf("object-only pair: %v %#v", err, report)
	}
	env, report, err = Pair(nil, bundle)
	if err != nil || env != nil || report.Status != "invalid" || report.Coverage == nil || report.Diagnostics == nil {
		t.Fatalf("nil snapshot: %v %#v", err, report)
	}

	dup := bundleFixture(t)
	dup.Bindings = append(dup.Bindings, dup.Bindings[0])
	_, duplicateReport, _ := PrepareSchemaBundle(dup)
	if duplicateReport.Status != "invalid" {
		t.Fatalf("duplicate pair accepted: %#v", duplicateReport)
	}
	if !strings.Contains(FormatReport(report), "invalid") {
		t.Fatal("report format missing")
	}
}

func TestPairRetainsValidBindingsAndValidationSurfaces(t *testing.T) {
	snapshot := preparedPairSnapshot(t, "first", "x=1")
	bundle := bundleFixture(t)
	bundle.Schemas = bundle.Schemas[:2]
	bundle.Bindings = []SchemaBinding{
		{SchemaID: "fields", ObjectID: "macro-a", Expected: ObjectIdentity{Kind: "macro", Name: "first", Namespace: "search", App: "main", Owner: "nobody"}, SourceCoverage: "complete"},
		{SchemaID: "closed", ObjectID: "missing", Expected: ObjectIdentity{Kind: "lookup", Name: "later", Namespace: "search", App: "main", Owner: "nobody"}, SourceCoverage: "complete"},
	}
	prepared, _, err := PrepareSchemaBundle(bundle)
	if err != nil {
		t.Fatal(err)
	}
	fixture := snapshot.Snapshot()
	for i := range fixture.Collections {
		if fixture.Collections[i].Kind == "lookup" {
			fixture.Collections[i] = Collection{Kind: "lookup", Coverage: "partial", Reason: "sampled"}
		}
	}
	partialSnapshot, _, err := PrepareSnapshot(fixture)
	if err != nil {
		t.Fatal(err)
	}
	env, report, err := Pair(partialSnapshot, prepared)
	if err != nil || env == nil || report.Status != "partial" || len(env.Bindings("macro-a")) != 1 || len(env.Bindings("missing")) != 0 {
		t.Fatalf("valid binding lost with unresolved neighbor: %v %#v", err, report)
	}
	if got := env.Report(); got.Status != "partial" || got.SchemaBundleDigest == "" || got.SnapshotDigest == "" {
		t.Fatalf("stored report: %#v", got)
	}
	value := env.Report()
	value.Diagnostics[0].Code = "mutated"
	if env.Report().Diagnostics[0].Code == "mutated" {
		t.Fatal("report accessor leaked diagnostics")
	}
	artifactReport, err := ValidateArtifacts(fixtureRaw(t, fixture), fixtureRaw(t, bundle))
	if err != nil || artifactReport.Status != "partial" || !hasDiagnostic(artifactReport, "binding_unresolved") {
		t.Fatalf("artifact pairing: %v %#v", err, artifactReport)
	}
	inline, err := ValidateJSON(fixtureRaw(t, map[string]any{"schema_version": 1, "snapshot": fixture, "schema_bundle": bundle}))
	if err != nil || !reflect.DeepEqual(artifactReport, inline) {
		t.Fatalf("inline parity: %v %#v", err, inline)
	}
	badBundle := []byte(`{"schema_version":1,"bundle_id":"bad","schemas":[]}`)
	invalid, err := ValidateArtifacts(fixtureRaw(t, fixture), badBundle)
	if err != nil || invalid.Status != "invalid" || !hasDiagnostic(invalid, "schema_bundle_invalid") || invalid.SnapshotDigest == "" {
		t.Fatalf("standalone invalid diagnostics lost: %v %#v", err, invalid)
	}
}

func TestPairLinkageDiagnosticsUseStableBindingPath(t *testing.T) {
	bundle := bundleFixture(t)
	bundle.Schemas = bundle.Schemas[:2]
	bundle.Bindings = []SchemaBinding{
		{SchemaID: "fields", ObjectID: "missing-z", Expected: ObjectIdentity{Kind: "macro", Name: "z", Namespace: "search", App: "main", Owner: "nobody"}, SourceCoverage: "complete"},
		{SchemaID: "closed", ObjectID: "missing-a", Expected: ObjectIdentity{Kind: "macro", Name: "a", Namespace: "search", App: "main", Owner: "nobody"}, SourceCoverage: "complete"},
	}
	snapshot := snapshotFixture()
	first, err := ValidateArtifacts(fixtureRaw(t, snapshot), fixtureRaw(t, bundle))
	if err != nil || first.Status != "invalid" {
		t.Fatalf("first report: %v %#v", err, first)
	}
	linkages := 0
	for _, diagnostic := range first.Diagnostics {
		if diagnostic.Code == "binding_object_absent" {
			linkages++
			if diagnostic.Path != "/bindings" {
				t.Fatalf("sorted binding index points at wrong input: %#v", diagnostic)
			}
		}
	}
	if linkages != 2 {
		t.Fatalf("expected two linkage diagnostics: %#v", first.Diagnostics)
	}
	bundle.Bindings[0], bundle.Bindings[1] = bundle.Bindings[1], bundle.Bindings[0]
	second, err := ValidateArtifacts(fixtureRaw(t, snapshot), fixtureRaw(t, bundle))
	if err != nil || !reflect.DeepEqual(first, second) {
		t.Fatalf("binding order changed report: %v\nfirst=%#v\nsecond=%#v", err, first, second)
	}
	formatted := FormatReport(first)
	for _, expected := range []string{"Status: invalid", "Snapshot digest: sha256:", "Schema bundle digest: sha256:", "Coverage: schema_bundle macro", "binding_object_absent"} {
		if !strings.Contains(formatted, expected) {
			t.Fatalf("plain report omitted %q: %s", expected, formatted)
		}
	}
}
