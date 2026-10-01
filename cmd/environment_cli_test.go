package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/delgado-jacob/spl-toolkit/pkg/api"
	"github.com/delgado-jacob/spl-toolkit/pkg/environment"
)

type environmentCase struct {
	Name         string          `json:"name"`
	Status       string          `json:"status"`
	Snapshot     json.RawMessage `json:"snapshot"`
	SchemaBundle json.RawMessage `json:"schema_bundle"`
}

func environmentCases(t *testing.T) []environmentCase {
	t.Helper()
	raw, err := os.ReadFile("../testdata/environment/cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []environmentCase
	if err := json.Unmarshal(raw, &cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) != 3 {
		t.Fatalf("want three environment cases, got %d", len(cases))
	}
	return cases
}

func TestEnvironmentCLIAndHTTPParity(t *testing.T) {
	handler := api.NewServer().Handler()
	for _, tc := range environmentCases(t) {
		t.Run(tc.Name, func(t *testing.T) {
			want, err := environment.ValidateArtifacts(tc.Snapshot, tc.SchemaBundle)
			if err != nil || want.Status != tc.Status {
				t.Fatalf("fixture: %v %#v", err, want)
			}
			wantJSON, err := json.Marshal(want)
			if err != nil {
				t.Fatal(err)
			}
			dir := t.TempDir()
			snapshotPath := filepath.Join(dir, "snapshot.json")
			bundlePath := filepath.Join(dir, "schemas.json")
			if err := os.WriteFile(snapshotPath, tc.Snapshot, 0600); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(bundlePath, tc.SchemaBundle, 0600); err != nil {
				t.Fatal(err)
			}
			var out, stderr bytes.Buffer
			code := runCLIWithInput([]string{"environment", "validate", "--snapshot", snapshotPath, "--schemas", bundlePath, "--format", "json"}, strings.NewReader(""), &out, &stderr)
			wantCode := map[string]int{"valid": 0, "invalid": 1, "partial": 3}[tc.Status]
			if code != wantCode || stderr.Len() != 0 || !bytes.Equal(bytes.TrimSpace(out.Bytes()), wantJSON) {
				t.Fatalf("CLI code=%d stderr=%s report=%s want=%s", code, &stderr, &out, wantJSON)
			}
			if bytes.Contains(out.Bytes(), []byte(dir)) {
				t.Fatal("local path leaked into canonical report")
			}
			requestJSON, err := json.Marshal(struct {
				SchemaVersion int             `json:"schema_version"`
				Snapshot      json.RawMessage `json:"snapshot"`
				SchemaBundle  json.RawMessage `json:"schema_bundle"`
			}{1, tc.Snapshot, tc.SchemaBundle})
			if err != nil {
				t.Fatal(err)
			}
			r := httptest.NewRequest(http.MethodPost, "/api/v1/environment/validate", bytes.NewReader(requestJSON))
			r.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, r)
			wantHTTP := 200
			if tc.Status == "invalid" {
				wantHTTP = 400
			}
			if w.Code != wantHTTP || !bytes.Equal(bytes.TrimSpace(w.Body.Bytes()), wantJSON) {
				t.Fatalf("HTTP status=%d report=%s want=%s", w.Code, w.Body.String(), wantJSON)
			}
		})
	}
}

func TestEnvironmentCLISchemaOnlyTextAndOutput(t *testing.T) {
	tc := environmentCases(t)[0]
	path := filepath.Join(t.TempDir(), "schemas.json")
	if err := os.WriteFile(path, tc.SchemaBundle, 0600); err != nil {
		t.Fatal(err)
	}
	want, err := environment.ValidateArtifacts(nil, tc.SchemaBundle)
	if err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(t.TempDir(), "report.txt")
	var stdout, stderr bytes.Buffer
	code := runCLIWithInput([]string{"environment", "validate", "--schemas", path, "--output", output}, strings.NewReader(""), &stdout, &stderr)
	got, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	if code != 0 || stdout.Len() != 0 || stderr.Len() != 0 || string(got) != environment.FormatReport(want) {
		t.Fatalf("code=%d stdout=%s stderr=%s report=%s", code, &stdout, &stderr, got)
	}
}

