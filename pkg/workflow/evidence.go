package workflow

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/closure"
	"github.com/delgado-jacob/spl-toolkit/pkg/compatibility"
	"github.com/delgado-jacob/spl-toolkit/pkg/corpus"
	"github.com/delgado-jacob/spl-toolkit/pkg/impact"
	"github.com/delgado-jacob/spl-toolkit/pkg/resolution"
	"github.com/delgado-jacob/spl-toolkit/pkg/validation"
)

// EvidenceJSON strictly admits the wire document before projecting saved facts.
func EvidenceJSON(raw []byte) (*EvidenceReport, error) {
	if _, _, _, err := readWire(raw, reflect.TypeOf(EvidenceRequest{})); err != nil {
		return nil, safeEvidenceError(err)
	}
	var q EvidenceRequest
	if err := json.Unmarshal(raw, &q); err != nil {
		return nil, requestErrorAt("request_invalid", "", "invalid evidence request")
	}
	return Evidence(q)
}

// Evidence never analyzes, resolves, executes or restores private engine proof.
// The default projection contains only fixed vocabulary and local traversal tokens.
func Evidence(q EvidenceRequest) (*EvidenceReport, error) {
	if !validUTF8(reflect.ValueOf(q)) {
		return nil, requestErrorAt("request_invalid", "", "invalid evidence request")
	}
	if q.SchemaVersion != 1 {
		return nil, requestErrorAt("request_invalid", "/schema_version", "schema_version must be 1")
	}
	if (q.Report == nil) == (q.Comparison == nil) {
		return nil, requestErrorAt("request_invalid", "", "exactly one evidence source is required")
	}
	include, err := admitDisclosure(q.Include)
	if err != nil {
		return nil, err
	}
	raw, err := json.Marshal(q)
	if err != nil {
		return nil, requestErrorAt("request_invalid", "", "invalid evidence request")
	}
	if _, _, _, err = readWire(raw, reflect.TypeOf(q)); err != nil {
		return nil, safeEvidenceError(err)
	}
	out := &EvidenceReport{SchemaVersion: 1, Counts: map[string]uint64{}, Items: []EvidenceItem{}, Disclosure: Disclosure{Requested: []string{}, Emitted: []string{}, Omitted: []string{}, PotentiallySensitive: []string{}}}
	projector := evidenceProjector{out: out, include: include, emitted: map[string]bool{}}
	if q.Report != nil {
		if err := admitSavedReport(*q.Report, "/report"); err != nil {
			return nil, safeEvidenceError(err)
		}
		out.SourceKind = "report"
		out.SourceCIExitCode = q.Report.CIExitCode
		out.ExecutionComplete = q.Report.ExecutionComplete
		if err := evidenceCounts(q.Report.Counts, out.Counts); err != nil {
			return nil, err
		}
		for i, e := range q.Report.Entries {
			projector.entry(e, fmt.Sprintf("/report/entries/%d", i))
		}
		projector.walk(reflect.ValueOf(q.Report.Provenance), "/report/provenance", false)
		projector.walk(reflect.ValueOf(q.Report.Selection), "/report/selection", false)
	} else {
		if err := admitSavedComparison(*q.Comparison); err != nil {
			return nil, safeEvidenceError(err)
		}
		out.SourceKind = "comparison"
		out.SourceCIExitCode = q.Comparison.CIExitCode
		out.ExecutionComplete = q.Comparison.ExecutionComplete
		if err := evidenceCounts(q.Comparison.Counts, out.Counts); err != nil {
			return nil, err
		}
		for i, e := range q.Comparison.Entries {
			p := fmt.Sprintf("/comparison/entries/%d", i)
			projector.add(p, "comparison", string(e.Classification), e.Classification != impact.Failed && entryEvidenceComplete(e.Before) && entryEvidenceComplete(e.After) && len(e.Unmatched) == 0 && len(e.Ambiguous) == 0)
			for j, reason := range e.Reasons {
				pointer := fmt.Sprintf("%s/reasons/%d", p, j)
				item := projector.add(pointer, "diagnostic", "reported", false)
				item.Codes = append(item.Codes, publicCode(reason))
				projector.detail(pointer, "diagnostic_details", reason)
			}
			projector.entry(e.Before, p+"/before")
			projector.entry(e.After, p+"/after")
			// Delta payloads duplicate canonical findings and are deliberately never
			// copied wholesale. Their retained sources above own all disclosures.
			for j, d := range e.Deltas {
				projector.add(fmt.Sprintf("%s/deltas/%d", p, j), "delta", publicEnum(d.Change, "introduced", "resolved", "changed", "unmatched", "ambiguous"), false)
			}
		}
		projector.walk(reflect.ValueOf(q.Comparison.BeforeProvenance), "/comparison/before_provenance", false)
		projector.walk(reflect.ValueOf(q.Comparison.AfterProvenance), "/comparison/after_provenance", false)
	}
	for _, c := range disclosureCategories {
		if include[c] {
			out.Disclosure.Requested = append(out.Disclosure.Requested, c)
		}
		if projector.emitted[c] {
			out.Disclosure.Emitted = append(out.Disclosure.Emitted, c)
			out.Disclosure.PotentiallySensitive = append(out.Disclosure.PotentiallySensitive, c)
		} else {
			out.Disclosure.Omitted = append(out.Disclosure.Omitted, c)
		}
	}
	return detachedExport(out)
}

