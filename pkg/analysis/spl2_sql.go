package analysis

import (
	"sort"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	"github.com/delgado-jacob/spl-toolkit/parser/spl2"
)

type spl2SQLCommand interface {
	spl2SQLClauses
	SqlSelectClause() spl2.ISqlSelectClauseContext
	SqlGroupClause() spl2.ISqlGroupClauseContext
	SqlOrderClause() spl2.ISqlOrderClauseContext
}

func spl2ScheduledSQL(c spl2SQLCommand) bool {
	return c.SqlSelectClause() != nil || c.SqlWhereClause() != nil || c.SqlGroupClause() != nil || c.SqlHavingClause() != nil || c.SqlOrderClause() != nil || c.SqlLimitClause() != nil || c.SqlOffsetClause() != nil
}

// Stages retain lexical clause order. Each phase uses its real owning clause;
// preparation and final restriction can therefore share the one SELECT stage.
func executeSPL2SQL(result *Result, parsed *spl2ParsedDocument, refinement *sourceRefinement, c spl2SQLCommand, env *environment, aliases map[string]bool, scheduler *spl2ScopeScheduler, scopeID string, parent, position int) *environment {
	logical := []antlr.ParserRuleContext{}
	for _, ctx := range []antlr.ParserRuleContext{c.SqlFromClause(), c.SqlWhereClause(), c.SqlGroupClause(), c.SqlSelectClause(), c.SqlHavingClause(), c.SqlOrderClause(), c.SqlLimitClause(), c.SqlOffsetClause()} {
		if ctx != nil {
			logical = append(logical, ctx)
		}
	}
	positions := map[antlr.ParserRuleContext]int{}
	for i, ctx := range logical {
		positions[ctx] = position + i
	}
	lexical := append([]antlr.ParserRuleContext{}, logical...)
	sort.SliceStable(lexical, func(i, j int) bool { return lexical[i].GetStart().GetStart() < lexical[j].GetStart().GetStart() })
	stages := map[antlr.ParserRuleContext]int{}
	for _, ctx := range lexical {
		stages[ctx] = registerSPL2Stage(result, parsed.source.contextLocation(ctx), strings.ToLower(ctx.GetStart().GetText()), positions[ctx], scopeID)
	}
	s := &spl2SemanticStage{semanticStage: &semanticStage{result: result, env: env, refinement: refinement}, parsed2: parsed, aliases: aliases}
	phase := func(ctx antlr.ParserRuleContext, name string, run func()) {
		if ctx == nil {
			return
		}
		s.stage = stages[ctx]
		s.transitions = []Transition{}
		before := s.env.snapshot()
		scheduler.runChildren(ctx, s.env, aliases, scopeID, parent)
		if spl2IntactSyntax(ctx) {
			run()
		} else {
			s.unsupported(ctx, "Recovered SQL clause effects are not yet modeled")
		}
		if !result.Stages[s.stage].SemanticComplete {
			s.env.uncertain = true
		}
		order := len(result.Lineage)
		st := result.Stages[s.stage]
		result.Lineage = append(result.Lineage, Lineage{StageID: st.ID, ScopeID: st.ScopeID, Before: before, After: s.env.snapshot(), Transitions: s.transitions, Phase: name, ExecutionOrder: &order})
	}
	phase(c.SqlFromClause(), "source", func() {
		s.applySource()
		if parent < 0 || scheduler.children[parent].input != "correlated" {
			clear(aliases)
		}
		from := c.SqlFromClause()
		s.dataset(from.Dataset())
		if a := from.SourceAlias(); a != nil {
			o := s.operand(a.Identifier())
			if o.Sound {
				aliases[o.Name] = true
			}
		}
		if len(from.AllSqlJoinClause()) > 0 {
			s.unsupported(from, "SQL join field effects are not yet modeled")
		}
	})
	phase(c.SqlWhereClause(), "filter", func() { s.expression(c.SqlWhereClause().SqlPredicate()) })
	pregroup := s.env
	phase(c.SqlGroupClause(), "group", func() {
		groups := []locatedOperand{}
		for _, key := range c.SqlGroupClause().AllSqlGroupKey() {
			field := spl2SQLDirectField(key.Expression())
			if field != nil && key.SqlSpanAssignment() == nil {
				groups = append(groups, s.selector(field))
			} else {
				s.expression(key)
				s.unsupported(key, "SQL grouping expression output is unproved")
			}
		}
		s.applyAggregation(nil, groups, false)
	})
	aggregate := spl2SQLHasAggregate(c.SqlSelectClause())
	selectPhase := "evaluate"
	if aggregate {
		selectPhase = "aggregate"
	}
	selected := []preparedSelection{}
	visible := map[string]bool{}
	phase(c.SqlSelectClause(), selectPhase, func() {
		selected, visible = s.prepareSQLSelection(c.SqlSelectClause(), pregroup, aggregate, c.SqlGroupClause() != nil)
	})
	phase(c.SqlHavingClause(), "having", func() {
		if c.SqlGroupClause() == nil {
			s.unsupported(c.SqlHavingClause(), "SQL HAVING without grouping is unproved")
		}
		s.sqlRestrictedExpression(c.SqlHavingClause().SqlPredicate(), visible)
	})
	phase(c.SqlOrderClause(), "order", func() {
		if c.SqlSelectClause() == nil {
			s.expression(c.SqlOrderClause())
		} else {
			s.sqlRestrictedExpression(c.SqlOrderClause(), visible)
		}
	})
	phase(c.SqlSelectClause(), "project", func() { s.applyPreparedProjection(selected, "table") })
	phase(c.SqlLimitClause(), "limit", func() {})
	phase(c.SqlOffsetClause(), "offset", func() {})
	return s.env
}

