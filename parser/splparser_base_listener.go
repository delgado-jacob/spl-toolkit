// Code generated from grammar/SPLParser.g4 by ANTLR 4.13.2. DO NOT EDIT.

package parser // SPLParser
import "github.com/antlr4-go/antlr/v4"

// BaseSPLParserListener is a complete listener for a parse tree produced by SPLParser.
type BaseSPLParserListener struct{}

var _ SPLParserListener = &BaseSPLParserListener{}

// VisitTerminal is called when a terminal node is visited.
func (s *BaseSPLParserListener) VisitTerminal(node antlr.TerminalNode) {}

// VisitErrorNode is called when an error node is visited.
func (s *BaseSPLParserListener) VisitErrorNode(node antlr.ErrorNode) {}

// EnterEveryRule is called when any rule is entered.
func (s *BaseSPLParserListener) EnterEveryRule(ctx antlr.ParserRuleContext) {}

// ExitEveryRule is called when any rule is exited.
func (s *BaseSPLParserListener) ExitEveryRule(ctx antlr.ParserRuleContext) {}

// EnterQuery is called when production query is entered.
func (s *BaseSPLParserListener) EnterQuery(ctx *QueryContext) {}

// ExitQuery is called when production query is exited.
func (s *BaseSPLParserListener) ExitQuery(ctx *QueryContext) {}

// EnterInitCommand is called when production initCommand is entered.
func (s *BaseSPLParserListener) EnterInitCommand(ctx *InitCommandContext) {}

// ExitInitCommand is called when production initCommand is exited.
func (s *BaseSPLParserListener) ExitInitCommand(ctx *InitCommandContext) {}

// EnterNextCommand is called when production nextCommand is entered.
func (s *BaseSPLParserListener) EnterNextCommand(ctx *NextCommandContext) {}

// ExitNextCommand is called when production nextCommand is exited.
func (s *BaseSPLParserListener) ExitNextCommand(ctx *NextCommandContext) {}

// EnterSubquery is called when production subquery is entered.
func (s *BaseSPLParserListener) EnterSubquery(ctx *SubqueryContext) {}

// ExitSubquery is called when production subquery is exited.
func (s *BaseSPLParserListener) ExitSubquery(ctx *SubqueryContext) {}

// EnterLIKEOP is called when production LIKEOP is entered.
func (s *BaseSPLParserListener) EnterLIKEOP(ctx *LIKEOPContext) {}

// ExitLIKEOP is called when production LIKEOP is exited.
func (s *BaseSPLParserListener) ExitLIKEOP(ctx *LIKEOPContext) {}

// EnterANDOP is called when production ANDOP is entered.
func (s *BaseSPLParserListener) EnterANDOP(ctx *ANDOPContext) {}

// ExitANDOP is called when production ANDOP is exited.
func (s *BaseSPLParserListener) ExitANDOP(ctx *ANDOPContext) {}

// EnterOROP is called when production OROP is entered.
func (s *BaseSPLParserListener) EnterOROP(ctx *OROPContext) {}

// ExitOROP is called when production OROP is exited.
func (s *BaseSPLParserListener) ExitOROP(ctx *OROPContext) {}

// EnterINOP is called when production INOP is entered.
func (s *BaseSPLParserListener) EnterINOP(ctx *INOPContext) {}

// ExitINOP is called when production INOP is exited.
func (s *BaseSPLParserListener) ExitINOP(ctx *INOPContext) {}

// EnterNOTOP is called when production NOTOP is entered.
func (s *BaseSPLParserListener) EnterNOTOP(ctx *NOTOPContext) {}

// ExitNOTOP is called when production NOTOP is exited.
func (s *BaseSPLParserListener) ExitNOTOP(ctx *NOTOPContext) {}

// EnterBYOP is called when production BYOP is entered.
func (s *BaseSPLParserListener) EnterBYOP(ctx *BYOPContext) {}

// ExitBYOP is called when production BYOP is exited.
func (s *BaseSPLParserListener) ExitBYOP(ctx *BYOPContext) {}

// EnterPARENOP is called when production PARENOP is entered.
func (s *BaseSPLParserListener) EnterPARENOP(ctx *PARENOPContext) {}

// ExitPARENOP is called when production PARENOP is exited.
func (s *BaseSPLParserListener) ExitPARENOP(ctx *PARENOPContext) {}