func evidenceCounts(counts any, out map[string]uint64) error {
	v := reflect.ValueOf(counts)
	typ := v.Type()
	// Only the two fixed, validated source count structs can supply map keys.
	if typ != reflect.TypeOf(Counts{}) && typ != reflect.TypeOf(ComparisonCounts{}) {
		return requestErrorAt("request_invalid", "/counts", "invalid evidence counts")
	}
	for i := 0; i < v.NumField(); i++ {
		n := v.Field(i).Int()
		if n < 0 {
			return requestErrorAt("request_invalid", "/counts", "invalid evidence counts")
		}
		out[typ.Field(i).Tag.Get("json")] = uint64(n)
	}
	return nil
}

// Reconstruct summary-only report shells to reuse canonical saved admission and
// saved-evidence comparison. This performs no current capability evaluation.
// A comparison does not retain selection traversal failures, so its explicit
// incomplete execution flag is preserved after validating all retained entries.
func admitSavedComparison(c ComparisonReport) error {
	if c.SchemaVersion != 1 || len(c.Entries) == 0 {
		return requestErrorAt("request_invalid", "/comparison", "invalid saved comparison")
	}
	b := Report{SchemaVersion: 1, Status: analysis.Valid, ExecutionComplete: true, Selection: corpus.Selection{Mode: "inline", Complete: true, IgnoredNames: []string{}, SkippedSymlinks: []string{}, TraversalFailures: []corpus.AcquisitionError{}}, Entries: []ReportEntry{}, Provenance: c.BeforeProvenance}
	a := b
	a.Entries = []ReportEntry{}
	a.Provenance = c.AfterProvenance
	for _, e := range c.Entries {
		if e.ID != e.Before.ID || e.ID != e.After.ID {
			return requestErrorAt("request_invalid", "/comparison/entries", "comparison identities disagree")
		}
		b.Entries = append(b.Entries, e.Before)
		a.Entries = append(a.Entries, e.After)
	}
	b.Counts.Selected = len(b.Entries)
	a.Counts.Selected = len(a.Entries)
	finalize(&b)
	finalize(&a)
	expected, err := Compare(CompareRequest{SchemaVersion: 1, Before: b, After: a})
	if err != nil {
		return err
	}
	if !c.ExecutionComplete {
		expected.ExecutionComplete = false
		expected.CIExitCode = 2
	}
	// JSON comparison normalizes RawMessage whitespace while preserving all
	// classifications, pointers, canonical delta values and source entries.
	x, err := exportCopy(c)
	if err != nil {
		return requestErrorAt("request_invalid", "/comparison", "invalid saved comparison")
	}
	y, err := exportCopy(*expected)
	if err != nil {
		return requestErrorAt("request_invalid", "/comparison", "invalid saved comparison")
	}
	var xm, ym any
	xr, _ := json.Marshal(x)
	yr, _ := json.Marshal(y)
	_ = json.Unmarshal(xr, &xm)
	_ = json.Unmarshal(yr, &ym)
	if !reflect.DeepEqual(xm, ym) {
		return requestErrorAt("request_invalid", "/comparison", "comparison summary or evidence disagrees with retained sources")
	}
	return nil
}

