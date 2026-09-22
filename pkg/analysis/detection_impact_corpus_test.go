package analysis

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
)

type detectionImpactManifest struct {
	Version        string                      `json:"version"`
	BaselineDigest string                      `json:"baseline_digest"`
	Baseline       detectionImpactBaseline     `json:"baseline"`
	Cases          []detectionImpactCorpusCase `json:"cases"`
}

type detectionImpactBaseline struct {
	SyntaxCompleteCases          int `json:"syntax_complete_cases"`
	SemanticCompleteTargetStages int `json:"semantic_complete_target_stages"`
	NonmacroGapCount             int `json:"nonmacro_gap_count"`
}

type detectionImpactCorpusCase struct {
	ID                      string                                  `json:"id"`
	Query                   string                                  `json:"query"`
	SourceFamily            string                                  `json:"source_family"`
	SelectionRationale      string                                  `json:"selection_rationale"`
	Targets                 []detectionImpactTargetStage            `json:"targets"`
	Expected                detectionImpactExpectedFacts            `json:"expected"`
	HeldBoundaries          []detectionImpactHeldBoundary           `json:"held_boundaries"`
	BranchScopeExpectations []detectionImpactBranchScopeExpectation `json:"branch_scope_expectations,omitempty"`
}

type detectionImpactTargetStage struct {
	Command          string `json:"command"`
	Occurrence       int    `json:"occurrence"`
	SyntaxComplete   *bool  `json:"syntax_complete"`
	SemanticComplete *bool  `json:"semantic_complete"`
}

type detectionImpactStageSelector struct {
	Command    string `json:"command"`
	Occurrence int    `json:"occurrence"`
}

type detectionImpactBranchScopeExpectation struct {
	Branch           detectionImpactStageSelector   `json:"branch"`
	ParentScopeKind  string                         `json:"parent_scope_kind"`
	ChildScopeKind   string                         `json:"child_scope_kind"`
	ChildStages      []detectionImpactStageSelector `json:"child_stages"`
	DownstreamStages []detectionImpactStageSelector `json:"downstream_stages"`
}

type detectionImpactExpectedFacts struct {
	RequiredNormalizedReferenceNames []string                    `json:"required_normalized_reference_names"`
	RequiredDependencyPairs          []detectionImpactFactPair   `json:"required_dependency_pairs"`
	RequiredTransitionTargets        []string                    `json:"required_transition_targets"`
	RequiredRequirementPairs         []detectionImpactFactPair   `json:"required_requirement_pairs"`
	RequiredGapCodes                 []string                    `json:"required_gap_codes"`
	RequiredStageFacts               []detectionImpactStageFacts `json:"required_stage_facts"`
}

type detectionImpactStageFacts struct {
	Command                          string                    `json:"command"`
	Occurrence                       int                       `json:"occurrence"`
	RequiredNormalizedReferenceNames []string                  `json:"required_normalized_reference_names"`
	RequiredTransitionTargets        []string                  `json:"required_transition_targets"`
	RequiredRequirementPairs         []detectionImpactFactPair `json:"required_requirement_pairs"`
}

type detectionImpactHeldBoundary struct {
	Command          string   `json:"command"`
	Occurrence       int      `json:"occurrence"`
	RequiredGapCodes []string `json:"required_gap_codes"`
}

type detectionImpactFactPair struct {
	Kind string `json:"kind"`
	Name string `json:"name"`
}

func TestDetectionImpactCorpusBaseline(t *testing.T) {
	manifest := loadDetectionImpactManifest(t)
	results := analyzeDetectionImpactCorpus(t, manifest)
	assertDetectionImpactSelectors(t, manifest, results)

	want := detectionImpactBaseline{
		SyntaxCompleteCases:          12,
		SemanticCompleteTargetStages: 1,
		NonmacroGapCount:             35,
	}
	if manifest.Baseline != want {
		t.Fatalf("detection-impact bound baseline = %+v, want immutable pre-change baseline %+v", manifest.Baseline, want)
	}
	const wantDigest = "sha256:63aa3e808bad38edd3b0c6935c0f7e9ceb9d8f68e133866ba33ca0ce553954e1"
	if manifest.BaselineDigest != wantDigest {
		t.Fatalf("detection-impact baseline digest = %q, want %q", manifest.BaselineDigest, wantDigest)
	}
}

