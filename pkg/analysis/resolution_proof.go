package analysis

import (
	"fmt"
	"reflect"
)

// ResolutionProof is session-bound substitution authority. Reporting values,
// including a JSON-decoded proof, cannot populate its private provenance.
type ResolutionProof struct {
	session        *ResolutionSession
	rendering      *ResolutionRendering
	candidate      *RewriteSession
	candidateTrace *requirementTrace
	original       ResolutionEvidence
	resolved       ResolutionEvidence
	evidence       ResolutionProofEvidence
	authorized     bool
}

func emptyResolutionProofEvidence() ResolutionProofEvidence {
	return ResolutionProofEvidence{References: []ResolutionReferencePair{}, Roles: []ResolutionRole{}, Limitations: []ResolutionLimitation{}}
}

// Evidence returns a detached audit projection, never proof authority.
func (p *ResolutionProof) Evidence() ResolutionProofEvidence {
	if p == nil || p.session == nil {
		return emptyResolutionProofEvidence()
	}
	return rewriteCopy(p.evidence)
}

// AssessmentEvidence retains both canonical snapshots even when correspondence
// is refused. Only Verify can establish the returned authority boolean.
func (p *ResolutionProof) AssessmentEvidence() (ResolutionEvidence, ResolutionEvidence, []ResolutionRole, bool) {
	if p == nil || p.session == nil {
		return ResolutionEvidence{}, ResolutionEvidence{}, []ResolutionRole{}, false
	}
	return rewriteCopy(p.original), rewriteCopy(p.resolved), rewriteCopy(p.evidence.Roles), p.authorized && p.rendering != nil && p.candidate != nil && p.candidateTrace != nil
}

// Verify reconstructs the typed rendering and analyzes those exact bytes with
// canonical analysis semantics. Ordinary rewrite correspondence remains strict.
func (s *ResolutionSession) Verify(r *ResolutionRendering) (*ResolutionProof, error) {
	if s == nil || s.rewrite == nil || s.trace == nil || r == nil || r.session != s || r.rewrite == nil {
		return nil, fmt.Errorf("rendering does not belong to resolution session")
	}
	fresh, err := s.Render(r.choices)
	if err != nil {
		return nil, err
	}
	if fresh.candidate != r.candidate || !reflect.DeepEqual(fresh.rewrite.edits, r.rewrite.edits) || !reflect.DeepEqual(fresh.rewrite.effects, r.rewrite.effects) || !reflect.DeepEqual(fresh.rewrite.requirements, r.rewrite.requirements) || !reflect.DeepEqual(fresh.rewrite.changes, r.rewrite.changes) || !reflect.DeepEqual(fresh.limitations, r.limitations) {
		return nil, fmt.Errorf("resolution rendering differs from exact reconstruction")
	}
	candidate := &RewriteSession{sites: []*rewriteSite{}, preserveCanonicalResult: true, renderUnchangedIdentities: true}
	result, trace, err := analyzeRewriteWithTrace(fresh.candidate, nil, candidate)
	if err != nil {
		return nil, err
	}
	candidate.result = result
	p := &ResolutionProof{session: s, rendering: fresh, candidate: candidate, candidateTrace: trace.clone(), original: s.Evidence(), resolved: ResolutionEvidence{Analysis: rewriteCopy(*result), Placeholders: []ResolutionPlaceholder{}, Coverage: cloneInputCoverage(result.InputCoverage)}, evidence: emptyResolutionProofEvidence()}
	p.evidence.Limitations = append(p.evidence.Limitations, fresh.limitations...)
	correspondence := s.rewrite.Verify(candidate, fresh.rewrite)
	for _, lim := range correspondence.Limitations {
		p.evidence.Limitations = append(p.evidence.Limitations, ResolutionLimitation{lim.Code, lim.Message, lim.Location})
	}
	if !correspondence.Proven || len(fresh.limitations) > 0 {
		return p, nil
	}
	mapping := map[string]string{}
	for _, pair := range correspondence.References {
		mapping[pair.OriginalID] = pair.CandidateID
		p.evidence.References = append(p.evidence.References, ResolutionReferencePair{pair.OriginalID, pair.CandidateID})
	}
	edits := fresh.rewrite.edits
	source := newSourceIndex(fresh.candidate.Text)
	effects := map[string]RewriteIdentity{}
	for _, effect := range fresh.rewrite.effects {
		effects[effect.ReferenceID] = effect.After
	}
	fail := func(code, message string, loc Location) (*ResolutionProof, error) {
		p.evidence.Limitations = append(p.evidence.Limitations, ResolutionLimitation{code, message, loc})
		return p, nil
	}
	// Private trace facts preserve supplier occurrences even if canonical input or
	// requirement groups merge after substitution. Compare situated owners rather
	// than names or opaque logical input IDs.
	if !resolutionTraceEqual(s.trace, trace, mapping, effects, edits, source) {
		return fail("requirement_correspondence", "Candidate changes original requirement flags or situated suppliers", Location{})
	}
	used := map[string]bool{}
	for _, original := range s.evidence.Analysis.Inputs {
		role := ResolutionRole{OriginalInput: original, Occurrences: []ResolutionOccurrencePair{}, Requirements: []ResolutionRequirementPair{}, Coverage: cloneInputCoverage(original.Evidence)}
		for _, occurrence := range original.Occurrences {
			found := false
			for _, next := range result.Inputs {
				for _, candidateOccurrence := range next.Occurrences {
					if used[candidateOccurrence.ID] || !resolutionInputEqual(original, next, occurrence, effects) || !resolutionOccurrenceEqual(occurrence, candidateOccurrence, mapping, edits, source) {
						continue
					}
					if role.CandidateInput.ID != "" && role.CandidateInput.ID != next.ID {
						return fail("source_correspondence", "Original logical source splits into candidate inputs", occurrence.Location)
					}
					role.CandidateInput = next
					role.Occurrences = append(role.Occurrences, ResolutionOccurrencePair{original.ID, next.ID, occurrence.ID, candidateOccurrence.ID})
					used[candidateOccurrence.ID] = true
					found = true
					break
				}
				if found {
					break
				}
			}
			if !found {
				return fail("source_correspondence", "Original source occurrence has no unique candidate supplier", occurrence.Location)
			}
		}
		p.evidence.Roles = append(p.evidence.Roles, role)
	}
	count := 0
	for _, input := range result.Inputs {
		count += len(input.Occurrences)
	}
	if count != len(used) {
		return fail("source_correspondence", "Candidate adds or reassigns a source occurrence", Location{})
	}
	resolutionPairRequirements(p, mapping, edits, source)
	// Audit edit endpoints only after all reference and source correspondences
	// are established. Byte ranges describe the exact candidate text.
	auditedChanges := rewriteCopy(fresh.changes)
	for i := range auditedChanges {
		change := &auditedChanges[i]
		loc, ok := rewriteTranslatedLocation(change.OriginalLocation, edits, source)
		if !ok {
			return fail("change_coordinates", "Candidate edit endpoints are unproved", change.OriginalLocation)
		}
		change.CandidateLocation = &loc
		change.CandidateReferenceIDs = []string{}
		for _, id := range change.OriginalReferenceIDs {
			change.CandidateReferenceIDs = append(change.CandidateReferenceIDs, mapping[id])
		}
	}
	r.changes = auditedChanges
	p.authorized = true
	p.evidence.Proven = true
	return p, nil
}

