package main

import (
	"io"
	"log"
	"os"
	"os/user"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/shubhamdhama/cockroach-mcp/pkg/clusterapi"
	"github.com/shubhamdhama/cockroach-mcp/pkg/db"
	"github.com/shubhamdhama/cockroach-mcp/pkg/server"
)

func main() {
	home, err := os.UserHomeDir()
	if err != nil {
		// Fallback method if os.UserHomeDir() fails
		log.Println("os.UserHomeDir() failed, trying another method...")

		usr, err := user.Current()
		if err != nil {
			log.Fatal("Failed to get user home directory:", err)
			return
		}
		home = usr.HomeDir
	}
	// Open a log file (create if it doesn't exist, append to it)
	logFilePath := filepath.Join(home, ".cockroach-mcp.log")
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("failed to open log file: %v", err)
	}

	multiWriter := io.MultiWriter(os.Stderr, logFile)
	log.SetOutput(multiWriter)

	envPath := filepath.Join(home, ".mcp-env")
	if err := godotenv.Load(envPath); err != nil {
		log.Println("Warning: No .env file found")
	}

	db.InitDB()
	clusterapi.InitAPIClient()

	log.Println("Starting cockroach-mcp server...")
	server.Start()
}
