package analysis

import (
	"reflect"
	"strings"
	"testing"
)

// Pattern stars must match literal stars in known field names, and typed quoted
// fragments must normalize without turning neighboring selectors into one name.
func TestSelectorClosedMembership(t *testing.T) {
	for _, tc := range []struct {
		name, field, selector, pattern string
		matches                        bool
	}{
		{"literal_star_prefix", "a*b", "a*", "a*", true},
		{"literal_star_suffix", "*ba", "*a", "*a", true},
		{"literal_star_backtracking", "a*bbbc", "'a*bc'", "a*bc", true},
		{"suffix_mismatch", "a*b", "'a*c'", "a*c", false},
		{"ordinary_prefix", "ab", "a*", "a*", true},
		{"empty_star_match", "a", "a*", "a*", true},
		{"quoted_prefix", "ab", "'a'*", "a*", true},
		{"quoted_suffix", "first name", "*'name'", "*name", true},
		{"quoted_both", "first name last", "*'name'*", "*name*", true},
		{"quoted_escaped_fragment", "a'b-tail", "'a\\'b'*", "a'b*", true},
		{"quoted_unicode", "café-tail", "'café'*", "café*", true},
		{"fully_quoted_pattern", "ab", "'a*'", "a*", true},
	} {
		for _, exclude := range []bool{false, true} {
			mode, operand, role, binding := "include", tc.selector, "read", "indeterminate"
			if exclude {
				mode, operand, role, binding = "exclude", "- "+tc.selector, "remove", "not_applicable"
			}
			t.Run(tc.name+"/"+mode, func(t *testing.T) {
				// Use an exact quoted expression identifier even if the field has '*'.
				quoted := "'" + stringsForQuotedField(tc.field) + "'"
				q := "| stats count AS " + quoted + " | fields " + operand + " | where " + quoted + "=1"
				r, err := Analyze(QueryDocument{Text: q})
				if err != nil {
					t.Fatal(err)
				}
				retained := tc.matches != exclude
				wantStatus, wantBinding := Valid, "derived"
				if !retained {
					wantStatus, wantBinding = Invalid, "unavailable"
				}
				if r.Status != wantStatus || !r.Coverage.SyntaxComplete || !r.Coverage.SemanticComplete {
					t.Errorf("status/coverage = %s %+v; want %s/complete", r.Status, r.Coverage, wantStatus)
				}
				if len(r.References) != 3 {
					t.Fatalf("want create, selector, final read; got %+v", r.References)
				}
				sel, last := r.References[1], r.References[2]
				if sel.OriginalName != tc.selector || sel.NormalizedName != tc.pattern || sel.Resolution != "wildcard" || sel.Role != role || sel.Binding != binding {
					t.Errorf("selector = %+v", sel)
				}
				if last.NormalizedName != tc.field || last.Resolution != "exact" || last.Binding != wantBinding {
					t.Errorf("final read = %+v; want exact/%s", last, wantBinding)
				}
				origins := []string{}
				if tc.matches {
					origins = []string{"ref-0"}
				}
				if !reflect.DeepEqual(sel.OriginReferenceIDs, origins) {
					t.Errorf("selector origins = %v; want %v", sel.OriginReferenceIDs, origins)
				}
				final := r.Lineage[len(r.Lineage)-1].After
				if retained {
					if len(final.Fields) != 1 || final.Fields[0].Name != tc.field || !reflect.DeepEqual(final.Fields[0].OriginReferenceIDs, []string{"ref-0"}) || !reflect.DeepEqual(last.OriginReferenceIDs, []string{"ref-0"}) || len(r.Diagnostics) != 0 {
						t.Errorf("retained field/provenance = %+v; final read %+v", final, last)
					}
				} else if len(final.Fields) != 0 || len(last.OriginReferenceIDs) != 0 || len(r.Diagnostics) != 1 || r.Diagnostics[0].Code != CodeUnavailableField || r.Diagnostics[0].Location != last.Location {
					t.Errorf("absence must apply only to final read: %+v %+v", final, r.Diagnostics)
				}
				assertCorpusIntegrity(t, r)
			})
		}
	}
}

func stringsForQuotedField(name string) string {
	// Test input quoting only; expected names/patterns above are handwritten.
	return strings.ReplaceAll(strings.ReplaceAll(name, "\\", "\\\\"), "'", "\\'")
}

func TestSelectorQuotedFragmentBoundaries(t *testing.T) {
	r, err := Analyze(QueryDocument{Text: "| stats count AS ab count AS keep | fields 'a'* keep | where ab=1 AND keep=1"})
	if err != nil {
		t.Fatal(err)
	}
	if r.Status != Valid || len(r.References) != 6 {
		t.Fatalf("neighbor selector lost: %+v", r)
	}
	for i, want := range []struct{ original, normalized, resolution, binding string }{
		{"'a'*", "a*", "wildcard", "indeterminate"},
		{"keep", "keep", "exact", "derived"},
	} {
		ref := r.References[i+2]
		if ref.OriginalName != want.original || ref.NormalizedName != want.normalized || ref.Resolution != want.resolution || ref.Binding != want.binding {
			t.Errorf("neighbor selector = %+v; want %+v", ref, want)
		}
	}
	if !reflect.DeepEqual(r.References[4].OriginReferenceIDs, []string{"ref-0"}) || !reflect.DeepEqual(r.References[5].OriginReferenceIDs, []string{"ref-1"}) {
		t.Error("neighbor origins mixed", r.References)
	}
	assertCorpusIntegrity(t, r)
}
