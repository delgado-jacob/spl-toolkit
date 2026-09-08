package rewrite

import (
	"encoding/json"
	"os"
	"reflect"
	"sync"
	"testing"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
)

type editFixture struct {
	ID, Category, Language, Query, Candidate, Reason string
	Rules                                            [][]string
	Condition                                        *Condition
}

func candidateForTest(t *testing.T, document analysis.QueryDocument, rules []Rule) (*rewriteCandidate, *analysis.RewriteSession) {
	t.Helper()
	prepared, err := prepareRules(rules)
	if err != nil {
		t.Fatal(err)
	}
	probes := conditionProbes(prepared)
	session, err := analysis.PrepareRewrite(document, probes)
	if err != nil {
		t.Fatal(err)
	}
	got, err := buildCandidate(session, prepared, probes, selectRules(prepared, probes, session.Evidence()))
	if err != nil {
		t.Fatal(err)
	}
	return got, session
}

func runEditFixtures(t *testing.T, category string) {
	t.Helper()
	data, err := os.ReadFile("../../testdata/rewrite/edits.json")
	if err != nil {
		t.Fatal(err)
	}
	var corpus struct{ Cases []editFixture }
	if err := json.Unmarshal(data, &corpus); err != nil {
		t.Fatal(err)
	}
	for _, tc := range corpus.Cases {
		if tc.Category != category {
			continue
		}
		t.Run(tc.ID, func(t *testing.T) {
			rules := []Rule{}
			for i, r := range tc.Rules {
				kind := "field"
				if len(r) > 2 {
					kind = r[2]
				}
				rules = append(rules, Rule{ID: string(rune('A' + i)), Kind: kind, Source: conditionIdentity(r[0]), Target: conditionIdentity(r[1]), When: tc.Condition})
			}
			got, _ := candidateForTest(t, analysis.QueryDocument{Text: tc.Query, Language: tc.Language}, rules)
			want := tc.Candidate
			if want == "" {
				want = tc.Query
			}
			if got.Text != want {
				t.Errorf("candidate = %q, want %q; groups=%+v", got.Text, want, got.Groups)
			}
			if tc.Reason != "" {
				found := false
				for _, c := range got.Changes {
					if c.Reason == tc.Reason && !c.CandidateApplied {
						found = true
					}
				}
				if !found || !got.Incomplete {
					t.Errorf("missing incomplete %s audit: %+v", tc.Reason, got)
				}
			}
			for _, c := range got.Changes {
				if c.Committed {
					t.Error("candidate committed an edit")
				}
				if c.CandidateApplied {
					if c.CandidateLocation == nil || c.OriginalLocation == nil {
						t.Fatalf("missing location: %+v", c)
					}
					if got.Text[c.CandidateLocation.Start.Offset:c.CandidateLocation.End.Offset] != c.NewText {
						t.Errorf("candidate audit not at replacement: %+v", c)
					}
					if tc.Query[c.OriginalLocation.Start.Offset:c.OriginalLocation.End.Offset] != c.OldText {
						t.Errorf("original audit not at old text: %+v", c)
					}
				}
			}
			if tc.ID == "compatible-catalog" {
				if len(got.Changes) != 1 || !reflect.DeepEqual(got.Changes[0].RuleIDs, []string{"A", "B"}) || len(got.Changes[0].OriginalReferenceIDs) != 2 || len(got.Changes[0].CandidateReferenceIDs) != 2 {
					t.Errorf("co-render lost provenance: %+v", got.Changes)
				}
			}
			if tc.ID == "identical-targets" {
				if len(got.Changes) != 2 {
					t.Errorf("duplicate edit audit: %+v", got.Changes)
				}
				for _, c := range got.Changes {
					if !reflect.DeepEqual(c.RuleIDs, []string{"A", "B"}) {
						t.Errorf("lost contributing IDs: %+v", c)
					}
				}
			}
		})
	}
}

func TestRewriteLinkedAliases(t *testing.T) { runEditFixtures(t, "linked") }
func TestRewriteConflicts(t *testing.T)     { runEditFixtures(t, "conflicts") }

func TestRewriteByteCanonicalRender(t *testing.T) { runEditFixtures(t, "byte") }

func TestRewriteConflictsAuditAlternatives(t *testing.T) {
	got, _ := candidateForTest(t, analysis.QueryDocument{Text: "search src=x"}, []Rule{
		{ID: "user", Kind: "field", Source: conditionIdentity("src"), Target: conditionIdentity("user")},
		{ID: "account", Kind: "field", Source: conditionIdentity("src"), Target: conditionIdentity("account")},
	})
	if len(got.Changes) != 2 {
		t.Fatalf("lost alternatives: %+v", got.Changes)
	}
	for _, c := range got.Changes {
		if c.OldText != "src" || c.Outcome != "ambiguous" || c.CandidateApplied || c.CandidateLocation != nil || len(c.CandidateReferenceIDs) != 0 {
			t.Errorf("ambiguous audit: %+v", c)
		}
		if !reflect.DeepEqual(c.RuleIDs, []string{c.NewText}) {
			t.Errorf("alternative rule evidence mixed: %+v", c)
		}
	}
}

