package workflow

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/impact"
)

// Findings retain canonical owner values. Normalization is comparison-only; the
// exported delta values and pointers always refer to the retained saved evidence.
type comparisonFinding struct {
	category, pointer string
	value             any
}

func classifyComparison(failed, definiteDelta, complete, aligned bool) impact.Classification {
	switch {
	case failed:
		return impact.Failed
	case definiteDelta:
		return impact.Affected
	case !complete || !aligned:
		return impact.Indeterminate
	default:
		return impact.Unchanged
	}
}

func comparisonComplete(e ReportEntry) bool {
	if e.Failure != nil {
		return false
	}
	if e.Compatibility != nil {
		// A supported closure can discharge gaps in the original unexpanded query.
		for _, coverage := range e.Compatibility.Coverage {
			if coverage.State != "complete" && coverage.State != "not_applicable" {
				return false
			}
		}
		return e.Compatibility.Outcome == "satisfied" || e.Compatibility.Outcome == "unsatisfied"
	}
	if e.Resolution != nil {
		if len(e.Resolution.Variants) == 0 || e.Resolution.Counts.Incomplete > 0 {
			return false
		}
		for _, v := range e.Resolution.Variants {
			if v.Compatibility != nil {
				for _, coverage := range v.Compatibility.Coverage {
					if coverage.Evidence.State != "complete" && coverage.Evidence.State != "not_applicable" {
						return false
					}
				}
			}
			if v.CandidateAnalysis != nil && v.CandidateAnalysis.Status == analysis.Incomplete && (v.Compatibility == nil || (v.Compatibility.Outcome != "satisfied" && v.Compatibility.Outcome != "unsatisfied")) {
				return false
			}
		}
		return true
	}
	return false
}

func evidenceMap(v any) map[string]any {
	b, _ := json.Marshal(v)
	var m map[string]any
	_ = json.Unmarshal(b, &m)
	return m
}
func pointerValue(root any, pointer string) any {
	v := root
	for _, part := range strings.Split(strings.TrimPrefix(pointer, "/"), "/") {
		switch x := v.(type) {
		case map[string]any:
			v = x[part]
		case []any:
			var i int
			if _, err := fmt.Sscan(part, &i); err != nil || i < 0 || i >= len(x) {
				return nil
			}
			v = x[i]
		default:
			return nil
		}
	}
	return v
}

func entryFindings(entry ReportEntry, p string) []comparisonFinding {
	m := evidenceMap(entry)
	out := []comparisonFinding{}
	add := func(category, path string, v any) {
		if v != nil {
			out = append(out, comparisonFinding{category, p + path, typedPointer(entry, path)})
		}
	}
	addClosure := func(base string, raw any) {
		c, ok := raw.(map[string]any)
		if !ok {
			return
		}
		keys := []string{}
		for key := range c {
			if key != "bundle_digest" {
				keys = append(keys, key)
			}
		}
		sort.Strings(keys)
		for _, key := range keys {
			if key == "definition_analyses" {
				for i, v := range c[key].([]any) {
					add("closure", fmt.Sprintf("%s/%s/%d", base, key, i), v)
				}
			} else {
				add("closure", base+"/"+key, c[key])
			}
		}
	}
	a := m["analysis"].(map[string]any)
	for _, name := range []string{"requirements", "diagnostics", "coverage", "correlation"} {
		add(name, "/analysis/"+name, a[name])
	}
	if c, ok := m["compatibility"].(map[string]any); ok {
		addClosure("/compatibility/closure", c["closure"])
		for _, name := range []string{"diagnostics", "coverage", "correlation"} {
			add(name, "/compatibility/"+name, c[name])
		}
		for _, name := range []string{"outcome", "reasons", "effective_requirements"} {
			add("compatibility", "/compatibility/"+name, c[name])
		}
		for _, name := range []string{"inputs", "requirement_outcomes", "input_bindings", "dependency_bindings", "query_scope"} {
			add("supporting_fact", "/compatibility/"+name, c[name])
		}
	}
	if r, ok := m["resolution"].(map[string]any); ok {
		for _, name := range []string{"resolutions", "max_variants", "total_combinations", "generated_count"} {
			add("resolution_selection", "/resolution/"+name, r[name])
		}
		variants := r["variants"].([]any)
		// Membership is a set of selections; enumeration order remains separately
		// visible in resolutions, without inventing additions or removals.
		add("resolution_selection", "/resolution/variants", variants)
		for i, raw := range variants {
			v := raw.(map[string]any)
			vp := fmt.Sprintf("/resolution/variants/%d", i)
			for _, name := range []string{"outcome", "candidate_text", "resolved_query", "changes", "proof", "diagnostics"} {
				add("resolution_result", vp+"/"+name, v[name])
			}
			if ca, ok := v["candidate_analysis"].(map[string]any); ok {
				for _, name := range []string{"requirements", "diagnostics", "coverage", "correlation"} {
					add(name, vp+"/candidate_analysis/"+name, ca[name])
				}
			}
			if c, ok := v["compatibility"].(map[string]any); ok {
				addClosure(vp+"/compatibility/closure", c["closure"])
				for _, name := range []string{"coverage", "correlation", "diagnostics"} {
					add(name, vp+"/compatibility/"+name, c[name])
				}
				for _, name := range []string{"outcome", "reasons", "effective_requirements"} {
					add("compatibility", vp+"/compatibility/"+name, c[name])
				}
				for _, name := range []string{"inputs", "requirement_outcomes", "original_role_decisions", "input_bindings", "dependency_bindings", "effective_dependency_bindings"} {
					add("supporting_fact", vp+"/compatibility/"+name, c[name])
				}
			}
		}
	}
	return out
}

