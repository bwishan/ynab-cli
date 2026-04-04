package cli

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/bwishan/ynab-cli/internal/config"
)

func (a *App) registerTransactionCommands() {
	cmd := &Command{
		Name:        "transactions",
		Description: "Manage transactions",
		Subcommands: map[string]*Subcommand{},
	}

	cmd.Subcommands["list"] = &Subcommand{
		Name:        "list",
		Description: "List all transactions for a plan",
		Usage:       "<plan-id> [--since-date DATE] [--type TYPE] [--last-knowledge N]",
		Run: func(ctx *Context, args []string) error {
			planExplicit, sinceDate, txType, lastKnowledge, err := parseTransactionListArgs(args)
			if err != nil {
				return err
			}

			planID, err := config.ResolvePlan(planExplicit)
			if err != nil {
				return err
			}

			data, err := ctx.Client.GetTransactions(planID, sinceDate, txType, lastKnowledge)
			if err != nil {
				return err
			}

			return ctx.Printer.PrintResult(data, transactionHeaders, extractTransactionRows)
		},
	}

	cmd.Subcommands["get"] = &Subcommand{
		Name:        "get",
		Description: "Get a specific transaction",
		Usage:       "<plan-id> <transaction-id>",
		Run: func(ctx *Context, args []string) error {
			var positional []string
			for _, arg := range args {
				if !strings.HasPrefix(arg, "--") {
					positional = append(positional, arg)
				}
			}

			if len(positional) < 2 {
				return fmt.Errorf("usage: ynab transactions get <plan-id> <transaction-id>")
			}

			planID, err := config.ResolvePlan(positional[0])
			if err != nil {
				return err
			}

			data, err := ctx.Client.GetTransactionByID(planID, positional[1])
			if err != nil {
				return err
			}

			return ctx.Printer.PrintResult(data, transactionHeaders, extractSingleTransactionRows)
		},
	}

	cmd.Subcommands["create"] = &Subcommand{
		Name:        "create",
		Description: "Create a new transaction",
		Usage:       "<plan-id> --account-id ID --date DATE --amount AMOUNT [--payee-id ID] [--payee-name NAME] [--category-id ID] [--memo MEMO] [--cleared cleared|uncleared|reconciled] [--approved] [--flag-color COLOR] [--import-id ID]",
		Run: func(ctx *Context, args []string) error {
			var planExplicit string
			var accountID, date, payeeID, payeeName, categoryID, memo, cleared, flagColor, importID string
			var amount int64
			amountSet := false
			approved := false

			positional := []string{}
			for i := 0; i < len(args); i++ {
				key, val, consumed, isFlag := parseFlag(args, i)
				if !isFlag {
					positional = append(positional, args[i])
					continue
				}
				if val == "" && consumed == 0 {
					if key == "approved" {
						approved = true
						continue
					}
					return fmt.Errorf("flag %s requires a value", args[i])
				}
				i += consumed
				switch key {
				case "account-id":
					accountID = val
				case "date":
					date = val
				case "amount":
					v, err := strconv.ParseInt(val, 10, 64)
					if err != nil {
						return fmt.Errorf("invalid --amount value: %s", val)
					}
					amount = v
					amountSet = true
				case "payee-id":
					payeeID = val
				case "payee-name":
					payeeName = val
				case "category-id":
					categoryID = val
				case "memo":
					memo = val
				case "cleared":
					cleared = val
				case "approved":
					approved = true
				case "flag-color":
					flagColor = val
				case "import-id":
					importID = val
				default:
					return fmt.Errorf("unknown flag: --%s", key)
				}
			}

			if len(positional) > 0 {
				planExplicit = positional[0]
			}

			planID, err := config.ResolvePlan(planExplicit)
			if err != nil {
				return err
			}

			if accountID == "" {
				return fmt.Errorf("--account-id is required")
			}
			if date == "" {
				return fmt.Errorf("--date is required")
			}
			if !amountSet {
				return fmt.Errorf("--amount is required")
			}

			tx := map[string]interface{}{
				"account_id": accountID,
				"date":       date,
				"amount":     amount,
			}
			if payeeID != "" {
				tx["payee_id"] = payeeID
			}
			if payeeName != "" {
				tx["payee_name"] = payeeName
			}
			if categoryID != "" {
				tx["category_id"] = categoryID
			}
			if memo != "" {
				tx["memo"] = memo
			}
			if cleared != "" {
				tx["cleared"] = cleared
			}
			if approved {
				tx["approved"] = true
			}
			if flagColor != "" {
				tx["flag_color"] = flagColor
			}
			if importID != "" {
				tx["import_id"] = importID
			}

			body := map[string]interface{}{
				"transaction": tx,
			}

			data, err := ctx.Client.CreateTransaction(planID, body)
			if err != nil {
				return err
			}

			return ctx.Printer.PrintResult(data, nil, nil)
		},
	}

	cmd.Subcommands["update"] = &Subcommand{
		Name:        "update",
		Description: "Update a transaction",
		Usage:       "<plan-id> <transaction-id> [--account-id ID] [--date DATE] [--amount AMOUNT] [--payee-id ID] [--payee-name NAME] [--category-id ID] [--memo MEMO] [--cleared cleared|uncleared|reconciled] [--approved] [--flag-color COLOR] [--import-id ID]",
		Run: func(ctx *Context, args []string) error {
			var accountID, date, payeeID, payeeName, categoryID, memo, cleared, flagColor, importID string
			var amount int64
			amountSet := false
			approved := false

			positional := []string{}
			for i := 0; i < len(args); i++ {
				key, val, consumed, isFlag := parseFlag(args, i)
				if !isFlag {
					positional = append(positional, args[i])
					continue
				}
				if val == "" && consumed == 0 {
					if key == "approved" {
						approved = true
						continue
					}
					return fmt.Errorf("flag %s requires a value", args[i])
				}
				i += consumed
				switch key {
				case "account-id":
					accountID = val
				case "date":
					date = val
				case "amount":
					v, err := strconv.ParseInt(val, 10, 64)
					if err != nil {
						return fmt.Errorf("invalid --amount value: %s", val)
					}
					amount = v
					amountSet = true
				case "payee-id":
					payeeID = val
				case "payee-name":
					payeeName = val
				case "category-id":
					categoryID = val
				case "memo":
					memo = val
				case "cleared":
					cleared = val
				case "approved":
					approved = true
				case "flag-color":
					flagColor = val
				case "import-id":
					importID = val
				default:
					return fmt.Errorf("unknown flag: --%s", key)
				}
			}

			if len(positional) < 2 {
				return fmt.Errorf("usage: ynab transactions update <plan-id> <transaction-id> [flags]")
			}

			planID, err := config.ResolvePlan(positional[0])
			if err != nil {
				return err
			}
			transactionID := positional[1]

			tx := map[string]interface{}{}
			if accountID != "" {
				tx["account_id"] = accountID
			}
			if date != "" {
				tx["date"] = date
			}
			if amountSet {
				tx["amount"] = amount
			}
			if payeeID != "" {
				tx["payee_id"] = payeeID
			}
			if payeeName != "" {
				tx["payee_name"] = payeeName
			}
			if categoryID != "" {
				tx["category_id"] = categoryID
			}
			if memo != "" {
				tx["memo"] = memo
			}
			if cleared != "" {
				tx["cleared"] = cleared
			}
			if approved {
				tx["approved"] = true
			}
			if flagColor != "" {
				tx["flag_color"] = flagColor
			}
			if importID != "" {
				tx["import_id"] = importID
			}

			if len(tx) == 0 {
				return fmt.Errorf("at least one field flag is required")
			}

			body := map[string]interface{}{
				"transaction": tx,
			}

			data, err := ctx.Client.UpdateTransaction(planID, transactionID, body)
			if err != nil {
				return err
			}

			return ctx.Printer.PrintResult(data, nil, nil)
		},
	}

	cmd.Subcommands["delete"] = &Subcommand{
		Name:        "delete",
		Description: "Delete a transaction",
		Usage:       "<plan-id> <transaction-id>",
		Run: func(ctx *Context, args []string) error {
			var positional []string
			for _, arg := range args {
				if !strings.HasPrefix(arg, "--") {
					positional = append(positional, arg)
				}
			}

			if len(positional) < 2 {
				return fmt.Errorf("usage: ynab transactions delete <plan-id> <transaction-id>")
			}

			planID, err := config.ResolvePlan(positional[0])
			if err != nil {
				return err
			}

			data, err := ctx.Client.DeleteTransaction(planID, positional[1])
			if err != nil {
				return err
			}

			return ctx.Printer.PrintResult(data, nil, nil)
		},
	}

	cmd.Subcommands["import"] = &Subcommand{
		Name:        "import",
		Description: "Import transactions for linked accounts",
		Usage:       "<plan-id>",
		Run: func(ctx *Context, args []string) error {
			var planExplicit string
			for _, arg := range args {
				if !strings.HasPrefix(arg, "--") {
					if planExplicit == "" {
						planExplicit = arg
					}
				}
			}

			planID, err := config.ResolvePlan(planExplicit)
			if err != nil {
				return err
			}

			data, err := ctx.Client.ImportTransactions(planID)
			if err != nil {
				return err
			}

			return ctx.Printer.PrintResult(data, nil, nil)
		},
	}

	cmd.Subcommands["list-by-account"] = &Subcommand{
		Name:        "list-by-account",
		Description: "List transactions for a specific account",
		Usage:       "<plan-id> <account-id> [--since-date DATE] [--type TYPE] [--last-knowledge N]",
		Run: func(ctx *Context, args []string) error {
			positional, sinceDate, txType, lastKnowledge, err := parseTransactionFilterArgs(args)
			if err != nil {
				return err
			}

			if len(positional) < 2 {
				return fmt.Errorf("usage: ynab transactions list-by-account <plan-id> <account-id> [flags]")
			}

			planID, err := config.ResolvePlan(positional[0])
			if err != nil {
				return err
			}

			data, err := ctx.Client.GetTransactionsByAccount(planID, positional[1], sinceDate, txType, lastKnowledge)
			if err != nil {
				return err
			}

			return ctx.Printer.PrintResult(data, transactionHeaders, extractTransactionRows)
		},
	}

	cmd.Subcommands["list-by-category"] = &Subcommand{
		Name:        "list-by-category",
		Description: "List transactions for a specific category",
		Usage:       "<plan-id> <category-id> [--since-date DATE] [--type TYPE] [--last-knowledge N]",
		Run: func(ctx *Context, args []string) error {
			positional, sinceDate, txType, lastKnowledge, err := parseTransactionFilterArgs(args)
			if err != nil {
				return err
			}

			if len(positional) < 2 {
				return fmt.Errorf("usage: ynab transactions list-by-category <plan-id> <category-id> [flags]")
			}

			planID, err := config.ResolvePlan(positional[0])
			if err != nil {
				return err
			}

			data, err := ctx.Client.GetTransactionsByCategory(planID, positional[1], sinceDate, txType, lastKnowledge)
			if err != nil {
				return err
			}

			return ctx.Printer.PrintResult(data, transactionHeaders, extractTransactionRows)
		},
	}

	cmd.Subcommands["list-by-payee"] = &Subcommand{
		Name:        "list-by-payee",
		Description: "List transactions for a specific payee",
		Usage:       "<plan-id> <payee-id> [--since-date DATE] [--type TYPE] [--last-knowledge N]",
		Run: func(ctx *Context, args []string) error {
			positional, sinceDate, txType, lastKnowledge, err := parseTransactionFilterArgs(args)
			if err != nil {
				return err
			}

			if len(positional) < 2 {
				return fmt.Errorf("usage: ynab transactions list-by-payee <plan-id> <payee-id> [flags]")
			}

			planID, err := config.ResolvePlan(positional[0])
			if err != nil {
				return err
			}

			data, err := ctx.Client.GetTransactionsByPayee(planID, positional[1], sinceDate, txType, lastKnowledge)
			if err != nil {
				return err
			}

			return ctx.Printer.PrintResult(data, transactionHeaders, extractTransactionRows)
		},
	}

	cmd.Subcommands["list-by-month"] = &Subcommand{
		Name:        "list-by-month",
		Description: "List transactions for a specific month",
		Usage:       "<plan-id> <month> [--since-date DATE] [--type TYPE] [--last-knowledge N]",
		Run: func(ctx *Context, args []string) error {
			positional, sinceDate, txType, lastKnowledge, err := parseTransactionFilterArgs(args)
			if err != nil {
				return err
			}

			if len(positional) < 2 {
				return fmt.Errorf("usage: ynab transactions list-by-month <plan-id> <month> [flags]")
			}

			planID, err := config.ResolvePlan(positional[0])
			if err != nil {
				return err
			}

			data, err := ctx.Client.GetTransactionsByMonth(planID, positional[1], sinceDate, txType, lastKnowledge)
			if err != nil {
				return err
			}

			return ctx.Printer.PrintResult(data, transactionHeaders, extractTransactionRows)
		},
	}

	cmd.Subcommands["bulk-update"] = &Subcommand{
		Name:        "bulk-update",
		Description: "Update multiple transactions at once",
		Usage:       "<plan-id> --data JSON_STRING",
		Run: func(ctx *Context, args []string) error {
			var planExplicit string
			var dataStr string

			positional := []string{}
			for i := 0; i < len(args); i++ {
				switch {
				case args[i] == "--data" && i+1 < len(args):
					dataStr = args[i+1]
					i++
				case strings.HasPrefix(args[i], "--data="):
					dataStr = strings.TrimPrefix(args[i], "--data=")
				case !strings.HasPrefix(args[i], "--"):
					positional = append(positional, args[i])
				default:
					return fmt.Errorf("unknown flag: %s", args[i])
				}
			}

			if len(positional) > 0 {
				planExplicit = positional[0]
			}

			planID, err := config.ResolvePlan(planExplicit)
			if err != nil {
				return err
			}

			if dataStr == "" {
				return fmt.Errorf("--data is required")
			}

			var txArray json.RawMessage
			if err := json.Unmarshal([]byte(dataStr), &txArray); err != nil {
				return fmt.Errorf("invalid --data JSON: %w", err)
			}

			body := map[string]json.RawMessage{
				"transactions": txArray,
			}

			data, err := ctx.Client.UpdateTransactions(planID, body)
			if err != nil {
				return err
			}

			return ctx.Printer.PrintResult(data, nil, nil)
		},
	}

	a.registerCommand(cmd)
}

