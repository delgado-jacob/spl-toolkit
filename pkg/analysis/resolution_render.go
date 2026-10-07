package analysis

import (
	"fmt"
	"sort"
	"unicode/utf8"
)

// ResolutionRendering retains original-source typed edits and session provenance.
// Reporting values returned by its accessors cannot establish correspondence.
type ResolutionRendering struct {
	session     *ResolutionSession
	rewrite     *RewriteRendering
	candidate   QueryDocument
	choices     []ResolutionChoice
	changes     []ResolutionChange
	limitations []ResolutionLimitation
}

func (r *ResolutionRendering) CandidateDocument() QueryDocument {
	if r == nil {
		return QueryDocument{}
	}
	return r.candidate
}

func (r *ResolutionRendering) Changes() []ResolutionChange {
	if r == nil {
		return []ResolutionChange{}
	}
	return rewriteCopy(r.changes)
}

// Render encodes each choice as a logical atom at every proven original owner.
// All edits are rendered together, so replacement bytes are never rediscovered.
func (s *ResolutionSession) Render(choices []ResolutionChoice) (*ResolutionRendering, error) {
	if s == nil || s.rewrite == nil || s.rewrite.result == nil || s.trace == nil {
		return nil, fmt.Errorf("missing resolution session")
	}
	groups := map[string]ResolutionPlaceholder{}
	kinds := map[string]string{}
	for _, group := range s.evidence.Placeholders {
		if kind, known := kinds[group.Placeholder]; known && kind != group.Kind {
			return nil, fmt.Errorf("placeholder %q has conflicting kinds", group.Placeholder)
		}
		kinds[group.Placeholder] = group.Kind
		groups[resolutionGroupKey(group.Placeholder, group.Kind)] = group
	}
	selected := map[string]ResolutionChoice{}
	for _, choice := range choices {
		key := resolutionGroupKey(choice.Placeholder, choice.Kind)
		if _, known := groups[key]; !known {
			return nil, fmt.Errorf("unknown resolution choice %q (%s)", choice.Placeholder, choice.Kind)
		}
		if _, duplicate := selected[key]; duplicate {
			return nil, fmt.Errorf("duplicate resolution choice %q", choice.Placeholder)
		}
		if choice.Value == "" || !utf8.ValidString(choice.Value) {
			return nil, fmt.Errorf("invalid resolution value for %q", choice.Placeholder)
		}
		selected[key] = choice
	}
	if len(selected) != len(groups) {
		return nil, fmt.Errorf("resolution choices must select every placeholder")
	}
	r := &ResolutionRendering{session: s, choices: rewriteCopy(choices), changes: []ResolutionChange{}, limitations: []ResolutionLimitation{}}
	replacements := []RewriteReplacement{}
	bySite := map[string]ResolutionChoice{}
	sites := []*rewriteSite{}
	for _, group := range s.evidence.Placeholders {
		key := resolutionGroupKey(group.Placeholder, group.Kind)
		choice := selected[key]
		groupSites := s.sites[key]
		for _, loc := range group.Locations {
			proved := false
			for _, site := range groupSites {
				proved = proved || site.public.OwnerLocation == loc
			}
			if !proved {
				r.limitations = append(r.limitations, ResolutionLimitation{Code: "unproved_owner", Message: "Placeholder has no proven typed rendering owner", Location: loc})
			}
		}
		for _, site := range groupSites {
			if prior, exists := bySite[site.public.ID]; exists {
				if prior != choice {
					return nil, fmt.Errorf("inconsistent typed resolution site %q", site.public.ID)
				}
				continue
			}
			bySite[site.public.ID] = choice
			sites = append(sites, site)
			replacements = append(replacements, RewriteReplacement{SiteID: site.public.ID, Target: rewriteAtom(choice.Value)})
		}
	}
	sort.SliceStable(sites, func(i, j int) bool {
		a, b := sites[i].public.Location, sites[j].public.Location
		if a.Start.Offset != b.Start.Offset {
			return a.Start.Offset < b.Start.Offset
		}
		return a.End.Offset < b.End.Offset
	})
	for i, site := range sites {
		loc := site.public.Location
		text := s.rewrite.result.Document.Text
		if loc.Start.Offset < 0 || loc.End.Offset < loc.Start.Offset || loc.End.Offset > len(text) {
			return nil, fmt.Errorf("invalid typed resolution range")
		}
		if i > 0 && loc.Start.Offset < sites[i-1].public.Location.End.Offset {
			prior := sites[i-1]
			if loc != prior.public.Location || bySite[site.public.ID] != bySite[prior.public.ID] {
				return nil, fmt.Errorf("overlapping typed resolution sites")
			}
		}
	}
	rendered, err := s.rewrite.Render(replacements)
	if err != nil {
		return nil, err
	}
	r.rewrite = rendered
	for _, requirement := range rendered.requirements {
		for _, limitation := range requirement.Limitations {
			r.limitations = append(r.limitations, ResolutionLimitation{Code: limitation.Code, Message: limitation.Message, Location: limitation.Location})
		}
		for _, required := range requirement.RequiredChanges {
			choice, known := bySite[required.SiteID]
			if !known || !rewriteIdentityEqual(required.Target, rewriteAtom(choice.Value)) {
				r.limitations = append(r.limitations, ResolutionLimitation{Code: "render_unproved", Message: "Typed rendering requires additional identity changes", Location: resolutionSitesLocation(sites, requirement.CauseSiteIDs)})
			}
		}
	}
	previousEnd := 0
	for _, edit := range rendered.edits {
		loc := edit.Location
		if loc.Start.Offset < previousEnd || loc.Start.Offset < 0 || loc.End.Offset < loc.Start.Offset || loc.End.Offset > len(s.rewrite.result.Document.Text) || s.rewrite.result.Document.Text[loc.Start.Offset:loc.End.Offset] != edit.Before {
			return nil, fmt.Errorf("invalid canonical resolution edit")
		}
		previousEnd = loc.End.Offset
		var choice ResolutionChoice
		references := []string{}
		for i, id := range edit.SiteIDs {
			current, known := bySite[id]
			if !known {
				return nil, fmt.Errorf("canonical edit has unknown resolution site")
			}
			if i > 0 && current != choice {
				return nil, fmt.Errorf("canonical edit has inconsistent resolution choices")
			}
			choice = current
			for _, site := range sites {
				if site.public.ID == id {
					references = uniqueIDs(references, []string{site.public.ReferenceID})
				}
			}
		}
		if len(edit.SiteIDs) == 0 {
			return nil, fmt.Errorf("canonical edit has no resolution site")
		}
		r.changes = append(r.changes, ResolutionChange{Placeholder: choice.Placeholder, Kind: choice.Kind, Value: choice.Value, OriginalReferenceIDs: references, CandidateReferenceIDs: []string{}, OriginalLocation: loc, Before: edit.Before, After: edit.After})
	}
	r.candidate = s.rewrite.result.Document
	r.candidate.Text = rewriteCandidateText(r.candidate.Text, rendered.edits)
	return r, nil
}

func resolutionSitesLocation(sites []*rewriteSite, ids []string) Location {
	for _, id := range ids {
		for _, site := range sites {
			if site.public.ID == id {
				return site.public.Location
			}
		}
	}
	return Location{}
}
