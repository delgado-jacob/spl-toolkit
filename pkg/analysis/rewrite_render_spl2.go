package analysis

import (
	"strconv"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	"github.com/delgado-jacob/spl-toolkit/parser/spl2"
)

func (s *spl2SemanticStage) rewriteSPL2Owner(ctx antlr.ParserRuleContext) rewriteOwner {
	if s.result.rewrite == nil {
		return rewriteOwner{}
	}
	o := rewriteOwner{location: s.parsed2.source.contextLocation(ctx)}
	switch c := ctx.(type) {
	case spl2.IIdentifierContext:
		o.role = "expression_atom"
	case spl2.IRenameSourceContext:
		o.role = "rename_input"
	case spl2.IQuotedNameContext:
		o.role = "quoted_catalog_atom"
	case spl2.ILookupDatasetContext:
		o.role = "catalog_atom"
	case spl2.ILookupColumnContext:
		o.role = "lookup_dual"
	case spl2.ILookupEventFieldContext:
		o.role = "lookup_local"
	case spl2.ISearchValueContext:
		o.role = "search_value"
		if c.SearchSignedNumber() != nil || c.SearchUnprovedLiteral() != nil || c.SearchDirective() != nil || c.RAW_STRING() != nil || strings.Contains(c.GetText(), "*") || (c.StringLiteral() != nil && len(c.StringLiteral().AllExpression()) > 0) {
			o.role = ""
		}
	case spl2.IStringLiteralContext:
		o.role = "metric_value"
	case spl2.IObjectKeyContext:
		o.role = "definition"
	}
	for parent := ctx.GetParent(); parent != nil; parent = parent.GetParent() {
		switch c := parent.(type) {
		case spl2.ICallContext:
			if _, known := spl2Functions[c.Identifier().GetText()]; !known {
				o.role = ""
				o.location = s.parsed2.source.contextLocation(c)
				return o
			}
		case spl2.IFieldTemplateContext, spl2.INamedArgumentContext:
			o.role = ""
			o.location = s.parsed2.source.contextLocation(parent.(antlr.ParserRuleContext))
			return o
		case spl2.IAccessContext:
			if len(c.AllAccessPart()) > 0 {
				o.role = "navigation"
				o.location = s.parsed2.source.contextLocation(c)
				return o
			}
		case spl2.ISearchAtomContext:
			if o.role == "expression_atom" {
				o.role = "search_field"
			}
			return o
		case spl2.IRenameSourceContext:
			o.role = "rename_input"
			return o
		case spl2.IDatasetContext:
			o.role = "catalog_atom"
			return o
		case spl2.ILookupMatchContext:
			o.location = s.parsed2.source.contextLocation(c)
			return o
		case spl2.IFieldSelectorContext, spl2.ITableFieldContext, spl2.IGroupFieldContext:
			o.role = "selector_atom"
			return o
		case spl2.ICommandContext:
			return o
		}
	}
	return o
}
func rewriteSPL2Quoted(name string, quote byte) string {
	var out strings.Builder
	out.WriteByte(quote)
	for _, r := range name {
		switch r {
		case '$':
			out.WriteString("\\u0024")
		case '\\':
			out.WriteString("\\\\")
		case '\n':
			out.WriteString("\\n")
		case '\r':
			out.WriteString("\\r")
		case '\t':
			out.WriteString("\\t")
		default:
			if r == rune(quote) {
				out.WriteByte('\\')
				out.WriteRune(r)
			} else if r < 32 {
				out.WriteString(strconv.QuoteRuneToASCII(r)[1 : len(strconv.QuoteRuneToASCII(r))-1])
			} else {
				out.WriteRune(r)
			}
		}
	}
	out.WriteByte(quote)
	return out.String()
}
func rewriteSPL2Bare(name string) bool {
	input := antlr.NewInputStream(name)
	lexer := spl2.NewSPL2Lexer(input)
	lexer.RemoveErrorListeners()
	token := lexer.NextToken()
	return token.GetTokenType() == spl2.SPL2LexerIDENTIFIER && token.GetText() == name && lexer.NextToken().GetTokenType() == antlr.TokenEOF
}
func rewriteSPL2Text(site *rewriteSite, target, before string) (string, bool) {
	switch site.public.Role {
	case "search_field":
		switch target {
		case "index", "source", "sourcetype", "earliest", "latest", "timeformat":
			return "", false
		}
		if !strings.HasPrefix(before, "'") && rewriteSPL2Bare(target) {
			return target, true
		}
		return rewriteSPL2Quoted(target, '\''), true
	case "search_value":
		if strings.Contains(target, "*") {
			return "", false
		}
		if !strings.HasPrefix(before, "\"") && rewriteSPL2Bare(target) {
			return target, true
		}
		return rewriteSPL2Quoted(target, '"'), true
	case "metric_value":
		if !plainCatalogComponent(target) {
			return "", false
		}
		return rewriteSPL2Quoted(target, '"'), true
	case "quoted_catalog_atom":
		if strings.ContainsAny(target, "*\\:") {
			return "", false
		}
		return rewriteSPL2Quoted(target, '\''), true
	case "expression_atom", "selector_atom", "rename_input", "lookup_local", "lookup_dual", "null_test", "catalog_atom":
		if (site.public.Role == "selector_atom" || site.public.Role == "rename_input") && strings.Contains(target, "*") {
			return "", false
		}
		if !strings.HasPrefix(before, "'") && rewriteSPL2Bare(target) {
			return target, true
		}
		return rewriteSPL2Quoted(target, '\''), true
	}
	return "", false
}

func (s *spl2SemanticStage) rewriteNavigation(access spl2.IAccessContext) rewriteOwner {
	owner := rewriteOwner{role: "navigation", location: s.parsed2.source.contextLocation(access)}
	if field := access.Primary().FieldName(); field != nil && field.Identifier() != nil {
		name, ok := spl2DecodeKey(field.Identifier().GetText())
		if !ok || s.aliases[name] {
			return owner
		}
		path := []string{name}
		for _, part := range access.AllAccessPart() {
			if part.Identifier() == nil {
				return owner
			}
			name, ok = spl2DecodeKey(part.Identifier().GetText())
			if !ok {
				return owner
			}
			path = append(path, name)
		}
		owner.identity = RewriteIdentity{Path: path}
	}
	return owner
}
