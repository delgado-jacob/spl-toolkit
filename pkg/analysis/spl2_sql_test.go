package analysis

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

// This witnesses lexical evidence independently of the SQL execution schedule.
func TestSPL2SQLLexicalStagesAndProjectionRead(t *testing.T) {
	text := "SELECT host FROM main WHERE bytes>0"
	r := spl2AnalyzeTest(t, text)
	loc := func(a, b int) Location { return Location{Position{a, 1, a + 1}, Position{b, 1, b + 1}} }
	wantStages := []Stage{{"stage-0", "select", 2, "scope-0", loc(0, 11), true}, {"stage-1", "from", 0, "scope-0", loc(12, 21), true}, {"stage-2", "where", 1, "scope-0", loc(22, 35), true}}
	if r.Status != Valid || !reflect.DeepEqual(r.Stages, wantStages) {
		t.Fatalf("status/stages = %s %+v", r.Status, r.Stages)
	}
	if len(r.References) != 3 || len(r.Lineage) != 4 {
		t.Fatalf("one read per lexical operand: refs=%+v lineage=%+v", r.References, r.Lineage)
	}
	wantNames := []string{"host", "main", "bytes"}
	wantStagesIDs := []string{"stage-0", "stage-1", "stage-2"}
	for i, ref := range r.References {
		if ref.NormalizedName != wantNames[i] || ref.StageID != wantStagesIDs[i] || ref.Role != "read" || text[ref.Location.Start.Offset:ref.Location.End.Offset] != ref.OriginalName {
			t.Fatalf("lexical ref %d = %+v", i, ref)
		}
	}
	assertSQLPhases(t, r, []string{"source", "filter", "evaluate", "project"}, []string{"stage-1", "stage-2", "stage-0", "stage-0"})
	open := FieldState{Fields: []FieldBinding{}, Removed: []string{}, Open: true}
	filtered := FieldState{Fields: []FieldBinding{{Name: "bytes", OriginReferenceIDs: []string{"ref-2"}}}, Removed: []string{}, Open: true}
	evaluated := FieldState{Fields: []FieldBinding{{Name: "bytes", OriginReferenceIDs: []string{"ref-2"}}, {Name: "host", OriginReferenceIDs: []string{"ref-0"}}}, Removed: []string{}, Open: true}
	projected := FieldState{Fields: []FieldBinding{{Name: "host", OriginReferenceIDs: []string{"ref-0"}}}, Removed: []string{}}
	wantStates := []FieldState{open, filtered, evaluated, projected}
	for i, l := range r.Lineage {
		before := open
		if i > 0 {
			before = wantStates[i-1]
		}
		if !reflect.DeepEqual(l.Before, before) || !reflect.DeepEqual(l.After, wantStates[i]) {
			t.Fatalf("phase %d before/after = %+v / %+v", i, l.Before, l.After)
		}
	}
	if !reflect.DeepEqual(r.Lineage[3].Transitions, []Transition{{Operation: "project", Output: "host", InputReferenceIDs: []string{"ref-0"}}}) {
		t.Fatalf("project reused read: %+v", r.Lineage[3])
	}
	want := newResult(QueryDocument{Text: text, Language: "spl2", Profile: "splunkd", Version: "current"})
	want.Status = Valid
	want.Stages = wantStages
	want.Scopes = []Scope{{ID: "scope-0", Kind: "root", Location: loc(0, 35)}}
	want.Dependencies.Datasets = []string{"main"}
	want.References = []Reference{
		{ID: "ref-0", OriginalName: "host", NormalizedName: "host", Kind: "field", Role: "read", StageID: "stage-0", ScopeID: "scope-0", Location: loc(7, 11), Resolution: "exact", Binding: "source", OriginReferenceIDs: []string{}},
		{ID: "ref-1", OriginalName: "main", NormalizedName: "main", Kind: "dataset", Role: "read", StageID: "stage-1", ScopeID: "scope-0", Location: loc(17, 21), Resolution: "exact", Binding: "not_applicable", OriginReferenceIDs: []string{}},
		{ID: "ref-2", OriginalName: "bytes", NormalizedName: "bytes", Kind: "field", Role: "read", StageID: "stage-2", ScopeID: "scope-0", Location: loc(28, 33), Resolution: "exact", Binding: "source", OriginReferenceIDs: []string{}},
	}
	orders := []int{0, 1, 2, 3}
	want.Lineage = []Lineage{
		{StageID: "stage-1", ScopeID: "scope-0", Before: open, After: open, Transitions: []Transition{}, Phase: "source", ExecutionOrder: &orders[0]},
		{StageID: "stage-2", ScopeID: "scope-0", Before: open, After: filtered, Transitions: []Transition{}, Phase: "filter", ExecutionOrder: &orders[1]},
		{StageID: "stage-0", ScopeID: "scope-0", Before: filtered, After: evaluated, Transitions: []Transition{}, Phase: "evaluate", ExecutionOrder: &orders[2]},
		{StageID: "stage-0", ScopeID: "scope-0", Before: evaluated, After: projected, Transitions: []Transition{{Operation: "project", Output: "host", InputReferenceIDs: []string{"ref-0"}}}, Phase: "project", ExecutionOrder: &orders[3]},
	}
	if !reflect.DeepEqual(r, want) {
		t.Fatalf("complete SQL report differs: %+v", r)
	}
	for _, u := range []SourceUniverse{{Fields: []string{"host", "bytes"}, Complete: true}, {Fields: []string{"host"}}} {
		source, err := AnalyzeWithSourceUniverse(want.Document, u)
		if err != nil || !reflect.DeepEqual(source, &SourceAnalysis{Result: want, Expansions: []FieldExpansion{}}) {
			t.Fatalf("source report %+v %v", source, err)
		}
	}
	assertCorpusIntegrity(t, r)
}