// EnterOUTPUTMULTIINOP is called when production OUTPUTMULTIINOP is entered.
func (s *BaseSPLParserListener) EnterOUTPUTMULTIINOP(ctx *OUTPUTMULTIINOPContext) {}

// ExitOUTPUTMULTIINOP is called when production OUTPUTMULTIINOP is exited.
func (s *BaseSPLParserListener) ExitOUTPUTMULTIINOP(ctx *OUTPUTMULTIINOPContext) {}

// EnterKEYVALUEOP is called when production KEYVALUEOP is entered.
func (s *BaseSPLParserListener) EnterKEYVALUEOP(ctx *KEYVALUEOPContext) {}

// ExitKEYVALUEOP is called when production KEYVALUEOP is exited.
func (s *BaseSPLParserListener) ExitKEYVALUEOP(ctx *KEYVALUEOPContext) {}

// EnterOUTPUTOP is called when production OUTPUTOP is entered.
func (s *BaseSPLParserListener) EnterOUTPUTOP(ctx *OUTPUTOPContext) {}

// ExitOUTPUTOP is called when production OUTPUTOP is exited.
func (s *BaseSPLParserListener) ExitOUTPUTOP(ctx *OUTPUTOPContext) {}

// EnterEXPRESSIONOP is called when production EXPRESSIONOP is entered.
func (s *BaseSPLParserListener) EnterEXPRESSIONOP(ctx *EXPRESSIONOPContext) {}

// ExitEXPRESSIONOP is called when production EXPRESSIONOP is exited.
func (s *BaseSPLParserListener) ExitEXPRESSIONOP(ctx *EXPRESSIONOPContext) {}

// EnterOUTPUTMULTIOP is called when production OUTPUTMULTIOP is entered.
func (s *BaseSPLParserListener) EnterOUTPUTMULTIOP(ctx *OUTPUTMULTIOPContext) {}

// ExitOUTPUTMULTIOP is called when production OUTPUTMULTIOP is exited.
func (s *BaseSPLParserListener) ExitOUTPUTMULTIOP(ctx *OUTPUTMULTIOPContext) {}

// EnterRENAMEOP is called when production RENAMEOP is entered.
func (s *BaseSPLParserListener) EnterRENAMEOP(ctx *RENAMEOPContext) {}

// ExitRENAMEOP is called when production RENAMEOP is exited.
func (s *BaseSPLParserListener) ExitRENAMEOP(ctx *RENAMEOPContext) {}

// EnterExpression is called when production expression is entered.
func (s *BaseSPLParserListener) EnterExpression(ctx *ExpressionContext) {}

// ExitExpression is called when production expression is exited.
func (s *BaseSPLParserListener) ExitExpression(ctx *ExpressionContext) {}

// EnterValue is called when production value is entered.
func (s *BaseSPLParserListener) EnterValue(ctx *ValueContext) {}

// ExitValue is called when production value is exited.
func (s *BaseSPLParserListener) ExitValue(ctx *ValueContext) {}

// EnterDate is called when production date is entered.
func (s *BaseSPLParserListener) EnterDate(ctx *DateContext) {}

// ExitDate is called when production date is exited.
func (s *BaseSPLParserListener) ExitDate(ctx *DateContext) {}

// EnterFieldUse is called when production FieldUse is entered.
func (s *BaseSPLParserListener) EnterFieldUse(ctx *FieldUseContext) {}

// ExitFieldUse is called when production FieldUse is exited.
func (s *BaseSPLParserListener) ExitFieldUse(ctx *FieldUseContext) {}

// EnterCommandUse is called when production CommandUse is entered.
func (s *BaseSPLParserListener) EnterCommandUse(ctx *CommandUseContext) {}

// ExitCommandUse is called when production CommandUse is exited.
func (s *BaseSPLParserListener) ExitCommandUse(ctx *CommandUseContext) {}

// EnterFunctionUse is called when production FunctionUse is entered.
func (s *BaseSPLParserListener) EnterFunctionUse(ctx *FunctionUseContext) {}

// ExitFunctionUse is called when production FunctionUse is exited.
func (s *BaseSPLParserListener) ExitFunctionUse(ctx *FunctionUseContext) {}

// EnterFunction is called when production function is entered.
func (s *BaseSPLParserListener) EnterFunction(ctx *FunctionContext) {}

// ExitFunction is called when production function is exited.
func (s *BaseSPLParserListener) ExitFunction(ctx *FunctionContext) {}

// EnterCommand is called when production command is entered.
func (s *BaseSPLParserListener) EnterCommand(ctx *CommandContext) {}

