package routes

import (
	"net/http"
	"rest-api/internal/handlers"
)

func SetupRoutes(mux *http.ServeMux, h *handlers.Handler) {
	SetupHealthRoute(mux, h)
}
