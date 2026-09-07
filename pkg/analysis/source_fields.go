package analysis

import (
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/delgado-jacob/spl-toolkit/parser"
)

// ExpandedField is a proven member of a selector's finite field universe.
type ExpandedField struct {
	Name    string `json:"name"`
	Binding string `json:"binding"`
}

// FieldExpansion distinguishes conclusive empty membership from unresolved membership.
type FieldExpansion struct {
	ReferenceID string          `json:"reference_id"`
	Complete    bool            `json:"complete"`
	Matches     []ExpandedField `json:"matches"`
}

// SourceAnalysis adds finite-source evidence without changing the analysis wire format.
type SourceAnalysis struct {
	Result     *Result          `json:"analysis"`
	Expansions []FieldExpansion `json:"expansions"`
}

type sourceRefinement struct {
	names      []string
	members    map[string]bool
	expansions []FieldExpansion
}

// AnalyzeWithSourceFields analyzes with a known finite source universe. Names are
// concrete and case-sensitive; nil and empty both denote a known empty universe.
func AnalyzeWithSourceFields(document QueryDocument, fields []string) (*SourceAnalysis, error) {
	r := &sourceRefinement{names: append([]string{}, fields...), members: map[string]bool{}, expansions: []FieldExpansion{}}
	for _, name := range r.names {
		if !utf8.ValidString(name) || strings.TrimSpace(name) == "" {
			return nil, fmt.Errorf("source field name must be nonempty, non-whitespace UTF-8")
		}
		if r.members[name] {
			return nil, fmt.Errorf("duplicate source field name %q", name)
		}
		r.members[name] = true
	}
	sort.Strings(r.names)
	result, err := analyze(document, r)
	if err != nil {
		return nil, err
	}
	return &SourceAnalysis{Result: result, Expansions: r.expansions}, nil
}

// Known conditional bindings shadow catalog declarations, just as derived ones
// do. An exact source obligation absent from the catalog is not a proven member.
func (s *semanticStage) refinedSelectorCandidates(pattern string) ([]string, map[string]trackedField, bool) {
	bindings := map[string]trackedField{}
	for name, field := range s.env.fields {
		if wildcardMatches(pattern, name) && (field.Conditional || !field.source || s.refinement.members[name]) {
			bindings[name] = field
		}
	}
	if s.env.open {
		for _, name := range s.refinement.names {
			if _, known := s.env.fields[name]; known || s.env.removed[name] || !wildcardMatches(pattern, name) {
				continue
			}
			bindings[name] = trackedField{FieldBinding: FieldBinding{Name: name, OriginReferenceIDs: []string{}, Conditional: s.env.uncertain}, source: true}
		}
	}
	names := make([]string, 0, len(bindings))
	for name := range bindings {
		names = append(names, name)
	}
	sort.Strings(names)
	return names, bindings, true
}

func allBindingsProven(bindings map[string]trackedField) bool {
	for _, field := range bindings {
		if field.Conditional {
			return false
		}
	}
	return true
}

func sortedExpandedFields(names []string, bindings map[string]trackedField) []ExpandedField {
	matches := make([]ExpandedField, 0, len(names))
	for _, name := range names {
		binding := "derived"
		if bindings[name].source {
			binding = "source"
		}
		matches = append(matches, ExpandedField{Name: name, Binding: binding})
	}
	return matches
}

func (s *semanticStage) recordExpansion(id string, complete bool, matches []ExpandedField) {
	if !complete {
		matches = nil
	}
	s.refinement.expansions = append(s.refinement.expansions, FieldExpansion{ReferenceID: id, Complete: complete, Matches: append([]ExpandedField{}, matches...)})
}

// Structural absence must be proven before filtering exact source obligations
// by catalog membership. A closed set can exclude a pattern outright; an open
// set can prove only the removal of all matching catalog declarations.
func (s *semanticStage) selectorStructurallyAbsent(pattern string) bool {
	for name := range s.env.fields {
		if wildcardMatches(pattern, name) {
			return false
		}
	}
	if !s.env.open {
		return true
	}
	removedMatch := false
	for _, name := range s.refinement.names {
		if !wildcardMatches(pattern, name) {
			continue
		}
		if !s.env.removed[name] {
			return false
		}
		removedMatch = true
	}
	return removedMatch
}

func (s *semanticStage) refinedSelector(c parser.IAnalysisSelectorContext, role, id string) []string {
	pattern := selectorName(c)
	// Removal operates on tracked obligations too, even when external validation
	// would find those names absent. These are transfer candidates, not evidence.
	removals := map[string]bool{}
	if role == "remove" {
		for name := range s.env.fields {
			if wildcardMatches(pattern, name) {
				removals[name] = true
			}
		}
	}
	names, bindings, exhaustive := s.refinedSelectorCandidates(pattern)
	complete := exhaustive && !s.env.uncertain && allBindingsProven(bindings)
	ref := &s.result.References[len(s.result.References)-1]
	if !complete {
		s.recordExpansion(id, false, nil)
		s.diagnostic(CodeUnresolvedWildcard, fmt.Sprintf("wildcard %q membership is unresolved", pattern), c)
	} else {
		matches := sortedExpandedFields(names, bindings)
		s.recordExpansion(id, true, matches)
		if role != "remove" {
			ref.Binding = "source"
			if len(matches) == 0 && s.selectorStructurallyAbsent(pattern) {
				ref.Binding = "unavailable"
				s.diagnostic(CodeUnavailableField, fmt.Sprintf("wildcard %q is unavailable after an earlier pipeline transfer", pattern), c)
			} else if len(matches) > 0 {
				ref.Binding = matches[0].Binding
				for _, match := range matches[1:] {
					if match.Binding != ref.Binding {
						ref.Binding = "indeterminate"
						break
					}
				}
			}
		}
	}
	origins := []string{}
	for _, name := range names {
		field := bindings[name]
		if role == "remove" {
			removals[name] = true
		} else {
			if _, known := s.env.fields[name]; !known {
				field.OriginReferenceIDs = []string{id}
				s.env.fields[name] = field
			}
		}
		// Newly materialized members originate at this selector; avoid a self-link
		// on the selector itself while preserving that provenance downstream.
		origins = uniqueIDs(origins, bindings[name].OriginReferenceIDs)
	}
	sort.Strings(origins)
	ref.OriginReferenceIDs = origins
	if role == "remove" {
		names = make([]string, 0, len(removals))
		for name := range removals {
			names = append(names, name)
		}
		sort.Strings(names)
	}
	return names
}

func (s *semanticStage) retainSourceInternals() {
	if !s.env.open {
		return
	}
	for _, name := range s.refinement.names {
		if _, known := s.env.fields[name]; known || s.env.removed[name] || !strings.HasPrefix(name, "_") {
			continue
		}
		s.env.fields[name] = trackedField{FieldBinding: FieldBinding{Name: name, OriginReferenceIDs: []string{}, Conditional: s.env.uncertain}, source: true}
	}
}

func (r *sourceRefinement) finalizeExpansions(references []Reference, mapping map[string]string) {
	order := map[string]int{}
	for i, ref := range references {
		order[ref.ID] = i
	}
	for i := range r.expansions {
		r.expansions[i].ReferenceID = mapping[r.expansions[i].ReferenceID]
	}
	sort.SliceStable(r.expansions, func(i, j int) bool { return order[r.expansions[i].ReferenceID] < order[r.expansions[j].ReferenceID] })
}
