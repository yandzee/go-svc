package router

import (
	"errors"
	"fmt"
	"io/fs"
	"iter"
	"net/http"
	"net/url"

	"github.com/yandzee/go-svc/httputils"
	"github.com/yandzee/go-svc/log"
)

const MethodAll = ""

type Builder struct {
	handlers    map[string]map[string]Handler
	files       map[string]fs.FS
	corsEnabled bool
	corsOptions *CORSOptions
	// notFoundHandler Handler
}

func New() Builder {
	return Builder{}
}

func (b *Builder) Get(p string, h Handler) {
	b.ensureHandlers(http.MethodGet)[p] = h
}

func (b *Builder) Post(p string, h Handler) {
	b.ensureHandlers(http.MethodPost)[p] = h
}

func (b *Builder) Put(p string, h Handler) {
	b.ensureHandlers(http.MethodPut)[p] = h
}

func (b *Builder) Head(p string, h Handler) {
	b.ensureHandlers(http.MethodHead)[p] = h
}

func (b *Builder) Options(p string, h Handler) {
	b.ensureHandlers(http.MethodOptions)[p] = h
}

func (b *Builder) Delete(p string, h Handler) {
	b.ensureHandlers(http.MethodDelete)[p] = h
}

func (b *Builder) Connect(p string, h Handler) {
	b.ensureHandlers(http.MethodConnect)[p] = h
}

func (b *Builder) Patch(p string, h Handler) {
	b.ensureHandlers(http.MethodPatch)[p] = h
}

func (b *Builder) Trace(p string, h Handler) {
	b.ensureHandlers(http.MethodTrace)[p] = h
}

func (b *Builder) All(p string, h Handler) {
	b.ensureHandlers(MethodAll)[p] = h
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
	default:
		panic(fmt.Sprintf("Router: unsupported method '%s' (path: '%s')", method, path))
	}
}

func (b *Builder) CORS(enabled bool, maybeOpts ...CORSOptions) {
	b.corsEnabled = enabled

	if !enabled {
		b.corsOptions = nil
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

	b.corsOptions = opts
}

func (b *Builder) Files(p string, fs fs.FS) {
	b.ensureFiles()[p] = fs
}

// func (b *RouterBuilder) NotFound(nfh Handler) {
// 	b.notFoundHandler = nfh
// }

func (b *Builder) IterRoutes() iter.Seq[*Route] {
	return func(yield func(*Route) bool) {
		for method, pathHandlers := range b.handlers {
			for path, handler := range pathHandlers {
				r := Route{
					Method:  method,
					Path:    path,
					Handler: handler,
				}

				if !yield(&r) {
					return
				}
			}
		}

		for path, fs := range b.files {
			r := Route{
				Method:     http.MethodGet,
				Path:       path,
				FileSystem: fs,
			}

			if !yield(&r) {
				return
			}
		}

		// if b.notFoundHandler != nil {
		// 	r := Route{
		// 		NotFoundHandler: b.notFoundHandler,
		// 	}
		//
		// 	if !yield(&r) {
		// 		return
		// 	}
		// }
	}
}

func (b *Builder) Extend(routes iter.Seq[*Route], prefix ...string) error {
	for route := range routes {
		// // NOTE: Not found handler makes sense only on empty prefix
		// if route.NotFoundHandler != nil && len(prefix) == 0 {
		// 	b.NotFound(route.NotFoundHandler)
		// 	continue
		// }

		path := route.Path

		if len(prefix) > 0 {
			_path, err := url.JoinPath(prefix[0], route.Path)
			if err != nil {
				return errors.Join(
					fmt.Errorf(
						"failed to join route paths: '%s' and '%s'",
						prefix[0],
						route.Path,
					),
					err,
				)
			}

			path = _path
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

// func (b *RouterBuilder) Handler() (http.Handler, error) {
// 	root := http.NewServeMux()
// 	handler := http.Handler(root)
//
// 	for route := range b.IterRoutes() {
// 		switch {
// 		case route.Handler != nil:
// 			root.Handle(route.Method, route.Path, b.makeHandle(route.Handler))
// 		case route.FileSystem != nil:
// 			root.ServeFiles(b.makeFilesPath(route.Path), route.FileSystem)
// 		case route.NotFoundHandler != nil:
// 			root.NotFound = b.makeHandler(route.NotFoundHandler)
// 		}
// 	}
//
// 	if b.corsEnabled {
// 		opts := cors.Options{
// 			AllowedOrigins:   b.corsOptions.AllowedOrigins,
// 			AllowCredentials: b.corsOptions.AllowCredentials,
// 			AllowedHeaders:   b.corsOptions.AllowedHeaders,
// 			AllowedMethods:   b.corsOptions.AllowedMethods,
// 			ExposedHeaders:   b.corsOptions.ExposedHeaders,
// 			Debug:            b.corsOptions.DebugEnabled,
// 			Logger:           nil,
// 		}
//
// 		if opts.Debug {
// 			opts.Logger = &corsLogger{
// 				Log: b.corsOptions.Logger,
// 			}
// 		}
//
// 		corsServer := cors.New(opts)
// 		handler = corsServer.Handler(handler)
// 	}
//
// 	return handler, nil
// }

// func (b *RouterBuilder) makeHandle(h Handler) httpHandle {
// 	return func(w http.ResponseWriter, req *http.Request, ps httpParams) {
// 		h(w, req, &HttprouterContext{
// 			ps: ps,
// 		})
// 	}
// }
//
// func (b *RouterBuilder) makeHandler(h Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
// 		params := httpParamsFromContext(req.Context())
//
// 		h(w, req, &HttprouterContext{
// 			ps: params,
// 		})
// 	})
// }

// func (b *RouterBuilder) makeFilesPath(p string) string {
// 	if strings.HasSuffix(p, "*filepath") {
// 		return p
// 	}
//
// 	if strings.HasSuffix(p, "/") {
// 		return p + "*filepath"
// 	}
//
// 	return p + "/*filepath"
// }

func (b *Builder) ensureHandlers(method string) map[string]Handler {
	if b.handlers == nil {
		b.handlers = make(map[string]map[string]Handler)
	}

	methodHandlers, ok := b.handlers[method]
	if ok {
		return methodHandlers
	}

	b.handlers[method] = make(map[string]Handler)
	return b.handlers[method]
}

func (b *Builder) ensureFiles() map[string]fs.FS {
	if b.files == nil {
		b.files = make(map[string]fs.FS)
	}

	return b.files
}
