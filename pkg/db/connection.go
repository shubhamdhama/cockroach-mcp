package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"

	_ "github.com/lib/pq"
)

var appDB, sysDB *sql.DB

func InitDB() {
	host := os.Getenv("COCKROACHDB_HOST")
	port := os.Getenv("COCKROACHDB_PORT")
	nPort, err := strconv.Atoi(port)
	if err != nil {
		log.Fatalf("port is invalid!")
	}

	// build connection string
	appTenantConnStr := fmt.Sprintf("postgresql://root:''@%s:%d?sslmode=disable", host, nPort)
	sysTenantConnStr := fmt.Sprintf("postgresql://root:''@%s:%d?sslmode=disable", host, nPort)

	appDB, err = sql.Open("postgres", appTenantConnStr)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	if err := appDB.Ping(); err != nil {
		log.Fatalf("DB ping failed: %v", err)
	}

	sysDB, err = sql.Open("postgres", sysTenantConnStr)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	if err := sysDB.Ping(); err != nil {
		log.Fatalf("DB ping failed: %v", err)
	}
	log.Println("Connected to CockroachDB!")
}

func GetAppDB() *sql.DB {
	return appDB
}

func GetSystemDB() *sql.DB {
	return sysDB
}
