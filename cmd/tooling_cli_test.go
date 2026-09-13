package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/delgado-jacob/spl-toolkit/internal/buildinfo"
	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/corpus"
	"github.com/delgado-jacob/spl-toolkit/pkg/corpusio"
	"github.com/delgado-jacob/spl-toolkit/pkg/document"
	"github.com/delgado-jacob/spl-toolkit/pkg/graph"
	"github.com/delgado-jacob/spl-toolkit/pkg/impact"
	"github.com/delgado-jacob/spl-toolkit/pkg/sarif"
)

func toolingFile(t *testing.T, dir, name, text string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(text), 0600); err != nil {
		t.Fatal(err)
	}
	return p
}
func toolingRun(args ...string) (int, string, string) {
	var out, err bytes.Buffer
	code := runCLIWithInput(args, strings.NewReader(""), &out, &err)
	return code, out.String(), err.String()
}
func TestScanCLIExitPrecedence(t *testing.T) {
	d := toolingTemp(t)
	manifest := toolingFile(t, d, "manifest.json", `{"schema_version":1,"documents":[{"id":"bad","text":" "},{"id":"lost","path":"missing.spl"}]}`)
	code, out, err := toolingRun("scan", "--manifest", manifest, "--format", "json")
	if code != 2 || !strings.Contains(out, `"execution_complete":false`) || !strings.Contains(out, `"status":"invalid"`) {
		t.Fatalf("%d %s %s", code, out, err)
	}
}
func TestCorpusSourceConflict(t *testing.T) {
	for _, args := range [][]string{{"scan"}, {"scan", "--directory", "x", "--manifest", "y"}, {"graph", "--directory", "x", "--query", "search host=x"}, {"scan", "--directory", "x", "--before-rules", "y"}} {
		if code, _, _ := toolingRun(args...); code != 2 {
			t.Fatal(args, code)
		}
	}
}
func TestImpactCLIReadonly(t *testing.T) {
	d := toolingTemp(t)
	query := toolingFile(t, d, "q.spl", "search host=x")
	manifest := toolingFile(t, d, "m.json", `{"schema_version":1,"documents":[{"id":"q","path":"q.spl"}]}`)
	rules := toolingFile(t, d, "rules.json", `{"schema_version":1,"rules":[]}`)
	for _, output := range []string{query, manifest, rules, filepath.Join(d, "alias")} {
		if output == filepath.Join(d, "alias") {
			if err := os.Link(query, output); err != nil {
				t.Fatal(err)
			}
		}
		code, _, _ := toolingRun("impact-mapping", "--manifest", manifest, "--before-rules", rules, "--after-rules", rules, "--output", output)
		if code != 2 {
			t.Fatalf("output %s: %d", output, code)
		}
	}
	raw, _ := os.ReadFile(query)
	if string(raw) != "search host=x" {
		t.Fatal("source changed")
	}
}