// ExitCommand is called when production command is exited.
func (s *BaseSPLParserListener) ExitCommand(ctx *CommandContext) {}

// EnterAnalysisQuery is called when production analysisQuery is entered.
func (s *BaseSPLParserListener) EnterAnalysisQuery(ctx *AnalysisQueryContext) {}

// ExitAnalysisQuery is called when production analysisQuery is exited.
func (s *BaseSPLParserListener) ExitAnalysisQuery(ctx *AnalysisQueryContext) {}

// EnterAnalysisPipeline is called when production analysisPipeline is entered.
func (s *BaseSPLParserListener) EnterAnalysisPipeline(ctx *AnalysisPipelineContext) {}

// ExitAnalysisPipeline is called when production analysisPipeline is exited.
func (s *BaseSPLParserListener) ExitAnalysisPipeline(ctx *AnalysisPipelineContext) {}

// EnterAnalysisInitialStage is called when production analysisInitialStage is entered.
func (s *BaseSPLParserListener) EnterAnalysisInitialStage(ctx *AnalysisInitialStageContext) {}

// ExitAnalysisInitialStage is called when production analysisInitialStage is exited.
func (s *BaseSPLParserListener) ExitAnalysisInitialStage(ctx *AnalysisInitialStageContext) {}

// EnterAnalysisImplicitSearch is called when production analysisImplicitSearch is entered.
func (s *BaseSPLParserListener) EnterAnalysisImplicitSearch(ctx *AnalysisImplicitSearchContext) {}

// ExitAnalysisImplicitSearch is called when production analysisImplicitSearch is exited.
func (s *BaseSPLParserListener) ExitAnalysisImplicitSearch(ctx *AnalysisImplicitSearchContext) {}

// EnterAnalysisSearchStage is called when production AnalysisSearchStage is entered.
func (s *BaseSPLParserListener) EnterAnalysisSearchStage(ctx *AnalysisSearchStageContext) {}

// ExitAnalysisSearchStage is called when production AnalysisSearchStage is exited.
func (s *BaseSPLParserListener) ExitAnalysisSearchStage(ctx *AnalysisSearchStageContext) {}

// EnterAnalysisWhereStage is called when production AnalysisWhereStage is entered.
func (s *BaseSPLParserListener) EnterAnalysisWhereStage(ctx *AnalysisWhereStageContext) {}

// ExitAnalysisWhereStage is called when production AnalysisWhereStage is exited.
func (s *BaseSPLParserListener) ExitAnalysisWhereStage(ctx *AnalysisWhereStageContext) {}

// EnterAnalysisEvalStage is called when production AnalysisEvalStage is entered.
func (s *BaseSPLParserListener) EnterAnalysisEvalStage(ctx *AnalysisEvalStageContext) {}

// ExitAnalysisEvalStage is called when production AnalysisEvalStage is exited.
func (s *BaseSPLParserListener) ExitAnalysisEvalStage(ctx *AnalysisEvalStageContext) {}

// EnterAnalysisRenameStage is called when production AnalysisRenameStage is entered.
func (s *BaseSPLParserListener) EnterAnalysisRenameStage(ctx *AnalysisRenameStageContext) {}

// ExitAnalysisRenameStage is called when production AnalysisRenameStage is exited.
func (s *BaseSPLParserListener) ExitAnalysisRenameStage(ctx *AnalysisRenameStageContext) {}

// EnterAnalysisFieldsStage is called when production AnalysisFieldsStage is entered.
func (s *BaseSPLParserListener) EnterAnalysisFieldsStage(ctx *AnalysisFieldsStageContext) {}

// ExitAnalysisFieldsStage is called when production AnalysisFieldsStage is exited.
func (s *BaseSPLParserListener) ExitAnalysisFieldsStage(ctx *AnalysisFieldsStageContext) {}

// EnterAnalysisStatsStage is called when production AnalysisStatsStage is entered.
func (s *BaseSPLParserListener) EnterAnalysisStatsStage(ctx *AnalysisStatsStageContext) {}

// ExitAnalysisStatsStage is called when production AnalysisStatsStage is exited.
func (s *BaseSPLParserListener) ExitAnalysisStatsStage(ctx *AnalysisStatsStageContext) {}

// EnterAnalysisLookupStage is called when production AnalysisLookupStage is entered.
func (s *BaseSPLParserListener) EnterAnalysisLookupStage(ctx *AnalysisLookupStageContext) {}

