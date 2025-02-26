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

	// TODO(shubham): improve descriptions

	s.AddTool(
		mcp.NewTool("list_databases",
			mcp.WithDescription("Fetches the list of databases"),
		), handleListDatabases)

	s.AddTool(
		mcp.NewTool("list_tables",
			mcp.WithDescription("Fetches the list of tables"),
			mcp.WithString("database",
				mcp.Description("Specify the database to list its tables (optional)"),
			),
		), handleListTables)

	s.AddTool(
		mcp.NewTool("list_crdb_internal_tables",
			mcp.WithDescription("Fetches the list of crdb_internal tables along with their description"),
		), handleListCRDBInternalTables)

	s.AddTool(
		mcp.NewTool("list_system_tables",
			mcp.WithDescription("Fetches the list of system database tables along with estimated row count. These tables store metadata for cluster operation."),
		), handleListSystemTables)

	s.AddTool(
		mcp.NewTool("list_cluster_settings",
			mcp.WithDescription("Fetches the list of cluster settings along with their description, default, and currently set values"),
		), handleListClusterSettings)

	s.AddTool(
		mcp.NewTool("run_sql",
			mcp.WithDescription("Execute a single SQL statement against the database"),
			mcp.WithString("sql", mcp.Required(),
				mcp.Description("The SQL query to execute"),
			),
			mcp.WithString("mode",
				mcp.Description("Use 'execute' for non-returning commands, otherwise 'query' (default)."),
			),
		), handleRunSQL)

	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("Failed to start MCP server: %v", err)
	}
}

func handleListDatabases(
	ctx context.Context, _ mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	databases, err := db.ListDatabases(ctx)
	if err != nil {
		return mcp.NewToolResultError("Failed to fetch databases: " + err.Error()), nil
	}
	return mcp.NewToolResultText(databases), nil
}

func handleListTables(
	ctx context.Context, req mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	database, ok := req.Params.Arguments["database"].(string)
	if !ok {
		database = ""
	}
	tables, err := db.ListTables(ctx, database)
	if err != nil {
		return mcp.NewToolResultError("Failed to fetch tables: " + err.Error()), nil
	}
	return mcp.NewToolResultText(tables), nil
}

func handleListCRDBInternalTables(
	ctx context.Context, _ mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	tables, err := db.ListCRDBInternalTables(ctx)
	if err != nil {
		return mcp.NewToolResultError("Failed to fetch list of crdb_internal tables: " + err.Error()), nil
	}
	return mcp.NewToolResultText(tables), nil
}

func handleListSystemTables(
	ctx context.Context, _ mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	tables, err := db.ListSystemTables(ctx)
	if err != nil {
		return mcp.NewToolResultError("Failed to fetch list of system database tables: " + err.Error()), nil
	}
	return mcp.NewToolResultText(tables), nil
}

func handleListClusterSettings(
	ctx context.Context, _ mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	settings, err := db.ListClusterSettings(ctx)
	if err != nil {
		return mcp.NewToolResultError("Failed to fetch all cluster settings: " + err.Error()), nil
	}
	return mcp.NewToolResultText(settings), nil
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
	return mcp.NewToolResultText(result), nil
}
