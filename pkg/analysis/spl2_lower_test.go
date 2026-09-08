package analysis

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"testing"
)

func spl2AnalyzeTest(t *testing.T, text string) *Result {
	t.Helper()
	r, err := Analyze(QueryDocument{Text: text, Language: "spl2"})
	if err != nil {
		t.Fatal(err)
	}
	return r
}
func spl2HasCode(r *Result, code string) bool {
	for _, d := range r.Diagnostics {
		if d.Code == code {
			return true
		}
	}
	return false
}
func spl2Ref(t *testing.T, r *Result, name, role string) Reference {
	t.Helper()
	for _, ref := range r.References {
		if ref.NormalizedName == name && ref.Role == role {
			return ref
		}
	}
	t.Fatalf("missing %s %s in %+v", role, name, r.References)
	return Reference{}
}
func TestSPL2SequentialFullState(t *testing.T) {
	text := "FROM main | eval x=bytes, y=x+1 | table y"
	got := spl2AnalyzeTest(t, text)
	loc := func(start, end int) Location {
		return Location{Position{start, 1, start + 1}, Position{end, 1, end + 1}}
	}
	ref := func(id, name, role, stage, bind string, start, end int, origins ...string) Reference {
		kind := "field"
		if name == "main" {
			kind = "dataset"
		}
		return Reference{ID: id, OriginalName: text[start:end], NormalizedName: name, Kind: kind, Role: role, StageID: stage, ScopeID: "scope-0", Location: loc(start, end), Resolution: "exact", Binding: bind, OriginReferenceIDs: append([]string{}, origins...)}
	}
	field := func(name string, ids ...string) FieldBinding {
		return FieldBinding{Name: name, OriginReferenceIDs: ids}
	}
	before := FieldState{Fields: []FieldBinding{}, Removed: []string{}, Open: true}
	assigned := FieldState{Fields: []FieldBinding{field("bytes", "ref-2"), field("x", "ref-1", "ref-2"), field("y", "ref-3", "ref-4", "ref-1", "ref-2")}, Removed: []string{}, Open: true}
	after := FieldState{Fields: []FieldBinding{field("y", "ref-3", "ref-4", "ref-1", "ref-2")}, Removed: []string{}}
	want := newResult(QueryDocument{Text: text, Language: "spl2", Profile: "splunkd", Version: "current"})
	want.Status = Valid
	want.Scopes = []Scope{{ID: "scope-0", Kind: "root", Location: loc(0, len(text))}}
	want.Stages = []Stage{{"stage-0", "from", 0, "scope-0", loc(0, 9), true}, {"stage-1", "eval", 1, "scope-0", loc(12, 31), true}, {"stage-2", "table", 2, "scope-0", loc(34, 41), true}}
	want.Dependencies.Datasets = []string{"main"}
	want.References = []Reference{ref("ref-0", "main", "read", "stage-0", "not_applicable", 5, 9), ref("ref-1", "x", "create", "stage-1", "not_applicable", 17, 18, "ref-2"), ref("ref-2", "bytes", "read", "stage-1", "source", 19, 24), ref("ref-3", "y", "create", "stage-1", "not_applicable", 26, 27, "ref-4", "ref-1", "ref-2"), ref("ref-4", "x", "read", "stage-1", "derived", 28, 29, "ref-1", "ref-2"), ref("ref-5", "y", "read", "stage-2", "derived", 40, 41, "ref-3", "ref-4", "ref-1", "ref-2")}
	want.Lineage = []Lineage{{"stage-0", "scope-0", before, before, []Transition{}}, {"stage-1", "scope-0", before, assigned, []Transition{{"create", "x", []string{"ref-2"}, "ref-1", false}, {"create", "y", []string{"ref-4"}, "ref-3", false}}}, {"stage-2", "scope-0", assigned, after, []Transition{{"project", "y", []string{"ref-5"}, "", false}}}}
	if !reflect.DeepEqual(got, want) {
		g, _ := json.MarshalIndent(got, "", "  ")
		w, _ := json.MarshalIndent(want, "", "  ")
		t.Fatalf("got %s\nwant %s", g, w)
	}
	for _, u := range []SourceUniverse{{Fields: []string{"bytes", "_time"}, Complete: true}, {Fields: []string{"bytes"}}} {
		s, e := AnalyzeWithSourceUniverse(want.Document, u)
		if e != nil || !reflect.DeepEqual(s, &SourceAnalysis{Result: want, Expansions: []FieldExpansion{}}) {
			t.Fatalf("source report %+v %v", s, e)
		}
	}
}
func TestSPL2OrdinaryReadAndTransferBoundaries(t *testing.T) {
	t.Run("search literals", func(t *testing.T) {
		r := spl2AnalyzeTest(t, `search index=main code=bytes status IN (true,false) flag="-1"`)
		if r.Status != Valid || !reflect.DeepEqual(r.Dependencies.Indexes, []string{"main"}) {
			t.Fatalf("%+v", r)
		}
		for _, ref := range r.References {
			if ref.Kind == "field" && ref.NormalizedName != "code" && ref.NormalizedName != "status" && ref.NormalizedName != "flag" {
				t.Fatalf("literal read %+v", ref)
			}
		}
	})
	t.Run("where reads", func(t *testing.T) {
		r := spl2AnalyzeTest(t, `FROM main | where bytes=rate`)
		if r.Status != Valid {
			t.Fatalf("%+v", r)
		}
		spl2Ref(t, r, "bytes", "read")
		spl2Ref(t, r, "rate", "read")
	})
	t.Run("null removes", func(t *testing.T) {
		r := spl2AnalyzeTest(t, `FROM main | eval x=bytes, x=null | where x>0`)
		if r.Status != Invalid || spl2Ref(t, r, "x", "remove").Binding != "not_applicable" || spl2Ref(t, r, "x", "read").Binding != "unavailable" {
			t.Fatalf("%+v", r)
		}
		last := r.Lineage[1].Transitions[1]
		if last.Operation != "remove" || len(last.InputReferenceIDs) != 0 {
			t.Fatalf("%+v", last)
		}
	})
	t.Run("fields open", func(t *testing.T) {
		r := spl2AnalyzeTest(t, `FROM main | eval x=bytes, y=x+1 | fields y`)
		if r.Status != Incomplete || spl2Ref(t, r, "x", "read").Binding != "derived" {
			t.Fatalf("%+v", r)
		}
	})
	t.Run("internal removal", func(t *testing.T) {
		r, e := AnalyzeWithSourceUniverse(QueryDocument{Text: `FROM main | fields - _time | eval y=bytes | fields y`, Language: "spl2"}, SourceUniverse{Fields: []string{"bytes", "_time", "_raw"}, Complete: true})
		if e != nil {
			t.Fatal(e)
		}
		last := r.Result.Lineage[len(r.Result.Lineage)-1].After
		if r.Result.Status != Valid || len(last.Fields) != 2 || last.Fields[0].Name != "_raw" || !reflect.DeepEqual(last.Removed, []string{"_time"}) {
			t.Fatalf("%+v", r)
		}
	})
	t.Run("independent rename", func(t *testing.T) {
		r := spl2AnalyzeTest(t, `FROM main | rename a AS x, b AS y | table x,y`)
		if r.Status != Valid || spl2Ref(t, r, "x", "read").Binding != "derived" {
			t.Fatalf("%+v", r)
		}
	})
	for _, tail := range []string{"a AS x, a AS y", "a AS x, b AS x", "a AS b, b AS c", "a AS b, b AS a"} {
		t.Run(tail, func(t *testing.T) {
			r := spl2AnalyzeTest(t, "FROM main | rename "+tail)
			if r.Status != Invalid {
				t.Fatalf("%+v", r)
			}
		})
	}
	t.Run("aggregate output", func(t *testing.T) {
		r := spl2AnalyzeTest(t, `FROM main | stats count(), sum(bytes), dc(action) BY host`)
		if r.Status != Valid {
			t.Fatalf("%+v", r)
		}
		if r.Lineage[1].After.Open || len(r.Lineage[1].After.Fields) != 4 {
			t.Fatalf("%+v", r.Lineage)
		}
		for _, name := range []string{"count", "sum(bytes)", "dc(action)"} {
			spl2Ref(t, r, name, "output")
		}
	})
	t.Run("lookup local", func(t *testing.T) {
		r := spl2AnalyzeTest(t, `FROM main | eval role="local" | lookup users uid AS user OUTPUTNEW role, label AS name | sort name | dedup user | head 5 | reverse`)
		if r.Status != Valid || len(r.Lineage[2].Transitions) != 1 || r.Lineage[2].Transitions[0].Output != "name" {
			t.Fatalf("%+v", r)
		}
		spl2Ref(t, r, "user", "read")
		for _, ref := range r.References {
			if ref.Kind == "field" && ref.NormalizedName == "uid" {
				t.Fatal("catalog column became event read")
			}
		}
	})
}
func TestSPL2MixedDialectConcurrent(t *testing.T) {
	splText := `search user=* | eval x=coalesce(user)`
	baseline, e := Analyze(QueryDocument{Text: splText})
	if e != nil {
		t.Fatal(e)
	}
	expected, _ := json.Marshal(baseline)
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 8; j++ {
				s, e := Analyze(QueryDocument{Text: splText})
				b, _ := json.Marshal(s)
				if e != nil || string(b) != string(expected) {
					t.Errorf("SPL changed %s %v", b, e)
				}
				r, e := Analyze(QueryDocument{Text: `FROM main | eval x=coalesce(user) | table x`, Language: "spl2"})
				if e != nil || r.Status != Valid {
					t.Errorf("SPL2 %+v %v", r, e)
				}
				universe := SourceUniverse{Fields: []string{"host", "host.child"}, Complete: true, Resolve: func(string) SourceFieldAdmission { return SourceFieldAdmitted }}
				legacy, err := AnalyzeWithSourceUniverse(QueryDocument{Text: "fields host*"}, universe)
				if err != nil || len(legacy.Expansions) != 1 || !legacy.Expansions[0].Complete || len(legacy.Expansions[0].Matches) != 2 {
					t.Errorf("SPL resolver isolation: %+v %v", legacy, err)
				}
				selected, err := AnalyzeWithSourceUniverse(QueryDocument{Text: "FROM main | fields 'host*'", Language: "spl2"}, universe)
				if err != nil || len(selected.Expansions) != 1 || selected.Expansions[0].Complete || len(selected.Expansions[0].Matches) != 1 || selected.Expansions[0].Matches[0].Name != "host" {
					t.Errorf("SPL2 resolver isolation: %+v %v", selected, err)
				}
			}
		}()
	}
	wg.Wait()
}
func TestSPL2Capabilities(t *testing.T) {
	m, e := CapabilitiesFor(CapabilityOptions{Language: "spl2"})
	if e != nil || m.Language != "spl2" {
		t.Fatalf("%+v %v", m, e)
	}
	found := false
	for i, c := range m.Functions {
		if c.Name == "coalesce" {
			found = true
			if !c.SemanticSupported || !strings.Contains(c.Limitations[0], "1") {
				t.Fatalf("%+v", c)
			}
			m.Functions[i].Limitations[0] = "mutated"
		}
	}
	if !found {
		t.Fatal("missing coalesce")
	}
	n, _ := CapabilitiesFor(CapabilityOptions{Language: "spl2"})
	for _, c := range n.Functions {
		if c.Limitations[0] == "mutated" {
			t.Fatal("shared mutable capability")
		}
	}
}

