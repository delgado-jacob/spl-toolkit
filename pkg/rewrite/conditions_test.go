package rewrite

import (
	"encoding/json"
	"os"
	"reflect"
	"sort"
	"testing"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
)

func conditionIdentity(name string) Identity { return Identity{Name: &name} }
func conditionLiteral(name, operator, value string) Condition {
	id := conditionIdentity(name)
	return Condition{Fact: "literal", Kind: "field", Identity: &id, Operator: operator, Value: json.RawMessage(value)}
}
func conditionProbe(name string) analysis.RewriteFactProbe {
	return analysis.RewriteFactProbe{Kind: "field", Identity: analysis.RewriteIdentity{Name: &name}}
}

// Reversing a controlling branch or short-circuiting away evidence must fail.
func TestConditionTruthTables(t *testing.T) {
	states := []string{"true", "false", "unknown"}
	all := [][]string{{"true", "false", "unknown"}, {"false", "false", "false"}, {"unknown", "false", "unknown"}}
	any := [][]string{{"true", "true", "true"}, {"true", "false", "unknown"}, {"true", "unknown", "unknown"}}
	left, right := conditionIdentity("left"), conditionIdentity("right")
	children := []Condition{{Fact: "source_reference_present", Kind: "field", Identity: &left}, {Fact: "source_reference_present", Kind: "field", Identity: &right}}
	probes := []analysis.RewriteFactProbe{conditionProbe("left"), conditionProbe("right")}
	for i, a := range states {
		for j, b := range states {
			for _, mode := range []string{"all", "any"} {
				t.Run(mode+"/"+a+"/"+b, func(t *testing.T) {
					c, want := Condition{All: children}, all[i][j]
					if mode == "any" {
						c, want = Condition{Any: children}, any[i][j]
					}
					site := analysis.RewriteSite{Facts: []analysis.RewriteFactEvidence{
						{ProbeIndex: 1, ReferenceState: b, ReferenceIDs: []string{"ref-right"}},
						{ProbeIndex: 0, ReferenceState: a, ReferenceIDs: []string{"ref-left"}},
					}}
					got := evaluateConditionAtSite(&c, probes, site)
					if got.State != want {
						t.Errorf("state = %s, want %s", got.State, want)
					}
					if len(got.Children) != 2 || got.Children[0].State != a || got.Children[1].State != b {
						t.Fatalf("lost child decisions: %+v", got)
					}
					if !reflect.DeepEqual(got.ReferenceIDs, []string{"ref-left", "ref-right"}) || !reflect.DeepEqual(got.Children[1].ReferenceIDs, []string{"ref-right"}) {
						t.Fatalf("lost evidence: %+v", got)
					}
					if got.Reason == "" || got.Children[0].Reason == "" || got.Children[1].Reason == "" {
						t.Fatalf("missing reason: %+v", got)
					}
				})
			}
		}
	}
}

