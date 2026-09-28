package analysis

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"slices"
	"sort"
	"strings"
)

//go:embed capabilitydata/ledger.json
var embeddedCapabilityLedger []byte

//go:embed capabilitydata/corpus.json
var embeddedCapabilityCorpus []byte

func loadEmbeddedCapabilityData() ([]CapabilityRecord, []CapabilityEvidence, error) {
	return decodeCapabilityAssets(embeddedCapabilityLedger, embeddedCapabilityCorpus)
}

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

type capabilitySemanticProofRequirement uint8

const (
	capabilityProofScope capabilitySemanticProofRequirement = 1 << iota
	capabilityProofLineage
	capabilityProofOrigins
	capabilityProofTransitions
	capabilityProofFinalState
	capabilityProofMerge
	capabilityProofIdentity

	capabilityProofScopedLineage = capabilityProofScope | capabilityProofLineage | capabilityProofFinalState
	capabilityProofField         = capabilityProofScopedLineage | capabilityProofOrigins
	capabilityProofTransition    = capabilityProofScopedLineage | capabilityProofTransitions
	capabilityProofFieldTransfer = capabilityProofField | capabilityProofTransitions
	capabilityProofFieldMerge    = capabilityProofFieldTransfer | capabilityProofMerge
)

var milestone11SemanticProofRequirements = map[string]capabilitySemanticProofRequirement{
	"spl2.command.bin.span-field":                          capabilityProofFieldTransfer,
	"spl2.command.eval.exact-assignment":                   capabilityProofFieldTransfer,
	"spl2.command.fields.exact-field-list":                 capabilityProofTransition,
	"spl2.command.from.dataset":                            capabilityProofScopedLineage,
	"spl2.command.if.subpipe":                              capabilityProofFieldMerge,
	"spl2.command.join.qualified-subsearch":                capabilityProofFieldMerge,
	"spl2.command.mvexpand.limited-field":                  capabilityProofField,
	"spl2.command.select.projection":                       capabilityProofFieldTransfer,
	"spl2.command.stats.aggregate-call":                    capabilityProofFieldTransfer,
	"spl2.command.union.dataset":                           capabilityProofFieldMerge,
	"spl2.command.where.predicate":                         capabilityProofField,
	"spl2.dataset.dataset.dynamic-descriptor":              capabilityProofScopedLineage,
	"spl2.dataset.dataset.parameter":                       capabilityProofField,
	"spl2.dataset.dataset.static-descriptor":               capabilityProofScopedLineage,
	"spl2.expression.field.identity-collision":             capabilityProofField | capabilityProofIdentity,
	"spl2.expression.field.quoted-dotted-atom":             capabilityProofField,
	"spl2.expression.field.structural-path":                capabilityProofField,
	"spl2.expression.function-call.invalid-selected-arity": capabilityProofFieldTransfer,
	"spl2.function.abs.one-positional":                     capabilityProofFieldTransfer,
	"spl2.function.any.lambda-positional":                  capabilityProofFieldTransfer,
	"spl2.function.avg.one-positional":                     capabilityProofFieldTransfer,
	"spl2.function.cidrmatch.two-positional":               capabilityProofFieldTransfer,
	"spl2.function.coalesce.two-positional":                capabilityProofFieldTransfer,
	"spl2.function.count.zero-positional":                  capabilityProofFieldTransfer,
	"spl2.function.dc.one-positional":                      capabilityProofFieldTransfer,
	"spl2.function.distinct_count.one-positional":          capabilityProofFieldTransfer,
	"spl2.function.json.one-positional":                    capabilityProofFieldTransfer,
	"spl2.function.json_array_to_mv.one-positional":        capabilityProofFieldTransfer,
	"spl2.function.like.two-positional":                    capabilityProofFieldTransfer,
	"spl2.function.lower.one-positional":                   capabilityProofFieldTransfer,
	"spl2.function.match.two-positional":                   capabilityProofFieldTransfer,
	"spl2.function.max.one-positional":                     capabilityProofFieldTransfer,
	"spl2.function.min.one-positional":                     capabilityProofFieldTransfer,
	"spl2.function.mvindex.two-positional":                 capabilityProofFieldTransfer,
	"spl2.function.round.one-positional":                   capabilityProofFieldTransfer,
	"spl2.function.rtrim.one-positional":                   capabilityProofFieldTransfer,
	"spl2.function.span.grouping-positional":               capabilityProofFieldTransfer,
	"spl2.function.sqrt.one-positional":                    capabilityProofFieldTransfer,
	"spl2.function.stdev.one-positional":                   capabilityProofFieldTransfer,
	"spl2.function.strftime.two-positional":                capabilityProofFieldTransfer,
	"spl2.function.sum.one-positional":                     capabilityProofFieldTransfer,
	"spl2.function.tonumber.one-positional":                capabilityProofFieldTransfer,
	"spl2.function.values.one-positional":                  capabilityProofFieldTransfer,
	"spl2.module.module.declaration-cycle":                 capabilityProofScopedLineage,
	"spl2.module.module.local-scalar-function":             capabilityProofFieldTransfer,
	"spl2.module.module.local-view":                        capabilityProofFieldTransfer,
	"spl2.module.module.unresolved-import":                 capabilityProofScopedLineage,
	"spl2.pipeline.branch.guarded-arms":                    capabilityProofFieldMerge,
	"spl2.pipeline.branch.unguarded-arms":                  capabilityProofFieldMerge,
	"spl2.pipeline.join.output-collision":                  capabilityProofFieldMerge,
	"spl2.pipeline.join.qualified-left-subsearch":          capabilityProofFieldMerge,
	"spl2.pipeline.join.qualified-outer-subsearch":         capabilityProofFieldMerge,
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
	for i := range corpus.Cases {
		if len(corpus.Cases[i].RewriteRequest) == 0 {
			continue
		}
		var compact bytes.Buffer
		if err := json.Compact(&compact, corpus.Cases[i].RewriteRequest); err != nil {
			return nil, nil, fmt.Errorf("compact capability evidence %q rewrite request: %w", corpus.Cases[i].ID, err)
		}
		corpus.Cases[i].RewriteRequest = json.RawMessage(compact.Bytes())
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
	if requirement, ok := milestone11SemanticProofRequirements[record.ID]; ok {
		for _, evidenceID := range record.Dimensions.Semantics.EvidenceIDs {
			if err := validateCapabilitySemanticProof(evidence[evidenceID].Observations.Semantics, requirement); err != nil {
				return fmt.Errorf("semantics evidence %q: %w", evidenceID, err)
			}
		}
	}
	if record.ID == "spl2.expression.field.identity-collision" && record.Dimensions.Requirements.State == CapabilitySupported {
		for _, evidenceID := range record.Dimensions.Requirements.EvidenceIDs {
			if err := validateCapabilityIdentityCollisionRequirementProof(evidence[evidenceID].Observations.Requirements); err != nil {
				return fmt.Errorf("requirements evidence %q: %w", evidenceID, err)
			}
		}
	}
	return nil
}

func validateCapabilityIdentityCollisionRequirementProof(observation *CapabilityRequirementsObservation) error {
	if observation == nil || !observation.Complete || observation.QueryStatus != Valid {
		return fmt.Errorf("complete valid identity-collision requirement proof is required")
	}
	fields := make([]CapabilityRequirementExpectation, 0, 2)
	for _, item := range observation.Items {
		if item.Kind == "field" {
			fields = append(fields, item)
		}
	}
	if len(fields) != 2 {
		return fmt.Errorf("two distinct exact same-display field requirements are required")
	}
	left, right := fields[0], fields[1]
	if left.Identity != right.Identity || left.FieldIdentity == nil || right.FieldIdentity == nil ||
		left.Role != "read" || right.Role != "read" ||
		left.Necessity != "required" || right.Necessity != "required" ||
		left.Resolution != "exact" || right.Resolution != "exact" ||
		capabilityFieldIdentitiesEqual(left.FieldIdentity, right.FieldIdentity) {
		return fmt.Errorf("two distinct exact same-display field requirements are required")
	}
	if left.FieldIdentity.Kind == right.FieldIdentity.Kind {
		return fmt.Errorf("atomic and path field requirements are both required")
	}
	return nil
}

func validateCapabilitySemanticProof(observation *CapabilitySemanticsObservation, requirement capabilitySemanticProofRequirement) error {
	if observation == nil {
		return fmt.Errorf("structured proof is required")
	}
	if requirement&capabilityProofScope != 0 && len(observation.Scopes) == 0 {
		return fmt.Errorf("scope proof is required")
	}
	if requirement&capabilityProofLineage != 0 && len(observation.Lineage) == 0 {
		return fmt.Errorf("lineage proof is required")
	}
	if requirement&capabilityProofOrigins != 0 && !capabilitySemanticsHaveOrigins(observation) {
		return fmt.Errorf("field-origin proof is required")
	}
	if requirement&capabilityProofTransitions != 0 && !capabilitySemanticsHaveExpandedTransitions(observation) {
		return fmt.Errorf("expanded transition proof is required")
	}
	if requirement&capabilityProofFinalState != 0 && observation.FinalFieldState == nil {
		return fmt.Errorf("final field-state proof is required")
	}
	if requirement&capabilityProofMerge != 0 && !capabilitySemanticsHaveMergeFacts(observation) {
		return fmt.Errorf("merge proof is required")
	}
	if requirement&capabilityProofIdentity != 0 && !capabilitySemanticsHaveIdentityCollisionFacts(observation) {
		return fmt.Errorf("distinct field-identity proof is required")
	}
	return nil
}

func capabilitySemanticsHaveOrigins(observation *CapabilitySemanticsObservation) bool {
	found := false
	visit := func(state *CapabilityFieldStateExpectation) {
		if state == nil {
			return
		}
		for _, field := range state.Fields {
			found = found || len(field.OriginReferenceIDs) != 0
		}
	}
	for _, lineage := range observation.Lineage {
		visit(lineage.Before)
		visit(lineage.After)
	}
	visit(observation.FinalFieldState)
	return found
}

func capabilitySemanticsHaveExpandedTransitions(observation *CapabilitySemanticsObservation) bool {
	for _, transition := range observation.Transitions {
		if len(transition.InputReferenceIDs) != 0 || transition.OutputReferenceID != "" || transition.Conditional {
			return true
		}
	}
	return false
}

func capabilitySemanticsHaveMergeFacts(observation *CapabilitySemanticsObservation) bool {
	found := false
	for _, scope := range observation.Scopes {
		found = found || scope.Kind == "search" && scope.ParentID != ""
	}
	visit := func(state *CapabilityFieldStateExpectation) {
		if state == nil {
			return
		}
		found = found || state.Uncertain
		for _, field := range state.Fields {
			found = found || field.Conditional || len(field.OriginReferenceIDs) > 1
		}
	}
	for _, lineage := range observation.Lineage {
		visit(lineage.Before)
		visit(lineage.After)
	}
	visit(observation.FinalFieldState)
	for _, transition := range observation.Transitions {
		found = found || transition.Conditional
	}
	return found
}

func capabilitySemanticsHaveIdentityCollisionFacts(observation *CapabilitySemanticsObservation) bool {
	if observation.FinalFieldState == nil {
		return false
	}
	fields := observation.FinalFieldState.Fields
	for i := 1; i < len(fields); i++ {
		left, right := fields[i-1], fields[i]
		if left.Name == right.Name && left.FieldIdentity != nil && right.FieldIdentity != nil && !capabilityFieldIdentitiesEqual(left.FieldIdentity, right.FieldIdentity) && len(left.OriginReferenceIDs) != 0 && len(right.OriginReferenceIDs) != 0 {
			return true
		}
	}
	return false
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
		if claim.State == CapabilitySupported && dimension == "linting" && evidence.Classification == CapabilityEvidenceNegative {
			if err := validateExactLintingDiagnostics(evidence.Document, evidence.Observations.Linting.Diagnostics); err != nil {
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
		return validateCapabilityDiagnostics(evidence.Document, observation.Diagnostics, false)
	case "semantics":
		observation := evidence.Observations.Semantics
		if err := validateCapabilityStatus(observation.Status); err != nil {
			return err
		}
		if positive && !observation.Complete {
			return fmt.Errorf("positive semantics evidence must be complete")
		}
		if len(observation.Stages)+len(observation.Scopes)+len(observation.References)+len(observation.Dependencies)+len(observation.Lineage)+len(observation.Transitions)+len(observation.Diagnostics) == 0 && observation.FinalFieldState == nil {
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
			if err := validateCapabilityLocation(evidence.Document, reference.Location); err != nil {
				return fmt.Errorf("reference %d: %w", i, err)
			}
			if err := validateCapabilityFieldIdentity(reference.FieldIdentity, reference.NormalizedName); err != nil {
				return fmt.Errorf("reference %d: %w", i, err)
			}
			if reference.FieldIdentity != nil && (reference.Kind != "field" || reference.Resolution != "exact") {
				return fmt.Errorf("reference %d has identity outside an exact field", i)
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
			if err := validateCapabilityFieldIdentity(transition.OutputIdentity, transition.Output); err != nil {
				return fmt.Errorf("transition %d: %w", i, err)
			}
		}
		if err := validateCapabilityStructuredSemantics(evidence.Document, observation); err != nil {
			return err
		}
		return validateCapabilityDiagnostics(evidence.Document, observation.Diagnostics, false)
	case "requirements":
		observation := evidence.Observations.Requirements
		if err := validateCapabilityStatus(observation.QueryStatus); err != nil {
			return err
		}
		if positive && !observation.Complete {
			return fmt.Errorf("positive requirements evidence must be complete")
		}
		if observation.Items == nil || observation.GapCodes == nil {
			return fmt.Errorf("requirements items and gap_codes must be exact arrays")
		}
		if !observation.Complete && len(observation.Items)+len(observation.GapCodes) == 0 {
			return fmt.Errorf("at least one typed requirement fact is required")
		}
		if observation.Complete && len(observation.GapCodes) != 0 {
			return fmt.Errorf("complete requirements observation must not contain gap codes")
		}
		for i, item := range observation.Items {
			if err := requireCapabilityFields("requirement item", item.Kind, item.Identity, item.Role, item.Necessity, item.Resolution); err != nil {
				return fmt.Errorf("item %d: %w", i, err)
			}
			if err := validateCapabilityFieldIdentity(item.FieldIdentity, item.Identity); err != nil {
				return fmt.Errorf("item %d: %w", i, err)
			}
			if item.FieldIdentity != nil && (item.Kind != "field" || item.Resolution != "exact") {
				return fmt.Errorf("item %d has identity outside an exact field", i)
			}
		}
		return validateNonemptyStrings("gap code", observation.GapCodes)
	case "linting":
		return validateCapabilityDiagnostics(evidence.Document, evidence.Observations.Linting.Diagnostics, false)
	case "safe_rewriting":
		observation := evidence.Observations.SafeRewriting
		if err := validateCapabilityStatus(observation.Status); err != nil {
			return err
		}
		if positive && !observation.RewriteComplete {
			return fmt.Errorf("positive rewrite evidence must be complete")
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

func validateCapabilityStructuredSemantics(document QueryDocument, observation *CapabilitySemanticsObservation) error {
	structured := len(observation.Scopes) != 0 || len(observation.Lineage) != 0 || observation.FinalFieldState != nil
	for _, stage := range observation.Stages {
		structured = structured || stage.ID != "" || stage.ScopeID != ""
	}
	for _, reference := range observation.References {
		structured = structured || reference.ID != ""
	}
	for _, transition := range observation.Transitions {
		structured = structured || transition.InputReferenceIDs != nil || transition.OutputReferenceID != "" || transition.Conditional
	}
	if !structured {
		return nil
	}

	stageIDs := make(map[string]struct{}, len(observation.Stages))
	for i, stage := range observation.Stages {
		if stage.ID == "" && stage.ScopeID == "" {
			continue
		}
		if err := requireCapabilityFields("structured stage", stage.ID, stage.ScopeID); err != nil {
			return fmt.Errorf("stage %d: %w", i, err)
		}
		if _, duplicate := stageIDs[stage.ID]; duplicate {
			return fmt.Errorf("stage %d has duplicate ID %q", i, stage.ID)
		}
		stageIDs[stage.ID] = struct{}{}
	}

	referenceRoles := make(map[string]string, len(observation.References))
	referenceIdentities := make(map[string]*FieldIdentity, len(observation.References))
	referenceOrder := make(map[string]int, len(observation.References))
	for i, reference := range observation.References {
		if strings.TrimSpace(reference.ID) == "" {
			continue
		}
		if _, duplicate := referenceRoles[reference.ID]; duplicate {
			return fmt.Errorf("reference %d has duplicate ID %q", i, reference.ID)
		}
		referenceRoles[reference.ID] = reference.Role
		referenceIdentities[reference.ID] = reference.FieldIdentity
		referenceOrder[reference.ID] = i
	}
	scopeIDs := make(map[string]struct{}, len(observation.Scopes))
	for i, scope := range observation.Scopes {
		if err := requireCapabilityFields("scope", scope.ID, scope.Kind); err != nil {
			return fmt.Errorf("scope %d: %w", i, err)
		}
		if i > 0 && !capabilityOrdinalIDLess(observation.Scopes[i-1].ID, scope.ID) {
			return fmt.Errorf("scopes are not in canonical ID order at indexes %d and %d", i-1, i)
		}
		if scope.StageID != "" {
			if _, ok := stageIDs[scope.StageID]; !ok {
				return fmt.Errorf("scope %d names unknown stage ID %q", i, scope.StageID)
			}
		}
		if scope.ParentID != "" {
			if _, ok := scopeIDs[scope.ParentID]; !ok {
				return fmt.Errorf("scope %d names unknown or later parent scope ID %q", i, scope.ParentID)
			}
		}
		if err := validateCapabilityLocation(document, scope.Location); err != nil {
			return fmt.Errorf("scope %d: %w", i, err)
		}
		scopeIDs[scope.ID] = struct{}{}
	}
	for i, stage := range observation.Stages {
		if stage.ID == "" && stage.ScopeID == "" {
			continue
		}
		if _, ok := scopeIDs[stage.ScopeID]; !ok {
			return fmt.Errorf("stage %d names unknown scope ID %q", i, stage.ScopeID)
		}
	}

	for i, lineage := range observation.Lineage {
		if err := requireCapabilityFields("lineage", lineage.StageID, lineage.ScopeID); err != nil {
			return fmt.Errorf("lineage %d: %w", i, err)
		}
		if _, ok := stageIDs[lineage.StageID]; !ok {
			return fmt.Errorf("lineage %d names unknown stage ID %q", i, lineage.StageID)
		}
		if _, ok := scopeIDs[lineage.ScopeID]; !ok {
			return fmt.Errorf("lineage %d names unknown scope ID %q", i, lineage.ScopeID)
		}
		if lineage.Before == nil && lineage.After == nil {
			return fmt.Errorf("lineage %d requires before or after field state", i)
		}
		if err := validateCapabilityFieldState("lineage before", lineage.Before, referenceOrder, observation.Transitions); err != nil {
			return fmt.Errorf("lineage %d: %w", i, err)
		}
		if err := validateCapabilityFieldState("lineage after", lineage.After, referenceOrder, observation.Transitions); err != nil {
			return fmt.Errorf("lineage %d: %w", i, err)
		}
	}

	for i, transition := range observation.Transitions {
		if transition.InputReferenceIDs == nil && transition.OutputReferenceID == "" && !transition.Conditional {
			continue
		}
		if !orderedCapabilityReferenceIDs(transition.InputReferenceIDs, referenceOrder) {
			return fmt.Errorf("transition %d input reference IDs are not in canonical order", i)
		}
		for _, id := range transition.InputReferenceIDs {
			role, ok := referenceRoles[id]
			if !ok {
				return fmt.Errorf("transition %d names unknown input reference ID %q", i, id)
			}
			if role == "output" {
				return fmt.Errorf("transition %d uses output reference %q as an input", i, id)
			}
		}
		if transition.OutputReferenceID != "" {
			role, ok := referenceRoles[transition.OutputReferenceID]
			if !ok {
				return fmt.Errorf("transition %d names unknown output reference ID %q", i, transition.OutputReferenceID)
			}
			if role != "output" && role != "create" && role != "remove" {
				return fmt.Errorf("transition %d output reference %q has role %q", i, transition.OutputReferenceID, role)
			}
			if referenceIdentity := referenceIdentities[transition.OutputReferenceID]; referenceIdentity != nil && transition.OutputIdentity != nil && !capabilityFieldIdentitiesEqual(referenceIdentity, transition.OutputIdentity) {
				return fmt.Errorf("transition %d output identity differs from reference %q", i, transition.OutputReferenceID)
			}
		}
	}

	if err := validateCapabilityFieldState("final field state", observation.FinalFieldState, referenceOrder, observation.Transitions); err != nil {
		return err
	}
	return nil
}

func validateCapabilityFieldState(label string, state *CapabilityFieldStateExpectation, referenceOrder map[string]int, transitions []CapabilityTransitionExpectation) error {
	if state == nil {
		return nil
	}
	if state.Fields == nil || state.Removed == nil {
		return fmt.Errorf("%s fields and removed must be exact arrays", label)
	}
	var previousField CapabilityFieldExpectation
	for i, field := range state.Fields {
		if strings.TrimSpace(field.Name) == "" {
			return fmt.Errorf("%s field %d requires a name", label, i)
		}
		if err := validateCapabilityFieldIdentity(field.FieldIdentity, field.Name); err != nil {
			return fmt.Errorf("%s field %d: %w", label, i, err)
		}
		if i > 0 && !capabilityFieldExpectationLess(previousField.Name, previousField.FieldIdentity, field.Name, field.FieldIdentity) {
			return fmt.Errorf("%s fields are not in canonical identity order at indexes %d and %d", label, i-1, i)
		}
		if field.OriginReferenceIDs == nil || !orderedCapabilityOriginReferenceIDs(field.Name, field.FieldIdentity, field.OriginReferenceIDs, referenceOrder, transitions) {
			return fmt.Errorf("%s field %d origin reference IDs must be an exact canonical array", label, i)
		}
		for _, id := range field.OriginReferenceIDs {
			if _, ok := referenceOrder[id]; !ok {
				return fmt.Errorf("%s field %d names unknown origin reference ID %q", label, i, id)
			}
		}
		previousField = field
	}
	for i, removal := range state.Removed {
		if strings.TrimSpace(removal.Name) == "" {
			return fmt.Errorf("%s removed field must not be empty", label)
		}
		if err := validateCapabilityFieldIdentity(removal.FieldIdentity, removal.Name); err != nil {
			return fmt.Errorf("%s removal %d: %w", label, i, err)
		}
		if i > 0 && !capabilityFieldExpectationLess(state.Removed[i-1].Name, state.Removed[i-1].FieldIdentity, removal.Name, removal.FieldIdentity) {
			return fmt.Errorf("%s removed fields are not in canonical identity order", label)
		}
		for _, field := range state.Fields {
			if field.Name == removal.Name && (field.FieldIdentity == nil || removal.FieldIdentity == nil || capabilityFieldIdentitiesEqual(field.FieldIdentity, removal.FieldIdentity)) {
				return fmt.Errorf("%s field %q cannot be both present and removed", label, removal.Name)
			}
		}
	}
	return nil
}

func validateCapabilityFieldIdentity(identity *FieldIdentity, name string) error {
	if identity == nil {
		return nil
	}
	if len(identity.Segments) == 0 || (identity.Kind != "atomic" && identity.Kind != "path") {
		return fmt.Errorf("field identity requires atomic or path kind and nonempty segments")
	}
	for _, segment := range identity.Segments {
		if segment == "" {
			return fmt.Errorf("field identity has an empty segment")
		}
	}
	if identity.Kind == "atomic" && (len(identity.Segments) != 1 || identity.Qualifier != "") || strings.Join(identity.Segments, ".") != name {
		return fmt.Errorf("field identity does not match name %q", name)
	}
	return nil
}

func capabilityFieldIdentitiesEqual(left, right *FieldIdentity) bool {
	return left.Kind == right.Kind && left.Qualifier == right.Qualifier && slices.Equal(left.Segments, right.Segments)
}

func capabilityFieldExpectationLess(leftName string, leftIdentity *FieldIdentity, rightName string, rightIdentity *FieldIdentity) bool {
	if leftName != rightName {
		return leftName < rightName
	}
	if leftIdentity == nil || rightIdentity == nil {
		return false
	}
	return fieldIdentityLess(*leftIdentity, *rightIdentity)
}

func orderedCapabilityOriginReferenceIDs(fieldName string, fieldIdentity *FieldIdentity, values []string, referenceOrder map[string]int, transitions []CapabilityTransitionExpectation) bool {
	if orderedCapabilityReferenceIDs(values, referenceOrder) {
		return true
	}
	producers := make([]CapabilityTransitionExpectation, 0, 2)
	for _, transition := range transitions {
		if transition.Output == fieldName && (fieldIdentity == nil || transition.OutputIdentity != nil && capabilityFieldIdentitiesEqual(fieldIdentity, transition.OutputIdentity)) && transition.OutputReferenceID != "" && slices.Contains(values, transition.OutputReferenceID) {
			producers = append(producers, transition)
		}
	}
	if len(producers) == 1 {
		producer := producers[0]
		prefixLength := len(producer.InputReferenceIDs) + 1
		if len(values) < prefixLength || values[0] != producer.OutputReferenceID || !slices.Equal(values[1:prefixLength], producer.InputReferenceIDs) {
			return false
		}
		return (producer.Operation == "aggregate" || len(values) > prefixLength) && orderedCapabilityReferenceIDs(values[prefixLength:], referenceOrder)
	}
	if len(producers) < 2 {
		return false
	}
	producerIDs := make(map[string]struct{})
	for _, producer := range producers {
		producerIDs[producer.OutputReferenceID] = struct{}{}
		for _, input := range producer.InputReferenceIDs {
			producerIDs[input] = struct{}{}
		}
	}
	expected := make([]string, 0, len(values))
	for _, value := range values {
		if _, produced := producerIDs[value]; !produced {
			expected = append(expected, value)
		}
	}
	if !orderedCapabilityReferenceIDs(expected, referenceOrder) {
		return false
	}
	for _, producer := range producers {
		if !orderedCapabilityReferenceIDs(producer.InputReferenceIDs, referenceOrder) {
			return false
		}
		expected = append(expected, producer.InputReferenceIDs...)
		expected = append(expected, producer.OutputReferenceID)
	}
	return slices.Equal(values, expected)
}

func orderedCapabilityReferenceIDs(values []string, order map[string]int) bool {
	seen := make(map[string]struct{}, len(values))
	previous := -1
	for _, value := range values {
		position, ok := order[value]
		if !ok || position <= previous {
			return false
		}
		if _, duplicate := seen[value]; duplicate {
			return false
		}
		seen[value] = struct{}{}
		previous = position
	}
	return true
}

func capabilityOrdinalIDLess(left, right string) bool {
	leftPrefix, leftOrdinal := referenceIDOrder(left)
	rightPrefix, rightOrdinal := referenceIDOrder(right)
	if leftPrefix != rightPrefix {
		return leftPrefix < rightPrefix
	}
	if leftOrdinal != rightOrdinal {
		return leftOrdinal < rightOrdinal
	}
	return left < right
}

func strictlyIncreasingStrings(values []string) bool {
	for i, value := range values {
		if strings.TrimSpace(value) == "" || (i > 0 && strings.Compare(values[i-1], value) >= 0) {
			return false
		}
	}
	return true
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

func validateExactLintingDiagnostics(document QueryDocument, diagnostics []CapabilityDiagnosticExpectation) error {
	return validateCapabilityDiagnostics(document, diagnostics, true)
}

func validateCapabilityDiagnostics(document QueryDocument, diagnostics []CapabilityDiagnosticExpectation, required bool) error {
	if required && len(diagnostics) == 0 {
		return fmt.Errorf("at least one diagnostic is required")
	}
	for i, diagnostic := range diagnostics {
		if strings.TrimSpace(diagnostic.Code) == "" || strings.TrimSpace(diagnostic.Category) == "" {
			return fmt.Errorf("diagnostic %d requires code, category, and severity", i)
		}
		switch diagnostic.Severity {
		case "error", "warning", "info":
		default:
			return fmt.Errorf("diagnostic %d has unsupported severity %q", i, diagnostic.Severity)
		}
		if err := validateCapabilityLocation(document, diagnostic.Location); err != nil {
			return fmt.Errorf("diagnostic %d: %w", i, err)
		}
	}
	return nil
}

func validateCapabilityLocation(document QueryDocument, location Location) error {
	index := newSourceIndex(document.Text)
	startIndex, endIndex := -1, -1
	for i, position := range index.positions {
		if position == location.Start {
			startIndex = i
		}
		if position == location.End {
			endIndex = i
		}
	}
	if startIndex < 0 || endIndex < 0 {
		return fmt.Errorf("source location must use canonical document coordinates")
	}
	if endIndex < startIndex {
		return fmt.Errorf("source location end precedes start")
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
