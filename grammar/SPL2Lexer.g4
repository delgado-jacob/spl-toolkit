lexer grammar SPL2Lexer;

// Every delimiter stack belongs to this lexer instance. Interpolation returns to
// its owning quote mode; object/lambda braces nest without closing that quote.
@structmembers { braceDepth int }
@members {
func (l *SPL2Lexer) openBrace() { l.braceDepth++; l.PushMode(antlr.LexerDefaultMode) }
func (l *SPL2Lexer) closeBrace() { if l.braceDepth > 0 { l.braceDepth--; l.PopMode() } }
}

FROM: 'FROM' | 'from'; SELECT: 'SELECT' | 'select';
SEARCH: 'search'; INDEX: 'index'; EVAL: 'eval'; WHERE: 'where' | 'WHERE';
FIELDS: 'fields'; TABLE: 'table'; AS: 'AS'; GROUP: 'GROUP' | 'group';
REX: 'rex' -> pushMode(REGEX_INPUT);
MAKERESULTS: 'makeresults'; SPL1: 'spl1';
IMPORT: 'import'; EXPORT: 'export'; FUNCTION: 'function'; RETURN: 'return';
AND: 'AND'; OR: 'OR'; XOR: 'XOR'; NOT: 'NOT';
BETWEEN: 'BETWEEN'; IN: 'IN'; LIKE: 'LIKE'; IS: 'IS';
NULL: 'null'; NULL_TEST: 'NULL'; BOOLEAN: 'true' | 'false';
TYPE: 'int' | 'long' | 'float' | 'double' | 'string' | 'boolean';
ARROW: '->'; LE: '<='; GE: '>='; NE: '!='; EQ: '==';
ASSIGN: '='; LT: '<'; GT: '>'; PLUS: '+'; MINUS: '-'; STAR: '*';
LINE_COMMENT: '//' ~[\r\n]* -> channel(HIDDEN);
BLOCK_COMMENT: '/*' .*? '*/' -> channel(HIDDEN);
SLASH: '/'; MOD: '%'; PIPE: '|'; COMMA: ','; COLON: ':'; DOT: '.';
LPAREN: '('; RPAREN: ')'; LBRACKET: '['; RBRACKET: ']';
LBRACE: '{' {l.openBrace()}; RBRACE: '}' {l.closeBrace()}; SEMI: ';';
RAW_STRING: '@"' ('""' | ~'"')* '"';
DQUOTE: '"' -> pushMode(STRING_MODE);
SQUOTE: '\'' -> pushMode(NAME_MODE);
BACKTICK: '`' -> pushMode(EMBEDDED_MODE);
NUMBER: [0-9]+ ('.' [0-9]+)? ([eE] [+-]? [0-9]+)? [LFD]?;
LOCAL: '$' [a-zA-Z_] [a-zA-Z_0-9]*;
IDENTIFIER: [a-zA-Z_] [a-zA-Z_0-9]*;
NL: '\r'? '\n' | '\r';
WS: [ \t]+ -> channel(HIDDEN);

mode STRING_MODE;
STRING_END: '"' -> popMode;
STRING_INTERPOLATION: '${' {l.openBrace()};
STRING_TEXT: ('\\' . | ~["\\$])+;
STRING_DOLLAR: '$';

mode NAME_MODE;
NAME_END: '\'' -> popMode;
NAME_INTERPOLATION: '${' {l.openBrace()};
NAME_TEXT: ('\\' . | ~['\\$])+;
NAME_DOLLAR: '$';

mode EMBEDDED_MODE;
EMBEDDED_END: '`' -> popMode;
EMBEDDED_TEXT: ~'`'+;

// Slash regexes exist only in rex command context. Ordinary a/b/c always
// remains three operands and two division tokens, including inside templates.
mode REGEX_INPUT;
REGEX: '/' ('\\' . | ~[/\r\n])+ '/' ;
RX_RAW: '@"' ('""' | ~'"')* '"' -> type(RAW_STRING);
RX_QUOTE: '"' -> type(DQUOTE), pushMode(STRING_MODE);
RX_NAME: '\'' -> type(SQUOTE), pushMode(NAME_MODE);
RX_ID: [a-zA-Z_] [a-zA-Z_0-9]* -> type(IDENTIFIER);
RX_NUMBER: [0-9]+ -> type(NUMBER);
RX_EQ: '=' -> type(ASSIGN);
RX_PIPE: '|' -> type(PIPE), popMode;
RX_WS: [ \t\r\n]+ -> channel(HIDDEN);
