package validation

import (
	"encoding/json"
	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"sort"
	"strconv"
	"strings"
)

const jsonSchemaDialect = "https://json-schema.org/draft/2020-12/schema"
const jsonSchemaRoot = "urn:spl-toolkit:json-schema:root"
const schemaProjectionBudget = 4096
const schemaPathSegmentBudget = 128
const schemaEnumerationBudget = 4096

// Capability reasons are stable identifiers shared with target metadata and docs.
var schemaCapabilityReasons = map[string]string{
	"object_shape_unknown":        "Requiredness depends on an unproved object shape.",
	"unrepresentable_source_name": "An admitted blank source name cannot be represented by the canonical source universe.",
	"array_traversal":             "Array traversal is outside field projection.", "partial_name_universe": "The source name universe is not exhaustive.", "enumeration_budget": "Candidate enumeration reached its work or name bound.", "traversal_budget": "The requested path exceeded the projection state or segment bound.", "recursive_ref": "A reference cycle made no path progress.", "unresolved_ref": "The reference has no supplied local schema target.", "unsupported_pattern": "The pattern is outside the supported ASCII subset.", "pattern_properties": "Patterns can admit unenumerated names.", "dynamic_ref": "Dynamic reference evaluation is unsupported.", "unevaluated_properties": "Unevaluated property tracking is unsupported.", "negation": "Negated field constraints are unsupported.", "conditional_schema": "Conditional field constraints are unsupported.", "dependent_schema": "Dependent schemas are unsupported.", "dependent_required": "Dependent requiredness is unsupported.", "property_names": "Property name constraints are unsupported.", "object_value_constraint": "Object value and cardinality constraints are unsupported.", "required_vocabulary": "A required custom vocabulary is unsupported.", "exclusive_branches": "Exclusive branch viability is unresolved.", "alternative_branches": "Alternative branches disagree on membership.", "literal_path_collision": "Literal dotted and nested names collide.",
}

func capabilityForKeyword(k string) string {
	switch k {
	case "unevaluatedProperties":
		return "unevaluated_properties"
	case "not":
		return "negation"
	case "dependentSchemas":
		return "dependent_schema"
	case "dependentRequired":
		return "dependent_required"
	case "propertyNames":
		return "property_names"
	default:
		return "object_value_constraint"
	}
}

type jsonSchemaTarget struct {
	root            *schemaNode
	index           *schemaIndex
	fields          []string
	complete        bool
	metadata        SchemaTargetInfo
	enumerationWork int
}

