package router

import (
	"errors"
	"fmt"
	"iter"
	"net/http"
	"net/url"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/rs/cors"

	"github.com/yandzee/go-svc/httputils"
	"github.com/yandzee/go-svc/log"
)

type RouterImpl struct {
	handlers        map[string]map[string]Handler
	files           map[string]http.FileSystem
	corsEnabled     bool
	corsOptions     *CORSOptions
	notFoundHandler Handler
}

func (ri *RouterImpl) Get(p string, h Handler) {
	ri.ensureHandlers(http.MethodGet)[p] = h
}

func (ri *RouterImpl) Post(p string, h Handler) {
	ri.ensureHandlers(http.MethodPost)[p] = h
}

func (ri *RouterImpl) Put(p string, h Handler) {
	ri.ensureHandlers(http.MethodPut)[p] = h
}

func (ri *RouterImpl) Head(p string, h Handler) {
	ri.ensureHandlers(http.MethodHead)[p] = h
}

func (ri *RouterImpl) Options(p string, h Handler) {
	ri.ensureHandlers(http.MethodOptions)[p] = h
}

func (ri *RouterImpl) Delete(p string, h Handler) {
	ri.ensureHandlers(http.MethodDelete)[p] = h
}

func (ri *RouterImpl) Connect(p string, h Handler) {
	ri.ensureHandlers(http.MethodConnect)[p] = h
}

func (ri *RouterImpl) Patch(p string, h Handler) {
	ri.ensureHandlers(http.MethodPatch)[p] = h
}

func (ri *RouterImpl) Trace(p string, h Handler) {
	ri.ensureHandlers(http.MethodTrace)[p] = h
}

func (ri *RouterImpl) Method(method, path string, handler Handler) {
	switch method {
	case http.MethodGet:
		ri.Get(path, handler)
	case http.MethodPost:
		ri.Post(path, handler)
	case http.MethodPut:
		ri.Put(path, handler)
	case http.MethodHead:
		ri.Head(path, handler)
	case http.MethodDelete:
		ri.Delete(path, handler)
	case http.MethodOptions:
		ri.Options(path, handler)
	case http.MethodConnect:
		ri.Connect(path, handler)
	case http.MethodPatch:
		ri.Patch(path, handler)
	case http.MethodTrace:
		ri.Trace(path, handler)
	default:
		panic(fmt.Sprintf("Router: unsupported method '%s' (path: '%s')", method, path))
	}
}

func (ri *RouterImpl) CORS(enabled bool, maybeOpts ...CORSOptions) {
	ri.corsEnabled = enabled

	if !enabled {
		ri.corsOptions = nil
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

	ri.corsOptions = opts
}

func (ri *RouterImpl) Files(p string, fsh http.FileSystem) {
	ri.ensureFiles()[p] = fsh
}

func (ri *RouterImpl) NotFound(nfh Handler) {
	ri.notFoundHandler = nfh
}

func (ri *RouterImpl) IterRoutes() iter.Seq[*Route] {
	return func(yield func(*Route) bool) {
		for method, pathHandlers := range ri.handlers {
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

		for path, fsh := range ri.files {
			r := Route{
				Method:     http.MethodGet,
				Path:       path,
				FileSystem: fsh,
			}

			if !yield(&r) {
				return
			}
		}

		if ri.notFoundHandler != nil {
			r := Route{
				NotFoundHandler: ri.notFoundHandler,
			}

			if !yield(&r) {
				return
			}
		}
	}
}

func (ri *RouterImpl) Extend(routes iter.Seq[*Route], prefix ...string) error {
	for route := range routes {
		// NOTE: Not found handler makes sense only on empty prefix
		if route.NotFoundHandler != nil && len(prefix) == 0 {
			ri.NotFound(route.NotFoundHandler)
			continue
		}

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
			ri.Files(path, route.FileSystem)
			continue
		}

		if route.Handler == nil {
			continue
		}

		ri.Method(route.Method, path, route.Handler)
	}

	return nil
}

func (ri *RouterImpl) Handler() (http.Handler, error) {
	root := httprouter.New()
	handler := http.Handler(root)

	for route := range ri.IterRoutes() {
		switch {
		case route.Handler != nil:
			root.Handle(route.Method, route.Path, ri.makeHandle(route.Handler))
		case route.FileSystem != nil:
			root.ServeFiles(ri.makeFilesPath(route.Path), route.FileSystem)
		case route.NotFoundHandler != nil:
			root.NotFound = ri.makeHandler(route.NotFoundHandler)
		}
	}

	if ri.corsEnabled {
		opts := cors.Options{
			AllowedOrigins:   ri.corsOptions.AllowedOrigins,
			AllowCredentials: ri.corsOptions.AllowCredentials,
			AllowedHeaders:   ri.corsOptions.AllowedHeaders,
			AllowedMethods:   ri.corsOptions.AllowedMethods,
			ExposedHeaders:   ri.corsOptions.ExposedHeaders,
			Debug:            ri.corsOptions.DebugEnabled,
			Logger:           nil,
		}

		if opts.Debug {
			opts.Logger = &corsLogger{
				Log: ri.corsOptions.Logger,
			}
		}

		corsServer := cors.New(opts)
		handler = corsServer.Handler(handler)
	}

	return handler, nil
}

func (ri *RouterImpl) makeHandle(h Handler) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
		h(w, req, &HttprouterContext{
			ps: ps,
		})
	}
}

func (ri *RouterImpl) makeHandler(h Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		params := httprouter.ParamsFromContext(req.Context())

		h(w, req, &HttprouterContext{
			ps: params,
		})
	})
}

func (ri *RouterImpl) makeFilesPath(p string) string {
	if strings.HasSuffix(p, "*filepath") {
		return p
	}

	if strings.HasSuffix(p, "/") {
		return p + "*filepath"
	}

	return p + "/*filepath"
}

func (ri *RouterImpl) ensureHandlers(method string) map[string]Handler {
	if ri.handlers == nil {
		ri.handlers = make(map[string]map[string]Handler)
	}

	methodHandlers, ok := ri.handlers[method]
	if ok {
		return methodHandlers
	}

	ri.handlers[method] = make(map[string]Handler)
	return ri.handlers[method]
}

func (ri *RouterImpl) ensureFiles() map[string]http.FileSystem {
	if ri.files == nil {
		ri.files = make(map[string]http.FileSystem)
	}

	return ri.files
}
