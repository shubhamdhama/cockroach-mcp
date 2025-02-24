package main

import (
	"log"

	"github.com/shubhamdhama/cockroach-mcp/pkg/server"
)

func main() {
	log.Println("Starting cockroach-mcp server...")
	server.Start()
}
