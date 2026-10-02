package environment

import (
	"encoding/json"
	"strings"
	"testing"
)

func observedFixture(t *testing.T) Snapshot {
	t.Helper()
	var value Snapshot
	if err := json.Unmarshal(fixtureRaw(t, snapshotFixture()), &value); err != nil {
		t.Fatal(err)
	}
	value.SchemaVersion = 2
	p := Provenance{SourceKind: "fixture", SourceID: "observed-inventory", ObservedAt: "2026-10-01T12:01:00Z"}
	value.Objects = []Object{{ID: "index-main", Kind: "index", Name: "main", Provenance: p}, {ID: "sourcetype-audit", Kind: "sourcetype", Name: "audit", Provenance: p}}
	yes := true
	value.Observation = &ObservationScope{
		IndexSelection: Selector{All: &yes}, Enumeration: IndexEnumeration{Method: "distributed_rest", PeerScope: "configured_search_peers", Coverage: "complete", Provenance: p},
		Indexes: []ObservationIndex{{IndexID: "index-main", CatalogDatatypes: []string{"event"}, RequiredDatatypes: []string{"event"}}}, UnmatchedIndexes: []string{},
		Method: "splunk_metadata", Visibility: "exporting_principal", Window: ObservationWindow{Mode: "all_retained"}, TimePrecision: "bucket_overlap", AbsenceMeaning: "not_observed",
		Captures: []ObservationCapture{{Kind: "source", IndexID: "index-main", Datatype: "event", ObjectIDs: []string{}, Coverage: "complete", Provenance: p}, {Kind: "sourcetype", IndexID: "index-main", Datatype: "event", ObjectIDs: []string{"sourcetype-audit"}, Coverage: "complete", Provenance: p}},
	}
	return value
}
func TestSnapshotV2ObservedScope(t *testing.T) {
	value := observedFixture(t)
	prepared, report, err := PrepareSnapshot(value)
	if err != nil || prepared == nil || report.Status != "valid" {
		t.Fatalf("prepare v2: %v %#v", err, report)
	}
	saved := prepared.Snapshot()
	value.Observation.Captures[1].ObjectIDs[0] = "mutated"
	if saved.Observation.Captures[1].ObjectIDs[0] != "sourcetype-audit" {
		t.Fatal("prepared observation did not detach input")
	}
	saved.Observation.Captures[1].ObjectIDs[0] = "changed"
	if prepared.Snapshot().Observation.Captures[1].ObjectIDs[0] != "sourcetype-audit" {
		t.Fatal("snapshot accessor exposed prepared state")
	}
	value = observedFixture(t)
	value.CaptureScope.App = Selector{Values: []string{"search"}}
	if got := fixtureReport(t, value); got.Status != "valid" {
		t.Fatalf("knowledge selector restricted observation: %#v", got)
	}
}
func TestSnapshotV1DigestCompatibility(t *testing.T) {
	if got := fixtureReport(t, snapshotFixture()).SnapshotDigest; got != "sha256:3d43db2e033c7d57830b886adb4d3dadf07cd2d27722f1db1e8a56a9f1973912" {
		t.Fatalf("v1 digest changed: %s", got)
	}
	for _, observation := range []any{nil, map[string]any{}} {
		value := snapshotFixture()
		value["observation"] = observation
		if got := fixtureReport(t, value); got.Status != "invalid" {
			t.Fatalf("v1 observation accepted: %#v", got)
		}
	}
}
func TestSnapshotV2RequiredModes(t *testing.T) {
	value := observedFixture(t)
	value.Observation.Indexes[0].CatalogDatatypes = []string{}
	value.Observation.Indexes[0].RequiredDatatypes = []string{"metric", "event", "event"}
	for _, c := range append([]ObservationCapture{}, value.Observation.Captures...) {
		c.Datatype = "metric"
		value.Observation.Captures = append(value.Observation.Captures, c)
	}
	prepared, report, err := PrepareSnapshot(value)
	if err != nil || prepared == nil || report.Status != "valid" {
		t.Fatalf("unreported datatypes: %v %#v", err, report)
	}
	digest := report.SnapshotDigest
	value.Observation.Indexes[0].RequiredDatatypes = []string{"event", "metric"}
	value.Observation.Captures[0], value.Observation.Captures[3] = value.Observation.Captures[3], value.Observation.Captures[0]
	if got := fixtureReport(t, value); got.Status != "valid" || got.SnapshotDigest != digest {
		t.Fatalf("set ordering changed digest: %#v", got)
	}
	value.Observation.Window = ObservationWindow{Mode: "bounded", Earliest: "2026-09-01T00:00:00Z"}
	if got := fixtureReport(t, value); got.Status != "valid" || got.SnapshotDigest == digest {
		t.Fatalf("window did not change digest: %#v", got)
	}
	value = observedFixture(t)
	value.Observation.Captures[1].ObjectIDs = []string{"sourcetype-audit", "sourcetype-audit"}
	if got := fixtureReport(t, value); got.Status != "valid" {
		t.Fatalf("repeated object reference: %#v", got)
	}
}
func setObservedCollection(value *Snapshot, kind, coverage, reason string) {
	for i := range value.Collections {
		if value.Collections[i].Kind == kind {
			value.Collections[i].Coverage = coverage
			value.Collections[i].Reason = reason
		}
	}
}
func TestSnapshotV2UnmatchedIndexes(t *testing.T) {
	value := observedFixture(t)
	value.Observation.IndexSelection = Selector{Values: []string{"main", "missing"}}
	value.Observation.UnmatchedIndexes = []string{"missing"}
	for _, kind := range []string{"source", "sourcetype"} {
		setObservedCollection(&value, kind, "partial", "requested index unconfirmed")
	}
	if got := fixtureReport(t, value); got.Status != "partial" {
		t.Fatalf("unmatched scope: %#v", got)
	}
	value.Observation.Enumeration.Coverage = "partial"
	value.Observation.Enumeration.Reason = "peer unavailable"
	setObservedCollection(&value, "index", "partial", "peer unavailable")
	if got := fixtureReport(t, value); got.Status != "partial" {
		t.Fatalf("partial enumeration with confirmed index: %#v", got)
	}
	value = observedFixture(t)
	value.Objects = []Object{}
	value.Observation.Indexes = []ObservationIndex{}
	value.Observation.Captures = []ObservationCapture{}
	if got := fixtureReport(t, value); got.Status != "valid" {
		t.Fatalf("complete empty all index observation: %#v", got)
	}
	value.Observation.IndexSelection = Selector{Values: []string{"missing"}}
	value.Observation.UnmatchedIndexes = []string{"missing"}
	if got := fixtureReport(t, value); got.Status != "invalid" {
		t.Fatalf("unmatched exact scope declared complete: %#v", got)
	}
	value.Observation.Enumeration.Coverage = "unavailable"
	value.Observation.Enumeration.Reason = "timeout"
	for _, kind := range []string{"index", "source", "sourcetype"} {
		setObservedCollection(&value, kind, "unavailable", "timeout")
	}
	if got := fixtureReport(t, value); got.Status != "partial" {
		t.Fatalf("unavailable enumeration: %#v", got)
	}
}
func TestSnapshotV2RejectsInvalidObservation(t *testing.T) {
	cases := map[string]func(*Snapshot){
		"missing observation": func(v *Snapshot) { v.Observation = nil },
		"missing indexes":     func(v *Snapshot) { v.Observation.Indexes = nil },
		"missing unmatched":   func(v *Snapshot) { v.Observation.UnmatchedIndexes = nil },
		"missing captures":    func(v *Snapshot) { v.Observation.Captures = nil },
		"missing catalog":     func(v *Snapshot) { v.Observation.Indexes[0].CatalogDatatypes = nil },
		"missing required":    func(v *Snapshot) { v.Observation.Indexes[0].RequiredDatatypes = nil },
		"missing object refs": func(v *Snapshot) { v.Observation.Captures[0].ObjectIDs = nil },
		"missing mode record": func(v *Snapshot) { v.Observation.Captures = v.Observation.Captures[:1] },
		"unknown datatype":    func(v *Snapshot) { v.Observation.Indexes[0].CatalogDatatypes = []string{"future"} },
		"wrong required mode": func(v *Snapshot) { v.Observation.Indexes[0].RequiredDatatypes = []string{"metric"} },
		"wrong index ref":     func(v *Snapshot) { v.Observation.Indexes[0].IndexID = "sourcetype-audit" },
		"wrong object ref":    func(v *Snapshot) { v.Observation.Captures[1].ObjectIDs = []string{"index-main"} },
		"unknown object ref":  func(v *Snapshot) { v.Observation.Captures[1].ObjectIDs = []string{"missing"} },
		"duplicate capture":   func(v *Snapshot) { v.Observation.Captures = append(v.Observation.Captures, v.Observation.Captures[0]) },
		"duplicate index":     func(v *Snapshot) { v.Observation.Indexes = append(v.Observation.Indexes, v.Observation.Indexes[0]) },
		"conflicting identity": func(v *Snapshot) {
			o := v.Objects[1]
			o.ID = "different"
			v.Objects = append(v.Objects, o)
			v.Observation.Captures[1].ObjectIDs = append(v.Observation.Captures[1].ObjectIDs, o.ID)
		},
		"unassociated object": func(v *Snapshot) { v.Observation.Captures[1].ObjectIDs = []string{} },
		"unavailable with objects": func(v *Snapshot) {
			v.Observation.Captures[1].Coverage = "unavailable"
			v.Observation.Captures[1].Reason = "timeout"
		},
		"aggregate contradiction": func(v *Snapshot) {
			v.Observation.Captures[0].Coverage = "partial"
			v.Observation.Captures[0].Reason = "timeout"
		},
		"missing reason":      func(v *Snapshot) { v.Observation.Enumeration.Coverage = "partial" },
		"unsupported host":    func(v *Snapshot) { v.Observation.Captures[0].Kind = "host" },
		"method":              func(v *Snapshot) { v.Observation.Method = "configured" },
		"visibility":          func(v *Snapshot) { v.Observation.Visibility = "all" },
		"precision":           func(v *Snapshot) { v.Observation.TimePrecision = "exact" },
		"absence":             func(v *Snapshot) { v.Observation.AbsenceMeaning = "absent" },
		"enumeration method":  func(v *Snapshot) { v.Observation.Enumeration.Method = "local" },
		"peer scope":          func(v *Snapshot) { v.Observation.Enumeration.PeerScope = "all" },
		"bounds all retained": func(v *Snapshot) { v.Observation.Window.Earliest = "2026-09-01T00:00:00Z" },
		"unbounded bounded":   func(v *Snapshot) { v.Observation.Window.Mode = "bounded" },
		"reversed bounds": func(v *Snapshot) {
			v.Observation.Window = ObservationWindow{Mode: "bounded", Earliest: "2026-09-02T00:00:00Z", Latest: "2026-09-01T00:00:00Z"}
		},
		"equal bounds": func(v *Snapshot) {
			v.Observation.Window = ObservationWindow{Mode: "bounded", Earliest: "2026-09-01T00:00:00Z", Latest: "2026-09-01T00:00:00Z"}
		},
		"non UTC bound": func(v *Snapshot) {
			v.Observation.Window = ObservationWindow{Mode: "bounded", Latest: "2026-09-01T00:00:00+01:00"}
		},
		"unknown mode":                func(v *Snapshot) { v.Observation.Window.Mode = "recent" },
		"invalid Unicode":             func(v *Snapshot) { v.Observation.Method = string([]byte{0xff}) },
		"unmatched all":               func(v *Snapshot) { v.Observation.UnmatchedIndexes = []string{"missing"} },
		"unselected index":            func(v *Snapshot) { v.Observation.IndexSelection = Selector{Values: []string{"other"}} },
		"provenance outside interval": func(v *Snapshot) { v.Observation.Captures[0].Provenance.ObservedAt = "2026-10-01T13:00:00Z" },
	}
	for name, change := range cases {
		t.Run(name, func(t *testing.T) {
			value := observedFixture(t)
			change(&value)
			_, report, err := PrepareSnapshot(value)
			if err != nil || report.Status != "invalid" {
				t.Fatalf("accepted invalid observation: %v %#v", err, report)
			}
		})
	}
}
func TestSnapshotV2StrictObservationJSON(t *testing.T) {
	raw := string(fixtureRaw(t, observedFixture(t)))
	for name, modified := range map[string]string{
		"unknown":   strings.Replace(raw, `"method":"splunk_metadata"`, `"method":"splunk_metadata","password":"secret"`, 1),
		"duplicate": strings.Replace(raw, `"mode":"all_retained"`, `"mode":"all_retained","mode":"bounded"`, 1),
		"Unicode":   strings.Replace(raw, `"method":"splunk_metadata"`, `"method":"\ud800"`, 1),
	} {
		t.Run(name, func(t *testing.T) {
			report, err := ValidateArtifacts([]byte(modified), nil)
			if err != nil || report.Status != "invalid" {
				t.Fatalf("invalid JSON accepted: %v %#v", err, report)
			}
		})
	}
	value := observedFixture(t)
	value.Observation = nil
	value.Collections = []Collection{}
	value.Objects = []Object{}
	if got := fixtureReport(t, value); got.Status != "partial" {
		t.Fatalf("knowledge-only v2 without observation: %#v", got)
	}
}

