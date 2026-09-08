parser grammar SPL2Parser;
options { tokenVocab=SPL2Lexer; }

@header {
import "strings"
// ANTLR also emits the header into the visitor interface file.
var _ = strings.EqualFold
}
@members {
// Casing is recognized only in an operator grammar position. The same lower-
// case token remains an ordinary identifier or search literal elsewhere.
func (p *SPL2Parser) contextualKeyword(word string) bool {
    return strings.EqualFold(p.GetTokenStream().LT(1).GetText(), word)
}
}

query: NL* (pipeline moduleSuffix? | moduleDeclaration) NL* EOF;
pipeline: start (NL* PIPE NL* command)*;
start: fromCommand | selectCommand | searchCommand | implicitSearch | generator | embeddedCommand;
command: evalCommand | whereCommand | fieldsCommand | tableCommand | renameCommand
    | statsCommand | eventstatsCommand | streamstatsCommand | lookupCommand
    | sortCommand | dedupCommand | headCommand | reverseCommand
    | fromCommand | selectCommand | searchCommand | rexCommand | embeddedCommand;
fromCommand: FROM dataset;
selectCommand: SELECT projection (COMMA projection)* FROM dataset;
projection: expression (AS identifier)?;
dataset: identifier | array;
generator: MAKERESULTS NUMBER?;
evalCommand: EVAL assignment (COMMA assignment)*;
assignment: fieldName ASSIGN expression;
whereCommand: WHERE expression;
fieldsCommand: FIELDS fieldSelection;
fieldSelection: (PLUS | MINUS)? fieldSelector (COMMA fieldSelector)*;
fieldSelector: identifier;
tableCommand: TABLE tableField (COMMA tableField)*;
tableField: identifier | stringLiteral;
renameCommand: RENAME renamePair (COMMA renamePair)*;
renamePair: renameSource aliasKeyword renameTarget;
renameSource: identifier;
renameTarget: identifier;
aliasKeyword: AS | AS_LOWER;
statsCommand: STATS statsOption* aggregate (COMMA aggregate)* aggregateGroup?;
statsOption: allnumOption | delimOption | partitionsOption | unknownOption;
allnumOption: ALLNUM ASSIGN BOOLEAN;
delimOption: DELIM ASSIGN stringLiteral;
partitionsOption: PARTITIONS ASSIGN (PLUS | MINUS)? NUMBER;
aggregate: call (aliasKeyword aggregateAlias)?;
aggregateAlias: identifier;
aggregateGroup: BY groupField (COMMA groupField)*;
groupField: identifier groupSpan?;
groupSpan: SPAN ASSIGN timeSpan;
timeSpan: NUMBER? IDENTIFIER;
eventstatsCommand: EVENTSTATS (allnumOption | unknownOption)* aggregate (COMMA aggregate)* aggregateGroup?;
streamstatsCommand: STREAMSTATS unknownOption* streamGroup? currentOption? resetClause? windowOption? aggregate (COMMA aggregate)* streamPostLayout?;
streamGroup: BY groupField (COMMA groupField)*;
currentOption: CURRENT ASSIGN BOOLEAN;
windowOption: WINDOW ASSIGN NUMBER;
resetClause: RESET (resetBefore resetAfter? resetOnchange? | resetAfter resetOnchange? | resetOnchange);
resetBefore: BEFORE expression;
resetAfter: AFTER expression;
resetOnchange: ONCHANGE;
// The documented contradictory layouts have explicit ownership and remain held.
streamPostLayout: streamGroup resetClause? | resetClause;
lookupCommand: LOOKUP unknownOption* lookupDataset lookupMatch (COMMA lookupMatch)* lookupOutputClause?;
lookupDataset: identifier;
lookupMatch: lookupColumn (aliasKeyword lookupEventField)?;
lookupOutputClause: (OUTPUT | OUTPUTNEW) lookupOutput (COMMA lookupOutput)*;
lookupOutput: lookupColumn (aliasKeyword lookupEventField)?;
lookupColumn: identifier;
lookupEventField: identifier;
sortCommand: SORT unknownOption* integerValue? sortTerm (COMMA sortTerm)*;
sortTerm: (PLUS | MINUS)? (sortWrapper LPAREN identifier RPAREN | identifier);
sortWrapper: AUTO | IP | NUM | STR;
integerValue: (PLUS | MINUS)? NUMBER;
dedupCommand: DEDUP unknownOption* integerValue? keepemptyOption? consecutiveOption? dedupField (COMMA dedupField)*;
keepemptyOption: KEEPEMPTY ASSIGN BOOLEAN;
consecutiveOption: CONSECUTIVE ASSIGN BOOLEAN;
dedupField: identifier;
headCommand: HEAD unknownOption* (keeplastOption? headWhile? integerValue? | headPostLayout);
keeplastOption: KEEPLAST ASSIGN BOOLEAN;
headWhile: WHILE LPAREN expression RPAREN;
headPostLayout: integerValue keeplastOption? headWhile;
reverseCommand: REVERSE;
// Unknown option literals are retained as bounded typed nodes, never an opaque
// command body. Known option tokens cannot fall through this alternative.
unknownOption: IDENTIFIER ASSIGN (literal | identifier);
rexCommand: REX rexOption* (REGEX | stringLiteral | RAW_STRING);
rexOption: IDENTIFIER ASSIGN (identifier | NUMBER);
embeddedCommand: SPL1? embeddedText;
embeddedText: BACKTICK EMBEDDED_TEXT? EMBEDDED_END;
moduleSuffix: SEMI .*?;
moduleDeclaration: (IMPORT | EXPORT | FUNCTION | LOCAL ASSIGN) .*?;

