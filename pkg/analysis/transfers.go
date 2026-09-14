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
	Location   Location
	Resolution string
	Sound      bool
	// UnresolvedSource preserves frontend distinctions absent from string universes.
	// Derived bindings and structurally proven absence remain independently sound.
	UnresolvedSource bool
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
	if trace := s.env.requirements.trace; trace != nil {
		entry := trace.reference(id)
		if kind == "field" {
			entry.reference.Binding, entry.directExternal, entry.conditional = s.env.requirements.read(entry.reference)
		} else {
			entry.directExternal, entry.conditional = requirementReferencePolicy(entry.reference)
		}
	}
	s.rewriteReference(id, operand, kind, role)
	return id
}
func (s *semanticStage) readAt(operand locatedOperand, role string) string {
	name := operand.Name
	id := s.operandReference(operand, "field", role)
	if id == "" {
		return id
	}
	ref := &s.result.References[len(s.result.References)-1]
	f, known := s.env.fields[name]
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
		}
	case s.env.removed[name] || !s.env.open:
		ref.Binding = "unavailable"
		if role != "null_test" {
			s.diagnosticAtOwned(CodeUnavailableField, "error", "unavailable_field", fmt.Sprintf("field %q is unavailable after an earlier pipeline transfer", name), operand.Location, false, []string{id})
		}
	default:
		ref.Binding = "source"
		if role != "null_test" {
			s.env.fields[name] = trackedField{FieldBinding: FieldBinding{Name: name, OriginReferenceIDs: []string{id}}, source: true}
		}
	}
	if operand.UnresolvedSource && s.refinement != nil && s.refinement.resolve != nil && ref.Binding == "source" {
		ref.Binding = "indeterminate"
		if role != "null_test" {
			field := s.env.fields[name]
			field.Conditional = true
			s.env.fields[name] = field
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
	s.env.install(name, uniqueIDs([]string{id}, origins), conditional)
	if trace := s.env.requirements.trace; trace != nil {
		entry := trace.reference(id)
		entry.reference.Binding = "definition"
		entry.reference.OriginReferenceIDs = traceOrigins(trace, inputs)
		s.env.requirements.install(name, uniqueIDs([]string{id}, entry.reference.OriginReferenceIDs), requirementConditional)
	}
	s.transitions = append(s.transitions, Transition{Operation: operation, Output: name, InputReferenceIDs: copyIDs(inputs), OutputReferenceID: id, Conditional: conditional})
	return id
}
func (s *semanticStage) selectorAt(operand locatedOperand, role string, allowWildcard bool) ([]string, []string) {
	name := operand.Name
	if operand.Resolution != "wildcard" {
		return []string{name}, []string{s.readAt(operand, role)}
	}
	id := s.operandReference(operand, "field", role)
	if id == "" {
		return []string{}, []string{}
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
		names := s.refinedSelectorAt(operand, role, id)
		if !allowWildcard {
			s.refinementDiagnosticAt(CodeUnsupportedSemantics, "warning", "unsupported_semantics", fmt.Sprintf("wildcard selectors for command %q are unmodeled", command), operand.Location, true)
			return []string{}, []string{id}
		}
		return names, []string{id}
	}
	if !allowWildcard {
		s.diagnosticAtOwned(CodeUnsupportedSemantics, "warning", "unsupported_semantics", fmt.Sprintf("wildcard selectors for command %q are unmodeled", command), operand.Location, true, []string{id})
		return []string{}, []string{id}
	}
	names := []string{}
	origins := []string{}
	for n, f := range s.env.fields {
		if wildcardMatches(name, n) {
			names = append(names, n)
			origins = uniqueIDs(origins, f.OriginReferenceIDs)
		}
	}
	sort.Strings(names)
	sort.Strings(origins)
	ref.OriginReferenceIDs = origins
	if s.env.open || s.env.uncertain {
		s.diagnosticAtOwned(CodeUnresolvedWildcard, "warning", "unsupported_semantics", fmt.Sprintf("wildcard %q membership is unresolved", name), operand.Location, true, []string{id})
	}
	return names, []string{id}
}

// applyProjection prepares lexical reads once before restricting the environment.
// mode is the frontend's finite policy: include, exclude, or table.
func (s *semanticStage) applyProjection(selectors []locatedOperand, mode string, retainKnownInternals bool) {
	exclude := mode == "exclude"
	selected := []preparedSelection{}
	if !exclude && retainKnownInternals {
		if s.refinement != nil && s.env.requirements.open {
			s.requirementDiagnosticAt(CodeUnsupportedSemantics, "warning", "unsupported_semantics", "fields inclusion retains internal fields with unresolved open-source membership", s.result.Stages[s.stage].Location, true, nil)
		}
		internalsComplete := true
		if s.refinement != nil {
			internalsComplete = s.retainSourceInternals()
		}
		for name, field := range s.env.fields {
			if strings.HasPrefix(name, "_") {
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
	for _, operand := range selectors {
		if exclude && operand.Resolution != "wildcard" {
			s.removeAt(operand)
			continue
		}
		role := "read"
		if exclude {
			role = "remove"
		}
		names, ids := s.selectorAt(operand, role, true)
		for _, name := range names {
			if exclude {
				s.env.remove(name)
				s.transitions = append(s.transitions, Transition{Operation: "remove", Output: name, InputReferenceIDs: copyIDs(ids)})
			} else if f, ok := s.projectedField(name, ids); ok {
				selected = append(selected, preparedSelection{Field: f, InputReferenceIDs: ids, EmitProjectTransition: true})
			}
		}
	}
	s.env.requirements.applyProjection(selectors, mode, retainKnownInternals, s.result.Stages[s.stage].ID)
	if !exclude {
		s.applyPreparedProjection(selected, mode)
	}
}

// applyPreparedProjection installs copied binding evidence without resolving it
// again. Ordered records intentionally retain repeated explicit projections.
func (s *semanticStage) applyPreparedProjection(selected []preparedSelection, mode string) {
	fields := map[string]trackedField{}
	for _, selection := range selected {
		field := selection.Field
		field.OriginReferenceIDs = copyIDs(field.OriginReferenceIDs)
		fields[field.Name] = field
		if selection.EmitProjectTransition {
			s.transitions = append(s.transitions, Transition{Operation: "project", Output: field.Name, InputReferenceIDs: copyIDs(selection.InputReferenceIDs)})
		}
	}
	s.env.rewriteProject(fields)
	s.env.fields = fields
	// Partial selectors retain an unknown remainder; finite compatibility keeps
	// its historical closed-output wire shape. Existing tombstones are retained.
	s.env.open = s.refinement != nil && !s.refinement.finiteCompatibility && s.env.open && !s.result.Stages[s.stage].SemanticComplete
	if mode == "table" && s.result.Stages[s.stage].SemanticComplete {
		s.env.uncertain = false
	}
}

// Exact projection fixes the output names without proving conditional inputs exist.
func (s *semanticStage) projectedField(name string, ids []string) (trackedField, bool) {
	if field, known := s.env.fields[name]; known {
		return field, true
	}
	if s.env.uncertain {
		return trackedField{FieldBinding: FieldBinding{Name: name, OriginReferenceIDs: s.origins(ids), Conditional: true}}, true
	}
	return trackedField{}, false
}

func (s *semanticStage) applyRename(pairs []renameOperands) {
	original := s.env
	before := s.env.clone()
	type rename struct {
		source, dest           string
		target                 locatedOperand
		input                  string
		conditional            bool
		requirementConditional bool
	}
	items := []rename{}
	sources, dests := map[string]bool{}, map[string]bool{}
	conflict := false
	for _, r := range pairs {
		if !r.Source.Sound || !r.Target.Sound {
			continue
		}
		src := r.Source.Name
		dst := r.Target.Name
		if r.Source.Resolution == "wildcard" || r.Target.Resolution == "wildcard" {
			s.env = before
			names, _ := s.selectorAt(r.Source, "read", true)
			for _, name := range names {
				sources[name] = true
			}
			dests[dst] = true
			s.diagnosticAt(CodeUnsupportedSemantics, "warning", "unsupported_semantics", "wildcard rename substitution is unmodeled", Location{Start: r.Source.Location.Start, End: r.Target.Location.End}, true)
			conflict = true
			continue
		}
		s.env = before
		id := s.readAt(r.Source, "read")
		_, existingDestination := original.fields[dst]
		if sources[src] || dests[dst] || existingDestination {
			conflict = true
		}
		sources[src] = true
		dests[dst] = true
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
	for name := range sources {
		if dests[name] {
			conflict = true
		}
	}
	if conflict {
		s.diagnosticAt(CodeUnsupportedSemantics, "warning", "unsupported_semantics", "rename has conflicting or unsupported source/destination mappings", s.result.Stages[s.stage].Location, true)
		for name := range sources {
			delete(s.env.fields, name)
		}
		for name := range dests {
			delete(s.env.fields, name)
		}
		return
	}
	for _, r := range items {
		s.env.remove(r.source)
		s.env.requirements.remove(r.source)
	}
	for _, r := range items {
		s.createAtWithRequirementConditional(r.target, "rename", "rename", []string{r.input}, r.conditional, r.requirementConditional)
	}
}
func (s *semanticStage) applyAggregation(outputs []aggregateOutput, groups []locatedOperand, preserveInput bool) {
	output := newEnvironmentWithRequirementTrace(s.env.requirements.trace)
	output.open = false
	output.requirements.open = false
	for _, operand := range groups {
		names, ids := s.selectorAt(operand, "group", false)
		if field, ok := s.env.requirements.fields[operand.Name]; ok && operand.Resolution == "exact" {
			field.origins = append([]string{}, field.origins...)
			output.requirements.fields[operand.Name] = field
		}
		for _, name := range names {
			if field, ok := s.projectedField(name, ids); ok {
				output.fields[name] = field
			}
			s.transitions = append(s.transitions, Transition{Operation: "project", Output: name, InputReferenceIDs: copyIDs(ids)})
		}
	}
	if !preserveInput {
		output.rewrite = s.env.rewrite.clone()
		output.rewriteProject(output.fields)
		output.uncertain = !s.result.Stages[s.stage].SemanticComplete
		output.requirements.uncertain = s.env.requirements.uncertain
		s.env = output
	}
	for _, output := range outputs {
		stageID := s.result.Stages[s.stage].ID
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
			if _, known := s.env.fields[output.Target.Name]; known {
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
	s.env.remove(operand.Name)
	s.env.requirements.remove(operand.Name)
	s.transitions = append(s.transitions, Transition{Operation: "remove", Output: operand.Name, InputReferenceIDs: []string{}, OutputReferenceID: id})
}

// applySource starts an independent external dataset without carrying prior fields.
func (s *semanticStage) applySource() {
	s.env = newEnvironmentWithRequirementTrace(s.env.requirements.trace)
}
