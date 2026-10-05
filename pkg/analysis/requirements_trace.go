package analysis

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

type requirementTrace struct {
	inputs                  []inputFact
	correlations            []correlationEvent
	sourceEvidenceBudget    *sourceEvidenceWorkBudget
	references              []requirementTraceReference
	diagnostics             []requirementTraceDiagnostic
	pendingReferenceIndexes map[string]int
	incompleteStageIDs      map[string]struct{}
	syntaxComplete          bool
	semanticComplete        bool
	nextOrdinal             int
}

type requirementTraceReference struct {
	pendingID       string
	reference       Reference
	fieldIdentity   fieldIdentity
	owners          []sourceOwner
	directExternal  bool
	conditional     bool
	pathConditional bool
	eventOrdinal    int
}

type requirementTraceDiagnostic struct {
	diagnostic          Diagnostic
	incomplete          bool
	pendingReferenceIDs []string
	eventOrdinal        int
}

type requirementTracePath struct {
	Ordinal   int
	Trace     *requirementTrace
	Reachable bool
}

type requirementField struct {
	identity       fieldIdentity
	owners         []sourceOwner
	ownerCollision bool // Simultaneous alternatives, rather than expression origins.
	source         bool
	unavailable    bool
	conditional    bool
	origins        []string
}

type requirementEnvironment struct {
	trace     *requirementTrace
	inputs    []inputFact
	fields    map[fieldIdentityKey]requirementField
	removed   map[fieldIdentityKey]bool
	ambiguous map[string]bool
	open      bool
	uncertain bool
}

func newRequirementTrace() *requirementTrace {
	return &requirementTrace{
		sourceEvidenceBudget:    &sourceEvidenceWorkBudget{},
		correlations:            []correlationEvent{},
		references:              []requirementTraceReference{},
		diagnostics:             []requirementTraceDiagnostic{},
		pendingReferenceIndexes: map[string]int{},
		incompleteStageIDs:      map[string]struct{}{},
		syntaxComplete:          true,
		semanticComplete:        true,
	}
}

func (t *requirementTrace) clone() *requirementTrace {
	if t == nil {
		return nil
	}
	out := &requirementTrace{
		inputs:                  cloneInputFacts(t.inputs),
		correlations:            cloneCorrelationEvents(t.correlations),
		sourceEvidenceBudget:    t.sourceEvidenceBudget,
		references:              make([]requirementTraceReference, len(t.references)),
		diagnostics:             make([]requirementTraceDiagnostic, len(t.diagnostics)),
		pendingReferenceIndexes: make(map[string]int, len(t.pendingReferenceIndexes)),
		incompleteStageIDs:      make(map[string]struct{}, len(t.incompleteStageIDs)),
		syntaxComplete:          t.syntaxComplete,
		semanticComplete:        t.semanticComplete,
		nextOrdinal:             t.nextOrdinal,
	}
	for i, entry := range t.references {
		entry.reference = cloneTraceReference(entry.reference)
		entry.fieldIdentity = cloneRequirementTraceIdentity(entry.fieldIdentity)
		entry.owners = cloneSourceOwners(entry.owners)
		out.references[i] = entry
	}
	for i, entry := range t.diagnostics {
		entry.pendingReferenceIDs = append([]string{}, entry.pendingReferenceIDs...)
		out.diagnostics[i] = entry
	}
	for id, index := range t.pendingReferenceIndexes {
		out.pendingReferenceIndexes[id] = index
	}
	for id := range t.incompleteStageIDs {
		out.incompleteStageIDs[id] = struct{}{}
	}
	return out
}

func (t *requirementTrace) forkBranch() *requirementTrace {
	return t.clone()
}

type requirementTraceMergeKey struct {
	kind       string
	identity   string
	fieldKey   fieldIdentityKey
	ownerKey   string
	role       string
	resolution string
}

type requirementTraceMergeEvent struct {
	branchOrdinal int
	eventOrdinal  int
	location      Location
	reference     *requirementTraceReference
	diagnostic    *requirementTraceDiagnostic
}

