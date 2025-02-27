# cockroach-mcp

cockroach-mcp is an MCP server implementation written in Go for integrating with
CockroachDB. Using the Model Context Protocol (MCP) (see modelcontextprotocol.io
for details), this project exposes your CockroachDB’s schema and query
capabilities as tools that MCP hosts can consume.

## Prerequisites

- **Go** installed (Go 1.23+ recommended)
- A running CockroachDB instance
- Your CockroachDB connection URI saved in an environment file (see below)

## Installation & Build

1. **Clone the repository:**

   ```bash
   git clone https://github.com/shubhamdhama/cockroach-mcp.git
   cd cockroach-mcp
   ```

2. **Build the binary:**

   ```bash
   go build -o bin/cockroach-mcp cmd/cockroach-mcp/main.go
   ```

3. **Set up your environment:**

   Save your CockroachDB connection URI in `~/.mcp-env`. For example:

   ```bash
   echo "COCKROACH_URL=postgresl://demo@127.0.0.1:26257/movr" > ~/.mcp-env
   ```

## Usage Example

- You can use Claude Code, Claude Desktop, or Cline (a VS Code extension).

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

In Cline chat, configure your favorite LLM provider. You can start with "VS Code
LM API," which uses your GitHub Copilot models, and then query anything on Cline
chat related to your cluster.

## Using MCP inspector

```
npx @modelcontextprotocol/inspector $HOME/repos/work/cockroach-mcp/bin/cockroach-mcp
```

---

Logs will be written to `~/.cockroach-mcp.log` for troubleshooting.

---

Feel free to reach out or contribute if you have suggestions for improvement!