type evidenceProjector struct {
	out                        *EvidenceReport
	include, emitted           map[string]bool
	detection, input, evidence int
}

func (p *evidenceProjector) add(pointer, kind, outcome string, complete bool) *EvidenceItem {
	token := ""
	if kind == "detection" {
		p.detection++
		token = fmt.Sprintf("detection-%d", p.detection)
	}
	if kind == "input" {
		p.input++
		token = fmt.Sprintf("input-%d", p.input)
	}
	if token == "" {
		p.evidence++
		token = fmt.Sprintf("evidence-%d", p.evidence)
	}
	p.out.Items = append(p.out.Items, EvidenceItem{Token: token, Pointer: pointer, Kind: kind, Outcome: outcome, Codes: []string{}, Complete: complete, Coverage: []EvidenceCoverage{}})
	return &p.out.Items[len(p.out.Items)-1]
}
func (p *evidenceProjector) entry(e ReportEntry, path string) {
	p.add(path, "detection", publicOutcome(string(e.Status)), entryEvidenceComplete(e))
	p.walk(reflect.ValueOf(e), path, false)
}
func (p *evidenceProjector) detail(path, category string, value any) {
	if !p.include[category] {
		return
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return
	} // Admission has already validated every raw value.
	item := p.add(path, "disclosure", "included", false)
	item.Details = map[string]json.RawMessage{category: raw}
	p.emitted[category] = true
}
func (p *evidenceProjector) walk(v reflect.Value, path string, definition bool) {
	p.walkCovered(v, path, definition, true)
}
func (p *evidenceProjector) walkCovered(v reflect.Value, path string, definition, complete bool) {
	if !v.IsValid() {
		return
	}
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return
		}
		p.walkCovered(v.Elem(), path, definition, complete)
		return
	}
	if v.Type() == reflect.TypeOf(json.RawMessage{}) {
		return
	}
	if v.Kind() == reflect.Slice {
		for i := 0; i < v.Len(); i++ {
			p.walkCovered(v.Index(i), fmt.Sprintf("%s/%d", path, i), definition, complete)
		}
		return
	}
	if v.Kind() != reflect.Struct {
		return
	}
	switch x := v.Interface().(type) {
	case analysis.Result:
		item := p.add(path, "analysis", publicOutcome(string(x.Status)), analysisEvidenceComplete(&x))
		item.Capability = &PublicCapability{Language: publicEnum(x.Document.Language, "spl", "spl2"), Profile: publicEnum(x.Document.Profile, "splunkd"), Version: publicEnum(x.Document.Version, "current"), Revision: publicRevision(x.Requirements.CapabilityRevision)}
		if item.Capability.Language == "unrecognized" || item.Capability.Profile == "unrecognized" || item.Capability.Version == "unrecognized" {
			item.Complete = false
		}
		item.Coverage = []EvidenceCoverage{{"syntax", booleanCoverage(x.Coverage.SyntaxComplete)}, {"semantics", booleanCoverage(x.Coverage.SemanticComplete)}, {"requirements", booleanCoverage(x.Requirements.Coverage.Complete)}}
	case analysis.QueryInput:
		item := p.add(path, "input", publicEnum(x.Evidence.State, "complete", "partial", "not_applicable"), knownInputKind(x.Kind) && (x.Evidence.State == "complete" || x.Evidence.State == "not_applicable"))
		item.Coverage = []EvidenceCoverage{{"target_discovery", publicState(x.Evidence.State)}}
	case analysis.RequirementItem:
		p.add(path, publicRequirementKind(x.Kind), publicEnum(x.Resolution, "exact", "dynamic", "wildcard", "unresolved"), false)
	case compatibility.Report:
		complete = complete && compatibilityReportEvidenceComplete(&x)
		p.add(path, "compatibility", publicOutcome(x.Outcome), complete)
	case compatibility.ResolutionReport:
		complete = complete && resolutionReportEvidenceComplete(&x)
		p.add(path, "compatibility", publicOutcome(x.Outcome), complete)
	case compatibility.InputOutcome:
		p.add(path, "input_assessment", publicOutcome(x.Outcome), complete && knownAssessmentOutcome(x.Outcome) && (x.Outcome == "satisfied" || x.Outcome == "missing" || x.Outcome == "neutral"))
	case compatibility.RequirementOutcome:
		p.add(path, "requirement_assessment", publicOutcome(x.Outcome), complete && requirementAssessmentLabelsRecognized(x) && (x.Outcome == "satisfied" || x.Outcome == "neutral" || x.Applicability == "inapplicable" || (x.Outcome == "missing" && x.Applicability == "applicable")))
	case compatibility.Coverage:
		item := p.add(path, "coverage", publicState(x.State), publicDimension(x.Dimension) != "unrecognized" && (x.State == "complete" || x.State == "not_applicable"))
		item.Coverage = []EvidenceCoverage{{publicDimension(x.Dimension), publicState(x.State)}}
	case analysis.InputCoverage:
		item := p.add(path, "coverage", publicState(x.State), x.State == "complete" || x.State == "not_applicable")
		item.Coverage = []EvidenceCoverage{{"input_evidence", publicState(x.State)}}
	case analysis.RequirementCoverage:
		item := p.add(path, "coverage", booleanCoverage(x.Complete), x.Complete)
		item.Coverage = []EvidenceCoverage{{"requirements", booleanCoverage(x.Complete)}}
	case closure.Report:
		complete = complete && closureEvidenceComplete(&x)
		p.add(path, "closure", publicOutcome(string(x.Status)), complete)
	case closure.TraversalEdge:
		p.add(path, publicRequirementKind(x.Kind), publicEnum(x.Resolution, "resolved", "bound", "missing", "unknown", "ambiguous", "cycle", "dynamic", "wildcard", "unavailable", "incomplete"), publicRequirementKind(x.Kind) != "unrecognized" && (x.Resolution == "resolved" || x.Resolution == "bound"))
	case validation.FieldProjection:
		p.add(path, "schema_projection", publicOutcome(x.Outcome), complete && (x.Outcome == "required" || x.Outcome == "optional" || x.Outcome == "missing"))
	case closure.ClosureCoverage:
		item := p.add(path, "coverage", booleanCoverage(x.Complete), complete && x.Complete)
		item.Coverage = []EvidenceCoverage{{"dependency_closure", booleanCoverage(x.Complete)}}
	case resolution.Variant:
		variantComplete := (x.Outcome == "verified" || x.Outcome == "failed") && (x.CandidateAnalysis == nil || analysisLabelsRecognized(x.CandidateAnalysis))
		if x.Compatibility != nil {
			variantComplete = variantComplete && resolutionReportEvidenceComplete(x.Compatibility)
		}
		p.add(path, "variant", publicOutcome(x.Outcome), variantComplete)
	case Failure:
		item := p.add(path, "failure", "incomplete", false)
		item.Codes = append(item.Codes, publicCode(x.Code))
	}
	typ := v.Type()
	for i := 0; i < v.NumField(); i++ {
		f := typ.Field(i)
		if !f.IsExported() {
			continue
		}
		name := jsonFieldName(f)
		if name == "" {
			continue
		}
		nested := v.Field(i)
		pointer := path + "/" + name
		category := disclosureField(typ.Name(), f.Name, definition)
		if f.Name == "Reasons" && nested.Kind() == reflect.Slice && nested.Type().Elem().Kind() == reflect.String {
			category = "diagnostic_details"
			for j := 0; j < nested.Len(); j++ {
				item := p.add(fmt.Sprintf("%s/%d", pointer, j), "diagnostic", "reported", false)
				item.Codes = append(item.Codes, publicCode(nested.Index(j).String()))
			}
		}
		if category != "" {
			if !isAbsent(nested) {
				p.detail(pointer, category, nested.Interface())
			}
		}
		if f.Name == "Code" && nested.Kind() == reflect.String && typ != reflect.TypeOf(Failure{}) {
			item := p.add(path, "diagnostic", "reported", false)
			item.Codes = append(item.Codes, publicCode(nested.String()))
		}
		// Definition analysis bodies are a separate disclosure category, including
		// their effective expansions. No document is copied as an environment object.
		childDefinition := definition || (typ == reflect.TypeOf(closure.Report{}) && f.Name == "EffectiveAnalysis") || typ == reflect.TypeOf(closure.DefinitionAnalysis{}) || (typ.Name() == "Object" && typ.PkgPath() == "github.com/delgado-jacob/spl-toolkit/pkg/environment")
		p.walkCovered(nested, pointer, childDefinition, complete)
	}
}

