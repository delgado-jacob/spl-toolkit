package analysis

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/delgado-jacob/spl-toolkit/internal/buildinfo"
)

type capabilityContract struct {
	manifest CapabilityManifest
	revision string
}

type capabilityRevisionPayload struct {
	Rewrite                   *RewriteCapabilityManifest           `json:"rewrite,omitempty"`
	SchemaVersion             int                                  `json:"schema_version"`
	Language                  string                               `json:"language"`
	Profile                   string                               `json:"profile"`
	Version                   string                               `json:"version"`
	DocumentationSnapshot     string                               `json:"documentation_snapshot,omitempty"`
	Commands                  []Capability                         `json:"commands"`
	Functions                 []Capability                         `json:"functions"`
	Records                   []CapabilityRecord                   `json:"records"`
	Summary                   CapabilitySummary                    `json:"summary"`
	Evidence                  []capabilityRevisionEvidence         `json:"evidence"`
	SemanticProofRequirements []capabilityRevisionProofRequirement `json:"semantic_proof_requirements,omitempty"`
}

type capabilityRevisionEvidence struct {
	ID             string                                 `json:"id"`
	Classification CapabilityEvidenceClassification       `json:"classification"`
	Document       QueryDocument                          `json:"document"`
	Observations   capabilityRevisionEvidenceObservations `json:"observations"`
	RewriteRequest json.RawMessage                        `json:"rewrite_request,omitempty" swaggertype:"object"`
	Provenance     CapabilityProvenance                   `json:"provenance"`
}

type capabilityRevisionEvidenceObservations struct {
	Syntax        *CapabilitySyntaxObservation               `json:"syntax,omitempty"`
	Semantics     *capabilityRevisionSemanticsObservation    `json:"semantics,omitempty"`
	Requirements  *capabilityRevisionRequirementsObservation `json:"requirements,omitempty"`
	Linting       *CapabilityLintingObservation              `json:"linting,omitempty"`
	SafeRewriting *CapabilityRewriteObservation              `json:"safe_rewriting,omitempty"`
}

type capabilityRevisionRequirementExpectation CapabilityRequirementExpectation

type capabilityRevisionRequirementsObservation struct {
	QueryStatus Status                                     `json:"query_status"`
	Complete    bool                                       `json:"complete"`
	Items       []capabilityRevisionRequirementExpectation `json:"items"`
	GapCodes    []string                                   `json:"gap_codes"`
}

type capabilityRevisionStageExpectation CapabilityStageExpectation
type capabilityRevisionReferenceExpectation CapabilityReferenceExpectation
type capabilityRevisionTransitionExpectation CapabilityTransitionExpectation

type capabilityRevisionSemanticsObservation struct {
	Status          Status                                    `json:"status"`
	Complete        bool                                      `json:"complete"`
	Stages          []capabilityRevisionStageExpectation      `json:"stages"`
	Scopes          []CapabilityScopeExpectation              `json:"scopes,omitempty"`
	References      []capabilityRevisionReferenceExpectation  `json:"references"`
	Dependencies    []CapabilityDependencyExpectation         `json:"dependencies"`
	Lineage         []CapabilityLineageExpectation            `json:"lineage,omitempty"`
	Transitions     []capabilityRevisionTransitionExpectation `json:"transitions"`
	FinalFieldState *CapabilityFieldStateExpectation          `json:"final_field_state,omitempty"`
	Diagnostics     []CapabilityDiagnosticExpectation         `json:"diagnostics"`
}

type capabilityRevisionProofRequirement struct {
	RecordID   string   `json:"record_id"`
	Categories []string `json:"categories"`
}

var capabilityCatalog = mustBuildCapabilityCatalog()

func mustBuildCapabilityCatalog() map[CapabilityOptions]capabilityContract {
	records, evidence, err := loadEmbeddedCapabilityData()
	if err != nil {
		panic(fmt.Sprintf("load embedded capability catalog: %v", err))
	}
	catalog := make(map[CapabilityOptions]capabilityContract, 2)
	for _, language := range []string{"spl", "spl2"} {
		options := CapabilityOptions{Language: language, Profile: "splunkd", Version: "current"}
		manifest := buildCapabilityManifest(options, records, evidence)
		payload := capabilityRevisionPayloadFor(manifest)
		encoded, err := json.Marshal(payload)
		if err != nil {
			panic(fmt.Sprintf("encode %s capability revision: %v", language, err))
		}
		sum := sha256.Sum256(encoded)
		catalog[options] = capabilityContract{
			manifest: manifest,
			revision: "sha256:" + hex.EncodeToString(sum[:]),
		}
	}
	return catalog
}

