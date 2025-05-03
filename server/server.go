package server

import (
	"context"
	"errors"
	"net/http"
	"net/url"

	"github.com/rs/cors"

	"chelnok-backend/internal/application/core"
	"chelnok-backend/internal/config"
	"log/slog"
)

type Server struct {
	Config      *config.Config
	Log         *slog.Logger
	Application core.Core

	listener ServerListener
}

func (srv *Server) Run(ctx context.Context) error {
	rootHandler := srv.setupHandler()

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

func (srv *Server) setupHandler() http.Handler {
	handler := http.Handler(srv.prepareRouter())

	if srv.Config.ServerCORSEnabled {
		tokenHeaders := srv.Application.Services().Auth().ExposedHTTPHeaders()

		opts := cors.Options{
			AllowedOrigins:   srv.Config.ServerCORSOrigins,
			AllowCredentials: true,
			AllowedHeaders:   []string{"*"},
			AllowedMethods: []string{
				http.MethodGet,
				http.MethodPost,
				http.MethodPut,
				http.MethodDelete,
			},
			ExposedHeaders: tokenHeaders,
			Debug:          false,
			Logger:         nil,
		}

		if srv.Config.ServerCORSDebugEnabled {
			opts.Debug = true
			opts.Logger = &corsLogger{
				Log: srv.Log.With("module", "cors"),
			}
		}

		corsServer := cors.New(opts)
		handler = corsServer.Handler(handler)
	}

	return handler
}

func (srv *Server) prefixed(p string) string {
	str, err := url.JoinPath(srv.Config.ServerAPIPrefix, p)
	if err != nil {
		panic(err.Error())
	}

	return str
}
