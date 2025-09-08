package apiserver

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

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

	taskHandler := &handlers.TaskHandler{
		Store: s.store,
		TokenService: s.tokenService,
		Respond: s.respond,
		Error: s.error,
	}

	secret := []byte(s.config.JWTSecret)

	// registration of authorization routs
	s.router.HandleFunc("/register", authHandler.Register())
	s.router.HandleFunc("/login", authHandler.Login())
	s.router.HandleFunc("/refresh", authHandler.Refresh())

	// registration of private test routs
	private := http.NewServeMux()
	private.Handle("/whoami", middleware.Compose(middleware.AuthMiddleware(secret, s.store))(authHandler.Whoami()))
	s.router.Handle("/private/", http.StripPrefix("/private", private))

	// registration of task (todo) routs
	s.router.Handle("/user/", middleware.Compose(middleware.AuthMiddleware(secret, s.store))(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := strings.Trim(r.URL.Path, "/")
			parts := strings.FieldsFunc(path, func(r rune) bool {
				return r == '/' || r == '?' || r == '=' || r == '&'
			})
			
			// expect /user/{user_id}/task
			if len(parts) == 3 && parts[0] == "user" && parts[2] == "task" {
				userID, err := strconv.Atoi(parts[1])
				if err != nil {
					s.error(w, r, http.StatusBadRequest, errors.New("invalid user_id"))
					return
				}

				switch r.Method {
				case http.MethodGet:
					taskHandler.GetTask(userID)(w, r)
				case http.MethodPost:
					taskHandler.CreateTask(userID)(w, r)
				case http.MethodDelete:
					query := r.URL.Query()["ids"]
					var taskIDs []int
					for _, idStr := range query {
						id, err := strconv.Atoi(idStr)
						if err != nil {
							s.error(w, r, http.StatusBadRequest, err)
							return
						}
						taskIDs = append(taskIDs, id)
					}
					taskHandler.DeleteTask(userID, taskIDs)(w, r)
				default:
					s.error(w, r, http.StatusMethodNotAllowed, errMethodNotAllowed)
				}
				return
			}
			
			// expect /user/{user_id}/task/{task_id} or /user/{user_id}/task?ids=1&ids=2
			if len(parts) == 4 && parts[0] == "user" && parts[2] == "task" {
				userID, err1 := strconv.Atoi(parts[1])
				taskID, err2 := strconv.Atoi(parts[3])
				if err1 != nil || err2 != nil {
					s.error(w, r, http.StatusBadRequest, errors.New("invalid user_id or task_id"))
					return
				}

				switch r.Method {
				case http.MethodPatch:
					taskHandler.UpdateTask(userID, taskID)(w, r)
				default:
					s.error(w, r, http.StatusMethodNotAllowed, errMethodNotAllowed)
				}
				return
			}


			http.NotFound(w, r)
		}),
	))

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