func compatibilityEvidenceComplete(outcome string, coverage []compatibility.Coverage, c *closure.Report, requirements analysis.RequirementSet, effective *analysis.RequirementSet) bool {
	if !requirementLabelsRecognized(requirements) || (effective != nil && !requirementLabelsRecognized(*effective)) {
		return false
	}
	if len(coverage) == 0 || (outcome != "satisfied" && outcome != "unsatisfied") {
		return false
	}
	for _, item := range coverage {
		if publicDimension(item.Dimension) == "unrecognized" || (item.State != "complete" && item.State != "not_applicable") {
			return false
		}
	}
	return c == nil || closureEvidenceComplete(c)
}

func publicRequirementKind(kind string) string {
	return publicEnum(kind, "field", "index", "source", "sourcetype", "dataset", "lookup", "data_model", "macro", "function", "command", "external_command", "saved_search", "event_type", "tag", "calculated_field", "field_extraction", "module")
}
func knownInputKind(kind string) bool {
	return publicEnum(kind, "explicit_dataset", "named_placeholder", "unresolved_source", "implicit_stream", "index", "source", "sourcetype") != "unrecognized"
}
func analysisEvidenceComplete(a *analysis.Result) bool {
	if !analysisLabelsRecognized(a) || a.Status == analysis.Incomplete || !a.Coverage.SyntaxComplete || !a.Coverage.SemanticComplete || !a.Requirements.Coverage.Complete {
		return false
	}
	if publicEnum(a.Document.Language, "spl", "spl2") == "unrecognized" || a.Document.Profile != "splunkd" || a.Document.Version != "current" {
		return false
	}
	for _, in := range a.Inputs {
		if !knownInputKind(in.Kind) || (in.Evidence.State != "complete" && in.Evidence.State != "not_applicable") {
			return false
		}
	}
	for _, item := range a.Requirements.Items {
		if publicRequirementKind(item.Kind) == "unrecognized" {
			return false
		}
	}
	return true
}
func entryEvidenceComplete(e ReportEntry) bool {
	if !comparisonComplete(e) {
		return false
	}
	if !analysisLabelsRecognized(e.Analysis) {
		return false
	}
	if e.Compatibility != nil {
		return compatibilityReportEvidenceComplete(e.Compatibility)
	}
	if e.Resolution != nil {
		for _, v := range e.Resolution.Variants {
			if v.CandidateAnalysis != nil && !analysisLabelsRecognized(v.CandidateAnalysis) {
				return false
			}
			if v.Compatibility != nil && !resolutionReportEvidenceComplete(v.Compatibility) {
				return false
			}
		}
	}
	return true
}

