package compatibility

import (
	"fmt"
	"slices"
	"sort"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/closure"
)

// Resolution closure keeps the submitted definitions immutable and uses only
// session-proved root coordinates to transfer original role authority.
func (p *Prepared) checkResolutionClosure(original, candidate analysis.ResolutionEvidence, roles []analysis.ResolutionRole, references []analysis.ResolutionReferencePair, assessment ResolutionAssessment) (*ResolutionReport, error) {
	// Complete input admission before interpreting any dependency content.
	admission := detach(assessment)
	admission.DependencyBindings = nil
	if _, err := p.checkResolutionDirect(original, candidate, roles, references, admission); err != nil {
		return nil, err
	}
	bundle, err := p.validateResolutionDependencies(original, assessment)
	if err != nil {
		return nil, err
	}
	bindings, err := translateResolutionBindings(original.Analysis, candidate.Analysis, references, assessment.DependencyBindings, bundle)
	if err != nil {
		return nil, err
	}
	for _, role := range roles {
		selected, err := p.resolveResolutionRole(original, role, role.CandidateInput, assessment)
		if err != nil {
			return nil, err
		}
		if selected.binding == nil || len(selected.objects) == 0 {
			continue
		}
		supplied := *selected.binding
		selectedBindings, err := sourceDependencyBindings(AssessmentRequest{Document: &candidate.Analysis.Document, QueryScope: assessment.QueryScope, InputBindings: []InputBinding{supplied}}, []assessmentQuery{{set: analysis.RequirementSet{Inputs: []analysis.QueryInput{resolutionRoleInput(role)}}}}, bindings, p)
		if err != nil {
			return nil, err
		}
		bindings = selectedBindings
	}
	sort.Slice(bindings, func(i, j int) bool { return stableKey(bindings[i]) < stableKey(bindings[j]) })
	evaluated, err := closure.Evaluate(closure.Request{SchemaVersion: 1, Document: candidate.Analysis.Document, Bundle: bundle, Bindings: bindings})
	if err != nil {
		return nil, closureRequestError(err)
	}
	queries := closureQueries(evaluated)
	effective := analysis.ResolutionEvidence{Analysis: *evaluated.EffectiveAnalysis, Coverage: candidate.Coverage, Placeholders: candidate.Placeholders}
	mapped := resolutionEffectiveRoles(roles, queries[0], candidate.Analysis.Document)
	effectiveReferences := resolutionEffectiveReferences(references, candidate.Analysis, *evaluated.EffectiveAnalysis, queries[0])
	report, err := p.checkResolutionRoles(original, effective, mapped, effectiveReferences, admission, evaluated)
	if err != nil {
		return nil, err
	}
	report.Requirements = detach(candidate.Analysis.Requirements)
	report.EffectiveRequirements = ptrResolutionSet(queries[0].set)
	report.Closure = evaluated
	report.DependencyBindings = detach(assessment.DependencyBindings)
	if report.DependencyBindings == nil {
		report.DependencyBindings = []closure.Binding{}
	}
	report.EffectiveDependencyBindings = detach(bindings)
	report.Provenance.QueryDigest = candidate.Analysis.Requirements.Query.QueryDigest
	report.Provenance.SourceID = candidate.Analysis.Requirements.Query.SourceID
	report.Correlation = detach(queries[0].set.Correlation)
	// Attach effective source coordinates independently to each role wrapper.
	for i := range report.RequirementOutcomes {
		out := &report.RequirementOutcomes[i]
		local := queries[0]
		local.set.Items = nil
		for _, item := range queries[0].set.Items {
			if item.ID == out.CandidateRequirementID {
				item.Occurrences = detach(out.AssessedOccurrences)
				local.set.Items = append(local.set.Items, item)
			}
		}
		child := &Report{RequirementOutcomes: []RequirementOutcome{out.Evidence}}
		attachOutcomeProvenance(child, local, evaluated)
		out.Evidence = child.RequirementOutcomes[0]
		for _, source := range out.Evidence.SourceIntervals {
			if source.Kind == "definition" && out.Evidence.DefinitionObjectID == "" {
				out.Evidence.DefinitionObjectID = source.ObjectID
			}
		}
	}
	for i := range report.Inputs {
		out := &report.Inputs[i]
		out.Evidence.Occurrences = nil
		for _, input := range queries[0].set.Inputs {
			if input.ID != out.CandidateInputID {
				continue
			}
			for _, occ := range input.Occurrences {
				if !slices.ContainsFunc(out.Occurrences, func(pair analysis.ResolutionOccurrencePair) bool { return pair.CandidateOccurrenceID == occ.ID }) {
					continue
				}
				sources, frames := mappedProvenance(queries[0], occ.Location.Start.Offset, occ.Location.End.Offset, evaluated)
				out.Evidence.Occurrences = append(out.Evidence.Occurrences, InputOccurrenceEvidence{Query: queries[0].set.Query, Occurrence: occ, SourceIntervals: sources, InvocationProvenance: frames})
			}
		}
	}
	// Expanded definition-origin input occurrences remain evidence even when
	// no original root role can authorize a schema selection for them.
	for _, input := range queries[0].set.Inputs {
		for _, occurrence := range input.Occurrences {
			if slices.ContainsFunc(mapped, func(role analysis.ResolutionRole) bool {
				return slices.ContainsFunc(role.Occurrences, func(pair analysis.ResolutionOccurrencePair) bool { return pair.CandidateOccurrenceID == occurrence.ID })
			}) {
				continue
			}
			local := detach(input)
			local.Occurrences = []analysis.InputOccurrence{detach(occurrence)}
			resolved, err := p.resolveUnselectedResolutionInput(local, assessment)
			if err != nil {
				return nil, err
			}
			out, coverage := p.assessInput(resolved, assessment.QueryScope)
			sources, frames := mappedProvenance(queries[0], occurrence.Location.Start.Offset, occurrence.Location.End.Offset, evaluated)
			evidence := InputOccurrenceEvidence{Query: queries[0].set.Query, Occurrence: occurrence, SourceIntervals: sources, InvocationProvenance: frames}
			if len(sources) == 1 && sources[0].Kind == "definition" {
				evidence.DefinitionObjectID = sources[0].ObjectID
			}
			out.Occurrences = []InputOccurrenceEvidence{evidence}
			report.Inputs = append(report.Inputs, ResolutionInputOutcome{CandidateInputID: input.ID, Occurrences: []analysis.ResolutionOccurrencePair{}, Evidence: out})
			report.Coverage = append(report.Coverage, ResolutionCoverage{Evidence: coverage})
			if out.Outcome != "satisfied" {
				required := false
				for _, item := range queries[0].set.Items {
					if item.Kind != "dataset" {
						continue
					}
					for _, read := range item.Occurrences {
						if read.Necessity == "required" && slices.Contains(read.InputOccurrenceIDs, occurrence.ID) {
							required = true
						}
					}
				}
				if out.Outcome == "missing" && required {
					report.Outcome = "unsatisfied"
				} else if report.Outcome != "unsatisfied" {
					report.Outcome = "incomplete"
				}
				for _, reason := range out.Reasons {
					report.Reasons = append(report.Reasons, ResolutionReason{Evidence: reason})
				}
			}
		}
	}
	// Definition contexts cannot borrow a root binding even when their canonical
	// input IDs, source IDs, or concrete names happen to coincide.
	for _, query := range queries[1:] {
		resolved := map[string]resolvedInput{}
		for _, input := range query.set.Inputs {
			value, err := p.resolveUnselectedResolutionInput(input, assessment)
			if err != nil {
				return nil, err
			}
			resolved[input.ID] = value
		}
		child := p.assess(AssessmentRequest{SchemaVersion: 1, Requirements: query.set, QueryScope: assessment.QueryScope, InputBindings: []InputBinding{}}, resolved, query, evaluated)
		attachOutcomeProvenance(child, query, evaluated)
		attachInputOccurrences(child, []assessmentQuery{query}, evaluated)
		p.selectedBodyLimits(child, resolved, evaluated)
		appendResolutionFragment(report, child, query.set)
	}
	// Expanded-away named dependencies retain their direct existence evidence.
	for _, item := range candidate.Analysis.Requirements.Items {
		if !expansionKind(item.Kind) || slices.ContainsFunc(queries[0].set.Items, func(effective analysis.RequirementItem) bool {
			return item.Kind == effective.Kind && item.Identity == effective.Identity
		}) {
			continue
		}
		query := assessmentQuery{set: candidate.Analysis.Requirements}
		out := p.selectedObjectOutcome(p.assessObject(item, query.set.Query, assessment.QueryScope), item, query, evaluated)
		child := &Report{Outcome: "satisfied", RequirementOutcomes: []RequirementOutcome{out}}
		attachOutcomeProvenance(child, query, evaluated)
		aggregateOutcome(child, out.Outcome, out.Applicability, out.Reasons)
		appendResolutionFragment(report, child, query.set)
	}
	reasons := []Reason{}
	for _, gap := range evaluated.Gaps {
		if gap.Code == "analysis_incomplete" || gap.Code == "collection_incomplete" && resolvedGap(gap, evaluated) {
			continue
		}
		reason := newReason("dependency_closure_incomplete", "Reachable dependency closure is incomplete: "+gap.Code+".", "dependency_closure")
		reason.ObjectID = gap.Source.ObjectID
		text := candidate.Analysis.Document.Text
		if gap.Source.Kind == "definition" {
			if object, found := p.env.Object(gap.Source.ObjectID); found && object.Document != nil {
				text = object.Document.Text
			}
		}
		location := resolutionSourceLocation(text, gap.Source.Start, gap.Source.End)
		reason.Location = &location
		reason.Locations = []analysis.Location{location}
		reasons = append(reasons, reason)
	}
	child := &Report{Outcome: "satisfied", Coverage: []Coverage{}}
	replaceClosureCoverage(child, p.reachableExpansion(evaluated), reasons)
	appendResolutionFragment(report, child, analysis.RequirementSet{})
	finalizeResolutionReport(report, assessment.QueryScope)
	return report, nil
}

