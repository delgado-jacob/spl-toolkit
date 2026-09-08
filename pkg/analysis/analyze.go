// Package analysis exposes deterministic, source-aware analysis of the supported SPL contract.
package analysis

import (
	"fmt"
	"sort"
	"unicode/utf8"
)

func normalizeDocument(document QueryDocument) (QueryDocument, error) {
	selectors, err := normalizeSelectors(CapabilityOptions{Language: document.Language, Profile: document.Profile, Version: document.Version})
	if err != nil {
		return document, err
	}
	document.Language, document.Profile, document.Version = selectors.Language, selectors.Profile, selectors.Version
	if !utf8.ValidString(document.Text) {
		return document, fmt.Errorf("document text is not valid UTF-8")
	}
	if !utf8.ValidString(document.SourceID) {
		return document, fmt.Errorf("document source_id is not valid UTF-8")
	}
	return document, nil
}
func newResult(document QueryDocument) *Result {
	return &Result{SchemaVersion: 1, Document: document, Coverage: Coverage{SyntaxComplete: true, SemanticComplete: true, Reasons: []string{}}, Stages: []Stage{}, Scopes: []Scope{}, References: []Reference{}, Lineage: []Lineage{}, Dependencies: Dependencies{Indexes: []string{}, Sources: []string{}, SourceTypes: []string{}, Datasets: []string{}, Lookups: []string{}, DataModels: []string{}, Macros: []string{}}, Diagnostics: []Diagnostic{}}
}

// Analyze retains original source and reports syntax errors as findings. Unsupported document options are API errors.
func Analyze(document QueryDocument) (*Result, error) {
	return analyze(document, nil)
}

func analyze(document QueryDocument, refinement *sourceRefinement) (*Result, error) {
	return analyzeRewrite(document, refinement, nil)
}

func analyzeRewrite(document QueryDocument, refinement *sourceRefinement, rewrite *RewriteSession) (*Result, error) {
	normalized, err := normalizeDocument(document)
	if err != nil {
		return nil, err
	}
	result := newResult(normalized)
	result.rewrite = rewrite
	if rewrite != nil {
		rewrite.source = newSourceIndex(normalized.Text)
	}
	if normalized.Language == "spl2" {
		if refinement != nil {
			refinement.literalSourceNames = true
		}
		analyzeSPL2(result, parseSPL2Document(normalized.Text), refinement)
	} else {
		parsed := parseDocument(normalized.Text)
		result.Diagnostics = append(result.Diagnostics, parsed.diagnostics...)
		analyzeParsed(result, parsed, refinement)
	}
	finalizeResult(result)
	return result, nil
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
