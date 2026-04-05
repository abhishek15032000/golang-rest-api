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

// @title Go Resume API
// @version 1.0
// @description This is a production-grade REST API server built with Go.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email defabhishekkumarsingh@gmail.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

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
