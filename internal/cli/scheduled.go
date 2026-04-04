package cli

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/bwishan/ynab-cli/internal/config"
)

func (a *App) registerScheduledCommands() {
	cmd := &Command{
		Name:        "scheduled",
		Description: "Manage scheduled transactions",
		Subcommands: map[string]*Subcommand{
			"list": {
				Name:        "list",
				Description: "List all scheduled transactions for a plan",
				Usage:       "<plan-id> [--last-knowledge N]",
				Run: func(ctx *Context, args []string) error {
					var planIDArg string
					var lastKnowledge int64

					for i := 0; i < len(args); i++ {
						switch {
						case args[i] == "--last-knowledge" && i+1 < len(args):
							v, err := strconv.ParseInt(args[i+1], 10, 64)
							if err != nil {
								return fmt.Errorf("invalid --last-knowledge value: %s", args[i+1])
							}
							lastKnowledge = v
							i++
						case strings.HasPrefix(args[i], "--last-knowledge="):
							val := strings.TrimPrefix(args[i], "--last-knowledge=")
							v, err := strconv.ParseInt(val, 10, 64)
							if err != nil {
								return fmt.Errorf("invalid --last-knowledge value: %s", val)
							}
							lastKnowledge = v
						case !strings.HasPrefix(args[i], "--"):
							if planIDArg == "" {
								planIDArg = args[i]
							}
						}
					}

					planID, err := config.ResolvePlan(planIDArg)
					if err != nil {
						return err
					}

					data, err := ctx.Client.GetScheduledTransactions(planID, lastKnowledge)
					if err != nil {
						return err
					}

					headers := []string{"ID", "DATE_FIRST", "DATE_NEXT", "FREQUENCY", "AMOUNT", "ACCOUNT_ID", "PAYEE_ID", "CATEGORY_ID", "MEMO", "FLAG_COLOR", "DELETED"}
					extractRows := func(data []byte) ([][]string, error) {
						var resp struct {
							Data struct {
								ScheduledTransactions []struct {
									ID         string `json:"id"`
									DateFirst  string `json:"date_first"`
									DateNext   string `json:"date_next"`
									Frequency  string `json:"frequency"`
									Amount     int64  `json:"amount"`
									AccountID  string `json:"account_id"`
									PayeeID    string `json:"payee_id"`
									CategoryID string `json:"category_id"`
									Memo       string `json:"memo"`
									FlagColor  string `json:"flag_color"`
									Deleted    bool   `json:"deleted"`
								} `json:"scheduled_transactions"`
							} `json:"data"`
						}
						if err := json.Unmarshal(data, &resp); err != nil {
							return nil, err
						}
						var rows [][]string
						for _, t := range resp.Data.ScheduledTransactions {
							rows = append(rows, []string{
								t.ID,
								t.DateFirst,
								t.DateNext,
								t.Frequency,
								strconv.FormatInt(t.Amount, 10),
								t.AccountID,
								t.PayeeID,
								t.CategoryID,
								t.Memo,
								t.FlagColor,
								fmt.Sprintf("%t", t.Deleted),
							})
						}
						return rows, nil
					}

					return ctx.Printer.PrintResult(data, headers, extractRows)
				},
			},
			"get": {
				Name:        "get",
				Description: "Get a specific scheduled transaction",
				Usage:       "<plan-id> <id>",
				Run: func(ctx *Context, args []string) error {
					var positional []string
					for _, arg := range args {
						if !strings.HasPrefix(arg, "--") {
							positional = append(positional, arg)
						}
					}

					if len(positional) < 2 {
						return fmt.Errorf("usage: ynab scheduled get <plan-id> <id>")
					}

					planID, err := config.ResolvePlan(positional[0])
					if err != nil {
						return err
					}

					data, err := ctx.Client.GetScheduledTransactionByID(planID, positional[1])
					if err != nil {
						return err
					}

					headers := []string{"ID", "DATE_FIRST", "DATE_NEXT", "FREQUENCY", "AMOUNT", "ACCOUNT_ID", "PAYEE_ID", "CATEGORY_ID", "MEMO", "FLAG_COLOR", "DELETED"}
					extractRows := func(data []byte) ([][]string, error) {
						var resp struct {
							Data struct {
								ScheduledTransaction struct {
									ID         string `json:"id"`
									DateFirst  string `json:"date_first"`
									DateNext   string `json:"date_next"`
									Frequency  string `json:"frequency"`
									Amount     int64  `json:"amount"`
									AccountID  string `json:"account_id"`
									PayeeID    string `json:"payee_id"`
									CategoryID string `json:"category_id"`
									Memo       string `json:"memo"`
									FlagColor  string `json:"flag_color"`
									Deleted    bool   `json:"deleted"`
								} `json:"scheduled_transaction"`
							} `json:"data"`
						}
						if err := json.Unmarshal(data, &resp); err != nil {
							return nil, err
						}
						t := resp.Data.ScheduledTransaction
						return [][]string{{
							t.ID,
							t.DateFirst,
							t.DateNext,
							t.Frequency,
							strconv.FormatInt(t.Amount, 10),
							t.AccountID,
							t.PayeeID,
							t.CategoryID,
							t.Memo,
							t.FlagColor,
							fmt.Sprintf("%t", t.Deleted),
						}}, nil
					}

					return ctx.Printer.PrintResult(data, headers, extractRows)
				},
			},
			"create": {
				Name:        "create",
				Description: "Create a scheduled transaction",
				Usage:       "<plan-id> --account-id ID --date DATE --amount AMOUNT --frequency FREQ [--payee-id ID] [--category-id ID] [--memo MEMO] [--flag-color COLOR]",
				Run: func(ctx *Context, args []string) error {
					var planIDArg string
					var accountID, date, frequency string
					var payeeID, categoryID, memo, flagColor string
					var amount int64
					amountSet := false

					for i := 0; i < len(args); i++ {
						switch {
						case args[i] == "--account-id" && i+1 < len(args):
							accountID = args[i+1]
							i++
						case strings.HasPrefix(args[i], "--account-id="):
							accountID = strings.TrimPrefix(args[i], "--account-id=")
						case args[i] == "--date" && i+1 < len(args):
							date = args[i+1]
							i++
						case strings.HasPrefix(args[i], "--date="):
							date = strings.TrimPrefix(args[i], "--date=")
						case args[i] == "--amount" && i+1 < len(args):
							v, err := strconv.ParseInt(args[i+1], 10, 64)
							if err != nil {
								return fmt.Errorf("invalid --amount value: %s", args[i+1])
							}
							amount = v
							amountSet = true
							i++
						case strings.HasPrefix(args[i], "--amount="):
							val := strings.TrimPrefix(args[i], "--amount=")
							v, err := strconv.ParseInt(val, 10, 64)
							if err != nil {
								return fmt.Errorf("invalid --amount value: %s", val)
							}
							amount = v
							amountSet = true
						case args[i] == "--frequency" && i+1 < len(args):
							frequency = args[i+1]
							i++
						case strings.HasPrefix(args[i], "--frequency="):
							frequency = strings.TrimPrefix(args[i], "--frequency=")
						case args[i] == "--payee-id" && i+1 < len(args):
							payeeID = args[i+1]
							i++
						case strings.HasPrefix(args[i], "--payee-id="):
							payeeID = strings.TrimPrefix(args[i], "--payee-id=")
						case args[i] == "--category-id" && i+1 < len(args):
							categoryID = args[i+1]
							i++
						case strings.HasPrefix(args[i], "--category-id="):
							categoryID = strings.TrimPrefix(args[i], "--category-id=")
						case args[i] == "--memo" && i+1 < len(args):
							memo = args[i+1]
							i++
						case strings.HasPrefix(args[i], "--memo="):
							memo = strings.TrimPrefix(args[i], "--memo=")
						case args[i] == "--flag-color" && i+1 < len(args):
							flagColor = args[i+1]
							i++
						case strings.HasPrefix(args[i], "--flag-color="):
							flagColor = strings.TrimPrefix(args[i], "--flag-color=")
						case !strings.HasPrefix(args[i], "--"):
							if planIDArg == "" {
								planIDArg = args[i]
							}
						}
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
					if frequency == "" {
						return fmt.Errorf("--frequency is required")
					}
					if err := validateFrequency(frequency); err != nil {
						return err
					}

					planID, err := config.ResolvePlan(planIDArg)
					if err != nil {
						return err
					}

					fields := map[string]interface{}{
						"account_id": accountID,
						"date_first": date,
						"amount":     amount,
						"frequency":  frequency,
					}
					if payeeID != "" {
						fields["payee_id"] = payeeID
					}
					if categoryID != "" {
						fields["category_id"] = categoryID
					}
					if memo != "" {
						fields["memo"] = memo
					}
					if flagColor != "" {
						fields["flag_color"] = flagColor
					}

					body := map[string]interface{}{
						"scheduled_transaction": fields,
					}

					data, err := ctx.Client.CreateScheduledTransaction(planID, body)
					if err != nil {
						return err
					}

					return ctx.Printer.PrintResult(data, nil, nil)
				},
			},
			"update": {
				Name:        "update",
				Description: "Update a scheduled transaction",
				Usage:       "<plan-id> <id> [--account-id ID] [--date DATE] [--amount AMOUNT] [--frequency FREQ] [--payee-id ID] [--category-id ID] [--memo MEMO] [--flag-color COLOR]",
				Run: func(ctx *Context, args []string) error {
					var positional []string
					var accountID, date, frequency string
					var payeeID, categoryID, memo, flagColor string
					var amount int64
					accountIDSet := false
					dateSet := false
					amountSet := false
					frequencySet := false
					payeeIDSet := false
					categoryIDSet := false
					memoSet := false
					flagColorSet := false

					for i := 0; i < len(args); i++ {
						switch {
						case args[i] == "--account-id" && i+1 < len(args):
							accountID = args[i+1]
							accountIDSet = true
							i++
						case strings.HasPrefix(args[i], "--account-id="):
							accountID = strings.TrimPrefix(args[i], "--account-id=")
							accountIDSet = true
						case args[i] == "--date" && i+1 < len(args):
							date = args[i+1]
							dateSet = true
							i++
						case strings.HasPrefix(args[i], "--date="):
							date = strings.TrimPrefix(args[i], "--date=")
							dateSet = true
						case args[i] == "--amount" && i+1 < len(args):
							v, err := strconv.ParseInt(args[i+1], 10, 64)
							if err != nil {
								return fmt.Errorf("invalid --amount value: %s", args[i+1])
							}
							amount = v
							amountSet = true
							i++
						case strings.HasPrefix(args[i], "--amount="):
							val := strings.TrimPrefix(args[i], "--amount=")
							v, err := strconv.ParseInt(val, 10, 64)
							if err != nil {
								return fmt.Errorf("invalid --amount value: %s", val)
							}
							amount = v
							amountSet = true
						case args[i] == "--frequency" && i+1 < len(args):
							frequency = args[i+1]
							frequencySet = true
							i++
						case strings.HasPrefix(args[i], "--frequency="):
							frequency = strings.TrimPrefix(args[i], "--frequency=")
							frequencySet = true
						case args[i] == "--payee-id" && i+1 < len(args):
							payeeID = args[i+1]
							payeeIDSet = true
							i++
						case strings.HasPrefix(args[i], "--payee-id="):
							payeeID = strings.TrimPrefix(args[i], "--payee-id=")
							payeeIDSet = true
						case args[i] == "--category-id" && i+1 < len(args):
							categoryID = args[i+1]
							categoryIDSet = true
							i++
						case strings.HasPrefix(args[i], "--category-id="):
							categoryID = strings.TrimPrefix(args[i], "--category-id=")
							categoryIDSet = true
						case args[i] == "--memo" && i+1 < len(args):
							memo = args[i+1]
							memoSet = true
							i++
						case strings.HasPrefix(args[i], "--memo="):
							memo = strings.TrimPrefix(args[i], "--memo=")
							memoSet = true
						case args[i] == "--flag-color" && i+1 < len(args):
							flagColor = args[i+1]
							flagColorSet = true
							i++
						case strings.HasPrefix(args[i], "--flag-color="):
							flagColor = strings.TrimPrefix(args[i], "--flag-color=")
							flagColorSet = true
						case !strings.HasPrefix(args[i], "--"):
							positional = append(positional, args[i])
						}
					}

					if len(positional) < 2 {
						return fmt.Errorf("usage: ynab scheduled update <plan-id> <id> [flags]")
					}

					if frequencySet {
						if err := validateFrequency(frequency); err != nil {
							return err
						}
					}

					planID, err := config.ResolvePlan(positional[0])
					if err != nil {
						return err
					}
					transactionID := positional[1]

					fields := map[string]interface{}{}
					if accountIDSet {
						fields["account_id"] = accountID
					}
					if dateSet {
						fields["date_first"] = date
					}
					if amountSet {
						fields["amount"] = amount
					}
					if frequencySet {
						fields["frequency"] = frequency
					}
					if payeeIDSet {
						fields["payee_id"] = payeeID
					}
					if categoryIDSet {
						fields["category_id"] = categoryID
					}
					if memoSet {
						fields["memo"] = memo
					}
					if flagColorSet {
						fields["flag_color"] = flagColor
					}

					if len(fields) == 0 {
						return fmt.Errorf("at least one flag is required for update")
					}

					body := map[string]interface{}{
						"scheduled_transaction": fields,
					}

					data, err := ctx.Client.UpdateScheduledTransaction(planID, transactionID, body)
					if err != nil {
						return err
					}

					return ctx.Printer.PrintResult(data, nil, nil)
				},
			},
			"delete": {
				Name:        "delete",
				Description: "Delete a scheduled transaction",
				Usage:       "<plan-id> <id>",
				Run: func(ctx *Context, args []string) error {
					var positional []string
					for _, arg := range args {
						if !strings.HasPrefix(arg, "--") {
							positional = append(positional, arg)
						}
					}

					if len(positional) < 2 {
						return fmt.Errorf("usage: ynab scheduled delete <plan-id> <id>")
					}

					planID, err := config.ResolvePlan(positional[0])
					if err != nil {
						return err
					}

					data, err := ctx.Client.DeleteScheduledTransaction(planID, positional[1])
					if err != nil {
						return err
					}

					return ctx.Printer.PrintResult(data, nil, nil)
				},
			},
		},
	}
	a.registerCommand(cmd)
}

var validFrequencies = map[string]bool{
	"never":          true,
	"daily":          true,
	"weekly":         true,
	"everyOtherWeek": true,
	"twiceAMonth":    true,
	"every4Weeks":    true,
	"monthly":        true,
	"everyOtherMonth": true,
	"every3Months":   true,
	"every4Months":   true,
	"twiceAYear":     true,
	"yearly":         true,
	"everyOtherYear": true,
}

func validateFrequency(freq string) error {
	if !validFrequencies[freq] {
		return fmt.Errorf("invalid frequency %q; valid values: never, daily, weekly, everyOtherWeek, twiceAMonth, every4Weeks, monthly, everyOtherMonth, every3Months, every4Months, twiceAYear, yearly, everyOtherYear", freq)
	}
	return nil
}