func toolingTemp(t *testing.T) string {
	t.Helper()
	p, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func TestToolingCLICanonicalParity(t *testing.T) {
	d := toolingTemp(t)
	raw := `{"schema_version":1,"documents":[{"id":"a","text":"search host=x"},{"id":"b","text":"| mystery | table host"}]}`
	path := toolingFile(t, d, "manifest.json", raw)
	m, err := corpusio.DecodeManifest([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	input, err := corpusio.LoadManifest(m, path)
	if err != nil {
		t.Fatal(err)
	}
	prepared, _ := corpus.Prepare(corpus.ScanOptions{})
	report, err := prepared.Scan(input)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ command, format string }{{"scan", "json"}, {"scan", "sarif"}, {"graph", "json"}} {
		var want any = report
		switch tc.format {
		case "sarif":
			want, err = sarif.Export(report)
		}
		if tc.command == "graph" {
			want, err = graph.Export(report)
		}
		if err != nil {
			t.Fatal(err)
		}
		expected, _ := json.Marshal(want)
		code, out, stderr := toolingRun(tc.command, "--manifest", path, "--format", tc.format)
		if code != 3 || strings.TrimSpace(out) != string(expected) {
			t.Fatalf("%v parity code %d stderr %s", tc, code, stderr)
		}
	}
}

func TestToolingPreflightAndAliases(t *testing.T) {
	d := toolingTemp(t)
	target := toolingFile(t, d, "bad.json", `{"kind":"field_list","catalog":{"fields":["host","host"]}}`)
	code, out, stderr := toolingRun("scan", "--directory", filepath.Join(d, "absent"), "--target", target)
	if code != 2 || out != "" || strings.Contains(stderr, "root:") {
		t.Fatalf("preflight %d %s %s", code, out, stderr)
	}
	query := toolingFile(t, d, "q.spl", "search host=x")
	alias := filepath.Join(d, "alias.spl")
	if err := os.Symlink(query, alias); err != nil {
		t.Fatal(err)
	}
	for _, output := range []string{query, alias} {
		code, _, _ := toolingRun("scan", "--directory", d, "--output", output)
		if code != 2 {
			t.Fatal("accepted source alias", output, code)
		}
	}
	output := filepath.Join(d, "report.json")
	code, out, stderr = toolingRun("scan", "--directory", d, "--format", "json", "--output", output)
	if code != 0 || out != "" || stderr != "" {
		t.Fatalf("output %d %s %s", code, out, stderr)
	}
	if _, err := os.Stat(output); err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(query)
	if string(raw) != "search host=x" {
		t.Fatal("source modified")
	}
}

func TestToolingImpactAndDocumentGoParity(t *testing.T) {
	d := toolingTemp(t)
	raw := `{"schema_version":1,"documents":[{"id":"one","text":"search host=x","source_id":"one"},{"id":"two","text":"| mystery | table host","source_id":"two"}]}`
	manifest := toolingFile(t, d, "manifest.json", raw)
	targetRaw := `{"kind":"field_list","catalog":{"fields":["host"]}}`
	targetPath := toolingFile(t, d, "target.json", targetRaw)
	rulesPath := toolingFile(t, d, "rules.json", `{"schema_version":1,"rules":[]}`)
	m, _ := corpusio.DecodeManifest([]byte(raw))
	input, err := corpusio.LoadManifest(m, manifest)
	if err != nil {
		t.Fatal(err)
	}
	target, err := corpus.DecodeValidationTarget([]byte(targetRaw))
	if err != nil {
		t.Fatal(err)
	}
	for _, command := range []string{"impact-schema", "impact-mapping"} {
		var want *impact.Report
		args := []string{command, "--manifest", manifest, "--format", "json"}
		if command == "impact-schema" {
			p, e := impact.PrepareSchemas(*target, *target)
			if e != nil {
				t.Fatal(e)
			}
			want, err = p.Compare(input)
			args = append(args, "--before-target", targetPath, "--after-target", targetPath)
		} else {
			side, e := readToolingMapping(rulesPath, "")
			if e != nil {
				t.Fatal(e)
			}
			p, e := impact.PrepareMappings(side, side)
			if e != nil {
				t.Fatal(e)
			}
			want, err = p.Compare(input)
			args = append(args, "--before-rules", rulesPath, "--after-rules", rulesPath)
		}
		if err != nil {
			t.Fatal(err)
		}
		expected, _ := json.Marshal(want)
		code, out, stderr := toolingRun(args...)
		if code != 3 || strings.TrimSpace(out) != string(expected) {
			t.Fatalf("%s parity: %d %s", command, code, stderr)
		}
	}
	query := " search host=\"😀\"\r\n| table host\n"
	result, err := analysis.Analyze(analysis.QueryDocument{Text: query, SourceID: "exact"})
	if err != nil {
		t.Fatal(err)
	}
	want, err := document.New(result, document.RevisionContext{ToolVersion: buildinfo.Version, ContractVersion: "1"})
	if err != nil {
		t.Fatal(err)
	}
	expected, _ := json.Marshal(want)
	code, out, stderr := toolingRun("document", query, "--source-id", "exact")
	if code != 0 || strings.TrimSpace(out) != string(expected) {
		t.Fatalf("document parity %d %s", code, stderr)
	}
}

func TestImpactCLIExitReduction(t *testing.T) {
	for _, tc := range []struct {
		before, after analysis.Status
		incomplete    bool
		code          int
	}{{analysis.Valid, analysis.Valid, false, 0}, {analysis.Valid, analysis.Incomplete, true, 3}, {analysis.Invalid, analysis.Incomplete, true, 1}} {
		report := &impact.Report{ExecutionComplete: true, Entries: []impact.ImpactEntry{{BeforeStatus: tc.before, AfterStatus: tc.after}}}
		if tc.incomplete {
			report.Counts.Indeterminate = 1
		}
		if got := impactCLIExit(report); got != tc.code {
			t.Fatalf("%v: %d", tc, got)
		}
		report.ExecutionComplete = false
		if impactCLIExit(report) != 2 {
			t.Fatal("operational precedence")
		}
	}
}
