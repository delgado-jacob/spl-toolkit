package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
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
	raw, err := os.ReadFile("../testdata/compatibility/cases.json")
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

func TestCompatibilityCLIReportsAndPolicy(t *testing.T) {
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
			if report.Outcome != c.Expected.Compatibility.Outcome || report.Correlation.Outcome != c.Expected.Correlation.Outcome {
				t.Fatalf("authored facts: %s / %s", report.Outcome, report.Correlation.Outcome)
			}
			path := filepath.Join(t.TempDir(), "request.json")
			if err := os.WriteFile(path, raw, 0600); err != nil {
				t.Fatal(err)
			}
			want, _ := json.Marshal(report)
			for _, connected := range []bool{false, true} {
				args := []string{"compatibility", "--request", path, "--format", "json"}
				if connected {
					args = append(args, "--require-connected")
				}
				var stdout, stderr bytes.Buffer
				code := runCLIWithInput(args, strings.NewReader(""), &stdout, &stderr)
				wantCode := map[string]int{"satisfied": 0, "unsatisfied": 1, "incomplete": 3, "not assessed": 3}[report.Outcome]
				if connected && report.Correlation.Outcome == "disconnected" {
					wantCode = 1
				}
				if connected && report.Correlation.Outcome == "indeterminate" && wantCode == 0 {
					wantCode = 3
				}
				if code != wantCode || stderr.Len() != 0 || !bytes.Equal(bytes.TrimSpace(stdout.Bytes()), want) {
					t.Fatalf("code=%d want=%d stderr=%s report=%s", code, wantCode, &stderr, &stdout)
				}
			}
			var stdout, stderr bytes.Buffer
			output := filepath.Join(t.TempDir(), "report.txt")
			code := runCLIWithInput([]string{"compatibility", "--request", path, "--output", output}, strings.NewReader(""), &stdout, &stderr)
			got, err := os.ReadFile(output)
			if err != nil {
				t.Fatal(err)
			}
			if code != map[string]int{"satisfied": 0, "unsatisfied": 1, "incomplete": 3, "not assessed": 3}[report.Outcome] || stdout.Len() != 0 || stderr.Len() != 0 || string(got) != compatibility.FormatReport(report) {
				t.Fatalf("text/output: code=%d error=%s", code, &stderr)
			}
		})
	}
}

func TestCompatibilityCLIRequestErrorDetails(t *testing.T) {
	for _, raw := range [][]byte{nil, []byte(`{`), []byte(`{"schema_version":1,"extra":true}`), {0xff}} {
		path := filepath.Join(t.TempDir(), "request.json")
		if err := os.WriteFile(path, raw, 0600); err != nil {
			t.Fatal(err)
		}
		_, err := compatibility.CheckJSON(raw)
		detail, ok := compatibility.RequestErrorDetails(err)
		if !ok {
			t.Fatalf("reference error: %v", err)
		}
		want, _ := json.Marshal(detail)
		var stdout, stderr bytes.Buffer
		code := runCLIWithInput([]string{"compatibility", "--request", path, "--format=json", "--require-connected"}, strings.NewReader(""), &stdout, &stderr)
		if code != 2 || stdout.Len() != 0 || !bytes.Equal(bytes.TrimSpace(stderr.Bytes()), want) {
			t.Fatalf("code=%d error=%s want=%s", code, &stderr, want)
		}
		stdout.Reset()
		stderr.Reset()
		code = runCLIWithInput([]string{"compatibility", "--request", path}, strings.NewReader(""), &stdout, &stderr)
		if code != 2 || strings.Contains(stderr.String(), `{"code"`) || !strings.Contains(stderr.String(), detail.Message) {
			t.Fatalf("plain text error: %s", &stderr)
		}
	}
}

