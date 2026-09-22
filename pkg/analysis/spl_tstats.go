package analysis

import (
	"strconv"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	"github.com/delgado-jacob/spl-toolkit/parser"
)

func tstatsCommand(s *semanticStage, node antlr.ParserRuleContext) {
	stage, ok := node.(*parser.AnalysisTstatsStageContext)
	if !ok || stage.AnalysisTstats() == nil {
		return
	}
	ctx := stage.AnalysisTstats()
	ordinaryOutputs := true
	for _, option := range ctx.AllAnalysisTstatsOption() {
		if s.tstatsOption(option) == "prestats" {
			ordinaryOutputs = false
		}
	}
	if macro := ctx.AnalysisMacro(); macro != nil {
		s.macro(macro)
	}

	// tstats is a generating command. Options and an inline macro are diagnosed
	// before the source reset so their uncertainty cannot turn exact source reads
	// into invented upstream bindings.
	s.applySource()
	outputs := []aggregateOutput{}
	for _, aggregate := range ctx.AllAnalysisAggregate() {
		output, understood := s.tstatsAggregate(aggregate)
		if output != nil {
			outputs = append(outputs, *output)
		}
		if !understood {
			// The failed aggregate's reads are already durable facts. Reset only the
			// source-field cache so later exact siblings remain direct requirements.
			s.applySource()
		}
	}
	if from := ctx.AnalysisTstatsFrom(); from != nil && from.AnalysisDataModelName() != nil {
		s.qualifiedCatalog(from.AnalysisDataModelName(), false)
		// FROM establishes the generating source even when an earlier boundary has
		// already made the stage incomplete or this catalog identity is dynamic.
		s.applySource()
	}
	if where := ctx.AnalysisTstatsWhere(); where != nil && where.AnalysisSearch() != nil {
		complete := s.result.Stages[s.stage].SemanticComplete
		s.searchPredicate(where.AnalysisSearch(), false)
		if complete && !s.result.Stages[s.stage].SemanticComplete {
			s.applySource()
		}
	}

	groups := []locatedOperand{}
	if group := ctx.AnalysisTstatsGroup(); group != nil {
		for _, item := range group.AllAnalysisTstatsGroupItem() {
			if item.AnalysisIdentifier() == nil {
				continue
			}
			identifier := item.AnalysisIdentifier()
			name := normalizedName(identifier.GetText())
			operand := s.operand(identifier, name)
			if strings.Contains(name, "*") {
				operand.Resolution = "wildcard"
				s.diagnostic(CodeUnsupportedSemantics, "tstats wildcard grouping fields are unmodeled", identifier)
				continue
			}
			groups = append(groups, operand)
			if span := item.AnalysisTstatsSpanOption(); span != nil {
				if !s.tstatsSpanModeled(name, span.AnalysisUnitOptionValue()) {
					s.diagnostic(CodeUnsupportedSemantics, "tstats span requires a supported positive numeric _time duration", span)
				}
			}
		}
	}

	if s.result.Stages[s.stage].SemanticComplete {
		s.applyAggregation(outputs, groups, false)
		return
	}
	if !ordinaryOutputs {
		outputs = nil
	}
	// A held form has an unknown remainder. Start from a clean source cache so
	// exact groups remain required, then retain only independently known output
	// bindings in an open, uncertain result environment.
	s.applySource()
	s.applyTstatsPartialAggregation(outputs, groups)
}

