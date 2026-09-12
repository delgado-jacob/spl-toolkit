//go:build windows

package corpusfs

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"unicode/utf16"
	"unsafe"

	"golang.org/x/sys/windows"
)

func openRoot(root string) (*Dir, error) {
	absolute := strings.ReplaceAll(root, "/", `\`)
	if !filepath.IsAbs(absolute) {
		volume := filepath.VolumeName(absolute)
		// Resolve only the implicit current-directory prefix; never give the
		// caller's named components to filepath.Abs/GetFullPathName cleanup.
		cwd, err := filepath.Abs(volume + ".")
		if err != nil {
			return nil, err
		}
		if strings.HasPrefix(absolute, `\`) {
			absolute = filepath.VolumeName(cwd) + absolute
		} else {
			absolute = cwd + `\` + absolute[len(volume):]
		}
	}
	// Reject device namespaces; only a drive or UNC share may anchor the walk.
	if strings.HasPrefix(absolute, `\\?\`) || strings.HasPrefix(absolute, `\\.\`) {
		return nil, ErrUnsafePath
	}
	volume := filepath.VolumeName(absolute)
	if volume == "" {
		return nil, ErrUnsafePath
	}
	rootName := volume + `\`
	name, err := windows.UTF16PtrFromString(rootName)
	if err != nil {
		return nil, err
	}
	h, err := windows.CreateFile(name, windows.FILE_GENERIC_READ, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE, nil, windows.OPEN_EXISTING, windows.FILE_FLAG_BACKUP_SEMANTICS|windows.FILE_FLAG_OPEN_REPARSE_POINT, 0)
	if err != nil {
		return nil, &os.PathError{Op: "open", Path: rootName, Err: err}
	}
	f, err := admit(h, rootName, true)
	if err != nil {
		return nil, err
	}
	return walkRoot(&Dir{file: f}, strings.Split(absolute[len(volume):], `\`))
}

func openChild(parent *os.File, name string, directory bool) (*os.File, error) {
	// Native relative names must not request an alternate data stream.
	if strings.Contains(name, ":") {
		return nil, &os.PathError{Op: "open", Path: name, Err: ErrUnsafePath}
	}
	native, err := windows.NewNTUnicodeString(name)
	if err != nil {
		return nil, err
	}
	oa := windows.OBJECT_ATTRIBUTES{RootDirectory: windows.Handle(parent.Fd()), ObjectName: native, Attributes: windows.OBJ_DONT_REPARSE | windows.OBJ_CASE_INSENSITIVE}
	oa.Length = uint32(unsafe.Sizeof(oa))
	options := uint32(windows.FILE_OPEN_REPARSE_POINT | windows.FILE_SYNCHRONOUS_IO_NONALERT)
	if directory {
		options |= windows.FILE_DIRECTORY_FILE
	} else {
		options |= windows.FILE_NON_DIRECTORY_FILE
	}
	var h windows.Handle
	var status windows.IO_STATUS_BLOCK
	err = windows.NtCreateFile(&h, windows.FILE_GENERIC_READ, &oa, &status, nil, 0, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE, windows.FILE_OPEN, options, 0, 0)
	runtime.KeepAlive(parent)
	if err != nil {
		if errors.Is(err, windows.STATUS_REPARSE_POINT_ENCOUNTERED) || errors.Is(err, windows.STATUS_STOPPED_ON_SYMLINK) {
			err = ErrSymlink
		} else if status, ok := err.(windows.NTStatus); ok {
			// Preserve portable os.ErrPermission/os.ErrNotExist classification.
			err = status.Errno()
		}
		return nil, &os.PathError{Op: "open", Path: name, Err: err}
	}
	opened(name, directory)
	return admit(h, name, directory)
}

func admit(h windows.Handle, name string, directory bool) (*os.File, error) {
	fail := func(err error) (*os.File, error) {
		windows.CloseHandle(h)
		return nil, &os.PathError{Op: "stat", Path: name, Err: err}
	}
	typ, err := windows.GetFileType(h)
	if err != nil {
		return fail(err)
	}
	if typ != windows.FILE_TYPE_DISK {
		return fail(ErrNotRegular)
	}
	var info windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(h, &info); err != nil {
		return fail(err)
	}
	if info.FileAttributes&windows.FILE_ATTRIBUTE_REPARSE_POINT != 0 {
		return fail(ErrSymlink)
	}
	if (info.FileAttributes&windows.FILE_ATTRIBUTE_DIRECTORY != 0) != directory {
		return fail(ErrNotRegular)
	}
	return os.NewFile(uintptr(h), name), nil
}

// Entries enumerates the held handle using FILE_ID_BOTH_DIR_INFO, including
// after a pathname rename. Unsupported filesystems fail explicitly.
func (d *Dir) Entries() ([]Entry, error) {
	defer runtime.KeepAlive(d)
	entries := []Entry{}
	class := uint32(windows.FileIdBothDirectoryRestartInfo)
	for {
		// uint64 storage guarantees the native structure's required 8-byte alignment.
		storage := make([]uint64, 8192)
		b := unsafe.Slice((*byte)(unsafe.Pointer(&storage[0])), len(storage)*8)
		err := windows.GetFileInformationByHandleEx(windows.Handle(d.file.Fd()), class, &b[0], uint32(len(b)))
		if errors.Is(err, windows.ERROR_NO_MORE_FILES) {
			return entries, nil
		}
		if err != nil {
			return entries, err
		}
		batch, err := directoryRecords(b)
		if err != nil {
			return entries, err
		}
		entries = append(entries, batch...)
		class = windows.FileIdBothDirectoryInfo
	}
}

// FILE_ID_BOTH_DIR_INFO has a 104-byte fixed prefix on Win32 and Win64.
// Offsets are byte counts, names are non-NUL-terminated UTF-16, and nonfinal
// records must advance to an 8-byte boundary beyond the entire current name.
func directoryRecords(b []byte) ([]Entry, error) {
	entries := []Entry{}
	bad := func() ([]Entry, error) { return nil, fmt.Errorf("invalid native directory record") }
	for {
		if len(b) < 104 {
			return bad()
		}
		next := uint64(binary.LittleEndian.Uint32(b[:4]))
		attrs := binary.LittleEndian.Uint32(b[56:60])
		length := uint64(binary.LittleEndian.Uint32(b[60:64]))
		if length == 0 || length%2 != 0 || length > uint64(len(b)-104) {
			return bad()
		}
		if next != 0 && (next%8 != 0 || next < 104+length || next > uint64(len(b)-104)) {
			return bad()
		}
		units := make([]uint16, int(length/2))
		for i := range units {
			units[i] = binary.LittleEndian.Uint16(b[104+2*i:])
		}
		for i := 0; i < len(units); i++ {
			u := units[i]
			if u == 0 {
				return bad()
			}
			if u >= 0xd800 && u <= 0xdbff {
				if i+1 == len(units) || units[i+1] < 0xdc00 || units[i+1] > 0xdfff {
					return bad()
				}
				i++
			} else if u >= 0xdc00 && u <= 0xdfff {
				return bad()
			}
		}
		name := string(utf16.Decode(units))
		if name != "." && name != ".." {
			if !component(name) || strings.Contains(name, ":") {
				return bad()
			}
			entries = append(entries, Entry{Name: name, Directory: attrs&windows.FILE_ATTRIBUTE_DIRECTORY != 0, Symlink: attrs&windows.FILE_ATTRIBUTE_REPARSE_POINT != 0})
		}
		if next == 0 {
			return entries, nil
		}
		b = b[int(next):]
	}
}
