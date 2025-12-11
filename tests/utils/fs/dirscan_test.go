package fs_test

import (
	"io/fs"
	"testing"
	"testing/fstest"

	fsutils "github.com/yandzee/go-svc/utils/fs"
)

func prepareFs() fstest.MapFS {
	tfs := make(fstest.MapFS)

	tfs["f1.txt"] = &fstest.MapFile{Data: []byte("f1.txt content")}
	tfs["f2.txt"] = &fstest.MapFile{Data: []byte("f2.txt content")}
	tfs["d1_empty"] = &fstest.MapFile{Mode: fs.ModeDir}
	tfs["dir1/f2.txt"] = &fstest.MapFile{Data: []byte("dir1/f2.txt content")}
	tfs["dir1/f3.txt"] = &fstest.MapFile{Data: []byte("dir1/f3.txt content")}
	tfs["dir1/d2"] = &fstest.MapFile{Mode: fs.ModeDir}
	tfs["dir1/dir2/f1.txt"] = &fstest.MapFile{Data: []byte("dir1/dir2/f1.txt content")}
	tfs["dir1/dir2/f2.txt"] = &fstest.MapFile{Data: []byte("dir1/dir2/f2.txt content")}
	tfs["dir1/dir2/d3"] = &fstest.MapFile{Mode: fs.ModeDir}

	return tfs
}

func TestScanDir(t *testing.T) {
	tfs := prepareFs()

	for entry := range fsutils.ScanDir(tfs) {
		t.Logf("entry %v\n", entry)
	}
}
