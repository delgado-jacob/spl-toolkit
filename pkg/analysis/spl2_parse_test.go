package analysis

import (
	"fmt"
	"strings"
	"testing"

	"github.com/antlr4-go/antlr/v4"
	"github.com/delgado-jacob/spl-toolkit/parser/spl2"
)

type spl2CountingCharStream struct {
	antlr.CharStream
	lookaheads int
}

func (s *spl2CountingCharStream) LA(offset int) int {
	s.lookaheads++
	return s.CharStream.LA(offset)
}

func spl2RequireNoDiagnostics(t *testing.T, text string) *spl2ParsedDocument {
	t.Helper()
	p := parseSPL2Document(text)
	if !p.syntaxComplete || len(p.diagnostics) != 0 {
		t.Fatalf("selected syntax rejected: %+v", p.diagnostics)
	}
	return p
}

func spl2RequireLocatedError(t *testing.T, text string) *spl2ParsedDocument {
	t.Helper()
	p := parseSPL2Document(text)
	found := false
	for _, diagnostic := range p.diagnostics {
		if diagnostic.Severity != "error" {
			continue
		}
		if diagnostic.Location.Start.Offset < 0 || diagnostic.Location.End.Offset > len(text) || diagnostic.Location.Start.Offset > diagnostic.Location.End.Offset {
			t.Fatalf("unlocated diagnostic for %q: %+v", text, diagnostic)
		}
		found = true
	}
	if !found || p.syntaxComplete {
		t.Fatalf("missing located error for %q: %+v", text, p.diagnostics)
	}
	return p
}

func spl2RequireTypedModule(t *testing.T, text string) *spl2ParsedDocument {
	t.Helper()
	p := parseSPL2Document(text)
	if p.tree == nil || p.tree.ModuleDeclaration() == nil {
		t.Fatalf("missing module declaration: %+v", p.diagnostics)
	}
	if len(p.diagnostics) != 0 || !p.syntaxComplete {
		t.Fatalf("selected typed module rejected: %+v", p.diagnostics)
	}
	return p
}

func TestSPL2StartNegativeSource(t *testing.T) {
	text := "failure index=app"
	parsed := parseSPL2Document(text)
	if parsed.source.text != text || len(parsed.diagnostics) == 0 {
		t.Fatal("invalid standalone start was prefixed, lost, or accepted")
	}
}

func TestSPL2ExpressionTyped(t *testing.T) {
	for _, text := range []string{
		"FROM main | eval x=a/b/c",
		"FROM main | eval 'résumé'=true, n=null, raw=@\"C:\\logs\"",
		"FROM main | eval x={a:[user,2],b:actor.name,}",
		"FROM main | eval x=if(code=200,1,0)",
		"FROM main | eval x=round(bytes,precision:2)",
		"FROM main | eval x=map(items,$v -> {$n=$v+factor; return $n})",
		"FROM main | eval x=\"a ${if(user=\"x\",1,0)} | b\"",
	} {
		t.Run(text, func(t *testing.T) {
			p := parseSPL2Document(text)
			if len(p.diagnostics) > 0 {
				t.Fatalf("%+v", p.diagnostics)
			}
			if p.tree == nil || p.syntax == nil {
				t.Fatal("missing typed tree")
			}
		})
	}
	for _, text := range []string{"FROM main | eval x=a+", "FROM main | eval x=[a,,b]", "FROM main | eval x=round(num:bytes,2)", "FROM main | eval x=\"${}\"", "FROM main | eval x=map(items,$v -> {$n=2})"} {
		if p := parseSPL2Document(text); len(p.diagnostics) == 0 {
			t.Errorf("accepted malformed %s", text)
		}
	}
}

