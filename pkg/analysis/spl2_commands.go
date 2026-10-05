package analysis

import (
	"encoding/json"
	"github.com/antlr4-go/antlr/v4"
	"github.com/delgado-jacob/spl-toolkit/parser/spl2"
	"sort"
	"strconv"
	"strings"
)

func (s *spl2SemanticStage) selectedCommandHandler() (spl2CommandHandler, bool) {
	policy, selected := s.typedPolicy()
	if !selected || s.stage < 0 || s.stage >= len(s.result.Stages) {
		return 0, false
	}
	handler, ok := policy.commands[s.result.Stages[s.stage].Command]
	return handler, ok
}

func (s *spl2SemanticStage) selectedCommand(ctx antlr.ParserRuleContext, handler spl2CommandHandler) {
	switch handler {
	case spl2WhereCommandHandler:
		c, ok := ctx.(*spl2.WhereCommandContext)
		if !ok {
			s.unsupported(ctx, "Selected command policy does not match parsed command")
			return
		}
		if c.Expression() == nil || s.damagedExpressionContinuation(c.Expression()) {
			s.unsupported(c, "Recovered SPL2 command effects are not yet modeled")
			return
		}
		s.rewritePredicate(c.Expression(), "spl2", false, false)
		s.expression(c.Expression())
	case spl2EvalCommandHandler:
		c, ok := ctx.(*spl2.EvalCommandContext)
		if !ok {
			s.unsupported(ctx, "Selected command policy does not match parsed command")
			return
		}
		for _, a := range c.AllAssignment() {
			if s.damagedExpressionContinuation(a.Expression()) {
				s.unsupported(a, "Recovered SPL2 assignment effects are not yet modeled")
				continue
			}
			value := s.expression(a.Expression())
			if !s.assignmentEffectSound(a) {
				continue
			}
			if a.FieldName().Identifier() == nil {
				s.expression(a.FieldName())
				s.unsupported(a.FieldName(), "Computed assignment target is unresolved")
				continue
			}
			s.applyAssignmentWithRequirementConditional(s.operand(a.FieldName().Identifier()), value.ids, !value.nonnull, !value.requirementNonnull, value.exactNull)
		}
	case spl2FieldsCommandHandler:
		c, ok := ctx.(*spl2.FieldsCommandContext)
		if !ok {
			s.unsupported(ctx, "Selected command policy does not match parsed command")
			return
		}
		selection := c.FieldSelection()
		fields := []locatedOperand{}
		priorFields := map[fieldIdentityKey]bool{}
		for key := range s.env.fields {
			priorFields[key] = true
		}
		priorExpansionCount := 0
		if s.refinement != nil {
			priorExpansionCount = len(s.refinement.expansions)
		}
		unproved := false
		for _, f := range selection.AllFieldSelector() {
			o := locatedOperand{}
			if f.Identifier() != nil {
				o = s.selector(f.Identifier())
			} else if f.StructuralFieldSelector() != nil {
				o = s.structuralSelector(f.StructuralFieldSelector())
			}
			if o.Sound && o.Name != "" {
				fields = append(fields, o)
			} else {
				unproved = true
			}
		}
		mode := "include"
		if selection.MINUS() != nil {
			mode = "exclude"
		}
		if unproved {
			s.unprovedProjection(c, fields)
		} else {
			s.applyProjection(fields, mode, false)
			if mode == "include" {
				s.closeSelectedFieldsWildcard(fields, priorFields, priorExpansionCount)
			}
		}
	case spl2StatsCommandHandler:
		c, ok := ctx.(*spl2.StatsCommandContext)
		if !ok {
			s.unsupported(ctx, "Selected command policy does not match parsed command")
			return
		}
		allnum := []spl2.IAllnumOptionContext{}
		for _, o := range c.AllStatsOption() {
			if a := o.AllnumOption(); a != nil {
				allnum = append(allnum, a)
			}
			if o.UnknownOption() != nil {
				s.unsupported(o, "Unmodeled stats option")
			}
		}
		s.aggregates(c.AllAggregate(), spl2Groups(c.AggregateGroup()), c.SelectedAggregateGroup(), false, s.allnum(allnum))
	case spl2BinCommandHandler:
		c, ok := ctx.(*spl2.BinCommandContext)
		if !ok {
			s.unsupported(ctx, "Selected command policy does not match parsed command")
			return
		}
		input := s.operand(c.Identifier(0))
		id := s.readAt(input, "read")
		if len(c.AllExtendedOption()) > 0 || !s.commandEffectSound(c) {
			s.rejectSelectedIdentityEffect(c, input, id)
			return
		}
		if id == "" {
			return
		}
		inputReference := s.result.References[len(s.result.References)-1]
		requirementBinding := inputReference.Binding
		requirementConditional := inputReference.Binding == "indeterminate"
		if trace := s.env.requirements.trace; trace != nil {
			entry := trace.reference(id)
			requirementBinding = entry.reference.Binding
			requirementConditional = entry.conditional
		}
		if inputReference.Binding == "unavailable" || requirementBinding == "unavailable" {
			return
		}
		target := input
		if len(c.AllIdentifier()) > 1 {
			target = s.operand(c.Identifier(1))
		}
		s.createAtWithRequirementConditional(target, "output", "bin", []string{id}, inputReference.Binding == "indeterminate", requirementConditional)
	case spl2MvexpandCommandHandler:
		c, ok := ctx.(*spl2.MvexpandCommandContext)
		if !ok {
			s.unsupported(ctx, "Selected command policy does not match parsed command")
			return
		}
		o := s.operand(c.Identifier())
		id := s.readAt(o, "read")
		if len(c.AllExtendedOption()) > 0 || !s.commandEffectSound(c) {
			s.rejectSelectedIdentityEffect(c, o, id)
			return
		}
		if id == "" {
			return
		}
		inputReference := s.result.References[len(s.result.References)-1]
		requirementBinding := inputReference.Binding
		if trace := s.env.requirements.trace; trace != nil {
			requirementBinding = trace.reference(id).reference.Binding
		}
		if inputReference.Binding == "unavailable" || requirementBinding == "unavailable" {
			s.unsupportedOwned(c, "mvexpand cannot select an unavailable field", []string{id})
			return
		}
		if !s.env.selectElement(o.fieldIdentity()) {
			s.unsupportedOwned(c, "mvexpand input value state is unresolved", []string{id})
		}
	}
}

