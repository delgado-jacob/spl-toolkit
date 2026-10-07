package main

import (
	"bytes"
	"encoding/json"
	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/compatibility"
	"github.com/delgado-jacob/spl-toolkit/pkg/corpus"
	"github.com/delgado-jacob/spl-toolkit/pkg/resolution"
	"github.com/delgado-jacob/spl-toolkit/pkg/workflow"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWorkflowEvidenceRequiresExactlyOneSource(t *testing.T) {
	var out, stderr bytes.Buffer
	code := runCLI([]string{"workflow", "evidence"}, &out, &stderr)
	if code != 2 || out.Len() != 0 || !strings.Contains(stderr.String(), "exactly one") {
		t.Fatalf("code=%d stdout=%q stderr=%q", code, out.String(), stderr.String())
	}
}

func workflowCLIRequest(t *testing.T, text string) workflow.Request {
	t.Helper()
	raw, err := os.ReadFile("../examples/resolution/request.json")
	if err != nil {
		t.Fatal(err)
	}
	seed, err := resolution.DecodeRequest(raw)
	if err != nil {
		t.Fatal(err)
	}
	return workflow.Request{SchemaVersion: 1, Documents: []corpus.RequestDocument{{ID: "d1", Document: analysis.QueryDocument{Text: text, Language: "spl2"}}}, Settings: workflow.Settings{SchemaVersion: 1, Snapshot: seed.Compatibility.Snapshot, SchemaBundle: seed.Compatibility.SchemaBundle, Entries: []workflow.EntrySettings{{ID: "d1", Compatibility: &workflow.CheckSettings{QueryScope: seed.Compatibility.QueryScope, InputBindings: []compatibility.InputBinding{}}}}}}
}
func workflowCLIFile(t *testing.T, name string, value any) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	return path
}
func TestWorkflowCLIReportsAndFormats(t *testing.T) {
	for _, tc := range []struct {
		text string
		code int
	}{{"from [{id:1}]", 0}, {"from [{id:1}] | eval broken =", 1}, {"from [{id:1}] | unknowable_command", 3}} {
		request := workflowCLIRequest(t, tc.text)
		request.Format = "json"
		path := workflowCLIFile(t, "request.json", request)
		report, err := workflow.Assess(request)
		if err != nil {
			t.Fatal(err)
		}
		for _, format := range []string{"", "text", "json", "sarif", "graph", "bom"} {
			args := []string{"workflow", "assess", "--request", path}
			selected := format
			if format == "" {
				selected = "json"
			} else {
				args = append(args, "--format", format)
			}
			var out, stderr bytes.Buffer
			code := runCLI(args, &out, &stderr)
			if code != tc.code || stderr.Len() != 0 {
				t.Fatalf("format %s code %d stderr %s", format, code, &stderr)
			}
			expected, err := workflow.RenderAssessment(report, selected)
			if err != nil {
				t.Fatal(err)
			}
			var want []byte
			if text, ok := expected.(string); ok {
				want = []byte(text)
			} else {
				want, _ = json.Marshal(expected)
				want = append(want, '\n')
			}
			if !bytes.Equal(out.Bytes(), want) {
				t.Fatalf("format %s projection differs", format)
			}
		}
	}
}
func TestWorkflowCLIOperations(t *testing.T) {
	request := workflowCLIRequest(t, "from [{id:1}]")
	report, err := workflow.Assess(request)
	if err != nil {
		t.Fatal(err)
	}
	reportPath := workflowCLIFile(t, "report.json", report)
	comparison, err := workflow.Compare(workflow.CompareRequest{SchemaVersion: 1, Before: *report, After: *report})
	if err != nil {
		t.Fatal(err)
	}
	comparisonPath := workflowCLIFile(t, "comparison.json", comparison)
	proposal := request.Documents[0].Document
	proposal.Text = "from [{id:1}] | eval broken ="
	recheck := workflow.RecheckRequest{SchemaVersion: 1, Context: workflow.RecheckContext{Original: request.Documents[0], Snapshot: request.Settings.Snapshot, SchemaBundle: request.Settings.SchemaBundle}, Proposal: workflow.Proposal{Document: &proposal, Settings: request.Settings.Entries[0]}}
	recheckPath := workflowCLIFile(t, "recheck.json", recheck)
	for _, tc := range []struct {
		args []string
		code int
	}{{[]string{"compare", "--before", reportPath, "--after", reportPath, "--format", "json"}, 0}, {[]string{"evidence", "--report", reportPath}, 0}, {[]string{"evidence", "--comparison", comparisonPath}, 0}, {[]string{"recheck", "--request", recheckPath, "--format", "json"}, 1}} {
		var out, stderr bytes.Buffer
		if code := runCLI(append([]string{"workflow"}, tc.args...), &out, &stderr); code != tc.code || !json.Valid(out.Bytes()) || stderr.Len() != 0 {
			t.Fatalf("%v code=%d out=%s err=%s", tc.args, code, &out, &stderr)
		}
	}
}
func TestWorkflowCLIRejectsOptions(t *testing.T) {
	for _, args := range [][]string{{"assess"}, {"assess", "--request", "x", "--directory", "y"}, {"assess", "--directory", "x"}, {"assess", "--request", "x", "--settings", "y"}, {"compare", "--before", "x"}, {"compare", "--before", "x", "--after", "y", "--format", "bom"}, {"evidence", "--report", "x", "--comparison", "y"}, {"evidence", "--report", "x", "--format", "json"}, {"recheck", "--request"}, {"recheck", "--request", "x", "--request", "y"}, {"recheck", "--unknown", "x"}} {
		var out, stderr bytes.Buffer
		if code := runCLI(append([]string{"workflow"}, args...), &out, &stderr); code != 2 || out.Len() != 0 || stderr.Len() == 0 {
			t.Fatalf("%v code=%d out=%s err=%s", args, code, &out, &stderr)
		}
	}
}
func TestWorkflowCLIProtectsInputAliases(t *testing.T) {
	request := workflowCLIRequest(t, "from [{id:1}]")
	path := workflowCLIFile(t, "request.json", request)
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(filepath.Dir(path), "link.json")
	if err = os.Link(path, link); err != nil {
		t.Fatal(err)
	}
	symlink := filepath.Join(filepath.Dir(path), "symlink.json")
	if err = os.Symlink(path, symlink); err != nil {
		t.Fatal(err)
	}
	for _, output := range []string{path, link, symlink} {
		var out, stderr bytes.Buffer
		if code := runCLI([]string{"workflow", "assess", "--request", path, "--output", output}, &out, &stderr); code != 2 || out.Len() != 0 || !strings.Contains(stderr.String(), "overwrite") {
			t.Fatalf("code=%d out=%s err=%s", code, &out, &stderr)
		}
	}
	after, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(before, after) {
		t.Fatal("input changed")
	}
}
func TestWorkflowCLIManifestRetainsFailedEntriesAndHashes(t *testing.T) {
	request := workflowCLIRequest(t, "from [{id:1}]")
	request.Settings.Entries = append(request.Settings.Entries, request.Settings.Entries[0])
	request.Settings.Entries[1].ID = "missing"
	settings := workflowCLIFile(t, "settings.json", request.Settings)
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	query := "from [{id:1}]\r\n"
	if err := os.WriteFile(filepath.Join(dir, "q.spl2"), []byte(query), 0600); err != nil {
		t.Fatal(err)
	}
	manifest := filepath.Join(dir, "manifest.json")
	raw := `{"schema_version":1,"base":".","documents":[{"id":"d1","path":"q.spl2"},{"id":"missing","path":"missing.spl2"}]}`
	if err := os.WriteFile(manifest, []byte(raw), 0600); err != nil {
		t.Fatal(err)
	}
	var out, stderr bytes.Buffer
	code := runCLI([]string{"workflow", "assess", "--manifest", manifest, "--settings", settings, "--format", "json"}, &out, &stderr)
	var report workflow.Report
	if err := json.Unmarshal(out.Bytes(), &report); err != nil {
		t.Fatalf("code=%d err=%s out=%s", code, &stderr, &out)
	}
	if code != 2 || len(report.Entries) != 2 || report.Entries[0].SourceHash != corpus.SourceHash(query) || report.Entries[1].Failure == nil {
		t.Fatalf("retained report %+v code=%d", report, code)
	}
	out.Reset()
	stderr.Reset()
	code = runCLI([]string{"workflow", "assess", "--manifest", manifest, "--settings", settings, "--output", filepath.Join(dir, "missing.spl2")}, &out, &stderr)
	if code != 2 || !strings.Contains(stderr.String(), "overwrite") {
		t.Fatalf("missing selected output code=%d err=%s", code, &stderr)
	}
}

