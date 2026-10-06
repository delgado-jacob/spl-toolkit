package compatibility

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"

	"github.com/delgado-jacob/spl-toolkit/pkg/environment"
)

func stableKey(value any) string { raw, _ := json.Marshal(value); return string(raw) }
func sortReasons(reasons []Reason) {
	for i := range reasons {
		r := &reasons[i]
		sort.Strings(r.ReferenceIDs)
		sort.Strings(r.CandidateInputIDs)
		sort.Slice(r.Locations, func(i, j int) bool { return stableKey(r.Locations[i]) < stableKey(r.Locations[j]) })
	}
	sort.SliceStable(reasons, func(i, j int) bool { return stableKey(reasons[i]) < stableKey(reasons[j]) })
}
func sortObjects(objects []ObjectEvidence) {
	sort.Slice(objects, func(i, j int) bool { return stableKey(objects[i]) < stableKey(objects[j]) })
}
func finalizeReport(report *Report) {
	for i := range report.Inputs {
		input := &report.Inputs[i]
		sortObjects(input.Objects)
		sort.Strings(input.RequirementIDs)
		sortReasons(input.Reasons)
	}
	sort.Slice(report.Inputs, func(i, j int) bool { return report.Inputs[i].InputID < report.Inputs[j].InputID })
	for i := range report.RequirementOutcomes {
		out := &report.RequirementOutcomes[i]
		sortReasons(out.Reasons)
		sortObjects(out.Objects)
		sort.Slice(out.Schemas, func(i, j int) bool { return stableKey(out.Schemas[i]) < stableKey(out.Schemas[j]) })
		sort.Slice(out.Capabilities, func(i, j int) bool { return stableKey(out.Capabilities[i]) < stableKey(out.Capabilities[j]) })
		sort.Slice(out.SourceIntervals, func(i, j int) bool { return stableKey(out.SourceIntervals[i]) < stableKey(out.SourceIntervals[j]) })
	}
	sort.Slice(report.RequirementOutcomes, func(i, j int) bool {
		a, b := report.RequirementOutcomes[i], report.RequirementOutcomes[j]
		if stableKey(a.Query) != stableKey(b.Query) {
			return stableKey(a.Query) < stableKey(b.Query)
		}
		if a.RequirementID != b.RequirementID {
			return a.RequirementID < b.RequirementID
		}
		return a.InputID < b.InputID
	})
	for i := range report.Coverage {
		sortReasons(report.Coverage[i].Reasons)
	}
	sort.Slice(report.Coverage, func(i, j int) bool { return stableKey(report.Coverage[i]) < stableKey(report.Coverage[j]) })
	sort.Slice(report.Diagnostics, func(i, j int) bool { return stableKey(report.Diagnostics[i]) < stableKey(report.Diagnostics[j]) })
	sortReasons(report.Reasons)
	sort.Slice(report.InputBindings, func(i, j int) bool { return report.InputBindings[i].InputID < report.InputBindings[j].InputID })
	// Struct field order is explicit and stable; selectors and input bindings are
	// normalized before hashing. No mutable query state enters Prepared.
	identity := struct {
		QueryDigest             string                   `json:"query_digest"`
		SourceID                string                   `json:"source_id"`
		CapabilityRevision      string                   `json:"capability_revision"`
		AnalysisContractVersion int                      `json:"analysis_contract_version"`
		RequirementSetVersion   int                      `json:"requirement_set_version"`
		EnvironmentDigest       string                   `json:"environment_digest"`
		SchemaBundleDigest      string                   `json:"schema_bundle_digest,omitempty"`
		QueryScope              environment.CaptureScope `json:"query_scope"`
		InputBindings           []InputBinding           `json:"input_bindings"`
	}{report.Provenance.QueryDigest, report.Provenance.SourceID, report.Provenance.CapabilityRevision, report.Provenance.AnalysisContractVersion, report.Provenance.RequirementSetVersion, report.Provenance.EnvironmentDigest, report.Provenance.SchemaBundleDigest, report.QueryScope, report.InputBindings}
	sum := sha256.Sum256([]byte(stableKey(identity)))
	report.Provenance.AssessmentIdentityDigest = "sha256:" + hex.EncodeToString(sum[:])
}
