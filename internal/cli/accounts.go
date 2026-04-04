package cli

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/bwishan/ynab-cli/internal/config"
)

func (a *App) registerAccountCommands() {
	cmd := &Command{
		Name:        "accounts",
		Description: "Manage accounts",
		Subcommands: map[string]*Subcommand{
			"list": {
				Name:        "list",
				Description: "List all accounts for a plan",
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

					data, err := ctx.Client.GetAccounts(planID, lastKnowledge)
					if err != nil {
						return err
					}

					headers := []string{"ID", "NAME", "TYPE", "BALANCE", "ON_BUDGET", "CLOSED"}
					extractRows := func(data []byte) ([][]string, error) {
						var resp struct {
							Data struct {
								Accounts []struct {
									ID       string `json:"id"`
									Name     string `json:"name"`
									Type     string `json:"type"`
									Balance  int64  `json:"balance"`
									OnBudget bool   `json:"on_budget"`
									Closed   bool   `json:"closed"`
								} `json:"accounts"`
							} `json:"data"`
						}
						if err := json.Unmarshal(data, &resp); err != nil {
							return nil, err
						}
						var rows [][]string
						for _, a := range resp.Data.Accounts {
							rows = append(rows, []string{
								a.ID,
								a.Name,
								a.Type,
								strconv.FormatInt(a.Balance, 10),
								strconv.FormatBool(a.OnBudget),
								strconv.FormatBool(a.Closed),
							})
						}
						return rows, nil
					}

					return ctx.Printer.PrintResult(data, headers, extractRows)
				},
			},
			"get": {
				Name:        "get",
				Description: "Get a specific account",
				Usage:       "<plan-id> <account-id>",
				Run: func(ctx *Context, args []string) error {
					var positional []string
					for _, arg := range args {
						if !strings.HasPrefix(arg, "--") {
							positional = append(positional, arg)
						}
					}

					if len(positional) < 2 {
						return fmt.Errorf("usage: ynab accounts get <plan-id> <account-id>")
					}

					planID, err := config.ResolvePlan(positional[0])
					if err != nil {
						return err
					}
					accountID := positional[1]

					data, err := ctx.Client.GetAccountByID(planID, accountID)
					if err != nil {
						return err
					}

					headers := []string{"ID", "NAME", "TYPE", "BALANCE", "ON_BUDGET", "CLOSED"}
					extractRows := func(data []byte) ([][]string, error) {
						var resp struct {
							Data struct {
								Account struct {
									ID       string `json:"id"`
									Name     string `json:"name"`
									Type     string `json:"type"`
									Balance  int64  `json:"balance"`
									OnBudget bool   `json:"on_budget"`
									Closed   bool   `json:"closed"`
								} `json:"account"`
							} `json:"data"`
						}
						if err := json.Unmarshal(data, &resp); err != nil {
							return nil, err
						}
						a := resp.Data.Account
						return [][]string{{
							a.ID,
							a.Name,
							a.Type,
							strconv.FormatInt(a.Balance, 10),
							strconv.FormatBool(a.OnBudget),
							strconv.FormatBool(a.Closed),
						}}, nil
					}

					return ctx.Printer.PrintResult(data, headers, extractRows)
				},
			},
			"create": {
				Name:        "create",
				Description: "Create a new account",
				Usage:       "<plan-id> --name NAME --type TYPE --balance AMOUNT",
				Run: func(ctx *Context, args []string) error {
					var planIDArg string
					var name, acctType string
					var balance int64

					for i := 0; i < len(args); i++ {
						switch {
						case args[i] == "--name" && i+1 < len(args):
							name = args[i+1]
							i++
						case strings.HasPrefix(args[i], "--name="):
							name = strings.TrimPrefix(args[i], "--name=")
						case args[i] == "--type" && i+1 < len(args):
							acctType = args[i+1]
							i++
						case strings.HasPrefix(args[i], "--type="):
							acctType = strings.TrimPrefix(args[i], "--type=")
						case args[i] == "--balance" && i+1 < len(args):
							v, err := strconv.ParseInt(args[i+1], 10, 64)
							if err != nil {
								return fmt.Errorf("invalid --balance value: %s", args[i+1])
							}
							balance = v
							i++
						case strings.HasPrefix(args[i], "--balance="):
							val := strings.TrimPrefix(args[i], "--balance=")
							v, err := strconv.ParseInt(val, 10, 64)
							if err != nil {
								return fmt.Errorf("invalid --balance value: %s", val)
							}
							balance = v
						case !strings.HasPrefix(args[i], "--"):
							if planIDArg == "" {
								planIDArg = args[i]
							}
						}
					}

					if name == "" {
						return fmt.Errorf("--name is required")
					}
					if acctType == "" {
						return fmt.Errorf("--type is required")
					}

					planID, err := config.ResolvePlan(planIDArg)
					if err != nil {
						return err
					}

					body := map[string]interface{}{
						"account": map[string]interface{}{
							"name":    name,
							"type":    acctType,
							"balance": balance,
						},
					}

					data, err := ctx.Client.CreateAccount(planID, body)
					if err != nil {
						return err
					}

					headers := []string{"ID", "NAME", "TYPE", "BALANCE"}
					extractRows := func(data []byte) ([][]string, error) {
						var resp struct {
							Data struct {
								Account struct {
									ID      string `json:"id"`
									Name    string `json:"name"`
									Type    string `json:"type"`
									Balance int64  `json:"balance"`
								} `json:"account"`
							} `json:"data"`
						}
						if err := json.Unmarshal(data, &resp); err != nil {
							return nil, err
						}
						a := resp.Data.Account
						return [][]string{{
							a.ID,
							a.Name,
							a.Type,
							strconv.FormatInt(a.Balance, 10),
						}}, nil
					}

					return ctx.Printer.PrintResult(data, headers, extractRows)
				},
			},
		},
	}
	a.registerCommand(cmd)
}
