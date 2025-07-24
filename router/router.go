package router

import (
	"io/fs"
	"iter"
	"log/slog"
	"net/http"
	// "github.com/yandzee/go-svc/httputils"
)

type Router interface {
	Method(string, string, Handler)

	Get(string, Handler)
	Post(string, Handler)
	Put(string, Handler)
	Head(string, Handler)
	Options(string, Handler)
	Delete(string, Handler)
	Connect(string, Handler)
	Patch(string, Handler)
	Trace(string, Handler)

	Files(string, fs.FS)
	NotFound(Handler)

	// Second param is optional: prefix for routes to add
	Extend(iter.Seq[*Route], ...string) error

	CORS(bool, ...CORSOptions)
	Handler() (http.Handler, error)

	IterRoutes() iter.Seq[*Route]
}

type Handler func(*RequestContext)

type Request interface {
	PathParam(string) (string, bool)
}

type Response interface{}

type RequestContext struct {
	Request  Request
	Response Response
}

type Route struct {
	Method          string
	Path            string
	Handler         Handler
	FileSystem      fs.FS
	NotFoundHandler Handler
}

type CORSOptions struct {
	AllowedMethods []string
	AllowedOrigins []string

	AllowedHeaders []string
	ExposedHeaders []string

	AllowCredentials bool

	DebugEnabled bool
	Logger       *slog.Logger
}

// type Context interface {
// 	Jsoner() *httputils.Jsoner
// }

// var NotFoundHandler = Handler(func(w http.ResponseWriter, _ *http.Request, _ Context) {
// 	http.Error(w, "", http.StatusNotFound)
// })
//
// var LoggedNotFound = func(log *slog.Logger) Handler {
// 	return func(w http.ResponseWriter, r *http.Request, _ Context) {
// 		log.Warn("resource is not found", "route", r.URL.Path)
// 		http.Error(w, "", http.StatusNotFound)
// 	}
// }
