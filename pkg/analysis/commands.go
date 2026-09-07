package analysis

import (
	"fmt"
	"github.com/antlr4-go/antlr/v4"
	"github.com/delgado-jacob/spl-toolkit/parser"
	"sort"
	"strconv"
	"strings"
)

type commandHandler func(*semanticStage, antlr.ParserRuleContext)
type commandSpec struct {
	handle     commandHandler
	limitation string
}

var commands = map[string]commandSpec{
	"search":      {searchCommand, "Search predicates and typed selectors."},
	"where":       {whereCommand, "Expression predicates; dynamic functions are incomplete."},
	"eval":        {evalCommand, "Sequential exact-name assignments; registered pure functions only."},
	"rename":      {renameCommand, "Snapshot source bindings; conflicting destinations are incomplete."},
	"fields":      {fieldsCommand, "Exact exclusion and closed-input inclusion; known internal fields are retained. Open-input inclusion has unresolved retained-internal membership; wildcards require proven membership."},
	"table":       {fieldsCommand, "Projection; wildcards require proven membership."},
	"stats":       {statsCommand, "Registered aggregates and exact grouping fields; no options."},
	"eventstats":  {statsCommand, "Additive registered aggregates; no options."},
	"streamstats": {statsCommand, "Additive registered aggregates; window/options are unmodeled."},
	"lookup":      {lookupCommand, "Explicit inputs and outputs; OUTPUTNEW preserves known fields; options are unmodeled."},
	"inputlookup": {inputlookupCommand, "Open source from catalog; options are unmodeled."},
	"sort":        {sortCommand, "Numeric limit and signed field selectors."},
	"dedup":       {dedupCommand, "Numeric limit and exact field lists; options are unmodeled."},
	"head":        {limitCommand, "Optional numeric limit only."},
	"tail":        {limitCommand, "Optional numeric limit only."},
	"append":      {nil, "Branch merging is unmodeled."},
	"appendpipe":  {nil, "Branch merging is unmodeled."},
	"join":        {nil, "Branch merging is unmodeled."},
	"datamodel":   {nil, "Field effects are unmodeled."},
	"from":        {nil, "Field effects are unmodeled."},
	"tstats":      {nil, "Field effects are unmodeled."},
}

