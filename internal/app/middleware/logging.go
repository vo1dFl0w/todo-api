package middleware

import (
	"net/http"
	"log/slog"
	"time"
	"fmt"
)

func LoggerMiddleware(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			
			log := log.With(
				"remote_addr", r.RemoteAddr,
				"http-method", r.Method,
				"path", r.URL.Path,
			)

			log.Info("started")

			rw := &responseWriter{w, http.StatusOK}
			
			next.ServeHTTP(rw, r)
			
			var level slog.Level
			switch {
			case rw.code >= 500:
				level = slog.LevelError
			case rw.code >= 400:
				level = slog.LevelWarn
			default:
				level = slog.LevelInfo
			}
			
			complited := time.Since(start)
			complitedStr := fmt.Sprintf("%.3fms", float64(complited.Microseconds())/1000)

			log.Info(
				"completed",
				slog.Int("code", rw.code),
				slog.String("level", level.String()),
				slog.String("status-text", http.StatusText(rw.code)),
				slog.String("time", complitedStr),
			)
		})
	}
}

type responseWriter struct {
	http.ResponseWriter
	code int
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.code = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}