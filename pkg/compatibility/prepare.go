package compatibility

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/environment"
)

// Prepared owns one detached environment and reusable compiled schema targets.
// Query bindings belong exclusively to individual Check calls.
type Prepared struct {
	env      *environment.PreparedEnvironment
	snapshot environment.Snapshot
	schemas  *environment.SchemaBundle
}

func Prepare(snapshot environment.Snapshot, schemas *environment.SchemaBundle) (*Prepared, error) {
	for _, array := range []struct {
		name    string
		missing bool
	}{{"capabilities", snapshot.Capabilities == nil}, {"collections", snapshot.Collections == nil}, {"objects", snapshot.Objects == nil}} {
		if array.missing {
			return nil, requestErrorAt("snapshot_invalid", "/snapshot/"+array.name, array.name+" must be an array")
		}
	}
	if schemas != nil {
		for _, array := range []struct {
			name    string
			missing bool
		}{{"schemas", schemas.Schemas == nil}, {"bindings", schemas.Bindings == nil}} {
			if array.missing {
				return nil, requestErrorAt("schema_bundle_invalid", "/schema_bundle/"+array.name, array.name+" must be an array")
			}
		}
	}

	prepared, report, err := environment.PrepareSnapshot(snapshot)
	if err != nil {
		return nil, err
	}
	if err := artifactError(report, "/snapshot", 0); err != nil {
		return nil, err
	}
	var bundle *environment.PreparedSchemaBundle
	if schemas != nil {
		bundle, report, err = environment.PrepareSchemaBundle(*schemas)
		if err != nil {
			return nil, err
		}
		if err := artifactError(report, "/schema_bundle", 0); err != nil {
			return nil, err
		}
	}
	env, report, err := environment.Pair(prepared, bundle)
	if err != nil {
		return nil, err
	}
	if err := artifactError(report, "/schema_bundle", 0); err != nil {
		return nil, err
	}
	out := &Prepared{env: env, snapshot: prepared.Snapshot()}
	if bundle != nil {
		normalized := bundle.Bundle()
		out.schemas = &normalized
	}
	return out, nil
}
func CheckJSON(raw []byte) (*Report, error) {
	request, err := DecodeRequest(raw)
	if err != nil {
		return nil, err
	}
	return Check(request)
}
func Check(request Request) (*Report, error) {
	assessment, err := normalizeAssessment(request.assessment())
	if err != nil {
		return nil, err
	}
	// Validate direct configuration before query content aggregation or optional
	// closure discovery. Effective inputs are validated separately after discovery.
	if err := validateBindingSet(assessment.Requirements.Inputs, assessment.InputBindings); err != nil {
		return nil, err
	}
	prepared, err := Prepare(request.Snapshot, request.SchemaBundle)
	if err != nil {
		return nil, err
	}
	return prepared.checkNormalized(assessment)
}
func (p *Prepared) Check(request AssessmentRequest) (*Report, error) {
	assessment, err := normalizeAssessment(request)
	if err != nil {
		return nil, err
	}
	return p.checkNormalized(assessment)
}
func (p *Prepared) checkNormalized(request AssessmentRequest) (*Report, error) {
	if p == nil || p.env == nil {
		return nil, requestErrorAt("request_invalid", "/snapshot", "prepared environment is required")
	}
	resolved, err := p.resolveInputs(request.Requirements.Inputs, request.QueryScope, request.InputBindings)
	if err != nil {
		return nil, err
	}
	// Evaluation and aggregation are added by the assessment owner. Admission
	// alone never claims that content obligations have been satisfied.
	environmentReport := p.env.Report()
	report := &Report{SchemaVersion: 1, Outcome: "incomplete", Requirements: request.Requirements, Correlation: detach(request.Requirements.Correlation), Inputs: []InputOutcome{}, RequirementOutcomes: []RequirementOutcome{}, Coverage: []Coverage{}, Diagnostics: environmentReport.Diagnostics, Reasons: []Reason{}, Provenance: Provenance{QueryDigest: request.Requirements.Query.QueryDigest, SourceID: request.Requirements.Query.SourceID, CapabilityRevision: request.Requirements.CapabilityRevision, AnalysisContractVersion: 1, RequirementSetVersion: request.Requirements.SchemaVersion, EnvironmentDigest: environmentReport.SnapshotDigest, SchemaBundleDigest: environmentReport.SchemaBundleDigest}}
	for _, input := range request.Requirements.Inputs {
		r := resolved[input.ID]
		evidence := []ObjectEvidence{}
		if r.binding != nil {
			e := ObjectEvidence{ObjectID: r.binding.ObjectID, Expected: r.binding.Expected}
			if len(r.objects) == 1 {
				object := r.objects[0]
				e.Object = &object
			}
			evidence = append(evidence, e)
		} else {
			for _, object := range r.objects {
				o := object
				evidence = append(evidence, ObjectEvidence{ObjectID: o.ID, Expected: objectIdentity(o), Object: &o})
			}
		}
		report.Inputs = append(report.Inputs, InputOutcome{InputID: input.ID, Outcome: "indeterminate", Objects: evidence, RequirementIDs: []string{}, Reasons: []Reason{}})
	}
	return report, nil
}

