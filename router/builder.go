package router

import (
	"fmt"
	"io/fs"
	"iter"
	"net/http"
	"strings"

	"github.com/yandzee/go-svc/httputils"
	"github.com/yandzee/go-svc/log"
)

const MethodAll = ""

type Builder struct {
	Routes      []Route
	CORSEnabled bool
	CORSOptions *CORSOptions
}

func NewBuilder() Builder {
	return Builder{}
}

func (b *Builder) Get(p string, h Handler) {
	b.ensureRoute(http.MethodGet, p, h)
}

func (b *Builder) Post(p string, h Handler) {
	b.ensureRoute(http.MethodPost, p, h)
}

func (b *Builder) Put(p string, h Handler) {
	b.ensureRoute(http.MethodPut, p, h)
}

func (b *Builder) Head(p string, h Handler) {
	b.ensureRoute(http.MethodHead, p, h)
}

func (b *Builder) Options(p string, h Handler) {
	b.ensureRoute(http.MethodOptions, p, h)
}

func (b *Builder) Delete(p string, h Handler) {
	b.ensureRoute(http.MethodDelete, p, h)
}

func (b *Builder) Connect(p string, h Handler) {
	b.ensureRoute(http.MethodConnect, p, h)
}

func (b *Builder) Patch(p string, h Handler) {
	b.ensureRoute(http.MethodPatch, p, h)
}

func (b *Builder) Trace(p string, h Handler) {
	b.ensureRoute(http.MethodTrace, p, h)
}

func (b *Builder) All(p string, h Handler) {
	b.ensureRoute(MethodAll, p, h)
}

func (b *Builder) Method(method, path string, handler Handler) {
	switch method {
	case http.MethodGet:
		b.Get(path, handler)
	case http.MethodPost:
		b.Post(path, handler)
	case http.MethodPut:
		b.Put(path, handler)
	case http.MethodHead:
		b.Head(path, handler)
	case http.MethodDelete:
		b.Delete(path, handler)
	case http.MethodOptions:
		b.Options(path, handler)
	case http.MethodConnect:
		b.Connect(path, handler)
	case http.MethodPatch:
		b.Patch(path, handler)
	case http.MethodTrace:
		b.Trace(path, handler)
	case MethodAll:
		b.All(path, handler)
	default:
		panic(fmt.Sprintf("Router: unsupported method '%s' (path: '%s')", method, path))
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

func (b *Builder) Files(p string, fs fs.FS) {
	b.ensureFiles(p, fs)
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

		if route.FileSystem != nil {
			b.Files(path, route.FileSystem)
			continue
		}

		if route.Handler == nil {
			continue
		}

		b.Method(route.Method, path, route.Handler)
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

func (b *Builder) ensureRoute(method, path string, h Handler) {
	for i := range b.Routes {
		route := &b.Routes[i]

		if route.Method != method || route.Path != path {
			continue
		}

		route.FileSystem = nil
		route.Handler = h

		return
	}

	b.Routes = append(b.Routes, Route{
		Method:  method,
		Path:    path,
		Handler: h,
	})
}

func (b *Builder) ensureFiles(path string, f fs.FS) {
	for i := range b.Routes {
		route := &b.Routes[i]

		if route.Path != path {
			continue
		}

		route.Handler = nil
		route.FileSystem = f

		return
	}

	b.Routes = append(b.Routes, Route{
		Method:     http.MethodGet,
		Path:       path,
		FileSystem: f,
	})
}