func (s *spl2SemanticStage) command(ctx antlr.ParserRuleContext) {
	if handler, selected := s.selectedCommandHandler(); selected {
		s.selectedCommand(ctx, handler)
		return
	}
	switch c := ctx.(type) {
	case *spl2.FromCommandContext:
		s.applySource()
		clear(s.aliases)
		source := c.SqlFromClause()
		if !s.exactDatasetSource(source.Dataset()) {
			s.dataset(source.Dataset())
		}
		s.joinDatasetIntentions(source)
		if alias := source.SourceAlias(); alias != nil {
			o := s.operand(alias.Identifier())
			if o.Sound {
				s.aliases[o.Name] = true
			}
		}
		if len(source.AllSqlJoinClause()) > 0 || c.SqlSelectClause() != nil || c.SqlWhereClause() != nil || c.SqlGroupClause() != nil || c.SqlHavingClause() != nil || c.SqlOrderClause() != nil || c.SqlLimitClause() != nil || c.SqlOffsetClause() != nil {
			s.unsupported(c, "SQL clause field scheduling is not yet modeled")
		}
	case *spl2.SearchCommandContext:
		s.rewritePredicate(c.SearchExpression(), "spl2", true, false)
		s.search(c.SearchExpression())
	case *spl2.ImplicitSearchContext:
		s.rewritePredicate(c.SearchExpression(), "spl2", true, false)
		s.search(c.SearchExpression())
	case *spl2.TableCommandContext:
		fields := []locatedOperand{}
		unproved := false
		for _, f := range c.AllTableField() {
			if f.Identifier() == nil {
				// Double-quoted expressions are not exact field selectors. Their
				// interpolation children can still contain real input reads.
				if literal := f.StringLiteral(); literal != nil {
					for _, child := range literal.AllExpression() {
						s.expression(child)
					}
				}
				unproved = true
				continue
			}
			o := s.selector(f.Identifier())
			if o.Sound && o.Name != "" {
				fields = append(fields, o)
			} else {
				unproved = true
			}
		}
		if unproved {
			s.unprovedProjection(c, fields)
		} else {
			s.applyProjection(fields, "table", false)
		}
	case *spl2.RenameCommandContext:
		pairs := []renameOperands{}
		for _, pair := range c.AllRenamePair() {
			pairs = append(pairs, renameOperands{s.selector(pair.RenameSource()), s.selector(pair.RenameTarget())})
		}
		s.applyRename(pairs)
	case *spl2.EventstatsCommandContext:
		s.aggregates(c.AllAggregate(), spl2Groups(c.AggregateGroup()), nil, true, s.allnum(c.AllAllnumOption()))
	case *spl2.StreamstatsCommandContext:
		groups := []spl2.IGroupFieldContext{}
		if g := c.StreamGroup(); g != nil {
			groups = g.AllGroupField()
		}
		if r := c.ResetClause(); r != nil {
			if b := r.ResetBefore(); b != nil {
				s.expression(b.Expression())
			}
			if a := r.ResetAfter(); a != nil {
				s.expression(a.Expression())
			}
		}
		conditional := c.CurrentOption() != nil && c.CurrentOption().BOOLEAN().GetText() == "false"
		s.aggregates(c.AllAggregate(), groups, nil, true, conditional)
	case *spl2.LookupCommandContext:
		s.lookup(c)
	case *spl2.SortCommandContext:
		for _, term := range c.AllSortTerm() {
			s.readIdentifier(term.Identifier(), "read")
		}
	case *spl2.DedupCommandContext:
		for _, field := range c.AllDedupField() {
			s.readIdentifier(field.Identifier(), "read")
		}
	case *spl2.HeadCommandContext:
		if w := c.HeadWhile(); w != nil {
			s.expression(w.Expression())
		}
	case *spl2.IfCommandContext:
		for _, condition := range c.AllExpression() {
			s.expression(condition)
		}
		s.unsupported(c, "Conditional child merge output is unproved")
	case *spl2.ReverseCommandContext:
	case *spl2.UnionCommandContext:
		s.unionDatasetIntentions(c)
		s.unsupported(c, "Union output merge effects are unproved")
	case *spl2.FillnullCommandContext:
		for _, field := range c.AllIdentifier() {
			s.operandReference(s.selector(field), "field", "output")
		}
		s.unsupported(c, "Fillnull output presence is unproved")
	case *spl2.LoadjobCommandContext:
		o := locatedOperand{}
		if number := c.NUMBER(); number != nil {
			token := number.GetSymbol()
			o = locatedOperand{Name: token.GetText(), Location: s.parsed2.source.location(token.GetStart(), token.GetStop()+1), Resolution: "exact", Sound: token.GetTokenIndex() >= 0}
		} else if c.Identifier() != nil {
			o = s.operand(c.Identifier())
		} else if literal := c.StringLiteral(); literal != nil && len(literal.AllExpression()) == 0 {
			o = s.operand(literal)
		}
		s.operandReference(o, "search_job", "read")
		s.unsupported(c, "External job output fields are unproved")
	case *spl2.JoinCommandContext:
		ids := s.joinPredicateIntentions(c)
		s.unsupportedOwned(c, "Join output merge and qualified input binding are unproved", ids)
	case *spl2.RexCommandContext:
		input, sed := "_raw", false
		for _, option := range c.AllRexOption() {
			if option.FIELD() != nil {
				o := s.operand(option.Identifier())
				s.readAt(o, "read")
				if o.Sound {
					input = o.Name
				}
			}
			if option.OFFSET_FIELD() != nil {
				s.operandReference(s.operand(option.Identifier()), "field", "output")
			}
			sed = sed || option.SED() != nil
		}
		var affected map[string]bool
		if sed {
			affected = map[string]bool{input: true}
		}
		s.deferredEffects(c, affected, false)
	case *spl2.SpathCommandContext:
		var affected map[string]bool
		for _, option := range c.AllSpathOption() {
			if option.INPUT() != nil {
				s.readIdentifier(option.Identifier(), "read")
			}
			if option.OUTPUT_LOWER() != nil {
				o := s.operand(option.Identifier())
				s.operandReference(o, "field", "output")
				if affected == nil {
					affected = map[string]bool{}
				}
				if o.Sound {
					affected[o.Name] = true
				}
			}
		}
		s.deferredEffects(c, affected, false)
	case *spl2.MetricsCommandContext:
		s.metricsInputs(c)
		s.deferredEffects(c, nil, true)
	case *spl2.TimechartCommandContext:
		s.timechartInputs(c)
		s.deferredEffects(c, nil, true)
	case *spl2.MakemvCommandContext:
		o := s.operand(c.Identifier())
		s.readAt(o, "read")
		s.deferredEffects(c, map[string]bool{o.Name: o.Sound}, false)
	case *spl2.MvcombineCommandContext:
		o := s.operand(c.Identifier())
		s.readAt(o, "read")
		s.deferredEffects(c, map[string]bool{o.Name: o.Sound}, false)
	default:
		s.unsupported(ctx, "Command field effects are not yet modeled")
	}
}

func (q *spl2ScopeScheduler) lowerSelectedFlowCommand(s *spl2SemanticStage, ctx antlr.ParserRuleContext, aliases map[string]bool, scopeID string, parent int) {
	switch command := ctx.(type) {
	case *spl2.IfCommandContext:
		q.lowerSelectedIf(s, command, aliases, scopeID, parent)
	case *spl2.BranchCommandContext:
		q.lowerSelectedBranch(s, command, aliases, scopeID, parent)
	case *spl2.UnionCommandContext:
		q.lowerSelectedUnion(s, command, aliases, scopeID, parent)
	case *spl2.JoinCommandContext:
		q.lowerSelectedJoin(s, command, aliases, scopeID, parent)
	default:
		s.command(ctx)
	}
}

func (q *spl2ScopeScheduler) lowerRecoveredSelectedFlowCommand(s *spl2SemanticStage, ctx antlr.ParserRuleContext, aliases map[string]bool, scopeID string, parent int, location Location) {
	base := s.env
	paths := []flowMergePath{}
	schedule := func(ordinal int, child antlr.ParserRuleContext) {
		checkpoint := q.selectedFlowTraceCheckpoint(base)
		execution, ok := q.executeDirectChild(child, base, aliases, scopeID, parent)
		q.reconcileSelectedFlowTrace(checkpoint, base, paths)
		if ok {
			paths = append(paths, flowMergePath{Ordinal: ordinal, Environment: execution.Environment, Reachable: true})
		}
	}

	switch command := ctx.(type) {
	case *spl2.IfCommandContext:
		expressions, children := command.AllExpression(), command.AllInheritedSubpipe()
		for _, expression := range expressions {
			if spl2IntactSyntax(expression) {
				s.expression(expression)
			}
		}
		for i := 0; i < len(expressions) && i < len(children); i++ {
			if spl2IntactSyntax(expressions[i]) && spl2IntactSyntax(children[i]) {
				schedule(i, children[i])
			}
		}
	case *spl2.BranchCommandContext:
		for _, arm := range command.AllBranchArm() {
			if arm.Expression() != nil && spl2IntactSyntax(arm.Expression()) {
				s.expression(arm.Expression())
			}
		}
		for ordinal, arm := range command.AllBranchArm() {
			if arm.Expression() != nil && arm.InheritedSubpipe() != nil && spl2IntactSyntax(arm.Expression()) && spl2IntactSyntax(arm.InheritedSubpipe()) {
				schedule(ordinal, arm.InheritedSubpipe())
			}
		}
	}

	s.installSelectedAlternativeMerge(ctx, base, paths, true)
	s.env.uncertain = true
	s.env.requirements.uncertain = true
	s.diagnosticAtOwned(CodeUnsupportedSemantics, "warning", "unsupported_semantics", "Recovered SPL2 command effects are not yet modeled", location, true, nil)
}

func (q *spl2ScopeScheduler) lowerSelectedIf(s *spl2SemanticStage, command *spl2.IfCommandContext, aliases map[string]bool, scopeID string, parent int) {
	for _, condition := range command.AllExpression() {
		s.expression(condition)
	}
	base := s.env
	paths := make([]flowMergePath, 0, len(command.AllInheritedSubpipe()))
	for ordinal, child := range command.AllInheritedSubpipe() {
		checkpoint := q.selectedFlowTraceCheckpoint(base)
		execution, ok := q.executeDirectChild(child, base, aliases, scopeID, parent)
		q.reconcileSelectedFlowTrace(checkpoint, base, paths)
		if !ok {
			s.unsupported(command, "Conditional child scope could not be scheduled")
			continue
		}
		paths = append(paths, flowMergePath{Ordinal: ordinal, Environment: execution.Environment, Reachable: true})
	}
	s.installSelectedAlternativeMerge(command, base, paths, command.ELSE() == nil)
}