func assertSQLPhases(t *testing.T, r *Result, phases, stages []string) {
	t.Helper()
	data, err := json.Marshal(r.Lineage)
	if err != nil {
		t.Fatal(err)
	}
	var entries []struct {
		Phase          string `json:"phase"`
		ExecutionOrder *int   `json:"execution_order"`
	}
	if err := json.Unmarshal(data, &entries); err != nil {
		t.Fatal(err)
	}
	if len(entries) != len(phases) {
		t.Fatalf("phases: %s", data)
	}
	for i, entry := range entries {
		if entry.Phase != phases[i] || entry.ExecutionOrder == nil || *entry.ExecutionOrder != i || r.Lineage[i].StageID != stages[i] {
			t.Fatalf("phase %d: %s", i, data)
		}
	}
}

func TestSPL2SQLAliasPreparationAndVisibility(t *testing.T) {
	r := spl2AnalyzeTest(t, "SELECT lower(user) AS owner FROM main ORDER BY owner")
	if r.Status != Valid {
		t.Fatalf("%+v", r)
	}
	assertSQLPhases(t, r, []string{"source", "evaluate", "order", "project"}, []string{"stage-1", "stage-0", "stage-2", "stage-0"})
	if len(r.References) != 4 || spl2Ref(t, r, "user", "read").Binding != "source" || spl2Ref(t, r, "owner", "read").Binding != "indeterminate" {
		t.Fatalf("refs %+v", r.References)
	}
	owner := FieldBinding{Name: "owner", OriginReferenceIDs: []string{"ref-1", "ref-0"}, Conditional: true}
	if !reflect.DeepEqual(r.Lineage[1].After.Fields, []FieldBinding{owner, {Name: "user", OriginReferenceIDs: []string{"ref-0"}}}) || !reflect.DeepEqual(r.Lineage[3].After, FieldState{Fields: []FieldBinding{owner}, Removed: []string{}}) {
		t.Fatalf("states %+v", r.Lineage)
	}
	if !reflect.DeepEqual(r.Lineage[3].Transitions, []Transition{{Operation: "project", Output: "owner", InputReferenceIDs: []string{"ref-1"}}}) {
		t.Fatalf("alias project %+v", r.Lineage[3])
	}
	where := spl2AnalyzeTest(t, "SELECT user AS owner FROM main WHERE owner>0 ORDER BY owner")
	if spl2Ref(t, where, "owner", "read").Binding != "source" {
		t.Fatalf("WHERE suppressed: %+v", where.References)
	}
	for _, text := range []string{
		"SELECT host FROM main ORDER BY bytes",
		"SELECT host FROM main HAVING host=1",
		"SELECT user AS user FROM main ORDER BY user",
		"SELECT user AS owner,owner FROM main ORDER BY owner",
		"SELECT user AS name,host AS name FROM main ORDER BY name",
		"SELECT user AS name,name AS other FROM main",
	} {
		t.Run(text, func(t *testing.T) {
			r := spl2AnalyzeTest(t, text)
			if r.Status != Incomplete || !spl2HasCode(r, CodeUnsupportedSemantics) || spl2HasCode(r, CodeUnavailableField) {
				t.Fatalf("%+v", r)
			}
		})
	}
	hidden := spl2AnalyzeTest(t, "SELECT host FROM main ORDER BY bytes")
	if spl2Ref(t, hidden, "bytes", "read").Binding != "indeterminate" {
		t.Fatalf("hidden source promoted: %+v", hidden.References)
	}
}

