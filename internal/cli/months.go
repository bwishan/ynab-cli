package cli

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/bwishan/ynab-cli/internal/config"
)

func (a *App) registerMonthCommands() {
	cmd := &Command{
		Name:        "months",
		Description: "Manage budget months",
		Subcommands: map[string]*Subcommand{
			"list": {
				Name:        "list",
				Description: "List all budget months for a plan",
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

					data, err := ctx.Client.GetMonths(planID, lastKnowledge)
					if err != nil {
						return err
					}

					headers := []string{"MONTH", "NOTE", "INCOME", "BUDGETED", "ACTIVITY", "TO_BE_BUDGETED", "DELETED"}
					extractRows := func(data []byte) ([][]string, error) {
						var resp struct {
							Data struct {
								Months []struct {
									Month         string `json:"month"`
									Note          string `json:"note"`
									Income        int64  `json:"income"`
									Budgeted      int64  `json:"budgeted"`
									Activity      int64  `json:"activity"`
									ToBeBudgeted  int64  `json:"to_be_budgeted"`
									Deleted       bool   `json:"deleted"`
								} `json:"months"`
							} `json:"data"`
						}
						if err := json.Unmarshal(data, &resp); err != nil {
							return nil, err
						}
						var rows [][]string
						for _, m := range resp.Data.Months {
							rows = append(rows, []string{
								m.Month,
								m.Note,
								strconv.FormatInt(m.Income, 10),
								strconv.FormatInt(m.Budgeted, 10),
								strconv.FormatInt(m.Activity, 10),
								strconv.FormatInt(m.ToBeBudgeted, 10),
								fmt.Sprintf("%t", m.Deleted),
							})
						}
						return rows, nil
					}

					return ctx.Printer.PrintResult(data, headers, extractRows)
				},
			},
			"get": {
				Name:        "get",
				Description: "Get a specific budget month",
				Usage:       "<plan-id> <month>",
				Run: func(ctx *Context, args []string) error {
					var positional []string
					for _, arg := range args {
						if !strings.HasPrefix(arg, "--") {
							positional = append(positional, arg)
						}
					}

					if len(positional) < 2 {
						return fmt.Errorf("usage: ynab months get <plan-id> <month>")
					}

					planID, err := config.ResolvePlan(positional[0])
					if err != nil {
						return err
					}

					data, err := ctx.Client.GetMonth(planID, positional[1])
					if err != nil {
						return err
					}

					headers := []string{"MONTH", "NOTE", "INCOME", "BUDGETED", "ACTIVITY", "TO_BE_BUDGETED"}
					extractRows := func(data []byte) ([][]string, error) {
						var resp struct {
							Data struct {
								Month struct {
									Month        string `json:"month"`
									Note         string `json:"note"`
									Income       int64  `json:"income"`
									Budgeted     int64  `json:"budgeted"`
									Activity     int64  `json:"activity"`
									ToBeBudgeted int64  `json:"to_be_budgeted"`
								} `json:"month"`
							} `json:"data"`
						}
						if err := json.Unmarshal(data, &resp); err != nil {
							return nil, err
						}
						m := resp.Data.Month
						return [][]string{{
							m.Month,
							m.Note,
							strconv.FormatInt(m.Income, 10),
							strconv.FormatInt(m.Budgeted, 10),
							strconv.FormatInt(m.Activity, 10),
							strconv.FormatInt(m.ToBeBudgeted, 10),
						}}, nil
					}

					return ctx.Printer.PrintResult(data, headers, extractRows)
				},
			},
		},
	}
	a.registerCommand(cmd)
}
