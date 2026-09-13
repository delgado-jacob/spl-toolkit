package lsp

import "testing"

func TestUTF16Positions(t *testing.T) {
	source := "a😀b\r\nc\rd\né"
	for _, tc := range []struct{ offset, line, char int }{{0, 0, 0}, {1, 0, 1}, {5, 0, 3}, {6, 0, 4}, {8, 1, 0}, {9, 1, 1}, {10, 2, 0}, {12, 3, 0}, {14, 3, 1}} {
		got, err := positionAt(source, tc.offset)
		if err != nil || got != (Position{tc.line, tc.char}) {
			t.Fatalf("offset %d: %+v %v", tc.offset, got, err)
		}
		offset, err := offsetAt(source, got)
		if err != nil || offset != tc.offset {
			t.Fatalf("position %+v: %d %v", got, offset, err)
		}
	}
	if n, err := offsetAt(source, Position{0, 99}); err != nil || n != 6 {
		t.Fatalf("line clamp: %d %v", n, err)
	}
	for _, p := range []Position{{-1, 0}, {0, -1}, {0, 2}, {4, 0}} {
		if _, err := offsetAt(source, p); err == nil {
			t.Fatalf("accepted %+v", p)
		}
	}
	for _, n := range []int{-1, 2, 3, 4, 15} {
		if _, err := positionAt(source, n); err == nil {
			t.Fatalf("accepted offset %d", n)
		}
	}
}
