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
		if from, ok := ctx.(*spl2.FromCommandContext); ok && spl2EmptyGroupAtSelect(from.SqlGroupClause()) {
			// Retry this established SQL range once with the same generated
			// parser, retaining the original diagnostics and token coordinates.
			p.tokens.Seek(boundary[0])
			local := &spl2ParsedDocument{source: p.source, tokens: p.tokens, diagnostics: []Diagnostic{}}
			parser := spl2.NewSPL2Parser(p.tokens)
			parser.RemoveErrorListeners()
			parser.AddErrorListener(&spl2SyntaxListener{DefaultErrorListener: antlr.NewDefaultErrorListener(), parsed: local})
			parser.SetErrorHandler(&spl2GroupSelectRecovery{DefaultErrorStrategy: antlr.NewDefaultErrorStrategy()})
			recovered := parser.FromCommand()
			if recovered.GetStop() != nil && recovered.GetStop().GetTokenIndex() <= boundary[1] && recovered.SqlSelectClause() != nil && spl2IntactSyntax(recovered.SqlSelectClause()) {
				p.recoveredSelects = append(p.recoveredSelects, spl2RecoveredSelect{from, recovered})
				ctx = recovered
				result.Diagnostics = append(result.Diagnostics, local.diagnostics...)
			}
		}
		if ctx != nil && ctx.GetStop() != nil && ctx.GetStop().GetTokenIndex() <= boundary[1] {
			site := spl2Sites([]antlr.ParserRuleContext{ctx}, p.source)[0]
			// Original default-channel operands after a recovered prefix remain
			// owned by this established command range. Hidden boundary comments
			// and neighboring pipes cannot enlarge the stage.
			for tokenIndex := ctx.GetStop().GetTokenIndex() + 1; tokenIndex <= boundary[1]; tokenIndex++ {
				token := p.tokens.Get(tokenIndex)
				if token.GetChannel() != antlr.TokenDefaultChannel {
					continue
				}
				tokenLocation := p.source.location(token.GetStart(), token.GetStop()+1)
				for _, d := range p.diagnostics {
					if d.Code == CodeSyntaxError && d.Category == "syntax" && d.Location.Start.Offset >= tokenLocation.Start.Offset && d.Location.Start.Offset < tokenLocation.End.Offset {
						site.location = location
					}
				}
			}
			sites = append(sites, site)
		} else {
			sites = append(sites, spl2CommandSite{location: location, command: name})
		}
	}
	return sites
}

// Record the already approved retry's original owners, never a new parse.
type spl2RecoveredSelect struct {
	original  *spl2.FromCommandContext
	recovered spl2.IFromCommandContext
}

