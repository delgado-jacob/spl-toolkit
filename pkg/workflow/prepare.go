package workflow

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"runtime/debug"
	"strconv"

	"github.com/delgado-jacob/spl-toolkit/internal/buildinfo"
	"github.com/delgado-jacob/spl-toolkit/pkg/compatibility"
	"github.com/delgado-jacob/spl-toolkit/pkg/validation"
)

// Prepared owns detached settings and reusable offline evidence. It performs no acquisition.
type Prepared struct {
	settings      Settings
	entries       map[string]EntrySettings
	compatibility *compatibility.Prepared
	provenance    Provenance
}

func Prepare(settings Settings) (*Prepared, error) {
	admitted, err := normalizeSettings(settings)
	if err != nil {
		return nil, err
	}
	evidence, err := compatibility.Prepare(admitted.Snapshot, admitted.SchemaBundle)
	if err != nil {
		return nil, preparationError(err)
	}
	raw, err := json.Marshal(admitted)
	if err != nil {
		return nil, err
	}
	digest := sha256.Sum256(raw)
	identity := evidence.ArtifactIdentity()
	provenance := Provenance{ToolkitVersion: buildinfo.Version, SettingsDigest: hex.EncodeToString(digest[:]), EnvironmentDigest: identity.EnvironmentDigest, SchemaBundleDigest: identity.SchemaBundleDigest}
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			switch setting.Key {
			case "vcs.revision":
				provenance.VCSRevision = setting.Value
			case "vcs.modified":
				if modified, err := strconv.ParseBool(setting.Value); err == nil {
					provenance.VCSModified = &modified
				}
			}
		}
	}
	prepared := &Prepared{settings: admitted, compatibility: evidence, provenance: provenance, entries: make(map[string]EntrySettings, len(admitted.Entries))}
	for _, entry := range admitted.Entries {
		prepared.entries[entry.ID] = entry
	}
	return prepared, nil
}

// Preserve the owner cause and detail at the workflow admission boundary.
func preparationError(err error) error {
	if detail, ok := compatibility.RequestErrorDetails(err); ok {
		return &validation.InputError{Err: &requestError{detail: RequestErrorDetail{Code: detail.Code, Path: detail.Path, ByteOffset: detail.ByteOffset, Message: detail.Message}, cause: err}}
	}
	return err
}