// Only structural ID fields participate in renumbering. Object and schema
// identities are captured evidence and must never be erased or normalized.
func structuralID(key string) bool {
	switch key {
	case "id", "input_id", "scope_id", "stage_id", "parent_id", "reference_id", "requirement_id", "occurrence_id", "original_reference_id", "original_input_id", "candidate_input_id", "original_id", "candidate_id", "original_occurrence_id", "candidate_occurrence_id", "original_requirement_id", "candidate_requirement_id", "output_reference_id", "input_occurrence_ids", "origin_reference_ids", "reference_ids", "candidate_input_ids", "requirement_ids", "input_reference_ids", "original_reference_ids", "candidate_reference_ids", "use_site_reference_ids", "components":
		return true
	}
	return false
}

func typedPointer(root any, pointer string) any {
	v := reflect.ValueOf(root)
	for _, part := range strings.Split(strings.TrimPrefix(pointer, "/"), "/") {
		for v.Kind() == reflect.Pointer {
			if v.IsNil() {
				return nil
			}
			v = v.Elem()
		}
		switch v.Kind() {
		case reflect.Struct:
			found := false
			for i := 0; i < v.NumField(); i++ {
				if strings.Split(v.Type().Field(i).Tag.Get("json"), ",")[0] == part {
					v = v.Field(i)
					found = true
					break
				}
			}
			if !found {
				return nil
			}
		case reflect.Slice:
			var i int
			if _, err := fmt.Sscan(part, &i); err != nil || i < 0 || i >= v.Len() {
				return nil
			}
			v = v.Index(i)
		default:
			return nil
		}
	}
	return v.Interface()
}

// Metadata exclusions name exact owner types and fields. Raw schema payloads
// are decoded as data and never traversed as owner metadata or structural IDs.
func comparisonMetadata(t reflect.Type, field string) bool {
	owner := t.PkgPath() + "." + t.Name()
	switch owner {
	case "github.com/delgado-jacob/spl-toolkit/pkg/analysis.QueryDocument":
		return field == "SourceID"
	case "github.com/delgado-jacob/spl-toolkit/pkg/analysis.RequirementQueryIdentity":
		return field == "SourceID" || field == "QueryDigest"
	case "github.com/delgado-jacob/spl-toolkit/pkg/analysis.RequirementSet":
		return field == "CapabilityRevision"
	case "github.com/delgado-jacob/spl-toolkit/pkg/environment.Provenance":
		return field == "SourceID" || field == "ObservedAt"
	case "github.com/delgado-jacob/spl-toolkit/pkg/closure.SourceInterval":
		return field == "SourceID"
	case "github.com/delgado-jacob/spl-toolkit/pkg/closure.Report":
		return field == "BundleDigest"
	case "github.com/delgado-jacob/spl-toolkit/pkg/compatibility.Provenance", "github.com/delgado-jacob/spl-toolkit/pkg/resolution.Provenance":
		switch field {
		case "EnvironmentDigest", "SchemaBundleDigest", "AssessmentIdentityDigest", "ResolutionInputDigest", "AssessmentInputDigest", "CapabilityRevision", "SourceID":
			return true
		}
	}
	return false
}
func normalizeFinding(value any, ids func(string) map[string]string, path string, stripUnknown bool) any {
	var walk func(reflect.Value, string, bool) any
	walk = func(v reflect.Value, p string, localID bool) any {
		if !v.IsValid() {
			return nil
		}
		if v.Type() == reflect.TypeOf(json.RawMessage{}) {
			var out any
			_ = json.Unmarshal(v.Interface().(json.RawMessage), &out)
			return out
		}
		switch v.Kind() {
		case reflect.Pointer, reflect.Interface:
			if v.IsNil() {
				return nil
			}
			return walk(v.Elem(), p, localID)
		case reflect.Struct:
			out := map[string]any{}
			t := v.Type()
			for i := 0; i < v.NumField(); i++ {
				f := t.Field(i)
				if !f.IsExported() || comparisonMetadata(t, f.Name) {
					continue
				}
				name := strings.Split(f.Tag.Get("json"), ",")[0]
				if name == "" || name == "-" {
					continue
				}
				local := structuralID(name) && (strings.HasSuffix(t.PkgPath(), "/analysis") || strings.HasSuffix(t.PkgPath(), "/compatibility") || strings.HasSuffix(t.PkgPath(), "/validation") || (strings.HasSuffix(t.PkgPath(), "/closure") && name == "reference_id"))
				out[name] = walk(v.Field(i), p+"/"+name, local)
			}
			return out
		case reflect.Slice, reflect.Array:
			out := make([]any, v.Len())
			for i := range out {
				out[i] = walk(v.Index(i), fmt.Sprintf("%s/%d", p, i), localID)
			}
			return out
		case reflect.String:
			if localID {
				if id, ok := ids(p)[v.String()]; ok {
					return id
				}
				if stripUnknown {
					return "unmapped-local-id"
				}
			}
			return v.String()
		default:
			return v.Interface()
		}
	}
	return walk(reflect.ValueOf(value), path, false)
}

