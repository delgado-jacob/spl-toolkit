// Code generated from grammar/SPLParser.g4 by ANTLR 4.13.2. DO NOT EDIT.

package parser // SPLParser
import "github.com/antlr4-go/antlr/v4"

// SPLParserListener is a complete listener for a parse tree produced by SPLParser.
type SPLParserListener interface {
	antlr.ParseTreeListener

	// EnterQuery is called when entering the query production.
	EnterQuery(c *QueryContext)

	// EnterInitCommand is called when entering the initCommand production.
	EnterInitCommand(c *InitCommandContext)

	// EnterNextCommand is called when entering the nextCommand production.
	EnterNextCommand(c *NextCommandContext)

	// EnterSubquery is called when entering the subquery production.
	EnterSubquery(c *SubqueryContext)

	// EnterLIKEOP is called when entering the LIKEOP production.
	EnterLIKEOP(c *LIKEOPContext)

	// EnterANDOP is called when entering the ANDOP production.
	EnterANDOP(c *ANDOPContext)

	// EnterOROP is called when entering the OROP production.
	EnterOROP(c *OROPContext)

	// EnterINOP is called when entering the INOP production.
	EnterINOP(c *INOPContext)

	// EnterNOTOP is called when entering the NOTOP production.
	EnterNOTOP(c *NOTOPContext)

	// EnterBYOP is called when entering the BYOP production.
	EnterBYOP(c *BYOPContext)

	// EnterPARENOP is called when entering the PARENOP production.
	EnterPARENOP(c *PARENOPContext)

	// EnterOUTPUTMULTIINOP is called when entering the OUTPUTMULTIINOP production.
	EnterOUTPUTMULTIINOP(c *OUTPUTMULTIINOPContext)

	// EnterKEYVALUEOP is called when entering the KEYVALUEOP production.
	EnterKEYVALUEOP(c *KEYVALUEOPContext)

	// EnterOUTPUTOP is called when entering the OUTPUTOP production.
	EnterOUTPUTOP(c *OUTPUTOPContext)

	// EnterEXPRESSIONOP is called when entering the EXPRESSIONOP production.
	EnterEXPRESSIONOP(c *EXPRESSIONOPContext)

	// EnterOUTPUTMULTIOP is called when entering the OUTPUTMULTIOP production.
	EnterOUTPUTMULTIOP(c *OUTPUTMULTIOPContext)

	// EnterRENAMEOP is called when entering the RENAMEOP production.
	EnterRENAMEOP(c *RENAMEOPContext)

	// EnterExpression is called when entering the expression production.
	EnterExpression(c *ExpressionContext)

	// EnterValue is called when entering the value production.
	EnterValue(c *ValueContext)

	// EnterDate is called when entering the date production.
	EnterDate(c *DateContext)

	// EnterFieldUse is called when entering the FieldUse production.
	EnterFieldUse(c *FieldUseContext)

	// EnterCommandUse is called when entering the CommandUse production.
	EnterCommandUse(c *CommandUseContext)

	// EnterFunctionUse is called when entering the FunctionUse production.
	EnterFunctionUse(c *FunctionUseContext)

	// EnterFunction is called when entering the function production.
	EnterFunction(c *FunctionContext)

	// EnterCommand is called when entering the command production.
	EnterCommand(c *CommandContext)

	// EnterAnalysisQuery is called when entering the analysisQuery production.
	EnterAnalysisQuery(c *AnalysisQueryContext)

	// EnterAnalysisPipeline is called when entering the analysisPipeline production.
	EnterAnalysisPipeline(c *AnalysisPipelineContext)

	// EnterAnalysisInitialStage is called when entering the analysisInitialStage production.
	EnterAnalysisInitialStage(c *AnalysisInitialStageContext)

	// EnterAnalysisImplicitSearch is called when entering the analysisImplicitSearch production.
	EnterAnalysisImplicitSearch(c *AnalysisImplicitSearchContext)

	// EnterAnalysisSearchStage is called when entering the AnalysisSearchStage production.
	EnterAnalysisSearchStage(c *AnalysisSearchStageContext)

	// EnterAnalysisWhereStage is called when entering the AnalysisWhereStage production.
	EnterAnalysisWhereStage(c *AnalysisWhereStageContext)

	// EnterAnalysisEvalStage is called when entering the AnalysisEvalStage production.
	EnterAnalysisEvalStage(c *AnalysisEvalStageContext)

	// EnterAnalysisRenameStage is called when entering the AnalysisRenameStage production.
	EnterAnalysisRenameStage(c *AnalysisRenameStageContext)

	// EnterAnalysisFieldsStage is called when entering the AnalysisFieldsStage production.
	EnterAnalysisFieldsStage(c *AnalysisFieldsStageContext)

	// EnterAnalysisStatsStage is called when entering the AnalysisStatsStage production.
	EnterAnalysisStatsStage(c *AnalysisStatsStageContext)

	// EnterAnalysisLookupStage is called when entering the AnalysisLookupStage production.
	EnterAnalysisLookupStage(c *AnalysisLookupStageContext)

	// EnterAnalysisSortStage is called when entering the AnalysisSortStage production.
	EnterAnalysisSortStage(c *AnalysisSortStageContext)

	// EnterAnalysisDedupStage is called when entering the AnalysisDedupStage production.
	EnterAnalysisDedupStage(c *AnalysisDedupStageContext)

	// EnterAnalysisLimitStage is called when entering the AnalysisLimitStage production.
	EnterAnalysisLimitStage(c *AnalysisLimitStageContext)

	// EnterAnalysisInputlookupStage is called when entering the AnalysisInputlookupStage production.
	EnterAnalysisInputlookupStage(c *AnalysisInputlookupStageContext)

	// EnterAnalysisDatamodelStage is called when entering the AnalysisDatamodelStage production.
	EnterAnalysisDatamodelStage(c *AnalysisDatamodelStageContext)

	// EnterAnalysisFromStage is called when entering the AnalysisFromStage production.
	EnterAnalysisFromStage(c *AnalysisFromStageContext)

	// EnterAnalysisTstatsStage is called when entering the AnalysisTstatsStage production.
	EnterAnalysisTstatsStage(c *AnalysisTstatsStageContext)

	// EnterAnalysisFillnullStage is called when entering the AnalysisFillnullStage production.
	EnterAnalysisFillnullStage(c *AnalysisFillnullStageContext)

	// EnterAnalysisRexStage is called when entering the AnalysisRexStage production.
	EnterAnalysisRexStage(c *AnalysisRexStageContext)

	// EnterAnalysisSpathStage is called when entering the AnalysisSpathStage production.
	EnterAnalysisSpathStage(c *AnalysisSpathStageContext)

	// EnterAnalysisBinStage is called when entering the AnalysisBinStage production.
	EnterAnalysisBinStage(c *AnalysisBinStageContext)

	// EnterAnalysisRegexStage is called when entering the AnalysisRegexStage production.
	EnterAnalysisRegexStage(c *AnalysisRegexStageContext)

	// EnterAnalysisMvexpandStage is called when entering the AnalysisMvexpandStage production.
	EnterAnalysisMvexpandStage(c *AnalysisMvexpandStageContext)

	// EnterAnalysisJoinStage is called when entering the AnalysisJoinStage production.
	EnterAnalysisJoinStage(c *AnalysisJoinStageContext)

	// EnterAnalysisBranchStage is called when entering the AnalysisBranchStage production.
	EnterAnalysisBranchStage(c *AnalysisBranchStageContext)

	// EnterAnalysisMacroStage is called when entering the AnalysisMacroStage production.
	EnterAnalysisMacroStage(c *AnalysisMacroStageContext)

	// EnterAnalysisOpaqueStage is called when entering the AnalysisOpaqueStage production.
	EnterAnalysisOpaqueStage(c *AnalysisOpaqueStageContext)

	// EnterAnalysisTstats is called when entering the analysisTstats production.
	EnterAnalysisTstats(c *AnalysisTstatsContext)

	// EnterAnalysisTstatsOption is called when entering the analysisTstatsOption production.
	EnterAnalysisTstatsOption(c *AnalysisTstatsOptionContext)

	// EnterAnalysisTstatsGroup is called when entering the analysisTstatsGroup production.
	EnterAnalysisTstatsGroup(c *AnalysisTstatsGroupContext)

	// EnterAnalysisTstatsGroupItem is called when entering the analysisTstatsGroupItem production.
	EnterAnalysisTstatsGroupItem(c *AnalysisTstatsGroupItemContext)

	// EnterAnalysisTstatsSpanOption is called when entering the analysisTstatsSpanOption production.
	EnterAnalysisTstatsSpanOption(c *AnalysisTstatsSpanOptionContext)

	// EnterAnalysisFillnull is called when entering the analysisFillnull production.
	EnterAnalysisFillnull(c *AnalysisFillnullContext)

	// EnterAnalysisFillnullValueOption is called when entering the analysisFillnullValueOption production.
	EnterAnalysisFillnullValueOption(c *AnalysisFillnullValueOptionContext)

	// EnterAnalysisRex is called when entering the analysisRex production.
	EnterAnalysisRex(c *AnalysisRexContext)

	// EnterAnalysisRexFieldOption is called when entering the analysisRexFieldOption production.
	EnterAnalysisRexFieldOption(c *AnalysisRexFieldOptionContext)

	// EnterAnalysisRexMaxMatchOption is called when entering the analysisRexMaxMatchOption production.
	EnterAnalysisRexMaxMatchOption(c *AnalysisRexMaxMatchOptionContext)

	// EnterAnalysisRexOffsetFieldOption is called when entering the analysisRexOffsetFieldOption production.
	EnterAnalysisRexOffsetFieldOption(c *AnalysisRexOffsetFieldOptionContext)

	// EnterAnalysisRexModeOption is called when entering the analysisRexModeOption production.
	EnterAnalysisRexModeOption(c *AnalysisRexModeOptionContext)

	// EnterAnalysisSpath is called when entering the analysisSpath production.
	EnterAnalysisSpath(c *AnalysisSpathContext)

	// EnterAnalysisSpathInputOption is called when entering the analysisSpathInputOption production.
	EnterAnalysisSpathInputOption(c *AnalysisSpathInputOptionContext)

	// EnterAnalysisSpathPathOption is called when entering the analysisSpathPathOption production.
	EnterAnalysisSpathPathOption(c *AnalysisSpathPathOptionContext)

	// EnterAnalysisSpathOutputOption is called when entering the analysisSpathOutputOption production.
	EnterAnalysisSpathOutputOption(c *AnalysisSpathOutputOptionContext)

	// EnterAnalysisBin is called when entering the analysisBin production.
	EnterAnalysisBin(c *AnalysisBinContext)

	// EnterAnalysisBinOption is called when entering the analysisBinOption production.
	EnterAnalysisBinOption(c *AnalysisBinOptionContext)

	// EnterAnalysisRegex is called when entering the analysisRegex production.
	EnterAnalysisRegex(c *AnalysisRegexContext)

	// EnterAnalysisMvexpand is called when entering the analysisMvexpand production.
	EnterAnalysisMvexpand(c *AnalysisMvexpandContext)

	// EnterAnalysisMvexpandOption is called when entering the analysisMvexpandOption production.
	EnterAnalysisMvexpandOption(c *AnalysisMvexpandOptionContext)

	// EnterAnalysisJoin is called when entering the analysisJoin production.
	EnterAnalysisJoin(c *AnalysisJoinContext)

	// EnterAnalysisJoinOption is called when entering the analysisJoinOption production.
	EnterAnalysisJoinOption(c *AnalysisJoinOptionContext)

	// EnterAnalysisBranch is called when entering the analysisBranch production.
	EnterAnalysisBranch(c *AnalysisBranchContext)

	// EnterAnalysisBranchOption is called when entering the analysisBranchOption production.
	EnterAnalysisBranchOption(c *AnalysisBranchOptionContext)

	// EnterAnalysisOptionValue is called when entering the analysisOptionValue production.
	EnterAnalysisOptionValue(c *AnalysisOptionValueContext)

	// EnterAnalysisUnitOptionValue is called when entering the analysisUnitOptionValue production.
	EnterAnalysisUnitOptionValue(c *AnalysisUnitOptionValueContext)

	// EnterAnalysisUnitSuffix is called when entering the analysisUnitSuffix production.
	EnterAnalysisUnitSuffix(c *AnalysisUnitSuffixContext)

	// EnterAnalysisInvalidOptionValue is called when entering the analysisInvalidOptionValue production.
	EnterAnalysisInvalidOptionValue(c *AnalysisInvalidOptionValueContext)

	// EnterAnalysisMissingOptionValue is called when entering the analysisMissingOptionValue production.
	EnterAnalysisMissingOptionValue(c *AnalysisMissingOptionValueContext)

	// EnterAnalysisMissingJoinKey is called when entering the analysisMissingJoinKey production.
	EnterAnalysisMissingJoinKey(c *AnalysisMissingJoinKeyContext)

	// EnterAnalysisMissingOperand is called when entering the analysisMissingOperand production.
	EnterAnalysisMissingOperand(c *AnalysisMissingOperandContext)

	// EnterAnalysisMissingEvalAssignment is called when entering the analysisMissingEvalAssignment production.
	EnterAnalysisMissingEvalAssignment(c *AnalysisMissingEvalAssignmentContext)

	// EnterAnalysisDataModelName is called when entering the analysisDataModelName production.
	EnterAnalysisDataModelName(c *AnalysisDataModelNameContext)

	// EnterAnalysisDataModelDataset is called when entering the analysisDataModelDataset production.
	EnterAnalysisDataModelDataset(c *AnalysisDataModelDatasetContext)

	// EnterAnalysisFromDataset is called when entering the analysisFromDataset production.
	EnterAnalysisFromDataset(c *AnalysisFromDatasetContext)

	// EnterAnalysisTstatsFrom is called when entering the analysisTstatsFrom production.
	EnterAnalysisTstatsFrom(c *AnalysisTstatsFromContext)

	// EnterAnalysisTstatsWhere is called when entering the analysisTstatsWhere production.
	EnterAnalysisTstatsWhere(c *AnalysisTstatsWhereContext)

	// EnterAnalysisCommandName is called when entering the analysisCommandName production.
	EnterAnalysisCommandName(c *AnalysisCommandNameContext)

	// EnterAnalysisSortField is called when entering the analysisSortField production.
	EnterAnalysisSortField(c *AnalysisSortFieldContext)

	// EnterAnalysisLimit is called when entering the analysisLimit production.
	EnterAnalysisLimit(c *AnalysisLimitContext)

	// EnterAnalysisOption is called when entering the analysisOption production.
	EnterAnalysisOption(c *AnalysisOptionContext)

	// EnterAnalysisCatalogName is called when entering the analysisCatalogName production.
	EnterAnalysisCatalogName(c *AnalysisCatalogNameContext)

	// EnterAnalysisAssignment is called when entering the analysisAssignment production.
	EnterAnalysisAssignment(c *AnalysisAssignmentContext)

	// EnterAnalysisRename is called when entering the analysisRename production.
	EnterAnalysisRename(c *AnalysisRenameContext)

	// EnterAnalysisAlias is called when entering the analysisAlias production.
	EnterAnalysisAlias(c *AnalysisAliasContext)

	// EnterAnalysisFieldList is called when entering the analysisFieldList production.
	EnterAnalysisFieldList(c *AnalysisFieldListContext)

	// EnterAnalysisGroup is called when entering the analysisGroup production.
	EnterAnalysisGroup(c *AnalysisGroupContext)

	// EnterAnalysisAggregate is called when entering the analysisAggregate production.
	EnterAnalysisAggregate(c *AnalysisAggregateContext)

	// EnterAnalysisLookup is called when entering the analysisLookup production.
	EnterAnalysisLookup(c *AnalysisLookupContext)

	// EnterAnalysisLookupInput is called when entering the analysisLookupInput production.
	EnterAnalysisLookupInput(c *AnalysisLookupInputContext)

	// EnterAnalysisOutput is called when entering the analysisOutput production.
	EnterAnalysisOutput(c *AnalysisOutputContext)

	// EnterAnalysisArgument is called when entering the analysisArgument production.
	EnterAnalysisArgument(c *AnalysisArgumentContext)

	// EnterAnalysisSubquery is called when entering the analysisSubquery production.
	EnterAnalysisSubquery(c *AnalysisSubqueryContext)

	// EnterAnalysisMacro is called when entering the analysisMacro production.
	EnterAnalysisMacro(c *AnalysisMacroContext)

	// EnterAnalysisArgumentList is called when entering the analysisArgumentList production.
	EnterAnalysisArgumentList(c *AnalysisArgumentListContext)

	// EnterAnalysisSearch is called when entering the analysisSearch production.
	EnterAnalysisSearch(c *AnalysisSearchContext)

	// EnterAnalysisSearchAnd is called when entering the analysisSearchAnd production.
	EnterAnalysisSearchAnd(c *AnalysisSearchAndContext)

	// EnterAnalysisSearchUnary is called when entering the analysisSearchUnary production.
	EnterAnalysisSearchUnary(c *AnalysisSearchUnaryContext)

	// EnterAnalysisSearchTerm is called when entering the analysisSearchTerm production.
	EnterAnalysisSearchTerm(c *AnalysisSearchTermContext)

	// EnterAnalysisSearchValue is called when entering the analysisSearchValue production.
	EnterAnalysisSearchValue(c *AnalysisSearchValueContext)

	// EnterAnalysisUnquotedValue is called when entering the analysisUnquotedValue production.
	EnterAnalysisUnquotedValue(c *AnalysisUnquotedValueContext)

	// EnterAnalysisUnquotedPart is called when entering the analysisUnquotedPart production.
	EnterAnalysisUnquotedPart(c *AnalysisUnquotedPartContext)

	// EnterAnalysisExpression is called when entering the analysisExpression production.
	EnterAnalysisExpression(c *AnalysisExpressionContext)

	// EnterAnalysisOr is called when entering the analysisOr production.
	EnterAnalysisOr(c *AnalysisOrContext)

	// EnterAnalysisAnd is called when entering the analysisAnd production.
	EnterAnalysisAnd(c *AnalysisAndContext)

	// EnterAnalysisNot is called when entering the analysisNot production.
	EnterAnalysisNot(c *AnalysisNotContext)

	// EnterAnalysisComparison is called when entering the analysisComparison production.
	EnterAnalysisComparison(c *AnalysisComparisonContext)

	// EnterAnalysisComparisonOperator is called when entering the analysisComparisonOperator production.
	EnterAnalysisComparisonOperator(c *AnalysisComparisonOperatorContext)

	// EnterAnalysisConcat is called when entering the analysisConcat production.
	EnterAnalysisConcat(c *AnalysisConcatContext)

	// EnterAnalysisAdd is called when entering the analysisAdd production.
	EnterAnalysisAdd(c *AnalysisAddContext)

	// EnterAnalysisMultiply is called when entering the analysisMultiply production.
	EnterAnalysisMultiply(c *AnalysisMultiplyContext)

	// EnterAnalysisPower is called when entering the analysisPower production.
	EnterAnalysisPower(c *AnalysisPowerContext)

	// EnterAnalysisUnary is called when entering the analysisUnary production.
	EnterAnalysisUnary(c *AnalysisUnaryContext)

	// EnterAnalysisAtom is called when entering the analysisAtom production.
	EnterAnalysisAtom(c *AnalysisAtomContext)

	// EnterAnalysisFunctionCall is called when entering the analysisFunctionCall production.
	EnterAnalysisFunctionCall(c *AnalysisFunctionCallContext)

	// EnterAnalysisFunctionName is called when entering the analysisFunctionName production.
	EnterAnalysisFunctionName(c *AnalysisFunctionNameContext)

	// EnterAnalysisLiteral is called when entering the analysisLiteral production.
	EnterAnalysisLiteral(c *AnalysisLiteralContext)

	// EnterAnalysisSelector is called when entering the analysisSelector production.
	EnterAnalysisSelector(c *AnalysisSelectorContext)

	// EnterAnalysisIdentifier is called when entering the analysisIdentifier production.
	EnterAnalysisIdentifier(c *AnalysisIdentifierContext)

	// ExitQuery is called when exiting the query production.
	ExitQuery(c *QueryContext)

	// ExitInitCommand is called when exiting the initCommand production.
	ExitInitCommand(c *InitCommandContext)

	// ExitNextCommand is called when exiting the nextCommand production.
	ExitNextCommand(c *NextCommandContext)

	// ExitSubquery is called when exiting the subquery production.
	ExitSubquery(c *SubqueryContext)

	// ExitLIKEOP is called when exiting the LIKEOP production.
	ExitLIKEOP(c *LIKEOPContext)

	// ExitANDOP is called when exiting the ANDOP production.
	ExitANDOP(c *ANDOPContext)

	// ExitOROP is called when exiting the OROP production.
	ExitOROP(c *OROPContext)

	// ExitINOP is called when exiting the INOP production.
	ExitINOP(c *INOPContext)

	// ExitNOTOP is called when exiting the NOTOP production.
	ExitNOTOP(c *NOTOPContext)

	// ExitBYOP is called when exiting the BYOP production.
	ExitBYOP(c *BYOPContext)

	// ExitPARENOP is called when exiting the PARENOP production.
	ExitPARENOP(c *PARENOPContext)

	// ExitOUTPUTMULTIINOP is called when exiting the OUTPUTMULTIINOP production.
	ExitOUTPUTMULTIINOP(c *OUTPUTMULTIINOPContext)

	// ExitKEYVALUEOP is called when exiting the KEYVALUEOP production.
	ExitKEYVALUEOP(c *KEYVALUEOPContext)

	// ExitOUTPUTOP is called when exiting the OUTPUTOP production.
	ExitOUTPUTOP(c *OUTPUTOPContext)

	// ExitEXPRESSIONOP is called when exiting the EXPRESSIONOP production.
	ExitEXPRESSIONOP(c *EXPRESSIONOPContext)

	// ExitOUTPUTMULTIOP is called when exiting the OUTPUTMULTIOP production.
	ExitOUTPUTMULTIOP(c *OUTPUTMULTIOPContext)

	// ExitRENAMEOP is called when exiting the RENAMEOP production.
	ExitRENAMEOP(c *RENAMEOPContext)

	// ExitExpression is called when exiting the expression production.
	ExitExpression(c *ExpressionContext)

	// ExitValue is called when exiting the value production.
	ExitValue(c *ValueContext)

	// ExitDate is called when exiting the date production.
	ExitDate(c *DateContext)

	// ExitFieldUse is called when exiting the FieldUse production.
	ExitFieldUse(c *FieldUseContext)

	// ExitCommandUse is called when exiting the CommandUse production.
	ExitCommandUse(c *CommandUseContext)

	// ExitFunctionUse is called when exiting the FunctionUse production.
	ExitFunctionUse(c *FunctionUseContext)

	// ExitFunction is called when exiting the function production.
	ExitFunction(c *FunctionContext)

	// ExitCommand is called when exiting the command production.
	ExitCommand(c *CommandContext)

	// ExitAnalysisQuery is called when exiting the analysisQuery production.
	ExitAnalysisQuery(c *AnalysisQueryContext)

	// ExitAnalysisPipeline is called when exiting the analysisPipeline production.
	ExitAnalysisPipeline(c *AnalysisPipelineContext)

	// ExitAnalysisInitialStage is called when exiting the analysisInitialStage production.
	ExitAnalysisInitialStage(c *AnalysisInitialStageContext)

	// ExitAnalysisImplicitSearch is called when exiting the analysisImplicitSearch production.
	ExitAnalysisImplicitSearch(c *AnalysisImplicitSearchContext)

	// ExitAnalysisSearchStage is called when exiting the AnalysisSearchStage production.
	ExitAnalysisSearchStage(c *AnalysisSearchStageContext)

	// ExitAnalysisWhereStage is called when exiting the AnalysisWhereStage production.
	ExitAnalysisWhereStage(c *AnalysisWhereStageContext)

	// ExitAnalysisEvalStage is called when exiting the AnalysisEvalStage production.
	ExitAnalysisEvalStage(c *AnalysisEvalStageContext)

	// ExitAnalysisRenameStage is called when exiting the AnalysisRenameStage production.
	ExitAnalysisRenameStage(c *AnalysisRenameStageContext)

	// ExitAnalysisFieldsStage is called when exiting the AnalysisFieldsStage production.
	ExitAnalysisFieldsStage(c *AnalysisFieldsStageContext)

	// ExitAnalysisStatsStage is called when exiting the AnalysisStatsStage production.
	ExitAnalysisStatsStage(c *AnalysisStatsStageContext)

	// ExitAnalysisLookupStage is called when exiting the AnalysisLookupStage production.
	ExitAnalysisLookupStage(c *AnalysisLookupStageContext)

	// ExitAnalysisSortStage is called when exiting the AnalysisSortStage production.
	ExitAnalysisSortStage(c *AnalysisSortStageContext)

	// ExitAnalysisDedupStage is called when exiting the AnalysisDedupStage production.
	ExitAnalysisDedupStage(c *AnalysisDedupStageContext)

	// ExitAnalysisLimitStage is called when exiting the AnalysisLimitStage production.
	ExitAnalysisLimitStage(c *AnalysisLimitStageContext)

	// ExitAnalysisInputlookupStage is called when exiting the AnalysisInputlookupStage production.
	ExitAnalysisInputlookupStage(c *AnalysisInputlookupStageContext)

	// ExitAnalysisDatamodelStage is called when exiting the AnalysisDatamodelStage production.
	ExitAnalysisDatamodelStage(c *AnalysisDatamodelStageContext)

	// ExitAnalysisFromStage is called when exiting the AnalysisFromStage production.
	ExitAnalysisFromStage(c *AnalysisFromStageContext)

	// ExitAnalysisTstatsStage is called when exiting the AnalysisTstatsStage production.
	ExitAnalysisTstatsStage(c *AnalysisTstatsStageContext)

	// ExitAnalysisFillnullStage is called when exiting the AnalysisFillnullStage production.
	ExitAnalysisFillnullStage(c *AnalysisFillnullStageContext)

	// ExitAnalysisRexStage is called when exiting the AnalysisRexStage production.
	ExitAnalysisRexStage(c *AnalysisRexStageContext)

	// ExitAnalysisSpathStage is called when exiting the AnalysisSpathStage production.
	ExitAnalysisSpathStage(c *AnalysisSpathStageContext)

	// ExitAnalysisBinStage is called when exiting the AnalysisBinStage production.
	ExitAnalysisBinStage(c *AnalysisBinStageContext)

	// ExitAnalysisRegexStage is called when exiting the AnalysisRegexStage production.
	ExitAnalysisRegexStage(c *AnalysisRegexStageContext)

	// ExitAnalysisMvexpandStage is called when exiting the AnalysisMvexpandStage production.
	ExitAnalysisMvexpandStage(c *AnalysisMvexpandStageContext)

	// ExitAnalysisJoinStage is called when exiting the AnalysisJoinStage production.
	ExitAnalysisJoinStage(c *AnalysisJoinStageContext)

	// ExitAnalysisBranchStage is called when exiting the AnalysisBranchStage production.
	ExitAnalysisBranchStage(c *AnalysisBranchStageContext)

	// ExitAnalysisMacroStage is called when exiting the AnalysisMacroStage production.
	ExitAnalysisMacroStage(c *AnalysisMacroStageContext)

	// ExitAnalysisOpaqueStage is called when exiting the AnalysisOpaqueStage production.
	ExitAnalysisOpaqueStage(c *AnalysisOpaqueStageContext)

	// ExitAnalysisTstats is called when exiting the analysisTstats production.
	ExitAnalysisTstats(c *AnalysisTstatsContext)

	// ExitAnalysisTstatsOption is called when exiting the analysisTstatsOption production.
	ExitAnalysisTstatsOption(c *AnalysisTstatsOptionContext)

	// ExitAnalysisTstatsGroup is called when exiting the analysisTstatsGroup production.
	ExitAnalysisTstatsGroup(c *AnalysisTstatsGroupContext)

	// ExitAnalysisTstatsGroupItem is called when exiting the analysisTstatsGroupItem production.
	ExitAnalysisTstatsGroupItem(c *AnalysisTstatsGroupItemContext)

	// ExitAnalysisTstatsSpanOption is called when exiting the analysisTstatsSpanOption production.
	ExitAnalysisTstatsSpanOption(c *AnalysisTstatsSpanOptionContext)

	// ExitAnalysisFillnull is called when exiting the analysisFillnull production.
	ExitAnalysisFillnull(c *AnalysisFillnullContext)

	// ExitAnalysisFillnullValueOption is called when exiting the analysisFillnullValueOption production.
	ExitAnalysisFillnullValueOption(c *AnalysisFillnullValueOptionContext)

	// ExitAnalysisRex is called when exiting the analysisRex production.
	ExitAnalysisRex(c *AnalysisRexContext)

	// ExitAnalysisRexFieldOption is called when exiting the analysisRexFieldOption production.
	ExitAnalysisRexFieldOption(c *AnalysisRexFieldOptionContext)

	// ExitAnalysisRexMaxMatchOption is called when exiting the analysisRexMaxMatchOption production.
	ExitAnalysisRexMaxMatchOption(c *AnalysisRexMaxMatchOptionContext)

	// ExitAnalysisRexOffsetFieldOption is called when exiting the analysisRexOffsetFieldOption production.
	ExitAnalysisRexOffsetFieldOption(c *AnalysisRexOffsetFieldOptionContext)

	// ExitAnalysisRexModeOption is called when exiting the analysisRexModeOption production.
	ExitAnalysisRexModeOption(c *AnalysisRexModeOptionContext)

	// ExitAnalysisSpath is called when exiting the analysisSpath production.
	ExitAnalysisSpath(c *AnalysisSpathContext)

	// ExitAnalysisSpathInputOption is called when exiting the analysisSpathInputOption production.
	ExitAnalysisSpathInputOption(c *AnalysisSpathInputOptionContext)

	// ExitAnalysisSpathPathOption is called when exiting the analysisSpathPathOption production.
	ExitAnalysisSpathPathOption(c *AnalysisSpathPathOptionContext)

	// ExitAnalysisSpathOutputOption is called when exiting the analysisSpathOutputOption production.
	ExitAnalysisSpathOutputOption(c *AnalysisSpathOutputOptionContext)

	// ExitAnalysisBin is called when exiting the analysisBin production.
	ExitAnalysisBin(c *AnalysisBinContext)

	// ExitAnalysisBinOption is called when exiting the analysisBinOption production.
	ExitAnalysisBinOption(c *AnalysisBinOptionContext)

	// ExitAnalysisRegex is called when exiting the analysisRegex production.
	ExitAnalysisRegex(c *AnalysisRegexContext)

	// ExitAnalysisMvexpand is called when exiting the analysisMvexpand production.
	ExitAnalysisMvexpand(c *AnalysisMvexpandContext)

	// ExitAnalysisMvexpandOption is called when exiting the analysisMvexpandOption production.
	ExitAnalysisMvexpandOption(c *AnalysisMvexpandOptionContext)

	// ExitAnalysisJoin is called when exiting the analysisJoin production.
	ExitAnalysisJoin(c *AnalysisJoinContext)

	// ExitAnalysisJoinOption is called when exiting the analysisJoinOption production.
	ExitAnalysisJoinOption(c *AnalysisJoinOptionContext)

	// ExitAnalysisBranch is called when exiting the analysisBranch production.
	ExitAnalysisBranch(c *AnalysisBranchContext)

	// ExitAnalysisBranchOption is called when exiting the analysisBranchOption production.
	ExitAnalysisBranchOption(c *AnalysisBranchOptionContext)

	// ExitAnalysisOptionValue is called when exiting the analysisOptionValue production.
	ExitAnalysisOptionValue(c *AnalysisOptionValueContext)

	// ExitAnalysisUnitOptionValue is called when exiting the analysisUnitOptionValue production.
	ExitAnalysisUnitOptionValue(c *AnalysisUnitOptionValueContext)

	// ExitAnalysisUnitSuffix is called when exiting the analysisUnitSuffix production.
	ExitAnalysisUnitSuffix(c *AnalysisUnitSuffixContext)

	// ExitAnalysisInvalidOptionValue is called when exiting the analysisInvalidOptionValue production.
	ExitAnalysisInvalidOptionValue(c *AnalysisInvalidOptionValueContext)

	// ExitAnalysisMissingOptionValue is called when exiting the analysisMissingOptionValue production.
	ExitAnalysisMissingOptionValue(c *AnalysisMissingOptionValueContext)

	// ExitAnalysisMissingJoinKey is called when exiting the analysisMissingJoinKey production.
	ExitAnalysisMissingJoinKey(c *AnalysisMissingJoinKeyContext)

	// ExitAnalysisMissingOperand is called when exiting the analysisMissingOperand production.
	ExitAnalysisMissingOperand(c *AnalysisMissingOperandContext)

	// ExitAnalysisMissingEvalAssignment is called when exiting the analysisMissingEvalAssignment production.
	ExitAnalysisMissingEvalAssignment(c *AnalysisMissingEvalAssignmentContext)

	// ExitAnalysisDataModelName is called when exiting the analysisDataModelName production.
	ExitAnalysisDataModelName(c *AnalysisDataModelNameContext)

	// ExitAnalysisDataModelDataset is called when exiting the analysisDataModelDataset production.
	ExitAnalysisDataModelDataset(c *AnalysisDataModelDatasetContext)

	// ExitAnalysisFromDataset is called when exiting the analysisFromDataset production.
	ExitAnalysisFromDataset(c *AnalysisFromDatasetContext)

	// ExitAnalysisTstatsFrom is called when exiting the analysisTstatsFrom production.
	ExitAnalysisTstatsFrom(c *AnalysisTstatsFromContext)

	// ExitAnalysisTstatsWhere is called when exiting the analysisTstatsWhere production.
	ExitAnalysisTstatsWhere(c *AnalysisTstatsWhereContext)

	// ExitAnalysisCommandName is called when exiting the analysisCommandName production.
	ExitAnalysisCommandName(c *AnalysisCommandNameContext)

	// ExitAnalysisSortField is called when exiting the analysisSortField production.
	ExitAnalysisSortField(c *AnalysisSortFieldContext)

	// ExitAnalysisLimit is called when exiting the analysisLimit production.
	ExitAnalysisLimit(c *AnalysisLimitContext)

	// ExitAnalysisOption is called when exiting the analysisOption production.
	ExitAnalysisOption(c *AnalysisOptionContext)

	// ExitAnalysisCatalogName is called when exiting the analysisCatalogName production.
	ExitAnalysisCatalogName(c *AnalysisCatalogNameContext)

	// ExitAnalysisAssignment is called when exiting the analysisAssignment production.
	ExitAnalysisAssignment(c *AnalysisAssignmentContext)

	// ExitAnalysisRename is called when exiting the analysisRename production.
	ExitAnalysisRename(c *AnalysisRenameContext)

	// ExitAnalysisAlias is called when exiting the analysisAlias production.
	ExitAnalysisAlias(c *AnalysisAliasContext)

	// ExitAnalysisFieldList is called when exiting the analysisFieldList production.
	ExitAnalysisFieldList(c *AnalysisFieldListContext)

	// ExitAnalysisGroup is called when exiting the analysisGroup production.
	ExitAnalysisGroup(c *AnalysisGroupContext)

	// ExitAnalysisAggregate is called when exiting the analysisAggregate production.
	ExitAnalysisAggregate(c *AnalysisAggregateContext)

	// ExitAnalysisLookup is called when exiting the analysisLookup production.
	ExitAnalysisLookup(c *AnalysisLookupContext)

	// ExitAnalysisLookupInput is called when exiting the analysisLookupInput production.
	ExitAnalysisLookupInput(c *AnalysisLookupInputContext)

	// ExitAnalysisOutput is called when exiting the analysisOutput production.
	ExitAnalysisOutput(c *AnalysisOutputContext)

	// ExitAnalysisArgument is called when exiting the analysisArgument production.
	ExitAnalysisArgument(c *AnalysisArgumentContext)

	// ExitAnalysisSubquery is called when exiting the analysisSubquery production.
	ExitAnalysisSubquery(c *AnalysisSubqueryContext)

	// ExitAnalysisMacro is called when exiting the analysisMacro production.
	ExitAnalysisMacro(c *AnalysisMacroContext)

	// ExitAnalysisArgumentList is called when exiting the analysisArgumentList production.
	ExitAnalysisArgumentList(c *AnalysisArgumentListContext)

	// ExitAnalysisSearch is called when exiting the analysisSearch production.
	ExitAnalysisSearch(c *AnalysisSearchContext)

	// ExitAnalysisSearchAnd is called when exiting the analysisSearchAnd production.
	ExitAnalysisSearchAnd(c *AnalysisSearchAndContext)

	// ExitAnalysisSearchUnary is called when exiting the analysisSearchUnary production.
	ExitAnalysisSearchUnary(c *AnalysisSearchUnaryContext)

	// ExitAnalysisSearchTerm is called when exiting the analysisSearchTerm production.
	ExitAnalysisSearchTerm(c *AnalysisSearchTermContext)

	// ExitAnalysisSearchValue is called when exiting the analysisSearchValue production.
	ExitAnalysisSearchValue(c *AnalysisSearchValueContext)

	// ExitAnalysisUnquotedValue is called when exiting the analysisUnquotedValue production.
	ExitAnalysisUnquotedValue(c *AnalysisUnquotedValueContext)

	// ExitAnalysisUnquotedPart is called when exiting the analysisUnquotedPart production.
	ExitAnalysisUnquotedPart(c *AnalysisUnquotedPartContext)

	// ExitAnalysisExpression is called when exiting the analysisExpression production.
	ExitAnalysisExpression(c *AnalysisExpressionContext)

	// ExitAnalysisOr is called when exiting the analysisOr production.
	ExitAnalysisOr(c *AnalysisOrContext)

	// ExitAnalysisAnd is called when exiting the analysisAnd production.
	ExitAnalysisAnd(c *AnalysisAndContext)

	// ExitAnalysisNot is called when exiting the analysisNot production.
	ExitAnalysisNot(c *AnalysisNotContext)

	// ExitAnalysisComparison is called when exiting the analysisComparison production.
	ExitAnalysisComparison(c *AnalysisComparisonContext)

	// ExitAnalysisComparisonOperator is called when exiting the analysisComparisonOperator production.
	ExitAnalysisComparisonOperator(c *AnalysisComparisonOperatorContext)

	// ExitAnalysisConcat is called when exiting the analysisConcat production.
	ExitAnalysisConcat(c *AnalysisConcatContext)

	// ExitAnalysisAdd is called when exiting the analysisAdd production.
	ExitAnalysisAdd(c *AnalysisAddContext)

	// ExitAnalysisMultiply is called when exiting the analysisMultiply production.
	ExitAnalysisMultiply(c *AnalysisMultiplyContext)

	// ExitAnalysisPower is called when exiting the analysisPower production.
	ExitAnalysisPower(c *AnalysisPowerContext)

	// ExitAnalysisUnary is called when exiting the analysisUnary production.
	ExitAnalysisUnary(c *AnalysisUnaryContext)

	// ExitAnalysisAtom is called when exiting the analysisAtom production.
	ExitAnalysisAtom(c *AnalysisAtomContext)

	// ExitAnalysisFunctionCall is called when exiting the analysisFunctionCall production.
	ExitAnalysisFunctionCall(c *AnalysisFunctionCallContext)

	// ExitAnalysisFunctionName is called when exiting the analysisFunctionName production.
	ExitAnalysisFunctionName(c *AnalysisFunctionNameContext)

	// ExitAnalysisLiteral is called when exiting the analysisLiteral production.
	ExitAnalysisLiteral(c *AnalysisLiteralContext)

	// ExitAnalysisSelector is called when exiting the analysisSelector production.
	ExitAnalysisSelector(c *AnalysisSelectorContext)

	// ExitAnalysisIdentifier is called when exiting the analysisIdentifier production.
	ExitAnalysisIdentifier(c *AnalysisIdentifierContext)
}
