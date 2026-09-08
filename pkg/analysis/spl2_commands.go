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
		s.search(c.SearchExpression())
	case *spl2.ImplicitSearchContext:
		s.search(c.SearchExpression())
	case *spl2.WhereCommandContext:
		s.expression(c.Expression())
	case *spl2.EvalCommandContext:
		for _, a := range c.AllAssignment() {
			value := s.expression(a.Expression())
			if a.FieldName().Identifier() == nil {
				s.expression(a.FieldName())
				s.unsupported(a.FieldName(), "Computed assignment target is unresolved")
				continue
			}
			s.applyAssignment(s.operand(a.FieldName().Identifier()), value.ids, !value.nonnull, value.exactNull)
		}
	case *spl2.FieldsCommandContext:
		selection := c.FieldSelection()
		fields := []locatedOperand{}
		for _, f := range selection.AllFieldSelector() {
			fields = append(fields, s.selector(f.Identifier()))
		}
		mode := "include"
		if selection.MINUS() != nil {
			mode = "exclude"
		}
		s.applyProjection(fields, mode, true)
	case *spl2.TableCommandContext:
		fields := []locatedOperand{}
		for _, f := range c.AllTableField() {
			fields = append(fields, s.selector(f))
		}
		s.applyProjection(fields, "table", false)
	case *spl2.RenameCommandContext:
		pairs := []renameOperands{}
		for _, pair := range c.AllRenamePair() {
			pairs = append(pairs, renameOperands{s.selector(pair.RenameSource()), s.selector(pair.RenameTarget())})
		}
		s.applyRename(pairs)
	case *spl2.StatsCommandContext:
		conditional := false
		for _, o := range c.AllStatsOption() {
			if a := o.AllnumOption(); a != nil && a.BOOLEAN().GetText() == "true" {
				conditional = true
			}
			if o.UnknownOption() != nil {
				s.unsupported(o, "Unmodeled stats option")
			}
		}
		s.aggregates(c.AllAggregate(), spl2Groups(c.AggregateGroup()), false, conditional)
	case *spl2.EventstatsCommandContext:
		conditional := false
		for _, o := range c.AllAllnumOption() {
			conditional = conditional || o.BOOLEAN().GetText() == "true"
		}
		s.aggregates(c.AllAggregate(), spl2Groups(c.AggregateGroup()), true, conditional)
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
	default:
		s.unsupported(ctx, "Command field effects are not yet modeled")
	}
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
				target = locatedOperand{Name: label, Location: s.parsed2.source.contextLocation(call), Resolution: "exact", Sound: true}
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
	for _, match := range c.AllLookupMatch() {
		var ctx antlr.ParserRuleContext = match.LookupColumn()
		if match.LookupEventField() != nil {
			ctx = match.LookupEventField()
		}
		ids = append(ids, s.readIdentifier(ctx, "read"))
	}
	output := c.LookupOutputClause()
	if output == nil {
		s.unsupported(c, "Lookup without explicit outputs requires catalog field membership")
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
