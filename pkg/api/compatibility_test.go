package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
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

func compatibilityREST(body []byte, contentType string, unknown bool) *httptest.ResponseRecorder {
	r := httptest.NewRequest(http.MethodPost, "/api/v1/query/compatibility", bytes.NewReader(body))
	r.Header.Set("Content-Type", contentType)
	if unknown {
		r.ContentLength = -1
	}
	w := httptest.NewRecorder()
	NewServer().Handler().ServeHTTP(w, r)
	return w
}

func TestCompatibilityRESTReports(t *testing.T) {
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
			want, _ := json.Marshal(report)
			w := compatibilityREST(raw, "application/json; charset=utf-8", false)
			if w.Code != http.StatusOK || !bytes.Equal(bytes.TrimSpace(w.Body.Bytes()), want) {
				t.Fatalf("%d %s", w.Code, w.Body)
			}
		})
	}
}

func TestCompatibilityRESTStructuredErrors(t *testing.T) {
	for _, raw := range [][]byte{nil, []byte(`{`), []byte(`{"schema_version":1,"unknown":true}`), {0xff}} {
		_, err := compatibility.CheckJSON(raw)
		detail, ok := compatibility.RequestErrorDetails(err)
		if !ok {
			t.Fatal(err)
		}
		want, _ := json.Marshal(detail)
		w := compatibilityREST(raw, "application/json", false)
		if w.Code != 400 || !bytes.Equal(bytes.TrimSpace(w.Body.Bytes()), want) {
			t.Fatalf("%d %s want %s", w.Code, w.Body, want)
		}
	}
}

func TestCompatibilityRESTBoundaries(t *testing.T) {
	raw := compatibilitySurfaceRequest(t, compatibilitySurfaceCases(t)[0])
	for _, ct := range []string{"", "text/plain", "application/json; malformed"} {
		if w := compatibilityREST(raw, ct, false); w.Code != 400 || !strings.Contains(w.Body.String(), `"error":true`) {
			t.Fatalf("%q: %d %s", ct, w.Code, w.Body)
		}
	}
	exact := append(raw, bytes.Repeat([]byte(" "), (8<<20)-len(raw))...)
	for _, unknown := range []bool{false, true} {
		if w := compatibilityREST(exact, "application/json", unknown); w.Code != 200 {
			t.Fatalf("exact limit: %d %s", w.Code, w.Body)
		}
		if w := compatibilityREST(append(exact, ' '), "application/json", unknown); w.Code != 400 || !strings.Contains(w.Body.String(), "request body exceeds 8 MiB limit") {
			t.Fatalf("over limit: %d %s", w.Code, w.Body)
		}
	}
}

func TestCompatibilityRESTInternalError(t *testing.T) {
	w := httptest.NewRecorder()
	NewServer().writeCompatibilityError(w, errors.New("internal test failure"))
	if w.Code != 500 || !strings.Contains(w.Body.String(), `"error":true`) {
		t.Fatalf("%d %s", w.Code, w.Body)
	}
}
