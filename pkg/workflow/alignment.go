package workflow

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/closure"
)

type alignmentFact struct{ key, pointer, id string }

func factKey(v any) string { b, _ := json.Marshal(v); return string(b) }

// pairFacts admits only unique one-to-one correspondences. IDs are deliberately
// excluded from keys because supplied historical contracts may rename them.
func pairFacts(e *ComparisonEntry, b, a []alignmentFact, basis string) map[string]string {
	bm, am := map[string][]alignmentFact{}, map[string][]alignmentFact{}
	for _, f := range b {
		bm[f.key] = append(bm[f.key], f)
	}
	for _, f := range a {
		am[f.key] = append(am[f.key], f)
	}
	keys := []string{}
	seen := map[string]bool{}
	for _, f := range append(append([]alignmentFact{}, b...), a...) {
		if !seen[f.key] {
			seen[f.key] = true
			keys = append(keys, f.key)
		}
	}
	mapping := map[string]string{}
	for _, key := range keys {
		x, y := bm[key], am[key]
		switch {
		case len(x) == 1 && len(y) == 1:
			e.Pairs = append(e.Pairs, EvidencePair{BeforePointer: x[0].pointer, AfterPointer: y[0].pointer, Basis: basis})
			mapping[x[0].id] = y[0].id
		case len(x) > 1 || len(y) > 1:
			for _, f := range append(append([]alignmentFact{}, x...), y...) {
				e.Ambiguous = append(e.Ambiguous, f.pointer)
			}
		default:
			for _, f := range append(append([]alignmentFact{}, x...), y...) {
				e.Unmatched = append(e.Unmatched, f.pointer)
			}
		}
	}
	return mapping
}
func alignEntry(e *ComparisonEntry, index int) {
	b, a := e.Before.Analysis, e.After.Analysis
	bp, ap := fmt.Sprintf("/entries/%d/before/analysis", index), fmt.Sprintf("/entries/%d/after/analysis", index)
	alignAnalysis(e, b, a, bp, ap)
	if e.Before.Compatibility != nil && e.After.Compatibility != nil {
		alignClosure(e, e.Before.Compatibility.Closure, e.After.Compatibility.Closure, fmt.Sprintf("/entries/%d/before/compatibility/closure", index), fmt.Sprintf("/entries/%d/after/compatibility/closure", index))
	}
	if e.Before.Resolution != nil && e.After.Resolution != nil {
		bf, af := []alignmentFact{}, []alignmentFact{}
		for i, v := range e.Before.Resolution.Variants {
			bf = append(bf, alignmentFact{key: variantKey(v.Selection), pointer: fmt.Sprintf("/entries/%d/before/resolution/variants/%d", index, i), id: fmt.Sprint(i)})
		}
		for i, v := range e.After.Resolution.Variants {
			af = append(af, alignmentFact{key: variantKey(v.Selection), pointer: fmt.Sprintf("/entries/%d/after/resolution/variants/%d", index, i), id: fmt.Sprint(i)})
		}
		mapping := pairFacts(e, bf, af, "canonical_resolution_selection")
		for x := range bf {
			ai, ok := mapping[fmt.Sprint(x)]
			if !ok {
				continue
			}
			var y int
			fmt.Sscan(ai, &y)
			bv, av := e.Before.Resolution.Variants[x], e.After.Resolution.Variants[y]
			if bv.CandidateAnalysis != nil && av.CandidateAnalysis != nil {
				bvp, avp := bf[x].pointer, af[y].pointer
				if sameQuery(bv.CandidateAnalysis.Document, av.CandidateAnalysis.Document) && originalRolesKey(bv.Changes) == originalRolesKey(av.Changes) {
					alignAnalysis(e, bv.CandidateAnalysis, av.CandidateAnalysis, bvp+"/candidate_analysis", avp+"/candidate_analysis")
					if bv.Compatibility != nil && av.Compatibility != nil {
						alignClosure(e, bv.Compatibility.Closure, av.Compatibility.Closure, bvp+"/compatibility/closure", avp+"/compatibility/closure")
					}
				} else {
					e.Unmatched = append(e.Unmatched, bvp+"/candidate_analysis", avp+"/candidate_analysis")
				}
			}
		}
	}
}
func variantKey(in []analysis.ResolutionChoice) string {
	out := append([]analysis.ResolutionChoice{}, in...)
	sort.Slice(out, func(i, j int) bool {
		if out[i].Placeholder != out[j].Placeholder {
			return out[i].Placeholder < out[j].Placeholder
		}
		return out[i].Kind < out[j].Kind
	})
	return factKey(out)
}
func alignAnalysis(e *ComparisonEntry, b, a *analysis.Result, bp, ap string) {
	refs := func(r *analysis.Result, p string) []alignmentFact {
		out := []alignmentFact{}
		for i, v := range r.References {
			out = append(out, alignmentFact{key: factKey([]any{v.Kind, v.Role, v.FieldIdentity, v.NormalizedName, v.Location, v.Binding, scopeInterval(r, v.ScopeID)}), pointer: fmt.Sprintf("%s/references/%d", p, i), id: v.ID})
		}
		return out
	}
	refmap := pairFacts(e, refs(b, bp), refs(a, ap), "original_interval_typed_role")
	inputs := func(r *analysis.Result, p string) []alignmentFact {
		out := []alignmentFact{}
		for i, v := range r.Inputs {
			sites := []any{}
			for _, o := range v.Occurrences {
				sites = append(sites, []any{o.Location, o.UseSiteLocations, scopeInterval(r, o.ScopeID)})
			}
			out = append(out, alignmentFact{key: factKey([]any{v.Kind, v.Identity, sites}), pointer: fmt.Sprintf("%s/inputs/%d", p, i), id: v.ID})
		}
		return out
	}
	inputmap := pairFacts(e, inputs(b, bp), inputs(a, ap), "original_input_identity_intervals")
	occurrenceFacts := func(r *analysis.Result, p string, before bool) []alignmentFact {
		out := []alignmentFact{}
		for i, in := range r.Inputs {
			owner := in.ID
			known := true
			if before {
				owner, known = inputmap[owner]
			}
			for j, o := range in.Occurrences {
				key := factKey([]any{owner, o.Location, o.UseSiteLocations, o.Alias, scopeInterval(r, o.ScopeID)})
				if !known {
					key = p + key
				}
				out = append(out, alignmentFact{key: key, pointer: fmt.Sprintf("%s/inputs/%d/occurrences/%d", p, i, j), id: o.ID})
			}
		}
		return out
	}
	occmap := pairFacts(e, occurrenceFacts(b, bp, true), occurrenceFacts(a, ap, false), "unique_input_original_occurrence")
	requirements := func(r *analysis.Result, p string, before bool) []alignmentFact {
		out := []alignmentFact{}
		for i, v := range r.Requirements.Items {
			owner := v.InputID
			owners := append([]string{}, v.Ownership.CandidateInputIDs...)
			known := true
			if before {
				if owner != "" {
					owner, known = inputmap[owner]
				}
				for j, id := range owners {
					mapped, ok := inputmap[id]
					owners[j] = mapped
					known = known && ok
				}
			}
			sort.Strings(owners)
			occurrences := []any{}
			for _, o := range v.Occurrences {
				rid := o.ReferenceID
				if before {
					mapped, ok := refmap[rid]
					rid = mapped
					known = known && ok
				}
				oids := append([]string{}, o.InputOccurrenceIDs...)
				if before {
					for j, id := range oids {
						mapped, ok := occmap[id]
						oids[j] = mapped
						known = known && ok
					}
				}
				sort.Strings(oids)
				occurrences = append(occurrences, []any{rid, o.Location, o.Necessity, oids})
			}
			key := factKey([]any{v.Kind, v.Identity, v.FieldIdentity, v.Role, v.Necessity, v.Origin, v.Ownership.State, owner, owners, occurrences})
			if !known {
				key = p + key
			}
			out = append(out, alignmentFact{key: key, pointer: fmt.Sprintf("%s/requirements/items/%d", p, i), id: v.ID})
		}
		return out
	}
	pairFacts(e, requirements(b, bp, true), requirements(a, ap, false), "unique_original_occurrences_and_ownership")
}
func scopeInterval(r *analysis.Result, id string) any {
	for _, s := range r.Scopes {
		if s.ID == id {
			return []any{s.Kind, s.Location}
		}
	}
	return nil
}