// parseFlag parses a flag at position i in args. Returns the key name (without --),
// value, number of additional args consumed, and whether it was a flag at all.
func parseFlag(args []string, i int) (key, val string, consumed int, isFlag bool) {
	arg := args[i]
	if !strings.HasPrefix(arg, "--") {
		return "", "", 0, false
	}

	arg = strings.TrimPrefix(arg, "--")

	if idx := strings.Index(arg, "="); idx >= 0 {
		return arg[:idx], arg[idx+1:], 0, true
	}

	// Boolean flag (no next arg or next arg is a flag)
	if i+1 >= len(args) || strings.HasPrefix(args[i+1], "--") {
		return arg, "", 0, true
	}

	return arg, args[i+1], 1, true
}

// parseTransactionListArgs parses args for the "list" subcommand, returning
// the plan ID arg and filter parameters.
func parseTransactionListArgs(args []string) (planExplicit, sinceDate, txType string, lastKnowledge int64, err error) {
	positional, sinceDate, txType, lastKnowledge, err := parseTransactionFilterArgs(args)
	if err != nil {
		return "", "", "", 0, err
	}
	if len(positional) > 0 {
		planExplicit = positional[0]
	}
	return planExplicit, sinceDate, txType, lastKnowledge, nil
}

// parseTransactionFilterArgs parses common list filter flags from args,
// returning positional args and filter values.
func parseTransactionFilterArgs(args []string) (positional []string, sinceDate, txType string, lastKnowledge int64, err error) {
	for i := 0; i < len(args); i++ {
		switch {
		case args[i] == "--since-date" && i+1 < len(args):
			sinceDate = args[i+1]
			i++
		case strings.HasPrefix(args[i], "--since-date="):
			sinceDate = strings.TrimPrefix(args[i], "--since-date=")
		case args[i] == "--type" && i+1 < len(args):
			txType = args[i+1]
			i++
		case strings.HasPrefix(args[i], "--type="):
			txType = strings.TrimPrefix(args[i], "--type=")
		case args[i] == "--last-knowledge" && i+1 < len(args):
			v, parseErr := strconv.ParseInt(args[i+1], 10, 64)
			if parseErr != nil {
				return nil, "", "", 0, fmt.Errorf("invalid --last-knowledge value: %s", args[i+1])
			}
			lastKnowledge = v
			i++
		case strings.HasPrefix(args[i], "--last-knowledge="):
			val := strings.TrimPrefix(args[i], "--last-knowledge=")
			v, parseErr := strconv.ParseInt(val, 10, 64)
			if parseErr != nil {
				return nil, "", "", 0, fmt.Errorf("invalid --last-knowledge value: %s", val)
			}
			lastKnowledge = v
		case !strings.HasPrefix(args[i], "--"):
			positional = append(positional, args[i])
		}
	}
	return positional, sinceDate, txType, lastKnowledge, nil
}

