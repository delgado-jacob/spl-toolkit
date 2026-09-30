package closure

import "testing"

func TestEvaluateEOFDiagnosticMapsToMacroDefinitionPoint(t *testing.T) {
	report, err := Evaluate(evalRequest("search index=main | `m`", []Collection{{Kind: "macro", Coverage: "complete"}}, macroDef("m", "m", "eval x=")))
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range report.Diagnostics {
		if d.Diagnostic.Code != "SPL_SYNTAX_ERROR" || d.Diagnostic.Location.Start.Offset != len(report.EffectiveAnalysis.Document.Text) {
			continue
		}
		want := SourceInterval{Kind: "definition", SourceID: "m.conf", ObjectID: "m", Start: len("eval x="), End: len("eval x=")}
		if d.Diagnostic.Location.End.Offset != d.Diagnostic.Location.Start.Offset || len(d.Origins) != 1 || d.Origins[0] != want || d.Source != want {
			t.Fatalf("EOF diagnostic source=%+v, want %+v", d, want)
		}
		return
	}
	t.Fatalf("expected zero-width EOF syntax diagnostic: %+v", report.Diagnostics)
}

func TestDiagnosticPointOriginChoosesFollowingSegmentAndRetainsPlaceholder(t *testing.T) {
	left := sourceInterval{Kind: "definition", SourceID: "left.conf", ObjectID: "left", Start: 4, End: 5}
	right := sourceInterval{Kind: "query", SourceID: "root.spl", Start: 10, End: 11}
	placeholder := sourceInterval{Kind: "definition", SourceID: "macro.conf", ObjectID: "macro", Start: 1, End: 4}
	expanded := expansion{Text: "ab", Segments: []provenanceSegment{
		{EffectiveStart: 0, EffectiveEnd: 1, Source: left},
		{EffectiveStart: 1, EffectiveEnd: 2, Source: right, Placeholders: []sourceInterval{placeholder}},
	}}
	atBoundary := diagnosticProvenance(expanded, 1, 1)
	if len(atBoundary) != 1 || atBoundary[0].Source != (sourceInterval{Kind: "query", SourceID: "root.spl", Start: 10, End: 10}) || len(atBoundary[0].Placeholders) != 1 || atBoundary[0].Placeholders[0] != placeholder {
		t.Fatalf("internal boundary mapped to wrong source: %+v", atBoundary)
	}
	atEOF := diagnosticProvenance(expanded, 2, 2)
	if len(atEOF) != 1 || atEOF[0].Source != (sourceInterval{Kind: "query", SourceID: "root.spl", Start: 11, End: 11}) || len(atEOF[0].Placeholders) != 1 || atEOF[0].Placeholders[0] != placeholder {
		t.Fatalf("EOF did not retain the final source and placeholder: %+v", atEOF)
	}
}
