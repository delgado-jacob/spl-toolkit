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
func (p *SPL2Parser) extendedUnknownOption() bool {
    token:=p.GetTokenStream().LA(1)
    for ctx:=p.GetParserRuleContext();ctx!=nil; {
        var known []int
        switch ctx.(type) {
        case *RexCommandContext: known=[]int{SPL2ParserFIELD,SPL2ParserMAX_MATCH,SPL2ParserOFFSET_FIELD,SPL2ParserMODE}
        case *AppendpipeCommandContext: known=[]int{SPL2ParserRUN_IN_PREVIEW}
        case *BinCommandContext: known=[]int{SPL2ParserBINS,SPL2ParserMINSPAN,SPL2ParserSPAN,SPL2ParserSTART,SPL2ParserEND,SPL2ParserALIGNTIME}
        case *SpathCommandContext: known=[]int{SPL2ParserINPUT,SPL2ParserPATH,SPL2ParserOUTPUT_LOWER}
        case *MakemvCommandContext: known=[]int{SPL2ParserDELIM,SPL2ParserTOKENIZER}
        case *MvexpandCommandContext: known=[]int{SPL2ParserLIMIT}
        case *MvcombineCommandContext: known=[]int{SPL2ParserDELIM}
        case *FillnullCommandContext: known=[]int{SPL2ParserVALUE}
        case *MetricsCommandContext: known=[]int{SPL2ParserAGGREGATES,SPL2ParserPREDICATE,SPL2ParserBYFIELDS,SPL2ParserDATAMODEL_NAME}
        case *TimechartCommandContext: known=[]int{SPL2ParserBINS,SPL2ParserMINSPAN,SPL2ParserSPAN,SPL2ParserSTART,SPL2ParserEND,SPL2ParserALIGNTIME,SPL2ParserSEP,SPL2ParserFORMAT,SPL2ParserPARTIAL,SPL2ParserCONT,SPL2ParserFIXEDRANGE,SPL2ParserLIMIT,SPL2ParserAGG,SPL2ParserUSENULL,SPL2ParserUSEOTHER,SPL2ParserNULLSTR,SPL2ParserOTHERSTR}
        }
        if known!=nil {for _,k:=range known {if k==token {return false}};return true}
        parent,_:=ctx.GetParent().(antlr.ParserRuleContext);ctx=parent
    }
    return true
}
func (p *SPL2Parser) adjacentPrevious() bool {
 return p.GetTokenStream().LT(-1).GetStop()+1 == p.GetTokenStream().LT(1).GetStart()
}
func (p *SPL2Parser) contextualKeyword(word string) bool {
    return strings.EqualFold(p.GetTokenStream().LT(1).GetText(), word)
}
func (p *SPL2Parser) selectedBranch() bool {
    if p.GetTokenStream().LT(1).GetText() != "branch" { return false }
    parens, brackets, braces := 0, 0, 0
    separators := 0
    for offset := 2; ; offset++ {
        token := p.GetTokenStream().LA(offset)
        if token == antlr.TokenEOF || (token == SPL2ParserPIPE || token == SPL2ParserRBRACKET) && parens == 0 && brackets == 0 && braces == 0 {
            return separators > 0 && parens == 0 && brackets == 0 && braces == 0
        }
        switch token {
        case SPL2ParserLPAREN:
            parens++
        case SPL2ParserRPAREN:
            if parens == 0 { return false }
            parens--
        case SPL2ParserLBRACKET:
            brackets++
        case SPL2ParserRBRACKET:
            if brackets == 0 { return false }
            brackets--
        case SPL2ParserLBRACE:
            braces++
        case SPL2ParserRBRACE:
            if braces == 0 { return false }
            braces--
        case SPL2ParserCOMMA:
            if parens == 0 && brackets == 0 && braces == 0 { separators++ }
        }
    }
}
// SQL clause words remain ordinary identifiers in previously supported pipeline
// expressions. Within SQL they are boundaries, not missing operand fallbacks.
func (p *SPL2Parser) outsideSQL() bool {
    for ctx := p.GetParserRuleContext(); ctx != nil; {
        switch ctx.(type) { case *FromCommandContext, *SelectCommandContext: return false }
        parent, _ := ctx.GetParent().(antlr.ParserRuleContext)
        ctx = parent
    }
    return true
}
// Supported options must use their owning command's strict production. Tokens
// documented for other commands remain typed, unproved options here.
func (p *SPL2Parser) unreviewedOption(command int) bool {
    token := p.GetTokenStream().LA(1)
    switch command {
    case SPL2ParserSTATS:
        return token != SPL2ParserBY && token != SPL2ParserALLNUM && token != SPL2ParserDELIM && token != SPL2ParserPARTITIONS
    case SPL2ParserEVENTSTATS:
        return token != SPL2ParserBY && token != SPL2ParserALLNUM
    case SPL2ParserSTREAMSTATS:
        return token != SPL2ParserBY && token != SPL2ParserCURRENT && token != SPL2ParserRESET && token != SPL2ParserWINDOW
    case SPL2ParserDEDUP:
        return token != SPL2ParserKEEPEMPTY && token != SPL2ParserCONSECUTIVE
    case SPL2ParserHEAD:
        return token != SPL2ParserKEEPLAST && token != SPL2ParserWHILE
    }
    return true
}
}

