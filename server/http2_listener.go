package server

import (
	"context"
	"net/http"
)

type Http2Listener struct {
	inner *http.Server
}

func (h2l *Http2Listener) Serve() error {
	switch {
	case h2l.inner.TLSConfig == nil:
		return h2l.inner.ListenAndServe()
	default:
		return h2l.inner.ListenAndServeTLS("", "")
	}
}

func (h2l *Http2Listener) Kind() string {
	return "http2"
}

func (h2l *Http2Listener) Shutdown(ctx context.Context) error {
	return h2l.inner.Shutdown(ctx)
}
