package apiserver

import (
	"log/slog"
	"net/http"
)

func Launch(config *Config) error {
	srv := newServer()

	srv.log = srv.configureLogger(config.Env)

	srv.log.Info("server started", slog.String("env", config.Env))

	return http.ListenAndServe(config.HTTPAddr, srv)
}


