lexer grammar SPL2Lexer;

// Every delimiter stack belongs to this lexer instance. Interpolation returns to
// its owning quote mode; object/lambda braces nest without closing that quote.
@structmembers { braceDepth int }
@members {
func (l *SPL2Lexer) openBrace() { l.braceDepth++; l.PushMode(antlr.LexerDefaultMode) }
func (l *SPL2Lexer) closeBrace() { if l.braceDepth > 0 { l.braceDepth--; l.PopMode() } }
}

FROM: 'FROM' | 'from'; SELECT: 'SELECT' | 'select';
DISTINCT: 'DISTINCT' | 'distinct'; HAVING: 'HAVING' | 'having';
GROUPBY: 'GROUPBY' | 'groupby'; ORDER: 'ORDER' | 'order'; ORDERBY: 'ORDERBY' | 'orderby';
LIMIT: 'LIMIT' | 'limit'; OFFSET: 'OFFSET' | 'offset';
ASC: 'ASC' | 'asc'; DESC: 'DESC' | 'desc'; BY_LOWER: 'by';
JOIN: 'JOIN' | 'join'; INNER: 'INNER' | 'inner'; LEFT: 'LEFT' | 'left'; OUTER: 'OUTER' | 'outer';
ON: 'ON' | 'on'; EXISTS: 'EXISTS' | 'exists';
SEARCH: 'search'; INDEX: 'index'; EVAL: 'eval'; WHERE: 'where' | 'WHERE';
FIELDS: 'fields'; TABLE: 'table'; AS: 'AS'; AS_LOWER: 'as'; GROUP: 'GROUP' | 'group';
RENAME: 'rename'; STATS: 'stats'; EVENTSTATS: 'eventstats'; STREAMSTATS: 'streamstats';
LOOKUP: 'lookup'; SORT: 'sort'; DEDUP: 'dedup'; HEAD: 'head'; REVERSE: 'reverse';
BY: 'BY'; OUTPUT: 'OUTPUT'; OUTPUTNEW: 'OUTPUTNEW';
ALLNUM: 'allnum'; DELIM: 'delim'; PARTITIONS: 'partitions'; SPAN: 'span';
CURRENT: 'current'; RESET: 'reset'; BEFORE: 'before'; AFTER: 'after'; ONCHANGE: 'onchange'; WINDOW: 'window';
KEEPEMPTY: 'keepempty'; CONSECUTIVE: 'consecutive'; KEEPLAST: 'keeplast'; WHILE: 'while';
AUTO: 'auto'; IP: 'ip'; NUM: 'num'; STR: 'str';
TERM: 'TERM'; CASE: 'CASE'; EARLIEST: 'earliest'; LATEST: 'latest';
INDEX_EARLIEST: '_index_earliest'; INDEX_LATEST: '_index_latest';
TIMEFORMAT: 'timeformat'; STARTTIME: 'starttime'; ENDTIME: 'endtime'; NOW: 'now'; AT: '@';
REX: 'rex' -> pushMode(REGEX_INPUT);
APPEND: 'append';
APPENDPIPE: 'appendpipe';
APPENDCOLS: 'appendcols';
UNION: 'union';
IF: 'if';
ELSEIF: 'elseif';
ELSE: 'else';
BIN: 'bin';
SPATH: 'spath';
LOADJOB: 'loadjob';
TSTATS: 'tstats';
MSTATS: 'mstats';
TIMECHART: 'timechart';
TIMEWRAP: 'timewrap';
MAKEMV: 'makemv';
MVEXPAND: 'mvexpand';
MVCOMBINE: 'mvcombine';
FILLNULL: 'fillnull';
RIGHT: 'right';
RUN_IN_PREVIEW: 'run_in_preview';
MAX: 'max';
BINS: 'bins';
MINSPAN: 'minspan';
START: 'start';
END: 'end';
ALIGNTIME: 'aligntime';
FIELD: 'field';
MAX_MATCH: 'max_match';
OFFSET_FIELD: 'offset_field';
MODE: 'mode';
SED: 'sed';
INPUT: 'input';
PATH: 'path';
OUTPUT_LOWER: 'output';
AGGREGATES: 'aggregates';
PREDICATE: 'predicate';
BYFIELDS: 'byfields';
DATAMODEL_NAME: 'datamodel_name';
SEP: 'sep';
FORMAT: 'format';
PARTIAL: 'partial';
CONT: 'cont';
FIXEDRANGE: 'fixedrange';
AGG: 'agg';
USENULL: 'usenull';
USEOTHER: 'useother';
NULLSTR: 'nullstr';
OTHERSTR: 'otherstr';
ALIGN: 'align';
TOKENIZER: 'tokenizer';
VALUE: 'value';
MAKERESULTS: 'makeresults'; SPL1: 'spl1';
IMPORT: 'import'; EXPORT: 'export'; FUNCTION: 'function'; RETURN: 'return';
AND: 'AND'; OR: 'OR'; XOR: 'XOR'; NOT: 'NOT';
BETWEEN: 'BETWEEN'; IN: 'IN'; LIKE: 'LIKE'; IS: 'IS';
NULL: 'null'; NULL_TEST: 'NULL'; BOOLEAN: 'true' | 'false';
TYPE_OPTION: 'type';
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
LOG_SPAN: 'log' ([0-9]+ ('.' [0-9]+)?)?;
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
RX_FIELD: 'field' -> type(FIELD);
RX_MAX_MATCH: 'max_match' -> type(MAX_MATCH);
RX_OFFSET_FIELD: 'offset_field' -> type(OFFSET_FIELD);
RX_MODE: 'mode' -> type(MODE);
RX_SED: 'sed' -> type(SED);
RX_ID: [a-zA-Z_] [a-zA-Z_0-9]* -> type(IDENTIFIER);
RX_NUMBER: [0-9]+ ('.' [0-9]+)? -> type(NUMBER);
RX_MINUS: '-' -> type(MINUS);
RX_PLUS: '+' -> type(PLUS);
RX_CLOSE: ']' -> type(RBRACKET), popMode;
RX_EQ: '=' -> type(ASSIGN);
RX_PIPE: '|' -> type(PIPE), popMode;
RX_WS: [ \t\r\n]+ -> channel(HIDDEN);
