package analysis

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/antlr4-go/antlr/v4"
	"github.com/delgado-jacob/spl-toolkit/parser/spl2"
)

// Fixtures are repository-owned source evidence, independent of parse output.
func spl2PipelineFixture(t *testing.T, id string) *spl2ParsedDocument {
	t.Helper()
	for _, file := range []string{"pipeline-commands.json", "pipeline-boundaries.json"} {
		data, err := os.ReadFile(filepath.Join("..", "..", "testdata", "spl2", file))
		if err != nil {
			t.Fatal(err)
		}
		var cases []spl2CorpusCase
		if err := json.Unmarshal(data, &cases); err != nil {
			t.Fatal(err)
		}
		for _, c := range cases {
			for _, alias := range c.ObligationIDs {
				if alias == id {
					p := parseSPL2Document(c.Document.Text)
					for _, d := range p.diagnostics {
						if c.Status != "invalid" && d.Severity == "error" {
							t.Fatalf("%s: %+v", id, p.diagnostics)
						}
					}
					return p
				}
			}
		}
	}
	t.Fatalf("missing fixture %s", id)
	return nil
}

func spl2PipelineText(t *testing.T, p *spl2ParsedDocument, ctx antlr.ParserRuleContext, want string) {
	t.Helper()
	if ctx == nil {
		t.Fatalf("missing typed operand %q", want)
	}
	loc := p.source.contextLocation(ctx)
	if got := p.source.text[loc.Start.Offset:loc.End.Offset]; got != want {
		t.Errorf("operand %q want %q", got, want)
	}
}

func TestSPL2PipelineAssignmentOwnership(t *testing.T) {
	p := spl2PipelineFixture(t, "T3.typed.assign")
	assignments := p.tree.Pipeline().Command(0).EvalCommand().AllAssignment()
	if len(assignments) != 2 {
		t.Fatalf("assignments %d", len(assignments))
	}
	for i, want := range []struct{ target, expression string }{{"x", "bytes"}, {"y", "x+1"}} {
		spl2PipelineText(t, p, assignments[i].FieldName(), want.target)
		spl2PipelineText(t, p, assignments[i].Expression(), want.expression)
	}
}

func TestSPL2PipelineSelectorIntent(t *testing.T) {
	for _, c := range []struct {
		id     string
		sign   int
		fields []string
	}{
		{"C05-P1", spl2.SPL2ParserPLUS, []string{"server", "account"}},
		{"C05-P2", spl2.SPL2ParserMINUS, []string{"'scratch*'", "debug"}},
		{"E.C05.include.P1", 0, []string{"user", "bytes"}},
	} {
		p := spl2PipelineFixture(t, c.id)
		selection := p.tree.Pipeline().Command(0).FieldsCommand().FieldSelection()
		if (selection.PLUS() != nil) != (c.sign == spl2.SPL2ParserPLUS) || (selection.MINUS() != nil) != (c.sign == spl2.SPL2ParserMINUS) {
			t.Fatal("lost list-wide intent", c.id)
		}
		fields := selection.AllFieldSelector()
		if len(fields) != len(c.fields) {
			t.Fatalf("fields %d", len(fields))
		}
		for i, want := range c.fields {
			spl2PipelineText(t, p, fields[i].Identifier(), want)
		}
	}
	p := spl2PipelineFixture(t, "C06-P2")
	fields := p.tree.Pipeline().Command(0).TableCommand().AllTableField()
	if len(fields) != 1 || fields[0].Identifier() == nil {
		t.Fatal("table exact field ownership")
	}
	spl2PipelineText(t, p, fields[0].Identifier(), "'remote-address'")
}