func TestSPL2QuotedDotsAcrossSharedTransfers(t *testing.T) {
	for _, tail := range []string{`table 'actor.name'`, `fields 'actor.name'`, `rename 'actor.name' AS name`, `stats count() BY 'actor.name'`} {
		document := QueryDocument{Text: "FROM main | " + tail, Language: "spl2"}
		for _, universe := range []SourceUniverse{{Fields: []string{"actor.name"}, Complete: true}, {Fields: []string{"actor.name"}}, {Fields: []string{"actor.name"}, Complete: true, Resolve: func(string) SourceFieldAdmission { return SourceFieldAdmitted }}} {
			source, err := AnalyzeWithSourceUniverse(document, universe)
			if err != nil {
				t.Fatal(err)
			}
			expected := "source"
			if universe.Resolve != nil || (!universe.Complete && strings.HasPrefix(tail, "fields")) {
				expected = "indeterminate"
			}
			found := false
			for _, ref := range source.Result.References {
				if ref.NormalizedName == "actor.name" {
					found = true
					if ref.Binding != expected {
						t.Fatalf("%s %+v", tail, ref)
					}
				}
			}
			if !found {
				t.Fatal("missing dotted source")
			}
		}
		r := spl2AnalyzeTest(t, document.Text)
		for _, ref := range r.References {
			if ref.NormalizedName == "actor.name" && ref.Binding != "source" && !strings.HasPrefix(tail, "fields") {
				t.Fatalf("plain name lost %+v", ref)
			}
		}
	}
}

