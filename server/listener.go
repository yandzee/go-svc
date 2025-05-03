package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
)

type ServerListener interface {
	Serve() error
	Shutdown(context.Context) error
	Kind() string
}

func (srv *Server) prepareListener(ctx context.Context, h http.Handler) (ServerListener, error) {
	kind := srv.Config.ServerProtocol

	switch {
	case kind == "http2":
		return srv.prepareHttp2Listener(ctx, h)
	case kind == "http3":
		return srv.prepareHttp3Listener(ctx, h)
	default:
		return nil, fmt.Errorf("unknown http listener kind '%s' is specified", kind)
	}
}

func (srv *Server) prepareHttp2Listener(ctx context.Context, h http.Handler) (ServerListener, error) {
	tlsConfig, err := srv.Config.BuildServerTLSConfig()
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("failed to setup tls config for http2 listener"))
	}

	inner := &http.Server{
		Addr:      fmt.Sprintf(":%d", srv.Config.ServerPort),
		Handler:   h,
		TLSConfig: tlsConfig,
		BaseContext: func(net.Listener) context.Context {
			return ctx
		},
		ConnContext: func(_ctx context.Context, c net.Conn) context.Context {
			return ctx
		},
		ReadHeaderTimeout: srv.Config.ServerTimeout,
	}

	return &Http2Listener{
		inner: inner,
	}, nil
}

func (srv *Server) prepareHttp3Listener(ctx context.Context, h http.Handler) (ServerListener, error) {
	tlsConfig, err := srv.Config.BuildServerTLSConfig()
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("failed to setup tls config for http3 listener"))
	}

	inner := http3.Server{
		Logger:          srv.Log.With("module", "http3.Server"),
		Port:            int(srv.Config.ServerPort),
		EnableDatagrams: false,
		TLSConfig:       tlsConfig,
		Addr:            fmt.Sprintf(":%d", srv.Config.ServerPort),
		ConnContext: func(_ctx context.Context, _c quic.Connection) context.Context {
			return ctx
		},
	}

	inner.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor < 3 {
			if err := inner.SetQUICHeaders(w.Header()); err != nil {
				srv.Log.Error("SetQUICHeaders failed", "err", err.Error())
			}
		}

		h.ServeHTTP(w, r)
	})

	return nil, nil
}
