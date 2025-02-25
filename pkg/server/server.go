package server

import (
	"context"
	"fmt"
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

	runSQLTool := mcp.NewTool("run_sql",
		mcp.WithDescription("Execute a single SQL statement against the database"),
		mcp.WithString("sql",
			mcp.Required(),
			mcp.Description("The SQL query to execute"),
		),
	)
	s.AddTool(listTablesTool, handleListTables)
	s.AddTool(runSQLTool, handleRunSQL)

	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("Failed to start MCP server: %v", err)
	}
}

func handleListTables(
	ctx context.Context, req mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	log.Printf("handleListTables: %v", req)
	tables, err := db.ListTables()
	if err != nil {
		log.Printf("Failed to fetch tables: %v", err)
		return mcp.NewToolResultError("Failed to fetch tables: " + err.Error()), nil
	}
	log.Printf("Fetched tables: %v", tables)
	return mcp.NewToolResultText("Tables: " + tables), nil
}

func handleRunSQL(
	ctx context.Context, req mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	log.Printf("handleRunSQL: %v", req)
	query, ok := req.Params.Arguments["sql"].(string)
	if !ok || query == "" {
		return mcp.NewToolResultError("Missing or invalid SQL query"), nil
	}
	result, err := db.RunSQL(ctx, query)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to run SQL: %v", err)), nil
	}
	return mcp.NewToolResultText(fmt.Sprintf("Result:\n%s", result)), nil
}
