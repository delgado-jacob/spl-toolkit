package analysis

import (
	"encoding/json"
	"fmt"
	"sort"
	"unicode/utf8"
)

// Rewrite identities and handles belong to one original analysis call. Name is
// one logical atom; Path is a distinct identity, never a dotted-name shortcut.
type RewriteIdentity struct {
	Name *string
	Path []string
}
type RewriteFactProbe struct {
	Kind     string
	Identity RewriteIdentity
}
type RewriteReplacement struct {
	SiteID string
	Target RewriteIdentity
}
type RewriteEvidence struct {
	Analysis Result
	Sites    []RewriteSite
}
type RewriteSite struct {
	ID, ReferenceID, BindingID, SourceEpochID string
	Kind, Role, Eligibility                   string
	Identity                                  RewriteIdentity
	Location, OwnerLocation                   Location
	Point                                     RewritePoint
	Facts                                     []RewriteFactEvidence
	Limitations                               []RewriteLimitation
}
type RewritePoint struct {
	ScopeID, StageID, Phase string
	LineageIndex, Ordinal   int
}
type RewriteFactEvidence struct {
	ProbeIndex       int
	GuaranteedValues []RewriteScalar
	LiteralComplete  bool
	ReferenceState   string
	ReferenceIDs     []string
	Locations        []Location
	Limitations      []RewriteLimitation
}
type RewriteScalar struct {
	Kind  string
	Value json.RawMessage
}
type RewriteLimitation struct {
	Code, Message string
	Location      Location
}
type RewriteTextEdit struct {
	Location      Location
	Before, After string
	SiteIDs       []string
}
type RewriteRequirement struct {
	CauseSiteIDs    []string
	RequiredChanges []RewriteReplacement
	Limitations     []RewriteLimitation
}
type RewriteIdentityEffect struct {
	ReferenceID   string
	Before, After RewriteIdentity
}
type RewriteReferencePair struct{ OriginalID, CandidateID string }
type RewriteProof struct {
	Proven      bool
	References  []RewriteReferencePair
	Limitations []RewriteLimitation
}

// RewriteSession retains canonical typed ownership privately. Public snapshots
// and render values are copies and cannot add authority to a session.
type RewriteSession struct {
	result *Result
	source *sourceIndex
	probes []RewriteFactProbe
	sites  []*rewriteSite
	epoch  int
}
type rewriteSite struct {
	public   RewriteSite
	owner    rewriteOwner
	binding  string
	inputs   []string
	implicit bool
}
type rewriteOwner struct {
	identity      RewriteIdentity
	role          string
	location      Location
	modelLocation Location
	prefix        string
	component     bool
	implicit      string
}
type RewriteRendering struct {
	session      *RewriteSession
	edits        []RewriteTextEdit
	requirements []RewriteRequirement
	effects      []RewriteIdentityEffect
	changes      []RewriteReplacement
}

