package stdrouter

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/rs/cors"

	"github.com/yandzee/go-svc/data/jsoner"
	"github.com/yandzee/go-svc/router"
)

func Build(b *router.Builder) http.Handler {
	mux := http.NewServeMux()
	handler := http.Handler(mux)
	jsoner := jsoner.Jsoner{}

	for route := range b.IterRoutes() {
		switch {
		case route.FileSystem != nil:
			mux.Handle(
				route.Path,
				http.StripPrefix(
					route.Path,
					http.FileServerFS(route.FileSystem),
				),
			)
		case route.Method == router.MethodAll:
			mux.Handle(route.Path, makeHandler(route.Handler, &jsoner))
		default:
			mux.Handle(
				fmt.Sprintf("%s %s", route.Method, route.Path),
				makeHandler(route.Handler, &jsoner),
			)
		}
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
