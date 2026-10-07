package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
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

func resolutionREST(body []byte, contentType string, unknown bool) *httptest.ResponseRecorder {
	r := httptest.NewRequest(http.MethodPost, "/api/v1/query/resolve", bytes.NewReader(body))
	r.Header.Set("Content-Type", contentType)
	if unknown {
		r.ContentLength = -1
	}
	w := httptest.NewRecorder()
	NewServer().Handler().ServeHTTP(w, r)
	return w
}
func TestResolutionRESTReports(t *testing.T) {
	server := httptest.NewServer(NewServer().Handler())
	defer server.Close()
	for _, c := range resolutionSurfaceCases(t) {
		if c.Expected.RequestError != "" {
			continue
		}
		t.Run(c.ID, func(t *testing.T) {
			raw, _ := json.Marshal(resolutionSurfaceRequest(t, c))
			report, err := resolution.ResolveJSON(raw)
			if err != nil {
				t.Fatal(err)
			}
			want, _ := json.Marshal(report)
			response, err := server.Client().Post(server.URL+"/api/v1/query/resolve", "application/json; charset=utf-8", bytes.NewReader(raw))
			if err != nil {
				t.Fatal(err)
			}
			got, readErr := io.ReadAll(response.Body)
			closeErr := response.Body.Close()
			if readErr != nil || closeErr != nil {
				t.Fatalf("body: %v %v", readErr, closeErr)
			}
			if response.StatusCode != 200 || !bytes.Equal(bytes.TrimSpace(got), want) {
				t.Fatalf("%d %s", response.StatusCode, got)
			}
		})
	}
}
func TestResolutionRESTStructuredErrors(t *testing.T) {
	raws := [][]byte{nil, []byte(`{`), []byte(`null`), []byte(`{"schema_version":1,"unknown":true}`), {0xff}}
	for _, c := range resolutionSurfaceCases(t) {
		if c.Expected.RequestError != "" {
			raw, _ := json.Marshal(resolutionSurfaceRequest(t, c))
			raws = append(raws, raw)
		}
	}
	valid, _ := json.Marshal(resolutionSurfaceRequest(t, resolutionSurfaceCases(t)[0]))
	var nullRequest map[string]any
	if err := json.Unmarshal(valid, &nullRequest); err != nil {
		t.Fatal(err)
	}
	nullRequest["resolutions"] = nil
	nullRaw, _ := json.Marshal(nullRequest)
	raws = append(raws, nullRaw)
	for _, raw := range raws {
		_, err := resolution.ResolveJSON(raw)
		detail, ok := resolution.RequestErrorDetails(err)
		if !ok {
			t.Fatal(err)
		}
		want, _ := json.Marshal(detail)
		w := resolutionREST(raw, "application/json", false)
		if w.Code != 400 || !bytes.Equal(bytes.TrimSpace(w.Body.Bytes()), want) {
			t.Fatalf("%d %s want %s", w.Code, w.Body, want)
		}
	}
}
func TestResolutionRESTBoundaries(t *testing.T) {
	raw, _ := json.Marshal(resolutionSurfaceRequest(t, resolutionSurfaceCases(t)[0]))
	for _, ct := range []string{"", "text/plain", "application/json; malformed"} {
		w := resolutionREST(raw, ct, false)
		if w.Code != 400 || !strings.Contains(w.Body.String(), `"error":true`) {
			t.Fatalf("%q: %d %s", ct, w.Code, w.Body)
		}
	}
	exact := append(raw, bytes.Repeat([]byte(" "), (8<<20)-len(raw))...)
	for _, unknown := range []bool{false, true} {
		if w := resolutionREST(exact, "application/json", unknown); w.Code != 200 {
			t.Fatalf("exact: %d %s", w.Code, w.Body)
		}
		if w := resolutionREST(append(exact, ' '), "application/json", unknown); w.Code != 400 || !strings.Contains(w.Body.String(), "request body exceeds 8 MiB limit") {
			t.Fatalf("over: %d %s", w.Code, w.Body)
		}
	}
	// Exercise real transfer framing and drain responses before server shutdown.
	server := httptest.NewServer(NewServer().Handler())
	defer server.Close()
	for _, chunked := range []bool{false, true} {
		request, err := http.NewRequest(http.MethodPost, server.URL+"/api/v1/query/resolve", bytes.NewReader(append(exact, ' ')))
		if err != nil {
			t.Fatal(err)
		}
		request.Header.Set("Content-Type", "application/json")
		if chunked {
			request.ContentLength = -1
		}
		response, err := server.Client().Do(request)
		if err != nil {
			t.Fatal(err)
		}
		body, readErr := io.ReadAll(response.Body)
		closeErr := response.Body.Close()
		if readErr != nil || closeErr != nil {
			t.Fatalf("body: %v %v", readErr, closeErr)
		}
		if response.StatusCode != 400 || !strings.Contains(string(body), "request body exceeds 8 MiB limit") {
			t.Fatalf("chunked=%v: %d %s", chunked, response.StatusCode, body)
		}
	}
}
func TestResolutionRESTInternalError(t *testing.T) {
	w := httptest.NewRecorder()
	NewServer().writeResolutionError(w, errors.New("internal test failure"))
	if w.Code != 500 || !strings.Contains(w.Body.String(), `"error":true`) {
		t.Fatalf("%d %s", w.Code, w.Body)
	}
}
