package main

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
)

func TestCapabilitiesBindingsReturnCompleteCanonicalOwnedManifests(t *testing.T) {
	handle := spl_mapper_new()
	defer spl_mapper_free(handle)

	for _, language := range []string{"spl", "spl2"} {
		t.Run(language, func(t *testing.T) {
			options := analysis.CapabilityOptions{Language: language, Profile: "splunkd", Version: "current"}
			want, err := analysis.CapabilitiesFor(options)
			if err != nil {
				t.Fatal(err)
			}
			if want.ToolkitVersion == "" || len(want.Records) == 0 || len(want.Evidence) == 0 {
				t.Fatalf("incomplete canonical manifest: toolkit=%q records=%d evidence=%d", want.ToolkitVersion, len(want.Records), len(want.Evidence))
			}

			got, err := capabilitiesJSON([]byte(`{"language":"` + language + `","profile":"splunkd","version":"current"}`))
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("Go binding manifest mismatch\ngot:  %#v\nwant: %#v", got, want)
			}
			got.Records[0].Dimensions.Syntax.EvidenceIDs = append(got.Records[0].Dimensions.Syntax.EvidenceIDs, "mutated")
			got.Evidence[0].ID = "mutated"
			fresh, err := capabilitiesJSON([]byte(`{"language":"` + language + `","profile":"splunkd","version":"current"}`))
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(fresh, want) {
				t.Fatal("Go binding result mutation escaped into a later call")
			}

			callNative := func() *_Ctype_SPLResult {
				if language == "spl" {
					return spl_mapper_capabilities(handle)
				}
				return spl_mapper_capabilities_for(handle, nativeTestCString(t, `{"language":"spl2","profile":"splunkd","version":"current"}`))
			}
			result := callNative()
			if result == nil {
				t.Fatal("native capability export returned nil")
			}
			defer spl_result_free(result)
			if result.error != nil || result.result == nil {
				t.Fatalf("native capability export = error %q, result %q", nativeTestGoString(result.error), nativeTestGoString(result.result))
			}
			var native analysis.CapabilityManifest
			if err := json.Unmarshal([]byte(nativeTestGoString(result.result)), &native); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(native, want) {
				t.Fatalf("native manifest mismatch\ngot:  %#v\nwant: %#v", native, want)
			}
			native.Records[0].Dimensions.Syntax.Limitations = append(native.Records[0].Dimensions.Syntax.Limitations, "mutated")
			native.Evidence[0].ID = "mutated"
			freshResult := callNative()
			if freshResult == nil {
				t.Fatal("fresh native capability export returned nil")
			}
			defer spl_result_free(freshResult)
			if freshResult.error != nil || freshResult.result == nil {
				t.Fatalf("fresh native capability export = error %q, result %q", nativeTestGoString(freshResult.error), nativeTestGoString(freshResult.result))
			}
			var freshNative analysis.CapabilityManifest
			if err := json.Unmarshal([]byte(nativeTestGoString(freshResult.result)), &freshNative); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(freshNative, want) {
				t.Fatal("native binding result mutation escaped into a later call")
			}
		})
	}
}

func TestOwnedNativeJSONResultUsesCompactUTF8WithoutHTMLEscaping(t *testing.T) {
	handle := spl_mapper_new()
	defer spl_mapper_free(handle)

	result := ownedMapperJSONResult(handle, func() (any, error) {
		return map[string]string{"value": "café <tag> &"}, nil
	})
	if result == nil {
		t.Fatal("owned native JSON result returned nil")
	}
	defer spl_result_free(result)
	if result.error != nil || result.result == nil {
		t.Fatalf("owned native JSON result = error %q, result %q", nativeTestGoString(result.error), nativeTestGoString(result.result))
	}
	if got, want := nativeTestGoString(result.result), `{"value":"café <tag> &"}`; got != want {
		t.Fatalf("owned native JSON result = %q, want compact UTF-8 JSON %q", got, want)
	}
}

func TestCapabilityNativeEntryPointsRemainByteIdentical(t *testing.T) {
	handle := spl_mapper_new()
	defer spl_mapper_free(handle)

	legacy := spl_mapper_capabilities(handle)
	if legacy == nil {
		t.Fatal("legacy native capability export returned nil")
	}
	defer spl_result_free(legacy)
	selected := spl_mapper_capabilities_for(handle, nativeTestCString(t, `{}`))
	if selected == nil {
		t.Fatal("selected native capability export returned nil")
	}
	defer spl_result_free(selected)
	if legacy.error != nil || legacy.result == nil || selected.error != nil || selected.result == nil {
		t.Fatalf("native capability exports = legacy error %q result %q, selected error %q result %q",
			nativeTestGoString(legacy.error), nativeTestGoString(legacy.result), nativeTestGoString(selected.error), nativeTestGoString(selected.result))
	}
	if got, want := nativeTestGoString(legacy.result), nativeTestGoString(selected.result); got != want {
		t.Fatal("native capability entry points returned different JSON bytes")
	}
}
