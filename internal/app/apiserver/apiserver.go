package apiserver

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"github.com/vo1dFl0w/taskmanager-api/internal/app/apiserver/config"
	"github.com/vo1dFl0w/taskmanager-api/internal/app/store/repository"
)

func Run(config *config.Config, logger *slog.Logger) error {
	db, err := initDB(config.DatabaseURL)
	if err != nil {
		return err
	}

	defer db.Close()

	store := repository.New(db)

	router := newServer(store, logger, config)

	
	s := &http.Server{
		Addr: config.HTTPAddr,
		Handler: router,
		ReadTimeout: 10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout: 120 * time.Second,
	}
	
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func()  {
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("failed to start server", slog.String("error", err.Error()))
		}
	}()

	logger.Info("server started", slog.String("env", config.Env))
	logger.Debug("debug messages are enable")

	<-done
	logger.Info("stopping server")

	ctx, cancel := context.WithTimeout(context.Background(), 10 * time.Second)
	defer cancel()

	if err := s.Shutdown(ctx); err != nil {
		logger.Error(fmt.Sprintf("failed to stop server: %s", err))
	}

	logger.Info("server stopped")

	return nil
}

func initDB(databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}


