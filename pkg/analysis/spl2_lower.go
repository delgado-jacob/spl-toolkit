package analysis

import (
	"fmt"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	"github.com/delgado-jacob/spl-toolkit/parser/spl2"
)

// Recognition is kept separate from a proved source environment. Non-SQL and
// correlated aliases retain their existing qualifier behavior without claiming
// ownership; SQL joins attach only the environment established for that alias.
type spl2SourceAlias struct {
	environment *environment
	conditional bool
}
type spl2Aliases map[string]spl2SourceAlias

func (a spl2Aliases) recognizes(name string) bool { _, ok := a[name]; return ok }

// SPL2 retains its own typed tree and locations. The embedded stage is only the
// shared field-transfer kernel; its legacy SPL parser pointer stays nil.
type spl2SemanticStage struct {
	*semanticStage
	parsed2                    *spl2ParsedDocument
	aliases                    spl2Aliases
	locals                     map[string]bool
	functionSummary            *spl2FunctionSummary
	program                    *spl2Program
	readRole                   string
	sqlVisibility              map[string]bool
	suppressLocalCallReference bool
}

func analyzeSPL2(result *Result, parsed *spl2ParsedDocument, refinement *sourceRefinement, trace *requirementTrace) {
	diagnostics, syntaxComplete := spl2BoundProgramDiagnostics(parsed)
	result.Coverage.SyntaxComplete = syntaxComplete
	result.Diagnostics = append(result.Diagnostics, diagnostics...)
	result.Scopes = append(result.Scopes, Scope{ID: "scope-0", Kind: "root", Location: parsed.source.location(0, len(parsed.source.positions)-1)})
	if !syntaxComplete {
		result.Coverage.SemanticComplete = false
	}
	trace.syntaxComplete = syntaxComplete
	if spl2BindableProgram(parsed) {
		initialDiagnosticCount := len(result.Diagnostics)
		for _, diagnostic := range result.Diagnostics[:initialDiagnosticCount] {
			trace.recordDiagnostic(diagnostic, true, nil, trace.nextEvent())
		}
		program := collectSPL2Program(parsed)
		program.bind(result, refinement, trace, initialDiagnosticCount)
	} else {
		sites := spl2RecoverySites(parsed, result)
		initialDiagnosticCount := len(result.Diagnostics)
		for _, diagnostic := range result.Diagnostics[:initialDiagnosticCount] {
			trace.recordDiagnostic(diagnostic, true, nil, trace.nextEvent())
		}
		trees := []antlr.Tree{}
		for _, site := range sites {
			if site.context != nil {
				trees = append(trees, site.context)
			}
		}
		scheduler := &spl2ScopeScheduler{result: result, parsed: parsed, refinement: refinement, trace: trace, initialDiagnosticCount: initialDiagnosticCount, children: spl2ChildScopesIn(parsed, trees), executed: map[int]bool{}}
		scheduler.pipeline(sites, newEnvironmentWithRequirementTrace(trace), spl2Aliases{}, "scope-0", -1)
		scheduler.syncParserDiagnostics()
	}
	if len(result.Scopes) > 1 {
		trace.remapStages(spl2FinalizeStages(result))
	}

	finalizeReferences(result, refinement, trace)
}

