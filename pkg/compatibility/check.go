package compatibility

import (
	"slices"
	"strings"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/closure"
	"github.com/delgado-jacob/spl-toolkit/pkg/environment"
)

func newReason(code, message, dimension string) Reason {
	return Reason{Code: code, Message: message, Dimension: dimension, Locations: []analysis.Location{}, ReferenceIDs: []string{}, CandidateInputIDs: []string{}}
}
func baseOutcome(item analysis.RequirementItem, query analysis.RequirementQueryIdentity) RequirementOutcome {
	applicability := "indeterminate"
	if item.Necessity == "required" {
		applicability = "applicable"
	}
	return RequirementOutcome{Query: query, RequirementID: item.ID, InputID: item.InputID, Applicability: applicability, Outcome: "indeterminate", Reasons: []Reason{}, Capabilities: []environment.Capability{}, Objects: []ObjectEvidence{}, Schemas: []SchemaEvidence{}, SourceIntervals: []closure.SourceInterval{}, InvocationProvenance: []closure.InvocationFrame{}}
}
func outcomeReason(out RequirementOutcome, item analysis.RequirementItem, code, message, dimension string) RequirementOutcome {
	reason := newReason(code, message, dimension)
	reason.RequirementID = item.ID
	reason.InputID = item.InputID
	reason.CandidateInputIDs = append(reason.CandidateInputIDs, item.Ownership.CandidateInputIDs...)
	for _, o := range item.Occurrences {
		reason.Locations = append(reason.Locations, o.Location)
		reason.ReferenceIDs = append(reason.ReferenceIDs, o.ReferenceID)
	}
	if len(reason.Locations) > 0 {
		location := reason.Locations[0]
		reason.Location = &location
	}
	if len(out.Objects) == 1 {
		reason.ObjectID = out.Objects[0].ObjectID
	}
	if len(out.Schemas) == 1 {
		reason.SchemaID = out.Schemas[0].SchemaID
	}
	out.Reasons = append(out.Reasons, reason)
	return out
}

