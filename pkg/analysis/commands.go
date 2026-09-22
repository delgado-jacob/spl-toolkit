package analysis

import (
	"github.com/antlr4-go/antlr/v4"
	"github.com/delgado-jacob/spl-toolkit/parser"
	"strconv"
	"strings"
)

type commandHandler func(*semanticStage, antlr.ParserRuleContext)
type commandSpec struct {
	handle     commandHandler
	limitation string
}

var commands = map[string]commandSpec{
	"search":      {searchCommand, "Search predicates and typed selectors. Damaged searches retain sound prefixes; unclosed delimiters can withhold later findings."},
	"where":       {whereCommand, "Expression predicates; dynamic functions are incomplete."},
	"eval":        {evalCommand, "Sequential exact-name assignments; registered pure functions only. Syntax recovery retains the sound assignment prefix and stops at the first damaged assignment."},
	"rename":      {renameCommand, "Snapshot source bindings; conflicting destinations are incomplete."},
	"fields":      {fieldsCommand, "Exact exclusion and closed-input inclusion; known internal fields are retained. Open-input inclusion has unresolved retained-internal membership; wildcards require proven membership."},
	"table":       {fieldsCommand, "Projection; quoted and unquoted wildcard selectors require proven membership."},
	"stats":       {statsCommand, "Registered aggregates and exact grouping fields; options and wildcard grouping are unmodeled."},
	"eventstats":  {statsCommand, "Additive registered aggregates; no options."},
	"streamstats": {statsCommand, "Additive registered aggregates; window/options are unmodeled."},
	"lookup":      {lookupCommand, "Explicit exact inputs and outputs; OUTPUTNEW preserves known fields; options and wildcard columns are unmodeled."},
	"inputlookup": {inputlookupCommand, "Open source from catalog; options are unmodeled."},
	"sort":        {sortCommand, "Numeric limit and signed exact field selectors; wildcard selectors are unmodeled."},
	"dedup":       {dedupCommand, "Numeric limit and exact field lists; options and wildcard selectors are unmodeled."},
	"head":        {limitCommand, "Optional numeric limit only."},
	"tail":        {limitCommand, "Optional numeric limit only."},
	"append":      {branchCommand, "Branch merging is unmodeled."},
	"appendpipe":  {branchCommand, "Branch merging is unmodeled."},
	"join":        {branchCommand, "Branch merging is unmodeled."},
	"datamodel":   {nil, "Exact model and optional dataset operands only; field effects are unmodeled. Qualified dataset references can cover the dataset component."},
	"from":        {nil, "One exact dataset operand; datamodel:model.dataset yields overlapping located root-model and dataset references. Field effects are unmodeled."},
	"tstats":      {tstatsCommand, "Exact data-model sources, predicates, registered aggregates, aliases, and exact grouping fields with literal _time spans. Dynamic catalogs, macros, wildcard grouping, prestats/append result-shape modes, unknown options, and unknown output identities remain held."},
	"fillnull":    {fillnullCommand, "Optional literal replacement and exact target fields. All-fields mode preserves field shape; dynamic selectors and unsupported options remain held."},
	"rex":         {rexCommand, "Exact or default _raw input, literal max_match, exact offset field, and unambiguous named captures. Sed mode and ambiguous captures remain held."},
	"spath":       {spathCommand, "Exact or default _raw input with an exact literal path and explicit exact output. Auto-extraction and dynamic operands remain held."},
	"bin":         {binCommand, "One exact input, supported literal value options, and an optional exact alias. Dynamic operands and unsupported options remain held."},
	"bucket":      {binCommand, "Alias of bin with one exact input, supported literal value options, and an optional exact alias. Dynamic operands and unsupported options remain held."},
	"regex":       {regexCommand, "Quoted expression against _raw or an exact field with = or !=. Dynamic fields and malformed forms remain held."},
	"mvexpand":    {mvexpandCommand, "One exact field with identity-preserving field flow. Options and dynamic fields remain held."},
	"macro":       {nil, "Synthetic category for a macro-only stage; exact macro name dependencies, unresolved expansion. A literal command named macro remains unmodeled."},
}

func searchCommand(s *semanticStage, node antlr.ParserRuleContext) {
	s.searchPredicate(node, true)
}

