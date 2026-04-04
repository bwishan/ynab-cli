package cli

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/bwishan/ynab-cli/internal/config"
)

func (a *App) registerMoneyMovementCommands() {
	// money-movements command
	mmCmd := &cobra.Command{
		Use:   "money-movements",
		Short: "Manage money movements",
	}

	mmListCmd := &cobra.Command{
		Use:   "list [plan-id]",
		Short: "List all money movements for a plan",
		RunE: func(cmd *cobra.Command, args []string) error {
			var planIDArg string
			if len(args) > 0 {
				planIDArg = args[0]
			}

			planID, err := config.ResolvePlan(planIDArg)
			if err != nil {
				return err
			}

			data, err := a.client.GetMoneyMovements(planID)
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

			return a.printer.PrintResult(data, headers, extractRows)
		},
	}

	mmListByMonthCmd := &cobra.Command{
		Use:   "list-by-month <plan-id> <month>",
		Short: "List money movements for a specific month",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			planID, err := config.ResolvePlan(args[0])
			if err != nil {
				return err
			}

			data, err := a.client.GetMoneyMovementsByMonth(planID, args[1])
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

			return a.printer.PrintResult(data, headers, extractRows)
		},
	}

	mmCmd.AddCommand(mmListCmd, mmListByMonthCmd)
	a.rootCmd.AddCommand(mmCmd)

	// money-movement-groups command
	mmgCmd := &cobra.Command{
		Use:   "money-movement-groups",
		Short: "Manage money movement groups",
	}

	mmgListCmd := &cobra.Command{
		Use:   "list [plan-id]",
		Short: "List all money movement groups for a plan",
		RunE: func(cmd *cobra.Command, args []string) error {
			var planIDArg string
			if len(args) > 0 {
				planIDArg = args[0]
			}

			planID, err := config.ResolvePlan(planIDArg)
			if err != nil {
				return err
			}

			data, err := a.client.GetMoneyMovementGroups(planID)
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

			return a.printer.PrintResult(data, headers, extractRows)
		},
	}

	mmgListByMonthCmd := &cobra.Command{
		Use:   "list-by-month <plan-id> <month>",
		Short: "List money movement groups for a specific month",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			planID, err := config.ResolvePlan(args[0])
			if err != nil {
				return err
			}

			data, err := a.client.GetMoneyMovementGroupsByMonth(planID, args[1])
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

			return a.printer.PrintResult(data, headers, extractRows)
		},
	}

	mmgCmd.AddCommand(mmgListCmd, mmgListByMonthCmd)
	a.rootCmd.AddCommand(mmgCmd)
}
