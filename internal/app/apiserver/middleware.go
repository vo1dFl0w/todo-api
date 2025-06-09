package apiserver

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
)

func (s *server) loggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		log := s.log.With(
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
			"complited",
			slog.Int("code", rw.code),
			slog.String("level", level.String()),
			slog.String("status-text", http.StatusText(rw.code)),
			slog.String("time", complitedStr),
		)
	})
}

func (s *server) userIdentity(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			s.error(w, r, http.StatusUnauthorized, errUnauthorized)
			return
		}

		var tokenString string
		fmt.Sscanf(authHeader, "Bearer %s", &tokenString)

		claims := &tokenClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
			return jwtSecretKey, nil
		})

		if err != nil || !token.Valid {
			s.error(w, r, http.StatusUnauthorized, errUnauthorized)
			return
		}

		u, err := s.store.User().Find(claims.UserId)
		if err != nil {
			s.error(w, r, http.StatusUnauthorized, errUnauthorized)
			return
		}

		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKeyUser, u)))
	}
}