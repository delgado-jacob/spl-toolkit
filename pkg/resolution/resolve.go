package resolution

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/closure"
	"github.com/delgado-jacob/spl-toolkit/pkg/compatibility"
	"github.com/delgado-jacob/spl-toolkit/pkg/environment"
)

// ResolveJSON performs strict wire admission before composing variants.
func ResolveJSON(raw []byte) (*Report, error) {
	request, err := DecodeRequest(raw)
	if err != nil {
		return nil, err
	}
	return Resolve(request)
}

// Resolve prepares detached environment evidence and evaluates every admitted selection.
func Resolve(request Request) (*Report, error) {
	request, err := normalizeRequestShape(request)
	if err != nil {
		return nil, err
	}
	prepared, err := Prepare(request.Compatibility.Snapshot, request.Compatibility.SchemaBundle)
	if err != nil {
		if detail, ok := compatibility.RequestErrorDetails(err); ok {
			if detail.ByteOffset != nil {
				return nil, requestErrorOffset("request_invalid", "/compatibility"+detail.Path, detail.Message, *detail.ByteOffset)
			}
			return nil, requestErrorAt("request_invalid", "/compatibility"+detail.Path, detail.Message)
		}
		return nil, requestErrorAt("assessment_invariant", "/compatibility", err.Error())
	}
	return prepared.Resolve(PreparedRequest{SchemaVersion: request.SchemaVersion, Document: request.Document, Resolutions: request.Resolutions, MaxVariants: request.MaxVariants, Compatibility: compatibility.ResolutionAssessment{QueryScope: request.Compatibility.QueryScope, InputBindings: request.Compatibility.InputBindings, DependencyBindings: request.Compatibility.DependencyBindings}})
}

