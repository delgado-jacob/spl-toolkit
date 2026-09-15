package analysis

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/antlr4-go/antlr/v4"
	"github.com/delgado-jacob/spl-toolkit/parser"
	"github.com/delgado-jacob/spl-toolkit/parser/spl2"
)

func TestSPLLexerWorkBudgetExactAndPlusOne(t *testing.T) {
	exact, plusOne := lexerBoundaryQueries(t, "spl")

	parsed := parseDocument(exact)
	if parsed.resourceLimit != nil || parsed.tree == nil {
		t.Fatalf("exact-limit SPL was not admitted: %+v", parsed.resourceLimit)
	}
	assertNotResourceLimited(t, mustAnalyzeDocument(t, QueryDocument{Text: exact}))
	result := mustAnalyzeDocument(t, QueryDocument{Text: plusOne})
	assertResourceLimitedResult(t, result, plusOne, rawLexerTokens(t, "spl", plusOne)[lexerWorkLimit])
}

func TestSPL2LexerWorkBudgetExactAndPlusOne(t *testing.T) {
	exact, plusOne := lexerBoundaryQueries(t, "spl2")

	parsed := parseSPL2Document(exact)
	if parsed.resourceLimit != nil || parsed.tree == nil {
		t.Fatalf("exact-limit SPL2 was not admitted: %+v", parsed.resourceLimit)
	}
	assertNotResourceLimited(t, mustAnalyzeDocument(t, QueryDocument{Text: exact, Language: "spl2"}))
	result := mustAnalyzeDocument(t, QueryDocument{Text: plusOne, Language: "spl2"})
	assertResourceLimitedResult(t, result, plusOne, rawLexerTokens(t, "spl2", plusOne)[lexerWorkLimit])
}

func TestSPLParserFactoryRunsOnlyAfterLexerAdmission(t *testing.T) {
	exact, plusOne := lexerBoundaryQueries(t, "spl")
	calls := 0
	factory := func(tokens antlr.TokenStream) *parser.SPLParser {
		calls++
		return parser.NewSPLParser(tokens)
	}

	parsed := parseDocumentWithParserFactory(exact, factory)
	if parsed.resourceLimit != nil || parsed.tree == nil || calls != 1 {
		t.Fatalf("exact-limit parse = resource %v tree %T factory calls %d, want admitted tree and one call", parsed.resourceLimit, parsed.tree, calls)
	}
	calls = 0
	parsed = parseDocumentWithParserFactory(plusOne, factory)
	if parsed.resourceLimit == nil || parsed.tree != nil || calls != 0 {
		t.Fatalf("plus-one parse = resource %v tree %T factory calls %d, want resource limit and zero calls", parsed.resourceLimit, parsed.tree, calls)
	}
}

func TestSPL2ParserFactoryRunsOnlyAfterLexerAdmission(t *testing.T) {
	exact, plusOne := lexerBoundaryQueries(t, "spl2")
	calls := 0
	factory := func(tokens antlr.TokenStream) *spl2.SPL2Parser {
		calls++
		return spl2.NewSPL2Parser(tokens)
	}

	parsed := parseSPL2DocumentWithParserFactory(exact, factory)
	if parsed.resourceLimit != nil || parsed.tree == nil || calls != 1 {
		t.Fatalf("exact-limit parse = resource %v tree %T factory calls %d, want admitted tree and one call", parsed.resourceLimit, parsed.tree, calls)
	}
	calls = 0
	parsed = parseSPL2DocumentWithParserFactory(plusOne, factory)
	if parsed.resourceLimit == nil || parsed.tree != nil || calls != 0 {
		t.Fatalf("plus-one parse = resource %v tree %T factory calls %d, want resource limit and zero calls", parsed.resourceLimit, parsed.tree, calls)
	}
}

