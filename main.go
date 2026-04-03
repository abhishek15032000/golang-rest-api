package main

import (
	"fmt"
	"log"
	"net/http"
	"rest-api/dbconfig"
	"rest-api/internal/handlers"
	"rest-api/internal/routes"
	"rest-api/serverconfig"
)

func main() {
	config, err := serverconfig.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// db connection

	db := dbconfig.ConnectDB(config.DatabaseURL)
	defer db.Close()

	// create a new handler

	handler := handlers.NewHandlers()

	// set up the http server
	mux := http.NewServeMux()

	// setup routes
	routes.SetupRoutes(mux, handler)

	// server instance
	serverAddr := fmt.Sprintf(":%s", config.ServerPort)
	server := &http.Server{
		Addr:    serverAddr,
		Handler: mux,
	}
	fmt.Println("App is running on port ", config.ServerPort)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
