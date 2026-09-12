package impact

import (
	"encoding/json"
	"reflect"
	"sort"
	"strings"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/corpus"
	"github.com/delgado-jacob/spl-toolkit/pkg/rewrite"
)

type outcomeEvidence struct {
	referenceID string
	value       any
}
type evidence struct {
	analysis      *analysis.Result
	status        analysis.Status
	complete      bool
	coverage      any
	outcomes      []outcomeEvidence
	diagnostics   []analysis.Diagnostic
	dependencies  analysis.Dependencies
	changes       []rewrite.Change
	rules         []rewrite.RuleEvaluation
	candidateText string
	mapping       bool
}

func evidenceFromEvaluation(v *corpus.Evaluation) evidence {
	if v.FieldValidation != nil {
		r := v.FieldValidation
		e := evidence{analysis: r.Analysis, status: r.Status, complete: r.Coverage.SyntaxComplete && r.Coverage.SemanticComplete && r.Coverage.SchemaComplete, coverage: r.Coverage, diagnostics: r.Diagnostics, dependencies: r.Analysis.Dependencies}
		for _, o := range r.Outcomes {
			e.outcomes = append(e.outcomes, outcomeEvidence{o.ReferenceID, o})
		}
		return e
	}
	r := v.SchemaValidation
	e := evidence{analysis: r.Analysis, status: r.Status, complete: r.Coverage.SyntaxComplete && r.Coverage.SemanticComplete && r.Coverage.SchemaComplete, coverage: r.Coverage, diagnostics: r.Diagnostics, dependencies: r.Analysis.Dependencies}
	for _, o := range r.Outcomes {
		e.outcomes = append(e.outcomes, outcomeEvidence{o.ReferenceID, o})
	}
	return e
}

func evidenceFromRewrite(r *rewrite.Result) evidence {
	e := evidence{analysis: r.CandidateAnalysis, status: r.Status, complete: r.Coverage.SyntaxComplete && r.Coverage.SemanticComplete && r.Coverage.RewriteComplete,
		coverage: r.Coverage, diagnostics: r.CandidateAnalysis.Diagnostics, dependencies: r.CandidateAnalysis.Dependencies,
		changes: r.Changes, rules: r.RuleEvaluations, candidateText: r.CandidateText, mapping: true}
	if v := r.CandidateValidation; v != nil {
		if v.FieldList != nil {
			e.complete = e.complete && v.FieldList.Coverage.SchemaComplete
			e.diagnostics = v.FieldList.Diagnostics
			for _, o := range v.FieldList.Outcomes {
				e.outcomes = append(e.outcomes, outcomeEvidence{o.ReferenceID, o})
			}
		} else if v.Schema != nil {
			e.complete = e.complete && v.Schema.Coverage.SchemaComplete
			e.diagnostics = v.Schema.Diagnostics
			for _, o := range v.Schema.Outcomes {
				e.outcomes = append(e.outcomes, outcomeEvidence{o.ReferenceID, o})
			}
		}
	}
	return e
}

func copySelection(s corpus.Selection) corpus.Selection {
	s.IgnoredNames = append([]string{}, s.IgnoredNames...)
	s.SkippedSymlinks = append([]string{}, s.SkippedSymlinks...)
	s.TraversalFailures = append([]corpus.AcquisitionError{}, s.TraversalFailures...)
	return s
}

func copyInput(in corpus.Input) corpus.Input {
	out := corpus.Input{Selection: copySelection(in.Selection), Entries: make([]corpus.Entry, len(in.Entries))}
	for i, entry := range in.Entries {
		out.Entries[i] = entry
		if entry.Document != nil {
			doc := *entry.Document
			out.Entries[i].Document = &doc
		}
		if entry.Failure != nil {
			failure := *entry.Failure
			out.Entries[i].Failure = &failure
		}
	}
	return out
}

