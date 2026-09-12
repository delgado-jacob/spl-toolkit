package corpusio

import (
	"path"
	"path/filepath"
	"sort"

	"github.com/delgado-jacob/spl-toolkit/internal/corpusfs"
	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/corpus"
)

var ignoredNames = []string{".git", ".hg", ".svn", ".worktrees", "node_modules", "vendor", ".venv", "venv", "build", "dist"}

func language(p string) string {
	switch path.Ext(p) {
	case ".spl":
		return "spl"
	case ".spl2":
		return "spl2"
	}
	return ""
}
func ignored(name string) bool {
	for _, n := range ignoredNames {
		if name == n {
			return true
		}
	}
	return false
}

// directoryReader keeps traversal error handling independently testable without
// relying on permission bits (which privileged hosts can bypass).
type directoryReader func(*corpusfs.Dir) ([]corpusfs.Entry, error)

// LoadDirectory selects lowercase .spl/.spl2 files in lexical relative-path order.
// Roots require physical ancestry. Traversal failures make selection incomplete;
// local file failures do not discard independently acquired neighbors.
func LoadDirectory(rootPath string) (corpus.Input, error) {
	absolute, err := filepath.Abs(rootPath)
	if err != nil {
		return corpus.Input{}, inputError("scan root: %v", err)
	}
	root, err := corpusfs.OpenRoot(rootPath)
	if err != nil {
		return corpus.Input{}, inputError("scan root: %v", err)
	}
	defer root.Close()
	return discover(root, baseURI(absolute), func(d *corpusfs.Dir) ([]corpusfs.Entry, error) { return d.Entries() })
}

func discover(root *corpusfs.Dir, uri string, readDirectory directoryReader) (corpus.Input, error) {
	out := selection("directory")
	out.Selection.IgnoredNames = append([]string(nil), ignoredNames...)
	failure := func(p string, err error) {
		e := acquisition("", p, err)
		e.Code = "traversal_failed"
		e.Phase = "traverse"
		out.Selection.Complete = false
		out.Selection.TraversalFailures = append(out.Selection.TraversalFailures, *e)
	}
	var walk func(*corpusfs.Dir, string)
	walk = func(dir *corpusfs.Dir, prefix string) {
		entries, err := readDirectory(dir)
		if err != nil {
			failure(prefix, err)
		}
		sort.Slice(entries, func(i, j int) bool { return entries[i].Name < entries[j].Name })
		for _, e := range entries {
			p := path.Join(prefix, e.Name)
			if e.Symlink {
				out.Selection.SkippedSymlinks = append(out.Selection.SkippedSymlinks, p)
				continue
			}
			if e.Directory {
				if ignored(e.Name) {
					continue
				}
				child, err := dir.OpenDir(e.Name)
				if err != nil {
					failure(p, err)
					continue
				}
				walk(child, p)
				child.Close()
				continue
			}
			lang := language(e.Name)
			if e.Err != nil {
				failure(p, e.Err)
				if lang == "" {
					continue
				}
			}
			if lang == "" {
				continue
			}
			c, _ := analysis.CapabilitiesFor(analysis.CapabilityOptions{Language: lang})
			// Read relative to the held containing directory, not the root pathname.
			name := e.Name
			item := loadedEntry(ManifestEntry{ID: p, Path: &name, Language: c.Language, Profile: c.Profile, Version: c.Version, SourceID: p}, dir, uri)
			item.Origin.RelativePath = p
			if item.Failure != nil {
				item.Failure.Path = p
			}
			out.Entries = append(out.Entries, item)
		}
	}
	walk(root, "")
	sort.Slice(out.Entries, func(i, j int) bool { return out.Entries[i].ID < out.Entries[j].ID })
	sort.Strings(out.Selection.SkippedSymlinks)
	sort.Slice(out.Selection.TraversalFailures, func(i, j int) bool {
		return out.Selection.TraversalFailures[i].Path < out.Selection.TraversalFailures[j].Path
	})
	if len(out.Entries) == 0 && out.Selection.Complete {
		return corpus.Input{}, inputError("selection contains no query documents")
	}
	return out, nil
}
