package server

import (
	"context"
	"log"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/shubhamdhama/cockroach-mcp/pkg/db"
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

func handleListTables(
	ctx context.Context, req mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	tables, err := db.ListTables()
	if err != nil {
		return mcp.NewToolResultError("Failed to fetch tables: " + err.Error()), nil
	}
	return mcp.NewToolResultText("Tables: " + tables), nil
}
