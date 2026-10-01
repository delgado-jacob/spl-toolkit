package main

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/delgado-jacob/spl-toolkit/pkg/environment"
)

func TestEnvironmentNativeOwnedReportParity(t *testing.T) {
	raw, err := os.ReadFile("../../testdata/environment/cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Name         string          `json:"name"`
		Status       string          `json:"status"`
		Snapshot     json.RawMessage `json:"snapshot"`
		SchemaBundle json.RawMessage `json:"schema_bundle"`
	}
	if err := json.Unmarshal(raw, &cases); err != nil {
		t.Fatal(err)
	}
	handle := spl_mapper_new()
	defer spl_mapper_free(handle)
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			request, err := json.Marshal(map[string]any{"schema_version": 1, "snapshot": tc.Snapshot, "schema_bundle": tc.SchemaBundle})
			if err != nil {
				t.Fatal(err)
			}
			want, err := environment.ValidateJSON(request)
			if err != nil || want.Status != tc.Status {
				t.Fatalf("reference: %v %#v", err, want)
			}
			result := spl_mapper_validate_environment(handle, nativeTestCStringBytes(t, request))
			if result == nil {
				t.Fatal("native result is nil")
			}
			defer spl_result_free(result)
			if result.error != nil || result.result == nil {
				t.Fatalf("native error=%q result=%q", nativeTestGoString(result.error), nativeTestGoString(result.result))
			}
			wantJSON, err := json.Marshal(want)
			if err != nil {
				t.Fatal(err)
			}
			var got, expected any
			if err := json.Unmarshal([]byte(nativeTestGoString(result.result)), &got); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal(wantJSON, &expected); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, expected) {
				t.Fatalf("native report differs\ngot: %s\nwant: %s", nativeTestGoString(result.result), wantJSON)
			}
		})
	}
}

func TestEnvironmentNativeMalformedRequestAndClosedMapper(t *testing.T) {
	handle := spl_mapper_new()
	raw := []byte(`{"schema_version":1,"snapshot":{"schema_version":1,},"schema_bundle":{}}`)
	want, err := environment.ValidateJSON(raw)
	if err != nil || want.Status != "invalid" || len(want.Diagnostics) == 0 || want.Diagnostics[0].Artifact != "request" {
		t.Fatalf("reference: %v %#v", err, want)
	}
	result := spl_mapper_validate_environment(handle, nativeTestCStringBytes(t, raw))
	if result == nil || result.error != nil || result.result == nil {
		t.Fatal("malformed inline JSON did not return an owned report")
	}
	wantJSON, err := json.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	var got, expected any
	if err := json.Unmarshal([]byte(nativeTestGoString(result.result)), &got); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(wantJSON, &expected); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("native report differs: %s %s", nativeTestGoString(result.result), wantJSON)
	}
	spl_result_free(result)
	spl_mapper_free(handle)
	closed := spl_mapper_validate_environment(handle, nativeTestCString(t, `{"schema_version":1}`))
	if closed == nil || closed.error == nil || closed.result != nil || nativeTestGoString(closed.error) != "Mapper not found" {
		t.Fatal("closed mapper did not return an owned error")
	}
	spl_result_free(closed)
}
