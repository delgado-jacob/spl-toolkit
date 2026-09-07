package analysis

import (
	"github.com/antlr4-go/antlr/v4"
	"github.com/delgado-jacob/spl-toolkit/parser"
)

// sound excludes parser insertions/deletions and lexer damage inside or
// immediately adjacent to a context (a deleted prefix/suffix can change a name).
func (s *semanticStage) sound(ctx antlr.ParserRuleContext) bool {
	if ctx == nil || !intact(ctx) {
		return false
	}
	loc := s.parsed.source.contextLocation(ctx)
	if s.recoveryLimit > 0 && loc.End.Offset > s.recoveryLimit {
		return false
	}
	for _, d := range s.parsed.diagnostics {
		if d.Location.End.Offset >= loc.Start.Offset && d.Location.Start.Offset <= loc.End.Offset {
			return false
		}
	}
	return true
}

// A generic recovered eval context can have lost every typed assignment. Re-run
// the assignment rule on the original token stream to retain its sound prefix.
// This never slices source or guesses expression structure. Stop at the first
// damaged assignment; the ordinary pipeline parse owns all later stages.
func (s *semanticStage) recoverStage(ctx antlr.ParserRuleContext) {
	command := s.result.Stages[s.stage].Command
	if command != "eval" && command != "search" {
		return
	}
	tokens := s.parsed.tokens
	saved := tokens.Index()
	defer tokens.Seek(saved)
	tokens.Seek(ctx.GetStart().GetTokenIndex() + 1)
	p := parser.NewSPLParser(tokens)
	p.RemoveErrorListeners()
	recovered := &parsedDocument{source: s.parsed.source, diagnostics: []Diagnostic{}}
	p.AddErrorListener(&syntaxListener{DefaultErrorListener: antlr.NewDefaultErrorListener(), parsed: recovered})
	if command == "search" {
		if _, implicit := ctx.(*parser.AnalysisImplicitSearchContext); implicit {
			tokens.Seek(ctx.GetStart().GetTokenIndex())
		}
		prefix := *s
		// An unclosed bracket may be synthetically closed before remaining tokens.
		// Keep only the prefix before a subquery rather than re-owning child terms.
		for _, token := range tokens.GetAllTokens() {
			if token.GetTokenIndex() >= ctx.GetStart().GetTokenIndex() && token.GetTokenType() == parser.SPLLexerLBRACK {
				prefix.recoveryLimit = s.parsed.source.location(token.GetStart(), token.GetStart()).Start.Offset
				break
			}
		}
		searchCommand(&prefix, p.AnalysisSearch())
		return
	}
	for tokens.LT(1).GetTokenIndex() <= ctx.GetStop().GetTokenIndex() {
		a := p.AnalysisAssignment()
		if len(recovered.diagnostics) > 0 || !s.sound(a) || a.GetStop().GetTokenIndex() > ctx.GetStop().GetTokenIndex() {
			break
		}
		if a.AnalysisIdentifier() == nil || a.AnalysisExpression() == nil {
			break
		}
		inputs := s.expression(a.AnalysisExpression())
		s.create(a.AnalysisIdentifier(), normalizedName(a.AnalysisIdentifier().GetText()), "create", "create", inputs, !s.result.Stages[s.stage].SemanticComplete)
		if tokens.LA(1) != parser.SPLLexerCOMMA {
			break
		}
		tokens.Consume()
	}
}

// A pipeline separator is trustworthy only at its scope's parenthesis depth.
// Bracket entry saves the outer depth, so a valid child pipeline in an outer
// expression still has its own independent command boundaries. Each token also
// retains the identity of its original opening bracket (-1 for the root). A
// depth alone cannot distinguish sibling scopes after parser recovery.
func originalTokenBoundaries(parsed *parsedDocument) (map[int]bool, map[int]int) {
	unsafe := map[int]bool{}
	owners := map[int]int{}
	depth := 0
	type bracket struct{ depth, token int }
	stack := []bracket{}
	badNext := false
	for _, token := range parsed.tokens.GetAllTokens() {
		if token.GetChannel() != antlr.TokenDefaultChannel {
			continue
		}
		owners[token.GetTokenIndex()] = -1
		if len(stack) > 0 {
			owners[token.GetTokenIndex()] = stack[len(stack)-1].token
		}
		if badNext {
			unsafe[token.GetTokenIndex()] = true
			badNext = false
		}
		switch token.GetTokenType() {
		case parser.SPLLexerLPAREN:
			depth++
		case parser.SPLLexerRPAREN:
			if depth > 0 {
				depth--
			}
		case parser.SPLLexerLBRACK:
			stack = append(stack, bracket{depth, token.GetTokenIndex()})
			depth = 0
		case parser.SPLLexerRBRACK:
			if len(stack) > 0 {
				depth = stack[len(stack)-1].depth
				stack = stack[:len(stack)-1]
			}
		case parser.SPLLexerPIPE:
			badNext = depth != 0
		}
	}
	return unsafe, owners
}
