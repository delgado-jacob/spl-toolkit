package rewrite

import (
	"bytes"
	"encoding/json"
	"math/big"
	"sort"
	"strings"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
)

// evaluateConditionAtSite consumes only original canonical facts. It deliberately
// does not source-match the site: linked render requirements must independently
// evaluate the originating condition at their own original flow point.
// Conditions and probes have already passed request preparation.
func evaluateConditionAtSite(condition *Condition, probes []analysis.RewriteFactProbe, site analysis.RewriteSite) ConditionEvaluation {
	if condition == nil {
		return conditionResult("true", nil, nil)
	}
	c := *condition
	if c.All != nil || c.Any != nil {
		children := c.All
		if c.Any != nil {
			children = c.Any
		}
		evaluations := make([]ConditionEvaluation, 0, len(children))
		ids := []string{}
		seen := map[string]bool{}
		for i := range children {
			child := evaluateConditionAtSite(&children[i], probes, site)
			evaluations = append(evaluations, child)
			ids = append(ids, child.ReferenceIDs...)
			seen[child.State] = true
		}
		state := "unknown"
		if c.All != nil {
			if seen["false"] {
				state = "false"
			} else if !seen["unknown"] {
				state = "true"
			}
		} else {
			if seen["true"] {
				state = "true"
			} else if !seen["unknown"] {
				state = "false"
			}
		}
		return conditionResult(state, ids, evaluations)
	}
	for probeIndex, probe := range probes {
		if c.Identity == nil || c.Kind != probe.Kind || !conditionIdentityEqual(*c.Identity, probe.Identity) {
			continue
		}
		for _, fact := range site.Facts {
			if fact.ProbeIndex != probeIndex {
				continue
			}
			if c.Fact == "source_reference_present" {
				state := fact.ReferenceState
				if state != "true" && state != "false" {
					state = "unknown"
				}
				return conditionResult(state, fact.ReferenceIDs, nil)
			}
			wantedKind, wanted, understood := conditionScalar(c.Value)
			complete := fact.LiteralComplete && understood
			for _, scalar := range fact.GuaranteedValues {
				kind, value, ok := conditionScalar(scalar.Value)
				if !ok || kind != scalar.Kind {
					complete = false
					continue
				}
				if !understood || kind != wantedKind {
					continue
				}
				match := c.Operator == "equals" && value == wanted
				if c.Operator == "contains" && kind == "string" {
					match = strings.Contains(value, wanted)
				}
				if match {
					return conditionResult("true", fact.ReferenceIDs, nil)
				}
			}
			if complete {
				return conditionResult("false", fact.ReferenceIDs, nil)
			}
			return conditionResult("unknown", fact.ReferenceIDs, nil)
		}
	}
	return conditionResult("unknown", nil, nil)
}

func conditionResult(state string, ids []string, children []ConditionEvaluation) ConditionEvaluation {
	reason := "condition_true"
	if state == "false" {
		reason = ReasonConditionFalse
	}
	if state == "unknown" {
		reason = ReasonConditionUnknown
	}
	return ConditionEvaluation{State: state, Reason: reason, ReferenceIDs: conditionReferenceIDs(ids), Children: append([]ConditionEvaluation{}, children...)}
}

func conditionReferenceIDs(ids []string) []string {
	out := append([]string{}, ids...)
	sort.Strings(out)
	unique := out[:0]
	for _, id := range out {
		if id != "" && (len(unique) == 0 || unique[len(unique)-1] != id) {
			unique = append(unique, id)
		}
	}
	return unique
}

func conditionIdentityEqual(a Identity, b analysis.RewriteIdentity) bool {
	if (a.Name == nil) != (b.Name == nil) || len(a.Path) != len(b.Path) {
		return false
	}
	if a.Name != nil && *a.Name != *b.Name {
		return false
	}
	for i := range a.Path {
		if a.Path[i] != b.Path[i] {
			return false
		}
	}
	return true
}