// transactionHeaders defines the table columns for transaction list output.
var transactionHeaders = []string{"DATE", "PAYEE", "CATEGORY", "MEMO", "AMOUNT", "CLEARED", "APPROVED"}

// extractTransactionRows parses a multi-transaction response for table output.
func extractTransactionRows(data []byte) ([][]string, error) {
	var resp struct {
		Data struct {
			Transactions []struct {
				Date         string `json:"date"`
				PayeeName    string `json:"payee_name"`
				CategoryName string `json:"category_name"`
				Memo         string `json:"memo"`
				Amount       int64  `json:"amount"`
				Cleared      string `json:"cleared"`
				Approved     bool   `json:"approved"`
			} `json:"transactions"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing transactions response: %w", err)
	}

	var rows [][]string
	for _, t := range resp.Data.Transactions {
		rows = append(rows, []string{
			t.Date,
			t.PayeeName,
			t.CategoryName,
			t.Memo,
			milliunitToString(t.Amount),
			t.Cleared,
			strconv.FormatBool(t.Approved),
		})
	}
	return rows, nil
}

// extractSingleTransactionRows parses a single-transaction response for table output.
func extractSingleTransactionRows(data []byte) ([][]string, error) {
	var resp struct {
		Data struct {
			Transaction struct {
				Date         string `json:"date"`
				PayeeName    string `json:"payee_name"`
				CategoryName string `json:"category_name"`
				Memo         string `json:"memo"`
				Amount       int64  `json:"amount"`
				Cleared      string `json:"cleared"`
				Approved     bool   `json:"approved"`
			} `json:"transaction"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing transaction response: %w", err)
	}

	t := resp.Data.Transaction
	return [][]string{{
		t.Date,
		t.PayeeName,
		t.CategoryName,
		t.Memo,
		milliunitToString(t.Amount),
		t.Cleared,
		strconv.FormatBool(t.Approved),
	}}, nil
}