// Original reference alignment requires the exact same analyzed source revision.
// A changed canonical ID can be paired only by unique, identical source evidence.
func alignOriginal(before, after *analysis.Result, beforeRevision, afterRevision string) Alignment {
	a := Alignment{Pairs: []ReferencePair{}, Unmatched: []string{}, Ambiguous: []string{}}
	if before == nil || after == nil || beforeRevision == "" || beforeRevision != afterRevision {
		if before != nil {
			for _, r := range before.References {
				a.Unmatched = append(a.Unmatched, "before:"+r.ID)
			}
		}
		if after != nil {
			for _, r := range after.References {
				a.Unmatched = append(a.Unmatched, "after:"+r.ID)
			}
		}
		return a
	}
	beforeIDs, afterIDs := map[string][]int{}, map[string][]int{}
	for i, r := range before.References {
		beforeIDs[r.ID] = append(beforeIDs[r.ID], i)
	}
	for i, r := range after.References {
		afterIDs[r.ID] = append(afterIDs[r.ID], i)
	}
	usedBefore, usedAfter := map[int]bool{}, map[int]bool{}
	// Reserve every exact identity before fallback can consume its candidate.
	for i, b := range before.References {
		matches := afterIDs[b.ID]
		if len(beforeIDs[b.ID]) != 1 || len(matches) != 1 {
			continue
		}
		j := matches[0]
		r := after.References[j]
		if referenceEvidence(b) == referenceEvidence(r) {
			a.Pairs = append(a.Pairs, ReferencePair{BeforeID: b.ID, AfterID: r.ID, Domain: "original", Basis: "canonical_original_id"})
			usedBefore[i], usedAfter[j] = true, true
		}
	}
	beforeEvidence, afterEvidence := map[string][]int{}, map[string][]int{}
	for i, r := range before.References {
		if !usedBefore[i] {
			key := referenceEvidence(r)
			beforeEvidence[key] = append(beforeEvidence[key], i)
		}
	}
	for i, r := range after.References {
		if !usedAfter[i] {
			key := referenceEvidence(r)
			afterEvidence[key] = append(afterEvidence[key], i)
		}
	}
	for key, bs := range beforeEvidence {
		as := afterEvidence[key]
		if len(bs) == 1 && len(as) == 1 {
			b, r := before.References[bs[0]], after.References[as[0]]
			a.Pairs = append(a.Pairs, ReferencePair{BeforeID: b.ID, AfterID: r.ID, Domain: "original", Basis: "unique_exact_original_evidence"})
			usedBefore[bs[0]], usedAfter[as[0]] = true, true
		} else if len(as) != 0 {
			for _, i := range bs {
				a.Ambiguous = append(a.Ambiguous, "before:"+before.References[i].ID)
			}
			for _, i := range as {
				a.Ambiguous = append(a.Ambiguous, "after:"+after.References[i].ID)
			}
		}
	}
	for i, r := range before.References {
		if !usedBefore[i] {
			a.Unmatched = append(a.Unmatched, "before:"+r.ID)
		}
	}
	for i, r := range after.References {
		if !usedAfter[i] {
			a.Unmatched = append(a.Unmatched, "after:"+r.ID)
		}
	}
	sortAlignment(&a)
	return a
}

func referenceEvidence(r analysis.Reference) string {
	raw, _ := json.Marshal([]any{r.Kind, r.Role, r.ScopeID, r.StageID, r.Location, r.OriginalName, r.NormalizedName})
	return string(raw)
}

