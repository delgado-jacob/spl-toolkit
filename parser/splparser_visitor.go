// Code generated from spl-toolkit/grammar/SPLParser.g4 by ANTLR 4.13.2. DO NOT EDIT.

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
}
