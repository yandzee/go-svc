package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
)

type ServerListener interface {
	Serve() error
	Shutdown(context.Context) error
	Kind() string
}

func (srv *Server) prepareListener(ctx context.Context, h http.Handler) (ServerListener, error) {
	switch srv.Kind {
	case HTTP2:
		return srv.prepareHttp2Listener(ctx, h)
	default:
		return nil, fmt.Errorf("unknown http listener kind '%d' is specified", srv.Kind)
	}
}

func (srv *Server) prepareHttp2Listener(ctx context.Context, h http.Handler) (ServerListener, error) {
	inner := &http.Server{
		Addr:    srv.Addr,
		Handler: h,
		BaseContext: func(net.Listener) context.Context {
			return ctx
		},
		ConnContext: func(_ctx context.Context, c net.Conn) context.Context {
			return ctx
		},
	}

	if srv.SetupFn != nil {
		srv.SetupFn(inner)
	}

	return &Http2Listener{
		inner: inner,
	}, nil
}
