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

func TestSPL2LookupImplicitOutputUncertainty(t *testing.T) {
	for _, pair := range []struct{ match, local string }{{"uid AS user", "user"}, {"id AS account", "account"}} {
		r := spl2AnalyzeTest(t, "FROM main | lookup users "+pair.match+" | table "+pair.local)
		var reads []Reference
		for _, ref := range r.References {
			if ref.Kind == "field" && ref.NormalizedName == pair.local {
				reads = append(reads, ref)
			}
		}
		if r.Status != Incomplete || len(reads) != 2 || reads[0].Binding != "source" || reads[1].Binding != "indeterminate" || !reflect.DeepEqual(reads[1].OriginReferenceIDs, []string{reads[0].ID}) {
			t.Fatalf("unknown lookup outputs may overwrite local alias: %+v", r)
		}
		if f := r.Lineage[1].After.Fields; len(f) != 1 || f[0].Name != pair.local || !f[0].Conditional || !reflect.DeepEqual(f[0].OriginReferenceIDs, []string{reads[0].ID}) {
			t.Fatalf("post-lookup field evidence: %+v", r)
		}
	}
	for _, suffix := range []string{"uid | table uid", "uid AS user OUTPUT name | table user", "uid AS user OUTPUTNEW name | table user"} {
		r := spl2AnalyzeTest(t, "FROM main | lookup users "+suffix)
		if r.References[len(r.References)-1].Binding != "source" {
			t.Fatalf("protected key/explicit output changed: %+v", r)
		}
	}
	r := spl2AnalyzeTest(t, "FROM main | eval gone=null | lookup users uid AS user | table gone")
	if r.Status != Incomplete || spl2Ref(t, r, "gone", "read").Binding != "indeterminate" || spl2HasCode(r, CodeUnavailableField) {
		t.Fatalf("unknown output cannot prove absence: %+v", r)
	}
}

func TestSPL2ConflictingAllnumOptions(t *testing.T) {
	for _, command := range []string{"stats", "eventstats"} {
		for _, values := range []string{"true allnum=false", "false allnum=true"} {
			r := spl2AnalyzeTest(t, "FROM main | "+command+" allnum="+values+" count()")
			if r.Status != Incomplete || !r.Coverage.SyntaxComplete || r.Coverage.SemanticComplete || !spl2HasCode(r, CodeUnsupportedSemantics) {
				t.Fatalf("conflicting option precedence is unproved: %+v", r)
			}
			located := false
			for _, d := range r.Diagnostics {
				if d.Code == CodeUnsupportedSemantics && strings.HasPrefix(r.Document.Text[d.Location.Start.Offset:d.Location.End.Offset], "allnum=") {
					located = true
				}
			}
			if !located {
				t.Fatalf("missing located conflicting option diagnostic: %+v", r)
			}
		}
		for _, value := range []string{"true", "false"} {
			r := spl2AnalyzeTest(t, "FROM main | "+command+" allnum="+value+" allnum="+value+" count()")
			if r.Status != Valid || !r.Coverage.SemanticComplete {
				t.Fatalf("identical repeated value changed: %+v", r)
			}
		}
	}
}

func TestSPL2RejectedAssignmentEffects(t *testing.T) {
	for _, tc := range []struct {
		expression string
		intact     bool
	}{
		{`@"She said ""yes""`, false},
		{`coalesce(values:primary,backup)`, false},
		{`[[1,2],bytes+1`, false},
		{`map(items,($x) ->)`, false},
		{`round(precision:2,bytes)`, false},
		{`coalesce(values:a,b)`, false},
		{`{a:1,a:2}`, true},
		{`{name:user,"name":host}`, true},
		{`{'display name':user,"display name":role}`, true},
		{`{a:1,a:2}`, true},
		{`{name:user,"name":host}`, true},
	} {
		t.Run(tc.expression, func(t *testing.T) {
			r := spl2AnalyzeTest(t, `FROM main | eval rejected=`+tc.expression)
			if r.Status != Invalid || r.Coverage.SyntaxComplete != tc.intact || r.Coverage.SemanticComplete || !spl2HasCode(r, CodeSyntaxError) {
				t.Fatalf("rejected assignment coverage: %+v", r)
			}
			for _, ref := range r.References {
				if ref.NormalizedName == "rejected" {
					t.Fatalf("damaged target reference: %+v", ref)
				}
			}
			for _, field := range r.Lineage[len(r.Lineage)-1].After.Fields {
				if field.Name == "rejected" {
					t.Fatalf("damaged target state: %+v", field)
				}
			}
			assertCorpusIntegrity(t, r)
		})
	}
	for _, expression := range []string{`{a:user,b:host}`, `[[1,2],bytes+1]`, `mystery(bytes)`} {
		r := spl2AnalyzeTest(t, `FROM main | eval kept=`+expression)
		spl2Ref(t, r, "kept", "create")
		if expression == `mystery(bytes)` && !r.Lineage[len(r.Lineage)-1].After.Fields[len(r.Lineage[len(r.Lineage)-1].After.Fields)-1].Conditional {
			t.Fatalf("unknown intact output promoted: %+v", r)
		}
	}
	r := spl2AnalyzeTest(t, `FROM main | eval first=1, rejected={a:user,"a":host}, last=2`)
	spl2Ref(t, r, "first", "create")
	spl2Ref(t, r, "last", "create")
	spl2Ref(t, r, "user", "read")
	spl2Ref(t, r, "host", "read")
	for _, ref := range r.References {
		if ref.NormalizedName == "rejected" {
			t.Fatalf("invalid constructor target survived: %+v", r)
		}
	}
}