func TestSPL2PipelineLookupOwnership(t *testing.T) {
	p := spl2PipelineFixture(t, "T3.typed.lookup")
	lookup := p.tree.Pipeline().Command(0).LookupCommand()
	spl2PipelineText(t, p, lookup.LookupDataset(), "people")
	matches, outputs := lookup.AllLookupMatch(), lookup.LookupOutputClause().AllLookupOutput()
	if len(matches) != 2 || len(outputs) != 2 || lookup.LookupOutputClause().OUTPUT() == nil {
		t.Fatal("lookup list ownership")
	}
	for i, want := range []struct{ column, event string }{{"key", "'café'"}, {"realm", "domain"}} {
		spl2PipelineText(t, p, matches[i].LookupColumn(), want.column)
		spl2PipelineText(t, p, matches[i].LookupEventField(), want.event)
	}
	loc := p.source.contextLocation(matches[0].LookupEventField())
	if loc.Start.Line != 2 || loc.End.Offset-loc.Start.Offset != 7 {
		t.Fatalf("Unicode/CRLF source %+v", loc)
	}
	spl2PipelineText(t, p, outputs[0].LookupColumn(), "title")
	spl2PipelineText(t, p, outputs[0].LookupEventField(), "role")
	spl2PipelineText(t, p, outputs[1].LookupColumn(), "team")
	if outputs[1].LookupEventField() != nil {
		t.Fatal("implicit output acquired an alias")
	}
	p = spl2PipelineFixture(t, "C11-P2")
	if p.tree.Pipeline().Command(0).LookupCommand().LookupOutputClause().OUTPUTNEW() == nil {
		t.Fatal("OUTPUTNEW intent lost")
	}
	p = spl2PipelineFixture(t, "E.C11.matches.P1")
	if p.tree.Pipeline().Command(0).LookupCommand().LookupOutputClause() != nil || p.semanticComplete {
		t.Fatal("absent catalog output shape invented")
	}
}

func TestSPL2PipelineAggregateAndResetOwnership(t *testing.T) {
	p := spl2PipelineFixture(t, "T3.stats.options")
	stats := p.tree.Pipeline().Command(0).StatsCommand()
	options := stats.AllStatsOption()
	if len(options) != 3 {
		t.Fatalf("stats options %d", len(options))
	}
	spl2PipelineText(t, p, options[0].AllnumOption(), "allnum=false")
	spl2PipelineText(t, p, options[1].DelimOption(), `delim=":"`)
	spl2PipelineText(t, p, options[2].PartitionsOption(), "partitions=2.5")
	spl2PipelineText(t, p, stats.Aggregate(0).AggregateAlias(), "n")
	spl2PipelineText(t, p, stats.AggregateGroup().GroupField(0).Identifier(), "host")
	p = spl2PipelineFixture(t, "E.C09.by_span.P2")
	groups := p.tree.Pipeline().Command(0).EventstatsCommand().AggregateGroup().AllGroupField()
	if len(groups) != 2 {
		t.Fatalf("group fields %d", len(groups))
	}
	spl2PipelineText(t, p, groups[0].Identifier(), "event_time")
	spl2PipelineText(t, p, groups[0].GroupSpan().TimeSpan(), "2hr")
	spl2PipelineText(t, p, groups[1].Identifier(), "host")
	p = spl2PipelineFixture(t, "T3.typed.stream")
	stream := p.tree.Pipeline().Command(0).StreamstatsCommand()
	spl2PipelineText(t, p, stream.StreamGroup(), "BY host")
	spl2PipelineText(t, p, stream.CurrentOption(), "current=false")
	reset := stream.ResetClause()
	spl2PipelineText(t, p, reset.ResetBefore().Expression(), "code=500")
	spl2PipelineText(t, p, reset.ResetAfter().Expression(), `action="STOP"`)
	spl2PipelineText(t, p, reset.ResetOnchange(), "onchange")
	spl2PipelineText(t, p, stream.WindowOption(), "window=3")
	spl2PipelineText(t, p, stream.Aggregate(0).Call(), "sum(bytes)")
	spl2PipelineText(t, p, stream.Aggregate(0).AggregateAlias(), "total")
	if stream.StreamPostLayout() != nil {
		t.Fatal("normative options became postaggregate")
	}
}

