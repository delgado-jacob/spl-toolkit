// Code generated from grammar/SPL2Parser.g4 by ANTLR 4.13.2. DO NOT EDIT.

package spl2 // SPL2Parser
import "github.com/antlr4-go/antlr/v4"

// BaseSPL2ParserListener is a complete listener for a parse tree produced by SPL2Parser.
type BaseSPL2ParserListener struct{}

var _ SPL2ParserListener = &BaseSPL2ParserListener{}

// VisitTerminal is called when a terminal node is visited.
func (s *BaseSPL2ParserListener) VisitTerminal(node antlr.TerminalNode) {}

// VisitErrorNode is called when an error node is visited.
func (s *BaseSPL2ParserListener) VisitErrorNode(node antlr.ErrorNode) {}

// EnterEveryRule is called when any rule is entered.
func (s *BaseSPL2ParserListener) EnterEveryRule(ctx antlr.ParserRuleContext) {}

// ExitEveryRule is called when any rule is exited.
func (s *BaseSPL2ParserListener) ExitEveryRule(ctx antlr.ParserRuleContext) {}

// EnterQuery is called when production query is entered.
func (s *BaseSPL2ParserListener) EnterQuery(ctx *QueryContext) {}

// ExitQuery is called when production query is exited.
func (s *BaseSPL2ParserListener) ExitQuery(ctx *QueryContext) {}

// EnterPipeline is called when production pipeline is entered.
func (s *BaseSPL2ParserListener) EnterPipeline(ctx *PipelineContext) {}

// ExitPipeline is called when production pipeline is exited.
func (s *BaseSPL2ParserListener) ExitPipeline(ctx *PipelineContext) {}

// EnterStart is called when production start is entered.
func (s *BaseSPL2ParserListener) EnterStart(ctx *StartContext) {}

// ExitStart is called when production start is exited.
func (s *BaseSPL2ParserListener) ExitStart(ctx *StartContext) {}

// EnterCommand is called when production command is entered.
func (s *BaseSPL2ParserListener) EnterCommand(ctx *CommandContext) {}

// ExitCommand is called when production command is exited.
func (s *BaseSPL2ParserListener) ExitCommand(ctx *CommandContext) {}

// EnterFromCommand is called when production fromCommand is entered.
func (s *BaseSPL2ParserListener) EnterFromCommand(ctx *FromCommandContext) {}

// ExitFromCommand is called when production fromCommand is exited.
func (s *BaseSPL2ParserListener) ExitFromCommand(ctx *FromCommandContext) {}

// EnterSelectCommand is called when production selectCommand is entered.
func (s *BaseSPL2ParserListener) EnterSelectCommand(ctx *SelectCommandContext) {}

// ExitSelectCommand is called when production selectCommand is exited.
func (s *BaseSPL2ParserListener) ExitSelectCommand(ctx *SelectCommandContext) {}

// EnterSqlFromClause is called when production sqlFromClause is entered.
func (s *BaseSPL2ParserListener) EnterSqlFromClause(ctx *SqlFromClauseContext) {}

// ExitSqlFromClause is called when production sqlFromClause is exited.
func (s *BaseSPL2ParserListener) ExitSqlFromClause(ctx *SqlFromClauseContext) {}

// EnterSourceAlias is called when production sourceAlias is entered.
func (s *BaseSPL2ParserListener) EnterSourceAlias(ctx *SourceAliasContext) {}

// ExitSourceAlias is called when production sourceAlias is exited.
func (s *BaseSPL2ParserListener) ExitSourceAlias(ctx *SourceAliasContext) {}

// EnterSqlJoinClause is called when production sqlJoinClause is entered.
func (s *BaseSPL2ParserListener) EnterSqlJoinClause(ctx *SqlJoinClauseContext) {}

// ExitSqlJoinClause is called when production sqlJoinClause is exited.
func (s *BaseSPL2ParserListener) ExitSqlJoinClause(ctx *SqlJoinClauseContext) {}

// EnterSqlJoinPredicate is called when production sqlJoinPredicate is entered.
func (s *BaseSPL2ParserListener) EnterSqlJoinPredicate(ctx *SqlJoinPredicateContext) {}

// ExitSqlJoinPredicate is called when production sqlJoinPredicate is exited.
func (s *BaseSPL2ParserListener) ExitSqlJoinPredicate(ctx *SqlJoinPredicateContext) {}

// EnterSqlJoinEquality is called when production sqlJoinEquality is entered.
func (s *BaseSPL2ParserListener) EnterSqlJoinEquality(ctx *SqlJoinEqualityContext) {}

// ExitSqlJoinEquality is called when production sqlJoinEquality is exited.
func (s *BaseSPL2ParserListener) ExitSqlJoinEquality(ctx *SqlJoinEqualityContext) {}

// EnterSqlJoinField is called when production sqlJoinField is entered.
func (s *BaseSPL2ParserListener) EnterSqlJoinField(ctx *SqlJoinFieldContext) {}

// ExitSqlJoinField is called when production sqlJoinField is exited.
func (s *BaseSPL2ParserListener) ExitSqlJoinField(ctx *SqlJoinFieldContext) {}

// EnterSqlSelectClause is called when production sqlSelectClause is entered.
func (s *BaseSPL2ParserListener) EnterSqlSelectClause(ctx *SqlSelectClauseContext) {}

// ExitSqlSelectClause is called when production sqlSelectClause is exited.
func (s *BaseSPL2ParserListener) ExitSqlSelectClause(ctx *SqlSelectClauseContext) {}

// EnterProjection is called when production projection is entered.
func (s *BaseSPL2ParserListener) EnterProjection(ctx *ProjectionContext) {}

// ExitProjection is called when production projection is exited.
func (s *BaseSPL2ParserListener) ExitProjection(ctx *ProjectionContext) {}

// EnterProjectionAlias is called when production projectionAlias is entered.
func (s *BaseSPL2ParserListener) EnterProjectionAlias(ctx *ProjectionAliasContext) {}

// ExitProjectionAlias is called when production projectionAlias is exited.
func (s *BaseSPL2ParserListener) ExitProjectionAlias(ctx *ProjectionAliasContext) {}

// EnterSqlWhereClause is called when production sqlWhereClause is entered.
func (s *BaseSPL2ParserListener) EnterSqlWhereClause(ctx *SqlWhereClauseContext) {}

// ExitSqlWhereClause is called when production sqlWhereClause is exited.
func (s *BaseSPL2ParserListener) ExitSqlWhereClause(ctx *SqlWhereClauseContext) {}

// EnterSqlHavingClause is called when production sqlHavingClause is entered.
func (s *BaseSPL2ParserListener) EnterSqlHavingClause(ctx *SqlHavingClauseContext) {}

// ExitSqlHavingClause is called when production sqlHavingClause is exited.
func (s *BaseSPL2ParserListener) ExitSqlHavingClause(ctx *SqlHavingClauseContext) {}

// EnterSqlPredicate is called when production sqlPredicate is entered.
func (s *BaseSPL2ParserListener) EnterSqlPredicate(ctx *SqlPredicateContext) {}

// ExitSqlPredicate is called when production sqlPredicate is exited.
func (s *BaseSPL2ParserListener) ExitSqlPredicate(ctx *SqlPredicateContext) {}

// EnterSqlGroupClause is called when production sqlGroupClause is entered.
func (s *BaseSPL2ParserListener) EnterSqlGroupClause(ctx *SqlGroupClauseContext) {}

// ExitSqlGroupClause is called when production sqlGroupClause is exited.
func (s *BaseSPL2ParserListener) ExitSqlGroupClause(ctx *SqlGroupClauseContext) {}

// EnterSqlBy is called when production sqlBy is entered.
func (s *BaseSPL2ParserListener) EnterSqlBy(ctx *SqlByContext) {}

// ExitSqlBy is called when production sqlBy is exited.
func (s *BaseSPL2ParserListener) ExitSqlBy(ctx *SqlByContext) {}

// EnterSqlGroupKey is called when production sqlGroupKey is entered.
func (s *BaseSPL2ParserListener) EnterSqlGroupKey(ctx *SqlGroupKeyContext) {}

// ExitSqlGroupKey is called when production sqlGroupKey is exited.
func (s *BaseSPL2ParserListener) ExitSqlGroupKey(ctx *SqlGroupKeyContext) {}

// EnterSqlSpanCall is called when production sqlSpanCall is entered.
func (s *BaseSPL2ParserListener) EnterSqlSpanCall(ctx *SqlSpanCallContext) {}

// ExitSqlSpanCall is called when production sqlSpanCall is exited.
func (s *BaseSPL2ParserListener) ExitSqlSpanCall(ctx *SqlSpanCallContext) {}

// EnterSqlSpanAssignment is called when production sqlSpanAssignment is entered.
func (s *BaseSPL2ParserListener) EnterSqlSpanAssignment(ctx *SqlSpanAssignmentContext) {}

// ExitSqlSpanAssignment is called when production sqlSpanAssignment is exited.
func (s *BaseSPL2ParserListener) ExitSqlSpanAssignment(ctx *SqlSpanAssignmentContext) {}

// EnterSqlUnparenthesizedSpan is called when production sqlUnparenthesizedSpan is entered.
func (s *BaseSPL2ParserListener) EnterSqlUnparenthesizedSpan(ctx *SqlUnparenthesizedSpanContext) {}

// ExitSqlUnparenthesizedSpan is called when production sqlUnparenthesizedSpan is exited.
func (s *BaseSPL2ParserListener) ExitSqlUnparenthesizedSpan(ctx *SqlUnparenthesizedSpanContext) {}

// EnterSqlOrderClause is called when production sqlOrderClause is entered.
func (s *BaseSPL2ParserListener) EnterSqlOrderClause(ctx *SqlOrderClauseContext) {}

// ExitSqlOrderClause is called when production sqlOrderClause is exited.
func (s *BaseSPL2ParserListener) ExitSqlOrderClause(ctx *SqlOrderClauseContext) {}

// EnterSqlOrderTerm is called when production sqlOrderTerm is entered.
func (s *BaseSPL2ParserListener) EnterSqlOrderTerm(ctx *SqlOrderTermContext) {}

// ExitSqlOrderTerm is called when production sqlOrderTerm is exited.
func (s *BaseSPL2ParserListener) ExitSqlOrderTerm(ctx *SqlOrderTermContext) {}

// EnterSqlDirection is called when production sqlDirection is entered.
func (s *BaseSPL2ParserListener) EnterSqlDirection(ctx *SqlDirectionContext) {}

// ExitSqlDirection is called when production sqlDirection is exited.
func (s *BaseSPL2ParserListener) ExitSqlDirection(ctx *SqlDirectionContext) {}

// EnterSqlLimitClause is called when production sqlLimitClause is entered.
func (s *BaseSPL2ParserListener) EnterSqlLimitClause(ctx *SqlLimitClauseContext) {}

// ExitSqlLimitClause is called when production sqlLimitClause is exited.
func (s *BaseSPL2ParserListener) ExitSqlLimitClause(ctx *SqlLimitClauseContext) {}

