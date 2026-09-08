package analysis

import (
	"fmt"
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
	out := spl2ExpressionEvidence{ids: []string{}}
	name := c.Identifier().GetText()
	values := []spl2ExpressionEvidence{}
	named := false
	if args := c.Arguments(); args != nil {
		for _, expr := range args.AllExpression() {
			var v spl2ExpressionEvidence
			access := spl2SQLFieldAccess(expr)
			if !aggregate && (name == "isnull" || name == "isnotnull") && spl2IntactSyntax(c) && len(args.AllExpression()) == 1 && len(args.AllNamedArgument()) == 0 && access != nil && len(access.AllAccessPart()) == 0 && access.Primary().FieldName() != nil && access.Primary().FieldName().Identifier() != nil {
				id := s.readIdentifier(access.Primary().FieldName().Identifier(), "null_test")
				v = spl2ExpressionEvidence{ids: []string{id}, modeled: true}
			} else if aggregate && access != nil && len(access.AllAccessPart()) == 0 && access.Primary().FieldName() != nil && access.Primary().FieldName().Identifier() != nil && strings.Contains(s.operand(access.Primary().FieldName().Identifier()).Name, "*") {
				_, ids := s.selectorAt(s.selector(access.Primary().FieldName().Identifier()), "read", false)
				v = spl2ExpressionEvidence{ids: ids}
			} else {
				v = s.expression(expr)
			}
			values = append(values, v)
			out.ids = append(out.ids, v.ids...)
		}
		for _, arg := range args.AllNamedArgument() {
			named = true
			v := s.expression(arg.Expression())
			out.ids = append(out.ids, v.ids...)
		}
	}
	if name == "batch_id" || name == "batch_time" {
		s.result.Coverage.SyntaxComplete = false
		s.diagnosticAt("SPL_PROFILE_MISMATCH", "error", "compatibility", fmt.Sprintf("function %q is unavailable in the splunkd profile", name), s.parsed2.source.contextLocation(c), true)
		return out
	}
	spec, known := spl2Functions[name]
	if !known {
		s.diagnosticAt(CodeUnsupportedFunction, "warning", "unsupported_semantics", fmt.Sprintf("function %q has unmodeled semantics", name), s.parsed2.source.contextLocation(c), true)
		return out
	}
	if named {
		s.unsupported(c, "Named function argument effects are not yet modeled")
		return out
	}
	if !aggregate && (name == "min" || name == "max") {
		s.unsupported(c, "Scalar min/max signatures are outside the approved positional core")
		return out
	}
	n := len(values)
	if n < spec.min || spec.max >= 0 && n > spec.max || name == "case" && n%2 != 0 || spec.aggregate != aggregate {
		s.diagnosticAt(CodeSyntaxError, "error", "contract", fmt.Sprintf("function %q has invalid positional arity or context", name), s.parsed2.source.contextLocation(c), true)
		return out
	}
	out.modeled = true
	for _, value := range values {
		out.modeled = out.modeled && value.modeled
	}
	switch name {
	case "count", "dc", "distinct_count", "isnull", "isnotnull":
		out.nonnull = out.modeled
	case "coalesce":
		out.exactNull = out.modeled
		for _, value := range values {
			out.nonnull = out.nonnull || value.nonnull
			out.exactNull = out.exactNull && value.exactNull
		}
	case "if":
		out.nonnull = values[1].nonnull && values[2].nonnull && out.modeled
		out.exactNull = values[1].exactNull && values[2].exactNull && out.modeled
	case "case":
		out.nonnull = out.modeled
		out.exactNull = out.modeled
		fallback := false
		for i := 0; i < n; i += 2 {
			out.nonnull = out.nonnull && values[i+1].nonnull
			out.exactNull = out.exactNull && values[i+1].exactNull
			if values[i].truth {
				fallback = true
				break
			}
		}
		out.nonnull = out.nonnull && fallback
	case "tonumber", "tostring", "mvcount": // Neither argument presence nor function spelling proves a value.
	case "abs", "ceil", "ceiling", "floor", "round":
		out.nonnull = out.modeled
		out.domain = "number"
		for _, value := range values {
			out.nonnull = out.nonnull && value.nonnull && value.domain == "number"
		}
	case "lower", "upper", "len", "trim", "ltrim", "rtrim", "split":
		out.nonnull = out.modeled
		out.domain = "string"
		for _, value := range values {
			out.nonnull = out.nonnull && value.nonnull && value.domain == "string"
		}
	case "substr":
		out.nonnull = out.modeled && values[0].nonnull && values[0].domain == "string"
		out.domain = "string"
		for _, value := range values[1:] {
			out.nonnull = out.nonnull && value.nonnull && value.domain == "number"
		}
		// Aggregate operand presence does not establish empty-group/null results.
		// Regex value semantics likewise require more than argument count.
	}
	out.nonnull = out.nonnull && out.modeled
	return out
}
