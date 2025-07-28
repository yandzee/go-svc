package stdrouter

import (
	"fmt"
	"net/http"

	"github.com/rs/cors"
	"github.com/yandzee/go-svc/router"
)

func Build(b *router.Builder) http.Handler {
	mux := http.NewServeMux()
	handler := http.Handler(mux)

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
			mux.Handle(route.Path, makeHandler(route.Handler))
		default:
			mux.Handle(
				fmt.Sprintf("%s %s", route.Method, route.Path),
				makeHandler(route.Handler),
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

func makeHandler(h router.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		h(&router.RequestContext{
			Request:  wrapRequest(req, res),
			Response: wrapResponse(res),
		})
	})
}

func wrapRequest(req *http.Request, res http.ResponseWriter) router.Request {
	return &Request{
		Original: req,
		Response: res,
	}
}

func wrapResponse(w http.ResponseWriter) router.Response {
	return &Response{
		Original: w,
	}
}
