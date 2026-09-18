package analysis_test

import (
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/rewrite"
)

type evidenceDimensionResult struct {
	passed   bool
	expected string
	actual   string
	refused  bool
}

type evidenceExecution struct {
	dimensions map[string]evidenceDimensionResult
}

type capabilityDimension struct {
	name  string
	claim analysis.CapabilityClaim
}

type diagnosticFact struct {
	Code     string
	Category string
	Severity string
	Location analysis.Location
}

type semanticFacts struct {
	Status       analysis.Status
	Complete     bool
	Stages       []analysis.CapabilityStageExpectation
	References   []analysis.CapabilityReferenceExpectation
	Dependencies []analysis.CapabilityDependencyExpectation
	Transitions  []analysis.CapabilityTransitionExpectation
	Diagnostics  []diagnosticFact
}

type requirementFacts struct {
	QueryStatus analysis.Status
	Complete    bool
	Items       []analysis.CapabilityRequirementExpectation
	GapCodes    []string
}

type rewriteFacts struct {
	Status                analysis.Status
	Committed             bool
	RewriteComplete       bool
	Text                  string
	CandidateText         string
	CoverageReasons       []string
	ChangeReasons         []string
	RuleEvaluationReasons []string
}

func TestCapabilityEvidenceCorpus(t *testing.T) {
	manifests := make([]analysis.CapabilityManifest, 0, 2)
	for _, language := range []string{"spl", "spl2"} {
		manifest, err := analysis.CapabilitiesFor(analysis.CapabilityOptions{
			Language: language,
			Profile:  "splunkd",
			Version:  "current",
		})
		if err != nil {
			t.Fatalf("record=<manifest> evidence=<selection> dimension=<all> expected=available %s capability manifest actual=%v", language, err)
		}
		manifests = append(manifests, manifest)
	}

	executions := make(map[string]evidenceExecution)
	evidenceByID := make(map[string]analysis.CapabilityEvidence)
	for _, manifest := range manifests {
		for _, evidence := range manifest.Evidence {
			if _, exists := executions[evidence.ID]; exists {
				t.Fatalf("record=<manifest> evidence=%s dimension=<all> expected=execute selected evidence once actual=duplicate selection", evidence.ID)
			}
			executions[evidence.ID] = executeEvidence(evidence)
			evidenceByID[evidence.ID] = evidence
		}
	}

	for _, manifest := range manifests {
		for _, record := range manifest.Records {
			for _, dimension := range recordDimensions(record.Dimensions) {
				assertClaimEvidence(t, record.ID, dimension, evidenceByID, executions)
			}
		}
	}
}

func executeEvidence(evidence analysis.CapabilityEvidence) evidenceExecution {
	execution := evidenceExecution{
		dimensions: make(map[string]evidenceDimensionResult, 5),
	}

	if evidence.Observations.Syntax != nil || evidence.Observations.Semantics != nil || evidence.Observations.Linting != nil {
		result, err := analysis.Analyze(evidence.Document)
		if err != nil {
			actual := fmt.Sprintf("Analyze error: %v", err)
			for _, dimension := range []struct {
				name     string
				observed bool
			}{
				{name: "syntax", observed: evidence.Observations.Syntax != nil},
				{name: "semantics", observed: evidence.Observations.Semantics != nil},
				{name: "linting", observed: evidence.Observations.Linting != nil},
			} {
				if dimension.observed {
					execution.dimensions[dimension.name] = evidenceDimensionResult{expected: "successful analysis", actual: actual}
				}
			}
		} else {
			if observation := evidence.Observations.Syntax; observation != nil {
				execution.dimensions["syntax"] = compareSyntax(*observation, result)
			}
			if observation := evidence.Observations.Semantics; observation != nil {
				execution.dimensions["semantics"] = compareSemantics(*observation, result)
			}
			if observation := evidence.Observations.Linting; observation != nil {
				execution.dimensions["linting"] = compareLinting(*observation, result)
			}
		}
	}

	if observation := evidence.Observations.Requirements; observation != nil {
		set, err := analysis.Requirements(evidence.Document)
		if err != nil {
			execution.dimensions["requirements"] = evidenceDimensionResult{
				expected: formatValue(*observation),
				actual:   fmt.Sprintf("Requirements error: %v", err),
			}
		} else {
			execution.dimensions["requirements"] = compareRequirements(*observation, set)
		}
	}

	if observation := evidence.Observations.SafeRewriting; observation != nil {
		execution.dimensions["safe_rewriting"] = compareRewrite(evidence, *observation)
	}
	return execution
}

