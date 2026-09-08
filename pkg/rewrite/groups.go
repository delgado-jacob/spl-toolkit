package rewrite

import (
	"fmt"
	"sort"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
)

// rewriteCandidate is preview evidence only. Task 5 owns the whole-request
// verification/validation gate and publication of authoritative apply text.
type rewriteCandidate struct {
	Text        string
	Changes     []Change
	Groups      []editGroup
	Evaluations []RuleEvaluation
	Rendering   *analysis.RewriteRendering
	Session     *analysis.RewriteSession
	Proof       analysis.RewriteProof
	Incomplete  bool
}

type editGroup struct {
	ID, Reason   string
	Replacements []analysis.RewriteReplacement
	RuleIDs      []string
	Limitations  []analysis.RewriteLimitation
	rendering    *analysis.RewriteRendering
	members      []groupOption
}

type groupOption struct {
	replacements []analysis.RewriteReplacement
	rules        map[string][]string
	reason       string
	limitations  []analysis.RewriteLimitation
	rendering    *analysis.RewriteRendering
}

type groupBuilder struct {
	session     *analysis.RewriteSession
	evidence    analysis.RewriteEvidence
	rules       []Rule
	probes      []analysis.RewriteFactProbe
	sites       map[string]analysis.RewriteSite
	refs        map[string]analysis.Reference
	epochs      map[string]string
	selected    map[string]ruleProposal
	evaluations []RuleEvaluation
	incomplete  bool
}

func replacementKey(siteID string, target Identity) string {
	return siteID + "\x00" + conditionIdentityKey(target)
}
func publicIdentity(id analysis.RewriteIdentity) Identity {
	return Identity{Name: id.Name, Path: id.Path}
}
func canonicalIdentity(id Identity) analysis.RewriteIdentity {
	id = conditionCopyIdentity(id)
	return analysis.RewriteIdentity{Name: id.Name, Path: id.Path}
}

func buildCandidate(session *analysis.RewriteSession, rules []Rule, probes []analysis.RewriteFactProbe, selection ruleSelection) (*rewriteCandidate, error) {
	b := groupBuilder{session: session, evidence: session.Evidence(), rules: rules, probes: probes, sites: map[string]analysis.RewriteSite{}, refs: map[string]analysis.Reference{}, epochs: map[string]string{}, selected: map[string]ruleProposal{}, evaluations: copyGroupEvaluations(selection.Evaluations), incomplete: selection.Incomplete}
	for _, s := range b.evidence.Sites {
		b.sites[s.ID] = s
		b.epochs[s.ReferenceID] = s.SourceEpochID
	}
	for _, r := range b.evidence.Analysis.References {
		b.refs[r.ID] = r
	}
	for _, p := range selection.Proposals {
		b.selected[replacementKey(p.SiteID, p.Target)] = p
	}
	options := []groupOption{}
	seen := map[string]bool{}
	for _, p := range selection.Proposals {
		site, ok := b.sites[p.SiteID]
		if !ok {
			return nil, fmt.Errorf("proposal has no canonical site %q", p.SiteID)
		}
		binding := site.BindingID
		if binding == "" {
			binding = site.ID
		}
		key := binding + "\x00" + conditionIdentityKey(p.Target)
		if seen[key] {
			continue
		}
		seen[key] = true
		option, err := b.expand(p)
		if err != nil {
			return nil, err
		}
		options = append(options, option)
	}
	groups, err := b.connect(options)
	if err != nil {
		return nil, err
	}
	// Interaction components are candidates for one simultaneous canonical proof,
	// not a preferred rewrite subset. In particular, swaps are never tried as a
	// sequence of one-sided rewrites.
	links := newGroupLinks(len(groups))
	for i := range groups {
		for j := 0; j < i; j++ {
			if b.interact(groups[i], groups[j]) {
				links.join(i, j)
			}
		}
	}
	for _, component := range links.components() {
		selected := []analysis.RewriteReplacement{}
		for _, i := range component {
			if groups[i].Reason == "" {
				selected = append(selected, groups[i].Replacements...)
			}
		}
		if len(selected) == 0 {
			continue
		}
		_, _, proof, err := b.prove(selected)
		if err != nil {
			return nil, err
		}
		if !proof.Proven {
			for _, i := range component {
				if groups[i].Reason == "" {
					groups[i].Reason = ReasonBindingCollision
					groups[i].Limitations = append(groups[i].Limitations, proof.Limitations...)
				}
			}
		}
	}
	result := &rewriteCandidate{Changes: []Change{}, Groups: groups, Evaluations: b.evaluations, Incomplete: b.incomplete}
	selected := []analysis.RewriteReplacement{}
	for _, g := range groups {
		if g.Reason == "" {
			selected = append(selected, g.Replacements...)
		} else {
			result.Incomplete = true
		}
	}
	rendering, next, proof, err := b.prove(selected)
	if err != nil {
		return nil, err
	}
	// A missed interaction cannot leave an unproved edit in the candidate. This
	// monotone closure removes the remaining set, without retries
	// that search for an optimally large subset. Task 5 still verifies the final
	// candidate and applies its independent destination-validation gate.
	if len(selected) > 0 && !proof.Proven {
		for i := range result.Groups {
			g := &result.Groups[i]
			if g.Reason == "" {
				g.Reason = ReasonLinkedEditUnproven
				g.Limitations = append(g.Limitations, proof.Limitations...)
			}
		}
		result.Incomplete = true
		rendering, next, proof, err = b.prove(nil)
		if err != nil {
			return nil, err
		}
	}
	result.Rendering, result.Session, result.Proof = rendering, next, proof
	result.Text = next.Evidence().Analysis.Document.Text
	if err := b.audit(result); err != nil {
		return nil, err
	}
	b.sortEvaluations(result.Evaluations)
	return result, nil
}