func spl2Nodes(node *spl2SyntaxNode, kind string) []*spl2SyntaxNode {
	var result []*spl2SyntaxNode
	if node.Kind == kind {
		result = append(result, node)
	}
	for _, child := range node.Children {
		result = append(result, spl2Nodes(child, kind)...)
	}
	return result
}
func TestSPL2ExpressionPrecedence(t *testing.T) {
	cases := []struct{ text, kind, shape string }{
		{"FROM main | eval x=a/b/c", "multiplicative", `(multiplicative IDENTIFIER:"a" SLASH:"/" IDENTIFIER:"b" SLASH:"/" IDENTIFIER:"c")`},
		{"FROM main | eval x=-a+b*2", "additive", `(additive (unary MINUS:"-" IDENTIFIER:"a") PLUS:"+" (multiplicative IDENTIFIER:"b" STAR:"*" NUMBER:"2"))`},
		{"FROM main | where a=1 OR b=2 AND c=3", "orExpression", `(orExpression (predicate IDENTIFIER:"a" ASSIGN:"=" NUMBER:"1") OR:"OR" (andExpression (predicate IDENTIFIER:"b" ASSIGN:"=" NUMBER:"2") AND:"AND" (predicate IDENTIFIER:"c" ASSIGN:"=" NUMBER:"3")))`},
		{"search a=1 OR b=2 AND c=3", "searchAnd", `(searchAnd (searchOr (searchAtom IDENTIFIER:"a" ASSIGN:"=" NUMBER:"1") OR:"OR" (searchAtom IDENTIFIER:"b" ASSIGN:"=" NUMBER:"2")) AND:"AND" (searchAtom IDENTIFIER:"c" ASSIGN:"=" NUMBER:"3"))`},
	}
	for _, tc := range cases {
		p := parseSPL2Document(tc.text)
		nodes := spl2Nodes(p.syntax, tc.kind)
		if len(p.diagnostics) > 0 || len(nodes) == 0 {
			t.Fatalf("%s: %+v", tc.text, p.diagnostics)
		}
		if got := nodes[0].shape(); got != tc.shape {
			t.Errorf("got %s\nwant %s", got, tc.shape)
		}
	}
}
func TestSPL2LexOwnership(t *testing.T) {
	text := "FROM main | eval x=\"${if(a=\"|\",{key:[b,2]},0)}\", y=map(items,$v -> {$n=$v+free; return $n})"
	p := parseSPL2Document(text)
	if len(p.diagnostics) > 0 {
		t.Fatalf("%+v", p.diagnostics)
	}
	for kind, want := range map[string]int{"PIPE": 1, "objectEntry": 1, "array": 1, "lambdaBlock": 1, "STRING_INTERPOLATION": 1, "namedArgument": 0, "LOCAL": 4} {
		if got := len(spl2Nodes(p.syntax, kind)); got != want {
			t.Errorf("%s got %d want %d", kind, got, want)
		}
	}
	p = parseSPL2Document("FROM main | eval x=round(bytes,precision:2), y=coalesce(values:[primary,backup])")
	if got := len(spl2Nodes(p.syntax, "namedArgument")); got != 2 {
		t.Fatalf("named arguments %d %+v", got, p.diagnostics)
	}
	named := spl2Nodes(p.syntax, "namedArgument")[0]
	if got := named.Children[0].shape(); got != `IDENTIFIER:"precision"` {
		t.Fatal(got)
	}
	// Labels and object keys have their own contexts, not access-expression reads.
	if len(spl2Nodes(named.Children[0], "access")) != 0 {
		t.Fatal("named label is a field read")
	}
	p = parseSPL2Document("FROM main | rex field=payload /(?<digits>[0-9]+)/ | eval x=a/b/c")
	if len(p.diagnostics) > 0 || len(spl2Nodes(p.syntax, "REGEX")) != 1 || len(spl2Nodes(p.syntax, "SLASH")) != 2 {
		t.Fatalf("regex/division ownership: %+v", p.diagnostics)
	}
	p = parseSPL2Document("FROM main | rex field=payload \"(?<tag>red|blue)\"")
	if len(p.diagnostics) > 0 || len(spl2Nodes(p.syntax, "PIPE")) != 1 {
		t.Fatalf("quoted alternation %+v", p.diagnostics)
	}
}
func TestSPL2SourceUnicodeCRLF(t *testing.T) {
	text := "FROM main // | ignored\r\n| eval 'résumé'='café', x=\"${'résumé'}\""
	p := parseSPL2Document(text)
	if len(p.diagnostics) > 0 {
		t.Fatalf("%+v", p.diagnostics)
	}
	names := spl2Nodes(p.syntax, "quotedName")
	if len(names) != 3 {
		t.Fatalf("quoted names %d", len(names))
	}
	first := names[0].Location
	if first.Start.Line != 2 || first.Start.Column != 8 || text[first.Start.Offset:first.End.Offset] != "'résumé'" {
		t.Fatalf("%+v", first)
	}
	for _, name := range names {
		slice := text[name.Location.Start.Offset:name.Location.End.Offset]
		decoded, ok := spl2DecodeKey(slice)
		if !ok || (decoded != "résumé" && decoded != "café") {
			t.Fatal(slice, decoded)
		}
	}
}
func TestSPL2ExpressionContractsAndHolds(t *testing.T) {
	for _, text := range []string{"FROM main | eval x={a:1,\"a\":2}", `FROM main | eval x={'a':1,"\u0061":2}`, "FROM main | eval x=map(items,$v -> map(items,$w -> $w))"} {
		p := parseSPL2Document(text)
		if !p.syntaxComplete || len(p.diagnostics) == 0 || p.diagnostics[0].Category != "contract" {
			t.Fatalf("contract: %s %+v", text, p.diagnostics)
		}
	}
	for _, text := range []string{"FROM main | eval x=[1,2,]", "FROM main | where user IS NOT string", "FROM main | rex field=payload /(?<tag>red|blue)/", "FROM main | eval x=`status=4*`"} {
		p := parseSPL2Document(text)
		if p.syntaxComplete || len(p.diagnostics) == 0 {
			t.Fatal("held form complete", text)
		}
		for _, d := range p.diagnostics {
			if d.Severity == "error" {
				t.Fatalf("held form made invalid %s: %+v", text, d)
			}
		}
	}
	p := parseSPL2Document("FROM main | eval x={a:1,nested:{a:2}}")
	if len(p.diagnostics) != 0 {
		t.Fatalf("object scope leaked: %+v", p.diagnostics)
	}
}