// ExitAnalysisLookupStage is called when production AnalysisLookupStage is exited.
func (s *BaseSPLParserListener) ExitAnalysisLookupStage(ctx *AnalysisLookupStageContext) {}

// EnterAnalysisSortStage is called when production AnalysisSortStage is entered.
func (s *BaseSPLParserListener) EnterAnalysisSortStage(ctx *AnalysisSortStageContext) {}

// ExitAnalysisSortStage is called when production AnalysisSortStage is exited.
func (s *BaseSPLParserListener) ExitAnalysisSortStage(ctx *AnalysisSortStageContext) {}

// EnterAnalysisDedupStage is called when production AnalysisDedupStage is entered.
func (s *BaseSPLParserListener) EnterAnalysisDedupStage(ctx *AnalysisDedupStageContext) {}

// ExitAnalysisDedupStage is called when production AnalysisDedupStage is exited.
func (s *BaseSPLParserListener) ExitAnalysisDedupStage(ctx *AnalysisDedupStageContext) {}

// EnterAnalysisLimitStage is called when production AnalysisLimitStage is entered.
func (s *BaseSPLParserListener) EnterAnalysisLimitStage(ctx *AnalysisLimitStageContext) {}

// ExitAnalysisLimitStage is called when production AnalysisLimitStage is exited.
func (s *BaseSPLParserListener) ExitAnalysisLimitStage(ctx *AnalysisLimitStageContext) {}

// EnterAnalysisInputlookupStage is called when production AnalysisInputlookupStage is entered.
func (s *BaseSPLParserListener) EnterAnalysisInputlookupStage(ctx *AnalysisInputlookupStageContext) {}

// ExitAnalysisInputlookupStage is called when production AnalysisInputlookupStage is exited.
func (s *BaseSPLParserListener) ExitAnalysisInputlookupStage(ctx *AnalysisInputlookupStageContext) {}

// EnterAnalysisDatamodelStage is called when production AnalysisDatamodelStage is entered.
func (s *BaseSPLParserListener) EnterAnalysisDatamodelStage(ctx *AnalysisDatamodelStageContext) {}

// ExitAnalysisDatamodelStage is called when production AnalysisDatamodelStage is exited.
func (s *BaseSPLParserListener) ExitAnalysisDatamodelStage(ctx *AnalysisDatamodelStageContext) {}

// EnterAnalysisFromStage is called when production AnalysisFromStage is entered.
func (s *BaseSPLParserListener) EnterAnalysisFromStage(ctx *AnalysisFromStageContext) {}

// ExitAnalysisFromStage is called when production AnalysisFromStage is exited.
func (s *BaseSPLParserListener) ExitAnalysisFromStage(ctx *AnalysisFromStageContext) {}

// EnterAnalysisTstatsStage is called when production AnalysisTstatsStage is entered.
func (s *BaseSPLParserListener) EnterAnalysisTstatsStage(ctx *AnalysisTstatsStageContext) {}

// ExitAnalysisTstatsStage is called when production AnalysisTstatsStage is exited.
func (s *BaseSPLParserListener) ExitAnalysisTstatsStage(ctx *AnalysisTstatsStageContext) {}

// EnterAnalysisMacroStage is called when production AnalysisMacroStage is entered.
func (s *BaseSPLParserListener) EnterAnalysisMacroStage(ctx *AnalysisMacroStageContext) {}

// ExitAnalysisMacroStage is called when production AnalysisMacroStage is exited.
func (s *BaseSPLParserListener) ExitAnalysisMacroStage(ctx *AnalysisMacroStageContext) {}

// EnterAnalysisOpaqueStage is called when production AnalysisOpaqueStage is entered.
func (s *BaseSPLParserListener) EnterAnalysisOpaqueStage(ctx *AnalysisOpaqueStageContext) {}

// ExitAnalysisOpaqueStage is called when production AnalysisOpaqueStage is exited.
func (s *BaseSPLParserListener) ExitAnalysisOpaqueStage(ctx *AnalysisOpaqueStageContext) {}

// EnterAnalysisDataModelName is called when production analysisDataModelName is entered.
func (s *BaseSPLParserListener) EnterAnalysisDataModelName(ctx *AnalysisDataModelNameContext) {}

// ExitAnalysisDataModelName is called when production analysisDataModelName is exited.
func (s *BaseSPLParserListener) ExitAnalysisDataModelName(ctx *AnalysisDataModelNameContext) {}