query: NL* (pipeline moduleSuffix? | moduleDeclaration) NL* EOF;
pipeline: start (NL* PIPE NL* command)*;
start: fromCommand | selectCommand | searchCommand | implicitSearch | generator | loadjobCommand | metricsCommand | unionCommand | branchCommand | embeddedCommand;
command: evalCommand | whereCommand | fieldsCommand | tableCommand | renameCommand
    | statsCommand | eventstatsCommand | streamstatsCommand | lookupCommand
    | sortCommand | dedupCommand | headCommand | reverseCommand
    | fromCommand | selectCommand | searchCommand | rexCommand | embeddedCommand
    | joinCommand | appendCommand | appendpipeCommand | appendcolsCommand | unionCommand | branchCommand | ifCommand
    | binCommand | spathCommand | timechartCommand | timewrapCommand | makemvCommand | mvexpandCommand | mvcombineCommand | fillnullCommand;
// Clause contexts own only their original lexical spans. Logical SQL scheduling
// is deliberately deferred to lowering, not represented by reordered source.
fromCommand: sqlFromClause (NL* sqlWhereClause)? (NL* sqlGroupClause NL* sqlSelectClause | NL* sqlSelectClause)?
    (NL* sqlHavingClause)? (NL* sqlOrderClause)? (NL* sqlLimitClause)? (NL* sqlOffsetClause)?;
selectCommand: sqlSelectClause NL* sqlFromClause (NL* sqlWhereClause)? (NL* sqlGroupClause)?
    (NL* sqlHavingClause)? (NL* sqlOrderClause)? (NL* sqlLimitClause)? (NL* sqlOffsetClause)?
    | {p.GetTokenStream().LA(2) == SPL2ParserNL}? multilineSqlSelectClause NL+ multilineSqlFromClause
    (NL* multilineSqlWhereClause)? (NL* multilineSqlGroupClause)? (NL* sqlHavingClause)?
    (NL* multilineSqlOrderClause)? (NL* sqlLimitClause)? (NL* sqlOffsetClause)?;
