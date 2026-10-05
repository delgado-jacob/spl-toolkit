package analysis

import "github.com/antlr4-go/antlr/v4"

const lexerWorkLimit = 4096

// Semantic source expansion uses the same bounded work scale as lexical input.
// Each new active-source or discovery fact consumes one unit before copying.
const sourceEvidenceWorkLimit = lexerWorkLimit

const sourceEvidenceResourceLimitMessage = "source discovery stopped because further copying would exceed the 4,096-unit source-evidence work limit"

type sourceEvidenceWorkBudget struct {
	units   int
	failure *Diagnostic
}

func (t *requirementTrace) reserveSourceEvidence(units int, stage Stage, location Location) bool {
	if t == nil || units == 0 {
		return true
	}
	budget := t.sourceEvidenceBudget
	if budget == nil {
		budget = &sourceEvidenceWorkBudget{}
		t.sourceEvidenceBudget = budget
	}
	if budget.failure == nil && units <= sourceEvidenceWorkLimit-budget.units {
		budget.units += units
		return true
	}
	if budget.failure == nil {
		budget.failure = &Diagnostic{Code: CodeAnalysisResourceLimit, Severity: "warning", Category: "resource_limit", Message: sourceEvidenceResourceLimitMessage, Location: location, StageID: stage.ID, ScopeID: stage.ScopeID}
	}
	return false
}

// Publish once from the canonical owner after branch merges and ID finalization.
// The shared budget never changes immutable branch reference/diagnostic prefixes.
func (t *requirementTrace) finishSourceEvidence(result *Result) {
	if t == nil || t.sourceEvidenceBudget == nil || t.sourceEvidenceBudget.failure == nil {
		return
	}
	diagnostic := *t.sourceEvidenceBudget.failure
	result.Diagnostics = append(result.Diagnostics, diagnostic)
	result.Coverage.SemanticComplete = false
	t.recordDiagnostic(diagnostic, true, nil, t.nextEvent())
}

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
	result.Inputs = cloneInputs(requirements.Inputs)
	result.InputCoverage = cloneInputCoverage(requirements.InputCoverage)
	result.FieldAttributionCoverage = cloneInputCoverage(requirements.FieldAttributionCoverage)
	result.Correlation = cloneCorrelation(requirements.Correlation)
	return result, trace, nil
}