func (b *groupBuilder) orderedRules(ids ...[]string) []string {
	wanted := map[string]bool{}
	for _, list := range ids {
		for _, id := range list {
			wanted[id] = true
		}
	}
	out := []string{}
	for _, r := range b.rules {
		if wanted[r.ID] {
			out = append(out, r.ID)
		}
	}
	return out
}

func (b *groupBuilder) expand(seed ruleProposal) (groupOption, error) {
	o := groupOption{replacements: []analysis.RewriteReplacement{}, rules: map[string][]string{}, limitations: []analysis.RewriteLimitation{}}
	source := b.sites[seed.SiteID]
	causeRules := []string{}
	for _, p := range b.selected {
		s := b.sites[p.SiteID]
		if s.BindingID == source.BindingID && conditionIdentityEqual(p.Target, canonicalIdentity(seed.Target)) {
			causeRules = b.orderedRules(causeRules, p.RuleIDs)
		}
	}
	// Binding IDs come from canonical transfer. A true rule at one flow point
	// does not authorize the other observed source members of that binding.
	for _, s := range b.evidence.Sites {
		if s.ID != source.ID && (source.BindingID == "" || s.BindingID != source.BindingID || s.Eligibility != "eligible") {
			continue
		}
		p, ok := b.selected[replacementKey(s.ID, seed.Target)]
		ids := p.RuleIDs
		if !ok {
			o.reason = ReasonLinkedEditUnproven
			ids = causeRules
		}
		o.replacements = append(o.replacements, analysis.RewriteReplacement{SiteID: s.ID, Target: canonicalIdentity(seed.Target)})
		o.rules[s.ID] = append([]string{}, ids...)
	}
	for {
		b.sortReplacements(o.replacements)
		rendered, err := b.session.Render(o.replacements)
		if err != nil {
			return o, err
		}
		added := false
		for _, req := range rendered.Requirements() {
			for _, lim := range req.Limitations {
				reason := ReasonLinkedEditUnproven
				if lim.Code == "target_not_renderable" {
					reason = ReasonTargetNotRenderable
				}
				if lim.Code == "render_conflict" {
					reason = ReasonOverlappingEdits
				}
				if o.reason == "" {
					o.reason = reason
				}
				o.limitations = appendLimitation(o.limitations, lim)
			}
			originIDs := []string{}
			for _, id := range req.CauseSiteIDs {
				originIDs = b.orderedRules(originIDs, o.rules[id])
			}
			for _, required := range req.RequiredChanges {
				existing := -1
				for i, c := range o.replacements {
					if c.SiteID == required.SiteID {
						existing = i
						break
					}
				}
				if existing >= 0 {
					if !conditionIdentityEqual(publicIdentity(required.Target), o.replacements[existing].Target) {
						o.reason = ReasonConflictingTargets
					}
					continue
				}
				site, ok := b.sites[required.SiteID]
				if !ok {
					return o, fmt.Errorf("canonical requirement has no site %q", required.SiteID)
				}
				ids := []string{}
				// Only canonical derived implicit consumers inherit source-rule authority;
				// coupled catalog dependencies require their own explicitly supplied rule.
				if site.Kind == "field" && b.refs[site.ReferenceID].Binding == "derived" {
					for _, r := range b.rules {
						if !hasID(originIDs, r.ID) {
							continue
						}
						condition := evaluateConditionAtSite(r.When, b.probes, site)
						loc := site.Location
						ev := RuleEvaluation{RuleID: r.ID, Outcome: "skipped", Reason: ReasonConditionFalse, ReferenceIDs: []string{site.ReferenceID}, Location: &loc}
						if r.When != nil {
							ev.Condition = &condition
						}
						if condition.State == "true" {
							ids = append(ids, r.ID)
							ev.Outcome = "proposed"
							ev.Reason = "linked"
						} else if condition.State == "unknown" {
							ev.Reason = ReasonConditionUnknown
							b.incomplete = true
						}
						b.evaluations = append(b.evaluations, ev)
					}
				} else if p, ok := b.selected[replacementKey(site.ID, publicIdentity(required.Target))]; ok {
					ids = append(ids, p.RuleIDs...)
				}
				if len(ids) == 0 {
					o.reason = ReasonLinkedEditUnproven
					ids = originIDs
				}
				o.replacements = append(o.replacements, required)
				o.rules[required.SiteID] = ids
				added = true
			}
		}
		if !added {
			o.rendering = rendered
			break
		}
	}
	if limitation, ok := b.uncertain(o.rendering); ok {
		if o.reason == "" {
			o.reason = ReasonLinkedEditUnproven
		}
		o.limitations = appendLimitation(o.limitations, limitation)
	}
	return o, nil
}