// EnterSqlOffsetClause is called when production sqlOffsetClause is entered.
func (s *BaseSPL2ParserListener) EnterSqlOffsetClause(ctx *SqlOffsetClauseContext) {}

// ExitSqlOffsetClause is called when production sqlOffsetClause is exited.
func (s *BaseSPL2ParserListener) ExitSqlOffsetClause(ctx *SqlOffsetClauseContext) {}

// EnterMultilineSqlSelectClause is called when production multilineSqlSelectClause is entered.
func (s *BaseSPL2ParserListener) EnterMultilineSqlSelectClause(ctx *MultilineSqlSelectClauseContext) {
}

// ExitMultilineSqlSelectClause is called when production multilineSqlSelectClause is exited.
func (s *BaseSPL2ParserListener) ExitMultilineSqlSelectClause(ctx *MultilineSqlSelectClauseContext) {}

// EnterMultilineSqlProjection is called when production multilineSqlProjection is entered.
func (s *BaseSPL2ParserListener) EnterMultilineSqlProjection(ctx *MultilineSqlProjectionContext) {}

// ExitMultilineSqlProjection is called when production multilineSqlProjection is exited.
func (s *BaseSPL2ParserListener) ExitMultilineSqlProjection(ctx *MultilineSqlProjectionContext) {}

// EnterMultilineSqlFromClause is called when production multilineSqlFromClause is entered.
func (s *BaseSPL2ParserListener) EnterMultilineSqlFromClause(ctx *MultilineSqlFromClauseContext) {}

// ExitMultilineSqlFromClause is called when production multilineSqlFromClause is exited.
func (s *BaseSPL2ParserListener) ExitMultilineSqlFromClause(ctx *MultilineSqlFromClauseContext) {}

// EnterMultilineSqlWhereClause is called when production multilineSqlWhereClause is entered.
func (s *BaseSPL2ParserListener) EnterMultilineSqlWhereClause(ctx *MultilineSqlWhereClauseContext) {}

// ExitMultilineSqlWhereClause is called when production multilineSqlWhereClause is exited.
func (s *BaseSPL2ParserListener) ExitMultilineSqlWhereClause(ctx *MultilineSqlWhereClauseContext) {}

// EnterMultilineSqlPredicate is called when production multilineSqlPredicate is entered.
func (s *BaseSPL2ParserListener) EnterMultilineSqlPredicate(ctx *MultilineSqlPredicateContext) {}

// ExitMultilineSqlPredicate is called when production multilineSqlPredicate is exited.
func (s *BaseSPL2ParserListener) ExitMultilineSqlPredicate(ctx *MultilineSqlPredicateContext) {}

// EnterMultilineSqlGroupClause is called when production multilineSqlGroupClause is entered.
func (s *BaseSPL2ParserListener) EnterMultilineSqlGroupClause(ctx *MultilineSqlGroupClauseContext) {}

// ExitMultilineSqlGroupClause is called when production multilineSqlGroupClause is exited.
func (s *BaseSPL2ParserListener) ExitMultilineSqlGroupClause(ctx *MultilineSqlGroupClauseContext) {}

// EnterMultilineSqlGroupKey is called when production multilineSqlGroupKey is entered.
func (s *BaseSPL2ParserListener) EnterMultilineSqlGroupKey(ctx *MultilineSqlGroupKeyContext) {}

// ExitMultilineSqlGroupKey is called when production multilineSqlGroupKey is exited.
func (s *BaseSPL2ParserListener) ExitMultilineSqlGroupKey(ctx *MultilineSqlGroupKeyContext) {}

// EnterMultilineSqlSpanCall is called when production multilineSqlSpanCall is entered.
func (s *BaseSPL2ParserListener) EnterMultilineSqlSpanCall(ctx *MultilineSqlSpanCallContext) {}

// ExitMultilineSqlSpanCall is called when production multilineSqlSpanCall is exited.
func (s *BaseSPL2ParserListener) ExitMultilineSqlSpanCall(ctx *MultilineSqlSpanCallContext) {}

// EnterMultilineSqlOrderClause is called when production multilineSqlOrderClause is entered.
func (s *BaseSPL2ParserListener) EnterMultilineSqlOrderClause(ctx *MultilineSqlOrderClauseContext) {}

// ExitMultilineSqlOrderClause is called when production multilineSqlOrderClause is exited.
func (s *BaseSPL2ParserListener) ExitMultilineSqlOrderClause(ctx *MultilineSqlOrderClauseContext) {}

// EnterMultilineSqlOrderTerm is called when production multilineSqlOrderTerm is entered.
func (s *BaseSPL2ParserListener) EnterMultilineSqlOrderTerm(ctx *MultilineSqlOrderTermContext) {}

// ExitMultilineSqlOrderTerm is called when production multilineSqlOrderTerm is exited.
func (s *BaseSPL2ParserListener) ExitMultilineSqlOrderTerm(ctx *MultilineSqlOrderTermContext) {}

// EnterExistsPredicate is called when production existsPredicate is entered.
func (s *BaseSPL2ParserListener) EnterExistsPredicate(ctx *ExistsPredicateContext) {}

// ExitExistsPredicate is called when production existsPredicate is exited.
func (s *BaseSPL2ParserListener) ExitExistsPredicate(ctx *ExistsPredicateContext) {}

// EnterDataset is called when production dataset is entered.
func (s *BaseSPL2ParserListener) EnterDataset(ctx *DatasetContext) {}

// ExitDataset is called when production dataset is exited.
func (s *BaseSPL2ParserListener) ExitDataset(ctx *DatasetContext) {}

// EnterDottedDataset is called when production dottedDataset is entered.
func (s *BaseSPL2ParserListener) EnterDottedDataset(ctx *DottedDatasetContext) {}

// ExitDottedDataset is called when production dottedDataset is exited.
func (s *BaseSPL2ParserListener) ExitDottedDataset(ctx *DottedDatasetContext) {}

// EnterDatasetPath is called when production datasetPath is entered.
func (s *BaseSPL2ParserListener) EnterDatasetPath(ctx *DatasetPathContext) {}

// ExitDatasetPath is called when production datasetPath is exited.
func (s *BaseSPL2ParserListener) ExitDatasetPath(ctx *DatasetPathContext) {}

// EnterDatasetParameter is called when production datasetParameter is entered.
func (s *BaseSPL2ParserListener) EnterDatasetParameter(ctx *DatasetParameterContext) {}

// ExitDatasetParameter is called when production datasetParameter is exited.
func (s *BaseSPL2ParserListener) ExitDatasetParameter(ctx *DatasetParameterContext) {}

// EnterStaticDatasetDescriptor is called when production staticDatasetDescriptor is entered.
func (s *BaseSPL2ParserListener) EnterStaticDatasetDescriptor(ctx *StaticDatasetDescriptorContext) {}

// ExitStaticDatasetDescriptor is called when production staticDatasetDescriptor is exited.
func (s *BaseSPL2ParserListener) ExitStaticDatasetDescriptor(ctx *StaticDatasetDescriptorContext) {}

// EnterDescriptorKindKey is called when production descriptorKindKey is entered.
func (s *BaseSPL2ParserListener) EnterDescriptorKindKey(ctx *DescriptorKindKeyContext) {}

// ExitDescriptorKindKey is called when production descriptorKindKey is exited.
func (s *BaseSPL2ParserListener) ExitDescriptorKindKey(ctx *DescriptorKindKeyContext) {}

// EnterDescriptorPropertiesKey is called when production descriptorPropertiesKey is entered.
func (s *BaseSPL2ParserListener) EnterDescriptorPropertiesKey(ctx *DescriptorPropertiesKeyContext) {}

// ExitDescriptorPropertiesKey is called when production descriptorPropertiesKey is exited.
func (s *BaseSPL2ParserListener) ExitDescriptorPropertiesKey(ctx *DescriptorPropertiesKeyContext) {}

// EnterDescriptorProperties is called when production descriptorProperties is entered.
func (s *BaseSPL2ParserListener) EnterDescriptorProperties(ctx *DescriptorPropertiesContext) {}

// ExitDescriptorProperties is called when production descriptorProperties is exited.
func (s *BaseSPL2ParserListener) ExitDescriptorProperties(ctx *DescriptorPropertiesContext) {}

// EnterDescriptorProperty is called when production descriptorProperty is entered.
func (s *BaseSPL2ParserListener) EnterDescriptorProperty(ctx *DescriptorPropertyContext) {}

// ExitDescriptorProperty is called when production descriptorProperty is exited.
func (s *BaseSPL2ParserListener) ExitDescriptorProperty(ctx *DescriptorPropertyContext) {}

// EnterJsonObjectKey is called when production jsonObjectKey is entered.
func (s *BaseSPL2ParserListener) EnterJsonObjectKey(ctx *JsonObjectKeyContext) {}

// ExitJsonObjectKey is called when production jsonObjectKey is exited.
func (s *BaseSPL2ParserListener) ExitJsonObjectKey(ctx *JsonObjectKeyContext) {}

// EnterJsonLiteral is called when production jsonLiteral is entered.
func (s *BaseSPL2ParserListener) EnterJsonLiteral(ctx *JsonLiteralContext) {}

// ExitJsonLiteral is called when production jsonLiteral is exited.
func (s *BaseSPL2ParserListener) ExitJsonLiteral(ctx *JsonLiteralContext) {}

// EnterJsonStringLiteral is called when production jsonStringLiteral is entered.
func (s *BaseSPL2ParserListener) EnterJsonStringLiteral(ctx *JsonStringLiteralContext) {}

// ExitJsonStringLiteral is called when production jsonStringLiteral is exited.
func (s *BaseSPL2ParserListener) ExitJsonStringLiteral(ctx *JsonStringLiteralContext) {}

// EnterJsonArray is called when production jsonArray is entered.
func (s *BaseSPL2ParserListener) EnterJsonArray(ctx *JsonArrayContext) {}

// ExitJsonArray is called when production jsonArray is exited.
func (s *BaseSPL2ParserListener) ExitJsonArray(ctx *JsonArrayContext) {}

// EnterJsonObject is called when production jsonObject is entered.
func (s *BaseSPL2ParserListener) EnterJsonObject(ctx *JsonObjectContext) {}

// ExitJsonObject is called when production jsonObject is exited.
func (s *BaseSPL2ParserListener) ExitJsonObject(ctx *JsonObjectContext) {}

// EnterGenerator is called when production generator is entered.
func (s *BaseSPL2ParserListener) EnterGenerator(ctx *GeneratorContext) {}

// ExitGenerator is called when production generator is exited.
func (s *BaseSPL2ParserListener) ExitGenerator(ctx *GeneratorContext) {}

// EnterEvalCommand is called when production evalCommand is entered.
func (s *BaseSPL2ParserListener) EnterEvalCommand(ctx *EvalCommandContext) {}

// ExitEvalCommand is called when production evalCommand is exited.
func (s *BaseSPL2ParserListener) ExitEvalCommand(ctx *EvalCommandContext) {}

