package analysis

import (
	"github.com/antlr4-go/antlr/v4"
	"github.com/delgado-jacob/spl-toolkit/parser"
	"github.com/delgado-jacob/spl-toolkit/pkg/mapper"
	"strings"
	"sync"
	"testing"
)

// Grammar acceptance covers structured arguments; modeled forms can be complete.
func TestParseStructuredStages(t *testing.T) {
	for index, q := range []string{
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
			want := Valid
			if index == 2 || index == 5 || index == 6 {
				want = Incomplete
			}
			if r.Status != want || r.Coverage.SemanticComplete != (want == Valid) {
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
	for _, tc := range []struct {
		name, query string
		want        map[string]int
	}{
		{
			name:  "existing typed expressions",
			query: `search host=web | eval x=if(user=host,1,2), y=3 | lookup people uid AS user OUTPUT name AS display | stats sum(x) AS total BY user`,
			want:  map[string]int{"assignment": 2, "function": 2, "comparison": 1, "alias": 3, "output": 1, "group": 1},
		},
		{
			name: "milestone 10 command operands",
			query: `| tstats summariesonly=true sum(bytes) AS total FROM datamodel=Web.Events WHERE user="café" BY 'hôte', _time span=5m
| fillnull value="unknown" user,'display name'
| rex field=_raw max_match=2 offset_field=offsets "(?<name>.+)"
| spath input=_raw path="event.id" output=event_id
| bin span=5m _time AS bucket_time
| regex user!="^svc_"
| mvexpand values limit=3
| join type=left user,'tenant id' [ search child=* ]
| append maxout=100 [ search child=* ]
| appendpipe run_in_preview=true [ stats count ]`,
			want: map[string]int{
				"tstats": 1, "tstats-option": 1, "tstats-group": 1,
				"fillnull": 1, "rex": 1, "spath": 1, "bin": 1, "regex": 1,
				"mvexpand": 1, "join": 1, "branch": 2,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := parseDocument(tc.query)
			if len(p.diagnostics) != 0 {
				t.Fatal(p.diagnostics)
			}
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
				case *parser.AnalysisTstatsContext:
					counts["tstats"]++
				case *parser.AnalysisTstatsOptionContext:
					counts["tstats-option"]++
				case *parser.AnalysisTstatsGroupContext:
					counts["tstats-group"]++
				case *parser.AnalysisFillnullContext:
					counts["fillnull"]++
				case *parser.AnalysisRexContext:
					counts["rex"]++
				case *parser.AnalysisSpathContext:
					counts["spath"]++
				case *parser.AnalysisBinContext:
					counts["bin"]++
				case *parser.AnalysisRegexContext:
					counts["regex"]++
				case *parser.AnalysisMvexpandContext:
					counts["mvexpand"]++
				case *parser.AnalysisJoinContext:
					counts["join"]++
				case *parser.AnalysisBranchContext:
					counts["branch"]++
				}
				for i := 0; i < n.GetChildCount(); i++ {
					walk(n.GetChild(i))
				}
			}
			walk(p.tree)
			for kind, want := range tc.want {
				if counts[kind] != want {
					t.Fatalf("%s count = %d want %d", kind, counts[kind], want)
				}
			}
		})
	}
}

