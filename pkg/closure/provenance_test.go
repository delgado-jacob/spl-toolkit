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