sqlFromClause: FROM dataset sourceAlias? (NL* sqlJoinClause)*;
sourceAlias: aliasKeyword identifier;
sqlJoinClause: (INNER | LEFT OUTER?)? JOIN dataset sourceAlias ON sqlJoinPredicate;
sqlJoinPredicate: sqlJoinEquality (AND sqlJoinEquality)*;
sqlJoinEquality: sqlJoinField ASSIGN sqlJoinField;
sqlJoinField: identifier DOT identifier accessPart*;
sqlSelectClause: SELECT DISTINCT? projection (COMMA NL* projection)*;
projection: expression (aliasKeyword projectionAlias)?;
projectionAlias: identifier;
sqlWhereClause: WHERE sqlPredicate;
sqlHavingClause: HAVING sqlPredicate;
sqlPredicate: expression;
sqlGroupClause: (GROUP sqlBy | GROUPBY) sqlGroupKey (COMMA sqlGroupKey)*;
sqlBy: BY | BY_LOWER;
sqlGroupKey: sqlSpanCall | {p.GetTokenStream().LA(1) != SPL2ParserSPAN}? expression sqlSpanAssignment?;
sqlSpanCall: SPAN LPAREN fieldName (COMMA timeSpan)? RPAREN;
sqlSpanAssignment: SPAN ASSIGN (LPAREN timeSpan RPAREN | sqlUnparenthesizedSpan);
sqlUnparenthesizedSpan: timeSpan;
sqlOrderClause: (ORDER sqlBy | ORDERBY) sqlOrderTerm (COMMA sqlOrderTerm)*;
sqlOrderTerm: expression sqlDirection?;
sqlDirection: ASC | DESC;
sqlLimitClause: LIMIT integerValue;
sqlOffsetClause: OFFSET integerValue;
multilineSqlSelectClause: SELECT NL+ DISTINCT? NL* multilineSqlProjection (NL* COMMA NL* multilineSqlProjection)*;
multilineSqlProjection: (multilineCall | multilineOperand) (aliasKeyword projectionAlias)?;
multilineSqlFromClause: FROM NL+ dataset (NL* sourceAlias)? (NL* sqlJoinClause)*;
multilineSqlWhereClause: WHERE NL+ multilineSqlPredicate;
multilineSqlPredicate: multilineOperand comparison multilineOperand;
multilineSqlGroupClause: (GROUP sqlBy | GROUPBY) NL+ multilineSqlGroupKey (NL* COMMA NL* multilineSqlGroupKey)*;
multilineSqlGroupKey: multilineSqlSpanCall | multilineOperand;
multilineSqlSpanCall: SPAN LPAREN multilineOperand COMMA timeSpan RPAREN;
multilineSqlOrderClause: (ORDER sqlBy | ORDERBY) NL+ multilineSqlOrderTerm (NL* COMMA NL* multilineSqlOrderTerm)*;
multilineSqlOrderTerm: multilineOperand sqlDirection?;
existsPredicate: EXISTS LPAREN (fromCommand | selectCommand) RPAREN;
dataset: identifier | dottedDataset | datasetParameter | staticDatasetDescriptor | array;
dottedDataset: identifier datasetPath;
datasetPath: (NL* DOT NL* identifier)+;
datasetParameter: LOCAL;
// Preserve dynamic descriptor expressions for located semantic evidence. Exact
// identity, key uniqueness, and canonical properties belong to lowering.
staticDatasetDescriptor: LBRACE NL* descriptorKindKey NL* COLON NL* (jsonStringLiteral | expression)
    (NL* COMMA NL* descriptorPropertiesKey NL* COLON NL* (descriptorProperties | expression))? NL* RBRACE;
descriptorKindKey: {p.contextualKeyword("kind")}? IDENTIFIER;
descriptorPropertiesKey: {p.contextualKeyword("properties")}? IDENTIFIER;
descriptorProperties: LBRACE NL* (descriptorProperty (NL* COMMA NL* descriptorProperty)* NL* COMMA?)? NL* RBRACE;
descriptorProperty: jsonObjectKey NL* COLON NL* jsonLiteral;
jsonObjectKey: identifier | jsonStringLiteral;
jsonLiteral: NUMBER | BOOLEAN | NULL | jsonStringLiteral | jsonArray | jsonObject;
jsonStringLiteral: DQUOTE (STRING_TEXT | STRING_DOLLAR)* STRING_END;
jsonArray: LBRACKET NL* (jsonLiteral (NL* COMMA NL* jsonLiteral)*)? NL* RBRACKET;
jsonObject: LBRACE NL* (descriptorProperty (NL* COMMA NL* descriptorProperty)*)? NL* RBRACE;
generator: MAKERESULTS extendedOption* integerValue?;
evalCommand: EVAL assignment (COMMA NL* assignment)*
    | EVAL NL+ assignment (NL* COMMA NL* assignment)*;
assignment: fieldName ASSIGN expression;
whereCommand: WHERE NL* expression;
fieldsCommand: FIELDS fieldSelection;
fieldSelection: (PLUS | MINUS)? fieldSelector (NL* COMMA NL* fieldSelector)*;
fieldSelector: identifier | structuralFieldSelector;
structuralFieldSelector: identifier fieldPath;
fieldPath: (NL* DOT NL* identifier)+;
tableCommand: TABLE tableField (COMMA tableField)*;
tableField: identifier | stringLiteral;
renameCommand: RENAME renamePair (COMMA renamePair)*;
renamePair: renameSource aliasKeyword renameTarget;
renameSource: identifier;
renameTarget: identifier;
aliasKeyword: AS | AS_LOWER;
statsCommand: STATS statsOption* aggregate (COMMA NL* aggregate)* (NL* (aggregateGroup | selectedAggregateGroup))?
    | STATS NL+ (statsOption NL*)* aggregate (NL* COMMA NL* aggregate)*
      (NL* (aggregateGroup | selectedAggregateGroup))?;
