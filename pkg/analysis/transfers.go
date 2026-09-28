package analysis

import (
	"fmt"
	"sort"
	"strings"
)

// locatedOperand is decoded and checked for soundness by its language frontend.
type locatedOperand struct {
	rewrite    rewriteOwner
	Name       string
	Identity   fieldIdentity
	Location   Location
	Resolution string
	Sound      bool
	// UnresolvedSource preserves frontend distinctions absent from string universes.
	// Derived bindings and structurally proven absence remain independently sound.
	UnresolvedSource bool
}

func (o locatedOperand) fieldIdentity() fieldIdentity {
	if _, exact := o.Identity.privateKey(); exact {
		return o.Identity
	}
	return atomicFieldIdentity(o.Name)
}

type renameOperands struct{ Source, Target locatedOperand }
type aggregateOutput struct {
	Target                 locatedOperand
	InputReferenceIDs      []string
	Conditional            bool
	RequirementConditional bool
}
type preparedSelection struct {
	Field                 trackedField
	InputReferenceIDs     []string
	EmitProjectTransition bool
}
type lookupOutput struct {
	Target           locatedOperand
	PreserveExisting bool
}

func transitionOutputIdentity(identity fieldIdentity) *FieldIdentity {
	publicIdentity, exact := identity.public()
	if !exact {
		return nil
	}
	return &publicIdentity
}

func (s *semanticStage) diagnosticAt(code, severity, category, message string, location Location, incomplete bool) {
	s.diagnosticAtOwned(code, severity, category, message, location, incomplete, nil)
}

func (s *semanticStage) diagnosticAtOwned(code, severity, category, message string, location Location, incomplete bool, pendingReferenceIDs []string) {
	s.appendDiagnostic(code, severity, category, message, location, incomplete, incomplete, pendingReferenceIDs, true)
}

func (s *semanticStage) diagnosticAtOwnedWithRequirementIncomplete(code, severity, category, message string, location Location, pendingReferenceIDs []string) {
	s.appendDiagnostic(code, severity, category, message, location, false, true, pendingReferenceIDs, true)
}

func (s *semanticStage) refinementDiagnosticAt(code, severity, category, message string, location Location, incomplete bool) {
	s.appendDiagnostic(code, severity, category, message, location, incomplete, false, nil, false)
}

func (s *semanticStage) appendDiagnostic(code, severity, category, message string, location Location, incomplete, requirementIncomplete bool, pendingReferenceIDs []string, recordRequirement bool) {
	st := &s.result.Stages[s.stage]
	if incomplete {
		st.SemanticComplete = false
		s.env.uncertain = true
		s.rewriteUncertain()
	}
	diagnostic := Diagnostic{Code: code, Severity: severity, Category: category, Message: message, Location: location, StageID: st.ID, ScopeID: st.ScopeID}
	s.result.Diagnostics = append(s.result.Diagnostics, diagnostic)
	if trace := s.env.requirements.trace; recordRequirement && trace != nil {
		if requirementIncomplete {
			s.env.requirements.uncertain = true
		}
		trace.recordDiagnostic(diagnostic, requirementIncomplete, pendingReferenceIDs, trace.nextEvent())
	}
}