func TestSPL2LexSearchCommentBoundary(t *testing.T) {
	for _, text := range []string{"search index=main /* forbidden */", "search index=main /* forbidden */ | eval x=2", "index=main /* forbidden */"} {
		p := parseSPL2Document(text)
		if len(p.diagnostics) == 0 || p.syntaxComplete {
			t.Errorf("accepted search comment %q", text)
		}
	}
}
func TestSPL2StartExcludedModule(t *testing.T) {
	spl2RequireTypedModule(t, "$saved = FROM main;")
	for _, text := range []string{"import foo", "function f() {}", "FROM main; FROM other"} {
		p := parseSPL2Document(text)
		found := false
		for _, d := range p.diagnostics {
			if d.Code == "SPL_UNSUPPORTED_MODULE" {
				found = true
				if d.Category != "unsupported_syntax" || d.Severity != "error" || d.Location.Start.Offset < 0 || d.Location.End.Offset > len(text) || d.Location.Start.Offset >= d.Location.End.Offset {
					t.Errorf("module diagnostic contract %q %+v", text, d)
				}
			}
		}
		if !found || p.syntaxComplete {
			t.Errorf("module not classified %q %+v", text, p.diagnostics)
		}
	}
}

func TestSPL2ExpressionLambdaSignedDefault(t *testing.T) {
	p := parseSPL2Document("FROM main | eval x=map(items,($v:int=-2) -> $v+factor), a=[], b={}")
	if len(p.diagnostics) > 0 || len(spl2Nodes(p.syntax, "lambdaParameter")) != 1 || len(spl2Nodes(p.syntax, "MINUS")) != 1 {
		t.Fatalf("signed constant/default ownership %+v", p.diagnostics)
	}
}