func TestSPL2PipelineSortAndHeadOwnership(t *testing.T) {
	p := spl2PipelineFixture(t, "C12-P1")
	sort := p.tree.Pipeline().Command(0).SortCommand()
	spl2PipelineText(t, p, sort.IntegerValue(), "0")
	terms := sort.AllSortTerm()
	if len(terms) != 2 || terms[0].MINUS() == nil || terms[1].PLUS() == nil {
		t.Fatal("per-term signs lost")
	}
	spl2PipelineText(t, p, terms[0].SortWrapper(), "num")
	spl2PipelineText(t, p, terms[0].Identifier(), "size")
	spl2PipelineText(t, p, terms[1].Identifier(), "server")
	p = spl2PipelineFixture(t, "T3.dedup.options")
	dedup := p.tree.Pipeline().Command(0).DedupCommand()
	spl2PipelineText(t, p, dedup.IntegerValue(), "2")
	spl2PipelineText(t, p, dedup.KeepemptyOption(), "keepempty=false")
	spl2PipelineText(t, p, dedup.ConsecutiveOption(), "consecutive=true")
	if len(dedup.AllDedupField()) != 2 {
		t.Fatal("dedup list lost")
	}
	p = spl2PipelineFixture(t, "C14-P2")
	head := p.tree.Pipeline().Command(0).HeadCommand()
	spl2PipelineText(t, p, head.KeeplastOption(), "keeplast=false")
	spl2PipelineText(t, p, head.HeadWhile().Expression(), "size<100")
	spl2PipelineText(t, p, head.IntegerValue(), "9")
}

func TestSPL2PipelineRenameAndUnknownOptionOwnership(t *testing.T) {
	p := spl2PipelineFixture(t, "C07-P1")
	pairs := p.tree.Pipeline().Command(0).RenameCommand().AllRenamePair()
	if len(pairs) != 2 {
		t.Fatal("rename pairs lost")
	}
	for i, want := range []struct{ source, target string }{{"server", "machine"}, {"account", "actor"}} {
		spl2PipelineText(t, p, pairs[i].RenameSource(), want.source)
		spl2PipelineText(t, p, pairs[i].RenameTarget(), want.target)
	}
	p = spl2PipelineFixture(t, "T3.stats.unknown")
	option := p.tree.Pipeline().Command(0).StatsCommand().StatsOption(0).UnknownOption()
	spl2PipelineText(t, p, option, "unproved=true")
	if p.syntaxComplete || p.semanticComplete {
		t.Fatal("unknown option acquired complete coverage")
	}
	p = spl2PipelineFixture(t, "T3.stats.repeated.option.0")
	options := p.tree.Pipeline().Command(0).StatsCommand().AllStatsOption()
	if len(options) != 2 {
		t.Fatal("repeated options silently discarded")
	}
	spl2PipelineText(t, p, options[0].AllnumOption(), "allnum=true")
	spl2PipelineText(t, p, options[1].AllnumOption(), "allnum=false")
}

func TestSPL2PipelineRecoveryContracts(t *testing.T) {
	for _, id := range []string{"T3.recovery.contract.0", "T3.recovery.contract.1", "T3.recovery.contract.2"} {
		p := spl2PipelineFixture(t, id)
		if p.syntaxComplete || len(p.diagnostics) == 0 {
			t.Fatal("malformed supported command became complete", id)
		}
		for _, diagnostic := range p.diagnostics {
			if diagnostic.Category == "contract" {
				t.Fatalf("%s: recovered operand produced derivative contract finding: %+v", id, diagnostic)
			}
		}
	}
}

func TestSPL2PipelineDynamicTableRemainsUnproved(t *testing.T) {
	p := spl2PipelineFixture(t, "T3.fix1.dynamic.star")
	if p.syntaxComplete {
		t.Fatal("dynamic table value became a static field")
	}
	for _, d := range p.diagnostics {
		if strings.HasPrefix(d.Message, "H01") {
			t.Fatal("interpolated expression classified as a static wildcard")
		}
	}
}

func TestSPL2PipelineSearchLiteralOwnership(t *testing.T) {
	for _, c := range []struct {
		id, value string
		signed    bool
	}{
		{"T3.fix1.search.true", "true", false}, {"T3.fix1.search.false", "false", false},
		{"T3.fix1.search.minus", "-1", true}, {"T3.fix1.search.plus", "+1", true},
	} {
		p := spl2PipelineFixture(t, c.id)
		atoms := p.tree.Pipeline().Start_().SearchCommand().SearchExpression().SearchXor().SearchAnd(0).AllSearchOr()
		value := atoms[1].SearchNot(0).SearchAtom().SearchValue(0)
		if c.signed {
			spl2PipelineText(t, p, value.SearchSignedNumber(), c.value)
			if p.syntaxComplete || len(p.diagnostics) != 1 {
				t.Fatal("signed search literal lost its limitation")
			}
		} else {
			spl2PipelineText(t, p, value.SearchWordLiteral(), c.value)
			if !p.syntaxComplete || len(p.diagnostics) != 0 {
				t.Fatal("literal word classified as unproved")
			}
		}
		if len(spl2Nodes(p.syntax, "expression")) != 0 || len(spl2Nodes(p.syntax, "access")) != 0 || p.semanticComplete {
			t.Fatal("search literal became an evaluated field expression")
		}
	}
}

