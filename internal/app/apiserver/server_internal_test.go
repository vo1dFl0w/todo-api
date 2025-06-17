package apiserver

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vo1dFl0w/taskmanager-api/internal/app/apiserver/config"
	"github.com/vo1dFl0w/taskmanager-api/internal/app/middleware"
	"github.com/vo1dFl0w/taskmanager-api/internal/app/model"
	"github.com/vo1dFl0w/taskmanager-api/internal/app/services/logger"
	"github.com/vo1dFl0w/taskmanager-api/internal/app/store/teststore"
)

func TestServer_AuthMiddleware(t *testing.T) {
	cfg := config.InitConfig()
	s := newServer(teststore.New(), logger.InitLogger(cfg.Env), cfg)
	u := model.TestUser(t)
	_ = s.store.User().Create(u)

	token, _ := s.tokenService.GenerateAccessToken(u.ID)

	testCases := []struct{
		name 		 string
		authHeader   string
		expectedCode int
	}{
		{
			name: "authenticated",
			authHeader: "Bearer " + token,
			expectedCode: http.StatusOK,
		},
		{
			name: "not autheticated",
			authHeader: "",
			expectedCode: http.StatusUnauthorized,
		},
		{
			name: "invalid token",
			authHeader: "invalid token",
			expectedCode: http.StatusUnauthorized,
		},
	}

	secret := []byte(s.config.JWTSecret)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "Hello, World!")
	})

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet,"/", nil)
			if tc.authHeader != "" {
				req.Header.Set("Authorization", tc.authHeader)
			}

			middleware.AuthMiddleware(secret, s.store)(handler).ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedCode, rec.Code)
		})
	}
}

func TestServer_HandleRegister(t *testing.T) {
	cfg := config.InitConfig()
	s := newServer(teststore.New(), logger.InitLogger(cfg.Env), cfg)

	testCases := []struct{
		name 		 string
		payload 	 interface{}
		expectedCode int
	}{
		{
			name: "valid",
			payload: map[string]string{
				"email": "user@example.org",
				"password": "password",
			},
			expectedCode: http.StatusCreated,
		},
		{
			name: "invalid payload",
			payload: "invalid payload",
			expectedCode: http.StatusBadRequest,
		},
				{
			name: "invalid params",
			payload: map[string]string{
				"email": "invalid",
				"password": "",
			},
			expectedCode: http.StatusUnprocessableEntity,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			b := &bytes.Buffer{}
			json.NewEncoder(b).Encode(tc.payload)
			req, _ := http.NewRequest(http.MethodPost, "/register", b)
			s.ServeHTTP(rec, req)
			assert.Equal(t, tc.expectedCode, rec.Code)
		})
	}
}

func TestServer_HandleLogin(t *testing.T) {
	cfg := config.InitConfig()
	s := newServer(teststore.New(), logger.InitLogger(cfg.Env), cfg)
	u := model.TestUser(t)
	s.store.User().Create(u)

	testCases := []struct{
		name 		 string
		payload 	 interface{}
		expectedCode int
	}{
		{
			name: "valid",
			payload: map[string]string{
				"email": u.Email,
				"password": u.Password,
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "invalid payload",
			payload: "invalid payload",
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "invalid email",
			payload: map[string]string{
				"email": "invaid",
				"password": u.Password,
			},
			expectedCode: http.StatusUnauthorized,
		},
						{
			name: "invalid password",
			payload: map[string]string{
				"email": u.Email,
				"password": "invalid",
			},
			expectedCode: http.StatusUnauthorized,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			b := &bytes.Buffer{}
			json.NewEncoder(b).Encode(tc.payload)
			req, _ := http.NewRequest(http.MethodPost, "/login", b)
			s.ServeHTTP(rec, req)
			assert.Equal(t, tc.expectedCode, rec.Code)
		})
	}
}