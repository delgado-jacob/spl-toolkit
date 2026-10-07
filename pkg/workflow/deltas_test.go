package workflow

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/delgado-jacob/spl-toolkit/pkg/environment"
	"github.com/delgado-jacob/spl-toolkit/pkg/impact"
)

func TestClassifyComparisonPrecedence(t *testing.T) {
	for _, tc := range []struct {
		failed, delta, complete, aligned bool
		want                             impact.Classification
	}{
		{true, true, true, true, impact.Failed}, {false, true, false, false, impact.Affected}, {false, false, false, true, impact.Indeterminate}, {false, false, true, false, impact.Indeterminate}, {false, false, true, true, impact.Unchanged},
	} {
		if got := classifyComparison(tc.failed, tc.delta, tc.complete, tc.aligned); got != tc.want {
			t.Fatalf("%+v: %s", tc, got)
		}
	}
}
func TestComparisonNormalizationPreservesCapturedAndRawIdentity(t *testing.T) {
	ids := func(string) map[string]string { return map[string]string{"ref-0": "renamed"} }
	object := environment.Object{ID: "ref-0", Kind: "dataset", Name: "captured"}
	got := normalizeFinding(object, ids, "", false).(map[string]any)
	if got["id"] != "ref-0" {
		t.Fatal("captured object identity normalized as a local reference")
	}
	target := json.RawMessage(`{"schema":{"properties":{"source_id":{"const":"one"},"schema_version":{"const":1},"id":{"const":"ref-0"},"captured_at":{"const":"then"}}}}`)
	var want any
	_ = json.Unmarshal(target, &want)
	if got := normalizeFinding(target, ids, "", false); !reflect.DeepEqual(got, want) {
		t.Fatal("raw schema facts erased or renamed")
	}
}
