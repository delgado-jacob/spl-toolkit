package closure

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

// BundleDigest identifies the normalized supplied definitions, including exact source text.
func BundleDigest(bundle DefinitionBundle) (string, error) {
	canonical, err := normalizeBundle(bundle)
	if err != nil {
		return "", err
	}
	raw, err := json.Marshal(canonical)
	if err != nil {
		return "", inputError("encode bundle: %v", err)
	}
	sum := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}
