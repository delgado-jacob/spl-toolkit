package analysis_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"slices"
	"strings"
	"testing"
	"unicode/utf8"

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

type capabilityRewriteRequest struct {
	SchemaVersion int            `json:"schema_version"`
	Mode          rewrite.Mode   `json:"mode"`
	Rules         []rewrite.Rule `json:"rules"`
}

type rewriteCapabilityFixture struct {
	Checks []rewriteCapabilityCheck `json:"capability_checks"`
}

type rewriteCapabilityCheck struct {
	Language  string                     `json:"language"`
	Kind      string                     `json:"kind"`
	Role      string                     `json:"role"`
	Supported bool                       `json:"supported"`
	Positive  *rewriteCapabilityPositive `json:"positive"`
	Negative  rewriteCapabilityNegative  `json:"negative"`
}

type rewriteCapabilityPositive struct {
	Query     string `json:"query"`
	Source    string `json:"source"`
	Target    string `json:"target"`
	Candidate string `json:"candidate"`
}

type rewriteCapabilityNegative struct {
	Query  string `json:"query"`
	Source string `json:"source"`
	Target string `json:"target"`
	Reason string `json:"reason"`
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

func TestAdvertisedRewriteFormsHaveLedgerEvidence(t *testing.T) {
	fixture := loadRewriteCapabilityFixture(t)
	manifests := make([]analysis.CapabilityManifest, 0, 2)
	for _, language := range []string{"spl", "spl2"} {
		manifest, err := analysis.CapabilitiesFor(analysis.CapabilityOptions{
			Language: language,
			Profile:  "splunkd",
			Version:  "current",
		})
		if err != nil {
			t.Fatal(err)
		}
		manifests = append(manifests, manifest)
	}
	if err := validateAdvertisedRewriteEvidence(manifests, fixture); err != nil {
		t.Fatal(err)
	}
}

func TestAdvertisedSupportedRewriteEvidenceRejectsSwappedForms(t *testing.T) {
	fixture := loadRewriteCapabilityFixture(t)
	manifests := make([]analysis.CapabilityManifest, 0, 2)
	for _, language := range []string{"spl", "spl2"} {
		manifest, err := analysis.CapabilitiesFor(analysis.CapabilityOptions{
			Language: language,
			Profile:  "splunkd",
			Version:  "current",
		})
		if err != nil {
			t.Fatal(err)
		}
		manifests = append(manifests, manifest)
	}
	if err := validateAdvertisedRewriteEvidence(manifests, fixture); err != nil {
		t.Fatalf("valid rewrite evidence was rejected before swap: %v", err)
	}
	manifest := &manifests[0]
	var expression, search *analysis.CapabilityRecord
	for i := range manifest.Records {
		record := &manifest.Records[i]
		if record.Kind != "expression" || record.Name != "field" {
			continue
		}
		switch record.Form {
		case "expression_atom":
			expression = record
		case "search_field":
			search = record
		}
	}
	if expression == nil || search == nil {
		t.Fatal("required rewrite records are missing")
	}
	expression.Dimensions.SafeRewriting.EvidenceIDs, search.Dimensions.SafeRewriting.EvidenceIDs =
		search.Dimensions.SafeRewriting.EvidenceIDs, expression.Dimensions.SafeRewriting.EvidenceIDs
	err := validateAdvertisedRewriteEvidence(manifests, fixture)
	if err == nil {
		t.Fatal("swapped rewrite evidence IDs were accepted")
	}
	if !strings.Contains(err.Error(), `spl/field/expression_atom`) || !strings.Contains(err.Error(), "document") {
		t.Fatalf("swapped rewrite evidence IDs failed for the wrong reason: %v", err)
	}
}

func loadRewriteCapabilityFixture(t *testing.T) rewriteCapabilityFixture {
	t.Helper()
	data, err := os.ReadFile("../../testdata/rewrite/forms.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture rewriteCapabilityFixture
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	return fixture
}

func validateAdvertisedRewriteEvidence(manifests []analysis.CapabilityManifest, fixture rewriteCapabilityFixture) error {
	checksByKey := make(map[string]rewriteCapabilityCheck)
	for _, check := range fixture.Checks {
		key := rewriteCapabilityKey(check.Language, check.Kind, check.Role)
		if check.Supported && check.Positive == nil {
			return fmt.Errorf("supported rewrite fixture %q has no positive case", key)
		}
		if !check.Supported && (check.Positive != nil || check.Negative.Query == "" || check.Negative.Reason == "") {
			return fmt.Errorf("unsupported rewrite fixture %q has no exact refusal case", key)
		}
		if _, exists := checksByKey[key]; exists {
			return fmt.Errorf("rewrite fixture %q is duplicated", key)
		}
		checksByKey[key] = check
	}

	evidenceOwners := make(map[string]string)
	seenForms := make(map[string]bool)
	for _, manifest := range manifests {
		if manifest.Rewrite == nil {
			return fmt.Errorf("%s rewrite compatibility projection is missing", manifest.Language)
		}
		evidenceByID := make(map[string]analysis.CapabilityEvidence, len(manifest.Evidence))
		for _, evidence := range manifest.Evidence {
			evidenceByID[evidence.ID] = evidence
		}
		for _, form := range manifest.Rewrite.Forms {
			key := rewriteCapabilityKey(manifest.Language, form.Kind, form.Role)
			check, found := checksByKey[key]
			if !found {
				return fmt.Errorf("advertised rewrite form %q has no compatibility fixture", key)
			}
			if check.Supported != form.Supported {
				return fmt.Errorf("advertised rewrite form %q support=%t, fixture support=%t", key, form.Supported, check.Supported)
			}
			seenForms[key] = true

			wantRecordKind := "dataset"
			if form.Kind == "field" {
				wantRecordKind = "expression"
			}
			var matches []analysis.CapabilityRecord
			for _, record := range manifest.Records {
				if record.Kind == wantRecordKind && record.Name == form.Kind && record.Form == form.Role {
					matches = append(matches, record)
				}
			}
			if len(matches) != 1 {
				return fmt.Errorf("advertised rewrite form %q has %d canonical ledger records, want 1", key, len(matches))
			}
			claim := matches[0].Dimensions.SafeRewriting
			wantState := analysis.CapabilityUnsupported
			if form.Supported {
				wantState = analysis.CapabilitySupported
			}
			if claim.State != wantState {
				return fmt.Errorf("advertised rewrite form %q has safe_rewriting state %q, want %q", key, claim.State, wantState)
			}
			if !slices.Equal(claim.Limitations, form.Limitations) {
				return fmt.Errorf("advertised rewrite form %q limitations=%v, ledger limitations=%v", key, form.Limitations, claim.Limitations)
			}
			if len(claim.EvidenceIDs) != 1 {
				return fmt.Errorf("advertised rewrite form %q has %d evidence cases, want 1", key, len(claim.EvidenceIDs))
			}
			evidenceID := claim.EvidenceIDs[0]
			if owner, exists := evidenceOwners[evidenceID]; exists {
				return fmt.Errorf("advertised rewrite forms %q and %q share evidence %q", owner, key, evidenceID)
			}
			evidenceOwners[evidenceID] = key
			evidence, found := evidenceByID[evidenceID]
			if !found {
				return fmt.Errorf("advertised rewrite form %q cites missing evidence %q", key, evidenceID)
			}
			if err := validateRewriteEvidenceCase(key, evidence, check); err != nil {
				return err
			}
		}
	}
	for key := range checksByKey {
		if !seenForms[key] {
			return fmt.Errorf("rewrite fixture %q has no advertised form", key)
		}
	}
	if len(evidenceOwners) != len(seenForms) {
		return fmt.Errorf("rewrite forms have %d distinct evidence cases, want %d", len(evidenceOwners), len(seenForms))
	}
	return nil
}

func validateRewriteEvidenceCase(key string, evidence analysis.CapabilityEvidence, check rewriteCapabilityCheck) error {
	query, source, target := check.Negative.Query, check.Negative.Source, check.Negative.Target
	wantClassification := analysis.CapabilityEvidenceIncomplete
	if check.Supported {
		query, source, target = check.Positive.Query, check.Positive.Source, check.Positive.Target
		wantClassification = analysis.CapabilityEvidencePositive
	} else if check.Negative.Reason == "no_match" {
		wantClassification = analysis.CapabilityEvidenceNegative
	}
	if evidence.Classification != wantClassification {
		return fmt.Errorf("advertised rewrite form %q evidence %q classification=%q, want %q", key, evidence.ID, evidence.Classification, wantClassification)
	}
	wantDocument := analysis.QueryDocument{
		Text:     query,
		Language: check.Language,
		Profile:  "splunkd",
		Version:  "current",
		SourceID: "capability:" + evidence.ID,
	}
	if evidence.Document != wantDocument {
		return fmt.Errorf("advertised rewrite form %q evidence %q document=%+v, want %+v", key, evidence.ID, evidence.Document, wantDocument)
	}
	request, err := decodeCapabilityRewriteRequest(evidence.RewriteRequest)
	if err != nil {
		return fmt.Errorf("advertised rewrite form %q evidence %q rewrite request: %w", key, evidence.ID, err)
	}
	if request.SchemaVersion != 1 || request.Mode != rewrite.Apply || len(request.Rules) != 1 {
		return fmt.Errorf("advertised rewrite form %q evidence %q request=%+v, want schema_version 1, apply mode, and one rule", key, evidence.ID, request)
	}
	rule := request.Rules[0]
	sourcePath := check.Role == "navigation"
	if rule.Kind != check.Kind || !rewriteFixtureIdentityEquals(rule.Source, source, sourcePath) || !rewriteFixtureIdentityEquals(rule.Target, target, false) {
		return fmt.Errorf("advertised rewrite form %q evidence %q rule=%+v, want kind=%q source=%q target=%q", key, evidence.ID, rule, check.Kind, source, target)
	}
	observation := evidence.Observations.SafeRewriting
	if observation == nil {
		return fmt.Errorf("advertised rewrite form %q evidence %q has no safe_rewriting observation", key, evidence.ID)
	}
	wantText, wantCommitted, wantComplete := query, false, false
	if check.Supported {
		wantText, wantCommitted, wantComplete = check.Positive.Candidate, true, true
	} else if wantClassification == analysis.CapabilityEvidenceNegative {
		wantComplete = true
	}
	if observation.Committed != wantCommitted || observation.RewriteComplete != wantComplete || observation.Text != wantText || observation.CandidateText != wantText {
		return fmt.Errorf("advertised rewrite form %q evidence %q result=%+v, want committed=%t complete=%t text=%q", key, evidence.ID, *observation, wantCommitted, wantComplete, wantText)
	}
	if wantClassification == analysis.CapabilityEvidenceNegative && (len(observation.CoverageReasons) != 0 || len(observation.ChangeReasons) != 0 || !slices.Equal(observation.RuleEvaluationReasons, []string{"no_match"})) {
		return fmt.Errorf("advertised rewrite form %q evidence %q result=%+v, want exact no-match boundary", key, evidence.ID, *observation)
	}
	if wantClassification == analysis.CapabilityEvidenceIncomplete && !slices.Contains(observation.CoverageReasons, check.Negative.Reason) && !slices.Contains(observation.ChangeReasons, check.Negative.Reason) && !slices.Contains(observation.RuleEvaluationReasons, check.Negative.Reason) {
		return fmt.Errorf("advertised rewrite form %q evidence %q result has no refusal reason %q", key, evidence.ID, check.Negative.Reason)
	}
	result := executeEvidence(evidence).dimensions["safe_rewriting"]
	wantRefused := wantClassification == analysis.CapabilityEvidenceIncomplete
	if !result.passed || result.refused != wantRefused {
		return fmt.Errorf("advertised rewrite form %q evidence %q did not prove its boundary: expected=%s actual=%s", key, evidence.ID, result.expected, result.actual)
	}
	return nil
}

func rewriteCapabilityKey(language, kind, role string) string {
	return language + "/" + kind + "/" + role
}

func rewriteFixtureIdentityEquals(identity rewrite.Identity, name string, path bool) bool {
	if path {
		return identity.Name == nil && slices.Equal(identity.Path, strings.Split(name, "."))
	}
	return identity.Name != nil && *identity.Name == name && len(identity.Path) == 0
}

func TestCapabilityRewriteEvidencePreservesSchemaVersion(t *testing.T) {
	manifest := analysis.Capabilities()
	for _, evidence := range manifest.Evidence {
		if evidence.ID != "spl.splunkd.baseline.rewrite" {
			continue
		}
		evidence.RewriteRequest = json.RawMessage(`{"schema_version":2,"mode":"preview","rules":[]}`)
		request, err := decodeCapabilityRewriteRequest(evidence.RewriteRequest)
		if err != nil || request.SchemaVersion != 2 {
			t.Fatalf("record=spl.profile_form.splunkd.baseline evidence=%s dimension=safe_rewriting expected=decoded schema_version 2 actual=request=%+v error=%v", evidence.ID, request, err)
		}
		result := compareRewrite(evidence, *evidence.Observations.SafeRewriting)
		if result.passed || !strings.Contains(result.actual, "schema_version must be integer 1") {
			t.Fatalf("record=spl.profile_form.splunkd.baseline evidence=%s dimension=safe_rewriting expected=raw schema_version 2 rejected actual=%s", evidence.ID, result.actual)
		}
		return
	}
	t.Fatal("record=spl.profile_form.splunkd.baseline evidence=spl.splunkd.baseline.rewrite dimension=safe_rewriting expected=selected regression evidence actual=missing")
}

func TestCapabilityRewriteRequestWrapperRejectsMalformedJSON(t *testing.T) {
	for name, raw := range map[string]json.RawMessage{
		"invalid syntax":            json.RawMessage(`{"schema_version":1`),
		"unknown property":          json.RawMessage(`{"schema_version":1,"mode":"preview","rules":[],"unexpected":true}`),
		"trailing JSON value":       json.RawMessage(`{"schema_version":1,"mode":"preview","rules":[]} {}`),
		"duplicate schema version":  json.RawMessage(`{"schema_version":1,"schema_version":1,"mode":"preview","rules":[]}`),
		"duplicate rules":           json.RawMessage(`{"schema_version":1,"mode":"preview","rules":[],"rules":[]}`),
		"missing schema version":    json.RawMessage(`{"mode":"preview","rules":[]}`),
		"missing rules":             json.RawMessage(`{"schema_version":1,"mode":"preview"}`),
		"null mode":                 json.RawMessage(`{"schema_version":1,"mode":null,"rules":[]}`),
		"null rules":                json.RawMessage(`{"schema_version":1,"mode":"preview","rules":null}`),
		"duplicate nested rule key": json.RawMessage(`{"schema_version":1,"rules":[{"id":"first","id":"second","kind":"field","source":{"name":"old"},"target":{"name":"new"}}]}`),
		"lone surrogate escape":     json.RawMessage(`{"schema_version":1,"rules":[{"id":"\ud800","kind":"field","source":{"name":"old"},"target":{"name":"new"}}]}`),
	} {
		t.Run(name, func(t *testing.T) {
			if request, err := decodeCapabilityRewriteRequest(raw); err == nil {
				t.Fatalf("record=<wrapper> evidence=<malformed> dimension=safe_rewriting expected=request rejection actual=%+v", request)
			}
		})
	}
}

func TestCapabilityRewriteRequestWrapperAllowsMissingMode(t *testing.T) {
	request, err := decodeCapabilityRewriteRequest(json.RawMessage(`{"schema_version":1,"rules":[]}`))
	if err != nil || request.Mode != "" {
		t.Fatalf("record=<wrapper> evidence=<mode-default> dimension=safe_rewriting expected=optional mode actual=request=%+v error=%v", request, err)
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
	request, err := decodeCapabilityRewriteRequest(evidence.RewriteRequest)
	if err != nil {
		return evidenceDimensionResult{expected: formatValue(expected), actual: fmt.Sprintf("decode rewrite request: %v", err)}
	}
	result, err := rewrite.Rewrite(rewrite.Request{
		SchemaVersion: request.SchemaVersion,
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

func decodeCapabilityRewriteRequest(raw json.RawMessage) (capabilityRewriteRequest, error) {
	if err := validateCapabilityRewriteUnicode(raw); err != nil {
		return capabilityRewriteRequest{}, err
	}
	if err := validateCapabilityJSONDuplicates(raw); err != nil {
		return capabilityRewriteRequest{}, err
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	start, err := decoder.Token()
	if err != nil {
		return capabilityRewriteRequest{}, err
	}
	if start != json.Delim('{') {
		return capabilityRewriteRequest{}, fmt.Errorf("expected an object")
	}
	seen := make(map[string]bool, 3)
	fields := make(map[string]json.RawMessage, 3)
	for decoder.More() {
		token, err := decoder.Token()
		if err != nil {
			return capabilityRewriteRequest{}, err
		}
		key := token.(string)
		if seen[key] {
			return capabilityRewriteRequest{}, fmt.Errorf("duplicate property %q", key)
		}
		seen[key] = true
		if key != "schema_version" && key != "mode" && key != "rules" {
			return capabilityRewriteRequest{}, fmt.Errorf("unknown property %q", key)
		}
		var value json.RawMessage
		if err := decoder.Decode(&value); err != nil {
			return capabilityRewriteRequest{}, fmt.Errorf("invalid property %q: %w", key, err)
		}
		fields[key] = value
	}
	if _, err := decoder.Token(); err != nil {
		return capabilityRewriteRequest{}, err
	}
	var request capabilityRewriteRequest
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return capabilityRewriteRequest{}, fmt.Errorf("expected exactly one JSON value")
		}
		return capabilityRewriteRequest{}, fmt.Errorf("trailing data: %w", err)
	}
	for _, required := range []string{"schema_version", "rules"} {
		if !seen[required] {
			return capabilityRewriteRequest{}, fmt.Errorf("missing property %q", required)
		}
	}
	for _, present := range []string{"mode", "rules"} {
		if value, found := fields[present]; found && bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return capabilityRewriteRequest{}, fmt.Errorf("property %q must not be null", present)
		}
	}
	typed := json.NewDecoder(bytes.NewReader(raw))
	typed.DisallowUnknownFields()
	if err := typed.Decode(&request); err != nil {
		return capabilityRewriteRequest{}, err
	}
	return request, nil
}

func validateCapabilityJSONDuplicates(raw []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	if err := validateCapabilityJSONValue(decoder); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return fmt.Errorf("expected exactly one JSON value")
		}
		return fmt.Errorf("trailing data: %w", err)
	}
	return nil
}

func validateCapabilityJSONValue(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delim, composite := token.(json.Delim)
	if !composite {
		return nil
	}
	switch delim {
	case '{':
		seen := make(map[string]bool)
		for decoder.More() {
			keyToken, err := decoder.Token()
			if err != nil {
				return err
			}
			key := keyToken.(string)
			if seen[key] {
				return fmt.Errorf("duplicate property %q", key)
			}
			seen[key] = true
			if err := validateCapabilityJSONValue(decoder); err != nil {
				return err
			}
		}
	case '[':
		for decoder.More() {
			if err := validateCapabilityJSONValue(decoder); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("unexpected JSON delimiter %q", delim)
	}
	_, err = decoder.Token()
	return err
}

func validateCapabilityRewriteUnicode(raw []byte) error {
	if !utf8.Valid(raw) {
		return fmt.Errorf("request body is not valid UTF-8")
	}
	inString := false
	for i := 0; i < len(raw); i++ {
		switch raw[i] {
		case '"':
			inString = !inString
		case '\\':
			if !inString || i+1 >= len(raw) {
				continue
			}
			if raw[i+1] != 'u' {
				i++
				continue
			}
			value, ok := capabilityUnicodeEscapeValue(raw, i)
			if !ok {
				continue
			}
			switch {
			case value >= 0xd800 && value <= 0xdbff:
				low, paired := capabilityUnicodeEscapeValue(raw, i+6)
				if !paired || low < 0xdc00 || low > 0xdfff {
					return fmt.Errorf("request body contains an unpaired UTF-16 surrogate escape at byte %d", i)
				}
				i += 11
			case value >= 0xdc00 && value <= 0xdfff:
				return fmt.Errorf("request body contains an unpaired UTF-16 surrogate escape at byte %d", i)
			default:
				i += 5
			}
		}
	}
	return nil
}

func capabilityUnicodeEscapeValue(raw []byte, start int) (uint16, bool) {
	if start < 0 || start+6 > len(raw) || raw[start] != '\\' || raw[start+1] != 'u' {
		return 0, false
	}
	var value uint16
	for _, digit := range raw[start+2 : start+6] {
		value <<= 4
		switch {
		case digit >= '0' && digit <= '9':
			value |= uint16(digit - '0')
		case digit >= 'a' && digit <= 'f':
			value |= uint16(digit-'a') + 10
		case digit >= 'A' && digit <= 'F':
			value |= uint16(digit-'A') + 10
		default:
			return 0, false
		}
	}
	return value, true
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
