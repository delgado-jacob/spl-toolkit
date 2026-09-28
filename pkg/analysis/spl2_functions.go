package analysis

import (
	"fmt"
	"github.com/antlr4-go/antlr/v4"
	"github.com/delgado-jacob/spl-toolkit/parser/spl2"
	"strings"
)

type spl2FunctionSpec struct {
	min, max  int
	aggregate bool
}

// Finite approved positional signatures. SPL's independent registry is unchanged.
var spl2Functions = map[string]spl2FunctionSpec{
	"abs": {1, 1, false}, "ceil": {1, 1, false}, "ceiling": {1, 1, false}, "floor": {1, 1, false}, "round": {1, 2, false},
	"len": {1, 1, false}, "lower": {1, 1, false}, "upper": {1, 1, false}, "trim": {1, 2, false}, "ltrim": {1, 2, false}, "rtrim": {1, 2, false}, "substr": {2, 3, false}, "replace": {3, 3, false},
	"coalesce": {1, -1, false}, "if": {3, 3, false}, "case": {2, -1, false}, "match": {2, 2, false}, "isnull": {1, 1, false}, "isnotnull": {1, 1, false}, "tonumber": {1, 2, false}, "tostring": {1, 2, false}, "mvcount": {1, 1, false}, "split": {2, 2, false},
	"count": {0, 1, true}, "sum": {1, 1, true}, "avg": {1, 1, true}, "min": {1, 1, true}, "max": {1, 1, true}, "dc": {1, 1, true}, "distinct_count": {1, 1, true}, "values": {1, 1, true}, "list": {1, 1, true}, "first": {1, 1, true}, "last": {1, 1, true},
}

func (s *spl2SemanticStage) call(c spl2.ICallContext, aggregate bool) spl2ExpressionEvidence {
	return s.callWithExpression(c, aggregate, s.expression)
}

// SQL compound preparation supplies aggregate-aware operands; all call contracts
// and direct null-test/wildcard handling still use this single shared policy.
func (s *spl2SemanticStage) callWithExpression(c spl2.ICallContext, aggregate bool, expression func(antlr.Tree) spl2ExpressionEvidence) spl2ExpressionEvidence {
	if s.program != nil {
		if value, handled := s.program.bindCall(s, c, aggregate); handled {
			return value
		}
	}
	out := spl2ExpressionEvidence{ids: []string{}}
	name := c.Identifier().GetText()
	policy, selected := s.typedPolicy()
	function, known := policy.functions[name]
	values := []spl2ExpressionEvidence{}
	named := false
	if args := c.Arguments(); args != nil {
		for index, expr := range args.AllExpression() {
			var v spl2ExpressionEvidence
			access := spl2SQLFieldAccess(expr)
			if known && function.lambda && index == 1 {
				v = s.selectedLambda(expr)
			} else if !aggregate && known && function.nullTest && spl2IntactSyntax(c) && len(args.AllExpression()) == 1 && len(args.AllNamedArgument()) == 0 && access != nil && len(access.AllAccessPart()) == 0 && access.Primary().FieldName() != nil && access.Primary().FieldName().Identifier() != nil {
				id := s.readIdentifier(access.Primary().FieldName().Identifier(), "null_test")
				v = spl2ExpressionEvidence{ids: []string{id}, modeled: true}
			} else if !aggregate && known && function.nullTest && spl2IntactSyntax(c) && len(args.AllExpression()) == 1 && len(args.AllNamedArgument()) == 0 && access != nil && len(access.AllAccessPart()) == 0 && access.Primary().LOCAL() != nil && s.functionSummary != nil {
				v = s.expressionWithRole(expr, "null_test")
			} else if aggregate && access != nil && len(access.AllAccessPart()) == 0 && access.Primary().FieldName() != nil && access.Primary().FieldName().Identifier() != nil && strings.Contains(s.operand(access.Primary().FieldName().Identifier()).Name, "*") {
				_, ids := s.selectorAt(s.selector(access.Primary().FieldName().Identifier()), "read", false)
				v = spl2ExpressionEvidence{ids: ids}
			} else {
				v = expression(expr)
			}
			values = append(values, v)
			out.ids = append(out.ids, v.ids...)
		}
		for _, arg := range args.AllNamedArgument() {
			named = true
			v := expression(arg.Expression())
			out.ids = append(out.ids, v.ids...)
		}
	}
	if name == "batch_id" || name == "batch_time" {
		s.result.Coverage.SyntaxComplete = false
		s.diagnosticAt("SPL_PROFILE_MISMATCH", "error", "compatibility", fmt.Sprintf("function %q is unavailable in the splunkd profile", name), s.parsed2.source.contextLocation(c), true)
		return out
	}
	if !selected || !known {
		s.diagnosticAt(CodeUnsupportedFunction, "warning", "unsupported_semantics", fmt.Sprintf("function %q has unmodeled semantics", name), s.parsed2.source.contextLocation(c), true)
		return out
	}
	spec := function.signature
	if named {
		s.unsupported(c, "Named function argument effects are not yet modeled")
		return out
	}
	if !aggregate && function.unmodeledScalar {
		s.unsupported(c, "Scalar min/max signatures are outside the approved positional core")
		return out
	}
	n := len(values)
	if n < spec.min || spec.max >= 0 && n > spec.max || function.pairedArguments && n%2 != 0 || spec.aggregate != aggregate {
		s.diagnosticAt(CodeSyntaxError, "error", "contract", fmt.Sprintf("function %q has invalid positional arity or context", name), s.parsed2.source.contextLocation(c), true)
		return out
	}
	return spl2SelectedFunctionEvidence(function, values)
}