func TestSPL2PipelineForeignKeywordOptionOwnership(t *testing.T) {
	p := spl2PipelineFixture(t, "T3.fix1.foreign.stats")
	option := p.tree.Pipeline().Command(0).StatsCommand().StatsOption(0).UnknownOption()
	spl2PipelineText(t, p, option.UnknownOptionName(), "window")
	spl2PipelineText(t, p, option.Literal(), "3")
	if option.UnknownOptionName().GetStart().GetTokenType() != spl2.SPL2ParserWINDOW || p.syntaxComplete {
		t.Fatal("foreign keyword lost its typed, unproved ownership")
	}
	p = spl2PipelineFixture(t, "T3.fix1.foreign.eventstats")
	option = p.tree.Pipeline().Command(0).EventstatsCommand().UnknownOption(0)
	spl2PipelineText(t, p, option.UnknownOptionName(), "delim")
	spl2PipelineText(t, p, option.Literal().StringLiteral(), `";"`)
	if p.syntaxComplete {
		t.Fatal("foreign option acquired complete coverage")
	}
}

func TestSPL2PipelineUnprovedLiteralOwnership(t *testing.T) {
	for _, c := range []struct{ id, value string }{
		{"T3.fix1.search.fieldminus", "-bytes"}, {"T3.fix1.search.fieldplus", "+bytes"},
		{"T3.fix1.search.arithmetic", "-1+rate"}, {"T3.fix1.search.incomplete", "-"},
		{"T3.fix1.search.expression", "-"},
	} {
		p := spl2PipelineFixture(t, c.id)
		atoms := p.tree.Pipeline().Start_().SearchCommand().SearchExpression().SearchXor().SearchAnd(0).AllSearchOr()
		literal := atoms[1].SearchNot(0).SearchAtom().SearchValue(0).SearchUnprovedLiteral()
		spl2PipelineText(t, p, literal, c.value)
		if p.syntaxComplete || p.semanticComplete || len(p.diagnostics) == 0 {
			t.Fatal("unproved literal lost its limitation", c.id)
		}
		if len(spl2Nodes(p.syntax, "expression")) != 0 || len(spl2Nodes(p.syntax, "access")) != 0 {
			t.Fatal("unproved search literal interpreted as field/arithmetic", c.id)
		}
	}
}

func TestSPL2SelectedMultilinePipelineSyntax(t *testing.T) {
	for _, text := range []string{
		"FROM events\n| eval total=(value\n  + delta)",
		"FROM events\n| eval choice=coalesce(\n  primary,\n  fallback\n)",
		"FROM events\n| stats\n  sum(value) AS total",
		"FROM events\n| stats count() AS total BY\n  region,\n  kind",
	} {
		t.Run(text, func(t *testing.T) {
			p := spl2RequireNoDiagnostics(t, text)
			if len(spl2Nodes(p.syntax, "NL")) == 0 {
				t.Fatal("multiline source lost newline ownership")
			}
		})
	}
}

func TestSPL2SelectedCommandListContinuationNewlines(t *testing.T) {
	for _, tc := range []struct {
		name  string
		query string
		kind  string
		want  int
	}{
		{"aggregate list", "FROM synthetic_dataset | stats count() AS synthetic_count,\nsum(synthetic_value) AS synthetic_sum", "aggregate", 2},
		{"group boundary", "FROM synthetic_dataset | stats count() AS synthetic_count\nBY\nspan(synthetic_time, 5m)", "selectedAggregateGroup", 1},
		{"assignment list", "FROM synthetic_dataset | eval synthetic_left=1,\nsynthetic_right=2", "assignment", 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			parsed := spl2RequireNoDiagnostics(t, tc.query)
			if got := len(spl2Nodes(parsed.syntax, tc.kind)); got != tc.want {
				t.Fatalf("%s nodes = %d want %d: %s", tc.kind, got, tc.want, parsed.syntax.shape())
			}
		})
	}

	for _, query := range []string{
		"FROM synthetic_dataset | stats count() AS synthetic_count,\n| fields synthetic_count",
		"FROM synthetic_dataset | eval synthetic_left=1,\n| fields synthetic_left",
	} {
		t.Run(query, func(t *testing.T) {
			parsed := parseSPL2Document(query)
			if parsed.syntaxComplete || len(parsed.diagnostics) == 0 {
				t.Fatalf("missing list element was accepted: %+v", parsed.diagnostics)
			}
		})
	}
}

