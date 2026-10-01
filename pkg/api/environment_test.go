package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/delgado-jacob/spl-toolkit/pkg/environment"
)

func environmentREST(t *testing.T, body []byte, contentType string, unknownLength bool) *httptest.ResponseRecorder {
	t.Helper()
	r := httptest.NewRequest(http.MethodPost, "/api/v1/environment/validate", bytes.NewReader(body))
	r.Header.Set("Content-Type", contentType)
	if unknownLength {
		r.ContentLength = -1
	}
	w := httptest.NewRecorder()
	NewServer().Handler().ServeHTTP(w, r)
	return w
}

func TestEnvironmentRESTFixtureReports(t *testing.T) {
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
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			request, err := json.Marshal(struct {
				SchemaVersion int             `json:"schema_version"`
				Snapshot      json.RawMessage `json:"snapshot"`
				SchemaBundle  json.RawMessage `json:"schema_bundle"`
			}{1, tc.Snapshot, tc.SchemaBundle})
			if err != nil {
				t.Fatal(err)
			}
			want, err := environment.ValidateJSON(request)
			if err != nil || want.Status != tc.Status {
				t.Fatalf("reference: %v %#v", err, want)
			}
			wantJSON, err := json.Marshal(want)
			if err != nil {
				t.Fatal(err)
			}
			w := environmentREST(t, request, "application/json; charset=utf-8", false)
			wantCode := http.StatusOK
			if tc.Status == "invalid" {
				wantCode = http.StatusBadRequest
			}
			if w.Code != wantCode || !bytes.Equal(bytes.TrimSpace(w.Body.Bytes()), wantJSON) {
				t.Fatalf("status=%d body=%s want=%s", w.Code, w.Body, wantJSON)
			}
		})
	}
}

func TestEnvironmentRESTMalformedInlineAttribution(t *testing.T) {
	for _, raw := range [][]byte{
		[]byte(`{"schema_version":1,"snapshot":{"schema_version":1,},"schema_bundle":{}}`),
		[]byte(`{"schema_version":1,"snapshot":{},"schema_bundle":{"schema_version":1,"schemas":[}`),
		[]byte(`{"schema_version":1,"snapshot":{},"extra":true}`),
	} {
		want, err := environment.ValidateJSON(raw)
		if err != nil || want.Status != "invalid" || len(want.Diagnostics) == 0 || want.Diagnostics[0].Artifact != "request" {
			t.Fatalf("reference: %v %#v", err, want)
		}
		wantJSON, err := json.Marshal(want)
		if err != nil {
			t.Fatal(err)
		}
		w := environmentREST(t, raw, "application/json", false)
		if w.Code != http.StatusBadRequest || !bytes.Equal(bytes.TrimSpace(w.Body.Bytes()), wantJSON) {
			t.Fatalf("status=%d body=%s want=%s", w.Code, w.Body, wantJSON)
		}
	}
}

func TestEnvironmentRESTContentTypeAndBodyLimit(t *testing.T) {
	for _, contentType := range []string{"", "text/plain", "application/json; malformed"} {
		w := environmentREST(t, []byte(`{"schema_version":1}`), contentType, false)
		if w.Code != http.StatusBadRequest || !strings.Contains(w.Body.String(), `"error":true`) {
			t.Fatalf("content type %q: %d %s", contentType, w.Code, w.Body)
		}
	}
	base := []byte(`{"schema_version":1,"schema_bundle":{"schema_version":1,"bundle_id":"bundle-a","provenance":{"source_kind":"fixture","source_id":"schema-source","observed_at":"2026-10-01T12:00:00Z"},"schemas":[],"bindings":[]}}`)
	exact := append(base, bytes.Repeat([]byte(" "), (8<<20)-len(base))...)
	if w := environmentREST(t, exact, "application/json", false); w.Code != http.StatusOK {
		t.Fatalf("exact limit: %d %s", w.Code, w.Body)
	}
	for _, unknown := range []bool{false, true} {
		w := environmentREST(t, append(exact, ' '), "application/json", unknown)
		if w.Code != http.StatusBadRequest || !strings.Contains(w.Body.String(), "request body exceeds 8 MiB limit") || !strings.Contains(w.Body.String(), `"error":true`) {
			t.Fatalf("over limit unknown=%v: %d %s", unknown, w.Code, w.Body)
		}
	}
}
