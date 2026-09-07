package analysis

import (
	"encoding/json"
	"reflect"
	"testing"
)

// Catches incorrect provenance IDs after source sorting and losing literal spelling.
func TestReferencesProvenanceAndLocations(t *testing.T) {
	q := "search src=1 | eval a=src, b=a+1 | stats SUM (b)"
	r, _ := Analyze(QueryDocument{Text: q})
	ids := map[string]Reference{}
	for _, ref := range r.References {
		ids[ref.ID] = ref
		if q[ref.Location.Start.Offset:ref.Location.End.Offset] != ref.OriginalName {
			t.Fatal(ref)
		}
	}
	found := false
	for _, ref := range r.References {
		if ref.Role == "output" && ref.NormalizedName == "sum(b)" {
			found = true
			if ref.OriginalName != "SUM (b)" {
				t.Fatal(ref)
			}
		}
		for _, id := range ref.OriginReferenceIDs {
			if _, ok := ids[id]; !ok {
				t.Fatalf("dangling origin %s", id)
			}
		}
	}
	if !found {
		t.Fatal("missing semantic aggregate output")
	}
	for _, l := range r.Lineage {
		for _, tr := range l.Transitions {
			for _, id := range tr.InputReferenceIDs {
				if _, ok := ids[id]; !ok {
					t.Fatal("dangling transition", tr)
				}
			}
		}
	}
	want, _ := json.Marshal(r)
	for i := 0; i < 10; i++ {
		again, _ := Analyze(QueryDocument{Text: q})
		got, _ := json.Marshal(again)
		if string(got) != string(want) {
			t.Fatal("non-deterministic JSON")
		}
	}
}

func TestReferencesCreationOrigins(t *testing.T) {
	r, _ := Analyze(QueryDocument{Text: "search src=1 | eval a=src, b=a+1"})
	for _, ref := range r.References {
		if ref.Role == "create" && len(ref.OriginReferenceIDs) == 0 {
			t.Fatal("creation has no input provenance", ref)
		}
	}
}
func TestReferencesRecoveryKeepsIntactReads(t *testing.T) {
	r, _ := Analyze(QueryDocument{Text: "search host=web | eval broken= | stats count by user"})
	host, user := false, false
	for _, ref := range r.References {
		host = host || ref.NormalizedName == "host"
		user = user || ref.NormalizedName == "user"
		if ref.NormalizedName == "broken" {
			t.Fatal("damaged assignment fabricated reference")
		}
	}
	if !host || !user || r.Status != Invalid {
		t.Fatal(r)
	}
}

// Catches omitted creation links and incorrect source provenance after ID reassignment.
func TestReferencesExactProvenance(t *testing.T) {
	r, _ := Analyze(QueryDocument{Text: "search src=1 | eval a=src, b=a+1 | table b"})
	want := map[string][]string{"ref-1": {"ref-2", "ref-0"}, "ref-3": {"ref-4", "ref-1", "ref-2", "ref-0"}, "ref-4": {"ref-1", "ref-2", "ref-0"}, "ref-5": {"ref-3", "ref-4", "ref-1", "ref-2", "ref-0"}}
	for id, ids := range want {
		found := false
		for _, ref := range r.References {
			if ref.ID == id {
				found = true
				if !reflect.DeepEqual(ref.OriginReferenceIDs, ids) {
					t.Errorf("%s origins %v want %v", id, ref.OriginReferenceIDs, ids)
				}
			}
		}
		if !found {
			t.Error("missing", id)
		}
	}
}

func TestReferencesUnresolvedRenameWildcard(t *testing.T) {
	r, _ := Analyze(QueryDocument{Text: `search a1=1 | rename a* AS b | where a1=2`})
	found := false
	for _, ref := range r.References {
		if ref.NormalizedName == "a*" && ref.Resolution == "wildcard" {
			found = true
		}
	}
	last := r.References[len(r.References)-1]
	if !found || last.Binding != "indeterminate" {
		t.Fatal("wildcard rename lost reference or retained false provenance", r.References)
	}
}
