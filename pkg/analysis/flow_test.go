package analysis

import (
	"reflect"
	"testing"
)

// Catches incorrect source/derived classification and missing destructive transfers.
func TestFlowRepresentativePipelines(t *testing.T) {
	for _, tc := range []struct {
		query         string
		status        Status
		reads, fields []string
		open          bool
	}{
		{"search host=web | where src=dest", Valid, []string{"host:source", "src:source", "dest:source"}, []string{"dest", "host", "src"}, true},
		{"search src=1 | eval a=src, b=a+1 | table b", Valid, []string{"src:source", "src:source", "a:derived", "b:derived"}, []string{"b"}, false},
		{"search a=1 | rename a AS b | where a=2", Invalid, []string{"a:source", "a:source", "a:unavailable"}, []string{"b"}, true},
		{"search a=1 b=2 | fields - a | where a=3", Invalid, []string{"a:source", "b:source", "a:unavailable"}, []string{"b"}, true},
		{"search a=1 b=2 | stats sum(a) AS total BY b | where a>1", Invalid, []string{"a:source", "b:source", "a:source", "b:source", "a:unavailable"}, []string{"b", "total"}, false},
		{"search a=1 | eval a=a+1 | where a>1", Valid, []string{"a:source", "a:source", "a:derived"}, []string{"a"}, true},
		{"search a=1 | eval b=if(a>0,lower(c),d)", Valid, []string{"a:source", "a:source", "c:source", "d:source"}, []string{"a", "b", "c", "d"}, true},
		{"search a=1 | lookup users key AS a OUTPUT name AS user", Valid, []string{"a:source", "a:source"}, []string{"a", "user"}, true},
		{"search a=1 | lookup users a OUTPUTNEW user | where user=x", Valid, []string{"a:source", "a:source", "user:indeterminate", "x:source"}, []string{"a", "user", "x"}, true},
		{"search a=1 | stats SUM (a) | table 'sum(a)'", Valid, []string{"a:source", "a:source", "sum(a):derived"}, []string{"sum(a)"}, false},
		{"search Host=web | sort -Host | dedup Host | head 3 | tail 1", Valid, []string{"Host:source", "Host:source", "Host:source"}, []string{"Host"}, true},
		{"search a=1 | eventstats sum(a) AS total | streamstats count AS n", Valid, []string{"a:source", "a:source"}, []string{"a", "n", "total"}, true},
	} {
		t.Run(tc.query, func(t *testing.T) {
			r, e := Analyze(QueryDocument{Text: tc.query})
			if e != nil {
				t.Fatal(e)
			}
			var reads []string
			for _, ref := range r.References {
				if ref.Kind == "field" && ref.Binding != "not_applicable" {
					reads = append(reads, ref.NormalizedName+":"+ref.Binding)
				}
			}
			if !reflect.DeepEqual(reads, tc.reads) {
				t.Errorf("reads = %v; want %v", reads, tc.reads)
			}
			if r.Status != tc.status {
				t.Errorf("status %s want %s: %+v", r.Status, tc.status, r.Diagnostics)
			}
			if len(r.Lineage) == 0 {
				t.Fatal("missing field lineage")
			}
			state := r.Lineage[len(r.Lineage)-1].After
			names := []string{}
			for _, f := range state.Fields {
				names = append(names, f.Name)
			}
			if !reflect.DeepEqual(names, tc.fields) || state.Open != tc.open {
				t.Errorf("final state %+v; want fields %v open %v", state, tc.fields, tc.open)
			}
		})
	}
}

// Catches unknown-effect identity, OUTPUTNEW overwrites, wildcard guessing and rename sequencing.
func TestFlowConservativeTransfers(t *testing.T) {
	for _, tc := range []struct {
		q      string
		status Status
		last   string
	}{
		{"search a=1 | fields - a | mystery | where a=2", Incomplete, "a:indeterminate"},
		{"search a=1 | fields - a | where a=2 | mystery", Invalid, "a:unavailable"},
		{"search a=1 | lookup users a OUTPUTNEW a | where a=2", Valid, "a:source"},
		{"search a=1 b=2 | rename a AS b,b AS a | where a=b", Incomplete, "b:indeterminate"},
		{"search a=1 b=2 | rename a AS c,b AS c | where c=1", Incomplete, "c:indeterminate"},
		{"search a=1 b=2 | table a b | fields a* | where b=1", Invalid, "b:unavailable"},
		{"search a=1 | fields a* | where a=1", Incomplete, "a:source"},
		{"search a=1 | table a | fields - * | where a=1", Invalid, "a:unavailable"},
		{"search a=1 | table a | inputlookup users | where z=1", Valid, "z:source"},
	} {
		t.Run(tc.q, func(t *testing.T) {
			r, _ := Analyze(QueryDocument{Text: tc.q})
			if r.Status != tc.status {
				t.Fatalf("status %s want %s: %+v", r.Status, tc.status, r.Diagnostics)
			}
			last := ""
			for _, ref := range r.References {
				if ref.Binding != "not_applicable" {
					last = ref.NormalizedName + ":" + ref.Binding
				}
			}
			if last != tc.last {
				t.Fatalf("last read %s want %s", last, tc.last)
			}
		})
	}
}

// Explicit modeled outputs after an opaque stage still establish new field bindings.
func TestFlowKnownOutputsAfterUncertainty(t *testing.T) {
	for _, q := range []string{`search a=1 | mystery | eval x=1 | where x>0`, `search a=1 | mystery | stats count AS x | where x>0`, `search a=1 | mystery | lookup users a OUTPUT x | where x>0`} {
		r, _ := Analyze(QueryDocument{Text: q})
		last := r.References[len(r.References)-1]
		if last.NormalizedName != "x" || last.Binding != "derived" {
			t.Errorf("%s: %+v", q, last)
		}
	}
}

// SPL fields inclusion retains internal fields; it cannot assert unknown internals absent.
func TestFlowFieldsInternalMembership(t *testing.T) {
	for _, tc := range []struct {
		q       string
		status  Status
		binding string
		removed bool
	}{
		{`search host=x | fields host | where _time>0`, Incomplete, "indeterminate", false},
		{`search host=x _time=1 | fields host | where _time>0`, Incomplete, "source", false},
		{`search host=x | fields - _time | fields host | where _time>0`, Incomplete, "indeterminate", true},
		{`search host=x | table host | fields host | where _time>0`, Invalid, "unavailable", false},
		{`search host=x _time=1 | table host _time | fields host | where _time>0`, Valid, "source", false},
	} {
		r, _ := Analyze(QueryDocument{Text: tc.q})
		last := r.References[len(r.References)-1]
		if r.Status != tc.status || last.NormalizedName != "_time" || last.Binding != tc.binding {
			t.Errorf("%s: %s %+v", tc.q, r.Status, last)
		}
		if tc.removed {
			found := false
			for _, name := range r.Lineage[len(r.Lineage)-1].After.Removed {
				found = found || name == "_time"
			}
			if !found {
				t.Fatal("removed internal tombstone lost")
			}
		}
	}
}