func buildCapabilityManifest(options CapabilityOptions, allRecords []CapabilityRecord, allEvidence []CapabilityEvidence) CapabilityManifest {
	records := make([]CapabilityRecord, 0, len(allRecords))
	referencedEvidence := make(map[string]struct{})
	for _, record := range allRecords {
		if record.Language != options.Language || record.Profile != options.Profile {
			continue
		}
		records = append(records, cloneCapabilityRecord(record))
		for _, dimension := range capabilityDimensionClaims(record.Dimensions) {
			for _, evidenceID := range dimension.claim.EvidenceIDs {
				referencedEvidence[evidenceID] = struct{}{}
			}
		}
	}
	evidence := make([]CapabilityEvidence, 0, len(referencedEvidence))
	for _, item := range allEvidence {
		if _, referenced := referencedEvidence[item.ID]; referenced {
			evidence = append(evidence, cloneCapabilityEvidence(item))
		}
	}
	manifest := CapabilityManifest{
		Rewrite:        rewriteCapabilities(options.Language),
		SchemaVersion:  1,
		Language:       options.Language,
		Profile:        options.Profile,
		Version:        options.Version,
		ToolkitVersion: buildinfo.Version,
		Commands:       projectLegacyCapabilities(records, "command"),
		Functions:      projectLegacyCapabilities(records, "function"),
		Records:        records,
		Summary:        summarizeCapabilityRecords(records),
		Evidence:       evidence,
	}
	if options.Language == "spl2" {
		manifest.DocumentationSnapshot = spl2DocumentationSnapshot
	}
	return manifest
}

func projectLegacyCapabilities(records []CapabilityRecord, kind string) []Capability {
	projected := []Capability{}
	indexes := make(map[string]int)
	seenLimitations := make(map[string]map[string]struct{})
	for _, record := range records {
		if record.Kind != kind {
			continue
		}
		index, found := indexes[record.Name]
		if !found {
			index = len(projected)
			indexes[record.Name] = index
			seenLimitations[record.Name] = make(map[string]struct{})
			projected = append(projected, Capability{Name: record.Name, Limitations: []string{}})
		}
		projected[index].SyntaxSupported = projected[index].SyntaxSupported || record.GrammarRegistered
		projected[index].SemanticSupported = projected[index].SemanticSupported || record.Dimensions.Semantics.State == CapabilitySupported
		for _, dimension := range capabilityDimensionClaims(record.Dimensions) {
			for _, limitation := range dimension.claim.Limitations {
				if _, seen := seenLimitations[record.Name][limitation]; seen {
					continue
				}
				seenLimitations[record.Name][limitation] = struct{}{}
				projected[index].Limitations = append(projected[index].Limitations, limitation)
			}
		}
	}
	return projected
}

func capabilityRevisionPayloadFor(manifest CapabilityManifest) capabilityRevisionPayload {
	return capabilityRevisionPayloadForRequirements(manifest, milestone11SemanticProofRequirements)
}

func capabilityRevisionPayloadForRequirements(manifest CapabilityManifest, proofRequirements map[string]capabilitySemanticProofRequirement) capabilityRevisionPayload {
	return capabilityRevisionPayload{
		Rewrite:                   manifest.Rewrite,
		SchemaVersion:             manifest.SchemaVersion,
		Language:                  manifest.Language,
		Profile:                   manifest.Profile,
		Version:                   manifest.Version,
		DocumentationSnapshot:     manifest.DocumentationSnapshot,
		Commands:                  manifest.Commands,
		Functions:                 manifest.Functions,
		Records:                   manifest.Records,
		Summary:                   manifest.Summary,
		Evidence:                  capabilityRevisionEvidenceFor(manifest.Evidence),
		SemanticProofRequirements: capabilityRevisionProofRequirementsFor(manifest.Records, proofRequirements),
	}
}

