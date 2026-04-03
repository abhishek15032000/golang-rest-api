package routes

import (
	"net/http"
	"rest-api/internal/handlers"
)

func SetupUserRoutes(mux *http.ServeMux, h *handlers.Handler) {
	userMux := http.NewServeMux()
	userMux.HandleFunc("POST /register", h.CreateUser())
	userMux.HandleFunc("POST /login", h.LoginUser())
	mux.Handle("/users/", http.StripPrefix("/users", userMux))
}