func alignMapping(before, after *rewrite.Result, revision string) Alignment {
	a := alignOriginal(before.OriginalAnalysis, after.OriginalAnalysis, revision, revision)
	for _, side := range []*rewrite.Result{before, after} {
		for _, change := range side.Changes {
			if change.Outcome == "ambiguous" || len(change.OriginalReferenceIDs) > 1 || len(change.CandidateReferenceIDs) > 1 {
				a.Ambiguous = append(a.Ambiguous, "group:"+change.GroupID)
			}
		}
	}
	if before.CandidateText == after.CandidateText {
		beforeCandidateRevision, beforeErr := corpus.AnalysisRevision(before.CandidateAnalysis.Document)
		afterCandidateRevision, afterErr := corpus.AnalysisRevision(after.CandidateAnalysis.Document)
		if beforeErr != nil || afterErr != nil {
			beforeCandidateRevision, afterCandidateRevision = "", ""
		}
		candidate := alignOriginal(before.CandidateAnalysis, after.CandidateAnalysis, beforeCandidateRevision, afterCandidateRevision)
		for _, pair := range candidate.Pairs {
			pair.Domain = "candidate"
			pair.Basis = "same_candidate_source_" + pair.Basis
			a.Pairs = append(a.Pairs, pair)
		}
		a.Ambiguous = append(a.Ambiguous, candidate.Ambiguous...)
	} else {
		bmap := provenCandidateOrigins(before)
		amap := provenCandidateOrigins(after)
		candidates := []ReferencePair{}
		beforeCount, afterCount := map[string]int{}, map[string]int{}
		for original, b := range bmap {
			if candidate, ok := amap[original]; ok {
				candidates = append(candidates, ReferencePair{BeforeID: b, AfterID: candidate, Domain: "candidate", Basis: "audit_original_reference_provenance"})
				beforeCount[b]++
				afterCount[candidate]++
			}
		}
		for _, pair := range candidates {
			if beforeCount[pair.BeforeID] != 1 || afterCount[pair.AfterID] != 1 {
				a.Ambiguous = append(a.Ambiguous, "candidate:"+pair.BeforeID+":"+pair.AfterID)
				continue
			}
			a.Pairs = append(a.Pairs, pair)
		}
	}
	pairedBefore, pairedAfter := map[string]bool{}, map[string]bool{}
	for _, pair := range a.Pairs {
		if pair.Domain == "candidate" {
			pairedBefore[pair.BeforeID], pairedAfter[pair.AfterID] = true, true
		}
	}
	for _, ref := range before.CandidateAnalysis.References {
		if !pairedBefore[ref.ID] {
			a.Unmatched = append(a.Unmatched, "candidate:before:"+ref.ID)
		}
	}
	for _, ref := range after.CandidateAnalysis.References {
		if !pairedAfter[ref.ID] {
			a.Unmatched = append(a.Unmatched, "candidate:after:"+ref.ID)
		}
	}
	sortAlignment(&a)
	return a
}

func provenCandidateOrigins(r *rewrite.Result) map[string]string {
	out := map[string]string{}
	conflicted := map[string]bool{}
	if r.CandidateText == r.OriginalText {
		aligned := alignOriginal(r.OriginalAnalysis, r.CandidateAnalysis, "same", "same")
		for _, p := range aligned.Pairs {
			out[p.BeforeID] = p.AfterID
		}
	}
	for _, change := range r.Changes {
		if len(change.OriginalReferenceIDs) != 1 || len(change.CandidateReferenceIDs) != 1 || !change.CandidateApplied {
			continue
		}
		original, candidate := change.OriginalReferenceIDs[0], change.CandidateReferenceIDs[0]
		if !hasReference(r.OriginalAnalysis, original) || !hasReference(r.CandidateAnalysis, candidate) {
			continue
		}
		if conflicted[original] {
			continue
		}
		if prior, ok := out[original]; ok && prior != candidate {
			delete(out, original)
			conflicted[original] = true
			continue
		}
		out[original] = candidate
	}
	return out
}

func hasReference(r *analysis.Result, id string) bool {
	if r == nil {
		return false
	}
	for _, ref := range r.References {
		if ref.ID == id {
			return true
		}
	}
	return false
}

func sortAlignment(a *Alignment) {
	sort.Slice(a.Pairs, func(i, j int) bool {
		if a.Pairs[i].Domain != a.Pairs[j].Domain {
			return a.Pairs[i].Domain < a.Pairs[j].Domain
		}
		if a.Pairs[i].BeforeID != a.Pairs[j].BeforeID {
			return a.Pairs[i].BeforeID < a.Pairs[j].BeforeID
		}
		return a.Pairs[i].AfterID < a.Pairs[j].AfterID
	})
	sort.Strings(a.Unmatched)
	sort.Strings(a.Ambiguous)
	a.Unmatched = uniqueStrings(a.Unmatched)
	a.Ambiguous = uniqueStrings(a.Ambiguous)
}