func spl2SQLDirectField(tree antlr.Tree) spl2.IIdentifierContext {
	access := spl2SQLFieldAccess(tree)
	if access != nil && len(access.AllAccessPart()) == 0 && access.Primary().FieldName() != nil {
		return access.Primary().FieldName().Identifier()
	}
	return nil
}

func spl2SQLHasAggregate(tree antlr.Tree) bool {
	if tree == nil {
		return false
	}
	if c, ok := tree.(spl2.ICallContext); ok && spl2Functions[c.Identifier().GetText()].aggregate {
		return true
	}
	// Child expressions own their own scheduling and aggregate context.
	switch tree.(type) {
	case spl2.IExistsPredicateContext, spl2.ISearchLiteralContext:
		return false
	}
	for _, child := range tree.GetChildren() {
		if spl2SQLHasAggregate(child) {
			return true
		}
	}
	return false
}

// Compound aggregate results remain unproved. Read their real arguments with
// the reviewed function policy, avoiding a false scalar-context contract error.
func (s *spl2SemanticStage) sqlUnprovedAggregateExpression(tree antlr.Tree) spl2ExpressionEvidence {
	if !spl2SQLHasAggregate(tree) {
		return s.expression(tree)
	}
	if c, ok := tree.(spl2.ICallContext); ok {
		aggregate := spl2Functions[c.Identifier().GetText()].aggregate
		out := s.callWithExpression(c, aggregate, s.sqlUnprovedAggregateExpression)
		if !aggregate {
			out.modeled, out.nonnull, out.exactNull, out.truth, out.domain = false, false, false, false, ""
		}
		return out
	}
	out := spl2ExpressionEvidence{ids: []string{}}
	for _, child := range tree.GetChildren() {
		out.ids = append(out.ids, s.sqlUnprovedAggregateExpression(child).ids...)
	}
	return out
}

