package closure

import (
	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"strings"
)

func (e *evaluator) inspectExpansion(expanded expansion, owner sourceInterval, result *analysis.Result, root bool) {
	for _, gap := range expanded.Gaps {
		source := SourceInterval{Kind: owner.Kind, SourceID: owner.SourceID, ObjectID: owner.ObjectID, Start: owner.Start, End: owner.End}
		if len(gap.Origins) > 0 {
			source = publicInterval(gap.Origins[0])
		}
		dimension := "expansion"
		e.addGap(ClosureGap{Code: "expansion_" + gap.Reason, Source: source, Path: append([]string{}, e.active...)}, dimension)
		e.expansionGapEdge(gap, owner, expanded)
	}
	if result.Status != analysis.Valid || !result.Requirements.Coverage.Complete {
		dimension := "definitions"
		if root {
			dimension = "effective"
		}
		e.addGap(ClosureGap{Code: "analysis_incomplete", Source: SourceInterval{Kind: owner.Kind, SourceID: owner.SourceID, ObjectID: owner.ObjectID, Start: owner.Start, End: owner.End}, Path: append([]string{}, e.active...)}, dimension)
	}
	for _, d := range result.Diagnostics {
		origin := expanded.Origins(d.Location.Start.Offset, d.Location.End.Offset)
		source := SourceInterval{Kind: owner.Kind, SourceID: owner.SourceID, ObjectID: owner.ObjectID, Start: owner.Start, End: owner.End}
		if len(origin) > 0 {
			source = publicInterval(origin[0].Source)
		}
		e.report.Diagnostics = append(e.report.Diagnostics, ClosureDiagnostic{Source: source, Diagnostic: d})
	}
	refs := map[string]analysis.Reference{}
	for _, r := range result.References {
		refs[r.ID] = r
	}
	for _, item := range result.Requirements.Items {
		if !isObjectKind(item.Kind) || item.Kind == "macro" {
			continue
		}
		kind, name := item.Kind, item.Identity
		if skipLegacyDataset(kind, name) {
			if item.Resolution == "exact" {
				continue
			}
			kind = "saved_search"
			name = name[len("savedsearch:"):]
		}
		for _, occ := range item.Occurrences {
			ref, found := refs[occ.ReferenceID]
			if !found {
				continue
			}
			origins := expanded.Origins(ref.Location.Start.Offset, ref.Location.End.Offset)
			evidence := []SourceInterval{}
			seenEvidence := map[SourceInterval]bool{}
			addOrigin := func(v sourceInterval) {
				value := publicInterval(v)
				if !seenEvidence[value] {
					evidence = append(evidence, value)
					seenEvidence[value] = true
				}
			}
			var chain []invocationFrame
			for _, origin := range origins {
				addOrigin(origin.Source)
				if len(origin.Placeholders) > 0 {
					for _, placeholder := range origin.Placeholders {
						addOrigin(placeholder)
					}
				} else if origin.Placeholder != nil {
					addOrigin(*origin.Placeholder)
				}
				if len(origin.InvocationChain) > len(chain) {
					chain = origin.InvocationChain
				}
			}
			source := SourceInterval{Kind: owner.Kind, SourceID: owner.SourceID, ObjectID: owner.ObjectID, Start: owner.Start, End: owner.End}
			if len(evidence) > 0 {
				source = evidence[0]
			}
			path := append([]string{}, e.active...)
			for _, frame := range chain {
				if len(path) == 0 || path[len(path)-1] != frame.ObjectID {
					path = append(path, frame.ObjectID)
				}
			}
			from := owner.ObjectID
			if source.ObjectID != "" {
				from = source.ObjectID
			}
			if len(chain) > 0 {
				from = chain[len(chain)-1].ObjectID
			}
			edge := TraversalEdge{FromObjectID: from, Kind: kind, Name: name, Source: source, ReferenceID: ref.ID, Path: path, InvocationChain: publicFrames(chain), Origins: evidence}
			if item.Resolution != "exact" || ref.Resolution != "exact" || len(evidence) == 0 || strings.Contains(name, "*") {
				edge.Resolution = "dynamic"
			}
			prior := e.active
			e.active = append([]string{}, path...)
			e.resolve(edge, nil, false)
			e.active = prior
		}
	}
	// Macro definition bodies may be inserted more than once with different
	// arguments. Each expansion frame is an instance, even when its body was
	// already analyzed as part of the effective query.
	instances := map[string]bool{}
	for _, segment := range expanded.Segments {
		for index, frame := range segment.InvocationChain {
			if instances[frame.InstanceID] {
				continue
			}
			instances[frame.InstanceID] = true
			if len(frame.Invocation) == 0 {
				continue
			}
			source := publicInterval(frame.Invocation[0])
			if source.Kind != "definition" {
				continue
			} // original root/query body calls are scanned below
			parent, ok := e.objects[source.ObjectID]
			if !ok || parent.Kind != "macro" {
				continue
			}
			target, ok := e.objects[frame.ObjectID]
			if !ok {
				continue
			}
			origins := make([]SourceInterval, 0, len(frame.Invocation))
			for _, v := range frame.Invocation {
				origins = append(origins, publicInterval(v))
			}
			path := append([]string{}, e.active...)
			for _, prior := range segment.InvocationChain[:index] {
				path = append(path, prior.ObjectID)
			}
			edge := TraversalEdge{FromObjectID: parent.ID, Kind: "macro", Name: target.Name, Source: source, Path: path, InvocationChain: publicFrames(segment.InvocationChain[:index+1]), Origins: origins}
			edge.ToObjectID = target.ID
			edge.Resolution = "resolved"
			matches := 0
			for _, candidate := range e.req.Bundle.Objects {
				if candidate.Kind == "macro" && candidate.Name == target.Name && candidate.Arity != nil && target.Arity != nil && *candidate.Arity == *target.Arity {
					matches++
				}
			}
			if matches > 1 {
				edge.Resolution = "bound"
			}
			e.checkCollection(edge)
			occurrenceID := e.addEdge(edge)
			prior := e.active
			e.active = append([]string{}, path...)
			e.visit(target, false, occurrenceID)
			e.active = prior
		}
	}
}