func rewriteCopy[T any](value T) T {
	data, _ := json.Marshal(value)
	var out T
	_ = json.Unmarshal(data, &out)
	return out
}
func rewriteAtom(name string) RewriteIdentity { return RewriteIdentity{Name: &name} }
func rewriteIdentityEqual(a, b RewriteIdentity) bool {
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
func rewriteValidIdentity(id RewriteIdentity) bool {
	if id.Name != nil {
		return id.Path == nil && *id.Name != "" && utf8.ValidString(*id.Name)
	}
	if len(id.Path) == 0 {
		return false
	}
	for _, part := range id.Path {
		if part == "" || !utf8.ValidString(part) {
			return false
		}
	}
	return true
}
func PrepareRewrite(document QueryDocument, probes []RewriteFactProbe) (*RewriteSession, error) {
	for i, p := range probes {
		if !rewriteValidIdentity(p.Identity) || !rewriteKind(p.Kind) {
			return nil, fmt.Errorf("invalid rewrite fact probe %d", i)
		}
	}
	s := &RewriteSession{probes: rewriteCopy(probes), sites: []*rewriteSite{}}
	result, err := analyzeRewrite(document, nil, s)
	if err != nil {
		return nil, err
	}
	s.result = result
	return s, nil
}
func (s *RewriteSession) Evidence() RewriteEvidence {
	if s == nil || s.result == nil {
		return RewriteEvidence{Sites: []RewriteSite{}}
	}
	sites := make([]RewriteSite, 0, len(s.sites))
	for _, site := range s.sites {
		sites = append(sites, site.public)
	}
	return rewriteCopy(RewriteEvidence{Analysis: *s.result, Sites: sites})
}
func (r *RewriteRendering) Edits() []RewriteTextEdit           { return rewriteCopy(r.edits) }
func (r *RewriteRendering) Requirements() []RewriteRequirement { return rewriteCopy(r.requirements) }
func (r *RewriteRendering) Effects() []RewriteIdentityEffect   { return rewriteCopy(r.effects) }
func rewriteKind(kind string) bool {
	switch kind {
	case "field", "index", "source", "sourcetype", "lookup", "dataset", "data_model":
		return true
	}
	return false
}
func (s *semanticStage) rewriteState() *rewriteFlow {
	if s.result.rewrite == nil {
		return nil
	}
	if s.env.rewrite == nil {
		c := s.result.rewrite
		c.epoch++
		s.env.rewrite = &rewriteFlow{complete: !s.env.uncertain && s.result.Stages[s.stage].SemanticComplete, epoch: fmt.Sprintf("source-%d", c.epoch), facts: map[string]rewriteFact{}, seen: map[string][]string{}}
	}
	return s.env.rewrite
}
func (s *semanticStage) rewriteReference(id string, operand locatedOperand, kind, role string) {
	c := s.result.rewrite
	if c == nil || id == "" || !rewriteKind(kind) {
		return
	}
	flow := s.rewriteState()
	st := s.result.Stages[s.stage]
	owner := operand.rewrite
	if owner.location.End.Offset == 0 {
		owner.location = operand.Location
	}
	if role == "null_test" && owner.role != "" && owner.role != "navigation" {
		owner.role = "null_test"
	}
	site := &rewriteSite{owner: owner, public: RewriteSite{ID: fmt.Sprintf("site-%d", len(c.sites)), ReferenceID: id, SourceEpochID: flow.epoch, Kind: kind, Role: owner.role, Eligibility: "ineligible", Identity: rewriteAtom(operand.Name), Location: operand.Location, OwnerLocation: owner.location, Point: RewritePoint{ScopeID: st.ScopeID, StageID: st.ID, Phase: s.rewritePhase, LineageIndex: len(s.result.Lineage), Ordinal: s.rewriteOrdinal}, Facts: []RewriteFactEvidence{}, Limitations: []RewriteLimitation{}}}
	s.rewriteOrdinal++
	if rewriteValidIdentity(owner.identity) {
		site.public.Identity = rewriteCopy(owner.identity)
	}
	if kind != "field" {
		site.binding = "not_applicable"
		site.public.BindingID = flow.epoch + ":" + kind + ":" + operand.Name
	}
	if operand.Resolution != "exact" {
		flow.complete = false
		site.public.Limitations = append(site.public.Limitations, RewriteLimitation{"dynamic_identity", "Wildcard or dynamic identity is not an exact rewrite operand", operand.Location})
	}
	if owner.role == "" || owner.role == "navigation" {
		flow.complete = false
		site.public.Limitations = append(site.public.Limitations, RewriteLimitation{"unproved_owner", "Canonical typed rendering ownership is unproved", operand.Location})
	}
	if kind != "field" && operand.Resolution == "exact" && owner.role != "" && owner.role != "navigation" {
		flow.seen[rewriteFactKey(kind, site.public.Identity)] = uniqueIDs(flow.seen[rewriteFactKey(kind, site.public.Identity)], []string{id})
	}
	site.public.Facts = s.rewriteFacts()
	c.sites = append(c.sites, site)
}
func (s *semanticStage) rewriteBinding(id string, binding string, inputs []string) {
	c := s.result.rewrite
	if c == nil || id == "" {
		return
	}
	for i := len(c.sites) - 1; i >= 0; i-- {
		site := c.sites[i]
		if site.public.ReferenceID != id {
			continue
		}
		site.binding = binding
		site.inputs = copyIDs(inputs)
		ref := s.result.References[len(s.result.References)-1]
		if binding == "source" {
			site.public.BindingID = site.public.SourceEpochID + ":field:" + ref.NormalizedName
			key := rewriteFactKey("field", site.public.Identity)
			flow := s.rewriteState()
			flow.seen[key] = uniqueIDs(flow.seen[key], []string{id})
			site.public.Facts = s.rewriteFacts()
		} else if binding == "derived" || binding == "indeterminate" {
			if len(ref.OriginReferenceIDs) > 0 {
				site.public.BindingID = ref.OriginReferenceIDs[0]
			}
		}
		if binding == "definition" {
			site.public.BindingID = id
			site.implicit = site.owner.implicit != ""
		}
		return
	}
}
func (c *RewriteSession) finalizeReferences(mapping map[string]string) {
	for _, site := range c.sites {
		p := &site.public
		p.ReferenceID = mapping[p.ReferenceID]
		if id, ok := mapping[p.BindingID]; ok {
			p.BindingID = id
		}
		for i, id := range site.inputs {
			site.inputs[i] = mapping[id]
		}
		for i := range p.Facts {
			f := &p.Facts[i]
			for j, id := range f.ReferenceIDs {
				f.ReferenceIDs[j] = mapping[id]
			}
			// Predicate guarantees can precede the physical reads in the same owner.
			for _, loc := range f.Locations {
				for _, other := range c.sites {
					if other.public.Point.ScopeID == p.Point.ScopeID && other.public.SourceEpochID == p.SourceEpochID && other.public.Location == loc {
						id := other.public.ReferenceID
						if mapped, ok := mapping[id]; ok {
							id = mapped
						}
						if id != "" {
							f.ReferenceIDs = uniqueIDs(f.ReferenceIDs, []string{id})
						}
					}
				}
			}
		}
		if len(p.Limitations) == 0 && (site.binding == "source" || (p.Kind != "field" && site.binding == "not_applicable")) {
			p.Eligibility = "eligible"
		}
		if p.Eligibility != "eligible" && len(p.Limitations) == 0 {
			p.Limitations = append(p.Limitations, RewriteLimitation{"binding_not_source", "Operand is not a selected source binding", p.Location})
		}
	}
}

// RewriteByteRange is a half-open interval of original UTF-8 bytes.
type RewriteByteRange struct{ Start, End int }

func LocateRewriteBytes(text string, ranges []RewriteByteRange) ([]Location, error) {
	if !utf8.ValidString(text) {
		return nil, fmt.Errorf("text is not valid UTF-8")
	}
	source := newSourceIndex(text)
	out := make([]Location, 0, len(ranges))
	for i, r := range ranges {
		loc, ok := source.byteLocation(r.Start, r.End)
		if !ok {
			return nil, fmt.Errorf("invalid rewrite byte range %d: [%d,%d)", i, r.Start, r.End)
		}
		out = append(out, loc)
	}
	return out, nil
}
func (s *sourceIndex) byteLocation(start, end int) (Location, bool) {
	if start < 0 || end < start || end > len(s.text) {
		return Location{}, false
	}
	a := sort.Search(len(s.positions), func(i int) bool { return s.positions[i].Offset >= start })
	b := sort.Search(len(s.positions), func(i int) bool { return s.positions[i].Offset >= end })
	if a == len(s.positions) || b == len(s.positions) || s.positions[a].Offset != start || s.positions[b].Offset != end {
		return Location{}, false
	}
	return Location{Start: s.positions[a], End: s.positions[b]}, true
}

func (s *semanticStage) rewriteDependencyRole(role string) {
	c := s.result.rewrite
	if c == nil || len(c.sites) == 0 {
		return
	}
	site := c.sites[len(c.sites)-1]
	site.owner.role, site.public.Role = role, role
}
func (s *semanticStage) rewriteUnprovedOperand(operand locatedOperand, kind, role, code string) {
	c := s.result.rewrite
	if c == nil {
		return
	}
	flow := s.rewriteState()
	flow.complete = false
	stage := s.result.Stages[s.stage]
	identity := RewriteIdentity{}
	if operand.Name != "" {
		identity = rewriteAtom(operand.Name)
	}
	c.sites = append(c.sites, &rewriteSite{public: RewriteSite{ID: fmt.Sprintf("site-%d", len(c.sites)), Kind: kind, Role: role, Eligibility: "ineligible", Identity: identity, Location: operand.Location, OwnerLocation: operand.Location, SourceEpochID: flow.epoch, Point: RewritePoint{ScopeID: stage.ScopeID, StageID: stage.ID, Phase: s.rewritePhase, LineageIndex: len(s.result.Lineage), Ordinal: s.rewriteOrdinal}, Facts: s.rewriteFacts(), Limitations: []RewriteLimitation{{code, "Canonical identity or rendering ownership is unproved", operand.Location}}}})
	s.rewriteOrdinal++
}