func TestCompatibilityCLIPathAndOptionsSafety(t *testing.T) {
	c := compatibilitySurfaceCases(t)[0]
	raw := compatibilitySurfaceRequest(t, c)
	dir := t.TempDir()
	input := filepath.Join(dir, "request.json")
	if err := os.WriteFile(input, raw, 0600); err != nil {
		t.Fatal(err)
	}
	hard := filepath.Join(dir, "hard.json")
	link := filepath.Join(dir, "link.json")
	parent := filepath.Join(dir, "parent")
	if err := os.Link(input, hard); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(input, link); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(dir, parent); err != nil {
		t.Fatal(err)
	}
	cases := [][]string{
		{}, {"--request", "-"}, {"--request", input, "--request", input}, {"--request", input, "--query", "x"}, {"--request", input, "--stdin"}, {"--request", input, "--unknown"}, {"--request", input, "--format", "yaml"}, {"--request", input, "--require-connected=true"}, {"--request", filepath.Join(dir, "missing")},
	}
	for _, output := range []string{input, filepath.Join(dir, ".", "request.json"), hard, link, filepath.Join(parent, "request.json"), filepath.Join(dir, "missing", "report.json")} {
		cases = append(cases, []string{"--request", input, "--output", output})
	}
	for _, args := range cases {
		var stdout, stderr bytes.Buffer
		code := runCLIWithInput(append([]string{"compatibility"}, args...), strings.NewReader(""), &stdout, &stderr)
		if code != 2 || stdout.Len() != 0 || stderr.Len() == 0 {
			t.Errorf("%v: code=%d stdout=%s stderr=%s", args, code, &stdout, &stderr)
		}
	}
	got, err := os.ReadFile(input)
	if err != nil || !bytes.Equal(got, raw) {
		t.Fatal("request changed")
	}
	files, err := filepath.Glob(filepath.Join(dir, ".spl-toolkit-output-*"))
	if err != nil || len(files) != 0 {
		t.Fatalf("temporary outputs leaked: %v", files)
	}
}

func TestCompatibilityCLIExitPriority(t *testing.T) {
	for _, outcome := range []string{"satisfied", "unsatisfied", "incomplete", "not assessed"} {
		for _, correlation := range []string{"connected", "disconnected", "indeterminate", "not applicable"} {
			for _, connected := range []bool{false, true} {
				report := &compatibility.Report{Outcome: outcome}
				report.Correlation.Outcome = correlation
				want := 0
				if outcome == "incomplete" || outcome == "not assessed" || (connected && correlation == "indeterminate") {
					want = 3
				}
				if outcome == "unsatisfied" || (connected && correlation == "disconnected") {
					want = 1
				}
				if got := compatibilityCLIExit(report, connected); got != want {
					t.Fatalf("%s/%s policy=%v: %d want %d", outcome, correlation, connected, got, want)
				}
			}
		}
	}
}

func TestCompatibilityCLIWriteFailurePrecedesPolicy(t *testing.T) {
	var c compatibilitySurfaceCase
	for _, candidate := range compatibilitySurfaceCases(t) {
		if candidate.Name == "closed-schema-missing-field" {
			c = candidate
			break
		}
	}
	raw := compatibilitySurfaceRequest(t, c)
	path := filepath.Join(t.TempDir(), "request.json")
	if err := os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	args := []string{"compatibility", "--request", path, "--format=json", "--require-connected"}
	var stderr bytes.Buffer
	if code := runCLIWithInput(args, strings.NewReader(""), failingCLIWriter{}, &stderr); code != 2 || stderr.Len() == 0 {
		t.Fatalf("write failure: code=%d stderr=%s", code, &stderr)
	}
	stderr.Reset()
	output := filepath.Join(t.TempDir(), "report.json")
	var stdout bytes.Buffer
	code := runCLIWithInput(append(args, "--output", output), strings.NewReader(""), &stdout, &stderr)
	report, err := compatibility.CheckJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	want, _ := json.Marshal(report)
	got, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	if code != 1 || stdout.Len() != 0 || stderr.Len() != 0 || !bytes.Equal(bytes.TrimSpace(got), want) {
		t.Fatalf("JSON output before policy: code=%d stdout=%s stderr=%s", code, &stdout, &stderr)
	}
}
