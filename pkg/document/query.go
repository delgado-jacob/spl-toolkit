package document

import (
	"fmt"
	"unicode/utf8"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
)

// Scope returns a detached scope with id from the current snapshot contents.
func (s *Snapshot) Scope(id string) (analysis.Scope, bool) {
	if s == nil {
		return analysis.Scope{}, false
	}
	for _, scope := range s.Scopes {
		if scope.ID == id {
			return scope, true
		}
	}
	return analysis.Scope{}, false
}

// Stage returns a detached stage with id from the current snapshot contents.
func (s *Snapshot) Stage(id string) (analysis.Stage, bool) {
	if s == nil {
		return analysis.Stage{}, false
	}
	for _, stage := range s.Stages {
		if stage.ID == id {
			return stage, true
		}
	}
	return analysis.Stage{}, false
}

// Reference returns a detached reference with id from the current snapshot contents.
func (s *Snapshot) Reference(id string) (analysis.Reference, bool) {
	if s == nil {
		return analysis.Reference{}, false
	}
	for _, reference := range s.References {
		if reference.ID == id {
			return cloneReferences([]analysis.Reference{reference})[0], true
		}
	}
	return analysis.Reference{}, false
}

// Source returns the exact original source text in a checked half-open byte range.
func (s *Snapshot) Source(startByte, endByte int) (string, error) {
	if s == nil {
		return "", fmt.Errorf("snapshot is required")
	}
	if err := s.checkRange(startByte, endByte); err != nil {
		return "", err
	}
	return s.Document.Text[startByte:endByte], nil
}

// ReferencesIn returns nonempty references that intersect r in canonical order.
// A zero-width query selects nothing; use ReferencesAt for an exact position.
func (s *Snapshot) ReferencesIn(startByte, endByte int) ([]analysis.Reference, error) {
	if s == nil {
		return nil, fmt.Errorf("snapshot is required")
	}
	if err := s.checkRange(startByte, endByte); err != nil {
		return nil, err
	}
	if startByte == endByte {
		return []analysis.Reference{}, nil
	}
	matches := []analysis.Reference{}
	for _, reference := range s.References {
		referenceStart, referenceEnd := reference.Location.Start.Offset, reference.Location.End.Offset
		if err := s.checkRange(referenceStart, referenceEnd); err != nil {
			return nil, err
		}
		if referenceStart == referenceEnd {
			continue
		}
		if referenceStart < endByte && startByte < referenceEnd {
			matches = append(matches, cloneReferences([]analysis.Reference{reference})[0])
		}
	}
	return matches, nil
}

// ReferencesAt returns nonempty references containing position and zero-width
// references located exactly at position, in canonical reference order.
func (s *Snapshot) ReferencesAt(position int) ([]analysis.Reference, error) {
	if s == nil {
		return nil, fmt.Errorf("snapshot is required")
	}
	if err := s.checkRange(position, position); err != nil {
		return nil, err
	}
	matches := []analysis.Reference{}
	for _, reference := range s.References {
		referenceStart, referenceEnd := reference.Location.Start.Offset, reference.Location.End.Offset
		if err := s.checkRange(referenceStart, referenceEnd); err != nil {
			return nil, err
		}
		if (referenceStart < referenceEnd && referenceStart <= position && position < referenceEnd) ||
			(referenceStart == referenceEnd && referenceStart == position) {
			matches = append(matches, cloneReferences([]analysis.Reference{reference})[0])
		}
	}
	return matches, nil
}

// OriginReferences returns the canonical direct origin references of id. It
// never infers transitive equivalence from spelling, scope, or source position.
func (s *Snapshot) OriginReferences(id string) ([]analysis.Reference, bool) {
	reference, found := s.Reference(id)
	if !found {
		return nil, false
	}
	origins := []analysis.Reference{}
	for _, originID := range reference.OriginReferenceIDs {
		if origin, found := s.Reference(originID); found {
			origins = append(origins, origin)
		}
	}
	return origins, true
}

// LineageForStage returns detached canonical lineage entries for stageID in
// stored evaluation order. It exposes existing transitions without deriving a
// new semantic flow.
func (s *Snapshot) LineageForStage(stageID string) []analysis.Lineage {
	if s == nil {
		return []analysis.Lineage{}
	}
	lineage := []analysis.Lineage{}
	for _, entry := range s.Lineage {
		if entry.StageID == stageID {
			lineage = append(lineage, cloneLineage([]analysis.Lineage{entry})[0])
		}
	}
	return lineage
}

func (s *Snapshot) checkRange(startByte, endByte int) error {
	text := s.Document.Text
	if startByte < 0 || endByte < startByte || endByte > len(text) || !utf8.ValidString(text) || !runeBoundary(text, startByte) || !runeBoundary(text, endByte) {
		return fmt.Errorf("%w: [%d,%d)", ErrInvalidRange, startByte, endByte)
	}
	return nil
}

func runeBoundary(text string, offset int) bool {
	return offset == 0 || offset == len(text) || (offset > 0 && offset < len(text) && utf8.RuneStart(text[offset]))
}
