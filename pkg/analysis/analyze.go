// Package analysis exposes deterministic, source-aware analysis of the supported SPL contract.
package analysis

import (
	"fmt"
	"github.com/antlr4-go/antlr/v4"
	"github.com/delgado-jacob/spl-toolkit/parser"
	"sort"
	"strings"
)

func normalizeDocument(document QueryDocument) (QueryDocument, error) {
	if document.Language == "" {
		document.Language = "spl"
	}
	if document.Language != "spl" {
		return document, fmt.Errorf("unsupported language %q", document.Language)
	}
	if document.Profile == "" {
		document.Profile = "splunkd"
	}
	if document.Profile != "splunkd" {
		return document, fmt.Errorf("unsupported profile %q", document.Profile)
	}
	if document.Version == "" {
		document.Version = "current"
	}
	if document.Version != "current" {
		return document, fmt.Errorf("unsupported compatibility version %q", document.Version)
	}
	return document, nil
}
func newResult(document QueryDocument) *Result {
	return &Result{SchemaVersion: 1, Document: document, Coverage: Coverage{SyntaxComplete: true, SemanticComplete: true, Reasons: []string{}}, Stages: []Stage{}, Scopes: []Scope{}, References: []Reference{}, Lineage: []Lineage{}, Dependencies: Dependencies{Indexes: []string{}, Sources: []string{}, SourceTypes: []string{}, Datasets: []string{}, Lookups: []string{}, DataModels: []string{}, Macros: []string{}}, Diagnostics: []Diagnostic{}}
}

// Analyze retains original source and reports syntax errors as findings. Unsupported document options are API errors.
func Analyze(document QueryDocument) (*Result, error) {
	normalized, err := normalizeDocument(document)
	if err != nil {
		return nil, err
	}
	result := newResult(normalized)
	parsed := parseDocument(normalized.Text)
	result.Diagnostics = append(result.Diagnostics, parsed.diagnostics...)
	analyzeParsed(result, parsed)
	finalizeResult(result)
	return result, nil
}
func analyzeParsed(result *Result, parsed *parsedDocument) {
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
	positions := map[string]int{}
	environments := map[string]*environment{root.ID: newEnvironment()}
	var visit func(antlr.Tree, string, string)
	visit = func(node antlr.Tree, scopeID, stageID string) {
		var stageContext antlr.ParserRuleContext
		command := ""
		switch ctx := node.(type) {
		case parser.IAnalysisStageContext:
			stageContext = ctx
			command = strings.ToLower(ctx.GetStart().GetText())
		case *parser.AnalysisImplicitSearchContext:
			stageContext = ctx
			command = "search"
		case *parser.AnalysisSubqueryContext:
			kind := "subsearch"
			for _, stage := range result.Stages {
				if stage.ID == stageID && (stage.Command == "join" || stage.Command == "append" || stage.Command == "appendpipe") {
					kind = stage.Command
					break
				}
			}
			child := Scope{ID: fmt.Sprintf("scope-%d", len(result.Scopes)), ParentID: scopeID, Kind: kind, StageID: stageID, Location: parsed.source.contextLocation(ctx)}
			result.Scopes = append(result.Scopes, child)
			env := newEnvironment()
			if kind == "appendpipe" {
				env = environments[scopeID].clone()
			}
			scopeID = child.ID
			environments[scopeID] = env
		}
		if stageContext != nil {
			stageID = fmt.Sprintf("stage-%d", len(result.Stages))
			stage := Stage{ID: stageID, Command: command, Position: positions[scopeID], ScopeID: scopeID, Location: parsed.source.contextLocation(stageContext), SemanticComplete: true}
			positions[scopeID]++
			result.Stages = append(result.Stages, stage)
			state := &semanticStage{result: result, parsed: parsed, stage: len(result.Stages) - 1, env: environments[scopeID], transitions: []Transition{}}
			before := state.env.snapshot()
			if _, damaged := stageContext.(*parser.AnalysisStageContext); damaged || damagedStage(stage.Location) {
				result.Stages[state.stage].SemanticComplete = false
				state.env.uncertain = true
			} else if spec, ok := commands[command]; ok && spec.handle != nil {
				spec.handle(state, stageContext)
			} else {
				code := CodeUnsupportedCommand
				if ok {
					code = CodeUnsupportedSemantics
				}
				state.diagnostic(code, fmt.Sprintf("command %q has unmodeled field effects", command), stageContext)
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
	finalizeReferences(result)
}
func finalizeResult(result *Result) {
	sort.SliceStable(result.Diagnostics, func(i, j int) bool {
		a, b := result.Diagnostics[i], result.Diagnostics[j]
		if a.Location.Start.Offset != b.Location.Start.Offset {
			return a.Location.Start.Offset < b.Location.Start.Offset
		}
		if a.Code != b.Code {
			return a.Code < b.Code
		}
		return a.Message < b.Message
	})
	result.Status = Valid
	seen := map[string]bool{}
	for _, stage := range result.Stages {
		if !stage.SemanticComplete {
			result.Coverage.SemanticComplete = false
		}
	}
	for _, d := range result.Diagnostics {
		if d.Category == "syntax" {
			result.Coverage.SyntaxComplete = false
			result.Coverage.SemanticComplete = false
		}
		if d.Severity == "error" {
			result.Status = Invalid
		}
		if !seen[d.Code] {
			result.Coverage.Reasons = append(result.Coverage.Reasons, d.Code)
			seen[d.Code] = true
		}
	}
	if result.Status != Invalid && (!result.Coverage.SyntaxComplete || !result.Coverage.SemanticComplete) {
		result.Status = Incomplete
	}
}