func (s *semanticStage) requirementDiagnosticAt(code, severity, category, message string, location Location, incomplete bool, pendingReferenceIDs []string) {
	trace := s.env.requirements.trace
	if trace == nil {
		return
	}
	st := s.result.Stages[s.stage]
	if incomplete {
		s.env.requirements.uncertain = true
	}
	trace.recordDiagnostic(Diagnostic{Code: code, Severity: severity, Category: category, Message: message, Location: location, StageID: st.ID, ScopeID: st.ScopeID}, incomplete, pendingReferenceIDs, trace.nextEvent())
}
func (s *semanticStage) operandReference(operand locatedOperand, kind, role string) string {
	if !operand.Sound {
		return ""
	}
	id := s.referenceAt(operand.Location, operand.Name, kind, role, operand.Resolution)
	if kind == "field" && operand.Resolution == "exact" {
		if identity, exact := operand.fieldIdentity().public(); exact {
			s.result.References[len(s.result.References)-1].FieldIdentity = &identity
		}
	}
	if trace := s.env.requirements.trace; trace != nil {
		entry := trace.reference(id)
		if kind == "field" {
			entry.reference.Binding, entry.directExternal, entry.conditional = s.env.requirements.readIdentity(entry.reference, operand.fieldIdentity())
		} else {
			entry.directExternal, entry.conditional = requirementReferencePolicy(entry.reference)
		}
	}
	s.rewriteReference(id, operand, kind, role)
	s.rewriteIdentityCoverage(operand, kind)
	return id
}
func (s *semanticStage) readAt(operand locatedOperand, role string) string {
	name := operand.Name
	identity := operand.fieldIdentity()
	id := s.operandReference(operand, "field", role)
	if id == "" {
		return id
	}
	ref := &s.result.References[len(s.result.References)-1]
	requirementBinding := ref.Binding
	if trace := s.env.requirements.trace; trace != nil {
		requirementBinding = trace.reference(id).reference.Binding
	}
	f, known := s.env.field(identity)
	key, exactIdentity := identity.privateKey()
	switch {
	case known && !f.Conditional:
		ref.Binding = "derived"
		if f.source {
			ref.Binding = "source"
		}
		ref.OriginReferenceIDs = copyIDs(f.OriginReferenceIDs)
	case s.env.uncertain || (known && f.Conditional):
		ref.Binding = "indeterminate"
		if known {
			ref.OriginReferenceIDs = copyIDs(f.OriginReferenceIDs)
		} else if exactIdentity && role != "null_test" && s.env.hasOtherIdentity(identity) {
			collision, owners := s.env.installIdentity(identity, []string{id}, true, true)
			if collision {
				s.fieldIdentityCollision(name, operand.Location, owners)
			}
		}
	case exactIdentity && s.env.removed[key] || !s.env.open:
		ref.Binding = "unavailable"
	default:
		ref.Binding = "source"
		if role != "null_test" {
			collision, owners := s.env.installIdentity(identity, []string{id}, false, true)
			if collision {
				s.fieldIdentityCollision(name, operand.Location, owners)
			}
		}
	}
	if role != "null_test" {
		message := fmt.Sprintf("field %q is unavailable after an earlier pipeline transfer", name)
		publicUnavailable := ref.Binding == "unavailable"
		requirementUnavailable := requirementBinding == "unavailable"
		switch {
		case publicUnavailable && requirementUnavailable:
			s.diagnosticAtOwned(CodeUnavailableField, "error", "unavailable_field", message, operand.Location, false, []string{id})
		case publicUnavailable:
			s.appendDiagnostic(CodeUnavailableField, "error", "unavailable_field", message, operand.Location, false, false, nil, false)
		case requirementUnavailable:
			s.requirementDiagnosticAt(CodeUnavailableField, "error", "unavailable_field", message, operand.Location, false, []string{id})
		}
	}
	if operand.UnresolvedSource && s.refinement != nil && s.refinement.resolve != nil && ref.Binding == "source" {
		ref.Binding = "indeterminate"
		if role != "null_test" {
			field, _ := s.env.field(identity)
			field.Conditional = true
			key, _ := identity.privateKey()
			s.env.fields[key] = field
		}
		s.refinementDiagnosticAt(CodeUnsupportedSemantics, "warning", "unsupported_semantics", "Source-name identity is not represented by the string-only source universe", operand.Location, false)
		s.result.Stages[s.stage].SemanticComplete = false
	}
	s.rewriteBinding(id, ref.Binding, nil)
	return id
}
func (s *semanticStage) createAt(operand locatedOperand, role, operation string, inputs []string, conditional bool) string {
	return s.createAtWithRequirementConditional(operand, role, operation, inputs, conditional, conditional)
}