func TestDetectionImpactCorpusImproves(t *testing.T) {
	manifest := loadDetectionImpactManifest(t)
	results := analyzeDetectionImpactCorpus(t, manifest)
	assertDetectionImpactSelectors(t, manifest, results)

	for i, corpusCase := range manifest.Cases {
		result := results[i]
		for _, target := range corpusCase.Targets {
			target := target
			t.Run(corpusCase.ID+"/target/"+detectionImpactSelectorName(target.Command, target.Occurrence), func(t *testing.T) {
				stage := resolveDetectionImpactStage(t, result, target.Command, target.Occurrence)
				if actual := detectionImpactStageSyntaxComplete(result, stage); actual != *target.SyntaxComplete {
					t.Errorf("target syntax_complete = %t, want %t", actual, *target.SyntaxComplete)
				}
				if stage.SemanticComplete != *target.SemanticComplete {
					t.Errorf("target semantic_complete = %t, want %t", stage.SemanticComplete, *target.SemanticComplete)
				}
			})
		}
		t.Run(corpusCase.ID+"/facts", func(t *testing.T) {
			assertDetectionImpactFacts(t, result, corpusCase.Expected)
		})
		for _, boundary := range corpusCase.HeldBoundaries {
			boundary := boundary
			t.Run(corpusCase.ID+"/held/"+detectionImpactSelectorName(boundary.Command, boundary.Occurrence), func(t *testing.T) {
				stage := resolveDetectionImpactStage(t, result, boundary.Command, boundary.Occurrence)
				for _, code := range boundary.RequiredGapCodes {
					if !hasDetectionImpactStageGap(result, stage.ID, code) {
						t.Errorf("held boundary lacks gap %q", code)
					}
				}
			})
		}
	}

	actual := detectionImpactMetrics(manifest, results)
	if actual.SyntaxCompleteCases <= manifest.Baseline.SyntaxCompleteCases {
		t.Errorf("syntax-complete cases = %d, want > baseline %d", actual.SyntaxCompleteCases, manifest.Baseline.SyntaxCompleteCases)
	}
	if actual.SemanticCompleteTargetStages <= manifest.Baseline.SemanticCompleteTargetStages {
		t.Errorf("semantic-complete target stages = %d, want > baseline %d", actual.SemanticCompleteTargetStages, manifest.Baseline.SemanticCompleteTargetStages)
	}
	if actual.NonmacroGapCount >= manifest.Baseline.NonmacroGapCount {
		t.Errorf("non-macro gaps = %d, want < baseline %d", actual.NonmacroGapCount, manifest.Baseline.NonmacroGapCount)
	}
}

func TestDetectionImpactManifestRequiresTargetCompletenessProperties(t *testing.T) {
	data, err := os.ReadFile("../../testdata/analysis/detection-impact.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name, property string
		remove         []byte
	}{
		{name: "syntax", property: "syntax_complete", remove: []byte(`"syntax_complete": true, `)},
		{name: "semantic", property: "semantic_complete", remove: []byte(`, "semantic_complete": true`)},
	} {
		t.Run(test.name, func(t *testing.T) {
			modified := bytes.Replace(data, test.remove, nil, 1)
			if bytes.Equal(modified, data) {
				t.Fatalf("fixture does not contain removable %s property", test.property)
			}
			_, err := decodeDetectionImpactManifest(modified)
			if err == nil || !strings.Contains(err.Error(), `omits required `+test.property) {
				t.Fatalf("missing %s error = %v", test.property, err)
			}
		})
	}
}

