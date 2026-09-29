package closure

import (
	"strings"
	"testing"
)

func TestProvenanceUnicodeOffsetsAndCrossSegmentReference(t *testing.T) {
	got := expandMacros(macroRequest("é `m(λ)` z", macroDef("m", "m", "$x$+$x$", "x")))
	if got.Text != "é λ+λ z" {
		t.Fatalf("text=%q", got.Text)
	}
	assertCovered(t, got)
	origins := got.Origins(3, 7)
	if len(origins) != 3 {
		t.Fatalf("cross-segment origins=%+v", origins)
	}
	if origins[0].Source.Start != 6 || origins[0].Source.End != 8 || origins[0].Placeholder == nil {
		t.Fatalf("unicode byte offset=%+v", origins[0])
	}
	if origins[1].Source.ObjectID != "m" || origins[1].Source.Start != 3 {
		t.Fatalf("definition origin=%+v", origins[1])
	}
	if origins[2].Source.Start != 6 || origins[2].Placeholder == nil {
		t.Fatalf("second argument origin=%+v", origins[2])
	}
	if len(got.Origins(99, 101)) != 0 {
		t.Fatal("out of range had origins")
	}
}

func TestProvenanceZeroWidthLimitSurvivesSubstitution(t *testing.T) {
	huge := strings.Repeat("x", maxExpansionBytes) + "y"
	got := expandMacros(macroRequest("`outer(`inner`)`", macroDef("outer", "outer", "$x$", "x"), macroDef("inner", "inner", huge)))
	if len(got.Text) != maxExpansionBytes || len(got.Gaps) != 1 || got.Gaps[0].Reason != "limit" || got.Gaps[0].EffectiveStart != len(got.Text) || got.Gaps[0].EffectiveEnd != len(got.Text) {
		t.Fatalf("nested limit lost: text=%d gaps=%+v", len(got.Text), got.Gaps)
	}
	assertCovered(t, got)
}

func TestProvenanceNestedPlaceholderIntervals(t *testing.T) {
	req := macroRequest("`outer(`inner(7)`)`", macroDef("outer", "outer", "$x$", "x"), macroDef("inner", "inner", "$y$", "y"))
	got := expandMacros(req)
	if got.Text != "7" || len(got.Gaps) != 0 || len(got.Segments) != 1 {
		t.Fatalf("nested expansion: %+v", got)
	}
	seg := got.Segments[0]
	queryOffset := strings.Index(req.Document.Text, "7")
	if seg.Source.Kind != "query" || seg.Source.Start != queryOffset || seg.Source.End != queryOffset+1 {
		t.Fatalf("query source: %+v", seg.Source)
	}
	if len(seg.Placeholders) != 2 || seg.Placeholders[0].ObjectID != "outer" || seg.Placeholders[0].Start != 0 || seg.Placeholders[0].End != 3 || seg.Placeholders[1].ObjectID != "inner" || seg.Placeholders[1].Start != 0 || seg.Placeholders[1].End != 3 {
		t.Fatalf("placeholder ancestry: %+v", seg.Placeholders)
	}
	if seg.Placeholder == nil || *seg.Placeholder != seg.Placeholders[0] {
		t.Fatalf("immediate placeholder: %+v", seg)
	}
}

func TestProvenanceRepeatedNestedPlaceholderDoesNotAccumulate(t *testing.T) {
	got := expandMacros(macroRequest("`outer(`inner(7)`)`", macroDef("outer", "outer", "$x$+$x$", "x"), macroDef("inner", "inner", "$y$", "y")))
	if got.Text != "7+7" || len(got.Gaps) != 0 {
		t.Fatalf("repeated nested text: %+v", got)
	}
	var substituted int
	for _, seg := range got.Segments {
		if seg.Source.Kind != "query" {
			continue
		}
		substituted++
		if len(seg.Placeholders) != 2 || seg.Placeholders[0].ObjectID != "outer" || seg.Placeholders[1].ObjectID != "inner" {
			t.Fatalf("placeholder ancestry accumulated: %+v", seg.Placeholders)
		}
	}
	if substituted != 2 {
		t.Fatalf("substituted segments=%d", substituted)
	}
}
