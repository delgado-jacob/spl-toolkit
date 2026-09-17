package analysis

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
)

type capabilityLedgerFile struct {
	SchemaVersion int                `json:"schema_version"`
	Records       []CapabilityRecord `json:"records"`
}

type capabilityCorpusFile struct {
	SchemaVersion int                  `json:"schema_version"`
	Cases         []CapabilityEvidence `json:"cases"`
}

var capabilityRecordKinds = map[string]struct{}{
	"command":      {},
	"function":     {},
	"expression":   {},
	"lexical_form": {},
	"pipeline":     {},
	"subsearch":    {},
	"dataset":      {},
	"macro":        {},
	"module":       {},
	"namespace":    {},
	"variable":     {},
	"annotation":   {},
	"profile_form": {},
}

var capabilityProvenanceFamilies = map[string]struct{}{
	"toolkit":            {},
	"spl2":               {},
	"splunk_analytics":   {},
	"security_detection": {},
}

func decodeCapabilityAssets(ledgerJSON, corpusJSON []byte) ([]CapabilityRecord, []CapabilityEvidence, error) {
	var ledger capabilityLedgerFile
	if err := decodeCapabilityJSON("capability ledger", ledgerJSON, &ledger); err != nil {
		return nil, nil, err
	}
	var corpus capabilityCorpusFile
	if err := decodeCapabilityJSON("capability corpus", corpusJSON, &corpus); err != nil {
		return nil, nil, err
	}
	if err := validateCapabilityDimensionPresence(ledgerJSON); err != nil {
		return nil, nil, err
	}
	if err := validateCapabilityAssets(ledger, corpus); err != nil {
		return nil, nil, err
	}
	return ledger.Records, corpus.Cases, nil
}

func decodeCapabilityJSON(name string, raw []byte, destination any) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return fmt.Errorf("decode %s: %w", name, err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return fmt.Errorf("decode %s: trailing JSON value", name)
		}
		return fmt.Errorf("decode %s: trailing data: %w", name, err)
	}
	return nil
}

func validateCapabilityDimensionPresence(raw []byte) error {
	type dimensionPresence struct {
		Syntax        json.RawMessage `json:"syntax"`
		Semantics     json.RawMessage `json:"semantics"`
		Requirements  json.RawMessage `json:"requirements"`
		Linting       json.RawMessage `json:"linting"`
		SafeRewriting json.RawMessage `json:"safe_rewriting"`
	}
	type recordPresence struct {
		Dimensions json.RawMessage `json:"dimensions"`
	}
	var ledger struct {
		Records []recordPresence `json:"records"`
	}
	if err := json.Unmarshal(raw, &ledger); err != nil {
		return fmt.Errorf("inspect capability dimensions: %w", err)
	}
	for i, record := range ledger.Records {
		if missingJSONValue(record.Dimensions) {
			return fmt.Errorf("record %d: missing dimensions", i)
		}
		var dimensions dimensionPresence
		if err := json.Unmarshal(record.Dimensions, &dimensions); err != nil {
			return fmt.Errorf("record %d dimensions: %w", i, err)
		}
		for _, dimension := range []struct {
			name string
			raw  json.RawMessage
		}{
			{"syntax", dimensions.Syntax},
			{"semantics", dimensions.Semantics},
			{"requirements", dimensions.Requirements},
			{"linting", dimensions.Linting},
			{"safe_rewriting", dimensions.SafeRewriting},
		} {
			if missingJSONValue(dimension.raw) {
				return fmt.Errorf("record %d: missing dimension %q", i, dimension.name)
			}
		}
	}
	return nil
}

func missingJSONValue(raw json.RawMessage) bool {
	trimmed := bytes.TrimSpace(raw)
	return len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null"))
}