statsOption: allnumOption | delimOption | partitionsOption | {p.unreviewedOption(SPL2ParserSTATS)}? unknownOption;
allnumOption: ALLNUM ASSIGN BOOLEAN;
delimOption: DELIM ASSIGN stringLiteral;
partitionsOption: PARTITIONS ASSIGN (PLUS | MINUS)? NUMBER;
aggregate: call (aliasKeyword aggregateAlias)?;
aggregateAlias: identifier;
aggregateGroup: BY groupField (COMMA groupField)*
    | BY NL+ groupField (NL* COMMA NL* groupField)*;
selectedAggregateGroup: SELECTED_BY NL* selectedGroupTerm (NL* COMMA NL* selectedGroupTerm)*;
selectedGroupTerm: selectedSpanGroup | expression;
selectedSpanGroup: SPAN LPAREN expression COMMA timeSpan RPAREN;
groupField: identifier groupSpan?;
groupSpan: SPAN ASSIGN timeSpan;
timeSpan: NUMBER? IDENTIFIER;
eventstatsCommand: EVENTSTATS (allnumOption | {p.unreviewedOption(SPL2ParserEVENTSTATS)}? unknownOption)* aggregate (COMMA aggregate)* aggregateGroup?;
streamstatsCommand: STREAMSTATS ({p.unreviewedOption(SPL2ParserSTREAMSTATS)}? unknownOption)* streamGroup? currentOption? resetClause? windowOption? aggregate (COMMA aggregate)* streamPostLayout?;
streamGroup: BY groupField (COMMA groupField)*;
currentOption: CURRENT ASSIGN BOOLEAN;
windowOption: WINDOW ASSIGN NUMBER;
resetClause: RESET (resetBefore resetAfter? resetOnchange? | resetAfter resetOnchange? | resetOnchange);
resetBefore: BEFORE expression;
resetAfter: AFTER expression;
resetOnchange: ONCHANGE;
// The documented contradictory layouts have explicit ownership and remain held.
streamPostLayout: streamGroup resetClause? | resetClause;
lookupCommand: LOOKUP ({p.unreviewedOption(SPL2ParserLOOKUP)}? unknownOption)* lookupDataset lookupMatch (COMMA lookupMatch)* lookupOutputClause?;
lookupDataset: identifier;
lookupMatch: lookupColumn (aliasKeyword lookupEventField)?;
lookupOutputClause: (OUTPUT | OUTPUTNEW) lookupOutput (COMMA lookupOutput)*;
lookupOutput: lookupColumn (aliasKeyword lookupEventField)?;
lookupColumn: identifier;
lookupEventField: identifier;
sortCommand: SORT ({p.unreviewedOption(SPL2ParserSORT)}? unknownOption)* integerValue? sortTerm (COMMA sortTerm)*;
sortTerm: (PLUS | MINUS)? (sortWrapper LPAREN identifier RPAREN | identifier);
sortWrapper: AUTO | IP | NUM | STR;
integerValue: (PLUS | MINUS)? NUMBER;
dedupCommand: DEDUP ({p.unreviewedOption(SPL2ParserDEDUP)}? unknownOption)* integerValue? keepemptyOption? consecutiveOption? dedupField (COMMA dedupField)*;
keepemptyOption: KEEPEMPTY ASSIGN BOOLEAN;
consecutiveOption: CONSECUTIVE ASSIGN BOOLEAN;
dedupField: identifier;
headCommand: HEAD ({p.unreviewedOption(SPL2ParserHEAD)}? unknownOption)* (keeplastOption? headWhile? integerValue? | headPostLayout);
keeplastOption: KEEPLAST ASSIGN BOOLEAN;
headWhile: WHILE LPAREN expression RPAREN;
headPostLayout: integerValue keeplastOption? headWhile;
reverseCommand: REVERSE;
// Unknown option literals are retained as bounded typed nodes, never an opaque
// command body. Known options of this command cannot fall through here.
unknownOption: unknownOptionName ASSIGN (literal | identifier);
unknownOptionName: IDENTIFIER | INDEX | pipelineKeyword | sqlKeyword;
// Bracket roles have distinct contexts; no array can become a command child.
independentSearch: LBRACKET NL* pipeline NL* RBRACKET;
inheritedSubpipe: LBRACKET NL* command (NL* PIPE NL* command)* NL* RBRACKET;
joinCommand: JOIN joinOption* WHERE sqlJoinPredicate independentSearch;
joinOption: LEFT ASSIGN identifier | RIGHT ASSIGN identifier | TYPE_OPTION ASSIGN joinType | MAX ASSIGN integerValue | {p.GetTokenStream().LA(1) != SPL2ParserLEFT && p.GetTokenStream().LA(1) != SPL2ParserRIGHT && p.GetTokenStream().LA(1) != SPL2ParserTYPE_OPTION && p.GetTokenStream().LA(1) != SPL2ParserMAX}? unknownOption;
joinType: INNER | LEFT | OUTER | identifier;
appendCommand: APPEND extendedOption* independentSearch;
appendpipeCommand: APPENDPIPE extendedOption* (RUN_IN_PREVIEW ASSIGN BOOLEAN)? inheritedSubpipe;
appendcolsCommand: APPENDCOLS extendedOption* independentSearch;
unionCommand: UNION NL* unionDataset (NL* COMMA NL* unionDataset)*;
unionDataset: independentSearch | dataset;
branchCommand: {p.selectedBranch()}? IDENTIFIER NL* branchArm (NL* COMMA NL* branchArm)+;
branchArm: LPAREN NL* expression NL* RPAREN NL* inheritedSubpipe;
ifCommand: IF NL* LPAREN NL* expression NL* RPAREN NL* inheritedSubpipe
    (NL* ELSEIF NL* LPAREN NL* expression NL* RPAREN NL* inheritedSubpipe)*
    (NL* ELSE NL* inheritedSubpipe)?;