func TestDetectionImpactBranchScopeAssertionsRejectBrokenOwnership(t *testing.T) {
	manifest := loadDetectionImpactManifest(t)
	results := analyzeDetectionImpactCorpus(t, manifest)
	checked := 0
	for i, corpusCase := range manifest.Cases {
		for _, expectation := range corpusCase.BranchScopeExpectations {
			checked++
			expectation := expectation
			result := results[i]
			name := corpusCase.ID + "/" + detectionImpactSelectorName(expectation.Branch.Command, expectation.Branch.Occurrence)
			t.Run(name+"/flattened_child", func(t *testing.T) {
				mutated := cloneDetectionImpactScopeResult(result)
				branch, _ := findDetectionImpactStage(mutated, expectation.Branch.Command, expectation.Branch.Occurrence)
				child := expectation.ChildStages[0]
				setDetectionImpactStageScope(t, mutated, child, branch.ScopeID)
				if err := validateDetectionImpactBranchScope(mutated, expectation); err == nil || !strings.Contains(err.Error(), "child stage") {
					t.Fatalf("flattened child error = %v", err)
				}
			})
			t.Run(name+"/wrong_owner", func(t *testing.T) {
				mutated := cloneDetectionImpactScopeResult(result)
				child, _ := findDetectionImpactStage(mutated, expectation.ChildStages[0].Command, expectation.ChildStages[0].Occurrence)
				for scopeIndex := range mutated.Scopes {
					if mutated.Scopes[scopeIndex].ID == child.ScopeID {
						mutated.Scopes[scopeIndex].StageID = "stage-wrong-owner"
					}
				}
				if err := validateDetectionImpactBranchScope(mutated, expectation); err == nil || !strings.Contains(err.Error(), "owned by") {
					t.Fatalf("wrong owner error = %v", err)
				}
			})
			t.Run(name+"/downstream_in_child", func(t *testing.T) {
				mutated := cloneDetectionImpactScopeResult(result)
				child, _ := findDetectionImpactStage(mutated, expectation.ChildStages[0].Command, expectation.ChildStages[0].Occurrence)
				setDetectionImpactStageScope(t, mutated, expectation.DownstreamStages[0], child.ScopeID)
				if err := validateDetectionImpactBranchScope(mutated, expectation); err == nil || !strings.Contains(err.Error(), "downstream stage") {
					t.Fatalf("downstream scope error = %v", err)
				}
			})
		}
	}
	if checked != 3 {
		t.Fatalf("checked %d branch-scope expectations, want 3", checked)
	}
}

func loadDetectionImpactManifest(t *testing.T) detectionImpactManifest {
	t.Helper()
	data, err := os.ReadFile("../../testdata/analysis/detection-impact.json")
	if err != nil {
		t.Fatal(err)
	}
	manifest, err := decodeDetectionImpactManifest(data)
	if err != nil {
		t.Fatal(err)
	}
	validateDetectionImpactManifest(t, manifest)
	return manifest
}

func decodeDetectionImpactManifest(data []byte) (detectionImpactManifest, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var manifest detectionImpactManifest
	if err := decoder.Decode(&manifest); err != nil {
		return detectionImpactManifest{}, err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return detectionImpactManifest{}, fmt.Errorf("detection-impact manifest has trailing JSON: %v", err)
	}
	for _, corpusCase := range manifest.Cases {
		for _, target := range corpusCase.Targets {
			selector := detectionImpactSelectorName(target.Command, target.Occurrence)
			if target.SyntaxComplete == nil {
				return detectionImpactManifest{}, fmt.Errorf("case %q target %s omits required syntax_complete", corpusCase.ID, selector)
			}
			if target.SemanticComplete == nil {
				return detectionImpactManifest{}, fmt.Errorf("case %q target %s omits required semantic_complete", corpusCase.ID, selector)
			}
		}
	}
	return manifest, nil
}

