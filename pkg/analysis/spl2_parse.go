package analysis

import (
	"github.com/antlr4-go/antlr/v4"
	"github.com/delgado-jacob/spl-toolkit/parser/spl2"
)

// This private frontend does not select a dialect or expose semantic completeness.
// Every mutable stream, source mapping, recovery listener and syntax fact is local.
type spl2ParsedDocument struct {
	tokens           *antlr.CommonTokenStream
	tree             spl2.IQueryContext
	source           *sourceIndex
	diagnostics      []Diagnostic
	syntax           *spl2SyntaxNode
	syntaxComplete   bool
	semanticComplete bool
}

type spl2SyntaxListener struct {
	*antlr.DefaultErrorListener
	parsed *spl2ParsedDocument
}

func (l *spl2SyntaxListener) SyntaxError(recognizer antlr.Recognizer, offending interface{}, line, column int, msg string, e antlr.RecognitionException) {
	s := l.parsed.source
	start, end := len(s.positions)-1, len(s.positions)-1
	if token, ok := offending.(antlr.Token); ok && token.GetStart() >= 0 {
		start = token.GetStart()
		end = token.GetStop() + 1
	} else if lexer, ok := recognizer.(antlr.Lexer); ok {
		start = lexer.GetInputStream().Index()
		end = start + 1
	}
	l.parsed.diagnostics = append(l.parsed.diagnostics, Diagnostic{Code: CodeSyntaxError, Severity: "error", Category: "syntax", Message: msg, Location: s.location(start, end)})
}
func parseSPL2Document(text string) *spl2ParsedDocument {
	parsed := &spl2ParsedDocument{source: newSourceIndex(text), diagnostics: []Diagnostic{}}
	listener := &spl2SyntaxListener{DefaultErrorListener: antlr.NewDefaultErrorListener(), parsed: parsed}
	lexer := spl2.NewSPL2Lexer(antlr.NewInputStream(text))
	lexer.RemoveErrorListeners()
	lexer.AddErrorListener(listener)
	parsed.tokens = antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	parser := spl2.NewSPL2Parser(parsed.tokens)
	parser.RemoveErrorListeners()
	parser.AddErrorListener(listener)
	parsed.tree = parser.Query()
	parsed.syntaxComplete = len(parsed.diagnostics) == 0
	parsed.syntax = spl2TreeFacts(parsed.tree, parsed.source, parser.RuleNames, parser.SymbolicNames)
	parsed.inspectSyntax(parsed.tree, 0)
	return parsed
}
