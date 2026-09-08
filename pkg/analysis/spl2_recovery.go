package analysis

import (
	"strings"

	"github.com/antlr4-go/antlr/v4"
	"github.com/delgado-jacob/spl-toolkit/parser/spl2"
)

// Recovery sites have original token locations and optionally a real generated
// command context. An unsupported command has no invented AST or field operands.
type spl2CommandSite struct {
	context  antlr.ParserRuleContext
	location Location
	command  string
}

func spl2Sites(contexts []antlr.ParserRuleContext, source *sourceIndex) []spl2CommandSite {
	sites := []spl2CommandSite{}
	for _, ctx := range contexts {
		command := strings.ToLower(ctx.GetStart().GetText())
		if _, ok := ctx.(*spl2.ImplicitSearchContext); ok {
			command = "search"
		}
		sites = append(sites, spl2CommandSite{ctx, source.contextLocation(ctx), command})
	}
	return sites
}

// Original lexer token kinds retain all quote/comment/regex modes. Only a pipe
// outside every original delimiter can authorize a root stage. An unmatched
// delimiter blocks later ownership, even if ANTLR inserted a closing token.
func spl2RootTokenRanges(p *spl2ParsedDocument) [][2]int {
	p.tokens.Fill()
	ranges := [][2]int{}
	start, end := -1, -1
	stack := []int{}
	for _, token := range p.tokens.GetAllTokens() {
		if token.GetChannel() != antlr.TokenDefaultChannel || token.GetTokenType() == antlr.TokenEOF {
			continue
		}
		kind, index := token.GetTokenType(), token.GetTokenIndex()
		if kind == spl2.SPL2ParserNL {
			continue
		}
		if len(stack) == 0 && (kind == spl2.SPL2ParserPIPE || kind == spl2.SPL2ParserSEMI) {
			if start >= 0 {
				ranges = append(ranges, [2]int{start, end})
			}
			start, end = -1, -1
			if kind == spl2.SPL2ParserSEMI {
				break
			}
			continue
		}
		if start < 0 {
			start = index
		}
		end = index
		switch kind {
		case spl2.SPL2ParserLPAREN:
			stack = append(stack, spl2.SPL2ParserRPAREN)
		case spl2.SPL2ParserLBRACKET:
			stack = append(stack, spl2.SPL2ParserRBRACKET)
		case spl2.SPL2ParserLBRACE, spl2.SPL2ParserSTRING_INTERPOLATION, spl2.SPL2ParserNAME_INTERPOLATION:
			stack = append(stack, spl2.SPL2ParserRBRACE)
		case spl2.SPL2ParserDQUOTE:
			stack = append(stack, spl2.SPL2ParserSTRING_END)
		case spl2.SPL2ParserSQUOTE:
			stack = append(stack, spl2.SPL2ParserNAME_END)
		case spl2.SPL2ParserBACKTICK:
			stack = append(stack, spl2.SPL2ParserEMBEDDED_END)
		case spl2.SPL2ParserRPAREN, spl2.SPL2ParserRBRACKET, spl2.SPL2ParserRBRACE, spl2.SPL2ParserSTRING_END, spl2.SPL2ParserNAME_END, spl2.SPL2ParserEMBEDDED_END:
			if len(stack) > 0 && stack[len(stack)-1] == kind {
				stack = stack[:len(stack)-1]
			} else {
				stack = append(stack, -1)
			}
		}
	}
	if start >= 0 {
		ranges = append(ranges, [2]int{start, end})
	}
	return ranges
}

func spl2RecoverySites(p *spl2ParsedDocument, result *Result) []spl2CommandSite {
	ordinary := spl2Sites(spl2PipelineContexts(p.tree.Pipeline()), p.source)
	if p.syntaxComplete || p.tree.ModuleDeclaration() != nil {
		return ordinary
	}
	ranges := spl2RootTokenRanges(p)
	lexicalErrors := map[Diagnostic]bool{}
	for _, d := range p.lexicalErrors {
		lexicalErrors[d] = true
	}
	existing := map[int]antlr.ParserRuleContext{}
	for _, site := range ordinary {
		existing[site.context.GetStart().GetTokenIndex()] = site.context
	}
	sites := []spl2CommandSite{}
	saved := p.tokens.Index()
	defer p.tokens.Seek(saved)
	for i, boundary := range ranges {
		first, last := p.tokens.Get(boundary[0]), p.tokens.Get(boundary[1])
		location := p.source.location(first.GetStart(), last.GetStop()+1)
		name := strings.ToLower(first.GetText())
		// The grammar reserves supported command tokens. An ordinary identifier in
		// a command slot is a located unknown/native-deferred/profile command.
		if first.GetTokenType() == spl2.SPL2ParserIDENTIFIER {
			code, severity, category, message := CodeUnsupportedSemantics, "warning", "unsupported_semantics", "Standalone command syntax and effects are unproved"
			if name == "decrypt" || name == "ocsf" || name == "route" {
				code, severity, category, message = "SPL_PROFILE_MISMATCH", "error", "compatibility", "Command is unavailable in the splunkd profile"
			}
			kept := result.Diagnostics[:0]
			for _, d := range result.Diagnostics {
				if lexicalErrors[d] || d.Code != CodeSyntaxError || d.Location.Start.Offset < location.Start.Offset || d.Location.Start.Offset > location.End.Offset {
					kept = append(kept, d)
				}
			}
			result.Diagnostics = append(kept, Diagnostic{Code: code, Severity: severity, Category: category, Message: message, Location: location})
			sites = append(sites, spl2CommandSite{location: location, command: name})
			continue
		}
		ctx := existing[boundary[0]]
		if ctx == nil {
			p.tokens.Seek(boundary[0])
			local := &spl2ParsedDocument{source: p.source, tokens: p.tokens, diagnostics: []Diagnostic{}, syntaxComplete: true}
			parser := spl2.NewSPL2Parser(p.tokens)
			parser.RemoveErrorListeners()
			parser.AddErrorListener(&spl2SyntaxListener{DefaultErrorListener: antlr.NewDefaultErrorListener(), parsed: local})
			var routing antlr.ParserRuleContext
			if i == 0 {
				routing = parser.Start_()
			} else {
				routing = parser.Command()
			}
			for _, child := range routing.GetChildren() {
				if candidate, ok := child.(antlr.ParserRuleContext); ok {
					ctx = candidate
					break
				}
			}
			if ctx != nil {
				local.inspectSyntax(ctx, 0)
			}
			result.Diagnostics = append(result.Diagnostics, local.diagnostics...)
		}
		if ctx != nil && ctx.GetStop() != nil && ctx.GetStop().GetTokenIndex() <= boundary[1] {
			site := spl2Sites([]antlr.ParserRuleContext{ctx}, p.source)[0]
			sites = append(sites, site)
		} else {
			sites = append(sites, spl2CommandSite{location: location, command: name})
		}
	}
	return sites
}

// Lexer damage can sit just outside a surviving name token. Keep that name
// untrusted rather than silently discarding the offending character.
func (p *spl2ParsedDocument) soundOperand(ctx antlr.ParserRuleContext) bool {
	if ctx == nil || !spl2IntactSyntax(ctx) {
		return false
	}
	location := p.source.contextLocation(ctx)
	for _, d := range p.diagnostics {
		if d.Code == CodeSyntaxError && d.Category == "syntax" && d.Location.End.Offset >= location.Start.Offset && d.Location.Start.Offset <= location.End.Offset {
			return false
		}
	}
	return true
}
