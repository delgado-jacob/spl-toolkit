package analysis

import (
	"fmt"
	"sort"
	"strings"
)

// locatedOperand is decoded and checked for soundness by its language frontend.
type locatedOperand struct {
	Name       string
	Location   Location
	Resolution string
	Sound      bool
}
type renameOperands struct{ Source, Target locatedOperand }
type aggregateOutput struct {
	Target            locatedOperand
	InputReferenceIDs []string
	Conditional       bool
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
	st := &s.result.Stages[s.stage]
	if incomplete {
		st.SemanticComplete = false
		s.env.uncertain = true
	}
	s.result.Diagnostics = append(s.result.Diagnostics, Diagnostic{Code: code, Severity: severity, Category: category, Message: message, Location: location, StageID: st.ID, ScopeID: st.ScopeID})
}
func (s *semanticStage) operandReference(operand locatedOperand, kind, role string) string {
	if !operand.Sound {
		return ""
	}
	return s.referenceAt(operand.Location, operand.Name, kind, role, operand.Resolution)
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
		s.diagnosticAt(CodeUnavailableField, "error", "unavailable_field", fmt.Sprintf("field %q is unavailable after an earlier pipeline transfer", name), operand.Location, false)
	default:
		ref.Binding = "source"
		s.env.fields[name] = trackedField{FieldBinding: FieldBinding{Name: name, OriginReferenceIDs: []string{id}}, source: true}
	}
	return id
}
func (s *semanticStage) createAt(operand locatedOperand, role, operation string, inputs []string, conditional bool) string {
	name := operand.Name
	id := s.operandReference(operand, "field", role)
	if id == "" {
		return id
	}
	origins := s.origins(inputs)
	s.result.References[len(s.result.References)-1].OriginReferenceIDs = copyIDs(origins)
	s.env.install(name, uniqueIDs([]string{id}, origins), conditional)
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
		names := s.refinedSelectorAt(operand, role, id)
		if !allowWildcard {
			s.diagnosticAt(CodeUnsupportedSemantics, "warning", "unsupported_semantics", fmt.Sprintf("wildcard selectors for command %q are unmodeled", command), operand.Location, true)
			return []string{}, []string{id}
		}
		return names, []string{id}
	}
	if !allowWildcard {
		s.diagnosticAt(CodeUnsupportedSemantics, "warning", "unsupported_semantics", fmt.Sprintf("wildcard selectors for command %q are unmodeled", command), operand.Location, true)
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
		s.diagnosticAt(CodeUnresolvedWildcard, "warning", "unsupported_semantics", fmt.Sprintf("wildcard %q membership is unresolved", name), operand.Location, true)
	}
	return names, []string{id}
}

// applyProjection prepares lexical reads once before restricting the environment.
// mode is the frontend's finite policy: include, exclude, or table.
func (s *semanticStage) applyProjection(selectors []locatedOperand, mode string, retainKnownInternals bool) {
	exclude := mode == "exclude"
	selected := []preparedSelection{}
	if !exclude && retainKnownInternals {
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
			s.diagnosticAt(CodeUnsupportedSemantics, "warning", "unsupported_semantics", "fields inclusion retains internal fields with unresolved open-source membership", s.result.Stages[s.stage].Location, true)
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
		source, dest string
		target       locatedOperand
		input        string
		conditional  bool
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
		items = append(items, rename{src, dst, r.Target, id, binding == "indeterminate" || binding == "unavailable"})
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
	}
	for _, r := range items {
		s.createAt(r.target, "rename", "rename", []string{r.input}, r.conditional)
	}
}
func (s *semanticStage) applyAggregation(outputs []aggregateOutput, groups []locatedOperand, preserveInput bool) {
	output := newEnvironment()
	output.open = false
	for _, operand := range groups {
		names, ids := s.selectorAt(operand, "group", false)
		for _, name := range names {
			if field, ok := s.projectedField(name, ids); ok {
				output.fields[name] = field
			}
			s.transitions = append(s.transitions, Transition{Operation: "project", Output: name, InputReferenceIDs: copyIDs(ids)})
		}
	}
	if !preserveInput {
		output.uncertain = !s.result.Stages[s.stage].SemanticComplete
		s.env = output
	}
	for _, output := range outputs {
		s.createAt(output.Target, "output", "aggregate", output.InputReferenceIDs, output.Conditional || !s.result.Stages[s.stage].SemanticComplete)
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
		s.createAt(output.Target, "output", "lookup", matchReferenceIDs, output.PreserveExisting || !s.result.Stages[s.stage].SemanticComplete)
	}
}

func (s *semanticStage) applyAssignment(target locatedOperand, inputs []string, conditional, removeNull bool) {
	if removeNull {
		if target.Sound {
			s.removeAt(target)
		}
		return
	}
	s.createAt(target, "create", "create", inputs, conditional)
}

// Exact-null assignments and exact exclusions share non-consuming removal evidence.
func (s *semanticStage) removeAt(operand locatedOperand) {
	id := s.operandReference(operand, "field", "remove")
	s.env.remove(operand.Name)
	s.transitions = append(s.transitions, Transition{Operation: "remove", Output: operand.Name, InputReferenceIDs: []string{}, OutputReferenceID: id})
}
