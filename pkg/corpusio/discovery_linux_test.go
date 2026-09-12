//go:build linux

package corpusio

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLinuxInvalidByteFilesystemNames(t *testing.T) {
	root := testRoot(t)
	for _, name := range []string{"a\xff.spl", "a\xfe.spl", "é.spl"} {
		writeQuery(t, root, name, []byte("| table host"))
	}
	if err := os.Mkdir(filepath.Join(root, "dir\xff"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("é.spl", filepath.Join(root, "link\xfe")); err != nil {
		t.Fatal(err)
	}
	in, err := LoadDirectory(root)
	if err != nil {
		t.Fatal(err)
	}
	if in.Selection.Complete || len(in.Entries) != 1 || in.Entries[0].ID != "é.spl" || len(in.Selection.TraversalFailures) != 4 || len(in.Selection.SkippedSymlinks) != 0 {
		t.Fatalf("unsafe native names: %+v", in)
	}
	raw, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), `\ufffd`) || strings.Contains(string(raw), "�") {
		t.Fatalf("lossy JSON: %s", raw)
	}
}
