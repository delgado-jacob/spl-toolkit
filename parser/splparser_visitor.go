// Code generated from grammar/SPLParser.g4 by ANTLR 4.13.2. DO NOT EDIT.

package parser // SPLParser
import "github.com/antlr4-go/antlr/v4"

// A complete Visitor for a parse tree produced by SPLParser.
type SPLParserVisitor interface {
	antlr.ParseTreeVisitor

	// Visit a parse tree produced by SPLParser#query.
	VisitQuery(ctx *QueryContext) interface{}

	// Visit a parse tree produced by SPLParser#initCommand.
	VisitInitCommand(ctx *InitCommandContext) interface{}

	// Visit a parse tree produced by SPLParser#nextCommand.
	VisitNextCommand(ctx *NextCommandContext) interface{}

	// Visit a parse tree produced by SPLParser#subquery.
	VisitSubquery(ctx *SubqueryContext) interface{}

	// Visit a parse tree produced by SPLParser#LIKEOP.
	VisitLIKEOP(ctx *LIKEOPContext) interface{}

	// Visit a parse tree produced by SPLParser#ANDOP.
	VisitANDOP(ctx *ANDOPContext) interface{}

	// Visit a parse tree produced by SPLParser#OROP.
	VisitOROP(ctx *OROPContext) interface{}

	// Visit a parse tree produced by SPLParser#INOP.
	VisitINOP(ctx *INOPContext) interface{}

	// Visit a parse tree produced by SPLParser#NOTOP.
	VisitNOTOP(ctx *NOTOPContext) interface{}

	// Visit a parse tree produced by SPLParser#BYOP.
	VisitBYOP(ctx *BYOPContext) interface{}

	// Visit a parse tree produced by SPLParser#PARENOP.
	VisitPARENOP(ctx *PARENOPContext) interface{}

	// Visit a parse tree produced by SPLParser#OUTPUTMULTIINOP.
	VisitOUTPUTMULTIINOP(ctx *OUTPUTMULTIINOPContext) interface{}

	// Visit a parse tree produced by SPLParser#KEYVALUEOP.
	VisitKEYVALUEOP(ctx *KEYVALUEOPContext) interface{}

	// Visit a parse tree produced by SPLParser#OUTPUTOP.
	VisitOUTPUTOP(ctx *OUTPUTOPContext) interface{}

	// Visit a parse tree produced by SPLParser#EXPRESSIONOP.
	VisitEXPRESSIONOP(ctx *EXPRESSIONOPContext) interface{}

	// Visit a parse tree produced by SPLParser#OUTPUTMULTIOP.
	VisitOUTPUTMULTIOP(ctx *OUTPUTMULTIOPContext) interface{}

	// Visit a parse tree produced by SPLParser#RENAMEOP.
	VisitRENAMEOP(ctx *RENAMEOPContext) interface{}

	// Visit a parse tree produced by SPLParser#expression.
	VisitExpression(ctx *ExpressionContext) interface{}

	// Visit a parse tree produced by SPLParser#value.
	VisitValue(ctx *ValueContext) interface{}

	// Visit a parse tree produced by SPLParser#date.
	VisitDate(ctx *DateContext) interface{}

	// Visit a parse tree produced by SPLParser#FieldUse.
	VisitFieldUse(ctx *FieldUseContext) interface{}

	// Visit a parse tree produced by SPLParser#CommandUse.
	VisitCommandUse(ctx *CommandUseContext) interface{}

	// Visit a parse tree produced by SPLParser#FunctionUse.
	VisitFunctionUse(ctx *FunctionUseContext) interface{}

	// Visit a parse tree produced by SPLParser#function.
	VisitFunction(ctx *FunctionContext) interface{}

	// Visit a parse tree produced by SPLParser#command.
	VisitCommand(ctx *CommandContext) interface{}

	// Visit a parse tree produced by SPLParser#analysisQuery.
	VisitAnalysisQuery(ctx *AnalysisQueryContext) interface{}

	// Visit a parse tree produced by SPLParser#analysisPipeline.
	VisitAnalysisPipeline(ctx *AnalysisPipelineContext) interface{}

	// Visit a parse tree produced by SPLParser#analysisInitialStage.
	VisitAnalysisInitialStage(ctx *AnalysisInitialStageContext) interface{}

	// Visit a parse tree produced by SPLParser#analysisImplicitSearch.
	VisitAnalysisImplicitSearch(ctx *AnalysisImplicitSearchContext) interface{}

	// Visit a parse tree produced by SPLParser#AnalysisSearchStage.
	VisitAnalysisSearchStage(ctx *AnalysisSearchStageContext) interface{}

	// Visit a parse tree produced by SPLParser#AnalysisWhereStage.
	VisitAnalysisWhereStage(ctx *AnalysisWhereStageContext) interface{}

	// Visit a parse tree produced by SPLParser#AnalysisEvalStage.
	VisitAnalysisEvalStage(ctx *AnalysisEvalStageContext) interface{}

	// Visit a parse tree produced by SPLParser#AnalysisRenameStage.
	VisitAnalysisRenameStage(ctx *AnalysisRenameStageContext) interface{}

	// Visit a parse tree produced by SPLParser#AnalysisFieldsStage.
	VisitAnalysisFieldsStage(ctx *AnalysisFieldsStageContext) interface{}

	// Visit a parse tree produced by SPLParser#AnalysisStatsStage.
	VisitAnalysisStatsStage(ctx *AnalysisStatsStageContext) interface{}

	// Visit a parse tree produced by SPLParser#AnalysisLookupStage.
	VisitAnalysisLookupStage(ctx *AnalysisLookupStageContext) interface{}

	// Visit a parse tree produced by SPLParser#AnalysisSortStage.
	VisitAnalysisSortStage(ctx *AnalysisSortStageContext) interface{}

	// Visit a parse tree produced by SPLParser#AnalysisDedupStage.
	VisitAnalysisDedupStage(ctx *AnalysisDedupStageContext) interface{}

	// Visit a parse tree produced by SPLParser#AnalysisLimitStage.
	VisitAnalysisLimitStage(ctx *AnalysisLimitStageContext) interface{}

	// Visit a parse tree produced by SPLParser#AnalysisInputlookupStage.
	VisitAnalysisInputlookupStage(ctx *AnalysisInputlookupStageContext) interface{}

	// Visit a parse tree produced by SPLParser#AnalysisDatamodelStage.
	VisitAnalysisDatamodelStage(ctx *AnalysisDatamodelStageContext) interface{}

	// Visit a parse tree produced by SPLParser#AnalysisFromStage.
	VisitAnalysisFromStage(ctx *AnalysisFromStageContext) interface{}

	// Visit a parse tree produced by SPLParser#AnalysisTstatsStage.
	VisitAnalysisTstatsStage(ctx *AnalysisTstatsStageContext) interface{}

	// Visit a parse tree produced by SPLParser#AnalysisFillnullStage.
	VisitAnalysisFillnullStage(ctx *AnalysisFillnullStageContext) interface{}

	// Visit a parse tree produced by SPLParser#AnalysisRexStage.
	VisitAnalysisRexStage(ctx *AnalysisRexStageContext) interface{}

	// Visit a parse tree produced by SPLParser#AnalysisSpathStage.
	VisitAnalysisSpathStage(ctx *AnalysisSpathStageContext) interface{}

	// Visit a parse tree produced by SPLParser#AnalysisBinStage.
	VisitAnalysisBinStage(ctx *AnalysisBinStageContext) interface{}

	// Visit a parse tree produced by SPLParser#AnalysisRegexStage.
	VisitAnalysisRegexStage(ctx *AnalysisRegexStageContext) interface{}

	// Visit a parse tree produced by SPLParser#AnalysisMvexpandStage.
	VisitAnalysisMvexpandStage(ctx *AnalysisMvexpandStageContext) interface{}

	// Visit a parse tree produced by SPLParser#AnalysisJoinStage.
	VisitAnalysisJoinStage(ctx *AnalysisJoinStageContext) interface{}

	// Visit a parse tree produced by SPLParser#AnalysisBranchStage.
	VisitAnalysisBranchStage(ctx *AnalysisBranchStageContext) interface{}

	// Visit a parse tree produced by SPLParser#AnalysisMacroStage.
	VisitAnalysisMacroStage(ctx *AnalysisMacroStageContext) interface{}

	// Visit a parse tree produced by SPLParser#AnalysisOpaqueStage.
	VisitAnalysisOpaqueStage(ctx *AnalysisOpaqueStageContext) interface{}

	// Visit a parse tree produced by SPLParser#analysisTstats.
	VisitAnalysisTstats(ctx *AnalysisTstatsContext) interface{}

	// Visit a parse tree produced by SPLParser#analysisTstatsOption.
	VisitAnalysisTstatsOption(ctx *AnalysisTstatsOptionContext) interface{}

	// Visit a parse tree produced by SPLParser#analysisTstatsGroup.
	VisitAnalysisTstatsGroup(ctx *AnalysisTstatsGroupContext) interface{}

	// Visit a parse tree produced by SPLParser#analysisTstatsGroupItem.
	VisitAnalysisTstatsGroupItem(ctx *AnalysisTstatsGroupItemContext) interface{}

	// Visit a parse tree produced by SPLParser#analysisTstatsSpanOption.
	VisitAnalysisTstatsSpanOption(ctx *AnalysisTstatsSpanOptionContext) interface{}

	// Visit a parse tree produced by SPLParser#analysisFillnull.
	VisitAnalysisFillnull(ctx *AnalysisFillnullContext) interface{}

	// Visit a parse tree produced by SPLParser#analysisFillnullValueOption.
	VisitAnalysisFillnullValueOption(ctx *AnalysisFillnullValueOptionContext) interface{}

	// Visit a parse tree produced by SPLParser#analysisRex.
	VisitAnalysisRex(ctx *AnalysisRexContext) interface{}

	// Visit a parse tree produced by SPLParser#analysisRexFieldOption.
	VisitAnalysisRexFieldOption(ctx *AnalysisRexFieldOptionContext) interface{}

	// Visit a parse tree produced by SPLParser#analysisRexMaxMatchOption.
	VisitAnalysisRexMaxMatchOption(ctx *AnalysisRexMaxMatchOptionContext) interface{}

	// Visit a parse tree produced by SPLParser#analysisRexOffsetFieldOption.
	VisitAnalysisRexOffsetFieldOption(ctx *AnalysisRexOffsetFieldOptionContext) interface{}

	// Visit a parse tree produced by SPLParser#analysisRexModeOption.
	VisitAnalysisRexModeOption(ctx *AnalysisRexModeOptionContext) interface{}

	// Visit a parse tree produced by SPLParser#analysisSpath.
	VisitAnalysisSpath(ctx *AnalysisSpathContext) interface{}

	// Visit a parse tree produced by SPLParser#analysisSpathInputOption.
	VisitAnalysisSpathInputOption(ctx *AnalysisSpathInputOptionContext) interface{}

	// Visit a parse tree produced by SPLParser#analysisSpathPathOption.
	VisitAnalysisSpathPathOption(ctx *AnalysisSpathPathOptionContext) interface{}

	// Visit a parse tree produced by SPLParser#analysisSpathOutputOption.
	VisitAnalysisSpathOutputOption(ctx *AnalysisSpathOutputOptionContext) interface{}

	// Visit a parse tree produced by SPLParser#analysisBin.
	VisitAnalysisBin(ctx *AnalysisBinContext) interface{}

	// Visit a parse tree produced by SPLParser#analysisBinOption.
	VisitAnalysisBinOption(ctx *AnalysisBinOptionContext) interface{}

	// Visit a parse tree produced by SPLParser#analysisRegex.
	VisitAnalysisRegex(ctx *AnalysisRegexContext) interface{}

	// Visit a parse tree produced by SPLParser#analysisMvexpand.
	VisitAnalysisMvexpand(ctx *AnalysisMvexpandContext) interface{}

	// Visit a parse tree produced by SPLParser#analysisMvexpandOption.
	VisitAnalysisMvexpandOption(ctx *AnalysisMvexpandOptionContext) interface{}

	// Visit a parse tree produced by SPLParser#analysisJoin.
	VisitAnalysisJoin(ctx *AnalysisJoinContext) interface{}

	// Visit a parse tree produced by SPLParser#analysisJoinOption.
	VisitAnalysisJoinOption(ctx *AnalysisJoinOptionContext) interface{}

	// Visit a parse tree produced by SPLParser#analysisBranch.
	VisitAnalysisBranch(ctx *AnalysisBranchContext) interface{}

	// Visit a parse tree produced by SPLParser#analysisBranchOption.
	VisitAnalysisBranchOption(ctx *AnalysisBranchOptionContext) interface{}

	// Visit a parse tree produced by SPLParser#analysisOptionValue.
	VisitAnalysisOptionValue(ctx *AnalysisOptionValueContext) interface{}

	// Visit a parse tree produced by SPLParser#analysisUnitOptionValue.
	VisitAnalysisUnitOptionValue(ctx *AnalysisUnitOptionValueContext) interface{}

	// Visit a parse tree produced by SPLParser#analysisUnitSuffix.
	VisitAnalysisUnitSuffix(ctx *AnalysisUnitSuffixContext) interface{}

	// Visit a parse tree produced by SPLParser#analysisInvalidOptionValue.
	VisitAnalysisInvalidOptionValue(ctx *AnalysisInvalidOptionValueContext) interface{}

	// Visit a parse tree produced by SPLParser#analysisMissingOptionValue.
	VisitAnalysisMissingOptionValue(ctx *AnalysisMissingOptionValueContext) interface{}

	// Visit a parse tree produced by SPLParser#analysisMissingJoinKey.
	VisitAnalysisMissingJoinKey(ctx *AnalysisMissingJoinKeyContext) interface{}

	// Visit a parse tree produced by SPLParser#analysisMissingOperand.
	VisitAnalysisMissingOperand(ctx *AnalysisMissingOperandContext) interface{}

	// Visit a parse tree produced by SPLParser#analysisDataModelName.
	VisitAnalysisDataModelName(ctx *AnalysisDataModelNameContext) interface{}

	// Visit a parse tree produced by SPLParser#analysisDataModelDataset.
	VisitAnalysisDataModelDataset(ctx *AnalysisDataModelDatasetContext) interface{}

	// Visit a parse tree produced by SPLParser#analysisFromDataset.
	VisitAnalysisFromDataset(ctx *AnalysisFromDatasetContext) interface{}

	// Visit a parse tree produced by SPLParser#analysisTstatsFrom.
	VisitAnalysisTstatsFrom(ctx *AnalysisTstatsFromContext) interface{}

	// Visit a parse tree produced by SPLParser#analysisTstatsWhere.
	VisitAnalysisTstatsWhere(ctx *AnalysisTstatsWhereContext) interface{}

	// Visit a parse tree produced by SPLParser#analysisCommandName.
	VisitAnalysisCommandName(ctx *AnalysisCommandNameContext) interface{}

	// Visit a parse tree produced by SPLParser#analysisSortField.
	VisitAnalysisSortField(ctx *AnalysisSortFieldContext) interface{}

	// Visit a parse tree produced by SPLParser#analysisLimit.
	VisitAnalysisLimit(ctx *AnalysisLimitContext) interface{}

	// Visit a parse tree produced by SPLParser#analysisOption.
	VisitAnalysisOption(ctx *AnalysisOptionContext) interface{}

	// Visit a parse tree produced by SPLParser#analysisCatalogName.
	VisitAnalysisCatalogName(ctx *AnalysisCatalogNameContext) interface{}

	// Visit a parse tree produced by SPLParser#analysisAssignment.
	VisitAnalysisAssignment(ctx *AnalysisAssignmentContext) interface{}

	// Visit a parse tree produced by SPLParser#analysisRename.
	VisitAnalysisRename(ctx *AnalysisRenameContext) interface{}

	// Visit a parse tree produced by SPLParser#analysisAlias.
	VisitAnalysisAlias(ctx *AnalysisAliasContext) interface{}

	// Visit a parse tree produced by SPLParser#analysisFieldList.
	VisitAnalysisFieldList(ctx *AnalysisFieldListContext) interface{}

	// Visit a parse tree produced by SPLParser#analysisGroup.
	VisitAnalysisGroup(ctx *AnalysisGroupContext) interface{}

	// Visit a parse tree produced by SPLParser#analysisAggregate.
	VisitAnalysisAggregate(ctx *AnalysisAggregateContext) interface{}

	// Visit a parse tree produced by SPLParser#analysisLookup.
	VisitAnalysisLookup(ctx *AnalysisLookupContext) interface{}

	// Visit a parse tree produced by SPLParser#analysisLookupInput.
	VisitAnalysisLookupInput(ctx *AnalysisLookupInputContext) interface{}

	// Visit a parse tree produced by SPLParser#analysisOutput.
	VisitAnalysisOutput(ctx *AnalysisOutputContext) interface{}

	// Visit a parse tree produced by SPLParser#analysisArgument.
	VisitAnalysisArgument(ctx *AnalysisArgumentContext) interface{}

	// Visit a parse tree produced by SPLParser#analysisSubquery.
	VisitAnalysisSubquery(ctx *AnalysisSubqueryContext) interface{}

	// Visit a parse tree produced by SPLParser#analysisMacro.
	VisitAnalysisMacro(ctx *AnalysisMacroContext) interface{}

	// Visit a parse tree produced by SPLParser#analysisArgumentList.
	VisitAnalysisArgumentList(ctx *AnalysisArgumentListContext) interface{}

	// Visit a parse tree produced by SPLParser#analysisSearch.
	VisitAnalysisSearch(ctx *AnalysisSearchContext) interface{}

	// Visit a parse tree produced by SPLParser#analysisSearchAnd.
	VisitAnalysisSearchAnd(ctx *AnalysisSearchAndContext) interface{}

	// Visit a parse tree produced by SPLParser#analysisSearchUnary.
	VisitAnalysisSearchUnary(ctx *AnalysisSearchUnaryContext) interface{}

	// Visit a parse tree produced by SPLParser#analysisSearchTerm.
	VisitAnalysisSearchTerm(ctx *AnalysisSearchTermContext) interface{}

	// Visit a parse tree produced by SPLParser#analysisSearchValue.
	VisitAnalysisSearchValue(ctx *AnalysisSearchValueContext) interface{}

	// Visit a parse tree produced by SPLParser#analysisUnquotedValue.
	VisitAnalysisUnquotedValue(ctx *AnalysisUnquotedValueContext) interface{}

	// Visit a parse tree produced by SPLParser#analysisUnquotedPart.
	VisitAnalysisUnquotedPart(ctx *AnalysisUnquotedPartContext) interface{}

	// Visit a parse tree produced by SPLParser#analysisExpression.
	VisitAnalysisExpression(ctx *AnalysisExpressionContext) interface{}

	// Visit a parse tree produced by SPLParser#analysisOr.
	VisitAnalysisOr(ctx *AnalysisOrContext) interface{}

	// Visit a parse tree produced by SPLParser#analysisAnd.
	VisitAnalysisAnd(ctx *AnalysisAndContext) interface{}

	// Visit a parse tree produced by SPLParser#analysisNot.
	VisitAnalysisNot(ctx *AnalysisNotContext) interface{}

	// Visit a parse tree produced by SPLParser#analysisComparison.
	VisitAnalysisComparison(ctx *AnalysisComparisonContext) interface{}

	// Visit a parse tree produced by SPLParser#analysisComparisonOperator.
	VisitAnalysisComparisonOperator(ctx *AnalysisComparisonOperatorContext) interface{}

	// Visit a parse tree produced by SPLParser#analysisConcat.
	VisitAnalysisConcat(ctx *AnalysisConcatContext) interface{}

	// Visit a parse tree produced by SPLParser#analysisAdd.
	VisitAnalysisAdd(ctx *AnalysisAddContext) interface{}

	// Visit a parse tree produced by SPLParser#analysisMultiply.
	VisitAnalysisMultiply(ctx *AnalysisMultiplyContext) interface{}

	// Visit a parse tree produced by SPLParser#analysisPower.
	VisitAnalysisPower(ctx *AnalysisPowerContext) interface{}

	// Visit a parse tree produced by SPLParser#analysisUnary.
	VisitAnalysisUnary(ctx *AnalysisUnaryContext) interface{}

	// Visit a parse tree produced by SPLParser#analysisAtom.
	VisitAnalysisAtom(ctx *AnalysisAtomContext) interface{}

	// Visit a parse tree produced by SPLParser#analysisFunctionCall.
	VisitAnalysisFunctionCall(ctx *AnalysisFunctionCallContext) interface{}

	// Visit a parse tree produced by SPLParser#analysisFunctionName.
	VisitAnalysisFunctionName(ctx *AnalysisFunctionNameContext) interface{}

	// Visit a parse tree produced by SPLParser#analysisLiteral.
	VisitAnalysisLiteral(ctx *AnalysisLiteralContext) interface{}

	// Visit a parse tree produced by SPLParser#analysisSelector.
	VisitAnalysisSelector(ctx *AnalysisSelectorContext) interface{}

	// Visit a parse tree produced by SPLParser#analysisIdentifier.
	VisitAnalysisIdentifier(ctx *AnalysisIdentifierContext) interface{}
}