func TestSPL2CommandSourceEvidence(t *testing.T) {
	t.Run("finite wildcard and internals", func(t *testing.T) {
		text := `FROM main | fields - _time | fields 'b*'`
		r, e := AnalyzeWithSourceUniverse(QueryDocument{Text: text, Language: "spl2"}, SourceUniverse{Fields: []string{"bytes", "bits", "_raw", "_time"}, Complete: true})
		if e != nil {
			t.Fatal(e)
		}
		if r.Result.Status != Valid || !reflect.DeepEqual(r.Expansions, []FieldExpansion{{ReferenceID: "ref-2", Complete: true, Matches: []ExpandedField{{Name: "bits", Binding: "source"}, {Name: "bytes", Binding: "source"}}}}) {
			t.Fatalf("%+v", r)
		}
		want := FieldState{Fields: []FieldBinding{{Name: "_raw", OriginReferenceIDs: []string{}}, {Name: "bits", OriginReferenceIDs: []string{"ref-2"}}, {Name: "bytes", OriginReferenceIDs: []string{"ref-2"}}}, Removed: []string{"_time"}}
		if !reflect.DeepEqual(r.Result.Lineage[2].After, want) {
			t.Fatalf("%+v", r.Result.Lineage)
		}
		if len(r.Result.References) != 3 {
			t.Fatalf("fabricated internals %+v", r.Result.References)
		}
	})
	for _, command := range []string{`eventstats count() AS n`, `streamstats BY host current=true reset before code=500 window=3 count() AS n`, `stats allnum=true count() AS n`, `streamstats current=false count() AS n`} {
		t.Run(command, func(t *testing.T) {
			r := spl2AnalyzeTest(t, "FROM main | eval prior=1 | "+command)
			if r.Status != Valid {
				t.Fatalf("%+v", r)
			}
			last := r.Lineage[2].After
			preserve := !strings.HasPrefix(command, "stats")
			seenPrior := false
			for _, f := range last.Fields {
				seenPrior = seenPrior || f.Name == "prior"
			}
			if preserve != seenPrior {
				t.Fatalf("%+v", last)
			}
			conditional := strings.Contains(command, "allnum=true") || strings.Contains(command, "current=false")
			if r.Lineage[2].Transitions[len(r.Lineage[2].Transitions)-1].Conditional != conditional {
				t.Fatalf("%+v", r.Lineage)
			}
			if strings.Contains(command, "code") {
				spl2Ref(t, r, "code", "read")
			}
		})
	}
	for _, query := range []string{`search index=main code=-1`, `search index=main code=-bytes`, `search index=main code=-(1+2)`} {
		r := spl2AnalyzeTest(t, query)
		if r.Status != Incomplete {
			t.Fatalf("%+v", r)
		}
		for _, ref := range r.References {
			if ref.Kind == "field" && ref.NormalizedName != "code" {
				t.Fatalf("search literal evaluated %+v", ref)
			}
		}
	}
	t.Run("source reset", func(t *testing.T) {
		r := spl2AnalyzeTest(t, `FROM first | eval local=1 | FROM second | where local>0`)
		if spl2Ref(t, r, "local", "read").Binding != "source" || len(r.Lineage[2].After.Fields) != 0 {
			t.Fatalf("%+v", r)
		}
	})
	t.Run("UTF8 CRLF caller data", func(t *testing.T) {
		text := "FROM 'données'\r\n| eval 'métrique'=bytes | where 'métrique'>0"
		r, e := Analyze(QueryDocument{Text: text, Language: "spl2", SourceID: " exact\r\n "})
		if e != nil || r.Status != Valid || r.Document.Text != text || r.Document.SourceID != " exact\r\n " {
			t.Fatalf("%+v %v", r, e)
		}
		ref := spl2Ref(t, r, "métrique", "create")
		if ref.OriginalName != "'métrique'" || ref.Location.Start.Line != 2 || ref.Location.Start.Column != 8 || text[ref.Location.Start.Offset:ref.Location.End.Offset] != ref.OriginalName {
			t.Fatalf("%+v", ref)
		}
	})
}