func TestSPL2DeferredCommandOriginalRoles(t *testing.T) {
	for _, tc := range []struct {
		query                  string
		reads, groups, outputs []string
	}{
		{`FROM main | bin span=15m _time AS slot`, []string{"_time"}, nil, []string{"slot"}},
		{`FROM main | rex field=message offset_field=offsets @"(?<digits>\d+)"`, []string{"message"}, nil, []string{"offsets"}},
		{`FROM main | spath input=payload output=user_name path="actor.name"`, []string{"payload"}, nil, []string{"user_name"}},
		{`tstats aggregates=[sum(bytes)] datamodel_name='Traffic.All' predicate=(port=443) byfields=[host,source]`, []string{"bytes", "port"}, []string{"host", "source"}, nil},
		{`mstats aggregates=[avg('cpu.load')] predicate=(index="metrics") byfields=[host]`, []string{"cpu.load"}, []string{"host"}, nil},
		{`FROM main | timechart agg=(sum(bytes)) avg(size) BY host`, []string{"bytes", "size"}, []string{"host"}, nil},
		{`FROM main | timechart agg=(max(bytes) AS peak) count() BY host`, []string{"bytes"}, []string{"host"}, []string{"peak"}},
		{`FROM main | makemv delim=":" labels`, []string{"labels"}, nil, nil},
		{`FROM main | mvexpand limit=2 tokens`, []string{"tokens"}, nil, nil},
		{`FROM main | mvcombine delim=";" account`, []string{"account"}, nil, nil},
	} {
		t.Run(tc.query, func(t *testing.T) {
			r := spl2AnalyzeTest(t, tc.query)
			if r.Status != Incomplete || !r.Coverage.SyntaxComplete || r.Coverage.SemanticComplete {
				t.Fatalf("deferred effects changed: %+v", r)
			}
			for _, names := range []struct {
				role  string
				names []string
			}{{"read", tc.reads}, {"group", tc.groups}, {"output", tc.outputs}} {
				for _, name := range names.names {
					ref := spl2Ref(t, r, name, names.role)
					if names.role == "output" && ref.Binding != "not_applicable" {
						t.Fatalf("candidate binding: %+v", ref)
					}
				}
			}
			if strings.HasPrefix(tc.query, "tstats") {
				found := false
				for _, ref := range r.References {
					if ref.Kind == "data_model" && ref.NormalizedName == "Traffic.All" && ref.OriginalName == "'Traffic.All'" {
						found = true
					}
				}
				if !found || !reflect.DeepEqual(r.Dependencies.DataModels, []string{"Traffic.All"}) || len(r.Dependencies.Datasets) != 0 {
					t.Fatalf("atomic model dependency: %+v", r)
				}
			}
			if strings.HasPrefix(tc.query, "mstats") && !reflect.DeepEqual(r.Dependencies.Indexes, []string{"metrics"}) {
				t.Fatalf("typed metrics index selector: %+v", r)
			}
			allowed := map[string]bool{}
			for _, name := range append(append(append([]string{}, tc.reads...), tc.groups...), tc.outputs...) {
				allowed[name] = true
			}
			for _, ref := range r.References {
				if ref.Kind == "field" && !allowed[ref.NormalizedName] {
					t.Fatalf("invented field role %+v", ref)
				}
			}
			for _, name := range tc.outputs {
				for _, field := range r.Lineage[len(r.Lineage)-1].After.Fields {
					if field.Name == name {
						t.Fatalf("candidate proved output %+v", r)
					}
				}
			}
			for _, lineage := range r.Lineage {
				if len(lineage.Transitions) != 0 {
					t.Fatalf("deferred effect invented transition: %+v", r)
				}
			}
			assertCorpusIntegrity(t, r)
		})
	}
}

