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

// ANTLR can recover child stages as root siblings when it loses a bracketed
// subtree. Such stages must not publish parent outputs, references, or lineage.
func TestScopesMalformedChildCannotBecomeParent(t *testing.T) {
	for _, tc := range []struct{ name, query string }{
		{"append", `search host=web | append [ search child=1 | eval good=child, broken= ] | where good=2`},
		{"join", `search host=web | join id [ search child=1 | eval good=child, broken= ] | where good=2`},
		{"subsearch", `search host=web [ search child=1 | eval good=child, broken= ] | where good=2`},
		{"appendpipe", `search host=web | appendpipe [ search child=1 | eval good=child, broken= ] | where good=2`},
		{"nested", `search host=web | append [ search outer=1 | append [ search child=1 | eval good=child, broken= ] ] | where good=2`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r, err := Analyze(QueryDocument{Text: tc.query})
			if err != nil || r.Status != Invalid {
				t.Fatal(r, err)
			}
			for _, ref := range r.References {
				if ref.ScopeID == "scope-0" && (ref.NormalizedName == "child" || (ref.NormalizedName == "good" && ref.Role != "read")) {
					t.Error("misowned child reference", ref)
				}
			}
			for _, line := range r.Lineage {
				if line.ScopeID == "scope-0" {
					for _, state := range []FieldState{line.Before, line.After} {
						for _, field := range state.Fields {
							if field.Name == "good" || field.Name == "child" {
								t.Error("child output leaked to parent state", line)
							}
						}
					}
				}
			}
			last := scopedReference(t, r, "scope-0", "good", "read")
			if last.Binding != "indeterminate" || len(last.OriginReferenceIDs) != 0 {
				t.Error("parent acquired child provenance", last)
			}
		})
	}
}