func (s *semanticStage) searchPredicate(node antlr.Tree, rewrite bool) {
	if rewrite {
		s.rewritePredicate(node, "spl", true, false)
	}
	var visit func(antlr.Tree)
	visit = func(n antlr.Tree) {
		switch c := n.(type) {
		case parser.IAnalysisSearchTermContext:
			if c.AnalysisIdentifier() != nil {
				if !s.sound(c) {
					return
				}
				name := normalizedName(c.AnalysisIdentifier().GetText())
				kind := strings.ToLower(name)
				values := c.AllAnalysisSearchValue()
				if kind == "index" || kind == "source" || kind == "sourcetype" {
					for _, v := range values {
						s.dependency(v, normalizedName(v.GetText()), kind)
					}
				} else {
					s.read(c.AnalysisIdentifier(), name, "filter")
				}
				return
			}
		case parser.IAnalysisMacroContext:
			s.macro(c)
			return
		case parser.IAnalysisSubqueryContext:
			s.diagnostic(CodeUnsupportedSemantics, "subsearch result field effects are unmodeled", c)
			return
		}
		for i := 0; i < n.GetChildCount(); i++ {
			visit(n.GetChild(i))
		}
	}
	visit(node)
}
func whereCommand(s *semanticStage, node antlr.ParserRuleContext) {
	s.rewritePredicate(node, "spl", false, false)
	s.expression(node.(*parser.AnalysisWhereStageContext).AnalysisExpression())
}
func evalCommand(s *semanticStage, node antlr.ParserRuleContext) {
	for _, a := range node.(*parser.AnalysisEvalStageContext).AllAnalysisAssignment() {
		if !intact(a) || a.AnalysisIdentifier() == nil || a.AnalysisExpression() == nil {
			continue
		}
		inputs := s.expression(a.AnalysisExpression())
		s.applyAssignment(s.operand(a.AnalysisIdentifier(), normalizedName(a.AnalysisIdentifier().GetText())), inputs, !s.result.Stages[s.stage].SemanticComplete, false)
	}
}
func wildcardMatches(pattern, name string) bool { // Selectors admit only '*' wildcard syntax.
	p, n := []rune(pattern), []rune(name)
	pi, ni, star, mark := 0, 0, -1, 0
	for ni < len(n) {
		if pi < len(p) && p[pi] != '*' && p[pi] == n[ni] {
			pi++
			ni++
		} else if pi < len(p) && p[pi] == '*' {
			star = pi
			pi++
			mark = ni
		} else if star >= 0 {
			pi = star + 1
			mark++
			ni = mark
		} else {
			return false
		}
	}
	for pi < len(p) && p[pi] == '*' {
		pi++
	}
	return pi == len(p)
}

// Selector operands carry pattern semantics independently of identifier quoting.
func selectorPattern(c parser.IAnalysisSelectorContext) bool {
	return len(c.AllMULT()) > 0 || (c.AnalysisIdentifier() != nil && strings.Contains(normalizedName(c.AnalysisIdentifier().GetText()), "*"))
}
func selectorName(c parser.IAnalysisSelectorContext) string {
	var name strings.Builder
	for _, child := range c.GetChildren() {
		switch part := child.(type) {
		case parser.IAnalysisIdentifierContext:
			name.WriteString(normalizedName(part.GetText()))
		case antlr.TerminalNode:
			if part.GetSymbol().GetTokenType() == parser.SPLParserMULT {
				name.WriteByte('*')
			}
		}
	}
	return name.String()
}
func (s *semanticStage) selector(c parser.IAnalysisSelectorContext, role string) ([]string, []string) {
	command := s.result.Stages[s.stage].Command
	return s.selectorAt(s.operand(c, selectorName(c)), role, command == "fields" || command == "table" || command == "rename")
}
func renameCommand(s *semanticStage, node antlr.ParserRuleContext) {
	pairs := []renameOperands{}
	for _, r := range node.(*parser.AnalysisRenameStageContext).AllAnalysisRename() {
		if !intact(r) {
			continue
		}
		source := s.operand(r.AnalysisSelector(), normalizedName(r.AnalysisSelector().GetText()))
		if selectorPattern(r.AnalysisSelector()) {
			source.Name = selectorName(r.AnalysisSelector())
		}
		target := s.operand(r.AnalysisAlias().AnalysisIdentifier(), normalizedName(r.AnalysisAlias().AnalysisIdentifier().GetText()))
		if strings.Contains(target.Name, "*") {
			target.Resolution = "wildcard"
		}
		pairs = append(pairs, renameOperands{Source: source, Target: target})
	}
	s.applyRename(pairs)
}
func fieldsCommand(s *semanticStage, node antlr.ParserRuleContext) {
	ctx := node.(*parser.AnalysisFieldsStageContext)
	mode := "include"
	if s.result.Stages[s.stage].Command == "table" {
		if ctx.SUB() != nil || ctx.ADD() != nil {
			s.diagnostic(CodeUnsupportedSemantics, "table does not support signed projection", ctx)
			return
		}
		mode = "table"
	} else if ctx.SUB() != nil {
		mode = "exclude"
	}
	selectors := []locatedOperand{}
	for _, c := range ctx.AnalysisFieldList().AllAnalysisSelector() {
		selectors = append(selectors, s.operand(c, selectorName(c)))
	}
	s.applyProjection(selectors, mode, mode == "include")
}

