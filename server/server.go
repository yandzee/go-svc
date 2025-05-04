package server

import (
	"context"
	"errors"
	"net/http"
	"net/url"

	"github.com/rs/cors"
	"github.com/yandzee/go-svc/server/router"

	"log/slog"
)

type ProtocolKind int

const (
	HTTP2 ProtocolKind = iota
)

type Server struct {
	Kind    ProtocolKind
	Router  router.Router
	Handler http.Handler
	Log     *slog.Logger

	listener ServerListener
}

func (srv *Server) Run(ctx context.Context) error {
	rootHandler, err := srv.setupHandler()
	if err != nil {
		return err
	}

	server, err := srv.prepareListener(ctx, rootHandler)
	if err != nil {
		return err
	}

	srv.Log.Info("running listener", "port", srv.Config.ServerPort, "kind", server.Kind())
	srv.listener = server

	err = server.Serve()

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
	var err error

	switch {
	case srv.Handler != nil:
		handler = srv.Handler
	case srv.Router != nil:
		handler, err = srv.Router.Handler()
	}

	return handler, err
}

func (srv *Server) prefixed(p string) string {
	str, err := url.JoinPath(srv.Config.ServerAPIPrefix, p)
	if err != nil {
		panic(err.Error())
	}

	return str
}