// Float decoding, type coercion, raw-text contains, or ignoring incompleteness must fail.
func TestConditionTypedLiterals(t *testing.T) {
	cases := []struct {
		name, kind, established, op, wanted, state string
		incomplete                                 bool
	}{
		{"string", "string", `"Alice"`, "equals", `"Alice"`, "true", false},
		{"decoded string", "string", `"caf\u00e9"`, "equals", `"café"`, "true", false},
		{"number differs from string", "number", `1`, "equals", `"1"`, "false", false},
		{"string differs from number", "string", `"1"`, "equals", `1`, "false", false},
		{"boolean", "boolean", `true`, "equals", `true`, "true", false},
		{"boolean differs from string", "boolean", `true`, "equals", `"true"`, "false", false},
		{"boolean differs from number", "boolean", `true`, "equals", `1`, "false", false},
		{"false boolean", "boolean", `false`, "equals", `true`, "false", false},
		{"null", "null", `null`, "equals", `null`, "true", false},
		{"null differs from string", "null", `null`, "equals", `"null"`, "false", false},
		{"exact integer differs", "number", `9007199254740993`, "equals", `9007199254740992`, "false", false},
		{"exact decimal equality", "number", `9007199254740993.0`, "equals", `9007199254740993`, "true", false},
		{"equivalent exponent", "number", `1e10000`, "equals", `10e9999`, "true", false},
		{"unbounded exponent", "number", `1e9999999999999999999999999999`, "equals", `10e9999999999999999999999999998`, "true", false},
		{"negative exponent", "number", `-0.00100`, "equals", `-1e-3`, "true", false},
		{"negative zero", "number", `-0e9999999999999999999999999999`, "equals", `0`, "true", false},
		{"different sign", "number", `-1`, "equals", `1`, "false", false},
		{"contains", "string", `"Alice Smith"`, "contains", `"ice S"`, "true", false},
		{"contains case sensitive", "string", `"Alice"`, "contains", `"alice"`, "false", false},
		{"contains empty", "string", `"Alice"`, "contains", `""`, "true", false},
		{"contains numeric is not coercion", "number", `123`, "contains", `"2"`, "false", false},
		{"contains exact asterisk", "string", `"x*"`, "contains", `"*"`, "true", false},
		{"unknown representation", "expression", `"Alice"`, "equals", `"Alice"`, "unknown", false},
		{"mismatched representation", "string", `true`, "equals", `true`, "unknown", false},
		{"invalid representation", "string", `"broken`, "equals", `"broken"`, "unknown", false},
		{"missing representation", "null", ``, "equals", `null`, "unknown", false},
		{"incomplete nonmatch", "string", `"a"`, "equals", `"b"`, "unknown", true},
		{"incomplete match", "string", `"a"`, "equals", `"a"`, "true", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := conditionLiteral("value", tc.op, tc.wanted)
			site := analysis.RewriteSite{Facts: []analysis.RewriteFactEvidence{{ProbeIndex: 0, GuaranteedValues: []analysis.RewriteScalar{{Kind: tc.kind, Value: json.RawMessage(tc.established)}}, LiteralComplete: !tc.incomplete, ReferenceState: "true", ReferenceIDs: []string{"ref-value"}}}}
			got := evaluateConditionAtSite(&c, []analysis.RewriteFactProbe{conditionProbe("value")}, site)
			if got.State != tc.state {
				t.Fatalf("got %+v, want %s", got, tc.state)
			}
			if !reflect.DeepEqual(got.ReferenceIDs, []string{"ref-value"}) {
				t.Fatalf("lost scalar evidence: %+v", got)
			}
		})
	}
}

func TestConditionReferenceAndAbsence(t *testing.T) {
	c := conditionLiteral("value", "equals", `1`)
	probes := []analysis.RewriteFactProbe{conditionProbe("value")}
	site := analysis.RewriteSite{Facts: []analysis.RewriteFactEvidence{{ProbeIndex: 0, LiteralComplete: true, ReferenceState: "true", ReferenceIDs: []string{"ref-value"}}}}
	if got := evaluateConditionAtSite(&c, probes, site); got.State != "false" {
		t.Fatalf("reference became equality: %+v", got)
	}
	c.Fact, c.Operator, c.Value = "source_reference_present", "", nil
	for _, state := range []string{"true", "false", "unknown"} {
		site.Facts[0].ReferenceState = state
		if got := evaluateConditionAtSite(&c, probes, site); got.State != state {
			t.Fatalf("reference %s became %+v", state, got)
		}
	}
	site.Facts = nil
	if got := evaluateConditionAtSite(&c, probes, site); got.State != "unknown" {
		t.Fatalf("missing evidence became %+v", got)
	}
	if got := evaluateConditionAtSite(nil, probes, site); got.State != "true" {
		t.Fatalf("unconditional rule became %+v", got)
	}
}

