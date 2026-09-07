package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
)

type analysisCorpusCase struct {
	ID       string                 `json:"id"`
	Document analysis.QueryDocument `json:"document"`
	Expected *analysis.Result       `json:"expected"`
}

func loadAnalysisCorpus(t *testing.T) []analysisCorpusCase {
	t.Helper()
	data, err := os.ReadFile("../testdata/analysis/cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var corpus struct {
		Version string               `json:"version"`
		Cases   []analysisCorpusCase `json:"cases"`
	}
	if err := json.Unmarshal(data, &corpus); err != nil {
		t.Fatal(err)
	}
	if corpus.Version != "1" || len(corpus.Cases) != 24 {
		t.Fatalf("missing reviewed corpus: version=%q cases=%d", corpus.Version, len(corpus.Cases))
	}
	return corpus.Cases
}

func analysisExitCode(status analysis.Status) int {
	switch status {
	case analysis.Valid:
		return 0
	case analysis.Invalid:
		return 1
	case analysis.Incomplete:
		return 3
	default:
		return -1
	}
}

func analysisCLIArgs(document analysis.QueryDocument) []string {
	return []string{
		"analyze",
		"--query", document.Text,
		"--language", document.Language,
		"--profile", document.Profile,
		"--compatibility-version", document.Version,
		"--source-id", document.SourceID,
		"--format", "json",
	}
}

func TestAnalysisCLIReportsMatchCorpus(t *testing.T) {
	for _, corpusCase := range loadAnalysisCorpus(t) {
		t.Run(corpusCase.ID, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := runCLI(analysisCLIArgs(corpusCase.Document), &stdout, &stderr)
			if code != analysisExitCode(corpusCase.Expected.Status) || stderr.Len() != 0 {
				t.Fatalf("code=%d stderr=%q", code, stderr.String())
			}
			var got analysis.Result
			if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
				t.Fatalf("decode report: %v\n%s", err, stdout.Bytes())
			}
			if !reflect.DeepEqual(&got, corpusCase.Expected) {
				t.Fatalf("report mismatch\ngot:  %#v\nwant: %#v", &got, corpusCase.Expected)
			}
		})
	}
}

func TestAnalysisCLIInputAndExitContracts(t *testing.T) {
	for _, test := range []struct {
		name string
		args []string
		code int
	}{
		{name: "omitted query", args: []string{"analyze", "--format", "json"}, code: 2},
		{name: "unsupported language", args: []string{"analyze", "--query", "search a=1", "--language", "spl2"}, code: 2},
		{name: "unsupported profile", args: []string{"analyze", "--query", "search a=1", "--profile", "cloud"}, code: 2},
		{name: "unsupported version", args: []string{"analyze", "--query", "search a=1", "--compatibility-version", "9.4"}, code: 2},
		{name: "invalid query UTF-8", args: []string{"analyze", "--query", string([]byte{0xff})}, code: 2},
		{name: "invalid source UTF-8", args: []string{"analyze", "--query", "search a=1", "--source-id", string([]byte{0xff})}, code: 2},
		{name: "capabilities rejects query", args: []string{"capabilities", "--query", "search a=1"}, code: 2},
		{name: "capabilities rejects language", args: []string{"capabilities", "--language", "spl"}, code: 2},
		{name: "legacy rejects analysis option", args: []string{"discover", "--query", "search a=1", "--source-id", "x"}, code: 2},
	} {
		t.Run(test.name, func(t *testing.T) {
			code, stdout, stderr := runCLITest(test.args...)
			if code != test.code || stdout != "" || stderr == "" {
				t.Fatalf("code=%d stdout=%q stderr=%q", code, stdout, stderr)
			}
		})
	}

	for _, args := range [][]string{
		{"analyze", "--query", "", "--format", "json"},
		{"analyze", "", "--format", "json"},
		{"analyze", "--query", " \t\r\n", "--format", "json"},
	} {
		code, stdout, stderr := runCLITest(args...)
		if code != 1 || stderr != "" {
			t.Fatalf("explicit empty query: code=%d stdout=%q stderr=%q", code, stdout, stderr)
		}
		var report analysis.Result
		if err := json.Unmarshal([]byte(stdout), &report); err != nil || report.Status != analysis.Invalid {
			t.Fatalf("explicit empty report: status=%q err=%v body=%q", report.Status, err, stdout)
		}
	}

	code, stdout, stderr := runCLITest(
		"analyze", "--query", "search a=1", "--language", "", "--profile", "", "--compatibility-version", "", "--format", "json",
	)
	if code != 0 || stderr != "" {
		t.Fatalf("empty defaults: code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	var defaulted analysis.Result
	if err := json.Unmarshal([]byte(stdout), &defaulted); err != nil {
		t.Fatal(err)
	}
	if defaulted.Document.Language != "spl" || defaulted.Document.Profile != "splunkd" || defaulted.Document.Version != "current" {
		t.Fatalf("document options were not normalized: %#v", defaulted.Document)
	}
}

func TestAnalysisCLIOutputFilePreservesCanonicalBytes(t *testing.T) {
	corpusCase := loadAnalysisCorpus(t)[0]
	want, err := json.Marshal(corpusCase.Expected)
	if err != nil {
		t.Fatal(err)
	}
	want = append(want, '\n')
	output := filepath.Join(t.TempDir(), "report.json")
	args := append(analysisCLIArgs(corpusCase.Document), "--output", output)
	code, stdout, stderr := runCLITest(args...)
	if code != analysisExitCode(corpusCase.Expected.Status) || stdout != "" || stderr != "" {
		t.Fatalf("code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	got, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("output bytes differ\ngot:  %q\nwant: %q", got, want)
	}
}

func TestAnalysisCLITextIncludesCoverageLocationsAndDiagnostics(t *testing.T) {
	code, stdout, stderr := runCLITest("analyze", "--query", "search a=1 | mystery x")
	if code != 3 || stderr != "" {
		t.Fatalf("code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	for _, fragment := range []string{
		"Status: incomplete",
		"Syntax coverage: complete",
		"Semantic coverage: incomplete",
		"References:",
		"bytes ",
		"Diagnostics:",
		"SPL_UNSUPPORTED_COMMAND",
	} {
		if !strings.Contains(stdout, fragment) {
			t.Errorf("text output missing %q:\n%s", fragment, stdout)
		}
	}
}

func TestCapabilitiesCLIFormatsAndOutput(t *testing.T) {
	want := analysis.Capabilities()
	code, stdout, stderr := runCLITest("capabilities", "--format", "json")
	if code != 0 || stderr != "" {
		t.Fatalf("code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	var got analysis.CapabilityManifest
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("manifest mismatch\ngot:  %#v\nwant: %#v", got, want)
	}

	output := filepath.Join(t.TempDir(), "capabilities.txt")
	code, stdout, stderr = runCLITest("capabilities", "--output", output)
	if code != 0 || stdout != "" || stderr != "" {
		t.Fatalf("file code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	contents, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	for _, fragment := range []string{"Language: spl", "Profile: splunkd", "Compatibility version: current", "Commands:", "Functions:"} {
		if !bytes.Contains(contents, []byte(fragment)) {
			t.Errorf("capabilities text missing %q:\n%s", fragment, contents)
		}
	}
}
