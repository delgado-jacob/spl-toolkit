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
	if corpus.Version != "1" || len(corpus.Cases) == 0 {
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

func requirementsCLIArgs(document analysis.QueryDocument) []string {
	return []string{
		"requirements",
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
		{name: "unsupported language", args: []string{"analyze", "--query", "search a=1", "--language", "unknown"}, code: 2},
		{name: "unsupported profile", args: []string{"analyze", "--query", "search a=1", "--profile", "cloud"}, code: 2},
		{name: "unsupported version", args: []string{"analyze", "--query", "search a=1", "--compatibility-version", "9.4"}, code: 2},
		{name: "invalid query UTF-8", args: []string{"analyze", "--query", string([]byte{0xff})}, code: 2},
		{name: "invalid source UTF-8", args: []string{"analyze", "--query", "search a=1", "--source-id", string([]byte{0xff})}, code: 2},
		{name: "capabilities rejects query", args: []string{"capabilities", "--query", "search a=1"}, code: 2},
		{name: "capabilities rejects source identity", args: []string{"capabilities", "--source-id", "x"}, code: 2},
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

func TestRequirementsCLIInputBoundary(t *testing.T) {
	for _, args := range [][]string{
		{"requirements", "search host=web", "--format", "json"},
		{"requirements", "--query", "search host=web", "--format", "json"},
	} {
		code, stdout, stderr := runCLITest(args...)
		if code != 0 || stderr != "" {
			t.Fatalf("accepted input %v: code=%d stdout=%q stderr=%q", args, code, stdout, stderr)
		}
		var got analysis.RequirementSet
		if err := json.Unmarshal([]byte(stdout), &got); err != nil {
			t.Fatalf("decode accepted input %v: %v\n%s", args, err, stdout)
		}
	}

	for _, test := range []struct {
		name string
		args []string
	}{
		{name: "duplicate input", args: []string{"requirements", "search host=web", "--query", "search user=alice"}},
		{name: "file input", args: []string{"requirements", "--file", "query.spl"}},
		{name: "stdin input", args: []string{"requirements", "--stdin"}},
		{name: "batch input", args: []string{"requirements", "--batch", "queries.json"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			code, stdout, stderr := runCLITest(test.args...)
			if code != 2 || stdout != "" || stderr == "" {
				t.Fatalf("code=%d stdout=%q stderr=%q", code, stdout, stderr)
			}
		})
	}

	code, stdout, stderr := runCLITest("help")
	if code != 0 || stderr != "" || !strings.Contains(stdout, "requirements [query]") {
		t.Fatalf("help: code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	for _, clause := range []string{
		"Analyze, validate-fields, validate-schema, and rewrite use status-only exits: 0 valid, 1 invalid content, 3 incomplete analysis.",
		"Requirements uses coverage-aware exits: 0 valid with complete requirement coverage, 1 invalid content, 3 incomplete analysis or requirement coverage.",
	} {
		if !strings.Contains(stdout, clause) {
			t.Errorf("help missing command-specific exit clause %q", clause)
		}
	}
	if strings.Contains(stdout, "Analyze, requirements") {
		t.Errorf("help still groups analyze with coverage-aware requirements exits: %q", stdout)
	}
}

func TestRequirementsCLIJSONMatchesGo(t *testing.T) {
	for _, document := range []analysis.QueryDocument{
		{Text: "search host=web", Language: "spl", Profile: "splunkd", Version: "current", SourceID: "queries/valid.spl"},
		{Text: "FROM main | mystery", Language: "spl2", Profile: "splunkd", Version: "current", SourceID: "queries/incomplete.spl2"},
	} {
		document := document
		t.Run(document.Language, func(t *testing.T) {
			want, err := analysis.Requirements(document)
			if err != nil {
				t.Fatal(err)
			}
			code, stdout, stderr := runCLITest(requirementsCLIArgs(document)...)
			if code != requirementsExitCode(want) || stderr != "" {
				t.Fatalf("code=%d stdout=%q stderr=%q", code, stdout, stderr)
			}
			var got analysis.RequirementSet
			if err := json.Unmarshal([]byte(stdout), &got); err != nil {
				t.Fatalf("decode: %v\n%s", err, stdout)
			}
			if !reflect.DeepEqual(&got, want) {
				t.Fatalf("requirement set mismatch\ngot:  %#v\nwant: %#v", &got, want)
			}
		})
	}

	document := analysis.QueryDocument{Text: "search host=web", Language: "spl", Profile: "splunkd", Version: "current", SourceID: "queries/output.spl"}
	want, err := analysis.Requirements(document)
	if err != nil {
		t.Fatal(err)
	}
	wantJSON, err := json.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	wantJSON = append(wantJSON, '\n')
	output := filepath.Join(t.TempDir(), "requirements.json")
	args := append(requirementsCLIArgs(document), "--output", output)
	code, stdout, stderr := runCLITest(args...)
	if code != 0 || stdout != "" || stderr != "" {
		t.Fatalf("output file: code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	gotJSON, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotJSON, wantJSON) {
		t.Fatalf("output bytes differ\ngot:  %q\nwant: %q", gotJSON, wantJSON)
	}
}

func TestRequirementsCLIExitCodes(t *testing.T) {
	for _, test := range []struct {
		name string
		set  analysis.RequirementSet
		want int
	}{
		{name: "valid complete", set: analysis.RequirementSet{QueryStatus: analysis.Valid, Coverage: analysis.RequirementCoverage{Complete: true}}, want: 0},
		{name: "invalid", set: analysis.RequirementSet{QueryStatus: analysis.Invalid, Coverage: analysis.RequirementCoverage{Complete: true}}, want: 1},
		{name: "incomplete query", set: analysis.RequirementSet{QueryStatus: analysis.Incomplete, Coverage: analysis.RequirementCoverage{Complete: false}}, want: 3},
		{name: "valid incomplete coverage", set: analysis.RequirementSet{QueryStatus: analysis.Valid, Coverage: analysis.RequirementCoverage{Complete: false}}, want: 3},
		{name: "unknown status", set: analysis.RequirementSet{QueryStatus: analysis.Status("future"), Coverage: analysis.RequirementCoverage{Complete: true}}, want: 2},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := requirementsExitCode(&test.set); got != test.want {
				t.Fatalf("exit=%d, want %d", got, test.want)
			}
		})
	}

	code, stdout, stderr := runCLITest("requirements", "--query", "search host=web", "--unknown")
	if code != 2 || stdout != "" || stderr == "" {
		t.Fatalf("invalid option: code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}

	var errOut bytes.Buffer
	code = runCLI([]string{"requirements", "--query", "search host=web"}, failingCLIWriter{}, &errOut)
	if code != 2 || !strings.Contains(errOut.String(), "write rejected") {
		t.Fatalf("output failure: code=%d stderr=%q", code, errOut.String())
	}
}

func TestRequirementsCLITextStableAndOrdered(t *testing.T) {
	set := &analysis.RequirementSet{
		QueryStatus: analysis.Incomplete,
		Coverage:    analysis.RequirementCoverage{Complete: false, Reasons: []string{"reason-1", "reason-2"}},
		Items: []analysis.RequirementItem{{
			ID: "req-1", Kind: "field", Identity: "host", Role: "read", Necessity: "conditional", Origin: "direct", Resolution: "wildcard",
			Occurrences: []analysis.RequirementOccurrence{{
				ReferenceID: "ref-1", OriginalName: "h*", Binding: "indeterminate", StageID: "stage-1", ScopeID: "scope-0",
				Location: analysis.Location{Start: analysis.Position{Offset: 2, Line: 1, Column: 3}, End: analysis.Position{Offset: 4, Line: 1, Column: 5}},
			}},
		}},
		Gaps: []analysis.RequirementGap{{Code: "gap-1", Message: "gap message", ReferenceIDs: []string{"ref-1"}, DiagnosticCodes: []string{"diag-1"}}},
		Diagnostics: []analysis.Diagnostic{{
			Code: "diag-1", Severity: "warning", Category: "test", Message: "diagnostic message",
			Location: analysis.Location{Start: analysis.Position{Offset: 2, Line: 1, Column: 3}, End: analysis.Position{Offset: 4, Line: 1, Column: 5}},
		}},
	}
	want := "Query status: incomplete\n" +
		"Requirement coverage: incomplete\n" +
		"Coverage reasons: reason-1, reason-2\n" +
		"Items:\n" +
		"  - req-1 field/read \"host\" (conditional, direct, wildcard)\n" +
		"    occurrence ref-1 \"h*\" (indeterminate) stage=stage-1 scope=scope-0 @ 1:3-1:5, bytes 2-4\n" +
		"Gaps:\n" +
		"  - gap-1: gap message (references: ref-1; diagnostics: diag-1)\n" +
		"Diagnostics:\n" +
		"  - diag-1 [warning/test] @ 1:3-1:5, bytes 2-4: diagnostic message\n"
	if got := string(formatRequirementsText(set)); got != want {
		t.Fatalf("text mismatch\ngot:\n%s\nwant:\n%s", got, want)
	}
	if got := string(formatRequirementsText(set)); got != want {
		t.Fatalf("second formatting changed\ngot:\n%s", got)
	}
}

func TestRequirementsCLILexerBudgetParity(t *testing.T) {
	for _, language := range []string{"spl", "spl2"} {
		language := language
		t.Run(language, func(t *testing.T) {
			exact := strings.Repeat("x ", 2048)
			for _, test := range []struct {
				name          string
				text          string
				resourceLimit bool
			}{
				{name: "exact", text: exact},
				{name: "plus one", text: exact + "x", resourceLimit: true},
				{name: "long sparse", text: `search note="` + strings.Repeat("a", 64*1024) + `"`},
			} {
				t.Run(test.name, func(t *testing.T) {
					document := analysis.QueryDocument{Text: test.text, Language: language, Profile: "splunkd", Version: "current", SourceID: "adapter-budget"}
					wantAnalysis, err := analysis.Analyze(document)
					if err != nil {
						t.Fatal(err)
					}
					wantRequirements, err := analysis.Requirements(document)
					if err != nil {
						t.Fatal(err)
					}
					gotLimit := len(wantAnalysis.Diagnostics) == 1 && wantAnalysis.Diagnostics[0].Code == analysis.CodeAnalysisResourceLimit
					if gotLimit != test.resourceLimit {
						t.Fatalf("resource limit=%t, want %t", gotLimit, test.resourceLimit)
					}

					code, stdout, stderr := runCLITest(analysisCLIArgs(document)...)
					if code != analysisStatusExitCode(wantAnalysis.Status) || stderr != "" {
						t.Fatalf("analyze code=%d stderr=%q", code, stderr)
					}
					var gotAnalysis analysis.Result
					if err := json.Unmarshal([]byte(stdout), &gotAnalysis); err != nil || !reflect.DeepEqual(&gotAnalysis, wantAnalysis) {
						t.Fatalf("analyze parity: decode=%v", err)
					}

					code, stdout, stderr = runCLITest(requirementsCLIArgs(document)...)
					if code != requirementsExitCode(wantRequirements) || stderr != "" {
						t.Fatalf("requirements code=%d stderr=%q", code, stderr)
					}
					var gotRequirements analysis.RequirementSet
					if err := json.Unmarshal([]byte(stdout), &gotRequirements); err != nil || !reflect.DeepEqual(&gotRequirements, wantRequirements) {
						t.Fatalf("requirements parity: decode=%v", err)
					}
					if test.resourceLimit && (code != 3 || gotAnalysis.Status != analysis.Incomplete || gotRequirements.QueryStatus != analysis.Incomplete) {
						t.Fatalf("resource outcome: code=%d analysis=%q requirements=%q", code, gotAnalysis.Status, gotRequirements.QueryStatus)
					}
				})
			}
		})
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

func TestCapabilitiesCLITextIsCanonicalAndComplete(t *testing.T) {
	manifest := analysis.CapabilityManifest{
		Rewrite: &analysis.RewriteCapabilityManifest{SchemaVersion: 1, Forms: []analysis.RewriteCapabilityForm{
			{Kind: "field", Role: "expression_atom", IdentityForms: []string{"atom"}, Supported: true, Limitations: []string{"exact binding only"}},
		}},
		SchemaVersion:  1,
		Language:       "spl2",
		Profile:        "splunkd",
		Version:        "current",
		ToolkitVersion: "9.9.9",
		Commands: []analysis.Capability{
			{Name: "eval", SyntaxSupported: true, SemanticSupported: true, Limitations: []string{"legacy command limit"}},
		},
		Functions: []analysis.Capability{
			{Name: "lower", SyntaxSupported: true, SemanticSupported: true, Limitations: []string{"legacy function limit"}},
		},
		Records: []analysis.CapabilityRecord{
			{
				ID: "rec-z", Language: "spl2", Profile: "splunkd", Kind: "function", Name: "lower", Form: "call",
				GrammarRegistered: true,
				Dimensions: analysis.CapabilityDimensions{
					Syntax:        analysis.CapabilityClaim{State: analysis.CapabilitySupported, EvidenceIDs: []string{"ev-z"}, Limitations: []string{}},
					Semantics:     analysis.CapabilityClaim{State: analysis.CapabilityPartial, EvidenceIDs: []string{"ev-z"}, Limitations: []string{"dynamic names remain incomplete"}},
					Requirements:  analysis.CapabilityClaim{State: analysis.CapabilitySupported, EvidenceIDs: []string{"ev-z"}, Limitations: []string{}},
					Linting:       analysis.CapabilityClaim{State: analysis.CapabilityNotApplicable, EvidenceIDs: []string{}, Limitations: []string{}},
					SafeRewriting: analysis.CapabilityClaim{State: analysis.CapabilityUnsupported, EvidenceIDs: []string{"ev-a"}, Limitations: []string{"calls are fixed"}},
				},
			},
			{
				ID: "rec-a", Language: "spl2", Profile: "splunkd", Kind: "command", Name: "eval", Form: "assignment",
				GrammarRegistered: false,
				Dimensions: analysis.CapabilityDimensions{
					Syntax:        analysis.CapabilityClaim{State: analysis.CapabilityUnsupported, EvidenceIDs: []string{"ev-a"}, Limitations: []string{"grammar deferred"}},
					Semantics:     analysis.CapabilityClaim{State: analysis.CapabilityUnassessed, EvidenceIDs: []string{}, Limitations: []string{}},
					Requirements:  analysis.CapabilityClaim{State: analysis.CapabilityPartial, EvidenceIDs: []string{"ev-z", "ev-a"}, Limitations: []string{"wildcards unresolved"}},
					Linting:       analysis.CapabilityClaim{State: analysis.CapabilitySupported, EvidenceIDs: []string{"ev-a"}, Limitations: []string{}},
					SafeRewriting: analysis.CapabilityClaim{State: analysis.CapabilityNotApplicable, EvidenceIDs: []string{}, Limitations: []string{}},
				},
			},
		},
		Summary: analysis.CapabilitySummary{
			Syntax:        analysis.CapabilityStateCounts{Applicable: 2, Covered: 1, Supported: 1, Unsupported: 1},
			Semantics:     analysis.CapabilityStateCounts{Applicable: 2, Partial: 1, Unassessed: 1},
			Requirements:  analysis.CapabilityStateCounts{Applicable: 2, Covered: 1, Supported: 1, Partial: 1},
			Linting:       analysis.CapabilityStateCounts{Applicable: 1, Covered: 1, Supported: 1, NotApplicable: 1},
			SafeRewriting: analysis.CapabilityStateCounts{Applicable: 1, Unsupported: 1, NotApplicable: 1},
		},
		Evidence: []analysis.CapabilityEvidence{
			{ID: "ev-z", Classification: analysis.CapabilityEvidenceIncomplete, Document: analysis.QueryDocument{Text: "FROM z", Language: "spl2", Profile: "splunkd", Version: "current", SourceID: "z.spl2"}},
			{ID: "ev-a", Classification: analysis.CapabilityEvidenceNegative, Document: analysis.QueryDocument{Text: "FROM a", Language: "spl2", Profile: "splunkd", Version: "current", SourceID: "a.spl2"}},
		},
	}

	want := "Schema version: 1\n" +
		"Language: spl2\n" +
		"Profile: splunkd\n" +
		"Compatibility version: current\n" +
		"Toolkit version: 9.9.9\n" +
		"Coverage counts:\n" +
		"  Syntax: applicable=2 covered=1 supported=1 partial=0 unsupported=1 not_applicable=0 unassessed=0\n" +
		"  Semantics: applicable=2 covered=0 supported=0 partial=1 unsupported=0 not_applicable=0 unassessed=1\n" +
		"  Requirements: applicable=2 covered=1 supported=1 partial=1 unsupported=0 not_applicable=0 unassessed=0\n" +
		"  Linting: applicable=1 covered=1 supported=1 partial=0 unsupported=0 not_applicable=1 unassessed=0\n" +
		"  Safe rewriting: applicable=1 covered=0 supported=0 partial=0 unsupported=1 not_applicable=1 unassessed=0\n" +
		"Records:\n" +
		"  command:\n" +
		"    - eval/assignment [rec-a]: grammar_registered=false syntax=unsupported semantics=unassessed requirements=partial linting=supported safe_rewriting=not_applicable\n" +
		"      syntax: evidence=ev-a; limitations=grammar deferred\n" +
		"      requirements: evidence=ev-z,ev-a; limitations=wildcards unresolved\n" +
		"  function:\n" +
		"    - lower/call [rec-z]: grammar_registered=true syntax=supported semantics=partial requirements=supported linting=not_applicable safe_rewriting=unsupported\n" +
		"      semantics: evidence=ev-z; limitations=dynamic names remain incomplete\n" +
		"      safe_rewriting: evidence=ev-a; limitations=calls are fixed\n" +
		"Evidence:\n" +
		"  - ev-a [negative]: spl2/splunkd/current source=\"a.spl2\" query=\"FROM a\"\n" +
		"  - ev-z [incomplete]: spl2/splunkd/current source=\"z.spl2\" query=\"FROM z\"\n" +
		"Rewrite schema version: 1\n" +
		"Rewrite forms:\n" +
		"  - field/expression_atom: supported=true identities=atom (exact binding only)\n" +
		"Commands:\n" +
		"  - eval: syntax=true semantic=true (legacy command limit)\n" +
		"Functions:\n" +
		"  - lower: syntax=true semantic=true (legacy function limit)\n"

	if got := string(formatCapabilitiesText(manifest)); got != want {
		t.Fatalf("capability text mismatch\ngot:\n%s\nwant:\n%s", got, want)
	}
}