func statsCommand(s *semanticStage, node antlr.ParserRuleContext) {
	ctx := node.(*parser.AnalysisStatsStageContext)
	if len(ctx.AllAnalysisOption()) > 0 {
		s.diagnostic(CodeUnsupportedSemantics, "aggregate command options are unmodeled", ctx)
	}
	outputs := []aggregateOutput{}
	for _, a := range ctx.AllAnalysisAggregate() {
		if !intact(a) {
			continue
		}
		name := ""
		ids := []string{}
		var outCtx antlr.ParserRuleContext = a
		if call := a.AnalysisFunctionCall(); call != nil {
			valid := s.function(call, true)
			if call.AnalysisArgumentList() != nil {
				ids = s.expression(call.AnalysisArgumentList())
			}
			if valid {
				functionName := strings.ToLower(call.AnalysisFunctionName().GetText())
				name = functionName + "()"
				if call.AnalysisArgumentList() != nil {
					field, exact := aggregateField(call.AnalysisArgumentList().AnalysisExpression(0))
					if exact {
						name = functionName + "(" + field + ")"
					} else {
						name = ""
						s.diagnostic(CodeUnsupportedSemantics, "aggregate arguments must be exact field identifiers", call)
					}
				} else if functionName == "count" {
					name = "count"
				}
			}
			outCtx = call
		} else if strings.EqualFold(a.AnalysisIdentifier().GetText(), "count") {
			name = "count"
			outCtx = a.AnalysisIdentifier()
		} else {
			s.diagnostic(CodeUnsupportedSemantics, "bare aggregate must be count", a)
		}
		if a.AnalysisAlias() != nil {
			name = normalizedName(a.AnalysisAlias().AnalysisIdentifier().GetText())
			outCtx = a.AnalysisAlias().AnalysisIdentifier()
		}
		if name != "" {
			outputs = append(outputs, aggregateOutput{Target: s.operand(outCtx, name), InputReferenceIDs: ids})
		}
	}
	groups := []locatedOperand{}
	if group := ctx.AnalysisGroup(); group != nil {
		for _, c := range group.AnalysisFieldList().AllAnalysisSelector() {
			groups = append(groups, s.operand(c, selectorName(c)))
		}
	}
	s.applyAggregation(outputs, groups, s.result.Stages[s.stage].Command != "stats")
}