// This local view applies only to an intact native zero-argument count in the
// recorded SELECT retry. Its original missing-SELECT diagnostic remains in the
// document and every other original/retry error still constrains the output.
func (p *spl2ParsedDocument) recoveredCountView(ctx antlr.ParserRuleContext) (*spl2ParsedDocument, *Diagnostic) {
	projection, ok := ctx.(spl2.IProjectionContext)
	if !ok || !spl2IntactSyntax(projection) {
		return p, nil
	}
	access := spl2SQLFieldAccess(projection.Expression())
	if access == nil || len(access.AllAccessPart()) != 0 || access.Primary().Call() == nil {
		return p, nil
	}
	call := access.Primary().Call()
	if call.Identifier().GetText() != "count" || call.Arguments() != nil {
		return p, nil
	}
	for _, retry := range p.recoveredSelects {
		original, recovered := retry.original, retry.recovered
		if projection.GetParent() != recovered.SqlSelectClause() || original.SqlSelectClause() != nil || !spl2EmptyGroupAtSelect(original.SqlGroupClause()) || original.GetStart() != recovered.GetStart() || original.GetStop() != recovered.GetStop() {
			continue
		}
		originalRange := true
		for _, token := range []antlr.Token{recovered.GetStart(), recovered.GetStop(), recovered.SqlSelectClause().GetStart(), projection.GetStart(), projection.GetStop()} {
			if token == nil || token.GetTokenIndex() < 0 || token.GetTokenIndex() >= p.tokens.Size() || p.tokens.Get(token.GetTokenIndex()) != token || token.GetTokenIndex() < original.GetStart().GetTokenIndex() || token.GetTokenIndex() > original.GetStop().GetTokenIndex() {
				originalRange = false
			}
		}
		for parent := original.GetParent(); parent != nil; parent = parent.GetParent() {
			switch parent.(type) {
			case *spl2.IndependentSearchContext, *spl2.InheritedSubpipeContext, *spl2.ExistsPredicateContext:
				originalRange = false
			}
		}
		if !originalRange {
			continue
		}
		for _, mismatch := range p.missingSelectEOF {
			token := mismatch.token
			if mismatch.owner != original || token.GetTokenIndex() < 0 || token.GetTokenIndex() >= p.tokens.Size() || p.tokens.Get(token.GetTokenIndex()) != token || token.GetStart() != len(p.source.positions)-1 || token.GetStop()+1 != token.GetStart() {
				continue
			}
			local := *p
			local.diagnostics = []Diagnostic{}
			for _, d := range p.diagnostics {
				if d != mismatch.diagnostic {
					local.diagnostics = append(local.diagnostics, d)
				}
			}
			return &local, &mismatch.diagnostic
		}
	}
	return p, nil
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

func spl2EmptyGroupAtSelect(group spl2.ISqlGroupClauseContext) bool {
	if group == nil {
		return false
	}
	for _, key := range group.AllSqlGroupKey() {
		if key.GetStart() != nil && key.GetStart().GetTokenType() == spl2.SPL2ParserSELECT {
			return true
		}
	}
	return false
}

// SELECT at entry to a GROUP key is a real next-clause boundary. Only this
// empty-key site can unwind without consuming it; nested expression recovery
// still uses ANTLR's ordinary strategy.
type spl2GroupSelectRecovery struct{ *antlr.DefaultErrorStrategy }

func (r *spl2GroupSelectRecovery) boundary(p antlr.Parser) bool {
	key, ok := p.GetParserRuleContext().(*spl2.SqlGroupKeyContext)
	return ok && len(key.GetChildren()) == 0 && p.GetCurrentToken().GetTokenType() == spl2.SPL2ParserSELECT && key.GetStart() == p.GetCurrentToken()
}
func (r *spl2GroupSelectRecovery) Sync(p antlr.Parser) {
	if r.boundary(p) {
		p.SetError(antlr.NewInputMisMatchException(p))
		return
	}
	r.DefaultErrorStrategy.Sync(p)
}
func (r *spl2GroupSelectRecovery) Recover(p antlr.Parser, e antlr.RecognitionException) {
	if r.boundary(p) {
		return
	}
	r.DefaultErrorStrategy.Recover(p, e)
}

// A single bare key before an actually empty HAVING can be rejected by ANTLR's
// full-context prediction. Prove only that original key with the same Identifier
// rule, without replacing the raw key tree or repairing the SQL text.
type spl2ProvedGroupKey struct {
	key        spl2.ISqlGroupKeyContext
	identifier spl2.IIdentifierContext
	diagnostic Diagnostic
}

func (p *spl2ParsedDocument) proveGroupKeyBeforeEmptyHaving(c spl2SQLCommand) *spl2ProvedGroupKey {
	group, having := c.SqlGroupClause(), c.SqlHavingClause()
	if group == nil || having == nil || having.GetStart() == nil || having.GetStart().GetTokenType() != spl2.SPL2ParserHAVING || spl2IntactSyntax(having) {
		return nil
	}
	keys := group.AllSqlGroupKey()
	if len(keys) == 0 {
		return nil
	}
	key := keys[len(keys)-1]
	first, last := key.GetStart(), key.GetStop()
	if first == nil || first != last || first.GetTokenIndex() < 0 || first.GetTokenType() != spl2.SPL2ParserIDENTIFIER {
		return nil
	}
	// Prove every original GROUP/BY token and every preceding key independently.
	if group.GROUP() != nil {
		if group.GROUP().GetSymbol().GetTokenIndex() < 0 || group.SqlBy() == nil || !spl2IntactSyntax(group.SqlBy()) {
			return nil
		}
	} else if group.GROUPBY() == nil || group.GROUPBY().GetSymbol().GetTokenIndex() < 0 {
		return nil
	}
	for _, earlier := range keys[:len(keys)-1] {
		if !p.soundOperand(earlier) {
			return nil
		}
	}
	next := func(index int) antlr.Token {
		for i := index + 1; i < p.tokens.Size(); i++ {
			token := p.tokens.Get(i)
			if token.GetChannel() == antlr.TokenDefaultChannel {
				return token
			}
		}
		return nil
	}
	if next(first.GetTokenIndex()) != having.GetStart() {
		return nil
	}
	after := next(having.GetStart().GetTokenIndex())
	if after == nil {
		return nil
	}
	switch after.GetTokenType() {
	case antlr.TokenEOF, spl2.SPL2ParserPIPE, spl2.SPL2ParserSEMI, spl2.SPL2ParserRPAREN:
	default:
		return nil
	}
	location := p.source.contextLocation(key)
	var prediction *spl2PredictionError
	for i := range p.predictionErrors {
		e := &p.predictionErrors[i]
		if e.diagnostic.Location == location && spl2Within(e.context, key) {
			if prediction != nil {
				return nil
			}
			prediction = e
		}
	}
	if prediction == nil {
		return nil
	}
	groupLocation := p.source.contextLocation(group)
	for _, d := range p.diagnostics {
		if d != prediction.diagnostic && d.Location.Start.Offset < groupLocation.End.Offset && d.Location.End.Offset >= groupLocation.Start.Offset {
			return nil
		}
	}
	saved := p.tokens.Index()
	defer p.tokens.Seek(saved)
	p.tokens.Seek(first.GetTokenIndex())
	local := &spl2ParsedDocument{source: p.source, tokens: p.tokens}
	parser := spl2.NewSPL2Parser(p.tokens)
	parser.RemoveErrorListeners()
	parser.AddErrorListener(&spl2SyntaxListener{DefaultErrorListener: antlr.NewDefaultErrorListener(), parsed: local})
	identifier := parser.Identifier()
	if len(local.diagnostics) != 0 || !spl2IntactSyntax(identifier) || identifier.GetStart() != first || identifier.GetStop() != last || parser.GetCurrentToken() != having.GetStart() {
		return nil
	}
	return &spl2ProvedGroupKey{key, identifier, prediction.diagnostic}
}

// Only original document EOF attributed to this FROM owner's missing SELECT
// can be excluded from an otherwise intact, already surviving GROUP operand.
func (p *spl2ParsedDocument) groupMissingSelectDiagnostic(c spl2SQLCommand) *Diagnostic {
	owner, ok := c.(*spl2.FromCommandContext)
	group := c.SqlGroupClause()
	if !ok || c.SqlSelectClause() != nil || group == nil || !spl2IntactSyntax(group) || len(group.AllSqlGroupKey()) == 0 {
		return nil
	}
	for parent := owner.GetParent(); parent != nil; parent = parent.GetParent() {
		switch parent.(type) {
		case *spl2.IndependentSearchContext, *spl2.InheritedSubpipeContext, *spl2.ExistsPredicateContext:
			return nil
		}
	}
	var found *Diagnostic
	for _, mismatch := range p.missingSelectEOF {
		token := mismatch.token
		if mismatch.owner != owner || token.GetTokenIndex() < 0 || token.GetTokenIndex() >= p.tokens.Size() || p.tokens.Get(token.GetTokenIndex()) != token || token.GetStart() != len(p.source.positions)-1 || token.GetStop()+1 != token.GetStart() {
			continue
		}
		if found != nil {
			return nil
		}
		d := mismatch.diagnostic
		found = &d
	}
	if found == nil {
		return nil
	}
	location := p.source.contextLocation(group)
	for _, d := range p.diagnostics {
		if d != *found && d.Location.End.Offset >= location.Start.Offset && d.Location.Start.Offset <= location.End.Offset {
			return nil
		}
	}
	return found
}
