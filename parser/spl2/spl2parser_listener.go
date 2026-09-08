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

	// EnterProjection is called when entering the projection production.
	EnterProjection(c *ProjectionContext)

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

	// EnterRexCommand is called when entering the rexCommand production.
	EnterRexCommand(c *RexCommandContext)

	// EnterRexOption is called when entering the rexOption production.
	EnterRexOption(c *RexOptionContext)

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

	// ExitProjection is called when exiting the projection production.
	ExitProjection(c *ProjectionContext)

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

	// ExitRexCommand is called when exiting the rexCommand production.
	ExitRexCommand(c *RexCommandContext)

	// ExitRexOption is called when exiting the rexOption production.
	ExitRexOption(c *RexOptionContext)

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
