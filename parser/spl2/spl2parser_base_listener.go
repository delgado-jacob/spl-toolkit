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

// EnterProjection is called when production projection is entered.
func (s *BaseSPL2ParserListener) EnterProjection(ctx *ProjectionContext) {}

// ExitProjection is called when production projection is exited.
func (s *BaseSPL2ParserListener) ExitProjection(ctx *ProjectionContext) {}

// EnterDataset is called when production dataset is entered.
func (s *BaseSPL2ParserListener) EnterDataset(ctx *DatasetContext) {}

// ExitDataset is called when production dataset is exited.
func (s *BaseSPL2ParserListener) ExitDataset(ctx *DatasetContext) {}

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

// EnterRexCommand is called when production rexCommand is entered.
func (s *BaseSPL2ParserListener) EnterRexCommand(ctx *RexCommandContext) {}

// ExitRexCommand is called when production rexCommand is exited.
func (s *BaseSPL2ParserListener) ExitRexCommand(ctx *RexCommandContext) {}

// EnterRexOption is called when production rexOption is entered.
func (s *BaseSPL2ParserListener) EnterRexOption(ctx *RexOptionContext) {}

// ExitRexOption is called when production rexOption is exited.
func (s *BaseSPL2ParserListener) ExitRexOption(ctx *RexOptionContext) {}

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

// EnterModuleDeclaration is called when production moduleDeclaration is entered.
func (s *BaseSPL2ParserListener) EnterModuleDeclaration(ctx *ModuleDeclarationContext) {}

// ExitModuleDeclaration is called when production moduleDeclaration is exited.
func (s *BaseSPL2ParserListener) ExitModuleDeclaration(ctx *ModuleDeclarationContext) {}

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

// EnterCall is called when production call is entered.
func (s *BaseSPL2ParserListener) EnterCall(ctx *CallContext) {}

// ExitCall is called when production call is exited.
func (s *BaseSPL2ParserListener) ExitCall(ctx *CallContext) {}

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