func validateDetectionImpactManifest(t *testing.T, manifest detectionImpactManifest) {
	t.Helper()
	if manifest.Version != "1" {
		t.Fatalf("detection-impact version = %q, want 1", manifest.Version)
	}
	if len(manifest.Cases) == 0 {
		t.Fatal("detection-impact manifest has no cases")
	}
	wantDigest := detectionImpactDigest(manifest.Cases)
	if manifest.BaselineDigest != wantDigest {
		t.Fatalf("detection-impact baseline_digest = %q, want %q", manifest.BaselineDigest, wantDigest)
	}
	seenCases := map[string]bool{}
	for _, corpusCase := range manifest.Cases {
		if corpusCase.ID == "" || strings.ContainsRune(corpusCase.ID, 0) || seenCases[corpusCase.ID] {
			t.Fatalf("invalid or duplicate detection-impact case id %q", corpusCase.ID)
		}
		seenCases[corpusCase.ID] = true
		if corpusCase.Query == "" || corpusCase.SourceFamily == "" || corpusCase.SelectionRationale == "" {
			t.Fatalf("case %q lacks query provenance", corpusCase.ID)
		}
		if len(corpusCase.Targets) == 0 || corpusCase.HeldBoundaries == nil {
			t.Fatalf("case %q lacks targets or held-boundary declaration", corpusCase.ID)
		}
		facts := corpusCase.Expected
		if facts.RequiredNormalizedReferenceNames == nil || facts.RequiredDependencyPairs == nil || facts.RequiredTransitionTargets == nil || facts.RequiredRequirementPairs == nil || facts.RequiredGapCodes == nil || facts.RequiredStageFacts == nil {
			t.Fatalf("case %q omits an expected-fact collection", corpusCase.ID)
		}
		seenTargets := map[string]bool{}
		for _, target := range corpusCase.Targets {
			key := detectionImpactSelectorName(target.Command, target.Occurrence)
			if target.Command == "" || target.Occurrence < 0 || target.SyntaxComplete == nil || target.SemanticComplete == nil || seenTargets[key] {
				t.Fatalf("case %q has invalid or duplicate target %q", corpusCase.ID, key)
			}
			seenTargets[key] = true
		}
		seenBoundaries := map[string]bool{}
		for _, boundary := range corpusCase.HeldBoundaries {
			key := detectionImpactSelectorName(boundary.Command, boundary.Occurrence)
			if boundary.Command == "" || boundary.Occurrence < 0 || len(boundary.RequiredGapCodes) == 0 || seenBoundaries[key] {
				t.Fatalf("case %q has invalid or duplicate held boundary %q", corpusCase.ID, key)
			}
			seenBoundaries[key] = true
		}
		validateDetectionImpactExpectedFacts(t, corpusCase.ID, facts)
		validateDetectionImpactBranchScopeExpectations(t, corpusCase)
	}
}

func validateDetectionImpactExpectedFacts(t *testing.T, caseID string, facts detectionImpactExpectedFacts) {
	t.Helper()
	seen := map[string]bool{}
	for _, name := range facts.RequiredNormalizedReferenceNames {
		if name == "" || seen["reference\x00"+name] {
			t.Fatalf("case %q has an invalid or duplicate required reference %q", caseID, name)
		}
		seen["reference\x00"+name] = true
	}
	for _, pair := range facts.RequiredDependencyPairs {
		key := "dependency\x00" + pair.Kind + "\x00" + pair.Name
		if pair.Kind == "" || pair.Name == "" || seen[key] {
			t.Fatalf("case %q has an invalid or duplicate required pair %+v", caseID, pair)
		}
		seen[key] = true
	}
	for _, pair := range facts.RequiredRequirementPairs {
		key := "requirement\x00" + pair.Kind + "\x00" + pair.Name
		if pair.Kind == "" || pair.Name == "" || seen[key] {
			t.Fatalf("case %q has an invalid or duplicate required pair %+v", caseID, pair)
		}
		seen[key] = true
	}
	for _, target := range facts.RequiredTransitionTargets {
		if target == "" || seen["transition\x00"+target] {
			t.Fatalf("case %q has an invalid or duplicate transition target %q", caseID, target)
		}
		seen["transition\x00"+target] = true
	}
	for _, code := range facts.RequiredGapCodes {
		if code == "" || seen["gap\x00"+code] {
			t.Fatalf("case %q has an invalid or duplicate gap code %q", caseID, code)
		}
		seen["gap\x00"+code] = true
	}
	for _, stageFacts := range facts.RequiredStageFacts {
		key := "stage\x00" + detectionImpactSelectorName(stageFacts.Command, stageFacts.Occurrence)
		if stageFacts.Command == "" || stageFacts.Occurrence < 0 || stageFacts.RequiredNormalizedReferenceNames == nil || stageFacts.RequiredTransitionTargets == nil || stageFacts.RequiredRequirementPairs == nil || seen[key] {
			t.Fatalf("case %q has invalid or duplicate required stage facts %q", caseID, key)
		}
		seen[key] = true
	}
}

