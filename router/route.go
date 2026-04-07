package router

import (
	"io/fs"
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

func (r *Route) Compression(enabled bool, opts ...*CompressionOptions) *Route {
	if !enabled {
		r.CompressionOptions = nil
		return r
	}

	if len(opts) > 0 {
		r.CompressionOptions = opts[0]
	} else {
		r.CompressionOptions = &CompressionOptions{}
	}

	return r
}
