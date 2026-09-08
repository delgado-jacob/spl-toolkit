package validation

import (
	"encoding/json"
	"strings"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
)

func (p *ocsfTarget) profileAdmission(a ocsfAttribute, class string) (analysis.SourceFieldAdmission, bool, string) {
	if len(a.Profiles) == 0 {
		return analysis.SourceFieldAdmitted, false, ""
	}
	enabled := 0
	for _, profile := range a.Profiles {
		if containsOCSF(p.selection.Profiles, profile) && containsOCSF(p.catalog.Classes[class].Profiles, profile) {
			enabled++
		}
	}
	if enabled == 0 {
		if p.unknownProfiles[class] {
			return analysis.SourceFieldIndeterminate, false, "ocsf_profile_inheritance"
		}
		return analysis.SourceFieldProhibited, false, ""
	}
	return analysis.SourceFieldAdmitted, enabled < len(a.Profiles) && a.Requirement == "required", ""
}
func ocsfPointerName(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, "~", "~0"), "/", "~1")
}
func (p *ocsfTarget) evidence(m SchemaClass, pointer, keyword, basis, requirement, reason string) SchemaEvidence {
	uid := m.UID
	return SchemaEvidence{ResourceURI: "urn:ocsf:" + p.catalog.Version, Pointer: pointer, Keyword: keyword, ClassKey: m.Key, ClassUID: &uid, DeclarationBasis: basis, Requirement: requirement, Reason: reason}
}
func (p *ocsfTarget) project(name string) fieldProjection {
	result := newProjection(analysis.SourceFieldProhibited, "missing")
	allRequired := true
	anyUnspecified := false
	unknownRequirement := false
	budget := schemaProjectionBudget
	for _, m := range p.members {
		got := p.projectClass(name, m, &budget)
		result.Evidence = append(result.Evidence, got.Evidence...)
		switch got.Admission {
		case analysis.SourceFieldAdmitted:
			result.SupportingClasses = append(result.SupportingClasses, m)
			allRequired = allRequired && got.Outcome == "required"
			anyUnspecified = anyUnspecified || got.Outcome == "permitted_unspecified"
			unknownRequirement = unknownRequirement || got.Outcome == "indeterminate"
		case analysis.SourceFieldProhibited:
			result.MissingClasses = append(result.MissingClasses, m)
		default:
			result.IndeterminateClasses = append(result.IndeterminateClasses, m)
		}
	}
	if len(result.IndeterminateClasses) > 0 {
		result.Admission = analysis.SourceFieldIndeterminate
		result.Outcome = "indeterminate"
	} else if len(result.SupportingClasses) == len(p.members) {
		result.Admission = analysis.SourceFieldAdmitted
		switch {
		case unknownRequirement:
			result.Outcome = "indeterminate"
		case allRequired:
			result.Outcome = "required"
		case anyUnspecified:
			result.Outcome = "permitted_unspecified"
		default:
			result.Outcome = "optional"
		}
	} else if len(result.SupportingClasses) > 0 {
		result.Admission = analysis.SourceFieldIndeterminate
		result.Outcome = "conditional"
	}
	return result
}
func (p *ocsfTarget) projectClass(name string, m SchemaClass, budget *int) fieldProjection {
	node := p.catalog.Classes[m.Key]
	pointer := "/classes/" + ocsfPointerName(m.Key)
	ev := []SchemaEvidence{}
	finish := func(a analysis.SourceFieldAdmission, out string) fieldProjection { return newProjection(a, out, ev...) }
	parts := strings.Split(name, ".")
	if !validSchemaName(name) || len(parts) > schemaPathSegmentBudget {
		ev = append(ev, p.evidence(m, pointer, "attributes", "unknown", "unknown", "traversal_budget"))
		return finish(analysis.SourceFieldIndeterminate, "indeterminate")
	}
	// 0 optional, 1 required, 2 unknown requiredness. Admission is independent.
	requirement := 1
	forcedThrough := 0
	for i := 0; i < len(parts); {
		if *budget <= 0 {
			ev = append(ev, p.evidence(m, pointer, "attributes", "unknown", "unknown", "traversal_budget"))
			return finish(analysis.SourceFieldIndeterminate, "indeterminate")
		}
		*budget--
		part := parts[i]
		advance := 1
		remaining := strings.Join(parts[i:], ".")
		forced, unknown, constraintEv := p.constraintFacts(node, remaining, m, pointer)
		ev = append(ev, constraintEv...)
		if forced > 0 && i+forced > forcedThrough {
			forcedThrough = i + forced
		}
		// A compiled attribute key is an exact declaration, including literal
		// dots. Multiple possible structural segmentations remain indeterminate.
		candidates := []string{}
		for key := range node.Attributes {
			if remaining == key || strings.HasPrefix(remaining, key+".") {
				candidates = append(candidates, key)
			}
		}
		if len(candidates) > 1 {
			ev = append(ev, p.evidence(m, pointer, "attributes", "unknown", "unknown", "literal_path_collision"))
			return finish(analysis.SourceFieldIndeterminate, "indeterminate")
		}
		if len(candidates) == 1 {
			part = candidates[0]
			advance = len(strings.Split(part, "."))
		}
		a, ok := node.Attributes[part]
		if !ok {
			reason := ""
			admission := analysis.SourceFieldProhibited
			out := "missing"
			if p.unknownProfiles[m.Key] {
				reason = "ocsf_profile_inheritance"
				admission = analysis.SourceFieldIndeterminate
				out = "indeterminate"
			}
			if unknown {
				reason = "ocsf_constraint"
				admission = analysis.SourceFieldIndeterminate
				out = "indeterminate"
			}
			ev = append(ev, p.evidence(m, pointer, "attributes", "unknown", "unknown", reason))
			return finish(admission, out)
		}
		attrPointer := pointer + "/attributes/" + ocsfPointerName(part)
		admission, uncertain, reason := p.profileAdmission(a, m.Key)
		local := a.Requirement
		if uncertain {
			local = "unknown"
			reason = "ocsf_profile_requirement"
		}
		ev = append(ev, p.evidence(m, attrPointer, "requirement", "ocsf_attribute", local, reason))
		if len(a.Profiles) > 0 {
			ev = append(ev, p.evidence(m, attrPointer+"/profiles", "profiles", "ocsf_attribute", local, reason))
		}
		if forcedThrough > i {
			local = "required"
		}
		if admission != analysis.SourceFieldAdmitted {
			if admission == analysis.SourceFieldProhibited {
				return finish(admission, "missing")
			}
			return finish(admission, "indeterminate")
		}
		if unknown {
			return finish(analysis.SourceFieldIndeterminate, "indeterminate")
		}
		if local == "optional" || local == "recommended" {
			requirement = 0
		} else if local == "unknown" && requirement != 0 {
			requirement = 2
		}
		if i+advance == len(parts) {
			switch requirement {
			case 1:
				return finish(analysis.SourceFieldAdmitted, "required")
			case 2:
				return finish(analysis.SourceFieldAdmitted, "indeterminate")
			default:
				return finish(analysis.SourceFieldAdmitted, "optional")
			}
		}
		if a.IsArray {
			ev = append(ev, p.evidence(m, attrPointer, "is_array", "unknown", "unknown", "array_traversal"))
			return finish(analysis.SourceFieldIndeterminate, "indeterminate")
		}
		if a.Type == "json_t" || (a.Type == "object_t" && a.ObjectType == "object") {
			ev = append(ev, p.evidence(m, attrPointer, "type", "generic_object", "unknown", ""))
			return finish(analysis.SourceFieldAdmitted, "permitted_unspecified")
		}
		if a.Type != "object_t" {
			ev = append(ev, p.evidence(m, attrPointer, "type", "ocsf_attribute", "unknown", ""))
			return finish(analysis.SourceFieldProhibited, "missing")
		}
		node = p.catalog.Objects[a.ObjectType]
		pointer = "/objects/" + ocsfPointerName(a.ObjectType)
		i += advance
	}
	return finish(analysis.SourceFieldProhibited, "missing")
}

