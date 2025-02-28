package server

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/shubhamdhama/cockroach-mcp/pkg/clusterapi"
	"github.com/shubhamdhama/cockroach-mcp/pkg/db"
)

func Start() {
	s := server.NewMCPServer(
		"CockroachDB MCP Server",
		"0.1.0",
		server.WithLogging(),
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
			mcp.WithString("sql", mcp.Required(), mcp.Description("The SQL query to execute")),
			mcp.WithString("mode", mcp.Description("Use 'execute' for non-returning commands, otherwise 'query' (default).")),
		), handleRunSQL)

	s.AddTool(
		mcp.NewTool("query_timeseries_metrics",
			mcp.WithDescription("Fetch metrics related to timeseries data (e.g., CPU usage, etc.)"),
			mcp.WithString("tenant", mcp.Required(), mcp.Description("Tenant name")),
			mcp.WithString("start_time", mcp.Required(), mcp.Description("Start time in 'YYYY-MM-DD HH:MM:SS' format")),
			mcp.WithString("end_time", mcp.Required(), mcp.Description("End time in 'YYYY-MM-DD HH:MM:SS' format")),
			mcp.WithString("query", mcp.Required(), mcp.Description("Query name for timeseries data")),
		), handleTSQueryAPI)

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

var DefaultTenant = "system"

func handleTSQueryAPI(
	ctx context.Context, req mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	tenant, ok := req.Params.Arguments["tenant"].(string)
	if !ok || tenant == "" {
		tenant = DefaultTenant
	}
	startStr, ok := req.Params.Arguments["start_time"].(string)
	if !ok || startStr == "" {
		return mcp.NewToolResultError("start must be provided as a string in the format 'YYYY-MM-dd HH-mm-ss'"), nil
	}
	endStr, ok := req.Params.Arguments["end_time"].(string)
	if !ok || endStr == "" {
		return mcp.NewToolResultError("end must be provided as a string in the format 'YYYY-MM-dd HH-mm-ss'"), nil
	}

	layout := time.DateTime
	startTime, err := time.Parse(layout, startStr)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to parse start: %v", err)), nil
	}
	endTime, err := time.Parse(layout, endStr)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to parse end: %v", err)), nil
	}

	startNanos := startTime.UnixNano()
	endNanos := endTime.UnixNano()

	query, ok := req.Params.Arguments["query"].(string)
	if !ok || query == "" {
		return mcp.NewToolResultError("Missing or invalid SQL query"), nil
	}
	response, err := clusterapi.GetClient().QueryTimeseries(ctx, tenant, startNanos, endNanos, query)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("query failed: %v", err)), nil
	}
	return mcp.NewToolResultText(response), nil
}