func TestSPL2SelectedExpressionGroupingAndSpan(t *testing.T) {
	text := "FROM events | stats sum(value) AS total BY bytes+delta, lower(region), payload.kind, span(_time, 5m)"
	p := spl2RequireNoDiagnostics(t, text)
	for kind, want := range map[string]int{"aggregate": 1, "selectedGroupTerm": 4, "selectedSpanGroup": 1, "timeSpan": 1} {
		if got := len(spl2Nodes(p.syntax, kind)); got != want {
			t.Errorf("%s count %d want %d: %s", kind, got, want, p.syntax.shape())
		}
	}
	span := spl2Nodes(p.syntax, "selectedSpanGroup")[0]
	if got := text[span.Location.Start.Offset:span.Location.End.Offset]; got != "span(_time, 5m)" {
		t.Fatalf("span source %q", got)
	}

	arithmetic := spl2RequireNoDiagnostics(t, "FROM events | stats count() BY bytes+delta")
	if got := len(spl2Nodes(arithmetic.syntax, "expression")); got != 1 {
		t.Fatalf("arithmetic group expressions %d: %s", got, arithmetic.syntax.shape())
	}
}

func TestSPL2SelectedMultilineExpressionGroupingAndSpan(t *testing.T) {
	for _, tc := range []struct {
		name       string
		query      string
		terms      int
		spanGroups int
	}{
		{"dotted", "FROM synthetic_dataset | stats count() BY\nsynthetic_parent.synthetic_field", 1, 0},
		{"span", "FROM synthetic_dataset | stats count() BY\nspan(synthetic_time, 5m)", 1, 1},
		{"expression list", "FROM synthetic_dataset | stats count() BY\nlower(synthetic_field),\nsynthetic_other", 2, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			parsed := spl2RequireNoDiagnostics(t, tc.query)
			if got := len(spl2Nodes(parsed.syntax, "selectedAggregateGroup")); got != 1 {
				t.Fatalf("selected groups = %d want 1: %s", got, parsed.syntax.shape())
			}
			if got := len(spl2Nodes(parsed.syntax, "selectedGroupTerm")); got != tc.terms {
				t.Fatalf("selected terms = %d want %d: %s", got, tc.terms, parsed.syntax.shape())
			}
			if got := len(spl2Nodes(parsed.syntax, "selectedSpanGroup")); got != tc.spanGroups {
				t.Fatalf("selected span groups = %d want %d: %s", got, tc.spanGroups, parsed.syntax.shape())
			}
			selected := 0
			for _, token := range parsed.tokens.GetAllTokens() {
				if token.GetTokenType() == spl2.SPL2LexerSELECTED_BY {
					selected++
				}
			}
			if selected != 1 {
				t.Fatalf("selected BY tokens = %d want 1", selected)
			}
		})
	}

	t.Run("missing selected term remains invalid", func(t *testing.T) {
		query := "FROM synthetic_dataset | stats count() BY\nlower(synthetic_field),\n| fields synthetic_field"
		parsed := parseSPL2Document(query)
		if parsed.syntaxComplete || len(parsed.diagnostics) == 0 {
			t.Fatalf("missing selected group term was accepted: %+v", parsed.diagnostics)
		}
	})
}

