package analysis

import (
	"fmt"
	"strings"
)

type requirementTrace struct {
	references              []requirementTraceReference
	diagnostics             []requirementTraceDiagnostic
	pendingReferenceIndexes map[string]int
	incompleteStageIDs      map[string]struct{}
	syntaxComplete          bool
	semanticComplete        bool
	nextOrdinal             int
}

type requirementTraceReference struct {
	pendingID      string
	reference      Reference
	directExternal bool
	conditional    bool
	eventOrdinal   int
}

type requirementTraceDiagnostic struct {
	diagnostic          Diagnostic
	incomplete          bool
	pendingReferenceIDs []string
	eventOrdinal        int
}

type requirementField struct {
	source      bool
	unavailable bool
	conditional bool
	origins     []string
}

type requirementEnvironment struct {
	trace     *requirementTrace
	fields    map[string]requirementField
	removed   map[string]bool
	open      bool
	uncertain bool
}

func newRequirementTrace() *requirementTrace {
	return &requirementTrace{
		references:              []requirementTraceReference{},
		diagnostics:             []requirementTraceDiagnostic{},
		pendingReferenceIndexes: map[string]int{},
		incompleteStageIDs:      map[string]struct{}{},
		syntaxComplete:          true,
		semanticComplete:        true,
	}
}

func (t *requirementTrace) nextEvent() int {
	ordinal := t.nextOrdinal
	t.nextOrdinal++
	return ordinal
}

func (t *requirementTrace) recordReference(reference Reference, directExternal, conditional bool, eventOrdinal int) {
	index := len(t.references)
	t.references = append(t.references, requirementTraceReference{
		pendingID:      reference.ID,
		reference:      cloneTraceReference(reference),
		directExternal: directExternal,
		conditional:    conditional,
		eventOrdinal:   eventOrdinal,
	})
	t.pendingReferenceIndexes[reference.ID] = index
}

func (t *requirementTrace) recordDiagnostic(diagnostic Diagnostic, incomplete bool, pendingReferenceIDs []string, eventOrdinal int) {
	if incomplete {
		t.semanticComplete = false
		if diagnostic.StageID != "" {
			t.incompleteStageIDs[diagnostic.StageID] = struct{}{}
		}
	}
	t.diagnostics = append(t.diagnostics, requirementTraceDiagnostic{
		diagnostic:          diagnostic,
		incomplete:          incomplete,
		pendingReferenceIDs: append([]string{}, pendingReferenceIDs...),
		eventOrdinal:        eventOrdinal,
	})
}

func (t *requirementTrace) remapReferences(mapping map[string]string) {
	seen := map[string]bool{}
	for i := range t.references {
		entry := &t.references[i]
		if seen[entry.pendingID] {
			panic(fmt.Sprintf("requirement trace pending reference %q was recorded more than once", entry.pendingID))
		}
		seen[entry.pendingID] = true
		id, ok := mapping[entry.pendingID]
		if !ok {
			panic(fmt.Sprintf("requirement trace pending reference %q was not finalized", entry.pendingID))
		}
		entry.reference.ID = id
		remapTraceIDs(entry.reference.OriginReferenceIDs, mapping)
	}
	for i := range t.diagnostics {
		remapTraceIDs(t.diagnostics[i].pendingReferenceIDs, mapping)
	}
}

func (t *requirementTrace) reference(pendingID string) *requirementTraceReference {
	index, ok := t.pendingReferenceIndexes[pendingID]
	if ok && index >= 0 && index < len(t.references) && t.references[index].pendingID == pendingID {
		return &t.references[index]
	}
	panic(fmt.Sprintf("requirement trace reference %q is missing", pendingID))
}

func (t *requirementTrace) syncParserDiagnostics(diagnostics []Diagnostic) {
	for i, diagnostic := range diagnostics {
		if i >= len(t.diagnostics) {
			break
		}
		t.diagnostics[i].diagnostic = diagnostic
	}
	t.rebuildIncompleteStageIndex()
}

func (t *requirementTrace) remapStages(mapping map[string]string) {
	for i := range t.references {
		if id := t.references[i].reference.StageID; id != "" {
			t.references[i].reference.StageID = mapping[id]
		}
	}
	for i := range t.diagnostics {
		if id := t.diagnostics[i].diagnostic.StageID; id != "" {
			t.diagnostics[i].diagnostic.StageID = mapping[id]
		}
	}
	t.rebuildIncompleteStageIndex()
}

func (t *requirementTrace) rebuildIncompleteStageIndex() {
	t.incompleteStageIDs = map[string]struct{}{}
	for _, diagnostic := range t.diagnostics {
		if diagnostic.incomplete && diagnostic.diagnostic.StageID != "" {
			t.incompleteStageIDs[diagnostic.diagnostic.StageID] = struct{}{}
		}
	}
}