// EnterAssignment is called when production assignment is entered.
func (s *BaseSPL2ParserListener) EnterAssignment(ctx *AssignmentContext) {}

// ExitAssignment is called when production assignment is exited.
func (s *BaseSPL2ParserListener) ExitAssignment(ctx *AssignmentContext) {}

// EnterWhereCommand is called when production whereCommand is entered.
func (s *BaseSPL2ParserListener) EnterWhereCommand(ctx *WhereCommandContext) {}

// ExitWhereCommand is called when production whereCommand is exited.
func (s *BaseSPL2ParserListener) ExitWhereCommand(ctx *WhereCommandContext) {}

// EnterFieldsCommand is called when production fieldsCommand is entered.
func (s *BaseSPL2ParserListener) EnterFieldsCommand(ctx *FieldsCommandContext) {}

// ExitFieldsCommand is called when production fieldsCommand is exited.
func (s *BaseSPL2ParserListener) ExitFieldsCommand(ctx *FieldsCommandContext) {}

// EnterFieldSelection is called when production fieldSelection is entered.
func (s *BaseSPL2ParserListener) EnterFieldSelection(ctx *FieldSelectionContext) {}

// ExitFieldSelection is called when production fieldSelection is exited.
func (s *BaseSPL2ParserListener) ExitFieldSelection(ctx *FieldSelectionContext) {}

// EnterFieldSelector is called when production fieldSelector is entered.
func (s *BaseSPL2ParserListener) EnterFieldSelector(ctx *FieldSelectorContext) {}

// ExitFieldSelector is called when production fieldSelector is exited.
func (s *BaseSPL2ParserListener) ExitFieldSelector(ctx *FieldSelectorContext) {}

// EnterStructuralFieldSelector is called when production structuralFieldSelector is entered.
func (s *BaseSPL2ParserListener) EnterStructuralFieldSelector(ctx *StructuralFieldSelectorContext) {}

// ExitStructuralFieldSelector is called when production structuralFieldSelector is exited.
func (s *BaseSPL2ParserListener) ExitStructuralFieldSelector(ctx *StructuralFieldSelectorContext) {}

// EnterFieldPath is called when production fieldPath is entered.
func (s *BaseSPL2ParserListener) EnterFieldPath(ctx *FieldPathContext) {}

// ExitFieldPath is called when production fieldPath is exited.
func (s *BaseSPL2ParserListener) ExitFieldPath(ctx *FieldPathContext) {}

// EnterTableCommand is called when production tableCommand is entered.
func (s *BaseSPL2ParserListener) EnterTableCommand(ctx *TableCommandContext) {}

// ExitTableCommand is called when production tableCommand is exited.
func (s *BaseSPL2ParserListener) ExitTableCommand(ctx *TableCommandContext) {}

// EnterTableField is called when production tableField is entered.
func (s *BaseSPL2ParserListener) EnterTableField(ctx *TableFieldContext) {}

// ExitTableField is called when production tableField is exited.
func (s *BaseSPL2ParserListener) ExitTableField(ctx *TableFieldContext) {}

// EnterRenameCommand is called when production renameCommand is entered.
func (s *BaseSPL2ParserListener) EnterRenameCommand(ctx *RenameCommandContext) {}

// ExitRenameCommand is called when production renameCommand is exited.
func (s *BaseSPL2ParserListener) ExitRenameCommand(ctx *RenameCommandContext) {}

// EnterRenamePair is called when production renamePair is entered.
func (s *BaseSPL2ParserListener) EnterRenamePair(ctx *RenamePairContext) {}

// ExitRenamePair is called when production renamePair is exited.
func (s *BaseSPL2ParserListener) ExitRenamePair(ctx *RenamePairContext) {}

// EnterRenameSource is called when production renameSource is entered.
func (s *BaseSPL2ParserListener) EnterRenameSource(ctx *RenameSourceContext) {}

// ExitRenameSource is called when production renameSource is exited.
func (s *BaseSPL2ParserListener) ExitRenameSource(ctx *RenameSourceContext) {}

// EnterRenameTarget is called when production renameTarget is entered.
func (s *BaseSPL2ParserListener) EnterRenameTarget(ctx *RenameTargetContext) {}

// ExitRenameTarget is called when production renameTarget is exited.
func (s *BaseSPL2ParserListener) ExitRenameTarget(ctx *RenameTargetContext) {}

// EnterAliasKeyword is called when production aliasKeyword is entered.
func (s *BaseSPL2ParserListener) EnterAliasKeyword(ctx *AliasKeywordContext) {}

// ExitAliasKeyword is called when production aliasKeyword is exited.
func (s *BaseSPL2ParserListener) ExitAliasKeyword(ctx *AliasKeywordContext) {}

// EnterStatsCommand is called when production statsCommand is entered.
func (s *BaseSPL2ParserListener) EnterStatsCommand(ctx *StatsCommandContext) {}

// ExitStatsCommand is called when production statsCommand is exited.
func (s *BaseSPL2ParserListener) ExitStatsCommand(ctx *StatsCommandContext) {}

// EnterStatsOption is called when production statsOption is entered.
func (s *BaseSPL2ParserListener) EnterStatsOption(ctx *StatsOptionContext) {}

// ExitStatsOption is called when production statsOption is exited.
func (s *BaseSPL2ParserListener) ExitStatsOption(ctx *StatsOptionContext) {}

// EnterAllnumOption is called when production allnumOption is entered.
func (s *BaseSPL2ParserListener) EnterAllnumOption(ctx *AllnumOptionContext) {}

// ExitAllnumOption is called when production allnumOption is exited.
func (s *BaseSPL2ParserListener) ExitAllnumOption(ctx *AllnumOptionContext) {}

// EnterDelimOption is called when production delimOption is entered.
func (s *BaseSPL2ParserListener) EnterDelimOption(ctx *DelimOptionContext) {}

// ExitDelimOption is called when production delimOption is exited.
func (s *BaseSPL2ParserListener) ExitDelimOption(ctx *DelimOptionContext) {}

// EnterPartitionsOption is called when production partitionsOption is entered.
func (s *BaseSPL2ParserListener) EnterPartitionsOption(ctx *PartitionsOptionContext) {}

// ExitPartitionsOption is called when production partitionsOption is exited.
func (s *BaseSPL2ParserListener) ExitPartitionsOption(ctx *PartitionsOptionContext) {}

// EnterAggregate is called when production aggregate is entered.
func (s *BaseSPL2ParserListener) EnterAggregate(ctx *AggregateContext) {}

// ExitAggregate is called when production aggregate is exited.
func (s *BaseSPL2ParserListener) ExitAggregate(ctx *AggregateContext) {}

// EnterAggregateAlias is called when production aggregateAlias is entered.
func (s *BaseSPL2ParserListener) EnterAggregateAlias(ctx *AggregateAliasContext) {}

// ExitAggregateAlias is called when production aggregateAlias is exited.
func (s *BaseSPL2ParserListener) ExitAggregateAlias(ctx *AggregateAliasContext) {}

// EnterAggregateGroup is called when production aggregateGroup is entered.
func (s *BaseSPL2ParserListener) EnterAggregateGroup(ctx *AggregateGroupContext) {}

// ExitAggregateGroup is called when production aggregateGroup is exited.
func (s *BaseSPL2ParserListener) ExitAggregateGroup(ctx *AggregateGroupContext) {}

// EnterSelectedAggregateGroup is called when production selectedAggregateGroup is entered.
func (s *BaseSPL2ParserListener) EnterSelectedAggregateGroup(ctx *SelectedAggregateGroupContext) {}

// ExitSelectedAggregateGroup is called when production selectedAggregateGroup is exited.
func (s *BaseSPL2ParserListener) ExitSelectedAggregateGroup(ctx *SelectedAggregateGroupContext) {}

// EnterSelectedGroupTerm is called when production selectedGroupTerm is entered.
func (s *BaseSPL2ParserListener) EnterSelectedGroupTerm(ctx *SelectedGroupTermContext) {}

// ExitSelectedGroupTerm is called when production selectedGroupTerm is exited.
func (s *BaseSPL2ParserListener) ExitSelectedGroupTerm(ctx *SelectedGroupTermContext) {}

// EnterSelectedSpanGroup is called when production selectedSpanGroup is entered.
func (s *BaseSPL2ParserListener) EnterSelectedSpanGroup(ctx *SelectedSpanGroupContext) {}

// ExitSelectedSpanGroup is called when production selectedSpanGroup is exited.
func (s *BaseSPL2ParserListener) ExitSelectedSpanGroup(ctx *SelectedSpanGroupContext) {}

// EnterGroupField is called when production groupField is entered.
func (s *BaseSPL2ParserListener) EnterGroupField(ctx *GroupFieldContext) {}

// ExitGroupField is called when production groupField is exited.
func (s *BaseSPL2ParserListener) ExitGroupField(ctx *GroupFieldContext) {}

// EnterGroupSpan is called when production groupSpan is entered.
func (s *BaseSPL2ParserListener) EnterGroupSpan(ctx *GroupSpanContext) {}

// ExitGroupSpan is called when production groupSpan is exited.
func (s *BaseSPL2ParserListener) ExitGroupSpan(ctx *GroupSpanContext) {}

// EnterTimeSpan is called when production timeSpan is entered.
func (s *BaseSPL2ParserListener) EnterTimeSpan(ctx *TimeSpanContext) {}

// ExitTimeSpan is called when production timeSpan is exited.
func (s *BaseSPL2ParserListener) ExitTimeSpan(ctx *TimeSpanContext) {}

// EnterEventstatsCommand is called when production eventstatsCommand is entered.
func (s *BaseSPL2ParserListener) EnterEventstatsCommand(ctx *EventstatsCommandContext) {}

// ExitEventstatsCommand is called when production eventstatsCommand is exited.
func (s *BaseSPL2ParserListener) ExitEventstatsCommand(ctx *EventstatsCommandContext) {}

// EnterStreamstatsCommand is called when production streamstatsCommand is entered.
func (s *BaseSPL2ParserListener) EnterStreamstatsCommand(ctx *StreamstatsCommandContext) {}

// ExitStreamstatsCommand is called when production streamstatsCommand is exited.
func (s *BaseSPL2ParserListener) ExitStreamstatsCommand(ctx *StreamstatsCommandContext) {}

// EnterStreamGroup is called when production streamGroup is entered.
func (s *BaseSPL2ParserListener) EnterStreamGroup(ctx *StreamGroupContext) {}

// ExitStreamGroup is called when production streamGroup is exited.
func (s *BaseSPL2ParserListener) ExitStreamGroup(ctx *StreamGroupContext) {}

// EnterCurrentOption is called when production currentOption is entered.
func (s *BaseSPL2ParserListener) EnterCurrentOption(ctx *CurrentOptionContext) {}

// ExitCurrentOption is called when production currentOption is exited.
func (s *BaseSPL2ParserListener) ExitCurrentOption(ctx *CurrentOptionContext) {}

// EnterWindowOption is called when production windowOption is entered.
func (s *BaseSPL2ParserListener) EnterWindowOption(ctx *WindowOptionContext) {}

