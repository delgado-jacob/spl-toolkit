// Code generated from grammar/SPL2Parser.g4 by ANTLR 4.13.2. DO NOT EDIT.

package spl2 // SPL2Parser
import "github.com/antlr4-go/antlr/v4"

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

	// Visit a parse tree produced by SPL2Parser#projection.
	VisitProjection(ctx *ProjectionContext) interface{}

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

	// Visit a parse tree produced by SPL2Parser#rexCommand.
	VisitRexCommand(ctx *RexCommandContext) interface{}

	// Visit a parse tree produced by SPL2Parser#rexOption.
	VisitRexOption(ctx *RexOptionContext) interface{}

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