func originalRolesKey(changes []analysis.ResolutionChange) string {
	out := []any{}
	for _, c := range changes {
		out = append(out, []any{c.Placeholder, c.Kind, c.OriginalLocation, c.Before})
	}
	return factKey(out)
}

// Definition bodies align only under captured object identity, exact original
// definition bytes and situated invocation source ownership. Expanded offsets
// alone are never correspondence evidence.
func alignClosure(e *ComparisonEntry, b, a *closure.Report, bp, ap string) {
	if b == nil || a == nil {
		return
	}
	// The closure's original and expanded analyses have their own ID domains.
	// Admit them only when the retained query bytes identify the same owner.
	if b.DirectAnalysis != nil && a.DirectAnalysis != nil && sameQuery(b.DirectAnalysis.Document, a.DirectAnalysis.Document) {
		alignAnalysis(e, b.DirectAnalysis, a.DirectAnalysis, bp+"/direct_analysis", ap+"/direct_analysis")
	}
	if b.EffectiveAnalysis != nil && a.EffectiveAnalysis != nil && sameQuery(b.EffectiveAnalysis.Document, a.EffectiveAnalysis.Document) {
		alignAnalysis(e, b.EffectiveAnalysis, a.EffectiveAnalysis, bp+"/effective_analysis", ap+"/effective_analysis")
	}
	facts := func(c *closure.Report, p string) []alignmentFact {
		out := []alignmentFact{}
		for i, d := range c.DefinitionAnalyses {
			contexts := []any{}
			for _, ctx := range c.DefinitionContexts {
				if ctx.ObjectID == d.ObjectID {
					for _, seg := range ctx.Provenance {
						frames := []any{}
						for _, f := range seg.InvocationChain {
							frames = append(frames, []any{f.ObjectID, alignmentIntervals(f.Invocation)})
						}
						contexts = append(contexts, []any{alignmentIntervals([]closure.SourceInterval{seg.Source}), frames, alignmentIntervals(seg.Placeholders)})
					}
				}
			}
			digest := ""
			if d.DirectAnalysis != nil {
				digest = d.DirectAnalysis.Requirements.Query.QueryDigest
			}
			out = append(out, alignmentFact{key: factKey([]any{d.ObjectID, digest, contexts}), pointer: fmt.Sprintf("%s/definition_analyses/%d", p, i), id: fmt.Sprint(i)})
		}
		return out
	}
	bf, af := facts(b, bp), facts(a, ap)
	mapping := pairFacts(e, bf, af, "captured_definition_digest_invocation_sources")
	// Iteration follows the retained before order rather than map order.
	for i := range bf {
		ai, ok := mapping[fmt.Sprint(i)]
		if !ok {
			continue
		}
		var j int
		fmt.Sscan(ai, &j)
		x, y := b.DefinitionAnalyses[i], a.DefinitionAnalyses[j]
		if x.DirectAnalysis != nil && y.DirectAnalysis != nil {
			alignAnalysis(e, x.DirectAnalysis, y.DirectAnalysis, bf[i].pointer+"/direct_analysis", af[j].pointer+"/direct_analysis")
		}
		if x.EffectiveAnalysis != nil && y.EffectiveAnalysis != nil && sameQuery(x.EffectiveAnalysis.Document, y.EffectiveAnalysis.Document) {
			alignAnalysis(e, x.EffectiveAnalysis, y.EffectiveAnalysis, bf[i].pointer+"/effective_analysis", af[j].pointer+"/effective_analysis")
		}
	}
}

// SourceID records provenance. Source ownership is the query/definition kind and
// captured ObjectID plus byte interval, so moving a source does not change it.
func alignmentIntervals(in []closure.SourceInterval) []closure.SourceInterval {
	out := append([]closure.SourceInterval{}, in...)
	for i := range out {
		out[i].SourceID = ""
	}
	return out
}
