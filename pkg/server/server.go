package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/shubhamdhama/cockroach-mcp/pkg/db"
	"github.com/shubhamdhama/cockroach-mcp/pkg/tsdb"
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
			mcp.WithBoolean("using_system_tenant",
				mcp.Description("should the request be made using the system tenant (optional)"),
			),
		), handleListDatabases)

	s.AddTool(
		mcp.NewTool("list_tables",
			mcp.WithDescription("Fetches the list of tables"),
			mcp.WithBoolean("using_system_tenant",
				mcp.Description("should the request be made using the system tenant (optional)"),
			),
			mcp.WithString("database",
				mcp.Description("Specify the database to list its tables (optional)"),
			),
		), handleListTables)

	s.AddTool(
		mcp.NewTool("list_crdb_internal_tables",
			mcp.WithDescription("Fetches the list of crdb_internal tables along with their description"),
			mcp.WithBoolean("using_system_tenant",
				mcp.Description("should the request be made using the system tenant (optional)"),
			),
		), handleListCRDBInternalTables)

	s.AddTool(
		mcp.NewTool("list_system_tables",
			mcp.WithDescription("Fetches the list of system database tables along with estimated row count. These tables store metadata for cluster operation."),
			mcp.WithBoolean("using_system_tenant",
				mcp.Description("should the request be made using the system tenant (optional)"),
			),
		), handleListSystemTables)

	s.AddTool(
		mcp.NewTool("list_cluster_settings",
			mcp.WithDescription("Fetches the list of cluster settings along with their description, default, and currently set values"),
		), handleListClusterSettings)

	s.AddTool(
		mcp.NewTool("run_sql",
			mcp.WithDescription("Execute a single SQL statement against the database"),
			mcp.WithBoolean("use_system_tenant", mcp.Required(),
				mcp.Description("should the request be made using the system tenant"),
			),
			mcp.WithString("sql", mcp.Required(),
				mcp.Description("The SQL query to execute"),
			),
			mcp.WithString("mode",
				mcp.Description("Use 'execute' for non-returning commands, otherwise 'query' (default)."),
			),
		), handleRunSQL)

	s.AddTool(
		mcp.NewTool("tsdb_query",
			mcp.WithDescription("Query timeseries data from TSDB"),
			mcp.WithNumber("start_nanos",
				mcp.Required(),
				mcp.Description(`A timestamp in nanoseconds which defines the
					early bound of the time span for this query.`),
			),
			mcp.WithNumber("end_nanos", mcp.Required(),
				mcp.Required(),
				mcp.Description(`A timestamp in nanoseconds which defines the
					late bound of the time span for this query. Must be greater
					than start_nanos.`),
			),
			mcp.WithNumber("sample_nanos", mcp.Required(),
				mcp.Description(`Duration of requested sample period in
					nanoseconds. Returned data for each query will be
					downsampled into periods of the supplied length. The
					supplied duration must be a multiple of ten seconds.`),
				mcp.DefaultNumber(float64((time.Second*10).Nanoseconds())),
			),
			mcp.WithString("sources", mcp.Required(),
				mcp.Description(`Timeseries sources to query. A request must
					have at least one timeseries source.`),
			),
		), handleTimeseriesQuery)

	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("Failed to start MCP server: %v", err)
	}
}

func handleListDatabases(
	ctx context.Context, req mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	st, _ := req.Params.Arguments["use_system_tenant"].(bool)
	databases, err := db.ListDatabases(ctx, st)
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
	st, _ := req.Params.Arguments["use_system_tenant"].(bool)
	tables, err := db.ListTables(ctx, st, database)
	if err != nil {
		return mcp.NewToolResultError("Failed to fetch tables: " + err.Error()), nil
	}
	return mcp.NewToolResultText(tables), nil
}

func handleListCRDBInternalTables(
	ctx context.Context, req mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	st, _ := req.Params.Arguments["use_system_tenant"].(bool)
	tables, err := db.ListCRDBInternalTables(ctx, st)
	if err != nil {
		return mcp.NewToolResultError("Failed to fetch list of crdb_internal tables: " + err.Error()), nil
	}
	return mcp.NewToolResultText(tables), nil
}

func handleListSystemTables(
	ctx context.Context, req mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	st, _ := req.Params.Arguments["use_system_tenant"].(bool)
	tables, err := db.ListSystemTables(ctx, st)
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
	st, _ := req.Params.Arguments["use_system_tenant"].(bool)
	mode, ok := req.Params.Arguments["mode"].(string)
	var result string
	var err error
	if !ok || strings.ToLower(mode) != "execute" {
		result, err = db.Query(ctx, st, query)
	} else {
		result, err = db.Execute(ctx, st, query)
	}
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to run SQL: %v", err)), nil
	}
	return mcp.NewToolResultText(result), nil
}

func handleTimeseriesQuery(
	ctx context.Context, req mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	if req.Params.Arguments["start_nanos"] == nil {
		return mcp.NewToolResultError("start time is required"), nil
	}
	if req.Params.Arguments["end_nanos"] == nil {
		return mcp.NewToolResultError("end time is required"), nil
	}
	if req.Params.Arguments["sample_nanos"] == nil {
		return mcp.NewToolResultError("sample time is required"), nil
	}

	tsdbReq := tsdb.TSQueryRequest{}
	tsdbReq.StartNanos = req.Params.Arguments["start_nanos"].(int64)
	tsdbReq.EndNanos = req.Params.Arguments["end_nanos"].(int64)
	tsdbReq.SampleNanos = req.Params.Arguments["sample_nanos"].(int64)
	if tsdbReq.StartNanos >= tsdbReq.EndNanos {
		return mcp.NewToolResultError("start time must be before end time"), nil
	}
	if req.Params.Arguments["sources"] == nil || req.Params.Arguments["sources"].(string) == "" {
		return mcp.NewToolResultError("timeseries sources are required"), nil
	}

	strings.Split(req.Params.Arguments["sources"].(string), ",")
	for _, source := range strings.Split(req.Params.Arguments["sources"].(string), ",") {
		tsdbReq.Queries = append(tsdbReq.Queries, tsdb.Query{
			Name:              source,
			DownSampler:       tsdb.QueryAggSum,
			SourceAggeregator: tsdb.QueryAggSum,
			Derivative:        tsdb.None,
			Sources:           []string{source},
		})
	}

	tsdbRes, err := tsdb.Client().Query(ctx, &tsdbReq)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to query tsdb: %v", err)), nil
	}
	resp, err := json.Marshal(tsdbRes)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to query tsdb: %v", err)), nil
	}

	return mcp.NewToolResultText(string(resp)), nil
}