func validateDetectionImpactBranchScopeExpectations(t *testing.T, corpusCase detectionImpactCorpusCase) {
	t.Helper()
	required := map[string]bool{}
	for _, target := range corpusCase.Targets {
		switch target.Command {
		case "join", "append", "appendpipe":
			required[detectionImpactSelectorName(target.Command, target.Occurrence)] = true
		}
	}
	seen := map[string]bool{}
	for _, expectation := range corpusCase.BranchScopeExpectations {
		key := detectionImpactSelectorName(expectation.Branch.Command, expectation.Branch.Occurrence)
		if !required[key] || expectation.Branch.Command == "" || expectation.Branch.Occurrence < 0 || expectation.ParentScopeKind == "" || expectation.ChildScopeKind == "" || len(expectation.ChildStages) == 0 || len(expectation.DownstreamStages) == 0 || seen[key] {
			t.Fatalf("case %q has invalid or duplicate branch-scope expectation %q", corpusCase.ID, key)
		}
		seen[key] = true
		for _, group := range [][]detectionImpactStageSelector{expectation.ChildStages, expectation.DownstreamStages} {
			for _, selector := range group {
				if selector.Command == "" || selector.Occurrence < 0 {
					t.Fatalf("case %q branch %q has invalid stage selector %+v", corpusCase.ID, key, selector)
				}
			}
		}
	}
	for key := range required {
		if !seen[key] {
			t.Fatalf("case %q omits branch-scope expectation %q", corpusCase.ID, key)
		}
	}
}

func detectionImpactDigest(cases []detectionImpactCorpusCase) string {
	hash := sha256.New()
	var queryLength [8]byte
	for _, corpusCase := range cases {
		_, _ = io.WriteString(hash, corpusCase.ID)
		_, _ = hash.Write([]byte{0})
		binary.BigEndian.PutUint64(queryLength[:], uint64(len([]byte(corpusCase.Query))))
		_, _ = hash.Write(queryLength[:])
		_, _ = io.WriteString(hash, corpusCase.Query)
	}
	return "sha256:" + hex.EncodeToString(hash.Sum(nil))
}

func analyzeDetectionImpactCorpus(t *testing.T, manifest detectionImpactManifest) []*Result {
	t.Helper()
	results := make([]*Result, 0, len(manifest.Cases))
	for _, corpusCase := range manifest.Cases {
		result, err := Analyze(QueryDocument{Text: corpusCase.Query})
		if err != nil {
			t.Fatalf("case %q: %v", corpusCase.ID, err)
		}
		results = append(results, result)
	}
	return results
}

func assertDetectionImpactSelectors(t *testing.T, manifest detectionImpactManifest, results []*Result) {
	t.Helper()
	for i, corpusCase := range manifest.Cases {
		for _, target := range corpusCase.Targets {
			resolveDetectionImpactStage(t, results[i], target.Command, target.Occurrence)
		}
		for _, boundary := range corpusCase.HeldBoundaries {
			resolveDetectionImpactStage(t, results[i], boundary.Command, boundary.Occurrence)
		}
		for _, expectation := range corpusCase.BranchScopeExpectations {
			if err := validateDetectionImpactBranchScope(results[i], expectation); err != nil {
				t.Errorf("case %q: %v", corpusCase.ID, err)
			}
		}
	}
}