func TestSPL2SQLGroupedAggregatePhases(t *testing.T) {
	text := `SELECT host,sum(bytes) AS total FROM main WHERE active=true GROUP BY host HAVING total>0 ORDER BY total DESC LIMIT 3 OFFSET 1`
	r := spl2AnalyzeTest(t, text)
	if r.Status != Valid {
		t.Fatalf("%+v", r)
	}
	assertSQLPhases(t, r, []string{"source", "filter", "group", "aggregate", "having", "order", "project", "limit", "offset"}, []string{"stage-1", "stage-2", "stage-3", "stage-0", "stage-4", "stage-5", "stage-0", "stage-6", "stage-7"})
	if len(r.References) != 8 {
		t.Fatalf("duplicated reads: %+v", r.References)
	}
	for _, ref := range r.References {
		if ref.NormalizedName == "bytes" && ref.Binding != "source" {
			t.Fatalf("lost pregroup input: %+v", ref)
		}
	}
	if spl2Ref(t, r, "host", "group").Binding != "source" || spl2Ref(t, r, "host", "read").Binding != "source" || spl2Ref(t, r, "total", "read").Binding != "indeterminate" {
		t.Fatalf("bindings %+v", r.References)
	}
	host := FieldBinding{Name: "host", OriginReferenceIDs: []string{"ref-5"}}
	total := FieldBinding{Name: "total", OriginReferenceIDs: []string{"ref-2", "ref-1"}, Conditional: true}
	if !reflect.DeepEqual(r.Lineage[2].After, FieldState{Fields: []FieldBinding{host}, Removed: []string{}}) || !reflect.DeepEqual(r.Lineage[3].Before, r.Lineage[2].After) || !reflect.DeepEqual(r.Lineage[3].After, FieldState{Fields: []FieldBinding{host, total}, Removed: []string{}}) {
		t.Fatalf("group/aggregate states %+v", r.Lineage)
	}
	if !reflect.DeepEqual(r.Lineage[3].Transitions, []Transition{{Operation: "aggregate", Output: "total", InputReferenceIDs: []string{"ref-1"}, OutputReferenceID: "ref-2", Conditional: true}}) || !reflect.DeepEqual(r.Lineage[6].Transitions, []Transition{{Operation: "project", Output: "host", InputReferenceIDs: []string{"ref-0"}}, {Operation: "project", Output: "total", InputReferenceIDs: []string{"ref-2"}}}) {
		t.Fatalf("aggregate/project provenance %+v", r.Lineage)
	}
	for i := 1; i < len(r.Lineage); i++ {
		if !reflect.DeepEqual(r.Lineage[i].Before, r.Lineage[i-1].After) {
			t.Fatalf("discontinuous phase %d", i)
		}
	}
	assertCorpusIntegrity(t, r)
}

