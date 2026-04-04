package cli

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/bwishan/ynab-cli/internal/config"
)

func (a *App) registerMonthCommands() {
	monthsCmd := &cobra.Command{
		Use:   "months",
		Short: "Manage budget months",
	}

	listCmd := &cobra.Command{
		Use:   "list [plan-id]",
		Short: "List all budget months for a plan",
		RunE: func(cmd *cobra.Command, args []string) error {
			var planIDArg string
			if len(args) > 0 {
				planIDArg = args[0]
			}

			planID, err := config.ResolvePlan(planIDArg)
			if err != nil {
				return err
			}

			lastKnowledge, _ := cmd.Flags().GetInt64("last-knowledge")

			data, err := a.client.GetMonths(planID, lastKnowledge)
			if err != nil {
				return err
			}

			headers := []string{"MONTH", "NOTE", "INCOME", "BUDGETED", "ACTIVITY", "TO_BE_BUDGETED", "DELETED"}
			extractRows := func(data []byte) ([][]string, error) {
				var resp struct {
					Data struct {
						Months []struct {
							Month        string `json:"month"`
							Note         string `json:"note"`
							Income       int64  `json:"income"`
							Budgeted     int64  `json:"budgeted"`
							Activity     int64  `json:"activity"`
							ToBeBudgeted int64  `json:"to_be_budgeted"`
							Deleted      bool   `json:"deleted"`
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

			return a.printer.PrintResult(data, headers, extractRows)
		},
	}
	listCmd.Flags().Int64("last-knowledge", 0, "Server knowledge for delta sync")

	getCmd := &cobra.Command{
		Use:   "get <plan-id> <month>",
		Short: "Get a specific budget month",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			planID, err := config.ResolvePlan(args[0])
			if err != nil {
				return err
			}

			data, err := a.client.GetMonth(planID, args[1])
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

			return a.printer.PrintResult(data, headers, extractRows)
		},
	}

	monthsCmd.AddCommand(listCmd, getCmd)
	a.rootCmd.AddCommand(monthsCmd)
}