func TestSPL2StartImplicitBooleanTree(t *testing.T) {
	cases := []struct{ text, shape string }{
		{"index=app AND host=api", `(searchAnd (searchAtom INDEX:"index" ASSIGN:"=" IDENTIFIER:"app") AND:"AND" (searchAtom IDENTIFIER:"host" ASSIGN:"=" IDENTIFIER:"api"))`},
		{"index=app OR host=api", `(searchOr (searchAtom INDEX:"index" ASSIGN:"=" IDENTIFIER:"app") OR:"OR" (searchAtom IDENTIFIER:"host" ASSIGN:"=" IDENTIFIER:"api"))`},
		{"index=app XOR host=api", `(searchXor (searchAtom INDEX:"index" ASSIGN:"=" IDENTIFIER:"app") XOR:"XOR" (searchAtom IDENTIFIER:"host" ASSIGN:"=" IDENTIFIER:"api"))`},
		{"index=app host=api", `(searchAnd (searchAtom INDEX:"index" ASSIGN:"=" IDENTIFIER:"app") (searchAtom IDENTIFIER:"host" ASSIGN:"=" IDENTIFIER:"api"))`},
		{"index=app OR host=api AND status=200 XOR kind=other", `(searchXor (searchAnd (searchOr (searchAtom INDEX:"index" ASSIGN:"=" IDENTIFIER:"app") OR:"OR" (searchAtom IDENTIFIER:"host" ASSIGN:"=" IDENTIFIER:"api")) AND:"AND" (searchAtom IDENTIFIER:"status" ASSIGN:"=" NUMBER:"200")) XOR:"XOR" (searchAtom IDENTIFIER:"kind" ASSIGN:"=" IDENTIFIER:"other"))`},
	}
	for _, tc := range cases {
		t.Run(tc.text, func(t *testing.T) {
			p := parseSPL2Document(tc.text)
			if !p.syntaxComplete || len(p.diagnostics) > 0 {
				t.Fatalf("%+v", p.diagnostics)
			}
			expressions := spl2Nodes(p.syntax, "searchExpression")
			if len(expressions) != 1 || expressions[0].shape() != tc.shape {
				t.Fatalf("initial index escaped Boolean tree: %s", p.syntax.shape())
			}
			loc := expressions[0].Location
			if p.source.text[loc.Start.Offset:loc.End.Offset] != tc.text {
				t.Fatal("initial constraint excluded from expression source span")
			}
		})
	}
	for _, text := range []string{"host=api AND index=app", "failure index=app", "index=app OR", "index=app AND AND host=api"} {
		p := parseSPL2Document(text)
		if len(p.diagnostics) == 0 || p.syntaxComplete {
			t.Errorf("accepted invalid start or continuation %q", text)
		}
	}
}
func TestSPL2ExpressionLambdaConstantDefault(t *testing.T) {
	for _, value := range []string{`"${fallback}"`, `"prefix ${fallback} suffix"`} {
		p := parseSPL2Document("FROM main | eval result=map(items,($x:string=" + value + ") -> $x)")
		if !p.syntaxComplete || len(p.diagnostics) != 1 {
			t.Fatalf("expected parsed but contract-invalid default: %+v", p.diagnostics)
		}
		d := p.diagnostics[0]
		if d.Code != CodeSyntaxError || d.Severity != "error" || d.Category != "contract" {
			t.Fatalf("%+v", d)
		}
		if p.source.text[d.Location.Start.Offset:d.Location.End.Offset] != value {
			t.Fatalf("default finding misplaced: %+v", d.Location)
		}
	}
	for _, text := range []string{`FROM main | eval x=map(items,($v:string="none") -> $v)`, `FROM main | eval x=map(items,($v:int=-2) -> $v+factor)`} {
		p := parseSPL2Document(text)
		if !p.syntaxComplete || len(p.diagnostics) > 0 {
			t.Fatalf("constant default rejected: %+v", p.diagnostics)
		}
	}
}
func TestSPL2ExpressionLowercaseBetween(t *testing.T) {
	p := parseSPL2Document("FROM main | where amount between 0 and 5")
	if !p.syntaxComplete || len(p.diagnostics) > 0 {
		t.Fatalf("evidenced lowercase predicate rejected: %+v", p.diagnostics)
	}
	nodes := spl2Nodes(p.syntax, "predicate")
	if len(nodes) != 1 {
		t.Fatalf("predicate ownership %s", p.syntax.shape())
	}
	if text := p.source.text[nodes[0].Location.Start.Offset:nodes[0].Location.End.Offset]; text != "amount between 0 and 5" {
		t.Fatalf("predicate source %q", text)
	}
}

func TestSPL2ExpressionContextualCasing(t *testing.T) {
	for _, text := range []string{
		"FROM main | where amount BETWEEN 0 and 5",
		"FROM main | where amount between 0 AND 5",
		"FROM main | where amount BeTwEeN 0 AnD 5",
		"FROM main | where amount NOT between 0 and 5",
		"FROM main | where amount not BETWEEN 0 AND 5",
		"FROM main | where a=1 and b=2",
		"FROM main | where a=1 AnD b=2",
		"FROM main | where a=1 or b=2",
		"FROM main | where a=1 xOr b=2",
		"FROM main | where nOt enabled",
	} {
		t.Run(text, func(t *testing.T) {
			p := parseSPL2Document(text)
			if p.syntaxComplete || len(p.diagnostics) == 0 {
				t.Fatal("unproved operator casing claimed complete")
			}
			for _, d := range p.diagnostics {
				if d.Severity == "error" || d.Code != CodeUnsupportedSemantics {
					t.Fatalf("held spelling became invalid: %+v", p.diagnostics)
				}
			}
		})
	}
	for _, text := range []string{
		"FROM main | eval and=or, x={and:or,not:xor}, y=round(and:or), z=and",
		"FROM main | where and=or",
		"FROM main | eval between=and, x={between:or}",
		"search and OR host=api",
		"search index=app and",
		"search index=app aNd host=api",
	} {
		t.Run(text, func(t *testing.T) {
			p := parseSPL2Document(text)
			if !p.syntaxComplete || len(p.diagnostics) > 0 {
				t.Fatalf("ordinary same-spelled name/literal changed meaning: %+v", p.diagnostics)
			}
			for _, kind := range []string{"logicalAnd", "logicalOr", "logicalXor", "logicalNot", "betweenOperator"} {
				if len(spl2Nodes(p.syntax, kind)) != 0 {
					t.Fatalf("name/literal became operator %s", kind)
				}
			}
		})
	}
	for _, text := range []string{"FROM main | where amount between 0 and", "FROM main | where amount between and 5"} {
		p := parseSPL2Document(text)
		found := false
		for _, d := range p.diagnostics {
			if d.Severity == "error" && d.Code == CodeSyntaxError {
				found = true
			}
		}
		if !found {
			t.Fatalf("malformed lowercase predicate was hidden by hold: %+v", p.diagnostics)
		}
	}
}

