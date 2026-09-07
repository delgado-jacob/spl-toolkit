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
	positions := map[string]int{}
	var visit func(antlr.Tree, string, string)
	visit = func(node antlr.Tree, scopeID, stageID string) {
		switch ctx := node.(type) {
		case parser.IAnalysisStageContext:
			// The typed context owns the command token; no query text is split or reparsed.
			command := strings.ToLower(ctx.GetStart().GetText())
			stageID = fmt.Sprintf("stage-%d", len(result.Stages))
			stage := Stage{ID: stageID, Command: command, Position: positions[scopeID], ScopeID: scopeID, Location: parsed.source.contextLocation(ctx)}
			positions[scopeID]++
			result.Stages = append(result.Stages, stage)
			code := CodeUnsupportedSemantics
			if ctx.GetStart().GetTokenType() == parser.SPLLexerIDENTIFIER {
				code = CodeUnsupportedCommand
			}
			result.Diagnostics = append(result.Diagnostics, Diagnostic{Code: code, Severity: "warning", Category: "semantic", Message: fmt.Sprintf("command %q has unmodeled field effects", command), Location: stage.Location, StageID: stageID, ScopeID: scopeID})
		case *parser.AnalysisImplicitSearchContext:
			stageID = fmt.Sprintf("stage-%d", len(result.Stages))
			stage := Stage{ID: stageID, Command: "search", Position: positions[scopeID], ScopeID: scopeID, Location: parsed.source.contextLocation(ctx)}
			positions[scopeID]++
			result.Stages = append(result.Stages, stage)
			result.Diagnostics = append(result.Diagnostics, Diagnostic{Code: CodeUnsupportedSemantics, Severity: "warning", Category: "semantic", Message: "command \"search\" has unmodeled field effects", Location: stage.Location, StageID: stageID, ScopeID: scopeID})
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
			scopeID = child.ID
		}
		for i := 0; i < node.GetChildCount(); i++ {
			visit(node.GetChild(i), scopeID, stageID)
		}
	}
	visit(parsed.tree, root.ID, "")
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

// Capabilities describes the analysis contract shipped by this build.
func Capabilities() CapabilityManifest {
	manifest := CapabilityManifest{SchemaVersion: 1, Language: "spl", Profile: "splunkd", Version: "current", Commands: []Capability{}, Functions: []Capability{}}
	for _, name := range []string{"append", "appendpipe", "datamodel", "dedup", "eval", "eventstats", "fields", "from", "head", "inputlookup", "join", "lookup", "rename", "search", "sort", "stats", "streamstats", "table", "tail", "tstats", "where"} {
		manifest.Commands = append(manifest.Commands, Capability{Name: name, SyntaxSupported: true, SemanticSupported: false, Limitations: []string{"Field effects are not modeled."}})
	}
	return manifest
}