func validateDetectionImpactBranchScope(result *Result, expectation detectionImpactBranchScopeExpectation) error {
	branch, found := findDetectionImpactStage(result, expectation.Branch.Command, expectation.Branch.Occurrence)
	if !found {
		return fmt.Errorf("branch stage %s does not resolve", detectionImpactSelectorName(expectation.Branch.Command, expectation.Branch.Occurrence))
	}
	parentScope, found := findDetectionImpactScope(result, branch.ScopeID)
	if !found {
		return fmt.Errorf("branch stage %s references missing parent scope %q", branch.ID, branch.ScopeID)
	}
	if parentScope.Kind != expectation.ParentScopeKind {
		return fmt.Errorf("branch stage %s parent scope kind = %q, want %q", branch.ID, parentScope.Kind, expectation.ParentScopeKind)
	}

	childStages := make([]Stage, 0, len(expectation.ChildStages))
	for _, selector := range expectation.ChildStages {
		stage, found := findDetectionImpactStage(result, selector.Command, selector.Occurrence)
		if !found {
			return fmt.Errorf("child stage %s does not resolve", detectionImpactSelectorName(selector.Command, selector.Occurrence))
		}
		childStages = append(childStages, stage)
	}
	childScopeID := childStages[0].ScopeID
	if childScopeID == branch.ScopeID {
		return fmt.Errorf("child stage %s was flattened into branch parent scope %q", childStages[0].ID, branch.ScopeID)
	}
	for _, stage := range childStages[1:] {
		if stage.ScopeID != childScopeID {
			return fmt.Errorf("child stage %s scope = %q, want shared child scope %q", stage.ID, stage.ScopeID, childScopeID)
		}
	}
	childScope, found := findDetectionImpactScope(result, childScopeID)
	if !found {
		return fmt.Errorf("child stage %s references missing scope %q", childStages[0].ID, childScopeID)
	}
	if childScope.ParentID != branch.ScopeID {
		return fmt.Errorf("child scope %q parent = %q, want branch parent scope %q", childScope.ID, childScope.ParentID, branch.ScopeID)
	}
	if childScope.StageID != branch.ID {
		return fmt.Errorf("child scope %q is owned by %q, want branch stage %q", childScope.ID, childScope.StageID, branch.ID)
	}
	if childScope.Kind != expectation.ChildScopeKind {
		return fmt.Errorf("child scope %q kind = %q, want %q", childScope.ID, childScope.Kind, expectation.ChildScopeKind)
	}
	for _, selector := range expectation.DownstreamStages {
		stage, found := findDetectionImpactStage(result, selector.Command, selector.Occurrence)
		if !found {
			return fmt.Errorf("downstream stage %s does not resolve", detectionImpactSelectorName(selector.Command, selector.Occurrence))
		}
		if stage.ScopeID != branch.ScopeID {
			return fmt.Errorf("downstream stage %s scope = %q, want branch parent scope %q", stage.ID, stage.ScopeID, branch.ScopeID)
		}
	}
	return nil
}

func findDetectionImpactScope(result *Result, scopeID string) (Scope, bool) {
	for _, scope := range result.Scopes {
		if scope.ID == scopeID {
			return scope, true
		}
	}
	return Scope{}, false
}

func cloneDetectionImpactScopeResult(result *Result) *Result {
	clone := *result
	clone.Stages = append([]Stage{}, result.Stages...)
	clone.Scopes = append([]Scope{}, result.Scopes...)
	return &clone
}

func setDetectionImpactStageScope(t *testing.T, result *Result, selector detectionImpactStageSelector, scopeID string) {
	t.Helper()
	stage, found := findDetectionImpactStage(result, selector.Command, selector.Occurrence)
	if !found {
		t.Fatalf("stage %s does not resolve", detectionImpactSelectorName(selector.Command, selector.Occurrence))
	}
	for stageIndex := range result.Stages {
		if result.Stages[stageIndex].ID == stage.ID {
			result.Stages[stageIndex].ScopeID = scopeID
			return
		}
	}
	t.Fatalf("stage %q disappeared", stage.ID)
}