// tstatsOption returns "prestats" only for the held true mode that suppresses
// ordinary aggregate output ownership. Every other held option retains exact
// sibling outputs conditionally.
func (s *semanticStage) tstatsOption(ctx parser.IAnalysisTstatsOptionContext) string {
	if ctx == nil || ctx.AnalysisIdentifier() == nil {
		return ""
	}
	name := strings.ToLower(normalizedName(ctx.AnalysisIdentifier().GetText()))
	value := ctx.AnalysisOptionValue()
	boolean := func() (bool, bool) {
		if value == nil || value.AnalysisIdentifier() == nil {
			return false, false
		}
		raw := value.AnalysisIdentifier().GetText()
		switch {
		case strings.EqualFold(raw, "true"):
			return true, true
		case strings.EqualFold(raw, "false"):
			return false, true
		default:
			return false, false
		}
	}
	unsupported := func(message string) {
		s.diagnostic(CodeUnsupportedSemantics, message, ctx)
	}
	switch name {
	case "summariesonly", "local", "include_reduced_buckets", "allow_old_summaries":
		if _, exact := boolean(); !exact {
			unsupported("tstats option requires an exact boolean literal")
		}
	case "chunk_size":
		literal := parser.IAnalysisLiteralContext(nil)
		if value != nil {
			literal = value.AnalysisLiteral()
		}
		if literal == nil || literal.NUMBER() == nil {
			unsupported("tstats chunk_size requires a non-negative integer literal")
			break
		}
		if _, err := strconv.ParseUint(literal.NUMBER().GetText(), 10, 64); err != nil {
			unsupported("tstats chunk_size requires a non-negative integer literal")
		}
	case "fillnull_value":
		if value == nil || value.AnalysisLiteral() == nil {
			unsupported("tstats fillnull_value requires an exact literal")
		}
	case "prestats", "append":
		enabled, exact := boolean()
		if !exact {
			unsupported("tstats result-shape option requires an exact boolean literal")
			break
		}
		if enabled {
			unsupported("tstats result-shape option is unmodeled when true")
			if name == "prestats" {
				return "prestats"
			}
		}
	default:
		unsupported("tstats option is unmodeled")
	}
	return ""
}

func (s *semanticStage) tstatsSpanModeled(group string, value parser.IAnalysisUnitOptionValueContext) bool {
	if group != "_time" || value == nil || value.NUMBER() == nil || !s.sound(value) {
		return false
	}
	magnitude := value.NUMBER().GetText()
	if !positiveNumber(magnitude) {
		return false
	}
	unit := value.AnalysisUnitSuffix()
	if unit == nil {
		return true
	}
	if strings.Contains(magnitude, ".") {
		return false
	}
	switch strings.ToLower(unit.GetText()) {
	case "s", "sec", "secs", "second", "seconds",
		"m", "min", "mins", "minute", "minutes",
		"h", "hr", "hrs", "hour", "hours",
		"d", "day", "days",
		"mon", "month", "months":
		return true
	default:
		return false
	}
}

func (s *semanticStage) tstatsAggregate(aggregate parser.IAnalysisAggregateContext) (*aggregateOutput, bool) {
	if aggregate == nil || !intact(aggregate) {
		return nil, false
	}
	name := ""
	inputs := []string{}
	understood := true
	var target antlr.ParserRuleContext = aggregate
	if call := aggregate.AnalysisFunctionCall(); call != nil {
		functionName := strings.ToLower(call.AnalysisFunctionName().GetText())
		arguments := call.AnalysisArgumentList()
		dynamicInput := false
		if arguments != nil {
			inputs, dynamicInput = s.tstatsAggregateInputs(arguments)
			if dynamicInput {
				understood = false
			}
		}
		if functionName == "prefix" {
			s.diagnostic(CodeUnsupportedSemantics, "tstats PREFIX output identity is unmodeled", call)
			understood = false
		} else {
			if s.function(call, true) && understood {
				name = functionName + "()"
				if arguments != nil {
					field, exact := aggregateField(arguments.AnalysisExpression(0))
					if exact {
						name = functionName + "(" + field + ")"
					} else {
						name = ""
						understood = false
						s.diagnostic(CodeUnsupportedSemantics, "aggregate arguments must be exact field identifiers", call)
					}
				} else if functionName == "count" {
					name = "count"
				}
			} else {
				understood = false
			}
		}
		target = call
	} else if aggregate.AnalysisIdentifier() != nil && strings.EqualFold(aggregate.AnalysisIdentifier().GetText(), "count") {
		name = "count"
		target = aggregate.AnalysisIdentifier()
	} else {
		understood = false
		s.diagnostic(CodeUnsupportedSemantics, "bare aggregate must be count", aggregate)
	}
	if alias := aggregate.AnalysisAlias(); alias != nil && alias.AnalysisIdentifier() != nil {
		aliasName := normalizedName(alias.AnalysisIdentifier().GetText())
		target = alias.AnalysisIdentifier()
		if strings.Contains(aliasName, "*") {
			name = ""
			understood = false
			s.diagnostic(CodeUnsupportedSemantics, "tstats aggregate output must be an exact identifier", alias.AnalysisIdentifier())
		} else if understood && name != "" {
			name = aliasName
		}
	}
	if !understood || name == "" {
		return nil, false
	}
	return &aggregateOutput{Target: s.operand(target, name), InputReferenceIDs: inputs}, true
}

