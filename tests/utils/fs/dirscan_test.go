package fs_test

import (
	"io/fs"
	"iter"
	"testing"
	"testing/fstest"

	fsutils "github.com/yandzee/go-svc/utils/fs"
)

func prepareFs() fstest.MapFS {
	tfs := make(fstest.MapFS)

	tfs["f1.txt"] = &fstest.MapFile{Data: []byte("f1.txt content")}
	tfs["f2.txt"] = &fstest.MapFile{Data: []byte("f2.txt content")}
	tfs["empty_dir"] = &fstest.MapFile{Mode: fs.ModeDir}
	tfs["dir1/f1.txt"] = &fstest.MapFile{Data: []byte("dir1/f1.txt content")}
	tfs["dir1/f2.txt"] = &fstest.MapFile{Data: []byte("dir1/f2.txt content")}
	tfs["dir1/empty_dir"] = &fstest.MapFile{Mode: fs.ModeDir}
	tfs["dir1/dir2/f1.txt"] = &fstest.MapFile{Data: []byte("dir1/dir2/f1.txt content")}
	tfs["dir1/dir2/f2.txt"] = &fstest.MapFile{Data: []byte("dir1/dir2/f2.txt content")}
	tfs["dir1/dir2/empty_dir"] = &fstest.MapFile{Mode: fs.ModeDir}

	return tfs
}

func TestScanDir(t *testing.T) {
	tfs := prepareFs()

	expectPaths(t, fsutils.ScanDir(tfs, "empty_dir"), []string{})

	expectPaths(t, fsutils.ScanDir(tfs), []string{
		"f1.txt",
		"f2.txt",
		"empty_dir",
		"dir1/f1.txt",
		"dir1/f2.txt",
		"dir1/empty_dir",
		"dir1/dir2/f1.txt",
		"dir1/dir2/f2.txt",
		"dir1/dir2/empty_dir",
	})

	expectPaths(t, fsutils.ScanDir(tfs, "dir1"), []string{
		"dir1/f1.txt",
		"dir1/f2.txt",
		"dir1/empty_dir",
		"dir1/dir2/f1.txt",
		"dir1/dir2/f2.txt",
		"dir1/dir2/empty_dir",
	})

	expectPaths(t, fsutils.ScanDir(tfs, "dir1/dir2"), []string{
		"dir1/dir2/f1.txt",
		"dir1/dir2/f2.txt",
		"dir1/dir2/empty_dir",
	})

	expectPaths(t, fsutils.ScanDir(tfs, "dir1/dir2/empty_dir"), []string{})
}

func expectPaths(
	t *testing.T,
	scan iter.Seq[fsutils.EntryResult],
	expected []string,
) {
	expectedMap := map[string]struct{}{}

	for _, exp := range expected {
		expectedMap[exp] = struct{}{}
	}

	for res := range scan {
		delete(expectedMap, res.Path)

		if res.Err != nil {
			t.Fatalf("Err on entry %v: %v\n", res.Path, res.Err)
		}
	}

	if len(expectedMap) > 0 {
		t.Fatalf("Paths expected to be in the scan: %v\n", expectedMap)
	}
}
