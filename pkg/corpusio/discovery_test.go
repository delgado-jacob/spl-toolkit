//go:build darwin || linux || windows

package corpusio

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/delgado-jacob/spl-toolkit/internal/corpusfs"
)

func TestDiscoveryInvalidByteNamesDoNotPublishLossyIdentities(t *testing.T) {
	rootPath := testRoot(t)
	writeQuery(t, rootPath, "é.spl", []byte("| table host"))
	root, err := corpusfs.OpenRoot(rootPath)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	in, err := discover(root, "file:///fixture/", func(*corpusfs.Dir) ([]corpusfs.Entry, error) {
		return []corpusfs.Entry{{Name: "a\xff.spl"}, {Name: "a\xfe.spl"}, {Name: "dir\xff", Directory: true}, {Name: "link\xfe", Symlink: true}, {Name: "é.spl"}}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if in.Selection.Complete || len(in.Entries) != 1 || in.Entries[0].ID != "é.spl" || in.Entries[0].Document == nil || in.Entries[0].Document.SourceID != "é.spl" {
		t.Fatalf("unsafe identities or lost valid neighbor: %+v", in)
	}
	if len(in.Selection.SkippedSymlinks) != 0 || len(in.Selection.TraversalFailures) != 4 {
		t.Fatalf("unsafe metadata: %+v", in.Selection)
	}
	var evidence []string
	for _, f := range in.Selection.TraversalFailures {
		if f.Code != "invalid_name" || f.Phase != "traverse" || !utf8.ValidString(f.Path) || !utf8.ValidString(f.Message) {
			t.Fatalf("unsafe failure: %+v", f)
		}
		evidence = append(evidence, f.Message)
	}
	for _, originalHex := range []string{"61ff2e73706c", "61fe2e73706c", "646972ff", "6c696e6bfe"} {
		if !strings.Contains(strings.Join(evidence, "\n"), originalHex) {
			t.Fatalf("lost raw-name evidence %s: %v", originalHex, evidence)
		}
	}
	raw, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(raw, []byte(`\ufffd`)) || bytes.Contains(raw, []byte("�")) {
		t.Fatalf("lossy JSON: %s", raw)
	}
}

func TestDirectorySelectionOrder(t *testing.T) {
	root := testRoot(t)
	for _, p := range []string{"z.spl2", "a/q.spl", "a.SPL", "note.txt", ".git/hidden.spl", "a/node_modules/hidden.spl", "dist/hidden.spl", "b.spl"} {
		writeQuery(t, root, p, []byte(""))
	}
	writeQuery(t, root, "bad.spl", []byte{0xff})
	in, err := LoadDirectory(root)
	if err != nil {
		t.Fatal(err)
	}
	var ids []string
	for _, e := range in.Entries {
		ids = append(ids, e.ID)
	}
	if !reflect.DeepEqual(ids, []string{"a/q.spl", "b.spl", "bad.spl", "z.spl2"}) {
		t.Fatalf("selection %v", ids)
	}
	if in.Entries[3].Document.Language != "spl2" || in.Entries[2].Failure == nil || !in.Selection.Complete || len(in.Selection.IgnoredNames) != 10 {
		t.Fatalf("metadata: %+v", in)
	}
}

func TestTraversalFailureRetainsNeighborsAndEmptyFailure(t *testing.T) {
	for _, withNeighbor := range []bool{false, true} {
		root := testRoot(t)
		if err := os.Mkdir(filepath.Join(root, "blocked"), 0700); err != nil {
			t.Fatal(err)
		}
		if withNeighbor {
			writeQuery(t, root, "ok.spl", []byte("| table host"))
		}
		d, err := corpusfs.OpenRoot(root)
		if err != nil {
			t.Fatal(err)
		}
		in, err := discover(d, baseURI(root), func(dir *corpusfs.Dir) ([]corpusfs.Entry, error) {
			if dir != d {
				return nil, os.ErrPermission
			}
			return dir.Entries()
		})
		d.Close()
		if err != nil {
			t.Fatal(err)
		}
		if in.Selection.Complete || len(in.Selection.TraversalFailures) != 1 || in.Selection.TraversalFailures[0].Path != "blocked" || in.Selection.TraversalFailures[0].Phase != "traverse" {
			t.Fatalf("traversal failure: %+v", in)
		}
		if withNeighbor && (len(in.Entries) != 1 || in.Entries[0].Document.Text != "| table host") {
			t.Fatalf("lost neighbor: %+v", in)
		}
	}
	root := testRoot(t)
	d, err := corpusfs.OpenRoot(root)
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()
	in, err := discover(d, baseURI(root), func(*corpusfs.Dir) ([]corpusfs.Entry, error) {
		return nil, errors.New("injected root enumeration error")
	})
	if err != nil || in.Selection.Complete || len(in.Selection.TraversalFailures) != 1 {
		t.Fatalf("root traversal misclassified as empty: %+v %v", in, err)
	}
}

func TestDiscoverySkipsSymlinksAndRejectsEmpty(t *testing.T) {
	root := testRoot(t)
	outside := testRoot(t)
	writeQuery(t, outside, "q.spl", []byte("SENTINEL"))
	writeQuery(t, root, "ok.spl", []byte(""))
	if err := os.Symlink(outside, filepath.Join(root, "link")); err != nil {
		t.Skip(err)
	}
	in, err := LoadDirectory(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(in.Entries) != 1 || !reflect.DeepEqual(in.Selection.SkippedSymlinks, []string{"link"}) {
		t.Fatalf("symlinks: %+v", in)
	}
	if _, err := LoadDirectory(testRoot(t)); err == nil {
		t.Fatal("empty success")
	}
}
