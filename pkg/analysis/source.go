package analysis

import "github.com/antlr4-go/antlr/v4"

// sourceIndex converts ANTLR's rune offsets to the public UTF-8 byte positions.
type sourceIndex struct {
	text      string
	positions []Position
}

func newSourceIndex(text string) *sourceIndex {
	s := &sourceIndex{text: text, positions: make([]Position, 0, len(text)+1)}
	line, column := 1, 1
	previous := rune(0)
	for offset, r := range text {
		s.positions = append(s.positions, Position{Offset: offset, Line: line, Column: column})
		switch r {
		case '\r':
			line++
			column = 1
		case '\n':
			if previous != '\r' {
				line++
			}
			column = 1
		default:
			column++
		}
		previous = r
	}
	s.positions = append(s.positions, Position{Offset: len(text), Line: line, Column: column})
	return s
}
func (s *sourceIndex) position(runeOffset int) Position {
	if runeOffset < 0 {
		runeOffset = 0
	}
	if runeOffset >= len(s.positions) {
		runeOffset = len(s.positions) - 1
	}
	return s.positions[runeOffset]
}
func (s *sourceIndex) location(start, end int) Location {
	if end < start {
		end = start
	}
	return Location{Start: s.position(start), End: s.position(end)}
}
func (s *sourceIndex) contextLocation(ctx antlr.ParserRuleContext) Location {
	start, end := 0, 0
	if ctx.GetStart() != nil {
		start = ctx.GetStart().GetStart()
		end = start
	}
	if ctx.GetStop() != nil && ctx.GetStop().GetStop() >= start {
		end = ctx.GetStop().GetStop() + 1
	}
	return s.location(start, end)
}