func resolveDetectionImpactStage(t *testing.T, result *Result, command string, occurrence int) Stage {
	t.Helper()
	seen := 0
	for _, stage := range result.Stages {
		if stage.Command != command {
			continue
		}
		if seen == occurrence {
			return stage
		}
		seen++
	}
	t.Fatalf("target %s does not resolve exactly; found %d occurrence(s)", detectionImpactSelectorName(command, occurrence), seen)
	return Stage{}
}

func detectionImpactSelectorName(command string, occurrence int) string {
	return fmt.Sprintf("%s[%d]", command, occurrence)
}

func detectionImpactStageSyntaxComplete(result *Result, stage Stage) bool {
	for _, diagnostic := range result.Diagnostics {
		if diagnostic.Code != CodeSyntaxError {
			continue
		}
		if diagnostic.StageID == stage.ID || (diagnostic.StageID == "" && detectionImpactLocationsOverlap(diagnostic.Location, stage.Location)) {
			return false
		}
	}
	return true
}

func detectionImpactLocationsOverlap(left, right Location) bool {
	return left.Start.Offset < right.End.Offset && right.Start.Offset < left.End.Offset
}

func detectionImpactMetrics(manifest detectionImpactManifest, results []*Result) detectionImpactBaseline {
	metrics := detectionImpactBaseline{}
	type gapKey struct {
		caseID, code, stageID string
		location              Location
	}
	gaps := map[gapKey]bool{}
	for i, corpusCase := range manifest.Cases {
		result := results[i]
		if result.Coverage.SyntaxComplete {
			metrics.SyntaxCompleteCases++
		}
		for _, target := range corpusCase.Targets {
			stage, found := findDetectionImpactStage(result, target.Command, target.Occurrence)
			if found && stage.SemanticComplete {
				metrics.SemanticCompleteTargetStages++
			}
		}
		for _, diagnostic := range result.Diagnostics {
			if !detectionImpactSourceLocated(corpusCase.Query, diagnostic) || !detectionImpactDiagnosticMakesSemanticsIncomplete(result, diagnostic) || detectionImpactMacroGap(result, diagnostic) {
				continue
			}
			gaps[gapKey{caseID: corpusCase.ID, code: diagnostic.Code, stageID: diagnostic.StageID, location: diagnostic.Location}] = true
		}
	}
	metrics.NonmacroGapCount = len(gaps)
	return metrics
}

func findDetectionImpactStage(result *Result, command string, occurrence int) (Stage, bool) {
	seen := 0
	for _, stage := range result.Stages {
		if stage.Command != command {
			continue
		}
		if seen == occurrence {
			return stage, true
		}
		seen++
	}
	return Stage{}, false
}

func detectionImpactDiagnosticMakesSemanticsIncomplete(result *Result, diagnostic Diagnostic) bool {
	if result.Coverage.SemanticComplete || diagnostic.Code == CodeUnavailableField {
		return false
	}
	if diagnostic.StageID == "" {
		return true
	}
	for _, stage := range result.Stages {
		if stage.ID == diagnostic.StageID {
			return !stage.SemanticComplete
		}
	}
	return false
}

func detectionImpactSourceLocated(query string, diagnostic Diagnostic) bool {
	return diagnostic.Location.Start.Offset >= 0 &&
		diagnostic.Location.End.Offset >= diagnostic.Location.Start.Offset &&
		diagnostic.Location.End.Offset <= len([]byte(query)) &&
		diagnostic.Location.Start.Line > 0 && diagnostic.Location.Start.Column > 0 &&
		diagnostic.Location.End.Line > 0 && diagnostic.Location.End.Column > 0
}

func detectionImpactMacroGap(result *Result, diagnostic Diagnostic) bool {
	if diagnostic.Code != CodeDynamicReference {
		return false
	}
	for _, reference := range result.References {
		if reference.Kind == "macro" && reference.StageID == diagnostic.StageID && reference.ScopeID == diagnostic.ScopeID && detectionImpactLocationContains(diagnostic.Location, reference.Location) {
			return true
		}
	}
	return false
}

