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