// validateBindingSet receives the discovered set rather than reading only the
// direct requirement set, so effective closure discovery can use the same gate.
func validateBindingSet(inputs []analysis.QueryInput, bindings []InputBinding) error {
	known := map[string]analysis.QueryInput{}
	for _, input := range inputs {
		known[input.ID] = input
	}
	seen := map[string]bool{}
	objects := map[string]environment.ObjectIdentity{}
	for i, b := range bindings {
		path := fmt.Sprintf("/input_bindings/%d", i)
		if _, exists := known[b.InputID]; !exists || !nonblank(b.InputID) || seen[b.InputID] {
			return requestErrorAt("binding_invalid", path+"/input_id", "binding requires one known input id without duplicates")
		}
		seen[b.InputID] = true
		if !nonblank(b.ObjectID) {
			return requestErrorAt("binding_invalid", path+"/object_id", "object id must be nonblank")
		}
		if err := validateExpected(b.Expected, path+"/expected"); err != nil {
			return err
		}
		if expected, exists := objects[b.ObjectID]; exists && expected != b.Expected {
			return requestErrorAt("binding_invalid", path+"/object_id", "object id has contradictory expected identities")
		}
		objects[b.ObjectID] = b.Expected
	}
	for _, input := range inputs {
		if input.Kind == "named_placeholder" && !seen[input.ID] {
			return requestErrorAt("missing_input_binding", "/input_bindings", "each discovered placeholder requires exactly one binding")
		}
	}
	return nil
}
func validateExpected(expected environment.ObjectIdentity, path string) error {
	if expected.Kind != "dataset" && expected.Kind != "index" && expected.Kind != "source" && expected.Kind != "sourcetype" {
		return requestErrorAt("binding_invalid", path+"/kind", "input bindings select dataset, index, source, or sourcetype")
	}
	if !nonblank(expected.Name) {
		return requestErrorAt("binding_invalid", path+"/name", "expected name must be nonblank")
	}
	if expected.Kind != "dataset" && (expected.Namespace != "" || expected.App != "" || expected.Owner != "") {
		return requestErrorAt("binding_invalid", path, "global source identity has inapplicable context")
	}
	return nil
}
func objectIdentity(o environment.Object) environment.ObjectIdentity {
	return environment.ObjectIdentity{Kind: o.Kind, Name: o.Name, Namespace: o.Namespace, App: o.App, Owner: o.Owner}
}
func selected(selector environment.Selector, value string) bool {
	if selector.All != nil {
		return *selector.All
	}
	for _, candidate := range selector.Values {
		if candidate == value {
			return true
		}
	}
	return false
}
func identityInScope(scope environment.CaptureScope, identity environment.ObjectIdentity) bool {
	if identity.Kind != "dataset" {
		return true
	}
	return selected(scope.Namespace, identity.Namespace) && selected(scope.App, identity.App) && selected(scope.Owner, identity.Owner)
}

type resolvedInput struct {
	input                  analysis.QueryInput
	binding                *InputBinding
	objects                []environment.Object
	identityMapped         bool
	explicitSourceEvidence bool
}

