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
	ID                 string                        `json:"id"`
	Query              string                        `json:"query"`
	SourceFamily       string                        `json:"source_family"`
	SelectionRationale string                        `json:"selection_rationale"`
	Targets            []detectionImpactTargetStage  `json:"targets"`
	Expected           detectionImpactExpectedFacts  `json:"expected"`
	HeldBoundaries     []detectionImpactHeldBoundary `json:"held_boundaries"`
}

type detectionImpactTargetStage struct {
	Command          string `json:"command"`
	Occurrence       int    `json:"occurrence"`
	SyntaxComplete   bool   `json:"syntax_complete"`
	SemanticComplete bool   `json:"semantic_complete"`
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

	actual := detectionImpactMetrics(manifest, results)
	if actual != manifest.Baseline {
		t.Fatalf("detection-impact baseline = %+v, want %+v", actual, manifest.Baseline)
	}
}

func TestDetectionImpactCorpusImproves(t *testing.T) {
	manifest := loadDetectionImpactManifest(t)
	results := analyzeDetectionImpactCorpus(t, manifest)
	allExpectedFactsPresent := true

	for i, corpusCase := range manifest.Cases {
		result := results[i]
		for _, target := range corpusCase.Targets {
			target := target
			if !t.Run(corpusCase.ID+"/target/"+detectionImpactSelectorName(target.Command, target.Occurrence), func(t *testing.T) {
				stage := resolveDetectionImpactStage(t, result, target.Command, target.Occurrence)
				if actual := detectionImpactStageSyntaxComplete(result, stage); actual != target.SyntaxComplete {
					t.Errorf("target syntax_complete = %t, want %t", actual, target.SyntaxComplete)
				}
				if stage.SemanticComplete != target.SemanticComplete {
					t.Errorf("target semantic_complete = %t, want %t", stage.SemanticComplete, target.SemanticComplete)
				}
			}) {
				allExpectedFactsPresent = false
			}
		}
		if !t.Run(corpusCase.ID+"/facts", func(t *testing.T) {
			assertDetectionImpactFacts(t, result, corpusCase.Expected)
		}) {
			allExpectedFactsPresent = false
		}
		for _, boundary := range corpusCase.HeldBoundaries {
			boundary := boundary
			if !t.Run(corpusCase.ID+"/held/"+detectionImpactSelectorName(boundary.Command, boundary.Occurrence), func(t *testing.T) {
				stage := resolveDetectionImpactStage(t, result, boundary.Command, boundary.Occurrence)
				for _, code := range boundary.RequiredGapCodes {
					if !hasDetectionImpactStageGap(result, stage.ID, code) {
						t.Errorf("held boundary lacks gap %q", code)
					}
				}
			}) {
				allExpectedFactsPresent = false
			}
		}
	}

	// The strict aggregate comparison becomes useful only after every reviewed
	// target and fact is present. Before semantic implementation, the failures
	// above identify the exact missing contract instead of reporting only counts.
	if !allExpectedFactsPresent {
		return
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

func loadDetectionImpactManifest(t *testing.T) detectionImpactManifest {
	t.Helper()
	data, err := os.ReadFile("../../testdata/analysis/detection-impact.json")
	if err != nil {
		t.Fatal(err)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var manifest detectionImpactManifest
	if err := decoder.Decode(&manifest); err != nil {
		t.Fatal(err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		t.Fatalf("detection-impact manifest has trailing JSON: %v", err)
	}
	validateDetectionImpactManifest(t, manifest)
	return manifest
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
			if target.Command == "" || target.Occurrence < 0 || seenTargets[key] {
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
	}
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
