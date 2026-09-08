package analysis

import (
	"github.com/antlr4-go/antlr/v4"
	"testing"
)

func TestSPL2SQLSyntaxBoundaries(t *testing.T) {
	for _, c := range []struct {
		name, text    string
		invalid, held bool
	}{
		{"join_inequality", "FROM main AS m JOIN users AS u ON m.id>u.id", true, false},
		{"join_literal", "FROM main AS m JOIN users AS u ON true", true, false},
		{"join_or", "FROM main AS m JOIN users AS u ON m.id=u.id OR m.realm=u.realm", true, false},
		{"join_unqualified", "FROM main AS m JOIN users AS u ON id=u.id", true, false},
		{"join_deep_path", "FROM main AS m JOIN users AS u ON m.user.id=u.id", false, true},
		{"join_reversed", "FROM main AS m JOIN users AS u ON u.id=m.id", false, false},
		{"foreign_sql_option", "FROM main | stats limit=3 count()", false, true},
		{"negative_huge", "FROM main LIMIT -9999999999999999999999999999999", false, true},
		{"span_fraction", "FROM main GROUP BY span(_time,1.5h) SELECT count()", true, false},
		{"span_assignment_fraction", "FROM main GROUP BY _time span=(1.5h) SELECT count()", true, false},
		{"exists_parenthesized_fields", "FROM orders AS o WHERE EXISTS (SELECT id FROM inventory WHERE (id)=(o.item))", false, false},
		{"exists_child_from", "FROM orders AS o WHERE EXISTS (FROM inventory WHERE id=o.item SELECT id)", false, false},
		{"exists_shadowed_alias", "FROM orders AS o WHERE EXISTS (SELECT id FROM inventory AS o WHERE id=o.item)", true, false},
		{"span_target_number", "FROM main GROUP BY 1 span=(h) SELECT count()", true, false},
		{"span_target_call", "FROM main GROUP BY lower(x) span=(h) SELECT count()", true, false},
		{"span_target_computation", "FROM main GROUP BY x+1 span=(h) SELECT count()", true, false},
		{"span_target_qualified", "FROM main AS m GROUP BY m._time span=(h) SELECT count()", false, true},
		{"span_target_deep", "FROM main GROUP BY payload.time.value span=(h) SELECT count()", false, true},
		{"span_target_literal_dot", "FROM main GROUP BY 'payload.time' span=(h) SELECT count()", false, false},
		{"group_expression_number", "FROM main GROUP BY 1 SELECT count()", false, false},
		{"group_expression_call", "FROM main GROUP BY lower(x) SELECT count()", false, false},
		{"select_where", "SELECT host FROM main WHERE bytes>0", false, false},
		{"wrong_order", "SELECT host WHERE bytes>0 FROM main", true, false},
		{"lowercase", "select distinct host from main where bytes>0 group by host having host=\"a\" order by host desc limit 3 offset 1", false, false},
		{"adjacency", "FROM main WHERE x=1 y=2", true, false},
		{"integer", "FROM main LIMIT 1.5", true, false},
		{"negative_limit", "FROM main LIMIT -1", false, true},
		{"negative_offset", "FROM main OFFSET -2", false, true},
		{"term_direction", "FROM main ORDER BY account ASC,size DESC", false, true},
		{"span_unparenthesized", "FROM main GROUP BY _time span=3h SELECT count()", false, true},
		{"span_unit", "SELECT count() FROM main GROUP BY span(_time,hour)", false, false},
		{"span_extra", "SELECT count() FROM main GROUP BY span(_time,hour,day)", true, false},
		{"exists_offset", "FROM orders AS o WHERE EXISTS (SELECT id FROM inventory WHERE id=o.item OFFSET 1)", true, false},
		{"exists_having", "FROM orders AS o WHERE active=true SELECT item HAVING EXISTS (SELECT id FROM inventory WHERE id=o.item)", false, false},
		{"exists_wrong_where", "FROM orders AS o WHERE EXISTS (SELECT id FROM inventory WHERE id=o.item) SELECT item HAVING item>0", true, false},
		{"exists_no_correlation", "FROM orders AS o WHERE EXISTS (SELECT id FROM inventory WHERE id=1)", true, false},
		{"exists_wrong_alias", "FROM orders AS o WHERE EXISTS (SELECT id FROM inventory WHERE id=z.item)", true, false},
		{"exists_inequality", "FROM orders AS o WHERE EXISTS (SELECT id FROM inventory WHERE id>o.item)", true, false},
		{"exists_literal_dot", "FROM orders AS o WHERE EXISTS (SELECT id FROM inventory WHERE id='o.item')", true, false},
		{"exists_projection", "SELECT EXISTS (SELECT id FROM inventory WHERE id=o.item) FROM orders AS o", true, false},
		{"exists_nested_projection", "FROM orders AS o WHERE EXISTS (SELECT EXISTS (SELECT id FROM child WHERE id=o.item) FROM inventory WHERE id=o.item)", true, false},
		{"exists_lambda", "FROM orders AS o WHERE map(a, ($v)->EXISTS (SELECT id FROM inventory WHERE id=o.item))", true, false},
		{"missing_group_key", "FROM main GROUP BY SELECT count()", true, false},
		{"group_wildcard_escape", "FROM main GROUP BY 'host\\u002a' SELECT count()", true, false},
	} {
		t.Run(c.name, func(t *testing.T) {
			p := parseSPL2Document(c.text)
			invalid := false
			for _, d := range p.diagnostics {
				if d.Severity == "error" {
					invalid = true
				}
			}
			if invalid != c.invalid {
				t.Errorf("invalid=%v want %v: %+v", invalid, c.invalid, p.diagnostics)
			}
			if !c.invalid && p.syntaxComplete == c.held {
				t.Errorf("syntax complete=%v held=%v: %+v", p.syntaxComplete, c.held, p.diagnostics)
			}
			if p.semanticComplete {
				t.Fatal("syntax frontend claimed semantics")
			}
		})
	}
}