func (b *groupBuilder) sortReplacements(changes []analysis.RewriteReplacement) {
	sort.SliceStable(changes, func(i, j int) bool {
		a, c := b.sites[changes[i].SiteID], b.sites[changes[j].SiteID]
		if a.Location.Start.Offset != c.Location.Start.Offset {
			return a.Location.Start.Offset < c.Location.Start.Offset
		}
		if a.Location.End.Offset != c.Location.End.Offset {
			return a.Location.End.Offset < c.Location.End.Offset
		}
		return a.ID < c.ID
	})
}

// Relevance comes from the kernel's live origin sets at a transition. An epoch
// or similarly spelled operand alone never makes an unrelated scope relevant.
func (b *groupBuilder) uncertain(rendering *analysis.RewriteRendering) (analysis.RewriteLimitation, bool) {
	affected := map[string]bool{}
	for _, effect := range rendering.Effects() {
		if b.refs[effect.ReferenceID].Kind == "field" {
			affected[effect.ReferenceID] = true
		}
	}
	for _, lineage := range b.evidence.Analysis.Lineage {
		live, owned := false, false
		for _, site := range b.evidence.Sites {
			if affected[site.ReferenceID] && site.Point.StageID == lineage.StageID && site.Point.Phase == lineage.Phase {
				owned = true
			}
		}
		for _, state := range []analysis.FieldState{lineage.Before, lineage.After} {
			for _, field := range state.Fields {
				for _, id := range field.OriginReferenceIDs {
					if affected[id] {
						live = true
					}
				}
			}
		}
		if !live && !owned {
			continue
		}
		for _, stage := range b.evidence.Analysis.Stages {
			if stage.ID == lineage.StageID {
				if lineage.Before.Uncertain || lineage.After.Uncertain || (owned && !stage.SemanticComplete) {
					return analysis.RewriteLimitation{Code: "linked_transition_unproved", Message: "An uncertain canonical transition carries an affected field origin", Location: stage.Location}, true
				}
			}
		}
		for _, site := range b.evidence.Sites {
			if site.Kind == "field" && site.Point.StageID == lineage.StageID {
				for _, lim := range site.Limitations {
					if lim.Code == "dynamic_identity" || lim.Code == "unproved_owner" {
						return lim, true
					}
				}
			}
		}
	}
	return analysis.RewriteLimitation{}, false
}

