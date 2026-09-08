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
	case *spl2.RenameCommandContext:
		p.inspectRename(ctx)
	case *spl2.TableFieldContext:
		if spl2IntactSyntax(ctx) {
			if literal := ctx.StringLiteral(); literal != nil && len(literal.AllSTRING_INTERPOLATION()) > 0 {
				p.heldSyntax(ctx, "Interpolated table field remains unproved")
			} else if name, ok := spl2DecodeKey(ctx.GetText()); !ok {
				p.heldSyntax(ctx, "Unproved quoted table field")
			} else if strings.Contains(name, "*") {
				p.heldSyntax(ctx, "H01 table wildcard quoting remains held")
			} else if ctx.StringLiteral() != nil {
				p.syntaxFinding(ctx, CodeSyntaxError, "contract", "Exact table fields require identifiers")
			}
		}
	case *spl2.GroupFieldContext:
		if span := ctx.GroupSpan(); span != nil {
			if _, streaming := ctx.GetParent().(*spl2.StreamGroupContext); streaming {
				p.heldSyntax(span, "Grouping span in streamstats remains unproved")
			}
		}
		if name := ctx.Identifier(); name != nil && spl2IntactSyntax(name) {
			if decoded, ok := spl2DecodeKey(name.GetText()); !ok {
				p.heldSyntax(name, "Unproved quoted grouping field")
			} else if strings.Contains(decoded, "*") {
				p.syntaxFinding(name, CodeSyntaxError, "contract", "Grouping fields cannot contain wildcards")
			}
		}
	case *spl2.StreamPostLayoutContext:
		p.heldSyntax(ctx, "H02 postaggregate streamstats layout remains held")
	case *spl2.HeadPostLayoutContext:
		p.heldSyntax(ctx, "H03 head count before while remains held")
	case *spl2.SearchTimeModifierContext:
		if op := ctx.Comparison(); op != nil && op.GetText() != "=" && op.GetText() != "!=" {
			p.heldSyntax(ctx, "EH03 time modifier comparison remains held")
		}
	case *spl2.SearchSignedNumberContext:
		if spl2IntactSyntax(ctx) {
			p.heldSyntax(ctx, "Unquoted signed search number remains unproved")
		}
	case *spl2.SearchAtomContext:
		if name := ctx.Identifier(); name != nil && ctx.Comparison() != nil {
			switch name.GetStart().GetTokenType() {
			case spl2.SPL2ParserEARLIEST, spl2.SPL2ParserLATEST, spl2.SPL2ParserINDEX_EARLIEST, spl2.SPL2ParserINDEX_LATEST, spl2.SPL2ParserSTARTTIME, spl2.SPL2ParserENDTIME, spl2.SPL2ParserTIMEFORMAT:
				p.heldSyntax(ctx, "Unproved time modifier value retains incomplete syntax coverage")
			}
		}
		if ctx.IN() != nil && len(ctx.AllSearchValue()) == 1 {
			p.heldSyntax(ctx, "Single-element search IN remains unproved")
		}
	case *spl2.IntegerValueContext:
		positive := false
		_, positive = ctx.GetParent().(*spl2.DedupCommandContext)
		p.inspectInteger(ctx, positive)
	case *spl2.WindowOptionContext:
		if ctx.NUMBER() != nil {
			p.inspectInteger(ctx, false)
		}
	case *spl2.UnknownOptionContext:
		p.inspectUnknownOption(ctx)
	case *spl2.AggregateContext:
		p.inspectAggregate(ctx)
	case *spl2.CallContext:
		if name := ctx.Identifier(); name != nil && spl2IntactSyntax(ctx) && spl2StatisticalFunction(name.GetText()) && name.GetText() != "min" && name.GetText() != "max" {
			for parent := ctx.GetParent(); parent != nil; parent = parent.GetParent() {
				if _, ok := parent.(*spl2.HeadWhileContext); ok {
					p.syntaxFinding(ctx, CodeSyntaxError, "contract", "Statistical functions are forbidden in head while")
					break
				}
			}
		}
	case *spl2.ModuleSuffixContext:
		p.syntaxComplete = false
		p.syntaxFinding(ctx, "SPL_UNSUPPORTED_MODULE", "unsupported", "Top-level statement terminators are outside the standalone contract")
	case *spl2.ModuleDeclarationContext:
		p.syntaxComplete = false
		p.syntaxFinding(ctx, "SPL_UNSUPPORTED_MODULE", "unsupported", "Module declarations are outside the standalone contract")
	case *spl2.LambdaParameterContext:
		if value := ctx.StringLiteral(); value != nil && len(value.AllSTRING_INTERPOLATION()) > 0 {
			p.syntaxFinding(value, CodeSyntaxError, "contract", "Lambda default strings must be constant")
		}
	case *spl2.LogicalAndContext:
		p.inspectOperatorCase(ctx, "AND")
	case *spl2.LogicalOrContext:
		p.inspectOperatorCase(ctx, "OR")
	case *spl2.LogicalXorContext:
		p.inspectOperatorCase(ctx, "XOR")
	case *spl2.LogicalNotContext:
		p.inspectOperatorCase(ctx, "NOT")
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
		if operator := ctx.BetweenOperator(); operator != nil && ctx.BetweenConjunction() != nil {
			between, conjunction := operator.GetText(), ctx.BetweenConjunction().GetText()
			// The lowercase pair does not authorize additional NOT casing combinations.
			complete := between == "BETWEEN" && conjunction == "AND" || between == "between" && conjunction == "and" && ctx.LogicalNot() == nil
			if !complete {
				p.heldSyntax(ctx, "H11 unproved BETWEEN/AND casing remains held")
			}
		}
		if ctx.IS() != nil && ctx.LogicalNot() != nil && ctx.TYPE() != nil {
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

func (p *spl2ParsedDocument) inspectOperatorCase(ctx antlr.ParserRuleContext, documented string) {
	if ctx.GetText() != documented {
		p.heldSyntax(ctx, "H11 unproved logical operator casing remains held")
	}
}

// Contract validation reads typed operands; it never infers command boundaries
// or rewrites the caller's query. Recovery-damaged operands are skipped.
func (p *spl2ParsedDocument) inspectRename(ctx *spl2.RenameCommandContext) {
	sources, targets := map[string]bool{}, map[string]bool{}
	for _, pair := range ctx.AllRenamePair() {
		if pair.RenameSource() == nil || pair.RenameTarget() == nil || !spl2IntactSyntax(pair) {
			continue
		}
		source, sourceOK := spl2DecodeKey(pair.RenameSource().GetText())
		target, targetOK := spl2DecodeKey(pair.RenameTarget().GetText())
		if !sourceOK || !targetOK || strings.Contains(source+target, "*") {
			continue
		}
		if sources[source] || targets[target] || targets[source] || sources[target] {
			p.syntaxFinding(pair, CodeSyntaxError, "contract", "Rename pairs must have independent sources and targets")
		}
		sources[source], targets[target] = true, true
	}
}

func (p *spl2ParsedDocument) inspectInteger(ctx antlr.ParserRuleContext, positive bool) {
	if !spl2IntactSyntax(ctx) {
		return
	}
	value := ctx.GetText()
	if window, ok := ctx.(*spl2.WindowOptionContext); ok {
		value = window.NUMBER().GetText()
	}
	unsigned := strings.TrimPrefix(strings.TrimPrefix(value, "+"), "-")
	integer, nonzero := unsigned != "", false
	for _, digit := range unsigned {
		integer = integer && digit >= '0' && digit <= '9'
		nonzero = nonzero || digit != '0'
	}
	if !integer || positive && (!nonzero || strings.HasPrefix(value, "-")) {
		p.syntaxFinding(ctx, CodeSyntaxError, "contract", "Option requires an integer in its documented domain")
	}
}

func (p *spl2ParsedDocument) inspectUnknownOption(ctx *spl2.UnknownOptionContext) {
	if ctx.UnknownOptionName() == nil {
		return
	}
	name := ctx.UnknownOptionName().GetText()
	removed, profile := false, false
	for parent := ctx.GetParent(); parent != nil; parent = parent.GetParent() {
		switch parent.(type) {
		case *spl2.HeadCommandContext:
			removed = name == "limit" || name == "null"
		case *spl2.SortCommandContext:
			removed = name == "count"
		case *spl2.DedupCommandContext:
			removed = name == "keepevents" || name == "sortby"
		case *spl2.LookupCommandContext:
			removed = name == "local" || name == "update"
		case *spl2.StatsCommandContext:
			profile = name == "mode" || name == "prestats" || name == "annotations"
		}
	}
	if profile {
		p.syntaxComplete = false
		p.syntaxFinding(ctx, "SPL_PROFILE_MISMATCH", "compatibility", "This stats option is unavailable in the splunkd profile")
	} else if removed {
		p.syntaxComplete = false
		p.syntaxFinding(ctx, CodeSyntaxError, "syntax", "Removed SPL option is not accepted by this SPL2 command")
	} else {
		p.heldSyntax(ctx, "Unproved command option retains incomplete syntax coverage")
	}
}

func spl2StatisticalFunction(name string) bool {
	switch name {
	case "count", "sum", "avg", "min", "max", "dc", "distinct_count", "values", "list", "first", "last":
		return true
	}
	return false
}

func (p *spl2ParsedDocument) inspectAggregate(ctx *spl2.AggregateContext) {
	call := ctx.Call()
	if call == nil || call.Identifier() == nil || call.RPAREN() == nil || !spl2IntactSyntax(call) {
		return
	}
	name := call.Identifier().GetText()
	if !spl2StatisticalFunction(name) {
		return
	}
	count := 0
	if args := call.Arguments(); args != nil {
		// Named signatures remain semantically unproved, not positional errors.
		if len(args.AllNamedArgument()) > 0 {
			return
		}
		count = len(args.AllExpression())
	}
	if count > 1 || name != "count" && count != 1 {
		p.syntaxFinding(call, CodeSyntaxError, "contract", "Statistical call has an invalid positional arity")
	}
}

func spl2IntactSyntax(tree antlr.Tree) bool {
	switch node := tree.(type) {
	case antlr.ErrorNode:
		return false
	case antlr.TerminalNode:
		return node.GetSymbol().GetTokenIndex() >= 0
	case antlr.ParserRuleContext:
		if len(node.GetChildren()) == 0 {
			return false
		}
	}
	for _, child := range tree.GetChildren() {
		if !spl2IntactSyntax(child) {
			return false
		}
	}
	return true
}