func compareSyntax(expected analysis.CapabilitySyntaxObservation, result *analysis.Result) evidenceDimensionResult {
	wantDiagnostics := diagnosticFacts(expected.Diagnostics)
	gotDiagnostics := resultDiagnosticFacts(result.Diagnostics)
	passed := expected.Complete == result.Coverage.SyntaxComplete && containsAll(gotDiagnostics, wantDiagnostics)
	return evidenceDimensionResult{
		passed: passed,
		expected: formatValue(struct {
			Complete    bool
			Diagnostics []diagnosticFact
		}{expected.Complete, wantDiagnostics}),
		actual: formatValue(struct {
			Complete    bool
			Diagnostics []diagnosticFact
		}{result.Coverage.SyntaxComplete, gotDiagnostics}),
	}
}

func compareSemantics(expected analysis.CapabilitySemanticsObservation, result *analysis.Result) evidenceDimensionResult {
	actual := semanticFacts{
		Status:       result.Status,
		Complete:     result.Coverage.SemanticComplete,
		Stages:       stageFacts(result.Stages),
		References:   referenceFacts(result.References),
		Dependencies: dependencyFacts(result.Dependencies),
		Transitions:  transitionFacts(result.Lineage),
		Diagnostics:  resultDiagnosticFacts(result.Diagnostics),
	}
	passed := expected.Status == actual.Status &&
		expected.Complete == actual.Complete &&
		containsAll(actual.Stages, expected.Stages) &&
		containsAll(actual.References, expected.References) &&
		containsAll(actual.Dependencies, expected.Dependencies) &&
		containsAll(actual.Transitions, expected.Transitions) &&
		containsAll(actual.Diagnostics, diagnosticFacts(expected.Diagnostics))
	return evidenceDimensionResult{passed: passed, expected: formatValue(expected), actual: formatValue(actual)}
}

func compareLinting(expected analysis.CapabilityLintingObservation, result *analysis.Result) evidenceDimensionResult {
	want := diagnosticFacts(expected.Diagnostics)
	got := resultDiagnosticFacts(result.Diagnostics)
	return evidenceDimensionResult{passed: containsAll(got, want), expected: formatValue(want), actual: formatValue(got)}
}

func compareRequirements(expected analysis.CapabilityRequirementsObservation, set *analysis.RequirementSet) evidenceDimensionResult {
	actual := requirementFacts{
		QueryStatus: set.QueryStatus,
		Complete:    set.Coverage.Complete,
		Items:       requirementItemFacts(set.Items),
		GapCodes:    requirementGapCodes(set.Gaps),
	}
	passed := expected.QueryStatus == actual.QueryStatus &&
		expected.Complete == actual.Complete &&
		containsAll(actual.Items, expected.Items) &&
		slices.Equal(expected.GapCodes, actual.GapCodes)
	return evidenceDimensionResult{passed: passed, expected: formatValue(expected), actual: formatValue(actual)}
}

func compareRewrite(evidence analysis.CapabilityEvidence, expected analysis.CapabilityRewriteObservation) evidenceDimensionResult {
	var request struct {
		Mode  rewrite.Mode   `json:"mode"`
		Rules []rewrite.Rule `json:"rules"`
	}
	if err := json.Unmarshal(evidence.RewriteRequest, &request); err != nil {
		return evidenceDimensionResult{expected: formatValue(expected), actual: fmt.Sprintf("decode rewrite request: %v", err)}
	}
	result, err := rewrite.Rewrite(rewrite.Request{
		SchemaVersion: 1,
		Mode:          request.Mode,
		Document:      evidence.Document,
		Rules:         request.Rules,
	})
	if err != nil {
		return evidenceDimensionResult{expected: formatValue(expected), actual: fmt.Sprintf("Rewrite error: %v", err)}
	}
	actual := rewriteFacts{
		Status:                result.Status,
		Committed:             result.Committed,
		RewriteComplete:       result.Coverage.RewriteComplete,
		Text:                  result.Text,
		CandidateText:         result.CandidateText,
		CoverageReasons:       result.Coverage.Reasons,
		ChangeReasons:         changeReasons(result.Changes),
		RuleEvaluationReasons: ruleEvaluationReasons(result.RuleEvaluations),
	}
	want := rewriteFacts{
		Status:                expected.Status,
		Committed:             expected.Committed,
		RewriteComplete:       expected.RewriteComplete,
		Text:                  expected.Text,
		CandidateText:         expected.CandidateText,
		CoverageReasons:       expected.CoverageReasons,
		ChangeReasons:         expected.ChangeReasons,
		RuleEvaluationReasons: expected.RuleEvaluationReasons,
	}
	return evidenceDimensionResult{
		passed:   reflect.DeepEqual(want, actual),
		expected: formatValue(want),
		actual:   formatValue(actual),
		refused:  !result.Coverage.RewriteComplete,
	}
}