func prepareJSONSchema(t SchemaTarget) (*jsonSchemaTarget, error) {
	t, e := normalizeSchemaTarget(t)
	if e != nil {
		return nil, e
	}
	if t.Kind != "json_schema" {
		return nil, inputError("expected json_schema target")
	}
	idx, root, e := newSchemaIndex(t)
	if e != nil {
		return nil, e
	}
	p := &jsonSchemaTarget{root: root, index: idx}
	p.enumerate()
	idx.limitations["array_traversal"] = true
	if !p.complete {
		idx.limitations["partial_name_universe"] = true
	}
	p.metadata = SchemaTargetInfo{Kind: "json_schema", Identity: t.Identity, Dialect: jsonSchemaDialect, BaseURI: root.base, ResourceURIs: sortedKeys(idx.resources), Members: []SchemaClass{}, Extensions: []SchemaExtension{}, Limitations: sortedKeys(idx.limitations)}
	return p, nil
}
func newProjection(a analysis.SourceFieldAdmission, out string, ev ...SchemaEvidence) fieldProjection {
	return fieldProjection{Admission: a, Outcome: out, Evidence: append([]SchemaEvidence{}, ev...), SupportingClasses: []SchemaClass{}, MissingClasses: []SchemaClass{}, IndeterminateClasses: []SchemaClass{}}
}
func schemaEv(n *schemaNode, key, basis, requirement, reason string) SchemaEvidence {
	return SchemaEvidence{ResourceURI: n.uri, Pointer: n.resourcePointer, Keyword: key, DeclarationBasis: basis, Requirement: requirement, Reason: reason}
}
func unknownProjection(n *schemaNode, key, reason string) fieldProjection {
	return newProjection(analysis.SourceFieldIndeterminate, "indeterminate", schemaEv(n, key, "unknown", "unknown", reason))
}
func intersectProjection(a, b fieldProjection) fieldProjection {
	a.Evidence = append(a.Evidence, b.Evidence...)
	if a.Admission == analysis.SourceFieldProhibited || b.Admission == analysis.SourceFieldProhibited {
		a.Admission = analysis.SourceFieldProhibited
		a.Outcome = "missing"
		return a
	}
	if a.Admission == analysis.SourceFieldIndeterminate || b.Admission == analysis.SourceFieldIndeterminate {
		a.Admission = analysis.SourceFieldIndeterminate
		if a.Outcome != "conditional" && b.Outcome != "conditional" {
			a.Outcome = "indeterminate"
		} else {
			a.Outcome = "conditional"
		}
		return a
	}
	if a.Outcome == "indeterminate" || b.Outcome == "indeterminate" {
		a.Outcome = "indeterminate"
	} else if b.Outcome == "required" || a.Outcome == "required" {
		a.Outcome = "required"
	} else if a.Outcome == "optional" || b.Outcome == "optional" {
		a.Outcome = "optional"
	}
	return a
}

type projectionState struct {
	node *schemaNode
	path string
}
type projectionContext struct {
	active          map[projectionState]bool
	seen            map[projectionState]bool
	memo            map[projectionState]fieldProjection
	requirementMemo map[projectionState][]schemaRequirementFact
}