// ExitWindowOption is called when production windowOption is exited.
func (s *BaseSPL2ParserListener) ExitWindowOption(ctx *WindowOptionContext) {}

// EnterResetClause is called when production resetClause is entered.
func (s *BaseSPL2ParserListener) EnterResetClause(ctx *ResetClauseContext) {}

// ExitResetClause is called when production resetClause is exited.
func (s *BaseSPL2ParserListener) ExitResetClause(ctx *ResetClauseContext) {}

// EnterResetBefore is called when production resetBefore is entered.
func (s *BaseSPL2ParserListener) EnterResetBefore(ctx *ResetBeforeContext) {}

// ExitResetBefore is called when production resetBefore is exited.
func (s *BaseSPL2ParserListener) ExitResetBefore(ctx *ResetBeforeContext) {}

// EnterResetAfter is called when production resetAfter is entered.
func (s *BaseSPL2ParserListener) EnterResetAfter(ctx *ResetAfterContext) {}

// ExitResetAfter is called when production resetAfter is exited.
func (s *BaseSPL2ParserListener) ExitResetAfter(ctx *ResetAfterContext) {}

// EnterResetOnchange is called when production resetOnchange is entered.
func (s *BaseSPL2ParserListener) EnterResetOnchange(ctx *ResetOnchangeContext) {}

// ExitResetOnchange is called when production resetOnchange is exited.
func (s *BaseSPL2ParserListener) ExitResetOnchange(ctx *ResetOnchangeContext) {}

// EnterStreamPostLayout is called when production streamPostLayout is entered.
func (s *BaseSPL2ParserListener) EnterStreamPostLayout(ctx *StreamPostLayoutContext) {}

// ExitStreamPostLayout is called when production streamPostLayout is exited.
func (s *BaseSPL2ParserListener) ExitStreamPostLayout(ctx *StreamPostLayoutContext) {}

// EnterLookupCommand is called when production lookupCommand is entered.
func (s *BaseSPL2ParserListener) EnterLookupCommand(ctx *LookupCommandContext) {}

// ExitLookupCommand is called when production lookupCommand is exited.
func (s *BaseSPL2ParserListener) ExitLookupCommand(ctx *LookupCommandContext) {}

// EnterLookupDataset is called when production lookupDataset is entered.
func (s *BaseSPL2ParserListener) EnterLookupDataset(ctx *LookupDatasetContext) {}

// ExitLookupDataset is called when production lookupDataset is exited.
func (s *BaseSPL2ParserListener) ExitLookupDataset(ctx *LookupDatasetContext) {}

// EnterLookupMatch is called when production lookupMatch is entered.
func (s *BaseSPL2ParserListener) EnterLookupMatch(ctx *LookupMatchContext) {}

// ExitLookupMatch is called when production lookupMatch is exited.
func (s *BaseSPL2ParserListener) ExitLookupMatch(ctx *LookupMatchContext) {}

// EnterLookupOutputClause is called when production lookupOutputClause is entered.
func (s *BaseSPL2ParserListener) EnterLookupOutputClause(ctx *LookupOutputClauseContext) {}

// ExitLookupOutputClause is called when production lookupOutputClause is exited.
func (s *BaseSPL2ParserListener) ExitLookupOutputClause(ctx *LookupOutputClauseContext) {}

// EnterLookupOutput is called when production lookupOutput is entered.
func (s *BaseSPL2ParserListener) EnterLookupOutput(ctx *LookupOutputContext) {}

// ExitLookupOutput is called when production lookupOutput is exited.
func (s *BaseSPL2ParserListener) ExitLookupOutput(ctx *LookupOutputContext) {}

// EnterLookupColumn is called when production lookupColumn is entered.
func (s *BaseSPL2ParserListener) EnterLookupColumn(ctx *LookupColumnContext) {}

// ExitLookupColumn is called when production lookupColumn is exited.
func (s *BaseSPL2ParserListener) ExitLookupColumn(ctx *LookupColumnContext) {}

// EnterLookupEventField is called when production lookupEventField is entered.
func (s *BaseSPL2ParserListener) EnterLookupEventField(ctx *LookupEventFieldContext) {}

// ExitLookupEventField is called when production lookupEventField is exited.
func (s *BaseSPL2ParserListener) ExitLookupEventField(ctx *LookupEventFieldContext) {}

// EnterSortCommand is called when production sortCommand is entered.
func (s *BaseSPL2ParserListener) EnterSortCommand(ctx *SortCommandContext) {}

// ExitSortCommand is called when production sortCommand is exited.
func (s *BaseSPL2ParserListener) ExitSortCommand(ctx *SortCommandContext) {}

// EnterSortTerm is called when production sortTerm is entered.
func (s *BaseSPL2ParserListener) EnterSortTerm(ctx *SortTermContext) {}

// ExitSortTerm is called when production sortTerm is exited.
func (s *BaseSPL2ParserListener) ExitSortTerm(ctx *SortTermContext) {}

// EnterSortWrapper is called when production sortWrapper is entered.
func (s *BaseSPL2ParserListener) EnterSortWrapper(ctx *SortWrapperContext) {}

// ExitSortWrapper is called when production sortWrapper is exited.
func (s *BaseSPL2ParserListener) ExitSortWrapper(ctx *SortWrapperContext) {}

// EnterIntegerValue is called when production integerValue is entered.
func (s *BaseSPL2ParserListener) EnterIntegerValue(ctx *IntegerValueContext) {}

// ExitIntegerValue is called when production integerValue is exited.
func (s *BaseSPL2ParserListener) ExitIntegerValue(ctx *IntegerValueContext) {}

// EnterDedupCommand is called when production dedupCommand is entered.
func (s *BaseSPL2ParserListener) EnterDedupCommand(ctx *DedupCommandContext) {}

// ExitDedupCommand is called when production dedupCommand is exited.
func (s *BaseSPL2ParserListener) ExitDedupCommand(ctx *DedupCommandContext) {}

// EnterKeepemptyOption is called when production keepemptyOption is entered.
func (s *BaseSPL2ParserListener) EnterKeepemptyOption(ctx *KeepemptyOptionContext) {}

// ExitKeepemptyOption is called when production keepemptyOption is exited.
func (s *BaseSPL2ParserListener) ExitKeepemptyOption(ctx *KeepemptyOptionContext) {}

// EnterConsecutiveOption is called when production consecutiveOption is entered.
func (s *BaseSPL2ParserListener) EnterConsecutiveOption(ctx *ConsecutiveOptionContext) {}

// ExitConsecutiveOption is called when production consecutiveOption is exited.
func (s *BaseSPL2ParserListener) ExitConsecutiveOption(ctx *ConsecutiveOptionContext) {}

// EnterDedupField is called when production dedupField is entered.
func (s *BaseSPL2ParserListener) EnterDedupField(ctx *DedupFieldContext) {}

// ExitDedupField is called when production dedupField is exited.
func (s *BaseSPL2ParserListener) ExitDedupField(ctx *DedupFieldContext) {}

// EnterHeadCommand is called when production headCommand is entered.
func (s *BaseSPL2ParserListener) EnterHeadCommand(ctx *HeadCommandContext) {}

// ExitHeadCommand is called when production headCommand is exited.
func (s *BaseSPL2ParserListener) ExitHeadCommand(ctx *HeadCommandContext) {}

// EnterKeeplastOption is called when production keeplastOption is entered.
func (s *BaseSPL2ParserListener) EnterKeeplastOption(ctx *KeeplastOptionContext) {}

// ExitKeeplastOption is called when production keeplastOption is exited.
func (s *BaseSPL2ParserListener) ExitKeeplastOption(ctx *KeeplastOptionContext) {}

// EnterHeadWhile is called when production headWhile is entered.
func (s *BaseSPL2ParserListener) EnterHeadWhile(ctx *HeadWhileContext) {}

// ExitHeadWhile is called when production headWhile is exited.
func (s *BaseSPL2ParserListener) ExitHeadWhile(ctx *HeadWhileContext) {}

// EnterHeadPostLayout is called when production headPostLayout is entered.
func (s *BaseSPL2ParserListener) EnterHeadPostLayout(ctx *HeadPostLayoutContext) {}

// ExitHeadPostLayout is called when production headPostLayout is exited.
func (s *BaseSPL2ParserListener) ExitHeadPostLayout(ctx *HeadPostLayoutContext) {}

// EnterReverseCommand is called when production reverseCommand is entered.
func (s *BaseSPL2ParserListener) EnterReverseCommand(ctx *ReverseCommandContext) {}

// ExitReverseCommand is called when production reverseCommand is exited.
func (s *BaseSPL2ParserListener) ExitReverseCommand(ctx *ReverseCommandContext) {}

// EnterUnknownOption is called when production unknownOption is entered.
func (s *BaseSPL2ParserListener) EnterUnknownOption(ctx *UnknownOptionContext) {}

// ExitUnknownOption is called when production unknownOption is exited.
func (s *BaseSPL2ParserListener) ExitUnknownOption(ctx *UnknownOptionContext) {}

// EnterUnknownOptionName is called when production unknownOptionName is entered.
func (s *BaseSPL2ParserListener) EnterUnknownOptionName(ctx *UnknownOptionNameContext) {}

// ExitUnknownOptionName is called when production unknownOptionName is exited.
func (s *BaseSPL2ParserListener) ExitUnknownOptionName(ctx *UnknownOptionNameContext) {}

// EnterIndependentSearch is called when production independentSearch is entered.
func (s *BaseSPL2ParserListener) EnterIndependentSearch(ctx *IndependentSearchContext) {}

// ExitIndependentSearch is called when production independentSearch is exited.
func (s *BaseSPL2ParserListener) ExitIndependentSearch(ctx *IndependentSearchContext) {}

// EnterInheritedSubpipe is called when production inheritedSubpipe is entered.
func (s *BaseSPL2ParserListener) EnterInheritedSubpipe(ctx *InheritedSubpipeContext) {}

// ExitInheritedSubpipe is called when production inheritedSubpipe is exited.
func (s *BaseSPL2ParserListener) ExitInheritedSubpipe(ctx *InheritedSubpipeContext) {}

// EnterJoinCommand is called when production joinCommand is entered.
func (s *BaseSPL2ParserListener) EnterJoinCommand(ctx *JoinCommandContext) {}

// ExitJoinCommand is called when production joinCommand is exited.
func (s *BaseSPL2ParserListener) ExitJoinCommand(ctx *JoinCommandContext) {}

// EnterJoinOption is called when production joinOption is entered.
func (s *BaseSPL2ParserListener) EnterJoinOption(ctx *JoinOptionContext) {}

// ExitJoinOption is called when production joinOption is exited.
func (s *BaseSPL2ParserListener) ExitJoinOption(ctx *JoinOptionContext) {}

// EnterJoinType is called when production joinType is entered.
func (s *BaseSPL2ParserListener) EnterJoinType(ctx *JoinTypeContext) {}

// ExitJoinType is called when production joinType is exited.
func (s *BaseSPL2ParserListener) ExitJoinType(ctx *JoinTypeContext) {}