// EnterAnalysisDataModelDataset is called when production analysisDataModelDataset is entered.
func (s *BaseSPLParserListener) EnterAnalysisDataModelDataset(ctx *AnalysisDataModelDatasetContext) {}

// ExitAnalysisDataModelDataset is called when production analysisDataModelDataset is exited.
func (s *BaseSPLParserListener) ExitAnalysisDataModelDataset(ctx *AnalysisDataModelDatasetContext) {}

// EnterAnalysisFromDataset is called when production analysisFromDataset is entered.
func (s *BaseSPLParserListener) EnterAnalysisFromDataset(ctx *AnalysisFromDatasetContext) {}

// ExitAnalysisFromDataset is called when production analysisFromDataset is exited.
func (s *BaseSPLParserListener) ExitAnalysisFromDataset(ctx *AnalysisFromDatasetContext) {}

// EnterAnalysisTstatsFrom is called when production analysisTstatsFrom is entered.
func (s *BaseSPLParserListener) EnterAnalysisTstatsFrom(ctx *AnalysisTstatsFromContext) {}

// ExitAnalysisTstatsFrom is called when production analysisTstatsFrom is exited.
func (s *BaseSPLParserListener) ExitAnalysisTstatsFrom(ctx *AnalysisTstatsFromContext) {}

// EnterAnalysisTstatsWhere is called when production analysisTstatsWhere is entered.
func (s *BaseSPLParserListener) EnterAnalysisTstatsWhere(ctx *AnalysisTstatsWhereContext) {}

// ExitAnalysisTstatsWhere is called when production analysisTstatsWhere is exited.
func (s *BaseSPLParserListener) ExitAnalysisTstatsWhere(ctx *AnalysisTstatsWhereContext) {}

// EnterAnalysisCommandName is called when production analysisCommandName is entered.
func (s *BaseSPLParserListener) EnterAnalysisCommandName(ctx *AnalysisCommandNameContext) {}

// ExitAnalysisCommandName is called when production analysisCommandName is exited.
func (s *BaseSPLParserListener) ExitAnalysisCommandName(ctx *AnalysisCommandNameContext) {}

// EnterAnalysisSortField is called when production analysisSortField is entered.
func (s *BaseSPLParserListener) EnterAnalysisSortField(ctx *AnalysisSortFieldContext) {}

// ExitAnalysisSortField is called when production analysisSortField is exited.
func (s *BaseSPLParserListener) ExitAnalysisSortField(ctx *AnalysisSortFieldContext) {}

// EnterAnalysisLimit is called when production analysisLimit is entered.
func (s *BaseSPLParserListener) EnterAnalysisLimit(ctx *AnalysisLimitContext) {}

// ExitAnalysisLimit is called when production analysisLimit is exited.
func (s *BaseSPLParserListener) ExitAnalysisLimit(ctx *AnalysisLimitContext) {}

// EnterAnalysisOption is called when production analysisOption is entered.
func (s *BaseSPLParserListener) EnterAnalysisOption(ctx *AnalysisOptionContext) {}

// ExitAnalysisOption is called when production analysisOption is exited.
func (s *BaseSPLParserListener) ExitAnalysisOption(ctx *AnalysisOptionContext) {}

// EnterAnalysisCatalogName is called when production analysisCatalogName is entered.
func (s *BaseSPLParserListener) EnterAnalysisCatalogName(ctx *AnalysisCatalogNameContext) {}

// ExitAnalysisCatalogName is called when production analysisCatalogName is exited.
func (s *BaseSPLParserListener) ExitAnalysisCatalogName(ctx *AnalysisCatalogNameContext) {}

// EnterAnalysisAssignment is called when production analysisAssignment is entered.
func (s *BaseSPLParserListener) EnterAnalysisAssignment(ctx *AnalysisAssignmentContext) {}

// ExitAnalysisAssignment is called when production analysisAssignment is exited.
func (s *BaseSPLParserListener) ExitAnalysisAssignment(ctx *AnalysisAssignmentContext) {}

// EnterAnalysisRename is called when production analysisRename is entered.
func (s *BaseSPLParserListener) EnterAnalysisRename(ctx *AnalysisRenameContext) {}

// ExitAnalysisRename is called when production analysisRename is exited.
func (s *BaseSPLParserListener) ExitAnalysisRename(ctx *AnalysisRenameContext) {}

// EnterAnalysisAlias is called when production analysisAlias is entered.
func (s *BaseSPLParserListener) EnterAnalysisAlias(ctx *AnalysisAliasContext) {}