func (b *groupBuilder) connect(options []groupOption) ([]editGroup, error) {
	links := newGroupLinks(len(options))
	for i, a := range options {
		for j := 0; j < i; j++ {
			joined := false
			for _, x := range a.replacements {
				for _, y := range options[j].replacements {
					if x.SiteID == y.SiteID {
						joined = true
					}
				}
			}
			for _, x := range a.rendering.Edits() {
				for _, y := range options[j].rendering.Edits() {
					if editIntervalsTouch(x, y) {
						joined = true
					}
				}
			}
			if joined {
				links.join(i, j)
			}
		}
	}
	groups := []editGroup{}
	for _, indices := range links.components() {
		g := editGroup{ID: fmt.Sprintf("group-%d", len(groups)), Replacements: []analysis.RewriteReplacement{}, RuleIDs: []string{}, Limitations: []analysis.RewriteLimitation{}, members: []groupOption{}}
		bySite := map[string]analysis.RewriteIdentity{}
		for _, i := range indices {
			o := options[i]
			g.members = append(g.members, o)
			if g.Reason == "" {
				g.Reason = o.reason
			}
			for _, lim := range o.limitations {
				g.Limitations = appendLimitation(g.Limitations, lim)
			}
			for _, c := range o.replacements {
				g.RuleIDs = b.orderedRules(g.RuleIDs, o.rules[c.SiteID])
				if target, ok := bySite[c.SiteID]; ok {
					if !conditionIdentityEqual(publicIdentity(c.Target), target) {
						g.Reason = ReasonConflictingTargets
					}
				} else {
					bySite[c.SiteID] = c.Target
					g.Replacements = append(g.Replacements, c)
				}
			}
		}
		b.sortReplacements(g.Replacements)
		if g.Reason != ReasonConflictingTargets {
			r, err := b.session.Render(g.Replacements)
			if err != nil {
				return nil, err
			}
			g.rendering = r
			for _, req := range r.Requirements() {
				for _, lim := range req.Limitations {
					if lim.Code == "render_conflict" {
						g.Reason = ReasonOverlappingEdits
					}
					g.Limitations = appendLimitation(g.Limitations, lim)
				}
			}
		}
		groups = append(groups, g)
	}
	return groups, nil
}

// Interaction edges only schedule a shared canonical proof. Name equality is
// never used as proof that a swap, chain, capture or collapse is safe.
func (b *groupBuilder) interact(a, c editGroup) bool {
	for _, x := range a.members {
		for _, y := range c.members {
			for _, e := range x.rendering.Effects() {
				for _, f := range y.rendering.Effects() {
					er, fr := b.refs[e.ReferenceID], b.refs[f.ReferenceID]
					if er.Kind != "field" || fr.Kind != "field" {
						continue
					}
					related := b.epochs[er.ID] != "" && b.epochs[er.ID] == b.epochs[fr.ID]
					if !related {
						for _, l := range b.evidence.Analysis.Lineage {
							for _, state := range []analysis.FieldState{l.Before, l.After} {
								first, second := false, false
								for _, field := range state.Fields {
									first = first || hasID(field.OriginReferenceIDs, er.ID)
									second = second || hasID(field.OriginReferenceIDs, fr.ID)
								}
								related = related || (first && second)
							}
						}
					}
					if related && (conditionIdentityEqual(publicIdentity(e.After), f.Before) || conditionIdentityEqual(publicIdentity(f.After), e.Before) || conditionIdentityEqual(publicIdentity(e.After), f.After)) {
						return true
					}
				}
			}
		}
	}
	return false
}