func TestSPL2ExpressionPrefixNotOwnership(t *testing.T) {
	cases := []struct{ expression, shape string }{
		{"not+1", `(additive IDENTIFIER:"not" PLUS:"+" NUMBER:"1")`},
		{"not[0]", `(access IDENTIFIER:"not" (accessPart LBRACKET:"[" NUMBER:"0" RBRACKET:"]"))`},
		{"not(1)", `(call IDENTIFIER:"not" LPAREN:"(" NUMBER:"1" RPAREN:")")`},
		{"not-1", `(additive IDENTIFIER:"not" MINUS:"-" NUMBER:"1")`},
		{"not.member", `(access IDENTIFIER:"not" (accessPart DOT:"." IDENTIFIER:"member"))`},
		{"nOt(1)", `(call IDENTIFIER:"nOt" LPAREN:"(" NUMBER:"1" RPAREN:")")`},
		{"not", `IDENTIFIER:"not"`},
	}
	for _, tc := range cases {
		t.Run(tc.expression, func(t *testing.T) {
			text := "FROM main | eval x=" + tc.expression
			p := parseSPL2Document(text)
			if !p.syntaxComplete || len(p.diagnostics) > 0 {
				t.Fatalf("ordinary identifier became held operator: %+v", p.diagnostics)
			}
			expressions := spl2Nodes(p.syntax, "expression")
			if len(expressions) == 0 || expressions[0].shape() != tc.shape {
				t.Fatalf("got %s want %s", p.syntax.shape(), tc.shape)
			}
			location := expressions[0].Location
			if text[location.Start.Offset:location.End.Offset] != tc.expression {
				t.Fatalf("expression source ownership %+v", location)
			}
			if len(spl2Nodes(p.syntax, "logicalNot")) != 0 || len(spl2Nodes(p.syntax, "array")) != 0 {
				t.Fatal("identifier stolen by prefix or array-literal context")
			}
		})
	}
	controls := []struct {
		expression, shape string
		held              bool
	}{
		{"nOt enabled", `(notExpression IDENTIFIER:"nOt" IDENTIFIER:"enabled")`, true},
		{"NOT enabled", `(notExpression NOT:"NOT" IDENTIFIER:"enabled")`, false},
		{"NOT a=1 AND b=2", `(andExpression (notExpression NOT:"NOT" (predicate IDENTIFIER:"a" ASSIGN:"=" NUMBER:"1")) AND:"AND" (predicate IDENTIFIER:"b" ASSIGN:"=" NUMBER:"2"))`, false},
		{"NOT not[0]", `(notExpression NOT:"NOT" (access IDENTIFIER:"not" (accessPart LBRACKET:"[" NUMBER:"0" RBRACKET:"]")))`, false},
	}
	for _, tc := range controls {
		t.Run(tc.expression, func(t *testing.T) {
			text := "FROM main | where " + tc.expression
			p := parseSPL2Document(text)
			if p.syntaxComplete == tc.held {
				t.Fatalf("prefix coverage changed %+v", p.diagnostics)
			}
			if tc.held {
				if len(p.diagnostics) != 1 || p.diagnostics[0].Code != CodeUnsupportedSemantics || p.diagnostics[0].Severity != "warning" {
					t.Fatalf("held prefix outcome %+v", p.diagnostics)
				}
			} else if len(p.diagnostics) != 0 {
				t.Fatalf("documented uppercase prefix changed %+v", p.diagnostics)
			}
			expression := spl2Nodes(p.syntax, "expression")[0]
			if expression.shape() != tc.shape {
				t.Fatalf("got %s want %s", expression.shape(), tc.shape)
			}
			location := expression.Location
			if text[location.Start.Offset:location.End.Offset] != tc.expression {
				t.Fatal("prefix source changed")
			}
			if len(spl2Nodes(p.syntax, "logicalNot")) != 1 {
				t.Fatalf("prefix operator ownership %s", p.syntax.shape())
			}
		})
	}
}

