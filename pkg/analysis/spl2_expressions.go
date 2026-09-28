package analysis

import (
	"fmt"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	"github.com/delgado-jacob/spl-toolkit/parser/spl2"
)

type spl2ExpressionEvidence struct {
	ids []string
	// requirementNonnull follows query-only bindings; nonnull may include refinement.
	nonnull, requirementNonnull, exactNull, truth, modeled bool
	domain                                                 string
	template                                               spl2ExpressionTemplate
}

type spl2FunctionParameterUse struct {
	parameter int
	role      string
}

type spl2ExpressionTemplate func(map[spl2FunctionParameterUse]spl2ExpressionEvidence) spl2ExpressionEvidence

func (e spl2ExpressionEvidence) instantiate(bindings map[spl2FunctionParameterUse]spl2ExpressionEvidence) spl2ExpressionEvidence {
	if e.template == nil {
		e.ids = append([]string{}, e.ids...)
		return e
	}
	return e.template(bindings)
}

func spl2TransformExpressionEvidence(value spl2ExpressionEvidence, transform func(spl2ExpressionEvidence) spl2ExpressionEvidence) spl2ExpressionEvidence {
	plain := value
	plain.template = nil
	out := transform(plain)
	out.template = nil
	if value.template != nil {
		out.template = func(bindings map[spl2FunctionParameterUse]spl2ExpressionEvidence) spl2ExpressionEvidence {
			return spl2TransformExpressionEvidence(value.instantiate(bindings), transform)
		}
	}
	return out
}

func spl2CombineExpressionEvidence(values []spl2ExpressionEvidence, combine func([]spl2ExpressionEvidence) spl2ExpressionEvidence) spl2ExpressionEvidence {
	plain := make([]spl2ExpressionEvidence, len(values))
	templated := false
	for i, value := range values {
		plain[i] = value
		plain[i].template = nil
		templated = templated || value.template != nil
	}
	out := combine(plain)
	out.template = nil
	if templated {
		captured := append([]spl2ExpressionEvidence{}, values...)
		out.template = func(bindings map[spl2FunctionParameterUse]spl2ExpressionEvidence) spl2ExpressionEvidence {
			instantiated := make([]spl2ExpressionEvidence, len(captured))
			for i, value := range captured {
				instantiated[i] = value.instantiate(bindings)
			}
			return spl2CombineExpressionEvidence(instantiated, combine)
		}
	}
	return out
}

func (s *spl2SemanticStage) readIdentifier(ctx antlr.ParserRuleContext, role string) string {
	o := s.operand(ctx)
	if !o.Sound {
		return ""
	}
	return s.readAt(o, role)
}

func (s *spl2SemanticStage) expressionRole() string {
	if s.readRole != "" {
		return s.readRole
	}
	return "read"
}

func (s *spl2SemanticStage) expressionWithRole(node antlr.Tree, role string) spl2ExpressionEvidence {
	previous := s.readRole
	s.readRole = role
	defer func() { s.readRole = previous }()
	return s.expression(node)
}