// ExitAnalysisAlias is called when production analysisAlias is exited.
func (s *BaseSPLParserListener) ExitAnalysisAlias(ctx *AnalysisAliasContext) {}

// EnterAnalysisFieldList is called when production analysisFieldList is entered.
func (s *BaseSPLParserListener) EnterAnalysisFieldList(ctx *AnalysisFieldListContext) {}

// ExitAnalysisFieldList is called when production analysisFieldList is exited.
func (s *BaseSPLParserListener) ExitAnalysisFieldList(ctx *AnalysisFieldListContext) {}

// EnterAnalysisGroup is called when production analysisGroup is entered.
func (s *BaseSPLParserListener) EnterAnalysisGroup(ctx *AnalysisGroupContext) {}

// ExitAnalysisGroup is called when production analysisGroup is exited.
func (s *BaseSPLParserListener) ExitAnalysisGroup(ctx *AnalysisGroupContext) {}

// EnterAnalysisAggregate is called when production analysisAggregate is entered.
func (s *BaseSPLParserListener) EnterAnalysisAggregate(ctx *AnalysisAggregateContext) {}

// ExitAnalysisAggregate is called when production analysisAggregate is exited.
func (s *BaseSPLParserListener) ExitAnalysisAggregate(ctx *AnalysisAggregateContext) {}

// EnterAnalysisLookup is called when production analysisLookup is entered.
func (s *BaseSPLParserListener) EnterAnalysisLookup(ctx *AnalysisLookupContext) {}

// ExitAnalysisLookup is called when production analysisLookup is exited.
func (s *BaseSPLParserListener) ExitAnalysisLookup(ctx *AnalysisLookupContext) {}

// EnterAnalysisLookupInput is called when production analysisLookupInput is entered.
func (s *BaseSPLParserListener) EnterAnalysisLookupInput(ctx *AnalysisLookupInputContext) {}

// ExitAnalysisLookupInput is called when production analysisLookupInput is exited.
func (s *BaseSPLParserListener) ExitAnalysisLookupInput(ctx *AnalysisLookupInputContext) {}

// EnterAnalysisOutput is called when production analysisOutput is entered.
func (s *BaseSPLParserListener) EnterAnalysisOutput(ctx *AnalysisOutputContext) {}

// ExitAnalysisOutput is called when production analysisOutput is exited.
func (s *BaseSPLParserListener) ExitAnalysisOutput(ctx *AnalysisOutputContext) {}

// EnterAnalysisArgument is called when production analysisArgument is entered.
func (s *BaseSPLParserListener) EnterAnalysisArgument(ctx *AnalysisArgumentContext) {}

// ExitAnalysisArgument is called when production analysisArgument is exited.
func (s *BaseSPLParserListener) ExitAnalysisArgument(ctx *AnalysisArgumentContext) {}

// EnterAnalysisSubquery is called when production analysisSubquery is entered.
func (s *BaseSPLParserListener) EnterAnalysisSubquery(ctx *AnalysisSubqueryContext) {}

// ExitAnalysisSubquery is called when production analysisSubquery is exited.
func (s *BaseSPLParserListener) ExitAnalysisSubquery(ctx *AnalysisSubqueryContext) {}

// EnterAnalysisMacro is called when production analysisMacro is entered.
func (s *BaseSPLParserListener) EnterAnalysisMacro(ctx *AnalysisMacroContext) {}

// ExitAnalysisMacro is called when production analysisMacro is exited.
func (s *BaseSPLParserListener) ExitAnalysisMacro(ctx *AnalysisMacroContext) {}

// EnterAnalysisArgumentList is called when production analysisArgumentList is entered.
func (s *BaseSPLParserListener) EnterAnalysisArgumentList(ctx *AnalysisArgumentListContext) {}

// ExitAnalysisArgumentList is called when production analysisArgumentList is exited.
func (s *BaseSPLParserListener) ExitAnalysisArgumentList(ctx *AnalysisArgumentListContext) {}

// EnterAnalysisSearch is called when production analysisSearch is entered.
func (s *BaseSPLParserListener) EnterAnalysisSearch(ctx *AnalysisSearchContext) {}

// ExitAnalysisSearch is called when production analysisSearch is exited.
func (s *BaseSPLParserListener) ExitAnalysisSearch(ctx *AnalysisSearchContext) {}