func TestSPLLexerWorkBudgetCountsErrorsBeforeReturnedToken(t *testing.T) {
	prefix, _ := lexerBoundaryQueriesAt(t, "spl", lexerWorkLimit-1)

	t.Run("error is unit 4096 then token is omitted", func(t *testing.T) {
		query := prefix + "\x00x"
		result := mustAnalyzeDocument(t, QueryDocument{Text: query})
		assertSingleResourceLimitLocation(t, result, newSourceIndex(query).location(len([]rune(prefix))+1, len([]rune(prefix))+2))
	})

	t.Run("error is first omitted event", func(t *testing.T) {
		exact, _ := lexerBoundaryQueriesAt(t, "spl", lexerWorkLimit)
		query := exact + "\x00x\x00"
		result := mustAnalyzeDocument(t, QueryDocument{Text: query})
		start := len([]rune(exact))
		assertSingleResourceLimitLocation(t, result, newSourceIndex(query).location(start, start+1))
	})
}

func TestSPL2LexerWorkBudgetCountsErrorsBeforeReturnedToken(t *testing.T) {
	prefix, _ := lexerBoundaryQueriesAt(t, "spl2", lexerWorkLimit-1)

	t.Run("error is unit 4096 then token is omitted", func(t *testing.T) {
		query := prefix + "\x00x"
		result := mustAnalyzeDocument(t, QueryDocument{Text: query, Language: "spl2"})
		assertSingleResourceLimitLocation(t, result, newSourceIndex(query).location(len([]rune(prefix))+1, len([]rune(prefix))+2))
	})

	t.Run("error is first omitted event", func(t *testing.T) {
		exact, _ := lexerBoundaryQueriesAt(t, "spl2", lexerWorkLimit)
		query := exact + "\x00x\x00"
		result := mustAnalyzeDocument(t, QueryDocument{Text: query, Language: "spl2"})
		start := len([]rune(exact))
		assertSingleResourceLimitLocation(t, result, newSourceIndex(query).location(start, start+1))
	})
}

func TestSPLLexerWorkBudgetAbortsConsecutiveErrorStorm(t *testing.T) {
	prefix, _ := lexerBoundaryQueriesAt(t, "spl", lexerWorkLimit-1)
	query := prefix + strings.Repeat("\x00", 32*1024)
	source := newSourceIndex(query)
	parsed := &parsedDocument{source: source, diagnostics: []Diagnostic{}}
	tracker := &lexerWorkTracker{}
	observer := &countingSyntaxErrorListener{DefaultErrorListener: antlr.NewDefaultErrorListener()}
	listener := &syntaxListener{DefaultErrorListener: antlr.NewDefaultErrorListener(), parsed: parsed, tracker: tracker}
	lexer := parser.NewSPLLexer(&analysisInputStream{InputStream: antlr.NewInputStream(query)})
	lexer.RemoveErrorListeners()
	lexer.AddErrorListener(observer)
	lexer.AddErrorListener(listener)
	preflightLexer(lexer, tracker, source)

	assertErrorStormStoppedAtFirstOmitted(t, query, prefix, tracker, observer.calls, lexer.GetInputStream().Index(), parsed.diagnostics)
	assertSingleResourceLimitLocation(t, mustAnalyzeDocument(t, QueryDocument{Text: query}), *tracker.resourceLimit)
}

func TestSPL2LexerWorkBudgetAbortsConsecutiveErrorStormBeforeClosureInspection(t *testing.T) {
	prefix, _ := lexerBoundaryQueriesAt(t, "spl2", lexerWorkLimit-1)
	query := prefix + strings.Repeat("\x00", 32*1024) + `"`
	source := newSourceIndex(query)
	parsed := &spl2ParsedDocument{source: source, diagnostics: []Diagnostic{}}
	tracker := &lexerWorkTracker{}
	observer := &countingSyntaxErrorListener{DefaultErrorListener: antlr.NewDefaultErrorListener()}
	listener := &spl2SyntaxListener{DefaultErrorListener: antlr.NewDefaultErrorListener(), parsed: parsed, tracker: tracker}
	lexer := spl2.NewSPL2Lexer(antlr.NewInputStream(query))
	lexer.RemoveErrorListeners()
	lexer.AddErrorListener(observer)
	lexer.AddErrorListener(listener)
	parsed.tokens = preflightLexer(lexer, tracker, source)

	assertErrorStormStoppedAtFirstOmitted(t, query, prefix, tracker, observer.calls, lexer.GetInputStream().Index(), parsed.diagnostics)
	if len(parsed.lexicalErrors) != 1 || parsed.literalClosureDiagnostic() != nil {
		t.Fatalf("post-limit SPL2 evidence survived: lexical=%+v closure=%+v", parsed.lexicalErrors, parsed.literalClosureDiagnostic())
	}
	result := mustAnalyzeDocument(t, QueryDocument{Text: query, Language: "spl2"})
	assertSingleResourceLimitLocation(t, result, *tracker.resourceLimit)
}

