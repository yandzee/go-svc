package stdrouter

import (
	"context"
	"io"
	"net/http"
	"net/url"
)

type Request struct {
	Original *http.Request
	Response http.ResponseWriter
}

func (r *Request) Context() context.Context {
	return r.Original.Context()
}

func (r *Request) URL() *url.URL {
	return r.Original.URL
}

func (r *Request) PathParam(key string) (string, bool) {
	p := r.Original.PathValue(key)

	return p, len(p) > 0
}

func (r *Request) Headers() http.Header {
	return r.Original.Header
}

func (r *Request) LimitedBody(limit uint) io.ReadCloser {
	r.Original.Body = http.MaxBytesReader(r.Response, r.Original.Body, int64(limit))
	return r.Original.Body
}
