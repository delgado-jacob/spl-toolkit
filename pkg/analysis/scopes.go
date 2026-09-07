package analysis

import (
	"fmt"
	"github.com/antlr4-go/antlr/v4"
	"github.com/delgado-jacob/spl-toolkit/parser"
	"strings"
)

func analyzeParsed(result *Result, parsed *parsedDocument, refinement *sourceRefinement) {
	root := Scope{ID: "scope-0", Kind: "root", Location: parsed.source.location(0, len(parsed.source.positions)-1)}
	result.Scopes = append(result.Scopes, root)
	// Recovery may leave a generic or structurally partial command context. Attribute
	// errors through the next intact stage boundary before invoking typed handlers.
	stageLocations := []Location{}
	var collectStages func(antlr.Tree)
	collectStages = func(n antlr.Tree) {
		switch c := n.(type) {
		case parser.IAnalysisStageContext:
			stageLocations = append(stageLocations, parsed.source.contextLocation(c))
		case *parser.AnalysisImplicitSearchContext:
			stageLocations = append(stageLocations, parsed.source.contextLocation(c))
		}
		for i := 0; i < n.GetChildCount(); i++ {
			collectStages(n.GetChild(i))
		}
	}
	collectStages(parsed.tree)
	damagedStage := func(location Location) bool {
		end := len(result.Document.Text) + 1
		for _, loc := range stageLocations {
			if loc.Start.Offset > location.End.Offset && loc.Start.Offset < end {
				end = loc.Start.Offset
			}
		}
		for _, d := range parsed.diagnostics {
			if d.Location.Start.Offset >= location.Start.Offset && d.Location.Start.Offset < end {
				return true
			}
		}
		return false
	}
	unsafeBoundaries, tokenOwners := originalTokenBoundaries(parsed)
	scopeOwners := map[string]int{root.ID: -1}
	positions := map[string]int{}
	environments := map[string]*environment{root.ID: newEnvironment()}
	branchInputs := map[string]*environment{}
	var visit func(antlr.Tree, string, string)
	visit = func(node antlr.Tree, scopeID, stageID string) {
		var stageContext antlr.ParserRuleContext
		command := ""
		switch ctx := node.(type) {
		case parser.IAnalysisStageContext:
			stageContext = ctx
			command = strings.ToLower(ctx.GetStart().GetText())
			if _, macro := ctx.(*parser.AnalysisMacroStageContext); macro {
				command = "macro"
			}
		case *parser.AnalysisInitialStageContext:
			if ctx.AnalysisStage() == nil && ctx.AnalysisImplicitSearch() == nil {
				stageContext = ctx
				command = "search"
				switch ctx.GetStart().GetTokenType() {
				case parser.SPLLexerINIT_COMMAND, parser.SPLLexerSTD_COMMAND, parser.SPLLexerSTD_COMMAND_AND_FUNCTION:
					command = strings.ToLower(ctx.GetStart().GetText())
				}
			}
		case *parser.AnalysisImplicitSearchContext:
			stageContext = ctx
			command = "search"
		case *parser.AnalysisSubqueryContext:
			// A recovered subtree can lose or synthesize its opening bracket. Never
			// assign it a scope unless the original token's owner matches its parent.
			start := ctx.GetStart()
			if owner, ok := tokenOwners[start.GetTokenIndex()]; !ok || owner != scopeOwners[scopeID] || start.GetTokenType() != parser.SPLLexerLBRACK {
				environments[scopeID].uncertain = true
				return
			}
			kind := "subsearch"
			for _, stage := range result.Stages {
				if stage.ID == stageID && (stage.Command == "join" || stage.Command == "append" || stage.Command == "appendpipe") {
					kind = stage.Command
					break
				}
			}
			child := Scope{ID: fmt.Sprintf("scope-%d", len(result.Scopes)), ParentID: scopeID, Kind: kind, StageID: stageID, Location: parsed.source.contextLocation(ctx)}
			result.Scopes = append(result.Scopes, child)
			scopeOwners[child.ID] = start.GetTokenIndex()
			env := newEnvironment()
			if kind == "appendpipe" {
				env = branchInputs[stageID].clone()
			}
			scopeID = child.ID
			environments[scopeID] = env
		}
		if stageContext != nil {
			// ANTLR recovery may flatten a damaged child stage into the parent's
			// pipeline. Withhold its entire subtree before publishing a stage or
			// running transfers/dependency collection against the wrong environment.
			if owner, ok := tokenOwners[stageContext.GetStart().GetTokenIndex()]; !ok || owner != scopeOwners[scopeID] {
				environments[scopeID].uncertain = true
				return
			}
			stageID = fmt.Sprintf("stage-%d", len(result.Stages))
			stage := Stage{ID: stageID, Command: command, Position: positions[scopeID], ScopeID: scopeID, Location: parsed.source.contextLocation(stageContext), SemanticComplete: true}
			positions[scopeID]++
			result.Stages = append(result.Stages, stage)
			state := &semanticStage{result: result, parsed: parsed, stage: len(result.Stages) - 1, env: environments[scopeID], transitions: []Transition{}, refinement: refinement}
			before := state.env.snapshot()
			branchInputs[stageID] = state.env.clone()
			if _, damaged := stageContext.(*parser.AnalysisStageContext); damaged || damagedStage(stage.Location) || unsafeBoundaries[stageContext.GetStart().GetTokenIndex()] {
				if !unsafeBoundaries[stageContext.GetStart().GetTokenIndex()] {
					state.recoverStage(stageContext)
				}
				result.Stages[state.stage].SemanticComplete = false
				state.env.uncertain = true
			} else if _, macro := stageContext.(*parser.AnalysisMacroStageContext); macro {
				state.diagnostic(CodeDynamicReference, "macro expansion is unresolved", stageContext)
			} else if spec, ok := commands[command]; ok && spec.handle != nil {
				spec.handle(state, stageContext)
			} else {
				code := CodeUnsupportedCommand
				if ok {
					code = CodeUnsupportedSemantics
				}
				state.diagnostic(code, fmt.Sprintf("command %q has unmodeled field effects", command), stageContext)
			}
			if !unsafeBoundaries[stageContext.GetStart().GetTokenIndex()] {
				state.dependencies(stageContext)
			}
			if !intact(stageContext) {
				result.Stages[state.stage].SemanticComplete = false
				state.env.uncertain = true
			}
			environments[scopeID] = state.env
			result.Lineage = append(result.Lineage, Lineage{StageID: stageID, ScopeID: scopeID, Before: before, After: state.env.snapshot(), Transitions: state.transitions})
		}
		for i := 0; i < node.GetChildCount(); i++ {
			visit(node.GetChild(i), scopeID, stageID)
		}
	}
	visit(parsed.tree, root.ID, "")
	finalizeReferences(result, refinement)
}
