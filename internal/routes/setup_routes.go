package routes

import (
	"net/http"
	"rest-api/internal/handlers"
	"rest-api/internal/utils"
)

func SetupRoutes(mux *http.ServeMux, h *handlers.Handler) {
	SetupHealthRoute(mux, h)
	SetupUserRoutes(mux, h)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		utils.RespondWithNotFound(w)
	})
}