type countingSyntaxErrorListener struct {
	*antlr.DefaultErrorListener
	calls int
}

func (l *countingSyntaxErrorListener) SyntaxError(antlr.Recognizer, interface{}, int, int, string, antlr.RecognitionException) {
	l.calls++
}

type panickingLexer struct {
	antlr.Lexer
	value any
}

func (l *panickingLexer) NextToken() antlr.Token {
	panic(l.value)
}

func TestLexerWorkLimiterRepanicsUnrelatedValues(t *testing.T) {
	marker := &struct{ value int }{value: 1}
	lexer := parser.NewSPLLexer(antlr.NewInputStream(""))
	limiter := &lexerWorkLimiter{
		Lexer:   &panickingLexer{Lexer: lexer, value: marker},
		tracker: &lexerWorkTracker{},
		source:  newSourceIndex(""),
	}
	defer func() {
		if recovered := recover(); recovered != marker {
			t.Fatalf("recovered panic = %#v, want original marker", recovered)
		}
	}()
	limiter.NextToken()
	t.Fatal("unrelated panic was swallowed")
}

func TestLexerWorkLimiterRepanicsAbortSentinelWithoutRecordedLimit(t *testing.T) {
	lexer := parser.NewSPLLexer(antlr.NewInputStream(""))
	limiter := &lexerWorkLimiter{
		Lexer:   &panickingLexer{Lexer: lexer, value: lexerWorkLimitAbortSignal},
		tracker: &lexerWorkTracker{},
		source:  newSourceIndex(""),
	}
	defer func() {
		if recovered := recover(); recovered != lexerWorkLimitAbortSignal {
			t.Fatalf("recovered panic = %#v, want original lexer work abort", recovered)
		}
	}()
	limiter.NextToken()
	t.Fatal("unowned lexer work abort was swallowed")
}

func assertErrorStormStoppedAtFirstOmitted(t *testing.T, query, prefix string, tracker *lexerWorkTracker, calls, inputIndex int, diagnostics []Diagnostic) {
	t.Helper()
	firstError := len([]rune(prefix))
	omittedError := firstError + 1
	want := newSourceIndex(query).location(omittedError, omittedError+1)
	if tracker.units != lexerWorkLimit || tracker.resourceLimit == nil || *tracker.resourceLimit != want {
		t.Fatalf("resource tracker = units %d location %+v, want units %d location %+v", tracker.units, tracker.resourceLimit, lexerWorkLimit, want)
	}
	if calls != 2 {
		t.Fatalf("lexer inspected %d errors, want the admitted error and first omitted error only", calls)
	}
	if inputIndex != omittedError {
		t.Fatalf("lexer input index = %d, want first omitted error index %d", inputIndex, omittedError)
	}
	if len(diagnostics) != 1 || diagnostics[0].Location != newSourceIndex(query).location(firstError, firstError+1) {
		t.Fatalf("ordinary lexer diagnostics = %+v, want only the admitted error", diagnostics)
	}
}

func TestLexerWorkBudgetEOFErrorUsesClampedRange(t *testing.T) {
	text := "x"
	parsed := &parsedDocument{source: newSourceIndex(text), diagnostics: []Diagnostic{}}
	tracker := &lexerWorkTracker{units: lexerWorkLimit}
	listener := &syntaxListener{DefaultErrorListener: antlr.NewDefaultErrorListener(), parsed: parsed, tracker: tracker}
	lexer := parser.NewSPLLexer(antlr.NewInputStream(text))
	lexer.GetInputStream().Seek(len([]rune(text)))
	func() {
		defer func() {
			if recovered := recover(); recovered != lexerWorkLimitAbortSignal {
				t.Fatalf("recovered panic = %#v, want lexer work abort", recovered)
			}
		}()
		listener.SyntaxError(lexer, nil, 1, 1, "EOF lexer error", nil)
		t.Fatal("rejected EOF lexer error did not abort")
	}()

	want := parsed.source.location(len([]rune(text)), len([]rune(text))+1)
	if tracker.resourceLimit == nil || *tracker.resourceLimit != want || want.Start != want.End {
		t.Fatalf("EOF lexer-error range = %+v, want clamped %+v", tracker.resourceLimit, want)
	}
}

