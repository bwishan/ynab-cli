# Agent Integration Guide

This document describes how to install, configure, and use `ynab-cli` from AI agents such as Claude Code, OpenClaw, ChatGPT, and other LLM-powered tools.

## Quick Start for Agents

```bash
# 1. Install
curl -L https://github.com/bwishan/ynab-cli/releases/latest/download/ynab-linux-amd64 -o /usr/local/bin/ynab && chmod +x /usr/local/bin/ynab

# 2. Configure (token provided by user)
export YNAB_ACCESS_TOKEN="<token>"

# 3. Verify
ynab user get
```

## Why This Tool is Agent-Friendly

- **JSON by default** — all output is structured JSON, parseable without regex
- **Consistent interface** — every command follows `ynab <resource> <action> [args] [flags]`
- **Deterministic errors** — errors are JSON with `id`, `name`, and `detail` fields
- **No interactive prompts** — all input via flags and arguments
- **Exit codes** — 0 for success, 1 for errors
- **Delta sync** — use `--last-knowledge` to avoid re-fetching entire datasets

## Installation by Platform

### Linux (amd64)
```bash
curl -L https://github.com/bwishan/ynab-cli/releases/latest/download/ynab-linux-amd64 -o ynab
chmod +x ynab
sudo mv ynab /usr/local/bin/
```

### Linux (arm64)
```bash
curl -L https://github.com/bwishan/ynab-cli/releases/latest/download/ynab-linux-arm64 -o ynab
chmod +x ynab
sudo mv ynab /usr/local/bin/
```

### macOS (Apple Silicon)
```bash
curl -L https://github.com/bwishan/ynab-cli/releases/latest/download/ynab-darwin-arm64 -o ynab
chmod +x ynab
sudo mv ynab /usr/local/bin/
```

### macOS (Intel)
```bash
curl -L https://github.com/bwishan/ynab-cli/releases/latest/download/ynab-darwin-amd64 -o ynab
chmod +x ynab
sudo mv ynab /usr/local/bin/
```

### From Source (Go 1.22+)
```bash
go install github.com/bwishan/ynab-cli/cmd/ynab@latest
```

## Authentication Setup

The user must provide a YNAB Personal Access Token. Tokens are created at:
https://app.ynab.com/settings/developer

### Configuration Priority

1. `--token <token>` flag (highest priority)
2. `YNAB_ACCESS_TOKEN` environment variable
3. `~/.config/ynab-cli/config.json` file

### Recommended Agent Setup

```bash
# Set via environment variable (preferred for agents)
export YNAB_ACCESS_TOKEN="<user-provided-token>"

# Optionally set a default plan to avoid passing plan-id every time
export YNAB_PLAN_ID="<plan-id>"
```

## Command Reference

### Discovering Plans and Accounts

```bash
# List all plans (budgets)
ynab plans list

# List accounts in a plan
ynab accounts list <plan-id>

# Get plan settings (currency, date format)
ynab plans settings <plan-id>
```

### Reading Financial Data

```bash
# All transactions (most recent first)
ynab transactions list <plan-id>

# Transactions since a date
ynab transactions list <plan-id> --since-date 2024-01-01

# Transactions by account/category/payee
ynab transactions list-by-account <plan-id> <account-id>
ynab transactions list-by-category <plan-id> <category-id>
ynab transactions list-by-payee <plan-id> <payee-id>

# Budget categories with current month data
ynab categories list <plan-id>

# Specific month's budget
ynab months get <plan-id> 2024-03-01

# Category budget for a specific month
ynab categories get-month <plan-id> 2024-03-01 <category-id>
```

### Making Changes

```bash
# Create a transaction (amounts in milliunits: $50 = 50000)
ynab transactions create <plan-id> \
  --account-id <id> --date 2024-03-15 --amount -50000 \
  --payee-name "Store" --category-id <id>

# Update a transaction
ynab transactions update <plan-id> <transaction-id> --memo "Updated memo"

# Set a category budget for a month
ynab categories update-month <plan-id> 2024-03-01 <category-id> --budgeted 500000

# Create a payee
ynab payees create <plan-id> --name "New Payee"

# Delete a transaction
ynab transactions delete <plan-id> <transaction-id>
```

### Efficient Data Fetching (Delta Sync)

```bash
# Initial fetch (note the server_knowledge value in response)
ynab transactions list <plan-id>
# Response: { "data": { "transactions": [...], "server_knowledge": 500 } }

# Subsequent fetch (only changes since knowledge=500)
ynab transactions list <plan-id> --last-knowledge 500
```

Delta sync is supported by: `accounts list`, `categories list`, `transactions list`, `payees list`, `scheduled list`, `months list`.

## Claude Code / OpenClaw Integration

