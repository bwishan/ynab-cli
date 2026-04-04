---
name: ynab-cli
description: Agent-friendly CLI for the YNAB (You Need A Budget) API. Use this skill whenever the user mentions YNAB, budgeting, personal finance, budget tracking, spending analysis, transaction management, or wants to interact with their YNAB account programmatically. Also trigger when the user asks about budget categories, payees, scheduled transactions, account balances, monthly budgets, or any financial data that lives in YNAB — even if they don't explicitly say "YNAB" but describe a budgeting workflow that matches what YNAB does.
---

# ynab-cli

A CLI tool that wraps the full YNAB API v1. Every command outputs structured JSON by default, uses no interactive prompts, and returns proper exit codes — making it ideal for agent automation. The binary is self-contained with no runtime dependencies.

## Quick Start

```bash
# Install (Linux amd64 — see references/commands.md for other platforms)
curl -L https://github.com/bwishan/ynab-cli/releases/latest/download/ynab-linux-amd64 -o /usr/local/bin/ynab && chmod +x /usr/local/bin/ynab

# Authenticate (user must provide their YNAB Personal Access Token)
export YNAB_ACCESS_TOKEN="<token>"

# Verify
ynab user get
```

Tokens are created at https://app.ynab.com/settings/developer. The user must provide one — never generate or guess tokens.

## Key Concepts

### Command Structure

Every command follows: `ynab <resource> <action> [positional-args] [--flags]`

Resources: `user`, `plans`, `accounts`, `categories`, `category-groups`, `transactions`, `payees`, `payee-locations`, `scheduled`, `months`, `money-movements`, `money-movement-groups`, `configure`

### Milliunits

YNAB represents all currency amounts in "milliunits" — 1/1000 of a currency unit. This avoids floating-point issues. Always convert when displaying to users or accepting input:

| User sees | CLI expects |
|-----------|-------------|
| $1.00     | 1000        |
| $50.00    | 50000       |
| -$123.45  | -123450     |

Negative amounts = outflows (spending). Positive = inflows (income).

### Plan IDs

Most commands require a `plan-id` (YNAB calls budgets "plans"). Discover it with `ynab plans list`, then either pass it to every command or set a default:

```bash
export YNAB_PLAN_ID="<plan-id>"
# or
ynab configure --default-plan <plan-id>
```

### Delta Sync

Many endpoints support incremental fetching via `--last-knowledge <number>`. The API response includes a `server_knowledge` value — pass it on the next request to get only what changed. This is critical for staying within the 200 requests/hour rate limit.

```bash
# First call returns server_knowledge: 500
ynab transactions list <plan-id>
# Next call gets only changes
ynab transactions list <plan-id> --last-knowledge 500
```

Supported by: `accounts list`, `categories list`, `transactions list`, `payees list`, `scheduled list`, `months list`.

## Configuration Priority

Tokens and plan IDs resolve in this order (highest priority first):

1. CLI flag (`--token`, positional `<plan-id>`)
2. Environment variable (`YNAB_ACCESS_TOKEN`, `YNAB_PLAN_ID`)
3. Config file (`~/.config/ynab-cli/config.json`)

For agents, environment variables are the recommended approach.

## Global Flags

| Flag | Description |
|------|-------------|
| `--token <token>` | Override access token |
| `--output <format>` | `json` (default), `table`, `csv` |
| `--pretty` | Pretty-print JSON |
| `--version` | Print version |
| `--help` | Print help |

## Common Workflows

### Budget Overview

```bash
ynab plans list                                     # Find the plan
ynab categories list <plan-id> --pretty             # All categories with budgeted/spent/available
ynab months get <plan-id> 2024-03-01 --pretty       # Specific month summary
```

### Transaction Management

```bash
# List recent transactions
ynab transactions list <plan-id> --since-date 2024-01-01

# Create a transaction (outflow of $50)
ynab transactions create <plan-id> \
  --account-id <id> --date 2024-03-15 --amount -50000 \
  --payee-name "Coffee Shop" --category-id <id> --memo "Morning coffee"

# Filter by account, category, or payee
ynab transactions list-by-account <plan-id> <account-id>
ynab transactions list-by-category <plan-id> <category-id>
ynab transactions list-by-payee <plan-id> <payee-id>

# Get uncategorized transactions
ynab transactions list <plan-id> --type uncategorized
```

### Budget Adjustments

```bash
# Set a category budget for a month ($500 = 500000 milliunits)
ynab categories update-month <plan-id> 2024-03-01 <category-id> --budgeted 500000
```

### Scheduled Transactions

```bash
ynab scheduled create <plan-id> \
  --account-id <id> --date 2024-04-01 --amount -100000 \
  --frequency monthly --payee-name "Electric Company" --category-id <id>
```

Valid frequencies: `never`, `daily`, `weekly`, `everyOtherWeek`, `twiceAMonth`, `every4Weeks`, `monthly`, `everyOtherMonth`, `every3Months`, `every4Months`, `twiceAYear`, `yearly`, `everyOtherYear`.

## Error Handling

Errors are JSON on stderr with exit code 1:

```json
{
  "error": {
    "id": "401",
    "name": "unauthorized",
    "detail": "The access token is invalid or has expired."
  }
}
```

Key errors: **401** (bad token), **404** (bad ID), **429** (rate limited — 200 req/hr), **400** (bad params).

When hitting 429, wait before retrying. Use delta sync and cache IDs to minimize API calls.

## Output Formats

- **JSON** (default): Raw API response, ideal for programmatic parsing
- **Table**: Tab-separated, human-readable (`--output table`)
- **CSV**: Properly escaped, importable into spreadsheets (`--output csv`)

## Reference Files

For the full command reference with every subcommand, flag, and positional argument, read `references/commands.md`.
