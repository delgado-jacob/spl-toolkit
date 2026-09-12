package analysis

import (
	"strings"
	"testing"
)

// A dependency mapping names the selected value; a literal condition names
// the grammar selector. Conflating either namespace loses real authorization.
func TestRewriteDependencyLiteralSelectorIdentity(t *testing.T) {
	for _, language := range []string{"spl", "spl2"} {
		for _, selector := range []struct{ kind, value string }{{"index", "main"}, {"source", "app"}, {"sourcetype", "auth"}} {
			t.Run(language+"/"+selector.kind, func(t *testing.T) {
				query := "search " + selector.kind + "=" + selector.value + " src=alice | table src"
				s := rewriteTestSession(t, language, query,
					RewriteFactProbe{Kind: selector.kind, Identity: rewriteName(selector.kind)},
					RewriteFactProbe{Kind: selector.kind, Identity: rewriteName(selector.value)},
					RewriteFactProbe{Kind: "field", Identity: rewriteName(selector.kind)})
				dependency := rewriteFind(t, s, selector.kind, selector.value, 0)
				if query[dependency.Location.Start.Offset:dependency.Location.End.Offset] != selector.value {
					t.Fatalf("dependency mapping lost value ownership: %+v", dependency)
				}
				for occurrence := 0; occurrence < 2; occurrence++ {
					facts := rewriteFind(t, s, "field", "src", occurrence).Facts
					fact := facts[0]
					if !fact.LiteralComplete || fact.ReferenceState != "true" || len(fact.GuaranteedValues) != 1 || fact.GuaranteedValues[0].Kind != "string" || string(fact.GuaranteedValues[0].Value) != `"`+selector.value+`"` || len(fact.ReferenceIDs) != 1 || fact.ReferenceIDs[0] != dependency.ReferenceID || len(fact.Locations) != 1 || fact.Locations[0] != dependency.Location {
						t.Errorf("selector literal lost value/location/reference evidence: %+v", fact)
					}
					for _, absent := range facts[1:] {
						if !absent.LiteralComplete || len(absent.GuaranteedValues) != 0 || len(absent.Locations) != 0 {
							t.Errorf("selector fact leaked into another typed identity: %+v", absent)
						}
					}
				}
			})
		}
		query := `search src=x | where sourcetype="auth" | table src`
		s := rewriteTestSession(t, language, query,
			RewriteFactProbe{Kind: "field", Identity: rewriteName("sourcetype")},
			RewriteFactProbe{Kind: "sourcetype", Identity: rewriteName("sourcetype")})
		facts := rewriteFind(t, s, "field", "src", 1).Facts
		if len(facts[0].GuaranteedValues) != 1 || len(facts[1].GuaranteedValues) != 0 || !facts[1].LiteralComplete {
			t.Fatalf("expression field reclassified by spelling: %+v", facts)
		}
	}
}

// Contradictory static restrictions are not event-cardinality assumptions or
// vacuous proof. Complete supported absence is false; unobserved effects remain
// unknown. Repeated compatible values and other identities retain their proof.
func TestRewriteConflictingLiteralGuarantees(t *testing.T) {
	cases := []struct {
		name, predicate, value string
		complete               bool
	}{
		{"different", `EventCode=1 AND EventCode=2`, "", true},
		{"typed different", `EventCode=1 AND EventCode="1"`, "", true},
		{"repeated after conflict", `EventCode=1 AND EventCode=2 AND EventCode=1`, "", true},
		{"common OR with conflict", `(EventCode=1 AND EventCode=2) OR EventCode=1`, "", true},
		{"later AND after OR", `((EventCode=1 AND EventCode=2) OR tag="x") AND EventCode=1`, "", true},
		{"conflict plus unobserved", `EventCode=1 AND EventCode=2 AND unknown(other)=3`, "", false},
		{"compatible repeated", `EventCode=1 AND EventCode=1`, "1", true},
		{"compatible numeric", `EventCode=1 AND EventCode=1.0`, "1", true},
		{"noncommon OR", `EventCode=1 OR EventCode=2`, "", true},
		{"NOT", `NOT EventCode=1`, "", true},
	}
	for _, language := range []string{"spl", "spl2"} {
		for _, tc := range cases {
			t.Run(language+"/"+tc.name, func(t *testing.T) {
				query := `search src=x | where (` + tc.predicate + `) AND flag="kept" | table src`
				s := rewriteTestSession(t, language, query,
					RewriteFactProbe{Kind: "field", Identity: rewriteName("EventCode")},
					RewriteFactProbe{Kind: "field", Identity: rewriteName("flag")})
				facts := rewriteFind(t, s, "field", "src", 1).Facts
				fact := facts[0]
				wantCount := 0
				if tc.value != "" {
					wantCount = 1
				}
				if fact.LiteralComplete != tc.complete || len(fact.GuaranteedValues) != wantCount || (wantCount == 1 && string(fact.GuaranteedValues[0].Value) != tc.value) {
					t.Errorf("incompatible/compatible proof boundary lost: %+v", fact)
				}
				if len(facts[1].GuaranteedValues) != 1 || string(facts[1].GuaranteedValues[0].Value) != `"kept"` {
					t.Errorf("unrelated positive guarantee discarded: %+v", facts[1])
				}
			})
		}
		for _, query := range []string{
			`search EventCode=1 EventCode=2 src=x | table src`,
			`search EventCode=1 src=x | where EventCode=2 | where EventCode=1 | table src`,
		} {
			s := rewriteTestSession(t, language, query, RewriteFactProbe{Kind: "field", Identity: rewriteName("EventCode")})
			fact := rewriteFind(t, s, "field", "src", 1).Facts[0]
			if !fact.LiteralComplete || len(fact.GuaranteedValues) != 0 || len(fact.ReferenceIDs) < 2 {
				t.Errorf("%s conflicting search/flow facts authorize or lose evidence: %+v", language, fact)
			}
			for _, loc := range fact.Locations {
				if !strings.HasPrefix(query[loc.Start.Offset:loc.End.Offset], "EventCode") {
					t.Errorf("invented contradiction evidence: %+v", loc)
				}
			}
		}
	}
}
