// Code generated from grammar/SPL2Parser.g4 by ANTLR 4.13.2. DO NOT EDIT.

package spl2 // SPL2Parser
import "github.com/antlr4-go/antlr/v4"

type BaseSPL2ParserVisitor struct {
	*antlr.BaseParseTreeVisitor
}

func (v *BaseSPL2ParserVisitor) VisitQuery(ctx *QueryContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitPipeline(ctx *PipelineContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitStart(ctx *StartContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitCommand(ctx *CommandContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitFromCommand(ctx *FromCommandContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitSelectCommand(ctx *SelectCommandContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitProjection(ctx *ProjectionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitDataset(ctx *DatasetContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitGenerator(ctx *GeneratorContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitEvalCommand(ctx *EvalCommandContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitAssignment(ctx *AssignmentContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitWhereCommand(ctx *WhereCommandContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitFieldsCommand(ctx *FieldsCommandContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitFieldSelection(ctx *FieldSelectionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitFieldSelector(ctx *FieldSelectorContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitTableCommand(ctx *TableCommandContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitTableField(ctx *TableFieldContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitRenameCommand(ctx *RenameCommandContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitRenamePair(ctx *RenamePairContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitRenameSource(ctx *RenameSourceContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitRenameTarget(ctx *RenameTargetContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitAliasKeyword(ctx *AliasKeywordContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitStatsCommand(ctx *StatsCommandContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitStatsOption(ctx *StatsOptionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitAllnumOption(ctx *AllnumOptionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitDelimOption(ctx *DelimOptionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitPartitionsOption(ctx *PartitionsOptionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitAggregate(ctx *AggregateContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitAggregateAlias(ctx *AggregateAliasContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitAggregateGroup(ctx *AggregateGroupContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitGroupField(ctx *GroupFieldContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitGroupSpan(ctx *GroupSpanContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitTimeSpan(ctx *TimeSpanContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitEventstatsCommand(ctx *EventstatsCommandContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitStreamstatsCommand(ctx *StreamstatsCommandContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitStreamGroup(ctx *StreamGroupContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitCurrentOption(ctx *CurrentOptionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitWindowOption(ctx *WindowOptionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitResetClause(ctx *ResetClauseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitResetBefore(ctx *ResetBeforeContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitResetAfter(ctx *ResetAfterContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitResetOnchange(ctx *ResetOnchangeContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitStreamPostLayout(ctx *StreamPostLayoutContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitLookupCommand(ctx *LookupCommandContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitLookupDataset(ctx *LookupDatasetContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitLookupMatch(ctx *LookupMatchContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitLookupOutputClause(ctx *LookupOutputClauseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitLookupOutput(ctx *LookupOutputContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitLookupColumn(ctx *LookupColumnContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitLookupEventField(ctx *LookupEventFieldContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitSortCommand(ctx *SortCommandContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitSortTerm(ctx *SortTermContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitSortWrapper(ctx *SortWrapperContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitIntegerValue(ctx *IntegerValueContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitDedupCommand(ctx *DedupCommandContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitKeepemptyOption(ctx *KeepemptyOptionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitConsecutiveOption(ctx *ConsecutiveOptionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitDedupField(ctx *DedupFieldContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitHeadCommand(ctx *HeadCommandContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitKeeplastOption(ctx *KeeplastOptionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitHeadWhile(ctx *HeadWhileContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitHeadPostLayout(ctx *HeadPostLayoutContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitReverseCommand(ctx *ReverseCommandContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitUnknownOption(ctx *UnknownOptionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitRexCommand(ctx *RexCommandContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitRexOption(ctx *RexOptionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitEmbeddedCommand(ctx *EmbeddedCommandContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitEmbeddedText(ctx *EmbeddedTextContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitModuleSuffix(ctx *ModuleSuffixContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitModuleDeclaration(ctx *ModuleDeclarationContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitSearchCommand(ctx *SearchCommandContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitImplicitSearch(ctx *ImplicitSearchContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitSearchExpression(ctx *SearchExpressionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitSearchXor(ctx *SearchXorContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitSearchAnd(ctx *SearchAndContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitSearchOr(ctx *SearchOrContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitSearchNot(ctx *SearchNotContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitSearchAtom(ctx *SearchAtomContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitSearchValue(ctx *SearchValueContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitSearchBareValue(ctx *SearchBareValueContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitSearchDirective(ctx *SearchDirectiveContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitSearchTimeModifier(ctx *SearchTimeModifierContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitTimeModifierKey(ctx *TimeModifierKeyContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitTimeModifierValue(ctx *TimeModifierValueContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitRelativeTime(ctx *RelativeTimeContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitExpression(ctx *ExpressionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitXorExpression(ctx *XorExpressionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitOrExpression(ctx *OrExpressionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitAndExpression(ctx *AndExpressionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitNotExpression(ctx *NotExpressionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitPredicate(ctx *PredicateContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitLogicalAnd(ctx *LogicalAndContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitLogicalOr(ctx *LogicalOrContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitLogicalXor(ctx *LogicalXorContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitLogicalNot(ctx *LogicalNotContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitBetweenOperator(ctx *BetweenOperatorContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitBetweenConjunction(ctx *BetweenConjunctionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitComparison(ctx *ComparisonContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitAdditive(ctx *AdditiveContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitMultiplicative(ctx *MultiplicativeContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitUnary(ctx *UnaryContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitAccess(ctx *AccessContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitAccessPart(ctx *AccessPartContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitPrimary(ctx *PrimaryContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitCall(ctx *CallContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitArguments(ctx *ArgumentsContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitNamedArgument(ctx *NamedArgumentContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitLiteral(ctx *LiteralContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitStringLiteral(ctx *StringLiteralContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitQuotedName(ctx *QuotedNameContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitFieldTemplate(ctx *FieldTemplateContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitFieldName(ctx *FieldNameContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitIdentifier(ctx *IdentifierContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitPipelineKeyword(ctx *PipelineKeywordContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitArray(ctx *ArrayContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitObject(ctx *ObjectContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitObjectEntry(ctx *ObjectEntryContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitObjectKey(ctx *ObjectKeyContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitSearchLiteral(ctx *SearchLiteralContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitLambdaExpression(ctx *LambdaExpressionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitLambdaParameter(ctx *LambdaParameterContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitLambdaBlock(ctx *LambdaBlockContext) interface{} {
	return v.VisitChildren(ctx)
}
