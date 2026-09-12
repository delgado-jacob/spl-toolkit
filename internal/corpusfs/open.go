// Package corpusfs acquires local snapshots through held directory handles.
// Roots must be physical paths: symlinks/reparse points in root ancestors are
// refused. A Dir must not be used concurrently with Close or enumeration.
package corpusfs

import (
	"errors"
	"io"
	"os"
	"strings"
	"unicode/utf8"
)

var (
	ErrSymlink     = errors.New("symlink or reparse point refused")
	ErrNotRegular  = errors.New("selected object is not a regular file")
	ErrUnsafePath  = errors.New("unsafe relative path")
	ErrUnsupported = errors.New("contained acquisition unsupported on this OS")
)

type Dir struct{ file *os.File }
type Entry struct {
	Name               string
	Directory, Symlink bool
	Err                error
}

// afterOpen is a package-local deterministic barrier for filesystem race tests.
// Production leaves it nil; tests that set it must not run concurrently.
var afterOpen func(name string, directory bool)

func opened(name string, directory bool) {
	if afterOpen != nil {
		afterOpen(name, directory)
	}
}
func (d *Dir) Close() error { return d.file.Close() }

func component(name string) bool {
	return name != "" && name != "." && name != ".." && utf8.ValidString(name) && !strings.ContainsAny(name, "/\\\x00")
}

// OpenRoot checks each ancestor starting at the filesystem/volume root.
func OpenRoot(root string) (*Dir, error) {
	if root == "" || !utf8.ValidString(root) || strings.ContainsRune(root, 0) {
		return nil, ErrUnsafePath
	}
	return openRoot(root)
}

// walkRoot preserves the supplied sequence: even a named component canceled by
// a later '..' must be opened without following aliases. Parent segments pop a
// held ancestor, never a mutable native '..' entry after a directory rename.
// Ownership of anchor transfers here; every handle except the result is closed.
func walkRoot(anchor *Dir, parts []string) (*Dir, error) {
	stack := []*Dir{anchor}
	defer func() {
		for _, d := range stack {
			d.Close()
		}
	}()
	for _, part := range parts {
		switch part {
		case "", ".":
			continue
		case "..":
			if len(stack) > 1 {
				stack[len(stack)-1].Close()
				stack = stack[:len(stack)-1]
			}
		default:
			child, err := stack[len(stack)-1].OpenDir(part)
			if err != nil {
				return nil, err
			}
			stack = append(stack, child)
		}
	}
	result := stack[len(stack)-1]
	stack = stack[:len(stack)-1]
	return result, nil
}

func (d *Dir) OpenDir(name string) (*Dir, error) {
	if !component(name) {
		return nil, ErrUnsafePath
	}
	f, err := openChild(d.file, name, true)
	if err != nil {
		return nil, err
	}
	return &Dir{file: f}, nil
}

// ReadFile opens every component relative to a held parent and reads only the
// admitted regular file handle. Names are never reopened through a full path.
func (d *Dir) ReadFile(path string) ([]byte, error) {
	parts := strings.Split(path, "/")
	for _, p := range parts {
		if !component(p) {
			return nil, ErrUnsafePath
		}
	}
	parent := d
	for _, p := range parts[:len(parts)-1] {
		child, err := parent.OpenDir(p)
		if parent != d {
			parent.Close()
		}
		if err != nil {
			return nil, err
		}
		parent = child
	}
	if parent != d {
		defer parent.Close()
	}
	f, err := openChild(parent.file, parts[len(parts)-1], false)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	b, err := io.ReadAll(f)
	if err != nil {
		return nil, &os.PathError{Op: "read", Path: path, Err: err}
	}
	return b, nil
}