func spl2SelectedFunctionEvidence(function spl2FunctionPolicy, values []spl2ExpressionEvidence) spl2ExpressionEvidence {
	return spl2CombineExpressionEvidence(values, func(values []spl2ExpressionEvidence) spl2ExpressionEvidence {
		out := spl2ExpressionEvidence{ids: []string{}, modeled: true}
		for _, value := range values {
			out.ids = append(out.ids, value.ids...)
			out.modeled = out.modeled && value.modeled
		}
		switch function.nullability {
		case spl2NullabilityAlways:
			out.nonnull = out.modeled
			out.requirementNonnull = out.modeled
		case spl2NullabilityAllArguments:
			out.nonnull = out.modeled
			out.requirementNonnull = out.modeled
			for _, value := range values {
				out.nonnull = out.nonnull && value.nonnull
				out.requirementNonnull = out.requirementNonnull && value.requirementNonnull
			}
		case spl2NullabilityCoalesce:
			out.exactNull = out.modeled
			for _, value := range values {
				out.nonnull = out.nonnull || value.nonnull
				out.requirementNonnull = out.requirementNonnull || value.requirementNonnull
				out.exactNull = out.exactNull && value.exactNull
			}
		case spl2NullabilityIfBranches:
			out.nonnull = values[1].nonnull && values[2].nonnull && out.modeled
			out.requirementNonnull = values[1].requirementNonnull && values[2].requirementNonnull && out.modeled
			out.exactNull = values[1].exactNull && values[2].exactNull && out.modeled
		case spl2NullabilityCaseBranches:
			out.nonnull = out.modeled
			out.requirementNonnull = out.modeled
			out.exactNull = out.modeled
			fallback := false
			for i := 0; i < len(values); i += 2 {
				out.nonnull = out.nonnull && values[i+1].nonnull
				out.requirementNonnull = out.requirementNonnull && values[i+1].requirementNonnull
				out.exactNull = out.exactNull && values[i+1].exactNull
				if values[i].truth {
					fallback = true
					break
				}
			}
			out.nonnull = out.nonnull && fallback
			out.requirementNonnull = out.requirementNonnull && fallback
		case spl2NullabilityNumericArguments:
			out.nonnull = out.modeled
			out.requirementNonnull = out.modeled
			for _, value := range values {
				out.nonnull = out.nonnull && value.nonnull && value.domain == "number"
				out.requirementNonnull = out.requirementNonnull && value.requirementNonnull && value.domain == "number"
			}
		case spl2NullabilityStringArguments:
			out.nonnull = out.modeled
			out.requirementNonnull = out.modeled
			for _, value := range values {
				out.nonnull = out.nonnull && value.nonnull && value.domain == "string"
				out.requirementNonnull = out.requirementNonnull && value.requirementNonnull && value.domain == "string"
			}
		case spl2NullabilitySubstringArguments:
			out.nonnull = out.modeled && values[0].nonnull && values[0].domain == "string"
			out.requirementNonnull = out.modeled && values[0].requirementNonnull && values[0].domain == "string"
			for _, value := range values[1:] {
				out.nonnull = out.nonnull && value.nonnull && value.domain == "number"
				out.requirementNonnull = out.requirementNonnull && value.requirementNonnull && value.domain == "number"
			}
		case spl2NullabilityFirstArgument:
			out.nonnull = out.modeled && values[0].nonnull
			out.requirementNonnull = out.modeled && values[0].requirementNonnull
		case spl2NullabilityConservative:
			// Aggregate operand presence, conversions, regex matching, and multivalue
			// selection require more than argument presence to prove a value.
		}
		if function.returnKind != "unknown" {
			out.domain = function.returnKind
		}
		out.nonnull = out.nonnull && out.modeled
		out.requirementNonnull = out.requirementNonnull && out.modeled
		return out
	})
}

func (s *spl2SemanticStage) selectedLambda(tree antlr.Tree) spl2ExpressionEvidence {
	expression, ok := tree.(spl2.IExpressionContext)
	if !ok {
		return spl2ExpressionEvidence{ids: []string{}}
	}
	if expression.LambdaExpression() == nil {
		value := s.expression(expression)
		s.diagnosticAt(CodeSyntaxError, "error", "contract", "function \"any\" requires an expression lambda as its second argument", s.parsed2.source.contextLocation(expression), true)
		value.modeled = false
		value.nonnull = false
		value.requirementNonnull = false
		return value
	}
	lambda := expression.LambdaExpression()
	parameters := lambda.AllLambdaParameter()
	if len(parameters) != 1 || parameters[0].LOCAL() == nil || lambda.Expression() == nil || lambda.LambdaBlock() != nil {
		s.diagnosticAt(CodeSyntaxError, "error", "contract", "function \"any\" requires one expression lambda parameter", s.parsed2.source.contextLocation(lambda), true)
		return spl2ExpressionEvidence{ids: []string{}}
	}
	previous := s.locals
	s.locals = make(map[string]bool, len(previous)+1)
	for name, bound := range previous {
		s.locals[name] = bound
	}
	s.locals[parameters[0].LOCAL().GetText()] = true
	value := s.expression(lambda.Expression())
	s.locals = previous
	return spl2TransformExpressionEvidence(value, func(value spl2ExpressionEvidence) spl2ExpressionEvidence {
		value.nonnull = value.modeled
		value.requirementNonnull = value.modeled
		value.domain = "boolean"
		return value
	})
}
