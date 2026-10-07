package compatibility

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"slices"
	"sort"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/closure"
	"github.com/delgado-jacob/spl-toolkit/pkg/environment"
)

// CheckResolution consumes session authority, never serialized proof evidence.
// Closure integration establishes effective-query correspondence before
// reusing the direct role evaluator.
func (p *Prepared) CheckResolution(proof *analysis.ResolutionProof, assessment ResolutionAssessment) (*ResolutionReport, error) {
	original, candidate, roles, authorized := proof.AssessmentEvidence()
	if !authorized {
		return nil, requestErrorAt("resolution_proof_invalid", "/proof", "session-bound substitution authority is required")
	}
	return p.checkResolutionClosure(original, candidate, roles, proof.Evidence().References, assessment)
}

func (p *Prepared) checkResolutionDirect(original, candidate analysis.ResolutionEvidence, roles []analysis.ResolutionRole, references []analysis.ResolutionReferencePair, assessment ResolutionAssessment) (*ResolutionReport, error) {
	return p.checkResolutionRoles(original, candidate, roles, references, assessment, nil)
}

func (p *Prepared) checkResolutionRoles(original, candidate analysis.ResolutionEvidence, roles []analysis.ResolutionRole, references []analysis.ResolutionReferencePair, assessment ResolutionAssessment, evaluated *closure.Report) (*ResolutionReport, error) {
	// Admission sees all flattened selections, including alternatives belonging to
	// other variants. Only the exact role/value for this proof is used below.
	choices := []analysis.ResolutionChoice{}
	for _, group := range original.Placeholders {
		for _, role := range roles {
			if slices.Contains(group.OriginalInputIDs, role.OriginalInput.ID) {
				choices = append(choices, analysis.ResolutionChoice{Placeholder: group.Placeholder, Kind: group.Kind, Value: role.CandidateInput.Name})
			}
		}
		for _, b := range assessment.InputBindings {
			if slices.Contains(group.OriginalInputIDs, b.OriginalInputID) && b.ResolvedValue != nil {
				choices = append(choices, analysis.ResolutionChoice{Placeholder: group.Placeholder, Kind: group.Kind, Value: *b.ResolvedValue})
			}
		}
	}
	if err := p.ValidateResolutionBindings(original, choices, assessment); err != nil {
		return nil, err
	}
	normalized, err := normalizeAssessment(AssessmentRequest{SchemaVersion: 1, Requirements: candidate.Analysis.Requirements, QueryScope: assessment.QueryScope, InputBindings: []InputBinding{}, DependencyBindings: assessment.DependencyBindings})
	if err != nil {
		return nil, err
	}
	assessment = detach(assessment)
	assessment.QueryScope = normalized.QueryScope
	if len(assessment.DependencyBindings) > 0 {
		return nil, requestErrorAt("request_invalid", "/dependency_bindings", "dependency bindings require resolution closure assessment")
	}
	set := candidate.Analysis.Requirements
	artifact := p.ArtifactIdentity()
	report := &ResolutionReport{SchemaVersion: 1, Outcome: "satisfied", Requirements: detach(set), Correlation: detach(set.Correlation), InputBindings: detach(assessment.InputBindings), DependencyBindings: []closure.Binding{}, Inputs: []ResolutionInputOutcome{}, RequirementOutcomes: []ResolutionRequirementOutcome{}, Coverage: []ResolutionCoverage{}, Reasons: []ResolutionReason{}, Diagnostics: p.env.Report().Diagnostics, Provenance: Provenance{QueryDigest: set.Query.QueryDigest, SourceID: set.Query.SourceID, CapabilityRevision: set.CapabilityRevision, AnalysisContractVersion: 1, RequirementSetVersion: set.SchemaVersion, EnvironmentDigest: artifact.EnvironmentDigest, SchemaBundleDigest: artifact.SchemaBundleDigest}}
	missing, unknown := false, false
	record := func(role string, outcome, applicability string, reasons []Reason) {
		if outcome == "satisfied" || applicability == "inapplicable" {
			return
		}
		if outcome == "missing" && applicability == "applicable" {
			missing = true
		} else {
			unknown = true
		}
		for _, reason := range reasons {
			report.Reasons = append(report.Reasons, ResolutionReason{OriginalInputID: role, Evidence: reason})
		}
	}
	byID := map[string]analysis.RequirementItem{}
	for _, item := range set.Items {
		byID[item.ID] = item
	}
	paired := map[string]bool{}
	for _, role := range roles {
		input := detach(role.CandidateInput)
		input.Occurrences = slices.DeleteFunc(input.Occurrences, func(o analysis.InputOccurrence) bool {
			return !slices.ContainsFunc(role.Occurrences, func(pair analysis.ResolutionOccurrencePair) bool { return pair.CandidateOccurrenceID == o.ID })
		})
		r, err := p.resolveResolutionRole(original, role, input, assessment)
		if err != nil {
			return nil, err
		}
		inputOut, inputCoverage := p.assessInput(r, assessment.QueryScope)
		linked := slices.ContainsFunc(role.Requirements, func(pair analysis.ResolutionRequirementPair) bool {
			return byID[pair.CandidateRequirementID].Kind == "dataset"
		})
		if linked {
			inputOut.Outcome = "satisfied"
			inputOut.Reasons = []Reason{}
		} else {
			record(role.OriginalInput.ID, inputOut.Outcome, "applicable", inputOut.Reasons)
		}
		report.Coverage = append(report.Coverage, ResolutionCoverage{OriginalInputID: role.OriginalInput.ID, Evidence: inputCoverage})
		// Each one-item projection preserves occurrence necessity before canonical
		// grouping. Its artificial subset is never published as Requirements.
		for _, pair := range role.Requirements {
			item := detach(byID[pair.CandidateRequirementID])
			item.InputID = input.ID
			item.Necessity = pair.OriginalOccurrence.Necessity
			item.Occurrences = []analysis.RequirementOccurrence{detach(pair.CandidateOccurrence)}
			item.Ownership.State = "proved"
			item.Ownership.CandidateInputIDs = []string{input.ID}
			projected := resolutionLocalSet(set, input, item)
			child := p.assess(AssessmentRequest{SchemaVersion: 1, Requirements: projected, QueryScope: assessment.QueryScope, InputBindings: []InputBinding{}}, map[string]resolvedInput{input.ID: r}, assessmentQuery{set: projected}, nil)
			for _, out := range child.RequirementOutcomes {
				if out.RequirementID != item.ID {
					continue
				}
				report.RequirementOutcomes = append(report.RequirementOutcomes, ResolutionRequirementOutcome{OriginalInputID: role.OriginalInput.ID, OriginalRequirementID: pair.OriginalRequirementID, CandidateInputID: input.ID, CandidateRequirementID: item.ID, AssessedOccurrences: detach(item.Occurrences), Evidence: out})
				record(role.OriginalInput.ID, out.Outcome, out.Applicability, out.Reasons)
				inputOut.RequirementIDs = append(inputOut.RequirementIDs, item.ID)
				updateResolutionInput(&inputOut, out)
			}
			for _, coverage := range child.Coverage {
				if coverage.Dimension == "field_schema" || coverage.Dimension == "environment_collection" && coverage.CollectionKind != "" {
					report.Coverage = append(report.Coverage, ResolutionCoverage{OriginalInputID: role.OriginalInput.ID, Evidence: coverage})
				}
			}
			paired[resolutionOccurrenceKey(item.ID, pair.CandidateOccurrence)] = true
		}
		limits := &Report{Outcome: "satisfied", Inputs: []InputOutcome{inputOut}, Coverage: []Coverage{{Dimension: "dependency_closure", State: "not_applicable", Reasons: []Reason{}}}}
		p.selectedBodyLimits(limits, map[string]resolvedInput{input.ID: r}, evaluated)
		if limits.Outcome != "satisfied" {
			unknown = true
			for _, reason := range limits.Reasons {
				report.Reasons = append(report.Reasons, ResolutionReason{OriginalInputID: role.OriginalInput.ID, Evidence: reason})
			}
			for _, coverage := range limits.Coverage {
				report.Coverage = append(report.Coverage, ResolutionCoverage{OriginalInputID: role.OriginalInput.ID, Evidence: coverage})
			}
		}
		attachInputOccurrencesToResolution(&inputOut, input, set.Query)
		report.Inputs = append(report.Inputs, ResolutionInputOutcome{OriginalInputID: role.OriginalInput.ID, CandidateInputID: input.ID, Occurrences: detach(role.Occurrences), Evidence: inputOut})
		coverage := resolutionRoleCoverage(role, references, candidate)
		report.Coverage = append(report.Coverage, ResolutionCoverage{OriginalInputID: role.OriginalInput.ID, Evidence: coverage})
		if coverage.State == "partial" {
			unknown = true
			for _, reason := range coverage.Reasons {
				report.Reasons = append(report.Reasons, ResolutionReason{OriginalInputID: role.OriginalInput.ID, Evidence: reason})
			}
		}
	}
	// Global/non-source obligations retain their ordinary owners and are assessed
	// once. An unpaired source read cannot acquire a convenient role after merging.
	global := detach(set)
	global.Inputs = []analysis.QueryInput{}
	global.Items = []analysis.RequirementItem{}
	// Direct roles already establish assessable obligations. Removing them from
	// this local remainder must not trigger the ordinary empty-query stop gate;
	// retained semantic/coverage gaps still determine incompleteness.
	if set.QueryStatus == analysis.Incomplete && len(set.Inputs)+len(set.Items) > 0 {
		global.QueryStatus = analysis.Valid
	}
	for _, item := range set.Items {
		item = detach(item)
		item.Occurrences = slices.DeleteFunc(item.Occurrences, func(o analysis.RequirementOccurrence) bool { return paired[resolutionOccurrenceKey(item.ID, o)] })
		if len(item.Occurrences) == 0 {
			continue
		}
		if item.Kind == "field" {
			item.InputID = ""
			item.Ownership.State = "unproved"
		}
		if item.InputID != "" {
			item.InputID = ""
			item.Resolution = "indeterminate"
		}
		global.Items = append(global.Items, item)
	}
	global.Gaps = slices.DeleteFunc(global.Gaps, func(gap analysis.RequirementGap) bool {
		return gap.Code == analysis.CodeRequirementIndeterminate && resolutionGapSatisfied(gap, report.RequirementOutcomes)
	})
	globalQuery := assessmentQuery{set: global}
	if evaluated != nil {
		globalQuery.provenance = evaluated.Provenance
	}
	child := p.assess(AssessmentRequest{SchemaVersion: 1, Requirements: global, QueryScope: assessment.QueryScope, InputBindings: []InputBinding{}}, map[string]resolvedInput{}, globalQuery, evaluated)
	p.selectedBodyLimits(child, map[string]resolvedInput{}, evaluated)
	// A conservative local ownership projection changes assessment, never the
	// canonical reporting coordinates of the obligation it assesses.
	canonicalReason := func(reason Reason) Reason {
		if item, ok := byID[reason.RequirementID]; ok {
			reason.InputID = item.InputID
		}
		return reason
	}
	for i := range child.RequirementOutcomes {
		out := &child.RequirementOutcomes[i]
		if item, ok := byID[out.RequirementID]; ok {
			out.InputID = item.InputID
		}
		for j := range out.Reasons {
			out.Reasons[j] = canonicalReason(out.Reasons[j])
		}
	}
	for i := range child.Reasons {
		child.Reasons[i] = canonicalReason(child.Reasons[i])
	}
	for i := range child.Coverage {
		for j := range child.Coverage[i].Reasons {
			child.Coverage[i].Reasons[j] = canonicalReason(child.Coverage[i].Reasons[j])
		}
		if len(child.Coverage[i].Reasons) > 0 && child.Coverage[i].InputID == "" && child.Coverage[i].Dimension == "field_schema" {
			child.Coverage[i].InputID = child.Coverage[i].Reasons[0].InputID
		}
	}
	for _, out := range child.RequirementOutcomes {
		occurrences := []analysis.RequirementOccurrence{}
		for _, item := range global.Items {
			if item.ID == out.RequirementID {
				occurrences = detach(item.Occurrences)
			}
		}
		report.RequirementOutcomes = append(report.RequirementOutcomes, ResolutionRequirementOutcome{CandidateInputID: out.InputID, CandidateRequirementID: out.RequirementID, AssessedOccurrences: occurrences, Evidence: out})
		record("", out.Outcome, out.Applicability, out.Reasons)
	}
	for _, coverage := range child.Coverage {
		report.Coverage = append(report.Coverage, ResolutionCoverage{Evidence: coverage})
	}
	for _, reason := range child.Reasons {
		if !slices.ContainsFunc(report.Reasons, func(prior ResolutionReason) bool {
			return prior.OriginalInputID == "" && stableKey(prior.Evidence) == stableKey(reason)
		}) {
			report.Reasons = append(report.Reasons, ResolutionReason{Evidence: reason})
		}
	}
	if child.Outcome == "unsatisfied" {
		missing = true
	} else if child.Outcome != "satisfied" {
		unknown = true
	}
	known := map[string]bool{}
	for _, role := range roles {
		known[role.OriginalInput.ID] = true
	}
	for _, binding := range assessment.InputBindings {
		if !known[binding.OriginalInputID] {
			unknown = true
			reason := newReason("resolution_role_incomplete", "The binding's original role is not established by partial query discovery.", "target_discovery")
			report.Reasons = append(report.Reasons, ResolutionReason{OriginalInputID: binding.OriginalInputID, Evidence: reason})
			report.Coverage = append(report.Coverage, ResolutionCoverage{OriginalInputID: binding.OriginalInputID, Evidence: Coverage{Dimension: "target_discovery", State: "partial", Reasons: []Reason{reason}}})
		}
	}
	if missing {
		report.Outcome = "unsatisfied"
	} else if unknown {
		report.Outcome = "incomplete"
	}
	finalizeResolutionReport(report, assessment.QueryScope)
	return report, nil
}

