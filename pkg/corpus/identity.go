package corpus

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"unicode/utf8"

	"github.com/delgado-jacob/spl-toolkit/internal/buildinfo"
	"github.com/delgado-jacob/spl-toolkit/internal/capabilityselector"
	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
)

type analysisRevisionContract struct {
	prefix []byte
	suffix []byte
}

var analysisRevisionContracts = mustBuildAnalysisRevisionContracts()

func mustBuildAnalysisRevisionContracts() map[capabilityselector.Selection]analysisRevisionContract {
	const sourceHashMarker = "__spl_toolkit_source_hash__"
	marker := []byte(`"` + sourceHashMarker + `"`)
	contracts := make(map[capabilityselector.Selection]analysisRevisionContract, 2)
	for _, language := range []string{"spl", "spl2"} {
		selection, err := capabilityselector.Normalize(language, "splunkd", "current")
		if err != nil {
			panic(fmt.Sprintf("normalize %s analysis revision selectors: %v", language, err))
		}
		manifest, err := analysis.CapabilitiesFor(analysis.CapabilityOptions{Language: selection.Language, Profile: selection.Profile, Version: selection.Version})
		if err != nil {
			panic(fmt.Sprintf("load %s analysis revision contract: %v", language, err))
		}
		encoded, err := json.Marshal([]any{"analysis-revision-v1", sourceHashMarker, selection.Language, selection.Profile, selection.Version, buildinfo.Version, "analysis-report-v1", manifest})
		if err != nil {
			panic(fmt.Sprintf("encode %s analysis revision contract: %v", language, err))
		}
		markerAt := bytes.Index(encoded, marker)
		if markerAt < 0 {
			panic(fmt.Sprintf("encode %s analysis revision contract: source hash marker missing", language))
		}
		contracts[selection] = analysisRevisionContract{
			prefix: append([]byte(nil), encoded[:markerAt]...),
			suffix: append([]byte(nil), encoded[markerAt+len(marker):]...),
		}
	}
	return contracts
}

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
	selection, err := capabilityselector.Normalize(document.Language, document.Profile, document.Version)
	if err != nil {
		return "", inputError("document options: %v", err)
	}
	contract, found := analysisRevisionContracts[selection]
	if !found {
		return "", inputError("document options: capability contract is unavailable for language %q, profile %q, version %q", selection.Language, selection.Profile, selection.Version)
	}
	hash := sha256.New()
	_, _ = hash.Write(contract.prefix)
	_, _ = hash.Write([]byte(`"`))
	_, _ = hash.Write([]byte(SourceHash(document.Text)))
	_, _ = hash.Write([]byte(`"`))
	_, _ = hash.Write(contract.suffix)
	return hex.EncodeToString(hash.Sum(nil)), nil
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
