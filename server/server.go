package server

import (
	"context"
	"errors"
	"net/http"

	"log/slog"
)

type ProtocolKind int

const (
	HTTP2 ProtocolKind = iota
)

type Server struct {
	Addr    string
	Kind    ProtocolKind
	Router  http.Handler
	Handler http.Handler
	Log     *slog.Logger
	SetupFn func(*http.Server)

	listener ServerListener
}

func (srv *Server) Run(ctx context.Context) error {
	rootHandler, err := srv.setupHandler()
	if err != nil {
		return err
	}

	listener, err := srv.prepareListener(ctx, rootHandler)
	if err != nil {
		return err
	}

	srv.Log.Info("running listener", "addr", srv.Addr, "kind", srv.Kind.String())
	srv.listener = listener

	err = listener.Serve()

	if errors.Is(err, http.ErrServerClosed) {
		srv.Log.Debug("Serve terminates with ErrServerClosed")
		return nil
	}

	return err
}

func (srv *Server) Shutdown(ctx context.Context) error {
	return srv.listener.Shutdown(ctx)
}

func (srv *Server) setupHandler() (http.Handler, error) {
	var handler http.Handler

	switch {
	case srv.Handler != nil:
		handler = srv.Handler
	case srv.Router != nil:
		handler = srv.Router
	}

	return handler, nil
}

func (pk ProtocolKind) String() string {
	switch pk {
	case HTTP2:
		return "HTTP2"
	default:
		return "Unknown"
	}
}