// EnterAnalysisSearchAnd is called when production analysisSearchAnd is entered.
func (s *BaseSPLParserListener) EnterAnalysisSearchAnd(ctx *AnalysisSearchAndContext) {}

// ExitAnalysisSearchAnd is called when production analysisSearchAnd is exited.
func (s *BaseSPLParserListener) ExitAnalysisSearchAnd(ctx *AnalysisSearchAndContext) {}

// EnterAnalysisSearchUnary is called when production analysisSearchUnary is entered.
func (s *BaseSPLParserListener) EnterAnalysisSearchUnary(ctx *AnalysisSearchUnaryContext) {}

// ExitAnalysisSearchUnary is called when production analysisSearchUnary is exited.
func (s *BaseSPLParserListener) ExitAnalysisSearchUnary(ctx *AnalysisSearchUnaryContext) {}

// EnterAnalysisSearchTerm is called when production analysisSearchTerm is entered.
func (s *BaseSPLParserListener) EnterAnalysisSearchTerm(ctx *AnalysisSearchTermContext) {}

// ExitAnalysisSearchTerm is called when production analysisSearchTerm is exited.
func (s *BaseSPLParserListener) ExitAnalysisSearchTerm(ctx *AnalysisSearchTermContext) {}

// EnterAnalysisSearchValue is called when production analysisSearchValue is entered.
func (s *BaseSPLParserListener) EnterAnalysisSearchValue(ctx *AnalysisSearchValueContext) {}

// ExitAnalysisSearchValue is called when production analysisSearchValue is exited.
func (s *BaseSPLParserListener) ExitAnalysisSearchValue(ctx *AnalysisSearchValueContext) {}

// EnterAnalysisUnquotedValue is called when production analysisUnquotedValue is entered.
func (s *BaseSPLParserListener) EnterAnalysisUnquotedValue(ctx *AnalysisUnquotedValueContext) {}

// ExitAnalysisUnquotedValue is called when production analysisUnquotedValue is exited.
func (s *BaseSPLParserListener) ExitAnalysisUnquotedValue(ctx *AnalysisUnquotedValueContext) {}

// EnterAnalysisUnquotedPart is called when production analysisUnquotedPart is entered.
func (s *BaseSPLParserListener) EnterAnalysisUnquotedPart(ctx *AnalysisUnquotedPartContext) {}

// ExitAnalysisUnquotedPart is called when production analysisUnquotedPart is exited.
func (s *BaseSPLParserListener) ExitAnalysisUnquotedPart(ctx *AnalysisUnquotedPartContext) {}

// EnterAnalysisExpression is called when production analysisExpression is entered.
func (s *BaseSPLParserListener) EnterAnalysisExpression(ctx *AnalysisExpressionContext) {}

// ExitAnalysisExpression is called when production analysisExpression is exited.
func (s *BaseSPLParserListener) ExitAnalysisExpression(ctx *AnalysisExpressionContext) {}

// EnterAnalysisOr is called when production analysisOr is entered.
func (s *BaseSPLParserListener) EnterAnalysisOr(ctx *AnalysisOrContext) {}

// ExitAnalysisOr is called when production analysisOr is exited.
func (s *BaseSPLParserListener) ExitAnalysisOr(ctx *AnalysisOrContext) {}

// EnterAnalysisAnd is called when production analysisAnd is entered.
func (s *BaseSPLParserListener) EnterAnalysisAnd(ctx *AnalysisAndContext) {}

// ExitAnalysisAnd is called when production analysisAnd is exited.
func (s *BaseSPLParserListener) ExitAnalysisAnd(ctx *AnalysisAndContext) {}

// EnterAnalysisNot is called when production analysisNot is entered.
func (s *BaseSPLParserListener) EnterAnalysisNot(ctx *AnalysisNotContext) {}

// ExitAnalysisNot is called when production analysisNot is exited.
func (s *BaseSPLParserListener) ExitAnalysisNot(ctx *AnalysisNotContext) {}

// EnterAnalysisComparison is called when production analysisComparison is entered.
func (s *BaseSPLParserListener) EnterAnalysisComparison(ctx *AnalysisComparisonContext) {}

// ExitAnalysisComparison is called when production analysisComparison is exited.
func (s *BaseSPLParserListener) ExitAnalysisComparison(ctx *AnalysisComparisonContext) {}

// EnterAnalysisComparisonOperator is called when production analysisComparisonOperator is entered.
func (s *BaseSPLParserListener) EnterAnalysisComparisonOperator(ctx *AnalysisComparisonOperatorContext) {
}

