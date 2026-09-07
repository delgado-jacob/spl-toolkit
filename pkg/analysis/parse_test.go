package analysis

import (
	"github.com/antlr4-go/antlr/v4"
	"github.com/delgado-jacob/spl-toolkit/parser"
	"strings"
	"testing"
)

// Grammar acceptance must cover structured arguments without declaring their semantics complete.
func TestParseStructuredStages(t *testing.T) {
	for _, q := range []string{
		`search host=web status IN (200,404) | eval x=if(isnull(user),lower(host),user), y=1+2*3 | where x=host AND NOT (y<2 OR match(x,"a"))`,
		`index=main | stats count, sum(bytes) AS total BY user, host`,
		`search source=/var/log/app.log src_ip=10.2.3.4 host=web* | rename user AS person, host AS node | fields + person node* | table person,node*`,
		`search * | lookup people uid AS user OUTPUT name AS display, age OUTPUTNEW group AS team`,
		`search * | sort 0 -bytes +host | dedup 2 user host | head 10 | tail 2`,
		"search * `filter(host)` | append [ search index=child | eval x=1 ] | appendpipe [ stats count ]",
		`search * | mystery alpha=beta [ search source="child" ] | table host`,
		`search café="é" | eval 'café'=lower('hôte') . "!"`,
	} {
		t.Run(q, func(t *testing.T) {
			r, e := Analyze(QueryDocument{Text: q})
			if e != nil {
				t.Fatal(e)
			}
			if !r.Coverage.SyntaxComplete {
				t.Fatalf("syntax: %+v", r.Diagnostics)
			}
			if r.Status != Incomplete || r.Coverage.SemanticComplete {
				t.Fatal(r.Status, r.Coverage)
			}
			if len(r.Stages) == 0 {
				t.Fatal("missing stages")
			}
		})
	}
}
func TestParseInvalidAndRecovery(t *testing.T) {
	for _, q := range []string{"", " \r\n\t ", "|", "search host=web | eval broken= | stats count by user", "search host=web | where @@@ | table user", "search host=\"unterminated"} {
		t.Run(q, func(t *testing.T) {
			r, e := Analyze(QueryDocument{Text: q})
			if e != nil {
				t.Fatal(e)
			}
			if r.Status != Invalid || r.Coverage.SyntaxComplete || r.Coverage.SemanticComplete {
				t.Fatal(r.Status, r.Coverage)
			}
			found := false
			for _, d := range r.Diagnostics {
				found = found || d.Category == "syntax"
			}
			if !found {
				t.Fatal(r.Diagnostics)
			}
			if strings.Contains(q, "broken=") {
				if len(r.Stages) != 3 || r.Stages[2].Command != "stats" {
					t.Fatal(r.Stages)
				}
			}
		})
	}
}
func TestParseTypedContexts(t *testing.T) {
	p := parseDocument(`search host=web | eval x=if(user=host,1,2), y=3 | lookup people uid AS user OUTPUT name AS display | stats sum(x) AS total BY user`)
	counts := map[string]int{}
	var walk func(antlr.Tree)
	walk = func(n antlr.Tree) {
		switch ctx := n.(type) {
		case *parser.AnalysisAssignmentContext:
			counts["assignment"]++
		case *parser.AnalysisFunctionCallContext:
			counts["function"]++
		case *parser.AnalysisComparisonContext:
			if ctx.AnalysisComparisonOperator() != nil {
				counts["comparison"]++
			}
		case *parser.AnalysisAliasContext:
			counts["alias"]++
		case *parser.AnalysisOutputContext:
			counts["output"]++
		case *parser.AnalysisGroupContext:
			counts["group"]++
		}
		for i := 0; i < n.GetChildCount(); i++ {
			walk(n.GetChild(i))
		}
	}
	walk(p.tree)
	for k, want := range map[string]int{"assignment": 2, "function": 2, "comparison": 1, "alias": 3, "output": 1, "group": 1} {
		if counts[k] != want {
			t.Fatalf("%s count = %d want %d", k, counts[k], want)
		}
	}
}
func TestParseLegacyTokenMeaning(t *testing.T) {
	for _, tc := range []struct {
		q    string
		want []int
	}{
		{"10.2.3.4", []int{parser.SPLLexerNUMBER, parser.SPLLexerIDENTIFIER}},
		{"/var/log/app.log", []int{parser.SPLLexerDIV, parser.SPLLexerFUNCTION, parser.SPLLexerDIV, parser.SPLLexerFUNCTION, parser.SPLLexerDIV, parser.SPLLexerIDENTIFIER}},
		{"host*", []int{parser.SPLLexerIDENTIFIER, parser.SPLLexerMULT}},
	} {
		l := parser.NewSPLLexer(antlr.NewInputStream(tc.q))
		var got []int
		for tok := l.NextToken(); tok.GetTokenType() != antlr.TokenEOF; tok = l.NextToken() {
			got = append(got, tok.GetTokenType())
		}
		if len(got) != len(tc.want) {
			t.Fatalf("%s tokens %v", tc.q, got)
		}
		for i := range got {
			if got[i] != tc.want[i] {
				t.Fatalf("%s tokens %v want %v", tc.q, got, tc.want)
			}
		}
	}
}