func mergeRequirementTraces(base *requirementTrace, paths []requirementTracePath) *requirementTrace {
	if base == nil {
		return nil
	}
	reachable := make([]requirementTracePath, 0, len(paths))
	for _, path := range paths {
		if path.Reachable && path.Trace != nil {
			reachable = append(reachable, path)
		}
	}
	sort.SliceStable(reachable, func(i, j int) bool { return reachable[i].Ordinal < reachable[j].Ordinal })
	if len(reachable) == 0 {
		return base.clone()
	}

	directPathCounts := map[requirementTraceMergeKey]int{}
	events := []requirementTraceMergeEvent{}
	for _, path := range reachable {
		assertRequirementTracePrefix(base, path.Trace)
		pathDirect := map[requirementTraceMergeKey]bool{}
		for i := len(base.references); i < len(path.Trace.references); i++ {
			entry := path.Trace.references[i]
			copy := entry
			copy.reference = cloneTraceReference(entry.reference)
			copy.fieldIdentity = cloneRequirementTraceIdentity(entry.fieldIdentity)
			copy.owners = cloneSourceOwners(entry.owners)
			events = append(events, requirementTraceMergeEvent{
				branchOrdinal: path.Ordinal,
				eventOrdinal:  entry.eventOrdinal,
				location:      entry.reference.Location,
				reference:     &copy,
			})
			if entry.directExternal {
				pathDirect[requirementTraceKey(entry)] = true
			}
		}
		for key := range pathDirect {
			directPathCounts[key]++
		}
		for i := len(base.diagnostics); i < len(path.Trace.diagnostics); i++ {
			entry := path.Trace.diagnostics[i]
			copy := entry
			copy.pendingReferenceIDs = append([]string{}, entry.pendingReferenceIDs...)
			events = append(events, requirementTraceMergeEvent{
				branchOrdinal: path.Ordinal,
				eventOrdinal:  entry.eventOrdinal,
				location:      entry.diagnostic.Location,
				diagnostic:    &copy,
			})
		}
	}
	sort.SliceStable(events, func(i, j int) bool {
		left, right := events[i], events[j]
		if left.location.Start.Offset != right.location.Start.Offset {
			return left.location.Start.Offset < right.location.Start.Offset
		}
		if left.location.End.Offset != right.location.End.Offset {
			return left.location.End.Offset < right.location.End.Offset
		}
		if left.branchOrdinal != right.branchOrdinal {
			return left.branchOrdinal < right.branchOrdinal
		}
		return left.eventOrdinal < right.eventOrdinal
	})

	merged := base.clone()
	for _, path := range reachable {
		merged.inputs = mergeInputFacts(merged.inputs, path.Trace.inputs)
		merged.correlations = mergeCorrelationEvents(merged.correlations, path.Trace.correlations)
	}
	for _, path := range reachable {
		merged.syntaxComplete = merged.syntaxComplete && path.Trace.syntaxComplete
		merged.semanticComplete = merged.semanticComplete && path.Trace.semanticComplete
		for stageID := range path.Trace.incompleteStageIDs {
			merged.incompleteStageIDs[stageID] = struct{}{}
		}
	}
	for _, event := range events {
		ordinal := merged.nextEvent()
		if event.reference != nil {
			entry := *event.reference
			if entry.directExternal || entry.conditional || entry.pathConditional {
				if entry.conditional {
					entry.directExternal = false
				} else if entry.directExternal && directPathCounts[requirementTraceKey(entry)] == len(reachable) {
					entry.directExternal = true
					entry.conditional = false
					entry.pathConditional = false
				} else {
					entry.directExternal = false
					entry.conditional = false
					entry.pathConditional = true
				}
			}
			merged.recordReference(entry.reference, entry.directExternal, entry.conditional, ordinal)
			merged.references[len(merged.references)-1].owners = cloneSourceOwners(entry.owners)
			merged.references[len(merged.references)-1].pathConditional = entry.pathConditional
			merged.references[len(merged.references)-1].fieldIdentity = cloneRequirementTraceIdentity(entry.fieldIdentity)
			continue
		}
		entry := *event.diagnostic
		merged.recordDiagnostic(entry.diagnostic, entry.incomplete, entry.pendingReferenceIDs, ordinal)
	}
	return merged
}

