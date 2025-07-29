package router

import (
	"io/fs"
	"log/slog"
)

type Handler func(*RequestContext)

type Route struct {
	Method     string
	Path       string
	Handler    Handler
	FileSystem fs.FS
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
