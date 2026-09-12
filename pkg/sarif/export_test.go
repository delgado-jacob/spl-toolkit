package sarif

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/corpus"
	"github.com/delgado-jacob/spl-toolkit/pkg/validation"
)

type cases struct {
	WindowsRoot        string `json:"windows_root"`
	RelativePath       string `json:"relative_path"`
	EncodedRelativeURI string `json:"encoded_relative_uri"`
	UnicodeSource      string `json:"unicode_source"`
	InlineID           string `json:"inline_id"`
	InlineURI          string `json:"inline_uri"`
}

func loadCases(t *testing.T) cases {
	t.Helper()
	data, err := os.ReadFile("../../testdata/tooling/sarif-cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var c cases
	if err := json.Unmarshal(data, &c); err != nil {
		t.Fatal(err)
	}
	return c
}

func analyzed(id, text string, origin corpus.Origin, diagnostics ...analysis.Diagnostic) corpus.ReportEntry {
	status := analysis.Valid
	for _, d := range diagnostics {
		if d.Severity == "error" {
			status = analysis.Invalid
		}
	}
	result := &analysis.Result{SchemaVersion: 1, Document: analysis.QueryDocument{Text: text, SourceID: "opaque://not/a/path"}, Status: status, Coverage: analysis.Coverage{SyntaxComplete: true, SemanticComplete: true}, Diagnostics: diagnostics}
	return corpus.ReportEntry{ID: id, Origin: origin, SourceHash: corpus.SourceHash(text), Evaluation: &corpus.Evaluation{Kind: "analysis", Analysis: result}}
}

func report(entries ...corpus.ReportEntry) *corpus.Report {
	r := &corpus.Report{SchemaVersion: 1, Status: analysis.Valid, ExecutionComplete: true, Mode: "analysis", Selection: corpus.Selection{Mode: "manifest", Complete: true}, Entries: entries}
	r.Counts.Selected = len(entries)
	for _, e := range entries {
		if e.Failure != nil {
			r.Counts.AcquisitionFailed++
			r.ExecutionComplete = false
			continue
		}
		r.Counts.Analyzed++
		r.Coverage.Syntax.Complete++
		r.Coverage.Syntax.Denominator++
		r.Coverage.Semantic.Complete++
		r.Coverage.Semantic.Denominator++
		if e.Evaluation.Analysis.Status == analysis.Invalid {
			r.Status = analysis.Invalid
			r.StatusCounts.Invalid++
		} else {
			r.StatusCounts.Valid++
		}
	}
	r.Coverage.Schema.NotRequested = true
	return r
}

func point(offset, line, column int) analysis.Position {
	return analysis.Position{Offset: offset, Line: line, Column: column}
}
func diag(code, severity, message string, start, end analysis.Position) analysis.Diagnostic {
	return analysis.Diagnostic{Code: code, Severity: severity, Category: "syntax", Message: message, Location: analysis.Location{Start: start, End: end}}
}

func TestSARIFRuleAndArtifactIndices(t *testing.T) {
	c := loadCases(t)
	root := corpus.Origin{Kind: "file", BaseURI: c.WindowsRoot, RelativePath: c.RelativePath}
	first := analyzed("one", "abc", root, diag("Z_RULE", "error", "first", point(0, 1, 1), point(1, 1, 2)), diag("A_RULE", "warning", "second", point(1, 1, 2), point(2, 1, 3)))
	second := analyzed("two", "abc", root, diag("Z_RULE", "error", "third", point(2, 1, 3), point(3, 1, 4)))
	got, err := Export(report(first, second))
	if err != nil {
		t.Fatal(err)
	}
	run := got.Runs[0]
	if got.Version != "2.1.0" || len(run.Tool.Driver.Rules) != 2 || run.Tool.Driver.Rules[0].ID != "A_RULE" || run.Tool.Driver.Rules[1].ID != "Z_RULE" {
		t.Fatalf("rule table: %+v", run.Tool.Driver.Rules)
	}
	if len(run.Artifacts) != 1 || len(run.Results) != 3 {
		t.Fatalf("artifacts/results: %+v", run)
	}
	for i, want := range []struct {
		id    string
		index int
	}{{"Z_RULE", 1}, {"A_RULE", 0}, {"Z_RULE", 1}} {
		result := run.Results[i]
		if result.RuleID != want.id || result.RuleIndex != want.index || len(result.Locations) != 1 || result.Locations[0].PhysicalLocation.ArtifactLocation.Index == nil || *result.Locations[0].PhysicalLocation.ArtifactLocation.Index != 0 {
			t.Fatalf("result %d: %+v", i, result)
		}
	}
	one, _ := json.Marshal(got)
	again, err := Export(report(first, second))
	if err != nil {
		t.Fatal(err)
	}
	two, _ := json.Marshal(again)
	if string(one) != string(two) {
		t.Fatal("export is not deterministic")
	}
}

func TestSARIFURIsAndUnicode(t *testing.T) {
	c := loadCases(t)
	root := corpus.Origin{Kind: "file", BaseURI: c.WindowsRoot, RelativePath: c.RelativePath}
	e := analyzed("file", c.UnicodeSource, root,
		diag("EMOJI", "warning", "emoji", point(1, 1, 2), point(5, 1, 3)),
		diag("CRLF", "warning", "across newline", point(5, 1, 3), point(8, 2, 1)),
		diag("EOF", "note", "insertion", point(10, 2, 2), point(10, 2, 2)))
	inline := analyzed(c.InlineID, "x\r\n", corpus.Origin{Kind: "inline"}, diag("INLINE", "warning", "end", point(3, 2, 1), point(3, 2, 1)))
	got, err := Export(report(e, inline))
	if err != nil {
		t.Fatal(err)
	}
	run := got.Runs[0]
	if run.ColumnKind != "unicodeCodePoints" || run.OriginalURIBases["SRCROOT"].URI != c.WindowsRoot {
		t.Fatalf("root/column: %+v", run)
	}
	if !reflect.DeepEqual(run.NewlineSequences, []string{"\r\n", "\r", "\n"}) {
		t.Fatalf("newline interpretation: %+v", run.NewlineSequences)
	}
	if run.Artifacts[0].Location.URI != c.EncodedRelativeURI || run.Artifacts[0].Location.URIBaseID != "SRCROOT" {
		t.Fatalf("file URI: %+v", run.Artifacts[0].Location)
	}
	if run.Artifacts[1].Location.URI != c.InlineURI || run.Artifacts[1].Contents == nil || run.Artifacts[1].Contents.Text != "x\r\n" {
		t.Fatalf("inline artifact: %+v", run.Artifacts[1])
	}
	for i, want := range []Region{{StartLine: 1, StartColumn: 2, EndLine: 1, EndColumn: 3}, {StartLine: 1, StartColumn: 3, EndLine: 2, EndColumn: 1}, {StartLine: 2, StartColumn: 2, EndLine: 2, EndColumn: 2}, {StartLine: 1, StartColumn: 4, EndLine: 1, EndColumn: 4}} {
		if got := run.Results[i].Locations[0].PhysicalLocation.Region; got == nil || *got != want {
			t.Fatalf("region %d: got %+v want %+v", i, got, want)
		}
	}
	raw, _ := json.Marshal(got)
	if strings.Contains(string(raw), "charOffset") || strings.Contains(string(raw), "charLength") {
		t.Fatalf("incorrect offsets or source URI: %s", raw)
	}
	if run.Results[0].Properties["source_id"] != "opaque://not/a/path" {
		t.Fatalf("source id lost: %+v", run.Results[0].Properties)
	}
}

func TestSARIFInvocationVsContent(t *testing.T) {
	invalid := analyzed("bad", "x", corpus.Origin{Kind: "inline"}, diag("BAD", "error", "invalid", point(0, 1, 1), point(1, 1, 2)))
	clean, err := Export(report(invalid))
	if err != nil {
		t.Fatal(err)
	}
	if !clean.Runs[0].Invocations[0].ExecutionSuccessful || len(clean.Runs[0].Results) != 1 {
		t.Fatalf("content error became tool failure: %+v", clean.Runs[0])
	}
	failed := corpus.ReportEntry{ID: "missing", Origin: corpus.Origin{Kind: "file", RelativePath: "missing.spl", BaseURI: "file:///tmp/scan/"}, Failure: &corpus.AcquisitionError{Code: "not_found", Phase: "open", Message: "file unavailable", ID: "missing", Path: "missing.spl"}}
	mixed := report(invalid, failed)
	mixed.Status = analysis.Invalid
	got, err := Export(mixed)
	if err != nil {
		t.Fatal(err)
	}
	invocation := got.Runs[0].Invocations[0]
	if invocation.ExecutionSuccessful || len(invocation.ToolExecutionNotifications) != 1 || len(got.Runs[0].Results) != 1 || got.Runs[0].Results[0].RuleID != "BAD" {
		t.Fatalf("mixed report lost independent states: %+v", got.Runs[0])
	}
	noFindings := report(analyzed("ok", "x", corpus.Origin{Kind: "inline"}))
	empty, err := Export(noFindings)
	if err != nil {
		t.Fatal(err)
	}
	if !empty.Runs[0].Invocations[0].ExecutionSuccessful || len(empty.Runs[0].Results) != 0 || len(empty.Runs[0].Artifacts) != 1 {
		t.Fatalf("empty findings: %+v", empty.Runs[0])
	}
	if empty.Runs[0].Artifacts[0].Properties["source_id"] != "opaque://not/a/path" {
		t.Fatalf("source id vanished without results: %+v", empty.Runs[0].Artifacts[0])
	}
}

func TestSARIFTraversalFailureWithoutEntries(t *testing.T) {
	r := &corpus.Report{SchemaVersion: 1, Status: analysis.Incomplete, ExecutionComplete: false, Mode: "analysis", Selection: corpus.Selection{Mode: "directory", Complete: false, TraversalFailures: []corpus.AcquisitionError{{Code: "traversal_failed", Phase: "traverse"}}}, Counts: corpus.Counts{TraversalFailed: 1}}
	got, err := Export(r)
	if err != nil {
		t.Fatal(err)
	}
	run := got.Runs[0]
	if len(run.Results) != 0 || len(run.Artifacts) != 0 || run.Invocations[0].ExecutionSuccessful || len(run.Invocations[0].ToolExecutionNotifications) != 1 || run.Invocations[0].ToolExecutionNotifications[0].Message.Text == "" {
		t.Fatalf("zero-entry traversal failure became unexplained or successful: %+v", run)
	}
}

func TestSARIFIncompleteDiagnostics(t *testing.T) {
	e := analyzed("partial", "x", corpus.Origin{Kind: "inline"}, diag("UNKNOWN", "warning", "unknown meaning", point(0, 1, 1), point(1, 1, 2)))
	e.Evaluation.Analysis.Status = analysis.Incomplete
	e.Evaluation.Analysis.Coverage.SemanticComplete = false
	e.Evaluation.Analysis.Coverage.Reasons = []string{"unresolved analysis"}
	r := report(e)
	r.Status = analysis.Incomplete
	r.StatusCounts = corpus.StatusCounts{Incomplete: 1}
	r.Coverage.Semantic = corpus.CoverageCount{Incomplete: 1, Denominator: 1}
	r.CoverageReasons = []string{"unresolved analysis"}
	got, err := Export(r)
	if err != nil {
		t.Fatal(err)
	}
	run := got.Runs[0]
	if !run.Invocations[0].ExecutionSuccessful || len(run.Results) < 1 || run.Results[0].RuleID != "UNKNOWN" || run.Results[0].Level != "warning" {
		t.Fatalf("incomplete diagnostic: %+v", run)
	}
	if !reflect.DeepEqual(run.Properties["coverage_reasons"], []string{"unresolved analysis"}) {
		t.Fatalf("coverage reasons: %+v", run.Properties)
	}
	if len(run.Results) != 2 || run.Results[1].RuleID != completenessRule || len(run.Results[1].Locations) != 0 {
		t.Fatalf("unlocated coverage gap hidden by unrelated located warning: %+v", run.Results)
	}
	unlocated := analyzed("unlocated", "x", corpus.Origin{Kind: "inline"}, diag("GAP", "warning", "unknown", analysis.Position{}, analysis.Position{}))
	unlocated.Evaluation.Analysis.Status = analysis.Incomplete
	unlocated.Evaluation.Analysis.Coverage.SemanticComplete = false
	partial := report(unlocated)
	partial.Status = analysis.Incomplete
	partial.Coverage.Semantic = corpus.CoverageCount{Incomplete: 1, Denominator: 1}
	out, err := Export(partial)
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Runs[0].Results) == 0 || len(out.Runs[0].Results[0].Locations) != 0 {
		t.Fatalf("invented coordinates: %+v", out.Runs[0].Results)
	}
}