// Lazy declaration binding can extend the canonical trace after a branch was
// forked. Rebase keeps that branch's immutable suffix exactly once.
func rebaseRequirementTrace(oldBase, newBase, branch *requirementTrace) *requirementTrace {
	if oldBase == nil || newBase == nil || branch == nil {
		panic("requirement trace rebase requires old base, new base, and branch")
	}
	assertRequirementTracePrefix(oldBase, branch)
	rebased := newBase.clone()
	rebased.inputs = mergeInputFacts(rebased.inputs, branch.inputs)
	rebased.correlations = mergeCorrelationEvents(rebased.correlations, branch.correlations)
	rebased.syntaxComplete = rebased.syntaxComplete && branch.syntaxComplete
	rebased.semanticComplete = rebased.semanticComplete && branch.semanticComplete
	for stageID := range branch.incompleteStageIDs {
		rebased.incompleteStageIDs[stageID] = struct{}{}
	}

	events := make([]requirementTraceMergeEvent, 0, len(branch.references)-len(oldBase.references)+len(branch.diagnostics)-len(oldBase.diagnostics))
	for i := len(oldBase.references); i < len(branch.references); i++ {
		entry := branch.references[i]
		copy := entry
		copy.reference = cloneTraceReference(entry.reference)
		copy.fieldIdentity = cloneRequirementTraceIdentity(entry.fieldIdentity)
		copy.owners = cloneSourceOwners(entry.owners)
		events = append(events, requirementTraceMergeEvent{eventOrdinal: entry.eventOrdinal, reference: &copy})
	}
	for i := len(oldBase.diagnostics); i < len(branch.diagnostics); i++ {
		entry := branch.diagnostics[i]
		copy := entry
		copy.pendingReferenceIDs = append([]string{}, entry.pendingReferenceIDs...)
		events = append(events, requirementTraceMergeEvent{eventOrdinal: entry.eventOrdinal, diagnostic: &copy})
	}
	sort.SliceStable(events, func(i, j int) bool { return events[i].eventOrdinal < events[j].eventOrdinal })
	for _, event := range events {
		ordinal := rebased.nextEvent()
		if event.reference != nil {
			entry := *event.reference
			rebased.recordReference(entry.reference, entry.directExternal, entry.conditional, ordinal)
			rebased.references[len(rebased.references)-1].owners = cloneSourceOwners(entry.owners)
			rebased.references[len(rebased.references)-1].pathConditional = entry.pathConditional
			rebased.references[len(rebased.references)-1].fieldIdentity = cloneRequirementTraceIdentity(entry.fieldIdentity)
			continue
		}
		entry := *event.diagnostic
		rebased.recordDiagnostic(entry.diagnostic, entry.incomplete, entry.pendingReferenceIDs, ordinal)
	}
	return rebased
}

func requirementTraceKey(entry requirementTraceReference) requirementTraceMergeKey {
	var fieldKey fieldIdentityKey
	if entry.reference.Kind == "field" && entry.reference.Resolution == "exact" {
		fieldKey, _ = requirementSourceIdentity(entry).privateKey()
	}
	return requirementTraceMergeKey{
		kind:       entry.reference.Kind,
		identity:   requirementSourceIdentity(entry).PublicName,
		fieldKey:   fieldKey,
		ownerKey:   requirementOwnerKey(entry),
		role:       entry.reference.Role,
		resolution: entry.reference.Resolution,
	}
}

