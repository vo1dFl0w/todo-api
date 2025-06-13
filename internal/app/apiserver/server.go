package apiserver

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/vo1dFl0w/taskmanager-api/internal/app/apiserver/config"
	"github.com/vo1dFl0w/taskmanager-api/internal/app/apiserver/handlers"
	"github.com/vo1dFl0w/taskmanager-api/internal/app/middleware"
	"github.com/vo1dFl0w/taskmanager-api/internal/app/services"
	"github.com/vo1dFl0w/taskmanager-api/internal/app/services/auth"
	"github.com/vo1dFl0w/taskmanager-api/internal/app/store"
)

var (
	errMethodNotAllowed = errors.New("method not allowed")
	errIncorrectedEmailOrPassword = errors.New("incorrected email or password")
	errUnauthorized = errors.New("unauthorized")
)

type Server struct {
	store 		store.Store
	router 		*http.ServeMux
	middleware 	http.Handler
	config      *config.Config
	log			*slog.Logger
	tokenService services.TokenService

}

func newServer(store store.Store, logger *slog.Logger, cfg *config.Config) *Server {
	s := &Server{
		store: store,
		router: http.NewServeMux(),
		config: cfg,
		log: logger,
		tokenService: auth.NewTokenService([]byte(cfg.JWTSecret)),
	}

	s.configureRouter()

	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.middleware.ServeHTTP(w, r)
}

func (s *Server) configureRouter() {
	authHandler := &handlers.AuthHandler{
		Store: s.store,
		TokenService: s.tokenService,
		Respond: s.respond,
		Error: s.error,
	}

	secret := []byte(s.config.JWTSecret)

	// registering regular routes
	s.router.HandleFunc("/register", authHandler.Register())
	s.router.HandleFunc("/login", authHandler.Login())
	s.router.HandleFunc("/refresh", authHandler.Refresh())

	// registering private routes
	private := http.NewServeMux()
	private.Handle("/whoami", middleware.Compose(middleware.AuthMiddleware(secret, s.store))(authHandler.Whoami()))
	s.router.Handle("/private/", http.StripPrefix("/private", private))

	// wrapping routes to middleware
	s.middleware = middleware.Compose(
		middleware.LoggerMiddleware(s.log),
		middleware.CorsMiddleware(),
	)(s.router)


}

func (s *Server) error(w http.ResponseWriter, r *http.Request, code int, err error) {
	s.respond(w, r, code, map[string]string{"error": err.Error()})
}

func (s *Server) respond(w http.ResponseWriter, r *http.Request, code int, data interface{}) {
	w.WriteHeader(code)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}


