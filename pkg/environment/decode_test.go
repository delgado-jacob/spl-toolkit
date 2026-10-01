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
