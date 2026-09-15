package analysis

import "github.com/antlr4-go/antlr/v4"

const lexerWorkLimit = 4096

const (
	analysisResourceLimitMessage    = "analysis stopped before parser prediction after reaching the 4,096-unit lexer work limit"
	requirementResourceLimitMessage = "requirement coverage is incomplete because analysis exceeded the 4,096-unit lexer work limit"
)

type lexerWorkTracker struct {
	units         int
	resourceLimit *Location
}

type lexerWorkLimitAbort struct{ marker byte }

var lexerWorkLimitAbortSignal = &lexerWorkLimitAbort{marker: 1}

func (t *lexerWorkTracker) consume(location Location) bool {
	if t.resourceLimit != nil {
		return false
	}
	if t.units == lexerWorkLimit {
		locationCopy := location
		t.resourceLimit = &locationCopy
		return false
	}
	t.units++
	return true
}

func (t *lexerWorkTracker) consumeLexerError(location Location) {
	if !t.consume(location) {
		panic(lexerWorkLimitAbortSignal)
	}
}

type lexerWorkLimiter struct {
	antlr.Lexer
	tracker *lexerWorkTracker
	source  *sourceIndex
}

func (l *lexerWorkLimiter) NextToken() (token antlr.Token) {
	defer func() {
		recovered := recover()
		if recovered == nil {
			return
		}
		if recovered != lexerWorkLimitAbortSignal || l.tracker == nil || l.tracker.resourceLimit == nil {
			panic(recovered)
		}
		token = resourceLimitEOF(l.Lexer.GetInputStream().Index())
	}()
	token = l.Lexer.NextToken()
	if l.tracker.resourceLimit != nil {
		return resourceLimitEOF(token.GetStart())
	}
	if token.GetTokenType() == antlr.TokenEOF {
		return token
	}
	if !l.tracker.consume(l.source.location(token.GetStart(), token.GetStop()+1)) {
		return resourceLimitEOF(token.GetStart())
	}
	return token
}

func resourceLimitEOF(offset int) antlr.Token {
	token := antlr.NewCommonToken(&antlr.TokenSourceCharStreamPair{}, antlr.TokenEOF, antlr.TokenDefaultChannel, offset, offset-1)
	token.SetText("<EOF>")
	return token
}

func preflightLexer(lexer antlr.Lexer, tracker *lexerWorkTracker, source *sourceIndex) *antlr.CommonTokenStream {
	tokens := antlr.NewCommonTokenStream(&lexerWorkLimiter{Lexer: lexer, tracker: tracker, source: source}, antlr.TokenDefaultChannel)
	tokens.Fill()
	return tokens
}

func resourceLimitedAnalysis(document QueryDocument, location Location) (*Result, *requirementTrace, error) {
	diagnostic := Diagnostic{
		Code:     CodeAnalysisResourceLimit,
		Severity: "warning",
		Category: "resource_limit",
		Message:  analysisResourceLimitMessage,
		Location: location,
	}
	result := newResult(document)
	result.Coverage.SyntaxComplete = false
	result.Coverage.SemanticComplete = false
	result.Diagnostics = append(result.Diagnostics, diagnostic)
	finalizeResult(result)

	trace := newRequirementTrace()
	trace.syntaxComplete = false
	trace.recordDiagnostic(diagnostic, true, nil, trace.nextEvent())
	requirements, err := projectRequirements(document, trace)
	if err != nil {
		return nil, nil, err
	}
	result.Requirements = requirements
	return result, trace, nil
}