func TestParseMilestone10CommandContexts(t *testing.T) {
	for _, tc := range []struct {
		name, query string
		assert      func(*testing.T, parser.IAnalysisStageContext)
	}{
		{
			name:  "tstats comma-present aggregates and groups with unicode quoted fields whitespace and span",
			query: "| tstats summariesonly=true count AS total,\r\n\t sum('octets') AS 'sómme' FROM datamodel=Network_Traffic.All_Traffic WHERE All_Traffic.action=\"allowed\" BY 'hôte',\t_time span=5m",
			assert: func(t *testing.T, stage parser.IAnalysisStageContext) {
				ctx, ok := stage.(*parser.AnalysisTstatsStageContext)
				if !ok || ctx.AnalysisTstats() == nil {
					t.Fatalf("stage = %T", stage)
				}
				body := ctx.AnalysisTstats()
				if len(body.AllAnalysisTstatsOption()) != 1 || len(body.AllAnalysisAggregate()) != 2 || body.AnalysisTstatsFrom() == nil || body.AnalysisTstatsWhere() == nil || body.AnalysisTstatsGroup() == nil {
					t.Fatalf("incomplete tstats context: %s", body.GetText())
				}
				if len(body.AnalysisTstatsGroup().AllAnalysisTstatsGroupItem()) != 2 || body.AnalysisTstatsGroup().AnalysisTstatsGroupItem(1).AnalysisTstatsSpanOption() == nil || body.AnalysisTstatsGroup().AnalysisTstatsGroupItem(1).AnalysisTstatsSpanOption().GetText() != "span=5m" {
					t.Fatalf("group context: %s", body.AnalysisTstatsGroup().GetText())
				}
			},
		},
		{
			name:  "tstats comma-omitted aggregates and groups",
			query: `| tstats count AS total sum(bytes) AS octets BY host _time span=5m`,
			assert: func(t *testing.T, stage parser.IAnalysisStageContext) {
				ctx, ok := stage.(*parser.AnalysisTstatsStageContext)
				if !ok || ctx.AnalysisTstats() == nil {
					t.Fatalf("stage = %T", stage)
				}
				body := ctx.AnalysisTstats()
				if len(body.AllAnalysisAggregate()) != 2 || body.AnalysisTstatsGroup() == nil || len(body.AnalysisTstatsGroup().AllAnalysisTstatsGroupItem()) != 2 || body.AnalysisTstatsGroup().AnalysisTstatsGroupItem(1).AnalysisTstatsSpanOption() == nil || body.AnalysisTstatsGroup().AnalysisTstatsGroupItem(1).AnalysisTstatsSpanOption().GetText() != "span=5m" {
					t.Fatalf("comma-omitted tstats context: %s", body.GetText())
				}
			},
		},
		{
			name:  "fillnull comma-present fields",
			query: `| fillnull value="unknown" café, 'display name'`,
			assert: func(t *testing.T, stage parser.IAnalysisStageContext) {
				ctx, ok := stage.(*parser.AnalysisFillnullStageContext)
				if !ok || ctx.AnalysisFillnull() == nil || ctx.AnalysisFillnull().AnalysisFillnullValueOption() == nil || len(ctx.AnalysisFillnull().AllAnalysisIdentifier()) != 2 {
					t.Fatalf("stage = %T %s", stage, stage.GetText())
				}
			},
		},
		{
			name:  "fillnull comma-omitted fields",
			query: `| fillnull value="unknown" café 'display name'`,
			assert: func(t *testing.T, stage parser.IAnalysisStageContext) {
				ctx, ok := stage.(*parser.AnalysisFillnullStageContext)
				if !ok || ctx.AnalysisFillnull() == nil || ctx.AnalysisFillnull().AnalysisFillnullValueOption() == nil || len(ctx.AnalysisFillnull().AllAnalysisIdentifier()) != 2 {
					t.Fatalf("stage = %T %s", stage, stage.GetText())
				}
			},
		},
		{
			name:  "rex typed slots",
			query: `| rex field='raw field' max_match=2 offset_field='offset list' "(?<café>.+)"`,
			assert: func(t *testing.T, stage parser.IAnalysisStageContext) {
				ctx, ok := stage.(*parser.AnalysisRexStageContext)
				if !ok || len(ctx.AnalysisRex().AllAnalysisRexFieldOption()) != 1 || len(ctx.AnalysisRex().AllAnalysisRexMaxMatchOption()) != 1 || len(ctx.AnalysisRex().AllAnalysisRexOffsetFieldOption()) != 1 || ctx.AnalysisRex().STRING() == nil {
					t.Fatalf("stage = %T %s", stage, stage.GetText())
				}
			},
		},
		{
			name:  "spath typed slots",
			query: `| spath input='raw payload' path="event.id" output='event id'`,
			assert: func(t *testing.T, stage parser.IAnalysisStageContext) {
				ctx, ok := stage.(*parser.AnalysisSpathStageContext)
				if !ok || len(ctx.AnalysisSpath().AllAnalysisSpathInputOption()) != 1 || len(ctx.AnalysisSpath().AllAnalysisSpathPathOption()) != 1 || len(ctx.AnalysisSpath().AllAnalysisSpathOutputOption()) != 1 {
					t.Fatalf("stage = %T %s", stage, stage.GetText())
				}
			},
		},
		{
			name:  "bin literal option and alias",
			query: `| bin span=5m 'event time' AS 'bucket time'`,
			assert: func(t *testing.T, stage parser.IAnalysisStageContext) {
				ctx, ok := stage.(*parser.AnalysisBinStageContext)
				if !ok || len(ctx.AnalysisBin().AllAnalysisBinOption()) != 1 || ctx.AnalysisBin().AnalysisAlias() == nil {
					t.Fatalf("stage = %T %s", stage, stage.GetText())
				}
			},
		},
		{
			name:  "regex exact field",
			query: `| regex 'user name'!="^svc_"`,
			assert: func(t *testing.T, stage parser.IAnalysisStageContext) {
				ctx, ok := stage.(*parser.AnalysisRegexStageContext)
				if !ok || ctx.AnalysisRegex().AnalysisIdentifier() == nil || ctx.AnalysisRegex().NE() == nil || ctx.AnalysisRegex().STRING() == nil {
					t.Fatalf("stage = %T %s", stage, stage.GetText())
				}
			},
		},
		{
			name:  "mvexpand exact field and option",
			query: `| mvexpand 'tag list' limit=3`,
			assert: func(t *testing.T, stage parser.IAnalysisStageContext) {
				ctx, ok := stage.(*parser.AnalysisMvexpandStageContext)
				if !ok || len(ctx.AnalysisMvexpand().AllAnalysisMvexpandOption()) != 1 {
					t.Fatalf("stage = %T %s", stage, stage.GetText())
				}
			},
		},
		{
			name:  "join comma-present exact keys and nested branch",
			query: `| join type=left user,'tenant id' [ search child=* | append [ search nested=* ] ]`,
			assert: func(t *testing.T, stage parser.IAnalysisStageContext) {
				ctx, ok := stage.(*parser.AnalysisJoinStageContext)
				if !ok || len(ctx.AnalysisJoin().AllAnalysisJoinOption()) != 1 || len(ctx.AnalysisJoin().AllAnalysisIdentifier()) != 2 || ctx.AnalysisJoin().AnalysisSubquery() == nil {
					t.Fatalf("stage = %T %s", stage, stage.GetText())
				}
				childStages := ctx.AnalysisJoin().AnalysisSubquery().AnalysisPipeline().AllAnalysisStage()
				if len(childStages) != 1 {
					t.Fatalf("nested child stages = %d", len(childStages))
				}
				if _, ok := childStages[0].(*parser.AnalysisBranchStageContext); !ok {
					t.Fatalf("nested stage = %T", childStages[0])
				}
			},
		},
		{
			name:  "join comma-omitted exact keys",
			query: `| join type=left user 'tenant id' [ search child=* ]`,
			assert: func(t *testing.T, stage parser.IAnalysisStageContext) {
				ctx, ok := stage.(*parser.AnalysisJoinStageContext)
				if !ok || len(ctx.AnalysisJoin().AllAnalysisJoinOption()) != 1 || len(ctx.AnalysisJoin().AllAnalysisIdentifier()) != 2 || ctx.AnalysisJoin().AnalysisSubquery() == nil {
					t.Fatalf("stage = %T %s", stage, stage.GetText())
				}
			},
		},
		{
			name:  "append option and subquery",
			query: `| append maxout=100 [ search child=* ]`,
			assert: func(t *testing.T, stage parser.IAnalysisStageContext) {
				ctx, ok := stage.(*parser.AnalysisBranchStageContext)
				if !ok || len(ctx.AnalysisBranch().AllAnalysisBranchOption()) != 1 || ctx.AnalysisBranch().AnalysisSubquery() == nil {
					t.Fatalf("stage = %T %s", stage, stage.GetText())
				}
			},
		},
		{
			name:  "appendpipe option and subquery",
			query: `| appendpipe run_in_preview=true [ stats count ]`,
			assert: func(t *testing.T, stage parser.IAnalysisStageContext) {
				ctx, ok := stage.(*parser.AnalysisBranchStageContext)
				if !ok || len(ctx.AnalysisBranch().AllAnalysisBranchOption()) != 1 || ctx.AnalysisBranch().AnalysisSubquery() == nil {
					t.Fatalf("stage = %T %s", stage, stage.GetText())
				}
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := parseDocument(tc.query)
			if len(p.diagnostics) != 0 {
				t.Fatal(p.diagnostics)
			}
			stages := p.tree.AnalysisPipeline().AllAnalysisStage()
			if len(stages) != 1 {
				t.Fatalf("top-level stages = %d", len(stages))
			}
			tc.assert(t, stages[0])
		})
	}
}