func (q *spl2ScopeScheduler) lowerSelectedBranch(s *spl2SemanticStage, command *spl2.BranchCommandContext, aliases map[string]bool, scopeID string, parent int) {
	base := s.env
	for _, arm := range command.AllBranchArm() {
		s.expression(arm.Expression())
	}
	paths := make([]flowMergePath, 0, len(command.AllBranchArm()))
	for ordinal, arm := range command.AllBranchArm() {
		checkpoint := q.selectedFlowTraceCheckpoint(base)
		execution, ok := q.executeDirectChild(arm.InheritedSubpipe(), base, aliases, scopeID, parent)
		q.reconcileSelectedFlowTrace(checkpoint, base, paths)
		if !ok {
			s.unsupported(command, "Guarded branch child scope could not be scheduled")
			continue
		}
		paths = append(paths, flowMergePath{Ordinal: ordinal, Environment: execution.Environment, Reachable: true})
	}
	s.installSelectedAlternativeMerge(command, base, paths, false)
}

func (q *spl2ScopeScheduler) lowerSelectedUnion(s *spl2SemanticStage, command *spl2.UnionCommandContext, aliases map[string]bool, scopeID string, parent int) {
	base := s.env
	paths := make([]flowMergePath, 0, len(command.AllUnionDataset()))
	for ordinal, input := range command.AllUnionDataset() {
		checkpoint := q.selectedFlowTraceCheckpoint(base)
		if child := input.IndependentSearch(); child != nil {
			execution, ok := q.executeDirectChild(child, base, aliases, scopeID, parent)
			q.reconcileSelectedFlowTrace(checkpoint, base, paths)
			if !ok {
				s.unsupported(input, "Union child scope could not be scheduled")
				continue
			}
			paths = append(paths, flowMergePath{Ordinal: ordinal, Environment: execution.Environment, Reachable: true})
			continue
		}

		trace := (*requirementTrace)(nil)
		if base.requirements.trace != nil {
			trace = base.requirements.trace.forkBranch()
		}
		branch := newEnvironmentWithRequirementTrace(trace)
		branchStage := &spl2SemanticStage{
			semanticStage: &semanticStage{result: s.result, stage: s.stage, env: branch, transitions: []Transition{}, refinement: s.refinement},
			parsed2:       s.parsed2,
			aliases:       map[string]bool{},
			locals:        map[string]bool{},
			program:       s.program,
		}
		dataset := input.Dataset()
		resolved := false
		if branchStage.program != nil && dataset != nil {
			resolved = branchStage.program.resolveViewSource(branchStage, dataset.DatasetParameter()) || branchStage.program.resolveImportedDataset(branchStage, dataset)
		}
		if !resolved && (dataset == nil || !branchStage.exactDatasetSource(dataset)) {
			branchStage.unsupported(input, "Union operand is not an exact dataset")
		}
		q.reconcileSelectedFlowTrace(checkpoint, base, paths)
		paths = append(paths, flowMergePath{Ordinal: ordinal, Environment: branchStage.env, Reachable: true})
	}
	s.installSelectedAlternativeMerge(command, base, paths, s.selectedUnionIncludesParent(command))
}

type selectedFlowTraceCheckpoint struct {
	base      *requirementTrace
	canonical *requirementTrace
}

func (q *spl2ScopeScheduler) selectedFlowTraceCheckpoint(base *environment) selectedFlowTraceCheckpoint {
	checkpoint := selectedFlowTraceCheckpoint{}
	if base != nil && base.requirements.trace != nil {
		checkpoint.base = base.requirements.trace.clone()
	}
	if q.trace != nil {
		checkpoint.canonical = q.trace.clone()
	}
	return checkpoint
}

// Lazy view binding can extend the canonical trace while an alternative is
// executing. Keep the shared base and every earlier sibling on that new prefix;
// the just-executed path is already rebased by its child or source binder.
func (q *spl2ScopeScheduler) reconcileSelectedFlowTrace(checkpoint selectedFlowTraceCheckpoint, base *environment, paths []flowMergePath) {
	if base == nil || base.requirements.trace == nil || checkpoint.base == nil {
		return
	}
	parserDiagnostics := q.result.Diagnostics[:q.initialDiagnosticCount]
	checkpoint.base.syncParserDiagnostics(parserDiagnostics)
	if checkpoint.canonical != nil {
		checkpoint.canonical.syncParserDiagnostics(parserDiagnostics)
	}
	if q.trace != nil && base.requirements.trace != q.trace && !requirementTraceHasPrefix(q.trace, base.requirements.trace) {
		if checkpoint.canonical == nil {
			return
		}
		base.requirements.trace = rebaseRequirementTrace(checkpoint.canonical, q.trace, base.requirements.trace)
	}
	for i := range paths {
		path := paths[i].Environment
		if path == nil || path.requirements.trace == nil {
			continue
		}
		path.requirements.trace.syncParserDiagnostics(parserDiagnostics)
		if requirementTraceHasPrefix(base.requirements.trace, path.requirements.trace) {
			continue
		}
		path.requirements.trace = rebaseRequirementTrace(checkpoint.base, base.requirements.trace, path.requirements.trace)
	}
}

func (s *spl2SemanticStage) selectedUnionIncludesParent(command *spl2.UnionCommandContext) bool {
	if s.stage >= 0 && s.stage < len(s.result.Stages) && s.result.Stages[s.stage].Position > 0 {
		return true
	}
	for ancestor := antlr.Tree(command.GetParent()); ancestor != nil; ancestor = ancestor.GetParent() {
		switch ancestor.(type) {
		case *spl2.InheritedSubpipeContext:
			return true
		case *spl2.IndependentSearchContext:
			return false
		}
	}
	return false
}

type spl2SelectedJoin struct {
	leftAlias  string
	rightAlias string
	joinType   string
	selected   bool
}

func (s *spl2SemanticStage) selectedJoin(command *spl2.JoinCommandContext) spl2SelectedJoin {
	selection := spl2SelectedJoin{selected: true, joinType: "inner"}
	leftCount, rightCount, typeCount, maxCount := 0, 0, 0, 0
	for _, option := range command.AllJoinOption() {
		switch {
		case option.LEFT() != nil:
			leftCount++
			operand := s.operand(option.Identifier())
			if operand.Sound {
				selection.leftAlias = operand.Name
			} else {
				selection.selected = false
			}
		case option.RIGHT() != nil:
			rightCount++
			operand := s.operand(option.Identifier())
			if operand.Sound {
				selection.rightAlias = operand.Name
			} else {
				selection.selected = false
			}
		case option.MAX() != nil:
			maxCount++
			value := option.IntegerValue()
			if value == nil || !s.parsed2.soundOperand(value) || !spl2SQLInteger(strings.TrimPrefix(value.GetText(), "+")) {
				selection.selected = false
			}
		case option.TYPE_OPTION() != nil:
			typeCount++
			joinType := option.JoinType()
			switch {
			case joinType != nil && joinType.INNER() != nil:
				selection.joinType = "inner"
			case joinType != nil && joinType.LEFT() != nil:
				selection.joinType = "left"
			case joinType != nil && joinType.OUTER() != nil:
				selection.joinType = "outer"
			default:
				selection.selected = false
			}
		default:
			selection.selected = false
		}
	}
	selection.selected = selection.selected && leftCount == 1 && rightCount == 1 && typeCount <= 1 && maxCount <= 1 && selection.leftAlias != "" && selection.rightAlias != "" && selection.leftAlias != selection.rightAlias
	return selection
}

func (q *spl2ScopeScheduler) lowerSelectedJoin(s *spl2SemanticStage, command *spl2.JoinCommandContext, aliases map[string]bool, scopeID string, parent int) {
	right, childOK := q.executeDirectChild(command.IndependentSearch(), s.env, aliases, scopeID, parent)
	baseAfterChild := (*requirementTrace)(nil)
	if s.env.requirements.trace != nil {
		baseAfterChild = s.env.requirements.trace.clone()
	}
	selection := s.selectedJoin(command)
	if !selection.selected {
		joinReferenceIDs := s.joinPredicateIntentions(command)
		if childOK {
			s.retainSelectedChildTrace(baseAfterChild, right.Trace)
			s.env.retainHeldJoinCandidates(right.Environment)
		}
		s.unsupportedOwned(command, "Join output merge and qualified input binding are unproved", joinReferenceIDs)
		return
	}
	joinReferenceIDs, predicateOK := s.selectedJoinPredicate(command, selection, s.env, right.Environment)

	if !childOK || !predicateOK {
		if childOK {
			s.retainSelectedChildTrace(baseAfterChild, right.Trace)
			s.env.retainHeldJoinCandidates(right.Environment)
		}
		s.unsupportedOwned(command, "Join layout is outside selected semantics", joinReferenceIDs)
		return
	}

	s.recordJoinCorrelations(command, joinReferenceIDs)
	combinedTrace := s.rebasedSelectedChildTrace(baseAfterChild, right.Trace)
	rightMatched := right.Environment.cloneWithRequirementTrace(combinedTrace)
	matched, collisions, composed := composeFlowEnvironments(s.env, rightMatched)
	if !composed {
		s.retainSelectedChildTrace(baseAfterChild, right.Trace)
		s.env.retainHeldJoinCandidates(right.Environment)
		s.unsupportedOwned(command, "Join child requirement trace does not extend the current parent", joinReferenceIDs)
		return
	}
	if len(collisions) != 0 {
		matched.uncertain = true
		matched.requirements.uncertain = true
	}

	mergeBase := s.env.cloneWithRequirementTrace(combinedTrace.clone())
	paths := []flowMergePath{{Ordinal: 0, Environment: matched, Reachable: true}}
	switch selection.joinType {
	case "left", "outer":
		leftOnly := s.env.cloneWithRequirementTrace(combinedTrace.clone())
		paths = append(paths, flowMergePath{Ordinal: 1, Environment: leftOnly, Reachable: true})
	}
	base := s.env
	s.installSelectedFlowMerge(base, mergeFlowEnvironments(mergeBase, paths, false))
	for _, name := range collisions {
		origins := selectedFlowFieldOrigins(s.env, name)
		s.diagnosticAtOwned(CodeAmbiguousField, "warning", "unsupported_semantics", "Join output contains fields with the same public name", s.parsed2.source.contextLocation(command), true, origins)
	}
}

