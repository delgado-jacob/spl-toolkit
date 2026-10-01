package environment

import (
	"strings"
	"testing"
)

func TestSnapshotNestedDuplicateMember(t *testing.T) {
	raw := fixtureRaw(t, snapshotFixture())
	// Duplicate a nested member to ensure the raw decoder checks every depth.
	raw = append(raw[:0:0], raw...)
	for i := 0; i+len(`"all":true`) <= len(raw); i++ {
		if string(raw[i:i+len(`"all":true`)]) == `"all":true` {
			duplicate := append([]byte{}, raw[:i+len(`"all":true`)]...)
			duplicate = append(duplicate, []byte(`,"all":true`)...)
			raw = append(duplicate, raw[i+len(`"all":true`):]...)
			break
		}
	}
	report, err := ValidateArtifacts(raw, nil)
	if err != nil {
		t.Fatal(err)
	}
	if report == nil || report.Status != "invalid" {
		t.Fatalf("expected invalid duplicate-member report: %#v", report)
	}
}

func TestSnapshotExactMembersAndLocations(t *testing.T) {
	cases := []struct {
		name, old, replacement, path string
		offset                       bool
	}{
		{name: "case variant top member", old: `"scope_id":"capture-a"`, replacement: `"scope_id":"capture-a","SCOPE_ID":"other"`, path: "/SCOPE_ID"},
		{name: "case variant nested member", old: `"producer":"fixture"`, replacement: `"producer":"fixture","PRODUCER":"other"`, path: "/origin/PRODUCER"},
		{name: "duplicate nested member", old: `"producer":"fixture"`, replacement: `"producer":"fixture","producer":"other"`, offset: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			raw := []byte(strings.Replace(string(fixtureRaw(t, snapshotFixture())), tc.old, tc.replacement, 1))
			report, err := ValidateArtifacts(raw, nil)
			if err != nil {
				t.Fatal(err)
			}
			if report.Status != "invalid" || len(report.Diagnostics) != 1 {
				t.Fatalf("expected invalid: %#v", report)
			}
			got := report.Diagnostics[0]
			if tc.path != "" && got.Path != tc.path {
				t.Fatalf("path = %q, want %q", got.Path, tc.path)
			}
			if tc.offset && got.ByteOffset == nil {
				t.Fatalf("missing byte offset: %#v", got)
			}
		})
	}
}

func TestInlineRequestDuplicateKeepsLocation(t *testing.T) {
	report, err := ValidateJSON([]byte(`{"schema_version":1,"schema_version":1,"snapshot":{}}`))
	if err != nil {
		t.Fatal(err)
	}
	if report.Status != "invalid" || len(report.Diagnostics) != 1 {
		t.Fatalf("request report: %#v", report)
	}
	diagnostic := report.Diagnostics[0]
	if diagnostic.Artifact != "request" || diagnostic.Path != "/schema_version" || diagnostic.ByteOffset == nil {
		t.Fatalf("request location lost: %#v", diagnostic)
	}
}

func TestSnapshotWrongTypeKeepsNestedLocation(t *testing.T) {
	raw := []byte(strings.Replace(string(fixtureRaw(t, snapshotFixture())), `"instance_id":"instance-a"`, `"instance_id":123`, 1))
	report, err := ValidateArtifacts(raw, nil)
	if err != nil {
		t.Fatal(err)
	}
	if report.Status != "invalid" || len(report.Diagnostics) != 1 {
		t.Fatalf("wrong-type report: %#v", report)
	}
	diagnostic := report.Diagnostics[0]
	if diagnostic.Path != "/origin/instance_id" || diagnostic.ByteOffset == nil {
		t.Fatalf("nested location lost: %#v", diagnostic)
	}
}

func TestSnapshotWrongTypeInsideArrayUsesByteOffset(t *testing.T) {
	value := snapshotFixture()
	object := macroFixture("m", "name", "x")
	object["document"].(map[string]any)["text"] = 123
	value["objects"] = []any{object}
	report, err := ValidateArtifacts(fixtureRaw(t, value), nil)
	if err != nil {
		t.Fatal(err)
	}
	if report.Status != "invalid" || len(report.Diagnostics) != 1 {
		t.Fatalf("wrong-type report: %#v", report)
	}
	diagnostic := report.Diagnostics[0]
	if diagnostic.Path != "" || diagnostic.ByteOffset == nil {
		t.Fatalf("array error location must use byte offset: %#v", diagnostic)
	}
}