func TestSPL2DeferredGeneratingInputsDoNotProveOutput(t *testing.T) {
	for _, query := range []string{`tstats aggregates=[sum(bytes)] | table bytes`, `mstats aggregates=[avg(bytes)] | table bytes`, `FROM main | timechart sum(bytes) | table bytes`} {
		r := spl2AnalyzeTest(t, query)
		var refs []Reference
		for _, ref := range r.References {
			if ref.Kind == "field" && ref.NormalizedName == "bytes" {
				refs = append(refs, ref)
			}
		}
		if len(refs) != 2 || refs[0].Binding != "source" || refs[1].Binding != "indeterminate" {
			t.Fatalf("input presence leaked through generator: %+v", r)
		}
		if !r.Lineage[len(r.Lineage)-2].After.Open || !r.Lineage[len(r.Lineage)-2].After.Uncertain {
			t.Fatalf("unproved generating output closed: %+v", r)
		}
	}
}

func TestSPL2DeferredOverwriteBoundaries(t *testing.T) {
	for _, tc := range []struct {
		command                             string
		firstConditional, secondConditional bool
	}{
		{`rex mode=sed field=first "s/x/y/g"`, true, false},
		{`rex field=first "(?<unknown>.*)"`, true, true},
		{`spath input=first output=second path="name"`, false, true},
		{`spath input=first`, true, true},
		{`bin span=1 first AS second`, false, true},
		{`makemv delim=":" first`, true, false},
	} {
		r := spl2AnalyzeTest(t, `FROM main | eval first="x", second="y" | `+tc.command)
		if r.Status != Incomplete {
			t.Fatalf("deferred control: %+v", r)
		}
		last := r.Lineage[len(r.Lineage)-1]
		if len(last.After.Fields) != 2 || last.After.Fields[0].Name != "first" || last.After.Fields[0].Conditional != tc.firstConditional || last.After.Fields[1].Name != "second" || last.After.Fields[1].Conditional != tc.secondConditional || len(last.Transitions) != 0 {
			t.Fatalf("overwrite boundary: %+v", r)
		}
		for i, f := range last.Before.Fields {
			if !reflect.DeepEqual(f.OriginReferenceIDs, last.After.Fields[i].OriginReferenceIDs) {
				t.Fatalf("origin changed: %+v", r)
			}
		}
	}
}

func TestSPL2ExternalJobAndJoinIntentions(t *testing.T) {
	for _, sid := range []string{`1780123000.5`, `"1780123000.5"`, `'saved_job'`} {
		t.Run(sid, func(t *testing.T) {
			r := spl2AnalyzeTest(t, "loadjob "+sid)
			if r.Status != Incomplete || len(r.References) != 1 || r.References[0].Kind != "search_job" || r.References[0].Role != "read" || r.References[0].Binding != "not_applicable" || r.References[0].OriginalName != sid || r.References[0].Location.Start.Offset != len("loadjob ") || len(r.Dependencies.Indexes) != 0 || len(r.Lineage[0].After.Fields) != 0 {
				t.Fatalf("static job intention: %+v", r)
			}
			assertCorpusIntegrity(t, r)
		})
	}
	for _, q := range []string{`loadjob "${job}"`, `loadjob "broken`, `loadjob`} {
		r := spl2AnalyzeTest(t, q)
		for _, ref := range r.References {
			if ref.Kind == "search_job" {
				t.Fatalf("unproved job identity: %+v", r)
			}
		}
	}
	r := spl2AnalyzeTest(t, `FROM main | join left=L right=R where L.id=R.uid [FROM other | table uid]`)
	if r.Status != Incomplete || len(r.Scopes) != 2 {
		t.Fatalf("join scope: %+v", r)
	}
	for _, name := range []string{"L.id", "R.uid"} {
		ref := spl2Ref(t, r, name, "read")
		if ref.Binding != "indeterminate" || ref.ScopeID != "scope-0" || len(ref.OriginReferenceIDs) != 0 {
			t.Fatalf("qualified join input: %+v", ref)
		}
	}
	for _, line := range r.Lineage {
		if line.ScopeID == "scope-0" && len(line.After.Fields) != 0 {
			t.Fatalf("join installed qualified inputs: %+v", r)
		}
	}
	if spl2Ref(t, r, "uid", "read").ScopeID != "scope-1" {
		t.Fatalf("child scope lost: %+v", r)
	}
	assertCorpusIntegrity(t, r)
}