func assertClaimEvidence(t *testing.T, recordID string, dimension capabilityDimension, evidenceByID map[string]analysis.CapabilityEvidence, executions map[string]evidenceExecution) {
	t.Helper()
	claim := dimension.claim
	if claim.State != analysis.CapabilitySupported && claim.State != analysis.CapabilityPartial && claim.State != analysis.CapabilityUnsupported {
		return
	}

	positive, incomplete, negative := 0, 0, 0
	for _, evidenceID := range claim.EvidenceIDs {
		evidence, found := evidenceByID[evidenceID]
		if !found {
			claimFailure(t, recordID, evidenceID, dimension.name, "selected evidence", "missing")
			continue
		}
		execution, found := executions[evidenceID]
		if !found {
			claimFailure(t, recordID, evidenceID, dimension.name, "evidence executed exactly once", "not executed")
			continue
		}
		result, found := execution.dimensions[dimension.name]
		if !found {
			claimFailure(t, recordID, evidenceID, dimension.name, "typed observation and execution result", "missing result")
			continue
		}
		if !result.passed {
			claimFailure(t, recordID, evidenceID, dimension.name, result.expected, result.actual)
			continue
		}
		switch evidence.Classification {
		case analysis.CapabilityEvidencePositive:
			if dimension.name != "safe_rewriting" || !result.refused {
				positive++
			}
		case analysis.CapabilityEvidenceIncomplete:
			incomplete++
		case analysis.CapabilityEvidenceNegative:
			negative++
		}
	}

	evidenceIDs := strings.Join(claim.EvidenceIDs, ",")
	switch claim.State {
	case analysis.CapabilitySupported:
		proved := positive > 0 || (dimension.name == "linting" && negative > 0)
		if !proved {
			claimFailure(t, recordID, evidenceIDs, dimension.name, "at least one passing proof of supported behavior", fmt.Sprintf("positive=%d incomplete=%d negative=%d", positive, incomplete, negative))
		}
	case analysis.CapabilityPartial:
		if positive == 0 || incomplete == 0 {
			claimFailure(t, recordID, evidenceIDs, dimension.name, "at least one passing proved behavior and one passing incomplete boundary", fmt.Sprintf("positive=%d incomplete=%d negative=%d", positive, incomplete, negative))
		}
	case analysis.CapabilityUnsupported:
		if incomplete == 0 && negative == 0 {
			claimFailure(t, recordID, evidenceIDs, dimension.name, "at least one passing incomplete or negative boundary", fmt.Sprintf("positive=%d incomplete=%d negative=%d", positive, incomplete, negative))
		}
	}
}

func claimFailure(t *testing.T, recordID, evidenceID, dimension, expected, actual string) {
	t.Helper()
	t.Errorf("record=%s evidence=%s dimension=%s expected=%s actual=%s", recordID, evidenceID, dimension, expected, actual)
}

func recordDimensions(dimensions analysis.CapabilityDimensions) []capabilityDimension {
	return []capabilityDimension{
		{name: "syntax", claim: dimensions.Syntax},
		{name: "semantics", claim: dimensions.Semantics},
		{name: "requirements", claim: dimensions.Requirements},
		{name: "linting", claim: dimensions.Linting},
		{name: "safe_rewriting", claim: dimensions.SafeRewriting},
	}
}

