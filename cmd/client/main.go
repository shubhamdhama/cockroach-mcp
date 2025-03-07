package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

func main() {
	mcpClient, err := client.NewStdioMCPClient("go", nil, "run", "cmd/cockroach-mcp/main.go")
	if err != nil {
		log.Fatalf("Failed to create MCP client: %v", err)
	}
	defer mcpClient.Close()

	ctx := context.Background()
	log.Println("Initializing client...")

	initRequest := mcp.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initRequest.Params.ClientInfo = mcp.Implementation{
		Name:    "test-client",
		Version: "1.0.0",
	}
	initResult, err := mcpClient.Initialize(ctx, initRequest)
	if err != nil {
		log.Fatalf("Failed to initialize MCP client: %v", err)
	}

	log.Printf("Initialized with server: %s(%s)",
		initResult.ServerInfo.Name, initResult.ServerInfo.Version)
	if err := mcpClient.Ping(ctx); err != nil {
		log.Fatalf("MCP server not responding: %v", err)
	}
	fmt.Println("Connected to MCP server")

	// req := mcp.CallToolRequest{}

	// req.Params.Name = "list_tables"

	// result, err := mcpClient.CallTool(ctx, req)
	// if err != nil {
	// 	log.Fatalf("Tool call failed: %v", err)
	// }
	// fmt.Println("Response from list_tables: ", result.Content)

	// Call TSDB query tool
	{
		req := mcp.CallToolRequest{}
		req.Params.Name = "tsdb_query"
		req.Params.Arguments = map[string]interface{}{
			"start":  time.Now().Add(-time.Minute * 2).UnixNano(),
			"end":    time.Now().UnixNano(),
			"sample": (time.Second * 10).Nanoseconds(),
			"queries": []interface{}{
				map[string]interface{}{
					"name":              "cpu",
					"downsampler":       "sum",
					"source_aggregator": "sum",
					"derivative":        "none",
					"sources":           []string{"cpu"},
				},
			},
		}

		resp, err := mcpClient.CallTool(ctx, req)
		if err != nil {
			log.Fatalf("Tool call failed: %v", err)
		}
		fmt.Println("Response from list_tables: ", resp.Content)
	}
}