// Constraints contribute local presence evidence; multi-member alternatives do
// not make any single member required. Unknown forms affect only named paths
// when the member list is recognizable, otherwise the scope is unresolved.
func (p *ocsfTarget) constraintFacts(n ocsfNode, path string, m SchemaClass, pointer string) (int, bool, []SchemaEvidence) {
	forced := 0
	unknown := false
	ev := []SchemaEvidence{}
	for _, operator := range sortedKeys(n.Constraints) {
		raw := n.Constraints[operator]
		var names []string
		parsed := json.Unmarshal(raw, &names) == nil && len(names) > 0
		relevant := !parsed
		for _, name := range names {
			if path == name || strings.HasPrefix(path, name+".") || strings.HasPrefix(name, path+".") {
				relevant = true
			}
		}
		if !relevant {
			continue
		}
		supported := (operator == "at_least_one" || operator == "just_one") && parsed
		requirement := "unknown"
		reason := ""
		if !supported {
			unknown = true
			reason = "ocsf_constraint"
		} else if len(names) == 1 && p.constraintPathSupported(n, names[0], m.Key) {
			forced = max(forced, len(strings.Split(names[0], ".")))
			requirement = "required"
		}
		e := p.evidence(m, pointer+"/constraints/"+ocsfPointerName(operator), "constraints", "ocsf_attribute", requirement, reason)
		e.Operator = operator
		ev = append(ev, e)
	}
	return forced, unknown, ev
}
func (p *ocsfTarget) constraintPathSupported(n ocsfNode, path, class string) bool {
	parts := strings.Split(path, ".")
	if len(parts) > schemaPathSegmentBudget {
		return false
	}
	for i, part := range parts {
		a, ok := n.Attributes[part]
		if !ok {
			return false
		}
		admission, uncertain, _ := p.profileAdmission(a, class)
		if admission != analysis.SourceFieldAdmitted || uncertain {
			return false
		}
		if i == len(parts)-1 {
			return true
		}
		if a.IsArray || a.Type != "object_t" || a.ObjectType == "object" {
			return false
		}
		n = p.catalog.Objects[a.ObjectType]
	}
	return false
}
