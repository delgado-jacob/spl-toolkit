package corpusio

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/delgado-jacob/spl-toolkit/internal/capabilityselector"
	"github.com/delgado-jacob/spl-toolkit/internal/corpusfs"
	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/corpus"
)

func inputError(format string, args ...any) error {
	return &corpus.InputError{Err: fmt.Errorf(format, args...)}
}
func selection(mode string) corpus.Input {
	return corpus.Input{Entries: []corpus.Entry{}, Selection: corpus.Selection{Mode: mode, Complete: true, IgnoredNames: []string{}, SkippedSymlinks: []string{}, TraversalFailures: []corpus.AcquisitionError{}}}
}

// validateManifest also protects callers constructing exported Go structs.
// Do not JSON-roundtrip here: Marshal replaces invalid UTF-8 and omitempty loses
// the explicit empty opaque source ID accepted by DecodeManifest.
func validateManifest(m Manifest) (Manifest, error) {
	if m.SchemaVersion != 1 || len(m.Documents) == 0 {
		return Manifest{}, inputError("manifest requires schema_version 1 and nonempty documents")
	}
	if m.Base == "" {
		m.Base = "."
	}
	if !utf8.ValidString(m.Base) || !validRelative(m.Base, true) {
		return Manifest{}, inputError("invalid relative base")
	}
	m.Documents = append([]ManifestEntry(nil), m.Documents...)
	seen := map[string]bool{}
	for i := range m.Documents {
		e := &m.Documents[i]
		if !utf8.ValidString(e.ID) || strings.TrimSpace(e.ID) == "" || seen[e.ID] || (e.Text == nil) == (e.Path == nil) {
			return Manifest{}, inputError("document %d has invalid identity or source", i)
		}
		seen[e.ID] = true
		if !utf8.ValidString(e.SourceID) || (e.Text != nil && !utf8.ValidString(*e.Text)) {
			return Manifest{}, inputError("document %d has invalid UTF-8", i)
		}
		if e.Path != nil {
			if !utf8.ValidString(*e.Path) || !validRelative(*e.Path, false) {
				return Manifest{}, inputError("document %d has invalid query path", i)
			}
			if e.Language == "" {
				e.Language = language(*e.Path)
				if e.Language == "" {
					return Manifest{}, inputError("document %d requires language", i)
				}
			}
		}
		selectors, err := capabilityselector.Normalize(e.Language, e.Profile, e.Version)
		if err != nil {
			return Manifest{}, inputError("document %d options: %v", i, err)
		}
		e.Language, e.Profile, e.Version = selectors.Language, selectors.Profile, selectors.Version
	}
	return m, nil
}

func baseURI(root string) string {
	p := filepath.ToSlash(root)
	if strings.HasPrefix(p, "//") {
		parts := strings.SplitN(strings.TrimPrefix(p, "//"), "/", 2)
		if len(parts) == 2 {
			return (&url.URL{Scheme: "file", Host: parts[0], Path: "/" + strings.TrimRight(parts[1], "/") + "/"}).String()
		}
	}
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return (&url.URL{Scheme: "file", Path: strings.TrimRight(p, "/") + "/"}).String()
}

func acquisition(id, path string, err error) *corpus.AcquisitionError {
	code, phase := "acquisition_failed", "open"
	switch {
	case errors.Is(err, corpusfs.ErrSymlink):
		code = "symlink_refused"
	case errors.Is(err, corpusfs.ErrNotRegular):
		code = "nonregular_file"
	case errors.Is(err, os.ErrPermission):
		code = "permission_denied"
	case errors.Is(err, os.ErrNotExist):
		code = "not_found"
	}
	var pe *os.PathError
	if errors.As(err, &pe) {
		phase = pe.Op
	}
	return &corpus.AcquisitionError{Code: code, Phase: phase, Message: err.Error(), ID: id, Path: path}
}

func loadedEntry(e ManifestEntry, root *corpusfs.Dir, uri string) corpus.Entry {
	out := corpus.Entry{ID: e.ID, Origin: corpus.Origin{Kind: "inline"}}
	var text string
	if e.Path == nil {
		text = *e.Text
	} else {
		out.Origin = corpus.Origin{Kind: "file", RelativePath: *e.Path, BaseURI: uri}
		b, err := root.ReadFile(*e.Path)
		if err != nil {
			out.Failure = acquisition(e.ID, *e.Path, err)
			return out
		}
		if !utf8.Valid(b) {
			out.Failure = &corpus.AcquisitionError{Code: "invalid_utf8", Phase: "decode", Message: "selected file is not valid UTF-8", ID: e.ID, Path: *e.Path}
			return out
		}
		text = string(b)
	}
	out.Document = &analysis.QueryDocument{Text: text, Language: e.Language, Profile: e.Profile, Version: e.Version, SourceID: e.SourceID}
	out.SourceHash = corpus.SourceHash(text)
	return out
}

// LoadManifest reads snapshots from the declared base relative to manifestPath.
// The manifest parent and selected base must have physical (non-symlink) ancestry.
// Structural/root errors abort; individual selected file failures remain entries.
func LoadManifest(manifest Manifest, manifestPath string) (corpus.Input, error) {
	if manifestPath == "" || !utf8.ValidString(manifestPath) || strings.ContainsRune(manifestPath, 0) {
		return corpus.Input{}, inputError("invalid manifest path")
	}
	m, err := validateManifest(manifest)
	if err != nil {
		return corpus.Input{}, err
	}
	// Split retains canceled components in the locator parent. Dir/Join/Abs
	// would erase them before the no-follow walk could check their objects.
	parent, _ := filepath.Split(manifestPath)
	if parent == "" {
		parent = "." + string(filepath.Separator)
	}
	// Check the supplied parent even when a base with '..' chooses another root.
	p, err := corpusfs.OpenRoot(parent)
	if err != nil {
		return corpus.Input{}, inputError("manifest parent: %v", err)
	}
	p.Close()
	base := parent + filepath.FromSlash(m.Base)
	root, err := corpusfs.OpenRoot(base)
	if err != nil {
		return corpus.Input{}, inputError("manifest base: %v", err)
	}
	defer root.Close()
	absolute, err := filepath.Abs(base)
	if err != nil {
		return corpus.Input{}, inputError("manifest base: %v", err)
	}
	out := selection("manifest")
	uri := baseURI(absolute)
	for _, e := range m.Documents {
		out.Entries = append(out.Entries, loadedEntry(e, root, uri))
	}
	return out, nil
}
