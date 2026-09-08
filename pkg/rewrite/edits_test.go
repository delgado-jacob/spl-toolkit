package rewrite

import (
	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"reflect"
	"testing"
)

func byteEdit(start, end int, before, after string) analysis.RewriteTextEdit {
	return analysis.RewriteTextEdit{Location: analysis.Location{Start: analysis.Position{Offset: start}, End: analysis.Position{Offset: end}}, Before: before, After: after}
}
func TestRewriteBytePreservationAndLocations(t *testing.T) {
	// Arbitrary spans intentionally have no reference; only the canonical locator
	// knows Unicode columns and CRLF behavior.
	original := "é\tα\r\n'x|y' src\r\n<EOF>"
	edits := []analysis.RewriteTextEdit{byteEdit(13, 16, "src", "用户\nβ"), byteEdit(3, 5, "α", "名字")}
	got, locs, err := reconstructEdits(original, edits)
	if err != nil {
		t.Fatal(err)
	}
	if want := "é\t名字\r\n'x|y' 用户\nβ\r\n<EOF>"; got != want {
		t.Fatalf("text=%q, want %q", got, want)
	}
	want := []analysis.Location{
		{Start: analysis.Position{Offset: 17, Line: 2, Column: 7}, End: analysis.Position{Offset: 26, Line: 3, Column: 2}},
		{Start: analysis.Position{Offset: 3, Line: 1, Column: 3}, End: analysis.Position{Offset: 9, Line: 1, Column: 5}},
	}
	if !reflect.DeepEqual(locs, want) {
		t.Errorf("locations=%+v, want %+v", locs, want)
	}
	if edits[0].Location.Start.Offset != 13 {
		t.Fatal("mutated caller edits")
	}
}
func TestRewriteByteInsertions(t *testing.T) {
	got, locs, err := reconstructEdits("α src!", []analysis.RewriteTextEdit{byteEdit(7, 7, "", "\r\n終"), byteEdit(3, 6, "src", "x"), byteEdit(3, 3, "", "["), byteEdit(6, 6, "", "]")})
	if err != nil {
		t.Fatal(err)
	}
	if got != "α [x]!\r\n終" {
		t.Fatalf("text=%q", got)
	}
	want := []analysis.Location{
		{Start: analysis.Position{Offset: 7, Line: 1, Column: 7}, End: analysis.Position{Offset: 12, Line: 2, Column: 2}},
		{Start: analysis.Position{Offset: 4, Line: 1, Column: 4}, End: analysis.Position{Offset: 5, Line: 1, Column: 5}},
		{Start: analysis.Position{Offset: 3, Line: 1, Column: 3}, End: analysis.Position{Offset: 4, Line: 1, Column: 4}},
		{Start: analysis.Position{Offset: 5, Line: 1, Column: 5}, End: analysis.Position{Offset: 6, Line: 1, Column: 6}},
	}
	if !reflect.DeepEqual(locs, want) {
		t.Errorf("locations=%+v, want %+v", locs, want)
	}
}
func TestRewriteByteRejectsInvalidEdits(t *testing.T) {
	for name, edits := range map[string][]analysis.RewriteTextEdit{
		"old text mismatch":    {byteEdit(0, 2, "x", "y")},
		"negative":             {byteEdit(-1, 0, "", "y")},
		"past end":             {byteEdit(0, 20, "", "y")},
		"reversed":             {byteEdit(3, 2, "", "y")},
		"split rune":           {byteEdit(1, 2, "\xa9", "x")},
		"invalid replacement":  {byteEdit(0, 2, "é", "\xff")},
		"overlap":              {byteEdit(0, 3, "éx", "a"), byteEdit(2, 3, "x", "b")},
		"interior insertion":   {byteEdit(0, 3, "éx", "a"), byteEdit(2, 2, "", "b")},
		"competing insertions": {byteEdit(2, 2, "", "a"), byteEdit(2, 2, "", "b")},
	} {
		t.Run(name, func(t *testing.T) {
			if _, _, err := reconstructEdits("éx", edits); err == nil {
				t.Fatal("invalid interval accepted")
			}
		})
	}
}
func TestRewriteByteUnchanged(t *testing.T) {
	const original = "search x=1 ``` é | comment ```\r\n<EOF>"
	got, locs, err := reconstructEdits(original, nil)
	if err != nil || got != original || locs == nil || len(locs) != 0 {
		t.Fatalf("unchanged bytes lost: %q, %+v, %v", got, locs, err)
	}
}

func TestRewriteByteCanonicalAuditLocations(t *testing.T) {
	got, _ := candidateForTest(t, analysis.QueryDocument{Language: "spl2", Text: "FROM main\r\n| where 'pré'=1 /* é|src */\r\n| table 'pré'"}, []Rule{{ID: "unicode", Kind: "field", Source: conditionIdentity("pré"), Target: conditionIdentity("用\n户")}})
	if len(got.Changes) != 2 {
		t.Fatalf("changes=%+v", got.Changes)
	}
	want := []analysis.Location{
		{Start: analysis.Position{Offset: 19, Line: 2, Column: 9}, End: analysis.Position{Offset: 29, Line: 2, Column: 15}},
		{Start: analysis.Position{Offset: 54, Line: 3, Column: 9}, End: analysis.Position{Offset: 64, Line: 3, Column: 15}},
	}
	for i, c := range got.Changes {
		if c.CandidateLocation == nil || *c.CandidateLocation != want[i] {
			t.Errorf("candidate location=%+v,want %+v", c.CandidateLocation, want[i])
		}
		if c.NewText != "'用\\n户'" || len(c.CandidateReferenceIDs) != 1 || len(c.OriginalReferenceIDs) != 1 {
			t.Errorf("quoted candidate audit=%+v", c)
		}
	}
}
