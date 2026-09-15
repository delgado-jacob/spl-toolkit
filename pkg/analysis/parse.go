package analysis

import (
	"github.com/antlr4-go/antlr/v4"
	"github.com/delgado-jacob/spl-toolkit/parser"
)

// analysisInputStream opts this document into the analysis lexer contract.
// Embedding preserves the ANTLR stream unchanged; the marker has no mutable state.
type analysisInputStream struct{ *antlr.InputStream }

func (*analysisInputStream) SPLAnalysisSyntax() bool { return true }

type parsedDocument struct {
	tokens        *antlr.CommonTokenStream
	tree          parser.IAnalysisQueryContext
	source        *sourceIndex
	diagnostics   []Diagnostic
	resourceLimit *Location
}
type syntaxListener struct {
	*antlr.DefaultErrorListener
	parsed  *parsedDocument
	tracker *lexerWorkTracker
}

type splParserFactory func(antlr.TokenStream) *parser.SPLParser

func (l *syntaxListener) SyntaxError(recognizer antlr.Recognizer, offendingSymbol interface{}, line, column int, msg string, e antlr.RecognitionException) {
	s := l.parsed.source
	start, end := len(s.positions)-1, len(s.positions)-1
	if token, ok := offendingSymbol.(antlr.Token); ok && token.GetStart() >= 0 {
		start = token.GetStart()
		end = token.GetStop() + 1
	} else if lexer, ok := recognizer.(antlr.Lexer); ok {
		start = lexer.GetInputStream().Index()
		end = start + 1
	}
	location := s.location(start, end)
	if _, lexical := recognizer.(antlr.Lexer); lexical && l.tracker != nil {
		l.tracker.consumeLexerError(location)
	}
	l.parsed.diagnostics = append(l.parsed.diagnostics, Diagnostic{Code: CodeSyntaxError, Severity: "error", Category: "syntax", Message: msg, Location: location})
}
func parseDocument(text string) *parsedDocument {
	return parseDocumentWithParserFactory(text, parser.NewSPLParser)
}

func parseDocumentWithParserFactory(text string, newParser splParserFactory) *parsedDocument {
	parsed := &parsedDocument{source: newSourceIndex(text), diagnostics: []Diagnostic{}}
	tracker := &lexerWorkTracker{}
	listener := &syntaxListener{DefaultErrorListener: antlr.NewDefaultErrorListener(), parsed: parsed, tracker: tracker}
	lexer := parser.NewSPLLexer(&analysisInputStream{InputStream: antlr.NewInputStream(text)})
	lexer.RemoveErrorListeners()
	lexer.AddErrorListener(listener)
	parsed.tokens = preflightLexer(lexer, tracker, parsed.source)
	if tracker.resourceLimit != nil {
		parsed.resourceLimit = tracker.resourceLimit
		return parsed
	}
	p := newParser(parsed.tokens)
	p.RemoveErrorListeners()
	p.AddErrorListener(listener)
	parsed.tree = p.AnalysisQuery()
	return parsed
}