func (s *semanticStage) createAtWithRequirementConditional(operand locatedOperand, role, operation string, inputs []string, conditional, requirementConditional bool) string {
	name := operand.Name
	id := s.operandReference(operand, "field", role)
	if id == "" {
		return id
	}
	s.rewriteBinding(id, "definition", inputs)
	origins := s.origins(inputs)
	s.result.References[len(s.result.References)-1].OriginReferenceIDs = copyIDs(origins)
	collision, owners := s.env.installIdentity(operand.fieldIdentity(), uniqueIDs([]string{id}, origins), conditional, false)
	if trace := s.env.requirements.trace; trace != nil {
		entry := trace.reference(id)
		entry.reference.Binding = "definition"
		entry.reference.OriginReferenceIDs = traceOrigins(trace, inputs)
		s.env.requirements.installIdentity(operand.fieldIdentity(), uniqueIDs([]string{id}, entry.reference.OriginReferenceIDs), requirementConditional, false)
	}
	if collision {
		s.fieldIdentityCollision(name, operand.Location, owners)
	}
	outputIdentity := transitionOutputIdentity(operand.fieldIdentity())
	s.appendTransition(Transition{Operation: operation, Output: name, OutputIdentity: outputIdentity, InputReferenceIDs: copyIDs(inputs), OutputReferenceID: id, Conditional: conditional})
	return id
}
func (s *semanticStage) selectorAt(operand locatedOperand, role string, allowWildcard bool) ([]fieldIdentity, []string) {
	name := operand.Name
	if operand.Resolution != "wildcard" {
		return []fieldIdentity{operand.fieldIdentity()}, []string{s.readAt(operand, role)}
	}
	id := s.operandReference(operand, "field", role)
	if id == "" {
		return []fieldIdentity{}, []string{}
	}
	ref := &s.result.References[len(s.result.References)-1]
	if role != "remove" {
		ref.Binding = "indeterminate"
	}
	command := s.result.Stages[s.stage].Command
	if s.refinement != nil {
		if !allowWildcard {
			s.requirementDiagnosticAt(CodeUnsupportedSemantics, "warning", "unsupported_semantics", fmt.Sprintf("wildcard selectors for command %q are unmodeled", command), operand.Location, true, []string{id})
		} else if s.env.requirements.open || s.env.requirements.uncertain {
			s.requirementDiagnosticAt(CodeUnresolvedWildcard, "warning", "unsupported_semantics", fmt.Sprintf("wildcard %q membership is unresolved", name), operand.Location, true, []string{id})
		}
		identities := s.refinedSelectorAt(operand, role, id)
		if !allowWildcard {
			s.refinementDiagnosticAt(CodeUnsupportedSemantics, "warning", "unsupported_semantics", fmt.Sprintf("wildcard selectors for command %q are unmodeled", command), operand.Location, true)
			return []fieldIdentity{}, []string{id}
		}
		return identities, []string{id}
	}
	if !allowWildcard {
		s.diagnosticAtOwned(CodeUnsupportedSemantics, "warning", "unsupported_semantics", fmt.Sprintf("wildcard selectors for command %q are unmodeled", command), operand.Location, true, []string{id})
		return []fieldIdentity{}, []string{id}
	}
	identities := []fieldIdentity{}
	origins := []string{}
	for _, key := range s.env.orderedFieldKeys() {
		f, exists := s.env.fields[key]
		if !exists {
			continue
		}
		if wildcardMatches(name, f.Name) {
			identities = append(identities, f.identity.clone())
			origins = uniqueIDs(origins, f.OriginReferenceIDs)
		}
	}
	sort.Slice(identities, func(i, j int) bool {
		if identities[i].PublicName != identities[j].PublicName {
			return identities[i].PublicName < identities[j].PublicName
		}
		left, _ := identities[i].privateKey()
		right, _ := identities[j].privateKey()
		return left < right
	})
	sort.Strings(origins)
	ref.OriginReferenceIDs = origins
	if s.env.open || s.env.uncertain {
		s.diagnosticAtOwned(CodeUnresolvedWildcard, "warning", "unsupported_semantics", fmt.Sprintf("wildcard %q membership is unresolved", name), operand.Location, true, []string{id})
	}
	return identities, []string{id}
}