func TestSPL2SQLGroupedLimitsAndFinalBoundary(t *testing.T) {
	for _, query := range []string{
		`SELECT sum(bytes) AS total FROM main GROUP BY host HAVING host="a"`,
		`SELECT sum(bytes) AS total FROM main GROUP BY host ORDER BY host`,
		`SELECT sum(bytes) AS total FROM main GROUP BY host HAVING sum(other)>0`,
		`SELECT user,sum(bytes) AS total FROM main GROUP BY host`,
		`SELECT host,count() FROM main`,
		`SELECT count(user) FROM main`,
		`SELECT sum(bytes)+1 AS total FROM main`,
		`SELECT count() FROM main GROUP BY lower(host)`,
	} {
		t.Run(query, func(t *testing.T) {
			r := spl2AnalyzeTest(t, query)
			if r.Status != Incomplete || !spl2HasCode(r, CodeUnsupportedSemantics) || spl2HasCode(r, CodeUnavailableField) {
				t.Fatalf("%+v", r)
			}
			assertCorpusIntegrity(t, r)
		})
	}
	r := spl2AnalyzeTest(t, `SELECT sum(bytes) AS total FROM main GROUP BY host HAVING host="a"`)
	if spl2Ref(t, r, "host", "group").Binding != "source" || spl2Ref(t, r, "host", "read").Binding != "indeterminate" || r.Lineage[len(r.Lineage)-1].StageID != "stage-0" {
		t.Fatalf("hidden key %+v", r)
	}
	last := r.Lineage[len(r.Lineage)-1].After
	if len(last.Fields) != 1 || last.Fields[0].Name != "total" {
		t.Fatalf("hidden key leaked %+v", last)
	}
	r = spl2AnalyzeTest(t, `SELECT count(),sum(bytes),dc(action) FROM main | where count>0`)
	if r.Status != Valid || spl2Ref(t, r, "count", "read").Binding != "derived" {
		t.Fatalf("native labels %+v", r)
	}
	r = spl2AnalyzeTest(t, `SELECT host FROM main WHERE bytes>0 | where bytes>0`)
	if r.Status != Invalid || r.References[len(r.References)-1].Binding != "unavailable" || len(r.Stages) != 4 || r.Stages[3].Position != 3 {
		t.Fatalf("piped boundary %+v", r)
	}
}

func TestSPL2SQLExpressionEvidenceAndSourceIdentity(t *testing.T) {
	for _, c := range []struct {
		query, name          string
		conditional, removed bool
	}{
		{`SELECT isnull(missing) AS ok FROM main`, "ok", false, false},
		{`SELECT if(flag,null,null) AS gone FROM main`, "gone", false, true},
		{`SELECT coalesce(null,1) AS present FROM main`, "present", false, false},
		{`SELECT tonumber("17") AS number FROM main`, "number", true, false},
	} {
		t.Run(c.query, func(t *testing.T) {
			r := spl2AnalyzeTest(t, c.query)
			if r.Status != Valid {
				t.Fatalf("%+v", r)
			}
			state := r.Lineage[len(r.Lineage)-1].After
			if c.removed {
				if len(state.Fields) != 0 || !reflect.DeepEqual(state.Removed, []string{c.name}) {
					t.Fatalf("%+v", state)
				}
			} else if len(state.Fields) != 1 || state.Fields[0].Name != c.name || state.Fields[0].Conditional != c.conditional {
				t.Fatalf("%+v", state)
			}
			if c.name == "ok" && (spl2Ref(t, r, "missing", "null_test").Binding != "source" || len(r.Lineage[1].After.Fields) != 1) {
				t.Fatalf("null test established input %+v", r)
			}
		})
	}
	doc := QueryDocument{Text: `SELECT 'actor.name' FROM main`, Language: "spl2"}
	u := SourceUniverse{Fields: []string{"actor.name"}, Complete: true, Resolve: func(string) SourceFieldAdmission { return SourceFieldAdmitted }}
	r, err := AnalyzeWithSourceUniverse(doc, u)
	if err != nil || r.Result.Status != Incomplete || spl2Ref(t, r.Result, "actor.name", "read").Binding != "indeterminate" || !r.Result.Lineage[2].After.Fields[0].Conditional {
		t.Fatalf("identity %+v %v", r, err)
	}
}

func TestSPL2SQLCollisionDoesNotProveFinalOutput(t *testing.T) {
	r := spl2AnalyzeTest(t, `SELECT user AS owner,owner FROM main | where owner>0`)
	last := r.References[len(r.References)-1]
	if r.Status != Incomplete || last.Binding != "indeterminate" || !r.Lineage[len(r.Lineage)-1].After.Fields[0].Conditional {
		t.Fatalf("collision gained certainty: %+v", r)
	}
}

