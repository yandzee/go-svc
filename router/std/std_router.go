package stdrouter

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/klauspost/compress/gzhttp"
	"github.com/rs/cors"

	"github.com/yandzee/go-svc/data/jsoner"
	"github.com/yandzee/go-svc/router"
)

func Build(b *router.Builder) http.Handler {
	mux := http.NewServeMux()
	handler := http.Handler(mux)
	jsoner := jsoner.Jsoner{}

	for route := range b.IterRoutes() {
		p, h := preparePathAndHandler(route, &jsoner)
		mux.Handle(p, h)
	}

	if b.CORSEnabled {
		opts := cors.Options{
			AllowedOrigins:   b.CORSOptions.AllowedOrigins,
			AllowCredentials: b.CORSOptions.AllowCredentials,
			AllowedHeaders:   b.CORSOptions.AllowedHeaders,
			AllowedMethods:   b.CORSOptions.AllowedMethods,
			ExposedHeaders:   b.CORSOptions.ExposedHeaders,
			Debug:            b.CORSOptions.DebugEnabled,
			Logger:           nil,
		}

		if opts.Debug {
			opts.Logger = &corsLogger{
				Log: b.CORSOptions.Logger,
			}
		}

		corsServer := cors.New(opts)
		handler = corsServer.Handler(handler)
	}

	return handler
}

var LoggedNotFound = func(log *slog.Logger) router.Handler {
	return func(rctx *router.RequestContext) {
		log.Warn("resource is not found", "route", rctx.Request.URL().Path)
		rctx.Response.String(http.StatusNotFound)
	}
}

func preparePathAndHandler(route *router.Route, j *jsoner.Jsoner) (string, http.Handler) {
	p := route.Path
	var h http.Handler

	switch {
	case route.FileSystem != nil && len(route.FileName) > 0:
		h = http.StripPrefix(
			route.Path,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.ServeFileFS(w, r, route.FileSystem, route.FileName)
			}),
		)
	case route.FileSystem != nil:
		h = http.StripPrefix(
			route.Path,
			http.FileServerFS(route.FileSystem),
		)
	case route.Method == router.MethodAll:
		h = makeHandler(route.Handler, j)
	default:
		p = fmt.Sprintf("%s %s", route.Method, route.Path)
		h = makeHandler(route.Handler, j)
	}

	return p, gzhttp.GzipHandler(h)
}

func makeHandler(h router.Handler, j *jsoner.Jsoner) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		h(&router.RequestContext{
			Request: &Request{
				Original: req,
				Response: res,
			},
			Response: &Response{
				Original: res,
				Request:  req,
				Jsoner:   j,
			},
		})
	})
}