func assertRequirementTracePrefix(base, branch *requirementTrace) {
	if len(branch.inputs) < len(base.inputs) || len(base.inputs) > 0 && !reflect.DeepEqual(branch.inputs[:len(base.inputs)], base.inputs) {
		panic("requirement trace branch changed its immutable input prefix")
	}
	if len(branch.references) < len(base.references) || len(branch.diagnostics) < len(base.diagnostics) {
		panic("requirement trace branch does not contain its base prefix")
	}
	if !reflect.DeepEqual(branch.references[:len(base.references)], base.references) || !reflect.DeepEqual(branch.diagnostics[:len(base.diagnostics)], base.diagnostics) {
		panic("requirement trace branch changed its immutable base prefix")
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
	remapInputReferences(t.inputs, mapping)
	t.remapCorrelationReferences(mapping)
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
		remapSourceOwnerReferences(entry.owners, mapping)
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
	if t.incompleteStageIDs == nil {
		t.rebuildIncompleteStageIndex()
	}
	rebuildIndex := false
	for i, diagnostic := range diagnostics {
		if i >= len(t.diagnostics) {
			break
		}
		entry := &t.diagnostics[i]
		oldStageID := entry.diagnostic.StageID
		entry.diagnostic = diagnostic
		if !entry.incomplete || oldStageID == diagnostic.StageID {
			continue
		}
		if oldStageID == "" && diagnostic.StageID != "" {
			t.incompleteStageIDs[diagnostic.StageID] = struct{}{}
			continue
		}
		rebuildIndex = true
	}
	if rebuildIndex {
		t.rebuildIncompleteStageIndex()
	}
}

func (t *requirementTrace) remapStages(mapping map[string]string) {
	if budget := t.sourceEvidenceBudget; budget != nil && budget.failure != nil {
		if id, ok := mapping[budget.failure.StageID]; ok {
			budget.failure.StageID = id
		}
	}
	remapInputStages(t.inputs, mapping)
	t.remapCorrelationStages(mapping)
	for i := range t.references {
		remapSourceOwnerStages(t.references[i].owners, mapping)
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
	if t.incompleteStageIDs == nil {
		t.incompleteStageIDs = map[string]struct{}{}
	} else {
		clear(t.incompleteStageIDs)
	}
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
		if got.NormalizedName != want.NormalizedName || got.OriginalName != want.OriginalName || got.Kind != want.Kind || got.Role != want.Role || got.StageID != want.StageID || got.ScopeID != want.ScopeID || got.Location != want.Location || got.Resolution != want.Resolution || !reflect.DeepEqual(got.FieldIdentity, want.FieldIdentity) {
			panic(fmt.Sprintf("requirement trace reference %q differs from its public counterpart: trace=%+v public=%+v", entry.reference.ID, want, got))
		}
	}
}

func cloneTraceReference(reference Reference) Reference {
	reference.OriginReferenceIDs = append([]string{}, reference.OriginReferenceIDs...)
	if reference.FieldIdentity != nil {
		identity := *reference.FieldIdentity
		identity.Segments = append([]string{}, identity.Segments...)
		reference.FieldIdentity = &identity
	}
	return reference
}

func cloneRequirementTraceIdentity(identity fieldIdentity) fieldIdentity {
	if identity.Segments == nil {
		return identity
	}
	return identity.clone()
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
		trace:     trace,
		inputs:    []inputFact{},
		fields:    map[fieldIdentityKey]requirementField{},
		removed:   map[fieldIdentityKey]bool{},
		ambiguous: map[string]bool{},
		open:      true,
	}
}

func (e requirementEnvironment) clone() requirementEnvironment {
	out := newRequirementEnvironment(e.trace)
	out.inputs = cloneInputFacts(e.inputs)
	out.open = e.open
	out.uncertain = e.uncertain
	for name, field := range e.fields {
		field.origins = append([]string{}, field.origins...)
		field.identity = field.identity.clone()
		field.owners = cloneSourceOwners(field.owners)
		out.fields[name] = field
	}
	for name, removed := range e.removed {
		out.removed[name] = removed
	}
	for name, ambiguous := range e.ambiguous {
		out.ambiguous[name] = ambiguous
	}
	return out
}

