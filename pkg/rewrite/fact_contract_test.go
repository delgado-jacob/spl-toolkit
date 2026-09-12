package rewrite

import (
	"testing"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
)

func TestRewriteDependencySelectorCondition(t *testing.T) {
	for _, language := range []string{"spl", "spl2"} {
		for _, selector := range []struct{ kind, value string }{{"index", "main"}, {"source", "app"}, {"sourcetype", "auth"}} {
			for _, identity := range []string{selector.kind, selector.value} {
				t.Run(language+"/"+selector.kind+"/"+identity, func(t *testing.T) {
					r := rewriteRequest("search " + selector.kind + "=" + selector.value + " src=alice | table src")
					r.Document.Language = language
					when := conditionLiteral(identity, "equals", `"`+selector.value+`"`)
					when.Kind = selector.kind
					r.Rules[0].When = &when
					got := requireRewrite(t, r)
					candidate, state := r.Document.Text, "false"
					if identity == selector.kind {
						candidate = "search " + selector.kind + "=" + selector.value + " user=alice | table user"
						state = "true"
					}
					if got.Status != analysis.Valid || got.CandidateText != candidate || got.Text != candidate || got.Committed != (state == "true") || len(got.RuleEvaluations) != 2 {
						t.Fatalf("wrong selector authorization: %+v", got)
					}
					for _, evaluation := range got.RuleEvaluations {
						if evaluation.Condition.State != state {
							t.Errorf("wrong condition: %+v", evaluation)
						}
					}
				})
			}
		}
	}
}

func TestRewriteConflictingPredicateCondition(t *testing.T) {
	for _, language := range []string{"spl", "spl2"} {
		for _, tc := range []struct{ predicate, state, want string }{
			{`EventCode=1 AND EventCode=2`, "false", "src"},
			{`EventCode=1 AND EventCode=1`, "true", "user"},
			{`EventCode=1 AND EventCode=2 AND unknown(other)=3`, "unknown", "src"},
			{`((EventCode=1 AND EventCode=2) OR tag="x") AND EventCode=1`, "false", "src"},
		} {
			t.Run(language+"/"+tc.predicate, func(t *testing.T) {
				r := rewriteRequest(`search index=main | where ` + tc.predicate + ` | table src`)
				r.Document.Language = language
				when := conditionLiteral("EventCode", "equals", `1`)
				r.Rules[0].When = &when
				got := requireRewrite(t, r)
				candidate := `search index=main | where ` + tc.predicate + ` | table ` + tc.want
				status := analysis.Valid
				if tc.state == "unknown" {
					status = analysis.Incomplete
				}
				if got.Status != status || got.CandidateText != candidate || got.Text != candidate || got.Committed != (tc.state == "true") || len(got.RuleEvaluations) != 1 || got.RuleEvaluations[0].Condition.State != tc.state {
					t.Fatalf("conflicting predicate crossed publication/evidence boundary: %+v", got)
				}
			})
		}
	}
}