func TestRewriteLinkedAuthorizationAlternatives(t *testing.T) {
	condition := conditionLiteral("EventCode", "equals", `1`)
	rules := []Rule{
		{ID: "guarded", Kind: "field", Source: conditionIdentity("src"), Target: conditionIdentity("user"), When: &condition},
		{ID: "always", Kind: "field", Source: conditionIdentity("src"), Target: conditionIdentity("user")},
	}
	got, _ := candidateForTest(t, analysis.QueryDocument{Text: `search EventCode=1 | stats sum(src) | table 'sum(src)'`}, rules)
	if got.Text != `search EventCode=1 | stats sum(user) | table 'sum(user)'` || !got.Incomplete {
		t.Fatalf("independent authorization lost: %+v", got)
	}
	if !reflect.DeepEqual(got.Changes[0].RuleIDs, []string{"guarded", "always"}) || !reflect.DeepEqual(got.Changes[1].RuleIDs, []string{"always"}) {
		t.Fatalf("flow-point rule evidence lost: %+v", got.Changes)
	}
	found := false
	for _, e := range got.Evaluations {
		if e.RuleID == "guarded" && e.Condition != nil && e.Condition.State == "unknown" {
			found = true
		}
	}
	if !found {
		t.Fatal("required consumer condition evidence lost")
	}
}

func TestRewriteConflictsDeterministicConcurrent(t *testing.T) {
	rules := []Rule{
		{ID: "z", Kind: "field", Source: conditionIdentity("a"), Target: conditionIdentity("b")},
		{ID: "x", Kind: "field", Source: conditionIdentity("b"), Target: conditionIdentity("a")},
	}
	probes := conditionProbes(rules)
	session, err := analysis.PrepareRewrite(analysis.QueryDocument{Text: "search a=x b=y | table a b", SourceID: "stable"}, probes)
	if err != nil {
		t.Fatal(err)
	}
	evidence := session.Evidence()
	selection := selectRules(rules, probes, evidence)
	before, _ := json.Marshal(struct {
		Rules     []Rule
		Selection ruleSelection
		Evidence  analysis.RewriteEvidence
	}{rules, selection, evidence})
	var wg sync.WaitGroup
	outputs := make(chan string, 8)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			got, err := buildCandidate(session, rules, probes, selection)
			if err != nil {
				t.Error(err)
				return
			}
			raw, err := json.Marshal(struct {
				Text        string
				Changes     []Change
				Evaluations []RuleEvaluation
				Groups      []editGroup
			}{got.Text, got.Changes, got.Evaluations, got.Groups})
			if err != nil {
				t.Error(err)
				return
			}
			outputs <- string(raw)
		}()
	}
	wg.Wait()
	close(outputs)
	first := ""
	count := 0
	for output := range outputs {
		count++
		if first == "" {
			first = output
		}
		if output != first {
			t.Fatal("concurrent audit is nondeterministic")
		}
	}
	if count != 8 {
		t.Fatalf("only %d calls completed", count)
	}
	after, _ := json.Marshal(struct {
		Rules     []Rule
		Selection ruleSelection
		Evidence  analysis.RewriteEvidence
	}{rules, selection, evidence})
	if string(after) != string(before) {
		t.Fatal("candidate construction mutated inputs")
	}
	reverse := []Rule{rules[1], rules[0]}
	got, err := buildCandidate(session, reverse, probes, selectRules(reverse, probes, evidence))
	if err != nil {
		t.Fatal(err)
	}
	if got.Text != "search b=x a=y | table b a" {
		t.Fatalf("rule order became precedence: %q", got.Text)
	}
}

func TestRewriteLinkedOutputOwnership(t *testing.T) {
	condition := conditionLiteral("EventCode", "equals", `1`)
	rules := []Rule{{ID: "guarded", Kind: "field", Source: conditionIdentity("src"), Target: conditionIdentity("user"), When: &condition}}
	probes := conditionProbes(rules)
	session, err := analysis.PrepareRewrite(analysis.QueryDocument{Text: "search EventCode=1 src=x"}, probes)
	if err != nil {
		t.Fatal(err)
	}
	selection := selectRules(rules, probes, session.Evidence())
	before, _ := json.Marshal(selection)
	got, err := buildCandidate(session, rules, probes, selection)
	if err != nil {
		t.Fatal(err)
	}
	got.Evaluations[0].Condition.State = "mutated"
	got.Evaluations[0].ReferenceIDs[0] = "mutated"
	got.Evaluations[0].Location.Start.Offset = -1
	after, _ := json.Marshal(selection)
	if string(after) != string(before) {
		t.Fatal("returned audit aliases input selection")
	}
}

func TestRewriteLinkedOrdinaryAndUnknownSkips(t *testing.T) {
	for _, tc := range []struct {
		name, source, target, query, reason string
		guard                               *Condition
		incomplete                          bool
	}{
		{name: "unmatched", source: "absent", target: "user", query: "search src=x", reason: ReasonNoMatch},
		{name: "same identity", source: "src", target: "src", query: "search src=x", reason: ReasonNoChange},
		{name: "false guard", source: "src", target: "user", query: "search src=x", reason: ReasonConditionFalse, guard: func() *Condition { c := conditionLiteral("missing", "equals", `1`); return &c }()},
		{name: "derived", source: "alias", target: "user", query: "search src=x | eval alias=src | table alias", reason: ReasonUnsupportedReference},
		{name: "unknown guard", source: "src", target: "user", query: "search EventCode=* src=x", reason: ReasonConditionUnknown, guard: func() *Condition {
			c := conditionLiteral("EventCode", "equals", `1`)
			return &c
		}(), incomplete: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, _ := candidateForTest(t, analysis.QueryDocument{Text: tc.query}, []Rule{{ID: "rule", Kind: "field", Source: conditionIdentity(tc.source), Target: conditionIdentity(tc.target), When: tc.guard}})
			if got.Text != tc.query || len(got.Changes) != 0 || got.Incomplete != tc.incomplete {
				t.Errorf("skip changed candidate/coverage: %+v", got)
			}
			found := false
			for _, e := range got.Evaluations {
				found = found || e.Reason == tc.reason
			}
			if !found {
				t.Errorf("skip reason lost: %+v", got.Evaluations)
			}
		})
	}
}
