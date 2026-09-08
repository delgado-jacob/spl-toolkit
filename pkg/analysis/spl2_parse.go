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
	lexicalErrors    []Diagnostic
	predictionErrors []spl2PredictionError
	missingSelectEOF []spl2MissingSelectEOF
	recoveredSelects []spl2RecoveredSelect
	syntax           *spl2SyntaxNode
	syntaxComplete   bool
	semanticComplete bool
}

// Private prediction provenance permits a narrowly proved recovery site to
// distinguish its ANTLR prediction error without inspecting message text.
type spl2PredictionError struct {
	diagnostic Diagnostic
	context    antlr.ParserRuleContext
}

// This is diagnostic attribution, never a replacement parse or syntax waiver.
type spl2MissingSelectEOF struct {
	diagnostic Diagnostic
	owner      *spl2.FromCommandContext
	token      antlr.Token
}

func spl2ExpectsSelectOrNewline(expected *antlr.IntervalSet) bool {
	count := 0
	for _, interval := range expected.GetIntervals() {
		for token := interval.Start; token < interval.Stop; token++ {
			if token != spl2.SPL2ParserSELECT && token != spl2.SPL2ParserNL {
				return false
			}
			count++
		}
	}
	return count == 2
}

type spl2SyntaxListener struct {
	*antlr.DefaultErrorListener
	parsed *spl2ParsedDocument
}

func (l *spl2SyntaxListener) SyntaxError(recognizer antlr.Recognizer, offending interface{}, line, column int, msg string, e antlr.RecognitionException) {
	s := l.parsed.source
	start, end := len(s.positions)-1, len(s.positions)-1
	lexical := false
	if token, ok := offending.(antlr.Token); ok && token.GetStart() >= 0 {
		start = token.GetStart()
		end = token.GetStop() + 1
	} else if lexer, ok := recognizer.(antlr.Lexer); ok {
		lexical = true
		start = lexer.GetInputStream().Index()
		end = start + 1
	}
	diagnostic := Diagnostic{Code: CodeSyntaxError, Severity: "error", Category: "syntax", Message: msg, Location: s.location(start, end)}
	l.parsed.diagnostics = append(l.parsed.diagnostics, diagnostic)
	if _, ok := e.(*antlr.NoViableAltException); ok {
		if parser, ok := recognizer.(antlr.Parser); ok {
			l.parsed.predictionErrors = append(l.parsed.predictionErrors, spl2PredictionError{diagnostic, parser.GetParserRuleContext()})
		}
	}
	if _, ok := e.(*antlr.InputMisMatchException); ok {
		if parser, ok := recognizer.(antlr.Parser); ok {
			if owner, ok := parser.GetParserRuleContext().(*spl2.FromCommandContext); ok && spl2ExpectsSelectOrNewline(parser.GetExpectedTokens()) {
				if token, ok := offending.(antlr.Token); ok && token == parser.GetCurrentToken() && token.GetTokenType() == antlr.TokenEOF {
					l.parsed.missingSelectEOF = append(l.parsed.missingSelectEOF, spl2MissingSelectEOF{diagnostic, owner, token})
				}
			}
		}
	}
	if lexical {
		l.parsed.lexicalErrors = append(l.parsed.lexicalErrors, diagnostic)
	}
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
	parsed.inspectLiteralClosure()
	parsed.syntaxComplete = len(parsed.diagnostics) == 0
	parsed.syntax = spl2TreeFacts(parsed.tree, parsed.source, parser.RuleNames, parser.SymbolicNames)
	parsed.inspectSyntax(parsed.tree, 0)
	return parsed
}

// EOF does not necessarily emit a lexer error from an open literal mode. Replay
// only its original opener/end tokens; text, escapes and raw strings are opaque,
// and literals inside interpolation nest without closing their owning literal.
func (p *spl2ParsedDocument) inspectLiteralClosure() {
	p.tokens.Fill()
	type literal struct {
		opener antlr.Token
		end    int
	}
	stack := []literal{}
	for _, token := range p.tokens.GetAllTokens() {
		kind := token.GetTokenType()
		switch kind {
		case spl2.SPL2LexerDQUOTE:
			stack = append(stack, literal{token, spl2.SPL2LexerSTRING_END})
		case spl2.SPL2LexerSQUOTE:
			stack = append(stack, literal{token, spl2.SPL2LexerNAME_END})
		case spl2.SPL2LexerBACKTICK:
			stack = append(stack, literal{token, spl2.SPL2LexerEMBEDDED_END})
		case spl2.SPL2LexerSTRING_END, spl2.SPL2LexerNAME_END, spl2.SPL2LexerEMBEDDED_END:
			if len(stack) > 0 && stack[len(stack)-1].end == kind {
				stack = stack[:len(stack)-1]
			}
		}
	}
	if len(stack) == 0 {
		return
	}
	opener := stack[0].opener
	diagnostic := Diagnostic{Code: CodeSyntaxError, Severity: "error", Category: "syntax", Message: "Unterminated literal", Location: p.source.location(opener.GetStart(), opener.GetStop()+1)}
	p.diagnostics = append(p.diagnostics, diagnostic)
	p.lexicalErrors = append(p.lexicalErrors, diagnostic)
}