func TestSPL2MetricsSelectorTypedPosition(t *testing.T) {
	for _, predicate := range []string{`index="metrics"`, `index=metrics`, `index="metrics" AND port=443`} {
		r := spl2AnalyzeTest(t, `mstats aggregates=[count()] predicate=(`+predicate+`) byfields=[index]`)
		if !reflect.DeepEqual(r.Dependencies.Indexes, []string{"metrics"}) {
			t.Fatalf("selector lost: %+v", r)
		}
		for _, ref := range r.References {
			if ref.Kind == "field" && (ref.NormalizedName == "metrics" || (ref.NormalizedName == "index" && ref.Role != "group")) {
				t.Fatalf("selector became input: %+v", ref)
			}
		}
		spl2Ref(t, r, "index", "group")
	}
	for _, predicate := range []string{`'index'="metrics"`, `index=lower("metrics")`, `index="metrics*"`, `index=actor.name`, `index="${catalog}"`} {
		r := spl2AnalyzeTest(t, `mstats aggregates=[count()] predicate=(`+predicate+`)`)
		if len(r.Dependencies.Indexes) != 0 || r.Status != Incomplete {
			t.Fatalf("unproved exact source membership: %+v", r)
		}
	}
	for _, query := range []string{`FROM main | rex "plain"`, `FROM main | spath`} {
		r := spl2AnalyzeTest(t, query)
		for _, ref := range r.References {
			if ref.NormalizedName == "_raw" {
				t.Fatalf("implicit span fabricated: %+v", r)
			}
		}
	}
}

func TestSPL2UnresolvedMetricsSelectorKeepsItsRole(t *testing.T) {
	for _, rhs := range []string{`"${catalog}"`, `lower(catalog)`, `catalog.name`} {
		r := spl2AnalyzeTest(t, `mstats aggregates=[count()] predicate=(index=`+rhs+`)`)
		if r.Status != Incomplete || len(r.Dependencies.Indexes) != 0 {
			t.Fatalf("unresolved selector promoted: %+v", r)
		}
		for _, ref := range r.References {
			if ref.Kind == "field" && ref.NormalizedName == "index" {
				t.Fatalf("selector label became field: %+v", r)
			}
		}
		spl2Ref(t, r, "catalog", "read")
	}
}