func TestSPL2SQLGroupedFieldUsesEvaluatePhase(t *testing.T) {
	r := spl2AnalyzeTest(t, `SELECT DISTINCT host FROM main GROUP BY host ORDER BY host`)
	if r.Status != Valid {
		t.Fatalf("%+v", r)
	}
	assertSQLPhases(t, r, []string{"source", "group", "evaluate", "order", "project"}, []string{"stage-1", "stage-2", "stage-0", "stage-3", "stage-0"})
}

func TestSPL2SQLEquivalentFormsRetainOwnLocations(t *testing.T) {
	for _, c := range []struct {
		text      string
		hostID    string
		hostStart int
	}{
		{"SELECT host FROM main WHERE bytes>0", "ref-0", 7},
		{"FROM main WHERE bytes>0 SELECT host", "ref-2", 31},
		{"FROM main | where bytes>0 | table host", "ref-2", 34},
	} {
		t.Run(c.text, func(t *testing.T) {
			r := spl2AnalyzeTest(t, c.text)
			want := FieldState{Fields: []FieldBinding{{Name: "host", OriginReferenceIDs: []string{c.hostID}}}, Removed: []string{}}
			if r.Status != Valid || !reflect.DeepEqual(r.Lineage[len(r.Lineage)-1].After, want) {
				t.Fatalf("%+v", r)
			}
			ref := spl2Ref(t, r, "host", "read")
			if ref.Location.Start.Offset != c.hostStart || c.text[ref.Location.Start.Offset:ref.Location.End.Offset] != "host" {
				t.Fatalf("%+v", ref)
			}
			assertCorpusIntegrity(t, r)
		})
	}
	text := "SELECT 'café' AS 'étiquette'\r\nFROM 'données'\r\nORDER BY 'étiquette' | where 'étiquette'!=\"\""
	r, err := Analyze(QueryDocument{Text: text, Language: "spl2", SourceID: " exact\r\n "})
	if err != nil || r.Status != Valid || r.Document.Text != text || r.Document.SourceID != " exact\r\n " {
		t.Fatalf("%+v %v", r, err)
	}
	ref := spl2Ref(t, r, "café", "read")
	if ref.Location.Start != (Position{7, 1, 8}) || ref.Location.End != (Position{14, 1, 14}) {
		t.Fatalf("%+v", ref)
	}
	for i, want := range []string{"SELECT 'café' AS 'étiquette'", "FROM 'données'", "ORDER BY 'étiquette'", `where 'étiquette'!=""`} {
		loc := r.Stages[i].Location
		if text[loc.Start.Offset:loc.End.Offset] != want {
			t.Fatalf("stage %d %+v", i, loc)
		}
	}
	assertCorpusIntegrity(t, r)
}

func TestFinalizeSQLPhaseAndSourceExpansionIDs(t *testing.T) {
	doc := QueryDocument{Text: `SELECT 1 AS 'actor.local',host FROM main | fields '*'`, Language: "spl2"}
	wantState := FieldState{Fields: []FieldBinding{{Name: "actor.local", OriginReferenceIDs: []string{"ref-0"}}, {Name: "host", OriginReferenceIDs: []string{"ref-1"}}}, Removed: []string{}}
	for _, u := range []SourceUniverse{{Fields: []string{"actor.name", "host"}, Complete: true}, {Fields: []string{"actor.name", "host"}}, {Fields: []string{"actor.name", "host"}, Complete: true, Resolve: func(string) SourceFieldAdmission { return SourceFieldAdmitted }}} {
		r, err := AnalyzeWithSourceUniverse(doc, u)
		if err != nil || r.Result.Status != Valid || !reflect.DeepEqual(r.Result.Lineage[3].After, wantState) {
			t.Fatalf("state %+v %v", r.Result.Lineage, err)
		}
		want := []FieldExpansion{{ReferenceID: "ref-3", Complete: true, Matches: []ExpandedField{{Name: "actor.local", Binding: "derived"}, {Name: "host", Binding: "source"}}}}
		if !reflect.DeepEqual(r.Expansions, want) {
			t.Fatalf("expansions %+v", r.Expansions)
		}
		if !reflect.DeepEqual(r.Result.Lineage[2].Transitions, []Transition{{Operation: "project", Output: "actor.local", InputReferenceIDs: []string{"ref-0"}}, {Operation: "project", Output: "host", InputReferenceIDs: []string{"ref-1"}}}) {
			t.Fatalf("project %+v", r.Result.Lineage[2])
		}
		assertCorpusIntegrity(t, r.Result)
	}
}