func uniqueStrings(xs []string) []string {
	if len(xs) == 0 {
		return xs
	}
	out := xs[:1]
	for _, x := range xs[1:] {
		if x != out[len(out)-1] {
			out = append(out, x)
		}
	}
	return out
}

func pairMap(a Alignment, domain string) map[string]string {
	m := map[string]string{}
	for _, p := range a.Pairs {
		if p.Domain == domain {
			m[p.BeforeID] = p.AfterID
		}
	}
	return m
}

func canonicalRaw(v any) json.RawMessage { raw, _ := json.Marshal(v); return raw }

func compareEvidence(before, after evidence, alignment Alignment) []EvidenceDelta {
	deltas := []EvidenceDelta{}
	addScalar := func(category string, b, a any) {
		br, ar := canonicalRaw(b), canonicalRaw(a)
		if string(br) != string(ar) {
			deltas = append(deltas, EvidenceDelta{Category: category, Change: "changed", Key: category, Before: br, After: ar})
		}
	}
	addScalar("status", before.status, after.status)
	addScalar("coverage", before.coverage, after.coverage)
	if before.mapping || after.mapping {
		addScalar("candidate_text", before.candidateText, after.candidateText)
	}
	deltas = append(deltas, dependencyDeltas(before.dependencies, after.dependencies)...)
	domain := "original"
	if before.mapping || after.mapping {
		domain = "candidate"
	}
	deltas = append(deltas, outcomeDeltas(before.outcomes, after.outcomes, alignment, domain)...)
	deltas = append(deltas, diagnosticDeltas(before.diagnostics, after.diagnostics, before.analysis, after.analysis, alignment, domain)...)
	if before.mapping || after.mapping {
		deltas = append(deltas, keyedDeltas("rewrite_change", changesAsFindings(before.changes), changesAsFindings(after.changes))...)
		deltas = append(deltas, ruleEvaluationDeltas(before.rules, after.rules, alignment)...)
	}
	sort.Slice(deltas, func(i, j int) bool {
		a, b := deltas[i], deltas[j]
		if a.Category != b.Category {
			return a.Category < b.Category
		}
		if a.Key != b.Key {
			return a.Key < b.Key
		}
		if a.Change != b.Change {
			return a.Change < b.Change
		}
		return string(a.Before)+string(a.After) < string(b.Before)+string(b.After)
	})
	return deltas
}

type finding struct {
	key   string
	value any
}

func outcomeDeltas(before, after []outcomeEvidence, a Alignment, domain string) []EvidenceDelta {
	b, af := map[string][]json.RawMessage{}, map[string][]json.RawMessage{}
	for _, o := range before {
		b[o.referenceID] = append(b[o.referenceID], canonicalRaw(o.value))
	}
	for _, o := range after {
		af[o.referenceID] = append(af[o.referenceID], canonicalRaw(o.value))
	}
	out := []EvidenceDelta{}
	usedBefore, usedAfter := map[string]bool{}, map[string]bool{}
	for _, pair := range a.Pairs {
		if pair.Domain != domain {
			continue
		}
		bv, av := b[pair.BeforeID], af[pair.AfterID]
		if len(bv) == 0 && len(av) == 0 {
			continue
		}
		usedBefore[pair.BeforeID], usedAfter[pair.AfterID] = true, true
		if len(bv) > 1 || len(av) > 1 {
			for _, value := range bv {
				out = append(out, EvidenceDelta{Category: "outcome", Change: "ambiguous", Key: pair.BeforeID, Before: value})
			}
			for _, value := range av {
				out = append(out, EvidenceDelta{Category: "outcome", Change: "ambiguous", Key: pair.BeforeID, After: value})
			}
		} else if len(bv) == 0 {
			out = append(out, EvidenceDelta{Category: "outcome", Change: "introduced", Key: pair.BeforeID, After: av[0]})
		} else if len(av) == 0 {
			out = append(out, EvidenceDelta{Category: "outcome", Change: "resolved", Key: pair.BeforeID, Before: bv[0]})
		} else if !sameOutcome(bv[0], av[0]) {
			out = append(out, EvidenceDelta{Category: "outcome", Change: "changed", Key: pair.BeforeID, Before: bv[0], After: av[0]})
		}
	}
	for id, values := range b {
		if !usedBefore[id] {
			for _, value := range values {
				out = append(out, EvidenceDelta{Category: "outcome", Change: "unmatched", Key: "before:" + id, Before: value})
			}
		}
	}
	for id, values := range af {
		if !usedAfter[id] {
			for _, value := range values {
				out = append(out, EvidenceDelta{Category: "outcome", Change: "unmatched", Key: "after:" + id, After: value})
			}
		}
	}
	return out
}

