package analysis

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

func (s *RewriteSession) Render(changes []RewriteReplacement) (*RewriteRendering, error) {
	if s == nil || s.result == nil {
		return nil, fmt.Errorf("missing rewrite session")
	}
	// Validate caller bytes before the JSON copy can replace malformed UTF-8.
	for _, change := range changes {
		if !rewriteValidIdentity(change.Target) {
			return nil, fmt.Errorf("invalid rewrite replacement %q", change.SiteID)
		}
	}
	r := &RewriteRendering{session: s, edits: []RewriteTextEdit{}, requirements: []RewriteRequirement{}, effects: []RewriteIdentityEffect{}, changes: rewriteCopy(changes)}
	byID := map[string]*rewriteSite{}
	byRef := map[string]*rewriteSite{}
	selected := map[string]RewriteIdentity{}
	for _, site := range s.sites {
		byID[site.public.ID] = site
		byRef[site.public.ReferenceID] = site
	}
	for _, change := range r.changes {
		if byID[change.SiteID] == nil {
			return nil, fmt.Errorf("invalid rewrite replacement %q", change.SiteID)
		}
		if _, ok := selected[change.SiteID]; ok {
			return nil, fmt.Errorf("duplicate rewrite replacement %q", change.SiteID)
		}
		selected[change.SiteID] = change.Target
	}
	refuse := func(ids []string, code, message string, loc Location) {
		r.requirements = append(r.requirements, RewriteRequirement{CauseSiteIDs: copyIDs(ids), RequiredChanges: []RewriteReplacement{}, Limitations: []RewriteLimitation{{code, message, loc}}})
	}
	if !s.result.Coverage.SyntaxComplete || s.result.Status == Invalid {
		for _, c := range changes {
			refuse([]string{c.SiteID}, "original_not_valid", "Original document contains syntax or definite analysis errors", byID[c.SiteID].public.Location)
		}
		return r, nil
	}
	linked := map[string]RewriteIdentity{}
	for _, output := range s.sites {
		if !output.implicit {
			continue
		}
		causes := []string{}
		newName := ""
		for _, id := range output.inputs {
			input := byRef[id]
			if input == nil {
				continue
			}
			target, ok := selected[input.public.ID]
			if !ok || rewriteIdentityEqual(target, input.public.Identity) {
				continue
			}
			causes = append(causes, input.public.ID)
			if target.Name != nil {
				newName = output.owner.implicit + "(" + *target.Name + ")"
			}
		}
		if len(causes) == 0 {
			continue
		}
		if s.result.Document.Language == "spl2" || newName == "" || len(output.inputs) != 1 {
			refuse(causes, "implicit_name_unproved", "Changed implicit output name is outside the canonical naming contract", output.public.Location)
			continue
		}
		r.effects = append(r.effects, RewriteIdentityEffect{ReferenceID: output.public.ReferenceID, Before: output.public.Identity, After: rewriteAtom(newName)})
		required := []RewriteReplacement{}
		for _, consumer := range s.sites {
			if consumer == output || consumer.public.BindingID != output.public.ReferenceID {
				continue
			}
			target := rewriteAtom(newName)
			linked[consumer.public.ID] = target
			required = append(required, RewriteReplacement{consumer.public.ID, target})
			if consumer.binding != "derived" || consumer.owner.role == "" {
				refuse(causes, "implicit_consumer_unproved", "Implicit output consumer has an indeterminate binding or owner", consumer.public.Location)
			}
		}
		if len(required) > 0 {
			r.requirements = append(r.requirements, RewriteRequirement{CauseSiteIDs: causes, RequiredChanges: required, Limitations: []RewriteLimitation{}})
		}
	}
	for _, site := range s.sites {
		target, ok := selected[site.public.ID]
		if !ok || rewriteIdentityEqual(target, site.public.Identity) {
			continue
		}
		ids := []string{site.public.ID}
		p := site.public
		if expected, ok := linked[p.ID]; p.Eligibility != "eligible" && (!ok || !rewriteIdentityEqual(expected, target) || site.binding != "derived") {
			refuse(ids, "binding_not_source", "Original operand is not an eligible source or required implicit consumer", p.Location)
			continue
		}
		if target.Name == nil {
			refuse(ids, "target_not_renderable", "Static path rendering is not established for this atom owner", p.Location)
			continue
		}
		name := *target.Name
		before := s.result.Document.Text[p.Location.Start.Offset:p.Location.End.Offset]
		if site.owner.prefix != "" {
			parts := strings.Split(name, ".")
			if len(parts) != 2 || !plainCatalogComponent(parts[0]) || !plainCatalogComponent(parts[1]) {
				refuse(ids, "target_not_renderable", "Dataset target needs exact model and dataset components", p.Location)
				continue
			}
			if parts[0]+"." != site.owner.prefix {
				var model *rewriteSite
				for _, other := range s.sites {
					if other.public.Kind == "data_model" && other.public.Location == site.owner.modelLocation {
						model = other
						break
					}
				}
				if model == nil {
					refuse(ids, "target_not_renderable", "Dataset model owner is unproved", p.Location)
					continue
				}
				r.requirements = append(r.requirements, RewriteRequirement{CauseSiteIDs: ids, RequiredChanges: []RewriteReplacement{{model.public.ID, rewriteAtom(parts[0])}}, Limitations: []RewriteLimitation{}})
			}
			name = parts[1]
		}
		var after string
		var renderable bool
		if p.Role == "qualified_dataset" {
			parts := strings.Split(name, ".")
			renderable = len(parts) == 2 && plainCatalogComponent(parts[0]) && plainCatalogComponent(parts[1]) && !strings.ContainsAny(name, " \t\r\n|'\"")
			after = name
		} else if s.result.Document.Language == "spl2" {
			after, renderable = rewriteSPL2Text(site, name, before)
		} else {
			after, renderable = rewriteSPLText(site, name, before)
		}
		if !renderable {
			refuse(ids, "target_not_renderable", "Target has no proven encoding in the original grammar role", p.Location)
			continue
		}
		loc := p.Location
		if p.Role == "lookup_dual" {
			loc.Start = loc.End
			before = ""
			after = " AS " + after
		}
		r.edits = append(r.edits, RewriteTextEdit{Location: loc, Before: before, After: after, SiteIDs: ids})
		r.effects = append(r.effects, RewriteIdentityEffect{ReferenceID: p.ReferenceID, Before: p.Identity, After: target})
	}
	sort.SliceStable(r.edits, func(i, j int) bool {
		a, b := r.edits[i].Location, r.edits[j].Location
		if a.Start.Offset != b.Start.Offset {
			return a.Start.Offset < b.Start.Offset
		}
		return a.End.Offset > b.End.Offset
	})
	// Overlapping catalog components are independently typed. Compatible edits
	// co-render once; the retained semantic effect accounts for each reference.
	edits := []RewriteTextEdit{}
	for _, edit := range r.edits {
		if len(edits) > 0 {
			last := &edits[len(edits)-1]
			if edit.Location.Start.Offset < last.Location.End.Offset {
				delta := edit.Location.Start.Offset - last.Location.Start.Offset
				end := delta + len(edit.Before)
				compatible := false
				if delta >= 0 && end <= len(last.Before) && delta+len(edit.After) <= len(last.After) {
					compatible = last.After[delta:delta+len(edit.After)] == edit.After
				}
				if compatible {
					last.SiteIDs = uniqueIDs(last.SiteIDs, edit.SiteIDs)
					continue
				}
				refuse(uniqueIDs(last.SiteIDs, edit.SiteIDs), "render_conflict", "Overlapping typed component proposals disagree", edit.Location)
				continue
			}
		}
		edits = append(edits, edit)
	}
	r.edits = edits
	// A whole qualified dataset owns its model component as well.
	for _, dataset := range s.sites {
		target, ok := selected[dataset.public.ID]
		if !ok || target.Name == nil || dataset.public.Role != "qualified_dataset" {
			continue
		}
		parts := strings.Split(*target.Name, ".")
		if len(parts) != 2 {
			continue
		}
		for _, model := range s.sites {
			if model.public.Kind == "data_model" && model.public.OwnerLocation == dataset.public.OwnerLocation {
				if _, ok := selected[model.public.ID]; !ok {
					r.effects = append(r.effects, RewriteIdentityEffect{ReferenceID: model.public.ReferenceID, Before: model.public.Identity, After: rewriteAtom(parts[0])})
				}
			}
		}
	}
	// A component replacement can normalize an enclosing dependency even when
	// that dependency has no independent editable token.
	for _, site := range s.sites {
		if site.public.Kind != "dataset" && site.public.Kind != "data_model" {
			continue
		}
		if _, ok := selected[site.public.ID]; ok {
			continue
		}
		p := site.public
		value := s.result.Document.Text[p.Location.Start.Offset:p.Location.End.Offset]
		changed := false
		for i := len(r.edits) - 1; i >= 0; i-- {
			e := r.edits[i]
			if e.Location.Start.Offset >= p.Location.Start.Offset && e.Location.End.Offset <= p.Location.End.Offset {
				a, b := e.Location.Start.Offset-p.Location.Start.Offset, e.Location.End.Offset-p.Location.Start.Offset
				value = value[:a] + e.After + value[b:]
				changed = true
			}
		}
		if site.owner.prefix != "" && p.Identity.Name != nil {
			// Separate datamodel operands retain a decoded composite identity;
			// their original dataset token may include grammar quotes.
			value = strings.TrimPrefix(*p.Identity.Name, site.owner.prefix)
			for _, model := range s.sites {
				if model.public.Kind == "data_model" && model.public.Location == site.owner.modelLocation {
					if target, ok := selected[model.public.ID]; ok && target.Name != nil {
						value = *target.Name + "." + value
						changed = true
					} else {
						value = site.owner.prefix + value
					}
					break
				}
			}
		}
		if changed && site.owner.component {
			r.effects = append(r.effects, RewriteIdentityEffect{ReferenceID: p.ReferenceID, Before: p.Identity, After: rewriteAtom(value)})
		}
	}
	return r, nil
}