func TestSPL2NullTestReadsDoNotEstablishPresence(t *testing.T) {
	for _, name := range []string{"isnull", "isnotnull"} {
		t.Run(name, func(t *testing.T) {
			r := spl2AnalyzeTest(t, "FROM main | eval answer="+name+"((absent))")
			ref := spl2Ref(t, r, "absent", "null_test")
			if ref.Binding != "source" || r.Status != Valid || len(r.Lineage[1].After.Fields) != 1 || r.Lineage[1].After.Fields[0].Name != "answer" {
				t.Fatalf("%+v", r)
			}
		})
	}
	r := spl2AnalyzeTest(t, `FROM main | eval x=null, answer=isnull(x)`)
	if r.Status != Valid || spl2Ref(t, r, "x", "null_test").Binding != "unavailable" || len(r.Lineage[1].After.Removed) != 1 {
		t.Fatalf("%+v", r)
	}
	r = spl2AnalyzeTest(t, `FROM main | eval x=null, answer=isnull(x) | where x>0`)
	if r.Status != Invalid || spl2Ref(t, r, "x", "read").Binding != "unavailable" {
		t.Fatalf("%+v", r)
	}
	r = spl2AnalyzeTest(t, `FROM main | eval x=tonumber("17"), answer=isnotnull(x)`)
	if r.Status != Valid || spl2Ref(t, r, "x", "null_test").Binding != "indeterminate" || !r.Lineage[1].After.Fields[1].Conditional {
		t.Fatalf("%+v", r)
	}
	for _, query := range []string{`FROM main | eval x=isnull(a+1)`, `FROM main | eval x=isnotnull(abs(a))`, `FROM main | eval x=isnull(a,b)`, `FROM main | eval x=isnull(value:a)`} {
		r := spl2AnalyzeTest(t, query)
		spl2Ref(t, r, "a", "read")
		for _, ref := range r.References {
			if ref.NormalizedName == "a" && ref.Role == "null_test" {
				t.Fatalf("arbitrary subtree exempted %+v", r)
			}
		}
	}
}