func (s *spl2SemanticStage) prepareSQLSelection(clause spl2.ISqlSelectClauseContext, pregroup *environment, aggregate, grouped bool) ([]preparedSelection, map[string]bool) {
	type projection struct {
		ctx       spl2.IProjectionContext
		target    locatedOperand
		value     spl2ExpressionEvidence
		field     *trackedField
		aggregate bool
	}
	items := []projection{}
	input := s.env
	if aggregate && !grouped {
		s.applyAggregation(nil, nil, false)
		input = s.env
	}
	groups := map[string]bool{}
	for name := range input.fields {
		groups[name] = true
	}
	for _, p := range clause.AllProjection() {
		item := projection{ctx: p}
		field := spl2SQLDirectField(p.Expression())
		access := spl2SQLFieldAccess(p.Expression())
		var call spl2.ICallContext
		if access != nil && len(access.AllAccessPart()) == 0 {
			call = access.Primary().Call()
		}
		item.aggregate = call != nil && spl2Functions[call.Identifier().GetText()].aggregate
		if item.aggregate {
			s.env = pregroup
			item.value = s.call(call, true)
			s.env = input
		} else if spl2SQLHasAggregate(p.Expression()) {
			s.env = pregroup
			item.value = s.sqlUnprovedAggregateExpression(p.Expression())
			s.env = input
			s.unsupported(p, "Compound SQL aggregate output is unproved")
		} else if aggregate || grouped {
			if field == nil || !groups[s.operand(field).Name] {
				s.unsupported(p, "Mixed SQL aggregate/non-grouped projection is unproved")
			}
			item.value = s.sqlRestrictedExpression(p.Expression(), groups)
		} else {
			item.value = s.expression(p.Expression())
		}
		if p.ProjectionAlias() != nil {
			item.target = s.operand(p.ProjectionAlias().Identifier())
		} else if item.aggregate {
			name := call.Identifier().GetText()
			label := ""
			if name == "count" && call.Arguments() == nil {
				label = "count"
			}
			if args := call.Arguments(); args != nil && len(args.AllNamedArgument()) == 0 && len(args.AllExpression()) == 1 {
				f := spl2SQLDirectField(args.Expression(0))
				if f != nil && ((name == "sum" && f.GetText() == "bytes") || (name == "dc" && f.GetText() == "action")) {
					label = name + "(" + f.GetText() + ")"
				}
			}
			if label != "" {
				item.target = locatedOperand{Name: label, Location: s.parsed2.source.contextLocation(call), Resolution: "exact", Sound: true}
			} else {
				s.unsupported(p, "Implicit aggregate output label is unproved; use AS")
			}
		} else if field != nil {
			if f, ok := s.projectedField(s.operand(field).Name, item.value.ids); ok {
				item.field = &f
			}
		} else {
			s.unsupported(p, "Implicit SQL expression output label is unproved; use AS")
		}
		items = append(items, item)
	}
	// All expressions see input, never a sibling SELECT alias. Ambiguous output
	// names remain conditional and cannot authorize later alias visibility.
	counts := map[string]int{}
	for _, item := range items {
		if item.target.Sound {
			counts[item.target.Name]++
		}
		if item.field != nil {
			counts[item.field.Name]++
		}
	}
	collisions := map[string]bool{}
	for _, item := range items {
		if !item.target.Sound {
			continue
		}
		_, source := pregroup.fields[item.target.Name]
		if source || counts[item.target.Name] > 1 {
			collisions[item.target.Name] = true
			s.unsupported(item.ctx, "SQL alias collision or sibling alias visibility is unproved")
		}
	}
	selected := []preparedSelection{}
	visible := map[string]bool{}
	for _, item := range items {
		if item.field != nil {
			item.field.Conditional = item.field.Conditional || collisions[item.field.Name]
			selected = append(selected, preparedSelection{Field: *item.field, InputReferenceIDs: item.value.ids, EmitProjectTransition: true})
			visible[item.field.Name] = true
		}
		if !item.target.Sound {
			continue
		}
		before := len(s.result.References)
		if item.aggregate {
			s.applyAggregation([]aggregateOutput{{Target: item.target, InputReferenceIDs: item.value.ids, Conditional: !item.value.nonnull}}, nil, true)
		} else {
			s.applyAssignment(item.target, item.value.ids, !item.value.nonnull || collisions[item.target.Name], item.value.exactNull)
		}
		visible[item.target.Name] = !collisions[item.target.Name]
		if !item.value.exactNull {
			ids := []string{s.result.References[before].ID}
			if f, ok := s.projectedField(item.target.Name, ids); ok {
				selected = append(selected, preparedSelection{Field: f, InputReferenceIDs: ids, EmitProjectTransition: true})
			}
		}
	}
	return selected, visible
}

// A temporary visibility view keeps hidden source/group origins without
// claiming availability or mutating the actual phase state. The shared reader
// still decides source/derived/null-test/conditional binding for every operand.
func (s *spl2SemanticStage) sqlRestrictedExpression(tree antlr.Tree, visible map[string]bool) spl2ExpressionEvidence {
	actual := s.env
	s.env = actual.clone()
	var inspect func(antlr.Tree)
	inspect = func(node antlr.Tree) {
		switch c := node.(type) {
		case spl2.IExistsPredicateContext, spl2.ISearchLiteralContext:
			return
		case spl2.IFieldNameContext:
			if c.Identifier() != nil {
				o := s.operand(c.Identifier())
				if o.Sound && !visible[o.Name] {
					f := s.env.fields[o.Name]
					f.Name = o.Name
					f.Conditional = true
					s.env.fields[o.Name] = f
					s.diagnosticAt(CodeUnsupportedSemantics, "warning", "unsupported_semantics", "SQL field visibility outside selected/grouped outputs is unproved", o.Location, false)
					s.result.Stages[s.stage].SemanticComplete = false
				}
			}
		}
		for _, child := range node.GetChildren() {
			inspect(child)
		}
	}
	inspect(tree)
	var value spl2ExpressionEvidence
	if spl2SQLHasAggregate(tree) {
		s.unsupported(tree.(antlr.ParserRuleContext), "SQL postaggregate expression visibility is unproved")
		value = s.sqlUnprovedAggregateExpression(tree)
	} else {
		value = s.expression(tree)
	}
	s.env = actual
	return value
}