// applyProjection prepares lexical reads once before restricting the environment.
// mode is the frontend's finite policy: include, exclude, or table.
func (s *semanticStage) applyProjection(selectors []locatedOperand, mode string, retainKnownInternals bool) {
	exclude := mode == "exclude"
	selected := []preparedSelection{}
	requirementReferenceIDs := make([][]string, len(selectors))
	if !exclude && retainKnownInternals {
		if s.refinement != nil && s.env.requirements.open {
			s.requirementDiagnosticAt(CodeUnsupportedSemantics, "warning", "unsupported_semantics", "fields inclusion retains internal fields with unresolved open-source membership", s.result.Stages[s.stage].Location, true, nil)
		}
		internalsComplete := true
		if s.refinement != nil {
			internalsComplete = s.retainSourceInternals()
		}
		for _, key := range s.env.orderedFieldKeys() {
			field := s.env.fields[key]
			if strings.HasPrefix(field.Name, "_") {
				selected = append(selected, preparedSelection{Field: field})
			}
		}
		if (s.env.open && s.refinement == nil) || !internalsComplete {
			if s.refinement != nil {
				s.refinementDiagnosticAt(CodeUnsupportedSemantics, "warning", "unsupported_semantics", "fields inclusion retains internal fields with unresolved open-source membership", s.result.Stages[s.stage].Location, true)
			} else {
				s.diagnosticAt(CodeUnsupportedSemantics, "warning", "unsupported_semantics", "fields inclusion retains internal fields with unresolved open-source membership", s.result.Stages[s.stage].Location, true)
			}
		}
	}
	for i, operand := range selectors {
		if exclude && operand.Resolution != "wildcard" {
			s.removeAt(operand)
			continue
		}
		role := "read"
		if exclude {
			role = "remove"
		}
		identities, ids := s.selectorAt(operand, role, true)
		requirementReferenceIDs[i] = copyIDs(ids)
		for _, identity := range identities {
			if exclude {
				s.env.removeIdentity(identity)
				outputIdentity := transitionOutputIdentity(identity)
				s.appendTransition(Transition{Operation: "remove", Output: identity.PublicName, OutputIdentity: outputIdentity, InputReferenceIDs: copyIDs(ids)})
			} else if f, ok := s.projectedIdentityField(identity, ids); ok {
				selected = append(selected, preparedSelection{Field: f, InputReferenceIDs: ids, EmitProjectTransition: true})
			}
		}
	}
	s.env.requirements.applyProjection(selectors, requirementReferenceIDs, mode, retainKnownInternals, s.result.Stages[s.stage].ID)
	if !exclude {
		s.applyPreparedProjection(selected, mode)
	}
}

// applyPreparedProjection installs copied binding evidence without resolving it
// again. Ordered records intentionally retain repeated explicit projections.
func (s *semanticStage) applyPreparedProjection(selected []preparedSelection, mode string) {
	previousFields := s.env.fields
	previousOrder := s.env.orderedFieldKeys()
	registrations := []preparedSelection{}
	registrationIndex := map[fieldIdentityKey]int{}
	register := func(selection preparedSelection) {
		field := selection.Field
		key, exact := field.identity.privateKey()
		if !exact {
			field.identity = atomicFieldIdentity(field.Name)
			key, _ = field.identity.privateKey()
		}
		selection.Field = field
		if index, exists := registrationIndex[key]; exists {
			registrations[index] = selection
			return
		}
		registrationIndex[key] = len(registrations)
		registrations = append(registrations, selection)
	}
	for _, selection := range selected {
		field := selection.Field
		key, exact := field.identity.privateKey()
		if !exact {
			field.identity = atomicFieldIdentity(field.Name)
			key, _ = field.identity.privateKey()
		}
		if s.env.ambiguous[field.Name] {
			matched := false
			for _, candidateKey := range previousOrder {
				candidate, exists := previousFields[candidateKey]
				if exists && candidate.Name == field.Name {
					candidateSelection := preparedSelection{Field: candidate}
					if candidateKey == key {
						candidateSelection = selection
						candidateSelection.Field = field
						matched = true
					}
					register(candidateSelection)
				}
			}
			if !matched {
				selection.Field = field
				register(selection)
			}
			continue
		}
		selection.Field = field
		register(selection)
	}

	s.env.fields = map[fieldIdentityKey]trackedField{}
	s.env.fieldOrder = nil
	for _, selection := range registrations {
		collision, owners := s.env.registerIdentityField(selection.Field)
		if collision {
			s.fieldIdentityCollision(selection.Field.Name, s.referenceLocation(selection.InputReferenceIDs), owners)
		}
	}
	for _, selection := range selected {
		if selection.EmitProjectTransition {
			outputIdentity := transitionOutputIdentity(selection.Field.identity)
			s.appendTransition(Transition{Operation: "project", Output: selection.Field.Name, OutputIdentity: outputIdentity, InputReferenceIDs: copyIDs(selection.InputReferenceIDs)})
		}
	}
	s.env.rewriteProject(s.env.fields)
	// Partial selectors retain an unknown remainder; finite compatibility keeps
	// its historical closed-output wire shape. Existing tombstones are retained.
	s.env.open = s.refinement != nil && !s.refinement.finiteCompatibility && s.env.open && !s.result.Stages[s.stage].SemanticComplete
	if mode == "table" && s.result.Stages[s.stage].SemanticComplete {
		s.env.uncertain = false
	}
}

