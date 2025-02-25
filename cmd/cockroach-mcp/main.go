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
	// Open a log file (create if it doesn't exist, append to it)
	logFilePath := filepath.Join(home, "data", "cockroach-mcp.log")
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("failed to open log file: %v", err)
	}

	// Redirect the default logger's output to the file
	log.SetOutput(logFile)

	envPath := filepath.Join(home, ".mcp-env")
	if err := godotenv.Load(envPath); err != nil {
		log.Println("Warning: No .env file found")
	}

	db.InitDB()

	log.Println("Starting cockroach-mcp server...")
	server.Start()
}