func TestSPL2SQLSyntaxClauseOwnership(t *testing.T) {
	text := "SELECT a.'café' AS label,'a.café',payload.user.name\r\nFROM main AS a\r\nWHERE a.bytes>0\r\nGROUP BY a.'café','a.café',payload.user.name\r\nHAVING label!=\"\"\r\nORDER BY label DESC\r\nLIMIT 2\r\nOFFSET 1 | table label"
	p := parseSPL2Document(text)
	if len(p.diagnostics) > 0 {
		t.Fatalf("%+v", p.diagnostics)
	}
	sql := p.tree.Pipeline().Start_().SelectCommand()
	for _, c := range []struct {
		ctx  antlr.ParserRuleContext
		text string
		line int
	}{
		{sql.SqlSelectClause(), "SELECT a.'café' AS label,'a.café',payload.user.name", 1},
		{sql.SqlFromClause(), "FROM main AS a", 2},
		{sql.SqlWhereClause(), "WHERE a.bytes>0", 3},
		{sql.SqlGroupClause(), "GROUP BY a.'café','a.café',payload.user.name", 4},
		{sql.SqlHavingClause(), "HAVING label!=\"\"", 5},
		{sql.SqlOrderClause(), "ORDER BY label DESC", 6},
		{sql.SqlLimitClause(), "LIMIT 2", 7},
		{sql.SqlOffsetClause(), "OFFSET 1", 8},
	} {
		spl2PipelineText(t, p, c.ctx, c.text)
		loc := p.source.contextLocation(c.ctx)
		if loc.Start.Line != c.line || loc.Start.Column != 1 {
			t.Errorf("clause location %+v", loc)
		}
	}
	projections := sql.SqlSelectClause().AllProjection()
	qualified := spl2SingleAccess(projections[0].Expression())
	literal := spl2SingleAccess(projections[1].Expression())
	nested := spl2SingleAccess(projections[2].Expression())
	if len(qualified.AllAccessPart()) != 1 || qualified.AccessPart(0).DOT() == nil || len(literal.AllAccessPart()) != 0 || len(nested.AllAccessPart()) != 2 {
		t.Fatal("qualified, literal-dot and nested access collapsed")
	}
	spl2PipelineText(t, p, qualified.Primary().FieldName(), "a")
	spl2PipelineText(t, p, qualified.AccessPart(0).Identifier(), "'café'")
	spl2PipelineText(t, p, literal.Primary().FieldName(), "'a.café'")
	spl2PipelineText(t, p, nested.Primary().FieldName(), "payload")
	spl2PipelineText(t, p, sql.SqlFromClause().Dataset(), "main")
	spl2PipelineText(t, p, sql.SqlFromClause().SourceAlias().Identifier(), "a")
	spl2PipelineText(t, p, projections[0].ProjectionAlias(), "label")
	loc := p.source.contextLocation(qualified.AccessPart(0).Identifier())
	if loc.Start.Offset != 9 || loc.End.Offset != 16 || loc.Start.Column != 10 || loc.End.Column != 16 {
		t.Fatalf("Unicode byte/column location %+v", loc)
	}
	if len(p.tree.Pipeline().AllCommand()) != 1 || p.tree.Pipeline().Command(0).TableCommand() == nil {
		t.Fatal("SQL swallowed downstream pipeline")
	}
}