// EnterAppendCommand is called when production appendCommand is entered.
func (s *BaseSPL2ParserListener) EnterAppendCommand(ctx *AppendCommandContext) {}

// ExitAppendCommand is called when production appendCommand is exited.
func (s *BaseSPL2ParserListener) ExitAppendCommand(ctx *AppendCommandContext) {}

// EnterAppendpipeCommand is called when production appendpipeCommand is entered.
func (s *BaseSPL2ParserListener) EnterAppendpipeCommand(ctx *AppendpipeCommandContext) {}

// ExitAppendpipeCommand is called when production appendpipeCommand is exited.
func (s *BaseSPL2ParserListener) ExitAppendpipeCommand(ctx *AppendpipeCommandContext) {}

// EnterAppendcolsCommand is called when production appendcolsCommand is entered.
func (s *BaseSPL2ParserListener) EnterAppendcolsCommand(ctx *AppendcolsCommandContext) {}

// ExitAppendcolsCommand is called when production appendcolsCommand is exited.
func (s *BaseSPL2ParserListener) ExitAppendcolsCommand(ctx *AppendcolsCommandContext) {}

// EnterUnionCommand is called when production unionCommand is entered.
func (s *BaseSPL2ParserListener) EnterUnionCommand(ctx *UnionCommandContext) {}

// ExitUnionCommand is called when production unionCommand is exited.
func (s *BaseSPL2ParserListener) ExitUnionCommand(ctx *UnionCommandContext) {}

// EnterUnionDataset is called when production unionDataset is entered.
func (s *BaseSPL2ParserListener) EnterUnionDataset(ctx *UnionDatasetContext) {}

// ExitUnionDataset is called when production unionDataset is exited.
func (s *BaseSPL2ParserListener) ExitUnionDataset(ctx *UnionDatasetContext) {}

// EnterBranchCommand is called when production branchCommand is entered.
func (s *BaseSPL2ParserListener) EnterBranchCommand(ctx *BranchCommandContext) {}

// ExitBranchCommand is called when production branchCommand is exited.
func (s *BaseSPL2ParserListener) ExitBranchCommand(ctx *BranchCommandContext) {}

// EnterBranchArm is called when production branchArm is entered.
func (s *BaseSPL2ParserListener) EnterBranchArm(ctx *BranchArmContext) {}

// ExitBranchArm is called when production branchArm is exited.
func (s *BaseSPL2ParserListener) ExitBranchArm(ctx *BranchArmContext) {}

// EnterIfCommand is called when production ifCommand is entered.
func (s *BaseSPL2ParserListener) EnterIfCommand(ctx *IfCommandContext) {}

// ExitIfCommand is called when production ifCommand is exited.
func (s *BaseSPL2ParserListener) ExitIfCommand(ctx *IfCommandContext) {}

// EnterBinCommand is called when production binCommand is entered.
func (s *BaseSPL2ParserListener) EnterBinCommand(ctx *BinCommandContext) {}

// ExitBinCommand is called when production binCommand is exited.
func (s *BaseSPL2ParserListener) ExitBinCommand(ctx *BinCommandContext) {}

// EnterBinOption is called when production binOption is entered.
func (s *BaseSPL2ParserListener) EnterBinOption(ctx *BinOptionContext) {}

// ExitBinOption is called when production binOption is exited.
func (s *BaseSPL2ParserListener) ExitBinOption(ctx *BinOptionContext) {}

// EnterAlignmentOption is called when production alignmentOption is entered.
func (s *BaseSPL2ParserListener) EnterAlignmentOption(ctx *AlignmentOptionContext) {}

// ExitAlignmentOption is called when production alignmentOption is exited.
func (s *BaseSPL2ParserListener) ExitAlignmentOption(ctx *AlignmentOptionContext) {}

// EnterBinSpan is called when production binSpan is entered.
func (s *BaseSPL2ParserListener) EnterBinSpan(ctx *BinSpanContext) {}

// ExitBinSpan is called when production binSpan is exited.
func (s *BaseSPL2ParserListener) ExitBinSpan(ctx *BinSpanContext) {}

// EnterSignedWeeklySpan is called when production signedWeeklySpan is entered.
func (s *BaseSPL2ParserListener) EnterSignedWeeklySpan(ctx *SignedWeeklySpanContext) {}

// ExitSignedWeeklySpan is called when production signedWeeklySpan is exited.
func (s *BaseSPL2ParserListener) ExitSignedWeeklySpan(ctx *SignedWeeklySpanContext) {}

// EnterLogarithmicSpan is called when production logarithmicSpan is entered.
func (s *BaseSPL2ParserListener) EnterLogarithmicSpan(ctx *LogarithmicSpanContext) {}

// ExitLogarithmicSpan is called when production logarithmicSpan is exited.
func (s *BaseSPL2ParserListener) ExitLogarithmicSpan(ctx *LogarithmicSpanContext) {}

// EnterExtendedOption is called when production extendedOption is entered.
func (s *BaseSPL2ParserListener) EnterExtendedOption(ctx *ExtendedOptionContext) {}

// ExitExtendedOption is called when production extendedOption is exited.
func (s *BaseSPL2ParserListener) ExitExtendedOption(ctx *ExtendedOptionContext) {}

// EnterSignedNumber is called when production signedNumber is entered.
func (s *BaseSPL2ParserListener) EnterSignedNumber(ctx *SignedNumberContext) {}

// ExitSignedNumber is called when production signedNumber is exited.
func (s *BaseSPL2ParserListener) ExitSignedNumber(ctx *SignedNumberContext) {}

// EnterRexCommand is called when production rexCommand is entered.
func (s *BaseSPL2ParserListener) EnterRexCommand(ctx *RexCommandContext) {}

// ExitRexCommand is called when production rexCommand is exited.
func (s *BaseSPL2ParserListener) ExitRexCommand(ctx *RexCommandContext) {}

// EnterRexOption is called when production rexOption is entered.
func (s *BaseSPL2ParserListener) EnterRexOption(ctx *RexOptionContext) {}

// ExitRexOption is called when production rexOption is exited.
func (s *BaseSPL2ParserListener) ExitRexOption(ctx *RexOptionContext) {}

// EnterSpathCommand is called when production spathCommand is entered.
func (s *BaseSPL2ParserListener) EnterSpathCommand(ctx *SpathCommandContext) {}

// ExitSpathCommand is called when production spathCommand is exited.
func (s *BaseSPL2ParserListener) ExitSpathCommand(ctx *SpathCommandContext) {}

// EnterSpathOption is called when production spathOption is entered.
func (s *BaseSPL2ParserListener) EnterSpathOption(ctx *SpathOptionContext) {}

// ExitSpathOption is called when production spathOption is exited.
func (s *BaseSPL2ParserListener) ExitSpathOption(ctx *SpathOptionContext) {}

// EnterLoadjobCommand is called when production loadjobCommand is entered.
func (s *BaseSPL2ParserListener) EnterLoadjobCommand(ctx *LoadjobCommandContext) {}

// ExitLoadjobCommand is called when production loadjobCommand is exited.
func (s *BaseSPL2ParserListener) ExitLoadjobCommand(ctx *LoadjobCommandContext) {}

// EnterMetricsCommand is called when production metricsCommand is entered.
func (s *BaseSPL2ParserListener) EnterMetricsCommand(ctx *MetricsCommandContext) {}

// ExitMetricsCommand is called when production metricsCommand is exited.
func (s *BaseSPL2ParserListener) ExitMetricsCommand(ctx *MetricsCommandContext) {}

// EnterMetricsAggregates is called when production metricsAggregates is entered.
func (s *BaseSPL2ParserListener) EnterMetricsAggregates(ctx *MetricsAggregatesContext) {}

// ExitMetricsAggregates is called when production metricsAggregates is exited.
func (s *BaseSPL2ParserListener) ExitMetricsAggregates(ctx *MetricsAggregatesContext) {}

// EnterMetricsOption is called when production metricsOption is entered.
func (s *BaseSPL2ParserListener) EnterMetricsOption(ctx *MetricsOptionContext) {}

// ExitMetricsOption is called when production metricsOption is exited.
func (s *BaseSPL2ParserListener) ExitMetricsOption(ctx *MetricsOptionContext) {}

// EnterTimechartCommand is called when production timechartCommand is entered.
func (s *BaseSPL2ParserListener) EnterTimechartCommand(ctx *TimechartCommandContext) {}

// ExitTimechartCommand is called when production timechartCommand is exited.
func (s *BaseSPL2ParserListener) ExitTimechartCommand(ctx *TimechartCommandContext) {}

// EnterTimechartExtraAggregate is called when production timechartExtraAggregate is entered.
func (s *BaseSPL2ParserListener) EnterTimechartExtraAggregate(ctx *TimechartExtraAggregateContext) {}

// ExitTimechartExtraAggregate is called when production timechartExtraAggregate is exited.
func (s *BaseSPL2ParserListener) ExitTimechartExtraAggregate(ctx *TimechartExtraAggregateContext) {}

// EnterTimechartOption is called when production timechartOption is entered.
func (s *BaseSPL2ParserListener) EnterTimechartOption(ctx *TimechartOptionContext) {}

// ExitTimechartOption is called when production timechartOption is exited.
func (s *BaseSPL2ParserListener) ExitTimechartOption(ctx *TimechartOptionContext) {}

// EnterTimechartSplit is called when production timechartSplit is entered.
func (s *BaseSPL2ParserListener) EnterTimechartSplit(ctx *TimechartSplitContext) {}

// ExitTimechartSplit is called when production timechartSplit is exited.
func (s *BaseSPL2ParserListener) ExitTimechartSplit(ctx *TimechartSplitContext) {}

// EnterTimechartSplitOption is called when production timechartSplitOption is entered.
func (s *BaseSPL2ParserListener) EnterTimechartSplitOption(ctx *TimechartSplitOptionContext) {}

// ExitTimechartSplitOption is called when production timechartSplitOption is exited.
func (s *BaseSPL2ParserListener) ExitTimechartSplitOption(ctx *TimechartSplitOptionContext) {}

// EnterTimewrapCommand is called when production timewrapCommand is entered.
func (s *BaseSPL2ParserListener) EnterTimewrapCommand(ctx *TimewrapCommandContext) {}

// ExitTimewrapCommand is called when production timewrapCommand is exited.
func (s *BaseSPL2ParserListener) ExitTimewrapCommand(ctx *TimewrapCommandContext) {}

// EnterMakemvCommand is called when production makemvCommand is entered.
func (s *BaseSPL2ParserListener) EnterMakemvCommand(ctx *MakemvCommandContext) {}

// ExitMakemvCommand is called when production makemvCommand is exited.
func (s *BaseSPL2ParserListener) ExitMakemvCommand(ctx *MakemvCommandContext) {}

// EnterMvexpandCommand is called when production mvexpandCommand is entered.
func (s *BaseSPL2ParserListener) EnterMvexpandCommand(ctx *MvexpandCommandContext) {}