binCommand: BIN (binOption | extendedOption)* identifier (aliasKeyword identifier)?;
binOption: BINS ASSIGN integerValue | MINSPAN ASSIGN binSpan | SPAN ASSIGN binSpan
    | (START | END) ASSIGN signedNumber | alignmentOption;
alignmentOption: ALIGNTIME ASSIGN (EARLIEST | LATEST | relativeTime);
binSpan: logarithmicSpan | signedWeeklySpan | (signedNumber ({p.adjacentPrevious()}? IDENTIFIER | {!p.adjacentPrevious()}?) | IDENTIFIER) (AT IDENTIFIER)?;
// A signed weekly snap may omit its magnitude; other signed unit-only forms
// do not acquire this alternative or a synthetic numeric operand.
signedWeeklySpan: (PLUS | MINUS)
    {p.GetTokenStream().LT(1).GetText() == "w" || p.GetTokenStream().LT(1).GetText() == "week" || p.GetTokenStream().LT(1).GetText() == "weeks"}?
    IDENTIFIER AT IDENTIFIER;
logarithmicSpan: signedNumber? LOG_SPAN;
extendedOption: {p.extendedUnknownOption()}? unknownOption;
signedNumber: (PLUS | MINUS)? NUMBER;
rexCommand: REX rexOption* (REGEX | stringLiteral | RAW_STRING);
rexOption: FIELD ASSIGN identifier | MAX_MATCH ASSIGN integerValue | OFFSET_FIELD ASSIGN identifier | MODE ASSIGN SED | extendedOption;
spathCommand: SPATH (spathOption | extendedOption)*;
spathOption: INPUT ASSIGN identifier | PATH ASSIGN stringLiteral | OUTPUT_LOWER ASSIGN identifier;
loadjobCommand: LOADJOB extendedOption* (NUMBER | identifier | stringLiteral);
metricsCommand: (TSTATS | MSTATS) metricsAggregates metricsOption*;
metricsAggregates: AGGREGATES ASSIGN LBRACKET aggregate (COMMA aggregate)* RBRACKET;
metricsOption: PREDICATE ASSIGN LPAREN expression RPAREN | BYFIELDS ASSIGN LBRACKET groupField (COMMA groupField)* RBRACKET | DATAMODEL_NAME ASSIGN quotedName | extendedOption;
timechartCommand: TIMECHART timechartOption* (EVAL LPAREN expression RPAREN timechartSplit | aggregate timechartExtraAggregate* timechartSplit?);
timechartExtraAggregate: COMMA? aggregate;
timechartOption: binOption | (SEP | FORMAT) ASSIGN stringLiteral | (PARTIAL | CONT | FIXEDRANGE) ASSIGN BOOLEAN
    | LIMIT ASSIGN integerValue | AGG ASSIGN (LPAREN aggregate RPAREN | identifier) | extendedOption;
