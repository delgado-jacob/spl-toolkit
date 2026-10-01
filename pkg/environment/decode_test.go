package environment

import "testing"

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