func TestSnapshotV2DeclaredKindsAndCatalogSelection(t *testing.T) {
	value := observedFixture(t)
	value.Collections = []Collection{{Kind: "index", Coverage: "complete"}}
	value.Objects = value.Objects[:1]
	value.Observation.Captures = []ObservationCapture{}
	if got := fixtureReport(t, value); got.Status != "partial" {
		t.Fatalf("index-only snapshot: %#v", got)
	}
	value.Observation.Captures = observedFixture(t).Observation.Captures[:1]
	if got := fixtureReport(t, value); got.Status != "invalid" {
		t.Fatalf("undeclared source capture accepted: %#v", got)
	}
	value = observedFixture(t)
	extra := value.Objects[0]
	extra.ID = "index-extra"
	extra.Name = "extra"
	value.Objects = append(value.Objects, extra)
	if got := fixtureReport(t, value); got.Status != "invalid" {
		t.Fatalf("all selection omitted confirmed catalog index: %#v", got)
	}
	value.Observation.IndexSelection = Selector{Values: []string{"main"}}
	if got := fixtureReport(t, value); got.Status != "valid" {
		t.Fatalf("unselected catalog index: %#v", got)
	}
	value.Observation.IndexSelection.Values = []string{"main", "extra"}
	value.Observation.UnmatchedIndexes = []string{"extra"}
	for _, kind := range []string{"source", "sourcetype"} {
		setObservedCollection(&value, kind, "partial", "unmatched")
	}
	if got := fixtureReport(t, value); got.Status != "invalid" {
		t.Fatalf("confirmed catalog entry declared unmatched: %#v", got)
	}
}
func TestSnapshotV2UnavailableAndMixedCaptures(t *testing.T) {
	value := observedFixture(t)
	value.Objects = value.Objects[:1]
	for i := range value.Observation.Captures {
		c := &value.Observation.Captures[i]
		c.Coverage = "unavailable"
		c.Reason = "timeout"
		c.ObjectIDs = []string{}
	}
	for _, kind := range []string{"source", "sourcetype"} {
		setObservedCollection(&value, kind, "unavailable", "timeout")
	}
	if got := fixtureReport(t, value); got.Status != "partial" {
		t.Fatalf("unavailable capture acquisition: %#v", got)
	}
	value.Observation.Indexes[0].CatalogDatatypes = []string{"event", "metric"}
	value.Observation.Indexes[0].RequiredDatatypes = []string{"event", "metric"}
	for _, c := range append([]ObservationCapture{}, value.Observation.Captures...) {
		c.Datatype = "metric"
		c.Coverage = "complete"
		c.Reason = ""
		value.Observation.Captures = append(value.Observation.Captures, c)
	}
	for _, kind := range []string{"source", "sourcetype"} {
		setObservedCollection(&value, kind, "partial", "event timeout")
	}
	if got := fixtureReport(t, value); got.Status != "partial" {
		t.Fatalf("mixed capture acquisition: %#v", got)
	}
}
func TestSnapshotV2ExplicitForbiddenMetadata(t *testing.T) {
	value := observedFixture(t)
	value.Objects[1].Arguments = []string{}
	if _, got, err := PrepareSnapshot(value); err != nil || got.Status != "invalid" {
		t.Fatalf("source query metadata accepted: %v %#v", err, got)
	}
	value = observedFixture(t)
	raw := string(fixtureRaw(t, value))
	raw = strings.Replace(raw, `"mode":"all_retained"`, `"mode":"all_retained","earliest":""`, 1)
	report, err := ValidateArtifacts([]byte(raw), nil)
	if err != nil || report.Status != "invalid" {
		t.Fatalf("explicit bound on all_retained: %v %#v", err, report)
	}
}