func TestLexerWorkBudgetTokenRangeUsesSourceIndexedBytes(t *testing.T) {
	for _, language := range []string{"spl", "spl2"} {
		t.Run(language, func(t *testing.T) {
			prefix, _ := lexerBoundaryQueriesAt(t, language, lexerWorkLimit-1)
			suffix := "|é"
			wantStart, wantEnd := len(prefix)+1, len(prefix)+3
			if language == "spl2" {
				suffix = "'é'"
			}
			query := prefix + suffix
			result := mustAnalyzeDocument(t, QueryDocument{Text: query, Language: language})
			omitted := rawLexerTokens(t, language, query)[lexerWorkLimit]
			assertResourceLimitedResult(t, result, query, omitted)
			if result.Diagnostics[0].Location.Start.Offset != wantStart || result.Diagnostics[0].Location.End.Offset != wantEnd {
				t.Fatalf("source-indexed byte range = %+v, want bytes [%d,%d)", result.Diagnostics[0].Location, wantStart, wantEnd)
			}
		})
	}
}

func TestSPL2LexerWorkBudgetLiteralClosure(t *testing.T) {
	t.Run("synthetic closure inside budget remains syntax", func(t *testing.T) {
		prefix, _ := lexerBoundaryQueriesAt(t, "spl2", lexerWorkLimit-2)
		query := prefix + `"`
		if got := len(rawLexerTokens(t, "spl2", query)); got != lexerWorkLimit-1 {
			t.Fatalf("raw token units = %d, want %d", got, lexerWorkLimit-1)
		}
		parsed := parseSPL2Document(query)
		if parsed.resourceLimit != nil || parsed.tree == nil {
			t.Fatalf("within-budget closure did not reach parser: %+v", parsed.resourceLimit)
		}
		found := false
		for _, diagnostic := range parsed.lexicalErrors {
			if diagnostic.Code == CodeSyntaxError && diagnostic.Message == "Unterminated literal" {
				found = true
			}
		}
		if !found {
			t.Fatalf("missing ordinary closure diagnostic: %+v", parsed.lexicalErrors)
		}
	})

	t.Run("synthetic closure is first omitted event", func(t *testing.T) {
		prefix, _ := lexerBoundaryQueriesAt(t, "spl2", lexerWorkLimit-1)
		query := prefix + `"`
		if got := len(rawLexerTokens(t, "spl2", query)); got != lexerWorkLimit {
			t.Fatalf("raw token units = %d, want %d", got, lexerWorkLimit)
		}
		result := mustAnalyzeDocument(t, QueryDocument{Text: query, Language: "spl2"})
		start := len([]rune(prefix))
		assertSingleResourceLimitLocation(t, result, newSourceIndex(query).location(start, start+1))
	})

	t.Run("earlier token overflow suppresses closure", func(t *testing.T) {
		_, plusOne := lexerBoundaryQueriesAt(t, "spl2", lexerWorkLimit)
		query := plusOne + ` "`
		result := mustAnalyzeDocument(t, QueryDocument{Text: query, Language: "spl2"})
		omitted := rawLexerTokens(t, "spl2", query)[lexerWorkLimit]
		assertResourceLimitedResult(t, result, query, omitted)
		if result.Diagnostics[0].Location.Start.Offset == len(plusOne)+1 {
			t.Fatal("literal closure replaced the earlier omitted token")
		}
	})
}

