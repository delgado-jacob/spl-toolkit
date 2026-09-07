/*
 [The "BSD licence"]
 Copyright (c) 2024 Clemens Sageder
 All rights reserved.

 Redistribution and use in source and binary forms, with or without
 modification, are permitted provided that the following conditions
 are met:
 1. Redistributions of source code must retain the above copyright
    notice, this list of conditions and the following disclaimer.
 2. Redistributions in binary form must reproduce the above copyright
    notice, this list of conditions and the following disclaimer in the
    documentation and/or other materials provided with the distribution.
 3. The name of the author may not be used to endorse or promote products
    derived from this software without specific prior written permission.

 THIS SOFTWARE IS PROVIDED BY THE AUTHOR ``AS IS'' AND ANY EXPRESS OR
 IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES
 OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED.
 IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY DIRECT, INDIRECT,
 INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT
 NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
 DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
 THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
 (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF
 THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
*/

parser grammar SPLParser;

options { tokenVocab = SPLLexer; }

@parser::header {import "strings"}

@parser::members {
func (p *SPLParser) analysisCommandIs(names ...string) bool {
    name := p.GetTokenStream().LT(1).GetText()
    for _, candidate := range names { if strings.EqualFold(name, candidate) { return true } }
    return false
}
func (p *SPLParser) analysisIsCommand() bool {
    kind := p.GetTokenStream().LA(1)
    return kind == SPLParserINIT_COMMAND || kind == SPLParserSTD_COMMAND || kind == SPLParserSTD_COMMAND_AND_FUNCTION
}
}

query
    : initCommand (PIPE nextCommand)* EOF
    ;

initCommand
    : PIPE? INIT_COMMAND? operation+ subquery?
    ;

nextCommand
    : command operation+ subquery?
    ;

subquery
    : LBRACK initCommand (PIPE nextCommand)* RBRACK
    ;

operation
    : operation AND operation   #ANDOP
    | operation OR operation    #OROP
    | expression LIKE value     #LIKEOP
    | expression IN LPAREN (expression (COMMA expression)*)? RPAREN     #INOP
    | NOT operation             #NOTOP
    | expression (OUTPUT | OUTPUTNEW) id    #OUTPUTOP
    | expression expression OUTPUT id COMMA id  #OUTPUTMULTIOP
    | expression expression expression (OUTPUT | OUTPUTNEW) id COMMA id COMMA id    # OUTPUTMULTIINOP
    | BY (id)+  #BYOP
    | expression AS id  #RENAMEOP
    | id (EQ | NE | GT | LT | GE | LE) expression   #KEYVALUEOP
    | expression    #EXPRESSIONOP
    | LPAREN operation RPAREN   #PARENOP
    ;

expression
    : function LPAREN (expression (COMMA expression)*)? RPAREN  // function call
    | LPAREN expression RPAREN                                  // paren
    | <assoc = right> expression POW expression                 // power
    | expression (MULT | MOD) expression                        // mult, div, mod
    | expression DIV expression                                 // div, path
    | MULT expression MULT                                      // wildcard
    | MULT expression                                           // wildcard
    | expression MULT                                           // wildcard
    | MULT                                                      // wildcard
    | (DIV id)+                                                 // path
    | expression (ADD | SUB) expression                         // add, sub
    | value                                                     // value
    ;

value
    : date
    | STRING
    | id
    | (ADD | SUB)? NUMBER
    ;

date
    : QUOTE? TIME_AND_FUNCTION QUOTE?
    | TIME
    ;

id
    : (IDENTIFIER | DOT | QUOTED_IDENTIFIER)    #FieldUse
    | command       #CommandUse
    | function      #FunctionUse
    ;

function
    : FUNCTION
    | STD_COMMAND_AND_FUNCTION
    | MODIFIER_AND_FUNCTION
    | TIME_AND_FUNCTION
    | LIKE
    ;

command
    : INIT_COMMAND
    | STD_COMMAND
    | STD_COMMAND_AND_FUNCTION
    ;


// Source-aware entry point. Legacy query contexts above remain available to mapper consumers.
analysisQuery : analysisPipeline EOF;
analysisPipeline
    : PIPE analysisStage (PIPE analysisStage)*
    | analysisInitialStage (PIPE analysisStage)*
    ;
analysisInitialStage
    : {p.analysisIsCommand()}? analysisStage
    | analysisImplicitSearch
    ;
