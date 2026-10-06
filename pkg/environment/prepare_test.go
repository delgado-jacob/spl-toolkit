package environment

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
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
				_ = env.Snapshot()
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
	if err != nil || env == nil || report.Status != "partial" || !hasDiagnostic(report, "binding_object_absent") {
		t.Fatalf("complete absence: %v %#v", err, report)
	}
	for _, coverage := range []string{"partial", "unavailable"} {
		value.Collections = []Collection{{Kind: "macro", Coverage: coverage, Reason: "limited"}}
		missing, _, _ = PrepareSnapshot(value)
		env, report, err = Pair(missing, bundle)
		if err != nil || env == nil || report.Status != "partial" || !hasDiagnostic(report, "binding_unresolved") || len(env.Bindings("macro-a")) != 2 {
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
	if err != nil || env == nil || report.Status != "partial" || len(env.Bindings("macro-a")) != 1 || len(env.Bindings("missing")) != 1 {
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
	if err != nil || first.Status != "partial" {
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
	for _, expected := range []string{"Status: partial", "Snapshot digest: sha256:", "Schema bundle digest: sha256:", "Coverage: schema_bundle macro", "binding_object_absent"} {
		if !strings.Contains(formatted, expected) {
			t.Fatalf("plain report omitted %q: %s", expected, formatted)
		}
	}
}

func TestPairObservedAbsence(t *testing.T) {
	for _, kind := range []string{"source", "sourcetype", "index"} {
		for _, version := range []int{1, 2} {
			t.Run(fmt.Sprintf("%s/v%d", kind, version), func(t *testing.T) {
				value := observedFixture(t)
				if version == 1 {
					value.SchemaVersion, value.Observation = 1, nil
				}
				bundle := bundleFixture(t)
				bundle.Schemas = bundle.Schemas[:2]
				bundle.Bindings = []SchemaBinding{
					{SchemaID: "fields", ObjectID: "missing", Expected: ObjectIdentity{Kind: kind, Name: "missing"}, SourceCoverage: "complete"},
					{SchemaID: "closed", ObjectID: "missing", Expected: ObjectIdentity{Kind: kind, Name: "missing"}, SourceCoverage: "complete"},
				}
				preparedBundle, bundleReport, err := PrepareSchemaBundle(bundle)
				if err != nil || preparedBundle == nil || bundleReport.Status != "valid" {
					t.Fatalf("prepare bundle: %v %#v", err, bundleReport)
				}
				prepared, snapshotReport, err := PrepareSnapshot(value)
				if err != nil || prepared == nil || snapshotReport.Status != "valid" {
					t.Fatalf("prepare snapshot: %v %#v", err, snapshotReport)
				}
				env, report, err := Pair(prepared, preparedBundle)
				if version == 2 && kind != "index" {
					if err != nil || env == nil || report.Status != "partial" || !hasDiagnostic(report, "binding_unresolved") || hasDiagnostic(report, "binding_object_absent") {
						t.Fatalf("observed absence must stay unresolved: %v %#v", err, report)
					}
					if len(env.Bindings("missing")) != 2 {
						t.Fatal("unobserved object lost bindings")
					}
					if target, ok := env.FieldCatalog("missing", "fields"); !ok || target == nil {
						t.Fatal("unobserved object lost field catalog")
					}
					if target, ok := env.SchemaTarget("missing", "closed"); !ok || target == nil {
						t.Fatal("unobserved object lost schema target")
					}
				} else if err != nil || env == nil || report.Status != "partial" || !hasDiagnostic(report, "binding_object_absent") {
					t.Fatalf("complete catalog absence must remain usable: %v %#v", err, report)
				}
				if report.SchemaVersion != 1 || report.SnapshotDigest != snapshotReport.SnapshotDigest || report.SchemaBundleDigest != bundleReport.SchemaBundleDigest {
					t.Fatalf("pair report changed artifact identities: %#v", report)
				}
			})
		}
	}
}

func TestPairReusesSchemaBundleAcrossSnapshotVersions(t *testing.T) {
	bundle := bundleFixture(t)
	bundle.Schemas = bundle.Schemas[:2]
	identity := ObjectIdentity{Kind: "sourcetype", Name: "audit"}
	bundle.Bindings = []SchemaBinding{
		{SchemaID: "fields", ObjectID: "sourcetype-audit", Expected: identity, SourceCoverage: "partial", Reason: "field source sampled"},
		{SchemaID: "closed", ObjectID: "sourcetype-audit", Expected: identity, SourceCoverage: "complete"},
	}
	preparedBundle, bundleReport, err := PrepareSchemaBundle(bundle)
	if err != nil || preparedBundle == nil || bundleReport.Status != "partial" {
		t.Fatalf("prepare bundle: %v %#v", err, bundleReport)
	}
	var previous *PreparedEnvironment
	var v1Digest string
	for _, capture := range []struct {
		name     string
		version  int
		coverage string
	}{
		{"v1", 1, "complete"}, {"v2-complete", 2, "complete"}, {"v2-partial", 2, "partial"},
	} {
		t.Run(capture.name, func(t *testing.T) {
			value := observedFixture(t)
			if capture.version == 1 {
				value.SchemaVersion, value.Observation = 1, nil
			} else if capture.coverage == "partial" {
				value.Observation.Captures[1].Coverage = "partial"
				value.Observation.Captures[1].Reason = "metadata interrupted"
				setObservedCollection(&value, "sourcetype", "partial", "metadata interrupted")
			}
			prepared, snapshotReport, err := PrepareSnapshot(value)
			if err != nil || prepared == nil || snapshotReport.Status == "invalid" {
				t.Fatalf("prepare snapshot: %v %#v", err, snapshotReport)
			}
			if capture.version == 1 {
				v1Digest = snapshotReport.SnapshotDigest
			} else {
				if snapshotReport.SnapshotDigest == v1Digest {
					t.Fatal("v2 observation did not distinguish snapshot digest")
				}
				copyOfSnapshot := prepared.Snapshot()
				copyOfSnapshot.Observation.AbsenceMeaning = "mutated"
				copyOfSnapshot.Observation.Captures[1].ObjectIDs[0] = "mutated"
				if again := prepared.Snapshot(); again.Observation.AbsenceMeaning != "not_observed" || again.Observation.Captures[1].ObjectIDs[0] != "sourcetype-audit" {
					t.Fatal("snapshot accessor leaked observation state")
				}
			}
			env, report, err := Pair(prepared, preparedBundle)
			if err != nil || env == nil || report.Status != "partial" || len(env.Bindings("sourcetype-audit")) != 2 {
				t.Fatalf("captured binding must resolve: %v %#v", err, report)
			}
			if object, ok := env.Object("sourcetype-audit"); !ok || object.Name != "audit" || object.ID != "sourcetype-audit" {
				t.Fatalf("captured identity changed: %#v", object)
			}
			if collection, ok := env.Collection("sourcetype"); !ok || collection.Coverage != capture.coverage {
				t.Fatalf("schema evidence changed acquisition coverage: %#v", collection)
			}
			for _, binding := range env.Bindings("sourcetype-audit") {
				if binding.SchemaID == "fields" && (binding.SourceCoverage != "partial" || binding.Reason != "field source sampled") || binding.SchemaID == "closed" && binding.SourceCoverage != "complete" {
					t.Fatalf("acquisition changed independent schema evidence: %#v", binding)
				}
			}
			fields, fieldsOK := env.FieldCatalog("sourcetype-audit", "fields")
			closed, closedOK := env.SchemaTarget("sourcetype-audit", "closed")
			if !fieldsOK || fields == nil || !closedOK || closed == nil {
				t.Fatal("captured object lost compiled targets")
			}
			if previous != nil {
				priorFields, _ := previous.FieldCatalog("sourcetype-audit", "fields")
				priorClosed, _ := previous.SchemaTarget("sourcetype-audit", "closed")
				if fields != priorFields || closed != priorClosed {
					t.Fatal("pairing recompiled reusable schema targets")
				}
			}
			previous = env
			if report.SnapshotDigest != snapshotReport.SnapshotDigest || report.SchemaBundleDigest != bundleReport.SchemaBundleDigest || preparedBundle.Bundle().SchemaVersion != 1 {
				t.Fatalf("pairing changed artifact identities: %#v", report)
			}
			mismatch := prepared.Snapshot()
			for i := range mismatch.Objects {
				if mismatch.Objects[i].ID == "sourcetype-audit" {
					mismatch.Objects[i].Name = "different"
				}
			}
			preparedMismatch, mismatchReport, err := PrepareSnapshot(mismatch)
			if err != nil || preparedMismatch == nil {
				t.Fatalf("prepare identity mismatch: %v %#v", err, mismatchReport)
			}
			if bad, report, err := Pair(preparedMismatch, preparedBundle); err != nil || bad != nil || report.Status != "invalid" || !hasDiagnostic(report, "binding_identity_mismatch") {
				t.Fatalf("same ID with wrong identity accepted: %v %#v", err, report)
			}
		})
	}
}

func TestPairAbsentBindingsRetainIndependentTargets(t *testing.T) {
	bundle := preparedPairBundle(t)
	var value Snapshot
	if err := json.Unmarshal(fixtureRaw(t, snapshotFixture()), &value); err != nil {
		t.Fatal(err)
	}
	for _, coverage := range []string{"complete", "partial"} {
		if coverage == "partial" {
			value.Collections = []Collection{{Kind: "macro", Coverage: "partial", Reason: "sampled"}}
		}
		snapshot, _, err := PrepareSnapshot(value)
		if err != nil {
			t.Fatal(err)
		}
		for i := 0; i < 2; i++ {
			env, report, err := Pair(snapshot, bundle)
			if err != nil || env == nil || report.Status != "partial" || len(env.Bindings("macro-a")) != 2 {
				t.Fatalf("%s: %v %#v", coverage, err, report)
			}
			if _, found := env.Object("macro-a"); found {
				t.Fatal("schema binding invented an object")
			}
			fields, found := env.FieldCatalog("macro-a", "fields")
			if !found || fields == nil {
				t.Fatal("absent field catalog lost")
			}
			target, found := env.SchemaTarget("macro-a", "closed")
			if !found || target == nil {
				t.Fatal("absent schema target lost")
			}
			field := analysis.FieldIdentity{Kind: "atomic", Segments: []string{"host"}}
			if got, err := fields.ProjectField(field); err != nil || got.Outcome != "required" {
				t.Fatalf("absent object's selected catalog: %+v %v", got, err)
			}
			if got, err := target.ProjectField(field); err != nil || got.Outcome != "optional" {
				t.Fatalf("absent object's selected schema: %+v %v", got, err)
			}
			if _, found := env.FieldCatalog("macro-a", "closed"); found {
				t.Fatal("independent target kinds mixed")
			}
			copy := env.Snapshot()
			*copy.CaptureScope.Namespace.All = false
			if again := env.Snapshot(); !*again.CaptureScope.Namespace.All {
				t.Fatal("snapshot accessor aliases prepared state")
			}
		}
	}
	conflicting := bundle.Bundle()
	conflicting.Bindings[1].Expected.Name = "contradiction"
	prepared, report, err := PrepareSchemaBundle(conflicting)
	if err != nil || prepared == nil {
		t.Fatalf("structurally valid bundle: %v %#v", err, report)
	}
	for _, coverage := range []string{"complete", "partial"} {
		value.Collections = []Collection{{Kind: "macro", Coverage: coverage}}
		if coverage == "partial" {
			value.Collections[0].Reason = "limited"
		}
		snapshot, _, _ := PrepareSnapshot(value)
		env, report, err := Pair(snapshot, prepared)
		if err != nil || env != nil || report.Status != "invalid" || !hasDiagnostic(report, "binding_identity_mismatch") {
			t.Fatalf("%s contradictory absent identities: %v %#v", coverage, err, report)
		}
	}
}

func TestPreparedEnvironmentSnapshotDetachedObservation(t *testing.T) {
	value := observedFixture(t)
	snapshot, report, err := PrepareSnapshot(value)
	if err != nil || snapshot == nil {
		t.Fatalf("snapshot: %v %#v", err, report)
	}
	env, report, err := Pair(snapshot, nil)
	if err != nil || env == nil {
		t.Fatalf("pair: %v %#v", err, report)
	}
	want := env.Snapshot()
	copy := env.Snapshot()
	*copy.CaptureScope.Namespace.All = false
	copy.Objects[0].Name = "changed"
	copy.Collections[0].Kind = "changed"
	copy.Observation.Indexes[0].CatalogDatatypes[0] = "changed"
	copy.Observation.Captures[1].ObjectIDs[0] = "changed"
	if got := env.Snapshot(); !reflect.DeepEqual(want, got) {
		t.Fatal("paired snapshot aliases observation or capture state")
	}
	var empty *PreparedEnvironment
	if got := empty.Snapshot(); got.SchemaVersion != 0 {
		t.Fatalf("nil accessor: %+v", got)
	}
}

func TestPairCompleteCoverageDoesNotCloseProjection(t *testing.T) {
	snapshot := preparedPairSnapshot(t, "first", "x=1")
	bundle := bundleFixture(t)
	bundle.Schemas = bundle.Schemas[2:3]
	bundle.Bindings = []SchemaBinding{{SchemaID: "open", ObjectID: "macro-a", Expected: ObjectIdentity{Kind: "macro", Name: "first", Namespace: "search", App: "main", Owner: "nobody"}, SourceCoverage: "complete"}}
	prepared, report, err := PrepareSchemaBundle(bundle)
	if err != nil || prepared == nil {
		t.Fatalf("bundle: %v %#v", err, report)
	}
	env, report, err := Pair(snapshot, prepared)
	if err != nil || env == nil || report.Status != "valid" {
		t.Fatalf("pair: %v %#v", err, report)
	}
	target, found := env.SchemaTarget("macro-a", "open")
	if !found {
		t.Fatal("missing selected target")
	}
	got, err := target.ProjectField(analysis.FieldIdentity{Kind: "path", Segments: []string{"unlisted", "child"}})
	if err != nil || got.Admission != analysis.SourceFieldIndeterminate {
		t.Fatalf("complete source coverage closed an open schema: %+v %v", got, err)
	}
}