func TestAnalysisResourceLimitFullResult(t *testing.T) {
	_, query := lexerBoundaryQueries(t, "spl")
	document := QueryDocument{Text: query, SourceID: "queries/dense.spl"}
	result, trace, err := analyzeRewriteWithTrace(document, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	normalized, err := normalizeDocument(document)
	if err != nil {
		t.Fatal(err)
	}
	if result.Document != normalized || result.Document.Text != query || result.Status != Incomplete {
		t.Fatalf("resource result identity/status = %+v", result)
	}
	if result.Coverage.SyntaxComplete || result.Coverage.SemanticComplete || !reflect.DeepEqual(result.Coverage.Reasons, []string{CodeAnalysisResourceLimit}) {
		t.Fatalf("resource coverage = %+v", result.Coverage)
	}
	if result.Stages == nil || result.Scopes == nil || result.References == nil || result.Lineage == nil ||
		len(result.Stages) != 0 || len(result.Scopes) != 0 || len(result.References) != 0 || len(result.Lineage) != 0 {
		t.Fatalf("partial canonical evidence survived: stages=%v scopes=%v references=%v lineage=%v", result.Stages, result.Scopes, result.References, result.Lineage)
	}
	if !emptyDependencies(result.Dependencies) {
		t.Fatalf("partial dependencies survived: %+v", result.Dependencies)
	}
	if len(trace.references) != 0 || len(trace.diagnostics) != 1 || trace.pendingReferenceIndexes == nil || trace.incompleteStageIDs == nil {
		t.Fatalf("partial or uninitialized trace survived: %+v", trace)
	}
	assertResourceLimitRequirements(t, &result.Requirements, normalized, result.Diagnostics)
}

func TestRequirementsResourceLimitIsDetachedAndDeterministic(t *testing.T) {
	query := denseWildcardQuery(65_533, "spl2")
	document := QueryDocument{Text: query, Language: "spl2", SourceID: "dense.spl2"}
	analysisResult := mustAnalyzeDocument(t, document)
	standalone, err := Requirements(document)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(*standalone, analysisResult.Requirements) {
		t.Fatalf("standalone requirements differ:\nstandalone=%+v\nembedded=%+v", *standalone, analysisResult.Requirements)
	}
	before, err := json.Marshal(analysisResult.Requirements)
	if err != nil {
		t.Fatal(err)
	}
	standalone.Coverage.Reasons[0] = "mutated"
	standalone.Gaps[0].DiagnosticCodes[0] = "mutated"
	standalone.Diagnostics[0].Message = "mutated"
	after, err := json.Marshal(analysisResult.Requirements)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(before, after) {
		t.Fatal("standalone resource-limit requirements alias embedded requirements")
	}
	for i := 0; i < 3; i++ {
		repeated := mustAnalyzeDocument(t, document)
		got, err := json.Marshal(repeated)
		if err != nil {
			t.Fatal(err)
		}
		want, err := json.Marshal(analysisResult)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("resource-limit JSON changed on repeat %d", i)
		}
	}
}

func TestAnalysisLexerWorkBudgetAdmitsLongSparseInput(t *testing.T) {
	for _, tc := range []struct {
		language string
		base     string
	}{
		{language: "spl", base: "search host=x"},
		{language: "spl2", base: "FROM main"},
	} {
		t.Run(tc.language, func(t *testing.T) {
			query := tc.base + strings.Repeat(" ", 300_000)
			if got := len(rawLexerTokens(t, tc.language, query)); got > lexerWorkLimit {
				t.Fatalf("sparse fixture has %d units", got)
			}
			result := mustAnalyzeDocument(t, QueryDocument{Text: query, Language: tc.language})
			if len(result.Diagnostics) == 1 && result.Diagnostics[0].Code == CodeAnalysisResourceLimit {
				t.Fatal("long sparse input hit a byte-size limit")
			}
			if len(result.Stages) == 0 {
				t.Fatalf("long sparse input did not reach parser: %+v", result)
			}
		})
	}
}

