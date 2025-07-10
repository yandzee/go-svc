package router

import (
	"iter"
	"log/slog"
	"net/http"

	"github.com/yandzee/go-svc/httputils"
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

	Files(string, http.FileSystem)
	NotFound(Handler)

	// Second param is optional: prefix for routes to add
	Extend(iter.Seq[*Route], ...string) error

	CORS(bool, ...CORSOptions)
	Handler() (http.Handler, error)

	IterRoutes() iter.Seq[*Route]
}

type Handler func(http.ResponseWriter, *http.Request, Context)

type Route struct {
	Method          string
	Path            string
	Handler         Handler
	FileSystem      http.FileSystem
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

type Context interface {
	Param(string) (string, bool)
	Jsoner() *httputils.Jsoner
}

var NotFoundHandler = Handler(func(w http.ResponseWriter, _ *http.Request, _ Context) {
	http.Error(w, "", http.StatusNotFound)
})

func New() Router {
	return &RouterImpl{}
}