func resolutionInputEqual(original, candidate QueryInput, occurrence InputOccurrence, effects map[string]RewriteIdentity) bool {
	if effect, ok := effects[occurrence.OriginalReferenceID]; ok {
		// A typed dataset parameter becomes one encoded identifier atom. This
		// is the only authorized change to the canonical source category.
		kindMatches := original.Kind == candidate.Kind
		if original.Kind == "named_placeholder" && original.Identity.Form == "parameter" {
			kindMatches = candidate.Kind == "explicit_dataset" && candidate.Identity.Form == "identifier"
		}
		return kindMatches && effect.Name != nil && candidate.Name == *effect.Name && candidate.Identity.Value == *effect.Name
	}
	if original.Kind != candidate.Kind {
		return false
	}
	return original.Name == candidate.Name && original.Identity == candidate.Identity
}

func resolutionOccurrenceEqual(a, b InputOccurrence, mapping map[string]string, edits []RewriteTextEdit, source *sourceIndex) bool {
	loc, ok := rewriteTranslatedLocation(a.Location, edits, source)
	if !ok || loc != b.Location || mapping[a.ReferenceID] != b.ReferenceID || mapping[a.OriginalReferenceID] != b.OriginalReferenceID || a.StageID != b.StageID || a.ScopeID != b.ScopeID || a.Alias != b.Alias || len(a.UseSiteLocations) != len(b.UseSiteLocations) || len(a.UseSiteReferenceIDs) != len(b.UseSiteReferenceIDs) {
		return false
	}
	for i, location := range a.UseSiteLocations {
		next, ok := rewriteTranslatedLocation(location, edits, source)
		if !ok || next != b.UseSiteLocations[i] {
			return false
		}
	}
	for i, id := range a.UseSiteReferenceIDs {
		if mapping[id] != b.UseSiteReferenceIDs[i] {
			return false
		}
	}
	return true
}

func resolutionOwnerEqual(a, b sourceOwner, mapping map[string]string, effects map[string]RewriteIdentity, edits []RewriteTextEdit, source *sourceIndex) bool {
	if a.unresolved != b.unresolved || !reflect.DeepEqual(a.identity, b.identity) || a.input.sourceID != b.input.sourceID {
		return false
	}
	original := QueryInput{Kind: a.input.kind, Name: a.input.name, Identity: a.input.identity}
	candidate := QueryInput{Kind: b.input.kind, Name: b.input.name, Identity: b.input.identity}
	return resolutionInputEqual(original, candidate, a.input.occurrence, effects) && resolutionOccurrenceEqual(a.input.occurrence, b.input.occurrence, mapping, edits, source)
}