func (e *requirementEnvironment) read(reference Reference) (binding string, directExternal, conditional bool) {
	return e.readIdentity(reference, atomicFieldIdentity(reference.NormalizedName))
}

func (e *requirementEnvironment) readIdentity(reference Reference, identity fieldIdentity) (binding string, directExternal, conditional bool) {
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
	key, exactIdentity := identity.privateKey()
	if reference.Resolution == "wildcard" || reference.Resolution == "dynamic" {
		return classified("indeterminate", false, true)
	}
	field, known := e.fields[key]
	owners := cloneSourceOwners(field.owners)
	if !known {
		owners = sourceOwnersForIdentity(e.inputs, identity)
	}
	switch {
	case known && (field.conditional || field.ownerCollision):
		return classified("indeterminate", false, true)
	case known:
		if field.source {
			if len(owners) > 0 && !sourceOwnersProved(owners) {
				return classified("indeterminate", false, true)
			}
			return classified("source", true, false)
		}
		return classified("derived", false, false)
	case e.uncertain:
		return classified("indeterminate", false, true)
	case exactIdentity && e.removed[key] || !e.open:
		return classified("unavailable", false, false)
	default:
		if len(owners) > 0 && !sourceOwnersProved(owners) {
			return classified("indeterminate", false, true)
		}
		if !nullTest {
			e.fields[key] = requirementField{identity: identity.clone(), source: true, owners: cloneSourceOwners(owners), origins: []string{reference.ID}}
			e.detectCollision(reference.NormalizedName)
		}
		return classified("source", true, false)
	}
}

func requirementReferencePolicy(reference Reference) (directExternal, conditional bool) {
	if reference.Role != "read" {
		return false, false
	}
	switch reference.Kind {
	case "index", "source", "sourcetype", "dataset", "data_model", "lookup", "macro", "saved_search", "module", "function":
		if reference.Resolution == "exact" {
			return true, false
		}
		return false, true
	default:
		return false, false
	}
}

func (e *requirementEnvironment) install(name string, origins []string, conditional bool) {
	e.installIdentity(atomicFieldIdentity(name), origins, conditional, false)
}

func (e *requirementEnvironment) installIdentity(identity fieldIdentity, origins []string, conditional, source bool) {
	key, exact := identity.privateKey()
	if !exact {
		return
	}
	e.registerIdentityField(requirementField{identity: identity.clone(), owners: traceSourceOwners(e.trace, origins), source: source, conditional: conditional, origins: append([]string{}, origins...)})
	delete(e.removed, key)
}

func (e *requirementEnvironment) registerIdentityField(field requirementField) {
	key, exact := field.identity.privateKey()
	if !exact {
		return
	}
	field.identity = field.identity.clone()
	field.origins = append([]string{}, field.origins...)
	field.owners = cloneSourceOwners(field.owners)
	e.fields[key] = field
	e.detectCollision(field.identity.PublicName)
}

func (e *requirementEnvironment) field(identity fieldIdentity) (requirementField, bool) {
	key, exact := identity.privateKey()
	if !exact {
		return requirementField{}, false
	}
	field, known := e.fields[key]
	return field, known
}

func (e *requirementEnvironment) detectCollision(publicName string) {
	identities := map[fieldIdentityKey]bool{}
	for _, field := range e.fields {
		if field.identity.PublicName != publicName {
			continue
		}
		key, exact := field.identity.privateKey()
		if !exact {
			key, _ = atomicFieldIdentity(publicName).privateKey()
		}
		identities[key] = true
	}
	if len(identities) > 1 {
		e.ambiguous[publicName] = true
	}
}

func (e *requirementEnvironment) markConditional(name string) {
	e.markConditionalIdentity(atomicFieldIdentity(name))
}