func TestSPL2SelectedExpressionGroupingUsesExistingExpressions(t *testing.T) {
	for _, tc := range []struct {
		text string
		held bool
	}{
		{`FROM events | stats count() BY if(bytes>0, region, "other")`, false},
		{"FROM events | stats count() BY (bytes+delta)", false},
		{"FROM events | stats count() BY -bytes", false},
		{"FROM events | stats count() BY ready AND active", false},
		{"FROM events | stats count() BY ready and active", true},
		{"FROM events | stats count() BY ready or active", true},
		{"FROM events | stats count() BY ready xor active", true},
		{"FROM events | stats count() BY NOT ready", false},
		{"FROM events | stats count() BY not ready", true},
		{"FROM events | stats count() BY bytes BETWEEN 1 AND 10", false},
		{`FROM events | stats count() BY region IN ("us", "eu")`, false},
		{`FROM events | stats count() BY name LIKE "a%"`, false},
		{"FROM events | stats count() BY value IS NOT NULL", false},
		{"FROM events | stats count() BY true", false},
		{"FROM events | stats count() BY null", false},
		{`FROM events | stats count() BY "other"`, false},
		{"FROM events | stats count() BY 42", false},
	} {
		t.Run(tc.text, func(t *testing.T) {
			p := parseSPL2Document(tc.text)
			for _, diagnostic := range p.diagnostics {
				if diagnostic.Severity == "error" || diagnostic.Code == CodeSyntaxError {
					t.Fatalf("grouping expression was not admitted: %+v", p.diagnostics)
				}
			}
			if tc.held == p.syntaxComplete {
				t.Fatalf("contextual casing hold = %t, syntax complete = %t: %+v", tc.held, p.syntaxComplete, p.diagnostics)
			}
			if len(spl2Nodes(p.syntax, "expression")) == 0 {
				t.Fatalf("grouping bypassed expression ownership: %s", p.syntax.shape())
			}
		})
	}
}

func TestSPL2SelectedExpressionGroupingClassifiesOnlyItsIntroducingBY(t *testing.T) {
	for _, tc := range []struct {
		name       string
		query      string
		introducer int
	}{
		{"single stats", "FROM events | stats count() BY BY+1", 1},
		{"independent stats", "FROM events | stats count() BY bytes+delta | stats count() BY BY+1", 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := spl2RequireNoDiagnostics(t, tc.query)
			selected, atoms := 0, 0
			for _, token := range p.tokens.GetAllTokens() {
				switch token.GetTokenType() {
				case spl2.SPL2LexerSELECTED_BY:
					selected++
				case spl2.SPL2LexerBY:
					atoms++
				}
			}
			if selected != tc.introducer || atoms != 1 {
				t.Fatalf("selected BY = %d want %d; expression atoms = %d want 1", selected, tc.introducer, atoms)
			}
		})
	}
}

func TestSPL2SelectedStatsDistinguishesExpressionAndCommandBrackets(t *testing.T) {
	for _, query := range []string{
		"FROM events | stats count() BY [stats, BY+1]",
		"FROM events | eval x=[stats, BY+1]",
	} {
		t.Run(query, func(t *testing.T) {
			p := spl2RequireNoDiagnostics(t, query)
			if got := len(spl2Nodes(p.syntax, "array")); got != 1 {
				t.Fatalf("expression arrays = %d want 1: %s", got, p.syntax.shape())
			}
		})
	}

	t.Run("inherited subpipe", func(t *testing.T) {
		query := "FROM events | appendpipe [stats count() BY bytes+delta]"
		p := spl2RequireNoDiagnostics(t, query)
		if got := len(spl2Nodes(p.syntax, "inheritedSubpipe")); got != 1 {
			t.Fatalf("inherited subpipes = %d want 1: %s", got, p.syntax.shape())
		}
		if got := len(spl2Nodes(p.syntax, "selectedAggregateGroup")); got != 1 {
			t.Fatalf("selected groups = %d want 1: %s", got, p.syntax.shape())
		}
	})

	t.Run("nested expression and subpipe", func(t *testing.T) {
		query := "FROM events | appendpipe [eval x=[stats, BY+1] | appendpipe [stats count() BY bytes+delta]]"
		p := spl2RequireNoDiagnostics(t, query)
		if got := len(spl2Nodes(p.syntax, "inheritedSubpipe")); got != 2 {
			t.Fatalf("inherited subpipes = %d want 2: %s", got, p.syntax.shape())
		}
		if got := len(spl2Nodes(p.syntax, "array")); got != 1 {
			t.Fatalf("expression arrays = %d want 1: %s", got, p.syntax.shape())
		}
		if got := len(spl2Nodes(p.syntax, "selectedAggregateGroup")); got != 1 {
			t.Fatalf("selected groups = %d want 1: %s", got, p.syntax.shape())
		}
	})
}