func TestSPL2SQLMetadataDoesNotChangeOrdinaryWire(t *testing.T) {
	for _, doc := range []QueryDocument{{Text: "search host=* | table host"}, {Text: "FROM main | eval x=bytes | table x", Language: "spl2"}} {
		r, err := Analyze(doc)
		if err != nil {
			t.Fatal(err)
		}
		data, err := json.Marshal(r)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(data), `"phase"`) || strings.Contains(string(data), `"execution_order"`) {
			t.Fatalf("metadata added to ordinary wire: %s", data)
		}
	}
}

func TestSPL2SQLRecoveryDoesNotInventClauses(t *testing.T) {
	r := spl2AnalyzeTest(t, "SELECT host")
	if r.Status != Invalid || len(r.Stages) != 1 || r.Stages[0].Command != "select" {
		t.Fatalf("recovered missing FROM %+v", r)
	}
	for _, l := range r.Lineage {
		if l.Phase == "source" {
			t.Fatal("invented missing source phase")
		}
	}
	for _, text := range []string{"SELECT FROM main", "SELECT host WHERE bytes>0 FROM main", "FROM main GROUP BY SELECT count()"} {
		t.Run(text, func(t *testing.T) {
			r := spl2AnalyzeTest(t, text)
			if r.Status != Invalid {
				t.Fatalf("%+v", r)
			}
			assertCorpusIntegrity(t, r)
		})
	}
}

func TestSPL2SQLGlobalOrderAfterPipelineAndSourceReset(t *testing.T) {
	r := spl2AnalyzeTest(t, `FROM first | eval prior=1 | SELECT host FROM main ORDER BY host`)
	if r.Status != Valid || len(r.Stages) != 5 || len(r.Lineage) != 6 {
		t.Fatalf("%+v", r)
	}
	for i, phase := range []string{"source", "evaluate", "order", "project"} {
		l := r.Lineage[i+2]
		if l.Phase != phase || l.ExecutionOrder == nil || *l.ExecutionOrder != i+2 {
			t.Fatalf("global order %+v", l)
		}
	}
	if r.Stages[2].Position != 3 || r.Stages[3].Position != 2 || r.Stages[4].Position != 4 || len(r.Lineage[2].After.Fields) != 0 {
		t.Fatalf("reset/positions %+v", r)
	}
	assertCorpusIntegrity(t, r)
}

func TestSPL2SQLAggregateWildcardKeepsSourceGuards(t *testing.T) {
	doc := QueryDocument{Text: `SELECT count('actor*') AS n FROM main`, Language: "spl2"}
	u := SourceUniverse{Fields: []string{"actor.name", "actor"}, Complete: true, Resolve: func(string) SourceFieldAdmission { return SourceFieldAdmitted }}
	r, err := AnalyzeWithSourceUniverse(doc, u)
	if err != nil || r.Result.Status != Incomplete || len(r.Expansions) != 1 {
		t.Fatalf("%+v %v", r, err)
	}
	want := []FieldExpansion{{ReferenceID: "ref-0", Complete: false, Matches: []ExpandedField{{Name: "actor", Binding: "source"}}}}
	if !reflect.DeepEqual(r.Expansions, want) || !reflect.DeepEqual(r.Result.Lineage[2].After, FieldState{Fields: []FieldBinding{{Name: "n", OriginReferenceIDs: []string{"ref-1", "ref-0"}, Conditional: true}}, Removed: []string{}, Uncertain: true}) {
		t.Fatalf("guard/IDs %+v %+v", r.Expansions, r.Result.Lineage)
	}
	assertCorpusIntegrity(t, r.Result)
}

