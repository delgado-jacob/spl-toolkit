// Code generated from spl-toolkit/grammar/SPLParser.g4 by ANTLR 4.13.2. DO NOT EDIT.

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
}