func TestSPL2SelectedDatasetParametersAndDottedPaths(t *testing.T) {
	parameter := spl2RequireNoDiagnostics(t, "FROM $target_1 | fields synthetic_value")
	if got := len(spl2Nodes(parameter.syntax, "datasetParameter")); got != 1 {
		t.Fatalf("dataset parameters %d: %s", got, parameter.syntax.shape())
	}
	param := spl2Nodes(parameter.syntax, "datasetParameter")[0]
	if text := parameter.source.text[param.Location.Start.Offset:param.Location.End.Offset]; text != "$target_1" {
		t.Fatalf("dataset parameter source %q", text)
	}

	paths := spl2RequireNoDiagnostics(t, "FROM catalog.events | eval leaf=payload.user.name | stats count() AS 'metrics.total' BY payload.region")
	if got := len(spl2Nodes(paths.syntax, "datasetPath")); got != 1 {
		t.Fatalf("dataset paths %d: %s", got, paths.syntax.shape())
	}
	if got := len(spl2Nodes(paths.syntax, "accessPart")); got != 3 {
		t.Fatalf("structural access parts %d: %s", got, paths.syntax.shape())
	}
	aliases := spl2Nodes(paths.syntax, "aggregateAlias")
	if len(aliases) != 1 || len(spl2Nodes(aliases[0], "accessPart")) != 0 {
		t.Fatalf("quoted dotted alias became structural: %s", paths.syntax.shape())
	}
	loc := aliases[0].Location
	if got := paths.source.text[loc.Start.Offset:loc.End.Offset]; got != "'metrics.total'" {
		t.Fatalf("atomic alias source %q", got)
	}
}

func TestSPL2SelectedTypedModuleSyntax(t *testing.T) {
	text := "@module(\"security\");\n" +
		"@source(\"catalog\") import events as source from acme/security/events;\n" +
		"import {users as identities, alerts} from acme/security;\n" +
		"import * as security from acme/security;\n" +
		"$view = FROM $target_1 | where source.enabled=true;\n" +
		"@memoized() function normalize($value, $fallback) {\n" +
		"  return coalesce($value, $fallback);\n" +
		"}\n" +
		"export view;\n" +
		"export {view as detections, normalize};"
	p := spl2RequireTypedModule(t, text)
	for kind, want := range map[string]int{
		"annotation":          3,
		"annotationStatement": 1,
		"importDeclaration":   3,
		"qualifiedName":       3,
		"viewDeclaration":     1,
		"functionDeclaration": 1,
		"functionParameter":   2,
		"returnStatement":     1,
		"exportDeclaration":   2,
	} {
		if got := len(spl2Nodes(p.syntax, kind)); got != want {
			t.Errorf("%s count %d want %d: %s", kind, got, want, p.syntax.shape())
		}
	}
	if got := len(spl2Nodes(p.syntax, "statementTerminator")); got != 8 {
		t.Fatalf("statement terminators %d want 8", got)
	}
}

func TestSPL2SelectedModuleBoundaries(t *testing.T) {
	for _, text := range []string{
		"$value = 1;",
		"dataset events = {kind:\"index\"};",
		"namespace security;",
		"function query_rows($value) { return FROM events; }",
		"function mutate($value) { $local=$value; return $local; }",
		"$sink = FROM events | into output;",
		"import events from acme/security",
		"export events;;",
		"export function helper($value) { return $value; }",
	} {
		t.Run(text, func(t *testing.T) {
			p := spl2RequireLocatedError(t, text)
			foundSyntax := false
			for _, diagnostic := range p.diagnostics {
				foundSyntax = foundSyntax || diagnostic.Code == CodeSyntaxError && diagnostic.Category == "syntax"
			}
			if !foundSyntax {
				t.Fatal("malformed statement received only the generic module boundary")
			}
		})
	}
}

func TestSPL2SelectedStatsLookaheadRespectsLexerBudget(t *testing.T) {
	text := "FROM main | stats count() BY " + strings.Repeat("x+", lexerWorkLimit*8) + "x"
	input := &spl2CountingCharStream{CharStream: antlr.NewInputStream(text)}
	source := newSourceIndex(text)
	parsed := &spl2ParsedDocument{source: source, diagnostics: []Diagnostic{}}
	tracker := &lexerWorkTracker{}
	lexer := spl2.NewSPL2Lexer(input)
	lexer.RemoveErrorListeners()
	lexer.AddErrorListener(&spl2SyntaxListener{DefaultErrorListener: antlr.NewDefaultErrorListener(), parsed: parsed, tracker: tracker})
	parsed.tokens = preflightLexer(lexer, tracker, source)
	if tracker.resourceLimit == nil || tracker.units != lexerWorkLimit {
		t.Fatalf("dense stats grouping bypassed lexer admission: units=%d limit=%+v", tracker.units, tracker.resourceLimit)
	}
	if input.lookaheads > lexerWorkLimit*12 {
		t.Fatalf("stats discriminator performed %d lookaheads before the 4,097th token rejection", input.lookaheads)
	}
	result := mustAnalyzeDocument(t, QueryDocument{Text: text, Language: "spl2"})
	assertSingleResourceLimitLocation(t, result, *tracker.resourceLimit)
}