func selectedFlowFieldOrigins(env *environment, name string) []string {
	origins := []string{}
	if env == nil {
		return origins
	}
	for _, key := range env.orderedFieldKeys() {
		field, ok := env.fields[key]
		if ok && field.Name == name {
			origins = uniqueIDs(origins, field.OriginReferenceIDs)
		}
	}
	return origins
}

func (s *spl2SemanticStage) installSelectedFlowMerge(base, merged *environment) {
	if base != nil && merged != nil && base.requirements.trace != nil && merged.requirements.trace != nil {
		canonical := base.requirements.trace
		*canonical = *merged.requirements.trace
		merged.requirements.trace = canonical
	}
	s.env = merged
}

func (s *spl2SemanticStage) installSelectedAlternativeMerge(owner antlr.ParserRuleContext, base *environment, paths []flowMergePath, includeParent bool) {
	merged := mergeFlowEnvironments(base, paths, includeParent)
	s.installSelectedFlowMerge(base, merged)
	names := make([]string, 0, len(merged.ambiguous))
	for name, ambiguous := range merged.ambiguous {
		if !ambiguous || base != nil && base.ambiguous[name] {
			continue
		}
		alreadyOwned := false
		for _, path := range paths {
			alreadyOwned = alreadyOwned || path.Environment != nil && path.Environment.ambiguous[name]
		}
		if !alreadyOwned {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	for _, name := range names {
		_, origins := merged.detectCollision(name)
		s.fieldIdentityCollision(name, s.parsed2.source.contextLocation(owner), origins)
	}
}

func (s *spl2SemanticStage) retainSelectedChildTrace(oldBase, child *requirementTrace) {
	if child == nil || s.env.requirements.trace == nil {
		return
	}
	if oldBase != nil {
		child = rebaseRequirementTrace(oldBase, s.env.requirements.trace, child)
	}
	merged := mergeRequirementTraces(s.env.requirements.trace, []requirementTracePath{{Ordinal: 0, Trace: child, Reachable: true}})
	canonical := s.env.requirements.trace
	*canonical = *merged
	s.env.requirements.trace = canonical
}

func (s *spl2SemanticStage) rebasedSelectedChildTrace(oldBase, child *requirementTrace) *requirementTrace {
	if child == nil || s.env.requirements.trace == nil {
		return s.env.requirements.trace
	}
	if oldBase == nil {
		return child
	}
	return rebaseRequirementTrace(oldBase, s.env.requirements.trace, child)
}

func (s *spl2SemanticStage) selectedJoinPredicate(command *spl2.JoinCommandContext, selection spl2SelectedJoin, left, right *environment) ([]string, bool) {
	ids := []string{}
	if command == nil || command.SqlJoinPredicate() == nil || left == nil || right == nil {
		return ids, false
	}
	valid := selection.leftAlias != "" && selection.rightAlias != ""
	for _, equality := range command.SqlJoinPredicate().AllSqlJoinEquality() {
		fields := equality.AllSqlJoinField()
		if len(fields) != 2 {
			valid = false
			continue
		}
		sides := map[string]bool{}
		for _, field := range fields {
			identifiers := field.AllIdentifier()
			if !s.parsed2.soundOperand(field) || len(field.AllAccessPart()) != 0 || len(identifiers) != 2 {
				valid = false
				continue
			}
			qualifier := s.operand(identifiers[0])
			name := s.operand(identifiers[1])
			if !qualifier.Sound || !name.Sound {
				valid = false
				continue
			}
			var source *environment
			switch qualifier.Name {
			case selection.leftAlias:
				source = left
				sides["left"] = true
			case selection.rightAlias:
				source = right
				sides["right"] = true
			default:
				valid = false
				continue
			}
			operand := locatedOperand{
				Name:       name.Name,
				Identity:   pathFieldIdentity(qualifier.Name, []string{name.Name}),
				Location:   s.parsed2.source.contextLocation(field),
				Resolution: "exact",
				Sound:      true,
			}
			id := s.selectedJoinReference(operand, source, atomicFieldIdentity(name.Name))
			if id == "" {
				valid = false
				continue
			}
			ids = append(ids, id)
		}
		if !sides["left"] || !sides["right"] {
			valid = false
		}
	}
	return ids, valid
}

func (s *spl2SemanticStage) selectedJoinReference(operand locatedOperand, source *environment, identity fieldIdentity) string {
	id := s.referenceAt(operand.Location, operand.Name, "field", "read", operand.Resolution)
	if id == "" {
		return ""
	}
	ref := &s.result.References[len(s.result.References)-1]
	if publicIdentity, exact := operand.Identity.public(); exact {
		ref.FieldIdentity = &publicIdentity
	}
	requirementBinding, directExternal, requirementConditional := "not_applicable", false, false
	if trace := s.env.requirements.trace; trace != nil {
		entry := trace.reference(id)
		entry.fieldIdentity = identity.clone()
		entry.reference.FieldIdentity = ref.FieldIdentity
		entry.reference = cloneTraceReference(entry.reference)
		entry.owners = source.requirements.sourceOwners(identity)
		requirementBinding, directExternal, requirementConditional = source.requirements.readIdentity(entry.reference, identity)
		entry.reference.Binding = requirementBinding
		entry.directExternal = directExternal
		entry.conditional = requirementConditional
	}

	binding, origins := "source", []string{}
	field, known := source.field(identity)
	switch {
	case known && (field.Conditional || field.ownerCollision) || source.uncertain:
		binding = "indeterminate"
		if known {
			origins = copyIDs(field.OriginReferenceIDs)
		}
	case known:
		binding = "derived"
		if field.source {
			binding = "source"
		}
		origins = copyIDs(field.OriginReferenceIDs)
	case !source.open:
		binding = "unavailable"
	default:
		collision, owners := source.installIdentity(identity, []string{id}, false, true)
		if collision {
			s.fieldIdentityCollision(operand.Name, operand.Location, owners)
		}
	}
	ref.Binding = binding
	ref.OriginReferenceIDs = origins
	s.rewriteReference(id, operand, "field", "read")
	s.rewriteIdentityCoverage(operand, "field")
	s.rewriteBinding(id, ref.Binding, origins)
	message := "join equality field is unavailable on its declared side"
	switch {
	case binding == "unavailable" && requirementBinding == "unavailable":
		s.diagnosticAtOwned(CodeUnavailableField, "error", "unavailable_field", message, operand.Location, false, []string{id})
	case binding == "unavailable":
		s.appendDiagnostic(CodeUnavailableField, "error", "unavailable_field", message, operand.Location, false, false, nil, false)
	case requirementBinding == "unavailable":
		s.requirementDiagnosticAt(CodeUnavailableField, "error", "unavailable_field", message, operand.Location, false, []string{id})
	}
	return id
}

// SPL2 fields inclusion closes over bindings proved before the command plus
// source names admitted by the caller. An unresolved source declaration keeps
// the wildcard incomplete, but it does not become a guessed output field.
func (s *spl2SemanticStage) closeSelectedFieldsWildcard(selectors []locatedOperand, priorFields map[fieldIdentityKey]bool, priorExpansionCount int) {
	if s.refinement == nil {
		return
	}
	patterns := []string{}
	exact := map[fieldIdentityKey]bool{}
	for _, selector := range selectors {
		if selector.Resolution == "wildcard" {
			patterns = append(patterns, selector.Name)
			continue
		}
		if key, ok := selector.fieldIdentity().privateKey(); ok {
			exact[key] = true
		}
	}
	if len(patterns) == 0 {
		return
	}
	pruned := map[string]bool{}
	for key, field := range s.env.fields {
		if priorFields[key] || exact[key] || !field.source || s.refinement.admission(field.Name) == SourceFieldAdmitted {
			continue
		}
		matched := false
		for _, pattern := range patterns {
			matched = matched || wildcardMatches(pattern, field.Name)
		}
		if !matched {
			continue
		}
		delete(s.env.fields, key)
		delete(s.env.identities, key)
		pruned[field.Name] = true
	}
	if len(pruned) == 0 {
		return
	}
	for i := priorExpansionCount; i < len(s.refinement.expansions); i++ {
		expansion := &s.refinement.expansions[i]
		matches := expansion.Matches[:0]
		for _, match := range expansion.Matches {
			if !pruned[match.Name] {
				matches = append(matches, match)
			}
		}
		expansion.Matches = matches
	}
	transitions := s.transitions[:0]
	for _, transition := range s.transitions {
		if transition.Operation != "project" || !pruned[transition.Output] {
			transitions = append(transitions, transition)
		}
	}
	s.transitions = transitions
	s.env.rewriteProject(s.env.fields)
}

// Independently parsed inputs survive rejected commands, but parser damage or
// a command-owned contract error cannot prove an output effect.
func (s *spl2SemanticStage) commandEffectSound(ctx antlr.ParserRuleContext) bool {
	if ctx == nil || !spl2IntactSyntax(ctx) {
		return false
	}
	if s.stage >= 0 && s.stage < len(s.result.Stages) && !s.result.Stages[s.stage].SemanticComplete {
		return false
	}
	location := s.parsed2.source.contextLocation(ctx)
	for _, d := range s.parsed2.diagnostics {
		if d.Location.Start.Offset >= location.Start.Offset && d.Location.Start.Offset < location.End.Offset {
			return false
		}
	}
	return true
}

func (s *spl2SemanticStage) joinPredicateIntentions(command *spl2.JoinCommandContext) []string {
	ids := []string{}
	if command == nil || command.SqlJoinPredicate() == nil {
		return ids
	}
	aliases := map[string]bool{}
	for _, option := range command.AllJoinOption() {
		if option.Identifier() == nil || option.LEFT() == nil && option.RIGHT() == nil {
			continue
		}
		alias := s.operand(option.Identifier())
		if alias.Sound {
			aliases[alias.Name] = true
		}
	}
	for _, equality := range command.SqlJoinPredicate().AllSqlJoinEquality() {
		for _, field := range equality.AllSqlJoinField() {
			if !s.parsed2.soundOperand(field) || len(field.AllAccessPart()) != 0 || len(field.AllIdentifier()) != 2 {
				continue
			}
			left, right := s.operand(field.Identifier(0)), s.operand(field.Identifier(1))
			if !left.Sound || !right.Sound || !aliases[left.Name] {
				continue
			}
			identity := pathFieldIdentity(left.Name, []string{right.Name})
			o := locatedOperand{Name: identity.PublicName, Identity: identity, Location: s.parsed2.source.contextLocation(field), Resolution: "exact", Sound: true}
			if id := s.structuralFieldReference(o); id != "" {
				ids = append(ids, id)
			}
		}
	}
	return ids
}

// A real target token cannot prove an effect from a damaged RHS or rejected
// constructor. Traverse the RHS first so independent original reads survive.
func (s *spl2SemanticStage) assignmentEffectSound(a spl2.IAssignmentContext) bool {
	if !s.parsed2.soundOperand(a) {
		return false
	}
	location := s.parsed2.source.contextLocation(a)
	for _, d := range s.parsed2.diagnostics {
		if d.Code == CodeSyntaxError && d.Category == "contract" && d.Location.Start.Offset >= location.Start.Offset && d.Location.Start.Offset < location.End.Offset {
			return false
		}
	}
	return true
}

func (s *spl2SemanticStage) damagedExpressionContinuation(ctx antlr.ParserRuleContext) bool {
	if ctx == nil {
		return true
	}
	isContinuation := func(token antlr.Token) bool {
		if token == nil {
			return false
		}
		switch token.GetTokenType() {
		case spl2.SPL2ParserSTAR, spl2.SPL2ParserSLASH, spl2.SPL2ParserMOD, spl2.SPL2ParserAND, spl2.SPL2ParserOR, spl2.SPL2ParserXOR, spl2.SPL2ParserNOT:
			return true
		}
		return false
	}
	if isContinuation(ctx.GetStop()) || isContinuation(s.nextCommandToken(ctx)) {
		return true
	}
	if ctx.GetStart() == nil || ctx.GetStart().GetTokenType() != spl2.SPL2ParserLPAREN {
		return false
	}
	location := s.parsed2.source.contextLocation(ctx)
	for _, diagnostic := range s.parsed2.diagnostics {
		if diagnostic.Code == CodeSyntaxError && diagnostic.Location.Start.Offset >= location.Start.Offset && diagnostic.Location.Start.Offset <= location.End.Offset {
			return true
		}
	}
	return !spl2IntactSyntax(ctx)
}

func (s *spl2SemanticStage) deferredAggregate(aggregate spl2.IAggregateContext) {
	if aggregate == nil || aggregate.Call() == nil {
		return
	}
	s.call(aggregate.Call(), true)
	if alias := aggregate.AggregateAlias(); alias != nil && spl2IntactSyntax(aggregate) {
		s.operandReference(s.operand(alias.Identifier()), "field", "output")
	}
}

// Input evidence precedes an unmodeled effect. Candidate destination references
// never install fields or transitions. Nil affected means unknown output names.
func (s *spl2SemanticStage) deferredEffects(ctx antlr.ParserRuleContext, affected map[string]bool, generating bool) {
	s.unsupported(ctx, "Command field effects are not yet modeled")
	for key, field := range s.env.fields {
		if affected == nil || affected[field.Name] {
			field.Conditional = true
			s.env.fields[key] = field
			s.env.requirements.markConditionalIdentity(field.identity)
		}
	}
	if generating {
		s.env.open = true
	}
}

func (s *spl2SemanticStage) markRejectedIdentityEffect(operand locatedOperand) {
	if !operand.Sound {
		return
	}
	identity := operand.fieldIdentity()
	key, exact := identity.privateKey()
	if !exact {
		return
	}
	field, known := s.env.fields[key]
	if !known {
		return
	}
	field.Conditional = true
	field.valueState = fieldValueUnknown
	s.env.fields[key] = field
	s.env.requirements.markConditionalIdentity(identity)
}

func (s *spl2SemanticStage) ownExistingUnsupportedEffect(id string) bool {
	if id == "" || s.env.requirements.trace == nil {
		return false
	}
	stageID := s.result.Stages[s.stage].ID
	trace := s.env.requirements.trace
	for index := len(trace.diagnostics) - 1; index >= 0; index-- {
		entry := &trace.diagnostics[index]
		if entry.incomplete && entry.diagnostic.StageID == stageID && entry.diagnostic.Code == CodeUnsupportedSemantics {
			entry.pendingReferenceIDs = uniqueIDs(entry.pendingReferenceIDs, []string{id})
			return true
		}
	}
	return false
}

func (s *spl2SemanticStage) rejectSelectedIdentityEffect(ctx antlr.ParserRuleContext, operand locatedOperand, id string) {
	s.markRejectedIdentityEffect(operand)
	if !s.ownExistingUnsupportedEffect(id) {
		pendingReferenceIDs := []string{}
		if id != "" {
			pendingReferenceIDs = append(pendingReferenceIDs, id)
		}
		s.unsupportedOwned(ctx, "Rejected command field effects are not modeled", pendingReferenceIDs)
	}
}

// Only the command-owned intact bare-index equality is a source selector.
// BY fields, aggregate operands and other expression positions remain fields.
func (s *spl2SemanticStage) metricsPredicate(tree antlr.Tree) {
	if p, ok := tree.(spl2.IPredicateContext); ok {
		if p.Comparison() != nil && (p.Comparison().ASSIGN() != nil || p.Comparison().EQ() != nil) && len(p.AllAdditive()) == 2 {
			left, right := spl2SingleAccess(p.Additive(0)), spl2SingleAccess(p.Additive(1))
			if left != nil && len(left.AllAccessPart()) == 0 && left.Primary().FieldName() != nil && left.Primary().FieldName().Identifier() != nil && left.Primary().FieldName().Identifier().INDEX() != nil && s.parsed2.soundOperand(left) {
				var value antlr.ParserRuleContext
				if spl2IntactSyntax(p) && right != nil && len(right.AllAccessPart()) == 0 {
					if literal := right.Primary().Literal(); literal != nil && literal.StringLiteral() != nil && len(literal.StringLiteral().AllExpression()) == 0 {
						value = literal.StringLiteral()
					}
					if field := right.Primary().FieldName(); field != nil && field.Identifier() != nil && field.Identifier().QuotedName() == nil {
						value = field.Identifier()
					}
				}
				if value != nil {
					o := s.operand(value)
					if o.Sound && plainCatalogComponent(o.Name) {
						s.dependency(value, "index")
						s.rewriteDependencyRole("metric_value")
						return
					}
					if o.Sound && strings.Contains(o.Name, "*") {
						o.Resolution = "wildcard"
						s.operandReference(o, "index", "read")
						s.unsupported(value, "Metrics index pattern membership is unproved")
						return
					}
				}
				s.expression(p.Additive(1))
				s.unsupported(p, "Metrics index selector identity is unproved")
				return
			}
		}
		if access := spl2SingleAccess(p); access != nil && len(access.AllAccessPart()) == 0 && access.Primary().Expression() != nil {
			s.metricsPredicate(access.Primary().Expression())
			return
		}
		s.expression(p)
		return
	}
	for _, child := range tree.GetChildren() {
		if _, ok := child.(antlr.ParserRuleContext); ok {
			s.metricsPredicate(child)
		}
	}
}

func (s *spl2SemanticStage) allnum(options []spl2.IAllnumOptionContext) bool {
	first := ""
	conditional := false
	for _, option := range options {
		value := option.BOOLEAN().GetText()
		if first != "" && first != value {
			s.unsupported(option, "Conflicting allnum option precedence is unproved")
		}
		if first == "" {
			first = value
		}
		conditional = conditional || value == "true"
	}
	return conditional
}
func spl2Groups(ctx spl2.IAggregateGroupContext) []spl2.IGroupFieldContext {
	if ctx == nil {
		return nil
	}
	return ctx.AllGroupField()
}

type spl2ExpressionGroup struct {
	Target                 locatedOperand
	InputReferenceIDs      []string
	Conditional            bool
	RequirementConditional bool
}

type spl2InstalledAggregateOutput struct {
	aggregateOutput
	OutputReferenceID string
	Conditional       bool
}

// Selected expression groups must finish installing every private identity
// before transitions are published. This mirrors the shared create transfer
// without emitting a transition that a later collision could invalidate.
func (s *spl2SemanticStage) installSelectedAggregateOutput(output aggregateOutput, stageID string) spl2InstalledAggregateOutput {
	id := s.operandReference(output.Target, "field", "output")
	if id == "" {
		return spl2InstalledAggregateOutput{}
	}
	s.rewriteBinding(id, "definition", output.InputReferenceIDs)
	origins := s.origins(output.InputReferenceIDs)
	s.result.References[len(s.result.References)-1].OriginReferenceIDs = copyIDs(origins)
	conditional := output.Conditional || !s.result.Stages[s.stage].SemanticComplete
	collision, owners := s.env.installIdentity(output.Target.fieldIdentity(), uniqueIDs([]string{id}, origins), conditional, false)
	if trace := s.env.requirements.trace; trace != nil {
		entry := trace.reference(id)
		entry.reference.Binding = "definition"
		entry.reference.OriginReferenceIDs = traceOrigins(trace, output.InputReferenceIDs)
		requirementConditional := output.RequirementConditional || s.env.requirements.stageIncomplete(stageID)
		s.env.requirements.installIdentity(output.Target.fieldIdentity(), uniqueIDs([]string{id}, entry.reference.OriginReferenceIDs), requirementConditional, false)
	}
	if collision {
		s.fieldIdentityCollision(output.Target.Name, output.Target.Location, owners)
	}
	return spl2InstalledAggregateOutput{aggregateOutput: output, OutputReferenceID: id, Conditional: conditional}
}

func (s *spl2SemanticStage) aggregates(calls []spl2.IAggregateContext, keys []spl2.IGroupFieldContext, selected spl2.ISelectedAggregateGroupContext, preserve, conditional bool) {
	outputs := []aggregateOutput{}
	groups := []locatedOperand{}
	expressionGroups := []spl2ExpressionGroup{}
	for _, a := range calls {
		value := s.call(a.Call(), true)
		target := locatedOperand{}
		if alias := a.AggregateAlias(); alias != nil {
			target = s.operand(alias.Identifier())
		} else if next := s.nextCommandToken(a); next != nil && (next.GetTokenType() == spl2.SPL2ParserAS || next.GetTokenType() == spl2.SPL2ParserAS_LOWER) {
			// A surviving AS token is explicit destination intent, even when its
			// missing identifier kept it outside the recovered aggregate context.
			s.unsupported(a, "Damaged explicit aggregate alias has no proved output")
		} else {
			call := a.Call()
			name := call.Identifier().GetText()
			args := call.Arguments()
			label := ""
			if name == "count" && args == nil {
				label = "count"
			} else if args != nil && len(args.AllNamedArgument()) == 0 && len(args.AllExpression()) == 1 {
				access := spl2SingleAccess(args.Expression(0))
				if access != nil && len(access.AllAccessPart()) == 0 && access.Primary().FieldName() != nil {
					field := access.Primary().FieldName().Identifier()
					if field != nil && ((name == "sum" && field.GetText() == "bytes") || (name == "dc" && field.GetText() == "action")) {
						label = name + "(" + field.GetText() + ")"
					}
				}
			}
			if label != "" {
				target = locatedOperand{Name: label, Location: s.parsed2.source.contextLocation(call), Resolution: "exact", Sound: true, rewrite: rewriteOwner{role: "implicit_output", location: s.parsed2.source.contextLocation(call), implicit: name}}
			} else {
				s.unsupported(call, "Implicit aggregate output label is unproved; use AS")
			}
		}
		if target.Sound {
			outputs = append(outputs, aggregateOutput{
				Target:                 target,
				InputReferenceIDs:      value.ids,
				Conditional:            conditional || !value.nonnull,
				RequirementConditional: conditional || !value.requirementNonnull,
			})
		}
	}
	for _, key := range keys {
		groups = append(groups, s.selector(key.Identifier()))
		if key.GroupSpan() != nil {
			s.unsupported(key.GroupSpan(), "Grouping span field effects are unmodeled")
		}
	}
	if selected != nil {
		for _, term := range selected.AllSelectedGroupTerm() {
			group := term.Expression()
			span := term.SelectedSpanGroup()
			if span != nil {
				group = span.Expression()
			}
			value := s.expressionWithRole(group, "group")
			target, exact := s.selectedGroupTarget(group)
			if span != nil && !exact {
				s.unsupportedOwned(span, "Grouping span requires one exact field path", value.ids)
				continue
			}
			if !exact {
				name := group.GetText()
				target = locatedOperand{Name: name, Identity: atomicFieldIdentity(name), Location: s.parsed2.source.contextLocation(group), Resolution: "exact", Sound: spl2IntactSyntax(group)}
			}
			if target.Sound {
				expressionGroups = append(expressionGroups, spl2ExpressionGroup{Target: target, InputReferenceIDs: value.ids, Conditional: !value.nonnull, RequirementConditional: !value.requirementNonnull})
			}
		}
	}
	if len(expressionGroups) > 0 {
		knownAmbiguity := map[string]bool{}
		for name, ambiguous := range s.env.ambiguous {
			knownAmbiguity[name] = ambiguous
		}
		s.applyAggregation(nil, nil, false)
		for _, group := range expressionGroups {
			origins := s.origins(group.InputReferenceIDs)
			collision, owners := s.env.installIdentity(group.Target.fieldIdentity(), origins, group.Conditional, false)
			if trace := s.env.requirements.trace; trace != nil {
				s.env.requirements.installIdentity(group.Target.fieldIdentity(), traceOrigins(trace, group.InputReferenceIDs), group.RequirementConditional, false)
			}
			if collision && !knownAmbiguity[group.Target.Name] {
				s.fieldIdentityCollision(group.Target.Name, group.Target.Location, owners)
				knownAmbiguity[group.Target.Name] = true
			}
		}
		stageID := s.result.Stages[s.stage].ID
		installedOutputs := make([]spl2InstalledAggregateOutput, 0, len(outputs))
		for _, output := range outputs {
			if installed := s.installSelectedAggregateOutput(output, stageID); installed.OutputReferenceID != "" {
				installedOutputs = append(installedOutputs, installed)
			}
		}
		for _, group := range expressionGroups {
			outputIdentity := transitionOutputIdentity(group.Target.fieldIdentity())
			s.appendTransition(Transition{Operation: "project", Output: group.Target.Name, OutputIdentity: outputIdentity, InputReferenceIDs: copyIDs(group.InputReferenceIDs), Conditional: group.Conditional})
		}
		for _, output := range installedOutputs {
			outputIdentity := transitionOutputIdentity(output.Target.fieldIdentity())
			s.appendTransition(Transition{Operation: "aggregate", Output: output.Target.Name, OutputIdentity: outputIdentity, InputReferenceIDs: copyIDs(output.InputReferenceIDs), OutputReferenceID: output.OutputReferenceID, Conditional: output.Conditional})
		}
		return
	}
	s.applyAggregation(outputs, groups, preserve)
}

func (s *spl2SemanticStage) selectedGroupTarget(tree antlr.Tree) (locatedOperand, bool) {
	access := spl2SQLFieldAccess(tree)
	if access == nil || access.Primary().FieldName() == nil || access.Primary().FieldName().Identifier() == nil {
		return locatedOperand{}, false
	}
	base := s.operand(access.Primary().FieldName().Identifier())
	if !base.Sound {
		return locatedOperand{}, false
	}
	if len(access.AllAccessPart()) == 0 {
		base.rewrite.role = "selector_atom"
		return base, true
	}
	segments := []string{base.Name}
	for _, part := range access.AllAccessPart() {
		if part.DOT() == nil || part.Identifier() == nil {
			return locatedOperand{}, false
		}
		segment := s.operand(part.Identifier())
		if !segment.Sound {
			return locatedOperand{}, false
		}
		segments = append(segments, segment.Name)
	}
	identity := pathFieldIdentity("", segments)
	return locatedOperand{Name: identity.PublicName, Identity: identity, Location: s.parsed2.source.contextLocation(access), Resolution: "exact", Sound: spl2IntactSyntax(access), UnresolvedSource: true, rewrite: s.rewriteNavigation(access)}, true
}
func (s *spl2SemanticStage) lookup(c *spl2.LookupCommandContext) {
	s.dependency(c.LookupDataset(), "lookup")
	ids := []string{}
	remoteKeys := map[string]bool{}
	for _, match := range c.AllLookupMatch() {
		remote := s.operand(match.LookupColumn())
		if remote.Sound {
			remoteKeys[remote.Name] = true
		}
		var ctx antlr.ParserRuleContext = match.LookupColumn()
		if match.LookupEventField() != nil {
			ctx = match.LookupEventField()
		}
		ids = append(ids, s.readIdentifier(ctx, "read"))
	}
	output := c.LookupOutputClause()
	if output == nil {
		s.unsupported(c, "Lookup without explicit outputs requires catalog field membership")
		for key, field := range s.env.fields {
			if !remoteKeys[field.Name] {
				field.Conditional = true
				s.env.fields[key] = field
				s.env.requirements.markConditionalIdentity(field.identity)
			}
		}
		return
	}
	for _, item := range output.AllLookupOutput() {
		var ctx antlr.ParserRuleContext = item.LookupColumn()
		if item.LookupEventField() != nil {
			ctx = item.LookupEventField()
		}
		s.applyLookupOutputs(ids, []lookupOutput{{Target: s.operand(ctx), PreserveExisting: output.OUTPUTNEW() != nil}})
	}
}
func (s *spl2SemanticStage) search(node antlr.Tree) {
	if atom, ok := node.(spl2.ISearchAtomContext); ok {
		if field := atom.Identifier(); field != nil {
			if spl2SearchModifierLabel(atom, field) {
				return
			}
			name := s.operand(field).Name
			source := name == "index" || name == "source" || name == "sourcetype"
			if source {
				for _, value := range atom.AllSearchValue() {
					o := s.operand(value)
					if value.SearchBareValue() != nil && strings.Contains(o.Name, "*") {
						o.Resolution = "wildcard"
					}
					if s.operandReference(o, name, "read") != "" {
						s.addDependency(o.Name, name)
					}
				}
			} else {
				s.readIdentifier(field, "read")
			}
		}
		if atom.SearchExpression() != nil {
			s.search(atom.SearchExpression())
		}
		return
	}
	for _, child := range node.GetChildren() {
		s.search(child)
	}
}

// A damaged selector cannot be fed to the shared projection kernel as an empty
// operand, nor can an unknown selector establish a closed output universe.
func (s *spl2SemanticStage) unprovedProjection(ctx antlr.ParserRuleContext, fields []locatedOperand) {
	for _, field := range fields {
		s.readAt(field, "read")
	}
	s.deferredEffects(ctx, nil, true)
}

func spl2SearchModifierLabel(atom spl2.ISearchAtomContext, field spl2.IIdentifierContext) bool {
	if atom.Comparison() == nil || atom.Comparison().ASSIGN() == nil || field.GetStart() != field.GetStop() {
		return false
	}
	switch field.GetStart().GetTokenType() {
	case spl2.SPL2ParserEARLIEST, spl2.SPL2ParserLATEST, spl2.SPL2ParserINDEX_EARLIEST, spl2.SPL2ParserINDEX_LATEST, spl2.SPL2ParserSTARTTIME, spl2.SPL2ParserENDTIME, spl2.SPL2ParserTIMEFORMAT:
		return true
	}
	return false
}

func (s *spl2SemanticStage) nextCommandToken(ctx antlr.ParserRuleContext) antlr.Token {
	stop := ctx.GetStop()
	if stop == nil {
		return nil
	}
	for i := stop.GetTokenIndex() + 1; i < s.parsed2.tokens.Size(); i++ {
		token := s.parsed2.tokens.Get(i)
		if token.GetChannel() != antlr.TokenDefaultChannel {
			continue
		}
		if token.GetTokenType() == spl2.SPL2ParserPIPE || token.GetTokenType() == spl2.SPL2ParserSEMI || token.GetTokenType() == antlr.TokenEOF {
			return nil
		}
		return token
	}
	return nil
}

// Right-side datasets are dependency intentions, not an instruction to install
// a joined array's row shape into the parent environment.
func (s *spl2SemanticStage) joinDatasetIntentions(from spl2.ISqlFromClauseContext) {
	for _, join := range from.AllSqlJoinClause() {
		if dataset := join.Dataset(); dataset != nil {
			s.datasetInputIntention(dataset)
		}
	}
}

func (s *spl2SemanticStage) recoveredInputs(ctx antlr.ParserRuleContext) []string {
	switch c := ctx.(type) {
	case *spl2.BinCommandContext:
		if len(c.AllIdentifier()) > 0 {
			input := s.operand(c.Identifier(0))
			id := s.readAt(input, "read")
			s.markRejectedIdentityEffect(input)
			if id != "" {
				return []string{id}
			}
		}
	case *spl2.MvexpandCommandContext:
		if c.Identifier() != nil {
			input := s.operand(c.Identifier())
			id := s.readAt(input, "read")
			s.markRejectedIdentityEffect(input)
			if id != "" {
				return []string{id}
			}
		}
	case *spl2.MetricsCommandContext:
		s.metricsInputs(c)
		s.deferredEffects(c, nil, true)
	case *spl2.TimechartCommandContext:
		s.timechartInputs(c)
		s.deferredEffects(c, nil, true)
	case *spl2.UnionCommandContext:
		s.unionDatasetIntentions(c)
	case *spl2.SearchCommandContext:
		if c.SearchExpression() != nil {
			s.rewritePredicate(c.SearchExpression(), "spl2", true, false)
			s.search(c.SearchExpression())
		}
	case *spl2.ImplicitSearchContext:
		if c.SearchExpression() != nil {
			s.rewritePredicate(c.SearchExpression(), "spl2", true, false)
			s.search(c.SearchExpression())
		}
	case *spl2.FromCommandContext:
		if from := c.SqlFromClause(); from != nil {
			s.sourceIntentions(from)
		}
	}
	return nil
}
func (s *spl2SemanticStage) sourceIntentions(from spl2.ISqlFromClauseContext) {
	if dataset := from.Dataset(); dataset != nil {
		s.exactDatasetSource(dataset)
	}
	s.joinDatasetIntentions(from)
}

func (s *spl2SemanticStage) metricsInputs(c *spl2.MetricsCommandContext) {
	if aggregates := c.MetricsAggregates(); aggregates != nil {
		for _, aggregate := range aggregates.AllAggregate() {
			s.deferredAggregate(aggregate)
		}
	}
	for _, option := range c.AllMetricsOption() {
		if option.Expression() != nil {
			s.rewritePredicate(option.Expression(), "spl2", false, true)
			s.metricsPredicate(option.Expression())
		}
		for _, group := range option.AllGroupField() {
			o := s.selector(group.Identifier())
			if o.Resolution == "wildcard" {
				if s.operandReference(o, "field", "group") != "" {
					s.result.References[len(s.result.References)-1].Binding = "indeterminate"
				}
			} else {
				s.readAt(o, "group")
			}
		}
		if c.TSTATS() != nil && option.QuotedName() != nil {
			o := s.operand(option.QuotedName())
			if o.Sound && o.Name != "" && !strings.ContainsAny(o.Name, "*\\:") {
				s.dependency(option.QuotedName(), "data_model")
			} else if o.Sound {
				s.rewriteUnprovedOperand(o, "data_model", "quoted_catalog_atom", "dynamic_identity")
			}
		}
	}
}

func (s *spl2SemanticStage) timechartInputs(c *spl2.TimechartCommandContext) {
	for _, option := range c.AllTimechartOption() {
		if option.Aggregate() != nil {
			s.deferredAggregate(option.Aggregate())
		}
	}
	if c.Aggregate() != nil {
		s.deferredAggregate(c.Aggregate())
	}
	for _, extra := range c.AllTimechartExtraAggregate() {
		s.deferredAggregate(extra.Aggregate())
	}
	if c.Expression() != nil {
		s.sqlUnprovedAggregateExpression(c.Expression())
	}
	if split := c.TimechartSplit(); split != nil {
		s.readIdentifier(split.Identifier(), "group")
	}
}

func (s *spl2SemanticStage) unionDatasetIntentions(c *spl2.UnionCommandContext) {
	for _, input := range c.AllUnionDataset() {
		if dataset := input.Dataset(); dataset != nil {
			s.datasetInputIntention(dataset)
		}
	}
}

func (s *spl2SemanticStage) exactDatasetSource(dataset spl2.IDatasetContext) bool {
	if dataset == nil {
		return false
	}
	// Lexical Dataset bindings precede external source discovery, including
	// dependency-only joined sources. Scalar locals never become inputs here.
	if s.program != nil && (s.program.resolveViewSource(s, dataset.DatasetParameter()) || s.program.resolveImportedDataset(s, dataset)) {
		return true
	}
	var owner antlr.ParserRuleContext
	name := ""
	form := ""
	switch {
	case dataset.Identifier() != nil:
		form = "identifier"
		owner = dataset.Identifier()
		name = s.operand(owner).Name
	case dataset.DottedDataset() != nil:
		form = "dotted"
		owner = dataset.DottedDataset()
		parts := []string{}
		identifiers := []spl2.IIdentifierContext{dataset.DottedDataset().Identifier()}
		identifiers = append(identifiers, dataset.DottedDataset().DatasetPath().AllIdentifier()...)
		for _, identifier := range identifiers {
			part := s.operand(identifier)
			if !part.Sound {
				return true
			}
			parts = append(parts, part.Name)
		}
		name = strings.Join(parts, ".")
	case dataset.DatasetParameter() != nil:
		form = "parameter"
		owner = dataset.DatasetParameter()
		if spl2IntactSyntax(owner) {
			name = owner.GetText()
		}
	case dataset.StaticDatasetDescriptor() != nil:
		form = "descriptor"
		owner = dataset.StaticDatasetDescriptor()
		canonical, ok := spl2CanonicalDatasetDescriptor(dataset.StaticDatasetDescriptor())
		if !ok {
			s.dynamicDatasetSource(owner)
			for _, expression := range dataset.StaticDatasetDescriptor().AllExpression() {
				s.expression(expression)
			}
			return true
		}
		name = canonical
	default:
		return false
	}
	if name == "" || !spl2IntactSyntax(owner) {
		return true
	}
	location := s.parsed2.source.contextLocation(owner)
	operand := locatedOperand{Name: name, Location: location, Resolution: "exact", Sound: true, rewrite: s.rewriteSPL2Owner(owner)}
	if id := s.operandReference(operand, "dataset", "read"); id != "" {
		kind := "explicit_dataset"
		if form == "parameter" {
			kind = "named_placeholder"
		}
		s.recordInput(kind, name, form, name, id, spl2DatasetAlias(dataset), location)
		s.addDependency(name, "dataset")
	}
	return true
}

func (s *spl2SemanticStage) dynamicDatasetSource(owner antlr.ParserRuleContext) {
	location := s.parsed2.source.contextLocation(owner)
	name := s.result.Document.Text[location.Start.Offset:location.End.Offset]
	operand := locatedOperand{Name: name, Location: location, Resolution: "dynamic", Sound: name != "", rewrite: s.rewriteSPL2Owner(owner)}
	id := s.operandReference(operand, "dataset", "read")
	alias := ""
	if dataset, ok := owner.GetParent().(spl2.IDatasetContext); ok {
		alias = spl2DatasetAlias(dataset)
	}
	s.recordInput("unresolved_source", name, "descriptor", name, id, alias, location)
	if id == "" {
		s.diagnosticAt(CodeDynamicReference, "warning", "unsupported_semantics", "Dynamic dataset descriptor identity is unresolved", location, true)
		return
	}
	ref := &s.result.References[len(s.result.References)-1]
	ref.Binding = "indeterminate"
	if trace := s.env.requirements.trace; trace != nil {
		trace.reference(id).reference.Binding = "indeterminate"
	}
	s.rewriteBinding(id, "indeterminate", nil)
	s.diagnosticAtOwned(CodeDynamicReference, "warning", "unsupported_semantics", "Dynamic dataset descriptor identity is unresolved", location, true, []string{id})
}

func spl2CanonicalDatasetDescriptor(descriptor spl2.IStaticDatasetDescriptorContext) (string, bool) {
	if descriptor == nil || !spl2IntactSyntax(descriptor) || descriptor.JsonStringLiteral() == nil || len(descriptor.AllExpression()) != 0 {
		return "", false
	}
	kind, ok := spl2StaticJSONString(descriptor.JsonStringLiteral())
	if !ok {
		return "", false
	}
	value := map[string]any{"kind": kind}
	if descriptor.DescriptorPropertiesKey() != nil {
		if descriptor.DescriptorProperties() == nil {
			return "", false
		}
		properties, ok := spl2StaticJSONObject(descriptor.DescriptorProperties().AllDescriptorProperty())
		if !ok {
			return "", false
		}
		value["properties"] = properties
	}
	encoded, err := json.Marshal(value)
	return string(encoded), err == nil
}

func spl2StaticJSONLiteral(literal spl2.IJsonLiteralContext) (any, bool) {
	switch {
	case literal == nil:
		return nil, false
	case literal.NUMBER() != nil:
		number := json.Number(literal.NUMBER().GetText())
		return number, true
	case literal.BOOLEAN() != nil:
		value, err := strconv.ParseBool(literal.BOOLEAN().GetText())
		return value, err == nil
	case literal.NULL() != nil:
		return nil, true
	case literal.JsonStringLiteral() != nil:
		return spl2StaticJSONString(literal.JsonStringLiteral())
	case literal.JsonArray() != nil:
		values := make([]any, 0, len(literal.JsonArray().AllJsonLiteral()))
		for _, child := range literal.JsonArray().AllJsonLiteral() {
			value, ok := spl2StaticJSONLiteral(child)
			if !ok {
				return nil, false
			}
			values = append(values, value)
		}
		return values, true
	case literal.JsonObject() != nil:
		return spl2StaticJSONObject(literal.JsonObject().AllDescriptorProperty())
	default:
		return nil, false
	}
}

func spl2StaticJSONObject(properties []spl2.IDescriptorPropertyContext) (map[string]any, bool) {
	out := map[string]any{}
	for _, property := range properties {
		if property == nil || property.JsonObjectKey() == nil {
			return nil, false
		}
		key, ok := spl2DecodeKey(property.JsonObjectKey().GetText())
		if !ok {
			return nil, false
		}
		if _, duplicate := out[key]; duplicate {
			return nil, false
		}
		value, ok := spl2StaticJSONLiteral(property.JsonLiteral())
		if !ok {
			return nil, false
		}
		out[key] = value
	}
	return out, true
}

func spl2StaticJSONString(literal spl2.IJsonStringLiteralContext) (string, bool) {
	if literal == nil || strings.Contains(literal.GetText(), "${") {
		return "", false
	}
	return spl2DecodeKey(literal.GetText())
}

// Aliases describe occurrences and never enter logical source identities.
func spl2DatasetAlias(dataset spl2.IDatasetContext) string {
	if dataset == nil {
		return ""
	}
	var alias spl2.ISourceAliasContext
	switch parent := dataset.GetParent().(type) {
	case spl2.ISqlFromClauseContext:
		alias = parent.SourceAlias()
	case spl2.ISqlJoinClauseContext:
		alias = parent.SourceAlias()
	}
	if alias == nil || alias.Identifier() == nil || !spl2IntactSyntax(alias) {
		return ""
	}
	name, ok := spl2DecodeKey(alias.Identifier().GetText())
	if !ok {
		return ""
	}
	return name
}

func (s *spl2SemanticStage) datasetInputIntention(dataset spl2.IDatasetContext) {
	source := *s
	semantic := *s.semanticStage
	semantic.env = s.env.clone()
	source.semanticStage = &semantic
	source.exactDatasetSource(dataset)
	s.env.inputs = mergeInputFacts(s.env.inputs, source.env.inputs)
	s.env.requirements.inputs = mergeInputFacts(s.env.requirements.inputs, source.env.requirements.inputs)
}