// Reusing later/child facts, parsing text for contains, or treating local derived
// values as source facts must fail these canonical integration cases.
func TestRewriteConditionScopeAndOrder(t *testing.T) {
	raw, err := os.ReadFile("../../testdata/rewrite/conditions.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixtures []struct {
		Name       string                 `json:"name"`
		Document   analysis.QueryDocument `json:"document"`
		Candidate  Identity               `json:"candidate"`
		Occurrence int                    `json:"occurrence"`
		Condition  Condition              `json:"condition"`
		Want       string                 `json:"want"`
	}
	if err := json.Unmarshal(raw, &fixtures); err != nil {
		t.Fatal(err)
	}
	for _, tc := range fixtures {
		t.Run(tc.Name, func(t *testing.T) {
			rules, err := prepareRules([]Rule{{ID: "check", Kind: "field", Source: tc.Candidate, Target: conditionIdentity("replacement"), When: &tc.Condition}})
			if err != nil {
				t.Fatal(err)
			}
			probes := conditionProbes(rules)
			session, err := analysis.PrepareRewrite(tc.Document, probes)
			if err != nil {
				t.Fatal(err)
			}
			evidence := session.Evidence()
			var found *analysis.RewriteSite
			occurrence := 0
			for _, site := range evidence.Sites {
				if site.Kind == "field" && conditionIdentityEqual(tc.Candidate, site.Identity) {
					if occurrence == tc.Occurrence {
						found = &site
						break
					}
					occurrence++
				}
			}
			if found == nil {
				t.Fatalf("candidate occurrence %d absent: %+v", tc.Occurrence, evidence.Sites)
			}
			got := evaluateConditionAtSite(rules[0].When, probes, *found)
			if got.State != tc.Want {
				t.Fatalf("got %+v, want %s; canonical site: %+v", got, tc.Want, *found)
			}
			realIDs := map[string]bool{}
			for _, site := range evidence.Sites {
				realIDs[site.ReferenceID] = true
			}
			for _, id := range got.ReferenceIDs {
				if !realIDs[id] {
					t.Fatalf("manufactured evidence ID %q", id)
				}
			}
		})
	}
}

