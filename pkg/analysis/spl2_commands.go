package analysis

import (
	"github.com/antlr4-go/antlr/v4"
	"github.com/delgado-jacob/spl-toolkit/parser/spl2"
	"strings"
)

func (s *spl2SemanticStage) command(ctx antlr.ParserRuleContext) {
	switch c := ctx.(type) {
	case *spl2.FromCommandContext:
		s.applySource()
		clear(s.aliases)
		source := c.SqlFromClause()
		s.dataset(source.Dataset())
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
	case *spl2.WhereCommandContext:
		s.rewritePredicate(c.Expression(), "spl2", false, false)
		s.expression(c.Expression())
	case *spl2.EvalCommandContext:
		for _, a := range c.AllAssignment() {
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
	case *spl2.FieldsCommandContext:
		selection := c.FieldSelection()
		fields := []locatedOperand{}
		unproved := false
		for _, f := range selection.AllFieldSelector() {
			o := s.selector(f.Identifier())
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
			s.applyProjection(fields, mode, true)
		}
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
	case *spl2.StatsCommandContext:
		allnum := []spl2.IAllnumOptionContext{}
		for _, o := range c.AllStatsOption() {
			if a := o.AllnumOption(); a != nil {
				allnum = append(allnum, a)
			}
			if o.UnknownOption() != nil {
				s.unsupported(o, "Unmodeled stats option")
			}
		}
		s.aggregates(c.AllAggregate(), spl2Groups(c.AggregateGroup()), false, s.allnum(allnum))
	case *spl2.EventstatsCommandContext:
		s.aggregates(c.AllAggregate(), spl2Groups(c.AggregateGroup()), true, s.allnum(c.AllAllnumOption()))
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
		s.aggregates(c.AllAggregate(), groups, true, conditional)
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
		s.joinPredicateIntentions(c.SqlJoinPredicate())
		s.unsupported(c, "Join output merge and qualified input binding are unproved")
	case *spl2.BinCommandContext:
		input := s.operand(c.Identifier(0))
		s.readAt(input, "read")
		target := input
		if len(c.AllIdentifier()) > 1 {
			target = s.operand(c.Identifier(1))
			s.operandReference(target, "field", "output")
		}
		s.deferredEffects(c, map[string]bool{target.Name: target.Sound}, false)
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
	case *spl2.MvexpandCommandContext:
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

func (s *spl2SemanticStage) joinPredicateIntentions(predicate spl2.ISqlJoinPredicateContext) {
	if predicate == nil {
		return
	}
	for _, equality := range predicate.AllSqlJoinEquality() {
		for _, field := range equality.AllSqlJoinField() {
			if !s.parsed2.soundOperand(field) || len(field.AllAccessPart()) != 0 || len(field.AllIdentifier()) != 2 {
				continue
			}
			left, right := s.operand(field.Identifier(0)), s.operand(field.Identifier(1))
			if !left.Sound || !right.Sound {
				continue
			}
			o := locatedOperand{Name: left.Name + "." + right.Name, Location: s.parsed2.source.contextLocation(field), Resolution: "exact", Sound: true}
			if s.operandReference(o, "field", "read") != "" {
				s.result.References[len(s.result.References)-1].Binding = "indeterminate"
			}
		}
	}
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
	for name, field := range s.env.fields {
		if affected == nil || affected[name] {
			field.Conditional = true
			s.env.fields[name] = field
			s.env.requirements.markConditional(name)
		}
	}
	if generating {
		s.env.open = true
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
func (s *spl2SemanticStage) aggregates(calls []spl2.IAggregateContext, keys []spl2.IGroupFieldContext, preserve, conditional bool) {
	outputs := []aggregateOutput{}
	groups := []locatedOperand{}
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
			outputs = append(outputs, aggregateOutput{Target: target, InputReferenceIDs: value.ids, Conditional: conditional || !value.nonnull})
		}
	}
	for _, key := range keys {
		groups = append(groups, s.selector(key.Identifier()))
		if key.GroupSpan() != nil {
			s.unsupported(key.GroupSpan(), "Grouping span field effects are unmodeled")
		}
	}
	s.applyAggregation(outputs, groups, preserve)
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
		for name, field := range s.env.fields {
			if !remoteKeys[name] {
				field.Conditional = true
				s.env.fields[name] = field
				s.env.requirements.markConditional(name)
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
		if dataset := join.Dataset(); dataset != nil && dataset.Identifier() != nil {
			s.dependency(dataset.Identifier(), "dataset")
		}
	}
}

func (s *spl2SemanticStage) recoveredInputs(ctx antlr.ParserRuleContext) {
	switch c := ctx.(type) {
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
}
func (s *spl2SemanticStage) sourceIntentions(from spl2.ISqlFromClauseContext) {
	if dataset := from.Dataset(); dataset != nil && dataset.Identifier() != nil {
		s.dependency(dataset.Identifier(), "dataset")
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
		if dataset := input.Dataset(); dataset != nil && dataset.Identifier() != nil {
			s.dependency(dataset.Identifier(), "dataset")
		}
	}
}