// Signs on whitespace-separated sort fields must not become arithmetic operands;
// command options and catalog names must not become local event-field reads.
func TestParseCommandArgumentRoles(t *testing.T) {
	p := parseDocument(`search * | sort 0 -bytes +host | dedup 2 keepempty=true user host | inputlookup append=true people | head 10`)
	if len(p.diagnostics) != 0 {
		t.Fatal(p.diagnostics)
	}
	var sorts, options, catalogs, limits int
	var walk func(antlr.Tree)
	walk = func(n antlr.Tree) {
		switch n.(type) {
		case *parser.AnalysisSortFieldContext:
			sorts++
		case *parser.AnalysisOptionContext:
			options++
		case *parser.AnalysisCatalogNameContext:
			catalogs++
		case *parser.AnalysisLimitContext:
			limits++
		}
		for i := 0; i < n.GetChildCount(); i++ {
			walk(n.GetChild(i))
		}
	}
	walk(p.tree)
	if sorts != 2 || options != 2 || catalogs != 1 || limits != 3 {
		t.Fatalf("sort=%d option=%d catalog=%d limit=%d", sorts, options, catalogs, limits)
	}
}

func TestParseModeledMalformedDoesNotFallBack(t *testing.T) {
	for _, q := range []string{`search * | eval`, `search * | eval x=`, `search * | where`, `search * | rename user`, `search * | fields`, `search * | stats sum(`, `search * | lookup people`, `search * | sort`, `search * | dedup`, `search * | inputlookup`, `search * | eval x=$`} {
		r, e := Analyze(QueryDocument{Text: q})
		if e != nil {
			t.Fatal(e)
		}
		if r.Coverage.SyntaxComplete || r.Status != Invalid {
			t.Fatalf("%s accepted: %+v", q, r)
		}
	}
}
func TestParseImplicitSearchLiteralAndPrecedence(t *testing.T) {
	p := parseDocument(`error`)
	if len(p.diagnostics) != 0 {
		t.Fatal(p.diagnostics)
	}
	initial := p.tree.AnalysisPipeline().AnalysisInitialStage()
	if initial.AnalysisImplicitSearch() == nil || initial.AnalysisImplicitSearch().AnalysisSearch().GetText() != "error" {
		t.Fatal("bare term lost its search role")
	}
	p = parseDocument(`| eval answer=1+2*3 . "x"`)
	if len(p.diagnostics) != 0 {
		t.Fatal(p.diagnostics)
	}
	stage := p.tree.AnalysisPipeline().AnalysisStage(0).(*parser.AnalysisEvalStageContext)
	concat := stage.AnalysisAssignment(0).AnalysisExpression().AnalysisOr().AnalysisAnd(0).AnalysisNot(0).AnalysisComparison().AnalysisConcat(0)
	if len(concat.AllAnalysisAdd()) != 2 {
		t.Fatal("concatenation precedence")
	}
	add := concat.AnalysisAdd(0)
	if len(add.AllAnalysisMultiply()) != 2 || len(add.AnalysisMultiply(1).AllAnalysisPower()) != 2 {
		t.Fatal("multiplication precedence")
	}
}
func TestParseScopePreorderAndSpans(t *testing.T) {
	text := `  search * | append [ search index=child | stats count ] | table host  `
	r, e := Analyze(QueryDocument{Text: text})
	if e != nil {
		t.Fatal(e)
	}
	if !r.Coverage.SyntaxComplete {
		t.Fatal(r.Diagnostics)
	}
	if len(r.Scopes) != 2 || len(r.Stages) != 5 {
		t.Fatal(r.Scopes, r.Stages)
	}
	root, child := r.Scopes[0], r.Scopes[1]
	if root.ID != "scope-0" || root.Location.Start.Offset != 0 || root.Location.End.Offset != len(text) {
		t.Fatal(root)
	}
	if child.ID != "scope-1" || child.ParentID != root.ID || child.StageID != "stage-1" || child.Kind != "append" || text[child.Location.Start.Offset:child.Location.End.Offset] != `[ search index=child | stats count ]` {
		t.Fatal(child)
	}
	for i, want := range []struct {
		command, scope, span string
		position             int
	}{{"search", "scope-0", "search *", 0}, {"append", "scope-0", "append [ search index=child | stats count ]", 1}, {"search", "scope-1", "search index=child", 0}, {"stats", "scope-1", "stats count", 1}, {"table", "scope-0", "table host", 2}} {
		stage := r.Stages[i]
		if stage.Command != want.command || stage.ScopeID != want.scope || stage.Position != want.position || text[stage.Location.Start.Offset:stage.Location.End.Offset] != want.span {
			t.Fatal(stage)
		}
	}
}