timechartSplit: BY identifier timechartSplitOption*;
timechartSplitOption: binOption | (USENULL | USEOTHER) ASSIGN BOOLEAN | (NULLSTR | OTHERSTR) ASSIGN stringLiteral | extendedOption;
timewrapCommand: TIMEWRAP timeSpan (ALIGN ASSIGN (NOW | END | identifier))?;
makemvCommand: MAKEMV (DELIM ASSIGN stringLiteral | TOKENIZER ASSIGN (stringLiteral | RAW_STRING) | extendedOption)* identifier;
mvexpandCommand: MVEXPAND extendedOption* (LIMIT ASSIGN integerValue)? identifier;
mvcombineCommand: MVCOMBINE extendedOption* delimOption? identifier;
fillnullCommand: FILLNULL extendedOption* (VALUE ASSIGN stringLiteral)? (identifier (COMMA identifier)*)?;
embeddedCommand: SPL1 stringLiteral | SPL1? embeddedText;

embeddedText: BACKTICK EMBEDDED_TEXT? EMBEDDED_END;
moduleSuffix: statementTerminator (NL* trailingPipelineBoundary)?;
trailingPipelineBoundary: FROM identifier;
moduleDeclaration: annotatedStatement (NL* annotatedStatement)*;
annotatedStatement: annotations NL* annotationStatement | annotations? NL* moduleStatement;
moduleStatement: viewDeclaration | functionDeclaration | importDeclaration | exportDeclaration | unsupportedModuleBoundary;
unsupportedModuleBoundary: unsupportedImportWildcard | unsupportedExportView | unsupportedFunctionTerminator;
unsupportedImportWildcard: IMPORT NL* STAR NL* FROM NL* qualifiedName NL* statementTerminator;
unsupportedExportView: EXPORT NL* LOCAL NL* ASSIGN NL* FROM NL* identifier NL* statementTerminator;
unsupportedFunctionTerminator: FUNCTION NL* identifier NL* LPAREN NL* functionParameters? NL* RPAREN NL*
    LBRACE NL* RETURN NL* expression NL* RBRACE;
viewDeclaration: LOCAL NL* ASSIGN NL* pipeline NL* statementTerminator;
functionDeclaration: FUNCTION NL* identifier NL* LPAREN NL* functionParameters? NL* RPAREN NL*
    LBRACE NL* returnStatement NL* RBRACE;
functionParameters: functionParameter (NL* COMMA NL* functionParameter)*;
functionParameter: LOCAL;
returnStatement: RETURN NL* expression NL* statementTerminator;
importDeclaration: IMPORT NL* importSelection NL* FROM NL* qualifiedName NL* statementTerminator;
importSelection: importWildcard | importList | aliasedImport;
importWildcard: STAR NL* aliasKeyword NL* identifier;
importList: LBRACE NL* aliasedImport (NL* COMMA NL* aliasedImport)* NL* RBRACE;
aliasedImport: identifier (NL* aliasKeyword NL* identifier)?;
exportDeclaration: EXPORT NL* exportSelection NL* statementTerminator;
exportSelection: exportList | aliasedExport;
exportList: LBRACE NL* aliasedExport (NL* COMMA NL* aliasedExport)* NL* RBRACE;
aliasedExport: identifier (NL* aliasKeyword NL* identifier)?;
qualifiedName: identifier (NL* (SLASH | DOT) NL* identifier)*;
annotations: annotation (NL* annotation)*;
annotation: AT identifier (NL* LPAREN NL* arguments? NL* RPAREN)?;
annotationStatement: statementTerminator;
statementTerminator: SEMI;