func (p *Prepared) Resolve(request PreparedRequest) (*Report, error) {
	if p == nil || p.compatibility == nil {
		return nil, requestErrorAt("request_invalid", "", "prepared environment is required")
	}
	request, err := normalizePreparedRequest(request)
	if err != nil {
		return nil, err
	}
	count, limit, err := admitFanout(request.Resolutions, request.MaxVariants)
	if err != nil {
		return nil, err
	}
	session, err := analysis.PrepareResolution(request.Document)
	if err != nil {
		return nil, requestErrorAt("analysis_error", "/document", err.Error())
	}
	original := session.Evidence()
	flattened := []analysis.ResolutionChoice{}
	selected := map[string]string{}
	for _, r := range request.Resolutions {
		selected[r.Placeholder] = r.Kind
		for _, value := range r.Values {
			flattened = append(flattened, analysis.ResolutionChoice{Placeholder: r.Placeholder, Kind: r.Kind, Value: value})
		}
	}
	known := map[string]string{}
	for _, group := range original.Placeholders {
		if prior, ok := known[group.Placeholder]; ok && prior != group.Kind {
			return nil, requestErrorAt("resolution_kind_mismatch", "/resolutions", "placeholder has conflicting kinds")
		}
		known[group.Placeholder] = group.Kind
		if kind, exists := selected[group.Placeholder]; !exists {
			return nil, requestErrorAt("resolution_missing", "/resolutions", "every eligible placeholder requires a resolution")
		} else if kind != group.Kind {
			return nil, requestErrorAt("resolution_kind_mismatch", "/resolutions", "resolution kind disagrees with eligible placeholder")
		}
	}
	for _, resolution := range request.Resolutions {
		if _, exists := known[resolution.Placeholder]; !exists && coverageComplete(original.Coverage) {
			return nil, requestErrorAt("resolution_unused", "/resolutions", "resolution has no eligible placeholder")
		}
	}
	if err := p.compatibility.ValidateResolutionBindings(original, flattened, request.Compatibility); err != nil {
		if detail, ok := compatibility.RequestErrorDetails(err); ok {
			return nil, requestErrorAt(detail.Code, "/compatibility"+detail.Path, detail.Message)
		}
		return nil, requestErrorAt("assessment_invariant", "/compatibility", err.Error())
	}
	resolutionDigest, err := digestJSON(struct {
		Resolutions []Resolution `json:"resolutions"`
		MaxVariants uint64       `json:"max_variants"`
	}{request.Resolutions, limit})
	if err != nil {
		return nil, requestErrorAt("resolution_invariant", "/resolutions", err.Error())
	}
	assessmentIdentity, err := assessmentDigest(request.Compatibility)
	if err != nil {
		return nil, requestErrorAt("assessment_invariant", "/compatibility", err.Error())
	}
	provenance := Provenance{QueryDigest: digestBytes([]byte(request.Document.Text)), CapabilityRevision: original.Analysis.Requirements.CapabilityRevision, AnalysisContractVersion: original.Analysis.SchemaVersion, RequirementSetVersion: original.Analysis.Requirements.SchemaVersion, EnvironmentDigest: p.identity.EnvironmentDigest, SchemaBundleDigest: p.identity.SchemaBundleDigest, ResolutionInputDigest: resolutionDigest, AssessmentInputDigest: assessmentIdentity}

	report := &Report{SchemaVersion: 1, Original: original, Resolutions: detach(request.Resolutions), MaxVariants: limit, TotalCombinations: count.String(), Variants: []Variant{}, Provenance: provenance}
	err = visitSelections(request.Resolutions, func(ordinal uint64, selection []analysis.ResolutionChoice) error {
		id, err := digestJSON(struct {
			Query     analysis.RequirementQueryIdentity `json:"query"`
			Selection []analysis.ResolutionChoice       `json:"selection"`
			Ordinal   uint64                            `json:"ordinal"`
		}{original.Analysis.Requirements.Query, selection, ordinal})
		if err != nil {
			return requestErrorAt("resolution_invariant", "/resolutions", err.Error())
		}
		variant := Variant{ID: id, Ordinal: ordinal, Selection: selection, Outcome: "incomplete", Changes: []analysis.ResolutionChange{}, Proof: analysis.ResolutionProofEvidence{References: []analysis.ResolutionReferencePair{}, Roles: []analysis.ResolutionRole{}, Limitations: []analysis.ResolutionLimitation{}}, Diagnostics: []Diagnostic{}, Provenance: provenance}

		for _, choice := range selection {
			if _, exists := known[choice.Placeholder]; !exists {
				variant.Diagnostics = append(variant.Diagnostics, Diagnostic{Code: "resolution_discovery_incomplete", Message: "Partial discovery cannot establish a rendering owner for this choice", Placeholder: choice.Placeholder})
			}
		}
		ownedSelection := []analysis.ResolutionChoice{}
		for _, choice := range selection {
			if _, exists := known[choice.Placeholder]; exists {
				ownedSelection = append(ownedSelection, choice)
			}
		}
		rendering, err := session.Render(ownedSelection)
		if err != nil {
			return requestErrorAt("resolution_invariant", "/resolutions", err.Error())
		}
		candidate := rendering.CandidateDocument()
		variant.CandidateText = &candidate.Text
		variant.Provenance.QueryDigest = digestBytes([]byte(candidate.Text))
		proof, err := session.Verify(rendering)
		if err != nil {
			return requestErrorAt("analysis_error", "/document", err.Error())
		}
		variant.Proof = proof.Evidence()
		for _, limitation := range variant.Proof.Limitations {
			loc := limitation.Location
			variant.Diagnostics = append(variant.Diagnostics, Diagnostic{Code: limitation.Code, Message: limitation.Message, Location: &loc})
		}
		_, evidence, _, authorized := proof.AssessmentEvidence()
		variant.CandidateAnalysis = &evidence.Analysis
		variant.Changes = rendering.Changes()
		if authorized {
			assessment, err := p.compatibility.CheckResolution(proof, request.Compatibility)
			if err != nil {
				if detail, ok := compatibility.RequestErrorDetails(err); ok {
					return requestErrorAt(detail.Code, "/compatibility"+detail.Path, detail.Message)
				} else {
					return requestErrorAt("assessment_invariant", "/compatibility", err.Error())
				}
			} else {
				variant.Compatibility = assessment
			}
		}
		complete := coverageComplete(original.Coverage) && coverageComplete(evidence.Coverage) && !hasExternalMarker(evidence.Analysis)
		if variant.Compatibility != nil && variant.Compatibility.Closure != nil {
			closure := variant.Compatibility.Closure
			if closure.EffectiveAnalysis != nil && hasExternalMarker(*closure.EffectiveAnalysis) {
				complete = false
			}
			for _, definition := range closure.DefinitionAnalyses {
				if definition.DirectAnalysis != nil && hasExternalMarker(*definition.DirectAnalysis) {
					complete = false
				}
				if definition.EffectiveAnalysis != nil && hasExternalMarker(*definition.EffectiveAnalysis) {
					complete = false
				}
			}
		}
		if !complete {
			variant.Diagnostics = append(variant.Diagnostics, Diagnostic{Code: "substitution_incomplete", Message: "External marker substitution coverage is incomplete"})
		}
		definitive := false
		for _, limitation := range variant.Proof.Limitations {
			if limitation.Code == "target_not_renderable" || limitation.Code == "render_conflict" {
				definitive = true
			}
		}
		if evidence.Analysis.Status == analysis.Invalid || definitive || (variant.Compatibility != nil && variant.Compatibility.Outcome == "unsatisfied") {
			variant.Outcome = "failed"
		} else if variant.Proof.Proven && complete && variant.Compatibility != nil && variant.Compatibility.Outcome == "satisfied" {
			variant.Outcome = "verified"
			published := candidate.Text
			variant.ResolvedQuery = &published
		}
		switch variant.Outcome {
		case "verified":
			report.Counts.Verified++
		case "failed":
			report.Counts.Failed++
		default:
			report.Counts.Incomplete++
		}
		report.Variants = append(report.Variants, variant)
		report.GeneratedCount++
		return nil
	})
	if err != nil {
		return nil, err
	}
	return report, nil
}

