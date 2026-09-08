package analysis

import (
	"fmt"
	"github.com/antlr4-go/antlr/v4"
	"github.com/delgado-jacob/spl-toolkit/parser/spl2"
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
	pipeline := parsed.tree.Pipeline()
	if pipeline == nil {
		result.Coverage.SemanticComplete = false
		return
	}
	contexts := []antlr.ParserRuleContext{}
	for _, node := range pipeline.GetChildren() {
		switch c := node.(type) {
		case *spl2.StartContext, *spl2.CommandContext:
			for _, child := range c.GetChildren() {
				if ctx, ok := child.(antlr.ParserRuleContext); ok {
					contexts = append(contexts, ctx)
					break
				}
			}
		}
	}
	env := newEnvironment()
	aliases := map[string]bool{}
	for _, ctx := range contexts {
		location := parsed.source.contextLocation(ctx)
		command := strings.ToLower(ctx.GetStart().GetText())
		if _, ok := ctx.(*spl2.ImplicitSearchContext); ok {
			command = "search"
		}
		index := len(result.Stages)
		id := fmt.Sprintf("stage-%d", index)
		result.Stages = append(result.Stages, Stage{ID: id, Command: command, Position: index, ScopeID: "scope-0", Location: location, SemanticComplete: true})
		s := &spl2SemanticStage{semanticStage: &semanticStage{result: result, stage: index, env: env, transitions: []Transition{}, refinement: refinement}, parsed2: parsed, aliases: aliases}
		before := env.snapshot()
		for i := range result.Diagnostics {
			d := &result.Diagnostics[i]
			if d.StageID == "" && d.Location.Start.Offset >= location.Start.Offset && d.Location.Start.Offset <= location.End.Offset {
				d.StageID = id
				d.ScopeID = "scope-0"
				result.Stages[index].SemanticComplete = false
			}
		}
		if spl2IntactSyntax(ctx) {
			s.command(ctx)
		} else {
			s.diagnosticAt(CodeUnsupportedSemantics, "warning", "unsupported_semantics", "Recovered SPL2 command effects are not yet modeled", location, true)
		}
		if !result.Stages[index].SemanticComplete {
			s.env.uncertain = true
		}
		env = s.env
		result.Lineage = append(result.Lineage, Lineage{StageID: id, ScopeID: "scope-0", Before: before, After: env.snapshot(), Transitions: s.transitions})
	}
	finalizeReferences(result, refinement)
}
func (s *spl2SemanticStage) operand(ctx antlr.ParserRuleContext) locatedOperand {
	if ctx == nil || !spl2IntactSyntax(ctx) {
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
