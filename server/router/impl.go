package router

import (
	"errors"
	"fmt"
	"iter"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/rs/cors"

	"github.com/yandzee/go-svc/httputils"
	"github.com/yandzee/go-svc/log"
)

type RouterImpl struct {
	handlers    map[string]map[string]Handler
	files       map[string]http.FileSystem
	attached    map[string]Router
	corsEnabled bool
	corsOptions *CORSOptions
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

func (ri *RouterImpl) Attach(p string, r Router) {
	if ri.attached == nil {
		ri.attached = make(map[string]Router)
	}

	ri.attached[p] = r
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
			AllowedMethods:    httputils.AllMethods,
			DisallowedMethods: []string{},
			AllowedOrigins:    []string{},
			AllowedHeaders:    []string{"*"},
			DisallowedHeaders: []string{},
			ExposedHeaders:    []string{"*"},
			DebugEnabled:      false,
			Logger:            log.Discard(),
		}
	}

	ri.corsOptions = opts
}

func (ri *RouterImpl) Files(p string, fsh http.FileSystem) {
	ri.ensureFiles()[p] = fsh
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
	}
}

func (ri *RouterImpl) Extend(rhs Router) {
	for route := range rhs.IterRoutes() {
		if route.FileSystem != nil {
			ri.Files(route.Path, route.FileSystem)
			continue
		}

		if route.Handler == nil {
			continue
		}

		switch route.Method {
		case http.MethodGet:
			ri.Get(route.Path, route.Handler)
		case http.MethodPost:
			ri.Post(route.Path, route.Handler)
		case http.MethodPut:
			ri.Put(route.Path, route.Handler)
		case http.MethodHead:
			ri.Head(route.Path, route.Handler)
		case http.MethodOptions:
			ri.Options(route.Path, route.Handler)
		case http.MethodConnect:
			ri.Connect(route.Path, route.Handler)
		case http.MethodPatch:
			ri.Patch(route.Path, route.Handler)
		case http.MethodTrace:
			ri.Trace(route.Path, route.Handler)
		}
	}
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
		}
	}

	for baseUrl, subrouter := range ri.attached {
		subhandler, err := subrouter.Handler()
		if err != nil {
			return nil, errors.Join(
				err,
				fmt.Errorf("failed to build handler for subrouter on path '%s'", baseUrl),
			)
		}

		for _, m := range httputils.AllMethods {
			root.Handler(m, baseUrl, subhandler)
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

	// 	combinedPath, err := url.JoinPath(p, route.Path)
	// 	if err != nil {
	// 		return errors.Join(
	// 			fmt.Errorf("failed to join route paths: '%s' and '%s'", p, route.Path),
	// 			err,
	// 		)
	// 	}
	//
	// 	switch {
	// 	case route.Handler != nil:
	// 		ri.ensureHandlers(route.Method)[combinedPath] = route.Handler
	// 	case route.FileSystem != nil:
	// 		ri.ensureFiles()[combinedPath] = route.FileSystem
	// 	}
	// }
	//
	// return nil
	// }

	return handler, nil
}

func (ri *RouterImpl) makeHandle(h Handler) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
		h(w, req, &HttprouterContext{
			ps: ps,
		})
	}
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