func lookupCommand(s *semanticStage, node antlr.ParserRuleContext) {
	c := node.(*parser.AnalysisLookupStageContext).AnalysisLookup()
	s.dependency(c.AnalysisCatalogName(), normalizedName(c.AnalysisCatalogName().GetText()), "lookup")
	if len(c.AllAnalysisOption()) > 0 {
		s.diagnostic(CodeUnsupportedSemantics, "lookup options are unmodeled", c)
	}
	ids := []string{}
	for _, in := range c.AllAnalysisLookupInput() {
		local := in.AnalysisIdentifier()
		if in.AnalysisAlias() != nil {
			local = in.AnalysisAlias().AnalysisIdentifier()
		}
		if strings.Contains(normalizedName(in.AnalysisIdentifier().GetText()), "*") || strings.Contains(normalizedName(local.GetText()), "*") {
			s.diagnostic(CodeUnsupportedSemantics, "lookup wildcard input columns are unmodeled", in)
			continue
		}
		ids = append(ids, s.read(local, normalizedName(local.GetText()), "read"))
	}
	if len(c.AllAnalysisOutput()) == 0 {
		s.diagnostic(CodeUnsupportedSemantics, "lookup without explicit outputs has unknown output fields", c)
	}
	for _, out := range c.AllAnalysisOutput() {
		for _, item := range out.AllAnalysisLookupInput() {
			local := item.AnalysisIdentifier()
			if item.AnalysisAlias() != nil {
				local = item.AnalysisAlias().AnalysisIdentifier()
			}
			name := normalizedName(local.GetText())
			if strings.Contains(normalizedName(item.AnalysisIdentifier().GetText()), "*") || strings.Contains(name, "*") {
				s.diagnostic(CodeUnsupportedSemantics, "lookup wildcard output columns are unmodeled", item)
				continue
			}
			// Preserve source-order diagnostics while every output shares the same
			// once-prepared local match references, including overlapping destinations.
			s.applyLookupOutputs(ids, []lookupOutput{{Target: s.operand(local, name), PreserveExisting: out.OUTPUTNEW() != nil}})
		}
	}
}
func inputlookupCommand(s *semanticStage, node antlr.ParserRuleContext) {
	c := node.(*parser.AnalysisInputlookupStageContext)
	s.applySource()
	s.dependency(c.AnalysisCatalogName(), normalizedName(c.AnalysisCatalogName().GetText()), "lookup")
	if len(c.AllAnalysisOption()) > 0 {
		s.diagnostic(CodeUnsupportedSemantics, "inputlookup options are unmodeled", c)
	}
}
func sortCommand(s *semanticStage, node antlr.ParserRuleContext) {
	s.numericLimit(node.(*parser.AnalysisSortStageContext).AnalysisLimit(), true)
	for _, c := range node.(*parser.AnalysisSortStageContext).AllAnalysisSortField() {
		s.selector(c.AnalysisSelector(), "sort")
	}
}
func dedupCommand(s *semanticStage, node antlr.ParserRuleContext) {
	c := node.(*parser.AnalysisDedupStageContext)
	s.numericLimit(c.AnalysisLimit(), false)
	if len(c.AllAnalysisOption()) > 0 {
		s.diagnostic(CodeUnsupportedSemantics, "dedup options are unmodeled", c)
	}
	for _, f := range c.AnalysisFieldList().AllAnalysisSelector() {
		s.selector(f, "read")
	}
}
func limitCommand(s *semanticStage, node antlr.ParserRuleContext) {
	c := node.(*parser.AnalysisLimitStageContext)
	s.numericLimit(c.AnalysisLimit(), true)
	if c.AnalysisExpression() != nil {
		s.expression(c.AnalysisExpression())
		s.diagnostic(CodeUnsupportedSemantics, "only numeric head/tail limits are modeled", c)
	}
}

func branchCommand(s *semanticStage, node antlr.ParserRuleContext) {
	if stage, ok := node.(*parser.AnalysisJoinStageContext); ok && stage.AnalysisJoin() != nil {
		for _, key := range stage.AnalysisJoin().AllAnalysisIdentifier() {
			s.read(key, normalizedName(key.GetText()), "read")
		}
	}
	command := s.result.Stages[s.stage].Command
	s.diagnostic(CodeUnsupportedSemantics, "command \""+command+"\" has unmodeled field effects", node)
}

// Only a single typed identifier establishes an implicit aggregate output name.
func aggregateField(node antlr.Tree) (string, bool) {
	if id, ok := node.(parser.IAnalysisIdentifierContext); ok {
		return normalizedName(id.GetText()), intact(id)
	}
	if node.GetChildCount() != 1 {
		return "", false
	}
	return aggregateField(node.GetChild(0))
}
func (s *semanticStage) numericLimit(c parser.IAnalysisLimitContext, zero bool) {
	if c == nil {
		return
	}
	n, err := strconv.ParseUint(c.GetText(), 10, 64)
	if err != nil || (!zero && n == 0) {
		s.diagnostic(CodeUnsupportedSemantics, "command limit must be a supported integer", c)
	}
}