func TestParseMilestone10HeldForms(t *testing.T) {
	for _, tc := range []struct {
		name, query string
		assert      func(*testing.T, parser.IAnalysisStageContext)
	}{
		{
			name:  "tstats exact macro context",
			query: `| tstats prestats=true ` + "`summariesonly`" + ` PREFIX(user) BY user`,
			assert: func(t *testing.T, stage parser.IAnalysisStageContext) {
				ctx, ok := stage.(*parser.AnalysisTstatsStageContext)
				if !ok || ctx.AnalysisTstats() == nil || ctx.AnalysisTstats().AnalysisMacro() == nil || ctx.AnalysisTstats().AnalysisMacro().GetText() != "`summariesonly`" {
					t.Fatalf("tstats macro context = %T %q", stage, stage.GetText())
				}
			},
		},
		{
			name:  "fillnull all-fields form",
			query: `| fillnull value="unknown"`,
			assert: func(t *testing.T, stage parser.IAnalysisStageContext) {
				ctx, ok := stage.(*parser.AnalysisFillnullStageContext)
				if !ok || ctx.AnalysisFillnull() == nil || ctx.AnalysisFillnull().AnalysisFillnullValueOption() == nil || len(ctx.AnalysisFillnull().AllAnalysisIdentifier()) != 0 {
					t.Fatalf("fillnull context = %T %q", stage, stage.GetText())
				}
			},
		},
		{
			name:  "rex sed mode and field contexts",
			query: `| rex mode=sed field=_raw "s/a/b/g"`,
			assert: func(t *testing.T, stage parser.IAnalysisStageContext) {
				ctx, ok := stage.(*parser.AnalysisRexStageContext)
				if !ok || ctx.AnalysisRex() == nil || len(ctx.AnalysisRex().AllAnalysisRexModeOption()) != 1 || len(ctx.AnalysisRex().AllAnalysisRexFieldOption()) != 1 || ctx.AnalysisRex().AnalysisRexModeOption(0).GetText() != "mode=sed" || ctx.AnalysisRex().AnalysisRexFieldOption(0).GetText() != "field=_raw" {
					t.Fatalf("rex contexts = %T %q", stage, stage.GetText())
				}
			},
		},
		{
			name:  "spath input present path and output absent",
			query: `| spath input=_raw`,
			assert: func(t *testing.T, stage parser.IAnalysisStageContext) {
				ctx, ok := stage.(*parser.AnalysisSpathStageContext)
				if !ok || ctx.AnalysisSpath() == nil || len(ctx.AnalysisSpath().AllAnalysisSpathInputOption()) != 1 || len(ctx.AnalysisSpath().AnalysisSpathInputOption(0).AllAnalysisIdentifier()) != 2 || ctx.AnalysisSpath().AnalysisSpathInputOption(0).AnalysisIdentifier(1).GetText() != "_raw" || len(ctx.AnalysisSpath().AllAnalysisSpathPathOption()) != 0 || len(ctx.AnalysisSpath().AllAnalysisSpathOutputOption()) != 0 {
					t.Fatalf("spath contexts = %T %q", stage, stage.GetText())
				}
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := parseDocument(tc.query)
			if len(p.diagnostics) != 0 {
				t.Fatalf("held form lost typed syntax: %+v", p.diagnostics)
			}
			stage := p.tree.AnalysisPipeline().AnalysisStage(0)
			if _, opaque := stage.(*parser.AnalysisOpaqueStageContext); opaque {
				t.Fatal("held form fell back to opaque stage")
			}
			tc.assert(t, stage)
		})
	}

	for _, q := range []string{
		`| tstats summariesonly= count`,
		`| rex field= "(?<x>.)"`,
		`| spath input=_raw output=`,
		`| bin span= _time`,
		`| mvexpand values limit=`,
		`| join type= [ search * ]`,
		`| append maxout= [ search * ]`,
	} {
		t.Run("malformed "+q, func(t *testing.T) {
			p := parseDocument(q)
			if len(p.diagnostics) == 0 {
				t.Fatal("malformed option accepted")
			}
			if _, opaque := p.tree.AnalysisPipeline().AnalysisStage(0).(*parser.AnalysisOpaqueStageContext); opaque {
				t.Fatal("malformed modeled command fell back to opaque stage")
			}
		})
	}

	for _, tc := range []struct {
		name, query string
	}{
		{"tstats unrelated option rejects adjacent unit-like suffix", `| tstats summariesonly=5garbage count | table safe`},
		{"append unrelated option rejects adjacent unit-like suffix", `| append maxout=5garbage [ search * ] | table safe`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := parseDocument(tc.query)
			if len(p.diagnostics) == 0 {
				t.Fatal("unrelated option accepted number plus arbitrary identifier")
			}
			outer := p.tree.AnalysisPipeline().AllAnalysisStage()
			if len(outer) != 2 {
				t.Fatalf("outer stages = %d", len(outer))
			}
			if _, ok := outer[1].(*parser.AnalysisFieldsStageContext); !ok || outer[1].GetText() != "tablesafe" {
				t.Fatalf("outer sibling stage = %T %q", outer[1], outer[1].GetText())
			}
		})
	}

	t.Run("bin invalid unit suffix preserves input and outer sibling", func(t *testing.T) {
		p := parseDocument(`| bin span=5garbage _time | table safe`)
		foundInvalidValue := false
		for _, diagnostic := range p.diagnostics {
			foundInvalidValue = foundInvalidValue || diagnostic.Message == "invalid option value"
		}
		if !foundInvalidValue {
			t.Fatalf("invalid bin unit suffix diagnostics = %+v", p.diagnostics)
		}
		outer := p.tree.AnalysisPipeline().AllAnalysisStage()
		if len(outer) != 2 {
			t.Fatalf("outer stages = %d", len(outer))
		}
		bin, ok := outer[0].(*parser.AnalysisBinStageContext)
		if !ok || bin.AnalysisBin() == nil || len(bin.AnalysisBin().AllAnalysisBinOption()) != 1 || bin.AnalysisBin().AnalysisBinOption(0).AnalysisUnitOptionValue() == nil || bin.AnalysisBin().AnalysisBinOption(0).AnalysisUnitOptionValue().GetText() != "5garbage" || bin.AnalysisBin().AnalysisIdentifier() == nil || bin.AnalysisBin().AnalysisIdentifier().GetText() != "_time" {
			t.Fatalf("bin context = %T %q", outer[0], outer[0].GetText())
		}
		if _, ok := outer[1].(*parser.AnalysisFieldsStageContext); !ok || outer[1].GetText() != "tablesafe" {
			t.Fatalf("outer sibling stage = %T %q", outer[1], outer[1].GetText())
		}
	})

	t.Run("tstats invalid span suffix preserves later groups and outer sibling", func(t *testing.T) {
		p := parseDocument(`| tstats count BY _time span=5garbage host region | table safe`)
		foundInvalidValue := false
		for _, diagnostic := range p.diagnostics {
			foundInvalidValue = foundInvalidValue || diagnostic.Message == "invalid option value"
		}
		if !foundInvalidValue {
			t.Fatalf("invalid tstats span suffix diagnostics = %+v", p.diagnostics)
		}
		outer := p.tree.AnalysisPipeline().AllAnalysisStage()
		if len(outer) != 2 {
			t.Fatalf("outer stages = %d", len(outer))
		}
		tstats, ok := outer[0].(*parser.AnalysisTstatsStageContext)
		if !ok || tstats.AnalysisTstats() == nil || tstats.AnalysisTstats().AnalysisTstatsGroup() == nil {
			t.Fatalf("tstats context = %T %q", outer[0], outer[0].GetText())
		}
		items := tstats.AnalysisTstats().AnalysisTstatsGroup().AllAnalysisTstatsGroupItem()
		if len(items) != 3 {
			t.Fatalf("group items = %d, text = %q", len(items), tstats.AnalysisTstats().AnalysisTstatsGroup().GetText())
		}
		for i, want := range []string{"_time", "host", "region"} {
			if items[i].AnalysisIdentifier() == nil || items[i].AnalysisIdentifier().GetText() != want {
				t.Fatalf("group item %d = %q, want %q", i, items[i].GetText(), want)
			}
		}
		span := items[0].AnalysisTstatsSpanOption()
		if span == nil || span.AnalysisUnitOptionValue() == nil || span.AnalysisUnitOptionValue().GetText() != "5garbage" {
			t.Fatalf("span context = %v", span)
		}
		if _, ok := outer[1].(*parser.AnalysisFieldsStageContext); !ok || outer[1].GetText() != "tablesafe" {
			t.Fatalf("outer sibling stage = %T %q", outer[1], outer[1].GetText())
		}
	})

	t.Run("tstats numeric-only span preserves following group and outer sibling", func(t *testing.T) {
		p := parseDocument(`| tstats count BY _time span=5 host | table safe`)
		if len(p.diagnostics) != 0 {
			t.Fatalf("numeric-only tstats span diagnostics = %+v", p.diagnostics)
		}
		outer := p.tree.AnalysisPipeline().AllAnalysisStage()
		if len(outer) != 2 {
			t.Fatalf("outer stages = %d", len(outer))
		}
		tstats, ok := outer[0].(*parser.AnalysisTstatsStageContext)
		if !ok || tstats.AnalysisTstats() == nil || tstats.AnalysisTstats().AnalysisTstatsGroup() == nil {
			t.Fatalf("tstats context = %T %q", outer[0], outer[0].GetText())
		}
		items := tstats.AnalysisTstats().AnalysisTstatsGroup().AllAnalysisTstatsGroupItem()
		if len(items) != 2 || items[0].AnalysisIdentifier().GetText() != "_time" || items[1].AnalysisIdentifier().GetText() != "host" || items[1].AnalysisTstatsSpanOption() != nil {
			t.Fatalf("group context = %q", tstats.AnalysisTstats().AnalysisTstatsGroup().GetText())
		}
		span := items[0].AnalysisTstatsSpanOption()
		if span == nil || span.AnalysisUnitOptionValue() == nil || span.AnalysisUnitOptionValue().GetText() != "5" || span.AnalysisUnitOptionValue().AnalysisUnitSuffix() != nil {
			t.Fatalf("span context = %v", span)
		}
		if _, ok := outer[1].(*parser.AnalysisFieldsStageContext); !ok || outer[1].GetText() != "tablesafe" {
			t.Fatalf("outer sibling stage = %T %q", outer[1], outer[1].GetText())
		}
	})

	t.Run("bin numeric-only span preserves input and outer sibling", func(t *testing.T) {
		p := parseDocument(`| bin span=5 _time | table safe`)
		if len(p.diagnostics) != 0 {
			t.Fatalf("numeric-only bin span diagnostics = %+v", p.diagnostics)
		}
		outer := p.tree.AnalysisPipeline().AllAnalysisStage()
		if len(outer) != 2 {
			t.Fatalf("outer stages = %d", len(outer))
		}
		bin, ok := outer[0].(*parser.AnalysisBinStageContext)
		if !ok || bin.AnalysisBin() == nil || len(bin.AnalysisBin().AllAnalysisBinOption()) != 1 || bin.AnalysisBin().AnalysisBinOption(0).AnalysisUnitOptionValue() == nil || bin.AnalysisBin().AnalysisBinOption(0).AnalysisUnitOptionValue().GetText() != "5" || bin.AnalysisBin().AnalysisBinOption(0).AnalysisUnitOptionValue().AnalysisUnitSuffix() != nil || bin.AnalysisBin().AnalysisIdentifier() == nil || bin.AnalysisBin().AnalysisIdentifier().GetText() != "_time" {
			t.Fatalf("bin context = %T %q", outer[0], outer[0].GetText())
		}
		if _, ok := outer[1].(*parser.AnalysisFieldsStageContext); !ok || outer[1].GetText() != "tablesafe" {
			t.Fatalf("outer sibling stage = %T %q", outer[1], outer[1].GetText())
		}
	})

	t.Run("append missing option value preserves child and outer sibling", func(t *testing.T) {
		p := parseDocument(`| append maxout= [ search * ] | table safe`)
		if len(p.diagnostics) == 0 {
			t.Fatal("missing branch option value accepted")
		}
		outer := p.tree.AnalysisPipeline().AllAnalysisStage()
		if len(outer) != 2 {
			t.Fatalf("outer stages = %d", len(outer))
		}
		branch, ok := outer[0].(*parser.AnalysisBranchStageContext)
		if !ok || branch.AnalysisBranch() == nil || len(branch.AnalysisBranch().AllAnalysisBranchOption()) != 1 || branch.AnalysisBranch().AnalysisBranchOption(0).AnalysisOptionValue() != nil || branch.AnalysisBranch().AnalysisSubquery() == nil {
			t.Fatalf("branch context = %T %q", outer[0], outer[0].GetText())
		}
		child := branch.AnalysisBranch().AnalysisSubquery().AnalysisPipeline().AnalysisInitialStage().AnalysisStage()
		if _, ok := child.(*parser.AnalysisSearchStageContext); !ok || child.GetText() != "search*" {
			t.Fatalf("child stage = %T %q", child, child.GetText())
		}
		if _, ok := outer[1].(*parser.AnalysisFieldsStageContext); !ok || outer[1].GetText() != "tablesafe" {
			t.Fatalf("outer sibling stage = %T %q", outer[1], outer[1].GetText())
		}
	})

	t.Run("spath missing output value preserves outer sibling", func(t *testing.T) {
		p := parseDocument(`| spath input=_raw output= | table safe`)
		if len(p.diagnostics) == 0 {
			t.Fatal("missing spath output accepted")
		}
		outer := p.tree.AnalysisPipeline().AllAnalysisStage()
		if len(outer) != 2 {
			t.Fatalf("outer stages = %d", len(outer))
		}
		spath, ok := outer[0].(*parser.AnalysisSpathStageContext)
		if !ok || spath.AnalysisSpath() == nil || len(spath.AnalysisSpath().AllAnalysisSpathInputOption()) != 1 || len(spath.AnalysisSpath().AllAnalysisSpathOutputOption()) != 1 || spath.AnalysisSpath().AnalysisSpathOutputOption(0).AnalysisIdentifier() != nil {
			t.Fatalf("spath context = %T %q", outer[0], outer[0].GetText())
		}
		if _, ok := outer[1].(*parser.AnalysisFieldsStageContext); !ok || outer[1].GetText() != "tablesafe" {
			t.Fatalf("outer sibling stage = %T %q", outer[1], outer[1].GetText())
		}
	})

	t.Run("nested spath missing output preserves closing bracket and outer sibling", func(t *testing.T) {
		p := parseDocument(`| append [ spath output= ] | table safe`)
		if len(p.diagnostics) == 0 {
			t.Fatal("missing nested spath output accepted")
		}
		outer := p.tree.AnalysisPipeline().AllAnalysisStage()
		if len(outer) != 2 {
			t.Fatalf("outer stages = %d", len(outer))
		}
		branch, ok := outer[0].(*parser.AnalysisBranchStageContext)
		if !ok || branch.AnalysisBranch() == nil || branch.AnalysisBranch().AnalysisSubquery() == nil {
			t.Fatalf("branch context = %T %q", outer[0], outer[0].GetText())
		}
		child := branch.AnalysisBranch().AnalysisSubquery().AnalysisPipeline().AnalysisInitialStage().AnalysisStage()
		spath, ok := child.(*parser.AnalysisSpathStageContext)
		if !ok || spath.AnalysisSpath() == nil || len(spath.AnalysisSpath().AllAnalysisSpathOutputOption()) != 1 || spath.AnalysisSpath().AnalysisSpathOutputOption(0).AnalysisIdentifier() != nil {
			t.Fatalf("child stage = %T %q", child, child.GetText())
		}
		if _, ok := outer[1].(*parser.AnalysisFieldsStageContext); !ok || outer[1].GetText() != "tablesafe" {
			t.Fatalf("outer sibling stage = %T %q", outer[1], outer[1].GetText())
		}
	})

	t.Run("join missing key preserves child and outer sibling", func(t *testing.T) {
		p := parseDocument(`| join type= [ search * ] | table safe`)
		foundMissingKey := false
		for _, diagnostic := range p.diagnostics {
			foundMissingKey = foundMissingKey || diagnostic.Message == "missing join key"
		}
		if !foundMissingKey {
			t.Fatalf("missing join key diagnostics = %+v", p.diagnostics)
		}
		outer := p.tree.AnalysisPipeline().AllAnalysisStage()
		if len(outer) != 2 {
			t.Fatalf("outer stages = %d", len(outer))
		}
		join, ok := outer[0].(*parser.AnalysisJoinStageContext)
		if !ok || join.AnalysisJoin() == nil || len(join.AnalysisJoin().AllAnalysisJoinOption()) != 1 || len(join.AnalysisJoin().AllAnalysisIdentifier()) != 0 || join.AnalysisJoin().AnalysisSubquery() == nil {
			t.Fatalf("join context = %T %q", outer[0], outer[0].GetText())
		}
		child := join.AnalysisJoin().AnalysisSubquery().AnalysisPipeline().AnalysisInitialStage().AnalysisStage()
		if _, ok := child.(*parser.AnalysisSearchStageContext); !ok || child.GetText() != "search*" {
			t.Fatalf("child stage = %T %q", child, child.GetText())
		}
		if _, ok := outer[1].(*parser.AnalysisFieldsStageContext); !ok || outer[1].GetText() != "tablesafe" {
			t.Fatalf("outer sibling stage = %T %q", outer[1], outer[1].GetText())
		}
	})

	t.Run("nested join missing key preserves both subqueries and outer sibling", func(t *testing.T) {
		p := parseDocument(`| append [ join type= [ search * ] ] | table safe`)
		foundMissingKey := false
		for _, diagnostic := range p.diagnostics {
			foundMissingKey = foundMissingKey || diagnostic.Message == "missing join key"
		}
		if !foundMissingKey {
			t.Fatalf("missing nested join key diagnostics = %+v", p.diagnostics)
		}
		outer := p.tree.AnalysisPipeline().AllAnalysisStage()
		if len(outer) != 2 {
			t.Fatalf("outer stages = %d", len(outer))
		}
		branch, ok := outer[0].(*parser.AnalysisBranchStageContext)
		if !ok || branch.AnalysisBranch() == nil || branch.AnalysisBranch().AnalysisSubquery() == nil {
			t.Fatalf("branch context = %T %q", outer[0], outer[0].GetText())
		}
		joinStage := branch.AnalysisBranch().AnalysisSubquery().AnalysisPipeline().AnalysisInitialStage().AnalysisStage()
		join, ok := joinStage.(*parser.AnalysisJoinStageContext)
		if !ok || join.AnalysisJoin() == nil || len(join.AnalysisJoin().AllAnalysisIdentifier()) != 0 || join.AnalysisJoin().AnalysisSubquery() == nil {
			t.Fatalf("join child = %T %q", joinStage, joinStage.GetText())
		}
		search := join.AnalysisJoin().AnalysisSubquery().AnalysisPipeline().AnalysisInitialStage().AnalysisStage()
		if _, ok := search.(*parser.AnalysisSearchStageContext); !ok || search.GetText() != "search*" {
			t.Fatalf("search child = %T %q", search, search.GetText())
		}
		if _, ok := outer[1].(*parser.AnalysisFieldsStageContext); !ok || outer[1].GetText() != "tablesafe" {
			t.Fatalf("outer sibling stage = %T %q", outer[1], outer[1].GetText())
		}
	})

	t.Run("damaged child remains isolated", func(t *testing.T) {
		p := parseDocument(`search root=* | append [ search child=* | mystery good $ | stats count BY child ] | table root`)
		if len(p.diagnostics) == 0 {
			t.Fatal("damaged child was accepted")
		}
		outer := p.tree.AnalysisPipeline().AllAnalysisStage()
		if len(outer) != 2 {
			for i, stage := range outer {
				t.Logf("outer stage %d: %T %q", i, stage, stage.GetText())
			}
			t.Fatalf("outer stages = %d", len(outer))
		}
		branch, ok := outer[0].(*parser.AnalysisBranchStageContext)
		if !ok {
			t.Fatalf("branch stage = %T", outer[0])
		}
		child := branch.AnalysisBranch().AnalysisSubquery().AnalysisPipeline().AllAnalysisStage()
		if len(child) != 2 {
			t.Fatalf("child stages = %d", len(child))
		}
		if _, ok := child[1].(*parser.AnalysisStatsStageContext); !ok {
			t.Fatalf("recovered child stage = %T", child[1])
		}
		if _, ok := outer[1].(*parser.AnalysisFieldsStageContext); !ok {
			t.Fatalf("outer sibling stage = %T", outer[1])
		}
	})
}

