package router

import (
	"io/fs"
	"log/slog"

	"github.com/yandzee/go-svc/log"
	httputils "github.com/yandzee/go-svc/utils/http"
)

type Handler func(*RequestContext)

type ZstdCompressionLevel int

const (
	FastestCompression ZstdCompressionLevel = iota + 1
	DefaultCompression
	BetterCompression
	BestCompression
)

type Route struct {
	Method     string
	Path       string
	Handler    Handler
	FileSystem fs.FS
	FileName   string

	CompressionOptions *CompressionOptions
}

type CompressionOptions struct {
	ZstdDisabled         bool
	ZstdCompressionLevel ZstdCompressionLevel
	GzipDisabled         bool
}

type CORSOptions struct {
	AllowedMethods []string `json:"allowedMethods"`
	AllowedOrigins []string `json:"allowedOrigins"`

	AllowedHeaders []string `json:"allowedHeaders"`
	ExposedHeaders []string `json:"exposedHeaders"`

	AllowCredentials bool `json:"allowCredentials"`

	DebugEnabled bool         `json:"debugEnabled"`
	Logger       *slog.Logger `json:"-"`
}

func (r *Route) Compression(enabled bool, opts ...CompressionOptions) *Route {
	r.CompressionOptions = ensureCompressionOptions(enabled, opts)

	return r
}

func ensureCORSOptions(enabled bool, opts []CORSOptions) *CORSOptions {
	if !enabled {
		return nil
	} else if len(opts) > 0 {
		return &opts[0]
	} else {
		return &CORSOptions{
			AllowedMethods: httputils.AllMethods,
			AllowedOrigins: []string{},
			AllowedHeaders: []string{"*"},
			ExposedHeaders: []string{"*"},
			DebugEnabled:   false,
			Logger:         log.Discard(),
		}
	}
}

func ensureCompressionOptions(enabled bool, opts []CompressionOptions) *CompressionOptions {
	if !enabled {
		return nil
	} else if len(opts) > 0 {
		return &opts[0]
	} else {
		return &CompressionOptions{}
	}
}
