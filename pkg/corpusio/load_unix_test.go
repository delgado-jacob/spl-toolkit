//go:build darwin || linux

package corpusio

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPermissionFailureContinuesManifestSelection(t *testing.T) {
	root := testRoot(t)
	writeQuery(t, root, "denied.spl", []byte("private"))
	writeQuery(t, root, "ok.spl", []byte("| table host"))
	p := filepath.Join(root, "denied.spl")
	if err := os.Chmod(p, 0); err != nil {
		t.Fatal(err)
	}
	defer os.Chmod(p, 0600)
	if _, err := os.ReadFile(p); err == nil {
		t.Skip("host bypasses file permissions; injected traversal permission test remains active")
	}
	m := manifest(t, `{"schema_version":1,"documents":[{"id":"denied","path":"denied.spl"},{"id":"ok","path":"ok.spl"}]}`)
	in, err := LoadManifest(m, filepath.Join(root, "m.json"))
	if err != nil {
		t.Fatal(err)
	}
	if in.Entries[0].Failure == nil || in.Entries[0].Failure.Code != "permission_denied" || in.Entries[0].Failure.Phase != "open" || in.Entries[1].Document == nil {
		t.Fatalf("permission continuation: %+v", in)
	}
}
