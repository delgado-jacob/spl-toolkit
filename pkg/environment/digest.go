package environment

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func snapshotDigest(snapshot Snapshot) (string, error) {
	snapshot.Digest = ""
	raw, err := json.Marshal(snapshot)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}
