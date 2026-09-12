//go:build darwin || linux

package corpusfs

import (
	"errors"
	"os"
	"runtime"
	"strings"

	"golang.org/x/sys/unix"
)

const directoryFlags = unix.O_RDONLY | unix.O_CLOEXEC | unix.O_NOFOLLOW | unix.O_DIRECTORY

func openRoot(absolute string) (*Dir, error) {
	fd, err := unix.Open("/", directoryFlags, 0)
	if err != nil {
		return nil, err
	}
	d := &Dir{file: os.NewFile(uintptr(fd), "/")}
	for _, p := range strings.Split(strings.TrimPrefix(absolute, "/"), "/") {
		if p == "" {
			continue
		}
		child, err := d.OpenDir(p)
		d.Close()
		if err != nil {
			return nil, err
		}
		d = child
	}
	return d, nil
}

func openChild(parent *os.File, name string, directory bool) (*os.File, error) {
	flags := unix.O_RDONLY | unix.O_CLOEXEC | unix.O_NOFOLLOW | unix.O_NONBLOCK | unix.O_NOCTTY
	if directory {
		flags = directoryFlags
	}
	fd, err := unix.Openat(int(parent.Fd()), name, flags, 0)
	runtime.KeepAlive(parent)
	if err != nil {
		if errors.Is(err, unix.ELOOP) {
			err = ErrSymlink
		}
		return nil, &os.PathError{Op: "open", Path: name, Err: err}
	}
	f := os.NewFile(uintptr(fd), name)
	opened(name, directory)
	var st unix.Stat_t
	if err := unix.Fstat(fd, &st); err != nil {
		f.Close()
		return nil, &os.PathError{Op: "stat", Path: name, Err: err}
	}
	if (directory && st.Mode&unix.S_IFMT != unix.S_IFDIR) || (!directory && st.Mode&unix.S_IFMT != unix.S_IFREG) {
		f.Close()
		return nil, &os.PathError{Op: "open", Path: name, Err: ErrNotRegular}
	}
	return f, nil
}

// Readdirnames enumerates the descriptor itself. Fstatat is relative to that
// same descriptor and never follows the returned entry, even after a rename.
func (d *Dir) Entries() ([]Entry, error) {
	defer runtime.KeepAlive(d)
	names, err := d.file.Readdirnames(-1)
	entries := make([]Entry, 0, len(names))
	for _, name := range names {
		e := Entry{Name: name}
		var st unix.Stat_t
		if !component(name) {
			e.Err = ErrUnsafePath
		} else if statErr := unix.Fstatat(int(d.file.Fd()), name, &st, unix.AT_SYMLINK_NOFOLLOW); statErr != nil {
			e.Err = statErr
		} else {
			e.Directory = st.Mode&unix.S_IFMT == unix.S_IFDIR
			e.Symlink = st.Mode&unix.S_IFMT == unix.S_IFLNK
		}
		entries = append(entries, e)
	}
	return entries, err
}