func capabilityRevisionEvidenceFor(evidence []CapabilityEvidence) []capabilityRevisionEvidence {
	out := make([]capabilityRevisionEvidence, len(evidence))
	for i, item := range evidence {
		out[i] = capabilityRevisionEvidence{
			ID:             item.ID,
			Classification: item.Classification,
			Document:       item.Document,
			Observations: capabilityRevisionEvidenceObservations{
				Syntax:        item.Observations.Syntax,
				Semantics:     capabilityRevisionSemanticsFor(item.Observations.Semantics),
				Requirements:  capabilityRevisionRequirementsFor(item.Observations.Requirements),
				Linting:       item.Observations.Linting,
				SafeRewriting: item.Observations.SafeRewriting,
			},
			RewriteRequest: item.RewriteRequest,
			Provenance:     item.Provenance,
		}
	}
	return out
}

func capabilityRevisionRequirementsFor(observation *CapabilityRequirementsObservation) *capabilityRevisionRequirementsObservation {
	if observation == nil {
		return nil
	}
	items := make([]capabilityRevisionRequirementExpectation, len(observation.Items))
	for i, item := range observation.Items {
		items[i] = capabilityRevisionRequirementExpectation(item)
	}
	return &capabilityRevisionRequirementsObservation{
		QueryStatus: observation.QueryStatus,
		Complete:    observation.Complete,
		Items:       items,
		GapCodes:    observation.GapCodes,
	}
}

func capabilityRevisionSemanticsFor(observation *CapabilitySemanticsObservation) *capabilityRevisionSemanticsObservation {
	if observation == nil {
		return nil
	}
	stages := make([]capabilityRevisionStageExpectation, len(observation.Stages))
	for i, stage := range observation.Stages {
		stages[i] = capabilityRevisionStageExpectation(stage)
	}
	references := make([]capabilityRevisionReferenceExpectation, len(observation.References))
	for i, reference := range observation.References {
		references[i] = capabilityRevisionReferenceExpectation(reference)
	}
	transitions := make([]capabilityRevisionTransitionExpectation, len(observation.Transitions))
	for i, transition := range observation.Transitions {
		transitions[i] = capabilityRevisionTransitionExpectation(transition)
	}
	return &capabilityRevisionSemanticsObservation{
		Status:          observation.Status,
		Complete:        observation.Complete,
		Stages:          stages,
		Scopes:          observation.Scopes,
		References:      references,
		Dependencies:    observation.Dependencies,
		Lineage:         observation.Lineage,
		Transitions:     transitions,
		FinalFieldState: observation.FinalFieldState,
		Diagnostics:     observation.Diagnostics,
	}
}

func capabilityRevisionProofRequirementsFor(records []CapabilityRecord, proofRequirements map[string]capabilitySemanticProofRequirement) []capabilityRevisionProofRequirement {
	out := make([]capabilityRevisionProofRequirement, 0, len(proofRequirements))
	for _, record := range records {
		requirement, ok := proofRequirements[record.ID]
		if !ok {
			continue
		}
		out = append(out, capabilityRevisionProofRequirement{
			RecordID:   record.ID,
			Categories: capabilityRevisionProofCategories(requirement),
		})
	}
	return out
}

func capabilityRevisionProofCategories(requirement capabilitySemanticProofRequirement) []string {
	categories := make([]string, 0, 7)
	for _, category := range []struct {
		flag capabilitySemanticProofRequirement
		name string
	}{
		{capabilityProofScope, "scope"},
		{capabilityProofLineage, "lineage"},
		{capabilityProofOrigins, "origins"},
		{capabilityProofTransitions, "transitions"},
		{capabilityProofFinalState, "final_state"},
		{capabilityProofMerge, "merge"},
		{capabilityProofIdentity, "identity"},
	} {
		if requirement&category.flag != 0 {
			categories = append(categories, category.name)
		}
	}
	return categories
}

// Capabilities returns a detached copy of the default capability contract.
func Capabilities() CapabilityManifest {
	return cloneCapabilityManifest(capabilityCatalog[CapabilityOptions{Language: "spl", Profile: "splunkd", Version: "current"}].manifest)
}

// CapabilitiesFor validates selectors and returns a fresh available manifest.
func CapabilitiesFor(options CapabilityOptions) (CapabilityManifest, error) {
	_, contract, err := lookupCapabilityContract(options)
	if err != nil {
		return CapabilityManifest{}, err
	}
	return cloneCapabilityManifest(contract.manifest), nil
}

func capabilityRevisionFor(options CapabilityOptions) (string, error) {
	_, contract, err := lookupCapabilityContract(options)
	if err != nil {
		return "", err
	}
	return contract.revision, nil
}

