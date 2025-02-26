package server

import (
	"context"
	"fmt"
	"log"
	"strings"

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

	runSQLTool := mcp.NewTool("run_sql",
		mcp.WithDescription("Execute a single SQL statement against the database"),
		mcp.WithString("sql",
			mcp.Required(),
			mcp.Description("The SQL query to execute"),
		),
		mcp.WithString("mode",
			mcp.Description("Use 'execute' for non-returning commands, otherwise 'query' (default)."),
		),
	)
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
	mode, ok := req.Params.Arguments["mode"].(string)
	var result string
	var err error
	if !ok || strings.ToLower(mode) != "execute" {
		result, err = db.Query(ctx, query)
	} else {
		result, err = db.Execute(ctx, query)
	}
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to run SQL: %v", err)), nil
	}
	return mcp.NewToolResultText(fmt.Sprintf("Output: \n%s", result)), nil
}
