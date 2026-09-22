// Code generated from grammar/SPLParser.g4 by ANTLR 4.13.2. DO NOT EDIT.

package parser // SPLParser
import "github.com/antlr4-go/antlr/v4"

type BaseSPLParserVisitor struct {
	*antlr.BaseParseTreeVisitor
}

func (v *BaseSPLParserVisitor) VisitQuery(ctx *QueryContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitInitCommand(ctx *InitCommandContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitNextCommand(ctx *NextCommandContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitSubquery(ctx *SubqueryContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitLIKEOP(ctx *LIKEOPContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitANDOP(ctx *ANDOPContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitOROP(ctx *OROPContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitINOP(ctx *INOPContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitNOTOP(ctx *NOTOPContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitBYOP(ctx *BYOPContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitPARENOP(ctx *PARENOPContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitOUTPUTMULTIINOP(ctx *OUTPUTMULTIINOPContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitKEYVALUEOP(ctx *KEYVALUEOPContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitOUTPUTOP(ctx *OUTPUTOPContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitEXPRESSIONOP(ctx *EXPRESSIONOPContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitOUTPUTMULTIOP(ctx *OUTPUTMULTIOPContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitRENAMEOP(ctx *RENAMEOPContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitExpression(ctx *ExpressionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitValue(ctx *ValueContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitDate(ctx *DateContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitFieldUse(ctx *FieldUseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitCommandUse(ctx *CommandUseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitFunctionUse(ctx *FunctionUseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitFunction(ctx *FunctionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitCommand(ctx *CommandContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisQuery(ctx *AnalysisQueryContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisPipeline(ctx *AnalysisPipelineContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisInitialStage(ctx *AnalysisInitialStageContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisImplicitSearch(ctx *AnalysisImplicitSearchContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisSearchStage(ctx *AnalysisSearchStageContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisWhereStage(ctx *AnalysisWhereStageContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisEvalStage(ctx *AnalysisEvalStageContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisRenameStage(ctx *AnalysisRenameStageContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisFieldsStage(ctx *AnalysisFieldsStageContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisStatsStage(ctx *AnalysisStatsStageContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisLookupStage(ctx *AnalysisLookupStageContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisSortStage(ctx *AnalysisSortStageContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisDedupStage(ctx *AnalysisDedupStageContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisLimitStage(ctx *AnalysisLimitStageContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisInputlookupStage(ctx *AnalysisInputlookupStageContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisDatamodelStage(ctx *AnalysisDatamodelStageContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisFromStage(ctx *AnalysisFromStageContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisTstatsStage(ctx *AnalysisTstatsStageContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisFillnullStage(ctx *AnalysisFillnullStageContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisRexStage(ctx *AnalysisRexStageContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisSpathStage(ctx *AnalysisSpathStageContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisBinStage(ctx *AnalysisBinStageContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisRegexStage(ctx *AnalysisRegexStageContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisMvexpandStage(ctx *AnalysisMvexpandStageContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisJoinStage(ctx *AnalysisJoinStageContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisBranchStage(ctx *AnalysisBranchStageContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisMacroStage(ctx *AnalysisMacroStageContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisOpaqueStage(ctx *AnalysisOpaqueStageContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisTstats(ctx *AnalysisTstatsContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisTstatsOption(ctx *AnalysisTstatsOptionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisTstatsGroup(ctx *AnalysisTstatsGroupContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisTstatsGroupItem(ctx *AnalysisTstatsGroupItemContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisTstatsSpanOption(ctx *AnalysisTstatsSpanOptionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisFillnull(ctx *AnalysisFillnullContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisFillnullValueOption(ctx *AnalysisFillnullValueOptionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisRex(ctx *AnalysisRexContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisRexFieldOption(ctx *AnalysisRexFieldOptionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisRexMaxMatchOption(ctx *AnalysisRexMaxMatchOptionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisRexOffsetFieldOption(ctx *AnalysisRexOffsetFieldOptionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisRexModeOption(ctx *AnalysisRexModeOptionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisSpath(ctx *AnalysisSpathContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisSpathInputOption(ctx *AnalysisSpathInputOptionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisSpathPathOption(ctx *AnalysisSpathPathOptionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisSpathOutputOption(ctx *AnalysisSpathOutputOptionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisBin(ctx *AnalysisBinContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisBinOption(ctx *AnalysisBinOptionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisRegex(ctx *AnalysisRegexContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisMvexpand(ctx *AnalysisMvexpandContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisMvexpandOption(ctx *AnalysisMvexpandOptionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisJoin(ctx *AnalysisJoinContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisJoinOption(ctx *AnalysisJoinOptionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisBranch(ctx *AnalysisBranchContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisBranchOption(ctx *AnalysisBranchOptionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisOptionValue(ctx *AnalysisOptionValueContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisUnitOptionValue(ctx *AnalysisUnitOptionValueContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisUnitSuffix(ctx *AnalysisUnitSuffixContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisInvalidOptionValue(ctx *AnalysisInvalidOptionValueContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisMissingOptionValue(ctx *AnalysisMissingOptionValueContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisMissingJoinKey(ctx *AnalysisMissingJoinKeyContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisMissingOperand(ctx *AnalysisMissingOperandContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisMissingEvalAssignment(ctx *AnalysisMissingEvalAssignmentContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisDataModelName(ctx *AnalysisDataModelNameContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisDataModelDataset(ctx *AnalysisDataModelDatasetContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisFromDataset(ctx *AnalysisFromDatasetContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisTstatsFrom(ctx *AnalysisTstatsFromContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisTstatsWhere(ctx *AnalysisTstatsWhereContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisCommandName(ctx *AnalysisCommandNameContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisSortField(ctx *AnalysisSortFieldContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisLimit(ctx *AnalysisLimitContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisOption(ctx *AnalysisOptionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisCatalogName(ctx *AnalysisCatalogNameContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisAssignment(ctx *AnalysisAssignmentContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisRename(ctx *AnalysisRenameContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisAlias(ctx *AnalysisAliasContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisFieldList(ctx *AnalysisFieldListContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisGroup(ctx *AnalysisGroupContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisAggregate(ctx *AnalysisAggregateContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisLookup(ctx *AnalysisLookupContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisLookupInput(ctx *AnalysisLookupInputContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisOutput(ctx *AnalysisOutputContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisArgument(ctx *AnalysisArgumentContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisSubquery(ctx *AnalysisSubqueryContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisMacro(ctx *AnalysisMacroContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisArgumentList(ctx *AnalysisArgumentListContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisSearch(ctx *AnalysisSearchContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisSearchAnd(ctx *AnalysisSearchAndContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisSearchUnary(ctx *AnalysisSearchUnaryContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisSearchTerm(ctx *AnalysisSearchTermContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisSearchValue(ctx *AnalysisSearchValueContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisUnquotedValue(ctx *AnalysisUnquotedValueContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisUnquotedPart(ctx *AnalysisUnquotedPartContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisExpression(ctx *AnalysisExpressionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisOr(ctx *AnalysisOrContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisAnd(ctx *AnalysisAndContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisNot(ctx *AnalysisNotContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisComparison(ctx *AnalysisComparisonContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisComparisonOperator(ctx *AnalysisComparisonOperatorContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisConcat(ctx *AnalysisConcatContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisAdd(ctx *AnalysisAddContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisMultiply(ctx *AnalysisMultiplyContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisPower(ctx *AnalysisPowerContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisUnary(ctx *AnalysisUnaryContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisAtom(ctx *AnalysisAtomContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisFunctionCall(ctx *AnalysisFunctionCallContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisFunctionName(ctx *AnalysisFunctionNameContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisLiteral(ctx *AnalysisLiteralContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisSelector(ctx *AnalysisSelectorContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPLParserVisitor) VisitAnalysisIdentifier(ctx *AnalysisIdentifierContext) interface{} {
	return v.VisitChildren(ctx)
}
