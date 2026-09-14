package analysis

import (
	"fmt"
	"github.com/antlr4-go/antlr/v4"
	"github.com/delgado-jacob/spl-toolkit/parser"
	"sort"
	"strings"
)

type semanticStage struct {
	rewritePhase   string
	rewriteOrdinal int
	result         *Result
	parsed         *parsedDocument
	stage          int
	env            *environment
	transitions    []Transition
	recoveryLimit  int
	refinement     *sourceRefinement
}

func normalizedName(text string) string {
	if len(text) >= 2 && (text[0] == '\'' || text[0] == '"') && text[len(text)-1] == text[0] {
		text = text[1 : len(text)-1]
		var b strings.Builder
		escape := false
		for _, r := range text {
			if escape {
				b.WriteRune(r)
				escape = false
			} else if r == '\\' {
				escape = true
			} else {
				b.WriteRune(r)
			}
		}
		return b.String()
	}
	return text
}
func intact(node antlr.Tree) bool {
	if _, bad := node.(antlr.ErrorNode); bad {
		return false
	}
	if t, ok := node.(antlr.TerminalNode); ok {
		return t.GetSymbol().GetTokenIndex() >= 0
	}
	for i := 0; i < node.GetChildCount(); i++ {
		if !intact(node.GetChild(i)) {
			return false
		}
	}
	return true
}
func (s *semanticStage) diagnostic(code, message string, ctx antlr.ParserRuleContext) {
	severity, category, incomplete := "warning", "unsupported_semantics", true
	if code == CodeUnavailableField {
		severity, category, incomplete = "error", "unavailable_field", false
	}
	s.diagnosticAt(code, severity, category, message, s.parsed.source.contextLocation(ctx), incomplete)
}
func (s *semanticStage) operand(ctx antlr.ParserRuleContext, name string) locatedOperand {
	if !s.sound(ctx) {
		return locatedOperand{Name: name}
	}
	loc := s.parsed.source.contextLocation(ctx)
	resolution := "exact"
	switch c := ctx.(type) {
	case parser.IAnalysisSelectorContext:
		if selectorPattern(c) {
			resolution = "wildcard"
		}
	case parser.IAnalysisSearchValueContext:
		// Search values are patterns even when double-quoted; this rule does not
		// apply to exact quoted field identifiers in assignment/expression contexts.
		if c.STRING() != nil && strings.Contains(name, "*") {
			resolution = "wildcard"
		}
		if value := c.AnalysisUnquotedValue(); value != nil {
			for _, part := range value.AllAnalysisUnquotedPart() {
				if part.MULT() != nil {
					resolution = "wildcard"
				}
			}
		}
	}
	return locatedOperand{Name: name, Location: loc, Resolution: resolution, Sound: true, rewrite: s.rewriteSPLOwner(ctx)}
}
func (s *semanticStage) reference(ctx antlr.ParserRuleContext, name, kind, role string) string {
	return s.operandReference(s.operand(ctx, name), kind, role)
}
func (s *semanticStage) referenceAt(loc Location, name, kind, role, resolution string) string {
	st := s.result.Stages[s.stage]
	id := fmt.Sprintf("pending-%d", len(s.result.References))
	reference := Reference{ID: id, OriginalName: s.result.Document.Text[loc.Start.Offset:loc.End.Offset], NormalizedName: name, Kind: kind, Role: role, StageID: st.ID, ScopeID: st.ScopeID, Location: loc, Resolution: resolution, Binding: "not_applicable", OriginReferenceIDs: []string{}}
	s.result.References = append(s.result.References, reference)
	if trace := s.env.requirements.trace; trace != nil {
		directExternal, conditional := requirementReferencePolicy(reference)
		trace.recordReference(reference, directExternal, conditional, trace.nextEvent())
	}
	return id
}
func (s *semanticStage) read(ctx antlr.ParserRuleContext, name, role string) string {
	return s.readAt(s.operand(ctx, name), role)
}
func (s *semanticStage) origins(ids []string) []string {
	out := copyIDs(ids)
	for _, id := range ids {
		for _, r := range s.result.References {
			if r.ID == id {
				out = uniqueIDs(out, r.OriginReferenceIDs)
				break
			}
		}
	}
	return out
}
func (s *semanticStage) create(ctx antlr.ParserRuleContext, name, role, operation string, inputs []string, conditional bool) string {
	return s.createAt(s.operand(ctx, name), role, operation, inputs, conditional)
}
func (s *semanticStage) expression(node antlr.Tree) []string {
	ids := []string{}
	if node == nil {
		return ids
	}
	switch ctx := node.(type) {
	case parser.IAnalysisIdentifierContext:
		if intact(ctx) {
			id := s.read(ctx, normalizedName(ctx.GetText()), "read")
			if id != "" {
				ids = append(ids, id)
			}
		}
		return ids
	case parser.IAnalysisLiteralContext:
		return ids
	case parser.IAnalysisFunctionCallContext:
		s.function(ctx, false)
		if ctx.AnalysisArgumentList() != nil {
			return s.expression(ctx.AnalysisArgumentList())
		}
		return ids
	case parser.IAnalysisMacroContext:
		s.diagnostic(CodeDynamicReference, "macro expansion is unresolved", ctx)
		return ids
	case parser.IAnalysisSubqueryContext:
		s.diagnostic(CodeUnsupportedSemantics, "subsearch result field effects are unmodeled", ctx)
		return ids
	case antlr.TerminalNode:
		if _, atom := node.GetParent().(parser.IAnalysisAtomContext); atom && ctx.GetSymbol().GetTokenType() == parser.SPLLexerMULT {
			s.diagnostic(CodeUnresolvedWildcard, "wildcard expression membership is unresolved", node.GetParent().(antlr.ParserRuleContext))
		}
		return ids
	}
	for i := 0; i < node.GetChildCount(); i++ {
		ids = append(ids, s.expression(node.GetChild(i))...)
	}
	return ids
}
func finalizeReferences(r *Result, refinement *sourceRefinement, traces ...*requirementTrace) {
	var trace *requirementTrace
	if len(traces) > 0 {
		trace = traces[0]
	}
	sort.SliceStable(r.References, func(i, j int) bool {
		a, b := r.References[i], r.References[j]
		if a.Location.Start.Offset != b.Location.Start.Offset {
			return a.Location.Start.Offset < b.Location.Start.Offset
		}
		if a.Location.End.Offset != b.Location.End.Offset {
			return a.Location.End.Offset < b.Location.End.Offset
		}
		if a.Kind != b.Kind {
			return a.Kind < b.Kind
		}
		return a.Role < b.Role
	})
	mapping := map[string]string{}
	for i := range r.References {
		ref := &r.References[i]
		mapping[ref.ID] = fmt.Sprintf("ref-%d", i)
		ref.ID = mapping[ref.ID]
	}
	if trace != nil {
		trace.remapReferences(mapping)
	}
	if r.rewrite != nil {
		r.rewrite.finalizeReferences(mapping)
	}
	if refinement != nil {
		refinement.finalizeExpansions(r.References, mapping)
	}
	remap := func(ids []string) {
		for i, id := range ids {
			ids[i] = mapping[id]
		}
	}
	for i := range r.References {
		remap(r.References[i].OriginReferenceIDs)
	}
	for i := range r.Lineage {
		l := &r.Lineage[i]
		for _, state := range []*FieldState{&l.Before, &l.After} {
			for j := range state.Fields {
				remap(state.Fields[j].OriginReferenceIDs)
			}
		}
		for j := range l.Transitions {
			tr := &l.Transitions[j]
			remap(tr.InputReferenceIDs)
			if tr.OutputReferenceID != "" {
				tr.OutputReferenceID = mapping[tr.OutputReferenceID]
			}
		}
	}
	if trace != nil {
		trace.assertReferences(r.References)
	}
	for _, values := range []*[]string{&r.Dependencies.Indexes, &r.Dependencies.Sources, &r.Dependencies.SourceTypes, &r.Dependencies.Lookups, &r.Dependencies.Datasets, &r.Dependencies.DataModels, &r.Dependencies.Macros} {
		sort.Strings(*values)
		out := []string{}
		for _, v := range *values {
			if len(out) == 0 || out[len(out)-1] != v {
				out = append(out, v)
			}
		}
		*values = out
	}
}
