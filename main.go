package main

import (
	"fmt"
	"log"
	"net/http"
	"rest-api/dbconfig"
	"rest-api/internal/handlers"
	"rest-api/internal/routes"
	"rest-api/internal/store"
	"rest-api/redisconfig"
	"rest-api/serverconfig"

	"github.com/redis/go-redis/v9"
)

func main() {
	config, err := serverconfig.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// db connection

	db := dbconfig.ConnectDB(config.DatabaseURL)
	defer db.Close()

	// redis instance

	redisinstance := redisconfig.ConnectRedis()
	defer func(rdb *redis.Client) {
		_ = rdb.Close()
	}(redisinstance)
	// retuns pointers to store.Queries
	queries := store.New(db)

	// create a new handler

	handler := handlers.NewHandlers(db, queries, redisinstance)

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
