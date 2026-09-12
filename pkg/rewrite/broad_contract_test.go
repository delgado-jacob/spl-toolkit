package rewrite

import (
	"fmt"
	"strings"
	"testing"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
)

func TestRewriteSQLHavingCondition(t *testing.T) {
	for _, prefix := range []string{"SELECT marker FROM main", "FROM main SELECT marker"} {
		for _, tc := range []struct{ clauses, field, value, state string }{
			{" GROUP BY marker HAVING marker=1", "marker", "1", "true"},
			{" WHERE marker=1 GROUP BY marker", "marker", "1", "true"},
			{" WHERE marker=1 GROUP BY marker HAVING marker=1", "marker", "1", "true"},
			{" WHERE marker=1 GROUP BY marker HAVING marker=2", "marker", "1", "false"},
			{" WHERE marker=1 GROUP BY marker HAVING marker=2", "marker", "2", "false"},
			{" WHERE marker=1 GROUP BY marker HAVING marker=2 AND hidden=1", "marker", "1", "unknown"},
			{" HAVING marker=1", "marker", "1", "unknown"},
			{" GROUP BY marker HAVING hidden=1", "hidden", "1", "unknown"},
			{" GROUP BY marker HAVING unknown(marker)=1", "marker", "1", "unknown"},
		} {
			t.Run(prefix+tc.clauses+"/"+tc.value, func(t *testing.T) {
				query := prefix + tc.clauses + " | lookup people marker OUTPUT label"
				if strings.HasPrefix(prefix, "FROM") {
					before, after, hasHaving := strings.Cut(tc.clauses, " HAVING ")
					query = "FROM main" + before + " SELECT marker"
					if hasHaving {
						query += " HAVING " + after
					}
					query += " | lookup people marker OUTPUT label"
				}
				request := rewriteRequest(query)
				request.Document.Language = "spl2"
				when := conditionLiteral(tc.field, "equals", tc.value)
				request.Rules = []Rule{{ID: "catalog", Kind: "lookup", Source: conditionIdentity("people"), Target: conditionIdentity("users"), When: &when}}
				got := requireRewrite(t, request)
				candidate := query
				if tc.state == "true" {
					candidate = strings.Replace(query, "lookup people", "lookup users", 1)
				}
				if got.CandidateText != candidate || got.Text != candidate || got.Committed != (tc.state == "true") || len(got.RuleEvaluations) != 1 || got.RuleEvaluations[0].Condition.State != tc.state {
					t.Fatalf("HAVING crossed public authorization boundary: %+v", got)
				}
				if tc.state != "unknown" && got.Status != analysis.Valid {
					t.Fatalf("supported HAVING status: %s", got.Status)
				}
			})
		}
	}
	for _, query := range []string{
		"SELECT count() AS n FROM main HAVING n=1 | lookup people n OUTPUT label",
		"FROM main SELECT count() AS n HAVING n=1 | lookup people n OUTPUT label",
	} {
		request := rewriteRequest(query)
		request.Document.Language = "spl2"
		when := conditionLiteral("n", "equals", "1")
		request.Rules = []Rule{{ID: "catalog", Kind: "lookup", Source: conditionIdentity("people"), Target: conditionIdentity("users"), When: &when}}
		got := requireRewrite(t, request)
		if got.Committed || got.Text != query || got.RuleEvaluations[0].Condition.State != "unknown" {
			t.Fatal("derived HAVING manufactured source guarantee")
		}
	}
}

// Decode a real request containing exact numeric JSON tokens: neither the
// request nor its condition may pass through float64 on the public path.
func TestRewriteLargeExponentCondition(t *testing.T) {
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
					raw := fmt.Sprintf(`{"schema_version":1,"document":{"text":%q,"language":%q},"mode":"apply","rules":[{"id":"r","kind":"field","source":{"name":"src"},"target":{"name":"dst"},"when":{"fact":"literal","kind":"field","identity":{"name":"n"},"operator":"equals","value":%s}}]}`, query, language, pair.left)
					request, err := DecodeRequest([]byte(raw))
					if err != nil {
						t.Fatal(err)
					}
					got := requireRewrite(t, request)
					candidate, state := query, "false"
					if pair.equal {
						candidate, state = strings.Replace(query, "table src", "table dst", 1), "true"
					}
					if got.Status != analysis.Valid || got.Committed != pair.equal || got.Text != candidate || got.CandidateText != candidate || len(got.RuleEvaluations) != 1 || got.RuleEvaluations[0].Condition.State != state {
						t.Fatalf("exact public numeric comparison failed: %+v", got)
					}
				})
			}
		}
	}
}

func TestRewriteSPLExactExponentConditionAndSyntaxBoundary(t *testing.T) {
	for _, value := range []string{"0e1000001", "0e-1000001", "0e9223372036854775809", "-0e-9223372036854775809"} {
		request := rewriteRequest("search index=main | where n=0 AND n=0.0 | table src")
		request.Document.Language = "spl"
		when := conditionLiteral("n", "equals", value)
		request.Rules[0].When = &when
		got := requireRewrite(t, request)
		if !got.Committed || got.Text != "search index=main | where n=0 AND n=0.0 | table user" {
			t.Fatal("exact JSON zero condition lost")
		}
	}
	for _, literal := range []string{"1e1000001", "1e-9223372036854775809"} {
		request := rewriteRequest("search index=main | where n=" + literal + " | table src")
		request.Document.Language = "spl"
		got := requireRewrite(t, request)
		if got.Status != analysis.Invalid || got.Committed || got.Text != request.Document.Text {
			t.Fatal("unsupported SPL numeric grammar promoted")
		}
	}
}