func (s *semanticStage) referenceLocation(ids []string) Location {
	for _, id := range ids {
		for _, reference := range s.result.References {
			if reference.ID == id {
				return reference.Location
			}
		}
	}
	return s.result.Stages[s.stage].Location
}

// Exact projection fixes the output names without proving conditional inputs exist.
func (s *semanticStage) projectedField(name string, ids []string) (trackedField, bool) {
	return s.projectedIdentityField(atomicFieldIdentity(name), ids)
}

func (s *semanticStage) projectedIdentityField(identity fieldIdentity, ids []string) (trackedField, bool) {
	if field, known := s.env.field(identity); known {
		return field, true
	}
	if s.env.uncertain {
		return trackedField{FieldBinding: FieldBinding{Name: identity.PublicName, OriginReferenceIDs: s.origins(ids), Conditional: true}, identity: identity.clone()}, true
	}
	return trackedField{}, false
}

func (s *semanticStage) applyRename(pairs []renameOperands) {
	original := s.env
	requirementOriginal := s.env.requirements.clone()
	before := s.env.clone()
	type rename struct {
		source, dest           string
		target                 locatedOperand
		input                  string
		output                 string
		conditional            bool
		requirementConditional bool
	}
	items := []rename{}
	publicSources, publicDests := map[string]bool{}, map[string]bool{}
	requirementSources, requirementDests := map[string]bool{}, map[string]bool{}
	publicSourceKeys, publicDestKeys := map[fieldIdentityKey]bool{}, map[fieldIdentityKey]bool{}
	requirementSourceKeys, requirementDestKeys := map[fieldIdentityKey]bool{}, map[fieldIdentityKey]bool{}
	trackIdentity := func(keys map[fieldIdentityKey]bool, identity fieldIdentity) {
		if key, exact := identity.privateKey(); exact {
			keys[key] = true
		}
	}
	publicConflict, requirementConflict := false, false
	for _, r := range pairs {
		if !r.Source.Sound || !r.Target.Sound {
			continue
		}
		src := r.Source.Name
		dst := r.Target.Name
		if r.Source.Resolution == "wildcard" || r.Target.Resolution == "wildcard" {
			s.env = before
			identities, ids := s.selectorAt(r.Source, "read", true)
			for _, identity := range identities {
				publicSources[identity.PublicName] = true
				trackIdentity(publicSourceKeys, identity)
			}
			if r.Source.Resolution == "wildcard" {
				for key, field := range requirementOriginal.fields {
					if wildcardMatches(src, field.identity.PublicName) {
						requirementSources[field.identity.PublicName] = true
						requirementSourceKeys[key] = true
					}
				}
			} else {
				requirementSources[src] = true
				trackIdentity(requirementSourceKeys, r.Source.fieldIdentity())
			}
			if r.Target.Resolution == "wildcard" {
				for key, field := range original.fields {
					if wildcardMatches(dst, field.Name) {
						publicDests[field.Name] = true
						publicDestKeys[key] = true
					}
				}
				for key, field := range requirementOriginal.fields {
					if wildcardMatches(dst, field.identity.PublicName) {
						requirementDests[field.identity.PublicName] = true
						requirementDestKeys[key] = true
					}
				}
			} else {
				publicDests[dst] = true
				requirementDests[dst] = true
				trackIdentity(publicDestKeys, r.Target.fieldIdentity())
				trackIdentity(requirementDestKeys, r.Target.fieldIdentity())
				if len(ids) > 0 {
					items = append(items, rename{source: src, dest: dst, target: r.Target, input: ids[0], conditional: true, requirementConditional: true})
				}
			}
			s.diagnosticAt(CodeUnsupportedSemantics, "warning", "unsupported_semantics", "wildcard rename substitution is unmodeled", Location{Start: r.Source.Location.Start, End: r.Target.Location.End}, true)
			publicConflict = true
			requirementConflict = true
			continue
		}
		s.env = before
		id := s.readAt(r.Source, "read")
		_, publicExistingDestination := original.field(r.Target.fieldIdentity())
		_, requirementExistingDestination := requirementOriginal.field(r.Target.fieldIdentity())
		if publicSources[src] || publicDests[dst] || publicExistingDestination {
			publicConflict = true
		}
		if requirementSources[src] || requirementDests[dst] || requirementExistingDestination {
			requirementConflict = true
		}
		publicSources[src] = true
		publicDests[dst] = true
		requirementSources[src] = true
		requirementDests[dst] = true
		trackIdentity(publicSourceKeys, r.Source.fieldIdentity())
		trackIdentity(publicDestKeys, r.Target.fieldIdentity())
		trackIdentity(requirementSourceKeys, r.Source.fieldIdentity())
		trackIdentity(requirementDestKeys, r.Target.fieldIdentity())
		binding := s.result.References[len(s.result.References)-1].Binding
		requirementBinding := binding
		if trace := s.env.requirements.trace; trace != nil {
			requirementBinding = trace.reference(id).reference.Binding
		}
		items = append(items, rename{
			source:                 src,
			dest:                   dst,
			target:                 r.Target,
			input:                  id,
			conditional:            binding == "indeterminate" || binding == "unavailable",
			requirementConditional: requirementBinding == "indeterminate" || requirementBinding == "unavailable",
		})
	}
	s.env = before // All reads observed the snapshot; no destination was installed while reading.
	for name := range publicSources {
		if publicDests[name] {
			publicConflict = true
		}
	}
	for name := range requirementSources {
		if requirementDests[name] {
			requirementConflict = true
		}
	}
	message := "rename has conflicting or unsupported source/destination mappings"
	location := s.result.Stages[s.stage].Location
	switch {
	case publicConflict && requirementConflict:
		s.diagnosticAt(CodeUnsupportedSemantics, "warning", "unsupported_semantics", "rename has conflicting or unsupported source/destination mappings", s.result.Stages[s.stage].Location, true)
	case publicConflict:
		s.appendDiagnostic(CodeUnsupportedSemantics, "warning", "unsupported_semantics", message, location, true, false, nil, false)
	case requirementConflict:
		s.requirementDiagnosticAt(CodeUnsupportedSemantics, "warning", "unsupported_semantics", message, location, true, nil)
	}
	if publicConflict {
		for key := range publicSourceKeys {
			delete(s.env.fields, key)
		}
		for key := range publicDestKeys {
			delete(s.env.fields, key)
		}
		for i := range items {
			items[i].output = s.recordRejectedRenameTarget(items[i].target, items[i].input)
		}
	} else {
		for _, r := range items {
			s.env.remove(r.source)
		}
		if !requirementConflict {
			for _, r := range items {
				s.env.requirements.remove(r.source)
			}
		}
		for i := range items {
			items[i].output = s.createAtWithRequirementConditional(items[i].target, "rename", "rename", []string{items[i].input}, items[i].conditional, items[i].requirementConditional)
		}
	}
	if requirementConflict {
		for key := range requirementSourceKeys {
			delete(s.env.requirements.fields, key)
		}
		for key := range requirementDestKeys {
			delete(s.env.requirements.fields, key)
		}
		return
	}
	if publicConflict {
		for _, r := range items {
			s.env.requirements.remove(r.source)
		}
		for _, r := range items {
			origins := traceOrigins(s.env.requirements.trace, []string{r.input})
			s.env.requirements.install(r.dest, uniqueIDs([]string{r.output}, origins), r.requirementConditional)
		}
	}
}