func (t *requirementTrace) assertReferences(public []Reference) {
	if len(t.references) != len(public) {
		panic(fmt.Sprintf("requirement trace has %d references for %d public references", len(t.references), len(public)))
	}
	byID := make(map[string]Reference, len(public))
	for _, ref := range public {
		byID[ref.ID] = ref
	}
	for _, entry := range t.references {
		got, ok := byID[entry.reference.ID]
		if !ok {
			panic(fmt.Sprintf("requirement trace reference %q has no public counterpart", entry.reference.ID))
		}
		want := entry.reference
		if got.NormalizedName != want.NormalizedName || got.OriginalName != want.OriginalName || got.Kind != want.Kind || got.Role != want.Role || got.StageID != want.StageID || got.ScopeID != want.ScopeID || got.Location != want.Location || got.Resolution != want.Resolution {
			panic(fmt.Sprintf("requirement trace reference %q differs from its public counterpart: trace=%+v public=%+v", entry.reference.ID, want, got))
		}
	}
}

func cloneTraceReference(reference Reference) Reference {
	reference.OriginReferenceIDs = append([]string{}, reference.OriginReferenceIDs...)
	return reference
}

func remapTraceIDs(ids []string, mapping map[string]string) {
	for i, id := range ids {
		replacement, ok := mapping[id]
		if !ok {
			panic(fmt.Sprintf("requirement trace link %q was not finalized", id))
		}
		ids[i] = replacement
	}
}

func traceOrigins(trace *requirementTrace, ids []string) []string {
	out := append([]string{}, ids...)
	for _, id := range ids {
		entry := trace.reference(id)
		out = uniqueIDs(out, entry.reference.OriginReferenceIDs)
	}
	return out
}

func newRequirementEnvironment(trace *requirementTrace) requirementEnvironment {
	return requirementEnvironment{
		trace:   trace,
		fields:  map[string]requirementField{},
		removed: map[string]bool{},
		open:    true,
	}
}

func (e requirementEnvironment) clone() requirementEnvironment {
	out := newRequirementEnvironment(e.trace)
	out.open = e.open
	out.uncertain = e.uncertain
	for name, field := range e.fields {
		field.origins = append([]string{}, field.origins...)
		out.fields[name] = field
	}
	for name, removed := range e.removed {
		out.removed[name] = removed
	}
	return out
}

func (e *requirementEnvironment) read(reference Reference) (binding string, directExternal, conditional bool) {
	if reference.Role == "remove" || reference.Role == "create" || reference.Role == "output" || reference.Role == "rename" {
		return "not_applicable", false, false
	}
	nullTest := reference.Role == "null_test"
	classified := func(binding string, directExternal, conditional bool) (string, bool, bool) {
		if nullTest {
			return binding, false, false
		}
		return binding, directExternal, conditional
	}
	name := reference.NormalizedName
	if reference.Resolution == "wildcard" || reference.Resolution == "dynamic" {
		return classified("indeterminate", false, true)
	}
	field, known := e.fields[name]
	switch {
	case known && field.conditional:
		return classified("indeterminate", false, true)
	case known:
		if field.source {
			return classified("source", true, false)
		}
		return classified("derived", false, false)
	case e.uncertain:
		return classified("indeterminate", false, true)
	case e.removed[name] || !e.open:
		return classified("unavailable", false, false)
	default:
		if !nullTest {
			e.fields[name] = requirementField{source: true, origins: []string{reference.ID}}
		}
		return classified("source", true, false)
	}
}

func requirementReferencePolicy(reference Reference) (directExternal, conditional bool) {
	if reference.Role != "read" {
		return false, false
	}
	switch reference.Kind {
	case "index", "source", "sourcetype", "dataset", "data_model", "lookup", "macro":
		if reference.Resolution == "exact" {
			return true, false
		}
		return false, true
	default:
		return false, false
	}
}

func (e *requirementEnvironment) install(name string, origins []string, conditional bool) {
	e.fields[name] = requirementField{conditional: conditional, origins: append([]string{}, origins...)}
	delete(e.removed, name)
}

func (e *requirementEnvironment) markConditional(name string) {
	field, known := e.fields[name]
	if !known {
		return
	}
	field.conditional = true
	e.fields[name] = field
}

func (e *requirementEnvironment) remove(name string) {
	delete(e.fields, name)
	e.removed[name] = true
}

func (e *requirementEnvironment) stageIncomplete(stageID string) bool {
	if e.trace == nil {
		return false
	}
	_, incomplete := e.trace.incompleteStageIDs[stageID]
	return incomplete
}

func (e *requirementEnvironment) applyProjection(selectors []locatedOperand, mode string, retainKnownInternals bool, stageID string) {
	if mode == "exclude" {
		for _, selector := range selectors {
			if selector.Resolution != "wildcard" {
				continue
			}
			for name := range e.fields {
				if wildcardMatches(selector.Name, name) {
					e.remove(name)
				}
			}
		}
		return
	}

	selected := map[string]requirementField{}
	if retainKnownInternals {
		for name, field := range e.fields {
			if strings.HasPrefix(name, "_") {
				selected[name] = field
			}
		}
	}
	for _, selector := range selectors {
		if selector.Resolution == "wildcard" {
			for name, field := range e.fields {
				if wildcardMatches(selector.Name, name) {
					selected[name] = field
				}
			}
			continue
		}
		if field, ok := e.fields[selector.Name]; ok {
			selected[selector.Name] = field
		}
	}
	e.fields = selected
	e.open = false
	if mode == "table" && !e.stageIncomplete(stageID) {
		e.uncertain = false
	}
}