func TestWorkflowCLIDirectoryOrderingAndSelectedOutputProtection(t *testing.T) {
	request := workflowCLIRequest(t, "from [{id:1}]")
	request.Settings.Entries[0].ID = "a.spl2"
	other := request.Settings.Entries[0]
	other.ID = "z.spl2"
	request.Settings.Entries = append(request.Settings.Entries, other)
	settings := workflowCLIFile(t, "settings.json", request.Settings)
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"z.spl2", "a.spl2"} {
		if err = os.WriteFile(filepath.Join(dir, name), []byte("from [{id:1}]"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	args := []string{"workflow", "assess", "--directory", dir, "--settings", settings, "--format", "json"}
	var out, stderr bytes.Buffer
	code := runCLI(args, &out, &stderr)
	var report workflow.Report
	if code != 0 || json.Unmarshal(out.Bytes(), &report) != nil || len(report.Entries) != 2 || report.Entries[0].ID != "a.spl2" || report.Entries[1].ID != "z.spl2" {
		t.Fatalf("code=%d report=%+v err=%s", code, report, &stderr)
	}
	alias := filepath.Join(dir, "output.json")
	if err = os.Link(filepath.Join(dir, "a.spl2"), alias); err != nil {
		t.Fatal(err)
	}
	for _, output := range []string{settings, filepath.Join(dir, "a.spl2"), alias} {
		out.Reset()
		stderr.Reset()
		if code = runCLI(append(args, "--output", output), &out, &stderr); code != 2 || out.Len() != 0 || !strings.Contains(stderr.String(), "overwrite") {
			t.Fatalf("output %s code=%d err=%s", output, code, &stderr)
		}
	}
}
func TestWorkflowCLIEvidencePreservesSourceCIAndRejectsDuplicateDisclosure(t *testing.T) {
	report, err := workflow.Assess(workflowCLIRequest(t, "from [{id:1}] | eval broken ="))
	if err != nil {
		t.Fatal(err)
	}
	path := workflowCLIFile(t, "report.json", report)
	var out, stderr bytes.Buffer
	code := runCLI([]string{"workflow", "evidence", "--report", path}, &out, &stderr)
	var evidence workflow.EvidenceReport
	if code != 0 || json.Unmarshal(out.Bytes(), &evidence) != nil || evidence.SourceCIExitCode != 1 || len(evidence.Disclosure.Requested) != 0 {
		t.Fatalf("code=%d evidence=%+v err=%s", code, evidence, &stderr)
	}
	for _, includes := range [][]string{{"query_text", "query_text"}, {"unknown"}} {
		args := []string{"workflow", "evidence", "--report", path}
		for _, include := range includes {
			args = append(args, "--include", include)
		}
		out.Reset()
		stderr.Reset()
		if code = runCLI(args, &out, &stderr); code != 2 || out.Len() != 0 {
			t.Fatalf("code=%d out=%s err=%s", code, &out, &stderr)
		}
	}
}
func TestWorkflowCLIOutputWriteFailureOverridesContentCI(t *testing.T) {
	path := workflowCLIFile(t, "request.json", workflowCLIRequest(t, "from [{id:1}] | eval broken ="))
	var out, stderr bytes.Buffer
	code := runCLI([]string{"workflow", "assess", "--request", path, "--output", t.TempDir()}, &out, &stderr)
	if code != 2 || out.Len() != 0 || stderr.Len() == 0 {
		t.Fatalf("code=%d out=%s err=%s", code, &out, &stderr)
	}
}
