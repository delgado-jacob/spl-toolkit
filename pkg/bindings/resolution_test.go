package main

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/compatibility"
	"github.com/delgado-jacob/spl-toolkit/pkg/environment"
	"github.com/delgado-jacob/spl-toolkit/pkg/resolution"
)

type resolutionSurfaceCase struct {
	ID            string                        `json:"id"`
	Document      analysis.QueryDocument        `json:"document"`
	Resolutions   []resolutionFixtureResolution `json:"resolutions"`
	MaxVariants   *int                          `json:"max_variants,omitempty"`
	Compatibility resolutionFixtureEvidence     `json:"compatibility"`
	Expected      resolutionFixtureExpectation  `json:"expected"`
}
type resolutionFixtureResolution struct {
	Marker string   `json:"marker"`
	Kind   string   `json:"kind"`
	Values []string `json:"values"`
}
type resolutionFixtureInputSelector struct {
	Kind     string                 `json:"kind"`
	Identity analysis.InputIdentity `json:"identity"`
}
type resolutionFixtureBinding struct {
	OriginalInput resolutionFixtureInputSelector `json:"original_input"`
	ResolvedValue string                         `json:"resolved_value,omitempty"`
	ObjectID      string                         `json:"object_id"`
	Expected      environment.ObjectIdentity     `json:"expected"`
	SchemaID      string                         `json:"schema_id,omitempty"`
}
type resolutionFixtureEvidence struct {
	Snapshot      environment.Snapshot       `json:"snapshot"`
	SchemaBundle  *environment.SchemaBundle  `json:"schema_bundle"`
	QueryScope    environment.CaptureScope   `json:"query_scope"`
	InputBindings []resolutionFixtureBinding `json:"input_bindings"`
}
type resolutionFixtureVariant struct {
	Selection     []string `json:"selection"`
	Outcome       string   `json:"outcome"`
	ResolvedQuery *string  `json:"resolved_query"`
}
type resolutionFixtureExpectation struct {
	Variants           []resolutionFixtureVariant `json:"variants"`
	Counts             map[string]int             `json:"counts"`
	SemanticAssertions []string                   `json:"semantic_assertions"`
	CombinationCount   int                        `json:"combination_count,omitempty"`
	RequestError       string                     `json:"request_error,omitempty"`
}

