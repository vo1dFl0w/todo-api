package apiserver

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/vo1dFl0w/taskmanager-api/internal/app/model"
	"github.com/vo1dFl0w/taskmanager-api/internal/app/store"
)

var (
	errMethodNotAllowed = errors.New("method not allowed")
	errIncorrectedEmailOrPassword = errors.New("incorrected email or password")
	errUnauthorized = errors.New("unauthorized")
)

type server struct {
	store 		store.Store
	router 		*http.ServeMux
	middleware 	http.Handler
	log         *slog.Logger
}

const (
	ctxKeyUser ctxKey = iota
)

type ctxKey uint8

var (
	jwtSecretKey = []byte("your_super_secret_key")
)

type tokenClaims struct {
	jwt.StandardClaims
	UserId int `json:"user_id"`
}

func newServer(store store.Store) *server {
	s := &server{
		store: store,
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
	s.router.HandleFunc("/register", s.handleRegister())
	s.router.HandleFunc("/login", s.handleLogin())
	s.router.HandleFunc("/refresh", s.handleRefresh())

	private := http.NewServeMux()
	private.HandleFunc("/whoami", s.userIdentity(s.handleWhoami()))
	s.router.Handle("/private/", http.StripPrefix("/private", private))
}

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

func (s *server) handleHello() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "Hello World!")
	}
}

func (s *server) handleRegister() http.HandlerFunc {
	type request struct {
		Email 	 string `json:"email"`
		Password string	`json:"password"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			s.error(w, r, http.StatusMethodNotAllowed, errMethodNotAllowed)
			return
		}

		req := &request{}

		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			s.error(w, r, http.StatusBadRequest, err)
			return
		}

		u := &model.User{
			Email: req.Email,
			Password: req.Password,
		}

		if err := s.store.User().Create(u); err != nil {
			s.error(w, r, http.StatusUnprocessableEntity, err)
			return
		}

		u.Sanitize()

		s.respond(w, r, http.StatusCreated, u)

	}
}


func (s *server) handleLogin() http.HandlerFunc{
	type request struct {
		Email 		string `json:"email"`
		Password 	string `json:"password"`
	}

	type loginResponse struct {
		AccessToken string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			s.error(w, r, http.StatusMethodNotAllowed, errMethodNotAllowed)
			return
		}

		req := &request{}

		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			s.error(w, r, http.StatusBadRequest, err)
			return
		}

		u, err := s.store.User().FindByEmail(req.Email)
		if err != nil || !u.ComparePassword(req.Password) {
			s.error(w, r, http.StatusUnauthorized, errIncorrectedEmailOrPassword)
			return
		}

		accessToken, err := s.newAccessToken(u.ID)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		refreshToken, err := s.newRefreshToken()
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		expiry := time.Now().Add(30 * 24 * time.Hour)
		err = s.store.User().SaveRefreshToken(u.ID, refreshToken, expiry)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		s.respond(w, r, http.StatusOK, loginResponse{
			AccessToken: accessToken,
			RefreshToken: refreshToken,
		})
	}
}

func (s *server) handleRefresh() http.HandlerFunc {
	type request struct {
		RefreshToken string `json:"refresh_token"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var req request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			s.error(w, r, http.StatusBadRequest, err)
			return
		}

		u, err := s.store.User().FindByRefreshToken(req.RefreshToken)
		if err != nil {
			s.error(w, r, http.StatusUnauthorized, errors.New("invalid refresh token"))
			return
		}

		if time.Now().After(u.RefreshTokenExpire) {
			s.error(w, r, http.StatusUnauthorized, errors.New("refresh token expired"))
			return
		}

		newAccessToken, err := s.newAccessToken(u.ID)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		s.respond(w, r, http.StatusOK, map[string]string{
			"access_token": newAccessToken,
		})
	}
}

func (s *server) handleWhoami() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			s.error(w, r, http.StatusMethodNotAllowed, errMethodNotAllowed)
			return
		}

		s.respond(w, r, http.StatusOK, r.Context().Value(ctxKeyUser).(*model.User))
	}
}

func (s *server) newAccessToken(id int) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &tokenClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Minute * 15).Unix(),
			IssuedAt: time.Now().Unix(),
		},
		UserId: id,
	})
	
	return token.SignedString(jwtSecretKey)
}

func (s *server) newRefreshToken() (string, error) {
	b := make([]byte, 32)

	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", b), nil
}

func (s *server) error(w http.ResponseWriter, r *http.Request, code int, err error) {
	s.respond(w, r, code, map[string]string{"error": err.Error()})
}

func (s *server) respond(w http.ResponseWriter, r *http.Request, code int, data interface{}) {
	w.WriteHeader(code)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}


