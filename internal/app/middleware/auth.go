package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/vo1dFl0w/taskmanager-api/internal/app/store"
)

type tokenClaims struct {
	UserID int `json:"user_id"`
	jwt.StandardClaims
}

func AuthMiddleware(secret []byte, s store.Store) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "cannot extract token", http.StatusUnauthorized)
				return
			}
			var tokenStr string
			fmt.Sscanf(authHeader, "Bearer %s", &tokenStr)

			claims := &tokenClaims{}
			token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
				return secret, nil 
			})

			if err != nil || !token.Valid {
				http.Error(w, "cannot parse token or token is not valid", http.StatusUnauthorized)
				return
			}

			u, err := s.User().FindByID(claims.UserID)
			if err != nil || time.Now().After(u.RefreshTokenExpire) {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
		}
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), CtxKeyUser, u)))
		})
	}
}

