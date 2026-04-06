# ynab-cli

An agent-friendly command-line interface for the [YNAB (You Need A Budget) API](https://api.ynab.com/). Designed for both human and AI agent use, with structured JSON output, minimal dependencies, and full API coverage.

## Features

- **Full YNAB API v1 coverage** — plans, accounts, categories, transactions, payees, scheduled transactions, months, money movements
- **Minimal dependencies** — only [cobra](https://github.com/spf13/cobra) for CLI parsing; the API client and output layer use only Go's standard library
- **Agent-friendly** — JSON output by default, consistent error codes, structured responses
- **Multiple output formats** — JSON (default), table, CSV
- **Cross-platform** — pre-built binaries for Linux, macOS, and Windows (amd64/arm64)
- **Delta sync support** — incremental data fetching via `--last-knowledge` to minimize API calls
- **MCP server mode** — built-in [Model Context Protocol](https://modelcontextprotocol.io) server for AI agent integration via stdio or HTTP

## Security

This tool handles financial data. We take a security-conscious approach:

- **Minimal dependencies** — only cobra/pflag for CLI parsing; the API client uses only Go's standard library
- **Tokens never logged** — access tokens are only sent in HTTP Authorization headers
- **Config file permissions** — stored with `0600` permissions in `~/.config/ynab-cli/`
- **Auditable** — small, readable codebase you can verify yourself

## Installation

### Pre-built binaries

Download the latest release from [GitHub Releases](https://github.com/bwishan/ynab-cli/releases):

```bash
# Linux (amd64)
curl -L https://github.com/bwishan/ynab-cli/releases/latest/download/ynab-linux-amd64 -o ynab
chmod +x ynab
sudo mv ynab /usr/local/bin/

# macOS (Apple Silicon)
curl -L https://github.com/bwishan/ynab-cli/releases/latest/download/ynab-darwin-arm64 -o ynab
chmod +x ynab
sudo mv ynab /usr/local/bin/

# macOS (Intel)
curl -L https://github.com/bwishan/ynab-cli/releases/latest/download/ynab-darwin-amd64 -o ynab
chmod +x ynab
sudo mv ynab /usr/local/bin/
```

### From source

Requires Go 1.22+:

```bash
go install github.com/bwishan/ynab-cli/cmd/ynab@latest
```

Or clone and build:

```bash
git clone https://github.com/bwishan/ynab-cli.git
cd ynab-cli
make build
./bin/ynab --version
```

## Configuration

### 1. Get an access token

Create a Personal Access Token at [YNAB Developer Settings](https://app.ynab.com/settings/developer).

### 2. Configure the CLI

```bash
# Option A: Save to config file
ynab configure --token <your-access-token>

# Option B: Environment variable
export YNAB_ACCESS_TOKEN=<your-access-token>

# Option C: Pass per-command
ynab --token <your-access-token> plans list
```

### 3. Set a default plan (optional)

```bash
# Find your plan ID
ynab plans list

# Set it as default so you don't need to pass it every time
ynab configure --default-plan <plan-id>
```

## Usage

```
ynab [global-flags] <command> <subcommand> [arguments]
```

### Global flags

| Flag | Description |
|------|-------------|
| `--token <token>` | YNAB access token (overrides env/config) |
| `--output <format>` | Output format: `json` (default), `table`, `csv` |
| `--pretty` | Pretty-print JSON output |
| `--version` | Print version |
| `--help` | Print help |

### Commands

| Command | Description |
|---------|-------------|
| `user` | Authenticated user info |
| `plans` | Budget plans |
| `accounts` | Bank accounts |
| `categories` | Budget categories |
| `category-groups` | Category groups |
| `transactions` | Transactions |
| `payees` | Payees |
| `payee-locations` | Payee locations |
| `scheduled` | Scheduled transactions |
| `months` | Budget months |
| `money-movements` | Money movements |
| `money-movement-groups` | Money movement groups |
| `configure` | Save CLI configuration |

### Examples

```bash
# List all plans
ynab plans list --pretty

# List accounts in a plan (table format)
ynab accounts list <plan-id> --output table

# Get all transactions since a date
ynab transactions list <plan-id> --since-date 2024-01-01

# Create a transaction (amounts in milliunits: $50.00 = 50000)
ynab transactions create <plan-id> \
  --account-id <id> \
  --date 2024-03-15 \
  --amount -50000 \
  --payee-name "Coffee Shop" \
  --category-id <id> \
  --memo "Morning coffee"

# Get uncategorized transactions
ynab transactions list <plan-id> --type uncategorized

# List categories with budgets
ynab categories list <plan-id> --output table

# Update a category budget for a month (assign $500)
ynab categories update-month <plan-id> 2024-03-01 <category-id> --budgeted 500000

# Delta sync (only get changes since last fetch)
ynab transactions list <plan-id> --last-knowledge 12345

# Import linked account transactions
ynab transactions import <plan-id>

# Create a scheduled transaction
ynab scheduled create <plan-id> \
  --account-id <id> \
  --date 2024-04-01 \
  --amount -100000 \
  --frequency monthly \
  --payee-id <payee-id> \
  --category-id <id>
```

## Environment Variables

| Variable | Description |
|----------|-------------|
| `YNAB_ACCESS_TOKEN` | YNAB API access token |
| `YNAB_PLAN_ID` | Default plan ID |

## YNAB API Concepts

### Milliunits

YNAB uses "milliunits" for currency amounts. 1,000 milliunits = 1 currency unit.

| Display | Milliunits |
|---------|------------|
| $1.00 | 1000 |
| $50.00 | 50000 |
| -$123.45 | -123450 |

### Delta requests

Many endpoints support incremental fetching via `--last-knowledge`. The API response includes a `server_knowledge` value. Pass it on subsequent requests to receive only data changed since that point.

```bash
# First request - gets everything
ynab transactions list <plan-id> --pretty
# Response includes: "server_knowledge": 500

# Later request - gets only changes
ynab transactions list <plan-id> --last-knowledge 500
```

### Rate Limiting

The YNAB API allows 200 requests per hour per access token. The CLI returns a clear error when the limit is hit (HTTP 429).

## MCP Server Mode

The CLI includes a built-in [Model Context Protocol](https://modelcontextprotocol.io) server, exposing two tools (`search_commands` and `run_command`) that let AI agents discover and execute any CLI command programmatically.

```bash
# stdio transport (default) — for Claude Code, Cursor, etc.
ynab mcp serve

# HTTP transport — for remote or multi-client setups
ynab mcp serve --transport http --port 8080
```

### Claude Code configuration

Add to your MCP settings (`.claude/settings.json` or project config):

```json
{
  "mcpServers": {
    "ynab": {
      "command": "ynab",
      "args": ["mcp", "serve"],
      "env": {
        "YNAB_ACCESS_TOKEN": "<your-token>"
      }
    }
  }
}
```

The agent can then call `search_commands` to discover available commands and `run_command` to execute them — no shell access needed.

## Agent Integration

See **[INTEGRATION.md](INTEGRATION.md)** for detailed instructions on integrating this CLI with AI agents like Claude Code, OpenClaw, and others — including install scripts, MCP server config, OpenClaw tool definitions, and common agent workflows.

## Development

```bash
# Build
make build

# Run tests
make test

# Lint
make lint

# Cross-compile all platforms
make build-all
```

## Releasing

Push a version tag to trigger a GitHub Actions release:

```bash
git tag v1.0.0
git push origin v1.0.0
```

This builds binaries for all platforms and creates a GitHub Release with checksums.

## License

MIT — see [LICENSE](LICENSE).
