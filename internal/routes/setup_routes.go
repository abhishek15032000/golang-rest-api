package routes

import (
	"net/http"
	"rest-api/internal/handlers"
	"rest-api/internal/utils"

	_ "rest-api/docs"

	httpSwagger "github.com/swaggo/http-swagger/v2"
)

func SetupRoutes(mux *http.ServeMux, h *handlers.Handler) {
	SetupHealthRoute(mux, h)
	SetupUserRoutes(mux, h)
	
	// Swagger Documentation Route
	mux.HandleFunc("/swagger/", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		utils.RespondWithNotFound(w)
	})
}