func TestSPL2SelectedMultilineAssignmentAnalysis(t *testing.T) {
	r := spl2AnalyzeTest(t, "FROM [{value:1, delta:2}] | eval total=(value\n + delta) | fields total")
	if r.Status != Valid || !r.Coverage.SyntaxComplete || !r.Coverage.SemanticComplete {
		t.Fatalf("multiline assignment remained unmodeled: %+v", r)
	}
	spl2Ref(t, r, "value", "read")
	spl2Ref(t, r, "delta", "read")
	spl2Ref(t, r, "total", "create")
	spl2Ref(t, r, "total", "read")
}

func TestSPL2SelectedGroupingAnalysisIsComplete(t *testing.T) {
	for _, tc := range []struct {
		name        string
		query       string
		groupOutput string
		inputs      []string
	}{
		{"expression", "FROM synthetic_events | stats count() AS synthetic_count BY synthetic_bytes+synthetic_delta", "synthetic_bytes+synthetic_delta", []string{"synthetic_bytes", "synthetic_delta"}},
		{"span", "FROM synthetic_events | stats count() AS synthetic_count BY span(synthetic_time, 5m)", "synthetic_time", []string{"synthetic_time"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := spl2AnalyzeTest(t, tc.query)
			if r.Status != Valid || !r.Coverage.SyntaxComplete || !r.Coverage.SemanticComplete || len(r.Diagnostics) != 0 {
				t.Fatalf("selected grouping did not receive complete semantic credit: %+v", r)
			}
			if r.Requirements.QueryStatus != Valid || !r.Requirements.Coverage.Complete || len(r.Requirements.Gaps) != 0 {
				t.Fatalf("selected grouping requirements are incomplete: %+v", r.Requirements)
			}
			const message = "Expression and span grouping effects are unmodeled"
			for _, diagnostic := range r.Diagnostics {
				if diagnostic.Message == message {
					t.Fatalf("temporary grouping diagnostic survived recovery: %+v", diagnostic)
				}
			}
			for _, gap := range r.Requirements.Gaps {
				if gap.Message == message {
					t.Fatalf("temporary grouping requirement gap survived recovery: %+v", gap)
				}
			}
			groupIDs := []string{}
			for _, name := range tc.inputs {
				ref := spl2Ref(t, r, name, "group")
				if ref.Binding != "source" || ref.Resolution != "exact" {
					t.Fatalf("group input %q lost exact source ownership: %+v", name, ref)
				}
				groupIDs = append(groupIDs, ref.ID)
				item := requirementItem(r.Requirements, "field", name, "group")
				if item == nil || item.Necessity != "required" || item.Resolution != "exact" || len(item.Occurrences) != 1 || item.Occurrences[0].ReferenceID != ref.ID {
					t.Fatalf("group input %q requirement is not exact: %+v", name, r.Requirements.Items)
				}
			}
			var output *FieldBinding
			for i := range r.Lineage[len(r.Lineage)-1].After.Fields {
				field := &r.Lineage[len(r.Lineage)-1].After.Fields[i]
				if field.Name == tc.groupOutput {
					output = field
				}
			}
			if output == nil || !reflect.DeepEqual(output.OriginReferenceIDs, groupIDs) || output.Conditional {
				t.Fatalf("modeled group output %q = %+v, want origins %v", tc.groupOutput, output, groupIDs)
			}
			alias := spl2Ref(t, r, "synthetic_count", "output")
			if alias.StageID != r.Stages[len(r.Stages)-1].ID || alias.Binding != "not_applicable" {
				t.Fatalf("aggregate alias lost output ownership: %+v", alias)
			}
			for _, ref := range r.References {
				if ref.NormalizedName == "span" || ref.NormalizedName == "5m" {
					t.Fatalf("span operator or duration became a field: %+v", ref)
				}
			}
		})
	}
}