func TestRewriteConditionRuleOrderAndProposals(t *testing.T) {
	when := conditionLiteral("a", "equals", `"b"`)
	when.Kind = "sourcetype"
	rules, err := prepareRules([]Rule{
		{ID: "rename", Kind: "field", Source: conditionIdentity("src"), Target: conditionIdentity("user")},
		{ID: "duplicate", Kind: "field", Source: conditionIdentity("src"), Target: conditionIdentity("user")},
		{ID: "conflict", Kind: "field", Source: conditionIdentity("src"), Target: conditionIdentity("account")},
		{ID: "type-map", Kind: "sourcetype", Source: conditionIdentity("a"), Target: conditionIdentity("b")},
		{ID: "trigger", Kind: "field", Source: conditionIdentity("src"), Target: conditionIdentity("wrong"), When: &when},
		{ID: "cascade", Kind: "sourcetype", Source: conditionIdentity("b"), Target: conditionIdentity("c")},
		{ID: "self", Kind: "field", Source: conditionIdentity("src"), Target: conditionIdentity("src")},
	})
	if err != nil {
		t.Fatal(err)
	}
	probes := conditionProbes(rules)
	session, err := analysis.PrepareRewrite(analysis.QueryDocument{Text: "search sourcetype=a src=x | table src"}, probes)
	if err != nil {
		t.Fatal(err)
	}
	evidence := session.Evidence()
	before, _ := json.Marshal(evidence)
	got := selectRules(rules, probes, evidence)
	if got.Incomplete || len(got.Proposals) != 5 {
		t.Fatalf("selection: %+v", got)
	}
	counts := map[string]int{}
	for _, p := range got.Proposals {
		for _, id := range p.RuleIDs {
			counts[id]++
		}
		if len(p.RuleIDs) == 2 && !reflect.DeepEqual(p.RuleIDs, []string{"rename", "duplicate"}) {
			t.Fatalf("coalescing lost rule IDs: %+v", p)
		}
	}
	if !reflect.DeepEqual(counts, map[string]int{"rename": 2, "duplicate": 2, "conflict": 2, "type-map": 1}) {
		t.Fatalf("precedence or cascading applied: %v", counts)
	}
	reasons := map[string][]string{}
	for _, e := range got.Evaluations {
		reasons[e.RuleID] = append(reasons[e.RuleID], e.Reason)
		if e.RuleID == "cascade" {
			if e.Location != nil || len(e.ReferenceIDs) != 0 {
				t.Fatalf("no-match fabricated location: %+v", e)
			}
		} else if e.Location == nil || len(e.ReferenceIDs) != 1 {
			t.Fatalf("lost real candidate location: %+v", e)
		}
	}
	if !reflect.DeepEqual(reasons["trigger"], []string{ReasonConditionFalse, ReasonConditionFalse}) || !reflect.DeepEqual(reasons["cascade"], []string{ReasonNoMatch}) || !reflect.DeepEqual(reasons["self"], []string{ReasonNoChange, ReasonNoChange}) {
		t.Fatalf("wrong skip reasons: %v", reasons)
	}
	sort.Slice(rules, func(i, j int) bool { return rules[i].ID > rules[j].ID })
	reorderedProbes := conditionProbes(rules)
	reordered := selectRules(rules, reorderedProbes, evidence)
	if !reflect.DeepEqual(probes, reorderedProbes) || len(reordered.Proposals) != len(got.Proposals) || reordered.Incomplete != got.Incomplete {
		t.Fatal("rule input order changed proposal eligibility")
	}
	for _, proposal := range got.Proposals {
		found := false
		for _, other := range reordered.Proposals {
			if proposal.SiteID != other.SiteID || !reflect.DeepEqual(proposal.Target, other.Target) {
				continue
			}
			firstIDs, secondIDs := append([]string{}, proposal.RuleIDs...), append([]string{}, other.RuleIDs...)
			sort.Strings(firstIDs)
			sort.Strings(secondIDs)
			found = reflect.DeepEqual(firstIDs, secondIDs) && reflect.DeepEqual(proposal.EvidenceReferenceIDs, other.EvidenceReferenceIDs)
		}
		if !found {
			t.Fatalf("rule input order changed proposal semantics: %+v", proposal)
		}
	}
	// Public report/proposal values must not alias input rule or canonical evidence.
	*got.Proposals[0].Target.Name = "mutated"
	*got.Evaluations[1].Location = analysis.Location{}
	after, _ := json.Marshal(evidence)
	if string(before) != string(after) {
		t.Fatal("selection mutated original evidence")
	}
	for _, rule := range rules {
		if *rule.Target.Name == "mutated" {
			t.Fatal("proposal aliases rule target")
		}
	}
}

func TestRewriteConditionEligibilityAndUnknown(t *testing.T) {
	for _, tc := range []struct {
		query, source, reason   string
		conditional, incomplete bool
	}{
		{`search index=main | eval src=1 | table src`, "src", ReasonUnsupportedReference, false, false},
		{`search index=main | table src*`, "src*", ReasonDynamicReference, false, true},
		{`search index=main | where unknown(EventCode)=1 | table src`, "src", ReasonConditionUnknown, true, true},
		{`search index=main | mystery | table src`, "src", ReasonUnsupportedReference, false, true},
		{`search index=main | eval src=1 | where unknown(src)=2`, "src", ReasonUnsupportedReference, false, true},
	} {
		t.Run(tc.reason, func(t *testing.T) {
			rule := Rule{ID: "rename", Kind: "field", Source: conditionIdentity(tc.source), Target: conditionIdentity("user")}
			if tc.conditional {
				c := conditionLiteral("EventCode", "equals", `1`)
				rule.When = &c
			}
			rules, err := prepareRules([]Rule{rule})
			if err != nil {
				t.Fatal(err)
			}
			probes := conditionProbes(rules)
			session, err := analysis.PrepareRewrite(analysis.QueryDocument{Text: tc.query}, probes)
			if err != nil {
				t.Fatal(err)
			}
			got := selectRules(rules, probes, session.Evidence())
			if got.Incomplete != tc.incomplete || len(got.Proposals) != 0 || len(got.Evaluations) == 0 {
				t.Fatalf("selection: %+v; want no proposals, located skips and incomplete=%t", got, tc.incomplete)
			}
			for _, e := range got.Evaluations {
				if e.Reason != tc.reason || e.Outcome != "skipped" || e.Location == nil {
					t.Fatalf("wrong refusal: %+v", e)
				}
			}
		})
	}
}