func TestAnalysisResourceLimitAdversarialResponseBounds(t *testing.T) {
	for _, language := range []string{"spl", "spl2"} {
		var first Location
		for _, size := range []int{65_533, 256 * 1024} {
			t.Run(fmt.Sprintf("%s/%d", language, size), func(t *testing.T) {
				query := denseWildcardQuery(size, language)
				result := mustAnalyzeDocument(t, QueryDocument{Text: query, Language: language})
				if len(result.Diagnostics) != 1 || result.Diagnostics[0].Code != CodeAnalysisResourceLimit {
					t.Fatalf("dense fixture was not bounded: %+v", result.Diagnostics)
				}
				if first == (Location{}) {
					first = result.Diagnostics[0].Location
				} else if result.Diagnostics[0].Location != first {
					t.Fatalf("first omitted event changed with input size: got %+v want %+v", result.Diagnostics[0].Location, first)
				}
				requirementsJSON, err := json.Marshal(result.Requirements)
				if err != nil {
					t.Fatal(err)
				}
				resultJSON, err := json.Marshal(result)
				if err != nil {
					t.Fatal(err)
				}
				if len(requirementsJSON) > 4096 {
					t.Fatalf("RequirementSet bytes = %d, want <= 4096", len(requirementsJSON))
				}
				if len(resultJSON) > len(query)+4096 {
					t.Fatalf("Result bytes = %d, want <= %d", len(resultJSON), len(query)+4096)
				}
				t.Logf("query=%d requirements=%d result=%d", len(query), len(requirementsJSON), len(resultJSON))
			})
		}
	}
}

var (
	benchmarkResourceLimitJSON       []byte
	benchmarkResourceRequirementJSON []byte
)

func BenchmarkAnalysisResourceLimitAdversarial(b *testing.B) {
	for _, language := range []string{"spl", "spl2"} {
		for _, size := range []int{65_533, 256 * 1024} {
			query := denseWildcardQuery(size, language)
			document := QueryDocument{Text: query, Language: language}
			b.Run(fmt.Sprintf("%s/%d", language, size), func(b *testing.B) {
				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					result, err := Analyze(document)
					if err != nil {
						b.Fatal(err)
					}
					benchmarkResourceLimitJSON, err = json.Marshal(result)
					if err != nil {
						b.Fatal(err)
					}
				}
				b.StopTimer()
				result, err := Analyze(document)
				if err != nil {
					b.Fatal(err)
				}
				benchmarkResourceRequirementJSON, err = json.Marshal(result.Requirements)
				if err != nil {
					b.Fatal(err)
				}
				b.ReportMetric(float64(len(benchmarkResourceLimitJSON)), "result-bytes")
				b.ReportMetric(float64(len(benchmarkResourceRequirementJSON)), "requirements-bytes")
			})
		}
	}
}

func lexerBoundaryQueries(t *testing.T, language string) (string, string) {
	t.Helper()
	return lexerBoundaryQueriesAt(t, language, lexerWorkLimit)
}

func lexerBoundaryQueriesAt(t *testing.T, language string, units int) (string, string) {
	t.Helper()
	seed := strings.Repeat("x ", units+16)
	tokens := rawLexerTokens(t, language, seed)
	if len(tokens) <= units {
		t.Fatalf("seed has %d tokens, need %d", len(tokens), units+1)
	}
	exactEnd := tokens[units-1].GetStop() + 1
	plusEnd := tokens[units].GetStop() + 1
	exact := string([]rune(seed)[:exactEnd])
	plusOne := string([]rune(seed)[:plusEnd])
	if got := len(rawLexerTokens(t, language, exact)); got != units {
		t.Fatalf("exact fixture has %d tokens, want %d", got, units)
	}
	if got := len(rawLexerTokens(t, language, plusOne)); got != units+1 {
		t.Fatalf("plus-one fixture has %d tokens, want %d", got, units+1)
	}
	return exact, plusOne
}

func rawLexerTokens(t *testing.T, language, text string) []antlr.Token {
	t.Helper()
	input := antlr.NewInputStream(text)
	var lexer antlr.Lexer
	if language == "spl2" {
		lexer = spl2.NewSPL2Lexer(input)
	} else {
		lexer = parser.NewSPLLexer(input)
	}
	lexer.RemoveErrorListeners()
	tokens := []antlr.Token{}
	for {
		token := lexer.NextToken()
		if token.GetTokenType() == antlr.TokenEOF {
			return tokens
		}
		tokens = append(tokens, token)
	}
}