// conditionScalar normalizes JSON scalar data, never query expressions. Numeric
// equality retains every digit and distinguishes numbers from all other kinds.
func conditionScalar(raw json.RawMessage) (kind, value string, ok bool) {
	if !json.Valid(raw) {
		return "", "", false
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var decoded any
	if decoder.Decode(&decoded) != nil {
		return "", "", false
	}
	switch v := decoded.(type) {
	case string:
		return "string", v, true
	case json.Number:
		return "number", conditionNumber(string(v)), true
	case bool:
		if v {
			return "boolean", "true", true
		}
		return "boolean", "false", true
	case nil:
		return "null", "null", true
	default:
		return "", "", false
	}
}

// Normalize a valid JSON number as coefficient * 10^exponent without expanding
// the exponent. Even very large accepted exponents require only input-sized
// storage, unlike converting the decimal into a rational numerator/denominator.
func conditionNumber(number string) string {
	sign := ""
	if strings.HasPrefix(number, "-") {
		sign, number = "-", number[1:]
	}
	exponent := new(big.Int)
	if i := strings.IndexAny(number, "eE"); i >= 0 {
		exponent.SetString(number[i+1:], 10)
		number = number[:i]
	}
	if i := strings.IndexByte(number, '.'); i >= 0 {
		exponent.Sub(exponent, big.NewInt(int64(len(number)-i-1)))
		number = number[:i] + number[i+1:]
	}
	number = strings.TrimLeft(number, "0")
	if number == "" {
		return "0"
	}
	coefficient := strings.TrimRight(number, "0")
	exponent.Add(exponent, big.NewInt(int64(len(number)-len(coefficient))))
	return sign + coefficient + "e" + exponent.String()
}

// ruleProposal leaves target conflicts to candidate construction while retaining
// every originating rule for independent linked-member condition evaluation.
type ruleProposal struct {
	SiteID               string
	Target               Identity
	RuleIDs              []string
	EvidenceReferenceIDs []string
}
type ruleSelection struct {
	Evaluations []RuleEvaluation
	Proposals   []ruleProposal
	Incomplete  bool
}

// conditionProbes requests every leaf before analysis. Stable, distinct typed
// keys make probe indices independent of rule order and preserve atom/path IDs.
func conditionProbes(rules []Rule) []analysis.RewriteFactProbe {
	byKey := map[string]analysis.RewriteFactProbe{}
	var visit func(*Condition)
	visit = func(c *Condition) {
		if c == nil {
			return
		}
		for i := range c.All {
			visit(&c.All[i])
		}
		for i := range c.Any {
			visit(&c.Any[i])
		}
		if c.Identity != nil {
			id := conditionCopyIdentity(*c.Identity)
			byKey[c.Kind+"\x00"+conditionIdentityKey(id)] = analysis.RewriteFactProbe{Kind: c.Kind, Identity: analysis.RewriteIdentity{Name: id.Name, Path: id.Path}}
		}
	}
	for i := range rules {
		visit(rules[i].When)
	}
	keys := make([]string, 0, len(byKey))
	for key := range byKey {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	probes := make([]analysis.RewriteFactProbe, 0, len(keys))
	for _, key := range keys {
		probes = append(probes, byKey[key])
	}
	return probes
}

// selectRules is a pure selection pass over prepared rules and one original
// canonical evidence snapshot. Evaluations and contributing rule IDs preserve
// request order; candidate sites follow source order with stable canonical ties.
// Equal targets share a proposal; distinct targets all survive for the
// candidate constructor to resolve as conflicts, never as rule precedence.
func selectRules(rules []Rule, probes []analysis.RewriteFactProbe, evidence analysis.RewriteEvidence) ruleSelection {
	sites := append([]analysis.RewriteSite{}, evidence.Sites...)
	sort.SliceStable(sites, func(i, j int) bool {
		if sites[i].Location.Start.Offset != sites[j].Location.Start.Offset {
			return sites[i].Location.Start.Offset < sites[j].Location.Start.Offset
		}
		return sites[i].Location.End.Offset < sites[j].Location.End.Offset
	})
	siteOrder := make(map[string]int, len(sites))
	for i, site := range sites {
		siteOrder[site.ID] = i
	}
	result := ruleSelection{Evaluations: []RuleEvaluation{}, Proposals: []ruleProposal{}}
	proposalByKey := map[string]int{}
	for _, rule := range rules {
		matched := false
		for _, site := range sites {
			if rule.Kind != site.Kind || !conditionIdentityEqual(rule.Source, site.Identity) {
				continue
			}
			matched = true
			location := site.Location
			evaluation := RuleEvaluation{RuleID: rule.ID, Outcome: "skipped", ReferenceIDs: []string{site.ReferenceID}, Location: &location}
			condition := evaluateConditionAtSite(rule.When, probes, site)
			if rule.When != nil {
				evaluation.Condition = &condition
			}
			switch {
			case condition.State == "false":
				evaluation.Reason = ReasonConditionFalse
			case condition.State == "unknown":
				evaluation.Reason = ReasonConditionUnknown
				result.Incomplete = true
			case conditionIdentityEqual(rule.Target, site.Identity):
				evaluation.Reason = ReasonNoChange
			case site.Eligibility != "eligible":
				evaluation.Reason = ReasonUnsupportedReference
				for _, limitation := range site.Limitations {
					if limitation.Code == "dynamic_identity" {
						evaluation.Reason = ReasonDynamicReference
					}
				}
				result.Incomplete = true
			default:
				evaluation.Outcome, evaluation.Reason = "proposed", "matched"
				key := site.ID + "\x00" + conditionIdentityKey(rule.Target)
				index, exists := proposalByKey[key]
				if !exists {
					index = len(result.Proposals)
					proposalByKey[key] = index
					result.Proposals = append(result.Proposals, ruleProposal{SiteID: site.ID, Target: conditionCopyIdentity(rule.Target), RuleIDs: []string{}, EvidenceReferenceIDs: []string{site.ReferenceID}})
				}
				proposal := &result.Proposals[index]
				proposal.RuleIDs = append(proposal.RuleIDs, rule.ID)
				proposal.EvidenceReferenceIDs = conditionReferenceIDs(append(proposal.EvidenceReferenceIDs, condition.ReferenceIDs...))
			}
			result.Evaluations = append(result.Evaluations, evaluation)
		}
		if !matched {
			result.Evaluations = append(result.Evaluations, RuleEvaluation{RuleID: rule.ID, Outcome: "skipped", Reason: ReasonNoMatch, ReferenceIDs: []string{}})
		}
	}
	sort.SliceStable(result.Proposals, func(i, j int) bool {
		return siteOrder[result.Proposals[i].SiteID] < siteOrder[result.Proposals[j].SiteID]
	})
	return result
}

func conditionIdentityKey(id Identity) string {
	raw, _ := json.Marshal(id)
	return string(raw)
}
func conditionCopyIdentity(id Identity) Identity {
	if id.Name != nil {
		name := *id.Name
		return Identity{Name: &name}
	}
	return Identity{Path: append([]string{}, id.Path...)}
}
