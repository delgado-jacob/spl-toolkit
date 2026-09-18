package analysis

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/antlr4-go/antlr/v4"
	"github.com/delgado-jacob/spl-toolkit/parser/spl2"
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
	want.Requirements = RequirementSet{
		SchemaVersion: 1,
		Query: RequirementQueryIdentity{
			Language: "spl2", Profile: "splunkd", Version: "current",
			QueryDigest: "sha256:aeacf92af76b0a8b74aacd3863908ec66d33802ee4cf5ea0ffa80f9c9d78129f",
		},
		CapabilityRevision: "sha256:0203cbeec2e0fd484080b1f512582a1c8bdc4bc541fb73ac42c013060695f84c",
		QueryStatus:        Valid,
		Coverage:           RequirementCoverage{Complete: true, Reasons: []string{}},
		Items: []RequirementItem{
			{ID: "req-1", Kind: "field", Identity: "host", Role: "read", Necessity: "required", Origin: "direct", Resolution: "exact", Occurrences: []RequirementOccurrence{{ReferenceID: "ref-0", OriginalName: "host", Binding: "source", StageID: "stage-0", ScopeID: "scope-0", Location: loc(7, 11)}}},
			{ID: "req-2", Kind: "dataset", Identity: "main", Role: "read", Necessity: "required", Origin: "direct", Resolution: "exact", Occurrences: []RequirementOccurrence{{ReferenceID: "ref-1", OriginalName: "main", Binding: "not_applicable", StageID: "stage-1", ScopeID: "scope-0", Location: loc(17, 21)}}},
			{ID: "req-3", Kind: "field", Identity: "bytes", Role: "read", Necessity: "required", Origin: "direct", Resolution: "exact", Occurrences: []RequirementOccurrence{{ReferenceID: "ref-2", OriginalName: "bytes", Binding: "source", StageID: "stage-2", ScopeID: "scope-0", Location: loc(28, 33)}}},
		},
		Gaps:        []RequirementGap{},
		Diagnostics: []Diagnostic{},
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

func TestSPL2SQLAggregateOnlyHaving(t *testing.T) {
	for _, text := range []string{"SELECT count() AS n FROM main HAVING n>2", "FROM main SELECT count() AS n HAVING n>2"} {
		t.Run(text, func(t *testing.T) {
			r := spl2AnalyzeTest(t, text)
			if r.Status != Valid || !r.Coverage.SyntaxComplete || !r.Coverage.SemanticComplete || len(r.Diagnostics) != 0 {
				t.Fatalf("whole-input aggregate HAVING: %+v", r)
			}
			selectID, sourceID := "stage-0", "stage-1"
			createID := "ref-0"
			if strings.HasPrefix(text, "FROM") {
				selectID, sourceID, createID = "stage-1", "stage-0", "ref-1"
			}
			assertSQLPhases(t, r, []string{"source", "aggregate", "having", "project"}, []string{sourceID, selectID, "stage-2", selectID})
			if len(r.References) != 3 || len(r.Scopes) != 1 || len(r.Stages) != 3 {
				t.Fatalf("extra lexical evidence %+v", r)
			}
			created, read := spl2Ref(t, r, "n", "output"), spl2Ref(t, r, "n", "read")
			if created.ID != createID || created.StageID != selectID || read.StageID != "stage-2" || read.Binding != "derived" || !reflect.DeepEqual(read.OriginReferenceIDs, []string{createID}) {
				t.Fatalf("HAVING did not consume selected aggregate: %+v", r.References)
			}
			field := FieldBinding{Name: "n", OriginReferenceIDs: []string{createID}}
			closed := FieldState{Fields: []FieldBinding{field}, Removed: []string{}}
			for _, phase := range r.Lineage[1:] {
				if !reflect.DeepEqual(phase.After, closed) {
					t.Fatalf("aggregate presence/projection: %+v", phase)
				}
			}
			if !reflect.DeepEqual(r.Lineage[1].Transitions, []Transition{{Operation: "aggregate", Output: "n", InputReferenceIDs: []string{}, OutputReferenceID: createID}}) || !reflect.DeepEqual(r.Lineage[3].Transitions, []Transition{{Operation: "project", Output: "n", InputReferenceIDs: []string{createID}}}) {
				t.Fatalf("phase transition ownership %+v", r.Lineage)
			}
			assertCorpusIntegrity(t, r)
		})
	}
	for _, text := range []string{
		"FROM main SELECT host HAVING host>2",
		"FROM main SELECT count() AS n HAVING hidden>2",
		"FROM main SELECT count() AS n,host HAVING n>2",
		"FROM main SELECT count() AS n,count() AS n HAVING n>2",
		"FROM main SELECT mystery() AS n HAVING n>2",
	} {
		r := spl2AnalyzeTest(t, text)
		if r.Status != Incomplete || r.Coverage.SemanticComplete {
			t.Fatalf("unproved HAVING promoted: %s %+v", text, r)
		}
	}
	r := spl2AnalyzeTest(t, "FROM main SELECT sum(bytes) AS total HAVING total>2")
	if r.Status != Valid || !r.Coverage.SemanticComplete || spl2Ref(t, r, "total", "read").Binding != "indeterminate" || !r.Lineage[len(r.Lineage)-1].After.Fields[0].Conditional {
		t.Fatalf("conditional aggregate presence promoted: %+v", r)
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

func TestSPL2SQLFiniteProjectionKeepsPartialExpansionUncertainty(t *testing.T) {
	for _, complete := range []bool{false, true} {
		u := SourceUniverse{Fields: []string{"actor"}, Complete: complete, Resolve: func(string) SourceFieldAdmission { return SourceFieldAdmitted }}
		r, err := AnalyzeWithSourceUniverse(QueryDocument{Text: `SELECT count('actor*') AS n FROM main`, Language: "spl2"}, u)
		if err != nil || len(r.Expansions) != 1 || r.Expansions[0].Complete != complete {
			t.Fatalf("source expansion changed: %+v %v", r, err)
		}
		last := r.Result.Lineage[len(r.Result.Lineage)-1].After
		if last.Open || last.Uncertain != !complete || len(last.Fields) != 1 || last.Fields[0].Name != "n" || !last.Fields[0].Conditional {
			t.Fatalf("partial=%v expansion final guard: %+v", !complete, last)
		}
		assertCorpusIntegrity(t, r.Result)
	}

	// An earlier expansion is not a dependency of the later SELECT's fresh input.
	u := SourceUniverse{Fields: []string{"actor"}, Complete: false, Resolve: func(string) SourceFieldAdmission { return SourceFieldAdmitted }}
	r, err := AnalyzeWithSourceUniverse(QueryDocument{Text: `FROM main | stats count('actor*') AS previous | SELECT mystery(host) AS output FROM other`, Language: "spl2"}, u)
	if err != nil || len(r.Expansions) != 1 || r.Expansions[0].Complete {
		t.Fatalf("expected earlier partial expansion: %+v %v", r, err)
	}
	last := r.Result.Lineage[len(r.Result.Lineage)-1].After
	if last.Open || last.Uncertain || len(last.Fields) != 1 || last.Fields[0].Name != "output" || !last.Fields[0].Conditional {
		t.Fatalf("unrelated expansion tainted exact later destination: %+v", last)
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

func TestSPL2SQLRecoveredClauseEvidence(t *testing.T) {
	for _, q := range []string{
		`SELECT L.id FROM main AS L JOIN users AS R ON L.id=R.id`,
		`SELECT host FROM main AS L JOIN users AS R ON`,
		`FROM main AS L JOIN users AS R ON`,
	} {
		t.Run(q, func(t *testing.T) {
			r := spl2AnalyzeTest(t, q)
			if !reflect.DeepEqual(r.Dependencies.Datasets, []string{"main", "users"}) {
				t.Fatalf("intact datasets lost: %+v", r)
			}
			assertCorpusIntegrity(t, r)
		})
	}
	for _, q := range []string{
		`FROM main GROUP BY SELECT host,count() AS n`,
		`FROM main GROUP BY host, SELECT host,count() AS n`,
		`FROM main | eval 'é'=1 | FROM main GROUP BY SELECT host,count() AS n | table n`,
	} {
		t.Run(q, func(t *testing.T) {
			r := spl2AnalyzeTest(t, q)
			if r.Status != Invalid || r.Coverage.SyntaxComplete {
				t.Fatalf("damage promoted: %+v", r)
			}
			spl2Ref(t, r, "n", "output")
			selectStart := strings.Index(q, "SELECT")
			found := false
			for _, stage := range r.Stages {
				if stage.Command == "select" && stage.Location.Start.Offset == selectStart {
					found = true
				}
			}
			if !found {
				t.Fatalf("real SELECT boundary lost: %+v", r)
			}
			assertCorpusIntegrity(t, r)
		})
	}
}

func TestSPL2SQLEmptyHavingRetainsGroupEvidence(t *testing.T) {
	q := `SELECT sum(size) AS total FROM main GROUP BY server HAVING`
	r := spl2AnalyzeTest(t, q)
	if r.Status != Invalid || r.Coverage.SyntaxComplete {
		t.Fatalf("empty HAVING promoted: %+v", r)
	}
	if spl2Ref(t, r, "size", "read").Binding != "source" || spl2Ref(t, r, "server", "group").Binding != "source" {
		t.Fatalf("intact pregroup inputs lost: %+v", r)
	}
	for _, stage := range r.Stages {
		if stage.Command == "group" && !stage.SemanticComplete {
			t.Fatalf("intact GROUP damaged: %+v", r)
		}
	}
}

func TestSPL2SQLRecoveryPreservesOriginalDiagnostics(t *testing.T) {
	for _, q := range []string{`FROM main GROUP BY SELECT host,count() AS n`, `SELECT sum(size) AS total FROM main GROUP BY server HAVING`} {
		p := parseSPL2Document(q)
		r := spl2AnalyzeTest(t, q)
		for _, original := range p.diagnostics {
			found := false
			for _, actual := range r.Diagnostics {
				if actual.Code == original.Code && actual.Message == original.Message && actual.Location == original.Location {
					found = true
				}
			}
			if !found {
				t.Fatalf("original error removed: %+v", original)
			}
		}
	}
}
func TestSPL2SQLGroupRecoveryNearMisses(t *testing.T) {
	for _, q := range []string{`FROM main GROUP BY coalesce(SELECT host) SELECT count() AS n`, `FROM main GROUP BY [SELECT host] SELECT count() AS n`} {
		r := spl2AnalyzeTest(t, q)
		first := strings.Index(q, "SELECT")
		for _, stage := range r.Stages {
			if stage.Command == "select" && stage.Location.Start.Offset == first {
				t.Fatalf("nested SELECT promoted: %+v", r)
			}
		}
		if r.Status != Invalid {
			t.Fatalf("nested damage promoted: %+v", r)
		}
	}
	for _, key := range []string{`server+`, `lower(server)`, `server@`} {
		r := spl2AnalyzeTest(t, `SELECT sum(size) AS total FROM main GROUP BY `+key+` HAVING`)
		if r.Status != Invalid {
			t.Fatalf("key near miss promoted: %+v", r)
		}
		for _, stage := range r.Stages {
			if stage.Command == "group" && stage.SemanticComplete {
				t.Fatalf("unproved key promoted: %+v", r)
			}
		}
	}
}

func TestSPL2GroupKeyProofRequiresExactPredictionProvenance(t *testing.T) {
	query := `SELECT sum(size) AS total FROM main GROUP BY server HAVING`
	p := parseSPL2Document(query)
	command := spl2PipelineContexts(p.tree.Pipeline())[0].(spl2SQLCommand)
	proof := p.proveGroupKeyBeforeEmptyHaving(command)
	if proof == nil || proof.identifier.GetText() != "server" {
		t.Fatal("real single-key proof absent")
	}
	original := append([]Diagnostic{}, p.diagnostics...)
	p.diagnostics = append(p.diagnostics, Diagnostic{Code: CodeSyntaxError, Severity: "error", Category: "syntax", Message: "independent earlier key damage", Location: proof.diagnostic.Location})
	if p.proveGroupKeyBeforeEmptyHaving(command) != nil {
		t.Fatal("prior error was waived")
	}
	p.diagnostics = original
	p.predictionErrors = nil
	if p.proveGroupKeyBeforeEmptyHaving(command) != nil {
		t.Fatal("message/location alone authorized a waiver")
	}
	for _, key := range []string{`'server'`, `lower(server)`, `server+`, `server@`} {
		p := parseSPL2Document(`SELECT sum(size) AS total FROM main GROUP BY ` + key + ` HAVING`)
		command := spl2PipelineContexts(p.tree.Pipeline())[0].(spl2SQLCommand)
		if p.proveGroupKeyBeforeEmptyHaving(command) != nil {
			t.Fatalf("non-bare key gained special proof: %s", key)
		}
	}
	q := `FROM main | eval 'é'=1 | SELECT sum(size) AS total FROM main GROUP BY server HAVING`
	r := spl2AnalyzeTest(t, q)
	ref := spl2Ref(t, r, "server", "group")
	if ref.Binding != "source" || ref.Location.Start.Offset != strings.Index(q, "server") || r.Status != Invalid {
		t.Fatalf("nonzero Unicode ownership: %+v", r)
	}
	assertCorpusIntegrity(t, r)
}

func TestSPL2SQLRecoveredTypedInputsDoNotInstallOutputs(t *testing.T) {
	for _, tc := range []struct{ query, name string }{
		{`FROM main SELECT lower(account) AS ORDER BY normalized`, `account`},
		{`SELECT lower(user) AS FROM main`, `user`},
		{`SELECT count() FROM main GROUP BY _time span=()`, `_time`},
	} {
		t.Run(tc.query, func(t *testing.T) {
			r := spl2AnalyzeTest(t, tc.query)
			if r.Status != Invalid {
				t.Fatalf("damaged clause promoted: %+v", r)
			}
			count := 0
			for _, ref := range r.References {
				if ref.NormalizedName == tc.name && ref.Role == "read" {
					count++
				}
			}
			if count != 1 {
				t.Fatalf("want one original input read, got %d: %+v", count, r)
			}
			for _, ref := range r.References {
				if ref.Role == "create" && (ref.NormalizedName == "main" || ref.NormalizedName == "normalized") {
					t.Fatalf("damaged alias installed: %+v", r)
				}
			}
			assertCorpusIntegrity(t, r)
		})
	}
}

func TestSPL2SQLMissingSelectEOFReadsOriginalGroup(t *testing.T) {
	for _, query := range []string{
		`FROM charges GROUP BY region`,
		`FROM main GROUP BY host`,
		`FROM main | FROM 'café' GROUP BY host`,
	} {
		t.Run(query, func(t *testing.T) {
			r := spl2AnalyzeTest(t, query)
			name := "host"
			if strings.HasSuffix(query, "region") {
				name = "region"
			}
			if r.Status != Invalid || r.Coverage.SyntaxComplete || r.Coverage.SemanticComplete {
				t.Fatalf("missing SELECT promoted: %+v", r)
			}
			found := false
			for _, ref := range r.References {
				if ref.NormalizedName == name && ref.Role == "group" && ref.OriginalName == name {
					found = true
					if query[ref.Location.Start.Offset:ref.Location.End.Offset] != name {
						t.Fatalf("wrong original span: %+v", ref)
					}
				}
			}
			if !found {
				t.Fatalf("intact original GROUP input lost: %+v", r)
			}
			for _, stage := range r.Stages {
				if stage.Command == "select" {
					t.Fatalf("invented SELECT: %+v", r)
				}
			}
			assertCorpusIntegrity(t, r)
		})
	}
}

func TestSPL2SQLMissingSelectEOFAttributionGuards(t *testing.T) {
	for _, tc := range []struct {
		query   string
		allowed bool
	}{
		{`FROM main GROUP BY host`, true},
		{`FROM main GROUP BY host SELECT host`, false},
		{`FROM main GROUP BY host+`, false},
		{`FROM main GROUP BY span(_time,)`, false},
		{`FROM main GROUP BY @host`, false},
		{`FROM main GROUP BY host | where x=1`, false},
		{`FROM main GROUP BY host;`, false},
		{`FROM main WHERE EXISTS (FROM main GROUP BY host) SELECT host`, false},
		{`FROM main WHERE EXISTS (FROM main GROUP BY host`, false},
	} {
		t.Run(tc.query, func(t *testing.T) {
			p := parseSPL2Document(tc.query)
			var from *spl2.FromCommandContext
			var walk func(antlr.Tree)
			walk = func(tree antlr.Tree) {
				if f, ok := tree.(*spl2.FromCommandContext); ok && f.SqlGroupClause() != nil {
					from = f
				}
				for _, child := range tree.GetChildren() {
					walk(child)
				}
			}
			walk(p.tree)
			var diagnostic *Diagnostic
			if from != nil {
				diagnostic = p.groupMissingSelectDiagnostic(from)
			}
			if (diagnostic != nil) != tc.allowed {
				t.Fatalf("EOF attribution=%+v, want %v", diagnostic, tc.allowed)
			}
			if tc.allowed {
				before := append([]Diagnostic{}, p.diagnostics...)
				local := *p
				local.source = newSourceIndex(tc.query + " original tail")
				if local.groupMissingSelectDiagnostic(from) != nil {
					t.Fatal("substream EOF treated as original document end")
				}
				local = *p
				local.tokens = antlr.NewCommonTokenStream(spl2.NewSPL2Lexer(antlr.NewInputStream(tc.query)), antlr.TokenDefaultChannel)
				local.tokens.Fill()
				if local.groupMissingSelectDiagnostic(from) != nil {
					t.Fatal("non-original EOF identity admitted")
				}
				if !reflect.DeepEqual(before, p.diagnostics) {
					t.Fatal("raw diagnostics mutated")
				}
			}
		})
	}
}

func TestSPL2SQLDamagedSpanDoesNotInventSiblingKeys(t *testing.T) {
	r := spl2AnalyzeTest(t, `SELECT count() FROM main GROUP BY span(_time,hour,day)`)
	for _, ref := range r.References {
		if ref.Kind == "field" && (ref.NormalizedName == "hour" || ref.NormalizedName == "day" || ref.NormalizedName == "_time") {
			t.Fatalf("damaged SPAN tokens inferred as inputs: %+v", ref)
		}
	}
	if r.Status != Invalid {
		t.Fatalf("malformed SPAN promoted: %+v", r)
	}
}

func TestSPL2SQLIndependentCountProofSurvivesSiblingLimitation(t *testing.T) {
	for _, tc := range []struct{ query, name string }{
		{`SELECT region,count() AS n FROM main Group By region`, "n"},
		{`SELECT host,count() FROM main HAVING count>1 GROUP BY host`, "count"},
		{`FROM main GROUPBY lower(user) SELECT lower(user) AS normalized,count()`, "count"},
		{`FROM orders AS o LEFT OUTER JOIN inventory AS i ON o.item=i.id WHERE active=true GROUP BY o.item SELECT o.item,count() AS n HAVING EXISTS (SELECT id FROM stock WHERE id=o.item) ORDERBY n DESC LIMIT 4 OFFSET 1 | where n>0`, "n"},
	} {
		t.Run(tc.query, func(t *testing.T) {
			r := spl2AnalyzeTest(t, tc.query)
			if r.Status == Valid || r.Coverage.SemanticComplete {
				t.Fatalf("surrounding limitation promoted: %+v", r)
			}
			found := false
			for _, lineage := range r.Lineage {
				for _, tr := range lineage.Transitions {
					if tr.Operation == "aggregate" && tr.Output == tc.name {
						found = true
						if tr.Conditional {
							t.Fatalf("independent count proof contaminated: %+v", tr)
						}
					}
				}
			}
			if !found {
				t.Fatalf("count output missing: %+v", r)
			}
			assertCorpusIntegrity(t, r)
		})
	}
}

func TestSPL2SQLCountProofDoesNotOverrideLocalDamage(t *testing.T) {
	for _, query := range []string{
		`SELECT sum(bytes) AS n,other FROM main`,
		`SELECT mystery() AS n FROM main`,
		`SELECT count()+1 AS n FROM main`,
		`SELECT count(host) AS n,other FROM main`,
		`SELECT count(host,other) AS n FROM main`,
		`SELECT count() AS n,count() AS n FROM main`,
		`SELECT count() AS n FROM main WHERE n>0`,
		`SELECT count() AS n,1 AS 'field_${x}' FROM main`,
		`SELECT count() AS FROM main`,
		`FROM main GROUP BY SELECT count(,) AS n`,
	} {
		t.Run(query, func(t *testing.T) {
			r := spl2AnalyzeTest(t, query)
			for _, lineage := range r.Lineage {
				for _, tr := range lineage.Transitions {
					if tr.Output == "n" && (tr.Operation == "aggregate" || tr.Operation == "create") && !tr.Conditional {
						t.Fatalf("unproved output made definite: %+v", tr)
					}
				}
			}
			for _, field := range r.Lineage[len(r.Lineage)-1].After.Fields {
				if field.Name == "n" && !field.Conditional {
					t.Fatalf("unproved final presence: %+v", field)
				}
			}
		})
	}
	for _, query := range []string{`SELECT count() FROM main GROUP BY _time span=()`, `SELECT count() FROM main GROUP BY timestamp span=()`} {
		r := spl2AnalyzeTest(t, query)
		name := "_time"
		if strings.Contains(query, "timestamp") {
			name = "timestamp"
		}
		if r.Status != Invalid || spl2Ref(t, r, name, "read").OriginalName != name {
			t.Fatalf("intact assignment-form field lost: %+v", r)
		}
	}
}

func TestSPL2SQLCountCollisionIncludesUninstalledProjection(t *testing.T) {
	for _, query := range []string{
		`SELECT o.n,count() AS n FROM main AS o`,
		`SELECT n,count() AS n FROM main`,
		`FROM main GROUPBY lower(user) SELECT n,count() AS n`,
		`SELECT o[field],count() AS n FROM main AS o`,
		`SELECT o.n.deep,count() AS n FROM main AS o`,
	} {
		t.Run(query, func(t *testing.T) {
			r := spl2AnalyzeTest(t, query)
			for _, lineage := range r.Lineage {
				for _, tr := range lineage.Transitions {
					if tr.Operation == "aggregate" && tr.Output == "n" && !tr.Conditional {
						t.Fatalf("uninstalled competing projection ignored: %+v", tr)
					}
				}
			}
		})
	}
}

func TestSPL2SQLFiniteProjectionMembershipPreservesConditionality(t *testing.T) {
	for _, tc := range []struct{ query, field string }{
		{`FROM main GROUPBY lower(user) SELECT lower(user) AS normalized,count()`, "normalized"},
		{`FROM main SELECT mystery(host) AS output`, "output"},
	} {
		t.Run(tc.query, func(t *testing.T) {
			r := spl2AnalyzeTest(t, tc.query)
			last := r.Lineage[len(r.Lineage)-1]
			if r.Coverage.SemanticComplete || r.Status != Incomplete || last.Phase != "project" || last.After.Open || last.After.Uncertain {
				t.Fatalf("fixed projected membership lost: %+v", r)
			}
			found := false
			for _, f := range last.After.Fields {
				if f.Name == tc.field {
					found = true
					if !f.Conditional {
						t.Fatalf("conditional output promoted: %+v", f)
					}
				}
			}
			if !found {
				t.Fatalf("missing prepared output: %+v", last)
			}
			if tc.field == "normalized" && (len(last.After.Fields) != 2 || last.After.Fields[0].Name != "count" || last.After.Fields[0].Conditional) {
				t.Fatalf("exact selected set drifted: %+v", last.After)
			}
			assertCorpusIntegrity(t, r)
		})
	}
	for _, query := range []string{
		`FROM main SELECT lower(host)`,
		`SELECT count() AS 'out_${host}' FROM main`,
		`SELECT count() AS n,count() AS n FROM main`,
		`SELECT lower(host) AS FROM main`,
		`SELECT o.item,count() AS n FROM main AS o`,
	} {
		r := spl2AnalyzeTest(t, query)
		if r.Coverage.SemanticComplete || !r.Lineage[len(r.Lineage)-1].After.Uncertain {
			t.Fatalf("unproved output universe closed: %s %+v", query, r)
		}
	}
}

func TestSPL2SQLRecoveredCountKeepsOriginalGroupDamage(t *testing.T) {
	queries := []string{
		`FROM main GROUP BY SELECT count()`,
		`FROM main | FROM 'café' GROUP BY SELECT count()`,
	}
	for _, group := range []string{"GROUP BY", "GROUPBY", "group by", "groupby"} {
		queries = append(queries, "FROM main "+group+" SELECT host,count()", "FROM main "+group+" host, SELECT host,count()")
	}
	for _, query := range queries {
		t.Run(query, func(t *testing.T) {
			r := spl2AnalyzeTest(t, query)
			if r.Status != Invalid || r.Coverage.SyntaxComplete || r.Coverage.SemanticComplete {
				t.Fatalf("recovered SELECT erased original invalidity: %+v", r)
			}
			name := "count"
			if strings.HasSuffix(query, "AS n") {
				name = "n"
			}
			found := false
			for _, lineage := range r.Lineage {
				for _, tr := range lineage.Transitions {
					if tr.Operation == "aggregate" && tr.Output == name {
						found = true
						if tr.Conditional {
							t.Fatalf("pre-retry missing SELECT tainted intact recovered count: %+v", tr)
						}
					}
				}
			}
			if !found {
				t.Fatalf("intact count output absent: %+v", r)
			}
			originalEOF, groupDamage := false, false
			for _, d := range r.Diagnostics {
				originalEOF = originalEOF || (d.Code == CodeSyntaxError && d.Location.Start.Offset == len(query) && d.Location.End.Offset == len(query))
				groupDamage = groupDamage || (d.Code == CodeSyntaxError && d.Location.Start.Offset == strings.Index(query, "SELECT"))
			}
			if !originalEOF || !groupDamage {
				t.Fatalf("original EOF/GROUP diagnostics were removed: %+v", r.Diagnostics)
			}
			for _, stage := range r.Stages {
				if stage.Command == "group" && stage.SemanticComplete {
					t.Fatal("damaged GROUP effects promoted")
				}
			}
			assertCorpusIntegrity(t, r)
		})
	}
}

func TestSPL2SQLRecoveredCountStillRequiresSoundDestination(t *testing.T) {
	for _, query := range []string{
		`FROM main GROUP BY SELECT count() AS n`,
		`FROM main | FROM 'café' GROUP BY SELECT count() AS n`,
		`FROM main GROUP BY SELECT count(,) AS n`,
		`FROM main GROUP BY SELECT count(host) AS n`,
		`FROM main GROUP BY SELECT count()+1 AS n`,
		`FROM main GROUP BY SELECT count() AS`,
		`FROM main GROUP BY SELECT count() AS 'n_${host}'`,
		`FROM main GROUP BY SELECT count() AS n,count() AS n`,
		`FROM main GROUP BY SELECT n,count() AS n`,
		`FROM main GROUP BY SELECT mystery(host),count() AS n`,
		`FROM main GROUP BY SELECT count() AS n@`,
		`FROM main WHERE EXISTS (FROM main GROUP BY SELECT count() AS n`,
	} {
		t.Run(query, func(t *testing.T) {
			r := spl2AnalyzeTest(t, query)
			if r.Status != Invalid {
				t.Fatalf("damaged SQL promoted: %+v", r)
			}
			for _, lineage := range r.Lineage {
				for _, tr := range lineage.Transitions {
					if tr.Output == "n" && (tr.Operation == "aggregate" || tr.Operation == "create") && !tr.Conditional {
						t.Fatalf("unproved recovered output made definite: %+v", tr)
					}
				}
			}
		})
	}
}

func TestSPL2SQLRecoveredCountRequiresPairedOriginalEOF(t *testing.T) {
	query := `FROM main GROUP BY SELECT count()`
	p := parseSPL2Document(query)
	result := &Result{Diagnostics: append([]Diagnostic{}, p.diagnostics...)}
	sites := spl2RecoverySites(p, result)
	command := sites[0].context.(spl2SQLCommand)
	projection := command.SqlSelectClause().Projection(0)
	stage := &spl2SemanticStage{semanticStage: &semanticStage{result: result}, parsed2: p}
	before := append([]Diagnostic{}, p.diagnostics...)
	if !stage.sqlProjectionEffectSound(projection) {
		t.Fatal("recorded original EOF did not admit intact count")
	}
	for _, change := range []struct {
		name  string
		apply func(*spl2ParsedDocument)
	}{
		{"no exception provenance", func(q *spl2ParsedDocument) { q.missingSelectEOF = nil }},
		{"no paired retry", func(q *spl2ParsedDocument) { q.recoveredSelects = nil }},
		{"not original document end", func(q *spl2ParsedDocument) { q.source = newSourceIndex(query + " tail") }},
		{"different original tokens", func(q *spl2ParsedDocument) {
			q.tokens = antlr.NewCommonTokenStream(spl2.NewSPL2Lexer(antlr.NewInputStream(query)), antlr.TokenDefaultChannel)
			q.tokens.Fill()
		}},
		{"independent overlapping syntax error", func(q *spl2ParsedDocument) {
			q.diagnostics = append(append([]Diagnostic{}, q.diagnostics...), Diagnostic{Code: CodeSyntaxError, Severity: "error", Category: "syntax", Message: "separate projection damage", Location: p.source.contextLocation(projection)})
		}},
	} {
		t.Run(change.name, func(t *testing.T) {
			local := *p
			change.apply(&local)
			stage.parsed2 = &local
			if stage.sqlProjectionEffectSound(projection) {
				t.Fatal("unproved recovered count admitted")
			}
		})
	}
	stage.parsed2 = p
	result.Diagnostics = append(result.Diagnostics, Diagnostic{Code: CodeSyntaxError, Severity: "error", Category: "contract", Message: "independent projection contract", Location: p.source.contextLocation(projection)})
	if stage.sqlProjectionEffectSound(projection) {
		t.Fatal("contract error waived with EOF")
	}
	if !reflect.DeepEqual(before, p.diagnostics) {
		t.Fatal("original diagnostics changed")
	}
	for _, query := range []string{
		`FROM main GROUP BY SELECT count() | table count`,
		`FROM main GROUP BY SELECT count();`,
		`FROM main WHERE EXISTS (FROM main GROUP BY SELECT count()) SELECT count()`,
		`FROM main WHERE EXISTS (FROM main GROUP BY SELECT count()`,
	} {
		p := parseSPL2Document(query)
		result := &Result{Diagnostics: append([]Diagnostic{}, p.diagnostics...)}
		for _, site := range spl2RecoverySites(p, result) {
			command, ok := site.context.(spl2SQLCommand)
			if !ok || command.SqlSelectClause() == nil {
				continue
			}
			for _, projection := range command.SqlSelectClause().AllProjection() {
				if _, diagnostic := p.recoveredCountView(projection); diagnostic != nil {
					t.Fatalf("neighbor/nested boundary gained EOF waiver: %s", query)
				}
			}
		}
	}
}
