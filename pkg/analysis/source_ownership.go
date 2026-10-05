package analysis

import "sort"

// Source facts travel with the existing field lineage. Neither schema evidence
// nor field spelling can select one of multiple situated source candidates.
type sourceOwner struct {
	input      inputFact
	identity   fieldIdentity
	unresolved bool
}

type InputOwnership struct {
	State             string   `json:"state"`
	CandidateInputIDs []string `json:"candidate_input_ids"`
}

func cloneSourceOwners(in []sourceOwner) []sourceOwner {
	if in == nil {
		return nil
	}
	out := append([]sourceOwner{}, in...)
	for i := range out {
		out[i].input = cloneInputFacts([]inputFact{in[i].input})[0]
		out[i].identity = cloneRequirementTraceIdentity(in[i].identity)
	}
	return out
}
func mergeSourceOwners(base []sourceOwner, additions ...[]sourceOwner) []sourceOwner {
	out := cloneSourceOwners(base)
	seen := map[string]bool{}
	key := func(owner sourceOwner) string {
		identity, _ := owner.identity.privateKey()
		return orderedStringSliceKey([]string{inputOccurrenceID(owner.input), identity, map[bool]string{true: "candidate", false: "origin"}[owner.unresolved]})
	}
	for _, owner := range out {
		seen[key(owner)] = true
	}
	for _, addition := range additions {
		for _, owner := range cloneSourceOwners(addition) {
			if !seen[key(owner)] {
				seen[key(owner)] = true
				out = append(out, owner)
			}
		}
	}
	return out
}
func sourceOwnersForIdentity(inputs []inputFact, identity fieldIdentity) []sourceOwner {
	owners := []sourceOwner{}
	for _, input := range inputs {
		if identity.Qualifier != "" && identity.Qualifier != input.occurrence.Alias {
			continue
		}
		relative := identity.clone()
		if relative.Qualifier != "" {
			relative.Qualifier = ""
			// A simple qualified join/source reference names an atomic field;
			// actual structural paths retain their typed segments.
			if len(relative.Segments) == 1 {
				relative = atomicFieldIdentity(relative.Segments[0])
			}
		}
		owners = append(owners, sourceOwner{input: cloneInputFacts([]inputFact{input})[0], identity: relative})
	}
	return owners
}
func (e *requirementEnvironment) sourceOwners(identity fieldIdentity) []sourceOwner {
	if field, known := e.field(identity); known {
		owners := cloneSourceOwners(field.owners)
		if field.ownerCollision || field.conditional && !field.source {
			for i := range owners {
				owners[i].unresolved = true
			}
		}
		return owners
	}
	return sourceOwnersForIdentity(e.inputs, identity)
}
func traceSourceOwners(trace *requirementTrace, ids []string) []sourceOwner {
	var out []sourceOwner
	if trace == nil {
		return out
	}
	for _, id := range ids {
		if _, found := trace.pendingReferenceIndexes[id]; found {
			out = mergeSourceOwners(out, trace.reference(id).owners)
		}
	}
	return out
}
func provedSourceOwner(owners []sourceOwner) (string, fieldIdentity, bool) {
	if len(owners) == 0 {
		return "", fieldIdentity{}, false
	}
	first := owners[0]
	id := opaqueInputID(first.input.kind, first.input.identity)
	identityKey, _ := first.identity.privateKey()
	for _, owner := range owners {
		key, _ := owner.identity.privateKey()
		if owner.unresolved || owner.input.kind == "unresolved_source" || opaqueInputID(owner.input.kind, owner.input.identity) != id || identityKey != key {
			return "", fieldIdentity{}, false
		}
	}
	return id, cloneRequirementTraceIdentity(first.identity), true
}
func sourceOwnerKey(owners []sourceOwner) string {
	if id, identity, proved := provedSourceOwner(owners); proved {
		key, _ := identity.privateKey()
		return orderedStringSliceKey([]string{id, key})
	}
	keys := []string{}
	for _, owner := range owners {
		key, _ := owner.identity.privateKey()
		keys = append(keys, orderedStringSliceKey([]string{inputOccurrenceID(owner.input), key}))
	}
	sort.Strings(keys)
	return orderedStringSliceKey(keys)
}
func requirementSourceIdentity(entry requirementTraceReference) fieldIdentity {
	if entry.reference.Kind == "field" {
		if _, identity, proved := provedSourceOwner(entry.owners); proved {
			return identity
		}
		if entry.fieldIdentity.PublicName != "" {
			return entry.fieldIdentity
		}
	}
	return atomicFieldIdentity(entry.reference.NormalizedName)
}
func remapSourceOwnerReferences(owners []sourceOwner, mapping map[string]string) {
	for i := range owners { // Copy the remapped fact back: the occurrence itself is a value.
		facts := []inputFact{owners[i].input}
		remapInputReferences(facts, mapping)
		owners[i].input = facts[0]
	}
}
func remapSourceOwnerStages(owners []sourceOwner, mapping map[string]string) {
	for i := range owners {
		facts := []inputFact{owners[i].input}
		remapInputStages(facts, mapping)
		owners[i].input = facts[0]
	}
}
func (e *environment) situateSourceOwners(referenceID string, stage Stage, location Location, alias string) {
	e.requirements.inputs = situatedViewInputs(e.requirements.inputs, referenceID, stage, location, alias)
	situate := func(owners []sourceOwner) []sourceOwner {
		out := cloneSourceOwners(owners)
		for i := range out {
			out[i].input = situatedViewInputs([]inputFact{out[i].input}, referenceID, stage, location, alias)[0]
		}
		return out
	}
	for key, field := range e.fields {
		field.owners = situate(field.owners)
		e.fields[key] = field
	}
	for key, field := range e.requirements.fields {
		field.owners = situate(field.owners)
		e.requirements.fields[key] = field
	}
}
func projectedOwnerOccurrenceIDs(owners []sourceOwner, inputs []QueryInput) []string {
	ids := []string{}
	for _, owner := range owners {
		logical := opaqueInputID(owner.input.kind, owner.input.identity)
		for _, input := range inputs {
			if input.ID != logical {
				continue
			}
			for _, occurrence := range input.Occurrences {
				original := owner.input.occurrence
				if original.Location != occurrence.Location || original.OriginalReferenceID != occurrence.OriginalReferenceID || len(original.UseSiteLocations) > len(occurrence.UseSiteLocations) {
					continue
				}
				matches := true
				for i, location := range original.UseSiteLocations {
					if location != occurrence.UseSiteLocations[i] {
						matches = false
						break
					}
				}
				if matches {
					ids = uniqueIDs(ids, []string{occurrence.ID})
				}
			}
		}
	}
	return ids
}
func projectInputOwnership(owners []sourceOwner) (string, InputOwnership) {
	ownership := InputOwnership{State: "unproved", CandidateInputIDs: []string{}}
	for _, owner := range owners {
		ownership.CandidateInputIDs = uniqueIDs(ownership.CandidateInputIDs, []string{opaqueInputID(owner.input.kind, owner.input.identity)})
	}
	if id, _, proved := provedSourceOwner(owners); proved {
		ownership.State = "proved"
		return id, ownership
	}
	return "", ownership
}

func sourceOwnersProved(owners []sourceOwner) bool {
	_, _, proved := provedSourceOwner(owners)
	return proved
}

func requirementOwnerKey(entry requirementTraceReference) string {
	if len(entry.owners) == 0 && entry.reference.Kind == "field" {
		return entry.reference.ScopeID
	}
	return sourceOwnerKey(entry.owners)
}
