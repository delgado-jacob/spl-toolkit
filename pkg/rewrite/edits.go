package rewrite

import (
	"fmt"
	"sort"
	"strings"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
)

// reconstructEdits translates original byte intervals during one forward copy.
// Returned locations correspond to the caller's edit order. The canonical
// locator alone validates UTF-8 boundaries and computes lines and columns.
func reconstructEdits(original string, edits []analysis.RewriteTextEdit) (string, []analysis.Location, error) {
	order := make([]int, len(edits))
	ranges := make([]analysis.RewriteByteRange, len(edits))
	for i, e := range edits {
		order[i] = i
		ranges[i] = analysis.RewriteByteRange{Start: e.Location.Start.Offset, End: e.Location.End.Offset}
	}
	if _, err := analysis.LocateRewriteBytes(original, ranges); err != nil {
		return "", nil, err
	}
	sort.SliceStable(order, func(i, j int) bool {
		a, b := ranges[order[i]], ranges[order[j]]
		if a.Start != b.Start {
			return a.Start < b.Start
		}
		return a.End < b.End
	})
	candidateRanges := make([]analysis.RewriteByteRange, len(edits))
	var text strings.Builder
	cursor := 0
	for n, i := range order {
		e, r := edits[i], ranges[i]
		if r.Start < cursor || (n > 0 && r.Start == r.End && ranges[order[n-1]] == r) {
			return "", nil, fmt.Errorf("overlapping rewrite interval at %d", r.Start)
		}
		if original[r.Start:r.End] != e.Before {
			return "", nil, fmt.Errorf("rewrite old text mismatch at %d", r.Start)
		}
		text.WriteString(original[cursor:r.Start])
		start := text.Len()
		text.WriteString(e.After)
		candidateRanges[i] = analysis.RewriteByteRange{Start: start, End: text.Len()}
		cursor = r.End
	}
	text.WriteString(original[cursor:])
	candidate := text.String()
	locations, err := analysis.LocateRewriteBytes(candidate, candidateRanges)
	if err != nil {
		return "", nil, err
	}
	return candidate, locations, nil
}

func (b *groupBuilder) audit(result *rewriteCandidate) error {
	original := b.evidence.Analysis.Document.Text
	selectedEdits := result.Rendering.Edits()
	_, locations, err := reconstructEdits(original, selectedEdits)
	if err != nil {
		return err
	}
	pairs := map[string]string{}
	for _, pair := range result.Proof.References {
		pairs[pair.OriginalID] = pair.CandidateID
	}
	for _, g := range result.Groups {
		edits := []analysis.RewriteTextEdit{}
		if g.Reason == "" {
			for _, edit := range selectedEdits {
				belongs := false
				for _, c := range g.Replacements {
					belongs = belongs || hasID(edit.SiteIDs, c.SiteID)
				}
				if belongs {
					edits = append(edits, edit)
				}
			}
		} else if g.rendering != nil && g.Reason != ReasonOverlappingEdits {
			edits = g.rendering.Edits()
		} else {
			for _, member := range g.members {
				edits = append(edits, member.rendering.Edits()...)
			}
		}
		seenSites := map[string]bool{}
		start := len(result.Changes)
		for _, edit := range edits {
			loc := edit.Location
			change := Change{Outcome: "proposed", Reason: "matched", GroupID: g.ID, RuleIDs: []string{}, OriginalReferenceIDs: []string{}, CandidateReferenceIDs: []string{}, OriginalLocation: &loc, OldText: edit.Before, NewText: edit.After, CandidateApplied: g.Reason == ""}
			for _, siteID := range edit.SiteIDs {
				seenSites[siteID] = true
				for _, member := range g.members {
					// A co-rendered interval retains all component contributions;
					// contradictory alternatives retain only their own rules.
					if g.Reason != ReasonConflictingTargets && g.Reason != ReasonOverlappingEdits {
						change.RuleIDs = b.orderedRules(change.RuleIDs, member.rules[siteID])
						continue
					}
					for _, alternative := range member.rendering.Edits() {
						if alternative.Location == edit.Location && alternative.After == edit.After && hasID(alternative.SiteIDs, siteID) {
							change.RuleIDs = b.orderedRules(change.RuleIDs, member.rules[siteID])
						}
					}
				}
			}
			for _, ref := range b.evidence.Analysis.References {
				for _, siteID := range edit.SiteIDs {
					if b.sites[siteID].ReferenceID == ref.ID && !hasID(change.OriginalReferenceIDs, ref.ID) {
						change.OriginalReferenceIDs = append(change.OriginalReferenceIDs, ref.ID)
					}
				}
			}
			if g.Reason != "" {
				change.Outcome = "skipped"
				change.Reason = g.Reason
				if g.Reason == ReasonConflictingTargets || g.Reason == ReasonOverlappingEdits || g.Reason == ReasonBindingCollision {
					change.Outcome = "ambiguous"
				}
			} else {
				for i, selected := range selectedEdits {
					if selected.Location == edit.Location && selected.After == edit.After {
						candidateLoc := locations[i]
						change.CandidateLocation = &candidateLoc
						break
					}
				}
				for _, id := range change.OriginalReferenceIDs {
					if candidateID := pairs[id]; candidateID != "" {
						change.CandidateReferenceIDs = append(change.CandidateReferenceIDs, candidateID)
					}
				}
			}
			duplicate := false
			for i := start; i < len(result.Changes); i++ {
				old := &result.Changes[i]
				if *old.OriginalLocation == loc && old.OldText == change.OldText && old.NewText == change.NewText {
					old.RuleIDs = b.orderedRules(old.RuleIDs, change.RuleIDs)
					old.OriginalReferenceIDs = conditionReferenceIDs(append(old.OriginalReferenceIDs, change.OriginalReferenceIDs...))
					duplicate = true
					break
				}
			}
			if !duplicate {
				result.Changes = append(result.Changes, change)
			}
		}
		// A renderer refusal can have no byte edit. Retain its located attempted
		// member without inventing a frontend encoding for the unavailable target.
		for _, c := range g.Replacements {
			if seenSites[c.SiteID] {
				continue
			}
			site := b.sites[c.SiteID]
			loc := site.Location
			ids := []string{}
			for _, member := range g.members {
				ids = b.orderedRules(ids, member.rules[c.SiteID])
			}
			refs := []string{}
			if site.ReferenceID != "" {
				refs = append(refs, site.ReferenceID)
			}
			result.Changes = append(result.Changes, Change{Outcome: "skipped", Reason: g.Reason, GroupID: g.ID, RuleIDs: ids, OriginalReferenceIDs: refs, CandidateReferenceIDs: []string{}, OriginalLocation: &loc, OldText: original[loc.Start.Offset:loc.End.Offset]})
		}
	}
	sort.SliceStable(result.Changes, func(i, j int) bool {
		a, c := result.Changes[i], result.Changes[j]
		if a.OriginalLocation.Start.Offset != c.OriginalLocation.Start.Offset {
			return a.OriginalLocation.Start.Offset < c.OriginalLocation.Start.Offset
		}
		if a.OriginalLocation.End.Offset != c.OriginalLocation.End.Offset {
			return a.OriginalLocation.End.Offset < c.OriginalLocation.End.Offset
		}
		if a.NewText != c.NewText {
			return a.NewText < c.NewText
		}
		return a.GroupID < c.GroupID
	})
	return nil
}
