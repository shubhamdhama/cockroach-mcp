package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/shubhamdhama/cockroach-mcp/pkg/db"
	"github.com/shubhamdhama/cockroach-mcp/pkg/server"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: No .env file found")
	}

	db.InitDB()

	log.Println("Starting cockroach-mcp server...")
	server.Start()
}
