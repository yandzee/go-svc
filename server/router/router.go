package router

import (
	"iter"
	"log/slog"
	"net/http"
)

type Router interface {
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

	Attach(string, Router)
	Extend(Router)

	CORS(bool, ...CORSOptions)
	Handler() (http.Handler, error)

	IterRoutes() iter.Seq[*Route]
}

type Handler func(http.ResponseWriter, *http.Request, Context)

type Route struct {
	Method     string
	Path       string
	Handler    Handler
	FileSystem http.FileSystem
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
}

func New() Router {
	return &RouterImpl{}
}