// A rejected rename still contributes its located target reference so
// refinement cannot renumber later query evidence. It does not change field or
// rewrite state and does not emit a transition.
func (s *semanticStage) recordRejectedRenameTarget(target locatedOperand, input string) string {
	id := s.operandReference(target, "field", "rename")
	if id == "" {
		return ""
	}
	origins := s.origins([]string{input})
	s.result.References[len(s.result.References)-1].OriginReferenceIDs = copyIDs(origins)
	if trace := s.env.requirements.trace; trace != nil {
		entry := trace.reference(id)
		entry.reference.Binding = "definition"
		entry.reference.OriginReferenceIDs = traceOrigins(trace, []string{input})
	}
	return id
}
func (s *semanticStage) applyAggregation(outputs []aggregateOutput, groups []locatedOperand, preserveInput bool) {
	stageID := s.result.Stages[s.stage].ID
	output := newEnvironmentWithRequirementTrace(s.env.requirements.trace)
	output.open = false
	output.requirements.open = false
	for _, operand := range groups {
		identities, ids := s.selectorAt(operand, "group", false)
		if operand.Resolution == "exact" {
			if field, ok := s.env.requirements.exactIdentityProjection(operand.fieldIdentity(), ids); ok {
				key, _ := operand.fieldIdentity().privateKey()
				output.requirements.fields[key] = field
			}
		}
		for _, identity := range identities {
			var outputIdentity *FieldIdentity
			if field, ok := s.projectedIdentityField(identity, ids); ok {
				key, _ := identity.privateKey()
				output.fields[key] = field
				outputIdentity = transitionOutputIdentity(identity)
			}
			s.appendTransition(Transition{Operation: "project", Output: identity.PublicName, OutputIdentity: outputIdentity, InputReferenceIDs: copyIDs(ids)})
		}
	}
	if !preserveInput {
		output.rewrite = s.env.rewrite.clone()
		output.rewriteProject(output.fields)
		output.uncertain = !s.result.Stages[s.stage].SemanticComplete
		output.requirements.uncertain = s.env.requirements.stageIncomplete(stageID)
		s.env = output
	}
	for _, output := range outputs {
		s.createAtWithRequirementConditional(output.Target, "output", "aggregate", output.InputReferenceIDs,
			output.Conditional || !s.result.Stages[s.stage].SemanticComplete,
			output.RequirementConditional || s.env.requirements.stageIncomplete(stageID))
	}
}

