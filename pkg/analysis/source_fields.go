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

// SourceFieldAdmission describes external source-name membership, not event presence.
type SourceFieldAdmission uint8

const (
	SourceFieldIndeterminate SourceFieldAdmission = iota
	SourceFieldAdmitted
	SourceFieldProhibited
)

// SourceUniverse supplies candidate concrete names and optional admission rules.
// Complete promises every potentially admitted name is listed in Fields. Names
// are copied, sorted, case-sensitive, and must be unique nonblank UTF-8 strings.
// Resolve must be pure, deterministic, concurrency-safe, and independent of query
// text, pipeline state, and event presence. It may be called repeatedly. Callers
// must not mutate captured target data during analysis; callback state is not
// copied and callback panics are not recovered. With nil Resolve, listed names
// are admitted; unlisted names are prohibited when complete and indeterminate
// otherwise. A complete universe never calls Resolve for unlisted names.
// Unknown admission values are treated as indeterminate.
type SourceUniverse struct {
	Fields   []string
	Complete bool
	Resolve  func(name string) SourceFieldAdmission
}

type sourceRefinement struct {
	complete            bool
	resolve             func(string) SourceFieldAdmission
	finiteCompatibility bool
	names               []string
	members             map[string]bool
	expansions          []FieldExpansion
}

// AnalyzeWithSourceFields analyzes with a known finite source universe. Names are
// concrete and case-sensitive; nil and empty both denote a known empty universe.
func AnalyzeWithSourceFields(document QueryDocument, fields []string) (*SourceAnalysis, error) {
	return analyzeSourceUniverse(document, SourceUniverse{Fields: fields, Complete: true}, true)
}

// AnalyzeWithSourceUniverse retains proven partial selector evidence while
// distinguishing unresolved source membership from structural field availability.
func AnalyzeWithSourceUniverse(document QueryDocument, universe SourceUniverse) (*SourceAnalysis, error) {
	return analyzeSourceUniverse(document, universe, false)
}

func analyzeSourceUniverse(document QueryDocument, universe SourceUniverse, finiteCompatibility bool) (*SourceAnalysis, error) {
	r := &sourceRefinement{names: append([]string{}, universe.Fields...), members: map[string]bool{}, expansions: []FieldExpansion{}, complete: universe.Complete, resolve: universe.Resolve, finiteCompatibility: finiteCompatibility}
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

func (r *sourceRefinement) admission(name string) SourceFieldAdmission {
	listed := r.members[name]
	if !listed && r.complete {
		return SourceFieldProhibited
	}
	if r.resolve != nil {
		result := r.resolve(name)
		if result == SourceFieldAdmitted || result == SourceFieldProhibited {
			return result
		}
		return SourceFieldIndeterminate
	}
	if listed {
		return SourceFieldAdmitted
	}
	return SourceFieldIndeterminate
}

// Tracked obligations control local exhaustiveness even when admission is
// unresolved. Conditional provenance shadows declarations and is never promoted.
func (s *semanticStage) refinedSelectorCandidates(pattern string) ([]string, map[string]trackedField, bool) {
	bindings := map[string]trackedField{}
	exhaustive := !s.env.open || s.refinement.complete
	add := func(name string, field trackedField) {
		if !wildcardMatches(pattern, name) {
			return
		}
		if field.source && !field.Conditional {
			switch s.refinement.admission(name) {
			case SourceFieldProhibited:
				return
			case SourceFieldIndeterminate:
				exhaustive = false
			}
		}
		bindings[name] = field
	}
	for name, field := range s.env.fields {
		add(name, field)
	}
	if s.env.open {
		for _, name := range s.refinement.names {
			if _, known := s.env.fields[name]; known || s.env.removed[name] {
				continue
			}
			// Prohibited declarations do not become fields, including after unknown stages.
			if s.refinement.admission(name) == SourceFieldProhibited {
				continue
			}
			add(name, trackedField{FieldBinding: FieldBinding{Name: name, OriginReferenceIDs: []string{}, Conditional: s.env.uncertain}, source: true})
		}
	}
	names := make([]string, 0, len(bindings))
	for name := range bindings {
		names = append(names, name)
	}
	sort.Strings(names)
	return names, bindings, exhaustive
}

func (s *semanticStage) provenExpandedFields(names []string, bindings map[string]trackedField) []ExpandedField {
	proven := []string{}
	for _, name := range names {
		field := bindings[name]
		if !field.Conditional && (!field.source || s.refinement.admission(name) == SourceFieldAdmitted) {
			proven = append(proven, name)
		}
	}
	return sortedExpandedFields(proven, bindings)
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
	if !complete && s.refinement.finiteCompatibility {
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
	if !s.refinement.complete {
		return false
	}
	removedMatch := false
	for _, name := range s.refinement.names {
		if !wildcardMatches(pattern, name) {
			continue
		}
		admission := s.refinement.admission(name)
		if admission == SourceFieldIndeterminate {
			return false
		}
		if admission == SourceFieldProhibited {
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
	matches := s.provenExpandedFields(names, bindings)
	if !complete {
		s.recordExpansion(id, false, matches)
		s.diagnostic(CodeUnresolvedWildcard, fmt.Sprintf("wildcard %q membership is unresolved", pattern), c)
	} else {
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

func (s *semanticStage) retainSourceInternals() bool {
	if !s.env.open {
		return true
	}
	complete := s.refinement.complete
	for _, name := range s.refinement.names {
		if _, known := s.env.fields[name]; known || s.env.removed[name] || !strings.HasPrefix(name, "_") {
			continue
		}
		admission := s.refinement.admission(name)
		if admission == SourceFieldProhibited {
			continue
		}
		if admission == SourceFieldIndeterminate {
			complete = false
		}
		s.env.fields[name] = trackedField{FieldBinding: FieldBinding{Name: name, OriginReferenceIDs: []string{}, Conditional: s.env.uncertain}, source: true}
	}
	return complete
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