func rewriteCandidateText(text string, edits []RewriteTextEdit) string {
	for i := len(edits) - 1; i >= 0; i-- {
		e := edits[i]
		text = text[:e.Location.Start.Offset] + e.After + text[e.Location.End.Offset:]
	}
	return text
}
func rewriteTranslatedOffset(offset int, edits []RewriteTextEdit, end bool) int {
	delta := 0
	for _, edit := range edits {
		a, b := edit.Location.Start.Offset, edit.Location.End.Offset
		if offset < a {
			break
		}
		if a == b && offset == a {
			if end {
				delta += len(edit.After)
			}
			continue
		}
		if offset == a {
			return a + delta
		}
		if offset < b {
			if end {
				return a + delta + len(edit.After)
			}
			return a + delta
		}
		delta += len(edit.After) - (b - a)
	}
	return offset + delta
}
func rewriteTranslatedLocation(loc Location, edits []RewriteTextEdit, source *sourceIndex) (Location, bool) {
	return source.byteLocation(rewriteTranslatedOffset(loc.Start.Offset, edits, false), rewriteTranslatedOffset(loc.End.Offset, edits, true))
}
func (s *RewriteSession) Verify(candidate *RewriteSession, rendered *RewriteRendering) RewriteProof {
	proof := RewriteProof{References: []RewriteReferencePair{}, Limitations: []RewriteLimitation{}}
	fail := func(code, message string, loc Location) RewriteProof {
		proof.Limitations = append(proof.Limitations, RewriteLimitation{code, message, loc})
		return proof
	}
	if s == nil || s.result == nil || candidate == nil || candidate.result == nil || rendered == nil || rendered.session != s {
		return fail("render_provenance", "Rendering does not belong to the original session", Location{})
	}
	a, b := s.result, candidate.result
	if a.Document.Language != b.Document.Language || a.Document.Profile != b.Document.Profile || a.Document.Version != b.Document.Version || a.Document.SourceID != b.Document.SourceID || rewriteCandidateText(a.Document.Text, rendered.edits) != b.Document.Text {
		return fail("candidate_mismatch", "Candidate document does not match the original rendering", Location{})
	}
	selected := map[string]RewriteIdentity{}
	for _, c := range rendered.changes {
		selected[c.SiteID] = c.Target
	}
	for _, r := range rendered.requirements {
		if len(r.Limitations) > 0 {
			return fail("render_unproved", "Rendering contains an unproved required effect", r.Limitations[0].Location)
		}
		for _, c := range r.RequiredChanges {
			if !rewriteIdentityEqual(c.Target, selected[c.SiteID]) {
				return fail("required_change_missing", "A linked original consumer change is missing", Location{})
			}
		}
	}
	if !a.Coverage.SyntaxComplete || a.Status == Invalid || !b.Coverage.SyntaxComplete || b.Status == Invalid {
		return fail("candidate_not_valid", "Original or candidate has syntax or definite analysis errors", Location{})
	}
	effects := map[string]RewriteIdentity{}
	for _, e := range rendered.effects {
		effects[e.ReferenceID] = e.After
	}
	source := newSourceIndex(b.Document.Text)
	mapping := map[string]string{}
	used := map[string]bool{}
	originalSites := map[string]*rewriteSite{}
	for _, site := range s.sites {
		originalSites[site.public.ReferenceID] = site
	}
	candidateSites := map[string]*rewriteSite{}
	for _, site := range candidate.sites {
		candidateSites[site.public.ReferenceID] = site
	}
	for _, ref := range a.References {
		loc, ok := s.rewriteReferenceLocation(ref, rendered, source)
		if !ok {
			return fail("reference_coordinates", "Reference endpoint translation is unproved", ref.Location)
		}
		if site := originalSites[ref.ID]; site != nil && site.public.Role == "lookup_dual" {
			for _, edit := range rendered.edits {
				if len(edit.SiteIDs) == 1 && edit.SiteIDs[0] == site.public.ID {
					start := rewriteTranslatedOffset(edit.Location.Start.Offset, rendered.edits, false) + len(" AS ")
					loc, ok = source.byteLocation(start, start+len(edit.After)-len(" AS "))
				}
			}
		}
		name := ref.NormalizedName
		if effect, ok := effects[ref.ID]; ok {
			if effect.Name == nil {
				return fail("identity_unproved", "Path correspondence is unproved", ref.Location)
			}
			name = *effect.Name
		}
		found := ""
		for _, next := range b.References {
			if !used[next.ID] && next.Location == loc && next.NormalizedName == name && next.Kind == ref.Kind && next.Role == ref.Role && next.ScopeID == ref.ScopeID && next.StageID == ref.StageID && next.Binding == ref.Binding && next.Resolution == ref.Resolution {
				found = next.ID
				break
			}
		}
		if found == "" {
			return fail("reference_correspondence", "Candidate changed typed reference identity, ownership, role or binding", ref.Location)
		}
		mapping[ref.ID] = found
		used[found] = true
		proof.References = append(proof.References, RewriteReferencePair{ref.ID, found})
	}
	if !s.rewriteUnknownOwnersEqual(candidate, rendered, source) {
		return fail("unknown_correspondence", "Candidate changes an unproved typed owner", Location{})
	}
	if len(mapping) != len(b.References) {
		return fail("reference_correspondence", "Candidate adds or removes canonical references", Location{})
	}
	for _, ref := range a.References {
		var next Reference
		for _, r := range b.References {
			if r.ID == mapping[ref.ID] {
				next = r
				break
			}
		}
		ids := []string{}
		for _, id := range ref.OriginReferenceIDs {
			ids = append(ids, mapping[id])
		}
		if !reflect.DeepEqual(ids, next.OriginReferenceIDs) {
			return fail("origin_correspondence", "Candidate changes canonical origins", ref.Location)
		}
		oldSite, newSite := originalSites[ref.ID], candidateSites[mapping[ref.ID]]
		if oldSite != nil && newSite != nil {
			oldPoint, newPoint := oldSite.public.Point, newSite.public.Point
			ownerLoc, ownerOK := rewriteTranslatedLocation(oldSite.public.OwnerLocation, rendered.edits, source)
			if !ownerOK || ownerLoc != newSite.public.OwnerLocation {
				return fail("owner_correspondence", "Candidate changes typed operand ownership", ref.Location)
			}
			if oldPoint != newPoint {
				return fail("phase_correspondence", "Candidate changes canonical flow point", ref.Location)
			}
			if oldSite.binding == "derived" && mapping[oldSite.public.BindingID] != newSite.public.BindingID {
				return fail("binding_correspondence", "Candidate changes a direct alias binding", ref.Location)
			}
		}
	}
	if len(a.Stages) != len(b.Stages) || len(a.Scopes) != len(b.Scopes) || len(a.Lineage) != len(b.Lineage) || len(a.Diagnostics) != len(b.Diagnostics) {
		return fail("structure_correspondence", "Candidate changes canonical scope, stage, phase or diagnostic structure", Location{})
	}
	for i, stage := range a.Stages {
		loc, ok := rewriteTranslatedLocation(stage.Location, rendered.edits, source)
		next := b.Stages[i]
		if !ok || loc != next.Location || stage.ID != next.ID || stage.ScopeID != next.ScopeID || stage.Command != next.Command || stage.SemanticComplete != next.SemanticComplete {
			return fail("stage_correspondence", "Candidate stage ownership or coverage changed", stage.Location)
		}
	}
	for i, scope := range a.Scopes {
		loc, ok := rewriteTranslatedLocation(scope.Location, rendered.edits, source)
		next := b.Scopes[i]
		if !ok || loc != next.Location || scope.ID != next.ID || scope.ParentID != next.ParentID || scope.StageID != next.StageID || scope.Kind != next.Kind {
			return fail("scope_correspondence", "Candidate scope ownership changed", scope.Location)
		}
	}
	for i, diag := range a.Diagnostics {
		loc, ok := rewriteTranslatedLocation(diag.Location, rendered.edits, source)
		next := b.Diagnostics[i]
		if !ok || loc != next.Location || diag.Code != next.Code || diag.Category != next.Category || diag.Severity != next.Severity || diag.StageID != next.StageID || diag.ScopeID != next.ScopeID || diag.Message != next.Message {
			return fail("diagnostic_correspondence", "Candidate changes a located unknown or diagnostic effect", diag.Location)
		}
	}
	for i, lineage := range a.Lineage {
		next := b.Lineage[i]
		if lineage.StageID != next.StageID || lineage.ScopeID != next.ScopeID || lineage.Phase != next.Phase || !reflect.DeepEqual(lineage.ExecutionOrder, next.ExecutionOrder) || len(lineage.Transitions) != len(next.Transitions) {
			return fail("transition_correspondence", "Candidate changes canonical phase scheduling", Location{})
		}
		if !s.rewriteFieldStateEqual(lineage.Before, next.Before, mapping, effects, lineage.ScopeID, i, false) || !s.rewriteFieldStateEqual(lineage.After, next.After, mapping, effects, lineage.ScopeID, i, true) {
			return fail("field_state_correspondence", "Candidate changes canonical field membership, origins or conditionality", Location{})
		}
		for j, tr := range lineage.Transitions {
			n := next.Transitions[j]
			inputs := []string{}
			for _, id := range tr.InputReferenceIDs {
				inputs = append(inputs, mapping[id])
			}
			output := tr.Output
			if effect, ok := effects[tr.OutputReferenceID]; ok && effect.Name != nil {
				output = *effect.Name
			} else if tr.Operation == "project" && len(tr.InputReferenceIDs) == 1 {
				if effect, ok := effects[tr.InputReferenceIDs[0]]; ok && effect.Name != nil {
					output = *effect.Name
				}
			}
			if tr.Operation != n.Operation || output != n.Output || tr.Conditional != n.Conditional || mapping[tr.OutputReferenceID] != n.OutputReferenceID || !reflect.DeepEqual(inputs, n.InputReferenceIDs) {
				return fail("transition_correspondence", "Candidate changes canonical transfer relationships", Location{})
			}
		}
	}
	proof.Proven = true
	return proof
}