func (p *jsonSchemaTarget) project(name string) fieldProjection {
	if !validSchemaName(name) {
		return unknownProjection(p.root, "", "invalid_field_name")
	}
	parts := strings.Split(name, ".")
	if len(parts) > schemaPathSegmentBudget {
		return unknownProjection(p.root, "", "traversal_budget")
	}
	ctx := &projectionContext{active: map[projectionState]bool{}, seen: map[projectionState]bool{}}
	result := p.walk(p.root, parts, ctx)
	if result.Admission == analysis.SourceFieldAdmitted {
		facts := p.requirementFacts(p.root, parts, ctx, map[projectionState]bool{})
		allRequired, allObjects := true, true
		for _, fact := range facts {
			allRequired = allRequired && fact.required
			allObjects = allObjects && fact.object
		}
		if allRequired {
			if allObjects {
				result.Outcome = "required"
			} else {
				result.Outcome = "indeterminate"
				result.Evidence = append(result.Evidence, schemaEv(p.root, "type", "unknown", "unknown", "object_shape_unknown"))
			}
		}
	}
	sort.SliceStable(result.Evidence, func(i, j int) bool {
		a, _ := json.Marshal(result.Evidence[i])
		b, _ := json.Marshal(result.Evidence[j])
		return string(a) < string(b)
	})
	return result
}
func (p *jsonSchemaTarget) walk(n *schemaNode, path []string, ctx *projectionContext) (result fieldProjection) {
	state := projectionState{n, schemaPathKey(path)}
	if value, ok := ctx.memo[state]; ok {
		value.Evidence = append([]SchemaEvidence{}, value.Evidence...)
		return value
	}
	if ctx.active[state] {
		return unknownProjection(n, "$ref", "recursive_ref")
	}
	if !ctx.seen[state] {
		if len(ctx.seen) >= schemaProjectionBudget {
			return unknownProjection(n, "", "traversal_budget")
		}
		ctx.seen[state] = true
	}
	ctx.active[state] = true
	defer func() {
		delete(ctx.active, state)
		result.Evidence = uniqueSchemaEvidence(result.Evidence)
		if ctx.memo == nil {
			ctx.memo = map[projectionState]fieldProjection{}
		}
		memo := result
		memo.Evidence = append([]SchemaEvidence{}, result.Evidence...)
		ctx.memo[state] = memo
	}()
	if b, ok := n.value.(bool); ok {
		if !b {
			return newProjection(analysis.SourceFieldProhibited, "missing", schemaEv(n, "", "unknown", "unknown", ""))
		}
		return newProjection(analysis.SourceFieldAdmitted, "permitted_unspecified", schemaEv(n, "", "generic_object", "optional", ""))
	}
	result, declared := p.local(n, path, ctx)
	if _, ok := n.object["$ref"]; ok {
		if target := n.refs["$ref"]; target != nil {
			result = intersectProjection(result, p.walk(target, path, ctx))
		} else {
			result = intersectProjection(result, unknownProjection(n, "$ref", "unresolved_ref"))
		}
	}
	for _, operator := range []string{"allOf", "anyOf", "oneOf"} {
		branches, ok := n.object[operator].([]any)
		if !ok {
			continue
		}
		values := make([]fieldProjection, 0, min(len(branches), schemaProjectionBudget))
		for i := range branches {
			child := n.children[operator+"/"+strconv.Itoa(i)]
			if len(ctx.seen) >= schemaProjectionBudget && !ctx.seen[projectionState{child, schemaPathKey(path)}] {
				values = append(values, unknownProjection(child, operator, "traversal_budget"))
				break
			}
			v := p.walk(child, path, ctx)
			v.Evidence = append(v.Evidence, SchemaEvidence{ResourceURI: child.uri, Pointer: child.resourcePointer, Operator: operator, Branch: child.resourcePointer})
			values = append(values, v)
		}
		if operator == "allOf" {
			for _, v := range values {
				result = intersectProjection(result, v)
			}
			continue
		}
		combined := values[0]
		for _, v := range values[1:] {
			combined.Evidence = append(combined.Evidence, v.Evidence...)
		}
		admitted, prohibited := 0, 0
		for _, v := range values {
			if v.Admission == analysis.SourceFieldAdmitted {
				admitted++
			}
			if v.Admission == analysis.SourceFieldProhibited {
				prohibited++
			}
		}
		if prohibited == len(values) {
			combined.Admission = analysis.SourceFieldProhibited
			combined.Outcome = "missing"
		} else if admitted == len(values) {
			combined.Admission = analysis.SourceFieldAdmitted
			combined.Outcome = values[0].Outcome
			for _, v := range values {
				if schemaOutcomeRank(v.Outcome) > schemaOutcomeRank(combined.Outcome) {
					combined.Outcome = v.Outcome
				}
			}
			if operator == "oneOf" && len(values) > 1 && !declared {
				combined.Admission = analysis.SourceFieldIndeterminate
				combined.Outcome = "indeterminate"
				combined.Evidence = append(combined.Evidence, schemaEv(n, operator, "unknown", "unknown", "exclusive_branches"))
			}
		} else {
			combined.Admission = analysis.SourceFieldIndeterminate
			combined.Outcome = "conditional"
			if admitted+prohibited != len(values) {
				combined.Outcome = "indeterminate"
			}
			combined.Evidence = append(combined.Evidence, schemaEv(n, operator, "unknown", "unknown", "alternative_branches"))
		}
		result = intersectProjection(result, combined)
	}
	for _, key := range []string{"$dynamicRef", "not", "dependentSchemas", "propertyNames", "minProperties", "maxProperties", "const", "enum"} {
		if v, ok := n.object[key]; ok && fieldConstraintRelevant(key, v, path, n) {
			reason := capabilityForKeyword(key)
			if key == "$dynamicRef" {
				reason = "dynamic_ref"
			}
			result = intersectProjection(result, unknownProjection(n, key, reason))
		}
	}
	if _, ok := n.object["if"]; ok {
		if _, a := n.object["then"]; a {
			result = intersectProjection(result, unknownProjection(n, "if", "conditional_schema"))
		} else if _, b := n.object["else"]; b {
			result = intersectProjection(result, unknownProjection(n, "if", "conditional_schema"))
		}
	}
	if _, ok := n.object["unevaluatedProperties"]; ok && !declared && len(path) > 0 {
		result = intersectProjection(result, unknownProjection(n, "unevaluatedProperties", "unevaluated_properties"))
	}
	if deps, ok := n.object["dependentRequired"].(map[string]any); ok && len(path) > 0 {
		affected := false
		for _, items := range deps {
			for _, value := range items.([]any) {
				if value == path[0] {
					affected = true
				}
			}
		}
		if affected {
			result.Evidence = append(result.Evidence, schemaEv(n, "dependentRequired", "unknown", "unknown", "dependent_required"))
			if result.Admission == analysis.SourceFieldAdmitted && result.Outcome != "required" {
				result.Outcome = "indeterminate"
			}
		}
	}
	if vocab, ok := n.object["$vocabulary"].(map[string]any); ok {
		for _, uri := range sortedKeys(vocab) {
			if vocab[uri] == true && !knownSchemaVocabulary(uri) {
				result = intersectProjection(result, unknownProjection(n, "$vocabulary", "required_vocabulary"))
			}
		}
	}
	return result
}
func fieldConstraintRelevant(key string, v any, path []string, n *schemaNode) bool {
	if key == "const" {
		_, ok := v.(map[string]any)
		return ok
	}
	if key == "enum" {
		for _, v := range v.([]any) {
			if _, ok := v.(map[string]any); ok {
				return true
			}
		}
		return false
	}
	if len(path) == 0 && (key == "minProperties" || key == "maxProperties" || key == "propertyNames" || key == "dependentSchemas") {
		return false
	}
	return true
}
func typeShape(n *schemaNode) (object, array, known bool) {
	v, ok := n.object["type"]
	if !ok {
		return false, false, false
	}
	known = true
	a, ok := v.([]any)
	if !ok {
		a = []any{v}
	}
	for _, v := range a {
		object = object || v == "object"
		array = array || v == "array"
	}
	return
}
func (p *jsonSchemaTarget) local(n *schemaNode, path []string, ctx *projectionContext) (fieldProjection, bool) {
	return p.localPath(n, path, ctx, false)
}
func (p *jsonSchemaTarget) localPath(n *schemaNode, path []string, ctx *projectionContext, ignoreLiteral bool) (fieldProjection, bool) {
	if len(path) == 0 {
		return newProjection(analysis.SourceFieldAdmitted, "permitted_unspecified"), false
	}
	obj, arr, known := typeShape(n)
	if arr {
		return unknownProjection(n, "type", "array_traversal"), false
	}
	objectCertain := schemaObjectOnly(n)
	if known && !obj {
		if arr {
			return unknownProjection(n, "type", "array_traversal"), false
		}
		return newProjection(analysis.SourceFieldProhibited, "missing", schemaEv(n, "type", "unknown", "unknown", "")), false
	}
	name := path[0]
	props, _ := n.object["properties"].(map[string]any)
	if len(path) > 1 && !ignoreLiteral {
		nested, nestedDeclared := p.localPath(n, path, ctx, true)
		for consumed := 2; consumed <= len(path); consumed++ {
			literalName := strings.Join(path[:consumed], ".")
			if _, literal := props[literalName]; literal {
				literalPath := append([]string{literalName}, path[consumed:]...)
				value, _ := p.localPath(n, literalPath, ctx, true)
				if nested.Admission == analysis.SourceFieldProhibited {
					return value, true
				}
				if value.Admission == analysis.SourceFieldProhibited {
					continue
				}
				uncertainty := unknownProjection(n, "properties", "literal_path_collision")
				uncertainty.Evidence = append(uncertainty.Evidence, nested.Evidence...)
				uncertainty.Evidence = append(uncertainty.Evidence, value.Evidence...)
				return uncertainty, false
			}
		}
		return nested, nestedDeclared
	}
	req := "optional"
	if names, ok := n.object["required"].([]any); ok {
		for _, v := range names {
			if v == name {
				req = "required"
			}
		}
	}
	result := newProjection(analysis.SourceFieldAdmitted, "permitted_unspecified")
	matched, uncertain := false, false
	apply := func(child *schemaNode, key, basis string) {
		v := p.walk(child, path[1:], ctx)
		localReq := req
		if req == "required" && !objectCertain {
			localReq = "unknown"
		}
		v.Evidence = append([]SchemaEvidence{schemaEv(n, key, basis, localReq, "")}, v.Evidence...)
		if v.Admission == analysis.SourceFieldAdmitted {
			if v.Outcome == "indeterminate" {
			} else if req == "required" && objectCertain && (len(path) == 1 || v.Outcome == "required") {
				v.Outcome = "required"
			} else {
				v.Outcome = "optional"
			}
		}
		result = intersectProjection(result, v)
	}
	if _, ok := props[name]; ok {
		matched = true
		apply(n.children["properties/"+pointerEscape(name)], "properties", "property")
	}
	for _, pattern := range sortedKeys(n.patterns) {
		matches, conclusive := n.patterns[pattern].match(name)
		if !conclusive {
			uncertain = true
			result.Evidence = append(result.Evidence, schemaEv(n, "patternProperties", "unknown", "unknown", "unsupported_pattern"))
			continue
		}
		if matches {
			matched = true
			apply(n.children["patternProperties/"+pointerEscape(pattern)], "patternProperties", "pattern")
		}
	}
	if uncertain {
		result = intersectProjection(result, unknownProjection(n, "patternProperties", "unsupported_pattern"))
	}
	if !matched && !uncertain {
		if child := n.children["additionalProperties"]; child != nil {
			apply(child, "additionalProperties", "additional_properties")
			if result.Admission == analysis.SourceFieldAdmitted && result.Outcome != "required" && result.Outcome != "indeterminate" {
				result.Outcome = "permitted_unspecified"
			}
		} else {
			result.Evidence = append(result.Evidence, schemaEv(n, "", "generic_object", req, ""))
			if req == "required" && objectCertain && len(path) == 1 {
				result.Outcome = "required"
			}
		}
	}
	return result, matched
}
func (p *jsonSchemaTarget) universe() analysis.SourceUniverse {
	return analysis.SourceUniverse{Fields: append([]string{}, p.fields...), Complete: p.complete, Resolve: func(name string) analysis.SourceFieldAdmission { return p.project(name).Admission }}
}
func (p *jsonSchemaTarget) info() SchemaTargetInfo {
	info := p.metadata
	info.ResourceURIs = append([]string{}, info.ResourceURIs...)
	info.Limitations = append([]string{}, info.Limitations...)
	info.Members = []SchemaClass{}
	info.Extensions = []SchemaExtension{}
	return info
}