func ptrResolutionSet(set analysis.RequirementSet) *analysis.RequirementSet {
	copy := detach(set)
	return &copy
}
func resolutionRoleInput(role analysis.ResolutionRole) analysis.QueryInput {
	input := detach(role.CandidateInput)
	input.Occurrences = slices.DeleteFunc(input.Occurrences, func(o analysis.InputOccurrence) bool {
		return !slices.ContainsFunc(role.Occurrences, func(pair analysis.ResolutionOccurrencePair) bool { return pair.CandidateOccurrenceID == o.ID })
	})
	return input
}

func translateResolutionBindings(original, candidate analysis.Result, pairs []analysis.ResolutionReferencePair, bindings []closure.Binding, bundle closure.DefinitionBundle) ([]closure.Binding, error) {
	out := []closure.Binding{}
	for i, b := range bindings {
		if b.DocumentDigest != queryDigest(original.Document.Text) {
			out = append(out, b)
			continue
		}
		// Digest bindings apply to every matching supplied document. Retain the
		// submitted coordinates for immutable definitions and independently
		// translate any matching root reference.
		sharedDefinition := slices.ContainsFunc(bundle.Objects, func(d closure.Definition) bool {
			return d.Document != nil && queryDigest(d.Document.Text) == b.DocumentDigest
		})
		if sharedDefinition && !slices.Contains(out, b) {
			out = append(out, b)
		}
		found := false
		for _, pair := range pairs {
			for _, before := range original.References {
				if before.ID != pair.OriginalID || before.Kind != b.Kind || !resolutionBindingReference(b, before, original.Document.Text) {
					continue
				}
				for _, after := range candidate.References {
					if after.ID != pair.CandidateID || after.Kind != before.Kind {
						continue
					}
					start := after.Location.Start.Offset - (before.Location.Start.Offset - b.Start)
					end := after.Location.End.Offset + (b.End - before.Location.End.Offset)
					if start < 0 || end > len(candidate.Document.Text) || end <= start || original.Document.Text[b.Start:b.End] != candidate.Document.Text[start:end] {
						return nil, requestErrorAt("binding_invalid", fmt.Sprintf("/dependency_bindings/%d", i), "bound reference bytes changed during rendering")
					}
					b.DocumentDigest = queryDigest(candidate.Document.Text)
					b.Start = start
					b.End = end
					found = true
				}
			}
		}
		if !found && !sharedDefinition {
			return nil, requestErrorAt("binding_invalid", fmt.Sprintf("/dependency_bindings/%d", i), "binding has no exact session-proved candidate reference")
		}
		if found && !slices.Contains(out, b) {
			out = append(out, b)
		}
	}
	return out, nil
}