searchCommand: SEARCH searchExpression;
// The token guard restricts entry, while the entire search owns precedence.
implicitSearch: {p.GetTokenStream().LA(1) == SPL2ParserINDEX && p.GetTokenStream().LA(2) == SPL2ParserASSIGN}? searchExpression;
searchExpression: searchXor;
searchXor: searchAnd (XOR searchAnd)*;
searchAnd: searchOr (AND? searchOr)*;
searchOr: searchNot (OR searchNot)*;
searchNot: NOT searchNot | searchAtom;
searchAtom: LPAREN searchExpression RPAREN | searchTimeModifier | identifier comparison searchValue | identifier IN LPAREN searchValue (COMMA searchValue)* RPAREN | searchValue;
searchValue: searchDirective | searchBareValue | stringLiteral | RAW_STRING;
searchBareValue: (identifier | NUMBER) ((DOT | MINUS | SLASH | COLON) (identifier | NUMBER) | STAR)* | STAR;
searchDirective: (TERM | CASE) LPAREN searchBareValue RPAREN;
searchTimeModifier: timeModifierKey comparison timeModifierValue | TIMEFORMAT ASSIGN stringLiteral;
timeModifierKey: EARLIEST | LATEST | INDEX_EARLIEST | INDEX_LATEST | STARTTIME | ENDTIME;
timeModifierValue: relativeTime | NUMBER | stringLiteral | NOW LPAREN RPAREN;
relativeTime: (PLUS | MINUS) NUMBER? IDENTIFIER (AT IDENTIFIER)? ((PLUS | MINUS) NUMBER? IDENTIFIER)? | AT IDENTIFIER ((PLUS | MINUS) NUMBER? IDENTIFIER)?;

