//go:build !darwin && !linux && !windows

package corpusfs

import "os"

func openRoot(string) (*Dir, error)                      { return nil, ErrUnsupported }
func openChild(*os.File, string, bool) (*os.File, error) { return nil, ErrUnsupported }
func (d *Dir) Entries() ([]Entry, error)                 { return nil, ErrUnsupported }
