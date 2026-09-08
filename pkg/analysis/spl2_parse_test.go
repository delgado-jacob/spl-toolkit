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
