package apiserver

import (
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"
)

type server struct {
	router 		*http.ServeMux
	middleware 	http.Handler
	log         *slog.Logger
}

func newServer() *server {
	s := &server{
		router: http.NewServeMux(),
	}

	s.configureRouter()

	return s
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.middleware.ServeHTTP(w, r)
}

func (s *server) configureRouter() {
	s.middleware = s.loggerMiddleware(s.router)

	s.router.HandleFunc("/hello", s.handleHello())
}

func (s *server) loggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rw := &responseWriter{w, http.StatusOK}
		
		log := s.log.With(
			"remote_addr", r.RemoteAddr,
		)
		
		log.Info("started", slog.String("http-method", r.Method))
		
		complited := time.Since(start).String()
		log.Info("complited", slog.String("code", strconv.Itoa(rw.code)), slog.String("status-text", http.StatusText(rw.code)), slog.String("time", complited))

		next.ServeHTTP(rw, r)
	})
}

func (s *server) handleHello() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "Hello World!")
	}
}

