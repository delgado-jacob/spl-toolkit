//go:build darwin || linux || windows

package corpusio

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/delgado-jacob/spl-toolkit/internal/corpusfs"
)

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
