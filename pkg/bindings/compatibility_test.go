package main

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/compatibility"
	"github.com/delgado-jacob/spl-toolkit/pkg/environment"
)

type compatibilitySurfaceCase struct {
	Name         string                    `json:"name"`
	Document     analysis.QueryDocument    `json:"document"`
	Snapshot     environment.Snapshot      `json:"snapshot"`
	SchemaBundle *environment.SchemaBundle `json:"schema_bundle"`
	QueryScope   environment.CaptureScope  `json:"query_scope"`
	Bindings     map[string]struct {
		ObjectID string                     `json:"object_id"`
		Expected environment.ObjectIdentity `json:"expected"`
		SchemaID string                     `json:"schema_id"`
	} `json:"bindings"`
	Expected struct {
		RequestError  bool `json:"request_error"`
		Compatibility struct {
			Outcome string `json:"outcome"`
		} `json:"compatibility"`
		Correlation struct {
			Outcome string `json:"outcome"`
		} `json:"correlation"`
	} `json:"expected"`
}

func compatibilitySurfaceCases(t *testing.T) []compatibilitySurfaceCase {
	t.Helper()
	raw, err := os.ReadFile("../../testdata/compatibility/cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []compatibilitySurfaceCase
	if err := json.Unmarshal(raw, &cases); err != nil {
		t.Fatal(err)
	}
	return cases
}

func compatibilitySurfaceRequest(t *testing.T, c compatibilitySurfaceCase) []byte {
	t.Helper()
	requirements, err := analysis.Requirements(c.Document)
	if err != nil {
		t.Fatal(err)
	}
	request := compatibility.Request{SchemaVersion: 1, Requirements: *requirements, Snapshot: c.Snapshot, SchemaBundle: c.SchemaBundle, QueryScope: c.QueryScope, InputBindings: []compatibility.InputBinding{}}
	for _, input := range requirements.Inputs {
		binding, exists := c.Bindings[input.Kind+":"+input.Name]
		if !exists {
			continue
		}
		if binding.Expected.Kind == "" {
			for _, object := range c.Snapshot.Objects {
				if object.ID == binding.ObjectID {
					binding.Expected = environment.ObjectIdentity{Kind: object.Kind, Name: object.Name, Namespace: object.Namespace, App: object.App, Owner: object.Owner}
				}
			}
		}
		request.InputBindings = append(request.InputBindings, compatibility.InputBinding{InputID: input.ID, ObjectID: binding.ObjectID, Expected: binding.Expected, SchemaID: binding.SchemaID})
	}
	raw, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func TestCompatibilityNativeOwnedReports(t *testing.T) {
	handle := spl_mapper_new()
	defer spl_mapper_free(handle)
	for _, c := range compatibilitySurfaceCases(t) {
		if c.Expected.RequestError {
			continue
		}
		t.Run(c.Name, func(t *testing.T) {
			raw := compatibilitySurfaceRequest(t, c)
			report, err := compatibility.CheckJSON(raw)
			if err != nil {
				t.Fatal(err)
			}
			if report.Outcome != c.Expected.Compatibility.Outcome {
				t.Fatal("authored outcome mismatch")
			}
			want, err := marshalNativeJSON(report)
			if err != nil {
				t.Fatal(err)
			}
			result := spl_mapper_check_compatibility(handle, nativeTestCStringBytes(t, raw))
			if result == nil {
				t.Fatal("nil owned result")
			}
			defer spl_result_free(result)
			if result.error != nil || result.result == nil || nativeTestGoString(result.result) != string(want) {
				t.Fatalf("error=%q result=%q", nativeTestGoString(result.error), nativeTestGoString(result.result))
			}
		})
	}
}

func TestCompatibilityNativeOwnedErrors(t *testing.T) {
	handle := spl_mapper_new()
	defer spl_mapper_free(handle)
	for _, raw := range [][]byte{nil, []byte(`{`), []byte(`{"schema_version":1,"unknown":true}`), {0xff}} {
		_, err := compatibility.CheckJSON(raw)
		detail, ok := compatibility.RequestErrorDetails(err)
		if !ok {
			t.Fatal(err)
		}
		want, _ := json.Marshal(detail)
		result := spl_mapper_check_compatibility(handle, nativeTestCStringBytes(t, raw))
		if result == nil {
			t.Fatal("nil owned error")
		}
		func() {
			defer spl_result_free(result)
			if result.error == nil || result.result != nil || nativeTestGoString(result.error) != string(want) {
				t.Fatalf("error=%q result=%q want=%s", nativeTestGoString(result.error), nativeTestGoString(result.result), want)
			}
		}()
	}
}

func TestCompatibilityNativeClosedAndInvalidHandle(t *testing.T) {
	handle := spl_mapper_new()
	spl_mapper_free(handle)
	for _, id := range []_Ctype_int{handle, -1, 2147483647} {
		result := spl_mapper_check_compatibility(id, nativeTestCString(t, `{}`))
		if result == nil {
			t.Fatal("nil owned error")
		}
		func() {
			defer spl_result_free(result)
			if result.error == nil || result.result != nil || nativeTestGoString(result.error) != "Mapper not found" {
				t.Fatal("invalid mapper was admitted")
			}
		}()
	}
	spl_result_free(nil)
}