func digestBytes(raw []byte) string {
	sum := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(sum[:])
}
func digestJSON(value any) (string, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return digestBytes(raw), nil
}
func coverageComplete(coverage analysis.InputCoverage) bool {
	return coverage.State == "complete" || coverage.State == "not_applicable"
}

// Use decoded canonical identities. Source strings, field literals, and local
// view symbols are not external substitution obligations.
func hasExternalMarker(result analysis.Result) bool {
	for _, input := range result.Inputs {
		if markerPattern.MatchString(input.Identity.Value) {
			return true
		}
	}
	for _, ref := range result.References {
		if supportedChoiceKind(ref.Kind) && ref.Binding != "local_view" && ref.Binding != "local" && markerPattern.MatchString(ref.NormalizedName) {
			return true
		}
	}
	return false
}

// Binding order does not alter the assessment projection. Input and dependency
// selections remain separate from the normalized query scope and artifacts.
func assessmentDigest(assessment compatibility.ResolutionAssessment) (string, error) {
	assessment = detach(assessment)
	if assessment.DependencyBindings == nil {
		assessment.DependencyBindings = []closure.Binding{}
	}
	sort.Slice(assessment.InputBindings, func(i, j int) bool {
		a, b := assessment.InputBindings[i], assessment.InputBindings[j]
		if a.OriginalInputID != b.OriginalInputID {
			return a.OriginalInputID < b.OriginalInputID
		}
		if a.ResolvedValue == nil || b.ResolvedValue == nil {
			return a.ResolvedValue == nil && b.ResolvedValue != nil
		}
		return *a.ResolvedValue < *b.ResolvedValue
	})
	sort.Slice(assessment.DependencyBindings, func(i, j int) bool {
		a, b := assessment.DependencyBindings[i], assessment.DependencyBindings[j]
		if a.DocumentDigest != b.DocumentDigest {
			return a.DocumentDigest < b.DocumentDigest
		}
		if a.Kind != b.Kind {
			return a.Kind < b.Kind
		}
		if a.Start != b.Start {
			return a.Start < b.Start
		}
		return a.End < b.End
	})
	return digestJSON(struct {
		QueryScope         environment.CaptureScope          `json:"query_scope"`
		InputBindings      []compatibility.ResolutionBinding `json:"input_bindings"`
		DependencyBindings []closure.Binding                 `json:"dependency_bindings"`
	}{assessment.QueryScope, assessment.InputBindings, assessment.DependencyBindings})
}
