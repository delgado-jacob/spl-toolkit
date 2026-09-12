//go:build darwin || linux || windows

package corpusfs

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func physicalTemp(t *testing.T) string {
	t.Helper()
	p, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return p
}
func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
func put(t *testing.T, p, text string) { t.Helper(); must(t, os.WriteFile(p, []byte(text), 0600)) }
func link(t *testing.T, target, name string) {
	t.Helper()
	if err := os.Symlink(target, name); err != nil {
		t.Skipf("symlink privilege unavailable: %v", err)
	}
}

func TestComponentSwapDoesNotEscape(t *testing.T) {
	for _, at := range []string{"parent", "leaf"} {
		t.Run(at, func(t *testing.T) {
			root := physicalTemp(t)
			must(t, os.Mkdir(filepath.Join(root, "parent"), 0700))
			outside := physicalTemp(t)
			put(t, filepath.Join(root, "parent", "q.spl"), "original")
			put(t, filepath.Join(outside, "q.spl"), "SENTINEL")
			d, err := OpenRoot(root)
			must(t, err)
			defer d.Close()
			afterOpen = func(name string, directory bool) {
				if (at == "parent" && directory && name == "parent") || (at == "leaf" && !directory && name == "q.spl") {
					afterOpen = nil
					p := filepath.Join(root, "parent")
					target := outside
					if at == "leaf" {
						p = filepath.Join(p, "q.spl")
						target = filepath.Join(outside, "q.spl")
					}
					must(t, os.Rename(p, p+"-moved"))
					link(t, target, p)
				}
			}
			defer func() { afterOpen = nil }()
			b, err := d.ReadFile("parent/q.spl")
			if err == nil && string(b) != "original" {
				t.Fatalf("escaped held object: %q", b)
			}
		})
	}
}

func TestRootAndAncestorSymlinksRefused(t *testing.T) {
	root := physicalTemp(t)
	must(t, os.Mkdir(filepath.Join(root, "real"), 0700))
	must(t, os.Mkdir(filepath.Join(root, "real", "child"), 0700))
	link(t, filepath.Join(root, "real"), filepath.Join(root, "alias"))
	for _, p := range []string{filepath.Join(root, "alias"), filepath.Join(root, "alias", "child")} {
		d, err := OpenRoot(p)
		if err == nil {
			d.Close()
			t.Fatalf("accepted aliased root %s", p)
		}
	}
}

func TestEnumerationAndReadRemainOnHeldRoot(t *testing.T) {
	base := physicalTemp(t)
	root := filepath.Join(base, "root")
	must(t, os.Mkdir(root, 0700))
	outside := physicalTemp(t)
	put(t, filepath.Join(root, "q.spl"), "original")
	put(t, filepath.Join(outside, "sentinel.spl"), "SENTINEL")
	d, err := OpenRoot(root)
	must(t, err)
	defer d.Close()
	must(t, os.Rename(root, root+"-moved"))
	link(t, outside, root)
	entries, err := d.Entries()
	must(t, err)
	if len(entries) != 1 || entries[0].Name != "q.spl" {
		t.Fatalf("escaped enumeration: %+v", entries)
	}
	b, err := d.ReadFile("q.spl")
	must(t, err)
	if string(b) != "original" {
		t.Fatalf("escaped read: %q", b)
	}
}

func TestUnsafeRelativePathsRefused(t *testing.T) {
	d, err := OpenRoot(physicalTemp(t))
	must(t, err)
	defer d.Close()
	for _, p := range []string{"../q.spl", "/q.spl", "a/../q.spl", "a\\q.spl", "a//q.spl", "./q.spl", "q\x00.spl"} {
		if _, err := d.ReadFile(p); err == nil {
			t.Errorf("accepted %q", p)
		}
	}
}

func TestLeafSwapBeforeOpenRefused(t *testing.T) {
	root := physicalTemp(t)
	must(t, os.Mkdir(filepath.Join(root, "sub"), 0700))
	put(t, filepath.Join(root, "sub", "q.spl"), "original")
	put(t, filepath.Join(root, "sentinel"), "SENTINEL")
	d, err := OpenRoot(root)
	must(t, err)
	defer d.Close()
	afterOpen = func(name string, directory bool) {
		if name == "sub" && directory {
			afterOpen = nil
			must(t, os.Remove(filepath.Join(root, "sub", "q.spl")))
			link(t, filepath.Join(root, "sentinel"), filepath.Join(root, "sub", "q.spl"))
		}
	}
	defer func() { afterOpen = nil }()
	if b, err := d.ReadFile("sub/q.spl"); err == nil {
		t.Fatalf("replacement link accepted: %q", b)
	}
}

func TestExplicitSymlinkAndDirectoryRefused(t *testing.T) {
	root := physicalTemp(t)
	put(t, filepath.Join(root, "sentinel"), "SENTINEL")
	must(t, os.Mkdir(filepath.Join(root, "directory.spl"), 0700))
	link(t, filepath.Join(root, "sentinel"), filepath.Join(root, "link.spl"))
	d, err := OpenRoot(root)
	must(t, err)
	defer d.Close()
	for _, p := range []string{"link.spl", "directory.spl"} {
		if b, err := d.ReadFile(p); err == nil {
			t.Fatalf("accepted %s: %q", p, b)
		}
	}
	entries, err := d.Entries()
	must(t, err)
	var links []string
	for _, e := range entries {
		if e.Symlink {
			links = append(links, e.Name)
		}
	}
	if !reflect.DeepEqual(links, []string{"link.spl"}) {
		t.Fatalf("links: %v", links)
	}
}
