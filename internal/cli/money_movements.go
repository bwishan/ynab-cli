package cli

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/bwishan/ynab-cli/internal/config"
)

func (a *App) registerMoneyMovementCommands() {
	// money-movements command
	mmCmd := &Command{
		Name:        "money-movements",
		Description: "Manage money movements",
		Subcommands: map[string]*Subcommand{
			"list": {
				Name:        "list",
				Description: "List all money movements for a plan",
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

					data, err := ctx.Client.GetMoneyMovements(planID)
					if err != nil {
						return err
					}

					headers := []string{"ID", "NAME", "AMOUNT", "TYPE", "DELETED"}
					extractRows := func(data []byte) ([][]string, error) {
						var resp struct {
							Data struct {
								MoneyMovements []struct {
									ID      string `json:"id"`
									Name    string `json:"name"`
									Amount  int64  `json:"amount"`
									Type    string `json:"type"`
									Deleted bool   `json:"deleted"`
								} `json:"money_movements"`
							} `json:"data"`
						}
						if err := json.Unmarshal(data, &resp); err != nil {
							return nil, err
						}
						var rows [][]string
						for _, mm := range resp.Data.MoneyMovements {
							rows = append(rows, []string{
								mm.ID,
								mm.Name,
								strconv.FormatInt(mm.Amount, 10),
								mm.Type,
								fmt.Sprintf("%t", mm.Deleted),
							})
						}
						return rows, nil
					}

					return ctx.Printer.PrintResult(data, headers, extractRows)
				},
			},
			"list-by-month": {
				Name:        "list-by-month",
				Description: "List money movements for a specific month",
				Usage:       "<plan-id> <month>",
				Run: func(ctx *Context, args []string) error {
					var positional []string
					for _, arg := range args {
						if !strings.HasPrefix(arg, "--") {
							positional = append(positional, arg)
						}
					}

					if len(positional) < 2 {
						return fmt.Errorf("usage: ynab money-movements list-by-month <plan-id> <month>")
					}

					planID, err := config.ResolvePlan(positional[0])
					if err != nil {
						return err
					}

					data, err := ctx.Client.GetMoneyMovementsByMonth(planID, positional[1])
					if err != nil {
						return err
					}

					headers := []string{"ID", "NAME", "AMOUNT", "TYPE", "DELETED"}
					extractRows := func(data []byte) ([][]string, error) {
						var resp struct {
							Data struct {
								MoneyMovements []struct {
									ID      string `json:"id"`
									Name    string `json:"name"`
									Amount  int64  `json:"amount"`
									Type    string `json:"type"`
									Deleted bool   `json:"deleted"`
								} `json:"money_movements"`
							} `json:"data"`
						}
						if err := json.Unmarshal(data, &resp); err != nil {
							return nil, err
						}
						var rows [][]string
						for _, mm := range resp.Data.MoneyMovements {
							rows = append(rows, []string{
								mm.ID,
								mm.Name,
								strconv.FormatInt(mm.Amount, 10),
								mm.Type,
								fmt.Sprintf("%t", mm.Deleted),
							})
						}
						return rows, nil
					}

					return ctx.Printer.PrintResult(data, headers, extractRows)
				},
			},
		},
	}
	a.registerCommand(mmCmd)

	// money-movement-groups command
	mmgCmd := &Command{
		Name:        "money-movement-groups",
		Description: "Manage money movement groups",
		Subcommands: map[string]*Subcommand{
			"list": {
				Name:        "list",
				Description: "List all money movement groups for a plan",
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

					data, err := ctx.Client.GetMoneyMovementGroups(planID)
					if err != nil {
						return err
					}

					headers := []string{"ID", "NAME", "DELETED"}
					extractRows := func(data []byte) ([][]string, error) {
						var resp struct {
							Data struct {
								MoneyMovementGroups []struct {
									ID      string `json:"id"`
									Name    string `json:"name"`
									Deleted bool   `json:"deleted"`
								} `json:"money_movement_groups"`
							} `json:"data"`
						}
						if err := json.Unmarshal(data, &resp); err != nil {
							return nil, err
						}
						var rows [][]string
						for _, g := range resp.Data.MoneyMovementGroups {
							rows = append(rows, []string{
								g.ID,
								g.Name,
								fmt.Sprintf("%t", g.Deleted),
							})
						}
						return rows, nil
					}

					return ctx.Printer.PrintResult(data, headers, extractRows)
				},
			},
			"list-by-month": {
				Name:        "list-by-month",
				Description: "List money movement groups for a specific month",
				Usage:       "<plan-id> <month>",
				Run: func(ctx *Context, args []string) error {
					var positional []string
					for _, arg := range args {
						if !strings.HasPrefix(arg, "--") {
							positional = append(positional, arg)
						}
					}

					if len(positional) < 2 {
						return fmt.Errorf("usage: ynab money-movement-groups list-by-month <plan-id> <month>")
					}

					planID, err := config.ResolvePlan(positional[0])
					if err != nil {
						return err
					}

					data, err := ctx.Client.GetMoneyMovementGroupsByMonth(planID, positional[1])
					if err != nil {
						return err
					}

					headers := []string{"ID", "NAME", "DELETED"}
					extractRows := func(data []byte) ([][]string, error) {
						var resp struct {
							Data struct {
								MoneyMovementGroups []struct {
									ID      string `json:"id"`
									Name    string `json:"name"`
									Deleted bool   `json:"deleted"`
								} `json:"money_movement_groups"`
							} `json:"data"`
						}
						if err := json.Unmarshal(data, &resp); err != nil {
							return nil, err
						}
						var rows [][]string
						for _, g := range resp.Data.MoneyMovementGroups {
							rows = append(rows, []string{
								g.ID,
								g.Name,
								fmt.Sprintf("%t", g.Deleted),
							})
						}
						return rows, nil
					}

					return ctx.Printer.PrintResult(data, headers, extractRows)
				},
			},
		},
	}
	a.registerCommand(mmgCmd)
}