searchCommand: SEARCH searchExpression;
// The token guard restricts entry, while the entire search owns precedence.
implicitSearch: {p.GetTokenStream().LA(1) == SPL2ParserINDEX && p.GetTokenStream().LA(2) == SPL2ParserASSIGN}? searchExpression;
searchExpression: searchXor;
searchXor: searchAnd (XOR searchAnd)*;
searchAnd: searchOr (AND? searchOr)*;
searchOr: searchNot (OR searchNot)*;
searchNot: NOT searchNot | searchAtom;
searchAtom: LPAREN searchExpression RPAREN | searchTimeModifier | identifier comparison searchValue | identifier IN LPAREN searchValue (COMMA searchValue)* RPAREN | searchValue;
searchValue: searchDirective | searchBareValue | searchWordLiteral | searchUnprovedLiteral | searchSignedNumber | stringLiteral | RAW_STRING;
// Boolean spellings here are literal search words, not expression evaluation.
searchWordLiteral: BOOLEAN;
// Signed values have narrow literal ownership and remain explicitly unproved.
searchSignedNumber: (PLUS | MINUS) NUMBER;
// These bounded punctuation-bearing literal shapes have no field/arithmetic
// interpretation. Recognition retains incomplete syntax coverage.
searchUnprovedLiteral: (PLUS | MINUS) (identifier | NUMBER (PLUS | MINUS) identifier)
    | {p.GetTokenStream().LA(2) != SPL2ParserNUMBER}? MINUS;
searchBareValue: (identifier | NUMBER) ((DOT | MINUS | SLASH | COLON) (identifier | NUMBER) | STAR)* | STAR;
searchDirective: (TERM | CASE) LPAREN searchBareValue RPAREN;
searchTimeModifier: timeModifierKey comparison timeModifierValue | TIMEFORMAT ASSIGN stringLiteral;
timeModifierKey: EARLIEST | LATEST | INDEX_EARLIEST | INDEX_LATEST | STARTTIME | ENDTIME;
timeModifierValue: relativeTime | NUMBER | stringLiteral | NOW LPAREN RPAREN;
relativeTime: (PLUS | MINUS) NUMBER? IDENTIFIER (AT IDENTIFIER)? ((PLUS | MINUS) NUMBER? IDENTIFIER)? | AT IDENTIFIER ((PLUS | MINUS) NUMBER? IDENTIFIER)?;

expression: lambdaExpression | xorExpression;
xorExpression: orExpression (NL* logicalXor NL* orExpression)*;
orExpression: andExpression (NL* logicalOr NL* andExpression)*;
andExpression: notExpression (NL* logicalAnd NL* notExpression)*;
// Prefer a complete ordinary identifier expression when a contextual spelling
// also permits prefix NOT. The reserved uppercase token remains unambiguous.
notExpression: predicate | logicalNot NL* notExpression;
predicate: additive (comparison additive | logicalNot? betweenOperator additive betweenConjunction additive | logicalNot? IN NL* LPAREN NL* expression (NL* COMMA NL* expression)* NL* RPAREN | logicalNot? LIKE additive | IS (logicalNot? (NULL | NULL_TEST) | logicalNot? TYPE))?;
logicalAnd: AND | {p.contextualKeyword("and")}? IDENTIFIER;
logicalOr: OR | {p.contextualKeyword("or")}? IDENTIFIER;
logicalXor: XOR | {p.contextualKeyword("xor")}? IDENTIFIER;
logicalNot: NOT | {p.contextualKeyword("not")}? IDENTIFIER;
betweenOperator: BETWEEN | {p.contextualKeyword("between")}? IDENTIFIER;
betweenConjunction: AND | {p.contextualKeyword("and")}? IDENTIFIER;
comparison: ASSIGN | EQ | NE | LT | LE | GT | GE;
additive: multiplicative ((PLUS | MINUS) multiplicative)*;
multiplicative: unary ((STAR | SLASH | MOD) NL* unary)*;
unary: (PLUS | MINUS) unary | access;
access: primary accessPart*;
accessPart: DOT identifier | LBRACKET expression RBRACKET;
primary: existsPredicate | call | fieldName | LOCAL | literal | array | object | LPAREN NL* expression NL* RPAREN | searchLiteral;
multilineOperator: PLUS | MINUS | STAR | SLASH | MOD | comparison | logicalAnd | logicalOr | logicalXor;
multilineOperand: multilineSimpleCall | multilineAtom multilineAccessPart*;
multilineSimpleCall: identifier LPAREN (multilineOperand (COMMA multilineOperand)*)? RPAREN;
multilineAtom: fieldName | literal | LOCAL;
multilineAccessPart: DOT identifier;
call: identifier LPAREN NL* arguments? NL* RPAREN;
multilineCall: {p.GetTokenStream().LA(3) == SPL2ParserNL}? identifier LPAREN NL+ multilineArguments? NL* RPAREN;
multilineArguments: multilineArgument (NL* COMMA NL* multilineArgument)*;
multilineArgument: identifier COLON multilineOperand | multilineOperand;
arguments: namedArgument (NL* COMMA NL* namedArgument)* | expression (NL* COMMA NL* expression)* (NL* COMMA NL* namedArgument)*;
namedArgument: identifier COLON expression;
literal: NUMBER | BOOLEAN | NULL | RAW_STRING | stringLiteral;
stringLiteral: DQUOTE (STRING_TEXT | STRING_DOLLAR | STRING_INTERPOLATION expression RBRACE)* STRING_END;
quotedName: SQUOTE (NAME_TEXT | NAME_DOLLAR)+ NAME_END;
fieldTemplate: SQUOTE (NAME_TEXT | NAME_DOLLAR)* NAME_INTERPOLATION expression RBRACE (NAME_TEXT | NAME_DOLLAR | NAME_INTERPOLATION expression RBRACE)* NAME_END;
fieldName: identifier | fieldTemplate;
identifier: IDENTIFIER | INDEX | quotedName | pipelineKeyword | {p.outsideSQL()}? sqlKeyword;
sqlKeyword: DISTINCT | HAVING | GROUPBY | ORDER | ORDERBY | LIMIT | OFFSET | ASC | DESC | BY_LOWER
    | JOIN | INNER | LEFT | OUTER | ON | EXISTS;