// A complete uninterrupted root interval establishes correspondence. Overlaps,
// inserted definition bytes, and formal-argument placeholders grant no authority.
func resolutionRootInterval(query assessmentQuery, location analysis.Location, document analysis.QueryDocument) (closure.SourceInterval, bool) {
	start, end := location.Start.Offset, location.End.Offset
	matches := []closure.SourceInterval{}
	for _, segment := range query.provenance {
		if segment.EffectiveStart <= start && segment.EffectiveEnd >= end && segment.Source.Kind == "query" && segment.Source.ObjectID == "" && len(segment.Placeholders) == 0 && segment.Placeholder == nil && len(segment.InvocationChain) == 0 {
			source := segment.Source
			source.Start += start - segment.EffectiveStart
			source.End -= segment.EffectiveEnd - end
			if source.End-source.Start == end-start && source.Start >= 0 && source.End <= len(document.Text) {
				matches = append(matches, source)
			}
		} else if max(start, segment.EffectiveStart) < min(end, segment.EffectiveEnd) {
			return closure.SourceInterval{}, false
		}
	}
	if len(matches) != 1 {
		return closure.SourceInterval{}, false
	}
	return matches[0], true
}
func resolutionEffectiveRoles(roles []analysis.ResolutionRole, query assessmentQuery, document analysis.QueryDocument) []analysis.ResolutionRole {
	out := []analysis.ResolutionRole{}
	for _, role := range roles {
		for _, input := range query.set.Inputs {
			mapped := analysis.ResolutionRole{OriginalInput: detach(role.OriginalInput), CandidateInput: detach(input), Coverage: detach(role.Coverage), Occurrences: []analysis.ResolutionOccurrencePair{}, Requirements: []analysis.ResolutionRequirementPair{}}
			for _, occ := range input.Occurrences {
				source, ok := resolutionRootInterval(query, occ.Location, document)
				if !ok {
					continue
				}
				for _, before := range role.CandidateInput.Occurrences {
					if source.Start != before.Location.Start.Offset || source.End != before.Location.End.Offset {
						continue
					}
					for _, pair := range role.Occurrences {
						if pair.CandidateOccurrenceID == before.ID {
							pair.CandidateInputID = input.ID
							pair.CandidateOccurrenceID = occ.ID
							mapped.Occurrences = append(mapped.Occurrences, pair)
						}
					}
				}
			}
			if len(mapped.Occurrences) == 0 {
				continue
			}
			for _, item := range query.set.Items {
				if item.Ownership.State != "proved" || len(item.Ownership.CandidateInputIDs) != 1 || item.Ownership.CandidateInputIDs[0] != input.ID {
					continue
				}
				for _, occ := range item.Occurrences {
					source, ok := resolutionRootInterval(query, occ.Location, document)
					if !ok {
						continue
					}
					if len(occ.InputOccurrenceIDs) == 0 || slices.ContainsFunc(occ.InputOccurrenceIDs, func(id string) bool {
						return !slices.ContainsFunc(mapped.Occurrences, func(pair analysis.ResolutionOccurrencePair) bool { return pair.CandidateOccurrenceID == id })
					}) {
						continue
					}
					if !slices.ContainsFunc(occ.InputOccurrenceIDs, func(id string) bool {
						return slices.ContainsFunc(mapped.Occurrences, func(pair analysis.ResolutionOccurrencePair) bool { return pair.CandidateOccurrenceID == id })
					}) {
						continue
					}
					for _, pair := range role.Requirements {
						if source.Start == pair.CandidateOccurrence.Location.Start.Offset && source.End == pair.CandidateOccurrence.Location.End.Offset {
							pair.CandidateRequirementID = item.ID
							pair.CandidateOccurrence = detach(occ)
							mapped.Requirements = append(mapped.Requirements, pair)
						}
					}
				}
			}
			out = append(out, mapped)
		}
	}
	return out
}
func appendResolutionFragment(report *ResolutionReport, child *Report, set analysis.RequirementSet) {
	if child.Outcome == "unsatisfied" {
		report.Outcome = "unsatisfied"
	} else if child.Outcome != "satisfied" && report.Outcome != "unsatisfied" {
		report.Outcome = "incomplete"
	}
	for _, out := range child.Inputs {
		report.Inputs = append(report.Inputs, ResolutionInputOutcome{CandidateInputID: out.InputID, Occurrences: []analysis.ResolutionOccurrencePair{}, Evidence: out})
	}
	for _, out := range child.RequirementOutcomes {
		occurrences := []analysis.RequirementOccurrence{}
		for _, item := range set.Items {
			if item.ID == out.RequirementID {
				occurrences = detach(item.Occurrences)
			}
		}
		report.RequirementOutcomes = append(report.RequirementOutcomes, ResolutionRequirementOutcome{CandidateInputID: out.InputID, CandidateRequirementID: out.RequirementID, AssessedOccurrences: occurrences, Evidence: out})
	}
	for _, reason := range child.Reasons {
		report.Reasons = append(report.Reasons, ResolutionReason{Evidence: reason})
	}
	for _, coverage := range child.Coverage {
		report.Coverage = append(report.Coverage, ResolutionCoverage{Evidence: coverage})
	}
}