func validateCapabilityAssets(ledger capabilityLedgerFile, corpus capabilityCorpusFile) error {
	if ledger.SchemaVersion != 1 {
		return fmt.Errorf("capability ledger schema_version must be 1, got %d", ledger.SchemaVersion)
	}
	if corpus.SchemaVersion != 1 {
		return fmt.Errorf("capability corpus schema_version must be 1, got %d", corpus.SchemaVersion)
	}
	if ledger.Records == nil {
		return fmt.Errorf("capability ledger records must be an array")
	}
	if corpus.Cases == nil {
		return fmt.Errorf("capability corpus cases must be an array")
	}

	evidenceByID := make(map[string]CapabilityEvidence, len(corpus.Cases))
	for i, evidence := range corpus.Cases {
		if err := validateCapabilityEvidence(evidence); err != nil {
			return fmt.Errorf("evidence case %d: %w", i, err)
		}
		if _, exists := evidenceByID[evidence.ID]; exists {
			return fmt.Errorf("duplicate evidence ID %q", evidence.ID)
		}
		evidenceByID[evidence.ID] = evidence
	}
	if err := validateCapabilityEvidenceOrder(corpus.Cases); err != nil {
		return err
	}

	recordIDs := make(map[string]struct{}, len(ledger.Records))
	recordKindsByLanguage := make(map[string]map[string]struct{}, 2)
	referencedEvidence := make(map[string]struct{}, len(corpus.Cases))
	for i, record := range ledger.Records {
		if err := validateCapabilityRecord(record, evidenceByID, referencedEvidence); err != nil {
			return fmt.Errorf("capability record %d: %w", i, err)
		}
		if _, exists := recordIDs[record.ID]; exists {
			return fmt.Errorf("duplicate record ID %q", record.ID)
		}
		recordIDs[record.ID] = struct{}{}
		if recordKindsByLanguage[record.Language] == nil {
			recordKindsByLanguage[record.Language] = make(map[string]struct{}, len(capabilityRecordKinds))
		}
		recordKindsByLanguage[record.Language][record.Kind] = struct{}{}
	}
	if err := validateCapabilityKindCoverage(recordKindsByLanguage); err != nil {
		return err
	}
	if err := validateCapabilityRecordOrder(ledger.Records); err != nil {
		return err
	}
	for _, evidence := range corpus.Cases {
		if _, referenced := referencedEvidence[evidence.ID]; !referenced {
			return fmt.Errorf("evidence %q is not referenced by a capability claim", evidence.ID)
		}
	}
	return nil
}

func validateCapabilityKindCoverage(kindsByLanguage map[string]map[string]struct{}) error {
	for _, language := range []string{"spl", "spl2"} {
		present := kindsByLanguage[language]
		missing := make([]string, 0, len(capabilityRecordKinds)-len(present))
		for kind := range capabilityRecordKinds {
			if _, exists := present[kind]; !exists {
				missing = append(missing, kind)
			}
		}
		if len(missing) != 0 {
			sort.Strings(missing)
			return fmt.Errorf("language %q is missing record kinds: %s", language, strings.Join(missing, ", "))
		}
	}
	return nil
}

func validateCapabilityRecord(record CapabilityRecord, evidence map[string]CapabilityEvidence, referenced map[string]struct{}) error {
	if strings.TrimSpace(record.ID) == "" {
		return fmt.Errorf("record ID must not be empty")
	}
	if !validCapabilityLanguage(record.Language) {
		return fmt.Errorf("unsupported language %q", record.Language)
	}
	if record.Profile != "splunkd" {
		return fmt.Errorf("unsupported profile %q", record.Profile)
	}
	if _, ok := capabilityRecordKinds[record.Kind]; !ok {
		return fmt.Errorf("unsupported record kind %q", record.Kind)
	}
	if strings.TrimSpace(record.Name) == "" {
		return fmt.Errorf("record name must not be empty")
	}
	if strings.TrimSpace(record.Form) == "" {
		return fmt.Errorf("record form must not be empty")
	}
	if err := validateCapabilityProvenance(record.Provenance); err != nil {
		return fmt.Errorf("provenance: %w", err)
	}
	for _, dimension := range capabilityDimensionClaims(record.Dimensions) {
		if err := validateCapabilityClaim(dimension.name, record, dimension.claim, evidence); err != nil {
			return fmt.Errorf("%s: %w", dimension.name, err)
		}
		for _, evidenceID := range dimension.claim.EvidenceIDs {
			referenced[evidenceID] = struct{}{}
		}
	}
	return nil
}

