package cli

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/bwishan/ynab-cli/internal/config"
)

func (a *App) registerAccountCommands() {
	accountsCmd := &cobra.Command{
		Use:   "accounts",
		Short: "Manage accounts",
	}

	listCmd := &cobra.Command{
		Use:   "list [plan-id]",
		Short: "List all accounts for a plan",
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

			data, err := a.client.GetAccounts(planID, lastKnowledge)
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

			return a.printer.PrintResult(data, headers, extractRows)
		},
	}
	listCmd.Flags().Int64("last-knowledge", 0, "Server knowledge for delta sync")

	getCmd := &cobra.Command{
		Use:   "get <plan-id> <account-id>",
		Short: "Get a specific account",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			planID, err := config.ResolvePlan(args[0])
			if err != nil {
				return err
			}
			accountID := args[1]

			data, err := a.client.GetAccountByID(planID, accountID)
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

			return a.printer.PrintResult(data, headers, extractRows)
		},
	}

	createCmd := &cobra.Command{
		Use:   "create [plan-id]",
		Short: "Create a new account",
		RunE: func(cmd *cobra.Command, args []string) error {
			var planIDArg string
			if len(args) > 0 {
				planIDArg = args[0]
			}

			name, _ := cmd.Flags().GetString("name")
			acctType, _ := cmd.Flags().GetString("type")
			balance, _ := cmd.Flags().GetInt64("balance")

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

			data, err := a.client.CreateAccount(planID, body)
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

			return a.printer.PrintResult(data, headers, extractRows)
		},
	}
	createCmd.Flags().String("name", "", "Account name (required)")
	createCmd.Flags().String("type", "", "Account type (required)")
	createCmd.Flags().Int64("balance", 0, "Initial balance in milliunits")

	accountsCmd.AddCommand(listCmd, getCmd, createCmd)
	a.rootCmd.AddCommand(accountsCmd)
}
