package analysis

import (
	"encoding/json"
	"os"
	"path/filepath"
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