func TestParseMilestone10LegacyEntryPoint(t *testing.T) {
	// These exact trees were captured through Query before regenerating the
	// analysis-only rules. They bind mapper behavior to the pre-change parser.
	for _, tc := range []struct {
		query, tree string
	}{
		{
			query: `search src_ip=1 dst_ip=2`,
			tree:  `(query (initCommand search (operation (id src_ip) = (expression (value 1))) (operation (id dst_ip) = (expression (value 2)))) <EOF>)`,
		},
		{
			query: `search host=web | eval result=lower(user) | stats count BY result`,
			tree:  `(query (initCommand search (operation (id host) = (expression (value (id web))))) | (nextCommand (command eval) (operation (id result) = (expression (function lower) ( (expression (value (id user))) )))) | (nextCommand (command stats) (operation (expression (value (id (function count))))) (operation BY (id result))) <EOF>)`,
		},
		{
			query: `| tstats count from datamodel=Web by Web.status`,
			tree:  `(query (initCommand | tstats (operation (expression (value (id (function count))))) (operation (expression (value (id (command from))))) (operation (id (command datamodel)) = (expression (value (id Web)))) (operation by (id Web.status))) <EOF>)`,
		},
	} {
		t.Run(tc.query, func(t *testing.T) {
			lexer := parser.NewSPLLexer(antlr.NewInputStream(tc.query))
			tokens := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
			legacy := parser.NewSPLParser(tokens)
			tree := legacy.Query()
			if got := tree.ToStringTree(legacy.GetRuleNames(), legacy); got != tc.tree {
				t.Fatalf("legacy tree changed\ngot:  %s\nwant: %s", got, tc.tree)
			}
			mapped, err := mapper.New().MapQuery(tc.query)
			if err != nil || mapped != tc.query {
				t.Fatalf("legacy mapper = %q, %v", mapped, err)
			}
		})
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
		{"host.name", []int{parser.SPLLexerIDENTIFIER}},
		{"host.", []int{parser.SPLLexerIDENTIFIER}},
		{".metadata", []int{parser.SPLLexerIDENTIFIER}},
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

// An adjacent string concatenation operator must not become part of its left identifier.
func TestParseAdjacentConcatenation(t *testing.T) {
	for _, q := range []string{`| eval x=host."!"`, `| eval x=host . "!"`, `| eval x=host."!"."?"`, `| eval x=host. "!"`, `| eval x=host.("!")`, `| eval x=host.'other'`} {
		p := parseDocument(q)
		if len(p.diagnostics) != 0 {
			t.Fatalf("%s: %+v", q, p.diagnostics)
		}
		stage := p.tree.AnalysisPipeline().AnalysisStage(0).(*parser.AnalysisEvalStageContext)
		concat := stage.AnalysisAssignment(0).AnalysisExpression().AnalysisOr().AnalysisAnd(0).AnalysisNot(0).AnalysisComparison().AnalysisConcat(0)
		if len(concat.AllAnalysisAdd()) < 2 {
			t.Fatalf("%s lost concatenation", q)
		}
		left := concat.AnalysisAdd(0).AnalysisMultiply(0).AnalysisPower(0).AnalysisUnary().AnalysisAtom().AnalysisIdentifier()
		if left == nil || left.GetText() != "host" {
			t.Fatalf("%s left operand %v", q, left)
		}
	}
}

// A comparison's complete unquoted RHS owns adjacent punctuation; spaces start another search term.
func TestParseUnquotedSearchValueAdjacency(t *testing.T) {
	for _, tc := range []struct {
		q, want string
		terms   int
	}{
		{`search host=web-01`, `web-01`, 1},
		{`search host=web-01 extra`, `web-01`, 2},
		{`search host=web -01`, `web`, 2},
		{`search host=web/* comment */-01`, `web`, 2},
		{`search host=café-01 status=200`, `café-01`, 2},
		{`search source=/var/log/web-01.log status=200`, `/var/log/web-01.log`, 2},
		{`search src_ip=10.2.3.4 status=200`, `10.2.3.4`, 2},
		{`search host=web* extra`, `web*`, 2},
	} {
		p := parseDocument(tc.q)
		if len(p.diagnostics) != 0 {
			t.Fatalf("%s: %+v", tc.q, p.diagnostics)
		}
		stage := p.tree.AnalysisPipeline().AnalysisInitialStage().AnalysisStage().(*parser.AnalysisSearchStageContext)
		terms := stage.AnalysisSearch().AnalysisSearchAnd(0).AllAnalysisSearchUnary()
		if len(terms) != tc.terms {
			t.Fatalf("%s has %d terms, want %d", tc.q, len(terms), tc.terms)
		}
		rhs := terms[0].AnalysisSearchTerm().AnalysisSearchValue(0)
		if rhs.GetText() != tc.want {
			t.Fatalf("%s RHS %q want %q", tc.q, rhs.GetText(), tc.want)
		}
		loc := p.source.contextLocation(rhs)
		if tc.q[loc.Start.Offset:loc.End.Offset] != tc.want {
			t.Fatalf("%s RHS source %+v", tc.q, loc)
		}
	}
}

// The same source requires analysis expression boundaries while the legacy entry
// must retain its original dotted-identifier tokenization.
func TestParseAnalysisAndLegacyDotBoundaries(t *testing.T) {
	for _, q := range []string{`| eval x=host."other"`, `| eval x=host. "other"`, `| eval x=host. (other)`} {
		for _, analysisFirst := range []bool{true, false} {
			var parsed *parsedDocument
			var legacy *antlr.CommonTokenStream
			runAnalysis := func() { parsed = parseDocument(q) }
			runLegacy := func() {
				legacy = antlr.NewCommonTokenStream(parser.NewSPLLexer(antlr.NewInputStream(q)), antlr.TokenDefaultChannel)
				legacy.Fill()
			}
			if analysisFirst {
				runAnalysis()
				runLegacy()
			} else {
				runLegacy()
				runAnalysis()
			}
			if len(parsed.diagnostics) != 0 {
				t.Fatalf("analysis %q: %+v", q, parsed.diagnostics)
			}
			start := strings.Index(q, "host.")
			for _, tc := range []struct {
				name   string
				tokens *antlr.CommonTokenStream
				want   string
			}{{"analysis", parsed.tokens, "host"}, {"legacy", legacy, "host."}} {
				found := false
				for _, token := range tc.tokens.GetAllTokens() {
					if token.GetStart() == start {
						found = true
						if token.GetText() != tc.want {
							t.Errorf("%s %q (analysis first=%v) token %q, want %q", tc.name, q, analysisFirst, token.GetText(), tc.want)
						}
					}
				}
				if !found {
					t.Fatalf("%s %q: no token at field start", tc.name, q)
				}
			}
		}
	}
}

func TestParseAnalysisAndLegacyConcurrent(t *testing.T) {
	const q = `search * | eval x=host. "other"`
	const legacyQuery = `search * | fields host. "other"`
	m := mapper.NewWithConfig(&mapper.MappingConfig{Version: "1.0", Mappings: []mapper.FieldMapping{{Source: "host", Target: "server"}}})
	var wg sync.WaitGroup
	for i := 0; i < 12; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			r, err := Analyze(QueryDocument{Text: q})
			if err != nil {
				t.Error(err)
				return
			}
			if !r.Coverage.SyntaxComplete {
				t.Errorf("analysis: %+v", r.Diagnostics)
			}
		}()
		go func() {
			defer wg.Done()
			got, err := m.MapQuery(legacyQuery)
			if err != nil {
				t.Error(err)
				return
			}
			if got != legacyQuery {
				t.Errorf("legacy mapping = %q, want %q", got, legacyQuery)
			}
		}()
	}
	wg.Wait()
}
