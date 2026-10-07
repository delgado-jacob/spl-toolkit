package analysis

import (
	"regexp"
	"sort"
)

var resolutionMarker = regexp.MustCompile(`^\$[A-Za-z_][A-Za-z_0-9]*$`)

// PrepareResolution discovers whole markers in external typed owners of the
// submitted document. Definitions and ordinary string values are not owners.
func PrepareResolution(document QueryDocument) (*ResolutionSession, error) {
	rewrite := &RewriteSession{sites: []*rewriteSite{}, preserveCanonicalResult: true}
	result, trace, err := analyzeRewriteWithTrace(document, nil, rewrite)
	if err != nil {
		return nil, err
	}
	rewrite.result = result
	s := &ResolutionSession{rewrite: rewrite, trace: trace.clone(), sites: map[string][]*rewriteSite{}, evidence: ResolutionEvidence{Analysis: rewriteCopy(*result), Placeholders: []ResolutionPlaceholder{}, Coverage: cloneInputCoverage(result.InputCoverage)}}
	s.discover()
	return s, nil
}

// Evidence returns detached reporting values, never rewriting authority.
func (s *ResolutionSession) Evidence() ResolutionEvidence {
	if s == nil || s.rewrite == nil {
		return ResolutionEvidence{Placeholders: []ResolutionPlaceholder{}, Coverage: InputCoverage{State: "partial", Reasons: []InputReason{}}}
	}
	return rewriteCopy(s.evidence)
}

func resolutionKind(kind string) bool {
	switch kind {
	case "index", "source", "sourcetype", "lookup", "dataset", "data_model":
		return true
	}
	return false
}

// Conflicting kinds retain separate reporting groups so request admission can
// reject the marker's ambiguity without silently selecting one kind.
func resolutionGroupKey(marker, kind string) string { return kind + ":" + marker }

func (s *ResolutionSession) discover() {
	groups := map[string]int{}
	references := map[string]Reference{}
	inputIDs := map[string][]string{}
	for _, ref := range s.evidence.Analysis.References {
		references[ref.ID] = ref
	}
	for _, input := range s.evidence.Analysis.Inputs {
		for _, occurrence := range input.Occurrences {
			for _, id := range []string{occurrence.ReferenceID, occurrence.OriginalReferenceID} {
				if id != "" {
					inputIDs[id] = uniqueIDs(inputIDs[id], []string{input.ID})
				}
			}
		}
	}
	sites := append([]*rewriteSite{}, s.rewrite.sites...)
	sort.SliceStable(sites, func(i, j int) bool {
		return sites[i].public.OwnerLocation.Start.Offset < sites[j].public.OwnerLocation.Start.Offset
	})
	for _, site := range sites {
		p := site.public
		if !resolutionKind(p.Kind) {
			continue
		}
		// Local-view references resolve through the root external occurrence. Their
		// declaration identifiers cannot become substitution sites.
		if p.ReferenceID != "" {
			ref, known := references[p.ReferenceID]
			if !known || ref.Binding == "local_view" || ref.Binding == "local" {
				continue
			}
		}
		if p.Eligibility != "eligible" {
			s.evidence.Coverage.State = "partial"
			referenceIDs := []string{}
			if p.ReferenceID != "" {
				referenceIDs = append(referenceIDs, p.ReferenceID)
			}
			for _, limitation := range p.Limitations {
				s.evidence.Coverage.Reasons = append(s.evidence.Coverage.Reasons, InputReason{Code: limitation.Code, Message: limitation.Message, Location: limitation.Location, StageID: p.Point.StageID, ScopeID: p.Point.ScopeID, ReferenceIDs: copyIDs(referenceIDs)})
			}
		}
		if p.Identity.Name == nil || len(p.Identity.Path) > 0 || !resolutionMarker.MatchString(*p.Identity.Name) {
			continue
		}
		marker := *p.Identity.Name
		key := resolutionGroupKey(marker, p.Kind)
		index, known := groups[key]
		if !known {
			index = len(s.evidence.Placeholders)
			groups[key] = index
			s.evidence.Placeholders = append(s.evidence.Placeholders, ResolutionPlaceholder{Placeholder: marker, Kind: p.Kind, ReferenceIDs: []string{}, OriginalInputIDs: []string{}, Locations: []Location{}})
		}
		group := &s.evidence.Placeholders[index]
		if p.ReferenceID != "" {
			group.ReferenceIDs = uniqueIDs(group.ReferenceIDs, []string{p.ReferenceID})
		}
		found := false
		for _, loc := range group.Locations {
			found = found || loc == p.OwnerLocation
		}
		if !found {
			group.Locations = append(group.Locations, p.OwnerLocation)
		}
		group.OriginalInputIDs = uniqueIDs(group.OriginalInputIDs, inputIDs[p.ReferenceID])
		if p.Eligibility == "eligible" && p.ReferenceID != "" {
			s.sites[key] = append(s.sites[key], site)
		}
	}
}
