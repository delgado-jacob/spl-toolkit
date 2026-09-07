package analysis

import "testing"

// Rune offsets must not be used as byte offsets, including after Unicode and CRLF.
func TestSourceUnicodeAndCRLF(t *testing.T) {
	s := newSourceIndex("  search host=\"é\"\n| eval 'café'=1  ")
	got := s.location(25, 31)
	want := Location{Start: Position{Offset: 26, Line: 2, Column: 8}, End: Position{Offset: 33, Line: 2, Column: 14}}
	if got != want {
		t.Fatalf("got %+v want %+v", got, want)
	}
	s = newSourceIndex("é\r\n\tx")
	for _, tc := range []struct {
		r int
		p Position
	}{{0, Position{0, 1, 1}}, {3, Position{4, 2, 1}}, {4, Position{5, 2, 2}}, {5, Position{6, 2, 3}}} {
		if p := s.position(tc.r); p != tc.p {
			t.Fatalf("%d: %+v != %+v", tc.r, p, tc.p)
		}
	}
}
