// Code generated from grammar/SPL2Parser.g4 by ANTLR 4.13.2. DO NOT EDIT.

package spl2 // SPL2Parser
import "github.com/antlr4-go/antlr/v4"

type BaseSPL2ParserVisitor struct {
	*antlr.BaseParseTreeVisitor
}

func (v *BaseSPL2ParserVisitor) VisitQuery(ctx *QueryContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitPipeline(ctx *PipelineContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitStart(ctx *StartContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitCommand(ctx *CommandContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitFromCommand(ctx *FromCommandContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitSelectCommand(ctx *SelectCommandContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitProjection(ctx *ProjectionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitDataset(ctx *DatasetContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitGenerator(ctx *GeneratorContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitEvalCommand(ctx *EvalCommandContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitAssignment(ctx *AssignmentContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitWhereCommand(ctx *WhereCommandContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitFieldsCommand(ctx *FieldsCommandContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitRexCommand(ctx *RexCommandContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitRexOption(ctx *RexOptionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitEmbeddedCommand(ctx *EmbeddedCommandContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitEmbeddedText(ctx *EmbeddedTextContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitModuleSuffix(ctx *ModuleSuffixContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitModuleDeclaration(ctx *ModuleDeclarationContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitSearchCommand(ctx *SearchCommandContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitImplicitSearch(ctx *ImplicitSearchContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitSearchExpression(ctx *SearchExpressionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitSearchXor(ctx *SearchXorContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitSearchAnd(ctx *SearchAndContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitSearchOr(ctx *SearchOrContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitSearchNot(ctx *SearchNotContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitSearchAtom(ctx *SearchAtomContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitSearchValue(ctx *SearchValueContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitExpression(ctx *ExpressionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitXorExpression(ctx *XorExpressionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitOrExpression(ctx *OrExpressionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitAndExpression(ctx *AndExpressionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitNotExpression(ctx *NotExpressionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitPredicate(ctx *PredicateContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitComparison(ctx *ComparisonContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitAdditive(ctx *AdditiveContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitMultiplicative(ctx *MultiplicativeContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitUnary(ctx *UnaryContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitAccess(ctx *AccessContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitAccessPart(ctx *AccessPartContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitPrimary(ctx *PrimaryContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitCall(ctx *CallContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitArguments(ctx *ArgumentsContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitNamedArgument(ctx *NamedArgumentContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitLiteral(ctx *LiteralContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitStringLiteral(ctx *StringLiteralContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitQuotedName(ctx *QuotedNameContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitFieldTemplate(ctx *FieldTemplateContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitFieldName(ctx *FieldNameContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitIdentifier(ctx *IdentifierContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitArray(ctx *ArrayContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitObject(ctx *ObjectContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitObjectEntry(ctx *ObjectEntryContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitObjectKey(ctx *ObjectKeyContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitSearchLiteral(ctx *SearchLiteralContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitLambdaExpression(ctx *LambdaExpressionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitLambdaParameter(ctx *LambdaParameterContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSPL2ParserVisitor) VisitLambdaBlock(ctx *LambdaBlockContext) interface{} {
	return v.VisitChildren(ctx)
}
