package analysis

import (
	"fmt"
	"github.com/antlr4-go/antlr/v4"
	"github.com/delgado-jacob/spl-toolkit/parser"
	"sort"
	"strings"
)

type semanticStage struct {
	result      *Result
	parsed      *parsedDocument
	stage       int
	env         *environment
	transitions []Transition
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
	st := &s.result.Stages[s.stage]
	severity, category := "warning", "unsupported_semantics"
	if code == CodeUnavailableField {
		severity, category = "error", "unavailable_field"
	} else {
		st.SemanticComplete = false
		s.env.uncertain = true
	}
	s.result.Diagnostics = append(s.result.Diagnostics, Diagnostic{Code: code, Severity: severity, Category: category, Message: message, Location: s.parsed.source.contextLocation(ctx), StageID: st.ID, ScopeID: st.ScopeID})
}
func (s *semanticStage) reference(ctx antlr.ParserRuleContext, name, kind, role string) string {
	if !intact(ctx) {
		return ""
	}
	st := s.result.Stages[s.stage]
	loc := s.parsed.source.contextLocation(ctx)
	resolution := "exact"
	if strings.Contains(name, "*") {
		resolution = "wildcard"
	}
	id := fmt.Sprintf("pending-%d", len(s.result.References))
	s.result.References = append(s.result.References, Reference{ID: id, OriginalName: s.result.Document.Text[loc.Start.Offset:loc.End.Offset], NormalizedName: name, Kind: kind, Role: role, StageID: st.ID, ScopeID: st.ScopeID, Location: loc, Resolution: resolution, Binding: "not_applicable", OriginReferenceIDs: []string{}})
	return id
}
func (s *semanticStage) read(ctx antlr.ParserRuleContext, name, role string) string {
	id := s.reference(ctx, name, "field", role)
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
		s.diagnostic(CodeUnavailableField, fmt.Sprintf("field %q is unavailable after an earlier pipeline transfer", name), ctx)
	default:
		ref.Binding = "source"
		s.env.fields[name] = trackedField{FieldBinding: FieldBinding{Name: name, OriginReferenceIDs: []string{id}}, source: true}
	}
	return id
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
	id := s.reference(ctx, name, "field", role)
	if id == "" {
		return id
	}
	origins := s.origins(inputs)
	s.result.References[len(s.result.References)-1].OriginReferenceIDs = copyIDs(origins)
	s.env.install(name, uniqueIDs([]string{id}, origins), conditional)
	s.transitions = append(s.transitions, Transition{Operation: operation, Output: name, InputReferenceIDs: copyIDs(inputs), OutputReferenceID: id, Conditional: conditional})
	return id
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
func (s *semanticStage) dependency(ctx antlr.ParserRuleContext, name, kind string) {
	s.reference(ctx, name, kind, "read")
	switch kind {
	case "index":
		s.result.Dependencies.Indexes = append(s.result.Dependencies.Indexes, name)
	case "source":
		s.result.Dependencies.Sources = append(s.result.Dependencies.Sources, name)
	case "sourcetype":
		s.result.Dependencies.SourceTypes = append(s.result.Dependencies.SourceTypes, name)
	case "lookup":
		s.result.Dependencies.Lookups = append(s.result.Dependencies.Lookups, name)
	}
}
func finalizeReferences(r *Result) {
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
