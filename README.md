# cockroach-mcp

This README was written in a hurry, but I hope it serves the purpose for now.

- **Build the binary:**
  `go build -o bin/cockroach-mcp cmd/cockroach-mcp/main.go`
- **Logs:** Logs are stored in `~/data/cockroach-mcp.log`
- **Environment:** Place your CockroachDB connection URI in `~/.mcp-env`

## How to use an MCP host

- You can use Claude Code, Claude Desktop or Cline (a VS Code extension).

### Configuring Cline

1. Install Cline.
2. Open the command palette and search for "MCP servers." Then, go to the
   Installed tab and select "Configure MCP servers."
3. Update the configuration to look like this:

   ```json
   {
     "mcpServers": {
       "cockroach-mcp": {
         "command": "/Users/<user name>/repos/work/cockroach-mcp/bin/cockroach-mcp",
         "autoApprove": []
       }
     }
   }
   ```

Cline should be able to connect to the MCP server, which will be running in
stdio mode—that is, Cline will start an MCP server process and communicate with
it over the stdio stream.

In Cline chat, configure your favorite LLM provider. You can start with "VS Code
LM API," which uses your GitHub Copilot models, and then query anything on Cline
chat related to your cluster.
