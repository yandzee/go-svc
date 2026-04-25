package router

import (
	"fmt"
	"io/fs"
	"iter"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/yandzee/go-svc/log"
	httputils "github.com/yandzee/go-svc/utils/http"
)

const MethodAll = ""

type Builder struct {
	Routes             []Route
	CORSEnabled        bool
	CORSOptions        *CORSOptions
	CompressionOptions *CompressionOptions
}

type CORSOptions struct {
	AllowedMethods []string `json:"allowedMethods"`
	AllowedOrigins []string `json:"allowedOrigins"`

	AllowedHeaders []string `json:"allowedHeaders"`
	ExposedHeaders []string `json:"exposedHeaders"`

	AllowCredentials bool `json:"allowCredentials"`

	DebugEnabled bool         `json:"debugEnabled"`
	Logger       *slog.Logger `json:"-"`
}

func NewBuilder() Builder {
	return Builder{}
}

func (b *Builder) Get(p string, h Handler) *Route {
	return b.ensureRoute(http.MethodGet, p, h)
}

func (b *Builder) Post(p string, h Handler) *Route {
	return b.ensureRoute(http.MethodPost, p, h)
}

func (b *Builder) Put(p string, h Handler) *Route {
	return b.ensureRoute(http.MethodPut, p, h)
}

func (b *Builder) Head(p string, h Handler) *Route {
	return b.ensureRoute(http.MethodHead, p, h)
}

func (b *Builder) Options(p string, h Handler) *Route {
	return b.ensureRoute(http.MethodOptions, p, h)
}

func (b *Builder) Delete(p string, h Handler) *Route {
	return b.ensureRoute(http.MethodDelete, p, h)
}

func (b *Builder) Connect(p string, h Handler) *Route {
	return b.ensureRoute(http.MethodConnect, p, h)
}

func (b *Builder) Patch(p string, h Handler) *Route {
	return b.ensureRoute(http.MethodPatch, p, h)
}

func (b *Builder) Trace(p string, h Handler) *Route {
	return b.ensureRoute(http.MethodTrace, p, h)
}

func (b *Builder) All(p string, h Handler) *Route {
	return b.ensureRoute(MethodAll, p, h)
}

func (b *Builder) Method(method, path string, handler Handler) *Route {
	switch method {
	case http.MethodGet:
		return b.Get(path, handler)
	case http.MethodPost:
		return b.Post(path, handler)
	case http.MethodPut:
		return b.Put(path, handler)
	case http.MethodHead:
		return b.Head(path, handler)
	case http.MethodDelete:
		return b.Delete(path, handler)
	case http.MethodOptions:
		return b.Options(path, handler)
	case http.MethodConnect:
		return b.Connect(path, handler)
	case http.MethodPatch:
		return b.Patch(path, handler)
	case http.MethodTrace:
		return b.Trace(path, handler)
	case MethodAll:
		return b.All(path, handler)
	default:
		panic(fmt.Sprintf("Router: unsupported method '%s' (path: '%s')", method, path))
	}
}

func (b *Builder) Compression(enabled bool, opts ...*CompressionOptions) {
	if !enabled {
		b.CompressionOptions = nil

		for route := range b.IterRoutes() {
			route.CompressionOptions = nil
		}

		return
	}

	if len(opts) > 0 {
		b.CompressionOptions = opts[0]
	} else {
		b.CompressionOptions = &CompressionOptions{}
	}
}

func (b *Builder) CORS(enabled bool, maybeOpts ...CORSOptions) {
	b.CORSEnabled = enabled

	if !enabled {
		b.CORSOptions = nil
		return
	}

	var opts *CORSOptions

	switch {
	case len(maybeOpts) > 0:
		opts = &maybeOpts[0]
	default:
		opts = &CORSOptions{
			AllowedMethods: httputils.AllMethods,
			AllowedOrigins: []string{},
			AllowedHeaders: []string{"*"},
			ExposedHeaders: []string{"*"},
			DebugEnabled:   false,
			Logger:         log.Discard(),
		}
	}

	b.CORSOptions = opts
}

func (b *Builder) Files(p string, fs fs.FS) *Route {
	return b.ensureFiles(p, fs)
}

func (b *Builder) File(p string, fs fs.FS, fileName ...string) *Route {
	fname := ""
	if len(fileName) > 0 {
		fname = fileName[0]
	}

	if len(fname) == 0 {
		fname = filepath.Base(p)
	}

	return b.ensureFiles(p, fs, fname)
}

func (b *Builder) IterRoutes() iter.Seq[*Route] {
	return func(yield func(*Route) bool) {
		for i := range b.Routes {
			if !yield(&b.Routes[i]) {
				return
			}
		}
	}
}

func (b *Builder) Extend(routes iter.Seq[*Route], prefixes ...string) error {
	for route := range routes {
		path := route.Path

		if len(prefixes) > 0 {
			prefix := prefixes[0]
			path = b.joinPathParts(prefix, path)
		}

		var r *Route

		switch {
		case route.FileSystem != nil && len(route.FileName) > 0:
			r = b.File(path, route.FileSystem, route.FileName)
		case route.FileSystem != nil:
			r = b.Files(path, route.FileSystem)
		case route.Handler != nil:
			r = b.Method(route.Method, path, route.Handler)
		}

		if r != nil {
			r.CompressionOptions = route.CompressionOptions
		}
	}

	return nil
}

func (b *Builder) joinPathParts(p, q string) string {
	switch {
	case strings.HasSuffix(p, "/"):
		if !strings.HasPrefix(q, "/") {
			return p + q
		}

		return p + strings.TrimLeft(q, "/")
	case strings.HasPrefix(q, "/"):
		return p + q
	default:
		return p + "/" + q
	}
}

func (b *Builder) ensureRoute(method, path string, h Handler) *Route {
	for i := range b.Routes {
		route := &b.Routes[i]

		if route.Method != method || route.Path != path {
			continue
		}

		route.FileSystem = nil
		route.Handler = h

		return route
	}

	b.Routes = append(b.Routes, Route{
		Method:             method,
		Path:               path,
		Handler:            h,
		CompressionOptions: b.CompressionOptions,
	})

	return &b.Routes[len(b.Routes)-1]
}

func (b *Builder) ensureFiles(path string, f fs.FS, fname ...string) *Route {
	fileName := ""
	if len(fname) > 0 {
		fileName = fname[0]
	}

	for i := range b.Routes {
		route := &b.Routes[i]

		if route.Path != path {
			continue
		}

		route.Handler = nil
		route.FileSystem = f
		route.FileName = fileName

		return route
	}

	b.Routes = append(b.Routes, Route{
		Method:             http.MethodGet,
		Path:               path,
		FileSystem:         f,
		FileName:           fileName,
		CompressionOptions: b.CompressionOptions,
	})

	return &b.Routes[len(b.Routes)-1]
}
