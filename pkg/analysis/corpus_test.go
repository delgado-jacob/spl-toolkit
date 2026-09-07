package analysis

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"testing"
)

type corpusCase struct {
	ID       string        `json:"id"`
	Document QueryDocument `json:"document"`
	Expected *Result       `json:"expected"`
}

func loadCorpus(t *testing.T) []corpusCase {
	t.Helper()
	data, err := os.ReadFile("../../testdata/analysis/cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var corpus struct {
		Version string       `json:"version"`
		Cases   []corpusCase `json:"cases"`
	}
	if err := json.Unmarshal(data, &corpus); err != nil {
		t.Fatal(err)
	}
	if corpus.Version != "1" || len(corpus.Cases) != 26 {
		t.Fatal("missing reviewed corpus", corpus.Version, len(corpus.Cases))
	}
	return corpus.Cases
}

// Hand-derived semantic oracles precede the full-report equality gate. They catch
// missing/reclassified reads, literals treated as fields, leaked scope outputs,
// lost projections, and incorrect status precedence independently of the goldens.
func TestCorpusSemanticReview(t *testing.T) {
	type oracle struct {
		status          Status
		names, fields   string
		stages, scopes  int
		open, uncertain bool
	}
	want := map[string]oracle{
		"unicode_repeated":          {Valid, "café,café,café,n,café,café,n", "café,n", 3, 1, false, false},
		"comparisons":               {Valid, "host,status,host,other,status,limit", "host,limit,other,status", 2, 1, true, false},
		"functions_assignments":     {Valid, "a,src,b,a,other,c,b,src,b", "a,b,c,other,src", 1, 1, true, false},
		"transfer_family":           {Valid, "a,b,a,c,b,c,c,c", "c", 8, 1, false, false},
		"aggregate_family":          {Valid, "bytes,user,bytes,total,user,n,total,all,user", "all,user", 4, 1, false, false},
		"lookup_conditional":        {Valid, "uid,people,uid,display,group,uid,display,group", "display,group,uid", 3, 1, false, false},
		"inputlookup":               {Valid, "users,id", "id", 2, 1, true, false},
		"wildcard_closed":           {Valid, "a1,a2,a1,a2,a*", "a1,a2", 3, 1, false, false},
		"wildcard_open":             {Incomplete, "a1,a*", "a1", 2, 1, false, true},
		"unknown_function":          {Incomplete, "host,out,host", "host,out", 2, 1, true, true},
		"macro":                     {Incomplete, "host,expand,count,user", "count,user", 3, 1, false, false},
		"join_scope":                {Incomplete, "root,child,local,child,local", "root", 5, 2, true, true},
		"appendpipe_scope":          {Incomplete, "root,local,root,local,local,local", "local,root", 6, 2, true, true},
		"partial_recovery":          {Invalid, "host,good,host,count,user", "count,user", 3, 1, false, false},
		"invalid_precedence":        {Invalid, "a,a,a", "", 4, 1, true, true},
		"dependencies":              {Incomplete, "main,/var/log/a,syslog,Network_Traffic,Network_Traffic.All_Traffic,Web,Web.All_Traffic,Authentication,Authentication.Authentication,users", "", 8, 4, true, false},
		"saved_dataset":             {Incomplete, "savedsearch:Daily", "", 1, 1, true, true},
		"quoted_asterisk":           {Valid, "a*,a*", "a*", 2, 1, true, false},
		"wildcard_exclusion_closed": {Valid, "a1,a2,keep,a1,a2,keep,a*,z*", "keep", 3, 1, false, false},
		"wildcard_exclusion_open":   {Incomplete, "a1,keep,a*,z*", "keep", 2, 1, true, true},
		"malformed_child_scope":     {Invalid, "host,good", "host", 3, 1, true, true},
		"quoted_projection":         {Incomplete, "a*,ab,a*", "a*,ab", 2, 1, false, true},
		"implicit_aggregate":        {Valid, "bytes,sum(bytes),bytes", "sum(bytes)", 2, 1, false, false},
		"exact_after_unknown":       {Invalid, "a,a,missing", "a", 4, 1, false, false},
		"fields_internal":           {Valid, "_time,a,b,_time,a,b,a,_time", "_time,a", 4, 1, false, false},
		"fields_open":               {Incomplete, "a,a,b", "a", 3, 1, false, true},
	}
	for _, c := range loadCorpus(t) {
		t.Run(c.ID, func(t *testing.T) {
			w, ok := want[c.ID]
			if !ok {
				t.Fatal("no handwritten oracle")
			}
			r, err := Analyze(c.Document)
			if err != nil {
				t.Fatal(err)
			}
			names := []string{}
			for _, ref := range r.References {
				names = append(names, ref.NormalizedName)
			}
			fields := []string{}
			last := r.Lineage[len(r.Lineage)-1].After
			for _, f := range last.Fields {
				fields = append(fields, f.Name)
			}
			if r.Status != w.status || strings.Join(names, ",") != w.names || strings.Join(fields, ",") != w.fields || len(r.Stages) != w.stages || len(r.Scopes) != w.scopes || last.Open != w.open || last.Uncertain != w.uncertain {
				t.Fatalf("status=%s refs=%q fields=%q stages=%d scopes=%d open=%v uncertain=%v", r.Status, strings.Join(names, ","), strings.Join(fields, ","), len(r.Stages), len(r.Scopes), last.Open, last.Uncertain)
			}
			if c.ID == "wildcard_exclusion_closed" || c.ID == "wildcard_exclusion_open" {
				for _, ref := range r.References[len(r.References)-2:] {
					if ref.Role != "remove" || ref.Binding != "not_applicable" || ref.Resolution != "wildcard" {
						t.Fatal("wildcard exclusion became consuming obligation", ref)
					}
				}
				for _, diag := range r.Diagnostics {
					if diag.Code == CodeUnavailableField {
						t.Fatal("exclusion invented unavailable field", diag)
					}
				}
			}
			if c.ID == "malformed_child_scope" {
				ref := r.References[len(r.References)-1]
				if ref.Binding != "indeterminate" || len(ref.OriginReferenceIDs) != 0 {
					t.Fatal("malformed child became parent provenance", ref)
				}
			}
			if c.ID == "unicode_repeated" {
				ref := r.References[1]
				if ref.OriginalName != "'café'" || ref.Location.Start.Line != 2 || ref.Location.Start.Column != 8 || ref.Location.Start.Offset != 26 {
					t.Fatal(ref)
				}
			}
			if c.ID == "lookup_conditional" {
				if !last.Fields[1].Conditional || r.References[7].Binding != "indeterminate" {
					t.Fatal("lost OUTPUTNEW conditionality", last, r.References[7])
				}
			}
			if c.ID == "unknown_function" {
				if !last.Fields[1].Conditional || r.Coverage.SemanticComplete {
					t.Fatal("unknown function invented certainty", r)
				}
			}
			if c.ID == "partial_recovery" || c.ID == "macro" {
				if !last.Fields[1].Conditional || r.Coverage.SemanticComplete {
					t.Fatal("later exact aggregation erased earlier incompleteness", r)
				}
			}
			if c.ID == "quoted_asterisk" {
				for _, ref := range r.References {
					if ref.Resolution != "exact" {
						t.Fatal(ref)
					}
				}
			}
			if c.ID == "wildcard_closed" || c.ID == "wildcard_open" {
				if r.References[len(r.References)-1].Resolution != "wildcard" {
					t.Fatal(r.References)
				}
			}
		})
	}
}