func detectionImpactLocationContains(outer, inner Location) bool {
	return outer.Start.Offset <= inner.Start.Offset && outer.End.Offset >= inner.End.Offset
}

func assertDetectionImpactFacts(t *testing.T, result *Result, expected detectionImpactExpectedFacts) {
	t.Helper()
	for _, name := range expected.RequiredNormalizedReferenceNames {
		found := false
		for _, reference := range result.References {
			found = found || reference.NormalizedName == name
		}
		if !found {
			t.Errorf("missing normalized reference %q", name)
		}
	}
	for _, pair := range expected.RequiredDependencyPairs {
		if !hasDetectionImpactDependency(result.Dependencies, pair) {
			t.Errorf("missing dependency %s:%s", pair.Kind, pair.Name)
		}
	}
	for _, target := range expected.RequiredTransitionTargets {
		found := false
		for _, lineage := range result.Lineage {
			for _, transition := range lineage.Transitions {
				found = found || transition.Output == target
			}
		}
		if !found {
			t.Errorf("missing transition target %q", target)
		}
	}
	for _, pair := range expected.RequiredRequirementPairs {
		found := false
		for _, item := range result.Requirements.Items {
			found = found || (item.Kind == pair.Kind && item.Identity == pair.Name)
		}
		if !found {
			t.Errorf("missing requirement %s:%s", pair.Kind, pair.Name)
		}
	}
	for _, code := range expected.RequiredGapCodes {
		found := false
		for _, diagnostic := range result.Diagnostics {
			found = found || diagnostic.Code == code
		}
		if !found {
			t.Errorf("missing analysis gap %q", code)
		}
	}
	for _, expectedStage := range expected.RequiredStageFacts {
		stage := resolveDetectionImpactStage(t, result, expectedStage.Command, expectedStage.Occurrence)
		for _, name := range expectedStage.RequiredNormalizedReferenceNames {
			found := false
			for _, reference := range result.References {
				found = found || (reference.StageID == stage.ID && reference.NormalizedName == name)
			}
			if !found {
				t.Errorf("stage %s lacks normalized reference %q", detectionImpactSelectorName(expectedStage.Command, expectedStage.Occurrence), name)
			}
		}
		for _, target := range expectedStage.RequiredTransitionTargets {
			found := false
			for _, lineage := range result.Lineage {
				if lineage.StageID != stage.ID {
					continue
				}
				for _, transition := range lineage.Transitions {
					found = found || transition.Output == target
				}
			}
			if !found {
				t.Errorf("stage %s lacks transition target %q", detectionImpactSelectorName(expectedStage.Command, expectedStage.Occurrence), target)
			}
		}
		for _, pair := range expectedStage.RequiredRequirementPairs {
			found := false
			for _, item := range result.Requirements.Items {
				if item.Kind != pair.Kind || item.Identity != pair.Name {
					continue
				}
				for _, occurrence := range item.Occurrences {
					found = found || occurrence.StageID == stage.ID
				}
			}
			if !found {
				t.Errorf("stage %s lacks requirement %s:%s", detectionImpactSelectorName(expectedStage.Command, expectedStage.Occurrence), pair.Kind, pair.Name)
			}
		}
	}
}

func hasDetectionImpactDependency(dependencies Dependencies, pair detectionImpactFactPair) bool {
	var values []string
	switch pair.Kind {
	case "index":
		values = dependencies.Indexes
	case "source":
		values = dependencies.Sources
	case "source_type":
		values = dependencies.SourceTypes
	case "dataset":
		values = dependencies.Datasets
	case "lookup":
		values = dependencies.Lookups
	case "data_model":
		values = dependencies.DataModels
	case "macro":
		values = dependencies.Macros
	default:
		return false
	}
	for _, value := range values {
		if value == pair.Name {
			return true
		}
	}
	return false
}

func hasDetectionImpactStageGap(result *Result, stageID, code string) bool {
	for _, diagnostic := range result.Diagnostics {
		if diagnostic.StageID == stageID && diagnostic.Code == code {
			return true
		}
	}
	return false
}