func (b *groupBuilder) prove(changes []analysis.RewriteReplacement) (*analysis.RewriteRendering, *analysis.RewriteSession, analysis.RewriteProof, error) {
	b.sortReplacements(changes)
	rendering, err := b.session.Render(changes)
	if err != nil {
		return nil, nil, analysis.RewriteProof{}, err
	}
	text, _, err := reconstructEdits(b.evidence.Analysis.Document.Text, rendering.Edits())
	if err != nil {
		return nil, nil, analysis.RewriteProof{}, err
	}
	document := b.evidence.Analysis.Document
	document.Text = text
	candidate, err := analysis.PrepareRewrite(document, b.probes)
	if err != nil {
		return nil, nil, analysis.RewriteProof{}, err
	}
	return rendering, candidate, b.session.Verify(candidate, rendering), nil
}

func editIntervalsTouch(a, b analysis.RewriteTextEdit) bool {
	x, y := a.Location, b.Location
	if x.Start.Offset == y.Start.Offset && x.End.Offset == y.End.Offset {
		return true
	}
	return x.Start.Offset < y.End.Offset && y.Start.Offset < x.End.Offset
}
func hasID(ids []string, id string) bool {
	for _, candidate := range ids {
		if id == candidate {
			return true
		}
	}
	return false
}
func appendLimitation(limits []analysis.RewriteLimitation, lim analysis.RewriteLimitation) []analysis.RewriteLimitation {
	for _, old := range limits {
		if old == lim {
			return limits
		}
	}
	return append(limits, lim)
}

type groupLinks []int

func newGroupLinks(n int) groupLinks {
	links := make(groupLinks, n)
	for i := range links {
		links[i] = i
	}
	return links
}
func (links groupLinks) root(i int) int {
	for links[i] != i {
		i = links[i]
	}
	return i
}
func (links groupLinks) join(a, b int) {
	a, b = links.root(a), links.root(b)
	if a < b {
		links[b] = a
	} else {
		links[a] = b
	}
}
func (links groupLinks) components() [][]int {
	out := [][]int{}
	byRoot := map[int]int{}
	for i := range links {
		r := links.root(i)
		n, ok := byRoot[r]
		if !ok {
			n = len(out)
			byRoot[r] = n
			out = append(out, []int{})
		}
		out[n] = append(out[n], i)
	}
	return out
}

func (b *groupBuilder) sortEvaluations(evaluations []RuleEvaluation) {
	rank := map[string]int{}
	for i, r := range b.rules {
		rank[r.ID] = i
	}
	sort.SliceStable(evaluations, func(i, j int) bool {
		a, c := evaluations[i], evaluations[j]
		if rank[a.RuleID] != rank[c.RuleID] {
			return rank[a.RuleID] < rank[c.RuleID]
		}
		if a.Location == nil || c.Location == nil {
			return a.Location != nil
		}
		if a.Location.Start.Offset != c.Location.Start.Offset {
			return a.Location.Start.Offset < c.Location.Start.Offset
		}
		return a.Location.End.Offset < c.Location.End.Offset
	})
}

func copyGroupEvaluations(input []RuleEvaluation) []RuleEvaluation {
	output := append([]RuleEvaluation{}, input...)
	var copyCondition func(ConditionEvaluation) ConditionEvaluation
	copyCondition = func(c ConditionEvaluation) ConditionEvaluation {
		c.ReferenceIDs = append([]string{}, c.ReferenceIDs...)
		children := make([]ConditionEvaluation, len(c.Children))
		for i, child := range c.Children {
			children[i] = copyCondition(child)
		}
		c.Children = children
		return c
	}
	for i := range output {
		e := &output[i]
		e.ReferenceIDs = append([]string{}, e.ReferenceIDs...)
		if e.Location != nil {
			loc := *e.Location
			e.Location = &loc
		}
		if e.Condition != nil {
			c := copyCondition(*e.Condition)
			e.Condition = &c
		}
	}
	return output
}
