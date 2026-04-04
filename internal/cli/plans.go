package cli

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/bwishan/ynab-cli/internal/config"
)

func (a *App) registerPlanCommands() {
	cmd := &Command{
		Name:        "plans",
		Description: "Manage budgets/plans",
		Subcommands: map[string]*Subcommand{
			"list": {
				Name:        "list",
				Description: "List all plans",
				Usage:       "[--include-accounts]",
				Run: func(ctx *Context, args []string) error {
					includeAccounts := false
					for _, arg := range args {
						if arg == "--include-accounts" {
							includeAccounts = true
						}
					}

					data, err := ctx.Client.GetPlans(includeAccounts)
					if err != nil {
						return err
					}

					headers := []string{"ID", "NAME", "LAST_MODIFIED"}
					extractRows := func(data []byte) ([][]string, error) {
						var resp struct {
							Data struct {
								Budgets []struct {
									ID           string `json:"id"`
									Name         string `json:"name"`
									LastModified string `json:"last_modified_on"`
								} `json:"budgets"`
							} `json:"data"`
						}
						if err := json.Unmarshal(data, &resp); err != nil {
							return nil, err
						}
						var rows [][]string
						for _, b := range resp.Data.Budgets {
							rows = append(rows, []string{b.ID, b.Name, b.LastModified})
						}
						return rows, nil
					}

					return ctx.Printer.PrintResult(data, headers, extractRows)
				},
			},
			"get": {
				Name:        "get",
				Description: "Get a specific plan",
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

					data, err := ctx.Client.GetPlanByID(planID, lastKnowledge)
					if err != nil {
						return err
					}

					headers := []string{"ID", "NAME", "LAST_MODIFIED"}
					extractRows := func(data []byte) ([][]string, error) {
						var resp struct {
							Data struct {
								Budget struct {
									ID           string `json:"id"`
									Name         string `json:"name"`
									LastModified string `json:"last_modified_on"`
								} `json:"budget"`
							} `json:"data"`
						}
						if err := json.Unmarshal(data, &resp); err != nil {
							return nil, err
						}
						b := resp.Data.Budget
						return [][]string{{b.ID, b.Name, b.LastModified}}, nil
					}

					return ctx.Printer.PrintResult(data, headers, extractRows)
				},
			},
			"settings": {
				Name:        "settings",
				Description: "Get plan settings",
				Usage:       "<plan-id>",
				Run: func(ctx *Context, args []string) error {
					var planIDArg string
					for _, arg := range args {
						if !strings.HasPrefix(arg, "--") {
							if planIDArg == "" {
								planIDArg = arg
							}
						}
					}

					planID, err := config.ResolvePlan(planIDArg)
					if err != nil {
						return err
					}

					data, err := ctx.Client.GetPlanSettings(planID)
					if err != nil {
						return err
					}

					headers := []string{"CURRENCY_FORMAT", "DATE_FORMAT"}
					extractRows := func(data []byte) ([][]string, error) {
						var resp struct {
							Data struct {
								Settings struct {
									DateFormat struct {
										Format string `json:"format"`
									} `json:"date_format"`
									CurrencyFormat struct {
										ISOCode string `json:"iso_code"`
									} `json:"currency_format"`
								} `json:"settings"`
							} `json:"data"`
						}
						if err := json.Unmarshal(data, &resp); err != nil {
							return nil, err
						}
						s := resp.Data.Settings
						return [][]string{{s.CurrencyFormat.ISOCode, s.DateFormat.Format}}, nil
					}

					return ctx.Printer.PrintResult(data, headers, extractRows)
				},
			},
		},
	}
	a.registerCommand(cmd)
}
