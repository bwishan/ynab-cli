---
name: ynab-cli
description: Agent-friendly interface for the YNAB (You Need A Budget) API. Use this skill whenever the user mentions YNAB, budgeting, personal finance, budget tracking, spending analysis, transaction management, or wants to interact with their YNAB account programmatically. Also trigger when the user asks about budget categories, payees, scheduled transactions, account balances, monthly budgets, or any financial data that lives in YNAB — even if they don't explicitly say "YNAB" but describe a budgeting workflow that matches what YNAB does.
---

# ynab-cli

A tool for reading and writing YNAB budget data. It works in three modes depending on how it's configured — pick the one that matches your setup and ignore the others.

## Determine Your Mode

Before doing anything, figure out which mode you're operating in:

**MCP mode** — You have `search_commands` and `run_command` tools available (possibly prefixed like `mcp__ynab__search_commands`). If so, use those tools directly. No shell needed.

**CLI mode** — You have shell access and the `ynab` binary is installed. Run commands via shell as `ynab <resource> <action> [args] [--flags]`.

If you're unsure, try calling `search_commands` first. If that tool exists, you're in MCP mode. If not, try running `ynab --version` in the shell to confirm CLI mode. If neither works, the user needs to set up ynab-cli first (see Setup below).

## MCP Mode

You have two tools:

**`search_commands`** — Discover available commands, their arguments, and flags. Call it before running an unfamiliar command to learn the exact syntax.

```
search_commands({})                           # list everything
search_commands({"query": "transaction"})     # search by keyword
search_commands({"resource": "categories"})   # filter by resource type
```

The response includes command names, positional arguments (with required/optional), and all available flags with types and descriptions.

**`run_command`** — Execute a command. Pass the command string exactly as you would after `ynab` on the command line — positional args and flags included.

```
run_command({"command": "plans list"})
run_command({"command": "accounts list <plan-id>"})
run_command({"command": "transactions list <plan-id> --since-date 2024-01-01"})
run_command({"command": "transactions create <plan-id> --account-id <id> --date 2024-03-15 --amount -50000 --payee-name \"Coffee Shop\""})
run_command({"command": "transactions create <plan-id> --account-id <id> --date 2024-03-15 --amount -153932 --payee-name \"Card Charge\" --split \"[{\\\"amount\\\":-129123,\\\"category_id\\\":\\\"abc\\\"},{\\\"amount\\\":-24809,\\\"category_id\\\":\\\"def\\\"}]\""})
```

Output is always JSON. Errors also come back as JSON with `id`, `name`, and `detail` fields.

When you don't know the exact flags for a command, call `search_commands` first to look them up rather than guessing.

## CLI Mode

Every command follows: `ynab <resource> <action> [positional-args] [--flags]`

Output is JSON by default. Use `--pretty` for human-readable JSON, `--output table` for tabular, or `--output csv` for CSV.

```bash
ynab plans list
ynab accounts list <plan-id>
ynab transactions list <plan-id> --since-date 2024-01-01
ynab transactions create <plan-id> --account-id <id> --date 2024-03-15 --amount -50000 --payee-name "Coffee Shop"
```

### Setup (CLI mode only)

```bash
# Install (macOS Apple Silicon example — see references/commands.md for all platforms)
curl -L https://github.com/bwishan/ynab-cli/releases/latest/download/ynab-darwin-arm64 -o /usr/local/bin/ynab && chmod +x /usr/local/bin/ynab

# Authenticate (user must provide their YNAB Personal Access Token)
export YNAB_ACCESS_TOKEN="<token>"

# Verify
ynab user get
```

Tokens are created at https://app.ynab.com/settings/developer. The user must provide one — never generate or guess tokens.

### Authentication Priority (CLI mode)

1. `--token <token>` flag (highest)
2. `YNAB_ACCESS_TOKEN` environment variable
3. `~/.config/ynab-cli/config.json` file

For MCP mode, authentication is handled by the server configuration — you don't need to worry about tokens.

## Key Concepts

These apply regardless of mode.

### Plan IDs

Most commands require a `plan-id` (YNAB calls budgets "plans"). Discover it first:

```
plans list
```

If the user has multiple plans, ask which one they mean before proceeding. If there's only one, use it directly.

Then either pass it to every command, or (CLI mode only) set a default:

