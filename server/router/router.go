package router

import (
	"iter"
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

	// Static serving
	Files(string, http.FileSystem)

	// Allows to decouple Router creating and attaching to base url
	Attach(string, Router) error

	Inspect() iter.Seq[*Route]
}

type Handler func(http.ResponseWriter, *http.Request, Context)

type Route struct {
	Method     string
	Path       string
	Handler    Handler
	FileSystem http.FileSystem
}

type Context interface {
	Param(string) (string, bool)
}

func New() Router {
	return &RouterImpl{}
}