// Encoded full reports are the cross-adapter contract; concurrency and source
// assertions additionally detect unstable IDs and byte/rune location confusion.
func TestCorpusFullReports(t *testing.T) {
	for _, c := range loadCorpus(t) {
		t.Run(c.ID, func(t *testing.T) {
			if c.Expected == nil {
				t.Fatal("full report not reviewed")
			}
			expected, _ := json.Marshal(c.Expected)
			var wg sync.WaitGroup
			for i := 0; i < 6; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					r, err := Analyze(c.Document)
					if err != nil {
						t.Error(err)
						return
					}
					got, err := json.Marshal(r)
					if err != nil {
						t.Error(err)
						return
					}
					if !bytes.Equal(got, expected) {
						t.Errorf("full report differs for %s", c.ID)
					}
					assertCorpusIntegrity(t, r)
				}()
			}
			wg.Wait()
		})
	}
}
func assertCorpusIntegrity(t *testing.T, r *Result) {
	t.Helper()
	ids := map[string]bool{}
	for i, ref := range r.References {
		if ref.ID != fmt.Sprintf("ref-%d", i) {
			t.Error("unstable reference ID", ref)
		}
		ids[ref.ID] = true
		loc := ref.Location
		if loc.Start.Offset < 0 || loc.End.Offset > len(r.Document.Text) || loc.End.Offset < loc.Start.Offset {
			t.Error("out of bounds", ref)
			continue
		}
		if r.Document.Text[loc.Start.Offset:loc.End.Offset] != ref.OriginalName {
			t.Error("source mismatch", ref)
		}
	}
	checkIDs := func(origins []string) {
		for _, id := range origins {
			if !ids[id] {
				t.Error("dangling provenance", id)
			}
		}
	}
	for _, ref := range r.References {
		checkIDs(ref.OriginReferenceIDs)
	}
	for _, line := range r.Lineage {
		for _, state := range []FieldState{line.Before, line.After} {
			if state.Fields == nil || state.Removed == nil {
				t.Error("null field state")
			}
			names := []string{}
			for _, f := range state.Fields {
				names = append(names, f.Name)
				checkIDs(f.OriginReferenceIDs)
			}
			if !sort.StringsAreSorted(names) {
				t.Error("unordered fields")
			}
		}
		for _, tr := range line.Transitions {
			checkIDs(tr.InputReferenceIDs)
			if tr.OutputReferenceID != "" {
				checkIDs([]string{tr.OutputReferenceID})
			}
		}
	}
	// Walk decoded values so every collection, including nested origins, is [] and never null.
	data, _ := json.Marshal(r)
	var value any
	_ = json.Unmarshal(data, &value)
	var walk func(any)
	walk = func(v any) {
		switch n := v.(type) {
		case nil:
			t.Error("null report value")
		case map[string]any:
			for _, x := range n {
				walk(x)
			}
		case []any:
			for _, x := range n {
				walk(x)
			}
		}
	}
	walk(value)
}

// Truncations exercise parser recovery through every rune boundary without
// hiding panics; unknown effects must never produce complete coverage.
func TestCorpusTruncatedInputs(t *testing.T) {
	for _, q := range []string{`search host=web | eval a=coalesce(host,"é"), b=1 | append [ search child=2 ]`, "search * | `expand(host)` | stats count by user"} {
		runes := []rune(q)
		for i := 0; i < len(runes); i++ {
			r, err := Analyze(QueryDocument{Text: string(runes[:i])})
			if err != nil {
				t.Fatal(err)
			}
			assertCorpusIntegrity(t, r)
		}
	}
	for _, q := range []string{`search * | mystery x`, `search * | eval a=unknownpure(host)`, `search * | macro host`} {
		r, _ := Analyze(QueryDocument{Text: q})
		if r.Status != Incomplete || r.Coverage.SemanticComplete {
			t.Fatal(r)
		}
	}
}
