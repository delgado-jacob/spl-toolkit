// Code generated from grammar/SPL2Parser.g4 by ANTLR 4.13.2. DO NOT EDIT.

package spl2 // SPL2Parser
import "github.com/antlr4-go/antlr/v4"

// SPL2ParserListener is a complete listener for a parse tree produced by SPL2Parser.
type SPL2ParserListener interface {
	antlr.ParseTreeListener

	// EnterQuery is called when entering the query production.
	EnterQuery(c *QueryContext)

	// EnterPipeline is called when entering the pipeline production.
	EnterPipeline(c *PipelineContext)

	// EnterStart is called when entering the start production.
	EnterStart(c *StartContext)

	// EnterCommand is called when entering the command production.
	EnterCommand(c *CommandContext)

	// EnterFromCommand is called when entering the fromCommand production.
	EnterFromCommand(c *FromCommandContext)

	// EnterSelectCommand is called when entering the selectCommand production.
	EnterSelectCommand(c *SelectCommandContext)

	// EnterSqlFromClause is called when entering the sqlFromClause production.
	EnterSqlFromClause(c *SqlFromClauseContext)

	// EnterSourceAlias is called when entering the sourceAlias production.
	EnterSourceAlias(c *SourceAliasContext)

	// EnterSqlJoinClause is called when entering the sqlJoinClause production.
	EnterSqlJoinClause(c *SqlJoinClauseContext)

	// EnterSqlJoinPredicate is called when entering the sqlJoinPredicate production.
	EnterSqlJoinPredicate(c *SqlJoinPredicateContext)

	// EnterSqlJoinEquality is called when entering the sqlJoinEquality production.
	EnterSqlJoinEquality(c *SqlJoinEqualityContext)

	// EnterSqlJoinField is called when entering the sqlJoinField production.
	EnterSqlJoinField(c *SqlJoinFieldContext)

	// EnterSqlSelectClause is called when entering the sqlSelectClause production.
	EnterSqlSelectClause(c *SqlSelectClauseContext)

	// EnterProjection is called when entering the projection production.
	EnterProjection(c *ProjectionContext)

	// EnterProjectionAlias is called when entering the projectionAlias production.
	EnterProjectionAlias(c *ProjectionAliasContext)

	// EnterSqlWhereClause is called when entering the sqlWhereClause production.
	EnterSqlWhereClause(c *SqlWhereClauseContext)

	// EnterSqlHavingClause is called when entering the sqlHavingClause production.
	EnterSqlHavingClause(c *SqlHavingClauseContext)

	// EnterSqlPredicate is called when entering the sqlPredicate production.
	EnterSqlPredicate(c *SqlPredicateContext)

	// EnterSqlGroupClause is called when entering the sqlGroupClause production.
	EnterSqlGroupClause(c *SqlGroupClauseContext)

	// EnterSqlBy is called when entering the sqlBy production.
	EnterSqlBy(c *SqlByContext)

	// EnterSqlGroupKey is called when entering the sqlGroupKey production.
	EnterSqlGroupKey(c *SqlGroupKeyContext)

	// EnterSqlSpanCall is called when entering the sqlSpanCall production.
	EnterSqlSpanCall(c *SqlSpanCallContext)

	// EnterSqlSpanAssignment is called when entering the sqlSpanAssignment production.
	EnterSqlSpanAssignment(c *SqlSpanAssignmentContext)

	// EnterSqlUnparenthesizedSpan is called when entering the sqlUnparenthesizedSpan production.
	EnterSqlUnparenthesizedSpan(c *SqlUnparenthesizedSpanContext)

	// EnterSqlOrderClause is called when entering the sqlOrderClause production.
	EnterSqlOrderClause(c *SqlOrderClauseContext)

	// EnterSqlOrderTerm is called when entering the sqlOrderTerm production.
	EnterSqlOrderTerm(c *SqlOrderTermContext)

	// EnterSqlDirection is called when entering the sqlDirection production.
	EnterSqlDirection(c *SqlDirectionContext)

	// EnterSqlLimitClause is called when entering the sqlLimitClause production.
	EnterSqlLimitClause(c *SqlLimitClauseContext)

	// EnterSqlOffsetClause is called when entering the sqlOffsetClause production.
	EnterSqlOffsetClause(c *SqlOffsetClauseContext)

	// EnterExistsPredicate is called when entering the existsPredicate production.
	EnterExistsPredicate(c *ExistsPredicateContext)

	// EnterDataset is called when entering the dataset production.
	EnterDataset(c *DatasetContext)

	// EnterGenerator is called when entering the generator production.
	EnterGenerator(c *GeneratorContext)

	// EnterEvalCommand is called when entering the evalCommand production.
	EnterEvalCommand(c *EvalCommandContext)

	// EnterAssignment is called when entering the assignment production.
	EnterAssignment(c *AssignmentContext)

	// EnterWhereCommand is called when entering the whereCommand production.
	EnterWhereCommand(c *WhereCommandContext)

	// EnterFieldsCommand is called when entering the fieldsCommand production.
	EnterFieldsCommand(c *FieldsCommandContext)

	// EnterFieldSelection is called when entering the fieldSelection production.
	EnterFieldSelection(c *FieldSelectionContext)

	// EnterFieldSelector is called when entering the fieldSelector production.
	EnterFieldSelector(c *FieldSelectorContext)

	// EnterTableCommand is called when entering the tableCommand production.
	EnterTableCommand(c *TableCommandContext)

	// EnterTableField is called when entering the tableField production.
	EnterTableField(c *TableFieldContext)

	// EnterRenameCommand is called when entering the renameCommand production.
	EnterRenameCommand(c *RenameCommandContext)

	// EnterRenamePair is called when entering the renamePair production.
	EnterRenamePair(c *RenamePairContext)

	// EnterRenameSource is called when entering the renameSource production.
	EnterRenameSource(c *RenameSourceContext)

	// EnterRenameTarget is called when entering the renameTarget production.
	EnterRenameTarget(c *RenameTargetContext)

	// EnterAliasKeyword is called when entering the aliasKeyword production.
	EnterAliasKeyword(c *AliasKeywordContext)

	// EnterStatsCommand is called when entering the statsCommand production.
	EnterStatsCommand(c *StatsCommandContext)

	// EnterStatsOption is called when entering the statsOption production.
	EnterStatsOption(c *StatsOptionContext)

	// EnterAllnumOption is called when entering the allnumOption production.
	EnterAllnumOption(c *AllnumOptionContext)

	// EnterDelimOption is called when entering the delimOption production.
	EnterDelimOption(c *DelimOptionContext)

	// EnterPartitionsOption is called when entering the partitionsOption production.
	EnterPartitionsOption(c *PartitionsOptionContext)

	// EnterAggregate is called when entering the aggregate production.
	EnterAggregate(c *AggregateContext)

	// EnterAggregateAlias is called when entering the aggregateAlias production.
	EnterAggregateAlias(c *AggregateAliasContext)

	// EnterAggregateGroup is called when entering the aggregateGroup production.
	EnterAggregateGroup(c *AggregateGroupContext)

	// EnterGroupField is called when entering the groupField production.
	EnterGroupField(c *GroupFieldContext)

	// EnterGroupSpan is called when entering the groupSpan production.
	EnterGroupSpan(c *GroupSpanContext)

	// EnterTimeSpan is called when entering the timeSpan production.
	EnterTimeSpan(c *TimeSpanContext)

	// EnterEventstatsCommand is called when entering the eventstatsCommand production.
	EnterEventstatsCommand(c *EventstatsCommandContext)

	// EnterStreamstatsCommand is called when entering the streamstatsCommand production.
	EnterStreamstatsCommand(c *StreamstatsCommandContext)

	// EnterStreamGroup is called when entering the streamGroup production.
	EnterStreamGroup(c *StreamGroupContext)

	// EnterCurrentOption is called when entering the currentOption production.
	EnterCurrentOption(c *CurrentOptionContext)

	// EnterWindowOption is called when entering the windowOption production.
	EnterWindowOption(c *WindowOptionContext)

	// EnterResetClause is called when entering the resetClause production.
	EnterResetClause(c *ResetClauseContext)

	// EnterResetBefore is called when entering the resetBefore production.
	EnterResetBefore(c *ResetBeforeContext)

	// EnterResetAfter is called when entering the resetAfter production.
	EnterResetAfter(c *ResetAfterContext)

	// EnterResetOnchange is called when entering the resetOnchange production.
	EnterResetOnchange(c *ResetOnchangeContext)

	// EnterStreamPostLayout is called when entering the streamPostLayout production.
	EnterStreamPostLayout(c *StreamPostLayoutContext)

	// EnterLookupCommand is called when entering the lookupCommand production.
	EnterLookupCommand(c *LookupCommandContext)

	// EnterLookupDataset is called when entering the lookupDataset production.
	EnterLookupDataset(c *LookupDatasetContext)

	// EnterLookupMatch is called when entering the lookupMatch production.
	EnterLookupMatch(c *LookupMatchContext)

	// EnterLookupOutputClause is called when entering the lookupOutputClause production.
	EnterLookupOutputClause(c *LookupOutputClauseContext)

	// EnterLookupOutput is called when entering the lookupOutput production.
	EnterLookupOutput(c *LookupOutputContext)

	// EnterLookupColumn is called when entering the lookupColumn production.
	EnterLookupColumn(c *LookupColumnContext)

	// EnterLookupEventField is called when entering the lookupEventField production.
	EnterLookupEventField(c *LookupEventFieldContext)

	// EnterSortCommand is called when entering the sortCommand production.
	EnterSortCommand(c *SortCommandContext)

	// EnterSortTerm is called when entering the sortTerm production.
	EnterSortTerm(c *SortTermContext)

	// EnterSortWrapper is called when entering the sortWrapper production.
	EnterSortWrapper(c *SortWrapperContext)

	// EnterIntegerValue is called when entering the integerValue production.
	EnterIntegerValue(c *IntegerValueContext)

	// EnterDedupCommand is called when entering the dedupCommand production.
	EnterDedupCommand(c *DedupCommandContext)

	// EnterKeepemptyOption is called when entering the keepemptyOption production.
	EnterKeepemptyOption(c *KeepemptyOptionContext)

	// EnterConsecutiveOption is called when entering the consecutiveOption production.
	EnterConsecutiveOption(c *ConsecutiveOptionContext)

	// EnterDedupField is called when entering the dedupField production.
	EnterDedupField(c *DedupFieldContext)

	// EnterHeadCommand is called when entering the headCommand production.
	EnterHeadCommand(c *HeadCommandContext)

	// EnterKeeplastOption is called when entering the keeplastOption production.
	EnterKeeplastOption(c *KeeplastOptionContext)

	// EnterHeadWhile is called when entering the headWhile production.
	EnterHeadWhile(c *HeadWhileContext)

	// EnterHeadPostLayout is called when entering the headPostLayout production.
	EnterHeadPostLayout(c *HeadPostLayoutContext)

	// EnterReverseCommand is called when entering the reverseCommand production.
	EnterReverseCommand(c *ReverseCommandContext)

	// EnterUnknownOption is called when entering the unknownOption production.
	EnterUnknownOption(c *UnknownOptionContext)

	// EnterUnknownOptionName is called when entering the unknownOptionName production.
	EnterUnknownOptionName(c *UnknownOptionNameContext)

	// EnterIndependentSearch is called when entering the independentSearch production.
	EnterIndependentSearch(c *IndependentSearchContext)

	// EnterInheritedSubpipe is called when entering the inheritedSubpipe production.
	EnterInheritedSubpipe(c *InheritedSubpipeContext)

	// EnterJoinCommand is called when entering the joinCommand production.
	EnterJoinCommand(c *JoinCommandContext)

	// EnterJoinOption is called when entering the joinOption production.
	EnterJoinOption(c *JoinOptionContext)

	// EnterJoinType is called when entering the joinType production.
	EnterJoinType(c *JoinTypeContext)

	// EnterAppendCommand is called when entering the appendCommand production.
	EnterAppendCommand(c *AppendCommandContext)

	// EnterAppendpipeCommand is called when entering the appendpipeCommand production.
	EnterAppendpipeCommand(c *AppendpipeCommandContext)

	// EnterAppendcolsCommand is called when entering the appendcolsCommand production.
	EnterAppendcolsCommand(c *AppendcolsCommandContext)

	// EnterUnionCommand is called when entering the unionCommand production.
	EnterUnionCommand(c *UnionCommandContext)

	// EnterUnionDataset is called when entering the unionDataset production.
	EnterUnionDataset(c *UnionDatasetContext)

	// EnterIfCommand is called when entering the ifCommand production.
	EnterIfCommand(c *IfCommandContext)

	// EnterBinCommand is called when entering the binCommand production.
	EnterBinCommand(c *BinCommandContext)

	// EnterBinOption is called when entering the binOption production.
	EnterBinOption(c *BinOptionContext)

	// EnterAlignmentOption is called when entering the alignmentOption production.
	EnterAlignmentOption(c *AlignmentOptionContext)

	// EnterBinSpan is called when entering the binSpan production.
	EnterBinSpan(c *BinSpanContext)

	// EnterLogarithmicSpan is called when entering the logarithmicSpan production.
	EnterLogarithmicSpan(c *LogarithmicSpanContext)

	// EnterExtendedOption is called when entering the extendedOption production.
	EnterExtendedOption(c *ExtendedOptionContext)

	// EnterSignedNumber is called when entering the signedNumber production.
	EnterSignedNumber(c *SignedNumberContext)

	// EnterRexCommand is called when entering the rexCommand production.
	EnterRexCommand(c *RexCommandContext)

	// EnterRexOption is called when entering the rexOption production.
	EnterRexOption(c *RexOptionContext)

	// EnterSpathCommand is called when entering the spathCommand production.
	EnterSpathCommand(c *SpathCommandContext)

	// EnterSpathOption is called when entering the spathOption production.
	EnterSpathOption(c *SpathOptionContext)

	// EnterLoadjobCommand is called when entering the loadjobCommand production.
	EnterLoadjobCommand(c *LoadjobCommandContext)

	// EnterMetricsCommand is called when entering the metricsCommand production.
	EnterMetricsCommand(c *MetricsCommandContext)

	// EnterMetricsAggregates is called when entering the metricsAggregates production.
	EnterMetricsAggregates(c *MetricsAggregatesContext)

	// EnterMetricsOption is called when entering the metricsOption production.
	EnterMetricsOption(c *MetricsOptionContext)

	// EnterTimechartCommand is called when entering the timechartCommand production.
	EnterTimechartCommand(c *TimechartCommandContext)

	// EnterTimechartExtraAggregate is called when entering the timechartExtraAggregate production.
	EnterTimechartExtraAggregate(c *TimechartExtraAggregateContext)

	// EnterTimechartOption is called when entering the timechartOption production.
	EnterTimechartOption(c *TimechartOptionContext)

	// EnterTimechartSplit is called when entering the timechartSplit production.
	EnterTimechartSplit(c *TimechartSplitContext)

	// EnterTimechartSplitOption is called when entering the timechartSplitOption production.
	EnterTimechartSplitOption(c *TimechartSplitOptionContext)

	// EnterTimewrapCommand is called when entering the timewrapCommand production.
	EnterTimewrapCommand(c *TimewrapCommandContext)

	// EnterMakemvCommand is called when entering the makemvCommand production.
	EnterMakemvCommand(c *MakemvCommandContext)

	// EnterMvexpandCommand is called when entering the mvexpandCommand production.
	EnterMvexpandCommand(c *MvexpandCommandContext)

	// EnterMvcombineCommand is called when entering the mvcombineCommand production.
	EnterMvcombineCommand(c *MvcombineCommandContext)

	// EnterFillnullCommand is called when entering the fillnullCommand production.
	EnterFillnullCommand(c *FillnullCommandContext)

	// EnterEmbeddedCommand is called when entering the embeddedCommand production.
	EnterEmbeddedCommand(c *EmbeddedCommandContext)

	// EnterEmbeddedText is called when entering the embeddedText production.
	EnterEmbeddedText(c *EmbeddedTextContext)

	// EnterModuleSuffix is called when entering the moduleSuffix production.
	EnterModuleSuffix(c *ModuleSuffixContext)

	// EnterModuleDeclaration is called when entering the moduleDeclaration production.
	EnterModuleDeclaration(c *ModuleDeclarationContext)

	// EnterSearchCommand is called when entering the searchCommand production.
	EnterSearchCommand(c *SearchCommandContext)

	// EnterImplicitSearch is called when entering the implicitSearch production.
	EnterImplicitSearch(c *ImplicitSearchContext)

	// EnterSearchExpression is called when entering the searchExpression production.
	EnterSearchExpression(c *SearchExpressionContext)

	// EnterSearchXor is called when entering the searchXor production.
	EnterSearchXor(c *SearchXorContext)

	// EnterSearchAnd is called when entering the searchAnd production.
	EnterSearchAnd(c *SearchAndContext)

	// EnterSearchOr is called when entering the searchOr production.
	EnterSearchOr(c *SearchOrContext)

	// EnterSearchNot is called when entering the searchNot production.
	EnterSearchNot(c *SearchNotContext)

	// EnterSearchAtom is called when entering the searchAtom production.
	EnterSearchAtom(c *SearchAtomContext)

	// EnterSearchValue is called when entering the searchValue production.
	EnterSearchValue(c *SearchValueContext)

	// EnterSearchWordLiteral is called when entering the searchWordLiteral production.
	EnterSearchWordLiteral(c *SearchWordLiteralContext)

	// EnterSearchSignedNumber is called when entering the searchSignedNumber production.
	EnterSearchSignedNumber(c *SearchSignedNumberContext)

	// EnterSearchUnprovedLiteral is called when entering the searchUnprovedLiteral production.
	EnterSearchUnprovedLiteral(c *SearchUnprovedLiteralContext)

	// EnterSearchBareValue is called when entering the searchBareValue production.
	EnterSearchBareValue(c *SearchBareValueContext)

	// EnterSearchDirective is called when entering the searchDirective production.
	EnterSearchDirective(c *SearchDirectiveContext)

	// EnterSearchTimeModifier is called when entering the searchTimeModifier production.
	EnterSearchTimeModifier(c *SearchTimeModifierContext)

	// EnterTimeModifierKey is called when entering the timeModifierKey production.
	EnterTimeModifierKey(c *TimeModifierKeyContext)

	// EnterTimeModifierValue is called when entering the timeModifierValue production.
	EnterTimeModifierValue(c *TimeModifierValueContext)

	// EnterRelativeTime is called when entering the relativeTime production.
	EnterRelativeTime(c *RelativeTimeContext)

	// EnterExpression is called when entering the expression production.
	EnterExpression(c *ExpressionContext)

	// EnterXorExpression is called when entering the xorExpression production.
	EnterXorExpression(c *XorExpressionContext)

	// EnterOrExpression is called when entering the orExpression production.
	EnterOrExpression(c *OrExpressionContext)

	// EnterAndExpression is called when entering the andExpression production.
	EnterAndExpression(c *AndExpressionContext)

	// EnterNotExpression is called when entering the notExpression production.
	EnterNotExpression(c *NotExpressionContext)

	// EnterPredicate is called when entering the predicate production.
	EnterPredicate(c *PredicateContext)

	// EnterLogicalAnd is called when entering the logicalAnd production.
	EnterLogicalAnd(c *LogicalAndContext)

	// EnterLogicalOr is called when entering the logicalOr production.
	EnterLogicalOr(c *LogicalOrContext)

	// EnterLogicalXor is called when entering the logicalXor production.
	EnterLogicalXor(c *LogicalXorContext)

	// EnterLogicalNot is called when entering the logicalNot production.
	EnterLogicalNot(c *LogicalNotContext)

	// EnterBetweenOperator is called when entering the betweenOperator production.
	EnterBetweenOperator(c *BetweenOperatorContext)

	// EnterBetweenConjunction is called when entering the betweenConjunction production.
	EnterBetweenConjunction(c *BetweenConjunctionContext)

	// EnterComparison is called when entering the comparison production.
	EnterComparison(c *ComparisonContext)

	// EnterAdditive is called when entering the additive production.
	EnterAdditive(c *AdditiveContext)

	// EnterMultiplicative is called when entering the multiplicative production.
	EnterMultiplicative(c *MultiplicativeContext)

	// EnterUnary is called when entering the unary production.
	EnterUnary(c *UnaryContext)

	// EnterAccess is called when entering the access production.
	EnterAccess(c *AccessContext)

	// EnterAccessPart is called when entering the accessPart production.
	EnterAccessPart(c *AccessPartContext)

	// EnterPrimary is called when entering the primary production.
	EnterPrimary(c *PrimaryContext)

	// EnterCall is called when entering the call production.
	EnterCall(c *CallContext)

	// EnterArguments is called when entering the arguments production.
	EnterArguments(c *ArgumentsContext)

	// EnterNamedArgument is called when entering the namedArgument production.
	EnterNamedArgument(c *NamedArgumentContext)

	// EnterLiteral is called when entering the literal production.
	EnterLiteral(c *LiteralContext)

	// EnterStringLiteral is called when entering the stringLiteral production.
	EnterStringLiteral(c *StringLiteralContext)

	// EnterQuotedName is called when entering the quotedName production.
	EnterQuotedName(c *QuotedNameContext)

	// EnterFieldTemplate is called when entering the fieldTemplate production.
	EnterFieldTemplate(c *FieldTemplateContext)

	// EnterFieldName is called when entering the fieldName production.
	EnterFieldName(c *FieldNameContext)

	// EnterIdentifier is called when entering the identifier production.
	EnterIdentifier(c *IdentifierContext)

	// EnterSqlKeyword is called when entering the sqlKeyword production.
	EnterSqlKeyword(c *SqlKeywordContext)

	// EnterPipelineKeyword is called when entering the pipelineKeyword production.
	EnterPipelineKeyword(c *PipelineKeywordContext)

	// EnterArray is called when entering the array production.
	EnterArray(c *ArrayContext)

	// EnterObject is called when entering the object production.
	EnterObject(c *ObjectContext)

	// EnterObjectEntry is called when entering the objectEntry production.
	EnterObjectEntry(c *ObjectEntryContext)

	// EnterObjectKey is called when entering the objectKey production.
	EnterObjectKey(c *ObjectKeyContext)

	// EnterSearchLiteral is called when entering the searchLiteral production.
	EnterSearchLiteral(c *SearchLiteralContext)

	// EnterLambdaExpression is called when entering the lambdaExpression production.
	EnterLambdaExpression(c *LambdaExpressionContext)

	// EnterLambdaParameter is called when entering the lambdaParameter production.
	EnterLambdaParameter(c *LambdaParameterContext)

	// EnterLambdaBlock is called when entering the lambdaBlock production.
	EnterLambdaBlock(c *LambdaBlockContext)

	// ExitQuery is called when exiting the query production.
	ExitQuery(c *QueryContext)

	// ExitPipeline is called when exiting the pipeline production.
	ExitPipeline(c *PipelineContext)

	// ExitStart is called when exiting the start production.
	ExitStart(c *StartContext)

	// ExitCommand is called when exiting the command production.
	ExitCommand(c *CommandContext)

	// ExitFromCommand is called when exiting the fromCommand production.
	ExitFromCommand(c *FromCommandContext)

	// ExitSelectCommand is called when exiting the selectCommand production.
	ExitSelectCommand(c *SelectCommandContext)

	// ExitSqlFromClause is called when exiting the sqlFromClause production.
	ExitSqlFromClause(c *SqlFromClauseContext)

	// ExitSourceAlias is called when exiting the sourceAlias production.
	ExitSourceAlias(c *SourceAliasContext)

	// ExitSqlJoinClause is called when exiting the sqlJoinClause production.
	ExitSqlJoinClause(c *SqlJoinClauseContext)

	// ExitSqlJoinPredicate is called when exiting the sqlJoinPredicate production.
	ExitSqlJoinPredicate(c *SqlJoinPredicateContext)

	// ExitSqlJoinEquality is called when exiting the sqlJoinEquality production.
	ExitSqlJoinEquality(c *SqlJoinEqualityContext)

	// ExitSqlJoinField is called when exiting the sqlJoinField production.
	ExitSqlJoinField(c *SqlJoinFieldContext)

	// ExitSqlSelectClause is called when exiting the sqlSelectClause production.
	ExitSqlSelectClause(c *SqlSelectClauseContext)

	// ExitProjection is called when exiting the projection production.
	ExitProjection(c *ProjectionContext)

	// ExitProjectionAlias is called when exiting the projectionAlias production.
	ExitProjectionAlias(c *ProjectionAliasContext)

	// ExitSqlWhereClause is called when exiting the sqlWhereClause production.
	ExitSqlWhereClause(c *SqlWhereClauseContext)

	// ExitSqlHavingClause is called when exiting the sqlHavingClause production.
	ExitSqlHavingClause(c *SqlHavingClauseContext)

	// ExitSqlPredicate is called when exiting the sqlPredicate production.
	ExitSqlPredicate(c *SqlPredicateContext)

	// ExitSqlGroupClause is called when exiting the sqlGroupClause production.
	ExitSqlGroupClause(c *SqlGroupClauseContext)

	// ExitSqlBy is called when exiting the sqlBy production.
	ExitSqlBy(c *SqlByContext)

	// ExitSqlGroupKey is called when exiting the sqlGroupKey production.
	ExitSqlGroupKey(c *SqlGroupKeyContext)

	// ExitSqlSpanCall is called when exiting the sqlSpanCall production.
	ExitSqlSpanCall(c *SqlSpanCallContext)

	// ExitSqlSpanAssignment is called when exiting the sqlSpanAssignment production.
	ExitSqlSpanAssignment(c *SqlSpanAssignmentContext)

	// ExitSqlUnparenthesizedSpan is called when exiting the sqlUnparenthesizedSpan production.
	ExitSqlUnparenthesizedSpan(c *SqlUnparenthesizedSpanContext)

	// ExitSqlOrderClause is called when exiting the sqlOrderClause production.
	ExitSqlOrderClause(c *SqlOrderClauseContext)

	// ExitSqlOrderTerm is called when exiting the sqlOrderTerm production.
	ExitSqlOrderTerm(c *SqlOrderTermContext)

	// ExitSqlDirection is called when exiting the sqlDirection production.
	ExitSqlDirection(c *SqlDirectionContext)

	// ExitSqlLimitClause is called when exiting the sqlLimitClause production.
	ExitSqlLimitClause(c *SqlLimitClauseContext)

	// ExitSqlOffsetClause is called when exiting the sqlOffsetClause production.
	ExitSqlOffsetClause(c *SqlOffsetClauseContext)

	// ExitExistsPredicate is called when exiting the existsPredicate production.
	ExitExistsPredicate(c *ExistsPredicateContext)

	// ExitDataset is called when exiting the dataset production.
	ExitDataset(c *DatasetContext)

	// ExitGenerator is called when exiting the generator production.
	ExitGenerator(c *GeneratorContext)

	// ExitEvalCommand is called when exiting the evalCommand production.
	ExitEvalCommand(c *EvalCommandContext)

	// ExitAssignment is called when exiting the assignment production.
	ExitAssignment(c *AssignmentContext)

	// ExitWhereCommand is called when exiting the whereCommand production.
	ExitWhereCommand(c *WhereCommandContext)

	// ExitFieldsCommand is called when exiting the fieldsCommand production.
	ExitFieldsCommand(c *FieldsCommandContext)

	// ExitFieldSelection is called when exiting the fieldSelection production.
	ExitFieldSelection(c *FieldSelectionContext)

	// ExitFieldSelector is called when exiting the fieldSelector production.
	ExitFieldSelector(c *FieldSelectorContext)

	// ExitTableCommand is called when exiting the tableCommand production.
	ExitTableCommand(c *TableCommandContext)

	// ExitTableField is called when exiting the tableField production.
	ExitTableField(c *TableFieldContext)

	// ExitRenameCommand is called when exiting the renameCommand production.
	ExitRenameCommand(c *RenameCommandContext)

	// ExitRenamePair is called when exiting the renamePair production.
	ExitRenamePair(c *RenamePairContext)

	// ExitRenameSource is called when exiting the renameSource production.
	ExitRenameSource(c *RenameSourceContext)

	// ExitRenameTarget is called when exiting the renameTarget production.
	ExitRenameTarget(c *RenameTargetContext)

	// ExitAliasKeyword is called when exiting the aliasKeyword production.
	ExitAliasKeyword(c *AliasKeywordContext)

	// ExitStatsCommand is called when exiting the statsCommand production.
	ExitStatsCommand(c *StatsCommandContext)

	// ExitStatsOption is called when exiting the statsOption production.
	ExitStatsOption(c *StatsOptionContext)

	// ExitAllnumOption is called when exiting the allnumOption production.
	ExitAllnumOption(c *AllnumOptionContext)

	// ExitDelimOption is called when exiting the delimOption production.
	ExitDelimOption(c *DelimOptionContext)

	// ExitPartitionsOption is called when exiting the partitionsOption production.
	ExitPartitionsOption(c *PartitionsOptionContext)

	// ExitAggregate is called when exiting the aggregate production.
	ExitAggregate(c *AggregateContext)

	// ExitAggregateAlias is called when exiting the aggregateAlias production.
	ExitAggregateAlias(c *AggregateAliasContext)

	// ExitAggregateGroup is called when exiting the aggregateGroup production.
	ExitAggregateGroup(c *AggregateGroupContext)

	// ExitGroupField is called when exiting the groupField production.
	ExitGroupField(c *GroupFieldContext)

	// ExitGroupSpan is called when exiting the groupSpan production.
	ExitGroupSpan(c *GroupSpanContext)

	// ExitTimeSpan is called when exiting the timeSpan production.
	ExitTimeSpan(c *TimeSpanContext)

	// ExitEventstatsCommand is called when exiting the eventstatsCommand production.
	ExitEventstatsCommand(c *EventstatsCommandContext)

	// ExitStreamstatsCommand is called when exiting the streamstatsCommand production.
	ExitStreamstatsCommand(c *StreamstatsCommandContext)

	// ExitStreamGroup is called when exiting the streamGroup production.
	ExitStreamGroup(c *StreamGroupContext)

	// ExitCurrentOption is called when exiting the currentOption production.
	ExitCurrentOption(c *CurrentOptionContext)

	// ExitWindowOption is called when exiting the windowOption production.
	ExitWindowOption(c *WindowOptionContext)

	// ExitResetClause is called when exiting the resetClause production.
	ExitResetClause(c *ResetClauseContext)

	// ExitResetBefore is called when exiting the resetBefore production.
	ExitResetBefore(c *ResetBeforeContext)

	// ExitResetAfter is called when exiting the resetAfter production.
	ExitResetAfter(c *ResetAfterContext)

	// ExitResetOnchange is called when exiting the resetOnchange production.
	ExitResetOnchange(c *ResetOnchangeContext)

	// ExitStreamPostLayout is called when exiting the streamPostLayout production.
	ExitStreamPostLayout(c *StreamPostLayoutContext)

	// ExitLookupCommand is called when exiting the lookupCommand production.
	ExitLookupCommand(c *LookupCommandContext)

	// ExitLookupDataset is called when exiting the lookupDataset production.
	ExitLookupDataset(c *LookupDatasetContext)

	// ExitLookupMatch is called when exiting the lookupMatch production.
	ExitLookupMatch(c *LookupMatchContext)

	// ExitLookupOutputClause is called when exiting the lookupOutputClause production.
	ExitLookupOutputClause(c *LookupOutputClauseContext)

	// ExitLookupOutput is called when exiting the lookupOutput production.
	ExitLookupOutput(c *LookupOutputContext)

	// ExitLookupColumn is called when exiting the lookupColumn production.
	ExitLookupColumn(c *LookupColumnContext)

	// ExitLookupEventField is called when exiting the lookupEventField production.
	ExitLookupEventField(c *LookupEventFieldContext)

	// ExitSortCommand is called when exiting the sortCommand production.
	ExitSortCommand(c *SortCommandContext)

	// ExitSortTerm is called when exiting the sortTerm production.
	ExitSortTerm(c *SortTermContext)

	// ExitSortWrapper is called when exiting the sortWrapper production.
	ExitSortWrapper(c *SortWrapperContext)

	// ExitIntegerValue is called when exiting the integerValue production.
	ExitIntegerValue(c *IntegerValueContext)

	// ExitDedupCommand is called when exiting the dedupCommand production.
	ExitDedupCommand(c *DedupCommandContext)

	// ExitKeepemptyOption is called when exiting the keepemptyOption production.
	ExitKeepemptyOption(c *KeepemptyOptionContext)

	// ExitConsecutiveOption is called when exiting the consecutiveOption production.
	ExitConsecutiveOption(c *ConsecutiveOptionContext)

	// ExitDedupField is called when exiting the dedupField production.
	ExitDedupField(c *DedupFieldContext)

	// ExitHeadCommand is called when exiting the headCommand production.
	ExitHeadCommand(c *HeadCommandContext)

	// ExitKeeplastOption is called when exiting the keeplastOption production.
	ExitKeeplastOption(c *KeeplastOptionContext)

	// ExitHeadWhile is called when exiting the headWhile production.
	ExitHeadWhile(c *HeadWhileContext)

	// ExitHeadPostLayout is called when exiting the headPostLayout production.
	ExitHeadPostLayout(c *HeadPostLayoutContext)

	// ExitReverseCommand is called when exiting the reverseCommand production.
	ExitReverseCommand(c *ReverseCommandContext)

	// ExitUnknownOption is called when exiting the unknownOption production.
	ExitUnknownOption(c *UnknownOptionContext)

	// ExitUnknownOptionName is called when exiting the unknownOptionName production.
	ExitUnknownOptionName(c *UnknownOptionNameContext)

	// ExitIndependentSearch is called when exiting the independentSearch production.
	ExitIndependentSearch(c *IndependentSearchContext)

	// ExitInheritedSubpipe is called when exiting the inheritedSubpipe production.
	ExitInheritedSubpipe(c *InheritedSubpipeContext)

	// ExitJoinCommand is called when exiting the joinCommand production.
	ExitJoinCommand(c *JoinCommandContext)

	// ExitJoinOption is called when exiting the joinOption production.
	ExitJoinOption(c *JoinOptionContext)

	// ExitJoinType is called when exiting the joinType production.
	ExitJoinType(c *JoinTypeContext)

	// ExitAppendCommand is called when exiting the appendCommand production.
	ExitAppendCommand(c *AppendCommandContext)

	// ExitAppendpipeCommand is called when exiting the appendpipeCommand production.
	ExitAppendpipeCommand(c *AppendpipeCommandContext)

	// ExitAppendcolsCommand is called when exiting the appendcolsCommand production.
	ExitAppendcolsCommand(c *AppendcolsCommandContext)

	// ExitUnionCommand is called when exiting the unionCommand production.
	ExitUnionCommand(c *UnionCommandContext)

	// ExitUnionDataset is called when exiting the unionDataset production.
	ExitUnionDataset(c *UnionDatasetContext)

	// ExitIfCommand is called when exiting the ifCommand production.
	ExitIfCommand(c *IfCommandContext)

	// ExitBinCommand is called when exiting the binCommand production.
	ExitBinCommand(c *BinCommandContext)

	// ExitBinOption is called when exiting the binOption production.
	ExitBinOption(c *BinOptionContext)

	// ExitAlignmentOption is called when exiting the alignmentOption production.
	ExitAlignmentOption(c *AlignmentOptionContext)

	// ExitBinSpan is called when exiting the binSpan production.
	ExitBinSpan(c *BinSpanContext)

	// ExitLogarithmicSpan is called when exiting the logarithmicSpan production.
	ExitLogarithmicSpan(c *LogarithmicSpanContext)

	// ExitExtendedOption is called when exiting the extendedOption production.
	ExitExtendedOption(c *ExtendedOptionContext)

	// ExitSignedNumber is called when exiting the signedNumber production.
	ExitSignedNumber(c *SignedNumberContext)

	// ExitRexCommand is called when exiting the rexCommand production.
	ExitRexCommand(c *RexCommandContext)

	// ExitRexOption is called when exiting the rexOption production.
	ExitRexOption(c *RexOptionContext)

	// ExitSpathCommand is called when exiting the spathCommand production.
	ExitSpathCommand(c *SpathCommandContext)

	// ExitSpathOption is called when exiting the spathOption production.
	ExitSpathOption(c *SpathOptionContext)

	// ExitLoadjobCommand is called when exiting the loadjobCommand production.
	ExitLoadjobCommand(c *LoadjobCommandContext)

	// ExitMetricsCommand is called when exiting the metricsCommand production.
	ExitMetricsCommand(c *MetricsCommandContext)

	// ExitMetricsAggregates is called when exiting the metricsAggregates production.
	ExitMetricsAggregates(c *MetricsAggregatesContext)

	// ExitMetricsOption is called when exiting the metricsOption production.
	ExitMetricsOption(c *MetricsOptionContext)

	// ExitTimechartCommand is called when exiting the timechartCommand production.
	ExitTimechartCommand(c *TimechartCommandContext)

	// ExitTimechartExtraAggregate is called when exiting the timechartExtraAggregate production.
	ExitTimechartExtraAggregate(c *TimechartExtraAggregateContext)

	// ExitTimechartOption is called when exiting the timechartOption production.
	ExitTimechartOption(c *TimechartOptionContext)

	// ExitTimechartSplit is called when exiting the timechartSplit production.
	ExitTimechartSplit(c *TimechartSplitContext)

	// ExitTimechartSplitOption is called when exiting the timechartSplitOption production.
	ExitTimechartSplitOption(c *TimechartSplitOptionContext)

	// ExitTimewrapCommand is called when exiting the timewrapCommand production.
	ExitTimewrapCommand(c *TimewrapCommandContext)

	// ExitMakemvCommand is called when exiting the makemvCommand production.
	ExitMakemvCommand(c *MakemvCommandContext)

	// ExitMvexpandCommand is called when exiting the mvexpandCommand production.
	ExitMvexpandCommand(c *MvexpandCommandContext)

	// ExitMvcombineCommand is called when exiting the mvcombineCommand production.
	ExitMvcombineCommand(c *MvcombineCommandContext)

	// ExitFillnullCommand is called when exiting the fillnullCommand production.
	ExitFillnullCommand(c *FillnullCommandContext)

	// ExitEmbeddedCommand is called when exiting the embeddedCommand production.
	ExitEmbeddedCommand(c *EmbeddedCommandContext)

	// ExitEmbeddedText is called when exiting the embeddedText production.
	ExitEmbeddedText(c *EmbeddedTextContext)

	// ExitModuleSuffix is called when exiting the moduleSuffix production.
	ExitModuleSuffix(c *ModuleSuffixContext)

	// ExitModuleDeclaration is called when exiting the moduleDeclaration production.
	ExitModuleDeclaration(c *ModuleDeclarationContext)

	// ExitSearchCommand is called when exiting the searchCommand production.
	ExitSearchCommand(c *SearchCommandContext)

	// ExitImplicitSearch is called when exiting the implicitSearch production.
	ExitImplicitSearch(c *ImplicitSearchContext)

	// ExitSearchExpression is called when exiting the searchExpression production.
	ExitSearchExpression(c *SearchExpressionContext)

	// ExitSearchXor is called when exiting the searchXor production.
	ExitSearchXor(c *SearchXorContext)

	// ExitSearchAnd is called when exiting the searchAnd production.
	ExitSearchAnd(c *SearchAndContext)

	// ExitSearchOr is called when exiting the searchOr production.
	ExitSearchOr(c *SearchOrContext)

	// ExitSearchNot is called when exiting the searchNot production.
	ExitSearchNot(c *SearchNotContext)

	// ExitSearchAtom is called when exiting the searchAtom production.
	ExitSearchAtom(c *SearchAtomContext)

	// ExitSearchValue is called when exiting the searchValue production.
	ExitSearchValue(c *SearchValueContext)

	// ExitSearchWordLiteral is called when exiting the searchWordLiteral production.
	ExitSearchWordLiteral(c *SearchWordLiteralContext)

	// ExitSearchSignedNumber is called when exiting the searchSignedNumber production.
	ExitSearchSignedNumber(c *SearchSignedNumberContext)

	// ExitSearchUnprovedLiteral is called when exiting the searchUnprovedLiteral production.
	ExitSearchUnprovedLiteral(c *SearchUnprovedLiteralContext)

	// ExitSearchBareValue is called when exiting the searchBareValue production.
	ExitSearchBareValue(c *SearchBareValueContext)

	// ExitSearchDirective is called when exiting the searchDirective production.
	ExitSearchDirective(c *SearchDirectiveContext)

	// ExitSearchTimeModifier is called when exiting the searchTimeModifier production.
	ExitSearchTimeModifier(c *SearchTimeModifierContext)

	// ExitTimeModifierKey is called when exiting the timeModifierKey production.
	ExitTimeModifierKey(c *TimeModifierKeyContext)

	// ExitTimeModifierValue is called when exiting the timeModifierValue production.
	ExitTimeModifierValue(c *TimeModifierValueContext)

	// ExitRelativeTime is called when exiting the relativeTime production.
	ExitRelativeTime(c *RelativeTimeContext)

	// ExitExpression is called when exiting the expression production.
	ExitExpression(c *ExpressionContext)

	// ExitXorExpression is called when exiting the xorExpression production.
	ExitXorExpression(c *XorExpressionContext)

	// ExitOrExpression is called when exiting the orExpression production.
	ExitOrExpression(c *OrExpressionContext)

	// ExitAndExpression is called when exiting the andExpression production.
	ExitAndExpression(c *AndExpressionContext)

	// ExitNotExpression is called when exiting the notExpression production.
	ExitNotExpression(c *NotExpressionContext)

	// ExitPredicate is called when exiting the predicate production.
	ExitPredicate(c *PredicateContext)

	// ExitLogicalAnd is called when exiting the logicalAnd production.
	ExitLogicalAnd(c *LogicalAndContext)

	// ExitLogicalOr is called when exiting the logicalOr production.
	ExitLogicalOr(c *LogicalOrContext)

	// ExitLogicalXor is called when exiting the logicalXor production.
	ExitLogicalXor(c *LogicalXorContext)

	// ExitLogicalNot is called when exiting the logicalNot production.
	ExitLogicalNot(c *LogicalNotContext)

	// ExitBetweenOperator is called when exiting the betweenOperator production.
	ExitBetweenOperator(c *BetweenOperatorContext)

	// ExitBetweenConjunction is called when exiting the betweenConjunction production.
	ExitBetweenConjunction(c *BetweenConjunctionContext)

	// ExitComparison is called when exiting the comparison production.
	ExitComparison(c *ComparisonContext)

	// ExitAdditive is called when exiting the additive production.
	ExitAdditive(c *AdditiveContext)

	// ExitMultiplicative is called when exiting the multiplicative production.
	ExitMultiplicative(c *MultiplicativeContext)

	// ExitUnary is called when exiting the unary production.
	ExitUnary(c *UnaryContext)

	// ExitAccess is called when exiting the access production.
	ExitAccess(c *AccessContext)

	// ExitAccessPart is called when exiting the accessPart production.
	ExitAccessPart(c *AccessPartContext)

	// ExitPrimary is called when exiting the primary production.
	ExitPrimary(c *PrimaryContext)

	// ExitCall is called when exiting the call production.
	ExitCall(c *CallContext)

	// ExitArguments is called when exiting the arguments production.
	ExitArguments(c *ArgumentsContext)

	// ExitNamedArgument is called when exiting the namedArgument production.
	ExitNamedArgument(c *NamedArgumentContext)

	// ExitLiteral is called when exiting the literal production.
	ExitLiteral(c *LiteralContext)

	// ExitStringLiteral is called when exiting the stringLiteral production.
	ExitStringLiteral(c *StringLiteralContext)

	// ExitQuotedName is called when exiting the quotedName production.
	ExitQuotedName(c *QuotedNameContext)

	// ExitFieldTemplate is called when exiting the fieldTemplate production.
	ExitFieldTemplate(c *FieldTemplateContext)

	// ExitFieldName is called when exiting the fieldName production.
	ExitFieldName(c *FieldNameContext)

	// ExitIdentifier is called when exiting the identifier production.
	ExitIdentifier(c *IdentifierContext)

	// ExitSqlKeyword is called when exiting the sqlKeyword production.
	ExitSqlKeyword(c *SqlKeywordContext)

	// ExitPipelineKeyword is called when exiting the pipelineKeyword production.
	ExitPipelineKeyword(c *PipelineKeywordContext)

	// ExitArray is called when exiting the array production.
	ExitArray(c *ArrayContext)

	// ExitObject is called when exiting the object production.
	ExitObject(c *ObjectContext)

	// ExitObjectEntry is called when exiting the objectEntry production.
	ExitObjectEntry(c *ObjectEntryContext)

	// ExitObjectKey is called when exiting the objectKey production.
	ExitObjectKey(c *ObjectKeyContext)

	// ExitSearchLiteral is called when exiting the searchLiteral production.
	ExitSearchLiteral(c *SearchLiteralContext)

	// ExitLambdaExpression is called when exiting the lambdaExpression production.
	ExitLambdaExpression(c *LambdaExpressionContext)

	// ExitLambdaParameter is called when exiting the lambdaParameter production.
	ExitLambdaParameter(c *LambdaParameterContext)

	// ExitLambdaBlock is called when exiting the lambdaBlock production.
	ExitLambdaBlock(c *LambdaBlockContext)
}