func TestConditionProbeIdentityAndOwnership(t *testing.T) {
	atom, path := conditionIdentity("actor.name"), Identity{Path: []string{"actor", "name"}}
	c := Condition{All: []Condition{{Fact: "source_reference_present", Kind: "field", Identity: &atom}, {Fact: "source_reference_present", Kind: "field", Identity: &path}, {Fact: "source_reference_present", Kind: "field", Identity: &atom}}}
	rules := []Rule{{When: &c}}
	probes := conditionProbes(rules)
	if len(probes) != 2 {
		t.Fatalf("atom/path collapsed or duplicate retained: %+v", probes)
	}
	site := analysis.RewriteSite{Facts: []analysis.RewriteFactEvidence{{ProbeIndex: 0, ReferenceState: "true", ReferenceIDs: []string{"ref-atom"}}, {ProbeIndex: 1, ReferenceState: "unknown", ReferenceIDs: []string{"ref-path"}}}}
	got := evaluateConditionAtSite(&c, probes, site)
	if got.State != "unknown" || len(got.Children) != 3 || got.Children[0].State != "true" || got.Children[1].State != "unknown" || got.Children[2].State != "true" {
		t.Fatalf("identity interpretation changed: %+v", got)
	}
	*probes[0].Identity.Name = "changed"
	probes[1].Identity.Path[0] = "changed"
	if *atom.Name != "actor.name" || path.Path[0] != "actor" {
		t.Fatal("probes alias condition identities")
	}
}

// Losing either condition's evidence during coalescing, or aliasing a returned
// evidence slice to the original canonical snapshot, must fail.
func TestRewriteConditionProposalEvidence(t *testing.T) {
	sourceType := conditionLiteral("a", "equals", `"a"`)
	sourceType.Kind = "sourcetype"
	eventCode := conditionLiteral("EventCode", "equals", `1`)
	rules, err := prepareRules([]Rule{
		{ID: "event", Kind: "field", Source: conditionIdentity("src"), Target: conditionIdentity("user"), When: &eventCode},
		{ID: "type", Kind: "field", Source: conditionIdentity("src"), Target: conditionIdentity("user"), When: &sourceType},
	})
	if err != nil {
		t.Fatal(err)
	}
	probes := conditionProbes(rules)
	session, err := analysis.PrepareRewrite(analysis.QueryDocument{Text: `search sourcetype=a EventCode=1 src=x | table src`}, probes)
	if err != nil {
		t.Fatal(err)
	}
	evidence := session.Evidence()
	before, _ := json.Marshal(evidence)
	got := selectRules(rules, probes, evidence)
	if got.Incomplete || len(got.Proposals) != 2 || len(got.Evaluations) != 4 {
		t.Fatalf("selection: %+v", got)
	}
	for i, wantIDs := range [][]string{{"ref-0", "ref-1", "ref-2"}, {"ref-0", "ref-1", "ref-3"}} {
		p := got.Proposals[i]
		if !reflect.DeepEqual(p.RuleIDs, []string{"event", "type"}) || !reflect.DeepEqual(p.EvidenceReferenceIDs, wantIDs) {
			t.Fatalf("coalesced evidence: %+v", p)
		}
	}
	for _, evaluation := range got.Evaluations {
		wantIDs := []string{"ref-1"}
		if evaluation.RuleID == "type" {
			wantIDs = []string{"ref-0"}
		}
		if evaluation.Condition == nil || evaluation.Condition.State != "true" || !reflect.DeepEqual(evaluation.Condition.ReferenceIDs, wantIDs) {
			t.Fatalf("lost rule's own proof: %+v", evaluation)
		}
	}
	got.Evaluations[0].Condition.ReferenceIDs[0] = "changed"
	got.Proposals[0].EvidenceReferenceIDs[0] = "changed"
	after, _ := json.Marshal(evidence)
	if string(before) != string(after) {
		t.Fatal("returned evidence aliases original facts")
	}
}