func (p *Prepared) resolveInputs(inputs []analysis.QueryInput, scope environment.CaptureScope, bindings []InputBinding) (map[string]resolvedInput, error) {
	if err := validateBindingSet(inputs, bindings); err != nil {
		return nil, err
	}
	byID := map[string]int{}
	for i, b := range bindings {
		byID[b.InputID] = i
	}
	resolved := make(map[string]resolvedInput, len(inputs))
	for _, input := range inputs {
		value := resolvedInput{input: input, objects: []environment.Object{}}
		intended, mapped := explicitIdentity(input)
		value.identityMapped = mapped
		if index, hasBinding := byID[input.ID]; hasBinding {
			b := bindings[index]
			path := fmt.Sprintf("/input_bindings/%d", index)
			if input.Kind == "unresolved_source" {
				return nil, requestErrorAt("binding_invalid", path+"/input_id", "unresolved source identity cannot be proved consistent with a binding")
			}
			if input.Kind == "explicit_dataset" {
				if !mapped || !consistentExplicit(input, intended, b.Expected) || !identityInScope(scope, b.Expected) {
					return nil, requestErrorAt("binding_invalid", path+"/expected", "binding is not consistent with the exact query identity and scope")
				}
			}
			if object, found := p.env.Object(b.ObjectID); found {
				if objectIdentity(object) != b.Expected {
					return nil, requestErrorAt("binding_invalid", path+"/expected", "captured object disagrees with expected identity")
				}
				value.objects = append(value.objects, object)
			}
			if b.SchemaID != "" {
				linked := false
				for _, schemaBinding := range p.env.Bindings(b.ObjectID) {
					if schemaBinding.SchemaID == b.SchemaID && schemaBinding.Expected == b.Expected {
						linked = true
						break
					}
				}
				if !linked {
					return nil, requestErrorAt("binding_invalid", path+"/schema_id", "selected schema requires a matching bundle entry and object identity binding")
				}
			}
			value.binding = &b
			value.identityMapped = true
			value.explicitSourceEvidence = input.Kind == "named_placeholder" || input.Kind == "explicit_dataset"
		} else if input.Kind == "explicit_dataset" && mapped {
			for _, object := range p.snapshot.Objects {
				identity := objectIdentity(object)
				if consistentExplicit(input, intended, identity) && identityInScope(scope, identity) {
					value.objects = append(value.objects, detach(object))
				}
			}
			value.explicitSourceEvidence = true
		}
		// A classic implicit stream does not itself prove complete source selection.
		resolved[input.ID] = value
	}
	return resolved, nil
}
func consistentExplicit(input analysis.QueryInput, intended, expected environment.ObjectIdentity) bool {
	if input.Identity.Form == "identifier" || input.Identity.Form == "dotted" {
		return expected.Kind == "dataset" && expected.Name == intended.Name
	}
	return intended == expected
}

// Descriptor mapping is deliberately bounded to the entire typed object identity.
// Extra properties carry semantics and cannot be ignored to manufacture a match.
func explicitIdentity(input analysis.QueryInput) (environment.ObjectIdentity, bool) {
	if input.Kind != "explicit_dataset" {
		return environment.ObjectIdentity{}, false
	}
	if input.Identity.Form == "identifier" || input.Identity.Form == "dotted" {
		return environment.ObjectIdentity{Kind: "dataset", Name: input.Identity.Value}, true
	}
	if input.Identity.Form != "descriptor" {
		return environment.ObjectIdentity{}, false
	}
	var descriptor map[string]json.RawMessage
	if json.Unmarshal([]byte(input.Identity.Value), &descriptor) != nil || len(descriptor) != 2 {
		return environment.ObjectIdentity{}, false
	}
	var identity environment.ObjectIdentity
	if json.Unmarshal(descriptor["kind"], &identity.Kind) != nil {
		return environment.ObjectIdentity{}, false
	}
	var properties map[string]json.RawMessage
	if json.Unmarshal(descriptor["properties"], &properties) != nil || properties == nil {
		return environment.ObjectIdentity{}, false
	}
	for name, raw := range properties {
		var destination *string
		switch name {
		case "name":
			destination = &identity.Name
		case "namespace":
			destination = &identity.Namespace
		case "app":
			destination = &identity.App
		case "owner":
			destination = &identity.Owner
		default:
			return environment.ObjectIdentity{}, false
		}
		if bytes.Equal(raw, []byte("null")) || json.Unmarshal(raw, destination) != nil {
			return environment.ObjectIdentity{}, false
		}
	}
	if validateExpected(identity, "") != nil {
		return environment.ObjectIdentity{}, false
	}
	return identity, true
}
