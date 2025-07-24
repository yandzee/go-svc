package stdrouter

import (
	"fmt"
	"net/http"

	"github.com/yandzee/go-svc/router"
)

func Build(b *router.Builder) *http.ServeMux {
	mux := http.NewServeMux()

	for route := range b.IterRoutes() {
		switch {
		case route.FileSystem != nil:
			mux.Handle(route.Path, http.FileServerFS(route.FileSystem))
		case route.Method == router.MethodAll:
			mux.Handle(route.Path, makeHandler(route.Handler))
		default:
			mux.Handle(
				fmt.Sprintf("%s %s", route.Method, route.Path),
				makeHandler(route.Handler),
			)
		}
	}

	return mux
}

func makeHandler(h router.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		h(&router.RequestContext{
			Request:  wrapRequest(req),
			Response: wrapResponse(res),
		})
	})
}

func wrapRequest(req *http.Request) router.Request {
	return &Request{
		Original: req,
	}
}

func wrapResponse(w http.ResponseWriter) router.Response {
	return &Response{
		Original: w,
	}
}
