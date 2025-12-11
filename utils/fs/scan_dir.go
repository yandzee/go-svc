package fs

import (
	"io/fs"
	"iter"
)

type EntryReadError struct {
	Path string
	Err  error
}

type EntryResult struct {
	Entry     fs.DirEntry
	ReadError EntryReadError
}

func ScanDir(fsys fs.FS) iter.Seq[EntryResult] {
	dirs := []string{"."}

	return func(yield func(EntryResult) bool) {
		for len(dirs) > 0 {
			dir := dirs[0]
			dirs = dirs[1:]

			entries, err := fs.ReadDir(fsys, dir)
			if err != nil {
				if !yield(EntryResult{
					ReadError: EntryReadError{
						Err:  err,
						Path: dir,
					},
				}) {
					return
				}

				continue
			}

			for _, entry := range entries {
				if !yield(EntryResult{
					Entry: entry,
				}) {
					return
				}
			}
		}
	}
}
