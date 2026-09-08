package analysis

import (
	"encoding/json"
	"math/big"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	"github.com/delgado-jacob/spl-toolkit/parser"
	"github.com/delgado-jacob/spl-toolkit/parser/spl2"
)

type rewriteFact struct {
	values    []RewriteScalar
	complete  bool
	locations []Location
}
type rewriteFlow struct {
	epoch string
	facts map[string]rewriteFact
	seen  map[string][]string
}

func (f *rewriteFlow) clone() *rewriteFlow {
	if f == nil {
		return nil
	}
	out := &rewriteFlow{epoch: f.epoch, facts: map[string]rewriteFact{}, seen: map[string][]string{}}
	for k, v := range f.facts {
		v.values = rewriteCopy(v.values)
		v.locations = append([]Location{}, v.locations...)
		out.facts[k] = v
	}
	for k, v := range f.seen {
		out.seen[k] = copyIDs(v)
	}
	return out
}
func rewriteFactKey(kind string, id RewriteIdentity) string {
	b, _ := json.Marshal(id)
	return kind + ":" + string(b)
}
func (e *environment) rewriteInvalidate(name string) {
	if e.rewrite == nil {
		return
	}
	key := rewriteFactKey("field", rewriteAtom(name))
	delete(e.rewrite.facts, key)
	delete(e.rewrite.seen, key)
}
func (e *environment) rewriteBarrier() {
	if e.rewrite != nil {
		e.rewrite.facts = map[string]rewriteFact{}
		e.rewrite.seen = map[string][]string{}
	}
}
func (e *environment) rewriteProject(fields map[string]trackedField) {
	if e.rewrite == nil {
		return
	}
	retained := map[string]bool{}
	for name := range fields {
		retained[rewriteFactKey("field", rewriteAtom(name))] = true
	}
	for key := range e.rewrite.facts {
		if strings.HasPrefix(key, "field:") && !retained[key] {
			delete(e.rewrite.facts, key)
		}
	}
	for key := range e.rewrite.seen {
		if strings.HasPrefix(key, "field:") && !retained[key] {
			delete(e.rewrite.seen, key)
		}
	}
}
func (s *semanticStage) rewriteRemoval(id string, operand locatedOperand) {
	if s.result.rewrite == nil || id == "" {
		return
	}
	binding := "unavailable"
	field, known := s.env.fields[operand.Name]
	if known && !field.Conditional {
		binding = "derived"
		if field.source {
			binding = "source"
		}
	} else if s.env.uncertain || known {
		binding = "indeterminate"
	} else if !s.env.removed[operand.Name] && s.env.open {
		binding = "source"
	}
	s.rewriteBinding(id, binding, nil)
}
func (s *semanticStage) rewriteFacts() []RewriteFactEvidence {
	flow := s.rewriteState()
	out := []RewriteFactEvidence{}
	for i, p := range s.result.rewrite.probes {
		key := rewriteFactKey(p.Kind, p.Identity)
		fact, known := flow.facts[key]
		refs := copyIDs(flow.seen[key])
		state := "false"
		if len(refs) > 0 || len(fact.locations) > 0 {
			state = "true"
		} else if s.env.uncertain || p.Identity.Name == nil {
			state = "unknown"
		}
		out = append(out, RewriteFactEvidence{ProbeIndex: i, GuaranteedValues: rewriteCopy(append([]RewriteScalar{}, fact.values...)), LiteralComplete: known && fact.complete, ReferenceState: state, ReferenceIDs: refs, Locations: append([]Location{}, fact.locations...), Limitations: []RewriteLimitation{}})
	}
	return out
}
func rewriteScalar(text string, search bool, language string) (RewriteScalar, bool) {
	var value any
	if strings.Contains(text, "${") {
		return RewriteScalar{}, false
	}
	if len(text) > 1 && text[0] == '"' && text[len(text)-1] == '"' {
		var decoded string
		var ok bool
		if language == "spl2" {
			decoded, ok = spl2DecodeKey(text)
		} else {
			decoded, ok = rewriteSPLDecodeLiteral(text)
		}
		if !ok {
			return RewriteScalar{}, false
		}
		value = decoded
	} else {
		dec := json.NewDecoder(strings.NewReader(text))
		dec.UseNumber()
		err := dec.Decode(&value)
		if err != nil || !json.Valid([]byte(text)) {
			if !search || strings.ContainsAny(text, "*\\\"' +") {
				return RewriteScalar{}, false
			}
			value = text
		}
		if search {
			if _, ok := value.(bool); ok {
				value = text
			}
		}
	}
	kind := ""
	switch value.(type) {
	case string:
		kind = "string"
	case json.Number:
		kind = "number"
	case bool:
		kind = "boolean"
	case nil:
		kind = "null"
	default:
		return RewriteScalar{}, false
	}
	b, err := json.Marshal(value)
	return RewriteScalar{Kind: kind, Value: b}, err == nil
}
func rewriteSPLDecodeLiteral(text string) (string, bool) {
	body := text[1 : len(text)-1]
	var out strings.Builder
	for i := 0; i < len(body); i++ {
		if body[i] == '\\' {
			i++
			if i == len(body) || (body[i] != '\\' && body[i] != '"') {
				return "", false
			}
		}
		out.WriteByte(body[i])
	}
	return out.String(), true
}
func rewriteMergeFacts(a, b map[string]rewriteFact, or bool) map[string]rewriteFact {
	out := map[string]rewriteFact{}
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		prior, ok := out[k]
		if !ok {
			if !or {
				out[k] = v
			}
			continue
		}
		values := []RewriteScalar{}
		if or {
			for _, x := range prior.values {
				for _, y := range v.values {
					if rewriteScalarEqual(x, y) {
						values = append(values, x)
						break
					}
				}
			}
		} else {
			values = append(values, prior.values...)
			for _, x := range v.values {
				found := false
				for _, y := range values {
					found = found || (rewriteScalarEqual(x, y))
				}
				if !found {
					values = append(values, x)
				}
			}
		}
		out[k] = rewriteFact{values: values, complete: prior.complete && v.complete, locations: append(append([]Location{}, prior.locations...), v.locations...)}
	}
	if or {
		for k := range out {
			if _, ok := b[k]; !ok {
				delete(out, k)
			}
		}
	}
	return out
}
func (s *semanticStage) rewritePredicate(node antlr.Tree, language string, search, metrics bool) {
	if s.result.rewrite == nil || node == nil {
		return
	}
	facts := s.rewritePredicateFacts(node, language, search, metrics)
	flow := s.rewriteState()
	flow.facts = rewriteMergeFacts(flow.facts, facts, false)
}
func (s *semanticStage) rewritePredicateFacts(node antlr.Tree, language string, search, metrics bool) map[string]rewriteFact {
	empty := map[string]rewriteFact{}
	if node == nil {
		return empty
	}
	if ctx, ok := node.(antlr.ParserRuleContext); ok && !intact(ctx) {
		return empty
	}
	var name, kind string
	var literal antlr.Tree
	var location Location
	source := s.parsed
	if language == "spl" && source == nil {
		return empty
	}
	locate := func(ctx antlr.ParserRuleContext) Location {
		if language == "spl" {
			return s.parsed.source.contextLocation(ctx)
		}
		return s.result.rewrite.source.contextLocation(ctx)
	}
	or := false
	switch c := node.(type) {
	case parser.IAnalysisSearchUnaryContext:
		if c.NOT() != nil {
			return empty
		}
	case parser.IAnalysisNotContext:
		if c.NOT() != nil {
			return empty
		}
	case parser.IAnalysisSearchContext:
		or = len(c.AllAnalysisSearchAnd()) > 1
	case parser.IAnalysisOrContext:
		or = len(c.AllAnalysisAnd()) > 1
	case parser.IAnalysisSearchTermContext:
		if c.AnalysisIdentifier() != nil && c.AnalysisComparisonOperator() != nil && c.AnalysisComparisonOperator().EQ() != nil && len(c.AllAnalysisSearchValue()) == 1 {
			name = normalizedName(c.AnalysisIdentifier().GetText())
			literal = c.AnalysisSearchValue(0)
			location = locate(c.AnalysisIdentifier())
		}
	case parser.IAnalysisComparisonContext:
		if c.AnalysisComparisonOperator() != nil && c.AnalysisComparisonOperator().EQ() != nil && len(c.AllAnalysisConcat()) == 2 {
			if id := rewriteSPLSingleIdentifier(c.AnalysisConcat(0)); id != nil {
				name = normalizedName(id.GetText())
				literal = rewriteSPLSingleLiteral(c.AnalysisConcat(1))
				location = locate(id)
			}
		}
	case spl2.ISearchNotContext:
		if c.NOT() != nil {
			return empty
		}
	case spl2.INotExpressionContext:
		if c.LogicalNot() != nil {
			return empty
		}
	case spl2.ISearchXorContext:
		if len(c.AllSearchAnd()) > 1 {
			return empty
		}
	case spl2.IXorExpressionContext:
		if len(c.AllOrExpression()) > 1 {
			return empty
		}
	case spl2.ISearchOrContext:
		or = len(c.AllSearchNot()) > 1
	case spl2.IOrExpressionContext:
		for _, op := range c.AllLogicalOr() {
			if op.GetText() != "OR" {
				return empty
			}
		}
		or = len(c.AllAndExpression()) > 1
	case spl2.IAndExpressionContext:
		for _, op := range c.AllLogicalAnd() {
			if op.GetText() != "AND" {
				return empty
			}
		}
	case spl2.ISearchAtomContext:
		if c.Identifier() != nil && c.Comparison() != nil && (c.Comparison().ASSIGN() != nil || c.Comparison().EQ() != nil) && len(c.AllSearchValue()) == 1 {
			v := c.SearchValue(0)
			if v.SearchSignedNumber() != nil || v.SearchUnprovedLiteral() != nil || v.SearchDirective() != nil || v.RAW_STRING() != nil || (v.StringLiteral() != nil && len(v.StringLiteral().AllExpression()) > 0) {
				return empty
			}
			name, _ = spl2DecodeKey(c.Identifier().GetText())
			literal = v
			location = locate(c.Identifier())
		}
	case spl2.IPredicateContext:
		if c.Comparison() != nil && (c.Comparison().ASSIGN() != nil || c.Comparison().EQ() != nil) && len(c.AllAdditive()) == 2 {
			id := spl2SQLDirectField(c.Additive(0))
			right := spl2SingleAccess(c.Additive(1))
			if id != nil && right != nil && len(right.AllAccessPart()) == 0 {
				name, _ = spl2DecodeKey(id.GetText())
				literal = right.Primary().Literal()
				location = locate(id)
				if metrics && id.INDEX() != nil {
					kind = "index"
					literal = nil
					lit := right.Primary().Literal()
					field := right.Primary().FieldName()
					if lit != nil && lit.StringLiteral() != nil && len(lit.StringLiteral().AllExpression()) == 0 {
						literal = lit.StringLiteral()
					} else if field != nil && field.Identifier() != nil && field.Identifier().QuotedName() == nil {
						literal = field.Identifier()
					}
				}
				if lit, ok := literal.(spl2.ILiteralContext); ok && (lit.RAW_STRING() != nil || (lit.StringLiteral() != nil && len(lit.StringLiteral().AllExpression()) > 0)) {
					literal = nil
				}
			}
		}
	case parser.IAnalysisFunctionCallContext, parser.IAnalysisSubqueryContext, parser.IAnalysisMacroContext, spl2.ICallContext, spl2.IExistsPredicateContext, spl2.ISearchLiteralContext, spl2.ILambdaExpressionContext:
		return empty
	}
	if name != "" && literal != nil {
		if kind == "" {
			kind = "field"
		}
		if search && (name == "index" || name == "source" || name == "sourcetype") {
			kind = name
		}
		if kind == "field" {
			field, known := s.env.fields[name]
			if (known && (!field.source || field.Conditional)) || (!known && (s.env.uncertain || s.env.removed[name] || !s.env.open)) {
				return empty
			}
		}
		scalar, ok := rewriteScalar(literal.(antlr.ParserRuleContext).GetText(), search || metrics, language)
		if !ok {
			return empty
		}
		identity := rewriteAtom(name)
		if kind != "field" {
			if scalar.Kind != "string" {
				scalar.Value, _ = json.Marshal(literal.(antlr.ParserRuleContext).GetText())
				scalar.Kind = "string"
			}
			var value string
			_ = json.Unmarshal(scalar.Value, &value)
			identity = rewriteAtom(value)
			location = locate(literal.(antlr.ParserRuleContext))
		}
		return map[string]rewriteFact{rewriteFactKey(kind, identity): {values: []RewriteScalar{scalar}, complete: true, locations: []Location{location}}}
	}
	// Only routing/Boolean nodes reach children: an unsuccessful comparison may
	// contain literal-looking nested computations, which are not equality proof.
	switch node.(type) {
	case parser.IAnalysisComparisonContext, spl2.IPredicateContext:
		if node.GetChildCount() != 1 {
			return empty
		}
	}
	out := empty
	first := true
	for _, child := range node.GetChildren() {
		if _, ok := child.(antlr.ParserRuleContext); !ok {
			continue
		}
		switch child.(type) {
		case spl2.ILogicalOrContext, spl2.ILogicalAndContext, spl2.ILogicalXorContext:
			continue
		}
		next := s.rewritePredicateFacts(child, language, search, metrics)
		if first {
			out = next
			first = false
		} else {
			out = rewriteMergeFacts(out, next, or)
		}
	}
	return out
}
func rewriteSPLSingleIdentifier(n antlr.Tree) parser.IAnalysisIdentifierContext {
	if c, ok := n.(parser.IAnalysisIdentifierContext); ok {
		return c
	}
	if n.GetChildCount() == 1 {
		return rewriteSPLSingleIdentifier(n.GetChild(0))
	}
	return nil
}
func rewriteSPLSingleLiteral(n antlr.Tree) parser.IAnalysisLiteralContext {
	if c, ok := n.(parser.IAnalysisLiteralContext); ok {
		return c
	}
	if n.GetChildCount() == 1 {
		return rewriteSPLSingleLiteral(n.GetChild(0))
	}
	return nil
}

func rewriteScalarEqual(a, b RewriteScalar) bool {
	if a.Kind != b.Kind {
		return false
	}
	if a.Kind != "number" {
		return string(a.Value) == string(b.Value)
	}
	x, ok := new(big.Rat).SetString(string(a.Value))
	if !ok {
		return false
	}
	y, ok := new(big.Rat).SetString(string(b.Value))
	return ok && x.Cmp(y) == 0
}
func (s *semanticStage) rewriteUncertain() {
	if s.env.rewrite == nil {
		return
	}
	command := s.result.Stages[s.stage].Command
	if command == "where" || command == "search" || s.rewritePhase == "filter" {
		for key, fact := range s.env.rewrite.facts {
			fact.complete = false
			s.env.rewrite.facts[key] = fact
		}
		return
	}
	s.env.rewriteBarrier()
}