func TestSnapshotV2AssociationDigest(t *testing.T) {
	value := observedFixture(t)
	index := value.Objects[0]
	index.ID = "index-other"
	index.Name = "other"
	audit := value.Objects[1]
	audit.ID = "sourcetype-other"
	audit.Name = "other"
	value.Objects = append(value.Objects, index, audit)
	value.Observation.Indexes = append(value.Observation.Indexes, ObservationIndex{IndexID: index.ID, CatalogDatatypes: []string{"event"}, RequiredDatatypes: []string{"event"}})
	for _, capture := range append([]ObservationCapture{}, value.Observation.Captures...) {
		capture.IndexID = index.ID
		if capture.Kind == "sourcetype" {
			capture.ObjectIDs = []string{audit.ID}
		}
		value.Observation.Captures = append(value.Observation.Captures, capture)
	}
	first := fixtureReport(t, value)
	if first.Status != "valid" {
		t.Fatalf("two indexes: %#v", first)
	}
	value.Observation.Indexes[0], value.Observation.Indexes[1] = value.Observation.Indexes[1], value.Observation.Indexes[0]
	value.Observation.Captures[0], value.Observation.Captures[3] = value.Observation.Captures[3], value.Observation.Captures[0]
	if got := fixtureReport(t, value); got.SnapshotDigest != first.SnapshotDigest {
		t.Fatalf("observation set order changed digest: %#v", got)
	}
	for i := range value.Observation.Captures {
		capture := &value.Observation.Captures[i]
		if capture.Kind == "sourcetype" {
			if capture.IndexID == "index-main" {
				capture.ObjectIDs = []string{audit.ID}
			} else {
				capture.ObjectIDs = []string{"sourcetype-audit"}
			}
		}
	}
	if got := fixtureReport(t, value); got.Status != "valid" || got.SnapshotDigest == first.SnapshotDigest {
		t.Fatalf("association change omitted from digest: %#v", got)
	}
}
func TestSnapshotV2RequiredObservationArraysJSON(t *testing.T) {
	for _, path := range [][]string{{"indexes"}, {"unmatched_indexes"}, {"captures"}, {"indexes", "catalog_datatypes"}, {"indexes", "required_datatypes"}, {"captures", "object_ids"}} {
		for _, mode := range []string{"missing", "null"} {
			t.Run(strings.Join(path, "/")+"/"+mode, func(t *testing.T) {
				var root map[string]any
				if err := json.Unmarshal(fixtureRaw(t, observedFixture(t)), &root); err != nil {
					t.Fatal(err)
				}
				target := root["observation"].(map[string]any)
				key := path[0]
				if len(path) == 2 {
					target = target[key].([]any)[0].(map[string]any)
					key = path[1]
				}
				if mode == "missing" {
					delete(target, key)
				} else {
					target[key] = nil
				}
				got := fixtureReport(t, root)
				if got.Status != "invalid" || !strings.HasPrefix(got.Diagnostics[0].Path, "/observation/") {
					t.Fatalf("required observation array accepted: %#v", got)
				}
			})
		}
	}
}

