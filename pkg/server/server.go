package server

import (
	"context"
	"log"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func Start() {
	s := server.NewMCPServer(
		"CockroachDB MCP Server",
		"0.1.0",
	)

	listTablesTool := mcp.NewTool(
		"list_tables",
		mcp.WithDescription("Fetches the list of tables from CockroachDB"),
	)
	s.AddTool(listTablesTool, handleListTables)

	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("Failed to start MCP server: %v", err)
	}
}

func handleListTables(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultText("Tables: [Placeholder]"), nil
}