// The originating condition cannot be reused unchanged at an implicit output
// consumer: its own original flow point has lost the source restriction.
func TestRewriteConditionAtImplicitMember(t *testing.T) {
	c := conditionLiteral("EventCode", "equals", `1`)
	rule := Rule{ID: "sum-source", Kind: "field", Source: conditionIdentity("src"), Target: conditionIdentity("user"), When: &c}
	probes := conditionProbes([]Rule{rule})
	session, err := analysis.PrepareRewrite(analysis.QueryDocument{Text: `search EventCode=1 | stats sum(src) | table 'sum(src)'`}, probes)
	if err != nil {
		t.Fatal(err)
	}
	evidence := session.Evidence()
	selected := selectRules([]Rule{rule}, probes, evidence)
	if selected.Incomplete || len(selected.Proposals) != 1 {
		t.Fatalf("source condition did not select: %+v", selected)
	}
	checked := false
	for _, site := range evidence.Sites {
		if site.Identity.Name != nil && *site.Identity.Name == "sum(src)" && site.Role == "selector_atom" {
			checked = true
			got := evaluateConditionAtSite(rule.When, probes, site)
			if site.Eligibility == "eligible" || got.State != "unknown" {
				t.Fatalf("linked member reused source authorization: %+v; site %+v", got, site)
			}
		}
	}
	if !checked {
		t.Fatal("canonical implicit consumer absent")
	}
}

// Request order is audit order, never precedence. SQL site evaluation order is
// logical, so sorting reports by actual canonical source locations matters too.
func TestRewriteConditionRequestOrder(t *testing.T) {
	rules, err := prepareRules([]Rule{
		{ID: "z-first", Kind: "field", Source: conditionIdentity("src"), Target: conditionIdentity("user")},
		{ID: "a-second", Kind: "field", Source: conditionIdentity("src"), Target: conditionIdentity("user")},
		{ID: "m-third", Kind: "field", Source: conditionIdentity("src"), Target: conditionIdentity("account")},
	})
	if err != nil {
		t.Fatal(err)
	}
	probes := conditionProbes(rules)
	session, err := analysis.PrepareRewrite(analysis.QueryDocument{Text: `SELECT src FROM main WHERE src=1 ORDER BY src`, Language: "spl2"}, probes)
	if err != nil {
		t.Fatal(err)
	}
	evidence := session.Evidence()
	got := selectRules(rules, probes, evidence)
	if got.Incomplete || len(got.Proposals) != 6 || len(got.Evaluations) != 9 {
		t.Fatalf("selection: %+v", got)
	}
	for i, wantID := range []string{"z-first", "a-second", "m-third"} {
		for j, wantOffset := range []int{7, 27, 42} {
			evaluation := got.Evaluations[i*3+j]
			if evaluation.RuleID != wantID || evaluation.Location == nil || evaluation.Location.Start.Offset != wantOffset {
				t.Errorf("evaluation %d: %+v; want rule %s at %d", i*3+j, evaluation, wantID, wantOffset)
			}
		}
	}
	sites := map[string]analysis.RewriteSite{}
	for _, site := range evidence.Sites {
		sites[site.ID] = site
	}
	for i, wantOffset := range []int{7, 7, 27, 27, 42, 42} {
		proposal := got.Proposals[i]
		wantRuleIDs := []string{"z-first", "a-second"}
		if i%2 != 0 {
			wantRuleIDs = []string{"m-third"}
		}
		if sites[proposal.SiteID].Location.Start.Offset != wantOffset || !reflect.DeepEqual(proposal.RuleIDs, wantRuleIDs) {
			t.Errorf("proposal %d: %+v; want source offset %d and rules %v", i, proposal, wantOffset, wantRuleIDs)
		}
	}
}

