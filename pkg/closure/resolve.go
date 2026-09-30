package closure

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
)

func textDigest(text string) string {
	sum := sha256.Sum256([]byte(text))
	return "sha256:" + hex.EncodeToString(sum[:])
}
func allMacroCalls(calls []macroInvocation) []macroInvocation {
	out := []macroInvocation{}
	var walk func([]macroInvocation)
	walk = func(calls []macroInvocation) {
		for _, c := range calls {
			out = append(out, c)
			walk(c.Nested)
		}
	}
	walk(calls)
	return out
}

type bindingDocument struct {
	document analysis.QueryDocument
	objectID string
}

func bindingOccurrence(binding Binding, document analysis.QueryDocument, result *analysis.Result, target Definition, owner *Definition) bool {
	if binding.Kind == "macro" {
		text := document.Text
		if owner != nil && owner.Kind == "macro" {
			for _, name := range owner.Arguments {
				placeholder := "$" + name + "$"
				text = strings.ReplaceAll(text, placeholder, strings.Repeat("0", len(placeholder)))
			}
		}
		for _, call := range allMacroCalls(scanMacroInvocations(text)) {
			if call.Unsupported == "" && call.Span.Start == binding.Start && call.Span.End == binding.End && call.Name == target.Name && target.Arity != nil && *target.Arity == len(call.Arguments) {
				return true
			}
		}
		return false
	}
	for _, item := range result.Requirements.Items {
		if item.Kind != binding.Kind || item.Identity != target.Name || item.Resolution != "exact" {
			continue
		}
		for _, occ := range item.Occurrences {
			if occ.Location.Start.Offset == binding.Start && occ.Location.End.Offset == binding.End {
				return true
			}
		}
	}
	if owner != nil {
		for _, relation := range owner.Relations {
			if relation.Kind == binding.Kind && relation.Name == target.Name && relation.Start != nil && relation.End != nil && *relation.Start == binding.Start && *relation.End == binding.End {
				return true
			}
		}
	}
	return false
}
func validateBindings(req Request, root *analysis.Result) (map[string]*analysis.Result, error) {
	cache := map[string]*analysis.Result{}
	objects := map[string]Definition{}
	for _, o := range req.Bundle.Objects {
		objects[o.ID] = o
	}
	docs := map[string][]bindingDocument{textDigest(req.Document.Text): {{document: req.Document}}}
	for _, o := range req.Bundle.Objects {
		if o.Document != nil {
			key := textDigest(o.Document.Text)
			docs[key] = append(docs[key], bindingDocument{document: *o.Document, objectID: o.ID})
		}
	}
	seen := map[string]string{}
	for _, binding := range req.Bindings {
		target, found := objects[binding.ObjectID]
		if !found || target.Kind != binding.Kind || !validDigest(binding.DocumentDigest) {
			return nil, inputError("binding has invalid object, kind, or digest")
		}
		key := fmt.Sprintf("%s:%s:%d:%d", binding.DocumentDigest, binding.Kind, binding.Start, binding.End)
		if prior, exists := seen[key]; exists && prior != binding.ObjectID {
			return nil, inputError("conflicting bindings for one occurrence")
		}
		seen[key] = binding.ObjectID
		valid := false
		for _, entry := range docs[binding.DocumentDigest] {
			doc := entry.document
			if !validRange(doc.Text, binding.Start, binding.End) {
				continue
			}
			result := root
			var owner *Definition
			if entry.objectID != "" {
				value := objects[entry.objectID]
				owner = &value
				result = cache[entry.objectID]
				if result == nil {
					var err error
					result, err = analysis.Analyze(doc)
					if err != nil {
						return nil, inputError("analyze binding document: %v", err)
					}
					cache[entry.objectID] = result
				}
			}
			if bindingOccurrence(binding, doc, result, target, owner) {
				valid = true
				break
			}
		}
		if !valid {
			return nil, inputError("binding does not match an exact %s reference in its original document", binding.Kind)
		}
	}
	return cache, nil
}
func (e *evaluator) sourceText(source SourceInterval) string {
	if source.Kind == "query" {
		return e.req.Document.Text
	}
	if o, found := e.objects[source.ObjectID]; found && o.Document != nil {
		return o.Document.Text
	}
	return ""
}
func (e *evaluator) boundTarget(edge TraversalEdge, candidates []Definition) (Definition, bool) {
	text := e.sourceText(edge.Source)
	if text == "" {
		return Definition{}, false
	}
	digest := textDigest(text)
	for _, binding := range e.req.Bindings {
		if binding.DocumentDigest != digest || binding.Kind != edge.Kind || binding.Start != edge.Source.Start || binding.End != edge.Source.End {
			continue
		}
		for _, candidate := range candidates {
			if candidate.ID == binding.ObjectID {
				return candidate, true
			}
		}
	}
	return Definition{}, false
}
func (e *evaluator) checkCollection(edge TraversalEdge) {
	if e.collections[edge.Kind] == "complete" {
		return
	}
	e.addGap(ClosureGap{Code: "collection_incomplete", Kind: edge.Kind, Name: edge.Name, Source: edge.Source, Property: edge.Property, Path: append([]string{}, edge.Path...)}, "collections")
}
func (e *evaluator) resolve(edge TraversalEdge, arity *int, explicit bool) {
	e.checkCollection(edge)
	candidates := []Definition{}
	for _, o := range e.req.Bundle.Objects {
		if o.Kind != edge.Kind || o.Name != edge.Name {
			continue
		}
		if edge.Kind == "macro" && arity != nil && (o.Arity == nil || *o.Arity != *arity) {
			continue
		}
		candidates = append(candidates, o)
	}
	var target Definition
	switch {
	case edge.Resolution == "dynamic" || edge.Resolution == "wildcard":
		e.addGap(ClosureGap{Code: "dynamic_reference", Kind: edge.Kind, Name: edge.Name, Source: edge.Source, Property: edge.Property, Path: append([]string{}, edge.Path...)}, "resolution")
	case len(candidates) == 0:
		if e.collections[edge.Kind] == "complete" {
			edge.Resolution = "missing"
		} else {
			edge.Resolution = "unknown"
			e.report.Coverage.Collections = false
		}
		code := "missing_object"
		if edge.Resolution == "unknown" {
			code = "unknown_object"
		}
		if explicit {
			code = "missing_relation_target"
		}
		e.addGap(ClosureGap{Code: code, Kind: edge.Kind, Name: edge.Name, Source: edge.Source, Property: edge.Property, Path: append([]string{}, edge.Path...)}, "resolution")
	case len(candidates) == 1:
		target = candidates[0]
		edge.Resolution = "resolved"
		edge.ToObjectID = target.ID
	default:
		if bound, ok := e.boundTarget(edge, candidates); ok {
			target = bound
			edge.Resolution = "bound"
			edge.ToObjectID = target.ID
		} else {
			edge.Resolution = "ambiguous"
			e.addGap(ClosureGap{Code: "ambiguous_object", Kind: edge.Kind, Name: edge.Name, Source: edge.Source, Property: edge.Property, Path: append([]string{}, edge.Path...)}, "resolution")
		}
	}
	if target.ID != "" {
		for i, id := range e.active {
			if id == target.ID {
				edge.Resolution = "cycle"
				edge.CyclePath = append(append([]string{}, e.active[i:]...), target.ID)
				e.addGap(ClosureGap{Code: "cycle", Kind: edge.Kind, Name: edge.Name, Source: edge.Source, Property: edge.Property, Path: append([]string{}, edge.CyclePath...)}, "resolution")
				break
			}
		}
	}
	occurrenceID := e.addEdge(edge)
	if target.ID != "" && edge.Resolution != "cycle" {
		e.visit(target, edge.Kind == "macro" && arity == nil, occurrenceID)
	}
}
func (e *evaluator) inspectOriginalMacros(doc analysis.QueryDocument, owner sourceInterval, path []string) {
	if doc.Language != "spl" {
		return
	}
	for _, call := range allMacroCalls(scanMacroInvocations(doc.Text)) {
		source := SourceInterval{Kind: owner.Kind, SourceID: owner.SourceID, ObjectID: owner.ObjectID, Start: call.Span.Start, End: call.Span.End}
		edge := TraversalEdge{FromObjectID: owner.ObjectID, Kind: "macro", Name: call.Name, Source: source, Path: append([]string{}, path...), Origins: []SourceInterval{source}}
		if call.Unsupported != "" {
			edge.Resolution = "dynamic"
			e.resolve(edge, nil, false)
			continue
		}
		arity := len(call.Arguments)
		e.resolve(edge, &arity, false)
	}
}
func (e *evaluator) visit(def Definition, contextFreeMacro bool, occurrenceID string) {
	e.active = append(e.active, def.ID)
	defer func() { e.active = e.active[:len(e.active)-1] }()
	for _, relation := range def.Relations {
		source := SourceInterval{Kind: "definition", SourceID: def.SourceID, ObjectID: def.ID}
		property := ""
		if relation.Property != nil {
			property = *relation.Property
		} else if relation.Start != nil && relation.End != nil {
			source.Start = *relation.Start
			source.End = *relation.End
		}
		edge := TraversalEdge{FromObjectID: def.ID, Kind: relation.Kind, Name: relation.Name, Source: source, Property: property, Path: append([]string{}, e.active...), Origins: []SourceInterval{source}}
		e.resolve(edge, nil, true)
	}
	if def.Kind == "macro" {
		if def.Document == nil {
			if contextFreeMacro {
				e.addGap(ClosureGap{Code: "macro_context_missing", Kind: def.Kind, Name: def.Name, Source: SourceInterval{Kind: "definition", SourceID: def.SourceID, ObjectID: def.ID}, Path: append([]string{}, e.active...)}, "expansion")
			}
			return
		}
		if def.EvalBased != nil && *def.EvalBased {
			if contextFreeMacro {
				owner := sourceInterval{Kind: "definition", SourceID: def.SourceID, ObjectID: def.ID, Start: 0, End: len(def.Document.Text)}
				e.addGap(ClosureGap{Code: "macro_context_missing", Kind: def.Kind, Name: def.Name, Source: publicInterval(owner), Path: append([]string{}, e.active...)}, "expansion")
			}
			return
		}
		direct := e.directCache[def.ID]
		if direct == nil {
			var err error
			direct, err = analysis.Analyze(*def.Document)
			if err != nil {
				e.addGap(ClosureGap{Code: "definition_analysis_error", Kind: def.Kind, Name: def.Name, Source: SourceInterval{Kind: "definition", SourceID: def.SourceID, ObjectID: def.ID}, Path: append([]string{}, e.active...)}, "definitions")
				return
			}
			e.directCache[def.ID] = direct
		}
		if !e.bodyDone[def.ID] {
			e.report.DefinitionAnalyses = append(e.report.DefinitionAnalyses, DefinitionAnalysis{ObjectID: def.ID, DirectAnalysis: direct})
			e.bodyDone[def.ID] = true
		}
		owner := sourceInterval{Kind: "definition", SourceID: def.SourceID, ObjectID: def.ID, Start: 0, End: len(def.Document.Text)}
		if contextFreeMacro {
			e.addGap(ClosureGap{Code: "macro_context_missing", Kind: def.Kind, Name: def.Name, Source: publicInterval(owner), Path: append([]string{}, e.active...)}, "expansion")
		}
		if contextFreeMacro || def.Validation != nil && *def.Validation != "" {
			e.inspectExpansion(directExpansion(def.Document.Text, owner), owner, direct, false)
			e.inspectOriginalMacros(*def.Document, owner, e.active)
		}
		return
	}
	if def.Kind == "dataset" && def.Document == nil {
		return
	}
	if !queryBearing(def.Kind) {
		return
	}
	if def.Document == nil {
		e.addGap(ClosureGap{Code: "missing_definition_body", Kind: def.Kind, Name: def.Name, Source: SourceInterval{Kind: "definition", SourceID: def.SourceID, ObjectID: def.ID}, Path: append([]string{}, e.active...)}, "definitions")
		return
	}
	body, err := e.bodyFor(def)
	if err != nil {
		e.addGap(ClosureGap{Code: "definition_analysis_error", Kind: def.Kind, Name: def.Name, Source: SourceInterval{Kind: "definition", SourceID: def.SourceID, ObjectID: def.ID}, Path: append([]string{}, e.active...)}, "definitions")
		return
	}
	if !e.bodyDone[def.ID] {
		e.report.DefinitionAnalyses = append(e.report.DefinitionAnalyses, DefinitionAnalysis{ObjectID: def.ID, DirectAnalysis: body.direct, EffectiveAnalysis: body.effective})
		e.bodyDone[def.ID] = true
	}
	expanded := body.expanded.forOccurrence(occurrenceID)
	owner := sourceInterval{Kind: "definition", SourceID: def.SourceID, ObjectID: def.ID, Start: 0, End: len(def.Document.Text)}
	e.inspectExpansion(expanded, owner, body.effective, false)
	e.inspectOriginalMacros(*def.Document, owner, e.active)
}

