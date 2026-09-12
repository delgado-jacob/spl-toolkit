//go:build windows

package corpusfs

import (
	"encoding/binary"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"unicode/utf16"
)

func TestWindowsMissingFilePreservesErrorKind(t *testing.T) {
	d, err := OpenRoot(physicalTemp(t))
	must(t, err)
	defer d.Close()
	if _, err := d.ReadFile("missing.spl"); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("missing file error lost: %v", err)
	}
}

func TestWindowsADSRefused(t *testing.T) {
	root := physicalTemp(t)
	put(t, filepath.Join(root, "q.spl"), "ordinary")
	must(t, os.WriteFile(filepath.Join(root, "q.spl:secret"), []byte("SENTINEL"), 0600))
	d, err := OpenRoot(root)
	must(t, err)
	defer d.Close()
	if b, err := d.ReadFile("q.spl:secret"); err == nil {
		t.Fatalf("ADS accepted: %q", b)
	}
}

func TestWindowsDirectoryRecordBounds(t *testing.T) {
	record := func(name string) []byte {
		u := utf16.Encode([]rune(name))
		b := make([]byte, 104+len(u)*2)
		binary.LittleEndian.PutUint32(b[60:64], uint32(len(u)*2))
		for i, c := range u {
			binary.LittleEndian.PutUint16(b[104+i*2:], c)
		}
		return b
	}
	good := record("é.spl")
	entries, err := directoryRecords(good)
	must(t, err)
	if len(entries) != 1 || entries[0].Name != "é.spl" {
		t.Fatalf("records: %+v", entries)
	}
	for _, tc := range []struct {
		name string
		edit func([]byte) []byte
	}{
		{"short header", func(b []byte) []byte { return b[:103] }},
		{"odd name", func(b []byte) []byte { binary.LittleEndian.PutUint32(b[60:64], 3); return b }},
		{"oversize name", func(b []byte) []byte { binary.LittleEndian.PutUint32(b[60:64], 0xffffffff); return b }},
		{"overlapping next", func(b []byte) []byte { binary.LittleEndian.PutUint32(b[:4], 8); return b }},
		{"out of bounds next", func(b []byte) []byte { binary.LittleEndian.PutUint32(b[:4], 0xfffffff8); return b }},
		{"surrogate", func(b []byte) []byte { binary.LittleEndian.PutUint16(b[104:106], 0xd800); return b }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := directoryRecords(tc.edit(record("é.spl"))); err == nil {
				t.Fatal("malformed record accepted")
			}
		})
	}
}
