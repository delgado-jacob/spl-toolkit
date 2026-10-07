package compatibility

import (
	"fmt"
	"reflect"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/environment"
)

// ValidateResolutionBindings admits original-role selection keys before variants
// exist. Choices are flattened alternatives, never combinations or candidate IDs.
func (p *Prepared) ValidateResolutionBindings(original analysis.ResolutionEvidence, values []analysis.ResolutionChoice, assessment ResolutionAssessment) error {
	if p == nil || p.env == nil {
		return requestErrorAt("request_invalid", "/snapshot", "prepared environment is required")
	}
	if !validUTF8(reflect.ValueOf(assessment)) {
		return requestErrorAt("request_invalid", "", "request contains invalid UTF-8")
	}
	if assessment.InputBindings == nil {
		return requestErrorAt("request_invalid", "/input_bindings", "input_bindings must be an array")
	}
	for _, selector := range []struct {
		name  string
		value environment.Selector
	}{{"namespace", assessment.QueryScope.Namespace}, {"app", assessment.QueryScope.App}, {"owner", assessment.QueryScope.Owner}} {
		if _, err := normalizeSelector(selector.value, "/query_scope/"+selector.name); err != nil {
			return err
		}
	}
	known := map[string]analysis.QueryInput{}
	for _, input := range original.Analysis.Inputs {
		known[input.ID] = input
	}
	selectedRoles := map[string]map[string]string{}
	for _, group := range original.Placeholders {
		for _, choice := range values {
			if choice.Placeholder != group.Placeholder || choice.Kind != group.Kind {
				continue
			}
			for _, id := range group.OriginalInputIDs {
				if selectedRoles[id] == nil {
					selectedRoles[id] = map[string]string{}
				}
				selectedRoles[id][choice.Value] = choice.Kind
			}
		}
	}
	type key struct {
		id      string
		present bool
		value   string
	}
	seen := map[key]bool{}
	objects := map[string]environment.ObjectIdentity{}
	exhaustive := original.Coverage.State == "complete" || original.Coverage.State == "not_applicable"
	for i, binding := range assessment.InputBindings {
		path := fmt.Sprintf("/input_bindings/%d", i)
		input, exists := known[binding.OriginalInputID]
		if !nonblank(binding.OriginalInputID) || (!exists && exhaustive) {
			return requestErrorAt("binding_invalid", path+"/original_input_id", "binding requires a discovered original input id")
		}
		k := key{id: binding.OriginalInputID, present: binding.ResolvedValue != nil}
		if binding.ResolvedValue != nil {
			k.value = *binding.ResolvedValue
			if !nonblank(k.value) {
				return requestErrorAt("binding_invalid", path+"/resolved_value", "resolved value must be nonblank")
			}
		}
		if seen[k] {
			return requestErrorAt("binding_invalid", path, "duplicate original role and resolved value")
		}
		seen[k] = true
		if !nonblank(binding.ObjectID) {
			return requestErrorAt("binding_invalid", path+"/object_id", "object id must be nonblank")
		}
		if err := validateExpected(binding.Expected, path+"/expected"); err != nil {
			return err
		}
		if prior, found := objects[binding.ObjectID]; found && prior != binding.Expected {
			return requestErrorAt("binding_invalid", path+"/object_id", "object id has contradictory expected identities")
		}
		objects[binding.ObjectID] = binding.Expected
		if exists {
			if choices := selectedRoles[input.ID]; len(choices) > 0 {
				if binding.ResolvedValue == nil {
					return requestErrorAt("binding_invalid", path+"/resolved_value", "substituted original role requires a selected value")
				}
				kind, chosen := choices[*binding.ResolvedValue]
				if !chosen {
					return requestErrorAt("binding_invalid", path+"/resolved_value", "value is outside this original role's choices")
				}
				if binding.Expected.Kind != kind || binding.Expected.Name != *binding.ResolvedValue || !identityInScope(assessment.QueryScope, binding.Expected) {
					return requestErrorAt("binding_invalid", path+"/expected", "binding disagrees with the selected identity and scope")
				}
			} else {
				if binding.ResolvedValue != nil {
					return requestErrorAt("binding_invalid", path+"/resolved_value", "unchanged original role must omit resolved value")
				}
				intended, mapped := explicitIdentity(input)
				if input.Kind == "unresolved_source" || (input.Kind == "explicit_dataset" && (!mapped || !consistentExplicit(input, intended, binding.Expected) || !identityInScope(assessment.QueryScope, binding.Expected))) {
					return requestErrorAt("binding_invalid", path+"/expected", "binding is not consistent with the exact query identity and scope")
				}
			}
		}
		if err := p.validateSelection(binding.ObjectID, binding.Expected, binding.SchemaID, path); err != nil {
			return err
		}
	}
	return nil
}