func TestSPL2ResolverWildcardCandidateIdentity(t *testing.T) {
	for _, pattern := range []string{"actor.*", "actor*", "*"} {
		for _, resolver := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/resolver=%v", pattern, resolver), func(t *testing.T) {
				text := `FROM main | eval 'actor.local'=1 | fields '` + pattern + `'`
				u := SourceUniverse{Fields: []string{"actor.name", "host"}, Complete: true}
				if resolver {
					u.Resolve = func(string) SourceFieldAdmission { return SourceFieldAdmitted }
				}
				r, e := AnalyzeWithSourceUniverse(QueryDocument{Text: text, Language: "spl2"}, u)
				if e != nil {
					t.Fatal(e)
				}
				if len(r.Expansions) != 1 {
					t.Fatalf("%+v", r)
				}
				exp := r.Expansions[0]
				if exp.Complete == resolver {
					t.Fatalf("bad completeness %+v", r)
				}
				matches := map[string]string{}
				for _, m := range exp.Matches {
					matches[m.Name] = m.Binding
				}
				if matches["actor.local"] != "derived" {
					t.Fatalf("lost derived %+v", r)
				}
				if (matches["actor.name"] != "") == resolver {
					t.Fatalf("ambiguous candidate proof %+v", r)
				}
				if pattern == "*" && matches["host"] != "source" {
					t.Fatalf("lost flat proof %+v", r)
				}
				if resolver {
					for _, f := range r.Result.Lineage[2].After.Fields {
						if f.Name == "actor.name" {
							t.Fatalf("materialized withheld source %+v", r)
						}
					}
					used, e := AnalyzeWithSourceUniverse(QueryDocument{Text: text + ` | where 'actor.local'>0 AND 'actor.name'>0`, Language: "spl2"}, u)
					if e != nil {
						t.Fatal(e)
					}
					if spl2Ref(t, used.Result, "actor.local", "read").Binding != "derived" || spl2Ref(t, used.Result, "actor.name", "read").Binding != "indeterminate" {
						t.Fatalf("%+v", used)
					}
				}
			})
		}
	}
	t.Run("implicit internals", func(t *testing.T) {
		r, e := AnalyzeWithSourceUniverse(QueryDocument{Text: `FROM main | fields host`, Language: "spl2"}, SourceUniverse{Fields: []string{"_object.name", "host"}, Complete: true, Resolve: func(string) SourceFieldAdmission { return SourceFieldAdmitted }})
		if e != nil {
			t.Fatal(e)
		}
		for _, f := range r.Result.Lineage[1].After.Fields {
			if f.Name == "_object.name" {
				t.Fatalf("implicit ambiguous internal %+v", r)
			}
		}
	})
	t.Run("partial flat removal", func(t *testing.T) {
		r, e := AnalyzeWithSourceUniverse(QueryDocument{Text: `FROM main | fields - 'actor*'`, Language: "spl2"}, SourceUniverse{Fields: []string{"actor.name"}})
		if e != nil {
			t.Fatal(e)
		}
		if len(r.Expansions) != 1 || r.Expansions[0].Complete || !reflect.DeepEqual(r.Expansions[0].Matches, []ExpandedField{{Name: "actor.name", Binding: "source"}}) {
			t.Fatalf("%+v", r)
		}
	})
}