func validateCapabilityEvidence(evidence CapabilityEvidence) error {
	if strings.TrimSpace(evidence.ID) == "" {
		return fmt.Errorf("evidence ID must not be empty")
	}
	switch evidence.Classification {
	case CapabilityEvidencePositive, CapabilityEvidenceNegative, CapabilityEvidenceIncomplete:
	default:
		return fmt.Errorf("unsupported classification %q", evidence.Classification)
	}
	if !validCapabilityLanguage(evidence.Document.Language) {
		return fmt.Errorf("unsupported document language %q", evidence.Document.Language)
	}
	if evidence.Document.Profile != "splunkd" {
		return fmt.Errorf("unsupported document profile %q", evidence.Document.Profile)
	}
	if evidence.Document.Version != "current" {
		return fmt.Errorf("unsupported document version %q", evidence.Document.Version)
	}
	wantSourceID := "capability:" + evidence.ID
	if evidence.Document.SourceID != wantSourceID {
		return fmt.Errorf("document source_id must be %q", wantSourceID)
	}
	if err := validateCapabilityEvidenceObservations(evidence); err != nil {
		return fmt.Errorf("observations: %w", err)
	}
	if err := validateCapabilityProvenance(evidence.Provenance); err != nil {
		return fmt.Errorf("provenance: %w", err)
	}
	return nil
}

func validateCapabilityProvenance(provenance CapabilityProvenance) error {
	if _, ok := capabilityProvenanceFamilies[provenance.SourceFamily]; !ok {
		return fmt.Errorf("unsupported source family %q", provenance.SourceFamily)
	}
	if strings.TrimSpace(provenance.Note) == "" {
		return fmt.Errorf("note must not be empty")
	}
	return nil
}

func validCapabilityLanguage(language string) bool {
	return language == "spl" || language == "spl2"
}

func validateCapabilityClaim(dimension string, record CapabilityRecord, claim CapabilityClaim, evidenceByID map[string]CapabilityEvidence) error {
	switch claim.State {
	case CapabilitySupported, CapabilityPartial, CapabilityUnsupported, CapabilityNotApplicable, CapabilityUnassessed:
	default:
		return fmt.Errorf("unsupported state %q", claim.State)
	}
	if err := validateNonemptyStrings("limitation", claim.Limitations); err != nil {
		return err
	}

	classifications := make(map[CapabilityEvidenceClassification]bool, 3)
	seenEvidence := make(map[string]struct{}, len(claim.EvidenceIDs))
	for _, evidenceID := range claim.EvidenceIDs {
		if strings.TrimSpace(evidenceID) == "" {
			return fmt.Errorf("evidence ID must not be empty")
		}
		if _, duplicate := seenEvidence[evidenceID]; duplicate {
			return fmt.Errorf("duplicate evidence ID %q in claim", evidenceID)
		}
		seenEvidence[evidenceID] = struct{}{}
		evidence, exists := evidenceByID[evidenceID]
		if !exists {
			return fmt.Errorf("unknown evidence ID %q", evidenceID)
		}
		if evidence.Document.Language != record.Language || evidence.Document.Profile != record.Profile {
			return fmt.Errorf("evidence %q selectors %s/%s do not match record selectors %s/%s", evidenceID, evidence.Document.Language, evidence.Document.Profile, record.Language, record.Profile)
		}
		if !capabilityEvidenceHasObservation(evidence, dimension) {
			return fmt.Errorf("evidence %q has no %s observation", evidenceID, dimension)
		}
		if err := validateCapabilityEvidenceObservation(evidence, dimension); err != nil {
			return fmt.Errorf("evidence %q has invalid %s observation: %w", evidenceID, dimension, err)
		}
		if dimension == "linting" && evidence.Classification == CapabilityEvidenceNegative {
			if err := validateExactLintingDiagnostics(evidence.Observations.Linting.Diagnostics); err != nil {
				return fmt.Errorf("negative evidence %q has no exact linting diagnostic observation: %w", evidenceID, err)
			}
		}
		classifications[evidence.Classification] = true
	}

	switch claim.State {
	case CapabilitySupported:
		negativeLintEvidence := dimension == "linting" && classifications[CapabilityEvidenceNegative]
		if len(claim.EvidenceIDs) == 0 || (!classifications[CapabilityEvidencePositive] && !negativeLintEvidence) {
			return fmt.Errorf("supported claim requires positive evidence or exact negative linting evidence")
		}
		if (classifications[CapabilityEvidenceNegative] && dimension != "linting") || classifications[CapabilityEvidenceIncomplete] {
			return fmt.Errorf("supported claim accepts only positive evidence except for exact negative linting evidence")
		}
	case CapabilityPartial:
		if len(claim.Limitations) == 0 {
			return fmt.Errorf("partial claim requires a limitation")
		}
		if !classifications[CapabilityEvidencePositive] || !classifications[CapabilityEvidenceIncomplete] {
			return fmt.Errorf("partial claim requires positive and incomplete evidence")
		}
		if classifications[CapabilityEvidenceNegative] {
			return fmt.Errorf("partial claim does not accept negative evidence")
		}
	case CapabilityUnsupported:
		if len(claim.Limitations) == 0 {
			return fmt.Errorf("unsupported claim requires a limitation")
		}
		if !classifications[CapabilityEvidenceNegative] && !classifications[CapabilityEvidenceIncomplete] {
			return fmt.Errorf("unsupported claim requires negative or incomplete evidence")
		}
		if classifications[CapabilityEvidencePositive] {
			return fmt.Errorf("unsupported claim does not accept positive evidence")
		}
	case CapabilityNotApplicable:
		if len(claim.EvidenceIDs) != 0 {
			return fmt.Errorf("not_applicable claim must not have evidence")
		}
		if len(claim.Limitations) == 0 {
			return fmt.Errorf("not_applicable claim requires a reason")
		}
	case CapabilityUnassessed:
		if len(claim.EvidenceIDs) != 0 || len(claim.Limitations) != 0 {
			return fmt.Errorf("unassessed claim must have empty evidence and limitations")
		}
	}
	return nil
}

