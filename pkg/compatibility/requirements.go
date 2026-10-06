package compatibility

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/delgado-jacob/spl-toolkit/internal/capabilityselector"
	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
)

func validDigest(value string) bool {
	if len(value) != 71 || !strings.HasPrefix(value, "sha256:") {
		return false
	}
	_, err := hex.DecodeString(value[7:])
	return err == nil && value == strings.ToLower(value)
}
func requirementError(path, message string) error {
	return requestErrorAt("requirements_inconsistent", "/requirements"+path, message)
}
func refreshError(path string) error {
	return requestErrorAt("requirements_refresh_required", "/requirements"+path, "revised requirement evidence is required")
}
func validLocation(location analysis.Location) bool {
	a, b := location.Start, location.End
	if a.Offset < 0 || b.Offset < a.Offset || a.Line < 0 || b.Line < a.Line || a.Column < 0 || b.Column < 0 {
		return false
	}
	if a.Line == b.Line && b.Column < a.Column {
		return false
	}
	// Zero locations are produced for unlocated coverage limits. Positive source
	// coordinates must have both a line and a column.
	return (a.Line == 0) == (a.Column == 0) && (b.Line == 0) == (b.Column == 0)
}
func validFieldIdentity(identity analysis.FieldIdentity) bool {
	if identity.Kind != "atomic" && identity.Kind != "path" {
		return false
	}
	if len(identity.Segments) == 0 || identity.Kind == "atomic" && (len(identity.Segments) != 1 || identity.Qualifier != "") {
		return false
	}
	for _, s := range identity.Segments {
		if !nonblank(s) {
			return false
		}
	}
	return true
}
func validateCoverage(coverage analysis.InputCoverage, path string) error {
	if coverage.State == "" {
		return refreshError(path + "/state")
	}
	if coverage.Reasons == nil {
		return refreshError(path + "/reasons")
	}
	if coverage.State != "complete" && coverage.State != "partial" && coverage.State != "not_applicable" {
		return requirementError(path+"/state", "invalid input coverage state")
	}
	for i, r := range coverage.Reasons {
		p := fmt.Sprintf("%s/reasons/%d", path, i)
		if !nonblank(r.Code) || !validLocation(r.Location) {
			return requirementError(p, "invalid located coverage reason")
		}
		if r.ReferenceIDs == nil {
			return refreshError(p + "/reference_ids")
		}
		if !uniqueNonblank(r.ReferenceIDs) {
			return requirementError(p+"/reference_ids", "reference links must be unique and nonblank")
		}
	}
	return nil
}
func uniqueNonblank(values []string) bool {
	seen := map[string]bool{}
	for _, id := range values {
		if !nonblank(id) || seen[id] {
			return false
		}
		seen[id] = true
	}
	return true
}
func validateRequirements(set analysis.RequirementSet) error {
	if set.SchemaVersion != 1 {
		return requirementError("/schema_version", "schema_version must be integer 1")
	}
	if set.Inputs == nil {
		return refreshError("/inputs")
	}
	if set.Items == nil {
		return refreshError("/items")
	}
	if set.Gaps == nil {
		return refreshError("/gaps")
	}
	if set.Diagnostics == nil {
		return refreshError("/diagnostics")
	}
	if set.Coverage.Reasons == nil {
		return refreshError("/coverage/reasons")
	}
	if set.CapabilityRevision == "" {
		return refreshError("/capability_revision")
	}
	query := set.Query
	if query.Language == "" || query.Profile == "" || query.Version == "" || !validDigest(query.QueryDigest) {
		return requirementError("/query", "query identity requires canonical selectors and a sha256 digest")
	}
	selection, err := capabilityselector.Normalize(query.Language, query.Profile, query.Version)
	if err != nil {
		return requirementError("/query", err.Error())
	}
	revision, err := analysis.CapabilityRevisionFor(analysis.CapabilityOptions{Language: selection.Language, Profile: selection.Profile, Version: selection.Version})
	if err != nil {
		return requirementError("/query", err.Error())
	}
	if set.CapabilityRevision != revision {
		return requestErrorAt("requirements_stale", "/requirements/capability_revision", "capability revision differs from the selected current semantics")
	}
	if set.QueryStatus != analysis.Valid && set.QueryStatus != analysis.Invalid && set.QueryStatus != analysis.Incomplete {
		return requirementError("/query_status", "invalid query status")
	}
	if err := validateCoverage(set.InputCoverage, "/input_coverage"); err != nil {
		return err
	}
	if err := validateCoverage(set.FieldAttributionCoverage, "/field_attribution_coverage"); err != nil {
		return err
	}
	inputs := map[string]analysis.QueryInput{}
	occurrences := map[string]string{}
	type sourceReferenceFact struct {
		location     analysis.Location
		stage, scope string
	}
	sourceRefs := map[string]sourceReferenceFact{}
	recordSourceReference := func(id string, location analysis.Location, stage, scope, path string) error {
		if id == "" {
			return nil
		}
		prior, known := sourceRefs[id]
		if known && (prior.location != location || prior.stage != "" && stage != "" && prior.stage != stage || prior.scope != "" && scope != "" && prior.scope != scope) {
			return requirementError(path, "contradictory source reference")
		}
		if known {
			if stage == "" {
				stage = prior.stage
			}
			if scope == "" {
				scope = prior.scope
			}
		}
		sourceRefs[id] = sourceReferenceFact{location: location, stage: stage, scope: scope}
		return nil
	}
	identities := map[analysis.InputIdentity]map[string]bool{}
	for i, input := range set.Inputs {
		path := fmt.Sprintf("/inputs/%d", i)
		if !nonblank(input.ID) {
			return requirementError(path+"/id", "input id must be nonblank")
		}
		if _, exists := inputs[input.ID]; exists {
			return requirementError(path+"/id", "duplicate input id")
		}
		if input.Identity.Form == "" {
			return refreshError(path + "/identity/form")
		}
		valid := false
		switch input.Kind {
		case "named_placeholder":
			valid = input.Identity.Form == "parameter" && strings.HasPrefix(input.Identity.Value, "$") && len(input.Identity.Value) > 1
		case "explicit_dataset":
			valid = input.Identity.Form == "identifier" || input.Identity.Form == "dotted" || input.Identity.Form == "descriptor"
		case "implicit_stream":
			valid = input.Identity.Form == "implicit" && input.Identity.Value == ""
		case "unresolved_source":
			valid = input.Identity.Form == "unresolved" || input.Identity.Form == "descriptor"
		}
		if !valid || input.Kind != "implicit_stream" && (!nonblank(input.Identity.Value) || input.Name == "") {
			return requirementError(path+"/identity", "invalid typed input identity")
		}
		if input.Kind != "unresolved_source" && input.Name != input.Identity.Value {
			return requirementError(path+"/identity", "input name disagrees with typed identity")
		}
		if input.Kind == "explicit_dataset" && input.Identity.Form == "descriptor" && !validDescriptor(input.Identity.Value) {
			return requirementError(path+"/identity", "descriptor identity must preserve the complete canonical literal object")
		}
		if identities[input.Identity] == nil {
			identities[input.Identity] = map[string]bool{}
		}
		if identities[input.Identity][input.Kind] {
			return requirementError(path+"/identity", "duplicate logical input identity")
		}
		identities[input.Identity][input.Kind] = true
		if err := validateCoverage(input.Evidence, path+"/evidence"); err != nil {
			return err
		}
		if input.Occurrences == nil {
			return refreshError(path + "/occurrences")
		}
		if len(input.Occurrences) == 0 {
			return requirementError(path+"/occurrences", "input requires an occurrence")
		}
		for j, o := range input.Occurrences {
			p := fmt.Sprintf("%s/occurrences/%d", path, j)
			if !nonblank(o.ID) || occurrences[o.ID] != "" {
				return requirementError(p+"/id", "occurrence ids must be unique and nonblank")
			}
			if !nonblank(o.StageID) || !nonblank(o.ScopeID) || !validLocation(o.Location) {
				return requirementError(p, "invalid situated input occurrence")
			}
			if o.UseSiteLocations == nil {
				return refreshError(p + "/use_site_locations")
			}
			if o.UseSiteReferenceIDs == nil {
				return refreshError(p + "/use_site_reference_ids")
			}
			if len(o.UseSiteLocations) != len(o.UseSiteReferenceIDs) || !uniqueNonblank(o.UseSiteReferenceIDs) {
				return requirementError(p+"/use_site_reference_ids", "inconsistent use-site links")
			}
			for _, loc := range o.UseSiteLocations {
				if !validLocation(loc) {
					return requirementError(p+"/use_site_locations", "invalid use-site location")
				}
			}
			// Situated views retain the original source location while replacing
			// reference/stage/scope with the terminal use. Original coordinates
			// have no corresponding original stage/scope in this public shape.
			if err := recordSourceReference(o.OriginalReferenceID, o.Location, "", "", p+"/original_reference_id"); err != nil {
				return err
			}
			if len(o.UseSiteReferenceIDs) > 0 && o.ReferenceID != o.UseSiteReferenceIDs[len(o.UseSiteReferenceIDs)-1] {
				return requirementError(p+"/reference_id", "terminal reference disagrees with the use-site chain")
			}
			for k, id := range o.UseSiteReferenceIDs {
				stage, scope := "", ""
				if k == len(o.UseSiteReferenceIDs)-1 {
					stage, scope = o.StageID, o.ScopeID
				}
				if err := recordSourceReference(id, o.UseSiteLocations[k], stage, scope, fmt.Sprintf("%s/use_site_locations/%d", p, k)); err != nil {
					return err
				}
			}
			if len(o.UseSiteLocations) == 0 {
				if err := recordSourceReference(o.ReferenceID, o.Location, o.StageID, o.ScopeID, p+"/reference_id"); err != nil {
					return err
				}
			}
			occurrences[o.ID] = input.ID
		}
		inputs[input.ID] = input
	}
	type referenceFact struct {
		kind, identity, role, original, binding, stage, scope string
		location                                              analysis.Location
	}
	referenceFacts := map[string]referenceFact{}
	itemIDs := map[string]bool{}
	referenceEvidence := map[string]correlationReferenceFact{}
	for i, item := range set.Items {
		path := fmt.Sprintf("/items/%d", i)
		if !nonblank(item.ID) || itemIDs[item.ID] {
			return requirementError(path+"/id", "requirement ids must be unique and nonblank")
		}
		itemIDs[item.ID] = true
		if !nonblank(item.Kind) || !nonblank(item.Identity) || !nonblank(item.Role) || item.Necessity != "required" && item.Necessity != "conditional" || item.Origin != "direct" || item.Resolution != "exact" && item.Resolution != "wildcard" && item.Resolution != "dynamic" {
			return requirementError(path, "invalid requirement identity or classification")
		}
		if item.FieldIdentity != nil && item.FieldIdentity.Segments == nil {
			return refreshError(path + "/field_identity/segments")
		}
		if item.FieldIdentity != nil && item.Identity != strings.Join(item.FieldIdentity.Segments, ".") {
			return requirementError(path+"/field_identity", "field identity disagrees with requirement identity")
		}
		if item.FieldIdentity != nil && !validFieldIdentity(*item.FieldIdentity) {
			return requirementError(path+"/field_identity", "invalid typed field identity")
		}
		if item.Kind == "field" && item.Resolution == "exact" && item.FieldIdentity == nil {
			return refreshError(path + "/field_identity")
		}
		if item.Kind != "field" && item.FieldIdentity != nil {
			return requirementError(path+"/field_identity", "field identity on a non-field requirement")
		}
		if item.Ownership.CandidateInputIDs == nil {
			return refreshError(path + "/ownership/candidate_input_ids")
		}
		if !uniqueNonblank(item.Ownership.CandidateInputIDs) {
			return requirementError(path+"/ownership/candidate_input_ids", "ownership links must be unique and nonblank")
		}
		for _, id := range item.Ownership.CandidateInputIDs {
			if _, known := inputs[id]; !known {
				return requirementError(path+"/ownership/candidate_input_ids", "unknown owner input")
			}
		}
		if item.InputID != "" {
			if _, known := inputs[item.InputID]; !known {
				return requirementError(path+"/input_id", "unknown proved input")
			}
		}
		if item.Ownership.State == "" {
			return refreshError(path + "/ownership/state")
		}
		switch item.Ownership.State {
		case "proved":
			if item.InputID == "" || len(item.Ownership.CandidateInputIDs) != 1 || item.Ownership.CandidateInputIDs[0] != item.InputID {
				return requirementError(path+"/ownership", "proved ownership requires exactly its proved input")
			}
		case "unproved":
			if item.InputID != "" {
				return requirementError(path+"/input_id", "unproved ownership cannot select an input")
			}
		default:
			return requirementError(path+"/ownership/state", "invalid ownership state")
		}
		if item.Occurrences == nil {
			return refreshError(path + "/occurrences")
		}
		if len(item.Occurrences) == 0 {
			return requirementError(path+"/occurrences", "requirement requires an occurrence")
		}
		refs := map[string]bool{}
		for j, o := range item.Occurrences {
			p := fmt.Sprintf("%s/occurrences/%d", path, j)
			if !nonblank(o.ReferenceID) || refs[o.ReferenceID] {
				return requirementError(p+"/reference_id", "requirement reference links must be unique and nonblank")
			}
			refs[o.ReferenceID] = true
			if !nonblank(o.StageID) || !nonblank(o.ScopeID) || !validLocation(o.Location) || o.Necessity != "required" && o.Necessity != "conditional" {
				return requirementError(p, "invalid requirement occurrence")
			}
			if o.InputOccurrenceIDs == nil {
				return refreshError(p + "/input_occurrence_ids")
			}
			if item.InputID != "" && len(o.InputOccurrenceIDs) == 0 {
				return requirementError(p+"/input_occurrence_ids", "proved ownership requires situated occurrence evidence")
			}
			if !uniqueNonblank(o.InputOccurrenceIDs) {
				return requirementError(p+"/input_occurrence_ids", "occurrence links must be unique and nonblank")
			}
			for _, id := range o.InputOccurrenceIDs {
				inputID, known := occurrences[id]
				if !known || item.InputID != "" && inputID != item.InputID {
					return requirementError(p+"/input_occurrence_ids", "inconsistent input occurrence link")
				}
				ownerKnown := false
				for _, candidate := range item.Ownership.CandidateInputIDs {
					ownerKnown = ownerKnown || candidate == inputID
				}
				if !ownerKnown {
					return requirementError(p+"/input_occurrence_ids", "occurrence is outside candidate ownership")
				}
			}
			fact := referenceFact{item.Kind, item.Identity, item.Role, o.OriginalName, o.Binding, o.StageID, o.ScopeID, o.Location}
			if prior, known := referenceFacts[o.ReferenceID]; known && prior != fact {
				return requirementError(p+"/reference_id", "contradictory reference identity or location")
			}
			referenceFacts[o.ReferenceID] = fact
			if prior, known := referenceEvidence[o.ReferenceID]; known && prior.inputID != item.InputID {
				return requirementError(p+"/reference_id", "contradictory proved reference owner")
			}
			evidence := correlationReferenceFact{inputID: item.InputID, kind: item.Kind, location: o.Location}
			// A conditional or derived destination can retain a requirement without
			// establishing the source field identity used by this reference.
			if item.Kind == "field" && item.Resolution == "exact" && item.Ownership.State == "proved" && o.Binding == "source" {
				evidence.sourceField = item.FieldIdentity
			}
			if prior, known := referenceEvidence[o.ReferenceID]; known && prior.sourceField != nil {
				if evidence.sourceField != nil && !sameFieldIdentity(*prior.sourceField, *evidence.sourceField) {
					return requirementError(p+"/reference_id", "contradictory proved source field identity")
				}
				// Later incomplete evidence cannot erase an available source proof.
				if evidence.sourceField == nil {
					evidence.sourceField = prior.sourceField
				}
			}
			referenceEvidence[o.ReferenceID] = evidence
			if inputRef, known := sourceRefs[o.ReferenceID]; known && (inputRef.location != o.Location || inputRef.stage != "" && inputRef.stage != o.StageID || inputRef.scope != "" && inputRef.scope != o.ScopeID) {
				return requirementError(p+"/reference_id", "contradictory source reference location")
			}
		}
	}
	for i, gap := range set.Gaps {
		p := fmt.Sprintf("/gaps/%d", i)
		if !nonblank(gap.Code) {
			return requirementError(p+"/code", "gap code must be nonblank")
		}
		if gap.ReferenceIDs == nil || gap.DiagnosticCodes == nil {
			return refreshError(p)
		}
		if !uniqueNonblank(gap.ReferenceIDs) || !uniqueNonblank(gap.DiagnosticCodes) {
			return requirementError(p, "gap links must be unique and nonblank")
		}
	}
	for i, d := range set.Diagnostics {
		if !validLocation(d.Location) {
			return requirementError(fmt.Sprintf("/diagnostics/%d/location", i), "invalid diagnostic location")
		}
	}
	return validateCorrelation(set.Correlation, occurrences, referenceEvidence)
}