func lookupCapabilityContract(options CapabilityOptions) (CapabilityOptions, capabilityContract, error) {
	normalized, err := normalizeSelectors(options)
	if err != nil {
		return CapabilityOptions{}, capabilityContract{}, err
	}
	contract, found := capabilityCatalog[normalized]
	if !found {
		return CapabilityOptions{}, capabilityContract{}, fmt.Errorf("capability contract is unavailable for language %q, profile %q, version %q", normalized.Language, normalized.Profile, normalized.Version)
	}
	return normalized, contract, nil
}

func cloneCapabilityManifest(manifest CapabilityManifest) CapabilityManifest {
	cloned := manifest
	if manifest.Rewrite != nil {
		rewrite := *manifest.Rewrite
		rewrite.Forms = cloneCapabilitySlice(manifest.Rewrite.Forms)
		for i := range rewrite.Forms {
			rewrite.Forms[i].IdentityForms = cloneCapabilitySlice(rewrite.Forms[i].IdentityForms)
			rewrite.Forms[i].Limitations = cloneCapabilitySlice(rewrite.Forms[i].Limitations)
		}
		cloned.Rewrite = &rewrite
	}
	cloned.Commands = cloneCapabilitySlice(manifest.Commands)
	for i := range cloned.Commands {
		cloned.Commands[i].Limitations = cloneCapabilitySlice(cloned.Commands[i].Limitations)
	}
	cloned.Functions = cloneCapabilitySlice(manifest.Functions)
	for i := range cloned.Functions {
		cloned.Functions[i].Limitations = cloneCapabilitySlice(cloned.Functions[i].Limitations)
	}
	cloned.Records = cloneCapabilityRecords(manifest.Records)
	cloned.Evidence = cloneCapabilityEvidenceCases(manifest.Evidence)
	return cloned
}

// These are canonical render roles, not promises of runtime equivalence. Binding,
// linked effects and whole-candidate correspondence still govern each edit.
func rewriteCapabilities(language string) *RewriteCapabilityManifest {
	m := &RewriteCapabilityManifest{SchemaVersion: 1, Forms: []RewriteCapabilityForm{}}
	add := func(kind, role string, supported bool, limitations ...string) {
		m.Forms = append(m.Forms, RewriteCapabilityForm{Kind: kind, Role: role, IdentityForms: []string{"atom"}, Supported: supported, Limitations: append([]string{}, limitations...)})
	}
	for _, role := range []string{"expression_atom", "search_field", "selector_atom", "rename_input", "null_test", "lookup_local", "lookup_dual"} {
		add("field", role, true, "Requires an exact source binding; explicit aliases, wildcard selectors and indeterminate operands remain fixed.")
	}
	add("field", "implicit_output", language == "spl", "Only canonical implicit labels and all required consumers may change together; SPL2 changed labels remain unproved.")
	m.Forms = append(m.Forms, RewriteCapabilityForm{Kind: "field", Role: "navigation", IdentityForms: []string{"path"}, Supported: false, Limitations: []string{"Navigated and alias-qualified identities have no canonical source binding proof."}})
	for _, kind := range []string{"index", "source", "sourcetype"} {
		add(kind, "search_value", true, "Typed exact static search values only; patterns, interpolation and uncertain decoding remain held.")
	}
	for _, kind := range []string{"lookup", "dataset"} {
		add(kind, "catalog_atom", true, "Named external dependency, independent of columns and source aliases.")
	}
	if language == "spl2" {
		add("index", "metric_value", true, "Only intact command-owned bare-index equality with a proven static value.")
		add("data_model", "quoted_catalog_atom", true, "Existing atomic tstats datamodel_name token; affected effects remain subject to proof.")
	} else {
		add("data_model", "catalog_atom", true, "Exact model token.")
		add("data_model", "catalog_component", true, "Original qualified owner, prefix and quote boundaries are retained.")
		add("dataset", "catalog_component", true, "Dataset component keeps its model prefix; linked model normalization is explicit.")
		add("dataset", "qualified_dataset", true, "Compatible overlapping proposals co-render; conflicts and backslashes remain refused.")
	}
	return m
}

// CapabilityRevisionFor returns the same normalized contract revision included
// in current-producer requirement reports.
func CapabilityRevisionFor(options CapabilityOptions) (string, error) {
	return capabilityRevisionFor(options)
}