func validateCapabilityEvidenceObservations(evidence CapabilityEvidence) error {
	observed := false
	for _, dimension := range []string{"syntax", "semantics", "requirements", "linting", "safe_rewriting"} {
		if !capabilityEvidenceHasObservation(evidence, dimension) {
			continue
		}
		observed = true
		if err := validateCapabilityEvidenceObservation(evidence, dimension); err != nil {
			return fmt.Errorf("%s: %w", dimension, err)
		}
	}
	if !observed {
		return fmt.Errorf("at least one typed observation is required")
	}
	return nil
}

func validateCapabilityEvidenceObservation(evidence CapabilityEvidence, dimension string) error {
	positive := evidence.Classification == CapabilityEvidencePositive
	switch dimension {
	case "syntax":
		observation := evidence.Observations.Syntax
		if positive && !observation.Complete {
			return fmt.Errorf("positive syntax evidence must be complete")
		}
		return validateCapabilityDiagnostics(observation.Diagnostics, false)
	case "semantics":
		observation := evidence.Observations.Semantics
		if err := validateCapabilityStatus(observation.Status); err != nil {
			return err
		}
		if observation.Complete != (observation.Status == Valid) {
			return fmt.Errorf("semantic completeness must agree with status")
		}
		if positive && (observation.Status != Valid || !observation.Complete) {
			return fmt.Errorf("positive semantics evidence must be valid and complete")
		}
		if len(observation.Stages)+len(observation.References)+len(observation.Dependencies)+len(observation.Transitions)+len(observation.Diagnostics) == 0 {
			return fmt.Errorf("at least one typed semantic fact is required")
		}
		for i, stage := range observation.Stages {
			if strings.TrimSpace(stage.Command) == "" {
				return fmt.Errorf("stage %d requires a command", i)
			}
			if positive && !stage.SemanticComplete {
				return fmt.Errorf("positive stage %d must be semantically complete", i)
			}
		}
		for i, reference := range observation.References {
			if err := requireCapabilityFields("reference", reference.NormalizedName, reference.Kind, reference.Role, reference.Resolution, reference.Binding); err != nil {
				return fmt.Errorf("reference %d: %w", i, err)
			}
			if err := validateCapabilityLocation(reference.Location); err != nil {
				return fmt.Errorf("reference %d: %w", i, err)
			}
		}
		for i, dependency := range observation.Dependencies {
			if err := requireCapabilityFields("dependency", dependency.Kind, dependency.Name); err != nil {
				return fmt.Errorf("dependency %d: %w", i, err)
			}
		}
		for i, transition := range observation.Transitions {
			if err := requireCapabilityFields("transition", transition.Operation, transition.Output); err != nil {
				return fmt.Errorf("transition %d: %w", i, err)
			}
		}
		return validateCapabilityDiagnostics(observation.Diagnostics, false)
	case "requirements":
		observation := evidence.Observations.Requirements
		if err := validateCapabilityStatus(observation.QueryStatus); err != nil {
			return err
		}
		if observation.Complete != (observation.QueryStatus == Valid) {
			return fmt.Errorf("requirements completeness must agree with status")
		}
		if positive && (observation.QueryStatus != Valid || !observation.Complete) {
			return fmt.Errorf("positive requirements evidence must be valid and complete")
		}
		if len(observation.Items)+len(observation.GapCodes) == 0 {
			return fmt.Errorf("at least one typed requirement fact is required")
		}
		if observation.Complete && len(observation.GapCodes) != 0 {
			return fmt.Errorf("complete requirements observation must not contain gap codes")
		}
		for i, item := range observation.Items {
			if err := requireCapabilityFields("requirement item", item.Kind, item.Identity, item.Role, item.Necessity, item.Resolution); err != nil {
				return fmt.Errorf("item %d: %w", i, err)
			}
		}
		return validateNonemptyStrings("gap code", observation.GapCodes)
	case "linting":
		return validateCapabilityDiagnostics(evidence.Observations.Linting.Diagnostics, true)
	case "safe_rewriting":
		observation := evidence.Observations.SafeRewriting
		if err := validateCapabilityStatus(observation.Status); err != nil {
			return err
		}
		if observation.RewriteComplete != (observation.Status == Valid) {
			return fmt.Errorf("rewrite completeness must agree with status")
		}
		if positive && (observation.Status != Valid || !observation.RewriteComplete) {
			return fmt.Errorf("positive rewrite evidence must be valid and complete")
		}
		if strings.TrimSpace(observation.Text) == "" || strings.TrimSpace(observation.CandidateText) == "" {
			return fmt.Errorf("rewrite observation requires text and candidate_text")
		}
		if observation.Committed && observation.Text != observation.CandidateText {
			return fmt.Errorf("committed rewrite text must equal candidate_text")
		}
		for _, field := range []struct {
			label   string
			reasons []string
		}{
			{"coverage reason", observation.CoverageReasons},
			{"change reason", observation.ChangeReasons},
			{"rule evaluation reason", observation.RuleEvaluationReasons},
		} {
			if err := validateNonemptyStrings(field.label, field.reasons); err != nil {
				return err
			}
		}
		return nil
	default:
		return fmt.Errorf("unknown observation dimension %q", dimension)
	}
}