// ExitAnalysisComparisonOperator is called when production analysisComparisonOperator is exited.
func (s *BaseSPLParserListener) ExitAnalysisComparisonOperator(ctx *AnalysisComparisonOperatorContext) {
}

// EnterAnalysisConcat is called when production analysisConcat is entered.
func (s *BaseSPLParserListener) EnterAnalysisConcat(ctx *AnalysisConcatContext) {}

// ExitAnalysisConcat is called when production analysisConcat is exited.
func (s *BaseSPLParserListener) ExitAnalysisConcat(ctx *AnalysisConcatContext) {}

// EnterAnalysisAdd is called when production analysisAdd is entered.
func (s *BaseSPLParserListener) EnterAnalysisAdd(ctx *AnalysisAddContext) {}

// ExitAnalysisAdd is called when production analysisAdd is exited.
func (s *BaseSPLParserListener) ExitAnalysisAdd(ctx *AnalysisAddContext) {}

// EnterAnalysisMultiply is called when production analysisMultiply is entered.
func (s *BaseSPLParserListener) EnterAnalysisMultiply(ctx *AnalysisMultiplyContext) {}

// ExitAnalysisMultiply is called when production analysisMultiply is exited.
func (s *BaseSPLParserListener) ExitAnalysisMultiply(ctx *AnalysisMultiplyContext) {}

// EnterAnalysisPower is called when production analysisPower is entered.
func (s *BaseSPLParserListener) EnterAnalysisPower(ctx *AnalysisPowerContext) {}

// ExitAnalysisPower is called when production analysisPower is exited.
func (s *BaseSPLParserListener) ExitAnalysisPower(ctx *AnalysisPowerContext) {}

// EnterAnalysisUnary is called when production analysisUnary is entered.
func (s *BaseSPLParserListener) EnterAnalysisUnary(ctx *AnalysisUnaryContext) {}

// ExitAnalysisUnary is called when production analysisUnary is exited.
func (s *BaseSPLParserListener) ExitAnalysisUnary(ctx *AnalysisUnaryContext) {}

// EnterAnalysisAtom is called when production analysisAtom is entered.
func (s *BaseSPLParserListener) EnterAnalysisAtom(ctx *AnalysisAtomContext) {}

// ExitAnalysisAtom is called when production analysisAtom is exited.
func (s *BaseSPLParserListener) ExitAnalysisAtom(ctx *AnalysisAtomContext) {}

// EnterAnalysisFunctionCall is called when production analysisFunctionCall is entered.
func (s *BaseSPLParserListener) EnterAnalysisFunctionCall(ctx *AnalysisFunctionCallContext) {}

// ExitAnalysisFunctionCall is called when production analysisFunctionCall is exited.
func (s *BaseSPLParserListener) ExitAnalysisFunctionCall(ctx *AnalysisFunctionCallContext) {}

// EnterAnalysisFunctionName is called when production analysisFunctionName is entered.
func (s *BaseSPLParserListener) EnterAnalysisFunctionName(ctx *AnalysisFunctionNameContext) {}

// ExitAnalysisFunctionName is called when production analysisFunctionName is exited.
func (s *BaseSPLParserListener) ExitAnalysisFunctionName(ctx *AnalysisFunctionNameContext) {}

// EnterAnalysisLiteral is called when production analysisLiteral is entered.
func (s *BaseSPLParserListener) EnterAnalysisLiteral(ctx *AnalysisLiteralContext) {}

// ExitAnalysisLiteral is called when production analysisLiteral is exited.
func (s *BaseSPLParserListener) ExitAnalysisLiteral(ctx *AnalysisLiteralContext) {}

// EnterAnalysisSelector is called when production analysisSelector is entered.
func (s *BaseSPLParserListener) EnterAnalysisSelector(ctx *AnalysisSelectorContext) {}

// ExitAnalysisSelector is called when production analysisSelector is exited.
func (s *BaseSPLParserListener) ExitAnalysisSelector(ctx *AnalysisSelectorContext) {}

// EnterAnalysisIdentifier is called when production analysisIdentifier is entered.
func (s *BaseSPLParserListener) EnterAnalysisIdentifier(ctx *AnalysisIdentifierContext) {}

// ExitAnalysisIdentifier is called when production analysisIdentifier is exited.
func (s *BaseSPLParserListener) ExitAnalysisIdentifier(ctx *AnalysisIdentifierContext) {}
