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

	// NOTE: Router-level wrappers
	cors        *cors.Cors
	compression compressionWrapper
}

type compressionWrapper func(http.Handler) http.HandlerFunc

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

	return sb.wrapCORS(handler, b.CORSOptions)
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

func (b *stdBuilder) wrapCORS(
	h http.Handler,
	opts *router.CORSOptions,
) http.Handler {
	if opts == nil {
		return h
	}

	// NOTE: Route has no specific CORS options or they are equal to router options
	b.ensureRouterCORS(opts)
	return b.cors.Handler(h)
}

func (b *stdBuilder) ensureRouterCORS(opts *router.CORSOptions) *cors.Cors {
	if b.cors == nil {
		b.cors = b.createCORS(opts)
	}

	return b.cors
}

func (b *stdBuilder) createCORS(opts *router.CORSOptions) *cors.Cors {
	o := cors.Options{
		AllowedOrigins:   opts.AllowedOrigins,
		AllowCredentials: opts.AllowCredentials,
		AllowedHeaders:   opts.AllowedHeaders,
		AllowedMethods:   opts.AllowedMethods,
		ExposedHeaders:   opts.ExposedHeaders,
		Debug:            opts.DebugEnabled,
		Logger:           nil,
	}

	if opts.DebugEnabled {
		o.Logger = &corsLogger{
			Log: opts.Logger,
		}
	}

	return cors.New(o)
}

func (b *stdBuilder) wrapCompression(
	h http.Handler,
	routeOptions *router.CompressionOptions,
	routerOptions *router.CompressionOptions,
) http.Handler {
	if routeOptions == nil {
		return h
	} else if routerOptions != routeOptions {
		routeCompression := b.createCompression(routeOptions)
		return routeCompression(h)
	}

	if routerOptions == nil {
		return h
	}

	b.ensureRouterCompression(routerOptions)
	return b.compression(h)
}

func (b *stdBuilder) ensureRouterCompression(opts *router.CompressionOptions) compressionWrapper {
	if b.compression == nil {
		b.compression = b.createCompression(opts)
	}

	return b.compression
}

func (b *stdBuilder) createCompression(opts *router.CompressionOptions) compressionWrapper {
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

	return wrapper
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
