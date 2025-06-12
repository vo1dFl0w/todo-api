package middleware

import "net/http"

func CorsMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Origin", "*")
			next.ServeHTTP(w, r)
		})
	}
}