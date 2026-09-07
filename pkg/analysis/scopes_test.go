package analysis

import (
	"reflect"
	"testing"
)

// Catches inheriting a branch's derived names or mutating the parent's bindings.
func TestScopesIndependentBranches(t *testing.T) {
	for _, tc := range []struct{ query, kind string }{
		{`search root=1 | join id [ search child=1 | eval local=child ] | where local=2`, "join"},
		{`search root=1 | append [ search child=1 | eval root=child ] | where root=2`, "append"},
		{`search root=1 [ search child=1 | eval local=child ] | where local=2`, "subsearch"},
	} {
		t.Run(tc.kind, func(t *testing.T) {
			r, _ := Analyze(QueryDocument{Text: tc.query})
			if r.Status != Incomplete || !r.Coverage.SyntaxComplete || len(r.Scopes) != 2 {
				t.Fatal(r)
			}
			child := r.Scopes[1]
			if child.Kind != tc.kind || child.ParentID != "scope-0" || child.StageID == "" {
				t.Fatal(child)
			}
			childRead := scopedReference(t, r, "scope-1", "child", "filter")
			if childRead.Binding != "source" {
				t.Fatal(childRead)
			}
			if len(r.Lineage[2].Before.Fields) != 0 && tc.kind != "subsearch" {
				t.Fatal("child inherited parent fields", r.Lineage)
			}
			last := r.References[len(r.References)-1]
			if tc.kind == "append" {
				if last.Binding != "source" || !reflect.DeepEqual(last.OriginReferenceIDs, []string{"ref-0"}) {
					t.Fatal(last)
				}
			} else if last.Binding != "indeterminate" || len(last.OriginReferenceIDs) != 0 {
				t.Fatal(last)
			}
		})
	}
}

// Catches taking appendpipe's inherited copy after merge uncertainty and leaking child removals.
func TestScopesAppendpipeInheritsBeforeMerge(t *testing.T) {
	r, _ := Analyze(QueryDocument{Text: `search root=1 | eval local=root | appendpipe [ where local>1 | fields - local | eval branch=root ] | where local=2`})
	if r.Status != Incomplete || !r.Coverage.SyntaxComplete {
		t.Fatal(r)
	}
	child := scopedReference(t, r, "scope-1", "local", "read")
	if child.Binding != "derived" {
		t.Fatal(child)
	}
	if r.Lineage[3].Before.Uncertain {
		t.Fatal("child inherited merge uncertainty", r.Lineage[3])
	}
	last := r.References[len(r.References)-1]
	if last.Binding != "derived" || !reflect.DeepEqual(last.OriginReferenceIDs, child.OriginReferenceIDs) {
		t.Fatal(last, child)
	}
	for _, f := range r.Lineage[len(r.Lineage)-1].After.Fields {
		if f.Name == "branch" {
			t.Fatal("child output leaked", f)
		}
	}
}
func scopedReference(t *testing.T, r *Result, scope, name, role string) Reference {
	t.Helper()
	for _, ref := range r.References {
		if ref.ScopeID == scope && ref.NormalizedName == name && ref.Role == role {
			return ref
		}
	}
	t.Fatalf("missing %s %s %s: %+v", scope, name, role, r.References)
	return Reference{}
}

// Catches global counters used for scope positions and misattributed nested dependencies.
func TestScopesNestedOwnership(t *testing.T) {
	r, _ := Analyze(QueryDocument{Text: `search index=z | append [ search index=b | appendpipe [ search index=a ] ] | head 2`})
	if len(r.Scopes) != 3 || len(r.Stages) != 6 {
		t.Fatal(r)
	}
	if r.Scopes[2].ParentID != "scope-1" || r.Scopes[2].StageID != "stage-3" || r.Scopes[2].Kind != "appendpipe" {
		t.Fatal(r.Scopes)
	}
	if !reflect.DeepEqual(r.Dependencies.Indexes, []string{"a", "b", "z"}) {
		t.Fatal(r.Dependencies)
	}
	for i, want := range []int{0, 1, 0, 1, 0, 2} {
		if r.Stages[i].Position != want {
			t.Fatal(r.Stages)
		}
	}
	scopedReference(t, r, "scope-2", "a", "read")
}
