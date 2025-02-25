package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/shubhamdhama/cockroach-mcp/pkg/db"
	"github.com/shubhamdhama/cockroach-mcp/pkg/server"
)

func main() {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	envPath := filepath.Join(home, ".mcp-env")
	if err := godotenv.Load(envPath); err != nil {
		log.Println("Warning: No .env file found")
	}

	db.InitDB()

	log.Println("Starting cockroach-mcp server...")
	server.Start()
}
