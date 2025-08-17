// Code generated from spl-toolkit/grammar/SPLParser.g4 by ANTLR 4.13.2. DO NOT EDIT.

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
