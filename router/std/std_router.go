package stdrouter

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/klauspost/compress/gzhttp"
	"github.com/klauspost/compress/zstd"
	"github.com/rs/cors"

	"github.com/yandzee/go-svc/data/jsoner"
	"github.com/yandzee/go-svc/router"
)

var LoggedNotFound = func(log *slog.Logger) router.Handler {
	return func(rctx *router.RequestContext) {
		log.Warn(
			"resource is not found",
			"route", rctx.Request.URL().Path,
			"method", rctx.Request.Method(),
		)

		rctx.Response.String(http.StatusNotFound)
	}
}

type stdBuilder struct {
	Jsoner jsoner.Jsoner
}

func Build(b *router.Builder) http.Handler {
	builder := stdBuilder{}

	return builder.Build(b)
}

func (sb *stdBuilder) Build(b *router.Builder) http.Handler {
	mux := http.NewServeMux()
	handler := http.Handler(mux)

	for route := range b.IterRoutes() {
		p, h := sb.PreparePathAndInnerHandler(route)
		h = sb.wrapCompression(h, route.CompressionOptions, b.CompressionOptions)

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

func (b *stdBuilder) PreparePathAndInnerHandler(route *router.Route) (string, http.Handler) {
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
		h = b.wrapHandler(route.Handler)
	default:
		p = fmt.Sprintf("%s %s", route.Method, route.Path)
		h = b.wrapHandler(route.Handler)
	}

	return p, h
}

func (b *stdBuilder) wrapCompression(
	h http.Handler,
	compressionOpts ...*router.CompressionOptions,
) http.Handler {
	var opts *router.CompressionOptions
	for _, o := range compressionOpts {
		// NOTE: First nil options means that compression is disabled for the route
		if o == nil {
			break
		}

		opts = o
	}

	if opts == nil {
		return h
	}

	// NOTE: At least gzip is enabled
	gzipEnabled := opts.ZstdDisabled || !opts.GzipDisabled

	wrapper, err := gzhttp.NewWrapper(
		gzhttp.CompressionLevel(b.ensureZstdCompressionLevel(opts.ZstdCompressionLevel)),
		gzhttp.EnableZstd(!opts.ZstdDisabled),
		gzhttp.EnableGzip(gzipEnabled),
	)

	if err != nil {
		panic(err.Error())
	}

	return wrapper(h)
}

func (b *stdBuilder) ensureZstdCompressionLevel(lvl router.ZstdCompressionLevel) int {
	ensured := int(zstd.SpeedDefault)

	switch lvl {
	case router.FastestCompression:
		ensured = int(zstd.SpeedFastest)
	case router.DefaultCompression:
		ensured = int(zstd.SpeedDefault)
	case router.BetterCompression:
		ensured = int(zstd.SpeedBetterCompression)
	case router.BestCompression:
		ensured = int(zstd.SpeedBestCompression)
	}

	return ensured
}

func (b *stdBuilder) wrapHandler(h router.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		h(&router.RequestContext{
			Request: &Request{
				Original: req,
				Response: res,
			},
			Response: &Response{
				Original: res,
				Request:  req,
				Jsoner:   &b.Jsoner,
			},
		})
	})
}
