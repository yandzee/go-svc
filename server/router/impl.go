package router

import (
	"iter"
	"net/http"

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

// for route := range r.Inspect() {
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

func (ri *RouterImpl) Files(p string, fsh http.FileSystem) {
	ri.ensureFiles()[p] = fsh
}

func (ri *RouterImpl) Inspect() iter.Seq[*Route] {
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
