package analysis

import (
	"strconv"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	"github.com/delgado-jacob/spl-toolkit/parser/spl2"
)

// Rule names are grammar-owned kinds, not guesses from source text. These facts
// retain nested operand ownership and exact locations for conformance assertions.
// Semantic lowering should consume the generated contexts, not parse Value.
type spl2SyntaxNode struct {
	Kind     string            `json:"kind"`
	Value    string            `json:"value,omitempty"`
	Location Location          `json:"location"`
	Children []*spl2SyntaxNode `json:"children,omitempty"`
}

func spl2TreeFacts(tree antlr.Tree, source *sourceIndex, rules, symbols []string) *spl2SyntaxNode {
	switch node := tree.(type) {
	case antlr.ParserRuleContext:
		result := &spl2SyntaxNode{Kind: rules[node.GetRuleIndex()], Location: source.contextLocation(node)}
		for _, child := range node.GetChildren() {
			if fact := spl2TreeFacts(child, source, rules, symbols); fact != nil {
				result.Children = append(result.Children, fact)
			}
		}
		return result
	case antlr.TerminalNode:
		token := node.GetSymbol()
		kind := token.GetTokenType()
		if kind <= 0 || kind >= len(symbols) {
			return nil
		}
		return &spl2SyntaxNode{Kind: symbols[kind], Value: token.GetText(), Location: source.location(token.GetStart(), token.GetStop()+1)}
	}
	return nil
}

// shape removes single-child grammar routing rules, retaining all branching
// constructs, operator tokens and literal/name leaves. It is stable test data.
func (n *spl2SyntaxNode) shape() string {
	if len(n.Children) == 0 {
		return n.Kind + ":" + strconv.Quote(n.Value)
	}
	if len(n.Children) == 1 {
		return n.Children[0].shape()
	}
	parts := []string{n.Kind}
	for _, child := range n.Children {
		parts = append(parts, child.shape())
	}
	return "(" + strings.Join(parts, " ") + ")"
}
func (p *spl2ParsedDocument) syntaxFinding(ctx antlr.ParserRuleContext, code, category, message string) {
	p.diagnostics = append(p.diagnostics, Diagnostic{Code: code, Severity: "error", Category: category, Message: message, Location: p.source.contextLocation(ctx)})
}
func (p *spl2ParsedDocument) inspectSyntax(tree antlr.Tree, lambdaDepth int) {
	switch ctx := tree.(type) {
	case *spl2.ModuleSuffixContext:
		p.syntaxComplete = false
		p.syntaxFinding(ctx, "SPL_UNSUPPORTED_MODULE", "unsupported", "Top-level statement terminators are outside the standalone contract")
	case *spl2.ModuleDeclarationContext:
		p.syntaxComplete = false
		p.syntaxFinding(ctx, "SPL_UNSUPPORTED_MODULE", "unsupported", "Module declarations are outside the standalone contract")
	case *spl2.LambdaExpressionContext:
		if lambdaDepth > 0 {
			p.syntaxFinding(ctx, CodeSyntaxError, "contract", "Nested lambdas are not supported by the language contract")
		}
		lambdaDepth++
	case *spl2.SearchCommandContext:
		p.rejectSearchComments(ctx)
	case *spl2.ImplicitSearchContext:
		p.rejectSearchComments(ctx)
	case *spl2.ObjectContext:
		keys := map[string]bool{}
		for _, entry := range ctx.AllObjectEntry() {
			key := entry.ObjectKey()
			decoded, ok := spl2DecodeKey(key.GetText())
			if !ok {
				continue
			}
			if keys[decoded] {
				p.syntaxFinding(key, CodeSyntaxError, "contract", "Duplicate object key "+strconv.Quote(decoded))
			}
			keys[decoded] = true
		}
	case *spl2.ArrayContext:
		if len(ctx.AllCOMMA()) == len(ctx.AllExpression()) && len(ctx.AllExpression()) > 0 {
			p.heldSyntax(ctx, "H05 array trailing comma remains held")
		}
	case *spl2.PredicateContext:
		if ctx.IS() != nil && ctx.NOT() != nil && ctx.TYPE() != nil {
			p.heldSyntax(ctx, "EH04 IS NOT type remains held")
		}
	case *spl2.SearchLiteralContext:
		p.heldSyntax(ctx, "Expression search literal body is not yet modeled")
	case *spl2.EmbeddedCommandContext:
		p.heldSyntax(ctx, "Embedded command body is not yet modeled")
	case *spl2.RexCommandContext:
		if regex := ctx.REGEX(); regex != nil && strings.Contains(regex.GetText(), "|") {
			p.syntaxComplete = false
			p.diagnostics = append(p.diagnostics, Diagnostic{Code: CodeUnsupportedSemantics, Severity: "warning", Category: "unsupported", Message: "H12 slash regex with internal pipe remains held", Location: p.source.contextLocation(ctx)})
		}
	}
	for _, child := range tree.GetChildren() {
		p.inspectSyntax(child, lambdaDepth)
	}
}
func (p *spl2ParsedDocument) rejectSearchComments(ctx antlr.ParserRuleContext) {
	start := ctx.GetStart().GetTokenIndex()
	for _, token := range p.tokens.GetAllTokens() {
		if token.GetTokenIndex() < start {
			continue
		}
		if token.GetTokenType() == spl2.SPL2LexerPIPE {
			break
		}
		if token.GetTokenType() == spl2.SPL2LexerBLOCK_COMMENT {
			p.syntaxComplete = false
			p.diagnostics = append(p.diagnostics, Diagnostic{Code: CodeSyntaxError, Severity: "error", Category: "syntax", Message: "Block comments are forbidden in search portions", Location: p.source.location(token.GetStart(), token.GetStop()+1)})
		}
	}
}
func spl2DecodeKey(text string) (string, bool) {
	if len(text) == 0 {
		return "", false
	}
	if text[0] != '\'' && text[0] != '"' {
		return text, true
	}
	quote := text[0]
	if len(text) < 2 || text[len(text)-1] != quote {
		return "", false
	}
	body := text[1 : len(text)-1]
	var decoded strings.Builder
	for len(body) > 0 {
		value, _, tail, err := strconv.UnquoteChar(body, quote)
		if err != nil {
			return "", false
		}
		decoded.WriteRune(value)
		body = tail
	}
	return decoded.String(), true
}

func (p *spl2ParsedDocument) heldSyntax(ctx antlr.ParserRuleContext, message string) {
	p.syntaxComplete = false
	p.diagnostics = append(p.diagnostics, Diagnostic{Code: CodeUnsupportedSemantics, Severity: "warning", Category: "unsupported", Message: message, Location: p.source.contextLocation(ctx)})
}
