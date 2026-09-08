// Code generated from grammar/SPL2Parser.g4 by ANTLR 4.13.2. DO NOT EDIT.

package spl2 // SPL2Parser
import "github.com/antlr4-go/antlr/v4"

import "strings"

// ANTLR also emits the header into the visitor interface file.
var _ = strings.EqualFold

// A complete Visitor for a parse tree produced by SPL2Parser.
type SPL2ParserVisitor interface {
	antlr.ParseTreeVisitor

	// Visit a parse tree produced by SPL2Parser#query.
	VisitQuery(ctx *QueryContext) interface{}

	// Visit a parse tree produced by SPL2Parser#pipeline.
	VisitPipeline(ctx *PipelineContext) interface{}

	// Visit a parse tree produced by SPL2Parser#start.
	VisitStart(ctx *StartContext) interface{}

	// Visit a parse tree produced by SPL2Parser#command.
	VisitCommand(ctx *CommandContext) interface{}

	// Visit a parse tree produced by SPL2Parser#fromCommand.
	VisitFromCommand(ctx *FromCommandContext) interface{}

	// Visit a parse tree produced by SPL2Parser#selectCommand.
	VisitSelectCommand(ctx *SelectCommandContext) interface{}

	// Visit a parse tree produced by SPL2Parser#sqlFromClause.
	VisitSqlFromClause(ctx *SqlFromClauseContext) interface{}

	// Visit a parse tree produced by SPL2Parser#sourceAlias.
	VisitSourceAlias(ctx *SourceAliasContext) interface{}

	// Visit a parse tree produced by SPL2Parser#sqlJoinClause.
	VisitSqlJoinClause(ctx *SqlJoinClauseContext) interface{}

	// Visit a parse tree produced by SPL2Parser#sqlJoinPredicate.
	VisitSqlJoinPredicate(ctx *SqlJoinPredicateContext) interface{}

	// Visit a parse tree produced by SPL2Parser#sqlJoinEquality.
	VisitSqlJoinEquality(ctx *SqlJoinEqualityContext) interface{}

	// Visit a parse tree produced by SPL2Parser#sqlJoinField.
	VisitSqlJoinField(ctx *SqlJoinFieldContext) interface{}

	// Visit a parse tree produced by SPL2Parser#sqlSelectClause.
	VisitSqlSelectClause(ctx *SqlSelectClauseContext) interface{}

	// Visit a parse tree produced by SPL2Parser#projection.
	VisitProjection(ctx *ProjectionContext) interface{}

	// Visit a parse tree produced by SPL2Parser#projectionAlias.
	VisitProjectionAlias(ctx *ProjectionAliasContext) interface{}

	// Visit a parse tree produced by SPL2Parser#sqlWhereClause.
	VisitSqlWhereClause(ctx *SqlWhereClauseContext) interface{}

	// Visit a parse tree produced by SPL2Parser#sqlHavingClause.
	VisitSqlHavingClause(ctx *SqlHavingClauseContext) interface{}

	// Visit a parse tree produced by SPL2Parser#sqlPredicate.
	VisitSqlPredicate(ctx *SqlPredicateContext) interface{}

	// Visit a parse tree produced by SPL2Parser#sqlGroupClause.
	VisitSqlGroupClause(ctx *SqlGroupClauseContext) interface{}

	// Visit a parse tree produced by SPL2Parser#sqlBy.
	VisitSqlBy(ctx *SqlByContext) interface{}

	// Visit a parse tree produced by SPL2Parser#sqlGroupKey.
	VisitSqlGroupKey(ctx *SqlGroupKeyContext) interface{}

	// Visit a parse tree produced by SPL2Parser#sqlSpanCall.
	VisitSqlSpanCall(ctx *SqlSpanCallContext) interface{}

	// Visit a parse tree produced by SPL2Parser#sqlSpanAssignment.
	VisitSqlSpanAssignment(ctx *SqlSpanAssignmentContext) interface{}

	// Visit a parse tree produced by SPL2Parser#sqlUnparenthesizedSpan.
	VisitSqlUnparenthesizedSpan(ctx *SqlUnparenthesizedSpanContext) interface{}

	// Visit a parse tree produced by SPL2Parser#sqlOrderClause.
	VisitSqlOrderClause(ctx *SqlOrderClauseContext) interface{}

	// Visit a parse tree produced by SPL2Parser#sqlOrderTerm.
	VisitSqlOrderTerm(ctx *SqlOrderTermContext) interface{}

	// Visit a parse tree produced by SPL2Parser#sqlDirection.
	VisitSqlDirection(ctx *SqlDirectionContext) interface{}

	// Visit a parse tree produced by SPL2Parser#sqlLimitClause.
	VisitSqlLimitClause(ctx *SqlLimitClauseContext) interface{}

	// Visit a parse tree produced by SPL2Parser#sqlOffsetClause.
	VisitSqlOffsetClause(ctx *SqlOffsetClauseContext) interface{}

	// Visit a parse tree produced by SPL2Parser#existsPredicate.
	VisitExistsPredicate(ctx *ExistsPredicateContext) interface{}

	// Visit a parse tree produced by SPL2Parser#dataset.
	VisitDataset(ctx *DatasetContext) interface{}

	// Visit a parse tree produced by SPL2Parser#generator.
	VisitGenerator(ctx *GeneratorContext) interface{}

	// Visit a parse tree produced by SPL2Parser#evalCommand.
	VisitEvalCommand(ctx *EvalCommandContext) interface{}

	// Visit a parse tree produced by SPL2Parser#assignment.
	VisitAssignment(ctx *AssignmentContext) interface{}

	// Visit a parse tree produced by SPL2Parser#whereCommand.
	VisitWhereCommand(ctx *WhereCommandContext) interface{}

	// Visit a parse tree produced by SPL2Parser#fieldsCommand.
	VisitFieldsCommand(ctx *FieldsCommandContext) interface{}

	// Visit a parse tree produced by SPL2Parser#fieldSelection.
	VisitFieldSelection(ctx *FieldSelectionContext) interface{}

	// Visit a parse tree produced by SPL2Parser#fieldSelector.
	VisitFieldSelector(ctx *FieldSelectorContext) interface{}

	// Visit a parse tree produced by SPL2Parser#tableCommand.
	VisitTableCommand(ctx *TableCommandContext) interface{}

	// Visit a parse tree produced by SPL2Parser#tableField.
	VisitTableField(ctx *TableFieldContext) interface{}

	// Visit a parse tree produced by SPL2Parser#renameCommand.
	VisitRenameCommand(ctx *RenameCommandContext) interface{}

	// Visit a parse tree produced by SPL2Parser#renamePair.
	VisitRenamePair(ctx *RenamePairContext) interface{}

	// Visit a parse tree produced by SPL2Parser#renameSource.
	VisitRenameSource(ctx *RenameSourceContext) interface{}

	// Visit a parse tree produced by SPL2Parser#renameTarget.
	VisitRenameTarget(ctx *RenameTargetContext) interface{}

	// Visit a parse tree produced by SPL2Parser#aliasKeyword.
	VisitAliasKeyword(ctx *AliasKeywordContext) interface{}

	// Visit a parse tree produced by SPL2Parser#statsCommand.
	VisitStatsCommand(ctx *StatsCommandContext) interface{}

	// Visit a parse tree produced by SPL2Parser#statsOption.
	VisitStatsOption(ctx *StatsOptionContext) interface{}

	// Visit a parse tree produced by SPL2Parser#allnumOption.
	VisitAllnumOption(ctx *AllnumOptionContext) interface{}

	// Visit a parse tree produced by SPL2Parser#delimOption.
	VisitDelimOption(ctx *DelimOptionContext) interface{}

	// Visit a parse tree produced by SPL2Parser#partitionsOption.
	VisitPartitionsOption(ctx *PartitionsOptionContext) interface{}

	// Visit a parse tree produced by SPL2Parser#aggregate.
	VisitAggregate(ctx *AggregateContext) interface{}

	// Visit a parse tree produced by SPL2Parser#aggregateAlias.
	VisitAggregateAlias(ctx *AggregateAliasContext) interface{}

	// Visit a parse tree produced by SPL2Parser#aggregateGroup.
	VisitAggregateGroup(ctx *AggregateGroupContext) interface{}

	// Visit a parse tree produced by SPL2Parser#groupField.
	VisitGroupField(ctx *GroupFieldContext) interface{}

	// Visit a parse tree produced by SPL2Parser#groupSpan.
	VisitGroupSpan(ctx *GroupSpanContext) interface{}

	// Visit a parse tree produced by SPL2Parser#timeSpan.
	VisitTimeSpan(ctx *TimeSpanContext) interface{}

	// Visit a parse tree produced by SPL2Parser#eventstatsCommand.
	VisitEventstatsCommand(ctx *EventstatsCommandContext) interface{}

	// Visit a parse tree produced by SPL2Parser#streamstatsCommand.
	VisitStreamstatsCommand(ctx *StreamstatsCommandContext) interface{}

	// Visit a parse tree produced by SPL2Parser#streamGroup.
	VisitStreamGroup(ctx *StreamGroupContext) interface{}

	// Visit a parse tree produced by SPL2Parser#currentOption.
	VisitCurrentOption(ctx *CurrentOptionContext) interface{}

	// Visit a parse tree produced by SPL2Parser#windowOption.
	VisitWindowOption(ctx *WindowOptionContext) interface{}

	// Visit a parse tree produced by SPL2Parser#resetClause.
	VisitResetClause(ctx *ResetClauseContext) interface{}

	// Visit a parse tree produced by SPL2Parser#resetBefore.
	VisitResetBefore(ctx *ResetBeforeContext) interface{}

	// Visit a parse tree produced by SPL2Parser#resetAfter.
	VisitResetAfter(ctx *ResetAfterContext) interface{}

	// Visit a parse tree produced by SPL2Parser#resetOnchange.
	VisitResetOnchange(ctx *ResetOnchangeContext) interface{}

	// Visit a parse tree produced by SPL2Parser#streamPostLayout.
	VisitStreamPostLayout(ctx *StreamPostLayoutContext) interface{}

	// Visit a parse tree produced by SPL2Parser#lookupCommand.
	VisitLookupCommand(ctx *LookupCommandContext) interface{}

	// Visit a parse tree produced by SPL2Parser#lookupDataset.
	VisitLookupDataset(ctx *LookupDatasetContext) interface{}

	// Visit a parse tree produced by SPL2Parser#lookupMatch.
	VisitLookupMatch(ctx *LookupMatchContext) interface{}

	// Visit a parse tree produced by SPL2Parser#lookupOutputClause.
	VisitLookupOutputClause(ctx *LookupOutputClauseContext) interface{}

	// Visit a parse tree produced by SPL2Parser#lookupOutput.
	VisitLookupOutput(ctx *LookupOutputContext) interface{}

	// Visit a parse tree produced by SPL2Parser#lookupColumn.
	VisitLookupColumn(ctx *LookupColumnContext) interface{}

	// Visit a parse tree produced by SPL2Parser#lookupEventField.
	VisitLookupEventField(ctx *LookupEventFieldContext) interface{}

	// Visit a parse tree produced by SPL2Parser#sortCommand.
	VisitSortCommand(ctx *SortCommandContext) interface{}

	// Visit a parse tree produced by SPL2Parser#sortTerm.
	VisitSortTerm(ctx *SortTermContext) interface{}

	// Visit a parse tree produced by SPL2Parser#sortWrapper.
	VisitSortWrapper(ctx *SortWrapperContext) interface{}

	// Visit a parse tree produced by SPL2Parser#integerValue.
	VisitIntegerValue(ctx *IntegerValueContext) interface{}

	// Visit a parse tree produced by SPL2Parser#dedupCommand.
	VisitDedupCommand(ctx *DedupCommandContext) interface{}

	// Visit a parse tree produced by SPL2Parser#keepemptyOption.
	VisitKeepemptyOption(ctx *KeepemptyOptionContext) interface{}

	// Visit a parse tree produced by SPL2Parser#consecutiveOption.
	VisitConsecutiveOption(ctx *ConsecutiveOptionContext) interface{}

	// Visit a parse tree produced by SPL2Parser#dedupField.
	VisitDedupField(ctx *DedupFieldContext) interface{}

	// Visit a parse tree produced by SPL2Parser#headCommand.
	VisitHeadCommand(ctx *HeadCommandContext) interface{}

	// Visit a parse tree produced by SPL2Parser#keeplastOption.
	VisitKeeplastOption(ctx *KeeplastOptionContext) interface{}

	// Visit a parse tree produced by SPL2Parser#headWhile.
	VisitHeadWhile(ctx *HeadWhileContext) interface{}

	// Visit a parse tree produced by SPL2Parser#headPostLayout.
	VisitHeadPostLayout(ctx *HeadPostLayoutContext) interface{}

	// Visit a parse tree produced by SPL2Parser#reverseCommand.
	VisitReverseCommand(ctx *ReverseCommandContext) interface{}

	// Visit a parse tree produced by SPL2Parser#unknownOption.
	VisitUnknownOption(ctx *UnknownOptionContext) interface{}

	// Visit a parse tree produced by SPL2Parser#unknownOptionName.
	VisitUnknownOptionName(ctx *UnknownOptionNameContext) interface{}

	// Visit a parse tree produced by SPL2Parser#independentSearch.
	VisitIndependentSearch(ctx *IndependentSearchContext) interface{}

	// Visit a parse tree produced by SPL2Parser#inheritedSubpipe.
	VisitInheritedSubpipe(ctx *InheritedSubpipeContext) interface{}

	// Visit a parse tree produced by SPL2Parser#joinCommand.
	VisitJoinCommand(ctx *JoinCommandContext) interface{}

	// Visit a parse tree produced by SPL2Parser#joinOption.
	VisitJoinOption(ctx *JoinOptionContext) interface{}

	// Visit a parse tree produced by SPL2Parser#joinType.
	VisitJoinType(ctx *JoinTypeContext) interface{}

	// Visit a parse tree produced by SPL2Parser#appendCommand.
	VisitAppendCommand(ctx *AppendCommandContext) interface{}

	// Visit a parse tree produced by SPL2Parser#appendpipeCommand.
	VisitAppendpipeCommand(ctx *AppendpipeCommandContext) interface{}

	// Visit a parse tree produced by SPL2Parser#appendcolsCommand.
	VisitAppendcolsCommand(ctx *AppendcolsCommandContext) interface{}

	// Visit a parse tree produced by SPL2Parser#unionCommand.
	VisitUnionCommand(ctx *UnionCommandContext) interface{}

	// Visit a parse tree produced by SPL2Parser#unionDataset.
	VisitUnionDataset(ctx *UnionDatasetContext) interface{}

	// Visit a parse tree produced by SPL2Parser#ifCommand.
	VisitIfCommand(ctx *IfCommandContext) interface{}

	// Visit a parse tree produced by SPL2Parser#binCommand.
	VisitBinCommand(ctx *BinCommandContext) interface{}

	// Visit a parse tree produced by SPL2Parser#binOption.
	VisitBinOption(ctx *BinOptionContext) interface{}

	// Visit a parse tree produced by SPL2Parser#alignmentOption.
	VisitAlignmentOption(ctx *AlignmentOptionContext) interface{}

	// Visit a parse tree produced by SPL2Parser#binSpan.
	VisitBinSpan(ctx *BinSpanContext) interface{}

	// Visit a parse tree produced by SPL2Parser#logarithmicSpan.
	VisitLogarithmicSpan(ctx *LogarithmicSpanContext) interface{}

	// Visit a parse tree produced by SPL2Parser#extendedOption.
	VisitExtendedOption(ctx *ExtendedOptionContext) interface{}

	// Visit a parse tree produced by SPL2Parser#signedNumber.
	VisitSignedNumber(ctx *SignedNumberContext) interface{}

	// Visit a parse tree produced by SPL2Parser#rexCommand.
	VisitRexCommand(ctx *RexCommandContext) interface{}

	// Visit a parse tree produced by SPL2Parser#rexOption.
	VisitRexOption(ctx *RexOptionContext) interface{}

	// Visit a parse tree produced by SPL2Parser#spathCommand.
	VisitSpathCommand(ctx *SpathCommandContext) interface{}

	// Visit a parse tree produced by SPL2Parser#spathOption.
	VisitSpathOption(ctx *SpathOptionContext) interface{}

	// Visit a parse tree produced by SPL2Parser#loadjobCommand.
	VisitLoadjobCommand(ctx *LoadjobCommandContext) interface{}

	// Visit a parse tree produced by SPL2Parser#metricsCommand.
	VisitMetricsCommand(ctx *MetricsCommandContext) interface{}

	// Visit a parse tree produced by SPL2Parser#metricsAggregates.
	VisitMetricsAggregates(ctx *MetricsAggregatesContext) interface{}

	// Visit a parse tree produced by SPL2Parser#metricsOption.
	VisitMetricsOption(ctx *MetricsOptionContext) interface{}

	// Visit a parse tree produced by SPL2Parser#timechartCommand.
	VisitTimechartCommand(ctx *TimechartCommandContext) interface{}

	// Visit a parse tree produced by SPL2Parser#timechartExtraAggregate.
	VisitTimechartExtraAggregate(ctx *TimechartExtraAggregateContext) interface{}

	// Visit a parse tree produced by SPL2Parser#timechartOption.
	VisitTimechartOption(ctx *TimechartOptionContext) interface{}

	// Visit a parse tree produced by SPL2Parser#timechartSplit.
	VisitTimechartSplit(ctx *TimechartSplitContext) interface{}

	// Visit a parse tree produced by SPL2Parser#timechartSplitOption.
	VisitTimechartSplitOption(ctx *TimechartSplitOptionContext) interface{}

	// Visit a parse tree produced by SPL2Parser#timewrapCommand.
	VisitTimewrapCommand(ctx *TimewrapCommandContext) interface{}

	// Visit a parse tree produced by SPL2Parser#makemvCommand.
	VisitMakemvCommand(ctx *MakemvCommandContext) interface{}

	// Visit a parse tree produced by SPL2Parser#mvexpandCommand.
	VisitMvexpandCommand(ctx *MvexpandCommandContext) interface{}

	// Visit a parse tree produced by SPL2Parser#mvcombineCommand.
	VisitMvcombineCommand(ctx *MvcombineCommandContext) interface{}

	// Visit a parse tree produced by SPL2Parser#fillnullCommand.
	VisitFillnullCommand(ctx *FillnullCommandContext) interface{}

	// Visit a parse tree produced by SPL2Parser#embeddedCommand.
	VisitEmbeddedCommand(ctx *EmbeddedCommandContext) interface{}

	// Visit a parse tree produced by SPL2Parser#embeddedText.
	VisitEmbeddedText(ctx *EmbeddedTextContext) interface{}

	// Visit a parse tree produced by SPL2Parser#moduleSuffix.
	VisitModuleSuffix(ctx *ModuleSuffixContext) interface{}

	// Visit a parse tree produced by SPL2Parser#moduleDeclaration.
	VisitModuleDeclaration(ctx *ModuleDeclarationContext) interface{}

	// Visit a parse tree produced by SPL2Parser#searchCommand.
	VisitSearchCommand(ctx *SearchCommandContext) interface{}

	// Visit a parse tree produced by SPL2Parser#implicitSearch.
	VisitImplicitSearch(ctx *ImplicitSearchContext) interface{}

	// Visit a parse tree produced by SPL2Parser#searchExpression.
	VisitSearchExpression(ctx *SearchExpressionContext) interface{}

	// Visit a parse tree produced by SPL2Parser#searchXor.
	VisitSearchXor(ctx *SearchXorContext) interface{}

	// Visit a parse tree produced by SPL2Parser#searchAnd.
	VisitSearchAnd(ctx *SearchAndContext) interface{}

	// Visit a parse tree produced by SPL2Parser#searchOr.
	VisitSearchOr(ctx *SearchOrContext) interface{}

	// Visit a parse tree produced by SPL2Parser#searchNot.
	VisitSearchNot(ctx *SearchNotContext) interface{}

	// Visit a parse tree produced by SPL2Parser#searchAtom.
	VisitSearchAtom(ctx *SearchAtomContext) interface{}

	// Visit a parse tree produced by SPL2Parser#searchValue.
	VisitSearchValue(ctx *SearchValueContext) interface{}

	// Visit a parse tree produced by SPL2Parser#searchWordLiteral.
	VisitSearchWordLiteral(ctx *SearchWordLiteralContext) interface{}

	// Visit a parse tree produced by SPL2Parser#searchSignedNumber.
	VisitSearchSignedNumber(ctx *SearchSignedNumberContext) interface{}

	// Visit a parse tree produced by SPL2Parser#searchUnprovedLiteral.
	VisitSearchUnprovedLiteral(ctx *SearchUnprovedLiteralContext) interface{}

	// Visit a parse tree produced by SPL2Parser#searchBareValue.
	VisitSearchBareValue(ctx *SearchBareValueContext) interface{}

	// Visit a parse tree produced by SPL2Parser#searchDirective.
	VisitSearchDirective(ctx *SearchDirectiveContext) interface{}

	// Visit a parse tree produced by SPL2Parser#searchTimeModifier.
	VisitSearchTimeModifier(ctx *SearchTimeModifierContext) interface{}

	// Visit a parse tree produced by SPL2Parser#timeModifierKey.
	VisitTimeModifierKey(ctx *TimeModifierKeyContext) interface{}

	// Visit a parse tree produced by SPL2Parser#timeModifierValue.
	VisitTimeModifierValue(ctx *TimeModifierValueContext) interface{}

	// Visit a parse tree produced by SPL2Parser#relativeTime.
	VisitRelativeTime(ctx *RelativeTimeContext) interface{}

	// Visit a parse tree produced by SPL2Parser#expression.
	VisitExpression(ctx *ExpressionContext) interface{}

	// Visit a parse tree produced by SPL2Parser#xorExpression.
	VisitXorExpression(ctx *XorExpressionContext) interface{}

	// Visit a parse tree produced by SPL2Parser#orExpression.
	VisitOrExpression(ctx *OrExpressionContext) interface{}

	// Visit a parse tree produced by SPL2Parser#andExpression.
	VisitAndExpression(ctx *AndExpressionContext) interface{}

	// Visit a parse tree produced by SPL2Parser#notExpression.
	VisitNotExpression(ctx *NotExpressionContext) interface{}

	// Visit a parse tree produced by SPL2Parser#predicate.
	VisitPredicate(ctx *PredicateContext) interface{}

	// Visit a parse tree produced by SPL2Parser#logicalAnd.
	VisitLogicalAnd(ctx *LogicalAndContext) interface{}

	// Visit a parse tree produced by SPL2Parser#logicalOr.
	VisitLogicalOr(ctx *LogicalOrContext) interface{}

	// Visit a parse tree produced by SPL2Parser#logicalXor.
	VisitLogicalXor(ctx *LogicalXorContext) interface{}

	// Visit a parse tree produced by SPL2Parser#logicalNot.
	VisitLogicalNot(ctx *LogicalNotContext) interface{}

	// Visit a parse tree produced by SPL2Parser#betweenOperator.
	VisitBetweenOperator(ctx *BetweenOperatorContext) interface{}

	// Visit a parse tree produced by SPL2Parser#betweenConjunction.
	VisitBetweenConjunction(ctx *BetweenConjunctionContext) interface{}

	// Visit a parse tree produced by SPL2Parser#comparison.
	VisitComparison(ctx *ComparisonContext) interface{}

	// Visit a parse tree produced by SPL2Parser#additive.
	VisitAdditive(ctx *AdditiveContext) interface{}

	// Visit a parse tree produced by SPL2Parser#multiplicative.
	VisitMultiplicative(ctx *MultiplicativeContext) interface{}

	// Visit a parse tree produced by SPL2Parser#unary.
	VisitUnary(ctx *UnaryContext) interface{}

	// Visit a parse tree produced by SPL2Parser#access.
	VisitAccess(ctx *AccessContext) interface{}

	// Visit a parse tree produced by SPL2Parser#accessPart.
	VisitAccessPart(ctx *AccessPartContext) interface{}

	// Visit a parse tree produced by SPL2Parser#primary.
	VisitPrimary(ctx *PrimaryContext) interface{}

	// Visit a parse tree produced by SPL2Parser#call.
	VisitCall(ctx *CallContext) interface{}

	// Visit a parse tree produced by SPL2Parser#arguments.
	VisitArguments(ctx *ArgumentsContext) interface{}

	// Visit a parse tree produced by SPL2Parser#namedArgument.
	VisitNamedArgument(ctx *NamedArgumentContext) interface{}

	// Visit a parse tree produced by SPL2Parser#literal.
	VisitLiteral(ctx *LiteralContext) interface{}

	// Visit a parse tree produced by SPL2Parser#stringLiteral.
	VisitStringLiteral(ctx *StringLiteralContext) interface{}

	// Visit a parse tree produced by SPL2Parser#quotedName.
	VisitQuotedName(ctx *QuotedNameContext) interface{}

	// Visit a parse tree produced by SPL2Parser#fieldTemplate.
	VisitFieldTemplate(ctx *FieldTemplateContext) interface{}

	// Visit a parse tree produced by SPL2Parser#fieldName.
	VisitFieldName(ctx *FieldNameContext) interface{}

	// Visit a parse tree produced by SPL2Parser#identifier.
	VisitIdentifier(ctx *IdentifierContext) interface{}

	// Visit a parse tree produced by SPL2Parser#sqlKeyword.
	VisitSqlKeyword(ctx *SqlKeywordContext) interface{}

	// Visit a parse tree produced by SPL2Parser#pipelineKeyword.
	VisitPipelineKeyword(ctx *PipelineKeywordContext) interface{}

	// Visit a parse tree produced by SPL2Parser#array.
	VisitArray(ctx *ArrayContext) interface{}

	// Visit a parse tree produced by SPL2Parser#object.
	VisitObject(ctx *ObjectContext) interface{}

	// Visit a parse tree produced by SPL2Parser#objectEntry.
	VisitObjectEntry(ctx *ObjectEntryContext) interface{}

	// Visit a parse tree produced by SPL2Parser#objectKey.
	VisitObjectKey(ctx *ObjectKeyContext) interface{}

	// Visit a parse tree produced by SPL2Parser#searchLiteral.
	VisitSearchLiteral(ctx *SearchLiteralContext) interface{}

	// Visit a parse tree produced by SPL2Parser#lambdaExpression.
	VisitLambdaExpression(ctx *LambdaExpressionContext) interface{}

	// Visit a parse tree produced by SPL2Parser#lambdaParameter.
	VisitLambdaParameter(ctx *LambdaParameterContext) interface{}

	// Visit a parse tree produced by SPL2Parser#lambdaBlock.
	VisitLambdaBlock(ctx *LambdaBlockContext) interface{}
}
