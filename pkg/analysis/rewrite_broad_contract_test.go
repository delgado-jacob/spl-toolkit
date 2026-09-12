package analysis

import (
	"reflect"
	"strings"
	"testing"
)

// HAVING must contribute only at its logical phase, using the same restricted
// source/derived visibility as its public references.
func TestRewriteSQLHavingFacts(t *testing.T) {
	for _, fromFirst := range []bool{false, true} {
		for _, tc := range []struct {
			name, selectText, where, group, having, field, value string
			complete                                             bool
		}{
			{"having only", "marker", "", " GROUP BY marker", "marker=1", "marker", "1", true},
			{"compatible", "marker", " WHERE marker=1", " GROUP BY marker", "marker=1", "marker", "1", true},
			{"conflicting", "marker", " WHERE marker=1", " GROUP BY marker", "marker=2", "marker", "", true},
			{"conflict with hidden operand", "marker", " WHERE marker=1", " GROUP BY marker", "marker=2 AND hidden=1", "marker", "", false},
			{"no grouping", "marker", "", "", "marker=1", "marker", "", false},
			{"hidden", "marker", "", " GROUP BY marker", "hidden=1", "hidden", "", false},
			{"derived", "count() AS n", "", "", "n=1", "n", "", false},
			{"unsupported", "marker", "", " GROUP BY marker", "unknown(marker)=1", "marker", "", false},
		} {
			t.Run(tc.name+map[bool]string{false: "/select-first", true: "/from-first"}[fromFirst], func(t *testing.T) {
				query := "SELECT " + tc.selectText + " FROM main" + tc.where + tc.group
				if fromFirst {
					query = "FROM main" + tc.where + tc.group + " SELECT " + tc.selectText
				}
				query += " HAVING " + tc.having + " | lookup people marker OUTPUT label"
				session := rewriteTestSession(t, "spl2", query, RewriteFactProbe{Kind: "field", Identity: rewriteName(tc.field)})
				fact := rewriteFind(t, session, "lookup", "people", 0).Facts[0]
				count := 0
				if tc.value != "" {
					count = 1
				}
				if fact.LiteralComplete != tc.complete || len(fact.GuaranteedValues) != count || (count == 1 && string(fact.GuaranteedValues[0].Value) != tc.value) {
					t.Fatalf("HAVING authorization boundary: %+v", fact)
				}
				if tc.complete {
					havingOffset := strings.Index(query, "HAVING ") + len("HAVING ")
					found := false
					for _, loc := range fact.Locations {
						found = found || loc.Start.Offset == havingOffset
					}
					if !found {
						t.Fatalf("HAVING predicate evidence missing: %+v", fact)
					}
				}
			})
		}
	}
	query := "SELECT marker FROM main GROUP BY marker HAVING marker=1 ORDER BY marker | lookup people marker OUTPUT label"
	session := rewriteTestSession(t, "spl2", query, RewriteFactProbe{Kind: "field", Identity: rewriteName("marker")})
	for _, site := range session.Evidence().Sites {
		if site.Kind != "field" || site.Identity.Name == nil || *site.Identity.Name != "marker" {
			continue
		}
		want := 0
		if site.Point.Phase == "having" || site.Point.Phase == "order" || site.Point.Phase == "" {
			want = 1
		}
		if len(site.Facts[0].GuaranteedValues) != want {
			t.Errorf("fact leaked across phase %s: %+v", site.Point.Phase, site.Facts[0])
		}
	}
	phases := []string{}
	for _, item := range session.Evidence().Analysis.Lineage {
		phases = append(phases, item.Phase)
	}
	if !reflect.DeepEqual(phases, []string{"source", "group", "evaluate", "having", "order", "project", ""}) {
		t.Fatalf("SQL logical phase order changed: %v", phases)
	}
	if len(session.Evidence().Analysis.Diagnostics) != 0 {
		t.Fatal("supported HAVING gained diagnostics")
	}
}

// A finite-size accepted numeric token must remain equal to itself and its exact
// decimal equivalents; exponent magnitude is not contradictory evidence.
func TestRewriteLargeExponentFacts(t *testing.T) {
	for _, language := range []string{"spl", "spl2"} {
		pairs := []struct {
			left, right string
			equal       bool
		}{
			{"1e1000001", "1e1000001", true},
			{"1e1000001", "10e1000000", true},
			{"1e1000001", "2e1000001", false},
			{"1e-1000001", "10e-1000002", true},
			{"1e-1000001", "1e-1000001", true},
			{"1e-1000001", "1e-1000002", false},
			{"1e9223372036854775809", "10e9223372036854775808", true},
			{"1e9223372036854775809", "1e9223372036854775809", true},
			{"1e9223372036854775809", "1e9223372036854775810", false},
			{"1e-9223372036854775809", "10e-9223372036854775810", true},
			{"1e-9223372036854775809", "1e-9223372036854775809", true},
			{"1e-9223372036854775809", "2e-9223372036854775809", false},
		}
		if language == "spl" {
			// SPL's numeric grammar has no exponent token. Keep its existing
			// accepted decimal equivalents on the shared numeric merge path.
			pairs = []struct {
				left, right string
				equal       bool
			}{
				{"900719925474099300", "900719925474099300.0", true},
				{"900719925474099300", "900719925474099301", false},
			}
		}
		for _, pair := range pairs {
			for _, op := range []string{"AND", "OR"} {
				t.Run(language+"/"+pair.left+"/"+op+"/"+pair.right, func(t *testing.T) {
					query := "search index=main | where n=" + pair.left + " " + op + " n=" + pair.right + " | table src"
					session := rewriteTestSession(t, language, query, RewriteFactProbe{Kind: "field", Identity: rewriteName("n")})
					fact := rewriteFind(t, session, "field", "src", 0).Facts[0]
					want := 0
					if pair.equal {
						want = 1
					}
					if !fact.LiteralComplete || len(fact.GuaranteedValues) != want || (want == 1 && string(fact.GuaranteedValues[0].Value) != pair.left) {
						t.Fatalf("exact numeric merge lost: %+v", fact)
					}
				})
			}
		}
	}
}
