//go:build darwin || linux || windows

package corpusio

import (
	"net/url"
	"os"
	"path/filepath"
	"testing"
)

func TestManifestParentBaseAndIgnoredOverride(t *testing.T) {
	root := testRoot(t)
	writeQuery(t, root, "queries/vendor/a #%.SPL", []byte("\t | table host\n"))
	if err := os.Mkdir(filepath.Join(root, "manifests"), 0700); err != nil {
		t.Fatal(err)
	}
	m := manifest(t, `{"schema_version":1,"base":"../queries","documents":[{"id":"override","path":"vendor/a #%.SPL","language":"spl2","source_id":""}]}`)
	in, err := LoadManifest(m, filepath.Join(root, "manifests", "m.json"))
	if err != nil {
		t.Fatal(err)
	}
	e := in.Entries[0]
	if e.Document == nil || e.Document.Text != "\t | table host\n" || e.Document.Language != "spl2" || e.Document.SourceID != "" || e.Origin.RelativePath != "vendor/a #%.SPL" {
		t.Fatalf("explicit selection: %+v", e)
	}
	u, err := url.Parse(e.Origin.BaseURI)
	if err != nil || u.Scheme != "file" || u.Path[len(u.Path)-1:] != "/" {
		t.Fatalf("base URI: %q %v", e.Origin.BaseURI, err)
	}
	if len(in.Selection.IgnoredNames) != 0 {
		t.Fatal("manifest applied discovery exclusions")
	}
}

func testRoot(t *testing.T) string {
	t.Helper()
	p, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return p
}
func writeQuery(t *testing.T, root, p string, b []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(filepath.Join(root, p)), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, p), b, 0600); err != nil {
		t.Fatal(err)
	}
}
func manifest(t *testing.T, raw string) Manifest {
	t.Helper()
	m, err := DecodeManifest([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func TestManifestContainment(t *testing.T) {
	root := testRoot(t)
	writeQuery(t, root, "queries/q.spl", []byte("\xef\xbb\xbf | table host\r\n"))
	writeQuery(t, root, "queries/bad.spl", []byte{0xff})
	m := manifest(t, `{"schema_version":1,"base":"queries","documents":[{"id":"second","path":"q.spl","source_id":"opaque"},{"id":"first","path":"q.spl"},{"id":"empty","text":""},{"id":"bad","path":"bad.spl"},{"id":"missing","path":"missing.spl"}]}`)
	in, err := LoadManifest(m, filepath.Join(root, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(in.Entries) != 5 || in.Entries[0].ID != "second" || in.Entries[1].ID != "first" || in.Entries[0].Document.Text != "\xef\xbb\xbf | table host\r\n" || in.Entries[0].Document.SourceID != "opaque" || in.Entries[2].Document.Text != "" {
		t.Fatalf("snapshot/order: %+v", in)
	}
	if in.Entries[2].Origin.Kind != "inline" || in.Entries[2].Origin.BaseURI != "" || in.Entries[3].Failure.Code != "invalid_utf8" || in.Entries[4].Failure == nil || !in.Selection.Complete {
		t.Fatalf("failure/origin: %+v", in)
	}
	writeQuery(t, root, "queries/q.spl", []byte("mutated"))
	if in.Entries[0].Document.Text != "\xef\xbb\xbf | table host\r\n" {
		t.Fatal("snapshot mutated")
	}
	bad := m
	bad.Documents = append([]ManifestEntry(nil), m.Documents...)
	escape := "../q.spl"
	bad.Documents[0].Path = &escape
	if _, err := LoadManifest(bad, filepath.Join(root, "manifest.json")); err == nil {
		t.Fatal("programmatic traversal accepted")
	}
}

func TestExplicitSymlinkFailure(t *testing.T) {
	root := testRoot(t)
	outside := testRoot(t)
	writeQuery(t, outside, "q.spl", []byte("SENTINEL"))
	writeQuery(t, root, "ok.spl", []byte(""))
	if err := os.Symlink(filepath.Join(outside, "q.spl"), filepath.Join(root, "link.spl")); err != nil {
		t.Skip(err)
	}
	m := manifest(t, `{"schema_version":1,"documents":[{"id":"link","path":"link.spl"},{"id":"ok","path":"ok.spl"}]}`)
	in, err := LoadManifest(m, filepath.Join(root, "m.json"))
	if err != nil {
		t.Fatal(err)
	}
	if in.Entries[0].Failure == nil || in.Entries[0].Document != nil || in.Entries[1].Document == nil {
		t.Fatalf("failure continuation: %+v", in)
	}
	if err := os.Symlink(root, filepath.Join(outside, "alias")); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadManifest(m, filepath.Join(outside, "alias", "m.json")); err == nil {
		t.Fatal("aliased manifest parent accepted")
	}
}

func TestManifestRejectsInvalidLocatorBeforeAcquisition(t *testing.T) {
	m := manifest(t, `{"schema_version":1,"documents":[{"id":"q","text":""}]}`)
	for _, locator := range []string{"", filepath.Join(testRoot(t), "bad\x00.json"), filepath.Join(testRoot(t), "bad\xff.json")} {
		if _, err := LoadManifest(m, locator); err == nil {
			t.Fatalf("accepted invalid locator %q", locator)
		}
	}
}
