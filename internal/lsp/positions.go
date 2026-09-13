package lsp

import (
	"fmt"
	"unicode/utf8"
)

// Position uses zero-based UTF-16 code units, never terminal display widths.
type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}
type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

func positionAt(text string, offset int) (Position, error) {
	p := Position{}
	if offset < 0 || offset > len(text) || !utf8.ValidString(text) || (offset < len(text) && !utf8.RuneStart(text[offset])) {
		return p, fmt.Errorf("invalid UTF-8 byte offset")
	}
	for i := 0; i < offset; {
		if text[i] == '\r' || text[i] == '\n' {
			if text[i] == '\r' && i+1 < len(text) && text[i+1] == '\n' {
				if offset == i+1 {
					return p, nil
				}
				i++
			}
			p.Line++
			p.Character = 0
			i++
			continue
		}
		r, n := utf8.DecodeRuneInString(text[i:])
		p.Character++
		if r > 0xffff {
			p.Character++
		}
		i += n
	}
	return p, nil
}
func offsetAt(text string, p Position) (int, error) {
	if p.Line < 0 || p.Character < 0 || !utf8.ValidString(text) {
		return 0, fmt.Errorf("invalid UTF-16 position")
	}
	i, line := 0, 0
	for line < p.Line && i < len(text) {
		if text[i] == '\r' || text[i] == '\n' {
			if text[i] == '\r' && i+1 < len(text) && text[i+1] == '\n' {
				i++
			}
			line++
		}
		i++
	}
	if line != p.Line {
		return 0, fmt.Errorf("line outside document")
	}
	units := 0
	for i < len(text) && text[i] != '\r' && text[i] != '\n' && units < p.Character {
		r, n := utf8.DecodeRuneInString(text[i:])
		u := 1
		if r > 0xffff {
			u = 2
		}
		if units+u > p.Character {
			return 0, fmt.Errorf("position splits surrogate pair")
		}
		units += u
		i += n
	}
	return i, nil
}