func searchCommand(s *semanticStage, node antlr.ParserRuleContext) {
	var visit func(antlr.Tree)
	visit = func(n antlr.Tree) {
		switch c := n.(type) {
		case parser.IAnalysisSearchTermContext:
			if c.AnalysisIdentifier() != nil {
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
			s.diagnostic(CodeDynamicReference, "macro expansion is unresolved", c)
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
	s.expression(node.(*parser.AnalysisWhereStageContext).AnalysisExpression())
}
func evalCommand(s *semanticStage, node antlr.ParserRuleContext) {
	for _, a := range node.(*parser.AnalysisEvalStageContext).AllAnalysisAssignment() {
		if !intact(a) || a.AnalysisIdentifier() == nil || a.AnalysisExpression() == nil {
			continue
		}
		inputs := s.expression(a.AnalysisExpression())
		s.create(a.AnalysisIdentifier(), normalizedName(a.AnalysisIdentifier().GetText()), "create", "create", inputs, !s.result.Stages[s.stage].SemanticComplete)
	}
}
func wildcardMatches(pattern, name string) bool { // Selectors admit only '*' wildcard syntax.
	p, n := []rune(pattern), []rune(name)
	pi, ni, star, mark := 0, 0, -1, 0
	for ni < len(n) {
		if pi < len(p) && p[pi] == n[ni] {
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
func (s *semanticStage) selector(c parser.IAnalysisSelectorContext, role string) ([]string, []string) {
	name := normalizedName(c.GetText())
	if len(c.AllMULT()) == 0 {
		return []string{name}, []string{s.read(c, name, role)}
	}
	id := s.reference(c, name, "field", role)
	if id == "" {
		return []string{}, []string{}
	}
	ref := &s.result.References[len(s.result.References)-1]
	ref.Binding = "indeterminate"
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
		s.diagnostic(CodeUnresolvedWildcard, fmt.Sprintf("wildcard %q membership is unresolved", name), c)
	}
	return names, []string{id}
}
func renameCommand(s *semanticStage, node antlr.ParserRuleContext) {
	ctx := node.(*parser.AnalysisRenameStageContext)
	original := s.env
	before := s.env.clone()
	type rename struct {
		source, dest string
		ctx          parser.IAnalysisAliasContext
		input        string
		conditional  bool
	}
	items := []rename{}
	sources, dests := map[string]bool{}, map[string]bool{}
	conflict := false
	for _, r := range ctx.AllAnalysisRename() {
		if !intact(r) {
			continue
		}
		src := normalizedName(r.AnalysisSelector().GetText())
		dst := normalizedName(r.AnalysisAlias().AnalysisIdentifier().GetText())
		if len(r.AnalysisSelector().AllMULT()) > 0 {
			s.env = before
			names, _ := s.selector(r.AnalysisSelector(), "read")
			for _, name := range names {
				sources[name] = true
			}
			dests[dst] = true
			s.diagnostic(CodeUnsupportedSemantics, "wildcard rename substitution is unmodeled", r)
			conflict = true
			continue
		}
		s.env = before
		id := s.read(r.AnalysisSelector(), src, "read")
		_, existingDestination := original.fields[dst]
		if sources[src] || dests[dst] || existingDestination {
			conflict = true
		}
		sources[src] = true
		dests[dst] = true
		binding := s.result.References[len(s.result.References)-1].Binding
		items = append(items, rename{src, dst, r.AnalysisAlias(), id, binding == "indeterminate" || binding == "unavailable"})
	}
	s.env = before // All reads observed the snapshot; no destination was installed while reading.
	for name := range sources {
		if dests[name] {
			conflict = true
		}
	}
	if conflict {
		s.diagnostic(CodeUnsupportedSemantics, "rename has conflicting or unsupported source/destination mappings", ctx)
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
		s.create(r.ctx.AnalysisIdentifier(), r.dest, "rename", "rename", []string{r.input}, r.conditional)
	}
}
func fieldsCommand(s *semanticStage, node antlr.ParserRuleContext) {
	ctx := node.(*parser.AnalysisFieldsStageContext)
	exclude := ctx.SUB() != nil
	if s.result.Stages[s.stage].Command == "table" && (ctx.SUB() != nil || ctx.ADD() != nil) {
		s.diagnostic(CodeUnsupportedSemantics, "table does not support signed projection", ctx)
		return
	}
	selected := map[string]trackedField{}
	if !exclude && s.result.Stages[s.stage].Command == "fields" {
		for name, field := range s.env.fields {
			if strings.HasPrefix(name, "_") {
				selected[name] = field
			}
		}
		if s.env.open {
			s.diagnostic(CodeUnsupportedSemantics, "fields inclusion retains internal fields with unresolved open-source membership", ctx)
		}
	}
	for _, c := range ctx.AnalysisFieldList().AllAnalysisSelector() {
		if exclude && len(c.AllMULT()) == 0 {
			name := normalizedName(c.GetText())
			id := s.reference(c, name, "field", "remove")
			s.env.remove(name)
			s.transitions = append(s.transitions, Transition{Operation: "remove", Output: name, InputReferenceIDs: []string{}, OutputReferenceID: id})
			continue
		}
		names, ids := s.selector(c, "read")
		for _, name := range names {
			if exclude {
				s.env.remove(name)
				s.transitions = append(s.transitions, Transition{Operation: "remove", Output: name, InputReferenceIDs: copyIDs(ids)})
			} else if f, ok := s.projectedField(name, ids); ok {
				selected[name] = f
				s.transitions = append(s.transitions, Transition{Operation: "project", Output: name, InputReferenceIDs: copyIDs(ids)})
			}
		}
	}
	if !exclude {
		s.env.fields = selected
		s.env.open = false
		if s.result.Stages[s.stage].Command == "table" && s.result.Stages[s.stage].SemanticComplete {
			s.env.uncertain = false
		}
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

func statsCommand(s *semanticStage, node antlr.ParserRuleContext) {
	ctx := node.(*parser.AnalysisStatsStageContext)
	if len(ctx.AllAnalysisOption()) > 0 {
		s.diagnostic(CodeUnsupportedSemantics, "aggregate command options are unmodeled", ctx)
	}
	output := newEnvironment()
	output.open = false
	type aggregate struct {
		ctx    antlr.ParserRuleContext
		name   string
		inputs []string
	}
	outputs := []aggregate{}
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
			outputs = append(outputs, aggregate{outCtx, name, ids})
		}
	}
	if group := ctx.AnalysisGroup(); group != nil {
		for _, c := range group.AnalysisFieldList().AllAnalysisSelector() {
			names, ids := s.selector(c, "group")
			for _, name := range names {
				if f, ok := s.projectedField(name, ids); ok {
					output.fields[name] = f
				}
				s.transitions = append(s.transitions, Transition{Operation: "project", Output: name, InputReferenceIDs: copyIDs(ids)})
			}
		}
	}
	if s.result.Stages[s.stage].Command == "stats" {
		output.uncertain = !s.result.Stages[s.stage].SemanticComplete
		s.env = output
	}
	for _, o := range outputs {
		s.create(o.ctx, o.name, "output", "aggregate", o.inputs, !s.result.Stages[s.stage].SemanticComplete)
	}
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
			conditional := out.OUTPUTNEW() != nil
			if conditional {
				if _, known := s.env.fields[name]; known {
					s.reference(local, name, "field", "output")
					continue
				}
			}
			s.create(local, name, "output", "lookup", ids, conditional || !s.result.Stages[s.stage].SemanticComplete)
		}
	}
}
func inputlookupCommand(s *semanticStage, node antlr.ParserRuleContext) {
	c := node.(*parser.AnalysisInputlookupStageContext)
	s.dependency(c.AnalysisCatalogName(), normalizedName(c.AnalysisCatalogName().GetText()), "lookup")
	s.env = newEnvironment()
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