func (s *RewriteSession) rewriteFieldStateEqual(original, candidate FieldState, mapping map[string]string, effects map[string]RewriteIdentity, scope string, lineageIndex int, after bool) bool {
	expected := rewriteCopy(original)
	refs := map[string]Reference{}
	for _, ref := range s.result.References {
		refs[ref.ID] = ref
	}
	for i := range expected.Fields {
		field := &expected.Fields[i]
		if len(field.OriginReferenceIDs) > 0 {
			id := field.OriginReferenceIDs[0]
			if effect, ok := effects[id]; ok && effect.Name != nil && refs[id].NormalizedName == field.Name {
				field.Name = *effect.Name
			}
		}
		for j, id := range field.OriginReferenceIDs {
			field.OriginReferenceIDs[j] = mapping[id]
		}
	}
	for i, name := range expected.Removed {
		var latest *rewriteSite
		for _, site := range s.sites {
			p := site.public
			if p.Point.ScopeID != scope || p.Kind != "field" || p.Identity.Name == nil || *p.Identity.Name != name || (refs[p.ReferenceID].Role != "remove" && p.Role != "rename_input") || p.Point.LineageIndex > lineageIndex || (!after && p.Point.LineageIndex == lineageIndex) {
				continue
			}
			if latest == nil || p.Point.LineageIndex > latest.public.Point.LineageIndex || (p.Point.LineageIndex == latest.public.Point.LineageIndex && p.Point.Ordinal > latest.public.Point.Ordinal) {
				latest = site
			}
		}
		if latest != nil {
			if effect, ok := effects[latest.public.ReferenceID]; ok && effect.Name != nil {
				expected.Removed[i] = *effect.Name
			}
		}
	}
	sort.Slice(expected.Fields, func(i, j int) bool { return expected.Fields[i].Name < expected.Fields[j].Name })
	sort.Strings(expected.Removed)
	return reflect.DeepEqual(expected, candidate)
}

