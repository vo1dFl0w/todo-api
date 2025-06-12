package apiserver

import (
	"database/sql"
	"log/slog"
	"net/http"

	_ "github.com/lib/pq"
	"github.com/vo1dFl0w/taskmanager-api/internal/app/services/logger"
	"github.com/vo1dFl0w/taskmanager-api/internal/app/store/dbstore"
)

func Launch(config *Config) error {
	db, err := newDB(config.DatabaseURL)
	if err != nil {
		return err
	}

	defer db.Close()

	store := dbstore.New(db)

	logger := logger.ConfigureLogger(config.Env)
	logger.Info("server started", slog.String("env", config.Env))
	logger.Debug("debug messages are enable")

	cfg := NewConfig()
	srv := newServer(store, logger, cfg)

	return http.ListenAndServe(config.HTTPAddr, srv)
}

func newDB(databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}


