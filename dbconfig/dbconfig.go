package dbconfig

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

func ConnectDB(databaseURL string) *sql.DB {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		log.Fatal("Failed to connect to db", err)
	}
	if err = db.Ping(); err != nil {
		log.Fatalf("Failed to ping db %v", err)
	}
	log.Println("Connected to db")
	return db
}
