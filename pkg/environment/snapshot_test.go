package environment

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

var snapshotKinds = []string{"index", "source", "sourcetype", "dataset", "data_model", "lookup", "macro", "saved_search", "event_type", "tag", "calculated_field", "field_extraction", "module", "function", "external_command"}

func snapshotFixture() map[string]any {
	collections := make([]any, 0, len(snapshotKinds))
	for _, kind := range snapshotKinds {
		collections = append(collections, map[string]any{"kind": kind, "coverage": "complete"})
	}
	return map[string]any{
		"schema_version": 1, "scope_id": "capture-a",
		"capture_scope": map[string]any{"namespace": map[string]any{"all": true}, "app": map[string]any{"all": true}, "owner": map[string]any{"all": true}},
		"origin":        map[string]any{"instance_id": "instance-a", "product_version": "9.4", "producer": "fixture", "producer_version": "1"},
		"capture":       map[string]any{"start": "2026-10-01T12:00:00Z", "end": "2026-10-01T12:05:00Z"},
		"capabilities":  []any{}, "collections": collections, "objects": []any{},
	}
}
func fixtureRaw(t *testing.T, v any) []byte {
	t.Helper()
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}
func fixtureReport(t *testing.T, v any) *Report {
	t.Helper()
	report, err := ValidateArtifacts(fixtureRaw(t, v), nil)
	if err != nil {
		t.Fatal(err)
	}
	if report == nil {
		t.Fatal("nil report")
	}
	return report
}
func scopedCollections(v map[string]any) {
	for _, item := range v["collections"].([]any) {
		c := item.(map[string]any)
		if c["kind"] == "index" || c["kind"] == "source" || c["kind"] == "sourcetype" {
			c["coverage"] = "partial"
			c["reason"] = "global kind outside restricted scope"
		}
	}
}
func macroFixture(id, name, definition string) map[string]any {
	return map[string]any{"id": id, "kind": "macro", "name": name, "namespace": "search", "app": "main", "owner": "nobody", "provenance": map[string]any{"source_kind": "rest", "source_id": id, "observed_at": "2026-10-01T12:02:00Z"}, "document": map[string]any{"text": definition}, "arity": 0, "arguments": []any{}, "relations": []any{}}
}
func TestSnapshotScopedCoverageAndDigest(t *testing.T) {
	complete := snapshotFixture()
	report := fixtureReport(t, complete)
	if report.Status != "valid" || !strings.HasPrefix(report.SnapshotDigest, "sha256:") {
		t.Fatalf("complete empty snapshot: %#v", report)
	}
	if len(report.Diagnostics) != 0 {
		t.Fatalf("unexpected complete diagnostics: %#v", report.Diagnostics)
	}
	partial := snapshotFixture()
	partial["collections"] = []any{map[string]any{"kind": "macro", "coverage": "complete"}}
	report = fixtureReport(t, partial)
	if report.Status != "partial" || len(report.Diagnostics) != 14 {
		t.Fatalf("omitted kinds: %#v", report)
	}
	for _, d := range report.Diagnostics {
		if d.Code != "collection_omitted" {
			t.Fatalf("unexpected diagnostic: %#v", d)
		}
	}

	first := snapshotFixture()
	scopedCollections(first)
	first["capture_scope"] = map[string]any{"namespace": map[string]any{"values": []any{"search", "default"}}, "app": map[string]any{"values": []any{"main", "other"}}, "owner": map[string]any{"values": []any{"nobody", "alice"}}}
	first["objects"] = []any{macroFixture("macro-b", "second", "x=2"), macroFixture("macro-a", "first", "x=1")}
	first["capabilities"] = []any{map[string]any{"id": "cap-b", "version": "1", "state": "unknown", "provenance": map[string]any{"source_kind": "product", "source_id": "facts", "observed_at": "2026-10-01T12:03:00Z"}}, map[string]any{"id": "cap-a", "version": "1", "state": "available", "provenance": map[string]any{"source_kind": "product", "source_id": "facts", "observed_at": "2026-10-01T12:03:00Z"}}}
	digest := fixtureReport(t, first).SnapshotDigest
	if !strings.HasPrefix(digest, "sha256:") {
		t.Fatal("scoped snapshot did not produce digest")
	}
	second := snapshotFixture()
	scopedCollections(second)
	second["capture_scope"] = map[string]any{"namespace": map[string]any{"values": []any{"default", "search"}}, "app": map[string]any{"values": []any{"other", "main"}}, "owner": map[string]any{"values": []any{"alice", "nobody"}}}
	originalCollections := second["collections"].([]any)
	for i, j := 0, len(originalCollections)-1; i < j; i, j = i+1, j-1 {
		originalCollections[i], originalCollections[j] = originalCollections[j], originalCollections[i]
	}
	second["objects"] = []any{macroFixture("macro-a", "first", "x=1"), macroFixture("macro-b", "second", "x=2")}
	caps := first["capabilities"].([]any)
	second["capabilities"] = []any{caps[1], caps[0]}
	if got := fixtureReport(t, second).SnapshotDigest; got != digest {
		t.Fatalf("set reordering changed digest: %s != %s", got, digest)
	}
	second["objects"].([]any)[0].(map[string]any)["document"].(map[string]any)["text"] = "x=1 "
	if got := fixtureReport(t, second).SnapshotDigest; got == digest {
		t.Fatal("definition byte change did not change digest")
	}
}