// A known definition/derived exclusion needs the real reference's binding and
// exact resolution; the generic binding_not_source limitation alone is not proof.
func TestRewriteConditionDerivedExclusionProof(t *testing.T) {
	rules, err := prepareRules([]Rule{{ID: "rename", Kind: "field", Source: conditionIdentity("src"), Target: conditionIdentity("user")}})
	if err != nil {
		t.Fatal(err)
	}
	probes := conditionProbes(rules)
	session, err := analysis.PrepareRewrite(analysis.QueryDocument{Text: `search index=main | eval src=1 | table src`}, probes)
	if err != nil {
		t.Fatal(err)
	}
	evidence := session.Evidence()
	bindings := []string{}
	for _, ref := range evidence.Analysis.References {
		if ref.NormalizedName == "src" {
			bindings = append(bindings, ref.Role+":"+ref.Binding)
		}
	}
	if !reflect.DeepEqual(bindings, []string{"create:not_applicable", "read:derived"}) {
		t.Fatalf("fixture lacks known binding proof: %v", bindings)
	}
	got := selectRules(rules, probes, evidence)
	if got.Incomplete || len(got.Proposals) != 0 || len(got.Evaluations) != 2 {
		t.Fatalf("known derived exclusion is not an ordinary skip: %+v", got)
	}
	evidence.Analysis.References = nil
	missing := selectRules(rules, probes, evidence)
	if !missing.Incomplete || len(missing.Proposals) != 0 || len(missing.Evaluations) != 2 {
		t.Fatalf("missing canonical binding proof treated as complete: %+v", missing)
	}
}

func TestRewriteConditionDerivedExclusionUnknownGuard(t *testing.T) {
	first, second := conditionLiteral("EventCode", "equals", `1`), conditionLiteral("EventCode", "equals", `2`)
	condition := Condition{Any: []Condition{first, second}}
	rules, err := prepareRules([]Rule{{ID: "rename", Kind: "field", Source: conditionIdentity("src"), Target: conditionIdentity("user"), When: &condition}})
	if err != nil {
		t.Fatal(err)
	}
	probes := conditionProbes(rules)
	for _, tc := range []struct {
		name, query string
		incomplete  bool
		reason      string
	}{
		{"known exclusion", `search index=main | where EventCode=other | eval src=1 | table src`, false, ReasonUnsupportedReference},
		{"eligible source", `search index=main | where EventCode=other | table src`, true, ReasonConditionUnknown},
	} {
		t.Run(tc.name, func(t *testing.T) {
			session, err := analysis.PrepareRewrite(analysis.QueryDocument{Text: tc.query}, probes)
			if err != nil {
				t.Fatal(err)
			}
			got := selectRules(rules, probes, session.Evidence())
			if got.Incomplete != tc.incomplete || len(got.Proposals) != 0 || len(got.Evaluations) == 0 {
				t.Fatalf("unknown guard changed source eligibility summary: %+v", got)
			}
			for _, evaluation := range got.Evaluations {
				c := evaluation.Condition
				if evaluation.Reason != tc.reason || c == nil || c.State != "unknown" || len(c.Children) != 2 || c.Children[0].State != "unknown" || c.Children[1].State != "unknown" {
					t.Fatalf("lost exclusion or child uncertainty: %+v", evaluation)
				}
			}
		})
	}
}
