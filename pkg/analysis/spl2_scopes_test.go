package analysis

import "testing"

func TestSPL2ScopesTypedOwnership(t *testing.T) {
	query := `FROM main | eval a=[{key:"[|]"}], b=($x)->{return $x;} | append [FROM other | appendpipe [eval nested=1]] | if (flag=true) [fields a] else [where flag=false]`
	p := parseSPL2Document(query)
	if !p.syntaxComplete {
		t.Fatalf("syntax: %+v", p.diagnostics)
	}
	// Typed child contexts, not lexical bracket counting, establish four scopes.
	if got := len(spl2Nodes(p.syntax, "independentSearch")); got != 1 {
		t.Fatalf("independent children %d", got)
	}
	if got := len(spl2Nodes(p.syntax, "inheritedSubpipe")); got != 3 {
		t.Fatalf("inherited children %d", got)
	}
}

func TestSPL2ScopesDescriptors(t *testing.T) {
	query := "FROM [{name:\"é\"}]\r\n | append [FROM other | appendpipe [eval n=1]] | if (true) [fields n] else [where n=1]"
	p := parseSPL2Document(query)
	scopes := spl2ChildScopes(p)
	want := []struct {
		text, input string
		parent      int
	}{
		{`[FROM other | appendpipe [eval n=1]]`, "independent", -1},
		{`[eval n=1]`, "inherited", 0},
		{`[fields n]`, "inherited", -1},
		{`[where n=1]`, "inherited", -1},
	}
	if len(scopes) != len(want) {
		t.Fatalf("scopes: %+v diagnostics %+v", scopes, p.diagnostics)
	}
	for i, s := range scopes {
		if query[s.location.Start.Offset:s.location.End.Offset] != want[i].text || s.input != want[i].input || s.parent != want[i].parent || s.owner == nil || s.body == nil {
			t.Fatalf("scope %d %+v", i, s)
		}
		if s.location.Start.Line != 2 {
			t.Fatal("lost CRLF location")
		}
	}
}

func TestSPL2ScopesSQLChildAndBracketRoles(t *testing.T) {
	query := `FROM main AS m WHERE EXISTS(SELECT c.id FROM child AS c WHERE c.id=m.id) | union [{name:"[FROM fake]"}], [FROM other | eval a=[1,2], b=($x)->{return $x;} ]`
	p := parseSPL2Document(query)
	scopes := spl2ChildScopes(p)
	if len(scopes) != 2 || scopes[0].kind != "exists" || scopes[0].input != "correlated" || scopes[1].kind != "search" {
		t.Fatalf("scopes %+v diagnostics %+v", scopes, p.diagnostics)
	}
	if query[scopes[0].location.Start.Offset:scopes[0].location.End.Offset] != `EXISTS(SELECT c.id FROM child AS c WHERE c.id=m.id)` {
		t.Fatal("SQL child scope rewritten")
	}
	if len(spl2Nodes(p.syntax, "array")) != 2 || len(spl2Nodes(p.syntax, "lambdaBlock")) != 1 {
		t.Fatal("compound literal scopes confused")
	}
}

func TestSPL2ScopesRecovery(t *testing.T) {
	for _, tt := range []struct {
		query string
		count int
	}{
		{`FROM main | append [FROM other | eval x=]`, 0},
		{`FROM main | append [FROM other] | where`, 1},
		{`FROM main | if (true) [fields]`, 0},
		{`FROM main | eval a=["[FROM x]"], b={key:"[eval x=1]"}`, 0},
	} {
		t.Run(tt.query, func(t *testing.T) {
			p := parseSPL2Document(tt.query)
			scopes := spl2ChildScopes(p)
			if len(scopes) != tt.count {
				t.Fatalf("scopes %+v diagnostics %+v", scopes, p.diagnostics)
			}
			for _, scope := range scopes {
				if scope.owner.GetStart().GetTokenIndex() < 0 || scope.owner.GetStop().GetTokenIndex() < 0 {
					t.Fatal("invented scope token")
				}
			}
		})
	}
}
