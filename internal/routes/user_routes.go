package routes

import (
	"net/http"
	"rest-api/internal/handlers"
	"rest-api/internal/middleware"
)

func SetupUserRoutes(mux *http.ServeMux, h *handlers.Handler) {
	userMux := http.NewServeMux()
	userMux.HandleFunc("POST /register", h.CreateUser())
	userMux.HandleFunc("POST /login", h.LoginUser())
	// it is going to middleware which is handleFunc , if it passes there then the middleware will itself has the handlerfunc we passed, it will move to that.
	userMux.Handle("GET /profile", middleware.AuthMiddleware(http.HandlerFunc(h.GetProfile())))
	mux.Handle("/users/", http.StripPrefix("/users", userMux))
}