func TestSPL2SQLSyntaxFromJoinAndChildOwnership(t *testing.T) {
	p := parseSPL2Document("FROM orders AS o LEFT OUTER JOIN inventory AS i ON o.item=i.id WHERE active=true GROUP BY o.item SELECT o.item,count() AS n HAVING EXISTS (SELECT id FROM stock WHERE id=o.item) ORDERBY n DESC LIMIT 4 OFFSET 1 | where n>0")
	if len(p.diagnostics) > 0 {
		t.Fatalf("%+v", p.diagnostics)
	}
	sql := p.tree.Pipeline().Start_().FromCommand()
	join := sql.SqlFromClause().SqlJoinClause(0)
	if join.LEFT() == nil || join.OUTER() == nil || join.INNER() != nil {
		t.Fatal("left outer join kind lost")
	}
	spl2PipelineText(t, p, join.Dataset(), "inventory")
	spl2PipelineText(t, p, join.SourceAlias(), "AS i")
	spl2PipelineText(t, p, join.SqlJoinPredicate(), "o.item=i.id")
	having := spl2SingleAccess(sql.SqlHavingClause().SqlPredicate().Expression()).Primary().ExistsPredicate()
	child := having.SelectCommand()
	spl2PipelineText(t, p, child.SqlFromClause().Dataset(), "stock")
	spl2PipelineText(t, p, child.SqlWhereClause(), "WHERE id=o.item")
	if child.SqlLimitClause() != nil || child.SqlOffsetClause() != nil {
		t.Fatal("parent limits leaked into child")
	}
	if child.GetParent() != having {
		t.Fatal("child scope lost EXISTS ownership")
	}
}

func TestSPL2SQLSyntaxContractLocations(t *testing.T) {
	for _, c := range []struct{ text, excerpt string }{
		{"FROM orders AS o WHERE EXISTS (SELECT id FROM inventory WHERE id=o.item LIMIT 1)", "LIMIT 1"},
		{"FROM orders AS o WHERE EXISTS (SELECT id FROM inventory WHERE id=o.item OFFSET 1)", "OFFSET 1"},
		{"FROM main LIMIT 1.5", "1.5"},
		{"FROM main GROUP BY 'host\\u002a' SELECT count()", "'host\\u002a'"},
		{"FROM main GROUP BY span(_time,1.5h) SELECT count()", "1.5h"},
	} {
		p := parseSPL2Document(c.text)
		found := false
		for _, d := range p.diagnostics {
			if d.Category == "contract" && d.Code == CodeSyntaxError && c.text[d.Location.Start.Offset:d.Location.End.Offset] == c.excerpt {
				found = true
			}
		}
		if !found {
			t.Errorf("missing exact contract %q: %+v", c.excerpt, p.diagnostics)
		}
	}
	for _, text := range []string{"FROM main LIMIT", "FROM main GROUP BY span(_time,) SELECT count()", "FROM main AS m WHERE EXISTS (SELECT FROM users LIMIT)"} {
		p := parseSPL2Document(text)
		for _, d := range p.diagnostics {
			if d.Category == "contract" {
				t.Errorf("recovered tokens generated contract: %s %+v", text, d)
			}
		}
	}
	p := parseSPL2Document("FROM orders AS o WHERE EXISTS (SELECT id FROM inventory WHERE id=o.item LIMIT -1)")
	invalid, held := false, false
	for _, d := range p.diagnostics {
		invalid = invalid || d.Severity == "error"
		held = held || d.Code == CodeUnsupportedSemantics
	}
	if !invalid || !held || p.syntaxComplete {
		t.Fatalf("invalid plus incomplete coverage lost: %+v", p.diagnostics)
	}
}
