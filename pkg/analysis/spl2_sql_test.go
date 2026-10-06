package analysis

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/antlr4-go/antlr/v4"
	"github.com/delgado-jacob/spl-toolkit/parser/spl2"
)

func testAtomicIdentity(name string) FieldIdentity {
	return FieldIdentity{Kind: "atomic", Segments: []string{name}}
}

func testAtomicIdentityPointer(name string) *FieldIdentity {
	identity := testAtomicIdentity(name)
	return &identity
}

func TestSPL2SQLSelectedLinusForms(t *testing.T) {
	tests := []struct {
		name  string
		query string
	}{
		{"source where", `SELECT synthetic_value FROM synthetic_events WHERE synthetic_enabled=true`},
		{"projection", `SELECT synthetic_value, synthetic_region FROM synthetic_events`},
		{"group aggregate", `SELECT synthetic_region, sum(synthetic_value) AS synthetic_total FROM synthetic_events GROUP BY synthetic_region`},
		{"aggregate alias", `SELECT max(synthetic_value) AS synthetic_maximum FROM synthetic_events`},
		{"private aggregate policy", `SELECT stdev(synthetic_value) AS synthetic_stdev FROM synthetic_events`},
		{"dotted read", `SELECT synthetic_object.synthetic_value AS synthetic_leaf FROM synthetic_events WHERE synthetic_object.synthetic_enabled=true`},
		{"multiline", "SELECT synthetic_region, count() AS synthetic_count\nFROM synthetic_events\nWHERE synthetic_value > 2\nGROUP BY synthetic_region"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := spl2AnalyzeTest(t, tc.query)
			if r.Status != Valid || !r.Coverage.SyntaxComplete || !r.Coverage.SemanticComplete || !r.Requirements.Coverage.Complete || len(r.Diagnostics) != 0 {
				t.Fatalf("selected SQL form must be complete: %+v", r)
			}
			assertCorpusIntegrity(t, r)
		})
	}

	r := spl2AnalyzeTest(t, `SELECT synthetic_region, sum(synthetic_value) AS synthetic_total FROM synthetic_events WHERE synthetic_enabled=true GROUP BY synthetic_region`)
	phases := []string{}
	for _, lineage := range r.Lineage {
		phases = append(phases, lineage.Phase)
	}
	if !reflect.DeepEqual(phases, []string{"source", "filter", "group", "aggregate", "project"}) {
		t.Fatalf("SQL phase schedule = %v", phases)
	}
	alias := spl2Ref(t, r, "synthetic_total", "output")
	input := spl2Ref(t, r, "synthetic_value", "read")
	if !reflect.DeepEqual(alias.OriginReferenceIDs, []string{input.ID}) {
		t.Fatalf("aggregate alias origin = %v, want %s", alias.OriginReferenceIDs, input.ID)
	}
	for _, name := range []string{"synthetic_enabled", "synthetic_value", "synthetic_region"} {
		role := "read"
		if name == "synthetic_region" {
			role = "group"
		}
		if item := requirementItem(r.Requirements, "field", name, role); item == nil || item.Necessity != "required" || item.Resolution != "exact" {
			t.Fatalf("SQL requirement %s/%s = %+v", name, role, item)
		}
	}

	dotted := spl2AnalyzeTest(t, `SELECT synthetic_object.synthetic_value AS synthetic_leaf FROM synthetic_events WHERE synthetic_object.synthetic_enabled=true`)
	for _, name := range []string{"synthetic_object.synthetic_value", "synthetic_object.synthetic_enabled"} {
		ref := spl2Ref(t, dotted, name, "read")
		if ref.Resolution != "exact" || ref.Binding != "source" {
			t.Fatalf("dotted SQL read %q = %+v", name, ref)
		}
		if item := requirementItem(dotted.Requirements, "field", name, "read"); item == nil || item.Resolution != "exact" {
			t.Fatalf("dotted SQL requirement %q = %+v", name, item)
		}
	}
}

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
	open := FieldState{Fields: []FieldBinding{}, Removed: []FieldRemoval{}, Open: true}
	filtered := FieldState{Fields: []FieldBinding{{Name: "bytes", FieldIdentity: testAtomicIdentity("bytes"), OriginReferenceIDs: []string{"ref-2"}}}, Removed: []FieldRemoval{}, Open: true}
	evaluated := FieldState{Fields: []FieldBinding{{Name: "bytes", FieldIdentity: testAtomicIdentity("bytes"), OriginReferenceIDs: []string{"ref-2"}}, {Name: "host", FieldIdentity: testAtomicIdentity("host"), OriginReferenceIDs: []string{"ref-0"}}}, Removed: []FieldRemoval{}, Open: true}
	projected := FieldState{Fields: []FieldBinding{{Name: "host", FieldIdentity: testAtomicIdentity("host"), OriginReferenceIDs: []string{"ref-0"}}}, Removed: []FieldRemoval{}}
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
	if !reflect.DeepEqual(r.Lineage[3].Transitions, []Transition{{Operation: "project", Output: "host", OutputIdentity: testAtomicIdentityPointer("host"), InputReferenceIDs: []string{"ref-0"}}}) {
		t.Fatalf("project reused read: %+v", r.Lineage[3])
	}
	want := newResult(QueryDocument{Text: text, Language: "spl2", Profile: "splunkd", Version: "current"})
	want.Status = Valid
	want.Stages = wantStages
	want.Scopes = []Scope{{ID: "scope-0", Kind: "root", Location: loc(0, 35)}}
	want.Dependencies.Datasets = []string{"main"}
	want.References = []Reference{
		{ID: "ref-0", OriginalName: "host", NormalizedName: "host", FieldIdentity: testAtomicIdentityPointer("host"), Kind: "field", Role: "read", StageID: "stage-0", ScopeID: "scope-0", Location: loc(7, 11), Resolution: "exact", Binding: "source", OriginReferenceIDs: []string{}},
		{ID: "ref-1", OriginalName: "main", NormalizedName: "main", Kind: "dataset", Role: "read", StageID: "stage-1", ScopeID: "scope-0", Location: loc(17, 21), Resolution: "exact", Binding: "not_applicable", OriginReferenceIDs: []string{}},
		{ID: "ref-2", OriginalName: "bytes", NormalizedName: "bytes", FieldIdentity: testAtomicIdentityPointer("bytes"), Kind: "field", Role: "read", StageID: "stage-2", ScopeID: "scope-0", Location: loc(28, 33), Resolution: "exact", Binding: "source", OriginReferenceIDs: []string{}},
	}
	orders := []int{0, 1, 2, 3}
	want.Lineage = []Lineage{
		{StageID: "stage-1", ScopeID: "scope-0", Before: open, After: open, Transitions: []Transition{}, Phase: "source", ExecutionOrder: &orders[0]},
		{StageID: "stage-2", ScopeID: "scope-0", Before: open, After: filtered, Transitions: []Transition{}, Phase: "filter", ExecutionOrder: &orders[1]},
		{StageID: "stage-0", ScopeID: "scope-0", Before: filtered, After: evaluated, Transitions: []Transition{}, Phase: "evaluate", ExecutionOrder: &orders[2]},
		{StageID: "stage-0", ScopeID: "scope-0", Before: evaluated, After: projected, Transitions: []Transition{{Operation: "project", Output: "host", OutputIdentity: testAtomicIdentityPointer("host"), InputReferenceIDs: []string{"ref-0"}}}, Phase: "project", ExecutionOrder: &orders[3]},
	}
	want.Requirements = RequirementSet{
		SchemaVersion: 1,
		Query: RequirementQueryIdentity{
			Language: "spl2", Profile: "splunkd", Version: "current",
			QueryDigest: "sha256:aeacf92af76b0a8b74aacd3863908ec66d33802ee4cf5ea0ffa80f9c9d78129f",
		},
		CapabilityRevision: "sha256:7134e06d345f6b2c6e58c3d29c727868320b47ff1fc0a94842aec35615223d9f",
		QueryStatus:        Valid,
		Coverage:           RequirementCoverage{Complete: true, Reasons: []string{}},
		Items: []RequirementItem{
			{ID: "req-1", Kind: "field", Identity: "host", FieldIdentity: testAtomicIdentityPointer("host"), Role: "read", Necessity: "required", Origin: "direct", Resolution: "exact", Occurrences: []RequirementOccurrence{{ReferenceID: "ref-0", OriginalName: "host", Binding: "source", StageID: "stage-0", ScopeID: "scope-0", Location: loc(7, 11)}}},
			{ID: "req-2", Kind: "dataset", Identity: "main", Role: "read", Necessity: "required", Origin: "direct", Resolution: "exact", Occurrences: []RequirementOccurrence{{ReferenceID: "ref-1", OriginalName: "main", Binding: "not_applicable", StageID: "stage-1", ScopeID: "scope-0", Location: loc(17, 21)}}},
			{ID: "req-3", Kind: "field", Identity: "bytes", FieldIdentity: testAtomicIdentityPointer("bytes"), Role: "read", Necessity: "required", Origin: "direct", Resolution: "exact", Occurrences: []RequirementOccurrence{{ReferenceID: "ref-2", OriginalName: "bytes", Binding: "source", StageID: "stage-2", ScopeID: "scope-0", Location: loc(28, 33)}}},
		},
		Gaps:        []RequirementGap{},
		Diagnostics: []Diagnostic{},
	}
	addExactInputTestEvidence(want, want.References[1], want.References[2], want.References[0])
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
	owner := FieldBinding{Name: "owner", FieldIdentity: testAtomicIdentity("owner"), OriginReferenceIDs: []string{"ref-1", "ref-0"}, Conditional: true}
	if !reflect.DeepEqual(r.Lineage[1].After.Fields, []FieldBinding{owner, {Name: "user", FieldIdentity: testAtomicIdentity("user"), OriginReferenceIDs: []string{"ref-0"}}}) || !reflect.DeepEqual(r.Lineage[3].After, FieldState{Fields: []FieldBinding{owner}, Removed: []FieldRemoval{}}) {
		t.Fatalf("states %+v", r.Lineage)
	}
	if !reflect.DeepEqual(r.Lineage[3].Transitions, []Transition{{Operation: "project", Output: "owner", OutputIdentity: testAtomicIdentityPointer("owner"), InputReferenceIDs: []string{"ref-1"}}}) {
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
			field := FieldBinding{Name: "n", FieldIdentity: testAtomicIdentity("n"), OriginReferenceIDs: []string{createID}}
			closed := FieldState{Fields: []FieldBinding{field}, Removed: []FieldRemoval{}}
			for _, phase := range r.Lineage[1:] {
				if !reflect.DeepEqual(phase.After, closed) {
					t.Fatalf("aggregate presence/projection: %+v", phase)
				}
			}
			if !reflect.DeepEqual(r.Lineage[1].Transitions, []Transition{{Operation: "aggregate", Output: "n", OutputIdentity: testAtomicIdentityPointer("n"), InputReferenceIDs: []string{}, OutputReferenceID: createID}}) || !reflect.DeepEqual(r.Lineage[3].Transitions, []Transition{{Operation: "project", Output: "n", OutputIdentity: testAtomicIdentityPointer("n"), InputReferenceIDs: []string{createID}}}) {
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
	host := FieldBinding{Name: "host", FieldIdentity: testAtomicIdentity("host"), OriginReferenceIDs: []string{"ref-5"}}
	total := FieldBinding{Name: "total", FieldIdentity: testAtomicIdentity("total"), OriginReferenceIDs: []string{"ref-2", "ref-1"}, Conditional: true}
	if !reflect.DeepEqual(r.Lineage[2].After, FieldState{Fields: []FieldBinding{host}, Removed: []FieldRemoval{}}) || !reflect.DeepEqual(r.Lineage[3].Before, r.Lineage[2].After) || !reflect.DeepEqual(r.Lineage[3].After, FieldState{Fields: []FieldBinding{host, total}, Removed: []FieldRemoval{}}) {
		t.Fatalf("group/aggregate states %+v", r.Lineage)
	}
	if !reflect.DeepEqual(r.Lineage[3].Transitions, []Transition{{Operation: "aggregate", Output: "total", OutputIdentity: testAtomicIdentityPointer("total"), InputReferenceIDs: []string{"ref-1"}, OutputReferenceID: "ref-2", Conditional: true}}) || !reflect.DeepEqual(r.Lineage[6].Transitions, []Transition{{Operation: "project", Output: "host", OutputIdentity: testAtomicIdentityPointer("host"), InputReferenceIDs: []string{"ref-0"}}, {Operation: "project", Output: "total", OutputIdentity: testAtomicIdentityPointer("total"), InputReferenceIDs: []string{"ref-2"}}}) {
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
				if len(state.Fields) != 0 || !reflect.DeepEqual(state.Removed, []FieldRemoval{{Name: c.name, FieldIdentity: FieldIdentity{Kind: "atomic", Segments: []string{c.name}}}}) {
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
			want := FieldState{Fields: []FieldBinding{{Name: "host", FieldIdentity: testAtomicIdentity("host"), OriginReferenceIDs: []string{c.hostID}}}, Removed: []FieldRemoval{}}
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
	wantState := FieldState{Fields: []FieldBinding{{Name: "actor.local", FieldIdentity: testAtomicIdentity("actor.local"), OriginReferenceIDs: []string{"ref-0"}}, {Name: "host", FieldIdentity: testAtomicIdentity("host"), OriginReferenceIDs: []string{"ref-1"}}}, Removed: []FieldRemoval{}}
	for _, u := range []SourceUniverse{{Fields: []string{"actor.name", "host"}, Complete: true}, {Fields: []string{"actor.name", "host"}}, {Fields: []string{"actor.name", "host"}, Complete: true, Resolve: func(string) SourceFieldAdmission { return SourceFieldAdmitted }}} {
		r, err := AnalyzeWithSourceUniverse(doc, u)
		if err != nil || r.Result.Status != Valid || !reflect.DeepEqual(r.Result.Lineage[3].After, wantState) {
			t.Fatalf("state %+v %v", r.Result.Lineage, err)
		}
		want := []FieldExpansion{{ReferenceID: "ref-3", Complete: true, Matches: []ExpandedField{{Name: "actor.local", Binding: "derived"}, {Name: "host", Binding: "source"}}}}
		if !reflect.DeepEqual(r.Expansions, want) {
			t.Fatalf("expansions %+v", r.Expansions)
		}
		if !reflect.DeepEqual(r.Result.Lineage[2].Transitions, []Transition{{Operation: "project", Output: "actor.local", OutputIdentity: testAtomicIdentityPointer("actor.local"), InputReferenceIDs: []string{"ref-0"}}, {Operation: "project", Output: "host", OutputIdentity: testAtomicIdentityPointer("host"), InputReferenceIDs: []string{"ref-1"}}}) {
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
	if !reflect.DeepEqual(r.Expansions, want) || !reflect.DeepEqual(r.Result.Lineage[2].After, FieldState{Fields: []FieldBinding{{Name: "n", FieldIdentity: testAtomicIdentity("n"), OriginReferenceIDs: []string{"ref-1", "ref-0"}, Conditional: true}}, Removed: []FieldRemoval{}, Uncertain: true}) {
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

func TestSPL2SQLSelectedJoinChains(t *testing.T) {
	for _, query := range []string{
		`SELECT a.host AS first_host, b.owner AS second_owner, c.region AS third_region FROM alpha AS a JOIN beta AS b ON b.beta_key=a.alpha_key INNER JOIN gamma AS c ON b.next_key=c.gamma_key WHERE b.enabled=true`,
		`FROM alpha AS a JOIN beta AS b ON a.alpha_key=b.beta_key JOIN gamma AS c ON c.gamma_key=b.next_key WHERE b.enabled=true SELECT a.host AS first_host, b.owner AS second_owner, c.region AS third_region`,
	} {
		t.Run(query, func(t *testing.T) {
			r := spl2AnalyzeTest(t, query)
			if r.Status != Valid || !r.Coverage.SemanticComplete || r.Correlation.Outcome != "connected" || len(r.Correlation.Nodes) != 3 || len(r.Correlation.Edges) != 2 {
				t.Fatalf("selected SQL chain: status=%s coverage=%+v correlation=%+v diagnostics=%+v", r.Status, r.Coverage, r.Correlation, r.Diagnostics)
			}
			for name, inputName := range map[string]string{"host": "alpha", "owner": "beta", "region": "gamma", "enabled": "beta"} {
				item := requirementItem(r.Requirements, "field", name, "read")
				if item == nil || item.Ownership.State != "proved" || item.Necessity != "required" {
					t.Fatalf("%s requirement: %+v", name, item)
				}
				found := false
				for _, input := range r.Inputs {
					if input.ID == item.InputID && input.Name == inputName {
						found = true
					}
				}
				if !found {
					t.Fatalf("%s owner: %+v inputs=%+v", name, item, r.Inputs)
				}
			}
			assertCorpusIntegrity(t, r)
		})
	}
}

func TestSPL2SQLLeftJoinConditionalReads(t *testing.T) {
	for _, join := range []string{"LEFT JOIN", "LEFT OUTER JOIN"} {
		r := spl2AnalyzeTest(t, `SELECT b.owner AS right_owner, a.host AS left_host FROM alpha AS a `+join+` beta AS b ON a.alpha_key=b.beta_key WHERE b.enabled=true`)
		if r.Status != Valid || r.Correlation.Outcome != "connected" {
			t.Fatalf("left join: %+v", r)
		}
		for _, name := range []string{"owner", "enabled"} {
			item := requirementItem(r.Requirements, "field", name, "read")
			if item == nil || item.Ownership.State != "proved" || item.Necessity != "conditional" {
				t.Fatalf("right read %s: %+v", name, item)
			}
		}
		if item := requirementItem(r.Requirements, "field", "host", "read"); item == nil || item.Necessity != "required" {
			t.Fatalf("left read: %+v", item)
		}
	}
}

func TestSPL2SQLJoinGroupingAndVisibility(t *testing.T) {
	r := spl2AnalyzeTest(t, `SELECT a.host, sum(b.bytes) AS total FROM alpha AS a JOIN beta AS b ON a.alpha_key=b.beta_key GROUP BY a.host HAVING a.host="x" ORDER BY total`)
	if r.Status != Valid || !r.Coverage.SemanticComplete || r.Correlation.Outcome != "connected" {
		t.Fatalf("grouped join: diagnostics=%+v refs=%+v", r.Diagnostics, r.References)
	}
	for _, check := range []struct{ name, role string }{{"host", "group"}, {"bytes", "read"}} {
		if item := requirementItem(r.Requirements, "field", check.name, check.role); item == nil || item.Ownership.State != "proved" {
			t.Fatalf("group requirement: %+v", item)
		}
	}
	held := spl2AnalyzeTest(t, `SELECT a.host, sum(b.bytes) AS total FROM alpha AS a JOIN beta AS b ON a.alpha_key=b.beta_key GROUP BY a.host HAVING b.owner="x"`)
	if held.Coverage.SemanticComplete {
		t.Fatalf("ungrouped qualified read admitted: %+v", held)
	}
}

func TestSPL2SQLJoinHeldOwnershipBoundaries(t *testing.T) {
	for _, query := range []string{
		`SELECT a.host FROM alpha AS a JOIN beta AS a ON a.key=a.key`,
		`SELECT a.host FROM alpha AS a JOIN beta AS b ON a.key=a.key`,
		`SELECT a.host FROM alpha AS a JOIN beta AS b ON c.key=b.key`,
		`SELECT a.host FROM alpha AS a JOIN beta AS b ON a.key.child=b.key`,
	} {
		r := spl2AnalyzeTest(t, query)
		if r.Coverage.SemanticComplete || r.Correlation.Outcome == "connected" || len(r.Inputs) != 2 {
			t.Fatalf("held join: %s %+v", query, r)
		}
	}
	r := spl2AnalyzeTest(t, `SELECT host FROM alpha AS a JOIN beta AS b ON a.key=b.key`)
	item := requirementItem(r.Requirements, "field", "host", "read")
	if item == nil || item.Ownership.State != "unproved" || len(item.Ownership.CandidateInputIDs) != 2 {
		t.Fatalf("unqualified join read: %+v", item)
	}
	self := spl2AnalyzeTest(t, `SELECT a.host AS first_host, b.host AS second_host FROM alpha AS a JOIN alpha AS b ON a.key=b.key`)
	if self.Correlation.Outcome != "connected" || len(self.Inputs) != 1 || len(self.Inputs[0].Occurrences) != 2 {
		t.Fatalf("self join: %+v", self)
	}
}

func TestSPL2SQLJoinQualifiedProjectionDisambiguatesSourceNames(t *testing.T) {
	query := `FROM $events AS e JOIN $users AS u ON e.user_id=u.id JOIN $assets AS a ON e.asset_id=a.id SELECT e.user_id AS user_id, u.id AS user_record, a.id AS asset_record`
	r := spl2AnalyzeTest(t, query)
	if r.Status != Valid || !r.Requirements.Coverage.Complete || r.FieldAttributionCoverage.State != "complete" || r.Correlation.Outcome != "connected" {
		t.Fatalf("qualified labels retain independent ownership: coverage=%+v requirements=%+v diagnostics=%+v", r.Coverage, r.Requirements, r.Diagnostics)
	}
	// Both id obligations remain independent despite the matched-row name clash.
	ids := map[string]bool{}
	for _, item := range r.Requirements.Items {
		if item.Kind == "field" && item.Identity == "id" {
			ids[item.InputID] = true
		}
	}
	if len(ids) != 2 {
		t.Fatalf("same-name source obligations coalesced: %+v", r.Requirements.Items)
	}
}

func TestSPL2SQLJoinOutputCollisionRetainsCandidateOwners(t *testing.T) {
	r := spl2AnalyzeTest(t, `SELECT a.host,b.host FROM alpha AS a JOIN beta AS b ON a.alpha_key=b.beta_key | where host="x"`)
	if r.Coverage.SemanticComplete || r.Correlation.Outcome != "connected" {
		t.Fatalf("output collision: coverage=%+v graph=%+v", r.Coverage, r.Correlation)
	}
	for _, item := range r.Requirements.Items {
		for _, occurrence := range item.Occurrences {
			if occurrence.OriginalName == "host" {
				if item.InputID != "" || item.Ownership.State != "unproved" || len(item.Ownership.CandidateInputIDs) != 2 || occurrence.Necessity != "conditional" {
					t.Fatalf("collision fabricated ownership: %+v", item)
				}
				return
			}
		}
	}
	t.Fatal("postprojection collision read missing")
}

func TestSPL2SQLJoinANDAndLeftChain(t *testing.T) {
	r := spl2AnalyzeTest(t, `FROM alpha AS a LEFT JOIN beta AS b ON a.alpha_key=b.beta_key AND b.name=a.name JOIN gamma AS c ON b.next_key=c.gamma_key SELECT a.host AS left_host,b.owner AS right_owner,c.region AS last_region`)
	if r.Status != Valid || r.Correlation.Outcome != "connected" || len(r.Correlation.Edges) != 2 || len(r.Correlation.Edges[0].Keys) != 2 {
		t.Fatalf("AND/left chain: diagnostics=%+v graph=%+v", r.Diagnostics, r.Correlation)
	}
	item := requirementItem(r.Requirements, "field", "next_key", "read")
	if item == nil || item.Necessity != "conditional" || item.Ownership.State != "proved" {
		t.Fatalf("prior left side conditional key: %+v", item)
	}
}

func TestSPL2SQLLeftJoinClosedQualifiedReadsRemainUnavailable(t *testing.T) {
	for _, query := range []string{
		`SELECT a.host AS selected_host FROM alpha AS a LEFT JOIN beta AS b ON a.id=b.uid | eval later=b.secret`,
		`SELECT a.host,count() AS n FROM alpha AS a LEFT JOIN beta AS b ON a.id=b.uid GROUP BY a.host HAVING b.secret="x"`,
	} {
		r := spl2AnalyzeTest(t, query)
		ref := spl2Ref(t, r, "secret", "read")
		if ref.Binding != "unavailable" {
			t.Fatalf("closed LEFT field restored: %+v", ref)
		}
		if item := requirementItem(r.Requirements, "field", "secret", "read"); item != nil {
			t.Fatalf("unavailable field became external obligation: %+v", item)
		}
	}
}

func TestSPL2SQLJoinUnqualifiedKeyDoesNotSelectLastSource(t *testing.T) {
	r := spl2AnalyzeTest(t, `FROM alpha AS a JOIN beta AS b ON a.id=b.id | where id>0`)
	item := requirementItem(r.Requirements, "field", "id", "read")
	// The unqualified read is separate from the two proved qualified obligations.
	for _, candidate := range r.Requirements.Items {
		for _, occurrence := range candidate.Occurrences {
			if occurrence.OriginalName == "id" {
				item = &candidate
			}
		}
	}
	if item == nil || item.InputID != "" || item.Ownership.State != "unproved" || len(item.Ownership.CandidateInputIDs) != 2 {
		t.Fatalf("unqualified key selected a source: %+v", item)
	}
}

func TestSPL2SQLLeftJoinStaticStructuralReadsKeepTheirSource(t *testing.T) {
	r := spl2AnalyzeTest(t, `SELECT e.payload.id AS event_id,u.payload.id AS user_id FROM events AS e LEFT JOIN users AS u ON e.user_key=u.id`)
	if r.Status != Valid || r.FieldAttributionCoverage.State != "complete" {
		t.Fatalf("static source paths: diagnostics=%+v coverage=%+v", r.Diagnostics, r.FieldAttributionCoverage)
	}
	seen := map[string]bool{}
	for _, item := range r.Requirements.Items {
		if item.Kind != "field" || item.Identity != "payload.id" {
			continue
		}
		if item.FieldIdentity == nil || !reflect.DeepEqual(*item.FieldIdentity, FieldIdentity{Kind: "path", Segments: []string{"payload", "id"}}) || item.Ownership.State != "proved" {
			t.Fatalf("source-relative structural identity: %+v", item)
		}
		for _, input := range r.Inputs {
			if input.ID == item.InputID {
				want := "required"
				if input.Name == "users" {
					want = "conditional"
				}
				if item.Necessity != want {
					t.Fatalf("structural source conditionality: %+v", item)
				}
				seen[input.Name] = true
			}
		}
	}
	if len(seen) != 2 {
		t.Fatalf("structural source paths coalesced or disappeared: %+v", r.Requirements.Items)
	}
	for _, ref := range r.References {
		if ref.OriginalName == "u.payload.id" || ref.OriginalName == "e.payload.id" {
			wantQualifier := "e"
			if ref.OriginalName == "u.payload.id" {
				wantQualifier = "u"
			}
			if ref.FieldIdentity == nil || ref.FieldIdentity.Qualifier != wantQualifier || !reflect.DeepEqual(ref.FieldIdentity.Segments, []string{"payload", "id"}) {
				t.Fatalf("original qualified path lost: %+v", ref)
			}
		}
	}
}

func TestSPL2SQLJoinConditionalAliasesPreserveDerivedFields(t *testing.T) {
	for _, tc := range []struct {
		name, query, outcome string
		inputs, edges        int
	}{
		{"generated rows", `FROM [{id:1}] AS a %s [{id:1}] AS b ON b.id=a.id SELECT a.id AS aid,b.id AS bid`, "not applicable", 0, 0},
		{"generated chain", `FROM [{id:1}] AS a %s [{id:1}] AS b ON b.id=a.id JOIN [{id:1}] AS c ON b.id=c.id SELECT b.id AS bid`, "not applicable", 0, 0},
		{"constant view", `$v=FROM beta | eval id=1; $out=SELECT a.host,b.id AS constant_id FROM alpha AS a %s $v AS b ON b.id=a.id;`, "indeterminate", 2, 0},
		{"source-derived view", `$v=FROM beta | eval id=uid; $out=SELECT a.host,b.id AS derived_id FROM alpha AS a %s $v AS b ON b.id=a.id;`, "connected", 2, 1},
	} {
		for _, join := range []string{"LEFT JOIN", "INNER JOIN"} {
			t.Run(tc.name+"/"+join, func(t *testing.T) {
				query := strings.Replace(tc.query, "%s", join, 1)
				r := spl2AnalyzeTest(t, query)
				if r.Status != Valid || len(r.Inputs) != tc.inputs || r.Correlation.Outcome != tc.outcome || len(r.Correlation.Edges) != tc.edges {
					t.Fatalf("derived join: status=%s inputs=%+v graph=%+v diagnostics=%+v", r.Status, r.Inputs, r.Correlation, r.Diagnostics)
				}
				for _, item := range r.Requirements.Items {
					if tc.inputs == 0 && item.Kind == "field" {
						t.Fatalf("generated field became an external obligation: %+v", item)
					}
					for _, occurrence := range item.Occurrences {
						if item.Kind == "field" && occurrence.OriginalName == "b.id" {
							t.Fatalf("derived destination became an external obligation: %+v", item)
						}
					}
				}
				output := map[string]string{"generated rows": "bid", "generated chain": "bid", "constant view": "constant_id", "source-derived view": "derived_id"}[tc.name]
				found := false
				for _, field := range r.Lineage[len(r.Lineage)-1].After.Fields {
					if field.Name == output {
						found = true
						if field.Conditional != (join == "LEFT JOIN") {
							t.Fatalf("derived output conditionality changed: %+v", field)
						}
					}
				}
				if !found {
					t.Fatalf("derived projection %s missing", output)
				}
				if tc.name == "source-derived view" {
					item := requirementItem(r.Requirements, "field", "uid", "read")
					if item == nil || item.Ownership.State != "proved" {
						t.Fatalf("derived source obligation lost: %+v", item)
					}
					for _, ref := range r.References {
						if ref.OriginalName == "b.id" && ref.Role == "read" && len(ref.OriginReferenceIDs) == 0 {
							t.Fatalf("derived source lineage lost: %+v", ref)
						}
					}
				}
			})
		}
	}
}

func TestSPL2SQLQualifiedReadsRespectTypedRemovalAndRecreation(t *testing.T) {
	for _, tc := range []struct {
		name, query, original, binding string
	}{
		{"inner removed", `FROM alpha AS a JOIN beta AS b ON a.id=b.uid | fields - secret | eval later=a.secret`, "a.secret", "unavailable"},
		{"left removed", `FROM alpha AS a LEFT JOIN beta AS b ON a.id=b.uid | fields - secret | eval later=b.secret`, "b.secret", "unavailable"},
		{"atomic dotted removed", `FROM alpha AS a JOIN beta AS b ON a.id=b.uid | fields - 'payload.id' | eval later=a.'payload.id'`, "a.'payload.id'", "unavailable"},
		{"structural removed", `FROM alpha AS a LEFT JOIN beta AS b ON a.id=b.uid | fields - payload.id | eval later=b.payload.id`, "b.payload.id", "unavailable"},
		{"atomic removal keeps structural", `FROM alpha AS a JOIN beta AS b ON a.id=b.uid | fields - 'payload.id' | eval later=a.payload.id`, "a.payload.id", "source"},
		{"structural removal keeps atomic", `FROM alpha AS a JOIN beta AS b ON a.id=b.uid | fields - payload.id | eval later=a.'payload.id'`, "a.'payload.id'", "source"},
		{"recreated", `FROM alpha AS a JOIN beta AS b ON a.id=b.uid | fields - secret | eval secret=1 | eval later=a.secret`, "a.secret", "derived"},
		{"atomic dotted recreated", `FROM alpha AS a JOIN beta AS b ON a.id=b.uid | fields - 'payload.id' | eval 'payload.id'=1 | eval later=a.'payload.id'`, "a.'payload.id'", "derived"},
		{"atomic recreation keeps structural removal", `FROM alpha AS a JOIN beta AS b ON a.id=b.uid | fields - payload.id | eval 'payload.id'=1 | eval later=a.payload.id`, "a.payload.id", "unavailable"},
		{"left recreated", `FROM alpha AS a LEFT JOIN beta AS b ON a.id=b.uid | fields - secret | eval secret=1 | eval later=b.secret`, "b.secret", "indeterminate"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := spl2AnalyzeTest(t, tc.query)
			found := false
			for _, ref := range r.References {
				if ref.Role != "read" || ref.OriginalName != tc.original {
					continue
				}
				found = true
				if ref.Binding != tc.binding {
					t.Fatalf("qualified active-row read: %+v, want %s", ref, tc.binding)
				}
				if strings.Contains(tc.name, "recreated") {
					if len(ref.OriginReferenceIDs) == 0 {
						t.Fatalf("recreated field lost its new origins: %+v", ref)
					}
					for _, id := range ref.OriginReferenceIDs {
						origin := r.References[referenceOrdinal(t, id)]
						expected := *ref.FieldIdentity
						expected.Qualifier = ""
						if len(expected.Segments) == 1 {
							expected.Kind = "atomic"
						}
						if origin.NormalizedName != ref.NormalizedName || origin.Role != "create" || origin.FieldIdentity == nil || !reflect.DeepEqual(*origin.FieldIdentity, expected) {
							t.Fatalf("recreated field resurrected source lineage: %+v", origin)
						}
					}
				}
			}
			if !found {
				t.Fatalf("missing qualified read %s", tc.original)
			}
			if tc.binding == "unavailable" || strings.Contains(tc.name, "recreated") {
				for _, item := range r.Requirements.Items {
					for _, occurrence := range item.Occurrences {
						if item.Kind == "field" && occurrence.OriginalName == tc.original {
							t.Fatalf("removed/recreated field became external obligation: %+v", item)
						}
					}
				}
			}
		})
	}
}

func TestSPL2SQLQualifiedGroupSuffixCollisionsKeepBothSources(t *testing.T) {
	for _, join := range []string{"JOIN", "LEFT JOIN"} {
		for _, groups := range []string{"a.host,b.host", "b.host,a.host"} {
			t.Run(join+"/"+groups, func(t *testing.T) {
				base := `SELECT a.host AS ahost,b.host AS bhost FROM alpha AS a ` + join + ` beta AS b ON a.id=b.uid GROUP BY ` + groups
				r := spl2AnalyzeTest(t, base+` ORDER BY ahost,bhost`)
				if r.Status != Valid || !r.Coverage.SemanticComplete || r.Correlation.Outcome != "connected" {
					t.Fatalf("qualified grouping: status=%s diagnostics=%+v graph=%+v", r.Status, r.Diagnostics, r.Correlation)
				}
				seen := map[string]bool{}
				for _, item := range r.Requirements.Items {
					if item.Kind != "field" || item.Identity != "host" || item.Role != "read" {
						continue
					}
					for _, input := range r.Inputs {
						if input.ID != item.InputID {
							continue
						}
						seen[input.Name] = true
						want := "required"
						if join == "LEFT JOIN" && input.Name == "beta" {
							want = "conditional"
						}
						if item.Ownership.State != "proved" || item.Necessity != want {
							t.Fatalf("grouped source ownership/necessity: %+v", item)
						}
					}
				}
				if len(seen) != 2 {
					t.Fatalf("grouped source lost: %+v", r.Requirements.Items)
				}
				last := r.Lineage[len(r.Lineage)-1].After
				if len(last.Fields) != 2 || last.Fields[0].Name != "ahost" || last.Fields[1].Name != "bhost" {
					t.Fatalf("qualified grouping changed output labels: %+v", last)
				}
				held := spl2AnalyzeTest(t, base+` HAVING a.secret="x"`)
				if spl2Ref(t, held, "secret", "read").Binding != "unavailable" {
					t.Fatal("closed grouping admitted missing qualified HAVING field")
				}
				visibleGroup := `SELECT a.host,b.host AS bhost FROM alpha AS a ` + join + ` beta AS b ON a.id=b.uid GROUP BY ` + groups
				visible := spl2AnalyzeTest(t, visibleGroup+` HAVING a.host="x" ORDER BY bhost`)
				if visible.Status != Valid || !visible.Coverage.SemanticComplete {
					t.Fatalf("visible qualified HAVING/order: %+v", visible.Diagnostics)
				}
				restricted := spl2AnalyzeTest(t, base+` HAVING a.host="x"`)
				if restricted.Coverage.SemanticComplete {
					t.Fatal("renamed group field bypassed HAVING visibility")
				}
				missing := spl2AnalyzeTest(t, `SELECT a.host AS ahost,b.host AS bhost FROM alpha AS a `+join+` beta AS b ON a.id=b.uid GROUP BY a.host`)
				for _, ref := range missing.References {
					if ref.Role == "read" && ref.OriginalName == "b.host" && ref.Binding != "unavailable" {
						t.Fatalf("same suffix admitted an ungrouped alias: %+v", ref)
					}
				}
				for _, query := range []string{base + ` | eval later=a.host`, visibleGroup + ` | fields - host | eval later=a.host`} {
					closed := spl2AnalyzeTest(t, query)
					for i := len(closed.References) - 1; i >= 0; i-- {
						ref := closed.References[i]
						if ref.Role == "read" && ref.OriginalName == "a.host" {
							if ref.Binding != "unavailable" {
								t.Fatalf("final projection/removal retained grouped source alias: %+v", ref)
							}
							break
						}
					}
				}
				ambiguous := spl2AnalyzeTest(t, `SELECT host AS chosen FROM alpha AS a `+join+` beta AS b ON a.id=b.uid GROUP BY `+groups)
				item := requirementItem(ambiguous.Requirements, "field", "host", "read")
				if item == nil || item.Ownership.State != "unproved" || len(item.Ownership.CandidateInputIDs) != 2 {
					t.Fatalf("unqualified grouped collision picked an owner: %+v", item)
				}
			})
		}
	}
}

func TestSPL2SQLQualifiedSelectedVisibilityKeepsSourceIdentity(t *testing.T) {
	for _, join := range []string{"JOIN", "LEFT JOIN"} {
		for _, selectFirst := range []bool{false, true} {
			for _, grouped := range []bool{false, true} {
				for _, self := range []bool{false, true} {
					from := `FROM alpha AS a ` + join + ` beta AS b ON a.id=b.uid`
					if self {
						from = `FROM $events AS a ` + join + ` $events AS b ON a.id=b.uid`
					}
					group := ""
					if grouped {
						group = ` GROUP BY a.host,b.host`
					}
					query := from + group + ` SELECT a.host,b.host AS bhost`
					if selectFirst {
						query = `SELECT a.host,b.host AS bhost ` + from + group
					}
					t.Run(query, func(t *testing.T) {
						clauses := []string{` ORDER BY b.host`, ` ORDER BY a.host,b.host`}
						if grouped {
							clauses = append(clauses, ` HAVING b.host="x"`, ` HAVING a.host="x" AND b.host="x"`)
						}
						for _, clause := range clauses {
							r := spl2AnalyzeTest(t, query+clause)
							if r.Coverage.SemanticComplete {
								t.Fatalf("another alias's suffix authorized hidden qualified read: %s", query+clause)
							}
							found := false
							for i := len(r.References) - 1; i >= 0; i-- {
								ref := r.References[i]
								if ref.OriginalName != "b.host" || ref.Role != "read" {
									continue
								}
								if ref.Binding != "indeterminate" {
									t.Fatalf("hidden qualified read classification: %+v", ref)
								}
								for _, diagnostic := range r.Diagnostics {
									if diagnostic.Code == CodeUnsupportedSemantics && diagnostic.Location == ref.Location && diagnostic.StageID == ref.StageID {
										found = true
									}
								}
								owned := false
								for _, gap := range r.Requirements.Gaps {
									if gap.Code == CodeUnsupportedSemantics && reflect.DeepEqual(gap.ReferenceIDs, []string{ref.ID}) {
										owned = true
									}
								}
								if !owned {
									t.Fatalf("visibility gap lost the exact hidden alias reference: %+v", r.Requirements.Gaps)
								}
								break
							}
							if !found {
								t.Fatal("hidden qualified read lost its located visibility diagnostic")
							}
						}
						positive := spl2AnalyzeTest(t, query+` ORDER BY a.host,bhost`)
						if positive.Status != Valid || !positive.Coverage.SemanticComplete {
							t.Fatalf("selected qualified/unqualified labels lost visibility: %+v", positive.Diagnostics)
						}
						explicit := spl2AnalyzeTest(t, strings.Replace(query, "SELECT a.host,", "SELECT a.host AS host,", 1)+` ORDER BY a.host,bhost`)
						if explicit.Status != Valid || !explicit.Coverage.SemanticComplete {
							t.Fatalf("explicit suffix-retaining qualified label lost visibility: %+v", explicit.Diagnostics)
						}
						if grouped {
							positive = spl2AnalyzeTest(t, query+` HAVING a.host="x" AND bhost="x"`)
							if positive.Status != Valid || !positive.Coverage.SemanticComplete {
								t.Fatalf("selected HAVING labels lost visibility: %+v", positive.Diagnostics)
							}
						}
					})
				}
			}
		}
	}
}

func TestSPL2SQLQualifiedSelectedVisibilityKeepsTypedAtomicName(t *testing.T) {
	query := `SELECT a.'payload.id',b.'payload.id' AS bvalue FROM alpha AS a JOIN beta AS b ON a.id=b.uid`
	for _, suffix := range []string{` ORDER BY b.'payload.id'`, ` ORDER BY a.payload.id`} {
		r := spl2AnalyzeTest(t, query+suffix)
		if r.Coverage.SemanticComplete {
			t.Fatalf("atomic label authorized a hidden alias/path: %s", suffix)
		}
	}
	positive := spl2AnalyzeTest(t, query+` ORDER BY a.'payload.id',bvalue`)
	if positive.Status != Valid || !positive.Coverage.SemanticComplete {
		t.Fatalf("qualified atomic label lost visibility: %+v", positive.Diagnostics)
	}
}