func (s *RewriteSession) rewriteReferenceLocation(ref Reference, rendered *RewriteRendering, source *sourceIndex) (Location, bool) {
	if ref.Kind == "data_model" {
		for _, model := range s.sites {
			if model.public.ReferenceID != ref.ID {
				continue
			}
			for _, dataset := range s.sites {
				if dataset.public.Role != "qualified_dataset" || dataset.public.OwnerLocation != model.public.OwnerLocation {
					continue
				}
				for _, edit := range rendered.edits {
					if edit.Location != dataset.public.Location {
						continue
					}
					for _, effect := range rendered.effects {
						if effect.ReferenceID == ref.ID && effect.After.Name != nil {
							start := rewriteTranslatedOffset(edit.Location.Start.Offset, rendered.edits, false)
							return source.byteLocation(start, start+len(*effect.After.Name))
						}
					}
				}
			}
		}
	}
	return rewriteTranslatedLocation(ref.Location, rendered.edits, source)
}

func (s *RewriteSession) rewriteUnknownOwnersEqual(candidate *RewriteSession, rendered *RewriteRendering, source *sourceIndex) bool {
	used := map[string]bool{}
	for _, site := range s.sites {
		p := site.public
		if p.ReferenceID != "" {
			continue
		}
		loc, ok := rewriteTranslatedLocation(p.Location, rendered.edits, source)
		if !ok {
			return false
		}
		owner, ok := rewriteTranslatedLocation(p.OwnerLocation, rendered.edits, source)
		if !ok {
			return false
		}
		found := false
		for _, next := range candidate.sites {
			q := next.public
			if q.ReferenceID == "" && !used[q.ID] && loc == q.Location && owner == q.OwnerLocation && p.Kind == q.Kind && p.Role == q.Role && p.Point == q.Point && rewriteIdentityEqual(p.Identity, q.Identity) {
				used[q.ID] = true
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	for _, site := range candidate.sites {
		if site.public.ReferenceID == "" && !used[site.public.ID] {
			return false
		}
	}
	return true
}