func TestSPL2SQLCompoundAggregateCallContracts(t *testing.T) {
	for _, c := range []struct {
		name, expression, code, diagnosticText string
		status                                 Status
		fields                                 []string
	}{
		{"invalid if", `if(count(),1)`, CodeSyntaxError, `if(count(),1)`, Invalid, nil},
		{"valid if", `if(count(),1,2)`, "", "", Incomplete, nil},
		{"invalid if with reads", `if(sum(bytes),backup)`, CodeSyntaxError, `if(sum(bytes),backup)`, Invalid, []string{"bytes", "backup"}},
		{"valid if with reads", `if(sum(bytes),backup,fallback)`, "", "", Incomplete, []string{"bytes", "backup", "fallback"}},
		{"nested invalid wrapper", `coalesce(if(sum(bytes),backup),fallback)`, CodeSyntaxError, `if(sum(bytes),backup)`, Invalid, []string{"bytes", "backup", "fallback"}},
		{"valid numeric wrapper", `round(sum(bytes),2)`, "", "", Incomplete, []string{"bytes"}},
		{"invalid numeric wrapper", `round(sum(bytes),precision,extra)`, CodeSyntaxError, `round(sum(bytes),precision,extra)`, Invalid, []string{"bytes", "precision", "extra"}},
		{"unknown wrapper", `mystery(sum(bytes),backup)`, CodeUnsupportedFunction, `mystery(sum(bytes),backup)`, Incomplete, []string{"bytes", "backup"}},
		{"named wrapper", `if(sum(bytes),then:backup)`, CodeUnsupportedSemantics, `if(sum(bytes),then:backup)`, Incomplete, []string{"bytes", "backup"}},
		{"wrong profile wrapper", `batch_id(sum(bytes))`, "SPL_PROFILE_MISMATCH", `batch_id(sum(bytes))`, Invalid, []string{"bytes"}},
		{"null inspection wrapper", `isnull(count())`, "", "", Incomplete, nil},
		{"all-null wrapper stays unproved", `if(count(),null,null)`, "", "", Incomplete, nil},
	} {
		t.Run(c.name, func(t *testing.T) {
			text := "SELECT " + c.expression + " AS total FROM main"
			r := spl2AnalyzeTest(t, text)
			if r.Status != c.status || r.Coverage.SemanticComplete || !spl2HasCode(r, CodeUnsupportedSemantics) {
				t.Fatalf("status/coverage: %+v", r)
			}
			if c.code != "" && !spl2HasCode(r, c.code) {
				t.Fatalf("missing wrapper contract %s: %+v", c.code, r.Diagnostics)
			}
			if c.status != Invalid && spl2HasCode(r, CodeSyntaxError) {
				t.Fatalf("false scalar-context error: %+v", r.Diagnostics)
			}
			if c.diagnosticText != "" {
				found := false
				for _, d := range r.Diagnostics {
					if d.Code == c.code && text[d.Location.Start.Offset:d.Location.End.Offset] == c.diagnosticText && d.StageID == "stage-0" {
						found = true
					}
				}
				if !found {
					t.Fatalf("missing located wrapper diagnostic %+v", r.Diagnostics)
				}
			}
			if len(r.References) != len(c.fields)+2 {
				t.Fatalf("duplicated or fabricated operands: %+v", r.References)
			}
			for i, name := range c.fields {
				ref := r.References[i]
				binding := "source"
				// The inner invalid call marks the environment uncertain before
				// the outer coalesce's later fallback operand is inspected.
				if c.name == "nested invalid wrapper" && name == "fallback" {
					binding = "indeterminate"
				}
				if ref.NormalizedName != name || ref.Role != "read" || ref.Binding != binding || ref.StageID != "stage-0" {
					t.Fatalf("original operand %d: %+v", i, ref)
				}
			}
			last := r.Lineage[len(r.Lineage)-1].After
			if len(last.Fields) != 1 || last.Fields[0].Name != "total" || !last.Fields[0].Conditional {
				t.Fatalf("unproved compound acquired certainty: %+v", last)
			}
			assertCorpusIntegrity(t, r)
		})
	}
}
