package analysis

import (
	"fmt"
	"github.com/antlr4-go/antlr/v4"
	"strings"
)

// SPL2 retains its own typed tree and locations. The embedded stage is only the
// shared field-transfer kernel; its legacy SPL parser pointer stays nil.
type spl2SemanticStage struct {
	*semanticStage
	parsed2               *spl2ParsedDocument
	aliases               map[string]bool
	structuralUncertainty map[string]spl2UncertaintySnapshot
}

type spl2UncertaintySnapshot struct {
	public, requirements bool
}

func analyzeSPL2(result *Result, parsed *spl2ParsedDocument, refinement *sourceRefinement, trace *requirementTrace) {
	result.Coverage.SyntaxComplete = parsed.syntaxComplete
	result.Diagnostics = append(result.Diagnostics, parsed.diagnostics...)
	result.Scopes = append(result.Scopes, Scope{ID: "scope-0", Kind: "root", Location: parsed.source.location(0, len(parsed.source.positions)-1)})
	if !parsed.syntaxComplete {
		result.Coverage.SemanticComplete = false
	}
	sites := spl2RecoverySites(parsed, result)
	initialDiagnosticCount := len(result.Diagnostics)
	trace.syntaxComplete = parsed.syntaxComplete
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
	scheduler.pipeline(sites, newEnvironmentWithRequirementTrace(trace), map[string]bool{}, "scope-0", -1)
	scheduler.syncParserDiagnostics()
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
	return locatedOperand{Name: name, Location: s.parsed2.source.contextLocation(ctx), Resolution: "exact", Sound: true, UnresolvedSource: strings.Contains(name, "."), rewrite: s.rewriteSPL2Owner(ctx)}
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
func (s *spl2SemanticStage) unsupported(ctx antlr.ParserRuleContext, message string) {
	s.diagnosticAt(CodeUnsupportedSemantics, "warning", "unsupported_semantics", message, s.parsed2.source.contextLocation(ctx), true)
}
func (s *spl2SemanticStage) unsupportedOwned(ctx antlr.ParserRuleContext, message string, pendingReferenceIDs []string) {
	s.diagnosticAtOwned(CodeUnsupportedSemantics, "warning", "unsupported_semantics", message, s.parsed2.source.contextLocation(ctx), true, pendingReferenceIDs)
}

// Remember pre-diagnostic uncertainty for a later quoted atomic field with the
// same decoded spelling. The diagnostic still updates stage uncertainty
// normally, so unrelated later reads remain conservative.
func (s *spl2SemanticStage) unsupportedStructuralReference(ctx antlr.ParserRuleContext, name, message string, pendingReferenceIDs []string) {
	if s.structuralUncertainty == nil {
		s.structuralUncertainty = map[string]spl2UncertaintySnapshot{}
	}
	if _, found := s.structuralUncertainty[name]; !found {
		s.structuralUncertainty[name] = spl2UncertaintySnapshot{public: s.env.uncertain, requirements: s.env.requirements.uncertain}
	}
	s.unsupportedOwned(ctx, message, pendingReferenceIDs)
}

// structuralFieldReference records one grammar-proven structural occurrence
// without reading from or writing to either string-keyed field environment.
func (s *spl2SemanticStage) structuralFieldReference(operand locatedOperand) string {
	if !operand.Sound {
		return ""
	}
	id := s.referenceAt(operand.Location, operand.Name, "field", "read", operand.Resolution)
	s.bindStructuralFieldReference(id, operand)
	return id
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