```bash
export YNAB_PLAN_ID="<plan-id>"
# or
ynab configure --default-plan <plan-id>
```

### Milliunits

YNAB represents all currency amounts in "milliunits" — 1/1000 of a currency unit. This avoids floating-point issues. Always convert when displaying to users or accepting input:

| User sees | API value |
|-----------|-----------|
| $1.00     | 1000      |
| $50.00    | 50000     |
| -$123.45  | -123450   |

Negative amounts = outflows (spending). Positive = inflows (income).

Many responses also include pre-formatted fields like `amount_formatted` ("$50.00"), `activity_formatted` ("-$12.50"), and `balance_formatted` ("$200.00") — use these when presenting values to users rather than doing manual conversion.

### Split Transactions (Subtransactions)

For split transactions, pass one or more `--split` flags (or one JSON array), and ensure the parent `--amount` equals the sum of subtransaction amounts. The parent transaction must not set `--category-id` when `--split` is used.

### Delta Sync

Many endpoints support incremental fetching via `--last-knowledge <number>`. The response includes a `server_knowledge` value — pass it on the next request to get only what changed. This is critical for staying within the 200 requests/hour rate limit.

```
# First call returns server_knowledge: 500
transactions list <plan-id>
# Next call gets only changes since then
transactions list <plan-id> --last-knowledge 500
```

Supported by: `accounts list`, `categories list`, `transactions list`, `payees list`, `scheduled list`, `months list`.

## Resources

Available resources: `user`, `plans`, `accounts`, `categories`, `category-groups`, `transactions`, `payees`, `payee-locations`, `scheduled`, `months`, `money-movements`, `money-movement-groups`

In MCP mode, call `search_commands({"resource": "<name>"})` to see all commands for a resource. In CLI mode, run `ynab <resource> --help`.

## Common Workflows

### Budget Overview

```
plans list                                     # Find the plan
categories list <plan-id>                      # All categories with budgeted/spent/available
months get <plan-id> 2024-03-01                # Specific month summary
```

### Transaction Management

```
# Recent transactions
transactions list <plan-id> --since-date 2024-01-01

# Create a transaction (outflow of $50)
transactions create <plan-id> --account-id <id> --date 2024-03-15 --amount -50000 --payee-name "Coffee Shop" --category-id <id> --memo "Morning coffee"

# Create a split transaction (parent category omitted; split amounts sum to parent amount)
transactions create <plan-id> --account-id <id> --date 2024-03-15 --amount -153932 --payee-name "Card Charge" --split '[{"amount":-129123,"category_id":"abc","memo":"Club dues"},{"amount":-24809,"category_id":"def","memo":"Dining"}]'

# Update with repeatable split flags
transactions update <plan-id> <transaction-id> --amount -153932 --split 'amount=-129123,category_id=abc,memo=Club dues' --split 'amount=-24809,category_id=def,memo=Dining'

# Filter by account, category, or payee
transactions list-by-account <plan-id> <account-id>
transactions list-by-category <plan-id> <category-id>
transactions list-by-payee <plan-id> <payee-id>

# Uncategorized transactions
transactions list <plan-id> --type uncategorized
```

### Budget Adjustments

```
# Set a category budget for a month ($500 = 500000 milliunits)
categories update-month <plan-id> 2024-03-01 <category-id> --budgeted 500000
```

### Scheduled Transactions

Creating a scheduled transaction requires `--account-id`, `--date`, `--amount`, and `--frequency`. Note that unlike `transactions create`, scheduled transactions use `--payee-id` (not `--payee-name`) — you may need to look up or create the payee first.

```
# Look up the payee ID first
payees list <plan-id>

# Then create the scheduled transaction
scheduled create <plan-id> --account-id <id> --date 2024-04-01 --amount -100000 --frequency monthly --payee-id <payee-id> --category-id <id>
```

Valid frequencies: `never`, `daily`, `weekly`, `everyOtherWeek`, `twiceAMonth`, `every4Weeks`, `monthly`, `everyOtherMonth`, `every3Months`, `every4Months`, `twiceAYear`, `yearly`, `everyOtherYear`.

## Error Handling

Errors are JSON with this structure:

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

## Reference Files

For the full command reference with every subcommand, flag, and positional argument, read `references/commands.md`.
