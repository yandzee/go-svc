package router

import (
	"io/fs"
	"log/slog"
	"net/http"
)

type Handler func(*RequestContext)

type Request interface {
	Headers() http.Header
	PathParam(string) (string, bool)
}

type Response interface {
	Status(int, ...string)
	Statusf(int, string, ...any)
}

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
