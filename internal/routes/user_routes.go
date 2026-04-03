package routes

import (
	"net/http"
	"rest-api/internal/handlers"
)

func SetupUserRoutes(mux *http.ServeMux, h *handlers.Handler) {
	mux.HandleFunc("POST /user/register", h.CreateUser())
}