func diagnosticFacts(expectations []analysis.CapabilityDiagnosticExpectation) []diagnosticFact {
	facts := make([]diagnosticFact, 0, len(expectations))
	for _, diagnostic := range expectations {
		facts = append(facts, diagnosticFact{diagnostic.Code, diagnostic.Category, diagnostic.Severity, diagnostic.Location})
	}
	return facts
}

func resultDiagnosticFacts(diagnostics []analysis.Diagnostic) []diagnosticFact {
	facts := make([]diagnosticFact, 0, len(diagnostics))
	for _, diagnostic := range diagnostics {
		facts = append(facts, diagnosticFact{diagnostic.Code, diagnostic.Category, diagnostic.Severity, diagnostic.Location})
	}
	return facts
}

func stageFacts(stages []analysis.Stage) []analysis.CapabilityStageExpectation {
	facts := make([]analysis.CapabilityStageExpectation, 0, len(stages))
	for _, stage := range stages {
		facts = append(facts, analysis.CapabilityStageExpectation{Command: stage.Command, SemanticComplete: stage.SemanticComplete})
	}
	return facts
}

func referenceFacts(references []analysis.Reference) []analysis.CapabilityReferenceExpectation {
	facts := make([]analysis.CapabilityReferenceExpectation, 0, len(references))
	for _, reference := range references {
		facts = append(facts, analysis.CapabilityReferenceExpectation{
			NormalizedName: reference.NormalizedName,
			Kind:           reference.Kind,
			Role:           reference.Role,
			Resolution:     reference.Resolution,
			Binding:        reference.Binding,
			Location:       reference.Location,
		})
	}
	return facts
}

func dependencyFacts(dependencies analysis.Dependencies) []analysis.CapabilityDependencyExpectation {
	facts := make([]analysis.CapabilityDependencyExpectation, 0)
	add := func(kind string, names []string) {
		for _, name := range names {
			facts = append(facts, analysis.CapabilityDependencyExpectation{Kind: kind, Name: name})
		}
	}
	add("index", dependencies.Indexes)
	add("source", dependencies.Sources)
	add("sourcetype", dependencies.SourceTypes)
	add("dataset", dependencies.Datasets)
	add("lookup", dependencies.Lookups)
	add("data_model", dependencies.DataModels)
	add("macro", dependencies.Macros)
	return facts
}

func transitionFacts(lineage []analysis.Lineage) []analysis.CapabilityTransitionExpectation {
	facts := make([]analysis.CapabilityTransitionExpectation, 0)
	for _, entry := range lineage {
		for _, transition := range entry.Transitions {
			facts = append(facts, analysis.CapabilityTransitionExpectation{Operation: transition.Operation, Output: transition.Output})
		}
	}
	return facts
}

func requirementItemFacts(items []analysis.RequirementItem) []analysis.CapabilityRequirementExpectation {
	facts := make([]analysis.CapabilityRequirementExpectation, 0, len(items))
	for _, item := range items {
		facts = append(facts, analysis.CapabilityRequirementExpectation{
			Kind:       item.Kind,
			Identity:   item.Identity,
			Role:       item.Role,
			Necessity:  item.Necessity,
			Resolution: item.Resolution,
		})
	}
	return facts
}

func requirementGapCodes(gaps []analysis.RequirementGap) []string {
	codes := make([]string, 0, len(gaps))
	for _, gap := range gaps {
		codes = append(codes, gap.Code)
	}
	return codes
}

func changeReasons(changes []rewrite.Change) []string {
	reasons := make([]string, 0, len(changes))
	for _, change := range changes {
		reasons = append(reasons, change.Reason)
	}
	return reasons
}

func ruleEvaluationReasons(evaluations []rewrite.RuleEvaluation) []string {
	reasons := make([]string, 0, len(evaluations))
	for _, evaluation := range evaluations {
		reasons = append(reasons, evaluation.Reason)
	}
	return reasons
}

func containsAll[T comparable](actual, selected []T) bool {
	used := make([]bool, len(actual))
	for _, expected := range selected {
		found := false
		for i, candidate := range actual {
			if !used[i] && candidate == expected {
				used[i] = true
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func formatValue(value any) string {
	return fmt.Sprintf("%#v", value)
}
