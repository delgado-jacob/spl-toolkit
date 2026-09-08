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
	parsed2 *spl2ParsedDocument
	aliases map[string]bool
}

func analyzeSPL2(result *Result, parsed *spl2ParsedDocument, refinement *sourceRefinement) {
	result.Coverage.SyntaxComplete = parsed.syntaxComplete
	result.Diagnostics = append(result.Diagnostics, parsed.diagnostics...)
	result.Scopes = append(result.Scopes, Scope{ID: "scope-0", Kind: "root", Location: parsed.source.location(0, len(parsed.source.positions)-1)})
	if !parsed.syntaxComplete {
		result.Coverage.SemanticComplete = false
	}
	sites := spl2RecoverySites(parsed, result)
	trees := []antlr.Tree{}
	for _, site := range sites {
		if site.context != nil {
			trees = append(trees, site.context)
		}
	}
	scheduler := &spl2ScopeScheduler{result: result, parsed: parsed, refinement: refinement, children: spl2ChildScopesIn(parsed, trees), executed: map[int]bool{}}
	scheduler.pipeline(sites, newEnvironment(), map[string]bool{}, "scope-0", -1)
	if len(result.Scopes) > 1 {
		spl2FinalizeStages(result)
	}

	finalizeReferences(result, refinement)
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
	return locatedOperand{Name: name, Location: s.parsed2.source.contextLocation(ctx), Resolution: "exact", Sound: true, UnresolvedSource: strings.Contains(name, ".")}
}
func (s *spl2SemanticStage) selector(ctx antlr.ParserRuleContext) locatedOperand {
	o := s.operand(ctx)
	if strings.Contains(o.Name, "*") {
		o.Resolution = "wildcard"
	}
	return o
}
func (s *spl2SemanticStage) unsupported(ctx antlr.ParserRuleContext, message string) {
	s.diagnosticAt(CodeUnsupportedSemantics, "warning", "unsupported_semantics", message, s.parsed2.source.contextLocation(ctx), true)
}
func (s *spl2SemanticStage) dependency(ctx antlr.ParserRuleContext, kind string) {
	o := s.operand(ctx)
	if s.operandReference(o, kind, "read") != "" {
		s.addDependency(o.Name, kind)
	}
}
