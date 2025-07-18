package dir_scanner

import (
	"io/fs"
	"iter"
	"log/slog"
	"path/filepath"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/yandzee/go-svc/log"
)

type DirScanner struct {
	EnableWatcher bool
	Log           *slog.Logger

	paths    []string
	lastRead map[string][]ScannedEntry
	fsw      *fsnotify.Watcher
	stop     chan struct{}
	mx       sync.RWMutex
}

// NOTE: This is simply the args of fs.WalkDirFunc
type ScannedEntry struct {
	// NOTE: Not 'AbsolutePath' because whether paths are absolute depends on
	// the value passed to `Add()`
	JoinedPath string

	// NOTE: `DirPath` is Base of `JoinedPath` + `RelativePath`
	DirPath string

	RootPath     string
	RelativePath string

	Name  string
	Entry fs.DirEntry
	Err   error
}

func (ds *DirScanner) Add(p ...string) error {
	ds.mx.Lock()
	defer ds.mx.Unlock()

	if ds.EnableWatcher && ds.fsw == nil {
		if err := ds.setupWatcher(); err != nil {
			return err
		}

		for _, fsPath := range p {
			if err := ds.fsw.Add(fsPath); err != nil {
				return err
			}
		}
	}

	ds.paths = append(ds.paths, p...)
	return nil
}

func (ds *DirScanner) IterFiles() iter.Seq2[string, ScannedEntry] {
	ds.mx.Lock()

	content := map[string][]ScannedEntry{}

	for _, dirPath := range ds.paths {
		cached, ok := ds.lastRead[dirPath]
		if ok {
			content[dirPath] = cached
			continue
		}

		scanned, err := ds.readDir(dirPath)
		if err != nil {
			log.OrDiscard(ds.Log).Error(
				"readDir failure",
				"dirPath", dirPath,
				log.Error("err", err),
			)

			continue
		}

		content[dirPath] = scanned
	}

	ds.lastRead = content
	ds.mx.Unlock()

	return func(yield func(string, ScannedEntry) bool) {
		for dirPath, entries := range content {
			for _, entry := range entries {
				if !yield(dirPath, entry) {
					return
				}
			}
		}
	}
}

func (ds *DirScanner) Stop() error {
	ds.mx.Lock()
	defer ds.mx.Unlock()

	if ds.fsw != nil {
		if err := ds.fsw.Close(); err != nil {
			return err
		}
	}

	select {
	case <-ds.stop:
	default:
		close(ds.stop)
	}

	return nil
}

func (ds *DirScanner) readDir(rootPath string) ([]ScannedEntry, error) {
	return Scan(rootPath)
}

func Scan(p string) ([]ScannedEntry, error) {
	scanned := []ScannedEntry{}

	err := filepath.WalkDir(p, func(fpath string, e fs.DirEntry, err error) error {
		relPath, _ := filepath.Rel(p, fpath)
		dirPath := filepath.Join("/", filepath.Base(p), relPath)

		name := ""
		if e != nil {
			name = e.Name()
		}

		scanned = append(scanned, ScannedEntry{
			JoinedPath:   fpath,
			RelativePath: relPath,
			DirPath:      dirPath,
			RootPath:     p,
			Name:         name,
			Entry:        e,
			Err:          err,
		})

		return err
	})

	return scanned, err
}

func (ds *DirScanner) setupWatcher() error {
	if ds.fsw != nil {
		return nil
	}

	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	ds.fsw = fsw
	ds.stop = make(chan struct{})

	go ds.listenWatcher()

	return nil
}

func (ds *DirScanner) listenWatcher() {
	logger := log.OrDiscard(ds.Log)
	defer logger.Debug("listenWatcher: terminated")

	for {
		select {
		case evt, ok := <-ds.fsw.Events:
			if !ok {
				return
			}

			logger.Debug("listenWatcher: event", "event", evt)

			ds.mx.Lock()
			for k := range ds.lastRead {
				delete(ds.lastRead, k)
			}

			ds.mx.Unlock()
		case err, ok := <-ds.fsw.Errors:
			if !ok {
				return
			}

			logger.Debug("listenWatcher: error", log.Error("err", err))
		case <-ds.stop:
			return
		}
	}
}
