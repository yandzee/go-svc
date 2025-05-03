package server

import (
	"github.com/quic-go/quic-go/http3"
)

type Http3Listener struct {
	inner *http3.Server
}

func (h3l *Http3Listener) Serve() error {
	return h3l.inner.ListenAndServeTLS("", "")
}

func (h3l *Http3Listener) Kind() string {
	return "http3"
}