// A saved coverage flag cannot establish completeness when its enclosing
// canonical status is missing or unknown. Historical known statuses remain
// independent of the currently installed capability revision.
func closureEvidenceComplete(c *closure.Report) bool {
	if !c.Coverage.Complete || (c.Status != analysis.Valid && c.Status != analysis.Invalid) || !analysisLabelsRecognized(c.DirectAnalysis) || !analysisLabelsRecognized(c.EffectiveAnalysis) {
		return false
	}
	for _, definition := range c.DefinitionAnalyses {
		for _, a := range []*analysis.Result{definition.DirectAnalysis, definition.EffectiveAnalysis} {
			if a != nil && !analysisLabelsRecognized(a) {
				return false
			}
		}
	}
	for _, edge := range c.Traversal {
		if publicRequirementKind(edge.Kind) == "unrecognized" || publicEnum(edge.Resolution, "resolved", "bound", "missing", "unknown", "ambiguous", "cycle", "dynamic", "wildcard", "unavailable", "incomplete") == "unrecognized" {
			return false
		}
	}
	return true
}

// Recognition is separate from completeness: supported closure may discharge
// a known gap in original analysis, but cannot assign meaning to imported labels.
func analysisLabelsRecognized(a *analysis.Result) bool {
	if a == nil || publicEnum(string(a.Status), "valid", "invalid", "incomplete") == "unrecognized" || publicEnum(a.Document.Language, "spl", "spl2") == "unrecognized" || a.Document.Profile != "splunkd" || a.Document.Version != "current" {
		return false
	}
	return requirementLabelsRecognized(a.Requirements) && inputLabelsRecognized(a.Inputs) && knownInputState(a.InputCoverage.State) && knownInputState(a.FieldAttributionCoverage.State) && knownInputState(a.Correlation.Coverage.State)
}
func requirementLabelsRecognized(set analysis.RequirementSet) bool {
	if publicEnum(string(set.QueryStatus), "valid", "invalid", "incomplete") == "unrecognized" || publicEnum(set.Query.Language, "spl", "spl2") == "unrecognized" || set.Query.Profile != "splunkd" || set.Query.Version != "current" || !inputLabelsRecognized(set.Inputs) || !knownInputState(set.InputCoverage.State) || !knownInputState(set.FieldAttributionCoverage.State) || !knownInputState(set.Correlation.Coverage.State) {
		return false
	}
	for _, item := range set.Items {
		if publicRequirementKind(item.Kind) == "unrecognized" || publicEnum(item.Resolution, "exact", "dynamic", "wildcard", "unresolved") == "unrecognized" {
			return false
		}
	}
	return true
}
func inputLabelsRecognized(inputs []analysis.QueryInput) bool {
	for _, in := range inputs {
		if !knownInputKind(in.Kind) || !knownInputState(in.Evidence.State) {
			return false
		}
	}
	return true
}

