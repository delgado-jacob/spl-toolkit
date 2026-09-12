package corpus

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"unicode/utf8"

	"github.com/delgado-jacob/spl-toolkit/internal/buildinfo"
	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
)

func digestJSON(value any) (string, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:]), nil
}

// SourceHash identifies exact original UTF-8 bytes; it does not normalize line endings.
func SourceHash(text string) string {
	sum := sha256.Sum256([]byte(text))
	return hex.EncodeToString(sum[:])
}

// AnalysisRevision is scoped to the normalized document and this engine's
// versioned analysis and capability contracts. It is not a cross-edit symbol ID.
func AnalysisRevision(document analysis.QueryDocument) (string, error) {
	if !utf8.ValidString(document.Text) || !utf8.ValidString(document.SourceID) {
		return "", inputError("document text and source_id must be valid UTF-8")
	}
	capability, err := analysis.CapabilitiesFor(analysis.CapabilityOptions{Language: document.Language, Profile: document.Profile, Version: document.Version})
	if err != nil {
		return "", inputError("document options: %v", err)
	}
	return digestJSON([]any{"analysis-revision-v1", SourceHash(document.Text), capability.Language, capability.Profile, capability.Version, buildinfo.Version, "analysis-report-v1", capability})
}

// TargetDigest includes all canonical target content, including resources and
// selection options. Caller labels alone never determine refinement identity.
func TargetDigest(target *ValidationTarget) (string, error) {
	if target == nil {
		return "", nil
	}
	raw, err := json.Marshal(target)
	if err != nil {
		return "", inputError("validation target: %v", err)
	}
	normalized, err := DecodeValidationTarget(raw)
	if err != nil {
		return "", err
	}
	canonical, err := json.Marshal(normalized)
	if err != nil {
		return "", inputError("validation target: %v", err)
	}
	return digestJSON([]any{"validation-target-v1", json.RawMessage(canonical)})
}
