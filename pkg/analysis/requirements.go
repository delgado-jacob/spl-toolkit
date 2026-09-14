package analysis

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

type RequirementQueryIdentity struct {
	SourceID    string `json:"source_id"`
	Language    string `json:"language"`
	Profile     string `json:"profile"`
	Version     string `json:"version"`
	QueryDigest string `json:"query_digest"`
}

type RequirementCoverage struct {
	Complete bool     `json:"complete"`
	Reasons  []string `json:"reasons"`
}

type RequirementOccurrence struct {
	ReferenceID  string   `json:"reference_id"`
	OriginalName string   `json:"original_name"`
	Binding      string   `json:"binding"`
	StageID      string   `json:"stage_id"`
	ScopeID      string   `json:"scope_id"`
	Location     Location `json:"location"`
}

type RequirementItem struct {
	ID          string                  `json:"id"`
	Kind        string                  `json:"kind"`
	Identity    string                  `json:"identity"`
	Role        string                  `json:"role"`
	Necessity   string                  `json:"necessity"`
	Origin      string                  `json:"origin"`
	Resolution  string                  `json:"resolution"`
	Occurrences []RequirementOccurrence `json:"occurrences"`
}

type RequirementGap struct {
	Code            string   `json:"code"`
	Message         string   `json:"message"`
	ReferenceIDs    []string `json:"reference_ids"`
	DiagnosticCodes []string `json:"diagnostic_codes"`
}

type RequirementSet struct {
	SchemaVersion      int                      `json:"schema_version"`
	Query              RequirementQueryIdentity `json:"query"`
	CapabilityRevision string                   `json:"capability_revision"`
	QueryStatus        Status                   `json:"query_status"`
	Coverage           RequirementCoverage      `json:"coverage"`
	Items              []RequirementItem        `json:"items"`
	Gaps               []RequirementGap         `json:"gaps"`
	Diagnostics        []Diagnostic             `json:"diagnostics"`
}

func queryDigest(text string) string {
	sum := sha256.Sum256([]byte(text))
	return "sha256:" + hex.EncodeToString(sum[:])
}

func capabilityRevision(document QueryDocument) (string, error) {
	manifest, err := CapabilitiesFor(CapabilityOptions{
		Language: document.Language,
		Profile:  document.Profile,
		Version:  document.Version,
	})
	if err != nil {
		return "", err
	}
	encoded, err := json.Marshal(manifest)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(encoded)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}

func cloneRequirementSet(in RequirementSet) RequirementSet {
	out := in
	out.Coverage.Reasons = append([]string{}, in.Coverage.Reasons...)
	out.Items = append([]RequirementItem{}, in.Items...)
	for i := range out.Items {
		out.Items[i].Occurrences = append([]RequirementOccurrence{}, in.Items[i].Occurrences...)
	}
	out.Gaps = append([]RequirementGap{}, in.Gaps...)
	for i := range out.Gaps {
		out.Gaps[i].ReferenceIDs = append([]string{}, in.Gaps[i].ReferenceIDs...)
		out.Gaps[i].DiagnosticCodes = append([]string{}, in.Gaps[i].DiagnosticCodes...)
	}
	out.Diagnostics = append([]Diagnostic{}, in.Diagnostics...)
	return out
}