func (e *evaluator) expansionGapEdge(gap opaqueGap, owner sourceInterval, expanded expansion) {
	if len(gap.Origins) == 0 {
		return
	}
	source := gap.Origins[0]
	if source.Kind != "definition" {
		return
	}
	def, ok := e.objects[source.ObjectID]
	if !ok || def.Document == nil || def.Kind != "macro" {
		return
	}
	for _, call := range allMacroCalls(scanMacroInvocations(def.Document.Text)) {
		if call.Span.Start != source.Start || call.Span.End != source.End {
			continue
		}
		edge := TraversalEdge{FromObjectID: def.ID, Kind: "macro", Name: call.Name, Source: publicInterval(source), Path: append([]string{}, e.active...), Origins: []SourceInterval{publicInterval(source)}}
		if gap.Reason == "cycle" {
			edge.Resolution = "cycle"
			path := []string{}
			for _, segment := range expanded.Segments {
				if segment.EffectiveStart <= gap.EffectiveStart && segment.EffectiveEnd >= gap.EffectiveEnd {
					for _, frame := range segment.InvocationChain {
						path = append(path, frame.ObjectID)
					}
					break
				}
			}
			active := map[string]bool{}
			for _, id := range path {
				active[id] = true
			}
			candidates := []Definition{}
			for _, candidate := range e.req.Bundle.Objects {
				if candidate.Kind == "macro" && candidate.Name == call.Name && candidate.Arity != nil && *candidate.Arity == len(call.Arguments) && active[candidate.ID] {
					candidates = append(candidates, candidate)
				}
			}
			targetID := ""
			if len(candidates) == 1 {
				targetID = candidates[0].ID
			} else if len(candidates) > 1 {
				if bound, ok := e.boundTarget(edge, candidates); ok {
					targetID = bound.ID
				}
			}
			if targetID != "" {
				edge.ToObjectID = targetID
				for i, id := range path {
					if id == targetID {
						edge.CyclePath = append(append([]string{}, path[i:]...), targetID)
						break
					}
				}
			}
			for _, segment := range expanded.Segments {
				if segment.EffectiveStart <= gap.EffectiveStart && segment.EffectiveEnd >= gap.EffectiveEnd {
					edge.InvocationChain = publicFrames(segment.InvocationChain)
					break
				}
			}
			e.checkCollection(edge)
			e.addEdge(edge)
			return
		}
		if call.Unsupported != "" {
			edge.Resolution = "dynamic"
			e.resolve(edge, nil, false)
			return
		}
		arity := len(call.Arguments)
		e.resolve(edge, &arity, false)
		return
	}
}