// The opaque capability ID asserts one supplied language/profile fact at the
// canonical query version. It never asserts toolkit semantic support or derives
// availability from product-version strings.
func (p *Prepared) assessCapability(query analysis.RequirementQueryIdentity) RequirementOutcome {
	id := "language:" + query.Language + ":profile:" + query.Profile
	out := baseOutcome(analysis.RequirementItem{ID: "capability:" + id, Necessity: "required"}, query)
	code, message := "environment_capability_not_supplied", "Supply the query language/profile capability at its canonical version."
	for _, capability := range p.snapshot.Capabilities {
		if capability.ID != id {
			continue
		}
		out.Capabilities = append(out.Capabilities, detach(capability))
		if capability.Version != query.Version {
			code = "environment_capability_version_unproven"
			message = "The supplied capability version differs from the canonical query version; supply matching version evidence."
			break
		}
		switch capability.State {
		case "available":
			out.Outcome = "satisfied"
			return out
		case "unavailable":
			out.Outcome = "missing"
			code = "environment_capability_unavailable"
			message = "The supplied matching language/profile capability is unavailable."
		default:
			code = "environment_capability_unknown"
			message = "The supplied matching capability has unknown availability; supply conclusive evidence."
		}
		break
	}
	reason := newReason(code, message, "environment_capability")
	reason.RequirementID = out.RequirementID
	out.Reasons = append(out.Reasons, reason)
	return out
}
func correlationOnly(code string) bool {
	return strings.HasPrefix(code, "correlation_") || strings.HasPrefix(code, "SPL_CORRELATION_")
}
func (p *Prepared) assess(request AssessmentRequest, resolved map[string]resolvedInput) *Report {
	set := request.Requirements
	envReport := p.env.Report()
	report := &Report{SchemaVersion: 1, Outcome: "satisfied", QueryScope: detach(request.QueryScope), InputBindings: detach(request.InputBindings), Observation: detach(p.snapshot.Observation), Requirements: set, Correlation: detach(set.Correlation), Inputs: []InputOutcome{}, RequirementOutcomes: []RequirementOutcome{}, Coverage: []Coverage{}, Diagnostics: envReport.Diagnostics, Reasons: []Reason{}, Provenance: Provenance{QueryDigest: set.Query.QueryDigest, SourceID: set.Query.SourceID, CapabilityRevision: set.CapabilityRevision, AnalysisContractVersion: 1, RequirementSetVersion: set.SchemaVersion, EnvironmentDigest: envReport.SnapshotDigest, SchemaBundleDigest: envReport.SchemaBundleDigest}}
	missing, unknown := false, false
	record := func(outcome, applicability string, reasons []Reason) {
		if applicability == "inapplicable" {
			return
		}
		switch outcome {
		case "satisfied":
			return
		case "missing":
			if applicability == "applicable" {
				missing = true
				report.Reasons = append(report.Reasons, reasons...)
				return
			}
			unknown = true
		case "ambiguous", "indeterminate":
			unknown = true
		}
		report.Reasons = append(report.Reasons, reasons...)
	}
	cap := p.assessCapability(set.Query)
	report.RequirementOutcomes = append(report.RequirementOutcomes, cap)
	record(cap.Outcome, cap.Applicability, cap.Reasons)
	capabilityCoverage := Coverage{Dimension: "environment_capability", State: "complete", Reasons: append([]Reason{}, cap.Reasons...)}
	if cap.Outcome == "indeterminate" {
		capabilityCoverage.State = "partial"
		if len(cap.Capabilities) == 0 {
			capabilityCoverage.State = "unavailable"
		}
	}
	report.Coverage = append(report.Coverage, capabilityCoverage)
	inputObjects := map[string]InputOutcome{}
	for _, input := range set.Inputs {
		out, coverage := p.assessInput(resolved[input.ID], request.QueryScope)
		inputObjects[input.ID] = out
		report.Coverage = append(report.Coverage, coverage)
		// Dataset requirement items carry necessity. Do not independently turn
		// a conditional source obligation into an unconditional blocker.
		linked := false
		for _, item := range set.Items {
			if item.Kind == "dataset" && item.InputID == input.ID {
				linked = true
				break
			}
		}
		if linked {
			// Seed compatibility aggregation independently of object existence.
			// The linked requirement outcomes retain the object fact and decide
			// whether absence is an applicable blocker or conditional unknown.
			out.Outcome = "satisfied"
		} else {
			record(out.Outcome, "applicable", out.Reasons)
		}
		report.Inputs = append(report.Inputs, out)
	}
	for _, item := range set.Items {
		var out RequirementOutcome
		if item.Kind == "field" {
			out = p.assessField(item, set.Query, resolved)
		} else if item.Kind == "dataset" && item.InputID != "" {
			input := inputObjects[item.InputID]
			out = baseOutcome(item, set.Query)
			out.Outcome = input.Outcome
			out.Objects = input.Objects
			for _, reason := range input.Reasons {
				out = outcomeReason(out, item, reason.Code, reason.Message, reason.Dimension)
			}
		} else {
			out = p.assessObject(item, set.Query, request.QueryScope)
		}
		if out.Outcome == "missing" && out.Applicability == "indeterminate" {
			out = outcomeReason(out, item, "conditional_applicability_unproven", "The missing fact is retained, but query evidence has not proved that its condition applies. Schema evidence cannot decide query execution.", "conditionality")
		}
		report.RequirementOutcomes = append(report.RequirementOutcomes, out)
		record(out.Outcome, out.Applicability, out.Reasons)
		for i := range report.Inputs {
			if report.Inputs[i].InputID == item.InputID || item.Ownership.State == "unproved" && slices.Contains(item.Ownership.CandidateInputIDs, report.Inputs[i].InputID) {
				report.Inputs[i].RequirementIDs = append(report.Inputs[i].RequirementIDs, item.ID)
				input := &report.Inputs[i]
				if out.Outcome != "satisfied" {
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
			}
		}
		if item.Kind == "field" {
			coverage := Coverage{Dimension: "field_schema", State: "unavailable", InputID: item.InputID, Reasons: []Reason{}}
			if len(out.Schemas) == 1 {
				binding := out.Schemas[0].Binding
				coverage.SchemaID = binding.SchemaID
				coverage.ObjectID = binding.ObjectID
				coverage.State = binding.SourceCoverage
				if binding.Reason != "" {
					reason := newReason("schema_partial", binding.Reason, "field_schema")
					reason.InputID = item.InputID
					reason.RequirementID = item.ID
					reason.ObjectID = binding.ObjectID
					reason.SchemaID = binding.SchemaID
					for _, occurrence := range item.Occurrences {
						reason.Locations = append(reason.Locations, occurrence.Location)
						reason.ReferenceIDs = append(reason.ReferenceIDs, occurrence.ReferenceID)
					}
					if len(reason.Locations) > 0 {
						location := reason.Locations[0]
						reason.Location = &location
					}
					coverage.Reasons = append(coverage.Reasons, reason)
				}
			}
			if out.Outcome == "indeterminate" || out.Outcome == "ambiguous" {
				coverage.Reasons = append(coverage.Reasons, out.Reasons...)
			}
			report.Coverage = append(report.Coverage, coverage)
		} else if kind, ok := collectionKind(item.Kind); ok {
			c, supplied := p.env.Collection(kind)
			state := "unavailable"
			if supplied {
				state = c.Coverage
			}
			coverage := Coverage{Dimension: "environment_collection", CollectionKind: kind, State: state, InputID: item.InputID, Reasons: append([]Reason{}, out.Reasons...)}
			if item.Kind == "dataset" && item.InputID != "" {
				identity, _ := explicitIdentity(resolved[item.InputID].input)
				if resolved[item.InputID].binding != nil {
					identity = resolved[item.InputID].binding.Expected
				}
				coverage.CollectionKind = identity.Kind
				c, supplied = p.env.Collection(identity.Kind)
				coverage.State = "unavailable"
				if supplied {
					coverage.State = c.Coverage
				}
			}
			if supplied && c.Reason != "" {
				code := "environment_collection_partial"
				if c.Coverage == "unavailable" {
					code = "environment_collection_unavailable"
				}
				reason := newReason(code, c.Reason, coverage.Dimension)
				reason.InputID = item.InputID
				reason.RequirementID = item.ID
				for _, occurrence := range item.Occurrences {
					reason.Locations = append(reason.Locations, occurrence.Location)
					reason.ReferenceIDs = append(reason.ReferenceIDs, occurrence.ReferenceID)
				}
				if len(reason.Locations) > 0 {
					location := reason.Locations[0]
					reason.Location = &location
				}
				coverage.Reasons = append(coverage.Reasons, reason)
			}
			report.Coverage = append(report.Coverage, coverage)
		}
	}
	addDimension := func(dimension string, coverage analysis.InputCoverage, code string) {
		entry := Coverage{Dimension: dimension, State: coverage.State, Reasons: []Reason{}}
		for _, r := range coverage.Reasons {
			reason := newReason(code, r.Message, dimension)
			location := r.Location
			reason.Location = &location
			reason.Locations = append(reason.Locations, location)
			reason.ReferenceIDs = append(reason.ReferenceIDs, r.ReferenceIDs...)
			entry.Reasons = append(entry.Reasons, reason)
		}
		report.Coverage = append(report.Coverage, entry)
		if coverage.State == "partial" && dimension != "correlation_analysis" {
			unknown = true
			report.Reasons = append(report.Reasons, entry.Reasons...)
			if len(entry.Reasons) == 0 {
				reason := newReason(code, "Supply complete query evidence for this assessment dimension.", dimension)
				report.Reasons = append(report.Reasons, reason)
			}
		}
	}
	addDimension("target_discovery", set.InputCoverage, "target_discovery_incomplete")
	addDimension("field_attribution", set.FieldAttributionCoverage, "field_ownership_ambiguous")
	addDimension("correlation_analysis", set.Correlation.Coverage, "unsupported_semantics")
	semantics := Coverage{Dimension: "query_semantics", State: "complete", Reasons: []Reason{}}
	closureCoverage := Coverage{Dimension: "dependency_closure", State: "not_applicable", Reasons: []Reason{}}
	for _, gap := range set.Gaps {
		if correlationOnly(gap.Code) {
			continue
		}
		// An origin uncertainty linked entirely to conclusively satisfied obligations
		// cannot undo those positive facts. Missing/undecidable facts remain visible.
		if gap.Code == analysis.CodeRequirementIndeterminate && gapSatisfied(gap, set, report.RequirementOutcomes) {
			continue
		}

		code, dimension := "unsupported_semantics", "query_semantics"
		if gap.Code == analysis.CodeUnresolvedModule {
			code = "dependency_closure_incomplete"
			dimension = "dependency_closure"
		}
		reason := newReason(code, gap.Message, dimension)
		reason.ReferenceIDs = append(reason.ReferenceIDs, gap.ReferenceIDs...)
		for _, item := range set.Items {
			for _, o := range item.Occurrences {
				for _, ref := range gap.ReferenceIDs {
					if ref == o.ReferenceID {
						reason.RequirementID = item.ID
						reason.InputID = item.InputID
						reason.Locations = append(reason.Locations, o.Location)
					}
				}
			}
		}
		for _, diagnostic := range set.Diagnostics {
			for _, code := range gap.DiagnosticCodes {
				if code == diagnostic.Code {
					reason.Locations = append(reason.Locations, diagnostic.Location)
				}
			}
		}
		if len(reason.Locations) > 0 {
			location := reason.Locations[0]
			reason.Location = &location
		}
		if dimension == "dependency_closure" {
			closureCoverage.State = "partial"
			closureCoverage.Reasons = append(closureCoverage.Reasons, reason)
		} else {
			semantics.State = "partial"
			semantics.Reasons = append(semantics.Reasons, reason)
		}
		// This exact producer warning concerns merged output names. It does
		// not invalidate already-proved external reads; downstream ambiguous
		// reads still aggregate via ownership outcomes and attribution limits.
		if gap.Code != analysis.CodeAmbiguousField || gap.Message != "Join output contains fields with the same public name" {
			unknown = true
			report.Reasons = append(report.Reasons, reason)
		}
	}
	// Existence does not prove hidden expansion obligations. Task 9 can discharge
	// this conservative limit using effective closure requirements.
	for _, item := range set.Items {
		if item.Kind == "macro" || item.Kind == "saved_search" || item.Kind == "module" || item.Kind == "function" {
			reason := newReason("dependency_closure_incomplete", "Supply supported dependency closure evidence to assess reachable hidden requirements.", "dependency_closure")
			reason.RequirementID = item.ID
			for _, o := range item.Occurrences {
				reason.Locations = append(reason.Locations, o.Location)
			}
			if len(reason.Locations) > 0 {
				location := reason.Locations[0]
				reason.Location = &location
			}
			closureCoverage.State = "partial"
			closureCoverage.Reasons = append(closureCoverage.Reasons, reason)
			report.Reasons = append(report.Reasons, reason)
			unknown = true
		}
	}
	report.Coverage = append(report.Coverage, semantics, closureCoverage)
	if set.QueryStatus == analysis.Invalid || set.QueryStatus == analysis.Incomplete && len(set.Inputs) == 0 && len(set.Items) == 0 && (set.InputCoverage.State == "partial" || !set.Coverage.Complete) {
		report.Outcome = "not assessed"
		report.Reasons = append(report.Reasons, newReason("unsupported_semantics", "Query content is invalid or analysis stopped before trustworthy assessable obligations were established.", "query_semantics"))
	} else if missing {
		report.Outcome = "unsatisfied"
	} else if unknown {
		report.Outcome = "incomplete"
	}
	finalizeReport(report)
	return report
}

// Positive evidence can settle every linked origin obligation. No matching
// obligation, or any unresolved linked obligation, retains the coverage limit.
func gapSatisfied(gap analysis.RequirementGap, set analysis.RequirementSet, outcomes []RequirementOutcome) bool {
	if len(gap.ReferenceIDs) == 0 {
		return false
	}
	byID := map[string]string{}
	for _, out := range outcomes {
		byID[out.RequirementID] = out.Outcome
	}
	references := map[string]bool{}
	for _, item := range set.Items {
		for _, occurrence := range item.Occurrences {
			satisfied, known := references[occurrence.ReferenceID]
			references[occurrence.ReferenceID] = (!known || satisfied) && byID[item.ID] == "satisfied"
		}
	}
	for _, id := range gap.ReferenceIDs {
		if !references[id] {
			return false
		}
	}
	return true
}