func knownInputState(state string) bool {
	return publicEnum(state, "complete", "partial", "not_applicable") != "unrecognized"
}

func knownAssessmentOutcome(outcome string) bool {
	return publicEnum(outcome, "satisfied", "missing", "ambiguous", "indeterminate", "neutral") != "unrecognized"
}
func requirementAssessmentLabelsRecognized(outcome compatibility.RequirementOutcome) bool {
	if outcome.FieldProjection != nil && publicEnum(outcome.FieldProjection.Outcome, "required", "optional", "missing", "conditional", "indeterminate") == "unrecognized" {
		return false
	}
	return knownAssessmentOutcome(outcome.Outcome) && publicEnum(outcome.Applicability, "applicable", "inapplicable", "indeterminate") != "unrecognized"
}
func compatibilityReportEvidenceComplete(c *compatibility.Report) bool {
	if !compatibilityEvidenceComplete(c.Outcome, c.Coverage, c.Closure, c.Requirements, c.EffectiveRequirements) {
		return false
	}
	for _, input := range c.Inputs {
		if !knownAssessmentOutcome(input.Outcome) {
			return false
		}
	}
	for _, outcome := range c.RequirementOutcomes {
		if !requirementAssessmentLabelsRecognized(outcome) {
			return false
		}
	}
	return true
}
func resolutionReportEvidenceComplete(c *compatibility.ResolutionReport) bool {
	coverage := make([]compatibility.Coverage, 0, len(c.Coverage))
	for _, item := range c.Coverage {
		coverage = append(coverage, item.Evidence)
	}
	if !compatibilityEvidenceComplete(c.Outcome, coverage, c.Closure, c.Requirements, c.EffectiveRequirements) {
		return false
	}
	for _, input := range c.Inputs {
		if !knownAssessmentOutcome(input.Evidence.Outcome) {
			return false
		}
	}
	for _, outcome := range c.RequirementOutcomes {
		if !requirementAssessmentLabelsRecognized(outcome.Evidence) {
			return false
		}
	}
	return true
}