func TestEnvironmentCLISnapshotOnly(t *testing.T) {
	tc := environmentCases(t)[0]
	path := filepath.Join(t.TempDir(), "snapshot.json")
	if err := os.WriteFile(path, tc.Snapshot, 0600); err != nil {
		t.Fatal(err)
	}
	want, err := environment.ValidateArtifacts(tc.Snapshot, nil)
	if err != nil {
		t.Fatal(err)
	}
	wantJSON, err := json.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	var out, stderr bytes.Buffer
	code := runCLIWithInput([]string{"environment", "validate", "--snapshot", path, "--format=json"}, strings.NewReader(""), &out, &stderr)
	if code != 0 || stderr.Len() != 0 || !bytes.Equal(bytes.TrimSpace(out.Bytes()), wantJSON) {
		t.Fatalf("code=%d stderr=%s report=%s", code, &stderr, &out)
	}
}

func TestEnvironmentCLIInputAndOutputFailures(t *testing.T) {
	tc := environmentCases(t)[0]
	dir := t.TempDir()
	snapshotPath := filepath.Join(dir, "snapshot.json")
	bundlePath := filepath.Join(dir, "schemas.json")
	if err := os.WriteFile(snapshotPath, tc.Snapshot, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(bundlePath, tc.SchemaBundle, 0600); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{
		{"environment", "validate"},
		{"environment", "validate", "--snapshot", "-"},
		{"environment", "validate", "--snapshot", filepath.Join(dir, "missing")},
		{"environment", "validate", "--snapshot", snapshotPath, "--format", "yaml"},
		{"environment", "validate", "--snapshot", snapshotPath, "--output", snapshotPath},
		{"environment", "validate", "--snapshot", snapshotPath, "--schemas", bundlePath, "--output", bundlePath},
		{"environment", "validate", "--snapshot", snapshotPath, "--output", filepath.Join(dir, "missing", "report.json")},
	} {
		var out, stderr bytes.Buffer
		if code := runCLIWithInput(args, strings.NewReader(""), &out, &stderr); code != 2 || out.Len() != 0 || stderr.Len() == 0 {
			t.Errorf("args=%v code=%d stdout=%s stderr=%s", args, code, &out, &stderr)
		}
	}
	still, err := os.ReadFile(snapshotPath)
	if err != nil || !reflect.DeepEqual(still, []byte(tc.Snapshot)) {
		t.Fatalf("snapshot overwritten: %v", err)
	}
}

func TestEnvironmentCLIMalformedArtifactAttribution(t *testing.T) {
	tc := environmentCases(t)[0]
	for _, malformed := range []string{"snapshot", "schema_bundle"} {
		t.Run(malformed, func(t *testing.T) {
			dir := t.TempDir()
			snapshot, bundle := []byte(tc.Snapshot), []byte(tc.SchemaBundle)
			if malformed == "snapshot" {
				snapshot = []byte(`{"schema_version":1,`)
			} else {
				bundle = []byte(`{"schema_version":1,`)
			}
			snapshotPath, bundlePath := filepath.Join(dir, "snapshot.json"), filepath.Join(dir, "bundle.json")
			if err := os.WriteFile(snapshotPath, snapshot, 0600); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(bundlePath, bundle, 0600); err != nil {
				t.Fatal(err)
			}
			want, err := environment.ValidateArtifacts(snapshot, bundle)
			if err != nil || want.Status != "invalid" || len(want.Diagnostics) == 0 {
				t.Fatalf("reference: %v %#v", err, want)
			}
			var out, stderr bytes.Buffer
			code := runCLIWithInput([]string{"environment", "validate", "--snapshot", snapshotPath, "--schemas", bundlePath, "--format=json"}, strings.NewReader(""), &out, &stderr)
			var got environment.Report
			if err := json.Unmarshal(out.Bytes(), &got); err != nil {
				t.Fatal(err)
			}
			if code != 1 || stderr.Len() != 0 || !reflect.DeepEqual(&got, want) || got.Diagnostics[0].Artifact != malformed {
				t.Fatalf("code=%d stderr=%s got=%#v want=%#v", code, &stderr, &got, want)
			}
		})
	}
}
