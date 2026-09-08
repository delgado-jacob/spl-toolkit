package analysis

import (
	"strings"
	"unicode"

	"github.com/antlr4-go/antlr/v4"
	"github.com/delgado-jacob/spl-toolkit/parser"
)

func (s *semanticStage) rewriteSPLOwner(ctx antlr.ParserRuleContext) rewriteOwner {
	if s.result.rewrite == nil {
		return rewriteOwner{}
	}
	o := rewriteOwner{location: s.parsed.source.contextLocation(ctx)}
	switch c := ctx.(type) {
	case parser.IAnalysisIdentifierContext:
		o.role = "expression_atom"
	case parser.IAnalysisSelectorContext:
		o.role = "selector_atom"
	case parser.IAnalysisSearchValueContext:
		o.role = "search_value"
	case parser.IAnalysisCatalogNameContext, parser.IAnalysisDataModelNameContext, parser.IAnalysisFromDatasetContext:
		o.role = "catalog_atom"
	case parser.IAnalysisDataModelDatasetContext:
		o.role = "catalog_component"
		o.component = true
		if parent, ok := c.GetParent().(*parser.AnalysisDatamodelStageContext); ok {
			o.prefix = normalizedName(parent.AnalysisDataModelName().GetText()) + "."
			o.modelLocation = s.parsed.source.contextLocation(parent.AnalysisDataModelName())
		}
	case parser.IAnalysisFunctionCallContext:
		o.role = "implicit_output"
		o.implicit = strings.ToLower(c.AnalysisFunctionName().GetText())
	}
	for parent := ctx.GetParent(); parent != nil; parent = parent.GetParent() {
		switch c := parent.(type) {
		case parser.IAnalysisFunctionCallContext:
			spec, known := functions[strings.ToLower(c.AnalysisFunctionName().GetText())]
			if !known || spec.dynamic {
				o.role = ""
				o.location = s.parsed.source.contextLocation(c)
				return o
			}
		case parser.IAnalysisSearchTermContext:
			if o.role == "expression_atom" {
				o.role = "search_field"
			}
			return o
		case parser.IAnalysisRenameContext:
			if ctx.GetStart().GetStart() <= c.AnalysisSelector().GetStop().GetStop() {
				o.role = "rename_input"
			}
			return o
		case parser.IAnalysisLookupInputContext:
			o.location = s.parsed.source.contextLocation(c)
			o.role = "lookup_local"
			if c.AnalysisAlias() == nil {
				o.role = "lookup_dual"
			}
			return o
		case parser.IAnalysisStageContext:
			return o
		}
	}
	return o
}
func rewriteSPLQuoted(name string, quote byte) (string, bool) {
	if strings.ContainsAny(name, "\r\n\x00") {
		return "", false
	}
	var out strings.Builder
	out.WriteByte(quote)
	for _, r := range name {
		if r == rune(quote) || r == '\\' {
			out.WriteByte('\\')
		}
		out.WriteRune(r)
	}
	out.WriteByte(quote)
	return out.String(), true
}
func rewriteSPLBare(name string) bool {
	if name == "" {
		return false
	}
	for i, r := range name {
		if !(unicode.IsLetter(r) || r == '_' || r == '.' || r == ':' || (i > 0 && unicode.IsDigit(r))) {
			return false
		}
	}
	switch strings.ToUpper(name) {
	case "AND", "OR", "NOT", "AS", "BY", "IN", "LIKE", "OUTPUT", "OUTPUTNEW":
		return false
	}
	return true
}
func rewriteSPLText(site *rewriteSite, target string, before string) (string, bool) {
	if site.owner.component {
		if !plainCatalogComponent(target) || strings.ContainsAny(target, " \t\r\n|'\"") {
			return "", false
		}
		if len(before) > 1 && (before[0] == '\'' || before[0] == '"') {
			return rewriteSPLQuoted(target, before[0])
		}
		return target, true
	}
	quote := byte(0)
	if len(before) > 1 && (before[0] == '\'' || before[0] == '"') {
		quote = before[0]
	}
	switch site.public.Role {
	case "search_field":
		switch strings.ToLower(target) {
		case "index", "source", "sourcetype":
			return "", false
		}
		if quote == 0 && rewriteSPLBare(target) {
			return target, true
		}
		return rewriteSPLQuoted(target, '\'')
	case "search_value":
		if strings.Contains(target, "*") {
			return "", false
		}
		if quote == 0 && rewriteSPLBare(target) {
			return target, true
		}
		return rewriteSPLQuoted(target, '"')
	case "catalog_atom":
		if quote == 0 && rewriteSPLBare(target) {
			return target, true
		}
		if quote == 0 {
			quote = '\''
		}
		return rewriteSPLQuoted(target, quote)
	case "expression_atom", "selector_atom", "rename_input", "lookup_local", "lookup_dual", "null_test":
		if (site.public.Role == "selector_atom" || site.public.Role == "rename_input") && strings.Contains(target, "*") {
			return "", false
		}
		if quote == 0 && rewriteSPLBare(target) {
			return target, true
		}
		return rewriteSPLQuoted(target, '\'')
	}
	return "", false
}