func TestSPL2SelectedStatsLookaheadStopsWhenLexicalErrorConsumesBudget(t *testing.T) {
	tail := " " + strings.Repeat("x ", 2047) + "\x00y " + strings.Repeat("z ", lexerWorkLimit*16)
	text := "FROM main | stats count() BY" + tail
	input := &spl2CountingCharStream{CharStream: antlr.NewInputStream(text)}
	source := newSourceIndex(text)
	parsed := &spl2ParsedDocument{source: source, diagnostics: []Diagnostic{}}
	tracker := &lexerWorkTracker{}
	lexer := spl2.NewSPL2Lexer(input)
	lexer.RemoveErrorListeners()
	lexer.AddErrorListener(&spl2SyntaxListener{DefaultErrorListener: antlr.NewDefaultErrorListener(), parsed: parsed, tracker: tracker})
	parsed.tokens = preflightLexer(lexer, tracker, source)
	if tracker.resourceLimit == nil || tracker.units != lexerWorkLimit {
		t.Fatalf("lexical-error stats grouping bypassed lexer admission: units=%d limit=%+v", tracker.units, tracker.resourceLimit)
	}
	if input.lookaheads > lexerWorkLimit*12 {
		t.Fatalf("stats discriminator performed %d lookaheads after a lexical error consumed unit 4,096", input.lookaheads)
	}
	result := mustAnalyzeDocument(t, QueryDocument{Text: text, Language: "spl2"})
	assertSingleResourceLimitLocation(t, result, *tracker.resourceLimit)
}

func TestSPL2SelectedStatsNestedBracketLookaheadSharesBudget(t *testing.T) {
	for _, depth := range []int{11, 20} {
		t.Run(fmt.Sprintf("depth-%d", depth), func(t *testing.T) {
			tail := "[stats, BY+1]"
			for range depth {
				tail = "[" + tail + ", stats, BY+1]"
			}
			text := "FROM events | eval x=" + tail
			spl2RequireNoDiagnostics(t, text)

			input := &spl2CountingCharStream{CharStream: antlr.NewInputStream(text)}
			source := newSourceIndex(text)
			parsed := &spl2ParsedDocument{source: source, diagnostics: []Diagnostic{}}
			tracker := &lexerWorkTracker{}
			lexer := spl2.NewSPL2Lexer(input)
			lexer.RemoveErrorListeners()
			lexer.AddErrorListener(&spl2SyntaxListener{DefaultErrorListener: antlr.NewDefaultErrorListener(), parsed: parsed, tracker: tracker})
			preflightLexer(lexer, tracker, source)
			if tracker.resourceLimit != nil || len(parsed.diagnostics) != 0 {
				t.Fatalf("valid nested array exhausted lexer work: limit=%+v diagnostics=%+v", tracker.resourceLimit, parsed.diagnostics)
			}
			if input.lookaheads > lexerWorkLimit*12 {
				t.Fatalf("nested bracket classification performed %d lookaheads", input.lookaheads)
			}
		})
	}
}

func TestSPL2SelectedStatsBelowLimitSubpipeSharesDocumentProbe(t *testing.T) {
	text := "FROM events | appendpipe [stats count() BY " + strings.Repeat("x+", 1049) + "x]"
	spl2RequireNoDiagnostics(t, text)

	input := &spl2CountingCharStream{CharStream: antlr.NewInputStream(text)}
	source := newSourceIndex(text)
	parsed := &spl2ParsedDocument{source: source, diagnostics: []Diagnostic{}}
	tracker := &lexerWorkTracker{}
	lexer := spl2.NewSPL2Lexer(input)
	lexer.RemoveErrorListeners()
	lexer.AddErrorListener(&spl2SyntaxListener{DefaultErrorListener: antlr.NewDefaultErrorListener(), parsed: parsed, tracker: tracker})
	preflightLexer(lexer, tracker, source)
	if tracker.resourceLimit != nil || tracker.units != 2117 || len(parsed.diagnostics) != 0 {
		t.Fatalf("below-limit subpipe admission: units=%d limit=%+v diagnostics=%+v", tracker.units, tracker.resourceLimit, parsed.diagnostics)
	}
	if input.lookaheads > lexerWorkLimit*12 {
		t.Fatalf("below-limit subpipe classification performed %d lookaheads", input.lookaheads)
	}
}

