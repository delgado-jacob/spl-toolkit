package compatibility

import (
	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/environment"
)

// collectionKind is an exact mapping of canonical requirement kinds, never a
// lookup of arbitrary query strings or a product version inference.
func collectionKind(kind string) (string, bool) {
	switch kind {
	case "index", "source", "sourcetype", "dataset", "data_model", "lookup", "macro", "saved_search", "event_type", "tag", "calculated_field", "field_extraction", "module", "function", "external_command":
		return kind, true
	}
	return "", false
}
func contextual(kind string) bool { return kind != "index" && kind != "source" && kind != "sourcetype" }
func contextInScope(scope environment.CaptureScope, identity environment.ObjectIdentity) bool {
	return !contextual(identity.Kind) || selected(scope.Namespace, identity.Namespace) && selected(scope.App, identity.App) && selected(scope.Owner, identity.Owner)
}
func selectorCovers(capture, query environment.Selector) bool {
	if capture.All != nil {
		return *capture.All
	}
	if query.All != nil {
		return false
	}
	for _, value := range query.Values {
		if !selected(capture, value) {
			return false
		}
	}
	return true
}
func scopeCovers(capture, query environment.CaptureScope) bool {
	return selectorCovers(capture.Namespace, query.Namespace) && selectorCovers(capture.App, query.App) && selectorCovers(capture.Owner, query.Owner)
}
func objectEvidence(r resolvedInput) []ObjectEvidence {
	out := []ObjectEvidence{}
	if r.binding != nil {
		e := ObjectEvidence{ObjectID: r.binding.ObjectID, Expected: r.binding.Expected}
		if len(r.objects) == 1 {
			object := detach(r.objects[0])
			e.Object = &object
		}
		return append(out, e)
	}
	for _, object := range r.objects {
		o := detach(object)
		out = append(out, ObjectEvidence{ObjectID: o.ID, Expected: objectIdentity(o), Object: &o})
	}
	return out
}
func inputReason(r resolvedInput, code, message string) Reason {
	reason := newReason(code, message, "target_discovery")
	reason.InputID = r.input.ID
	for _, occurrence := range r.input.Occurrences {
		reason.Locations = append(reason.Locations, occurrence.Location)
		if occurrence.ReferenceID != "" {
			reason.ReferenceIDs = append(reason.ReferenceIDs, occurrence.ReferenceID)
		}
	}
	if len(reason.Locations) > 0 {
		location := reason.Locations[0]
		reason.Location = &location
	}
	if r.binding != nil {
		reason.ObjectID = r.binding.ObjectID
		reason.SchemaID = r.binding.SchemaID
	}
	return reason
}
func (p *Prepared) assessInput(r resolvedInput, scope environment.CaptureScope) (InputOutcome, Coverage) {
	out := InputOutcome{InputID: r.input.ID, Outcome: "indeterminate", Objects: objectEvidence(r), RequirementIDs: []string{}, Reasons: []Reason{}}
	coverage := Coverage{Dimension: "environment_collection", State: "unavailable", InputID: r.input.ID, Reasons: []Reason{}}
	identity, mapped := explicitIdentity(r.input)
	if r.binding != nil {
		identity = r.binding.Expected
		mapped = true
		coverage.ObjectID = r.binding.ObjectID
	}
	fail := func(code, message string) (InputOutcome, Coverage) {
		reason := inputReason(r, code, message)
		reason.Dimension = coverage.Dimension
		out.Reasons = append(out.Reasons, reason)
		coverage.Reasons = append(coverage.Reasons, reason)
		return out, coverage
	}
	if !mapped || !r.explicitSourceEvidence {
		coverage.Dimension = "target_discovery"
		coverage.State = "partial"
		if r.input.Kind == "explicit_dataset" && r.input.Identity.Form == "descriptor" && !mapped {
			reason := inputReason(r, "dataset_identity_unmapped", "The complete static Dataset descriptor cannot be mapped by the supported typed identity rules; supply supported exact source evidence.")
			reason.Dimension = coverage.Dimension
			out.Reasons = append(out.Reasons, reason)
			coverage.Reasons = append(coverage.Reasons, reason)
		}
		return fail("target_discovery_incomplete", "Supply exact query source evidence; the input cannot yet be mapped to a complete supplying source.")
	}
	collection, supplied := p.env.Collection(identity.Kind)
	if supplied {
		coverage.State = collection.Coverage
		coverage.CollectionKind = identity.Kind
		if collection.Reason != "" {
			code := "environment_collection_partial"
			if collection.Coverage == "unavailable" {
				code = "environment_collection_unavailable"
			}
			reason := inputReason(r, code, collection.Reason)
			reason.Dimension = coverage.Dimension
			coverage.Reasons = append(coverage.Reasons, reason)
		}
	}
	eligible := 0
	for _, object := range r.objects {
		if contextInScope(scope, objectIdentity(object)) && contextInScope(p.snapshot.CaptureScope, objectIdentity(object)) {
			eligible++
		}
	}
	if eligible == 1 {
		out.Outcome = "satisfied"
		return out, coverage
	}
	if eligible > 1 {
		out.Outcome = "ambiguous"
		return fail("environment_object_ambiguous", "Several exact eligible objects match; select one with an identity-consistent input binding.")
	}
	covered := !contextual(identity.Kind)
	if contextual(identity.Kind) {
		if r.binding != nil || r.input.Identity.Form == "descriptor" {
			covered = contextInScope(scope, identity) && contextInScope(p.snapshot.CaptureScope, identity)
		} else {
			covered = scopeCovers(p.snapshot.CaptureScope, scope)
		}
	}
	if !covered {
		return fail("environment_scope_not_covered", "Capture scope does not cover this exact identity or the full requested query scope; supply covered evidence.")
	}
	if !supplied || collection.Coverage == "unavailable" {
		return fail("environment_collection_unavailable", "Supply the relevant object collection to assess absence.")
	}
	if collection.Coverage != "complete" {
		return fail("environment_collection_partial", "Complete the relevant collection before absence can prove the object missing.")
	}
	if p.snapshot.Observation != nil && (identity.Kind == "source" || identity.Kind == "sourcetype") {
		return fail("observation_scope_insufficient", "Captured not_observed evidence is limited to the recorded visibility, indexes, datatypes, method and window; it cannot prove global absence.")
	}
	out.Outcome = "missing"
	return fail("environment_object_missing", "The exact selected object is absent from a complete covered collection.")
}
func (p *Prepared) assessObject(item analysis.RequirementItem, query analysis.RequirementQueryIdentity, scope environment.CaptureScope) RequirementOutcome {
	out := baseOutcome(item, query)
	if item.Resolution != "exact" {
		return outcomeReason(out, item, "unsupported_semantics", "A dynamic or wildcard identity needs proof of the entire obligation.", "environment_collection")
	}
	kind, known := collectionKind(item.Kind)
	if !known {
		return outcomeReason(out, item, "unsupported_semantics", "No supported canonical collection mapping exists for this requirement kind.", "query_semantics")
	}
	for _, object := range p.snapshot.Objects {
		identity := objectIdentity(object)
		if object.Kind == kind && object.Name == item.Identity && contextInScope(scope, identity) && contextInScope(p.snapshot.CaptureScope, identity) {
			o := detach(object)
			out.Objects = append(out.Objects, ObjectEvidence{ObjectID: o.ID, Expected: identity, Object: &o})
		}
	}
	if len(out.Objects) == 1 {
		out.Outcome = "satisfied"
		return out
	}
	if len(out.Objects) > 1 {
		out.Outcome = "ambiguous"
		return outcomeReason(out, item, "environment_object_ambiguous", "Several eligible exact objects match this requirement; supply disambiguating dependency evidence.", "environment_collection")
	}
	collection, supplied := p.env.Collection(kind)
	if contextual(kind) && !scopeCovers(p.snapshot.CaptureScope, scope) {
		return outcomeReason(out, item, "environment_scope_not_covered", "Supply capture evidence covering the requested query scope.", "environment_collection")
	}
	if !supplied || collection.Coverage == "unavailable" {
		return outcomeReason(out, item, "environment_collection_unavailable", "Supply the relevant object collection to assess absence.", "environment_collection")
	}
	if collection.Coverage != "complete" {
		return outcomeReason(out, item, "environment_collection_partial", "Supply a complete relevant collection to prove absence.", "environment_collection")
	}
	if p.snapshot.Observation != nil && (kind == "source" || kind == "sourcetype") {
		return outcomeReason(out, item, "observation_scope_insufficient", "Observed absence cannot prove global source or sourcetype nonexistence.", "environment_collection")
	}
	out.Outcome = "missing"
	return outcomeReason(out, item, "environment_object_missing", "The exact requirement is absent from a complete covered collection.", "environment_collection")
}