### As a Shell Tool

The simplest integration is to invoke `ynab` commands directly via shell execution:

```
User: "What's my current budget status?"
Agent: Runs `ynab plans list` → picks plan → runs `ynab categories list <plan-id> --pretty`
Agent: Summarizes category budgets, spending, and available amounts
```

### MCP Server Wrapper (Advanced)

For tighter integration, you can wrap the CLI as an MCP (Model Context Protocol) server. Example `mcp.json` configuration:

```json
{
  "mcpServers": {
    "ynab": {
      "command": "ynab",
      "args": ["--output", "json"],
      "env": {
        "YNAB_ACCESS_TOKEN": "<token>"
      }
    }
  }
}
```

### OpenClaw Tool Definition

```yaml
name: ynab
description: "CLI tool for managing YNAB budgets. Supports reading and writing budget data including plans, accounts, categories, transactions, payees, and scheduled transactions."
install:
  command: |
    curl -L https://github.com/bwishan/ynab-cli/releases/latest/download/ynab-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/') -o /usr/local/bin/ynab && chmod +x /usr/local/bin/ynab
  verify: ynab --version
env:
  - name: YNAB_ACCESS_TOKEN
    description: "YNAB Personal Access Token from https://app.ynab.com/settings/developer"
    required: true
  - name: YNAB_PLAN_ID
    description: "Default YNAB plan/budget ID (optional, avoids passing plan-id to every command)"
    required: false
commands:
  - name: list-plans
    command: ynab plans list
    description: "List all YNAB budget plans"
  - name: list-accounts
    command: "ynab accounts list {plan_id}"
    description: "List all accounts in a plan"
    params:
      - name: plan_id
        description: "Plan/budget ID"
        required: true
  - name: list-transactions
    command: "ynab transactions list {plan_id} --since-date {since_date}"
    description: "List transactions, optionally filtered by date"
    params:
      - name: plan_id
        required: true
      - name: since_date
        required: false
  - name: list-categories
    command: "ynab categories list {plan_id}"
    description: "List all budget categories with current amounts"
    params:
      - name: plan_id
        required: true
  - name: create-transaction
    command: "ynab transactions create {plan_id} --account-id {account_id} --date {date} --amount {amount} --payee-name {payee_name} --category-id {category_id} --memo {memo}"
    description: "Create a new transaction (amount in milliunits)"
    params:
      - name: plan_id
        required: true
      - name: account_id
        required: true
      - name: date
        description: "ISO date (YYYY-MM-DD)"
        required: true
      - name: amount
        description: "Amount in milliunits (1000 = $1.00, negative for outflow)"
        required: true
      - name: payee_name
        required: false
      - name: category_id
        required: false
      - name: memo
        required: false
```

## Common Agent Workflows

### Budget Analysis
```bash
ynab plans list                                    # Find the plan
ynab categories list <plan-id> --pretty            # Get all categories with budgets
ynab transactions list <plan-id> --since-date $(date -d '30 days ago' +%Y-%m-%d)  # Recent spending
```

### Spending Report by Category
```bash
ynab transactions list-by-category <plan-id> <category-id> --since-date 2024-01-01
```

### Monthly Budget Review
```bash
ynab months get <plan-id> 2024-03-01 --pretty     # Month summary
ynab categories list <plan-id> --pretty            # Category details
```

### Account Reconciliation
```bash
ynab accounts list <plan-id> --output table        # See all balances
ynab transactions list-by-account <plan-id> <account-id> --since-date 2024-03-01
```

## Amount Format (Milliunits)

All amounts in the YNAB API use "milliunits" (1/1000 of a currency unit):

| Display Amount | Milliunits Value |
|---------------|-----------------|
| $1.00 | 1000 |
| $50.00 | 50000 |
| $1,234.56 | 1234560 |
| -$25.00 | -25000 |

When creating/updating transactions, always use milliunits for the `--amount` flag.

## Error Handling

Errors are returned as JSON to stderr with a non-zero exit code:

```json
{
  "error": {
    "id": "401",
    "name": "unauthorized",
    "detail": "The access token is invalid or has expired."
  }
}
```

Common error codes:
- **401** — Invalid or expired token
- **404** — Resource not found (bad plan-id, transaction-id, etc.)
- **429** — Rate limited (200 requests/hour/token). Wait and retry.
- **400** — Invalid request parameters

## Rate Limits

The YNAB API allows **200 requests per hour** per access token. To stay within limits:

1. Use delta sync (`--last-knowledge`) to fetch only changes
2. Cache plan IDs and account IDs between calls
3. Use `YNAB_PLAN_ID` to avoid extra `plans list` calls
4. Batch reads when possible (e.g., `categories list` returns all categories at once)