func resolutionSurfaceCases(t *testing.T) []resolutionSurfaceCase {
	t.Helper()
	data, err := os.ReadFile("../../testdata/resolution/cases.json")
	if err != nil {
		t.Fatal(err)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var cases []resolutionSurfaceCase
	if err := decoder.Decode(&cases); err != nil {
		t.Fatal(err)
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		t.Fatalf("trailing fixture data: %v", err)
	}
	return cases
}

// Canonical typed identity selects original inputs. No opaque input or occurrence
// hashes are fixture literals, and equal replacement names cannot alter selection.
func resolutionOriginalInputID(t *testing.T, result *analysis.Result, selector resolutionFixtureInputSelector) string {
	t.Helper()
	var id string
	for _, input := range result.Inputs {
		if input.Kind == selector.Kind && input.Identity == selector.Identity {
			if id != "" {
				t.Fatalf("ambiguous fixture selector: %+v", selector)
			}
			id = input.ID
		}
	}
	if id == "" {
		t.Fatalf("original input not found: %+v in %+v", selector, result.Inputs)
	}
	return id
}

func resolutionSurfaceRequest(t *testing.T, c resolutionSurfaceCase) resolution.Request {
	t.Helper()
	result, err := analysis.Analyze(c.Document)
	if err != nil {
		t.Fatal(err)
	}
	r := resolution.Request{SchemaVersion: 1, Document: c.Document, Resolutions: []resolution.Resolution{}, Compatibility: resolution.CompatibilityInputs{Snapshot: c.Compatibility.Snapshot, SchemaBundle: c.Compatibility.SchemaBundle, QueryScope: c.Compatibility.QueryScope, InputBindings: []compatibility.ResolutionBinding{}}}
	for _, v := range c.Resolutions {
		r.Resolutions = append(r.Resolutions, resolution.Resolution{Placeholder: v.Marker, Kind: v.Kind, Values: v.Values})
	}
	if c.MaxVariants != nil {
		maximum := uint64(*c.MaxVariants)
		r.MaxVariants = &maximum
	}
	for _, b := range c.Compatibility.InputBindings {
		binding := compatibility.ResolutionBinding{OriginalInputID: resolutionOriginalInputID(t, result, b.OriginalInput), ObjectID: b.ObjectID, Expected: b.Expected, SchemaID: b.SchemaID}
		if b.ResolvedValue != "" {
			value := b.ResolvedValue
			binding.ResolvedValue = &value
		}
		r.Compatibility.InputBindings = append(r.Compatibility.InputBindings, binding)
	}
	return r
}

func TestResolutionNativeOwnedReports(t *testing.T) {
	handle := spl_mapper_new()
	defer spl_mapper_free(handle)
	for _, c := range resolutionSurfaceCases(t) {
		if c.Expected.RequestError != "" {
			continue
		}
		t.Run(c.ID, func(t *testing.T) {
			request := resolutionSurfaceRequest(t, c)
			raw, err := json.Marshal(request)
			if err != nil {
				t.Fatal(err)
			}
			report, err := resolution.ResolveJSON(raw)
			if err != nil {
				t.Fatal(err)
			}
			if report.Counts.Verified != uint64(c.Expected.Counts["verified"]) || report.Counts.Failed != uint64(c.Expected.Counts["failed"]) || report.Counts.Incomplete != uint64(c.Expected.Counts["incomplete"]) {
				t.Fatal("authored counts mismatch")
			}
			for i, expected := range c.Expected.Variants {
				got := report.Variants[i]
				if got.Outcome != expected.Outcome || !reflect.DeepEqual(got.ResolvedQuery, expected.ResolvedQuery) {
					t.Fatalf("authored variant %d mismatch", i)
				}
			}
			want, err := marshalNativeJSON(report)
			if err != nil {
				t.Fatal(err)
			}
			first := spl_mapper_resolve(handle, nativeTestCStringBytes(t, raw))
			if first == nil {
				t.Fatal("nil owned result")
			}
			defer spl_result_free(first)
			second := spl_mapper_resolve(handle, nativeTestCString(t, `{`))
			if second == nil {
				t.Fatal("nil owned error")
			}
			spl_result_free(second)
			if first.error != nil || first.result == nil || nativeTestGoString(first.result) != string(want) {
				t.Fatalf("error=%q result=%q", nativeTestGoString(first.error), nativeTestGoString(first.result))
			}
			if string(raw) != mustNativeTestJSON(t, request) {
				t.Fatal("request mutated")
			}
		})
	}
}

func TestResolutionNativeOwnedErrors(t *testing.T) {
	handle := spl_mapper_new()
	defer spl_mapper_free(handle)
	raws := [][]byte{nil, []byte(`{`), []byte(`null`), []byte(`{"schema_version":1,"unknown":true}`), {0xff}}
	for _, c := range resolutionSurfaceCases(t) {
		if c.Expected.RequestError == "" {
			continue
		}
		raw, _ := json.Marshal(resolutionSurfaceRequest(t, c))
		_, err := resolution.ResolveJSON(raw)
		detail, ok := resolution.RequestErrorDetails(err)
		if !ok || detail.Code != c.Expected.RequestError || detail.TotalCombinations != strconv.Itoa(c.Expected.CombinationCount) {
			t.Fatalf("authored rejection mismatch: %v", err)
		}
		raws = append(raws, raw)
	}
	valid := resolutionSurfaceRequest(t, resolutionSurfaceCases(t)[0])
	valid.SchemaVersion = 2
	invalidVersion, _ := json.Marshal(valid)
	raws = append(raws, invalidVersion)
	valid.SchemaVersion = 1
	valid.Resolutions[0].Kind = "unknown"
	invalidKind, _ := json.Marshal(valid)
	raws = append(raws, invalidKind)
	for _, raw := range raws {
		_, err := resolution.ResolveJSON(raw)
		detail, ok := resolution.RequestErrorDetails(err)
		if !ok {
			t.Fatal(err)
		}
		want, _ := json.Marshal(detail)
		result := spl_mapper_resolve(handle, nativeTestCStringBytes(t, raw))
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
	result := spl_mapper_resolve(handle, nil)
	if result == nil {
		t.Fatal("nil null-request result")
	}
	defer spl_result_free(result)
	if result.error == nil || result.result != nil {
		t.Fatal("null request admitted")
	}
}

func TestResolutionNativeClosedAndInvalidHandle(t *testing.T) {
	handle := spl_mapper_new()
	spl_mapper_free(handle)
	for _, id := range []_Ctype_int{handle, -1, 2147483647} {
		result := spl_mapper_resolve(id, nil)
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

func TestResolutionNativeUnicodeAudit(t *testing.T) {
	c := resolutionSurfaceCases(t)[0]
	c.Document.Text = "from $events | eval label=\"雪é\" | fields id"
	c.Document.SourceID = "résolution:雪"
	request := resolutionSurfaceRequest(t, c)
	raw, _ := json.Marshal(request)
	report, err := resolution.ResolveJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	want, _ := marshalNativeJSON(report)
	handle := spl_mapper_new()
	defer spl_mapper_free(handle)
	result := spl_mapper_resolve(handle, nativeTestCStringBytes(t, raw))
	if result == nil {
		t.Fatal("nil owned result")
	}
	defer spl_result_free(result)
	if result.error != nil || nativeTestGoString(result.result) != string(want) {
		t.Fatal("Unicode report differs")
	}
	if !strings.Contains(nativeTestGoString(result.result), "雪é") {
		t.Fatal("Unicode audit missing")
	}
	for _, variant := range report.Variants {
		for _, change := range variant.Changes {
			loc := change.OriginalLocation
			if request.Document.Text[loc.Start.Offset:loc.End.Offset] != change.Before {
				t.Fatal("original byte range differs")
			}
			if change.CandidateLocation != nil {
				loc = *change.CandidateLocation
				if (*variant.CandidateText)[loc.Start.Offset:loc.End.Offset] != change.After {
					t.Fatal("candidate byte range differs")
				}
			}
		}
	}
}
