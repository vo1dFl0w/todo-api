package apiserver

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"time"

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

func newServer(store store.Store, logger *slog.Logger) *server {
	s := &server{
		store: store,
		router: http.NewServeMux(),
		log: logger,
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

func (s *server) error(w http.ResponseWriter, r *http.Request, code int, err error) {
	s.respond(w, r, code, map[string]string{"error": err.Error()})
}

func (s *server) respond(w http.ResponseWriter, r *http.Request, code int, data interface{}) {
	w.WriteHeader(code)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}