func evidenceDeltas(e *ComparisonEntry, index int) bool {
	bp, ap := fmt.Sprintf("/entries/%d/before", index), fmt.Sprintf("/entries/%d/after", index)
	bm, am := evidenceMap(e.Before), evidenceMap(e.After)
	// Scope the ID map to a particular analysis domain. Reused candidate IDs
	// cannot authorize correspondence with another variant or definition.
	mappings := map[string]map[string]string{}
	for _, pair := range e.Pairs {
		if !strings.Contains(pair.BeforePointer, "/analysis/") && !strings.Contains(pair.BeforePointer, "_analysis/") {
			continue
		}
		x := pointerValue(bm, strings.TrimPrefix(pair.BeforePointer, bp))
		y := pointerValue(am, strings.TrimPrefix(pair.AfterPointer, ap))
		xm, xok := x.(map[string]any)
		ym, yok := y.(map[string]any)
		if !xok || !yok {
			continue
		}
		bid, bok := xm["id"].(string)
		aid, aok := ym["id"].(string)
		if !bok || !aok {
			continue
		}
		domain := pair.AfterPointer[:strings.LastIndex(pair.AfterPointer, "/")]
		if j := strings.LastIndex(domain, "/requirements/items"); j >= 0 {
			domain = domain[:j]
		} else if j := strings.LastIndex(domain, "/inputs/"); j >= 0 {
			domain = domain[:j]
		} else {
			domain = domain[:strings.LastIndex(domain, "/")]
		}
		if mappings[domain] == nil {
			mappings[domain] = map[string]string{}
		}
		mappings[domain][aid] = bid
		beforeDomain := strings.TrimPrefix(pair.BeforePointer[:len(pair.BeforePointer)-len(pair.AfterPointer[len(domain):])], bp)
		afterDomain := strings.TrimPrefix(domain, ap)
		bd, bok := pointerValue(bm, beforeDomain).(map[string]any)
		ad, aok := pointerValue(am, afterDomain).(map[string]any)
		if bok && aok {
			addSituatedIDs(bd, ad, domain, mappings)
		}
	}
	// Stages/scopes have no public pair list; exact unique source intervals and
	// typed owners establish their local renumbering correspondence.
	addSituatedIDs(bm["analysis"].(map[string]any), am["analysis"].(map[string]any), ap+"/analysis", mappings)
	idsFor := func(path string) map[string]string {
		best := ap + "/analysis"
		if strings.Contains(path, "/original_") {
			return mappings[best]
		}
		for domain := range mappings {
			if strings.HasPrefix(path, domain) && len(domain) > len(best) {
				best = domain
			}
		}
		if strings.Contains(path, "/resolution/variants/") {
			for domain := range mappings {
				if strings.Contains(domain, "/candidate_analysis") {
					base := strings.TrimSuffix(domain, "/candidate_analysis")
					if strings.HasPrefix(path, base) && len(domain) > len(best) {
						best = domain
					}
				}
			}
		}
		return mappings[best]
	}
	beforeIDs := func(path string) map[string]string {
		afterMap := idsFor(path)
		out := map[string]string{}
		for _, id := range afterMap {
			out[id] = id
		}
		return out
	}
	before, after := entryFindings(e.Before, bp), entryFindings(e.After, ap)
	af := map[string]comparisonFinding{}
	// Map paired variant positions, so reordered enumeration compares each
	// candidate to its established selection rather than its array ordinal.
	pairedPath := func(p string) string {
		bestB, bestA := "", ""
		for _, pair := range e.Pairs {
			if (pair.Basis == "canonical_resolution_selection" || pair.Basis == "captured_definition_digest_invocation_sources") && strings.HasPrefix(p, pair.BeforePointer+"/") && len(pair.BeforePointer) > len(bestB) {
				bestB, bestA = pair.BeforePointer, pair.AfterPointer
			}
		}
		if bestB != "" {
			return bestA + strings.TrimPrefix(p, bestB)
		}
		return ap + strings.TrimPrefix(p, bp)
	}
	for _, f := range after {
		af[f.pointer] = f
	}
	definite := false
	emit := func(b, a *comparisonFinding) {
		f := a
		if f == nil {
			f = b
		}
		d := impact.EvidenceDelta{Category: f.category, Change: "changed"}
		if b != nil && a != nil {
			e.Pairs = append(e.Pairs, EvidencePair{BeforePointer: b.pointer, AfterPointer: a.pointer, Basis: "same_query_owner_finding"})
		}
		if b != nil {
			d.Before, _ = json.Marshal(b.value)
			d.Key = b.pointer
		} else {
			d.Change = "introduced"
		}
		if a != nil {
			d.After, _ = json.Marshal(a.value)
			if b == nil {
				d.Key = a.pointer
			}
		} else {
			d.Change = "resolved"
		}
		e.Deltas = append(e.Deltas, d)
	}
	for _, b := range before {
		path := pairedPath(b.pointer)
		a, ok := af[path]
		if !ok {
			emit(&b, nil)
			definite = true
			continue
		}
		delete(af, path)
		bn := normalizeFinding(b.value, beforeIDs, pairedPath(b.pointer), false)
		an := normalizeFinding(a.value, idsFor, a.pointer, false)
		if strings.HasSuffix(b.pointer, "/resolution/variants") {
			bn = variantMembership(b.value)
			an = variantMembership(a.value)
		}
		if reflect.DeepEqual(bn, an) {
			continue
		}
		emit(&b, &a)
		// An ID-only difference without unique correspondence is uncertainty.
		semanticChange := !reflect.DeepEqual(normalizeFinding(b.value, beforeIDs, pairedPath(b.pointer), true), normalizeFinding(a.value, idsFor, a.pointer, true))
		if !semanticChange {
			e.Unmatched = append(e.Unmatched, b.pointer)
		}
		unresolvedDefinition := false
		for _, p := range append(append([]string{}, e.Unmatched...), e.Ambiguous...) {
			if strings.Contains(p, "/closure/definition_analyses/") {
				unresolvedDefinition = true
			}
		}
		definitionContent := strings.Contains(b.pointer, "/definition_analyses/") || strings.HasSuffix(b.pointer, "/effective_analysis")
		if semanticChange && !(b.category == "closure" && definitionContent && unresolvedDefinition) {
			definite = true
		}
	}
	for _, a := range after {
		if _, ok := af[a.pointer]; ok {
			emit(nil, &a)
			definite = true
		}
	}
	return definite
}
func addSituatedIDs(b, a map[string]any, domain string, mappings map[string]map[string]string) {
	if mappings[domain] == nil {
		mappings[domain] = map[string]string{}
	}
	for _, name := range []string{"stages", "scopes"} {
		bm, am := map[string][]string{}, map[string][]string{}
		for _, side := range []struct {
			m   map[string]any
			out map[string][]string
		}{{b, bm}, {a, am}} {
			vs, _ := side.m[name].([]any)
			for _, v := range vs {
				m := v.(map[string]any)
				key := factKey([]any{m["location"], m["kind"], m["command"]})
				id, _ := m["id"].(string)
				side.out[key] = append(side.out[key], id)
			}
		}
		for key, ids := range bm {
			if len(ids) == 1 && len(am[key]) == 1 {
				mappings[domain][am[key][0]] = ids[0]
			}
		}
	}
}

func variantMembership(value any) []string {
	out := []string{}
	raw, _ := json.Marshal(value)
	var variants []any
	_ = json.Unmarshal(raw, &variants)
	for _, v := range variants {
		out = append(out, factKey(v.(map[string]any)["selection"]))
	}
	sort.Strings(out)
	return out
}