// ExitMvexpandCommand is called when production mvexpandCommand is exited.
func (s *BaseSPL2ParserListener) ExitMvexpandCommand(ctx *MvexpandCommandContext) {}

// EnterMvcombineCommand is called when production mvcombineCommand is entered.
func (s *BaseSPL2ParserListener) EnterMvcombineCommand(ctx *MvcombineCommandContext) {}

// ExitMvcombineCommand is called when production mvcombineCommand is exited.
func (s *BaseSPL2ParserListener) ExitMvcombineCommand(ctx *MvcombineCommandContext) {}

// EnterFillnullCommand is called when production fillnullCommand is entered.
func (s *BaseSPL2ParserListener) EnterFillnullCommand(ctx *FillnullCommandContext) {}

// ExitFillnullCommand is called when production fillnullCommand is exited.
func (s *BaseSPL2ParserListener) ExitFillnullCommand(ctx *FillnullCommandContext) {}

// EnterEmbeddedCommand is called when production embeddedCommand is entered.
func (s *BaseSPL2ParserListener) EnterEmbeddedCommand(ctx *EmbeddedCommandContext) {}

// ExitEmbeddedCommand is called when production embeddedCommand is exited.
func (s *BaseSPL2ParserListener) ExitEmbeddedCommand(ctx *EmbeddedCommandContext) {}

// EnterEmbeddedText is called when production embeddedText is entered.
func (s *BaseSPL2ParserListener) EnterEmbeddedText(ctx *EmbeddedTextContext) {}

// ExitEmbeddedText is called when production embeddedText is exited.
func (s *BaseSPL2ParserListener) ExitEmbeddedText(ctx *EmbeddedTextContext) {}

// EnterModuleSuffix is called when production moduleSuffix is entered.
func (s *BaseSPL2ParserListener) EnterModuleSuffix(ctx *ModuleSuffixContext) {}

// ExitModuleSuffix is called when production moduleSuffix is exited.
func (s *BaseSPL2ParserListener) ExitModuleSuffix(ctx *ModuleSuffixContext) {}

// EnterTrailingPipelineBoundary is called when production trailingPipelineBoundary is entered.
func (s *BaseSPL2ParserListener) EnterTrailingPipelineBoundary(ctx *TrailingPipelineBoundaryContext) {
}

// ExitTrailingPipelineBoundary is called when production trailingPipelineBoundary is exited.
func (s *BaseSPL2ParserListener) ExitTrailingPipelineBoundary(ctx *TrailingPipelineBoundaryContext) {}

// EnterModuleDeclaration is called when production moduleDeclaration is entered.
func (s *BaseSPL2ParserListener) EnterModuleDeclaration(ctx *ModuleDeclarationContext) {}

// ExitModuleDeclaration is called when production moduleDeclaration is exited.
func (s *BaseSPL2ParserListener) ExitModuleDeclaration(ctx *ModuleDeclarationContext) {}

// EnterAnnotatedStatement is called when production annotatedStatement is entered.
func (s *BaseSPL2ParserListener) EnterAnnotatedStatement(ctx *AnnotatedStatementContext) {}

// ExitAnnotatedStatement is called when production annotatedStatement is exited.
func (s *BaseSPL2ParserListener) ExitAnnotatedStatement(ctx *AnnotatedStatementContext) {}

// EnterModuleStatement is called when production moduleStatement is entered.
func (s *BaseSPL2ParserListener) EnterModuleStatement(ctx *ModuleStatementContext) {}

// ExitModuleStatement is called when production moduleStatement is exited.
func (s *BaseSPL2ParserListener) ExitModuleStatement(ctx *ModuleStatementContext) {}

// EnterUnsupportedModuleBoundary is called when production unsupportedModuleBoundary is entered.
func (s *BaseSPL2ParserListener) EnterUnsupportedModuleBoundary(ctx *UnsupportedModuleBoundaryContext) {
}

// ExitUnsupportedModuleBoundary is called when production unsupportedModuleBoundary is exited.
func (s *BaseSPL2ParserListener) ExitUnsupportedModuleBoundary(ctx *UnsupportedModuleBoundaryContext) {
}

// EnterUnsupportedImportWildcard is called when production unsupportedImportWildcard is entered.
func (s *BaseSPL2ParserListener) EnterUnsupportedImportWildcard(ctx *UnsupportedImportWildcardContext) {
}

// ExitUnsupportedImportWildcard is called when production unsupportedImportWildcard is exited.
func (s *BaseSPL2ParserListener) ExitUnsupportedImportWildcard(ctx *UnsupportedImportWildcardContext) {
}

// EnterUnsupportedExportView is called when production unsupportedExportView is entered.
func (s *BaseSPL2ParserListener) EnterUnsupportedExportView(ctx *UnsupportedExportViewContext) {}

// ExitUnsupportedExportView is called when production unsupportedExportView is exited.
func (s *BaseSPL2ParserListener) ExitUnsupportedExportView(ctx *UnsupportedExportViewContext) {}

// EnterUnsupportedFunctionTerminator is called when production unsupportedFunctionTerminator is entered.
func (s *BaseSPL2ParserListener) EnterUnsupportedFunctionTerminator(ctx *UnsupportedFunctionTerminatorContext) {
}

// ExitUnsupportedFunctionTerminator is called when production unsupportedFunctionTerminator is exited.
func (s *BaseSPL2ParserListener) ExitUnsupportedFunctionTerminator(ctx *UnsupportedFunctionTerminatorContext) {
}

// EnterViewDeclaration is called when production viewDeclaration is entered.
func (s *BaseSPL2ParserListener) EnterViewDeclaration(ctx *ViewDeclarationContext) {}

// ExitViewDeclaration is called when production viewDeclaration is exited.
func (s *BaseSPL2ParserListener) ExitViewDeclaration(ctx *ViewDeclarationContext) {}

// EnterFunctionDeclaration is called when production functionDeclaration is entered.
func (s *BaseSPL2ParserListener) EnterFunctionDeclaration(ctx *FunctionDeclarationContext) {}

// ExitFunctionDeclaration is called when production functionDeclaration is exited.
func (s *BaseSPL2ParserListener) ExitFunctionDeclaration(ctx *FunctionDeclarationContext) {}

// EnterFunctionParameters is called when production functionParameters is entered.
func (s *BaseSPL2ParserListener) EnterFunctionParameters(ctx *FunctionParametersContext) {}

// ExitFunctionParameters is called when production functionParameters is exited.
func (s *BaseSPL2ParserListener) ExitFunctionParameters(ctx *FunctionParametersContext) {}

// EnterFunctionParameter is called when production functionParameter is entered.
func (s *BaseSPL2ParserListener) EnterFunctionParameter(ctx *FunctionParameterContext) {}

// ExitFunctionParameter is called when production functionParameter is exited.
func (s *BaseSPL2ParserListener) ExitFunctionParameter(ctx *FunctionParameterContext) {}

// EnterReturnStatement is called when production returnStatement is entered.
func (s *BaseSPL2ParserListener) EnterReturnStatement(ctx *ReturnStatementContext) {}

// ExitReturnStatement is called when production returnStatement is exited.
func (s *BaseSPL2ParserListener) ExitReturnStatement(ctx *ReturnStatementContext) {}

// EnterImportDeclaration is called when production importDeclaration is entered.
func (s *BaseSPL2ParserListener) EnterImportDeclaration(ctx *ImportDeclarationContext) {}

// ExitImportDeclaration is called when production importDeclaration is exited.
func (s *BaseSPL2ParserListener) ExitImportDeclaration(ctx *ImportDeclarationContext) {}

// EnterImportSelection is called when production importSelection is entered.
func (s *BaseSPL2ParserListener) EnterImportSelection(ctx *ImportSelectionContext) {}

// ExitImportSelection is called when production importSelection is exited.
func (s *BaseSPL2ParserListener) ExitImportSelection(ctx *ImportSelectionContext) {}

// EnterImportWildcard is called when production importWildcard is entered.
func (s *BaseSPL2ParserListener) EnterImportWildcard(ctx *ImportWildcardContext) {}

// ExitImportWildcard is called when production importWildcard is exited.
func (s *BaseSPL2ParserListener) ExitImportWildcard(ctx *ImportWildcardContext) {}

// EnterImportList is called when production importList is entered.
func (s *BaseSPL2ParserListener) EnterImportList(ctx *ImportListContext) {}

// ExitImportList is called when production importList is exited.
func (s *BaseSPL2ParserListener) ExitImportList(ctx *ImportListContext) {}

// EnterAliasedImport is called when production aliasedImport is entered.
func (s *BaseSPL2ParserListener) EnterAliasedImport(ctx *AliasedImportContext) {}

// ExitAliasedImport is called when production aliasedImport is exited.
func (s *BaseSPL2ParserListener) ExitAliasedImport(ctx *AliasedImportContext) {}

// EnterExportDeclaration is called when production exportDeclaration is entered.
func (s *BaseSPL2ParserListener) EnterExportDeclaration(ctx *ExportDeclarationContext) {}

// ExitExportDeclaration is called when production exportDeclaration is exited.
func (s *BaseSPL2ParserListener) ExitExportDeclaration(ctx *ExportDeclarationContext) {}

// EnterExportSelection is called when production exportSelection is entered.
func (s *BaseSPL2ParserListener) EnterExportSelection(ctx *ExportSelectionContext) {}

// ExitExportSelection is called when production exportSelection is exited.
func (s *BaseSPL2ParserListener) ExitExportSelection(ctx *ExportSelectionContext) {}

// EnterExportList is called when production exportList is entered.
func (s *BaseSPL2ParserListener) EnterExportList(ctx *ExportListContext) {}

// ExitExportList is called when production exportList is exited.
func (s *BaseSPL2ParserListener) ExitExportList(ctx *ExportListContext) {}

// EnterAliasedExport is called when production aliasedExport is entered.
func (s *BaseSPL2ParserListener) EnterAliasedExport(ctx *AliasedExportContext) {}

// ExitAliasedExport is called when production aliasedExport is exited.
func (s *BaseSPL2ParserListener) ExitAliasedExport(ctx *AliasedExportContext) {}

// EnterQualifiedName is called when production qualifiedName is entered.
func (s *BaseSPL2ParserListener) EnterQualifiedName(ctx *QualifiedNameContext) {}

// ExitQualifiedName is called when production qualifiedName is exited.
func (s *BaseSPL2ParserListener) ExitQualifiedName(ctx *QualifiedNameContext) {}

// EnterAnnotations is called when production annotations is entered.
func (s *BaseSPL2ParserListener) EnterAnnotations(ctx *AnnotationsContext) {}

// ExitAnnotations is called when production annotations is exited.
func (s *BaseSPL2ParserListener) ExitAnnotations(ctx *AnnotationsContext) {}

// EnterAnnotation is called when production annotation is entered.
func (s *BaseSPL2ParserListener) EnterAnnotation(ctx *AnnotationContext) {}

// ExitAnnotation is called when production annotation is exited.
func (s *BaseSPL2ParserListener) ExitAnnotation(ctx *AnnotationContext) {}