expression: lambdaExpression | xorExpression;
xorExpression: orExpression (logicalXor orExpression)*;
orExpression: andExpression (logicalOr andExpression)*;
andExpression: notExpression (logicalAnd notExpression)*;
// Prefer a complete ordinary identifier expression when a contextual spelling
// also permits prefix NOT. The reserved uppercase token remains unambiguous.
notExpression: predicate | logicalNot notExpression;
predicate: additive (comparison additive | logicalNot? betweenOperator additive betweenConjunction additive | logicalNot? IN LPAREN expression (COMMA expression)* RPAREN | logicalNot? LIKE additive | IS (logicalNot? (NULL | NULL_TEST) | logicalNot? TYPE))?;
logicalAnd: AND | {p.contextualKeyword("and")}? IDENTIFIER;
logicalOr: OR | {p.contextualKeyword("or")}? IDENTIFIER;
logicalXor: XOR | {p.contextualKeyword("xor")}? IDENTIFIER;
logicalNot: NOT | {p.contextualKeyword("not")}? IDENTIFIER;
betweenOperator: BETWEEN | {p.contextualKeyword("between")}? IDENTIFIER;
betweenConjunction: AND | {p.contextualKeyword("and")}? IDENTIFIER;
comparison: ASSIGN | EQ | NE | LT | LE | GT | GE;
additive: multiplicative ((PLUS | MINUS) multiplicative)*;
multiplicative: unary ((STAR | SLASH | MOD) unary)*;
unary: (PLUS | MINUS) unary | access;
access: primary accessPart*;
accessPart: DOT identifier | LBRACKET expression RBRACKET;
primary: call | fieldName | LOCAL | literal | array | object | LPAREN expression RPAREN | searchLiteral;
call: identifier LPAREN arguments? RPAREN;
arguments: namedArgument (COMMA namedArgument)* | expression (COMMA expression)* (COMMA namedArgument)*;
namedArgument: identifier COLON expression;
literal: NUMBER | BOOLEAN | NULL | RAW_STRING | stringLiteral;
stringLiteral: DQUOTE (STRING_TEXT | STRING_DOLLAR | STRING_INTERPOLATION expression RBRACE)* STRING_END;
quotedName: SQUOTE (NAME_TEXT | NAME_DOLLAR)+ NAME_END;
fieldTemplate: SQUOTE (NAME_TEXT | NAME_DOLLAR)* NAME_INTERPOLATION expression RBRACE (NAME_TEXT | NAME_DOLLAR | NAME_INTERPOLATION expression RBRACE)* NAME_END;
fieldName: identifier | fieldTemplate;
identifier: IDENTIFIER | INDEX | quotedName | pipelineKeyword;
// New command/option tokens remain ordinary names in expression/name positions.
pipelineKeyword: AS_LOWER | RENAME | STATS | EVENTSTATS | STREAMSTATS | LOOKUP | SORT | DEDUP | HEAD | REVERSE
    | BY | OUTPUT | OUTPUTNEW | ALLNUM | DELIM | PARTITIONS | SPAN | CURRENT | RESET | BEFORE | AFTER | ONCHANGE | WINDOW
    | KEEPEMPTY | CONSECUTIVE | KEEPLAST | WHILE | AUTO | IP | NUM | STR | TERM | CASE
    | EARLIEST | LATEST | INDEX_EARLIEST | INDEX_LATEST | TIMEFORMAT | STARTTIME | ENDTIME | NOW;
array: LBRACKET (expression (COMMA expression)* COMMA?)? RBRACKET;
object: LBRACE (objectEntry (COMMA objectEntry)* COMMA?)? RBRACE;
objectEntry: objectKey COLON expression;
objectKey: identifier | stringLiteral;
searchLiteral: embeddedText;
lambdaExpression: (lambdaParameter | LPAREN (lambdaParameter (COMMA lambdaParameter)*)? RPAREN) ARROW (expression | lambdaBlock);
lambdaParameter: LOCAL (COLON TYPE)? (ASSIGN ((PLUS | MINUS)? NUMBER | stringLiteral))?;
lambdaBlock: LBRACE NL* (LOCAL ASSIGN expression (SEMI NL* | NL+))* RETURN expression SEMI? NL* RBRACE;