func resolutionTraceEqual(a, b *requirementTrace, mapping map[string]string, effects map[string]RewriteIdentity, edits []RewriteTextEdit, source *sourceIndex) bool {
	if a == nil || b == nil || len(a.references) != len(b.references) {
		return false
	}
	used := map[int]bool{}
	for _, entry := range a.references {
		found := false
		for j, next := range b.references {
			if used[j] || mapping[entry.reference.ID] != next.reference.ID || !resolutionTraceFlagsEqual(entry, next) || !reflect.DeepEqual(entry.fieldIdentity, next.fieldIdentity) || len(entry.owners) != len(next.owners) {
				continue
			}
			ownersUsed := map[int]bool{}
			ownersMatch := true
			for _, owner := range entry.owners {
				ownerMatch := false
				for k, nextOwner := range next.owners {
					if resolutionTraceFlagsCollapsed(entry, next) && owner.unresolved && !nextOwner.unresolved {
						nextOwner.unresolved = true
					}
					if !ownersUsed[k] && resolutionOwnerEqual(owner, nextOwner, mapping, effects, edits, source) {
						ownersUsed[k] = true
						ownerMatch = true
						break
					}
				}
				if !ownerMatch {
					ownersMatch = false
					break
				}
			}
			if ownersMatch {
				used[j] = true
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// Source grouping can strengthen a candidate trace when several mapped
// original logical suppliers become one canonical input. This narrow exception
// retains the original flags for assessment; it never proves original ownership.
// The caller separately establishes every owner occurrence and field identity.
func resolutionTraceFlagsEqual(a, b requirementTraceReference) bool {
	if a.directExternal == b.directExternal && a.conditional == b.conditional && a.pathConditional == b.pathConditional {
		return true
	}
	return resolutionTraceFlagsCollapsed(a, b)
}

func resolutionTraceFlagsCollapsed(a, b requirementTraceReference) bool {
	if a.reference.Kind != "field" || a.reference.Resolution != "exact" || b.reference.Kind != "field" || b.reference.Resolution != "exact" || a.directExternal || !a.conditional || a.pathConditional || !b.directExternal || b.conditional || b.pathConditional || len(a.owners) < 2 {
		return false
	}
	_, _, originalProved := provedSourceOwner(a.owners)
	_, _, candidateProved := provedSourceOwner(b.owners)
	originalInputs := map[string]bool{}
	for _, owner := range a.owners {
		originalInputs[opaqueInputID(owner.input.kind, owner.input.identity)] = true
	}
	return !originalProved && candidateProved && len(originalInputs) > 1
}

func resolutionPairRequirements(p *ResolutionProof, mapping map[string]string, edits []RewriteTextEdit, source *sourceIndex) {
	for _, item := range p.original.Analysis.Requirements.Items {
		for _, occurrence := range item.Occurrences {
			// Unproved original ownership stays unowned. A candidate group becoming
			// proved cannot supply authority for an original role.
			if item.InputID == "" {
				if item.Kind == "field" {
					for i := range p.evidence.Roles {
						role := &p.evidence.Roles[i]
						if len(item.Ownership.CandidateInputIDs) == 0 || containsID(item.Ownership.CandidateInputIDs, role.OriginalInput.ID) {
							resolutionRolePartial(role, occurrence, "Original field supplier is unproved")
						}
					}
				}
				continue
			}
			for i := range p.evidence.Roles {
				role := &p.evidence.Roles[i]
				if role.OriginalInput.ID != item.InputID {
					continue
				}
				found := false
				loc, ok := rewriteTranslatedLocation(occurrence.Location, edits, source)
				if ok {
					for _, nextItem := range p.resolved.Analysis.Requirements.Items {
						if nextItem.InputID != role.CandidateInput.ID || nextItem.Kind != item.Kind || nextItem.Role != item.Role || nextItem.Resolution != item.Resolution {
							continue
						}
						for _, next := range nextItem.Occurrences {
							if next.ReferenceID == mapping[occurrence.ReferenceID] && next.Location == loc && next.ScopeID == occurrence.ScopeID && next.StageID == occurrence.StageID && next.Binding == occurrence.Binding && next.Necessity == occurrence.Necessity {
								role.Requirements = append(role.Requirements, ResolutionRequirementPair{item.ID, nextItem.ID, occurrence, next})
								found = true
								break
							}
						}
						if found {
							break
						}
					}
				}
				if !found {
					resolutionRolePartial(role, occurrence, "Original requirement has no proved candidate occurrence")
				}
			}
		}
	}
}

func containsID(ids []string, id string) bool {
	for _, value := range ids {
		if value == id {
			return true
		}
	}
	return false
}
func resolutionRolePartial(role *ResolutionRole, occurrence RequirementOccurrence, message string) {
	role.Coverage.State = "partial"
	role.Coverage.Reasons = append(role.Coverage.Reasons, InputReason{Code: "resolution_role_incomplete", Message: message, Location: occurrence.Location, ScopeID: occurrence.ScopeID, StageID: occurrence.StageID, ReferenceIDs: []string{occurrence.ReferenceID}})
}
