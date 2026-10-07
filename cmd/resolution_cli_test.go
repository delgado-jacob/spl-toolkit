package main

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
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
	data, err := os.ReadFile("../testdata/resolution/cases.json")
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

func resolutionRequestFile(t *testing.T, raw []byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "request.json")
	if err := os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	return path
}
func TestResolutionCLIReports(t *testing.T) {
	for _, c := range resolutionSurfaceCases(t) {
		if c.Expected.RequestError != "" {
			continue
		}
		t.Run(c.ID, func(t *testing.T) {
			raw, err := json.Marshal(resolutionSurfaceRequest(t, c))
			if err != nil {
				t.Fatal(err)
			}
			report, err := resolution.ResolveJSON(raw)
			if err != nil {
				t.Fatal(err)
			}
			want, _ := json.Marshal(report)
			wantCode := 0
			if c.Expected.Counts["incomplete"] > 0 {
				wantCode = 3
			}
			if c.Expected.Counts["failed"] > 0 {
				wantCode = 1
			}
			input := resolutionRequestFile(t, raw)
			for _, format := range []string{"json", "text"} {
				var out, errors bytes.Buffer
				output := filepath.Join(t.TempDir(), "report."+format)
				code := runCLIWithInput([]string{"resolve", "--request", input, "--format", format, "--output", output}, strings.NewReader(""), &out, &errors)
				got, err := os.ReadFile(output)
				if err != nil {
					t.Fatal(err)
				}
				expected := want
				if format == "text" {
					expected = []byte(resolution.FormatReport(report))
				}
				if code != wantCode || out.Len() != 0 || errors.Len() != 0 || !bytes.Equal(bytes.TrimSpace(got), bytes.TrimSpace(expected)) {
					t.Fatalf("code=%d want=%d errors=%s report=%s", code, wantCode, &errors, got)
				}
			}
			var out, errors bytes.Buffer
			if code := runCLIWithInput([]string{"resolve", "--request", input, "--format=json"}, strings.NewReader(""), &out, &errors); code != wantCode || errors.Len() != 0 || !bytes.Equal(bytes.TrimSpace(out.Bytes()), want) {
				t.Fatalf("stdout parity: code=%d errors=%s", code, &errors)
			}
		})
	}
}
func TestResolutionCLIRequestErrors(t *testing.T) {
	raws := [][]byte{nil, []byte(`{`), []byte(`null`), []byte(`{"schema_version":1,"extra":true}`), {0xff}}
	for _, c := range resolutionSurfaceCases(t) {
		if c.Expected.RequestError != "" {
			raw, _ := json.Marshal(resolutionSurfaceRequest(t, c))
			raws = append(raws, raw)
		}
	}
	for _, raw := range raws {
		_, err := resolution.ResolveJSON(raw)
		detail, ok := resolution.RequestErrorDetails(err)
		if !ok {
			t.Fatal(err)
		}
		want, _ := json.Marshal(detail)
		input := resolutionRequestFile(t, raw)
		for _, format := range []string{"json", "text"} {
			var out, errors bytes.Buffer
			code := runCLIWithInput([]string{"resolve", "--request", input, "--format", format}, strings.NewReader(""), &out, &errors)
			if code != 2 || out.Len() != 0 {
				t.Fatalf("code=%d stdout=%s", code, &out)
			}
			if format == "json" && !bytes.Equal(bytes.TrimSpace(errors.Bytes()), want) {
				t.Fatalf("%s want %s", &errors, want)
			}
			if format == "text" && !strings.Contains(errors.String(), detail.Message) {
				t.Fatalf("%s", &errors)
			}
		}
	}
}
func TestResolutionCLIPathAndOptionsSafety(t *testing.T) {
	raw, _ := json.Marshal(resolutionSurfaceRequest(t, resolutionSurfaceCases(t)[0]))
	input := resolutionRequestFile(t, raw)
	dir := filepath.Dir(input)
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
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	relative, err := filepath.Rel(cwd, input)
	if err != nil {
		t.Fatal(err)
	}
	cases := [][]string{{}, {"--request", "-"}, {"--request", dir}, {"--request", filepath.Join(dir, "missing")}, {"--request", input, "--request", input}, {"--request", input, "--stdin"}, {"--request", input, "--format=yaml"}, {"--request", input, "--require-connected"}}
	for _, output := range []string{input, relative, hard, link, filepath.Join(parent, "request.json"), filepath.Join(dir, "missing", "report.json")} {
		cases = append(cases, []string{"--request", input, "--output", output})
	}
	for _, args := range cases {
		var out, errors bytes.Buffer
		if code := runCLIWithInput(append([]string{"resolve"}, args...), strings.NewReader(""), &out, &errors); code != 2 || out.Len() != 0 || errors.Len() == 0 {
			t.Fatalf("%v code=%d out=%s errors=%s", args, code, &out, &errors)
		}
	}
	got, err := os.ReadFile(input)
	if err != nil || !bytes.Equal(got, raw) {
		t.Fatal("request changed")
	}
	files, err := filepath.Glob(filepath.Join(dir, ".spl-toolkit-output-*"))
	if err != nil || len(files) != 0 {
		t.Fatalf("leaked outputs: %v", files)
	}
	var out, errors bytes.Buffer
	if code := runCLIWithInput([]string{"resolve", "--help"}, strings.NewReader(""), &out, &errors); code != 0 || !strings.Contains(out.String(), "resolve --request") {
		t.Fatalf("help: %d %s %s", code, &out, &errors)
	}
}
func TestResolutionCLIWriteFailurePrecedesPolicy(t *testing.T) {
	for _, c := range resolutionSurfaceCases(t) {
		if c.ID != "missing-field-sibling" {
			continue
		}
		raw, _ := json.Marshal(resolutionSurfaceRequest(t, c))
		input := resolutionRequestFile(t, raw)
		var errors bytes.Buffer
		if code := runCLIWithInput([]string{"resolve", "--request", input, "--format=json"}, strings.NewReader(""), failingCLIWriter{}, &errors); code != 2 || errors.Len() == 0 {
			t.Fatalf("code=%d errors=%s", code, &errors)
		}
	}
}
func TestResolutionCLIExitPriority(t *testing.T) {
	for _, counts := range []resolution.Counts{{Verified: 1}, {Failed: 1}, {Incomplete: 1}, {Failed: 1, Incomplete: 1}} {
		want := 0
		if counts.Incomplete > 0 {
			want = 3
		}
		if counts.Failed > 0 {
			want = 1
		}
		if got := resolutionCLIExit(&resolution.Report{Counts: counts}); got != want {
			t.Fatalf("%+v: %d want %d", counts, got, want)
		}
	}
}
