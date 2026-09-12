//go:build darwin || linux

package corpusfs

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"golang.org/x/sys/unix"
)

func TestFIFONonblockingAndDeviceRefused(t *testing.T) {
	root := physicalTemp(t)
	must(t, unix.Mkfifo(filepath.Join(root, "pipe.spl"), 0600))
	d, err := OpenRoot(root)
	must(t, err)
	defer d.Close()
	done := make(chan error, 1)
	go func() { _, err := d.ReadFile("pipe.spl"); done <- err }()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("FIFO accepted")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("FIFO open blocked without a writer")
	}
	dev, err := OpenRoot("/dev")
	must(t, err)
	defer dev.Close()
	if _, err := dev.ReadFile("null"); err == nil {
		t.Fatal("device accepted")
	}
}

func TestOpenFailuresCloseHandles(t *testing.T) {
	root := physicalTemp(t)
	must(t, os.Mkdir(filepath.Join(root, "sub"), 0700))
	must(t, os.Mkdir(filepath.Join(root, "sub", "dir.spl"), 0700))
	must(t, unix.Mkfifo(filepath.Join(root, "sub", "pipe.spl"), 0600))
	d, err := OpenRoot(root)
	must(t, err)
	defer d.Close()
	count := func() int {
		dir, err := os.Open("/dev/fd")
		must(t, err)
		defer dir.Close()
		files, err := dir.Readdirnames(-1)
		must(t, err)
		return len(files)
	}
	before := count()
	for i := 0; i < 100; i++ {
		if opened, err := OpenRoot(root + "/sub/../missing"); err == nil {
			opened.Close()
			t.Fatal("expected failed root traversal")
		}
		for _, p := range []string{"sub/missing.spl", "sub/dir.spl", "sub/pipe.spl", "sub/missing/q.spl"} {
			if _, err := d.ReadFile(p); err == nil {
				t.Fatal("expected failure")
			}
		}
	}
	if after := count(); after != before {
		t.Fatalf("descriptor leak: before %d after %d", before, after)
	}
}