func TestSPL2RecoveredInputAndSelectorSoundness(t *testing.T) {
	for _, tail := range []string{`severity=`, `status IN (401,,403)`, `"unterminated`, `a=1 AND`, `earliest=`, `_index_latest=-`} {
		t.Run(tail, func(t *testing.T) {
			r := spl2AnalyzeTest(t, `search index=app `+tail)
			if r.Status != Invalid || !reflect.DeepEqual(r.Dependencies.Indexes, []string{"app"}) {
				t.Fatalf("intact index lost: %+v", r)
			}
			for _, ref := range r.References {
				if ref.Kind == "field" && (ref.NormalizedName == "index" || ref.NormalizedName == "_index_latest" || ref.NormalizedName == "earliest") {
					t.Fatalf("modifier label became field: %+v", r)
				}
			}
		})
	}
	for _, command := range []string{`table server account`, `eval x=1 y=2`, `reverse a`, `head null=true`, `stats count() sum(bytes) AS total`, `sort -host +num(bytes)`} {
		t.Run(command, func(t *testing.T) {
			r := spl2AnalyzeTest(t, `FROM main | `+command)
			if r.Status != Invalid || r.Stages[1].SemanticComplete {
				t.Fatalf("malformed owner promoted: %+v", r)
			}
		})
	}
	for _, command := range []string{`stats`, `eventstats`, `streamstats`} {
		t.Run(command+" alias", func(t *testing.T) {
			r := spl2AnalyzeTest(t, `FROM main | `+command+` count() as`)
			for _, ref := range r.References {
				if ref.Kind == "field" && ref.NormalizedName == "count" {
					t.Fatalf("damaged alias fell back: %+v", r)
				}
			}
		})
	}
	for _, command := range []string{`table "host"`, `table "host${suffix}"`, `table "${if(flag,"*","host")}"`, `fields host*`} {
		t.Run(command, func(t *testing.T) {
			r := spl2AnalyzeTest(t, `FROM main | `+command)
			for _, ref := range r.References {
				if ref.NormalizedName == "host${suffix}" || ref.OriginalName == `"host"` {
					t.Fatalf("unproved static selector: %+v", r)
				}
			}
			for _, l := range r.Lineage {
				for _, f := range l.After.Fields {
					if f.Name == "" {
						t.Fatalf("empty field: %+v", r)
					}
				}
			}
			assertCorpusIntegrity(t, r)
		})
	}
	for _, q := range []string{`search index=app earliest=unproved`, `search index=app _index_latest=-`} {
		r := spl2AnalyzeTest(t, q)
		for _, ref := range r.References {
			if ref.Kind == "field" && (ref.NormalizedName == "earliest" || ref.NormalizedName == "_index_latest") {
				t.Fatalf("time modifier is not field: %+v", r)
			}
		}
	}
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
	want.Lineage = []Lineage{{StageID: "stage-0", ScopeID: "scope-0", Before: before, After: before, Transitions: []Transition{}}, {StageID: "stage-1", ScopeID: "scope-0", Before: before, After: assigned, Transitions: []Transition{{"create", "x", []string{"ref-2"}, "ref-1", false}, {"create", "y", []string{"ref-4"}, "ref-3", false}}}, {StageID: "stage-2", ScopeID: "scope-0", Before: assigned, After: after, Transitions: []Transition{{"project", "y", []string{"ref-5"}, "", false}}}}
	want.Requirements = RequirementSet{
		SchemaVersion: 1,
		Query: RequirementQueryIdentity{
			Language: "spl2", Profile: "splunkd", Version: "current",
			QueryDigest: "sha256:159f5b7d4683ac55dd0efac524ba3ea30133c495a91e81af6c695b70d1e16afa",
		},
		CapabilityRevision: "sha256:f1391296cfbc616e9bb1b1828e2471e37b60a35c0555654c0734640e072a0437",
		QueryStatus:        Valid,
		Coverage:           RequirementCoverage{Complete: true, Reasons: []string{}},
		Items: []RequirementItem{
			{ID: "req-1", Kind: "dataset", Identity: "main", Role: "read", Necessity: "required", Origin: "direct", Resolution: "exact", Occurrences: []RequirementOccurrence{{ReferenceID: "ref-0", OriginalName: "main", Binding: "not_applicable", StageID: "stage-0", ScopeID: "scope-0", Location: loc(5, 9)}}},
			{ID: "req-2", Kind: "field", Identity: "bytes", Role: "read", Necessity: "required", Origin: "direct", Resolution: "exact", Occurrences: []RequirementOccurrence{{ReferenceID: "ref-2", OriginalName: "bytes", Binding: "source", StageID: "stage-1", ScopeID: "scope-0", Location: loc(19, 24)}}},
		},
		Gaps:        []RequirementGap{},
		Diagnostics: []Diagnostic{},
	}
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

func TestSPL2DeferredDestinationAndPrerequisiteBoundaries(t *testing.T) {
	t.Run("fillnull intentions", func(t *testing.T) {
		r := spl2AnalyzeTest(t, `FROM main | fillnull value="0" host,user`)
		for _, name := range []string{"host", "user"} {
			ref := spl2Ref(t, r, name, "output")
			if ref.Binding != "not_applicable" {
				t.Fatalf("destination binding: %+v", ref)
			}
		}
		if len(r.Lineage[len(r.Lineage)-1].After.Fields) != 0 {
			t.Fatalf("destination invented state: %+v", r)
		}
	})
	t.Run("dynamic SID", func(t *testing.T) {
		r := spl2AnalyzeTest(t, `loadjob 'job_${sid}'`)
		for _, ref := range r.References {
			if ref.Kind == "search_job" {
				t.Fatalf("dynamic SID promoted: %+v", r)
			}
		}
	})
	t.Run("timewrap predecessor", func(t *testing.T) {
		r := spl2AnalyzeTest(t, `FROM main | timewrap fortnight`)
		if r.Status != Invalid || !spl2HasCode(r, CodeSyntaxError) || !spl2HasCode(r, CodeUnsupportedSemantics) {
			t.Fatalf("missing prerequisite: %+v", r)
		}
		for _, q := range []string{`FROM main | timechart count() | timewrap fortnight`, `unknown | timewrap fortnight`} {
			r = spl2AnalyzeTest(t, q)
			if r.Status != Incomplete || spl2HasCode(r, CodeSyntaxError) {
				t.Fatalf("unproved/available predecessor: %+v", r)
			}
		}
	})
}

func TestSPL2DeferredRecoveredInputsAndUnionIntentions(t *testing.T) {
	for _, tc := range []struct{ query, name, role string }{
		{`tstats aggregates=[sum(bytes)] byfields=[host source]`, `bytes`, `read`},
		{`FROM main | timechart count() BY host usenull=1`, `host`, `group`},
		{`FROM main | timechart eval(avg(bytes)*avg(duration))`, `bytes`, `read`},
		{`FROM main | timechart eval(avg(bytes)*avg(duration))`, `duration`, `read`},
		{`FROM main | timechart bins= avg(bytes)`, `bytes`, `read`},
	} {
		t.Run(tc.query+tc.name, func(t *testing.T) {
			r := spl2AnalyzeTest(t, tc.query)
			if r.Status != Invalid {
				t.Fatalf("damage promoted: %+v", r)
			}
			spl2Ref(t, r, tc.name, tc.role)
			assertCorpusIntegrity(t, r)
		})
	}
	for _, q := range []string{`mstats aggregates=[avg('cpu.load')] byfields=['host*']`} {
		t.Run(q, func(t *testing.T) {
			r := spl2AnalyzeTest(t, q)
			ref := spl2Ref(t, r, "host*", "group")
			if ref.Binding != "indeterminate" || ref.Resolution != "wildcard" {
				t.Fatalf("wildcard membership invented: %+v", r)
			}
			for _, l := range r.Lineage {
				for _, f := range l.After.Fields {
					if f.Name == "host*" {
						t.Fatalf("pattern installed as field: %+v", r)
					}
				}
			}
		})
	}
	for _, q := range []string{`union main, archive`, `union main`, `FROM main | union other, [FROM archive]`} {
		t.Run(q, func(t *testing.T) {
			r := spl2AnalyzeTest(t, q)
			if len(r.Dependencies.Datasets) == 0 {
				t.Fatalf("named union inputs lost: %+v", r)
			}
			spl2Ref(t, r, "main", "read")
			assertCorpusIntegrity(t, r)
		})
	}
	for _, q := range []string{`FROM main | append [search index=audit | stats count()`, `FROM main | appendcols [FROM other | eval x=1`} {
		t.Run(q, func(t *testing.T) {
			r := spl2AnalyzeTest(t, q)
			if r.Status != Invalid || len(r.Scopes) != 1 {
				t.Fatalf("unclosed original child promoted: %+v", r)
			}
			spl2Ref(t, r, "main", "read")
			assertCorpusIntegrity(t, r)
		})
	}
	control := spl2AnalyzeTest(t, `FROM main | append [FROM child | eval x=1]`)
	if len(control.Scopes) != 2 {
		t.Fatalf("intact child lost: %+v", control)
	}
	spl2Ref(t, control, "x", "create")
}

func TestSPL2DeferredPatternNeverExpandsAndUnionDoesNotInstallRows(t *testing.T) {
	source, err := AnalyzeWithSourceUniverse(QueryDocument{Text: `mstats aggregates=[avg('cpu.load')] byfields=['host*']`, Language: "spl2"}, SourceUniverse{Fields: []string{"cpu.load", "host1"}, Complete: true})
	if err != nil {
		t.Fatal(err)
	}
	ref := spl2Ref(t, source.Result, "host*", "group")
	if ref.Binding != "indeterminate" || ref.Resolution != "wildcard" {
		t.Fatalf("forbidden group expanded: %+v", source)
	}
	for _, lineage := range source.Result.Lineage {
		for _, field := range lineage.After.Fields {
			if field.Name == "host*" || field.Name == "host1" {
				t.Fatalf("forbidden group membership: %+v", source)
			}
		}
	}
	for _, q := range []string{`union [{a:1}],[{b:2}]`, `FROM main | union [{a:1}],[{b:2}]`} {
		r := spl2AnalyzeTest(t, q)
		for _, ref := range r.References {
			if ref.Kind == "field" {
				t.Fatalf("union row shape leaked: %+v", r)
			}
		}
		if len(r.Scopes) != 1 || len(r.Lineage[len(r.Lineage)-1].After.Fields) != 0 {
			t.Fatalf("union row ownership invented: %+v", r)
		}
	}
}
