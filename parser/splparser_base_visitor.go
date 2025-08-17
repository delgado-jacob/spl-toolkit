// Code generated from spl-toolkit/grammar/SPLParser.g4 by ANTLR 4.13.2. DO NOT EDIT.

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