// New command/option tokens remain ordinary names in expression/name positions.
pipelineKeyword: AS_LOWER | RENAME | STATS | EVENTSTATS | STREAMSTATS | LOOKUP | SORT | DEDUP | HEAD | REVERSE
    | BY | OUTPUT | OUTPUTNEW | ALLNUM | DELIM | PARTITIONS | SPAN | CURRENT | RESET | BEFORE | AFTER | ONCHANGE | WINDOW
    | KEEPEMPTY | CONSECUTIVE | KEEPLAST | WHILE | AUTO | IP | NUM | STR | TERM | CASE
    | EARLIEST | LATEST | INDEX_EARLIEST | INDEX_LATEST | TIMEFORMAT | STARTTIME | ENDTIME | NOW
    | APPEND | APPENDPIPE | APPENDCOLS | UNION | IF | ELSEIF | ELSE | BIN | SPATH | LOADJOB | TSTATS | MSTATS | TIMECHART | TIMEWRAP | MAKEMV | MVEXPAND | MVCOMBINE | FILLNULL | RIGHT | RUN_IN_PREVIEW | MAX | BINS | MINSPAN | START | END | ALIGNTIME | FIELD | MAX_MATCH | OFFSET_FIELD | MODE | SED | INPUT | PATH | OUTPUT_LOWER | AGGREGATES | PREDICATE | BYFIELDS | DATAMODEL_NAME | SEP | FORMAT | PARTIAL | CONT | FIXEDRANGE | AGG | USENULL | USEOTHER | NULLSTR | OTHERSTR | ALIGN | TOKENIZER | VALUE | TYPE_OPTION | LOG_SPAN;
array: LBRACKET (expression (COMMA expression)* COMMA?)? RBRACKET;
object: LBRACE (objectEntry (COMMA objectEntry)* COMMA?)? RBRACE;
objectEntry: objectKey COLON expression;
objectKey: identifier | stringLiteral;
searchLiteral: embeddedText;
lambdaExpression: (lambdaParameter | LPAREN (lambdaParameter (COMMA lambdaParameter)*)? RPAREN) ARROW NL* (expression | lambdaBlock);
lambdaParameter: LOCAL (COLON TYPE)? (ASSIGN ((PLUS | MINUS)? NUMBER | stringLiteral))?;
lambdaBlock: LBRACE NL* (LOCAL ASSIGN expression (SEMI NL* | NL+))* RETURN expression SEMI? NL* RBRACE;
