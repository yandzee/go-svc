package fs

import (
	"io/fs"
	"iter"
	"path/filepath"
)

type EntryResult struct {
	// NOTE: For path "/d1/d2/f1.dat" and start argument "d1":
	// Path is "/d1/d2/f1.dat"
	// DirPath is "/d1/d2"
	// Name is "f1.dat"
	// StartDir is "d1"

	Path     string
	Dir      string
	StartDir string
	Entry    fs.DirEntry
	Err      error
}

// NOTE: Breadth First Search over `fsys`
func ScanDir(fsys fs.FS, start ...string) iter.Seq[EntryResult] {
	dirs := []string{"."}

	if len(start) > 0 {
		dirs[0] = start[0]
	}

	startDir := dirs[0]

	return func(yield func(EntryResult) bool) {
		for len(dirs) > 0 {
			dir := dirs[0]
			dirs = dirs[1:]

			entries, err := fs.ReadDir(fsys, dir)
			if err != nil {
				if !yield(EntryResult{
					Err:      err,
					Path:     dir,
					Dir:      dir,
					StartDir: startDir,
				}) {
					return
				}

				continue
			}

			for _, entry := range entries {
				entryPath := filepath.Join(dir, entry.Name())

				if !yield(EntryResult{
					Path:     entryPath,
					Dir:      dir,
					Entry:    entry,
					StartDir: startDir,
				}) {
					return
				}

				if entry.IsDir() {
					dirs = append(dirs, entryPath)
				}
			}
		}
	}
}
