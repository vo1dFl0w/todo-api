package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/vo1dFl0w/taskmanager-api/internal/app/middleware"
	"github.com/vo1dFl0w/taskmanager-api/internal/app/model"
	"github.com/vo1dFl0w/taskmanager-api/internal/app/services"
	"github.com/vo1dFl0w/taskmanager-api/internal/app/store"
)

var (
	errMethodNotAllowed = errors.New("method not allowed")
	errIncorrectedEmailOrPassword = errors.New("incorrected email or password")
	errUnauthorized = errors.New("unauthorized")
)

type AuthHandler struct {
	Store	store.Store
	TokenService services.TokenService
	Respond	func(http.ResponseWriter, *http.Request, int, interface{})
	Error   func(http.ResponseWriter, *http.Request, int, error)
}

func (h *AuthHandler) Register() http.HandlerFunc {
	type request struct {
		Email 	 string `json:"email"`
		Password string	`json:"password"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			h.Error(w, r, http.StatusMethodNotAllowed, errMethodNotAllowed)
			return
		}

		req := &request{}

		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			h.Error(w, r, http.StatusBadRequest, err)
			return
		}

		u := &model.User{
			Email: req.Email,
			Password: req.Password,
		}

		if err := h.Store.User().Create(u); err != nil {
			h.Error(w, r, http.StatusUnprocessableEntity, err)
			return
		}

		u.Password = ""

		h.Respond(w, r, http.StatusCreated, u)
	}
}


func (h *AuthHandler) Login() http.HandlerFunc{
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
			h.Error(w, r, http.StatusMethodNotAllowed, errMethodNotAllowed)
			return
		}

		req := &request{}

		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			h.Error(w, r, http.StatusBadRequest, err)
			return
		}

		u, err := h.Store.User().FindByEmail(req.Email)
		if err != nil || !u.ComparePassword(req.Password) {
			h.Error(w, r, http.StatusUnauthorized, errIncorrectedEmailOrPassword)
			return
		}

		accessToken, err := h.TokenService.GenerateAccessToken(u.ID)
		if err != nil {
			h.Error(w, r, http.StatusInternalServerError, err)
			return
		}

		refreshToken, err := h.TokenService.GenerateRefreshToken()
		if err != nil {
			h.Error(w, r, http.StatusInternalServerError, err)
			return
		}

		expiry := time.Now().Add(30 * 24 * time.Hour)
		err = h.Store.User().SaveRefreshToken(u.ID, refreshToken, expiry)
		if err != nil {
			h.Error(w, r, http.StatusInternalServerError, err)
			return
		}

		h.Respond(w, r, http.StatusOK, loginResponse{
			AccessToken: accessToken,
			RefreshToken: refreshToken,
		})
	}
}

func (h *AuthHandler) Refresh() http.HandlerFunc {
	type request struct {
		RefreshToken string `json:"refresh_token"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var req request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.Error(w, r, http.StatusBadRequest, err)
			return
		}

		u, err := h.Store.User().GetRefreshTokenExpire(req.RefreshToken)
		if err != nil {
			h.Error(w, r, http.StatusUnauthorized, errors.New("invalid refresh token"))
			return
		}

		if time.Now().After(u.RefreshTokenExpire) {
			h.Error(w, r, http.StatusUnauthorized, errors.New("refresh token expired"))
			return
		}

		newAccessToken, err := h.TokenService.GenerateAccessToken(u.ID)
		if err != nil {
			h.Error(w, r, http.StatusInternalServerError, err)
			return
		}

		h.Respond(w, r, http.StatusOK, map[string]string{
			"access_token": newAccessToken,
		})
	}
}

func (h *AuthHandler) Whoami() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			h.Error(w, r, http.StatusMethodNotAllowed, errMethodNotAllowed)
			return
		}

		h.Respond(w, r, http.StatusOK, r.Context().Value(middleware.CtxKeyUser).(*model.User))
	}
}