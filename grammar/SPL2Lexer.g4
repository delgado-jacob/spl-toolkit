lexer grammar SPL2Lexer;

// The second reserved discriminator keeps downstream token numbering stable.
tokens { SELECTED_BY, SELECTED_SPAN_START }

// Every delimiter stack belongs to this lexer instance. Interpolation returns to
// its owning quote mode; object/lambda braces nest without closing that quote.
@structmembers {
braceDepth int
inStats bool
inEval bool
parenDepth int
evalBraceDepth int
statsGroupClassified bool
bracketRoles []spl2BracketRole
buildingProbeTokens bool
probeContext *spl2ProbeContext
}
@members {
const statsGroupProbeLimit = 4096

type spl2ProbeBudget struct {
    units int
    truncated bool
}
type spl2StatsGroupClassification struct {
    classified bool
    selected bool
}
type spl2ProbeErrorRange struct {
    start int
    end int
}
type spl2ProbeContext struct {
    budget *spl2ProbeBudget
    errors []spl2ProbeErrorRange
    tokens []antlr.Token
    positions map[int]int
    brackets map[int]int
    bracketRoles map[int]bool
    bracketActive map[int]bool
    statsGroups map[int]spl2StatsGroupClassification
}
type spl2BracketRole struct {
    start int
    classified bool
    command bool
}

type statsGroupProbeLimitAbort struct{}
var statsGroupProbeLimitAbortSignal = &statsGroupProbeLimitAbort{}

type statsGroupProbeErrors struct {
    *antlr.DefaultErrorListener
    failed bool
    source *statsGroupProbeSource
    ranges *[]spl2ProbeErrorRange
}
func (e *statsGroupProbeErrors) SyntaxError(antlr.Recognizer, interface{}, int, int, string, antlr.RecognitionException) {
    e.failed = true
    if e.source != nil && e.ranges != nil {
        start := e.source.GetInputStream().Index()
        *e.ranges = append(*e.ranges, spl2ProbeErrorRange{start: start, end: start+1})
    }
    if e.source == nil { return }
    if e.source.budget.units >= statsGroupProbeLimit {
        e.source.budget.truncated = true
        panic(statsGroupProbeLimitAbortSignal)
    }
    e.source.budget.units++
}

type statsGroupProbeSource struct {
    *SPL2Lexer
    budget *spl2ProbeBudget
    depth int
    complete bool
    wholeDocument bool
}
func (s *statsGroupProbeSource) eof() antlr.Token {
    offset := s.GetInputStream().Index()
    token := antlr.NewCommonToken(&antlr.TokenSourceCharStreamPair{}, antlr.TokenEOF, antlr.TokenDefaultChannel, offset, offset-1)
    token.SetText("<EOF>")
    return token
}
func (s *statsGroupProbeSource) NextToken() (token antlr.Token) {
    defer func() {
        recovered := recover()
        if recovered == nil { return }
        if recovered != statsGroupProbeLimitAbortSignal { panic(recovered) }
        token = s.eof()
    }()
    if s.complete || s.budget.truncated { return s.eof() }
    token = s.SPL2Lexer.NextToken()
    if token.GetTokenType() == antlr.TokenEOF {
        s.complete = true
        return token
    }
    if s.budget.units >= statsGroupProbeLimit {
        s.budget.truncated = true
        return s.eof()
    }
    s.budget.units++
    switch token.GetTokenType() {
    case SPL2LexerLPAREN, SPL2LexerLBRACE:
        s.depth++
    case SPL2LexerLBRACKET:
        s.depth++
    case SPL2LexerRPAREN, SPL2LexerRBRACE:
        if s.depth > 0 { s.depth-- }
    case SPL2LexerRBRACKET:
        if s.depth > 0 { s.depth-- } else if !s.wholeDocument { s.complete = true; return s.eof() }
    case SPL2LexerPIPE, SPL2LexerSEMI, SPL2LexerNL:
        if !s.wholeDocument && s.depth == 0 { s.complete = true; return s.eof() }
    }
    return token
}

func (l *SPL2Lexer) openBrace() { l.braceDepth++; l.PushMode(antlr.LexerDefaultMode) }
func (l *SPL2Lexer) closeBrace() { if l.braceDepth > 0 { l.braceDepth--; l.PopMode() } }
func (l *SPL2Lexer) openBracket() {
    l.bracketRoles = append(l.bracketRoles, spl2BracketRole{start: l.TokenStartCharIndex})
}
func (l *SPL2Lexer) closeBracket() {
    if len(l.bracketRoles) > 0 { l.bracketRoles = l.bracketRoles[:len(l.bracketRoles)-1] }
}
func (l *SPL2Lexer) previousNonspace(start int) int {
    input := l.GetInputStream()
    for index := start - 1; index >= 0; index-- {
        text := input.GetText(index, index)
        if text != " " && text != "\t" { return index }
    }
    return -1
}
func (l *SPL2Lexer) characterCommandStart() bool {
    index := l.previousNonspace(l.TokenStartCharIndex)
    if index < 0 { return true }
    switch l.GetInputStream().GetText(index, index) {
    case "|", "[", "\n", "\r":
        return true
    }
    return false
}
func (l *SPL2Lexer) statsCommandStart() bool {
    if !l.buildingProbeTokens && len(l.bracketRoles) > 0 {
        index := len(l.bracketRoles)-1
        role := &l.bracketRoles[index]
        if !role.classified {
            role.command = l.bracketStartsCommand(role.start)
            role.classified = true
        }
        if !role.command { return false }
    }
    return l.characterCommandStart()
}
func (l *SPL2Lexer) startStats() {
    if l.statsCommandStart() {
        l.inStats = true
        l.statsGroupClassified = false
    }
}
type spl2CachedTokenSource struct {
    *antlr.BaseLexer
    tokens []antlr.Token
    index int
}
func newSPL2CachedTokenSource(tokens []antlr.Token) *spl2CachedTokenSource {
    return &spl2CachedTokenSource{BaseLexer: antlr.NewBaseLexer(antlr.NewInputStream("")), tokens: tokens}
}
func (s *spl2CachedTokenSource) NextToken() antlr.Token {
    if s.index < len(s.tokens) {
        token := s.tokens[s.index]
        s.index++
        return token
    }
    token := antlr.NewCommonToken(&antlr.TokenSourceCharStreamPair{}, antlr.TokenEOF, antlr.TokenDefaultChannel, 0, -1)
    token.SetText("<EOF>")
    return token
}
func spl2CloneProbeToken(token antlr.Token, tokenType int) antlr.Token {
    clone := antlr.NewCommonToken(&antlr.TokenSourceCharStreamPair{}, tokenType, token.GetChannel(), token.GetStart(), token.GetStop())
    clone.SetText(token.GetText())
    return clone
}
func (c *spl2ProbeContext) stream(start, end int, retags map[int]int) *antlr.CommonTokenStream {
    tokens := make([]antlr.Token, 0, end-start+1)
    for _, token := range c.tokens[start:end] {
        if token.GetTokenType() == antlr.TokenEOF { break }
        tokenType := token.GetTokenType()
        if retag, ok := retags[token.GetStart()]; ok { tokenType = retag }
        tokens = append(tokens, spl2CloneProbeToken(token, tokenType))
    }
    offset := 0
    if len(tokens) > 0 { offset = tokens[len(tokens)-1].GetStop()+1 }
    eof := antlr.NewCommonToken(&antlr.TokenSourceCharStreamPair{}, antlr.TokenEOF, antlr.TokenDefaultChannel, offset, offset-1)
    eof.SetText("<EOF>")
    tokens = append(tokens, eof)
    return antlr.NewCommonTokenStream(newSPL2CachedTokenSource(tokens), antlr.TokenDefaultChannel)
}
func (l *SPL2Lexer) documentProbeContext() *spl2ProbeContext {
    if l.probeContext != nil { return l.probeContext }
    context := &spl2ProbeContext{
        budget: &spl2ProbeBudget{},
        positions: map[int]int{},
        brackets: map[int]int{},
        bracketRoles: map[int]bool{},
        bracketActive: map[int]bool{},
        statsGroups: map[int]spl2StatsGroupClassification{},
    }
    l.probeContext = context
    input := l.GetInputStream()
    saved := input.Index()
    input.Seek(0)
    defer input.Seek(saved)
    lexer := NewSPL2Lexer(input)
    lexer.buildingProbeTokens = true
    source := &statsGroupProbeSource{SPL2Lexer: lexer, budget: context.budget, wholeDocument: true}
    errors := &statsGroupProbeErrors{DefaultErrorListener: antlr.NewDefaultErrorListener(), source: source, ranges: &context.errors}
    lexer.RemoveErrorListeners()
    lexer.AddErrorListener(errors)
    tokens := antlr.NewCommonTokenStream(source, antlr.TokenDefaultChannel)
    tokens.Fill()
    context.tokens = append(context.tokens, tokens.GetAllTokens()...)
    stack := []int{}
    for index, token := range context.tokens {
        if _, exists := context.positions[token.GetStart()]; !exists { context.positions[token.GetStart()] = index }
        switch token.GetTokenType() {
        case SPL2LexerLBRACKET:
            stack = append(stack, token.GetStart())
        case SPL2LexerRBRACKET:
            if len(stack) > 0 {
                start := stack[len(stack)-1]
                stack = stack[:len(stack)-1]
                context.brackets[start] = index
            }
        }
    }
    return context
}
func (c *spl2ProbeContext) hasLexicalError(start, end int) bool {
    for _, damaged := range c.errors {
        if damaged.start < end && damaged.end > start { return true }
    }
    return false
}
func (l *SPL2Lexer) bracketCommandProbe(tokens *antlr.CommonTokenStream, independent bool) bool {
    tokens.Seek(0)
    errors := &statsGroupProbeErrors{DefaultErrorListener: antlr.NewDefaultErrorListener()}
    parser := NewSPL2Parser(tokens)
    parser.RemoveErrorListeners()
    parser.AddErrorListener(errors)
    parser.BuildParseTrees = false
    if independent { parser.IndependentSearch() } else { parser.InheritedSubpipe() }
    return !errors.failed && tokens.LA(1) == antlr.TokenEOF
}
func (c *spl2ProbeContext) statsTailRange(byStart int) (int, int, bool) {
    by, ok := c.positions[byStart]
    if !ok { return 0, 0, false }
    depth := 0
    for index := by+1; index < len(c.tokens); index++ {
        token := c.tokens[index].GetTokenType()
        switch token {
        case antlr.TokenEOF:
            return by+1, index, true
        case SPL2LexerLPAREN, SPL2LexerLBRACKET, SPL2LexerLBRACE:
            depth++
        case SPL2LexerRPAREN, SPL2LexerRBRACE:
            if depth > 0 { depth-- }
        case SPL2LexerRBRACKET:
            if depth == 0 { return by+1, index, true }
            depth--
        case SPL2LexerPIPE, SPL2LexerSEMI, SPL2LexerNL:
            if depth == 0 { return by+1, index, true }
        }
    }
    return 0, 0, false
}
func (l *SPL2Lexer) statsGroupProbe(tokens *antlr.CommonTokenStream, selected bool) bool {
    tokens.Seek(0)
    errors := &statsGroupProbeErrors{DefaultErrorListener: antlr.NewDefaultErrorListener()}
    parser := NewSPL2Parser(tokens)
    parser.RemoveErrorListeners()
    parser.AddErrorListener(errors)
    parser.BuildParseTrees = false
    parseTerm := func() {
        if selected { parser.SelectedGroupTerm() } else { parser.GroupField() }
    }
    parseTerm()
    for !errors.failed && tokens.LA(1) == SPL2ParserCOMMA {
        parser.Consume()
        parseTerm()
    }
    return !errors.failed && tokens.LA(1) == antlr.TokenEOF
}
func (l *SPL2Lexer) classifyStatsGroup(context *spl2ProbeContext, byStart int) spl2StatsGroupClassification {
    if classification, ok := context.statsGroups[byStart]; ok { return classification }
    classification := spl2StatsGroupClassification{}
    if context.budget.truncated { return classification }
    start, end, ok := context.statsTailRange(byStart)
    if !ok { return classification }
    by := context.positions[byStart]
    if context.hasLexicalError(context.tokens[by].GetStop()+1, context.tokens[end].GetStart()) { return classification }
    tokens := context.stream(start, end, nil)
    if l.statsGroupProbe(tokens, false) {
        classification.classified = true
    } else if l.statsGroupProbe(tokens, true) {
        classification.classified = true
        classification.selected = true
    }
    context.statsGroups[byStart] = classification
    return classification
}
func (l *SPL2Lexer) statsGroupBY(context *spl2ProbeContext, stats, end int) (int, bool) {
    depth := 0
    for index := stats+1; index < end; index++ {
        token := context.tokens[index]
        switch token.GetTokenType() {
        case SPL2LexerLPAREN, SPL2LexerLBRACKET, SPL2LexerLBRACE:
            depth++
        case SPL2LexerRPAREN, SPL2LexerRBRACKET, SPL2LexerRBRACE:
            if depth > 0 { depth-- }
        case SPL2LexerPIPE:
            if depth == 0 { return 0, false }
        case SPL2LexerBY:
            if depth == 0 {
                classification := l.classifyStatsGroup(context, token.GetStart())
                if classification.classified { return token.GetStart(), classification.selected }
            }
        }
    }
    return 0, false
}
func (l *SPL2Lexer) collectBracketRetags(context *spl2ProbeContext, start int, retags map[int]int) {
    first, ok := context.positions[start]
    end, closed := context.brackets[start]
    if !ok || !closed { return }
    commandStart := true
    parens, braces := 0, 0
    for index := first+1; index < end; index++ {
        token := context.tokens[index]
        tokenType := token.GetTokenType()
        if tokenType == SPL2LexerLBRACKET {
            nestedStart := token.GetStart()
            if l.classifyBracketRole(context, nestedStart) { l.collectBracketRetags(context, nestedStart, retags) }
            if nestedEnd, ok := context.brackets[nestedStart]; ok { index = nestedEnd }
            commandStart = false
            continue
        }
        switch tokenType {
        case SPL2LexerLPAREN:
            parens++
        case SPL2LexerRPAREN:
            if parens > 0 { parens-- }
        case SPL2LexerLBRACE:
            braces++
        case SPL2LexerRBRACE:
            if braces > 0 { braces-- }
        case SPL2LexerPIPE:
            if parens == 0 && braces == 0 { commandStart = true }
        case SPL2LexerNL:
        default:
            if token.GetChannel() != antlr.TokenDefaultChannel { continue }
            if commandStart {
                if tokenType == SPL2LexerSTATS {
                    if by, selected := l.statsGroupBY(context, index, end); selected { retags[by] = SPL2LexerSELECTED_BY }
                }
                commandStart = false
            }
        }
    }
}
func (l *SPL2Lexer) classifyBracketRole(context *spl2ProbeContext, start int) bool {
    if context.budget.truncated { return false }
    if command, ok := context.bracketRoles[start]; ok { return command }
    if context.bracketActive[start] { return false }
    first, ok := context.positions[start]
    end, closed := context.brackets[start]
    if !ok || !closed { context.bracketRoles[start] = false; return false }
    if context.hasLexicalError(start, context.tokens[end].GetStop()+1) { context.bracketRoles[start] = false; return false }
    context.bracketActive[start] = true
    defer delete(context.bracketActive, start)
    retags := map[int]int{}
    l.collectBracketRetags(context, start, retags)
    tokens := context.stream(first, end+1, retags)
    command := l.bracketCommandProbe(tokens, false) || l.bracketCommandProbe(tokens, true)
    context.bracketRoles[start] = command
    return command
}
func (l *SPL2Lexer) bracketStartsCommand(start int) bool {
    return l.classifyBracketRole(l.documentProbeContext(), start)
}
func (l *SPL2Lexer) classifyStatsGroupTail() (bool, bool) {
    classification := l.classifyStatsGroup(l.documentProbeContext(), l.TokenStartCharIndex)
    return classification.classified, classification.selected
}
func (l *SPL2Lexer) selectStatsBy() {
    if l.buildingProbeTokens || !l.inStats || l.statsGroupClassified { return }
    classified, selected := l.classifyStatsGroupTail()
    if !classified { return }
    l.statsGroupClassified = true
    if selected { l.SetType(SPL2LexerSELECTED_BY) }
}
}

