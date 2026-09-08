package analysis

import "testing"

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
	for _, text := range []string{"$saved = FROM main;", "import foo", "function f() {}", "FROM main; FROM other"} {
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