func TestSPL2SelectedStatsExactLexerBudgetRemainsClassified(t *testing.T) {
	for _, tc := range []struct {
		name string
		text string
	}{
		{"top-level", "FROM events | stats count() BY " + strings.Repeat("x+", 2040) + "x "},
		{"inherited-subpipe", "FROM events | appendpipe [stats count() BY " + strings.Repeat("x+", 2038) + "x] "},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source := newSourceIndex(tc.text)
			parsed := &spl2ParsedDocument{source: source, diagnostics: []Diagnostic{}}
			tracker := &lexerWorkTracker{}
			lexer := spl2.NewSPL2Lexer(antlr.NewInputStream(tc.text))
			lexer.RemoveErrorListeners()
			lexer.AddErrorListener(&spl2SyntaxListener{DefaultErrorListener: antlr.NewDefaultErrorListener(), parsed: parsed, tracker: tracker})
			preflightLexer(lexer, tracker, source)
			if tracker.resourceLimit != nil || tracker.units != lexerWorkLimit || len(parsed.diagnostics) != 0 {
				t.Fatalf("exact-limit stats admission: units=%d limit=%+v diagnostics=%+v", tracker.units, tracker.resourceLimit, parsed.diagnostics)
			}
			spl2RequireNoDiagnostics(t, tc.text)
		})
	}
}

func TestSPL2SelectedStatsProbeErrorsAreRangeLocal(t *testing.T) {
	for _, tc := range []struct {
		name string
		text string
	}{
		{"top-level-later-error", "FROM events | stats count() BY bytes+delta | eval x=1 \x00"},
		{"inherited-subpipe-later-error", "FROM events | appendpipe [stats count() BY bytes+delta] | eval x=1 \x00"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			parsed := parseSPL2Document(tc.text)
			if len(parsed.lexicalErrors) != 1 || len(parsed.diagnostics) != 1 {
				t.Fatalf("later lexical error changed parser ownership: %+v", parsed.diagnostics)
			}
			if got := len(spl2Nodes(parsed.syntax, "selectedAggregateGroup")); got != 1 {
				t.Fatalf("earlier selected group lost classification: got %d, syntax=%s", got, parsed.syntax.shape())
			}
		})
	}

	for _, tc := range []struct {
		name string
		text string
	}{
		{"top-level-group-error", "FROM events | stats count() BY bytes+\x00delta"},
		{"inherited-subpipe-group-error", "FROM events | appendpipe [stats count() BY bytes+\x00delta]"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			parsed := parseSPL2Document(tc.text)
			if len(parsed.lexicalErrors) != 1 || len(spl2Nodes(parsed.syntax, "selectedAggregateGroup")) != 0 {
				t.Fatalf("damaged group was classified as selected: diagnostics=%+v syntax=%s", parsed.diagnostics, parsed.syntax.shape())
			}
		})
	}
}

func TestSPL2SelectedStatsLongLowTokenTailsRemainAdmitted(t *testing.T) {
	for _, tc := range []struct {
		name   string
		query  string
		groups []string
	}{
		{"whitespace", "FROM main | stats count() BY " + strings.Repeat(" ", 5000) + "bytes+delta", []string{"bytes", "delta"}},
		{"literal", `FROM main | stats count() BY if(ready, "` + strings.Repeat("x", 5000) + `", "other")`, []string{"ready"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			spl2RequireNoDiagnostics(t, tc.query)
			result := spl2AnalyzeTest(t, tc.query)
			if result.Status != Valid || !result.Coverage.SyntaxComplete || !result.Coverage.SemanticComplete || !result.Requirements.Coverage.Complete {
				t.Fatalf("low-token selected grouping was not completely analyzed: %+v", result)
			}
			if spl2HasCode(result, CodeAnalysisResourceLimit) || spl2HasCode(result, CodeUnsupportedSemantics) {
				t.Fatalf("low-token selected grouping retained a temporary limitation: %+v", result.Diagnostics)
			}
			groupRefs := map[string]bool{}
			for _, ref := range result.References {
				if ref.Role == "group" && ref.Binding == "source" && ref.Resolution == "exact" {
					groupRefs[ref.NormalizedName] = true
				}
			}
			groupRequirements := map[string]bool{}
			for _, item := range result.Requirements.Items {
				if item.Role == "group" && item.Resolution == "exact" {
					groupRequirements[item.Identity] = true
				}
			}
			for _, group := range tc.groups {
				if !groupRefs[group] || !groupRequirements[group] {
					t.Fatalf("missing exact group evidence for %q: refs=%+v requirements=%+v", group, result.References, result.Requirements)
				}
			}
			after := result.Lineage[len(result.Lineage)-1].After
			if after.Open || after.Uncertain || len(after.Fields) != 2 {
				t.Fatalf("group plus aggregate output is not closed: %+v", after)
			}
			count := false
			for _, field := range after.Fields {
				count = count || field.Name == "count"
			}
			if !count {
				t.Fatalf("closed output lacks aggregate: %+v", after)
			}
		})
	}
}