FROM: 'FROM' | 'from'; SELECT: 'SELECT' | 'select';
DISTINCT: 'DISTINCT' | 'distinct'; HAVING: 'HAVING' | 'having';
GROUPBY: 'GROUPBY' | 'groupby'; ORDER: 'ORDER' | 'order'; ORDERBY: 'ORDERBY' | 'orderby';
LIMIT: 'LIMIT' | 'limit'; OFFSET: 'OFFSET' | 'offset';
ASC: 'ASC' | 'asc'; DESC: 'DESC' | 'desc'; BY_LOWER: 'by';
JOIN: 'JOIN' | 'join'; INNER: 'INNER' | 'inner'; LEFT: 'LEFT' | 'left'; OUTER: 'OUTER' | 'outer';
ON: 'ON' | 'on'; EXISTS: 'EXISTS' | 'exists';
SEARCH: 'search'; INDEX: 'index'; EVAL: 'eval' {l.inEval = l.statsCommandStart(); if l.inEval {l.parenDepth=0; l.evalBraceDepth=0}}; WHERE: 'where' | 'WHERE';
FIELDS: 'fields'; TABLE: 'table'; AS: 'AS'; AS_LOWER: 'as'; GROUP: 'GROUP' | 'group';
RENAME: 'rename'; STATS: 'stats' {l.startStats()}; EVENTSTATS: 'eventstats'; STREAMSTATS: 'streamstats';
LOOKUP: 'lookup'; SORT: 'sort'; DEDUP: 'dedup'; HEAD: 'head'; REVERSE: 'reverse';
BY: 'BY' {l.selectStatsBy()}; OUTPUT: 'OUTPUT'; OUTPUTNEW: 'OUTPUTNEW';
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
SLASH: '/'; MOD: '%'; PIPE: '|' {l.inStats=false; l.statsGroupClassified=false; l.inEval=false; l.parenDepth=0; l.evalBraceDepth=0}; COMMA: ','; COLON: ':'; DOT: '.';
LPAREN: '(' {l.parenDepth++}; RPAREN: ')' {if l.parenDepth > 0 {l.parenDepth--}}; LBRACKET: '[' {l.openBracket()}; RBRACKET: ']' {l.closeBracket()};
LBRACE: '{' {l.openBrace(); if l.inEval {l.evalBraceDepth++}};
RBRACE: '}' {l.closeBrace(); if l.inEval && l.evalBraceDepth > 0 {l.evalBraceDepth--}};
SEMI: ';' {if l.evalBraceDepth == 0 {l.inStats=false; l.statsGroupClassified=false; l.inEval=false; l.parenDepth=0}};
RAW_STRING: '@"' ('""' | ~'"')* '"';
DQUOTE: '"' -> pushMode(STRING_MODE);
SQUOTE: '\'' -> pushMode(NAME_MODE);
BACKTICK: '`' -> pushMode(EMBEDDED_MODE);
LOG_SPAN: 'log' ([0-9]+ ('.' [0-9]+)?)?;
NUMBER: [0-9]+ ('.' [0-9]+)? ([eE] [+-]? [0-9]+)? [LFD]?;
LOCAL: '$' [a-zA-Z_] [a-zA-Z_0-9]*;
IDENTIFIER: [a-zA-Z_] [a-zA-Z_0-9]*;
NL: ('\r'? '\n' | '\r') {if l.inEval && l.parenDepth > 0 && l.evalBraceDepth == 0 {l.SetChannel(antlr.TokenHiddenChannel)}};
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
RX_CLOSE: ']' {l.closeBracket()} -> type(RBRACKET), popMode;
RX_EQ: '=' -> type(ASSIGN);
RX_PIPE: '|' -> type(PIPE), popMode;
RX_WS: [ \t\r\n]+ -> channel(HIDDEN);