func (p *Prepared) resolveResolutionRole(original analysis.ResolutionEvidence, role analysis.ResolutionRole, input analysis.QueryInput, a ResolutionAssessment) (resolvedInput, error) {
	substituted := slices.ContainsFunc(original.Placeholders, func(group analysis.ResolutionPlaceholder) bool {
		return slices.Contains(group.OriginalInputIDs, role.OriginalInput.ID)
	})
	bindings := []InputBinding{}
	for i, b := range a.InputBindings {
		if b.OriginalInputID != role.OriginalInput.ID {
			continue
		}
		if substituted && (b.ResolvedValue == nil || *b.ResolvedValue != input.Name) || !substituted && b.ResolvedValue != nil {
			continue
		}
		intended, mapped := explicitIdentity(input)
		if input.Kind == "explicit_dataset" && (!mapped || !consistentExplicit(input, intended, b.Expected) || !identityInScope(a.QueryScope, b.Expected)) {
			return resolvedInput{}, requestErrorAt("binding_invalid", fmt.Sprintf("/input_bindings/%d/expected", i), "binding disagrees with the proved candidate identity and scope")
		}
		bindings = append(bindings, InputBinding{InputID: input.ID, ObjectID: b.ObjectID, Expected: b.Expected, SchemaID: b.SchemaID})
	}
	resolved, err := p.resolveInputs([]analysis.QueryInput{input}, a.QueryScope, bindings)
	if err != nil {
		return resolvedInput{}, err
	}
	r := resolved[input.ID]
	if substituted && len(bindings) == 0 {
		r.objects = []environment.Object{}
		r.explicitSourceEvidence = false
	}
	return r, nil
}
func resolutionLocalSet(set analysis.RequirementSet, input analysis.QueryInput, item analysis.RequirementItem) analysis.RequirementSet {
	out := detach(set)
	out.Inputs = []analysis.QueryInput{input}
	out.Items = []analysis.RequirementItem{item}
	out.Gaps = []analysis.RequirementGap{}
	out.Diagnostics = nil
	out.InputCoverage = analysis.InputCoverage{State: "complete", Reasons: []analysis.InputReason{}}
	out.FieldAttributionCoverage = analysis.InputCoverage{State: "complete", Reasons: []analysis.InputReason{}}
	return out
}
func resolutionOccurrenceKey(id string, o analysis.RequirementOccurrence) string {
	return id + ":" + stableKey(o)
}
func updateResolutionInput(input *InputOutcome, out RequirementOutcome) {
	if out.Outcome == "satisfied" {
		return
	}
	input.Reasons = append(input.Reasons, out.Reasons...)
	if out.Outcome == "missing" && out.Applicability == "applicable" {
		input.Outcome = "missing"
	} else if input.Outcome != "missing" {
		if out.Outcome == "ambiguous" {
			input.Outcome = "ambiguous"
		} else if input.Outcome != "ambiguous" {
			input.Outcome = "indeterminate"
		}
	}
}
func attachInputOccurrencesToResolution(out *InputOutcome, input analysis.QueryInput, query analysis.RequirementQueryIdentity) {
	out.Occurrences = []InputOccurrenceEvidence{}
	for _, o := range input.Occurrences {
		out.Occurrences = append(out.Occurrences, InputOccurrenceEvidence{Query: query, Occurrence: o, SourceIntervals: []closure.SourceInterval{{Kind: "query", SourceID: query.SourceID, Start: o.Location.Start.Offset, End: o.Location.End.Offset}}, InvocationProvenance: []closure.InvocationFrame{}})
	}
}
func resolutionRoleCoverage(role analysis.ResolutionRole, references []analysis.ResolutionReferencePair, candidate analysis.ResolutionEvidence) Coverage {
	c := Coverage{Dimension: "field_attribution", State: role.Coverage.State, InputID: role.CandidateInput.ID, Reasons: []Reason{}}
	for _, r := range role.Coverage.Reasons {
		reason := newReason(r.Code, r.Message, c.Dimension)
		reason.InputID = role.CandidateInput.ID
		for _, ref := range r.ReferenceIDs {
			for _, pair := range references {
				if pair.OriginalID == ref {
					reason.ReferenceIDs = append(reason.ReferenceIDs, pair.CandidateID)
					for _, item := range candidate.Analysis.Requirements.Items {
						for _, o := range item.Occurrences {
							if o.ReferenceID == pair.CandidateID {
								reason.Locations = append(reason.Locations, o.Location)
							}
						}
					}
				}
			}
		}
		if len(reason.Locations) > 0 {
			location := reason.Locations[0]
			reason.Location = &location
		}
		c.Reasons = append(c.Reasons, reason)
	}
	return c
}
func resolutionGapSatisfied(gap analysis.RequirementGap, outcomes []ResolutionRequirementOutcome) bool {
	if len(gap.ReferenceIDs) == 0 {
		return false
	}
	for _, ref := range gap.ReferenceIDs {
		found := false
		for _, out := range outcomes {
			for _, o := range out.AssessedOccurrences {
				if o.ReferenceID == ref {
					found = true
					if out.Evidence.Outcome != "satisfied" {
						return false
					}
				}
			}
		}
		if !found {
			return false
		}
	}
	return true
}
func finalizeResolutionReport(report *ResolutionReport, scope environment.CaptureScope) {
	// Normalize children using their owner, then retain all role wrappers; ordinary
	// report merging and its candidate-ID binding digest are not applicable here.
	for i := range report.Inputs {
		r := &Report{Inputs: []InputOutcome{report.Inputs[i].Evidence}}
		finalizeReport(r)
		report.Inputs[i].Evidence = r.Inputs[0]
	}
	for i := range report.RequirementOutcomes {
		r := &Report{RequirementOutcomes: []RequirementOutcome{report.RequirementOutcomes[i].Evidence}}
		finalizeReport(r)
		report.RequirementOutcomes[i].Evidence = r.RequirementOutcomes[0]
		sort.Slice(report.RequirementOutcomes[i].AssessedOccurrences, func(a, b int) bool {
			return stableKey(report.RequirementOutcomes[i].AssessedOccurrences[a]) < stableKey(report.RequirementOutcomes[i].AssessedOccurrences[b])
		})
	}
	sort.Slice(report.Inputs, func(i, j int) bool {
		a, b := report.Inputs[i], report.Inputs[j]
		if a.OriginalInputID != b.OriginalInputID {
			return a.OriginalInputID < b.OriginalInputID
		}
		return stableKey(a) < stableKey(b)
	})
	sort.Slice(report.RequirementOutcomes, func(i, j int) bool {
		a, b := report.RequirementOutcomes[i], report.RequirementOutcomes[j]
		if a.OriginalInputID != b.OriginalInputID {
			return a.OriginalInputID < b.OriginalInputID
		}
		if a.CandidateRequirementID != b.CandidateRequirementID {
			return a.CandidateRequirementID < b.CandidateRequirementID
		}
		return stableKey(a.AssessedOccurrences) < stableKey(b.AssessedOccurrences)
	})
	for i := range report.Reasons {
		r := []Reason{report.Reasons[i].Evidence}
		sortReasons(r)
		report.Reasons[i].Evidence = r[0]
	}
	for i := range report.Coverage {
		sortReasons(report.Coverage[i].Evidence.Reasons)
	}
	sort.Slice(report.Reasons, func(i, j int) bool {
		a, b := report.Reasons[i], report.Reasons[j]
		if a.OriginalInputID != b.OriginalInputID {
			return a.OriginalInputID < b.OriginalInputID
		}
		return stableKey(a) < stableKey(b)
	})
	sort.Slice(report.Coverage, func(i, j int) bool {
		a, b := report.Coverage[i], report.Coverage[j]
		if a.OriginalInputID != b.OriginalInputID {
			return a.OriginalInputID < b.OriginalInputID
		}
		return stableKey(a) < stableKey(b)
	})
	sort.Slice(report.InputBindings, func(i, j int) bool { return stableKey(report.InputBindings[i]) < stableKey(report.InputBindings[j]) })
	sort.Slice(report.DependencyBindings, func(i, j int) bool {
		return stableKey(report.DependencyBindings[i]) < stableKey(report.DependencyBindings[j])
	})
	sort.Slice(report.Diagnostics, func(i, j int) bool { return stableKey(report.Diagnostics[i]) < stableKey(report.Diagnostics[j]) })
	identity := struct {
		Provenance         Provenance               `json:"provenance"`
		QueryScope         environment.CaptureScope `json:"query_scope"`
		InputBindings      []ResolutionBinding      `json:"input_bindings"`
		DependencyBindings []closure.Binding        `json:"dependency_bindings"`
	}{report.Provenance, scope, report.InputBindings, report.DependencyBindings}
	identity.Provenance.AssessmentIdentityDigest = ""
	sum := sha256.Sum256([]byte(stableKey(identity)))
	report.Provenance.AssessmentIdentityDigest = "sha256:" + hex.EncodeToString(sum[:])
}
