package analysis

import (
	"github.com/antlr4-go/antlr/v4"
	"github.com/delgado-jacob/spl-toolkit/parser"
)

type parsedDocument struct {
	tokens      *antlr.CommonTokenStream
	tree        parser.IAnalysisQueryContext
	source      *sourceIndex
	diagnostics []Diagnostic
}
type syntaxListener struct {
	*antlr.DefaultErrorListener
	parsed *parsedDocument
}

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
	l.parsed.diagnostics = append(l.parsed.diagnostics, Diagnostic{Code: CodeSyntaxError, Severity: "error", Category: "syntax", Message: msg, Location: s.location(start, end)})
}
func parseDocument(text string) *parsedDocument {
	parsed := &parsedDocument{source: newSourceIndex(text), diagnostics: []Diagnostic{}}
	listener := &syntaxListener{DefaultErrorListener: antlr.NewDefaultErrorListener(), parsed: parsed}
	lexer := parser.NewSPLLexer(antlr.NewInputStream(text))
	lexer.RemoveErrorListeners()
	lexer.AddErrorListener(listener)
	parsed.tokens = antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	p := parser.NewSPLParser(parsed.tokens)
	p.RemoveErrorListeners()
	p.AddErrorListener(listener)
	parsed.tree = p.AnalysisQuery()
	return parsed
}
