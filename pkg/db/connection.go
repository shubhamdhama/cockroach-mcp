package db

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var db *sql.DB

func InitDB() {
	connStr := os.Getenv("COCKROACHDB_URL")

	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("DB ping failed: %v", err)
	}
	log.Println("Connected to CockroachDB!")
}

func GetDB() *sql.DB {
	return db
}
