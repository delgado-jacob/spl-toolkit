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
    : IDENTIFIER    #FieldUse
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