analysisImplicitSearch : analysisSearch;
analysisStage
    : {p.analysisCommandIs("search")}? analysisCommandName analysisSearch #AnalysisSearchStage
    | {p.analysisCommandIs("where")}? analysisCommandName analysisExpression #AnalysisWhereStage
    | {p.analysisCommandIs("eval")}? analysisCommandName analysisAssignment (COMMA analysisAssignment)* #AnalysisEvalStage
    | {p.analysisCommandIs("rename")}? analysisCommandName analysisRename (COMMA? analysisRename)* #AnalysisRenameStage
    | {p.analysisCommandIs("fields", "table")}? analysisCommandName (ADD | SUB)? analysisFieldList #AnalysisFieldsStage
    | {p.analysisCommandIs("stats", "eventstats", "streamstats")}? analysisCommandName analysisAggregate (COMMA? analysisAggregate)* analysisGroup? #AnalysisStatsStage
    | {p.analysisCommandIs("lookup")}? analysisCommandName analysisLookup #AnalysisLookupStage
    | {p.analysisCommandIs("sort")}? analysisCommandName analysisLimit? analysisSortField (COMMA? analysisSortField)* #AnalysisSortStage
    | {p.analysisCommandIs("dedup")}? analysisCommandName analysisLimit? analysisOption* analysisFieldList #AnalysisDedupStage
    | {p.analysisCommandIs("head", "tail")}? analysisCommandName (analysisLimit | analysisExpression)? #AnalysisLimitStage
    | {p.analysisCommandIs("inputlookup")}? analysisCommandName analysisOption* analysisCatalogName #AnalysisInputlookupStage
    | {!p.analysisCommandIs("search", "where", "eval", "rename", "fields", "table", "stats", "eventstats", "streamstats", "lookup", "sort", "dedup", "head", "tail", "inputlookup")}? analysisCommandName analysisArgument* #AnalysisOpaqueStage
    ;
analysisCommandName : INIT_COMMAND | STD_COMMAND | STD_COMMAND_AND_FUNCTION | IDENTIFIER;
analysisSortField : (ADD | SUB)? analysisSelector;
analysisLimit : NUMBER;
analysisOption : analysisIdentifier EQ (analysisLiteral | analysisIdentifier);
analysisCatalogName : analysisIdentifier | STRING;
analysisAssignment : analysisIdentifier EQ analysisExpression;
analysisRename : analysisSelector analysisAlias;
analysisAlias : AS analysisIdentifier;
analysisFieldList : analysisSelector (COMMA? analysisSelector)*;
analysisGroup : BY analysisFieldList;
analysisAggregate : (analysisFunctionCall | analysisIdentifier) analysisAlias?;
analysisLookup : analysisOption* analysisCatalogName analysisLookupInput (COMMA? analysisLookupInput)* analysisOutput*;
analysisLookupInput : analysisIdentifier analysisAlias?;
analysisOutput : (OUTPUT | OUTPUTNEW) analysisLookupInput (COMMA? analysisLookupInput)*;
analysisArgument
    : analysisAssignment
    | analysisExpression analysisAlias?
    | analysisGroup
    | analysisOutput
    | COMMA
    ;
analysisSubquery : LBRACK analysisPipeline RBRACK;
analysisMacro : BACKTICK analysisIdentifier (LPAREN analysisArgumentList? RPAREN)? BACKTICK;
analysisArgumentList : analysisExpression (COMMA analysisExpression)*;

// Search values have their own context: bare comparison RHS and unqualified terms are literals.
analysisSearch : analysisSearchAnd (OR analysisSearchAnd)*;
analysisSearchAnd : analysisSearchUnary (AND? analysisSearchUnary)*;
analysisSearchUnary : NOT analysisSearchUnary | analysisSearchTerm;
analysisSearchTerm
    : analysisIdentifier analysisComparisonOperator analysisSearchValue
    | analysisIdentifier IN LPAREN analysisSearchValue (COMMA analysisSearchValue)* RPAREN
    | LPAREN analysisSearch RPAREN
    | analysisMacro
    | analysisSubquery
    | analysisSearchValue
    ;
analysisSearchValue
    : STRING | TIME
    | (ADD | SUB)? NUMBER analysisIdentifier?
    | MULT? analysisIdentifier (DIV analysisIdentifier)* MULT?
    | DIV analysisIdentifier (DIV analysisIdentifier)* MULT?
    | MULT
    ;
analysisExpression : analysisOr;
analysisOr : analysisAnd (OR analysisAnd)*;
analysisAnd : analysisNot (AND analysisNot)*;
analysisNot : NOT analysisNot | analysisComparison;
analysisComparison : analysisConcat (analysisComparisonOperator analysisConcat | IN LPAREN analysisArgumentList? RPAREN | LIKE analysisConcat)?;
analysisComparisonOperator : EQ | NE | LT | LE | GT | GE;
analysisConcat : analysisAdd (DOT analysisAdd)*;
analysisAdd : analysisMultiply ((ADD | SUB) analysisMultiply)*;
analysisMultiply : analysisPower ((MULT | DIV | MOD) analysisPower)*;
analysisPower : analysisUnary (POW analysisPower)?;
analysisUnary : (ADD | SUB) analysisUnary | analysisAtom;
analysisAtom
    : analysisFunctionCall
    | analysisLiteral
    | analysisIdentifier
    | MULT
    | LPAREN analysisExpression RPAREN
    | analysisSubquery
    | analysisMacro
    ;
analysisFunctionCall : analysisFunctionName LPAREN analysisArgumentList? RPAREN;
analysisFunctionName : analysisIdentifier | IN | LIKE;
analysisLiteral : STRING | NUMBER | TIME;
analysisSelector : MULT? analysisIdentifier MULT? | MULT;
analysisIdentifier
    : IDENTIFIER | QUOTED_IDENTIFIER | INIT_COMMAND | STD_COMMAND | STD_COMMAND_AND_FUNCTION
    | FUNCTION | MODIFIER_AND_FUNCTION | TIME_AND_FUNCTION
    ;
