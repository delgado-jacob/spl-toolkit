// Package corpusio describes and later loads explicit local query selections.
package corpusio

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/delgado-jacob/spl-toolkit/internal/jsoninput"
	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
)

type Manifest struct {
	SchemaVersion int             `json:"schema_version"`
	Base          string          `json:"base"`
	Documents     []ManifestEntry `json:"documents"`
}

type ManifestEntry struct {
	ID       string  `json:"id"`
	Text     *string `json:"text,omitempty"`
	Path     *string `json:"path,omitempty"`
	Language string  `json:"language,omitempty"`
	Profile  string  `json:"profile,omitempty"`
	Version  string  `json:"version,omitempty"`
	SourceID string  `json:"source_id,omitempty"`
}

func manifestObject(data []byte, required []string, allowed ...string) (map[string]json.RawMessage, error) {
	if err := jsoninput.ValidateUnicode(data); err != nil {
		return nil, err
	}
	d := json.NewDecoder(bytes.NewReader(data))
	start, err := d.Token()
	if err != nil || start != json.Delim('{') {
		return nil, fmt.Errorf("expected an object")
	}
	f := map[string]json.RawMessage{}
	for d.More() {
		keyToken, err := d.Token()
		if err != nil {
			return nil, err
		}
		key := keyToken.(string)
		if _, ok := f[key]; ok {
			return nil, fmt.Errorf("duplicate property %q", key)
		}
		known := false
		for _, name := range allowed {
			if name == key {
				known = true
				break
			}
		}
		if !known {
			return nil, fmt.Errorf("unknown property %q", key)
		}
		var raw json.RawMessage
		if err := d.Decode(&raw); err != nil {
			return nil, err
		}
		f[key] = raw
	}
	if _, err := d.Token(); err != nil {
		return nil, err
	}
	var trailing any
	if err := d.Decode(&trailing); err != io.EOF {
		return nil, fmt.Errorf("expected exactly one JSON object")
	}
	for _, key := range required {
		if _, ok := f[key]; !ok {
			return nil, fmt.Errorf("missing property %q", key)
		}
	}
	return f, nil
}

func manifestString(raw []byte) (string, error) {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || raw[0] != '"' {
		return "", fmt.Errorf("expected a string")
	}
	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return "", err
	}
	return value, nil
}

func validRelative(p string, allowParent bool) bool {
	if p == "" || strings.HasPrefix(p, "/") || strings.ContainsAny(p, "\\\x00") || (len(p) >= 2 && p[1] == ':' && ((p[0] >= 'A' && p[0] <= 'Z') || (p[0] >= 'a' && p[0] <= 'z'))) {
		return false
	}
	for _, part := range strings.Split(p, "/") {
		if part == "" || (!allowParent && (part == "." || part == "..")) {
			return false
		}
	}
	return true
}

// DecodeManifest validates structure and selectors before any local acquisition.
// The base may select parent directories; each query path stays inside that base.
func DecodeManifest(data []byte) (Manifest, error) {
	f, err := manifestObject(data, []string{"schema_version", "documents"}, "schema_version", "base", "documents")
	if err != nil {
		return Manifest{}, err
	}
	if string(bytes.TrimSpace(f["schema_version"])) != "1" {
		return Manifest{}, fmt.Errorf("schema_version must be integer 1")
	}
	m := Manifest{SchemaVersion: 1, Base: ".", Documents: []ManifestEntry{}}
	if raw, ok := f["base"]; ok {
		m.Base, err = manifestString(raw)
		if err != nil || !validRelative(m.Base, true) {
			return Manifest{}, fmt.Errorf("invalid relative base")
		}
	}
	raw := bytes.TrimSpace(f["documents"])
	if len(raw) == 0 || raw[0] != '[' {
		return Manifest{}, fmt.Errorf("documents must be an array")
	}
	var items []json.RawMessage
	if err := json.Unmarshal(raw, &items); err != nil || len(items) == 0 {
		return Manifest{}, fmt.Errorf("documents must be a nonempty array")
	}
	seen := map[string]bool{}
	for i, item := range items {
		fields, err := manifestObject(item, []string{"id"}, "id", "text", "path", "language", "profile", "version", "source_id")
		if err != nil {
			return Manifest{}, fmt.Errorf("document %d: %w", i, err)
		}
		var entry ManifestEntry
		entry.ID, err = manifestString(fields["id"])
		if err != nil || strings.TrimSpace(entry.ID) == "" || seen[entry.ID] {
			return Manifest{}, fmt.Errorf("document %d has invalid or duplicate id", i)
		}
		seen[entry.ID] = true
		_, hasText := fields["text"]
		_, hasPath := fields["path"]
		if hasText == hasPath {
			return Manifest{}, fmt.Errorf("document %d requires exactly one text or path", i)
		}
		if hasText {
			value, err := manifestString(fields["text"])
			if err != nil {
				return Manifest{}, fmt.Errorf("document %d text: %w", i, err)
			}
			entry.Text = &value
		} else {
			value, err := manifestString(fields["path"])
			if err != nil || !validRelative(value, false) {
				return Manifest{}, fmt.Errorf("document %d has invalid query path", i)
			}
			entry.Path = &value
		}
		for _, option := range []struct {
			key  string
			dest *string
		}{{"language", &entry.Language}, {"profile", &entry.Profile}, {"version", &entry.Version}, {"source_id", &entry.SourceID}} {
			if value, ok := fields[option.key]; ok {
				*option.dest, err = manifestString(value)
				if err != nil {
					return Manifest{}, fmt.Errorf("document %d %s: %w", i, option.key, err)
				}
			}
		}
		if entry.Language == "" && entry.Path != nil {
			switch path.Ext(*entry.Path) {
			case ".spl":
				entry.Language = "spl"
			case ".spl2":
				entry.Language = "spl2"
			default:
				return Manifest{}, fmt.Errorf("document %d requires language for extension", i)
			}
		}
		selectors, err := analysis.CapabilitiesFor(analysis.CapabilityOptions{Language: entry.Language, Profile: entry.Profile, Version: entry.Version})
		if err != nil {
			return Manifest{}, fmt.Errorf("document %d options: %w", i, err)
		}
		entry.Language, entry.Profile, entry.Version = selectors.Language, selectors.Profile, selectors.Version
		if _, supplied := fields["source_id"]; !supplied {
			if entry.Path != nil {
				entry.SourceID = *entry.Path
			} else {
				entry.SourceID = entry.ID
			}
		}
		m.Documents = append(m.Documents, entry)
	}
	return m, nil
}
