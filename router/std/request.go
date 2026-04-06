package stdrouter

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
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

func (r *Request) Cookie(name string) *http.Cookie {
	c, err := r.Original.Cookie(name)
	if errors.Is(err, http.ErrNoCookie) {
		return nil
	}

	return c
}

func (r *Request) AllCookies() []*http.Cookie {
	return r.Original.Cookies()
}

func (r *Request) Revalidates(sum string) bool {
	h := r.Original.Header
	if noCache := strings.Contains(h.Get("Cache-Control"), "no-cache"); noCache {
		return false
	}

	if noCache := strings.Contains(h.Get("Pragma"), "no-cache"); noCache {
		return false
	}

	return h.Get("If-None-Match") == sum
}
