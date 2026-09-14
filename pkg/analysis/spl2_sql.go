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
	provedKey := parsed.proveGroupKeyBeforeEmptyHaving(c)
	missingSelect := parsed.groupMissingSelectDiagnostic(c)
	selectedShape := false
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
	if provedKey != nil {
		index := stages[c.SqlGroupClause()]
		result.Stages[index].SemanticComplete = true
		for i := range result.Diagnostics {
			d := &result.Diagnostics[i]
			original := *d
			original.StageID = ""
			original.ScopeID = ""
			if original == provedKey.diagnostic {
				// This original parser finding remains document recovery evidence.
				d.StageID = ""
				d.ScopeID = ""
			} else if d.StageID == result.Stages[index].ID {
				result.Stages[index].SemanticComplete = false
			}
		}
	}
	s := &spl2SemanticStage{semanticStage: &semanticStage{result: result, env: env, refinement: refinement}, parsed2: parsed, aliases: aliases}
	phase := func(ctx antlr.ParserRuleContext, name string, run func()) {
		if ctx == nil {
			return
		}
		s.stage = stages[ctx]
		s.rewritePhase, s.rewriteOrdinal = name, 0
		s.transitions = []Transition{}
		before := s.env.snapshot()
		scheduler.runChildren(ctx, s.env, aliases, scopeID, parent)
		if spl2IntactSyntax(ctx) || (provedKey != nil && ctx == c.SqlGroupClause()) {
			if (provedKey != nil || missingSelect != nil) && ctx == c.SqlGroupClause() {
				local := *parsed
				local.diagnostics = []Diagnostic{}
				for _, d := range parsed.diagnostics {
					if (provedKey == nil || d != provedKey.diagnostic) && (missingSelect == nil || d != *missingSelect) {
						local.diagnostics = append(local.diagnostics, d)
					}
				}
				s.parsed2 = &local
			}
			run()
			s.parsed2 = parsed
		} else {
			if from, ok := ctx.(spl2.ISqlFromClauseContext); ok {
				s.sourceIntentions(from)
			} else if name != "project" {
				s.recoveredSQLInputs(ctx)
			}
			s.unsupported(ctx, "Recovered SQL clause effects are not yet modeled")
		}
		if !result.Stages[s.stage].SemanticComplete && !(name == "project" && selectedShape) {
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
		s.joinDatasetIntentions(from)
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
	phase(c.SqlWhereClause(), "filter", func() {
		s.rewritePredicate(c.SqlWhereClause().SqlPredicate(), "spl2", false, false)
		s.expression(c.SqlWhereClause().SqlPredicate())
	})
	pregroup := s.env
	phase(c.SqlGroupClause(), "group", func() {
		groups := []locatedOperand{}
		for _, key := range c.SqlGroupClause().AllSqlGroupKey() {
			field := spl2SQLDirectField(key.Expression())
			if provedKey != nil && key == provedKey.key {
				field = provedKey.identifier
			}
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
		selected, visible, selectedShape = s.prepareSQLSelection(c.SqlSelectClause(), pregroup, aggregate, c.SqlGroupClause() != nil)
	})
	phase(c.SqlHavingClause(), "having", func() {
		predicate := c.SqlGroupClause() != nil || aggregate
		if !predicate {
			s.unsupported(c.SqlHavingClause(), "SQL HAVING without grouping is unproved")
		}
		s.sqlRestrictedExpression(c.SqlHavingClause().SqlPredicate(), visible, predicate)
	})
	phase(c.SqlOrderClause(), "order", func() {
		if c.SqlSelectClause() == nil {
			s.expression(c.SqlOrderClause())
		} else {
			s.sqlRestrictedExpression(c.SqlOrderClause(), visible, false)
		}
	})
	phase(c.SqlSelectClause(), "project", func() {
		s.applyPreparedProjection(selected, "table")
		if selectedShape {
			s.env.open, s.env.uncertain = false, false
			fields := map[string]requirementField{}
			for _, selection := range selected {
				if field, ok := s.env.requirements.fields[selection.Field.Name]; ok {
					field.origins = append([]string{}, field.origins...)
					fields[selection.Field.Name] = field
				}
			}
			s.env.requirements.fields = fields
			s.env.requirements.open, s.env.requirements.uncertain = false, false
		}
	})
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

func (s *spl2SemanticStage) prepareSQLSelection(clause spl2.ISqlSelectClauseContext, pregroup *environment, aggregate, grouped bool) ([]preparedSelection, map[string]bool, bool) {
	type projection struct {
		ctx          spl2.IProjectionContext
		target       locatedOperand
		value        spl2ExpressionEvidence
		field        *trackedField
		aggregate    bool
		countCertain bool
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
			item.countCertain = s.sqlCountOutputSound(p, call, item.value)
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
			item.value = s.sqlRestrictedExpression(p.Expression(), groups, false)
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
				item.target = locatedOperand{Name: label, Location: s.parsed2.source.contextLocation(call), Resolution: "exact", Sound: true, rewrite: rewriteOwner{role: "implicit_output", location: s.parsed2.source.contextLocation(call), implicit: name}}
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
	ambiguousDestination := false
	for _, item := range items {
		if name, known := s.sqlProjectionCollisionName(item.ctx, item.target); known {
			counts[name]++
		} else {
			ambiguousDestination = true
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
	finiteShape := !ambiguousDestination && len(collisions) == 0 && s.sqlProjectionEffectSound(clause)
	incompleteInputs := map[string]bool{}
	if s.refinement != nil {
		for _, expansion := range s.refinement.expansions {
			if !expansion.Complete {
				incompleteInputs[expansion.ReferenceID] = true
			}
		}
	}
	for _, item := range items {
		for _, id := range s.origins(item.value.ids) {
			if incompleteInputs[id] {
				finiteShape = false
			}
		}
		if item.field != nil {
			item.field.Conditional = item.field.Conditional || collisions[item.field.Name]
			selected = append(selected, preparedSelection{Field: *item.field, InputReferenceIDs: item.value.ids, EmitProjectTransition: true})
			visible[item.field.Name] = true
		}
		if !item.target.Sound {
			if item.field == nil {
				finiteShape = false
			}
			continue
		}
		if item.target.Name == "" || item.target.Resolution != "exact" {
			finiteShape = false
		}
		before := len(s.result.References)
		if item.aggregate && item.countCertain && !collisions[item.target.Name] && !ambiguousDestination && item.target.Resolution == "exact" {
			// This one intact output has its own non-null proof; sibling/owner
			// limitations still govern coverage, visibility and all other outputs.
			s.createAt(item.target, "output", "aggregate", item.value.ids, false)
		} else if item.aggregate {
			s.applyAggregation([]aggregateOutput{{Target: item.target, InputReferenceIDs: item.value.ids, Conditional: !item.value.nonnull}}, nil, true)
		} else {
			s.applyAssignment(item.target, item.value.ids, !item.value.nonnull || collisions[item.target.Name], item.value.exactNull)
		}
		visible[item.target.Name] = !collisions[item.target.Name]
		if !item.value.exactNull {
			ids := []string{s.result.References[before].ID}
			if f, ok := s.projectedField(item.target.Name, ids); ok {
				selected = append(selected, preparedSelection{Field: f, InputReferenceIDs: ids, EmitProjectTransition: true})
			} else {
				finiteShape = false
			}
		}
	}
	return selected, visible, finiteShape
}

// A potential SQL destination participates in collision checks even when its
// field state is unproved. L19's static alias-qualified suffix policy is used
// only here: it cannot install a suffix field or authorize a source binding.
func (s *spl2SemanticStage) sqlProjectionCollisionName(p spl2.IProjectionContext, target locatedOperand) (string, bool) {
	local, _ := s.parsed2.recoveredCountView(p)
	if !local.soundOperand(p) {
		return "", false
	}
	if target.Sound && target.Resolution == "exact" {
		return target.Name, true
	}
	if p.ProjectionAlias() != nil {
		return "", false
	}
	access := spl2SQLFieldAccess(p.Expression())
	if access == nil || access.Primary().FieldName() == nil {
		return "", false
	}
	base := s.operand(access.Primary().FieldName().Identifier())
	if !base.Sound || base.Resolution != "exact" {
		return "", false
	}
	parts := access.AllAccessPart()
	if len(parts) == 0 {
		return base.Name, true
	}
	if len(parts) != 1 || !s.aliases[base.Name] || parts[0].DOT() == nil || parts[0].Identifier() == nil {
		return "", false
	}
	suffix := s.operand(parts[0].Identifier())
	return suffix.Name, suffix.Sound && suffix.Resolution == "exact"
}

// Per-output count proof is deliberately narrower than SELECT owner coverage.
func (s *spl2SemanticStage) sqlCountOutputSound(projection spl2.IProjectionContext, call spl2.ICallContext, value spl2ExpressionEvidence) bool {
	return value.modeled && value.nonnull && call.Identifier().GetText() == "count" && call.Arguments() == nil && s.sqlProjectionEffectSound(projection)
}

func (s *spl2SemanticStage) sqlProjectionEffectSound(ctx antlr.ParserRuleContext) bool {
	local, missingSelect := s.parsed2.recoveredCountView(ctx)
	if !local.soundOperand(ctx) {
		return false
	}
	location := s.parsed2.source.contextLocation(ctx)
	for _, d := range s.result.Diagnostics {
		original := d
		original.StageID, original.ScopeID = "", ""
		if missingSelect != nil && original == *missingSelect {
			continue
		}
		if d.Severity == "error" && d.Location.End.Offset >= location.Start.Offset && d.Location.Start.Offset <= location.End.Offset {
			return false
		}
	}
	return true
}

// A temporary visibility view keeps hidden source/group origins without
// claiming availability or mutating the actual phase state. The shared reader
// still decides source/derived/null-test/conditional binding for every operand.
func (s *spl2SemanticStage) sqlRestrictedExpression(tree antlr.Tree, visible map[string]bool, predicate bool) spl2ExpressionEvidence {
	actual := s.env
	s.env = actual.clone()
	hidden := []locatedOperand{}
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
					s.env.requirements.uncertain = true
					hidden = append(hidden, o)
				}
			}
		}
		for _, child := range node.GetChildren() {
			inspect(child)
		}
	}
	inspect(tree)
	if predicate {
		// Extract at the HAVING phase inside the same restricted source/derived
		// visibility view as its reads. Earlier SQL phases cannot see these facts.
		// A hidden operand must not suppress a conflicting visible-key restriction.
		s.rewritePredicate(tree, "spl2", false, false)
	}
	var value spl2ExpressionEvidence
	if spl2SQLHasAggregate(tree) {
		s.unsupported(tree.(antlr.ParserRuleContext), "SQL postaggregate expression visibility is unproved")
		value = s.sqlUnprovedAggregateExpression(tree)
	} else {
		value = s.expression(tree)
	}
	hiddenIDs := map[string][]string{}
	if trace := s.env.requirements.trace; trace != nil {
		for _, id := range value.ids {
			entry := trace.reference(id)
			hiddenIDs[entry.reference.NormalizedName] = append(hiddenIDs[entry.reference.NormalizedName], id)
		}
	}
	for _, operand := range hidden {
		owners := hiddenIDs[operand.Name]
		if len(owners) > 1 {
			hiddenIDs[operand.Name] = owners[1:]
		} else {
			hiddenIDs[operand.Name] = nil
		}
		if len(owners) > 0 {
			owners = owners[:1]
		}
		s.diagnosticAtOwned(CodeUnsupportedSemantics, "warning", "unsupported_semantics", "SQL field visibility outside selected/grouped outputs is unproved", operand.Location, true, owners)
	}
	if predicate {
		// Persist predicate/observed-reference evidence, not the temporary field
		// visibility or expression effects used to classify HAVING operands.
		actual.rewrite = s.env.rewrite
	}
	s.env = actual
	return value
}

// Preserve real expression inputs of damaged clauses once, without preparing
// any projection alias, group output, or final output shape.
func (s *spl2SemanticStage) recoveredSQLInputs(ctx antlr.ParserRuleContext) {
	switch c := ctx.(type) {
	case spl2.ISqlSelectClauseContext:
		for _, projection := range c.AllProjection() {
			if projection.Expression() != nil {
				s.sqlUnprovedAggregateExpression(projection.Expression())
			}
		}
	case spl2.ISqlGroupClauseContext:
		for _, key := range c.AllSqlGroupKey() {
			if span := key.SqlSpanCall(); span != nil {
				if field := span.FieldName(); field != nil {
					s.readIdentifier(field.Identifier(), "read")
				}
			} else if key.SqlSpanAssignment() != nil {
				if field := spl2SQLDirectField(key.Expression()); field != nil {
					s.readIdentifier(field, "read")
				}
			}
		}
	}
}