func (s *spl2SemanticStage) expression(node antlr.Tree) spl2ExpressionEvidence {
	out := spl2ExpressionEvidence{ids: []string{}, nonnull: true, requirementNonnull: true, modeled: true}
	if node == nil {
		return out
	}
	switch c := node.(type) {
	case spl2.INotExpressionContext:
		if operator := c.LogicalNot(); operator != nil && operator.GetText() != "NOT" {
			return spl2TransformExpressionEvidence(s.expression(c.NotExpression()), func(value spl2ExpressionEvidence) spl2ExpressionEvidence {
				value.modeled, value.nonnull, value.requirementNonnull, value.exactNull, value.truth, value.domain = false, false, false, false, false, ""
				return value
			})
		}
	case spl2.IUnaryContext:
		if c.Unary() != nil {
			return spl2TransformExpressionEvidence(s.expression(c.Unary()), func(value spl2ExpressionEvidence) spl2ExpressionEvidence {
				value.exactNull = false
				value.truth = false
				if value.domain != "number" {
					value.domain = ""
				}
				return value
			})
		}
		return s.expression(c.Access())
	case spl2.ICallContext:
		if s.program != nil {
			if value, handled := s.program.bindCall(s, c, false); handled {
				return value
			}
		}
		return s.call(c, false)
	case spl2.IAccessContext:
		if len(c.AllAccessPart()) == 0 {
			return s.expression(c.Primary())
		}
		base := c.Primary()
		if base.LOCAL() != nil && s.locals[base.LOCAL().GetText()] {
			for _, part := range c.AllAccessPart() {
				if part.Expression() != nil {
					value := s.expression(part.Expression())
					out.ids = append(out.ids, value.ids...)
					out.modeled = out.modeled && value.modeled
				}
			}
			out.nonnull = false
			out.requirementNonnull = false
			return out
		}
		alias := false
		baseName := ""
		if f := base.FieldName(); f != nil && f.Identifier() != nil {
			baseName = s.operand(f.Identifier()).Name
			alias = s.aliases[baseName]
		}
		if s.program != nil && baseName != "" {
			if binding := s.program.imports[baseName]; binding != nil && binding.container {
				parts := []string{}
				static := true
				for _, part := range c.AllAccessPart() {
					if part.DOT() == nil || part.Identifier() == nil {
						static = false
						break
					}
					parts = append(parts, spl2ProgramIdentifier(part.Identifier()))
				}
				if static {
					s.program.useImport(s, binding, c, strings.Join(parts, "."))
					out.modeled = false
					out.nonnull = false
					out.requirementNonnull = false
					return out
				}
			}
		}
		segments := []string{}
		static := baseName != ""
		for _, part := range c.AllAccessPart() {
			if part.DOT() == nil || part.Identifier() == nil {
				static = false
				break
			}
			segment := s.operand(part.Identifier())
			if !segment.Sound {
				static = false
				break
			}
			segments = append(segments, segment.Name)
		}
		if static {
			qualifier := ""
			if alias {
				qualifier = baseName
			} else {
				segments = append([]string{baseName}, segments...)
			}
			identity := pathFieldIdentity(qualifier, segments)
			o := locatedOperand{Name: identity.PublicName, Identity: identity, Location: s.parsed2.source.contextLocation(c), Resolution: "exact", Sound: spl2IntactSyntax(c), UnresolvedSource: true, rewrite: s.rewriteNavigation(c)}
			id := s.readAt(o, s.expressionRole())
			if id != "" {
				out.ids = append(out.ids, id)
				r := s.result.References[len(s.result.References)-1]
				out.nonnull = r.Binding == "source" || r.Binding == "derived"
				out.requirementNonnull = out.nonnull
				if trace := s.env.requirements.trace; trace != nil {
					entry := trace.reference(id)
					out.requirementNonnull = entry.reference.Binding == "source" || entry.reference.Binding == "derived"
				}
			}
			return out
		}
		if !alias {
			out = s.expression(base)
		}
		for _, part := range c.AllAccessPart() {
			if part.Expression() != nil {
				e := s.expression(part.Expression())
				out.ids = append(out.ids, e.ids...)
			}
		}
		name := baseName
		if name == "" {
			name = base.GetText()
		}
		for _, part := range c.AllAccessPart() {
			if part.Identifier() != nil {
				name += "." + s.operand(part.Identifier()).Name
			} else {
				name += "[]"
			}
		}
		o := locatedOperand{Name: name, Identity: dynamicFieldIdentity(name), Location: s.parsed2.source.contextLocation(c), Resolution: "dynamic", Sound: spl2IntactSyntax(c), rewrite: s.rewriteNavigation(c)}
		if id := s.operandReference(o, "field", s.expressionRole()); id != "" {
			out.ids = append(out.ids, id)
			ref := &s.result.References[len(s.result.References)-1]
			ref.Binding = "indeterminate"
			s.rewriteBinding(id, "indeterminate", nil)
			s.unsupportedOwned(c, "Dynamic field navigation has no exact identity", []string{id})
		} else {
			s.unsupported(c, "Dynamic field navigation has no exact identity")
		}
		out.nonnull = false
		out.requirementNonnull = false
		out.exactNull = false
		out.modeled = false
		return out
	case spl2.IFieldNameContext:
		if c.Identifier() != nil {
			if s.program != nil {
				name := spl2ProgramIdentifier(c.Identifier())
				if binding := s.program.imports[name]; binding != nil {
					s.program.useImport(s, binding, c.Identifier(), "")
					out.modeled = false
					out.nonnull = false
					out.requirementNonnull = false
					return out
				}
				if s.program.functions[name] != nil {
					s.diagnosticAtOwned(CodeUnresolvedSymbol, "error", "contract", fmt.Sprintf("function %q is not a scalar value", name), s.parsed2.source.contextLocation(c), true, nil)
					out.modeled = false
					out.nonnull = false
					out.requirementNonnull = false
					return out
				}
			}
			id := s.readIdentifier(c.Identifier(), s.expressionRole())
			if id != "" {
				out.ids = append(out.ids, id)
				r := s.result.References[len(s.result.References)-1]
				out.nonnull = r.Binding == "source" || r.Binding == "derived"
				out.requirementNonnull = out.nonnull
				if trace := s.env.requirements.trace; trace != nil {
					entry := trace.reference(id)
					out.requirementNonnull = entry.reference.Binding == "source" || entry.reference.Binding == "derived"
				}
			}
			return out
		}
		out = s.expression(c.FieldTemplate())
		out.nonnull = false
		out.requirementNonnull = false
		return out
	case spl2.ILiteralContext:
		if c.NULL() != nil {
			out.nonnull = false
			out.requirementNonnull = false
			out.exactNull = true
			return out
		}
		if c.NUMBER() != nil {
			out.domain = "number"
		}
		if c.RAW_STRING() != nil {
			out.domain = "string"
		}
		if c.BOOLEAN() != nil {
			out.domain = "boolean"
		}
		out.truth = c.BOOLEAN() != nil && c.BOOLEAN().GetText() == "true"
		if c.StringLiteral() != nil {
			return s.expression(c.StringLiteral())
		}
		return out
	case spl2.IStringLiteralContext:
		values := []spl2ExpressionEvidence{}
		for _, e := range c.AllExpression() {
			values = append(values, s.expression(e))
		}
		// Intact interpolation renders a string even when an operand is null.
		// Unmodeled children still prevent a proven assignment effect.
		return spl2CombineExpressionEvidence(values, func(values []spl2ExpressionEvidence) spl2ExpressionEvidence {
			out := spl2ExpressionEvidence{ids: []string{}, nonnull: true, requirementNonnull: true, modeled: true, domain: "string"}
			for _, value := range values {
				out.ids = append(out.ids, value.ids...)
				out.modeled = out.modeled && value.modeled
			}
			out.nonnull = out.modeled && spl2IntactSyntax(c)
			out.requirementNonnull = out.nonnull
			return out
		})
	case spl2.IFieldTemplateContext:
		s.rewriteUnprovedOperand(locatedOperand{Location: s.parsed2.source.contextLocation(c)}, "field", "dynamic_name", "dynamic_identity")
		for _, e := range c.AllExpression() {
			value := s.expression(e)
			out.ids = append(out.ids, value.ids...)
			out.modeled = out.modeled && value.modeled
		}
		out.modeled = false
		s.unsupported(c, "Computed field name is unresolved")
		out.nonnull = false
		out.requirementNonnull = false
		return out
	case spl2.IArrayContext:
		values := []spl2ExpressionEvidence{}
		for _, e := range c.AllExpression() {
			values = append(values, s.expression(e))
		}
		return spl2CombineExpressionEvidence(values, func(values []spl2ExpressionEvidence) spl2ExpressionEvidence {
			out := spl2ExpressionEvidence{ids: []string{}, nonnull: true, requirementNonnull: true, modeled: true, domain: "array"}
			for _, value := range values {
				out.ids = append(out.ids, value.ids...)
				out.modeled = out.modeled && value.modeled
			}
			return out
		})
	case spl2.IObjectContext:
		values := []spl2ExpressionEvidence{}
		for _, e := range c.AllObjectEntry() {
			values = append(values, s.expression(e.Expression()))
		}
		return spl2CombineExpressionEvidence(values, func(values []spl2ExpressionEvidence) spl2ExpressionEvidence {
			out := spl2ExpressionEvidence{ids: []string{}, nonnull: true, requirementNonnull: true, modeled: true, domain: "object"}
			for _, value := range values {
				out.ids = append(out.ids, value.ids...)
				out.modeled = out.modeled && value.modeled
			}
			return out
		})
	case spl2.ILambdaExpressionContext:
		for _, child := range c.GetChildren() {
			switch child.(type) {
			case spl2.IExpressionContext, spl2.ILambdaBlockContext:
				value := s.expression(child)
				out.ids = append(out.ids, value.ids...)
				out.modeled = out.modeled && value.modeled
			}
		}
		out.modeled = false
		s.unsupported(c, "Lambda result effects are unmodeled")
		out.nonnull = false
		out.requirementNonnull = false
		return out
	case spl2.IExistsPredicateContext, spl2.ISearchLiteralContext:
		out.modeled = false
		s.unsupported(node.(antlr.ParserRuleContext), "Child search expression result effects are not yet modeled")
		out.nonnull = false
		out.requirementNonnull = false
		return out
	case spl2.IIdentifierContext, spl2.IObjectKeyContext, spl2.ILambdaParameterContext:
		return out // Names/labels are read only in their owning field role.
	case spl2.IPrimaryContext:
		if c.LOCAL() != nil {
			if s.functionSummary != nil {
				if value, ok := s.functionSummary.parameter(c.LOCAL().GetText(), s.expressionRole()); ok {
					return value
				}
			}
			if s.locals[c.LOCAL().GetText()] {
				out.nonnull = false
				out.requirementNonnull = false
				return out
			}
			if len(s.locals) > 0 {
				if s.program != nil {
					s.diagnosticAtOwned(CodeUnresolvedSymbol, "error", "contract", fmt.Sprintf("local symbol %q is unresolved", c.LOCAL().GetText()), s.parsed2.source.contextLocation(c), true, nil)
				} else {
					s.unsupported(c, "Unbound lambda local is unresolved")
				}
			} else if s.program != nil {
				s.diagnosticAtOwned(CodeUnresolvedSymbol, "error", "contract", fmt.Sprintf("local symbol %q is unresolved", c.LOCAL().GetText()), s.parsed2.source.contextLocation(c), true, nil)
			}
			out.nonnull = false
			out.requirementNonnull = false
			out.modeled = false
			return out
		}
		for _, child := range c.GetChildren() {
			if _, ok := child.(antlr.ParserRuleContext); ok {
				return s.expression(child)
			}
		}
		return out
	case antlr.TerminalNode:
		if c.GetSymbol().GetTokenType() == spl2.SPL2ParserLOCAL {
			out.nonnull = false
			out.requirementNonnull = false
		}
		return out
	}
	values := []spl2ExpressionEvidence{}
	for _, child := range node.GetChildren() {
		if _, terminal := child.(antlr.TerminalNode); terminal {
			continue
		}
		values = append(values, s.expression(child))
	}
	return spl2CombineExpressionEvidence(values, func(values []spl2ExpressionEvidence) spl2ExpressionEvidence {
		out := spl2ExpressionEvidence{ids: []string{}, nonnull: true, requirementNonnull: true, modeled: true}
		for i, value := range values {
			out.ids = append(out.ids, value.ids...)
			out.modeled = out.modeled && value.modeled
			out.nonnull = out.nonnull && value.nonnull
			out.requirementNonnull = out.requirementNonnull && value.requirementNonnull
			if i == 0 {
				out.exactNull = value.exactNull
				out.truth = value.truth
				out.domain = value.domain
			} else {
				out.exactNull = false
				out.truth = false
				out.domain = ""
			}
		}
		// Only routing and parentheses preserve the exact literal-null identity.
		if node.GetChildCount() > 1 {
			switch node.(type) {
			case spl2.IPrimaryContext:
			default:
				out.exactNull = false
				out.truth = false
				out.domain = ""
			}
		}
		return out
	})
}