// Match reads belong to frontend preparation, before any output effects. Calling
// once per output preserves a frontend's intervening source-order diagnostics.
func (s *semanticStage) applyLookupOutputs(matchReferenceIDs []string, outputs []lookupOutput) {
	for _, output := range outputs {
		if output.PreserveExisting {
			if _, known := s.env.field(output.Target.fieldIdentity()); known {
				s.operandReference(output.Target, "field", "output")
				continue
			}
		}
		stageID := s.result.Stages[s.stage].ID
		s.createAtWithRequirementConditional(output.Target, "output", "lookup", matchReferenceIDs,
			output.PreserveExisting || !s.result.Stages[s.stage].SemanticComplete,
			output.PreserveExisting || s.env.requirements.stageIncomplete(stageID))
	}
}

func (s *semanticStage) applyAssignment(target locatedOperand, inputs []string, conditional, removeNull bool) {
	s.applyAssignmentWithRequirementConditional(target, inputs, conditional, conditional, removeNull)
}

func (s *semanticStage) applyAssignmentWithRequirementConditional(target locatedOperand, inputs []string, conditional, requirementConditional, removeNull bool) {
	if removeNull {
		if target.Sound {
			s.removeAt(target)
		}
		return
	}
	s.createAtWithRequirementConditional(target, "create", "create", inputs, conditional, requirementConditional)
}

// Exact-null assignments and exact exclusions share non-consuming removal evidence.
func (s *semanticStage) removeAt(operand locatedOperand) {
	id := s.operandReference(operand, "field", "remove")
	s.rewriteRemoval(id, operand)
	s.env.removeIdentity(operand.fieldIdentity())
	s.env.requirements.removeIdentity(operand.fieldIdentity())
	outputIdentity := transitionOutputIdentity(operand.fieldIdentity())
	s.appendTransition(Transition{Operation: "remove", Output: operand.Name, OutputIdentity: outputIdentity, InputReferenceIDs: []string{}, OutputReferenceID: id})
}

// applySource starts an independent external dataset without carrying prior fields.
func (s *semanticStage) applySource() {
	s.env = newEnvironmentWithRequirementTrace(s.env.requirements.trace)
}

func (s *semanticStage) appendTransition(transition Transition) {
	s.transitions = append(s.transitions, transition)
}

func (s *semanticStage) fieldIdentityCollision(name string, location Location, pendingReferenceIDs []string) {
	message := fmt.Sprintf("field %q has multiple private identities that schema version 1 cannot distinguish", name)
	s.diagnosticAtOwned(CodeAmbiguousField, "warning", "unsupported_semantics", message, location, true, pendingReferenceIDs)
}
