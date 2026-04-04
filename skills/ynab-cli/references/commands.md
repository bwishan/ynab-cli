# Full Command Reference

Every command follows: `ynab [global-flags] <resource> <action> [positional-args] [--flags]`

## Table of Contents

- [user](#user)
- [plans](#plans)
- [accounts](#accounts)
- [categories](#categories)
- [category-groups](#category-groups)
- [transactions](#transactions)
- [payees](#payees)
- [payee-locations](#payee-locations)
- [scheduled](#scheduled)
- [months](#months)
- [money-movements](#money-movements)
- [money-movement-groups](#money-movement-groups)
- [configure](#configure)

---

## user

| Subcommand | Usage | Description |
|------------|-------|-------------|
| `get` | `ynab user get` | Get authenticated user info |

---

## plans

| Subcommand | Usage | Description |
|------------|-------|-------------|
| `list` | `ynab plans list [--include-accounts]` | List all budget plans |
| `get` | `ynab plans get <plan-id> [--last-knowledge N]` | Get a specific plan |
| `settings` | `ynab plans settings <plan-id>` | Get plan settings (currency, date format) |

---

## accounts

| Subcommand | Usage | Description |
|------------|-------|-------------|
| `list` | `ynab accounts list <plan-id> [--last-knowledge N]` | List all accounts |
| `get` | `ynab accounts get <plan-id> <account-id>` | Get a specific account |
| `create` | `ynab accounts create <plan-id> --name <name> --type <type> [--balance N]` | Create an account |

Account types: `checking`, `savings`, `creditCard`, `cash`, `lineOfCredit`, `otherAsset`, `otherLiability`, `mortgage`, `autoLoan`, `studentLoan`, `personalLoan`, `medicalDebt`, `otherDebt`.

---

## categories

| Subcommand | Usage | Description |
|------------|-------|-------------|
| `list` | `ynab categories list <plan-id> [--last-knowledge N]` | List all categories (nested under groups) |
| `get` | `ynab categories get <plan-id> <category-id>` | Get a specific category |
| `create` | `ynab categories create <plan-id> --category-group-id <id> --name <name>` | Create a category |
| `update` | `ynab categories update <plan-id> <category-id> [--name <name>] [--note <note>]` | Update a category |
| `get-month` | `ynab categories get-month <plan-id> <month> <category-id>` | Get category for a specific month |
| `update-month` | `ynab categories update-month <plan-id> <month> <category-id> --budgeted <amount>` | Set budget for a category/month |

Month format: `YYYY-MM-DD` (always use first of month, e.g., `2024-03-01`).
Budgeted amounts are in milliunits.

---

## category-groups

| Subcommand | Usage | Description |
|------------|-------|-------------|
| `create` | `ynab category-groups create <plan-id> --name <name>` | Create a category group |
| `update` | `ynab category-groups update <plan-id> <group-id> --name <name>` | Update a category group |

---

## transactions

| Subcommand | Usage | Description |
|------------|-------|-------------|
| `list` | `ynab transactions list <plan-id> [--since-date DATE] [--type TYPE] [--last-knowledge N] [--transaction-sync] [--transaction-sync-db PATH]` | List transactions |
| `get` | `ynab transactions get <plan-id> <transaction-id> [--transaction-sync] [--transaction-sync-db PATH]` | Get a specific transaction |
| `create` | `ynab transactions create <plan-id> --account-id <id> --date <date> --amount <milliunits> [flags]` | Create a transaction |
| `update` | `ynab transactions update <plan-id> <transaction-id> [flags]` | Update a transaction |
| `delete` | `ynab transactions delete <plan-id> <transaction-id>` | Delete a transaction |
| `import` | `ynab transactions import <plan-id>` | Import linked account transactions |
| `sync` | `ynab transactions sync <plan-id> [--since-date DATE] [--type TYPE] [--transaction-sync-db PATH]` | Delta-sync transactions into the local SQLite cache |
| `sync-status` | `ynab transactions sync-status <plan-id> [--transaction-sync-db PATH]` | Show cache sync status |
| `search` | `ynab transactions search <plan-id> [--since-date DATE] [--before-date DATE] [--type TYPE] [--account-id ID] [--category-id ID] [--payee-id ID] [--memo TEXT] [--limit N] [--transaction-sync-db PATH]` | Search cached transactions locally |
| `list-by-account` | `ynab transactions list-by-account <plan-id> <account-id> [--since-date DATE] [--type TYPE] [--last-knowledge N]` | Transactions by account |
| `list-by-category` | `ynab transactions list-by-category <plan-id> <category-id> [--since-date DATE] [--type TYPE] [--last-knowledge N]` | Transactions by category |
| `list-by-payee` | `ynab transactions list-by-payee <plan-id> <payee-id> [--since-date DATE] [--type TYPE] [--last-knowledge N]` | Transactions by payee |
| `list-by-month` | `ynab transactions list-by-month <plan-id> <month> [--since-date DATE] [--type TYPE] [--last-knowledge N]` | Transactions by month |
| `bulk-update` | `ynab transactions bulk-update <plan-id> --data <json>` | Bulk update transactions |

### Transaction create/update flags

| Flag | Required | Description |
|------|----------|-------------|
| `--account-id <id>` | create | Account ID |
| `--date <YYYY-MM-DD>` | create | Transaction date |
| `--amount <milliunits>` | create | Amount in milliunits (negative = outflow) |
| `--payee-id <id>` | no | Payee ID |
| `--payee-name <name>` | no | Payee name (creates payee if needed) |
| `--category-id <id>` | no | Category ID |
| `--memo <text>` | no | Memo text |
| `--cleared <status>` | no | `cleared`, `uncleared`, or `reconciled` |
| `--approved` | no | Mark as approved (flag, no value) |
| `--flag-color <color>` | no | Transaction flag color |
| `--import-id <id>` | no | Import ID for deduplication |

### Transaction type filter

`--type` accepts: `uncategorized`, `unapproved`.

---

## payees

| Subcommand | Usage | Description |
|------------|-------|-------------|
| `list` | `ynab payees list <plan-id> [--last-knowledge N]` | List all payees |
| `get` | `ynab payees get <plan-id> <payee-id>` | Get a specific payee |
| `create` | `ynab payees create <plan-id> --name <name>` | Create a payee |
| `update` | `ynab payees update <plan-id> <payee-id> --name <name>` | Update a payee |

---

## payee-locations

| Subcommand | Usage | Description |
|------------|-------|-------------|
| `list` | `ynab payee-locations list <plan-id>` | List all payee locations |
| `get` | `ynab payee-locations get <plan-id> <location-id>` | Get a specific location |
| `list-by-payee` | `ynab payee-locations list-by-payee <plan-id> <payee-id>` | Locations for a payee |

---

## scheduled

| Subcommand | Usage | Description |
|------------|-------|-------------|
| `list` | `ynab scheduled list <plan-id> [--last-knowledge N]` | List scheduled transactions |
| `get` | `ynab scheduled get <plan-id> <scheduled-id>` | Get a scheduled transaction |
| `create` | `ynab scheduled create <plan-id> --account-id <id> --date <date> --amount <milliunits> --frequency <freq> [flags]` | Create scheduled transaction |
| `update` | `ynab scheduled update <plan-id> <scheduled-id> [flags]` | Update scheduled transaction |
| `delete` | `ynab scheduled delete <plan-id> <scheduled-id>` | Delete scheduled transaction |

### Valid frequencies

`never`, `daily`, `weekly`, `everyOtherWeek`, `twiceAMonth`, `every4Weeks`, `monthly`, `everyOtherMonth`, `every3Months`, `every4Months`, `twiceAYear`, `yearly`, `everyOtherYear`

---

## months

| Subcommand | Usage | Description |
|------------|-------|-------------|
| `list` | `ynab months list <plan-id> [--last-knowledge N]` | List all budget months |
| `get` | `ynab months get <plan-id> <month>` | Get a specific month's budget |

Month format: `YYYY-MM-DD` (first of month).

---

## money-movements

| Subcommand | Usage | Description |
|------------|-------|-------------|
| `list` | `ynab money-movements list <plan-id>` | List all money movements |
| `list-by-month` | `ynab money-movements list-by-month <plan-id> <month>` | Money movements for a month |

---

## money-movement-groups

| Subcommand | Usage | Description |
|------------|-------|-------------|
| `list` | `ynab money-movement-groups list <plan-id>` | List all money movement groups |
| `list-by-month` | `ynab money-movement-groups list-by-month <plan-id> <month>` | Groups for a month |

---

## configure

```bash
ynab configure --token <access-token> [--default-plan <plan-id>] [--transaction-sync] [--transaction-sync-off] [--transaction-sync-db <path>]
```

Saves configuration to `~/.config/ynab-cli/config.json` with `0600` permissions.

Transaction sync notes:
- `--transaction-sync` enables remembered local SQLite-backed sync/cache mode
- `--transaction-sync-off` disables remembered sync mode
- `--transaction-sync-db` sets the SQLite database path used for cached transactions

## Installation by Platform

```bash
# Linux amd64
curl -L https://github.com/bwishan/ynab-cli/releases/latest/download/ynab-linux-amd64 -o ynab

# Linux arm64
curl -L https://github.com/bwishan/ynab-cli/releases/latest/download/ynab-linux-arm64 -o ynab

# macOS Apple Silicon
curl -L https://github.com/bwishan/ynab-cli/releases/latest/download/ynab-darwin-arm64 -o ynab

# macOS Intel
curl -L https://github.com/bwishan/ynab-cli/releases/latest/download/ynab-darwin-amd64 -o ynab

# Windows amd64
curl -L https://github.com/bwishan/ynab-cli/releases/latest/download/ynab-windows-amd64.exe -o ynab.exe

# From source (Go 1.22+)
go install github.com/bwishan/ynab-cli/cmd/ynab@latest
```

After downloading, `chmod +x ynab` and move to a directory in your PATH.