func registerSPL2Stage(result *Result, location Location, command string, position int, scopeID string) int {
	index := len(result.Stages)
	id := fmt.Sprintf("stage-%d", index)
	result.Stages = append(result.Stages, Stage{ID: id, Command: command, Position: position, ScopeID: scopeID, Location: location, SemanticComplete: true})
	for i := range result.Diagnostics {
		d := &result.Diagnostics[i]
		if d.StageID == "" && d.Location.Start.Offset >= location.Start.Offset && d.Location.Start.Offset <= location.End.Offset {
			d.StageID = id
			d.ScopeID = scopeID
			result.Stages[index].SemanticComplete = false
		}
	}
	return index
}
func (s *spl2SemanticStage) operand(ctx antlr.ParserRuleContext) locatedOperand {
	if ctx == nil || !s.parsed2.soundOperand(ctx) {
		return locatedOperand{}
	}
	name, ok := spl2DecodeKey(ctx.GetText())
	if !ok {
		s.unsupported(ctx, "Identifier decoding is unproved")
		return locatedOperand{}
	}
	return locatedOperand{Name: name, Identity: atomicFieldIdentity(name), Location: s.parsed2.source.contextLocation(ctx), Resolution: "exact", Sound: true, UnresolvedSource: strings.Contains(name, "."), rewrite: s.rewriteSPL2Owner(ctx)}
}
func (s *spl2SemanticStage) selector(ctx antlr.ParserRuleContext) locatedOperand {
	o := s.operand(ctx)
	if o.rewrite.role != "rename_input" {
		o.rewrite.role = "selector_atom"
	}
	if strings.Contains(o.Name, "*") {
		o.Resolution = "wildcard"
	}
	return o
}

func (s *spl2SemanticStage) structuralSelector(ctx spl2.IStructuralFieldSelectorContext) locatedOperand {
	if ctx == nil || !spl2IntactSyntax(ctx) || ctx.Identifier() == nil || ctx.FieldPath() == nil {
		return locatedOperand{}
	}
	identifiers := []spl2.IIdentifierContext{ctx.Identifier()}
	identifiers = append(identifiers, ctx.FieldPath().AllIdentifier()...)
	segments := make([]string, 0, len(identifiers))
	for _, identifier := range identifiers {
		part := s.operand(identifier)
		if !part.Sound {
			return locatedOperand{}
		}
		segments = append(segments, part.Name)
	}
	identity := pathFieldIdentity("", segments)
	owner := rewriteOwner{role: "navigation", location: s.parsed2.source.contextLocation(ctx), identity: RewriteIdentity{Path: append([]string{}, segments...)}}
	return locatedOperand{Name: identity.PublicName, Identity: identity, Location: s.parsed2.source.contextLocation(ctx), Resolution: "exact", Sound: true, UnresolvedSource: true, rewrite: owner}
}
func (s *spl2SemanticStage) unsupported(ctx antlr.ParserRuleContext, message string) {
	s.diagnosticAt(CodeUnsupportedSemantics, "warning", "unsupported_semantics", message, s.parsed2.source.contextLocation(ctx), true)
}
func (s *spl2SemanticStage) unsupportedOwned(ctx antlr.ParserRuleContext, message string, pendingReferenceIDs []string) {
	s.diagnosticAtOwned(CodeUnsupportedSemantics, "warning", "unsupported_semantics", message, s.parsed2.source.contextLocation(ctx), true, pendingReferenceIDs)
}

// structuralFieldReference routes grammar-proven structural occurrences through
// the same shared environment as atomic SPL and SPL2 fields.
func (s *spl2SemanticStage) structuralFieldReference(operand locatedOperand) string {
	if !operand.Sound {
		return ""
	}
	return s.readAt(operand, "read")
}

func (s *spl2SemanticStage) bindStructuralFieldReference(id string, operand locatedOperand) {
	ref := &s.result.References[len(s.result.References)-1]
	ref.Binding = "indeterminate"
	ref.OriginReferenceIDs = copyIDs(ref.OriginReferenceIDs)
	if trace := s.env.requirements.trace; trace != nil {
		entry := trace.reference(id)
		entry.reference.Binding = "indeterminate"
		entry.reference.OriginReferenceIDs = copyIDs(ref.OriginReferenceIDs)
		entry.directExternal = false
		entry.conditional = true
	}
	s.rewriteReference(id, operand, "field", "read")
	s.rewriteBinding(id, "indeterminate", copyIDs(ref.OriginReferenceIDs))
}
func (s *spl2SemanticStage) dependency(ctx antlr.ParserRuleContext, kind string) {
	o := s.operand(ctx)
	if s.operandReference(o, kind, "read") != "" {
		s.addDependency(o.Name, kind)
	}
}