func TestSARIFUsesFinalValidationDiagnostics(t *testing.T) {
	e := analyzed("checked", "x", corpus.Origin{Kind: "inline"})
	base := e.Evaluation.Analysis
	e.Evaluation = &corpus.Evaluation{Kind: "field_list", FieldValidation: &validation.Report{
		Analysis: base, Status: analysis.Invalid,
		Coverage:    validation.Coverage{SyntaxComplete: true, SemanticComplete: true, SchemaComplete: true},
		Diagnostics: []analysis.Diagnostic{diag("SPL_UNKNOWN_FIELD", "error", "missing target field", point(0, 1, 1), point(1, 1, 2))},
	}}
	r := report(analyzed("checked", "x", corpus.Origin{Kind: "inline"}))
	r.Entries[0] = e
	r.Mode = "field_list"
	r.Status = analysis.Invalid
	r.Coverage.Schema = corpus.CoverageCount{Complete: 1, Denominator: 1}
	got, err := Export(r)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Runs[0].Results) != 1 || got.Runs[0].Results[0].RuleID != "SPL_UNKNOWN_FIELD" {
		t.Fatalf("final validation finding lost: %+v", got.Runs[0].Results)
	}
}

func TestSARIFConsumesCanonicalCorpusScan(t *testing.T) {
	catalog := validation.FieldCatalog{Fields: []string{"host"}}
	r, err := corpus.Scan(corpus.Request{SchemaVersion: 1, Documents: []corpus.RequestDocument{{ID: "scan", Document: analysis.QueryDocument{Text: "search missing=x", SourceID: "caller-owned-metadata"}}}, ValidationTarget: &corpus.ValidationTarget{Kind: "field_list", Catalog: &catalog}})
	if err != nil {
		t.Fatal(err)
	}
	got, err := Export(r)
	if err != nil {
		t.Fatal(err)
	}
	results := got.Runs[0].Results
	found := false
	for _, result := range results {
		if result.RuleID == validation.CodeUnknownField && result.Level == "error" && result.Properties["source_id"] == "caller-owned-metadata" && len(result.Locations) == 1 {
			found = true
		}
	}
	if !found || !got.Runs[0].Invocations[0].ExecutionSuccessful {
		t.Fatalf("canonical validation evidence lost: %+v", got.Runs[0])
	}
}