func denseWildcardQuery(size int, language string) string {
	pattern := "fields a* | "
	if language == "spl2" {
		pattern = "FROM main | fields a* | "
	}
	var out strings.Builder
	out.Grow(size)
	for out.Len() < size {
		out.WriteString(pattern)
	}
	return out.String()[:size]
}

func mustAnalyzeDocument(t *testing.T, document QueryDocument) *Result {
	t.Helper()
	result, err := Analyze(document)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func assertResourceLimitedResult(t *testing.T, result *Result, text string, omitted antlr.Token) {
	t.Helper()
	want := newSourceIndex(text).location(omitted.GetStart(), omitted.GetStop()+1)
	assertSingleResourceLimitLocation(t, result, want)
}

func assertSingleResourceLimitLocation(t *testing.T, result *Result, want Location) {
	t.Helper()
	if len(result.Diagnostics) != 1 {
		t.Fatalf("diagnostics = %+v", result.Diagnostics)
	}
	diagnostic := result.Diagnostics[0]
	if diagnostic.Code != CodeAnalysisResourceLimit || diagnostic.Severity != "warning" || diagnostic.Category != "resource_limit" ||
		diagnostic.Message != analysisResourceLimitMessage || diagnostic.StageID != "" || diagnostic.ScopeID != "" || diagnostic.Location != want {
		t.Fatalf("resource diagnostic = %+v, want location %+v", diagnostic, want)
	}
}

func assertNotResourceLimited(t *testing.T, result *Result) {
	t.Helper()
	for _, diagnostic := range result.Diagnostics {
		if diagnostic.Code == CodeAnalysisResourceLimit {
			t.Fatalf("exact-limit analysis was resource limited: %+v", diagnostic)
		}
	}
}

func assertResourceLimitRequirements(t *testing.T, set *RequirementSet, document QueryDocument, diagnostics []Diagnostic) {
	t.Helper()
	if set.SchemaVersion != 1 || set.Query.SourceID != document.SourceID || set.Query.Language != document.Language || set.Query.Profile != document.Profile || set.Query.Version != document.Version || set.Query.QueryDigest != queryDigest(document.Text) {
		t.Fatalf("resource requirement identity = %+v", set.Query)
	}
	wantRevision, err := capabilityRevision(document)
	if err != nil {
		t.Fatal(err)
	}
	if set.CapabilityRevision != wantRevision || set.QueryStatus != Incomplete || set.Coverage.Complete || !reflect.DeepEqual(set.Coverage.Reasons, []string{CodeAnalysisResourceLimit}) {
		t.Fatalf("resource requirement status = %+v", set)
	}
	if set.Items == nil || len(set.Items) != 0 || set.Gaps == nil || len(set.Gaps) != 1 || set.Diagnostics == nil || !reflect.DeepEqual(set.Diagnostics, diagnostics) {
		t.Fatalf("resource requirement collections = %+v", set)
	}
	wantGap := RequirementGap{Code: CodeAnalysisResourceLimit, Message: requirementResourceLimitMessage, ReferenceIDs: []string{}, DiagnosticCodes: []string{CodeAnalysisResourceLimit}}
	if !reflect.DeepEqual(set.Gaps[0], wantGap) {
		t.Fatalf("resource requirement gap = %+v, want %+v", set.Gaps[0], wantGap)
	}
	encoded, err := json.Marshal(set)
	if err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{`"reasons":[`, `"items":[]`, `"reference_ids":[]`, `"diagnostic_codes":[`, `"diagnostics":[`} {
		if !strings.Contains(string(encoded), field) {
			t.Fatalf("resource requirements missing initialized array %s: %s", field, encoded)
		}
	}
}

func emptyDependencies(dependencies Dependencies) bool {
	return dependencies.Indexes != nil && len(dependencies.Indexes) == 0 &&
		dependencies.Sources != nil && len(dependencies.Sources) == 0 &&
		dependencies.SourceTypes != nil && len(dependencies.SourceTypes) == 0 &&
		dependencies.Datasets != nil && len(dependencies.Datasets) == 0 &&
		dependencies.Lookups != nil && len(dependencies.Lookups) == 0 &&
		dependencies.DataModels != nil && len(dependencies.DataModels) == 0 &&
		dependencies.Macros != nil && len(dependencies.Macros) == 0
}