func (e *requirementEnvironment) markConditionalIdentity(identity fieldIdentity) {
	key, _ := identity.privateKey()
	field, known := e.fields[key]
	if !known {
		return
	}
	field.conditional = true
	e.fields[key] = field
}

func (e *requirementEnvironment) remove(name string) {
	e.removeIdentity(atomicFieldIdentity(name))
}

func (e *requirementEnvironment) removeIdentity(identity fieldIdentity) {
	key, exact := identity.privateKey()
	if !exact {
		return
	}
	delete(e.fields, key)
	e.removed[key] = true
}

func (e *requirementEnvironment) stageIncomplete(stageID string) bool {
	if e.trace == nil {
		return false
	}
	_, incomplete := e.trace.incompleteStageIDs[stageID]
	return incomplete
}

func (e *requirementEnvironment) exactProjection(name string, referenceIDs []string) (requirementField, bool) {
	return e.exactIdentityProjection(atomicFieldIdentity(name), referenceIDs)
}

func (e *requirementEnvironment) exactIdentityProjection(identity fieldIdentity, referenceIDs []string) (requirementField, bool) {
	key, _ := identity.privateKey()
	if field, ok := e.fields[key]; ok {
		field.owners = cloneSourceOwners(field.owners)
		field.origins = append([]string{}, field.origins...)
		return field, true
	}
	if e.trace == nil {
		return requirementField{}, false
	}
	matching := []string{}
	for _, id := range referenceIDs {
		if id == "" {
			continue
		}
		entry := e.trace.reference(id)
		reference := entry.reference
		entryKey, entryExact := entry.fieldIdentity.privateKey()
		if reference.Kind != "field" || !entryExact || entryKey != key || reference.Resolution != "exact" || reference.Binding != "indeterminate" || !entry.conditional || reference.Role == "null_test" {
			continue
		}
		matching = append(matching, id)
	}
	if len(matching) == 0 {
		return requirementField{}, false
	}
	return requirementField{identity: identity.clone(), owners: traceSourceOwners(e.trace, matching), source: true, conditional: true, origins: traceOrigins(e.trace, matching)}, true
}

func (e *requirementEnvironment) applyProjection(selectors []locatedOperand, referenceIDs [][]string, mode string, retainKnownInternals bool, stageID string) {
	if mode == "exclude" {
		for _, selector := range selectors {
			if selector.Resolution != "wildcard" {
				continue
			}
			for key, field := range e.fields {
				if wildcardMatches(selector.Name, field.identity.PublicName) {
					e.removeIdentity(field.identity)
					delete(e.fields, key)
				}
			}
		}
		return
	}

	selected := []requirementField{}
	selectedIndex := map[fieldIdentityKey]int{}
	selectField := func(field requirementField) {
		key, exact := field.identity.privateKey()
		if !exact {
			return
		}
		if index, exists := selectedIndex[key]; exists {
			selected[index] = field
			return
		}
		selectedIndex[key] = len(selected)
		selected = append(selected, field)
	}
	if retainKnownInternals {
		for _, key := range orderedIdentityKeys(nil, e.fields) {
			field := e.fields[key]
			if strings.HasPrefix(field.identity.PublicName, "_") {
				selectField(field)
			}
		}
	}
	for i, selector := range selectors {
		if selector.Resolution == "wildcard" {
			for _, key := range orderedIdentityKeys(nil, e.fields) {
				field := e.fields[key]
				if wildcardMatches(selector.Name, field.identity.PublicName) {
					selectField(field)
				}
			}
			continue
		}
		var ids []string
		if i < len(referenceIDs) {
			ids = referenceIDs[i]
		}
		if field, ok := e.exactIdentityProjection(selector.fieldIdentity(), ids); ok {
			selectField(field)
		}
	}
	e.fields = map[fieldIdentityKey]requirementField{}
	for _, field := range selected {
		e.registerIdentityField(field)
	}
	e.open = false
	if mode == "table" && !e.stageIncomplete(stageID) {
		e.uncertain = false
	}
}
