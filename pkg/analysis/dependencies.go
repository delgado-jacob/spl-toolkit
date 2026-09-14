package analysis

import (
	"github.com/antlr4-go/antlr/v4"
	"github.com/delgado-jacob/spl-toolkit/parser"
	"strings"
)

func (s *semanticStage) dependency(ctx antlr.ParserRuleContext, name, kind string) {
	id := s.reference(ctx, name, kind, "read")
	if id == "" {
		return
	}
	s.addDependency(name, kind)
}

func (s *semanticStage) macro(ctx parser.IAnalysisMacroContext) {
	if ctx == nil || !s.sound(ctx) || ctx.AnalysisIdentifier() == nil {
		return
	}
	if s.macroEvidence == nil {
		s.macroEvidence = map[antlr.ParserRuleContext]string{}
	}
	if s.macroEvidence[ctx] != "" {
		return
	}
	name := normalizedName(ctx.AnalysisIdentifier().GetText())
	id := s.reference(ctx.AnalysisIdentifier(), name, "macro", "read")
	if id == "" {
		return
	}
	s.macroEvidence[ctx] = id
	s.addDependency(name, "macro")
	s.diagnosticAtOwned(CodeDynamicReference, "warning", "unsupported_semantics", "macro expansion is unresolved", s.parsed.source.contextLocation(ctx), true, []string{id})
}
func (s *semanticStage) addDependency(name, kind string) {
	switch kind {
	case "index":
		s.result.Dependencies.Indexes = append(s.result.Dependencies.Indexes, name)
	case "source":
		s.result.Dependencies.Sources = append(s.result.Dependencies.Sources, name)
	case "sourcetype":
		s.result.Dependencies.SourceTypes = append(s.result.Dependencies.SourceTypes, name)
	case "dataset":
		s.result.Dependencies.Datasets = append(s.result.Dependencies.Datasets, name)
	case "data_model":
		s.result.Dependencies.DataModels = append(s.result.Dependencies.DataModels, name)
	case "macro":
		s.result.Dependencies.Macros = append(s.result.Dependencies.Macros, name)
	case "lookup":
		s.result.Dependencies.Lookups = append(s.result.Dependencies.Lookups, name)
	}
}

// This walk sees only grammar-established roles and never enters another scope.
// Opaque arguments are not expression reads. Macro arguments require expansion.
func (s *semanticStage) dependencies(node antlr.Tree) {
	switch c := node.(type) {
	case parser.IAnalysisSubqueryContext:
		return
	case parser.IAnalysisMacroContext:
		s.macro(c)
		return
	case *parser.AnalysisDatamodelStageContext:
		if !s.sound(c.AnalysisDataModelName()) {
			return
		}
		model := normalizedName(c.AnalysisDataModelName().GetText())
		if !plainCatalogComponent(model) {
			s.diagnostic(CodeUnsupportedSemantics, "data model name must be an exact catalog component", c.AnalysisDataModelName())
			return
		}
		s.dependency(c.AnalysisDataModelName(), model, "data_model")
		if d := c.AnalysisDataModelDataset(); d != nil && s.sound(d) {
			dataset := normalizedName(d.GetText())
			if !plainCatalogComponent(dataset) {
				s.diagnostic(CodeUnsupportedSemantics, "data model dataset must be an exact catalog component", d)
				return
			}
			s.dependency(d, model+"."+dataset, "dataset")
		}
	case parser.IAnalysisFromDatasetContext:
		s.qualifiedCatalog(c, true)
		return
	case parser.IAnalysisTstatsWhereContext:
		s.searchDependencies(c.AnalysisSearch())
		return
	case parser.IAnalysisTstatsFromContext:
		s.qualifiedCatalog(c.AnalysisDataModelName(), false)
		return
	}
	for i := 0; i < node.GetChildCount(); i++ {
		s.dependencies(node.GetChild(i))
	}
}
func plainCatalogComponent(name string) bool {
	return name != "" && !strings.ContainsAny(name, ".:*\\")
}

// Qualified catalog names are single lexer tokens. Decomposition is bounded to
// this typed operand; overlapping root/dataset spans deliberately retain source
// components for future rewrite eligibility checks, rather than guessed edits.
func (s *semanticStage) qualifiedCatalog(ctx antlr.ParserRuleContext, from bool) {
	if !s.sound(ctx) {
		return
	}
	raw := ctx.GetText()
	name := normalizedName(raw)
	prefix := 0
	if from {
		if !strings.HasPrefix(strings.ToLower(name), "datamodel:") {
			s.dependency(ctx, name, "dataset")
			return
		}
		prefix = len("datamodel:")
	}
	content := name[prefix:]
	parts := strings.Split(content, ".")
	if len(parts) > 2 || !plainCatalogComponent(parts[0]) || (len(parts) == 2 && !plainCatalogComponent(parts[1])) || strings.Contains(raw, "\\") {
		s.diagnostic(CodeUnsupportedSemantics, "qualified data model catalog requires exact model and optional dataset components", ctx)
		return
	}
	start := ctx.GetStart().GetStart() + prefix
	if len(raw) > 0 && (raw[0] == '\'' || raw[0] == '"') {
		start++
	}
	rootEnd := start + len([]rune(parts[0]))
	rootID := s.referenceAt(s.parsed.source.location(start, rootEnd), parts[0], "data_model", "read", "exact")
	s.rewriteReference(rootID, locatedOperand{Name: parts[0], Location: s.parsed.source.location(start, rootEnd), Resolution: "exact", Sound: true, rewrite: rewriteOwner{role: "catalog_component", location: s.parsed.source.contextLocation(ctx), component: true}}, "data_model", "read")
	s.addDependency(parts[0], "data_model")
	if len(parts) == 2 {
		datasetID := s.referenceAt(s.parsed.source.location(start, start+len([]rune(content))), content, "dataset", "read", "exact")
		s.rewriteReference(datasetID, locatedOperand{Name: content, Location: s.parsed.source.location(start, start+len([]rune(content))), Resolution: "exact", Sound: true, rewrite: rewriteOwner{role: "qualified_dataset", location: s.parsed.source.contextLocation(ctx), component: true}}, "dataset", "read")
		s.addDependency(content, "dataset")
	}
}

func (s *semanticStage) searchDependencies(node antlr.Tree) {
	if term, ok := node.(parser.IAnalysisSearchTermContext); ok && term.AnalysisIdentifier() != nil {
		kind := strings.ToLower(normalizedName(term.AnalysisIdentifier().GetText()))
		if s.sound(term) && (kind == "index" || kind == "source" || kind == "sourcetype") {
			for _, value := range term.AllAnalysisSearchValue() {
				s.dependency(value, normalizedName(value.GetText()), kind)
			}
		}
		return
	}
	if _, child := node.(parser.IAnalysisSubqueryContext); child {
		return
	}
	for i := 0; i < node.GetChildCount(); i++ {
		s.searchDependencies(node.GetChild(i))
	}
}