func TestSARIFRejectsNonSnapshotLocations(t *testing.T) {
	for name, d := range map[string]analysis.Diagnostic{
		"inside Unicode code point": diag("BAD", "error", "bad", point(2, 1, 2), point(5, 1, 3)),
		"after EOF":                 diag("BAD", "error", "bad", point(1, 1, 2), point(6, 1, 3)),
		"reversed":                  diag("BAD", "error", "bad", point(5, 1, 3), point(1, 1, 2)),
	} {
		t.Run(name, func(t *testing.T) {
			_, err := Export(report(analyzed("bad", "a😀", corpus.Origin{Kind: "inline"}, d)))
			if err == nil {
				t.Fatal("invalid source range exported")
			}
		})
	}
}

func TestSARIFColonFileNameStaysRelative(t *testing.T) {
	e := analyzed("colon", "x", corpus.Origin{Kind: "file", RelativePath: "a:b.spl", BaseURI: "file:///tmp/scan/"})
	got, err := Export(report(e))
	if err != nil {
		t.Fatal(err)
	}
	if got.Runs[0].Artifacts[0].Location.URI != "a%3Ab.spl" {
		t.Fatalf("first path segment looked like a URI scheme: %q", got.Runs[0].Artifacts[0].Location.URI)
	}
}

func TestSARIFRejectsNonDirectoryBaseURI(t *testing.T) {
	for _, base := range []string{"file:///tmp/scan/#", "file:///tmp/scan/?", "file:///C:\\scan/"} {
		e := analyzed("bad-root", "x", corpus.Origin{Kind: "file", RelativePath: "x.spl", BaseURI: base})
		if _, err := Export(report(e)); err == nil {
			t.Fatalf("non-directory file URI accepted: %q", base)
		}
	}
}

func TestSARIFRejectsInvalidUTF8WithoutDiagnostics(t *testing.T) {
	e := analyzed("bad-text", string([]byte{0xff}), corpus.Origin{Kind: "inline"})
	if _, err := Export(report(e)); err == nil {
		t.Fatal("invalid snapshot silently encoded with replacement character")
	}
}