func sameOutcome(before, after json.RawMessage) bool {
	var b, a map[string]json.RawMessage
	if json.Unmarshal(before, &b) != nil || json.Unmarshal(after, &a) != nil {
		return false
	}
	delete(b, "reference_id")
	delete(a, "reference_id")
	return reflect.DeepEqual(b, a)
}

func diagnosticDeltas(before, after []analysis.Diagnostic, bAnalysis, aAnalysis *analysis.Result, a Alignment, domain string) []EvidenceDelta {
	pairs := pairMap(a, domain)
	reverse := map[string]string{}
	for bID, aID := range pairs {
		reverse[aID] = bID
	}
	b, af := []finding{}, []finding{}
	out := []EvidenceDelta{}
	for _, d := range before {
		ref := diagnosticReference(d, bAnalysis)
		if _, ok := pairs[ref]; ref == "" || !ok {
			out = append(out, EvidenceDelta{Category: "diagnostic", Change: "unmatched", Key: "before:" + d.Code + "|" + d.Category + "|" + ref, Before: canonicalRaw(d)})
			continue
		}
		b = append(b, finding{d.Code + "|" + d.Category + "|" + ref, d})
	}
	for _, d := range after {
		ref := diagnosticReference(d, aAnalysis)
		if bID, ok := reverse[ref]; ok && ref != "" {
			af = append(af, finding{d.Code + "|" + d.Category + "|" + bID, d})
		} else {
			out = append(out, EvidenceDelta{Category: "diagnostic", Change: "unmatched", Key: "after:" + d.Code + "|" + d.Category + "|" + ref, After: canonicalRaw(d)})
		}
	}
	return append(out, keyedDeltas("diagnostic", b, af)...)
}

func diagnosticReference(d analysis.Diagnostic, r *analysis.Result) string {
	if r == nil {
		return ""
	}
	var found string
	for _, ref := range r.References {
		if ref.Location != d.Location || (d.ScopeID != "" && d.ScopeID != ref.ScopeID) || (d.StageID != "" && d.StageID != ref.StageID) {
			continue
		}
		if found != "" {
			return ""
		}
		found = ref.ID
	}
	return found
}

func dependencyDeltas(b, a analysis.Dependencies) []EvidenceDelta {
	before, after := []finding{}, []finding{}
	add := func(dst *[]finding, d analysis.Dependencies) {
		for _, group := range []struct {
			kind  string
			names []string
		}{{"index", d.Indexes}, {"source", d.Sources}, {"sourcetype", d.SourceTypes}, {"dataset", d.Datasets}, {"lookup", d.Lookups}, {"data_model", d.DataModels}, {"macro", d.Macros}} {
			for _, name := range group.names {
				*dst = append(*dst, finding{group.kind + "|" + name, map[string]string{"kind": group.kind, "name": name}})
			}
		}
	}
	add(&before, b)
	add(&after, a)
	return keyedDeltas("dependency", before, after)
}

