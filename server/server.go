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

	errCh := make(chan error)
	go func() {
		errCh <- listener.Serve()
	}()

	select {
	case err = <-errCh:
	case <-ctx.Done():
		sherr := listener.Shutdown(ctx)
		err = errors.Join(err, sherr, <-errCh)
	}

	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}

	return errors.Join(err, ctx.Err())
}

func (srv *Server) Shutdown(ctx context.Context) error {
	l := srv.listener
	if l == nil {
		return nil
	}

	err := l.Shutdown(ctx)
	srv.listener = nil

	return err
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