func TestSnapshotRejectsContradictions(t *testing.T) {
	cases := map[string]func(map[string]any) []byte{
		"duplicate object ID": func(v map[string]any) []byte {
			v["objects"] = []any{macroFixture("same", "first", "x"), macroFixture("same", "second", "y")}
			return fixtureRaw(t, v)
		},
		"duplicate capability fact": func(v map[string]any) []byte {
			p := map[string]any{"source_kind": "product", "source_id": "facts", "observed_at": "2026-10-01T12:03:00Z"}
			c := map[string]any{"id": "same", "version": "1", "state": "available", "provenance": p}
			v["capabilities"] = []any{c, c}
			return fixtureRaw(t, v)
		},
		"unsupported capability state": func(v map[string]any) []byte {
			v["capabilities"] = []any{map[string]any{"id": "cap", "version": "1", "state": "maybe", "provenance": map[string]any{"source_kind": "product", "source_id": "facts", "observed_at": "2026-10-01T12:03:00Z"}}}
			return fixtureRaw(t, v)
		},
		"null objects":         func(v map[string]any) []byte { v["objects"] = nil; return fixtureRaw(t, v) },
		"missing capabilities": func(v map[string]any) []byte { delete(v, "capabilities"); return fixtureRaw(t, v) },
		"null selector branch": func(v map[string]any) []byte {
			v["capture_scope"].(map[string]any)["app"] = map[string]any{"all": nil}
			return fixtureRaw(t, v)
		},
		"both selector branches": func(v map[string]any) []byte {
			v["capture_scope"].(map[string]any)["app"] = map[string]any{"all": true, "values": nil}
			return fixtureRaw(t, v)
		},
		"unsupported version": func(v map[string]any) []byte { v["schema_version"] = 2; return fixtureRaw(t, v) },
		"object outside scope": func(v map[string]any) []byte {
			v["capture_scope"].(map[string]any)["app"] = map[string]any{"values": []any{"main"}}
			v["objects"] = []any{macroFixture("m", "foo", "x")}
			v["objects"].([]any)[0].(map[string]any)["app"] = "other"
			return fixtureRaw(t, v)
		},
		"unavailable with object": func(v map[string]any) []byte {
			for _, c := range v["collections"].([]any) {
				if c.(map[string]any)["kind"] == "macro" {
					c.(map[string]any)["coverage"] = "unavailable"
					c.(map[string]any)["reason"] = "blocked"
				}
			}
			v["objects"] = []any{macroFixture("m", "foo", "x")}
			return fixtureRaw(t, v)
		},
		"inapplicable restricted app": func(v map[string]any) []byte {
			v["capture_scope"].(map[string]any)["app"] = map[string]any{"values": []any{"main"}}
			return fixtureRaw(t, v)
		},
		"reversed interval": func(v map[string]any) []byte {
			v["capture"].(map[string]any)["end"] = "2026-10-01T11:59:00Z"
			return fixtureRaw(t, v)
		},
		"observation outside interval": func(v map[string]any) []byte {
			o := macroFixture("m", "foo", "x")
			o["provenance"].(map[string]any)["observed_at"] = "2026-10-01T13:00:00Z"
			v["objects"] = []any{o}
			return fixtureRaw(t, v)
		},
		"unknown credential property": func(v map[string]any) []byte { v["password"] = "secret"; return fixtureRaw(t, v) },
		"asserted digest mismatch": func(v map[string]any) []byte {
			v["digest"] = "sha256:" + strings.Repeat("0", 64)
			return fixtureRaw(t, v)
		},
		"duplicate JSON key": func(v map[string]any) []byte {
			raw := fixtureRaw(t, v)
			return []byte(strings.Replace(string(raw), `"scope_id":"capture-a"`, `"scope_id":"capture-a","scope_id":"other"`, 1))
		},
		"invalid Unicode": func(v map[string]any) []byte {
			raw := fixtureRaw(t, v)
			return []byte(strings.Replace(string(raw), `"scope_id":"capture-a"`, `"scope_id":"\ud800"`, 1))
		},
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			raw := mutate(snapshotFixture())
			report, err := ValidateArtifacts(raw, nil)
			if err != nil {
				t.Fatal(err)
			}
			if report == nil || report.Status != "invalid" || len(report.Diagnostics) == 0 {
				t.Fatalf("%s: expected invalid diagnostic; got %#v; raw=%s", name, report, fmt.Sprint(string(raw)))
			}
		})
	}
}

func TestSnapshotTypedInputRejectsInvalidUnicodeAndDoesNotMutate(t *testing.T) {
	var value Snapshot
	if err := json.Unmarshal(fixtureRaw(t, snapshotFixture()), &value); err != nil {
		t.Fatal(err)
	}
	before := fixtureRaw(t, value)
	prepared, report, err := PrepareSnapshot(value)
	if err != nil || prepared == nil || report.Status != "valid" {
		t.Fatalf("prepare: %v %#v", err, report)
	}
	if string(fixtureRaw(t, value)) != string(before) {
		t.Fatal("caller snapshot mutated")
	}
	yes := true
	value.CaptureScope.App = Selector{All: &yes, Values: []string{}}
	prepared, report, err = PrepareSnapshot(value)
	if err != nil || prepared != nil || report.Status != "invalid" {
		t.Fatalf("contradictory typed selector accepted: %v %#v", err, report)
	}
	value.CaptureScope.App = Selector{All: &yes}
	value.ScopeID = string([]byte{0xff})
	prepared, report, err = PrepareSnapshot(value)
	if err != nil || prepared != nil || report.Status != "invalid" {
		t.Fatalf("invalid typed UTF-8 accepted: %v %#v", err, report)
	}
}