// EnterAnnotationStatement is called when production annotationStatement is entered.
func (s *BaseSPL2ParserListener) EnterAnnotationStatement(ctx *AnnotationStatementContext) {}

// ExitAnnotationStatement is called when production annotationStatement is exited.
func (s *BaseSPL2ParserListener) ExitAnnotationStatement(ctx *AnnotationStatementContext) {}

// EnterStatementTerminator is called when production statementTerminator is entered.
func (s *BaseSPL2ParserListener) EnterStatementTerminator(ctx *StatementTerminatorContext) {}

// ExitStatementTerminator is called when production statementTerminator is exited.
func (s *BaseSPL2ParserListener) ExitStatementTerminator(ctx *StatementTerminatorContext) {}

// EnterSearchCommand is called when production searchCommand is entered.
func (s *BaseSPL2ParserListener) EnterSearchCommand(ctx *SearchCommandContext) {}

// ExitSearchCommand is called when production searchCommand is exited.
func (s *BaseSPL2ParserListener) ExitSearchCommand(ctx *SearchCommandContext) {}

// EnterImplicitSearch is called when production implicitSearch is entered.
func (s *BaseSPL2ParserListener) EnterImplicitSearch(ctx *ImplicitSearchContext) {}

// ExitImplicitSearch is called when production implicitSearch is exited.
func (s *BaseSPL2ParserListener) ExitImplicitSearch(ctx *ImplicitSearchContext) {}

// EnterSearchExpression is called when production searchExpression is entered.
func (s *BaseSPL2ParserListener) EnterSearchExpression(ctx *SearchExpressionContext) {}

// ExitSearchExpression is called when production searchExpression is exited.
func (s *BaseSPL2ParserListener) ExitSearchExpression(ctx *SearchExpressionContext) {}

// EnterSearchXor is called when production searchXor is entered.
func (s *BaseSPL2ParserListener) EnterSearchXor(ctx *SearchXorContext) {}

// ExitSearchXor is called when production searchXor is exited.
func (s *BaseSPL2ParserListener) ExitSearchXor(ctx *SearchXorContext) {}

// EnterSearchAnd is called when production searchAnd is entered.
func (s *BaseSPL2ParserListener) EnterSearchAnd(ctx *SearchAndContext) {}

// ExitSearchAnd is called when production searchAnd is exited.
func (s *BaseSPL2ParserListener) ExitSearchAnd(ctx *SearchAndContext) {}

// EnterSearchOr is called when production searchOr is entered.
func (s *BaseSPL2ParserListener) EnterSearchOr(ctx *SearchOrContext) {}

// ExitSearchOr is called when production searchOr is exited.
func (s *BaseSPL2ParserListener) ExitSearchOr(ctx *SearchOrContext) {}

// EnterSearchNot is called when production searchNot is entered.
func (s *BaseSPL2ParserListener) EnterSearchNot(ctx *SearchNotContext) {}

// ExitSearchNot is called when production searchNot is exited.
func (s *BaseSPL2ParserListener) ExitSearchNot(ctx *SearchNotContext) {}

// EnterSearchAtom is called when production searchAtom is entered.
func (s *BaseSPL2ParserListener) EnterSearchAtom(ctx *SearchAtomContext) {}

// ExitSearchAtom is called when production searchAtom is exited.
func (s *BaseSPL2ParserListener) ExitSearchAtom(ctx *SearchAtomContext) {}

// EnterSearchValue is called when production searchValue is entered.
func (s *BaseSPL2ParserListener) EnterSearchValue(ctx *SearchValueContext) {}

// ExitSearchValue is called when production searchValue is exited.
func (s *BaseSPL2ParserListener) ExitSearchValue(ctx *SearchValueContext) {}

// EnterSearchWordLiteral is called when production searchWordLiteral is entered.
func (s *BaseSPL2ParserListener) EnterSearchWordLiteral(ctx *SearchWordLiteralContext) {}

// ExitSearchWordLiteral is called when production searchWordLiteral is exited.
func (s *BaseSPL2ParserListener) ExitSearchWordLiteral(ctx *SearchWordLiteralContext) {}

// EnterSearchSignedNumber is called when production searchSignedNumber is entered.
func (s *BaseSPL2ParserListener) EnterSearchSignedNumber(ctx *SearchSignedNumberContext) {}

// ExitSearchSignedNumber is called when production searchSignedNumber is exited.
func (s *BaseSPL2ParserListener) ExitSearchSignedNumber(ctx *SearchSignedNumberContext) {}

// EnterSearchUnprovedLiteral is called when production searchUnprovedLiteral is entered.
func (s *BaseSPL2ParserListener) EnterSearchUnprovedLiteral(ctx *SearchUnprovedLiteralContext) {}

// ExitSearchUnprovedLiteral is called when production searchUnprovedLiteral is exited.
func (s *BaseSPL2ParserListener) ExitSearchUnprovedLiteral(ctx *SearchUnprovedLiteralContext) {}

// EnterSearchBareValue is called when production searchBareValue is entered.
func (s *BaseSPL2ParserListener) EnterSearchBareValue(ctx *SearchBareValueContext) {}

// ExitSearchBareValue is called when production searchBareValue is exited.
func (s *BaseSPL2ParserListener) ExitSearchBareValue(ctx *SearchBareValueContext) {}

// EnterSearchDirective is called when production searchDirective is entered.
func (s *BaseSPL2ParserListener) EnterSearchDirective(ctx *SearchDirectiveContext) {}

// ExitSearchDirective is called when production searchDirective is exited.
func (s *BaseSPL2ParserListener) ExitSearchDirective(ctx *SearchDirectiveContext) {}

// EnterSearchTimeModifier is called when production searchTimeModifier is entered.
func (s *BaseSPL2ParserListener) EnterSearchTimeModifier(ctx *SearchTimeModifierContext) {}

// ExitSearchTimeModifier is called when production searchTimeModifier is exited.
func (s *BaseSPL2ParserListener) ExitSearchTimeModifier(ctx *SearchTimeModifierContext) {}

// EnterTimeModifierKey is called when production timeModifierKey is entered.
func (s *BaseSPL2ParserListener) EnterTimeModifierKey(ctx *TimeModifierKeyContext) {}

// ExitTimeModifierKey is called when production timeModifierKey is exited.
func (s *BaseSPL2ParserListener) ExitTimeModifierKey(ctx *TimeModifierKeyContext) {}

// EnterTimeModifierValue is called when production timeModifierValue is entered.
func (s *BaseSPL2ParserListener) EnterTimeModifierValue(ctx *TimeModifierValueContext) {}

// ExitTimeModifierValue is called when production timeModifierValue is exited.
func (s *BaseSPL2ParserListener) ExitTimeModifierValue(ctx *TimeModifierValueContext) {}

// EnterRelativeTime is called when production relativeTime is entered.
func (s *BaseSPL2ParserListener) EnterRelativeTime(ctx *RelativeTimeContext) {}

// ExitRelativeTime is called when production relativeTime is exited.
func (s *BaseSPL2ParserListener) ExitRelativeTime(ctx *RelativeTimeContext) {}

// EnterExpression is called when production expression is entered.
func (s *BaseSPL2ParserListener) EnterExpression(ctx *ExpressionContext) {}

// ExitExpression is called when production expression is exited.
func (s *BaseSPL2ParserListener) ExitExpression(ctx *ExpressionContext) {}

// EnterXorExpression is called when production xorExpression is entered.
func (s *BaseSPL2ParserListener) EnterXorExpression(ctx *XorExpressionContext) {}

// ExitXorExpression is called when production xorExpression is exited.
func (s *BaseSPL2ParserListener) ExitXorExpression(ctx *XorExpressionContext) {}

// EnterOrExpression is called when production orExpression is entered.
func (s *BaseSPL2ParserListener) EnterOrExpression(ctx *OrExpressionContext) {}

// ExitOrExpression is called when production orExpression is exited.
func (s *BaseSPL2ParserListener) ExitOrExpression(ctx *OrExpressionContext) {}

// EnterAndExpression is called when production andExpression is entered.
func (s *BaseSPL2ParserListener) EnterAndExpression(ctx *AndExpressionContext) {}

// ExitAndExpression is called when production andExpression is exited.
func (s *BaseSPL2ParserListener) ExitAndExpression(ctx *AndExpressionContext) {}

// EnterNotExpression is called when production notExpression is entered.
func (s *BaseSPL2ParserListener) EnterNotExpression(ctx *NotExpressionContext) {}

// ExitNotExpression is called when production notExpression is exited.
func (s *BaseSPL2ParserListener) ExitNotExpression(ctx *NotExpressionContext) {}

// EnterPredicate is called when production predicate is entered.
func (s *BaseSPL2ParserListener) EnterPredicate(ctx *PredicateContext) {}

// ExitPredicate is called when production predicate is exited.
func (s *BaseSPL2ParserListener) ExitPredicate(ctx *PredicateContext) {}

// EnterLogicalAnd is called when production logicalAnd is entered.
func (s *BaseSPL2ParserListener) EnterLogicalAnd(ctx *LogicalAndContext) {}

// ExitLogicalAnd is called when production logicalAnd is exited.
func (s *BaseSPL2ParserListener) ExitLogicalAnd(ctx *LogicalAndContext) {}

// EnterLogicalOr is called when production logicalOr is entered.
func (s *BaseSPL2ParserListener) EnterLogicalOr(ctx *LogicalOrContext) {}

// ExitLogicalOr is called when production logicalOr is exited.
func (s *BaseSPL2ParserListener) ExitLogicalOr(ctx *LogicalOrContext) {}

// EnterLogicalXor is called when production logicalXor is entered.
func (s *BaseSPL2ParserListener) EnterLogicalXor(ctx *LogicalXorContext) {}

// ExitLogicalXor is called when production logicalXor is exited.
func (s *BaseSPL2ParserListener) ExitLogicalXor(ctx *LogicalXorContext) {}

// EnterLogicalNot is called when production logicalNot is entered.
func (s *BaseSPL2ParserListener) EnterLogicalNot(ctx *LogicalNotContext) {}

// ExitLogicalNot is called when production logicalNot is exited.
func (s *BaseSPL2ParserListener) ExitLogicalNot(ctx *LogicalNotContext) {}

// EnterBetweenOperator is called when production betweenOperator is entered.
func (s *BaseSPL2ParserListener) EnterBetweenOperator(ctx *BetweenOperatorContext) {}

// ExitBetweenOperator is called when production betweenOperator is exited.
func (s *BaseSPL2ParserListener) ExitBetweenOperator(ctx *BetweenOperatorContext) {}

// EnterBetweenConjunction is called when production betweenConjunction is entered.
func (s *BaseSPL2ParserListener) EnterBetweenConjunction(ctx *BetweenConjunctionContext) {}

// ExitBetweenConjunction is called when production betweenConjunction is exited.
func (s *BaseSPL2ParserListener) ExitBetweenConjunction(ctx *BetweenConjunctionContext) {}