func validateCapabilityStatus(status Status) error {
	switch status {
	case Valid, Invalid, Incomplete:
		return nil
	default:
		return fmt.Errorf("unsupported status %q", status)
	}
}

func requireCapabilityFields(label string, values ...string) error {
	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s fields must not be empty", label)
		}
	}
	return nil
}

func validateExactLintingDiagnostics(diagnostics []CapabilityDiagnosticExpectation) error {
	return validateCapabilityDiagnostics(diagnostics, true)
}

func validateCapabilityDiagnostics(diagnostics []CapabilityDiagnosticExpectation, required bool) error {
	if required && len(diagnostics) == 0 {
		return fmt.Errorf("at least one diagnostic is required")
	}
	for i, diagnostic := range diagnostics {
		if strings.TrimSpace(diagnostic.Code) == "" || strings.TrimSpace(diagnostic.Category) == "" || strings.TrimSpace(diagnostic.Severity) == "" {
			return fmt.Errorf("diagnostic %d requires code, category, and severity", i)
		}
		if err := validateCapabilityLocation(diagnostic.Location); err != nil {
			return fmt.Errorf("diagnostic %d: %w", i, err)
		}
	}
	return nil
}

func validateCapabilityLocation(location Location) error {
	if location.Start.Offset < 0 || location.Start.Line <= 0 || location.Start.Column <= 0 ||
		location.End.Offset <= location.Start.Offset || location.End.Line < location.Start.Line || location.End.Column <= 0 ||
		(location.End.Line == location.Start.Line && location.End.Column <= location.Start.Column) {
		return fmt.Errorf("complete nonempty source location is required")
	}
	return nil
}