func (s *semanticStage) tstatsAggregateInputs(node antlr.Tree) ([]string, bool) {
	type heldInput struct {
		context parser.IAnalysisIdentifierContext
		operand locatedOperand
		id      string
	}
	inputs := []string{}
	held := []heldInput{}
	var visit func(antlr.Tree)
	visit = func(node antlr.Tree) {
		if node == nil {
			return
		}
		switch ctx := node.(type) {
		case parser.IAnalysisIdentifierContext:
			if !intact(ctx) {
				return
			}
			operand := fieldCommandOperand(s, ctx)
			if operand.Resolution == "exact" {
				role := "read"
				if s.splOwner(ctx).role == "null_test" {
					role = "null_test"
				}
				if id := s.readAt(operand, role); id != "" {
					inputs = append(inputs, id)
				}
				return
			}
			id := fieldCommandHeldRead(s, operand, "read")
			if id != "" {
				inputs = append(inputs, id)
			}
			held = append(held, heldInput{context: ctx, operand: operand, id: id})
			return
		case parser.IAnalysisLiteralContext:
			return
		case parser.IAnalysisFunctionCallContext:
			s.function(ctx, false)
			if ctx.AnalysisArgumentList() != nil {
				visit(ctx.AnalysisArgumentList())
			}
			return
		case parser.IAnalysisMacroContext:
			s.macro(ctx)
			return
		case parser.IAnalysisSubqueryContext:
			s.diagnostic(CodeUnsupportedSemantics, "subsearch result field effects are unmodeled", ctx)
			return
		case antlr.TerminalNode:
			if atom, ok := node.GetParent().(parser.IAnalysisAtomContext); ok && ctx.GetSymbol().GetTokenType() == parser.SPLLexerMULT {
				s.diagnostic(CodeUnresolvedWildcard, "wildcard expression membership is unresolved", atom)
			}
			return
		}
		for i := 0; i < node.GetChildCount(); i++ {
			visit(node.GetChild(i))
		}
	}
	visit(node)
	for _, input := range held {
		if input.id != "" {
			s.diagnosticAtOwned(CodeDynamicReference, "warning", "unsupported_semantics", "tstats aggregate operand identity must be exact", input.operand.Location, true, []string{input.id})
			continue
		}
		s.diagnostic(CodeDynamicReference, "tstats aggregate operand identity must be exact", input.context)
	}
	return inputs, len(held) > 0
}

func (s *semanticStage) applyTstatsPartialAggregation(outputs []aggregateOutput, groups []locatedOperand) {
	type retainedGroup struct {
		operand            locatedOperand
		inputReferenceID   string
		origins            []string
		requirementOrigins []string
	}
	retained := []retainedGroup{}
	for _, group := range groups {
		id := s.readAt(group, "group")
		if id == "" {
			continue
		}
		item := retainedGroup{operand: group, inputReferenceID: id, origins: s.origins([]string{id})}
		if trace := s.env.requirements.trace; trace != nil {
			item.requirementOrigins = traceOrigins(trace, []string{id})
		}
		retained = append(retained, item)
	}

	trace := s.env.requirements.trace
	output := newEnvironmentWithRequirementTrace(trace)
	output.rewrite = s.env.rewrite.clone()
	output.open = true
	output.uncertain = true
	output.requirements.open = true
	output.requirements.uncertain = true
	s.env = output
	for _, group := range retained {
		s.env.install(group.operand.Name, group.origins, true)
		s.env.requirements.install(group.operand.Name, group.requirementOrigins, true)
		s.transitions = append(s.transitions, Transition{Operation: "project", Output: group.operand.Name, InputReferenceIDs: []string{group.inputReferenceID}, Conditional: true})
	}
	for _, output := range outputs {
		s.createAtWithRequirementConditional(output.Target, "output", "aggregate", output.InputReferenceIDs, true, true)
	}
}