func resolutionEffectiveReferences(pairs []analysis.ResolutionReferencePair, candidate, effective analysis.Result, query assessmentQuery) []analysis.ResolutionReferencePair {
	out := []analysis.ResolutionReferencePair{}
	for _, pair := range pairs {
		for _, before := range candidate.References {
			if before.ID != pair.CandidateID {
				continue
			}
			for _, after := range effective.References {
				source, ok := resolutionRootInterval(query, after.Location, candidate.Document)
				if ok && source.Start == before.Location.Start.Offset && source.End == before.Location.End.Offset && before.Kind == after.Kind {
					out = append(out, analysis.ResolutionReferencePair{OriginalID: pair.OriginalID, CandidateID: after.ID})
				}
			}
		}
	}
	return out
}

// Macro closure bindings cover the invocation delimiters and arguments; the
// canonical analysis reference covers its name. Both are validated by closure.
func resolutionBindingReference(binding closure.Binding, ref analysis.Reference, text string) bool {
	if ref.Location.Start.Offset == binding.Start && ref.Location.End.Offset == binding.End {
		return true
	}
	return binding.Kind == "macro" && binding.Start >= 0 && binding.End <= len(text) && binding.End > binding.Start+1 && text[binding.Start] == '`' && text[binding.End-1] == '`' && ref.Location.Start.Offset == binding.Start+1 && ref.Location.End.Offset <= binding.End-1
}