func TestSnapshotV2ZeroTimeBounds(t *testing.T) {
	for _, window := range []ObservationWindow{
		{Mode: "bounded", Earliest: "0002-01-01T00:00:00Z", Latest: "0001-01-01T00:00:00Z"},
		{Mode: "bounded", Earliest: "0001-01-01T00:00:00Z", Latest: "0001-01-01T00:00:00Z"},
	} {
		t.Run(window.Earliest+"/"+window.Latest, func(t *testing.T) {
			value := observedFixture(t)
			value.Observation.Window = window
			report, err := ValidateArtifacts(fixtureRaw(t, value), nil)
			if err != nil || report.Status != "invalid" || report.Diagnostics[0].Path != "/observation/window" {
				t.Fatalf("unordered zero-time bounds accepted: %v %#v", err, report)
			}
		})
	}
	value := observedFixture(t)
	value.Observation.Window = ObservationWindow{Mode: "bounded", Earliest: "0001-01-01T00:00:00Z", Latest: "0002-01-01T00:00:00Z"}
	if got := fixtureReport(t, value); got.Status != "valid" {
		t.Fatalf("ordered zero-time bounds rejected: %#v", got)
	}
	value.Observation.Window = ObservationWindow{Mode: "bounded", Latest: "0001-01-01T00:00:00Z"}
	if got := fixtureReport(t, value); got.Status != "valid" {
		t.Fatalf("single zero-time bound rejected: %#v", got)
	}
}
func TestSnapshotV2ObservedSharingContext(t *testing.T) {
	for _, kind := range []string{"index", "source", "sourcetype"} {
		t.Run(kind, func(t *testing.T) {
			value := observedFixture(t)
			if kind == "source" {
				object := value.Objects[1]
				object.Kind = "source"
				object.ID = "source-audit"
				value.Objects = append(value.Objects, object)
				value.Observation.Captures[0].ObjectIDs = []string{object.ID}
			}
			for i := range value.Objects {
				if value.Objects[i].Kind == kind {
					value.Objects[i].Sharing = "app"
				}
			}
			if got := fixtureReport(t, value); got.Status != "invalid" || !strings.HasPrefix(got.Diagnostics[0].Path, "/objects/") {
				t.Fatalf("v2 observed ACL context accepted: %#v", got)
			}
			value.SchemaVersion = 1
			value.Observation = nil
			if got := fixtureReport(t, value); got.Status != "valid" {
				t.Fatalf("v1 sharing behavior changed: %#v", got)
			}
		})
	}
}
