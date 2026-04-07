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
	Disabled             bool
	ZstdDisabled         bool
	ZstdCompressionLevel ZstdCompressionLevel
	GzipDisabled         bool
}

func (r *Route) Compression(opts ...*CompressionOptions) *Route {
	if len(opts) > 0 {
		r.CompressionOptions = opts[0]
	} else {
		r.CompressionOptions = &CompressionOptions{}
	}

	return r
}