// Enumeration is an independent, deterministic bounded DFS. Every traversed
// schema/ref/property edge and candidate emission consumes a work unit first.
// Active locations detect cycles only on the current ancestry, never across
// different path prefixes of a shared reference DAG.
func (p *jsonSchemaTarget) enumerate() {
	names := map[string]bool{}
	active := map[*schemaNode]bool{}
	truncated := false
	charge := func() bool {
		if p.enumerationWork >= schemaEnumerationBudget {
			truncated = true
			return false
		}
		p.enumerationWork++
		return true
	}
	var visit func(*schemaNode, string)
	visit = func(n *schemaNode, prefix string) {
		if truncated {
			return
		}
		if active[n] {
			return
		}
		active[n] = true
		defer delete(active, n)
		props, _ := n.object["properties"].(map[string]any)
		for _, name := range sortedKeys(props) {
			full := prefix + name
			if !names[full] {
				if len(names) >= schemaEnumerationBudget {
					truncated = true
					return
				}
				if !charge() {
					return
				}
				names[full] = true
			}
			if !charge() {
				return
			}
			visit(n.children["properties/"+pointerEscape(name)], full+".")
			if truncated {
				return
			}
		}
		if ref := n.refs["$ref"]; ref != nil {
			if !charge() {
				return
			}
			visit(ref, prefix)
		}
		for _, key := range []string{"allOf", "anyOf", "oneOf"} {
			if a, ok := n.object[key].([]any); ok {
				for i := range a {
					if !charge() {
						return
					}
					visit(n.children[key+"/"+strconv.Itoa(i)], prefix)
					if truncated {
						return
					}
				}
			}
		}
	}
	visit(p.root, "")
	p.fields = []string{}
	unrepresentable := false
	for _, name := range sortedKeys(names) {
		if !validSchemaName(name) {
			ctx := &projectionContext{active: map[projectionState]bool{}, seen: map[projectionState]bool{}}
			if p.walk(p.root, []string{name}, ctx).Admission != analysis.SourceFieldProhibited {
				unrepresentable = true
			}
			continue
		}
		if p.project(name).Admission != analysis.SourceFieldProhibited {
			p.fields = append(p.fields, name)
		}
	}
	if unrepresentable {
		p.index.limitations["unrepresentable_source_name"] = true
	}
	p.complete = !truncated && !unrepresentable && p.finite(p.root, map[*schemaNode]bool{})
	if truncated {
		p.index.limitations["enumeration_budget"] = true
	}
}
func (p *jsonSchemaTarget) finite(n *schemaNode, active map[*schemaNode]bool) bool {
	if active[n] {
		return false
	}
	active[n] = true
	defer delete(active, n)
	if b, ok := n.value.(bool); ok {
		return !b
	}
	obj, arr, known := typeShape(n)
	if known && !obj && !arr {
		return true
	}
	for _, k := range []string{"$dynamicRef", "not", "if", "dependentSchemas", "unevaluatedProperties", "propertyNames", "const", "enum", "minProperties", "maxProperties", "$vocabulary"} {
		if _, ok := n.object[k]; ok {
			return false
		}
	}
	if len(n.patterns) > 0 {
		return false
	}
	if a, ok := n.object["allOf"].([]any); ok {
		for i := range a {
			if p.finite(n.children["allOf/"+strconv.Itoa(i)], active) {
				return true
			}
		}
	}
	if _, ok := n.object["$ref"]; ok {
		if ref := n.refs["$ref"]; ref != nil && p.finite(ref, active) {
			return true
		}
	}
	for _, key := range []string{"anyOf", "oneOf"} {
		if a, ok := n.object[key].([]any); ok {
			for i := range a {
				if !p.finite(n.children[key+"/"+strconv.Itoa(i)], active) {
					return false
				}
			}
			return true
		}
	}
	extra := n.children["additionalProperties"]
	if extra == nil || extra.value != false || arr {
		return false
	}
	props, _ := n.object["properties"].(map[string]any)
	for _, key := range sortedKeys(props) {
		if !p.finite(n.children["properties/"+pointerEscape(key)], active) {
			return false
		}
	}
	return true
}