func TestSPL2SelectedStructuralFieldSelectorReportsExactUnavailability(t *testing.T) {
	r := spl2AnalyzeTest(t, "FROM [{payload:1, other:2}] | fields payload.user.name")
	if r.Status != Invalid || !r.Coverage.SyntaxComplete || !r.Coverage.SemanticComplete || !spl2HasCode(r, CodeUnavailableField) {
		t.Fatalf("structural selector did not retain exact unavailable evidence: %+v", r)
	}
	if r.Requirements.QueryStatus != Invalid || !r.Requirements.Coverage.Complete || len(r.Requirements.Items) != 0 || len(r.Requirements.Gaps) != 0 {
		t.Fatalf("structural selector requirements disagree with exact unavailability: %+v", r.Requirements)
	}
	structural := spl2Ref(t, r, "payload.user.name", "read")
	if structural.Binding != "unavailable" || structural.OriginalName != "payload.user.name" {
		t.Fatalf("structural selector lost its exact identity: %+v", structural)
	}
	for _, ref := range r.References {
		if ref.NormalizedName == "payload" && (ref.Role == "read" || ref.Role == "output") {
			t.Fatalf("structural selector collapsed to its root: %+v", ref)
		}
	}
}

func TestSPL2SelectedConditionalBranchUnionAndJoin(t *testing.T) {
	for _, text := range []string{
		"FROM events | if (ready=true) [eval state=\"ready\"] elseif (failed=true) [eval state=\"failed\"] else [eval state=\"pending\"]",
		"FROM events | if (ready=true) [eval state=\"ready\"] elseif (failed=true) [eval state=\"failed\"]",
	} {
		p := spl2RequireNoDiagnostics(t, text)
		if got := len(spl2Nodes(p.syntax, "ifCommand")); got != 1 {
			t.Fatalf("if commands %d", got)
		}
	}

	branchText := "FROM events | branch (status=\"ok\") [where active=true], (status!=\"ok\") [eval failed=true]"
	branch := spl2RequireNoDiagnostics(t, branchText)
	if got := len(spl2Nodes(branch.syntax, "branchArm")); got != 2 {
		t.Fatalf("branch arms %d: %s", got, branch.syntax.shape())
	}

	unionText := "union $target_1, catalog.events, [FROM backup | where active=true]"
	union := spl2RequireNoDiagnostics(t, unionText)
	if len(spl2Nodes(union.syntax, "datasetParameter")) != 1 || len(spl2Nodes(union.syntax, "datasetPath")) != 1 || len(spl2Nodes(union.syntax, "independentSearch")) != 1 {
		t.Fatalf("union operands lost typed ownership: %s", union.syntax.shape())
	}

	joinText := "FROM events | join type=left left=L right=R where L.id=R.id AND L.realm=R.realm [FROM identities | where enabled=true]"
	join := spl2RequireNoDiagnostics(t, joinText)
	if len(spl2Nodes(join.syntax, "joinCommand")) != 1 || len(spl2Nodes(join.syntax, "sqlJoinEquality")) != 2 || len(spl2Nodes(join.syntax, "independentSearch")) != 1 {
		t.Fatalf("join ownership: %s", join.syntax.shape())
	}
}

func TestSPL2SelectedFinalBranchInBracketedPipelines(t *testing.T) {
	for _, tc := range []struct {
		name string
		text string
		kind string
	}{
		{
			name: "inherited",
			text: `FROM events | appendpipe [branch (status="ok") [where active=true], (status!="ok") [eval failed=true]]`,
			kind: "inheritedSubpipe",
		},
		{
			name: "independent",
			text: `FROM seed | append [FROM events | branch (status="ok") [where active=true], (status!="ok") [eval failed=true]]`,
			kind: "independentSearch",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			parsed := spl2RequireNoDiagnostics(t, tc.text)
			if got := len(spl2Nodes(parsed.syntax, "branchArm")); got != 2 {
				t.Fatalf("branch arms = %d, want 2: %s", got, parsed.syntax.shape())
			}
			if got := len(spl2Nodes(parsed.syntax, tc.kind)); got == 0 {
				t.Fatalf("missing %s ownership: %s", tc.kind, parsed.syntax.shape())
			}
		})
	}
}

func TestSPL2SelectedBranchUnionAndJoinBoundaries(t *testing.T) {
	for _, text := range []string{
		"FROM events | branch [where active=true], [eval failed=true]",
		"union $left.$right, events",
		"FROM events | join left=L right=R [FROM identities]",
		"FROM events | join left=L right=R where L.id>R.id [FROM identities]",
		"FROM events | join left=L right=R where L.id=R.id [$target_1]",
		"FROM events | join type=left left=L right=R where L.id=R.id [FROM identities] [FROM extra]",
	} {
		t.Run(text, func(t *testing.T) {
			spl2RequireLocatedError(t, text)
		})
	}
}