// EnterComparison is called when production comparison is entered.
func (s *BaseSPL2ParserListener) EnterComparison(ctx *ComparisonContext) {}

// ExitComparison is called when production comparison is exited.
func (s *BaseSPL2ParserListener) ExitComparison(ctx *ComparisonContext) {}

// EnterAdditive is called when production additive is entered.
func (s *BaseSPL2ParserListener) EnterAdditive(ctx *AdditiveContext) {}

// ExitAdditive is called when production additive is exited.
func (s *BaseSPL2ParserListener) ExitAdditive(ctx *AdditiveContext) {}

// EnterMultiplicative is called when production multiplicative is entered.
func (s *BaseSPL2ParserListener) EnterMultiplicative(ctx *MultiplicativeContext) {}

// ExitMultiplicative is called when production multiplicative is exited.
func (s *BaseSPL2ParserListener) ExitMultiplicative(ctx *MultiplicativeContext) {}

// EnterUnary is called when production unary is entered.
func (s *BaseSPL2ParserListener) EnterUnary(ctx *UnaryContext) {}

// ExitUnary is called when production unary is exited.
func (s *BaseSPL2ParserListener) ExitUnary(ctx *UnaryContext) {}

// EnterAccess is called when production access is entered.
func (s *BaseSPL2ParserListener) EnterAccess(ctx *AccessContext) {}

// ExitAccess is called when production access is exited.
func (s *BaseSPL2ParserListener) ExitAccess(ctx *AccessContext) {}

// EnterAccessPart is called when production accessPart is entered.
func (s *BaseSPL2ParserListener) EnterAccessPart(ctx *AccessPartContext) {}

// ExitAccessPart is called when production accessPart is exited.
func (s *BaseSPL2ParserListener) ExitAccessPart(ctx *AccessPartContext) {}

// EnterPrimary is called when production primary is entered.
func (s *BaseSPL2ParserListener) EnterPrimary(ctx *PrimaryContext) {}

// ExitPrimary is called when production primary is exited.
func (s *BaseSPL2ParserListener) ExitPrimary(ctx *PrimaryContext) {}

// EnterMultilineOperator is called when production multilineOperator is entered.
func (s *BaseSPL2ParserListener) EnterMultilineOperator(ctx *MultilineOperatorContext) {}

// ExitMultilineOperator is called when production multilineOperator is exited.
func (s *BaseSPL2ParserListener) ExitMultilineOperator(ctx *MultilineOperatorContext) {}

// EnterMultilineOperand is called when production multilineOperand is entered.
func (s *BaseSPL2ParserListener) EnterMultilineOperand(ctx *MultilineOperandContext) {}

// ExitMultilineOperand is called when production multilineOperand is exited.
func (s *BaseSPL2ParserListener) ExitMultilineOperand(ctx *MultilineOperandContext) {}

// EnterMultilineSimpleCall is called when production multilineSimpleCall is entered.
func (s *BaseSPL2ParserListener) EnterMultilineSimpleCall(ctx *MultilineSimpleCallContext) {}

// ExitMultilineSimpleCall is called when production multilineSimpleCall is exited.
func (s *BaseSPL2ParserListener) ExitMultilineSimpleCall(ctx *MultilineSimpleCallContext) {}

// EnterMultilineAtom is called when production multilineAtom is entered.
func (s *BaseSPL2ParserListener) EnterMultilineAtom(ctx *MultilineAtomContext) {}

// ExitMultilineAtom is called when production multilineAtom is exited.
func (s *BaseSPL2ParserListener) ExitMultilineAtom(ctx *MultilineAtomContext) {}

// EnterMultilineAccessPart is called when production multilineAccessPart is entered.
func (s *BaseSPL2ParserListener) EnterMultilineAccessPart(ctx *MultilineAccessPartContext) {}

// ExitMultilineAccessPart is called when production multilineAccessPart is exited.
func (s *BaseSPL2ParserListener) ExitMultilineAccessPart(ctx *MultilineAccessPartContext) {}

// EnterCall is called when production call is entered.
func (s *BaseSPL2ParserListener) EnterCall(ctx *CallContext) {}

// ExitCall is called when production call is exited.
func (s *BaseSPL2ParserListener) ExitCall(ctx *CallContext) {}

// EnterMultilineCall is called when production multilineCall is entered.
func (s *BaseSPL2ParserListener) EnterMultilineCall(ctx *MultilineCallContext) {}

// ExitMultilineCall is called when production multilineCall is exited.
func (s *BaseSPL2ParserListener) ExitMultilineCall(ctx *MultilineCallContext) {}

// EnterMultilineArguments is called when production multilineArguments is entered.
func (s *BaseSPL2ParserListener) EnterMultilineArguments(ctx *MultilineArgumentsContext) {}

// ExitMultilineArguments is called when production multilineArguments is exited.
func (s *BaseSPL2ParserListener) ExitMultilineArguments(ctx *MultilineArgumentsContext) {}

// EnterMultilineArgument is called when production multilineArgument is entered.
func (s *BaseSPL2ParserListener) EnterMultilineArgument(ctx *MultilineArgumentContext) {}

// ExitMultilineArgument is called when production multilineArgument is exited.
func (s *BaseSPL2ParserListener) ExitMultilineArgument(ctx *MultilineArgumentContext) {}

// EnterArguments is called when production arguments is entered.
func (s *BaseSPL2ParserListener) EnterArguments(ctx *ArgumentsContext) {}

// ExitArguments is called when production arguments is exited.
func (s *BaseSPL2ParserListener) ExitArguments(ctx *ArgumentsContext) {}

// EnterNamedArgument is called when production namedArgument is entered.
func (s *BaseSPL2ParserListener) EnterNamedArgument(ctx *NamedArgumentContext) {}

// ExitNamedArgument is called when production namedArgument is exited.
func (s *BaseSPL2ParserListener) ExitNamedArgument(ctx *NamedArgumentContext) {}

// EnterLiteral is called when production literal is entered.
func (s *BaseSPL2ParserListener) EnterLiteral(ctx *LiteralContext) {}

// ExitLiteral is called when production literal is exited.
func (s *BaseSPL2ParserListener) ExitLiteral(ctx *LiteralContext) {}

// EnterStringLiteral is called when production stringLiteral is entered.
func (s *BaseSPL2ParserListener) EnterStringLiteral(ctx *StringLiteralContext) {}

// ExitStringLiteral is called when production stringLiteral is exited.
func (s *BaseSPL2ParserListener) ExitStringLiteral(ctx *StringLiteralContext) {}

// EnterQuotedName is called when production quotedName is entered.
func (s *BaseSPL2ParserListener) EnterQuotedName(ctx *QuotedNameContext) {}

// ExitQuotedName is called when production quotedName is exited.
func (s *BaseSPL2ParserListener) ExitQuotedName(ctx *QuotedNameContext) {}

// EnterFieldTemplate is called when production fieldTemplate is entered.
func (s *BaseSPL2ParserListener) EnterFieldTemplate(ctx *FieldTemplateContext) {}

// ExitFieldTemplate is called when production fieldTemplate is exited.
func (s *BaseSPL2ParserListener) ExitFieldTemplate(ctx *FieldTemplateContext) {}

// EnterFieldName is called when production fieldName is entered.
func (s *BaseSPL2ParserListener) EnterFieldName(ctx *FieldNameContext) {}

// ExitFieldName is called when production fieldName is exited.
func (s *BaseSPL2ParserListener) ExitFieldName(ctx *FieldNameContext) {}

// EnterIdentifier is called when production identifier is entered.
func (s *BaseSPL2ParserListener) EnterIdentifier(ctx *IdentifierContext) {}

// ExitIdentifier is called when production identifier is exited.
func (s *BaseSPL2ParserListener) ExitIdentifier(ctx *IdentifierContext) {}

// EnterSqlKeyword is called when production sqlKeyword is entered.
func (s *BaseSPL2ParserListener) EnterSqlKeyword(ctx *SqlKeywordContext) {}

// ExitSqlKeyword is called when production sqlKeyword is exited.
func (s *BaseSPL2ParserListener) ExitSqlKeyword(ctx *SqlKeywordContext) {}

// EnterPipelineKeyword is called when production pipelineKeyword is entered.
func (s *BaseSPL2ParserListener) EnterPipelineKeyword(ctx *PipelineKeywordContext) {}

// ExitPipelineKeyword is called when production pipelineKeyword is exited.
func (s *BaseSPL2ParserListener) ExitPipelineKeyword(ctx *PipelineKeywordContext) {}

// EnterArray is called when production array is entered.
func (s *BaseSPL2ParserListener) EnterArray(ctx *ArrayContext) {}

// ExitArray is called when production array is exited.
func (s *BaseSPL2ParserListener) ExitArray(ctx *ArrayContext) {}

// EnterObject is called when production object is entered.
func (s *BaseSPL2ParserListener) EnterObject(ctx *ObjectContext) {}

// ExitObject is called when production object is exited.
func (s *BaseSPL2ParserListener) ExitObject(ctx *ObjectContext) {}

// EnterObjectEntry is called when production objectEntry is entered.
func (s *BaseSPL2ParserListener) EnterObjectEntry(ctx *ObjectEntryContext) {}

// ExitObjectEntry is called when production objectEntry is exited.
func (s *BaseSPL2ParserListener) ExitObjectEntry(ctx *ObjectEntryContext) {}

// EnterObjectKey is called when production objectKey is entered.
func (s *BaseSPL2ParserListener) EnterObjectKey(ctx *ObjectKeyContext) {}

// ExitObjectKey is called when production objectKey is exited.
func (s *BaseSPL2ParserListener) ExitObjectKey(ctx *ObjectKeyContext) {}

// EnterSearchLiteral is called when production searchLiteral is entered.
func (s *BaseSPL2ParserListener) EnterSearchLiteral(ctx *SearchLiteralContext) {}

// ExitSearchLiteral is called when production searchLiteral is exited.
func (s *BaseSPL2ParserListener) ExitSearchLiteral(ctx *SearchLiteralContext) {}

// EnterLambdaExpression is called when production lambdaExpression is entered.
func (s *BaseSPL2ParserListener) EnterLambdaExpression(ctx *LambdaExpressionContext) {}

// ExitLambdaExpression is called when production lambdaExpression is exited.
func (s *BaseSPL2ParserListener) ExitLambdaExpression(ctx *LambdaExpressionContext) {}

// EnterLambdaParameter is called when production lambdaParameter is entered.
func (s *BaseSPL2ParserListener) EnterLambdaParameter(ctx *LambdaParameterContext) {}

// ExitLambdaParameter is called when production lambdaParameter is exited.
func (s *BaseSPL2ParserListener) ExitLambdaParameter(ctx *LambdaParameterContext) {}

// EnterLambdaBlock is called when production lambdaBlock is entered.
func (s *BaseSPL2ParserListener) EnterLambdaBlock(ctx *LambdaBlockContext) {}

// ExitLambdaBlock is called when production lambdaBlock is exited.
func (s *BaseSPL2ParserListener) ExitLambdaBlock(ctx *LambdaBlockContext) {}