var _ preparedSchemaTarget = (*jsonSchemaTarget)(nil)

// Requirement facts combine per path level rather than flattening declarations.
// Conjunction can supply required and object facts from different sibling schemas;
// alternatives retain only common facts. This is presence projection, not a solver.
type schemaRequirementFact struct{ required, object bool }

func schemaObjectOnly(n *schemaNode) bool {
	v := n.object["type"]
	if v == "object" {
		return true
	}
	if a, ok := v.([]any); ok {
		return len(a) == 1 && a[0] == "object"
	}
	return false
}
func (p *jsonSchemaTarget) requirementFacts(n *schemaNode, path []string, ctx *projectionContext, active map[projectionState]bool) (facts []schemaRequirementFact) {
	facts = make([]schemaRequirementFact, len(path))
	if len(path) == 0 {
		return facts
	}
	state := projectionState{n, schemaPathKey(path)}
	if value, ok := ctx.requirementMemo[state]; ok {
		return append([]schemaRequirementFact{}, value...)
	}
	if active[state] {
		return facts
	}
	if !ctx.seen[state] {
		if len(ctx.seen) >= schemaProjectionBudget {
			return facts
		}
		ctx.seen[state] = true
	}
	active[state] = true
	defer func() {
		delete(active, state)
		if ctx.requirementMemo == nil {
			ctx.requirementMemo = map[projectionState][]schemaRequirementFact{}
		}
		ctx.requirementMemo[state] = append([]schemaRequirementFact{}, facts...)
	}()
	facts[0].object = schemaObjectOnly(n)
	if req, ok := n.object["required"].([]any); ok {
		for _, v := range req {
			facts[0].required = facts[0].required || v == path[0]
		}
	}
	merge := func(other []schemaRequirementFact) {
		for i, v := range other {
			facts[i].required = facts[i].required || v.required
			facts[i].object = facts[i].object || v.object
		}
	}
	if child := n.children["properties/"+pointerEscape(path[0])]; child != nil {
		copy(facts[1:], p.requirementFacts(child, path[1:], ctx, active))
	}
	matchedOrUnknown := n.children["properties/"+pointerEscape(path[0])] != nil
	for _, pattern := range sortedKeys(n.patterns) {
		match, known := n.patterns[pattern].match(path[0])
		matchedOrUnknown = matchedOrUnknown || match || !known
	}
	if !matchedOrUnknown {
		if child := n.children["additionalProperties"]; child != nil {
			copy(facts[1:], p.requirementFacts(child, path[1:], ctx, active))
		}
	}
	for _, pattern := range sortedKeys(n.patterns) {
		if match, known := n.patterns[pattern].match(path[0]); known && match {
			child := p.requirementFacts(n.children["patternProperties/"+pointerEscape(pattern)], path[1:], ctx, active)
			for i, v := range child {
				facts[i+1].required = facts[i+1].required || v.required
				facts[i+1].object = facts[i+1].object || v.object
			}
		}
	}
	if ref := n.refs["$ref"]; ref != nil {
		merge(p.requirementFacts(ref, path, ctx, active))
	}
	for _, op := range []string{"allOf", "anyOf", "oneOf"} {
		if branches, ok := n.object[op].([]any); ok {
			var common []schemaRequirementFact
			for i := range branches {
				v := p.requirementFacts(n.children[op+"/"+strconv.Itoa(i)], path, ctx, active)
				if op == "allOf" {
					merge(v)
				} else if i == 0 {
					common = v
				} else {
					for j := range common {
						common[j].required = common[j].required && v[j].required
						common[j].object = common[j].object && v[j].object
					}
				}
			}
			if op != "allOf" {
				merge(common)
			}
		}
	}
	return facts
}

func schemaOutcomeRank(outcome string) int {
	switch outcome {
	case "required":
		return 0
	case "optional":
		return 1
	case "permitted_unspecified":
		return 2
	default:
		return 3
	}
}
func uniqueSchemaEvidence(evidence []SchemaEvidence) []SchemaEvidence {
	seen := map[SchemaEvidence]bool{}
	result := make([]SchemaEvidence, 0, len(evidence))
	for _, e := range evidence {
		if !seen[e] {
			seen[e] = true
			result = append(result, e)
		}
	}
	return result
}

func schemaPathKey(path []string) string {
	var key strings.Builder
	for _, part := range path {
		key.WriteString(strconv.Itoa(len(part)))
		key.WriteByte(':')
		key.WriteString(part)
	}
	return key.String()
}