// Definition-only placeholders have no resolution selection contract. Their
// absence of source evidence is an assessment limitation, not malformed input.
func (p *Prepared) resolveUnselectedResolutionInput(input analysis.QueryInput, assessment ResolutionAssessment) (resolvedInput, error) {
	if input.Kind == "named_placeholder" {
		return resolvedInput{input: input}, nil
	}
	values, err := p.resolveInputs([]analysis.QueryInput{input}, assessment.QueryScope, []InputBinding{})
	return values[input.ID], err
}

func resolutionSourceLocation(text string, start, end int) analysis.Location {
	position := func(offset int) analysis.Position {
		offset = max(0, min(offset, len(text)))
		p := analysis.Position{Offset: offset, Line: 1, Column: 1}
		for _, r := range text[:offset] {
			if r == '\n' {
				p.Line++
				p.Column = 1
			} else {
				p.Column++
			}
		}
		return p
	}
	return analysis.Location{Start: position(start), End: position(end)}
}

// The closure owner admits submitted bindings against original root bytes and
// every immutable supplied definition before any rendered variant is assessed.
func (p *Prepared) validateResolutionDependencies(original analysis.ResolutionEvidence, assessment ResolutionAssessment) (closure.DefinitionBundle, error) {
	bundle, err := p.env.DefinitionBundle(assessment.QueryScope)
	if err != nil {
		return bundle, err
	}
	if len(assessment.DependencyBindings) > 0 {
		if _, err := closure.Evaluate(closure.Request{SchemaVersion: 1, Document: original.Analysis.Document, Bundle: bundle, Bindings: assessment.DependencyBindings}); err != nil {
			return bundle, closureRequestError(err)
		}
	}
	return bundle, nil
}