func validateNonemptyStrings(label string, values []string) error {
	for i, value := range values {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s %d must not be empty", label, i)
		}
	}
	return nil
}

func capabilityEvidenceHasObservation(evidence CapabilityEvidence, dimension string) bool {
	switch dimension {
	case "syntax":
		return evidence.Observations.Syntax != nil
	case "semantics":
		return evidence.Observations.Semantics != nil
	case "requirements":
		return evidence.Observations.Requirements != nil
	case "linting":
		return evidence.Observations.Linting != nil
	case "safe_rewriting":
		return evidence.Observations.SafeRewriting != nil
	default:
		return false
	}
}

func capabilityDimensionClaims(dimensions CapabilityDimensions) []struct {
	name  string
	claim CapabilityClaim
} {
	return []struct {
		name  string
		claim CapabilityClaim
	}{
		{"syntax", dimensions.Syntax},
		{"semantics", dimensions.Semantics},
		{"requirements", dimensions.Requirements},
		{"linting", dimensions.Linting},
		{"safe_rewriting", dimensions.SafeRewriting},
	}
}

func validateCapabilityRecordOrder(records []CapabilityRecord) error {
	for i := 1; i < len(records); i++ {
		if compareCapabilityRecords(records[i-1], records[i]) >= 0 {
			return fmt.Errorf("capability records are not in canonical order at indexes %d and %d", i-1, i)
		}
	}
	return nil
}

func compareCapabilityRecords(left, right CapabilityRecord) int {
	for _, values := range [][2]string{
		{left.Language, right.Language},
		{left.Profile, right.Profile},
		{left.Kind, right.Kind},
		{left.Name, right.Name},
		{left.Form, right.Form},
		{left.ID, right.ID},
	} {
		if comparison := strings.Compare(values[0], values[1]); comparison != 0 {
			return comparison
		}
	}
	return 0
}

func validateCapabilityEvidenceOrder(cases []CapabilityEvidence) error {
	for i := 1; i < len(cases); i++ {
		if compareCapabilityEvidence(cases[i-1], cases[i]) >= 0 {
			return fmt.Errorf("capability evidence is not in canonical order at indexes %d and %d", i-1, i)
		}
	}
	return nil
}

func compareCapabilityEvidence(left, right CapabilityEvidence) int {
	for _, values := range [][2]string{
		{left.Document.Language, right.Document.Language},
		{left.Document.Profile, right.Document.Profile},
		{left.ID, right.ID},
	} {
		if comparison := strings.Compare(values[0], values[1]); comparison != 0 {
			return comparison
		}
	}
	return 0
}

func summarizeCapabilityRecords(records []CapabilityRecord) CapabilitySummary {
	var summary CapabilitySummary
	for _, record := range records {
		countCapabilityState(&summary.Syntax, record.Dimensions.Syntax.State)
		countCapabilityState(&summary.Semantics, record.Dimensions.Semantics.State)
		countCapabilityState(&summary.Requirements, record.Dimensions.Requirements.State)
		countCapabilityState(&summary.Linting, record.Dimensions.Linting.State)
		countCapabilityState(&summary.SafeRewriting, record.Dimensions.SafeRewriting.State)
	}
	return summary
}

func countCapabilityState(counts *CapabilityStateCounts, state CapabilityState) {
	switch state {
	case CapabilitySupported:
		counts.Supported++
		counts.Covered++
		counts.Applicable++
	case CapabilityPartial:
		counts.Partial++
		counts.Applicable++
	case CapabilityUnsupported:
		counts.Unsupported++
		counts.Applicable++
	case CapabilityNotApplicable:
		counts.NotApplicable++
	case CapabilityUnassessed:
		counts.Unassessed++
		counts.Applicable++
	}
}