func changesAsFindings(changes []rewrite.Change) []finding {
	out := []finding{}
	for _, c := range changes {
		out = append(out, finding{c.GroupID + "|" + strings.Join(c.RuleIDs, ",") + "|" + strings.Join(c.OriginalReferenceIDs, ","), c})
	}
	return out
}
func ruleEvaluationDeltas(before, after []rewrite.RuleEvaluation, alignment Alignment) []EvidenceDelta {
	beforeIDs, afterIDs := map[string]string{}, map[string]string{}
	for _, p := range alignment.Pairs {
		if p.Domain == "original" {
			beforeIDs[p.BeforeID], afterIDs[p.AfterID] = p.BeforeID, p.BeforeID
		}
	}
	out := []EvidenceDelta{}
	findings := func(rules []rewrite.RuleEvaluation, ids map[string]string, side string) []finding {
		values := []finding{}
		for _, r := range rules {
			refs := []string{}
			aligned := true
			for _, id := range r.ReferenceIDs {
				if original, ok := ids[id]; ok {
					refs = append(refs, original)
				} else {
					aligned = false
				}
			}
			sort.Strings(refs)
			key := string(canonicalRaw([]any{r.RuleID, refs, r.Location}))
			if !aligned {
				d := EvidenceDelta{Category: "rule_evaluation", Change: "unmatched", Key: side + ":" + string(canonicalRaw([]any{r.RuleID, r.ReferenceIDs, r.Location}))}
				if side == "before" {
					d.Before = canonicalRaw(r)
				} else {
					d.After = canonicalRaw(r)
				}
				out = append(out, d)
				continue
			}
			values = append(values, finding{key, r})
		}
		return values
	}
	b, a := findings(before, beforeIDs, "before"), findings(after, afterIDs, "after")
	return append(out, keyedDeltas("rule_evaluation", b, a)...)
}

func keyedDeltas(category string, before, after []finding) []EvidenceDelta {
	b, a := map[string][]json.RawMessage{}, map[string][]json.RawMessage{}
	for _, f := range before {
		b[f.key] = append(b[f.key], canonicalRaw(f.value))
	}
	for _, f := range after {
		a[f.key] = append(a[f.key], canonicalRaw(f.value))
	}
	keys := map[string]bool{}
	for key := range b {
		keys[key] = true
	}
	for key := range a {
		keys[key] = true
	}
	out := []EvidenceDelta{}
	for key := range keys {
		bv, av := b[key], a[key]
		if len(bv) > 1 || len(av) > 1 || (category == "diagnostic" && strings.HasSuffix(key, "|")) {
			for _, value := range bv {
				out = append(out, EvidenceDelta{Category: category, Change: "ambiguous", Key: key, Before: value})
			}
			for _, value := range av {
				out = append(out, EvidenceDelta{Category: category, Change: "ambiguous", Key: key, After: value})
			}
			continue
		}
		if len(bv) == 0 {
			out = append(out, EvidenceDelta{Category: category, Change: "introduced", Key: key, After: av[0]})
		} else if len(av) == 0 {
			out = append(out, EvidenceDelta{Category: category, Change: "resolved", Key: key, Before: bv[0]})
		} else if !reflect.DeepEqual(bv[0], av[0]) {
			out = append(out, EvidenceDelta{Category: category, Change: "changed", Key: key, Before: bv[0], After: av[0]})
		}
	}
	return out
}

func classify(deltas []EvidenceDelta, complete bool, unresolvedOnly bool, alignment Alignment) (Classification, []string) {
	if unresolvedOnly {
		return Indeterminate, []string{"rewrite_or_validation_incomplete"}
	}
	if len(deltas) != 0 {
		if hasObservedDelta(deltas) {
			reasons := []string{"observed_static_delta"}
			if !complete {
				reasons = append(reasons, "comparison_coverage_incomplete")
			}
			return Affected, reasons
		}
		return Indeterminate, []string{"alignment_ambiguous"}
	}
	if len(alignment.Ambiguous) != 0 || len(alignment.Unmatched) != 0 {
		return Indeterminate, []string{"alignment_unresolved"}
	}
	if !complete {
		return Indeterminate, []string{"comparison_coverage_incomplete"}
	}
	return Unchanged, []string{}
}

func hasObservedDelta(deltas []EvidenceDelta) bool {
	for _, d := range deltas {
		if d.Change != "ambiguous" && d.Change != "unmatched" {
			return true
		}
	}
	return false
}