type bodyEvaluation struct {
	direct    *analysis.Result
	effective *analysis.Result
	expanded  expansion
}

func (e *evaluator) bodyFor(def Definition) (bodyEvaluation, error) {
	if cached, found := e.bodyCache[def.ID]; found {
		return cached, nil
	}
	direct := e.directCache[def.ID]
	if direct == nil {
		var err error
		direct, err = analysis.Analyze(*def.Document)
		if err != nil {
			return bodyEvaluation{}, err
		}
		e.directCache[def.ID] = direct
	}
	input := directExpansion(def.Document.Text, sourceInterval{Kind: "definition", SourceID: def.SourceID, ObjectID: def.ID, Start: 0, End: len(def.Document.Text)})
	expanded := input
	if def.Document.Language == "spl" {
		expanded = (&macroExpander{request: e.req}).expand(input, map[string]bool{}, 0)
	}
	effectiveDoc := *def.Document
	effectiveDoc.Text = expanded.Text
	effective, err := analysis.Analyze(effectiveDoc)
	if err != nil {
		return bodyEvaluation{}, err
	}
	body := bodyEvaluation{direct: direct, effective: effective, expanded: expanded}
	if e.bodyCache == nil {
		e.bodyCache = map[string]bodyEvaluation{}
	}
	e.bodyCache[def.ID] = body
	if e.expandedByResult == nil {
		e.expandedByResult = map[*analysis.Result]expansion{}
	}
	e.expandedByResult[effective] = expanded
	return body, nil
}
func (cached expansion) forOccurrence(occurrenceID string) expansion {
	if occurrenceID == "" {
		return cached
	}
	out := cached
	out.Segments = append([]provenanceSegment{}, cached.Segments...)
	for i := range out.Segments {
		chain := append([]invocationFrame{}, out.Segments[i].InvocationChain...)
		for j := range chain {
			chain[j].InstanceID = occurrenceID + "/" + chain[j].InstanceID
		}
		out.Segments[i].InvocationChain = chain
	}
	return out
}
func queryBearing(kind string) bool {
	switch kind {
	case "saved_search", "event_type", "dataset", "module", "function":
		return true
	}
	return false
}
func isObjectKind(kind string) bool { return objectKinds[kind] }
func skipLegacyDataset(kind, name string) bool {
	return kind == "dataset" && strings.HasPrefix(strings.ToLower(name), "savedsearch:")
}
