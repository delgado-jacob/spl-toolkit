package main

import (
	"encoding/json"
	"strings"
	"testing"
	"unsafe"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
)

func TestRequirementsExportReturnsOwnedJSON(t *testing.T) {
	document := analysis.QueryDocument{Text: "search host=web"}
	want, err := analysis.Requirements(document)
	if err != nil {
		t.Fatalf("Requirements() error = %v", err)
	}

	handle := spl_mapper_new()
	defer spl_mapper_free(handle)
	result := spl_mapper_requirements_query(handle, nativeTestCString(t, `{"text":"search host=web"}`))
	if result == nil {
		t.Fatal("requirements export returned nil")
	}
	defer spl_result_free(result)
	if result.error != nil || result.result == nil {
		t.Fatalf("requirements export = error %q, result %q", nativeTestGoString(result.error), nativeTestGoString(result.result))
	}

	var got analysis.RequirementSet
	if err := json.Unmarshal([]byte(nativeTestGoString(result.result)), &got); err != nil {
		t.Fatalf("decode native result: %v", err)
	}
	if encodedGot, encodedWant := mustNativeTestJSON(t, got), mustNativeTestJSON(t, want); encodedGot != encodedWant {
		t.Fatalf("native requirements mismatch\ngot:  %s\nwant: %s", encodedGot, encodedWant)
	}
}

func TestRequirementsExportRejectsBadDocumentsAndClosedHandles(t *testing.T) {
	handle := spl_mapper_new()
	badDocuments := [][]byte{
		[]byte("{"),
		[]byte(`[]`),
		[]byte(`{"text":"x"},{"text":"y"}`),
		[]byte(`{"text":"x","text":"y"}`),
		[]byte(`{"text":"x","unknown":true}`),
		{'{', '"', 't', 'e', 'x', 't', '"', ':', '"', 0xff, '"', '}'},
	}
	for _, document := range badDocuments {
		result := spl_mapper_requirements_query(handle, nativeTestCStringBytes(t, document))
		if result == nil || result.error == nil || result.result != nil {
			t.Fatalf("document %q = error %q, result %q; want owned error only", document, nativeTestGoString(result.error), nativeTestGoString(result.result))
		}
		spl_result_free(result)
	}
	spl_mapper_free(handle)

	result := spl_mapper_requirements_query(handle, nativeTestCString(t, `{"text":"search host=web"}`))
	if result == nil || nativeTestGoString(result.error) != "Mapper not found" || result.result != nil {
		t.Fatalf("closed handle = error %q, result %q; want owned Mapper not found error", nativeTestGoString(result.error), nativeTestGoString(result.result))
	}
	spl_result_free(result)
}

func TestRequirementsExportsPreserveResourceLimitSuccess(t *testing.T) {
	for _, language := range []string{"spl", "spl2"} {
		t.Run(language, func(t *testing.T) {
			document := analysis.QueryDocument{Text: strings.Repeat("a ", 2048) + "a", Language: language, SourceID: "dense." + language}
			wantAnalysis, err := analysis.Analyze(document)
			if err != nil {
				t.Fatalf("Analyze() error = %v", err)
			}
			wantRequirements, err := analysis.Requirements(document)
			if err != nil {
				t.Fatalf("Requirements() error = %v", err)
			}
			if wantAnalysis.Status != "incomplete" || len(wantAnalysis.Diagnostics) != 1 || wantAnalysis.Diagnostics[0].Code != "SPL_ANALYSIS_RESOURCE_LIMIT" {
				t.Fatalf("fixture did not produce resource-limit analysis: %+v", wantAnalysis.Diagnostics)
			}

			payload := mustNativeTestJSON(t, document)
			handle := spl_mapper_new()
			defer spl_mapper_free(handle)
			t.Run("analysis", func(t *testing.T) {
				result := spl_mapper_analyze_query(handle, nativeTestCString(t, payload))
				if result == nil {
					t.Fatal("native export returned nil")
				}
				defer spl_result_free(result)
				if result.error != nil || result.result == nil {
					t.Fatalf("native export = error %q, result %q", nativeTestGoString(result.error), nativeTestGoString(result.result))
				}
				if got, want := nativeTestGoString(result.result), mustNativeTestJSON(t, wantAnalysis); got != want {
					t.Fatal("native analysis resource result mismatch")
				}
			})
			t.Run("requirements", func(t *testing.T) {
				result := spl_mapper_requirements_query(handle, nativeTestCString(t, payload))
				if result == nil {
					t.Fatal("native export returned nil")
				}
				defer spl_result_free(result)
				if result.error != nil || result.result == nil {
					t.Fatalf("native export = error %q, result %q", nativeTestGoString(result.error), nativeTestGoString(result.result))
				}
				if got, want := nativeTestGoString(result.result), mustNativeTestJSON(t, wantRequirements); got != want {
					t.Fatal("native requirements resource result mismatch")
				}
			})
		})
	}
}

func nativeTestCString(t *testing.T, value string) *_Ctype_char {
	t.Helper()
	return nativeTestCStringBytes(t, []byte(value))
}

func nativeTestCStringBytes(t *testing.T, value []byte) *_Ctype_char {
	t.Helper()
	if strings.IndexByte(string(value), 0) >= 0 {
		t.Fatal("native test C string contains NUL")
	}
	value = append(value, 0)
	return (*_Ctype_char)(unsafe.Pointer(&value[0]))
}

func nativeTestGoString[T any](pointer *T) string {
	if pointer == nil {
		return ""
	}
	base := unsafe.Pointer(pointer)
	bytes := make([]byte, 0, 256)
	for index := uintptr(0); ; index++ {
		value := *(*byte)(unsafe.Add(base, index))
		if value == 0 {
			return string(bytes)
		}
		bytes = append(bytes, value)
	}
}

func mustNativeTestJSON(t *testing.T, value any) string {
	t.Helper()
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	return string(encoded)
}