// The Items projection is a partial reference inventory. Retain only facts
// actually submitted there; absent derived-key references remain unknown.
type correlationReferenceFact struct {
	inputID, kind string
	location      analysis.Location
	sourceField   *analysis.FieldIdentity
}

func sameFieldIdentity(a, b analysis.FieldIdentity) bool {
	return a.Kind == b.Kind && a.Qualifier == b.Qualifier && slices.Equal(a.Segments, b.Segments)
}
func sameCorrelationEndpoint(a, b analysis.CorrelationEndpoint) bool {
	return a.InputID == b.InputID && a.OccurrenceID == b.OccurrenceID && sameFieldIdentity(a.FieldIdentity, b.FieldIdentity) && slices.Equal(a.ReferenceIDs, b.ReferenceIDs) && a.Location == b.Location
}
func validateCorrelation(graph analysis.CorrelationGraph, occurrences map[string]string, referenceEvidence map[string]correlationReferenceFact) error {
	if graph.Nodes == nil {
		return refreshError("/correlation/nodes")
	}
	if graph.Edges == nil {
		return refreshError("/correlation/edges")
	}
	if graph.Components == nil {
		return refreshError("/correlation/components")
	}
	if err := validateCoverage(graph.Coverage, "/correlation/coverage"); err != nil {
		return err
	}
	if graph.Outcome != "connected" && graph.Outcome != "disconnected" && graph.Outcome != "indeterminate" && graph.Outcome != "not applicable" {
		return requirementError("/correlation/outcome", "invalid query correlation outcome")
	}
	nodes := map[string]string{}
	for i, n := range graph.Nodes {
		p := fmt.Sprintf("/correlation/nodes/%d", i)
		if occurrences[n.OccurrenceID] != n.InputID || n.InputID == "" || nodes[n.OccurrenceID] != "" {
			return requirementError(p, "node must identify one known occurrence of its input")
		}
		nodes[n.OccurrenceID] = n.InputID
	}
	if len(nodes) != len(occurrences) {
		return requirementError("/correlation/nodes", "graph omits input occurrences")
	}
	endpoint := func(e analysis.CorrelationEndpoint, path string) error {
		if nodes[e.OccurrenceID] != e.InputID || e.InputID == "" {
			return requirementError(path, "endpoint disagrees with its node")
		}
		if e.FieldIdentity.Segments == nil {
			return refreshError(path + "/field_identity/segments")
		}
		if !validFieldIdentity(e.FieldIdentity) || !validLocation(e.Location) {
			return requirementError(path, "invalid typed correlation endpoint")
		}
		if e.ReferenceIDs == nil {
			return refreshError(path + "/reference_ids")
		}
		if len(e.ReferenceIDs) == 0 || !uniqueNonblank(e.ReferenceIDs) {
			return requirementError(path+"/reference_ids", "endpoint requires unique reference links")
		}
		for _, id := range e.ReferenceIDs {
			fact, known := referenceEvidence[id]
			if !known {
				continue
			}
			if fact.inputID != "" && fact.inputID != e.InputID {
				return requirementError(path+"/reference_ids", "endpoint contradicts a proved reference owner")
			}
			if fact.kind != "field" {
				return requirementError(path+"/reference_ids", "endpoint names a known non-field reference")
			}
			if fact.location != e.Location {
				return requirementError(path+"/location", "endpoint contradicts the known reference location")
			}
			if fact.sourceField != nil && !sameFieldIdentity(*fact.sourceField, e.FieldIdentity) {
				return requirementError(path+"/field_identity", "endpoint contradicts the proved source field identity")
			}
		}
		return nil
	}
	components := map[string]int{}
	for i, component := range graph.Components {
		if component == nil {
			return refreshError(fmt.Sprintf("/correlation/components/%d", i))
		}
		if len(component) == 0 || !uniqueNonblank(component) {
			return requirementError("/correlation/components", "components must contain unique occurrences")
		}
		for _, id := range component {
			if _, known := nodes[id]; !known {
				return requirementError("/correlation/components", "component contains unknown node")
			}
			if _, duplicate := components[id]; duplicate {
				return requirementError("/correlation/components", "node appears in multiple components")
			}
			components[id] = i
		}
	}
	if len(components) != len(nodes) {
		return requirementError("/correlation/components", "components omit nodes")
	}
	parents := map[string]string{}
	for id := range nodes {
		parents[id] = id
	}
	root := func(id string) string {
		for parents[id] != id {
			id = parents[id]
		}
		return id
	}
	edges := map[string]bool{}
	for i, e := range graph.Edges {
		p := fmt.Sprintf("/correlation/edges/%d", i)
		if !nonblank(e.ID) || edges[e.ID] {
			return requirementError(p+"/id", "edge ids must be unique and nonblank")
		}
		edges[e.ID] = true
		if err := endpoint(e.Left, p+"/left"); err != nil {
			return err
		}
		if err := endpoint(e.Right, p+"/right"); err != nil {
			return err
		}
		if e.Left.OccurrenceID == e.Right.OccurrenceID || components[e.Left.OccurrenceID] != components[e.Right.OccurrenceID] {
			return requirementError(p, "contradictory graph edge")
		}
		parents[root(e.Right.OccurrenceID)] = root(e.Left.OccurrenceID)
		if !nonblank(e.StageID) || !nonblank(e.ScopeID) || !validLocation(e.Location) {
			return requirementError(p, "invalid situated edge")
		}
		if e.Keys == nil {
			return refreshError(p + "/keys")
		}
		if len(e.Keys) == 0 {
			return requirementError(p+"/keys", "edge requires predicate evidence")
		}
		for j, key := range e.Keys {
			kp := fmt.Sprintf("%s/keys/%d", p, j)
			if key.Predicate != "equals" || !validLocation(key.Location) {
				return requirementError(kp, "invalid key predicate")
			}
			if err := endpoint(key.Left, kp+"/left"); err != nil {
				return err
			}
			if err := endpoint(key.Right, kp+"/right"); err != nil {
				return err
			}
			// Edge endpoints are a copy of the first predicate endpoints. Later
			// AND predicates retain their own fields, references, and locations.
			if j == 0 {
				if !sameCorrelationEndpoint(key.Left, e.Left) {
					return requirementError(kp+"/left", "first key endpoint contradicts its edge endpoint")
				}
				if !sameCorrelationEndpoint(key.Right, e.Right) {
					return requirementError(kp+"/right", "first key endpoint contradicts its edge endpoint")
				}
			}
			if key.Left.InputID != e.Left.InputID || key.Left.OccurrenceID != e.Left.OccurrenceID || key.Right.InputID != e.Right.InputID || key.Right.OccurrenceID != e.Right.OccurrenceID {
				return requirementError(kp, "key endpoints contradict their edge")
			}
		}
	}
	for _, component := range graph.Components {
		first := root(component[0])
		for _, id := range component[1:] {
			if root(id) != first {
				return requirementError("/correlation/components", "declared component is not connected by proved edges")
			}
		}
	}
	if graph.Outcome == "connected" && (len(nodes) <= 1 || len(graph.Components) != 1) || graph.Outcome == "disconnected" && len(graph.Components) <= 1 || graph.Outcome == "not applicable" && len(nodes) > 1 {
		return requirementError("/correlation/outcome", "outcome contradicts the graph components")
	}
	return nil
}

func validDescriptor(identity string) bool {
	decoder := json.NewDecoder(strings.NewReader(identity))
	decoder.UseNumber()
	var descriptor map[string]any
	if decoder.Decode(&descriptor) != nil || descriptor == nil {
		return false
	}
	if _, err := decoder.Token(); err != io.EOF {
		return false
	}
	kind, ok := descriptor["kind"].(string)
	if !ok || !nonblank(kind) {
		return false
	}
	for key, value := range descriptor {
		switch key {
		case "kind":
		case "properties":
			if _, ok := value.(map[string]any); !ok {
				return false
			}
		default:
			return false
		}
	}
	canonical, err := json.Marshal(descriptor)
	return err == nil && bytes.Equal(canonical, []byte(identity))
}
